# Stakeholder Update Email

After work ships, sends (or drafts) a short, truthful update email to the person who needs to know — a client, a producer, a collaborator, your team. The skill encodes the discipline: updates go out only when something stakeholder-visible actually changed; the email describes shipped behavior in the recipient's vocabulary, not implementation details; nothing unverified gets called done; the format stays consistent (what changed, what it means for you, what's next); and you're CC'd or shown a draft first, per your preference.

## Why Build It
Communication is the half of client and team work that agent workflows usually drop. The skill's real content isn't email mechanics — it's the rules: only when shipped, only what's true, only in their language. A consistent, honest update cadence after real changes builds more trust than any amount of polish, and making it a skill means it actually happens instead of being the thing you'll do after lunch.

## What You Need


## Prompt / Setup
```xml
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
