# PLAN — Tool & Complex-Component Travelling Docs (PLAN + NOTES)

**Created:** 2026-07-04
**Last updated:** 2026-07-04 (rev 2 — storage decision reversed to DB-as-truth; two-table design grounded)
**Status:** design agreed; not yet built. This document is the spec.

---

## What we are aiming to achieve

Give every tool and every complex/interactive component its own **travelling
documentation** — a `PLAN` (its intent) and a `NOTES` log (its history), keyed to
the component's `function` — so that whenever an agent or a human fixes a tool or
the next improvement-loop cycle touches it, it **loads the tool's intent and
history first**, the way we already load site context. The payoff is fixes that
build on prior decisions instead of re-deriving lost context and fighting earlier,
deliberate choices.

This **extends** the tool-doc header system that shipped 2026-06-11 (019 §Tool Doc
Header). The inline header stays the short audit/parse anchor; PLAN and NOTES are
the substantive intent and the missing change-history.

Scope: tools (`content_components.component_level = 'tool'`) first, then widen the
same hooks to interactive/stateful non-tool components. Not site- or
platform-level docs (those are the `direction`/`mission`/`roadmap` specs and
016/016b) — kept separate to preserve distinct responsibilities.

---

## Reuse baseline — what already exists (do not rebuild)

| Piece | Where | What it gives us |
|---|---|---|
| Inline tool-doc header | `platform/content/tool_doc_header.go`; enforced in `create_tool_component` via `HasToolDocHeader`; 019 | `function`/`purpose`/`behaviour`/`inputs`/`outputs` anchor. Stripped at deploy, retained in `rendered_html`. |
| Creation provenance | `content_components.source_agent_type`, `source_orchestration_id` | Who/what/which orchestration created the component. Provenance is already data. |
| Versioning pattern | `site_specs` + `site_specs_supersede_log_20260422` (columns: `is_current`, `superseded_at`, `source`, `source_agent`, `source_item_id`, `notes`, `pinned`, `created_by`) and the 002 supersede/restore rollback pattern | A proven, in-codebase way to version a document in Postgres without git. |
| RAG index + retrieval | `rag_index` / `rag_lookup` (`rag_actions.go`), `knowledge_base` (content-hash keyed, collections e.g. `standards`) | Generic chunk→embed→store + vector/trigram lookup. Reuse verbatim as the derived index. |
| Provenance columns on KB | `knowledge_base.source_agent_type`, `source_orchestration_id`, `source_type` | The `source_*` convention to mirror on the new tables. |
| Diagnosis bundle assembler | `diagnose_assemble_bundle` + the `diagnose_load_runtime` read action (`maintenance_actions.go` mould) | The pattern for injecting extra context (runtime evidence, schema) into a diagnosis bundle: a thin `params.DB`+`QueryContext` read action + one compose line. |
| Lifecycle hooks | create `create_tool_component`; fork `deploy_tool_to_site`; modify `tool-improver`/`update_component_html`/`component-template-fixer`; sweep `check_tool_health` | The exact points a doc is written or updated already exist as agent steps. |

**Consequence:** the new work is narrow — the NOTES history, a structured PLAN
richer than the header, two small tables, two thin write actions, and reuse of
`rag_index`/`rag_lookup` for retrieval. Everything else is reuse.

---

## The genuine gaps (what is actually new)

1. **NOTES (change history).** Nothing records symptom → root cause → fix →
   verification, dead ends, or category tags per tool. Highest-value gap.
2. **Structured PLAN beyond the header.** Header + `description` is close but
   unstructured and missing first-class *delivery-mechanism*, *dependencies*, and
   *deliberate-decisions* fields.
3. **A framework-writable home.** The header/provenance shipped as human-applied
   migrations because the framework is read-only on `agentchassis`. For docs to be
   written automatically we need a store the running framework controls.
4. **Verify the `knowledge_base` `tool_docs` write** 019 claims — the uploaded
   `create_tool_component` writes `description` + `source_*` but no KB row.
   (Non-blocking: the KB copy is a `rag_index` of the truth regardless.)

---

## Storage decision — DB is the source of truth; git is an optional mirror

**Reversed from rev 1.** Rev 1 recommended flat files in a new writable docs repo
as truth. The uploaded git path and KB schema overturned that:

- The git commit path (`adapter.go` / `github_client.go`) **rejects an empty
  `Domain`** and **force-prefixes every path with `{domain}/`** (these docs are
  library-level, no domain); it has **no file-read** and does **whole-file
  commits** with `updateRef(force:false)` and **no conflict retry**; and every
  commit **serialises through one Kafka git-adapter**. So a NOTES append becomes a
  read-modify-write the client can't do atomically, and the *source of truth would
  live behind an external, rate-limited, retry-less service*.
- `knowledge_base` is **content-addressed** (`UNIQUE (collection, content_hash)`),
  embedding-coupled, and has **no `is_current`/version chain and no `function`
  key** — the shape of a derived index, not a record of truth.

The system's own norm is DB-as-truth (specs, components, work items, snapshots)
with git as the *publish* path. So:

- **Source of truth = two Postgres tables** the framework writes transactionally
  (below), keyed by `function`.
- **Retrieval index = `knowledge_base`** via `rag_index` into a `tool_docs`
  collection, queried by `rag_lookup` — reused verbatim.
- **Git = an OPTIONAL mirror** (render current PLAN/NOTES → markdown → commit to a
  docs repo) for human browsing/diffs. Because git is no longer authoritative, the
  adapter's domain-prefix / whole-file / no-retry / serialization limits are all
  **non-fatal**: a failed mirror commit costs nothing; truth is safe in Postgres.

Why DB over git here: the framework must write these docs; truth belongs in the
store it controls and can guarantee transactionally; the append-heavy NOTES side
is a natural fit for row-per-entry inserts (no RMW, no lost updates); and the PLAN
gets real version history by reusing the proven `site_specs` supersede pattern.

---

## Two-table design (grounded in the supersede-log + rag signatures)

Confirm no existing `tool_doc*` table on the live DB before migrating (the schema
dump shows none). New tables — no existing table has the right key
(`site_specs` is `site_id,aspect`; `knowledge_base` is content-hash). Both keyed
by `function` (library-level; survives forks/regenerations). `body` is `text`
(markdown prose), not `jsonb`, matching `knowledge_base.content`.

### `tool_doc_plan` — the intent (versioned; supersede pattern re-keyed to `function`)
Re-keys the `site_specs_supersede_log` columns from `(site_id, aspect)` to
`function`, dropping `site_id` (PLAN describes the tool *type*).

```
tool_doc_plan
  id             uuid pk default gen_random_uuid()
  function       text not null                      -- library key (matches content_components.function)
  body           text not null                      -- the PLAN markdown
  source         text                               -- 'tool-generator' | 'human' | 'component-template-fixer' | ...
  source_agent   text                               -- mirrors site_specs
  source_item_id uuid                               -- the work item that caused the write (mirrors site_specs)
  notes          text                               -- why this version changed
  is_current     boolean not null default true
  pinned         boolean not null default false     -- human hold (mirrors site_specs)
  created_by     text
  created_at     timestamptz not null default now()
  superseded_at  timestamptz
  updated_at     timestamptz not null default now()
  -- one current PLAN per function:
  UNIQUE (function) WHERE is_current
```
Write = the 002 supersede/restore pattern, in one tx:
`UPDATE tool_doc_plan SET is_current=false, superseded_at=now() WHERE function=$1 AND is_current;`
then `INSERT ... is_current=true`. Rollback = restore a prior row (flip `is_current`).

### `tool_doc_notes` — the history (append-only; not versioned)
```
tool_doc_notes
  id             uuid pk default gen_random_uuid()
  function       text not null                      -- library key
  site_id        uuid                               -- set for a per-site incident; null for a library-level note
  body           text not null                      -- Observed / Root cause / Fix / Verified
  categories     jsonb not null default '[]'::jsonb -- taxonomy tags
  source         text
  source_agent   text
  source_item_id uuid
  created_by     text
  created_at     timestamptz not null default now()
  INDEX btree (function, created_at DESC)           -- direct-by-function latest-N
  INDEX gin (categories jsonb_ops)                  -- roll-up: WHERE categories ? 'detool-on-rebuild'
```
Append = one `INSERT` (no read-modify-write; Postgres serialises concurrent
inserts; no lost updates — the reason NOTES are a table, not a git file).
GIN with the default `jsonb_ops` opclass (not `jsonb_path_ops`) so the `?`
operator is indexable.

Both carry the `source_*`/`source_item_id` provenance already used on
`content_components`, `knowledge_base`, and `site_specs`.

---

## Write hooks (as a byproduct of steps that already run)

- **PLAN — at creation and on later intent edits.** A thin `write_tool_plan`
  action implementing the supersede write above. Called at first creation of a
  `function` (from `create_tool_component`, using the reasoning it already
  holds — spec slice, delivery mechanism, deliberate decisions) and again on any
  direction change. Library-level → not written on `deploy_tool_to_site` forks.
- **NOTES — at modification.** A thin `append_tool_note` action (one `INSERT`),
  called as the **last step** of `tool-improver` / `update_component_html` /
  `component-template-fixer` (symptom → root cause → fix → verification +
  `categories`). A per-site incident sets `site_id`.
- **Index step (reuse).** After either write, a workflow step calling the existing
  `rag_index` with `collection='tool_docs'` and `content_field` pointing at the
  doc body. (Optional one-line change: parameterise `rag_index`'s hardcoded
  `source_type='scrape'`, default unchanged — note it if adopted.)
- **Mirror step (optional).** A git-adapter `commit` request rendering the current
  docs to the docs repo. Non-authoritative.

Keep workflows flat: each write is one Go action; `rag_index` and the git commit
are existing actions used as steps. No sub-workflows — the fix agents already own
these workflows. Do **not** create a doc per `content_components` row; generic
presentational components rely on 003 contracts.

---

## Retrieval (grounded in the actual actions)

- **Direct-by-function (primary, fix-time).** The improvement-loop fixers
  (`tool-improver`, `component-template-fixer`) already know the `function`, so
  they read `tool_doc_plan` (current) + `tool_doc_notes` (latest-N) directly
  before fixing. A thin read action in the `diagnose_load_runtime` /
  `maintenance_actions.go` mould (`params.DB` + `QueryContext`, returns text).
- **Code-diagnosis loop (secondary).** `diagnose_assemble_bundle` does **not** call
  `docselect` today, so tool docs reach it via the same pattern as
  `runtime_evidence`: a sibling read action that resolves the in-scope tool's
  `function` (from the runtime target / affected page) and returns its PLAN+NOTES
  as text, plus one compose line in the assembler. Not a `docselect` catalogue
  entry.
- **Semantic discovery.** `rag_lookup` with `collection='tool_docs'` for
  topic-similar tool-doc chunks when the `function` isn't known. Reused as-is
  (note: `rag_lookup` has no `function` filter — discovery only, not exact load).
- **`docselect` catalogue (only if wired later).** `SelectDocs` matches on
  keyword/path-glob/`always` and reads doc **files**; the diagnosis loop's scope
  is *code symbols*, and a `function` name won't appear in a Go path, so the
  keyword (function + symptom) is the lever — and it needs the git **mirror** to
  supply the file. Deferred; the two routes above don't need it.

---

## Deliberate-decisions — prose now, no locks
A `PLAN` section ("Deliberate decisions — do not re-fix") plus the NOTES
narrative. No locks (deferred). Prose is protective because the fixer **loads** it
(direct-by-function read). Revisit enforcement (a lock / `direction` must_have)
only if a regression class shows the prose isn't honoured.

---

## Document format

**PLAN sections:** Aim · Source spec · Behaviour contract (superset of the
header's `behaviour:`) · Delivery mechanism + why (Path 1 inline `<script>` →
`/tools/assets/{function}.js` vs Path 2 `js_snippet` → `snippets.js` vs build-time
content) · Dependencies · Deliberate decisions.

**NOTES entry (import-shaped, uniform, dated):**
```
## 2026-07-04 — <short title>
Observed: <symptom, where>
Root cause: <cause>
Fix: <what changed> (<site_id> if per-site)
Verified: <how confirmed>
Categories: detool-on-rebuild, schema-template-drift
```
Stored one row per entry in `tool_doc_notes` (`body` = the entry, `categories` =
the tags).

**Category taxonomy (reuse 037's; extend):** `css-variable-mismatch`,
`empty-shell`/`mode-b-template`, `broken-template-slots`,
`content-vs-runtime-mismatch`, `detool-on-rebuild`, `js-not-extracted`,
`js-bundle-stale`, `schema-template-drift`. Roll up into 016/016b via the GIN
query when a category recurs across tools.

---

## Phasing

**Phase A (now — two tables, two write actions, reuse `rag_index`)**
1. Lock the PLAN sections + NOTES entry format (storage-independent).
2. Migration: `tool_doc_plan` + `tool_doc_notes` (confirm absence on live DB first).
3. `write_tool_plan` (supersede) wired into `create_tool_component`;
   `append_tool_note` wired as the last step of the fix agents.
4. `rag_index` step (`collection='tool_docs'`) after each write; direct-by-function
   read action for the fixers.
5. Verify/settle the KB `tool_docs` write (gap #4).

**Phase B (when useful)**
6. Optional git mirror (render → docs repo) for human browsing.
7. Optional `docselect` wiring + assembler doc-injection if the code-diagnosis loop
   should also see tool docs deterministically.
8. Category roll-up query surfaced into 016/016b.

---

## Open questions / dependencies
- **KB `tool_docs` write** — confirm whether it exists (uploaded action lacks it);
  it becomes the `rag_index` step here regardless.
- **`deploy_tool_to_site`** — confirm forks stamp `source_*` and need only a NOTES
  entry (no PLAN write). Minor follow-up.
- **`rag_index` `source_type`** — decide whether to parameterise (default
  `'scrape'`) so tool docs aren't mislabelled. One-line, backward-compatible.
- **Mirror repo** — only if the optional git mirror is adopted (Phase B).

---

## Reuse ledger (functions/tables/patterns to reuse, not recreate)
`site_specs` supersede columns + 002 supersede/restore pattern (PLAN versioning) ·
`rag_index` / `rag_lookup` / `knowledge_base` (`rag_actions.go`) verbatim
(retrieval index) · `source_*`/`source_item_id` provenance convention
(`content_components`, `knowledge_base`, `site_specs`) · `diagnose_load_runtime` /
`maintenance_actions.go` read-action mould (direct-by-function loader + assembler
injection) · `content_components.function` (identity key) · inline tool-doc header
+ `StripToolDocHeader` (untouched) · `create_tool_component` /
`deploy_tool_to_site` / `tool-improver` / `update_component_html` /
`component-template-fixer` (write hooks) · git adapter `commit` action (optional
mirror) · `check_tool_health` (can flag a missing PLAN/NOTES later).
