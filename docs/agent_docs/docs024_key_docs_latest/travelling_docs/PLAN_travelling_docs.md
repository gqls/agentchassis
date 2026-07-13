# PLAN — Tool & Complex-Component Travelling Docs (PLAN + NOTES)

**Created:** 2026-07-04
**Last updated:** 2026-07-04
**Status:** design agreed in principle; not yet built. This document is the spec.

---

## What we are aiming to achieve

Give every tool and every complex/interactive component its own **travelling
documentation** — a `PLAN` (its intent) and a `NOTES` log (its history) — keyed
to the component's `function`, so that whenever an agent or a human fixes a tool
or the next improvement-loop cycle touches it, it **loads the tool's intent and
history first**, the same way we already load site context. The payoff is fixes
that build on prior decisions instead of re-deriving lost context and fighting
earlier, deliberate choices.

This **extends** the tool-doc header system that shipped on 2026-06-11. It does
not replace it. The inline header stays the audit/parse anchor; PLAN and NOTES
are the substantive intent and the missing change-history.

Scope of this plan: tools (`content_components.component_level = 'tool'`) first,
then widen the same hooks to non-tool components that are interactive / stateful
/ have a runtime contract. It does **not** cover site-level or platform-level
docs (those are the `direction`/`mission`/`roadmap` specs and 016/016b) — kept
separate to preserve distinct responsibilities.

---

## Reuse baseline — what already exists (do not rebuild)

| Piece | Where | What it gives us |
|---|---|---|
| Inline tool-doc header | `platform/content/tool_doc_header.go`; enforced in `create_tool_component_action.go` via `HasToolDocHeader`; 019 §Tool Doc Header | `function`/`purpose`/`behaviour`/`inputs`/`outputs` anchor. Stripped at deploy, retained in `rendered_html`. Anti-drift anchor across LLM rewrites. |
| Creation provenance | `content_components.source_agent_type`, `source_orchestration_id` (`NNN_add_component_provenance.sql`), mirrors `knowledge_base`'s pair | Who/what/which orchestration created the component. Provenance is already **data**, not prose. |
| Full prose blurb | `content_components.description`, written at generation | A single-blob description (unstructured). |
| RAG doc index | `knowledge_base` via `rag_index`/`rag_lookup` (content-hash keyed, collections e.g. `standards`, `tool_docs`) | Semantic retrieval of docs. **Derived index, not the source of truth.** |
| Deterministic doc retrieval | `docselect.go` (`SelectDocs`, `LoadDocCatalogue`), `diagnose_doc_catalogue*.json` | Pastes the matching flat-file doc into the diagnosis bundle by keyword / path-glob / `always`. |
| Lifecycle hooks | creation `tool-generator`→`create_tool_component`; fork `deploy_tool_to_site`; modify `tool-improver`/`update_component_html`; audit `tool-auditor`; sweep `check_tool_health` | The exact points a doc should be written or updated already exist as agent steps. |
| Doc storage principle | synthesis notes: "flat files stay the source of truth, the DB copy is a derived retrieval index" | Decides git-vs-DB for us: git is truth, `knowledge_base` is the index. |

**Consequence:** the genuinely new work is narrow — the NOTES history log, a
structured PLAN richer than the header, a writable home for both, and the
retrieval wiring. Everything else is reuse.

---

## The genuine gaps (what is actually new)

1. **NOTES (change history).** Nothing today records symptom → root cause → fix
   → verification, dead ends, or category tags per tool. `tool_health` flags
   issues but discards the reasoning of the fix. This is the highest-value gap
   for "get up to speed at fix time."
2. **Structured PLAN beyond the header.** The header (`purpose`/`behaviour`) plus
   `description` is close to a PLAN but unstructured and missing first-class
   *delivery-mechanism*, *dependencies*, and *deliberate-decisions* fields.
3. **A writable home.** The header/provenance shipped as human-applied
   migrations because the framework is read-only on `agentchassis`. For docs to
   be written **automatically** we need a location the running framework can
   write to (a new docs repo, or DB), not `agentchassis/docs`.
4. **Verify the `knowledge_base` `tool_docs` write.** 019 says the generation
   step writes it; the uploaded `create_tool_component_action.go` does not.
   Confirm whether it is written elsewhere or is unimplemented before we build on
   the assumption.

---

## Design decisions

### 1. Two docs per `function`, library-level
`PLAN_<function>` (intent; changes rarely) and `NOTES_<function>` (history;
append-only-ish). Keyed by `content_components.function` (unique for tools via
`idx_cc_tool_function_unique`), so they travel through forks/regenerations —
which are keyed by `function`, not by a transient `component_id`. A per-site
incident is a NOTES entry that names its `site_id`; the PLAN describes the tool
*type*.

### 2. Storage — flat-file truth in a **writable docs repo**, RAG-indexed
Follow the system's own principle: **flat files are the source of truth; the DB
copy is a derived retrieval index.**

- **Source of truth:** flat markdown files in a **new docs repo the framework
  can write to** — e.g. `<docs-repo>/tools/<function>/PLAN.md` and `NOTES.md`.
  Git gives the versioning, diff and review the PLAN needs (037's core want),
  and a *new* repo sidesteps the `agentchassis` read-only blocker that a
  `chassis/docs` location would hit. Written via the **same GitHub adapter
  pattern already used for the sites repo** (new adapter target + creds; no
  Backblaze/S3 step needed for docs).
- **Retrieval index:** `rag_index` the PLAN (and a NOTES digest) into a
  `knowledge_base` collection (as `standards` docs already do), retrieved by
  `rag_lookup`. Reuse the rag path as-is.
- **NOTES fallback → DB:** if per-entry git commits prove too heavy or
  conflict-prone for agent appends (037's stated concern), move **NOTES only**
  to a DB table (one row per entry: `function`, `site_id?`, `ts`,
  `categories jsonb`, `body`), keeping PLAN in git. This is the documented
  Phase B trigger, not a day-one build. The PLAN **never** lives in a bare,
  unversioned DB text column — versioning is its whole point.

Rationale for a repo over DB-first: the user prefers git; git is auto-writable
once it's a *new* repo (unlike read-only `agentchassis`); and it stays
consistent with how `standards` docs already work (flat file + rag index).

### 3. Write hooks — as a byproduct of steps that already run
- **PLAN — at creation.** `create_tool_component` already has the generator's
  reasoning in hand (it writes `description` + `source_*`). Add a step that
  drafts `PLAN_<function>.md` from that same material (spec slice, chosen
  delivery mechanism, deliberate decisions). Library-level → written once at
  first creation of a `function`, **not** on per-site forks.
- **NOTES — at modification.** `tool-improver` / `update_component_html` /
  `component-template-fixer` append a NOTES entry as their **last step**
  (symptom → root cause → fix → verification + `Categories:` line).
- **Complex non-tool components:** widen the same hooks to components that are
  interactive/stateful. Do **not** create a doc per `content_components` row;
  generic presentational components rely on 003 contracts instead.

Keep the workflow flat: the doc write is one Go action (complexity in the
action), no sub-workflow. Match existing variable names the actions expect.

### 4. Retrieval — three routes, all reused
- **Direct-by-function (primary, fix-time):** an agent already fixing a known
  `function` loads its PLAN + NOTES directly. No fuzzy matching. This is the
  main "get up to speed" path.
- **Deterministic catalogue:** add one `DocRule` per tool to
  `diagnose_doc_catalogue` — `keywords: [<function>, symptom terms]`,
  `path_globs: [<function>, "create_tool_component"]`. `SelectDocs` surfaces the
  PLAN when the diagnosis loop's hypothesis/scope hits the tool. Benign failure
  mode (a miss = today's behaviour).
- **Semantic:** `rag_lookup` over the `knowledge_base` copy for free-text recall.

### 5. Deliberate-decisions — prose now, no locks
Deliberate decisions live in a `PLAN` section ("Deliberate decisions — do not
re-fix") and the NOTES narrative. No locks yet (deferred by decision). What
makes prose protective is that the fixer **loads** it — guaranteed by the
direct-by-function read and the catalogue entry. If a regression class later
shows the prose isn't honoured, that's the signal to add enforcement (a lock or
a `direction` must_have), not before.

---

## Document format

### PLAN_<function>.md (sections)
- **Aim** — what the tool is for, in product terms.
- **Source spec** — the slice of site spec / roadmap it derives from.
- **Behaviour contract** — states/inputs/outputs (superset of the header's
  `behaviour:`; the header stays the short anchor).
- **Delivery mechanism** — Path 1 (inline `<script>` → `/tools/assets/{function}.js`)
  vs Path 2 (library `js_snippet` → `snippets.js`) vs build-time content — and
  **why**.
- **Dependencies** — data feeds, assets, other components, scheduled tasks.
- **Deliberate decisions** — choices a later pass must NOT "fix".

### NOTES_<function>.md (append-only; import-shaped)
Uniform, dated entries so a later migration into a DB table is an import, not a
reformat:
```
## 2026-07-04 — <short title>
Observed: <symptom>
Root cause: <cause>
Fix: <what changed> (<site_id> if per-site)
Verified: <how confirmed>
Categories: detool-on-rebuild, schema-template-drift
```

### Category taxonomy (reuse 037's; extend as needed)
`css-variable-mismatch`, `empty-shell`/`mode-b-template`, `broken-template-slots`,
`content-vs-runtime-mismatch`, `detool-on-rebuild`, `js-not-extracted`,
`js-bundle-stale`, `schema-template-drift`. Tags roll up into 016/016b when a
category recurs across tools.

---

## Phasing

**Phase A (now — reuses everything, minimal new infra)**
1. Fix the format (above) — storage-independent; both file and future-table
   versions must be identical.
2. Provision the writable docs repo + adapter target (dependency below).
3. `create_tool_component` drafts `PLAN_<function>.md` from reasoning in hand.
4. Direct-by-function load in the fix agents + one `DocRule` per tool in the
   catalogue; `rag_index` the PLAN.
5. Verify / wire the `knowledge_base` `tool_docs` write (gap #4).

**Phase B (when agents append NOTES at volume / cross-tool tag queries recur)**
6. NOTES → DB table; agent append = trivial INSERT; render a NOTES digest for
   the bundle; category roll-up becomes a SQL query feeding 016/016b. PLAN stays
   git.

---

## Open questions / dependencies (verify before building)
- **Adapter write target** for the new docs repo (creds, path convention). The
  sites-repo adapter is the pattern to extend.
- **`knowledge_base` schema** — confirm columns, collection key, and whether a
  `tool_docs` write already exists (019 claims it; the uploaded action lacks it).
- **`docselect` doc-root** — confirm how `-doc` resolves a path so docs in a
  separate repo are reachable (user has said repointing the diagnosis loop is
  acceptable).
- **`deploy_tool_to_site`** — confirm forks need no PLAN write (they reuse the
  library `function`; a per-site incident is a NOTES entry only).

---

## Reuse ledger (functions/tables to reuse, not recreate)
`rag_index` / `rag_lookup` / `knowledge_base` (retrieval index) ·
`docselect.SelectDocs` / `LoadDocCatalogue` / `diagnose_doc_catalogue` (deterministic retrieval) ·
`content_components.function` + `source_*` columns (identity + provenance) ·
`StripToolDocHeader` + `ToolDocOpen`/`ToolDocClose` (inline anchor, untouched) ·
`create_tool_component` / `deploy_tool_to_site` / `tool-improver` /
`update_component_html` / `component-template-fixer` (write hooks) ·
`check_tool_health` (sweep that can flag a missing PLAN/NOTES later) ·
sites-repo GitHub adapter (write path pattern) ·
`site_specs` supersede-log pattern (if a versioned DB variant is ever needed).
