---
name: building-ai-agents
description: >-
  Design and build production LLM agents — agent anatomy (model, instructions, tools, run
  loop), function/tool calling, structured outputs, multi-agent handoffs, memory and sessions,
  guardrails, and tracing/evaluation. Use this when building or debugging an AI agent or
  assistant, writing function tools or tool schemas, wiring the OpenAI Agents SDK (Agent,
  Runner, function_tool, handoffs, guardrails, Sessions), adding RAG- or memory-backed
  assistants, routing between specialized agents, fixing infinite loops or runaway tool calls,
  enforcing input/output validation, retrieving tools over a large catalog, building
  evidence-first diagnosis agents that verify before concluding, or deciding whether an agent
  is worth building. Triggers include "build an agent", "tool calling", "function tool",
  "agent handoff", "multi-agent", "guardrail", "agent keeps looping", "trace my agent", "tool
  retrieval", "agent picks the wrong tool", and "agent agrees with me too easily".
---

# Building AI Agents

This skill equips Claude to design, implement, debug, and evaluate production LLM agents — autonomous systems that reason toward a goal and act on the world through tools, with the OpenAI Agents SDK as the primary reference implementation and provider-agnostic principles throughout.

## When to use this skill

- Building any LLM-powered agent or assistant: support bot, research agent, coding agent, data/analytics agent, RAG-backed Q&A, workflow automation.
- Writing or fixing **function tools** / tool schemas, structured (typed) outputs, or tool-choice behavior.
- Scaling to a **large tool catalog** — tool retrieval (embedding vs. parametric), choosing/evaluating a retriever, or auditing whether the agent actually selects the right tool on realistic queries.
- Building agents that **diagnose from unreliable input** — interactive troubleshooting, support, debugging, triage — where the user volunteers a possibly-wrong cause and you must verify before concluding.
- Wiring the **OpenAI Agents SDK**: `Agent`, `Runner`, `@function_tool`, `handoffs`, guardrails, `Sessions`, tracing.
- Designing **multi-agent** systems — triage/routing, specialist agents, handoffs vs. agent-as-tool.
- Adding **memory** (short-term sessions, long-term persistence, structured recall) or RAG-backed knowledge.
- Adding **guardrails** (input/output validation, relevance/safety tripwires) or fixing runaway loops, redundant tool calls, hallucinated tool args.
- Setting up **tracing, observability, and evaluation** (end-to-end + unit tests, LLM-as-judge).
- Deciding **whether to build** an agent at all, or whether a deterministic pipeline is the better call.

For deep multi-agent **orchestration** (recursive/hierarchical harnesses, agent-environment engineering, long-horizon control loops) see the `llm-agent-orchestration` skill. For **RAG internals** (chunking, embedding models, retrievers, vector DB tuning, reranking, knowledge graphs) see the `rag-and-knowledge-graphs` skill. This skill covers using RAG *as a tool inside an agent*, not building the retrieval stack.

## Core concepts

### The anatomy of an agent
An agent is not a model — it is four parts working in a loop:
1. **Model** — the reasoning engine (an LLM). Choice trades cost, latency, and capability. A triage agent can use a cheap fast model; a math/research step may need a reasoning model.
2. **Instructions** — the system prompt defining identity, role, goals, and constraints. Most agent debugging is prompt tuning.
3. **Tools** — typed functions the model can call to perceive and act on the world (APIs, DBs, code execution, other agents).
4. **The agent loop** — the runtime that repeatedly: sends context to the model → gets a final answer OR a tool call/handoff → executes it → appends the result → re-runs, until a final output or a turn cap.

Memory/knowledge wraps around this: working memory (conversation history in the prompt), long-term memory (persisted across sessions), and retrieved knowledge (RAG). The model is the brain; tools are the hands; memory is the notebook.

This maps cleanly onto the **perceive → reason → act → learn** cognitive loop. What makes a system "agentic" is the loop and the autonomy to choose *which* action — not the model alone.

### The minimal OpenAI Agents SDK program
The SDK exposes six primitives: **Agent, Runner, tools, handoffs, guardrails, tracing**. A working agent is a few lines:

```python
from agents import Agent, Runner

agent = Agent(
    name="Assistant",
    instructions="You are a helpful assistant.",
    model="gpt-4o",
)
result = Runner.run_sync(agent, "Tell me a joke")
print(result.final_output)
```

`Runner.run_sync` (sync) and `await Runner.run(...)` (async, preferred for tool/agent parallelism) drive the loop. `Runner.run(..., max_turns=N)` caps iterations — a safety valve against infinite spins. `result.final_output` is the answer; `result.to_input_list()` and `result.new_items` expose the full transcript for chaining and testing.

Install: `pip install openai-agents` (add `[litellm]` for non-OpenAI models, `[viz]` for graphs). Set `OPENAI_API_KEY` via a `.env` file — never hardcode keys. The SDK is model-agnostic: swap the `model=` string (e.g. `litellm/anthropic/claude-...`) without changing agent logic.

### Tools are just typed, documented functions
Decorate any Python function with `@function_tool`. The SDK derives the tool's name, description, and JSON schema from the function name, **docstring**, and **type hints** — so write good ones; the model decides whether and how to call the tool from exactly that metadata.

```python
from agents import Agent, function_tool

@function_tool
def get_order_status(order_id: int) -> str:
    """Return the delivery status for a given order ID.

    Args:
        order_id: The customer's numeric order ID.
    """
    return lookup_status(order_id)  # your DB/API call

agent = Agent(name="Support", instructions="Help customers.",
              tools=[get_order_status])
```

The model autonomously decides *if*, *which*, and *in what order* to call tools — and can chain calls (feed one tool's output into the next). Tool choice is non-deterministic by design. See `references/tools-and-function-calling.md` for Pydantic inputs, `tool_choice`, `tool_use_behavior`, hosted tools (web search, file search, code interpreter), agents-as-tools, and MCP.

### Handoffs vs. agent-as-tool
Two ways for agents to collaborate, and they are not the same:
- **Handoff** (`handoffs=[other_agent]`): a *transfer of control*. The receiving agent takes over the conversation with full context; the original steps out. Use for routing into a domain another agent owns.
- **Agent-as-tool** (`other_agent.as_tool(...)`): a *call and return*. The orchestrator stays in charge, calls the sub-agent for one subtask, and continues. Use when you need central control and synthesis.

See `references/multi-agent-and-handoffs.md`.

## Workflow: how to build an agent

1. **Decide if you should build it at all.** Agents shine on ambiguous, multi-step, adaptive tasks; they are slower, costlier (often 10–100× a deterministic pipeline), and non-deterministic. If the task is fixed, one-shot, and well-specified, write normal code. Empirical study of abandoned AI projects shows the dominant killers are usually *not* ethics but **resource constraints, development-lifecycle challenges, and organizational dynamics** — undefined success criteria, insufficient data/ground truth, integration difficulty, cost-to-maintain. Name your success criteria and your kill criteria up front. See `references/agent-anatomy.md`.

2. **Pick the model per agent.** Match capability to the job; don't default everything to the biggest model. Cheap model for triage/guardrails, stronger model for reasoning-heavy steps. Tune `ModelSettings` (`temperature`, `max_tokens`) for tone and length. If an agent's skills come from *your own* fine-tunes, also decide how to supply them — merge specialists into one model, route between them, or fine-tune fresh. See `references/model-merging-and-model-ops.md`.

3. **Write the instructions.** State role, scope, when to use which tool, when to hand off, and what *not* to do. This is the highest-leverage knob.

4. **Define tools.** One clear responsibility each; precise name + docstring + type hints. Use Pydantic models for complex/nested inputs (you get validation for free). Push real computation into tools — never let the LLM do arithmetic or invent live data. Use OpenAI-hosted tools (web/file search, code interpreter) instead of reinventing them.

5. **Add memory if conversational.** Use `Sessions` for multi-turn; `SQLiteSession(id, db_path=...)` to persist across restarts; structured (tool-based) memory to store only key facts. Manage context growth with sliding windows or summarization. See `references/memory-and-sessions.md`.

6. **Compose multiple agents if needed.** Start with a centralized triage agent routing to specialists via handoffs (the most common, debuggable pattern). Reach for decentralized/swarm patterns only when creativity/parallel exploration justifies the coordination cost. See `references/multi-agent-and-handoffs.md`.

7. **Add guardrails.** Input guardrails screen prompts *before* expensive work (cheap classifier agent → tripwire). Output guardrails validate responses before they reach the user (schema/PII/relevance). See `references/guardrails-and-safety.md`.

8. **Instrument and evaluate.** Tracing is on by default — group runs with `with trace("name"):`, add `custom_span`s for timing. Write end-to-end tests (LLM-as-judge for non-determinism) and unit tests that assert tool calls / handoffs fired by inspecting `result.new_items`. See `references/tracing-and-evaluation.md`.

9. **Iterate from traces.** When behavior is wrong, open the trace, find the failing step, and fix the *root cause* — usually instructions or a tool description, not the loop.

For ready-to-adapt blueprints (research, coding, support, data, RAG-backed, workflow agents), see `references/agent-cookbook.md`.

## Common pitfalls & anti-patterns

- **Building an agent when a script would do.** Non-determinism, latency, and cost are real. Default to deterministic code for fixed pipelines; use agents for genuine ambiguity. (This is the single most common and most expensive mistake.)
- **Letting the LLM compute or recall live facts.** LLMs hallucinate arithmetic and stale data. Put math in a tool or the code interpreter; put live/proprietary data behind a tool. Consider `tool_choice="required"` when an answer *must* come from a trusted source.
- **Vague tool names/docstrings.** The model selects tools purely from metadata. "get_data" with no docstring will be mis-called or ignored. Be specific; use `name_override`/`description_override` when the function name is too generic.
- **No turn cap.** Always rely on / set `max_turns`. A bad config or unsolvable task otherwise spins forever, burning tokens.
- **Unbounded conversation history.** Appending every message eventually blows the context window and inflates cost/latency. Use sessions + sliding window or summarization.
- **Confusing handoff with agent-as-tool.** Handoff = give up control; as-tool = call and return. Choosing wrong breaks either oversight or modularity.
- **Input guardrails only fire on the first agent.** They gate the *entry point* of a multi-agent system, not every downstream agent. Plan placement accordingly.
- **Skipping guardrails on autonomous/expensive systems.** Any non-deterministic system that spends money or touches sensitive data needs validation. A cheap guardrail agent pays for itself by short-circuiting junk before the expensive pipeline.
- **Treating "the agent answered once" as tested.** Same input → different output. Test with LLM-as-judge over scenarios, and unit-test intermediate steps (tool fired? handoff happened?).
- **Naive auth/SQL inside tools.** Don't interpolate keys into SQL or trust LLM-built queries blind. Use parameterized queries, real authz (OAuth/RBAC), and least-privilege.
- **Trusting tool-retrieval recall as a proxy for real selection.** Over a large catalog, high recall on verbose/in-distribution benchmark queries can collapse 50+ points on the short, intent-focused queries real users type — and a model that retrieves well may not actually *know* the tool. Evaluate on realistic, paraphrased queries plus a factual probe, and curate non-overlapping tool descriptions. See `references/tools-and-function-calling.md`.
- **User-driven sycophancy in diagnosis.** Agents adopt the cause a user suggests and answer before ruling out alternatives — models challenge a misleading hypothesis spontaneously in only ~1–2 of 30 cases. For troubleshooting/support/debugging, run an evidence-first loop: competing hypotheses, discriminating questions, commit only when evidence is decisive. See `references/agent-reasoning-patterns.md`.

## Reference files

- **`references/agent-anatomy.md`** — Read when designing an agent from scratch, choosing models, or deciding build-vs-not-build. Covers the four components, the loop pseudocode, design patterns (ReAct, CoT, planner-executor, hierarchical), and the abandonment-factor checklist.
- **`references/agent-reasoning-patterns.md`** — Read when building an agent that diagnoses problems from incomplete, ambiguous, or user-biased input (support, troubleshooting, debugging, triage). Covers user-driven sycophancy, evidence-first reasoning, the hypothesis-tracking investigation loop, clarification budgets, and separating semantic reasoning from deterministic control.
- **`references/openai-agents-sdk.md`** — Read when setting up the SDK, wiring `Agent`/`Runner`, going async, integrating third-party models via LiteLLM, or using local/run context. The hands-on API reference.
- **`references/tools-and-function-calling.md`** — Read when writing tools, structured outputs, Pydantic inputs, controlling `tool_choice`/`tool_use_behavior`, chaining tools, using hosted tools, agents-as-tools, MCP servers, or doing tool retrieval over a large catalog (embedding vs. parametric retrieval, and auditing whether the model actually knows its tools).
- **`references/multi-agent-and-handoffs.md`** — Read when building multi-agent systems: orchestration types, handoff setup/customization, multi-agent switching, and the centralized/hierarchical/decentralized/swarm patterns.
- **`references/memory-and-sessions.md`** — Read when adding short-term or long-term memory, `Sessions`/`SQLiteSession`, structured memory recall, context-window management, or RAG-as-a-tool.
- **`references/guardrails-and-safety.md`** — Read when adding input/output guardrails, tripwires, relevance/safety/PII checks, and graceful failure handling.
- **`references/tracing-and-evaluation.md`** — Read when instrumenting traces/spans, debugging via the Traces UI, disabling tracing for compliance, or writing end-to-end and unit tests for non-deterministic agents.
- **`references/agent-cookbook.md`** — Read when you need a concrete blueprint for a specific agent type: research, coding, customer support, data/analytics, RAG-backed Q&A, or workflow-automation agents, plus the cognitive-architecture patterns (decision-making, planning, memory-augmented).
- **`references/model-merging-and-model-ops.md`** — Read when deciding how to *supply* an agent's model capabilities from your own fine-tunes: model merging and task arithmetic (task vectors, static vs. dynamic merging), learnable task-vector compression for cheap on-the-fly merging (Auto-FlexSwitch), and the merge-vs-route-vs-fine-tune decision when several task skills must power one agent or fleet.
