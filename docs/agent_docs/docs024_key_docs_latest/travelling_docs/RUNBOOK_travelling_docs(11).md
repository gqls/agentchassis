# RUNBOOK — Travelling Docs (PLAN + NOTES): Tools, Complex Components, Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-06 (rev 12 — run-2: error routing FIRED but target name wrong (`assemble` vs live `assemble_bundle`); corrective migration derives target from next_step; live step map recorded)

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

**Position as of 2026-07-06 (late):**
Stage 0 ✅ · Stage 1 ✅ · Stage 2 ✅ (+ helper tests 6/6 PASS) · Stage 3a wiring
APPLIED · run-1 failed pre-gate (no anchor, no routing) → routing migration
APPLIED · **run-2: routing FIRED (`error_routed` in ProcessingHistory) but the
target name was wrong — `assemble` vs the live step `assemble_bundle`; gate
still NOT verified** ·
**▶ NEXT: apply `drafts/0NN_fix_load_runtime_error_step_target.sql` (derives the
target from `load_runtime.next_step` — no name guessing), re-run the anchorless
trigger → gate checks** · 3b + Stages 4–6 pending. Pilot PLAN: unblocked.

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

### Stage 3 ⏳ first wiring: `persist_diagnosis_note` — MIGRATION DRAFTED
> **2026-07-06 guideline fix (re-pull before applying):** the persist_note step
> now carries `"error_step": "complete"` INSIDE its `config` — 001 §16: the
> coordinator reads `step.Config["error_step"]`; step-level `error_step` is
> silently ignored. A doc_notes write failure now degrades to `complete`, so the
> caller still receives the diagnosis (`result_from: "diagnosis"` reads emit's
> output, not persist's).
Resolved from the live `diagnose-agent` `default_config->'workflow'` (object-keyed
steps): `emit` runs `diagnose_emit` (`output_field: "diagnosis"`,
`next_step: "complete"`); `complete` runs `complete_workflow`
(`result_from: "diagnosis"`). The wiring inserts `persist_note` BETWEEN them
(`emit.next_step → persist_note → complete`), so `result_from` is unaffected and
the caller still gets the diagnosis. Migration:
`drafts/0NN_wire_persist_diagnosis_note.sql` (patches `default_config` — the live
column; leaves the deprecated `orchestrator_workflow` copy alone; targets the
current version row; guards that exactly one row is wired).

**⚠ Subject-threading caveat (read before you run):** `diagnose-agent`'s input
contract requires only `symptom`; it does NOT yet carry `subject_type`/
`subject_key`, and `diagnose-orchestrator` doesn't pass them. So the FIRST runs
will correctly **skip** (`persisted:false`, no row) via the action's
skip-don't-guess gate. That is expected, not a bug — the step is live and inert
until a subject-aware caller supplies a subject (Stage 3b).

**3a — APPLIED ✅ 2026-07-06.** Migration applied on live (BEGIN, UPDATE 1, DO
guard, COMMIT); verify query confirmed the shape: `emit.next_step =
"persist_note"`; `persist_note` carries `config.error_step = "complete"`,
`next_step: "complete"`, `output_field: "diagnosis_note"`. (The garbled psql
paste mid-file was a terminal display artifact — execution was clean.)

**3a-RUN 2026-07-06 — FAILED PRE-GATE; gate NOT yet verified.** The first
trigger spawned the child, which ran ~91s (analyse + lookup) then hard-failed at
`load_runtime`: `need at least one of site_id / correlation_id / domain in
collected_data`. The workflow never reached `persist_note`, so this run's 0-row
count is NOT decisive (the 0-rows rule, applied to ourselves) and no skip line
exists. Two causes: the smoke input carried no runtime anchor, and `load_runtime`
has no effective error routing — runtime evidence is an optional bundle tier
(`diagnose_assemble_bundle` omits it when empty) but the step makes it mandatory
in practice. Observed en route: the orchestrator row shows COMPLETED while the
child FAILED — the child reports failure with header `status: complete` and
`body.status: failed`, so a diagnosis consumer must check the BODY status.

Fix paths:
1. **Structural (recommended):** apply
   `drafts/0NN_diagnose_load_runtime_error_step.sql` — config-level
   `error_step: "assemble"` on `load_runtime` (001 §16 mechanism; success and
   failure converge on `assemble`, so an anchorless code-only diagnosis proceeds
   with a code+schema bundle). Then re-run the SAME anchorless trigger.
2. **Immediate alternative (no migration):** re-run with an anchor:
   `RUNTIME_SITE=gamesdesign.co.uk ./drafts/084_TRIGGER_diagnose_v1.sh`.
Follow-up (next chassis build): soften `diagnose_load_runtime` to return empty
evidence + `skipped:true` for the no-anchor case, keeping hard errors for real
DB failures.

**Pod label key is `agent-type` (hyphen)** — corrected below and in the script;
the underscore selector matches nothing.

**3a-RUN-2 2026-07-06 — routing FIRED; target name wrong; gate still NOT
verified.** ProcessingHistory: `error_routed … routed to assemble: step
load_runtime failed …` — the config-level mechanism is live-validated (and the
config merge preserved load_runtime's five existing keys) — then the coordinator
failed: `step 'assemble' not found` (`routeToErrorStepOrFail →
continueExecution`). The **live step map**, captured from the run dump:
`analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict →
route → emit → persist_note → complete` — the assemble step is
**`assemble_bundle`**; the previous target was an unverified name. Corrective:
`drafts/0NN_fix_load_runtime_error_step_target.sql` sets `error_step` to the
value of `load_runtime.next_step` **read from the same row** (convergence by
construction; the dump confirms next_step = `assemble_bundle`; guard checks the
target exists in the step map). Loop-back note: `route.config.gather_step =
"load_runtime"`, so loop-back failures now also converge on `assemble_bundle`.
Also observed: this failure class notifies the parent with
`status: error_unrecoverable` / `code: CHILD_ORCHESTRATION_FAILED` — a second
failure-envelope shape alongside run-1's header-complete + body-failed.

**3a-trigger — run one SUBJECTLESS diagnosis.** Canonical trigger:
`drafts/084_TRIGGER_diagnose_v1.sh` (envelope mirrors 082/083c). Target is
**`diagnose-orchestrator`** — NOT `diagnose-agent` directly: an in-place run on
a shared chassis pod has no `GITHUB_READ_TOKEN`, so `analyse_repo_local` fails
immediately (083c pre-flight). `REF` explicit, never HEAD. This run WILL skip
persistence by design — the orchestrator's `input_mapping` doesn't forward
`subject_type`/`subject_key` until 3b; proving the skip IS the test.

```bash
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid); ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid);     MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
kubectl -n kafka run -i --rm kcat-diagnose-$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID -H orchestration_name=manual-diagnose-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start -H client_id=demo_client -H message_type=request -H action=orchestrate \
  -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"diagnose-orchestrator"},"input_data":{"symptom":"smoke: subjectless diagnosis to verify persist_note skip gate","owner":"gqls","repo":"agentchassis","ref":"main"}}
JSON
```

**3a checks (in order):**
1. Spawned pod + token: `kubectl -n ai-persona-system get pods -l
   agent-type=diagnose-agent --sort-by=.metadata.creationTimestamp | tail -3`;
   `describe` it and grep `GITHUB_READ_TOKEN` (mind the empty-grep trap — an
   empty match + bare `describe pod` describes EVERY pod; 083c note).
2. Follow by correlation id: `kubectl -n ai-persona-system logs -f -l
   agent-type=diagnose-agent --tail=500 | grep $CORRELATION_ID` — markers:
   `analyse_repo_local → lookup_code_symbols → diagnose_load_runtime →
   diagnose_assemble_bundle → verdict → route → diagnose_emit →
   persist_diagnosis_note`.
3. State by correlation id, never `created_at`:
   `SELECT status, current_step, substring(COALESCE(error,''),1,200) FROM
   orchestration_states WHERE correlation_id='<id>'::uuid ORDER BY created_at;`
4. **Gate verification — only meaningful once status = COMPLETED (0-rows rule):**
   the log line `persist_diagnosis_note: no explicit subject — skipping (do not
   guess)` is present, AND
   `SELECT count(*) FROM doc_notes WHERE categories ? 'diagnosis' AND created_at
   > now() - interval '2 hours';` returns 0. The COMPLETED status + the skip
   line are what make the 0 decisive.

Timing: minutes, not seconds (repo tarball fetch + up to 5 verdict iterations;
timeout 1800s).

**3b — make it write (subject threading, small follow-on):** to actually persist,
a caller must pass a subject. Two edits, drafted when you want them:
- `diagnose-agent` input contract + `assemble`/emit unaffected — just accept
  `subject_type?`, `subject_key?` in `input_contract.optional` (they already flow
  via `input_data.*` if present).
- `diagnose-orchestrator` `call_diagnoser.input_mapping` gains
  `"subject_type?": "input_data.subject_type"`, `"subject_key?": "input_data.subject_key"`,
  so whoever spawns a diagnosis for a known tool/pipeline (e.g. a tool audit that
  has the `function`) threads it through.
Then a subject-carrying run leaves a `diagnosis` (or `unconfirmed-diagnosis`) row:
```sql
SELECT subject_type, subject_key, categories, left(body,120) AS body_head, created_at
FROM doc_notes ORDER BY created_at DESC LIMIT 5;
```

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
