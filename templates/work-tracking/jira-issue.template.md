# Jira issue template

> Fields map to Jira's create screen. Body uses Jira wiki markup (`h2.`, `*bold*`,
> `{code}`). CLI: `jira issue create -t Story -s "…" -b "$(cat issue.txt)"` (ankitpokhrel/jira-cli).

**Project:** `<KEY>` · **Issue Type:** Story | Bug | Task | Epic | Spike
**Summary:** `<imperative one-liner>`
**Priority:** Highest/High/Medium/Low · **Story Points:** `<n>`
**Components:** `<comp>` · **Labels:** `<label>` · **Fix Version/s:** `<release>`
**Epic Link:** `<KEY-###>` · **Sprint:** `<sprint>` · **Assignee:** `<user>`

----

h2. Description
*Context:* <why this exists, linked evidence>
*Scope:* <what's included / excluded>

h2. Acceptance Criteria
* *Given* <context> *when* <action> *then* <result>
* *Given* … *when* … *then* …

h2. Steps to Reproduce  _(Bug)_
# …
*Expected:* …
*Actual:* …
{code}<logs / stack trace>{code}

h2. Links
* blocks / is blocked by: <KEY-###>
* relates to: <url>
