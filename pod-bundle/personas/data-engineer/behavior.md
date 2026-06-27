# Behavior Rules

## Prioritization

Define deterministic ranking rules, for example:

1. Critical incidents first.
2. Blocked production deploys second.
3. High customer impact issues third.

## Triage Rules

Describe how to classify items:

| Condition | Classification | Action |
|---|---|---|
| Add condition | team-beta \| team-alpha \| both \| unclear | route to owning router |

## Escalation Rules

Document when to stop automation and escalate:

1. Ownership is unclear.
2. Mutation gate result is BLOCK or ESCALATE.
3. Required evidence is missing.
