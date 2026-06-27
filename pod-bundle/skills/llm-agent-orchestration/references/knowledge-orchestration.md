# Knowledge Orchestration for Agent Systems

How to organize the knowledge your agents reason over so retrieval is auditable, joins
are deterministic, and extraction stays affordable at scale. This is the *what knowledge
and how it is organized* half of an agent system — distinct from *how agents plan and
act*. Read this when you are designing the knowledge substrate (graph, index, document
store) behind a research agent, a RAG pipeline, a deep-research loop, or any system where
an LLM has to combine evidence across many sources and later justify its answer.

## Table of contents

1. [Agent orchestration vs knowledge orchestration](#1-agent-orchestration-vs-knowledge-orchestration)
2. [Stable identifiers as join keys](#2-stable-identifiers-as-join-keys)
3. [Semantic anchors for multimodal and cross-source linking](#3-semantic-anchors-for-multimodal-and-cross-source-linking)
4. [Typed knowledge, durable-vs-ephemeral facts, citation intent](#4-typed-knowledge-durable-vs-ephemeral-facts-citation-intent)
5. [Multi-source intent-routed retrieval with a coverage guarantee](#5-multi-source-intent-routed-retrieval-with-a-coverage-guarantee)
6. [Deterministic typed operators below an LLM composition layer](#6-deterministic-typed-operators-below-an-llm-composition-layer)
7. [Cost-factored core-then-views multi-view extraction](#7-cost-factored-core-then-views-multi-view-extraction)
8. [Weight-frozen self-improving skill libraries](#8-weight-frozen-self-improving-skill-libraries)
9. [Auditable coordinator-worker-aggregator swarm execution](#9-auditable-coordinator-worker-aggregator-swarm-execution)

---

## 1. Agent orchestration vs knowledge orchestration

**Pattern: treat "how agents act" and "what knowledge they act on" as two separate
engineering problems, and invest in the second one explicitly.**

Most multi-agent effort goes into *agent orchestration*: planning loops, tool selection,
delegation, debate, retries. The other half — *knowledge orchestration*, deciding what
knowledge the agents can reach and how it is structured — is usually an afterthought
("just throw the docs in a vector store"). That asymmetry is where agent systems quietly
fail. A planner can be flawless and still answer wrong if the knowledge layer only exposes
abstracts, surface mentions, and a flat "A references B" edge.

### Why it matters

When the knowledge layer is thin, the agent is forced to *re-derive structure at query
time*: it re-reads raw documents on every request, re-extracts the same entities, and
cannot point at the exact span that justifies a claim. This is slow, non-reproducible, and
unauditable. The same query run twice can extract different facts.

The fix is to move structure extraction **offline**, once, into a durable representation,
and keep the **online** path a thin retrieval-and-reason loop over that representation.
Two jobs that are usually conflated get cleanly split:

- **Offline construction** — parse sources, extract typed entities/claims/relations,
  build a graph or index, assign stable IDs, attach provenance. Done once per document,
  amortized across all future queries.
- **Online use** — retrieve evidence, follow provenance, join across sources, compose an
  answer, cite back to stable IDs. No re-extraction.

### Design checklist

- Can your agent trace any answer back to a **stable identifier + exact span**? If not,
  you have an agent-orchestration system bolted onto an unaudited knowledge blob.
- Is extraction repeated per query? Move it offline.
- Does your "knowledge" preserve *claims, evidence, mechanisms, and relations*, or just
  keywords and chunk text? Thin knowledge caps the ceiling of any planner above it.

The rest of this document is a toolbox of patterns for building the offline
representation and the online interface so that the knowledge layer is as engineered as
the agent layer.

---

## 2. Stable identifiers as join keys

**Pattern: give every entity a globally unique, durable ID at creation time. Use that ID
as the only join key. Store names, aliases, and surface forms as *attributes*, never as
keys.**

When you fuse evidence from multiple sources/views (web result + graph node + a second
extraction pass), you have to decide what makes two records "the same entity." There are
two choices, and only one scales:

1. **Surface-form matching** (join on the name string, or fuzzy string similarity). Cheap
   to start, catastrophic later.
2. **Stable-ID matching** (join on an opaque identifier assigned once). More work upfront,
   correct and fast forever after.

### Why stable IDs win

**Cost.** If every view/source keys records by the same stable ID, a cross-source join is
a **hash join in O(number of keys)**. If sources have no shared IDs, you fall back to
pairwise surface-form comparison, which is worst-case `O(|A| × |B|)` — quadratic in the
two candidate sets — because without a blocking key there is nothing to hash on.

**Correctness.** Surface forms collapse *homonyms*. "Apple" the company and "apple" the
fruit; "BERT" the model and "Bert" a person; two distinct datasets that happen to share a
name. A matcher that merges identical strings will **silently fuse distinct entities**. The
risk of a false merge scales with the fraction of your entities that share a surface form
with another distinct entity (the homonym set), which in real corpora is not small. False
merges are insidious because they corrupt downstream reasoning invisibly: the graph looks
fine, the answer is wrong. The dual failure also bites — the same entity spelled two ways
*splits* into two nodes — and is just as silent.

### How to implement

- **Mint the ID at extraction time**, before any merging. A UUID, a content hash, or a
  resolver-assigned canonical ID all work. The key property is that it is opaque and never
  derived from the mutable surface form.
- **Entity linking / resolution is a deliberate, evidence-backed step**, not a
  string-equality shortcut. Use embedding similarity + controlled vocabularies + (for
  high-stakes merges) human-in-the-loop, and record the decision. Merging two IDs is a
  logged, reversible event. This is the hard problem worth spending effort on, because
  everything downstream trusts the ID.
- **Every view of the same document shares the same node IDs.** A "view" is a deterministic
  projection that filters/relabels edges but *never renames nodes*. This is what makes
  pivoting from one view to another (e.g., "this entity showed up in lexical retrieval —
  now show me every n-ary relation it participates in") a pure ID lookup instead of a
  fuzzy re-match.

### The payoff: identifier-preserving joins enable coverage guarantees

Because every view keeps the same vertex set, the **union of views** never reaches *fewer*
nodes than any single view, and reaches *strictly more* whenever the gold reasoning path
runs through a relation that a single view hides (see §5 and §7). That monotonicity is the
formal backbone behind "fuse more sources, never lose recall." It only holds if joins are
ID-respecting — i.e., keyed by stable IDs and monotone in the edge set. Build on stable
IDs and you get this for free; build on surface forms and you get false merges plus
quadratic cost.

---

## 3. Semantic anchors for multimodal and cross-source linking

**Pattern: don't align modalities (or sources) directly. Insert an intermediate
*semantic anchor* layer — one modality-agnostic summary node per content unit — and route
all cross-modal/cross-source links *through* the anchors.**

The hard problem in multimodal knowledge is the *semantic gap*: a figure, a table, an
equation, and a paragraph all carry related evidence but live in incompatible
representations. Two naive approaches both fail:

- **Flatten everything to text** (e.g., reduce a figure to its caption). You lose the
  structure and most of the evidence.
- **Keep modality-specific graphs and align them post-hoc.** Cross-modal entity alignment
  ("is this table cell the same thing as that figure region?") is brittle and expensive —
  it is the same `N²` pairwise-matching trap as surface-form joins.

### The anchor layer

Decompose each source into **content units** `(modality_type, raw_content, metadata)` where
modality is one of `{text, figure, table, equation, code, ...}` and metadata carries
section hierarchy, page/location, caption associations. Then build a three-layer
heterogeneous graph:

```
  ┌─────────────────────────────────────────────┐
  │  STRUCTURE layer    (sections, documents)     │  ← global context, hierarchical retrieval
  └───────────────▲───────────────────────────────┘
                  │ belongs_to
  ┌───────────────┴───────────────────────────────┐
  │  ANCHOR layer       (one per content unit)      │  ← modality-agnostic semantic summary
  │  a_j = summarize(content_unit_j, local_context) │     + salient entities/relations
  └───────────────▲───────────────────────────────┘
                  │ grounded_in
  ┌───────────────┴───────────────────────────────┐
  │  ENTITY layer       (fine-grained nodes from    │  ← named entities, table entries,
  │  text/figures/tables/equations)                 │     symbolic variables, code symbols
  └─────────────────────────────────────────────────┘
```

- Each **anchor** is generated by an LLM/MLLM from the content unit plus its local context.
  It holds a short modality-agnostic summary plus the salient entities and relations it
  contains. It is a *stable bridge*: a figure-region anchor and a paragraph anchor can be
  linked because both are summaries in the same semantic space, even though their raw
  payloads are not comparable.
- **Entities link to anchors** via `grounded_in`; **anchors link to structure** via
  `belongs_to`. Anchor-to-anchor edges are induced from explicit references, shared
  canonical entities, and embedding similarity.
- **Provenance is preserved at the leaf**: each anchor keeps `(doc_id, page, bbox)` for
  figure/table content, `(file_path, symbol, line_span)` for code. So you get the
  robustness of routing through summaries *and* the precision of pointing at the exact
  pixel region or code line.

A cross-modal question ("what does the diagram say about the entity this paragraph
names?") becomes a traversal *through the shared anchor*, not a fragile heuristic match
between a paragraph and an image caption.

### Why this generalizes beyond multimodal

The anchor layer is really a **decoupling layer**. Anywhere you would otherwise do direct
N-way alignment — multiple OCR engines, multiple extraction passes, text + structured
records, several heterogeneous APIs — insert anchors so that components connect through a
shared abstraction instead of `N²` pairwise alignments. It keeps the graph lightweight
(you don't materialize every cross-modal edge) and extensible (a new modality just needs a
parser and an anchor generator).

### Caveats

- The anchor summary is an LLM artifact; cache it and treat it as derived data tied to its
  source content hash, so it can be regenerated when the model or source changes.
- Don't over-summarize: an anchor that loses the salient entities defeats the purpose.
  Keep entity references in the anchor so retrieval can drill from anchor → entity →
  provenance.

---

## 4. Typed knowledge, durable-vs-ephemeral facts, citation intent

**Pattern: extract *typed* knowledge with explicit relation types, and deliberately admit
only *durable* facts into the relation graph. Push ephemeral, replication-varying values
into entity attributes, not edges. For reference/citation edges, encode *intent and
strength*, not just existence.**

A flat triple store of "everything mentioned" rots fast. Three discipline moves keep a
knowledge graph a *reasoning surface* rather than a noisy catalog.

### 4a. Organize entities by abstraction level

A useful schema separates kinds of knowledge by how they are extracted and how stable they
are. *(The schema below is drawn from a research-paper knowledge-graph illustration; the
abstraction-level discipline generalizes to any document domain — contracts, incident
reports, product docs, support tickets. Read the level names as roles, not a fixed
ontology.)*

- **Meta / factual entities** — low-variance, verifiable metadata (ids, dates, authors,
  affiliations, URLs pinned to commits/hashes). This is the *deduplication and provenance
  backbone*. Verify it strictly; zero tolerance for hallucination.
- **Explicitly mentioned entities** — objects named verbatim in the source (methods,
  datasets, metrics, components). Normalize through controlled vocabularies and embedding
  linking. Synonyms/abbreviations are valid matches; tolerate format, be strict on facts.
- **Implicit / abstracted entities** — what the source *claims, assumes, finds*:
  motivations, contributions, hypotheses, findings, mechanisms, limitations. These are
  synthesized across sections via rhetorical-role tagging, judged on *semantic essence*
  not exact wording, and merged across documents when equivalent.
- **Relations between entities** — fine-grained typed triples connecting the above.

### 4b. Typed relations split into controlled vs open

- **Controlled relations** require both head and tail to already exist as canonical
  entities (e.g., `BUILDS_ON`, `USES_COMPONENT`, `ALTERNATIVE_TO`, `SOLVES`, `APPLIED_TO`).
  They preserve ontological discipline over your core entity space.
- **Open relations** admit new concepts when a mechanism lacks a canonical name, grouped
  into semantic zones: causal (`CAUSES`, `ENABLES`, `INHIBITS`, `CORRELATED_WITH`),
  composition (`CONSISTS_OF`, `IMPLEMENTS`, `COMBINES`, `REQUIRES`), comparison
  (`DERIVED_FROM`, `DIFFERS_FROM`, `HAS_LIMITATION`), structure (`SUBSET_OF`,
  `HAS_PROPERTY`).

Typing edges lets traversal be *selective* ("follow only `BUILDS_ON` edges to recover a
lineage chain") and lets validation reject nonsensical links.

Triples arrive through **two reinforcing channels**, and you should tag which:
- **Structural edges** are materialized *deterministically* from metadata (e.g.,
  `AFFILIATED_WITH` from author records, `USES_COMPONENT` from declared submodules). No LLM
  needed; cheap and exact.
- **Semantic edges** are *mined* with section awareness (harvest causal claims from
  intro/discussion, composition from methods, gaps from related-work, failures from
  limitations) and reconciled against canonical entities to prevent fragmentation.

Record each triple as `(head, head_type, relation, tail, tail_type)` with a **verbatim
evidence span**, a **calibrated confidence**, and a **source tag** distinguishing semantic
from structural provenance.

### 4c. The durable-vs-ephemeral rule (the most important one)

**Admit only durable knowledge into the relation graph.** Numerical results,
hyperparameters, learning rates, seeds, and per-run experimental values are *ephemeral*:
they vary across replications and would dilute the graph with noise that goes stale. They
belong as **attributes on entity nodes**, not as edges in the reasoning graph.

Litmus test: *would this fact still be true in a faithful re-run or a sibling
deployment?* If yes (a method builds on another method; a system targets a problem), it is
durable — make it an edge. If no (this run hit 87.3% accuracy with lr=3e-4), it is
ephemeral — make it an attribute or a separate metrics record.

This rule keeps the graph small, mergeable across documents, and stable over time. It is
also what makes cross-document merging safe: durable relations are the same across
sources; ephemeral values are not.

The same separation applies to **agent working state**. A worker's intermediate
hypothesis or a session's scratch set is ephemeral — it lives in a session/working layer
and is discarded or *promoted deliberately*, never silently merged into the durable store.
Mixing transient agent state into the authoritative graph destroys your ability to know
what is ground truth. (This is the knowledge-substrate face of the working-vs-durable
distinction in agent-memory.)

### 4d. Citation/reference intent levels

A reference edge that only says "A cites B" is nearly useless for reasoning. Encode the
**argumentative intent and strength** of every reference edge:

- A **relation role**: does the citing work *support*, *contrast*, *extend*, or merely use
  as *background*?
- A **strength score** computed from frequency, rhetorical-zone coverage, lexical cues
  ("we adopt", "we improve upon", "unlike"), and proximity to key entities — thresholded
  into levels (e.g., peripheral → contextual → moderate → strong → foundational).
- **Evidence**: the section/paragraph indices and in-text spans.

With intent-typed edges you can trace *method lineage* ("what does this extend, and what
extends it?"), find *pivotal works*, and map how claims propagate — none of which a flat
citation graph supports. The same idea generalizes to any link between records that has an
argumentative or dependency character (supports/refutes in a debate corpus, depends-on in a
codebase, supersedes in a document set).

---

## 5. Multi-source intent-routed retrieval with a coverage guarantee

**Pattern: expose several retrieval sources behind one typed interface, classify query
intent with a single cheap LLM call, re-weight the source mix per intent, and fuse by
stable ID so you inherit a candidate-coverage guarantee.**

No single retrieval source is sufficient:

- **Web / external search** — broad coverage and *recency*, but noisy and document-level.
  The only source that can surface things newer than your last index build.
- **Vector / semantic retrieval** — good semantic recall, no structural reasoning.
- **Graph traversal** — precise and structure-aware, but blind to anything outside the
  graph snapshot.

### The fusion mechanism

Run the relevant sources, normalize each source's scores to `[0,1]`, and combine:

```
score(e) = λ_web · s_web(e) + λ_graph · s_graph(e) + λ_kn · s_kn(e)
fused(q)  = TopK over (web ∪ graph ∪ knowledge-network), by score(e)
```

Each fused entry **retains its source label, evidence type, reference ID, and
provenance**, so the consumer can later show exactly which figure / paragraph / citation
path / code symbol supported each output. Fusion deduplicates on the stable ID (§2), so
the same fact arriving from two sources collapses to one ranked item rather than
double-counting.

### Intent routing

A **lightweight intent classifier** (a few-shot prompt over a frozen LLM, *one call per
query*) maps the query to a category and re-balances `(λ_web, λ_graph, λ_kn)`:

| Intent        | Bias toward         | Rationale                                  |
|---------------|---------------------|--------------------------------------------|
| `recency`     | web                 | newest work won't be in the graph yet      |
| `multimodal`  | graph anchors       | figures/tables/equations live in the graph |
| `lineage`     | knowledge network   | needs typed cross-document traversal       |
| `comparative` | knowledge network   | needs structured relation queries          |
| `general`     | balanced default    | no strong prior                            |

The routing cost is bounded by a single LLM call regardless of how many fan-out retrievals
follow — cheap insurance against spending graph-traversal budget on a recency question.

### The coverage guarantee (and its precondition)

Provided that every linked entry is keyed by a **stable node ID** (§2), the fused candidate
set is *at least as good in gold-answer coverage as the best single linked source*, before
final ranking and truncation. (Stable-ID keying is framed here as a *sufficient* condition
for the monotonic-coverage property — it is what makes the union-of-views argument go
through; it is not claimed to be the only way to get coverage.) It is **strictly better**
whenever joining the sources surfaces gold candidates that no single source exposed
(formally, whenever the gold path uses a relation a single view hides — see §3 and §7).
This is the practical statement of "more identifier-preserving sources never hurt recall,
and often help."

Two operational consequences:

- **Linked vs unlinked evidence are different tiers.** Records that resolve to a graph node
  enjoy the deterministic-join + coverage guarantee. Web records that can't yet be linked
  are kept as document-level evidence but stay *outside* the ID-based guarantee until entity
  linking promotes them. Track which tier each piece of evidence is in.
- **Fuse, don't pick.** The guarantee is the reason to combine sources rather than route to
  exactly one. Routing changes the *weights*; it should rarely zero out a source entirely.

---

## 6. Deterministic typed operators below an LLM composition layer

**Pattern: expose graph/knowledge computation as a small set of *deterministic, typed*
operators (typed input → typed output + provenance). Keep open-ended reasoning, synonym
handling, and multi-step planning in the LLM layer *above*. Let capabilities emerge from
*composition* of the two layers, not from a monolithic prompt.**

The anti-pattern is dumping the graph (or a giant text blob) into the prompt and asking the
LLM to "reason over it." That is unverifiable, non-reproducible, and expensive. Instead,
draw a hard line:

```
   ┌──────────────────────────────────────────────────────────┐
   │  LLM COMPOSITION LAYER                                      │
   │  synonym expansion · query planning · multi-hop chaining   │
   │  open-ended judgement (e.g., novelty rubric)               │
   └───────────▲────────────────────────────────────▲──────────┘
               │ typed calls                          │ typed results + provenance
   ┌───────────┴──────────────────────────────────────┴──────────┐
   │  DETERMINISTIC TYPED OPERATORS                                │
   │  each: typed input → typed output, explicit provenance,       │
   │  reproducible, no hidden LLM call                             │
   └───────────────────────────────────────────────────────────────┘
```

The operators — graph traversal, ID joins, filtering, ranking, aggregation, validation —
are ordinary code with predictable, reproducible behavior, and they are the part you want
*decoupled from the model family*. The LLM's job is to pick *which* operators and *in what
order*, not to perform the traversal or scoring itself.

### Designing the operator set

Keep operators **small, single-purpose, and deterministic**. A representative set for a
knowledge-graph agent (generalize the names to your domain):

- **Seed resolution** — map a mention string to a canonical node set via case-insensitive
  matching with degree-based disambiguation. *Synonym expansion is NOT in here* — the agent
  issues multiple resolution calls; the operator stays deterministic.
- **Lineage reconstruction** — forward/backward traversal along typed edges + shortest path
  weighted by lineage relations; returns an ordered `⟨node, relation, node, ...⟩` chain.
- **Comparative retrieval** — "find all X evaluated on Y under metric Z" as a *single graph
  query* instead of per-document LLM inspection.
- **Anchor retrieval** — return relevant anchors with their raw payloads + bbox/line-span
  provenance (§3).
- **Gap detection** — surface structural signals: orphan nodes (proposed but unused),
  singletons, disconnected components, sparse cells in a projection. Deterministic
  structural queries, no judgement.
- **Grounding/judging** — the *one* place an LLM judge is invoked, on a structured rubric,
  with the retrieved related set as input. Keep the retrieval deterministic and the
  judgement explicit and rubric-bound.

### Why split this way

- **Each layer stays simple.** Graph ops are efficient and verifiable; the agent absorbs
  the messy parts (synonyms, planning, open reasoning).
- **Capabilities emerge from composition**, by design, not accident:
  - deterministic seed resolution **+** agent synonym expansion = semantic retrieval
  - single-hop primitives **+** agent chaining with intermediate filtering = multi-hop
    workflows
  - structural gap detection **+** an LLM rubric = methodological novelty judgement
- **Provenance is structural, not narrated.** Because operators emit typed outputs with
  explicit provenance, you can audit a result without trusting the LLM's prose about what
  it did.

### Exposing the operators

Expose the same primitives through multiple interfaces so different agents can use them:
a **programmatic API**, a **CLI**, and an **MCP server** that registers each primitive as a
typed tool. Provider-agnostic typed tool calls mean the operator layer is reusable across
agent frameworks. The CLI form doubles as a human-debuggable surface and a reproducible
script unit. (For how to design those tool descriptions and schemas so agents call them
correctly, see reliability-and-operations.)

---

## 7. Cost-factored core-then-views multi-view extraction

**Pattern: when one document must be turned into several graph shapes (binary triples,
n-ary hyperedges, temporal edges, event frames, domain-specific schemas), do NOT run N
independent extraction pipelines. Extract a shared canonical *core* once, then derive each
view from it — by free deterministic projection where possible, by one cheap upgrade pass
where not.**

Different downstream tasks want different graph granularities over the *same* document:

- lexical retrieval / path queries → compact **binary triples** `(h, r, t)`
- multi-hop QA with bundled arguments → **n-ary hyperedges** (`k ≥ 3`)
- timeline / change tracking → **temporal edges** with time qualifiers
- biography / compliance → **person-centric** subgraphs
- narrative analysis → **event** nodes with roles
- regulated verticals → **DIY schema** declared in config (see §8)

Running an independent extractor per view multiplies LLM cost by the number of views and
defeats the whole point on a realistic corpus.

### The core-then-views factoring

```
            {chunks}
               │
        ┌──────▼───────┐   2 LLM passes / chunk:
        │  CORE STAGE  │   (1) typed entities, merged across chunks by canonical name
        │ f_core(...)  │   (2) strictly-binary relations w/ evidence, reject any
        └──────┬───────┘       triple whose head/tail is outside the entity set
               │  (V_doc, E_skeleton)  ← canonical IDs assigned ONCE here
       ┌───────┼───────────────────────┐
       ▼       ▼                       ▼
  PROJECTION  PROJECTION           UPGRADE modes
  modes       modes                (nary, temporal, event, diy)
  (binary,    (person, ...)        1 LLM pass / chunk, takes the
   ...)                            skeleton as a structural anchor
  0 LLM calls 0 LLM calls          and lifts it into the target form
```

- **Core stage** (2 LLM passes/chunk): extract typed entities and merge them across chunks
  by canonical name; then, *conditioned on the merged entity set*, extract strictly binary
  relations with evidence quotes, **rejecting any triple whose head or tail isn't in the
  entity set**. Assign stable IDs here, once.
- **Projection modes** derive output *deterministically* by filtering/relabeling the
  skeleton — **zero additional LLM calls**.
- **Upgrade modes** issue *one* LLM pass per chunk that takes the skeleton as a structural
  anchor and lifts it into the target shape (merge co-occurring binary edges into
  hyperedges, attach `{point_time, start_time, end_time, before, after}` qualifiers, etc.).

### The two wins

**Cost.** With `c_core` core passes, `n` chunks, `M` activated modes of which `n_up` are
upgrade modes:

```
C_factored = n · (c_core + n_up)      vs.      C_naive = c_core · n · M
```

Because every upgrade pass operates on an *already-canonicalized, strictly smaller
hypothesis space* than an end-to-end extractor would face, prompts are shorter and
hallucinated participants drop. Projection modes are free. (Illustrative numbers from the
source paper: `n=8, n_up=4, M=6, c_core=2` → 48 calls instead of 96, a 50% saving, while
still producing the full six-view output. Treat the *factoring* as the lesson, not the
exact number — your savings depend on how many of your views are projections.)

**Cross-view joins are free.** Every view references the **same node IDs** (§2), so
pivoting from a binary-retrieval hit to the n-ary hyperedges containing the same head
entity is an *identifier lookup*, not a fuzzy match. This is exactly what lets the
retrieval layer (§5) treat different views as interchangeable, ID-joinable sources.

### When to use it

Any time you need multiple representations of the same source for different consumers.
Extraction depth becomes a *per-source cost dial* rather than a schema change: spend the
expensive multi-view upgrades on high-value or high-ambiguity sources and cheap projection
on the rest — all output lands in the same ID-keyed graph. The discipline — *extract
canonical core once, derive views cheaply, share IDs* — applies well beyond graphs (e.g.,
one parse feeding both a search index and a structured DB, or one transcript feeding
summaries at several granularities).

---

## 8. Weight-frozen self-improving skill libraries

**Pattern: improve extraction/agent quality on a new domain by accumulating a
*version-controlled library of natural-language skills* retrieved into the prompt at
inference time — WITHOUT touching model weights. Distill the skills from a tiny seed of
gold examples via a rollout → classify → induce loop.**

Fine-tuning per vertical is slow (needs a GPU run), risks cross-domain regression, and
produces opaque weights. For onboarding a new domain or fixing systematic extraction
misses, a weight-frozen skill library is often the better tool. It is knowledge
orchestration applied to the agent's own competence: past solutions become durable,
retrievable facts, keyed and version-controlled like any other entry in the store.

### The distillation loop (per gold document)

Given 10–20 hand-curated gold documents for a vertical:

1. **Rollout** — sample the current extractor `K` times with non-zero temperature to get
   genuine sampling variance, producing candidate outputs `{R_1..R_K}`.
2. **Classify** each gold item against the rollouts under a normalizing matcher (e.g.,
   relation normalized to `UPPER_SNAKE_CASE`, participants as a case-insensitive set):
   - **stable** — matched in *all* `K` rollouts → already handled, **skip it**.
   - **unstable** — matched in *some but not all* → drives **path induction**: give the LLM
     the gold item's evidence span and ask it to verbalize the *minimal trigger pattern*
     that would have produced the item on every rollout.
   - **miss** — matched in *none* → drives **hindsight reasoning**: give the LLM the gold
     item *and* the extractor's failing output, ask it to diagnose why the item was missed,
     and propose a corrective pattern.
3. **Induce** candidate skills, each **anchored by a concrete evidence quote** (this is
   essential for later retrievability — abstract skills don't retrieve well).
4. **Fold** each candidate into the library via a *deterministic controller* choosing among
   `{add, modify, merge, keep}`, using token-overlap (e.g., Jaccard) against existing skills
   with a tunable threshold to detect overlap.

### Retrieval at inference time

A lightweight retriever scores every skill against the document title + the first ~2000
chars of body, and **prepends the top-k skills (default k=3) to the extraction system
prompt**. Critically: **when the library is empty or disabled, the retriever returns
nothing and the pipeline is byte-identical to the skill-free baseline.** The skill loop is
a *pure addition*, never a required path — so it can never make the base pipeline worse.

### Why weight-frozen matters in production

- **No GPU to cold-start a new vertical** — onboarding cost becomes gold-annotation time,
  not training time.
- **Cross-domain regression is impossible by construction** — adding a skill for vertical A
  cannot degrade vertical B, because weights never change and skills are retrieved per
  document.
- **Every skill is human-auditable and version-controlled** — decisive in regulated domains
  where *prompt provenance is a compliance requirement*, not a nice-to-have. You can diff,
  review, and roll back skills like code.

### Composing with config-declared schemas

For verticals needing entirely new entity/relation types (e.g., `drug`/`enzyme`/
`adverse_effect`, or `party`/`obligation`/`deadline`/`penalty`), declare the ontology in a
**self-contained config/template** (entity types, relation types, qualifiers, a few-shot
block) that the extractor parses at runtime — *no code deploy per vertical*. The template
and the skill library compose cleanly: the template **fixes the target ontology** and seeds
canonical examples; the skill library **accumulates extraction patterns within that
ontology**. Schema evolution is decoupled from code-deployment cycles, and the marginal
cost of a new vertical becomes dominated by annotation time rather than engineering time.

---

## 9. Auditable coordinator-worker-aggregator swarm execution

**Pattern: when a task contains heterogeneous sub-work that fails in different ways,
decompose it into a coordinator that emits a *typed plan*, specialized workers each fed a
*compact evidence bundle*, and an aggregator that writes an *on-disk manifest of
artifacts* — not a conversational answer. In this design the value is auditability and
failure isolation; reasoning-quality gains, where you want them, come from *collaborative*
patterns (debate, voting, generator-critic — see collaborative-and-single-agent-patterns),
which this structure can host but does not provide on its own.**

A complex research task mixes retrieving sources, inspecting figures/tables, recovering
lineage, reading code, judging novelty, and drafting output — each failing differently.
Hiding all of it inside one chat loop makes the final answer impossible to audit. Promote
the structure to first-class objects instead.

### The three roles

```
   user task q, graph G, mode m
            │
        ┌───▼──────────┐
        │ COORDINATOR  │  emits typed plan P = { job_i = (role_i, payload_i,
        │              │                          output_contract_i, deps_i) }
        └───┬──────────┘
            │ each worker gets a COMPACT evidence bundle (not the whole graph)
   ┌────────┼─────────────────────────┐
   ▼        ▼                          ▼
 WORKER   WORKER  ...               WORKER   (specialized: survey, code-doc,
 (typed output checked vs           idea/critique, prototype, ...)
  output_contract_i)
   └────────┼─────────────────────────┘
            ▼
        ┌──────────────┐
        │ AGGREGATOR   │  writes manifest A = (plan, {worker outputs},
        │              │  {evidence used}) → on-disk artifacts, surfaces failures
        └──────────────┘
```

### The three properties this structure actually buys

This particular decomposition — typed plan, isolated specialized workers, manifest
aggregator — is chosen for *operability*, not for emergent intelligence from agents
talking to each other. There is no built-in worker debate, majority voting, or aggregator
arbitration of disagreements here. Reasoning-quality improvements from collaboration are
real and worth pursuing (branch 3 in the SKILL front door), but they are a *separate*
design choice you would layer in (e.g. an idea/critique worker pair), not a free side
effect of fanning work out. What this structure gives you is three concrete, narrow,
valuable properties:

1. **Typed output contracts.** Each job's `output_contract` is a *schema*, not a free-form
   prompt. Worker outputs are validated against it before the aggregator accepts them, so a
   contract failure becomes a **routable signal** rather than a silent textual error buried
   in prose.
2. **Manifest, not chat history.** The run output is a manifest with on-disk artifacts. You
   can **rerun a single worker, replace a single evidence bundle, or audit a single
   artifact** without replaying the whole session. Compare to a chat transcript, where the
   only unit of replay is "the entire conversation."
3. **Computable failure isolation.** The dependency set `deps_i` makes the affected
   sub-tree of a failure *read off the graph*, not reconstructed from a chat trace. Recovery
   cost scales with the size of the dependent sub-graph, not the size of the whole plan.

### Recovery policy

On worker failure, the coordinator either **retries with an enlarged evidence bundle** (for
transient errors — the worker was starved of context) or **replans the affected sub-tree**
(for persistent errors). Failed jobs are recorded **verbatim** in the manifest so
downstream consumers can inspect them — surfacing failures beats hiding them in a tidy
final summary.

### Evidence discipline

- Workers receive a **compact, fused evidence bundle** (from §5), *not* the whole graph.
  This keeps prompts small and makes "what did this worker see?" answerable.
- The aggregator records, per worker, the **evidence IDs used**, so every line of output
  traces to specific stable IDs and provenance (§2, §3). A swarm you cannot audit is just a
  more expensive single agent — and it still has to beat a strong single-agent baseline to
  earn its keep.
- Writing agent activity (accepted ideas, specs, prototypes) *back* into the graph as new
  typed nodes is fine — but **gate the write on a human decision**, not the agent's own
  judgement (recall §4c: agent state is ephemeral until deliberately promoted), so the
  knowledge layer records both source literature and agent activity without the agent
  silently mutating its own ground truth.

### When to reach for this

Use the swarm pattern when (a) sub-tasks are genuinely heterogeneous, (b) you need to audit
*which evidence supported which output*, and (c) partial reruns are valuable. If what you
actually want is a *better answer to one hard question*, reach instead for a collaborative
pattern (debate, generator-critic, voting) — possibly hosted inside this same structure as
specialized worker roles. The win *this* structure provides is operability: typed
contracts, replayable artifacts, and computable blast radius.

---

## Cross-cutting summary

These nine patterns reinforce each other:

- **Stable IDs (§2)** are the precondition for cheap cross-view/cross-source joins, which
  underpins multi-view extraction (§7) and the retrieval coverage guarantee (§5).
- **Semantic anchors (§3)** and **typed durable knowledge (§4)** are what make the offline
  representation rich enough that the online loop (§1) never has to re-extract.
- **Deterministic operators (§6)** keep the auditable parts verifiable and the LLM confined
  to where it adds value, which is what makes the **swarm (§9)** manifests trustworthy.
- **Cost-factoring (§7)** and **weight-frozen skill libraries (§8)** are how all of this
  stays affordable and adaptable as the corpus and domains grow.

The throughline: engineer the knowledge layer as deliberately as the agent layer, key
everything by stable identifiers, keep durable structure separate from ephemeral values,
and make every answer traceable to an exact span.
