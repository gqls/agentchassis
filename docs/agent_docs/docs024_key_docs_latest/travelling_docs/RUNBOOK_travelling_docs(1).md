# RUNBOOK — Tool & Complex-Component Travelling Docs (PLAN + NOTES)

**Created:** 2026-07-04
**Last updated:** 2026-07-04 (rev 2 — DB-as-truth two tables; git mirror optional)
**Applies to:** tools (`content_components.component_level = 'tool'`) and, later,
interactive/stateful non-tool components.

---

## What we are aiming to achieve

Operational how-to for the travelling-docs feature: where a tool's `PLAN` and
`NOTES` live, how they are written and versioned, how they are indexed for
retrieval, and how a fixer loads them at the moment of a fix. Assumes the design in
`PLAN_travelling_docs.md`. Extends — does not replace — the tool-doc header (019).

> Prereqs not yet in place are flagged **[TODO]**. Source of truth is Postgres;
> git is an optional non-authoritative mirror.

---

## 1. Where the docs live

- **Source of truth (Postgres):** `tool_doc_plan` (current + history, keyed by
  `function`) and `tool_doc_notes` (append-only, keyed by `function`, `site_id`
  for per-site incidents). **[TODO: migration; confirm no `tool_doc*` table exists
  on the live DB first — the schema dump shows none, but verify, don't take the
  dump as decisive.]**
- **Retrieval index (derived):** a `knowledge_base` collection `tool_docs`,
  populated by the existing `rag_index`, queried by `rag_lookup`. Never edited
  directly.
- **Human-facing mirror (optional):** a render of the current PLAN/NOTES to
  markdown committed to a docs repo via the git adapter. Non-authoritative — if a
  commit fails, truth is untouched in Postgres.

`function` is the identity — it must match `content_components.function`
byte-for-byte (kebab-case, `tool-` prefix). Docs are keyed by `function` so they
survive forks and regenerations.

---

## 2. Writing / updating a PLAN

**When:** once at first creation of a `function` (inside `tool-generator` →
`create_tool_component`), and again on any intent/direction change. **Not** on
`deploy_tool_to_site` forks (a fork reuses the library `function`).

**How (via the `write_tool_plan` action — supersede pattern, one tx):**
1. `UPDATE tool_doc_plan SET is_current=false, superseded_at=now() WHERE function=$1 AND is_current;`
2. `INSERT INTO tool_doc_plan (function, body, source, source_agent, source_item_id, notes, created_by) VALUES (...);`
   with `is_current` defaulting true.
3. Workflow step: `rag_index` with `collection='tool_docs'`, `content_field` →
   the new PLAN body.
4. (Optional) mirror commit.

Rollback a PLAN = restore a prior row (flip `is_current` back), same as the 002
spec rollback. `pinned=true` marks a human hold.

Keep orchestration ids / agent names / dates **out** of the inline `<script>`
header (019 rule) — that provenance lives in the `source_*` columns and in the
PLAN/NOTES rows.

---

## 3. Appending a NOTES entry

**When:** any change or fix — `tool-improver`, `update_component_html`,
`component-template-fixer`, or a manual fix. The fixing agent's **last step**.

**How (via the `append_tool_note` action — one `INSERT`, no read-modify-write):**
```
INSERT INTO tool_doc_notes (function, site_id, body, categories, source, source_agent, source_item_id, created_by)
VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8);
```
`body` uses the import-shaped entry (Observed / Root cause / Fix / Verified);
`categories` is the tag array; `site_id` is set for a per-site incident. Then
re-`rag_index` the NOTES digest (optional per-append; can batch).

**Category taxonomy:** `css-variable-mismatch`, `empty-shell`/`mode-b-template`,
`broken-template-slots`, `content-vs-runtime-mismatch`, `detool-on-rebuild`,
`js-not-extracted`, `js-bundle-stale`, `schema-template-drift`.

---

## 4. Loading context at fix time (the payoff)

1. **Direct-by-function (primary).** The fixer knows the `function`; read truth
   directly via a thin read action (`diagnose_load_runtime` / `maintenance_actions`
   mould, `params.DB`+`QueryContext`):
   ```
   SELECT body FROM tool_doc_plan  WHERE function=$1 AND is_current;
   SELECT body, categories, created_at FROM tool_doc_notes WHERE function=$1 ORDER BY created_at DESC LIMIT $2;
   ```
   Return as text for the agent's prompt. Read the PLAN's **Deliberate decisions**
   before changing anything.
2. **Code-diagnosis loop.** `diagnose_assemble_bundle` doesn't call `docselect`
   today, so inject tool docs the same way `runtime_evidence` is injected: a
   sibling read action resolves the in-scope tool's `function` (from the runtime
   target / affected page) and returns its PLAN+NOTES text; add one compose line to
   the assembler. **[TODO if the code loop should see tool docs.]**
3. **Semantic discovery.** `rag_lookup` `collection='tool_docs'` for topic-similar
   chunks when the `function` isn't known. (No `function` filter — discovery only.)

---

## 5. Verification

1. **PLAN row present + current:** `SELECT function, is_current FROM tool_doc_plan
   WHERE function=$1 AND is_current;` returns one row; its Aim matches
   `content_components.function`.
2. **Supersede works:** after a second `write_tool_plan`, the prior row has
   `is_current=false, superseded_at` set and exactly one `is_current=true` remains
   (the partial unique index enforces this).
3. **NOTES append works:** after a fix, a new `tool_doc_notes` row exists with a
   non-empty `categories`.
4. **Roll-up query works:** `SELECT function, created_at FROM tool_doc_notes WHERE
   categories ? 'detool-on-rebuild' ORDER BY created_at DESC;` uses the GIN index.
5. **RAG returns it:** `rag_lookup` for the tool's purpose surfaces `tool_docs`
   chunks.
6. **Inline header untouched:** `<script>` still opens with the sentinel block;
   `StripToolDocHeader` still removes it from the shipped page and
   `/tools/assets/<function>.js` (grep the committed output — `tool-doc` finds
   nothing); `check_tool_health` raises no `no_doc_header`/`malformed_doc_header`.

Don't treat a 0-row result as decisive until the query itself is ruled out (right
`function`, right `is_current` filter, right collection).

---

## 6. Gotchas

- **`agentchassis` is read-only to the framework**, and the git commit path
  (`adapter.go`/`github_client.go`) rejects an empty `Domain`, force-prefixes
  `{domain}/`, has no file-read, does whole-file commits with no conflict retry,
  and serialises through one adapter. That is why truth is in Postgres and git is
  only an optional mirror — never the source of a NOTES append.
- **NOTES = one row per entry, never a shared file.** Appending to a single
  `NOTES.md` is a read-modify-write with lost-update risk under the retry-less
  commit path; a DB `INSERT` avoids it entirely.
- **Library vs site.** PLAN describes the tool *type* (one current row per
  `function`). A per-site incident is a `tool_doc_notes` row with `site_id` set —
  not a new PLAN.
- **Don't duplicate the inline header's job.** The header is the short audit/parse
  anchor (6–12 lines); the PLAN is the substantive intent. Keep them consistent.
- **`rag_index` labels `source_type='scrape'`** by default — parameterise (default
  unchanged) if a correct label matters for tool docs; note the change if made.
- **Keep workflows simple.** Each write is one Go action; `rag_index` and the git
  commit are existing actions used as steps. No sub-workflows — spawn/append within
  the fix agents' own workflows. Use `logger.Info` (not `logger.Debug`).
- **Kubernetes / build:** actions run in the chassis binary; changes are drafted
  for review and shipped via GitHub Actions → Backblaze, not built in-container.
  Pods: `kubectl -n ai-persona-system get pods`; Kafka: `kubectl -n kafka get pods`.
