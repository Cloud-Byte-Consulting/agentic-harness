# Token-Efficient Output (TEO) 🪙📐

The `air` CLI emits **TEO** — a line-oriented, indentation-structured format that
declares repeated structure once and drops JSON's per-value punctuation, so an agent
reads `air` output cheaply and parses it unambiguously. **TEO is the default**; pass
`--human` (or `--format human`, or `AIR_HUMAN=1`) for human-readable output.

```bash
air status                 # TEO (default)
air install --profile team-alpha
air personas list
air targets list
air skills list
air version

air status --human         # opt out to human-readable
```

Example (`air status --teo`):

```
description: AIR manifest components
harness: AIR
release: "2026.06"
count: 6
components[6]{id,kind,layer,lifecycle}:
  urn:air:cbc:discovery:ard,pipx,platform,tool
  urn:air:cbc:proxy:cachy,oci,platform,service
  …
```

The field schema (`{id,kind,layer,lifecycle}`) is declared once; each row is
positional values — no repeated keys, no braces, no quotes unless a value needs them.

## Implementation & validation

- **Emitter + parser + validator:** the external [`truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo`](https://truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo) module implements
  `Encode`, `Parse`, and `Validate` per the TEO grammar (typed `null`/`true`/`false`/
  numbers, quote-only-when-needed strings with `\"`/`\n` escaping, `count: e of t total`
  metadata, `help[n]` blocks, empty-state `count: 0`).
- **Round-trip oracle:** the spec's correctness rule — `parse(emit(data)) == data` —
  is a test. The TEO suite encodes known structures (incl. comma'd titles and null
  cells), parses them back, and asserts equality on the typed values.
- **Per-command conformance:** the cmd suite runs each `--teo` command, `teo.Parse`s
  the output (must succeed), and asserts the expected blocks/fields/counts. So
  "the output meets TEO" is a green test, not a claim.

Run it: `make ginkgo` (or `go test ./cmd/`).

> The canonical grammar lives with the **`teo`** project; this page documents how
> `air` conforms to it. Commands not yet TEO-enabled fall back to human-readable text.
