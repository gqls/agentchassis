# Cluster: diagnosis-and-context
Categories included: diagnosis-loop, new:context-assembly, new:contextkit-toolchain, new:context-pack-tooling, new:context-engineering-principles


<!-- SOURCE: U03_idea_uk_section_data.md -->
### cmd/bundle read-only context composer
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Used for bundles 1–3 and the differentiators handoff; notes (Sa) record its failure modes ("-step framing emitting structure rather than source — use -step debug"; -doc paths must be ls-verified).
- **what:** The investigation tooling that assembled evidence bundles for this thread: `go run ./cmd/bundle` with an analysis JSON, `-root`, `-constitution`, `-step debug|framing|implementation`, `-task` (one-sentence brief), `-scope file[:Symbol]` code selections, `-include`, `-doc` paths, `-psql` connection command, `-schema-tables`, `-runtime-site`/`-runtime-page` live evidence, `-out`. Operational lore: `-step framing` yields signatures only; doc paths silently fail if wrong; bundles can arrive as thin slices (runtime data excluded) so live queries still need running separately.
- **sources:** 001_bundling_context.md; bundle3; RUNBOOK_scheme_to_components(50).md#Bundle-command; running_notes_scheme_to_components(55).md#Sa #Sh
- **relations:** docs019 contextkit (its home); check-based investigation method.
- **verify-later:** cmd/bundle source under docs019 go_files/contextkit; flag semantics.

<!-- SOURCE: U05_content_quality_linking.md -->
### Context packaging + code-bundle tooling for fresh chats
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §11 gives working bundle invocations + "Known cmd/bundle run errors (to fix before regenerating)".
- **what:** Two generations of context assembly for hand-off to a fresh assistant chat: (1) the module packager shell script (package_content_quality_debug.sh) bundling a code slice + docs + an optional read-only live SQL capture into one context file; (2) contextkit `cmd/bundle`, which expands from named symbol scopes via the call graph with -doc/-schema-tables/-runtime flags and a constitution. The 001_context files record filled-in bundle invocations per defect thread (phantom-CTA, clobber). Known operational failures documented: unquoted parentheses in -doc paths break bash; empty /tmp/analysis_repo.json breaks analysis load; bundle can't reach session docs outside the repo.
- **sources:** package_module/package_content_quality_debug(3).sh (header); game_lost_its_tool/001_context; phantom_hero_ctas/001_context; HANDOFF_page_pipeline(11).md#11
- **relations:** docs019 contextkit; division-of-labour operating model; documentation system.
- **verify-later:** contextkit cmd/bundle; contextkit_bundle_issues.md.

<!-- SOURCE: U06_finetuning.md -->
### Docubundle context packager (thunder-checkpoint-race package)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** docubundle/README.md usage note + the generated 633KB production context file dated in-tree; packager header "Patterned on package_page_build_debug.sh".
- **what:** A self-contained packager script (`package_thunder_checkpoint_race.sh`) that bundles the async-await + loop machinery of the chassis, the checkpoint-upload path, the working docs, and optionally a read-only live capture (schemas, decisive queries, workflows, runtime state) into one context file to seed a fresh AI-assistant thread on a specific blocker. Paired with hand-written CONTEXT_PACK / NEXT_CHAT_MANIFEST docs that state the blocker, the verified root cause, the applied fix, and next actions. An instance of the wider bundle/context-package pattern (cf. docs019 contextkit) applied to the finetuning workstream; the targeted CHASSIS_await_loop_extract ("use the targeted extract, not the 72k-line file") shows deliberate context-size curation.
- **sources:** working/docubundle/README.md; working/docubundle/package_thunder_checkpoint_race.sh (header); working/phase5/NEXT_CHAT_MANIFEST.md; working/phase5/CHASSIS_await_loop_extract.txt (header)
- **relations:** diagnosis-loop bundles/contextkit; send-before-register race (its subject)
- **verify-later:** relation to z_bundles/context_packages tooling at repo root

<!-- SOURCE: U08_travelling_docs.md -->
### persist_diagnosis_note — skip-don't-guess subject gate; dead ends persisted
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Stage 3a CLOSED 2026-07-06 (skip gate proven ×3); Stage 3b CLOSED 2026-07-06/07 — first machine-written NOTES row `('pipeline','build')`, categories `["diagnosis","unconfirmed-diagnosis"]`, stop reason `scope-not-narrowing`.
- **what:** A config-gated step after `diagnose_emit` (emit stays read-only by its own design) that persists the diagnosis as a NOTES entry ONLY when the run carries an explicit subject in input_data — skip, never guess (a mis-filed note poisons history; the gate is the action's first check, before any DB access). UNVERIFIABLE verdicts are persisted too, tagged `unconfirmed-diagnosis`, so dead ends stop retries. First payoff on record: the machine-written note itself answered the open "why did the run finish fast" question.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-3,#§4; 0NN_wire_persist_diagnosis_note.sql; RUNNING_NOTES_travelling_docs(39).md#rev3,#rev20,#rev21
- **relations:** subject threading (3b); anchorless-diagnosis degrade; 037 pipeline-integration vision (realised).
- **verify-later:** diagnose-agent workflow `emit → persist_note → complete`; `persist_diagnosis_note_action.go` subject gate.

<!-- SOURCE: U08_travelling_docs.md -->
### Diagnosis subject threading through orchestrator input_mapping + both contracts
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** 3b.2 APPLIED + 3b.3 VERIFIED 2026-07-06 (map paths `input_data.subject_type`/`subject_key`; both contracts t/t).
- **what:** For a spawned child to receive optional fields, the mapping must satisfy the callee's input_contract — so threading `subject_type?`/`subject_key?` took THREE edits (orchestrator input_mapping merge + `optional` additions on BOTH diagnose-orchestrator and diagnose-agent contracts), not two. DB-only, effective immediately. Establishes the general spawn+call contract rule: an input the workflow depends on must be declared.
- **sources:** RUNBOOK_travelling_docs(38).md#3b; RUNNING_NOTES_travelling_docs(39).md#rev17; HANDOFF_2026-07-08…md#§2
- **relations:** spawn+call input-shape pattern (016b); dangling-doc rule (same "declare your inputs" class — migration 137's `spec` declaration).
- **verify-later:** diagnose-orchestrator `call_diagnoser.input_mapping`; both input_contracts.

<!-- SOURCE: U08_travelling_docs.md -->
### Anchorless (code-only) diagnosis degrade at load_runtime
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Corrective APPLIED 2026-07-06; fired ×5 per anchorless run ("NORMAL, not a fault"); softening (`skipped:true` return) still a chassis-build follow-up.
- **what:** Runtime evidence is an optional bundle tier, but `diagnose_load_runtime` hard-errored with no site/correlation/domain anchor and had no error routing — making the tier mandatory in practice and killing legitimate code-only diagnosis runs. Fixed by config-level error_step on load_runtime targeting its own next_step (`assemble_bundle`); since `route.gather_step` re-enters load_runtime every iteration, each loop-back degrades per-iteration to a code+schema bundle. Cost of a full anchorless loop: ≈26 min, 5 iterations.
- **sources:** 016b_debugging_guide_7_3_(7).md#anchorless-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12,#rev14; 084_TRIGGER_diagnose_v1(2).sh (ANCHOR NOTE)
- **relations:** error_step mechanics; diagnosis loop step map (analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict → route → emit → persist_note → complete).
- **verify-later:** `diagnose_load_runtime` no-anchor softening (shipped or not); load_runtime.config.error_step live value.

<!-- SOURCE: U08_travelling_docs.md -->
### Verdict symptom-coverage gate (symptom_check) on the diagnose-agent
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** FYI 2026-07-10: prompt rule 8 + `symptom_check` schema field applied (snapshot 34f4afc8); engine coercion rides the next chassis image post-v1.0.1101; F0.6 addendum adds `cites`/`context` members.
- **what:** A CONFIRMED verdict must account for every distinct observation of the ORIGINAL symptom via `symptom_check: [{observation, explained, how, cites, context}]`; the chassis engine (`pkg/diagnose`) coerces to UNVERIFIABLE any CONFIRMED verdict whose symptom_check is missing, carries an unexplained entry, or marks explained without a valid citation index; comparative/background clauses are exempted as `context` rather than grade-inflated. Terminal diagnosis notes gain a "Symptom coverage:" block. Owned by the fix-loop workstream; delivered to this unit as a courtesy collision-rule FYI.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md (whole)
- **relations:** persist_diagnosis_note (note bodies change); fix-loop council/verdict work (fixloop_eg_dartsonline docs).
- **verify-later:** diagnose-agent verdict prompt_template; `verdict_wire.go` symptom_check parsing.

<!-- SOURCE: U08_travelling_docs.md -->
### Context-bundle command for cross-chat handoffs (cmd/bundle + registry-based scope resolution)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Two bundles built and used (07-08 toolgen bug; 07-09 recreation); resolver rewritten after 3 misses (rev 44).
- **what:** `cmd/bundle` renders a task bundle (constitution + task text + code scopes + docs + live schema + runtime evidence incl. an agent_error_log "Recent errors" section — the section that settled the 07-08 diagnosis). Path facts banked: resolve actions via registry.go (action name → constructor → defining file), not filename convention (`execute_llm_prompt` lives in ai_actions.go; validate_page_content.go lacks the _action suffix); misses are non-fatal and print grep candidates.
- **sources:** bundle_recreation_v1(1).sh (header + resolve_action); HANDOFF_2026-07-08…md#§6; RUNNING_NOTES_travelling_docs(39).md#rev44
- **relations:** docs019 contextkit/bundles; agent_error_log first read.
- **verify-later:** cmd/bundle flags; whether the runtime errors section is standard.

<!-- SOURCE: U09_adoption.md -->
### Docubundle context-pack tooling (module packager + dbcontext + deploy guide)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Working scripts and generated 1.6MB context files present (package_page_build_debug.sh + output_contexts); GUIDE_deploy_from_context_packs documents the operating loop across four live packs.
- **what:** Tooling for packaging a subsystem's blast-radius into one AI-consumable context file: `package_page_build_debug.sh` (self-contained packager bundling the page-build/section-resolution/render-deploy/dispatch code, keeping the reuse-discovery catalogue layers — registry.go, datahelpers, input_contracts — plus an optional read-only live capture: schema, decisive queries, agent-def workflows, runtime state); `dbcontext` Go CLI (shells out to configurable psql; `\d` schema dumps and multipass-sized row fetches — never an unbounded dump); and the GUIDE's general loop (attach pack → gather live context → verify the decisive fact → work → deploy via mechanism A–F → verify with positive evidence). Deploy mechanisms taxonomy: A chassis image, B database/migrations, C work-items, D orchestrate-message triggers, E generated static sites (git→Actions→B2), F idea.uk binary.
- **sources:** docubundle/GUIDE_deploy_from_context_packs.md, docubundle/dbcontext.go header, docubundle/package_module/package_page_build_debug.sh header, CONTEXT_PACK_adoption_skinner_box.md
- **relations:** context packs (CONTEXT_PACK_* docs); docs019 bundles/contextkit; thin-slice constitution (included in every bundle)
- **verify-later:** whether packagers/dbcontext live in the repo proper or only in docs; output_contexts freshness

<!-- SOURCE: U09_adoption.md -->
### Adoption context pack (skinner-box) as a worked fresh-thread starter
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** CONTEXT_PACK_adoption_skinner_box.md exists and was consumed by the 2026-06-08 session that closed the bug (running_notes_15 Part 1 resolves the pack's named "decisive fork").
- **what:** A structured resume pack for one open bug: state + next action, the decisive fork to verify first, standing rules (constitution excerpt), code to pull fresh vs re-attach, schema/rows/runtime capture commands, and the minimum fast-start set. Demonstrates the pack contract: packs restate earlier context and inherit its staleness — the fresh pull is the source of truth (the pack's own causal story about the content-writer was corrected by the session).
- **sources:** CONTEXT_PACK_adoption_skinner_box.md, NEXT_CHAT_INPUTS_2026-06-06.md, running_notes_15(10)#part-1
- **relations:** docubundle tooling; GUIDE_deploy_from_context_packs per-project quick reference
- **verify-later:** n/a (artifact of method)

<!-- SOURCE: U12_docs024_archives.md -->
### `error_step`: config-level placement requirement + derive-from-next_step fix pattern
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the routing FIRES (error_routed in ProcessingHistory)... Live-validated ×5 in one run."
- **what:** The chassis workflow coordinator only consults `step.Config["error_step"]` (config-level); a step-level `error_step` is parsed but never read, so placing it outside `config` is silently inert. Fix pattern: derive `error_step` from the step's own `next_step`. This entry and its three siblings below are genuinely absent from the canonical live `016b_debugging_guide_8_consolidated.md`/`merged(1).md` — they continue only in a parallel `travelling_docs/016b_debugging_guide_7_3_(2..7).md` fork the canonical consolidation's "verified against ALL forks" claim did not actually reconcile.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"error_step: config-level placement..."
- **relations:** dormant instances of the buggy shape found still live in `tool-recreation-handler` and `tool-auditor` agent definitions
- **verify-later:** grep `agent_definitions` for step-level `error_step` occurrences in those two agents.

<!-- SOURCE: U12_docs024_archives.md -->
### Anchorless (code-only) diagnosis dies at load_runtime
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "Fix. Config-level error_step on load_runtime... Live-validated ×5" (deployed) but "Pending softening (next chassis build)" (aspirational remainder).
- **what:** A diagnosis run with no anchor was treated as optional by bundle-assembly but hard-errored the whole child workflow at `load_runtime`. Interim fix routes the error back to its own `next_step`; a proper code-level softening (treat no-anchor as a skip) was identified but not yet shipped.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Anchorless (code-only) diagnosis..."
- **relations:** sibling of the error_step concept above; also absent from canonical live 016b
- **verify-later:** check `diagnose_load_runtime` action source for the `skipped:true` softening.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Bundle write-through (DiagnoseAssembleBundleAction, F0.1b)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.1b... work end to end in production" (NOTES(10)#Turn 6, F0 plumbing criteria)
- **what:** Each diagnosis iteration's evidence bundle is persisted to `diagnosis_artifacts` from inside the Go action `DiagnoseAssembleBundleAction`, immediately before its existing return — zero workflow-shape change, staying off the tools-chat's active `emit → persist_note → complete` surface. A persistence failure degrades to a logged warning on all paths; it never fails the diagnosis itself, because observability must never cost a diagnosis.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql#design note, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1b, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 3
- **relations:** diagnosis_artifacts table; retention knob
- **verify-later:** DiagnoseAssembleBundleAction source; ON CONFLICT clause used for the write

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Loop-worthiness test (doctrine)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "LOOP-WORTHINESS TEST (doctrine — apply before every intake)" (RUNBOOK(10)#LOOP-WORTHINESS TEST)
- **what:** Five criteria applied before any bug enters the loop: it's a behaviour symptom not a feature request; a causal mechanism plausibly exists across code/data/runtime; it is NOT answerable by one or two direct queries (mandatory cheap pre-check first); it is bounded to one symptom; the symptom is verified current at intake. Three successive candidates were dissolved by criterion 3 on this platform, leading to the empirical conclusion that "bug mechanisms tend to be legible to schema access plus grep" — reframing the workstream's value proposition from discovery to unattended/cited/consistent diagnosis.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#LOOP-WORTHINESS TEST, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#LOOP-WORTHINESS TEST, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §6
- **relations:** known-answer benchmark methodology; abandoned pilot candidates
- **verify-later:** n/a — methodology, not code

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Symptom anchor (F0.4a)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4a... ✅ CODE-COMPLETE 2026-07-09" then verified live in run 2 (PLAN_fixloop_pilot.md §3b, NOTES(10)#Turn 10)
- **what:** The evidence bundle always renders "## Original symptom" above "## Hypothesis under test," restoring visibility of the user's original question once the loop's working hypothesis has drifted from it. Fixes a finding that the verdict never saw the original symptom text after iteration 2 in benchmark run 1.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b
- **relations:** hypothesis drift (engine behaviour); symptom-closure gate (F0.4d)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Follow-the-error-log enrichment (F0.4b)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4b... ✅ CODE-COMPLETE, its SQL verified live" (PLAN_fixloop_pilot.md §3b)
- **what:** Bridges the loop's Go-only static-evidence corpus gap: since `code_symbols` indexes `.go` files only and load-bearing platform logic can live in `agent_definitions.default_config` JSON, this enrichment regexes `agent/step (action)` references out of runtime evidence (agent_error_log lines) and inlines the named workflow step's JSON into the bundle, capped at 8KB. Directly converted the benchmark bug's cause B into cited static evidence.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7, #Turn 10
- **relations:** symptom anchor; workflow-JSON-as-load-bearing-logic gotcha
- **verify-later:** grep/inspect `code_symbols`; `.go`; `agent_definitions.default_config`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Same-file sibling signatures + fair-share budgeting (F0.4c)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4c... ✅ CODE-COMPLETE" then "fair-share worked end to end" (NOTES(10)#Turn 8, #Turn 16)
- **what:** When retrieval scopes a symbol, the bundle also lists the signatures of that file's other functions (capped), fixing the case where symbol-granular retrieval found the right file but the wrong function. Initial implementation starved small files' budget with first-come-first-served ordering; fixed with fair-share-per-file budgeting (`capChars/n`, floor 600) plus a "+N more" affordance.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 8, #Turn 15, #Turn 16
- **relations:** follow-the-error-log enrichment; must-claim-4 blind spot
- **verify-later:** grep/inspect `capChars/n`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tier-coverage guard (F0.4e)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "First production firing of any of the new guards" (NOTES(10)#Turn 12, run 3)
- **what:** A shared `coerceVerdict()` engine gate requiring a CONFIRMED verdict to carry at least one `static` citation AND at least one `state|runtime` citation, or it degrades to Unverifiable. REFUTED is exempt. Directly answers the benchmark run-1 finding that "cite-or-abstain does not prevent confirming the wrong cause."
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7, #Turn 8, #Turn 12
- **relations:** symptom-closure gate; context disposition (F0.6); cite-or-abstain doctrine
- **verify-later:** grep/inspect `coerceVerdict()`; `static`; `state|runtime`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Symptom-closure gate / symptom_check (F0.4d)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4d — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md §3b)
- **what:** A CONFIRMED verdict must carry a `symptom_check` — mapping each observation in the original symptom to the confirmed mechanism (`explained:true/false` + `how`) — or the engine coerces it to Unverifiable. Motivated by benchmark run 2, where a well-cited confirm dismissed half the symptom as "not a nav issue." The verdict prompt lives in the diagnose-agent workflow JSON, a different workstream's active surface, so the edit was done fetch-first with a snapshot and an FYI filed.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 11, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#BENCHMARK RUN 3
- **relations:** tier-coverage guard; context disposition (F0.6); doc_notes/travelling-docs coordination boundary
- **verify-later:** grep/inspect `symptom_check`; `explained:true/false`; `how`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Context disposition + citation-backed "explained" (F0.6)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.6 — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md §3b)
- **what:** Refines symptom_check with a `context bool` flag (comparative/background clauses exempt from explained/unexplained accounting) and requires `explained:true` entries to carry an in-range `cites` index — an unsupported "explained" is now rejected. Fixes a grade-inflation defect where run 4 marked comparison clauses `explained:true` while their own text said "unverifiable from this bundle."
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 14, #Turn 15
- **relations:** symptom-closure gate; tier-coverage guard
- **verify-later:** grep/inspect `context bool`; `explained:true`; `cites`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### data_request persistence across iterations (F0.5)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.5 — ✅ CODE-COMPLETE 2026-07-10 (from run 3)" (PLAN_fixloop_pilot.md §3b)
- **what:** Fixes a defect where fetched data_request answers evaporated from the bundle after one iteration, tripping the scope-not-narrowing guard. Reuses `LoopState.SeenRequests` by forwarding the UNION of current-verdict and prior-seen request keys (deduped, capped at 12) so `load_runtime` re-runs them every iteration — "re-run, don't store," avoiding the collected_data-bloat class of incident.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 12, #Turn 13
- **relations:** tier-coverage guard; collected_data-bloat gotcha
- **verify-later:** grep/inspect `LoopState.SeenRequests`; `load_runtime`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### contextkit bundle regeneration procedure
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "verified end-to-end 2026-07-09... 306,897 B, 468 files analysed, all three gathers succeeded" (RUNBOOK(10)#REGENERATING THE CONTEXT BUNDLE)
- **what:** The documented, tested procedure for regenerating a human/chat-facing evidence bundle via the `contextkit` CLI (a separate Go module, not the live loop's in-cluster assembler): analyser with excludes, then bundle with `-psql` as ONE quoted argument and `-schema-tables` including the tables relevant to the bug. This bundle is for humans; the live loop's own retrieval is a separate, in-process mechanism.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#REGENERATING THE CONTEXT BUNDLE, fixloop_eg_dartsonline/HANDOFF_fixloop_thread(8).md#CODE CONTEXT
- **relations:** blinding discipline
- **verify-later:** grep/inspect `contextkit`; `-psql`; `-schema-tables`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Corpus gap: workflow JSON invisible to the static tier
- **category:** diagnosis-loop
- **status-signal:** deployed (documented gotcha, partially mitigated by F0.4b)
- **status-evidence:** "code_symbols indexes .go files only. Workflow definitions live in agent_definitions.default_config as JSON and are therefore INVISIBLE to the loop's static tier" (RUNBOOK(10)#Inherited gotchas)
- **what:** The diagnosis loop's static evidence tier is built entirely from indexed Go source; workflow definitions stored as JSON in `agent_definitions.default_config` — which frequently contain the actual load-bearing control flow — are structurally invisible to it. Partially mitigated by the follow-the-error-log enrichment (F0.4b); no general mechanism exists for the static tier to discover workflow-JSON logic it hasn't been pointed at.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7
- **relations:** follow-the-error-log enrichment (F0.4b); dartsonline guides defect
- **verify-later:** grep/inspect `agent_definitions.default_config`

<!-- SOURCE: U14_docs019_runbooks.md -->
### contextkit — task-scoped codebase bundle toolkit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Done: call-graph neighbourhood; live schema + row data via dbcontext …; the cmd/bundle orchestration wrapper"; RUNBOOK(31) header "This project builds contextkit … developed against, and dogfooded on, the agent-chassis repository" (2026-06-24).
- **what:** A small Go module (`contextkit/`: cmd/analyser, assembler, embed, dbcontext, resolve_targets, fuse, eval_targets, bundle, diagnose) that assembles a tightly-scoped slice of a codebase — the in-scope source in full, its call-graph neighbourhood as signatures, DB schema, runtime evidence, and authored guidance/constitution — into one paste-ready "bundle" per task. Two shared contracts (`internal/analysis`, `internal/candidates`) defined once, no per-tool copies. The deployed chassis diagnosis agent is its descendant; the CLI remains the dev/eval harness.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#the-pipeline; docs019/RUNBOOK(31)_diagnosis_loop.md#what-this-is; docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** read-only diagnosis loop; bundle altitudes; dbcontext; cmd/bundle wrapper
- **verify-later:** `docs019/go_files/contextkit/` module; `$CK/cmd/*`; `internal/analysis`, `internal/candidates`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Read-only cite-or-abstain diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) checklist "§6G eval gate PASSED — run 51f95cda (2026-07-01): abstain → correct reads → REFUTE the naive framing → CONFIRM the grounded cause"; code_retrieval_route(21) "§7 ROUTE CLOSED — 2026-07-03 (run 73ed55c6)".
- **what:** An AI agent that investigates a bug strictly READ-ONLY: forms a hypothesis, gathers scoped evidence (code bodies + read-only DB rows + runtime records), issues a verdict that must CITE evidence or ABSTAIN (CONFIRMED/REFUTED/UNVERIFIABLE), then re-scopes by FOLLOWING the evidence (call graph for code, vetted queries for data) rather than re-searching the symptom. Never edits code, never runs builds, human-gated; the hard problem it targets is falsification — abandoning a wrong hypothesis.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#what-this-is; docs019/RUNBOOK_design_diagnosis_loop(7).md#overview; docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** convergence guards; verdict wire format; three-tier citation; falsification-first eval gate; diagnosis→fix loop (v2)
- **verify-later:** chassis `pkg/diagnose/` (loop.go, step.go, advance.go); `platform/orchestration/actions/diagnose_*_action.go`; agent_definitions rows diagnose-agent/diagnose-orchestrator

<!-- SOURCE: U14_docs019_runbooks.md -->
### Bundle step altitudes: framing vs implementation vs debug
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) assembler flags: "-step framing | implementation | debug … framing: in-scope shown as signatures (intent over detail)".
- **what:** A bundle declares its altitude: `framing` shows in-scope code as signatures only (used to expand an under-specified brief into a spec before targets can be picked), `implementation`/`debug` show full bodies, and `debug` adds a runtime-evidence section. Encodes the framing-vs-implementation altitude split as an explicit pipeline parameter.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#assembler-flags; docs019/RUNBOOK_thin_slice(27).md#fuzzy-tasks
- **relations:** contextkit toolkit; reasoning-state handoff
- **verify-later:** `$CK/cmd/assembler/main.go` step handling

<!-- SOURCE: U14_docs019_runbooks.md -->
### Call-graph neighbourhood selection with forced -include for wiring files
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "-include … Closes the blind spot the first adoption run found"; known-limits section "call-graph neighbourhood is name-matched, not type-resolved".
- **what:** The bundle's surrounding context is the call-graph neighbourhood (callees/callers/types) of the in-scope symbols, rendered as signatures, with `-neighbour package` as fallback when name-matching misses (interface dispatch). Registration/wiring files (e.g. registry.go, reached via init not calls) are force-included with `-include`. Ubiquitous names (Run, String, New) are dropped when the loop follows the graph, to avoid scope explosion.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#assembler-flags; docs019/RUNBOOK_design_diagnosis_loop(7).md#design-and-build-choices
- **relations:** named-scope guard vs capped expansion; ReadSymbolBody slicer
- **verify-later:** `internal/analysis/analyse.go` calls extraction; `pkg/diagnose/callgraph.go` ubiquitous-name drop list

<!-- SOURCE: U14_docs019_runbooks.md -->
### dbcontext — bounded read-only DB context gather
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) pipeline step 2 with worked flags; "-rows … multipass sizing (probe LIMIT N+1 …). Never an unbounded dump."
- **what:** CLI that pulls live DB context for a bundle: `-schema` (`\d` per table), `-rows` (SELECT with multipass sizing and a row cap), and `-runtime-site`/`-runtime-page` (recent agent_error_log rows + site_work_items lifecycle as a "Runtime evidence" block). All read-only; queries are appended as `-c` args, not shell-interpolated.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#dbcontext-flags
- **relations:** cmd/bundle wrapper; three-guard read-only SQL model; diagnose_load_runtime
- **verify-later:** `$CK/cmd/dbcontext/`

<!-- SOURCE: U14_docs019_runbooks.md -->
### cmd/bundle orchestration wrapper and the pure-composer boundary
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) design note "(Status: wrapper not yet built — flagged for decision)" superseded in the same file's Done list "the cmd/bundle orchestration wrapper (gather via dbcontext → assemble, composer stays read-only)".
- **what:** The assembler is a PURE COMPOSER — it never runs SQL or chooses tables; `cmd/bundle` is the orchestration wrapper that runs the requested read-only dbcontext gathers and then calls the assembler with the outputs wired in. Keeps query execution inside the bounded read-only tool while offering "one command including the SQL". Automatic table-selection was deliberately deferred and must propose-then-confirm.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#one-command; docs019/RUNBOOK_thin_slice(27).md#assembler-boundary
- **relations:** dbcontext; diagnosis loop gatherer (BundleGatherer shells out to cmd/bundle)
- **verify-later:** `$CK/cmd/bundle/`; `pkg/diagnose/gatherer.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Bundle size doctrine — "a large bundle is a smell, not a goal"
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Context-window facts (verified against the Claude docs, June 2026)"; "aim to keep a working bundle under ~200K tokens (~800 KB)".
- **what:** Working rule for feeding bundles to models: keep under ~200K tokens; context rot means a full 1M window is not used evenly; the fix for an oversized bundle is narrower selection, not a bigger window. Includes the three feeding routes (chat paste, claude.ai Project, API with prompt caching of the stable prefix).
- **sources:** docs019/RUNBOOK_thin_slice(27).md#large-bundles
- **relations:** responses-are-summaries doctrine (Kafka side); call-graph neighbourhood (the narrowing instrument)
- **verify-later:** n/a (doctrine); bundle sizes in diagnosis_artifacts once built

<!-- SOURCE: U14_docs019_runbooks.md -->
### B4a finding — the symptom→infrastructure retrieval ceiling
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "OUTCOME (2026-06-17, 2 ground-truth tasks): skinner-box lexical 0.50, semantic 0.00; resultspec lexical 0.00, semantic 0.00, fused 0.00 … DECISION: embeddings do NOT earn a place in the code path on this evidence".
- **what:** Measured finding that when a bug's cause lives in shared infrastructure named for its FUNCTION rather than its FAILURE MODE, symptom-based code retrieval (lexical, semantic, or fused) has a hard ceiling — symptom words and mechanism words don't intersect, and no embedding closes a zero-overlap gap. Secondary finding: naive RRF fusion can be worse than lexical alone. This is the empirical justification for the diagnosis loop's re-scope-by-following-evidence design.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#B4a; docs019/RUNBOOK_design_diagnosis_loop(7).md#the-empirical-finding
- **relations:** lexical/semantic/fused target resolution; read-only diagnosis loop (the lever pulled instead)
- **verify-later:** `$CK/groundtruth_targets.json`; `docs019/go_files/contextkit/{lex,sem}.json`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Lexical/semantic/fused target resolution (resolve_targets, embed, fuse, eval_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) pipeline 1b–1d with all four tools runnable; B4a decision "lexical (trigram + resolve_targets) carries the spine and embeddings are the tie-breaker".
- **what:** The target-resolution layer: a lexical (trigram) candidate proposer, an Ollama-backed semantic index (nomic-embed-text with search_document/search_query prefixes matching the chassis rag pipeline exactly), RRF rank fusion, and a recall@N/MRR scorer against a ground-truth task set. Built to answer "does semantic beat lexical for code" — the measured answer was no for this corpus.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#the-pipeline; docs019/RUNBOOK_thin_slice(27).md#B4a-task-1
- **relations:** B4a ceiling finding; code_symbols index (production analogue); evidence-fed scope resolver (later reuse of the same vector search)
- **verify-later:** `$CK/cmd/{resolve_targets,embed,fuse,eval_targets}/`; ollama-adapter service

<!-- SOURCE: U14_docs019_runbooks.md -->
### Ground-truth eval harness and its measurement-trap discipline
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "THE TRAP (hit 2026-06-14): resolve_targets was run with a DIFFERENT task … eval then scored … a meaningless 0/2"; the task-string bind guard and `-task-id` requirement.
- **what:** groundtruth_targets.json holds task→expected-symbol pairs; every eval binds the task string once, guards it against the truth file, uses ONE matched index for lexical and semantic, and forbids answer-vocabulary leaks in task wording (a leaked symbol name contaminated the ceiling test once). Three prior B4a attempts failed on METHOD, not result — the harness encodes the corrections.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#B4a-task-1; docs019/RUNBOOK_thin_slice(27).md#B4a-task-2
- **relations:** instrument-skepticism doctrine; B4a ceiling finding
- **verify-later:** `$CK/groundtruth_targets.json`; `$CK/cmd/eval_targets/`

<!-- SOURCE: U14_docs019_runbooks.md -->
### ReadSymbolBody — the single shared symbol-body slicer
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §3 "This collapse is DONE and verified: the merged assembler … diffed against the pre-collapse binary … byte-identical"; checklist "§1 ReadSymbolBody written + unit-tested".
- **what:** One implementation of symbol-body slicing (`analysis.ReadSymbolBody`) placed in BOTH module copies of `internal/analysis` (contextkit and chassis): body = file lines [StartLine, EndLine] inclusive, 1-indexed, exactly as the analyser records; resolves bare names and receiver-qualified `Type.Method`; whole-file for a path with no `:Symbol`. `cmd/assembler`'s duplicate slicing (splitScope/locateSymbol/readLines) was collapsed onto it — "two copies of one convention is the drift this project keeps getting bitten by".
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#1; docs019/RUNBOOK(31)_diagnosis_loop.md#3
- **relations:** diagnose_assemble_bundle; contextkit toolkit; module-copy drift (the two analyse.go copies noted drifted)
- **verify-later:** `internal/analysis/symbolbody.go` in both modules; `symbolbody_test.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose_assemble_bundle action
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) checklist "§2 diagnose_assemble_bundle merged (gofmt-clean)" and "§6C build + register the four diagnose actions … DONE 2026-06-29".
- **what:** The chassis action that, per iteration, reads the in-scope symbols' bodies via ReadSymbolBody from a decoded `repo_analysis` Output, composes hypothesis + code + runtime (+ live schema) into the `bundle` the verdict step reads. Scope fallback chain: `route.scope` (loop-back) → `input_data.seed_scope` → `code_lookup.code_results`. Unknown symbols are logged and skipped, not fatal.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#2; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** ReadSymbolBody; loop_scope_field lesson; diagnosis_artifacts egress (planned write-through here)
- **verify-later:** `platform/orchestration/actions/diagnose_assemble_bundle_action.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Four convergence guards plus engine-level failsafes
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6E.1 "the loop stopped at stopped_by: evidence-not-growing — a guard, not luck. So the guards + the max_iterations cap are armed" (2026-06-29).
- **what:** Deterministic stop conditions independent of model behaviour: iteration-cap, scope-not-narrowing, evidence-not-growing, hypothesis-thrash — plus engine-level `timeout_seconds: 1800` and `fuel_budget: 1000` that bound a runaway even if the loop's bookkeeping is disarmed. Behaviour-tested (26-test suite), not eyeballed; the guards are the safety layer that lets a model verdict be untrusted.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position; docs019/RUNBOOK_design_diagnosis_loop(7).md#0
- **relations:** SeenRequests progress rule; named-scope guard; state threading self-check
- **verify-later:** `pkg/diagnose/loop.go` guards; `loop_test.go`, `step_test.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### SeenRequests — a new data_request counts as loop progress
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "#1 fix … guardAfter now tracks issued read-only data_requests in a SeenRequests set … evidence-not-growing (and hypothesis-thrash) yield when the verdict issues a NEW unseen request"; validated in run 51f95cda ("3 iters, new queries each, no premature stop").
- **what:** Fix for the loop stopping one iteration before its own good query ran: guards treat a NEW unseen read-only data_request as progress (its answer arrives next gather), while a re-issue of the same query still trips the guard. Required the `verdict_wire.go` sync (an older chassis copy silently mapped DataRequests to null, making the engine fix inert).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#1 fix and #1 status)
- **relations:** convergence guards; data_requests channel; verdict wire seam
- **verify-later:** `pkg/diagnose/advance.go` SeenRequests; `loop_datarequest_test.go`; `pkg/diagnose/verdict_wire.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Named-scope guard vs capped call-graph expansion
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "BOTH FIXES DELIVERED 2026-07-03 … Guard now measures the MODEL-NAMED scope …; expansion runs only after the guard passes and is CAPPED (Config.MaxExpandedScope, engine default 18)"; route-close run 73ed55c6 "the expansion cap bounding iterations 2–3 at exactly 18 with all named entries kept".
- **what:** Blocker found when the real 515-file corpus replaced the stale 69-file one: guardAfter measured the POST-EXPANSION scope, and unbounded Neighbourhood expansion of six named symbols tripped scope-not-narrowing at iteration 1. Fix: the narrowing guard compares the MODEL-NAMED scope (deduped NextScope, no expansion); expansion is used only for the gather and capped at MaxExpandedScope (default 18, named entries always kept). A data_request escape on the scope guard was considered and REJECTED (would render it near-inert).
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E attempt-1 blocker 2); docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** convergence guards; call-graph neighbourhood; stale-corpus masking
- **verify-later:** `pkg/diagnose/{loop,step,advance}.go` NamedScopeSize/MaxExpandedScope; `loop_scopeguard_test.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Deterministic scaffold / model-only-verdict split
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) "The scaffold is deterministic; the verdict is the only model-dependent part … This puts the SAFETY … in code that is verified, and isolates the part that needs a model."
- **what:** Architecture decision: loop control, guards, evidence trail, and re-scope are pure tested Go; the cite-or-abstain judgement is an interface (stub that always abstains, scripted verdicts, or the live model). The verdict runs as its OWN observable workflow step (`execute_llm_prompt`), not buried in a monolith. A model-less run can never fabricate a conclusion.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#design-and-build-choices; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** diagnose_run monolith (rejected alternative); verdict wire seam; convergence guards
- **verify-later:** `pkg/diagnose/` package purity (no DB imports); workflow verdict step config

<!-- SOURCE: U14_docs019_runbooks.md -->
### Verdict wire format seam (script IS the wire format)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) §4a "because the script format IS the wire format, every scripted scenario in §1 is a faithful dry-run of the model path"; RUNBOOK(31) §7.5 "keep the prompt's output schema and verdict_wire.go in lockstep".
- **what:** The model returns one JSON object (`outcome` ∈ CONFIRMED|REFUTED|UNVERIFIABLE, citations with `tier` ∈ static|state|runtime, revised_hypothesis, next_scope, data_requests) per PROMPT_diagnosis_verdict.md; `diagnose.ParseVerdict`/`verdict_wire.go` map it to the domain Verdict, with fail-safes: unknown outcome → UNVERIFIABLE, citation-less confirm/refute coerced to UNVERIFIABLE. Verdict scripts for testing use the identical format, so scripted runs are faithful dry-runs of the model path.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4a; docs019/RUNBOOK(31)_diagnosis_loop.md#7.5
- **relations:** cite-or-abstain loop; SeenRequests (wire sync incident); three-tier citation
- **verify-later:** `pkg/diagnose/verdict_wire.go` + `verdict_wire_test.go`; PROMPT_diagnosis_verdict.md

<!-- SOURCE: U14_docs019_runbooks.md -->
### Falsification-first evaluation gate
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6G "[x] PASSED 2026-07-01 (run 51f95cda)"; design_diagnosis_loop(7) §5 "A loop that confirms the first guess on every known bug is the failure mode, not the success — judge it on the reversals."
- **what:** The loop is not trusted on scaffold correctness; it must be run against known bugs and (a) reproduce mid-course REVERSALS (refute wrong hypotheses on evidence), (b) converge on causes the symptom could never retrieve, and (c) ABSTAIN naming the missing evidence when the bundle doesn't settle it. "Scaffold correct ≠ reasons well." The §6G pass showed UNVERIFIABLE→REFUTED→CONFIRMED over 3 iterations with cited evidence.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G; docs019/RUNBOOK_design_diagnosis_loop(7).md#5
- **relations:** gamesdesign resolveResultSpec fixture; three-tier citation; loop-worthiness test
- **verify-later:** evidence trails of runs 51f95cda, 5537ffdb, 73ed55c6 in orchestration_states

<!-- SOURCE: U14_docs019_runbooks.md -->
### gamesdesign resolveResultSpec fixture (the reference bug trajectory)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK(31) 2026-07-01 "STILL not resolveResultSpec — now for a substantive reason: reading real data, the model found a coherent cause … FORK for the user: (a) the fixture is stale … retire the 'reach resolveResultSpec' yardstick".
- **what:** The canonical eval scenario built from the real gamesdesign bug: seed "sections never reach save" → REFUTE on runtime evidence → REFUTE "token cap" → CONFIRM `resolveResultSpec` (singular output_field collapsed the page to a stub). Used as both the scripted-verdict reference and the live-eval yardstick; superseded as a yardstick once the site's current state no longer exhibited the symptom (the loop instead correctly diagnosed the missing `site_specs.cta` aspect), and the route was closed on the refute-and-confirm-a-grounded-cause bar instead.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#7.1; docs019/RUNBOOK(31)_diagnosis_loop.md#6G-passed; docs019/RUNBOOK_gamesdesign_index_rebuild.md
- **relations:** falsification eval gate; workflow result contract; B4a resultspec ceiling task
- **verify-later:** `/tmp` verdict scripts are ephemeral; groundtruth_targets.json resultspec entry

<!-- SOURCE: U14_docs019_runbooks.md -->
### Workflow-driven loop via next_step override (diagnose_route)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6C "[x] DONE"; "§6C coordinator next_step override CONFIRMED (coordinator.go:1093 getNextStepFromResult)"; §6E "[x] DONE 2026-06-29 (5× loop-back, CONFIRMED)".
- **what:** The loop is workflow-driven, not action-internal: `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict (execute_llm_prompt) → route (diagnose_route) → [loop back | emit] → complete`. `diagnose_route` runs the engine's Advance (guards + call-graph re-scope) once per iteration and overrides `next_step` in its result (the conditional_route pattern); it sets no output_field so its results are read as `route.*`. The workflow lives in agent_definitions `default_config` (not the legacy orchestration_workflow columns).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** diagnose_run (abandoned alternative); state threading; coordinator getNextStepFromResult
- **verify-later:** `platform/orchestration/actions/diagnose_route_action.go`; `coordinator.go` getNextStepFromResult; diagnose-agent default_config

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose_run internal-iteration monolith
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** RUNBOOK(5) §6E "In this design there is NO workflow loop-back: the iteration lives inside the diagnose_run action (the engine Run())"; RUNBOOK(31) §6C "The BUILT design is the workflow-driven loop, NOT a diagnose_run action — there is no diagnose_run"; design_diagnosis_loop(7) "(The earlier diagnose_run monolith was removed.)"
- **what:** The earlier design where a single `diagnose_run` action executed the whole capped loop internally (orchestration shows one `run_loop` step; iteration visible only in logs/trail). Dropped in favour of the workflow-driven loop so each iteration's verdict and routing are separately observable orchestration steps. The seeded diagnose-agent briefly referenced the nonexistent action — the workflow-fix migration removed it. Family-delta: present in RUNBOOK(2)–(7), gone by RUNBOOK(8).
- **sources:** docs019/RUNBOOK(5).md#6E; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** workflow-driven loop (replacement); deterministic scaffold split
- **verify-later:** absence of diagnose_run in registry.go; diagnose-agent workflow JSON

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnostician draft and the seed→fix migration path
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK(31) §6C "Do NOT seed a new one (the diagnostician draft is superseded)"; "Do NOT apply the older NNN_move_diagnose_workflow_to_default_config.sql (bannered superseded)"; RUNBOOK(2) §E was "apply the seed migration (NNN_seed_diagnose_agents.sql)".
- **what:** The lineage of getting the diagnose pair into agent_definitions: an early `diagnostician` single-agent draft, then a seed-agents migration (RUNBOOK(2) era), superseded by fixing the ALREADY-seeded diagnose-agent/diagnose-orchestrator pair in place (workflow rewritten to diagnose_route shape in default_config, orchestrator workflow separately restored after the move migration nulled it). Every agent_definitions-touching migration snapshots the row first (`snapshot_agent`).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK(2).md#E
- **relations:** standing evidence rules (snapshot_agent); workflow-driven loop
- **verify-later:** migrations NNN_fix_diagnose_agent_workflow.sql, NNN_restore_diagnose_orchestrator_workflow.sql; agent_definitions snapshots

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose-orchestrator spawn-wrapper pattern
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6F "Target the ORCHESTRATOR; diagnose-agent is the worker it spawns … keeping the loop off shared pods"; run evidence throughout §6E–§6G.
- **what:** The diagnosis entry point is a thin orchestrator (spawn_diagnoser → call_diagnoser → complete) that spawns a dedicated diagnose-agent pod and forwards its result, keeping heavy in-chassis loop work off the shared chassis pods. The same pattern was replicated for indexing (index-orchestrator, §7B.1) when in-place `orchestrate` proved token-less on shared pods.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1
- **relations:** repo-cloning token gate; generic orchestrate envelope; code-indexer
- **verify-later:** diagnose-orchestrator/index-orchestrator agent_definitions; spawn_actions.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### data_requests channel — model-authored read-only SQL
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) update 2026-07-01 (run 51f95cda) "the model's data_requests RAN (verdict_wire.go confirmed live)"; §6C "The data_requests channel — now wired (was dormant from a wiring gap, not by design)."
- **what:** The verdict may emit `data_requests` (single read-only SELECTs with `sql`/`why`); `diagnose_route` reads them from the verdict wire, keeps only read-only ones, forwards to `route.data_requests`; `diagnose_load_runtime` executes each on loop-back in a READ ONLY transaction with SET LOCAL statement_timeout and appends rows to runtime_evidence. Code re-scope and data re-gather are deliberately separate channels. This is the "DB-following" arm of evidence-following.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6B; docs019/RUNBOOK(31)_diagnosis_loop.md#6C (data_requests wiring)
- **relations:** three-guard model; EXPLAIN size guard; SeenRequests; live schema section
- **verify-later:** `diagnose_load_runtime_action.go` runDataRequests; `diagnose_route_action.go` readOnlyDataRequestsFromWire

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three-guard read-only SQL enforcement model
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "SELECT-only is enforced at THREE layers (confirmed in the code)"; design_diagnosis_loop(7) §4d "CONFIRMED on this cluster (2026-06-17): pool_mode = transaction … a live BEGIN READ ONLY; DELETE … WHERE false probe refused the write".
- **what:** Defence in depth for model SQL: Guard 1 = the verdict prompt constrains to a single read-only SELECT; Guard 2 = `IsReadOnlySQL` lint applied twice (route boundary and pre-execution); Guard 3 = the actual guarantee, a `BeginTx(ReadOnly:true)` transaction (+ statement_timeout) that rejects any write including data-modifying CTEs. The `WHERE false` DELETE probe is the standard non-destructive verification. Guards 1–2 are hygiene, never the safety boundary.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#4; docs019/RUNBOOK(31)_diagnosis_loop.md#6B; docs019/RUNBOOK_design_diagnosis_loop(7).md#4d
- **relations:** sqlguard stripQuoted; diagnose_ro role; data_requests channel
- **verify-later:** `pkg/diagnose/sqlguard.go`; BeginTx call in diagnose_load_runtime

<!-- SOURCE: U14_docs019_runbooks.md -->
### sqlguard stripQuoted — lint false-positive on quoted literals
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK(31) (run 5537ffdb, 2026-07-01) "the page slug 'tool-drop-rate-simulator' contains 'drop' … FIXED (sqlguard.go stripQuoted blanks literal/identifier contents before the scan; regression test added)"; §6G banner "REMAINING … (a) DEPLOY the lint fix — latent" (2026-07-02).
- **what:** Keystone bug: the read-only lint scanned raw SQL, so a keyword substring inside a string literal (slug containing "drop") caused legitimate reads to be silently dropped — neutralising both the schema-section content read and the progress rule. Fix blanks literal/identifier contents before keyword scanning. Written + tested; the runbooks record deployment as still pending at the family's last update.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G-passed; docs019/RUNBOOK(31)_diagnosis_loop.md#update-5537ffdb
- **relations:** three-guard model; data_requests channel
- **verify-later:** `pkg/diagnose/sqlguard.go` stripQuoted + test; whether the deployed image carries it

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose_ro role and pooler-aware read-only enforcement
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK(31) checklist "[x] §6B diagnose_ro role migration written … [ ] role migration applied"; RUNBOOK(31) "data_requests run via db.BeginTx(ReadOnly) on params.DB (clients_user), NOT a restricted role".
- **what:** A GRANT-only SELECT role (`diagnose_ro`) for the harness path, where `psql -c` statement stacking makes a transaction wrapper unsafe. Key doctrine: under pgbouncer enforce read-only by GRANT, never by `SET default_transaction_read_only` (session settings leak across pooled backends); transaction pooling makes BeginTx(ReadOnly) safe; statement_timeout goes in the DSN options. The chassis path deliberately runs as clients_user under the read-only transaction instead, so content tables stay SELECTable without grants.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4d; docs019/RUNBOOK(31)_diagnosis_loop.md#6B
- **relations:** three-guard model; dbcontext harness
- **verify-later:** NNN_create_diagnose_ro_role.sql applied?; pgbouncer-config pool_mode

<!-- SOURCE: U14_docs019_runbooks.md -->
### EXPLAIN pre-flight size guard on data requests
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "EXPLAIN size-guard added … runDataRequests now runs EXPLAIN (FORMAT JSON) inside the read-only tx BEFORE each query"; 51f95cda validation "the EXPLAIN guard (didn't block site-scoped queries)".
- **what:** Before executing each model query, the action plans it (EXPLAIN FORMAT JSON, no execution) and skips with feedback if estimated rows exceed budget (explain_max_rows 50000; cost cap opt-in); output rows are capped (row_cap 200) and cells truncated rune-safe (cell_chars 600); statement_timeout remains the execution backstop. A skip is feedback the model narrows from — a new narrower request counts as progress.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel; SeenRequests; responses-are-summaries doctrine
- **verify-later:** runDataRequests EXPLAIN branch in diagnose_load_runtime_action.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### Live schema section in the bundle (gatherSchema)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "#2 IMPLEMENTED … the bundle now carries a ## Schema (live tables) section"; 51f95cda "using REAL table/column names — the schema section paying off, no more page_sections guessing".
- **what:** `diagnose_load_runtime` gains one read-only information_schema.columns query, DENYLIST-driven (%backup%/%bak%/%archive%/%supersede%, deliberately not %snapshot% since site_snapshots is live) plus a broad relevance include (site%/page%/content%/flow%) unless `schema_full=true`; rendered into the bundle via Go-defaulted config so no migration was needed. Stops the model guessing table names (it had invented `page_sections`).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel; denylist-over-allowlist style (shared with index-hygiene excludes)
- **verify-later:** gatherSchema in diagnose_load_runtime_action.go; runtime.schema render path

<!-- SOURCE: U14_docs019_runbooks.md -->
### Loop state threading and the re-seed self-check
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "§6E.1 DONE — state threading verified … trail_len == iteration (3) and stopped on a guard"; the self-check "aborts loudly with stopped_by: state-threading-error rather than silently re-seeding into a runaway".
- **what:** Loop state (iteration, trail, seen_citations, hyp_history) threads across iterations via `state_field = route.diagnose_state`; a mis-pointed state_field silently disarmed the cap/trail/guards (each iteration re-seeded fresh). Fix = migration + a code self-check: if diagnose_route is about to seed but route.diagnose_state already exists, abort loudly — a regression tripwire for the exact bug class.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position; docs019/RUNBOOK(31)_diagnosis_loop.md#6E
- **relations:** convergence guards; loop_scope_field lesson (same dotted-path config family)
- **verify-later:** NNN_fix_diagnose_route_state_threading.sql; self-check branch in diagnose_route_action.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### loop_scope_field / EncodeScope shape-mismatch lesson
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6C "CONFIRMED ISSUE + FIX … EncodeScope is json.Marshal of the Scope struct … keys are the Go field names … ExtractStringListHelper … coerces that OBJECT to empty — so on every loop-back the scope … NEVER advanced"; "loop_scope_field migration confirmed live (Run 2 error read route.scope.Symbols)".
- **what:** A silent contract mismatch between an action's encoded output (untagged Go struct → `{"Symbols":[...]}`) and a downstream dotted-path reader expecting a plain list: first-pass worked, every re-scope was inert — invisible to engine tests because it lived in workflow config. Fix was config-only: point `loop_scope_field` at `route.scope.Symbols`. Emblematic of the dotted-lookup config contract class (also: analysis_field, result_from, repo_field).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C
- **relations:** state threading; repo-label asymmetry; workflow result contract
- **verify-later:** NNN_fix_assemble_bundle_loop_scope_field.sql; ExtractNestedField 3-level traversal

<!-- SOURCE: U14_docs019_runbooks.md -->
### code_symbols index + code-indexer agent
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D "[x] index populated (436 rows)"; code_retrieval_route(21) §7C "4,155 symbols; 499 distinct paths … min=max=commit 36710be; prune cleared all 436 old rows"; measured "≈ 5 symbols/sec through the single ollama-adapter".
- **what:** The retrieval corpus: `code_symbols` (repo, path, symbol, kind, signature, doc, content, embedding, commit_sha) written solely by the `code-indexer` agent (request_repo_analysis → await analyser → index_code_symbols; later analyse_repo_local in-process) and read by `lookup_code_symbols` (vector + trigram). UPSERT-safe via uq_code_symbols_identity; prune removes rows whose commit_sha differs from the new index commit; embedded text is name + signature + first doc line + path. Triggered via index-orchestrator (spawning wrapper so the pod holds the read token).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_code_retrieval_route(21).md#7A–7C; docs019/RUNBOOK_thin_slice(27).md#in-cluster-path
- **relations:** analyser adapter; analyse_repo_local; repo-label convention; evidence-fed resolver
- **verify-later:** code_symbols table + constraints; code-indexer/index-orchestrator agent rows; index_code_symbols action

<!-- SOURCE: U14_docs019_runbooks.md -->
### Analyser adapter — repo analysis as a Kafka service
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D "Adapter availability — RESOLVED 2026-06-24 (secret name/key mismatch) … the adapter came up 1/1 Running"; thin_slice(27) in-cluster deployment section (kustomize dry-build, topic CRD, health checks).
- **what:** A deployed adapter (`internal/adapters/analyser`, topic system.adapter.analyser.requests) that clones a GitHub repo (read-only PAT) and returns the analysis Output over Kafka. Deployment lessons captured: inject the single needed secret via secretKeyRef (never envFrom, which exposes every platform secret); topic auto-create is off so the KafkaTopic CRD must exist; topic-addressed adapters legitimately show target_agent_type='unknown' in awaited_requests; idle consumer-poll timeouts log at ERROR cosmetically. Its per-iteration use by the loop was later removed (analyse_repo_local), but indexing and other consumers remain.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_thin_slice(27).md#in-cluster-path
- **relations:** analyse_repo_local (supersedes the loop's cross-pod call); code-indexer; repo-cloning token gate
- **verify-later:** deployments/kustomize/services/analyser-adapter; personae-platform-secrets GITHUB_READ_TOKEN wiring

<!-- SOURCE: U14_docs019_runbooks.md -->
### Repo-label composition convention (owner/repo) and the lookup asymmetry bug
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Label convention — DECIDED 2026-06-11: code_symbols.repo is the owner/repo form … COMPOSED by index_code_symbols"; RUNBOOK(31) "ROOT CAUSE CONFIRMED (2026-06-26) — repo-label asymmetry … the lookup queried WHERE repo='agentchassis' against rows under 'gqls/agentchassis' → 0 hits"; "Structural patch APPLIED".
- **what:** `code_symbols.repo` is always the composed `owner/repo` label. The index composed it but the lookup didn't → iteration-1 seeding returned nothing ("no scope"). Fixed twice: a config-only workaround (literal repo on the lookup step) then the structural `resolveCodeRepoLabel` shared by index AND lookup so they cannot diverge. Also the standing diagnostic rule it produced: confirm by correlation_id, never by `LIMIT 1` (a COMPLETED LIMIT-1 row was a red herring twice).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F-run1; docs019/RUNBOOK_thin_slice(27).md#label-convention
- **relations:** loop_scope_field lesson (same config-contract class); standing evidence rules
- **verify-later:** resolveCodeRepoLabel in code_symbols_actions.go; lookup step config (no repo_field literal)

<!-- SOURCE: U14_docs019_runbooks.md -->
### analyse_repo_local — in-process tarball fetch + analysis
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6F "DECIDED for option 3, REFINED turn 27 … BUILT (turn 27, gofmt-clean)"; checklist "§6C … image also carries analyse_repo_local + lifted internal/reposource — DONE 2026-06-29"; §7B swap "migration applied; snapshot 971da9c9".
- **what:** Resolution of the "no repo checkout on the diagnose pod" blocker: the agent fetches the repo itself via the analyser's tarball fetcher (`GET /repos/{o}/{r}/tarball/{ref}`, no git in the chassis) lifted into a neutral `internal/reposource` package, runs `analysis.Analyse(dir)` in-process for spans + call graph, and reads bodies from that checkout. `pin_to_index_commit` pins the fetch to the dominant code_symbols commit so seeded path:Symbol entries resolve (the indexer sets it false — it DEFINES the commit). Options weighed and rejected: bodies-in-DB (whole-repo Kafka payloads) and a stateful analyser serving slices (per-iteration coupling).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F (Run 3 + deploy sequence); docs019/RUNBOOK_code_retrieval_route(21).md#7B
- **relations:** analyser adapter; code_symbols; index hygiene excludes; repo-cloning token gate
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; NNN_swap_analyse_repo_to_local.sql

<!-- SOURCE: U14_docs019_runbooks.md -->
### Repo-cloning token gate (isRepoCloningAgent)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "spawn_actions.go injects GITHUB_READ_TOKEN via secretKeyRef … gated by isRepoCloningAgent -> ONLY diagnose-agent pods get the token; the spawner never holds it"; §7B.1 "isRepoCloningAgent gained 'code-indexer' … Verified end to end by run 93ba14e6".
- **what:** Least-privilege credential injection at spawn time: only agent types allowlisted as repo-cloning receive the read-only GitHub token env (secretKeyRef into the spawned pod), and the shared chassis pods never hold it — which is why indexing/diagnosis run through spawning orchestrators rather than in-place.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1
- **relations:** diagnose-orchestrator wrapper; analyser adapter secret lesson
- **verify-later:** spawn_actions.go isRepoCloningAgent list

<!-- SOURCE: U14_docs019_runbooks.md -->
### Stale-corpus class: HEAD pinning, explicit refs, CI-triggered indexing
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** code_retrieval_route(21) §7A "it was a faithful index of a YEAR-OLD tree … Decision: … envelopes ALWAYS carry an explicit branch/sha"; queue item 4 "CI-triggered indexing: GitHub Actions step firing the index-orchestrator envelope with ${GITHUB_SHA} on push … [queued]".
- **what:** A recurring staleness class: consumers pinned to `HEAD`/`latest` silently track an ancient artefact (remote HEAD = unmoved main from 2025; agent image_tag 'latest' = pre-architecture build). Adopted: explicit refs in every envelope, derive REF from the working checkout. Designed (aspirational): Structural A — a post-deploy CI step indexes at ${GITHUB_SHA} so index commit == deployed commit by construction; Structural B — fast-forward main to the deployed sha. Rejected: resolving "most recently pushed branch" via API (latest-pushed ≠ deployed).
- **sources:** docs019/RUNBOOK_code_retrieval_route(9).md#ref-strategy; docs019/RUNBOOK_code_retrieval_route(21).md#7A; docs019/RUNBOOK_builder_route(21).md#queue (item 4)
- **relations:** image_tag 'latest' trap (same class); code_symbols prune semantics
- **verify-later:** GitHub Actions workflow for post-deploy indexing (absent?); git ls-remote origin HEAD

<!-- SOURCE: U14_docs019_runbooks.md -->
### Index hygiene — exclude archived code copies, prune by commit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) §7C.1 "[x] census: docs-archived 430 symbols / 50 files … interim DELETE run"; "[x] reindex f284b749 VERIFIED (2026-07-03): commit e3176f8, 3,723 symbols, docs_rows=0".
- **what:** The repo stores archived copies of its own code under docs/ (and download-suffixed `name(N).go` files); indexing them pollutes retrieval with dead duplicates (observed: nine duplicate assembler copies as ranks 1–9). Fixes: the analyser skips `*(N).go` unconditionally; `analyse_repo_local` gained `exclude_patterns` (Go default ["docs/"]) calling AnalyseWithExclude; prune semantics (`commit_sha IS DISTINCT FROM $new`) clear old-commit rows on the next reindex. Same trap documented CLI-side ("analyse the RIGHT ROOT", relative -exclude substrings).
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7C.1; docs019/RUNBOOK_thin_slice(27).md#known-limits; docs019/RUNBOOK_thin_slice(27).md#B4a (build the index over REAL source)
- **relations:** analyse_repo_local; stale-corpus class; B4a eval discipline
- **verify-later:** exclude_patterns config on analyse_repo_local; code_symbols docs/% row count

<!-- SOURCE: U14_docs019_runbooks.md -->
### Evidence-fed fuzzy-scope resolver (§7D)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) route-close run 73ed55c6: "resolver canonicalisation (model's basenames → full paths @0.81–0.87) AND descriptive resolution … both load-bearing in the confirming scope"; "[x] code WRITTEN 2026-07-02 … resolver image is LIVE".
- **what:** Many verdict `next_scope` entries are English descriptions, not path:Symbol handles — previously inert (no call-graph match, no body sliced). The resolver, inside diagnose_route after verdict-parse and before Advance, embeds each non-exact entry (same nomic client/prefixes) and vector-searches code_symbols, replacing it with the top hits (resolver_top_k default 2 — tuned so substitution stays inside the narrowing guard's +2 allowance; min similarity 0.55; unresolvable entries survive as labels, "no worse"). Flagged deliberate change: the trail records the RESOLVED scope, the more auditable record. Reuses the seed lookup's retrieval machinery wholesale.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D; docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** code_symbols; named-scope guard (the +2 interplay); §7F seed reorder (retired by this)
- **verify-later:** diagnose_route_action.go resolver step 3.5; diagnose_route_resolver_test.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three-tier citation standard (static / data / runtime)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "CONFIRMED on citations spanning ALL THREE TIERS: Tier 2 work item, Tier 1 site_specs cta → (0 rows) (query+result cited together), Tier 0 plan_sections_action.go:planSection quoting 'case \"skip_field\"'" (run 73ed55c6, 2026-07-03).
- **what:** Verdict citations carry a tier (static code / live data reads / runtime records); the route's success bar — and the strongest diagnosis shape — is a CONFIRMED grounded across all three tiers, with query+result cited together for data reads and a quoted code branch for the mechanism. Distinguishes "confirmed by inference at the data layer" from "code-level mechanism named".
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#route-closed; docs019/RUNBOOK_design_diagnosis_loop(7).md#1 (tier vocabulary)
- **relations:** verdict wire seam; falsification eval gate; fix-loop council opinions (verdict-wire-style contract)
- **verify-later:** citation tier handling in verdict_wire.go; evidence_trail of run 73ed55c6

<!-- SOURCE: U14_docs019_runbooks.md -->
### §7F seed-query reorder (lookup after load_runtime)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** code_retrieval_route(21) "§7F RETIRED" (banner) after "SEED RELEVANCE MET — all twelve seed symbols build-domain … first time ever; §7F (seed reorder) substantially retired".
- **what:** A deferred design to reorder lookup_symbols after load_runtime so the seed query could be built from the symptom PLUS salient error-log lines. Made unnecessary once the corpus was current and the resolver landed — seed relevance was proven by content twice. Family-delta: the idea persists as a section in every version but flips from DEFERRED to RETIRED at (18)+.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7F; docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E scoring)
- **relations:** evidence-fed resolver (the retiring cause); code_symbols currency
- **verify-later:** n/a (not built)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Corpus enrichment policy — measure first, mechanical before authored
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** code_retrieval_route(21) "Should every function carry a human description for embedding-match? NO … Order of investment, gated on the §7E measurement" (question raised 2026-07-02).
- **what:** Position on enriching the retrieval corpus: (1) mechanical, rot-free first — extend composeSymbolContent with a function's string literals (diagnosis queries quote log lines and the literals ARE the log lines); (2) Go-convention one-sentence docs only on the exported surface + action entrypoints; (3) explicitly NO separate tag system — the doc first line is the tag surface. Rationale: stale docs make retrieval confidently wrong, the worst failure mode for a cite-or-abstain loop.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#corpus-enrichment
- **relations:** code_symbols; F3 learning layer (doc enrichment feed-back)
- **verify-later:** composeSymbolContent; exported_no_doc census query

<!-- SOURCE: U14_docs019_runbooks.md -->
### Reasoning-state as a first-class handoff artefact
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** thin_slice(27) improvement 5 "A bundle carries CODE + SCHEMA + RUNTIME EVIDENCE, but NOT reasoning state … The stopgap is a hand-written 'diagnosis so far' preamble (PREAMBLE_gamesdesign_diagnosis_handoff.md)"; the loop's evidence_trail later persists per-iteration hypothesis/scope/verdict.
- **what:** The insight that a context bundle without the evidence trail forces a fresh reader to re-derive falsified hypotheses; the design goal is a structured reasoning-state section accumulating across iterations (hypotheses tried, verdict + citation each, open discriminator). Partially realised by the loop's evidence trail in collected_data; the bundle-intrinsic version and per-iteration notes (F0.3 via doc_notes) remain in flight.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#next-improvements (item 5); docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F0.3)
- **relations:** per-task running notes; diagnosis_artifacts egress; falsification eval gate
- **verify-later:** evidence_trail shape in collected_data; doc_notes diagnosis category rows

<!-- SOURCE: U14_docs019_runbooks.md -->
### Instrument-skepticism doctrine
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) "almost every wrong measurement came from the test instrument, not the system under test — a wrong task string, a contaminated index, a stale shell variable, a task description that leaked the answer's vocabulary".
- **what:** Standing caution carried into the loop's design: apply cite-or-abstain suspicion to one's OWN inputs (the bundle, the query, the ground truth) before suspecting the target system. Surfaced repeatedly in B4a and encoded in the eval harness guards; named as the thing to watch when evaluating the model verdict.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#a-standing-caution; docs019/RUNBOOK_thin_slice(27).md#B4a-task-1
- **relations:** ground-truth eval harness; standing evidence rules (0-rows not decisive)
- **verify-later:** n/a (doctrine)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Base-runbook gated-items framing (PLAN.md linkage)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK.md §6 "Gated items (carried — see PLAN.md) … None are unblocked by this thread's work alone" — replaced from RUNBOOK(1) onward by the inlined "§6 Completing the whole task (what remains)" with per-step DoD.
- **what:** The earliest form of the diagnosis runbook kept only in-flight build steps and deferred the roadmap to a separate PLAN.md; within one version the roadmap was inlined as §6 with per-step definitions of done and live status, and the runbook became the single self-contained thread state (later §7 split out to its own file when §6 closed). Family-delta record of the project's documentation style converging on self-contained travelling runbooks.
- **sources:** docs019/RUNBOOK.md#6; docs019/RUNBOOK(1).md#6; docs019/RUNBOOK(31)_diagnosis_loop.md#6 (ACTIVE ROUTE MOVED banner)
- **relations:** parallel-thread convention; documentation-system
- **verify-later:** PLAN.md in docs019 (sibling file, other unit)

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnosis loop (contextkit)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v4 STATE OF THE WORLD (2026-07-02): "Diagnosis loop (§6): DONE. §6A–§6G all passed; §6G accepted on run `51f95cda`... Engine (pkg/diagnose) + diagnose-agent workflow live."
- **what:** A read-only, human-gated agent that diagnoses chassis bugs by iterating hypothesise → gather scoped evidence → cite-or-abstain verdict → re-scope by following the evidence (call graph for code, vetted/model-written queries for data) — never re-searching the symptom, never fixing, never triggering a run. Built first as a standalone Go engine (`contextkit/internal/diagnose/`) with a tested scaffold (guards, trail, verdict parsing), then ported into the chassis as a workflow-driven agent (`diagnose-agent`/`diagnose-orchestrator`) where each iteration is an observable sequence of steps (gather → verdict via `execute_llm_prompt` → `diagnose_route` → loop-back or `diagnose_emit`).
- **sources:** NOTES_running_synthesis_v2(36).md §STATE DIGEST 2026-06-17; NOTES_running_synthesis_v3(32).md §STATE DIGEST; NOTES_running_synthesis_v4(39).md §STATE OF THE WORLD; NOTES_running_synthesis_principles(59) "diagnosis-loop design updated" entries.
- **relations:** contextkit CLI toolchain; convergence guards; verdict cite-or-abstain contract; call-graph re-scope mechanism; B4a embedding-quality finding; diagnosis→fix loop workstream (successor pivot).
- **verify-later:** `pkg/diagnose/` in chassis repo; `platform/orchestration/actions/diagnose_*.go`; `agent_definitions` rows for `diagnose-agent`/`diagnose-orchestrator`; whether the "eval gate" (reproduce the gamesdesign reversals on a live model) was ever actually run.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnosis-loop chassis integration architecture
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v3 STATE DIGEST: "the agent rewrite landed: `diagnose-agent` now has the `diagnose_route` workflow in `default_config`"; v4: "§7 ROUTE CLOSED: run 73ed55c6 full trail read; §7E green".
- **what:** The loop is realised as a chassis AGENT (workflow of steps), not a new CLI or long-running service, following "every agent is an orchestrator": a thin `diagnose-orchestrator` spawns a `diagnose-agent` worker whose workflow (in `default_config`, not the three NULL `*_workflow` columns) is `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict(execute_llm_prompt) → route(diagnose_route) → [loop to assemble_bundle | emit] → complete`. The verdict step reuses the existing `execute_llm_prompt` action rather than a new action; `diagnose_route` is a router action that sets no `output_field` (its result lands under step-name `route`) and returns `next_step` per the coordinator's `getNextStepFromResult` mechanism.
- **sources:** NOTES_running_synthesis_v2(36).md (chassis integration entries, 2026-06-17); NOTES_running_synthesis_v3(32).md DECISIONS (diagnose_route seeding/state-threading fixes).
- **relations:** Diagnosis loop; Workflow default_config location convention; SagaCoordinator output_field contract.
- **verify-later:** `agent_definitions` rows for diagnose-agent/orchestrator; `coordinator.go` `getNextStepFromResult`/`ProcessResponse`.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Verdict cite-or-abstain contract
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the diagnosis loop is now COMPLETE as far as can be built off-chassis — design + tested scaffold + tested adapters + runnable entrypoint + the model prompt + the tested prompt↔scaffold seam" (principles(59), 2026-06-17 entry, mirrored in v2(36)).
- **what:** The model-facing prompt contract (`PROMPT_diagnosis_verdict.md`) requires every verdict to CONFIRM or REFUTE only with a verbatim-quoted citation from the bundle, else the outcome is coerced to UNVERIFIABLE; abstention is asymmetric (runtime evidence readily refutes, but confirms only on direct mechanism, never "consistent with"); the re-scope must follow what the evidence names, not re-search the symptom; and the model is told to apply the same suspicion to its own reading of the bundle that the loop applies to hypotheses. A parallel wire format (`verdict_wire.go`) parses model output as human-legible strings (CONFIRMED/REFUTED/UNVERIFIABLE) with fail-safe unknown→UNVERIFIABLE coercion.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "VERDICT PROMPT drafted"; NOTES_running_synthesis_principles(59) DB discipline / diagnosis-loop design entries.
- **relations:** Diagnosis loop; convergence guards; data-request channel (model-written SQL).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Convergence guards for the diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "The four convergence guards (iteration-cap, scope-not-narrowing, evidence-not-growing, hypothesis-thrash) + the no-citation→UNVERIFIABLE coercion are all behaviour-tested" (v2(36) STATE DIGEST); v3(32) DECISIONS: "A new data_request counts as forward progress in the spin guards (turn 31)".
- **what:** A set of anti-spin safety mechanisms bounding the loop: an iteration cap, a rule that re-scope can't balloon past prior scope + 2, a rule that a verdict adding no new citation halts the loop, and thrash detection for hypothesis oscillation without new discriminating evidence. Later hardened so an issued (not yet cited) read-only data request also counts as forward progress, preventing the loop from stopping one iteration before a fixed query's result would have arrived.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 scaffold entries; NOTES_running_synthesis_v3(32).md DECISIONS (turns 30-31).
- **relations:** Diagnosis loop; verdict cite-or-abstain contract.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Call-graph re-scope mechanism
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "CallGraph adapter... reads analyser `calls`, resolves callee NAMES back to defining symbols for re-scope" (v2(36) 2026-06-17).
- **what:** Re-scoping in the diagnosis loop follows the analyser's recorded (name-based, not type-resolved) call graph outward from an evidence-named site, deliberately dropping ubiquitous names (Run/String/Error/New/... plus any name resolving to more than 8 definitions) so following doesn't explode into noise — described as "the symptom-vocabulary trap in call-graph form."
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "diagnosis-loop adapters... BUILT & tested".
- **relations:** Diagnosis loop; B4a embedding-quality finding.

<!-- SOURCE: U15_docs019_running_notes.md -->
### B4a embedding-quality evaluation & symptom-vs-mechanism retrieval ceiling
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "embeddings do NOT earn a code-path place; the lever is the diagnosis loop... Retrieval is necessary-not-sufficient"; v4(39) STATE OF THE WORLD: "The code-retrieval channel contributes nothing (measured: flat similarity band 0.547–0.574 across all 12 seed hits; zero code citations in four full runs)."
- **what:** A rigorous two-task (later extended) evaluation of lexical vs. semantic (nomic/Ollama) vs. fused (RRF) code-symbol retrieval against real bugs, run through five corrected measurement setups (wrong task string, contaminated index, duplicate-symbol pollution, stale shell var, task-string vocabulary leakage — "the instrument, not the system, was the fault" every time). Conclusion: when a bug's cause lives in shared infrastructure named for its function, not its failure mode, symptom-based retrieval — lexical AND semantic alike — has a category-level ceiling (zero vocabulary overlap, not a ranking problem); naive RRF fusion can make results WORSE than lexical alone by demoting a lone correct hit. Later (v4) measured that the code-retrieval channel contributed essentially nothing across real runs, while runtime/DB evidence carried every successful diagnosis.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-14/06-17 B4a entries; NOTES_running_synthesis_v4(39).md STATE OF THE WORLD.
- **relations:** Diagnosis loop; call-graph re-scope mechanism; code-context retrieval infrastructure; reuse-checking retrieval architecture.
- **verify-later:** `groundtruth_targets.json`, `eval_targets` results in the repo, whether ground truth was ever widened beyond ~2 tasks.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnose-agent self-contained repo fetch
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "Symbol bodies come from a git checkout the diagnose-agent makes itself (Option 3, turns 25-26)."
- **what:** Rather than adding a body column to `code_symbols` or coupling every diagnosis iteration to a live analyser holding a checkout, the diagnose-agent fetches its own tarball (reusing the analyser's `FetchToDir`, lifted into a neutral `internal/reposource` package so both the analyser adapter and the diagnose action share one fetcher) and runs `internal/analysis` in-process for both the call graph and symbol-body slicing — one fetch, no cross-pod coupling, git stays the only source of truth for code. Fetches are pinned to the same commit the `code_symbols` index was built on (best-effort, falls back to `ref`/HEAD) so lookup-seeded symbols resolve in the fetched tree.
- **sources:** NOTES_running_synthesis_v3(32).md turns 25-27 (DECISIONS).
- **relations:** Analyser adapter build; code-context retrieval infrastructure; symbol-body slicer (ReadSymbolBody).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Data-request channel (adaptive DB-evidence gather)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** v4(39) STATE OF THE WORLD: "Correction (the data_requests channel is real and now wired)... it was dormant from a 3-part wiring gap, now fixed."
- **what:** The mechanism by which a diagnosis-loop verdict can name its own read-only SQL query as a `data_request`, which the loop lints (Guard 2), executes read-only (Guard 3), and folds into the next iteration's bundle — replacing an earlier, more limited "vetted query catalogue only" design once the read-only transaction guard was proven sufficient as the real safety boundary. The catalogue survives as a fast-path/few-shot-examples layer, not the only path. Was found dormant (misdiagnosed twice — first as "a gap to wire", then over-corrected to "dormant by design") due to a three-part wiring gap between `diagnose_route`, `diagnose_load_runtime`, and the migration's `gather_step`.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 DB-evidence design entries; NOTES_running_synthesis_v3(32).md STATE DIGEST "Correction" paragraph.
- **relations:** Model-written SQL guard model; diagnosis loop; doc/query catalogue relevance selection.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Doc/query catalogue relevance-keyed selection
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "NEW internal/diagnose/docselect.go: DocRule{Doc, Keywords, PathGlobs, Always} + SelectDocs(hypothesis, scope, rules)... TESTS: docselect_test.go" (v2(36), 2026-06-17).
- **what:** A pure, tested, per-iteration selector (`SelectDocs`/`SelectQueries`, sharing helpers) that pulls task-specific reference documents or SQL query templates into a diagnosis bundle only when their keywords/path-globs match the current hypothesis/scope, keeping the always-on constitution small while still surfacing the relevant 003-style contract or domain query "by relevance" rather than dumping every doc into every bundle (a deliberate anti-bloat decision citing the B4a context-rot lesson).
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "per-hypothesis -doc selection wired into the loop" and "adaptive DB-evidence gather" entries.
- **relations:** Data-request channel; context substrate principles (context rot avoidance).

<!-- SOURCE: U16_docs019_design_plans.md -->
### Iterative-bundle diagnosis loop (the automated debugging motion)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN_diagnosis_loop(3) "Status: … the engine is now BUILT. As of 2026-06-24, `pkg/diagnose` … exists and is tested"; PLAN.md DONE list; README_flows describes live runs.
- **what:** Automates the five-move debugging loop performed by hand in the gamesdesign session: hypothesise from a symptom, gather read-only evidence (bundle), test the hypothesis against the evidence (verdict), re-scope from what the evidence revealed, iterate until pinned or capped. Output is always a diagnosis plus full evidence trail, never a fix. Moves 1/2/4/5 are mechanical; move 3 (falsification) is the crux.
- **sources:** DESIGN_diagnosis_loop(3).md#0-1; README_iterate_until_bugfix_notes.md; README_overview.md; PLAN.md
- **relations:** cite-or-abstain verdict contract; convergence guards; chassis diagnose_route realisation; diagnosis→fix loop
- **verify-later:** pkg/diagnose/{loop,step,advance,callgraph,verdict_wire}.go; contextkit/internal/diagnose; agent_definitions rows diagnose-agent/diagnose-orchestrator

<!-- SOURCE: U16_docs019_design_plans.md -->
### Cite-or-abstain verdict contract + diagnosis verdict prompt
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Prompt is inline (JSON-escaped) in the applied NNN_fix_diagnose_agent_workflow(2).sql verdict step; verdict_wire.go parses its output (tested).
- **what:** Per iteration the model must return exactly one of CONFIRMED / REFUTED / UNVERIFIABLE with verbatim citations; a citation-less confirm/refute is coerced to UNVERIFIABLE; CONFIRMED only on direct evidence ("consistent with" = UNVERIFIABLE — the abstention asymmetry); no fix may be proposed; each citation tier-tagged with freshness. The prompt carries a worked REFUTED example (the gamesdesign reversal) and a self-suspicion caution. Schema must stay in lockstep with verdict_wire.go.
- **sources:** PROMPT_diagnosis_verdict(1).md; DESIGN_diagnosis_loop(3).md#2; NNN_fix_diagnose_agent_workflow(2).sql
- **relations:** doc-drift classifier evidence-or-abstain (its origin); falsification-first principle
- **verify-later:** pkg/diagnose/verdict_wire.go tests; live diagnose-agent default_config verdict step

<!-- SOURCE: U16_docs019_design_plans.md -->
### Falsification-first / confident wrongness as the single enemy
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** README_02 "The single enemy is confident wrongness. Runs 1–2 of the benchmark produced CONFIRMED verdicts that were wrong… Everything since … is aimed at that one failure mode" (2026-07-09 context).
- **what:** The design premise of the whole project: LLMs rationalise their first hypothesis, so every mechanism (citation mandate, REFUTED-is-correct framing, guards, council, closure gate) exists to force explicit falsification and make abandoning a wrong hypothesis cheap. The most valuable move in the founding debug was the model twice saying "my hypothesis is wrong".
- **sources:** DESIGN_diagnosis_loop(3).md#0; README_iterate_until_bugfix_notes.md; README_02_evidence_backed_proposals.md; README_overview.md
- **relations:** cite-or-abstain contract; real-bug eval gate; council pattern
- **verify-later:** eval-run artefacts; benchmark run records (runs 1–2 wrong CONFIRMED)

<!-- SOURCE: U16_docs019_design_plans.md -->
### B4a retrieval ceiling (symptom cannot reach infrastructure-layer causes)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN_diagnosis_loop(3) §1a "B4a (2026-06-17) measured… ALL of lexical, semantic, and fused scored 0.00" on the real gamesdesign fix symbols.
- **what:** Empirical measurement that one-shot retrieval from a symptom description cannot reach a cause living in shared infrastructure named for its function not its failure mode (resolveResultSpec/extractWorkflowResult): the symptom's words and the mechanism's words do not intersect. Lexical beat semantic on the mechanism-named task (0.50 vs 0.00). Consequence: embeddings did not earn a code-path place; the lever is iterative re-scoping following runtime evidence, not better retrieval. Retrieval seeds only the first scope.
- **sources:** DESIGN_diagnosis_loop(3).md#1a; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17 changelog); README_overview.md
- **relations:** evidence-follows re-scoping; text-vs-code embedding split (B4b); code_symbols index
- **verify-later:** contextkit eval_targets + groundtruth_targets.json; go_files/contextkit/{lex,sem}.json

<!-- SOURCE: U16_docs019_design_plans.md -->
### Evidence-follows re-scoping (call graph + runtime-named next scope)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Engine callgraph.go exists/tested per DESIGN(3) status; prompt rule 4 "Follow the evidence to the next scope — do not re-search the symptom".
- **what:** On REFUTED/UNVERIFIABLE the next bundle scopes the symbols/files the evidence names plus their call-graph neighbourhood (the analyser records `calls`), and prefers a runtime-named fault site over a retrieval-proposed one. This is the move retrieval cannot do; it reached the coordinator's result extraction in the real case — a symbol the symptom could never name.
- **sources:** DESIGN_diagnosis_loop(3).md#1a; PROMPT_diagnosis_verdict(1).md rule 4; NNN_fix_assemble_bundle_loop_scope_field.sql
- **relations:** B4a ceiling; convergence guards; Go analyser call graph
- **verify-later:** pkg/diagnose/callgraph.go; diagnose_route re-scope path

<!-- SOURCE: U16_docs019_design_plans.md -->
### Convergence guards (cap, narrow, grow, no-thrash)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN(3) status: "loop control, convergence guards … exists and is tested"; migration(19): "Four convergence guards + the no-citation→UNVERIFIABLE coercion all behaviour-tested".
- **what:** Deterministic Go guards so the loop cannot run forever or wander: iteration cap (5); scope must narrow (widening = not converging); evidence must grow (two iterations without new grounded evidence → stop with best-effort); no hypothesis thrash (oscillation without discriminating evidence → report both). Deliberately kept in tested Go, never re-expressed as workflow conditionals.
- **sources:** DESIGN_diagnosis_loop(3).md#3; PLAN_workflows_and_actions_migration(19).md; NNN_fix_diagnose_route_state_threading(1).sql
- **relations:** state-threading fix (guards were silently inert live); thin-workflows rule
- **verify-later:** pkg/diagnose/step.go DecideStep + tests

<!-- SOURCE: U16_docs019_design_plans.md -->
### Read-only, human-gated boundary of the diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN chassis integration(6) §7 "boundaries … do not relax in the chassis"; HANDOFF_fixloop(8) "Loop core is READ-ONLY by contract; sqlguard allowlists reads — keep it so".
- **what:** The loop gathers read-only (analyser, code_symbols, `\d`/capped SELECT/existing-log reads), proposes a diagnosis + suggested fix surface, and never applies fixes or triggers runs to test hypotheses. The human is kept at the two points that mattered: deciding the fix and backstopping the model's willingness to abandon a hypothesis. The F1 write surface is deliberately a separate agent with isolated credentials.
- **sources:** DESIGN_diagnosis_loop(3).md#4; DESIGN_diagnosis_loop_chassis_integration(6).md#7; HANDOFF_fixloop_thread(8).md#3
- **relations:** fix-implementer (the separate write surface); doc-drift read-only rule; three-guard read-only SQL
- **verify-later:** pkg/diagnose/sqlguard.go; spawn token-gate in spawn_actions.go

<!-- SOURCE: U16_docs019_design_plans.md -->
### Evidence tiers with freshness tagging (static / state / runtime)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Prompt rule 6 (tier + `fresh` per citation) in the applied workflow migration.
- **what:** Every citation is tagged static (code/schema), state (a DB row at a point in time) or runtime (log/work-item from an actual run), with observation time for state/runtime, so a verdict resting on stale evidence is visibly weak. Adapted from the doc-drift classifier's T1/T2/T3.
- **sources:** PROMPT_diagnosis_verdict(1).md rule 6; DESIGN_diagnosis_loop(3).md#2; DESIGN_doc_drift_classifier.md#2
- **relations:** doc-drift evidence tiers; misattribution asymmetry
- **verify-later:** verdict_wire.go citation struct

<!-- SOURCE: U16_docs019_design_plans.md -->
### Chassis realisation: diagnose_route workflow-driven loop (four diagnose actions)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Chassis-integration(6) banner (2026-06-24) "the BUILT design is `diagnose_route` … The actions that actually exist are diagnose_load_runtime, diagnose_assemble_bundle, diagnose_route, diagnose_emit"; live runs 8d488e01/73ed55c6 in fix migrations.
- **what:** On the chassis the loop is an agent workflow: analyse_repo → lookup_symbols (seeds iteration-1 scope from the symptom) → load_runtime → assemble_bundle → verdict (its own execute_llm_prompt step, observable) → diagnose_route (engine Advance: guards + call-graph re-scope, then a next_step override, the conditional_route pattern) → back to load_runtime or emit. The loop returns to load_runtime, not assemble, so the prior verdict's data_requests run and runtime re-gathers each iteration. Gather reuses existing actions (request_repo_analysis→analyse_repo_local, lookup_code_symbols, execute_llm_prompt) per the STEP-ZERO reuse audit; only the bundle composer was genuinely new.
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md#0,#status; NNN_fix_diagnose_agent_workflow(2).sql; PLAN.md; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17 entry)
- **relations:** abandoned diagnose_run and diagnostician designs; one-decision-core two realisations
- **verify-later:** platform/orchestration/actions/diagnose_*_action.go; coordinator.go getNextStepFromResult; registry.go Category "diagnose"

<!-- SOURCE: U16_docs019_design_plans.md -->
### Abandoned design: diagnose_run single engine-wrapping action
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** Chassis-integration(6) banner: "the §4–§6 `diagnose_run` recommendation below is the ABANDONED path … there is no `diagnose_run` action"; the seeded workflow referencing it was rewritten by NNN_fix_diagnose_agent_workflow.
- **what:** The originally recommended shape — one `diagnose_run` action calling `diagnose.Run()` with an injected Verdicter, keeping the whole loop inside a single step. Dropped in favour of the workflow-driven observable loop (verdict as its own step, router action). A prompt-registry reference `diagnose-verdict-v1` belonged to this design and is also unused; the prompt went inline instead. Kept here because seeded rows briefly referenced the non-existent action (a real incident class: workflow names an action that does not exist).
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md banner,#4-6; NNN_fix_diagnose_agent_workflow(2).sql header; NNN_move_diagnose_workflow_to_default_config(1).sql banner
- **relations:** superseded by diagnose_route realisation
- **verify-later:** absence of diagnose_run in registry.go

<!-- SOURCE: U16_docs019_design_plans.md -->
### Abandoned design: diagnostician per-iteration re-invocation (spawn-next chain)
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** NNN_seed_diagnose_agents(2).sql banner "SUPERSEDED — DO NOT APPLY … kept only as a record of the re-invocation design that was considered and dropped."
- **what:** A third loop shape: each orchestration runs ONE iteration (load_runtime → analyse → lookup → assemble → verdict → route → conditional), and on continue spawns+calls a fresh `diagnostician` of the same type with revised hypothesis/scope/iteration in input_data, the terminal verdict bubbling up the child chain. Motivated by doubt that the engine supported a workflow-internal cycle and by the build-dispatch-loop one-unit-per-orchestration precedent. Dropped once the next_step-override loop-back was confirmed to work.
- **sources:** NNN_seed_diagnose_agents(2).sql header + workflow body
- **relations:** superseded by diagnose_route loop-back; build-dispatch-loop pattern
- **verify-later:** no `diagnostician` row in agent_definitions

<!-- SOURCE: U16_docs019_design_plans.md -->
### One decision core, two realisations (Run vs Advance/DecideStep)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN(3) §1 note: "Both share ONE decision core — step.go's pure DecideStep … advance_test.go proves Advance threaded across iterations … reproduces Run()".
- **what:** The standalone dev harness (`internal/diagnose/loop.go` Run, a Go for-loop with inline IO) and the chassis workflow loop share the same pure per-iteration decision function, with `advance.go`'s Advance exposing it statefully to the chassis; equality is proven by test. cmd/diagnose stays the dev/test harness (scripted verdicts, dry-bundle), never a production entrypoint.
- **sources:** DESIGN_diagnosis_loop(3).md#1; DESIGN_diagnosis_loop_chassis_integration(6).md#status,#3; PLAN_workflows_and_actions_migration(19).md
- **relations:** engine/harness file-placement split; travelling contextkit module
- **verify-later:** pkg/diagnose/advance.go + advance_test.go

<!-- SOURCE: U16_docs019_design_plans.md -->
### Model-written data_requests under three-guard read-only SQL
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** PLAN.md DONE: "Three-guard model-SQL feature: Guard 1 (prompt), Guard 2 (IsReadOnlySQL), Guard 3 (read-only transaction, confirmed on the chassis side)"; GATED item 1: execution wiring in diagnose_load_runtime is deploy-side and untested; README_flows: "terminal-verdict data_requests that never run".
- **what:** The verdict may request specific evidence as single read-only SELECTs (`data_requests: [{sql, why}]`), defended by three independent guards: the prompt contract, a Go lint (sqlguard.IsReadOnlySQL), and execution inside a read-only transaction with statement timeout; the harness analogue is a GRANT-based SELECT-only `diagnose_ro` role (not default_transaction_read_only, unreliable under pgbouncer). Notably this reversed an earlier stance — chassis-integration(6) recorded "the model never writes SQL" and called runDataRequests dormant/beyond the boundary; the bounded, guarded version was then built deliberately.
- **sources:** PROMPT_diagnosis_verdict(1).md rule 7; NNN_create_diagnose_ro_role.sql; PLAN.md; DESIGN_diagnosis_loop_chassis_integration(6).md#status (the earlier stance)
- **relations:** read-only boundary; self-verification in the council (same move at review time)
- **verify-later:** pkg/diagnose/sqlguard.go in chassis; diagnose_load_runtime data-request execution path; diagnose_ro role existence

<!-- SOURCE: U16_docs019_design_plans.md -->
### Real-bug evaluation gate (scaffold correct ≠ reasons well)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** PLAN.md GATED 5: "THE EVAL GATE … MUST reproduce the mid-course REVERSALS and ABSTAIN when unsettled … No automatic triggering until this passes"; README_02 shows benchmark runs happened (runs 1–2 confidently wrong) and design responded.
- **what:** Before any unsupervised or automatic triggering, the live loop must be run against known bugs — the gamesdesign two-fault bug (with its captured reversals) and the 016 §9 silent-no-op catalogue — and must reproduce hypothesis reversals and abstain when evidence does not settle, rather than confirming first guesses. "Compiling isn't behaving" is the standing lesson; a loop that confirms its first guess every time is the failure mode dressed as success.
- **sources:** DESIGN_diagnosis_loop(3).md#6; PLAN.md GATED; README_whats_next.md; README_02_evidence_backed_proposals.md
- **relations:** gamesdesign bug fixture; falsification-first; later trigger modes gated on this
- **verify-later:** eval run records; whether triggers (b)/(c) were ever enabled

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnose agent pair + generic-request trigger envelope
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Restore/fix migrations applied against live rows (snapshot ids 34f4afc8/e8e96d24 in HANDOFF_builder_thread); TRIGGER pattern proven (084_TRIGGER_diagnose_v1.sh referenced as canonical in HANDOFF_fixloop(8)).
- **what:** A thin diagnose-orchestrator (spawn_agent → call_agent → complete, per "every agent is an orchestrator" and the wrapper rule for substantive work) spawns the diagnose-agent worker pod that runs the loop. Triggering is the existing generic-request envelope — kcat to system.agent.generic.requests with agent_type diagnose-orchestrator and input_data {symptom, seed_scope, runtime_site, …} — no new triggering code; later triggers (on build failure, proactive sweep) are the same envelope from a different sender, gated on the eval. Sub-agents reply on the caller's responses topic.
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md#2; NNN_restore_diagnose_orchestrator_workflow(1).sql; PLAN_workflows_and_actions_migration(19).md
- **relations:** wrapper-orchestrator pattern; index-orchestrator (same pattern reused)
- **verify-later:** agent_definitions diagnose-orchestrator/diagnose-agent; drafts/084_TRIGGER_diagnose_v1.sh

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnose loop-back plumbing fault class (state threading, scope encoding)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Both fix migrations are operative live fixes with run evidence (run 8d488e01: "there is a 'route' key, no 'diagnose_state' key"; trail truncated to 1 entry).
- **what:** Two producer/consumer field mismatches that left the loop running but silently degraded: (1) diagnose_route read its LoopState from bare `diagnose_state` while its own output lands under `route.diagnose_state` — so the loop re-seeded every iteration, never enforcing max_iterations, truncating the evidence trail, and resetting the cross-iteration guards; (2) route.scope is EncodeScope's untagged-struct JSON, so the string-list reader needed `route.scope.Symbols` (capital S) — before the fix every re-scope silently fell through to the fallback chain and iterations 2+ never moved scope. Both were invisible-success faults: the loop "worked" while its defining features were inert.
- **sources:** NNN_fix_diagnose_route_state_threading(1).sql; NNN_fix_assemble_bundle_loop_scope_field.sql; DESIGN_diagnosis_loop_chassis_integration(6).md#status (the round-trip flagged as unverified)
- **relations:** workflow-variable-sync rule; result-contract dead-key class; convergence guards
- **verify-later:** diagnose_route_action.go default state_field; Scope struct json tags

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnosis persistence + documented intake (diagnosis_artifacts, needs_diagnosis)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_fixloop(8): "First action: slice F0.1 with pre-registered criteria — (1) diagnosis_artifacts migration … (2) assemble-action write-through … (3) the needs_diagnosis envelope"; decisions recorded 2026-07-07, not yet built in these files.
- **what:** F0 of the fix loop: make each iteration's bundle durably fetchable and add per-iteration running notes — a `diagnosis_artifacts` table (correlation_id, iteration, kind ∈ {bundle, iteration_note}, body, retention knob per kind) written through from the assemble action Go-side (no workflow-shape change); plus a documented intake route: a `needs_diagnosis` envelope / pipeline='diagnose' work item carrying subject_type/subject_key with null-site allowed. Bundle egress via completion payloads is bounded (max_response_bytes) — persist and reference, don't ship megabytes.
- **sources:** HANDOFF_fixloop_thread(8).md#4; HANDOFF_fixloop_thread(8).md#3
- **relations:** result-contract size guard; travelling-docs pattern (notes per iteration); work-item relay
- **verify-later:** diagnosis_artifacts table existence; persist_note step in diagnose-agent workflow (tools thread's wiring)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Verdict-quality wrinkles + dead code-retrieval channel (measured)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** README_flows: "seed similarities in the 0.55 band, no page-build symbols … a stale line carried into a conclusion that its own citation contradicts, and terminal-verdict data_requests that never run".
- **what:** Post-run findings on the live loop: the lookup channel contributes nothing measurable (work is on the query side — seed the lookup from runtime evidence or expand the query, a self-contained lookup_symbols change); the trigger's site_id is intermittent across runs (reproducibility); and two verdict-quality defects point at the confirm/emit step (a conclusion contradicted by its own citation; data_requests emitted on terminal verdicts that never execute).
- **sources:** README_flows.md; PLAN.md GATED
- **relations:** B4a ceiling; data_requests wiring; eval gate
- **verify-later:** lookup_symbols seeding config; emit/confirm step handling of terminal data_requests

<!-- SOURCE: U17b_docs019_gofiles.md -->
### contextkit CLI toolkit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "go build ./...  # compiles all seven commands" (README.md); real invocation example wired to a live site in cmd/bundle/README.md (`-runtime-site gamesdesign.co.uk`).
- **what:** A standalone Go module (`module contextkit`, go 1.22) of CLI tools for building LLM context bundles from a repo without a live cluster: analyser, assembler, dbcontext, bundle, embed, resolve_targets, fuse, eval_targets, dedup, thin_versions. Compiles and runs independently of the agentchassis repo; two of its packages (`internal/analysis`) are shared verbatim with the chassis.
- **sources:** contextkit/README.md, contextkit/README(2).md, contextkit/go.mod
- **relations:** diagnosis loop (internal/diagnose), analyser-adapter deployment plan, thin-slice constitution
- **verify-later:** does `internal/analysis` in this tree still match `internal/analysis/` at the agentchassis repo root byte-for-byte (README.md flags this as a manual sync obligation, not automated)

<!-- SOURCE: U17b_docs019_gofiles.md -->
### analyser (cmd/analyser)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "thin CLI wrapper over analysis.AnalyseWithExclude — the same parsing primitive the in-cluster analyser adapter imports"
- **what:** Walks a Go source tree and emits a structural-summary JSON (files, packages, imports, function/method signatures with callee names, struct/interface declarations with line ranges). Always skips vendor/, testdata/, hidden dirs, `*_test.go`, and `*(N).go` download-duplicates; takes an `-exclude` list for repos (like this one) that store archived copies of their own code under docs/.
- **sources:** contextkit/cmd/analyser/main.go#header, contextkit/README.md
- **relations:** internal/analysis package, code-indexer agent (chassis-side counterpart), embed/resolve_targets (consume analyser JSON)
- **verify-later:** internal/analysis (agentchassis repo root) — confirm the in-cluster analyser adapter still calls `analysis.Analyse` (no-exclude) as documented

<!-- SOURCE: U17b_docs019_gofiles.md -->
### internal/analysis package (analyser output contract)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Package analysis defines the contract the analyser emits and the assembler, embed, and resolve_targets consume... defined only here" (types.go header); "the harness and production parse identically" (analyse.go header)
- **what:** The single-source-of-truth `Output`/`FileInfo`/`FuncDef`/`TypeDef` contract for repo structural analysis, plus `Analyse`/`AnalyseWithExclude` (the layer-1 AST walk) and `ReadSymbolBody` (slices a `path:Symbol` scope into source text using the analyser's recorded line span, never re-parsing). Intentionally Go-only; a non-Go producer would fill the same contract behind the analyser adapter.
- **sources:** contextkit/internal/analysis/types.go#header, contextkit/internal/analysis/analyse.go#header, contextkit/internal/analysis/symbolbody.go#header
- **relations:** analyser, assembler, embed, resolve_targets, cmd/bundle (also uses the symbol slicer), chassis diagnose_assemble_bundle action
- **verify-later:** whether the chassis's diagnose_assemble_bundle action's old inline `readSymbolBody` stub has actually been replaced by a call to `ReadSymbolBody` as the header claims is the intent

<!-- SOURCE: U17b_docs019_gofiles.md -->
### assembler (cmd/assembler)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "That's the whole thin slice working: analyser, constitution, assembler." (README_how_to_run_analyser.md)
- **what:** Builds one paste-ready markdown bundle for a single task: consumes the analyser JSON, the repo (to pull full bodies by line range), the flat constitution, and a task+scope spec; renders constitution, task, in-scope code in full, neighbourhood signatures (same-package, capped ~60/package), schema (hand-fed), and a pointers note of what was omitted. `-step` (framing/implementation/debug) controls altitude — framing shows signatures only, implementation/debug show full in-scope bodies.
- **sources:** contextkit/cmd/assembler/main.go#header, contextkit/README_how_to_run_analyser.md, contextkit/001_more_potential_thin_slice_prompt.md
- **relations:** internal/analysis (symbol slicing), thin_slice_constitution.md, bundle (wraps it), docselect/queryselect (chassis analogues for doc/query selection instead of hand-specified scope)
- **verify-later:** confirm the neighbourhood-signature cap and package-scoping behaviour match what 001_more_potential_thin_slice_prompt.md's design notes describe

<!-- SOURCE: U17b_docs019_gofiles.md -->
### dbcontext (cmd/dbcontext)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** used directly in cmd/bundle/README.md's worked example (`-psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db'`)
- **what:** Fetches DB context for a bundle by shelling out to a configurable `psql` — schema (`\d <table>`, complete+bounded), rows (multipass-sized SELECT: full if within cap, else sample + pointer query, never an unbounded dump), and capabilities (`\dx`, `\df`). No Go DB driver; psql does the talking, so it inherits whatever connection role/permissions the operator supplies.
- **sources:** contextkit/cmd/dbcontext/main.go#header
- **relations:** bundle (wraps it), sqlguard (lints model-written queries elsewhere in the pipeline), database-and-infrastructure conventions
- **verify-later:** whether the psql connection used in production is provisioned as a read-only role (sqlguard.go explicitly says the lint alone is not the safety boundary — the read-only transaction/role is)

<!-- SOURCE: U17b_docs019_gofiles.md -->
### bundle (cmd/bundle)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** concrete real invocation against `gamesdesign.co.uk` in cmd/bundle/README.md; used by bundle_diagnosis_loop.sh, bundle_minilobby_trim2/3.sh, bundle_recreation_v1.sh
- **what:** A thin orchestration wrapper around dbcontext + assembler: gathers read-only DB context (schema/capabilities/runtime evidence), writes each to a temp file, then invokes the assembler with those files wired in. Deliberately never runs SQL itself (that stays in dbcontext) so the assembler stays a pure, read-only, offline composer — the wrapper "triggers NOTHING — no builds, no spawns, no writes."
- **sources:** contextkit/cmd/bundle/main.go#header, contextkit/cmd/bundle/README.md
- **relations:** dbcontext, assembler, gatherer.go (BundleGatherer shells out to this exact binary from the diagnosis loop)
- **verify-later:** BundleGatherer.buildArgs (gatherer.go) — confirm the flag set it constructs still matches this binary's real flags

<!-- SOURCE: U17b_docs019_gofiles.md -->
### embed (cmd/embed)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "-local ... proves the pipeline (index → cosine → rank) WITHOUT a model, but is NOT semantic — use -ollama for real recall" (header)
- **what:** Builds/queries a semantic vector index over the analyser's symbols — the recall layer for target resolution sitting above the lexical baseline (resolve_targets). Model-agnostic via an embedder interface: `-ollama` (real embeddings, e.g. nomic-embed-text) or `-local` (deterministic offline token-hashing stand-in for pipeline-proving only). Index and query must use the same embedder/vector space.
- **sources:** contextkit/cmd/embed/main.go#header
- **relations:** resolve_targets, fuse (RRF-merges embed's output with resolve_targets'), eval_targets (scores it), code-indexer agent (chassis-side embedding via the same ollama-adapter/nomic-embed-text pairing)
- **verify-later:** whether production bundle-building actually runs `embed` with `-ollama` or still relies on the `-local` stand-in

<!-- SOURCE: U17b_docs019_gofiles.md -->
### resolve_targets (cmd/resolve_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the deterministic baseline — the layer that runs before any embeddings" (header)
- **what:** A first-cut, lexical-overlap target resolver: given a task string and the analyser JSON, proposes ranked candidate symbols/files to `-scope` by matching the task's distinctive words against each symbol's name, path, and docstring. Does not decide — proposes a ranked candidate set for a human or the assembler to confirm.
- **sources:** contextkit/cmd/resolve_targets/main.go#header
- **relations:** embed (semantic counterpart), fuse (merges both), internal/candidates (shared output contract), eval_targets
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### fuse (cmd/fuse)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "score(item) = sum over lists of 1/(k + rank_in_list), k=60 (standard)" (header)
- **what:** Merges ranked candidate lists (resolve_targets' lexical output + embed's semantic output) into one ranking via reciprocal-rank fusion (RRF). Combines by RANK not score specifically because the lexical integer scores and semantic cosine scores aren't on a comparable scale.
- **sources:** contextkit/cmd/fuse/main.go#header
- **relations:** resolve_targets, embed, internal/candidates, eval_targets (scores fuse's output too)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### eval_targets (cmd/eval_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** the ground-truth file contains a task tied to a "REAL fix (2026-01 chassis regression)" already applied, showing the harness is exercised against real cases, not just synthetic ones
- **what:** Scores a resolver's candidate list (`-json` output of resolve_targets/embed/fuse) against a ground-truth set mapping tasks to the symbols they actually needed — turns "the fused list looks better" into numbers: recall@N over decisive symbols, and MRR contribution (rank of first decisive hit). Match is on `path:name`.
- **sources:** contextkit/cmd/eval_targets/main.go#header, contextkit/groundtruth_targets.json
- **relations:** resolve_targets, embed, fuse (all scored by this), llm-quality-testing (evaluation-harness pattern), ground-truth eval set concept below
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### diagnose (cmd/diagnose)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "THE VERDICT STEP IS NOT THE REAL MODEL HERE... a chassis-side follow-on (needs a model). This entrypoint ships two stand-ins" (header)
- **what:** Wires the diagnosis-loop scaffold (internal/diagnose) to real adapters — BundleGatherer (shells to cmd/bundle, read-only) and AnalysisCallGraph (follows the analyser's `calls` for re-scope). The verdict step is stubbed (either a scripted JSON array of verdicts for testing, or a trivial always-UNVERIFIABLE default) since the real cite-or-abstain LLM verdicter needs a model and lives chassis-side. Explicitly read-only and human-gated: emits a diagnosis + evidence trail, never a fix, never a triggered run.
- **sources:** contextkit/cmd/diagnose/main.go#header
- **relations:** internal/diagnose (loop.go, step.go, verdict_wire.go, callgraph.go, gatherer.go), fixloop workstream (the diagnose→fix pipeline this scaffold feeds)
- **verify-later:** docs024_key_docs_latest/fixloop_eg_dartsonline/ for whether/how a real LLM verdicter has since been wired in chassis-side

<!-- SOURCE: U17b_docs019_gofiles.md -->
### diagnosis-loop scaffold (internal/diagnose, loop.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "WHAT LIVES HERE (deterministic, testable without a model): loop control, the guards..., the evidence trail, and the re-scope mechanism. WHAT DOES NOT... the Verdict step" (loop.go header); backed by loop_test.go, loop_datarequest_test.go, loop_scopeguard_test.go
- **what:** The deterministic core of the diagnosis loop: wraps a read-only gather step around a pluggable verdict step, enforces convergence guards (iteration cap, scope-must-narrow, evidence-must-grow, no-thrash), accumulates an evidence trail, and re-scopes by FOLLOWING runtime/call-graph evidence rather than re-searching the symptom — named as the fix for a "ceiling" where symptom-only retrieval fails on infrastructure-layer causes. Non-negotiable boundary: never applies a fix, never triggers a run to test a hypothesis.
- **sources:** contextkit/internal/diagnose/loop.go#header, contextkit/internal/diagnose/loop_scopeguard_test.go#header, contextkit/internal/diagnose/loop_datarequest_test.go#header
- **relations:** step.go (DecideStep), advance.go (chassis-facing wrapper), callgraph.go, verdict_wire.go, docselect.go, queryselect.go, sqlguard.go, gatherer.go, fixloop workstream
- **verify-later:** whether the "guard-vs-expansion" bugfix noted in loop_scopeguard_test.go (run 17933a83) and the data_request evidence-growth fix (loop_datarequest_test.go, "truncated the live gamesdesign runs at iteration 3") are reflected in the currently-deployed chassis diagnose_run/diagnose_route actions

<!-- SOURCE: U17b_docs019_gofiles.md -->
### DecideStep — shared pure per-iteration decision (step.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Extracting it keeps ONE source of truth for the guard + re-scope logic instead of two copies that could drift. Run() is refactored to call Step(); the existing tests are the proof the behaviour is unchanged." (header)
- **what:** The per-iteration decision (given iteration state, a verdict, the call graph, and guard memory) extracted as a pure function, shared by the standalone `Run()` loop and the chassis `diagnose_run` workflow action (where the verdict is a separate workflow step). Guarantees one source of truth instead of two logic copies that could drift apart.
- **sources:** contextkit/internal/diagnose/step.go#header, contextkit/internal/diagnose/step_test.go
- **relations:** loop.go, advance.go (LoopState calls this per-iteration)
- **verify-later:** confirm the chassis `diagnose_run` action actually calls this shared `Step()`/`DecideStep` rather than a re-implementation

<!-- SOURCE: U17b_docs019_gofiles.md -->
### LoopState — chassis-facing per-iteration API (advance.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "the chassis needs: (1) a LoopState it can thread through workflow collected_data between iterations, (2) a single Advance() call per iteration, and (3) parse helpers for a verdict that arrives as an already-unmarshalled map" (header)
- **what:** The workflow-driven realisation of the loop: since the chassis loop is `gather → verdict step → diagnose_route → back | emit` (not an in-process loop), LoopState carries loop memory across workflow steps via `collected_data`, with `Advance()` as the one call per iteration and `EncodeLoopState`/`DecodeLoopState` for the JSON round-trip. Adds no new decision logic beyond step.go's DecideStep plus state bookkeeping.
- **sources:** contextkit/internal/diagnose/advance.go#header, contextkit/internal/diagnose/advance_test.go
- **relations:** step.go, loop.go, chassis diagnose_route workflow step
- **verify-later:** platform/orchestration — the actual `diagnose_route` step and its `collected_data` schema, to confirm it matches `EncodeLoopState`'s shape

<!-- SOURCE: U17b_docs019_gofiles.md -->
### AnalysisCallGraph — call-graph re-scope (callgraph.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "HONEST LIMIT (from the analyser): `calls` is NAME-BASED, not type-resolved... DELIBERATELY DROPS ubiquitous names (Run, String, Error, …)" (header)
- **what:** A CallGraph implementation backed by the analyser's recorded (name-based, not type-resolved) `calls` field, letting the diagnosis loop re-scope by following the call graph from an evidence-named site rather than re-searching the symptom. Explicitly drops ubiquitous method names that would otherwise explode the neighbourhood into noise — the loop's narrowing guard is the backstop, but dropping known-ubiquitous names keeps re-scope sharp at the source.
- **sources:** contextkit/internal/diagnose/callgraph.go#header
- **relations:** internal/analysis (the `calls` data it consumes), loop.go's re-scope mechanism, cmd/diagnose (wires this in as the real adapter)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### verdict wire format (verdict_wire.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "If the prompt's output schema and this struct ever drift, the loop breaks at this join — so this file is tested against example model outputs." (header)
- **what:** The JSON wire format the (LLM) Verdicter emits — human-legible strings (`"REFUTED"`, `"runtime"`) and snake_case keys rather than the domain type's int enums, so a model can produce it reliably — and the parser (`ParseVerdict`) translating it into the domain `Verdict`. Named as the ONE seam between the verdict prompt's specified output (`docs/PROMPT_diagnosis_verdict.md`) and the scaffold; a verdict-script in this format is a faithful stand-in for the real model.
- **sources:** contextkit/internal/diagnose/verdict_wire.go#header, contextkit/internal/diagnose/verdict_wire_test.go
- **relations:** diagnose (cmd), loop.go, docs/PROMPT_diagnosis_verdict.md (referenced, not in this unit)
- **verify-later:** docs/PROMPT_diagnosis_verdict.md — confirm its schema still matches this struct (the header itself flags drift risk here)

<!-- SOURCE: U17b_docs019_gofiles.md -->
### docselect — per-hypothesis doc selection (docselect.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Selection is DETERMINISTIC and testable... so it can be exercised without a model. It is a HEURISTIC (keyword/path substring)" (header); docselect_test.go exercises keyword/always/path-glob rules
- **what:** Per-hypothesis selection of authored context docs (the 003 contract sections, 016 §9 entries, dev-guide sections) to paste into the CURRENT iteration's bundle rather than every doc into every bundle — avoiding the "irrelevant context buries the signal" failure mode. A future extension is floated (not built): letting the verdict NAME a needed doc via a `needed_docs` field mirroring `needed_evidence`/`next_scope`.
- **sources:** contextkit/internal/diagnose/docselect.go#header, contextkit/internal/diagnose/docselect_test.go
- **relations:** thin_slice_constitution.md (the always-on layer this supplements), queryselect.go (data analogue), contracts-and-standards (003)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### queryselect — vetted read-only query catalogue (queryselect.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "WHY A CATALOGUE, NOT MODEL-WRITTEN SQL (the safety boundary)... the queries are HAND-WRITTEN, parameterised, and \\d-verified ONCE; the loop only SELECTS among them by hypothesis. The model never writes SQL." (header)
- **what:** Per-hypothesis selection of vetted, read-only, parameterised DB queries for the runtime-evidence gather — the data analogue of docselect.go. Queries bind to the loop's existing context (site_id, domain, page, correlation_id already in input_data/seed), so no wire-format change or model-supplied SQL parameters are needed. This is presented as THE safety boundary for runtime evidence, distinct from sqlguard's lint-only role.
- **sources:** contextkit/internal/diagnose/queryselect.go#header, contextkit/internal/diagnose/queryselect_test.go
- **relations:** docselect.go, sqlguard.go, dbcontext (executes the chosen queries)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### sqlguard — IsReadOnlySQL lint (sqlguard.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "READ THIS FIRST — what this is NOT: This is NOT the safety boundary... The REAL guarantee is the EXECUTION SUBSTRATE" (header); three dedicated test files including a literal-false-positive regression (sqlguard_literal_test.go)
- **what:** A cheap pre-flight lint for model-written diagnosis queries, explicitly documented as defence-in-depth, NOT the safety guarantee — the real guarantee is the execution substrate (chassis: read-only transaction + non-multi-statement protocol; harness: a read-only DB role) plus a statement_timeout. Includes a regression fix for keywords/`;` appearing inside quoted string literals (triggered by a real page slug `tool-drop-rate-simulator` containing "drop").
- **sources:** contextkit/internal/diagnose/sqlguard.go#header, contextkit/internal/diagnose/sqlguard_literal_test.go#header
- **relations:** queryselect.go (the actual safety boundary via hand-vetted catalogue), dbcontext
- **verify-later:** confirm the chassis execution path really does use `db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})` as claimed, and that the harness's `-psql` role is genuinely read-only in practice

<!-- SOURCE: U17b_docs019_gofiles.md -->
### BundleGatherer (gatherer.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "READ-ONLY by construction (DESIGN §4): bundle runs dbcontext... + the pure assembler. Nothing here triggers a build, spawn, or write." (header)
- **what:** A Gatherer that shells out to `cmd/bundle` to produce each iteration's bundle, translating a `Scope` into bundle flags and returning the written bundle path. Adds no capability beyond what `cmd/bundle` already does — just drives it per iteration with the loop's evolving scope.
- **sources:** contextkit/internal/diagnose/gatherer.go#header
- **relations:** cmd/bundle, cmd/diagnose (wires this in as the real gatherer)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### ranked-candidate contract (internal/candidates)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Defined once here so the shape isn't re-declared as `candFile`/`jc` in each tool." (header)
- **what:** The shared `Candidate`/`File` JSON contract (`path`, `name`, `kind`, `score` as float64, `rank`, `task`, `method`) that resolve_targets, embed, and fuse all emit with `-json`, and that fuse and eval_targets read — replacing what used to be duplicated per-tool struct definitions.
- **sources:** contextkit/internal/candidates/types.go#header
- **relations:** resolve_targets, embed, fuse, eval_targets
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### ground-truth eval set for target resolution
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** task `silent-norebuild-resultspec` is tagged "REAL fix (2026-01 chassis regression): coordinator honoured only plural output_fields; writer declares singular output_field; compiled page collapsed into the skipPatterns dump -> 'completed' stub -> stale page. Fix = resolveResultSpec treats singular as flatten."
- **what:** `groundtruth_targets.json` maps tasks (deliberately symptom-only task strings, scrubbed of the vocabulary a resolver could trivially match — see the note on an earlier version leaking "extracted"/"result" and inflating a lexical rank) to the "expect" (decisive) and "also_useful" symbols a resolver must surface. Grew across versions: the `.orig` predecessor holds only the `skinner-box` task; the current file adds `silent-norebuild-resultspec`, drawn from a real, already-fixed chassis regression (result-spec singular vs plural output_field handling).
- **sources:** contextkit/groundtruth_targets.json, contextkit/groundtruth_targets.json.orig
- **relations:** eval_targets, resolve_targets/embed/fuse (evaluated against this set), llm-quality-testing
- **verify-later:** platform code for `result_spec.go:resolveResultSpec` / `coordinator.go:extractWorkflowResult` — confirm the fix described is actually live

<!-- SOURCE: U17b_docs019_gofiles.md -->
### code-indexer agent (analyser-adapter's chassis-side counterpart)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** "DRAFT — modelled on the real agent_definitions rows you sent... Confirm the live schema before applying" (NNN_create_code_indexer_agent.sql); status column set to `'experimental'` in the INSERT itself
- **what:** A draft `agent_definitions` row for a `code-indexer` orchestrator agent: workflow is `request_analysis` (calls `request_repo_analysis` action, asking the analyser adapter to parse a repo@ref into symbols) → `index_symbols` (calls `index_code_symbols`, upserting into `code_symbols`, embedding changed symbols via an ollama/nomic-embed-text endpoint, pruning symbols absent from the commit) → `complete`. Retrieval side is a separate `lookup_code_symbols` action used by other agents. Coordination-only orchestrator; the real parsing work happens in the analyser-adapter pod.
- **sources:** NNN_create_code_indexer_agent.sql
- **relations:** analyser (the parsing primitive this indexes), embed (same embedder pairing: ollama + nomic-embed-text), analyser-adapter deployment plan, snapshot-before-mutate practice
- **verify-later:** `agent_definitions` table (`\d agent_definitions`) for the real CHECK constraint on `agent_category` and NOT NULL/default columns before this migration is applied; whether `code_symbols`, `index_code_symbols`, `lookup_code_symbols`, `request_repo_analysis` exist yet

<!-- SOURCE: U17b_docs019_gofiles.md -->
### action-name-to-file resolver (bundle_recreation_v1.sh)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "v1 grepped only for the quoted name inside action files and excluded registry.go — which is precisely where the mapping lives. File names are not consistent either: validate_page_content.go has no _action suffix." (header)
- **what:** A path-resolution helper that, given a registered action NAME, finds its source FILE by (1) reading the registration line in `registry.go` to get the constructor/type name, (2) finding the file defining that constructor/type, (3) falling back to CamelCasing the action name and searching, (4) last-resort whole-platform-tree search — built specifically because file naming is inconsistent (some action files lack the `_action` suffix) and a prior version's naive grep missed paths by excluding the one file (`registry.go`) where the authoritative name→type mapping actually lives.
- **sources:** contextkit/bundle_recreation_v1.sh#header
- **relations:** bundle, resolve_targets (a cruder, deterministic alternative to lexical/semantic resolution for a KNOWN action name)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### dogfooding bundle for building the diagnosis loop itself (bundle_diagnosis_loop.sh)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** "CONFIRM BEFORE RUNNING (flagged — I could not verify these from the mounted files; only the contextkit engine .go files were available)... the four diagnose actions are DRAFTS (chassis-drafts/). If they are not yet committed to ~/projects/agentchassis AND re-analysed into chassis_clean.json, cmd/bundle will SKIP those -scope entries" (header)
- **what:** A read-only bundle recipe whose SUBJECT is the diagnosis loop's own code (its decisive symbols + the four diagnose actions + governing docs + the constitution), for continuing the loop's own gated build in a fresh chat/sub-agent without re-reading the whole tree — a self-referential use of the tool it is building context about. Self-flags an unverified assumption: the four action files may only exist as drafts not yet analysed into the chassis index.
- **sources:** contextkit/bundle_diagnosis_loop.sh#header
- **relations:** diagnosis-loop scaffold, bundle, cmd/diagnose
- **verify-later:** whether the "four diagnose actions" referenced are now committed to agentchassis proper (outside chassis-drafts/)

## Proposed NEW categories
None — all 30 concepts fit existing taxonomy slugs: `diagnosis-loop` (23), `documentation-system` (5), `adapters` (1), `database-and-infrastructure` (1), `content-governance` (1), `development-guide` (1). (Counts overlap because some concepts touch two slugs; each was filed under its single best-fit home per the tagging rules.)

<!-- SOURCE: U18_sql_for_agents.md -->
### Diagnosis loop agents (diagnose-orchestrator / diagnose-agent)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** 122 seeds both as 'experimental' ("until the real-bug evaluation gate passes; promote to 'active' after"); 126–129 wire persistence and subject threading; incidents in 127/128 show live runs 2026-07-06/10.
- **what:** Read-only diagnosis: hypothesise → gather scoped evidence (code + runtime) → cite-or-abstain verdict → re-scope by following evidence; emits a diagnosis + evidence trail for a human, never changes code. Loop CONTROL lives in the Go engine (diagnose_run), not workflow conditionals; gather steps stay explicit for log visibility. Wrapper-mandated (substantive in-chassis LLM work). Runtime evidence is an optional bundle tier — error routing makes anchorless (code-only) runs survive.
- **sources:** 122_diagnose_agents.sql; 126_wire_persist_diagnosis_note.sql; 127_diagnose_load_runtime_error_step.sql; 129_wire_diagnosis_subject_threading.sql
- **relations:** code-indexer supplies code_symbols retrieval; travelling docs receive diagnosis notes; docs019 diagnosis programme
- **verify-later:** diagnose_run engine; promotion to active; evaluation gate results

<!-- SOURCE: U18_sql_for_agents.md -->
### code-indexer (repo → code_symbols for the analyser)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** 118 marked DRAFT/experimental but with a checked label convention "[checked 2026-06-11: composition implemented in IndexCodeSymbolsAction]".
- **what:** Orchestrator that asks the analyser adapter to parse a repo at a ref into symbols, then index_code_symbols upserts them (embedding changed symbols, pruning absent ones). repo label is composed as "owner/repo" from the analyser reply so labels always match what was fetched; retrieval side is lookup_code_symbols used by diagnosis agents. Non-git corpora may override repo (e.g. 'domain:kruste.com').
- **sources:** 118_code_indexer_for_analyser.sql
- **relations:** diagnose-agent evidence gathering; analyser adapter; docs019 contextkit
- **verify-later:** code_symbols table; agent status live

<!-- SOURCE: U19_sql_tables_components.md -->
### code_symbols per-repo code index (context tool)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "CONFIRMED (2026-06-09, clients_db): \dt found no existing *code*/*symbol* table, and \dx shows vector 0.8.0... Both gates pass — HNSW stands."
- **what:** The context tool's code index: one row per symbol keyed (repo, path, symbol) with kind CHECK (func/method/struct/interface/alias/type/var/const), signature/doc/line range (bodies read from the repo at commit_sha, not stored), content text that is both embedded (HNSW cosine, chosen over IVFFlat for incremental churn) and trigram-matched, content_hash to skip re-embedding unchanged symbols. Deliberate departures flagged: no version/soft-delete — a rebuildable cache versioned by commit_sha, pruned by hard delete. Ships the full usage contract in comments: indexing upsert, prune, semantic/lexical retrieval, and hybrid RRF fusion in SQL (constant 60) replacing in-Go fuse.
- **sources:** docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql
- **relations:** knowledge_base shape reuse; diagnosis-loop code retrieval; contextkit.
- **verify-later:** indexing workflow; code_symbols row counts per repo.

<!-- SOURCE: U23_docs_root_vonc.md -->
### cmd/bundle context-assembly harness (contextkit)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "bundle_minilobby_trim.sh (v4) completed: bundle written to /tmp/bundle_minilobby_trim.md" (2026-07-09) after four documented failures.
- **what:** A read-only Go tool (in the contextkit tree, a SEPARATE Go module) that assembles a decision bundle for an LLM verdict: required -analysis/-root/-constitution/-task, repeatable -scope path[:Symbol]/-include/-doc, DB gathers via -psql (-schema-tables, -runtime-site/-page), -dry-run. Operational lessons made durable: resolve an action's file from the REGISTRY (key → Handler: symbol → function definition), never from header-comment conventions; scope a dedicated <key>_action.go file WHOLE but a shared file BY SYMBOL (attention dilution); run from inside contextkit with absolute -analysis/-constitution/-doc/-out and root-relative -scope; prefer the authored runbook's invocation over an example's shorthand. Used here to settle the sanctioned template-edit path before touching anything.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-bundle-v1→v4 (four entries); docs/bundle_minilobby_trim(4).sh (header); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4.0
- **relations:** cmd/diagnose harness; sanctioned edit paths (the question it settles)
- **verify-later:** docs/agent_docs/docs019.../go_files/contextkit/cmd/bundle; RUNBOOK_thin_slice invocation form

<!-- SOURCE: U23_docs_root_vonc.md -->
### cmd/diagnose read-only diagnosis harness
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_vonc_write_site_spec: runtime evidence "read read-only via the diagnosis harness's runtime gather"; full re-run command given with -seed-hypothesis/-dry-bundle.
- **what:** The diagnosis loop entry point: `go run ./cmd/diagnose` with -analysis (callgraph json), -constitution, -psql (read-only runtime gather against the cluster DB: agent_error_log, site_work_items), -seed-hypothesis/-seed-scope, -runtime-site/-page, producing per-iteration bundles (/tmp/diag_bundle_N.md, bundle-<id>/runtime.md); a -verdict-script drives the loop, the stub abstains without a model. The write_site_spec handoff shows the intended usage pattern: harness gathers evidence, a fresh session re-scopes and reads the real code.
- **sources:** docs/HANDOFF_vonc_write_site_spec_spec_data.md#how-to-get-the-evidence
- **relations:** cmd/bundle; fix-loop council (later consumers)
- **verify-later:** cmd/diagnose flags vs docs019 contextkit docs

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### backend_unreachable discovery check
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(f) "backend_unreachable REWRITTEN against the real DiscoveryCheck interface … Run(dctx DiscoveryCheckContext)(*CheckResult,error) … gofmt-clean"; enable pending.
- **what:** A discovery_checks check that NOOPs unless deploy_config.target='vm', GETs each backend site's public `/health`, and on failure emits a site_work_items row (source='discovery', item_type='backend_unreachable', item_key for dedup). Self-clearing. Alert-only: HandlerAgent "" because a down VM isn't chassis-fixable (the P5 vmhost adapter becomes the handler later). A `missing_beacon` check was floated too.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f, traffic_probe_plan(11).md#p4
- **relations:** ties to P5 vmhost adapter as future handler
- **verify-later:** discovery_checks/check_backend_unreachable.go; site_work_items idx_swi_dedup

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### contextkit target-resolution & bundle-assembly toolchain
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK_thin_slice(10).md: "The thin slice has been reasoned about but not yet run on a real task and checked." README(1).md's directory map shows a standalone `contextkit` Go module (7 commands) plus a `chassis-drafts/analyser-adapter` staging tree destined for the real repo — and that adapter is confirmed **built and deployed**: `internal/adapters/analyser/`, `cmd/analyser-adapter/`, and a live `deployments/kustomize/services/analyser-adapter/` overlay all exist in the working tree.
- **what:** A seven-command Go pipeline for assembling task-scoped LLM context bundles from a codebase: `analyser` (AST walk → JSON structural summary of package/imports/functions/types with line ranges), `resolve_targets` (deterministic lexical-overlap baseline that proposes scope candidates), `embed` (semantic vector index over symbols, Ollama-backed with a non-semantic offline stand-in for pipeline-proving), `fuse` (reciprocal-rank fusion of lexical + semantic candidate lists, k=60), `eval_targets` (recall@N / MRR scorer against a hand-authored `groundtruth_targets.json`), `assembler` (renders the final paste-ready bundle: constitution + task + in-scope code in full + neighbourhood signatures + schema + a "what was left out" pointer note), and `dbcontext` (shells out to psql for schema/rows/runtime-evidence with multipass row sizing — never an unbounded dump). Two contracts (`internal/analysis`, `internal/candidates`) are defined once and shared across commands.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/README(1).md, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/RUNBOOK_thin_slice(10).md, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/{analyser(2).go,assembler(2).go,dbcontext(1).go,go_files/embed.go,go_files/resolve_targets.go,go_files/fuse.go,go_files/eval_targets.go}
- **relations:** flat-file constitution (below); reuse-check retrieval pipeline design (below); adapter response-envelope contract (below, the chassis-integration half); fix-loop council (docs024 fixloop_eg_dartsonline)
- **verify-later:** internal/adapters/analyser/, cmd/analyser-adapter/, deployments/kustomize/services/analyser-adapter/ — confirm whether the standalone contextkit CLI tools themselves (analyser/assembler/embed/fuse/eval_targets/dbcontext binaries) were ever run on a real task per the runbook's "first real run" checklist, or whether only the adapter integration shipped.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### deploy-mechanics taxonomy (six ways a change ships)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** GUIDE_deploy_from_context_packs.md is written as an operational runbook with concrete `make`/`kubectl`/`kcat` commands and per-project worked examples (gamesdesign adoption, thunder checkpoint race, idea.uk go-live) dated against real image tags (`v1.0.1057`).
- **what:** A named taxonomy of the distinct ways a change actually reaches production in this system, used to scope any task before touching it: (A) chassis platform image — Go code changes need a rebuilt/pushed/retagged image and a k8s rollout; (B) database — SQL/migrations via psql, snapshot-first, re-query to verify; (C) work-items — insert a `site_work_items` row for the dispatch loop to claim; (D) orchestration trigger — a kcat `orchestrate` message to `system.agent.generic.requests`; (E) generated static sites — downstream/automatic via git → GitHub Actions → Backblaze once `build_status='deployed'`; (F) the idea.uk binary — a separate non-k8s, file-based Go binary with its own build/scp/restart cycle. Cross-cutting cautions: bump image tags or a rollout won't pull the change; "complete" is not "succeeded" — verify positive evidence, not just terminal status.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/GUIDE_deploy_from_context_packs.md
- **relations:** contextkit toolchain (above); deployment-github
- **verify-later:** Makefile targets referenced (build-*, deploy-agents, update-kustomization-images)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### reuse-check retrieval pipeline design (catalog → lexical/structural → embeddings → rerank)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** Directly implemented by the contextkit toolchain's `resolve_targets`/`embed`/`fuse` commands (see contextkit concept above, status partial/unrun); the design principles themselves are framed as reusable lenses.
- **what:** A layered design for "has this already been solved" reuse-checking that treats it as a retrieval problem with a judgement tail, not a generation problem: a maintained capability catalog is the cheapest check (lookup, not search); "identical" (token/AST fingerprinting, algorithmic, high-precision) and "similar" (semantic, embeddings) are split into different mechanisms because lexical/structural matching misses genuine near-duplicates with different names; every narrowing layer is tuned for **recall over precision** since a false-negative reuse check manufactures confident duplication that's worse than no check at all; a cheap model narrows the candidate set, a strong model decides on the shortlist — never the reverse; near-duplicate detection runs post-generation against a concrete draft (a real artifact to fingerprint), while fuzzy "what's there to build on" retrieval runs pre-generation. Signature+docstring embeddings are framed as a general retrieval substrate (also serving target resolution and capability-catalog curation), not a narrow dedup optimisation, and — at the scale of a few thousand symbols — need no vector database, just in-memory cosine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Reuse-checking; contextkit toolchain go files (embed.go, fuse.go)
- **relations:** contextkit toolchain (above); change-layer integration contract (above, reuse_index_refresh trigger)
- **verify-later:** whether any capability catalog or reuse index has been built beyond the contextkit prototype

<!-- SOURCE: U25_leopardess_social.md -->
### Read-only code bundle to settle method before editing (bundle verdict practice)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** VERDICT (2026-07-09) is the artifact: five method questions answered from a contextkit bundle + scoped Go reads before any write; "nothing is edited until its verdict lands".
- **what:** When the supported path for a change is genuinely unclear ("that is a code question → a bundle"), assemble a read-only context bundle (cmd/bundle + contextkit: dbcontext \d + capped SELECTs, scoped Go sources by symbol, relevant docs) and produce a written VERDICT answering the deciding questions before touching anything. The mini-lobby verdict overturned the handoff's premise (remove_element cannot reach the component — its header oversold it) and discovered the rendered-artifact template model. Operational notes: contextkit is a separate Go module; -scope/-include relative to -root; pure assembler triggers nothing.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md; docs/social001_vonc_tiktok_social/minilobby_task/bundle_minilobby_trim(4).sh (header); docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#4.0
- **relations:** diagnosis-loop (docs019 contextkit); operator discipline; section-editor
- **verify-later:** cmd/bundle + contextkit module

<!-- SOURCE: U03_idea_uk_section_data.md -->
### cmd/bundle read-only context composer
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Used for bundles 1–3 and the differentiators handoff; notes (Sa) record its failure modes ("-step framing emitting structure rather than source — use -step debug"; -doc paths must be ls-verified).
- **what:** The investigation tooling that assembled evidence bundles for this thread: `go run ./cmd/bundle` with an analysis JSON, `-root`, `-constitution`, `-step debug|framing|implementation`, `-task` (one-sentence brief), `-scope file[:Symbol]` code selections, `-include`, `-doc` paths, `-psql` connection command, `-schema-tables`, `-runtime-site`/`-runtime-page` live evidence, `-out`. Operational lore: `-step framing` yields signatures only; doc paths silently fail if wrong; bundles can arrive as thin slices (runtime data excluded) so live queries still need running separately.
- **sources:** 001_bundling_context.md; bundle3; RUNBOOK_scheme_to_components(50).md#Bundle-command; running_notes_scheme_to_components(55).md#Sa #Sh
- **relations:** docs019 contextkit (its home); check-based investigation method.
- **verify-later:** cmd/bundle source under docs019 go_files/contextkit; flag semantics.

<!-- SOURCE: U05_content_quality_linking.md -->
### Context packaging + code-bundle tooling for fresh chats
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §11 gives working bundle invocations + "Known cmd/bundle run errors (to fix before regenerating)".
- **what:** Two generations of context assembly for hand-off to a fresh assistant chat: (1) the module packager shell script (package_content_quality_debug.sh) bundling a code slice + docs + an optional read-only live SQL capture into one context file; (2) contextkit `cmd/bundle`, which expands from named symbol scopes via the call graph with -doc/-schema-tables/-runtime flags and a constitution. The 001_context files record filled-in bundle invocations per defect thread (phantom-CTA, clobber). Known operational failures documented: unquoted parentheses in -doc paths break bash; empty /tmp/analysis_repo.json breaks analysis load; bundle can't reach session docs outside the repo.
- **sources:** package_module/package_content_quality_debug(3).sh (header); game_lost_its_tool/001_context; phantom_hero_ctas/001_context; HANDOFF_page_pipeline(11).md#11
- **relations:** docs019 contextkit; division-of-labour operating model; documentation system.
- **verify-later:** contextkit cmd/bundle; contextkit_bundle_issues.md.

<!-- SOURCE: U06_finetuning.md -->
### Docubundle context packager (thunder-checkpoint-race package)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** docubundle/README.md usage note + the generated 633KB production context file dated in-tree; packager header "Patterned on package_page_build_debug.sh".
- **what:** A self-contained packager script (`package_thunder_checkpoint_race.sh`) that bundles the async-await + loop machinery of the chassis, the checkpoint-upload path, the working docs, and optionally a read-only live capture (schemas, decisive queries, workflows, runtime state) into one context file to seed a fresh AI-assistant thread on a specific blocker. Paired with hand-written CONTEXT_PACK / NEXT_CHAT_MANIFEST docs that state the blocker, the verified root cause, the applied fix, and next actions. An instance of the wider bundle/context-package pattern (cf. docs019 contextkit) applied to the finetuning workstream; the targeted CHASSIS_await_loop_extract ("use the targeted extract, not the 72k-line file") shows deliberate context-size curation.
- **sources:** working/docubundle/README.md; working/docubundle/package_thunder_checkpoint_race.sh (header); working/phase5/NEXT_CHAT_MANIFEST.md; working/phase5/CHASSIS_await_loop_extract.txt (header)
- **relations:** diagnosis-loop bundles/contextkit; send-before-register race (its subject)
- **verify-later:** relation to z_bundles/context_packages tooling at repo root

<!-- SOURCE: U08_travelling_docs.md -->
### persist_diagnosis_note — skip-don't-guess subject gate; dead ends persisted
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Stage 3a CLOSED 2026-07-06 (skip gate proven ×3); Stage 3b CLOSED 2026-07-06/07 — first machine-written NOTES row `('pipeline','build')`, categories `["diagnosis","unconfirmed-diagnosis"]`, stop reason `scope-not-narrowing`.
- **what:** A config-gated step after `diagnose_emit` (emit stays read-only by its own design) that persists the diagnosis as a NOTES entry ONLY when the run carries an explicit subject in input_data — skip, never guess (a mis-filed note poisons history; the gate is the action's first check, before any DB access). UNVERIFIABLE verdicts are persisted too, tagged `unconfirmed-diagnosis`, so dead ends stop retries. First payoff on record: the machine-written note itself answered the open "why did the run finish fast" question.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-3,#§4; 0NN_wire_persist_diagnosis_note.sql; RUNNING_NOTES_travelling_docs(39).md#rev3,#rev20,#rev21
- **relations:** subject threading (3b); anchorless-diagnosis degrade; 037 pipeline-integration vision (realised).
- **verify-later:** diagnose-agent workflow `emit → persist_note → complete`; `persist_diagnosis_note_action.go` subject gate.

<!-- SOURCE: U08_travelling_docs.md -->
### Diagnosis subject threading through orchestrator input_mapping + both contracts
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** 3b.2 APPLIED + 3b.3 VERIFIED 2026-07-06 (map paths `input_data.subject_type`/`subject_key`; both contracts t/t).
- **what:** For a spawned child to receive optional fields, the mapping must satisfy the callee's input_contract — so threading `subject_type?`/`subject_key?` took THREE edits (orchestrator input_mapping merge + `optional` additions on BOTH diagnose-orchestrator and diagnose-agent contracts), not two. DB-only, effective immediately. Establishes the general spawn+call contract rule: an input the workflow depends on must be declared.
- **sources:** RUNBOOK_travelling_docs(38).md#3b; RUNNING_NOTES_travelling_docs(39).md#rev17; HANDOFF_2026-07-08…md#§2
- **relations:** spawn+call input-shape pattern (016b); dangling-doc rule (same "declare your inputs" class — migration 137's `spec` declaration).
- **verify-later:** diagnose-orchestrator `call_diagnoser.input_mapping`; both input_contracts.

<!-- SOURCE: U08_travelling_docs.md -->
### Anchorless (code-only) diagnosis degrade at load_runtime
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Corrective APPLIED 2026-07-06; fired ×5 per anchorless run ("NORMAL, not a fault"); softening (`skipped:true` return) still a chassis-build follow-up.
- **what:** Runtime evidence is an optional bundle tier, but `diagnose_load_runtime` hard-errored with no site/correlation/domain anchor and had no error routing — making the tier mandatory in practice and killing legitimate code-only diagnosis runs. Fixed by config-level error_step on load_runtime targeting its own next_step (`assemble_bundle`); since `route.gather_step` re-enters load_runtime every iteration, each loop-back degrades per-iteration to a code+schema bundle. Cost of a full anchorless loop: ≈26 min, 5 iterations.
- **sources:** 016b_debugging_guide_7_3_(7).md#anchorless-entry; RUNNING_NOTES_travelling_docs(39).md#rev11,#rev12,#rev14; 084_TRIGGER_diagnose_v1(2).sh (ANCHOR NOTE)
- **relations:** error_step mechanics; diagnosis loop step map (analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict → route → emit → persist_note → complete).
- **verify-later:** `diagnose_load_runtime` no-anchor softening (shipped or not); load_runtime.config.error_step live value.

<!-- SOURCE: U08_travelling_docs.md -->
### Verdict symptom-coverage gate (symptom_check) on the diagnose-agent
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** FYI 2026-07-10: prompt rule 8 + `symptom_check` schema field applied (snapshot 34f4afc8); engine coercion rides the next chassis image post-v1.0.1101; F0.6 addendum adds `cites`/`context` members.
- **what:** A CONFIRMED verdict must account for every distinct observation of the ORIGINAL symptom via `symptom_check: [{observation, explained, how, cites, context}]`; the chassis engine (`pkg/diagnose`) coerces to UNVERIFIABLE any CONFIRMED verdict whose symptom_check is missing, carries an unexplained entry, or marks explained without a valid citation index; comparative/background clauses are exempted as `context` rather than grade-inflated. Terminal diagnosis notes gain a "Symptom coverage:" block. Owned by the fix-loop workstream; delivered to this unit as a courtesy collision-rule FYI.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md (whole)
- **relations:** persist_diagnosis_note (note bodies change); fix-loop council/verdict work (fixloop_eg_dartsonline docs).
- **verify-later:** diagnose-agent verdict prompt_template; `verdict_wire.go` symptom_check parsing.

<!-- SOURCE: U08_travelling_docs.md -->
### Context-bundle command for cross-chat handoffs (cmd/bundle + registry-based scope resolution)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Two bundles built and used (07-08 toolgen bug; 07-09 recreation); resolver rewritten after 3 misses (rev 44).
- **what:** `cmd/bundle` renders a task bundle (constitution + task text + code scopes + docs + live schema + runtime evidence incl. an agent_error_log "Recent errors" section — the section that settled the 07-08 diagnosis). Path facts banked: resolve actions via registry.go (action name → constructor → defining file), not filename convention (`execute_llm_prompt` lives in ai_actions.go; validate_page_content.go lacks the _action suffix); misses are non-fatal and print grep candidates.
- **sources:** bundle_recreation_v1(1).sh (header + resolve_action); HANDOFF_2026-07-08…md#§6; RUNNING_NOTES_travelling_docs(39).md#rev44
- **relations:** docs019 contextkit/bundles; agent_error_log first read.
- **verify-later:** cmd/bundle flags; whether the runtime errors section is standard.

<!-- SOURCE: U09_adoption.md -->
### Docubundle context-pack tooling (module packager + dbcontext + deploy guide)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Working scripts and generated 1.6MB context files present (package_page_build_debug.sh + output_contexts); GUIDE_deploy_from_context_packs documents the operating loop across four live packs.
- **what:** Tooling for packaging a subsystem's blast-radius into one AI-consumable context file: `package_page_build_debug.sh` (self-contained packager bundling the page-build/section-resolution/render-deploy/dispatch code, keeping the reuse-discovery catalogue layers — registry.go, datahelpers, input_contracts — plus an optional read-only live capture: schema, decisive queries, agent-def workflows, runtime state); `dbcontext` Go CLI (shells out to configurable psql; `\d` schema dumps and multipass-sized row fetches — never an unbounded dump); and the GUIDE's general loop (attach pack → gather live context → verify the decisive fact → work → deploy via mechanism A–F → verify with positive evidence). Deploy mechanisms taxonomy: A chassis image, B database/migrations, C work-items, D orchestrate-message triggers, E generated static sites (git→Actions→B2), F idea.uk binary.
- **sources:** docubundle/GUIDE_deploy_from_context_packs.md, docubundle/dbcontext.go header, docubundle/package_module/package_page_build_debug.sh header, CONTEXT_PACK_adoption_skinner_box.md
- **relations:** context packs (CONTEXT_PACK_* docs); docs019 bundles/contextkit; thin-slice constitution (included in every bundle)
- **verify-later:** whether packagers/dbcontext live in the repo proper or only in docs; output_contexts freshness

<!-- SOURCE: U09_adoption.md -->
### Adoption context pack (skinner-box) as a worked fresh-thread starter
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** CONTEXT_PACK_adoption_skinner_box.md exists and was consumed by the 2026-06-08 session that closed the bug (running_notes_15 Part 1 resolves the pack's named "decisive fork").
- **what:** A structured resume pack for one open bug: state + next action, the decisive fork to verify first, standing rules (constitution excerpt), code to pull fresh vs re-attach, schema/rows/runtime capture commands, and the minimum fast-start set. Demonstrates the pack contract: packs restate earlier context and inherit its staleness — the fresh pull is the source of truth (the pack's own causal story about the content-writer was corrected by the session).
- **sources:** CONTEXT_PACK_adoption_skinner_box.md, NEXT_CHAT_INPUTS_2026-06-06.md, running_notes_15(10)#part-1
- **relations:** docubundle tooling; GUIDE_deploy_from_context_packs per-project quick reference
- **verify-later:** n/a (artifact of method)

<!-- SOURCE: U12_docs024_archives.md -->
### `error_step`: config-level placement requirement + derive-from-next_step fix pattern
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the routing FIRES (error_routed in ProcessingHistory)... Live-validated ×5 in one run."
- **what:** The chassis workflow coordinator only consults `step.Config["error_step"]` (config-level); a step-level `error_step` is parsed but never read, so placing it outside `config` is silently inert. Fix pattern: derive `error_step` from the step's own `next_step`. This entry and its three siblings below are genuinely absent from the canonical live `016b_debugging_guide_8_consolidated.md`/`merged(1).md` — they continue only in a parallel `travelling_docs/016b_debugging_guide_7_3_(2..7).md` fork the canonical consolidation's "verified against ALL forks" claim did not actually reconcile.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"error_step: config-level placement..."
- **relations:** dormant instances of the buggy shape found still live in `tool-recreation-handler` and `tool-auditor` agent definitions
- **verify-later:** grep `agent_definitions` for step-level `error_step` occurrences in those two agents.

<!-- SOURCE: U12_docs024_archives.md -->
### Anchorless (code-only) diagnosis dies at load_runtime
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "Fix. Config-level error_step on load_runtime... Live-validated ×5" (deployed) but "Pending softening (next chassis build)" (aspirational remainder).
- **what:** A diagnosis run with no anchor was treated as optional by bundle-assembly but hard-errored the whole child workflow at `load_runtime`. Interim fix routes the error back to its own `next_step`; a proper code-level softening (treat no-anchor as a skip) was identified but not yet shipped.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Anchorless (code-only) diagnosis..."
- **relations:** sibling of the error_step concept above; also absent from canonical live 016b
- **verify-later:** check `diagnose_load_runtime` action source for the `skipped:true` softening.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Bundle write-through (DiagnoseAssembleBundleAction, F0.1b)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.1b... work end to end in production" (NOTES(10)#Turn 6, F0 plumbing criteria)
- **what:** Each diagnosis iteration's evidence bundle is persisted to `diagnosis_artifacts` from inside the Go action `DiagnoseAssembleBundleAction`, immediately before its existing return — zero workflow-shape change, staying off the tools-chat's active `emit → persist_note → complete` surface. A persistence failure degrades to a logged warning on all paths; it never fails the diagnosis itself, because observability must never cost a diagnosis.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql#design note, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1b, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 3
- **relations:** diagnosis_artifacts table; retention knob
- **verify-later:** DiagnoseAssembleBundleAction source; ON CONFLICT clause used for the write

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Loop-worthiness test (doctrine)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "LOOP-WORTHINESS TEST (doctrine — apply before every intake)" (RUNBOOK(10)#LOOP-WORTHINESS TEST)
- **what:** Five criteria applied before any bug enters the loop: it's a behaviour symptom not a feature request; a causal mechanism plausibly exists across code/data/runtime; it is NOT answerable by one or two direct queries (mandatory cheap pre-check first); it is bounded to one symptom; the symptom is verified current at intake. Three successive candidates were dissolved by criterion 3 on this platform, leading to the empirical conclusion that "bug mechanisms tend to be legible to schema access plus grep" — reframing the workstream's value proposition from discovery to unattended/cited/consistent diagnosis.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#LOOP-WORTHINESS TEST, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#LOOP-WORTHINESS TEST, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §6
- **relations:** known-answer benchmark methodology; abandoned pilot candidates
- **verify-later:** n/a — methodology, not code

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Symptom anchor (F0.4a)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4a... ✅ CODE-COMPLETE 2026-07-09" then verified live in run 2 (PLAN_fixloop_pilot.md §3b, NOTES(10)#Turn 10)
- **what:** The evidence bundle always renders "## Original symptom" above "## Hypothesis under test," restoring visibility of the user's original question once the loop's working hypothesis has drifted from it. Fixes a finding that the verdict never saw the original symptom text after iteration 2 in benchmark run 1.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b
- **relations:** hypothesis drift (engine behaviour); symptom-closure gate (F0.4d)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Follow-the-error-log enrichment (F0.4b)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4b... ✅ CODE-COMPLETE, its SQL verified live" (PLAN_fixloop_pilot.md §3b)
- **what:** Bridges the loop's Go-only static-evidence corpus gap: since `code_symbols` indexes `.go` files only and load-bearing platform logic can live in `agent_definitions.default_config` JSON, this enrichment regexes `agent/step (action)` references out of runtime evidence (agent_error_log lines) and inlines the named workflow step's JSON into the bundle, capped at 8KB. Directly converted the benchmark bug's cause B into cited static evidence.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7, #Turn 10
- **relations:** symptom anchor; workflow-JSON-as-load-bearing-logic gotcha
- **verify-later:** grep/inspect `code_symbols`; `.go`; `agent_definitions.default_config`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Same-file sibling signatures + fair-share budgeting (F0.4c)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4c... ✅ CODE-COMPLETE" then "fair-share worked end to end" (NOTES(10)#Turn 8, #Turn 16)
- **what:** When retrieval scopes a symbol, the bundle also lists the signatures of that file's other functions (capped), fixing the case where symbol-granular retrieval found the right file but the wrong function. Initial implementation starved small files' budget with first-come-first-served ordering; fixed with fair-share-per-file budgeting (`capChars/n`, floor 600) plus a "+N more" affordance.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 8, #Turn 15, #Turn 16
- **relations:** follow-the-error-log enrichment; must-claim-4 blind spot
- **verify-later:** grep/inspect `capChars/n`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tier-coverage guard (F0.4e)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "First production firing of any of the new guards" (NOTES(10)#Turn 12, run 3)
- **what:** A shared `coerceVerdict()` engine gate requiring a CONFIRMED verdict to carry at least one `static` citation AND at least one `state|runtime` citation, or it degrades to Unverifiable. REFUTED is exempt. Directly answers the benchmark run-1 finding that "cite-or-abstain does not prevent confirming the wrong cause."
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7, #Turn 8, #Turn 12
- **relations:** symptom-closure gate; context disposition (F0.6); cite-or-abstain doctrine
- **verify-later:** grep/inspect `coerceVerdict()`; `static`; `state|runtime`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Symptom-closure gate / symptom_check (F0.4d)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.4d — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md §3b)
- **what:** A CONFIRMED verdict must carry a `symptom_check` — mapping each observation in the original symptom to the confirmed mechanism (`explained:true/false` + `how`) — or the engine coerces it to Unverifiable. Motivated by benchmark run 2, where a well-cited confirm dismissed half the symptom as "not a nav issue." The verdict prompt lives in the diagnose-agent workflow JSON, a different workstream's active surface, so the edit was done fetch-first with a snapshot and an FYI filed.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 11, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#BENCHMARK RUN 3
- **relations:** tier-coverage guard; context disposition (F0.6); doc_notes/travelling-docs coordination boundary
- **verify-later:** grep/inspect `symptom_check`; `explained:true/false`; `how`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Context disposition + citation-backed "explained" (F0.6)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.6 — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md §3b)
- **what:** Refines symptom_check with a `context bool` flag (comparative/background clauses exempt from explained/unexplained accounting) and requires `explained:true` entries to carry an in-range `cites` index — an unsupported "explained" is now rejected. Fixes a grade-inflation defect where run 4 marked comparison clauses `explained:true` while their own text said "unverifiable from this bundle."
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 14, #Turn 15
- **relations:** symptom-closure gate; tier-coverage guard
- **verify-later:** grep/inspect `context bool`; `explained:true`; `cites`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### data_request persistence across iterations (F0.5)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "F0.5 — ✅ CODE-COMPLETE 2026-07-10 (from run 3)" (PLAN_fixloop_pilot.md §3b)
- **what:** Fixes a defect where fetched data_request answers evaporated from the bundle after one iteration, tripping the scope-not-narrowing guard. Reuses `LoopState.SeenRequests` by forwarding the UNION of current-verdict and prior-seen request keys (deduped, capped at 12) so `load_runtime` re-runs them every iteration — "re-run, don't store," avoiding the collected_data-bloat class of incident.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 12, #Turn 13
- **relations:** tier-coverage guard; collected_data-bloat gotcha
- **verify-later:** grep/inspect `LoopState.SeenRequests`; `load_runtime`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### contextkit bundle regeneration procedure
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "verified end-to-end 2026-07-09... 306,897 B, 468 files analysed, all three gathers succeeded" (RUNBOOK(10)#REGENERATING THE CONTEXT BUNDLE)
- **what:** The documented, tested procedure for regenerating a human/chat-facing evidence bundle via the `contextkit` CLI (a separate Go module, not the live loop's in-cluster assembler): analyser with excludes, then bundle with `-psql` as ONE quoted argument and `-schema-tables` including the tables relevant to the bug. This bundle is for humans; the live loop's own retrieval is a separate, in-process mechanism.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#REGENERATING THE CONTEXT BUNDLE, fixloop_eg_dartsonline/HANDOFF_fixloop_thread(8).md#CODE CONTEXT
- **relations:** blinding discipline
- **verify-later:** grep/inspect `contextkit`; `-psql`; `-schema-tables`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Corpus gap: workflow JSON invisible to the static tier
- **category:** diagnosis-loop
- **status-signal:** deployed (documented gotcha, partially mitigated by F0.4b)
- **status-evidence:** "code_symbols indexes .go files only. Workflow definitions live in agent_definitions.default_config as JSON and are therefore INVISIBLE to the loop's static tier" (RUNBOOK(10)#Inherited gotchas)
- **what:** The diagnosis loop's static evidence tier is built entirely from indexed Go source; workflow definitions stored as JSON in `agent_definitions.default_config` — which frequently contain the actual load-bearing control flow — are structurally invisible to it. Partially mitigated by the follow-the-error-log enrichment (F0.4b); no general mechanism exists for the static tier to discover workflow-JSON logic it hasn't been pointed at.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 7
- **relations:** follow-the-error-log enrichment (F0.4b); dartsonline guides defect
- **verify-later:** grep/inspect `agent_definitions.default_config`

<!-- SOURCE: U14_docs019_runbooks.md -->
### contextkit — task-scoped codebase bundle toolkit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Done: call-graph neighbourhood; live schema + row data via dbcontext …; the cmd/bundle orchestration wrapper"; RUNBOOK(31) header "This project builds contextkit … developed against, and dogfooded on, the agent-chassis repository" (2026-06-24).
- **what:** A small Go module (`contextkit/`: cmd/analyser, assembler, embed, dbcontext, resolve_targets, fuse, eval_targets, bundle, diagnose) that assembles a tightly-scoped slice of a codebase — the in-scope source in full, its call-graph neighbourhood as signatures, DB schema, runtime evidence, and authored guidance/constitution — into one paste-ready "bundle" per task. Two shared contracts (`internal/analysis`, `internal/candidates`) defined once, no per-tool copies. The deployed chassis diagnosis agent is its descendant; the CLI remains the dev/eval harness.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#the-pipeline; docs019/RUNBOOK(31)_diagnosis_loop.md#what-this-is; docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** read-only diagnosis loop; bundle altitudes; dbcontext; cmd/bundle wrapper
- **verify-later:** `docs019/go_files/contextkit/` module; `$CK/cmd/*`; `internal/analysis`, `internal/candidates`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Read-only cite-or-abstain diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) checklist "§6G eval gate PASSED — run 51f95cda (2026-07-01): abstain → correct reads → REFUTE the naive framing → CONFIRM the grounded cause"; code_retrieval_route(21) "§7 ROUTE CLOSED — 2026-07-03 (run 73ed55c6)".
- **what:** An AI agent that investigates a bug strictly READ-ONLY: forms a hypothesis, gathers scoped evidence (code bodies + read-only DB rows + runtime records), issues a verdict that must CITE evidence or ABSTAIN (CONFIRMED/REFUTED/UNVERIFIABLE), then re-scopes by FOLLOWING the evidence (call graph for code, vetted queries for data) rather than re-searching the symptom. Never edits code, never runs builds, human-gated; the hard problem it targets is falsification — abandoning a wrong hypothesis.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#what-this-is; docs019/RUNBOOK_design_diagnosis_loop(7).md#overview; docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** convergence guards; verdict wire format; three-tier citation; falsification-first eval gate; diagnosis→fix loop (v2)
- **verify-later:** chassis `pkg/diagnose/` (loop.go, step.go, advance.go); `platform/orchestration/actions/diagnose_*_action.go`; agent_definitions rows diagnose-agent/diagnose-orchestrator

<!-- SOURCE: U14_docs019_runbooks.md -->
### Bundle step altitudes: framing vs implementation vs debug
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) assembler flags: "-step framing | implementation | debug … framing: in-scope shown as signatures (intent over detail)".
- **what:** A bundle declares its altitude: `framing` shows in-scope code as signatures only (used to expand an under-specified brief into a spec before targets can be picked), `implementation`/`debug` show full bodies, and `debug` adds a runtime-evidence section. Encodes the framing-vs-implementation altitude split as an explicit pipeline parameter.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#assembler-flags; docs019/RUNBOOK_thin_slice(27).md#fuzzy-tasks
- **relations:** contextkit toolkit; reasoning-state handoff
- **verify-later:** `$CK/cmd/assembler/main.go` step handling

<!-- SOURCE: U14_docs019_runbooks.md -->
### Call-graph neighbourhood selection with forced -include for wiring files
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "-include … Closes the blind spot the first adoption run found"; known-limits section "call-graph neighbourhood is name-matched, not type-resolved".
- **what:** The bundle's surrounding context is the call-graph neighbourhood (callees/callers/types) of the in-scope symbols, rendered as signatures, with `-neighbour package` as fallback when name-matching misses (interface dispatch). Registration/wiring files (e.g. registry.go, reached via init not calls) are force-included with `-include`. Ubiquitous names (Run, String, New) are dropped when the loop follows the graph, to avoid scope explosion.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#assembler-flags; docs019/RUNBOOK_design_diagnosis_loop(7).md#design-and-build-choices
- **relations:** named-scope guard vs capped expansion; ReadSymbolBody slicer
- **verify-later:** `internal/analysis/analyse.go` calls extraction; `pkg/diagnose/callgraph.go` ubiquitous-name drop list

<!-- SOURCE: U14_docs019_runbooks.md -->
### dbcontext — bounded read-only DB context gather
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) pipeline step 2 with worked flags; "-rows … multipass sizing (probe LIMIT N+1 …). Never an unbounded dump."
- **what:** CLI that pulls live DB context for a bundle: `-schema` (`\d` per table), `-rows` (SELECT with multipass sizing and a row cap), and `-runtime-site`/`-runtime-page` (recent agent_error_log rows + site_work_items lifecycle as a "Runtime evidence" block). All read-only; queries are appended as `-c` args, not shell-interpolated.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#dbcontext-flags
- **relations:** cmd/bundle wrapper; three-guard read-only SQL model; diagnose_load_runtime
- **verify-later:** `$CK/cmd/dbcontext/`

<!-- SOURCE: U14_docs019_runbooks.md -->
### cmd/bundle orchestration wrapper and the pure-composer boundary
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) design note "(Status: wrapper not yet built — flagged for decision)" superseded in the same file's Done list "the cmd/bundle orchestration wrapper (gather via dbcontext → assemble, composer stays read-only)".
- **what:** The assembler is a PURE COMPOSER — it never runs SQL or chooses tables; `cmd/bundle` is the orchestration wrapper that runs the requested read-only dbcontext gathers and then calls the assembler with the outputs wired in. Keeps query execution inside the bounded read-only tool while offering "one command including the SQL". Automatic table-selection was deliberately deferred and must propose-then-confirm.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#one-command; docs019/RUNBOOK_thin_slice(27).md#assembler-boundary
- **relations:** dbcontext; diagnosis loop gatherer (BundleGatherer shells out to cmd/bundle)
- **verify-later:** `$CK/cmd/bundle/`; `pkg/diagnose/gatherer.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Bundle size doctrine — "a large bundle is a smell, not a goal"
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Context-window facts (verified against the Claude docs, June 2026)"; "aim to keep a working bundle under ~200K tokens (~800 KB)".
- **what:** Working rule for feeding bundles to models: keep under ~200K tokens; context rot means a full 1M window is not used evenly; the fix for an oversized bundle is narrower selection, not a bigger window. Includes the three feeding routes (chat paste, claude.ai Project, API with prompt caching of the stable prefix).
- **sources:** docs019/RUNBOOK_thin_slice(27).md#large-bundles
- **relations:** responses-are-summaries doctrine (Kafka side); call-graph neighbourhood (the narrowing instrument)
- **verify-later:** n/a (doctrine); bundle sizes in diagnosis_artifacts once built

<!-- SOURCE: U14_docs019_runbooks.md -->
### B4a finding — the symptom→infrastructure retrieval ceiling
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "OUTCOME (2026-06-17, 2 ground-truth tasks): skinner-box lexical 0.50, semantic 0.00; resultspec lexical 0.00, semantic 0.00, fused 0.00 … DECISION: embeddings do NOT earn a place in the code path on this evidence".
- **what:** Measured finding that when a bug's cause lives in shared infrastructure named for its FUNCTION rather than its FAILURE MODE, symptom-based code retrieval (lexical, semantic, or fused) has a hard ceiling — symptom words and mechanism words don't intersect, and no embedding closes a zero-overlap gap. Secondary finding: naive RRF fusion can be worse than lexical alone. This is the empirical justification for the diagnosis loop's re-scope-by-following-evidence design.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#B4a; docs019/RUNBOOK_design_diagnosis_loop(7).md#the-empirical-finding
- **relations:** lexical/semantic/fused target resolution; read-only diagnosis loop (the lever pulled instead)
- **verify-later:** `$CK/groundtruth_targets.json`; `docs019/go_files/contextkit/{lex,sem}.json`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Lexical/semantic/fused target resolution (resolve_targets, embed, fuse, eval_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) pipeline 1b–1d with all four tools runnable; B4a decision "lexical (trigram + resolve_targets) carries the spine and embeddings are the tie-breaker".
- **what:** The target-resolution layer: a lexical (trigram) candidate proposer, an Ollama-backed semantic index (nomic-embed-text with search_document/search_query prefixes matching the chassis rag pipeline exactly), RRF rank fusion, and a recall@N/MRR scorer against a ground-truth task set. Built to answer "does semantic beat lexical for code" — the measured answer was no for this corpus.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#the-pipeline; docs019/RUNBOOK_thin_slice(27).md#B4a-task-1
- **relations:** B4a ceiling finding; code_symbols index (production analogue); evidence-fed scope resolver (later reuse of the same vector search)
- **verify-later:** `$CK/cmd/{resolve_targets,embed,fuse,eval_targets}/`; ollama-adapter service

<!-- SOURCE: U14_docs019_runbooks.md -->
### Ground-truth eval harness and its measurement-trap discipline
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "THE TRAP (hit 2026-06-14): resolve_targets was run with a DIFFERENT task … eval then scored … a meaningless 0/2"; the task-string bind guard and `-task-id` requirement.
- **what:** groundtruth_targets.json holds task→expected-symbol pairs; every eval binds the task string once, guards it against the truth file, uses ONE matched index for lexical and semantic, and forbids answer-vocabulary leaks in task wording (a leaked symbol name contaminated the ceiling test once). Three prior B4a attempts failed on METHOD, not result — the harness encodes the corrections.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#B4a-task-1; docs019/RUNBOOK_thin_slice(27).md#B4a-task-2
- **relations:** instrument-skepticism doctrine; B4a ceiling finding
- **verify-later:** `$CK/groundtruth_targets.json`; `$CK/cmd/eval_targets/`

<!-- SOURCE: U14_docs019_runbooks.md -->
### ReadSymbolBody — the single shared symbol-body slicer
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §3 "This collapse is DONE and verified: the merged assembler … diffed against the pre-collapse binary … byte-identical"; checklist "§1 ReadSymbolBody written + unit-tested".
- **what:** One implementation of symbol-body slicing (`analysis.ReadSymbolBody`) placed in BOTH module copies of `internal/analysis` (contextkit and chassis): body = file lines [StartLine, EndLine] inclusive, 1-indexed, exactly as the analyser records; resolves bare names and receiver-qualified `Type.Method`; whole-file for a path with no `:Symbol`. `cmd/assembler`'s duplicate slicing (splitScope/locateSymbol/readLines) was collapsed onto it — "two copies of one convention is the drift this project keeps getting bitten by".
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#1; docs019/RUNBOOK(31)_diagnosis_loop.md#3
- **relations:** diagnose_assemble_bundle; contextkit toolkit; module-copy drift (the two analyse.go copies noted drifted)
- **verify-later:** `internal/analysis/symbolbody.go` in both modules; `symbolbody_test.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose_assemble_bundle action
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) checklist "§2 diagnose_assemble_bundle merged (gofmt-clean)" and "§6C build + register the four diagnose actions … DONE 2026-06-29".
- **what:** The chassis action that, per iteration, reads the in-scope symbols' bodies via ReadSymbolBody from a decoded `repo_analysis` Output, composes hypothesis + code + runtime (+ live schema) into the `bundle` the verdict step reads. Scope fallback chain: `route.scope` (loop-back) → `input_data.seed_scope` → `code_lookup.code_results`. Unknown symbols are logged and skipped, not fatal.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#2; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** ReadSymbolBody; loop_scope_field lesson; diagnosis_artifacts egress (planned write-through here)
- **verify-later:** `platform/orchestration/actions/diagnose_assemble_bundle_action.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Four convergence guards plus engine-level failsafes
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6E.1 "the loop stopped at stopped_by: evidence-not-growing — a guard, not luck. So the guards + the max_iterations cap are armed" (2026-06-29).
- **what:** Deterministic stop conditions independent of model behaviour: iteration-cap, scope-not-narrowing, evidence-not-growing, hypothesis-thrash — plus engine-level `timeout_seconds: 1800` and `fuel_budget: 1000` that bound a runaway even if the loop's bookkeeping is disarmed. Behaviour-tested (26-test suite), not eyeballed; the guards are the safety layer that lets a model verdict be untrusted.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position; docs019/RUNBOOK_design_diagnosis_loop(7).md#0
- **relations:** SeenRequests progress rule; named-scope guard; state threading self-check
- **verify-later:** `pkg/diagnose/loop.go` guards; `loop_test.go`, `step_test.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### SeenRequests — a new data_request counts as loop progress
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "#1 fix … guardAfter now tracks issued read-only data_requests in a SeenRequests set … evidence-not-growing (and hypothesis-thrash) yield when the verdict issues a NEW unseen request"; validated in run 51f95cda ("3 iters, new queries each, no premature stop").
- **what:** Fix for the loop stopping one iteration before its own good query ran: guards treat a NEW unseen read-only data_request as progress (its answer arrives next gather), while a re-issue of the same query still trips the guard. Required the `verdict_wire.go` sync (an older chassis copy silently mapped DataRequests to null, making the engine fix inert).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#1 fix and #1 status)
- **relations:** convergence guards; data_requests channel; verdict wire seam
- **verify-later:** `pkg/diagnose/advance.go` SeenRequests; `loop_datarequest_test.go`; `pkg/diagnose/verdict_wire.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Named-scope guard vs capped call-graph expansion
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "BOTH FIXES DELIVERED 2026-07-03 … Guard now measures the MODEL-NAMED scope …; expansion runs only after the guard passes and is CAPPED (Config.MaxExpandedScope, engine default 18)"; route-close run 73ed55c6 "the expansion cap bounding iterations 2–3 at exactly 18 with all named entries kept".
- **what:** Blocker found when the real 515-file corpus replaced the stale 69-file one: guardAfter measured the POST-EXPANSION scope, and unbounded Neighbourhood expansion of six named symbols tripped scope-not-narrowing at iteration 1. Fix: the narrowing guard compares the MODEL-NAMED scope (deduped NextScope, no expansion); expansion is used only for the gather and capped at MaxExpandedScope (default 18, named entries always kept). A data_request escape on the scope guard was considered and REJECTED (would render it near-inert).
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E attempt-1 blocker 2); docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** convergence guards; call-graph neighbourhood; stale-corpus masking
- **verify-later:** `pkg/diagnose/{loop,step,advance}.go` NamedScopeSize/MaxExpandedScope; `loop_scopeguard_test.go`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Deterministic scaffold / model-only-verdict split
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) "The scaffold is deterministic; the verdict is the only model-dependent part … This puts the SAFETY … in code that is verified, and isolates the part that needs a model."
- **what:** Architecture decision: loop control, guards, evidence trail, and re-scope are pure tested Go; the cite-or-abstain judgement is an interface (stub that always abstains, scripted verdicts, or the live model). The verdict runs as its OWN observable workflow step (`execute_llm_prompt`), not buried in a monolith. A model-less run can never fabricate a conclusion.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#design-and-build-choices; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** diagnose_run monolith (rejected alternative); verdict wire seam; convergence guards
- **verify-later:** `pkg/diagnose/` package purity (no DB imports); workflow verdict step config

<!-- SOURCE: U14_docs019_runbooks.md -->
### Verdict wire format seam (script IS the wire format)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) §4a "because the script format IS the wire format, every scripted scenario in §1 is a faithful dry-run of the model path"; RUNBOOK(31) §7.5 "keep the prompt's output schema and verdict_wire.go in lockstep".
- **what:** The model returns one JSON object (`outcome` ∈ CONFIRMED|REFUTED|UNVERIFIABLE, citations with `tier` ∈ static|state|runtime, revised_hypothesis, next_scope, data_requests) per PROMPT_diagnosis_verdict.md; `diagnose.ParseVerdict`/`verdict_wire.go` map it to the domain Verdict, with fail-safes: unknown outcome → UNVERIFIABLE, citation-less confirm/refute coerced to UNVERIFIABLE. Verdict scripts for testing use the identical format, so scripted runs are faithful dry-runs of the model path.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4a; docs019/RUNBOOK(31)_diagnosis_loop.md#7.5
- **relations:** cite-or-abstain loop; SeenRequests (wire sync incident); three-tier citation
- **verify-later:** `pkg/diagnose/verdict_wire.go` + `verdict_wire_test.go`; PROMPT_diagnosis_verdict.md

<!-- SOURCE: U14_docs019_runbooks.md -->
### Falsification-first evaluation gate
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6G "[x] PASSED 2026-07-01 (run 51f95cda)"; design_diagnosis_loop(7) §5 "A loop that confirms the first guess on every known bug is the failure mode, not the success — judge it on the reversals."
- **what:** The loop is not trusted on scaffold correctness; it must be run against known bugs and (a) reproduce mid-course REVERSALS (refute wrong hypotheses on evidence), (b) converge on causes the symptom could never retrieve, and (c) ABSTAIN naming the missing evidence when the bundle doesn't settle it. "Scaffold correct ≠ reasons well." The §6G pass showed UNVERIFIABLE→REFUTED→CONFIRMED over 3 iterations with cited evidence.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G; docs019/RUNBOOK_design_diagnosis_loop(7).md#5
- **relations:** gamesdesign resolveResultSpec fixture; three-tier citation; loop-worthiness test
- **verify-later:** evidence trails of runs 51f95cda, 5537ffdb, 73ed55c6 in orchestration_states

<!-- SOURCE: U14_docs019_runbooks.md -->
### gamesdesign resolveResultSpec fixture (the reference bug trajectory)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK(31) 2026-07-01 "STILL not resolveResultSpec — now for a substantive reason: reading real data, the model found a coherent cause … FORK for the user: (a) the fixture is stale … retire the 'reach resolveResultSpec' yardstick".
- **what:** The canonical eval scenario built from the real gamesdesign bug: seed "sections never reach save" → REFUTE on runtime evidence → REFUTE "token cap" → CONFIRM `resolveResultSpec` (singular output_field collapsed the page to a stub). Used as both the scripted-verdict reference and the live-eval yardstick; superseded as a yardstick once the site's current state no longer exhibited the symptom (the loop instead correctly diagnosed the missing `site_specs.cta` aspect), and the route was closed on the refute-and-confirm-a-grounded-cause bar instead.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#7.1; docs019/RUNBOOK(31)_diagnosis_loop.md#6G-passed; docs019/RUNBOOK_gamesdesign_index_rebuild.md
- **relations:** falsification eval gate; workflow result contract; B4a resultspec ceiling task
- **verify-later:** `/tmp` verdict scripts are ephemeral; groundtruth_targets.json resultspec entry

<!-- SOURCE: U14_docs019_runbooks.md -->
### Workflow-driven loop via next_step override (diagnose_route)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6C "[x] DONE"; "§6C coordinator next_step override CONFIRMED (coordinator.go:1093 getNextStepFromResult)"; §6E "[x] DONE 2026-06-29 (5× loop-back, CONFIRMED)".
- **what:** The loop is workflow-driven, not action-internal: `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict (execute_llm_prompt) → route (diagnose_route) → [loop back | emit] → complete`. `diagnose_route` runs the engine's Advance (guards + call-graph re-scope) once per iteration and overrides `next_step` in its result (the conditional_route pattern); it sets no output_field so its results are read as `route.*`. The workflow lives in agent_definitions `default_config` (not the legacy orchestration_workflow columns).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** diagnose_run (abandoned alternative); state threading; coordinator getNextStepFromResult
- **verify-later:** `platform/orchestration/actions/diagnose_route_action.go`; `coordinator.go` getNextStepFromResult; diagnose-agent default_config

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose_run internal-iteration monolith
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** RUNBOOK(5) §6E "In this design there is NO workflow loop-back: the iteration lives inside the diagnose_run action (the engine Run())"; RUNBOOK(31) §6C "The BUILT design is the workflow-driven loop, NOT a diagnose_run action — there is no diagnose_run"; design_diagnosis_loop(7) "(The earlier diagnose_run monolith was removed.)"
- **what:** The earlier design where a single `diagnose_run` action executed the whole capped loop internally (orchestration shows one `run_loop` step; iteration visible only in logs/trail). Dropped in favour of the workflow-driven loop so each iteration's verdict and routing are separately observable orchestration steps. The seeded diagnose-agent briefly referenced the nonexistent action — the workflow-fix migration removed it. Family-delta: present in RUNBOOK(2)–(7), gone by RUNBOOK(8).
- **sources:** docs019/RUNBOOK(5).md#6E; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** workflow-driven loop (replacement); deterministic scaffold split
- **verify-later:** absence of diagnose_run in registry.go; diagnose-agent workflow JSON

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnostician draft and the seed→fix migration path
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK(31) §6C "Do NOT seed a new one (the diagnostician draft is superseded)"; "Do NOT apply the older NNN_move_diagnose_workflow_to_default_config.sql (bannered superseded)"; RUNBOOK(2) §E was "apply the seed migration (NNN_seed_diagnose_agents.sql)".
- **what:** The lineage of getting the diagnose pair into agent_definitions: an early `diagnostician` single-agent draft, then a seed-agents migration (RUNBOOK(2) era), superseded by fixing the ALREADY-seeded diagnose-agent/diagnose-orchestrator pair in place (workflow rewritten to diagnose_route shape in default_config, orchestrator workflow separately restored after the move migration nulled it). Every agent_definitions-touching migration snapshots the row first (`snapshot_agent`).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK(2).md#E
- **relations:** standing evidence rules (snapshot_agent); workflow-driven loop
- **verify-later:** migrations NNN_fix_diagnose_agent_workflow.sql, NNN_restore_diagnose_orchestrator_workflow.sql; agent_definitions snapshots

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose-orchestrator spawn-wrapper pattern
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6F "Target the ORCHESTRATOR; diagnose-agent is the worker it spawns … keeping the loop off shared pods"; run evidence throughout §6E–§6G.
- **what:** The diagnosis entry point is a thin orchestrator (spawn_diagnoser → call_diagnoser → complete) that spawns a dedicated diagnose-agent pod and forwards its result, keeping heavy in-chassis loop work off the shared chassis pods. The same pattern was replicated for indexing (index-orchestrator, §7B.1) when in-place `orchestrate` proved token-less on shared pods.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1
- **relations:** repo-cloning token gate; generic orchestrate envelope; code-indexer
- **verify-later:** diagnose-orchestrator/index-orchestrator agent_definitions; spawn_actions.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### data_requests channel — model-authored read-only SQL
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) update 2026-07-01 (run 51f95cda) "the model's data_requests RAN (verdict_wire.go confirmed live)"; §6C "The data_requests channel — now wired (was dormant from a wiring gap, not by design)."
- **what:** The verdict may emit `data_requests` (single read-only SELECTs with `sql`/`why`); `diagnose_route` reads them from the verdict wire, keeps only read-only ones, forwards to `route.data_requests`; `diagnose_load_runtime` executes each on loop-back in a READ ONLY transaction with SET LOCAL statement_timeout and appends rows to runtime_evidence. Code re-scope and data re-gather are deliberately separate channels. This is the "DB-following" arm of evidence-following.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6B; docs019/RUNBOOK(31)_diagnosis_loop.md#6C (data_requests wiring)
- **relations:** three-guard model; EXPLAIN size guard; SeenRequests; live schema section
- **verify-later:** `diagnose_load_runtime_action.go` runDataRequests; `diagnose_route_action.go` readOnlyDataRequestsFromWire

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three-guard read-only SQL enforcement model
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "SELECT-only is enforced at THREE layers (confirmed in the code)"; design_diagnosis_loop(7) §4d "CONFIRMED on this cluster (2026-06-17): pool_mode = transaction … a live BEGIN READ ONLY; DELETE … WHERE false probe refused the write".
- **what:** Defence in depth for model SQL: Guard 1 = the verdict prompt constrains to a single read-only SELECT; Guard 2 = `IsReadOnlySQL` lint applied twice (route boundary and pre-execution); Guard 3 = the actual guarantee, a `BeginTx(ReadOnly:true)` transaction (+ statement_timeout) that rejects any write including data-modifying CTEs. The `WHERE false` DELETE probe is the standard non-destructive verification. Guards 1–2 are hygiene, never the safety boundary.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#4; docs019/RUNBOOK(31)_diagnosis_loop.md#6B; docs019/RUNBOOK_design_diagnosis_loop(7).md#4d
- **relations:** sqlguard stripQuoted; diagnose_ro role; data_requests channel
- **verify-later:** `pkg/diagnose/sqlguard.go`; BeginTx call in diagnose_load_runtime

<!-- SOURCE: U14_docs019_runbooks.md -->
### sqlguard stripQuoted — lint false-positive on quoted literals
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK(31) (run 5537ffdb, 2026-07-01) "the page slug 'tool-drop-rate-simulator' contains 'drop' … FIXED (sqlguard.go stripQuoted blanks literal/identifier contents before the scan; regression test added)"; §6G banner "REMAINING … (a) DEPLOY the lint fix — latent" (2026-07-02).
- **what:** Keystone bug: the read-only lint scanned raw SQL, so a keyword substring inside a string literal (slug containing "drop") caused legitimate reads to be silently dropped — neutralising both the schema-section content read and the progress rule. Fix blanks literal/identifier contents before keyword scanning. Written + tested; the runbooks record deployment as still pending at the family's last update.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G-passed; docs019/RUNBOOK(31)_diagnosis_loop.md#update-5537ffdb
- **relations:** three-guard model; data_requests channel
- **verify-later:** `pkg/diagnose/sqlguard.go` stripQuoted + test; whether the deployed image carries it

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnose_ro role and pooler-aware read-only enforcement
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK(31) checklist "[x] §6B diagnose_ro role migration written … [ ] role migration applied"; RUNBOOK(31) "data_requests run via db.BeginTx(ReadOnly) on params.DB (clients_user), NOT a restricted role".
- **what:** A GRANT-only SELECT role (`diagnose_ro`) for the harness path, where `psql -c` statement stacking makes a transaction wrapper unsafe. Key doctrine: under pgbouncer enforce read-only by GRANT, never by `SET default_transaction_read_only` (session settings leak across pooled backends); transaction pooling makes BeginTx(ReadOnly) safe; statement_timeout goes in the DSN options. The chassis path deliberately runs as clients_user under the read-only transaction instead, so content tables stay SELECTable without grants.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4d; docs019/RUNBOOK(31)_diagnosis_loop.md#6B
- **relations:** three-guard model; dbcontext harness
- **verify-later:** NNN_create_diagnose_ro_role.sql applied?; pgbouncer-config pool_mode

<!-- SOURCE: U14_docs019_runbooks.md -->
### EXPLAIN pre-flight size guard on data requests
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "EXPLAIN size-guard added … runDataRequests now runs EXPLAIN (FORMAT JSON) inside the read-only tx BEFORE each query"; 51f95cda validation "the EXPLAIN guard (didn't block site-scoped queries)".
- **what:** Before executing each model query, the action plans it (EXPLAIN FORMAT JSON, no execution) and skips with feedback if estimated rows exceed budget (explain_max_rows 50000; cost cap opt-in); output rows are capped (row_cap 200) and cells truncated rune-safe (cell_chars 600); statement_timeout remains the execution backstop. A skip is feedback the model narrows from — a new narrower request counts as progress.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel; SeenRequests; responses-are-summaries doctrine
- **verify-later:** runDataRequests EXPLAIN branch in diagnose_load_runtime_action.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### Live schema section in the bundle (gatherSchema)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "#2 IMPLEMENTED … the bundle now carries a ## Schema (live tables) section"; 51f95cda "using REAL table/column names — the schema section paying off, no more page_sections guessing".
- **what:** `diagnose_load_runtime` gains one read-only information_schema.columns query, DENYLIST-driven (%backup%/%bak%/%archive%/%supersede%, deliberately not %snapshot% since site_snapshots is live) plus a broad relevance include (site%/page%/content%/flow%) unless `schema_full=true`; rendered into the bundle via Go-defaulted config so no migration was needed. Stops the model guessing table names (it had invented `page_sections`).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel; denylist-over-allowlist style (shared with index-hygiene excludes)
- **verify-later:** gatherSchema in diagnose_load_runtime_action.go; runtime.schema render path

<!-- SOURCE: U14_docs019_runbooks.md -->
### Loop state threading and the re-seed self-check
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "§6E.1 DONE — state threading verified … trail_len == iteration (3) and stopped on a guard"; the self-check "aborts loudly with stopped_by: state-threading-error rather than silently re-seeding into a runaway".
- **what:** Loop state (iteration, trail, seen_citations, hyp_history) threads across iterations via `state_field = route.diagnose_state`; a mis-pointed state_field silently disarmed the cap/trail/guards (each iteration re-seeded fresh). Fix = migration + a code self-check: if diagnose_route is about to seed but route.diagnose_state already exists, abort loudly — a regression tripwire for the exact bug class.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position; docs019/RUNBOOK(31)_diagnosis_loop.md#6E
- **relations:** convergence guards; loop_scope_field lesson (same dotted-path config family)
- **verify-later:** NNN_fix_diagnose_route_state_threading.sql; self-check branch in diagnose_route_action.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### loop_scope_field / EncodeScope shape-mismatch lesson
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6C "CONFIRMED ISSUE + FIX … EncodeScope is json.Marshal of the Scope struct … keys are the Go field names … ExtractStringListHelper … coerces that OBJECT to empty — so on every loop-back the scope … NEVER advanced"; "loop_scope_field migration confirmed live (Run 2 error read route.scope.Symbols)".
- **what:** A silent contract mismatch between an action's encoded output (untagged Go struct → `{"Symbols":[...]}`) and a downstream dotted-path reader expecting a plain list: first-pass worked, every re-scope was inert — invisible to engine tests because it lived in workflow config. Fix was config-only: point `loop_scope_field` at `route.scope.Symbols`. Emblematic of the dotted-lookup config contract class (also: analysis_field, result_from, repo_field).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C
- **relations:** state threading; repo-label asymmetry; workflow result contract
- **verify-later:** NNN_fix_assemble_bundle_loop_scope_field.sql; ExtractNestedField 3-level traversal

<!-- SOURCE: U14_docs019_runbooks.md -->
### code_symbols index + code-indexer agent
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D "[x] index populated (436 rows)"; code_retrieval_route(21) §7C "4,155 symbols; 499 distinct paths … min=max=commit 36710be; prune cleared all 436 old rows"; measured "≈ 5 symbols/sec through the single ollama-adapter".
- **what:** The retrieval corpus: `code_symbols` (repo, path, symbol, kind, signature, doc, content, embedding, commit_sha) written solely by the `code-indexer` agent (request_repo_analysis → await analyser → index_code_symbols; later analyse_repo_local in-process) and read by `lookup_code_symbols` (vector + trigram). UPSERT-safe via uq_code_symbols_identity; prune removes rows whose commit_sha differs from the new index commit; embedded text is name + signature + first doc line + path. Triggered via index-orchestrator (spawning wrapper so the pod holds the read token).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_code_retrieval_route(21).md#7A–7C; docs019/RUNBOOK_thin_slice(27).md#in-cluster-path
- **relations:** analyser adapter; analyse_repo_local; repo-label convention; evidence-fed resolver
- **verify-later:** code_symbols table + constraints; code-indexer/index-orchestrator agent rows; index_code_symbols action

<!-- SOURCE: U14_docs019_runbooks.md -->
### Analyser adapter — repo analysis as a Kafka service
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D "Adapter availability — RESOLVED 2026-06-24 (secret name/key mismatch) … the adapter came up 1/1 Running"; thin_slice(27) in-cluster deployment section (kustomize dry-build, topic CRD, health checks).
- **what:** A deployed adapter (`internal/adapters/analyser`, topic system.adapter.analyser.requests) that clones a GitHub repo (read-only PAT) and returns the analysis Output over Kafka. Deployment lessons captured: inject the single needed secret via secretKeyRef (never envFrom, which exposes every platform secret); topic auto-create is off so the KafkaTopic CRD must exist; topic-addressed adapters legitimately show target_agent_type='unknown' in awaited_requests; idle consumer-poll timeouts log at ERROR cosmetically. Its per-iteration use by the loop was later removed (analyse_repo_local), but indexing and other consumers remain.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_thin_slice(27).md#in-cluster-path
- **relations:** analyse_repo_local (supersedes the loop's cross-pod call); code-indexer; repo-cloning token gate
- **verify-later:** deployments/kustomize/services/analyser-adapter; personae-platform-secrets GITHUB_READ_TOKEN wiring

<!-- SOURCE: U14_docs019_runbooks.md -->
### Repo-label composition convention (owner/repo) and the lookup asymmetry bug
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Label convention — DECIDED 2026-06-11: code_symbols.repo is the owner/repo form … COMPOSED by index_code_symbols"; RUNBOOK(31) "ROOT CAUSE CONFIRMED (2026-06-26) — repo-label asymmetry … the lookup queried WHERE repo='agentchassis' against rows under 'gqls/agentchassis' → 0 hits"; "Structural patch APPLIED".
- **what:** `code_symbols.repo` is always the composed `owner/repo` label. The index composed it but the lookup didn't → iteration-1 seeding returned nothing ("no scope"). Fixed twice: a config-only workaround (literal repo on the lookup step) then the structural `resolveCodeRepoLabel` shared by index AND lookup so they cannot diverge. Also the standing diagnostic rule it produced: confirm by correlation_id, never by `LIMIT 1` (a COMPLETED LIMIT-1 row was a red herring twice).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F-run1; docs019/RUNBOOK_thin_slice(27).md#label-convention
- **relations:** loop_scope_field lesson (same config-contract class); standing evidence rules
- **verify-later:** resolveCodeRepoLabel in code_symbols_actions.go; lookup step config (no repo_field literal)

<!-- SOURCE: U14_docs019_runbooks.md -->
### analyse_repo_local — in-process tarball fetch + analysis
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6F "DECIDED for option 3, REFINED turn 27 … BUILT (turn 27, gofmt-clean)"; checklist "§6C … image also carries analyse_repo_local + lifted internal/reposource — DONE 2026-06-29"; §7B swap "migration applied; snapshot 971da9c9".
- **what:** Resolution of the "no repo checkout on the diagnose pod" blocker: the agent fetches the repo itself via the analyser's tarball fetcher (`GET /repos/{o}/{r}/tarball/{ref}`, no git in the chassis) lifted into a neutral `internal/reposource` package, runs `analysis.Analyse(dir)` in-process for spans + call graph, and reads bodies from that checkout. `pin_to_index_commit` pins the fetch to the dominant code_symbols commit so seeded path:Symbol entries resolve (the indexer sets it false — it DEFINES the commit). Options weighed and rejected: bodies-in-DB (whole-repo Kafka payloads) and a stateful analyser serving slices (per-iteration coupling).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F (Run 3 + deploy sequence); docs019/RUNBOOK_code_retrieval_route(21).md#7B
- **relations:** analyser adapter; code_symbols; index hygiene excludes; repo-cloning token gate
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; NNN_swap_analyse_repo_to_local.sql

<!-- SOURCE: U14_docs019_runbooks.md -->
### Repo-cloning token gate (isRepoCloningAgent)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "spawn_actions.go injects GITHUB_READ_TOKEN via secretKeyRef … gated by isRepoCloningAgent -> ONLY diagnose-agent pods get the token; the spawner never holds it"; §7B.1 "isRepoCloningAgent gained 'code-indexer' … Verified end to end by run 93ba14e6".
- **what:** Least-privilege credential injection at spawn time: only agent types allowlisted as repo-cloning receive the read-only GitHub token env (secretKeyRef into the spawned pod), and the shared chassis pods never hold it — which is why indexing/diagnosis run through spawning orchestrators rather than in-place.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1
- **relations:** diagnose-orchestrator wrapper; analyser adapter secret lesson
- **verify-later:** spawn_actions.go isRepoCloningAgent list

<!-- SOURCE: U14_docs019_runbooks.md -->
### Stale-corpus class: HEAD pinning, explicit refs, CI-triggered indexing
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** code_retrieval_route(21) §7A "it was a faithful index of a YEAR-OLD tree … Decision: … envelopes ALWAYS carry an explicit branch/sha"; queue item 4 "CI-triggered indexing: GitHub Actions step firing the index-orchestrator envelope with ${GITHUB_SHA} on push … [queued]".
- **what:** A recurring staleness class: consumers pinned to `HEAD`/`latest` silently track an ancient artefact (remote HEAD = unmoved main from 2025; agent image_tag 'latest' = pre-architecture build). Adopted: explicit refs in every envelope, derive REF from the working checkout. Designed (aspirational): Structural A — a post-deploy CI step indexes at ${GITHUB_SHA} so index commit == deployed commit by construction; Structural B — fast-forward main to the deployed sha. Rejected: resolving "most recently pushed branch" via API (latest-pushed ≠ deployed).
- **sources:** docs019/RUNBOOK_code_retrieval_route(9).md#ref-strategy; docs019/RUNBOOK_code_retrieval_route(21).md#7A; docs019/RUNBOOK_builder_route(21).md#queue (item 4)
- **relations:** image_tag 'latest' trap (same class); code_symbols prune semantics
- **verify-later:** GitHub Actions workflow for post-deploy indexing (absent?); git ls-remote origin HEAD

<!-- SOURCE: U14_docs019_runbooks.md -->
### Index hygiene — exclude archived code copies, prune by commit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) §7C.1 "[x] census: docs-archived 430 symbols / 50 files … interim DELETE run"; "[x] reindex f284b749 VERIFIED (2026-07-03): commit e3176f8, 3,723 symbols, docs_rows=0".
- **what:** The repo stores archived copies of its own code under docs/ (and download-suffixed `name(N).go` files); indexing them pollutes retrieval with dead duplicates (observed: nine duplicate assembler copies as ranks 1–9). Fixes: the analyser skips `*(N).go` unconditionally; `analyse_repo_local` gained `exclude_patterns` (Go default ["docs/"]) calling AnalyseWithExclude; prune semantics (`commit_sha IS DISTINCT FROM $new`) clear old-commit rows on the next reindex. Same trap documented CLI-side ("analyse the RIGHT ROOT", relative -exclude substrings).
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7C.1; docs019/RUNBOOK_thin_slice(27).md#known-limits; docs019/RUNBOOK_thin_slice(27).md#B4a (build the index over REAL source)
- **relations:** analyse_repo_local; stale-corpus class; B4a eval discipline
- **verify-later:** exclude_patterns config on analyse_repo_local; code_symbols docs/% row count

<!-- SOURCE: U14_docs019_runbooks.md -->
### Evidence-fed fuzzy-scope resolver (§7D)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) route-close run 73ed55c6: "resolver canonicalisation (model's basenames → full paths @0.81–0.87) AND descriptive resolution … both load-bearing in the confirming scope"; "[x] code WRITTEN 2026-07-02 … resolver image is LIVE".
- **what:** Many verdict `next_scope` entries are English descriptions, not path:Symbol handles — previously inert (no call-graph match, no body sliced). The resolver, inside diagnose_route after verdict-parse and before Advance, embeds each non-exact entry (same nomic client/prefixes) and vector-searches code_symbols, replacing it with the top hits (resolver_top_k default 2 — tuned so substitution stays inside the narrowing guard's +2 allowance; min similarity 0.55; unresolvable entries survive as labels, "no worse"). Flagged deliberate change: the trail records the RESOLVED scope, the more auditable record. Reuses the seed lookup's retrieval machinery wholesale.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D; docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** code_symbols; named-scope guard (the +2 interplay); §7F seed reorder (retired by this)
- **verify-later:** diagnose_route_action.go resolver step 3.5; diagnose_route_resolver_test.go

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three-tier citation standard (static / data / runtime)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "CONFIRMED on citations spanning ALL THREE TIERS: Tier 2 work item, Tier 1 site_specs cta → (0 rows) (query+result cited together), Tier 0 plan_sections_action.go:planSection quoting 'case \"skip_field\"'" (run 73ed55c6, 2026-07-03).
- **what:** Verdict citations carry a tier (static code / live data reads / runtime records); the route's success bar — and the strongest diagnosis shape — is a CONFIRMED grounded across all three tiers, with query+result cited together for data reads and a quoted code branch for the mechanism. Distinguishes "confirmed by inference at the data layer" from "code-level mechanism named".
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#route-closed; docs019/RUNBOOK_design_diagnosis_loop(7).md#1 (tier vocabulary)
- **relations:** verdict wire seam; falsification eval gate; fix-loop council opinions (verdict-wire-style contract)
- **verify-later:** citation tier handling in verdict_wire.go; evidence_trail of run 73ed55c6

<!-- SOURCE: U14_docs019_runbooks.md -->
### §7F seed-query reorder (lookup after load_runtime)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** code_retrieval_route(21) "§7F RETIRED" (banner) after "SEED RELEVANCE MET — all twelve seed symbols build-domain … first time ever; §7F (seed reorder) substantially retired".
- **what:** A deferred design to reorder lookup_symbols after load_runtime so the seed query could be built from the symptom PLUS salient error-log lines. Made unnecessary once the corpus was current and the resolver landed — seed relevance was proven by content twice. Family-delta: the idea persists as a section in every version but flips from DEFERRED to RETIRED at (18)+.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7F; docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E scoring)
- **relations:** evidence-fed resolver (the retiring cause); code_symbols currency
- **verify-later:** n/a (not built)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Corpus enrichment policy — measure first, mechanical before authored
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** code_retrieval_route(21) "Should every function carry a human description for embedding-match? NO … Order of investment, gated on the §7E measurement" (question raised 2026-07-02).
- **what:** Position on enriching the retrieval corpus: (1) mechanical, rot-free first — extend composeSymbolContent with a function's string literals (diagnosis queries quote log lines and the literals ARE the log lines); (2) Go-convention one-sentence docs only on the exported surface + action entrypoints; (3) explicitly NO separate tag system — the doc first line is the tag surface. Rationale: stale docs make retrieval confidently wrong, the worst failure mode for a cite-or-abstain loop.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#corpus-enrichment
- **relations:** code_symbols; F3 learning layer (doc enrichment feed-back)
- **verify-later:** composeSymbolContent; exported_no_doc census query

<!-- SOURCE: U14_docs019_runbooks.md -->
### Reasoning-state as a first-class handoff artefact
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** thin_slice(27) improvement 5 "A bundle carries CODE + SCHEMA + RUNTIME EVIDENCE, but NOT reasoning state … The stopgap is a hand-written 'diagnosis so far' preamble (PREAMBLE_gamesdesign_diagnosis_handoff.md)"; the loop's evidence_trail later persists per-iteration hypothesis/scope/verdict.
- **what:** The insight that a context bundle without the evidence trail forces a fresh reader to re-derive falsified hypotheses; the design goal is a structured reasoning-state section accumulating across iterations (hypotheses tried, verdict + citation each, open discriminator). Partially realised by the loop's evidence trail in collected_data; the bundle-intrinsic version and per-iteration notes (F0.3 via doc_notes) remain in flight.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#next-improvements (item 5); docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F0.3)
- **relations:** per-task running notes; diagnosis_artifacts egress; falsification eval gate
- **verify-later:** evidence_trail shape in collected_data; doc_notes diagnosis category rows

<!-- SOURCE: U14_docs019_runbooks.md -->
### Instrument-skepticism doctrine
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) "almost every wrong measurement came from the test instrument, not the system under test — a wrong task string, a contaminated index, a stale shell variable, a task description that leaked the answer's vocabulary".
- **what:** Standing caution carried into the loop's design: apply cite-or-abstain suspicion to one's OWN inputs (the bundle, the query, the ground truth) before suspecting the target system. Surfaced repeatedly in B4a and encoded in the eval harness guards; named as the thing to watch when evaluating the model verdict.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#a-standing-caution; docs019/RUNBOOK_thin_slice(27).md#B4a-task-1
- **relations:** ground-truth eval harness; standing evidence rules (0-rows not decisive)
- **verify-later:** n/a (doctrine)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Base-runbook gated-items framing (PLAN.md linkage)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK.md §6 "Gated items (carried — see PLAN.md) … None are unblocked by this thread's work alone" — replaced from RUNBOOK(1) onward by the inlined "§6 Completing the whole task (what remains)" with per-step DoD.
- **what:** The earliest form of the diagnosis runbook kept only in-flight build steps and deferred the roadmap to a separate PLAN.md; within one version the roadmap was inlined as §6 with per-step definitions of done and live status, and the runbook became the single self-contained thread state (later §7 split out to its own file when §6 closed). Family-delta record of the project's documentation style converging on self-contained travelling runbooks.
- **sources:** docs019/RUNBOOK.md#6; docs019/RUNBOOK(1).md#6; docs019/RUNBOOK(31)_diagnosis_loop.md#6 (ACTIVE ROUTE MOVED banner)
- **relations:** parallel-thread convention; documentation-system
- **verify-later:** PLAN.md in docs019 (sibling file, other unit)

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnosis loop (contextkit)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v4 STATE OF THE WORLD (2026-07-02): "Diagnosis loop (§6): DONE. §6A–§6G all passed; §6G accepted on run `51f95cda`... Engine (pkg/diagnose) + diagnose-agent workflow live."
- **what:** A read-only, human-gated agent that diagnoses chassis bugs by iterating hypothesise → gather scoped evidence → cite-or-abstain verdict → re-scope by following the evidence (call graph for code, vetted/model-written queries for data) — never re-searching the symptom, never fixing, never triggering a run. Built first as a standalone Go engine (`contextkit/internal/diagnose/`) with a tested scaffold (guards, trail, verdict parsing), then ported into the chassis as a workflow-driven agent (`diagnose-agent`/`diagnose-orchestrator`) where each iteration is an observable sequence of steps (gather → verdict via `execute_llm_prompt` → `diagnose_route` → loop-back or `diagnose_emit`).
- **sources:** NOTES_running_synthesis_v2(36).md §STATE DIGEST 2026-06-17; NOTES_running_synthesis_v3(32).md §STATE DIGEST; NOTES_running_synthesis_v4(39).md §STATE OF THE WORLD; NOTES_running_synthesis_principles(59) "diagnosis-loop design updated" entries.
- **relations:** contextkit CLI toolchain; convergence guards; verdict cite-or-abstain contract; call-graph re-scope mechanism; B4a embedding-quality finding; diagnosis→fix loop workstream (successor pivot).
- **verify-later:** `pkg/diagnose/` in chassis repo; `platform/orchestration/actions/diagnose_*.go`; `agent_definitions` rows for `diagnose-agent`/`diagnose-orchestrator`; whether the "eval gate" (reproduce the gamesdesign reversals on a live model) was ever actually run.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnosis-loop chassis integration architecture
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v3 STATE DIGEST: "the agent rewrite landed: `diagnose-agent` now has the `diagnose_route` workflow in `default_config`"; v4: "§7 ROUTE CLOSED: run 73ed55c6 full trail read; §7E green".
- **what:** The loop is realised as a chassis AGENT (workflow of steps), not a new CLI or long-running service, following "every agent is an orchestrator": a thin `diagnose-orchestrator` spawns a `diagnose-agent` worker whose workflow (in `default_config`, not the three NULL `*_workflow` columns) is `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict(execute_llm_prompt) → route(diagnose_route) → [loop to assemble_bundle | emit] → complete`. The verdict step reuses the existing `execute_llm_prompt` action rather than a new action; `diagnose_route` is a router action that sets no `output_field` (its result lands under step-name `route`) and returns `next_step` per the coordinator's `getNextStepFromResult` mechanism.
- **sources:** NOTES_running_synthesis_v2(36).md (chassis integration entries, 2026-06-17); NOTES_running_synthesis_v3(32).md DECISIONS (diagnose_route seeding/state-threading fixes).
- **relations:** Diagnosis loop; Workflow default_config location convention; SagaCoordinator output_field contract.
- **verify-later:** `agent_definitions` rows for diagnose-agent/orchestrator; `coordinator.go` `getNextStepFromResult`/`ProcessResponse`.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Verdict cite-or-abstain contract
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the diagnosis loop is now COMPLETE as far as can be built off-chassis — design + tested scaffold + tested adapters + runnable entrypoint + the model prompt + the tested prompt↔scaffold seam" (principles(59), 2026-06-17 entry, mirrored in v2(36)).
- **what:** The model-facing prompt contract (`PROMPT_diagnosis_verdict.md`) requires every verdict to CONFIRM or REFUTE only with a verbatim-quoted citation from the bundle, else the outcome is coerced to UNVERIFIABLE; abstention is asymmetric (runtime evidence readily refutes, but confirms only on direct mechanism, never "consistent with"); the re-scope must follow what the evidence names, not re-search the symptom; and the model is told to apply the same suspicion to its own reading of the bundle that the loop applies to hypotheses. A parallel wire format (`verdict_wire.go`) parses model output as human-legible strings (CONFIRMED/REFUTED/UNVERIFIABLE) with fail-safe unknown→UNVERIFIABLE coercion.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "VERDICT PROMPT drafted"; NOTES_running_synthesis_principles(59) DB discipline / diagnosis-loop design entries.
- **relations:** Diagnosis loop; convergence guards; data-request channel (model-written SQL).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Convergence guards for the diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "The four convergence guards (iteration-cap, scope-not-narrowing, evidence-not-growing, hypothesis-thrash) + the no-citation→UNVERIFIABLE coercion are all behaviour-tested" (v2(36) STATE DIGEST); v3(32) DECISIONS: "A new data_request counts as forward progress in the spin guards (turn 31)".
- **what:** A set of anti-spin safety mechanisms bounding the loop: an iteration cap, a rule that re-scope can't balloon past prior scope + 2, a rule that a verdict adding no new citation halts the loop, and thrash detection for hypothesis oscillation without new discriminating evidence. Later hardened so an issued (not yet cited) read-only data request also counts as forward progress, preventing the loop from stopping one iteration before a fixed query's result would have arrived.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 scaffold entries; NOTES_running_synthesis_v3(32).md DECISIONS (turns 30-31).
- **relations:** Diagnosis loop; verdict cite-or-abstain contract.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Call-graph re-scope mechanism
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "CallGraph adapter... reads analyser `calls`, resolves callee NAMES back to defining symbols for re-scope" (v2(36) 2026-06-17).
- **what:** Re-scoping in the diagnosis loop follows the analyser's recorded (name-based, not type-resolved) call graph outward from an evidence-named site, deliberately dropping ubiquitous names (Run/String/Error/New/... plus any name resolving to more than 8 definitions) so following doesn't explode into noise — described as "the symptom-vocabulary trap in call-graph form."
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "diagnosis-loop adapters... BUILT & tested".
- **relations:** Diagnosis loop; B4a embedding-quality finding.

<!-- SOURCE: U15_docs019_running_notes.md -->
### B4a embedding-quality evaluation & symptom-vs-mechanism retrieval ceiling
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "embeddings do NOT earn a code-path place; the lever is the diagnosis loop... Retrieval is necessary-not-sufficient"; v4(39) STATE OF THE WORLD: "The code-retrieval channel contributes nothing (measured: flat similarity band 0.547–0.574 across all 12 seed hits; zero code citations in four full runs)."
- **what:** A rigorous two-task (later extended) evaluation of lexical vs. semantic (nomic/Ollama) vs. fused (RRF) code-symbol retrieval against real bugs, run through five corrected measurement setups (wrong task string, contaminated index, duplicate-symbol pollution, stale shell var, task-string vocabulary leakage — "the instrument, not the system, was the fault" every time). Conclusion: when a bug's cause lives in shared infrastructure named for its function, not its failure mode, symptom-based retrieval — lexical AND semantic alike — has a category-level ceiling (zero vocabulary overlap, not a ranking problem); naive RRF fusion can make results WORSE than lexical alone by demoting a lone correct hit. Later (v4) measured that the code-retrieval channel contributed essentially nothing across real runs, while runtime/DB evidence carried every successful diagnosis.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-14/06-17 B4a entries; NOTES_running_synthesis_v4(39).md STATE OF THE WORLD.
- **relations:** Diagnosis loop; call-graph re-scope mechanism; code-context retrieval infrastructure; reuse-checking retrieval architecture.
- **verify-later:** `groundtruth_targets.json`, `eval_targets` results in the repo, whether ground truth was ever widened beyond ~2 tasks.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnose-agent self-contained repo fetch
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "Symbol bodies come from a git checkout the diagnose-agent makes itself (Option 3, turns 25-26)."
- **what:** Rather than adding a body column to `code_symbols` or coupling every diagnosis iteration to a live analyser holding a checkout, the diagnose-agent fetches its own tarball (reusing the analyser's `FetchToDir`, lifted into a neutral `internal/reposource` package so both the analyser adapter and the diagnose action share one fetcher) and runs `internal/analysis` in-process for both the call graph and symbol-body slicing — one fetch, no cross-pod coupling, git stays the only source of truth for code. Fetches are pinned to the same commit the `code_symbols` index was built on (best-effort, falls back to `ref`/HEAD) so lookup-seeded symbols resolve in the fetched tree.
- **sources:** NOTES_running_synthesis_v3(32).md turns 25-27 (DECISIONS).
- **relations:** Analyser adapter build; code-context retrieval infrastructure; symbol-body slicer (ReadSymbolBody).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Data-request channel (adaptive DB-evidence gather)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** v4(39) STATE OF THE WORLD: "Correction (the data_requests channel is real and now wired)... it was dormant from a 3-part wiring gap, now fixed."
- **what:** The mechanism by which a diagnosis-loop verdict can name its own read-only SQL query as a `data_request`, which the loop lints (Guard 2), executes read-only (Guard 3), and folds into the next iteration's bundle — replacing an earlier, more limited "vetted query catalogue only" design once the read-only transaction guard was proven sufficient as the real safety boundary. The catalogue survives as a fast-path/few-shot-examples layer, not the only path. Was found dormant (misdiagnosed twice — first as "a gap to wire", then over-corrected to "dormant by design") due to a three-part wiring gap between `diagnose_route`, `diagnose_load_runtime`, and the migration's `gather_step`.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 DB-evidence design entries; NOTES_running_synthesis_v3(32).md STATE DIGEST "Correction" paragraph.
- **relations:** Model-written SQL guard model; diagnosis loop; doc/query catalogue relevance selection.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Doc/query catalogue relevance-keyed selection
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "NEW internal/diagnose/docselect.go: DocRule{Doc, Keywords, PathGlobs, Always} + SelectDocs(hypothesis, scope, rules)... TESTS: docselect_test.go" (v2(36), 2026-06-17).
- **what:** A pure, tested, per-iteration selector (`SelectDocs`/`SelectQueries`, sharing helpers) that pulls task-specific reference documents or SQL query templates into a diagnosis bundle only when their keywords/path-globs match the current hypothesis/scope, keeping the always-on constitution small while still surfacing the relevant 003-style contract or domain query "by relevance" rather than dumping every doc into every bundle (a deliberate anti-bloat decision citing the B4a context-rot lesson).
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "per-hypothesis -doc selection wired into the loop" and "adaptive DB-evidence gather" entries.
- **relations:** Data-request channel; context substrate principles (context rot avoidance).

<!-- SOURCE: U16_docs019_design_plans.md -->
### Iterative-bundle diagnosis loop (the automated debugging motion)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN_diagnosis_loop(3) "Status: … the engine is now BUILT. As of 2026-06-24, `pkg/diagnose` … exists and is tested"; PLAN.md DONE list; README_flows describes live runs.
- **what:** Automates the five-move debugging loop performed by hand in the gamesdesign session: hypothesise from a symptom, gather read-only evidence (bundle), test the hypothesis against the evidence (verdict), re-scope from what the evidence revealed, iterate until pinned or capped. Output is always a diagnosis plus full evidence trail, never a fix. Moves 1/2/4/5 are mechanical; move 3 (falsification) is the crux.
- **sources:** DESIGN_diagnosis_loop(3).md#0-1; README_iterate_until_bugfix_notes.md; README_overview.md; PLAN.md
- **relations:** cite-or-abstain verdict contract; convergence guards; chassis diagnose_route realisation; diagnosis→fix loop
- **verify-later:** pkg/diagnose/{loop,step,advance,callgraph,verdict_wire}.go; contextkit/internal/diagnose; agent_definitions rows diagnose-agent/diagnose-orchestrator

<!-- SOURCE: U16_docs019_design_plans.md -->
### Cite-or-abstain verdict contract + diagnosis verdict prompt
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Prompt is inline (JSON-escaped) in the applied NNN_fix_diagnose_agent_workflow(2).sql verdict step; verdict_wire.go parses its output (tested).
- **what:** Per iteration the model must return exactly one of CONFIRMED / REFUTED / UNVERIFIABLE with verbatim citations; a citation-less confirm/refute is coerced to UNVERIFIABLE; CONFIRMED only on direct evidence ("consistent with" = UNVERIFIABLE — the abstention asymmetry); no fix may be proposed; each citation tier-tagged with freshness. The prompt carries a worked REFUTED example (the gamesdesign reversal) and a self-suspicion caution. Schema must stay in lockstep with verdict_wire.go.
- **sources:** PROMPT_diagnosis_verdict(1).md; DESIGN_diagnosis_loop(3).md#2; NNN_fix_diagnose_agent_workflow(2).sql
- **relations:** doc-drift classifier evidence-or-abstain (its origin); falsification-first principle
- **verify-later:** pkg/diagnose/verdict_wire.go tests; live diagnose-agent default_config verdict step

<!-- SOURCE: U16_docs019_design_plans.md -->
### Falsification-first / confident wrongness as the single enemy
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** README_02 "The single enemy is confident wrongness. Runs 1–2 of the benchmark produced CONFIRMED verdicts that were wrong… Everything since … is aimed at that one failure mode" (2026-07-09 context).
- **what:** The design premise of the whole project: LLMs rationalise their first hypothesis, so every mechanism (citation mandate, REFUTED-is-correct framing, guards, council, closure gate) exists to force explicit falsification and make abandoning a wrong hypothesis cheap. The most valuable move in the founding debug was the model twice saying "my hypothesis is wrong".
- **sources:** DESIGN_diagnosis_loop(3).md#0; README_iterate_until_bugfix_notes.md; README_02_evidence_backed_proposals.md; README_overview.md
- **relations:** cite-or-abstain contract; real-bug eval gate; council pattern
- **verify-later:** eval-run artefacts; benchmark run records (runs 1–2 wrong CONFIRMED)

<!-- SOURCE: U16_docs019_design_plans.md -->
### B4a retrieval ceiling (symptom cannot reach infrastructure-layer causes)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN_diagnosis_loop(3) §1a "B4a (2026-06-17) measured… ALL of lexical, semantic, and fused scored 0.00" on the real gamesdesign fix symbols.
- **what:** Empirical measurement that one-shot retrieval from a symptom description cannot reach a cause living in shared infrastructure named for its function not its failure mode (resolveResultSpec/extractWorkflowResult): the symptom's words and the mechanism's words do not intersect. Lexical beat semantic on the mechanism-named task (0.50 vs 0.00). Consequence: embeddings did not earn a code-path place; the lever is iterative re-scoping following runtime evidence, not better retrieval. Retrieval seeds only the first scope.
- **sources:** DESIGN_diagnosis_loop(3).md#1a; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17 changelog); README_overview.md
- **relations:** evidence-follows re-scoping; text-vs-code embedding split (B4b); code_symbols index
- **verify-later:** contextkit eval_targets + groundtruth_targets.json; go_files/contextkit/{lex,sem}.json

<!-- SOURCE: U16_docs019_design_plans.md -->
### Evidence-follows re-scoping (call graph + runtime-named next scope)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Engine callgraph.go exists/tested per DESIGN(3) status; prompt rule 4 "Follow the evidence to the next scope — do not re-search the symptom".
- **what:** On REFUTED/UNVERIFIABLE the next bundle scopes the symbols/files the evidence names plus their call-graph neighbourhood (the analyser records `calls`), and prefers a runtime-named fault site over a retrieval-proposed one. This is the move retrieval cannot do; it reached the coordinator's result extraction in the real case — a symbol the symptom could never name.
- **sources:** DESIGN_diagnosis_loop(3).md#1a; PROMPT_diagnosis_verdict(1).md rule 4; NNN_fix_assemble_bundle_loop_scope_field.sql
- **relations:** B4a ceiling; convergence guards; Go analyser call graph
- **verify-later:** pkg/diagnose/callgraph.go; diagnose_route re-scope path

<!-- SOURCE: U16_docs019_design_plans.md -->
### Convergence guards (cap, narrow, grow, no-thrash)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN(3) status: "loop control, convergence guards … exists and is tested"; migration(19): "Four convergence guards + the no-citation→UNVERIFIABLE coercion all behaviour-tested".
- **what:** Deterministic Go guards so the loop cannot run forever or wander: iteration cap (5); scope must narrow (widening = not converging); evidence must grow (two iterations without new grounded evidence → stop with best-effort); no hypothesis thrash (oscillation without discriminating evidence → report both). Deliberately kept in tested Go, never re-expressed as workflow conditionals.
- **sources:** DESIGN_diagnosis_loop(3).md#3; PLAN_workflows_and_actions_migration(19).md; NNN_fix_diagnose_route_state_threading(1).sql
- **relations:** state-threading fix (guards were silently inert live); thin-workflows rule
- **verify-later:** pkg/diagnose/step.go DecideStep + tests

<!-- SOURCE: U16_docs019_design_plans.md -->
### Read-only, human-gated boundary of the diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN chassis integration(6) §7 "boundaries … do not relax in the chassis"; HANDOFF_fixloop(8) "Loop core is READ-ONLY by contract; sqlguard allowlists reads — keep it so".
- **what:** The loop gathers read-only (analyser, code_symbols, `\d`/capped SELECT/existing-log reads), proposes a diagnosis + suggested fix surface, and never applies fixes or triggers runs to test hypotheses. The human is kept at the two points that mattered: deciding the fix and backstopping the model's willingness to abandon a hypothesis. The F1 write surface is deliberately a separate agent with isolated credentials.
- **sources:** DESIGN_diagnosis_loop(3).md#4; DESIGN_diagnosis_loop_chassis_integration(6).md#7; HANDOFF_fixloop_thread(8).md#3
- **relations:** fix-implementer (the separate write surface); doc-drift read-only rule; three-guard read-only SQL
- **verify-later:** pkg/diagnose/sqlguard.go; spawn token-gate in spawn_actions.go

<!-- SOURCE: U16_docs019_design_plans.md -->
### Evidence tiers with freshness tagging (static / state / runtime)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Prompt rule 6 (tier + `fresh` per citation) in the applied workflow migration.
- **what:** Every citation is tagged static (code/schema), state (a DB row at a point in time) or runtime (log/work-item from an actual run), with observation time for state/runtime, so a verdict resting on stale evidence is visibly weak. Adapted from the doc-drift classifier's T1/T2/T3.
- **sources:** PROMPT_diagnosis_verdict(1).md rule 6; DESIGN_diagnosis_loop(3).md#2; DESIGN_doc_drift_classifier.md#2
- **relations:** doc-drift evidence tiers; misattribution asymmetry
- **verify-later:** verdict_wire.go citation struct

<!-- SOURCE: U16_docs019_design_plans.md -->
### Chassis realisation: diagnose_route workflow-driven loop (four diagnose actions)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Chassis-integration(6) banner (2026-06-24) "the BUILT design is `diagnose_route` … The actions that actually exist are diagnose_load_runtime, diagnose_assemble_bundle, diagnose_route, diagnose_emit"; live runs 8d488e01/73ed55c6 in fix migrations.
- **what:** On the chassis the loop is an agent workflow: analyse_repo → lookup_symbols (seeds iteration-1 scope from the symptom) → load_runtime → assemble_bundle → verdict (its own execute_llm_prompt step, observable) → diagnose_route (engine Advance: guards + call-graph re-scope, then a next_step override, the conditional_route pattern) → back to load_runtime or emit. The loop returns to load_runtime, not assemble, so the prior verdict's data_requests run and runtime re-gathers each iteration. Gather reuses existing actions (request_repo_analysis→analyse_repo_local, lookup_code_symbols, execute_llm_prompt) per the STEP-ZERO reuse audit; only the bundle composer was genuinely new.
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md#0,#status; NNN_fix_diagnose_agent_workflow(2).sql; PLAN.md; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17 entry)
- **relations:** abandoned diagnose_run and diagnostician designs; one-decision-core two realisations
- **verify-later:** platform/orchestration/actions/diagnose_*_action.go; coordinator.go getNextStepFromResult; registry.go Category "diagnose"

<!-- SOURCE: U16_docs019_design_plans.md -->
### Abandoned design: diagnose_run single engine-wrapping action
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** Chassis-integration(6) banner: "the §4–§6 `diagnose_run` recommendation below is the ABANDONED path … there is no `diagnose_run` action"; the seeded workflow referencing it was rewritten by NNN_fix_diagnose_agent_workflow.
- **what:** The originally recommended shape — one `diagnose_run` action calling `diagnose.Run()` with an injected Verdicter, keeping the whole loop inside a single step. Dropped in favour of the workflow-driven observable loop (verdict as its own step, router action). A prompt-registry reference `diagnose-verdict-v1` belonged to this design and is also unused; the prompt went inline instead. Kept here because seeded rows briefly referenced the non-existent action (a real incident class: workflow names an action that does not exist).
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md banner,#4-6; NNN_fix_diagnose_agent_workflow(2).sql header; NNN_move_diagnose_workflow_to_default_config(1).sql banner
- **relations:** superseded by diagnose_route realisation
- **verify-later:** absence of diagnose_run in registry.go

<!-- SOURCE: U16_docs019_design_plans.md -->
### Abandoned design: diagnostician per-iteration re-invocation (spawn-next chain)
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** NNN_seed_diagnose_agents(2).sql banner "SUPERSEDED — DO NOT APPLY … kept only as a record of the re-invocation design that was considered and dropped."
- **what:** A third loop shape: each orchestration runs ONE iteration (load_runtime → analyse → lookup → assemble → verdict → route → conditional), and on continue spawns+calls a fresh `diagnostician` of the same type with revised hypothesis/scope/iteration in input_data, the terminal verdict bubbling up the child chain. Motivated by doubt that the engine supported a workflow-internal cycle and by the build-dispatch-loop one-unit-per-orchestration precedent. Dropped once the next_step-override loop-back was confirmed to work.
- **sources:** NNN_seed_diagnose_agents(2).sql header + workflow body
- **relations:** superseded by diagnose_route loop-back; build-dispatch-loop pattern
- **verify-later:** no `diagnostician` row in agent_definitions

<!-- SOURCE: U16_docs019_design_plans.md -->
### One decision core, two realisations (Run vs Advance/DecideStep)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** DESIGN(3) §1 note: "Both share ONE decision core — step.go's pure DecideStep … advance_test.go proves Advance threaded across iterations … reproduces Run()".
- **what:** The standalone dev harness (`internal/diagnose/loop.go` Run, a Go for-loop with inline IO) and the chassis workflow loop share the same pure per-iteration decision function, with `advance.go`'s Advance exposing it statefully to the chassis; equality is proven by test. cmd/diagnose stays the dev/test harness (scripted verdicts, dry-bundle), never a production entrypoint.
- **sources:** DESIGN_diagnosis_loop(3).md#1; DESIGN_diagnosis_loop_chassis_integration(6).md#status,#3; PLAN_workflows_and_actions_migration(19).md
- **relations:** engine/harness file-placement split; travelling contextkit module
- **verify-later:** pkg/diagnose/advance.go + advance_test.go

<!-- SOURCE: U16_docs019_design_plans.md -->
### Model-written data_requests under three-guard read-only SQL
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** PLAN.md DONE: "Three-guard model-SQL feature: Guard 1 (prompt), Guard 2 (IsReadOnlySQL), Guard 3 (read-only transaction, confirmed on the chassis side)"; GATED item 1: execution wiring in diagnose_load_runtime is deploy-side and untested; README_flows: "terminal-verdict data_requests that never run".
- **what:** The verdict may request specific evidence as single read-only SELECTs (`data_requests: [{sql, why}]`), defended by three independent guards: the prompt contract, a Go lint (sqlguard.IsReadOnlySQL), and execution inside a read-only transaction with statement timeout; the harness analogue is a GRANT-based SELECT-only `diagnose_ro` role (not default_transaction_read_only, unreliable under pgbouncer). Notably this reversed an earlier stance — chassis-integration(6) recorded "the model never writes SQL" and called runDataRequests dormant/beyond the boundary; the bounded, guarded version was then built deliberately.
- **sources:** PROMPT_diagnosis_verdict(1).md rule 7; NNN_create_diagnose_ro_role.sql; PLAN.md; DESIGN_diagnosis_loop_chassis_integration(6).md#status (the earlier stance)
- **relations:** read-only boundary; self-verification in the council (same move at review time)
- **verify-later:** pkg/diagnose/sqlguard.go in chassis; diagnose_load_runtime data-request execution path; diagnose_ro role existence

<!-- SOURCE: U16_docs019_design_plans.md -->
### Real-bug evaluation gate (scaffold correct ≠ reasons well)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** PLAN.md GATED 5: "THE EVAL GATE … MUST reproduce the mid-course REVERSALS and ABSTAIN when unsettled … No automatic triggering until this passes"; README_02 shows benchmark runs happened (runs 1–2 confidently wrong) and design responded.
- **what:** Before any unsupervised or automatic triggering, the live loop must be run against known bugs — the gamesdesign two-fault bug (with its captured reversals) and the 016 §9 silent-no-op catalogue — and must reproduce hypothesis reversals and abstain when evidence does not settle, rather than confirming first guesses. "Compiling isn't behaving" is the standing lesson; a loop that confirms its first guess every time is the failure mode dressed as success.
- **sources:** DESIGN_diagnosis_loop(3).md#6; PLAN.md GATED; README_whats_next.md; README_02_evidence_backed_proposals.md
- **relations:** gamesdesign bug fixture; falsification-first; later trigger modes gated on this
- **verify-later:** eval run records; whether triggers (b)/(c) were ever enabled

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnose agent pair + generic-request trigger envelope
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Restore/fix migrations applied against live rows (snapshot ids 34f4afc8/e8e96d24 in HANDOFF_builder_thread); TRIGGER pattern proven (084_TRIGGER_diagnose_v1.sh referenced as canonical in HANDOFF_fixloop(8)).
- **what:** A thin diagnose-orchestrator (spawn_agent → call_agent → complete, per "every agent is an orchestrator" and the wrapper rule for substantive work) spawns the diagnose-agent worker pod that runs the loop. Triggering is the existing generic-request envelope — kcat to system.agent.generic.requests with agent_type diagnose-orchestrator and input_data {symptom, seed_scope, runtime_site, …} — no new triggering code; later triggers (on build failure, proactive sweep) are the same envelope from a different sender, gated on the eval. Sub-agents reply on the caller's responses topic.
- **sources:** DESIGN_diagnosis_loop_chassis_integration(6).md#2; NNN_restore_diagnose_orchestrator_workflow(1).sql; PLAN_workflows_and_actions_migration(19).md
- **relations:** wrapper-orchestrator pattern; index-orchestrator (same pattern reused)
- **verify-later:** agent_definitions diagnose-orchestrator/diagnose-agent; drafts/084_TRIGGER_diagnose_v1.sh

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnose loop-back plumbing fault class (state threading, scope encoding)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Both fix migrations are operative live fixes with run evidence (run 8d488e01: "there is a 'route' key, no 'diagnose_state' key"; trail truncated to 1 entry).
- **what:** Two producer/consumer field mismatches that left the loop running but silently degraded: (1) diagnose_route read its LoopState from bare `diagnose_state` while its own output lands under `route.diagnose_state` — so the loop re-seeded every iteration, never enforcing max_iterations, truncating the evidence trail, and resetting the cross-iteration guards; (2) route.scope is EncodeScope's untagged-struct JSON, so the string-list reader needed `route.scope.Symbols` (capital S) — before the fix every re-scope silently fell through to the fallback chain and iterations 2+ never moved scope. Both were invisible-success faults: the loop "worked" while its defining features were inert.
- **sources:** NNN_fix_diagnose_route_state_threading(1).sql; NNN_fix_assemble_bundle_loop_scope_field.sql; DESIGN_diagnosis_loop_chassis_integration(6).md#status (the round-trip flagged as unverified)
- **relations:** workflow-variable-sync rule; result-contract dead-key class; convergence guards
- **verify-later:** diagnose_route_action.go default state_field; Scope struct json tags

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnosis persistence + documented intake (diagnosis_artifacts, needs_diagnosis)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_fixloop(8): "First action: slice F0.1 with pre-registered criteria — (1) diagnosis_artifacts migration … (2) assemble-action write-through … (3) the needs_diagnosis envelope"; decisions recorded 2026-07-07, not yet built in these files.
- **what:** F0 of the fix loop: make each iteration's bundle durably fetchable and add per-iteration running notes — a `diagnosis_artifacts` table (correlation_id, iteration, kind ∈ {bundle, iteration_note}, body, retention knob per kind) written through from the assemble action Go-side (no workflow-shape change); plus a documented intake route: a `needs_diagnosis` envelope / pipeline='diagnose' work item carrying subject_type/subject_key with null-site allowed. Bundle egress via completion payloads is bounded (max_response_bytes) — persist and reference, don't ship megabytes.
- **sources:** HANDOFF_fixloop_thread(8).md#4; HANDOFF_fixloop_thread(8).md#3
- **relations:** result-contract size guard; travelling-docs pattern (notes per iteration); work-item relay
- **verify-later:** diagnosis_artifacts table existence; persist_note step in diagnose-agent workflow (tools thread's wiring)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Verdict-quality wrinkles + dead code-retrieval channel (measured)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** README_flows: "seed similarities in the 0.55 band, no page-build symbols … a stale line carried into a conclusion that its own citation contradicts, and terminal-verdict data_requests that never run".
- **what:** Post-run findings on the live loop: the lookup channel contributes nothing measurable (work is on the query side — seed the lookup from runtime evidence or expand the query, a self-contained lookup_symbols change); the trigger's site_id is intermittent across runs (reproducibility); and two verdict-quality defects point at the confirm/emit step (a conclusion contradicted by its own citation; data_requests emitted on terminal verdicts that never execute).
- **sources:** README_flows.md; PLAN.md GATED
- **relations:** B4a ceiling; data_requests wiring; eval gate
- **verify-later:** lookup_symbols seeding config; emit/confirm step handling of terminal data_requests

<!-- SOURCE: U17b_docs019_gofiles.md -->
### contextkit CLI toolkit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "go build ./...  # compiles all seven commands" (README.md); real invocation example wired to a live site in cmd/bundle/README.md (`-runtime-site gamesdesign.co.uk`).
- **what:** A standalone Go module (`module contextkit`, go 1.22) of CLI tools for building LLM context bundles from a repo without a live cluster: analyser, assembler, dbcontext, bundle, embed, resolve_targets, fuse, eval_targets, dedup, thin_versions. Compiles and runs independently of the agentchassis repo; two of its packages (`internal/analysis`) are shared verbatim with the chassis.
- **sources:** contextkit/README.md, contextkit/README(2).md, contextkit/go.mod
- **relations:** diagnosis loop (internal/diagnose), analyser-adapter deployment plan, thin-slice constitution
- **verify-later:** does `internal/analysis` in this tree still match `internal/analysis/` at the agentchassis repo root byte-for-byte (README.md flags this as a manual sync obligation, not automated)

<!-- SOURCE: U17b_docs019_gofiles.md -->
### analyser (cmd/analyser)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "thin CLI wrapper over analysis.AnalyseWithExclude — the same parsing primitive the in-cluster analyser adapter imports"
- **what:** Walks a Go source tree and emits a structural-summary JSON (files, packages, imports, function/method signatures with callee names, struct/interface declarations with line ranges). Always skips vendor/, testdata/, hidden dirs, `*_test.go`, and `*(N).go` download-duplicates; takes an `-exclude` list for repos (like this one) that store archived copies of their own code under docs/.
- **sources:** contextkit/cmd/analyser/main.go#header, contextkit/README.md
- **relations:** internal/analysis package, code-indexer agent (chassis-side counterpart), embed/resolve_targets (consume analyser JSON)
- **verify-later:** internal/analysis (agentchassis repo root) — confirm the in-cluster analyser adapter still calls `analysis.Analyse` (no-exclude) as documented

<!-- SOURCE: U17b_docs019_gofiles.md -->
### internal/analysis package (analyser output contract)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Package analysis defines the contract the analyser emits and the assembler, embed, and resolve_targets consume... defined only here" (types.go header); "the harness and production parse identically" (analyse.go header)
- **what:** The single-source-of-truth `Output`/`FileInfo`/`FuncDef`/`TypeDef` contract for repo structural analysis, plus `Analyse`/`AnalyseWithExclude` (the layer-1 AST walk) and `ReadSymbolBody` (slices a `path:Symbol` scope into source text using the analyser's recorded line span, never re-parsing). Intentionally Go-only; a non-Go producer would fill the same contract behind the analyser adapter.
- **sources:** contextkit/internal/analysis/types.go#header, contextkit/internal/analysis/analyse.go#header, contextkit/internal/analysis/symbolbody.go#header
- **relations:** analyser, assembler, embed, resolve_targets, cmd/bundle (also uses the symbol slicer), chassis diagnose_assemble_bundle action
- **verify-later:** whether the chassis's diagnose_assemble_bundle action's old inline `readSymbolBody` stub has actually been replaced by a call to `ReadSymbolBody` as the header claims is the intent

<!-- SOURCE: U17b_docs019_gofiles.md -->
### assembler (cmd/assembler)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "That's the whole thin slice working: analyser, constitution, assembler." (README_how_to_run_analyser.md)
- **what:** Builds one paste-ready markdown bundle for a single task: consumes the analyser JSON, the repo (to pull full bodies by line range), the flat constitution, and a task+scope spec; renders constitution, task, in-scope code in full, neighbourhood signatures (same-package, capped ~60/package), schema (hand-fed), and a pointers note of what was omitted. `-step` (framing/implementation/debug) controls altitude — framing shows signatures only, implementation/debug show full in-scope bodies.
- **sources:** contextkit/cmd/assembler/main.go#header, contextkit/README_how_to_run_analyser.md, contextkit/001_more_potential_thin_slice_prompt.md
- **relations:** internal/analysis (symbol slicing), thin_slice_constitution.md, bundle (wraps it), docselect/queryselect (chassis analogues for doc/query selection instead of hand-specified scope)
- **verify-later:** confirm the neighbourhood-signature cap and package-scoping behaviour match what 001_more_potential_thin_slice_prompt.md's design notes describe

<!-- SOURCE: U17b_docs019_gofiles.md -->
### dbcontext (cmd/dbcontext)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** used directly in cmd/bundle/README.md's worked example (`-psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db'`)
- **what:** Fetches DB context for a bundle by shelling out to a configurable `psql` — schema (`\d <table>`, complete+bounded), rows (multipass-sized SELECT: full if within cap, else sample + pointer query, never an unbounded dump), and capabilities (`\dx`, `\df`). No Go DB driver; psql does the talking, so it inherits whatever connection role/permissions the operator supplies.
- **sources:** contextkit/cmd/dbcontext/main.go#header
- **relations:** bundle (wraps it), sqlguard (lints model-written queries elsewhere in the pipeline), database-and-infrastructure conventions
- **verify-later:** whether the psql connection used in production is provisioned as a read-only role (sqlguard.go explicitly says the lint alone is not the safety boundary — the read-only transaction/role is)

<!-- SOURCE: U17b_docs019_gofiles.md -->
### bundle (cmd/bundle)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** concrete real invocation against `gamesdesign.co.uk` in cmd/bundle/README.md; used by bundle_diagnosis_loop.sh, bundle_minilobby_trim2/3.sh, bundle_recreation_v1.sh
- **what:** A thin orchestration wrapper around dbcontext + assembler: gathers read-only DB context (schema/capabilities/runtime evidence), writes each to a temp file, then invokes the assembler with those files wired in. Deliberately never runs SQL itself (that stays in dbcontext) so the assembler stays a pure, read-only, offline composer — the wrapper "triggers NOTHING — no builds, no spawns, no writes."
- **sources:** contextkit/cmd/bundle/main.go#header, contextkit/cmd/bundle/README.md
- **relations:** dbcontext, assembler, gatherer.go (BundleGatherer shells out to this exact binary from the diagnosis loop)
- **verify-later:** BundleGatherer.buildArgs (gatherer.go) — confirm the flag set it constructs still matches this binary's real flags

<!-- SOURCE: U17b_docs019_gofiles.md -->
### embed (cmd/embed)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "-local ... proves the pipeline (index → cosine → rank) WITHOUT a model, but is NOT semantic — use -ollama for real recall" (header)
- **what:** Builds/queries a semantic vector index over the analyser's symbols — the recall layer for target resolution sitting above the lexical baseline (resolve_targets). Model-agnostic via an embedder interface: `-ollama` (real embeddings, e.g. nomic-embed-text) or `-local` (deterministic offline token-hashing stand-in for pipeline-proving only). Index and query must use the same embedder/vector space.
- **sources:** contextkit/cmd/embed/main.go#header
- **relations:** resolve_targets, fuse (RRF-merges embed's output with resolve_targets'), eval_targets (scores it), code-indexer agent (chassis-side embedding via the same ollama-adapter/nomic-embed-text pairing)
- **verify-later:** whether production bundle-building actually runs `embed` with `-ollama` or still relies on the `-local` stand-in

<!-- SOURCE: U17b_docs019_gofiles.md -->
### resolve_targets (cmd/resolve_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the deterministic baseline — the layer that runs before any embeddings" (header)
- **what:** A first-cut, lexical-overlap target resolver: given a task string and the analyser JSON, proposes ranked candidate symbols/files to `-scope` by matching the task's distinctive words against each symbol's name, path, and docstring. Does not decide — proposes a ranked candidate set for a human or the assembler to confirm.
- **sources:** contextkit/cmd/resolve_targets/main.go#header
- **relations:** embed (semantic counterpart), fuse (merges both), internal/candidates (shared output contract), eval_targets
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### fuse (cmd/fuse)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "score(item) = sum over lists of 1/(k + rank_in_list), k=60 (standard)" (header)
- **what:** Merges ranked candidate lists (resolve_targets' lexical output + embed's semantic output) into one ranking via reciprocal-rank fusion (RRF). Combines by RANK not score specifically because the lexical integer scores and semantic cosine scores aren't on a comparable scale.
- **sources:** contextkit/cmd/fuse/main.go#header
- **relations:** resolve_targets, embed, internal/candidates, eval_targets (scores fuse's output too)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### eval_targets (cmd/eval_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** the ground-truth file contains a task tied to a "REAL fix (2026-01 chassis regression)" already applied, showing the harness is exercised against real cases, not just synthetic ones
- **what:** Scores a resolver's candidate list (`-json` output of resolve_targets/embed/fuse) against a ground-truth set mapping tasks to the symbols they actually needed — turns "the fused list looks better" into numbers: recall@N over decisive symbols, and MRR contribution (rank of first decisive hit). Match is on `path:name`.
- **sources:** contextkit/cmd/eval_targets/main.go#header, contextkit/groundtruth_targets.json
- **relations:** resolve_targets, embed, fuse (all scored by this), llm-quality-testing (evaluation-harness pattern), ground-truth eval set concept below
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### diagnose (cmd/diagnose)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "THE VERDICT STEP IS NOT THE REAL MODEL HERE... a chassis-side follow-on (needs a model). This entrypoint ships two stand-ins" (header)
- **what:** Wires the diagnosis-loop scaffold (internal/diagnose) to real adapters — BundleGatherer (shells to cmd/bundle, read-only) and AnalysisCallGraph (follows the analyser's `calls` for re-scope). The verdict step is stubbed (either a scripted JSON array of verdicts for testing, or a trivial always-UNVERIFIABLE default) since the real cite-or-abstain LLM verdicter needs a model and lives chassis-side. Explicitly read-only and human-gated: emits a diagnosis + evidence trail, never a fix, never a triggered run.
- **sources:** contextkit/cmd/diagnose/main.go#header
- **relations:** internal/diagnose (loop.go, step.go, verdict_wire.go, callgraph.go, gatherer.go), fixloop workstream (the diagnose→fix pipeline this scaffold feeds)
- **verify-later:** docs024_key_docs_latest/fixloop_eg_dartsonline/ for whether/how a real LLM verdicter has since been wired in chassis-side

<!-- SOURCE: U17b_docs019_gofiles.md -->
### diagnosis-loop scaffold (internal/diagnose, loop.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "WHAT LIVES HERE (deterministic, testable without a model): loop control, the guards..., the evidence trail, and the re-scope mechanism. WHAT DOES NOT... the Verdict step" (loop.go header); backed by loop_test.go, loop_datarequest_test.go, loop_scopeguard_test.go
- **what:** The deterministic core of the diagnosis loop: wraps a read-only gather step around a pluggable verdict step, enforces convergence guards (iteration cap, scope-must-narrow, evidence-must-grow, no-thrash), accumulates an evidence trail, and re-scopes by FOLLOWING runtime/call-graph evidence rather than re-searching the symptom — named as the fix for a "ceiling" where symptom-only retrieval fails on infrastructure-layer causes. Non-negotiable boundary: never applies a fix, never triggers a run to test a hypothesis.
- **sources:** contextkit/internal/diagnose/loop.go#header, contextkit/internal/diagnose/loop_scopeguard_test.go#header, contextkit/internal/diagnose/loop_datarequest_test.go#header
- **relations:** step.go (DecideStep), advance.go (chassis-facing wrapper), callgraph.go, verdict_wire.go, docselect.go, queryselect.go, sqlguard.go, gatherer.go, fixloop workstream
- **verify-later:** whether the "guard-vs-expansion" bugfix noted in loop_scopeguard_test.go (run 17933a83) and the data_request evidence-growth fix (loop_datarequest_test.go, "truncated the live gamesdesign runs at iteration 3") are reflected in the currently-deployed chassis diagnose_run/diagnose_route actions

<!-- SOURCE: U17b_docs019_gofiles.md -->
### DecideStep — shared pure per-iteration decision (step.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Extracting it keeps ONE source of truth for the guard + re-scope logic instead of two copies that could drift. Run() is refactored to call Step(); the existing tests are the proof the behaviour is unchanged." (header)
- **what:** The per-iteration decision (given iteration state, a verdict, the call graph, and guard memory) extracted as a pure function, shared by the standalone `Run()` loop and the chassis `diagnose_run` workflow action (where the verdict is a separate workflow step). Guarantees one source of truth instead of two logic copies that could drift apart.
- **sources:** contextkit/internal/diagnose/step.go#header, contextkit/internal/diagnose/step_test.go
- **relations:** loop.go, advance.go (LoopState calls this per-iteration)
- **verify-later:** confirm the chassis `diagnose_run` action actually calls this shared `Step()`/`DecideStep` rather than a re-implementation

<!-- SOURCE: U17b_docs019_gofiles.md -->
### LoopState — chassis-facing per-iteration API (advance.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "the chassis needs: (1) a LoopState it can thread through workflow collected_data between iterations, (2) a single Advance() call per iteration, and (3) parse helpers for a verdict that arrives as an already-unmarshalled map" (header)
- **what:** The workflow-driven realisation of the loop: since the chassis loop is `gather → verdict step → diagnose_route → back | emit` (not an in-process loop), LoopState carries loop memory across workflow steps via `collected_data`, with `Advance()` as the one call per iteration and `EncodeLoopState`/`DecodeLoopState` for the JSON round-trip. Adds no new decision logic beyond step.go's DecideStep plus state bookkeeping.
- **sources:** contextkit/internal/diagnose/advance.go#header, contextkit/internal/diagnose/advance_test.go
- **relations:** step.go, loop.go, chassis diagnose_route workflow step
- **verify-later:** platform/orchestration — the actual `diagnose_route` step and its `collected_data` schema, to confirm it matches `EncodeLoopState`'s shape

<!-- SOURCE: U17b_docs019_gofiles.md -->
### AnalysisCallGraph — call-graph re-scope (callgraph.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "HONEST LIMIT (from the analyser): `calls` is NAME-BASED, not type-resolved... DELIBERATELY DROPS ubiquitous names (Run, String, Error, …)" (header)
- **what:** A CallGraph implementation backed by the analyser's recorded (name-based, not type-resolved) `calls` field, letting the diagnosis loop re-scope by following the call graph from an evidence-named site rather than re-searching the symptom. Explicitly drops ubiquitous method names that would otherwise explode the neighbourhood into noise — the loop's narrowing guard is the backstop, but dropping known-ubiquitous names keeps re-scope sharp at the source.
- **sources:** contextkit/internal/diagnose/callgraph.go#header
- **relations:** internal/analysis (the `calls` data it consumes), loop.go's re-scope mechanism, cmd/diagnose (wires this in as the real adapter)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### verdict wire format (verdict_wire.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "If the prompt's output schema and this struct ever drift, the loop breaks at this join — so this file is tested against example model outputs." (header)
- **what:** The JSON wire format the (LLM) Verdicter emits — human-legible strings (`"REFUTED"`, `"runtime"`) and snake_case keys rather than the domain type's int enums, so a model can produce it reliably — and the parser (`ParseVerdict`) translating it into the domain `Verdict`. Named as the ONE seam between the verdict prompt's specified output (`docs/PROMPT_diagnosis_verdict.md`) and the scaffold; a verdict-script in this format is a faithful stand-in for the real model.
- **sources:** contextkit/internal/diagnose/verdict_wire.go#header, contextkit/internal/diagnose/verdict_wire_test.go
- **relations:** diagnose (cmd), loop.go, docs/PROMPT_diagnosis_verdict.md (referenced, not in this unit)
- **verify-later:** docs/PROMPT_diagnosis_verdict.md — confirm its schema still matches this struct (the header itself flags drift risk here)

<!-- SOURCE: U17b_docs019_gofiles.md -->
### docselect — per-hypothesis doc selection (docselect.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Selection is DETERMINISTIC and testable... so it can be exercised without a model. It is a HEURISTIC (keyword/path substring)" (header); docselect_test.go exercises keyword/always/path-glob rules
- **what:** Per-hypothesis selection of authored context docs (the 003 contract sections, 016 §9 entries, dev-guide sections) to paste into the CURRENT iteration's bundle rather than every doc into every bundle — avoiding the "irrelevant context buries the signal" failure mode. A future extension is floated (not built): letting the verdict NAME a needed doc via a `needed_docs` field mirroring `needed_evidence`/`next_scope`.
- **sources:** contextkit/internal/diagnose/docselect.go#header, contextkit/internal/diagnose/docselect_test.go
- **relations:** thin_slice_constitution.md (the always-on layer this supplements), queryselect.go (data analogue), contracts-and-standards (003)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### queryselect — vetted read-only query catalogue (queryselect.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "WHY A CATALOGUE, NOT MODEL-WRITTEN SQL (the safety boundary)... the queries are HAND-WRITTEN, parameterised, and \\d-verified ONCE; the loop only SELECTS among them by hypothesis. The model never writes SQL." (header)
- **what:** Per-hypothesis selection of vetted, read-only, parameterised DB queries for the runtime-evidence gather — the data analogue of docselect.go. Queries bind to the loop's existing context (site_id, domain, page, correlation_id already in input_data/seed), so no wire-format change or model-supplied SQL parameters are needed. This is presented as THE safety boundary for runtime evidence, distinct from sqlguard's lint-only role.
- **sources:** contextkit/internal/diagnose/queryselect.go#header, contextkit/internal/diagnose/queryselect_test.go
- **relations:** docselect.go, sqlguard.go, dbcontext (executes the chosen queries)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### sqlguard — IsReadOnlySQL lint (sqlguard.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "READ THIS FIRST — what this is NOT: This is NOT the safety boundary... The REAL guarantee is the EXECUTION SUBSTRATE" (header); three dedicated test files including a literal-false-positive regression (sqlguard_literal_test.go)
- **what:** A cheap pre-flight lint for model-written diagnosis queries, explicitly documented as defence-in-depth, NOT the safety guarantee — the real guarantee is the execution substrate (chassis: read-only transaction + non-multi-statement protocol; harness: a read-only DB role) plus a statement_timeout. Includes a regression fix for keywords/`;` appearing inside quoted string literals (triggered by a real page slug `tool-drop-rate-simulator` containing "drop").
- **sources:** contextkit/internal/diagnose/sqlguard.go#header, contextkit/internal/diagnose/sqlguard_literal_test.go#header
- **relations:** queryselect.go (the actual safety boundary via hand-vetted catalogue), dbcontext
- **verify-later:** confirm the chassis execution path really does use `db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})` as claimed, and that the harness's `-psql` role is genuinely read-only in practice

<!-- SOURCE: U17b_docs019_gofiles.md -->
### BundleGatherer (gatherer.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "READ-ONLY by construction (DESIGN §4): bundle runs dbcontext... + the pure assembler. Nothing here triggers a build, spawn, or write." (header)
- **what:** A Gatherer that shells out to `cmd/bundle` to produce each iteration's bundle, translating a `Scope` into bundle flags and returning the written bundle path. Adds no capability beyond what `cmd/bundle` already does — just drives it per iteration with the loop's evolving scope.
- **sources:** contextkit/internal/diagnose/gatherer.go#header
- **relations:** cmd/bundle, cmd/diagnose (wires this in as the real gatherer)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### ranked-candidate contract (internal/candidates)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Defined once here so the shape isn't re-declared as `candFile`/`jc` in each tool." (header)
- **what:** The shared `Candidate`/`File` JSON contract (`path`, `name`, `kind`, `score` as float64, `rank`, `task`, `method`) that resolve_targets, embed, and fuse all emit with `-json`, and that fuse and eval_targets read — replacing what used to be duplicated per-tool struct definitions.
- **sources:** contextkit/internal/candidates/types.go#header
- **relations:** resolve_targets, embed, fuse, eval_targets
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### ground-truth eval set for target resolution
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** task `silent-norebuild-resultspec` is tagged "REAL fix (2026-01 chassis regression): coordinator honoured only plural output_fields; writer declares singular output_field; compiled page collapsed into the skipPatterns dump -> 'completed' stub -> stale page. Fix = resolveResultSpec treats singular as flatten."
- **what:** `groundtruth_targets.json` maps tasks (deliberately symptom-only task strings, scrubbed of the vocabulary a resolver could trivially match — see the note on an earlier version leaking "extracted"/"result" and inflating a lexical rank) to the "expect" (decisive) and "also_useful" symbols a resolver must surface. Grew across versions: the `.orig` predecessor holds only the `skinner-box` task; the current file adds `silent-norebuild-resultspec`, drawn from a real, already-fixed chassis regression (result-spec singular vs plural output_field handling).
- **sources:** contextkit/groundtruth_targets.json, contextkit/groundtruth_targets.json.orig
- **relations:** eval_targets, resolve_targets/embed/fuse (evaluated against this set), llm-quality-testing
- **verify-later:** platform code for `result_spec.go:resolveResultSpec` / `coordinator.go:extractWorkflowResult` — confirm the fix described is actually live

<!-- SOURCE: U17b_docs019_gofiles.md -->
### code-indexer agent (analyser-adapter's chassis-side counterpart)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** "DRAFT — modelled on the real agent_definitions rows you sent... Confirm the live schema before applying" (NNN_create_code_indexer_agent.sql); status column set to `'experimental'` in the INSERT itself
- **what:** A draft `agent_definitions` row for a `code-indexer` orchestrator agent: workflow is `request_analysis` (calls `request_repo_analysis` action, asking the analyser adapter to parse a repo@ref into symbols) → `index_symbols` (calls `index_code_symbols`, upserting into `code_symbols`, embedding changed symbols via an ollama/nomic-embed-text endpoint, pruning symbols absent from the commit) → `complete`. Retrieval side is a separate `lookup_code_symbols` action used by other agents. Coordination-only orchestrator; the real parsing work happens in the analyser-adapter pod.
- **sources:** NNN_create_code_indexer_agent.sql
- **relations:** analyser (the parsing primitive this indexes), embed (same embedder pairing: ollama + nomic-embed-text), analyser-adapter deployment plan, snapshot-before-mutate practice
- **verify-later:** `agent_definitions` table (`\d agent_definitions`) for the real CHECK constraint on `agent_category` and NOT NULL/default columns before this migration is applied; whether `code_symbols`, `index_code_symbols`, `lookup_code_symbols`, `request_repo_analysis` exist yet

<!-- SOURCE: U17b_docs019_gofiles.md -->
### action-name-to-file resolver (bundle_recreation_v1.sh)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "v1 grepped only for the quoted name inside action files and excluded registry.go — which is precisely where the mapping lives. File names are not consistent either: validate_page_content.go has no _action suffix." (header)
- **what:** A path-resolution helper that, given a registered action NAME, finds its source FILE by (1) reading the registration line in `registry.go` to get the constructor/type name, (2) finding the file defining that constructor/type, (3) falling back to CamelCasing the action name and searching, (4) last-resort whole-platform-tree search — built specifically because file naming is inconsistent (some action files lack the `_action` suffix) and a prior version's naive grep missed paths by excluding the one file (`registry.go`) where the authoritative name→type mapping actually lives.
- **sources:** contextkit/bundle_recreation_v1.sh#header
- **relations:** bundle, resolve_targets (a cruder, deterministic alternative to lexical/semantic resolution for a KNOWN action name)
- **verify-later:** —

<!-- SOURCE: U17b_docs019_gofiles.md -->
### dogfooding bundle for building the diagnosis loop itself (bundle_diagnosis_loop.sh)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** "CONFIRM BEFORE RUNNING (flagged — I could not verify these from the mounted files; only the contextkit engine .go files were available)... the four diagnose actions are DRAFTS (chassis-drafts/). If they are not yet committed to ~/projects/agentchassis AND re-analysed into chassis_clean.json, cmd/bundle will SKIP those -scope entries" (header)
- **what:** A read-only bundle recipe whose SUBJECT is the diagnosis loop's own code (its decisive symbols + the four diagnose actions + governing docs + the constitution), for continuing the loop's own gated build in a fresh chat/sub-agent without re-reading the whole tree — a self-referential use of the tool it is building context about. Self-flags an unverified assumption: the four action files may only exist as drafts not yet analysed into the chassis index.
- **sources:** contextkit/bundle_diagnosis_loop.sh#header
- **relations:** diagnosis-loop scaffold, bundle, cmd/diagnose
- **verify-later:** whether the "four diagnose actions" referenced are now committed to agentchassis proper (outside chassis-drafts/)

## Proposed NEW categories
None — all 30 concepts fit existing taxonomy slugs: `diagnosis-loop` (23), `documentation-system` (5), `adapters` (1), `database-and-infrastructure` (1), `content-governance` (1), `development-guide` (1). (Counts overlap because some concepts touch two slugs; each was filed under its single best-fit home per the tagging rules.)

<!-- SOURCE: U18_sql_for_agents.md -->
### Diagnosis loop agents (diagnose-orchestrator / diagnose-agent)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** 122 seeds both as 'experimental' ("until the real-bug evaluation gate passes; promote to 'active' after"); 126–129 wire persistence and subject threading; incidents in 127/128 show live runs 2026-07-06/10.
- **what:** Read-only diagnosis: hypothesise → gather scoped evidence (code + runtime) → cite-or-abstain verdict → re-scope by following evidence; emits a diagnosis + evidence trail for a human, never changes code. Loop CONTROL lives in the Go engine (diagnose_run), not workflow conditionals; gather steps stay explicit for log visibility. Wrapper-mandated (substantive in-chassis LLM work). Runtime evidence is an optional bundle tier — error routing makes anchorless (code-only) runs survive.
- **sources:** 122_diagnose_agents.sql; 126_wire_persist_diagnosis_note.sql; 127_diagnose_load_runtime_error_step.sql; 129_wire_diagnosis_subject_threading.sql
- **relations:** code-indexer supplies code_symbols retrieval; travelling docs receive diagnosis notes; docs019 diagnosis programme
- **verify-later:** diagnose_run engine; promotion to active; evaluation gate results

<!-- SOURCE: U18_sql_for_agents.md -->
### code-indexer (repo → code_symbols for the analyser)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** 118 marked DRAFT/experimental but with a checked label convention "[checked 2026-06-11: composition implemented in IndexCodeSymbolsAction]".
- **what:** Orchestrator that asks the analyser adapter to parse a repo at a ref into symbols, then index_code_symbols upserts them (embedding changed symbols, pruning absent ones). repo label is composed as "owner/repo" from the analyser reply so labels always match what was fetched; retrieval side is lookup_code_symbols used by diagnosis agents. Non-git corpora may override repo (e.g. 'domain:kruste.com').
- **sources:** 118_code_indexer_for_analyser.sql
- **relations:** diagnose-agent evidence gathering; analyser adapter; docs019 contextkit
- **verify-later:** code_symbols table; agent status live

<!-- SOURCE: U19_sql_tables_components.md -->
### code_symbols per-repo code index (context tool)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "CONFIRMED (2026-06-09, clients_db): \dt found no existing *code*/*symbol* table, and \dx shows vector 0.8.0... Both gates pass — HNSW stands."
- **what:** The context tool's code index: one row per symbol keyed (repo, path, symbol) with kind CHECK (func/method/struct/interface/alias/type/var/const), signature/doc/line range (bodies read from the repo at commit_sha, not stored), content text that is both embedded (HNSW cosine, chosen over IVFFlat for incremental churn) and trigram-matched, content_hash to skip re-embedding unchanged symbols. Deliberate departures flagged: no version/soft-delete — a rebuildable cache versioned by commit_sha, pruned by hard delete. Ships the full usage contract in comments: indexing upsert, prune, semantic/lexical retrieval, and hybrid RRF fusion in SQL (constant 60) replacing in-Go fuse.
- **sources:** docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql
- **relations:** knowledge_base shape reuse; diagnosis-loop code retrieval; contextkit.
- **verify-later:** indexing workflow; code_symbols row counts per repo.

<!-- SOURCE: U23_docs_root_vonc.md -->
### cmd/bundle context-assembly harness (contextkit)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "bundle_minilobby_trim.sh (v4) completed: bundle written to /tmp/bundle_minilobby_trim.md" (2026-07-09) after four documented failures.
- **what:** A read-only Go tool (in the contextkit tree, a SEPARATE Go module) that assembles a decision bundle for an LLM verdict: required -analysis/-root/-constitution/-task, repeatable -scope path[:Symbol]/-include/-doc, DB gathers via -psql (-schema-tables, -runtime-site/-page), -dry-run. Operational lessons made durable: resolve an action's file from the REGISTRY (key → Handler: symbol → function definition), never from header-comment conventions; scope a dedicated <key>_action.go file WHOLE but a shared file BY SYMBOL (attention dilution); run from inside contextkit with absolute -analysis/-constitution/-doc/-out and root-relative -scope; prefer the authored runbook's invocation over an example's shorthand. Used here to settle the sanctioned template-edit path before touching anything.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-bundle-v1→v4 (four entries); docs/bundle_minilobby_trim(4).sh (header); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4.0
- **relations:** cmd/diagnose harness; sanctioned edit paths (the question it settles)
- **verify-later:** docs/agent_docs/docs019.../go_files/contextkit/cmd/bundle; RUNBOOK_thin_slice invocation form

<!-- SOURCE: U23_docs_root_vonc.md -->
### cmd/diagnose read-only diagnosis harness
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_vonc_write_site_spec: runtime evidence "read read-only via the diagnosis harness's runtime gather"; full re-run command given with -seed-hypothesis/-dry-bundle.
- **what:** The diagnosis loop entry point: `go run ./cmd/diagnose` with -analysis (callgraph json), -constitution, -psql (read-only runtime gather against the cluster DB: agent_error_log, site_work_items), -seed-hypothesis/-seed-scope, -runtime-site/-page, producing per-iteration bundles (/tmp/diag_bundle_N.md, bundle-<id>/runtime.md); a -verdict-script drives the loop, the stub abstains without a model. The write_site_spec handoff shows the intended usage pattern: harness gathers evidence, a fresh session re-scopes and reads the real code.
- **sources:** docs/HANDOFF_vonc_write_site_spec_spec_data.md#how-to-get-the-evidence
- **relations:** cmd/bundle; fix-loop council (later consumers)
- **verify-later:** cmd/diagnose flags vs docs019 contextkit docs

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### backend_unreachable discovery check
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(f) "backend_unreachable REWRITTEN against the real DiscoveryCheck interface … Run(dctx DiscoveryCheckContext)(*CheckResult,error) … gofmt-clean"; enable pending.
- **what:** A discovery_checks check that NOOPs unless deploy_config.target='vm', GETs each backend site's public `/health`, and on failure emits a site_work_items row (source='discovery', item_type='backend_unreachable', item_key for dedup). Self-clearing. Alert-only: HandlerAgent "" because a down VM isn't chassis-fixable (the P5 vmhost adapter becomes the handler later). A `missing_beacon` check was floated too.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f, traffic_probe_plan(11).md#p4
- **relations:** ties to P5 vmhost adapter as future handler
- **verify-later:** discovery_checks/check_backend_unreachable.go; site_work_items idx_swi_dedup

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### contextkit target-resolution & bundle-assembly toolchain
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK_thin_slice(10).md: "The thin slice has been reasoned about but not yet run on a real task and checked." README(1).md's directory map shows a standalone `contextkit` Go module (7 commands) plus a `chassis-drafts/analyser-adapter` staging tree destined for the real repo — and that adapter is confirmed **built and deployed**: `internal/adapters/analyser/`, `cmd/analyser-adapter/`, and a live `deployments/kustomize/services/analyser-adapter/` overlay all exist in the working tree.
- **what:** A seven-command Go pipeline for assembling task-scoped LLM context bundles from a codebase: `analyser` (AST walk → JSON structural summary of package/imports/functions/types with line ranges), `resolve_targets` (deterministic lexical-overlap baseline that proposes scope candidates), `embed` (semantic vector index over symbols, Ollama-backed with a non-semantic offline stand-in for pipeline-proving), `fuse` (reciprocal-rank fusion of lexical + semantic candidate lists, k=60), `eval_targets` (recall@N / MRR scorer against a hand-authored `groundtruth_targets.json`), `assembler` (renders the final paste-ready bundle: constitution + task + in-scope code in full + neighbourhood signatures + schema + a "what was left out" pointer note), and `dbcontext` (shells out to psql for schema/rows/runtime-evidence with multipass row sizing — never an unbounded dump). Two contracts (`internal/analysis`, `internal/candidates`) are defined once and shared across commands.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/README(1).md, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/RUNBOOK_thin_slice(10).md, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/{analyser(2).go,assembler(2).go,dbcontext(1).go,go_files/embed.go,go_files/resolve_targets.go,go_files/fuse.go,go_files/eval_targets.go}
- **relations:** flat-file constitution (below); reuse-check retrieval pipeline design (below); adapter response-envelope contract (below, the chassis-integration half); fix-loop council (docs024 fixloop_eg_dartsonline)
- **verify-later:** internal/adapters/analyser/, cmd/analyser-adapter/, deployments/kustomize/services/analyser-adapter/ — confirm whether the standalone contextkit CLI tools themselves (analyser/assembler/embed/fuse/eval_targets/dbcontext binaries) were ever run on a real task per the runbook's "first real run" checklist, or whether only the adapter integration shipped.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### deploy-mechanics taxonomy (six ways a change ships)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** GUIDE_deploy_from_context_packs.md is written as an operational runbook with concrete `make`/`kubectl`/`kcat` commands and per-project worked examples (gamesdesign adoption, thunder checkpoint race, idea.uk go-live) dated against real image tags (`v1.0.1057`).
- **what:** A named taxonomy of the distinct ways a change actually reaches production in this system, used to scope any task before touching it: (A) chassis platform image — Go code changes need a rebuilt/pushed/retagged image and a k8s rollout; (B) database — SQL/migrations via psql, snapshot-first, re-query to verify; (C) work-items — insert a `site_work_items` row for the dispatch loop to claim; (D) orchestration trigger — a kcat `orchestrate` message to `system.agent.generic.requests`; (E) generated static sites — downstream/automatic via git → GitHub Actions → Backblaze once `build_status='deployed'`; (F) the idea.uk binary — a separate non-k8s, file-based Go binary with its own build/scp/restart cycle. Cross-cutting cautions: bump image tags or a rollout won't pull the change; "complete" is not "succeeded" — verify positive evidence, not just terminal status.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/GUIDE_deploy_from_context_packs.md
- **relations:** contextkit toolchain (above); deployment-github
- **verify-later:** Makefile targets referenced (build-*, deploy-agents, update-kustomization-images)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### reuse-check retrieval pipeline design (catalog → lexical/structural → embeddings → rerank)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** Directly implemented by the contextkit toolchain's `resolve_targets`/`embed`/`fuse` commands (see contextkit concept above, status partial/unrun); the design principles themselves are framed as reusable lenses.
- **what:** A layered design for "has this already been solved" reuse-checking that treats it as a retrieval problem with a judgement tail, not a generation problem: a maintained capability catalog is the cheapest check (lookup, not search); "identical" (token/AST fingerprinting, algorithmic, high-precision) and "similar" (semantic, embeddings) are split into different mechanisms because lexical/structural matching misses genuine near-duplicates with different names; every narrowing layer is tuned for **recall over precision** since a false-negative reuse check manufactures confident duplication that's worse than no check at all; a cheap model narrows the candidate set, a strong model decides on the shortlist — never the reverse; near-duplicate detection runs post-generation against a concrete draft (a real artifact to fingerprint), while fuzzy "what's there to build on" retrieval runs pre-generation. Signature+docstring embeddings are framed as a general retrieval substrate (also serving target resolution and capability-catalog curation), not a narrow dedup optimisation, and — at the scale of a few thousand symbols — need no vector database, just in-memory cosine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Reuse-checking; contextkit toolchain go files (embed.go, fuse.go)
- **relations:** contextkit toolchain (above); change-layer integration contract (above, reuse_index_refresh trigger)
- **verify-later:** whether any capability catalog or reuse index has been built beyond the contextkit prototype

<!-- SOURCE: U25_leopardess_social.md -->
### Read-only code bundle to settle method before editing (bundle verdict practice)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** VERDICT (2026-07-09) is the artifact: five method questions answered from a contextkit bundle + scoped Go reads before any write; "nothing is edited until its verdict lands".
- **what:** When the supported path for a change is genuinely unclear ("that is a code question → a bundle"), assemble a read-only context bundle (cmd/bundle + contextkit: dbcontext \d + capped SELECTs, scoped Go sources by symbol, relevant docs) and produce a written VERDICT answering the deciding questions before touching anything. The mini-lobby verdict overturned the handoff's premise (remove_element cannot reach the component — its header oversold it) and discovered the rendered-artifact template model. Operational notes: contextkit is a separate Go module; -scope/-include relative to -root; pure assembler triggers nothing.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md; docs/social001_vonc_tiktok_social/minilobby_task/bundle_minilobby_trim(4).sh (header); docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#4.0
- **relations:** diagnosis-loop (docs019 contextkit); operator discipline; section-editor
- **verify-later:** cmd/bundle + contextkit module

<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols repo-label symmetry (shared owner/repo resolver)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PATCH_code_symbols_shared_repo_label written; NNN_fix_lookup_repo_label_workaround header: "TEST-NOW ENABLER … apply that and REVERT this when the rebuilt image ships".
- **what:** index_code_symbols composes the code_symbols.repo label as owner/repo ("gqls/agentchassis") but lookup_code_symbols did not compose, so the diagnose workflow's lookup queried the bare name → 0 hits → empty code_results → "assemble_bundle: no scope". Structural fix: one shared resolveCodeRepoLabel used by both writer and reader so they can never drift; temporary config-only workaround hard-codes the literal until the image ships. General lesson: writer and reader of a keyed store must share the key resolver.
- **sources:** PATCH_code_symbols_shared_repo_label.md; NNN_fix_lookup_repo_label_workaround.sql; NNN_create_code_indexer_agent(2).sql (label convention)
- **relations:** code_symbols index; loop-back plumbing fault class
- **verify-later:** code_symbols_actions.go resolveCodeRepoLabel; whether the workaround REVERT ran

<!-- SOURCE: U16_docs019_design_plans.md -->
### analyse_repo_local in-process analysis + the stale-index incident
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** NNN_swap_indexer_to_local_analysis header documents the incident and the applied swap; PATCH_lift_fetcher_and_register gives the registration; README_flows/handoffs treat the local path as current.
- **what:** The adapter round-trip left the diagnose-agent with no local checkout (ReadSymbolBody could not slice bodies) and once resolved ref "HEAD" to a year-old commit, silently indexing a July-2025 tree (69 files/436 symbols vs 572 today). analyse_repo_local fetches the repo tarball at an explicit ref to a local temp dir and analyses in-process (shared internal/reposource fetcher), with pin_to_index_commit defaulting true for the diagnose loop (bodies match the index) and set false for the indexer (the indexer defines the commit). Corollary rules: explicit git refs never HEAD; the spawned pod needs GITHUB_READ_TOKEN via the isRepoCloningAgent spawn gate.
- **sources:** NNN_swap_analyse_repo_to_local.sql; NNN_swap_indexer_to_local_analysis.sql; PATCH_lift_fetcher_and_register.md; TRIGGER_code_indexer_v2(1).sh
- **relations:** analyser adapter (request_repo_analysis stays for the code-indexer); index freshness / CI-triggered indexing
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; registry entry

<!-- SOURCE: U16_docs019_design_plans.md -->
### "Documentation is code" — the context-assembly tool and paid service
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PLAN_context_assembly_tool_and_service(2) status: "A thin slice of Phase 1 is built and exercised on real code … first real run: a 30 KB bundle vs the script's 1.7 MB"; Phases 2–4 and the service unbuilt.
- **what:** For any development task, assemble a task-scoped context bundle from documentation + codebase and feed generation against ground-truth verification, so results are "more likely to be correct" than pasting code into a chat. The thesis: in an AI-driven workflow documentation (standards, intent, trajectory) is an operational input — version it, drift-detect it, compose it deterministically. Two audiences: dogfood on the chassis repo first, then a paid multi-tenant service behind the gateway. Design principles: engine/config split (tenant-agnostic engine + per-stack adapters, the decision that makes the service possible); seams for the optimal machinery (cascade router, decision-point checkers, mediator) defined as interfaces from v1; dogfood first. Phases: 0 contracts+constitution, 1 bundle builder MVP, 2 verification loop, 3 service (sandboxed verification is gating), 4 cascade/checkers/mediator.
- **sources:** PLAN_context_assembly_tool_and_service(2).md; 001_onboarding_discussion.txt; MAPPING_tool_to_actions_and_agents(2).md
- **relations:** bundle shape contract; six governance contracts (Phase 0); onboarding as the hard service problem; thin-slice-first principle
- **verify-later:** contextkit module state vs the plan's thin-slice claims; gateway project status

<!-- SOURCE: U16_docs019_design_plans.md -->
### Bundle shape contract (the task-scoped context package)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) "Status: contract specification"; flagged by FOCUS_whole_plan_review §1.1 as "the next load-bearing contract".
- **what:** A bundle is assembled fresh per task (never stored/reused, so never stale) with fixed sections: metadata; task/target; the authored layer (constitution, why-chain, priority profile, direction-of-travel, matched standards); code context (in-scope code in full, neighbourhood signatures, reuse-search results, schema, definition data); database data in three kinds; pointers to everything not inlined; and provenance (exactly what went in, logged as the decision log's inputs_used). Exists in a canonical structured form and a rendered text view. Two integrity rules from the edge-case pass: assemble from a consistent snapshot (no torn reads), and log what the generator SAW (rendered form), not what was assembled.
- **sources:** PLAN_bundle_shape_contract(2).md; FOCUS_whole_plan_review.md#1.1; FOCUS_pre_build_edge_cases(1).md#1.4,#4.3
- **relations:** decision log inputs_used; altitude/step-type; multipass fetch; contextkit assembler (the harness prototype)
- **verify-later:** whether any structured bundle object exists in code vs the markdown-emitting harness

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three kinds of database data (definition / operational / content)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.1, verified against the live schema ("workflows are jsonb columns on agent_definitions … there is no workflows table and no tools table").
- **what:** In a data-defined system much of "the code" is DB rows, so the bundle distinguishes: definition data (the system's design as data — workflows in agent_definitions, tools as content_components rows, prompts as text columns; fetched routinely, covered by reuse-search); operational data (telemetry — work items, orchestration_states, error logs; multipass-capped); and content data (the output — sites/pages/tenant data; the gated set where privacy matters in the service).
- **sources:** PLAN_bundle_shape_contract(2).md#2.1
- **relations:** multipass fetch; reuse search over definitions; sensitivity gates
- **verify-later:** n/a (contract)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Multipass fetch: probe → gate → include/reduce/point
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext "row data with multipass sizing" built in the harness (tool plan status); the full gate flow is contract-only.
- **what:** Query results have unknown size until run, so the builder probes with LIMIT N+1 (not count(*) — counting a filtered query costs as much as running it), checks a size gate and a sensitivity gate, then includes rows in full, reduces (aggregate / representative sample / pointer), or gates behind confirmation. An oversized result becomes an aggregate or pointer, never an unbounded dump.
- **sources:** PLAN_bundle_shape_contract(2).md#3; FOCUS_pre_build_edge_cases(1).md#4.2; GUIDE_deploy_from_context_packs(1).md (dbcontext)
- **relations:** three kinds of DB data; bounded bundle egress
- **verify-later:** dbcontext.go sizing logic

<!-- SOURCE: U16_docs019_design_plans.md -->
### Runtime evidence keyed by orchestration_id (the run narrative)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext -runtime-site built and used throughout (bundles carried agent_error_log/site_work_items); the fuller run-trace composition is contract-level.
- **what:** For debugging, the most useful context is the narrative of a run, reconstructable from one key: orchestration_id (+ correlation_id) spans orchestration_states (spawn tree, topics, status), llm_call_log (time-ordered step sequence), agent_error_log (error trail), pod logs (grep by run id) and the Kafka messages. Three cheap reads give a coherent single-run story instead of a scatter of lines. Log-correlation only works where the id is actually in the log line — a convention whose coverage the conventions agent audits.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2; PLAN_onboarding_agent_specs(6).md#2.9,#1.9; README_02 ("everything durable, one correlation id")
- **relations:** run signatures; codebase-conditional capabilities; diagnose_load_runtime
- **verify-later:** whether orchestration_id reliably appears in pod log lines (named open item)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Run signatures: expected-vs-actual sequence diff
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2 — designed; capture/storage named open in §9.
- **what:** Capture a healthy run's step sequence and spawn-tree shape once (from known-good runs, confirmed), store as authored reference, and on a debug task diff the actual run against it to surface the divergence point — "matched the healthy path to step 7, then diverged here" instead of "read the logs". Verification applied to runtime.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** runtime evidence by orchestration_id; diagnostic playbooks
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnostic playbooks / failure fingerprints as authored knowledge
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2: "seeded from those guides rather than authored fresh"; home (standards atoms vs sibling table) open.
- **what:** Known failure fingerprints — a failure's signature + the commands that confirm it + the fix pattern — curated from the existing debugging guides and failure writeups, surfaced into debug bundles the way standards are surfaced into build bundles, and grown as run-signature diffs reveal new ones.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** run signatures; debugging guide 016 (the seed corpus)
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Codebase-conditional capabilities (degrade, don't break; partial config is normal)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.3/§2.4 and agent-specs §2.9 — design rules, engine unbuilt.
- **what:** The definition-data / run-trace / run-signature / log-correlation capabilities rest on structural facts (behaviour stored as data, a run-correlation key, named logged steps, a known log fetch) that hold on our codebase but may not elsewhere. Stack-discovery records which facts hold; each capability degrades to a weaker form or states "unavailable, because this codebase has no X" instead of breaking. Companion rule: distinguish not-yet-authored config (degrade gracefully, note what's pending) from malformed config (fail loud) — the no-fallbacks rule applies to malformed data only.
- **sources:** PLAN_bundle_shape_contract(2).md#2.3-2.4; PLAN_onboarding_agent_specs(6).md#2.9; FOCUS_pre_build_edge_cases(1).md#2.3; FOCUS_whole_plan_review.md#2.5
- **relations:** stack-discovery agent; convention coverage = capability reliability
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Altitude: step type decides what the bundle emphasises
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** The harness assembler takes `-step framing|debug` (used in real bundle invocations, e.g. "-step debug", "the framing bundle that produced this plan used -step framing").
- **what:** The same task at different stages needs different context: framing/decision steps get full intent (why-chain, priority profile, direction-of-travel) with light code; implementation gets full code with a thin intent tether; debug leads with errors + runtime evidence + the expected-vs-actual diff. "Right altitude at the right moment" made concrete in the bundle composer.
- **sources:** PLAN_bundle_shape_contract(2).md#4; PLAN_imagery_sprite_sheet.md (framing-bundle use); tasks/gameslink_missing_index_rerender/RUNBOOK…(2).md (-step debug)
- **relations:** bundle shape contract; salience-loss problem
- **verify-later:** assembler.go step handling

<!-- SOURCE: U16_docs019_design_plans.md -->
### Go analyser + call-graph neighbourhood (and the wiring-include gap)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: analyser/assembler "built and exercised on real code"; internal/analysis "already exists in the repo" per migration(19); bundles record "533 files analysed".
- **what:** A go/ast walk producing a structured index (signatures, types, per-function `calls` by name-based matching — full go/types resolution deliberately avoided as too heavy/fragile); the assembler slices in-scope symbols in full plus a signature-level caller/callee neighbourhood. Known structural blind spot: registration/init wiring (registry.go) is unreachable via calls, so `-include` exists for wiring files — the same gap as manually-named docs. ReadSymbolBody (span-slice over start_line/end_line) was extracted, tested byte-identical to cmd/assembler, and shared with diagnose_assemble_bundle.
- **sources:** 001_claude_reasoning; PLAN_context_assembly_tool_and_service(2).md#5 status; GUIDE_deploy_from_context_packs(1).md; PLAN.md changelog (ReadSymbolBody)
- **relations:** analyser adapter (wraps the same library); evidence-follows re-scoping; broad-script-vs-lean-assembler tradeoff
- **verify-later:** internal/analysis (chassis) vs contextkit copy drift (flagged in PLAN.md)

<!-- SOURCE: U16_docs019_design_plans.md -->
### contextkit module packaging and the graduation seam
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: "The tools are now a single Go module, contextkit/ … the two contracts defined once — internal/analysis and internal/candidates"; production-relevant parts graduated in-cluster per migration(19).
- **what:** All harness tools (analyser, embed, resolve_targets, fuse, eval_targets, assembler, dbcontext, bundle, dedup, diagnose) live in one module with two shared contracts; graduation moves the internal packages under the chassis module path and turns command mains into actions without changing the contracts. Production runs in-cluster; the harness remains the dev/measurement scaffold (eval_targets stays offline, the flywheel's measurement tool). The trial's output is throwaway; the rule it teaches is durable.
- **sources:** PLAN_context_assembly_tool_and_service(2).md status; PLAN_workflows_and_actions_migration(19).md (analyser-adapter sections); MAPPING_tool_to_actions_and_agents(2).md
- **relations:** analyser adapter; one-decision-core two realisations (same seam idea)
- **verify-later:** go_files/contextkit module contents (unit U17 territory)

<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols: the per-repo code index (pgvector sibling table)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-09: "code_symbols applied cleanly (table + 4 indexes)"; later populated (436 symbols, embeddings 436/436 per the repo-label workaround header).
- **what:** A sibling to knowledge_base reusing its proven shape (vector(768) nomic + trigram on content + idempotent dedup, same AIService embedder with caller-applied nomic prefixes) but keyed for code: unique(repo,path,symbol), SHA-versioned via commit_sha, identity upsert (ON CONFLICT DO UPDATE WHERE content_hash IS DISTINCT — a symbol persists across commits, re-embeds only on change), pruned hard on re-index. Deliberate departures from chassis conventions (no version/previous_version_id, no deleted_at) because it is a rebuildable cache. HNSW chosen over the KB's IVFFlat for incremental churn (pgvector 0.8.0 confirmed both). One symbol = one row; no prose chunker (rag_index's character windows fragment Go mid-function).
- **sources:** NNN_create_code_symbols_index.sql; PLAN_workflows_and_actions_migration(19).md A5/B4/B4b + code-indexer reuse mapping
- **relations:** lookup/index_code_symbols; embedding policy split; knowledge_base reuse-not-copy
- **verify-later:** \d code_symbols; row counts per repo; embedding_model column use

<!-- SOURCE: U16_docs019_design_plans.md -->
### Hybrid code retrieval: index/lookup_code_symbols (vector + trigram, RRF in SQL)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19): actions built against real rag_actions.go, registered, deployed with the analyser (2026-06-12); lookup live in the diagnose workflow.
- **what:** index_code_symbols flattens analyser Output to symbol rows, skips unchanged content hashes, embeds non-fatally (trigram still works without an embedding), upserts and prunes; lookup_code_symbols mirrors rag_lookup (embed query → cosine vector search → trigram fallback → top-k + code_context). The hybrid RRF fusion moved into SQL, so the in-Go fuse tool never graduated. Deliberately a sibling action, not a parameterised rag_lookup (KB columns are baked into vectorSearchKB); the three embedding helpers are shared package-level functions.
- **sources:** PLAN_workflows_and_actions_migration(19).md (code-indexer reuse mapping + consumer-side-built entries); NNN_create_code_symbols_index.sql (query set)
- **relations:** code_symbols table; rag_lookup/rag_index (the mechanism source); repo-label symmetry
- **verify-later:** code_symbols_actions.go; registry entries (storage/code categories)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Analyser adapter: in-cluster polyglot parsing service
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-12: "ANALYSER DEPLOYED TO PRODUCTION (uk_001) … the code-indexer agent row is applied (INSERT 0 1)."
- **what:** A Kafka message-based adapter (mirroring git/thunder, not HTTP) consuming `analyse` requests (owner, repo, ref, language), fetching source read-only via the GitHub tarball endpoint (one request, recovers the exact commit_sha, path-traversal-guarded), parsing through the analysis.Analyse library behind an `Analyser` seam so per-language parsers (JS next) drop in, and replying on the caller's responses topic. Security: its own least-privilege repo-scoped read-only token as a k8s Secret mounted only on this pod — two narrow credentials (analyser read, git-adapter write), never one broad token. Built polyglot-ready NOW because the JS tools already exist (tech debt, not future planning).
- **sources:** PLAN_workflows_and_actions_migration(19).md (analyser adapter sections + repo access & security); FOCUS_js_tools_documentation.md
- **relations:** analyse_repo_local (the in-process alternative that later took the diagnose+indexer paths); adapter envelope contract
- **verify-later:** internal/adapters/analyser; analyser-adapter deployment in uk_001; whether request_repo_analysis still has callers

<!-- SOURCE: U16_docs019_design_plans.md -->
### Text-vs-code embeddings: share the mechanism, separate the policy (B4b)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) B4b "decided 2026-06-09"; schema embodies it (sibling table + embedding_model column).
- **what:** Shared and not duplicated: the AIService embedder seam, provider implementations, pgvector, and the semantic+trigram hybrid pattern. Separate per domain so each upgrades independently: the model (prose vs code-specific), dimension, preprocessing (nomic search_ prefixes are caller-side), retrieval tuning (HNSW vs IVFFlat, lexical-heavier for code), row definition, and evaluation. Turns B4a into "which model for code", measurable independently. Caution recorded: separation pays only if the mechanism stays shared.
- **sources:** PLAN_workflows_and_actions_migration(19).md B4a/B4b/A5 resolutions
- **relations:** code_symbols; B4a ceiling; CPU-Ollama feasibility (bulk-index speed, code-domain recall)
- **verify-later:** embedding_model column values; whether a code-specific model was ever adopted

<!-- SOURCE: U16_docs019_design_plans.md -->
### Code-indexer agent, index-orchestrator wrapper, and CI-triggered indexing
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** code-indexer row applied (2026-06-12); index-orchestrator seeded (v2 after the v1 `name`-column failure, 2026-07-02); CI trigger still a queue item ("CI-triggered indexing (self-contained)" in HANDOFF_builder_thread).
- **what:** The code-indexer is a thin orchestrator (analyse → index_code_symbols → complete). Run 6dfa37cd proved orchestrate+agent_type is adopted IN-PLACE on the shared chassis pod — which never holds GITHUB_READ_TOKEN — so the index-orchestrator wraps it in the proven spawn pattern so the spawned pod receives the secret (isRepoCloningAgent gate). Manual trigger TRIGGER_code_indexer_v2.sh sends the explicit branch ref; the planned durable form is a GitHub Actions step firing the envelope with GITHUB_SHA on push, retiring the index-staleness class for the diagnosis corpus.
- **sources:** NNN_create_code_indexer_agent(2).sql; NNN_seed_index_orchestrator(1).sql; TRIGGER_code_indexer_v2(1).sh; HANDOFF_builder_thread.md#3
- **relations:** analyse_repo_local staleness incident; spawn-consumed columns lesson; reuse-index freshness (governance)
- **verify-later:** index-orchestrator row; CI workflow file existence

<!-- SOURCE: U16_docs019_design_plans.md -->
### Documentation indexing rides the prose rag path
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Migration(19) "Documentation indexing (related, lighter…)" — direction recorded, not reported built.
- **what:** Docs are prose, so they fit rag_index/knowledge_base as-is: index each guide under a collection (e.g. `standards`), retrieve with rag_lookup; flat files in git remain the editable source of truth, the DB copy a derived rebuildable index so the assembler pulls relevant sections instead of pasting 124KB guides. Precondition: docs must live in a versioned repo. Separate, smaller workstream from code_symbols.
- **sources:** PLAN_workflows_and_actions_migration(19).md (documentation indexing section); FOCUS_js_tools_documentation.md
- **relations:** JS tools documentation gap; standards/docs agent (the matched-guidelines slot)
- **verify-later:** knowledge_base collections for docs

<!-- SOURCE: U16_docs019_design_plans.md -->
### cmd/bundle robustness contract (validate early, fail loud, manifest input)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** 003_contextkit_bundle_issues (2026-06-24) — two real failed runs analysed; HANDOFF_fixloop(8) notes the skip-message/usage patches exist but "needs gofmt + build".
- **what:** Field-found tool defects and the rules they teach: validate the cheapest input (-analysis JSON) BEFORE the slow psql-shelling gather phases, with an actionable message naming file/size/regeneration; accept a manifest/config file instead of 20-line backslash shell commands (kills the unquoted-parentheses class — real filenames contain "(1)"); a missing -doc/-scope path must fail loudly naming the path, because a silently-omitted file means a downstream chat reasons from incomplete context without knowing; single quoted -psql argument, no TTY.
- **sources:** 003_contextkit_bundle_issues.md; HANDOFF_fixloop_thread(8).md#2
- **relations:** bundle-first handoff practice; fail-loud-vs-degrade rule
- **verify-later:** cmd/bundle/main.go precondition ordering; -config support

<!-- SOURCE: U16_docs019_design_plans.md -->
### Reuse search before generation (code AND definition rows)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Named in the tool plan §2 and bundle contract §2.1/§8; only the retrieval substrate (code_symbols, pgvector) exists — the pre-generation reuse step is unbuilt.
- **what:** Before generating, run "what already does something like this" over existing functions/structs and — critically in a data-defined system — over definition rows (workflows/agents/tools), so reuse-before-recreate is mechanical rather than a remembered habit and near-copies of existing workflows are caught like duplicate functions. Needs a searchable text projection for jsonb definitions (named open). The index is derived state that goes stale silently — re-index on change events and stamp freshness.
- **sources:** PLAN_context_assembly_tool_and_service(2).md#2; PLAN_bundle_shape_contract(2).md#2.1,#8; FOCUS_pre_build_edge_cases(1).md#2.4,#15
- **relations:** reuse_index_refresh trigger; code_symbols; dev-guide reuse discipline
- **verify-later:** any reuse-search action; definition-row indexing

<!-- SOURCE: U16_docs019_design_plans.md -->
### DB capabilities capture (\dx/\df into the bundle)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** Migration(19) B2a "decided, mechanism built" in the harness (dbcontext -capabilities, assembler -dbfacts); the durable indexing-workflow half is future.
- **what:** Generation that writes SQL should see installed extensions and helper functions (so it knows pgvector exists and reuses snapshot_agent instead of hand-rolling a backup — the migration-110 footgun). Captured as DB context (not the analyser's job), included for DB-touching tasks with a reuse nudge; the durable plan folds capture into the indexing workflow on a migration cadence so bundles always carry current DB facts without anyone remembering a flag.
- **sources:** PLAN_workflows_and_actions_migration(19).md B2a + 2026-06-09 changelog
- **relations:** multipass fetch; schema-before-SQL discipline
- **verify-later:** dbcontext -capabilities flag; assembler dbfacts section

<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols repo-label symmetry (shared owner/repo resolver)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PATCH_code_symbols_shared_repo_label written; NNN_fix_lookup_repo_label_workaround header: "TEST-NOW ENABLER … apply that and REVERT this when the rebuilt image ships".
- **what:** index_code_symbols composes the code_symbols.repo label as owner/repo ("gqls/agentchassis") but lookup_code_symbols did not compose, so the diagnose workflow's lookup queried the bare name → 0 hits → empty code_results → "assemble_bundle: no scope". Structural fix: one shared resolveCodeRepoLabel used by both writer and reader so they can never drift; temporary config-only workaround hard-codes the literal until the image ships. General lesson: writer and reader of a keyed store must share the key resolver.
- **sources:** PATCH_code_symbols_shared_repo_label.md; NNN_fix_lookup_repo_label_workaround.sql; NNN_create_code_indexer_agent(2).sql (label convention)
- **relations:** code_symbols index; loop-back plumbing fault class
- **verify-later:** code_symbols_actions.go resolveCodeRepoLabel; whether the workaround REVERT ran

<!-- SOURCE: U16_docs019_design_plans.md -->
### analyse_repo_local in-process analysis + the stale-index incident
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** NNN_swap_indexer_to_local_analysis header documents the incident and the applied swap; PATCH_lift_fetcher_and_register gives the registration; README_flows/handoffs treat the local path as current.
- **what:** The adapter round-trip left the diagnose-agent with no local checkout (ReadSymbolBody could not slice bodies) and once resolved ref "HEAD" to a year-old commit, silently indexing a July-2025 tree (69 files/436 symbols vs 572 today). analyse_repo_local fetches the repo tarball at an explicit ref to a local temp dir and analyses in-process (shared internal/reposource fetcher), with pin_to_index_commit defaulting true for the diagnose loop (bodies match the index) and set false for the indexer (the indexer defines the commit). Corollary rules: explicit git refs never HEAD; the spawned pod needs GITHUB_READ_TOKEN via the isRepoCloningAgent spawn gate.
- **sources:** NNN_swap_analyse_repo_to_local.sql; NNN_swap_indexer_to_local_analysis.sql; PATCH_lift_fetcher_and_register.md; TRIGGER_code_indexer_v2(1).sh
- **relations:** analyser adapter (request_repo_analysis stays for the code-indexer); index freshness / CI-triggered indexing
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; registry entry

<!-- SOURCE: U16_docs019_design_plans.md -->
### "Documentation is code" — the context-assembly tool and paid service
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** PLAN_context_assembly_tool_and_service(2) status: "A thin slice of Phase 1 is built and exercised on real code … first real run: a 30 KB bundle vs the script's 1.7 MB"; Phases 2–4 and the service unbuilt.
- **what:** For any development task, assemble a task-scoped context bundle from documentation + codebase and feed generation against ground-truth verification, so results are "more likely to be correct" than pasting code into a chat. The thesis: in an AI-driven workflow documentation (standards, intent, trajectory) is an operational input — version it, drift-detect it, compose it deterministically. Two audiences: dogfood on the chassis repo first, then a paid multi-tenant service behind the gateway. Design principles: engine/config split (tenant-agnostic engine + per-stack adapters, the decision that makes the service possible); seams for the optimal machinery (cascade router, decision-point checkers, mediator) defined as interfaces from v1; dogfood first. Phases: 0 contracts+constitution, 1 bundle builder MVP, 2 verification loop, 3 service (sandboxed verification is gating), 4 cascade/checkers/mediator.
- **sources:** PLAN_context_assembly_tool_and_service(2).md; 001_onboarding_discussion.txt; MAPPING_tool_to_actions_and_agents(2).md
- **relations:** bundle shape contract; six governance contracts (Phase 0); onboarding as the hard service problem; thin-slice-first principle
- **verify-later:** contextkit module state vs the plan's thin-slice claims; gateway project status

<!-- SOURCE: U16_docs019_design_plans.md -->
### Bundle shape contract (the task-scoped context package)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) "Status: contract specification"; flagged by FOCUS_whole_plan_review §1.1 as "the next load-bearing contract".
- **what:** A bundle is assembled fresh per task (never stored/reused, so never stale) with fixed sections: metadata; task/target; the authored layer (constitution, why-chain, priority profile, direction-of-travel, matched standards); code context (in-scope code in full, neighbourhood signatures, reuse-search results, schema, definition data); database data in three kinds; pointers to everything not inlined; and provenance (exactly what went in, logged as the decision log's inputs_used). Exists in a canonical structured form and a rendered text view. Two integrity rules from the edge-case pass: assemble from a consistent snapshot (no torn reads), and log what the generator SAW (rendered form), not what was assembled.
- **sources:** PLAN_bundle_shape_contract(2).md; FOCUS_whole_plan_review.md#1.1; FOCUS_pre_build_edge_cases(1).md#1.4,#4.3
- **relations:** decision log inputs_used; altitude/step-type; multipass fetch; contextkit assembler (the harness prototype)
- **verify-later:** whether any structured bundle object exists in code vs the markdown-emitting harness

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three kinds of database data (definition / operational / content)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.1, verified against the live schema ("workflows are jsonb columns on agent_definitions … there is no workflows table and no tools table").
- **what:** In a data-defined system much of "the code" is DB rows, so the bundle distinguishes: definition data (the system's design as data — workflows in agent_definitions, tools as content_components rows, prompts as text columns; fetched routinely, covered by reuse-search); operational data (telemetry — work items, orchestration_states, error logs; multipass-capped); and content data (the output — sites/pages/tenant data; the gated set where privacy matters in the service).
- **sources:** PLAN_bundle_shape_contract(2).md#2.1
- **relations:** multipass fetch; reuse search over definitions; sensitivity gates
- **verify-later:** n/a (contract)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Multipass fetch: probe → gate → include/reduce/point
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext "row data with multipass sizing" built in the harness (tool plan status); the full gate flow is contract-only.
- **what:** Query results have unknown size until run, so the builder probes with LIMIT N+1 (not count(*) — counting a filtered query costs as much as running it), checks a size gate and a sensitivity gate, then includes rows in full, reduces (aggregate / representative sample / pointer), or gates behind confirmation. An oversized result becomes an aggregate or pointer, never an unbounded dump.
- **sources:** PLAN_bundle_shape_contract(2).md#3; FOCUS_pre_build_edge_cases(1).md#4.2; GUIDE_deploy_from_context_packs(1).md (dbcontext)
- **relations:** three kinds of DB data; bounded bundle egress
- **verify-later:** dbcontext.go sizing logic

<!-- SOURCE: U16_docs019_design_plans.md -->
### Runtime evidence keyed by orchestration_id (the run narrative)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** dbcontext -runtime-site built and used throughout (bundles carried agent_error_log/site_work_items); the fuller run-trace composition is contract-level.
- **what:** For debugging, the most useful context is the narrative of a run, reconstructable from one key: orchestration_id (+ correlation_id) spans orchestration_states (spawn tree, topics, status), llm_call_log (time-ordered step sequence), agent_error_log (error trail), pod logs (grep by run id) and the Kafka messages. Three cheap reads give a coherent single-run story instead of a scatter of lines. Log-correlation only works where the id is actually in the log line — a convention whose coverage the conventions agent audits.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2; PLAN_onboarding_agent_specs(6).md#2.9,#1.9; README_02 ("everything durable, one correlation id")
- **relations:** run signatures; codebase-conditional capabilities; diagnose_load_runtime
- **verify-later:** whether orchestration_id reliably appears in pod log lines (named open item)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Run signatures: expected-vs-actual sequence diff
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2 — designed; capture/storage named open in §9.
- **what:** Capture a healthy run's step sequence and spawn-tree shape once (from known-good runs, confirmed), store as authored reference, and on a debug task diff the actual run against it to surface the divergence point — "matched the healthy path to step 7, then diverged here" instead of "read the logs". Verification applied to runtime.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** runtime evidence by orchestration_id; diagnostic playbooks
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnostic playbooks / failure fingerprints as authored knowledge
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.2: "seeded from those guides rather than authored fresh"; home (standards atoms vs sibling table) open.
- **what:** Known failure fingerprints — a failure's signature + the commands that confirm it + the fix pattern — curated from the existing debugging guides and failure writeups, surfaced into debug bundles the way standards are surfaced into build bundles, and grown as run-signature diffs reveal new ones.
- **sources:** PLAN_bundle_shape_contract(2).md#2.2,#9
- **relations:** run signatures; debugging guide 016 (the seed corpus)
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Codebase-conditional capabilities (degrade, don't break; partial config is normal)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** PLAN_bundle_shape_contract(2) §2.3/§2.4 and agent-specs §2.9 — design rules, engine unbuilt.
- **what:** The definition-data / run-trace / run-signature / log-correlation capabilities rest on structural facts (behaviour stored as data, a run-correlation key, named logged steps, a known log fetch) that hold on our codebase but may not elsewhere. Stack-discovery records which facts hold; each capability degrades to a weaker form or states "unavailable, because this codebase has no X" instead of breaking. Companion rule: distinguish not-yet-authored config (degrade gracefully, note what's pending) from malformed config (fail loud) — the no-fallbacks rule applies to malformed data only.
- **sources:** PLAN_bundle_shape_contract(2).md#2.3-2.4; PLAN_onboarding_agent_specs(6).md#2.9; FOCUS_pre_build_edge_cases(1).md#2.3; FOCUS_whole_plan_review.md#2.5
- **relations:** stack-discovery agent; convention coverage = capability reliability
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Altitude: step type decides what the bundle emphasises
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** The harness assembler takes `-step framing|debug` (used in real bundle invocations, e.g. "-step debug", "the framing bundle that produced this plan used -step framing").
- **what:** The same task at different stages needs different context: framing/decision steps get full intent (why-chain, priority profile, direction-of-travel) with light code; implementation gets full code with a thin intent tether; debug leads with errors + runtime evidence + the expected-vs-actual diff. "Right altitude at the right moment" made concrete in the bundle composer.
- **sources:** PLAN_bundle_shape_contract(2).md#4; PLAN_imagery_sprite_sheet.md (framing-bundle use); tasks/gameslink_missing_index_rerender/RUNBOOK…(2).md (-step debug)
- **relations:** bundle shape contract; salience-loss problem
- **verify-later:** assembler.go step handling

<!-- SOURCE: U16_docs019_design_plans.md -->
### Go analyser + call-graph neighbourhood (and the wiring-include gap)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: analyser/assembler "built and exercised on real code"; internal/analysis "already exists in the repo" per migration(19); bundles record "533 files analysed".
- **what:** A go/ast walk producing a structured index (signatures, types, per-function `calls` by name-based matching — full go/types resolution deliberately avoided as too heavy/fragile); the assembler slices in-scope symbols in full plus a signature-level caller/callee neighbourhood. Known structural blind spot: registration/init wiring (registry.go) is unreachable via calls, so `-include` exists for wiring files — the same gap as manually-named docs. ReadSymbolBody (span-slice over start_line/end_line) was extracted, tested byte-identical to cmd/assembler, and shared with diagnose_assemble_bundle.
- **sources:** 001_claude_reasoning; PLAN_context_assembly_tool_and_service(2).md#5 status; GUIDE_deploy_from_context_packs(1).md; PLAN.md changelog (ReadSymbolBody)
- **relations:** analyser adapter (wraps the same library); evidence-follows re-scoping; broad-script-vs-lean-assembler tradeoff
- **verify-later:** internal/analysis (chassis) vs contextkit copy drift (flagged in PLAN.md)

<!-- SOURCE: U16_docs019_design_plans.md -->
### contextkit module packaging and the graduation seam
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Tool plan status: "The tools are now a single Go module, contextkit/ … the two contracts defined once — internal/analysis and internal/candidates"; production-relevant parts graduated in-cluster per migration(19).
- **what:** All harness tools (analyser, embed, resolve_targets, fuse, eval_targets, assembler, dbcontext, bundle, dedup, diagnose) live in one module with two shared contracts; graduation moves the internal packages under the chassis module path and turns command mains into actions without changing the contracts. Production runs in-cluster; the harness remains the dev/measurement scaffold (eval_targets stays offline, the flywheel's measurement tool). The trial's output is throwaway; the rule it teaches is durable.
- **sources:** PLAN_context_assembly_tool_and_service(2).md status; PLAN_workflows_and_actions_migration(19).md (analyser-adapter sections); MAPPING_tool_to_actions_and_agents(2).md
- **relations:** analyser adapter; one-decision-core two realisations (same seam idea)
- **verify-later:** go_files/contextkit module contents (unit U17 territory)

<!-- SOURCE: U16_docs019_design_plans.md -->
### code_symbols: the per-repo code index (pgvector sibling table)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-09: "code_symbols applied cleanly (table + 4 indexes)"; later populated (436 symbols, embeddings 436/436 per the repo-label workaround header).
- **what:** A sibling to knowledge_base reusing its proven shape (vector(768) nomic + trigram on content + idempotent dedup, same AIService embedder with caller-applied nomic prefixes) but keyed for code: unique(repo,path,symbol), SHA-versioned via commit_sha, identity upsert (ON CONFLICT DO UPDATE WHERE content_hash IS DISTINCT — a symbol persists across commits, re-embeds only on change), pruned hard on re-index. Deliberate departures from chassis conventions (no version/previous_version_id, no deleted_at) because it is a rebuildable cache. HNSW chosen over the KB's IVFFlat for incremental churn (pgvector 0.8.0 confirmed both). One symbol = one row; no prose chunker (rag_index's character windows fragment Go mid-function).
- **sources:** NNN_create_code_symbols_index.sql; PLAN_workflows_and_actions_migration(19).md A5/B4/B4b + code-indexer reuse mapping
- **relations:** lookup/index_code_symbols; embedding policy split; knowledge_base reuse-not-copy
- **verify-later:** \d code_symbols; row counts per repo; embedding_model column use

<!-- SOURCE: U16_docs019_design_plans.md -->
### Hybrid code retrieval: index/lookup_code_symbols (vector + trigram, RRF in SQL)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19): actions built against real rag_actions.go, registered, deployed with the analyser (2026-06-12); lookup live in the diagnose workflow.
- **what:** index_code_symbols flattens analyser Output to symbol rows, skips unchanged content hashes, embeds non-fatally (trigram still works without an embedding), upserts and prunes; lookup_code_symbols mirrors rag_lookup (embed query → cosine vector search → trigram fallback → top-k + code_context). The hybrid RRF fusion moved into SQL, so the in-Go fuse tool never graduated. Deliberately a sibling action, not a parameterised rag_lookup (KB columns are baked into vectorSearchKB); the three embedding helpers are shared package-level functions.
- **sources:** PLAN_workflows_and_actions_migration(19).md (code-indexer reuse mapping + consumer-side-built entries); NNN_create_code_symbols_index.sql (query set)
- **relations:** code_symbols table; rag_lookup/rag_index (the mechanism source); repo-label symmetry
- **verify-later:** code_symbols_actions.go; registry entries (storage/code categories)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Analyser adapter: in-cluster polyglot parsing service
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) changelog 2026-06-12: "ANALYSER DEPLOYED TO PRODUCTION (uk_001) … the code-indexer agent row is applied (INSERT 0 1)."
- **what:** A Kafka message-based adapter (mirroring git/thunder, not HTTP) consuming `analyse` requests (owner, repo, ref, language), fetching source read-only via the GitHub tarball endpoint (one request, recovers the exact commit_sha, path-traversal-guarded), parsing through the analysis.Analyse library behind an `Analyser` seam so per-language parsers (JS next) drop in, and replying on the caller's responses topic. Security: its own least-privilege repo-scoped read-only token as a k8s Secret mounted only on this pod — two narrow credentials (analyser read, git-adapter write), never one broad token. Built polyglot-ready NOW because the JS tools already exist (tech debt, not future planning).
- **sources:** PLAN_workflows_and_actions_migration(19).md (analyser adapter sections + repo access & security); FOCUS_js_tools_documentation.md
- **relations:** analyse_repo_local (the in-process alternative that later took the diagnose+indexer paths); adapter envelope contract
- **verify-later:** internal/adapters/analyser; analyser-adapter deployment in uk_001; whether request_repo_analysis still has callers

<!-- SOURCE: U16_docs019_design_plans.md -->
### Text-vs-code embeddings: share the mechanism, separate the policy (B4b)
- **category:** NEW:context-assembly
- **status-signal:** deployed
- **status-evidence:** Migration(19) B4b "decided 2026-06-09"; schema embodies it (sibling table + embedding_model column).
- **what:** Shared and not duplicated: the AIService embedder seam, provider implementations, pgvector, and the semantic+trigram hybrid pattern. Separate per domain so each upgrades independently: the model (prose vs code-specific), dimension, preprocessing (nomic search_ prefixes are caller-side), retrieval tuning (HNSW vs IVFFlat, lexical-heavier for code), row definition, and evaluation. Turns B4a into "which model for code", measurable independently. Caution recorded: separation pays only if the mechanism stays shared.
- **sources:** PLAN_workflows_and_actions_migration(19).md B4a/B4b/A5 resolutions
- **relations:** code_symbols; B4a ceiling; CPU-Ollama feasibility (bulk-index speed, code-domain recall)
- **verify-later:** embedding_model column values; whether a code-specific model was ever adopted

<!-- SOURCE: U16_docs019_design_plans.md -->
### Code-indexer agent, index-orchestrator wrapper, and CI-triggered indexing
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** code-indexer row applied (2026-06-12); index-orchestrator seeded (v2 after the v1 `name`-column failure, 2026-07-02); CI trigger still a queue item ("CI-triggered indexing (self-contained)" in HANDOFF_builder_thread).
- **what:** The code-indexer is a thin orchestrator (analyse → index_code_symbols → complete). Run 6dfa37cd proved orchestrate+agent_type is adopted IN-PLACE on the shared chassis pod — which never holds GITHUB_READ_TOKEN — so the index-orchestrator wraps it in the proven spawn pattern so the spawned pod receives the secret (isRepoCloningAgent gate). Manual trigger TRIGGER_code_indexer_v2.sh sends the explicit branch ref; the planned durable form is a GitHub Actions step firing the envelope with GITHUB_SHA on push, retiring the index-staleness class for the diagnosis corpus.
- **sources:** NNN_create_code_indexer_agent(2).sql; NNN_seed_index_orchestrator(1).sql; TRIGGER_code_indexer_v2(1).sh; HANDOFF_builder_thread.md#3
- **relations:** analyse_repo_local staleness incident; spawn-consumed columns lesson; reuse-index freshness (governance)
- **verify-later:** index-orchestrator row; CI workflow file existence

<!-- SOURCE: U16_docs019_design_plans.md -->
### Documentation indexing rides the prose rag path
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Migration(19) "Documentation indexing (related, lighter…)" — direction recorded, not reported built.
- **what:** Docs are prose, so they fit rag_index/knowledge_base as-is: index each guide under a collection (e.g. `standards`), retrieve with rag_lookup; flat files in git remain the editable source of truth, the DB copy a derived rebuildable index so the assembler pulls relevant sections instead of pasting 124KB guides. Precondition: docs must live in a versioned repo. Separate, smaller workstream from code_symbols.
- **sources:** PLAN_workflows_and_actions_migration(19).md (documentation indexing section); FOCUS_js_tools_documentation.md
- **relations:** JS tools documentation gap; standards/docs agent (the matched-guidelines slot)
- **verify-later:** knowledge_base collections for docs

<!-- SOURCE: U16_docs019_design_plans.md -->
### cmd/bundle robustness contract (validate early, fail loud, manifest input)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** 003_contextkit_bundle_issues (2026-06-24) — two real failed runs analysed; HANDOFF_fixloop(8) notes the skip-message/usage patches exist but "needs gofmt + build".
- **what:** Field-found tool defects and the rules they teach: validate the cheapest input (-analysis JSON) BEFORE the slow psql-shelling gather phases, with an actionable message naming file/size/regeneration; accept a manifest/config file instead of 20-line backslash shell commands (kills the unquoted-parentheses class — real filenames contain "(1)"); a missing -doc/-scope path must fail loudly naming the path, because a silently-omitted file means a downstream chat reasons from incomplete context without knowing; single quoted -psql argument, no TTY.
- **sources:** 003_contextkit_bundle_issues.md; HANDOFF_fixloop_thread(8).md#2
- **relations:** bundle-first handoff practice; fail-loud-vs-degrade rule
- **verify-later:** cmd/bundle/main.go precondition ordering; -config support

<!-- SOURCE: U16_docs019_design_plans.md -->
### Reuse search before generation (code AND definition rows)
- **category:** NEW:context-assembly
- **status-signal:** aspirational
- **status-evidence:** Named in the tool plan §2 and bundle contract §2.1/§8; only the retrieval substrate (code_symbols, pgvector) exists — the pre-generation reuse step is unbuilt.
- **what:** Before generating, run "what already does something like this" over existing functions/structs and — critically in a data-defined system — over definition rows (workflows/agents/tools), so reuse-before-recreate is mechanical rather than a remembered habit and near-copies of existing workflows are caught like duplicate functions. Needs a searchable text projection for jsonb definitions (named open). The index is derived state that goes stale silently — re-index on change events and stamp freshness.
- **sources:** PLAN_context_assembly_tool_and_service(2).md#2; PLAN_bundle_shape_contract(2).md#2.1,#8; FOCUS_pre_build_edge_cases(1).md#2.4,#15
- **relations:** reuse_index_refresh trigger; code_symbols; dev-guide reuse discipline
- **verify-later:** any reuse-search action; definition-row indexing

<!-- SOURCE: U16_docs019_design_plans.md -->
### DB capabilities capture (\dx/\df into the bundle)
- **category:** NEW:context-assembly
- **status-signal:** partial
- **status-evidence:** Migration(19) B2a "decided, mechanism built" in the harness (dbcontext -capabilities, assembler -dbfacts); the durable indexing-workflow half is future.
- **what:** Generation that writes SQL should see installed extensions and helper functions (so it knows pgvector exists and reuses snapshot_agent instead of hand-rolling a backup — the migration-110 footgun). Captured as DB context (not the analyser's job), included for DB-touching tasks with a reuse nudge; the durable plan folds capture into the indexing workflow on a migration cadence so bundles always carry current DB facts without anyone remembering a flag.
- **sources:** PLAN_workflows_and_actions_migration(19).md B2a + 2026-06-09 changelog
- **relations:** multipass fetch; schema-before-SQL discipline
- **verify-later:** dbcontext -capabilities flag; assembler dbfacts section

<!-- SOURCE: U15_docs019_running_notes.md -->
### contextkit CLI toolchain
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "analyser (-exclude + *(N).go skip...), resolve_targets (lexical), embed (semantic, Ollama nomic), fuse (RRF -json k=60), eval_targets..., assembler, dbcontext, dedup, thin_versions, cmd/bundle..., cmd/diagnose."
- **what:** A family of small, report-first, behaviour-tested Go CLIs built to prototype and measure context-assembly/diagnosis before chassis porting: `analyser` (Go-AST symbol index with exclude/dedup-skip), `resolve_targets` (lexical scoring), `embed`/`fuse` (semantic + RRF), `eval_targets` (recall@N/MRR against ground truth), `assembler` (composes a bundle: constitution + docs + schema + symbol bodies + runtime), `dbcontext` (read-only DB gather: `\d`, `-rows`, `-capabilities`, `-runtime-site`), `dedup`/`thin_versions` (docs-archiving tools), `cmd/bundle` (orchestration wrapper: runs dbcontext then assembler), `cmd/diagnose` (dev/test harness for the loop).
- **sources:** NOTES_running_synthesis_principles(59) multiple 2026-06-13 entries (tool-building); NOTES_running_synthesis_v2(36).md STATE DIGEST.
- **relations:** Diagnosis loop; docs archiving toolchain; code-context retrieval infrastructure.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Symbol-body slicer (analysis.ReadSymbolBody)
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v3(32) STATE DIGEST: "`analysis.ReadSymbolBody`... TESTED... and proven BYTE-IDENTICAL to `cmd/assembler`'s real bundle output."
- **what:** The single shared implementation (in `internal/analysis`, not duplicated per-consumer) that turns a `path:Symbol` scope entry into source text by slicing the analyser's already-recorded `start_line..end_line` spans — never re-parsing — used identically by `cmd/assembler` and the chassis `diagnose_assemble_bundle` action, closing a prior stub.
- **sources:** NOTES_running_synthesis_v3(32).md STATE DIGEST + DECISIONS.
- **relations:** Diagnose-agent self-contained repo fetch; contextkit CLI toolchain.

<!-- SOURCE: U15_docs019_running_notes.md -->
### contextkit CLI toolchain
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "analyser (-exclude + *(N).go skip...), resolve_targets (lexical), embed (semantic, Ollama nomic), fuse (RRF -json k=60), eval_targets..., assembler, dbcontext, dedup, thin_versions, cmd/bundle..., cmd/diagnose."
- **what:** A family of small, report-first, behaviour-tested Go CLIs built to prototype and measure context-assembly/diagnosis before chassis porting: `analyser` (Go-AST symbol index with exclude/dedup-skip), `resolve_targets` (lexical scoring), `embed`/`fuse` (semantic + RRF), `eval_targets` (recall@N/MRR against ground truth), `assembler` (composes a bundle: constitution + docs + schema + symbol bodies + runtime), `dbcontext` (read-only DB gather: `\d`, `-rows`, `-capabilities`, `-runtime-site`), `dedup`/`thin_versions` (docs-archiving tools), `cmd/bundle` (orchestration wrapper: runs dbcontext then assembler), `cmd/diagnose` (dev/test harness for the loop).
- **sources:** NOTES_running_synthesis_principles(59) multiple 2026-06-13 entries (tool-building); NOTES_running_synthesis_v2(36).md STATE DIGEST.
- **relations:** Diagnosis loop; docs archiving toolchain; code-context retrieval infrastructure.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Symbol-body slicer (analysis.ReadSymbolBody)
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v3(32) STATE DIGEST: "`analysis.ReadSymbolBody`... TESTED... and proven BYTE-IDENTICAL to `cmd/assembler`'s real bundle output."
- **what:** The single shared implementation (in `internal/analysis`, not duplicated per-consumer) that turns a `path:Symbol` scope entry into source text by slicing the analyser's already-recorded `start_line..end_line` spans — never re-parsing — used identically by `cmd/assembler` and the chassis `diagnose_assemble_bundle` action, closing a prior stub.
- **sources:** NOTES_running_synthesis_v3(32).md STATE DIGEST + DECISIONS.
- **relations:** Diagnose-agent self-contained repo fetch; contextkit CLI toolchain.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Context-pack code-gathering tooling: analyser/assembler vs package_*.sh
- **category:** NEW:context-pack-tooling
- **status-signal:** superseded
- **status-evidence:** The archived `GUIDE_deploy_from_context_packs(1).md` (adoption/docubundle) devotes half its length to a considered trade-off between the directory-walking `package_page_build_debug.sh` script ("broad, thorough, large... 110 files when the task really turns on about 14... caught `registry.go`") and the call-graph-based `analyser.go`/`assembler.go` pair ("leaner... but currently blind to wiring files" like `registry.go`, because registration happens via an init/registry mechanism the call graph never reaches). The live rewritten guide (`docubundle/GUIDE_deploy_from_context_packs.md`) drops this entire discussion — it restructures around a deploy-mechanism reference (A–F) and per-project quick reference, never mentioning the analyser/assembler tool or its registry.go blind spot at all.
- **what:** Two competing tools for assembling an LLM chat's working context from a Go repo: (1) `package_*.sh`, which concatenates whole hand-picked directories plus a live DB/pod capture into one text bundle; (2) `analyser.go` (structural JSON index of the repo) + `assembler.go` (pulls only named functions plus their call-graph neighbourhood into a tight bundle, given a `-scope`/`-task`/`-constitution` spec). The archived guide is careful to flag that the assembler's call-graph approach misses non-call wiring (e.g. `init()`-based registry.go registration) that the script's brute-force directory walk happens to catch.
- **sources:** adoption/docubundle/GUIDE_deploy_from_context_packs(1).md; live docubundle/GUIDE_deploy_from_context_packs.md; adoption/docubundle/analyser(2).go; adoption/docubundle/assembler(2).go
- **relations:** the analyser/assembler pair itself lives on with a proper home under `docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/` (`thin_slice_constitution.md`) — so the *tool* isn't abandoned, but this guide's discussion of when/why to reach for it over the script was dropped from the doc lineage.
- **verify-later:** whether analyser.go/assembler.go are still invoked anywhere in practice, or whether package_*.sh fully displaced them for chassis debugging work.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Context-pack code-gathering tooling: analyser/assembler vs package_*.sh
- **category:** NEW:context-pack-tooling
- **status-signal:** superseded
- **status-evidence:** The archived `GUIDE_deploy_from_context_packs(1).md` (adoption/docubundle) devotes half its length to a considered trade-off between the directory-walking `package_page_build_debug.sh` script ("broad, thorough, large... 110 files when the task really turns on about 14... caught `registry.go`") and the call-graph-based `analyser.go`/`assembler.go` pair ("leaner... but currently blind to wiring files" like `registry.go`, because registration happens via an init/registry mechanism the call graph never reaches). The live rewritten guide (`docubundle/GUIDE_deploy_from_context_packs.md`) drops this entire discussion — it restructures around a deploy-mechanism reference (A–F) and per-project quick reference, never mentioning the analyser/assembler tool or its registry.go blind spot at all.
- **what:** Two competing tools for assembling an LLM chat's working context from a Go repo: (1) `package_*.sh`, which concatenates whole hand-picked directories plus a live DB/pod capture into one text bundle; (2) `analyser.go` (structural JSON index of the repo) + `assembler.go` (pulls only named functions plus their call-graph neighbourhood into a tight bundle, given a `-scope`/`-task`/`-constitution` spec). The archived guide is careful to flag that the assembler's call-graph approach misses non-call wiring (e.g. `init()`-based registry.go registration) that the script's brute-force directory walk happens to catch.
- **sources:** adoption/docubundle/GUIDE_deploy_from_context_packs(1).md; live docubundle/GUIDE_deploy_from_context_packs.md; adoption/docubundle/analyser(2).go; adoption/docubundle/assembler(2).go
- **relations:** the analyser/assembler pair itself lives on with a proper home under `docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/` (`thin_slice_constitution.md`) — so the *tool* isn't abandoned, but this guide's discussion of when/why to reach for it over the script was dropped from the doc lineage.
- **verify-later:** whether analyser.go/assembler.go are still invoked anywhere in practice, or whether package_*.sh fully displaced them for chassis debugging work.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Context substrate principles (context engineering framing)
- **category:** NEW:context-engineering-principles
- **status-signal:** aspirational
- **status-evidence:** "Documentation is code... Authored vs derived... References, not copies... Salience over presence" (principles(59) §Context substrate).
- **what:** A cluster of framing principles for how an autonomous system should hold and pass context: distinguish authored (owner + lifecycle, can be wrong) from derived (auto-generated, true-by-being-actual, can't be wrong, only stale) sources; authored layers should point at derived artifacts rather than paraphrase them, so they don't drift; models lose the bigger picture from local-detail salience during reasoning, not from context-window overflow, so the lever is salience management, not window size; a bigger context window is explicitly named as "not the fix for too much context" (context rot).
- **sources:** NOTES_running_synthesis_principles(59) §Context substrate; §Building discipline "A bigger context window is not the fix."
- **relations:** Diagnosis loop (embodies several of these — small bundles, references not pasted copies); B4a embedding-quality finding.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Context substrate principles (context engineering framing)
- **category:** NEW:context-engineering-principles
- **status-signal:** aspirational
- **status-evidence:** "Documentation is code... Authored vs derived... References, not copies... Salience over presence" (principles(59) §Context substrate).
- **what:** A cluster of framing principles for how an autonomous system should hold and pass context: distinguish authored (owner + lifecycle, can be wrong) from derived (auto-generated, true-by-being-actual, can't be wrong, only stale) sources; authored layers should point at derived artifacts rather than paraphrase them, so they don't drift; models lose the bigger picture from local-detail salience during reasoning, not from context-window overflow, so the lever is salience management, not window size; a bigger context window is explicitly named as "not the fix for too much context" (context rot).
- **sources:** NOTES_running_synthesis_principles(59) §Context substrate; §Building discipline "A bigger context window is not the fix."
- **relations:** Diagnosis loop (embodies several of these — small bundles, references not pasted copies); B4a embedding-quality finding.
