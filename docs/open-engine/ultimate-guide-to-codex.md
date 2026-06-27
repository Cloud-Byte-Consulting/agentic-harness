# The Ultimate Guide to Codex | Unlock AI

Generated 2.39:1 Codex guide masthead showing an operational Codex workspace cockpit on the right side.

# The Ultimate Guide to Codex.

The Ultimate Guide to Codex is an operating manual for running the app like a working system: installation, projects, the side panel, browser annotations, threading, and skills.

By the end, you should know how to set up Codex, steer it with the right UI surfaces, delegate work across threads, and preserve the patterns that make future runs better.

It is AI-first on purpose: every item ships with a copy-paste prompt, labeled with what you do and what the AI does for you.

June 10, 2026Last verified5 pagesHub, deep-dives, libraryEvery itemHas a prompt

**Start hereAn AI-first guide, with receipts.Every item below comes with a copy-paste prompt. The text tells you what the thing is, the labels tell you who does what, and the prompt makes it happen.**

The pattern is the same everywhere in this guide: read the short explanation, check the You-do / AI-does split so you know which parts are clicks only you can click, then copy the prompt. Until Codex is installed, paste prompts into any assistant (ChatGPT, Claude). After install, paste them into Codex itself unless an item says otherwise. Press / or ⌘K to search the whole guide from any page.

### What Codex actually is

Codex is OpenAI's coding agent: it reads codebases, writes and edits files, runs commands, reviews changes, and automates development work. It shows up on five surfaces — the desktop app (Mac and Windows), the CLI, IDE extensions, the cloud at chatgpt.com/codex, and mobile — Codex lives in the ChatGPT app on iOS and Android — for monitoring long-running work from anywhere. They all share your account, your projects, and your skills.

This guide focuses on the desktop app, because it is the surface where Codex stops feeling like autocomplete and starts feeling like a colleague with a workbench: a side panel with files, review, terminal, and a browser you can click on.

![Diagram of the five Codex surfaces — desktop app, CLI, IDE extension, cloud, and mobile — all connected to one Codex agent.](/guides/codex/surfaces-map.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
One agent, five surfaces. Your account carries projects and skills across all of them.You do

Skim the surfaces map once so the names stop being mysterious.

The AI does

Interviews you about your setup and recommends which surfaces to actually adopt, in what order.

Copy prompt
```
<prompt>
  <task>
    I'm new to OpenAI's Codex and deciding which surfaces to set up. Interview me, one question at a time, about:

1. My computer (Mac/Windows/Linux) and phone.
2. Whether I live in a terminal, an IDE, or neither.
3. The kind of work I want an AI agent for (coding, writing, file wrangling, automation).
4. My ChatGPT plan, if I know it.

Then recommend which Codex surfaces to set up this week (desktop app, CLI, IDE extension, cloud, mobile), in order, with one sentence on why each earns its spot. Keep the whole answer under 250 words.
  </task>
</prompt>
```

### How to work through this guide

This is a living guide: it gets re-verified against the current app and updated as Codex changes. Work it top to bottom the first time — install, first steps, projects — then treat the deep-dive pages (side panel, browser and annotations, threading, skills) as references you return to.

Long prompts are collapsed so the page stays readable; the copy button always grabs the full text. Anything that is genuinely its own discipline lives on its own page and is linked where it belongs.

You do

Bookmark this page. Decide which project you'll use as your learning sandbox.

The AI does

Acts as your concierge: fetches the guide, figures out where you are, and routes you to the right section.

Copy prompt
```
<prompt>
  <task>
    I'm working through the Unlock AI guide to Codex at https://unlock-ai.natebjones.com/guides/codex — fetch it if you can browse.

Ask me three quick questions to locate me:
1. Is Codex installed and signed in yet?
2. Have I run a real task in one of my own folders?
3. What's the next thing I'm trying to get out of it?

Based on my answers, point me to the next one or two sections of the guide to do today, and give me a single concrete task to try in my own project that proves the section worked.
  </task>
</prompt>
```

![A tidy desk at night with an open laptop, a fresh notebook, and a warm desk lamp — day one with a new tool.](/_next/image?url=%2Fguides%2Fcodex%2Fcodex-day-one-desk.jpg&w=3840&q=75&dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)

**Step 1 · InstallFrom zero to a signed-in agent.Installation is the one stretch of this guide where your hands do most of the work. It is four moves: know your plan, install the app, add the CLI, run one safe first task.**

Pre-install prompts go to any assistant you already use. The moment the app launches and you are signed in, switch to pasting prompts into Codex itself — including the one that installs its own CLI.

### Know what your plan gets you

Every ChatGPT plan includes Codex — Free for basic exploration, Go ($8/mo) for light use, Plus ($20/mo) for weekly coding sessions, Pro ($100+/mo) for 5x to 20x higher limits. Usage is measured in a rolling 5-hour window: on Plus, roughly 15–80 GPT-5.5 messages or 20–100 GPT-5.4 messages per window. You can also sign in with an OpenAI API key for usage-based billing, though some app features need a ChatGPT login.

Do not overthink this. Plus is the sensible default; upgrade when you actually hit limits, not before.

You do

Confirm which plan you're on at chatgpt.com → Settings.

The AI does

Estimates whether your intended usage fits your plan and flags the realistic upgrade trigger.

Copy prompt
```
<prompt>
  <task>
    Help me sanity-check my ChatGPT plan for Codex usage. Facts as of June 2026: every plan includes Codex; Plus ($20/mo) allows roughly 15-80 GPT-5.5 messages or 20-100 GPT-5.4 messages per rolling 5-hour window; Pro ($100+/mo) raises that 5x-20x; an OpenAI API key is an alternative with usage-based pricing.

Interview me briefly:
1. What plan am I on now?
2. How many hours a week do I expect to use a coding agent?
3. Do I tend to run long autonomous tasks or short bursts?

Then tell me: does my plan fit, what usage pattern would make me hit the window limits, and what the first sign of needing an upgrade will look like. Verify current pricing at https://developers.openai.com/codex/pricing if you can browse, and flag anything that changed.
  </task>
</prompt>
```

### Install the desktop app

Two clean paths on a Mac: download the app from developers.openai.com/codex, or install the Homebrew cask. Launch it, sign in with your ChatGPT account, and you land on the home screen: a composer asking what you should work on, with a permission picker, model picker, and project selector built into it.

This is the one genuinely manual step in the whole guide. Everything after it can be delegated.

You do

Run the install, drag to Applications if you downloaded the DMG, launch, sign in with your ChatGPT account, and grant the permissions macOS asks for.

The AI does

Once the app is open, audits its own install: versions, auth, defaults, and where files will land.

Copy prompt

**Show the full prompt**

```
<prompt>
  <context>
    Run immediately after your first launch, before changing defaults.
  </context>
  <task>
    You are freshly installed. Run a first-launch audit and report back in a table.
  </task>
  <requirements>
    <requirement>Your app version and the model you're currently set to use.</requirement>
    <requirement>Which account I'm signed in with (name the workspace/plan if visible to you, never paste tokens or keys).</requirement>
    <requirement>Your current permission/approval mode and what it means for what you can touch.</requirement>
    <requirement>Your current working directory (run pwd) and whether it's a generated scratch workspace or a folder I chose.</requirement>
    <requirement>Whether the codex CLI is also installed (run: which codex && codex --version).</requirement>
  </requirements>
  <constraints>
    <constraint>Make no changes.</constraint>
  </constraints>
  <deliverables>
    <deliverable>A table with all five audit rows.</deliverable>
    <deliverable>The one default you'd recommend I change first, and why.</deliverable>
  </deliverables>
</prompt>
```

Mac install: `brew install --cask codex` or the download at developers.openai.com/codex. Windows has its own installer; the CLI covers Linux.

### Let Codex install its own CLI

The CLI is the same agent in your terminal — useful for quick tasks, scripting, and the day the app is busy rendering something. It shares auth and configuration with the app, so installing it second is nearly free. This is also your first real delegation: let the agent install and verify its own tooling while you watch the commands go by.

You do

Approve the install commands when Codex asks.

The AI does

Picks the right install method for your system, runs it, verifies the result, and teaches you the three commands worth knowing.

Copy prompt
```
<prompt>
  <task>
    Install the Codex CLI on this machine and verify it works.

1. Check if it's already present: which codex && codex --version.
2. If missing, install using the official method that fits this system best — curl -fsSL https://chatgpt.com/codex/install.sh | sh on Mac/Linux, or npm install -g @openai/codex if I clearly use npm for global tools. Tell me which you chose and why before running it.
3. Verify: codex --version, and confirm it shares my existing sign-in.
4. Finish by teaching me, in five lines: how to start an interactive session, how to run a one-shot task with codex exec, and how to switch models with /model.

If anything fails, show me the exact error and fix it before moving on.
  </task>
</prompt>
```

### Run one safe first task

Your first real thread should be read-only recon on a folder you actually care about — not a toy repo, not a hello world. You want to feel the loop: it reads, it reasons, it reports, and nothing on disk changes. Pick the project you'll be working in next week and have Codex tell you what it sees.

You do

Open a real project folder as the thread's workspace and make sure Local execution is selected. Set the permission picker to Read only.

The AI does

Maps the folder, explains it back to you, and proposes — without making — three improvements.

Copy prompt
```
<prompt>
  <task>
    This is a read-only reconnaissance task. Change nothing on disk.

1. Confirm your working directory and that you're in read-only mode.
2. Map this folder: what kind of project is it, what are the key files and entry points, what's generated vs hand-written?
3. Tell me three things about this codebase you'd want to know if you were about to work in it (conventions, traps, oddities).
4. Propose three small improvements, each one sentence, ranked by value. Do NOT implement any of them.

Format: a short overview paragraph, then the three findings, then the three proposals. If anything in the folder looks like a secret or credential file, name the path but never print its contents.
  </task>
</prompt>
```

**Step 2 · First stepsGet your bearings, then make it yours.Ten minutes of orientation saves you weeks of fighting defaults. Learn the layout, then deliberately set three things: your model, your approval mode, and your house rules.**

Everything in this section happens inside the Codex app. The prompts here have Codex demonstrate and explain its own interface — which means the answers stay accurate even when the UI shifts under this guide.

### Learn the layout in ten minutes

The home screen is a composer asking what you should work on. Built into it: the permission picker (Full access, Ask before edits, Read only), the model picker, and a project selector. The left sidebar holds New chat, Search (⌘G), Plugins, Automations, your Pinned items, Projects, and Chats. Settings lives at the bottom.

The fastest way to learn it is to have Codex walk you through itself.

![Recreated Codex home screen showing the composer with the permission picker open: Full access, Ask before edits, and Read only options.](/_next/image?url=%2Fguides%2Fcodex%2Fcodex-home-permission-picker.png&w=3840&q=75&dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
The home screen, recreated as a scripted mockup: composer front and center, permission state one click away.You do

Follow along and click what it points at.

The AI does

Gives you a guided tour of its own interface, demonstrating each control with a harmless action.

Copy prompt
```
<prompt>
  <task>
    Give me a guided tour of this app, one element at a time. For each: what it's called, what it's for, and — where you can — demonstrate with a harmless action that changes nothing in my files.

Cover in order:
1. The composer and its three built-in controls: permission picker, model picker, project selector.
2. The sidebar: New chat, Search, Plugins, Automations, Pinned, Projects, Chats.
3. The side panel and each of its tabs (Files, Side chat, Review, Terminal, Browser) — one sentence each; we'll go deep later.
4. Settings: the three settings a new user should actually look at first.

Pause after each numbered group and wait for me to say "next". Keep each explanation under four sentences.
  </task>
</prompt>
```

### Pick your model and reasoning depth

As of June 2026 the lineup is: GPT-5.5, the frontier model for complex coding and agentic work; GPT-5.4, the flagship for professional everyday tasks; and GPT-5.4-mini, fast and cheap for light tasks and subagents. Older Codex-tuned models (GPT-5.2-codex, GPT-5.3-codex) have been retired for ChatGPT-plan users, so don't be surprised if a blog post names a model your picker no longer shows. Reasoning effort (low, medium, high — the picker shows variants like "5.5 Extra High") trades speed for depth.

The honest default: GPT-5.5 on medium for real work, mini for chores. Spend depth on tasks where being wrong is expensive.

![Recreated Codex composer with the model picker open, showing 5.5 Extra High selected and other model options below.](/_next/image?url=%2Fguides%2Fcodex%2Fcodex-model-picker.png&w=3840&q=75&dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
Model and reasoning depth live one click from the composer. Mockup recreation of the picker.You do

Open the model picker and set your default.

The AI does

Recommends a per-task-type model policy based on how you actually work, and shows you how to switch mid-thread.

Copy prompt
```
<prompt>
  <task>
    Help me set a model policy for this app. Current lineup as of June 2026: GPT-5.5 (frontier, complex coding and agentic workflows), GPT-5.4 (flagship for everyday professional work), GPT-5.4-mini (fast/cheap, good for subagents and chores). Reasoning effort can be low, medium, or high.

1. Tell me which model and reasoning level you're set to right now.
2. Ask me what my three most common task types are.
3. Give me a policy table: task type → model → reasoning level → one-line why.
4. Show me exactly how to switch (the picker in the composer, and /model in the CLI).

If the lineup has changed since June 2026, say so and use the current one instead.
  </task>
</prompt>
```

### Choose an approval mode you can live with

The permission picker is the most consequential control in the app. Read only means Codex can look but not touch — right for recon and unfamiliar codebases. Ask before edits is the working default: it proposes, you approve. Full access lets it edit and run commands without asking — earned, not granted, one project at a time.

The failure mode isn't recklessness, it's friction: people leave a trusted repo on ask-everything, get numb to approval dialogs, and stop reading them. Match the mode to the stakes and actually read the dialogs that remain.

You do

Set the mode per project: read-only for new codebases, ask-before-edits as your default, full access only where you'd let a junior dev commit unsupervised.

The AI does

Audits what your current mode allowed recently, so the policy is grounded in evidence instead of vibes.

Copy prompt
```
<prompt>
  <task>
    Audit my permission settings against my actual usage.

1. State your current approval mode in this thread and what it permits.
2. Look back over our recent threads in this project: list the last five actions that required (or would have required) approval — file edits, commands, deletions.
3. For each, tell me which mode would have blocked it, allowed it with a prompt, or allowed it silently.
4. Recommend a mode for THIS project specifically, and name the one category of action I should always keep behind an approval, regardless of mode.

Don't change any settings yourself — tell me what to click.
  </task>
</prompt>
```

### Teach Codex your house rules with AGENTS.md

AGENTS.md is a plain markdown file in your repo that Codex reads for standing instructions: how to navigate the codebase, what commands run the tests, which conventions matter, what to never touch. Think of it as the README that's actually for the agent. It is the single highest-leverage file in this whole workflow — rules written once stop being prompts you repeat forever.

You do

Answer the interview honestly, then review the file like it's a policy doc — because it is.

The AI does

Interviews you about the project, drafts AGENTS.md, and saves it at the repo root after you approve.

Copy prompt
```
<prompt>
  <task>
    Create an AGENTS.md for this project. Process:

1. Inspect the repo first: stack, structure, test setup, anything that looks like a convention.
2. Interview me, max six questions, about what you couldn't infer: commands to run and verify, code style I care about, files/folders that are off-limits, how I want changes proposed, anything that has burned me before.
3. Draft AGENTS.md with sections: Project overview · Commands that matter (build/test/lint) · Conventions · Boundaries (never touch) · How to verify your work before telling me it's done.
4. Show me the draft. After I approve, save it to the repo root.

Keep it under 60 lines. Rules an agent can act on, not aspirations — every line should change behavior.
  </task>
</prompt>
```

### Know where your config lives

User-level configuration sits in ~/.codex/config.toml — default model, provider settings, reasoning defaults. The schema isn't fully documented publicly, so the reliable move is to have Codex read your actual file and explain what's set rather than copying config blocks from blog posts that may be stale. Change one key at a time, and keep a dated backup comment when you do.

You do

Approve any edits one at a time. Keep a copy of the working version.

The AI does

Reads your real config, explains every line, and proposes at most two safe improvements.

Copy prompt
```
<prompt>
  <task>
    Walk me through my Codex configuration.

1. Read ~/.codex/config.toml (if it doesn't exist, say so and explain what defaults are in effect).
2. Explain every key that's set, in plain language: what it does, whether it's a default or something custom.
3. Propose at most TWO changes that would genuinely help based on how I've been using you — e.g., a saner default model or reasoning level. For each: the exact line, what it changes, the risk if any.
4. Wait for my approval, then apply changes one at a time, adding a dated comment above each. Show me the diff.

Never print anything from the file that looks like a key, token, or credential — refer to those as [redacted].
  </task>
</prompt>
```

**Step 3 · ProjectsSet up projects like you mean to come back.The difference between Codex as a toy and Codex as infrastructure is where your threads live. Real work starts from a real folder.**

Run these prompts inside a thread that's pointed at the project they concern. When a prompt moves files, it uses full paths and confirms before acting — keep that habit in your own prompts too.

### Scratch vs durable: know where your thread lives

New threads often start inside a generated per-thread workspace under ~/Documents/Codex/… — useful scratch space for experiments and throwaway output, but it is not your project. Files created there are easy to lose, aren't versioned, and won't be where you look next month. Durable work belongs in folders you chose on purpose: your repos, your workspace directories.

The rule that prevents 90% of the pain: before any work you intend to keep, confirm where you are. Threads are project-scoped at creation — pick the folder first, then start the thread, not the other way around.

![Diagram comparing the auto-generated scratch workspace under ~/Documents/Codex with a durable, deliberately chosen project folder.](/guides/codex/scratch-vs-durable.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
Scratch is a whiteboard. Durable is a filing cabinet. Decide which one you're writing on before it matters.You do

Before real work: check which folder the thread is in. Starting a new thread? Choose the project first.

The AI does

Audits where the current thread's files will land and moves anything stranded in scratch — with confirmation.

Copy prompt

**Show the full prompt**

```
<prompt>
  <task>
    Workspace audit. Before we do anything else in this thread:

1. Run pwd and tell me the working directory.
2. Classify it: a generated scratch workspace (e.g. ~/Documents/Codex/...) or a durable folder I chose on purpose.
3. If scratch: list any files created here in this thread that look worth keeping, and propose a destination using this exact pattern —
   Source: /full/path/to/file
   Destination: /full/path/to/durable/location
   Confirm both with me before moving anything.
4. If durable: confirm it's the right project for what I said I want to do.

Never use the scratch folder for anything I'd care about losing. If I ask for real work while we're in scratch, stop me and re-run this audit.
  </task>
</prompt>
```

### Induct a real project properly

When you bring Codex into a project you care about, do it as a ceremony, not a drive-by. One induction thread: the agent maps the codebase, you correct its misreadings, it writes AGENTS.md, and you end with a shared mental model that every future thread inherits. Twenty minutes here pays back on every thread after.

You do

Open the project folder, start a fresh thread in it, and correct anything it gets wrong — the corrections are the value.

The AI does

Maps the project, narrates its understanding, drafts the house rules, and proposes the first three real tasks.

Copy prompt
```
<prompt>
  <task>
    This is your induction into a project I intend to use you in regularly. In order:

1. Map the codebase: structure, stack, entry points, how it builds/tests/ships. Narrate your understanding in plain language and flag anything confusing — I'll correct you, and the corrections matter.
2. Read any existing docs (README, AGENTS.md, docs/). Tell me where docs disagree with the code.
3. If there's no AGENTS.md, draft one (interview me briefly first). If there is one, propose improvements based on what you just learned.
4. End with three concrete first tasks you'd recommend, ranked: one quick win, one cleanup that lowers future friction, one thing that's wrong and should be fixed soon.

Change nothing without approval. The deliverable is shared understanding, not edits.
  </task>
</prompt>
```

### One thread, one outcome

Threads are cheap; muddled context is expensive. Give each thread a single goal you could state in one sentence, and when the goal shifts, hand off to a fresh thread instead of dragging fifty messages of stale context behind you. The handoff prompt is the skill: a tight summary of state, decisions, and next actions that a fresh thread can pick up cold.

A thread that started in the wrong project can't be moved — write the handoff, start clean in the right place, and keep moving.

You do

Name threads by their goal. When a thread sprawls or sits in the wrong project, ask for the handoff and start fresh.

The AI does

Compresses everything that matters about the current thread into a handoff the next thread can run with.

Copy prompt

**Show the full prompt**

```
<prompt>
  <task>
    Write a handoff for continuing this work in a fresh thread.

Include, in this order:
1. Goal: what we're trying to accomplish, one sentence.
2. State: what's done, what's in flight, what's untouched.
3. Decisions: choices we made and the reasoning, so the next thread doesn't relitigate them.
4. Paths: full paths to every file and folder that matters.
5. Next actions: the first three things the new thread should do, in order.
6. Traps: anything that already burned us once.

Format it as a single copy-paste block. Assume the next thread is smart but knows nothing about this conversation.
  </task>
</prompt>
```

### Run projects in parallel without the chaos

The app runs multiple project threads side by side, and built-in Git worktree support means two threads can work the same repo without stepping on each other's changes. Add the iPhone app and you can check on a long-running thread from the couch. The habit that makes parallelism safe: every lane gets its own worktree, its own thread, and its own one-sentence goal.

You do

Decide what's actually parallelizable. Two lanes touching the same files isn't parallel, it's a merge conflict on a timer.

The AI does

Sets up worktree-isolated lanes and tells you exactly what's safe to run simultaneously.

Copy prompt
```
<prompt>
  <task>
    I want to run parallel work lanes in this repo. Set it up safely.

1. Ask me what the two (or three) workstreams are.
2. Judge honestly: which can run in parallel without touching the same files, and which should stay sequential. Tell me if my split is a bad idea.
3. For each parallel lane, create a Git worktree on its own branch (name them clearly), and tell me which thread should own which worktree.
4. Give me the rules of the road in five lines: where each lane works, how changes merge back, and the one command to clean up a finished worktree (git worktree remove, never rm -rf).

Confirm the worktree locations with me before creating anything.
  </task>
</prompt>
```

**Tour stop · Side panelThe workbench next to the chat.Files, Side chat, Review, Terminal, Browser. The chat is where you talk; the side panel is where you verify. It deserves its own page — here's the door.**

### Meet all five tools in one drill

The side panel is what separates the app from every chat window you've used: you can watch files change, question the agent mid-run without derailing it, walk diffs before accepting them, see every command execute, and click around the page you're building. One drill touches all five.

You do

Run the drill in a project with a few files. Click each panel as Codex references it.

The AI does

Performs a tiny task that deliberately exercises Files, Side chat, Review, Terminal, and Browser in sequence.

Copy prompt
```
<prompt>
  <task>
    Run a side-panel drill so I learn all five tools in one pass. Do a deliberately small task — create a single self-contained HTML page in this folder called hello-panel.html with a heading and one button — and as you work, tell me when to look at each panel:

1. Files: show me where the new file appears and what changed.
2. Terminal: run a harmless command (like listing the folder) so I see command output streaming.
3. Review: walk me through the diff of what you created before I accept it.
4. Browser: open hello-panel.html so I can see and click the result.
5. Side chat: tell me to ask you "why did you structure the HTML that way?" in the side chat, and answer it there.

Narrate each step in one or two sentences. The task is trivial on purpose — the panels are the lesson.
  </task>
</prompt>
```
 [](/guides/codex/side-panel) 

Read the full side panel deep-dive →

**Tour stop · BrowserThe killer feature: annotate the actual page.The in-app browser doesn't just show your work — you can click any element, attach a note, and Codex sees exactly what you see. Enter sends one; ⌘Enter stacks a batch. This changes how you do front-end work.**

### Your first annotation pass

Instead of describing a UI problem in words — "the button, no, the other button, the blue one in the header" — you click the button, type "this one, make it match the brand color," and Codex receives your note pinned to a screenshot of that exact element. Stack several annotations with ⌘Enter and related fixes ship together as one instruction set.

You do

Open a page in the browser panel, click three things you'd change, write one short note on each, then ⌘Enter to send the batch.

The AI does

Receives each note with its element screenshot, makes the changes, and reloads the page so you can verify by looking.

Copy prompt
```
<prompt>
  <task>
    I'm about to do my first annotation pass on the page that's open in your browser panel. Help me make it a clean rep:

1. Confirm you can see the page and tell me its URL or file path.
2. After I send a batch of annotations, restate each one back to me: the element, what I asked for, and what you'll change in the code.
3. Make the changes, then reload the page and summarize what's visibly different.
4. If any annotation was ambiguous, say what you assumed instead of guessing silently.

When we're done, tell me one way to write sharper annotations next time.
  </task>
</prompt>
```
 [](/guides/codex/browser-annotations) 

Read the full browser & annotations deep-dive →

**Tour stop · ThreadingThreads are the unit of work.Steering versus queueing, parent goals delegating to child threads, long-running goal mode — threading is where Codex goes from assistant to operation. It's mega. It gets its own page.**

### Watch a parent thread delegate

A parent thread can hold a goal and a plan, then spin up child threads to execute steps — each child focused, each reporting back. Once you've seen the pattern, you stop cramming everything into one conversation and start running Codex like a small team.

![Recreated Codex interface showing a parent goal plan creating a child thread, with the child thread executing the first step.](/_next/image?url=%2Fguides%2Fcodex%2Fcodex-child-thread-orchestration.png&w=3840&q=75&dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
Parent holds the plan, child does the step. Mockup recreation of goal orchestration in the app.You do

Pick a goal with two or three clearly separable steps. Watch the parent stay clean while children do the work.

The AI does

Creates a goal with a concrete outcome, drafts the plan, and executes the first step in a child thread.

Copy prompt
```
<prompt>
  <task>
    I want to see goal-and-child-thread orchestration on a real but small piece of work.

1. Ask me for a goal that has 2-3 separable steps (if I don't have one, propose one for this project that touches nothing risky).
2. Create the goal with concrete success criteria — what done looks like, verifiable.
3. Draft the plan as numbered steps. Keep this parent thread for plan, coordination, and review only.
4. Execute step one by creating a child thread for it. Tell me when the child starts and summarize its result here when it finishes.
5. End by telling me the status of the goal: what's done, what's next, what you need from me.

Narrate the mechanics as you go — I'm learning the pattern, not just shipping the task.
  </task>
</prompt>
```
 [](/guides/codex/threading) 

Read the full threading deep-dive →

**Tour stop · SkillsStop repeating prompts. Install them.A skill is a folder with a SKILL.md file: instructions, trigger conditions, optional scripts. Once installed, Codex pulls it in automatically when the task matches, or you call it by name. The Open Skills library has 31 of them ready to adapt.**

### Build your first skill in ten minutes

Skills live in ~/.codex/skills (personal, global) or .agents/skills inside a repo (shared with the project). Each is a folder with a SKILL.md — YAML frontmatter carrying the name and a description of when it should trigger, then markdown instructions. Codex ships a built-in creator: mention $skill-creator and it scaffolds one for you.

The right first skill is whatever instruction you've already typed three times this week.

You do

Pick the instruction you keep repeating. Confirm where the skill should live: personal (~/.codex/skills) or this repo (.agents/skills).

The AI does

Interviews you, writes the SKILL.md with honest trigger conditions, installs it, and runs a test invocation.

Copy prompt
```
<prompt>
  <task>
    Help me create my first Codex skill using $skill-creator (or manually if that's unavailable).

1. Ask me: what instruction do I keep repeating to you? That's the skill.
2. Ask whether it's personal (install to ~/.codex/skills) or project-wide (install to .agents/skills in this repo).
3. Write the SKILL.md: frontmatter with a kebab-case name and a description that says exactly when the skill SHOULD and SHOULD NOT trigger, then the instructions as numbered steps with any guardrails.
4. Install it, then prove it works: invoke it by name ($skillname) on a tiny example and show me the result.
5. Tell me how to edit it later and how to tell whether it's actually triggering when it should. No secrets in the skill file, ever — reference environment variables or local config instead.
  </task>
</prompt>
```
 [](/open-skills) 

Browse the Open Skills library →

**Stay currentThis guide has a verification date. So should you.Codex ships changes weekly. This guide was last verified against the app on June 10, 2026. The habit that keeps you current is the same one that keeps this guide honest: diff reality against your assumptions on a schedule.**

### The monthly changelog sweep

OpenAI keeps a dated changelog at developers.openai.com/codex/changelog, and it is the single best source for what actually changed — features often ship there before any documentation exists. Once a month, have Codex read it for you and report only what affects your workflow. Five minutes, and new capabilities stop arriving by rumor.

You do

Put a monthly reminder somewhere you'll see it. That's the entire manual step.

The AI does

Fetches the changelog, filters it against how you actually use Codex, and flags anything this guide's advice no longer matches.

Copy prompt

**Show the full prompt**

```
<prompt>
  <task>
    Run my monthly Codex changelog sweep.

1. Fetch https://developers.openai.com/codex/changelog and read the entries since [LAST CHECKED DATE — fill this in].
2. Sort what you find into three buckets:
   - Changes my workflow: features or changes that touch how I use the app, threads, skills, browser, or models.
   - Worth knowing: interesting but not urgent.
   - Ignore: platform/enterprise items that don't apply to me.
3. For each item in the first bucket, one sentence on what to do differently.
4. Flag anything that contradicts advice I'm following from the Unlock AI Codex guide (last verified June 10, 2026), so I know what to re-check there.

Keep the whole report under 300 words.
  </task>
</prompt>
```

## Three deep-dives, one library.

The hub gets you operational. The deep-dives are where the compounding skills live — return to them as the work demands.

 [

## Side panel deep-dive

Files, Side chat, Review, Terminal, Browser — the verification workbench.

Open page](/guides/codex/side-panel)  [

## Browser & annotations

Click the page, attach notes, steer the agent from inside your UI.

Open page](/guides/codex/browser-annotations)  [

## Threading & child threads

Steering vs queueing, parent goals, child execution, handoffs.

Open page](/guides/codex/threading)  [

## Open Skills

31 installable skills and 7 runbooks that compose them.

Open page](/open-skills) 

This is a living guide, last verified against the app on June 10, 2026. Codex ships weekly; when something here disagrees with the app in front of you, the app is right and this page is next in line for an update.