# The Teams Agent Ownership Playbook 🤖📋
> [!IMPORTANT]
> **The Fast Track to Dangerous AI:** "The fastest way to make an AI agent dangerous is to let everyone use it and nobody own it." 
> When an agent reads files, drafts messages, or changes code without a clear human owner, it leads to a "haunted house" company—automated systems running on rotted policies and stale instructions.

This playbook provides a concrete operational framework to bring **Agent Owner's Cards** and an **Agent Registry** into **Microsoft Teams** and **Microsoft Loop**.

---

## 1. The Teams & Loop Registry Architecture

To make agents visible to humans, you need an **Agent Registry**. In the Microsoft ecosystem, the most effective way to build this without heavy IT overhead is using **Microsoft Loop Components** or a **Teams Channel**.

### Option A: The Microsoft Loop Table Component (Recommended)
You can create a live, collaborative registry using a Loop Table component inside a shared Teams channel or Loop workspace.

```
+-----------------------------------------------------------------------------------+
|                              TEAMS AGENT REGISTRY                                 |
+-------------+------------------+-----------------------------+--------------------+
| Agent Name  | Human Owner      | Job (One Sentence)          | Permissions/Access |
+-------------+------------------+-----------------------------+--------------------+
| @StoryPrep  | @Jane Doe (PM)   | Draft refinement packets    | Read-only (M365)   |
| @TriageBot  | @John Smith (QA) | Label and route QA tickets  | Draft-only (Jira)  |
+-------------+------------------+-----------------------------+--------------------+
```

*   **How to set it up:**
    1. In a Teams chat or channel post, type `/table` to insert a Loop table.
    2. Add columns: `Agent Name`, `Human Owner`, `Job`, `Diet (Sources)`, `Permissions`, `Review Loop Cadence`, and `Link to Owner Card`.
    3. Use **@mentions** in the `Human Owner` column. This creates a **People Chip** linking directly to the owner's Microsoft 365 Contact Card so anyone can instantly message them.
    4. Pin this Loop component to the top of your Teams channel for easy access.

### Option B: The Teams Channel "Owner Cards"
If you prefer a post-by-post format similar to a Slack channel:
1. Create a dedicated Teams Channel (e.g., `#agent-registry`).
2. Have each owner post their **Agent Owner's Card** (using the template below) as a new conversation thread.
3. Keep the thread updated with changelogs, rotted instruction alerts, or run reviews.

---

## 2. The Agent Owner's Card (Loop/Teams Template)

Copy and paste this markdown template into your Teams Channel posts or Microsoft Loop pages to document your agents:

```markdown
# 📇 Agent Owner's Card: [Agent Name]

> **The One-Sentence Ownership Test:** If this agent breaks or drifts, who is the single human responsible for the business outcome?
> **Human Owner:** @[Name of Owner] (e.g., Product Manager, Tech Lead)

---

### 📋 The Job
*   **One-Sentence Mission:** [What is this agent supposed to do? (e.g., "Draft first-pass backlog items for sprint refinement.")]
*   **Target Output:** [What work product does it create? (e.g., "A Loop page refinement packet with user stories and acceptance criteria.")]

### 🥗 The Diet (Context & Inputs)
*   **What it Reads:** [List folders, files, repositories, or APIs (e.g., "PRD folder, last 20 support tickets, 3 baseline backlog examples.")]
*   **Staleness Risk:** [How often do these inputs change? How does the agent get updated? (e.g., "Updated weekly before backlog grooming.")]

### 🚧 Boundaries (Permissions)
*   **Read Access:** [e.g., "Read-only access to SharePoint onboarding folder."]
*   **Write/Action Access:** [e.g., "Draft-only. The agent CANNOT create Jira tickets directly; it must output to a draft review page."]
*   **Tooling/Execution:** [e.g., "Can execute local tests in sandboxed environment, cannot push to main branch."]

### 🔄 The Review Loop
*   **Human Reviewer:** @[Reviewer Name]
*   **Review Cadence:** [e.g., "Weekly during backlog grooming, post-sprint retrospective check."]
*   **Failure Modes to Watch For:**
    1. *Stale Context:* Relying on old design briefs if the Figma link changes.
    2. *Tone Drift:* Generating dry, overly-formal text instead of the team's voice.
    3. *Hallucinated Dependencies:* Listing nonexistent APIs for backlog items.
```

---

## 3. The Self-Documentation Prompt

Point an existing agent (such as your GPT, Claude, or a custom script) at this prompt to have it draft its own **Owner's Card** fields. It will fill in what it knows from its context and prompt instructions, then ask you for the human-centric parts.

```markdown
You are an AI agent currently executing a workflow. To ensure operational safety and alignment, I need to generate your "Agent Owner's Card."

Please read your current system instructions, active files, and configuration, then draft the following fields:

1. **Job (One-Sentence Mission):** Based on your instructions, summarize your exact operational job in one clear sentence.
2. **Diet (What you read):** List the files, context, parameters, and databases you have access to or read in this conversation. Highlight which ones might go stale.
3. **Boundaries (What you can and cannot do):** List your current tool capabilities (e.g., file reads, terminal commands, web search) and what actions you are strictly restricted from taking (or require manual approval for).
4. **Draft Failure Modes:** Based on your limits and instructions, what are 2 or 3 ways you could fail, drift, or produce "polished noise" if your inputs or system rules become outdated?

Leave the "Human Owner" and "Review Loop Cadence" fields blank or marked as [TO BE COMPLETED BY HUMAN] so I can finalize them. Output the result in standard markdown using the "Agent Owner's Card" structure.
```

---

## 4. Care and Feeding: The Operational Rules

To prevent your Microsoft Teams registry from becoming a graveyard of dead cards:

> [!TIP]
> **Start Read/Draft Only:** If you are unsure of an agent's boundaries, restrict it to `Read-Only` or `Draft-Only`. Let it earn write/send permissions by demonstrating consistency in the review loop.

> [!WARNING]
> **Decommission Rotted Agents:** If an agent's owner leaves the team, or if the review loop hasn't run in 30 days, decommission the agent or pause its execution until a new owner accepts the card.
