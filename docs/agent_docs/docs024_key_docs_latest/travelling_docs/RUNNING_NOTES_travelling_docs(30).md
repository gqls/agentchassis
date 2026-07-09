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
