# Postmortem: <incident title>

> Blameless. Focus on systems and contributing factors, never individuals.

- **Incident ID:** INC-<n> · **Severity:** SEV-1 … SEV-4 · **Status:** Draft | Final
- **Authors:** <names> · **Date:** YYYY-MM-DD
- **Duration:** <detection → resolution> · **Customer impact:** <who, what, how much>

## Summary
2–3 sentences: what broke, the impact, and the fix — readable by an exec.

## Impact
- **Users / revenue / SLA:** quantify.
- **Services affected:** …

## Timeline (UTC)
| Time | Event |
| :--- | :--- |
| 00:00 | <trigger / change deployed> |
| 00:07 | <alert fired> |
| 00:35 | <mitigated> |
| 01:10 | <resolved> |

## Root cause & contributing factors
What actually happened (cite logs/metrics/dashboards). Separate the **trigger**
from the **root cause** and the **factors that amplified** it.

## Detection & response
- How was it detected (alert / customer)? Time to detect.
- What slowed diagnosis or recovery?

## What went well / what didn't
- 👍 …
- 👎 …

## Action items
| Action | Type (prevent/detect/mitigate) | Owner | Due | Ticket |
| :--- | :--- | :--- | :--- | :--- |

## Lessons
Durable takeaways worth propagating to other teams.
