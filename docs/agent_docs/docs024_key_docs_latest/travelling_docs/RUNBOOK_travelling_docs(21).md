# RUNBOOK — Travelling Docs (PLAN + NOTES): Tools, Complex Components, Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-07 (rev 24 — Task 3 APPROVED and DRAFTED: drafts/0NN_tool_generator_plan_writing.sql (snapshot embedded, three doc steps with error containment, save_tool rewire, timeout 300→480, the ×3 error_step corrections))

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

**Position as of 2026-07-07 (later): STAGE 3 ✓ · PILOT ✓ · TASK 3 MIGRATION DRAFTED — apply + verify next**
Stamp facts: `('pipeline','build')`, categories `["diagnosis","unconfirmed-diagnosis"]`,
`source='diagnosis-loop'`; the body records the loop's stop reason —
**`scope-not-narrowing`** — answering the fast-run question from the note
itself (first payoff of the feature). `source_agent` came back empty (the
`agent_type` header is absent in that step context; backlog item — provenance
intact via `source`).

### YOUR TASKS — plain version

**TASK 1 ✅ DONE 2026-07-07 — snapshots taken; storage convention learned.**
Two uuids returned (the SOURCE row ids); the inspect shows agent_definitions
still holds ONLY the two original rows → snapshots live in a SEPARATE store.
Consequence: the `is_snapshot` current-row selector predicate is NOT
load-bearing — dropped from the migration mandate (harmless if present).
Optional curiosity, not blocking: `\sf snapshot_agent(text,text)` shows the
storage table; `\df *agent*` likely reveals the companion restore function.

**TASK 2A ✅ DONE 2026-07-07 12:32** — the system's FIRST tool PLAN is live:
`tool-archetype-taster-quiz`, `is_current=t`, `has_fence=t`, 2,761 chars.
Stage-5's precondition (≥1 tool PLAN with criteria) is met.
*About the `EDIT:` markers — nothing for you to do now.* Optional
fill-in-the-blank spots for details neither of us had to hand (exact answer
selectors; the spec slice). The PLAN is fully valid as seeded; details get
filled later by SUPERSEDE (never edit history in place) — when this PLAN is
next loaded alongside the component html, the filled body + supersede
statement get produced for you. Checks whose id ends `-EDIT` are skipped by
verification until real selectors replace them.
*Duplicates explained (indexdef paste):* uniqueness covers LIBRARY ORIGINALS
that are ACTIVE only (`forked_from IS NULL AND is_active = true`) — extra
rows are forks/inactive versions. Note: the quiz currently has NO active row
anywhere — fully dormant; a product fact, not a blocker.

**TASK 3 (APPROVED — MIGRATION DRAFTED) — automate PLAN-writing at tool creation.**
File: `drafts/0NN_tool_generator_plan_writing.sql`. The snapshot call is the
FIRST statement inside the transaction, so applying the file satisfies the
standing rule in one go (MVCC captures the pre-update state; a failure rolls
everything back together).
*Your part now (2 steps + optional proof):*
1. Renumber `0NN` and apply the file:
   `psql "$CLIENTS_DB_URL" -f drafts/0NN_tool_generator_plan_writing.sql`
   (expect: one NOTICE "Snapshot captured: type=tool-generator …", three
   UPDATE 1, DO, COMMIT).
2. Run the verify at the file tail — expect `save_next = compose_plan`,
   `subject_key_field = input_data.spec.function`, `timeout = 480`,
   `step_level_removed = t` — and paste it.
3. *Proof (optional):* the NEXT tool creation should leave a `doc_plans` row —
   `SELECT subject_key, is_current, body LIKE '%```criteria%' AS has_fence,
   source, created_at FROM doc_plans WHERE source='tool-generator' ORDER BY
   created_at DESC LIMIT 3;` Proof options: (a) wait for the next organic
   creation via the tool pipeline, or (b) a manual orchestrate against a TEST
   site only — it creates a REAL component + page + nav entry (side effects),
   so not against a live site. Say the word if you want the manual trigger
   scripted for a test site.
*What the file changes (all declared inside it):* three new steps after
`save_tool` — `compose_plan` (Sonnet drafts the PLAN body; five standard
checks verbatim; an interaction check ONLY from real selectors copied out of
the generated HTML, never invented; ≤3000 chars) → `write_plan`
(`write_doc_plan`) → `index_plan` (`rag_index` into `tool_docs`) — each with
`config.error_step: "complete"` so docs can never fail tool creation; the
`save_tool.next_step` rewire; `timeout_seconds` 300 → 480 (second Sonnet
call); and the correct-while-touching corrections — the three inert
step-level `error_step`s move into `config` with their ORIGINAL targets and
the dead keys are deleted.

**Background — no action now:** next chassis build: helper tests +
`diagnose_load_runtime` softening + `source_agent` fallback. Standing opens:
KB `tool_docs` write; `deploy_tool_to_site` `source_*` stamp; `rag_index`
`source_type`.

### Plain terms (mini-glossary)

- **`EDIT:` marker** — a fill-in-the-blank in a seeded document; optional,
  fill whenever the detail is known; the doc is valid meanwhile.
- **`-EDIT` check id** — an acceptance check with placeholder selectors;
  verification skips ids ending `-EDIT` until real selectors replace them.
- **fence / `has_fence`** — the ```criteria …``` block inside a PLAN body;
  `has_fence = t` means the machine-readable criteria are intact.
- **supersede** — how a PLAN is updated: flip the current row to
  `is_current=false`, insert the new body as current; never edit history in
  place.
- **subject / subject_key** — what a doc is about: `('tool', <function>)` or
  `('pipeline', build|content|design|maintenance)`.
- **anchorless / degrade** — a diagnosis run with no site/domain anchor;
  `load_runtime` fails and error-routes to `assemble_bundle`, continuing with
  a code+schema bundle; the `error_routed` lines are normal.
- **correct-while-touching** — when a migration already modifies a workflow,
  it also fixes known-inert bugs in that SAME workflow (e.g. step-level
  `error_step` → into `config`), declared explicitly; bounded repair, no
  separate campaign. (Norm adopted in this chat, 2026-07-06.)
- **"yes, draft the tool-generator migration"** — the approval phrase for
  Task 3 (formerly the shorthand "go B").

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
*(Placeholders like `<uuid>` are replaced INCLUDING the angle brackets — a
leading `<` inside the quotes is invalid uuid syntax.)*

**Snapshot before updating (STANDING RULE 2026-07-06)** — the platform's OWN
function; prepend to every agent-updating migration:
```sql
SELECT snapshot_agent('<type>', '<migration filename>: pre-update');
```
(`snapshot_agent(p_agent_type text, p_reason text DEFAULT NULL) RETURNS uuid`.)
Write-location CONFIRMED 2026-07-07: snapshots are stored OUTSIDE
`agent_definitions` (post-call inspect showed only the original rows; the
function returns the SOURCE row id, NOTICE carries source_version). The
`is_snapshot` selector predicate is NOT load-bearing — dropped from the
mandate (harmless if present). Restore path when first needed: `\df *agent*`
for the companion function. The side-table draft
(`drafts/0NN_agent_definition_snapshots.sql`) is SUPERSEDED un-applied.

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
**VERIFIED + CLOSED 2026-07-06.** Evidence (state form — the pod outlived its
stdout window): (1) child COMPLETED at `complete`, empty err (psql); (2)
ProcessingHistory: `route 14:30:09 → emit 14:30–14:31 → persist_note
14:31:37–14:33:02 → complete 14:33:09–14:33:32` — `persist_note` EXECUTED and
handed to `complete`; (3) `doc_notes` diagnosis-category count = 0.
Skip-vs-error settled structurally: the subject gate is the action's FIRST
check (returns `persisted:false` before any DB access) and no error appears in
state or the parent row. Observed cost of a full anchorless run: ≈26 min
(14:07→14:33), 5 iterations, 5 normal `error_routed` degrades. The loop's
verdict on the smoke symptom is not extractable from this capture (no
CollectedData; stdout window closed) — immaterial to the gate.

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
   *(Pod reaped — idle 3600s — before you grepped? A post-completion STATE DUMP
   is the accepted substitute: ProcessingHistory showing the step EXECUTED then
   the terminal step, plus the terminal status. Run-3 was closed this way.)*
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

**3b — make it write.**

**3b.1 ✅ DONE (2026-07-06).** Pasted facts: `call_diagnoser.input_mapping` has
NINE keys (no subject fields); BOTH input_contracts are identical twins —
required `["symptom"]`, eight optionals, no subject fields. Consequence
(016b §9 spawn+call rule: the mapping must satisfy the input_contract): the
threading is THREE edits, not two — mapping merge + `optional` additions on
BOTH contracts.

**3b.2 — migration APPLIED ✓ 2026-07-06** (`UPDATE 1` ×2, guard, COMMIT).
Applied BEFORE the snapshot standing rule — retroactive snapshots via
`snapshot_agent()` (§0 NEXT step 1; side-table draft superseded). File:
`drafts/0NN_wire_diagnosis_subject_threading.sql`
— orchestrator `input_mapping` += `"subject_type?"`/`"subject_key?"`
(object-merge preserves the nine keys); `optional` += the two names on both
contracts via a DISTINCT aggregate (idempotent; array order not significant);
current-version targeting; guards assert the exact mapping paths + both
contracts. DB-only — effective immediately, no deploy. Apply:
```bash
psql "$CLIENTS_DB_URL" -f drafts/0NN_wire_diagnosis_subject_threading.sql
```

**3b.3 — verify shape** ✅ VERIFIED 2026-07-06 (map paths = `input_data.subject_type`/`input_data.subject_key`; both contracts t/t):
```sql
SELECT default_config #>> '{workflow,steps,call_diagnoser,config,input_mapping,subject_type?}' AS map_type,
       default_config #>> '{workflow,steps,call_diagnoser,config,input_mapping,subject_key?}'  AS map_key,
       input_contract->'optional' ? 'subject_type' AS c_type,
       input_contract->'optional' ? 'subject_key'  AS c_key
FROM agent_definitions
WHERE type='diagnose-orchestrator' AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;

SELECT input_contract->'optional' ? 'subject_type' AS c_type,
       input_contract->'optional' ? 'subject_key'  AS c_key
FROM agent_definitions
WHERE type='diagnose-agent' AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
```

**3b.4 — positive run (the first machine-written NOTES row):**
```bash
SUBJECT_TYPE=pipeline SUBJECT_KEY=build \
  ./drafts/084_TRIGGER_diagnose_v1.sh "smoke: subject-carrying diagnosis — first machine-written NOTES row"
```
**Interim run `03bebec9…` (16:20:42, child at `assemble_bundle`) is
SUBJECTLESS** — its banner shows no `Subject:` line and the 3a default symptom,
so it will SKIP: a free post-3b regression of the gate, NOT this test. The env
vars above are load-bearing.
**Attempt `324456a9…` (18:2x) was ALSO subjectless** — the vars were set in the
shell but NOT exported, so the child saw nothing (banner tell: no `Subject:`
line, default symptom). Same-line prefix or `export`; the banner now prints an
explicit `Subject: NONE — will SKIP` when unset.
Expectations: ~26 min for a full anchorless loop (subject ≠ runtime anchor —
the routed degrade still fires ×5, still NORMAL); on completion,
```sql
SELECT subject_type, subject_key, categories, left(body,140) AS body_head, created_at
FROM doc_notes ORDER BY created_at DESC LIMIT 5;
```
expect ONE row: `pipeline`/`build`, categories containing `diagnosis` **and
`unconfirmed-diagnosis`** — a dead-end entry BY DESIGN for a smoke symptom
(UNVERIFIABLE), not a fault. Capture the state/log evidence inside the 3600s
reaper window or use the state-dump substitute. Subject sanity: pipeline keys
are `build|content|design|maintenance`; tool keys must match
`content_components.function` byte-for-byte (§1).
**3b CLOSED ✓ 2026-07-06.** Closure chain: banner `Subject: pipeline/build`
(same-line env prefix — the earlier lesson applied) → 3b.3-verified mapping
forwards to the child's `input_data.subject_*` → both orchestrations COMPLETED
at `complete`, empty err → diagnosis-category count 0 → **1**, and only a
subject-carrying run reaches the INSERT (subjectless runs skip — proven ×3), so
the row is `pipeline`/`build`. Evidence stamp = the SELECT in the §0 position.
Observations: run completed in ≈3–4 min (vs ≈26) — plausibly an early loop stop
+ warm `code_symbols`; unverified (optional grep for `cc61fad8` inside the
reaper window). `diagnoselogs7` does NOT contain this run (correlation absent —
a 2-iteration capture from another run; attribution rule applied).
**Stamp pasted 2026-07-07 11:41** — row `('pipeline','build')`, categories
`["diagnosis","unconfirmed-diagnosis"]`, `source='diagnosis-loop'`; the body
carries the stop reason `scope-not-narrowing` (the fast-run question answered
by the note itself). `source_agent` empty (header absent in that step context;
backlog — provenance intact via `source`). Dating correction: the positive run
completed 2026-07-07 morning; rev-20's "06-night" was the session boundary.

### Pilot PLAN (optional — UNBLOCKED NOW; tables live, no wiring needed)
> ✅ DONE 2026-07-07 12:32 — seeded for `tool-archetype-taster-quiz`
> (`has_fence=t`, 2,761 chars) via
> `drafts/pilot_PLAN_tool-archetype-taster-quiz.sql`. `EDIT:` markers are
> fill-later; no action needed.
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
> carry step-LEVEL `error_step` on several steps, and `tool-generator` on ALL
> THREE of its routed steps (observed 2026-07-07) — dormant instances of the
> documented bug (silently ignored). Do not copy that shape; correct adjacent
> instances while touching those workflows, as its own noted change (the B
> migration does this for tool-generator).
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

- **Snapshot before updating `agent_definitions`** (standing rule 2026-07-06):
  `SELECT snapshot_agent('<type>','<migration>: pre-update');` — §0-REF. The
  four pre-rule updates in this arc get retroactive snapshots + have
  per-migration rollback comments.
- **Env prefixes vs shell vars:** `VAR=x` on its own line sets a SHELL variable
  a child process never sees — use a same-line prefix (`VAR=x ./script`) or
  `export`. The 084 banner's `Subject:` line is the go/no-go tell (it now
  prints `Subject: NONE — persist_note will SKIP` when unset).
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
