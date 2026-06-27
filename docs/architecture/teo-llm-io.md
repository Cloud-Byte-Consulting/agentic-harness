# TEO LLM I/O — carve-out reference

TEO (`teo/`) is the default I/O grammar for model-facing text in the Open Engine
governed loop. This document defines what IS and IS NOT TEO, and how a runner or
agent calls the library.

## Rule of thumb

> If the next reader is the **model**, use TEO.
> If the next reader is a **human glancing at Linear**, use prose.
> If the next reader is a **tool, protocol, or identity standard**, keep its mandated format.

## NOT-TEO surfaces

These surfaces must use their native format. Never wrap them in TEO.

| Surface | Mandated format | Why |
| :--- | :--- | :--- |
| MCP `tools/call` (Linear MCP, Sentry gateway, any MCP server) | JSON-RPC | the MCP server's wire contract |
| OPA decision input / output (`data.mcp.auth.decision`) | JSON | Rego input schema |
| Omnigent session API (`POST /v1/sessions`, session events) | JSON | server contract |
| Tool / function-call arguments (harness `Bash`, `Read`, `Edit`, etc.) | harness-native JSON schema | the model's tool interface |
| Entra JWT claims | JWT | identity standard |
| Human-read Linear comments: status ledger lines, receipt verbs | Markdown prose | ledger is the human-facing projection of the runtime plane — dense TEO defeats it |

Receipt verbs that stay prose: `AGENT CLAIMED`, `AGENT DONE`, `AGENT BLOCKED`,
`AGENT HUMAN HOLD`, `HUMAN ANSWERED`, `AGENT FAILED`, `holding ENG-NNN`,
`completed ENG-NNN`. These are written to Linear for a human to scan; they are
not model-consumed artifacts.

## TEO surfaces

| Surface | Direction | Notes |
| :--- | :--- | :--- |
| Task record assembled for the model | input (context window) | task title, labels, body, linked context — render as TEO scalars/records |
| Status-ledger state snapshot fed to the model | input | current ledger entries, not the raw comment stream |
| Retrieved standing skills / standing context | input | field-per-scalar, block per list |
| AGENT DONE evidence body | output | the long artifact another model agent will consume; validate with `teo.Validate` |
| Inter-agent artifact payloads (reviews, summaries a second agent reads) | output | same; parse with `teo.Parse` |

## How the runner / agent calls teo

The `teo` package is dependency-free stdlib-only Go. Import path:
`truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo`

**Validate agent output** (cheapest check; non-nil error = malformed):

```go
if err := teo.Validate(agentOutput); err != nil {
    // emit AGENT FAILED or re-prompt
}
```

**Parse agent output** (when you need to extract fields):

```go
doc, err := teo.Parse(agentOutput)
if err != nil { ... }
val := doc.GetScalar("status")      // any (string/int/bool/nil)
blk := doc.FindBlock("findings")    // *teo.Item, nil if absent
```

**Build model input** (context assembly):

```go
// Block().Row() returns *BlockHandle (no String()); keep the *Document
// to render. Call String() on the Document, not the BlockHandle.
doc := teo.New().
    Scalar("task", "ENG-123").
    Scalar("title", issueTitle).
    Record("ledger",
        teo.KV{Key: "status", Value: "Agent Working"},
        teo.KV{Key: "claimed_at", Value: claimedAt},
    )
doc.Block("skills", "name", "scope").
    Row("code-review", "repo").
    Row("security-review", "repo")
modelInput := doc.String() // dense TEO, not JSON
```

No TEO↔JSON shim at tool boundaries. No bespoke TEO compressor — wire-level
compression is Cachy's job (deferred, API-key mode).

## Model-vs-human split rule (summary)

```
model consumer  → TEO  (context window + inter-agent artifacts)
human consumer  → prose  (Linear ledger, receipt lines)
tool/protocol   → native JSON/JWT/Rego  (never TEO)
```

This rule is instructed in the per-agent profile
(`docs/open-engine/templates/starter-private-context-file.md`, TEO I/O section)
and applies to all Open Engine agent codes.

## Checkpoint CP-2 criteria

Run one task with TEO I/O active and assert:
- `teo.Validate(agentOutput)` returns nil (clean round-trip)
- Token-dashboard shows a measurable drop vs the prose baseline
- All tool calls still emit native JSON (zero format regressions)
