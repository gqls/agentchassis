# EXTRACTION U14 — docs019 RUNBOOK* files (diagnosis/fix-loop project runbooks)
Extracted 2026-07-13. Files in scope: 108. Concepts found: 84.

Path alias used throughout: `docs019/` = `docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/`.
All files below live directly in that directory (maxdepth 1).

Eight version families. In every family the documents are cumulative working
journals: the highest-numbered version supersets earlier ones, so earlier
versions were scanned only for concepts that were dropped or replaced
(family-delta). The genuinely dropped/replaced ideas found: the `diagnose_run`
internal-iteration monolith, the `diagnostician` draft agent + seed-agents
migration path, the §7F seed-query reorder, and the §6 "gated items carried —
see PLAN.md" framing of the base RUNBOOK.md.

## Coverage

| file | treatment |
|---|---|
| docs019/RUNBOOK.md | family-delta |
| docs019/RUNBOOK(1).md | family-delta |
| docs019/RUNBOOK(2).md | family-delta |
| docs019/RUNBOOK(3).md | family-delta |
| docs019/RUNBOOK(4).md | family-delta |
| docs019/RUNBOOK(5).md | family-delta |
| docs019/RUNBOOK(6).md | family-delta |
| docs019/RUNBOOK(7).md | family-delta |
| docs019/RUNBOOK(8).md | family-delta |
| docs019/RUNBOOK(9).md | family-delta |
| docs019/RUNBOOK(10).md | family-delta |
| docs019/RUNBOOK(11).md | family-delta |
| docs019/RUNBOOK(12).md | family-delta |
| docs019/RUNBOOK(13).md | family-delta |
| docs019/RUNBOOK(14).md | family-delta |
| docs019/RUNBOOK(15).md | family-delta |
| docs019/RUNBOOK(16).md | family-delta |
| docs019/RUNBOOK(17).md | family-delta |
| docs019/RUNBOOK(18).md | family-delta |
| docs019/RUNBOOK(19).md | family-delta |
| docs019/RUNBOOK(20).md | family-delta |
| docs019/RUNBOOK(21).md | family-delta |
| docs019/RUNBOOK(22).md | family-delta |
| docs019/RUNBOOK(23).md | family-delta |
| docs019/RUNBOOK(25).md | family-delta |
| docs019/RUNBOOK(26).md | family-delta |
| docs019/RUNBOOK(27).md | family-delta |
| docs019/RUNBOOK(28).md | family-delta |
| docs019/RUNBOOK(29).md | family-delta |
| docs019/RUNBOOK(30).md | family-delta |
| docs019/RUNBOOK(31)_diagnosis_loop.md | family-latest |
| docs019/RUNBOOK_builder_route.md | family-delta |
| docs019/RUNBOOK_builder_route(1).md | family-delta |
| docs019/RUNBOOK_builder_route(2).md | family-delta |
| docs019/RUNBOOK_builder_route(3).md | family-delta |
| docs019/RUNBOOK_builder_route(4).md | family-delta |
| docs019/RUNBOOK_builder_route(5).md | family-delta |
| docs019/RUNBOOK_builder_route(6).md | family-delta |
| docs019/RUNBOOK_builder_route(7).md | family-delta |
| docs019/RUNBOOK_builder_route(8).md | family-delta |
| docs019/RUNBOOK_builder_route(9).md | family-delta |
| docs019/RUNBOOK_builder_route(10).md | family-delta |
| docs019/RUNBOOK_builder_route(11).md | family-delta |
| docs019/RUNBOOK_builder_route(12).md | family-delta |
| docs019/RUNBOOK_builder_route(13).md | family-delta |
| docs019/RUNBOOK_builder_route(14).md | family-delta |
| docs019/RUNBOOK_builder_route(15).md | family-delta |
| docs019/RUNBOOK_builder_route(16).md | family-delta |
| docs019/RUNBOOK_builder_route(17).md | family-delta |
| docs019/RUNBOOK_builder_route(18).md | family-delta |
| docs019/RUNBOOK_builder_route(19).md | family-delta |
| docs019/RUNBOOK_builder_route(20).md | family-delta |
| docs019/RUNBOOK_builder_route(21).md | family-latest |
| docs019/RUNBOOK_code_retrieval_route.md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(1).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(2).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(3).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(4).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(5).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(6).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(7).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(8).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(9).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(10).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(11).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(12).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(13).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(14).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(15).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(16).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(17).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(18).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(19).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(20).md | family-delta |
| docs019/RUNBOOK_code_retrieval_route(21).md | family-latest |
| docs019/RUNBOOK_thin_slice(12).md | family-delta |
| docs019/RUNBOOK_thin_slice(13).md | family-delta |
| docs019/RUNBOOK_thin_slice(14).md | family-delta |
| docs019/RUNBOOK_thin_slice(18).md | family-delta |
| docs019/RUNBOOK_thin_slice(19).md | family-delta |
| docs019/RUNBOOK_thin_slice(20).md | family-delta |
| docs019/RUNBOOK_thin_slice(21).md | family-delta |
| docs019/RUNBOOK_thin_slice(22).md | family-delta |
| docs019/RUNBOOK_thin_slice(23).md | family-delta |
| docs019/RUNBOOK_thin_slice(24).md | family-delta |
| docs019/RUNBOOK_thin_slice(25).md | family-delta |
| docs019/RUNBOOK_thin_slice(26).md | family-delta |
| docs019/RUNBOOK_thin_slice(27).md | family-latest |
| docs019/RUNBOOK_diagnosis_fix_loop.md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(1).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(2).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(3).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(4).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(5).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(6).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(7).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(8).md | family-delta |
| docs019/RUNBOOK_diagnosis_fix_loop(9).md | family-latest |
| docs019/RUNBOOK_design_diagnosis_loop.md | family-delta |
| docs019/RUNBOOK_design_diagnosis_loop(1).md | family-delta |
| docs019/RUNBOOK_design_diagnosis_loop(2).md | family-delta |
| docs019/RUNBOOK_design_diagnosis_loop(3).md | family-delta |
| docs019/RUNBOOK_design_diagnosis_loop(4).md | family-delta |
| docs019/RUNBOOK_design_diagnosis_loop(5).md | family-delta |
| docs019/RUNBOOK_design_diagnosis_loop(6).md | family-delta |
| docs019/RUNBOOK_design_diagnosis_loop(7).md | family-latest |
| docs019/RUNBOOK_site_quality(1).md | full |
| docs019/RUNBOOK_gamesdesign_index_rebuild.md | full |

## Concepts

### contextkit — task-scoped codebase bundle toolkit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Done: call-graph neighbourhood; live schema + row data via dbcontext …; the cmd/bundle orchestration wrapper"; RUNBOOK(31) header "This project builds contextkit … developed against, and dogfooded on, the agent-chassis repository" (2026-06-24).
- **what:** A small Go module (`contextkit/`: cmd/analyser, assembler, embed, dbcontext, resolve_targets, fuse, eval_targets, bundle, diagnose) that assembles a tightly-scoped slice of a codebase — the in-scope source in full, its call-graph neighbourhood as signatures, DB schema, runtime evidence, and authored guidance/constitution — into one paste-ready "bundle" per task. Two shared contracts (`internal/analysis`, `internal/candidates`) defined once, no per-tool copies. The deployed chassis diagnosis agent is its descendant; the CLI remains the dev/eval harness.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#the-pipeline; docs019/RUNBOOK(31)_diagnosis_loop.md#what-this-is; docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** read-only diagnosis loop; bundle altitudes; dbcontext; cmd/bundle wrapper
- **verify-later:** `docs019/go_files/contextkit/` module; `$CK/cmd/*`; `internal/analysis`, `internal/candidates`

### Read-only cite-or-abstain diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) checklist "§6G eval gate PASSED — run 51f95cda (2026-07-01): abstain → correct reads → REFUTE the naive framing → CONFIRM the grounded cause"; code_retrieval_route(21) "§7 ROUTE CLOSED — 2026-07-03 (run 73ed55c6)".
- **what:** An AI agent that investigates a bug strictly READ-ONLY: forms a hypothesis, gathers scoped evidence (code bodies + read-only DB rows + runtime records), issues a verdict that must CITE evidence or ABSTAIN (CONFIRMED/REFUTED/UNVERIFIABLE), then re-scopes by FOLLOWING the evidence (call graph for code, vetted queries for data) rather than re-searching the symptom. Never edits code, never runs builds, human-gated; the hard problem it targets is falsification — abandoning a wrong hypothesis.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#what-this-is; docs019/RUNBOOK_design_diagnosis_loop(7).md#overview; docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** convergence guards; verdict wire format; three-tier citation; falsification-first eval gate; diagnosis→fix loop (v2)
- **verify-later:** chassis `pkg/diagnose/` (loop.go, step.go, advance.go); `platform/orchestration/actions/diagnose_*_action.go`; agent_definitions rows diagnose-agent/diagnose-orchestrator

### Bundle step altitudes: framing vs implementation vs debug
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) assembler flags: "-step framing | implementation | debug … framing: in-scope shown as signatures (intent over detail)".
- **what:** A bundle declares its altitude: `framing` shows in-scope code as signatures only (used to expand an under-specified brief into a spec before targets can be picked), `implementation`/`debug` show full bodies, and `debug` adds a runtime-evidence section. Encodes the framing-vs-implementation altitude split as an explicit pipeline parameter.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#assembler-flags; docs019/RUNBOOK_thin_slice(27).md#fuzzy-tasks
- **relations:** contextkit toolkit; reasoning-state handoff
- **verify-later:** `$CK/cmd/assembler/main.go` step handling

### Call-graph neighbourhood selection with forced -include for wiring files
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "-include … Closes the blind spot the first adoption run found"; known-limits section "call-graph neighbourhood is name-matched, not type-resolved".
- **what:** The bundle's surrounding context is the call-graph neighbourhood (callees/callers/types) of the in-scope symbols, rendered as signatures, with `-neighbour package` as fallback when name-matching misses (interface dispatch). Registration/wiring files (e.g. registry.go, reached via init not calls) are force-included with `-include`. Ubiquitous names (Run, String, New) are dropped when the loop follows the graph, to avoid scope explosion.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#assembler-flags; docs019/RUNBOOK_design_diagnosis_loop(7).md#design-and-build-choices
- **relations:** named-scope guard vs capped expansion; ReadSymbolBody slicer
- **verify-later:** `internal/analysis/analyse.go` calls extraction; `pkg/diagnose/callgraph.go` ubiquitous-name drop list

### dbcontext — bounded read-only DB context gather
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) pipeline step 2 with worked flags; "-rows … multipass sizing (probe LIMIT N+1 …). Never an unbounded dump."
- **what:** CLI that pulls live DB context for a bundle: `-schema` (`\d` per table), `-rows` (SELECT with multipass sizing and a row cap), and `-runtime-site`/`-runtime-page` (recent agent_error_log rows + site_work_items lifecycle as a "Runtime evidence" block). All read-only; queries are appended as `-c` args, not shell-interpolated.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#dbcontext-flags
- **relations:** cmd/bundle wrapper; three-guard read-only SQL model; diagnose_load_runtime
- **verify-later:** `$CK/cmd/dbcontext/`

### cmd/bundle orchestration wrapper and the pure-composer boundary
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) design note "(Status: wrapper not yet built — flagged for decision)" superseded in the same file's Done list "the cmd/bundle orchestration wrapper (gather via dbcontext → assemble, composer stays read-only)".
- **what:** The assembler is a PURE COMPOSER — it never runs SQL or chooses tables; `cmd/bundle` is the orchestration wrapper that runs the requested read-only dbcontext gathers and then calls the assembler with the outputs wired in. Keeps query execution inside the bounded read-only tool while offering "one command including the SQL". Automatic table-selection was deliberately deferred and must propose-then-confirm.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#one-command; docs019/RUNBOOK_thin_slice(27).md#assembler-boundary
- **relations:** dbcontext; diagnosis loop gatherer (BundleGatherer shells out to cmd/bundle)
- **verify-later:** `$CK/cmd/bundle/`; `pkg/diagnose/gatherer.go`

### Bundle size doctrine — "a large bundle is a smell, not a goal"
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Context-window facts (verified against the Claude docs, June 2026)"; "aim to keep a working bundle under ~200K tokens (~800 KB)".
- **what:** Working rule for feeding bundles to models: keep under ~200K tokens; context rot means a full 1M window is not used evenly; the fix for an oversized bundle is narrower selection, not a bigger window. Includes the three feeding routes (chat paste, claude.ai Project, API with prompt caching of the stable prefix).
- **sources:** docs019/RUNBOOK_thin_slice(27).md#large-bundles
- **relations:** responses-are-summaries doctrine (Kafka side); call-graph neighbourhood (the narrowing instrument)
- **verify-later:** n/a (doctrine); bundle sizes in diagnosis_artifacts once built

### B4a finding — the symptom→infrastructure retrieval ceiling
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "OUTCOME (2026-06-17, 2 ground-truth tasks): skinner-box lexical 0.50, semantic 0.00; resultspec lexical 0.00, semantic 0.00, fused 0.00 … DECISION: embeddings do NOT earn a place in the code path on this evidence".
- **what:** Measured finding that when a bug's cause lives in shared infrastructure named for its FUNCTION rather than its FAILURE MODE, symptom-based code retrieval (lexical, semantic, or fused) has a hard ceiling — symptom words and mechanism words don't intersect, and no embedding closes a zero-overlap gap. Secondary finding: naive RRF fusion can be worse than lexical alone. This is the empirical justification for the diagnosis loop's re-scope-by-following-evidence design.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#B4a; docs019/RUNBOOK_design_diagnosis_loop(7).md#the-empirical-finding
- **relations:** lexical/semantic/fused target resolution; read-only diagnosis loop (the lever pulled instead)
- **verify-later:** `$CK/groundtruth_targets.json`; `docs019/go_files/contextkit/{lex,sem}.json`

### Lexical/semantic/fused target resolution (resolve_targets, embed, fuse, eval_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) pipeline 1b–1d with all four tools runnable; B4a decision "lexical (trigram + resolve_targets) carries the spine and embeddings are the tie-breaker".
- **what:** The target-resolution layer: a lexical (trigram) candidate proposer, an Ollama-backed semantic index (nomic-embed-text with search_document/search_query prefixes matching the chassis rag pipeline exactly), RRF rank fusion, and a recall@N/MRR scorer against a ground-truth task set. Built to answer "does semantic beat lexical for code" — the measured answer was no for this corpus.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#the-pipeline; docs019/RUNBOOK_thin_slice(27).md#B4a-task-1
- **relations:** B4a ceiling finding; code_symbols index (production analogue); evidence-fed scope resolver (later reuse of the same vector search)
- **verify-later:** `$CK/cmd/{resolve_targets,embed,fuse,eval_targets}/`; ollama-adapter service

### Ground-truth eval harness and its measurement-trap discipline
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "THE TRAP (hit 2026-06-14): resolve_targets was run with a DIFFERENT task … eval then scored … a meaningless 0/2"; the task-string bind guard and `-task-id` requirement.
- **what:** groundtruth_targets.json holds task→expected-symbol pairs; every eval binds the task string once, guards it against the truth file, uses ONE matched index for lexical and semantic, and forbids answer-vocabulary leaks in task wording (a leaked symbol name contaminated the ceiling test once). Three prior B4a attempts failed on METHOD, not result — the harness encodes the corrections.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#B4a-task-1; docs019/RUNBOOK_thin_slice(27).md#B4a-task-2
- **relations:** instrument-skepticism doctrine; B4a ceiling finding
- **verify-later:** `$CK/groundtruth_targets.json`; `$CK/cmd/eval_targets/`

### ReadSymbolBody — the single shared symbol-body slicer
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §3 "This collapse is DONE and verified: the merged assembler … diffed against the pre-collapse binary … byte-identical"; checklist "§1 ReadSymbolBody written + unit-tested".
- **what:** One implementation of symbol-body slicing (`analysis.ReadSymbolBody`) placed in BOTH module copies of `internal/analysis` (contextkit and chassis): body = file lines [StartLine, EndLine] inclusive, 1-indexed, exactly as the analyser records; resolves bare names and receiver-qualified `Type.Method`; whole-file for a path with no `:Symbol`. `cmd/assembler`'s duplicate slicing (splitScope/locateSymbol/readLines) was collapsed onto it — "two copies of one convention is the drift this project keeps getting bitten by".
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#1; docs019/RUNBOOK(31)_diagnosis_loop.md#3
- **relations:** diagnose_assemble_bundle; contextkit toolkit; module-copy drift (the two analyse.go copies noted drifted)
- **verify-later:** `internal/analysis/symbolbody.go` in both modules; `symbolbody_test.go`

### diagnose_assemble_bundle action
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) checklist "§2 diagnose_assemble_bundle merged (gofmt-clean)" and "§6C build + register the four diagnose actions … DONE 2026-06-29".
- **what:** The chassis action that, per iteration, reads the in-scope symbols' bodies via ReadSymbolBody from a decoded `repo_analysis` Output, composes hypothesis + code + runtime (+ live schema) into the `bundle` the verdict step reads. Scope fallback chain: `route.scope` (loop-back) → `input_data.seed_scope` → `code_lookup.code_results`. Unknown symbols are logged and skipped, not fatal.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#2; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** ReadSymbolBody; loop_scope_field lesson; diagnosis_artifacts egress (planned write-through here)
- **verify-later:** `platform/orchestration/actions/diagnose_assemble_bundle_action.go`

### Four convergence guards plus engine-level failsafes
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6E.1 "the loop stopped at stopped_by: evidence-not-growing — a guard, not luck. So the guards + the max_iterations cap are armed" (2026-06-29).
- **what:** Deterministic stop conditions independent of model behaviour: iteration-cap, scope-not-narrowing, evidence-not-growing, hypothesis-thrash — plus engine-level `timeout_seconds: 1800` and `fuel_budget: 1000` that bound a runaway even if the loop's bookkeeping is disarmed. Behaviour-tested (26-test suite), not eyeballed; the guards are the safety layer that lets a model verdict be untrusted.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position; docs019/RUNBOOK_design_diagnosis_loop(7).md#0
- **relations:** SeenRequests progress rule; named-scope guard; state threading self-check
- **verify-later:** `pkg/diagnose/loop.go` guards; `loop_test.go`, `step_test.go`

### SeenRequests — a new data_request counts as loop progress
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "#1 fix … guardAfter now tracks issued read-only data_requests in a SeenRequests set … evidence-not-growing (and hypothesis-thrash) yield when the verdict issues a NEW unseen request"; validated in run 51f95cda ("3 iters, new queries each, no premature stop").
- **what:** Fix for the loop stopping one iteration before its own good query ran: guards treat a NEW unseen read-only data_request as progress (its answer arrives next gather), while a re-issue of the same query still trips the guard. Required the `verdict_wire.go` sync (an older chassis copy silently mapped DataRequests to null, making the engine fix inert).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#1 fix and #1 status)
- **relations:** convergence guards; data_requests channel; verdict wire seam
- **verify-later:** `pkg/diagnose/advance.go` SeenRequests; `loop_datarequest_test.go`; `pkg/diagnose/verdict_wire.go`

### Named-scope guard vs capped call-graph expansion
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "BOTH FIXES DELIVERED 2026-07-03 … Guard now measures the MODEL-NAMED scope …; expansion runs only after the guard passes and is CAPPED (Config.MaxExpandedScope, engine default 18)"; route-close run 73ed55c6 "the expansion cap bounding iterations 2–3 at exactly 18 with all named entries kept".
- **what:** Blocker found when the real 515-file corpus replaced the stale 69-file one: guardAfter measured the POST-EXPANSION scope, and unbounded Neighbourhood expansion of six named symbols tripped scope-not-narrowing at iteration 1. Fix: the narrowing guard compares the MODEL-NAMED scope (deduped NextScope, no expansion); expansion is used only for the gather and capped at MaxExpandedScope (default 18, named entries always kept). A data_request escape on the scope guard was considered and REJECTED (would render it near-inert).
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E attempt-1 blocker 2); docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** convergence guards; call-graph neighbourhood; stale-corpus masking
- **verify-later:** `pkg/diagnose/{loop,step,advance}.go` NamedScopeSize/MaxExpandedScope; `loop_scopeguard_test.go`

### Deterministic scaffold / model-only-verdict split
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) "The scaffold is deterministic; the verdict is the only model-dependent part … This puts the SAFETY … in code that is verified, and isolates the part that needs a model."
- **what:** Architecture decision: loop control, guards, evidence trail, and re-scope are pure tested Go; the cite-or-abstain judgement is an interface (stub that always abstains, scripted verdicts, or the live model). The verdict runs as its OWN observable workflow step (`execute_llm_prompt`), not buried in a monolith. A model-less run can never fabricate a conclusion.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#design-and-build-choices; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** diagnose_run monolith (rejected alternative); verdict wire seam; convergence guards
- **verify-later:** `pkg/diagnose/` package purity (no DB imports); workflow verdict step config

### Verdict wire format seam (script IS the wire format)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) §4a "because the script format IS the wire format, every scripted scenario in §1 is a faithful dry-run of the model path"; RUNBOOK(31) §7.5 "keep the prompt's output schema and verdict_wire.go in lockstep".
- **what:** The model returns one JSON object (`outcome` ∈ CONFIRMED|REFUTED|UNVERIFIABLE, citations with `tier` ∈ static|state|runtime, revised_hypothesis, next_scope, data_requests) per PROMPT_diagnosis_verdict.md; `diagnose.ParseVerdict`/`verdict_wire.go` map it to the domain Verdict, with fail-safes: unknown outcome → UNVERIFIABLE, citation-less confirm/refute coerced to UNVERIFIABLE. Verdict scripts for testing use the identical format, so scripted runs are faithful dry-runs of the model path.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4a; docs019/RUNBOOK(31)_diagnosis_loop.md#7.5
- **relations:** cite-or-abstain loop; SeenRequests (wire sync incident); three-tier citation
- **verify-later:** `pkg/diagnose/verdict_wire.go` + `verdict_wire_test.go`; PROMPT_diagnosis_verdict.md

### Falsification-first evaluation gate
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6G "[x] PASSED 2026-07-01 (run 51f95cda)"; design_diagnosis_loop(7) §5 "A loop that confirms the first guess on every known bug is the failure mode, not the success — judge it on the reversals."
- **what:** The loop is not trusted on scaffold correctness; it must be run against known bugs and (a) reproduce mid-course REVERSALS (refute wrong hypotheses on evidence), (b) converge on causes the symptom could never retrieve, and (c) ABSTAIN naming the missing evidence when the bundle doesn't settle it. "Scaffold correct ≠ reasons well." The §6G pass showed UNVERIFIABLE→REFUTED→CONFIRMED over 3 iterations with cited evidence.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G; docs019/RUNBOOK_design_diagnosis_loop(7).md#5
- **relations:** gamesdesign resolveResultSpec fixture; three-tier citation; loop-worthiness test
- **verify-later:** evidence trails of runs 51f95cda, 5537ffdb, 73ed55c6 in orchestration_states

### gamesdesign resolveResultSpec fixture (the reference bug trajectory)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK(31) 2026-07-01 "STILL not resolveResultSpec — now for a substantive reason: reading real data, the model found a coherent cause … FORK for the user: (a) the fixture is stale … retire the 'reach resolveResultSpec' yardstick".
- **what:** The canonical eval scenario built from the real gamesdesign bug: seed "sections never reach save" → REFUTE on runtime evidence → REFUTE "token cap" → CONFIRM `resolveResultSpec` (singular output_field collapsed the page to a stub). Used as both the scripted-verdict reference and the live-eval yardstick; superseded as a yardstick once the site's current state no longer exhibited the symptom (the loop instead correctly diagnosed the missing `site_specs.cta` aspect), and the route was closed on the refute-and-confirm-a-grounded-cause bar instead.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#7.1; docs019/RUNBOOK(31)_diagnosis_loop.md#6G-passed; docs019/RUNBOOK_gamesdesign_index_rebuild.md
- **relations:** falsification eval gate; workflow result contract; B4a resultspec ceiling task
- **verify-later:** `/tmp` verdict scripts are ephemeral; groundtruth_targets.json resultspec entry

### Workflow-driven loop via next_step override (diagnose_route)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6C "[x] DONE"; "§6C coordinator next_step override CONFIRMED (coordinator.go:1093 getNextStepFromResult)"; §6E "[x] DONE 2026-06-29 (5× loop-back, CONFIRMED)".
- **what:** The loop is workflow-driven, not action-internal: `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict (execute_llm_prompt) → route (diagnose_route) → [loop back | emit] → complete`. `diagnose_route` runs the engine's Advance (guards + call-graph re-scope) once per iteration and overrides `next_step` in its result (the conditional_route pattern); it sets no output_field so its results are read as `route.*`. The workflow lives in agent_definitions `default_config` (not the legacy orchestration_workflow columns).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** diagnose_run (abandoned alternative); state threading; coordinator getNextStepFromResult
- **verify-later:** `platform/orchestration/actions/diagnose_route_action.go`; `coordinator.go` getNextStepFromResult; diagnose-agent default_config

### diagnose_run internal-iteration monolith
- **category:** diagnosis-loop
- **status-signal:** abandoned
- **status-evidence:** RUNBOOK(5) §6E "In this design there is NO workflow loop-back: the iteration lives inside the diagnose_run action (the engine Run())"; RUNBOOK(31) §6C "The BUILT design is the workflow-driven loop, NOT a diagnose_run action — there is no diagnose_run"; design_diagnosis_loop(7) "(The earlier diagnose_run monolith was removed.)"
- **what:** The earlier design where a single `diagnose_run` action executed the whole capped loop internally (orchestration shows one `run_loop` step; iteration visible only in logs/trail). Dropped in favour of the workflow-driven loop so each iteration's verdict and routing are separately observable orchestration steps. The seeded diagnose-agent briefly referenced the nonexistent action — the workflow-fix migration removed it. Family-delta: present in RUNBOOK(2)–(7), gone by RUNBOOK(8).
- **sources:** docs019/RUNBOOK(5).md#6E; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_design_diagnosis_loop(7).md#4b
- **relations:** workflow-driven loop (replacement); deterministic scaffold split
- **verify-later:** absence of diagnose_run in registry.go; diagnose-agent workflow JSON

### diagnostician draft and the seed→fix migration path
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK(31) §6C "Do NOT seed a new one (the diagnostician draft is superseded)"; "Do NOT apply the older NNN_move_diagnose_workflow_to_default_config.sql (bannered superseded)"; RUNBOOK(2) §E was "apply the seed migration (NNN_seed_diagnose_agents.sql)".
- **what:** The lineage of getting the diagnose pair into agent_definitions: an early `diagnostician` single-agent draft, then a seed-agents migration (RUNBOOK(2) era), superseded by fixing the ALREADY-seeded diagnose-agent/diagnose-orchestrator pair in place (workflow rewritten to diagnose_route shape in default_config, orchestrator workflow separately restored after the move migration nulled it). Every agent_definitions-touching migration snapshots the row first (`snapshot_agent`).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK(2).md#E
- **relations:** standing evidence rules (snapshot_agent); workflow-driven loop
- **verify-later:** migrations NNN_fix_diagnose_agent_workflow.sql, NNN_restore_diagnose_orchestrator_workflow.sql; agent_definitions snapshots

### diagnose-orchestrator spawn-wrapper pattern
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6F "Target the ORCHESTRATOR; diagnose-agent is the worker it spawns … keeping the loop off shared pods"; run evidence throughout §6E–§6G.
- **what:** The diagnosis entry point is a thin orchestrator (spawn_diagnoser → call_diagnoser → complete) that spawns a dedicated diagnose-agent pod and forwards its result, keeping heavy in-chassis loop work off the shared chassis pods. The same pattern was replicated for indexing (index-orchestrator, §7B.1) when in-place `orchestrate` proved token-less on shared pods.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1
- **relations:** repo-cloning token gate; generic orchestrate envelope; code-indexer
- **verify-later:** diagnose-orchestrator/index-orchestrator agent_definitions; spawn_actions.go

### Generic orchestrate envelope as the universal manual trigger
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D/§6F full kcat scripts: "The envelope's action: orchestrate + config.agent_type is the generic entry point"; reused verbatim for code-indexer, diagnose, and §7E.
- **what:** One trigger shape for hand-running any agent: kcat-produce to `system.agent.generic.requests` with correlation/orchestration/message/request ids, `action=orchestrate`, `config.agent_type=<entry agent>`, and task-specific `input_data`. Known wrinkles recorded: `site_id` intermittently arrives empty (reproducibility bug, parked), and runtime selectors (`site_id` vs `runtime_site` domain) drive different evidence filters in load_runtime.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7E
- **relations:** diagnose-orchestrator wrapper; needs_diagnosis intake (future replacement); correlation-id discipline
- **verify-later:** drafts/084_TRIGGER_diagnose_v1.sh; 080c/082 trigger scripts; site_id-empty envelope bug

### data_requests channel — model-authored read-only SQL
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) update 2026-07-01 (run 51f95cda) "the model's data_requests RAN (verdict_wire.go confirmed live)"; §6C "The data_requests channel — now wired (was dormant from a wiring gap, not by design)."
- **what:** The verdict may emit `data_requests` (single read-only SELECTs with `sql`/`why`); `diagnose_route` reads them from the verdict wire, keeps only read-only ones, forwards to `route.data_requests`; `diagnose_load_runtime` executes each on loop-back in a READ ONLY transaction with SET LOCAL statement_timeout and appends rows to runtime_evidence. Code re-scope and data re-gather are deliberately separate channels. This is the "DB-following" arm of evidence-following.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6B; docs019/RUNBOOK(31)_diagnosis_loop.md#6C (data_requests wiring)
- **relations:** three-guard model; EXPLAIN size guard; SeenRequests; live schema section
- **verify-later:** `diagnose_load_runtime_action.go` runDataRequests; `diagnose_route_action.go` readOnlyDataRequestsFromWire

### Three-guard read-only SQL enforcement model
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "SELECT-only is enforced at THREE layers (confirmed in the code)"; design_diagnosis_loop(7) §4d "CONFIRMED on this cluster (2026-06-17): pool_mode = transaction … a live BEGIN READ ONLY; DELETE … WHERE false probe refused the write".
- **what:** Defence in depth for model SQL: Guard 1 = the verdict prompt constrains to a single read-only SELECT; Guard 2 = `IsReadOnlySQL` lint applied twice (route boundary and pre-execution); Guard 3 = the actual guarantee, a `BeginTx(ReadOnly:true)` transaction (+ statement_timeout) that rejects any write including data-modifying CTEs. The `WHERE false` DELETE probe is the standard non-destructive verification. Guards 1–2 are hygiene, never the safety boundary.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#4; docs019/RUNBOOK(31)_diagnosis_loop.md#6B; docs019/RUNBOOK_design_diagnosis_loop(7).md#4d
- **relations:** sqlguard stripQuoted; diagnose_ro role; data_requests channel
- **verify-later:** `pkg/diagnose/sqlguard.go`; BeginTx call in diagnose_load_runtime

### sqlguard stripQuoted — lint false-positive on quoted literals
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK(31) (run 5537ffdb, 2026-07-01) "the page slug 'tool-drop-rate-simulator' contains 'drop' … FIXED (sqlguard.go stripQuoted blanks literal/identifier contents before the scan; regression test added)"; §6G banner "REMAINING … (a) DEPLOY the lint fix — latent" (2026-07-02).
- **what:** Keystone bug: the read-only lint scanned raw SQL, so a keyword substring inside a string literal (slug containing "drop") caused legitimate reads to be silently dropped — neutralising both the schema-section content read and the progress rule. Fix blanks literal/identifier contents before keyword scanning. Written + tested; the runbooks record deployment as still pending at the family's last update.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6G-passed; docs019/RUNBOOK(31)_diagnosis_loop.md#update-5537ffdb
- **relations:** three-guard model; data_requests channel
- **verify-later:** `pkg/diagnose/sqlguard.go` stripQuoted + test; whether the deployed image carries it

### diagnose_ro role and pooler-aware read-only enforcement
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** RUNBOOK(31) checklist "[x] §6B diagnose_ro role migration written … [ ] role migration applied"; RUNBOOK(31) "data_requests run via db.BeginTx(ReadOnly) on params.DB (clients_user), NOT a restricted role".
- **what:** A GRANT-only SELECT role (`diagnose_ro`) for the harness path, where `psql -c` statement stacking makes a transaction wrapper unsafe. Key doctrine: under pgbouncer enforce read-only by GRANT, never by `SET default_transaction_read_only` (session settings leak across pooled backends); transaction pooling makes BeginTx(ReadOnly) safe; statement_timeout goes in the DSN options. The chassis path deliberately runs as clients_user under the read-only transaction instead, so content tables stay SELECTable without grants.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4d; docs019/RUNBOOK(31)_diagnosis_loop.md#6B
- **relations:** three-guard model; dbcontext harness
- **verify-later:** NNN_create_diagnose_ro_role.sql applied?; pgbouncer-config pool_mode

### EXPLAIN pre-flight size guard on data requests
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "EXPLAIN size-guard added … runDataRequests now runs EXPLAIN (FORMAT JSON) inside the read-only tx BEFORE each query"; 51f95cda validation "the EXPLAIN guard (didn't block site-scoped queries)".
- **what:** Before executing each model query, the action plans it (EXPLAIN FORMAT JSON, no execution) and skips with feedback if estimated rows exceed budget (explain_max_rows 50000; cost cap opt-in); output rows are capped (row_cap 200) and cells truncated rune-safe (cell_chars 600); statement_timeout remains the execution backstop. A skip is feedback the model narrows from — a new narrower request counts as progress.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel; SeenRequests; responses-are-summaries doctrine
- **verify-later:** runDataRequests EXPLAIN branch in diagnose_load_runtime_action.go

### Live schema section in the bundle (gatherSchema)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "#2 IMPLEMENTED … the bundle now carries a ## Schema (live tables) section"; 51f95cda "using REAL table/column names — the schema section paying off, no more page_sections guessing".
- **what:** `diagnose_load_runtime` gains one read-only information_schema.columns query, DENYLIST-driven (%backup%/%bak%/%archive%/%supersede%, deliberately not %snapshot% since site_snapshots is live) plus a broad relevance include (site%/page%/content%/flow%) unless `schema_full=true`; rendered into the bundle via Go-defaulted config so no migration was needed. Stops the model guessing table names (it had invented `page_sections`).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#2)
- **relations:** data_requests channel; denylist-over-allowlist style (shared with index-hygiene excludes)
- **verify-later:** gatherSchema in diagnose_load_runtime_action.go; runtime.schema render path

### Loop state threading and the re-seed self-check
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "§6E.1 DONE — state threading verified … trail_len == iteration (3) and stopped on a guard"; the self-check "aborts loudly with stopped_by: state-threading-error rather than silently re-seeding into a runaway".
- **what:** Loop state (iteration, trail, seen_citations, hyp_history) threads across iterations via `state_field = route.diagnose_state`; a mis-pointed state_field silently disarmed the cap/trail/guards (each iteration re-seeded fresh). Fix = migration + a code self-check: if diagnose_route is about to seed but route.diagnose_state already exists, abort loudly — a regression tripwire for the exact bug class.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#current-position; docs019/RUNBOOK(31)_diagnosis_loop.md#6E
- **relations:** convergence guards; loop_scope_field lesson (same dotted-path config family)
- **verify-later:** NNN_fix_diagnose_route_state_threading.sql; self-check branch in diagnose_route_action.go

### loop_scope_field / EncodeScope shape-mismatch lesson
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6C "CONFIRMED ISSUE + FIX … EncodeScope is json.Marshal of the Scope struct … keys are the Go field names … ExtractStringListHelper … coerces that OBJECT to empty — so on every loop-back the scope … NEVER advanced"; "loop_scope_field migration confirmed live (Run 2 error read route.scope.Symbols)".
- **what:** A silent contract mismatch between an action's encoded output (untagged Go struct → `{"Symbols":[...]}`) and a downstream dotted-path reader expecting a plain list: first-pass worked, every re-scope was inert — invisible to engine tests because it lived in workflow config. Fix was config-only: point `loop_scope_field` at `route.scope.Symbols`. Emblematic of the dotted-lookup config contract class (also: analysis_field, result_from, repo_field).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6C
- **relations:** state threading; repo-label asymmetry; workflow result contract
- **verify-later:** NNN_fix_assemble_bundle_loop_scope_field.sql; ExtractNestedField 3-level traversal

### code_symbols index + code-indexer agent
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D "[x] index populated (436 rows)"; code_retrieval_route(21) §7C "4,155 symbols; 499 distinct paths … min=max=commit 36710be; prune cleared all 436 old rows"; measured "≈ 5 symbols/sec through the single ollama-adapter".
- **what:** The retrieval corpus: `code_symbols` (repo, path, symbol, kind, signature, doc, content, embedding, commit_sha) written solely by the `code-indexer` agent (request_repo_analysis → await analyser → index_code_symbols; later analyse_repo_local in-process) and read by `lookup_code_symbols` (vector + trigram). UPSERT-safe via uq_code_symbols_identity; prune removes rows whose commit_sha differs from the new index commit; embedded text is name + signature + first doc line + path. Triggered via index-orchestrator (spawning wrapper so the pod holds the read token).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_code_retrieval_route(21).md#7A–7C; docs019/RUNBOOK_thin_slice(27).md#in-cluster-path
- **relations:** analyser adapter; analyse_repo_local; repo-label convention; evidence-fed resolver
- **verify-later:** code_symbols table + constraints; code-indexer/index-orchestrator agent rows; index_code_symbols action

### Analyser adapter — repo analysis as a Kafka service
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D "Adapter availability — RESOLVED 2026-06-24 (secret name/key mismatch) … the adapter came up 1/1 Running"; thin_slice(27) in-cluster deployment section (kustomize dry-build, topic CRD, health checks).
- **what:** A deployed adapter (`internal/adapters/analyser`, topic system.adapter.analyser.requests) that clones a GitHub repo (read-only PAT) and returns the analysis Output over Kafka. Deployment lessons captured: inject the single needed secret via secretKeyRef (never envFrom, which exposes every platform secret); topic auto-create is off so the KafkaTopic CRD must exist; topic-addressed adapters legitimately show target_agent_type='unknown' in awaited_requests; idle consumer-poll timeouts log at ERROR cosmetically. Its per-iteration use by the loop was later removed (analyse_repo_local), but indexing and other consumers remain.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK_thin_slice(27).md#in-cluster-path
- **relations:** analyse_repo_local (supersedes the loop's cross-pod call); code-indexer; repo-cloning token gate
- **verify-later:** deployments/kustomize/services/analyser-adapter; personae-platform-secrets GITHUB_READ_TOKEN wiring

### Repo-label composition convention (owner/repo) and the lookup asymmetry bug
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) "Label convention — DECIDED 2026-06-11: code_symbols.repo is the owner/repo form … COMPOSED by index_code_symbols"; RUNBOOK(31) "ROOT CAUSE CONFIRMED (2026-06-26) — repo-label asymmetry … the lookup queried WHERE repo='agentchassis' against rows under 'gqls/agentchassis' → 0 hits"; "Structural patch APPLIED".
- **what:** `code_symbols.repo` is always the composed `owner/repo` label. The index composed it but the lookup didn't → iteration-1 seeding returned nothing ("no scope"). Fixed twice: a config-only workaround (literal repo on the lookup step) then the structural `resolveCodeRepoLabel` shared by index AND lookup so they cannot diverge. Also the standing diagnostic rule it produced: confirm by correlation_id, never by `LIMIT 1` (a COMPLETED LIMIT-1 row was a red herring twice).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F-run1; docs019/RUNBOOK_thin_slice(27).md#label-convention
- **relations:** loop_scope_field lesson (same config-contract class); standing evidence rules
- **verify-later:** resolveCodeRepoLabel in code_symbols_actions.go; lookup step config (no repo_field literal)

### analyse_repo_local — in-process tarball fetch + analysis
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6F "DECIDED for option 3, REFINED turn 27 … BUILT (turn 27, gofmt-clean)"; checklist "§6C … image also carries analyse_repo_local + lifted internal/reposource — DONE 2026-06-29"; §7B swap "migration applied; snapshot 971da9c9".
- **what:** Resolution of the "no repo checkout on the diagnose pod" blocker: the agent fetches the repo itself via the analyser's tarball fetcher (`GET /repos/{o}/{r}/tarball/{ref}`, no git in the chassis) lifted into a neutral `internal/reposource` package, runs `analysis.Analyse(dir)` in-process for spans + call graph, and reads bodies from that checkout. `pin_to_index_commit` pins the fetch to the dominant code_symbols commit so seeded path:Symbol entries resolve (the indexer sets it false — it DEFINES the commit). Options weighed and rejected: bodies-in-DB (whole-repo Kafka payloads) and a stateful analyser serving slices (per-iteration coupling).
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F (Run 3 + deploy sequence); docs019/RUNBOOK_code_retrieval_route(21).md#7B
- **relations:** analyser adapter; code_symbols; index hygiene excludes; repo-cloning token gate
- **verify-later:** internal/reposource/github_source.go; analyse_repo_local_action.go; NNN_swap_analyse_repo_to_local.sql

### Repo-cloning token gate (isRepoCloningAgent)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) "spawn_actions.go injects GITHUB_READ_TOKEN via secretKeyRef … gated by isRepoCloningAgent -> ONLY diagnose-agent pods get the token; the spawner never holds it"; §7B.1 "isRepoCloningAgent gained 'code-indexer' … Verified end to end by run 93ba14e6".
- **what:** Least-privilege credential injection at spawn time: only agent types allowlisted as repo-cloning receive the read-only GitHub token env (secretKeyRef into the spawned pod), and the shared chassis pods never hold it — which is why indexing/diagnosis run through spawning orchestrators rather than in-place.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7B.1
- **relations:** diagnose-orchestrator wrapper; analyser adapter secret lesson
- **verify-later:** spawn_actions.go isRepoCloningAgent list

### Stale-corpus class: HEAD pinning, explicit refs, CI-triggered indexing
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** code_retrieval_route(21) §7A "it was a faithful index of a YEAR-OLD tree … Decision: … envelopes ALWAYS carry an explicit branch/sha"; queue item 4 "CI-triggered indexing: GitHub Actions step firing the index-orchestrator envelope with ${GITHUB_SHA} on push … [queued]".
- **what:** A recurring staleness class: consumers pinned to `HEAD`/`latest` silently track an ancient artefact (remote HEAD = unmoved main from 2025; agent image_tag 'latest' = pre-architecture build). Adopted: explicit refs in every envelope, derive REF from the working checkout. Designed (aspirational): Structural A — a post-deploy CI step indexes at ${GITHUB_SHA} so index commit == deployed commit by construction; Structural B — fast-forward main to the deployed sha. Rejected: resolving "most recently pushed branch" via API (latest-pushed ≠ deployed).
- **sources:** docs019/RUNBOOK_code_retrieval_route(9).md#ref-strategy; docs019/RUNBOOK_code_retrieval_route(21).md#7A; docs019/RUNBOOK_builder_route(21).md#queue (item 4)
- **relations:** image_tag 'latest' trap (same class); code_symbols prune semantics
- **verify-later:** GitHub Actions workflow for post-deploy indexing (absent?); git ls-remote origin HEAD

### Index hygiene — exclude archived code copies, prune by commit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) §7C.1 "[x] census: docs-archived 430 symbols / 50 files … interim DELETE run"; "[x] reindex f284b749 VERIFIED (2026-07-03): commit e3176f8, 3,723 symbols, docs_rows=0".
- **what:** The repo stores archived copies of its own code under docs/ (and download-suffixed `name(N).go` files); indexing them pollutes retrieval with dead duplicates (observed: nine duplicate assembler copies as ranks 1–9). Fixes: the analyser skips `*(N).go` unconditionally; `analyse_repo_local` gained `exclude_patterns` (Go default ["docs/"]) calling AnalyseWithExclude; prune semantics (`commit_sha IS DISTINCT FROM $new`) clear old-commit rows on the next reindex. Same trap documented CLI-side ("analyse the RIGHT ROOT", relative -exclude substrings).
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7C.1; docs019/RUNBOOK_thin_slice(27).md#known-limits; docs019/RUNBOOK_thin_slice(27).md#B4a (build the index over REAL source)
- **relations:** analyse_repo_local; stale-corpus class; B4a eval discipline
- **verify-later:** exclude_patterns config on analyse_repo_local; code_symbols docs/% row count

### Evidence-fed fuzzy-scope resolver (§7D)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) route-close run 73ed55c6: "resolver canonicalisation (model's basenames → full paths @0.81–0.87) AND descriptive resolution … both load-bearing in the confirming scope"; "[x] code WRITTEN 2026-07-02 … resolver image is LIVE".
- **what:** Many verdict `next_scope` entries are English descriptions, not path:Symbol handles — previously inert (no call-graph match, no body sliced). The resolver, inside diagnose_route after verdict-parse and before Advance, embeds each non-exact entry (same nomic client/prefixes) and vector-searches code_symbols, replacing it with the top hits (resolver_top_k default 2 — tuned so substitution stays inside the narrowing guard's +2 allowance; min similarity 0.55; unresolvable entries survive as labels, "no worse"). Flagged deliberate change: the trail records the RESOLVED scope, the more auditable record. Reuses the seed lookup's retrieval machinery wholesale.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D; docs019/RUNBOOK_code_retrieval_route(21).md#route-closed
- **relations:** code_symbols; named-scope guard (the +2 interplay); §7F seed reorder (retired by this)
- **verify-later:** diagnose_route_action.go resolver step 3.5; diagnose_route_resolver_test.go

### Three-tier citation standard (static / data / runtime)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "CONFIRMED on citations spanning ALL THREE TIERS: Tier 2 work item, Tier 1 site_specs cta → (0 rows) (query+result cited together), Tier 0 plan_sections_action.go:planSection quoting 'case \"skip_field\"'" (run 73ed55c6, 2026-07-03).
- **what:** Verdict citations carry a tier (static code / live data reads / runtime records); the route's success bar — and the strongest diagnosis shape — is a CONFIRMED grounded across all three tiers, with query+result cited together for data reads and a quoted code branch for the mechanism. Distinguishes "confirmed by inference at the data layer" from "code-level mechanism named".
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#route-closed; docs019/RUNBOOK_design_diagnosis_loop(7).md#1 (tier vocabulary)
- **relations:** verdict wire seam; falsification eval gate; fix-loop council opinions (verdict-wire-style contract)
- **verify-later:** citation tier handling in verdict_wire.go; evidence_trail of run 73ed55c6

### §7F seed-query reorder (lookup after load_runtime)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** code_retrieval_route(21) "§7F RETIRED" (banner) after "SEED RELEVANCE MET — all twelve seed symbols build-domain … first time ever; §7F (seed reorder) substantially retired".
- **what:** A deferred design to reorder lookup_symbols after load_runtime so the seed query could be built from the symptom PLUS salient error-log lines. Made unnecessary once the corpus was current and the resolver landed — seed relevance was proven by content twice. Family-delta: the idea persists as a section in every version but flips from DEFERRED to RETIRED at (18)+.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7F; docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E scoring)
- **relations:** evidence-fed resolver (the retiring cause); code_symbols currency
- **verify-later:** n/a (not built)

### Corpus enrichment policy — measure first, mechanical before authored
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** code_retrieval_route(21) "Should every function carry a human description for embedding-match? NO … Order of investment, gated on the §7E measurement" (question raised 2026-07-02).
- **what:** Position on enriching the retrieval corpus: (1) mechanical, rot-free first — extend composeSymbolContent with a function's string literals (diagnosis queries quote log lines and the literals ARE the log lines); (2) Go-convention one-sentence docs only on the exported surface + action entrypoints; (3) explicitly NO separate tag system — the doc first line is the tag surface. Rationale: stale docs make retrieval confidently wrong, the worst failure mode for a cite-or-abstain loop.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#corpus-enrichment
- **relations:** code_symbols; F3 learning layer (doc enrichment feed-back)
- **verify-later:** composeSymbolContent; exported_no_doc census query

### Reasoning-state as a first-class handoff artefact
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** thin_slice(27) improvement 5 "A bundle carries CODE + SCHEMA + RUNTIME EVIDENCE, but NOT reasoning state … The stopgap is a hand-written 'diagnosis so far' preamble (PREAMBLE_gamesdesign_diagnosis_handoff.md)"; the loop's evidence_trail later persists per-iteration hypothesis/scope/verdict.
- **what:** The insight that a context bundle without the evidence trail forces a fresh reader to re-derive falsified hypotheses; the design goal is a structured reasoning-state section accumulating across iterations (hypotheses tried, verdict + citation each, open discriminator). Partially realised by the loop's evidence trail in collected_data; the bundle-intrinsic version and per-iteration notes (F0.3 via doc_notes) remain in flight.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#next-improvements (item 5); docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F0.3)
- **relations:** per-task running notes; diagnosis_artifacts egress; falsification eval gate
- **verify-later:** evidence_trail shape in collected_data; doc_notes diagnosis category rows

### Instrument-skepticism doctrine
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) "almost every wrong measurement came from the test instrument, not the system under test — a wrong task string, a contaminated index, a stale shell variable, a task description that leaked the answer's vocabulary".
- **what:** Standing caution carried into the loop's design: apply cite-or-abstain suspicion to one's OWN inputs (the bundle, the query, the ground truth) before suspecting the target system. Surfaced repeatedly in B4a and encoded in the eval harness guards; named as the thing to watch when evaluating the model verdict.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#a-standing-caution; docs019/RUNBOOK_thin_slice(27).md#B4a-task-1
- **relations:** ground-truth eval harness; standing evidence rules (0-rows not decisive)
- **verify-later:** n/a (doctrine)

### Cross-module engine port procedure
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) §4c "surfaced FOUR build errors in the first real make build-core-manager … None of these four were logic bugs … doing Steps 1–4 in order pays it once."
- **what:** The validated sequence for moving a package between Go modules (contextkit internal/diagnose → chassis pkg/diagnose): (1) copy the WHOLE package as a unit and diff file lists; (2) rewrite the moved-package import path everywhere; (3) build+test the package alone before the binary; (4) grep every shared-package call the new code makes against the REAL helper surface (datahelpers.ExtractStringSlice didn't exist; compose ExtractStringListHelper(ExtractNestedField(...))). The chassis copies keep the agentchassis import (step.go) while the prototype keeps contextkit's.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4c; docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#1 build note)
- **relations:** ReadSymbolBody dual placement; module-copy drift
- **verify-later:** pkg/diagnose file list vs contextkit/internal/diagnose

### Workflow result contract — resolveResultSpec and preferred-key unification
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** gamesdesign_index_rebuild "page-content-writer's singular output_field now flattens into the response … resolveResultSpec emits a resolved-contract line, a deprecation Warn per old key" (deployed 2026-06-18); code_retrieval_route(21) "STATUS 2026-07-04: DEPLOYED (image v1.0.1092) and the rename migration APPLIED + VERIFIED — all four agents on preferred result_from".
- **what:** The contract by which a completing workflow's result is extracted (result_from / output_field(s) / flatten / mapping). Its failure was the root of the flagship silent no-op (singular output_field ignored → page collapsed to a stub). Evolution: flatten fix + contract logging + deprecation census (2026-06-18) → Option A unification: the resolution table lifted verbatim into datahelpers/result_contract.go (ResolveResultSpec + ApplyResultSpec), complete_workflow delegating to it, agents migrated to preferred key spellings — one source of truth for coordinator and action paths.
- **sources:** docs019/RUNBOOK_gamesdesign_index_rebuild.md#what-this-exercises; docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 2); docs019/RUNBOOK(31)_diagnosis_loop.md#6G (the fixture cause)
- **relations:** gamesdesign fixture; oversize-result delivery; loop_scope_field lesson (dotted-config class); parent result key under step name
- **verify-later:** datahelpers/result_contract.go + 7-case test; NNN_rename_complete_keys_preferred.sql applied state

### Oversize-result delivery: loud failure, size guards, responses-are-summaries
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "MEASURED (2026-07-03): … cd_bytes=1,270,781 — THE COMPLETION RESPONSE CARRIED THE FULL collected_data alongside the 6KB result"; "FIX = NNN_fix_diagnose_complete_output_fields.sql … APPLIED 2026-07-03"; gamesdesign_index_rebuild fix #3 "an undeliverable (oversize) result now fails loudly (error_unrecoverable + agent_error_log) instead of a status:'completed' stub".
- **what:** The family of oversize-result failures and their doctrine. Mechanism found in code: `result_from` is a key CompleteWorkflowAction never reads, so its fallback shipped the ENTIRE collected_data (1.27MB > Kafka ~1MB cap → Message Size Too Large; child fails at complete). Fixes: output_fields selection per agent; a response SIZE GUARD (max_response_bytes, truncatedResponseStub naming where the real result lives); earlier, removal of the silent "Full result exceeded size limit" completed-stub in favour of loud error_unrecoverable. Doctrine: responses are summaries; heavy artifacts live in the DB, retrievable by correlation_id; raising the broker cap is inversion, last resort.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E attempt-1 blocker 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#6; docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 2)
- **relations:** workflow result contract; bundle size doctrine; diagnosis_artifacts egress (same principle)
- **verify-later:** workflow_actions.go size guard + truncatedResponseStub; MaxResultSizeBytes const

### Parent stores child result under the step name
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) follow-up 1 "RESOLVED 2026-07-03: jsonb_object_keys shows the child response stored under the STEP NAME call_diagnoser … 'diagnose-agent_result' never existed. Migration WRITTEN: NNN_fix_orchestrator_complete_key.sql".
- **what:** Engine behaviour: when a call step has no output_field, the child's response is stored in the parent's collected_data under the STEP NAME — imagined synthetic keys like `<agent>_result` never exist. Two orchestrators carried imagined keys; fixed by pointing complete steps at the real step-name keys. A recurring class with the dotted-config lookups.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 1)
- **relations:** workflow result contract; loop_scope_field lesson
- **verify-later:** collected_data ? 'call_diagnoser' on a post-migration diagnose run

### Orchestrator COMPLETED while child FAILED (body.status check)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B5 "orchestrator can show COMPLETED while the child FAILED — header status complete, body.status failed; consumers of child results must check body.status (behaviour, recorded)".
- **what:** Recorded platform behaviour: a parent orchestration's header status can read complete while the child's embedded body carries failed; any consumer of child results must check body.status, not the header. Adopted cross-thread from the tools chat's notes.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B5 (useful from their notes)
- **relations:** oversize delivery (child fails at complete); stage-by-stage verification
- **verify-later:** response-building code paths; parent/child rows of a failed diagnose run

### Diagnosis→Fix Loop programme (F0–F3)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "DISCUSSION COMPLETE for F0/F1 (2026-07-07): Q-A/B/C/D/F all decided — CUTOVER-READY … First slice: F0.1"; no build claimed.
- **what:** The v2 workstream turning the read-only diagnosis loop into a diagnosis→fix system, phased: F0 intake/observability/egress (documented route in and out, fetchable bundles, per-task running notes); F1 fix-on-a-branch; F2 council of reviewers + decision-maker with architecture-change visibility; F3 learning (bug records, guideline amendments, corpus enrichment). Mission: use everything available — code corpus, schemas, runtime, the guidelines themselves — with checks, balances and second opinions built in.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task; docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan; docs019/RUNBOOK_diagnosis_fix_loop(9).md#current-position
- **relations:** read-only diagnosis loop (the base); council of reviewers; docs026 stage-3 council agents (this register's own consumer)
- **verify-later:** diagnosis_artifacts migration; needs_diagnosis items; fixer agent existence

### diagnosis_artifacts bundle egress
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "DECIDED 2026-07-07 (owner): Q-A diagnosis_artifacts table, written through inside assemble (unified-table refinement: kind ∈ {bundle, iteration_note})".
- **what:** Durable per-iteration bundle persistence: a diagnosis_artifacts table written through inside the assemble action (zero workflow-shape change, deliberately off the tools chat's emit-adjacent surface), with a documented fetch route. doc_notes was considered and set aside (notes are prose for humans; bundles are machine-replayable evidence with different retention). Sizing memory: bundles ~60KB × ≤5 iterations vs the 1.27MB collected_data incident.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-A); docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** oversize delivery doctrine; per-task running notes; diagnose_assemble_bundle
- **verify-later:** diagnosis_artifacts table (exists?); assemble write-through code

### needs_diagnosis intake in a pipeline='diagnose' namespace
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-B needs_diagnosis item in a NEW pipeline='diagnose' namespace (null-site allowed; envelope extends 084; manual trigger retained)"; "ENABLER CONFIRMED 2026-07-07: anchorless (site-less) diagnosis runs now SURVIVE".
- **what:** Task input rides the existing work-item dispatch + immune system: a needs_diagnosis site_work_items row in its own pipeline namespace, with null-site allowed for pure code bugs (enabled by the tools chat's load_runtime error-routing so anchorless runs degrade gracefully — ~26 min / 5 iterations observed). The canonical envelope adopts/extends the tools chat's 084 trigger with subject_type/subject_key.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-B); docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** build pump + immune system (the ride); generic envelope trigger (retained manual path)
- **verify-later:** pipeline='diagnose' rows; 084_TRIGGER_diagnose_v1.sh subject fields

### Fix-on-a-branch with an isolated fixer agent
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-C separate fixer agent (isolated write token; constrained edit plan; gofmt+build in a spawned job pre-PR)" — decided 2026-07-07, not built.
- **what:** F1: a CONFIRMED diagnosis drives a proposed fix committed to a separate git branch via the git adapter, PR opened, human amends/ditches/applies. The loop's core stays read-only; the write surface is a SEPARATE fixer agent holding the only write token (the spawn token-gate pattern), producing a constrained edit plan validated by gofmt+build in a spawned job before the PR.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 2); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-C)
- **relations:** repo-cloning token gate (the pattern); council of reviewers (gate before finalising)
- **verify-later:** fixer agent definition; git-adapter write paths

### Council of reviewers with a decision-maker
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) F2 "Independent reviewers (roster above), each a small agent with its own curated context … a decision-maker aggregates" — designed 2026-07-06/07, open Q-E/G/H.
- **what:** Before any fix is finalised, independent specialist agents each judge it from their own perspective and send structured opinions (verdict-wire-style: verdict + citations + objections + suggested alternative) to a decision-maker. Initial roster: guidelines agent (adherence to 000-0xx — or did the guideline fall short), reuse agent (code AND docs), bug-historian, compliance/legal, pipeline guardians (one per master workflow, seeded from the builder relay map), and specialist knowledge agents ("we already have one of these"). Precursor idea from the thin slice: build-time liability and MORALITY review contributors applying a configured, layered standard with contested calls routed to a human.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 3); docs019/RUNBOOK_thin_slice(27).md#next-improvements (item 3)
- **relations:** hard-veto semantics; three-tier citation (opinion contract); docs026 council-agents stage
- **verify-later:** reviewer agent definitions (none yet); Q-G reviewer-context decision

### Hard-veto flag semantics for reviewers
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-D council topology — VETO SEMANTICS DECIDED (owner, 2026-07-07): parallel reviewers → decision-maker BY DEFAULT … a hard_veto flag, attachable at multiple scopes … converts that reviewer's negative verdict into a BLOCK".
- **what:** All council opinions are advisory by default and weighed together; a hard_veto flag — attachable per reviewer agent, per pipeline, or per tool/component, most-specific-scope contemplated — makes that reviewer's negative verdict blocking. Accessibility and legal are the motivating hard-veto cases. A guidelines-reviewer "the guideline itself fell short" finding leans side-task (gap, not violation), not block.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-D)
- **relations:** council of reviewers; learning layer (guideline-gap side-task)
- **verify-later:** where the flag lives (reviewer column vs council config)

### Architecture-change visibility detector
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-E architecture-change signals: packages touched breadth; platform/ vs actions/; exported-signature diffs vs the corpus; message/topic/schema/contract changes; migration presence. Which are load-bearing?" (open, F2-phase).
- **what:** Make it loud when a proposed change is accidentally fundamental — touching platform contracts, message shapes, many packages, exported signatures — before it ships; runs as one council reviewer. Candidate signals enumerated; exported-signature diffs against the code_symbols corpus is the notable reuse of the diagnosis infrastructure.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 4); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-E)
- **relations:** council of reviewers; code_symbols corpus
- **verify-later:** n/a (not built)

### Learning layer — bug records and guideline-amendment side-tasks
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) F3 "bug_records (category taxonomy, recurrence checks feeding the historian); guideline-amendment proposals routed to the human"; Q-D "guideline-gap SIDE-TASK (amendment PR against the guideline docs, human terminal, fix unblocked, F3 recurrence record)".
- **what:** The feedback layer: recorded bugs with a category taxonomy and recurrence checks (feeding the bug-historian reviewer so a class never repeats); when a fix exposes a guideline gap, a side-task raises an amendment PR against the guideline docs with the human as terminal approver while the fix itself proceeds; corpus and doc enrichment feed back into retrieval.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F3); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-D)
- **relations:** corpus enrichment policy; council (bug-historian); coverage baseline (guideline home)
- **verify-later:** bug_records table (absent); amendment-PR mechanism

### Loop-worthiness test (five-criteria intake doctrine)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** diagnosis_fix_loop(9) "LOOP-WORTHINESS TEST (doctrine — apply before every intake)" — applied three times in the same file (pilot #1 downgraded, candidate 2 forked, guides pilot confirmed).
- **what:** A task is loop material only when ALL hold: (1) a SYMPTOM about system behaviour, not a feature request; (2) a causal mechanism plausibly exists in code+data+runtime; (3) not answerable by one or two direct queries (mandatory cheap pre-check first); (4) bounded to one symptom; (5) verified CURRENT at intake — symptoms are perishable. Feature absences → build routes; quality judgements → council/auditors; one-query questions → the query. Demonstrated by downgrading the roadmap-gap "bug" (findable by reading two files) to a builder-queue item.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#loop-worthiness; docs019/RUNBOOK_diagnosis_fix_loop(9).md#previous-pilot-1
- **relations:** F0 guides pilot; falsification eval gate
- **verify-later:** n/a (doctrine)

### F0 pilot — the guides-route differential diagnosis
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "★ F0 PILOT — CONFIRMED 2026-07-07: nav links to a guides section that has no content" with pre-registered criteria; ordered after F0.1 plumbing.
- **what:** The chosen first fix-loop pilot: dartsonline published a Guides nav link and blank /guides/index.html while gamesdesign (same platform) has working guides — a two-site DIFFERENTIAL, the strongest evidence shape. Standing hypothesis for the loop to confirm/refute FROM CODE: reconcile_site_plan's routing table has no "guide" entry (blog-index present, tool commented out), so planner-emitted guide pages were silently dropped while nav — generated from the PLAN, not the built set — published the link. Two earlier pilot candidates were downgraded via the loop-worthiness test.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#f0-pilot; docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** loop-worthiness test; reconcile routing table; nav-grounded-in-built-set principle
- **verify-later:** load_work_item_actions.go routing table; the pilot's run artifacts once executed

### Per-task running notes via doc_notes (travelling docs reuse)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** diagnosis_fix_loop(9) "Q-F DIRECTION SET (2026-07-07): REUSE doc_notes. The terminal-diagnosis note already exists on their side (pending their 3b subject threading)"; "the diagnose-agent workflow is ALREADY rewired by them: emit → persist_note → complete".
- **what:** Live monitoring of what the loop is doing and why: per-iteration and per-step reasoning written to a task-specific notes home. Decision: reuse the tools chat's doc_plans/doc_notes infrastructure (terminal diagnosis note already wired via persist_note with a strict no-guessing subject gate); per-iteration rows are additional doc_notes entries pending the owning thread's sign-off; category convention `diagnosis`.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-F); docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** doc_plans/doc_notes infrastructure; reasoning-state handoff; thread-boundary convention
- **verify-later:** doc_notes rows with category diagnosis; persist_diagnosis_note action

### doc_plans/doc_notes travelling-docs infrastructure
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** diagnosis_fix_loop(9) "tools chat's travelling-docs infrastructure — REV-22 READ 2026-07-07: doc_plans/doc_notes tables LIVE (Stages 0–2 shipped)".
- **what:** DB-backed travelling documentation owned by the parallel tools thread: doc_plans (with a criteria-fence pattern usable for acceptance criteria) and doc_notes keyed by subject_type/subject_key; agents persist notes as workflow steps (persist_note with config.error_step routing and a subject gate that refuses to guess). Recorded here because the diagnosis workflow was rewired through it and the fix loop adopts it rather than building a rival.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** per-task running notes; tool-doc header rollout; tiered tool acceptance
- **verify-later:** doc_plans/doc_notes DDL; persist_note action wiring in diagnose-agent workflow

### error_step-inside-config gotcha and pod-reap evidence substitute
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** diagnosis_fix_loop(9) "New gotcha ADOPTED (their 001 §16 finding): error_step belongs INSIDE a step's config — step-LEVEL error_step is silently ignored (dormant bug instances exist in tool agents) … idle pods reap at ~3600s — the post-completion STATE DUMP (ProcessingHistory) is the accepted evidence substitute."
- **what:** Two operational facts: workflow error routing only works when `error_step` sits inside the step's `config` object (top-level placement is silently ignored — dormant instances exist and should be corrected when touching a workflow, as its own noted change); and spawned agent pods are reaped ~3600s after idle, so post-mortem evidence comes from the orchestration state's ProcessingHistory dump, not pod logs.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** stage-by-stage verification; standing evidence rules
- **verify-later:** error_step placement across tool agent workflows; agent-job-cleanup timing

### Builder route method — map what exists before building (§B0 census)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "Rule honoured: map what EXISTS against what the problem statement wants BEFORE creating anything. Sources: the 147-row agent_definitions census (2026-07-03)"; §B0 findings enumerated.
- **what:** The builder route's opening method: an inventory matrix of problem-statement capabilities (intake, research, planning, design, content, tools, feeds, infographics, build/deploy, improvement, observability) against the ~147 existing agent types. Findings: every section except infographics has agents; the real defect is ~8 overlapping top-tier "build the site" orchestrators; the per-section content family is already prototyped; genuine gaps are the infographics owner and the success-factor synthesis step. Liveness comes from pump + handler references, not the status column.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B0; docs019/RUNBOOK_builder_route(21).md#B0-findings
- **relations:** three builder generations; work-item relay spine; vertical-exemplar researcher (the gap filled)
- **verify-later:** agent_definitions census queries; duplicate-row Q1

### Three coexisting builder generations
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B1 "Three generations coexist: GEN-1 (template era) … GEN-2 (in-memory multipage) … GEN-3 (component/spec/DB era — the LIVE architecture): pageflow-builder v20 (ACTIVE)" (dumps read 2026-07-04).
- **what:** The archaeology of site building: GEN-1 template chains (strategist→architect→writer→html-assembler→site-deployer), GEN-2 in-memory multipage (chief-strategist→content loop→assemble→deployer-agent, no components/specs/review), GEN-3 component/spec/DB (pageflow-builder v20's full inline build; site-work-orchestrator as its queue-native sibling with dynamic per-item handlers and maintenance mode). Explains duplicate deployers (Q3: site-deployer serves GEN-1; deployer-agent GEN-2/3) and frames consolidation.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#open-questions (Q3)
- **relations:** builder census; work-item relay spine (the decision among them)
- **verify-later:** workflow dumps of the nine builders; pageflow-builder v20 definition

### The work-item relay spine (baton/hop model)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B3 "DECISION (pre-stated rule fires): the relay reaches page-build-handler natively ⇒ THE SPINE = THE WORK-ITEM RELAY"; MILESTONE 2026-07-06 "first end-to-end domain→deployed site through the relay" (dartsonline.com).
- **what:** The settled build architecture: work moves as a relay of site_work_items batons — each names a handler_agent; the 30s pump claims unclaimed batons and spawns the named agent; the agent does one job, writes findings to site_specs (the site's shared notebook — spec-not-message, the 1.27MB lesson), creates the next baton, stops. Full chain: domain-submitter/adoption → classifier → (vertical research) → strategist → briefing → build-site-planner (emits needs_page/design/imagery/rerender items) → page-build-handler per page → rerender/deploy. Observed extra hops: needs_composition→site-design-planner, needs_design→webdesign-agent, needs_imagery→image-build-handler, needs_rerender→rerender-pages; page items are item_type needs_page. pageflow-builder survives as intake's initial-build convenience.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3; docs019/RUNBOOK_builder_route(21).md#B4 (plain-language explainer, map corrections); docs019/RUNBOOK_builder_route(21).md#milestone
- **relations:** build pump + immune system; builder generations; roadmap scope-decision gap; site quality programme (first output's gaps)
- **verify-later:** load_work_item_actions.go routing; the 37-row dartsonline item chain

### Build pump and the queue immune system
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "the scheduler fires build-pipeline-trigger EVERY 30s … The queue's immune system is all ENABLED: claimed-item-timeout (evidence-based auto-complete …), feasibility-recheck, stale-orchestration-reaper, stale-work-item-reaper (48h), work-item-archiver, database-cleanup. FLAG: improvement-sweep is DISABLED."
- **what:** What drives the relay: scheduled build-pipeline-trigger (30s, pre_query gated, concurrency dispatch/8) → build-dispatch-loop → atomic claim → spawn dynamic handler → complete/fail → touch scheduled_tasks. The immune system self-heals the queue (claimed-item-timeout does evidence-based auto-complete, its SQL documenting the gamesdesign false-positive lesson; feasibility-recheck unblocks when handlers appear; reapers and archiver bound staleness). Standing flag: improvement-sweep is disabled platform-wide, so the improvement loop is not running; content-feed-refresh is enabled 6-hourly.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2
- **relations:** work-item relay spine; needs_diagnosis intake (rides the same machinery); site quality LEG 6
- **verify-later:** scheduled_tasks rows (build-pipeline-trigger, improvement-sweep enabled flags); claimed-item-timeout SQL

### Two front doors and duplicate classifiers (Q5)
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "Two front doors, two classifiers (overlap)"; queue item 2 "[MAIN] Q5 front-door consolidation — two classifiers, one responsibility" (queued, undecided).
- **what:** Intake exists twice: the queue door (domain-submitter → work-item relay with domain-research-classifier) and intake-orchestrator v3 (HITL: site-classifier → confirm type → questionnaire → briefing-agent → spawn dynamic builder). site-classifier and domain-research-classifier hold the same responsibility; the classifier prompt hardcodes recommended_builder="pageflow-builder"; intake carries orphaned rerender steps. Consolidation direction (deprecate the intake door vs align contracts) is an open decision.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#queue (item 2)
- **relations:** work-item relay spine; site_type taxonomy drift; adoption fidelity inversion
- **verify-later:** intake-orchestrator usage evidence (orchestration_names ILIKE intake); site-classifier workflow

### Adoption-first fidelity inversion
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "Q4 ANSWERED — adoption does NOT call the classifier; the lean is inverted … adoption writes the specs FIRST; the classifier, when the relay later reaches it, CONSUMES them under its fidelity rules"; Q7 answered in §B3.
- **what:** How site adoption meets the relay: site-adoption-agent does the heavy work (firecrawl 30 pages, no-LLM design/interactive fingerprints, three LLM analyses — site analysis, archetype snapshot with improvement-loop constraints, content-direction guide), writes specs + pages + work items, then hands off needs_domain_research into the relay; the classifier's adoption-fidelity block treats adopted identity/archetype/content_direction/design_intent as ground truth outranking its own search+scrape. apply_adoption_plan writes site_archetype reading from collected data regardless of declared input_fields.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2; docs019/RUNBOOK_builder_route(21).md#B3 (Q7)
- **relations:** work-item relay spine; vertical-exemplar researcher (adopted sites run the hop too)
- **verify-later:** site-adoption-agent workflow; check_adoption_skip_scrape branch; classifier fidelity prompt block

### vertical-exemplar-researcher — the exemplar-research relay hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "§B4 CLOSED — QUALITY VERIFIED 2026-07-06 … ✔ CONSUMPTION PROVEN: the strategy's gap_opportunity QUOTES the hop … ✔ TRANSMISSION THREE HOPS DEEP".
- **what:** The first new build of the builder route: a reuse-only agent (one DB row, zero new Go) inserted as needs_vertical_research between classifier and strategist. Twelve-step workflow: read specs → LLM exemplar selection (3 of the vertical's best sites, flat keys, own domain forbidden) → 3× shallow firecrawl + format → synthesis LLM (per-exemplar success factors, cross-exemplar patterns, adopt/adapt/avoid lessons, differentiation opportunity — REASONS NOT COPIES) → write_site_spec aspect=vertical_landscape → chain needs_strategy. Verified end-to-end on dartsonline: real vertical leaders selected, causal synthesis, quoted by the strategy, differentiator surfaced in the plan. Design calls: shallow-many vs adoption's deep-one; specs-not-messages; strategist prompt nudge so the research is read.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (design calls, change-set, re-run verification)
- **relations:** work-item relay spine; adoption fidelity; coverage baseline (curated-list reuse candidate); image_tag trap (its spawn incident)
- **verify-later:** NNN_seed_vertical_exemplar_researcher.sql; vertical_landscape spec rows; reroute migration chain_config

### image_tag 'latest' stale-default trap
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "INCIDENT 2026-07-06 — first claim STALLED … THE ONE REAL DIFFERENCE: image_tag='latest' (column default) … the registry's latest is an ANCIENT chassis build … FIX APPLIED … NEW PARKED TRAP (systemic): agent_definitions.image_tag DEFAULTS to 'latest' — every future seeded agent inherits it."
- **what:** Seeded agents inherit image_tag='latest', which points at an ancient pre-architecture chassis build (boots the retired generic.process consumer regardless of env) — the newly seeded researcher stalled on it. Immediate fix: copy image columns from a live donor in every seed. Systemic options parked: repoint/retire `latest`, ALTER the column default, or a New Agent checklist line. Rollback convention is the same lever inverted: revert by repointing image_tag to the prior tag. Same staleness class as the HEAD-pinned index. Follow-up question: does deploy bulk-bump pinned tags (all five tool rows updated at once suggests yes)?
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (incident); docs019/RUNBOOK_builder_route(21).md#queue (item 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#8 (rollback)
- **relations:** stale-corpus class; standing evidence rules (seed hygiene)
- **verify-later:** agent_definitions image_tag column default; whether redeploy-agents bumps rows

### Roadmap-phases scope decision gap (nav grounded in built reality)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 6 "PROMOTED 2026-07-07 — THE BUG IS PLATFORM-WIDE … 082_submit_domain_unified.sh accepts ONLY --mission … AND build-site-planner's prompt has NO ELSE BRANCH for the roadmap-authority block … an absent decision point, not a missing default."
- **what:** No submitted site ever gets a roadmap/phases decision: the submit script has no --roadmap path and the planner's phase-discipline instructions vanish (not degrade) without one — so commerce-shaped domains get aspirational full plans and nav links to unbuildable pages. Fix shape (relay-wide, by construction): a post-classification scope-decision hop writes a phased roadmap_brief (P1 content/guides/tools; P2 legal-gated affiliate; P3 catalogue); planner prompt gains the ELSE branch (default phase-1-only or HITL hold); nav generation grounded in the BUILT set regardless of plan. Guidelines 001 already define the roadmap/phases mechanism — the docs had it, intake didn't. The legal gate on P2 is named as the fix-loop council's first concrete reviewer job.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 6); docs019/RUNBOOK_diagnosis_fix_loop(9).md#root-context
- **relations:** F0 guides pilot (nav-vs-built strand); coverage baseline; council compliance reviewer
- **verify-later:** 082_submit_domain_unified.sh flags; build-site-planner roadmap_brief template block; nav-updater

### Coverage baseline — guides, tools, news, curated top-N on most sites
- **category:** NEW:build-pipeline
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 7 "standing expectation going forward is most sites should carry guides + tools + news + a curated (LLM-picked, non-affiliate) top-N list … the curated-list mechanism, which IS new"; "STANDING EXPECTATION HOME: 001_development_guide … NOT the per-message prompt (decays), NOT the constitution (dev method)."
- **what:** A platform content-coverage policy: most sites should carry guides, tools, news, and a curated non-affiliate top-N list of the vertical's best products/services with outbound links; "pages need not be original to be best-in-class — genuinely useful common content counts". Enforcement points are the strategist/planner prompts (relay-wide-fixes-every-site logic); the curated-list mechanism is the one genuinely new build (reuse candidates: research-agent or the exemplar-researcher crawl pattern feeding a curation step). The mechanism for guides/tools/news EXISTS (gamesdesign, gaswholesalers prove it) — dartsonline's absence is a broken route, not a missing feature.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** F0 guides pilot (the broken route); roadmap gap (same enforcement points); site quality LEG 5
- **verify-later:** 001 guideline amendment; strategist/planner prompt coverage clauses

### Commented-out tool route and the planned-tool-page seam
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B3 "COMMENTED-OUT FUTURE ROUTES present: entity-directory, entity-page, and 'tool' → tool-build-handler (needs_tool_page)"; §B5 "ON HOLD: coordination with the parallel tools chat … The §B5 interface — how a PLANNED tool page reaches the pipeline — is a JOINT decision".
- **what:** The relay's reconcile routing table carries a commented "tool" → tool-build-handler route, so planned tool pages (e.g. dartsonline's headline tool-setup-builder differentiator) ship as prose via page-build-handler. Design fork recorded for the joint decision: (i) thin tool-build-handler driving generation into the synced page (page-creation conflict); (ii) tool-generator gains an existing-page mode; (iii) most reuse-shaped — no handler, a relay hop runs tool-suggester after site_plan and its pipeline owns page creation end-to-end. Accepted sequencing: ship prose first, upgrade later.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3 (C3); docs019/RUNBOOK_builder_route(21).md#B4 (§B5 candidate); docs019/RUNBOOK_builder_route(21).md#B5
- **relations:** work-item relay; tool pipeline (active suggester/generator/deployer); thread-boundary convention
- **verify-later:** load_work_item_actions.go commented routes; the joint-seam decision record

### site_type taxonomy drift between classifier and strategist (Q8)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B2 "vocabulary drift between hops — the classifier's site_type set (brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social) vs the strategist's canonical set (brochure|authority-portal|local-directory|review-site|content-hub|landing-page|portfolio). Two taxonomies for the same concept, one spec chain." Queued item 3.
- **what:** Two adjacent relay hops use different canonical vocabularies for the same site_type concept flowing through one spec chain — a contract-drift hygiene defect awaiting a one-canonical-set decision plus two snapshot migrations.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2 (Q8); docs019/RUNBOOK_builder_route(21).md#queue (item 3)
- **relations:** two front doors Q5; workflow result contract (drift class)
- **verify-later:** classifier and strategist prompt enumerations

### Site Quality Programme — the three-way split and seven legs
- **category:** NEW:site-quality
- **status-signal:** partial
- **status-evidence:** site_quality(1) "MEASURED BASELINE (2026-07-06, the four rendered pages)" table (zero nav/img/svg/script everywhere) and the A/B/C split with legs 1–7; handed off from builder §B6 2026-07-06.
- **what:** The programme closing the gap between "deploys" and "best in class" for relay-built sites, evidence-first: split failures into A dispatched-but-stuck (LEG 1 site chrome, LEG 2 design/stylesheet delivery, LEG 3 imagery items), B delivered-but-poor (LEG 4 content depth, LEG 7 link integrity), C never-in-scope (LEG 5 feeds/graphics/games as planning criteria, LEG 6 the disabled improvement loop) — and fix in that order (dispatch before content before scope). Pre-stated decision rules; the diagnosis loop named as the deeper instrument when a direct read is ambiguous.
- **sources:** docs019/RUNBOOK_site_quality(1).md#the-task; docs019/RUNBOOK_site_quality(1).md#three-way-split; docs019/RUNBOOK_builder_route(21).md#B6
- **relations:** work-item relay; build pump (improvement-sweep disabled); content-regression guard; coverage baseline
- **verify-later:** §B6 query set results; /assets/css/styles.css existence in the sites repo

### Site-chrome gap hypothesis (relay path lacks chrome rendering)
- **category:** NEW:site-quality
- **status-signal:** unknown
- **status-evidence:** site_quality(1) "Zero <nav> on every page ⇒ hypothesis: the RELAY build path lacks the site-chrome rendering step (pageflow-builder has render_site_components …; build-site-planner … has populate_nav_tables — nav DATA — but no chrome-render step was observed)."
- **what:** Open hypothesis from the measured baseline: relay-built pages ship without header/footer/nav because the relay path never runs an equivalent of pageflow-builder's render_site_components; nav DATA exists (populate_nav) but chrome is never rendered/injected at assembly. Was briefly the F0 pilot before being reassigned; remains LEG 1's core question.
- **sources:** docs019/RUNBOOK_site_quality(1).md#measured-baseline; docs019/RUNBOOK_diagnosis_fix_loop(9).md#f0-pilot-original
- **relations:** site quality programme; work-item relay spine; F0 pilot history
- **verify-later:** site_components rows for dartsonline; assembler chrome injection; render_site_components reachability from the relay

### Content-regression guard on section save
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) worked example "NOTE the content-regression guard (L227) PROTECTS the content-rich page from being overwritten by an empty shell"; gamesdesign_index_rebuild Stage B "content regression blocked: new content has N chars vs M existing → … a loud, correct block, not a silent no-op. … Do not disable the guard to force it through."
- **what:** save_page_sections refuses to overwrite substantially richer live content with a much thinner regeneration (~3k vs ~13–15k chars observed) — so a stale page can be the guard PROTECTING content, presenting as success-no-change. Doctrine: a block means investigate why the writer produced thin content upstream, never disable the guard. Also the designed skip path: empty sections → skipped→complete via the workflow's check_skipped conditional.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#worked-example; docs019/RUNBOOK_gamesdesign_index_rebuild.md#5 (Stage B); docs019/RUNBOOK(31)_diagnosis_loop.md#update-5537ffdb
- **relations:** workflow result contract (the upstream thin-content cause); site quality LEG 4
- **verify-later:** save_page_sections_action.go regression check; check_skipped conditional

### Stage-by-stage rebuild verification and the false-complete rule
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** gamesdesign_index_rebuild §5 stages A–E with per-stage SQL; "status='complete' is only meaningful together with Stage C showing changed components; complete + unchanged components = the old false-complete".
- **what:** The verification method for a page rebuild: A writer delivered a flat result to the parent (sections_metadata path check) → B save attempted/blocked loudly (agent_error_log) → C components actually changed (content_hash/updated_at fingerprint vs baseline) → D work item completed on a REAL save (complete only meaningful with changed components) → E deploy. Baseline-first (capture fingerprints before triggering), re-open the existing work item rather than fabricating one, and the triage table maps each stopping stage to its likely cause.
- **sources:** docs019/RUNBOOK_gamesdesign_index_rebuild.md#2; docs019/RUNBOOK_gamesdesign_index_rebuild.md#5; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7
- **relations:** oversize delivery (fix #3); content-regression guard; standing evidence rules
- **verify-later:** page_components fingerprint queries; site_work_items re-open pattern

### Standing evidence rules (the working-method contract)
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) header "Standing rules: user runs all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not decisive until the query itself is checked."
- **what:** The recurring operating contract of every runbook in this unit: the human runs all mutations/builds; outcomes are read by correlation_id, never `ORDER BY … LIMIT 1` (twice a red herring); `\d <table>` before every query; every agent_definitions change snapshots the row first (`snapshot_agent` = byte-exact revert path); a 0-rows result proves nothing until the query/selector is validated (wrong key, wrong label, wrong nesting all produced false zeros); migrations are self-guarded (UPDATE 0 = assumption wrong, nothing changed) and carry REVERT blocks.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#header; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7 (0-rows reminder)
- **relations:** instrument skepticism; repo-label bug (LIMIT-1 lesson); diagnostician seed→fix (snapshot rule)
- **verify-later:** snapshot_agent function; snapshots table growth

### Parallel-thread boundary and handoff convention
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "THREAD HANDED OFF 2026-07-06 → HANDOFF_builder_thread.md"; §B5 "BOUNDARY (adopted): the other chat owns everything INSIDE the tool pipeline …; This chat owns the RELAY …; The §B5 interface … is a JOINT decision, not taken unilaterally"; fix_loop "RULE: any fix-loop change to diagnose workflows is fetch-first against the CURRENT JSON and coordinated".
- **what:** Multiple concurrent working threads (builder, tools, quality, fix-loop, imagery) each own declared surfaces; runbooks record explicit boundaries, joint-decision seams, collision surfaces, and fetch-first rules for shared state; work moves between threads via handoff documents and "this item retains / that thread owns" dispositions. This is how the runbook families themselves relate — each family is one thread's travelling state.
- **sources:** docs019/RUNBOOK_builder_route(21).md#handoff-banner; docs019/RUNBOOK_builder_route(21).md#B5; docs019/RUNBOOK_diagnosis_fix_loop(9).md#boundaries; docs019/RUNBOOK_site_quality(1).md#boundaries
- **relations:** doc_plans/doc_notes; per-task notes; documentation-system travelling docs
- **verify-later:** HANDOFF_builder_thread.md; boundary sections across sibling runbooks

### Tool-doc header rollout (provenance + stripped headers)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** thin_slice(27) "Tool-doc header rollout (2026-06-11) — apply order is load-bearing. … Three stages; do not reorder — the gate without the prompt fails every generation, and the stamps without the columns fail every insert." No completion claim in this unit's files.
- **what:** Rollout procedure for tool documentation headers: (1) provenance columns on content_components (source_agent_type, source_orchestration_id), (2) anchored idempotent prompt updates adding the `=== tool-doc ===` header requirement (abort if prompts drifted), (3) one binary release (tool_doc_header.go + five action edits) so headers are stamped in the DB template but STRIPPED from shipped pages/CDN assets, with a tool_health no_doc_header WARNING converging old tools on the normal sweep — no retrofit campaign.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#tool-doc-header-rollout
- **relations:** doc_plans/doc_notes (the tools thread's later system); tiered tool acceptance
- **verify-later:** content_components source_% columns; '=== tool-doc ===' in html_template rows; tool_health sweep items

### Tiered tool acceptance (static contract check + browser-runner)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Their Stages 5–6 define a TIERED ACCEPTANCE system for tools: a static Tier-2 contract-presence check and a Tier-4 browser-runner adapter (Chromium+Playwright, Kafka request/response per the 035 Adapter Guide) — their 'loop for complicated tools' is acceptance/verification + docs, NOT a rival diagnosis loop."
- **what:** The tools thread's acceptance ladder for generated tools, recorded here as a shared component: Tier-2 static contract-presence verification and a Tier-4 browser-runner adapter executing tools in real Chromium — also earmarked as a future verification service for fix-loop F1 fixes touching pages and as a council reviewer's instrument.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** council of reviewers; tool pipeline; adapters (035 guide)
- **verify-later:** browser-runner adapter existence; tool acceptance stages in the tools thread docs

### Base-runbook gated-items framing (PLAN.md linkage)
- **category:** diagnosis-loop
- **status-signal:** superseded
- **status-evidence:** RUNBOOK.md §6 "Gated items (carried — see PLAN.md) … None are unblocked by this thread's work alone" — replaced from RUNBOOK(1) onward by the inlined "§6 Completing the whole task (what remains)" with per-step DoD.
- **what:** The earliest form of the diagnosis runbook kept only in-flight build steps and deferred the roadmap to a separate PLAN.md; within one version the roadmap was inlined as §6 with per-step definitions of done and live status, and the runbook became the single self-contained thread state (later §7 split out to its own file when §6 closed). Family-delta record of the project's documentation style converging on self-contained travelling runbooks.
- **sources:** docs019/RUNBOOK.md#6; docs019/RUNBOOK(1).md#6; docs019/RUNBOOK(31)_diagnosis_loop.md#6 (ACTIVE ROUTE MOVED banner)
- **relations:** parallel-thread convention; documentation-system
- **verify-later:** PLAN.md in docs019 (sibling file, other unit)
