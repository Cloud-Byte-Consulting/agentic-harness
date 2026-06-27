# Codex browser & annotations | Unlock AI

Generated 2.39:1 Codex browser annotations masthead showing selected browser elements and comment pins on the right side.

# Browser & annotations.

Browser annotations turn the page you are building into the work order. Click an element, attach a note, and Codex receives the selected UI with screenshot context.

Use this guide to send one precise fix, stack a full-page review, and make Codex resolve the actual component behind the browser evidence before it edits.

Pointing beats describing. Every play below is built on that one move.

![Four-step diagram of the annotation flow: click an element, attach a note, send with Enter or stack with command-Enter, then Codex changes the code and the page reloads.](/guides/codex/annotation-flow.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)

**The casePoint instead of describe.Most front-end frustration with AI agents is a translation problem: you can see the issue, and you're stuck converting pixels into prose. Annotations delete the translation step.**

Everything on this page happens with a page open in the Codex browser panel — a local file, a dev server, anything it can render. Prompts marked as annotation notes are what you type INTO the annotation box on a clicked element; the rest paste into the main chat like usual.

### What Codex receives when you click

Click any element on the page and you get a note box pinned to it. Type what you want — a change, a question, a complaint — and Codex receives your words together with a screenshot of the exact element you clicked. No selectors, no "the third button in the header," no ambiguity about which div you meant.

You can ask for changes, ask for explanations, or just leave observations. The element context rides along automatically.

You do

Click the element. Write the note like you'd talk to a designer sitting next to you.

The AI does

Resolves your click to the actual code behind the element and acts on the note with that context.

Copy prompt

**Show the full prompt**

```
<prompt>
  <context>
    Paste into the main chat before your first annotation click.
  </context>
  <task>
    This is my first annotation on this page. When it arrives, resolve the element and act on my note.
  </task>
  <requirements>
    <requirement>Tell me what element you received — tag, role, and your best description of what it is on the page.</requirement>
    <requirement>Tell me which file and lines of code render it.</requirement>
    <requirement>Restate my note as the change you're about to make (or the question you're about to answer).</requirement>
    <requirement>If my note could mean two different things, pick the most likely, say so, and proceed — don't stall.</requirement>
  </requirements>
  <deliverables>
    <deliverable>Execute the change or answer.</deliverable>
    <deliverable>Reload the page and tell me where to look.</deliverable>
  </deliverables>
</prompt>
```

**MechanicsEnter sends one. ⌘Enter builds a batch.Two keys, two modes of working. Single annotations interrupt politely; stacked annotations arrive as one coordinated instruction set.**

### Single shot vs the stack

Enter submits the annotation on its own — it joins the queue, or steers the current run, depending on your default submit behavior. That's right for one urgent fix.

⌘Enter stacks the annotation instead: keep clicking, keep noting, and the batch goes when you send it. Stacking matters more than it looks: five related notes that arrive as one instruction set get implemented as one coherent change, instead of five sequential edits where the third undoes the first. Sweep first, send once.

You do

One-off fix: Enter. Review pass: ⌘Enter each note, then send the stack when the sweep is done.

The AI does

Treats a stacked batch as a single work order — resolves conflicts between notes before touching code instead of discovering them mid-edit.

Copy prompt

**Show the full prompt**

```
<prompt>
  <context>
    Paste into the main chat before sending a stacked batch.
  </context>
  <task>
    I'm about to send you a stacked batch of annotations from a full-page sweep. Process it as one work order.
  </task>
  <requirements>
    <requirement>List every annotation in the batch: element, note, your read of the intent.</requirement>
    <requirement>Check the set for conflicts and overlaps — notes that touch the same component, requests that contradict each other, changes that should share one implementation. Tell me what you found before coding.</requirement>
    <requirement>Propose the implementation order and where you'll consolidate (one CSS change serving three notes beats three patches).</requirement>
  </requirements>
  <deliverables>
    <deliverable>Execute the consolidated changes.</deliverable>
    <deliverable>Reload the page and give me a checklist mapping each annotation to what changed, so I can verify them one by one.</deliverable>
  </deliverables>
</prompt>
```

### Know your default: queue or steer

When Codex is mid-run, a submitted annotation either queues — waits politely for the next safe pause — or steers, redirecting the work in flight. Your default behavior is a setting, and not knowing it is how annotations seem to vanish (they queued) or how a long task seems to wander (you kept steering it).

Know what your Enter does before you trust it during a critical run.

You do

Check your submit default in Settings. Decide your own house rule for when interrupting a run is worth it.

The AI does

Tells you how it handles mid-run input and confirms what happened to each note you sent while it was busy.

Copy prompt
```
<prompt>
  <task>
    Explain how you're currently handling input that arrives while you're working.

1. When I submit an annotation or message mid-run, does it queue for your next pause or steer you immediately? Tell me what my current default is if you can see it, and where I change it.
2. From this thread so far: did any of my messages queue or steer a run in flight? Reconstruct what happened to each.
3. Give me your honest recommendation: for annotation batches during long tasks, should I queue or steer — and what's the one situation where the other choice is right?
  </task>
</prompt>
```

![A printed webpage mockup on a drafting table with cyan sticky notes attached to specific elements and one element circled in pencil.](/_next/image?url=%2Fguides%2Fcodex%2Fannotation-drafting-table.jpg&w=3840&q=75&dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)

**WorkflowsFour annotation plays that earn their keep.The mechanics take five minutes. These are the plays that turn them into a working style — each one is a prompt plus a way of sweeping the page.**

Each play pairs a main-chat prompt (paste first, it sets the contract) with an annotation style for the sweep itself. The plays assume stacking with ⌘Enter.

### The visual bug hunt

Something on the page is broken and you can see it. Click the broken thing, describe what you see versus what you expect, and let Codex connect the pixels to the code. This beats pasting console errors for the whole class of bugs that are visible before they're loggable.

You do

Click the broken element. Note format: "Seeing X, expected Y, happens when Z."

The AI does

Traces the element to its code, diagnoses, fixes, and tells you how to confirm the fix on the reloaded page.

Copy prompt
```
<prompt>
  <task>
    Bug-hunt contract for the annotations that follow:

Each note describes a visual defect as: what I see / what I expect / when it happens. For each one:
1. Diagnose before you fix — name the cause in one line (CSS specificity, state not updating, layout overflow, stale data, whatever it truly is).
2. Fix the cause, not the symptom. If the honest fix is bigger than the annotation implies, say so and ask before the big version.
3. After the batch: reload, then walk me through each defect — cause, fix, and exactly where to look to confirm it's gone.
  </task>
</prompt>
```

### The design review pass

Open the page, put on your design hat, and sweep: spacing that's off, colors that drift from the brand, type that's almost right. Stack the whole critique with ⌘Enter and send it as one order. This is the play where annotations replace an entire class of meeting.

You do

Sweep the full page before sending anything. Note taste decisively: "tighten this gap," "this should match the header cyan."

The AI does

Implements the critique as one coherent polish pass, consolidating repeated notes into shared fixes.

Copy prompt
```
<prompt>
  <task>
    Design-pass contract for the incoming annotation stack:

1. These notes are taste calls, not bug reports — implement them as a single coherent polish pass.
2. Where several notes point at the same underlying token (one spacing scale, one color variable, one type size), fix the token, not each instance — and tell me you did.
3. Respect the existing design system: if a note fights the project's established patterns, flag it instead of silently complying.
4. After: reload and summarize the pass in design language (what got tightened, aligned, unified), then list anything you flagged as fighting the system.
  </task>
</prompt>
```

### The copy QA sweep

Words are UI. Click every headline, label, and button that reads wrong, write the better version (or just "this is clunky — improve"), and stack the sweep. Annotations are dramatically better than a copy doc here because every note is attached to its exact context: the line, the space it has to fit in, the words around it.

You do

Read the page like a stranger. Click anything you stumble on. Give either replacement text or a direction.

The AI does

Applies exact replacements verbatim, writes the "improve this" notes in the page's established voice, and never rewrites copy you didn't click.

Copy prompt
```
<prompt>
  <task>
    Copy-QA contract for the incoming annotations:

1. Notes with replacement text in quotes: apply verbatim, exactly as written.
2. Notes asking for improvement without text: rewrite in the page's existing voice — match the register of the copy around it, don't import a new tone.
3. Hard rule: touch ONLY annotated copy. No drive-by rewording of text I didn't click.
4. Watch the containers: if new copy threatens to wrap badly or overflow at mobile widths, shorten it and tell me.
5. After: a before/after table of every string you changed.
  </task>
</prompt>
```

### The explain-this-component play

Annotations aren't only for changing things. Click a component on an unfamiliar page — inherited project, agent-built UI you weren't watching, code you wrote six months ago — and ask what it is, where it lives, and why. The page becomes a clickable map of its own codebase.

You do

Click the mystery component. Ask the actual question: "what is this, where's the code, why is it here?"

The AI does

Answers from the real source: the component's purpose, its file and data flow, and the context you're missing.

Copy prompt
```
<prompt>
  <task>
    The annotations that follow are questions, not change requests. For each clicked element:

1. What it is: the component's job on this page, in plain language.
2. Where it lives: file, the key lines, and what data flows into it.
3. Why it's here: what breaks or degrades if it's removed.
4. Anything surprising: known quirks, tech debt, or history visible in the code or comments.

Change nothing. This is a guided tour, and the page is the map.
  </task>
</prompt>
```