# Agentic Orchestration Patterns 📐🤖

Orchestration patterns define how agents interact with tools, review systems, and other agents to produce high-quality work. This repository supports four primary design patterns.

---

## 🗺️ Orchestration Guides

Click on any guide below to view the detailed implementation rules, visual flowcharts, and state contracts:

*   **[Hill-Climb Refinement Loop](hill_climb.md)**: A sequential, state-persisted verification loop designed to incrementally improve a candidate solution based on test/validation feedback. Enforces execution budgets and plateau constraints.
*   **[Fan-Out / Best-of-N](fan_out.md)**: A parallel execution pattern that generates multiple distinct candidates concurrently and selects the best candidate using LLM judging or consensus voting.
*   **[Adversary / Red-Team](adversary.md)**: A multi-agent debate workflow separating concerns between a Generator (builder), an Adversary (critic), and an Arbitrator (judge) to stress-test safety and security.
*   **[Tree of Thoughts (ToT)](tree_of_thoughts.md)**: A branching search pattern that allows agents to evaluate intermediate reasoning steps, backtrack when a path fails, and systematically explore alternatives using BFS/DFS.

---

## 📊 Selection Framework

Use this comparison matrix to select the correct orchestration pattern for your specific task:

| Feature / Trait | [Hill-Climb](hill_climb.md) | [Fan-Out](fan_out.md) | [Adversary](adversary.md) | [Tree of Thoughts](tree_of_thoughts.md) |
| :--- | :--- | :--- | :--- | :--- |
| **Execution Style** | Sequential (Iterative) | Parallel (Concurrent) | Alternating (Debate) | Branching (Tree Search) |
| **Latency Profile** | High (Multi-turn wait) | Low (Single-turn wait) | High (Multi-turn debate) | Very High (Tree exploration) |
| **API Token Cost** | Moderate | High (Proportional to $N$) | Moderate-to-High | Very High |
| **Primary Strength** | Incremental correction | Broad solution coverage | Rigorous security/safety | Complex logical paths |
| **Implementation Complexity** | Moderate | Low | High | Very High |
| **Best Fit** | Code bug-fixing & linting | Content creation & voting | High-risk system logic | Math proofs & algorithms |
