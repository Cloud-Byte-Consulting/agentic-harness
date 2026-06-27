# air workspace integration plan

**Status: Proposed (analysis).** Generated 2026-06-23 from a deep, multi-agent
review of all seven `air/workspace` projects. This is the plan-of-record for
bringing the projects together under `air` while keeping each one independently
shippable. It records the target architecture, the current state of every
cross-project seam, a severity-ranked gap matrix, and a phased plan.

Findings here were spot-checked against code; the four highest-stakes ones were
verified directly (teo field-name escaping, token-dashboard Dockerfile, the
role-router bash gap, the Sentry proxy stub).

---

## Progress — 2026-06-23 (Phase 0 underway)

**Done since this plan was written:**
- **teo published `v0.1.0`** on the Cloud-Byte-Consulting Gitea; module path renamed to
  `truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo` (served via Gitea `go-import`
  meta over HTTPS) — **no GitHub**.
- **air resolves teo as a versioned module** — dropped the relative `replace`, `go.sum`
  populated. Resolution needs `GOPRIVATE=truenas-scale-1.tail5a208d.ts.net` + a git
  `insteadOf` rewrite to the `:30008` endpoint. The standalone-build blocker is resolved for
  authenticated consumers.
- **Manifest reconciled** — Cachy pinned `~0.1` to match its tag.
- **8 doc reconciliations** landed (stack-doc Policy layer; teo_output paths; Sentry
  OPA-overlay + proxy-status; ARD ADR-0007 erratum; role-router links; Cachy proxy framing;
  token-dashboard phantom paths).
- **All 7 repos tagged** (teo/agentic-harness/Sentry/Cachy/token-dashboard `v0.1.0`,
  ARD `v0.9.0`, role-router `v1.4.0`) and pushed — ARD is a read-only mirror (local tag only).

**Phase 0 fixes completed 2026-06-23 (second pass):**
- **teo** — converter now escapes `,`/`}` in block field names so converted TEO re-parses; tagged **v0.1.1**.
- **token-dashboard** — added `public/.gitkeep` so the Docker `COPY /app/public` (and CI image push) succeed.
- **copilot-role-router** — added `package.json` `bin` (`role-router` → `core/cli.mjs`) + `exports` (stable import contract for air); 50 tests still green.
- **air CI** — authenticated private-teo fetch wired into both Go jobs (`GOPRIVATE`/`GOINSECURE` + git `insteadOf` to the `:30008` endpoint).
- **ARD** — `pyproject.toml` authored to expose `conformance-test` for `kind:pipx`. **Uncommitted** — ARD is a read-only mirror; apply this on the canonical repo (a local commit would be clobbered on sync).

- **Cachy** — module path renamed `cachy` → `…/Cloud-Byte-Consulting/Cachy` (Gitea-hosted, `v0.1.0` re-pointed); `go get` resolution proven from a fresh module. `composable-go-api.md` updated; the hardcoded Windows `replace` removed. (One pre-existing, rename-unrelated test fails: `TestResolvePathsLinuxFallsBackToHome` hardcodes `/home/alex`.)

**Still open:**
- **air CI** uses the built-in **`GITEA_TOKEN`** (auto-provided per run; must have cross-repo read for teo, else swap to a PAT) + runner HTTPS(443) reachability to the Gitea for the go-import meta — verifies on the next CI run.

---

## 1. Target architecture — `air` as a contract-driven composition hub

`air` never code-imports the runtime projects. It binds them through **four
stable contracts** plus one packaging/lifecycle plane. The five documented
layers gain a sixth, cross-cutting **Policy/Security** layer (Agentic-Sentry),
which the current stack doc omits.

| Layer | Project | Question it owns |
|---|---|---|
| Discovery | Agentic-Resource-Discovery | "What resources exist, how do I find them?" |
| Orchestration | copilot-role-router | "How do agents collaborate without shipping broken work?" |
| Optimization | Cachy | "How do I cut token cost in-flight?" |
| Observability | token-dashboard | "Where is the cost actually going?" |
| Output | teo | "How do agents emit results without inflating context?" |
| **Policy/Security** | **Agentic-Sentry** | **"What is each agent actually allowed to do?"** |

### The four shared contracts (coupling is through these, not source)

1. **TEO output** (`github.com/cloud-byte/teo`) — the *only* sanctioned cross-repo
   **code** dependency, and only for Go consumers (`air` today). For non-Go
   consumers (role-router, token-dashboard) TEO is a **text/wire** contract (the
   grammar in `teo/teo-format.md`): emit/parse strings, never import Go.
2. **ARD discovery** — `/.well-known/ai-catalog.json` + registry `POST /search`,
   keyed by the `urn:air:<publisher>:<ns>:<name>` identity scheme already used in
   the harness manifest. How an endpoint advertises itself and how orchestrators
   resolve one.
3. **OPA policy decision** — `POST {input:{server_name,tool_name,arguments,
   subject_id,subject_email,provider,groups}} -> {allow,reason}` at
   `data.mcp.auth.decision`, with the data overlay `data.mcp.auth.config.group_rbac`.
   Any caller can defer authz to this HTTP contract without linking Sentry.
4. **Config injection** — `air bootstrap`/`agents-link` writes each editor's MCP
   client config pointing tool traffic at the Sentry gateway URL (and optionally
   the Cachy proxy base URL). This is the seam that funnels orchestrated agents
   through the policy boundary.

### Independence rule

> A project may be **composed** by `air` (run as a service, have endpoints
> injected, be discovered via the manifest) but must always **build, test, and
> run with no sibling repo present.** The only sanctioned cross-repo *source*
> dependency is `air → teo`, as a published, versioned Go module.

---

## 2. Seam status — what is actually wired today

`wired` = live code path · `partial` = named in manifest/docs, no runtime · `missing` = intended, no code.

| From → To | Mechanism | Status |
|---|---|---|
| air → teo | Go import for all `--teo` output (8 cmd files) | **wired** (but standalone-broken: relative `replace`, no tag) |
| air → pod-bundle | skills projection + persona scaffolding | **wired** |
| Sentry → OPA | `queryOPA()` decision contract; group-RBAC overlay | **wired** (26 rego tests) |
| Sentry → Entra/Graph | client-credentials bootstrap + group sync | **wired** (untested) |
| Cachy → upstream LLMs | transparent HTTP/SSE proxy | **wired** (pass-through only; compression *not* in path) |
| role-router → harnesses | shared `core/*.mjs` across 5 tools | **wired** |
| token-dashboard → local logs | `~/.claude`, `~/.codex`, chat exports → daily-burn JSON | **wired** |
| air → ARD / Cachy / token-dashboard / role-router | named in `harness.manifest.yaml` | **partial** (install is plan-only) |
| air → Sentry (Seam ①) | bootstrap writes MCP config → gateway URL | **missing** (Sentry not even in manifest) |
| Sentry → backend MCP servers | "forward only on allow" proxy | **missing** (`handleToolsCall` is a stub) |
| Sentry → ARD (Seam ②) | gateway self-registers in catalog | **missing** (ARD ships no runnable registry) |
| air → ARD (runtime) | resolve endpoint via `POST /search` | **missing** |
| role-router → Sentry/OPA | output-guard defers to policy decision | **missing** (overlapping guards) |
| role-router → teo | emit verdict/dispatch as TEO text | **missing** |
| role-router → token-dashboard | per-role usage into daily-burn lane | **missing** |
| Cachy → Sentry/OPA | pre-forward policy hook | **missing** |
| Cachy → token-dashboard | savings telemetry into daily-burn | **missing** |

**Takeaway:** one live code seam (air→teo), the rest of the integration is
contracts-on-paper. The two highest-value missing seams are the **policy
chokepoint** (air→Sentry config injection + Sentry's proxy) and **real install
providers** (turning the plan-only `air install` into actual composition).

---

## 3. Gap matrix (severity-ranked)

### High

| Project | Gap | Recommendation |
|---|---|---|
| agentic-harness | Standalone build broken: `teo v0.0.0` + relative `replace`, no `go.sum` entry | Tag teo v0.1.0; versioned require; move replace into a root `go.work` |
| agentic-harness | In-flight teo migration uncommitted; `docs/teo_output.md` points at deleted `air/internal/teo` | Commit migration; fix the doc |
| agentic-harness | Policy layer absent from manifest **and** stack doc | Add Sentry component + a Policy layer section |
| agentic-harness | Seam ① over-claimed (Sentry doc says `air bootstrap` writes MCP config; it doesn't) | Implement MCP-config injection, gated by a flag |
| teo | No semver tag → every consumer forced into relative replace; no CI | Cut v0.1.0; add vet/build/test/gofmt CI |
| Agentic-Sentry | "forward only on allow" proxy is a stub; `ARD_MCP_URL` never read | Build the v2 proxy; until then soften the diagram |
| Agentic-Sentry | No Go CI, zero tests for gateway/admin/Entra surface | Add CI + httptest-based gateway tests |
| ARD | No `pyproject.toml` → manifest `kind:pipx` can never succeed | Add pyproject exposing `conformance-test` console_script |
| ARD | Schema contradicts spec: `attestation.mediaType` REQUIRED but examples omit it | Make mediaType optional; CI every in-spec example |
| Cachy | Advertised compression/CCR not wired into proxy (`upstream.go` is `io.Copy`) | Wire compress pipeline behind `--compress`; reconcile README |
| copilot-role-router | Read-only roles not enforced for bash (`SHELL_TOOLS` omits bash) | Add shell/exec names to a readonly-deny set |
| token-dashboard | Broken Docker build (`COPY /app/public`, no `public/`) blocks the only air seam | Add `public/.gitkeep`; unblocks CI + OCI |

### Medium / Low (summary)

- **agentic-harness** — install is plan-only for all providers; README front-matter
  stale ("implementation pending" while CLI is done); `persona.yaml` selection
  strings dangle against real skill dirs (no validator).
- **teo** — converter emits **invalid TEO** when a key contains `,` or `}` (field
  names not sanitized — *confirmed*); integer-float and `count: … total` type drift.
- **Agentic-Sentry** — `agent_security_opa.md` tells users to hardcode group IDs in
  rego, contradicting the overlay model; `MCP-Session-Id` required but unbound.
- **ARD** — no runnable reference registry; URN regex diverges from schema; 4
  conflicting version numbers; stale ADR-0007 (`urn:ai`).
- **Cachy** — admin/telemetry dead in production; bare module path `cachy` blocks `go get`.
- **role-router** — task classifier effectively non-functional; CI never runs the
  50 tests; no `package.json` exports/bin; broken README links.
- **token-dashboard** — zero tests/build gate; `npm run data` clobbers tracked
  sample; dead exact columns; no sibling-ingestion seam.

---

## 4. Phased plan

Dependency-ordered. Effort: S/M/L.

**Phase 0 — Unblock & reconcile (foundation)**
- Tag+publish **teo v0.1.0** + CI; switch air to a versioned require; move the
  relative replace into a root `go.work`; commit the `internal/teo` deletion. *(M)*
- Fix `docs/teo_output.md`; rewrite agentic-harness README front-matter; add a root LICENSE. *(S)*
- Fix token-dashboard Dockerfile (`public/.gitkeep`); fix `next lint`; add a CI build/tsc gate. *(S)*
- Add ARD `pyproject.toml`; stable module path for Cachy; `package.json` exports+bin for role-router. *(M)*

**Phase 1 — Manifest & install providers (discover + configure)**
- Add Agentic-Sentry to `harness.manifest.yaml` as `urn:air:cbc:policy:sentry` (oci, service, platform). *(S)*
- Implement real install providers (archive/oci/pipx/binary); materialize
  `~/.air/services/` + write `~/.air/endpoints.yaml`; trim doctor. *(L)*

**Phase 2 — Policy chokepoint (highest-value runtime seam)**
- Implement MCP-config injection in `air bootstrap` → Sentry gateway (Seam ①). *(M)*
- Implement Sentry's v2 "forward only on allow" proxy + Go CI + gateway tests; fix the group-overlay doc. *(L)*
- Code-enforce read-only roles in role-router (bash); resolve the output-guard ↔ Sentry boundary. *(M)*

**Phase 3 — Discovery seam**
- Ship a minimal ARD reference registry + Sentry catalog-publisher; air resolves
  the gateway via `POST /search` (Seam ②); fix the ARD mediaType schema/spec contradiction. *(L)*

**Phase 4 — Optimization seam**
- Wire Cachy compression/CCR behind `--compress`; start `cachy admin` with a shared
  TelemetryRecorder; air injects the Cachy proxy base URL. *(L)*

**Phase 5 — Observability & output ingestion**
- Define a daily-burn ingestion contract in token-dashboard; emit per-role usage
  (role-router) + savings (Cachy) into it; add a TEO text emitter to role-router. *(L)*

**Phase 6 — Hardening & validation**
- `air personas check` validator; teo field-name fix + round-trip tests;
  token-dashboard lane-invariant tests; role-router classifier/CLI tests; reconcile
  remaining stale docs. *(M)*

---

## 5. Documentation updates (tracked separately)

The doc reconciliations that are true *today* (independent of the code work
above) are applied as part of this analysis. See the companion commit/PR for:
the stack doc Policy layer, `teo_output.md`, the Sentry `agent_security_opa.md`
overlay fix, ARD ADR-0007 erratum, role-router README links, and the
token-dashboard SKILL/data-contract path fixes. Updates that assert *future*
state (e.g. "teo is published vX") are deferred until the corresponding Phase
lands.
