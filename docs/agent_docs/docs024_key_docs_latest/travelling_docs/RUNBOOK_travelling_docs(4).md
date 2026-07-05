# RUNBOOK — Travelling Docs (PLAN + NOTES): Tools, Complex Components, Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-04 (rev 5 — rollout tracker added; Tier-2 and Runner P0 defined; live pipeline values recorded; migration APPLIED)
**Applies to:** tools (`content_components.component_level='tool'`),
interactive/stateful components, and pipelines. Pipeline subject keys (live,
confirmed 2026-07-04): `build`, `content`, `design`, `maintenance`.

---

## 0. ROLLOUT TRACKER — YOU ARE HERE

**Position as of 2026-07-04:**
Stage 0 ✅ · Stage 1 ✅ · **Stage 2 ⏳ IN PROGRESS (Go actions deploying)** ·
Stage 3 ▶ next · Stages 4–6 pending.

---

### Stage 0 ✅ Pre-migration gates — DONE 2026-07-04
`drafts/verify_before_migration.sql` results recorded: no `doc_plans`/`doc_notes`
collision; `site_work_items.pipeline` live values = `build` (3579), `content`
(24), `design` (13), `maintenance` (2); no CHECK constraint. These four values
are the valid pipeline `subject_key`s.

### Stage 1 ✅ Migration — APPLIED 2026-07-04
Statement tally verified from psql output: 2 CREATE TABLE, 5 CREATE INDEX,
3 COMMENT, COMMIT — `doc_plans` + `doc_notes` and all indexes are live.

**Optional one-time table smoke (run any time; includes cleanup).** Note the
second INSERT is EXPECTED TO ERROR — that is the partial unique index
`idx_doc_plans_current` proving one-current-per-subject (the `write_doc_plan`
action avoids this by superseding first, in one tx):

```sql
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool','smoke-test-tool','smoke v1','human','smoke');

-- EXPECTED ERROR: duplicate key violates "idx_doc_plans_current"
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool','smoke-test-tool','smoke v2','human','smoke');

-- the supersede the action performs:
UPDATE doc_plans SET is_current=false, superseded_at=now()
 WHERE subject_type='tool' AND subject_key='smoke-test-tool' AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool','smoke-test-tool','smoke v2','human','smoke');

SELECT subject_key, is_current, superseded_at IS NOT NULL AS superseded, created_at
FROM doc_plans WHERE subject_key='smoke-test-tool' ORDER BY created_at;

-- notes + GIN operator smoke
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('tool','smoke-test-tool','## 2026-07-04 — smoke','["diagnosis"]'::jsonb,'human','smoke');
SELECT subject_key, categories FROM doc_notes WHERE categories ? 'diagnosis';

-- cleanup
DELETE FROM doc_notes WHERE subject_key='smoke-test-tool';
DELETE FROM doc_plans WHERE subject_key='smoke-test-tool';
```

### Stage 2 ⏳ Go actions build + deploy — IN PROGRESS
The four drafts are being deployed. What must be true before Stage 3:

1. **registry.go carries the four entries** (consolidated from the file headers):
```go
"write_doc_plan":         {Handler: WriteDocPlanAction,        Category: "documentation", Description: "Supersede-write the current PLAN doc for a tool/pipeline subject", IsLocal: true},
"append_doc_note":        {Handler: AppendDocNoteAction,       Category: "documentation", Description: "Append one NOTES entry (row) for a tool/pipeline subject",        IsLocal: true},
"load_doc_context":       {Handler: LoadDocContextAction,      Category: "documentation", Description: "Load current PLAN + latest NOTES for a tool/pipeline subject",    IsLocal: true},
"persist_diagnosis_note": {Handler: PersistDiagnosisNoteAction, Category: "documentation", Description: "Persist a diagnosis report as a NOTES entry when the subject is explicit", IsLocal: true},
```
2. Ship via the normal path (GitHub → Actions → Backblaze; pods on the new binary):
```bash
kubectl -n ai-persona-system get pods
kubectl -n ai-persona-system rollout status deployment/<chassis-deployment-name>
# optional, if action registration is logged at startup:
kubectl -n ai-persona-system logs <pod> | grep -iE "write_doc_plan|append_doc_note|load_doc_context|persist_diagnosis_note"
```
**Done when:** pods are on the new image. (If startup doesn't log registrations,
Stage 3's first run is the confirmation.)

### Stage 3 ▶ NEXT — first wiring: `persist_diagnosis_note` (the smoke consumer)
Chosen first because it is additive, config-gated, sits at the END of an
existing workflow, and you can trigger it on demand with a diagnosis run.

1. **Fetch the current definitions first** (rule: read before writing jsonb —
   the 072 lesson: steps hide in nested sub_workflows):
```sql
SELECT type, processing_mode
FROM agent_definitions
WHERE type ILIKE '%diagnos%'
   OR type IN ('tool-generator','tool-improver','component-template-fixer');

SELECT jsonb_pretty(default_config->'workflow')
FROM agent_definitions
WHERE type = '<the-diagnosis-agent-type-from-above>';
```
2. **Paste the workflow JSON back** → the UPDATE agent_definitions migration is
   drafted against the real nesting (inserting `persist_diagnosis_note` after
   the emit step, with `diag_*_field` defaults matching emit's output_field
   `diagnosis`).
3. Apply, then trigger ONE diagnosis run with an explicit subject in input
   (`subject_type`:`tool`|`pipeline`, `subject_key`), e.g. a known tool function
   or `pipeline`/`build`.
4. **Verify:**
```sql
SELECT subject_type, subject_key, categories, left(body, 120) AS body_head, created_at
FROM doc_notes ORDER BY created_at DESC LIMIT 5;
```
Expect one `diagnosis` (or `unconfirmed-diagnosis`) entry. A run WITHOUT an
explicit subject must produce `persisted:false` in the step output and NO row
(skip-don't-guess).

### Stage 4 — remaining wiring (same fetch-first pattern per agent)
- **tool-generator / `create_tool_component` path → PLAN at creation.** Needs a
  PLAN-body composition step before `write_doc_plan` (the action reads
  `doc_plan_body` from collected data): compose from the material in hand (spec
  slice, delivery mechanism, deliberate decisions, and an initial
  ```criteria block). Then a `rag_index` step (`collection='tool_docs'`).
- **Fix agents (`tool-improver`, `component-template-fixer`,
  `update_component_html` path) → `append_doc_note` as the LAST step** (body =
  the uniform entry; categories from the taxonomy; `site_id` for per-site).
- Fetch each definition, paste, migration drafted against the real JSON.

### Stage 5 — Tier-2 contract-presence check  ← DEFINITION
**What it is: a static, browserless verifier.** A Go check that loads the
tool's `criteria_json` (same query as `load_doc_context`) and asserts the
**statically visible subset** against the DEPLOYED artifacts — parse the
deployed page HTML and check `selector_exists` / `selector_count`; confirm
`asset_loads` (`/tools/assets/<function>.js` present); plus the standing shell
checks (`<no value>`, empty schema, header not leaked). **Nothing executes** —
that is what makes it cheap enough for every sweep, and why it catches the
markup-visible categories (`empty-shell`, `detool-on-rebuild`,
`js-not-extracted`, `broken-template-slots`) while `no_console_errors`,
`interaction`, and overflow remain Tier 4's job.
**Home:** a new pass inside `check_tool_health` (it already sweeps deployed
tools) — read `check_tool_health.go` first to place it. Runs only when a
current PLAN with a criteria block exists; a tool without one gets a
`needs_criteria` note, never a fake pass. **Failures →** `improve_tool` item
carrying the failing criterion (as `acceptance_test`) + an `acceptance-fail`
note. **Precondition:** at least one tool PLAN with criteria exists (Stage 4,
or one hand-written PLAN via `write_doc_plan`/SQL to pilot).

### Stage 6 — Runner P0  ← DEFINITION
**What it is: the smallest end-to-end slice of the Tier-4 headless runner.**
One new adapter deployment (`browser-runner-adapter` in `ai-persona-system`;
image = Chromium + Playwright, `playwright-go` consumer), two Kafka topics in
the analyser-adapter request/response shape, implementing EXACTLY three check
types on the DESKTOP profile only (1366×900): `page_status_ok`,
`selector_exists`, `no_console_errors`.
Request: `{run_id, urls, profiles:["desktop"], criteria_json, function,
site_id}` → Response: `{run_id, results:[{check_id, profile, url, pass,
detail}]}` on the caller's response topic.
**Driven by a hand-produced request** against ONE known live tool page — no
agent workflow, no improvement-loop wiring, no mobile, no screenshots in P0.
**Exit test:** the hand request's pass/fail matches manual inspection of that
page. **Deliverables:** Dockerfile/image, k8s Deployment, topic creation,
consumer main.go, response schema. Mobile profile, full check interpreter,
interactions, screenshots, and the `tool-acceptance-agent` are P1–P3
(`PLAN_tool_acceptance_runner.md`).

---

## 1. Where the docs live

- **Truth (Postgres, LIVE):** `doc_plans` (current + history) and `doc_notes`
  (append-only), keyed `(subject_type, subject_key)`:
  `('tool', <function>)` byte-for-byte from `content_components.function`;
  `('pipeline', <one of: build | content | design | maintenance>)`.
- **Retrieval index (derived):** `knowledge_base` collections via `rag_index`;
  `rag_lookup` for discovery only. Never edited directly.
- **Mirror (optional, Phase B):** rendered markdown to a docs repo.

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

**How:** `write_doc_plan` (supersede tx; one current row enforced by
`idx_doc_plans_current`). Then `rag_index`. Rollback = restore a prior row.
`pinned=true` = human hold. Keep orchestration ids/dates out of the inline
`<script>` header (019) — provenance is columns.

## 3. Appending a NOTES entry

**When:** any fix (fix agent's **last step**); workflow-altering migrations
(pipeline note: number, what, why); acceptance runs (§0 Stage 5/6); diagnosis
runs (§4). **How:** `append_doc_note` (one INSERT; body = the uniform entry;
`site_id` for per-site incidents).

**Taxonomy:** `css-variable-mismatch`, `empty-shell`/`mode-b-template`,
`broken-template-slots`, `content-vs-runtime-mismatch`, `detool-on-rebuild`,
`js-not-extracted`, `js-bundle-stale`, `schema-template-drift`, `diagnosis`,
`unconfirmed-diagnosis`, `migration`, `seam`, `acceptance-run`,
`acceptance-fail`, `truncated-output`, `needs_criteria`.

## 4. Persisting diagnosis output

`persist_diagnosis_note` runs **after** `diagnose_emit` (emit stays read-only).
Explicit subject in `input_data` only — **skip, don't guess**. CONFIRMED →
root-cause entry; UNVERIFIABLE → persisted tagged `unconfirmed-diagnosis` (dead
ends stop retries). `source='diagnosis-loop'`.

## 5. Tool acceptance & iteration

1. **Criteria live in the tool's PLAN** (fenced block; extracted by
   `load_doc_context` as `criteria_json`). Per-site parametrisation →
   `direction.must_have`, not the PLAN.
2. **Ladder:** Tier 0 generation-time (`HasToolDocHeader` +
   `check_tool_completeness`, flags-but-passes deliberately) → Tier 1
   structural (`check_tool_health`) → Tier 2 contract-presence (§0 Stage 5) →
   Tier 3 acceptance audit (`tool-auditor` vs criteria) → Tier 4 headless
   runner (§0 Stage 6; `PLAN_tool_acceptance_runner.md`).
3. **Iteration:** deploy → run → failing criterion → `improve_tool` item
   (criterion as `acceptance_test`, bounded by `max_fix_attempts`) → fixer
   loads PLAN+NOTES → fix → note → re-run. *Working* = criteria pass.
4. **Multi-page prerequisite:** preserve-sections re-render +
   interactivity-aware save guard (pending) before scaling page counts.

## 6. Loading context at fix time

`load_doc_context`: current PLAN + latest-N NOTES + `criteria_json`, composed
as one prompt-ready `doc_context` block. `has_plan=false` is a normal state.
For the code-diagnosis loop, hand `doc_context` in the way `runtime_evidence`
is handed to `diagnose_assemble_bundle`. `rag_lookup` for discovery. Read
**Deliberate decisions** before changing anything.

## 7. Verification

1. One current PLAN per subject (partial unique enforces).
2. Notes append with non-empty categories; diagnosis runs on an explicit
   subject leave a `diagnosis` entry; runs without one leave NO row.
3. Roll-up uses GIN: `SELECT subject_key, created_at FROM doc_notes WHERE
   categories ? 'detool-on-rebuild';`
4. `criteria_json` extracts from the PLAN (fence intact) and parses as JSON.
5. Acceptance: failures created `improve_tool` items carrying the criterion;
   passes left an `acceptance-run` note.
6. Inline header untouched: sentinel in `rendered_html`, stripped from shipped
   page + `/tools/assets/<function>.js`.
7. Pipeline topology claims: regenerate from `agent_definitions`.

0 rows is not decisive until the query itself is ruled out.

## 8. Gotchas

- Truth is Postgres, not git — never the home of an append.
- NOTES = one row per entry, never a shared file (RMW/lost-update).
- Skip, don't guess the diagnosis subject; a mis-filed note poisons history.
- `check_tool_completeness` **deliberately** lets flagged output through — wire
  a `truncated-output` note instead of "fixing" it.
- Tier 2 is static: it can never verify behaviour; don't let a Tier-2 pass be
  read as "the tool works" — that claim belongs to Tier 4.
- Don't hand-draw pipeline topology; author invariants/rationale/seams only.
- Criteria describe the tool TYPE; site-specific expectations →
  `direction.must_have`.
- `rag_index` labels `source_type='scrape'` — parameterise if it matters.
- Workflows flat; complexity in Go; `logger.Info`; check schemas before SQL;
  pods `-n ai-persona-system`, Kafka `-n kafka`; deploys via GitHub → Actions →
  Backblaze.
