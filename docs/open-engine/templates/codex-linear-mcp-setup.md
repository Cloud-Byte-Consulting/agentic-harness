# Codex Linear MCP setup

```markdown
# Recommended Codex local MCP setup
codex mcp add linear --url https://mcp.linear.app/mcp

# If this is your first remote MCP in Codex, enable remote MCP in ~/.codex/config.toml:
[features]
rmcp_client = true

# Manual Codex config alternative:
[mcp_servers.linear]
url = "https://mcp.linear.app/mcp"

# Then authenticate:
codex mcp login linear
```
