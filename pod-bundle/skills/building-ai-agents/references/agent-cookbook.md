# Agent Cookbook

Concrete blueprints for the agent types you'll build most. Each recipe lists the shape
(components), the key decisions, and a minimal SDK sketch. Adapt freely. Cross-references point
to the other reference files for depth.

## Table of contents
- The cognitive-architecture taxonomy (which building blocks)
- Recipe: customer-support agent
- Recipe: research agent
- Recipe: coding agent
- Recipe: data-analysis agent
- Recipe: RAG-backed Q&A agent
- Recipe: workflow-automation (multi-agent pipeline)
- Worked example: full support employee
- Worked example: outreach workflow

## The cognitive-architecture taxonomy (which building blocks)

Most agents are a composition of three reusable cognitive patterns — pick the ones the task
needs:

- **Autonomous decision-making** — perceive → reason → act → learn each turn. Reactive,
  single-step. Good for triage, support, alerting.
- **Planning** — decompose a goal into a task tree, sequence/monitor/revise. Good for
  long-horizon, multi-step work (research, workflows).
- **Memory-augmented** — working + episodic + semantic memory for continuity/personalization.
  Good for assistants and anything spanning sessions.

These map to interaction-paradigm levels: direct LLM call (stateless) → proxy/translator agent
→ tool-augmented assistant → autonomous agent → multi-agent system. Climb only as high as the
task demands.

## Recipe: customer-support agent

**Shape:** instructions (persona + policy) → function tool(s) for order/account lookups (with
auth) → `FileSearchTool` for policy/FAQ docs → input guardrail (relevance) → handoff to a
specialist (e.g. retention) → `Session` for multi-turn state.

**Key decisions:** keep order lookups behind a tool with real authz (not LLM-built SQL); use a
cheap relevance guardrail to reject off-topic prompts before spending; hand off to a retention
agent on cancel intent; persist the session for context across turns.

```python
from agents import Agent, Runner, SQLiteSession, function_tool, FileSearchTool

@function_tool
def get_order_status(order_id: int) -> str:
    """Return delivery status for an order id (after authz)."""
    ...

retention_agent = Agent(name="Retention",
    instructions="Empathize, find the pain point, offer up to a credit to retain the customer.",
    tools=[get_order_status])

support = Agent(
    name="Support",
    instructions=("Help with orders and complaints. Use get_order_status for order data "
                  "(ask for the auth key). Use file search for policy questions. "
                  "Hand off to Retention if the user wants to cancel."),
    tools=[get_order_status, FileSearchTool(vector_store_ids=["vs_..."])],
    handoffs=[retention_agent],
    # input_guardrails=[relevance_guardrail]  # see guardrails-and-safety.md
)
session = SQLiteSession("cust_123")
Runner.run_sync(support, "Where is order 1002?", session=session)
```

See `multi-agent-and-handoffs.md`, `guardrails-and-safety.md`, `memory-and-sessions.md`.

## Recipe: research agent

**Shape:** planning agent + `WebSearchTool` (+ optional `FileSearchTool` for internal corpora)
+ structured output. Often hierarchical: a plan agent outlines → a search/synthesis agent
gathers → a report agent writes.

**Key decisions:** use a strong reasoning model for the planning/synthesis steps; cite sources
(web search returns links); cap turns; for deep multi-stage research consider a hierarchical
manager → specialist layout, but mind the latency/cost of extra tiers.

```python
from agents import Agent, Runner, WebSearchTool

researcher = Agent(
    name="Researcher",
    instructions=("Plan the research, search the web for current facts, and synthesize a "
                  "cited summary. Note source links."),
    tools=[WebSearchTool()],
    model="gpt-4o",
)
Runner.run_sync(researcher, "Summarize the latest on agentic AI frameworks, with sources.")
```

For recursive/long-horizon research harnesses, see the **llm-agent-orchestration** skill.

## Recipe: coding agent

**Shape:** instructions (coding standards) + `CodeInterpreterTool` (run/verify) or
`LocalShellTool` (run commands) + optional file/repo tools + tight guardrails.

**Key decisions:** never let the model "execute" in its head — run real code in the sandbox and
feed back results; gate shell/file access carefully (least privilege); add output guardrails to
catch unsafe commands; verify by running tests, not by trusting the explanation.

```python
from agents import Agent, Runner, CodeInterpreterTool
from agents.tool import CodeInterpreter

code_agent = Agent(
    name="Coder",
    instructions="Write and run Python to solve the task; verify by executing it.",
    tools=[CodeInterpreterTool(tool_config=CodeInterpreter(
        container={"type": "auto"}, type="code_interpreter"))],
)
Runner.run_sync(code_agent, "Compute the monthly payment on an 800k loan at 6% for 30 years.")
```

## Recipe: data-analysis agent

**Shape:** `CodeInterpreterTool` for computation/plotting, or composable function tools
(`load_csv`, `aggregate`, `plot_chart`, `validate`) + DB query tool. Often `output_type` for a
structured result (metrics + chart reference).

**Key decisions:** do all math/aggregation in tools or the interpreter (LLMs hallucinate
arithmetic); validate outputs; for DB access use parameterized queries and read-only roles.

```python
from agents import Agent, Runner, function_tool

@function_tool
def run_query(sql: str) -> list:
    """Run a read-only SQL query and return rows."""
    ...   # parameterized / least-privilege connection

analyst = Agent(name="Analyst",
    instructions="Answer data questions. Use run_query for data; never guess numbers.",
    tools=[run_query])
```

## Recipe: RAG-backed Q&A agent

**Shape:** the simplest knowledge agent — instructions + `FileSearchTool` over a vector store
(+ `Session` for follow-ups). The agent retrieves relevant chunks and grounds its answer.

**Key decisions:** instruct it to answer *only* from retrieved content and to say "not found"
rather than hallucinate; handle ambiguous queries and conflicting sources; cite chunks.

```python
from agents import Agent, Runner, FileSearchTool, SQLiteSession

kb_agent = Agent(name="KB",
    instructions=("Answer only from the vector store. If the answer isn't there, say so. "
                  "Cite the source."),
    tools=[FileSearchTool(vector_store_ids=["vs_..."])])
session = SQLiteSession("kb_session")
Runner.run_sync(kb_agent, "How high can the drone fly?", session=session)
```

RAG internals (chunking, embeddings, retrievers, reranking) → **rag-and-knowledge-graphs**.

## Recipe: workflow-automation (multi-agent pipeline)

**Shape:** deterministic orchestration of sequential agents — agent A gathers/produces, its
output feeds agent B, etc. Each step traced. Final step often `output_type` for a sendable
object.

**Key decisions:** use deterministic Python to chain steps when the sequence is fixed (more
predictable/auditable than letting an LLM route); give the writing/creative step a
writing-strong model; structure the final output for downstream systems (email, ticket, etc.).

```python
from agents import Agent, Runner, trace
from pydantic import BaseModel

class Email(BaseModel):
    to: str; subject: str; body: str

research = Agent(name="Research", instructions="Build a customer profile.", tools=[...])
writer   = Agent(name="Writer", instructions="Write a concise personalized email.",
                 output_type=Email, model="gpt-4.1")

for user_id in ["1", "2", "3"]:
    with trace(f"Outreach {user_id}"):
        profile = Runner.run_sync(research, user_id).final_output
        email = Runner.run_sync(writer, profile).final_output   # Email object
        send(email)   # plug into your delivery system
```

## Worked example: full support employee

Combine everything: SQLite order DB behind a `query_orders` tool (with an authz filter),
`FileSearchTool` over a policy doc, an input relevance guardrail (cheap classifier agent),
handoff to a retention agent on cancel intent, and a `SQLiteSession` for multi-turn state,
all under a `trace()`. Test it: off-topic prompt → guardrail blocks; order query → asks for
auth, then answers from the DB; policy question → answers from vector search; "cancel my
account" → hands off to retention, which offers a credit. (This is the assembly of the support
recipe plus guardrails plus memory.)

## Worked example: outreach workflow

A research agent gathers a customer's DB record + past chat transcripts (a JSON-reading tool) +
current news on their interests (`WebSearchTool`), producing a profile. A second agent (a
writing-tuned model, `output_type=Email`) turns the profile into a short personalized email
that subtly includes a promotion. A deterministic loop runs both per customer and saves each
`Email`. Extend with a tool to actually send via SMTP, or an agent to pick which customers to
target. This is the workflow-automation recipe made concrete — sequential agents, structured
output, traced per run.
