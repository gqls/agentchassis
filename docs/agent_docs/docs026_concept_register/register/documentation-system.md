# Register — documentation-system

66 concepts, consolidated from 81 raw extractions (each doubled by an input-file
duplication artifact, so 162 raw headers collapsed to 81 before merging) across
units U01, U03, U04, U05, U06, U07, U08, U10, U11, U12, U13, U14, U15, U16, U17a,
U17b, U18, U23, U24a, U24b, U24c, U24e, U24f, U25.

### DOC-001 — Documentation consolidation system (numbered canonical docs + index)
- **status:** deployed
- **status-evidence:** 000(2) "Consolidated from 57 source files"; consolidation notes at the head of 001/002/003 recording patch-incorporation and supersessions; 016 v2_58 full-diff consolidation note.
- **what:** The docs024 set is the consolidated canonical documentation: one numbered doc per area with an index, consolidation notes stating which patches are already incorporated (and must not be re-applied), version families closed by full heading+content diffs, and continuation volumes when a doc hits size limits (016→016b). A "Plans (review for currency)" section separates aspirationals from settled fact.
- **sources:** 000_documentation_index(2).md; consolidation notes in 001(5)/002(4)/003(8)/016 v2_58
- **relations:** per-tool docs (DOC-003); travelling doc conventions (DOC-010)
- **verify-later:** —

### DOC-002 — Anthropic product-knowledge skill (verify, don't recall)
- **status:** deployed
- **status-evidence:** 106 is the skill file itself.
- **what:** A skill instructing agents to consult official Anthropic docs (docs maps per product) rather than memory for any Claude API/Code/claude.ai facts — accuracy over guessing, source everything, distinguish the three products.
- **sources:** 106_claude_anthropic_skill.md
- **relations:** dated-claim verification convention (DOC-040)
- **verify-later:** where the skill is installed/used

### DOC-003 — Per-tool travelling documentation convention (PLAN_/NOTES_ files + taxonomy)
- **status:** partial
- **status-evidence:** Written as 037_TOOL_DOCS_convention 2026-06-29; "layer 1 (manual markdown) now, DB tool_docs table recommended target"; the social/tiktok copy states plainly "It is not built yet; this convention is the spec" while 12+ NOTES/PLAN files already exist in active manual use.
- **what:** Every tool/complex component carries PLAN_<function>.md (aim, source spec, behaviour contract, delivery mechanism Path1/Path2/build-time with WHY, dependencies, deliberate decisions that must not be "fixed") and NOTES_<function>.md (timestamped choices/bugs/dead-ends tagged with a shared problem-category taxonomy: css-variable-mismatch, empty-shell/mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift). The same convention, written once 2026-06-29, was instantiated near-identically across at least three project trees: the docs024 core copy, the vonc site-building thread (provocation-card, brief-explanation, lobby-grid, provocations-index/archive-list), and the social/tiktok thread. Storage decision: files/git now (versioned, human-reviewable), NOTES→DB table only once agents start writing them or tag-queries become routine; PLAN stays in git — a DB text column without versioning is the named worst option. Sits below 016/016b and site runbooks; end-state is docs generated automatically at creation and grown per change. Superseded in the travelling_docs thread by the Postgres-based system (DOC-010).
- **sources:** 037_TOOL_DOCS_convention(1).md; docs/RUNNING_NOTES_vonc(36).md#2026-06-29; docs/PLAN_provocation-card(3).md; docs/social001_vonc_tiktok_social/tool_docs/TOOL_DOCS_convention(3).md
- **relations:** tool doc header (in-code anchor); 016b category tags; DOC-010 (Postgres successor); debugging-guide fork/merge (DOC-059)
- **verify-later:** TOOL_DOCS_convention.md canonical location; whether any tool_docs/tool_doc_notes DB table exists (expect none — superseded by doc_plans/doc_notes)

### DOC-004 — Running-notes checkpoint journal + distilled HANDOFF discipline (idea.uk, memory-off)
- **status:** deployed
- **status-evidence:** Main journal header: "Memory is OFF; this doc is the journal. Present this file at the END OF EVERY TURN."; running_notes.md grew to ~5,690 lines (checkpoints a…kkk) before archival into running_notes_2.md; the scheme_to_components sub-thread independently ran the same discipline (55 versions, lettered checkpoints Sa…Un).
- **what:** The idea.uk project's working method for coherent long multi-session threads without model memory: an append-only running-notes journal presented every turn, dated/lettered checkpoints, a carry-over state block (preferences, architecture conventions, project facts, "the fix in one line"), explicit AWAITING/NEXT lines, and CORRECTION entries that supersede earlier readings in place. Paired with a distilled HANDOFF kept fresh (current state, strict user preferences, schemas, backlog, marked "canonical cold-start doc") and companion PLAN (forward map) / RUNBOOK (commands + results, superseding "WHERE WE ARE" position blocks) / SPEC (decision record) documents, each with a defined role. The same discipline was run independently both on the main idea.uk thread and on its scheme_to_components sub-thread. Operational lore: attachments arrived unreadable repeatedly; pasted text and file uploads are the working channels.
- **sources:** idea.uk/running_notes_2(6).md (header); idea.uk/HANDOFF(13).md (header); running_notes_scheme_to_components(55).md; HANDOFF_scheme_to_components_for_claude_code(1).md
- **relations:** house rules; docs026's own charter (this journal family is a model input for the council); docs037 travelling-docs conventions; bundle packagers (DOC-005)
- **verify-later:** n/a — documentary practice

### DOC-005 — Docubundle context packagers + curated attach-lists for fresh threads (idea.uk)
- **status:** deployed
- **status-evidence:** Packager scripts "fixed for the real (messy) repo layout" 2026-06-10; two bundles produced (go-live 596KB, chassis-engine 1.5MB context files present).
- **what:** Self-contained bash packagers assemble a single context file for an AI assistant per task — the idea.uk go-live bundle (Go code + embedded page + deploy + the go-live docs; explicitly no live capture because there's no DB/k8s) and the chassis idea-engine bundle (the engine to port + the chassis framework to build it in, action catalogue for reuse-discovery). Copes with the messy folder by resolving docs to the newest "(N)" variant by mtime and dropping *.orig* backups and binaries. Complemented by hand-written attach-lists (BUNDLE_1/2, CONTEXT_PACK, CONTEXT_FOR_NEXT_CHAT) spelling out which files a fresh thread needs, warning the idea.uk files are NOT in the chassis project.
- **sources:** idea.uk/docubundle_idea_golive/package_idea_uk_golive.sh; idea.uk/docubundle_idea_within_chassis/package_chassis_idea_engine(3).sh; idea.uk/BUNDLE_1_idea_uk_golive.md; idea.uk/CONTEXT_FOR_NEXT_CHAT.md
- **relations:** diagnosis-loop bundle tooling (cmd/bundle); running-notes discipline (DOC-004)
- **verify-later:** n/a

### DOC-006 — Interactive HTML runbook checklist
- **status:** deployed
- **status-evidence:** RUNBOOK_gamesdesign_index_rebuild(29) header: "RUNBOOK_gamesdesign_checklist.html mirrors these parts/steps as tickable boxes … Seeded to the 2026-06-24 status."
- **what:** A self-contained HTML companion to the markdown runbook: per-part tickable checkboxes with locally-persisted state and progress bars, mirroring the runbook's `[ ]` step-checkbox convention. A small documentation-system pattern: dual-surface runbooks (grep-able markdown + stateful visual checklist), versions evolving in lockstep with the runbook.
- **sources:** RUNBOOK_gamesdesign_checklist(7).html; RUNBOOK_gamesdesign_index_rebuild(29).md
- **relations:** runbook discipline; travelling-doc conventions (DOC-010)
- **verify-later:** n/a (documentation artefact)

### DOC-007 — Packaged canonical-doc copies as debug context (003 contracts copy)
- **status:** deployed
- **status-evidence:** package_module/output_contexts/ contains a consolidated 003 copy ("Canonical 003 — Contracts & Standards, superseding v8–v10") plus the 4.3MB production code dump the packager emitted.
- **what:** The packaging workflow drops canonical guideline docs and a generated whole-slice code dump alongside the running notes so a fresh chat starts with ground truth. Registered here purely as provenance: the 003 contracts content and the production code dump are owned by their home units — this unit only evidences the packaging practice.
- **sources:** package_module/output_contexts/003_contracts_and_standards.md; package_module/output_contexts/production_content-and-linking-debug_context.txt
- **relations:** context packaging tooling; contracts-and-standards unit
- **verify-later:** n/a

### DOC-008 — Epistemic tagging and handoff-correction discipline
- **status:** deployed
- **status-evidence:** NOTES(45) header: "Epistemic tags used below: [verified-source]… [verified-db]… [deployed?]… [assumed]… [gap]"; §10 "Handoff-correction log (institutional memory)… Pattern: verify against code, not the handoff."
- **what:** The phase-5 notes operate a working epistemology: every claim carries a tag distinguishing read-from-source, confirmed-by-production-query, assumed, or known-gap; a dedicated correction log records where inherited handoffs contradicted deployed reality (reply-topic direction, prepare_object_url existence, the "list-keys gap" that already existed as ListObjects). Multiple bugs in this unit trace to trusting a doc over code (templates_db pin, backup-vs-live def divergence, runbook "safe to re-run"). A documentation-system convention worth institutionalising: docs are claims; code and DB state are evidence.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#header,#10; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md
- **relations:** docs026 programme itself (stage-2 verification mirrors this); hand-applied migrations lesson
- **verify-later:** n/a (convention)

### DOC-009 — Cold-start documentation bundle practice (BUNDLE/HANDOFF/PLAN/RUNBOOK + cmd/bundle)
- **status:** deployed
- **status-evidence:** "Decision: produce a cmd/bundle invocation + cold-start docs (BUNDLE/HANDOFF/PLAN/RUNBOOK) so a fresh chat could pick it up" (NOTES §2 Start); HANDOFF(7) explicitly "the cold-start entry point".
- **what:** A four-document travelling set per investigation — BUNDLE (a `cmd/bundle` invocation composing constitution + task + scoped code symbols + schemas + runtime evidence into one context file; `-step debug` for bodies, verified doc paths), HANDOFF (cold-start entry with operating model + status), PLAN (phased with gates/done-whens), RUNBOOK (live action document with YOU-ARE-HERE banner, per-step SQL + expected + CHECK blocks, ticked progress) — plus NOTES as the append-only journal owning every correction. Operational gotchas: pasted attachments extract empty (capture via `kubectl … psql -c > file`, not `\o`); runbooks rewritten wholesale when history outgrows action (old kept as `*_pre_cleanup_backup`).
- **sources:** BUNDLE(3).md; HANDOFF(7).md; NOTES(43).md §2, §9av, §9bc; RUNBOOK(49).md
- **relations:** documentation-system conventions (037, DOC-003); F2 discriminator discipline; bundle-first handoff practice (DOC-046 [see U16 entry below])
- **verify-later:** cmd/bundle tool flags (-scope/-schema-tables/-runtime-site/-df-filter)

### DOC-010 — Travelling documentation (PLAN + NOTES) in Postgres
- **status:** deployed
- **status-evidence:** RUNBOOK(38) §1 "Truth (Postgres, LIVE): doc_plans … doc_notes"; migration 125 applied 2026-07-04 (Stage 1 verified); migration 131 wired tool-generator PLAN writing, 132 wired fix-agent NOTES writing; first machine-written PLAN 2026-07-09 (migration 136); adopted as infrastructure by the separate fix-loop/diagnosis thread rather than that thread building a rival system.
- **what:** Every tool/complex component and every pipeline carries its own living documentation in two Postgres tables, keyed by (subject_type, subject_key): a PLAN (intent — aim, source spec, behaviour contract, acceptance criteria, delivery mechanism, dependencies, deliberate decisions) and a NOTES stream (every fix, diagnosis, and dead end). Agents write these as a byproduct of the steps that create and fix things, and load them before touching a subject, so fixes build on prior decisions instead of re-deriving lost context. Solves two failure modes: lost intent, and "deployed ≠ works". Successor to the file-based per-tool convention (DOC-003); its own umbrella covers many specific mechanisms registered separately below (DOC-011 through DOC-032).
- **sources:** RUNBOOK_travelling_docs(38).md#intro,§1; PLAN_travelling_docs(6).md#aim; OVERVIEW_self_verifying_tools.md#mechanism-1; docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; 125_doc_plans_and_notes.sql; 131_tool_generator_plan_writing.sql; 132_fix_agents_note_writing.sql
- **relations:** tool-doc header system (019, extended not replaced); DOC-003 (file-based predecessor); verification ladder; DOC-011 (supersede versioning); DOC-012 (append-only NOTES); DOC-017 (four doc actions); rag tool_docs indexing (DOC-027)
- **verify-later:** tables doc_plans, doc_notes in clients_db; migration sql_for_agents/125–146 (arc renumbered); actions in platform/orchestration/actions/write_doc_plan_action.go etc.; doc_notes row growth

### DOC-011 — doc_plans supersede versioning (one current row, never edit history)
- **status:** deployed
- **status-evidence:** RUNBOOK(38) §2 "supersede tx; one current row enforced by idx_doc_plans_current"; xp-curve PLAN chains v1→v2→v3 verified 2026-07-10.
- **what:** PLAN updates flip the current row to `is_current=false` + `superseded_at` and insert the new body as current; a partial unique index enforces exactly one current row per subject. History is never edited in place; rollback = restore a prior row; `pinned=true` is a human hold. The pattern is the `site_specs` supersede log re-keyed to the doc subject.
- **sources:** RUNBOOK_travelling_docs(38).md#§2; RUNNING_NOTES_travelling_docs(39).md#rev2; 0NN_supersede_xp_curve_plan_selectors(2).sql; write_doc_plan_action.go
- **relations:** site_specs supersede log (pattern source); EDIT-marker fill-by-supersede convention (DOC-028)
- **verify-later:** idx_doc_plans_current partial unique index; write_doc_plan supersede transaction in the action

### DOC-012 — doc_notes append-only log with jsonb category roll-up
- **status:** deployed
- **status-evidence:** RUNNING_NOTES(39) rev 2 "NOTES = table … Postgres serialises concurrent inserts"; GIN roll-up query in RUNBOOK §0-REF/§7.
- **what:** NOTES are one DB row per entry (never a shared file — a file append is a read-modify-write with lost-update risk under the retry-less git adapter). `categories jsonb` with a GIN `jsonb_ops` index makes `categories ? 'tag'` cross-tool roll-ups index-backed. `site_id` scopes per-site incidents. Entry format is uniform and dated (Observed / Root cause / Fix / Verified / Categories).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev2-GIN-RMW; PLAN_travelling_docs(6).md#table-design; RUNBOOK_travelling_docs(38).md#§3
- **relations:** NOTES category taxonomy (DOC-023); git-adapter constraints (why not git)
- **verify-later:** doc_notes schema + GIN index; roll-up queries in 016/016b

### DOC-013 — DB-as-truth storage decision (knowledge_base = derived index; git = optional mirror)
- **status:** deployed
- **status-evidence:** RUNNING_NOTES(39) rev 2 "DECISION: DB is the source of truth; git is an optional mirror"; PLAN(6) "Unchanged from rev 2".
- **what:** Postgres tables written transactionally by the framework are the record of truth; `knowledge_base` (content-hash keyed, no version chain) is only a derived retrieval index via `rag_index`/`rag_lookup`; git is a non-authoritative optional mirror for human browsing (Phase B, unbuilt). Grounded in git-adapter evidence: commits hard-reject empty Domain, force-prefix `{domain}/`, whole-file only, no read action, no conflict retry, all serialised through one Kafka adapter.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev2; PLAN_travelling_docs(6).md#storage-decision; RUNBOOK_travelling_docs(38).md#§1
- **relations:** supersedes the flat-file docs-repo decision (DOC-014); rag_index tool_docs collection (DOC-027)
- **verify-later:** adapter.go/github_client.go commit path; knowledge_base UNIQUE(collection, content_hash)

### DOC-014 — Abandoned: flat-file docs-repo as truth + docselect catalogue retrieval
- **status:** superseded
- **status-evidence:** RUNNING_NOTES(39) rev 1 "Rev-1 storage decision (later reversed) … Categories: (storage, superseded)"; the RUNBOOK rev-1 §4 "Making a doc retrievable (catalogue entry)" section vanishes from later revisions.
- **what:** The original (2026-07-04 rev 1) design: flat markdown files (`<docs-repo>/tools/<function>/PLAN.md`, `NOTES.md`) in a new writable docs repo as source of truth, RAG-indexed, plus a `DocRule` entry per tool in `diagnose_doc_catalogue*.json` so the code-diagnosis loop's `docselect.go` picks docs by keyword/path-glob. Reversed to DB-as-truth within the same day; the docselect route remains deferred for pipelines only ("needs the git mirror for files — Phase B").
- **sources:** PLAN_travelling_docs.md#design-decision-2 (rev 1); RUNBOOK_travelling_docs.md#§4 (rev 1); RUNNING_NOTES_travelling_docs(39).md#rev1
- **relations:** DB-as-truth decision (DOC-013, replacement); pipeline retrieval symmetry (docselect kept as a Phase-B idea for pipelines)
- **verify-later:** docselect.go DocRule selection; whether any doc catalogue entry for tool docs ever landed (expect none)

### DOC-015 — Doc subject convention — ('tool', function) and ('pipeline', build|content|design|maintenance)
- **status:** deployed
- **status-evidence:** Stage 0 gates 2026-07-04: live pipeline values build (3579), content (24), design (13), maintenance (2), no CHECK constraint; RUNBOOK §1.
- **what:** Docs key on `(subject_type, subject_key)`: tools by `content_components.function` byte-for-byte (survives forks — vindicated by the unique-index predicate covering active library originals only), pipelines by the four live `site_work_items.pipeline` values (convention, not schema — the column is unconstrained text). Generalising from tool_doc_* to subject-keyed tables was a deliberate rename made "while the migration was free".
- **sources:** RUNBOOK_travelling_docs(38).md#§1,#stage-0; RUNNING_NOTES_travelling_docs(39).md#rev3; verify_before_migration.sql
- **relations:** dangling-doc prevention rule (DOC-016); idx_cc_tool_function_unique
- **verify-later:** site_work_items.pipeline live values; content_components.function uniqueness predicate

### DOC-016 — The dangling-doc prevention rule — subject must be something the agent actually owns
- **status:** deployed
- **status-evidence:** Migration 137 applied 2026-07-09/10 ("append_note re-subjected ('tool', spec.function) → ('pipeline','build')"); PLAN(6) "Rollout outcomes" first bullet.
- **what:** A NOTES subject must reference an artifact the writing agent actually creates/owns. `tool-recreation-handler` writes page sections and never creates a `content_components` row, so a `('tool', spec.function)` note there would key a doc to a function no component owns — a dangling doc. It was re-subjected to `('pipeline','build')` + site stamp, mirroring component-template-fixer. Found by reading the definition, not by a failed run.
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev42-blocker-ii; HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§3
- **relations:** recreation-writes-page-sections fact; migration 137
- **verify-later:** tool-recreation-handler append_note config in agent_definitions

### DOC-017 — The four doc actions (write_doc_plan, append_doc_note, load_doc_context, persist_diagnosis_note)
- **status:** deployed
- **status-evidence:** RUNBOOK Stage 2 "✅ Go actions — ON PRODUCTION 2026-07-04 … deployed with their registry entries".
- **what:** The chassis-side write/read surface of travelling docs: `write_doc_plan` (supersede tx), `append_doc_note` (single INSERT), `load_doc_context` (current PLAN + latest-N NOTES + extracted `criteria_json`, composed as one prompt-ready block; `has_plan=false` is a normal state, not an error), `persist_diagnosis_note` (diagnosis output → NOTES). Conventions: prefixed InputSpec field names, error containment via `config.error_step`, pure-helper unit tests.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-2,#§6; RUNNING_NOTES_travelling_docs(39).md#rev4-drafts; write_doc_plan_action.go
- **relations:** all write hooks (DOC-018, DOC-019); criteria fenced block
- **verify-later:** platform/orchestration/actions/{write_doc_plan,append_doc_note,load_doc_context,persist_diagnosis_note}_action.go + registry.go entries

### DOC-018 — PLAN-at-birth write hook in tool-generator (compose_plan → write_plan → index_plan)
- **status:** deployed
- **status-evidence:** TASK 3 APPLIED 2026-07-07, PROVEN 2026-07-09 (run 1923badd: system wrote its own first PLAN, real selectors, fence intact, 2,982 chars).
- **what:** After `save_tool` succeeds, a Sonnet step composes a full PLAN body (standard checks verbatim; an interaction check ONLY from real selectors copied out of the generated HTML, never invented; ≤3,000 chars), `write_doc_plan` persists it (`source='tool-generator'`), and `rag_index` indexes it into `tool_docs`. Every doc step carries `config.error_step: "complete"` — docs can never fail tool creation. Composer later fixed by migration 144 (five → four standard checks, inline delivery).
- **sources:** RUNBOOK_travelling_docs(38).md#task-3,#task-3-proven; RUNNING_NOTES_travelling_docs(39).md#rev24,#rev33; HANDOFF_2026-07-10…md#§1
- **relations:** docs-never-fail-the-work containment (DOC-020); composer selector invention incident
- **verify-later:** tool-generator workflow (save_tool → compose_plan → write_plan → index_plan); doc_plans rows with source='tool-generator'

### DOC-019 — NOTES-at-every-fix hook on the three fix agents
- **status:** deployed
- **status-evidence:** TASK 4 APPLIED 2026-07-07, PROVEN 2026-07-09 (two machine-written `fix` notes from the economy-simulator recreation, 19:36:35Z and 20:33:04Z).
- **what:** `component-template-fixer`, `tool-improver`, and `tool-recreation-handler` each gained `compose_note → append_note` on their SUCCESS paths only (both fixer branches covered), with error containment to the terminal step. Subjects per agent: fixer → pipeline/build + site stamp; improver → tool/`tool_data.function`; recreation → re-subjected to pipeline/build (migration 137). Machine categories v1 = `["fix"]`; failure-class tags live in the body Categories line.
- **sources:** RUNBOOK_travelling_docs(38).md#task-4,#task-5-closed; RUNNING_NOTES_travelling_docs(39).md#rev26,#rev27
- **relations:** dangling-doc rule (DOC-016); acceptance iteration loop
- **verify-later:** the three agent workflows' compose_note/append_note tails; doc_notes WHERE categories ? 'fix'

### DOC-020 — "Docs never fail the work" containment principle — and its limit
- **status:** deployed
- **status-evidence:** Every doc step wired with `config.error_step` to the terminal step; corollary on record (HANDOFF 07-09 §5): "containment covers errors, not crashes or stalls".
- **what:** Documentation writes are strictly subordinate to the work: any doc-step failure routes to the workflow's terminal step so a fix/creation never fails because its documentation did. The limit was learned live: error containment protects against raised errors only — an OOMKilled pod or a stall raises nothing, so the step freezes instead of degrading (the index_plan incident).
- **sources:** HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§5; RUNNING_NOTES_travelling_docs(39).md#rev31,#rev33; 016b_debugging_guide_7_3_(5).md#§9
- **relations:** error_step-in-config mechanics; EXECUTING_STEP-forever pattern
- **verify-later:** config.error_step on all doc steps in the touched workflows

### DOC-021 — Pipeline documentation model — derive the topology, author the intent
- **status:** partial
- **status-evidence:** PLAN(6) Phase B item 9 "Pipeline PLAN bodies distilled from 004–008" still pending; the migration write-hook IS live ("live practice from 140 onward, not aspiration").
- **what:** A pipeline's step map is generated from `agent_definitions` (callgraph pattern) — never hand-drawn, so it can't drift. The authored pipeline PLAN holds only: end-to-end invariants (e.g. "interactive sections survive every rebuild route"), branch rationale, seams (pipelines sharing one handler is where seam bugs live), and deliberate decisions. Pipeline NOTES = incidents + migration entries + persisted diagnoses; 016/016b stays the global roll-up.
- **sources:** PLAN_travelling_docs(6).md#pipeline-documentation; RUNNING_NOTES_travelling_docs(39).md#rev3; RUNBOOK_travelling_docs(38).md#§2
- **relations:** migration write-hook (DOC-022); docselect Phase-B retrieval for pipelines; docs 004–008 as prose base
- **verify-later:** whether any pipeline PLAN body exists in doc_plans (subject_type='pipeline' with is_current)

### DOC-022 — Workflow-altering migrations write pipeline NOTES
- **status:** deployed
- **status-evidence:** "Workflow-altering migrations now leave the runbook-§3 pipeline note (140 and 141 both carry one)" — RUNBOOK rev 44; 145/146 also carry pipeline notes.
- **what:** Every migration that alters an agent workflow appends a `('pipeline','build')` doc_notes entry recording the migration number, what changed, and why — making the migration ledger part of the pipeline's travelling history. The 005 "SQL Migrations Applied" table was identified as the embryo of this write hook.
- **sources:** RUNBOOK_travelling_docs(38).md#task-5-closed; PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#2026-07-10-migrations
- **relations:** migrations system (ledger + runner); doc_notes migration category
- **verify-later:** doc_notes rows with categories containing 'migration' from 140 onward

### DOC-023 — NOTES category taxonomy
- **status:** deployed
- **status-evidence:** RUNBOOK(38) §3 lists the taxonomy as operating vocabulary.
- **what:** The tag set for NOTES roll-ups: `css-variable-mismatch`, `empty-shell`/`mode-b-template`, `broken-template-slots`, `content-vs-runtime-mismatch`, `detool-on-rebuild`, `js-not-extracted`, `js-bundle-stale`, `schema-template-drift`, `diagnosis`, `unconfirmed-diagnosis`, `migration`, `seam`, `acceptance-run`, `acceptance-fail`, `truncated-output`, `needs_criteria`. Extends 037's taxonomy (DOC-003); GIN-queryable.
- **sources:** RUNBOOK_travelling_docs(38).md#§3; PLAN_travelling_docs(6).md#document-formats
- **relations:** doc_notes jsonb roll-up (DOC-012); 037 documentation-system conventions (DOC-003)
- **verify-later:** live distinct categories in doc_notes

### DOC-024 — Deliberate-decisions sections + the graduation rule (prose → structured → enforced)
- **status:** deployed
- **status-evidence:** PLAN(6) "Graduation rule: prose → structured → enforced, only when recurrence proves the need. (Locks stay deferred; criteria live as a fenced block, not a column, until a checker consumes them at volume.)"
- **what:** A PLAN carries a "Deliberate decisions — do not re-fix" prose section, protective because it is loaded at fix time; no lock mechanism yet. Knowledge graduates from prose to structure to enforcement only on demonstrated recurrence — the reason criteria are a fenced block rather than a column, and locks are deferred. Runbook prose is "un-compiled residue" that retires as it is compiled into guards/fixes.
- **sources:** PLAN_travelling_docs(6).md#framing,#deliberate-decisions; RUNNING_NOTES_travelling_docs(39).md#rev3-framing
- **relations:** framing concept (DOC-025); locks category (031)
- **verify-later:** whether any lock/enforcement mechanism for deliberate decisions has since appeared

### DOC-025 — Framing: plan = enforced desired state; pipeline = compiled runbook; NOTES = the reasoning log
- **status:** deployed
- **status-evidence:** "Framing (agreed 2026-07-04)" section, stable across PLAN revs 3–6.
- **what:** Where each artifact sits within the Postgres travelling-docs system: `site_plans`+specs are the ENFORCED plan (reconciler drives realised state toward it — "the plan table is ground truth; the rest is weather"); the build pipeline is the compiled happy-path runbook; written runbooks are the un-compiled residue (exception knowledge); NOTES is the reasoning log nothing machine-side captures; contracts/constitution sit above as admission rules.
- **sources:** PLAN_travelling_docs(6).md#framing; RUNNING_NOTES_travelling_docs(39).md#rev3
- **relations:** site-plan-and-reconciler (030); graduation rule (DOC-024); contracts-and-standards
- **verify-later:** n/a (conceptual framing) — cross-check with docs 030/016b claims in their units

### DOC-026 — load_doc_context fix-time retrieval
- **status:** deployed
- **status-evidence:** Deployed 2026-07-04 (Stage 2); consumed live by the Tier-2 checker and tool-acceptance-agent.
- **what:** The primary direct-by-key read: current PLAN + latest-N NOTES + extracted `criteria_json`, composed into one prompt-ready `doc_context` block. `has_plan=false` is a normal state. For the code-diagnosis loop, `doc_context` is to be handed to `diagnose_assemble_bundle` the way `runtime_evidence` is (one compose line) — `rag_lookup` is discovery-only (no function filter).
- **sources:** RUNBOOK_travelling_docs(38).md#§6; PLAN_travelling_docs(6).md#retrieval; RUNNING_NOTES_travelling_docs(39).md#rev2
- **relations:** four doc actions (DOC-017); criteria fenced block; diagnose_assemble_bundle injection (still unwired — verify)
- **verify-later:** load_doc_context_action.go; whether the diagnosis bundle now includes doc_context

### DOC-027 — tool_docs knowledge-base indexing of PLANs (rag_index derived index)
- **status:** deployed
- **status-evidence:** TASK 6 CLOSED 2026-07-10: run 05d1fc97 wrote the FIRST knowledge_base collection='tool_docs' rows (4 chunks, 4 embeddings, ~5.5s) after the chunkContent fix + migration 141 re-enable.
- **what:** After each PLAN write, `rag_index` chunks and embeds the body into the `tool_docs` collection for semantic discovery. The 019 claim that generation already wrote tool_docs was verified UNIMPLEMENTED (a standing open from day 1); the write first became real 2026-07-10. Standing open: `rag_index` hardcodes `source_type='scrape'` (parameterisation open item).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev1-open-thread,#task-6-closed; RUNBOOK_travelling_docs(38).md#position-2026-07-10
- **relations:** chunkContent infinite-loop bug; DB-as-truth (DOC-013); rag_lookup discovery
- **verify-later:** knowledge_base rows collection='tool_docs'; rag_actions.go source_type parameter

### DOC-028 — EDIT-marker / -EDIT check-id convention (honest unknowns in seeded docs)
- **status:** deployed
- **status-evidence:** RUNBOOK(38) mini-glossary; Tier-2 checker implements "-EDIT skipped" (unit-tested).
- **what:** Fill-in-the-blank markers for details not known at seeding time: `EDIT:` prose markers are optional fill-later blanks (doc valid meanwhile; fills arrive by supersede, never in-place edits); acceptance checks whose id ends `-EDIT` carry placeholder selectors and are skipped by every verification tier until real selectors replace them.
- **sources:** RUNBOOK_travelling_docs(38).md#mini-glossary; RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23
- **relations:** pilot PLAN seeding (DOC-029); anchor rule (fail ⇒ drop or mark -EDIT)
- **verify-later:** -EDIT handling in discovery_checks/check_tool_acceptance.go and the browser-runner

### DOC-029 — Pilot PLAN seeding by SQL (dogfooding the format)
- **status:** deployed
- **status-evidence:** TASK 2A DONE 2026-07-07 12:32 — first tool PLAN live for `tool-archetype-taster-quiz` (is_current=t, has_fence=t, 2,761 chars).
- **what:** Before any workflow wiring existed, the first real tool PLAN was seeded by a hand-written dollar-quoted INSERT (source='human', created_by='pilot'), satisfying Stage-5's precondition (≥1 tool PLAN with criteria) and road-testing the section format. Later `write_doc_plan` calls supersede it cleanly. Includes a seeded deliberate decision ("exactly THREE questions — the taster must not be improved into the Gauntlet").
- **sources:** RUNBOOK_travelling_docs(38).md#pilot-plan; RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23
- **relations:** EDIT markers (DOC-028); acceptance-criteria-in-PLAN decision
- **verify-later:** doc_plans row for tool-archetype-taster-quiz (superseded chain)

### DOC-030 — Provenance stamps the chassis, not the logical agent — config-declared source is the reliable provenance
- **status:** deployed
- **status-evidence:** "Provenance stamps the CHASSIS: … source_agent_type='generic' … the planned doc-action fallback is DROPPED — backlog item closed by evidence rather than by code" (rev 32).
- **what:** Both `Headers["agent_type"]` (empty in step context) and `ExecutionContext.Sender.AgentType` (stamps 'generic' — the shared chassis pod) fail to identify the logical agent. Doc rows therefore rely on the config-declared `source`/`plan_source`/`note_source` fields for provenance, which the actions already carry. Applies equally to `content_components.source_agent_type` and rag_actions.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev31-watch,#rev32; HANDOFF_2026-07-09_recreation_and_chassis_1_.md#§5; 016b_debugging_guide_7_3_(7).md#schema-traps
- **relations:** four doc actions (DOC-017); component provenance columns
- **verify-later:** source vs source_agent population on live doc_plans/doc_notes rows

### DOC-031 — Handoff-document discipline (updated-every-turn, supersede chain, turn log) — travelling_docs thread
- **status:** deployed
- **status-evidence:** Three-generation chain in this unit (07-08 → 07-09 → 07-10, each "supersedes" the prior); 07-10 handoff carries an 11-entry turn log through 2026-07-13.
- **what:** Long-running work travels between chat sessions via a standing HANDOFF doc: first-actions list, state-of-play with dates and snapshot ids, blocker sections with ranked hypotheses and data-to-collect checklists, durable rules, key identifiers, file inventory, and a newest-first turn log updated EVERY turn. Companions pattern: RUNBOOK (position tracker) + RUNNING_NOTES (chronology) + PLAN (spec) + 016b (durable patterns) — the travelling-docs idea applied to the work itself. Includes the cross-workstream "collision rule" courtesy FYI when another chat touches a shared surface.
- **sources:** HANDOFF_2026-07-08…md; HANDOFF_2026-07-09…_1_.md; HANDOFF_2026-07-10…md#turn-log; FYI_from_fixloop…md
- **relations:** doc traveller / docs037 conventions (DOC-003); bundle command; the general travelling-docs pattern (DOC-052); vonc's own HANDOFF convention (DOC-060)
- **verify-later:** n/a (working practice)

### DOC-032 — Standing opens ledger of the travelling-docs arc
- **status:** partial
- **status-evidence:** Repeated "Standing opens" list in every rev's open threads and HANDOFF §5, still open at unit close.
- **what:** The carried-forward small items: `deploy_tool_to_site` should stamp `source_*` on forks (NOTES-only on fork — unverified); `rag_index` `source_type='scrape'` parameterisation; the Tier-4 vocabulary "select" verb; P1 mobile / P2 interactions; a real (non-manufactured) acceptance failure through tool-improver and back; github-actions-runner cgroup-driver CrashLoopBackOff (infra, not app); chassis memory slope watch (leak neither proven nor needed after the chunkContent answer).
- **sources:** HANDOFF_2026-07-10…md#§4,§5; RUNNING_NOTES_travelling_docs(39).md; RUNBOOK_travelling_docs(38).md#background
- **relations:** most DOC-010-series concepts above
- **verify-later:** each item individually in stage 2

### DOC-033 — Context-bundle seeding for fresh agent threads (imagery)
- **status:** deployed
- **status-evidence:** RUNBOOK A4 standing ritual: "z_bundles/imagery_seed_docs/imagery_bundle.sh… Output lands at z_bundles/imagery_bundle.md"; used to seed Turn 1.
- **what:** A repeatable script assembles the imagery workstream's context bundle (key docs + live schema/runtime sections queried from the cluster) for cold-starting a fresh agent session; run after credential refresh or when the DB sections come out empty. Paired with the document-set discipline: PLAN (map) / RUNNING_NOTES (turn-by-turn evidence) / RUNBOOK (human task queue) / HANDOFF (single cold-start entry point, updated every turn) / SHOWCASE (shareable summaries).
- **sources:** RUNBOOK_imagery_best_in_class.md#A4; HANDOFF_imagery_best_in_class.md#Document-map; CONTEXT_PACK_imagery_sprite_sheet.md
- **relations:** documentation-system conventions (travelling docs, DOC-010); CONTEXT_PACK as the sprite-specific instance
- **verify-later:** z_bundles/imagery_seed_docs/imagery_bundle.sh existence

### DOC-034 — Traffic-probe context packaging (docubundle)
- **status:** deployed
- **status-evidence:** package_traffic_probe.sh header + its output traffic-probe_context.txt (445 KB) and repo_summary bundles present in the unit.
- **what:** A self-contained packager bundles the task brief, domain list, reusable Go service, and deploy/persistence/VM docs into one context file so a new chat can start cold — coping with the messy versioned folder by resolving each doc to the newest (N) variant by mtime and dropping *.orig* backups. Companion scripts (outputtotext.sh, reduce_output_dir.sh) flatten captured site directories into repo_summary.txt bundles. The same cold-start pattern produced the HANDOFF file for the permanent thread.
- **sources:** package_traffic_probe.sh; docubundle/output_contexts/relojistas/outputtotext.sh; HANDOFF_vm_sites_permanent_thread.md
- **relations:** documentation-system (context packaging, travelling docs); per-domain notes convention
- **verify-later:** n/a (tooling snapshot)

### DOC-035 — Single-source relocation with pointer / canonical-doc-home discipline
- **status:** deployed
- **status-evidence:** Two independent content areas (adapter deployment mechanics, adapter response envelope shape) were both removed from their original host docs and consolidated into `035_adapter_guide.md`, with the live contracts doc leaving a one-line "Moved to X, now the single source for it" pointer; concretely dated 2026-06-11: "FOCUS_adapter_design merged into 035 — FOCUS retired".
- **what:** A recurring documentation practice: when a topic is found duplicated across docs (a debugging guide and a contracts doc; or 003 vs FOCUS_adapter_design — the "003 vs FOCUS contradiction" root-caused as duplication-then-drift, not a genuine disagreement), maintainers consolidate it into ONE canonical numbered doc and replace the other locations with a short pointer sentence, rather than letting copies drift out of sync. Numbered docs (001/002/003/019/020…) are canonical/permanent; `FOCUS_*` docs are transient design notes meant to be retired once their content graduates. A validator to enforce this mechanically was proposed but not built.
- **sources:** docs024_key_docs_latest/003_contracts_and_standards(8).md; docs024_key_docs_latest/035_adapter_guide.md; debugging_old/016_debugging_guide_v2(1).md; NOTES_running_synthesis_principles(59) "Doc restructure: adapter docs + 003"
- **relations:** adapters; documentation-system; 000_documentation_index (DOC-001); doc claim-verification convention (DOC-040)
- **verify-later:** check 000_documentation_index.md/travelling_docs conventions for whether this is a formal rule

### DOC-036 — Full heading+content-line diff across all forked copies before consolidating a travelling doc
- **status:** deployed
- **status-evidence:** Used explicitly twice: v2_58 ("A full heading-level AND content-line diff across all 14 files confirmed these were the ONLY entries missing") and the 016b consolidation ("Verified against ALL forked copies... proved this copy already contains every one of the 9 distinct §9 entries").
- **what:** A consolidation methodology: before promoting one copy of a travelling/forked doc to canonical, diff it against every other known fork at both heading and content-line granularity, explicitly asserting "no content was removed," and recover anything found missing.
- **sources:** docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md; docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md
- **relations:** the method's completeness claim does not always hold in practice — see the diagnosis-loop fork, which the 016b audit's own "verified against ALL forks" claim did not actually catch
- **verify-later:** none code-related — a documentation-process note for docs026 itself

### DOC-037 — verify_before_migration pre-flight convention
- **status:** deployed
- **status-evidence:** "Follows the travelling-docs precedent (verify_before_migration.sql)" (verify_before_migration_diagnosis_artifacts.sql header).
- **what:** A house convention (shared with the travelling-docs workstream) of writing a companion pre-flight SQL script before any hand-applied migration, checking for table/index name collisions and confirming assumptions about existing constraints, with results pasted back into the running notes doc.
- **sources:** fixloop_eg_dartsonline/verify_before_migration_diagnosis_artifacts.sql; fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 2
- **relations:** diagnosis_artifacts table
- **verify-later:** n/a — process/design record

### DOC-038 — doc_notes / travelling-docs integration boundary
- **status:** deployed
- **status-evidence:** "Q-F's integration is verified live: our envelope's subject fields opened the tools chat's persist_note gate" (NOTES(10)#Turn 6).
- **what:** The fix-loop's terminal diagnosis note lands in a separate workstream's `doc_notes` table via that workstream's own `persist_note` wiring, gated on the envelope carrying `subject_type`/`subject_key`. The fix-loop treats the diagnose-agent workflow JSON as an active surface owned by that other workstream — any change is fetch-first, snapshotted, with a written FYI. Per-iteration/per-step notes (F0.3) were designed to reuse this table but never built.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#What already exists; fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 6,#Turn 11
- **relations:** F0.3 per-iteration notes (never built); symptom-closure gate; Postgres travelling-docs system (DOC-010)
- **verify-later:** grep/inspect doc_notes; persist_note; subject_type

### DOC-039 — Guideline-compliance review methodology (001/002/003 walkthrough before shipping)
- **status:** deployed
- **status-evidence:** used as the actual pre-ship review for the js_snippets deliverable.
- **what:** A documented review convention: before applying new Go actions + migrations, walk every guideline document point by point against the actual deliverable — reuse-before-creating check, canonical field-path helpers, action/wrapper split rules, spawn-before-call pattern, logging conventions — producing a test plan as the final artifact.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md; js_snippets_news_gaswholesalers/old/guidelines_compliance_check(1).md
- **relations:** js_snippets site-level rendering pipeline
- **verify-later:** n/a — process/design record

### DOC-040 — Doc claim-verification / dated-claim convention
- **status:** deployed
- **status-evidence:** "DONE 2026-06-11: claim-verification discipline CODIFIED (016 v2_34 item 24 + 001 pointer + dated-claim convention)" (principles(59)).
- **what:** Falsifiable, load-bearing doc claims carry a `[checked YYYY-MM-DD: <evidence>]` tag (one date = last checked, updated in place); negative claims ("X isn't built") carry their falsifying command; whole-document "verified" stamps are explicitly banned (verification attaches to claims, never documents); docs update in the same change as the decision that makes them true. Motivated by a real incident where a stale negative claim in doc 019 nearly caused a reuse-before-recreate violation, and the team's own freshly-written docs went stale within hours — "staleness is a coupling property, not an age property."
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 "DISCUSSED... doc up-to-dateness" and "CODIFIED" entries
- **relations:** doc-drift claim classifier (DOC-041); canonical-doc-home discipline (DOC-035)
- **verify-later:** n/a

### DOC-041 — Doc-drift claim classifier (design only, named across three units, never built)
- **status:** aspirational
- **status-evidence:** First named 2026-06-13 in the docs019 principles notes ("doc-drift claim classifier (DESIGN_doc_drift_classifier.md) is design-only, gated on testing vs known bugs"); given a full design doc (DESIGN_doc_drift_classifier.md: "Status: design only. The prompt contract (§3) is the part that must be right before any code."); still referenced as "not built" in the later docs-archiving runbook ("that is the doc-drift classifier … a separate, later, T1+T2-first tool").
- **what:** A designed-but-never-built tool that classifies individual doc claims as current/stale/unverifiable against the real system, under an evidence-or-abstain contract (quote or no verdict; no proposed rewrites; unverifiable routes to keep-untouched), gathering evidence at the shallowest settling tier (T1 static code/schema, T2 read-only DB state, T3 behavioural — existing logs only, never triggering a run). Two hard rules: read-only, and abstention asymmetry (behavioural evidence supports "stale" only on direct contradiction, since misattributing an unrelated bug to a correct doc is worse than silence). Explicitly classify-don't-merge: it can check a claim but must never generatively rewrite/merge docs, since a rewrite can silently drop caveats no code-check would catch. Historically the parent of the diagnosis loop's cite-or-abstain verdict contract. Consistently deferred across all three sightings spanning at least a month.
- **sources:** DESIGN_doc_drift_classifier.md (full); NOTES_running_synthesis_principles(59) 2026-06-13; contextkit/RUNBOOK_doc_archiving.md#What-this-subproject-does-NOT-do; DESIGN_diagnosis_loop(3).md
- **relations:** doc claim-verification convention (DOC-040); docs archiving toolchain (DOC-046); claim taxonomy (DOC-042); classify-don't-merge principle (DOC-043); date/version-as-triage (DOC-044); standing conformance suite (DOC-045)
- **verify-later:** whether any classifier code exists

### DOC-042 — Claim taxonomy: code-checkable / superseded-but-not-wrong / code-invisible
- **status:** aspirational
- **status-evidence:** DESIGN_doc_drift_classifier §1 ("carried from item 24") — design only.
- **what:** Three buckets of doc claims by checkability: mechanically confirmable facts (the classifier's target); decisions whose holding-but-not-rationale the code confirms (partial signal); and design intent / negative results the code says nothing about — disproportionately why old docs are worth keeping. Buckets 2/3 must reliably route to keep-untouched, never a confident verdict.
- **sources:** DESIGN_doc_drift_classifier.md#1
- **relations:** doc-drift classifier (DOC-041); classify-don't-merge (DOC-043)
- **verify-later:** n/a (design)

### DOC-043 — Classify, do NOT merge (the human consolidates)
- **status:** deployed
- **status-evidence:** DESIGN_doc_drift_classifier §5 "the line held firmest"; echoed as working practice in engines_tree_proposal ("selective-carry-with-the-LLM-as-assistant, never a generative merge").
- **what:** Grounding makes checking tractable but does not make generative merging safe: an LLM rewriting N docs into one fails silently (a dropped caveat reads as clean prose; no code-check catches an omission). The tool finds and cites; the human decides and writes; every canonical doc stays human-authored. Applied as a standing rule across the doc work — including, per this project's own instructions, the docs026 consolidation stage itself.
- **sources:** DESIGN_doc_drift_classifier.md#5; engines_tree_proposal.md
- **relations:** doc-drift classifier (DOC-041); engines tree migration (DOC-051)
- **verify-later:** n/a (principle)

### DOC-044 — Date/version as triage, not truth
- **status:** deployed
- **status-evidence:** DESIGN_doc_drift_classifier §4.1 — stated as a settled refinement.
- **what:** A recent file is more likely current; an old file is not more likely wrong — it is more likely unchecked. Dates order the queue and break ties; they never override a code check (recent docs went stale within hours in observed cases). Code decides; date orders.
- **sources:** DESIGN_doc_drift_classifier.md#4.1
- **relations:** doc-drift classifier (DOC-041); misattribution asymmetry
- **verify-later:** n/a (principle)

### DOC-045 — Standing conformance suite (carved out, deliberately not built)
- **status:** aspirational
- **status-evidence:** DESIGN_doc_drift_classifier §0/§6: "CARVED OUT … Built later, on demand, as its own thing."
- **what:** A continuous "does the live system behave as documented" monitor on the existing DiscoveryCheckContext/CheckResult rails, scheduled, allowed fenced probes the doc pass forbids. Deliberately separated from the one-off classifier so the heavyweight always-on thing doesn't get built under a cleanup's banner and sink both.
- **sources:** DESIGN_doc_drift_classifier.md#0,#6
- **relations:** doc-drift classifier (DOC-041); improvement-loop checkers
- **verify-later:** n/a (unbuilt)

### DOC-046 — Documentation archiving subproject (docs019 cleanup: dedup + thin_versions + staged migration)
- **status:** partial
- **status-evidence:** Measured 2026-06-13: "2,729 files collapse to 1,917 subjects; 1,734 are singletons; the noise is concentrated in 18 fat clusters of 10+ versions each (~489 files)"; the same-period principles notes still logged "Doc archiving — tools built + tested, RUN not yet done"; but docs/dedup-manifest.tsv and docs/thin-manifest.tsv in the live docs/ root already contain completed move rows (moved_from → _archive/... destinations), while docs/PROPOSED_MOVES.tsv still carries an unfilled editorial ACTION column.
- **what:** A four-step, report-first, fully-reversible plan to de-noise the docs tree so the analyser's index isn't diluted: dedup (exact/near-duplicate copies) → thin_versions (older versions) → editorial re-home of surviving current docs into an engines/+runbooks/ structure (human judgement, "classify don't merge") → re-point links and rebuild the analyser index, verified by a python one-liner checking zero stale/archived paths remain indexed. Explicitly out of scope: merging documents, judging content currency (deferred to the doc-drift classifier), touching the 1,734 singleton subjects. Execution status is mixed: manifests in the live docs/ root prove the deterministic dedup/thin passes DID run against at least the top-level docs tree, but the editorial third step (PROPOSED_MOVES.tsv) is still awaiting human keep/move/archive decisions, and it is unconfirmed whether the docs019-specific pass described in the runbook is the same run or a separate, still-unexecuted one.
- **sources:** contextkit/RUNBOOK_doc_archiving.md; NOTES_running_synthesis_principles(59) 2026-06-13; docs/dedup-manifest.tsv; docs/thin-manifest.tsv; docs/PROPOSED_MOVES.tsv; docs/main_docs_directory_tree.txt
- **relations:** dedup tool (DOC-047); thin_versions tool (DOC-048); docs019 migration staging script (DOC-049); doc-drift classifier (DOC-041, deliberately separate); documentation-system (037 conventions, DOC-003)
- **verify-later:** whether the docs019-specific pass was ever executed as distinct from the docs/ root pass; docs/_archive contents vs the manifests; whether engines/ tree was created

### DOC-047 — dedup (cmd/dedup)
- **status:** deployed
- **status-evidence:** "these tools were behaviour-tested, but they MOVE FILES" (RUNBOOK_doc_archiving.md); default-report/`-move` design documented as used in the docs019 archiving runbook.
- **what:** Finds duplicate/near-duplicate files in a directory tree; two passes — EXACT (SHA-256, deterministic) and NEAR (optional, shingled-token Jaccard similarity ≥ threshold, heuristic). Report-only by default; `-move` relocates non-canonical copies into an archive dir with a full undo manifest (TSV), never deletes. Canonical-selection tie-break: not-archived > not a `(N)` download-dup > shallowest > shortest > newest.
- **sources:** contextkit/cmd/dedup/main.go#header; contextkit/RUNBOOK_doc_archiving.md#Step-1
- **relations:** thin_versions (DOC-048, distinct: copies vs versions); stage_docs019_migration.sh (DOC-049, delegates to it)
- **verify-later:** dedup-manifest.tsv output location and whether the docs019 archiving pass was actually executed against the live docs tree

### DOC-048 — thin_versions (cmd/thin_versions)
- **status:** deployed
- **status-evidence:** "go build ./cmd/dedup ./cmd/thin_versions    # both compile" (RUNBOOK_doc_archiving.md pre-flight).
- **what:** Reduces version-sprawl by grouping files that are successive versions of the same document (subject-stem derived by stripping `.patch/.orig/.bak`, extension, trailing `(N)` bracket, trailing `_vX`/`_vX_Y`, then a second `(N)`), keeps the newest N per group, and moves older versions to an archive dir on request. Recency rank within a subject is version-number first, then `(N)` bracket, then mtime — deliberately so a stale-dated-but-later version still ranks above an earlier one.
- **sources:** contextkit/cmd/thin_versions/main.go#header; contextkit/RUNBOOK_doc_archiving.md#Step-2
- **relations:** dedup (DOC-047, run first to clear exact copies before thinning versions)
- **verify-later:** whether the docs024 "18 fat clusters of 10+ versions each" identified in the runbook were actually thinned

### DOC-049 — docs019 migration staging script (stage_docs019_migration.sh)
- **status:** partial
- **status-evidence:** "DEFAULT IS REPORT ONLY. Nothing moves without --apply. Even --apply touches only (1); (2) needs --dedup; (3) is never auto-applied." (header)
- **what:** Automates the deterministic half of the docs019 archiving plan: auto-moves obviously-superseded archive directories (`go_files_old/`, `thin_slice_run/`, `working/`, a stray `archive_april_26/`) into `_archive/` on `--apply`; can also invoke the dedup tool (`--dedup`); and for the editorial third (loose `FOCUS_`/`PLAN_`/`RUNBOOK_`/`016_`/`NOTES_` docs), only writes a `PROPOSED_MOVES.tsv` for a human to edit (action column: move|archive|skip|keep) and apply by hand — deliberately never auto-applied.
- **sources:** contextkit/cmd/dedup/stage_docs019_migration.sh#header
- **relations:** dedup (DOC-047); thin_versions (DOC-048); documentation archiving subproject (DOC-046)
- **verify-later:** —

### DOC-050 — Bundle-first handoff practice (context packs; broad script vs lean assembler)
- **status:** deployed
- **status-evidence:** The tasks/ folder is this practice operating: one-sentence descriptions + primed bundle commands per task; the GUIDE compares the two gathering modes on real artefacts (1.7MB script vs 30KB bundle).
- **what:** Every task handoff pairs a one-sentence problem statement with a filled-in cmd/bundle invocation (scope, docs, schema tables, runtime target), so a fresh chat starts from assembled context. Two gathering modes with a stated tradeoff: the package_*.sh directory-concatenating script (broad, thorough, catches wiring like registry.go, over-includes) vs the analyser/assembler (narrow, lean, call-graph-blind to wiring). Manual expert manifests were used as ground truth to validate the tool ("we're automating what experts already do by hand"). Advanced form: self-resolving bundle scripts that grep the analysis to locate action files (bundle_minilobby_trim v2's resolver, with PIN_ overrides).
- **sources:** tasks/any_project_handoff/001_build_bundle_ask_for_handoff; GUIDE_deploy_from_context_packs(1).md; tasks/vonc_provocations_lobby/bundle_minilobby_trim(3).sh; tasks/missing_game_on_games_page/001_bundle
- **relations:** travelling-docs pattern (DOC-052); cmd/bundle robustness; constitution-in-every-bundle; cold-start documentation bundle practice (DOC-009)
- **verify-later:** n/a (practice)

### DOC-051 — Engines docs tree + single _archive graveyard
- **status:** aspirational
- **status-evidence:** engines_tree_proposal: "a TARGET to migrate toward … not a big-bang restructure"; _archive/ dirs now exist in git status (partially enacted).
- **what:** Three kinds of thing kept apart: engine code (in the module), engine docs (one canonical file per engine under engines/, pointing at canonical sources rather than restating), and archive (one _archive/ root, never indexed, replacing the go_files_old/docubundle/(N).go sprawl; the dedup tool's default target, giving the analyser a single -exclude). Runbooks split from engine docs because how-to-run rots at a different rate than how-it-works. Migration: dedup report → move → human editorial consolidation → re-point links → re-index.
- **sources:** engines_tree_proposal.md
- **relations:** classify-don't-merge (DOC-043); dedup tool (DOC-047)
- **verify-later:** whether engines/ was created; _archive contents

### DOC-052 — Travelling-docs pattern (runbook = plan, notes = history, handoff = session)
- **status:** deployed
- **status-evidence:** Used throughout: HANDOFF_fixloop(8) header names it; running_notes_2 "Memory is OFF; this doc is the journal. Present this file at the END OF EVERY TURN."
- **what:** Each long-running thread maintains a runbook (the plan/map), running notes (chronological decisions with rationale, checkpointed), and a handoff (the complete start state for a fresh context, updated as discussion takes positions, with a file manifest and an opening move). Handoffs restate standing rules every time — the manual precursor the constitution automates. Parallel threads carry explicit boundary-awareness sections (what NOT to work on here).
- **sources:** HANDOFF_fixloop_thread(8).md; HANDOFF_builder_thread.md; tasks/005site_scheme_palette_and_components/running_notes_2(5).md; tasks/005site_scheme_palette_and_components/HANDOFF_scheme_to_components.md
- **relations:** bundle-first handoffs (DOC-050); three-thread working (DOC-053); constitution
- **verify-later:** n/a (practice)

### DOC-053 — Three parallel threads with hard boundaries
- **status:** deployed
- **status-evidence:** HANDOFF_builder_thread §1: "THREE PARALLEL THREADS with hard boundaries" (builder/spine, tools, site-quality), with joint decisions marked ON HOLD.
- **what:** Concurrent chat threads own non-overlapping territories (relay spine + coordination; tool-pipeline internals; page-facing quality), each with its own runbook/notes; cross-territory scope changes route back through the owning thread; joint seams (e.g. the planned-tool-page seam §B5) are explicitly flagged as joint decisions and parked. Boundary files ride in each thread's manifest read-only.
- **sources:** HANDOFF_builder_thread.md#1,#4; HANDOFF_fixloop_thread(8).md
- **relations:** travelling docs (DOC-052); classifier-consolidation queue
- **verify-later:** n/a (practice)

### DOC-054 — Concept register and the council-of-concept-experts mission
- **status:** aspirational
- **status-evidence:** README_comprehensive_documentation_categorisation is the originating user prompt (extract → classify → later verify against code → later create per-concept council agents); stage 1 in progress as this register.
- **what:** The user's programme: sweep every docs/ file for concepts (aspirational, deployed, superseded, unfulfilled), classify them into docs024-style categories, later verify each concept's true state against chassis code/workflows/DB, and ultimately seed an expert agent per concept area to join the diagnosis/fix-loop council. Documentation categories are intended to correlate with council-reviewer expertise areas. This docs026 project is itself the concept being described.
- **sources:** README_comprehensive_documentation_categorisation.md; README_claude_conversation.md
- **relations:** expanded council bench; docs024 documentation index (DOC-001)
- **verify-later:** n/a (this project)

### DOC-055 — Four-layer documentation model for automation
- **status:** aspirational
- **status-evidence:** MASTER(4) §5 "Four layers — two existing, two new — plus governance".
- **what:** Automation's connective tissue is four doc layers plus governance across them: the existing standards tree (prescriptive), the existing authored/derived context substrate, a NEW known-good solution library, and a NEW trust ledger; governance (curators/coordinator/advocate) sits across all four.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5
- **relations:** atomic standard; authored-vs-derived context (DOC-058); known-good library; trust ledger; standards curation
- **verify-later:** none

### DOC-056 — Published reasoning as substrate + drift detection
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §11 "Every decision publishes its reasoning, not just its outcome … Drift is the gap between a decision's logged premise and the current premise".
- **what:** Every decision publishes the premise it rested on, not just the outcome, upgrading the trace into reusable reasoning and training signal. This is what makes drift detection possible: a detector flags where a standard's stated premise no longer matches the premises recent decisions were actually made on.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#10,#11
- **relations:** priority profile; mediator trace; confirm-not-initiate
- **verify-later:** none

### DOC-057 — Objective tree vs concern tree (two orthogonal axes)
- **status:** aspirational
- **status-evidence:** FOCUS_best_practice_doc_tree(1) §1 "Two orthogonal structures, kept separate and cross-referenced, never merged".
- **what:** Two doc structures kept separate: the vertical objective tree (mission→branch→leaf) read downward, and the horizontal concern tree (best-practice cross-cutting standards) consulted sideways for any change. Each objective node carries `standing_concerns` linking to the standards it always pulls.
- **sources:** ED/FOCUS_best_practice_doc_tree(1).md#1,#2.5
- **relations:** atomic standard; why-chain; concern curators
- **verify-later:** agent_definitions.domain_tags (proposed storage)

### DOC-058 — Authored vs derived context (one substrate, change layer between)
- **status:** aspirational
- **status-evidence:** FOCUS_context_authored_derived_change.md#2 "The distinction that does matter: authored vs derived"; §3 "The change layer between".
- **what:** Everything grounding a reasoning step is retrievable evidence; the load-bearing split is authored (owned, maintainable, can drift/be-wrong) vs derived (emitted by the system running, no owner). A third change layer (diffs of code/logs) is derived-but-narrative and is the natural audit/learning surface. Two staleness modes get two fixes: keep authored thin+pointer-rich; fetch derived at reasoning time.
- **sources:** ED/FOCUS_context_authored_derived_change.md#2,#3,#4,#5
- **relations:** four-layer doc model (DOC-055); salience over presence; known-good library
- **verify-later:** none

### DOC-059 — Debugging-guide fork-and-merge maintenance (cumulative 016b copy)
- **status:** deployed
- **status-evidence:** merged v5 changelog line 2026-07-04: "Guide had forked across chats; this is the cumulative version"; HANDOFF §9 item 10: "Apply 016b_debugging_guide_merged.md to the project."
- **what:** The 016b debugging guide is maintained across parallel chat threads and FORKS: the project copy and thread working copies diverge (each gaining entries the other lacks). Practice: merge into a cumulative copy (v5 folded three vonc-thread entries + the silent-noop entry into the parallel-chat version), version-stamp the changelog, and apply the merged copy back to the project. The docs/ root copies (guide_6_, merged(2)/(3)) are these thread artifacts; their unique-to-thread entries are the deferral drop, artifact-not-pod-logs, marker anchoring, silent no-op, hidden-vs-author-CSS, plus parallel-chat entries (two chrome paths, SQL pitfalls, sites.status).
- **sources:** docs/016b_debugging_guide_merged(3).md#v5-changelog; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-guide-merged; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** travelling docs convention (DOC-010); category tag roll-up (DOC-023); DOC-003
- **verify-later:** which 016b copy the docs024 consolidated guide corresponds to; whether the merge was applied

### DOC-060 — Handoff document convention (stand-alone dated brief for a fresh chat) — vonc thread
- **status:** deployed
- **status-evidence:** HANDOFF_2026-07-09 written to the convention and used (§1 first actions ... §10 file inventory; Categories: tags); the write_site_spec handoff is the evidence-only variant.
- **what:** A dated, self-contained handoff document lets a fresh chat (memory off) start work: orientation, verified DONE state with copy-paste ids, the next task's full scope/method/acceptance, data to collect first, commands/triggers, schema notes, gotchas (each "paid for" in the thread), backlog, file inventory. The diagnostic variant carries EVIDENCE and context but deliberately NO diagnosis ("the cause is still to be read from the real code"). Related authoring hygiene rules: no bare angle-bracket tags in markdown prose (breaks readers); quote heredoc delimiters; /home/claude resets between sessions while outputs persists (re-seed working copies before appends — the fragment-clobber incident).
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-handoff
- **relations:** travelling docs; the travelling_docs thread's own HANDOFF chain (DOC-031); README_summary_paragraph
- **verify-later:** n/a (convention)

### DOC-061 — API documentation system (OpenAPI external + per-service internal API.md)
- **status:** unknown
- **status-evidence:** File dated Aug 2025 (per the directory-tree listing) and flagged "unclassified" in PROPOSED_MOVES.tsv; no corroborating recent activity found.
- **what:** A two-tier API documentation practice for the platform: customer-facing APIs documented as OpenAPI 3.0 (internal/auth-service/api/openapi.yaml + swagger annotations in *_swagger.go, `make swagger`/`make validate-openapi`, Swagger UI/Redoc/Editor via docker-compose) and internal service communication documented as per-service API.md files (auth-service, core-manager, agents/*, adapters/*) covering Kafka topics, message formats, DB schemas, env vars; CI validation workflow proposed. Predates the vonc corpus; whether the practice is followed is unverified.
- **sources:** docs/API_DOCUMENTATION.md; docs/PROPOSED_MOVES.tsv
- **relations:** admin-dashboard-and-api; documentation-system
- **verify-later:** internal/auth-service/api/openapi.yaml exists?; make targets swagger/validate-openapi; API.md coverage

### DOC-062 — Classic pre-docs024 documentation tree (emptied)
- **status:** abandoned
- **status-evidence:** All files under docs/_archive/{architecture,deployment,development,operations} are 0 bytes (dated Aug 18 2025); live docs/operations/README.md is also 0 bytes, and live docs/architecture/ is a different numbered discussion set (001-030).
- **what:** The original top-level documentation set (agent architecture, service dependencies, kustomize/terraform deployment guides, local-setup/testing dev guides, monitoring/troubleshooting ops) that existed as named slots. Every archived file in these trees is empty — the content was never migrated or was stripped, and no live markdown counterpart exists for most slots. The doc slots survive as skeletons only.
- **sources:** docs/_archive/architecture/; docs/_archive/deployment/; docs/_archive/development/; docs/_archive/operations/
- **relations:** superseded in practice by docs024_key_docs_latest and live docs/architecture numbered docs
- **verify-later:** git history of docs/_archive/architecture/agent-architecture.md to see if it ever had content

### DOC-063 — 2026-05-24 launcher build handoff (superseded Option A)
- **status:** superseded
- **status-evidence:** RUNBOOK_iter0_pretrigger(3) §0a "treat the 2026-05-24 handoff as superseded (it's the source of the prepare_object_url-added and __parent_responses_topic__ claims that turned out wrong)".
- **what:** The handoff documenting the first real training-launcher build ("Option A": a 5-step orchestrator workflow of dispatch-action clones of DispatchThunderDecommissionAction, the run.sh on-VM launch chain, migration 102). It carries two claims later proven wrong — that `prepare_object_url` had been added to the deployed adapter (it hadn't) and that replies should route to `__parent_responses_topic__`. The uploaded 102 was also a pre-revision draft that must not be re-run.
- **sources:** flywheel_docs/HANDOFF_2026-05-24_phase5_launcher_build.md; phase5/RUNBOOK_iter0_pretrigger(3).md#0a
- **relations:** superseded by the D4 own-topic decision and the handoff-correction log (DOC-008)
- **verify-later:** migration 102_training_launcher_real.sql (revised vs draft)

### DOC-064 — Deploy-from-context-packs guide — six deploy mechanisms (A–F) — dropped from the live tree
- **status:** abandoned
- **status-evidence:** A copy is used as live reference material in one archive location (flywheel_docs' docubundle, framing the thunder-checkpoint-race context pack); but an exhaustive search (`find -iname`) of the live idea.uk docs tree found no file named `GUIDE_deploy_from_context_packs.md` (or any variant) anywhere. Two byte-identical archive copies exist (docubundle_idea_golive/ and docubundle_idea_within_chassis/), yet the live tree carries none, even though it kept the sibling CONTEXT_PACK and the .sh packaging scripts from the same bundles.
- **what:** A cross-project methodology doc for taking a "context pack" (docs+code bundle for a fresh chat thread) into production, given six distinct deploy mechanisms observed across the platform: A chassis platform image (build→tag-bump→k8s rollout), B database (snapshot-first SQL via kubectl exec psql), C work-items (insert site_work_items, build-dispatch-loop claims it), D orchestration trigger (kcat → system.agent.generic.requests), E generated static sites (git→GitHub Actions→Backblaze B2, mostly automatic), F the idea.uk binary (self-contained Go binary, scp+mv-f+systemctl, not k8s, not B2). Includes a per-project quick reference (gamesdesign, Flywheel-C thunder, idea.uk go-live, imagery) and the caution "Complete" ≠ "succeeded" — verify positive evidence, not terminal status. A genuinely useful cross-cutting operational doc that appears to have been silently dropped from the live tree rather than superseded by a named replacement.
- **sources:** docubundle_idea_golive/GUIDE_deploy_from_context_packs.md (full text); docubundle/.../CONTEXT_PACK_thunder_checkpoint_race.md
- **relations:** idea.uk deployment topology (mechanism F); service-deployer pattern; travelling-docs workstream (a plausible successor concept, unconfirmed); docubundle context packagers (DOC-005)
- **verify-later:** whether this content was folded into a differently-named doc elsewhere in the live tree, or genuinely lost

### DOC-065 — HANDOFF permanent-thread scope split (Threads A–D)
- **status:** partial
- **status-evidence:** running_notes 2026-06-13(b) "everything else → HANDOFF_vm_sites_permanent_thread.md (Threads A manifest / B framework integration / C more domains / D global bot blocklist)".
- **what:** Work was split so P4 collection stayed active while the rest handed off to a permanent thread: Thread A = static-build relojistas as a manifest→framework build; Thread B = framework integration (a backend site becomes a normal multi-page chassis build); Thread C = more domains on existing boxes; Thread D = a global bot-IP blocklist sharing the access-log digest source.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b,#2026-06-13-relojistas-verdict
- **relations:** Thread D shares /access-digest source
- **verify-later:** HANDOFF_vm_sites_permanent_thread.md (live)

### DOC-066 — docs019 working/main snapshot bundle (duplicate early-draft staging copy)
- **status:** superseded
- **status-evidence:** Header-by-header comparison against each doc's live counterpart shows this whole directory is a pure prefix/subset with zero unique content: `001_development_guide(3).md` (identical structure to live `(5).md`); `016_debugging_guide` v2_21/v2_36b (exact ordered subset of live v2_58_consolidated); `026_component_regeneration_flow.md` (live version adds a confirmed-2026-06-22 section plus a second appended doc); `030_phase1_plan_and_reconciler(2)/(3).md` (byte-identical to each other and to live headers); `FOCUS_imagery_assessment.md` (identical through section 8, live continues to 13); `old/012d_tool_lifecycle_guide_v4.md` (byte-identical via md5 to the docs024 archive copy).
- **what:** A nested archive-of-archive preserving a working-copy staging snapshot of six of the platform's core numbered guides plus a duplicate tool-lifecycle-guide vintage, all captured mid-iteration before being superseded by later-numbered/consolidated versions that already live under docs024_key_docs_latest and docs014_documentation_collection. No content unique to this snapshot survives comparison against the live versions — its value is purely as a dated waypoint in each guide's version history, not as a source of new concepts.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/{001_development_guide(3).md,016_debugging_guide_v2_21_.md,016_debugging_guide_v2_36b.md,026_component_regeneration_flow.md,030_phase1_plan_and_reconciler(2).md,FOCUS_imagery_assessment.md}; .../_archive/old/012d_tool_lifecycle_guide_v4.md
- **relations:** development-guide (001 anchor); debugging (016/016b anchor); imagery; site-plan-and-reconciler; tool-lifecycle (020 anchor)
- **verify-later:** none — superseded in full by already-covered live docs
