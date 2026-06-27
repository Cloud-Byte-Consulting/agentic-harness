# Claude Code Jira MCP setup

```markdown
# Claude Code Jira MCP setup (official Atlassian Remote MCP)
claude mcp add --transport sse jira-server https://mcp.atlassian.com/v1/sse

# Alternatively, use the community Jira MCP server (requires an API token):
# claude mcp add jira -- npx -y @modelcontextprotocol/server-jira

# Auth: the official MCP uses OAuth (browser flow on first /mcp).
# For the community server, set:
#   JIRA_BASE_URL=https://yourinstance.atlassian.net
#   JIRA_EMAIL=you@example.com
#   JIRA_API_TOKEN=<token from https://id.atlassian.com/manage-profile/security/api-tokens>

# Then open Claude Code and run:
/mcp

# Follow the Atlassian authentication flow and enable the Jira tools for the session.
```

## Workflow setup (once per project)

Jira is workflow-based: status changes require a permitted **transition**, not a
direct status write. The runner calls `getTransitions` (or `list-transitions`)
on the issue, finds the transition whose target matches the desired OE state,
and applies it. A transition not permitted from the current status is a hard
failure — the runner posts `AGENT FAILED` with the reason and stops.

Create (or map to existing) these five statuses in your project workflow, and
ensure the transitions below are permitted:

| Open Engine state  | Recommended Jira status | Fallback name  |
|--------------------|-------------------------|----------------|
| Agent Todo         | Agent Todo              | To Do          |
| Agent Working      | Agent Working           | In Progress    |
| Agent Review       | Agent Review            | In Review      |
| Agent Needs Input  | Agent Needs Input       | Blocked        |
| Agent Done         | Agent Done              | Done           |

Required transitions (at minimum):

```
Agent Todo        → Agent Working      (runner claims)
Agent Working     → Agent Review       (runner submits)
Agent Working     → Agent Needs Input  (runner blocks / human hold)
Agent Needs Input → Agent Working      (runner resumes after human answer)
Agent Review      → Agent Done         (reviewer accepts)
Agent Review      → Agent Todo         (reviewer rejects, re-queues)
Agent Review      → Agent Needs Input  (reviewer needs more info)
```

Add these in **Project settings → Workflows** for the relevant issue type.
If you reuse existing statuses (e.g. "In Progress" for Agent Working) record
the mapping in your operator's `private-context.md` so the runner knows which
Jira status name corresponds to which OE state.

## Create the `agent-instructions` label

Run once per project (replace PROJECT_KEY and adjust the base URL):

```bash
JIRA_BASE_URL=https://yourinstance.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_API_TOKEN=<your-api-token>
PROJECT_KEY=PROJ

# Create the eligibility label via the Jira REST API
curl -s -u "$JIRA_EMAIL:$JIRA_API_TOKEN" \
  -X POST "$JIRA_BASE_URL/rest/api/3/label" \
  -H "Content-Type: application/json" \
  -d '{"name":"agent-instructions"}' | jq .

# Labels in Jira are free-text; the above ensures it exists in the instance.
# Attach it to issues via the UI or the MCP addLabel tool.
```

## Connectivity check

```bash
# Verify the MCP can read an issue (replace ISSUE-KEY):
# In Claude Code, after /mcp:
#   "Use the Jira MCP to get the details of issue PROJ-1"
# Or via REST:
curl -s -u "$JIRA_EMAIL:$JIRA_API_TOKEN" \
  "$JIRA_BASE_URL/rest/api/3/issue/PROJ-1?fields=summary,status,labels,assignee" \
  | jq '{summary:.fields.summary, status:.fields.status.name, labels:.fields.labels}'
```

A successful response returns the issue `summary`, `status`, and `labels`.
The join key for this provider is `jira:PROJ-1` (Jira keys are instance-unique;
no repo qualifier is needed).
