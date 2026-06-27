# Tracing, Observability & Evaluation 📊🔍

You cannot trust or debug an agent—especially a multi-agent system—without seeing what it did internally, and you cannot ship one without a way to evaluate non-deterministic behavior. This document covers tracing execution flows and setting up basic evaluation.

---

## 1. Traces and Spans

*   **Trace:** One full execution flow of an agent run, start to finish, for a given input. It is the comprehensive play-by-play of everything that happened.
*   **Span:** A single operation within a trace (a model call, a tool call, a handoff, a guardrail check), each with start and end times. Spans nest and carry debug data.

This model lets you see *which step took how long* (e.g., of a 3.2s run, 1.5s was the model, 0.5s a database tool call) and exactly which decisions, tool arguments/outputs, and handoffs occurred.

---

## 2. Automatic Tracing

Tracing is typically **on by default** in standard Agent SDKs—every run is recorded with no extra code, viewable in traces dashboards. It captures the initial input + instructions, the model's reasoning, tool calls (arguments + outputs), handoffs (with context), guardrail tripwires, and the final response.

```python
from agents import Agent, Runner
agent = Agent(name="QA", instructions="Answer concisely.")
Runner.run_sync(agent, "Where is the Eiffel Tower?")   # Automatically traced
```

---

## 3. Custom Traces and Spans

You can name a trace to find it easily, and add custom spans to measure or segment your own steps:

```python
from agents import Agent, Runner, trace, custom_span
import time

agent = Agent(name="QA", instructions="Answer concisely.")
with trace("Nightly Workflow"):           # Named trace
    with custom_span("Fetch"):
        time.sleep(1)
    with custom_span("Answer"):
        Runner.run_sync(agent, "Where is the Eiffel Tower?")
```

Anything inside the `trace()` context is logged under it. Custom spans break a complex workflow into measurable steps so you can pinpoint bottlenecks and errors. Spans can nest—allowing you to group "research" and "text-generation" sub-steps separately.

---

## 4. Distributed Traces

By default each `Runner.run` is its own trace. Wrap several runs in one `trace()` to treat them as a single workflow:

```python
with trace("Pipeline"):
    a = Runner.run_sync(agent, "Step 1")
    b = Runner.run_sync(agent, "Step 2")   # Both under one trace
```

To stitch runs across separate processes, machines, or times into one trace, pass a shared `trace_id`:

```python
with trace("Pipeline", trace_id="A1B2C3"):
    Runner.run_sync(agent, "Step run elsewhere")
```

This is useful for long-running or distributed workflows whose pieces execute at different times.

---

## 5. Evaluation: End-to-End (LLM-as-Judge)

Agents are non-deterministic—the same input can yield different outputs—so exact-match testing fails. End-to-end evaluation asks: does the full system produce a *desirable* output? An LLM/agent can judge output against an expected result.

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

Re-run the suite after any change to the agent. Failures point you at instructions, tool descriptions, or tool implementations to fix.

---

## 6. Evaluation: Unit Testing (Asserting Control Flow)

Don't only check final text—assert the agent did the right *intermediate* things. Inspect `result.new_items` for the tool calls and handoffs that fired.

```python
from agents import Runner, ToolCallItem

result = Runner.run_sync(agent, "Status of order 101?")
tool_calls = [it.raw_item.name for it in result.new_items if isinstance(it, ToolCallItem)]
assert "get_order_status" in tool_calls, "expected the order-status tool to be called"
```

This pinpoints failures far better than judging only the final answer. Combine: 
*   **LLM-as-judge** for output quality across scenarios.
*   **Unit tests** for the control flow (right tool, right handoff, right guardrail fired).
