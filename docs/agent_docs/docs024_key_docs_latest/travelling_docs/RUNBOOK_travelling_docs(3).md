# RUNBOOK — Travelling Docs (PLAN + NOTES): Tools, Complex Components, Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-04 (rev 4 — criteria fenced-block convention; Tier 0 completeness; Tier 4 runner active; drafted actions referenced)
**Applies to:** tools (`content_components.component_level='tool'`),
interactive/stateful components, and pipelines (keyed by
`site_work_items.pipeline` values — unconstrained text, confirm live set).

---

## What we are aiming to achieve

Operational how-to for the travelling-docs feature: where a subject's `PLAN` and
`NOTES` live, how they are written/versioned, how acceptance criteria are stated
and consumed, and how a fixer loads context at fix time. Assumes
`PLAN_travelling_docs.md` (rev 4). Extends — does not replace — the tool-doc
header (019). Source of truth is Postgres; git is an optional mirror.

> Gates: **[GATE]** run `drafts/verify_before_migration.sql` clean before
> applying `drafts/0NN_doc_plans_and_notes.sql`; registry + workflow wiring
> migrations follow the drafted actions.

---

## 1. Where the docs live

- **Truth (Postgres):** `doc_plans` (current + history) and `doc_notes`
  (append-only), keyed `(subject_type, subject_key)`:
  `('tool', <function>)` byte-for-byte from `content_components.function`;
  `('pipeline', <pipeline>)` from `site_work_items.pipeline` (live values
  **[GATE]**).
- **Retrieval index (derived):** `knowledge_base` collections via `rag_index`;
  `rag_lookup` for discovery only. Never edited directly.
- **Mirror (optional):** rendered markdown to a docs repo; non-authoritative.

## 2. Writing / updating a PLAN

**Tools:** at first creation of a `function` (`create_tool_component`) and on
intent change — not on forks. Sections: Aim · Source spec · Behaviour contract →
**## Acceptance criteria** with a fenced block:

    ## Acceptance criteria
    ```criteria
    { "profiles": ["desktop","mobile"],
      "checks": [
        {"id":"boots","type":"selector_exists","selector":"#tool-root"},
        {"id":"console","type":"no_console_errors"},
        {"id":"asset","type":"asset_loads","path":"/tools/assets/<function>.js"} ] }
    ```

(schema v0 + check types: `PLAN_tool_acceptance_runner.md`). Multi-page tools
add Page set & inter-page contract. · Delivery mechanism + why · Dependencies ·
Deliberate decisions.

**Pipelines:** authored initially (distil 004–008); Aim · Invariants · Branch
rationale · Seams · Deliberate decisions. **Never embed the step map** — derive
from `agent_definitions`.

**How:** the `write_doc_plan` action (drafted — supersede tx; one current row
per subject enforced by the partial unique index). Then a `rag_index` step.
Rollback = restore a prior row. `pinned=true` = human hold. Keep orchestration
ids/dates out of the inline `<script>` header (019) — provenance is columns.

## 3. Appending a NOTES entry

**When:** any fix (fix agent's **last step**); workflow-altering migrations
(pipeline note: number, what, why); acceptance runs (§5); diagnosis runs (§4).
**How:** `append_doc_note` (drafted — one INSERT; body = the uniform entry;
`site_id` for per-site incidents; categories from the taxonomy).

**Taxonomy:** `css-variable-mismatch`, `empty-shell`/`mode-b-template`,
`broken-template-slots`, `content-vs-runtime-mismatch`, `detool-on-rebuild`,
`js-not-extracted`, `js-bundle-stale`, `schema-template-drift`, `diagnosis`,
`unconfirmed-diagnosis`, `migration`, `seam`, `acceptance-run`,
`acceptance-fail`, `truncated-output`, `needs_criteria`.

## 4. Persisting diagnosis output

`persist_diagnosis_note` (drafted) runs **after** `diagnose_emit` (emit stays
read-only). Explicit subject in `input_data` only — **skip, don't guess**.
CONFIRMED → root-cause entry; UNVERIFIABLE → persisted tagged
`unconfirmed-diagnosis` (dead ends stop retries). `source='diagnosis-loop'`.

## 5. Tool acceptance & iteration

1. **Criteria live in the tool's PLAN** (fenced block; extracted by
   `load_doc_context` as `criteria_json`). Per-site parametrisation, when it
   exists, goes in that site's `direction.must_have` — not in the PLAN.
2. **Ladder:** Tier 0 generation-time (`HasToolDocHeader` gate +
   `check_tool_completeness` — marker/balanced-tags/length; flags-but-passes
   deliberately; optional wiring: `complete=false` → `truncated-output` note) →
   Tier 1 structural (`check_tool_health`) → Tier 2 contract-presence (thin
   check from `criteria_json`) **[Phase A]** → Tier 3 acceptance audit
   (`tool-auditor` vs criteria) **[Phase B]** → **Tier 4 headless runner
   (ACTIVE)**: desktop + mobile profiles via the browser-runner adapter —
   `PLAN_tool_acceptance_runner.md`.
3. **Iteration:** deploy → run → failing criterion → `improve_tool` item
   (criterion as `acceptance_test`, bounded by `max_fix_attempts`) → fixer loads
   PLAN+NOTES → fix → note → re-run. *Working* = criteria pass. A tool with no
   criteria gets a `needs_criteria` note, not a pass.
4. **Multi-page prerequisite:** preserve-sections re-render +
   interactivity-aware save guard (pending) before scaling page counts.

## 6. Loading context at fix time

`load_doc_context` (drafted): current PLAN + latest-N NOTES + `criteria_json`,
composed as one prompt-ready `doc_context` block. `has_plan=false` is a normal
state. For the code-diagnosis loop, hand `doc_context` in the way
`runtime_evidence` is handed to `diagnose_assemble_bundle`. `rag_lookup` for
discovery when the key is unknown. Read **Deliberate decisions** before changing
anything.

## 7. Verification

1. One current PLAN per subject (partial unique enforces; check after a
   supersede: exactly one `is_current=true`).
2. Notes append with non-empty categories; diagnosis runs on an explicit
   subject leave a `diagnosis` entry.
3. Roll-up uses GIN: `SELECT subject_key, created_at FROM doc_notes WHERE
   categories ? 'detool-on-rebuild';`
4. `criteria_json` extracts from the PLAN (fence intact) and parses as JSON.
5. Acceptance: failures created `improve_tool` items carrying the criterion;
   passes left an `acceptance-run` note.
6. Inline header untouched: sentinel in `rendered_html`, stripped from shipped
   page + `/tools/assets/<function>.js`; no `no_doc_header` from
   `check_tool_health`.
7. Pipeline topology claims: regenerate from `agent_definitions`.

0 rows is not decisive until the query itself is ruled out (right subject key,
right `is_current`, right collection).

## 8. Gotchas

- Truth is Postgres, not git (commit path: domain-required, whole-file, no
  read, no retry, serialised) — never the home of an append.
- NOTES = one row per entry, never a shared file (RMW/lost-update).
- One current PLAN per subject; per-site incidents = notes with `site_id`.
- Skip, don't guess the diagnosis subject; a mis-filed note poisons history.
- `check_tool_completeness` **deliberately** lets flagged output through for
  review — don't "fix" that behaviour; wire a note instead.
- Don't hand-draw pipeline topology; author invariants/rationale/seams only.
- Don't restate the PLAN in the inline header (short anchor only).
- Criteria describe the tool TYPE; site-specific expectations →
  `direction.must_have`.
- `rag_index` labels `source_type='scrape'` — parameterise if it matters.
- Workflows flat; complexity in Go; `logger.Info`; check schemas before SQL;
  pods `-n ai-persona-system`, Kafka `-n kafka`; deploys via GitHub → Actions →
  Backblaze.
