# SLO: <service name>

- **Owner:** <team> · **Status:** Draft | Active · **Review cadence:** quarterly

## Service & users
What the service does and who depends on it.

## SLIs (what we measure)
| SLI | Definition (good events / valid events) | Source |
| :--- | :--- | :--- |
| Availability | successful requests / total requests | <metric> |
| Latency | requests faster than <T> ms / total | <metric> |

## SLO targets
| SLI | Target | Window |
| :--- | :--- | :--- |
| Availability | 99.9% | rolling 28d |
| Latency (p95) | < 300 ms, 99% | rolling 28d |

## Error budget
- **Budget:** <100% − SLO> over the window (e.g. 0.1% ≈ 40m/28d).
- **Policy when exhausted:** freeze risky changes; prioritize reliability work;
  escalate to <role>.

## Alerting
- **Fast burn:** <multiwindow rule> → page.
- **Slow burn:** <rule> → ticket.

## Dependencies & exclusions
- Upstream SLOs relied on: <…>
- Excluded from valid events: <maintenance windows, synthetic, …>
