# Agent Operations - Open Skills | Unlock AI

Generated 2.39:1 Open Skills category masthead showing categorized skill drawers on the right side. [←Skills directory](/open-skills/skills) 

# Agent Operations

Meta-skills: skills about running agents well. This is where heavy agent users get disproportionate returns.

Use this category page to compare the primitives, check their requirements, and copy the setup prompt for the smallest skill that makes the next workflow repeatable.

Pick one primitive, install it, test it, then come back when your workflow teaches it a better default.

7 skills

### Goal Prompt Generator

Transforms a fuzzy implementation plan into a bounded, autonomous objective for an agent: a goal prompt with an explicit definition of done, repo constraints (what may be touched, what must not be), verification gates the agent must pass before claiming completion, and stop conditions for when to halt and ask rather than improvise. The output is a prompt you hand to a fresh agent session — or a different agent entirely — that can be pursued without supervision and checked cleanly afterward.

Why build it

The difference between an agent that works autonomously and one that wanders is almost entirely in how the objective is specified. "Definition of done + constraints + verification gates" is the specification pattern that works, and this skill makes producing it a procedure instead of an art. It's also the natural bridge between agents: one agent plans, this skill packages, another agent executes.

What you need

Nothing

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "goal-prompt-generator", stored
wherever my harness loads skills from.

The skill's job: turn an implementation plan or task description into a bounded goal
prompt another agent session can pursue autonomously and be checked against.

The skill must include: (1) trigger conditions — when I ask you to package work for
another session, write a goal prompt, or prepare a task for autonomous execution;
(2) a required structure for every goal prompt: the objective in one paragraph; an
explicit DEFINITION OF DONE as a checklist of verifiable statements; repo constraints
(files/areas that may be modified, files/areas that must NOT be touched); verification
gates — the exact commands to run and expected results before claiming completion; and
stop conditions — situations where the agent must halt and ask instead of improvising;
(3) a self-containment rule: the receiving session has none of our conversation
context, so the prompt must include exact paths and all needed background; (4) a
quality check before delivering: "could a competent agent with zero context execute
this and could I verify the result without re-deriving the plan?"

After writing it, test it by packaging the next real task I describe into a goal
prompt.
  </task>
</prompt>
```

### Visible Delegation

Lets one agent orchestrate another while keeping the delegated work visible — running the delegate in a shared terminal session (tmux is the proven mechanism) instead of a hidden background process, so you can watch, interrupt, and course-correct in real time. The skill covers launching the delegate with a packaged goal prompt, monitoring its progress, when the orchestrator should intervene versus wait, and how results get verified on the way back.

Why build it

Multi-agent setups usually fail on supervision: hidden background agents drift for twenty minutes before anyone notices. Keeping delegation observable preserves the leverage of parallel agents without surrendering oversight — and watching how a delegated session goes wrong is also how you learn to write better goal prompts. Pairs directly with Goal Prompt Generator: one packages the work, this runs and supervises it.

What you need

tmux (or equivalent) · Two agent harnesses or two sessions of one

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "visible-delegation", stored wherever
my harness loads skills from.

The skill's job: delegate work to another agent session while keeping it visible and
supervisable — shared terminal sessions, never hidden background runs.

Before writing it, check that tmux is installed (install it if not) and confirm which
agent CLI(s) I use for delegate sessions.

The skill must include: (1) trigger conditions — when I ask you to delegate, run
something in parallel, or hand work to another agent; (2) the launch procedure: create
a named tmux session, start the delegate agent in it, and pass a goal prompt (built
with my goal-prompt-generator skill if available) — then tell me how to attach and
watch; (3) monitoring rules: check the session at sensible intervals, and define what
warrants intervention (stuck loops, scope drift, destructive commands) versus patience;
(4) a results protocol: when the delegate claims completion, run the verification
gates from the goal prompt yourself before reporting success to me; (5) cleanup —
sessions get closed, not abandoned.

After writing it, test it by delegating one small real task end to end while I watch.
  </task>
</prompt>
```

### Session Operating Map

Sets up and maintains a per-project map of your parallel agent sessions: which session/thread owns which lane of work, naming conventions so lanes are identifiable at a glance, where coordination state lives (a small repo-local map file), how blockers between lanes get recorded, and rules for archiving finished lanes and promoting durable lessons into the project's docs or skills. It's the answer to "which conversation was that in?"

Why build it

Past two or three concurrent sessions on a project, you become the bottleneck — re-explaining state, losing decisions in closed threads, duplicating work across lanes. A repo-local operating map externalizes that coordination so any session (or any future you) can read what's in flight, what's blocked, and what's been decided. This is project management for agent work, kept lightweight enough to actually maintain.

What you need

Nothing; matters once you run multiple concurrent sessions per project

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "session-operating-map", stored
wherever my harness loads skills from.

The skill's job: set up and maintain a repo-local operating map for projects where I
run multiple agent sessions in parallel — which lane owns what, current state,
blockers, and decisions.

The skill must include: (1) trigger conditions — when I start parallel workstreams in
a project, ask "what's in flight," or ask you to set up coordination for a repo;
(2) the map file: a single repo-local doc (suggest docs/operating-map.md) listing each
lane with a short name, its objective, owning session, current state, and blockers;
(3) lane discipline: one lane per concern, named so its purpose is obvious; (4) update
rules: a lane's entry gets updated when its state meaningfully changes — start, block,
handoff, done — not as a journal; (5) archive rules: finished lanes move to a done
section with a one-line outcome, and lessons worth keeping get promoted into the
project's docs or skills rather than dying with the lane; (6) a read-first rule: any
session joining the project reads the map before starting work.

After writing it, set up the map for my current project and populate it with what's
actually in flight.
  </task>
</prompt>
```

### Self-Authored PR Merge

A clean workflow for reviewing and merging pull requests you authored yourself — the daily reality of solo developers and agent-heavy workflows, which GitHub's approval model doesn't really accommodate (you can't approve your own PR). The skill runs a genuine self-review pass (diff inspection with fresh eyes, not a rubber stamp), checks CI status and mergeability, handles the merge with the right strategy, and finishes with branch and worktree-safe cleanup.

Why build it

Solo shipping needs more review discipline, not less — nobody else is going to catch the bug. Encoding the review-check-merge-cleanup sequence as a skill means it happens the same way every time, including the steps that get skipped when you're moving fast (actually reading the diff; actually deleting the branch). It also handles the practical GitHub friction honestly instead of pretending the approval model works differently than it does.

What you need

The GitHub CLI (gh) authenticated

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "self-pr-merge", stored wherever my
harness loads skills from.

The skill's job: review and merge pull requests I authored myself, with real review
discipline despite GitHub not allowing self-approval.

Before writing it, confirm the gh CLI is authenticated and ask me for my merge
strategy preference (squash, merge, rebase) and my branch cleanup preference.

The skill must include: (1) trigger conditions — when I ask to merge my own PR or
review-and-merge something I wrote; (2) a genuine review pass FIRST: read the full
diff with fresh eyes, list anything questionable (bugs, debug leftovers, missing
tests, scope creep) and show me findings before merging — finding nothing must be a
conclusion, never a default; (3) pre-merge checks: CI status, mergeability, conflicts,
and an honest note about the self-approval limitation rather than working around it;
(4) the merge with my preferred strategy; (5) cleanup: delete the remote branch per my
preference, and if local worktrees are involved, use worktree-safe removal — never
plain branch deletion under a worktree; (6) a stop rule: any failing check or
unresolved review finding halts the merge and comes back to me.

After writing it, test it on my next real PR.
  </task>
</prompt>
```

### Stakeholder Update Email

After work ships, sends (or drafts) a short, truthful update email to the person who needs to know — a client, a producer, a collaborator, your team. The skill encodes the discipline: updates go out only when something stakeholder-visible actually changed; the email describes shipped behavior in the recipient's vocabulary, not implementation details; nothing unverified gets called done; the format stays consistent (what changed, what it means for you, what's next); and you're CC'd or shown a draft first, per your preference.

Why build it

Communication is the half of client and team work that agent workflows usually drop. The skill's real content isn't email mechanics — it's the rules: only when shipped, only what's true, only in their language. A consistent, honest update cadence after real changes builds more trust than any amount of polish, and making it a skill means it actually happens instead of being the thing you'll do after lunch.

What you need

An email path your agent can use — a sending API like Resend, your mail provider's API, or just draft-for-you mode (no setup at all)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "stakeholder-update-email", stored
wherever my harness loads skills from.

The skill's job: after work ships with stakeholder-visible impact, send or draft a
short, truthful update email to the right person.

Before writing it, interview me for: who my recurring stakeholders are and what each
cares about, whether you should send directly (and through what — e.g. a Resend API
key in an env file) or always draft for my review, and whether I should be CC'd on
sends.

The skill must include: (1) trigger conditions — when work merges or ships with
visible impact for a stakeholder, or when I ask for an update email; (2) a gate: if
nothing stakeholder-visible changed, say so and send nothing; (3) writing rules:
describe shipped behavior in the recipient's vocabulary, not implementation detail;
never call anything done that wasn't verified; if something shipped partially, say
which part; (4) a consistent short format: what changed, what it means for them,
what's next; (5) the send/draft mechanics per my preference, with send requiring my
explicit confirmation.

After writing it, test it by drafting an update for the most recent thing I shipped.
  </task>
</prompt>
```

### Session-to-Skill Extractor

A continuous-learning loop for your skill library: at the end of substantial work sessions, this skill reviews what happened and asks whether any pattern is worth preserving — a workflow you'd repeat, a hard-won API discovery, a debugging path, a decision procedure. When a pattern clears the bar (recurring, non-obvious, codifiable), it drafts the new skill or updates an existing one, then files it for your review. Your library grows out of your actual work instead of requiring dedicated authoring time.

Why build it

Every skill in this library started as a session where someone solved a problem and refused to let the solution evaporate. This skill automates that refusal. It has a high bar built in — most sessions yield nothing, and that's correct — but compounding is the point: six months of extraction produces a library shaped precisely like your work. This is the skill that makes all the other skills self-multiplying.

What you need

Nothing — except a willingness to review what it proposes rather than auto-accepting

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "session-to-skill-extractor", stored
wherever my harness loads skills from.

The skill's job: at the end of substantial work sessions, evaluate whether anything we
did is worth preserving as a new skill or an update to an existing one — and if so,
draft it.

The skill must include: (1) trigger conditions — when I say "wrap up," "anything worth
keeping?", or at the natural end of a session where we solved something non-trivially;
(2) a high extraction bar, stated explicitly: the pattern must be RECURRING (I'll
plausibly need it again), NON-OBVIOUS (a fresh session wouldn't just derive it), and
CODIFIABLE (it can be written as a procedure) — most sessions yield nothing, and
"nothing worth extracting" is a good answer; (3) a check against my existing skill
library first: if an existing skill covers 80% of the pattern, propose an update, not
a new skill; (4) drafts follow my skills' standard format with trigger conditions, and
land somewhere for my review — never silently into the live library; (5) a sanitize
rule: extracted skills generalize the pattern and strip project/client specifics,
which stay in repo-local runbooks.

After writing it, test it on THIS session: evaluate whether our setup work contains an
extractable pattern, and show me your reasoning either way.
  </task>
</prompt>
```

### Agentic Harness Designer

A design-review skill for building agent-powered products and systems: when the problem is "how should this AI system actually work," it walks the real architecture questions — tool-use design, permission and approval models, workflow state and durability, context and memory strategy, evaluation approach, observability, and operator visibility — and produces a phased implementation plan. It encodes the hard-won principle that most "AI product" problems are agent-system problems: the model matters less than the harness around it.

Why build it

If you build anything agent-powered — internal tools, products, even sophisticated personal automations — the failure modes live in the harness: missing approval gates, no durable state, no evaluation plan, no way to see what the agent did. A skill that forces those questions in order, every time, is the difference between a demo and a system. It's the most conceptual skill in the library, and for builders, frequently the most valuable.

What you need

Nothing

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "agentic-harness-designer", stored
wherever my harness loads skills from.

The skill's job: when I'm designing or reviewing an agent-powered system or product,
walk the real architecture questions and produce a phased plan — treating the problem
as an agent-SYSTEM problem, not a model-choice problem.

The skill must include: (1) trigger conditions — designing, evaluating, or debugging
any AI-agent-powered product, tool, or serious automation; (2) the design walk, in
order: what tools the agent gets and their exact contracts; the permission model
(what's autonomous, what needs approval, what's forbidden); workflow state and
durability (what survives a crash or restart); context and memory strategy (what the
agent knows, from where, and what it must not accumulate); evaluation (how we'll know
it works — concrete checks, not vibes); observability (what's logged, what the
operator can see mid-run); (3) failure-mode review against the common killers: missing
approval gates, non-durable state, unbounded context growth, no evals, invisible
execution; (4) output: a design doc with decisions and rationale, plus a phased
implementation plan where each phase is independently shippable and testable.

After writing it, test it by reviewing an agent system or automation I describe — or
one we've already built — and showing me the design doc.
  </task>
</prompt>
```
 [](/open-skills/skills)  [](/open-skills/runbooks) 

Back to the Skills directory or continue into runbook compositions.