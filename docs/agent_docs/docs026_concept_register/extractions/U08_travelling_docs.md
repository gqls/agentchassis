# EXTRACTION U08 — docs024_key_docs_latest/travelling_docs (travelling documentation system)
Extracted 2026-07-13. Files in scope: 116. Concepts found: 52.

All paths below are relative to `docs/agent_docs/docs024_key_docs_latest/travelling_docs/`
unless otherwise stated. This unit documents the 2026-07-04 → 2026-07-13 arc that built the
travelling-docs (PLAN + NOTES) system and the tool verification ladder up to a self-driving
headless-browser tier. The RUNNING_NOTES family is append-only (every numbered version is a
byte-prefix of (39), verified by diff), so (39) covers the whole chronology; the RUNBOOK and
PLAN families are edited-in-place, checked for dropped sections (only renames/compressions —
dropped narratives were deliberately moved into RUNNING_NOTES/016b).

## Coverage
| file | treatment |
|---|---|
| 001_README_acceptance_criteria.md | full |
| 016b_debugging_guide_7_3_(2).md | family-delta |
| 016b_debugging_guide_7_3_(3).md | family-delta |
| 016b_debugging_guide_7_3_(4).md | family-delta |
| 016b_debugging_guide_7_3_(5).md | family-delta (one §9 entry "Error containment does not protect against a HANG" dropped in (7) — superseded by the OOMKill correction, see concept below) |
| 016b_debugging_guide_7_3_(6).md | family-delta |
| 016b_debugging_guide_7_3_(7).md | family-latest (full) |
| 084_TRIGGER_diagnose_v1(2).sh | header-scan |
| 085_TRIGGER_toolgen_gamesdesign_v1.sh | header-scan |
| 086_TRIGGER_recreate_economy_simulator.sh | header-scan |
| 086_input_data_recreate_economy_simulator.json | full |
| 087_TRIGGER_tool_acceptance.sh | header-scan |
| 0NN_supersede_xp_curve_plan_selectors.sql | family-delta |
| 0NN_supersede_xp_curve_plan_selectors(1).sql | family-delta |
| 0NN_supersede_xp_curve_plan_selectors(2).sql | family-latest (full) |
| 0NN_wire_persist_diagnosis_note.sql | full |
| FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md | full |
| HANDOFF_2026-07-08_travelling_docs_and_toolgen_bug.md | full |
| HANDOFF_2026-07-09_recreation_and_chassis.md | family-delta |
| HANDOFF_2026-07-09_recreation_and_chassis_1_.md | family-latest (full) |
| HANDOFF_2026-07-10_stage5_live_and_next_fronts.md | full |
| OVERVIEW_self_verifying_tools.md | full |
| PLAN_tool_acceptance_runner(1).md | family-delta |
| PLAN_tool_acceptance_runner(2).md | family-latest (full) |
| PLAN_travelling_docs.md | family-delta (rev 1: abandoned flat-file storage — extracted below) |
| PLAN_travelling_docs(1).md | family-delta |
| PLAN_travelling_docs(2).md | family-delta |
| PLAN_travelling_docs(3).md | family-delta |
| PLAN_travelling_docs(4).md | family-delta |
| PLAN_travelling_docs(5).md | family-delta |
| PLAN_travelling_docs(6).md | family-latest (full) |
| README_summary_paragraph2_for_discussion.md | full |
| README_summary_paragraph_for_handoff.md | full |
| RUNBOOK_travelling_docs.md | family-delta (rev 1: docselect catalogue procedure — extracted below) |
| RUNBOOK_travelling_docs(1).md | family-delta |
| RUNBOOK_travelling_docs(2).md | family-delta |
| RUNBOOK_travelling_docs(3).md | family-delta |
| RUNBOOK_travelling_docs(4).md | family-delta |
| RUNBOOK_travelling_docs(5).md | family-delta |
| RUNBOOK_travelling_docs(6).md | family-delta |
| RUNBOOK_travelling_docs(7).md | family-delta |
| RUNBOOK_travelling_docs(8).md | family-delta |
| RUNBOOK_travelling_docs(9).md | family-delta |
| RUNBOOK_travelling_docs(10).md | family-delta |
| RUNBOOK_travelling_docs(11).md | family-delta |
| RUNBOOK_travelling_docs(12).md | family-delta |
| RUNBOOK_travelling_docs(13).md | family-delta |
| RUNBOOK_travelling_docs(14).md | family-delta |
| RUNBOOK_travelling_docs(15).md | family-delta |
| RUNBOOK_travelling_docs(16).md | family-delta |
| RUNBOOK_travelling_docs(17).md | family-delta |
| RUNBOOK_travelling_docs(18).md | family-delta |
| RUNBOOK_travelling_docs(19).md | family-delta |
| RUNBOOK_travelling_docs(20).md | family-delta |
| RUNBOOK_travelling_docs(21).md | family-delta |
| RUNBOOK_travelling_docs(22).md | family-delta |
| RUNBOOK_travelling_docs(23).md | family-delta |
| RUNBOOK_travelling_docs(24).md | family-delta (duplicate of (23) by Last-updated line) |
| RUNBOOK_travelling_docs(25).md | family-delta |
| RUNBOOK_travelling_docs(26).md | family-delta |
| RUNBOOK_travelling_docs(27).md | family-delta |
| RUNBOOK_travelling_docs(28).md | family-delta |
| RUNBOOK_travelling_docs(29).md | family-delta |
| RUNBOOK_travelling_docs(30).md | family-delta |
| RUNBOOK_travelling_docs(31).md | family-delta |
| RUNBOOK_travelling_docs(32).md | family-delta |
| RUNBOOK_travelling_docs(33).md | family-delta |
| RUNBOOK_travelling_docs(34).md | family-delta |
| RUNBOOK_travelling_docs(35).md | family-delta |
| RUNBOOK_travelling_docs(36).md | family-delta |
| RUNBOOK_travelling_docs(37).md | family-delta |
| RUNBOOK_travelling_docs(38).md | family-latest (full) |
| RUNNING_NOTES_travelling_docs.md | family-delta (rev-1-only wording; content preserved compressed in (39)) |
| RUNNING_NOTES_travelling_docs(1).md | family-delta (verified byte-prefix of (39)) |
| RUNNING_NOTES_travelling_docs(2).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(3).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(4).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(5).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(6).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(7).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(8).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(9).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(10).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(11).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(12).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(13).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(14).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(15).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(16).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(17).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(18).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(19).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(20).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(21).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(22).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(23).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(24).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(25).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(26).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(27).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(28).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(29).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(30).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(31).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(32).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(33).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(34).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(35).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(36).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(37).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(38).md | family-delta (prefix) |
| RUNNING_NOTES_travelling_docs(39).md | family-latest (full) |
| bundle_recreation_v1.sh | family-delta |
| bundle_recreation_v1(1).sh | family-latest (header-scan) |
| verify_before_migration.sql | full |
| write_doc_plan_action.go | header-scan |

## Concepts

### Travelling documentation (PLAN + NOTES) in Postgres
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(38) §1 "Truth (Postgres, LIVE): doc_plans … doc_notes"; migration applied 2026-07-04 (Stage 1 ✅, statement tally verified); PLAN(6) rev 5 "Phase A write/read hooks proven in production".
- **what:** Every tool/complex component and every pipeline carries its own living documentation in two Postgres tables: a PLAN (intent — aim, source spec, behaviour contract, acceptance criteria, delivery mechanism, dependencies, deliberate decisions) and a NOTES stream (every fix, diagnosis, and dead end). Agents write these as a byproduct of the steps that create and fix things, and load them before touching a subject, so fixes build on prior decisions instead of re-deriving lost context. Solves two failure modes: lost intent, and "deployed ≠ works".
- **sources:** RUNBOOK_travelling_docs(38).md#intro,§1; PLAN_travelling_docs(6).md#aim; OVERVIEW_self_verifying_tools.md#mechanism-1; RUNNING_NOTES_travelling_docs(39).md#rev5
- **relations:** tool-doc header system (019, extended not replaced); verification ladder; doc_plans supersede versioning; doc_notes append-only log.
- **verify-later:** tables `doc_plans`, `doc_notes` in clients_db; migration `sql_for_agents/125*` (arc renumbered 125–146); actions in `platform/orchestration/actions/write_doc_plan_action.go` etc.

### doc_plans supersede versioning (one current row, never edit history)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(38) §2 "supersede tx; one current row enforced by idx_doc_plans_current"; xp-curve PLAN chains v1→v2→v3 verified 2026-07-10.
- **what:** PLAN updates flip the current row to `is_current=false` + `superseded_at` and insert the new body as current; a partial unique index enforces exactly one current row per subject. History is never edited in place; rollback = restore a prior row; `pinned=true` is a human hold. The pattern is the `site_specs` supersede log re-keyed to the doc subject.
- **sources:** RUNBOOK_travelling_docs(38).md#§2; RUNNING_NOTES_travelling_docs(39).md#rev2 (supersede-log pattern confirmed); 0NN_supersede_xp_curve_plan_selectors(2).sql (live example); write_doc_plan_action.go (header)
- **relations:** site_specs supersede log (pattern source); EDIT-marker fill-by-supersede convention.
- **verify-later:** `idx_doc_plans_current` partial unique index; `write_doc_plan` supersede transaction in the action.

### doc_notes append-only log with jsonb category roll-up
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES(39) rev 2 "NOTES = table … Postgres serialises concurrent inserts"; GIN roll-up query in RUNBOOK §0-REF/§7.
- **what:** NOTES are one DB row per entry (never a shared file — a file append is a read-modify-write with lost-update risk under the retry-less git adapter). `categories jsonb` with a GIN `jsonb_ops` index makes `categories ? 'tag'` cross-tool roll-ups index-backed. `site_id` scopes per-site incidents. Entry format is uniform and dated (Observed / Root cause / Fix / Verified / Categories).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev2-GIN-RMW; PLAN_travelling_docs(6).md#table-design,#document-formats; RUNBOOK_travelling_docs(38).md#§3
- **relations:** NOTES category taxonomy; git-adapter constraints (why not git).
- **verify-later:** `doc_notes` schema + GIN index; roll-up queries in 016/016b.

### DB-as-truth storage decision (knowledge_base = derived index; git = optional mirror)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES(39) rev 2 "DECISION: DB is the source of truth; git is an optional mirror"; PLAN(6) "Unchanged from rev 2".
- **what:** Postgres tables written transactionally by the framework are the record of truth; `knowledge_base` (content-hash keyed, no version chain) is only a derived retrieval index via `rag_index`/`rag_lookup`; git is a non-authoritative optional mirror for human browsing (Phase B, unbuilt). Grounded in git-adapter evidence: commits hard-reject empty Domain, force-prefix `{domain}/`, whole-file only, no read action, no conflict retry, all serialised through one Kafka adapter.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev2; PLAN_travelling_docs(6).md#storage-decision; RUNBOOK_travelling_docs(38).md#§1
- **relations:** supersedes the flat-file docs-repo decision (below); rag_index tool_docs collection.
- **verify-later:** `adapter.go`/`github_client.go` commit path; `knowledge_base` UNIQUE(collection, content_hash).

### Abandoned: flat-file docs-repo as truth + docselect catalogue retrieval
- **category:** documentation-system
- **status-signal:** superseded
- **status-evidence:** RUNNING_NOTES(39) rev 1 "Rev-1 storage decision (later reversed) … Categories: (storage, superseded)"; RUNBOOK rev-1 §4 "Making a doc retrievable (catalogue entry)" section vanishes from later revisions.
- **what:** The original (2026-07-04 rev 1) design: flat markdown files (`<docs-repo>/tools/<function>/PLAN.md`, `NOTES.md`) in a new writable docs repo as source of truth, RAG-indexed, plus a `DocRule` entry per tool in `diagnose_doc_catalogue*.json` so the code-diagnosis loop's `docselect.go` picks docs by keyword/path-glob. Reversed to DB-as-truth within the same day; the docselect route remains deferred for pipelines only ("needs the git mirror for files — Phase B").
- **sources:** PLAN_travelling_docs.md#design-decision-2 (rev 1); RUNBOOK_travelling_docs.md#§4 (rev 1); RUNNING_NOTES_travelling_docs(39).md#rev1
- **relations:** DB-as-truth decision (replacement); pipeline retrieval symmetry (docselect kept as a Phase-B idea for pipelines).
- **verify-later:** `docselect.go` DocRule selection; whether any doc catalogue entry for tool docs ever landed (expect none).

### Doc subject convention — ('tool', function) and ('pipeline', build|content|design|maintenance)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Stage 0 gates 2026-07-04: live pipeline values build (3579), content (24), design (13), maintenance (2), no CHECK constraint; RUNBOOK §1.
- **what:** Docs key on `(subject_type, subject_key)`: tools by `content_components.function` byte-for-byte (survives forks — vindicated by the unique-index predicate covering active library originals only), pipelines by the four live `site_work_items.pipeline` values (convention, not schema — the column is unconstrained text). Generalising from tool_doc_* to subject-keyed tables was a deliberate rename made "while the migration was free".
- **sources:** RUNBOOK_travelling_docs(38).md#§1,#stage-0; RUNNING_NOTES_travelling_docs(39).md#rev3 (PROPOSED: generalise to subjects); verify_before_migration.sql
- **relations:** dangling-doc prevention rule; idx_cc_tool_function_unique.
- **verify-later:** `site_work_items.pipeline` live values; `content_components.function` uniqueness predicate.

### The dangling-doc prevention rule — subject must be something the agent actually owns
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Migration 137 applied 2026-07-09/10 ("spec declared … append_note re-subjected ('tool', spec.function) → ('pipeline','build')"); PLAN(6) "Rollout outcomes" first bullet.
- **what:** A NOTES subject must reference an artifact the writing agent actually creates/owns. `tool-recreation-handler` writes page sections and never creates a `content_components` row, so a `('tool', spec.function)` note there would key a doc to a function no component owns — a dangling doc. It was re-subjected to `('pipeline','build')` + site stamp, mirroring component-template-fixer. Found by reading the definition, not by a failed run.
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev42-blocker-ii; HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§3
- **relations:** recreation-writes-page-sections fact; migration 137.
- **verify-later:** tool-recreation-handler `append_note` config in agent_definitions.

### The four doc actions (write_doc_plan, append_doc_note, load_doc_context, persist_diagnosis_note)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNBOOK Stage 2 "✅ Go actions — ON PRODUCTION 2026-07-04 … deployed with their registry entries".
- **what:** The chassis-side write/read surface of travelling docs: `write_doc_plan` (supersede tx), `append_doc_note` (single INSERT), `load_doc_context` (current PLAN + latest-N NOTES + extracted `criteria_json`, composed as one prompt-ready block; `has_plan=false` is a normal state, not an error), `persist_diagnosis_note` (diagnosis output → NOTES). Conventions: prefixed InputSpec field names, error containment via `config.error_step`, pure-helper unit tests (`doc_actions_helpers_test.go`).
- **sources:** RUNBOOK_travelling_docs(38).md#stage-2,#§6; RUNNING_NOTES_travelling_docs(39).md#rev4-drafts,#rev9; write_doc_plan_action.go
- **relations:** all write hooks below; criteria fenced block.
- **verify-later:** `platform/orchestration/actions/{write_doc_plan,append_doc_note,load_doc_context,persist_diagnosis_note}_action.go` + registry.go entries.

### persist_diagnosis_note — skip-don't-guess subject gate; dead ends persisted
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Stage 3a CLOSED 2026-07-06 (skip gate proven ×3); Stage 3b CLOSED 2026-07-06/07 — first machine-written NOTES row `('pipeline','build')`, categories `["diagnosis","unconfirmed-diagnosis"]`, stop reason `scope-not-narrowing`.
- **what:** A config-gated step after `diagnose_emit` (emit stays read-only by its own design) that persists the diagnosis as a NOTES entry ONLY when the run carries an explicit subject in input_data — skip, never guess (a mis-filed note poisons history; the gate is the action's first check, before any DB access). UNVERIFIABLE verdicts are persisted too, tagged `unconfirmed-diagnosis`, so dead ends stop retries. First payoff on record: the machine-written note itself answered the open "why did the run finish fast" question.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-3,#§4; 0NN_wire_persist_diagnosis_note.sql; RUNNING_NOTES_travelling_docs(39).md#rev3,#rev20,#rev21
- **relations:** subject threading (3b); anchorless-diagnosis degrade; 037 pipeline-integration vision (realised).
- **verify-later:** diagnose-agent workflow `emit → persist_note → complete`; `persist_diagnosis_note_action.go` subject gate.

### Diagnosis subject threading through orchestrator input_mapping + both contracts
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** 3b.2 APPLIED + 3b.3 VERIFIED 2026-07-06 (map paths `input_data.subject_type`/`subject_key`; both contracts t/t).
- **what:** For a spawned child to receive optional fields, the mapping must satisfy the callee's input_contract — so threading `subject_type?`/`subject_key?` took THREE edits (orchestrator input_mapping merge + `optional` additions on BOTH diagnose-orchestrator and diagnose-agent contracts), not two. DB-only, effective immediately. Establishes the general spawn+call contract rule: an input the workflow depends on must be declared.
- **sources:** RUNBOOK_travelling_docs(38).md#3b; RUNNING_NOTES_travelling_docs(39).md#rev17; HANDOFF_2026-07-08…md#§2
- **relations:** spawn+call input-shape pattern (016b); dangling-doc rule (same "declare your inputs" class — migration 137's `spec` declaration).
- **verify-later:** diagnose-orchestrator `call_diagnoser.input_mapping`; both input_contracts.

### PLAN-at-birth write hook in tool-generator (compose_plan → write_plan → index_plan)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** TASK 3 APPLIED 2026-07-07, PROVEN 2026-07-09 (run 1923badd: system wrote its own first PLAN, real selectors, fence intact, 2,982 chars).
- **what:** After `save_tool` succeeds, a Sonnet step composes a full PLAN body (standard checks verbatim; an interaction check ONLY from real selectors copied out of the generated HTML, never invented; ≤3,000 chars), `write_doc_plan` persists it (`source='tool-generator'`), and `rag_index` indexes it into `tool_docs`. Every doc step carries `config.error_step: "complete"` — docs can never fail tool creation. Composer later fixed by migration 144 (five → four standard checks, inline delivery).
- **sources:** RUNBOOK_travelling_docs(38).md#task-3,#task-3-proven; RUNNING_NOTES_travelling_docs(39).md#rev24,#rev33,#rev34; HANDOFF_2026-07-10…md#§1
- **relations:** docs-never-fail-the-work containment; composer selector invention incident; delivered-reality principle (144).
- **verify-later:** tool-generator workflow (save_tool → compose_plan → write_plan → index_plan); doc_plans rows with source='tool-generator'.

### NOTES-at-every-fix hook on the three fix agents
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** TASK 4 APPLIED 2026-07-07, PROVEN 2026-07-09 (two machine-written `fix` notes from the economy-simulator recreation, 19:36:35Z and 20:33:04Z).
- **what:** `component-template-fixer`, `tool-improver`, and `tool-recreation-handler` each gained `compose_note → append_note` on their SUCCESS paths only (both fixer branches covered), with error containment to the terminal step. Subjects per agent: fixer → pipeline/build + site stamp; improver → tool/`tool_data.function`; recreation → re-subjected to pipeline/build (migration 137). Machine categories v1 = `["fix"]`; failure-class tags live in the body Categories line.
- **sources:** RUNBOOK_travelling_docs(38).md#task-4,#task-5-closed; RUNNING_NOTES_travelling_docs(39).md#rev26,#rev27,#rev45
- **relations:** dangling-doc rule; acceptance iteration loop (fixer loads PLAN+NOTES first).
- **verify-later:** the three agent workflows' compose_note/append_note tails; `doc_notes WHERE categories ? 'fix'`.

### "Docs never fail the work" containment principle — and its limit
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Every doc step wired with `config.error_step` to the terminal step; corollary on record (HANDOFF 07-09 §5): "containment covers errors, not crashes or stalls".
- **what:** Documentation writes are strictly subordinate to the work: any doc-step failure routes to the workflow's terminal step so a fix/creation never fails because its documentation did. The limit was learned live: error containment protects against raised errors only — an OOMKilled pod or a stall raises nothing, so the step freezes instead of degrading (the index_plan incident).
- **sources:** HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§5; RUNNING_NOTES_travelling_docs(39).md#rev31,#rev33; 016b_debugging_guide_7_3_(5).md#§9 (superseded HANG entry)
- **relations:** error_step-in-config mechanics; EXECUTING_STEP-forever pattern.
- **verify-later:** config.error_step on all doc steps in the touched workflows.

### Pipeline documentation model — derive the topology, author the intent
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** PLAN(6) Phase B item 9 "Pipeline PLAN bodies distilled from 004–008" still pending; the migration write-hook IS live ("live practice from 140 onward, not aspiration").
- **what:** A pipeline's step map is generated from `agent_definitions` (callgraph pattern) — never hand-drawn, so it can't drift. The authored pipeline PLAN holds only: end-to-end invariants (e.g. "interactive sections survive every rebuild route"), branch rationale, seams (pipelines sharing one handler is where seam bugs live), and deliberate decisions. Pipeline NOTES = incidents + migration entries + persisted diagnoses; 016/016b stays the global roll-up.
- **sources:** PLAN_travelling_docs(6).md#pipeline-documentation; RUNNING_NOTES_travelling_docs(39).md#rev3; RUNBOOK_travelling_docs(38).md#§2 ("Never embed the step map")
- **relations:** migration write-hook; docselect Phase-B retrieval for pipelines; docs 004–008 as prose base.
- **verify-later:** whether any pipeline PLAN body exists in doc_plans (`subject_type='pipeline'` with is_current).

### Workflow-altering migrations write pipeline NOTES
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "Workflow-altering migrations now leave the runbook-§3 pipeline note (140 and 141 both carry one)" — RUNBOOK rev 44; 145/146 also carry pipeline notes.
- **what:** Every migration that alters an agent workflow appends a `('pipeline','build')` doc_notes entry recording the migration number, what changed, and why — making the migration ledger part of the pipeline's travelling history. The 005 "SQL Migrations Applied" table was identified as the embryo of this write hook.
- **sources:** RUNBOOK_travelling_docs(38).md#task-5-closed (migrations system); PLAN_travelling_docs(6).md#rollout-outcomes,#write-hooks; RUNNING_NOTES_travelling_docs(39).md#2026-07-10-migrations
- **relations:** migrations system (ledger + runner); doc_notes `migration` category.
- **verify-later:** doc_notes rows with categories containing 'migration' from 140 onward.

### NOTES category taxonomy
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(38) §3 lists the taxonomy as operating vocabulary.
- **what:** The tag set for NOTES roll-ups: `css-variable-mismatch`, `empty-shell`/`mode-b-template`, `broken-template-slots`, `content-vs-runtime-mismatch`, `detool-on-rebuild`, `js-not-extracted`, `js-bundle-stale`, `schema-template-drift`, `diagnosis`, `unconfirmed-diagnosis`, `migration`, `seam`, `acceptance-run`, `acceptance-fail`, `truncated-output`, `needs_criteria`. Extends 037's taxonomy; GIN-queryable.
- **sources:** RUNBOOK_travelling_docs(38).md#§3; PLAN_travelling_docs(6).md#document-formats
- **relations:** doc_notes jsonb roll-up; 037 documentation-system conventions.
- **verify-later:** live distinct categories in doc_notes.

### Deliberate-decisions sections + the graduation rule (prose → structured → enforced)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** PLAN(6) "Graduation rule: prose → structured → enforced, only when recurrence proves the need. (Locks stay deferred; criteria live as a fenced block, not a column, until a checker consumes them at volume.)"
- **what:** A PLAN carries a "Deliberate decisions — do not re-fix" prose section, protective because it is loaded at fix time; no lock mechanism yet. Knowledge graduates from prose to structure to enforcement only on demonstrated recurrence — the reason criteria are a fenced block rather than a column, and locks are deferred. Runbook prose is "un-compiled residue" that retires as it is compiled into guards/fixes.
- **sources:** PLAN_travelling_docs(6).md#framing,#deliberate-decisions; RUNNING_NOTES_travelling_docs(39).md#rev3-framing
- **relations:** framing concept below; locks category (031).
- **verify-later:** whether any lock/enforcement mechanism for deliberate decisions has since appeared.

### Framing: plan = enforced desired state; pipeline = compiled runbook; NOTES = the reasoning log
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "Framing (agreed 2026-07-04)" section, stable across PLAN revs 3–6.
- **what:** Where each artifact sits: `site_plans`+specs are the ENFORCED plan (reconciler drives realised state toward it — "the plan table is ground truth; the rest is weather"); the build pipeline is the compiled happy-path runbook; written runbooks are the un-compiled residue (exception knowledge); NOTES is the reasoning log nothing machine-side captures; contracts/constitution sit above as admission rules.
- **sources:** PLAN_travelling_docs(6).md#framing; RUNNING_NOTES_travelling_docs(39).md#rev3
- **relations:** site-plan-and-reconciler (030); graduation rule; contracts-and-standards.
- **verify-later:** n/a (conceptual framing) — cross-check with docs 030/016b claims in their units.

### load_doc_context fix-time retrieval
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Deployed 2026-07-04 (Stage 2); consumed live by the Tier-2 checker and tool-acceptance-agent.
- **what:** The primary direct-by-key read: current PLAN + latest-N NOTES + extracted `criteria_json`, composed into one prompt-ready `doc_context` block. `has_plan=false` is a normal state. For the code-diagnosis loop, `doc_context` is to be handed to `diagnose_assemble_bundle` the way `runtime_evidence` is (one compose line) — `rag_lookup` is discovery-only (no function filter).
- **sources:** RUNBOOK_travelling_docs(38).md#§6; PLAN_travelling_docs(6).md#retrieval; RUNNING_NOTES_travelling_docs(39).md#rev2 (rag signatures grounded)
- **relations:** four doc actions; criteria fenced block; diagnose_assemble_bundle injection (still unwired — verify).
- **verify-later:** `load_doc_context_action.go`; whether the diagnosis bundle now includes doc_context.

### tool_docs knowledge-base indexing of PLANs (rag_index derived index)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** TASK 6 CLOSED 2026-07-10: run 05d1fc97 wrote the FIRST `knowledge_base` `collection='tool_docs'` rows (4 chunks, 4 embeddings, ~5.5s) after the chunkContent fix + migration 141 re-enable.
- **what:** After each PLAN write, `rag_index` chunks and embeds the body into the `tool_docs` collection for semantic discovery. The 019 claim that generation already wrote tool_docs was verified UNIMPLEMENTED (a standing open from day 1); the write first became real 2026-07-10. Standing open: `rag_index` hardcodes `source_type='scrape'` (parameterisation open item).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev1-open-thread,#task-6-closed; RUNBOOK_travelling_docs(38).md#position-2026-07-10; PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** chunkContent infinite-loop bug; DB-as-truth (KB is derived); rag_lookup discovery.
- **verify-later:** knowledge_base rows collection='tool_docs'; `rag_actions.go` source_type parameter.

### EDIT-marker / -EDIT check-id convention (honest unknowns in seeded docs)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(38) mini-glossary; Tier-2 checker implements "-EDIT skipped" (unit-tested).
- **what:** Fill-in-the-blank markers for details not known at seeding time: `EDIT:` prose markers are optional fill-later blanks (doc valid meanwhile; fills arrive by supersede, never in-place edits); acceptance checks whose id ends `-EDIT` carry placeholder selectors and are skipped by every verification tier until real selectors replace them.
- **sources:** RUNBOOK_travelling_docs(38).md#mini-glossary; RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23
- **relations:** pilot PLAN seeding; anchor rule (fail ⇒ drop or mark -EDIT).
- **verify-later:** -EDIT handling in `discovery_checks/check_tool_acceptance.go` and the browser-runner.

### Pilot PLAN seeding by SQL (dogfooding the format)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** TASK 2A DONE 2026-07-07 12:32 — first tool PLAN live for `tool-archetype-taster-quiz` (is_current=t, has_fence=t, 2,761 chars).
- **what:** Before any workflow wiring existed, the first real tool PLAN was seeded by a hand-written dollar-quoted INSERT (source='human', created_by='pilot'), satisfying Stage-5's precondition (≥1 tool PLAN with criteria) and road-testing the section format. Later `write_doc_plan` calls supersede it cleanly. Includes a seeded deliberate decision ("exactly THREE questions — the taster must not be improved into the Gauntlet").
- **sources:** RUNBOOK_travelling_docs(38).md#pilot-plan; RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23
- **relations:** EDIT markers; acceptance-criteria-in-PLAN decision.
- **verify-later:** doc_plans row for tool-archetype-taster-quiz (superseded chain).

### Acceptance criteria live in the tool's PLAN (fenced ```criteria JSON block)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECISION: acceptance criteria live in the tool's doc_plans PLAN" (2026-07-04 rev 4); consumed live by Tier 2 and Tier 4 checkers.
- **what:** The per-tool definition of *working*. Candidates judged on key/lifecycle/owner: `site_specs` — right machinery, wrong key (site-scoped; per-site copies drift); `site_plans`/directives — wrong lifecycle and owner (churniest artifact, planner-owned; "never store the bar in the artifact that regenerates most"); findings' `acceptance_test` — right pattern, wrong duration (dies with the work item; the standing criteria SEED it). The PLAN wins on all three axes. Format: a machine-extractable fenced ```criteria JSON block (tool-doc-header precedent), extracted by `load_doc_context` as `criteria_json`; lifts to a column only on volume. Per-site parametrisation goes to `direction.must_have`, not the PLAN.
- **sources:** PLAN_travelling_docs(6).md#where-acceptance-criteria-live; 001_README_acceptance_criteria.md; RUNNING_NOTES_travelling_docs(39).md#rev4
- **relations:** verification ladder; findings acceptance_test/max_fix_attempts (improvement-loop 004); direction.must_have.
- **verify-later:** criteria fence extraction in load_doc_context; `has_fence` on live PLANs.

### Criteria describe DELIVERED reality, not aspiration (Option B)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECIDED 2026-07-10: Option B — inline reality (user: 'I choose option B and surrender')"; migrations 143 (PLANs superseded) + 144 (composer fixed) applied and verified.
- **what:** The composer had asserted a designed-but-never-built JS extraction (`asset_loads /tools/assets/<fn>.js`) in every PLAN; Tier-2's first sweep failed every tool on it by construction. Principle on record: criteria must describe what the system delivers; aspirations live in roadmaps. If extraction ever ships, PLANs supersede forward again. Corollary: the composer's standard checks became boots/console/status/mobile-fit + optional interaction from real selectors.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5 (Option B block); HANDOFF_2026-07-10…md#§2; PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** js-not-extracted delivery gap; Tier-2 first sweep; PLAN supersede versioning.
- **verify-later:** current PLAN fences have no asset check; compose_plan prompt (four checks, inline delivery line).

### The tool verification ladder (Tiers 0–4)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "The verification ladder is whole (Tier 0/1/2/4) and closes on both outcomes" — RUNBOOK position 2026-07-12; Tier 3 remains Phase B.
- **what:** Cheap-to-expensive tiers, each catching a different class: Tier 0 generation-time output integrity (`HasToolDocHeader` gate + `check_tool_completeness`, deliberately flags-but-passes); Tier 1 structural post-deploy (`check_tool_health`); Tier 2 static contract-presence against deployed HTML (anchor rule); Tier 3 acceptance audit (`tool-auditor` vs criteria — Phase B, unbuilt extension); Tier 4 behavioural — drive the deployed tool in headless Chromium until criteria pass. Standing rule: never read a Tier-2 pass as "the tool works" — that claim belongs to Tier 4.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_travelling_docs(6).md#tool-assurance; OVERVIEW_self_verifying_tools.md#mechanism-2
- **relations:** every tier concept below; "passed checks ≠ working".
- **verify-later:** check_tool_completeness + check_tool_health + discovery_checks/check_tool_acceptance.go + browser-runner adapter, all in the chassis repo.

### "Completeness + validation passed" ≠ working — twice demonstrated
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN(6) rollout outcomes: "The June recreation introduced the economy-simulator's two bugs and passed; run 2 of the repair faithfully recreated them and passed."
- **what:** The standing empirical argument for the behavioural tier: structural/validation checks measure output integrity, not behaviour. The same game shipped broken twice while passing every existing check — the June 2026-06-05 recreation introduced the bugs (proven from tool_recreation_training rows and the origin game.js which has neither bug), and repair run 2 recreated them while its own note truthfully said "completeness + validation passed".
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45; OVERVIEW_self_verifying_tools.md#problem
- **relations:** Tier 4; seam rule; economy-simulator case.
- **verify-later:** tool_recreation_training rows for page d9a8e6e8 dated 2026-06-05.

### Tier-2 static acceptance checker (discovery check `tool_acceptance`)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Stage 5 — LIVE 2026-07-10 ✓ (first sweep proven)" — run cd0d9731 on v1.0.1107 produced exactly the pre-verified findings (2 improve_tool items + 2 acceptance-fail notes, check-level precision confirmed).
- **what:** A browserless discovery check (sibling of `tool_health` in `discovery_checks/`): loads the current PLAN's criteria fence, fetches the deployed page (bounded 12s/2MB, cached per run), and evaluates the statically-visible subset under the anchor rule, plus shell checks (tool-doc header not leaked, no `<no value>` residue). No criteria → a `needs_criteria` note (30-day cooldown), never a fake pass. Failures → one improve_tool item (criteria embedded as `acceptance_test`, 7-day cooldown, cancelled items excluded since migration 146's correct-while-touching) + an acceptance-fail note. Scope limit by construction: only generator-created tools have content_components rows; adopted/recreated page-section tools are invisible to Tier 2.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5; RUNNING_NOTES_travelling_docs(39).md#stage-5-built,#stage-5-live; HANDOFF_2026-07-10…md#§1,§2
- **relations:** anchor rule; migration 142; Tier 4 (reaches page-section tools via pages).
- **verify-later:** `discovery_checks/check_tool_acceptance.go`; design-discovery-agent run_checks list.

### The anchor rule — static checks confirm, never refute
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "STAGE-5 RULE SETTLED" 2026-07-09 after the #tableWrap inspection (empty div filled by JS); implemented + unit-tested incl. the founding cases.
- **what:** Validate only a criteria selector's ANCHOR (leftmost id/class token) against `html_template`: `#tableWrap` exists ⇒ `#tableWrap tr` passes (rows are JS-built; Tier 4 asserts them for real); `#xpTableBody` exists nowhere ⇒ fails ⇒ drop or -EDIT. Static validation can confirm a selector but never refute one — never delete a check merely because the DOM is constructed at runtime. Motivated by the composer inventing selectors it ASSERTS on while copying real ones it ACTS on; the remedy is a check made by the system on itself, not a sterner prompt. Implementation detail banked: CSS class tokens are whitespace-delimited (Go regexp `\b` wrongly splits on hyphens).
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5-rule; RUNNING_NOTES_travelling_docs(39).md#rev39,#rev40; OVERVIEW_self_verifying_tools.md#tier-2
- **relations:** composer selector-invention incident; Tier 4 runtime assertions; tool-auditor (same logic belongs there — unbuilt).
- **verify-later:** anchor extraction + class-token comparison in check_tool_acceptance.go tests.

### Composer selector invention — caught twice, machine-corrected
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** CONFIRMED 2026-07-09 (`#xpTableBody`/`#statsStrip` token_anywhere=f); second sighting caught by Tier 2 itself 2026-07-10 (kebab `#drop-chance` vs real camelCase `#dropChance`).
- **what:** The PLAN-composer LLM invented DOM ids for assertion targets despite an explicit "never invent a selector" instruction — the rule held for controls it acts on and failed for things it asserts on. First instance corrected by a guarded supersede migration that itself initially refused a valid runtime selector (leading to the anchor rule); second caught automatically by the live Tier-2 sweep and corrected by migration 143. Demonstrates the design stance: hallucination is countered by verification, not prompt escalation.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#stage-5-live; 0NN_supersede_xp_curve_plan_selectors(2).sql; RUNBOOK_travelling_docs(38).md#task-3-proven
- **relations:** anchor rule; supersede versioning (correction recorded as a NOTES entry — "the travelling-docs loop applied to itself").
- **verify-later:** xp-curve PLAN v1→v2 chain + its correction note in doc_notes.

### Tier-4 browser-runner adapter (headless Chromium over Kafka)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Stage 6 P0 DEPLOYED + SMOKE PASSED" 2026-07-11 (v1.0.1107; §2.15 smoke matched manual inspection; real bool headers; in_response_to_request_id matcher).
- **what:** A dedicated adapter deployment (image = debian-slim + Playwright + Chromium, playwright-go) consuming `system.adapter.browser-runner.requests` (035 Convention A) and replying on the caller's topic with `{results:[{check_id, profile, url, pass, detail}]}`. P0 scope: desktop 1366×900, three check types (`page_status_ok`, `selector_exists` asserted against the LIVE DOM after settle, `no_console_errors`); everything else honestly reported in `skipped[]`, never faked; browser launched per request so a crash poisons one run, not the pod; navigation failure is a check-fail, not an infra error. Contract pinned to the 035 Adapter Guide as normative after a compliance pass (typed header struct with real bools; `in_response_to_request_id` = incoming request_id is THE matcher; ProduceWithValidation never plain Produce). Build gotchas banked: playwright.azureedge.net CDN dead; v0.6100.0 must be required under its declared (pre-rename) module path; the driver ignores XDG_CACHE_HOME — set HOME in the image.
- **sources:** PLAN_tool_acceptance_runner(2).md (whole); RUNBOOK_travelling_docs(38).md#stage-6; RUNNING_NOTES_travelling_docs(39).md#stage-6-built,#2026-07-11; HANDOFF_2026-07-10…md#T3–T6
- **relations:** analyser-adapter mould (pattern source); 035 adapter guide; tool-acceptance-agent (caller).
- **verify-later:** `cmd/browser-runner-adapter/main.go`; `internal/adapters/browserrunner/`; KafkaTopic CR; dockerfile HOME=/pw-home.

### tool-acceptance-agent — Tier 4 self-driving orchestrator
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** First machine acceptance-run note (run bf330ac6, 2026-07-12, "Tier-4 acceptance PASSED — all 3 evaluated checks"); fail path proven live via a controlled reverted test (failed=1, improve_tool_created=true, full teardown verified).
- **what:** An agent (migration 145) closing the loop with zero humans: `ensure_site_record → load_docs → request_browser_run (Kafka await; resolves the tool's deployed URL from pages itself; NO-OP skips without awaiting when the PLAN has no criteria) → judge_acceptance_results → complete`. Judge recomputes the verdict from results: all pass → acceptance-run note; any fail → acceptance-fail note + ONE improve_tool item (criteria embedded as acceptance_test, handler tool-improver); component-less recreated/adopted tools get the note but no item — logged honestly for manual routing. Trigger 087 (dry-run default).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tool-acceptance-agent-built,#tier-4-self-driving,#fail-path-proven; README_summary_paragraph2_for_discussion.md; 087_TRIGGER_tool_acceptance.sh (header)
- **relations:** browser-runner adapter; acceptance iteration loop; continuous sweep.
- **verify-later:** `platform/orchestration/actions/tool_acceptance_actions.go`; agent_definitions row tool-acceptance-agent; migration 145.

### Continuous acceptance — the `tool_acceptance_due` periodic sweep
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Built + migration 146 applied 2026-07-12, but "v1.0.1111 … the continuous sweep is NOT in the binary" (untracked-file trap); "GATE: continuous acceptance activates on the next image built from 83ba9bd4+" (T11, 2026-07-13 — state at unit close).
- **what:** A discovery check that emits one `acceptance_run` work item per active tool with a deployed page and current PLAN criteria, unless a verdict landed within 7 days or a run is open. Design calls: post-creation/post-improve hooks deliberately NOT used (they'd fire before the page redeploys — creation ends at 'planned', improve merely queues a rerender; the sweep only ever sees deployed pages); items emitted straight to `triaged` (acceptance needs no human judgment; `detected` items were observed sitting unswept); priority 90 so acceptance tests the NEW page after builds/rerenders.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tier-4-continuous,#v1.0.1111; HANDOFF_2026-07-10…md#T10,T11; README_summary_paragraph2_for_discussion.md
- **relations:** tool-acceptance-agent; untracked-file deploy trap; improve_tool cooldown (cancelled items excluded).
- **verify-later:** `discovery_checks/check_tool_acceptance_due.go` in the deployed image; first unattended acceptance-run note.

### Acceptance iteration loop — iterate until criteria pass
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Both halves proven separately (fail path via controlled test 2026-07-12; fix agents write notes); "let a REAL failure flow through to tool-improver and back" still open at unit close.
- **what:** deploy → acceptance run → failing criterion → `improve_tool` item (criterion as `acceptance_test`, bounded by `max_fix_attempts`) → fixer loads PLAN+NOTES first → fix → append note → redeploy → re-run. Criteria hold the bar still across iterations; NOTES stop iterations fighting each other. *Working* = criteria pass. The one link proven only with a synthetic input is a real failure flowing through tool-improver and back.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_tool_acceptance_runner(2).md#flow; OVERVIEW_self_verifying_tools.md#autonomous-loop
- **relations:** findings acceptance_test pattern (improvement-loop); tool-improver; continuous sweep.
- **verify-later:** an improve_tool item with source 'acceptance' processed end-to-end by tool-improver.

### Criteria contract v0 (check-type vocabulary + profiles)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** P0 implements 3 of 7 check types; "the composer emitted "action":"select" … a verb the Tier-4 criteria vocabulary must now define" (open).
- **what:** The machine-readable criteria schema: `profiles: [desktop, mobile]`; check types `selector_exists`, `selector_count`, `no_console_errors`, `asset_loads`, `interaction` (fill/click/select steps + expect), `no_horizontal_overflow`, `page_status_ok`. Deterministic only in v0 — no LLM drives the browser. Desktop = Chromium 1366×900; mobile = one stable Playwright device descriptor (emulation first; real devices out of scope). Phasing P0 boot checks → P1 interpreter+mobile → P2 interactions → P3 screenshots (via the existing Backblaze deploy path) → P4 optional LLM-exploratory mode.
- **sources:** PLAN_tool_acceptance_runner(2).md#criteria-contract,#profiles,#phasing; RUNBOOK_travelling_docs(38).md#stage-6
- **relations:** browser-runner adapter (P0); multi-page tool criteria (open question — url_role field).
- **verify-later:** criteria interpreter coverage in run_checks_action.go; whether "select" verb was added.

### Multi-page tool documentation prerequisites
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK §5.4 "Multi-page prerequisite: preserve-sections re-render + interactivity-aware save guard (pending) before scaling page counts."
- **what:** Multi-page tools add a "Page set & inter-page contract" PLAN section (URLs, shared state keys, data feeds) and may need per-page checks (a `url_role` field). Scaling page counts is explicitly gated on the pending preserve-sections re-render and interactivity-aware save guard.
- **sources:** RUNBOOK_travelling_docs(38).md#§2,#§5; PLAN_travelling_docs(6).md#tool-assurance; PLAN_tool_acceptance_runner(2).md#open-questions
- **relations:** interactive-section clobber (Part 4) below; criteria contract.
- **verify-later:** save_page_sections interactivity guard deployment status.

### Snapshot-before-update standing rule + the platform's snapshot_agent()
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Snapshot before updating (STANDING RULE 2026-07-06)"; write-location confirmed 2026-07-07 (stored OUTSIDE agent_definitions); every subsequent migration carries the call.
- **what:** `SELECT snapshot_agent('<type>', '<migration>: pre-update')` is prepended to every agent-updating migration. The function already existed (a reuse win — the drafted side-table migration was superseded un-applied; lesson: fetch-first applies to FUNCTIONS too, `\df` before drafting backup machinery). Snapshots live in a separate store (later identified as agent_definitions_backup in the FYI), so the defensive `is_snapshot` selector predicate is not load-bearing. Companion `revert_agent('<type>')` exists per 016b.
- **sources:** RUNBOOK_travelling_docs(38).md#§0-REF,#task-1; RUNNING_NOTES_travelling_docs(39).md#rev18,#rev19,#rev22; FYI_from_fixloop_2026-07-10…md
- **relations:** migrations system; correct-while-touching.
- **verify-later:** `snapshot_agent`/`revert_agent` function definitions; agent_definitions_backup table.

### Correct-while-touching norm (bounded repair of adjacent inert bugs)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Defined in the RUNBOOK mini-glossary, "Norm adopted in this chat, 2026-07-06"; exercised in migrations 125–146 (e.g. all ten of recreation's step-level error_steps corrected while adding its note tail; 146's cooldown fix).
- **what:** When a migration already modifies a workflow, it also fixes known-inert bugs in that SAME workflow (e.g. step-level `error_step` moved into config with original targets, dead keys deleted), declared explicitly in the file — bounded repair, no separate campaign, never copying the broken shape into new steps.
- **sources:** RUNBOOK_travelling_docs(38).md#mini-glossary; RUNNING_NOTES_travelling_docs(39).md#rev23,#rev26,#tier-4-continuous
- **relations:** error_step-in-config; snapshot rule.
- **verify-later:** the declared correct-while-touching sections inside migrations 125–146.

### Migrations system — schema_migrations ledger + guarded runner
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Migrations system live (2026-07-10)"; 140 was "the system's first real apply"; applied through 146 at unit close.
- **what:** `schema_migrations` table (filename PK, applied_at, md5 checksum, applied_by, notes) + `scripts/migration/run-migrations.sh` (dry-run default, `--apply`, per-file re-check, stops-without-recording on failure, LOUD warning for near-miss filenames since the repo really uses `NNNb_`/hyphenated names). Home `docs/agent_docs/sql_for_agents/`, baseline 124 (001–123 = pre-system history, never auto-applied). The travelling-docs arc was renumbered 125–139 in applied order and backfilled with dates from the runbook; 128 is an honest reconstruction stub (original lost with an old workspace; effect verified live; NULL checksum). Every migration carries its own guard DO block; parking a re-enable migration outside sql_for_agents/ (141) was used as a deliberate safety gate against a stray --apply.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#2026-07-10-migrations,#fix-140-141; RUNBOOK_travelling_docs(38).md#task-5-closed; HANDOFF_2026-07-10…md#§0.3
- **relations:** migration write-hook to pipeline NOTES; snapshot rule; guard-design patterns (016b).
- **verify-later:** schema_migrations rows; run-migrations.sh behaviour; the 128 stub.

### error_step mechanics — config-level placement, existing target, derive-from-next_step, loop corollary
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016b(7) §9 entry (live-validated ×5 in run-3); mechanism documented in 001 §16; dormant step-level instances corrected across tool-generator/fix agents by migrations.
- **what:** The coordinator reads only `step.Config["error_step"]` — step-LEVEL error_step is parsed but silently ignored. Once placed correctly, the target must name an EXISTING step or the coordinator fails the whole workflow (a typo converts a recoverable failure into a fatal one). Pattern: derive `error_step` from the step's own `next_step` read from the same row (convergence by construction, nothing guessed); `jsonb_set` does not create parents — COALESCE-merge config. Loop corollary: inside loop substeps, `error_step`/`then_step`/`fallback_step` are iteration-prefixed at expansion and must name substeps of the same loop; `continue_on_error: true` is the iteration-scoped alternative.
- **sources:** 016b_debugging_guide_7_3_(7).md#error_step-entry; RUNBOOK_travelling_docs(38).md#§8; RUNNING_NOTES_travelling_docs(39).md#rev9,#rev12,#rev13
- **relations:** docs-never-fail containment; correct-while-touching; loop_expansion_handler.go.
- **verify-later:** `routeToErrorStepOrFail`/`continueExecution` in coordinator; loop_expansion_handler.go prefixing.

### Anchorless (code-only) diagnosis degrade at load_runtime
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Corrective APPLIED 2026-07-06; fired ×5 per anchorless run ("NORMAL, not a fault"); softening (`skipped:true` return) still a chassis-build follow-up.
- **what:** Runtime evidence is an optional bundle tier, but `diagnose_load_runtime` hard-errored with no site/correlation/domain anchor and had no error routing — making the tier mandatory in practice and killing legitimate code-only diagnosis runs. Fixed by config-level error_step on load_runtime targeting its own next_step (`assemble_bundle`); since `route.gather_step` re-enters load_runtime every iteration, each loop-back degrades per-iteration to a code+schema bundle. Cost of a full anchorless loop: ≈26 min, 5 iterations.
- **sources:** 016b_debugging_guide_7_3_(7).md#anchorless-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12,#rev14; 084_TRIGGER_diagnose_v1(2).sh (ANCHOR NOTE)
- **relations:** error_step mechanics; diagnosis loop step map (analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict → route → emit → persist_note → complete).
- **verify-later:** `diagnose_load_runtime` no-anchor softening (shipped or not); load_runtime.config.error_step live value.

### Verdict symptom-coverage gate (symptom_check) on the diagnose-agent
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** FYI 2026-07-10: prompt rule 8 + `symptom_check` schema field applied (snapshot 34f4afc8); engine coercion rides the next chassis image post-v1.0.1101; F0.6 addendum adds `cites`/`context` members.
- **what:** A CONFIRMED verdict must account for every distinct observation of the ORIGINAL symptom via `symptom_check: [{observation, explained, how, cites, context}]`; the chassis engine (`pkg/diagnose`) coerces to UNVERIFIABLE any CONFIRMED verdict whose symptom_check is missing, carries an unexplained entry, or marks explained without a valid citation index; comparative/background clauses are exempted as `context` rather than grade-inflated. Terminal diagnosis notes gain a "Symptom coverage:" block. Owned by the fix-loop workstream; delivered to this unit as a courtesy collision-rule FYI.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md (whole)
- **relations:** persist_diagnosis_note (note bodies change); fix-loop council/verdict work (fixloop_eg_dartsonline docs).
- **verify-later:** diagnose-agent verdict prompt_template; `verdict_wire.go` symptom_check parsing.

### fix-proposer agent (F1.1a) — read-only proposal writer
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** FYI addendum 2026-07-10: "a fix-proposer agent (F1.1a) now exists".
- **what:** An agent in the diagnosis→fix loop that reads only orchestration_states/diagnosis_artifacts and writes only kind='fix_plan' artifacts — no code writes, no git token. Noted here as a boundary fact for the travelling-docs surface owners.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#addendum
- **relations:** diagnosis loop; fix-loop workstream (primary docs elsewhere).
- **verify-later:** fix-proposer agent_definitions row; diagnosis_artifacts table.

### max_tokens placement rule — dead config outside ai_service
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** FYI second addendum 2026-07-10: "verdict step max_tokens was DEAD CONFIG … Fixed on both agents (snapshot first)".
- **what:** `execute_llm_prompt` reads `max_tokens` from the agent's top-level config or from INSIDE the step's `ai_service` block — never from the step-config root, where several agents had it; the Anthropic client then silently defaults to 2048 output tokens. A truncated verdict JSON parses to UNVERIFIABLE. Standing grep: `config.max_tokens` outside `ai_service` is dead wherever execute_llm_prompt is the action.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#second-addendum
- **relations:** silent-no-op config-path heuristic (016b durable invariants); execute_llm_prompt shared action.
- **verify-later:** ai_actions.go:252-256 max_tokens resolution; remaining workflows with root-level max_tokens.

### agent_error_log is the FIRST read
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v6 entry (2026-07-08), promoted after it settled the tool-generation blocker in one query.
- **what:** Step failures persist to `agent_error_log` (orchestration_id — TEXT not uuid — step_name, action, error_message, error_code, context) and outlive the pod. Read it first, filtered by orchestration_id; only then pod logs (may be reaped) or collected_data (may be enormous). `current_step` from polling is a sample, not an attribution (a 120s poll blamed the LLM step when save_tool failed); a terminal step's success_message can name the wrong phase.
- **sources:** 016b_debugging_guide_7_3_(7).md#agent-error-log-entry; HANDOFF_2026-07-08…md#§3,§5; RUNNING_NOTES_travelling_docs(39).md#rev29
- **relations:** schema drift incident; two failure envelopes; 0-rows rule.
- **verify-later:** agent_error_log schema (orchestration_id type).

### Code-ahead-of-DB schema drift (SQLSTATE 42703, latent until first caller)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Root-caused 2026-07-08 (create_tool_component vs missing content_components provenance columns, latent ~2 months since 2026-05-16); fix applied + proven 2026-07-09.
- **what:** A binary referencing new columns deployed before its migration; nothing fails until the rare code path is called. Detection: the failing INSERT's own comment names the missing migration; a last-successful-call latency probe distinguishes latent drift from fresh regression. Fix pattern: MIRROR column types dynamically from the table the code says it mirrors (format_type/pg_attribute + ADD COLUMN IF NOT EXISTS), additive/nullable/idempotent. The canonical migration file existed but was parked in a docs folder, never renumbered into the migrations path — the exact mechanism by which a deploy skips a migration (one motivation for the migrations system). Standing pre-deploy check: grep the diff for new column names and assert each exists in production.
- **sources:** 016b_debugging_guide_7_3_(7).md#schema-drift-entry; HANDOFF_2026-07-08…md#§3; RUNNING_NOTES_travelling_docs(39).md#rev29,#rev30
- **relations:** migrations system; content_components provenance columns (migration 133); "provenance stamps the chassis".
- **verify-later:** sql_for_agents/133_add_component_provenance.sql vs the docs019 design copy.

### Provenance stamps the chassis, not the logical agent — config-declared source is the reliable provenance
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "Provenance stamps the CHASSIS: … source_agent_type='generic' … the planned doc-action fallback is DROPPED — backlog item closed by evidence rather than by code" (rev 32).
- **what:** Both `Headers["agent_type"]` (empty in step context) and `ExecutionContext.Sender.AgentType` (stamps 'generic' — the shared chassis pod) fail to identify the logical agent. Doc rows therefore rely on the config-declared `source`/`plan_source`/`note_source` fields for provenance, which the actions already carry. Applies equally to `content_components.source_agent_type` and rag_actions.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev31-watch,#rev32; HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§5; 016b_debugging_guide_7_3_(7).md#schema-traps
- **relations:** four doc actions; component provenance columns.
- **verify-later:** source vs source_agent population on live doc_plans/doc_notes rows.

### Prompt-template vs config-path resolvers (TEMPLATE_FIELD_ERROR)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v7 entry; root cause of the missing first auto-PLAN (rev 32); latent same-class bug in Task-4 templates caught before it fired.
- **what:** `execute_llm_prompt` with `output_format: text` hands the prompt template the BARE string (`{{.X}}`); with `json` a map (live form `{{.X.result | toJSON}}`); action CONFIG field paths are a different resolver and keep `.result`. Never reach an unverified nested key from a template — dump whole objects with `| toJSON`. A render-time error fires before tokens are spent and, with error containment, the workflow "succeeds" while the step's product is missing (reading rule: normal terminal + missing downstream artefact = contained step failure).
- **sources:** 016b_debugging_guide_7_3_(7).md#template-entry; RUNNING_NOTES_travelling_docs(39).md#rev32; HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§2
- **relations:** docs-never-fail containment (masking effect); seam rule.
- **verify-later:** template data shaping in ai_actions.go by output_format.

### The seam rule — every prompt consuming a spec field must render it
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Migration 138 applied ("Mandatory Behaviour Requirements" section rendered from spec.interactive_features, marked as overriding the source); PLAN(6) rollout-outcomes bullet.
- **what:** A requirement can survive the analysis step and still be ignored at generation if the generation prompt never renders it: `analyze_tool` rendered `spec.interactive_features`, `recreate_tool` didn't, and Opus trusted the visible source HTML over requirements buried in a 20KB analysis JSON — faithfully recreating the bugs it was asked to fix. Rule: when adding a spec field, grep EVERY prompt_template that should render it; render requirements explicitly, marked as overriding the source.
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45-run2; HANDOFF_2026-07-10…md#§4
- **relations:** economy-simulator case; "passed checks ≠ working"; doc_notes `seam` category.
- **verify-later:** recreate_tool prompt's Mandatory Behaviour Requirements section (migration 138).

### EXECUTING_STEP forever = the worker died (OOMKill triage), superseding stall/leak readings
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v8 rewrite 2026-07-09; the earlier v5-era "error containment does not protect against a HANG" entry and the slow-leak hypothesis are explicitly superseded on the evidence trail kept in RUNBOOK.
- **what:** `orchestration_states` is written BY the worker: a dead pod (OOMKill exit 137, eviction, panic) writes nothing, so the row freezes at EXECUTING_STEP and `since_s` measures time since the crash. Triage order: RESTARTS column → describe pod Last State → `logs --previous` (capture crash logs IMMEDIATELY — a ReplicaSet replacement erases them). Probe suspected-stalled dependencies with a bound (`curl -m 5`) before assuming a hang. Related-but-distinct: genuine stalls from missing context deadlines deserve fixing as hygiene. The arc walked through three wrong hypotheses (stall → missing deadline → slow leak) before the real cause (chunkContent loop), each correction documented rather than discarded.
- **sources:** 016b_debugging_guide_7_3_(7).md#executing-step-entry; RUNBOOK_travelling_docs(38).md#superseded-incident-block; RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#rev36
- **relations:** chunkContent bug (the answer); containment-limit corollary.
- **verify-later:** n/a (operational pattern).

### chunkContent() infinite loop — the OOM root cause, fixed with timeout regression tests
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "THE OOM ANSWER (closes the incident chain for good)" — confirmed 2026-07-10; fix deployed v1.0.1104; proof run 05d1fc97 with 0 pod restarts.
- **what:** `chunkContent()` in rag_actions.go never terminated on content > chunk_size: the final chunk ends at len(content), `start = end - overlap` steps BACKWARDS, the same tail appends forever → 2Gi in seconds; content ≤ 1000 chars returned early, hiding the bug for weeks (both OOMKills were PLAN-sized bodies through index_plan). Fixed with a final-chunk break + forward-progress guard and four regression tests with a 30s timeout that catches loop regressions. Durable class rule: content-below-threshold early returns can hide a non-terminating path; "a proof run is a probe — fire proofs early" (the 139 proof run found the real cause within the hour).
- **sources:** RUNBOOK_travelling_docs(38).md#task-6; RUNNING_NOTES_travelling_docs(39).md#v1.0.1103-proof-run,#fix-140-141; HANDOFF_2026-07-10…md#§1,§4
- **relations:** tool_docs indexing (unblocked); migrations 140/141; EXECUTING_STEP pattern.
- **verify-later:** rag_actions_chunk_test.go; chunkContent forward-progress guard.

### kcat -P is line-delimited — single-line trigger bodies
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Run 464102f4 post-mortem (rev 45); 086 and 087 scripts enforce single-line bodies.
- **what:** A pretty-printed JSON body piped to `kcat -P` becomes one message per line; the chassis can then marry your headers to a NEIGHBOURING message's body (observed: our correlation id completing "after 0 steps" holding a scheduler no-op's body — also flagged a chassis stale-buffer wrinkle worth a look). Trigger bodies must be compacted to a single line and scripts must refuse multi-line.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev45-run1; RUNBOOK_travelling_docs(38).md#new-durable-rules; 086/087 script headers
- **relations:** manual kcat trigger scripts; env-prefix trap (sibling).
- **verify-later:** the stale-buffer wrinkle in the chassis consumer (never followed up).

### Env-prefix trap — VAR=x on its own line (or with `;`) never reaches the child
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Cost two 3b.4 runs and one 085 run; 084/086 banners hardened to print the go/no-go tell ("Subject: NONE — will SKIP").
- **what:** Shell variables set on their own line (or terminated by `;` before the command) are not exported to child processes, so triggers silently run with defaults. Correct forms: same-line prefix or `export`. Scripts now print explicit banners of the effective values as the load-bearing tell.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev19,#rev33; RUNBOOK_travelling_docs(38).md#§8
- **relations:** trigger scripts; banner-tell convention.
- **verify-later:** n/a (operational pattern).

### Two failure envelopes — parent COMPLETED ≠ child success
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v5 entry; observed live in runs 1–2 of the 3a arc.
- **what:** A mid-run child failure returns header `status: complete` with the failure in the BODY (`body.status: failed`) — the parent forwards it and completes (a COMPLETED parent row with non-empty `error` = a forwarded child failure); a failed-to-START child sends `status: error_unrecoverable` / `CHILD_ORCHESTRATION_FAILED`. Consumers must check the body, never the header alone; which shape appears tells WHERE the child died.
- **sources:** 016b_debugging_guide_7_3_(7).md#failure-envelopes-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12
- **relations:** agent_error_log first read; §0-REF reading rules.
- **verify-later:** sendWorkflowResponse / notifyParentOfFailure paths.

### Pod label `agent-type` (hyphen) + multi-pod log attribution
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v5 entry; proven by a working command vs a zero-match selector.
- **what:** The pod LABEL key is `agent-type` (hyphen) while log JSON fields say `agent_type` (underscore) — the underscore selector silently matches zero pods. A type-wide selector spans ALL live pods (idle reaper 3600s), so tails contain residue from earlier runs: attribute every line by orchestration id / pod / timestamp before reading it as current.
- **sources:** 016b_debugging_guide_7_3_(7).md#label-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev13
- **relations:** 0-rows rule; §0-REF.
- **verify-later:** n/a (operational pattern).

### 0-rows rule + gate-evidence capture window + state-dump substitute
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Run-3 closed via the state-dump substitute 2026-07-06; codified in 016b anchorless entry and RUNBOOK §7.
- **what:** 0 rows is decisive only after the query itself AND the run's completion are ruled in (a run that died upstream also produces 0 rows). When a step's NON-firing is the success condition, closure needs a COMPLETED child + the step's explicit skip log line + the 0-count. Skip log lines have a 3600s capture window (idle reaper); past it, a post-completion state dump (ProcessingHistory showing the step executed + terminal status + 0-count) is the accepted substitute. Placeholders are replaced INCLUDING the angle brackets.
- **sources:** 016b_debugging_guide_7_3_(7).md#anchorless-entry (verification discipline); RUNBOOK_travelling_docs(38).md#§7,#stage-3; RUNNING_NOTES_travelling_docs(39).md#rev16
- **relations:** persist_diagnosis_note gate proof; agent_error_log.
- **verify-later:** idle-reaper timeout value (3600s).

### Postgres guard-writing gotchas — RE_DUP_MAX 255, sticky aborted transactions, psql -f over paste
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b entry written from the supersede-migration attempts 1–3 (2026-07-09).
- **what:** Postgres ARE caps bounded regex repetition `{m,n}` at 255 — prefer `strpos`/`substr` in guards (plainer intent, no engine limit). After any in-transaction error the session is stuck (`clients_db=!#`) and ignores everything including BEGIN — `ROLLBACK;` is the only exit; migration files should open with a defensive ROLLBACK and be run with `psql -f`/`\i` (pasting mangles comments and dollar-quoted bodies). A guard that refuses a write can be RIGHT (it blocked an unverified selector) or WRONG (it refused a valid runtime-built selector) — guard design evolved to accept static OR dynamic evidence with a NOTICE saying which path verified.
- **sources:** 016b_debugging_guide_7_3_(7).md#postgres-regex-entry; RUNNING_NOTES_travelling_docs(39).md#rev37,#rev38,#rev39; 0NN_supersede_xp_curve_plan_selectors(2).sql
- **relations:** needle-gate template-surgery pattern; anchor rule (the design insight that came out of guard 1's refusal).
- **verify-later:** n/a (operational pattern).

### Untracked-file deploy trap — verify deploys by ancestry, not by tag or commit message
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Struck TWICE (Tier-2 checker missed two releases; check_tool_acceptance_due missed v1.0.1111); rules banked in HANDOFF T11 and the durable-rules list.
- **what:** `git commit -a` commits modified-tracked files only — an untracked (`??`) new file silently misses any number of release commits while its sibling changes ship. Guards: `git status` for `??` before every release (or commit new files as written); verify a deploy carries your files by ANCESTRY (`git merge-base --is-ancestor <commit> <release>`); this repo also reuses version tags, so pod-start-time vs commit-time settles what a tag actually contains, not the commit message. Safe-failure companion: unknown discovery-check names warn+skip (the 142 precedent), so wiring a check by migration before its binary deploys is safe.
- **sources:** HANDOFF_2026-07-10…md#T8,T11,#§4; RUNNING_NOTES_travelling_docs(39).md#stage-5-live,#v1.0.1111; README_summary_paragraph2_for_discussion.md
- **relations:** continuous sweep gate; migrations-before-binary safety.
- **verify-later:** n/a (operational pattern).

### Manual kcat trigger scripts (084–087) — the canonical manual-orchestrate envelope
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** All four scripts exist, exercised repeatedly; 086/087 add DRY-RUN-by-default + single-line enforcement.
- **what:** A family of shell triggers producing `action=orchestrate` messages to `system.agent.generic.requests` with the house envelope (correlation/orchestration/request/message ids; `config.agent_type=<target>`). Encoded operational knowledge: target the spawn-wrapper orchestrator, not the agent directly (the SPAWNED pod gets GITHUB_READ_TOKEN via the spawn gate; an in-place run on a shared pod fails pre-fetch); REF explicit, never HEAD (user decision 2026-07-02); banners print effective subject/function as the go/no-go tell; DRY-RUN default with SEND=1; declared real side effects on the live trial site.
- **sources:** 084_TRIGGER_diagnose_v1(2).sh, 085/086/087 headers; RUNNING_NOTES_travelling_docs(39).md#rev10,#rev27; 086_input_data_recreate_economy_simulator.json
- **relations:** kcat line-delimited trap; env-prefix trap; spawn+call input shape.
- **verify-later:** whether these scripts were promoted anywhere canonical (they live in drafts/ and this docs dir).

### Manual spawn+call input shape — satisfy the contract top-level AND the workflow's input_data.spec.*
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016b entry with the 081/082/083 trigger trilogy as evidence (contract violation vs empty-context generation vs correct).
- **what:** `call_agent` validates input_mapping output against the target's input_contract at TOP level, while workflows for work-item-driven agents read `input_data.spec.*` — so manual invocation must provide required fields both top-level and inside a spec object (or better, drive the agent via its designed work-item trigger). Flagged as a latent design smell: contract and workflow should agree. Companion: `store_generated_component` keys regeneration on the LLM's EMITTED function — a mismatched name INSERTs a stray duplicate.
- **sources:** 016b_debugging_guide_7_3_(7).md#spawn-call-entry
- **relations:** diagnosis subject threading (same contract rule); trigger scripts.
- **verify-later:** call_agent contract validation code; component-creator contract vs workflow paths.

### Handoff-document discipline (updated-every-turn, supersede chain, turn log)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Three-generation chain in this unit (07-08 → 07-09 → 07-10, each "supersedes" the prior); 07-10 handoff carries an 11-entry turn log through 2026-07-13.
- **what:** Long-running work travels between chat sessions via a standing HANDOFF doc: first-actions list, state-of-play with dates and snapshot ids, blocker sections with ranked hypotheses and data-to-collect checklists, durable rules, key identifiers, file inventory, and a newest-first turn log updated EVERY turn. Companions pattern: RUNBOOK (position tracker) + RUNNING_NOTES (chronology) + PLAN (spec) + 016b (durable patterns) — the travelling-docs idea applied to the work itself. Includes the cross-workstream "collision rule" courtesy FYI when another chat touches a shared surface.
- **sources:** HANDOFF_2026-07-08…md; HANDOFF_2026-07-09…_1_.md; HANDOFF_2026-07-10…md#turn-log; FYI_from_fixloop…md (collision rule); README_summary_paragraph_for_handoff.md
- **relations:** doc traveller / docs037 conventions; bundle command.
- **verify-later:** n/a (working practice).

### Context-bundle command for cross-chat handoffs (cmd/bundle + registry-based scope resolution)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Two bundles built and used (07-08 toolgen bug; 07-09 recreation); resolver rewritten after 3 misses (rev 44).
- **what:** `cmd/bundle` renders a task bundle (constitution + task text + code scopes + docs + live schema + runtime evidence incl. an agent_error_log "Recent errors" section — the section that settled the 07-08 diagnosis). Path facts banked: resolve actions via registry.go (action name → constructor → defining file), not filename convention (`execute_llm_prompt` lives in ai_actions.go; validate_page_content.go lacks the _action suffix); misses are non-fatal and print grep candidates.
- **sources:** bundle_recreation_v1(1).sh (header + resolve_action); HANDOFF_2026-07-08…md#§6; RUNNING_NOTES_travelling_docs(39).md#rev44
- **relations:** docs019 contextkit/bundles; agent_error_log first read.
- **verify-later:** cmd/bundle flags; whether the runtime errors section is standard.

### Recreation writes page sections — component-less tools and their visibility gap
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Established by query 2026-07-09 ("pages.sections is EMPTY … the 32 KB game body exists only as deployed HTML in the sites repo"); Tier-2 scope note 2026-07-10.
- **what:** `tool-recreation-handler` ends save_page_sections → update_status → deploy_page and never creates a `content_components` row — adopted/recreated tools exist only as page sections + deployed HTML (source in adoption-crawl research_results: adoption_crawl full markdown+rawHTML, adoption_page per-page; `spec.mode="recreate"` is the handshake set by apply_adoption_plan). Consequences: no component address for tool-improver; invisible to Tier 2 by construction (Tier 4 reaches them via pages); NOTES subject must be pipeline-scoped. `site_plan_sections` is site-plan STRUCTURE, not HTML.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev41,#rev42,#rev43; HANDOFF_2026-07-09…_1_.md#§4; RUNBOOK_travelling_docs(38).md#task-5-record
- **relations:** dangling-doc rule; adoption pipeline (007); Tier-2 scope limit.
- **verify-later:** tool-recreation-handler workflow steps; research_results result_types.

### create_tool_component updates in place by function; unique index covers active library originals
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Side finding rev 33 (same function re-run → one component row, same id); index predicate read 2026-07-07.
- **what:** `idx_cc_tool_function_unique` = UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND is_active=true — uniqueness covers ACTIVE LIBRARY ORIGINALS only (duplicate function rows are forks/inactive versions), and `create_tool_component` updates an existing function in place rather than duplicating. Vindicates function-keyed docs (they span all instances). Also banked: content_components has NO site_id column (site scoping via page_components/site_components only); created_from CHECK {manual,generated,adopted,tool,forked}.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23,#rev28,#rev33; HANDOFF_2026-07-08…md#§7
- **relations:** doc subject convention; provenance columns.
- **verify-later:** pg_indexes indexdef for idx_cc_tool_function_unique.

### Tool creation never enqueues the final page deploy (planned-pages gap)
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Gap on record: tool creation ends at `complete` without enqueuing a page_rerender item — the pages deploy only when something else sweeps" (2026-07-10; pages hand-deployed by inserting page_rerender items).
- **what:** tool-generator creates component + page + nav but leaves the page `build_status='planned'`; nothing enqueues the render+deploy hop, so new tool pages 404 until an unrelated sweep. Recorded follow-up: a `create_rerender_item` tail on tool-generator. Interacts with acceptance timing (the reason post-creation acceptance hooks were rejected in favour of the deployed-pages-only sweep).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#both-planned-tool-pages-deployed; HANDOFF_2026-07-10…md#§5.3
- **relations:** continuous sweep design; build/rerender pipeline.
- **verify-later:** whether tool-generator gained a rerender tail.

### Inline-JS extraction ("Path 1" /tools/assets/<fn>.js) — designed, partly real, not on the deploy path
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** 016b v3 entry (separateInlineJS/collectJSAssets exist and are correct for the store path); Tier-2 first sweep proved the deploy path ships JS INLINE and never references the asset (js-not-extracted, Option B superseded the criteria).
- **what:** The store path's `separateInlineJS` extracts a bare inline `<script>` into `js_content`, replaced by a `<script src="/tools/assets/{function}.js">` reference deployed by `collectJSAssets` — but only for attribute-less tags, and legacy/seeded rows predate it (empty shells with raw inline scripts; provocation-card additionally truncated mid-script — store validation checks unclosed `<style>` but not `<script>`). Meanwhile the generator/deploy route for new tools ships everything inline, so "Path 1 extraction" is delivered reality nowhere on that route. Hardening recorded: script-balance check at store time; regenerate broken shells through the current path.
- **sources:** 016b_debugging_guide_7_3_(7).md#js-not-extracted-entry; RUNBOOK_travelling_docs(38).md#stage-5 (pre-verification); PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** delivered-reality principle (Option B); empty-shell/mode-b categories; vonc case evidence.
- **verify-later:** store_generated_component_action.go separateInlineJS; whether extraction ever ships.

### Interactive-section clobber + interactivity-aware save guard (preserve-sections)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 016b Part 4 "CAUSE CONFIRMED; FIX PENDING" → 2026-06-24 "fix WRITTEN (un-deployed)"; still listed as the pending multi-page prerequisite in RUNBOOK §5.4 at unit close.
- **what:** Any full rebuild regenerates a page from plan_sections, which knows nothing of an interactive tool stored only as a section's rendered_html — the game is silently discarded (detool-on-rebuild). Layered fix, both layers in `save_page_sections` (the only place holding the markup): (1) interactivity guard blocking a non-interactive set replacing a deployed interactive one; (2) carry-forward of existing interactive sections; plus source_item_id stamping. The unstated invariant "interactive sections survive every rebuild route" is the canonical example of what pipeline PLAN invariants should record.
- **sources:** 016b_debugging_guide_7_3_(7).md#open-threads-part-4; RUNBOOK_travelling_docs(38).md#§5.4; PLAN_travelling_docs(6).md#tool-assurance
- **relations:** multi-page prerequisites; pipeline documentation model; page build/rerender pipeline threads (Parts 1–5).
- **verify-later:** whether the patched save_page_sections_action.go deployed; page_component_history source_item_id.

### Page build/rerender failure-shape thread family (Parts 1–5 + wrong-turns log)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b open-threads status header: Part 1 DONE/verified; Part 2 partially verified; Part 3 code prepared not applied; Part 4 written un-deployed; Part 5 triaged.
- **what:** A connected series on "work that reports success but doesn't happen": result-contract drop replaced child output with a success stub (fixed, shipped 2026-06-18); no-LLM re-render pre-pass (partially verified); item_key canonicalization drift (needs_page vs needs_tool_recreation colliding on the dedup index — builder prepared); the interactive clobber (above); system-stats dropped because content_data and the component template share ZERO keys (a content↔template key-contract mismatch — the visible-content filter was correct). The companion "Wrong turns" log records false leads with the durable heuristic each violated — a deliberate documentation convention so the next pass doesn't re-walk them.
- **sources:** 016b_debugging_guide_7_3_(7).md#open-threads,#wrong-turns; (fix detail lives in the gamesdesign/scheme runbooks outside this unit)
- **relations:** silent-completion invariants below; travelling copies of 016b.
- **verify-later:** current state of Parts 2/3/5 fixes.

### Debugging durable invariants (trust the artefact; sampled steps; silent no-op config paths)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b "Durable invariants and heuristics" section, carried forward from 016 and extended through this arc.
- **what:** The distilled heuristics: trust the rendered artefact, not the status (work items/commits can report success on no-op work); completed_at is the orchestration end, not the write instant; a config key read on a different path than it's set is a silent no-op, not an error; only save_page_sections writes page_components; 0 rows is not decisive; a negative inference from an artefact's shape needs the mechanism checked in ALL cases; reuse before rebuild; check the schema before SQL. Plus 016b v4 additions: two page-assembly paths with different chrome sources (stale `site_components` renders fossilise; only a full page-build rebuild re-renders templates; provenance greps + legacy-variable tell); the needle-gate template-surgery pattern (LIKE booleans + occurrence counts + backup + guarded idempotent UPDATE + RETURNING + rollback); `sites.status` vocabulary (draft/building/review/published/deployed/archived/error — 'active' is legacy; nothing filters on it; never scope blast-radius by it).
- **sources:** 016b_debugging_guide_7_3_(7).md#durable-invariants,#light-site-dark-chrome,#sql-pitfalls,#sites-status
- **relations:** the whole debugging category; 016 back-catalogue (other unit).
- **verify-later:** n/a (heuristics; primary copies of 016/016b covered by their own unit).

### Aspiration: agent-creation and inter-agent message logging workstream
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES "Note on a separate, out-of-scope item" — "kept out of these docs to preserve separation of concerns. Can be specced separately." No later mention.
- **what:** A stated desire to closely log/track agent creation and inter-agent messages (headers + body) as a distinct workstream from travelling docs — different responsibility and data. Never designed or built within this unit's horizon.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#out-of-scope-note
- **relations:** travelling docs (deliberately separated); message envelope standards (035).
- **verify-later:** whether any message-logging workstream exists elsewhere in docs/.

### Standing opens ledger of the travelling-docs arc
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** Repeated "Standing opens" list in every rev's open threads and HANDOFF §5, still open at unit close.
- **what:** The carried-forward small items: `deploy_tool_to_site` should stamp `source_*` on forks (NOTES-only on fork — unverified); `rag_index` `source_type='scrape'` parameterisation; the Tier-4 vocabulary "select" verb; P1 mobile / P2 interactions; a real (non-manufactured) acceptance failure through tool-improver and back; github-actions-runner cgroup-driver CrashLoopBackOff (infra, not app); chassis memory slope watch (leak neither proven nor needed after the chunkContent answer).
- **sources:** HANDOFF_2026-07-10…md#§4,§5; RUNNING_NOTES_travelling_docs(39).md (open-threads sections); RUNBOOK_travelling_docs(38).md#background
- **relations:** most concepts above.
- **verify-later:** each item individually in stage 2.
