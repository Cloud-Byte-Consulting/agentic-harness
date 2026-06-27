# Emotional Intelligence and Using AI

The two skills that increasingly define a TPM's edge: the human soft skills GenAI can't replace, and the disciplined use of GenAI for the work it can accelerate.

## Contents
- Why EQ matters for TPMs
- The four components of emotional intelligence
- Growing your EQ
- Assessing your EQ
- Applying EQ: decisions, risk attitude, stakeholders
- Using GenAI safely (what it's for and not for)
- GenAI across the key management areas
- Prompt engineering
- Limitations, grounding, and ethics

## Why EQ matters for TPMs

Emotional intelligence (EQ) is the ability to recognize, understand, manage, and use emotions. As GenAI absorbs functional/administrative TPM tasks (risk lists, document generation, even planning), the durable differentiator is the soft skills it cannot replicate — chief among them EQ. It's what lets you form relationships with stakeholders, use empathy to drive change, ease tension in an escalation, and adapt to change with intent rather than reaction. "EQ trumps IQ" (Satya Nadella) — emotionally intelligent leaders reach their teams and effect change in ways that raw intelligence alone cannot.

## The four components of emotional intelligence

A pyramid, each layer the foundation for the next:
1. **Self-awareness** (base) — knowing your emotions and what triggers them.
2. **Self-regulation** — pausing between an emotion and a reaction so your response matches your values (and building adaptability to change).
3. **Empathy** — understanding others' feelings (impossible to do well without understanding your own first).
4. **Social skills** (top, most visible) — effective communication, conflict resolution, and motivation, built on the three below. Motivation is part of social skills, not a separate component.

## Growing your EQ

- **Self-awareness**: be curious about yourself; be vocally self-critical (constructively, not negatively); practice mindfulness (journaling — handwriting aids retention; short meditation reduces stress); step *outside* yourself to view your own actions as a bystander would; seek human connection (in-person beats video — the brain mirrors people you're with, building empathy).
- **Self-regulation**: anticipate your emotional responses from known triggers and preempt them to create a pause before acting.
- **Empathy**: imagine being in the other person's shoes; use **active listening** (be present, use non-verbal cues, ask open-ended questions, stay neutral — see `communication-and-influence.md`).
- **Social skills**: the outward result — drive conversations to resolution rather than demanding an outcome; use positive emotion to motivate (Jeff Bezos keeping shareholders steady through deliberate loss-for-growth; Satya Nadella's "One Microsoft" empathy-and-growth vision turning around an adversarial culture). The counter-example — a leader throwing a chair in frustration — shows poor regulation poisons a whole team's culture.

## Assessing your EQ

Five common assessments, all measuring the same high-level areas (identify, understand, perceive, regulate, use emotions; two add stress management): **EQ-i 2.0**, **MSCEIT**, **PEC**, **TEIQue**, **WEIS**. No one test is clearly superior — any that covers the high-level areas gets you started. Prefer an in-person workshop if budget allows (you grow alongside others on the same journey); online is a fine alternative. The assessment points the direction; growing EQ is a long journey beyond the workshop.

## Applying EQ: decisions, risk attitude, stakeholders

- **Decisions** — TPMs make small decisions all day (which task next, how to hit the milestone, how to frame a blocker). Balance the **rational** (KPIs, metrics) with the **emotional impact** on the team. When cutting a feature or pausing work, relay it with compassion and a focus on the vision — understanding *why* lessens the emotional toll (the layoff example: a director who prioritized employees' mental health and job placement got 60 of 66 placed and a smooth transition).
- **Risk attitude / tolerance** — people range from risk-averse (→ aversion → paranoia) to risk-seeking (→ addiction). When stakeholders span the spectrum, a **set risk-scoring procedure** smooths emotional responses and removes bias by forcing discussion of the true inherent risk.
- **Stakeholders** — most stakeholder best practices *are* acts of empathy (regular open communication, pre-warning before a status lands). The vocal-stakeholder example: rather than match a stakeholder's raised voice (the path of least friction), recognize their risk aversion (two decades of failed projects), de-escalate by repeating their concerns and acknowledging their experience, then move the conversation to a rational footing (the mandatory field can be relaxed later but not added later) — and they came around. Acknowledging emotions keeps the doors open for the real disagreement to surface and be addressed.

## Using GenAI safely (what it's for and not for)

GenAI is here to stay and changes *what* the TPM does day-to-day. Use it for the **mundane, repeatable, generative** tasks; lean on **EQ** for everything relational. It is **not** a substitute for knowing your stakeholders, personalizing real outreach, or stakeholder analysis — those are your EQ strengths. Core discipline: **write first, optimize second** — draft it yourself so the tone and facts are yours, then use GenAI to refine. Treat GenAI like any third-party tool: with caution and restraint.

## GenAI across the key management areas

- **Planning** — can generate a Gantt from a plan (though purpose-built tools do it better) and, more usefully, give *insights*: task-order/dependency analysis, resourcing/crashing what-ifs, optimization suggestions — like asking a colleague for a perspective.
- **Risk** — generate a first-pass risk list from a charter/description (often surfaces domain-specific categories beyond the usual technical/security/operational). A company-internal GenAI trained on company data is a powerful risk-analysis asset (essentially a giant searchable risk log, including non-logged incidents). Predictive analysis can flag patterns (e.g., delays correlated with a particular team) you wouldn't have logged as a risk.
- **Stakeholder/communication** — check **tone/sentiment** before sending (it has no bad days and isn't biased); remove **weasel words** to force concrete data; analyze a stakeholder's message sentiment if reading tone is hard (helpful for people on the autism spectrum). Don't use it to mass-personalize stakeholder emails — that's cold-outreach behavior, not relationship-building.
- **Bridging gaps** — ramp up on a new domain or language by conversing with it (ask clarifying questions, demand references to fact-check). It doesn't know everything (only its training data) — pair it with a search engine, which knows "everything" but needs you to know *what* to ask.

## Prompt engineering

How to ask for the most accurate, detailed answer:
- **Be concise** — more words = more room for misinterpretation.
- **Be detailed** — state exact output expectations; explain as if to an inexperienced newcomer.
- **Drive toward clarity** — ask follow-ups until you reach the root of what you need.
- **Scenario/role-play** — set context by assigning the model a role ("You are a sentiment-analysis expert grading text 1–5...").

## Limitations, grounding, and ethics

- **Data quality** — large training sets mean some poor/biased data slips through, causing wrong or biased output (the cautionary tale of an unsupervised chatbot turned offensive within hours).
- **Tokenization artifacts** — words are broken into tokens, which can lose information (the "how many r's in strawberry" failure — and models will *confidently argue* wrong answers). The output sounds confident; verify.
- **Grounding** — better systems cite sources/inline references; still imperfect (cited URLs sometimes don't support the claim). Policy: **ask the tool for its sources, read them, and prefer the source over the generated text** — treat GenAI as a search mechanism. (Echoes the early-Wikipedia stance: check the sources and use the sources.)
- **Ethics / data leakage** — IP can be regurgitated from training data; anything you paste in can become training data and resurface for other users. A meaningful share of pasted enterprise data is confidential. **Obfuscate before pasting**: rename tasks/resources to single letters (keep a key), strip names and even email aliases — but keep *dates* (shifting them breaks weekend/holiday alignment). Prefer a locked-down/private company instance for sensitive work.
