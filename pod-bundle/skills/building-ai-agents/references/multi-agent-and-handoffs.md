# Multi-Agent Systems & Handoffs

When one agent gets overloaded — too many tools, too many domains — split it into a team.
A router/triage agent plus specialists almost always beats one mega-agent, for the same
reason a company has departments. This file covers how agents collaborate and the
architectural patterns for organizing them.

> For deep, recursive, long-horizon orchestration (hierarchical controllers, agent-environment
> engineering, self-spawning sub-agent harnesses), see the **llm-agent-orchestration** skill.
> This file covers the practical multi-agent patterns you build directly with the SDK.

## Table of contents
- Handoff vs. agent-as-tool (the core distinction)
- Defining handoffs
- Multi-agent switching (keep control with the last agent)
- Customizing handoffs (callbacks, input filters, overrides)
- Handoff prompting (RECOMMENDED_PROMPT_PREFIX)
- Orchestration: deterministic vs. dynamic
- Architectural patterns: centralized, hierarchical, decentralized, swarm
- Choosing a pattern

## Handoff vs. agent-as-tool (the core distinction)

Two mechanisms, fundamentally different in who stays in charge:

| | **Handoff** | **Agent-as-tool** |
|---|---|---|
| Control | Full **transfer** — new agent owns the rest of the turn | Orchestrator **keeps** control, calls sub-agent like a function |
| What passes | Full conversation history/context | A scoped sub-request; result returns to the caller |
| Analogy | Transferred to another department | Put on hold while the rep asks a colleague, then resumes |
| Use when | The work belongs in another agent's domain; that agent should own the user interaction; tight oversight of intermediate steps isn't needed | You need central control, must synthesize multiple workers' outputs, or need maximum oversight |

Both are supported and can be combined. `as_tool()` is covered in
`tools-and-function-calling.md`; handoffs are below.

## Defining handoffs

Add a `handoffs=` list of agents the agent may transfer to. The SDK exposes each sub-agent's
name and instructions to the routing agent so it can decide when to delegate.

```python
from agents import Agent, Runner

complaints_agent = Agent(name="Complaints Agent",
    instructions="Handle complaints with empathy and clear next steps.")
inquiry_agent = Agent(name="General Inquiry Agent",
    instructions="Answer general questions promptly.")

triage_agent = Agent(
    name="Triage Agent",
    instructions="Triage the user's request and route to the appropriate agent.",
    handoffs=[complaints_agent, inquiry_agent],
)
print(Runner.run_sync(triage_agent, "My meal is too hot").final_output)
```

The triage agent reasons about intent (no keyword matching) and transfers control to the
right specialist, which then fully owns the response.

## Multi-agent switching (keep control with the last agent)

By default a sub-agent that receives a handoff is then "stuck" — it can't transfer onward. To
build a fully dynamic network where any agent can route to any other (including back to
triage), give every agent handoffs to the others and continue the loop with
`result.last_agent`:

```python
from agents import Agent, Runner, SQLiteSession, trace

complaints_agent = Agent(name="Complaints Agent", instructions="...")
sales_agent      = Agent(name="Sales Agent", instructions="...")
triage_agent     = Agent(name="Triage Agent", instructions="...")

# wire mutual handoffs
complaints_agent.handoffs = [sales_agent, triage_agent]
sales_agent.handoffs      = [complaints_agent, triage_agent]
triage_agent.handoffs     = [complaints_agent, sales_agent]

session = SQLiteSession("convo")
last_agent = triage_agent
with trace("Multi-agent system"):
    while True:
        q = input("You: ")
        result = Runner.run_sync(last_agent, q, session=session)
        print("Agent:", result.final_output)
        last_agent = result.last_agent   # continue with whoever is now active
```

The `session` keeps shared state across turns; `last_agent` ensures the next user message
goes to the currently active agent, not always back to triage.

## Customizing handoffs (callbacks, input filters, overrides)

Use the `handoff()` helper instead of a bare agent to attach behavior:

```python
from agents import handoff
from pydantic import BaseModel

class HandoffReason(BaseModel):
    target_agent: str

def on_transfer(ctx, data: HandoffReason):
    print(f"Transferring to: {data.target_agent}")   # logging/analytics/notify user

complaints_handoff = handoff(
    agent=complaints_agent,
    on_handoff=on_transfer,          # fires when the handoff happens
    input_type=HandoffReason,        # structured input passed to the callback
    # tool_name_override / tool_description_override: relabel in traces
    # input_filter: trim/summarize the history passed to the next agent
)
triage_agent.handoffs = [complaints_handoff]
```

- `on_handoff` — side-effect hook (logging, user notification, auditing).
- `input_type` — structured payload the SDK validates and passes to the callback.
- `input_filter` — reshape the context handed over (e.g. summarize, keep last N messages).
- `tool_name_override` / `tool_description_override` — change the trace label.

By default a handoff passes the **full conversation history** to the new agent — the user
need not repeat anything.

## Handoff prompting (RECOMMENDED_PROMPT_PREFIX)

Handoffs are only as good as each agent's instructions. The SDK ships a recommended prefix
that tells the model it's in a multi-agent system and that transfers are seamless/invisible
to the user:

```python
from agents.extensions.handoff_prompt import RECOMMENDED_PROMPT_PREFIX

triage_agent = Agent(
    name="Triage Agent",
    instructions=f"{RECOMMENDED_PROMPT_PREFIX} Triage the request and route appropriately.",
    handoffs=[complaints_agent, inquiry_agent],
)
```

Also give explicit routing rules ("if the user asks about sales, route to the sales agent")
and clear per-agent purpose statements so the router can choose well.

## Orchestration: deterministic vs. dynamic

How the flow between agents is decided:

- **Deterministic orchestration** — you hardcode the routing in Python (`if/else`, fixed
  sequence). Predictable, testable, auditable, cost-knowable. But brittle: misroutes anything
  phrased outside the rules, and every new case needs a code change. Use for stable, repeatable
  pipelines where you must guarantee the path.
  ```python
  def orchestrate(msg):
      agent = complaints_agent if "complaint" in msg.lower() else inquiry_agent
      return Runner.run_sync(agent, msg).final_output   # misses "my meal is too hot"
  ```
- **Dynamic orchestration** — an LLM agent decides routing at runtime (triage agent +
  handoffs or as-tool). Flexible, handles unseen phrasings and multi-intent requests, but the
  path (and cost/latency) is less predictable.

Most real systems blend both: deterministic where you need guarantees, dynamic where you need
to absorb ambiguity.

## Architectural patterns

### Centralized (most common)
One manager/orchestrator/triage agent routes to specialist agents, each an expert in one
function. Clear separation of concerns; easy to add specialists. Weakness: the whole system is
only as good as the router, and specialists are siloed (can't talk to each other). Best for
top-down flows: support triage, internal IT/HR/facilities helpdesk.

### Hierarchical (subset of centralized)
Multiple tiers: a top triage agent → manager agents → their own specialist sub-agents (CEO →
C-suite → teams). Excels at large, multi-domain tasks by recursive decomposition; promotes
reuse and focused scope. Cost: more layers = more latency, cost, failure surface, and harder
debugging. Best for deep research and complex multi-stage queries. Don't over-decompose simple
tasks.

```python
physics_agent  = Agent(name="Physics Agent",  instructions="Answer physics questions.")
warfare_agent  = Agent(name="Warfare Agent",  instructions="Answer military-history questions.")
science_manager = Agent(name="Science Manager",
    instructions="Route science queries to the right sub-agent.",
    handoffs=[physics_agent, ...])
history_manager = Agent(name="History Manager",
    instructions="Route history queries to the right sub-agent.",
    handoffs=[warfare_agent, ...])
triage = Agent(name="Research Triage",
    instructions="Decide science vs. history and route.",
    handoffs=[science_manager, history_manager])
```

### Decentralized
No central router — peer agents collaborate directly, all able to communicate (a roundtable,
not a hierarchy). Great for creativity, debate, brainstorming, negotiation. Weakness: no
built-in coordination — you must drive the conversation flow yourself with deterministic code
(the SDK doesn't auto-orchestrate this). Example: a landlord agent and tenant agent debate for
N rounds, then a summarizer agent consolidates.

### Swarm (subset of decentralized)
Many simple agents work in parallel; intelligence is an *emergent* property of the collective
(like cells forming an organism, or a random forest of weak learners). Scalable, no single
point of failure, good for exploration/optimization (generate many candidates, then select).
Weakness: coordination/cost overhead at scale, and emergence isn't guaranteed. Example: 10
role-playing agents each propose ideas in parallel; an aggregator synthesizes a final plan.

```python
import concurrent.futures
roles = ["Urban Planner", "Artist", "Engineer", "Doctor", ...]
agents = [Agent(name=f"{r} Agent", instructions=f"As a {r}, answer: ...") for r in roles]
# run in parallel, then feed all outputs to an aggregator agent
```

## Choosing a pattern

- Start **centralized** (triage → specialists). It's the most debuggable and covers most apps.
- Go **hierarchical** only when a domain genuinely needs sub-decomposition (deep research).
- Use **decentralized/swarm** only when creativity/parallel exploration justifies the
  coordination cost and reduced predictability.
- Prefer **handoff** when another agent should own the interaction; **agent-as-tool** when you
  must keep control and synthesize. Visualize the wiring with `draw_graph` (see
  `openai-agents-sdk.md`) to confirm it matches intent.
