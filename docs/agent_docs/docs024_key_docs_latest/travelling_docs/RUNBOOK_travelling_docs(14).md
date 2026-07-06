# RUNBOOK — Travelling Docs (PLAN + NOTES): Tools, Complex Components, Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-06 (rev 15 — Stage 3 tidied: status/history compressed, §3a CLOSE-OUT checks consolidated, §0-REF state-check queries added; 3a = trigger RUN, verification PENDING)

---

**For someone new to this runbook:** this system builds and maintains websites
with autonomous agents. Every complex tool/interactive component and every
pipeline now carries its own **travelling documentation in Postgres** — a
**PLAN** (what it is for, how it is delivered, its deliberate decisions, and
its **acceptance criteria**: the per-tool definition of *working*) and a
**NOTES** log (every fix, diagnosis, and dead end, tagged by category). Agents
write these docs as a byproduct of the steps that create and fix things, and
**load them before touching a subject**, so fixes build on prior decisions
instead of re-deriving lost context. Acceptance criteria are then checked in
tiers — from static presence checks up to driving the deployed tool in a
headless browser (desktop + mobile) — iterating until the criteria pass. This
runbook is the operating manual: where the docs live, how each write/read
happens, and (§0) the staged rollout with exactly what to run at each step and
a marker showing where we are. Design rationale: `PLAN_travelling_docs.md`;
the headless runner's own plan: `PLAN_tool_acceptance_runner.md`.

**Applies to:** tools (`content_components.component_level='tool'`),
interactive/stateful components, and pipelines. Pipeline subject keys (live,
confirmed 2026-07-04): `build`, `content`, `design`, `maintenance`.

---

## 0. ROLLOUT TRACKER — YOU ARE HERE

**Position as of 2026-07-06:**
Stages 0–2 ✅ (tables live · four actions on production · helper tests 6/6) ·
**Stage 3a: wiring APPLIED ✓ · `load_runtime` error-routing APPLIED ✓ (fired ×5
in run-3 — normal for anchorless) · trigger RUN ✓ (run-3, corr `8332c2a5…`) ·
VERIFICATION PENDING → run the §3a CLOSE-OUT block** ·
3b next · Stages 4–6 pending · Pilot PLAN: unblocked (any time after Stage 3).

### §0-REF — State-check queries (copy-paste)

```sql
-- 1. Orchestration by correlation (never by created_at):
SELECT status, current_step,
       EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s,
       substring(COALESCE(error,''),1,200) AS err
FROM orchestration_states
WHERE correlation_id = '<uuid>'::uuid ORDER BY created_at;

-- 2. Children of an orchestration:
SELECT status, current_step, orchestration_name, created_at
FROM orchestration_states
WHERE parent_orchestration_id = '<orchestration-uuid>'::uuid ORDER BY created_at;

-- 3. Travelling docs — current PLAN / latest NOTES / category roll-up:
SELECT subject_key, is_current, pinned, created_at
FROM doc_plans WHERE subject_type = '<tool|pipeline>' AND subject_key = '<key>';

SELECT created_at, categories, left(body,100) AS head
FROM doc_notes WHERE subject_type = '<tool|pipeline>' AND subject_key = '<key>'
ORDER BY created_at DESC LIMIT 10;

SELECT subject_key, count(*) FROM doc_notes
WHERE categories ? '<tag>' GROUP BY 1 ORDER BY 2 DESC;
```
```bash
# Pods — label key is agent-type (HYPHEN); the selector spans ALL live pods of
# the type, so attribute lines by pod / orchestration_id / timestamp:
kubectl -n ai-persona-system get pods -l agent-type=<agent> --sort-by=.metadata.creationTimestamp
```
Reading rules: parent COMPLETED ≠ child success — check the CHILD row and the
BODY status; a COMPLETED parent with a non-empty `error` = a forwarded child
failure; 0 rows is decisive only after the query itself AND the run's
completion are ruled in.

---

### Stage 0 ✅ Pre-migration gates — DONE 2026-07-04
No `doc_plans`/`doc_notes` collision; `site_work_items.pipeline` live values =
`build` (3579), `content` (24), `design` (13), `maintenance` (2); no CHECK
constraint. These four values are the valid pipeline `subject_key`s.

### Stage 1 ✅ Migration — APPLIED 2026-07-04
Statement tally verified (2 CREATE TABLE, 5 CREATE INDEX, 3 COMMENT, COMMIT).
Optional table smoke (incl. the DELIBERATE duplicate-insert error proving
`idx_doc_plans_current`, and cleanup):

```sql
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool','smoke-test-tool','smoke v1','human','smoke');

-- EXPECTED ERROR: duplicate key violates "idx_doc_plans_current"
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool','smoke-test-tool','smoke v2','human','smoke');

UPDATE doc_plans SET is_current=false, superseded_at=now()
 WHERE subject_type='tool' AND subject_key='smoke-test-tool' AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool','smoke-test-tool','smoke v2','human','smoke');

SELECT subject_key, is_current, superseded_at IS NOT NULL AS superseded, created_at
FROM doc_plans WHERE subject_key='smoke-test-tool' ORDER BY created_at;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('tool','smoke-test-tool','## 2026-07-04 — smoke','["diagnosis"]'::jsonb,'human','smoke');
SELECT subject_key, categories FROM doc_notes WHERE categories ? 'diagnosis';

DELETE FROM doc_notes WHERE subject_key='smoke-test-tool';
DELETE FROM doc_plans WHERE subject_key='smoke-test-tool';
```

### Stage 2 ✅ Go actions — ON PRODUCTION 2026-07-04
`write_doc_plan`, `append_doc_note`, `load_doc_context`,
`persist_diagnosis_note` deployed with their registry entries.

### Stage 3 — `persist_diagnosis_note` wiring + gate verification

**Status:** wiring APPLIED ✓ (live shape: `emit` → `persist_note`
(`config.error_step="complete"`) → `complete`; `result_from:"diagnosis"`
untouched) · `load_runtime` error-routing APPLIED ✓ (target DERIVED =
`assemble_bundle`; fired ×5 in run-3 — NORMAL for anchorless runs) · trigger
RUN ✓ (run-3: correlation `8332c2a5-bf89-49e0-b7b3-fb2cf83e868e`, child
`diagnose-agent-workflow-1407`, pod `agent-diagnose-agent-135b53b4-w4gbl`) ·
**VERIFICATION PENDING** — the last capture (diagnoselogs4) was mid-loop, and
the gate cannot pass on an in-flight capture by its own definition.

**Why runs 1–2 didn't count (compressed):** run-1 died at `load_runtime` (no
anchor, no error routing); run-2's routing fired but targeted a non-existent
step name. Full narratives + durable rules live in RUNNING_NOTES (2026-07-06
entries) and 016b v5 §9 — error_step placement/targets, anchorless-diagnosis,
`agent-type` label, failure envelopes. Do not retell them here.

**Design facts that hold:** this run SKIPS persistence BY DESIGN —
`diagnose-orchestrator.call_diagnoser.input_mapping` doesn't forward
`subject_type`/`subject_key` until 3b, and proving the skip IS the 3a test.
Trigger via the orchestrator only (the spawned pod gets `GITHUB_READ_TOKEN`);
`REF` explicit, never HEAD.

**3a CLOSE-OUT — run these now, in order:**

1. **Run state** (child row = the one at workflow steps; parent =
   `call_diagnoser`):
```sql
SELECT status, current_step,
       EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s,
       substring(COALESCE(error,''),1,200) AS err
FROM orchestration_states
WHERE correlation_id = '8332c2a5-bf89-49e0-b7b3-fb2cf83e868e'::uuid
ORDER BY created_at;
```
   Read: child COMPLETED at `complete` with empty err → step 2. Child still
   EXECUTING with small `since_s` → still looping (≤5 iterations, minutes
   each); re-check later. `since_s` ≫ 1800 or FAILED → paste back, don't
   proceed. A COMPLETED parent alone proves nothing (body-status rule).
2. **Gate log line** (pod-scoped to dodge multi-pod residue):
```bash
kubectl -n ai-persona-system logs agent-diagnose-agent-135b53b4-w4gbl --tail=4000 | grep persist_diagnosis_note
# expect: persist_diagnosis_note: no explicit subject — skipping (do not guess)
```
   (Fallback: `-l agent-type=diagnose-agent` — hyphen — then attribute by
   pod / orchestration id / timestamp.)
3. **DB gate** — decisive only with 1 (child COMPLETED) + 2 (line present):
```sql
SELECT count(*) AS diagnosis_notes
FROM doc_notes
WHERE categories ? 'diagnosis'
  AND created_at > now() - interval '6 hours';   -- expect 0
```
4. *(Optional — what the loop concluded for the smoke symptom):*
```bash
kubectl -n ai-persona-system logs agent-diagnose-agent-135b53b4-w4gbl --tail=4000 | grep -E "report shaped|stopped_by"
# expected: status UNVERIFIABLE, stopped_by iteration-cap
```
All green → **3a CLOSED**: update the §0 position line, proceed to 3b.

**Canonical trigger** (re-runs + 3b's positive run):
`drafts/084_TRIGGER_diagnose_v1.sh` — kcat → `system.agent.generic.requests`,
`agent_type=diagnose-orchestrator`, envelope mirrors 082/083c; subject fields
commented until 3b; anchorless runs now survive via the routed degrade.

**3b — make it write (after close-out):** thread the subject through, fetch-first:
- `diagnose-agent.input_contract.optional` += `subject_type`, `subject_key`
  (they already flow via `input_data.*` when present);
- `diagnose-orchestrator` `call_diagnoser.input_mapping` +=
  `"subject_type?": "input_data.subject_type"`,
  `"subject_key?": "input_data.subject_key"`.
Then the SAME trigger with subject fields uncommented (e.g. `pipeline`/`build`)
exercises the positive path — the first machine-written NOTES row. Verify:
```sql
SELECT subject_type, subject_key, categories, left(body,120) AS body_head, created_at
FROM doc_notes ORDER BY created_at DESC LIMIT 5;
```
The 3b migrations get drafted on request (against the two live definitions,
targets verified per the routing rule).

### Pilot PLAN (optional — UNBLOCKED NOW; tables live, no wiring needed)
Seeds the first real tool PLAN by SQL, satisfies Stage 5's precondition early,
and dogfoods the format. Later `write_doc_plan` calls supersede it cleanly.

```sql
-- 1) candidates (pick one function, byte-for-byte):
SELECT function, component_level, is_active, left(description,60) AS descr
FROM content_components
WHERE component_level = 'tool'
ORDER BY function;

-- 2) seed (replace <function> throughout; body is dollar-quoted):
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool', '<function>', $plan$# PLAN — <function>

## Aim
<what the tool is for, in product terms>

## Source spec
<the site-spec / roadmap slice it derives from>

## Behaviour contract
<states, inputs, outputs>

## Acceptance criteria
```criteria
{ "profiles": ["desktop","mobile"],
  "checks": [
    {"id":"boots","type":"selector_exists","selector":"#<root-element-id>"},
    {"id":"console","type":"no_console_errors"},
    {"id":"asset","type":"asset_loads","path":"/tools/assets/<function>.js"} ] }
```

## Delivery mechanism
<Path 1 inline script -> /tools/assets/<function>.js | Path 2 js_snippet | build-time content> — and why

## Dependencies
<data feeds, assets, other components, scheduled tasks>

## Deliberate decisions — do not re-fix
<choices a later pass must not "fix">
$plan$, 'human', 'pilot');

-- 3) verify the fence is intact:
SELECT subject_key, is_current, body LIKE '%```criteria%' AS has_fence
FROM doc_plans WHERE subject_type='tool' AND is_current;
```

### Stage 4 — remaining wiring (same fetch-first pattern per agent)
> **Caution (001 §16):** every step added here puts `error_step` INSIDE
> `config`. The live `tool-recreation-handler` and `tool-auditor` definitions
> carry step-LEVEL `error_step` on several steps — dormant instances of the
> documented bug (silently ignored). Do not copy that shape; optionally correct
> adjacent instances while touching those workflows, as its own noted change.
Resolved agent types (from the live `agent_definitions` list):
- **`tool-generator` → PLAN at creation.** Needs a PLAN-body composition step
  before `write_doc_plan` (the action reads `doc_plan_body`): compose from
  material in hand (spec slice, delivery mechanism, deliberate decisions,
  initial criteria block). Then `rag_index` (`collection='tool_docs'`). Same
  hook later on **`component-creator`** for non-tool complex components.
- **Fix agents → `append_doc_note` as the LAST step:** `component-template-fixer`,
  `tool-improver`, and **`tool-recreation-handler`** (the recreation path;
  `update_component_html` is an ACTION, not an agent). 
- Fetch each definition (same A/B/C pattern), paste, migration drafted against
  the real JSON, targeting the current version row.

### Stage 5 — Tier-2 contract-presence check  ← DEFINITION
**A static, browserless verifier.** Loads the tool's `criteria_json` (same
query as `load_doc_context`) and asserts the **statically visible subset**
against the DEPLOYED artifacts — parse deployed page HTML for
`selector_exists`/`selector_count`; confirm `asset_loads`
(`/tools/assets/<function>.js` present); plus shell checks (`<no value>`,
empty schema, header not leaked). **Nothing executes** — cheap enough for
every sweep; catches markup-visible categories (`empty-shell`,
`detool-on-rebuild`, `js-not-extracted`, `broken-template-slots`).
`no_console_errors`, `interaction`, overflow remain Tier 4's job.
**Home:** a new pass inside `check_tool_health` — read `check_tool_health.go`
first to place it. Runs only when a current PLAN with criteria exists; a tool
without one gets a `needs_criteria` note, never a fake pass. **Failures →**
`improve_tool` item carrying the failing criterion (as `acceptance_test`) +
an `acceptance-fail` note. **Precondition:** ≥1 tool PLAN with criteria (the
pilot above, or Stage 4).

### Stage 6 — Runner P0  ← DEFINITION
> **2026-07-06:** contract pinned to the 035 Adapter Guide (normative): topic
> `system.adapter.browser-runner.requests` (Convention A), matcher
> `in_response_to_request_id` = incoming `request_id`, typed header struct with
> real bools, fresh `message_id`, `ProduceWithValidation` never plain `Produce`.
> Body `run_id` demoted to payload detail. Full contract:
> `PLAN_tool_acceptance_runner.md` (rev 2).
**The smallest end-to-end slice of the Tier-4 headless runner.** One adapter
deployment (`browser-runner-adapter` in `ai-persona-system`; image = Chromium
+ Playwright, `playwright-go` consumer), two Kafka topics (analyser-adapter
request/response shape, Strimzi KafkaTopic CRs on `personae-kafka-cluster`),
implementing EXACTLY three check types, DESKTOP profile only (1366×900):
`page_status_ok`, `selector_exists`, `no_console_errors`.
Request `{run_id, urls, profiles:["desktop"], criteria_json, function,
site_id}` → Response `{run_id, results:[{check_id, profile, url, pass,
detail}]}` on the caller's response topic.
**Driven by a hand-produced request** against ONE known live tool page — no
agent workflow, no loop wiring, no mobile, no screenshots. **Exit test:**
results match manual inspection of that page. **Deliverables:**
Dockerfile/image, k8s Deployment, KafkaTopic CRs, consumer main.go, response
schema. Mobile, full interpreter, interactions, screenshots, and the
`tool-acceptance-agent` are P1–P3 (`PLAN_tool_acceptance_runner.md`).

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
**## Acceptance criteria** with a fenced block (schema v0 + check types:
`PLAN_tool_acceptance_runner.md`; skeleton: the pilot INSERT above). Multi-page
tools add Page set & inter-page contract. · Delivery mechanism + why ·
Dependencies · Deliberate decisions.

**Pipelines:** authored initially (distil 004–008); Aim · Invariants · Branch
rationale · Seams · Deliberate decisions. **Never embed the step map** — derive
from `agent_definitions`.

**How:** `write_doc_plan` (supersede tx; one current row enforced by
`idx_doc_plans_current`). Then `rag_index`. Rollback = restore a prior row.
`pinned=true` = human hold. Keep orchestration ids/dates out of the inline
`<script>` header (019) — provenance is columns.

## 3. Appending a NOTES entry

**When:** any fix (fix agent's **last step**); workflow-altering migrations
(pipeline note: number, what, why); acceptance runs; diagnosis runs (§4).
**How:** `append_doc_note` (one INSERT; body = the uniform entry; `site_id`
for per-site incidents).

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
   runner (§0 Stage 6).
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

- `error_step` (and any routing target) must name an EXISTING step — the
  coordinator fails the whole workflow on an unknown target (`step 'X' not
  found`). Verify against the live step map, or derive the target from the
  step's own `next_step` when converging success/failure paths.
- Loop corollary (001 Appendix C + `loop_expansion_handler.go`): inside a
  loop's substeps, `error_step`/`then_step`/`fallback_step` values are
  ITERATION-PREFIXED at expansion — they must name SUBSTEPS of the same loop,
  never a top-level step. `continue_on_error: true` on the loop is the
  iteration-scoped alternative (record the error, advance to the next
  iteration — `loop_error_handler.go`). Applies to any Stage-4 note-append
  step placed inside a loop body.
- Truth is Postgres, not git — never the home of an append.
- NOTES = one row per entry, never a shared file (RMW/lost-update).
- Skip, don't guess the diagnosis subject; a mis-filed note poisons history.
- `check_tool_completeness` **deliberately** lets flagged output through — wire
  a `truncated-output` note instead of "fixing" it.
- Tier 2 is static: never read a Tier-2 pass as "the tool works" — that claim
  belongs to Tier 4.
- Don't hand-draw pipeline topology; author invariants/rationale/seams only.
- Criteria describe the tool TYPE; site-specific expectations →
  `direction.must_have`.
- `rag_index` labels `source_type='scrape'` — parameterise if it matters.
- Workflows flat; complexity in Go; `logger.Info`; check schemas before SQL;
  pods `-n ai-persona-system`, Kafka `-n kafka`; deploys via GitHub → Actions →
  Backblaze.
