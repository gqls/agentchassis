# RUNNING NOTES — Tool & Complex-Component Travelling Docs (PLAN + NOTES)

**Created:** 2026-07-04
**Last updated:** 2026-07-16 (rev 41 — P3 screenshots-on-failure BUILT (deploy-gated); idle-adapter Kafka ERROR spam fixed chassis-wide. Prior: the self-verifying loop CLOSED GREEN on a real bug. Latest session log at the BOTTOM.)
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

---

## Session log — 2026-07-04 (rev 3)

### Framing discussed: "the site-plan is the plan, the build pipeline the runbook"
Grounded in 030/016b: `site_plans` is versioned desired state
(`built_from_plan_version` stamped at deploy; reconciler `decideEmit` computes
drift) — "the plan table is ground truth; the rest is weather." Conclusions: the
site-plan is an **enforced** plan (stronger than a document); the pipeline is the
**compiled happy-path runbook**; written runbooks are the **un-compiled residue**
(exception knowledge — e.g. the `upsertPage` flip standing in for the deferred
stamp, the two-canonicalisation-surfaces split — which retires as it is compiled
into fixes/guards); **NOTES is the reasoning log** nothing machine-side captures
(`diagnose_emit` deliberately persists nothing); contracts/constitution sit above
as admission rules. Graduation rule adopted: prose → structured → enforced on
recurrence (locks deferred under this rule). Categories: (framing, decision)

### DECISION: diagnosis output → doc_notes, folded into Phase A
`persist_diagnosis_note` as a config-gated step AFTER `diagnose_emit` (emit stays
read-only per its own design comment). Writes only when the run's subject
(`function` or `pipeline`) is explicit in input — skip, don't guess. CONFIRMED →
root-cause entry; UNVERIFIABLE → still written, tagged `unconfirmed-diagnosis`
(dead ends prevent retries). Realises 037's pipeline-integration vision:
machine-written NOTES. Categories: (design, decision)

### PROPOSED: generalise tables to subjects (adopt unless vetoed)
`tool_doc_plan`/`tool_doc_notes` → `doc_plans`/`doc_notes` keyed by
`(subject_type, subject_key)`: `('tool', function)`, `('pipeline',
site_work_items.pipeline value — seen 'build'; verify enum)`. Same columns,
partial unique per subject, one write path — avoids a parallel pipeline_doc_*
later. Migration not yet written, so the rename is free now. Categories:
(schema, decision)

### Pipeline documentation designed: derive topology, author intent
A pipeline's step map is generated from `agent_definitions` (callgraph.go
pattern) — never hand-drawn, so it can't drift. Authored PLAN holds only:
**invariants** (e.g. "interactive sections survive every rebuild route" — the
unstated invariant behind de-tool), **branch rationale** (page-build-handler
re-resolves sources vs page-rebuild deliberately doesn't — previously only 016b
lore), **seams** (pipelines sharing one handler is where seam bugs live),
**deliberate decisions** (priorities, cooldowns, NULL-as-stale). Write hooks:
workflow-altering migrations append a note (005's migrations table is the
embryo); persisted diagnoses cover incidents; 016/016b stays the global roll-up.
Retrieval symmetry: pipeline scope IS code, so `docselect` + `path_globs` fits
pipelines (unlike tool functions); needs the git mirror for files (Phase B).
004–008 remain the prose base for first PLAN bodies. Categories: (design,
retrieval)

### Tool assurance designed: criteria + ladder + iteration
"Fully tested" needs a per-tool definition of *working*: the PLAN behaviour
contract written as **acceptance criteria**; multi-page tools add page set +
inter-page contract (URLs, shared state keys, feeds). Verification ladder:
1 structural (`check_tool_health`, exists) → 2 contract-presence (thin DOM/asset
assertions from criteria; catches empty-shell/detool/js-not-extracted) →
3 acceptance audit (`tool-auditor` extended to deployed-pages-vs-criteria;
failures spawn improvement items — findings `acceptance_test` pattern) →
4 headless behavioural (new infra; deferred decision). Iteration loop: deploy →
audit → failing criterion → improvement item → fixer loads PLAN+NOTES → fix →
note → re-audit; criteria hold the bar still, notes stop iterations fighting.
Checked 023: model/prompt evaluation only — no behavioural test layer exists
today, gap confirmed. Prerequisite before multi-page scaling: preserve-sections
re-render + interactivity-aware save guard (pending). It IS a documentation
challenge insofar as docs define "working" and carry iteration memory; machinery
does the checking, fed by the docs. Categories: (design, testing)

### Docs updated to rev 3
PLAN (framing, subject tables, Phase A additions 5-step, pipeline + tool
sections), RUNBOOK (procedures §4 diagnosis persist, §5 acceptance & iteration,
pipeline rules), this log. Categories: (docs)

---

## Open threads (rev 3 state)

1. Migration: `doc_plans`/`doc_notes` — confirm no live-DB collision AND the
   `site_work_items.pipeline` value set.
2. KB `tool_docs` write (019 claim) — still unverified.
3. `deploy_tool_to_site` — confirm forks stamp `source_*`; NOTES-only on fork.
4. `rag_index` `source_type` — parameterise or accept `'scrape'`.
5. Acceptance-audit ingestion of criteria (prose-via-LLM first; structure only on
   volume, per graduation rule).
6. Headless behavioural testing — future infrastructure decision.
7. Next build step: draft the migration + four actions (`write_doc_plan`,
   `append_doc_note`, direct-by-key loader, `persist_diagnosis_note`) as chassis
   drafts, then the Tier-2 contract-presence check.

---

## Session log — 2026-07-04 (rev 4)

### DECISION: acceptance criteria live in the tool's doc_plans PLAN
Judged candidates on key/lifecycle/owner. `site_specs`: right machinery, wrong
key (site-scoped; tool criteria are function-scoped; per-site copies drift) —
`direction.must_have` stays the site-scale musts home and takes rare per-site
tool parametrisation. `site_plans`/directives: wrong lifecycle and owner
(churniest artifact, superseded per re-plan, planner-owned — never store the
bar in the artifact that regenerates most; directives configure building, not
verification; table absent from the schema dump, unverified, not needed for the
decision). Findings' `acceptance_test` (verified 004 line 92: each finding
requires `current_value`, `acceptance_test`, `suggestion`, `max_fix_attempts`):
right pattern, wrong duration — the standing criteria SEED it per iteration.
doc_plans matches all three axes (function-keyed, supersede+pin, creation/human
owned, loaded at consumption moments). Format: fenced ```criteria JSON block in
the PLAN body (tool-doc-header precedent), extracted by load_doc_context;
column only on volume. Categories: (decision, criteria)

### check_tool_completeness slotted as Tier 0
Uploaded action: completion marker + balanced script/style + length at
generation/recreation time; deliberately flags-but-passes. Ladder renumbered:
Tier 0 generation-time (header gate + completeness) → 1 structural → 2
contract-presence → 3 acceptance audit → 4 headless runner. Optional wiring:
complete=false → `truncated-output` note. Categories: (reuse, testing)

### Tier 4 promoted to ACTIVE; runner plan written
Per user direction ("this is where I want to get to — iterating until it meets
acceptance criteria"). `PLAN_tool_acceptance_runner.md` created: analyser-
adapter mould (request → browser-runner adapter pod → response on caller's
topic), Playwright+Chromium image (playwright-go; chromedp fallback), profiles
desktop (1366×900) + mobile (stable device descriptor, e.g. Pixel 7 — emulation
first, real devices out of scope), criteria contract v0 (selector_exists,
no_console_errors, asset_loads, interaction fill/click/expect,
no_horizontal_overflow, page_status_ok), flow acceptance item →
load_doc_context → resolve URLs via page_components → run both profiles → pass:
acceptance-run note / fail: acceptance-fail note + improve_tool item bounded by
max_fix_attempts. Phasing P0 skeleton+boot checks desktop → P1 interpreter+
mobile → P2 interactions → P3 screenshots → P4 optional LLM-exploratory.
Categories: (design, testing, decision)

### Phase A drafts written (pending live-DB gates)
`drafts/verify_before_migration.sql` (collision check; live pipeline values —
dump shows `site_work_items.pipeline` is unconstrained text NOT NULL DEFAULT
'build', btree (pipeline,status), NO CHECK, so values are convention);
`drafts/0NN_doc_plans_and_notes.sql` (two tables, partial unique is_current,
btree + GIN jsonb_ops, comments, renumber before apply);
`drafts/write_doc_plan_action.go` (supersede tx; docResolveSubject helper);
`drafts/append_doc_note_action.go` (single INSERT; insertDocNote +
docCategoriesJSON helpers); `drafts/load_doc_context_action.go` (PLAN + latest-N
notes + composed doc_context + extractCriteriaBlock; no-plan = has_plan=false,
not an error); `drafts/persist_diagnosis_note_action.go` (after diagnose_emit;
skip-don't-guess subject gate; unconfirmed persisted as dead ends). All carry
registry.go snippets in headers; conventions matched to rag_actions /
check_tool_completeness / diagnose_* (InputSpec+init, initialize short-circuit,
datahelpers, params.DB, Headers agent_type, map returns, logger.Info).
Categories: (drafts, schema)

---

## Open threads (rev 4 state)

1. Run `drafts/verify_before_migration.sql` on live DB; paste results; renumber
   and apply the migration if clean.
2. Registry.go entries + workflow wiring migrations for the four actions
   (create_tool_component → write_doc_plan; fix agents → append_doc_note;
   diagnosis agent → persist_diagnosis_note; rag_index steps).
3. Tier-2 contract-presence check consuming criteria_json.
4. Acceptance runner P0 (adapter skeleton — image, topics, boot checks).
5. KB `tool_docs` write (019 claim) — still unverified.
6. `deploy_tool_to_site` forks: confirm source_* stamp; NOTES-only on fork.
7. `rag_index` source_type parameterisation decision.

---

## Session log — 2026-07-04 (rev 5 — rollout underway)

### Gates run on live DB; migration APPLIED
verify_before_migration.sql results: no doc_plans/doc_notes collision; pipeline
live values = build (3579), content (24), design (13), maintenance (2) — these
four are the valid pipeline subject_keys; no CHECK constraint on
site_work_items. Migration applied same session — statement tally verified
from psql output (2 CREATE TABLE, 5 CREATE INDEX, 3 COMMENT, COMMIT): both
tables and all indexes live. Categories: (schema, milestone)

### Go actions deploying; rollout tracker added to RUNBOOK (rev 5)
The four action drafts are being deployed (Stage 2). RUNBOOK now opens with a
stage tracker (Stage 0 ✅ gates, Stage 1 ✅ migration, Stage 2 ⏳ deploy,
Stage 3 ▶ persist_diagnosis_note wiring as the smoke consumer, Stage 4 wiring
for tool-generator + fix agents, Stage 5 Tier-2 check, Stage 6 Runner P0) with
per-stage SQL/bash and done-when conditions; position to be updated every turn.
Stage-3 choice rationale: additive, config-gated, end-of-workflow, triggerable
on demand; verification includes the negative case (no explicit subject → no
row). Categories: (docs, rollout)

### Tier-2 and Runner P0 defined explicitly (user request)
Tier 2 = static, browserless contract-presence: parse deployed HTML + asset
presence against criteria_json's statically visible subset; catches
markup-visible categories; can never verify behaviour (that claim belongs to
Tier 4); home = a new pass inside check_tool_health (read it first);
needs_criteria note when no PLAN criteria exist. Runner P0 = smallest
end-to-end Tier-4 slice: browser-runner-adapter deployment (Chromium +
Playwright, playwright-go), two topics in the analyser-adapter shape, three
check types desktop-only (page_status_ok, selector_exists, no_console_errors),
hand-produced request against one live tool page; exit = results match manual
inspection. Categories: (design, testing)

---

## Open threads (rev 5 state)

1. Stage 2 done-when: pods on the new binary (registry entries in place).
2. Stage 3: paste the diagnosis agent's workflow JSON (fetch SQL in RUNBOOK §0)
   → wiring migration drafted against the real nesting → one triggered run →
   verify doc_notes row + the negative case.
3. Stage 4: fetch tool-generator + fix-agent definitions; PLAN-body composition
   step needed before write_doc_plan in the generator.
4. Stage 5: read check_tool_health.go, then draft the contract-presence pass.
5. Stage 6: Runner P0 deliverables (image, Deployment, topics, consumer).
6. Still open: KB tool_docs write (019 claim); deploy_tool_to_site source_*
   stamp; rag_index source_type parameterisation.

---

## Session log — 2026-07-04 (rev 6 — actions on production)

### Stage 2 complete: four actions ON PRODUCTION
write_doc_plan, append_doc_note, load_doc_context, persist_diagnosis_note
deployed. Tracker advanced: Stage 3 (persist_diagnosis_note wiring) is YOU ARE
HERE. Per the fetch-first rule, the wiring migration waits for the diagnosis
agent's real workflow JSON (RUNBOOK §0 3.1 queries); a ready-to-place step
fragment is prepared with EMPTY config because the action's InputSpec defaults
already match emit's output_field 'diagnosis' and the input_data.* subject
fields — placement after emit, before complete_workflow. Categories: (rollout)

### Newcomer intro added to RUNBOOK (user request)
Plain-language paragraph at the top stating the task for someone new: what
travelling docs are, that agents write them as byproducts and load them before
fixes, the acceptance-criteria/tier ladder up to the headless runner, and where
the design docs live. Categories: (docs)

### Pilot PLAN path opened (unblocked now)
Tables are live, so the FIRST real tool PLAN can be seeded by plain SQL without
any workflow wiring: candidates SELECT on content_components (component_level=
'tool'), dollar-quoted INSERT with the full section skeleton including a
```criteria fence, and a fence-intact verify. Satisfies Stage 5's precondition
early; later write_doc_plan calls supersede it cleanly (source='human',
created_by='pilot'). Categories: (rollout, criteria)

---

## Open threads (rev 6 state)

1. Stage 3.1: run the two fetch queries; paste the diagnosis agent's workflow
   JSON → wiring migration drafted against the real nesting → apply → one
   triggered run with explicit subject → verify positive AND negative cases.
2. Optional pilot: run the candidates SELECT, pick a function, seed the PLAN.
3. Stage 4: tool-generator (needs the PLAN-body composition step) + fix agents.
4. Stage 5: read check_tool_health.go, then draft the contract-presence pass.
5. Stage 6: Runner P0 deliverables (image, Deployment, KafkaTopic CRs, consumer).
6. Still open: KB tool_docs write (019 claim); deploy_tool_to_site source_*
   stamp; rag_index source_type parameterisation.

---

## Session log — 2026-07-04 (rev 7 — Stage-3 fetch corrected)

### processing_mode does not exist; workflow columns + diagnosis agents resolved
The Stage-3 fetch assumed a `processing_mode` column — it does not exist
(caught on live DB). agent_definitions holds workflows in task_workflow /
orchestrator_workflow / orchestration_workflow / default_config->'workflow'
(jsonb). Diagnosis family resolved from the live type list:
diagnose-orchestrator, diagnose-agent, code-indexer — diagnose_emit runs in one
of the first two. RUNBOOK 3.1 replaced with A/B/C queries that (A) show which
column each populates, (B) locate the diagnose_emit step across all four homes,
(C) pretty-print the right one to paste back. No guessing which agent/column.
Categories: (schema, rollout)

### Write-hook + creation agent types pinned to real names
Resolved against the live list: PLAN-at-creation = tool-generator (and
component-creator for non-tool complex components later). Note-append fix agents
= component-template-fixer, tool-improver, tool-recreation-handler
(update_component_html is an ACTION, not an agent). Stage 4 updated.
Categories: (schema, design)

### Migration will target the CURRENT version row
agent_definitions is versioned (version + previous_version_id + unique
(type,version)). The wiring UPDATE targets type + max(version) + deleted_at IS
NULL, matching the house pattern — not a blind UPDATE by type. Categories:
(schema, decision)

---

## Open threads (rev 7 state)

1. Stage 3.1: run A/B/C; paste the workflow JSON of whichever diagnosis agent
   holds diagnose_emit (+ which column) → wiring migration drafted against real
   nesting, targeting the current version row → apply → one run with explicit
   subject → verify positive AND negative.
2. Optional pilot PLAN: candidates SELECT → seed one tool PLAN by SQL.
3. Stage 4: tool-generator (PLAN-body composition step) + component-template-fixer
   / tool-improver / tool-recreation-handler (append_doc_note last step).
4. Stage 5: read check_tool_health.go → contract-presence pass.
5. Stage 6: Runner P0 (image, Deployment, KafkaTopic CRs, consumer).
6. Still open: KB tool_docs write (019 claim); deploy_tool_to_site source_*
   stamp; rag_index source_type parameterisation.

---

## Session log — 2026-07-04 (rev 8 — Stage-3 migration drafted)

### diagnose_emit located: diagnose-agent, default_config workflow
A/B results: diagnose_emit runs in diagnose-agent only (in_orch true there is the
deprecated orchestrator_workflow copy; default_config is live). diagnose-
orchestrator's in_orch=true is because it NAMES the emit output when spawning the
child, not because it runs the step. Full live workflow retrieved: object-keyed
steps; emit (diagnose_emit, output_field "diagnosis", next_step "complete");
complete (complete_workflow, result_from "diagnosis"). Categories: (schema)

### Stage-3 wiring migration drafted (insert between emit and complete)
drafts/0NN_wire_persist_diagnosis_note.sql: jsonb_set adds
workflow.steps.persist_note (persist_diagnosis_note, empty config, next_step
complete) and redirects emit.next_step to persist_note. Patches default_config
(live column), leaves deprecated orchestrator_workflow alone, targets current
version row (type + max version + deleted_at IS NULL), guards exactly-one-row.
result_from "diagnosis" unaffected since persist runs before complete.
Categories: (rollout, decision)

### Caveat recorded: subject not yet threaded → first runs SKIP (by design)
diagnose-agent requires only `symptom`; subject_type/subject_key are not in its
contract and diagnose-orchestrator doesn't pass them. So persist_diagnosis_note
skips (persisted:false) until Stage 3b threads subject_type?/subject_key? through
the diagnosis input_contract + orchestrator input_mapping. Chose to wire the step
now (inert, safe) and thread the subject when the first subject-aware caller
exists (e.g. a tool audit that has the function) — NOT to infer subject from
runtime_site (that's a domain, not a function/pipeline key — would be guessing).
Categories: (design, decision)

---

## Open threads (rev 8 state)

1. Stage 3a: apply drafts/0NN_wire_persist_diagnosis_note.sql; confirm shape;
   trigger a subjectless run; verify persisted:false + no row (proves the gate).
2. Stage 3b (when wanted): thread subject_type?/subject_key? through
   diagnose-agent input_contract + diagnose-orchestrator call_diagnoser
   input_mapping; then a subject-carrying run leaves a doc_notes row.
3. Optional pilot PLAN: candidates SELECT → seed one tool PLAN by SQL.
4. Stage 4: tool-generator (add PLAN-body compose step before a write_doc_plan
   step; NOTE its save step is create_tool_component with html_content —
   compose doc_plan_body from input_data.spec + generated_html reasoning);
   append_doc_note as last step on tool-improver (after create_rerender_item),
   component-template-fixer, tool-recreation-handler (after deploy_page).
5. Stage 5: read check_tool_health.go → contract-presence pass.
6. Stage 6: Runner P0 (image, Deployment, KafkaTopic CRs, consumer).
7. Still open: KB tool_docs write (019 claim); deploy_tool_to_site source_*
   stamp; rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 9 — guideline compliance pass)

### Solution checked against 001 / 002 / 003 / 035 (user request)
Verdicts: reuse-first, orchestrator/no-subworkflow, complexity-in-Go,
logger.Info, schemas-before-SQL, prefixed InputSpec/config field names (001
§Field Name Collisions), nil-guarded initialize, version-targeted in-place
UPDATE (user-approved; matches 070–073 precedent), response-topic rule
untouched — all conform. Appendix B data-path checklist run for the wiring:
emit precedes persist; complete's result_from reads emit's output so persist
failure cannot corrupt the caller's result once error routing exists.
Categories: (compliance)

### FINDING 1 (fixed): error_step must be INSIDE step config (001 §16)
The drafted persist_note step had no error routing; and step-level error_step is
silently ignored by the coordinator (it reads step.Config["error_step"]).
Re-drafted 0NN_wire_persist_diagnosis_note.sql: config now carries
"error_step": "complete" — a doc_notes write failure degrades gracefully; the
diagnosis still reaches the caller. Also observed: tool-recreation-handler and
tool-auditor definitions carry step-LEVEL error_step on several steps — dormant
instances of the exact bug 001 documents; recorded as a Stage-4 caution (fix
when touching, don't copy). Categories: (compliance, detool-guard, decision)

### FINDING 2 (fixed): Runner P0 contract was not 035-conformant
The runner plan predated reading 035 (normative envelope): it used a body
run_id as the implied matcher and a loose response shape. Re-pinned to the
favoured git/thunder pattern: Convention-A topic
system.adapter.browser-runner.requests + browser-runner.adapter.group; action
from body.action, payload at body.data; typed header struct with real bools;
in_response_to_request_id = incoming request_id (THE matcher); request_id
reused; fresh message_id; Tier-1 echoes; status complete/error_*;
ProduceWithValidation never plain Produce; §2.15 smoke IS the P0 exit test.
§2.1 decision table confirms adapter over agent (heavy external dep, multiple
callers). PLAN_tool_acceptance_runner.md rev 2. Categories: (compliance,
testing)

### FINDING 3 (added): pure-helper tests drafted (repo convention)
No guide mandate found, but the repo tests pure parts (e.g.
diagnose_route_resolver_test.go). Drafted drafts/doc_actions_helpers_test.go:
docResolveSubject (direct/field-path/bad-type/missing), docCategoriesJSON
(direct list, field list with non-strings skipped, string coercion, absent→[]),
docCategoriesJSONFromList, extractCriteriaBlock (present/absent/unclosed/
first-only), firstNonEmptyDoc. Categories: (testing)

---

## Open threads (rev 9 state)

1. Stage 3a: RE-PULL drafts/0NN_wire_persist_diagnosis_note.sql (error_step fix),
   apply, confirm shape, subjectless run → persisted:false + no row.
2. Stage 3b (when wanted): thread subject_type?/subject_key? through
   diagnose-agent input_contract + diagnose-orchestrator input_mapping.
3. Ship drafts/doc_actions_helpers_test.go with the next chassis build.
4. Optional pilot PLAN: candidates SELECT → seed one tool PLAN by SQL.
5. Stage 4: tool-generator compose+write_doc_plan; append_doc_note last step on
   component-template-fixer / tool-improver / tool-recreation-handler — with
   error_step INSIDE config on every added step.
6. Stage 5: read check_tool_health.go → contract-presence pass.
7. Stage 6: Runner P0 per the 035-pinned contract (rev 2 plan).
8. Still open: KB tool_docs write (019 claim); deploy_tool_to_site source_*
   stamp; rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 10 — Stage 3a applied; trigger documented)

### Helper tests pass; wiring migration APPLIED and shape-verified
doc_actions_helpers_test.go: 6/6 PASS in the user's build env. Migration
applied on live (BEGIN, UPDATE 1, DO guard, COMMIT); verify query shows
emit.next_step="persist_note" and persist_note with config.error_step=
"complete", next_step="complete", output_field="diagnosis_note". The garbled
mid-file psql paste was a terminal display artifact; execution clean.
Categories: (rollout, milestone)

### Diagnosis trigger documented (user request; shape from 082/083c)
drafts/084_TRIGGER_diagnose_v1.sh created + canonical copy in RUNBOOK §3a:
kcat pod → system.agent.generic.requests, action=orchestrate, agent_type=
diagnose-orchestrator (the spawn wrapper — NOT diagnose-agent directly: an
in-place run on a shared pod lacks GITHUB_READ_TOKEN and analyse_repo_local
fails pre-fetch; same spawn-gate reasoning as 083c/index-orchestrator). REF
explicit never HEAD (2026-07-02 decision). input_data: symptom (required) +
owner/repo/ref + optional runtime_site/site_id; subject fields commented out —
they have NO effect until 3b threads them through the orchestrator's
input_mapping. Checks: spawned pod + token grep (empty-grep trap noted), log
markers through persist_diagnosis_note, orchestration_states by correlation id,
and the gate verification: skip log line + 0 diagnosis-category doc_notes rows,
decisive only with COMPLETED status (0-rows rule). Categories: (rollout, docs)

---

## Open threads (rev 10 state)

1. Stage 3a: RUN drafts/084_TRIGGER_diagnose_v1.sh → paste orchestration status,
   the skip log line, and the 0-count → 3a closes.
2. Stage 3b (when wanted): thread subject_type?/subject_key? through
   diagnose-agent input_contract + diagnose-orchestrator call_diagnoser
   input_mapping — then the SAME trigger with the subject fields uncommented
   exercises the positive persist path.
3. Optional pilot PLAN: candidates SELECT → seed one tool PLAN by SQL.
4. Stage 4: tool-generator compose+write_doc_plan; append_doc_note last step on
   component-template-fixer / tool-improver / tool-recreation-handler —
   error_step INSIDE config on every added step.
5. Stage 5: read check_tool_health.go → contract-presence pass.
6. Stage 6: Runner P0 per the 035-pinned contract.
7. Still open: KB tool_docs write (019 claim); deploy_tool_to_site source_*
   stamp; rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 11 — first 3a run failed pre-gate; structural fix)

### Incident: anchorless diagnosis dies at load_runtime (gate NOT verified)
Trigger 48f937e5 ran; child diagnose-agent-workflow-1319 failed at step
load_runtime after ~91s (analyse+lookup done): "diagnose_load_runtime: need at
least one of site_id / correlation_id / domain in collected_data". persist_note
never ran, so the 0-row doc_notes count was NOT decisive (0-rows rule applied to
ourselves) and no skip line exists — Stage 3a remains unverified. Root cause is
structural, not the smoke input: runtime evidence is an OPTIONAL bundle tier
(assemble omits it when empty) but load_runtime has no effective error routing,
making the tier mandatory in practice — a code-only diagnosis mode could not
survive. Categories: (diagnosis, gap, rollout)

### Fix drafted: config-level error_step on load_runtime (001 §16 mechanism)
drafts/0NN_diagnose_load_runtime_error_step.sql — jsonb merge of
{"error_step":"assemble"} into load_runtime.config (COALESCE guards a missing
config object; jsonb_set does not create parents), current-version targeting,
exactly-one-row guard. Success and failure converge on assemble, so anchorless
runs proceed with a code+schema bundle. Follow-up (next chassis build): soften
the action to return skipped:true for no-anchor while keeping hard errors for
real DB failures. Immediate alternative recorded: rerun with
RUNTIME_SITE=<domain>. Categories: (decision, rollout)

### Pod label key corrected: agent-type (hyphen), not agent_type
User's working command proved the label; the underscore selector matches
nothing. Fixed in 084 script (3 occurrences) and RUNBOOK §3a checks (2). Note:
082/083c example scripts carry the underscore form in their monitor echoes.
Categories: (gotcha)

### Observed: orchestrator COMPLETED while child FAILED
Child reports workflow failure with header status complete + body.status
failed; orchestrator forwards it and completes. Consumers of diagnosis results
must check body.status. Recorded as behaviour, not changed. Categories:
(observation)

---

## Open threads (rev 11 state)

1. Apply drafts/0NN_diagnose_load_runtime_error_step.sql → re-run the anchorless
   trigger → gate checks (skip log line via agent-type selector + 0-count with
   COMPLETED child) → 3a closes. (Or interim: RUNTIME_SITE rerun.)
2. Follow-up chassis change: diagnose_load_runtime no-anchor softening
   (skipped:true), ship with next build alongside doc_actions_helpers_test.go.
3. Stage 3b: thread subject_type?/subject_key? through orchestrator
   input_mapping + agent input_contract → positive persist path.
4. Optional pilot PLAN seed; Stage 4 wiring; Stage 5 check_tool_health read;
   Stage 6 Runner P0.
5. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 12 — run-2: routing fired, target wrong)

### Run-2 (correlation 9870a08b): mechanism validated, my target name wrong
Routing migration applied cleanly (BEGIN, UPDATE 1, DO, COMMIT; the COALESCE
merge preserved load_runtime's five existing config keys — correlation_id_field,
data_requests_field, domain_field, site_id_field + new error_step). Anchorless
re-run: ProcessingHistory shows error_routed "routed to assemble: step
load_runtime failed ..." — config-level error_step FIRED (001 §16 mechanism
live-validated) — then coordinator failed "step 'assemble' not found"
(routeToErrorStepOrFail → continueExecution). Gate still NOT verified;
persist_note never ran. The wrong target was mine: I verified emit/persist/
load_runtime names against live JSON but inferred the assemble step's name from
its ACTION name. Categories: (diagnosis, gotcha, rollout)

### Authoritative live step map captured (from the run dump)
analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict →
route → emit → persist_note → complete. Also: route.config = {emit_step:
"emit", gather_step: "load_runtime", max_iterations: 5} — loop-backs re-enter
load_runtime, so the corrected routing also covers loop-back failures
(converge on assemble_bundle, "continue with the evidence we have").
assemble_bundle.config.loop_scope_field = "route.scope.Symbols". Categories:
(schema, observation)

### Corrective drafted: derive error_step from the row's own next_step
drafts/0NN_fix_load_runtime_error_step_target.sql — sets
load_runtime.config.error_step to the VALUE of load_runtime.next_step read from
the same row (to_jsonb(#>> next_step)); convergence by construction, no name
guessing; DO-guard requires error_step = next_step AND that the target step
exists in the step map. Rule promoted to RUNBOOK gotchas: routing targets must
name an existing step or the coordinator fails the whole workflow. Categories:
(decision, rollout)

### Observed: second failure-envelope shape
Workflow-start failure notifies the parent with status error_unrecoverable /
code CHILD_ORCHESTRATION_FAILED (is_error true), vs run-1's step-level report
(header status complete + body.status failed). Diagnosis consumers must handle
both. Categories: (observation)

---

## Open threads (rev 12 state)

1. Apply drafts/0NN_fix_load_runtime_error_step_target.sql (verify select:
   next_step = error_step = assemble_bundle) → re-run anchorless trigger →
   gate checks (COMPLETED child + skip log line via agent-type selector +
   0 diagnosis rows) → 3a closes.
2. Follow-up chassis build: diagnose_load_runtime no-anchor softening
   (skipped:true) + ship doc_actions_helpers_test.go.
3. Stage 3b: thread subject_type?/subject_key? through orchestrator
   input_mapping + agent input_contract → positive persist path.
4. Optional pilot PLAN seed; Stage 4 wiring (error_step INSIDE config, targets
   verified against live step maps); Stage 5 check_tool_health read; Stage 6
   Runner P0.
5. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 13 — corrective applied; run-3 in flight)

### Corrective APPLIED + verified; run-3 launched with the corrected plan
0NN_fix_load_runtime_error_step_target.sql applied (BEGIN, UPDATE 1, DO,
COMMIT); verify select: next_step = error_step = assemble_bundle. Run-3
(correlation 8332c2a5) spawned pod …135b53b4… with GITHUB_READ_TOKEN present;
the child's executing WorkflowPlan (workflow-1407 state dump) confirms
load_runtime.config.error_step = "assemble_bundle" — the corrective is live in
the running plan, not just the definition row. At capture the run was
mid-flight at lookup_symbols (~66s), no errors. Outcome pending; gate NOT yet
claimed. Categories: (rollout)

### Log-attribution caution recorded (multi-pod selector residue)
logs -l agent-type=diagnose-agent spans all live diagnoser pods (three running:
48m/29m/21s; idle timeout 3600s reaps them), so run-2's failure dump and
"step 'assemble' not found" lines appear in run-3's capture window. Attribute
by orchestration id / pod name / timestamp before reading a line as current —
the "plan is ground truth, the rest is weather" trap, log edition. Categories:
(gotcha, observation)

### Loop files reviewed; error_step loop corollary recorded
loop_expansion_handler.go / loop_error_handler.go are NOT implicated in run-3
(the diagnosis loop-back is route's next_step override, not loop expansion).
Their relevance: per 001 Appendix C + the expansion handler, error_step/
then_step/fallback_step values INSIDE loop substeps are iteration-prefixed at
expansion, so they must name substeps of the same loop, never top-level steps;
continue_on_error:true is the iteration-scoped alternative (loop_error_handler:
record error, advance iteration). Recorded as a RUNBOOK gotcha — binds on
Stage-4 note-append steps placed inside fix-agent loop bodies. Categories:
(gotcha, design)

---

## Open threads (rev 13 state)

1. Run-3 completion → gate checks: COMPLETED child by correlation 8332c2a5 +
   persist skip line on pod …135b53b4… + doc_notes diagnosis 0-count → 3a
   closes.
2. Stage 3b: thread subject_type?/subject_key? through orchestrator
   input_mapping + agent input_contract → positive persist path.
3. Follow-up chassis build: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
4. Pilot PLAN seed; Stage 4 wiring (error_step in config; loop-substep
   corollary; targets verified against live step maps); Stage 5
   check_tool_health read; Stage 6 Runner P0.
5. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 14 — run-3 progress; 016b updated)

### Run-3: corrective fired ×5, no failures — deep in the loop
diagnoselogs4: five error_routed → assemble_bundle events, matching the loop
design (route.gather_step re-enters load_runtime every iteration; anchorless →
per-iteration degrade to a code+schema bundle). Status paste: orchestrator
AWAITING_RESPONSES at call_diagnoser; child EXECUTING_STEP at assemble_bundle,
no error. Reading rule recorded: five error_routed lines are NORMAL for
anchorless runs. Cost note: full ≤5-iteration loop (one Sonnet verdict each)
before emit. Gate checks still pending natural completion. Categories:
(rollout, observation)

### Debugging guide updated (user request): 016b → internal v5
Copied to outputs as 016b_debugging_guide_7_3_.md (filename incremented per
their scheme; internal version line v5 2026-07-06). Four §9 entries in house
shape (Symptom/Mechanism/Fix/Cross-refs + tags): error_step placement +
existing-target + derive-from-next_step (with jsonb-parents COALESCE, loop-
substep prefix corollary, dormant step-level instances in
tool-recreation-handler/tool-auditor); anchorless diagnosis dies at
load_runtime (optional tier made mandatory; ×5 live validation; 0-rows gate
edition); agent-type hyphen label + multi-pod log residue; two failure
envelopes (parent COMPLETED ≠ child success; check body.status). Rollout state
deliberately left in RUNBOOK (guide cross-refs it). Categories: (docs)

---

## Open threads (rev 14 state)

1. Run-3 completion → gate checks (COMPLETED child + skip line on pod
   …135b53b4… + doc_notes diagnosis 0-count) → 3a closes.
2. Stage 3b subject threading → positive persist path (first machine-written
   NOTES row).
3. Follow-up chassis build: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
4. Pilot PLAN seed; Stage 4 wiring; Stage 5 check_tool_health read; Stage 6
   Runner P0.
5. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 15 — runbook tidy; 3a status made explicit)

### 3a status answered explicitly (user question)
Trigger RUN: yes — three attempts; run-3 (corr 8332c2a5) is the valid one.
Verification: NO — diagnoselogs4 was a mid-loop capture (child EXECUTING at
assemble_bundle), and the gate's pass condition (COMPLETED child + persist skip
log line + 0-count made decisive by both) cannot exist in an in-flight capture
by definition. Likely finished since; the CLOSE-OUT block settles it in three
pastes. Categories: (rollout, decision)

### Runbook tidied (rev 15)
Stage 3 rewritten: one Status line (✓/✓/✓/PENDING), runs 1–2 compressed to
pointers (full narratives live in RUNNING_NOTES + 016b v5 §9 — runbook no
longer retells them), a single consolidated "3a CLOSE-OUT" block (state query
with reading rules incl. the ≫1800s stall threshold, pod-scoped gate grep, DB
gate, optional loop-conclusion grep), canonical-trigger pointer, 3b spelled
out. NEW §0-REF added under the position line: copy-paste state-check queries
(orchestration by correlation — never created_at; children by
parent_orchestration_id; doc_plans/doc_notes/roll-up inspections; pod selector
with hyphen + attribution caution) + the reading rules (parent COMPLETED ≠
child success; non-empty parent error = forwarded child failure; 0-rows rule).
Categories: (docs)

---

## Open threads (rev 15 state)

1. Run the §3a CLOSE-OUT block (status → skip line → 0-count) → paste → 3a
   closes; position line updates.
2. 3b subject threading (two migrations, fetch-first) → positive persist run →
   first machine-written NOTES row.
3. Follow-up chassis build: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
4. Pilot PLAN seed; Stage 4 wiring; Stage 5 check_tool_health read; Stage 6
   Runner P0.
5. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 16 — STAGE 3a CLOSED)

### Gate verified in state form; 3a CLOSED
Evidence trio: (1) both orchestrations COMPLETED at complete, empty err
(since_s ≈4756/4769); (2) diagnoselogs6 is a post-completion STATE DUMP whose
ProcessingHistory shows route 14:30:09 → emit 14:30–14:31 → persist_note
14:31:37–14:33:02 EXECUTED → complete 14:33:09–14:33:32; (3) doc_notes
diagnosis-category count = 0. The literal skip stdout line was not capturable
(pod past the 3600s idle window); skip-vs-error settled structurally — the
subject gate is the action's FIRST check, before any DB access, and no error
appears anywhere. Token caution honoured: UNVERIFIABLE/CONFIRMED counts in the
dump are the embedded verdict-prompt text, not the run outcome; the loop's
actual conclusion is not extractable and is immaterial to the gate. Observed
cost of a full anchorless run: ≈26 min, 5 iterations, 5 normal degrades.
Categories: (milestone, rollout)

### Lesson banked: gate-evidence capture window + the state-dump substitute
Runbook close-out block and 016b v5 anchorless entry now carry: the skip log
line has a 3600s capture window; past it, the post-completion state dump
(step-executed history + terminal status + 0-count) is the accepted
substitute. Categories: (gotcha, docs)

### NEXT: 3b.1 fetch queries live in the runbook
call_diagnoser step JSON + both input_contracts; on paste, the two wiring
migrations (input_mapping += subject_type?/subject_key?; input_contract
optional += the two keys) get drafted with targets verified per the routing
rule. Then the SAME trigger with subject fields uncommented (e.g.
pipeline/build) → the first machine-written NOTES row. Categories: (rollout)

---

## Open threads (rev 16 state)

1. 3b.1: run the two fetch queries (runbook §3b.1) → paste → 3b migrations
   drafted → apply → subject-carrying run → verify the first doc_notes row.
2. Follow-up chassis build: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
3. Pilot PLAN seed; Stage 4 wiring; Stage 5 check_tool_health read; Stage 6
   Runner P0.
4. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 17 — 3b drafted against the pasted definitions)

### 3b.1 facts recorded
call_diagnoser.input_mapping = nine keys (ref?/repo?/owner?/symptom/site_id?/
seed_scope?/runtime_page?/runtime_site?/correlation_id?) — no subject fields;
input_contracts of diagnose-agent AND diagnose-orchestrator are identical twins
(required ["symptom"], eight optionals). Confirms the rev-8 inference exactly.
Categories: (schema)

### DECISION: three edits, not two (016b spawn+call contract rule)
Because input_mapping must satisfy the callee's input_contract, the callee
(diagnose-agent) MUST declare the subject optionals; the orchestrator's own
contract gets them too so the trigger's input is a declared interface. Drafted
drafts/0NN_wire_diagnosis_subject_threading.sql: mapping object-merge (preserves
the nine keys; "?"-suffix optional convention kept), optional-array append via
DISTINCT aggregate (idempotent; order not significant), current-version
targeting, guards on exact mapping paths + both contracts. DB-only — effective
immediately, no deploy. Categories: (design, decision, rollout)

### 084 trigger: SUBJECT_TYPE/SUBJECT_KEY env support
Commented hint replaced with real env-driven append into input_data; banner
line added. Positive run = SUBJECT_TYPE=pipeline SUBJECT_KEY=build ./084 "...".
Expectation set in RUNBOOK 3b.4: ONE doc_notes row pipeline/build with
categories diagnosis + unconfirmed-diagnosis — a dead-end entry BY DESIGN for a
smoke symptom; ~26 min; degrade ×5 still normal; evidence within the reaper
window or via the state-dump substitute. Categories: (rollout, docs)

---

## Open threads (rev 17 state)

1. Apply 0NN_wire_diagnosis_subject_threading.sql → 3b.3 verify → 3b.4 positive
   run → paste the doc_notes row → 3b CLOSES (travelling docs live end-to-end
   for diagnoses).
2. Follow-up chassis build: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
3. Pilot PLAN seed; Stage 4 wiring (write_doc_plan into tool-generator;
   append_doc_note into the three fix agents); Stage 5 check_tool_health read;
   Stage 6 Runner P0.
4. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 18 — snapshot rule adopted; 3b applied; interim run flagged)

### NEW STANDING RULE: snapshot agents before updating
Adopted from the constitution addition. Consequence owned: the four
agent_definitions updates this arc (persist_note wiring; load_runtime
error_step; error_step target fix; 3b subject threading) predate the rule —
restore for those was rollback-comments-only. Remedy:
drafts/0NN_agent_definition_snapshots.sql — a SIDE table
(CREATE TABLE ... AS TABLE agent_definitions WITH NO DATA + snapshot_taken_at/
snapshot_reason + index), a retroactive snapshot of both touched agents'
CURRENT (post-3b) rows, a guard, the standing pre-update INSERT snippet, and a
restore example. Side table over in-table is_snapshot rows because UNIQUE
(type,version) + the max(version) current-row selectors would make an in-table
snapshot hijack "current" unless every selector grew an is_snapshot filter;
the file includes an inspection query to detect the 009/024b is_snapshot
convention — converge on it if it turns out to be live. SELECT ad.* is
positional: extend the snapshot table FIRST if agent_definitions ever gains
columns. Categories: (decision, rollout, gotcha)

### 3b.2 APPLIED (pre-rule); 3b.3 NOT yet run
Migration applied cleanly (UPDATE 1 ×2, guard DO, COMMIT — psql paste garbling
again display-only). Verify selects still pending — sequenced after the
snapshot apply. Categories: (rollout)

### Interim run 03bebec9 is SUBJECTLESS — expected to SKIP
Banner evidence: no Subject: line (the 084 banner prints one when
SUBJECT_TYPE is set) and the 3a default symptom text. Child healthy at
assemble_bundle 16:20:42 (degrade fired). It serves as a free post-3b
regression of the skip gate; the 3b.4 positive test still requires
SUBJECT_TYPE=pipeline SUBJECT_KEY=build. Categories: (observation, rollout)

### uuid placeholder paste artifact
Query 1 failed on '<03bebec9-…' — the placeholder's angle bracket survived
inside the quotes. §0-REF now says: replace placeholders INCLUDING the
brackets. Categories: (gotcha)

---

## Open threads (rev 18 state)

1. Apply drafts/0NN_agent_definition_snapshots.sql (standing rule; retroactive
   snapshot; run its is_snapshot inspection query once and report).
2. 3b.3 verify selects → paste.
3. 3b.4 positive run WITH env vars → paste the doc_notes row → 3b CLOSES.
4. (Optional) let 03bebec9 finish as the post-3b skip regression.
5. Follow-up chassis build: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
6. Pilot PLAN seed; Stage 4 wiring (snapshot INSERT prepended to every
   migration from now on); Stage 5 check_tool_health read; Stage 6 Runner P0.
7. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 19 — snapshot_agent surfaced; 3b.3 green; env-var lesson)

### Reuse win: snapshot_agent() already exists — side-table draft SUPERSEDED un-applied
The platform has snapshot_agent(p_agent_type text, p_reason text DEFAULT NULL)
RETURNS uuid (two overloads). It is a FUNCTION, which is why table-level schema
checks never surfaced it — lesson: fetch-first applies to functions too (\df on
the verb) before drafting snapshot/backup machinery. Standing rule updated to
SELECT snapshot_agent('<type>','<migration>: pre-update'); side-table migration
marked SUPERSEDED, DO NOT APPLY (drop-if-applied one-liner included). Pending:
paste of the post-call inspect (or \sf) to learn WHERE it writes — until then,
future current-row selectors defensively add AND is_snapshot IS NOT TRUE
(harmless if snapshots live elsewhere; load-bearing if they are higher-version
in-table rows). Categories: (reuse, decision, gotcha)

### 3b.3 VERIFIED
map paths exact (input_data.subject_type / input_data.subject_key); both
contracts t/t. Definition-level threading confirmed. Categories: (rollout)

### 3b.4 attempt-1 (324456a9) subjectless — env vars set but not exported
echo showed the vars in the interactive shell, but bare VAR=x on its own line
never reaches a child process; banner tell: no Subject: line + default 3a
symptom. The 0-rows doc_notes at 18:35 was doubly expected (no subject; likely
mid-flight). Fix: same-line prefix or export. 084 hardened: prints an explicit
"Subject: NONE — persist_note will SKIP" when unset. Runs 03bebec9/324456a9
stand as post-3b skip regressions. Categories: (gotcha, rollout)

---

## Open threads (rev 19 state)

1. Run the two snapshot_agent() calls + the inspect query → paste (reveals the
   snapshot convention; decides whether the is_snapshot selector predicate is
   load-bearing).
2. 3b.4 positive run with SAME-LINE prefix (banner must show Subject:) → paste
   the doc_notes row → 3b CLOSES.
3. Follow-up chassis build: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
4. Pilot PLAN seed; Stage 4 wiring (snapshot_agent call + is_snapshot-safe
   selectors in every migration); Stage 5 check_tool_health read; Stage 6
   Runner P0.
5. Still open: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-06 (rev 20 — STAGE 3 CLOSED: first machine-written NOTES row)

### MILESTONE: 3b closed; travelling docs live end-to-end for diagnoses
Run cc61fad8 (banner: Subject pipeline/build — same-line prefix fix applied):
both orchestrations COMPLETED at complete, empty err; doc_notes diagnosis
count 0 → 1. Attribution is structural: only a subject-carrying run reaches
the INSERT (subjectless skip proven ×3 in 3a and the two regressions), so the
row is ('pipeline','build') — the build pipeline's NOTES stream opens with a
recorded diagnosis dead-end, which is exactly the reasoning-history content
037 identified as captured nowhere. The 037 pipeline-integration vision has
its first automated writer. Evidence stamp SELECT queued (also the first read
of the note body). Categories: (milestone, rollout)

### Log attribution honoured: diagnoselogs7 is NOT the positive run
Correlation cc61fad8 appears 0 times in the upload; the file is a 2-iteration
capture from another run (persist_note match = plan-definition text). Closure
therefore rests on the structural chain above, not on a log line. Fast-run
observation: ≈3–4 min vs ≈26 — plausibly early loop stop on a nonsense symptom
+ warm code_symbols lookups; unverified, optional pod grep within the reaper
window if the stop reason should go on record. Categories: (observation)

### Housekeeping outstanding (before any Stage-4 migration)
Retroactive snapshot_agent('diagnose-agent'|'diagnose-orchestrator', ...) ×2
+ the inspect query — not yet run (user proceeded to the positive run first;
no breach — no agent updates occurred). The inspect decides whether the
is_snapshot selector predicate is load-bearing. Categories: (rollout)

---

## Open threads (rev 20 state)

1. Paste the evidence-stamp SELECT (formality + first read of the note).
2. Retroactive snapshot_agent calls ×2 + inspect paste — REQUIRED before any
   Stage-4 agent migration.
3. Fork: pilot PLAN seed (minutes; unblocks Stage-5 criteria) OR Stage 4 first
   wiring (tool-generator → compose + write_doc_plan; fetch definitions first;
   snapshot + is_snapshot-safe selector).
4. Chassis build follow-ups: diagnose_load_runtime no-anchor softening +
   doc_actions_helpers_test.go.
5. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-07 (rev 21 — stamp confirmed; tasks made plain)

### Evidence stamp pasted: the first machine-written note, read in full
Row: ('pipeline','build'), site_id NULL, categories ["diagnosis",
"unconfirmed-diagnosis"], source='diagnosis-loop', created_at 2026-07-07
11:41:35. Body heading dated 2026-07-07; "Root cause: unconfirmed
(scope-not-narrowing)" — the loop's stop reason, ON RECORD IN THE NOTE ITSELF,
answering the prior "why so fast" observation without a log grep. First payoff
of the feature: the machine-written history resolved our own open question.
Dating correction: the positive run completed 07-07 morning (rev-20 said
"06-night" — session boundary). dioagnoselogs8 contains cc61fad8 only as plan
dumps (no runtime persisted line) — DB row is primary evidence, sufficient.
Categories: (milestone, observation)

### Observation + backlog: source_agent empty on the persisted row
The action reads params.Headers["agent_type"], which is absent in that step's
execution context; provenance intact via source='diagnosis-loop'. Backlog for
the next chassis build: fallback population (ExecutionContext / agent config)
in the doc actions. Categories: (gotcha)

### Tasks rewritten in plain form (user request)
Runbook §0 now carries "YOUR TASKS — plain version": TASK 1 snapshots (what/
why/how/done-when, ≈2 min, before any Stage-4 migration); TASK 2 the fork with
both options spelled out (A pilot PLAN seed — candidates SELECT then pick a
function; B tool-generator write hook — workflow fetch then migration), the
recommendation A-then-B, and a separate background list needing no action.
Categories: (docs)

---

## Open threads (rev 21 state)

1. TASK 1: two snapshot_agent calls + inspect → paste (decides the is_snapshot
   selector predicate).
2. TASK 2: reply with A (recommended first) or B — A needs the candidates
   SELECT output + your pick; B needs the tool-generator workflow paste.
3. Next chassis build: doc_actions_helpers_test.go + diagnose_load_runtime
   no-anchor softening + source_agent fallback in the doc actions.
4. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation.

---

## Session log — 2026-07-07 (rev 22 — Task 1 done; pilot delivered; B input banked)

### TASK 1 DONE: snapshot convention learned empirically
snapshot_agent fired for both agents (NOTICEs with source_version=1; returns
the SOURCE row id). The inspect (no deleted_at filter) shows agent_definitions
still holds ONLY the two original rows → snapshots are stored in a SEPARATE
table. The is_snapshot current-row selector predicate is therefore NOT
load-bearing — dropped from the migration mandate (harmless if present).
Optional later: \sf for the storage table; \df *agent* for the companion
restore function. Categories: (schema, decision)

### Observation: duplicate function rows in content_components (tools)
tool-archetype-taster-quiz ×2, tool-password-entropy ×5, roi-estimator ×3,
llm-cost-calculator ×3 — so idx_cc_tool_function_unique must be partial beyond
component_level='tool' (most plausibly library vs per-site instance rows from
forking). No impact on docs — they key by function, spanning all instances,
which this vindicates. Optional check: SELECT indexdef FROM pg_indexes WHERE
indexname='idx_cc_tool_function_unique'. Categories: (schema, observation)

### TASK 2A delivered: pilot PLAN file for tool-archetype-taster-quiz
drafts/pilot_PLAN_tool-archetype-taster-quiz.sql — concrete: .tool-container
boot check (generator invariant), asset_loads for the Path-1 extracted JS
(with a VERIFY caveat — the tool predates parts of the pipeline), console/
status/mobile-overflow checks, pairing with archetype-result-card, is_active=f
status note, and one seeded deliberate decision (exactly THREE questions — the
taster must not be "improved" into the Gauntlet). Unknowns honest as EDIT:
markers; the interaction check uses placeholder selectors flagged via the
criteria block's "note" + -EDIT id suffix so checkers skip it. Categories:
(criteria, rollout)

### B input banked; tool-generator carries the dormant 001 §16 bug ×3
User pasted the tool-generator workflow while choosing A. Findings: steps
ensure_site_record → load_brand_context (read_site_spec) → generate_tool_html
(execute_llm_prompt; prompt embeds the 019 tool-doc header sentinels) →
save_tool (create_tool_component) → complete; ALL THREE routed steps carry
error_step at STEP level (inert). B migration plan parked: snapshot first;
insert compose_plan (LLM drafts PLAN body incl. criteria from input_data.spec
+ generated_html) → write_doc_plan (subject_key_field input_data.spec.function;
config.error_step=complete — docs never fail tool creation) → complete; move
the three error_steps into config as its own noted change. 016b dormant-
instances sentence extended with tool-generator. Categories: (design, gotcha)

---

## Open threads (rev 22 state)

1. TASK 2A: run drafts/pilot_PLAN_tool-archetype-taster-quiz.sql (+ optional
   EDIT fills) → paste the fence verify → pilot closes; Stage-5 precondition
   satisfied.
2. TASK 3: say "go B" → snapshot-prefixed tool-generator migration drafted
   (compose_plan + write_doc_plan + the ×3 error_step corrections).
3. Next chassis build: helper tests + load_runtime softening + source_agent
   fallback.
4. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; optional index-def check for the duplicate-function
   observation; optional \sf snapshot_agent.

---

## Session log — 2026-07-07 (rev 23 — pilot seeded; jargon clarified)

### TASK 2A DONE: first tool PLAN live
INSERT 0 1; verify: tool-archetype-taster-quiz, is_current=t, has_fence=t,
2,761 chars, 12:32:49. Stage-5 precondition (≥1 tool PLAN with criteria) met.
Categories: (milestone, criteria)

### Duplicates observation CLOSED by the real index predicate
idx_cc_tool_function_unique = UNIQUE (function) WHERE component_level='tool'
AND forked_from IS NULL AND is_active=true — uniqueness covers ACTIVE LIBRARY
ORIGINALS only; extra rows are forks (forked_from set) and inactive versions.
Vindicates function-keyed docs. Side fact: tool-archetype-taster-quiz has NO
active row anywhere (fully dormant) — product fact, not a blocker.
Categories: (schema, observation)

### Jargon clarified on request; mini-glossary added to RUNBOOK
EDIT: markers = optional fill-later blanks (doc valid as seeded; fills arrive
by supersede; -EDIT check ids skipped by verification). "go B" retired —
replaced by the explicit approval phrase "yes, draft the tool-generator
migration"; Task 3 reworded in plain terms (what it automates; the user's 3
steps; what the file contains). correct-while-touching defined and attributed
honestly as a norm adopted IN THIS CHAT (2026-07-06), now in runbook/016b:
when a migration already modifies a workflow, fix known-inert bugs in that
same workflow, declared explicitly — bounded repair, no separate campaign.
Categories: (docs, decision)

---

## Open threads (rev 23 state)

1. TASK 3: awaiting the explicit "yes, draft the tool-generator migration" →
   snapshot-prefixed migration (compose_plan + write_doc_plan + the ×3
   error_step corrections) → apply → optional test-tool proof.
2. Later fills for the pilot PLAN's EDIT markers (by supersede, produced when
   the PLAN is next loaded with the component html).
3. Next chassis build: helper tests + load_runtime softening + source_agent
   fallback.
4. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; optional \sf snapshot_agent (storage table) and
   \df *agent* (restore companion).

---

## Session log — 2026-07-07 (rev 24 — Task 3 approved; migration drafted)

### tool-generator PLAN-writing migration drafted (Task 3 / Stage-4 first hook)
drafts/0NN_tool_generator_plan_writing.sql. Placement: PLAN written AFTER
save_tool succeeds (no component => no PLAN; failed save still routes to
complete_error untouched). Chain: save_tool -> compose_plan (execute_llm_prompt,
Sonnet 4.6, max_tokens 4000, inputs input_data/site_record/generated_html;
strict template — five standard checks verbatim, interaction check ONLY from
real selectors copied out of the generated HTML, never invented; <=3000 chars;
apostrophe-free template, criteria fence embedded) -> write_plan
(write_doc_plan; subject_key_field input_data.spec.function; plan_body_field
doc_plan.result; source/created_by tool-generator) -> index_plan (rag_index,
collection tool_docs; source_type='scrape' accepted open item) -> complete.
Every new step config.error_step="complete" — docs can never fail creation.
Declared deliberate changes: save_tool.next_step complete->compose_plan;
workflow.timeout_seconds 300->480 (second Sonnet call). Correct-while-touching
executed: the three inert step-LEVEL error_steps (save_tool,
generate_tool_html, load_brand_context) set in config with ORIGINAL targets
AND the dead step-level keys deleted. Snapshot embedded as the tx's first
statement (MVCC captures pre-update state; rollback removes snapshot too).
Guard asserts the full shape incl. deletions + timeout. JSON blocks
machine-validated (all three parse; prompt 2,3xx chars, fence present, no
apostrophes). Proof recipe: organic creation or manual orchestrate against a
TEST site only (real component/page/nav side effects). Categories: (rollout,
design, decision)

---

## Open threads (rev 24 state)

1. Apply drafts/0NN_tool_generator_plan_writing.sql (snapshot embedded) →
   paste the verify (save_next=compose_plan, subject_key_field, timeout=480,
   step_level_removed=t).
2. Proof: next tool creation leaves a doc_plans row (source='tool-generator');
   manual test-site trigger scripted on request.
3. Remaining Stage-4 hooks: append_doc_note as last step on
   component-template-fixer / tool-improver / tool-recreation-handler (same
   fetch-first + snapshot + correct-while-touching discipline).
4. Later fills for the pilot PLAN's EDIT markers (by supersede).
5. Next chassis build: helper tests + load_runtime softening + source_agent
   fallback.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; optional \sf snapshot_agent / \df *agent*.

---

## Session log — 2026-07-07 (rev 25 — Task 3 applied; creation-side hook LIVE)

### TASK 3 APPLIED + SHAPE-VERIFIED: tools write their PLAN at birth
Snapshot captured (type=tool-generator, source_id 1bca62f6, embedded as the
tx's first statement); UPDATE 1 ×3; guard DO; COMMIT. Verify quadruple green:
save_next=compose_plan · subject_key_field=input_data.spec.function ·
timeout=480 · step_level_removed=t (the three inert step-level error_steps are
gone; config-level targets set). Creation-side write hook is LIVE. Proof
pending: next tool creation → doc_plans row (source='tool-generator');
one-time composer acceptance review planned when the first auto-PLAN body is
pasted (five standard checks verbatim; selectors copied-or-omitted; fence
parses; ≤3,000 chars). Categories: (milestone, rollout)

### TASK 4 opened: fix-agents NOTES hook (fetch queued)
One query fetches all three workflows (tool-improver,
component-template-fixer, tool-recreation-handler). Design sketch parked for
the drafting turn: compose_note (LLM drafts the uniform entry from each
workflow's real collected data) + append_doc_note as the tail; config.error_step
to each terminal step; tag strategy per agent decided against the fetched
JSON; loop-substep corollary checked; correct-while-touching for
tool-recreation-handler's step-level error_steps (tool-auditor deferred to
Stage 5). One migration, three snapshot calls. Categories: (design, rollout)

### Constitution re-paste note
This turn's constitution paste omits the "snapshot before updating" line
(present in the 2026-07-06 paste) — treated as a stale copy, NOT a retraction:
the rule remains in force and is embedded in migration practice.
Categories: (observation)

---

## Open threads (rev 25 state)

1. TASK 3 proof: next tool creation → paste the doc_plans row + the first
   auto-PLAN body for the one-time composer review. (Manual test-site trigger
   scripted on request — real component/page/nav side effects.)
2. TASK 4: run the three-workflow fetch → paste → one snapshot-prefixed
   migration drafted (compose_note + append_doc_note tails ×3).
3. Later fills for the pilot PLAN's EDIT markers (by supersede).
4. Next chassis build: helper tests + load_runtime softening + source_agent
   fallback.
5. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; optional \sf snapshot_agent / \df *agent*.

---

## Session log — 2026-07-07 (rev 26 — Task 4 drafted; gamesdesign dogfood opened)

### DECISION: gamesdesign.co.uk as the dogfood site (over a new domain)
Already in pipelines with specs; live tools; a real broken component. A new
domain tests intake, not docs. Categories: (decision)

### Economy-simulator root cause (from index_6_.html — two compounding defects)
LOGIC: tick uses the raw slider index as the rate — popInflux =
parseInt(el.slPop.value) with slider min=0 max=5; labels
None/Slow/Medium/Fast/Surge/Flood are display-only. Flood = +5 players/tick.
Screenshot arithmetic confirms live behaviour: 191 = 50 + 5×~28 ticks; gold
generation scales with players (line 928) so figures DO move, ~20× too weakly.
DISPLAY: Players dataset yAxisID 'yGold' (right axis autoscaling to ~100k) —
~200 players renders flat on the baseline ("green line stays on 0"). Fix
shape: index→rate map (e.g. [0,1,5,15,40,100]) + own scale for Players. The
bug report IS the Tier-4 criterion: influx max → players figure + series rise
within N ticks. Detection honesty: the chassis diagnosis loop is the wrong
instrument for in-page JS; the Tier-4 runner is the layer that would catch
this. Categories: (diagnosis, criteria)

### Task 4 migration drafted (fix-agents note writing)
drafts/0NN_fix_agents_note_writing.sql. Grounded in the fetched 403-line
workflows. Success-path tails only; both fixer branches covered
(create_rerender.next_step AND check_needs_rerender.else_step →
compose_note). Subjects: fixer pipeline/build + note_site_id; improver
tool/tool_data.function; recreation tool/input_data.spec.function with
skip-on-absence via error containment (refinement on record: stamp function
into recreation item specs at creation). Machine categories v1 ["fix"];
failure-class tags in the body Categories line; entry heading carries no date
(created_at is the timestamp — deviation from persist_diagnosis_note,
accepted). Timeouts declared: fixer 120→240, improver 300→480, recreation
2400 kept. Correct-while-touching: ALL TEN recreation step-level error_steps
→ config with original targets + dead keys deleted; guard asserts
NOT EXISTS any step-level error_step across all three agents. Six JSON blocks
machine-validated (parse clean; prompts apostrophe-free). Observation banked:
recreate_tool runs Opus 4.6 at 64k tokens. Categories: (rollout, design,
decision)

---

## Open threads (rev 26 state)

1. Apply 0NN_fix_agents_note_writing.sql → paste verify (3 rows: has_compose/
   has_append true; timeouts 240/480/2400; no_step_level_error true).
2. Run the gamesdesign component lookup (runbook TASK 5) → paste → I script
   the improve_tool item (game fix; first auto fix-NOTES) + the tool-generator
   trigger (Task-3 proof, new tool on gamesdesign).
3. Task 3 proof + first auto-PLAN composer review (arrives with the trigger).
4. Pilot PLAN EDIT fills (by supersede) — the quiz; later the game's PLAN
   gains the influx criterion.
5. Chassis build: helper tests + load_runtime softening + source_agent
   fallback.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function.

---

## Session log — 2026-07-07 (rev 27 — Task 4 applied; discovery opened; 085 ready)

### TASK 4 APPLIED: fix agents write NOTES
Three snapshots (076455bf fixer, 1f3ebb4a improver, 8701375f recreation);
UPDATE 1 ×5; guard DO passed (the guard asserts the full shape incl. zero
step-level error_steps across all three agents — it IS the verification);
COMMIT. Tail verify = optional stamp. Categories: (milestone, rollout)

### Lookup surprise: the game body has no component address (yet)
The economy-simulator PAGE exists (game-economy-simulator, site e33263f4) but
its only page-linked components are a shared hero (same component_id as the
guide page) + a guide text block. tool-improver.load_tool requires a
component_id via the page_components join — the fix path is BLOCKED pending
discovery. Hypotheses (unconfirmed, in order): unlinked/differently-levelled
content_components row; sections written by save_page_sections (recreation
path) rather than a single tool component; body only in research_results +
deployed repo. Discovery queries queued in RUNBOOK TASK 5 (page-scoped no-name
filter; sitewide incl. html_template match on the game H1; \dt *section* if
both empty). Don't-jump honoured: no improve item scripted until the body has
an address. Categories: (observation, gotcha)

### 085 trigger drafted (Task-3 proof on gamesdesign)
drafts/085_TRIGGER_toolgen_gamesdesign_v1.sh — envelope copied from 084
verbatim (headers + body shape), target tool-generator, input_data carries
BOTH domain and site_id (e33263f4) plus spec
{function tool-xp-curve-designer, name, description, complexity simple};
banner declares the REAL side effects; post-run checks embedded (status,
doc_plans proof, component/page selects, PLAN-body paste for the one-time
composer review). bash -n clean. Categories: (rollout)

---

## Open threads (rev 27 state)

1. TASK 5 discovery: run Q1 + Q2 (and \dt *section* if both empty) → paste →
   improve_tool item scripted once the game body has an address.
2. TASK 3 proof: run 085 when ready → paste status + doc_plans row + PLAN body
   (composer review).
3. First fix-NOTES entry arrives with whichever fix runs first (categories ?
   'fix').
4. Chassis build: helper tests + load_runtime softening + source_agent
   fallback; recreation items to carry spec.function.
5. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type.

---

## Session log — 2026-07-08 (rev 28 — proof run found an upstream defect; game reframed)

### Task-3 proof run: generate_tool_html -> complete_error (0 PLANs = correct)
Run 00688389 (tool-xp-curve-designer, gamesdesign): EXECUTING at
generate_tool_html for ~57s then COMPLETED at complete_error (~96s). Read
correctly: complete_error bypasses compose_plan, so 0 doc_plans rows is the
CORRECT outcome — no content_components row, no new page (the three pages shown
are pre-existing, created_at order). NOT a Task-3 regression; the PLAN steps
are downstream of a save that never happened. The corrected error_step (moved
into config by Task 3) fired as designed and caught a real generation failure
underneath. err column empty at each poll; the failure detail is in
collected_data (complete_error emits generated_html) and the pod log. Two reads
queued in RUNBOOK CURRENT BLOCKER: collected_data by correlation; tool-generator
pod log grepped by id. Don't-jump + fetch-first honoured — no fix guessed.
Categories: (incident, rollout)

### Schema correction: content_components has NO site_id column
Q2 errored (column site_id does not exist). Per \d: components scope to sites
via page_components / site_components only. Q1 (page join) is authoritative.
Banked: created_from CHECK {manual,generated,adopted,tool,forked};
component_level default 'section'; name UNIQUE; idx_cc_tool_function_unique
confirmed (function WHERE component_level='tool' AND forked_from IS NULL AND
is_active). Categories: (schema)

### Game reframed: recreation case, not improve case
Q1 conclusive: the economy-simulator page has only a shared hero section
(2,790 chars) linked — no game-body component. tool-improver has nothing to
load. Correct path: tool-recreation-handler (analyse -> recreate -> save ->
deploy) — the agent Task 4 just fixed (ten error_steps). Live source uploaded
(index_7_.html) + URL on record. The influx (slider-index rate) + axis
(Players on yGold) bug becomes the recreated tool's first acceptance criterion.
A recreation trigger gets drafted when we turn to it (scoped to the game page;
spec.mode + spec.function set so the note tail has a subject). Categories:
(decision, diagnosis)

---

## Open threads (rev 28 state)

1. CURRENT BLOCKER: paste collected_data (A) + tool-generator pod log (B) for
   00688389 -> diagnose the generate_tool_html failure -> fix -> re-run 085
   (clean run = tool + auto-PLAN, Task-3 proof).
2. Game = recreation: draft a tool-recreation-handler trigger for the
   economy-simulator page once the generator is healthy (adopt the live body
   into a component; note tail writes the first recreation NOTES entry).
3. First fix-NOTES entry still pending a successful fix/recreation.
4. Chassis build: helper tests + load_runtime softening + source_agent
   fallback; recreation items to carry spec.function.
5. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type.

---

## Session log — 2026-07-08 (chat handoff prepared)

### Handoff document created (user request)
HANDOFF_2026-07-08_travelling_docs_and_toolgen_bug.md — stands alone for a
fresh chat: first-actions, orientation, DONE state (with snapshot ids), the
generate_tool_html/execute_llm_prompt blocker (run 00688389, ranked
hypotheses, shared-dependency warning), the gamesdesign recreation reframe
(root cause + acceptance criterion), a DATA-TO-COLLECT checklist (the run's
collected_data + pod log + current workflow dump + systemic/latent probe +
pod env + model cross-check), a filled bundle command modelled on
example_bundle.txt, the key facts/schema/gotchas, backlog, and file inventory.

### Bundle command reasoning
Built around execute_llm_prompt as the fulcrum (shared by generate_tool_html,
recreate/analyze, compose_plan, compose_note). Step-0b resolvers for the two
least-certain paths ($LLM_ACTION, $AI_CLIENT); whole-file scopes for the
failing action + AI client; symbol scopes elsewhere; docs 005/020/003/023/001;
schema agent_definitions+doc tables+content_components/pages/page_components/
sites/site_specs/orchestration_states; runtime gamesdesign game page; optional
recreation-action scopes listed separately. Paired with the §5 data checklist
because the bundle carries code/docs/schema/runtime but NOT the failing run's
own error text. Categories: (docs, handoff)

---

## Open threads (handoff state)
Unchanged from rev 28 + carried into the handoff §8. Immediate: collect §5
data, run §6 bundle, diagnose the execute_llm_prompt failure, re-run 085 for
the Task-3 proof, then draft the recreation trigger for the game (Task-4
proof).

---

## Session log — 2026-07-08 (rev 29 — blocker root-caused: schema drift)

### ROOT CAUSE: create_tool_component INSERTs a column production lacks
agent_error_log (orchestration 9f93a988, 16:14:44): "step save_tool failed: ...
create_tool_component: ERROR: column source_agent_type of relation
content_components does not exist (SQLSTATE 42703)". So generate_tool_html
SUCCEEDED and save_tool failed one step later — my earlier reading (LLM step
failed) was an artefact of 120s polling sampling current_step during
generation. The binary shipped ahead of its migration; the action's own comment
names it: NNN_add_component_provenance.sql, "mirroring knowledge_base's pair ...
apply that migration before this binary deploys". knowledge_base HAS both
columns (rag_actions.go inserts them); content_components has neither.
LATENT ~2 months: last created_from='generated' tool = 2026-05-16;
component-creator inserts a different column set and kept working. Probe D
(designed in the handoff) answered exactly this. Task 3 not broken: 0 PLAN rows
is the correct consequence of complete_error bypassing compose_plan.
Categories: (incident, root-cause, schema)

### Fix drafted: 0NN_add_component_provenance.sql
Mirrors knowledge_base's column TYPES dynamically (format_type + pg_attribute +
EXECUTE format ADD COLUMN IF NOT EXISTS) rather than guessing; additive,
nullable, idempotent; guard on both the mirror source and the result; no
snapshot_agent (data table, not agent_definitions). Reuse-first instruction at
the top: find the canonical repo file (find -name '*provenance*' / git grep)
and apply THAT if it exists; else apply the draft and commit it so the code's
reference resolves. Re-run 085 unchanged afterwards — the function name is free
(the failed run inserted nothing; component INSERT was its first statement).
Second-drift check done offline: pages / page_components / site_work_items
INSERT column lists all match the bundle's production schema. Categories:
(rollout, decision)

### Bonus: the empty source_agent mystery is solved
write_doc_plan_action.go:676 and append_doc_note_action.go:837 both set
sourceAgent = params.Headers["agent_type"] (absent in that step context);
create_tool_component_action.go:292 uses
params.ExecutionContext.Sender.AgentType, which populates. rag_actions.go:1257
shares the weakness. Next chassis build: fall back to the execution context in
all three. Categories: (root-cause, backlog)

### 016b -> v6: two durable entries
(1) agent_error_log is the FIRST read — it outlives the pod, names step_name +
action + error_message + context, and must be filtered by orchestration_id;
current_step from polling is sampled, not an attribution; a terminal step's
success_message can name the wrong phase ("Tool generation failed" when saving
failed). (2) Code ahead of DB: SQLSTATE 42703 schema drift, latent until the
first caller; detect with a last-successful-call probe; fix by applying the
named migration and MIRRORING types dynamically; standing pre-deploy check =
grep the diff for new column names and assert they exist in production.
Categories: (docs, gotcha)

### Bundle observations
The rendered bundle carried exactly the right in-scope file plus schema and a
runtime "Recent errors (agent_error_log)" section — that section is what
settled the diagnosis, so future bundles should always include it. Both Step-0b
resolvers misfired (LLM_ACTION -> ai_errors.go; AI_CLIENT -> a stray .txt);
harmless here because the fault lay in an explicitly-scoped file. Categories:
(observation, tooling)

---

## Open threads (rev 29 state)

1. Reuse-check the repo for the provenance migration; apply it (or the draft);
   verify the two columns; re-run 085 → Task-3 proof + first auto-PLAN body for
   the composer review.
2. Then the game: tool-recreation-handler trigger for the economy-simulator
   page (Task-4 proof; first 'fix'-category NOTES row).
3. Chassis build: source_agent fallback (Headers -> ExecutionContext.Sender) in
   write_doc_plan/append_doc_note/rag_actions; doc_actions_helpers_test.go;
   diagnose_load_runtime no-anchor softening; pre-deploy column-existence check.
4. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation; recreation items to carry
   spec.function.

---

## Session log — 2026-07-08 (rev 30 — provenance applied; canonical file found parked)

### Provenance migration APPLIED — mirror vindicated
NOTICE: source_agent_type character varying(100), source_orchestration_id
character varying(255) (mirrored from knowledge_base); verify: both present,
nullable. Note the orchestration id is VARCHAR(255), not uuid — the dynamic
format_type mirror avoided a wrong hardcoded type (a uuid column would have
required a cast for the string the action passes via nullIfEmpty).
Categories: (rollout, decision)

### The canonical migration existed — parked in a docs folder
find: docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/
NNN_add_component_provenance.sql. git grep for the name hits only the Go
comment. So the file was written but never renumbered into the migrations
directory — the exact mechanism by which a deploy skips a migration. Follow-ups
recorded: (a) cat + diff it against what we applied (may add indexes, other
tables, or a backfill we have not applied); (b) move it into the migrations
path with a proper number so the code reference resolves. Categories:
(gotcha, process)

### Pre-flight written for the three unchecked tables
create_tool_component writes five tables. content_components fixed; pages +
page_components verified from the bundle schema; site_work_items,
site_nav_groups, site_nav_items were never covered and are equally unexercised
since 2026-05-16. drafts/preflight_toolgen_columns.sql — read-only; query 1
lists missing columns (expect 0 rows), query 2 asserts the ON CONFLICT targets
pages(site_id,name) and site_nav_groups(site_id,group_key) exist as unique
indexes. Run before re-running 085 rather than discovering a second drift
mid-run. Categories: (rollout, gotcha)

---

## Open threads (rev 30 state)

1. Run drafts/preflight_toolgen_columns.sql (expect 0 missing columns; two
   unique indexes present) → then re-run 085 → paste status, doc_plans proof,
   and the PLAN body (composer review). Task-3 proof closes.
2. Diff the parked canonical migration; move it into the migrations directory.
3. Then the game: tool-recreation-handler trigger for the economy-simulator
   page (Task-4 proof; first 'fix'-category NOTES row).
4. Chassis build: source_agent fallback (Headers -> ExecutionContext.Sender) in
   write_doc_plan/append_doc_note/rag_actions; doc_actions_helpers_test.go;
   diagnose_load_runtime no-anchor softening; pre-deploy column-existence check.
5. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type parameterisation; recreation items to carry
   spec.function.

---

## Session log — 2026-07-09 (rev 31 — drift fix proven; PLAN write failed into containment)

### Drift fix PROVEN: tool creation restored
Pre-flight clean (0 missing columns across site_work_items/site_nav_groups/
site_nav_items; both ON CONFLICT unique indexes present). Re-run f3bec131 /
orch 036cd3bd: COMPLETED at `complete`; component f70cce61 created
(component_level=tool, is_active=t); pages /tools/tool-xp-curve-designer.html
and /guides/tool-xp-curve-designer-guide.html active. save_tool works; the
42703 drift is closed. Categories: (milestone, rollout)

### Task 3 HALF-proven: 0 doc_plans rows, run ended at `complete`
Shape is diagnostic: terminal `complete` (not complete_error) + no PLAN = a doc
step errored into its containment target (config.error_step="complete"), which
is the designed "docs never fail creation" behaviour. Wiring routes correctly;
the PLAN body was never persisted. Elimination: index_plan cannot be the
culprit (it runs after write_plan, so a doc_plans row would exist). Config keys
verified against action source (plan_body_field, plan_source, subject_type,
subject_key_field; collection, content_field) — all match the migration.
docResolveSubject uses ExtractNestedFieldString, which takes a full dotted path,
so the three-level input_data.spec.function is not obviously wrong. No further
guessing: agent_error_log by orchestration_id is the decisive read (the lesson
banked yesterday). Categories: (incident, rollout)

### Watch item: provenance source may be 'generic', not the agent type
The earlier error-log row showed agent_type='generic' on pod agent-chassis-...
(tool-generator runs on the generic chassis). If the new
content_components.source_agent_type stamps 'generic', then
ExecutionContext.Sender.AgentType is no better than Headers["agent_type"] for
provenance, and the planned doc-action fallback needs rethinking (the
config-declared `source` field is already reliable). Query queued.
Categories: (observation, backlog)

### Mitigation in pocket
If the error proves to be call size/timeout on compose_plan (its input carries
the entire generated HTML), truncate inside the template with
{{printf "%.4000s" .generated_html.result}} — workflow-only, no Go build.
Categories: (design)

---

## Open threads (rev 31 state)

1. Paste agent_error_log for 036cd3bd (+ the provenance stamp + tool_docs
   count) → identify compose_plan vs write_plan failure → fix → re-run 085 with
   a fresh function name (tool-xp-curve-designer now exists; either supersede
   via a new run on a new function, or delete the artefacts first).
2. Diff + relocate the parked canonical provenance migration.
3. Game recreation (Task-4 proof; first 'fix'-category NOTES row).
4. Chassis build: source_agent fallback (pending the watch item);
   doc_actions_helpers_test.go; diagnose_load_runtime no-anchor softening;
   pre-deploy column-existence check.
5. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function.

---

## Session log — 2026-07-09 (rev 32 — compose_plan diagnosed; template rule established)

### ROOT CAUSE of the missing PLAN: my template, not the action
agent_error_log (orch 036cd3bd, 11:41:36, TEMPLATE_FIELD_ERROR): compose_plan
failed rendering `{{.generated_html.result}}` — "can't evaluate field result in
type interface {}". Established rule from this run: execute_llm_prompt with
output_format "text" hands the template the BARE STRING ({{.generated_html}});
with output_format "json" it hands a map ({{.tool_analysis.result | toJSON}} —
the live precedent in recreate_tool). Action CONFIG paths are a different
resolver and keep .result (save_tool's html_content resolved fine in the same
run; it hard-fails on empty html and did not). Corroboration: {{.site_record.domain}}
renders in the same template, so map traversal was never the issue. Containment
behaved as designed: run COMPLETED at `complete`, tool created, docs skipped.
Categories: (root-cause, gotcha)

### Latent same-class bug caught in Task 4 before it fired
My recreation compose_note reached {{.tool_analysis.result.tool_name}} /
.tool_type / .purpose — keys never verified (they appear only inside
analyze_tool's own prompt schema). Replaced with the proven
{{.tool_analysis.result | toJSON}} dump. fixer/improver note templates switched
to | toJSON for whole-object dumps. Migration:
drafts/0NN_fix_prompt_template_field_paths.sql (4 snapshots, idempotent
replaces, guards asserting new forms present and old forms absent).
Categories: (design, rollout)

### Two schema traps banked
(1) agent_error_log.orchestration_id is TEXT — the ::uuid cast errors with
"operator does not exist: text = uuid"; §0-REF query corrected.
(2) Provenance stamps the CHASSIS: the new component has
source_agent_type='generic' (pods are agent-chassis-*; agent_error_log.agent_type
also 'generic'). Therefore ExecutionContext.Sender.AgentType is no better than
Headers["agent_type"], and the planned doc-action fallback is DROPPED — the
config-declared plan_source/note_source are the reliable provenance and are
already carried. Backlog item closed by evidence rather than by code.
Categories: (schema, decision)

### Provenance columns proven in use
content_components row for tool-xp-curve-designer: created_from='generated',
source_agent_type='generic', source_orchestration_id='036cd3bd-...'. The
columns added this morning are populated on the first new component.
Categories: (rollout)

---

## Open threads (rev 32 state)

1. Apply 0NN_fix_prompt_template_field_paths.sql → re-run 085 with
   SPEC_FUNCTION=tool-drop-rate-tuner (the old name now occupies the unique
   index) → paste doc_plans proof + PLAN body (composer review) → Task 3 CLOSES.
2. tool-xp-curve-designer has no PLAN (predates the working hook) — leave;
   a later improve/recreate can write one.
3. Diff + relocate the parked canonical provenance migration.
4. Game recreation (Task-4 proof; first 'fix'-category NOTES row) — now safer,
   its note template no longer reaches unverified keys.
5. Chassis build: doc_actions_helpers_test.go; diagnose_load_runtime no-anchor
   softening; pre-deploy column-existence check. (source_agent fallback: DROPPED
   — see above.)
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function.

---

## Session log — 2026-07-09 (rev 33 — TASK 3 PROVEN; index_plan hang)

### MILESTONE: the system wrote its own first PLAN (Task 3 proven)
Run 1923badd (after the template fix): doc_plans body for
tool-xp-curve-designer. Composer review PASSES: sections in order; five
standard checks verbatim; fence intact; interaction check built from REAL
selectors copied from the generated HTML (#curveType, #xpTableBody tr) — the
"never invent a selector" instruction held; deliberate-decisions bullets record
real intent (reused growth-factor field; raw canvas by choice; <=25-level dot
threshold). Deviations noted: uses "action":"select" with a value (a verb the
Tier-4 criteria vocabulary must define); folded the required "v1 kept simple by
design" sentence into the last bullet rather than its own. Confirmation queries
queued (source/length/has_fence; selector existence in html_template; KB row
count). Categories: (milestone, criteria)

### NEW INCIDENT: index_plan hung 2641s+, workflow timeout 480s did not fire
rag_index chunks content (1000/200) then calls GenerateEmbedding per chunk
against ollama-adapter...:11434. Embedding ERRORS are non-fatal by design
("storing without embeddings"); a STALL is not an error and no deadline exists,
so config.error_step never fired. Observation: timeout_seconds does not govern
in-process action execution (hypothesis: only awaited responses / child
orchestrations; unconfirmed). Stopgap drafted:
drafts/0NN_bypass_index_plan_until_embed_timeout.sql — write_plan.next_step ->
complete; index_plan left defined, unreachable, annotated with the reason and
the re-enable UPDATE. Safe because write_plan persists the PLAN BEFORE
index_plan (Postgres = truth; KB copy is derived). Structural fix for the next
build: context.WithTimeout around GenerateEmbedding (or http.Client{Timeout} in
aiservice.OllamaClient) so a stall degrades into the existing non-fatal path;
consider an action-level deadline in the chassis. 016b v8 entry written:
"error containment does not protect against a hang". Categories: (incident,
gotcha, design)

### Env-prefix trap repeated; and create_tool_component updates in place
The command was `SPEC_FUNCTION=... SPEC_NAME=... SPEC_DESC="..."; bash ./085...`
— the semicolon ends the assignments, so the child saw none of them and the
script used its defaults (banner: Function: tool-xp-curve-designer). So
tool-drop-rate-tuner was never created and the SAME tool was re-run. Side
finding: save_tool succeeded against the existing function and there is still
exactly one component row (f70cce61) — create_tool_component updates an existing
function in place rather than duplicating (consistent with
idx_cc_tool_function_unique). Categories: (gotcha, schema)

---

## Open threads (rev 33 state)

1. Apply 0NN_bypass_index_plan_until_embed_timeout.sql → verify write_next =
   complete.
2. Paste the three confirmation queries (auto-PLAN provenance/length; selector
   existence; KB tool_docs count) → composer review closes formally.
3. Game recreation via tool-recreation-handler (Task-4 proof; first
   'fix'-category NOTES row).
4. Chassis build: embedding deadline (+ action-level deadline);
   doc_actions_helpers_test.go; diagnose_load_runtime no-anchor softening;
   pre-deploy column-existence check. Then re-enable index_plan.
5. Tier-4 runner: criteria vocabulary must include the "select" interaction verb
   the composer emitted.
6. Standing opens: KB tool_docs write (now blocked on the embedding deadline);
   deploy_tool_to_site source_* stamp; rag_index source_type; recreation items
   to carry spec.function; diff + relocate the parked provenance migration.

---

## Session log — 2026-07-09 (rev 34 — Task 3 proven formally; selector doubt; stall evidence)

### TASK 3 PROVEN (formal): auto-PLAN provenance + size confirmed
doc_plans: subject_key tool-xp-curve-designer, source='tool-generator',
is_current=t, has_fence=t, body_len=2982 (under the 3000-char instruction,
tightly), created 13:07:50. The creation-side hook is proven end to end.
Categories: (milestone)

### OPEN: one criteria selector may be invented (#xpTableBody)
Check returned has_curvetype=t, has_xptablebody=f. NOT yet a conclusion — my
LIKE demanded the exact string id="xpTableBody"; curveType matching proves the
HTML uses double-quoted ids, so that exact form is absent, but the token may
exist under a different id or only inside the JS. Decisive query queued
(token_anywhere + substring of the actual <tbody ...> tag + formulaBox/
statsStrip/xpChart tokens). If the id genuinely differs, the composer obeyed
"never invent" for the control (#curveType) but not for the expectation
(#xpTableBody tr). REMEDY IS A CHECK, NOT A PROMPT SCOLDING: Stage 5 (Tier-2
static) should validate every criteria selector against the component's
html_template at write time, dropping or -EDIT-marking unverifiable ones. Then
supersede the PLAN with a corrected criteria block. Categories: (criteria,
design, backlog)

### index_plan bypass APPLIED; stall evidence gathered
Snapshot 1bca62f6; UPDATE 1 x2; guard; COMMIT. write_plan -> complete;
index_plan defined, unreachable, annotated. Cluster: ollama-adapter pod Running
(2d19h) + Service present -> reachable but unresponsive (stall profile);
knowledge_base tool_docs = 0 -> hung on the FIRST chunk (INSERT follows embed);
agent-chassis pod shows RESTARTS 1 (57m ago), timing consistent with the hang
tripping a liveness probe -> a wedged action may kill the pod (blast radius
larger than one slow step), and it explains the empty log grep (logs are
pre-restart; use --previous). Confirmation commands queued: describe pod (Last
State), logs --previous grepped by orchestration id, and BOUNDED curl probes
(-m 5 /api/tags, -m 20 /api/embeddings) — never probe a suspected stall
unbounded. Structural fix unchanged: context deadline around GenerateEmbedding;
consider an action-level deadline so no action outlives timeout_seconds.
Stuck run 46cd5299 likely orphaned by the restart — leave to the reaper, do not
hand-edit orchestration state. Categories: (incident, gotcha)

---

## Open threads (rev 34 state)

1. Selector query -> close the composer review (and, if needed, supersede the
   PLAN's criteria block).
2. Cluster: restart cause (Last State), --previous logs, bounded curl probes.
3. Game recreation via tool-recreation-handler (Task-4 proof; first 'fix' NOTES
   row).
4. Chassis build: embedding deadline + action-level deadline;
   doc_actions_helpers_test.go; diagnose_load_runtime no-anchor softening;
   pre-deploy column-existence check. Then re-enable index_plan.
5. Stage 5 (Tier-2 static): NEW requirement — validate criteria selectors
   against html_template at PLAN-write time; plus the "select" interaction verb
   for the Tier-4 vocabulary.
6. Standing opens: KB tool_docs write (blocked on the deadline);
   deploy_tool_to_site source_* stamp; rag_index source_type; recreation items
   to carry spec.function; diff + relocate the parked provenance migration.

---

## Session log — 2026-07-09 (rev 35 — CORRECTION: OOMKill, not stall; selector invention confirmed)

### CORRECTION: the "hang" was an OOMKill; my stall hypothesis was wrong
describe pod: Last State Terminated / OOMKilled / Exit Code 137 / Finished
13:08:24 UTC. Log: index_plan "Calling action handler" 13:08:01.218Z → dead ~23s
later. orchestration_states stayed EXECUTING_STEP because a dead pod writes
nothing; since_s measured time since the crash. Embedder proven healthy by
BOUNDED probes: /api/tags lists nomic-embed-text; /api/embeddings returned a
768-dim vector well inside -m 20. 016b v8 rewritten accordingly: "EXECUTING_STEP
forever usually means the worker died, not that the step is slow" — triage via
RESTARTS, describe (Last State), logs --previous. The missing-deadline advice is
retained as hygiene, explicitly demoted from "cause". The bypass migration
stands (it removes the step that triggered the OOM) but its stated mechanism is
superseded. Categories: (correction, incident, gotcha)

### Memory suspects (measure, do not theorise)
Params dump shows CollectedData carrying agent_config, agent_definition, the
entire site_specs research blob, and __raw_message__ nested inside
__raw_message__; plus the chassis logs the WHOLE params object at Info on every
action (DEBUGaa lines). index_plan was the last, largest step. Queued: size +
raw_message_copies of collected_data for the run; pod Limits/Requests; top pod;
logs --previous tail before the kill. Candidate fixes ranked after measurement:
remove DEBUGaa full-params logging; stop nesting __raw_message__; drop consumed
blobs (site_specs/agent_definition) from carried state; memory limit as stopgap
only. Categories: (incident, design, backlog)

### CONFIRMED: composer invented #xpTableBody and #statsStrip
token_anywhere=f; the real element is a bare <tbody> (no id); statsStrip absent;
formulaBox/xpChart/curveType real. So the interaction check's expect selector is
unsatisfiable and the behaviour contract names two non-existent ids. The rule
held for the control it acts on and failed for the thing it asserts on. Remedy
is a CHECK, not a sterner prompt (the prompt already forbids invention):
Stage-5 Tier-2 must validate every criteria selector against html_template at
write time (drop or -EDIT-mark unverifiable ones); same for tool-auditor. This
PLAN gets superseded with a corrected criteria block once the real ids are
listed. Categories: (criteria, design, milestone-caveat)

### Side observation
github-actions-runner-...-lhg9l: CrashLoopBackOff, 3207 restarts, StartError
(runc cgroupsPath format). Deploys still work via the healthy replica. Unrelated
to this arc; worth cleaning up. Categories: (observation)

---

## Open threads (rev 35 state)

1. Real-ids query → supersede the PLAN (criteria + behaviour contract).
2. Memory measurements → OOM fix (DEBUGaa logging / __raw_message__ nesting /
   carried blobs / limits).
3. Game recreation via tool-recreation-handler (Task-4 proof; first 'fix' NOTES
   row).
4. Stage 5 (Tier-2 static): validate criteria selectors against html_template —
   now firm, justified by a real miss. Plus the "select" interaction verb for
   the Tier-4 vocabulary.
5. Chassis build: embedding deadline (hygiene); doc_actions_helpers_test.go;
   diagnose_load_runtime no-anchor softening; pre-deploy column-existence check;
   then re-enable index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function; diff +
   relocate the parked provenance migration; github-actions-runner crashloop.

---

## Session log — 2026-07-09 (rev 36 — memory hypothesis disconfirmed; supersede drafted)

### Memory-growth hypothesis DISCONFIRMED
collected_data = 192 kB, raw_message_copies = 2. Not a bomb. The chassis ran
2d19h before its first OOM — consistent with a slow leak, not with indexing a
3 KB PLAN; index_plan was likely the last straw. Attribution withdrawn. The old
pod is now NotFound (ReplicaSet replaced it) so --previous logs are LOST —
lesson: capture crash logs immediately. Re-ranked actions: measure the new pod's
limits + working set (baseline now, repeat in a day); strip the DEBUGaa
full-params Info logging (serialises 192 kB CollectedData twice per action and
prints headers); hunt the leak with pprof (the generic chassis runs all agent
types, so a leak there hits everything); embedding deadline remains hygiene.
Categories: (correction, incident)

### Real ids listed; PLAN supersede + correction NOTE drafted
html_template ids: baseXP, curveHint, curveType, formulaBox, growthFactor,
growthHint, growthLabel, maxLevel, statRow, tableWrap, xpChart. So statsStrip ->
statRow; xpTableBody -> does not exist (table lives in tableWrap; its <tbody>
has no id). drafts/0NN_supersede_xp_curve_plan_selectors.sql: retires v1
(is_current=false, superseded_at) and inserts the corrected body as current
(partial unique index enforces one current row); guards assert a <table> within
300 chars after id="tableWrap" (else rename the check curve-switch-EDIT) and
that the removed ids are absent; ALSO appends a doc_notes correction entry
(categories criteria, docs) — using the NOTES stream for exactly what it exists
for. Offline validation: body < 3000 chars, criteria JSON parses, five standard
checks intact, interaction expects #tableWrap tr. Categories: (criteria,
rollout, dogfood)

### github-actions-runner crashloop diagnosed (side thread)
StartError: runc "expected cgroupsPath to be of format slice:prefix:name for
systemd cgroups" — a node cgroup-driver mismatch (runtime on systemd, pod/runtime
args on cgroupfs), not an app bug. 3,213 restarts. Deploys survive on the healthy
replica. Fix or delete the pod. Categories: (observation, infra)

---

## Open threads (rev 36 state)

1. Apply the supersede migration → verify (v1 retired, v2 current, correction
   NOTE present).
2. Chassis memory: limits + top-pod baseline on the new pod; strip DEBUGaa
   logging next build; pprof if the working set climbs.
3. THE GAME: tool-recreation-handler on the economy-simulator page — Task-4
   proof (first machine-written 'fix' NOTES row) and a repaired tool whose PLAN
   carries the influx/axis criterion.
4. Stage 5 (Tier-2 static): validate criteria selectors against html_template at
   write time (firm requirement, justified by a real miss); add the "select"
   interaction verb to the Tier-4 vocabulary.
5. Chassis build: embedding deadline (hygiene); doc_actions_helpers_test.go;
   diagnose_load_runtime no-anchor softening; pre-deploy column-existence check;
   then re-enable index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function; diff +
   relocate the parked provenance migration; runner crashloop.

---

## Session log — 2026-07-09 (rev 37 — supersede guard fixed: RE_DUP_MAX)

### Guard 1 aborted: Postgres bounded repetition caps at 255
`html_template ~ 'id="tableWrap".{0,300}<table'` → ERROR: invalid regular
expression: invalid repetition count(s). Postgres ARE enforces RE_DUP_MAX=255.
The failure was SAFE: it fired inside BEGIN, so nothing was written; the session
sits in an aborted transaction (clients_db=!#) and needs ROLLBACK before the
retry. Rewrote guard 1 with strpos/substr rather than trimming to {0,255} —
states the containment test plainly (find id="tableWrap", require '<table'
within the next 400 chars) and has no engine limit. 016b entry added
(postgres-regex / re-dup-max / aborted-transaction / guard-design). File
re-validated offline: criteria JSON parses, body 2,943 chars, neither invented
id present, no {m,n} counts remain. Categories: (gotcha, rollout)

---

## Open threads (rev 37 state)

1. ROLLBACK; then apply the corrected supersede migration → verify (v1 retired,
   v2 current, correction NOTE present).
2. Chassis memory: limits + top-pod baseline on the new pod; strip DEBUGaa
   logging next build; pprof if the working set climbs.
3. THE GAME: tool-recreation-handler on the economy-simulator page — Task-4
   proof (first machine-written 'fix' NOTES row).
4. Stage 5 (Tier-2 static): validate criteria selectors against html_template at
   write time; add the "select" interaction verb to the Tier-4 vocabulary.
5. Chassis build: embedding deadline (hygiene); doc_actions_helpers_test.go;
   diagnose_load_runtime no-anchor softening; pre-deploy column-existence check;
   then re-enable index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function; diff +
   relocate the parked provenance migration; runner crashloop (cgroup driver).

---

## Session log — 2026-07-09 (rev 38 — sticky aborted transaction)

### Attempt 2 was a no-op: the aborted transaction persisted
"current transaction is aborted, commands ignored until end of transaction
block" — the session from attempt 1 (RE_DUP_MAX guard failure) was still
poisoned (prompt clients_db=!#), so every statement including the file's own
BEGIN was discarded. ROLLBACK; is the only exit. Hardened:
(a) the supersede file now opens with a defensive ROLLBACK; (clean sessions just
warn "no transaction in progress");
(b) recommend psql -f / \i over pasting — the pastes all session show readline
mangling of comment lines and dollar-quoted bodies, which then confuses the
debugging of the paste itself.
016b entry extended (sticky-abort, psql-paste-mangling). Categories: (gotcha,
tooling)

---

## Open threads (rev 38 state)

1. ROLLBACK; then `psql -f drafts/0NN_supersede_xp_curve_plan_selectors.sql` →
   expect 2 silent guards, UPDATE 1, INSERT 0 1 x2, guard 3, COMMIT → verify.
2. Chassis memory baseline (limits + top pod on the new pod); strip DEBUGaa
   logging next build.
3. THE GAME: tool-recreation-handler on the economy-simulator page — Task-4
   proof (first machine-written 'fix' NOTES row).
4. Stage 5 (Tier-2 static): validate criteria selectors against html_template;
   add the "select" verb to the Tier-4 vocabulary.
5. Chassis build: embedding deadline; doc_actions_helpers_test.go;
   diagnose_load_runtime softening; pre-deploy column check; re-enable
   index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function; diff +
   relocate the parked provenance migration; runner crashloop (cgroup driver).

---

## Session log — 2026-07-09 (rev 39 — guard 1 refused the supersede; static vs runtime insight)

### Guard 1 fired: no <table> within 400 chars after id="tableWrap"
The supersede's own guard blocked the write. Correct behaviour: the replacement
selector #tableWrap tr was an unverified assumption, and the guard stopped a
second unverifiable assertion being written while repairing the first. Nothing
persisted; session aborted (ROLLBACK needed). Categories: (gotcha, rollout)

### DESIGN INSIGHT: acceptance criteria are RUNTIME assertions
Tier-4 evaluates criteria against the RENDERED DOM, not html_template. If the
tool builds rows in JS (plausible: the earlier <tbody> match may be a template
string inside <script>), then #tableWrap tr is valid at runtime yet absent from
the static HTML. Therefore static validation can CONFIRM a selector but never
REFUTE one. Stage-5 requirement refined: check html_template AND its inline JS;
found statically -> confirmed; found only in the script (or created dynamically)
-> mark runtime-only and leave to Tier-4; found nowhere -> drop or -EDIT. The
composer's #xpTableBody / #statsStrip are the "nowhere" bucket (genuine
invention); today's #tableWrap tr may be the middle bucket. Same logic belongs
in tool-auditor. Categories: (design, criteria, decision)

### Inspection queried before finalising
strpos positions of id="tableWrap" / <table / <tbody / <script, whether JS fills
the wrapper (tableWrap...innerHTML), and the 300 chars after the wrapper.
p_table > p_script => JS-built table => relax guard 1 to accept "JS fills
#tableWrap" as evidence and keep the check. p_table = 0 and no <tr => rows are
not a table => retarget or -EDIT. Categories: (rollout)

---

## Open threads (rev 39 state)

1. ROLLBACK; paste the two inspection queries → finalise guard 1 + the criteria
   check → apply the supersede with `psql -f`.
2. Chassis memory baseline (limits + top pod on the new pod); strip DEBUGaa
   logging next build.
3. THE GAME: tool-recreation-handler on the economy-simulator page — Task-4
   proof (first machine-written 'fix' NOTES row).
4. Stage 5 (Tier-2 static): html_template AND inline-JS selector validation;
   runtime-only marking; "select" verb for the Tier-4 vocabulary.
5. Chassis build: embedding deadline; doc_actions_helpers_test.go;
   diagnose_load_runtime softening; pre-deploy column check; re-enable
   index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function; diff +
   relocate the parked provenance migration; runner crashloop (cgroup driver).

---

## Session log — 2026-07-09 (rev 40 — resolved: JS-rendered table; anchor rule settled)

### Inspection result: #tableWrap is an empty div filled by JS
p_wrap=7066, p_script=7116, p_table=14384, p_tbody=14465, wrap_filled_by_js=t,
has_tr_markup=t; the wrapper reads `<div id="tableWrap"></div>`. So the table
and its rows are rendered by script at runtime. #tableWrap tr is therefore a
VALID runtime selector; guard 1 was wrong, the check was right. Guard v3 now
accepts static <table> inside the wrapper OR dynamic evidence (innerHTML fill +
<tr> markup positioned after <script>), and raises a NOTICE stating which path
verified. PLAN body updated to say the script re-renders the table into the
empty #tableWrap. Categories: (rollout, criteria)

### STAGE-5 RULE SETTLED: validate the selector's ANCHOR, not its path
No new criteria schema needed. Check the leftmost id/class token against
html_template: #tableWrap exists -> #tableWrap tr passes (Tier-4 asserts the
rows for real); #xpTableBody exists nowhere -> #xpTableBody tr fails -> drop or
-EDIT. Static checks CONFIRM, never REFUTE; never delete a check just because
the DOM is built at runtime. Applies to tool-auditor too. This distinguishes the
composer's genuine invention from a legitimate dynamic selector. Categories:
(design, decision, criteria)

---

## Open threads (rev 40 state)

1. ROLLBACK; `psql -f drafts/0NN_supersede_xp_curve_plan_selectors.sql` → expect
   NOTICE (static=f, dynamic=t), UPDATE 1, INSERT 0 1 x2, COMMIT → verify.
2. Chassis memory baseline (limits + top pod); strip DEBUGaa logging next build.
3. THE GAME: tool-recreation-handler on the economy-simulator page — Task-4
   proof (first machine-written 'fix' NOTES row).
4. Stage 5: implement anchor validation (html_template) + the "select" verb in
   the Tier-4 vocabulary.
5. Chassis build: embedding deadline; doc_actions_helpers_test.go;
   diagnose_load_runtime softening; pre-deploy column check; re-enable
   index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function; diff +
   relocate the parked provenance migration; runner crashloop (cgroup driver).

---

## Session log — 2026-07-09 (rev 41 — supersede applied; Task 5 activated)

### Supersede APPLIED: corrected PLAN + first correction NOTE
NOTICE: tableWrap anchor verified (static=f, dynamic=t) — the dynamic path, as
predicted. Guard 2 silent; UPDATE 1 (v1 retired with superseded_at); INSERT 0 1
x2 (corrected PLAN current + doc_notes correction entry, categories
["criteria","docs"]); guard 3; COMMIT. tool-xp-curve-designer now has a PLAN
whose interaction check can pass and a NOTES stream that opens with the record
of why it changed — the travelling-docs loop applied to itself. Categories:
(milestone, rollout)

### Task 5 wrinkle identified BEFORE triggering: recreation writes SECTIONS
tool-recreation-handler ends save_page_sections -> update_status -> deploy_page;
it never calls create_tool_component. That explains why the game page has only a
shared hero in page_components: the body is not a component. Consequence: the
note tail's subject_key_field (input_data.spec.function) may name a tool that
never exists as a component row, breaking the docs convention (subject_key =
content_components.function byte-for-byte). Three options parked for the fetch
to decide: (a) accept a tool-subject for a page-rendered artefact; (b) subject
the note to pipeline/build with note_site_id; (c) create the component properly
and have the page reference it. Fetch queued: input_contract,
load_existing_content step config (valid spec.mode values; adoption
research_results?), \dt *section*, and the section count/bytes for
game-economy-simulator. Do not guess. Categories: (design, schema, rollout)

---

## Open threads (rev 41 state)

1. Paste the four Task-5 fetch outputs → settle the subject question → I draft
   086_TRIGGER_recreate_economy_simulator.sh (page_name, spec.mode,
   spec.function, spec.interactive_features naming the influx/axis fixes).
2. Optional: supersede verification selects.
3. Chassis memory baseline (limits + top pod); strip DEBUGaa logging next build.
4. Stage 5: implement ANCHOR validation (leftmost id/class vs html_template);
   add the "select" verb to the Tier-4 vocabulary.
5. Chassis build: embedding deadline; doc_actions_helpers_test.go;
   diagnose_load_runtime softening; pre-deploy column check; re-enable
   index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; recreation items to carry spec.function; diff +
   relocate the parked provenance migration; runner crashloop (cgroup driver).

---

## Session log — 2026-07-09 (rev 42 — supersede verified; two Task-5 blockers found by reading)

### Supersede VERIFIED
doc_plans: v1 (source tool-generator, 2982 chars) is_current=f, retired=t; v2
(source migration, 2998 chars) is_current=t, has_fence=t. doc_notes: one row,
categories ["criteria","docs"], source migration. Categories: (rollout)

### BLOCKER (i): spec is undeclared in tool-recreation-handler's input_contract
required [site_id, domain]; optional [page_name, page_id, sections]. But the
workflow reads input_data.spec.mode, .spec.interactive_features and (Task-4
tail) .spec.function. Same class as the 3b finding (an input the workflow
depends on must be declared). Fix: snapshot + add `spec` to optional.
Categories: (schema, design)

### BLOCKER (ii): recreation's NOTES subject is wrong — it writes no component
load_existing_content = "Load rawHtml and markdown from adoption crawl
research_results"; the workflow ends save_page_sections -> update_status ->
deploy_page and never calls create_tool_component. So ('tool', spec.function)
would be a DANGLING doc (subject_key with no matching content_components.function).
Recreation is site-scoped page work, like component-template-fixer. Correction to
the Task-4 wiring: re-subject append_note to ('pipeline','build') +
note_site_id_field site_record.site_id; drop subject_key_field. Found by reading,
not by a failed run. Side effect: retires the "recreation items must carry
spec.function" backlog item for notes. Categories: (design, decision, correction)

### Correction: site_plan_sections is NOT html storage
Columns plan_id, page_name, ordering, component_name -> site-plan structure
(which component sits where), FK to site_plans. The page body most likely lives
in pages.sections (seen in the pre-flight column list). Confirm before relying.
Categories: (schema)

### Fetch queued
pages.sections type/bytes for game-economy-simulator + tool-xp-curve-designer;
\dt *research*; repo grep for load_existing_content's accepted spec.mode values
(and whether it reads rawHtml/markdown). If the crawl content is unreachable,
load_existing_content error-routes to load_related_context (contained) and the
LLM would rebuild from the spec alone — losing existing behaviour, so prefer
feeding it the real content. Categories: (rollout)

---

## Open threads (rev 42 state)

1. Paste the three reads → then ONE migration: (a) add `spec` to
   tool-recreation-handler.input_contract.optional; (b) re-subject its
   append_note to pipeline/build + note_site_id. Snapshot first.
2. Then 086_TRIGGER_recreate_economy_simulator.sh (page_name, spec.mode,
   spec.interactive_features naming the influx/axis fixes) → Task-4 proof.
3. Chassis memory baseline; strip DEBUGaa logging next build.
4. Stage 5: anchor validation; "select" verb in the Tier-4 vocabulary.
5. Chassis build: embedding deadline; doc_actions_helpers_test.go;
   diagnose_load_runtime softening; pre-deploy column check; re-enable
   index_plan.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; diff + relocate the parked provenance migration;
   runner crashloop (cgroup driver).

---

## Session log — 2026-07-09 (rev 43 — handoff + bundle script; recreation facts settled)

### Where the game body actually lives (settled)
pages.sections is jsonb and EMPTY for game-economy-simulator (30 bytes);
tool-xp-curve-designer sections = [] (2 bytes) and build_status='planned'.
site_plan_sections is site-plan STRUCTURE (plan_id, page_name, ordering,
component_name), not HTML. So the 32 KB game body exists only as deployed HTML
in the sites repo, sourced from the adoption crawl. research_results exists; per
doc 007 it holds result_type 'adoption_crawl' (full markdown + rawHTML) and
'adoption_page' (per-page markdown), and doc 007 line 815 lists
"existing_content / mode: recreate support (pending)" for the content writer.
Categories: (schema, observation)

### spec.mode = "recreate" (fetched, not guessed)
apply_adoption_plan_action.go:625 sets pageSpec["mode"] = "recreate" with the
comment "load_existing_content checks for this value";
load_existing_content_action.go Optional: ["page_id","mode"]. Categories:
(schema)

### Loose end: the new tool page is not deployed
tool-xp-curve-designer page build_status='planned' — the component and page
exist, but the site has not rendered/deployed it. Check the build pipeline
picked up the work item create_tool_component created. Categories: (rollout)

### Handoff + bundle script written (user request)
HANDOFF_2026-07-09_recreation_and_chassis.md — supersedes the 07-08 handoff.
Sections: first actions; where the project stands (Task 3 proven, Task 4 applied
but unproven); a table of faults already fixed (provenance drift, template
resolver, invented selectors, the OOM misdiagnosis); TASK A (two
tool-recreation-handler fixes: declare spec; re-subject the note to
pipeline/build); TASK B (recreate the economy simulator with mode=recreate; the
two game bugs as acceptance criteria; the Task-4 proof query); standing design
decisions (anchor rule; provenance stamps the chassis; containment covers errors
not crashes); TASK C chassis hygiene (OOM leak hunt, DEBUGaa logging removal,
rag_index deadline, re-enable index_plan, planned tool page, runner crashloop);
the bundle; the 016b rules; file inventory.
drafts/bundle_recreation_v1.sh — resolves scope paths first (both guessed paths
missed last time), prints them, then runs cmd/bundle with the recreation +
chassis scopes, docs 007/005/020/003/001/019, schema incl. research_results and
agent_error_log, runtime gamesdesign.co.uk / game-economy-simulator. bash -n
clean. Categories: (docs, handoff)

---

## Open threads (rev 43 state — carried into the handoff)

1. TASK A: one migration — declare `spec` in tool-recreation-handler's
   input_contract; re-subject its append_note to pipeline/build + note_site_id.
2. TASK B: verify research_results holds the game's rawHtml → trigger
   recreation (mode=recreate, interactive_features naming the influx/axis
   fixes) → first machine-written 'fix' NOTES row (Task-4 proof).
3. TASK C: chassis memory (limits, top pod, events, pprof); remove DEBUGaa
   logging; rag_index context deadline; then re-enable index_plan.
4. Deploy the xp-curve tool page (build_status='planned').
5. Stage 5: anchor validation; "select" verb in the Tier-4 vocabulary.
6. Standing opens: KB tool_docs write; deploy_tool_to_site source_* stamp;
   rag_index source_type; diff + relocate the parked provenance migration;
   runner crashloop (cgroup driver); doc_actions_helpers_test.go;
   diagnose_load_runtime no-anchor softening.

---

## Session log — 2026-07-09 (rev 44 — bundle resolver fixed)

### Scope-path resolution: three misses, and why
First run of bundle_recreation_v1.sh resolved 10/13 scopes. Facts learned:
- execute_llm_prompt lives in platform/orchestration/actions/ai_actions.go (no
  _action suffix) — the shared action behind generate_tool_html, compose_plan,
  compose_note, analyze_tool, recreate_tool.
- validate_page_content.go also lacks the _action suffix → filename convention
  is unreliable.
- save_page_sections / update_page_status / check_tool_health did not resolve.
  check_tool_health may not exist yet (Stage 5 unbuilt) — absence is expected,
  not a fault.
Root cause of the misses: v1's resolver grepped for the quoted action name only
inside action files and EXCLUDED registry.go — which is exactly where the
action-name → constructor mapping lives. Rewritten resolve_action(): registry
line → constructor/type → defining file; CamelCase fallback; whole-platform
fallback; misses are non-fatal, dropped from -scope, and print grep candidates
so the miss is diagnosable. bash -n clean. Handoff §7 updated with the path
facts. Categories: (tooling, gotcha)

---

## Session log — 2026-07-09 evening → 2026-07-10 (rev 45 — TASK 5 closed, Task 4 PROVEN, OOM solved, migrations system)

### TASK A applied (→ migration 137)
Snapshot `8701375f` → `UPDATE 1` → guard → `COMMIT`. Verified:
`spec_declared=t · pipeline · build · has_key_field=f · site_record.site_id`.
Both tool-recreation-handler fixes live: `spec` in the contract; note
re-subjected to `('pipeline','build')` mirroring component-template-fixer.
Categories: (migration)

### Pre-flight discoveries: where the game's source actually lives
- `research_results` intact: 1 `adoption_crawl`, 19 `adoption_page`,
  17 `tool_recreation_training`.
- The economy-simulator `adoption_page.raw_html` is the **5.8KB HTML SHELL** —
  the origin site loads an external `game.js` the crawl never captured.
- Fetched the origin `game.js` (still live, 6.5KB): **neither documented bug is
  in it** — no Players chart series at all (its green line is Total Gold on its
  own y1 axis), and population grows probabilistically ≤1/tick.
- Our live 32KB inline page HAS both bugs (line 918 `popInflux =
  parseInt(el.slPop.value)` → `state.players += popInflux`; Players dataset on
  `yAxisID:'yGold'`), and `d9a8e6e8` has TWO `tool_recreation_training` rows
  from 2026-06-05 → **the June recreation generated the live page and
  introduced both bugs itself**. The machine broke it; Task 4's proof is the
  machine documenting its own repair.
- User decision on record: **clean rebuild** (no research_results enrichment).
Categories: (provenance, decision)

### Run 1 `464102f4` — the kcat line-split trap
`kcat -P` is line-delimited: the pretty-printed `input_data` inlined into the
heredoc became ~37 one-line fragment messages, all carrying our headers. The
orchestration under our correlation id completed "after 0 steps" holding a
NEIGHBOURING scheduler message's no-op body (`pre_query`) and `input_data:
null` — headers married to someone else's body (chassis stale-buffer wrinkle
noted). Harmless: no side effects. 086 now compacts to a single line
(`json.dumps` separators) and refuses to send multi-line. Env-prefix trap's
sibling — banked.
Categories: (gotcha, tooling)

### Run 2 `e1018366` — Task 4 PROVEN, bugs faithfully recreated
Full path walked (analyze ~2.5min → recreate ~1.5min → save/validate/deploy →
note). **FIRST machine-written `fix` note** 19:36:35Z: `('pipeline','build')`,
site-stamped, `["fix"]`, source/created_by `tool-recreation-handler`. But the
deployed page had BOTH bugs again: `players += parseInt(slPop.value)`; Players
on the shared gold axis. Trace through `collected_data`: Sonnet's analysis
carried the fixes verbatim (`INFLUX_MAP [0,1,5,15,40,100]`, "raw value must be
mapped to rate array before use") — **but `recreate_tool`'s prompt never
renders `input_data.spec`**; the requirements reached Opus only buried in the
20KB analysis JSON while the shell's label read "New players joining per
tick". Opus trusted the visible markup. THE SEAM RULE: check every consumer of
a spec field, not just the first.
Categories: (proof, seam, gotcha)

### Migration 138 + Run 3 `ca00a1dd` — everything lands
138: `recreate_tool` prompt gains "## Mandatory Behaviour Requirements (from
the recreation spec)" rendered from `spec.interactive_features`, placed after
the functional spec, marked as OVERRIDING the original source. Run 3
(20:27→20:33): `INFLUX_RATES = [0, 1, 5, 15, 40, 100]` consumed as
`state.players += INFLUX_RATES[popIndex]`; three axes `yPrice/yGold/yPlayers`;
live page (33,314 bytes, deployed 20:32:32Z) verified — old bug lines gone.
Second `fix` note 20:33:04Z. **TASK 5 CLOSED.** Note: run 2's note truthfully
said "completeness + validation passed" while shipping recreated bugs — the
precise gap Tier 4 exists to close; the economy-simulator arc is now TWO
independent demonstrations (June's bug-introducing recreation + run 2).
Categories: (migration, proof, fix)

## Session log — 2026-07-10

### v1.0.1102 verified → 139 applied
Commit `8f9fe537` ancestor-checked; both removed DEBUGaa dumps absent from
5,000 lines of new-pod logs; embedding deadline present (default aligned to
the OllamaClient http 120s cap after user pushback — 15s would have TIGHTENED
the real ceiling on a slow embedder; defaults must not change behavior).
139 applied: `write_plan.next_step = index_plan` restored.
Categories: (deploy, migration)

### Migrations system (user: "we could set it up")
Survey: `sql_for_agents/` numbered 001–123 is the live convention;
`scripts/migration/run-migrations.sh` was an EMPTY file; prod has
`migration_backups` but no applied-ledger. Built: `schema_migrations` table
(filename PK, applied_at, checksum, applied_by, notes) via
`124_schema_migrations.sql`; runner filled in (dry-run default, `--apply`,
per-file re-check before executing, md5 recorded, stop-without-record on
failure); arc renumbered **125–139 in applied order** (git mv, content
untouched — snapshot reasons inside are historical); backfill with RUNBOOK
dates. **128 = reconstruction stub** (original lost with the old workspace;
effect verified live: `load_runtime.config.error_step='assemble_bundle'`;
NULL checksum + honest note). 136 = the applied v3 supersede file, identified
by strpos/dynamic-evidence markers; 133 = the APPLIED provenance version
(docs019 file = design doc, diffed, differs). Bootstrap sequence: apply 124
manually (idempotent) → dry run collapses to 124-only → `--apply` self-records.
Categories: (tooling, migration)

### v1.0.1103 + the index_plan proof run that answered the OOM (`75c512bf`)
1103 deployed (fixloop verdict-gate Go half rides it — FYI file in this dir;
no action for us). Proof run fired via 085 (`SPEC_FUNCTION=tool-drop-rate-tuner`,
same-line env prefix, banner tell checked). Result: **pod OOMKilled again**
(exit 137, restart 1 on a 28-min-old pod) minutes after entering `index_plan`.
`--previous` logs empty this time, but Last State + timeline nailed it, and the
code read confirmed: **`chunkContent()` never terminates on content >
chunk_size** — final chunk ends at `len(content)`, `start = end - overlap`
steps BACKWARDS, the same ~200-char tail appends forever → 2Gi in seconds.
Small content returns early (hid it for weeks). BOTH OOMKills were PLAN-sized
bodies (2,982 / 3,010 chars). The "stall", "missing deadline" AND "slow leak"
hypotheses all dead; **the user's no-leak hunch was right**. The deadline
shipped in 1102 was hygiene — the loop runs BEFORE any embed call.
Crashed run's artifacts intact: component + page (`planned`) + PLAN (fence
intact). `tool_docs` still 0.
Categories: (incident, root-cause, gotcha)

### Fix + 140 + 141 (parked)
`chunkContent`: break after final chunk + forward-progress guard for any
overlap config; four regression tests (`rag_actions_chunk_test.go`, 30s
timeout catches loop regressions; pathological overlap==chunkSize covered).
Build clean. **140 applied via the runner** (its first real apply): re-bypass
`write_plan → complete`, root cause recorded in the migration AND a
runbook-§3 pipeline note (`('pipeline','build')`, categories
["migration","diagnosis"]). **141 drafted but PARKED in travelling_docs/** —
moving it into sql_for_agents/ before the fixed image runs would let a stray
`--apply` re-arm the OOM. Runner hardened after user asked "is it safe with
the old files?": empirical audit (133 entries → 123 match regex → only ≥124
are candidates → all ledgered → "Up to date"), plus a LOUD warning for
near-miss names (`NNN-hyphens.sql` etc. — this repo really uses `NNNb_` and
hyphenated names; a well-formed name that silently skips would read as
"applied").
Categories: (fix, migration, tooling)

### Where this leaves the loop
Tasks 3 AND 4 both proven in production. Remaining: deploy the chunk fix →
141 → re-run 085 proof (same function, updates in place) → first
`tool_docs` KB rows. Then Stage 5 (Tier-2 anchor checker) and Stage 6
(Tier-4 runner) — now with doubled evidence they are the missing layer.
Categories: (position)

### TASK 6 CLOSED — index_plan proven on the fixed binary (`05d1fc97`)
v1.0.1104 verified carrying the chunkContent fix (commit `2654d0d1`; tests in
`7813f3eb`) → 141 moved into sql_for_agents/ and applied via the runner
(snapshot, UPDATE 1, pipeline note, guard, COMMIT) → 085 re-fired with
SPEC_FUNCTION=tool-drop-rate-tuner (same-line env prefix; banner tell checked).
COMPLETED at `complete`, empty err, **pod 0 restarts** (28Mi after). `index_plan`
took ~5.5s (14:03:06.9 → 14:03:12.5Z) and wrote the **FIRST
`knowledge_base` `collection='tool_docs'` rows: 4 chunks, 4 embeddings**
(731–976 chars from the 2,904-char PLAN). The crashed run's 3,010-char PLAN
superseded cleanly (write_doc_plan supersede tx, unattended). The step that
killed two pods is a five-second no-event. **Phase A of travelling docs is
proven end-to-end, including the derived index.** Next front: Stage 5 (Tier-2
anchor checker), Stage 6 (Tier-4 runner), and the two `planned` tool pages.
Categories: (proof, position)

### Both planned tool pages deployed (loose end closed)
`needs_content_page` items had completed but nothing enqueued the final
render+deploy hop — pages had full rendered components (17/22KB) yet 404'd.
Inserted two `page_rerender` work items exactly as `create_rerender_items`
does (same spec/item_key/priority); the dispatch loop deployed both within
minutes: xp-curve 31,899 bytes, drop-rate-tuner 37,242 bytes, both HTTP 200,
build_status=deployed. Gap on record: tool creation ends at `complete`
without enqueuing a page_rerender item — the pages deploy only when something
else sweeps. Worth a `create_rerender_item` tail on tool-generator later.
Categories: (fix, seam)

### Stage 5 BUILT — discovery check `tool_acceptance` (+ migration 142)
Tier-2 static checker implemented as a sibling of `tool_health` in
`discovery_checks/` (the runbook's "inside check_tool_health" resolved to the
check-plugin framework once read). Loads the current PLAN's ```criteria fence
(same extraction as load_doc_context), fetches the DEPLOYED page (bounded
12s, 2MB, cached per run), and evaluates the static subset under the ANCHOR
RULE: selector_exists/selector_count/interaction anchors (leftmost id/class
token; confirm never refute; -EDIT skipped), asset_loads (path referenced),
page_status_ok (the fetch), plus built-in shell checks (tool-doc header not
leaked; no '<no value>' residue). No criteria → needs_criteria note (30-day
cooldown), never a fake pass. Failures → improve_tool item (criteria embedded
as acceptance_test, handler tool-improver, shares tool_health's 7-day
per-component cooldown) + acceptance-fail note (source
'tool-acceptance-tier2'). Unit tests pass — including a real catch during
writing: Go regexp \b treats '-' as a boundary, so .tool would have matched
class="tool-page"; class tokens are now compared whitespace-delimited.
Migration 142 wired it into design-discovery-agent.run_checks (safe
pre-deploy: unknown check names warn+skip; activates with the next image).

**Pre-verified against production (manual probe of the live pages): the first
sweep will legitimately fail BOTH tools.** (1) drop-rate-tuner's interaction
anchors #drop-chance/#stat-median exist nowhere — the composer wrote
kebab-case ids while the generator emitted camelCase (#dropChance/#statHalf):
the invented-selector class, caught statically, second sighting. (2) BOTH
tools fail asset_loads — the JS ships inline; /tools/assets/<fn>.js is never
referenced: the PLANs' "Path 1 — extracted on rerender" delivery mechanism is
not what the deploy path does (js-not-extracted). Design decision pending
when the items land: implement extraction or supersede the PLANs' delivery
mechanism + asset criterion.
Categories: (build, migration, proof)

### Stage 5 LIVE — first sweep exactly as pre-verified (`cd0d9731`, v1.0.1107)
The checker missed two release commits (untracked files — v1.0.1106 shipped
without it; caught by ancestry check, committed as ad39ec6e on user's say-so).
v1.0.1107 carries it; triggered design-discovery on gamesdesign: COMPLETED,
two improve_tool items + two acceptance-fail notes, check-level precision
matching the pre-verification byte for byte — drop-rate-tuner failed
["asset","slider-updates-stat"] (interaction anchor #drop-chance absent;
boots/status passed), xp-curve failed asset only. Items carry failing_checks
+ full criteria as acceptance_test. Scope note: only generator-created tools
have content_components rows; adopted/recreated tools are page-sections and
invisible to this check by construction (Tier 4 will see them via pages).
The ladder now has a working Tier 2. Pending decision when the items process:
implement JS extraction (PLAN delivery "Path 1") or supersede the PLANs to
inline delivery. Lesson banked: verify a deploy carries your files by
ANCESTRY (git merge-base --is-ancestor), not by assumption — untracked files
survive any number of release commits.
Categories: (proof, position, gotcha)

### Option B executed — PLANs surrender to delivered reality (143 + 144)
User decision on record: criteria must describe what the system DELIVERS, not
what it aspires to. Migration 143 superseded both PLANs (window-surgery on
the Delivery mechanism section — robust to line-wrapping; exact-byte guard on
the asset check line; both old rows retired, new current rows source=human,
created_by=143_...): asset_loads check REMOVED, delivery section now states
inline + the supersede rationale; drop-rate-tuner's invented kebab-case
interaction selectors corrected to the live page's real ids
(#dropChance range 0.1–25, #statMedian) — second invented-selector fix, this
time caught by the machine (Tier 2) rather than by hand. Correction notes in
both tools' streams; the two Tier-2 items CANCELLED with the resolution in
result (leaving them would send tool-improver chasing a stale contract).
Migration 144 (snapshot 1bca62f6) fixed the composer: compose_plan's standard
checks five → four (asset_loads dropped), Delivery instruction now "inline —
all JS and CSS ship inside the page HTML". Verified: fence_has_asset_check=f
on both current PLANs; all remaining anchors probed present on the live
pages; PLAN chains intact (v1→v2→v3, history preserved). Footnote: the
7-day improve_tool cooldown counts CANCELLED items, so sweeps skip these two
tools until ~07-17 — future tweak: exclude cancelled from the cooldown query.
If extraction ever ships, PLANs supersede forward again.
Categories: (decision, migration, fix)

## Session log — 2026-07-10 (cont.) — Option B executed + Stage 6 adapter built

### Option B (delivery decision) — migrations 143/144
User chose "surrender to reality". 143 superseded both PLANs to inline delivery
(asset_loads check removed; drop-rate's invented kebab selectors #drop-chance/
#stat-median corrected to the live #dropChance/#statMedian) + cancelled the two
Tier-2 items (PLAN-side resolution). 144 fixed compose_plan (five→four standard
checks, inline delivery line). Verified: no asset check on either current fence;
all remaining anchors present on the live pages; chains v1→v2→v3 intact.
Principle banked: criteria describe DELIVERED reality; aspirations live in
roadmaps. Residue: 7-day cooldown counts cancelled items (sweeps skip these two
tools till ~07-17) — future tweak, exclude cancelled.
Categories: (decision, migration, fix)

### Stage 6 — browser-runner-adapter BUILT (code prep; deploy is user's)
Read the mould first (035 adapter guide §1 normative envelope; 002/001/003;
analyser adapter internal/adapters/analyser as the pattern; runner PLAN rev 2).
STEP ZERO: no existing headless/browser capability (webscrape adapter is
Firecrawl, not a driver). Added playwright-go v0.5200.0 (the latest tag
@v0.6100.0 declares a broken upstream module path mxschmitt→playwright-community,
so pinned to 0.5200.0). Wrote:
- cmd/browser-runner-adapter/main.go — signal-handler shell, mirrors analyser.
- internal/adapters/browserrunner/adapter.go — dispatcher. Envelope exactly per
  035: action from body.action ("run_checks"), payload body.data, reply topic
  from responses_topic|reply_to_topic(headers)|reply_to_topic(body); response
  body headers via canonical types.ResponseHeaders (real bools — the bool trap);
  in_response_to_request_id=incoming request_id + request_id reused + fresh
  message_id in kafka headers; ProduceWithValidation; sequential handling;
  shutdown sync.Once; /health+/ready (draining, no browser-per-probe). run_checks
  errors are error_recoverable (fresh Chromium/pod may succeed); unknown action
  unrecoverable.
- run_checks_action.go — playwright-go headless Chromium, desktop 1366×900,
  browser launched PER REQUEST (a crash poisons one run, not the pod). Three P0
  checks: page_status_ok (nav response 200), selector_exists (FULL selector
  Locator.Count in the live DOM after a 2s settle — THE tier that can assert
  #tableWrap tr for real, which Tier 2 only anchors statically), no_console_errors
  (console.error + pageerror). Everything else (interaction, no_horizontal_
  overflow, asset_loads, selector_count, mobile) reported in skipped[], never a
  fake pass; -EDIT skipped; navigation failure = check fail, not infra error;
  mobile-only request skips everything honestly. Probe func injectable → 5 unit
  tests (healthy/failures/nav-fail/empty-criteria/mobile-only) pass.
Packaging: dockerfile (debian-slim + `playwright-cli install --with-deps
chromium`, shared world-readable driver/browser homes, runs as appuser),
kustomize base+overlay (Recreate, 512Mi/2Gi, /health+/ready probes), the
system.adapter.browser-runner.requests KafkaTopic CR (apply once, kafka ns —
NOT in the overlay which forces ai-persona-system), config YAML (local +
overlay copies), and all FIVE makefile points (build-browser-runner-adapter,
build-adapters, push-backend, deploy sed+apply block, redeploy rollout). Full
go build ./... + vet clean. User deploys (Chromium image ~1.2GB, slow build).
Exit test after deploy: §2.15 smoke — hand-produced run_checks against one live
tool page → response with in_response_to_request_id + status=complete + results
matching manual inspection. Then the tool-acceptance-agent orchestrator.
Categories: (build, adapter)

## Session log — 2026-07-11

### Docker build 404 → dead CDN → v0.6100.0 via its declared path; Tier-4 core PROVEN live
User's image build failed at the driver download: playwright.azureedge.net
(and its akamai/verizon mirrors) are DECOMMISSIONED — every playwright-go tag
through v0.6000.0 hardcodes them. The only tag with the fix (v0.6100.0,
driver 1.61.1, installs from registry.npmjs.org + nodejs.org/dist instead) has
a broken go.mod: it declares the project's PRE-RENAME path
github.com/mxschmitt/playwright-go (upstream release accident; confirmed on
the raw tag). A replace directive hits "used for two different module paths",
so the clean route is requiring/importing the module under its DECLARED path:
go.mod now requires github.com/mxschmitt/playwright-go v0.6100.0 and
run_checks_action.go + the dockerfile CLI build line import that path (swap
back when upstream ships a fixed tag — noted at both sites). Build + unit
tests pass unchanged.

Proof before the user rebuilds: (1) the exact image install command ran
locally — driver via npm + Chrome Headless Shell 149 downloaded, exit 0;
(2) NEW env-gated integration test (integration_test.go, BROWSER_RUNNER_IT=1)
drove REAL headless Chromium against the live xp-curve page: status HTTP 200 ✓,
boots ✓, **rows: 20 elements match #tableWrap tr in the live DOM ✓ — the
JS-built assertion the whole tier ladder was argued from**, console clean ✓,
negative control (#definitely-not-here) correctly fails ✓. 4.7s. The Tier-4
core is proven against production reality before the image exists.
Lesson banked: pin dependencies by their DECLARED module path, and when a
download URL 404s, suspect a dead CDN before a wrong version — check what the
newest tag does differently.
Categories: (fix, gotcha, proof)

### Second image-build failure: the driver ignores XDG_CACHE_HOME
Rebuild got much further (CDN fix good: deps installed, Chrome for Testing +
headless shell + ffmpeg all downloaded to /pw-browsers) then failed at
`chmod: cannot access '/pw-cache'`. Root cause read from the module source:
v0.6100.0's getDefaultCacheDirectory() is `$HOME/.cache` on Linux — it IGNORES
XDG_CACHE_HOME. The driver had gone to /root/.cache; /pw-cache never existed.
(My earlier local "proof" that XDG_CACHE_HOME worked was a red herring — the
116MB in the scratch pw-cache was Go's BUILD cache from `go run`, which does
honour XDG_CACHE_HOME. The driver had quietly gone to ~/.cache both times,
which is also why the integration test worked.) Fix: HOME is the knob —
dockerfile now sets ENV HOME=/pw-home (persists across USER appuser, so
install-as-root and runtime-as-appuser resolve the same
/pw-home/.cache/ms-playwright-go path), appuser owns /pw-home (Chromium writes
crashpad scratch under $HOME), browsers root-owned world-readable. Comment
moved out of the RUN continuation (parser gotcha). Lesson: when an env knob
"works" locally, confirm WHICH component consumed it — two caches both
answering to the same variable name masked the miss.
Categories: (fix, gotcha)

### Stage 6 P0 DEPLOYED + SMOKE PASSED — the Tier-4 adapter is live
Third image build succeeded (v1.0.1107 tag via deploy-agents). Two deploy
findings, both anticipated by the docs: (1) the requests topic CR had not been
applied — the adapter showed the exact 035 §2.12 signature (`context deadline
exceeded` on every fetch); applied
system.adapter.browser-runner.requests, fixed. (2) ~10 Evicted pod husks from
node pressure while pulling the ~1.2GB image — one healthy pod landed on
another node; husks deleted.
§2.15 smoke against production: produced a run_checks request (single-line
envelope, real criteria_json from the tool's CURRENT PLAN fence) to the
requests topic with a scratch response topic; response received and verified:
in_response_to_request_id == the produced request_id (the load-bearing
matcher), request_id reused, message_type=response, status=complete,
sender_agent_type=browser-runner-adapter, body headers REAL JSON BOOLS
(is_complete=true/is_error=false as bool — the thunder bool trap avoided),
success=true, run_id echoed. Results (desktop, live xp-curve page): status
HTTP 200 ✓, boots .tool-container in live DOM ✓, console clean ✓; mobile-fit
and interaction honestly SKIPPED (P0 scope). Scratch topic deleted.
**Stage 6 P0 is COMPLETE: the tool verification ladder has a working Tier 4.**
Next: the tool-acceptance-agent orchestrator (criteria → URLs → adapter →
judge → acceptance-run/fail notes + improve_tool items), then P1 (mobile +
full interpreter), P2 (interactions — the tier that would have caught the
economy-simulator bugs end-to-end).
Categories: (deploy, proof, position)

### tool-acceptance-agent BUILT + 145 applied — Tier 4 goes self-driving (gated on next chassis deploy)
Two Go actions in platform/orchestration/actions/tool_acceptance_actions.go:
- request_browser_run — mirrors the analyser/webscrape await pattern
  (AwaitResponse=true; envelope per 035 §1.2; reply lands at output_field).
  Complexity in Go, workflow flat: it resolves the tool's deployed URL from
  pages itself (name == function for generator tools; url_field override
  available) and NO-OP SKIPS without awaiting when the PLAN has no criteria —
  Tier 2 owns needs_criteria; a browser run is never faked.
- judge_acceptance_results — reads the awaited reply through a fallback chain
  (.response.data.results / .response.results / .data.results / .results —
  response shapes vary across the codebase; unit-tested on all three) and
  recomputes the verdict from the results. All pass → acceptance-run doc_note.
  Any fail → acceptance-fail doc_note + ONE improve_tool item (source
  'acceptance', criteria embedded as acceptance_test, failing_checks listed,
  item_key acceptance_fail:<fn>:<site>, handler tool-improver). Recreated/
  adopted tools (no content_components row) get the note but NO item — logged
  honestly for manual routing.
Registry entries added beside request_repo_analysis. Migration 145 applied via
the runner (INSERT new agent, idempotent WHERE NOT EXISTS, no snapshot needed
for a new type; guard asserts the workflow shape; pipeline note per §3):
ensure_site_record → load_docs → request_run (await) → judge → complete, all
error_steps inside config → complete_error. Trigger 087 written (dry-run
default, single-line body, prints the ancestry check first). **GATE: do NOT
fire 087 until the chassis image carrying tool_acceptance_actions.go deploys —
unknown action fails the workflow.** After the first pass verdict lands, wire
trigger points (post-recreation, post-improve, periodic) as the runbook lists.
Categories: (build, migration)

### Tier 4 SELF-DRIVING — first machine acceptance-run note (run bf330ac6)
Chassis carrying tool_acceptance_actions.go deployed (v1.0.1108, pod started
16:11 — the tag was rebuilt from HEAD; the commit-message ancestry was a
tag-reuse red herring, settled by pod-start-time > actions-commit-time and
then by the run itself). Fired 087 on tool-xp-curve-designer: COMPLETED at
`complete` (not complete_error), and the FIRST machine-written acceptance-run
note landed — ('tool','tool-xp-curve-designer'), ["acceptance-run"], source
tool-acceptance, created_by tool-acceptance-agent: "Tier-4 acceptance PASSED —
all 3 evaluated checks passed in headless Chromium (2 skipped: mobile-fit,
curve-switch)". Verdict all_passed=true passed=3 failed=0; no improve_tool
item (correct). Results identical to the hand-produced smoke → the orchestrator
path is equivalent, with zero human in the loop:
PLAN criteria → tool-acceptance-agent → request_browser_run (Kafka) →
browser-runner-adapter (Chromium, live page) → reply → judge → doc_note.
**Stage 6 P0 is COMPLETE and self-driving. The tool verification ladder is
whole: Tier 0 (generation) · Tier 1 (structural) · Tier 2 (static anchor) ·
Tier 4 (behavioural, headless).** The fail path (acceptance-fail note +
improve_tool item) is unit-tested but not yet exercised live — the first
genuinely-failing tool will demonstrate it. Next: wire trigger points
(post-recreation, post-improve, periodic sweep) + P1 mobile / P2 interactions.
Categories: (proof, position)

### Fail path PROVEN live (controlled, reverted) + overview doc written
Wrote OVERVIEW_self_verifying_tools.md — a plain-language explainer of the whole
mechanism (travelling docs + the verification ladder + the autonomous loop) for
talking about it.
Then closed the other half of the Tier-4 loop with a controlled test (runbook
Stage-1 smoke precedent — prove then clean up). Temporarily added ONE
genuinely-failing criterion to drop-rate-tuner's current PLAN
(selector_exists #zzz-failpath-proof, confirmed absent from the live page;
in-place edit, exact-inverse revert, no supersede pollution), fired 087.
Verdict: failed=1, failing_checks=[failpath-proof], improve_tool_created=true.
Both artifacts correct: (1) acceptance-fail doc_note — "Tier-4 acceptance FAILED
— 1 of 4 evaluated checks failed in headless Chromium: failpath-proof: no
element matches #zzz-failpath-proof in the live DOM after settle"; (2)
improve_tool item — status detected, handler tool-improver, severity medium,
failing_checks ["failpath-proof"], acceptance_test = the criteria embedded,
item_key acceptance_fail:tool-drop-rate-tuner:<site>. Cleanup: item cancelled
(was still 'detected' — tool-improver never touched it); PLAN reverted exactly
(len 3046, 5 checks, fence_has_asset_CHECK=false); test note deleted; zero
orphan failpath references. **The full loop is now proven both ways:
pass → acceptance-run note; fail → acceptance-fail note + fix ticket carrying
the criteria. The ladder detects, documents, and hands off a repair with no
human in the loop.** Remaining: let a REAL failure flow through to tool-improver
and back (not manufactured); wire trigger points; P1 mobile / P2 interactions.
Categories: (proof, position)

### Tier 4 goes CONTINUOUS — tool_acceptance_due sweep (+ migration 146)
The trigger-points item, taken the timing-clean way: post-creation/post-improve
hooks would fire BEFORE the page redeploys (testing the old page — creation
ends at 'planned' and improve queues a rerender), so the periodic discovery
sweep is the correct first trigger: it only ever sees deployed pages. New
check discovery_checks/check_tool_acceptance_due.go — for every active tool
with a DEPLOYED page and a current PLAN criteria fence, emits ONE
acceptance_run item (handler tool-acceptance-agent, spec {function,
component_id, page_id}, item_key acceptance_run:<fn>:<site>) unless a verdict
note (acceptance-run OR acceptance-fail — a fail already has its fix ticket in
flight) landed within 7 days or a run is already open. Emitted at status
'triaged' + pipeline 'build' directly (create_rerender_items precedent):
acceptance needs no human judgment, and 'detected' items were observed sitting
unswept on this site (the Tier-2 items sat two days). Priority 90 = after
builds/rerenders, so acceptance tests the NEW page. Verified the dispatch
loop's input_mapping passes spec whole → input_data.spec.function matches the
145 contract exactly. Correct-while-touching (declared in 146):
check_tool_acceptance's improve_tool cooldown now EXCLUDES cancelled items
(the recorded 07-10 follow-up — a cancelled item means resolved another way).
Migration 146 applied (snapshot b294bf7d; design-discovery-agent checks list
+= tool_acceptance_due; pipeline note). GATE: the check rides the next chassis
image (warn-skip until then, the 142 precedent). Proof after deploy: a
discovery sweep should produce an acceptance_run item → dispatch → a fresh
acceptance-run note with NO manual trigger — full autonomy for the top tier.
Categories: (build, migration)

### v1.0.1111 deploy: cooldown fix landed, continuous check MISSED (untracked-file trap, again)
Verified the deploy by commit lineage (not tag): v1.0.1111 == commit f2fb87a,
message "acceptance loop runs in scheduled tasks" — intent was to ship the
continuous sweep. But check_tool_acceptance_due.go was UNTRACKED (??), so
`git commit -a` (modified-tracked-only) caught its sibling cooldown fix in
check_tool_acceptance.go — which IS in v1.0.1111 — while silently skipping the
new file. Migration 146 added the check name to design-discovery-agent, so the
deployed binary warn-skips it (unknown check; the 142-precedent safety holds —
no error). Net: the cooldown fix is LIVE; the continuous sweep is NOT (the
check isn't in the binary). Committed the file as 83ba9bd4; it needs one more
chassis image. Second occurrence of this exact trap (T4 was the Tier-2 checker
itself) — the durable guard is `git status` for `??` before every release, or
committing new files as they're written rather than at release time. GATE:
continuous acceptance activates on the next image built from 83ba9bd4+.
Categories: (deploy, gotcha)

### FULL AUTONOMY PROVEN — discovery → verdict, no human in the chain (v1.0.1112)
v1.0.1112 (commit 83ba9bd4+, pod 11:25Z) carries tool_acceptance_due; verified
in-binary via the sweep's checks_run list (not warn-skipped). First sweep
emitted nothing — CORRECT: both tools had verdict notes inside the 7-day
cooldown (my own testing this week). drop-rate's only blocker was a STALE
Tier-2 acceptance-fail note (2026-07-10 16:25) describing asset+kebab-selector
failures that migration 143 fixed 26 min later (16:51) — deleted as obsolete
cleanup, which made drop-rate eligible. Re-sweep → acceptance_run item emitted
(handler tool-acceptance-agent, priority 90) → dispatch loop CLAIMED it →
tool-acceptance-agent loaded criteria → browser-runner-adapter drove Chromium
on the live page → acceptance-run note "Tier-4 acceptance PASSED —
tool-drop-rate-tuner — all 3 evaluated checks passed (2 skipped)" → item
complete. ZERO manual triggers from the sweep onward (and the sweep itself is a
scheduled maintenance tick in production). **The whole mechanism now runs
unattended: discovery finds a due tool → drives it in a real browser against
its own PLAN's criteria → writes the verdict into its travelling docs.**
Follow-up noted: the cooldown counts Tier-2 acceptance-fail notes too, so a
stale/independent Tier-2 verdict can suppress a Tier-4 run for 7 days — worth
scoping the cooldown query to Tier-4 verdicts (source='tool-acceptance') in a
future refinement; not urgent (coarse don't-spam guard is defensible).
Remaining: a REAL failure flowing through tool-improver + back; P1 mobile / P2
interactions.
Categories: (proof, position)

## Session log — 2026-07-13 (cont.) — status docs + P1/P2 runner + repo-completeness fix

### Summary docs
Wrote STATUS_2026-07-13_where_we_are.md (state-of-play snapshot: milestone
ladder, architecture picture, what's live, what's next). Refreshed
OVERVIEW_self_verifying_tools.md (T9 draft) to record full autonomy + reorder
"next" to P1/P2.

### P1 (mobile) + P2 (interactions) built + PROVEN LIVE
Rewrote run_checks_action.go behind a testable browserPage interface
(real=Playwright chromiumPage, fake=tests). Preserved every P0 behavior
(honest skips, nav-fail=check-fail, -EDIT skip). Added: per-profile runs
(desktop 1366x900; mobile 390x844 touch+mobileUA+DPR3 via
BrowserNewContextOptions IsMobile/HasTouch); no_horizontal_overflow (Evaluate
scrollWidth-clientWidth, 2px tol); interaction (fill/click/select steps via
Locator, then expect selector-exists + text_matches regex). Console errors
evaluated LAST so interaction-triggered errors count. 9 unit tests pass.
LIVE PROOF on xp-curve (real browser): curve-switch interaction actually
SELECTED 'exponential' in #curveType → JS rebuilt the table → #tableWrap tr
present → PASS, on desktop AND mobile; mobile-fit no-overflow PASS at 390px;
mobile-fit correctly SKIPPED on desktop. 9.3s desktop+mobile. This is the
tier that asserts a tool DOES something — the economy-simulator bug class.
Migration 147 sets tool-acceptance-agent.request_run profiles=["desktop",
"mobile"] (safe with the P0 adapter — still desktop-only+skips; activates on
the P1/P2 image).

### Repo-completeness fix (important)
git status exposed that the ENTIRE browser-runner-adapter package was NEVER
git-tracked — prod images worked only because the Dockerfile does COPY . .
(includes untracked working-tree files), so a fresh clone was missing the
whole Tier-4 runner. Committed it complete (53a5b518): cmd, dispatcher,
P0/P1/P2 runner + tests, dockerfile, config, kustomize base+overlay. Durable
lesson (extends the ?? guard): Docker COPY . . masks untracked Go — the build
succeeding is NOT proof the repo is complete; check `git status` for whole
untracked packages, not just individual files.
GATE: P1/P2 activate when a new browser-runner-adapter image (from 53a5b518+)
deploys AND the chassis carrying 147's config is in effect (147 is DB-only,
already applied). Both the adapter image and mobile/interaction criteria then
go live together.
Categories: (build, proof, gotcha)

---

## Session log — 2026-07-14 → 2026-07-16 — Tier-4 goes live end-to-end; the loop closes GREEN on a real bug

*(Detailed turn-by-turn narrative lives in the HANDOFF Turn log, T14–T17. This is
the durable chronological summary.)*

### 2026-07-14 — P1/P2 verified live; composer fix; FIRST real failure

**P1/P2 already deployed (verify against the pod, not git).** The v1.0.1114
browser-runner-adapter build already carried P1 (mobile) + P2 (interactions):
images build from the LOCAL working tree, so the 07-13 source landed in the
image regardless of commit order. Confirmed via the binary's SYMBOL TABLE
(`go tool nm`: `runInteraction`, `splitByProfile`, `(*chromiumPage).HorizontalOverflow`).
**Durable trap banked:** grepping a Go binary for a SHORT string literal proves
nothing — Go compiles ≤16-byte constants used only in equality comparisons into
integer immediates, so `page_status_ok` (14B) greps ABSENT from a binary that
plainly implements it while `no_horizontal_overflow` (22B) greps present. Use
`go tool nm` (these images ship unstripped); the in-pod BusyBox has no `strings`,
so `kubectl cp` the binary out (sha256 both ends). Fixed the memory that had
recommended the wrong technique.

**Tier-4 P1/P2 proven live** (run `af5a4ac5`, xp-curve): **9 passed / 0 failed /
1 skipped** vs T8's P0 baseline of 3-evaluated/2-skipped. Adapter log per
(check,profile): curve-switch interaction "produced the expected result
(#tableWrap tr)" on desktop AND mobile; mobile-fit "no horizontal overflow on
mobile"; mobile-fit correctly SKIPPED on desktop.

**Composer/judge defect found + fixed (chassis).** The pass note said "1 skipped:
mobile-fit" while mobile-fit had PASSED on mobile — the judge counted skips per
check *id*, not per (check,profile), so a check that ran on mobile and was
correctly skipped on desktop read as "never checked". Fixed `extractRunResults`
to label every result `id@profile` (Passed/Failed/SkipList/Details); note gains
"across profiles: …"; `failing_checks` stays bare deduped ids (fixer matches PLAN
ids), new `failing_instances` carries the profile detail. Also scoped the
acceptance-due COOLDOWN to `source='tool-acceptance'` — Tier-2 writes
`tool-acceptance-tier2`, and letting its STATIC fails suppress a BEHAVIOURAL run
had it backwards (a static failure is when you most want the browser to look). 6
unit tests, regressions pinned to the live payload.

**FIRST real (non-manufactured) failure.** The only never-verified tool with
criteria was `tool-archetype-taster-quiz` on **vonc.com** (a different site, a
different render path). It failed 3 of 7 checks — of two OPPOSITE kinds:
- `boots` — FALSE failure: the PLAN asserted `.tool-container` (the generator
  convention); this page-section tool ships `.tool-archetype-taster-quiz-section`;
  `.tool-container` occurs ZERO times. Stale criteria (Option-B/143 class).
- `mobile-fit` — GENUINE failure: the page really overflows at 390px.
PARKED the improve_tool item before dispatch (it bundled the false criterion with
the real bug — the fixer would have chased a stale contract). Diagnosed the real
one in real Chromium: culprit `div.footer-legal` (506px in a 390px viewport),
inside vonc's SITE FOOTER, overflowing every page (homepage included) — NOT the
tool.

**Design defect this exposed + fix: ATTRIBUTE, then ROUTE.** `no_horizontal_overflow`
measures the whole DOCUMENT, but the ticket it raised was TOOL-scoped — one
overflowing footer would raise an unfixable improve_tool ticket for every tool on
the site, every run. Fix (user chose "attribute then route"): the adapter now
locates the tool's `container`, names the widest offender, and stamps each result
`scope` = tool | chrome | unknown (+ culprit, component). The judge routes:
`chrome` → a `responsive_fix` item for `component-template-fixer` (the existing,
dispatched route), `tool`/`unknown` → improve_tool as before. `unknown` NEVER
routes to chrome (an unlocatable container falls back to the tool). Dedup key
`chrome_overflow:<component>:<profile>` (no tool → one site ticket, not one per
tool). Container fallback `.tool-container, [class*="tool-"][class*="-section"]`
covers both delivery paths. **Verified live on vonc:** scope=chrome,
component=site-footer, culprit=div.footer-legal.

**Migration 148** superseded the quiz PLAN (143 precedent): boots re-anchored to
the real section class; the pre-144 `asset_loads` check removed; `container`
added; and `quiz-flow-EDIT` replaced with a REAL interaction (click
`.quiz-option-btn` ×3 → `.result-archetype-name`) PROBED PASSING in real Chromium
before being written into the PLAN (never author a criterion you haven't watched
pass). **Migration 149** (`149_composer_emits_container`) taught the composer to
emit `container` (copied from the generated HTML, never assumed) and anchor
`boots` on that same root. Numbering collided: another workstream landed its own
149 + 150 — TWO 149s in the ledger; next free was 151. (144's lesson re-learned:
`prompt_template` holds REAL newlines, so anchor guards on single-line substrings
and inject the container line with `chr(10)`.)

**Proven on v1.0.1116** (user built chassis+adapter mid-turn; verified by symbol
table): run `f0019bd6` on the quiz → note "PASSED — all 8 of the tool's own checks
passed", root cause "site chrome, not this tool: mobile-fit@mobile —
div.footer-legal in site-footer", ONE responsive_fix item
(`chrome_overflow:site-footer:mobile`, handler component-template-fixer), NO
improve_tool item. The tool was not blamed for the footer.

Categories: (proof, build, gotcha, decision)

### 2026-07-14 (cont.) — the fixer LIED; targeted chrome_overflow_fix built

Promoted the routed responsive_fix item. `component-template-fixer` returned
**`"fixed": true`** and changed NOTHING in the footer — re-measured the live page,
identical 506px overflow. **The behavioural tier caught a fixer that lied
("complete" ≠ fixed).** Root cause: `fixInjectResponsiveCSS` reads
`spec.slot_name`, and when absent DEFAULTS to `slotName := "header"`, then injects
a hardcoded header-nav CSS block — so it "fixed" the HEADER for a FOOTER defect.
Systemic: ALL 54 responsive_fix items ever raised have NO slot_name — every one
defaulted to the header (page-section responsive findings were never really
fixed, just closed). *(Left untouched — another workstream's backlog; flagged.)*
REVERTED the unrequested header injection (marker-guarded; it had not gone live).

**Built `chrome_overflow_fix`** (user chose: new targeted fix type, legacy path
untouched so no mass edits fire elsewhere). Adapter now also emits machine handles
`culprit_selector` (div.footer-legal) and `slot` (footer, via `closest('footer')`);
judge puts them in the spec (`fix_type=chrome_overflow_fix, slot_name,
overflow_selector`); the new fixer REFUSES to run without them rather than guess.
Selector regex-validated before it reaches a `<style>` block (browser→HTML
boundary); idempotent per-selector marker; appended after the slot's own style so
it wins on order without `!important`. **Proved the CSS on the live page BEFORE
shipping:** injected exactly what `buildOverflowCSS` produces into vonc.com in real
Chromium → footer-legal 506px→326px, document overflow 58px→0, flex-wrap:wrap.

Categories: (bug, build, proof, gotcha)

### 2026-07-15 — targeted fixer works; then the DEEPER bug: wrong LAYER

v1.0.1119 verified live (both binaries by symbol table). Re-ran 087 (one run
orphaned in the pod-rollover response-topic gap — re-fired cleanly). Adapter
returned full attribution (scope=chrome, culprit_selector=div.footer-legal,
slot=footer). Judge raised a CORRECTLY-specced ticket. Fixer ran `fixChromeOverflow`
and this time patched the FOOTER slot (header untouched) with the right CSS —
verified in DB (footer 3741→4017, flex-wrap:wrap targeting div.footer-legal).
**detect→attribute→route→fix-correct-target proven.**

**But it did not go live — the vonc landmine.** Drove the deploy via the queued
`stale_sc_footer` rerender, which has `refresh_site_components: true` — that
REGENERATED the footer from its content_component template and WIPED the patch
(footer 4017→3935, still_patched=f). `site_components.rendered_html` is a RENDERED
ARTIFACT; both `chrome_overflow_fix` and the legacy fixer wrote to it, so any
refresh erases the fix. The durable source is `content_components.html_template`.
This is the `[[vonc-spark-workstream]]` memory landmine, now shown to bite the
fixer path.

**Root bug at the durable layer:** vonc's footer renders from content_component
`footer-4-column` (id 09034086-a581-4bba-a5b4-760d863bb2df), whose template had
`.footer-legal { display:flex; gap:2rem; }` with NO flex-wrap. Shared by 8 sites
(incl. gamesdesign e33263f4) — this one rule overflows the footer on 8 sites.

**User chose: fix template + redesign fixer.** Migration **151**
(`151_footer4col_flexwrap`; numbering collided AGAIN — a `151_gripper_spec_sheet_component`
from another workstream also exists and FAILS on a duplicate content_component,
blocking the runner past it — next free number is 152) added flex-wrap:wrap +
justify-content:center to `.footer-legal` in the shared template, full pre-edit
template backed up in a doc_note (`created_by='151_footer4col_flexwrap'`;
subject_type must be tool|pipeline — 'component' is rejected by a check
constraint). Re-triggered stale_sc_footer (refresh now pulls the FIXED template) →
footer regenerated WITH the wrap → live on vonc. **Re-ran 087: `mobile-fit@mobile`
now PASSES** — note "all 9 of the tool's own checks passed", mobile-fit@mobile in
the Verified list. **THE REAL-BUG LOOP IS PROVEN GREEN END-TO-END.** The other 7
sites have the fixed template but stale rendered_html; they self-heal on their
next refresh (not force-rerendered — 7 live sites, left to natural cadence).

**Fixer redesigned (part 2).** `fixChromeOverflow` now resolves the slot's backing
component (`site_components.component_id` → `content_components.html_template`) and
patches the DURABLE layer (survives refresh). Falls back to rendered_html ONLY
when a slot has no backing component, and then reports `durable:false`/"TRANSIENT"
honestly. Reports `shared_sites` (blast radius) — patching a shared template is
correct for a genuine shared CSS defect but never silent. Proven against real
data: the resolution query for (vonc, footer) → footer-4-column, shared_sites=8 —
the exact durable target 151 fixed by hand. Builds clean; guard tests pass.

Categories: (bug, build, proof, decision, gotcha)

### 2026-07-16 — redesigned fixer deployed; docs rolled forward

v1.0.1123 verified against the pod (symbol table: `patched the DURABLE
content_component template`, `shared_sites`, `chrome_overflow_fix`, `TRANSIENT`).
The redesigned durable-layer fixer is LIVE. Rolled the docs forward for a clean
new-chat start: RUNBOOK rev 45 (this §0), RUNNING_NOTES rev 40 (this entry),
PLAN rev 6→7, STATUS_2026-07-16. Migration errors from the concurrent-workstream
number collisions (two 149s/150s/151s; the gripper 151 blocking the runner) are
being handled in a SEPARATE chat.

Categories: (deploy, docs)

### 2026-07-16 — P3 screenshots-on-failure built (deploy-gated)

The last named polish item from the completed loop, built in one pass. When any
check fails on a (url, profile) run, the adapter photographs the FULL page
while it is still open (so the shot carries post-interaction state) and uploads
it through the same S3/B2 client the imagegenerator uses — key
`acceptance-evidence/<site>/<function>/<run>_<profile>.png` in
`personae-prod-uk001-images`. The response gains `screenshots` refs: durable
`s3://` URI, 7-day presigned `view_url`, and the failing `id@profile` list.

Two design rules worth keeping:
- **Evidence is never load-bearing.** No storage configured, capture error,
  upload error → one log line, verdict untouched. Nav-failed pages are not
  photographed (nothing loaded — no state worth keeping). The "docs never fail
  the work" rule, applied to pixels.
- **Presigned URLs never enter doc_notes.** Notes are loaded into LLM prompt
  contexts by load_doc_context; a signed URL is hundreds of chars of expiring
  signature. The note's new `Evidence:` line carries the durable URI only;
  item specs (improve_tool = all evidence; each chrome_overflow item = its own
  profile's) carry both forms for humans to click. There is a unit test
  pinning this invariant.

Also fixed chassis-wide while in there: the shared `platform/kafka` consumer
ERROR-logged the NORMAL empty-poll `context deadline exceeded` every 10s on
every idle adapter (T14's log-drowning observation). Now `errors.Is`-matched
and logged at debug; real fetch errors still ERROR. And checked: the old
`DEBUGaa` coordinator logging (07-09 handoff Task C) is still in the tree —
left for its own turn, it is a wide sweep across processor.go/agent.go.

Tests: 6 new adapter + 4 new judge, all three touched packages green. Deploy
gate: next chassis image (judge + consumer) + next adapter image (capture) +
re-apply the adapter overlay (new object_storage config + B2 secret env).
Proof after deploy: 087 at a failing tool → `Evidence: s3://…` in the
acceptance-fail note, `screenshots` in the item spec.

Categories: (build, design, gotcha)

### 2026-07-16 — P3 proven live on v1.0.1125; the proof caught a dedup bug (157)

Deploy verified against the pods (chassis: symbols + string constants, sha256'd
copy; adapter: startup log `failure screenshots enabled`, idle ERROR spam gone).
Proof was a T9-style controlled failure on drop-rate (`p3-proof` check injected,
in-DB backup first, everything reverted after). Run 1: two full-page screenshots
stored (desktop 142KB / mobile 443KB), the acceptance-fail note carried the
`Evidence:` line with both durable URIs — **but no improve_tool item appeared**.

**The screenshots feature paid for itself on its first run**: the missing item
was a real bug. `idx_swi_dedup` treated `cancelled` as an OPEN status, so T9's
cancelled 2026-07-12 test ticket still held the `(site_id, item_key)` slot and
the judge's `ON CONFLICT DO NOTHING` insert vanished silently — the exact
opposite of the recorded intent (cooldown queries exclude cancelled; T15 parked
the vonc quiz item as "regenerable", which this index made false). Migration
**157** rebuilt the index with `cancelled` in the closed set (safe: strict
subset predicate). Numbering: 152–156 were taken by other workstreams and sit
PENDING behind the failing gripper-151, so 157 went in OUT OF BAND (`psql -f` +
manual ledger row) — running `--apply` would have dragged five foreign files
in. **Next free number: 158.**

Run 2 (post-157): improve_tool item created WITH `spec.screenshots`; curl of
the presigned view_url → HTTP 200, byte-identical to the adapter log, a real
1170×5457 PNG of the live tool (note: HEAD gives 403 — the presign signs GET).
Cleanup zero-orphans: PLAN restored byte-exact (md5 match), both manufactured
fail notes deleted, ticket cancelled with `result.resolution` (the actual
convention — site_work_items has NO notes/resolution column), tmp table dropped.
Left deliberately: 157 + its pipeline note + ledger row; 4 inert PNGs in B2.

Categories: (proof, bug, fix, gotcha)

### 2026-07-16 — summary doc; tool pipeline → Sonnet 5 (158)

Wrote `SUMMARY_travelling_docs_2026-07-16.md` (journey / position / roadmap).
Model audit: agent models live at `default_config → workflow.steps.<step>
.config.ai_service.model`; the workflow columns are NULL for these agents and
the chassis client default is claude-sonnet-4-6 (`anthropic.go`). All 7 Sonnet
steps in the tool pipeline ran claude-sonnet-4-6; recreate_tool runs
claude-opus-4-6 (64k, deliberate). Migration **158** (snapshots ×4; guard
demands exactly 7 updates) moved the 7 steps to **claude-sonnet-5** — no
rebuild needed (alias pass-through, no temperature sent) and diagnose-agent
had already proven sonnet-5 in prod through this chassis. recreate_tool left
on Opus 4.6; Opus 4.8 and the ~31 other sonnet-4-6 agents are separate calls.
Applied out of band (runner still blocked at gripper-151). Next free: **159**.
Aliases for claude-sonnet-5 / claude-opus-4-8 added to model_aliases.go.

Categories: (docs, build, decision)

### 2026-07-16 — recreate_tool → claude-opus-4-8 (159)

Completes the model refresh. Migration 159 (snapshot taken, guard = exactly 1
row) moved tool-recreation-handler's recreate_tool step from claude-opus-4-6
to claude-opus-4-8; 64k max_tokens and everything else untouched; verified
post-apply. The tool pipeline now runs 7× claude-sonnet-5 + 1× claude-opus-4-8.
Applied out of band (runner still blocked at gripper-151). Next free: **160**.

Categories: (build, decision)

### 2026-07-16 — runner unblocked: gripper-151 was a ledger omission

The "failing" 151 belonged to the empty-sections/loop-integrity workstream,
whose own handoff (§7) says 149–156 were applied — but 151–156 never got
schema_migrations rows, so the runner replayed 151 into its duplicate-component
error and halted there. Verified all six files' artifacts live in the DB
(component / 5 products / page slots / plan section / drift check on
completeness-discovery-agent — not design-discovery, mind — / refresher agent),
then backfilled six ledger rows (applied_by='ledger-backfill'). Runner:
"Up to date — no pending migrations". Their handoff updated.

Durable rule: whoever applies out of band inserts the ledger row themselves —
an applied-but-unrecorded migration turns into a runner roadblock that looks
like someone else's broken SQL.

Categories: (fix, gotcha, cross-workstream)

### 2026-07-17 — new models through the real pipeline; tool-birth gap closed; REAL failure → tool-improver

v1.0.1128. Proved Sonnet 5 through the ACTUAL generator (not just the DB flip):
fired tool-generator for a new tool `tool-loot-table-balancer` on gamesdesign —
llm_call_log confirms generate_tool_html + compose_plan both on claude-sonnet-5,
component + fenced PLAN-at-birth (container present) + index_plan (3 chunks/3
embeddings), zero errors.

**Composer shape defect (fixed at birth).** Sonnet-5's compose_plan wrote the
interaction check as `{"type":"click","expect":"<string>"}` — not a Tier-4
check type, expect must be an object — so the runner skips it and the tool's
behaviour goes untested. The SELECTORS were real (no-invention rule held); only
the shape was improvised, because the compose_plan prompt described interactions
in prose and never showed the JSON. Every prior well-formed interaction check
was hand-written in a migration (143/148) → gap never exercised. Migration 160
adds the exact shape to the prompt; 161 supersedes the already-born PLAN to the
real shape, PROBED passing in live Chromium first (row-4 absent pre-click on
both profiles → expect not vacuous; #ltbAddItem click produces it).

**Tool-birth deploy gap closed (code).** tool-generator never enqueues a
page_rerender (its create_result even carries an unused `needs_rerender:true`),
so a new tool page sits build_status=planned until an unrelated sweep — all 3
births needed a hand-inserted item. Taught `create_rerender_items` a single-page
mode (scalar page_id/name/filename config, tolerates leading "/") so one action
still owns the item shape/status/dedup key; a tool-generator tail step can now
enqueue it. Trap re-learned: create_rerender_items inserts at **triaged**, not
detected (detected sits unswept on gamesdesign) — my first hand-insert guessed
detected and stalled. GATE: tail wiring + Go change ride the next image.

**REAL failure → tool-improver → fix (closes T8/T9 open milestone).**
Acceptance on the corrected PLAN: 8/9 passed (incl. add-item on BOTH profiles —
the reshaped interaction genuinely drove the tool), and mobile-fit@mobile FAILED
on a genuine defect — a fieldset 419px wide at 390px, attributed INSIDE the tool
→ routed to improve_tool with the P3 screenshot. First non-manufactured failure
to reach tool-improver. Promoted it; tool-improver ran on **claude-sonnet-5**
(improve_tool 4767 tok + compose_note 250 tok), constrained the fieldset
(max-width:100% + box-sizing + flex-wrap/min-width:0), wrote a machine `fix`
note, and enqueued its OWN rerender (0743bfa9). Re-verify pends on a large prod page-rerender backlog (83 items) draining — external to this workstream; the fix is durably in the component template and the continuous sweep will re-verify green autonomously once it deploys.

Migrations 160/161 applied out of band (ledger rows same sitting). Next: 162.

Categories: (proof, build, fix, gotcha, milestone)

### 2026-07-17 (later) — re-verify RED twice: the fix loop doesn't converge on intrinsic overflow

The loot fix finally deployed (rerender drained the 83-deep backlog). Re-verify:
STILL RED, `mobile-fit@mobile fieldset 419px`. tool-improver (Sonnet 5) had
constrained the fieldset the adapter NAMED but not the grid child inside
`#ltbRows .ltb-row-grid` that forces the width. A SECOND cycle — Sonnet 5,
loading its own prior fix note — produced a materially identical fix and stayed
RED. The behavioural tier caught an insufficient fix twice (the whole point),
but the loop does NOT converge here.

Two causes: (1) the overflow signal names the widest ANCESTOR (fieldset), so the
one-shot fixer keeps targeting it and never reaches the forcing descendant — the
same class T15 solved for chrome; (2) nothing bounds a non-converging loop (each
fail = a fresh improve_tool item, so the 3-attempt cap never engages; only the
7-day cooldown gates re-tries → weekly re-fail forever). Filed
`bugs_open/010_HANDOFF_2026-07-17_fix_loop_stuck_on_intrinsic_overflow.md`
(candidates: drill-down overflow attribution; convergence guard →
needs_human_review after N cycles). Tool left overflowing as the benchmark.

Corrected an earlier grep guess: the culprit is NOT `.ltb-summary div`
(min-width:140px) — that div isn't inside the flagged fieldset. Pin computed
widths in a browser, not source CSS. Two transient infra notes: a
`request_browser_run` timeout → `complete_error` (re-fire cleanly); and
`AWAITING_RESPONSES` is PLURAL (a poll excluding the singular exits early).

Categories: (finding, loop-limit, gotcha)
