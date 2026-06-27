# Tracing, Observability & Evaluation

You cannot trust or debug an agent — especially a multi-agent one — without seeing what it did
internally, and you cannot ship one without a way to evaluate non-deterministic behavior. Turn
both on early.

## Table of contents
- Traces and spans
- Automatic tracing
- Custom traces and spans
- Grouping runs; distributed traces
- Observability backends
- Disabling tracing
- Visualization
- Evaluation: end-to-end (LLM-as-judge)
- Evaluation: unit testing (assert tool calls / handoffs)

## Traces and spans

- **Trace** — one full execution flow of an agent run, start to finish, for a given input. The
  play-by-play of everything that happened.
- **Span** — a single operation within a trace (a model call, a tool call, a handoff, a
  guardrail check), each with start/end times. Spans nest and carry debug data.

This model lets you see *which step took how long* — e.g. of a 3.2s run, 1.5s was the model,
0.5s a DB tool call — and exactly which decisions, tool args/outputs, and handoffs occurred.

## Automatic tracing

Tracing is **on by default** in the OpenAI Agents SDK — every run is recorded with no extra
code, viewable in the OpenAI Traces dashboard (platform → Dashboard → Traces). It captures the
initial input + instructions, the model's reasoning, tool calls (args + outputs), handoffs
(with context), guardrail tripwires, and the final response.

```python
from agents import Agent, Runner
agent = Agent(name="QA", instructions="Answer concisely.")
Runner.run_sync(agent, "Where is the Eiffel Tower?")   # already traced — open the dashboard
```

## Custom traces and spans

Name a trace to find it easily, and add custom spans to measure/segment your own steps:

```python
from agents import Agent, Runner, trace, custom_span
import time

agent = Agent(name="QA", instructions="Answer concisely.")
with trace("Nightly Workflow"):           # named trace
    with custom_span("Fetch"):
        time.sleep(1)
    with custom_span("Answer"):
        Runner.run_sync(agent, "Where is the Eiffel Tower?")
```

Anything inside the `trace()` context is logged under it. Custom spans break a complex workflow
into measurable steps so you can pinpoint bottlenecks and errors. Spans nest — group "research"
and "text-generation" sub-steps separately.

## Grouping runs; distributed traces

By default each `Runner.run` is its own trace. Wrap several runs in one `trace()` to treat them
as a single workflow:

```python
with trace("Pipeline"):
    a = Runner.run_sync(agent, "Step 1")
    b = Runner.run_sync(agent, "Step 2")   # both under one trace
```

To stitch runs across separate processes/machines/times into one trace, pass a shared
`trace_id`:

```python
with trace("Pipeline", trace_id="A1B2C3"):
    Runner.run_sync(agent, "Step run elsewhere")
```

Useful for long-running or distributed workflows whose pieces execute at different times.

## Observability backends

Traces default to the OpenAI Traces UI but can be exported to your own telemetry stack
(Datadog, Azure Monitor Logs, etc.). In broader agent practice, teams also wire metrics to
Prometheus/Grafana/LangSmith and track: task-success rate, average response time, tool
invocation latency, fallback frequency (how often the agent defers/fails), and hallucination
rate. A drop in tool-success or a spike in hallucinations should trigger alerts → human review
→ prompt/tooling fixes.

## Disabling tracing

For regulatory reasons or sensitive material you don't want logged, disable it:

```python
import os
os.environ["OPENAI_AGENTS_DISABLE_TRACING"] = "1"
```

Only disable for genuine compliance/PII reasons — you lose your primary debugging tool.

## Visualization

`draw_graph(agent, filename="graph")` (with `openai-agents[viz]`) renders the static structure
— agents as boxes, tools as ellipses, solid arrows = handoffs, dotted = tool calls. Use it to
confirm the system is wired as intended and as documentation. (See `openai-agents-sdk.md`.)

## Evaluation: end-to-end (LLM-as-judge)

Agents are non-deterministic — the same input can yield different outputs — so exact-match
testing fails. End-to-end evaluation asks: does the full system produce a *desirable* output?
Two approaches: a human verifies, or (automatable) an LLM/agent judges output against an
expected result.

```python
from pydantic import BaseModel
from agents import Agent, Runner

class Scenario(BaseModel):
    name: str; input: str; expected: str

class Verdict(BaseModel):
    passed: bool

judge = Agent(name="Judge",
    instructions="Decide whether the agent output satisfies the expected output.",
    output_type=Verdict)

scenarios = [
    Scenario(name="delayed", input="Why hasn't order 200 arrived?",
             expected="The order is delayed"),
    Scenario(name="missing", input="Status of order 400?",
             expected="No such order found"),
]
for s in scenarios:
    actual = Runner.run_sync(your_agent, s.input).final_output
    verdict = Runner.run_sync(judge, f"Output: {actual} ||| Expected: {s.expected}")
    print(s.name, verdict.final_output.passed)
```

Re-run the suite after any change to the agent. Failures point you at instructions, tool
descriptions, or tool implementations to fix.

## Evaluation: unit testing (assert tool calls / handoffs)

Don't only check final text — assert the agent did the right *intermediate* things. Inspect
`result.new_items` for the tool calls and handoffs that fired.

```python
from agents import Runner, ToolCallItem

result = Runner.run_sync(agent, "Status of order 101?")
tool_calls = [it.raw_item.name for it in result.new_items if isinstance(it, ToolCallItem)]
assert "get_order_status" in tool_calls, "expected the order-status tool to be called"
```

This pinpoints failures (didn't call the tool → prompt/description issue; wrong handoff →
routing issue) far better than judging only the final answer. Combine: LLM-as-judge for
output quality across scenarios, unit tests for the control flow (right tool, right handoff,
right guardrail fired). Wire both into CI so regressions surface early.
