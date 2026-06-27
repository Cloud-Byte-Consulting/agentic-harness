# Claude Code GitHub MCP setup

```markdown
# Claude Code GitHub MCP setup
claude mcp add --transport sse github-server https://api.githubcopilot.com/mcp/v1/sse

# Alternatively, use the community GitHub MCP server (requires a PAT):
# claude mcp add github -- npx -y @modelcontextprotocol/server-github

# Token scope needed (classic PAT or fine-grained):
#   repo  (read/write issues, labels, comments)
#   — no extra scopes required for Open Engine use

# Then open Claude Code and run:
/mcp

# Follow the GitHub authentication flow and enable the GitHub tools for the session.
```

## Create the required labels in each repo

Run once per repo (replace OWNER/REPO):

```bash
GH_REPO=OWNER/REPO

# Status labels (exactly one oe:* label on an issue at a time)
gh label create "oe:todo"        --repo $GH_REPO --color "0075ca" --description "Agent Todo — eligible to claim"
gh label create "oe:working"     --repo $GH_REPO --color "e4e669" --description "Agent Working"
gh label create "oe:review"      --repo $GH_REPO --color "d93f0b" --description "Agent Review"
gh label create "oe:needs-input" --repo $GH_REPO --color "cc317c" --description "Agent Needs Input (blocked or human hold)"
gh label create "oe:done"        --repo $GH_REPO --color "0e8a16" --description "Agent Done"

# Eligibility label
gh label create "agent-instructions" --repo $GH_REPO --color "bfd4f2" --description "Open Engine task issue"
```

## Connectivity check

```bash
# Verify the MCP server can see issues (replace OWNER/REPO and NUMBER):
gh issue view NUMBER --repo OWNER/REPO --json title,labels,state
```

A successful response returns the issue's `title`, `labels`, and `state`.
The join key for this provider is `github:OWNER/REPO#NUMBER`.
