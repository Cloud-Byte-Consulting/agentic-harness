# Open Skills Runbooks | Unlock AI

Generated 2.39:1 Open Skills runbooks masthead showing primitive modules chained into a workflow on the right side. [←Open Skills overview](/open-skills) 

# Open Skills runbooks.

Runbooks are what happen when primitives stop living alone. Each one chains named skills into a workflow with a concrete outcome: publish a page, brief a release, produce a video, verify a deploy, or preserve operational memory.

Use this page to see which skills combine, what each stage hands off to the next, and how a library of small primitives becomes a repeatable production system.

The primitive is the unit. The runbook is the machine you build from those units.

7 runbooks

## Same primitives, recombined into outcomes.

Runbooks show why small skills matter. You can swap, refine, and reuse the same primitives across different workflows instead of rebuilding the whole operation every time.

![Diagram showing individual skills as primitives composing into a runbook pipeline that turns a voice memo into a published page.](/open-skills/skills-runbooks.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)

### Runbook 01 · Talk to Published

 [](/open-skills/core-infrastructure#media-transcription)  [](/open-skills/research-thinking#brain-dump-processor)  [](/open-skills/writing-voice-content#personal-voice-skill)  [](/open-skills/core-infrastructure#html-artifact-builder)  [](/open-skills/web-publishing-frontend#personal-site-publisher) 

Media Transcription → Brain Dump Processor → Personal Voice → HTML Artifact Builder → Personal Site Publisher

You record a voice memo on a walk. Transcription turns it into clean text; the Brain Dump Processor separates and evaluates the ideas in it; you pick the one worth writing; the Voice skill drafts the piece as you'd write it; the Artifact Builder lays it out; the Site Publisher ships it to a clean URL with a proper link preview.

The payoff: A voice memo becomes a published page.

### Runbook 02 · Release Day

 [](/open-skills/core-infrastructure#current-information-search)  [](/open-skills/writing-voice-content#new-release-briefing)  [](/open-skills/writing-voice-content#branded-image-prompting-guide)  [](/open-skills/core-infrastructure#image-generation-gateway)  [](/open-skills/web-publishing-frontend#personal-site-publisher)  [](/open-skills/agent-operations#stakeholder-update-email) 

Current-Information Search → New Release Briefing → Branded Image Prompting → Image Generation Gateway → Personal Site Publisher → Stakeholder Update Email

Something big ships in your field at 10am and you want an accurate, on-brand briefing live by noon. Search gathers primary-source facts with dates (this step is what keeps you from publishing training-data hallucinations); the Briefing skill packages them into your standard format; the image skills produce a matching branded thumbnail; the Publisher ships it; the Update Email tells your list or team it's live.

The payoff: An accurate, on-brand briefing published the same day with speed that never costs correctness.

### Runbook 03 · The Video Production Line

 [](/open-skills/core-infrastructure#media-transcription)  [](/open-skills/video-media-production#radio-edit)  [](/open-skills/video-media-production#broll-pipeline)  [](/open-skills/video-media-production#ai-editing-assistant)  [](/open-skills/agent-operations#stakeholder-update-email) 

Media Transcription → Radio Edit → B-Roll Pipeline → AI Editing Assistant → Stakeholder Update Email

Raw talking-head footage in, finished video with motion graphics out. Transcription produces the timestamped foundation everything else reads. Radio Edit fixes the spoken narrative and hands you a paper edit to approve — the editorial decisions happen here, on paper, where they're cheap to change. The B-Roll Pipeline scouts the approved cut for graphic-worthy moments and generates consistent animated overlays. The NLE Assistant assembles it in your editor. The Update Email tells your editor or client it's ready for review.

The payoff: A raw video becomes a finished, graphics-laden edit with the editorial work front-loaded and cheap to change.

### Runbook 04 · Ship a Page You Can Trust

 [](/open-skills/web-publishing-frontend#frontend-taste-system)  [](/open-skills/web-publishing-frontend#personal-site-publisher)  [](/open-skills/testing-quality#browser-automation-qa)  [](/open-skills/testing-quality#testing-runbook-creator) 

Frontend Taste System → Personal Site Publisher → Browser Automation QA → Testing Runbook Creator

The difference between shipping a page and shipping a page you'd bet on. The Taste System builds it well; the Publisher takes it live; Browser QA then verifies the live page with instruments rather than vibes — screenshots across breakpoints, Core Web Vitals, console and network checks — and everything QA learned about testing this page lands in the repo's runbook, so the next deploy verifies in minutes.

The payoff: A personal site with a regression-test habit — every page shipped with verified quality.

### Runbook 05 · The Research Engine

 [](/open-skills/core-infrastructure#heavy-file-ingestion)  [](/open-skills/core-infrastructure#current-information-search)  [](/open-skills/research-thinking#assumption-checker)  [](/open-skills/research-thinking#meeting-synthesis)  [](/open-skills/core-infrastructure#html-artifact-builder)  [](/open-skills/research-thinking#reading-pack-builder) 

Heavy File Ingestion → Current-Information Search → Assumption Checker → Meeting Synthesis → HTML Artifact Builder → Reading Pack Builder

For real research questions with messy inputs: a folder of PDFs, some meeting recordings, and a claim you're not sure you believe. Ingestion converts the heavy sources into clean artifacts first (this ordering is the whole trick — analysis over converted text is faster, cheaper, and reusable). Search fills the gaps with current information. The Assumption Checker runs adversarially against the emerging conclusions — a separate skill with a skeptic's posture, not the same conversation grading its own homework. The output ships as a styled artifact, and when the material needs human review, the Reading Pack presents it in order.

The payoff: Research with a chain of custody: every claim traceable to an artifact, every conclusion stress-tested.

### Runbook 06 · Delegate and Verify

 [](/open-skills/agent-operations#session-operating-map)  [](/open-skills/agent-operations#goal-prompt-generator)  [](/open-skills/agent-operations#visible-delegation)  [](/open-skills/agent-operations#self-authored-pr-merge)  [](/open-skills/agent-operations#stakeholder-update-email) 

Session Operating Map → Goal Prompt Generator → Visible Delegation → Self-Authored PR Merge → Stakeholder Update Email

How one person runs parallel engineering lanes without becoming the bottleneck. The Operating Map records what each lane owns, so no session needs you to explain the project. The Goal Prompt Generator packages a task with a definition of done and verification gates; Visible Delegation runs it in a watchable session; when the delegate finishes, its work is verified against the gates it was given — the goal prompt is also the acceptance test. The PR Merge skill reviews and lands it; the Update Email closes the loop with whoever's waiting.

The payoff: Parallel engineering lanes with you only touching the two decisions that need you: what 'done' means, and whether the diff is good.

### Runbook 07 · The Flywheel

 [](/open-skills/agent-operations#session-to-skill-extractor)  [](/open-skills/testing-quality#testing-runbook-creator)  [](/open-skills/testing-quality#page-testing-memory)  [](/open-skills/agent-operations#session-operating-map) 

Session-to-Skill Extractor → Testing Runbook Creator → Page Testing Memory → Session Operating Map

This one is different: it's not a pipeline you run, it's a posture that runs under every other runbook. The Extractor watches your sessions for patterns worth keeping and drafts new skills from them. The Runbook Creator banks every testing discovery in the repo it belongs to. Page Testing Memory keeps the global/local boundary clean as both libraries grow. The Operating Map preserves coordination state across sessions.

The payoff: No useful discovery dies in chat — the mechanism by which a skill library compounds.

 [](/open-skills/skills) 

Back to the Skills directory.