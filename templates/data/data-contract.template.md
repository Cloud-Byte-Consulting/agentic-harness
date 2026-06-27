# Data Contract: <dataset / table / topic name>

- **Owner:** <team> · **Status:** Draft | Active | Deprecated · **Version:** vN
- **Source system:** <…> · **Location:** <warehouse.schema.table / topic / path>
- **Classification:** public | internal | confidential | PII/PHI

## Purpose
What this dataset represents and the questions it answers.

## Schema
| Field | Type | Nullable | Description | Example |
| :--- | :--- | :--- | :--- | :--- |
| id | string | no | primary key | `a1b2` |
| event_ts | timestamp | no | event time (UTC) | `2026-06-21T00:00:00Z` |

- **Primary key:** <field(s)> · **Partitioning:** <field> · **Grain:** one row per <…>

## Semantics & quality guarantees
- **Freshness / SLA:** updated every <interval>; available by <time>.
- **Completeness:** <expected volume / null rules>.
- **Uniqueness:** <key uniqueness guarantee>.
- **Lineage:** upstream <sources> → this dataset → downstream <consumers>.

## Producer / consumer obligations
- **Producer:** schema changes are additive; breaking changes bump major version + notice.
- **Consumer:** pin to a version; tolerate added columns.

## Change policy
SemVer for the schema. Breaking change process: <notice period, comms, dual-write>.
