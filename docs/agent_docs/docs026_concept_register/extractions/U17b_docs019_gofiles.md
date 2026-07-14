# EXTRACTION U17b — docs019_documentation_audit_autonomous_build_and_operate/go_files/
Extracted 2026-07-13. Files in scope: 74. Concepts found: 31.

Scope note: this unit narrows the failed U17 to just the `go_files/` subtree (the
contextkit tooling tree: analyser, assembler, dbcontext, bundle, diagnose, embed,
dedup, fuse, resolve_targets, eval_targets, thin_versions + internal/). Per the
task charter for this unit, Go code bodies were NOT analysed — only README*.md,
other *.md docs, `thin_slice_constitution.md`, `groundtruth_targets.json` (small),
and each Go/shell file's header comment (header-scan) were read for intent.

## Coverage
| file | treatment |
|---|---|
| NNN_create_code_indexer_agent.sql | full |
| contextkit(6).tar.gz | skipped-binary |
| contextkit.tar.gz | skipped-binary |
| contextkit/001_more_potential_thin_slice_prompt.md | full |
| contextkit/README(0).md | family-delta |
| contextkit/README(2).md | full |
| contextkit/README.md | full |
| contextkit/README_how_to_run_analyser.md | full |
| contextkit/RUNBOOK_doc_archiving.md | full |
| contextkit/analysis.json | skipped-generated |
| contextkit/analysis1.json | skipped-generated |
| contextkit/analysis2.json | skipped-generated |
| contextkit/analysis3.json | skipped-generated |
| contextkit/analysis4.json | skipped-generated |
| contextkit/analysis5.json | skipped-generated |
| contextkit/analysis6.json | skipped-generated |
| contextkit/bundle_diagnosis_loop.sh | header-scan |
| contextkit/bundle_minilobby_trim2.sh | header-scan |
| contextkit/bundle_minilobby_trim3.sh | header-scan |
| contextkit/bundle_recreation_v1.sh | header-scan |
| contextkit/chassis.json | skipped-generated (>1MB) |
| contextkit/cmd/analyser/main.go | header-scan |
| contextkit/cmd/assembler/main.go | header-scan |
| contextkit/cmd/bundle/README.md | full |
| contextkit/cmd/bundle/main.go | header-scan |
| contextkit/cmd/dbcontext/main.go | header-scan |
| contextkit/cmd/dedup/main.go | header-scan |
| contextkit/cmd/dedup/stage_docs019_migration.sh | header-scan |
| contextkit/cmd/diagnose/main.go | header-scan |
| contextkit/cmd/embed/main.go | header-scan |
| contextkit/cmd/eval_targets/main.go | header-scan |
| contextkit/cmd/fuse/main.go | header-scan |
| contextkit/cmd/resolve_targets/main.go | header-scan |
| contextkit/cmd/thin_versions/main.go | header-scan |
| contextkit/embeddings.json | skipped-generated (>1MB, ~43MB) |
| contextkit/go.mod | full |
| contextkit/groundtruth_targets.json | full |
| contextkit/groundtruth_targets.json.orig | family-delta |
| contextkit/image-analysis.json | skipped-generated (>1MB) |
| contextkit/internal/analysis/analyse.go | header-scan |
| contextkit/internal/analysis/symbolbody.go | header-scan |
| contextkit/internal/analysis/symbolbody_test.go | header-scan |
| contextkit/internal/analysis/types.go | header-scan |
| contextkit/internal/candidates/types.go | header-scan |
| contextkit/internal/diagnose/adapters_test.go | header-scan |
| contextkit/internal/diagnose/advance.go | header-scan |
| contextkit/internal/diagnose/advance_test.go | header-scan |
| contextkit/internal/diagnose/callgraph.go | header-scan |
| contextkit/internal/diagnose/diagnose_doc_catalogue.example.json | skipped-generated |
| contextkit/internal/diagnose/diagnose_query_catalogue.example.json | skipped-generated |
| contextkit/internal/diagnose/docselect.go | header-scan |
| contextkit/internal/diagnose/docselect_test.go | header-scan |
| contextkit/internal/diagnose/gatherer.go | header-scan |
| contextkit/internal/diagnose/loop.go | header-scan |
| contextkit/internal/diagnose/loop_datarequest_test.go | header-scan |
| contextkit/internal/diagnose/loop_scopeguard_test.go | header-scan |
| contextkit/internal/diagnose/loop_test.go | header-scan |
| contextkit/internal/diagnose/queryselect.go | header-scan |
| contextkit/internal/diagnose/queryselect_test.go | header-scan |
| contextkit/internal/diagnose/sqlguard.go | header-scan |
| contextkit/internal/diagnose/sqlguard_literal_test.go | header-scan |
| contextkit/internal/diagnose/sqlguard_test.go | header-scan |
| contextkit/internal/diagnose/step.go | header-scan |
| contextkit/internal/diagnose/step_test.go | header-scan |
| contextkit/internal/diagnose/verdict_wire.go | header-scan |
| contextkit/internal/diagnose/verdict_wire_test.go | header-scan |
| contextkit/lex.json | skipped-generated |
| contextkit/sem.json | skipped-generated |
| contextkit/thin_slice_constitution.md | full |
| contextkit/vonc-analysis.json | skipped-generated (>1MB) |
| groundtruth_targets(1).json | skipped-generated |
| groundtruth_targets(2).json | skipped-generated |
| groundtruth_targets(3).json | skipped-generated |
| groundtruth_targets(4).json | skipped-generated |

Note on README(0)/README(2)/README.md: these three are NOT a strict linear
version chain — `README(0).md` is a short superseded stub (tree diagram only);
`README.md` and `README(2).md` diverge (each carries content absent from the
other: `README.md` has the chassis-drafts staging + Makefile/kustomize "still
to create" plan, `README(2).md` has the fuller "where everything lands in the
agentchassis tree" destination map). Both were read in full; `README(0).md` was
scanned only for delta.

## Concepts

### contextkit CLI toolkit
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "go build ./...  # compiles all seven commands" (README.md); real invocation example wired to a live site in cmd/bundle/README.md (`-runtime-site gamesdesign.co.uk`).
- **what:** A standalone Go module (`module contextkit`, go 1.22) of CLI tools for building LLM context bundles from a repo without a live cluster: analyser, assembler, dbcontext, bundle, embed, resolve_targets, fuse, eval_targets, dedup, thin_versions. Compiles and runs independently of the agentchassis repo; two of its packages (`internal/analysis`) are shared verbatim with the chassis.
- **sources:** contextkit/README.md, contextkit/README(2).md, contextkit/go.mod
- **relations:** diagnosis loop (internal/diagnose), analyser-adapter deployment plan, thin-slice constitution
- **verify-later:** does `internal/analysis` in this tree still match `internal/analysis/` at the agentchassis repo root byte-for-byte (README.md flags this as a manual sync obligation, not automated)

### analyser (cmd/analyser)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "thin CLI wrapper over analysis.AnalyseWithExclude — the same parsing primitive the in-cluster analyser adapter imports"
- **what:** Walks a Go source tree and emits a structural-summary JSON (files, packages, imports, function/method signatures with callee names, struct/interface declarations with line ranges). Always skips vendor/, testdata/, hidden dirs, `*_test.go`, and `*(N).go` download-duplicates; takes an `-exclude` list for repos (like this one) that store archived copies of their own code under docs/.
- **sources:** contextkit/cmd/analyser/main.go#header, contextkit/README.md
- **relations:** internal/analysis package, code-indexer agent (chassis-side counterpart), embed/resolve_targets (consume analyser JSON)
- **verify-later:** internal/analysis (agentchassis repo root) — confirm the in-cluster analyser adapter still calls `analysis.Analyse` (no-exclude) as documented

### internal/analysis package (analyser output contract)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Package analysis defines the contract the analyser emits and the assembler, embed, and resolve_targets consume... defined only here" (types.go header); "the harness and production parse identically" (analyse.go header)
- **what:** The single-source-of-truth `Output`/`FileInfo`/`FuncDef`/`TypeDef` contract for repo structural analysis, plus `Analyse`/`AnalyseWithExclude` (the layer-1 AST walk) and `ReadSymbolBody` (slices a `path:Symbol` scope into source text using the analyser's recorded line span, never re-parsing). Intentionally Go-only; a non-Go producer would fill the same contract behind the analyser adapter.
- **sources:** contextkit/internal/analysis/types.go#header, contextkit/internal/analysis/analyse.go#header, contextkit/internal/analysis/symbolbody.go#header
- **relations:** analyser, assembler, embed, resolve_targets, cmd/bundle (also uses the symbol slicer), chassis diagnose_assemble_bundle action
- **verify-later:** whether the chassis's diagnose_assemble_bundle action's old inline `readSymbolBody` stub has actually been replaced by a call to `ReadSymbolBody` as the header claims is the intent

### assembler (cmd/assembler)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "That's the whole thin slice working: analyser, constitution, assembler." (README_how_to_run_analyser.md)
- **what:** Builds one paste-ready markdown bundle for a single task: consumes the analyser JSON, the repo (to pull full bodies by line range), the flat constitution, and a task+scope spec; renders constitution, task, in-scope code in full, neighbourhood signatures (same-package, capped ~60/package), schema (hand-fed), and a pointers note of what was omitted. `-step` (framing/implementation/debug) controls altitude — framing shows signatures only, implementation/debug show full in-scope bodies.
- **sources:** contextkit/cmd/assembler/main.go#header, contextkit/README_how_to_run_analyser.md, contextkit/001_more_potential_thin_slice_prompt.md
- **relations:** internal/analysis (symbol slicing), thin_slice_constitution.md, bundle (wraps it), docselect/queryselect (chassis analogues for doc/query selection instead of hand-specified scope)
- **verify-later:** confirm the neighbourhood-signature cap and package-scoping behaviour match what 001_more_potential_thin_slice_prompt.md's design notes describe

### dbcontext (cmd/dbcontext)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** used directly in cmd/bundle/README.md's worked example (`-psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db'`)
- **what:** Fetches DB context for a bundle by shelling out to a configurable `psql` — schema (`\d <table>`, complete+bounded), rows (multipass-sized SELECT: full if within cap, else sample + pointer query, never an unbounded dump), and capabilities (`\dx`, `\df`). No Go DB driver; psql does the talking, so it inherits whatever connection role/permissions the operator supplies.
- **sources:** contextkit/cmd/dbcontext/main.go#header
- **relations:** bundle (wraps it), sqlguard (lints model-written queries elsewhere in the pipeline), database-and-infrastructure conventions
- **verify-later:** whether the psql connection used in production is provisioned as a read-only role (sqlguard.go explicitly says the lint alone is not the safety boundary — the read-only transaction/role is)

### bundle (cmd/bundle)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** concrete real invocation against `gamesdesign.co.uk` in cmd/bundle/README.md; used by bundle_diagnosis_loop.sh, bundle_minilobby_trim2/3.sh, bundle_recreation_v1.sh
- **what:** A thin orchestration wrapper around dbcontext + assembler: gathers read-only DB context (schema/capabilities/runtime evidence), writes each to a temp file, then invokes the assembler with those files wired in. Deliberately never runs SQL itself (that stays in dbcontext) so the assembler stays a pure, read-only, offline composer — the wrapper "triggers NOTHING — no builds, no spawns, no writes."
- **sources:** contextkit/cmd/bundle/main.go#header, contextkit/cmd/bundle/README.md
- **relations:** dbcontext, assembler, gatherer.go (BundleGatherer shells out to this exact binary from the diagnosis loop)
- **verify-later:** BundleGatherer.buildArgs (gatherer.go) — confirm the flag set it constructs still matches this binary's real flags

### dedup (cmd/dedup)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "these tools were behaviour-tested, but they MOVE FILES" (RUNBOOK_doc_archiving.md); default-report/`-move` design documented as used in the docs019 archiving runbook
- **what:** Finds duplicate/near-duplicate files in a directory tree; two passes — EXACT (SHA-256, deterministic) and NEAR (optional, shingled-token Jaccard similarity ≥ threshold, heuristic). Report-only by default; `-move` relocates non-canonical copies into an archive dir with a full undo manifest (TSV), never deletes. Canonical-selection tie-break: not-archived > not a `(N)` download-dup > shallowest > shortest > newest.
- **sources:** contextkit/cmd/dedup/main.go#header, contextkit/RUNBOOK_doc_archiving.md#Step-1
- **relations:** thin_versions (distinct: copies vs versions), stage_docs019_migration.sh (delegates to it for step 2)
- **verify-later:** `dedup-manifest.tsv` output location and whether the docs019 archiving pass in RUNBOOK_doc_archiving.md was actually executed against the live docs tree

### thin_versions (cmd/thin_versions)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "go build ./cmd/dedup ./cmd/thin_versions    # both compile" (RUNBOOK_doc_archiving.md pre-flight)
- **what:** Reduces version-sprawl by grouping files that are successive versions of the same document (subject-stem derived by stripping `.patch/.orig/.bak`, extension, trailing `(N)` bracket, trailing `_vX`/`_vX_Y`, then a second `(N)`), keeps the newest N per group, and moves older versions to an archive dir on request. Recency rank within a subject is version-number first, then `(N)` bracket, then mtime — deliberately so a stale-dated-but-later version still ranks above an earlier one.
- **sources:** contextkit/cmd/thin_versions/main.go#header, contextkit/RUNBOOK_doc_archiving.md#Step-2
- **relations:** dedup (run first to clear exact copies before thinning versions)
- **verify-later:** whether the docs024 "18 fat clusters of 10+ versions each" identified in the runbook were actually thinned

### embed (cmd/embed)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "-local ... proves the pipeline (index → cosine → rank) WITHOUT a model, but is NOT semantic — use -ollama for real recall" (header)
- **what:** Builds/queries a semantic vector index over the analyser's symbols — the recall layer for target resolution sitting above the lexical baseline (resolve_targets). Model-agnostic via an embedder interface: `-ollama` (real embeddings, e.g. nomic-embed-text) or `-local` (deterministic offline token-hashing stand-in for pipeline-proving only). Index and query must use the same embedder/vector space.
- **sources:** contextkit/cmd/embed/main.go#header
- **relations:** resolve_targets, fuse (RRF-merges embed's output with resolve_targets'), eval_targets (scores it), code-indexer agent (chassis-side embedding via the same ollama-adapter/nomic-embed-text pairing)
- **verify-later:** whether production bundle-building actually runs `embed` with `-ollama` or still relies on the `-local` stand-in

### resolve_targets (cmd/resolve_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the deterministic baseline — the layer that runs before any embeddings" (header)
- **what:** A first-cut, lexical-overlap target resolver: given a task string and the analyser JSON, proposes ranked candidate symbols/files to `-scope` by matching the task's distinctive words against each symbol's name, path, and docstring. Does not decide — proposes a ranked candidate set for a human or the assembler to confirm.
- **sources:** contextkit/cmd/resolve_targets/main.go#header
- **relations:** embed (semantic counterpart), fuse (merges both), internal/candidates (shared output contract), eval_targets
- **verify-later:** —

### fuse (cmd/fuse)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "score(item) = sum over lists of 1/(k + rank_in_list), k=60 (standard)" (header)
- **what:** Merges ranked candidate lists (resolve_targets' lexical output + embed's semantic output) into one ranking via reciprocal-rank fusion (RRF). Combines by RANK not score specifically because the lexical integer scores and semantic cosine scores aren't on a comparable scale.
- **sources:** contextkit/cmd/fuse/main.go#header
- **relations:** resolve_targets, embed, internal/candidates, eval_targets (scores fuse's output too)
- **verify-later:** —

### eval_targets (cmd/eval_targets)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** the ground-truth file contains a task tied to a "REAL fix (2026-01 chassis regression)" already applied, showing the harness is exercised against real cases, not just synthetic ones
- **what:** Scores a resolver's candidate list (`-json` output of resolve_targets/embed/fuse) against a ground-truth set mapping tasks to the symbols they actually needed — turns "the fused list looks better" into numbers: recall@N over decisive symbols, and MRR contribution (rank of first decisive hit). Match is on `path:name`.
- **sources:** contextkit/cmd/eval_targets/main.go#header, contextkit/groundtruth_targets.json
- **relations:** resolve_targets, embed, fuse (all scored by this), llm-quality-testing (evaluation-harness pattern), ground-truth eval set concept below
- **verify-later:** —

### diagnose (cmd/diagnose)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "THE VERDICT STEP IS NOT THE REAL MODEL HERE... a chassis-side follow-on (needs a model). This entrypoint ships two stand-ins" (header)
- **what:** Wires the diagnosis-loop scaffold (internal/diagnose) to real adapters — BundleGatherer (shells to cmd/bundle, read-only) and AnalysisCallGraph (follows the analyser's `calls` for re-scope). The verdict step is stubbed (either a scripted JSON array of verdicts for testing, or a trivial always-UNVERIFIABLE default) since the real cite-or-abstain LLM verdicter needs a model and lives chassis-side. Explicitly read-only and human-gated: emits a diagnosis + evidence trail, never a fix, never a triggered run.
- **sources:** contextkit/cmd/diagnose/main.go#header
- **relations:** internal/diagnose (loop.go, step.go, verdict_wire.go, callgraph.go, gatherer.go), fixloop workstream (the diagnose→fix pipeline this scaffold feeds)
- **verify-later:** docs024_key_docs_latest/fixloop_eg_dartsonline/ for whether/how a real LLM verdicter has since been wired in chassis-side

### diagnosis-loop scaffold (internal/diagnose, loop.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "WHAT LIVES HERE (deterministic, testable without a model): loop control, the guards..., the evidence trail, and the re-scope mechanism. WHAT DOES NOT... the Verdict step" (loop.go header); backed by loop_test.go, loop_datarequest_test.go, loop_scopeguard_test.go
- **what:** The deterministic core of the diagnosis loop: wraps a read-only gather step around a pluggable verdict step, enforces convergence guards (iteration cap, scope-must-narrow, evidence-must-grow, no-thrash), accumulates an evidence trail, and re-scopes by FOLLOWING runtime/call-graph evidence rather than re-searching the symptom — named as the fix for a "ceiling" where symptom-only retrieval fails on infrastructure-layer causes. Non-negotiable boundary: never applies a fix, never triggers a run to test a hypothesis.
- **sources:** contextkit/internal/diagnose/loop.go#header, contextkit/internal/diagnose/loop_scopeguard_test.go#header, contextkit/internal/diagnose/loop_datarequest_test.go#header
- **relations:** step.go (DecideStep), advance.go (chassis-facing wrapper), callgraph.go, verdict_wire.go, docselect.go, queryselect.go, sqlguard.go, gatherer.go, fixloop workstream
- **verify-later:** whether the "guard-vs-expansion" bugfix noted in loop_scopeguard_test.go (run 17933a83) and the data_request evidence-growth fix (loop_datarequest_test.go, "truncated the live gamesdesign runs at iteration 3") are reflected in the currently-deployed chassis diagnose_run/diagnose_route actions

### DecideStep — shared pure per-iteration decision (step.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Extracting it keeps ONE source of truth for the guard + re-scope logic instead of two copies that could drift. Run() is refactored to call Step(); the existing tests are the proof the behaviour is unchanged." (header)
- **what:** The per-iteration decision (given iteration state, a verdict, the call graph, and guard memory) extracted as a pure function, shared by the standalone `Run()` loop and the chassis `diagnose_run` workflow action (where the verdict is a separate workflow step). Guarantees one source of truth instead of two logic copies that could drift apart.
- **sources:** contextkit/internal/diagnose/step.go#header, contextkit/internal/diagnose/step_test.go
- **relations:** loop.go, advance.go (LoopState calls this per-iteration)
- **verify-later:** confirm the chassis `diagnose_run` action actually calls this shared `Step()`/`DecideStep` rather than a re-implementation

### LoopState — chassis-facing per-iteration API (advance.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "the chassis needs: (1) a LoopState it can thread through workflow collected_data between iterations, (2) a single Advance() call per iteration, and (3) parse helpers for a verdict that arrives as an already-unmarshalled map" (header)
- **what:** The workflow-driven realisation of the loop: since the chassis loop is `gather → verdict step → diagnose_route → back | emit` (not an in-process loop), LoopState carries loop memory across workflow steps via `collected_data`, with `Advance()` as the one call per iteration and `EncodeLoopState`/`DecodeLoopState` for the JSON round-trip. Adds no new decision logic beyond step.go's DecideStep plus state bookkeeping.
- **sources:** contextkit/internal/diagnose/advance.go#header, contextkit/internal/diagnose/advance_test.go
- **relations:** step.go, loop.go, chassis diagnose_route workflow step
- **verify-later:** platform/orchestration — the actual `diagnose_route` step and its `collected_data` schema, to confirm it matches `EncodeLoopState`'s shape

### AnalysisCallGraph — call-graph re-scope (callgraph.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "HONEST LIMIT (from the analyser): `calls` is NAME-BASED, not type-resolved... DELIBERATELY DROPS ubiquitous names (Run, String, Error, …)" (header)
- **what:** A CallGraph implementation backed by the analyser's recorded (name-based, not type-resolved) `calls` field, letting the diagnosis loop re-scope by following the call graph from an evidence-named site rather than re-searching the symptom. Explicitly drops ubiquitous method names that would otherwise explode the neighbourhood into noise — the loop's narrowing guard is the backstop, but dropping known-ubiquitous names keeps re-scope sharp at the source.
- **sources:** contextkit/internal/diagnose/callgraph.go#header
- **relations:** internal/analysis (the `calls` data it consumes), loop.go's re-scope mechanism, cmd/diagnose (wires this in as the real adapter)
- **verify-later:** —

### verdict wire format (verdict_wire.go)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "If the prompt's output schema and this struct ever drift, the loop breaks at this join — so this file is tested against example model outputs." (header)
- **what:** The JSON wire format the (LLM) Verdicter emits — human-legible strings (`"REFUTED"`, `"runtime"`) and snake_case keys rather than the domain type's int enums, so a model can produce it reliably — and the parser (`ParseVerdict`) translating it into the domain `Verdict`. Named as the ONE seam between the verdict prompt's specified output (`docs/PROMPT_diagnosis_verdict.md`) and the scaffold; a verdict-script in this format is a faithful stand-in for the real model.
- **sources:** contextkit/internal/diagnose/verdict_wire.go#header, contextkit/internal/diagnose/verdict_wire_test.go
- **relations:** diagnose (cmd), loop.go, docs/PROMPT_diagnosis_verdict.md (referenced, not in this unit)
- **verify-later:** docs/PROMPT_diagnosis_verdict.md — confirm its schema still matches this struct (the header itself flags drift risk here)

### docselect — per-hypothesis doc selection (docselect.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Selection is DETERMINISTIC and testable... so it can be exercised without a model. It is a HEURISTIC (keyword/path substring)" (header); docselect_test.go exercises keyword/always/path-glob rules
- **what:** Per-hypothesis selection of authored context docs (the 003 contract sections, 016 §9 entries, dev-guide sections) to paste into the CURRENT iteration's bundle rather than every doc into every bundle — avoiding the "irrelevant context buries the signal" failure mode. A future extension is floated (not built): letting the verdict NAME a needed doc via a `needed_docs` field mirroring `needed_evidence`/`next_scope`.
- **sources:** contextkit/internal/diagnose/docselect.go#header, contextkit/internal/diagnose/docselect_test.go
- **relations:** thin_slice_constitution.md (the always-on layer this supplements), queryselect.go (data analogue), contracts-and-standards (003)
- **verify-later:** —

### queryselect — vetted read-only query catalogue (queryselect.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "WHY A CATALOGUE, NOT MODEL-WRITTEN SQL (the safety boundary)... the queries are HAND-WRITTEN, parameterised, and \\d-verified ONCE; the loop only SELECTS among them by hypothesis. The model never writes SQL." (header)
- **what:** Per-hypothesis selection of vetted, read-only, parameterised DB queries for the runtime-evidence gather — the data analogue of docselect.go. Queries bind to the loop's existing context (site_id, domain, page, correlation_id already in input_data/seed), so no wire-format change or model-supplied SQL parameters are needed. This is presented as THE safety boundary for runtime evidence, distinct from sqlguard's lint-only role.
- **sources:** contextkit/internal/diagnose/queryselect.go#header, contextkit/internal/diagnose/queryselect_test.go
- **relations:** docselect.go, sqlguard.go, dbcontext (executes the chosen queries)
- **verify-later:** —

### sqlguard — IsReadOnlySQL lint (sqlguard.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "READ THIS FIRST — what this is NOT: This is NOT the safety boundary... The REAL guarantee is the EXECUTION SUBSTRATE" (header); three dedicated test files including a literal-false-positive regression (sqlguard_literal_test.go)
- **what:** A cheap pre-flight lint for model-written diagnosis queries, explicitly documented as defence-in-depth, NOT the safety guarantee — the real guarantee is the execution substrate (chassis: read-only transaction + non-multi-statement protocol; harness: a read-only DB role) plus a statement_timeout. Includes a regression fix for keywords/`;` appearing inside quoted string literals (triggered by a real page slug `tool-drop-rate-simulator` containing "drop").
- **sources:** contextkit/internal/diagnose/sqlguard.go#header, contextkit/internal/diagnose/sqlguard_literal_test.go#header
- **relations:** queryselect.go (the actual safety boundary via hand-vetted catalogue), dbcontext
- **verify-later:** confirm the chassis execution path really does use `db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})` as claimed, and that the harness's `-psql` role is genuinely read-only in practice

### BundleGatherer (gatherer.go)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "READ-ONLY by construction (DESIGN §4): bundle runs dbcontext... + the pure assembler. Nothing here triggers a build, spawn, or write." (header)
- **what:** A Gatherer that shells out to `cmd/bundle` to produce each iteration's bundle, translating a `Scope` into bundle flags and returning the written bundle path. Adds no capability beyond what `cmd/bundle` already does — just drives it per iteration with the loop's evolving scope.
- **sources:** contextkit/internal/diagnose/gatherer.go#header
- **relations:** cmd/bundle, cmd/diagnose (wires this in as the real gatherer)
- **verify-later:** —

### ranked-candidate contract (internal/candidates)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "Defined once here so the shape isn't re-declared as `candFile`/`jc` in each tool." (header)
- **what:** The shared `Candidate`/`File` JSON contract (`path`, `name`, `kind`, `score` as float64, `rank`, `task`, `method`) that resolve_targets, embed, and fuse all emit with `-json`, and that fuse and eval_targets read — replacing what used to be duplicated per-tool struct definitions.
- **sources:** contextkit/internal/candidates/types.go#header
- **relations:** resolve_targets, embed, fuse, eval_targets
- **verify-later:** —

### thin-slice constitution (always-on rules doc)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Included in full in every bundle... Later it becomes the `standards` rows with `scope = constitution`; the content is the same." (thin_slice_constitution.md)
- **what:** The flat-file version of the chassis's always-on rules, pasted into every assembler/bundle output: reuse-before-recreate, fix-structural-not-symptoms, every-agent-is-an-orchestrator, no subworkflows-in-SQL (spawn sub-agents instead), the snake_case/kebab-case naming split, storage conventions (text+CHECK enums, version+previous_version_id, deleted_at soft-delete), logging rules (no `logger.Debug`, log the orchestration_id/correlation_id), deployment path (GitHub → Actions → Backblaze S3), and plain/pragmatic generated-text tone. Task-specific 003 contracts are listed but pulled in only when a task touches them.
- **sources:** contextkit/thin_slice_constitution.md
- **relations:** assembler (always includes it), docselect.go (adds the task-specific 003 sections this doc defers), contracts-and-standards (003)
- **verify-later:** whether the constitution has since actually migrated to `standards` rows with `scope = constitution` as the doc anticipates, or is still the flat file

### ground-truth eval set for target resolution
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** task `silent-norebuild-resultspec` is tagged "REAL fix (2026-01 chassis regression): coordinator honoured only plural output_fields; writer declares singular output_field; compiled page collapsed into the skipPatterns dump -> 'completed' stub -> stale page. Fix = resolveResultSpec treats singular as flatten."
- **what:** `groundtruth_targets.json` maps tasks (deliberately symptom-only task strings, scrubbed of the vocabulary a resolver could trivially match — see the note on an earlier version leaking "extracted"/"result" and inflating a lexical rank) to the "expect" (decisive) and "also_useful" symbols a resolver must surface. Grew across versions: the `.orig` predecessor holds only the `skinner-box` task; the current file adds `silent-norebuild-resultspec`, drawn from a real, already-fixed chassis regression (result-spec singular vs plural output_field handling).
- **sources:** contextkit/groundtruth_targets.json, contextkit/groundtruth_targets.json.orig
- **relations:** eval_targets, resolve_targets/embed/fuse (evaluated against this set), llm-quality-testing
- **verify-later:** platform code for `result_spec.go:resolveResultSpec` / `coordinator.go:extractWorkflowResult` — confirm the fix described is actually live

### code-indexer agent (analyser-adapter's chassis-side counterpart)
- **category:** diagnosis-loop
- **status-signal:** aspirational
- **status-evidence:** "DRAFT — modelled on the real agent_definitions rows you sent... Confirm the live schema before applying" (NNN_create_code_indexer_agent.sql); status column set to `'experimental'` in the INSERT itself
- **what:** A draft `agent_definitions` row for a `code-indexer` orchestrator agent: workflow is `request_analysis` (calls `request_repo_analysis` action, asking the analyser adapter to parse a repo@ref into symbols) → `index_symbols` (calls `index_code_symbols`, upserting into `code_symbols`, embedding changed symbols via an ollama/nomic-embed-text endpoint, pruning symbols absent from the commit) → `complete`. Retrieval side is a separate `lookup_code_symbols` action used by other agents. Coordination-only orchestrator; the real parsing work happens in the analyser-adapter pod.
- **sources:** NNN_create_code_indexer_agent.sql
- **relations:** analyser (the parsing primitive this indexes), embed (same embedder pairing: ollama + nomic-embed-text), analyser-adapter deployment plan, snapshot-before-mutate practice
- **verify-later:** `agent_definitions` table (`\d agent_definitions`) for the real CHECK constraint on `agent_category` and NOT NULL/default columns before this migration is applied; whether `code_symbols`, `index_code_symbols`, `lookup_code_symbols`, `request_repo_analysis` exist yet

### analyser-adapter deployment/migration plan
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** README.md marks most destinations "(NEW)"/"(EDIT)" against the real repo tree (tree -d, 2026-06-11), but notes "`NNN_create_code_symbols_index.sql` → workspace root (ALREADY APPLIED — commit for the record, your numbering)"
- **what:** A directory-by-directory migration map from a `chassis-drafts/analyser-adapter` staging area (which does not compile in this tree) to real agentchassis destinations: `cmd/analyser-adapter/main.go`, `internal/adapters/analyser/{adapter,analyse_action,github_source}.go`, `platform/orchestration/actions/{code_symbols_actions,analyser_request_action}.go` (+ registry.go insertion), the code-indexer migration, `configs/analyser-adapter.yaml`, a two-stage Dockerfile, and kustomize base/overlay scaffolding — all following the conventions already used for thunder-adapter. Also flags un-placeable items needing a human call: the `035_adapter_guide.md` doc home, the `system.adapter.analyser.requests` KafkaTopic CRD location, and the `analyser-github-read` Secret (never committed with a real token).
- **sources:** contextkit/README.md, contextkit/README(2).md
- **relations:** code-indexer agent, adapters (033/035 thunder/webscrape pattern), deployment-github (034)
- **verify-later:** build/docker/backend/, deployments/kustomize/services/analyser-adapter/, Makefile — confirm which of the four described insertions actually exist

### documentation archiving subproject (docs019 cleanup)
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** "measured 2026-06-13 from main_docs_directory_tree.txt: 2,729 files collapse to 1,917 subjects; 1,734 are singletons; the noise is concentrated in 18 fat clusters of 10+ versions each (~489 files)"
- **what:** A four-step, report-first, fully-reversible plan to reduce the docs tree so the analyser's index isn't diluted: dedup (exact/near copies) → thin_versions (older versions) → editorial re-home of surviving current docs into an `engines/`+`runbooks/` structure (human judgement, "classify don't merge") → re-point links and rebuild the analyser index, verified by a python one-liner checking zero stale/archived paths remain indexed. Explicitly out of scope: merging documents, judging content currency (deferred to a separate `DESIGN_doc_drift_classifier.md` tool), touching the 1,734 singleton subjects.
- **sources:** contextkit/RUNBOOK_doc_archiving.md
- **relations:** dedup, thin_versions, stage_docs019_migration.sh, doc-drift classifier (below), documentation-system (037)
- **verify-later:** whether `_archive/`, `dedup-manifest.tsv`, `thin-manifest.tsv`, and `PROPOSED_MOVES.tsv` exist in the live docs tree, i.e. whether this runbook was actually executed

### docs019 migration staging script (stage_docs019_migration.sh)
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** "DEFAULT IS REPORT ONLY. Nothing moves without --apply. Even --apply touches only (1); (2) needs --dedup; (3) is never auto-applied." (header)
- **what:** Automates the deterministic half of the docs019 archiving plan: auto-moves obviously-superseded archive directories (`go_files_old/`, `thin_slice_run/`, `working/`, a stray `archive_april_26/`) into `_archive/` on `--apply`; can also invoke the dedup tool (`--dedup`); and for the editorial third (loose `FOCUS_`/`PLAN_`/`RUNBOOK_`/`016_`/`NOTES_` docs), only writes a `PROPOSED_MOVES.tsv` for a human to edit (action column: move|archive|skip|keep) and apply by hand — deliberately never auto-applied.
- **sources:** contextkit/cmd/dedup/stage_docs019_migration.sh#header
- **relations:** dedup, thin_versions, documentation archiving subproject
- **verify-later:** —

### doc-drift classifier (named, not built)
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** "It does not judge document CONTENT currency — that is the doc-drift classifier (`DESIGN_doc_drift_classifier.md`), a separate, later, T1+T2-first tool." (RUNBOOK_doc_archiving.md)
- **what:** An idea named but explicitly deferred and not built in this unit's scope: a future tool to judge whether a document's CONTENT is still current (distinct from dedup/thin_versions, which only judge copy/version redundancy, never content currency).
- **sources:** contextkit/RUNBOOK_doc_archiving.md#What-this-subproject-does-NOT-do
- **relations:** documentation archiving subproject, dedup, thin_versions
- **verify-later:** search the wider docs tree for `DESIGN_doc_drift_classifier.md` — it may exist as a design doc outside this unit's scope even though not built as a tool

### snapshot-before-mutate practice (snapshot_agent / take_site_snapshot)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** pasted `\df`-style listing shows both functions already exist: `snapshot_agent(uuid, p_agent_type text[, p_reason text])` and `take_site_snapshot(uuid, p_site_id, p_trigger, p_git_sha, p_label, p_created_by)`; also invoked directly in the code-indexer SQL comment ("Snapshot before re-applying to an existing row: SELECT snapshot_agent('code-indexer');")
- **what:** A working convention of snapshotting an agent_definitions row (or a site) via a DB function before applying a migration that touches it, so changes are reversible without relying on git history alone. Paired with a documentation-discipline request to "start or update a runbook, a running notes and a plan" alongside any such migration.
- **sources:** contextkit/001_more_potential_thin_slice_prompt.md, NNN_create_code_indexer_agent.sql
- **relations:** code-indexer agent (applies this practice), site-snapshots-and-revert (014)
- **verify-later:** `\df snapshot_agent` / `\df take_site_snapshot` against the live clients_db schema to confirm current signatures

### vonc.com mini-lobby content-edit re-render architecture
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "docs 003/002 say HTML patching was REJECTED as an edit mechanism ('content_data is always the source of truth … if we only patched rendered_html, the edit would be lost on the next re-render')" (bundle_minilobby_trim2.sh / bundle_minilobby_trim3.sh header)
- **what:** The established (not this-unit's-invention) rule that content edits must go through `content_data`, never by patching `rendered_html` directly, because the next re-render would discard a raw HTML patch. Two re-render paths exist: the full path (`needs_page` → page-content-writer, LLM) and the light path (`rerender_page_sections` behind a `page_rerender` item, no LLM, re-renders each section from stored `content_data` via `RenderComponentAction`) — and neither is `rerender_single_page`, which only assembles already-rendered components. `fix_component_template_action`'s `remove_element` fix_type explicitly does NOT touch `page_components.rendered_html` content, because "content changes go through the section-editor workflow" — leaving the correct edit path for a structural trim (like the provocation-card mini-lobby) genuinely unclear without a bundle to check.
- **sources:** contextkit/bundle_minilobby_trim2.sh#header, contextkit/bundle_minilobby_trim3.sh#header
- **relations:** vonc (site-case-studies), section-editor workflow, content-governance (013), styling-render-pipeline (036)
- **verify-later:** platform/orchestration/actions/fix_component_template_action.go and rerender_page_sections/rerender_single_page to confirm the scope boundaries described are still accurate; whether the mini-lobby trim task itself was ever completed

### action-name-to-file resolver (bundle_recreation_v1.sh)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "v1 grepped only for the quoted name inside action files and excluded registry.go — which is precisely where the mapping lives. File names are not consistent either: validate_page_content.go has no _action suffix." (header)
- **what:** A path-resolution helper that, given a registered action NAME, finds its source FILE by (1) reading the registration line in `registry.go` to get the constructor/type name, (2) finding the file defining that constructor/type, (3) falling back to CamelCasing the action name and searching, (4) last-resort whole-platform-tree search — built specifically because file naming is inconsistent (some action files lack the `_action` suffix) and a prior version's naive grep missed paths by excluding the one file (`registry.go`) where the authoritative name→type mapping actually lives.
- **sources:** contextkit/bundle_recreation_v1.sh#header
- **relations:** bundle, resolve_targets (a cruder, deterministic alternative to lexical/semantic resolution for a KNOWN action name)
- **verify-later:** —

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
