# Self-Hosted Docker MCP Gateway 🐳🚪

Agents are only as capable as the tools they can reach. The **Docker MCP Gateway** runs MCP (Model Context Protocol) servers as Docker containers and exposes them — aggregated behind a single endpoint — to any MCP client (Claude Code, Cursor, VS Code, etc.). It is the open-source engine behind Docker Desktop's "MCP Toolkit," and it runs on **native Linux without Docker Desktop**. This guide sets it up as a persistent, always-on gateway.

The payoff: one gateway, one config line per editor, N containerized tool servers — each isolated, version-pinned, and disposable.

---

## 1. Architecture

```
  ┌────────────┐   HTTP :8811 (or stdio)   ┌─────────────────┐   spawns   ┌──────────────┐
  │ MCP client │ ────────────────────────▶ │  docker mcp     │ ─────────▶ │ fetch (ctr)  │
  │ (editor /  │   Bearer token (HTTP)     │  gateway        │            │ git   (ctr)  │
  │  agent)    │ ◀──── aggregated tools ── │  (profile-based)│            │ memory(ctr)  │
  └────────────┘                           └─────────────────┘            └──────────────┘
```

*   **Plugin (`docker mcp`):** a Docker CLI plugin that manages catalogs, profiles, secrets, and the gateway.
*   **Catalog:** an index of available MCP servers (image, tools, required secrets). Pulled as an OCI artifact.
*   **Profile:** a named subset of catalog servers you actually want to run (e.g. `main`).
*   **Gateway:** the long-lived process that starts the profile's servers as containers and exposes their tools over **stdio** (per-client) or a **streaming HTTP port** (shared server).

---

## 2. Prerequisites

*   Docker Engine (rootful) installed. On Arch/CachyOS: `pacman -S docker`.
*   A running daemon and a user in the `docker` group:

```bash
sudo groupadd -f docker
sudo usermod -aG docker "$USER"
sudo systemctl enable --now docker.service
# Group membership is set at login — re-login (or `newgrp docker`) so your
# shell and the systemd --user manager can reach the socket without sudo.
```

> **Note:** processes that predate the group change won't have it. For one-off commands, `sg docker -c '<cmd>'` grants the group on demand without a re-login (used in the service unit below).

---

## 3. Install the plugin

Docker CLI plugins live in `~/.docker/cli-plugins/`. Drop the release binary there (newest, no build deps):

```bash
mkdir -p ~/.docker/cli-plugins
VER=v0.42.3   # check github.com/docker/mcp-gateway/releases for the latest
curl -fsSL "https://github.com/docker/mcp-gateway/releases/download/$VER/docker-mcp-linux-amd64.tar.gz" \
  | tar -xz -C ~/.docker/cli-plugins
chmod +x ~/.docker/cli-plugins/docker-mcp
docker mcp version          # → v0.42.3
```

On Arch/CachyOS the AUR package `docker-mcp` is an alternative, but it may lag behind upstream.

---

## 4. Pick servers: catalog + profile

Pull the curated catalog (clean names; ~300 servers), then build a profile from the ones you want:

```bash
# Official curated catalog (OCI artifact)
docker mcp catalog pull mcp/docker-mcp-catalog:latest
docker mcp catalog server ls mcp/docker-mcp-catalog:latest --format human   # browse
docker mcp catalog server ls mcp/docker-mcp-catalog:latest -f name=git      # filter

# A profile = the servers you want to expose. Reference syntax:
#   catalog://<catalog-name>/<server1>+<server2>+...
docker mcp profile create --name "Main" --id main \
  --server "catalog://mcp/docker-mcp-catalog/fetch+duckduckgo+git+memory+time+markitdown+sequentialthinking"

docker mcp profile show main          # inspect resolved servers
```

*   Servers flagged with `secrets` need credentials before they work: `docker mcp secret set <KEY>` (stored in the OS keychain), or pass `--secrets <file>.env` to `gateway run`.
*   The huge **community registry** (`--from-community-registry registry.modelcontextprotocol.io`) is also available but uses verbose reverse-DNS names — prefer the curated catalog for hand-picking.

Validate the whole profile without listening:

```bash
docker mcp gateway run --profile main --dry-run
```

---

## 5. Run it as a persistent service

Two transports:

*   **stdio** (default) — the client launches the gateway itself, per session. Simplest, no port, no token.
*   **streaming HTTP** — a long-lived server multiple clients share. Use this for an always-on "gateway server."

For the HTTP server, protect localhost with a bearer token:

```bash
mkdir -p ~/.config
printf 'MCP_GATEWAY_AUTH_TOKEN=%s\n' "$(openssl rand -hex 24)" > ~/.config/docker-mcp-gateway.env
chmod 600 ~/.config/docker-mcp-gateway.env
```

Create a **systemd user service** at `~/.config/systemd/user/docker-mcp-gateway.service`:

```ini
[Unit]
Description=Docker MCP Gateway (streaming, port 8811, profile "main")
After=docker.service
Wants=docker.service

[Service]
Type=simple
EnvironmentFile=%h/.config/docker-mcp-gateway.env
ExecStartPre=/bin/sh -c 'until test -S /var/run/docker.sock; do sleep 1; done'
# `sg docker` grants the docker group even before a full re-login picks it up.
ExecStart=/usr/bin/sg docker -c '/usr/bin/docker mcp gateway run --transport streaming --port 8811 --profile main'
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

Enable it (and `linger` so it survives logout/reboot):

```bash
systemctl --user daemon-reload
systemctl --user enable --now docker-mcp-gateway.service
loginctl enable-linger "$USER"
```

The gateway logs its endpoint: `Gateway URL: http://localhost:8811/mcp`.

---

## 6. Connect editors and agents (JSON)

Every client points at the **one** gateway endpoint. Config file location and schema vary:

| Client | Config file | Top key | Remote (HTTP) |
|--------|-------------|---------|---------------|
| Cursor | `~/.cursor/mcp.json` or `.cursor/mcp.json` | `mcpServers` | ✅ `url` + `headers` |
| VS Code | `.vscode/mcp.json` or user `mcp.json` | `servers` | ✅ `type:"http"` |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` | ✅ `serverUrl` |
| Claude Desktop | `claude_desktop_config.json` | `mcpServers` | stdio only |
| Zed | `settings.json` | `context_servers` | mostly stdio |

**HTTP — Cursor / Windsurf style:**

```json
{
  "mcpServers": {
    "docker-mcp": {
      "url": "http://localhost:8811/mcp",
      "headers": { "Authorization": "Bearer PASTE_TOKEN_HERE" }
    }
  }
}
```

**HTTP — VS Code style** (note `servers` + `type`):

```json
{
  "servers": {
    "docker-mcp": {
      "type": "http",
      "url": "http://localhost:8811/mcp",
      "headers": { "Authorization": "Bearer PASTE_TOKEN_HERE" }
    }
  }
}
```

**stdio — universal, no token** (client launches its own gateway):

```json
{
  "mcpServers": {
    "docker-mcp": {
      "command": "docker",
      "args": ["mcp", "gateway", "run", "--profile", "main"]
    }
  }
}
```

**Annotated stdio fields** — the object you nest under the client's top key (`mcpServers` / `servers` / `context_servers`). Comments are JSONC-only (VS Code, Cursor, Zed); strip them for Claude Desktop's strict JSON:

```jsonc
{
  // The server's name — your label, shown in the client
  "docker-mcp": {
    // The command which runs the MCP server (must be a non-empty string)
    "command": "docker",
    // The arguments to pass to the MCP server
    "args": ["mcp", "gateway", "run", "--profile", "main"],
    // Environment variables to set for the server process
    "env": {}
  }
}
```

**Claude Code** is easiest via CLI:

```bash
TOKEN=$(grep -oP 'MCP_GATEWAY_AUTH_TOKEN=\K.*' ~/.config/docker-mcp-gateway.env)
claude mcp add --transport http docker-mcp http://localhost:8811/mcp \
  --header "Authorization: Bearer $TOKEN" --scope user
```

**Auto-writer:** the plugin can edit a client's JSON for you (stdio mode):

```bash
docker mcp client connect cursor      # or vscode, claude-desktop, zed, cline, continue, kiro…
docker mcp client ls                  # connection status per client
```

> **Secret hygiene:** never commit a bearer token into a project-scoped config (`.vscode/`, `.cursor/`). Use the global/user config, or use the stdio transport (no token at all).

---

## 7. Verify end to end

```bash
TOKEN=$(grep -oP 'MCP_GATEWAY_AUTH_TOKEN=\K.*' ~/.config/docker-mcp-gateway.env)
H=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json"
   -H "Accept: application/json, text/event-stream")

# Reject without a token (expect 401):
curl -s -o /dev/null -w '%{http_code}\n' -X POST http://localhost:8811/mcp "${H[@]:1}" -d '{}'

# Initialize handshake (expect 200 + serverInfo):
curl -s -X POST http://localhost:8811/mcp "${H[@]}" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"c","version":"1"}}}'
```

A healthy gateway returns `"serverInfo":{"name":"Docker AI MCP Gateway",...}` and, after `tools/list`, the union of every profile server's tools.

---

## 8. Manage & troubleshoot

```bash
systemctl --user status docker-mcp-gateway      # state
systemctl --user restart docker-mcp-gateway     # apply profile changes
journalctl --user -u docker-mcp-gateway -f      # follow logs

docker mcp profile server add main \
  --server catalog://mcp/docker-mcp-catalog/wikipedia-mcp   # add a server, then restart
docker ps --filter label=docker-mcp=true                   # running server containers
```

| Symptom | Cause / fix |
|---------|-------------|
| `permission denied … /var/run/docker.sock` | Process lacks the `docker` group. Re-login, or wrap in `sg docker -c '…'`. |
| `No catalogs found` | Run `docker mcp catalog pull mcp/docker-mcp-catalog:latest`. |
| `is not a self-describing image` | Don't reference raw `docker://mcp/<x>` images; use `catalog://…` references. |
| `401` from the gateway | Missing/incorrect `Authorization: Bearer` header. |
| Service dies after logout | `loginctl enable-linger "$USER"`. |
| `docker-credential-notfound` warnings | Harmless for public images (no Docker Hub login configured). |

---

## 9. Security notes

*   The streaming port binds all interfaces by default (`*:8811`). The bearer token gates access; for strict local-only use, firewall the port or bind to loopback.
*   Containerized servers are isolated, but `gateway run` flags tighten them further: `--block-network`, `--block-secrets` (on by default), `--cpus`, `--memory`, `--verify-signatures`.
*   Treat the auth token and any server secrets as credentials — keep them out of version control.
