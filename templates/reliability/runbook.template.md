# Runbook: <service / alert name>

- **Owner:** <team> · **Last reviewed:** YYYY-MM-DD · **Tier:** <criticality>
- **Dashboards:** <links> · **Logs:** <query/link> · **Source:** <repo>

## What this covers
The service/alert and the symptom this runbook responds to.

## Preconditions / access
Roles, credentials, VPN, and tools needed before starting. (Production actions
are **tier3** — confirm authorization per the mutation gate.)

## Triage
1. Check <dashboard/metric> — is it <threshold>?
2. Confirm blast radius: <how>.
3. Decide severity → page / escalate if <criteria>.

## Mitigation (fastest safe path first)
| Symptom | Action | Command / link |
| :--- | :--- | :--- |
| <symptom> | <action> | `<cmd>` |

## Rollback
```
<exact rollback command / steps>
```

## Verify recovery
- <metric back to normal> · <synthetic check passes>

## Escalation
- Primary: <oncall> · Secondary: <…> · Vendor/SME: <…>

## Related
- Postmortems: <links> · Architecture: <PDR/ADR>
