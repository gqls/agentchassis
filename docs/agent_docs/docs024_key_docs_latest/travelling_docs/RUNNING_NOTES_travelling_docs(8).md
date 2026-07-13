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
