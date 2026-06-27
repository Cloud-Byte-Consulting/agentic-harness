# OpenAI Agents SDK Reference

The hands-on API for building agents. Six primitives: **Agent, Runner, tools, handoffs, guardrails, tracing**. Minimal, Pythonic, model-agnostic, open source. Idiomatically similar to Flask/Pydantic — plain Python objects and decorated functions, no YAML, no hidden meta-language.

## Table of contents
- Install & environment
- Agent
- Runner (the agent loop)
- Sync vs. async
- The result object
- Model selection & ModelSettings
- Third-party models via LiteLLM
- Local (run) context
- Visualization
- Low-level loop control

## Install & environment

```bash
pip install openai-agents
pip install "openai-agents[litellm]"   # non-OpenAI providers
pip install "openai-agents[viz]"       # graph visualization
pip install python-dotenv              # load .env
```

Requires Python 3.9+. Store the key in a `.env` file (never hardcode, never commit):

```
OPENAI_API_KEY="sk-..."
```

```python
from dotenv import load_dotenv
load_dotenv()  # loads OPENAI_API_KEY into the environment
```

Treat the key like a password; revoke from the OpenAI dashboard if leaked. Google Colab works too: `!pip install openai-agents` then set `os.environ["OPENAI_API_KEY"]`.

## Agent

A highly configurable wrapper that turns an LLM into an agent. There is a single `Agent` class — behavior comes entirely from config, not subclassing.

```python
from agents import Agent, Runner, function_tool

agent = Agent(
    name="Customer Service Agent",       # identification/tracing
    model="gpt-4o",                       # the brain
    instructions="You are an AI agent that resolves customer issues cheerfully.",
    tools=[get_account_information, refund_customer_payment],  # callable tools
    handoffs=[retention_agent],           # agents it can transfer control to
    # also: input_guardrails, output_guardrails, output_type, model_settings,
    #       tool_use_behavior  (see other reference files)
)
```

Key params:
- `name` — label only.
- `instructions` — the system prompt (role, scope, tool/handoff conditions, prohibitions).
- `model` — model name string (or a LiteLLM string, or a model object).
- `tools` — list of `@function_tool` functions, hosted tools, or `agent.as_tool(...)`.
- `handoffs` — list of agents (or `handoff(...)` objects) it can delegate full control to.
- `output_type` — a Pydantic model to force structured output.
- `model_settings` — a `ModelSettings(...)` for `temperature`, `max_tokens`, `tool_choice`, etc.

The same agent either answers in natural language or calls a tool / hands off — the runtime decides based on the model's output.

## Runner (the agent loop)

The engine that drives the perceive→reason→act loop: call model → handle tool calls/handoffs → repeat until final output.

```python
result = await Runner.run(agent, "Where is my order? It's #XYZ")          # async
result = Runner.run_sync(agent, "Where is my order? It's #XYZ")           # sync
```

Per turn: the runtime sends the message list to the model. If the model returns a final answer, the loop ends and `final_output` is set. If it returns a tool call or handoff, the runtime executes it, appends the output, and re-runs — until a final answer or `max_turns`.

**`max_turns`** caps iterations — the safety valve against infinite spins from a bad config or unsolvable task:

```python
result = await Runner.run(agent, prompt, max_turns=10)
```

Pass a `session=` for automatic conversation memory, and `context=` for local run context (see below and `memory-and-sessions.md`).

## Sync vs. async

The SDK supports both; **async is preferred** because agentic workflows spend most of their time waiting on I/O (model calls, tool/API calls), and async enables parallel tools/agents.

```python
import asyncio
from agents import Agent, Runner

async def main():
    agent = Agent(name="A", instructions="Answer questions.")
    result = await Runner.run(agent, "Hello")
    print(result.final_output)

asyncio.run(main())
```

`async def` functions must be `await`ed and run inside an event loop (`asyncio.run(...)`). For simple single-agent, single-tool demos, `run_sync` is fine. Async guardrails are required when a guardrail itself calls another agent (see `guardrails-and-safety.md`).

## The result object

`Runner.run`/`run_sync` returns a `RunResult`:
- `result.final_output` — the answer (a string, or your `output_type` Pydantic object).
- `result.to_input_list()` — the full conversation as a list of message items, ready to pass back in for the next turn (manual memory).
- `result.new_items` — the items produced this run (model messages, `ToolCallItem`s, handoffs) — inspect for unit testing.
- `result.last_agent` — which agent ended the run; feed it back as the starting agent to continue a multi-agent conversation.

## Model selection & ModelSettings

Every agent can use a different model:

```python
fast   = Agent(name="Triage", instructions="...", model="gpt-4o")
deep   = Agent(name="Solver", instructions="...", model="o3-pro")   # slower, ~10x cost
```

Tune generation without changing models:

```python
from agents.model_settings import ModelSettings

creative = Agent(name="Writer", instructions="...", model="gpt-4o",
                 model_settings=ModelSettings(temperature=1.0, max_tokens=300))
precise  = Agent(name="Classifier", instructions="...", model="gpt-4o",
                 model_settings=ModelSettings(temperature=0.2, max_tokens=50))
```

`temperature` (0.2 ≈ deterministic, 0.8–1.0 ≈ creative), `max_tokens` (caps length — too small truncates), and `tool_choice` (see `tools-and-function-calling.md`).

## Third-party models via LiteLLM

The SDK is model-agnostic. With `openai-agents[litellm]` installed and the provider's key in `.env`, set `model=` to a LiteLLM string:

```python
agent = Agent(name="Claude Agent", instructions="You are an AI agent.",
              model="litellm/anthropic/claude-opus-4-20250514")
```

Other examples: `litellm/gemini/gemini-pro`, `litellm/meta_llama/Llama-3.3-70B-Instruct`. LiteLLM handles key routing and response formatting, so agent logic is unchanged. Useful for benchmarking providers or honoring privacy/latency/cost constraints (e.g. prototype on GPT-4o, ship on Claude).
> Note: model IDs above are illustrative; use current model names from each provider. When the provider is Anthropic/Claude, consult the `claude-api` skill for current model ids.
> Note: OpenAI-hosted tools (web search, file search, code interpreter, etc.) require OpenAI models and won't work with third-party models.

## Local (run) context

Pass privileged/session data into tools **without** putting it in the model's prompt. Define a context dataclass, parameterize the agent with it, and have tools accept a `RunContextWrapper`:

```python
from dataclasses import dataclass
from agents import Agent, Runner, RunContextWrapper, function_tool

@dataclass
class OrderContext:
    customer_name: str
    order_id: str
    shipping_status: str

@function_tool
def get_shipping_status(wrapper: RunContextWrapper[OrderContext]) -> str:
    """Provide the shipping status for the current order."""
    ctx = wrapper.context
    return f"Hi {ctx.customer_name}, order {ctx.order_id} is: {ctx.shipping_status}."

agent = Agent[OrderContext](name="Shipping Support",
                            instructions="Check the user's shipping status.",
                            tools=[get_shipping_status])

ctx = OrderContext("Henry", "123", "Delayed")
result = Runner.run_sync(agent, "Where is my order?", context=ctx)
```

The data (auth tokens, user prefs, internal state) reaches tools but never the LLM prompt — good for sensitive/proprietary info.

## Visualization

With `openai-agents[viz]`:

```python
from agents.extensions.visualization import draw_graph
draw_graph(triage_agent, filename="graph_visualization")  # writes a .png
```

Agents are boxes, tools are ellipses, solid arrows = handoffs, dotted = tool calls. Great for verifying a multi-agent system is wired as intended and for documentation.

## Low-level loop control

The high-level `Runner` handles retries, step limits, and concurrency. Advanced users can drive lower-level primitives to interleave agent execution with other async work (e.g. inside a FastAPI route) for full control over timing, concurrency, and resource management. The execution loop itself is modular and swappable — but `Runner.run` is the right default for almost everything.
