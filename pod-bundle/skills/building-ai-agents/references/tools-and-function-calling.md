# Tools, Function Calling & Structured Outputs

Tools are how an agent acts on the world. In the OpenAI Agents SDK, any Python function becomes a tool with one decorator.

## Table of contents
- Defining a function tool
- Overriding tool metadata
- Async tools
- Complex inputs with Pydantic
- Structured outputs (output_type)
- Controlling tool behavior: tool_choice
- Controlling tool behavior: tool_use_behavior
- Chained tool calls
- OpenAI-hosted tools
- Agents as tools
- MCP (Model Context Protocol)
- Tool retrieval at scale & auditing tool knowledge
- Tool design rules

## Defining a function tool

```python
from agents import Agent, Runner, function_tool

@function_tool
def get_order_status(order_id: int) -> str:
    """Return the order status given an order ID.

    Args:
        order_id: The customer's numeric order ID.
    """
    if order_id in (100, 101):
        return "Delivered"
    if order_id in (200, 201):
        return "Delayed"
    return "Unknown"

agent = Agent(name="Support", instructions="Help with orders.",
              tools=[get_order_status])
result = Runner.run_sync(agent, "Status of order 200?")
```

The SDK derives the JSON schema automatically from the **function name** (tool name), **docstring** (description + arg docs), and **type hints** (parameter types). You write zero schema by hand. The model sees only this metadata when deciding whether/how to call — so a clear, human-readable name and a docstring that states what the tool does and what inputs it expects (with units/formats) are essential.

A tool being available does not mean it gets called — tool choice is non-deterministic and context-driven. Asking "how do I reset my password?" won't trigger `get_order_status`.

## Overriding tool metadata

When the function name/docstring isn't expressive enough, override:

```python
@function_tool(
    name_override="Get Status of Current Order",
    description_override="Returns the status of an order given the customer's Order ID",
)
def get_order_status(order_id: int) -> str:
    ...
```

Useful for disambiguating similar tools, enforcing naming/formatting standards, or giving the model a clearer signal than a generic function name.

## Async tools

`@function_tool` works on `async def` too — the SDK awaits the result automatically. Use it for tools that hit external APIs/DBs:

```python
@function_tool
async def fetch_user(user_id: str) -> str:
    """Fetch a user record by ID."""
    return await db.get_user(user_id)
```

## Complex inputs with Pydantic

For nested/hierarchical/multi-field inputs, use a Pydantic `BaseModel` as the single parameter. The SDK generates the nested schema **and** validates the model's arguments — if the LLM hallucinates a missing field or wrong type, Pydantic raises `ValidationError`, which you can catch for resilience.

```python
from pydantic import BaseModel
from typing import List
from agents import function_tool

class Crypto(BaseModel):
    coin_ids: List[str]   # e.g. ["bitcoin", "ethereum"]

@function_tool
def get_crypto_prices(crypto: Crypto) -> str:
    """Get current USD prices for a list of cryptocurrencies."""
    ids = ",".join(crypto.coin_ids)
    return fetch_prices(ids)   # one call for all coins
```

Tip: if you want one batched call instead of N calls, say so in the instructions ("call the tool only once for all requests") — otherwise the agent may call per item.

## Structured outputs (output_type)

Force an agent's *final answer* into a typed object by setting `output_type`. The SDK parses the model output through Pydantic; `result.final_output` is then a typed instance, not a string — ideal for downstream code/APIs.

```python
from pydantic import BaseModel
from agents import Agent

class EmailOutput(BaseModel):
    to_email: str
    from_email: str
    subject: str
    html_email: str

email_agent = Agent(
    name="Email Creation Agent",
    instructions="Write a concise, personalized email.",
    output_type=EmailOutput,
    model="gpt-4.1",
)
# result.final_output is an EmailOutput; access result.final_output.subject, etc.
```

## Controlling tool behavior: tool_choice

`ModelSettings.tool_choice` controls the model's approach to calling tools:
- `"auto"` — model decides whether and which tool to call (**default**).
- `"required"` — model **must** call a tool; it cannot answer from internal knowledge. If no tool fits, it errors/refuses. Use when answers must come from a trusted backend (compliance, auditability, sandboxed testing).
- `"none"` — model may not call any tool.
- `"get_weather"` — a specific tool name forces that exact tool (good for isolating/testing one tool).

```python
from agents import Agent, ModelSettings

agent = Agent(
    name="Strict Support",
    instructions="Always check the backend for order status. Do not guess.",
    tools=[get_order_status],
    model_settings=ModelSettings(tool_choice="required"),
)
```

## Controlling tool behavior: tool_use_behavior

`Agent.tool_use_behavior` controls what happens *after* a tool returns:
- `"run_llm_again"` — feed the tool output back to the model, which interprets it and produces the final answer (**default**). Use when the output needs phrasing/synthesis.
- `"stop_on_first_tool"` — the first tool's output **is** the final answer; no further model call. Use when the raw output is the answer (a computation, a DB row, a generated invoice) — saves a model call and guarantees the exact tool output reaches the user.
- `StopAtTools.stop_at_tool_names([...])` — stop and return raw output when any of the named tools fire (e.g. trigger points like "speak to a manager", or sensitive outputs like invoices that must not be reworded).

```python
from agents import Agent, ModelSettings

agent = Agent(
    name="MortgageAdvisor",
    instructions="You are a mortgage assistant.",
    tools=[calculate_mortgage],
    tool_use_behavior="stop_on_first_tool",
    model_settings=ModelSettings(tool_choice="required"),
)
# Forces the tool, returns its exact output -> deterministic, auditable, one fewer LLM call.
```

## Chained tool calls

The model decides the order of tool calls and can feed one tool's output into the next, or call the same tool multiple times — no hardcoded workflow needed.

```python
@function_tool
def get_customer_orders(customer_id: str) -> list:
    """Return all order IDs for a customer."""
    ...

@function_tool
def get_order_information(order_id: str) -> str:
    """Return detailed info for one order."""
    ...

agent = Agent(name="Support", instructions="Help customers.",
              tools=[get_customer_orders, get_order_information])
# "Status of all my orders? Customer ID CUST123" -> agent calls get_customer_orders,
# then get_order_information for each returned ID, then synthesizes.
```

## OpenAI-hosted tools

Prebuilt, OpenAI-managed tools — import, instantiate, attach. They require OpenAI models and incur token costs. Don't reinvent these:

| Tool | What it does |
|---|---|
| `WebSearchTool` | Real-time web search (optional `user_location`, `search_context_size`) |
| `FileSearchTool` | RAG over a vector store (`vector_store_ids=[...]`) |
| `CodeInterpreterTool` | Run Python in a sandbox (for math/data) |
| `ImageGenerationTool` | Generate images from prompts |
| `ComputerTool` | Drive a computer/browser |
| `LocalShellTool` | Run shell commands locally |

```python
from agents import Agent, Runner, WebSearchTool

agent = Agent(name="WebTool", instructions="Answer web questions in one sentence.",
              tools=[WebSearchTool()])
result = Runner.run_sync(agent, "Who won the 2025 Stanley Cup?")
```

`FileSearchTool` is OpenAI's hosted RAG: upload files to a vector store in the platform UI, then pass its ID. For RAG internals (chunking, embeddings, retrievers), see the `rag-and-knowledge-graphs` skill. `CodeInterpreterTool` and `ImageGenerationTool` take a config object (`container={"type":"auto"}` for the interpreter). Image generation is especially prone to hallucination.

## Agents as tools

Wrap an entire agent as a callable tool with `as_tool()`. The orchestrator stays in control and calls the sub-agent for a subtask (call-and-return), unlike a handoff (transfer of control).

```python
orchestrator = Agent(
    name="Distance Agent",
    instructions="Use LocationAgent for coordinates, DistanceCalculator for distance.",
    tools=[
        location_agent.as_tool(
            tool_name="LocationAgent",
            tool_description="Returns latitude/longitude for a location"),
        distance_agent.as_tool(
            tool_name="DistanceCalculator",
            tool_description="Computes distance between two lat/long points"),
    ],
)
```

The `tool_name`/`tool_description` drive the orchestrator's selection — make them specific. Use agents-as-tool when you need central control, multi-worker synthesis, and maximum oversight. See `multi-agent-and-handoffs.md` for the handoff vs. as-tool decision.

## MCP (Model Context Protocol)

A standard "USB-C for agents and tools": an MCP server exposes tools that any MCP-compatible host can consume, regardless of framework or provider. Consume a server's tools without writing your own:

```python
from agents import Agent, Runner, HostedMCPTool
from agents.tool import Mcp

mcp_tool = HostedMCPTool(tool_config=Mcp(
    server_label="CryptocurrencyPriceFetcher",
    server_url="https://mcp.api.coingecko.com/sse",
    type="mcp",
    require_approval="never",   # or require approval before each call
))
agent = Agent(name="Crypto Agent", instructions="Return crypto prices.",
              tools=[mcp_tool])
```

The tool logic runs on the remote MCP server, not your machine. The agent can reason over all tools the server exposes. **Security:** enforce auth, rate-limit, and watch data exposure when sending inputs to / receiving outputs from external MCP servers.

## Tool retrieval at scale & auditing tool knowledge

A handful of tools fit in the prompt; thousands do not. Once a catalog has hundreds-to-tens-of-thousands of tools (large MCP fleets, enterprise API hubs), you can't list them all every turn — you need a **retrieval step** that narrows the catalog to a few candidate tools before the agent reasons. This retrieval is now the dominant bottleneck in large-tool-catalog agents, and it fails in non-obvious ways. Two patterns dominate, both with traps:

- **Embedding-based retrieval (the default).** Embed each tool's name+description, embed the query, return top-k by cosine similarity. Simple and model-agnostic, but a small encoder under-captures the semantics of many near-duplicate APIs (two QR-code generators, two enterprise-search APIs), and the retrieved descriptions still have to be injected into context, compounding token cost across a multi-step loop. Sparse lexical retrieval (BM25) remains a surprisingly competitive baseline on short, natural queries — keep it as a sanity check.
- **Parametric / generative retrieval.** Bake the catalog into a model that *generates* a tool identifier directly (one token or a short sequence per tool), often constrained at decode time to valid tools. Can post very high recall on in-distribution benchmarks — but that number is easy to over-trust (see below).

**The headline failure mode: high benchmark recall is a poor proxy for real-world tool selection.** A diagnostic study of parametric tool retrieval over a ~47k-tool catalog (ToolSense, on the ToolBench catalog) found that configurations scoring ~0.90–0.96 recall on the verbose, fully-specified benchmark queries *collapsed by ~50–64 percentage points* on short, intent-focused queries phrased the way a real user types ("show me odds for today's matches"), several falling **below** a plain embedding baseline. The benchmark queries paraphrased the API docs, so the model won by surface lexical overlap, not understanding. Lessons that generalize to any retrieval approach:

- **Evaluate on realistic, paraphrased queries, not doc-style ones.** Build your eval set from how users actually phrase requests (short, ambiguous, no API jargon), at several ambiguity tiers (one obvious tool / a few plausible tools / a vague multi-tool goal). A retriever tuned on verbose descriptions silently overfits to vocabulary.
- **Don't trust a single recall number; audit whether the model *knows* the tool.** The same study probed selected tools with simple factual MCQ/QA ("what does this tool do?", "does it support X?") and found models that retrieved well scored **near random** (~25–31% on 4-way MCQ, ~50% on yes/no) — they had learned a lookup table, not tool semantics, and broke the moment phrasing shifted. Add a cheap probe: ask the agent factual questions about a sampled tool and check the answers against the spec.
- **Beware decode-time crutches masking the gap.** If valid outputs are constrained to a fixed set at inference (a trie, an enum, a fixed tool list), strong "recall" can hide that the model couldn't have produced the right tool on its own. Test free-form too: the gap between constrained and unconstrained selection is your real internalization signal.
- **Over-specializing the retriever to one query style trades away another.** Retraining purely on user-style queries recovered the realistic-query score (+35pp) but dropped the doc-style score and *worsened* factual probes — narrow retrieval tuning erodes general tool knowledge. Keep eval coverage broad.

Practical takeaway for agent builders: treat tool retrieval as a first-class, separately-evaluated subsystem. Curate distinct, non-overlapping tool descriptions (near-duplicates are where retrieval fails); keep catalogs scoped per agent so the retriever has fewer confusables; prefer a hybrid (embeddings + lexical) retriever with a reranker over any single method; and gate large catalogs behind an eval that uses realistic queries plus a factual probe, not just in-distribution recall.

## Tool design rules

- **One responsibility per tool**, precise name, full docstring, exact type hints — the model selects purely from this metadata.
- **Push real work into tools.** LLMs hallucinate arithmetic and stale facts; do math in a tool or the code interpreter, fetch live/proprietary data behind a tool. Tools are live extensions of the model's knowledge and capability.
- **Validate inputs** with Pydantic; catch `ValidationError` for resilience against hallucinated args.
- **Handle errors inside the tool** and return a clear string the model can act on (e.g. `"No order found for {id}"`) rather than throwing raw.
- **Don't trust LLM-built SQL/queries blindly.** Use parameterized queries, least-privilege, and real authz (OAuth/RBAC) — never interpolate keys/credentials into queries (SQL-injection and key-leakage risk).
- **Prefer hosted/MCP tools** over rebuilding common capabilities (web search, file search, code execution).
