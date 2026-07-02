<role>
You are helping me wire `tools/opa_hook.py` as the native-tool authorization gate
for whichever agentic CLIs I have installed on this machine (Claude Code, Codex,
GitHub Copilot CLI, Gemini CLI), backed by an OPA instance you deploy for me as
either a **Docker** container or into **minikube** — no other deployment target
is in scope for this walkthrough. The code, policy, and deploy scripts already
exist in this repo (`tools/opa_hook.py`, `Agentic-Sentry/mcp-policies/`,
`deploy/opa-hook/`) — your job is to deploy the backend on my chosen target,
detect which CLIs I actually have installed, wire each one's hook config
correctly, prove it works in shadow mode, and only then flip it to enforce. Do
not assume I've read the architecture docs; explain each step in plain terms as
you go. Reference docs if you need exact detail: `deploy/opa-hook/README.md`,
`docs/architecture/opa-hook.md`,
`docs/architecture/agent-integration-claude-code.md`,
`docs/architecture/agent-integration-codex.md`,
`docs/architecture/agent-integration-copilot-cli.md`,
`docs/architecture/agent-integration-gemini-cli.md`,
`docs/open-engine/templates/opa-hook-setup.md`.
</role>

<how_to_work_with_me>
- Detect before you assume: check which CLI binaries actually exist and their
  versions rather than asking me to enumerate them.
- Do exactly one CLI at a time. Verify it works in shadow mode before moving to
  the next one.
- Show me the exact file path and exact file contents you're about to write
  BEFORE writing it. Never overwrite an existing hooks config file without
  showing me the diff and getting my go-ahead — other hooks may already live
  there.
- Substitute the real absolute path to this repo's `tools/opa_hook.py` into
  every config snippet; don't leave a placeholder in a file you actually write.
- Never set `OMNIGENT_OPA_DELEGATE_MODE=enforce` without me explicitly
  confirming after a shadow-mode dry run for that specific CLI. Enforce mode
  can block real tool calls, including ones I need mid-task.
</how_to_work_with_me>

<step_0_prerequisites>
1. Confirm `python3` is on PATH.
2. Run the offline self-check: `python3 tools/opa_hook.py --self-check`. It
   must print "opa-hook self-check passed" before you touch any live config —
   this proves the hook logic itself is sound, independent of any CLI wiring
   or backend deployment.
3. Ask me how I want to run the OPA backend that answers `oe_decision`
   queries. Offer exactly two choices — **do not offer or improvise a third**
   (no bare-host `opa run`, no cloud target, no other orchestrator):
   - **Docker** — single container, fastest local loop.
     `deploy/opa-hook/docker/run.sh`
   - **minikube** — local Kubernetes.
     `minikube start` (if not running) → `kubectl config use-context minikube`
     → `deploy/opa-hook/minikube/deploy.sh`
   If I ask for anything else (a bare host process, a real/managed cluster, a
   cloud target), tell me plainly that this prompt only covers Docker and
   minikube, and point me at `deploy/opa-hook/README.md` and
   `docs/architecture/deploy/` for other options — don't invent a third path.
4. Run the chosen script. Both print a `curl` command against
   `/v1/data/mcp/auth/oe_decision` when done — run it and confirm a JSON
   response with a `result.verdict` field. If this fails, stop and tell me
   what's wrong (container/pod not ready, port-forward not running, wrong
   context) rather than proceeding to wire a hook against a dead backend. Full
   detail, including a real gotcha already hit and fixed in the minikube
   manifest (ConfigMap-as-directory causing a duplicate-rule load error): see
   `deploy/opa-hook/README.md`.
</step_0_prerequisites>

<step_1_detect_clis>
Check which of these are on PATH and get their version, comparing against the
minimum floor for each (from the architecture docs):
- `claude --version` — Claude Code. Any version (gold-standard baseline).
- `codex --version` — Codex CLI. Any version with hook support.
- `copilot --version` (or `gh copilot --version`, check which binary is
  actually present) — GitHub Copilot CLI. Needs **>=1.0.6** for the
  Claude-compatible hook contract at all; **>=1.0.62** recommended for
  hardened fail-closed-on-crash behavior. If below 1.0.6, tell me hooks won't
  work on this version and ask whether to skip it or upgrade first.
- `gemini --version` — Gemini CLI. Needs **>=0.26.0** (hooks didn't exist
  before this). If below floor, tell me and ask whether to skip or upgrade.

Report back a short table: which CLIs are present, their versions, and whether
each meets its floor. Confirm with me which ones to wire before proceeding —
don't wire all of them unprompted if I only use one or two day to day.
</step_1_detect_clis>

<step_2_wire_claude_code>
Only if Claude Code was confirmed in step 1.
1. Find or create `~/.claude/settings.json`. If it already has a `hooks` key,
   show me the existing content and propose a merge that ADDS the opa-hook
   entry to `PreToolUse` without removing anything already there.
2. Show me the exact JSON you're about to write (matcher `"*"`, command
   `python3 <absolute-path-to-this-repo>/tools/opa_hook.py`), then write it
   after I confirm.
3. Set `OMNIGENT_OPA_DELEGATE_MODE=shadow` in the shell/session Claude Code
   will run in (not enforce yet).
4. Ask me to run any real tool call in Claude Code (e.g. a Bash command). Check
   for the `opa-hook[shadow]: tool=... verdict=...` line on stderr. If it
   doesn't appear, the hook isn't firing — debug the settings.json path/syntax
   before moving on.
</step_2_wire_claude_code>

<step_3_wire_codex>
Only if Codex was confirmed in step 1. Same pattern as step 2, but the config
lives under `CODEX_HOME` as a `PreToolUse`/`PostToolUse` command hook (see
`agent-integration-codex.md` for the exact key). Show me the file and content
before writing; shadow-mode verify before moving on.
</step_3_wire_codex>

<step_4_wire_copilot_cli>
Only if Copilot CLI was confirmed in step 1 and met its version floor. Ask me
whether to use the repo-level `.github/hooks/opa.json` (loads only after folder
trust, shared with the team if committed) or the personal
`~/.copilot/hooks/opa.json` (always loads, private to me) — explain the
tradeoff in one sentence and let me choose. Show the exact JSON (from
`docs/open-engine/templates/opa-hook-setup.md`) before writing. Shadow-mode
verify: run a tool call in Copilot CLI, confirm the shadow log line.
</step_4_wire_copilot_cli>

<step_5_wire_gemini_cli>
Only if Gemini CLI was confirmed in step 1 and met its version floor. Explain
BEFORE wiring: Gemini's hook contract has no `ask`/escalate verdict, so any
policy verdict of `require_approval` will surface as a flat `deny` here, not a
prompt. Confirm I understand and accept that before writing the config. Show
the exact `~/.gemini/settings.json` JSON (event name `BeforeTool`, command
includes `--format gemini`) before writing. Shadow-mode verify: run a tool
call, confirm the shadow log line.
</step_5_wire_gemini_cli>

<step_6_enforce>
Only after EVERY CLI wired in steps 2-5 has been individually shadow-verified,
and only for the CLIs I explicitly say to enforce (I may want some in shadow
indefinitely). Per CLI, one at a time:
1. Ask me to explicitly confirm "enforce <cli name>" before changing anything.
2. Set `OMNIGENT_OPA_DELEGATE_MODE=enforce` for that CLI's environment.
3. Run the three-way smoke test and report the actual outcome of each:
   - A tool call you expect OPA to `allow` — should proceed with no visible
     hook interference.
   - A tool call you expect OPA to `deny` — should be blocked, with the OPA
     reason surfaced to me.
   - A tool call you expect `require_approval` — on Claude Code/Codex/Copilot
     CLI this should prompt me (`ask`); on Gemini CLI this should be a flat
     `deny` (per step 5's caveat) — confirm that's what actually happens, not
     a silent allow.
4. If any of the three don't match, the hook is not correctly in the path —
   stop and debug (check the config file was actually picked up, check
   `OMNIGENT_OPA_URL` if OPA isn't on the default port) before calling this
   CLI done.
</step_6_enforce>

<step_7_mcp_coverage_check>
For at least one enforced CLI, fire a tool call through an MCP server (a tool
named like `mcp__<server>__<tool>`), not just a native tool. Confirm the hook
still fires and the verdict is sane — this proves `_parse_tool()` is correctly
splitting the MCP tool name into `(server, tool)` for the OPA query, which
people often assume without checking.
</step_7_mcp_coverage_check>

<safety>
- Only deploy the OPA backend to Docker or minikube. Never run
  `deploy/opa-hook/minikube/deploy.sh` (or any `kubectl apply`) against a
  context that isn't literally `minikube` — the script already refuses this,
  don't work around that refusal, and don't hand-roll an equivalent command
  against a different context "to save time."
- Never set `OMNIGENT_OPA_DELEGATE_MODE=enforce` without my explicit
  confirmation, per CLI, after a successful shadow-mode dry run.
- Never write to a hooks/settings file without showing me the exact path and
  exact content first, and never silently overwrite existing hook entries.
- If OPA is unreachable at any point, tell me plainly. Do not fall back to
  `mode=off` or skip verification and imply the gate is active when it isn't.
- If a CLI is below its documented version floor, don't wire it — tell me and
  let me decide whether to upgrade or skip.
</safety>

<start_here>
Start with step 0: confirm prerequisites, run the offline self-check, start
OPA, and confirm the bundle is queryable. Report back before moving to
detection in step 1.
</start_here>
