# Data Quality Report: <dataset>

- **Date / window:** YYYY-MM-DD (<window>) · **Owner:** <team>
- **Contract:** <link to data-contract> · **Run:** <pipeline/job id>

## Verdict
✅ Pass | ⚠️ Degraded | ❌ Fail — one line on overall health.

## Dimensions
| Dimension | Check | Result | Threshold | Status |
| :--- | :--- | :--- | :--- | :--- |
| Completeness | null rate of `<field>` | <x>% | < <y>% | ✅ |
| Uniqueness | dup rate on PK | <x>% | 0% | ✅ |
| Freshness | lag vs SLA | <t> | < <sla> | ⚠️ |
| Validity | values in allowed set | <x>% | 100% | ❌ |
| Accuracy/consistency | row count vs source | Δ<x>% | < <y>% | ✅ |

## Failures & anomalies
- <check> failed: <evidence — sample rows, query, count>.

## Impact & action
- **Affected downstream:** <consumers> · **Severity:** …
- **Action items:** <fix / backfill / quarantine> — owner, due.

## Trend
How key metrics moved vs prior runs (regressions flagged).
