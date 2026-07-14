# EXTRACTION U16 — docs019 design core (DESIGN/PLAN/FOCUS/HANDOFF/SQL seeds + tasks/ + 001_claude_reasoning)
Extracted 2026-07-13. Files in scope: 123. Concepts found: 104.

Paths in this file are relative to `docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/` unless prefixed otherwise.

## Coverage
| file | treatment |
|---|---|
| 001_claude_reasoning | full |
| 001_onboarding_discussion.txt | full |
| 003_contextkit_bundle_issues.md | full |
| DESIGN_diagnosis_loop.md | family-delta |
| DESIGN_diagnosis_loop(1).md | family-delta |
| DESIGN_diagnosis_loop(2).md | family-delta |
| DESIGN_diagnosis_loop(3).md | family-latest |
| DESIGN_diagnosis_loop_chassis_integration.md | family-delta |
| DESIGN_diagnosis_loop_chassis_integration(1).md | family-delta |
| DESIGN_diagnosis_loop_chassis_integration(2).md | family-delta |
| DESIGN_diagnosis_loop_chassis_integration(3).md | family-delta |
| DESIGN_diagnosis_loop_chassis_integration(4).md | family-delta |
| DESIGN_diagnosis_loop_chassis_integration(5).md | family-delta |
| DESIGN_diagnosis_loop_chassis_integration(6).md | family-latest |
| DESIGN_doc_drift_classifier.md | full |
| FOCUS_contract_set_review.md | full |
| FOCUS_js_tools_documentation.md | full |
| FOCUS_onboarding_system_view_check.md | full |
| FOCUS_pre_build_edge_cases(1).md | full |
| FOCUS_schema_verification_findings.md | full |
| FOCUS_whole_plan_review.md | full |
| GUIDE_deploy_from_context_packs(1).md | full |
| HANDOFF_builder_thread.md | full |
| HANDOFF_fixloop_thread.md | family-delta |
| HANDOFF_fixloop_thread(1).md | family-delta |
| HANDOFF_fixloop_thread(2).md | family-delta |
| HANDOFF_fixloop_thread(3).md | family-delta |
| HANDOFF_fixloop_thread(4).md | family-delta |
| HANDOFF_fixloop_thread(5).md | family-delta |
| HANDOFF_fixloop_thread(6).md | family-delta |
| HANDOFF_fixloop_thread(7).md | family-delta |
| HANDOFF_fixloop_thread(8).md | family-latest |
| MAPPING_tool_to_actions_and_agents(2).md | full |
| NNN_add_component_provenance.sql | full (header + DDL) |
| NNN_create_code_indexer_agent(1).sql | family-delta |
| NNN_create_code_indexer_agent(2).sql | family-latest |
| NNN_create_code_symbols_index.sql | full (header) |
| NNN_create_diagnose_ro_role.sql | full (header) |
| NNN_fix_assemble_bundle_loop_scope_field.sql | full (header) |
| NNN_fix_diagnose_agent_workflow.sql | family-delta |
| NNN_fix_diagnose_agent_workflow(1).sql | family-delta |
| NNN_fix_diagnose_agent_workflow(2).sql | family-latest |
| NNN_fix_diagnose_complete_output_fields.sql | full (header) |
| NNN_fix_diagnose_route_state_threading.sql | family-delta (identical to (1)) |
| NNN_fix_diagnose_route_state_threading(1).sql | family-latest (header) |
| NNN_fix_lookup_repo_label_workaround.sql | full (header) |
| NNN_fix_orchestrator_complete_key.sql | full (header) |
| NNN_fix_researcher_spawn_columns.sql | full (header) |
| NNN_move_diagnose_workflow_to_default_config.sql | family-delta |
| NNN_move_diagnose_workflow_to_default_config(1).sql | family-latest (header) |
| NNN_rename_complete_keys_preferred.sql | full (header) |
| NNN_reroute_classifier_to_vertical_research.sql | full (header) |
| NNN_restore_diagnose_orchestrator_workflow.sql | family-delta (identical to (1)) |
| NNN_restore_diagnose_orchestrator_workflow(1).sql | family-latest (header + body) |
| NNN_seed_diagnose_agents(1).sql | family-delta |
| NNN_seed_diagnose_agents(2).sql | family-latest (full incl. prompt body) |
| NNN_seed_index_orchestrator.sql | family-delta |
| NNN_seed_index_orchestrator(1).sql | family-latest (header) |
| NNN_seed_vertical_exemplar_researcher.sql | family-delta |
| NNN_seed_vertical_exemplar_researcher(1).sql | family-delta |
| NNN_seed_vertical_exemplar_researcher(2).sql | family-latest (header) |
| NNN_strategist_vertical_landscape_nudge.sql | full |
| NNN_swap_analyse_repo_to_local.sql | full (header) |
| NNN_swap_indexer_to_local_analysis.sql | full (header) |
| NNN_update_tool_prompts_doc_header.sql | full (header) |
| PATCH_code_symbols_shared_repo_label.md | full |
| PATCH_lift_fetcher_and_register.md | full |
| PLAN.md | full |
| PLAN_active_config_schema(3).md | full |
| PLAN_bundle_shape_contract(2).md | full |
| PLAN_capabilities_catalog_contract(1).md | full |
| PLAN_change_layer_integration_contract(4).md | full |
| PLAN_config_work_items_contract(3).md | full |
| PLAN_context_assembly_tool_and_service(2).md | full |
| PLAN_decision_log_contract(2).md | full |
| PLAN_imagery_sprite_sheet.md | full |
| PLAN_onboarding_agent_specs(6).md | full |
| PLAN_onboarding_config_derivation.md | full |
| PLAN_trust_ledger_contract(4).md | full |
| PLAN_workflows_and_actions_migration(15).md | family-delta (subset of (19)) |
| PLAN_workflows_and_actions_migration(16).md | family-delta (subset of (19)) |
| PLAN_workflows_and_actions_migration(17).md | family-delta (subset of (19)) |
| PLAN_workflows_and_actions_migration(19).md | family-latest (full) |
| PREAMBLE_gamesdesign_diagnosis_handoff.md | full |
| PROMPT_diagnosis_verdict.md | family-delta |
| PROMPT_diagnosis_verdict(1).md | family-latest (full) |
| README_02_evidence_backed_proposals.md | full |
| README_claude_conversation.md | full (URL only) |
| README_comprehensive_documentation_categorisation.md | full |
| README_flows.md | full |
| README_iterate_until_bugfix_notes.md | full |
| README_overview.md | full |
| README_useful_grep_for_documentation.sh | header-scan |
| README_whats_next.md | full |
| README_worker_statuses.md | full |
| TRIGGER_code_indexer_v2.sh | family-delta |
| TRIGGER_code_indexer_v2(1).sh | family-latest (header-scan) |
| directory_tree.txt | skipped-generated |
| engines_tree_proposal.md | full |
| tasks/005site_scheme_palette_and_components/HANDOFF_scheme_to_components.md | full |
| tasks/005site_scheme_palette_and_components/TODO_chassis_and_idea_uk(1).md | full |
| tasks/005site_scheme_palette_and_components/one_sentence_description.md | full |
| tasks/005site_scheme_palette_and_components/running_notes_2(5).md | full |
| tasks/any_project_handoff/001_build_bundle_ask_for_handoff | full |
| tasks/gameslink_missing_index_rerender/001_one_sentence_description.txt | full |
| tasks/gameslink_missing_index_rerender/RUNBOOK_gamesdesign_silent_norebuild_bug.md | family-delta |
| tasks/gameslink_missing_index_rerender/RUNBOOK_gamesdesign_silent_norebuild_bug(1).md | family-delta |
| tasks/gameslink_missing_index_rerender/RUNBOOK_gamesdesign_silent_norebuild_bug(2).md | family-latest (full) |
| tasks/gameslink_missing_index_rerender/bundle_gamesdesign.md | skipped-generated (header noted) |
| tasks/gameslink_missing_index_rerender/bundle_gamesdesign_generation.md | skipped-generated (header noted) |
| tasks/gameslink_missing_index_rerender/content_writer_prompt.md | skipped-generated (DB dump; header noted) |
| tasks/missing_game_on_games_page/001_bundle | full |
| tasks/this_project_itself/002_more_sql | header-scan |
| tasks/this_project_itself/003_more_sql_collector_scripts | header-scan |
| tasks/this_project_itself/004_more_sql_simple_collector_script | header-scan |
| tasks/this_project_itself/005_raw_sql_tmp.sql | skipped-generated (empty file, 0 bytes) |
| tasks/this_project_itself/2_raw_sql_temp.sql | skipped-generated (psql session dump; identical to raw_sql_temp.sql) |
| tasks/this_project_itself/raw_sql_temp.sql | skipped-generated (psql session dump) |
| tasks/this_project_itself/bundle_diagnosis_loop.sh | header-scan |
| tasks/vonc_provocations_lobby/bundle_minilobby_trim.sh | family-delta |
| tasks/vonc_provocations_lobby/bundle_minilobby_trim1.sh | family-delta |
| tasks/vonc_provocations_lobby/bundle_minilobby_trim(2).sh | family-delta |
| tasks/vonc_provocations_lobby/bundle_minilobby_trim(3).sh | family-latest (header-scan) |

## Concepts

### Iterative-bundle diagnosis loop (the automated debugging motion)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN_diagnosis_loop(3) "Status: … the engine is now BUILT. As of 2026-06-24, `pkg/diagnose` … exists and is tested"; PLAN.md DONE list; README_flows describes live runs.
- **what:** Automates the five-move debugging loop performed by hand in the gamesdesign session: hypothesise from a symptom, gather read-only evidence (bundle), test the hypothesis against the evidence (verdict), re-scope from what the evidence revealed, iterate until pinned or capped. Output is always a diagnosis plus full evidence trail, never a fix. Moves 1/2/4/5 are mechanical; move 3 (falsification) is the crux.
- **sources:** DESIGN_diagnosis_loop(3).md#0-1; README_iterate_until_bugfix_notes.md; README_overview.md; PLAN.md
- **relations:** cite-or-abstain verdict contract; convergence guards; chassis diagnose_route realisation; diagnosis→fix loop
- **verify-later:** pkg/diagnose/{loop,step,advance,callgraph,verdict_wire}.go; contextkit/internal/diagnose; agent_definitions rows diagnose-agent/diagnose-orchestrator

### Cite-or-abstain verdict contract + diagnosis verdict prompt
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Prompt is inline (JSON-escaped) in the applied NNN_fix_diagnose_agent_workflow(2).sql verdict step; verdict_wire.go parses its output (tested).
- **what:** Per iteration the model must return exactly one of CONFIRMED / REFUTED / UNVERIFIABLE with verbatim citations; a citation-less confirm/refute is coerced to UNVERIFIABLE; CONFIRMED only on direct evidence ("consistent with" = UNVERIFIABLE — the abstention asymmetry); no fix may be proposed; each citation tier-tagged with freshness. The prompt carries a worked REFUTED example (the gamesdesign reversal) and a self-suspicion caution. Schema must stay in lockstep with verdict_wire.go.
- **sources:** PROMPT_diagnosis_verdict(1).md; DESIGN_diagnosis_loop(3).md#2; NNN_fix_diagnose_agent_workflow(2).sql
- **relations:** doc-drift classifier evidence-or-abstain (its origin); falsification-first principle
- **verify-later:** pkg/diagnose/verdict_wire.go tests; live diagnose-agent default_config verdict step

### Falsification-first / confident wrongness as the single enemy
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** README_02 "The single enemy is confident wrongness. Runs 1–2 of the benchmark produced CONFIRMED verdicts that were wrong… Everything since … is aimed at that one failure mode" (2026-07-09 context).
- **what:** The design premise of the whole project: LLMs rationalise their first hypothesis, so every mechanism (citation mandate, REFUTED-is-correct framing, guards, council, closure gate) exists to force explicit falsification and make abandoning a wrong hypothesis cheap. The most valuable move in the founding debug was the model twice saying "my hypothesis is wrong".
- **sources:** DESIGN_diagnosis_loop(3).md#0; README_iterate_until_bugfix_notes.md; README_02_evidence_backed_proposals.md; README_overview.md
- **relations:** cite-or-abstain contract; real-bug eval gate; council pattern
- **verify-later:** eval-run artefacts; benchmark run records (runs 1–2 wrong CONFIRMED)

### B4a retrieval ceiling (symptom cannot reach infrastructure-layer causes)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN_diagnosis_loop(3) §1a "B4a (2026-06-17) measured… ALL of lexical, semantic, and fused scored 0.00" on the real gamesdesign fix symbols.
- **what:** Empirical measurement that one-shot retrieval from a symptom description cannot reach a cause living in shared infrastructure named for its function not its failure mode (resolveResultSpec/extractWorkflowResult): the symptom's words and the mechanism's words do not intersect. Lexical beat semantic on the mechanism-named task (0.50 vs 0.00). Consequence: embeddings did not earn a code-path place; the lever is iterative re-scoping following runtime evidence, not better retrieval. Retrieval seeds only the first scope.
- **sources:** DESIGN_diagnosis_loop(3).md#1a; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17 changelog); README_overview.md
- **relations:** evidence-follows re-scoping; text-vs-code embedding split (B4b); code_symbols index
- **verify-later:** contextkit eval_targets + groundtruth_targets.json; go_files/contextkit/{lex,sem}.json

### Evidence-follows re-scoping (call graph + runtime-named next scope)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Engine callgraph.go exists/tested per DESIGN(3) status; prompt rule 4 "Follow the evidence to the next scope — do not re-search the symptom".
- **what:** On REFUTED/UNVERIFIABLE the next bundle scopes the symbols/files the evidence names plus their call-graph neighbourhood (the analyser records `calls`), and prefers a runtime-named fault site over a retrieval-proposed one. This is the move retrieval cannot do; it reached the coordinator's result extraction in the real case — a symbol the symptom could never name.
- **sources:** DESIGN_diagnosis_loop(3).md#1a; PROMPT_diagnosis_verdict(1).md rule 4; NNN_fix_assemble_bundle_loop_scope_field.sql
- **relations:** B4a ceiling; convergence guards; Go analyser call graph
- **verify-later:** pkg/diagnose/callgraph.go; diagnose_route re-scope path

### Convergence guards (cap, narrow, grow, no-thrash)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN(3) status: "loop control, convergence guards … exists and is tested"; migration(19): "Four convergence guards + the no-citation→UNVERIFIABLE coercion all behaviour-tested".
- **what:** Deterministic Go guards so the loop cannot run forever or wander: iteration cap (5); scope must narrow (widening = not converging); evidence must grow (two iterations without new grounded evidence → stop with best-effort); no hypothesis thrash (oscillation without discriminating evidence → report both). Deliberately kept in tested Go, never re-expressed as workflow conditionals.
- **sources:** DESIGN_diagnosis_loop(3).md#3; PLAN_workflows_and_actions_migration(19).md; NNN_fix_diagnose_route_state_threading(1).sql
- **relations:** state-threading fix (guards were silently inert live); thin-workflows rule
- **verify-later:** pkg/diagnose/step.go DecideStep + tests

### Read-only, human-gated boundary of the diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN chassis integration(6) §7 "boundaries … do not relax in the chassis"; HANDOFF_fixloop(8) "Loop core is READ-ONLY by contract; sqlguard allowlists reads — keep it so".
- **what:** The loop gathers read-only (analyser, code_symbols, `\d`/capped SELECT/existing-log reads), proposes a diagnosis + suggested fix surface, and never applies fixes or triggers runs to test hypotheses. The human is kept at the two points that mattered: deciding the fix and backstopping the model's willingness to abandon a hypothesis. The F1 write surface is deliberately a separate agent with isolated credentials.
- **sources:** DESIGN_diagnosis_loop(3).md#4; DESIGN_diagnosis_loop_chassis_integration(6).md#7; HANDOFF_fixloop_thread(8).md#3
- **relations:** fix-implementer (the separate write surface); doc-drift read-only rule; three-guard read-only SQL
- **verify-later:** pkg/diagnose/sqlguard.go; spawn token-gate in spawn_actions.go

### Evidence tiers with freshness tagging (static / state / runtime)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Prompt rule 6 (tier + `fresh` per citation) in the applied workflow migration.
- **what:** Every citation is tagged static (code/schema), state (a DB row at a point in time) or runtime (log/work-item from an actual run), with observation time for state/runtime, so a verdict resting on stale evidence is visibly weak. Adapted from the doc-drift classifier's T1/T2/T3.
- **sources:** PROMPT_diagnosis_verdict(1).md rule 6; DESIGN_diagnosis_loop(3).md#2; DESIGN_doc_drift_classifier.md#2
- **relations:** doc-drift evidence tiers; misattribution asymmetry
- **verify-later:** verdict_wire.go citation struct

### Chassis realisation: diagnose_route workflow-driven loop (four diagnose actions)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Chassis-integration(6) banner (2026-06-24) "the BUILT design is `diagnose_route` … The actions that actually exist are diagnose_load_runtime, diagnose_assemble_bundle, diagnose_route, diagnose_emit"; live runs 8d488e01/73ed55c6 in fix migrations.
- **what:** On the chassis the loop is an agent workflow: analyse_repo → lookup_symbols (seeds iteration-1 scope from the symptom) → load_runtime → assemble_bundle → verdict (its own execute_llm_prompt step, observable) → diagnose_route (engine Advance: guards + call-graph re-scope, then a next_step override, the conditional_route pattern) → back to load_runtime or emit. The loop returns to load_runtime, not assemble, so the prior verdict's data_requests run and runtime re-gathers each iteration. Gather reuses existing actions (request_repo_analysis→analyse_repo_local, lookup_code_symbols, execute_llm_prompt) per the STEP-ZERO reuse audit; only the bundle composer was genuinely new.
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md#0,#status; NNN_fix_diagnose_agent_workflow(2).sql; PLAN.md; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17 entry)
- **relations:** abandoned diagnose_run and diagnostician designs; one-decision-core two realisations
- **verify-later:** platform/orchestration/actions/diagnose_*_action.go; coordinator.go getNextStepFromResult; registry.go Category "diagnose"

### Abandoned design: diagnose_run single engine-wrapping action
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** Chassis-integration(6) banner: "the §4–§6 `diagnose_run` recommendation below is the ABANDONED path … there is no `diagnose_run` action"; the seeded workflow referencing it was rewritten by NNN_fix_diagnose_agent_workflow.
- **what:** The originally recommended shape — one `diagnose_run` action calling `diagnose.Run()` with an injected Verdicter, keeping the whole loop inside a single step. Dropped in favour of the workflow-driven observable loop (verdict as its own step, router action). A prompt-registry reference `diagnose-verdict-v1` belonged to this design and is also unused; the prompt went inline instead. Kept here because seeded rows briefly referenced the non-existent action (a real incident class: workflow names an action that does not exist).
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md banner,#4-6; NNN_fix_diagnose_agent_workflow(2).sql header; NNN_move_diagnose_workflow_to_default_config(1).sql banner
- **relations:** superseded by diagnose_route realisation
- **verify-later:** absence of diagnose_run in registry.go

### Abandoned design: diagnostician per-iteration re-invocation (spawn-next chain)
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** NNN_seed_diagnose_agents(2).sql banner "SUPERSEDED — DO NOT APPLY … kept only as a record of the re-invocation design that was considered and dropped."
- **what:** A third loop shape: each orchestration runs ONE iteration (load_runtime → analyse → lookup → assemble → verdict → route → conditional), and on continue spawns+calls a fresh `diagnostician` of the same type with revised hypothesis/scope/iteration in input_data, the terminal verdict bubbling up the child chain. Motivated by doubt that the engine supported a workflow-internal cycle and by the build-dispatch-loop one-unit-per-orchestration precedent. Dropped once the next_step-override loop-back was confirmed to work.
- **sources:** NNN_seed_diagnose_agents(2).sql header + workflow body
- **relations:** superseded by diagnose_route loop-back; build-dispatch-loop pattern
- **verify-later:** no `diagnostician` row in agent_definitions

### One decision core, two realisations (Run vs Advance/DecideStep)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN(3) §1 note: "Both share ONE decision core — step.go's pure DecideStep … advance_test.go proves Advance threaded across iterations … reproduces Run()".
- **what:** The standalone dev harness (`internal/diagnose/loop.go` Run, a Go for-loop with inline IO) and the chassis workflow loop share the same pure per-iteration decision function, with `advance.go`'s Advance exposing it statefully to the chassis; equality is proven by test. cmd/diagnose stays the dev/test harness (scripted verdicts, dry-bundle), never a production entrypoint.
- **sources:** DESIGN_diagnosis_loop(3).md#1; DESIGN_diagnosis_loop_chassis_integration(6).md#status,#3; PLAN_workflows_and_actions_migration(19).md
- **relations:** engine/harness file-placement split; travelling contextkit module
- **verify-later:** pkg/diagnose/advance.go + advance_test.go

### Model-written data_requests under three-guard read-only SQL
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** PLAN.md DONE: "Three-guard model-SQL feature: Guard 1 (prompt), Guard 2 (IsReadOnlySQL), Guard 3 (read-only transaction, confirmed on the chassis side)"; GATED item 1: execution wiring in diagnose_load_runtime is deploy-side and untested; README_flows: "terminal-verdict data_requests that never run".
- **what:** The verdict may request specific evidence as single read-only SELECTs (`data_requests: [{sql, why}]`), defended by three independent guards: the prompt contract, a Go lint (sqlguard.IsReadOnlySQL), and execution inside a read-only transaction with statement timeout; the harness analogue is a GRANT-based SELECT-only `diagnose_ro` role (not default_transaction_read_only, unreliable under pgbouncer). Notably this reversed an earlier stance — chassis-integration(6) recorded "the model never writes SQL" and called runDataRequests dormant/beyond the boundary; the bounded, guarded version was then built deliberately.
- **sources:** PROMPT_diagnosis_verdict(1).md rule 7; NNN_create_diagnose_ro_role.sql; PLAN.md; DESIGN_diagnosis_loop_chassis_integration(6).md#status (the earlier stance)
- **relations:** read-only boundary; self-verification in the council (same move at review time)
- **verify-later:** pkg/diagnose/sqlguard.go in chassis; diagnose_load_runtime data-request execution path; diagnose_ro role existence

### Real-bug evaluation gate (scaffold correct ≠ reasons well)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** PLAN.md GATED 5: "THE EVAL GATE … MUST reproduce the mid-course REVERSALS and ABSTAIN when unsettled … No automatic triggering until this passes"; README_02 shows benchmark runs happened (runs 1–2 confidently wrong) and design responded.
- **what:** Before any unsupervised or automatic triggering, the live loop must be run against known bugs — the gamesdesign two-fault bug (with its captured reversals) and the 016 §9 silent-no-op catalogue — and must reproduce hypothesis reversals and abstain when evidence does not settle, rather than confirming first guesses. "Compiling isn't behaving" is the standing lesson; a loop that confirms its first guess every time is the failure mode dressed as success.
- **sources:** DESIGN_diagnosis_loop(3).md#6; PLAN.md GATED; README_whats_next.md; README_02_evidence_backed_proposals.md
- **relations:** gamesdesign bug fixture; falsification-first; later trigger modes gated on this
- **verify-later:** eval run records; whether triggers (b)/(c) were ever enabled

### Diagnose agent pair + generic-request trigger envelope
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Restore/fix migrations applied against live rows (snapshot ids 34f4afc8/e8e96d24 in HANDOFF_builder_thread); TRIGGER pattern proven (084_TRIGGER_diagnose_v1.sh referenced as canonical in HANDOFF_fixloop(8)).
- **what:** A thin diagnose-orchestrator (spawn_agent → call_agent → complete, per "every agent is an orchestrator" and the wrapper rule for substantive work) spawns the diagnose-agent worker pod that runs the loop. Triggering is the existing generic-request envelope — kcat to system.agent.generic.requests with agent_type diagnose-orchestrator and input_data {symptom, seed_scope, runtime_site, …} — no new triggering code; later triggers (on build failure, proactive sweep) are the same envelope from a different sender, gated on the eval. Sub-agents reply on the caller's responses topic.
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md#2; NNN_restore_diagnose_orchestrator_workflow(1).sql; PLAN_workflows_and_actions_migration(19).md
- **relations:** wrapper-orchestrator pattern; index-orchestrator (same pattern reused)
- **verify-later:** agent_definitions diagnose-orchestrator/diagnose-agent; drafts/084_TRIGGER_diagnose_v1.sh

### Workflow lives in default_config, not the workflow columns
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NNN_seed_diagnose_agents(2) header: "confirmed: those agents put their workflow there; the task_workflow / orchestrator_workflow columns are NULL"; the move/fix/restore migration sequence applied.
- **what:** The loader reads `agent_definitions.default_config` ({workflow:{start_step,steps}, processing_mode, timeout_seconds}); the three workflow columns are dead for working agents. A workflow seeded into orchestration_workflow silently never loads — this bit the diagnose pair (seeded 2026-06-20 into the wrong column, then the orchestrator's workflow was lost entirely during the move and had to be re-seeded). Key correction learned by reading a real row rather than docs.
- **sources:** NNN_move_diagnose_workflow_to_default_config(1).sql; NNN_restore_diagnose_orchestrator_workflow(1).sql; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17)
- **relations:** schema-before-SQL discipline; spawn-consumed columns lesson (sibling class)
- **verify-later:** workflow loader code path; whether the three columns are still consulted anywhere

### Diagnose loop-back plumbing fault class (state threading, scope encoding)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Both fix migrations are operative live fixes with run evidence (run 8d488e01: "there is a 'route' key, no 'diagnose_state' key"; trail truncated to 1 entry).
- **what:** Two producer/consumer field mismatches that left the loop running but silently degraded: (1) diagnose_route read its LoopState from bare `diagnose_state` while its own output lands under `route.diagnose_state` — so the loop re-seeded every iteration, never enforcing max_iterations, truncating the evidence trail, and resetting the cross-iteration guards; (2) route.scope is EncodeScope's untagged-struct JSON, so the string-list reader needed `route.scope.Symbols` (capital S) — before the fix every re-scope silently fell through to the fallback chain and iterations 2+ never moved scope. Both were invisible-success faults: the loop "worked" while its defining features were inert.
- **sources:** NNN_fix_diagnose_route_state_threading(1).sql; NNN_fix_assemble_bundle_loop_scope_field.sql; DESIGN_diagnosis_loop_chassis_integration(6).md#status (the round-trip flagged as unverified)
- **relations:** workflow-variable-sync rule; result-contract dead-key class; convergence guards
- **verify-later:** diagnose_route_action.go default state_field; Scope struct json tags

### Result-contract dead-key class and Option A unification
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** NNN_fix_diagnose_complete_output_fields (run 17933a83, Kafka Message Size Too Large at 1.27MB); HANDOFF_builder_thread "Option A CLOSED: shared result contract (datahelpers/result_contract.go), response size guard, four agents on preferred result_from; deployed v1.0.1092+".
- **what:** CompleteWorkflowAction honoured output_field/output_fields and otherwise shipped the ENTIRE collected_data; `result_from` was a key the action never read, so diagnose completions always shipped everything (masked until the 515-file analysis blew the Kafka cap). Same class: the orchestrator pointed output_fields at an imagined key ("diagnose-agent_result") when the engine stores a call step's response under the STEP NAME. Fixes: point at real keys; then Option A — a shared result contract with both readers honouring result_from/output_fields plus a response size guard; NNN_rename_complete_keys_preferred moves the four diagnose/index agents to the preferred key once that image is deployed. Standing rule made mechanical: keep workflow variable names in sync with what actions read.
- **sources:** NNN_fix_diagnose_complete_output_fields.sql; NNN_fix_orchestrator_complete_key.sql; NNN_rename_complete_keys_preferred.sql; HANDOFF_builder_thread.md#2
- **relations:** loop-back plumbing fault class; bounded bundle egress (persist-and-reference)
- **verify-later:** datahelpers/result_contract.go; extractFinalResult size guard; census query for deprecated keys

### code_symbols repo-label symmetry (shared owner/repo resolver)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PATCH_code_symbols_shared_repo_label written; NNN_fix_lookup_repo_label_workaround header: "TEST-NOW ENABLER … apply that and REVERT this when the rebuilt image ships".
- **what:** index_code_symbols composes the code_symbols.repo label as owner/repo ("gqls/agentchassis") but lookup_code_symbols did not compose, so the diagnose workflow's lookup queried the bare name → 0 hits → empty code_results → "assemble_bundle: no scope". Structural fix: one shared resolveCodeRepoLabel used by both writer and reader so they can never drift; temporary config-only workaround hard-codes the literal until the image ships. General lesson: writer and reader of a keyed store must share the key resolver.
- **sources:** PATCH_code_symbols_shared_repo_label.md; NNN_fix_lookup_repo_label_workaround.sql; NNN_create_code_indexer_agent(2).sql (label convention)
- **relations:** code_symbols index; loop-back plumbing fault class
- **verify-later:** code_symbols_actions.go resolveCodeRepoLabel; whether the workaround REVERT ran

### analyse_repo_local in-process analysis + the stale-index incident
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** NNN_swap_indexer_to_local_analysis header documents the incident and the applied swap; PATCH_lift_fetcher_and_register gives the registration; README_flows/handoffs treat the local path as current.
- **what:** The adapter round-trip left the diagnose-agent with no local checkout (ReadSymbolBody could not slice bodies) and once resolved ref "HEAD" to a year-old commit, silently indexing a July-2025 tree (69 files/436 symbols vs 572 today). analyse_repo_local fetches the repo tarball at an explicit ref to a local temp dir and analyses in-process (shared internal/reposource fetcher), with pin_to_index_commit defaulting true for the diagnose loop (bodies match the index) and set false for the indexer (the indexer defines the commit). Corollary rules: explicit git refs never HEAD; the spawned pod needs GITHUB_READ_TOKEN via the isRepoCloningAgent spawn gate.
- **sources:** NNN_swap_analyse_repo_to_local.sql; NNN_swap_indexer_to_local_analysis.sql; PATCH_lift_fetcher_and_register.md; TRIGGER_code_indexer_v2(1).sh
- **relations:** analyser adapter (request_repo_analysis stays for the code-indexer); index freshness / CI-triggered indexing
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; registry entry

### Diagnosis persistence + documented intake (diagnosis_artifacts, needs_diagnosis)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_fixloop(8): "First action: slice F0.1 with pre-registered criteria — (1) diagnosis_artifacts migration … (2) assemble-action write-through … (3) the needs_diagnosis envelope"; decisions recorded 2026-07-07, not yet built in these files.
- **what:** F0 of the fix loop: make each iteration's bundle durably fetchable and add per-iteration running notes — a `diagnosis_artifacts` table (correlation_id, iteration, kind ∈ {bundle, iteration_note}, body, retention knob per kind) written through from the assemble action Go-side (no workflow-shape change); plus a documented intake route: a `needs_diagnosis` envelope / pipeline='diagnose' work item carrying subject_type/subject_key with null-site allowed. Bundle egress via completion payloads is bounded (max_response_bytes) — persist and reference, don't ship megabytes.
- **sources:** HANDOFF_fixloop_thread(8).md#4; HANDOFF_fixloop_thread(8).md#3
- **relations:** result-contract size guard; travelling-docs pattern (notes per iteration); work-item relay
- **verify-later:** diagnosis_artifacts table existence; persist_note step in diagnose-agent workflow (tools thread's wiring)

### Verdict-quality wrinkles + dead code-retrieval channel (measured)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** README_flows: "seed similarities in the 0.55 band, no page-build symbols … a stale line carried into a conclusion that its own citation contradicts, and terminal-verdict data_requests that never run".
- **what:** Post-run findings on the live loop: the lookup channel contributes nothing measurable (work is on the query side — seed the lookup from runtime evidence or expand the query, a self-contained lookup_symbols change); the trigger's site_id is intermittent across runs (reproducibility); and two verdict-quality defects point at the confirm/emit step (a conclusion contradicted by its own citation; data_requests emitted on terminal verdicts that never execute).
- **sources:** README_flows.md; PLAN.md GATED
- **relations:** B4a ceiling; data_requests wiring; eval gate
- **verify-later:** lookup_symbols seeding config; emit/confirm step handling of terminal data_requests

### Diagnosis→fix loop programme (F0–F3)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** HANDOFF_fixloop(8) (2026-07-07): "All F0/F1 design questions are DECIDED"; README_02: "a pipeline that turns 'something is wrong' into 'here is a reviewable, evidence-backed proposal' … exercised on exactly one bug. The write step (plan → PR) is half-built."
- **what:** The evolution of the diagnosis loop into a fix pipeline: symptom → cited diagnosis → constrained edit plan → adversarial council review → revision informed by reviewer-requested DB queries → approved plan or honest escalation. Phased F0 (persistence/intake) → F1 (write step) → F2 (council expansion) → F3, driven by open questions Q-A…Q-H resolved in the discussion thread. The valuable output is the general pattern, not the bug-fixing.
- **sources:** HANDOFF_fixloop_thread(8).md; README_02_evidence_backed_proposals.md; README_overview.md (F1.1b(c) status)
- **relations:** council pattern; fix-implementer; pilot worthiness test; docs026 concept-council mission
- **verify-later:** RUNBOOK_diagnosis_fix_loop.md + NOTES_running_fixloop.md (units U14/U15); fix-implementer seed

### Council pattern: adversarial multi-agent review with deterministic aggregation
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** README_02: "the roster of 2 (edit-quality, guardian) is explicitly a skeleton"; "the guardian vetoed an architecture change dressed as a fix"; veto semantics decided per HANDOFF_fixloop(8) delta history ("flag-based hard veto, default advisory → decision-maker").
- **what:** Multiple reviewer agents each examine the proposed fix plan from one lens; a deterministic rule (not a third model) aggregates their positions; specified veto semantics are flag-based hard veto with advisory as default. Reviewers can demand facts, and the loop runs the queries itself rather than letting the proposer argue (self-verification instead of self-belief). Three runs running, the council correctly ruled the test bug's proper fix beyond a constrained plan's mandate.
- **sources:** README_02_evidence_backed_proposals.md; HANDOFF_fixloop_thread(4)-(8).md deltas; README_comprehensive_documentation_categorisation.md (veto description)
- **relations:** expanded council bench; escalation as success; guardian-from-decision-record (Q-G)
- **verify-later:** council agent seeds + the aggregation rule in Go; fixloop_eg_dartsonline docs (docs024)

### Expanded council bench (expert-per-area reviewers)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02: "the runbook's F2 design already names the full bench … Adding a reviewer is a seed change + prompt + curated context."
- **what:** The planned F2 roster beyond the two-agent skeleton: a guidelines agent (conformance to 000–0xx, or did the guidelines fall short), reuse agent (are we rebuilding something that exists, code and docs), bug-historian (has this class recurred), compliance eye, pipeline guardians one per master workflow, and specialist knowledge agents. Reviewer areas are expected to correlate with the docs024 documentation categories — the direct bridge to the docs026 concept register's council-agent goal.
- **sources:** README_02_evidence_backed_proposals.md#3; README_comprehensive_documentation_categorisation.md
- **relations:** council pattern; concept register mission; documentation categories as expertise areas
- **verify-later:** whether any bench agents beyond edit-quality/guardian were seeded

### Fix-implementer constrained write step
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** README_overview: "F1.1b(c) is code complete (b367a602) … validated as far as it can be without the deploys"; deploy checklist listed, first end-to-end run pending.
- **what:** A 15-step seeded agent: load plan/council/diagnosis → APPROVED gate (mirror of the CONFIRMED gate) → diagnose_read_repo_files fetches current file bodies via the GitHub contents API with a hard rule that a modify-file 404 is a refusal (whole-file rewrites of unseen files would be hallucination by construction) → sketch_to_files whole-file rewrites ("the diff a human reviews must contain ONLY the plan") → deterministic file allowlist → create fix/* branch → commit via git_adapter_request (one generic adapter caller; verbs allowlisted to commit/create_branch/create_pull_request so delete_repo is structurally unreachable) → build gate (golang Job): green → PR into main, red → NO PR, branch + build log left. Runs on the read-token spawn gate.
- **sources:** README_overview.md (landed pieces + deploy checklist); README_02_evidence_backed_proposals.md
- **relations:** hard deterministic gates; human-gate-never-moves; seeded-bug first run; build gate options A/B/C
- **verify-later:** 0NN_fix_implementer.sql; 092_TRIGGER_fix_implementer_v1.sh; git-adapter branch/PR ops; RBAC for pods/log

### Hard deterministic gates between every LLM step
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02 lists them as built pattern: "CONFIRMED gate, plan validator, file allowlist. The models propose; plain Go code decides what proceeds."
- **what:** No LLM output passes into consequence unchecked: the diagnosis must be CONFIRMED (gate), the plan must validate, the files must be on a deterministic allowlist, the build must pass, before anything advances. Complexity and authority live in plain Go; the models only propose. The same shape as keeping convergence guards in the engine rather than in workflow conditionals.
- **sources:** README_02_evidence_backed_proposals.md#1; README_overview.md
- **relations:** council aggregation rule; fix-implementer; thin-workflows rule
- **verify-later:** the gate implementations in the fixloop actions

### The human gate never moves (nothing merges itself)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02: "one structural commitment: the human gate never moves. More autonomy upstream … never past the PR. Nothing merges itself."
- **what:** Autonomy may widen upstream (diagnose, plan, revise, commit-to-branch) but the merge is permanently human. The PR is the fixed boundary of machine authority in the fix loop — a simpler, harder commitment than the graduated trust machinery, and orthogonal to it.
- **sources:** README_02_evidence_backed_proposals.md#2; README_overview.md (red build → NO PR)
- **relations:** trust ledger (graduated autonomy elsewhere); awareness surface; fork isolation
- **verify-later:** absence of any auto-merge path in the write step

### Escalation as a first-class success
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02 pattern list: "'this is beyond my mandate' is a correct output, packaged for you"; the council produced exactly this on the test bug three times.
- **what:** When a fix exceeds the constrained plan's mandate (architecture-level causes), the loop's correct output is an honest escalation package for the human, not a forced plan. Treating refusal-to-proceed as success is the organisational analogue of UNVERIFIABLE-beats-guessing.
- **sources:** README_02_evidence_backed_proposals.md; README_02 §6 (the escalate decision explained)
- **relations:** cite-or-abstain; council pattern
- **verify-later:** escalation package format in the fixloop runbook

### Fix-loop value proposition: unattended, cited, consistent
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02: "The value proposition (decided 2026-07-09): not 'the loop finds what humans can't' … The proposition is unattended, cited, consistent — the 3am diagnosis with a paper trail."
- **what:** A recorded decision reframing what the loop is for: on this platform bugs are legible to anyone with schema access and patience, so the differentiation is not superhuman insight but unattended operation with citations and consistency — a package instead of a hunch, reconstructible after the fact by one correlation id. Every design choice flows from it.
- **sources:** README_02_evidence_backed_proposals.md#2
- **relations:** falsification-first; awareness surface; diagnosis artifacts persistence
- **verify-later:** decision record in NOTES_running_fixloop (U15)

### Awareness surface before wider autonomy
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02: "the missing organ is a push surface: a periodic digest … before autonomy widens … the awareness surface gets built first" — a recommendation, not built.
- **what:** The named risk is not wrong action but unknown action — drift compounding silently while trails exist only pull-side. Proposed standing gate: before councils multiply or migration agents exist, build a push digest (what ran, what was decided and by which rule, what was escalated, what the council almost approved). "It must explain what it's doing, or it doesn't get to do more." The grown-up form of the parked F0.3 per-iteration notes.
- **sources:** README_02_evidence_backed_proposals.md#4
- **relations:** diagnosis artifacts persistence; decision log (the governance twin); human-gate-never-moves
- **verify-later:** whether any digest mechanism exists

### Fork isolation of the write surface
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §5: "the strongest isolation is that the write surface points only at the fork … a designed slice, not huge" — proposed, not executed in these files.
- **what:** Point the loop's git-adapter credential, intake defaults and corpus indexing at a fork of the repo, making the main repo physically unwritable by the loop rather than protected by review discipline; the human pulls reviewed changes across. Folds in "mission and objectives correct in the first place": the fork's constitution/mission docs become the councils' curated context so conformance is checked against human-authored documents.
- **sources:** README_02_evidence_backed_proposals.md#5
- **relations:** human gate; guardian-from-decision-record; external rollback
- **verify-later:** git-adapter repo config; whether a fork exists

### Pilot worthiness test and the dartsonline guides pilot
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_fixloop(8): "★ THE PILOT IS CONFIRMED (2026-07-07) … Two earlier candidates were rejected … that triage history is itself the worthiness test working."
- **what:** A five-criteria test for whether a bug deserves the diagnosis loop, exercised through three candidates: the chrome/nav defect (dropped — got fixed; perishability lesson), the nav-links-to-never-rendered-pages defect (downgraded — root cause found by direct code reading, a known platform gap, reclassified to the builder route), and the confirmed pilot: dartsonline published a Guides nav link and a blank /guides/index.html while gamesdesign has working guides — a broken route, not a missing feature, with a standing hypothesis (reconcile_site_plan's routing table omits "guide"; nav derives from the plan, not the built set), mandatory pre-check queries and a cross-site differential as evidence. Establishes "genuinely mechanism-unclear" as the admission bar.
- **sources:** HANDOFF_fixloop_thread(8).md; HANDOFF_fixloop_thread(3)-(5).md deltas (the triage history)
- **relations:** eval gate; site-plan reconciler routing table (the suspected mechanism)
- **verify-later:** reconcile_site_plan routing table in load_work_item_actions.go; the F0 PILOT section of the fixloop runbook (U14)

### Seeded-bug strategy for the first end-to-end write run
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §6 and README_overview both recommend it; "the system's first-ever PR will have earned every gate it passed" — proposed, awaiting deploys.
- **what:** Because the only real test bug never yields an approvable plan (correctly escalated as architecture-level), the write step is tested by planting a contained single-file defect with an obvious symptom on a low-stakes surface and running the full pipeline — diagnose → plan → council (genuine approval) → implementer → PR. Rejected alternatives: hand-approving a known-flawed plan (contradicts the reviewers), waiting for an organic small bug (unbounded).
- **sources:** README_02_evidence_backed_proposals.md#6; README_overview.md
- **relations:** fix-implementer; eval gate
- **verify-later:** whether the first PR happened (git history for fix/* branches)

### Transferable machinery: legacy-migration and feature intakes
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §3: migration agents "not built, but it's the same machinery with a different intake"; features from specs "honestly furthest away … plausible; not designed".
- **what:** The allowlist/gate/council scaffolding is intake-agnostic: a legacy migration is "pattern X supersedes pattern Y" (scanner finds Y-shaped code, proposer writes constrained plans, council reviews, PRs flow); feature-building from mission docs needs a new grounding tier ("cite the spec clause this serves") — same shape as causal citation but not designed.
- **sources:** README_02_evidence_backed_proposals.md#3
- **relations:** council pattern; hard gates
- **verify-later:** n/a (unbuilt)

### "Documentation is code" — the context-assembly tool and paid service
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PLAN_context_assembly_tool_and_service(2) status: "A thin slice of Phase 1 is built and exercised on real code … first real run: a 30 KB bundle vs the script's 1.7 MB"; Phases 2–4 and the service unbuilt.
- **what:** For any development task, assemble a task-scoped context bundle from documentation + codebase and feed generation against ground-truth verification, so results are "more likely to be correct" than pasting code into a chat. The thesis: in an AI-driven workflow documentation (standards, intent, trajectory) is an operational input — version it, drift-detect it, compose it deterministically. Two audiences: dogfood on the chassis repo first, then a paid multi-tenant service behind the gateway. Design principles: engine/config split (tenant-agnostic engine + per-stack adapters, the decision that makes the service possible); seams for the optimal machinery (cascade router, decision-point checkers, mediator) defined as interfaces from v1; dogfood first. Phases: 0 contracts+constitution, 1 bundle builder MVP, 2 verification loop, 3 service (sandboxed verification is gating), 4 cascade/checkers/mediator.
- **sources:** PLAN_context_assembly_tool_and_service(2).md; 001_onboarding_discussion.txt; MAPPING_tool_to_actions_and_agents(2).md
- **relations:** bundle shape contract; six governance contracts (Phase 0); onboarding as the hard service problem; thin-slice-first principle
- **verify-later:** contextkit module state vs the plan's thin-slice claims; gateway project status

### Bundle shape contract (the task-scoped context package)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) "Status: contract specification"; flagged by FOCUS_whole_plan_review §1.1 as "the next load-bearing contract".
- **what:** A bundle is assembled fresh per task (never stored/reused, so never stale) with fixed sections: metadata; task/target; the authored layer (constitution, why-chain, priority profile, direction-of-travel, matched standards); code context (in-scope code in full, neighbourhood signatures, reuse-search results, schema, definition data); database data in three kinds; pointers to everything not inlined; and provenance (exactly what went in, logged as the decision log's inputs_used). Exists in a canonical structured form and a rendered text view. Two integrity rules from the edge-case pass: assemble from a consistent snapshot (no torn reads), and log what the generator SAW (rendered form), not what was assembled.
- **sources:** PLAN_bundle_shape_contract(2).md; FOCUS_whole_plan_review.md#1.1; FOCUS_pre_build_edge_cases(1).md#1.4,#4.3
- **relations:** decision log inputs_used; altitude/step-type; multipass fetch; contextkit assembler (the harness prototype)
- **verify-later:** whether any structured bundle object exists in code vs the markdown-emitting harness

### Three kinds of database data (definition / operational / content)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.1, verified against the live schema ("workflows are jsonb columns on agent_definitions … there is no workflows table and no tools table").
- **what:** In a data-defined system much of "the code" is DB rows, so the bundle distinguishes: definition data (the system's design as data — workflows in agent_definitions, tools as content_components rows, prompts as text columns; fetched routinely, covered by reuse-search); operational data (telemetry — work items, orchestration_states, error logs; multipass-capped); and content data (the output — sites/pages/tenant data; the gated set where privacy matters in the service).
- **sources:** PLAN_bundle_shape_contract(2).md#2.1
- **relations:** multipass fetch; reuse search over definitions; sensitivity gates
- **verify-later:** n/a (contract)

### Multipass fetch: probe → gate → include/reduce/point
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext "row data with multipass sizing" built in the harness (tool plan status); the full gate flow is contract-only.
- **what:** Query results have unknown size until run, so the builder probes with LIMIT N+1 (not count(*) — counting a filtered query costs as much as running it), checks a size gate and a sensitivity gate, then includes rows in full, reduces (aggregate / representative sample / pointer), or gates behind confirmation. An oversized result becomes an aggregate or pointer, never an unbounded dump.
- **sources:** PLAN_bundle_shape_contract(2).md#3; FOCUS_pre_build_edge_cases(1).md#4.2; GUIDE_deploy_from_context_packs(1).md (dbcontext)
- **relations:** three kinds of DB data; bounded bundle egress
- **verify-later:** dbcontext.go sizing logic

### Runtime evidence keyed by orchestration_id (the run narrative)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext -runtime-site built and used throughout (bundles carried agent_error_log/site_work_items); the fuller run-trace composition is contract-level.
- **what:** For debugging, the most useful context is the narrative of a run, reconstructable from one key: orchestration_id (+ correlation_id) spans orchestration_states (spawn tree, topics, status), llm_call_log (time-ordered step sequence), agent_error_log (error trail), pod logs (grep by run id) and the Kafka messages. Three cheap reads give a coherent single-run story instead of a scatter of lines. Log-correlation only works where the id is actually in the log line — a convention whose coverage the conventions agent audits.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2; PLAN_onboarding_agent_specs(6).md#2.9,#1.9; README_02 ("everything durable, one correlation id")
- **relations:** run signatures; codebase-conditional capabilities; diagnose_load_runtime
- **verify-later:** whether orchestration_id reliably appears in pod log lines (named open item)

### Run signatures: expected-vs-actual sequence diff
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2 — designed; capture/storage named open in §9.
- **what:** Capture a healthy run's step sequence and spawn-tree shape once (from known-good runs, confirmed), store as authored reference, and on a debug task diff the actual run against it to surface the divergence point — "matched the healthy path to step 7, then diverged here" instead of "read the logs". Verification applied to runtime.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** runtime evidence by orchestration_id; diagnostic playbooks
- **verify-later:** n/a (unbuilt)

### Diagnostic playbooks / failure fingerprints as authored knowledge
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2: "seeded from those guides rather than authored fresh"; home (standards atoms vs sibling table) open.
- **what:** Known failure fingerprints — a failure's signature + the commands that confirm it + the fix pattern — curated from the existing debugging guides and failure writeups, surfaced into debug bundles the way standards are surfaced into build bundles, and grown as run-signature diffs reveal new ones.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** run signatures; debugging guide 016 (the seed corpus)
- **verify-later:** n/a (unbuilt)

### Codebase-conditional capabilities (degrade, don't break; partial config is normal)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.3/§2.4 and agent-specs §2.9 — design rules, engine unbuilt.
- **what:** The definition-data / run-trace / run-signature / log-correlation capabilities rest on structural facts (behaviour stored as data, a run-correlation key, named logged steps, a known log fetch) that hold on our codebase but may not elsewhere. Stack-discovery records which facts hold; each capability degrades to a weaker form or states "unavailable, because this codebase has no X" instead of breaking. Companion rule: distinguish not-yet-authored config (degrade gracefully, note what's pending) from malformed config (fail loud) — the no-fallbacks rule applies to malformed data only.
- **sources:** PLAN_bundle_shape_contract(2).md#2.3-2.4; PLAN_onboarding_agent_specs(6).md#2.9; FOCUS_pre_build_edge_cases(1).md#2.3; FOCUS_whole_plan_review.md#2.5
- **relations:** stack-discovery agent; convention coverage = capability reliability
- **verify-later:** n/a (design)

### Altitude: step type decides what the bundle emphasises
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** The harness assembler takes `-step framing|debug` (used in real bundle invocations, e.g. "-step debug", "the framing bundle that produced this plan used -step framing").
- **what:** The same task at different stages needs different context: framing/decision steps get full intent (why-chain, priority profile, direction-of-travel) with light code; implementation gets full code with a thin intent tether; debug leads with errors + runtime evidence + the expected-vs-actual diff. "Right altitude at the right moment" made concrete in the bundle composer.
- **sources:** PLAN_bundle_shape_contract(2).md#4; PLAN_imagery_sprite_sheet.md (framing-bundle use); tasks/gameslink_missing_index_rerender/RUNBOOK…(2).md (-step debug)
- **relations:** bundle shape contract; salience-loss problem
- **verify-later:** assembler.go step handling

### Go analyser + call-graph neighbourhood (and the wiring-include gap)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: analyser/assembler "built and exercised on real code"; internal/analysis "already exists in the repo" per migration(19); bundles record "533 files analysed".
- **what:** A go/ast walk producing a structured index (signatures, types, per-function `calls` by name-based matching — full go/types resolution deliberately avoided as too heavy/fragile); the assembler slices in-scope symbols in full plus a signature-level caller/callee neighbourhood. Known structural blind spot: registration/init wiring (registry.go) is unreachable via calls, so `-include` exists for wiring files — the same gap as manually-named docs. ReadSymbolBody (span-slice over start_line/end_line) was extracted, tested byte-identical to cmd/assembler, and shared with diagnose_assemble_bundle.
- **sources:** 001_claude_reasoning; PLAN_context_assembly_tool_and_service(2).md#5 status; GUIDE_deploy_from_context_packs(1).md; PLAN.md changelog (ReadSymbolBody)
- **relations:** analyser adapter (wraps the same library); evidence-follows re-scoping; broad-script-vs-lean-assembler tradeoff
- **verify-later:** internal/analysis (chassis) vs contextkit copy drift (flagged in PLAN.md)

### contextkit module packaging and the graduation seam
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: "The tools are now a single Go module, contextkit/ … the two contracts defined once — internal/analysis and internal/candidates"; production-relevant parts graduated in-cluster per migration(19).
- **what:** All harness tools (analyser, embed, resolve_targets, fuse, eval_targets, assembler, dbcontext, bundle, dedup, diagnose) live in one module with two shared contracts; graduation moves the internal packages under the chassis module path and turns command mains into actions without changing the contracts. Production runs in-cluster; the harness remains the dev/measurement scaffold (eval_targets stays offline, the flywheel's measurement tool). The trial's output is throwaway; the rule it teaches is durable.
- **sources:** PLAN_context_assembly_tool_and_service(2).md status; PLAN_workflows_and_actions_migration(19).md (analyser-adapter sections); MAPPING_tool_to_actions_and_agents(2).md
- **relations:** analyser adapter; one-decision-core two realisations (same seam idea)
- **verify-later:** go_files/contextkit module contents (unit U17 territory)

### code_symbols: the per-repo code index (pgvector sibling table)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-09: "code_symbols applied cleanly (table + 4 indexes)"; later populated (436 symbols, embeddings 436/436 per the repo-label workaround header).
- **what:** A sibling to knowledge_base reusing its proven shape (vector(768) nomic + trigram on content + idempotent dedup, same AIService embedder with caller-applied nomic prefixes) but keyed for code: unique(repo,path,symbol), SHA-versioned via commit_sha, identity upsert (ON CONFLICT DO UPDATE WHERE content_hash IS DISTINCT — a symbol persists across commits, re-embeds only on change), pruned hard on re-index. Deliberate departures from chassis conventions (no version/previous_version_id, no deleted_at) because it is a rebuildable cache. HNSW chosen over the KB's IVFFlat for incremental churn (pgvector 0.8.0 confirmed both). One symbol = one row; no prose chunker (rag_index's character windows fragment Go mid-function).
- **sources:** NNN_create_code_symbols_index.sql; PLAN_workflows_and_actions_migration(19).md A5/B4/B4b + code-indexer reuse mapping
- **relations:** lookup/index_code_symbols; embedding policy split; knowledge_base reuse-not-copy
- **verify-later:** \d code_symbols; row counts per repo; embedding_model column use

### Hybrid code retrieval: index/lookup_code_symbols (vector + trigram, RRF in SQL)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19): actions built against real rag_actions.go, registered, deployed with the analyser (2026-06-12); lookup live in the diagnose workflow.
- **what:** index_code_symbols flattens analyser Output to symbol rows, skips unchanged content hashes, embeds non-fatally (trigram still works without an embedding), upserts and prunes; lookup_code_symbols mirrors rag_lookup (embed query → cosine vector search → trigram fallback → top-k + code_context). The hybrid RRF fusion moved into SQL, so the in-Go fuse tool never graduated. Deliberately a sibling action, not a parameterised rag_lookup (KB columns are baked into vectorSearchKB); the three embedding helpers are shared package-level functions.
- **sources:** PLAN_workflows_and_actions_migration(19).md (code-indexer reuse mapping + consumer-side-built entries); NNN_create_code_symbols_index.sql (query set)
- **relations:** code_symbols table; rag_lookup/rag_index (the mechanism source); repo-label symmetry
- **verify-later:** code_symbols_actions.go; registry entries (storage/code categories)

### Analyser adapter: in-cluster polyglot parsing service
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-12: "ANALYSER DEPLOYED TO PRODUCTION (uk_001) … the code-indexer agent row is applied (INSERT 0 1)."
- **what:** A Kafka message-based adapter (mirroring git/thunder, not HTTP) consuming `analyse` requests (owner, repo, ref, language), fetching source read-only via the GitHub tarball endpoint (one request, recovers the exact commit_sha, path-traversal-guarded), parsing through the analysis.Analyse library behind an `Analyser` seam so per-language parsers (JS next) drop in, and replying on the caller's responses topic. Security: its own least-privilege repo-scoped read-only token as a k8s Secret mounted only on this pod — two narrow credentials (analyser read, git-adapter write), never one broad token. Built polyglot-ready NOW because the JS tools already exist (tech debt, not future planning).
- **sources:** PLAN_workflows_and_actions_migration(19).md (analyser adapter sections + repo access & security); FOCUS_js_tools_documentation.md
- **relations:** analyse_repo_local (the in-process alternative that later took the diagnose+indexer paths); adapter envelope contract
- **verify-later:** internal/adapters/analyser; analyser-adapter deployment in uk_001; whether request_repo_analysis still has callers

### Adapter response envelope contract (single-sourced)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Migration(19) 2026-06-10/11: "Envelope contract resolved empirically … The contract now lives once in 035_adapter_guide.md §1 (FOCUS_adapter_design fully merged … and retired)".
- **what:** Resolved from code, not docs: the coordinator claims awaited requests on in_response_to_request_id first (request_id fallback); working adapters use typed body headers with real booleans + ProduceWithValidation; every adapter reads `action` from the BODY; a reply without the right Kafka headers silently falls through to process-as-work and times out (the documented thunder fault — found and fixed in the analyser adapter before deploy). Import-reuse verdict: reuse canonical types for the body, add a local Kafka-header builder (canonical ResponseHeaders lacks request_id/message_id/ToKafkaHeaders). A 003-vs-FOCUS documentation contradiction was settled empirically and single-sourced.
- **sources:** PLAN_workflows_and_actions_migration(19).md (guideline audit + dispatcher fix + envelope resolution entries)
- **relations:** analyser adapter; doc-drift (docs contradicted; code decided); 035_adapter_guide (canonical home, another unit)
- **verify-later:** 035_adapter_guide.md §1; whether 003 §832-890 was replaced with the pointer (was PENDING)

### Text-vs-code embeddings: share the mechanism, separate the policy (B4b)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) B4b "decided 2026-06-09"; schema embodies it (sibling table + embedding_model column).
- **what:** Shared and not duplicated: the AIService embedder seam, provider implementations, pgvector, and the semantic+trigram hybrid pattern. Separate per domain so each upgrades independently: the model (prose vs code-specific), dimension, preprocessing (nomic search_ prefixes are caller-side), retrieval tuning (HNSW vs IVFFlat, lexical-heavier for code), row definition, and evaluation. Turns B4a into "which model for code", measurable independently. Caution recorded: separation pays only if the mechanism stays shared.
- **sources:** PLAN_workflows_and_actions_migration(19).md B4a/B4b/A5 resolutions
- **relations:** code_symbols; B4a ceiling; CPU-Ollama feasibility (bulk-index speed, code-domain recall)
- **verify-later:** embedding_model column values; whether a code-specific model was ever adopted

### Code-indexer agent, index-orchestrator wrapper, and CI-triggered indexing
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** code-indexer row applied (2026-06-12); index-orchestrator seeded (v2 after the v1 `name`-column failure, 2026-07-02); CI trigger still a queue item ("CI-triggered indexing (self-contained)" in HANDOFF_builder_thread).
- **what:** The code-indexer is a thin orchestrator (analyse → index_code_symbols → complete). Run 6dfa37cd proved orchestrate+agent_type is adopted IN-PLACE on the shared chassis pod — which never holds GITHUB_READ_TOKEN — so the index-orchestrator wraps it in the proven spawn pattern so the spawned pod receives the secret (isRepoCloningAgent gate). Manual trigger TRIGGER_code_indexer_v2.sh sends the explicit branch ref; the planned durable form is a GitHub Actions step firing the envelope with GITHUB_SHA on push, retiring the index-staleness class for the diagnosis corpus.
- **sources:** NNN_create_code_indexer_agent(2).sql; NNN_seed_index_orchestrator(1).sql; TRIGGER_code_indexer_v2(1).sh; HANDOFF_builder_thread.md#3
- **relations:** analyse_repo_local staleness incident; spawn-consumed columns lesson; reuse-index freshness (governance)
- **verify-later:** index-orchestrator row; CI workflow file existence

### Documentation indexing rides the prose rag path
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Migration(19) "Documentation indexing (related, lighter…)" — direction recorded, not reported built.
- **what:** Docs are prose, so they fit rag_index/knowledge_base as-is: index each guide under a collection (e.g. `standards`), retrieve with rag_lookup; flat files in git remain the editable source of truth, the DB copy a derived rebuildable index so the assembler pulls relevant sections instead of pasting 124KB guides. Precondition: docs must live in a versioned repo. Separate, smaller workstream from code_symbols.
- **sources:** PLAN_workflows_and_actions_migration(19).md (documentation indexing section); FOCUS_js_tools_documentation.md
- **relations:** JS tools documentation gap; standards/docs agent (the matched-guidelines slot)
- **verify-later:** knowledge_base collections for docs

### JS tools documentation and provenance gap
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** FOCUS_js_tools_documentation: "Status: flagged 2026-06-09. Not started."
- **what:** The platform's JS tools have no prose docs and no code-symbol provenance; the only documentation is origin history (site/plan specs). Three separated needs: prose documentation (language-agnostic rag path, the main gap), code-symbol provenance (waits on the analyser adapter's JS parser drop-in), origin history (exists, a seed not a substitute). Open: docs' git home, a coverage signal, and whether docs and symbols share a tool identity key.
- **sources:** FOCUS_js_tools_documentation.md
- **relations:** analyser adapter polyglot seam; documentation indexing; tool-doc header contract
- **verify-later:** where JS tool sources live; any tool docs collection

### cmd/bundle robustness contract (validate early, fail loud, manifest input)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** 003_contextkit_bundle_issues (2026-06-24) — two real failed runs analysed; HANDOFF_fixloop(8) notes the skip-message/usage patches exist but "needs gofmt + build".
- **what:** Field-found tool defects and the rules they teach: validate the cheapest input (-analysis JSON) BEFORE the slow psql-shelling gather phases, with an actionable message naming file/size/regeneration; accept a manifest/config file instead of 20-line backslash shell commands (kills the unquoted-parentheses class — real filenames contain "(1)"); a missing -doc/-scope path must fail loudly naming the path, because a silently-omitted file means a downstream chat reasons from incomplete context without knowing; single quoted -psql argument, no TTY.
- **sources:** 003_contextkit_bundle_issues.md; HANDOFF_fixloop_thread(8).md#2
- **relations:** bundle-first handoff practice; fail-loud-vs-degrade rule
- **verify-later:** cmd/bundle/main.go precondition ordering; -config support

### Bundle-first handoff practice (context packs; broad script vs lean assembler)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** The tasks/ folder is this practice operating: one-sentence descriptions + primed bundle commands per task; GUIDE compares the two gathering modes on real artefacts (1.7MB script vs 30KB bundle).
- **what:** Every task handoff pairs a one-sentence problem statement with a filled-in cmd/bundle invocation (scope, docs, schema tables, runtime target), so a fresh chat starts from assembled context. Two gathering modes with a stated tradeoff: the package_*.sh directory-concatenating script (broad, thorough, catches wiring like registry.go, over-includes) vs the analyser/assembler (narrow, lean, call-graph-blind to wiring). Manual expert manifests were used as ground truth to validate the tool ("we're automating what experts already do by hand: call-graph slices, constitution rules, reference docs"). Advanced form: self-resolving bundle scripts that grep the analysis to locate action files (bundle_minilobby_trim v2's resolver, with PIN_ overrides).
- **sources:** tasks/any_project_handoff/001_build_bundle_ask_for_handoff; GUIDE_deploy_from_context_packs(1).md; 001_claude_reasoning; tasks/vonc_provocations_lobby/bundle_minilobby_trim(3).sh; tasks/missing_game_on_games_page/001_bundle
- **relations:** travelling-docs pattern; cmd/bundle robustness; constitution-in-every-bundle
- **verify-later:** n/a (practice)

### Reuse search before generation (code AND definition rows)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Named in the tool plan §2 and bundle contract §2.1/§8; only the retrieval substrate (code_symbols, pgvector) exists — the pre-generation reuse step is unbuilt.
- **what:** Before generating, run "what already does something like this" over existing functions/structs and — critically in a data-defined system — over definition rows (workflows/agents/tools), so reuse-before-recreate is mechanical rather than a remembered habit and near-copies of existing workflows are caught like duplicate functions. Needs a searchable text projection for jsonb definitions (named open). The index is derived state that goes stale silently — re-index on change events and stamp freshness.
- **sources:** PLAN_context_assembly_tool_and_service(2).md#2; PLAN_bundle_shape_contract(2).md#2.1,#8; FOCUS_pre_build_edge_cases(1).md#2.4,#15
- **relations:** reuse_index_refresh trigger; code_symbols; dev-guide reuse discipline
- **verify-later:** any reuse-search action; definition-row indexing

### DB capabilities capture (\dx/\df into the bundle)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** Migration(19) B2a "decided, mechanism built" in the harness (dbcontext -capabilities, assembler -dbfacts); the durable indexing-workflow half is future.
- **what:** Generation that writes SQL should see installed extensions and helper functions (so it knows pgvector exists and reuses snapshot_agent instead of hand-rolling a backup — the migration-110 footgun). Captured as DB context (not the analyser's job), included for DB-touching tasks with a reuse nudge; the durable plan folds capture into the indexing workflow on a migration cadence so bundles always carry current DB facts without anyone remembering a flag.
- **sources:** PLAN_workflows_and_actions_migration(19).md B2a + 2026-06-09 changelog
- **relations:** multipass fetch; schema-before-SQL discipline
- **verify-later:** dbcontext -capabilities flag; assembler dbfacts section

### Doc-drift claim classifier (grounded, tiered, read-only)
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** DESIGN_doc_drift_classifier: "Status: design only. The prompt contract (§3) is the part that must be right before any code."
- **what:** A per-claim pass over a document deciding current / stale / unverifiable against the real system with the evidence-or-abstain contract (quote or you may not verdict; no proposed rewrites; unverifiable routes to keep-untouched). Evidence gathered at the shallowest settling tier: T1 static (code_symbols, \d), T2 state (read-only SELECT), T3 behavioural (existing logs/rows, NEVER triggering a run). The output is a per-document report; no file is moved or merged. Historically the parent of the diagnosis loop's verdict contract.
- **sources:** DESIGN_doc_drift_classifier.md; DESIGN_diagnosis_loop(3).md (contract reuse)
- **relations:** cite-or-abstain verdict; conformance-suite carve-out; claim taxonomy
- **verify-later:** whether any classifier code exists

### Claim taxonomy: code-checkable / superseded-but-not-wrong / code-invisible
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** DESIGN_doc_drift_classifier §1 ("carried from item 24") — design.
- **what:** Three buckets of doc claims by checkability: mechanically confirmable facts (the classifier's target); decisions whose holding-but-not-rationale the code confirms (partial signal); and design intent / negative results the code says nothing about — disproportionately why old docs are worth keeping. Buckets 2/3 must reliably route to keep-untouched, never a confident verdict.
- **sources:** DESIGN_doc_drift_classifier.md#1
- **relations:** classifier; classify-don't-merge
- **verify-later:** n/a (design)

### Classify, do NOT merge (the human consolidates)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** DESIGN_doc_drift_classifier §5 "the line held firmest"; echoed as working practice in engines_tree_proposal ("selective-carry-with-the-LLM-as-assistant, never a generative merge").
- **what:** Grounding makes checking tractable but does not make generative merging safe: an LLM rewriting N docs into one fails silently (a dropped caveat reads as clean prose; no code-check catches an omission). The tool finds and cites; the human decides and writes; every canonical doc stays human-authored. Applied as a standing rule across the doc work.
- **sources:** DESIGN_doc_drift_classifier.md#5; engines_tree_proposal.md
- **relations:** classifier; engines tree migration
- **verify-later:** n/a (principle)

### Date/version as triage, not truth
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** DESIGN_doc_drift_classifier §4.1 — stated as a settled refinement.
- **what:** A recent file is more likely current; an old file is not more likely wrong — it is more likely unchecked. Dates order the queue and break ties; they never override a code check (recent docs went stale within hours in observed cases). Code decides; date orders.
- **sources:** DESIGN_doc_drift_classifier.md#4.1
- **relations:** claim classifier; misattribution asymmetry
- **verify-later:** n/a (principle)

### Standing conformance suite (carved out, deliberately not built)
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** DESIGN_doc_drift_classifier §0/§6: "CARVED OUT … Built later, on demand, as its own thing."
- **what:** A continuous "does the live system behave as documented" monitor on the existing DiscoveryCheckContext/CheckResult rails, scheduled, allowed fenced probes the doc pass forbids. Deliberately separated from the one-off classifier so the heavyweight always-on thing doesn't get built under a cleanup's banner and sink both.
- **sources:** DESIGN_doc_drift_classifier.md#0,#6
- **relations:** classifier; improvement-loop checkers
- **verify-later:** n/a (unbuilt)

### Engines docs tree + single _archive graveyard
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** engines_tree_proposal: "a TARGET to migrate toward … not a big-bang restructure"; _archive/ dirs now exist in git status (partially enacted).
- **what:** Three kinds of thing kept apart: engine code (in the module), engine docs (one canonical file per engine under engines/, pointing at canonical sources rather than restating), and archive (one _archive/ root, never indexed, replacing the go_files_old/docubundle/(N).go sprawl; the dedup tool's default target, giving the analyser a single -exclude). Runbooks split from engine docs because how-to-run rots at a different rate than how-it-works. Migration: dedup report → move → human editorial consolidation → re-point links → re-index.
- **sources:** engines_tree_proposal.md
- **relations:** classify-don't-merge; B4a clean-index prerequisite; dedup tool
- **verify-later:** whether engines/ was created; _archive contents

### Travelling-docs pattern (runbook = plan, notes = history, handoff = session)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Used throughout: HANDOFF_fixloop(8) header names it; running_notes_2 "Memory is OFF; this doc is the journal. Present this file at the END OF EVERY TURN."
- **what:** Each long-running thread maintains a runbook (the plan/map), running notes (chronological decisions with rationale, checkpointed), and a handoff (the complete start state for a fresh context, updated as discussion takes positions, with a file manifest and an opening move). Handoffs restate standing rules every time — the manual precursor the constitution automates. Parallel threads carry explicit boundary-awareness sections (what NOT to work on here).
- **sources:** HANDOFF_fixloop_thread(8).md; HANDOFF_builder_thread.md; tasks/005site_scheme_palette_and_components/running_notes_2(5).md; tasks/005site_scheme_palette_and_components/HANDOFF_scheme_to_components.md
- **relations:** bundle-first handoffs; three-thread working; constitution
- **verify-later:** n/a (practice)

### Three parallel threads with hard boundaries
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread §1: "THREE PARALLEL THREADS with hard boundaries" (builder/spine, tools, site-quality), with joint decisions marked ON HOLD.
- **what:** Concurrent chat threads own non-overlapping territories (relay spine + coordination; tool-pipeline internals; page-facing quality), each with its own runbook/notes; cross-territory scope changes route back through the owning thread; joint seams (e.g. the planned-tool-page seam §B5) are explicitly flagged as joint decisions and parked. Boundary files ride in each thread's manifest read-only.
- **sources:** HANDOFF_builder_thread.md#1,#4; HANDOFF_fixloop_thread(8).md (tools-chat courtesies)
- **relations:** travelling docs; classifier-consolidation queue
- **verify-later:** n/a (practice)

### Concept register and the council-of-concept-experts mission
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** README_comprehensive_documentation_categorisation is the originating user prompt (extract → classify → later verify against code → later create per-concept council agents); stage 1 in progress (this register).
- **what:** The user's programme: sweep every docs/ file for concepts (aspirational, deployed, superseded, unfulfilled), classify them into the docs024-style categories, later verify each concept's true state against chassis code/workflows/DB, and ultimately seed an expert agent per concept area to join the diagnosis/fix-loop council. Documentation categories are intended to correlate with council-reviewer expertise areas.
- **sources:** README_comprehensive_documentation_categorisation.md; README_claude_conversation.md (source chat URL)
- **relations:** expanded council bench; docs024 documentation index
- **verify-later:** n/a (this project)

### Three-layer config: mechanical / conventions / intent (different derivability)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation "Status: plan"; nothing in these files claims implementation.
- **what:** Onboarding is three processes, not one: the mechanical layer is discoverable (inspect + probe; low stakes, confirmable by reality); conventions are inferred or doc-sourced (a strong draft, weak authority — code shows what it does, not what it should do); intent and standards are elicited (not derivable from source; the tenant is the source, and the part delivering the tool's distinctive value).
- **sources:** PLAN_onboarding_config_derivation.md#1; 001_onboarding_discussion.txt
- **relations:** the five onboarding agents; docs-authoritative decision
- **verify-later:** n/a (plan)

### Progressive onboarding — a ramp, never "done"
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §2 — plan.
- **what:** A tenant gets value from the mechanical layer alone (fresh code context, signatures, reuse search, schema) before any intent is captured; conventions and intent fill incrementally and the tool deepens as they arrive. Onboarding tracks the repo forever — active-with-pending is the steady state, and leaf-level intent is captured just-in-time during use rather than as a setup tax.
- **sources:** PLAN_onboarding_config_derivation.md#2; PLAN_onboarding_agent_specs(6).md#3.7,#4.3
- **relations:** intent-elicitation agent; config-maintenance agent
- **verify-later:** n/a (plan)

### Config as a maintained artifact (the wizard is the first pass; the lifecycle is the deliverable)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §3 — plan.
- **what:** The derived config drifts as the repo changes, so it gets the standards' own upkeep machinery: periodic re-derivation with divergence flagging, confirm-not-initiate on proposed changes, and per-entry provenance (discovered/inferred/supplied) determining trust and change authority. "Onboarding as a first-class deliverable" means this lifecycle, not a good setup script.
- **sources:** PLAN_onboarding_config_derivation.md#3; 001_onboarding_discussion.txt
- **relations:** config-maintenance agent; active-config provenance shape
- **verify-later:** n/a (plan)

### Inference quality scales with codebase quality — surface uncertainty
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §4 — named tension, design mitigation.
- **what:** On a messy repo, convention inference confidently drafts the repo's bad habits, and confirming that codifies the mess — so the more a tenant needs the tool, the less their repo can teach it. Mitigation: surface inconsistency as questions to resolve, never a silent majority pick; inconsistency found during onboarding is itself valuable output ("your conventions aren't actually conventions").
- **sources:** PLAN_onboarding_config_derivation.md#4; 001_onboarding_discussion.txt
- **relations:** conventions agent; docs-authoritative mode
- **verify-later:** n/a (plan)

### Docs-authoritative conventions for our own repo (the free drift audit)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §5 "Ours (decided): docs-authoritative" — the decision is recorded; the audit has not run.
- **what:** Source-of-truth for conventions is chosen per tenant by doc availability; for our repo, 001/003/the naming FOCUS are authoritative and code is read only to find disagreements. Each disagreement is recorded, not silently resolved — the set is a free audit of where the codebase drifted from its own documented standards, the drift detector's first run, on us. Our own onboarding is the template, not a special case.
- **sources:** PLAN_onboarding_config_derivation.md#5,#7; 001_onboarding_discussion.txt
- **relations:** conventions agent; drift audit three-bucket output
- **verify-later:** whether any drift audit ran

### Conventions agent (extract-cite-confirm, then audit)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1 — spec only.
- **what:** Owns the conventions layer: extracts discrete convention atoms from the standards docs (each citing its exact doc span — extraction is inferred-then-confirmed, because auditing code against an invented convention manufactures fake drift, the one failure that would discredit the audit), gets the set human-confirmed BEFORE any audit, then checks code and records disagreements with location/convention/tier/confidence and a default disposition (code-drifted, doc-drifted, or legitimate exception — human confirms). Accepted exceptions are remembered so audits become incremental.
- **sources:** PLAN_onboarding_agent_specs(6).md#1
- **relations:** three checking tiers; docs-authoritative decision; check_*.go validators
- **verify-later:** n/a (spec)

### Three checking tiers + three-bucket audit output (coverage honesty)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 — spec; the pattern recurs in maintenance §5.5.
- **what:** Conventions (and drift) are checked at three tiers: deterministic (static check settles it → violations), heuristic proxy (a measurable indicator flags candidates, not violations — "where to look, not what's wrong"; an optional LLM pass is still only a candidate flag, never a verdict), judgement-only (no proxy → reported as a coverage gap). The audit reports three numbers, never one — a clean tier-1 count beside many unchecked tier-3 conventions is a partial audit with known limits, and must say so. Companion role split: un-auditable conventions still serve as generation guidance (an atom can be audited, guiding, or both).
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#1.6; PLAN_onboarding_agent_specs(6).md#5.5
- **relations:** conventions agent; config-maintenance drift tiers; LLM-as-candidate principle
- **verify-later:** n/a (spec)

### Convention coverage IS capability reliability
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 closing — spec insight.
- **what:** When a bundle capability rests on a manual convention (log-correlation needs orchestration_id in every log line), the capability is only as reliable as the convention's coverage, not its existence. For any capability-bearing convention the audit reports how completely it is followed, and gaps surface as fixable (add the missing log statements) rather than hard limits — even on our own codebase, where the structure exists but coverage is unverified.
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#2.9
- **relations:** codebase-conditional capabilities; runtime evidence by orchestration_id
- **verify-later:** an orchestration_id logging coverage scan

### Stack-discovery agent (inspect → interpret → declared probe plan → probe → confirm)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2 — spec only.
- **what:** Owns the mechanical layer: read-only inspection emits facts; interpretation ("this Makefile target is probably the test command") emits proposals with confidence — the subtle point being that interpretation has uncertainty even at the mechanical layer; a declared probe plan (the security contract, kept even for our own use as audit) precedes sandboxed probes; probe results update confidence. A failing build is useful output, candidate-only interpreted, never fixed by this agent. The output document carries per-entry source/confidence/probe-result with uncertainties listed separately. Also records the structural facts bundle capabilities depend on (§2.9).
- **sources:** PLAN_onboarding_agent_specs(6).md#2
- **relations:** confirmation by reality; sandboxing envelope; codebase-conditional capabilities
- **verify-later:** n/a (spec)

### Confirmation by reality (the mechanical layer climbs the ratchet first)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.8; contract-set review §3.1 records the reconciliation.
- **what:** The mechanical layer can be confirmed by observation (the probed command actually works) — the strongest confirmation any config layer carries — so stack-discovery is the natural first capability to graduate past confirm_every. Reconciled with the gate: probe success is initially strong evidence inside the work-item gate (near-rubber-stamp, human still activates); only after trust-ledger graduation does probe success auto-activate. The gate is the starting position; graduation relaxes it — not a bypass.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.8; FOCUS_contract_set_review.md#3.1; PLAN_active_config_schema(3).md#5
- **relations:** trust ledger; confirm-not-initiate
- **verify-later:** n/a (design)

### Sandboxed probing — the tenant-code security envelope
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.6: "gating for the service phase: no tenant code runs until sandboxing is solid."
- **what:** The first agent that may execute tenant code does so inside an ephemeral sandbox: repo mounted read-only, restricted network, time limit, no persistent state; the emitted probe plan is the contract the sandbox approves/restricts/denies per command. The Tier-C security concern made concrete; same gate applies to Phase-2 verification running tenant code.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.6; PLAN_context_assembly_tool_and_service(2).md#6
- **relations:** stack-discovery; service phase
- **verify-later:** n/a (unbuilt)

### Intent-elicitation agent (progressive, value-returning interview)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §3 — spec only; reuse target (briefing_questionnaire column) verified real in FOCUS_schema_verification_findings §3.
- **what:** Captures the why-chain, per-node priority profiles and direction-of-travel via an interview that interleaves proposal-confirmation (where evidence exists — low friction, anchoring risk mitigated by citing the evidence so proposals are contestable) with free elicitation (blank page, unavoidable). Every exchange returns value (the captured piece changes the next bundle/mediation); the interview is not finite — leaf intent is captured just-in-time in the flow of work. A descendant of the briefing questionnaire / intake orchestrator, pointed at a codebase. Capture and use are separate roles (the user-rep advocate consumes what this captures). Open: detecting rubber-stamping.
- **sources:** PLAN_onboarding_agent_specs(6).md#3; FOCUS_schema_verification_findings.md#3
- **relations:** onboarding orchestrator; objectives table; user-rep advocate (salience doc, other unit)
- **verify-later:** n/a (spec)

### Onboarding orchestrator (dependency-graph flow; active-with-pending)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §4 — spec only.
- **what:** Coordinates the three layer agents: stack-discovery first (both others depend on its mechanical config), conventions and intent in parallel (independent of each other) — sequencing follows dependencies, not policy. Routes all proposals through confirm-not-initiate; surfaces a compact onboarding-state artifact (per-layer confirmed/partial/blocked, pending, drift-audit counts); a blocked layer doesn't stop the others; a tenant walking away pauses cleanly. Terminal state is active-with-pending, handing over to maintenance — never "fully done".
- **sources:** PLAN_onboarding_agent_specs(6).md#4; FOCUS_onboarding_system_view_check.md#1,#3.4
- **relations:** the three layer agents; config-maintenance handoff; work-items queue
- **verify-later:** n/a (spec)

### Config-maintenance agent (drift detection as the trust ratchet's signal source)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §5 — spec only.
- **what:** After baseline, detects drift across all three layers, event-driven (change-layer diffs) plus a periodic sweep, dispatching to the layer agents for re-checks rather than reimplementing them; targeted re-validation (implicated-only recheck) instead of full sweeps. Drift evidence uses the same three tiers; surfacing is prioritised to avoid alert fatigue (high-impact deterministic first, heuristic in paced batches, freshness nudges background). Its deeper role: sustained no-drift is graduation evidence and repeated drift is de-graduation evidence — without this agent the bidirectional ratchet has nothing to act on at the right timescale.
- **sources:** PLAN_onboarding_agent_specs(6).md#5; FOCUS_onboarding_system_view_check.md#2
- **relations:** trust ledger; change-layer integration; published-reasoning gap detection
- **verify-later:** n/a (spec)

### Active-config schema (four tables, computed-on-read effective values)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_active_config_schema(3) "Status: contract specification"; corrected to chassis conventions after schema verification.
- **what:** The load-bearing contract: tenant_configs (scope-holder row per tenant, created directly at init — not a gate violation), mechanical_config (one JSONB row, per-field embedded provenance), standards (flat concern atoms with scope constitution/domain/leaf, applies_to change types, rule/rationale/check/check_kind), objectives (nested why-chain nodes with priority_profile, direction_of_travel, standing_concerns). A common provenance shape (source/source_ref/confidence/status/last_verified_at/verified_by/freshness_until/version/previous_version_id/deleted_at) across all layers so consumers reason uniformly. Effective priority profile is computed at read time by walking root→node (store authored differences, compute effective on read); acyclicity must be enforced on write AND the walk bounded, since a human can confirm a cycle. The constitution is a view over standards WHERE scope=constitution, not a table. Two atom trees deliberately kept distinct: flat concern tree vs nested objective tree.
- **sources:** PLAN_active_config_schema(3).md; FOCUS_onboarding_system_view_check.md#3.1,#3.7; FOCUS_pre_build_edge_cases(1).md#1.1
- **relations:** all six contracts hang off it; bundle authored layer reads it
- **verify-later:** whether any of the four tables exist in clients_db

### Governed vocabularies and the hand-authored first constitution (prerequisites)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §6 — named prerequisites, "currently assumed, not called out"; a thin_slice_constitution.md flat file exists and rides in every bundle.
- **what:** The concern taxonomy (standards.concern) and priority dimensions are fixed vocabularies the conventions/intent agents classify INTO, so they must be authored before those agents run. The first constitution is hand-written from 001/003 + working preferences (the tool that would help write it doesn't exist yet); the thin-slice flat-file constitution is its interim form, later becoming standards rows with scope=constitution. Also: "us" is a real tenant row, not a sentinel, so single-tenant exercises the multi-tenant code path.
- **sources:** FOCUS_pre_build_edge_cases(1).md#6; PLAN_active_config_schema(3).md#1,#3.1; tasks/gameslink bundles (constitution section present)
- **relations:** active-config schema; thin-slice-first
- **verify-later:** thin_slice_constitution.md content vs standards rows

### Confirm-not-initiate + the single central confirmer (one path to active)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_contract_set_review §2.1 "Resolution applied: a single central confirmer"; all contracts remain "contract specification" status.
- **what:** Agents propose (status=proposed rows + a work item); a human confirms; ONE component applies uniformly — flip to active, set last_verified_at/verified_by, deprecate the prior version, write the decision-log entry, emit the in-band change event — so confirm-not-initiate is a status-transition rule enforced in one auditable place, not a discipline reimplemented per agent. Hardening from the edge-case pass: the apply is one DB transaction with the change event in an outbox (crash-consistent, retry-safe), idempotent (re-applying an active version is a no-op), one live proposal per target extending down to layer rows (a new proposal replaces the proposed row; expiring a work item deprecates its row), and work items reference proposed rows by identity not pinned version.
- **sources:** FOCUS_contract_set_review.md#2.1,#2.3; FOCUS_pre_build_edge_cases(1).md#1.2,#1.3,#12; PLAN_config_work_items_contract(3).md#4
- **relations:** config_work_items; decision log; change layer in_band guard
- **verify-later:** n/a (unbuilt)

### config_work_items contract (mirror of site_work_items, tenant-scoped)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_config_work_items_contract(3) "contract specification"; corrected against the real site_work_items shape (FOCUS_schema_verification_findings §2).
- **what:** The shared queue every onboarding/maintenance agent emits into and the tenant reads: a parallel table (site_id NOT NULL blocks direct reuse) mirroring the verified shape — item_type/spec/result naming, integer priority, the real status lifecycle (detected→triaged→approved|rejected→claimed→complete|failed), reuse of approval_mode (the pre-existing confirm-not-initiate field; config defaults manual, 'auto' only for graduated capabilities), item_key unique-partial dedup (one live item per target), depends_on/parent_item_id, retry machinery. Batch confirmation for the initial onboarding flood (approval granularity adapts; apply still honours dependency order). Explicit scope: gates config, not deliverables.
- **sources:** PLAN_config_work_items_contract(3).md; FOCUS_schema_verification_findings.md#2; FOCUS_pre_build_edge_cases(1).md#2.1,#15
- **relations:** central confirmer; two-gated-paths; site_work_items (the reuse source)
- **verify-later:** table existence; site_work_items approval_mode semantics in code

### Decision log (immutable; premise vs rule_trace; inputs_used)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_decision_log_contract(2) "contract specification".
- **what:** The published-reasoning log given shape: append-only, one row per decision, carrying either a human-readable premise (judgement decisions) or a rule_trace (mechanical ones — exactly one of the two, so mechanical steps don't produce noise premises), plus inputs_used: the active-config slice in hand at decision time (compact atom id+version references + merged-view hashes by default; full snapshot inline for high-stakes kinds). Resolves freshness-vs-retrospect: compute on read for freshness, log at point of use for reconstruction. Write discipline: every decision logs, the entry precedes the apply, logging is not itself a logged decision. Read patterns: drift detection (premise vs current profile), heuristic invalidation, trust-ledger evidence, retrospective audit, compliance review. Open seam flagged: bundle assemblies would dominate a reasoning log by volume — bundle provenance may belong as the consuming decision's inputs_used instead.
- **sources:** PLAN_decision_log_contract(2).md; FOCUS_pre_build_edge_cases(1).md#4.4,#11; FOCUS_whole_plan_review.md#2.2
- **relations:** bundle provenance; trust ledger; work-item resolutions feed premises
- **verify-later:** n/a (unbuilt)

### Trust ledger + bidirectional ratchet (asymmetric by design)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_trust_ledger_contract(4) "contract specification".
- **what:** Mutable state, one row per (tenant, capability): trust_level ∈ confirm_every/confirm_exceptions/notify/autonomous plus derived gate_policy and evidence_summary; derived from but separate from the immutable decision log (different access patterns). Cold start is always confirm_every (trust earned per tenant; no cross-tenant inheritance — deferred deliberately). Graduation up is always confirm-not-initiate; de-graduation down may auto-apply with notification on severe evidence — losing trust shouldn't wait on a human, gaining it should — but de-graduation evidence must first pass the defect-vs-partition filter so a flaky test or infra blip can't drop a capability and trigger a confirmation flood. Cascade routers and gate policy engines read it at runtime.
- **sources:** PLAN_trust_ledger_contract(4).md; FOCUS_pre_build_edge_cases(1).md#2.2; FOCUS_whole_plan_review.md#2.1
- **relations:** capabilities catalog (the ceiling); maintenance agent (the evidence source); outcome-record gap
- **verify-later:** n/a (unbuilt)

### Capabilities catalog: the ceiling lives on the capability (blast radius caps trust)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_capabilities_catalog_contract(1) "contract specification. The sixth contract, closing the trust ledger's open dependency."
- **what:** A sibling table to agent_definitions (its existing capabilities jsonb holds free descriptive kebab tags for discovery, deliberately left alone; catalog capability_ids are snake_case dispatch keys — a recorded naming decision) holding per-capability ceiling, verifiability and containment. The ceiling is a judgement over the two factors (the weaker holds it); stored for cheap reads but the factors are authoritative — a factor change triggers a gated ceiling re-proposal. Capabilities aren't 1:1 with agents; the operation→capability mapping is declared at the action level. Seeding principle made explicit: the more a capability can break — especially chassis-editing ones — the lower its ceiling, regardless of verifiability; never fully autonomous for chassis-touching capabilities.
- **sources:** PLAN_capabilities_catalog_contract(1).md; FOCUS_pre_build_edge_cases(1).md#13; FOCUS_whole_plan_review.md#1.4
- **relations:** trust ledger; recursive self-improvement risk; cascade router
- **verify-later:** n/a (unbuilt)

### Change-layer integration (change_events; in_band closes the self-modification loop)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_change_layer_integration_contract(4) "contract specification. … Closes the final contract gap before implementation."
- **what:** Change events from four sources (git_webhook, polling, in_band, periodic_sweep, + manual) land as first-class change_events rows (at-least-once with commit dedup; no event silently dropped — triggers_fired=[] is an explicit record). The trigger filter mapping changed paths to typed maintenance triggers is computed from the mechanical config, not stored (compute-on-read applied to routing). in_band emission — the tool's own applies emit events — is what keeps self-modification visible to the drift detector and decision log; rule: state changes emit, computed-view refreshes don't. Guard: a confirmer apply doesn't re-trigger maintenance on the entry just confirmed, but genuine downstream effects (audit code against a newly-active convention) still fire, and generation-origin in_band events are never exempt. reuse_index_refresh is its own trigger because a stale reuse index fails silently.
- **sources:** PLAN_change_layer_integration_contract(4).md; FOCUS_whole_plan_review.md#1.2; FOCUS_pre_build_edge_cases(1).md#4.1,#2.4
- **relations:** config-maintenance agent; central confirmer; reuse-search freshness
- **verify-later:** n/a (unbuilt)

### Two gated paths: config changes vs deliverables
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §3 "a conceptual conflation to fix"; restated in the trust-ledger and work-items contracts.
- **what:** Changes to the tool's knowledge of the codebase (standards/objectives/mechanical) flow config_work_items → central confirmer → active config. The tool's outputs (generated code; edits to workflows/agent definitions — deliverables even though they're DB rows) flow cascade → trust-ledger gate → apply+commit+in_band event. The decision log spans both; the gates are not the same gate, and there are correspondingly two gated-mutation mechanisms (config confirmer; ledger ratchet-evaluator with asymmetric de-graduation).
- **sources:** FOCUS_pre_build_edge_cases(1).md#3; PLAN_trust_ledger_contract(4).md#1; PLAN_config_work_items_contract(3).md#5; FOCUS_whole_plan_review.md#2.1
- **relations:** trust ledger; config_work_items
- **verify-later:** n/a (design)

### The outcome-record gap (the loop runs on outcomes nobody sources)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §9 "real gap … You can log every input and decision and still have no feedback signal."
- **what:** The contracts log decisions and inputs, but nothing records whether a deliverable succeeded — verification pass/fail, reverted, human-corrected, accepted-as-is — the raw signal evidence_summary must aggregate for the ratchet to move. Companion gap (§10): "the bundle helped" has no defined metric (candidate signals: fewer correction rounds, fewer convention violations, less manual context-gathering); both needed before Phase 2.
- **sources:** FOCUS_pre_build_edge_cases(1).md#9,#10
- **relations:** trust ledger; thin-slice premise test
- **verify-later:** n/a (gap)

### Thin vertical slice before the six-contract infrastructure
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** FOCUS_pre_build_edge_cases §8 recommended it; the tool plan's status note shows it happened ("a thin slice of Phase 1 is built … deliberately ahead of [the contracts] to test the core thesis first").
- **what:** The whole design rests on one unproven premise — an assembled bundle beats paste-and-rot — and none of the six contracts test it. So: hand-write a minimal flat-file constitution, build analyser+schema extractor, assemble ONE bundle for ONE real task, paste it by hand; only if it visibly helps build the infrastructure. This sequencing was followed: the thin-slice harness shipped and was used on real bugs while the contracts stayed specifications.
- **sources:** FOCUS_pre_build_edge_cases(1).md#8,#16; PLAN_context_assembly_tool_and_service(2).md status
- **relations:** contextkit harness; the six contracts (their build deliberately deferred)
- **verify-later:** n/a (executed strategy)

### External rollback (the self-hosting trap) + recursive self-improvement as residual risk
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §5, §14 — stated rules/risks, no implementation claimed.
- **what:** The tool runs on the chassis it modifies; a bad change could break the chassis badly enough that the tool can't run to fix it. Rule: rollback to known-good must be runnable externally, with no dependency on the agents/orchestrator being rolled back. And a self-improvement that passes verification can still degrade the tool's judgement gradually — not fully solvable; managed by conservative early trust, the human gate, external rollback and low ceilings for chassis-touching capabilities, and named as an accepted residual risk rather than assumed closed.
- **sources:** FOCUS_pre_build_edge_cases(1).md#5,#13,#14
- **relations:** capabilities-catalog ceilings; human gate; fork isolation
- **verify-later:** existence of any external rollback path

### Morality review as a configured, layered standard (not a baked-in view)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** MAPPING_tool_to_actions_and_agents(2) — design discussion; review contributors "(none yet)" in the thin-slice column.
- **what:** Distinct from liability ("will this get us sued" vs "is this right"): a build-time review contributor applying a layered standard held in the active config — an operator-chosen recognised base source (ASA/CAP Code, CMA guidance; OECD/UNESCO/NIST for the AI angle), operator values layered above it, jurisdiction/current-focus overlays later. Two altitudes: per-output, and a vertical-level gate at intake (should we build this site/industry at all). Contested cases route to HITL; the tool applies the configured standard and flags — it is not the moral authority.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md (morality review section)
- **relations:** build-time review contributors; active-config standards layer; council compliance eye
- **verify-later:** n/a (unbuilt)

### Contributors vs checkers (build-path reviews ≠ improvement-loop monitors)
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** MAPPING(2): "Checkers are a different concept — not these" — a settled terminology/architecture distinction (reuse overlap flagged to investigate).
- **what:** Context contributors assemble bundle slices (code/data/runtime/standards); build-time review contributors (reuse, near-duplicate, liability, morality, correctness) review a PROPOSED change before it ships, raising concerns that revise or HITL-gate; improvement-loop checkers (the check_*.go family) continuously monitor DEPLOYED sites against plan/spec in the operate layer. Two layers restated: the website-builder builds sites; the context tool builds reliable changes to the builder.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md; PLAN_workflows_and_actions_migration(19).md (group-agent reviews)
- **relations:** council pattern (the reviews' fix-loop descendant); improvement-loop category
- **verify-later:** whether build-time reviews reuse check_*.go logic (flagged open)

### Chassis conventions verified: text+CHECK, previous_version_id, deleted_at
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** FOCUS_schema_verification_findings §1 — read off the live schemas; applied as corrections across the contract set.
- **what:** The live chassis conventions every new table must follow: enumerated values are text with CHECK constraints (never native enums); versioning is version integer + previous_version_id uuid self-FK with unique (type,version); soft delete is deleted_at (never a status=archived); timestamptz defaults now(); jsonb for flexible payloads. The verification pass also corrected wrong reuse assumptions in the contracts ("the contracts are corrected to match reality, not the reverse") and confirmed real fields: approval_mode, pipeline, item_key, briefing_questionnaire, input/output_contract, agent_category CHECK set.
- **sources:** FOCUS_schema_verification_findings.md; PLAN_active_config_schema(3).md#2 note
- **relations:** schema-before-SQL discipline; all six contracts
- **verify-later:** n/a (verified from live schema dumps)

### Work-item relay spine (batons, handler_agent, the 30s pump)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread §2 "Builder route §B0–§B4 CLOSED: spine decided = the work-item relay (§B3, pre-registered criteria)"; README_flows describes the live pump.
- **what:** Builds move as a relay: the baton is a site_work_items row naming a handler_agent; build-pipeline-trigger (every 30s, behind a pre-query gate) seeds build_queue and picks one dispatchable site; build-dispatch-loop claims items atomically and spawns a dynamic handler per item. One hop = one baton, one agent, one site_specs entry, one new baton. Around it a fully enabled immune system: evidence-based claimed-item timeout (its SQL documents the gamesdesign false-positive lesson), feasibility-recheck, both reapers, archiver, cleanup — with improvement-sweep currently disabled (the improvement loop not running) while content-feed-refresh runs six-hourly.
- **sources:** HANDOFF_builder_thread.md#1,#2; README_flows.md
- **relations:** hop-insertion pattern; needs_diagnosis intake (reuses the relay); scheduler-and-tasks category
- **verify-later:** scheduled_tasks rows; improvement-sweep flag state

### Vertical-exemplar research hop (best-of-niche synthesis into vertical_landscape)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread: "§B4 vertical-exemplar-researcher LIVE and quality-verified end to end on dartsonline.com … causal synthesis (confidence 0.82) → strategy QUOTING the landscape".
- **what:** A new relay hop between classification and strategy: find the vertical's best existing sites, read three of them shallowly (deliberate budget: limit 6, markdown, main-content only, depth 1 — vs adoption's one-site-deep 30/rawHtml/4), and distil WHY they succeed — reasons, not copies — into spec aspect vertical_landscape for the strategist and planner. Reuse-only (every step an existing action; the whole agent is one DB row, no Go, no image build); written as a spec because specs are the per-site shared memory across hops; inserted via reroute (classifier chains needs_vertical_research; researcher chains needs_strategy onward; priority 7 below strategy's 8 in the ascending ladder); an optional strategist prompt nudge makes the strategy step weigh the new aspect (research nobody reads is wasted). First bare-domain→deployed-site milestone followed.
- **sources:** README_flows.md (the plain-language explainer); NNN_seed_vertical_exemplar_researcher(2).sql; NNN_reroute_classifier_to_vertical_research.sql; NNN_strategist_vertical_landscape_nudge.sql; HANDOFF_builder_thread.md#2
- **relations:** relay spine; adoption pipeline (contrasting crawl budget); site-spec-and-classifier
- **verify-later:** vertical-exemplar-researcher row; vertical_landscape aspects in site_specs

### Spawn-consumed columns lesson (seeds copy image columns from a live donor)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NNN_fix_researcher_spawn_columns documents the incident and fix; HANDOFF_builder_thread carries it as a standing guard ("seeds must copy image columns from a live donor (the amended seed does)").
- **what:** getAgentDefinition SELECTs image_repository/image_tag/command/resources/health_config/env_vars/idle_timeout_seconds and gates on is_active=true; a seed populating only default_config leaves command NULL → the image's default entrypoint boots the GENERIC chassis service, which never reads the injected AGENT_TYPE env, so the dispatcher's call goes unheard and the item stays claimed. Fix and rule: copy the spawn-consumed infrastructure columns from a proven donor (deliberately NOT capabilities/topics/default_config). Related: image_tag DEFAULT 'latest' pointed at an ancient build; the makefile now pins IMAGE_TAG. Sibling gotchas carried with it: pod label key is agent-type (hyphen); check body.status not just the header; error_step belongs INSIDE step config (step-level silently ignored); idle pods reap ~3600s with ProcessingHistory dumps as post-reap evidence.
- **sources:** NNN_fix_researcher_spawn_columns.sql; HANDOFF_builder_thread.md#2,#5; HANDOFF_fixloop_thread(8).md#3
- **relations:** workflow-in-default_config lesson; index-orchestrator spawn wrapper
- **verify-later:** guidelines 001 New Agent checklist line (flagged residual)

### Confirmed chassis workflow model (group agents, promotion pattern, wrapper orchestrator)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Migration(19) "Confirmed model (from the guideline docs …)" plus 2026-06-09 changelog confirming against two real agent_definitions rows and the action code.
- **what:** The model the migration work verified: workflows are declarative JSON steps in default_config; a generic config-driven action library (query_database, spawn_agent, call_agent, loop, conditional, rag_lookup, work-item lifecycle) is reused before writing Go; LLM-assisted checks group into one agent per shared context load (explicitly rejecting a registry of mini-action agents), promoted to spawned sub-agents only when one needs independence (a one-line workflow change); the wrapper orchestrator (spawn→call→complete) is the canonical small form; spawning is spawnAgentKubernetesJobFromDefinition with per-spawn job topics. Reuse discipline is encoded as queries (search agent_definitions; default_config::text ILIKE '%<action>%').
- **sources:** PLAN_workflows_and_actions_migration(19).md (confirmed model + changelog); DESIGN_diagnosis_loop_chassis_integration(6).md#0
- **relations:** diagnose pair; index-orchestrator; onboarding agents (all reuse it)
- **verify-later:** 001/002 guideline docs (canonical home, other units)

### Adoption writes first; classifier consumes (the corrected lean)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** README_flows: "your instinct was half right, inverted … apply_adoption_plan writes the specs, pages, and work items itself — it never calls domain-research-classifier."
- **what:** The adoption orchestrator is a thin spawn→call wrapper; the agent crawls via firecrawl, extracts design/interactive fingerprints without an LLM, runs three LLM analyses (site structure; archetype snapshot with improvement-loop constraints; content-direction style guide), and apply_adoption_plan writes specs/pages/work items directly. The classifier later consumes adopted specs under fidelity rules when the relay reaches the site. Parked question: does apply_adoption_plan write site_archetype (classify_archetype's output isn't in its declared inputs).
- **sources:** README_flows.md
- **relations:** relay spine; classifier consolidation queue
- **verify-later:** apply_adoption_plan_action.go; site_archetype writer

### Classifier consolidation + site_type taxonomy alignment (queued work)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_builder_thread Q5/Q8 — a queued brief with a read-first plan, not executed in these files.
- **what:** Two classifiers overlap: domain-research-classifier (newer, relay-native) and site-classifier (pageflow/intake path, which does NOT use work-item triage — its differences may be load-bearing: intake's hitl_confirm_type keys off confirmed_type.recommended_builder). Brief: diff both rows, map dependency points, check hard before changing; merge additions both ways with snapshot migrations; deprecate only at zero usage. Behind it: the classifier and strategist use different canonical site_type vocabularies in the same spec chain — one decision, two snapshot prompt migrations.
- **sources:** HANDOFF_builder_thread.md#3; README_flows.md (Q8 note)
- **relations:** adoption-writes-first; vertical-exemplar hop
- **verify-later:** both classifier rows; intake usage query on orchestration_states

### Thunder training-worker probe status taxonomy
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** README_worker_statuses: "The worker's probe status_command encodes exactly this taxonomy, and it lines up with the plan."
- **what:** GPU training worker liveness as four probe outcomes: ALIVE (pgrep finds the training run → reset streak), DONE_OK (RUN_SH_DONE marker + adapter_config.json exists → mark_complete → decommission), DONE_FAIL (RUN_SH_FATAL → mark_failed → decommission), GONE_UNKNOWN (process gone, no marker — crash/OOM/reap → bump streak, mark_failed at 3 consecutive unreachable probes).
- **sources:** README_worker_statuses.md
- **relations:** model-infrastructure lifecycle/reaper concepts (docs009 units)
- **verify-later:** the status_command in the thunder worker config

<!-- SECTION-F -->
