# RUNNING NOTES — Tool & Complex-Component Travelling Docs (PLAN + NOTES)

**Created:** 2026-07-04
**Last updated:** 2026-07-04
(Times are per working session on the date shown; add HH:MM in your timezone if
you want finer granularity.)

---

## What we are aiming to achieve

Formalise per-tool / per-complex-component travelling documentation — a `PLAN`
(intent) and a `NOTES` log (history), keyed by `function` — so any agent or human
fixing a tool or running the next improvement-loop cycle loads the tool's intent
and history first. Today this is a manual request each time; the goal is to
determine how it gets written, where it is stored, and how it is retrieved, and to
do so by **reusing** existing machinery rather than building parallel structures.

This log is the chronological record of the design discussion. `PLAN` and
`RUNBOOK` companions hold the forward design and the how-to.

---

## Session log — 2026-07-04

### 2026-07-04 — Scope fixed
Decided to formalise `PLAN` + `NOTES` for **tools and complicated components**
only (not site- or platform-level docs, which already live in the
`direction`/`mission`/`roadmap` specs and 016/016b). Structured, catalogue-
addressable entries. **Deliberate-decisions captured in prose** (PLAN section +
NOTES narrative). **Locks deferred** — the prose states what's decided; add
enforcement later only if a regression class shows prose isn't honoured.
Categories: (scope)

### 2026-07-04 — Retrieval contract established (`docselect.go`)
Read `docselect.go`. A doc is selected for the diagnosis bundle when its
`DocRule` is `Always`, OR a keyword is a case-insensitive substring of the
hypothesis, OR a path-glob is a substring of an in-scope symbol path. `Doc` is a
**file path** forwarded verbatim via bundle `-doc`; the catalogue is a JSON array
loaded by `LoadDocCatalogue`. Two consequences: (1) "catalogue-addressable via
the existing mechanism" implies **files**, not DB rows, unless the loader is
extended; (2) the current consumer is the code-diagnosis loop reasoning over
chassis symbols. Failure mode is benign (false match adds a doc; miss = today's
behaviour). Categories: (retrieval)

### 2026-07-04 — Primary consumer is direct-by-function, not the catalogue
Concluded the main "get up to speed" path is an agent already fixing a known
`function` loading its PLAN+NOTES **directly** — no fuzzy matching. The catalogue
entry is the fallback for the diagnosis loop that hasn't localised to a
`function` yet; `rag_lookup` is the semantic fallback. Categories: (retrieval)

### 2026-07-04 — Library key confirmed (`content_components`)
Checked schema dump. Tools are uniquely keyed by `function`
(`idx_cc_tool_function_unique` WHERE `component_level='tool'`), kebab-case
constrained. Confirms docs should be keyed by `function` (library-level, survives
forks) — matches the 037 convention. `component_versions` already holds component
version history (not reasoning notes). Categories: (schema)

### 2026-07-04 — Write-access facts resolved the storage fork
User confirmed: deployments go to a **single sites repo** (per-domain
directories); the **`agentchassis` code repo is read-only** to the framework; a
**separate docs repo or DB** is acceptable; the diagnosis loop's docs-directory
lookup can be repointed if needed; git is preferred over DB for the plans. This
removes last session's load-bearing unknown ("can we commit to the chassis repo?"
— no). Categories: (storage, constraint)

### 2026-07-04 — Reuse discovery: the tool-doc header system already ships
Read the uploaded `create_tool_component_action.go` and 019 §Tool Doc Header.
Found a substantial existing system, shipped **2026-06-11**:
- Inline **tool-doc header** gate (`HasToolDocHeader`) — hard-fails generation
  without a sentinel block stating `function`/`purpose`/`behaviour`/`inputs`/
  `outputs`. Stripped at deploy (`StripToolDocHeader`), retained in
  `rendered_html`. It is the **audit/parse anchor only**.
- Creation **provenance** columns `source_agent_type` / `source_orchestration_id`
  on `content_components` (mirrors `knowledge_base`'s pair). Provenance is
  already data, not prose.
- Full prose written to `content_components.description` at generation.
- Lifecycle hooks already exist: create / fork / `tool-improver` /
  `update_component_html` / `tool-auditor` / `check_tool_health` sweep
  (`no_doc_header`, `malformed_doc_header`).
So PLAN/NOTES **extend** this system; they don't replace it. Categories: (reuse)

### 2026-07-04 — `knowledge_base` is a RAG index, and the storage principle is settled
Searched project knowledge. `knowledge_base` is the RAG store (`rag_index`
chunk→embed→store, `rag_lookup` vector+trigram, content-hash keyed, collections
like `standards`). 019's "`knowledge_base` `tool_docs` row" is a RAG **collection**
— a derived retrieval copy. The synthesis notes state the governing principle
explicitly: **"flat files stay the source of truth, the DB copy is a derived
retrieval index."** This decides git-vs-DB: **git is truth; `knowledge_base` is
the index.** "Chassis read-only" means the framework can't push commits to the
`agentchassis` source repo (hence the 06-11 rollout as human migrations + a
binary release); it can write the DB and push to repos it has creds for.
Categories: (storage)

### 2026-07-04 — Storage decision
Source of truth = flat files in a **new writable docs repo**
(`<docs-repo>/tools/<function>/PLAN.md` + `NOTES.md`), written via the same GitHub
adapter pattern as the sites repo (no S3 step needed). `rag_index` a copy into
`knowledge_base` for retrieval, plus one `DocRule` per tool in the deterministic
catalogue. **NOTES → DB table** only if per-entry git commits prove too
heavy/conflict-prone once agents append (037's Phase B trigger); PLAN stays git.
PLAN never a bare unversioned DB column. Categories: (storage, decision)

### 2026-07-04 — Write hooks
PLAN drafted at **creation** inside `create_tool_component` (reasoning already in
hand: spec slice, delivery mechanism, deliberate decisions), library-level so
once per `function`, not on forks. NOTES appended at **modification** by
`tool-improver` / `update_component_html` / `component-template-fixer` as their
last step. Widen the same hooks to interactive/stateful non-tool components; do
not create a doc per `content_components` row. One flat Go action, complexity in
Go, no sub-workflow. Categories: (design)

### 2026-07-04 — OPEN THREAD: `knowledge_base` `tool_docs` write may be unimplemented
019 says the generation step writes the substantive doc to a `knowledge_base`
`tool_docs` row, but the uploaded `create_tool_component_action.go` writes
`description` and `source_*` only — **no `knowledge_base` write**. Verify whether
it happens in another step of the tool-generator workflow or is unimplemented,
before building on it. Do not assume either way. Categories: (verify, gap)

---

## Open threads (carry forward)

1. **Docs-repo + adapter write target** — provision the repo and the GitHub
   adapter credentials/path; the sites-repo adapter is the pattern to extend.
2. **`knowledge_base` schema + `tool_docs` write** — confirm columns/collection
   key and whether the write in 019 exists (uploaded action lacks it).
3. **`docselect` doc-root** — confirm how `-doc` resolves a path so docs in a
   separate repo are reachable; repoint the diagnosis loop if needed.
4. **`deploy_tool_to_site`** — confirm forks need no PLAN write (reuse the
   library `function`; per-site incidents are NOTES entries only).
5. **Format lock** — agree the PLAN sections and the NOTES entry header +
   `Categories:` line (storage-independent; both file and future-DB versions must
   be identical) before wiring any write.

---

## Note on a separate, out-of-scope item mentioned this session

The broader aspiration to closely **log/track agent creation and inter-agent
messages (headers + body)** was restated as context. It is a distinct workstream
from these travelling docs (different responsibility, different data), and has
been kept out of the PLAN/RUNBOOK to preserve separation of concerns. Can be
specced separately if wanted.
