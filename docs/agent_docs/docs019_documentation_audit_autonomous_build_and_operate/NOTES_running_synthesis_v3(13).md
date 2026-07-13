# NOTES — running synthesis (contextkit diagnosis loop, continuation thread)

Update this EVERY turn: append a dated entry under TURN LOG, fold anything
load-bearing into DECISIONS or RISKS, and refresh the STATE DIGEST so a fresh
chat can resume from the top. Timestamps are the date plus the deploy image
version (the build environment exposes the date, not wall-clock time); add
precise times if you need them. Memory is OFF; this file is the durable record.
Companion files: `RUNBOOK.md` (how to build/test/apply the work in flight) and
`PLAN.md` (workstream + gated items + changelog). For the loop-run procedure and
the broader project history, see the prior `RUNBOOK_design_diagnosis_loop*.md`
and `NOTES_running_synthesis_v2*.md` (already in the repo).

---

## STATE DIGEST (read first)

**Deploy status (2026-06-24).** Release image **v1.0.1073** deploying
(GitHub → GitHub Actions → Backblaze S3). The deploy carries the current chassis
state. Post-deploy verification: the §2 chassis build check, then the gated
analyser-path / eval items below — a deployed image is not yet evidence the loop
reasons.

**Current blocker (2026-06-24).** The agent rewrite landed: `diagnose-agent` now has
the `diagnose_route` workflow in `default_config` (verified `has_route=t`, no
`run_loop`). Remaining: (1) the `diagnose-orchestrator` move did NOT land (it shows
`has_wf=f` — workflow nulled out of `orchestration_workflow` without arriving in
`default_config`); restore it with `NNN_restore_diagnose_orchestrator_workflow.sql`.
(2) build + register the four diagnose actions and deploy. The coordinator
`next_step` override is CONFIRMED real (coordinator.go:1093 `getNextStepFromResult`).
RUNBOOK §6C. Standing rule added: snapshot every `agent_definitions` row before
changing it — `SELECT snapshot_agent('<type>','<reason>')` (the migrations do this).

**Correction (the data_requests channel is real and now wired).** Earlier I called it
a "gap to wire", then over-corrected to "dormant by design (catalogue boundary)". Both
wrong. `sqlguard.go`: the chassis runs model SQL under a **read-only transaction** (the
real guarantee), so model `data_requests` are SAFE in the chassis — the catalogue is
the STANDALONE's mechanism (role-based). So `runDataRequests` is the chassis's
DB-following channel and should be live. It was dormant from a 3-part wiring gap, now
fixed: gather_step=load_runtime (migration) + `diagnose_route` forwards
`verdict.DataRequests`→`route.data_requests` (PATCH_wire_data_requests.md) +
`diagnose_load_runtime` default `route.data_requests`. The engine's `Scope` carries
only code re-scope (call graph); the data channel is bridged by `diagnose_route`
directly. Still open: the `route.scope` vs `ExtractStringListHelper` field shape (CODE
re-scope round-trip), unchanged.

**What we are building.** contextkit — a task-scoped context-assembly + diagnosis
tool, dogfooded on the agent-chassis repo (`github.com/gqls/agentchassis`, Go +
Postgres + Kafka, k8s ns `ai-persona-system`). The immediate sub-thing is the
**diagnosis loop**: an agent that diagnoses chassis bugs READ-ONLY —
hypothesise → gather scoped read-only evidence → cite-or-abstain verdict →
re-scope by FOLLOWING evidence (call graph for code, vetted queries for data),
never re-searching the symptom. Human-gated, never a fix, never triggers a run.

**Where the code actually lives (corrected this thread).**
- Engine: `contextkit/internal/diagnose/` (package `diagnose`). The chassis has a
  PARTIAL port at `pkg/diagnose/` — it has advance/step/loop/callgraph/verdict_wire
  but is MISSING `sqlguard.go` (and likely docselect/queryselect). That partial
  port is why a bundle scope on `sqlguard.go:IsReadOnlySQL` did not resolve.
- The bundle assembler that slices symbol bodies is **`cmd/assembler`**.
  `cmd/bundle` is only an orchestration wrapper (it shells `cmd/dbcontext` for the
  read-only DB gather and `cmd/assembler` for the compose).
- Analyser contract: `internal/analysis` — `analyse.go` + `types.go`. `types.go`
  is the single source of truth for `Output`/`FileInfo`/`FuncDef`/`TypeDef`. NOTE:
  the contextkit and chassis copies of `analyse.go` have DRIFTED (contextkit has
  `AnalyseWithExclude` + the `name(N).go` duplicate-skip; the chassis copy is the
  older plain `Analyse`). The TYPES appear shared; the walker has diverged.

**DONE this thread.**
1. `bundle_diagnosis_loop.sh` — a `cmd/bundle` script that bundles the loop's own
   code + the four diagnose actions + governing docs + schema, for working on the
   loop (dogfooding). Flags verified against `gatherer.go` + `cmd/bundle/main.go`.
2. `analysis.ReadSymbolBody` (`symbolbody.go` + `symbolbody_test.go`) — the ONE
   slicer that turns a `path:Symbol` (or `path`) scope entry into source text by
   slicing the analyser's `start_line..end_line` spans. TESTED (go vet + unit test
   pass) and proven BYTE-IDENTICAL to `cmd/assembler`'s real bundle output.
3. `diagnose_assemble_bundle_action.go` — MERGED: the old `readSymbolBody` stub is
   deleted; the action now decodes `repo_analysis` (from `collected_data`) into a
   typed `analysis.Output` and calls `analysis.ReadSymbolBody`. gofmt-clean; the
   map→Output decode round-trip is tested.

**NEXT (in rough priority).**
- Drop `symbolbody.go` into BOTH `internal/analysis` copies (contextkit + chassis;
  separate module copies). Run `go test ./internal/analysis -run ReadSymbolBody`.
- Build the chassis action in your env (`go build ./platform/orchestration/actions/`)
  — it cannot compile in the sandbox (datahelpers/zap/ActionParams unavailable).
- Collapse `cmd/assembler` onto `analysis.ReadSymbolBody` (reuse — one slicer), and
  confirm its symbol resolution matches `spanOf` on the method-name-collision case.
- Port `sqlguard.go` (+ `sqlguard_test.go`) into chassis `pkg/diagnose` IF the
  chassis verdict path lints model SQL (it should — Guard 2). Then re-run the
  bundle script and confirm `sqlguard.go:IsReadOnlySQL` resolves.
- The larger GATED items carried from the handoff (unchanged): data-request
  EXECUTION wiring in `diagnose_load_runtime` (read-only tx + statement_timeout);
  analyser path verification in production; build/deploy + one hand-triggered run;
  and THE EVAL GATE — run the live loop on the gamesdesign bug + the 016 §9
  catalogue and confirm it reproduces the mid-course REVERSALS and ABSTAINS when
  unsettled. Scaffold correct ≠ reasons well.

---

## TURN LOG

### 2026-06-24 — turn 1: bundle script for the diagnosis-loop task
- Asked for a `cmd/bundle` script for the "diagnosis loop" task (template example
  given for the gamesdesign bug).
- `cmd/bundle` source was not mounted, so the flag contract was read off
  `gatherer.go` (`BundleGatherer.buildArgs`, the code that actually shells bundle):
  required `-analysis -root -constitution -task`, plus `-step -out`, repeatable
  `-scope`/`-doc`, DB block `-psql -schema-tables -runtime-site -runtime-page
  -capabilities`. `-include`/`-df-filter` came from the example.
- Engine symbol names were taken from the source files (not invented).
- Difficulty: empty `grep` of `some_schemas` for CREATE TABLE → checked the file
  rather than concluding; it is a `\d` dump, so the tables (agent_error_log,
  site_work_items, page_components, pages) ARE present. (The "0 rows is not
  decisive until the query is checked" rule paid off.)
- Decision: chassis root as the default (matches the example + RUNBOOK convention),
  with a clearly-commented contextkit ALT. Flagged that the four action filenames
  were convention-derived and that `pkg/diagnose`/actions being drafts could make
  scopes silently skip.
- Decision: bash ARRAY arg list over inline backtick-comment continuation (robust;
  parse-checked, incl. the `PSQL=""` drop path).

### 2026-06-24 — turn 2: read the produced bundle; design + write ReadSymbolBody
- The uploaded `bundle_diagnosis_loop.md` resolved the `pkg/diagnose/*` paths and
  the four actions → the chassis port is on disk. Six of seven engine scopes came
  through; `sqlguard.go:IsReadOnlySQL` did not (flagged).
- Target = the `readSymbolBody` STUB in `diagnose_assemble_bundle_action.go`.
- Structural read: the stub signature `(repoRoot, symbol)` cannot do the job the
  note asks for (slice from the analyser spans) because it is never handed the
  spans; re-parsing would duplicate the analyser. Decision: ONE shared slicer in
  the `analysis` package, reused by `cmd/assembler` and the action — not a
  chassis-local re-implementation.
- The `analysis` types file and `cmd/assembler` were not mounted. Rather than guess
  the `Output` shape (the exact mistake the handoff warns about), recovered the
  field names from how `analyse.go` CONSTRUCTS the structs, and the slicing
  convention from the bundle's OWN output — diffed `advance.go:Advance` lines 67–115
  against the rendered block: byte-identical (inclusive, 1-indexed, doc comment
  excluded).
- Wrote `symbolbody.go` + `symbolbody_test.go` but could not compile/test (no types
  file in the sandbox subset). Flagged it as needing the full module, and asked for
  `types.go` + `cmd/assembler`.

### 2026-06-24 — turn 3: uploads → tested + merged
- Uploads closed the gaps: `types.go` (confirmed every field ReadSymbolBody uses),
  both `analyse.go` copies, the real action file, `sqlguard.go`(+test), and
  `cmd/bundle/main.go`.
- The grep settled the sqlguard question: `IsReadOnlySQL` exists ONLY at
  `contextkit/internal/diagnose/sqlguard.go` → the chassis `pkg/diagnose` port
  never got `sqlguard.go`. That is the scope miss.
- `main.go` showed `cmd/bundle` is a wrapper → the real slicer is `cmd/assembler`;
  the byte-identical diff was against `cmd/assembler`'s output.
- Found `analyse.go` DRIFT between the two `internal/analysis` copies (flagged; the
  types look shared, the walker diverged).
- Built the real `analysis` package in a temp module (`analyse.go` + `types.go` +
  the new files): `go vet` clean, unit test PASS, and ReadSymbolBody on the real
  source reproduced `cmd/assembler`'s body byte-for-byte (the only delta was a
  trailing newline owned by the shared render template, which the helper correctly
  leaves to the caller). Validated the map→JSON→typed `Output` round-trip the
  action performs (12 files survive; ReadSymbolBody works on the result).
- MERGED the action: imports `encoding/json` + `internal/analysis`; decodes the
  Output once after the repoRoot check; loop calls `analysis.ReadSymbolBody`;
  deleted the stub; dropped the now-redundant `_ = analysisField`. gofmt-clean,
  syntax OK (full compile still needs your env).
- Created this NOTES file, `RUNBOOK.md`, `PLAN.md`.

### 2026-06-24 — turn 4: rename + deploy
- Renamed the running notes to `NOTES_running_synthesis_v3.md` (continues the
  repo's `_v2` series); updated the cross-references in `RUNBOOK.md` and `PLAN.md`.
- Release image **v1.0.1073** deploying (GitHub → Actions → Backblaze S3). No code
  change this turn; the deploy carries the current chassis state. Next verification
  is the §2 build check, then the gated analyser-path / eval items.

### 2026-06-24 — turn 5: runbook expansion
- Read `cmd/assembler/main.go`. Its body-slicer is `splitScope` + `locateSymbol` +
  `readLines`; `locateSymbol` matches by BARE name, first-match, NO `Type.Method`
  form, so `ReadSymbolBody`'s `spanOf` is a strict SUPERSET — collapsing the
  assembler onto the shared slicer cannot regress (it only adds the `Type.Method`
  case the assembler currently skips). `readLines` uses the same inclusive
  1-indexed convention (hence the byte-identical diff), differing only in an
  `end == -1` whole-file sentinel the named-symbol path never uses.
- Expanded `RUNBOOK.md`: a context-free high-level intro at the top; `$CK`/`$CHASSIS`
  exports threaded through every script (the contextkit path is
  `$CHASSIS/docs/agent_docs/docs019_.../go_files/contextkit`); §3 now details the
  collapse (which call to swap, what `locateSymbol` keeps, the byte-identical diff
  check); §4 now shows the analyse→bundle→grep recipe to confirm
  `sqlguard.go:IsReadOnlySQL` resolves; §6 now outlines the whole remaining task
  A–G to the eval gate.

### 2026-06-24 — turn 6: assembler collapse merged + verified; §6 expanded
- Merged the collapse into `cmd/assembler/main.go` (delivered as `assembler_main.go`):
  BOTH body reads now go through `analysis.ReadSymbolBody` — the named-symbol branch
  (`...(*root, *an, sc)`) and the whole-file branch (`...(*root, *an, path)`) — and
  the now-unused `readLines` is deleted. Real var name is `an` (not `ana`); the
  `filepath` import stays (still used at line 166).
- VERIFIED for real: built the merged assembler in a `contextkit`-shaped module
  against the analysis package (compiles), generated an analysis JSON of /mnt/project,
  ran the original and merged binaries over a bare-name func + a whole-file + a
  `Type.Method` scope — output BYTE-IDENTICAL.
- Correction (was wrong in turn-5's runbook §3): a `Type.Method` scope does NOT flip
  to a body under the collapse, because the named-symbol branch still calls
  `locateSymbol` for the header + `found` gate, so `ReadSymbolBody`'s receiver-qualified
  support is not reached in the assembler. The collapse changes nothing observable.
  Runbook §3 fixed.
- Expanded `RUNBOOK.md` §6 into runnable steps A–G (build/test, the read-only probe,
  the data-request execution code sketch, the `diagnose_ro` GRANT role, workflow
  loop-back checks via `orchestration_states.execution_path`, analyser-path checks
  via `code_symbols`, seed/register/deploy, a hand-trigger skeleton, and the eval
  gate). SQL grounded in the real `\d` columns; Go/trigger marked as sketch/skeleton.
  All 20 embedded bash blocks parse-checked.

### 2026-06-24 — turn 7 (image era: post-v1.0.1073)
- KEY FINDING (answers "is the diagnose workflow what's missing?" — yes): from the
  user's live `\d`, `agent_definitions WHERE category='diagnose'` → 0 rows and
  `orchestration_states WHERE workflow_plan LIKE '%diagnose%'` → 0 rows. Columns
  verified, so the 0 rows are real: NO diagnose agent is seeded. The missing piece
  is the seed migration (agent definition + workflow + the four registered actions).
  Reordered RUNBOOK §6 to put the seed up front as 6C (the current blocker).
- Schema note: `agent_definitions` has `default_config` AND `task_workflow` /
  `orchestrator_workflow` / `orchestration_workflow`. The design says the workflow
  lives in `default_config`, but §6C now FIRST inspects an existing live agent to
  confirm which column actually holds a workflow (reuse the live pattern, don't
  assume). Also: `category` is a free-text column (so `category='diagnose'` is
  fine); `agent_category` has a CHECK (strategist|executor|analyst|integrator|
  coordinator|specialist) that does NOT include 'diagnose'; `status` CHECK is
  active|experimental|deprecated|demo|template.
- Explained `type` in the runbook (the earlier "<the diagnose agent type from the
  row above>" confusion): it is the `agent_definitions.type` string you set in the
  seed INSERT; it fills the `{type}` topic templates and is the trigger's
  `target_agent_type`. The placeholder returned nothing only because the row is not
  seeded yet.
- Merged `runDataRequests` into the REAL `diagnose_load_runtime_action.go`: reads
  the prior verdict's `data_requests` from `route.data_requests`, runs each in a
  READ ONLY tx with `SET LOCAL statement_timeout`, appends rows; imports
  `pkg/diagnose` for `IsReadOnlySQL` (Guard 2). gofmt-clean. Needs §4 (sqlguard in
  `pkg/diagnose`) to build; confirm the `sql`/`why` wire keys against the verdict
  prompt; grep for an existing generic row formatter before keeping `formatRowsText`.
- Read-only probe switched to `site_flows` (user's request — safer; the table is
  empty and `WHERE false` touches nothing). User verified `BEGIN READ ONLY; DELETE
  FROM site_flows WHERE false; COMMIT;` errors with "cannot execute DELETE in a
  read-only transaction". Substrate (Guard 3) confirmed live.
- Added the `diagnose_ro` read-only role as a migration
  (`NNN_create_diagnose_ro_role.sql`): idempotent, GRANT-based SELECT-only, password
  from a secret, needs CREATEROLE to apply.
- RUNBOOK: added Created/Last-updated dates; a Status checklist with checkboxes; a
  "Where the files live" section answering that the `cmd/*` tools (assembler, bundle,
  dbcontext, analyser) are contextkit's and live under `$CK`, NOT in the chassis —
  the deployed agent uses the registered actions + `analysis.ReadSymbolBody`, not
  `cmd/assembler`; inlined the RUN procedure as a self-contained §7 (was a pointer
  to `RUNBOOK_design_diagnosis_loop*.md`). 24 bash blocks parse-checked.

### 2026-06-24 — turn 8 (image era: v1.0.1074)
- Confirmed from five live reference agents (build-dispatch-loop,
  site-work-orchestrator, domain-research-classifier, site-adoption-agent/-orchestrator):
  the workflow lives in `default_config.workflow` (task_workflow / orchestrator_workflow
  are NULL). Reference image tag is v1.0.1074.
- Read the workflow-creation guidelines (001 §"Workflows simple, complexity in Go",
  §"Don't create subworkflows in SQL", §"Spawn before call", §"How call_agent finds
  the spawned agent" → use target_role, §"Dynamic dispatch", Appendix C "Loop
  Mechanisms").
- DESIGN DECISION (the diagnosis-loop iteration): the `loop` action is FOR-EACH over
  a collection fixed at loop entry (Appendix C), so it cannot express the loop's
  WHILE semantics (next scope known only after each verdict). A workflow-internal
  cycle (route→assemble) relies on cyclic next_step the engine isn't shown to
  support. Per 001 ("spawn sub-agents, not subworkflows") and the build-dispatch-loop
  exemplar ("one unit per orchestration, re-invoke; separate orchestrations = clean
  logs"), chose ONE ITERATION PER ORCHESTRATION: load_runtime → request_analysis →
  lookup_symbols → assemble → verdict → route → route_branch(conditional) →
  [continue: spawn_next → call_next → emit | terminal: emit] → complete. On
  continue it spawns+calls a fresh diagnostician for the next iteration.
  Consequence: cross-iteration state travels in input_data (revised hypothesis,
  next_scope, data_requests, iteration+1 via call_next.input_mapping), NOT
  collected_data.
- Realigned `diagnose_load_runtime_action.go`: data_requests_field default
  "route.data_requests" → "input_data.data_requests" (matches re-invocation; the
  route.* default reflected the abandoned cycle design). Flagged variable change.
  Still gofmt-clean.
- WROTE `NNN_seed_diagnose_agents.sql` (the user has none): seeds the `diagnostician`
  agent, modelled on the reference agents — category='diagnose' (free text),
  agent_category='analyst' (CHECK list), status='experimental', is_active=true,
  workflow in default_config. Workflow JSON validated (parses; all step refs
  resolve). Seed step config overrides diagnose_assemble_bundle's scope/hypothesis
  fields to input_data.* (re-invocation). FLAGGED in the file: request_repo_analysis
  / lookup_code_symbols config keys; diagnose_route (outputs {continue, iteration})
  + diagnose_emit (picks child vs this-iteration verdict) must exist + be registered;
  paste PROMPT_diagnosis_verdict.md into the verdict step; re-invocation nesting
  (spawn+call same type, ≤max deep) vs flat alternative (fire to generic entry point
  + complete). RUNBOOK §6C updated to point at the seed; default_config confirmed.
- User confirmed the diagnose_ro role migration applied (DO / GRANT×3 / ALTER
  DEFAULT PRIVILEGES).

### 2026-06-24 — turn 9 (image era: v1.0.1074)
- Found the diagnose pair ALREADY EXISTS: `diagnose-agent` (worker) +
  `diagnose-orchestrator` (thin spawn-and-forward wrapper), seeded 2026-06-20.
  My `diagnostician` seed (turn 8) DUPLICATED it → SUPERSEDED (bannered the file,
  do not apply). Reuse-before-recreate caught late because the pair wasn't visible
  until this turn's `SELECT type FROM agent_definitions` (146 agents).
- COMPARED three designs for the loop: (1) the seeded `diagnose-agent` runs the
  whole loop inside ONE Go action `diagnose_run` (wraps engine Run()); (2) the
  chassis-integration doc's 06-17 banner wanted workflow-driven steps + a
  `diagnose_route` cyclic loop-back; (3) my re-invocation `diagnostician`. Best =
  (1): complexity in the tested Go engine, thin workflow, no cyclic next_step
  gamble, maximal reuse of Run/Advance/DecideStep/guards — and it matches the
  chassis-integration doc's OWN §4–§6 recommendation ("prefer diagnose_run wrapping
  the engine; don't re-express the guards as workflow branches"). Dropped (2)+(3).
- WORKFLOW COLUMN: the pair's workflow is in `orchestration_workflow` (json), but
  the loader reads `default_config` (where all 144 working agents keep it). Per the
  user's directive, wrote `NNN_move_diagnose_workflow_to_default_config.sql`: moves
  the column verbatim (UPDATE ... SET default_config = orchestration_workflow::jsonb
  || processing_mode/timeout_seconds, orchestration_workflow = NULL), with
  IS-NOT-NULL + default_config='{}' guards (idempotent, non-clobbering). Adds the
  processing_mode/timeout_seconds the seeded orchestration_workflow lacked.
- ACTION SET corrected from the seeded workflow (analyse_repo → lookup_symbols →
  load_runtime → assemble_bundle → run_loop → emit → complete): request_repo_analysis
  + lookup_code_symbols + complete_workflow EXIST; diagnose_load_runtime +
  diagnose_assemble_bundle drafted; `diagnose_run` (the engine wrapper — substantive,
  reads bundle_field/analysis_field/max_iterations/seed_scope_field/verdict_prompt_ref/
  seed_hypothesis_field) and `diagnose_emit` (thin) NOT built. `diagnose_route` NOT
  used (dropped). The model's per-iteration read-only data_requests belong in the
  engine's gather inside diagnose_run, NOT in diagnose_load_runtime (which runs once);
  the runDataRequests helper is reusable there.
- Verdict prompt: `diagnose_run` uses `verdict_prompt_ref: "diagnose-verdict-v1"` —
  ensure it resolves to PROMPT_diagnosis_verdict.md (the cite-or-abstain, falsify-first
  contract, DESIGN §2).
- Updated RUNBOOK §6C (use the pair + column move + build diagnose_run/diagnose_emit),
  §6E (loop is inside diagnose_run, no workflow loop-back), the status checklist, and
  the digest blocker. Updated the two DESIGN docs to current.

### 2026-06-24 — turn 10 (image era: v1.0.1074)
- The user uploaded the four drafted actions + verdict_wire + the verdict prompt,
  and confirmed THERE IS NO `diagnose_run`. So last turn's conclusion was wrong:
  the BUILT design is the workflow-driven `diagnose_route` loop, not a `diagnose_run`
  engine wrapper. Reverted the chassis-integration doc banner (it had been flipped
  to "diagnose_run wins"; corrected to "diagnose_route is built; no diagnose_run").
- VERIFIED `diagnose_route` compiles against the engine: `advance.go` provides
  `LoopState`, `Advance` → `AdvanceResult{Stop,Status,Conclusion,StoppedBy,
  Hypothesis,NextScope}`, `ParseVerdictValue`, `NewCallGraphFromValue`,
  `EncodeLoopState`/`DecodeLoopState`/`EncodeScope`/`EncodeTrail`; `loop.go` has
  `type Outcome` + `Outcome.String()` + `Scope` + `CallGraph`. All symbols route
  reads exist.
- REVIEW of the four actions vs guidelines — sound parts: complexity stays in the
  tested engine (Advance/DecideStep, 15 tests); thin workflow; verdict is its own
  observable execute_llm_prompt step; read-only throughout; verdict_wire Guard 2
  (IsReadOnlySQL) + fail-safes (unknown outcome / citation-less → UNVERIFIABLE).
  Findings: (a) the seeded workflow names the non-existent diagnose_run → rewrite to
  diagnose_route; (b) GAP — the model's data_requests + runtime_site are parsed and
  lint-checked but never executed in the loop: diagnose_route returns next_step/scope/
  hypothesis only (not data_requests/runtime_site) and loop-back goes to
  assemble_bundle (not load_runtime), so runDataRequests is never reached. Code
  re-scope works; runtime-following re-gather is an unwired enhancement; (c) the loop
  rests on the coordinator honouring a next_step key in an action result
  (getNextStepFromResult) — confirm in-tree; (d) load_runtime data_requests_field was
  input_data.data_requests (my turn-8 re-invocation edit) — for diagnose_route it
  should be route.data_requests (set so in the fixed workflow; dormant until the gap
  is wired).
- WROTE `NNN_fix_diagnose_agent_workflow.sql` (supersedes the move-only migration for
  the agent): rewrites diagnose-agent.default_config to the diagnose_route workflow
  (analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict → route →
  [loop|emit] → complete), with the verdict prompt INLINE (built programmatically from
  PROMPT_diagnosis_verdict.md + a `{{.bundle.bundle}}` injection, since the prompt has
  no bundle placeholder) and an `ai_service` block (user asked — model changeable in
  the verdict step + top-level). JSON validated; step graph resolves; route loops to
  assemble_bundle / stops at emit. Orchestrator moved to default_config + processing_mode/
  timeout. emit config reads route.* (matches the drafted diagnose_emit).
- ANSWERED "what is diagnose-verdict-v1": a prompt-registry ref in the STALE
  diagnose_run config; the built diagnose_route design has the verdict prompt inline in
  the execute_llm_prompt step, so it's not used.
- Updated RUNBOOK §6C/§6E + status checklist + this digest. Updated DESIGN docs.

### 2026-06-24 — turn 11 (image era: v1.0.1074)
- User uploaded the full engine (loop/step/advance/callgraph/gatherer/docselect/
  queryselect/sqlguard + verdict_wire + tests) and pushed back on the "gap".
- RE-CHECKED and RETRACTED the data_requests "gap": `Scope` and `AdvanceResult` carry
  no `DataRequests`, `DecideStep` never forwards them, and `queryselect.go` is explicit
  — DB evidence is a vetted `\d`-verified CATALOGUE selected by hypothesis, "the model
  never writes SQL". So re-scope = call-graph (code) + catalogue (DB) by design. The
  `runDataRequests` in `diagnose_load_runtime` is dormant (nothing feeds it) and beyond
  that boundary — decide drop vs keep as a bounded capability. Corrected §6C + the
  design doc (removed the overstated gap).
- REAL verify item found instead: `diagnose_route` writes `EncodeScope(NextScope)` (an
  OBJECT {symbols,...}) at `route.scope`, but `diagnose_assemble_bundle` reads it via
  `ExtractStringListHelper` (a LIST). Confirm `route.scope.symbols` or that the helper
  unwraps — else loop-back code re-scope won't flow. Not covered by the engine tests
  (they test Advance/DecideStep, not the actions).
- Also noted: the chassis loop gathers runtime once + re-assembles code each iteration;
  the per-hypothesis query catalogue the standalone runs isn't wired into the assemble
  action yet (later parity). Runtime-once is fine (completed run, static rows).
- DB state: `diagnose-agent` rewrite landed (has_route=t, run_loop gone, orch_null=t).
  `diagnose-orchestrator` move did NOT (has_wf=f) — wrote
  `NNN_restore_diagnose_orchestrator_workflow.sql` to re-seed its spawn→call→complete
  workflow explicitly (source column now null), with call_diagnoser timeout raised to
  1800 to match the agent loop. `getNextStepFromResult` confirmed (coordinator.go:1093)
  — the routing mechanism is real.
- Standing rule: `snapshot_agent('<type>','<reason>')` before any agent_definitions
  change; added to the fix + restore migrations. diagnose_ro confirmed SELECT-yes /
  DELETE-no on site_flows.

### 2026-06-24 — turn 12 (image era: v1.0.1074)
- User pushed back on "runDataRequests dormant". Re-checked `sqlguard.go`: chassis runs
  model SQL under a READ-ONLY TRANSACTION (the real guarantee), so model data_requests
  are SAFE in the chassis; the catalogue (queryselect) is the STANDALONE's mechanism
  (read-only role). So runDataRequests is the chassis DB-following channel and should be
  LIVE — dormant only from a 3-part wiring gap. Both my prior framings (gap; then
  boundary) were wrong.
- WIRED it: (1) regenerated NNN_fix_diagnose_agent_workflow.sql with
  `gather_step: load_runtime` so the loop returns to the SQL-runner (runs the prior
  verdict's data_requests under the read-only tx + re-gathers runtime, then
  assemble_bundle). (2) PATCH_wire_data_requests.md: `diagnose_route` forwards
  `verdict.DataRequests` → `route.data_requests` (it has the parsed verdict before
  Advance, so the engine not carrying DataRequests through Scope is no obstacle — code
  re-scope and data re-gather are separate channels) + a small `encodeDataRequests`
  helper. (3) `diagnose_load_runtime` InputSpec default → `route.data_requests`
  (was stale `input_data.data_requests` from the abandoned re-invocation design) +
  comment fix.
- Corrected §6C + the design doc (data_requests = chassis channel, safe under read-only
  tx, now wired) and this digest. Restated the still-open CODE re-scope verify item
  (route.scope object vs assemble_bundle's ExtractStringListHelper list).
- Orchestrator restore migration (turn 11) confirmed still pending; its current row
  shows default_config={} (full dump provided), so the explicit re-seed is needed.

### 2026-06-24 — turn 13 (image era: v1.0.1074)
- Migrations APPLIED (user's psql): diagnose-agent rewritten (snapshot captured,
  UPDATE 1) with gather_step=load_runtime; diagnose-orchestrator restored (snapshot,
  UPDATE 1). ReadSymbolBody passes in both contextkit and chassis trees.
- MERGED the data_requests patch into the two action files and provided them
  (gofmt-clean, parse-verified): diagnose_route_action.go (encodeDataRequests +
  forwards verdict.DataRequests → route.data_requests in the iterate branch + doc/
  logger updates); diagnose_load_runtime_action.go (InputSpec default →
  route.data_requests; corrected the stale re-invocation comments; documented the
  three-layer SELECT-only guarantee in the body comment).
- CONFIRMED SELECT-only at three layers against the real code: (1) prompt §7
  (single read-only SELECT/WITH … SELECT only); (2) IsReadOnlySQL filter applied
  TWICE — verdict_wire.toVerdict (parse, line 91) and runDataRequests (execution,
  line ~287); (3) read-only transaction BeginTx{ReadOnly:true} + statement_timeout
  (the real backstop). Documented in §6C + the load_runtime code.
- Remaining: build + register the four actions (Category 'diagnose') + deploy; then
  VERIFY the route.scope→assemble_bundle field shape (route.scope.symbols vs
  ExtractStringListHelper) before trusting CODE re-scope.

### 2026-06-24 — turn 14 (image era: v1.0.1074)
- `diagnose_route` now reads the verdict's `data_requests` from the RAW verdict wire
  (`verdict.result`) via `readOnlyDataRequestsFromWire`, keeping only read-only ones via
  `diagnose.IsReadOnlySQL`, rather than from a typed `Verdict` field — this keeps the
  action decoupled from the engine package (`data_requests` are model-supplied wire data,
  so reading + linting them at the route boundary is the natural layer). Added the
  `strings` import; gofmt-clean, parses.
- 3-layer SELECT-only guarantee with the wire read: prompt §7 → route-layer lint
  (`readOnlyDataRequestsFromWire`) → `runDataRequests` re-lint → read-only tx. Two lints +
  the tx, plus the prompt. §6C + the load_runtime comment describe the route-layer lint.

### 2026-06-24 — turn 15 (image era: v1.0.1076)
- analyser-adapter (the diagnose loop's step-1 dependency: analyse_repo →
  request_repo_analysis → system.adapter.analyser.requests → this adapter) was stuck
  `Pending`/`CreateContainerConfigError`. `state.waiting.message` =
  `secret "analyser-github-read" not found`.
- Root cause (from the uploaded Terraform): the deployment's env used
  `secretKeyRef{name: analyser-github-read, key: token}`, but Terraform creates no such
  secret — the read-only PAT lives in `personae-platform-secrets` under key
  `GITHUB_READ_TOKEN` (`var.github_read_token`, renamed from the old hyphenated
  `analyser-github-read` var). Three-way name tangle (TF var name ≠ k8s secret name ≠
  env var name); the deployment pointed at a secret/key that never existed.
- Fix: point the env at the real secret+key with `secretKeyRef` (name
  personae-platform-secrets, key GITHUB_READ_TOKEN) — NOT `envFrom`, which would inject
  EVERY platform-secret key (incl. the read-write GITHUB_TOKEN + DB passwords + JWT) into
  a read-only adapter. GITHUB_API_BASE is a plain env (defaults to api.github.com), never
  a secret, never the cause.
- VERIFIED: analyser-adapter `1/1 Running`; the GITHUB_READ_TOKEN secret key decodes to a
  real token. §6D request→adapter→index half is unblocked; the code_symbols count for
  gqls/agentchassis is the remaining §6D check (rises once a diagnosis/analyse runs).
- Tidy-ups noted (not blockers): correct the analyser-adapter.yaml config comment
  ("via envFrom / analyser-github-read Secret" → secretKeyRef to
  personae-platform-secrets/GITHUB_READ_TOKEN); ensure variables.tf declares
  github_read_token and tfvars sets it.

### 2026-06-24 — turn 16 (image era: v1.0.1076)
- code_symbols is genuinely empty (total 0, no repos) — confirmed not a query fault.
  Not a regression: the table is WRITTEN only by index_code_symbols (storage) and only
  READ by the loop's lookup_code_symbols. The diagnose agent does request_repo_analysis
  -> lookup_code_symbols (read), so diagnosis never populates it; nothing has indexed on
  prod, so empty is expected.
- The writer flow is a SEPARATE agent: `code-indexer` (seeded 2026-06-12), workflow
  request_repo_analysis -> await analyser -> index_code_symbols. To populate
  gqls/agentchassis: confirm the agent exists, then hand-trigger it (080c envelope,
  target_agent_type=code-indexer, input_data {owner:gqls, repo:agentchassis, ref:HEAD});
  watch adapter logs (analysed repo -> response sent) + indexer logs (index_code_symbols
  counts); re-run the count (UPSERT on uq_code_symbols_identity, safe to repeat). repo
  label = owner/repo, composed by index_code_symbols, so WHERE repo='gqls/agentchassis'
  is correct.
- Also confirmed this turn (user's psql/grep): the four diagnose actions ARE registered
  in registry.go (load_runtime/assemble_bundle/route/emit); diagnose-orchestrator
  has_wf=t/has_spawn=t. So the build went green after the wire-based route fix, and the
  §6C register/workflow side is done.
- Consequence for a first diagnosis: until code_symbols has chassis rows, lookup_symbols
  returns empty, so pass input_data.seed_scope on the trigger or assemble_bundle errors
  "no scope". Added the §6D note (prerequisite + how to populate + the data-vs-body caveat).

### 2026-06-24 — turn 17 (image era: v1.0.1076)
- Added the concrete code-indexer kcat trigger to §6D (envelope to
  system.agent.generic.requests with action=orchestrate, config.agent_type=code-indexer,
  input_data {owner,repo,ref,language}). Corrected copy-paste "classifier" labels +
  the workflow-step grep to code-indexer signals; kcat produce kept verbatim.
- Reframed the data_requests note in §6C (and neutralised turn-14 here): route reading
  data_requests from the verdict wire is a deliberate boundary choice (model SQL is wire
  data), not a workaround for an engine-copy gap — dropped the "behind/divergence/build
  failed" framing. No code change; the wire read + route-layer lint + load_runtime
  re-lint + read-only tx are unchanged. (If we later want the chassis pkg/diagnose in
  sync with contextkit so route can use the typed Verdict.DataRequests field, that's a
  small additive engine change + a one-line route edit — offered, not done.)

### 2026-06-26 — turn 18 (image era: v1.0.1076)
- code-indexer trigger RAN: analyser `analyse response sent` (complete) → code-indexer
  `request_analysis` awaited request `processed` → `index_symbols` → code_symbols = 436
  rows for gqls/agentchassis (sample paths cmd/.../main.go, internal/adapters/...). §6D
  index prerequisite SATISFIED; lookup_symbols now has a live index to seed from.
- Benign observations (recorded in §6D, not blockers): analyser logs idle
  `context deadline exceeded` fetch timeouts at error level on the requests topic AFTER
  the analyse succeeded (cosmetic log-level issue, topic is fine); awaited_requests
  target_agent_type='unknown' for request_analysis is expected (analyser is topic-addressed,
  not agent-addressed). Indexer orchestration snapshot caught mid-run at
  EXECUTING_STEP/index_symbols — worth a quick confirm it reached a terminal status.
- Remaining before a §6F diagnosis pass: VERIFY the route.scope→assemble_bundle field shape
  (route.scope.symbols vs ExtractStringListHelper) so loop-back code re-scope flows; then
  hand-trigger one read-only diagnosis (seed_scope optional now that the index is populated,
  but still useful to steer iter-1).

### 2026-06-26 — turn 19 (image era: v1.0.1076)
- Indexer orchestration confirmed terminal: COMPLETED/complete (the §6D mid-run snapshot
  has resolved). No diagnose orchestration yet (workflow_plan LIKE '%diagnose%' -> 0 rows),
  expected until §6F runs.
- Rewrote RUNBOOK §6F to the REAL hand-trigger: the same 080c generic-request kcat envelope
  as the §6D code-indexer trigger (system.agent.generic.requests, action=orchestrate), with
  config.agent_type='diagnose-orchestrator' (the entry point; it spawns the diagnose-agent
  and forwards the result). input_data populated with what we know: owner=gqls, repo=
  agentchassis, ref=HEAD (the repo we just indexed), symptom='index page completed but
  content is a stub', runtime_site='gamesdesign.co.uk'; SITE_ID/RUN_CORRELATION as vars to
  fill from SQL (empty tolerated -> NULL filters). Replaced the wrong old skeleton (kafka-
  produce + input_data {site,symptom}; 'site' isn't a diagnose field).
- Listed the lookup SQL in §6F step 1: (a) confirm diagnose-orchestrator/diagnose-agent +
  has_wf; (b) gamesdesign site_id (sites table primary, agent_error_log fallback) — site_id
  is the reliable runtime selector since site_work_items (the stub evidence) only filters by
  it; (c) optional recent gamesdesign correlation_id. Added a 'selector miss vs absence'
  caveat: if the runtime bundle is empty, check load_runtime's site_id/domain field configs
  against the keys sent before concluding no evidence.
- Kept step 3 write-nothing checks (orchestration_states/site_work_items/pages max(created_at)
  unchanged) + DoD. All 27 runbook bash blocks parse.
- Open before running §6F: the route.scope->assemble_bundle field-shape check (route.scope.
  symbols vs ExtractStringListHelper) still pending — a mismatch silently re-uses the seed
  scope on loop-back. After §6F: §6E (re-scope visible) then §6G (eval gate).

---

## DECISIONS (with rationale)

- **Backup before any agent_definitions change (standing rule, turn 11).** Use
  `SELECT snapshot_agent('<type>','<reason>')` (DB function, signatures
  `snapshot_agent(text)` and `snapshot_agent(text, text)`) before every UPDATE. The
  migrations include it.

- **One shared slicer in `analysis`, not a chassis-local one.** The convention must
  be identical everywhere a symbol body is rendered; duplicating it invites drift
  (the project has been bitten by stale copies). `analysis.ReadSymbolBody` is the
  single implementation; `cmd/assembler` and the action both call it.
- **Slice from the analyser's recorded spans, never re-parse.** The spans are
  already computed and shipped in `repo_analysis`; re-parsing in the action would
  duplicate the analyser and could diverge on edge cases.
- **Inclusive, 1-indexed `[StartLine,EndLine]`, doc comment excluded.** Verified
  byte-for-byte against `cmd/assembler`'s real output; the render template owns the
  trailing newline.
- **Decode the Output via JSON round-trip in the action.** `collected_data` holds a
  decoded map; `json.Marshal(map)` → `json.Unmarshal(&Output)` round-trips on the
  shared `json` tags (tested). Defensive string/[]byte cases are belt-and-braces.
- **Recover-from-ground-truth over guess-the-shape.** When the types file and
  `cmd/assembler` were absent, the field names came from `analyse.go`'s struct
  construction and the convention from the bundle's own output — not assumption.

---

## RISKS / OPEN QUESTIONS

- **`analyse.go` drift between copies.** Confirm the chassis `internal/analysis`
  `types.go` matches the contextkit one before relying on `symbolbody.go` there.
  The fields ReadSymbolBody touches are fundamental, but the walker has diverged,
  so the types could too.
- **`cmd/assembler` resolution vs `spanOf`.** The byte-identical diff covers a plain
  func. Confirm assembler's handling of method-name collisions and whole-file
  scopes matches before collapsing the two onto one slicer.
- **Chassis compile is unverified in the sandbox.** The action build (datahelpers,
  zap, ActionParams) must be confirmed in your env.
- **The eval gate is still the real test.** A green build and a resolving bundle are
  not evidence the loop reasons (reproduces reversals, abstains when unsettled).
