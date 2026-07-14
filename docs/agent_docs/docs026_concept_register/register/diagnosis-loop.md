# Register — diagnosis-loop
41 concepts, consolidated from 130 raw extractions across units U03, U05, U06, U08,
U09, U12, U13, U14, U15, U16, U17b, U18, U19, U23, U24c, U24f, U25.

### DIAG-001 — Read-only, cite-or-abstain diagnosis loop (core concept)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6G "PASSED — run 51f95cda (2026-07-01): abstain → correct reads → REFUTE the naive framing → CONFIRM the grounded cause"; code_retrieval_route(21) "ROUTE CLOSED — 2026-07-03 (run 73ed55c6)"; NOTES v4(39) "Diagnosis loop (§6): DONE."
- **what:** An AI agent that investigates a bug strictly read-only: forms a hypothesis, gathers scoped evidence (code bodies, read-only DB rows, runtime records), issues a verdict that must CITE evidence or ABSTAIN (CONFIRMED/REFUTED/UNVERIFIABLE), then re-scopes by FOLLOWING the evidence (call graph for code, vetted/model-written queries for data) rather than re-searching the symptom. Never edits code, never runs builds, human-gated. Built first as a standalone Go engine (`contextkit/internal/diagnose/`) with a tested scaffold, then ported into the chassis as the `diagnose-agent`/`diagnose-orchestrator` pair. The hard problem it targets is falsification — abandoning a wrong hypothesis — not evidence discovery, which turned out to be comparatively easy.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#what-this-is; docs019/RUNBOOK_design_diagnosis_loop(7).md#overview; docs019/RUNBOOK_code_retrieval_route(21).md#route-closed; docs019/NOTES_running_synthesis_v4(39).md; docs019/DESIGN_diagnosis_loop(3).md#0-1
- **relations:** verdict cite-or-abstain contract (DIAG-003); convergence guards (DIAG-004); falsification-first eval gate (DIAG-018); chassis workflow architecture (DIAG-002); contextkit CLI toolchain (register/contextkit-toolchain.md) as its dev/eval harness ancestor
- **verify-later:** chassis `pkg/diagnose/` (loop.go, step.go, advance.go); `platform/orchestration/actions/diagnose_*_action.go`; agent_definitions rows diagnose-agent/diagnose-orchestrator

### DIAG-002 — Chassis workflow architecture (diagnose_route, next_step override)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6C "DONE"; "coordinator next_step override CONFIRMED (coordinator.go:1093 getNextStepFromResult)"; §6E "DONE 2026-06-29 (5× loop-back, CONFIRMED)".
- **what:** The loop is realised as a chassis agent workflow, not a new CLI or action-internal monolith: a thin `diagnose-orchestrator` (spawn_diagnoser → call_diagnoser → complete) spawns a `diagnose-agent` worker whose workflow is `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict (execute_llm_prompt) → route (diagnose_route) → [loop back to load_runtime | emit] → persist_note → complete`. `diagnose_route` runs the engine's Advance (guards + call-graph re-scope) once per iteration and overrides `next_step` in its result (the conditional_route pattern); it sets no `output_field` so its results are read as `route.*`. The workflow lives in `agent_definitions.default_config`, not the legacy `*_workflow` columns. Each verdict is its own observable `execute_llm_prompt` step rather than buried in a monolith.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C,#4b; docs019/NOTES_running_synthesis_v3(32).md STATE DIGEST; docs019/DESIGN_diagnosis_loop_chassis_integration(6).md#0,#status; NNN_fix_diagnose_agent_workflow(2).sql
- **relations:** diagnose-orchestrator spawn-wrapper pattern (DIAG-019); abandoned diagnose_run monolith (DIAG-020); abandoned diagnostician re-invocation design (DIAG-021); deterministic scaffold split (DIAG-017)
- **verify-later:** `platform/orchestration/actions/diagnose_route_action.go`; `coordinator.go` getNextStepFromResult; diagnose-agent default_config

### DIAG-003 — Verdict cite-or-abstain contract + wire format seam
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §7.5 "keep the prompt's output schema and verdict_wire.go in lockstep"; NOTES principles(59) 2026-06-17 "the diagnosis loop is now COMPLETE as far as can be built off-chassis... the tested prompt↔scaffold seam".
- **what:** The model-facing prompt (`PROMPT_diagnosis_verdict.md`) requires every verdict to return exactly one of CONFIRMED/REFUTED/UNVERIFIABLE with verbatim-quoted citations tagged by tier (static/state/runtime) with freshness; a citation-less confirm/refute is coerced to UNVERIFIABLE; abstention is asymmetric — runtime evidence readily refutes but confirms only on direct mechanism, never "consistent with"; re-scope must follow what the evidence names, not re-search the symptom. The wire format uses human-legible strings and snake_case keys (not int enums) so a model can produce it reliably; `diagnose.ParseVerdict`/`verdict_wire.go` map it to the domain Verdict with fail-safe unknown→UNVERIFIABLE coercion. Because the script format IS the wire format, every scripted test scenario is a faithful dry-run of the real model path.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4a; docs019/PROMPT_diagnosis_verdict(1).md; docs019/NOTES_running_synthesis_v2(36).md 2026-06-17; docs019/go_files/contextkit/internal/diagnose/verdict_wire.go#header
- **relations:** three-tier citation standard (DIAG-008); SeenRequests / data_request progress rule (DIAG-005); symptom-coverage gate family (DIAG-011); falsification-first eval gate (DIAG-018)
- **verify-later:** `pkg/diagnose/verdict_wire.go` + `verdict_wire_test.go`; PROMPT_diagnosis_verdict.md; live diagnose-agent default_config verdict step

### DIAG-004 — Four convergence guards plus engine-level failsafes
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6E.1 "the loop stopped at stopped_by: evidence-not-growing — a guard, not luck" (2026-06-29); migration(19) "Four convergence guards + the no-citation→UNVERIFIABLE coercion are all behaviour-tested."
- **what:** Deterministic, model-independent stop conditions: iteration cap (5), scope-not-narrowing (re-scope can't balloon past prior scope + 2), evidence-not-growing (two iterations with no new grounded citation → stop with best effort), and hypothesis-thrash (oscillation without discriminating evidence → report both) — plus engine-level `timeout_seconds: 1800` and `fuel_budget: 1000` bounding a runaway even if the loop's own bookkeeping is disarmed. Behaviour-tested (26-test suite), not eyeballed; deliberately kept in tested Go rather than re-expressed as workflow conditionals, so a model verdict can be left untrusted.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position,#6E.1; docs019/RUNBOOK_design_diagnosis_loop(7).md#0,#3; docs019/PLAN_workflows_and_actions_migration(19).md; docs019/NOTES_running_synthesis_v2(36).md
- **relations:** SeenRequests progress rule (DIAG-005); named-scope guard (DIAG-006); loop state threading self-check (DIAG-015); verdict cite-or-abstain contract (DIAG-003)
- **verify-later:** `pkg/diagnose/loop.go` guards; `loop_test.go`, `step_test.go`; step.go DecideStep

### DIAG-005 — SeenRequests / data_request progress-counting mechanism
- **status:** deployed
- **status-evidence:** RUNBOOK(31) "guardAfter now tracks issued read-only data_requests in a SeenRequests set... validated in run 51f95cda (3 iters, new queries each, no premature stop)"; NOTES(10)#Turn 12/13 "F0.5 — CODE-COMPLETE 2026-07-10 (from run 3)".
- **what:** Fix for the loop stopping one iteration before its own good query's answer would have arrived: the evidence-not-growing and hypothesis-thrash guards treat an issued NEW unseen read-only `data_request` as forward progress, while re-issuing the same query still trips the guard. Required a `verdict_wire.go` sync (an older chassis copy silently mapped DataRequests to null, making the fix inert). Later extended by the fix-loop's F0.5: fetched data_request answers were evaporating from the bundle after one iteration and tripping scope-not-narrowing, fixed by forwarding the UNION of current-verdict and prior-seen request keys (deduped, capped at 12) each loop-back — "re-run, don't store," avoiding collected_data-bloat.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#1 fix); fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 12,#Turn 13; docs019/NOTES_running_synthesis_v3(32).md DECISIONS (turns 30-31)
- **relations:** convergence guards (DIAG-004); data_requests channel (DIAG-010); three-guard SQL model (DIAG-009)
- **verify-later:** `pkg/diagnose/advance.go` SeenRequests; `loop_datarequest_test.go`; `pkg/diagnose/verdict_wire.go`

### DIAG-006 — Named-scope guard vs capped call-graph expansion
- **status:** deployed
- **status-evidence:** code_retrieval_route(21) "BOTH FIXES DELIVERED 2026-07-03... Guard now measures the MODEL-NAMED scope; expansion runs only after the guard passes and is CAPPED (Config.MaxExpandedScope, engine default 18)"; route-close run 73ed55c6 "the expansion cap bounding iterations 2–3 at exactly 18".
- **what:** Blocker found when the real 515-file corpus replaced a stale 69-file one: `guardAfter` measured the POST-EXPANSION scope, so unbounded call-graph expansion of six named symbols tripped scope-not-narrowing at iteration 1. Fix: the narrowing guard compares the MODEL-NAMED scope only (deduped NextScope, no expansion); expansion is used purely for the gather and capped at MaxExpandedScope (default 18, named entries always kept). A data_request escape on the scope guard was considered and rejected as near-inert.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D,#route-closed
- **relations:** convergence guards (DIAG-004); call-graph re-scope mechanism (DIAG-007); stale-corpus class (DIAG-023)
- **verify-later:** `pkg/diagnose/{loop,step,advance}.go` NamedScopeSize/MaxExpandedScope; `loop_scopeguard_test.go`

### DIAG-007 — Call-graph re-scope mechanism (evidence-follows re-scoping)
- **status:** deployed
- **status-evidence:** NOTES v2(36) 2026-06-17 "diagnosis-loop adapters... BUILT & tested"; code_retrieval_route(21) run 73ed55c6 confirming re-scope onto a symbol the symptom never named.
- **what:** On REFUTED/UNVERIFIABLE, the next bundle scopes the symbols/files the evidence NAMES plus their call-graph neighbourhood (callees/callers, name-matched not type-resolved, per the analyser's recorded `calls`), preferring a runtime-named fault site over a retrieval-proposed one, and deliberately dropping ubiquitous names (Run, String, Error, New, ...) so following the graph doesn't explode into noise — "the symptom-vocabulary trap in call-graph form." This is the move symptom-based retrieval cannot do, and it is the empirical answer to the B4a retrieval ceiling (see register/context-engineering-principles.md).
- **sources:** docs019/RUNBOOK_thin_slice(27).md#assembler-flags; docs019/RUNBOOK_design_diagnosis_loop(7).md#1a,#design-and-build-choices; docs019/NOTES_running_synthesis_v2(36).md; docs019/PROMPT_diagnosis_verdict(1).md rule 4; contextkit/internal/diagnose/callgraph.go#header
- **relations:** named-scope guard (DIAG-006); B4a retrieval ceiling (register/context-engineering-principles.md); Go analyser call-graph neighbourhood (register/context-assembly.md — the tool building the graph)
- **verify-later:** `pkg/diagnose/callgraph.go`; `internal/analysis/analyse.go` calls extraction; ubiquitous-name drop list

### DIAG-008 — Three-tier citation standard (static / state / runtime) with freshness
- **status:** deployed
- **status-evidence:** code_retrieval_route(21) "CONFIRMED on citations spanning ALL THREE TIERS... query+result cited together... a quoted code branch for the mechanism" (run 73ed55c6, 2026-07-03).
- **what:** Every verdict citation carries a tier — static (code/schema), state (a DB row at a point in time), or runtime (log/work-item from an actual run) — with observation time recorded for state/runtime citations, so a verdict resting on stale evidence is visibly weaker. The route's success bar is a CONFIRMED verdict grounded across all three tiers, distinguishing "confirmed by inference at the data layer" from "code-level mechanism named." Adapted from the doc-drift classifier's T1/T2/T3 scheme.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#route-closed; docs019/RUNBOOK_design_diagnosis_loop(7).md#1,#tier-vocabulary; docs019/PROMPT_diagnosis_verdict(1).md rule 6; docs019/DESIGN_doc_drift_classifier.md#2
- **relations:** verdict cite-or-abstain contract (DIAG-003); tier-coverage guard (DIAG-012); verdict wire format (verdict_wire.go citation struct)
- **verify-later:** citation tier handling in verdict_wire.go; evidence_trail of run 73ed55c6

### DIAG-009 — Three-guard read-only SQL enforcement model
- **status:** deployed
- **status-evidence:** RUNBOOK(31) "SELECT-only is enforced at THREE layers (confirmed in the code)"; design_diagnosis_loop(7) §4d "CONFIRMED on this cluster (2026-06-17): pool_mode = transaction... a live BEGIN READ ONLY; DELETE... WHERE false probe refused the write."
- **what:** Defence-in-depth for model-written SQL: Guard 1 = the verdict prompt constrains to one read-only SELECT; Guard 2 = `IsReadOnlySQL` lint applied twice (route boundary and pre-execution); Guard 3 = the actual guarantee, a `BeginTx(ReadOnly:true)` transaction plus statement_timeout that rejects any write including data-modifying CTEs. Guards 1–2 are hygiene, never the safety boundary. This reversed an earlier stance that "the model never writes SQL" — the bounded, triple-guarded version was built deliberately once the read-only transaction was proven sufficient as the real boundary.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#4,#6B; docs019/RUNBOOK_design_diagnosis_loop(7).md#4d; docs019/PROMPT_diagnosis_verdict(1).md rule 7; docs019/DESIGN_diagnosis_loop_chassis_integration(6).md#status; NNN_create_diagnose_ro_role.sql
- **relations:** data_requests channel (DIAG-010); sqlguard stripQuoted fix (DIAG-013); diagnose_ro role (DIAG-014); EXPLAIN pre-flight guard (DIAG-016)
- **verify-later:** `pkg/diagnose/sqlguard.go`; BeginTx call in diagnose_load_runtime; PLAN.md GATED item 1 (execution wiring untested at last update)

### DIAG-010 — data_requests channel: model-authored read-only SQL evidence gather
- **status:** deployed
- **status-evidence:** RUNBOOK(31) update 2026-07-01 (run 51f95cda) "the model's data_requests RAN (verdict_wire.go confirmed live)"; §6C "now wired (was dormant from a wiring gap, not by design)."
- **what:** The verdict may emit `data_requests` (single read-only SELECTs with `sql`/`why`); `diagnose_route` reads them from the verdict wire, keeps only read-only ones, forwards to `route.data_requests`; `diagnose_load_runtime` executes each on loop-back in a READ ONLY transaction with `SET LOCAL statement_timeout` and appends rows to runtime_evidence. Code re-scope and data re-gather are deliberately separate channels — this is the "DB-following" arm of evidence-following, replacing an earlier, more limited "vetted query catalogue only" design (which survives as a fast-path/few-shot layer, not the only path) once the read-only transaction was proven sufficient.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6B,#6C; docs019/NOTES_running_synthesis_v2(36).md; docs019/NOTES_running_synthesis_v3(32).md STATE DIGEST "Correction" paragraph
- **relations:** three-guard SQL model (DIAG-009); SeenRequests (DIAG-005); doc/query catalogue selection (register/context-assembly.md docselect/queryselect); live schema section (DIAG-015)
- **verify-later:** `diagnose_load_runtime_action.go` runDataRequests; `diagnose_route_action.go` readOnlyDataRequestsFromWire

### DIAG-011 — Symptom-coverage gate family (symptom_check → context disposition)
- **status:** deployed
- **status-evidence:** FYI 2026-07-10: "prompt rule 8 + symptom_check schema field applied (snapshot 34f4afc8); engine coercion rides the next chassis image post-v1.0.1101"; PLAN_fixloop_pilot.md §3b "F0.4d BUILT 2026-07-10", "F0.6 BUILT 2026-07-10".
- **what:** A CONFIRMED verdict must account for every distinct observation in the ORIGINAL symptom via `symptom_check: [{observation, explained, how, cites, context}]`, or the engine coerces it to UNVERIFIABLE. Built in three stages: (F0.4d) the base gate, motivated by a well-cited confirm that dismissed half the symptom as "not a nav issue"; (F0.6) added a `context: bool` flag exempting comparative/background clauses from explained/unexplained accounting, and required `explained:true` entries to carry an in-range `cites` index (fixing a grade-inflation defect where comparison clauses were marked explained despite their own text saying "unverifiable"). Terminal diagnosis notes gain a "Symptom coverage:" block. Owned by the fix-loop workstream; delivered to the docs019/travelling-docs unit as a courtesy collision-rule FYI since it edits the shared diagnose-agent verdict prompt.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md; fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 11,#Turn 14,#Turn 15; fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#BENCHMARK RUN 3
- **relations:** persist_diagnosis_note (note bodies change); tier-coverage guard (DIAG-012); symptom anchor (DIAG-024); doc_notes/travelling-docs coordination boundary
- **verify-later:** diagnose-agent verdict prompt_template; `verdict_wire.go` symptom_check parsing

### DIAG-012 — Tier-coverage guard (F0.4e)
- **status:** deployed
- **status-evidence:** "First production firing of any of the new guards" (NOTES(10)#Turn 12, run 3).
- **what:** A shared `coerceVerdict()` engine gate requiring a CONFIRMED verdict to carry at least one `static` citation AND at least one `state|runtime` citation, or it degrades to Unverifiable; REFUTED is exempt. Directly answers the benchmark finding that "cite-or-abstain does not prevent confirming the wrong cause" when all citations come from a single tier.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7,#Turn 8,#Turn 12
- **relations:** symptom-coverage gate family (DIAG-011); three-tier citation standard (DIAG-008); cite-or-abstain doctrine (DIAG-003)
- **verify-later:** grep/inspect `coerceVerdict()`; `static`; `state|runtime`

### DIAG-013 — sqlguard stripQuoted: lint false-positive on quoted literals
- **status:** partial
- **status-evidence:** RUNBOOK(31) run 5537ffdb (2026-07-01) "the page slug 'tool-drop-rate-simulator' contains 'drop'... FIXED (sqlguard.go stripQuoted blanks literal/identifier contents before the scan; regression test added)"; §6G banner "REMAINING: DEPLOY the lint fix — latent" (2026-07-02).
- **what:** Keystone bug: the read-only lint scanned raw SQL text, so a keyword substring inside a string literal (a slug containing "drop") caused a legitimate read to be silently dropped — neutralising both the schema-section content read and the SeenRequests progress rule. Fix blanks literal/identifier contents before keyword scanning. Written and tested; the runbooks record deployment to the live image as still pending at the family's last update.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G-passed,#update-5537ffdb; contextkit/internal/diagnose/sqlguard_literal_test.go#header
- **relations:** three-guard SQL model (DIAG-009); SeenRequests (DIAG-005)
- **verify-later:** `pkg/diagnose/sqlguard.go` stripQuoted + test; whether the deployed chassis image carries it

### DIAG-014 — diagnose_ro role and pooler-aware read-only enforcement
- **status:** partial
- **status-evidence:** RUNBOOK(31) checklist "§6B diagnose_ro role migration written... [ ] role migration applied"; "data_requests run via db.BeginTx(ReadOnly) on params.DB (clients_user), NOT a restricted role."
- **what:** A GRANT-only SELECT role (`diagnose_ro`) designed for the CLI harness path, where `psql -c` statement stacking makes a transaction wrapper unsafe. Standing doctrine: under pgbouncer, enforce read-only by GRANT, never by `SET default_transaction_read_only` (session settings leak across pooled backends); transaction pooling makes `BeginTx(ReadOnly)` safe; statement_timeout goes in the DSN options. The chassis (production) path deliberately runs as `clients_user` under the read-only transaction instead, so content tables stay SELECTable without extra grants.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4d; docs019/RUNBOOK(31)_diagnosis_loop.md#6B
- **relations:** three-guard SQL model (DIAG-009); dbcontext CLI (register/contextkit-toolchain.md)
- **verify-later:** NNN_create_diagnose_ro_role.sql applied?; pgbouncer-config pool_mode

### DIAG-015 — Live schema section in the bundle (gatherSchema)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) "the bundle now carries a ## Schema (live tables) section"; run 51f95cda "using REAL table/column names — the schema section paying off, no more page_sections guessing."
- **what:** `diagnose_load_runtime` gained one read-only `information_schema.columns` query, denylist-driven (`%backup%/%bak%/%archive%/%supersede%`, deliberately not `%snapshot%` since site_snapshots is live) plus a broad relevance include (`site%/page%/content%/flow%`) unless `schema_full=true`; rendered into the bundle via Go-defaulted config so no migration was needed. Stops the model guessing table names (it had previously invented `page_sections`).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel (DIAG-010); denylist-over-allowlist style (shared with index-hygiene excludes, DIAG-022)
- **verify-later:** gatherSchema in diagnose_load_runtime_action.go; runtime.schema render path

### DIAG-016 — EXPLAIN pre-flight size guard on data requests
- **status:** deployed
- **status-evidence:** RUNBOOK(31) "EXPLAIN size-guard added... runDataRequests now runs EXPLAIN (FORMAT JSON) inside the read-only tx BEFORE each query"; 51f95cda validation "the EXPLAIN guard (didn't block site-scoped queries)".
- **what:** Before executing each model-written query, the action plans it (EXPLAIN FORMAT JSON, no execution) and skips with feedback if estimated rows exceed budget (explain_max_rows 50000; cost cap opt-in); output rows are capped (row_cap 200) and cells truncated rune-safe (cell_chars 600); statement_timeout remains the execution backstop. A skip is feedback the model narrows from — a new narrower request still counts as progress.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel (DIAG-010); SeenRequests (DIAG-005); three-guard SQL model (DIAG-009)
- **verify-later:** runDataRequests EXPLAIN branch in diagnose_load_runtime_action.go

### DIAG-017 — Deterministic scaffold / model-only-verdict split
- **status:** deployed
- **status-evidence:** design_diagnosis_loop(7) "The scaffold is deterministic; the verdict is the only model-dependent part... This puts the SAFETY in code that is verified, and isolates the part that needs a model."; RUNBOOK(31) §3 "the merged assembler diffed against the pre-collapse binary — byte-identical."
- **what:** Architecture decision: loop control, guards, evidence trail, and re-scope are pure tested Go (`step.go`'s `DecideStep`, shared by the standalone `Run()` loop and the chassis-facing `advance.go`/`LoopState`, proven equivalent by test); the cite-or-abstain judgement is an interface (a stub that always abstains, scripted verdicts, or the live model) that runs as its own observable workflow step. A model-less run can never fabricate a conclusion. `cmd/diagnose` stays the dev/test harness, never a production entrypoint.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#design-and-build-choices,#4b; docs019/DESIGN_diagnosis_loop(3).md#1; contextkit/internal/diagnose/step.go#header; contextkit/internal/diagnose/advance.go#header
- **relations:** chassis workflow architecture (DIAG-002); verdict wire format (DIAG-003); diagnosis-loop scaffold internals (register/contextkit-toolchain.md)
- **verify-later:** `pkg/diagnose/` package purity (no DB imports); workflow verdict step config; advance_test.go proving Advance reproduces Run()

### DIAG-018 — Falsification-first evaluation gate (scaffold correct ≠ reasons well)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6G "PASSED 2026-07-01 (run 51f95cda)"; design_diagnosis_loop(7) §5 "A loop that confirms the first guess on every known bug is the failure mode, not the success — judge it on the reversals."; README_02 "The single enemy is confident wrongness. Runs 1–2 of the benchmark produced CONFIRMED verdicts that were wrong."
- **what:** The loop is never trusted on scaffold correctness alone; it must run against known bugs and (a) reproduce mid-course REVERSALS (refute wrong hypotheses on evidence), (b) converge on causes the symptom could never retrieve, and (c) ABSTAIN when the bundle doesn't settle the question — never confirm the first guess. This is the design premise of the whole project: LLMs rationalise their first hypothesis, so every downstream mechanism (citation mandate, REFUTED-is-correct framing, guards, symptom-coverage gate) exists to force explicit falsification. The §6G pass showed UNVERIFIABLE→REFUTED→CONFIRMED over 3 cited iterations.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G; docs019/RUNBOOK_design_diagnosis_loop(7).md#5,#6; docs019/DESIGN_diagnosis_loop(3).md#0; docs019/README_02_evidence_backed_proposals.md; docs019/PLAN.md GATED item 5
- **relations:** gamesdesign resolveResultSpec fixture (DIAG-025); verdict cite-or-abstain contract (DIAG-003); loop-worthiness test (DIAG-030); confident-wrongness framing (register/context-engineering-principles.md)
- **verify-later:** evidence trails of runs 51f95cda, 5537ffdb, 73ed55c6 in orchestration_states; whether trigger modes (b)/(c) beyond human-gated were ever enabled

### DIAG-019 — diagnose-orchestrator spawn-wrapper pattern + generic-request trigger envelope
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6F "Target the ORCHESTRATOR; diagnose-agent is the worker it spawns... keeping the loop off shared pods"; 122_diagnose_agents.sql seeds both agents 'experimental' "until the real-bug evaluation gate passes; promote to 'active' after"; incidents in migrations 126–129 show live runs 2026-07-06/10.
- **what:** The diagnosis entry point is a thin `diagnose-orchestrator` (spawn_diagnoser → call_diagnoser → complete), per "every agent is an orchestrator," that spawns a dedicated `diagnose-agent` pod and forwards its result — keeping heavy in-chassis loop work off the shared chassis pods (which never hold the repo-read token). Triggering is the existing generic-request envelope (kcat to `system.agent.generic.requests`, agent_type `diagnose-orchestrator`, input_data `{symptom, seed_scope, runtime_site, ...}`) — no new triggering code; later trigger modes (on build failure, proactive sweep) are the same envelope from a different sender, gated on the falsification eval passing. The same spawn-wrapper pattern was replicated for indexing (`index-orchestrator`) once in-place `orchestrate` proved token-less on shared pods.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1; docs019/DESIGN_diagnosis_loop_chassis_integration(6).md#2; NNN_restore_diagnose_orchestrator_workflow(1).sql; 122_diagnose_agents.sql; drafts/084_TRIGGER_diagnose_v1.sh
- **relations:** repo-cloning token gate (DIAG-022); code-indexer/index-orchestrator (register/context-assembly.md); chassis workflow architecture (DIAG-002)
- **verify-later:** diagnose-orchestrator/index-orchestrator agent_definitions rows; spawn_actions.go; promotion to 'active' status; evaluation gate results

### DIAG-020 — Abandoned design: diagnose_run single engine-wrapping action
- **status:** abandoned
- **status-evidence:** Chassis-integration(6) banner: "the §4–§6 diagnose_run recommendation below is the ABANDONED path... there is no diagnose_run action"; RUNBOOK(5) §6E "In this design there is NO workflow loop-back: the iteration lives inside the diagnose_run action."
- **what:** The originally recommended shape — one `diagnose_run` action calling `diagnose.Run()` with an injected Verdicter, keeping the whole capped loop inside a single orchestration step (iteration visible only in logs/trail). Dropped in favour of the workflow-driven observable loop (verdict as its own step, router action) so each iteration's verdict and routing are separately auditable. A prompt-registry reference `diagnose-verdict-v1` belonged to this design and is also unused; the seeded diagnose-agent briefly referenced the nonexistent action until the workflow-fix migration removed it. Family-delta: present in RUNBOOK(2)–(7), gone by RUNBOOK(8).
- **sources:** docs019/DESIGN_diagnosis_loop_chassis_integration(6).md#banner,#4-6; docs019/RUNBOOK(5).md#6E; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; NNN_fix_diagnose_agent_workflow(2).sql#header
- **relations:** chassis workflow architecture (DIAG-002, the replacement); deterministic scaffold split (DIAG-017)
- **verify-later:** absence of diagnose_run in registry.go; diagnose-agent workflow JSON

### DIAG-021 — Abandoned design: diagnostician per-iteration re-invocation (spawn-next chain)
- **status:** abandoned
- **status-evidence:** NNN_seed_diagnose_agents(2).sql banner "SUPERSEDED — DO NOT APPLY... kept only as a record of the re-invocation design that was considered and dropped."; RUNBOOK(31) §6C "Do NOT seed a new one (the diagnostician draft is superseded)."
- **what:** A third loop shape considered and dropped: each orchestration runs ONE iteration (load_runtime → analyse → lookup → assemble → verdict → route → conditional), and on continue spawns+calls a fresh `diagnostician` of the same type with revised hypothesis/scope/iteration in input_data, the terminal verdict bubbling up the child chain. Motivated by doubt that the engine supported a workflow-internal cycle, and by the build-dispatch-loop's one-unit-per-orchestration precedent. Dropped once the `next_step`-override loop-back was confirmed to work. Also covers the lineage: an early `diagnostician` single-agent draft → a seed-agents migration → superseded by fixing the already-seeded diagnose-agent/diagnose-orchestrator pair in place.
- **sources:** NNN_seed_diagnose_agents(2).sql#header; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK(2).md#E
- **relations:** chassis workflow architecture (DIAG-002, the replacement); standing evidence rules (snapshot_agent used before every agent_definitions mutation)
- **verify-later:** no `diagnostician` row in agent_definitions; migrations NNN_fix_diagnose_agent_workflow.sql, NNN_restore_diagnose_orchestrator_workflow.sql

### DIAG-022 — Repo-cloning token gate (isRepoCloningAgent)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) "spawn_actions.go injects GITHUB_READ_TOKEN via secretKeyRef... gated by isRepoCloningAgent -> ONLY diagnose-agent pods get the token; the spawner never holds it"; §7B.1 "isRepoCloningAgent gained 'code-indexer'... Verified end to end by run 93ba14e6."
- **what:** Least-privilege credential injection at spawn time: only agent types allowlisted as repo-cloning receive the read-only GitHub token env (via secretKeyRef into the spawned pod); the shared chassis pods never hold it. This is why indexing and diagnosis both run through spawning orchestrators rather than in-place.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1
- **relations:** diagnose-orchestrator spawn-wrapper (DIAG-019); analyse_repo_local (DIAG-023); analyser adapter secret lesson (register/context-assembly.md)
- **verify-later:** spawn_actions.go isRepoCloningAgent list

### DIAG-023 — analyse_repo_local: the diagnose-agent's self-contained repo fetch
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6F "DECIDED for option 3... BUILT (turn 27, gofmt-clean)"; checklist "image also carries analyse_repo_local + lifted internal/reposource — DONE 2026-06-29"; NOTES v3(32) "Symbol bodies come from a git checkout the diagnose-agent makes itself (Option 3, turns 25-26)."
- **what:** Resolves the "no repo checkout on the diagnose pod" blocker: the agent fetches its own tarball (`GET /repos/{o}/{r}/tarball/{ref}`, no git binary in the chassis image), reusing the analyser adapter's fetcher lifted into a neutral `internal/reposource` package, then runs `analysis.Analyse(dir)` in-process for both the call graph and symbol-body slicing — one fetch, no cross-pod coupling, git remains the only source of truth for code. `pin_to_index_commit` pins the fetch to the dominant code_symbols commit (true for the diagnose loop, false for the indexer, which defines the commit) so lookup-seeded symbols resolve in the fetched tree. Options rejected: storing bodies in the DB (whole-repo Kafka payloads), or a stateful analyser serving slices (per-iteration coupling).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B; docs019/NOTES_running_synthesis_v3(32).md turns 25-27
- **relations:** repo-cloning token gate (DIAG-022); analyser adapter (register/context-assembly.md); stale-index incident (register/context-assembly.md — same mechanism from the corpus-freshness angle)
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; NNN_swap_analyse_repo_to_local.sql

### DIAG-024 — Symptom anchor (F0.4a)
- **status:** deployed
- **status-evidence:** "F0.4a... CODE-COMPLETE 2026-07-09" then verified live in run 2 (PLAN_fixloop_pilot.md §3b, NOTES(10)#Turn 10).
- **what:** The evidence bundle always renders "## Original symptom" above "## Hypothesis under test," restoring visibility of the user's original question once the loop's working hypothesis has drifted from it. Fixes a finding that the verdict never saw the original symptom text after iteration 2 in benchmark run 1.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7; fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b
- **relations:** symptom-coverage gate family (DIAG-011); hypothesis drift (engine behaviour)
- **verify-later:** n/a — process/design record

### DIAG-025 — gamesdesign resolveResultSpec fixture (reference bug trajectory)
- **status:** superseded
- **status-evidence:** RUNBOOK(31) 2026-07-01 "STILL not resolveResultSpec — now for a substantive reason: reading real data, the model found a coherent cause... FORK for the user: (a) the fixture is stale... retire the 'reach resolveResultSpec' yardstick."
- **what:** The canonical eval scenario built from a real gamesdesign bug: seed "sections never reach save" → REFUTE on runtime evidence → REFUTE "token cap" → CONFIRM `resolveResultSpec` (a singular `output_field` collapsed the page to a stub). Used as both the scripted-verdict reference and the live-eval yardstick; superseded once the site's current state no longer exhibited the symptom (the loop instead correctly diagnosed a missing `site_specs.cta` aspect), and the route was closed on the refute-and-confirm-a-grounded-cause bar instead of this specific fixture.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#7.1,#6G-passed; docs019/RUNBOOK_gamesdesign_index_rebuild.md
- **relations:** falsification-first eval gate (DIAG-018); ground-truth eval set (register/contextkit-toolchain.md)
- **verify-later:** `/tmp` verdict scripts are ephemeral; groundtruth_targets.json resultspec entry

### DIAG-026 — Diagnose loop-back plumbing fault class (state threading + scope encoding)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) "§6E.1 DONE — state threading verified... trail_len == iteration (3) and stopped on a guard"; run 8d488e01 "there is a 'route' key, no 'diagnose_state' key"; loop_scope_field "CONFIRMED ISSUE + FIX... EncodeScope is json.Marshal of the Scope struct... coerces that OBJECT to empty."
- **what:** Two producer/consumer field-mismatch faults that left the loop running but silently degraded — both invisible-success bugs where the loop "worked" while its defining features were inert: (1) `diagnose_route` read its LoopState from bare `diagnose_state` while its own output lands under `route.diagnose_state`, so the loop re-seeded fresh every iteration, never enforcing max_iterations, truncating the evidence trail, and resetting the cross-iteration guards; (2) `route.scope` is `EncodeScope`'s untagged-struct JSON (`{"Symbols":[...]}`), so the dotted-path string-list reader needed `route.scope.Symbols` (capital S) — before the fix every re-scope silently fell through to the fallback chain and iterations 2+ never moved scope. Fix for (1) added a code self-check: if diagnose_route is about to seed but `route.diagnose_state` already exists, abort loudly (`stopped_by: state-threading-error`) rather than silently re-seed into a runaway. Emblematic of the wider dotted-lookup config contract class (also: analysis_field, result_from, repo_field, loop_scope_field).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position,#6C,#6E; NNN_fix_diagnose_route_state_threading(1).sql; NNN_fix_assemble_bundle_loop_scope_field.sql; docs019/DESIGN_diagnosis_loop_chassis_integration(6).md#status
- **relations:** convergence guards (DIAG-004); repo-label composition asymmetry (register/context-assembly.md — same config-contract class); workflow result-contract dead-key class
- **verify-later:** NNN_fix_diagnose_route_state_threading.sql; self-check branch in diagnose_route_action.go; Scope struct json tags; ExtractNestedField 3-level traversal

### DIAG-027 — diagnose_assemble_bundle action
- **status:** deployed
- **status-evidence:** RUNBOOK(31) checklist "§2 diagnose_assemble_bundle merged (gofmt-clean)" and "§6C build + register the four diagnose actions... DONE 2026-06-29."
- **what:** The chassis action that, per iteration, reads the in-scope symbols' bodies via `ReadSymbolBody` from a decoded `repo_analysis` Output, composing hypothesis + code + runtime (+ live schema) into the `bundle` the verdict step reads. Scope fallback chain: `route.scope` (loop-back) → `input_data.seed_scope` → `code_lookup.code_results`. Unknown symbols are logged and skipped, not fatal. Later gained the write-through to `diagnosis_artifacts` described in DIAG-028.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#2; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** ReadSymbolBody (register/contextkit-toolchain.md); loop_scope_field lesson (DIAG-026); diagnosis persistence (DIAG-028)
- **verify-later:** `platform/orchestration/actions/diagnose_assemble_bundle_action.go`

### DIAG-028 — Diagnosis persistence: diagnosis_artifacts write-through + persist_diagnosis_note
- **status:** deployed
- **status-evidence:** "F0.1b... work end to end in production" (fixloop NOTES(10)#Turn 6); Stage 3a CLOSED 2026-07-06 (skip gate proven ×3); Stage 3b CLOSED 2026-07-06/07 — first machine-written NOTES row `('pipeline','build')`, categories `["diagnosis","unconfirmed-diagnosis"]`.
- **what:** Two durability mechanisms layered onto the loop. (1) Each iteration's evidence bundle is persisted to `diagnosis_artifacts` (correlation_id, iteration, kind ∈ {bundle, iteration_note}, retention knob per kind) from inside `DiagnoseAssembleBundleAction`, immediately before its existing return — zero workflow-shape change; a persistence failure degrades to a logged warning on all paths, never failing the diagnosis, because observability must never cost a diagnosis. (2) After `diagnose_emit` (which stays read-only by design), a config-gated `persist_diagnosis_note` step writes the diagnosis as a NOTES entry ONLY when the run carries an explicit subject in input_data — skip, never guess, since a mis-filed note poisons history; the gate is the action's first check, before any DB access. UNVERIFIABLE verdicts are persisted too, tagged `unconfirmed-diagnosis`, so dead ends stop retries rather than repeating.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql#design note; fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1b; fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 3; RUNBOOK_travelling_docs(38).md#stage-3,#§4; 0NN_wire_persist_diagnosis_note.sql; RUNNING_NOTES_travelling_docs(39).md#rev3,#rev20,#rev21; docs019/HANDOFF_fixloop_thread(8).md#3,#4 (original F0 plan, aspirational when written)
- **relations:** subject threading (DIAG-029); anchorless-diagnosis degrade (DIAG-031); symptom-coverage gate (DIAG-011); diagnose_assemble_bundle (DIAG-027)
- **verify-later:** DiagnoseAssembleBundleAction source; ON CONFLICT clause used for the write; `persist_diagnosis_note_action.go` subject gate; diagnose-agent workflow `emit → persist_note → complete`

### DIAG-029 — Diagnosis subject threading through orchestrator input_mapping + both contracts
- **status:** deployed
- **status-evidence:** 3b.2 APPLIED + 3b.3 VERIFIED 2026-07-06 (map paths `input_data.subject_type`/`subject_key`; both contracts t/t); 129_wire_diagnosis_subject_threading.sql.
- **what:** For a spawned child to receive optional fields, the mapping must satisfy the callee's input_contract — so threading `subject_type?`/`subject_key?` took THREE edits (orchestrator input_mapping merge + `optional` additions on BOTH diagnose-orchestrator and diagnose-agent contracts), not two. DB-only, effective immediately. Establishes the general spawn+call contract rule: an input the workflow depends on must be declared on both ends.
- **sources:** RUNBOOK_travelling_docs(38).md#3b; RUNNING_NOTES_travelling_docs(39).md#rev17; HANDOFF_2026-07-08...md#§2; 129_wire_diagnosis_subject_threading.sql
- **relations:** persist_diagnosis_note (DIAG-028); spawn+call input-shape pattern; dangling-doc "declare your inputs" rule
- **verify-later:** diagnose-orchestrator `call_diagnoser.input_mapping`; both input_contracts

### DIAG-030 — Anchorless (code-only) diagnosis degrade at load_runtime + error_step lesson
- **status:** partial
- **status-evidence:** Corrective APPLIED 2026-07-06; fired ×5 per anchorless run ("NORMAL, not a fault"); softening (`skipped:true` return) still a chassis-build follow-up per both docs024 archive and travelling-docs accounts.
- **what:** Runtime evidence is meant to be an optional bundle tier, but `diagnose_load_runtime` hard-errored with no site/correlation/domain anchor and had no error routing — making the tier mandatory in practice and killing legitimate code-only diagnosis runs. Fixed by a config-level `error_step` on load_runtime targeting its own `next_step` (`assemble_bundle`); since `route.gather_step` re-enters load_runtime every iteration, each loop-back degrades per-iteration to a code+schema bundle. Cost of a full anchorless loop: approximately 26 minutes, 5 iterations. The general mechanism this exposed: the chassis workflow coordinator only consults `step.Config["error_step"]` (config-level) — a step-level `error_step` is parsed but never read, so placing it outside `config` is silently inert; dormant instances of the same buggy shape were found still live in the `tool-recreation-handler` and `tool-auditor` agent definitions. A proper code-level softening (treat no-anchor as a clean `skipped:true` skip) was identified but not yet shipped.
- **sources:** 016b_debugging_guide_7_3_(7).md#anchorless-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12,#rev14; 084_TRIGGER_diagnose_v1(2).sh (ANCHOR NOTE); archive_april_26/016b_debugging_guide_7(4).md#"error_step: config-level placement...","Anchorless (code-only) diagnosis..."; 127_diagnose_load_runtime_error_step.sql
- **relations:** chassis workflow step map (DIAG-002); dotted-path config contract class (DIAG-026)
- **verify-later:** `diagnose_load_runtime` no-anchor softening (shipped or not); load_runtime.config.error_step live value; grep `agent_definitions` for step-level `error_step` in tool-recreation-handler/tool-auditor

### DIAG-031 — Loop-worthiness test (fix-loop intake doctrine)
- **status:** deployed
- **status-evidence:** "LOOP-WORTHINESS TEST (doctrine — apply before every intake)" (RUNBOOK(10)#LOOP-WORTHINESS TEST).
- **what:** Five criteria applied before any bug enters the fix loop: it's a behaviour symptom, not a feature request; a causal mechanism plausibly exists across code/data/runtime; it is NOT answerable by one or two direct queries (mandatory cheap pre-check first); it is bounded to one symptom; the symptom is verified current at intake. Three successive candidates were dissolved by criterion 3 on this platform, leading to the empirical conclusion that "bug mechanisms tend to be legible to schema access plus grep" — reframing the workstream's value proposition from discovery to unattended/cited/consistent diagnosis.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#LOOP-WORTHINESS TEST; fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#LOOP-WORTHINESS TEST; fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §6
- **relations:** falsification-first eval gate (DIAG-018); known-answer benchmark methodology
- **verify-later:** n/a — methodology, not code

### DIAG-032 — Same-file sibling signatures + fair-share budgeting (F0.4c)
- **status:** deployed
- **status-evidence:** "F0.4c... CODE-COMPLETE" then "fair-share worked end to end" (NOTES(10)#Turn 8,#Turn 16).
- **what:** When retrieval scopes a symbol, the bundle also lists the signatures of that file's other functions (capped), fixing the case where symbol-granular retrieval found the right file but the wrong function. The initial implementation starved small files' budget with first-come-first-served ordering; fixed with fair-share-per-file budgeting (`capChars/n`, floor 600) plus a "+N more" affordance.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 8,#Turn 15,#Turn 16
- **relations:** static-tier corpus gaps (DIAG-033); must-claim-4 blind spot
- **verify-later:** grep/inspect `capChars/n`

### DIAG-033 — Static-tier corpus gaps: workflow-JSON invisibility + follow-the-error-log enrichment
- **status:** partial
- **status-evidence:** "code_symbols indexes .go files only. Workflow definitions live in agent_definitions.default_config as JSON and are therefore INVISIBLE to the loop's static tier" (RUNBOOK(10)#Inherited gotchas); "F0.4b... CODE-COMPLETE, its SQL verified live" (PLAN_fixloop_pilot.md §3b).
- **what:** The diagnosis loop's static evidence tier is built entirely from indexed Go source; workflow definitions stored as JSON in `agent_definitions.default_config` — which frequently contain the actual load-bearing control flow — are structurally invisible to it. Partially mitigated by the F0.4b enrichment: since load-bearing logic can live in that JSON, this regexes `agent/step (action)` references out of runtime evidence (agent_error_log lines) and inlines the named workflow step's JSON into the bundle, capped at 8KB — directly converting a benchmark bug's cause into cited static evidence. No general mechanism exists yet for the static tier to discover workflow-JSON logic it hasn't been pointed at.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas; fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7,#Turn 10
- **relations:** same-file sibling signatures (DIAG-032); corpus enrichment policy (register/context-engineering-principles.md); dartsonline guides defect
- **verify-later:** grep/inspect `code_symbols`; `.go`; `agent_definitions.default_config`

### DIAG-034 — Verdict-quality wrinkles + measured-dead code-retrieval channel
- **status:** partial
- **status-evidence:** README_flows: "seed similarities in the 0.55 band, no page-build symbols... a stale line carried into a conclusion that its own citation contradicts, and terminal-verdict data_requests that never run."; NOTES v4(39) "The code-retrieval channel contributes nothing (measured: flat similarity band 0.547–0.574 across all 12 seed hits; zero code citations in four full runs)."
- **what:** Post-run findings on the live loop, distinct from the earlier B4a retrieval-ceiling measurement: the lexical/semantic lookup channel contributes essentially nothing measurable across real diagnosis runs (work is on the query side — seed the lookup from runtime evidence, a self-contained lookup_symbols change, rather than better matching); the trigger's `site_id` is intermittent across runs (a reproducibility gap); and two verdict-quality defects point at the confirm/emit step — a conclusion contradicted by its own citation, and data_requests emitted on terminal verdicts that never execute.
- **sources:** docs019/README_flows.md; docs019/PLAN.md GATED; docs019/NOTES_running_synthesis_v4(39).md STATE OF THE WORLD
- **relations:** B4a retrieval ceiling (register/context-engineering-principles.md); data_requests channel (DIAG-010); eval gate (DIAG-018)
- **verify-later:** lookup_symbols seeding config; emit/confirm step handling of terminal data_requests

### DIAG-035 — Reasoning-state as a first-class handoff artefact
- **status:** partial
- **status-evidence:** thin_slice(27) improvement 5 "A bundle carries CODE + SCHEMA + RUNTIME EVIDENCE, but NOT reasoning state... The stopgap is a hand-written 'diagnosis so far' preamble."
- **what:** The insight that a context bundle without the evidence trail forces a fresh reader to re-derive already-falsified hypotheses; the design goal is a structured reasoning-state section accumulating across iterations (hypotheses tried, verdict + citation each, open discriminator). Partially realised by the loop's evidence trail in `collected_data`; the bundle-intrinsic version and per-iteration notes (F0.3 via doc_notes) remain in flight.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#next-improvements (item 5); docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F0.3)
- **relations:** diagnosis persistence (DIAG-028); falsification-first eval gate (DIAG-018)
- **verify-later:** evidence_trail shape in collected_data; doc_notes diagnosis category rows

### DIAG-036 — code_symbols retrieval as used by the diagnosis loop (index + lookup)
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6D "index populated (436 rows)"; code_retrieval_route(21) §7C "4,155 symbols; 499 distinct paths... min=max=commit 36710be"; 122_diagnose_agents.sql / 118_code_indexer_for_analyser.sql confirm the agent-side wiring.
- **what:** The retrieval corpus the diagnose-agent seeds iteration-1 scope from: `code_symbols` (repo, path, symbol, kind, signature, doc, content, embedding, commit_sha), written solely by the `code-indexer` agent and read by `lookup_code_symbols` (vector + trigram). UPSERT-safe via a unique identity constraint; prune removes rows whose commit_sha differs from the new index commit. This entry covers the loop's CONSUMPTION of the index; the index's own schema, the analyser adapter that feeds it, and the code-indexer/index-orchestrator agents are catalogued in register/context-assembly.md.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_code_retrieval_route(21).md#7A–7C; 122_diagnose_agents.sql; 118_code_indexer_for_analyser.sql
- **relations:** code_symbols table + hybrid retrieval (register/context-assembly.md); analyse_repo_local (DIAG-023); stale-corpus class (DIAG-037)
- **verify-later:** code_symbols table + constraints; code-indexer/index-orchestrator agent rows; index_code_symbols action

### DIAG-037 — Stale-corpus class: HEAD pinning, explicit refs, CI-triggered indexing doctrine
- **status:** partial
- **status-evidence:** code_retrieval_route(21) §7A "it was a faithful index of a YEAR-OLD tree... Decision: envelopes ALWAYS carry an explicit branch/sha"; queue item 4 "CI-triggered indexing: GitHub Actions step firing the index-orchestrator envelope with ${GITHUB_SHA} on push [queued]."
- **what:** A recurring staleness class: consumers pinned to `HEAD`/`latest` silently track an ancient artefact (remote HEAD unmoved from 2025; an agent image_tag of 'latest' resolving to a pre-architecture build). Adopted doctrine: explicit refs in every envelope, derived from the working checkout, never HEAD. Designed but not yet built: Structural A, a post-deploy CI step indexing at `${GITHUB_SHA}` so index commit == deployed commit by construction; Structural B, fast-forwarding main to the deployed sha. Rejected: resolving "most recently pushed branch" via API (latest-pushed ≠ deployed).
- **sources:** docs019/RUNBOOK_code_retrieval_route(9).md#ref-strategy; docs019/RUNBOOK_code_retrieval_route(21).md#7A; docs019/RUNBOOK_builder_route(21).md#queue (item 4)
- **relations:** code_symbols retrieval (DIAG-036); code-indexer/index-orchestrator + CI indexing (register/context-assembly.md); named-scope guard (DIAG-006, same corpus-size incident)
- **verify-later:** GitHub Actions workflow for post-deploy indexing (absent?); git ls-remote origin HEAD

### DIAG-038 — Index hygiene: exclude archived code copies, prune by commit
- **status:** deployed
- **status-evidence:** code_retrieval_route(21) §7C.1 "census: docs-archived 430 symbols / 50 files... interim DELETE run"; "reindex f284b749 VERIFIED (2026-07-03): commit e3176f8, 3,723 symbols, docs_rows=0."
- **what:** The repo stores archived copies of its own code under docs/ (and download-suffixed `name(N).go` files); indexing them pollutes retrieval with dead duplicates (nine duplicate assembler copies observed as ranks 1–9 for one query). Fixes: the analyser unconditionally skips `*(N).go`; `analyse_repo_local` gained `exclude_patterns` (Go default `["docs/"]`, via `AnalyseWithExclude`); prune semantics (`commit_sha IS DISTINCT FROM $new`) clear old-commit rows on the next reindex. Same trap documented CLI-side ("analyse the RIGHT ROOT", relative `-exclude` substrings).
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7C.1; docs019/RUNBOOK_thin_slice(27).md#known-limits,#B4a
- **relations:** stale-corpus class (DIAG-037); analyse_repo_local (DIAG-023); B4a eval discipline (register/context-engineering-principles.md)
- **verify-later:** exclude_patterns config on analyse_repo_local; code_symbols docs/% row count

### DIAG-039 — Evidence-fed fuzzy-scope resolver (§7D)
- **status:** deployed
- **status-evidence:** code_retrieval_route(21) route-close run 73ed55c6: "resolver canonicalisation (model's basenames → full paths @0.81–0.87) AND descriptive resolution... both load-bearing in the confirming scope"; "code WRITTEN 2026-07-02... resolver image is LIVE."
- **what:** Many verdict `next_scope` entries are English descriptions, not `path:Symbol` handles — previously inert (no call-graph match, no body sliced). The resolver, inside `diagnose_route` after verdict-parse and before Advance, embeds each non-exact entry (same nomic client/prefixes as retrieval) and vector-searches code_symbols, replacing it with the top hits (`resolver_top_k` default 2, tuned so substitution stays inside the narrowing guard's +2 allowance; min similarity 0.55; unresolvable entries survive as plain labels — "no worse"). The trail records the RESOLVED scope, the more auditable form. Reuses the seed lookup's retrieval machinery wholesale, retiring the earlier §7F seed-reorder design once relevance was proven by content twice.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D,#route-closed
- **relations:** call-graph re-scope mechanism (DIAG-007); named-scope guard (DIAG-006); §7F seed reorder (superseded, folded into this entry's history)
- **verify-later:** diagnose_route_action.go resolver step 3.5; diagnose_route_resolver_test.go

### DIAG-040 — Base-runbook gated-items framing (documentation-style lineage note)
- **status:** superseded
- **status-evidence:** RUNBOOK.md §6 "Gated items (carried — see PLAN.md)... None are unblocked by this thread's work alone" — replaced from RUNBOOK(1) onward by an inlined "§6 Completing the whole task (what remains)" with per-step definitions of done.
- **what:** The earliest form of the diagnosis-loop runbook kept only in-flight build steps and deferred the roadmap to a separate PLAN.md; within one version the roadmap was inlined as §6 with per-step DoD and live status, and the runbook became the single self-contained thread-state document (later §7 split to its own file once §6 closed). Recorded here as a family-delta of the project's documentation style converging on self-contained travelling runbooks; it is a documentation-system convention, not a diagnosis-loop mechanism, and is filed here only because its evidence lives entirely inside the diagnosis-loop runbook family.
- **sources:** docs019/RUNBOOK.md#6; docs019/RUNBOOK(1).md#6; docs019/RUNBOOK(31)_diagnosis_loop.md#6 (ACTIVE ROUTE MOVED banner)
- **relations:** category mismatch note — properly a documentation-system convention; parallel-thread convention
- **verify-later:** PLAN.md in docs019 (sibling file, other unit)

### DIAG-041 — backend_unreachable discovery check
- **status:** partial
- **status-evidence:** running_notes 2026-06-13(f) "backend_unreachable REWRITTEN against the real DiscoveryCheck interface... Run(dctx DiscoveryCheckContext)(*CheckResult,error)... gofmt-clean"; enable pending.
- **what:** A `discovery_checks` check that no-ops unless `deploy_config.target='vm'`, GETs each backend site's public `/health`, and on failure emits a `site_work_items` row (source='discovery', item_type='backend_unreachable', item_key for dedup). Self-clearing; alert-only (HandlerAgent "") because a down VM isn't chassis-fixable — a future P5 vmhost adapter becomes the handler. Category mismatch note: this is a site-monitoring/discovery-checks concept, not diagnosis-loop machinery — filed here per the consolidation instructions because it was tagged diagnosis-loop at extraction and no closer-fitting category exists in this cluster's assignment.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e,#2026-06-13-f; traffic_probe_plan(11).md#p4
- **relations:** category mismatch — belongs with site-monitoring/discovery_checks, not diagnose-agent; P5 vmhost adapter (future handler)
- **verify-later:** discovery_checks/check_backend_unreachable.go; site_work_items idx_swi_dedup
