# Communication and Influence

The connective tissue of the role: translating between worlds, listening well, influencing without authority, and shaping the message so the right people act.

## Contents
- The communication bridge
- Active listening
- Influence without authority
- Driving toward clarity as a communication act
- Status, escalation, and executive updates
- Decide your narrative
- Tailoring to audience
- Avoiding wasted effort and churn

## The communication bridge

A TPM's technical background lets them both *understand* engineers and *translate* in both directions:
- **Business → technical**: turn business requirements (and domain jargon — tax legalese, medical terms) into functional specs the dev team can build from. Watch for words that mean different things to different teams (what "a purchase" means in law vs in an e-commerce workflow) — same word, two meanings, causes missed requirements.
- **Technical → business**: turn an engineering blocker into something a VP of marketing understands.

This bidirectional translation is the **single highest-leverage career skill** and shows up most in promotion/growth criteria. You don't always need the domain to land the job, but the ability to *learn the domain on the job* is essential. Sometimes the bridge is built ahead of time (creating short courses so the dev team learns the business domain's major concepts, and vice versa) so a common lexicon exists before requirements arrive.

## Active listening

The core technique for both empathy and clarity. Listen to *understand*, not to reply:
- **Be present** — phone away, occasional eye contact, undivided attention.
- **Don't interrupt** — when you're hunting for a gap to insert your thought, you're not absorbing theirs.
- **Paraphrase** what you heard and ask **open-ended follow-ups** ("What did you think about that?" not "Did you like that?").
- **Read non-verbal cues** (speech speed signals nervousness; crossed arms/furrowed brow shift the tone defensive) and **stay neutral** (no eye-rolls, no passive-aggressive sighs).

Follow-ups fill the gaps and paint a fuller picture; over time you learn each business team's domain and become a far more effective communicator.

## Influence without authority

A TPM rarely has org authority over the people they depend on; influence comes from **clarity and trust**:
- Be the person who manufactures clarity and **drives consensus** — without clarity you can't get everyone on the same page, let alone aligned on a path.
- Earn trust through **consistent, transparent communication**. Trust is largely *perception* of your ability to deliver, and your main lever on that perception is status communications. A missed or late status — even for a legitimate reason — reads as disorganized and invites micromanaging. Consistency is a measurable artifact that shows up in reviews.
- When a status is challenged and the program is actually well-run, the issue is usually **poorly worded/formatted information** leaving room for interpretation — not your delivery. Don't get defensive; fix the wording (some of the best status reports come from hour-long sessions with stakeholders on exactly how to format goals/risks/issues).
- Treat stakeholders as **partners and information sources**, not report recipients. They have perspectives that solve issues and reveal how problems intersect.

## Driving toward clarity as a communication act

Clarity is the defining behavior, and most of it happens through communication:
- **Issue resolution** — pull in the right stakeholders for their perspectives (e.g., a Windows↔macOS conversation revealing different TCP congestion-control algorithms — a risk worth a deep dive). Connect a problem to the right person *and* the right place; when you don't know either, ask stakeholders to find answers faster.
- **Estimation handoff** — when a task goes to a developer, drive maximum clarity: give full context up front, then probe. Ask how confident they are, what the estimate *includes* (design? review? testing? deployment? — many include only code-writing), and questions about the design. Use the answers to size the buffer. Be transparent about *why* you're asking ("I'm here so the project runs smoothly and tasks land when promised; to do that I evaluate estimates and add buffers for outside factors") — directness defuses pushback, and teams often start volunteering a confidence score.
- Questions to weigh against a developer's ideal-conditions estimate: deploy blockers on the release calendar? developer's experience in that area? new tech/unknowns? other projects touching the same code?

## Status, escalation, and executive updates

(Full report mechanics live in `stakeholder-management.md`; this is the craft.)
- **Every action needs an owner and a date** (or a date-for-a-date). This signals you're in control and proactive.
- **Escalate early, don't wait for a review.** Leadership/senior reviews are *reporting* forums, not where you first surface a blocker — if blocked, unblock now. Anything new in an LR/SLR should already be resolved or in progress.
- **Replace weasel words** (many, most, some, a lot, few, all) with concrete data — they sound quantitative but mean nothing and erode precision.
- **Hard messages**: pre-communicate to the affected stakeholder before it's broadcast, so they can prepare and don't feel set up. Frame decisions (a cut feature, a paused project) with compassion and a focus on the vision — understanding *why* reduces the emotional toll.

## Decide your narrative

The key to any report, regardless of audience, is deciding **what you're trying to convey** before writing. Even if you don't write in a narrative frame, knowing the narrative tells you which information to include and how to frame it.

## Tailoring to audience

The same issue is reported differently by audience. Example — a design delayed because two services can't agree on a contract (one wants `null`, the other can't handle it):
- **Technical, internal-platform audience**: name it precisely — a service-contract disagreement; options are default values vs updating the receiver to accept null; next steps.
- **Non-technical / customer-facing audience**: "design sign-off delayed due to a requirements mismatch; new ETA is X," with a link to the technical detail. Keep the surface at a level they understand: requirements/steps unclear, team driving clarity to close it.

Match depth to seniority and to technical fluency. Execs want trajectory and risk and confidence that the team has it handled; engineers and team leads want the day-to-day.

## Avoiding wasted effort and churn

Poor communication wastes time in two classic ways:
- **Duplicated/colliding work** — multiple teams solving the same problem or editing the same shared code (the "too many cooks" / overwriting-each-other's-commits problem). Cross-team communication catches the overlap; update the plan and correct fast.
- **Churn during issue resolution** — e.g., three requirements that logically can't coexist; a writer, developer, and tester fix one and break another in circles because they aren't talking. A TPM getting everyone in a room for an hour finds the root cause (the requirements conflict) and decides which to keep. The fix is almost always: get the right people communicating sooner — ideally with a TPM involved from the start.

These misses are rarely intentional — you're heads-down unblocking — but being an effective practitioner means always keeping stakeholders in the loop. It saves a lot of unnecessary work.
