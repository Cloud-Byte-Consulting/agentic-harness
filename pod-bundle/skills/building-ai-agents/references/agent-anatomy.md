# Agent Anatomy, the Loop, and Build-vs-Not

Mental model and design decisions for an LLM agent, before you write SDK code.

## Table of contents
- What an agent is (and is not)
- The four components
- The agent loop
- Cognitive architectures: decision-making, planning, memory-augmented
- Design patterns: CoT, ReAct, planner-executor, hierarchical
- Choosing the model
- Should you build it at all? (abandonment-factor checklist)

## What an agent is (and is not)

An **AI agent** is an intelligent system that pursues a goal by perceiving context, reasoning about what to do, and taking actions through tools — autonomously and iteratively. The defining traits are: reasoning from an ambiguous goal, forming/adapting a plan, and acting via tools in a loop.

Contrast with traditional software, which is deterministic (fixed if-X-then-Y rules) and stalls on anything outside its script. An agent can adapt, substitute, and recover. The trade: agents hallucinate, take unexpected paths, are slow, can be 10–100× costlier than the deterministic equivalent, and are hard to explain. Autonomy buys flexibility at the cost of predictability — which is exactly why guardrails and tracing exist.

A chatbot, a fraud classifier, or a RAG Q&A endpoint is **not** automatically an agent. The loop + autonomous tool/action selection is what makes it agentic.

## The four components

1. **Model (the brain).** A general-purpose pretrained LLM, instructed to behave like an agent. The same base model (e.g. GPT-4o) powers wildly different agents — what differs is instructions, tools, and wiring, not the weights. The model interprets input, plans, generates tool arguments, reviews results, and iterates.

2. **Instructions (the system prompt).** Defines identity, role, objectives, behavior, and persona. This is where most agent debugging happens — when an agent misbehaves, tune the prompt first. Be explicit about scope, tool-use conditions, handoff conditions, and prohibitions.

3. **Tools (hands and eyes).** Let the agent act on the world: send email, query a DB, search the web, run code, call another agent. Each tool exposes a name, description, and parameter schema to the model; the runtime executes the call (client-side function or server-side API) and feeds the result back. Tool design — granularity, descriptions, error handling — is a first-class engineering task.

4. **Memory & knowledge (the notebook).**
   - **Working memory:** conversation history kept in the prompt (enables follow-up questions). Bounded by the context window.
   - **Long-term memory:** facts persisted across sessions (turns the agent from stateless to stateful).
   - **Training knowledge:** baked into model weights; broad but frozen at the cutoff and not proprietary. Often you deliberately instruct the agent to *ignore* it and rely on retrieval.
   - **Retrieved knowledge (RAG):** dynamic, real-time facts pulled from DBs/docs/web via tools, grounded and citable.

## The agent loop

The runtime ("agent loop" / Runner) drives this cycle:

```
Read the user's goal and form an action plan
For each step:
    Create the tool/action inputs
    Execute the action
    Get the result
    Append the result to context (memory)
    Revise the plan if needed
    If the goal is achieved: return the final output
(stop early at a max-turn cap)
```

Concretely each turn: the runtime sends the current message list to the model; the model returns either a **final answer** (loop ends) or a **tool call / handoff** (runtime executes it, appends the output, loops). This loop logic is the historically hard part of agent building that frameworks abstract away. Always bound it with a max-turn cap.

## Cognitive architectures

Three reusable building blocks (combine them):

- **Autonomous decision-making agent** — perceive → reason → act → learn in real time. Best for reactive, single-step responses (triage, customer service, alerting). Risk: speed over depth, hallucination without constraints.
- **Planning agent** — decomposes a high-level goal into a task tree, sequences/allocates, monitors progress, and revises the plan dynamically. Best for long-horizon, multi-step projects (campaign launch, workflow orchestration). Uses hierarchical decomposition; symbolic planning (STRIPS/PDDL) for well-defined domains, LLM-powered dynamic planning (Tree-of-Thought, Self-Ask) for novel ones. Slower; needs control logic.
- **Memory-augmented agent** — adds working + episodic (timestamped past interactions, vector-searched) + semantic (factual knowledge base) memory for continuity and personalization. Best for assistants, CRM, anything spanning sessions. Cost: retrieval design and memory maintenance.

These realize the same perceive/reason/act/learn loop; pick which capabilities a given task needs.

## Design patterns (control logic)

- **Chain-of-Thought (CoT):** model produces step-by-step reasoning before answering. Can't act or adapt mid-plan.
- **ReAct (Reason + Act):** iteratively reason → pick a tool → observe result → adjust. The default for real-world tool-using agents. (The OpenAI Agents SDK blends ReAct with hierarchical/multi-agent.)
- **Planner-executor:** one agent plans, others execute steps. Good for long, modular tasks.
- **Hierarchical / multi-agent:** a manager delegates to specialized sub-agents (like a CEO → C-suite → teams). Good for complex, multi-domain work; can parallelize.

These are not exclusive — production agents blend them.

For agents that diagnose problems from **incomplete or user-biased input** (support, troubleshooting, debugging, triage), add an **evidence-first** layer on top of these patterns: treat any cause the user suggests as one hypothesis to test, not a fact to reason from. By default models exhibit *user-driven sycophancy* — adopting the user's guess and committing early. See `agent-reasoning-patterns.md`.

## Choosing the model

Per-agent, match capability to need. Axes:
- **Cost** — $/token; foundation models are far cheaper than heavy reasoning models.
- **Latency** — reasoning models can be 8× slower (and ~10× costlier) than a fast chat model for the same prompt.
- **Performance** — coding vs. creative vs. multimodal vs. context-window size; pick per task.
- **Bias & cutoff** — knowledge cutoffs and leanings; mitigate with RAG/prompting.

Pattern: cheap fast model for triage and guardrails, stronger model for reasoning/math/research steps, possibly a writing-tuned model for content. The SDK lets every agent use a different model and `ModelSettings` (`temperature`, `max_tokens`).

## Should you build it at all?

Before committing, sanity-check against the factors that actually kill AI projects. Empirical analysis of abandoned AI systems finds that while ethical concerns dominate *publicly reported* incidents (mostly post-deployment), practitioner-reported abandonment — especially *pre*-deployment — is driven far more by non-ethics factors. A large share of enterprise GenAI pilots never reach production, so treat "build" as a decision to justify, not a default. Use this as a pre-build checklist:

- **Resource constraints** — too expensive to build/maintain; compute/GPU availability and cost; timeline too long; lacking in-house expertise; cheaper/easier to outsource or buy off-the-shelf. *Often the most salient real lever, even when ethics concerns also exist.*
- **Development-lifecycle challenges** — can you even define and measure the target/success criteria? Is there ground truth? Can you collect/label/curate enough data? Will it integrate into existing workflows? Will a pilot get adopted? Undefined success criteria and integration failure are repeat offenders.
- **Organizational dynamics** — executive sponsor + internal champion present? Aligned with strategy? Did the client actually want it, or just want "AI"? Shifting priorities?
- **Ethical concerns** — discrimination, privacy/surveillance, misinformation, misuse, over-reliance, labor displacement, environmental cost.
- **Stakeholder feedback** — will employees, users, affected communities, or domain experts resist?
- **Legal/regulatory** — AI/data-protection/domain regulation, liability, IP/copyright, "too high risk."

These layer and interact (e.g. insufficient budget → small dataset → poor performance). Two practical rules:
1. Make resource and lifecycle costs *explicit* early — they're a legitimate, often-decisive reason not to build, and they compound other risks.
2. Define **kill criteria** alongside success criteria, and the conditions under which an abandoned idea could be revisited (more data, better tooling, a real internal champion). Non-development is a valid, sometimes optimal outcome — abandoning a non-viable system early is cheaper than failing late.

If the task is fixed, one-shot, repetitive, and fully specifiable: a deterministic pipeline beats an agent on cost, speed, and reliability. Reserve agents for genuine ambiguity and adaptivity.
