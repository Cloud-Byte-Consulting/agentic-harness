# Codex side panel deep-dive | Unlock AI

Generated 2.39:1 Codex side panel masthead showing a vertical verification rail on the right side.

# The side panel.

The side panel is Codex's verification workbench. Five tabs sit beside every conversation: Files, Side chat, Review, Terminal, and Browser.

Use this guide to understand what each panel is for, when to open it, and how the side rail turns agent output into inspectable, testable work.

![Diagram of the Codex side panel with its five tabs — Files, Side chat, Review, Terminal, Browser — and the job of each.](/guides/codex/side-panel-map.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)

**OrientationThe chat talks. The panel proves.Everything Codex claims to have done is verifiable one tab away: the files it touched, the diff it produced, the commands it ran, the page it built. Trust comes from looking.**

Run these prompts inside a project thread in the Codex app. Each one is designed to make a specific panel earn its keep — keep the side panel open and click what the prompts point at.

### Build the verification habit

The most common failure mode with coding agents isn't bad output — it's unverified output. The side panel exists so that checking is cheaper than trusting. The habit to build in week one: every time Codex says it did something, find the evidence in a panel before you say "looks good."

You do

After each meaningful Codex action, glance at the relevant tab: edit → Files, claim of success → Terminal output, finished feature → Browser.

The AI does

Tells you, for each action it takes, exactly which panel shows the evidence.

Copy prompt
```
<prompt>
  <task>
    For the rest of this thread, end every action you take with a one-line "Verify:" pointer telling me which side panel tab shows the evidence and what I should see there. Examples:

- Verify: Files → src/app.ts shows the new function highlighted.
- Verify: Terminal → test run with 14 passed, 0 failed.
- Verify: Browser → reload shows the new header layout.

If an action produces no panel-visible evidence, say "Verify: none — take my word for it" so the gaps are visible too. Start now by doing a trivial demonstration: list this folder and point me to where I can see the command output.
  </task>
</prompt>
```

**Tab 1 · FilesSee what Codex sees.The Files tab is the project tree from the agent's point of view — what it can read, what it changed, what just appeared.**

### Use Files as a map, not a tree

Browsing your own repo through the Files tab feels redundant until you realize what it's really for: confirming the agent's reality matches yours. After any multi-file change, the Files tab is the fastest answer to "what did you actually touch?" — no git command required.

You do

Open the Files tab after any task that creates or edits more than one file.

The AI does

Narrates a guided tour of the files it considers most important and explains every path it modified this session.

Copy prompt
```
<prompt>
  <task>
    Using the Files tab as our shared map:

1. List every file you've created or modified in this thread, with full paths and a one-line reason for each.
2. Name the five files in this project you'd read first if you were onboarding a new developer, and why those five.
3. Point out anything in the tree that surprised you when you first read this project — generated folders, naming oddities, things that look misplaced.

I'll follow along in the Files tab. Where a file you mention is one you changed, tell me so I can open it and look at exactly what's different.
  </task>
</prompt>
```

**Tab 2 · Side chatAsk questions without derailing the run.Side chat is a second conversation lane: the main thread keeps executing while you ask why, what, and what-if on the side.**

### Main thread for work, side chat for understanding

Every question you drop into the main thread becomes part of the working context the agent has to carry. Side chat keeps curiosity from polluting execution: ask why it chose an approach, what a piece of code does, whether an alternative was considered — all without nudging the run off course.

The discipline: if your message is meant to change what Codex does, it belongs in the main thread. If it's meant to change what you understand, it belongs in side chat.

You do

Route your questions: instructions to the main thread, curiosity to the side chat.

The AI does

Answers in the side lane with full awareness of the main thread's work, without letting the questions alter the plan.

Copy prompt
```
<prompt>
  <task>
    (Ask this in the side chat while a task runs in the main thread.)

Without changing anything about what you're doing in the main thread:

1. Explain the approach you're currently taking and the main alternative you considered but rejected.
2. What's the riskiest assumption in the current plan?
3. If this approach fails, what would the failure look like, and what's plan B?

Answer as commentary, not as new instructions — the main thread's plan stays exactly as it is unless I say otherwise over there.
  </task>
</prompt>
```

**Tab 3 · ReviewWalk the diff before you accept it.Review is where proposed changes become accepted changes. Treat it like a pull request from a fast, eager teammate: usually right, never above checking.**

### The two-pass review ritual

Pass one: have Codex review its own diff like a skeptical senior engineer — it routinely catches its own issues when explicitly asked to look. Pass two: you walk the diff in the Review tab, reading the files you understand best first. Accept when both passes are clean. This sounds heavy; it takes three minutes and catches the category of bug that costs an afternoon.

You do

Read the diff in the Review tab — at minimum every file you'd be embarrassed to break.

The AI does

Runs the skeptical self-review first and hands you a prioritized list of what deserves human eyes.

Copy prompt
```
<prompt>
  <task>
    Before I review your changes, review them yourself — as a skeptical senior engineer who did not write this code.

1. Walk the full diff. For each file: what changed and why it's safe.
2. Hunt specifically for: unintended changes (things in the diff that don't serve the task), broken assumptions (code that works only if something else is true), missing pieces (the task says done, but an edge is unhandled), and anything that touches files outside the task's scope.
3. Rate your confidence per file: high / medium / look-at-this.
4. Hand me a review order: which files deserve my human eyes first and what to look for in each.

Then wait. Do not consider the changes accepted until I've walked the Review tab myself.
  </task>
</prompt>
```

![A workshop pegboard with five hand tools hung in a neat row, each in its outlined spot, lit by a cool accent strip.](/_next/image?url=%2Fguides%2Fcodex%2Fside-panel-workbench.jpg&w=3840&q=75&dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)

**Tab 4 · TerminalWatch the commands. Run your own.The Terminal tab streams every command Codex runs — and gives you a shell in the same context. Settings lets you dock it at the bottom or the right.**

### Make command runs legible

An agent that runs commands silently trains you to stop watching. Flip it: have Codex announce intent before each command, then check the stream. Within a week you'll recognize its normal rhythms — and that's exactly what makes the abnormal moment (a command you didn't expect, a path you don't recognize) jump out at you.

You do

Keep the Terminal tab visible during execution-heavy tasks. If a command surprises you, pause and ask — that instinct is the safety system.

The AI does

Announces every command before running it: what, why, expected outcome.

Copy prompt
```
<prompt>
  <task>
    For all command execution in this thread, use announce-then-run:

Before each command, one line: `command` — why I'm running it — what success looks like.
After it finishes, one line: ✓ matched expectation, or ✗ unexpected, with what differed and what you'll do about it.

If a command fails twice, stop and bring it to me instead of trying a third variation silently. Start with the project's test suite (find the right command from the repo's config), and let me watch the pattern in the Terminal tab.
  </task>
</prompt>
```

**Tab 5 · BrowserThe page you're building, inside the app.The Browser tab renders local pages and dev servers next to the conversation — with the controls you'd expect and one you wouldn't: annotations. The basics live here; the annotation workflow has its own page.**

### The everyday browser toolkit

The unglamorous controls earn their keep daily: reload and force reload when you don't trust what you're seeing; the device toolbar to check phone widths without leaving the app; zoom for detail work; clear cookies and clear cache when state from an old session is lying to you; and screenshot-to-clipboard, which drops a capture you can paste straight into the chat for instant shared context.

That screenshot move is the gateway drug for the annotation workflow — when pointing at a picture beats describing a page, you're ready for the deep-dive.

You do

Drive the browser yourself: reload, device toolbar, zoom. Paste a screenshot into the chat when words get slow.

The AI does

Diagnoses stale-state problems and tells you which browser control un-lies the page.

Copy prompt
```
<prompt>
  <task>
    The page in your browser panel doesn't look right to me and I suspect stale state. Run the un-lying protocol:

1. Tell me what URL/file is loaded and when it was last rebuilt or reloaded.
2. Force reload. Did anything visibly change? Say what.
3. If it still looks wrong, walk me through the order of operations: clear cache, then clear cookies, then reload again — and explain what each step rules out.
4. If the page still misbehaves, check the dev server: is it running, did the last build succeed, is the browser even pointed at the right port?
5. Conclude: was it stale state, a build problem, or an actual bug — and what's the fix?

I'll paste a screenshot from the browser panel if you need to see what I see.
  </task>
</prompt>
```
 [](/guides/codex/browser-annotations) 

Continue to browser & annotations →