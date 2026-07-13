# RUNNING NOTES — Tool & Complex-Component Travelling Docs (PLAN + NOTES)

**Created:** 2026-07-04
**Last updated:** 2026-07-04 (rev 2)
(Times are per working session on the date shown; add HH:MM in your timezone for
finer granularity.)

---

## What we are aiming to achieve

Formalise per-tool / per-complex-component travelling documentation — a `PLAN`
(intent) and a `NOTES` log (history), keyed by `function` — so any agent or human
fixing a tool or running the next improvement-loop cycle loads the tool's intent
and history first. Today this is a manual request each time; the goal is to
determine how it gets written, where it is stored, and how it is retrieved, by
**reusing** existing machinery. This log is the chronological record of the design.

---

## Session log — 2026-07-04 (rev 1)

### Scope fixed
`PLAN` + `NOTES` for tools and complicated components only (not site/platform docs,
which live in `direction`/`mission`/`roadmap` specs and 016/016b). Structured,
catalogue-addressable, deliberate-decisions in prose, locks deferred. Categories:
(scope)

### Retrieval contract (`docselect.go`)
A doc is selected when its `DocRule` is `Always`, OR a keyword substrings the
hypothesis, OR a path-glob substrings an in-scope symbol path. `Doc` is a **file
path** forwarded via `-doc`. Implications: "catalogue-addressable via the existing
mechanism" implies files; the consumer is the code-diagnosis loop over chassis
symbols. Primary "get up to speed" path = direct-by-function load by the fixer,
not the catalogue. Categories: (retrieval)

### Library key (`content_components`)
Tools uniquely keyed by `function` (`idx_cc_tool_function_unique`), kebab-case.
Docs keyed by `function` → survive forks. Categories: (schema)

### Write-access facts
Deployments → single sites repo (per-domain dirs); `agentchassis` read-only to the
framework; a separate docs repo or DB acceptable; diagnosis docs-lookup
repointable; git preferred over DB. Categories: (constraint)

### Reuse discovery: tool-doc header system already ships (2026-06-11)
Inline header gate (`HasToolDocHeader`), `source_*` provenance columns, full prose
to `content_components.description`, a `knowledge_base` `tool_docs` collection
(claimed by 019), and lifecycle hooks (create/fork/improve/audit/`tool_health`
sweep). PLAN/NOTES **extend** this, not replace it. Categories: (reuse)

### `knowledge_base` is a RAG index; storage principle stated
`knowledge_base` = RAG store (`rag_index`/`rag_lookup`, content-hash keyed,
collections like `standards`). Synthesis-note principle: "flat files stay the
source of truth, the DB copy is a derived retrieval index" — but that was scoped to
human-authored `standards` reference docs. Categories: (storage)

### Rev-1 storage decision (later reversed)
Recommended flat files in a new writable docs repo as truth + `rag_index` copy +
`docselect` entry; NOTES → DB only on friction. Categories: (storage, superseded)

### OPEN THREAD: KB `tool_docs` write may be unimplemented
019 says generation writes a `knowledge_base` `tool_docs` row; the uploaded
`create_tool_component` writes `description` + `source_*` only. Verify. Categories:
(verify, gap)

---

## Session log — 2026-07-04 (rev 2 — storage reversed to DB-as-truth)

### Git commit path evidence (`adapter.go`, `github_client.go`)
`handleCommitAction` **hard-rejects an empty `Domain`**; `CommitToRepo`
force-prefixes every path with `{domain}/`. No file-read action; whole-file
commits; `updateRef(force:false)` with **no conflict retry**; all commits
**serialise through one Kafka git-adapter**. So a NOTES append via git is a
read-modify-write the client can't do atomically, and using git as truth puts the
record of truth behind an external, retry-less service. `RepoName` is per-message
(a separate repo is reachable) but the domain prefix still applies. Categories:
(storage, constraint)

### `knowledge_base` schema confirmed
`UNIQUE (collection, content_hash)` (content-addressed), `embedding vector(768)`,
`source_*` columns, **no `is_current`/version chain, no `function` key**. Shape of a
derived index, not a record of truth. As truth for an evolving PLAN it is wrong
(editing → new hash → orphan row; versioning absent); as the retrieval index it is
right and already built. An external vector DB as truth is strictly worse (same
problems + infra). Categories: (storage)

### GIN and RMW clarified (user question)
GIN = generalized inverted index; indexes the elements inside a composite value
(e.g. `jsonb` array), so `categories ? 'tag'` is index-backed — enables cross-tool
roll-up at scale. Use `jsonb_ops` (not `jsonb_path_ops`) so `?` is indexable;
`knowledge_base` already uses a GIN trigram index, so GIN is in-codebase. RMW =
read-modify-write; appending to one shared file is RMW with lost-update risk under
the retry-less commit path; a DB `INSERT` (row per entry) avoids RMW — Postgres
serialises concurrent inserts. This is the concrete reason NOTES = table, PLAN
(rare, whole-doc) can tolerate a file. Categories: (storage, decision)

### DECISION: DB is the source of truth; git is an optional mirror
Two Postgres tables the framework writes transactionally are the truth;
`knowledge_base` (`rag_index`/`rag_lookup`) is the derived index; git is an
OPTIONAL non-authoritative mirror (render → docs repo) for human browsing. Because
git isn't authoritative, the adapter's domain-prefix/whole-file/no-retry/
serialization limits are non-fatal. Consistent with the system norm (DB truth, git
publish). Supersedes the rev-1 flat-files recommendation. Categories: (storage,
decision)

### Supersede-log pattern confirmed (`site_specs_supersede_log_20260422`)
Columns: `is_current`, `superseded_at`, `source`, `source_agent`, `source_item_id`,
`notes`, `pinned`, `created_by`. Reuse the *pattern* (not the table — it's keyed by
`site_id,aspect`) re-keyed to `function` for `tool_doc_plan`, so the PLAN gets real
version history in Postgres. Categories: (reuse, schema)

### `rag_index` / `rag_lookup` signatures grounded (`rag_actions.go`)
`rag_index`: generic chunk→embed→`INSERT ON CONFLICT (collection, content_hash) DO
NOTHING`; config `content_field` + `collection`; stamps `source_agent_type`/
`source_orchestration_id`; hardcodes `source_type='scrape'` (optionally
parameterise). → Indexing a PLAN/NOTES digest is a workflow step with
`collection='tool_docs'`, pure reuse. `rag_lookup`: filters by `collection` (+
optional `industry`); **no `function` filter** → discovery only; exact "load docs
for function X" must query the truth table. Categories: (reuse, retrieval)

### `diagnose_assemble_bundle` does NOT call `docselect`
The chassis bundle = hypothesis + in-scope code bodies + live schema +
`runtime_evidence`. Authored-doc injection isn't wired for any doc. So feed tool
docs to the code loop via the `diagnose_load_runtime` pattern (thin `params.DB`
read action resolving the in-scope `function` + one assembler compose line), not a
`docselect` entry. `docselect` route stays deferred and would need the git mirror
to supply files. Categories: (retrieval)

### Two-table design settled and grounded
`tool_doc_plan` (supersede-pattern, `function`-keyed, `body text`, partial unique
on `is_current`) + `tool_doc_notes` (append-only, `function`+`site_id?`,
`categories jsonb` GIN, btree `function,created_at`). Both carry `source_*`/
`source_item_id`. Writes: `write_tool_plan` (supersede) at creation/edit,
`append_tool_note` (INSERT) at modification, `rag_index` step after each. No
existing `tool_doc*` table in the dump — new migration (confirm on live DB).
Categories: (schema, design)

---

## Open threads (carry forward)

1. **Migration** — create `tool_doc_plan` + `tool_doc_notes`; confirm no
   `tool_doc*` on the live DB first (don't take the dump as decisive).
2. **KB `tool_docs` write** — confirm whether it exists (uploaded action lacks it);
   it becomes the `rag_index` step regardless.
3. **`deploy_tool_to_site`** — confirm forks stamp `source_*` and need only a NOTES
   entry (no PLAN write). Minor follow-up read.
4. **`rag_index` `source_type`** — decide whether to parameterise (default
   `'scrape'`).
5. **Format lock** — agree PLAN sections + NOTES entry header before wiring writes.
6. **Optional (Phase B)** — git mirror; `docselect`+assembler doc-injection;
   roll-up query into 016/016b.

---

## Note on a separate, out-of-scope item

The aspiration to closely **log/track agent creation and inter-agent messages
(headers + body)** is a distinct workstream (different responsibility/data), kept
out of these docs to preserve separation of concerns. Can be specced separately.
