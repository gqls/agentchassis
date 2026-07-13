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

### 2026-06-26 — turn 20 (image era: v1.0.1077)
- Where we are: §6F, ready to fire the hand-trigger. Everything upstream confirmed green
  by the user's paste.
- §6C CLOSED by live evidence: diagnose-agent has_wf=t/has_route=t/run_loop GONE (f)/orch_null=t;
  diagnose-orchestrator has_wf=t/has_spawn=t/has_route=f (correct — spawn+call, not route).
  The next_step override mechanism is live in coordinator.go (913-918 checks the action
  result for an override; 1093/1105 getNextStepFromResult reads next_step, 1110 legacy
  next_step_override). So diagnose_route's loop control works as designed.
- The diagnose-agent default_config (uploaded) is the REAL route-shape workflow:
  analyse_repo -> lookup_symbols -> load_runtime -> assemble_bundle -> verdict -> route ->
  (load_runtime | emit) -> complete. verdict step carries the full cite-or-abstain prompt
  inline with {{.bundle.bundle}}, ai_service sonnet-4-6, temp 0.0, json. Matches the migration.
- load_runtime source (uploaded) CONFIRMS: domain_field default = input_data.runtime_site
  (so runtime_site populates the agent_error_log domain filter — resolves the §6F hedge);
  site_work_items only when site_id set; data_requests read from route.data_requests, run
  under BeginTx ReadOnly + SET LOCAL statement_timeout='15s', re-linted via diagnose.IsReadOnlySQL.
  (File still carries PRE-MERGE grep-for-dup-helper notes for nullUUID/nullText/errSuffix/
  formatRowsText — harmless now that it built green at v1.0.1077; dedupe only if a collision
  ever surfaces.)
- ROUTE.SCOPE item RESOLVED (was the long-flagged open verify): CONFIRMED a real mismatch.
  EncodeScope = json.Marshal(Scope); Scope has NO json tags, so route.scope is an OBJECT with
  CAPITALISED keys {"Symbols":[...],...}. assemble_bundle reads loop_scope_field "route.scope"
  via ExtractStringListHelper (a list reader) -> empty on loop-back -> falls back to seed/
  code_results -> NO re-scope. First pass unaffected (iter-1 uses code_results); re-scope
  (iter 2+) was silently inert -> would defeat §6E/§6G. FIX (config-only, no rebuild):
  migration NNN_fix_assemble_bundle_loop_scope_field.sql sets loop_scope_field =
  "route.scope.Symbols" (capital S; ExtractNestedField does 3 levels — precedent
  input_data.section_plan.sections_ready, site_specs.identity.team). Apply before §6E/§6G.
- Minor (noted, non-blocking): the stale orchestrator_workflow/task_workflow columns still
  hold the OLD run_loop/diagnose_run shape, but they are UNUSED (the loader reads
  default_config). Don't be alarmed seeing diagnose_run there; optional tidy is to null them.
- orchestration_states diagnose query still 0 rows (no diagnosis triggered yet) — expected.

### 2026-06-26 — turn 21 (image era: v1.0.1077)
- §6F FIRED (correlation 03c4604a-4aa3-4b47-8397-212a13fd7e10, header orch c43e3295). The
  read-only DoD half is met: latest orchestration COMPLETED/complete/no-error; site_work_items
  last write 2026-06-25, pages 2026-06-22 — both pre-date the run, so nothing written to the
  build tables (and the newest orchestration touching neither corroborates a read-only pass).
- PENDING to close §6F: (a) confirm the run by correlation (orchestrator + spawned diagnose-
  agent both COMPLETED) since the LIMIT 1 row (8b5744fc) wasn't matched to the correlation yet;
  (b) read the verdict outcome from diagnose-agent logs (grep the correlation for iterat/
  REFUTED/CONFIRMED/UNVERIFIABLE/emit). The full diagnosis returns on
  system.agent.generic.responses; logs are the reliable read.
- CAVEAT recorded in §6F: this run predates NNN_fix_assemble_bundle_loop_scope_field.sql, so
  it used the old loop_scope_field=route.scope. If it looped, re-scope re-used the seed scope.
  Fine for a first-pass smoke (one verdict over one bundle = the bar); NOT evidence about §6E.
  Apply the migration before §6E/§6G.

### 2026-06-26 — turn 22 (image era: v1.0.1077)
- §6F Run 1 FAILED at assemble_bundle (both orchestrator c43e3295 + spawned agent d2f85ef5):
  "no scope (tried route.scope, input_data.seed_scope, then code_results)". READ-ONLY held
  (site_work_items 06-25 / pages 06-22 unchanged). The failure is ITERATION-1 SEEDING, not the
  loop-back bug: route.scope empty (route not run) + seed_scope empty (none passed) are
  expected, but code_lookup.code_results (the iter-1 seed from lookup_code_symbols) was ALSO
  empty. lookup RAN (reached the next step) but returned no usable {path,symbol}. So
  loop_scope_field migration is NOT the fix for this (that's iter 2+).
- Root-cause candidates for empty code_lookup.code_results (need chassis-side checks I can't
  run): (B) OUTPUT-KEY mismatch — assemble_bundle reads code_lookup.code_results;
  lookup_code_symbols may return results under a different key or as a bare list -> grep its
  return in code_symbols_actions.go (deterministic, most likely); (C) EMBEDDINGS null at index
  time -> vector search empty + trigram (prose symptom vs Go identifiers) finds little -> check
  count(embedding); (A) REPO filter mismatch repo_analysis.repo vs stored 'gqls/agentchassis'.
  scopeFromCodeResults (my assemble_bundle) expects []{path,symbol} and skips anything else, so
  a shape/key mismatch silently yields empty.
- Unblock path given: pass input_data.seed_scope (a list of "path:Symbol") so assemble_bundle
  uses it instead of code_results — tests verdict/route/emit now. For §6G eval we want the
  lookup seeding from the symptom, so the lookup fix is the real item.
- gamesdesign ids: sites.id = e33263f4-74f8-494f-b191-546845dbbddf (the sites query works).
  agent_error_log has ~23 site_ids for the domain (per-build instances) -> runtime_site=domain
  is the better runtime selector than a single site_id here.
- Still to apply (separate, before §6E re-scope): NNN_fix_assemble_bundle_loop_scope_field.sql.
- Answer to "re-run?": not unchanged — it fails identically. Fix the lookup seed (or pass
  seed_scope) first.

### 2026-06-26 — turn 23 (image era: v1.0.1077)
- ROOT CAUSE of the §6F "no scope" CONFIRMED from code_symbols_actions.go (uploaded):
  repo-LABEL asymmetry. index_code_symbols COMPOSES owner/repo (repo_analysis.owner + "/" +
  repo_analysis.repo = gqls/agentchassis, lines 146-154); lookup_code_symbols does NOT compose
  (line 59) and the diagnose workflow set lookup_symbols.repo_field=repo_analysis.repo (bare
  "agentchassis"). vectorSearchCodeSymbols filters WHERE repo=$1, so lookup queried
  WHERE repo='agentchassis' vs rows stored 'gqls/agentchassis' -> 0 hits -> empty code_results
  -> assemble_bundle no scope. Ruled OUT: (B) output-key (key IS code_results, elements
  {path,symbol}, lines 111/99-100) and (C) embeddings (436/436 have embeddings).
- Run 2 (960ec6bf) ran with NO change (no seed_scope, no fix) so almost certainly failed the
  same way. The LIMIT-1 COMPLETED row (b931e1c6) is a RED HERRING — in Run 1 the LIMIT-1 was
  also COMPLETED (8b5744fc) while the actual diagnosis (c43e3295/d2f85ef5) FAILED. Must confirm
  by correlation, not LIMIT 1. (Lesson recorded: never read run success off ORDER BY created_at
  LIMIT 1 — always filter by correlation_id; the orchestrator spawns a child, and unrelated
  orchestrations interleave.)
- FIXES authored:
  * NNN_fix_lookup_repo_label_workaround.sql — config-only, no rebuild: drop lookup_symbols
    repo_field, set literal config.repo='gqls/agentchassis' (resolveRAGConfigField honours
    config.repo override). Repo-SPECIFIC + temporary; unblocks the full loop/eval NOW. Includes
    a REVERT block.
  * PATCH_code_symbols_shared_repo_label.md — structural (rebuild): one resolveCodeRepoLabel
    helper used by BOTH index and lookup (so they can't diverge), then drop the bare
    repo_field/literal so lookup composes owner/repo for ANY repo. Ship -> run the workaround
    REVERT -> re-trigger.
- Still pending (separate): NNN_fix_assemble_bundle_loop_scope_field.sql (route.scope.Symbols,
  the loop-back re-scope). Order to a green eval: workaround (or shipped composing-lookup) +
  loop_scope migration, then re-trigger §6F, then §6E re-scope, then §6G eval.

### 2026-06-26 — turn 24 (image era: v1.0.1077)
- Run 2 (960ec6bf) confirmed by correlation: BOTH orchestrations FAILED at assemble_bundle
  "no scope (tried route.scope.Symbols, input_data.seed_scope, then code_results)". Two facts
  from that: (1) the error says route.scope.Symbols, so NNN_fix_assemble_bundle_loop_scope_field
  .sql was already applied (loop-back fix live); (2) still the repo-label/empty-code_results
  cause, now fixed by the workaround.
- Workaround NNN_fix_lookup_repo_label_workaround.sql APPLIED + verified: snapshot captured,
  UPDATE 1, lookup_config = {"repo":"gqls/agentchassis","top_k":12,"query_field":
  "input_data.symptom"} (repo_field dropped). So the lookup now filters the correct label;
  next §6F trigger should clear seeding.
- STRUCTURAL patch APPLIED into code_symbols_actions.go (writable copy in outputs): added one
  resolveCodeRepoLabel(config, collected) (config.repo/repo_field -> compose owner/repo from
  repo_analysis -> input_data.repo) and routed BOTH LookupCodeSymbolsAction and
  IndexCodeSymbolsAction through it; index keeps its repo-not-found guard. Only the inline
  repo-resolution was changed (query_field still uses resolveRAGConfigField). gofmt-clean,
  imports all still used. Original upload was gofmt-clean; the patch is minimal.
- Deploy sequence: build+push code_symbols_actions.go -> run the workaround REVERT block (drops
  the literal config.repo + bare repo_field) so the rebuilt lookup composes owner/repo for ANY
  repo. Until then the literal keeps the eval unblocked.
- Next: re-trigger §6F (expect it to clear assemble_bundle now), confirm by CORRELATION (not
  LIMIT 1) that orchestrator+agent reach a verdict, read the outcome from diagnose-agent logs;
  then §6E re-scope, then §6G eval.

### 2026-06-26 — turn 25 (image era: v1.0.1077)
- Run 3: seeding cleared after deploy (past "no scope"); NEW failure at assemble_bundle:
  "repo root not found at repo_analysis.root". Architectural finding: analyser returns SPANS
  not bodies; code_symbols.content = composeSymbolContent (kind+symbol+signature+doc+path),
  retrieval text NOT bodies (search SELECTs return line_start/line_end/signature only); the
  diagnose-agent pod has no checkout (analyser clones in its own pod, returns metadata). So the
  bundle's in-scope code cannot be filled as wired.
- Options weighed (two turns): (1) bodies in DB column -> whole-repo Kafka payload at index
  time, rejected; (2) stateful analyser serving slices -> cleanest on responsibility but
  couples every iteration to a live analyser (and the loop re-enters at load_runtime, so only
  the body-fetch would be per-iteration; LLM verdict cost dwarfs the round-trip anyway, the
  real cost is the analyser gaining a checkout cache w/ TTL); (3) diagnose-agent re-checks-out
  on demand. User corrected me that stateful temp pods already exist (loops hold pods until the
  orchestration completes), weakening option 2's stateful objection — but chose OPTION 3 to
  keep the agent self-contained and avoid per-iteration coupling. Bundles are small (~30-100kB;
  assemble_bundle caps code at maxBodyChars=60000) so data movement was never the deciding cost.
- DONE: patched spawn_actions.go (uploaded) — credential injection scoped to the diagnoser.
  Added isRepoCloningAgent (mirrors isStorageEnabledAgent; list {"diagnose-agent"}) and, in
  spawnAgentKubernetesJobFromDefinition, a GITHUB_READ_TOKEN env via secretKeyRef
  {personae-platform-secrets, GITHUB_READ_TOKEN} gated by isRepoCloningAgent. secretKeyRef
  (NOT a passthrough from the spawner's os.Getenv) so the spawning chassis pod never holds the
  token and ONLY diagnose-agent pods receive it. gofmt-clean. No RBAC change: that Secret is
  already referenced for CLIENTS_DB_PASSWORD/ANTHROPIC etc., so the spawned pod's SA can read
  it. Reused the same secret/key as the analyser-adapter fix (read-only single-repo PAT).
- REMAINING for option 3 (decisions needed before I write the clone/body code):
  (a) clone mechanism: go-git (pure Go, no chassis image change) vs shell git (needs git in the
      image). (b) clone+repoRoot wiring — prefer the SELF-CONTAINED variant (clone once + run
      analysis in-process if the analysis package exposes an entrypoint, dropping the cross-pod
      analyse call); else keep the analyser call for the graph + clone only for bodies (2 clones
      /run). (c) pin commit_sha = the index's commit so retrieved symbols and sliced bodies agree.
- Open questions for the user: go-git vs git-binary; does the analysis package expose an
  in-process analyse entrypoint; confirm the chassis image will carry whatever git mechanism we pick.

### 2026-06-26 — turn 26 (image era: v1.0.1077)
- Recorded the body-source architecture thread (it was only in turn logs, not the DECISIONS
  record) and folded in three choices the user made this turn: go-git over a shelled git binary;
  the diagnose-agent FULLY self-contained (clone once + analyse in-process, dropping the cross-pod
  analyser call); and pinning the index's commit_sha for the checkout. See DECISIONS below.
- Dependency flagged for "fully self-contained": it needs the `analysis` package to expose an
  in-process analyse entrypoint (AnalyseRepo(root)->Output). If only the adapter wrapper is
  exported, that entrypoint must be exported first (chassis change) or we fall back to the
  2-clone variant. Also noted: lookup_symbols still reads the code_symbols index for retrieval
  seeding (built by the code-indexer) — "self-contained" = no runtime analyser-adapter call, not
  independence from the shared index.
- Ready to implement next: a clone+in-process-analyse action (go-git) replacing analyse_repo in
  the diagnose workflow, repoRoot = the local checkout, ReadSymbolBody unchanged; then the
  assemble_bundle repoRoot error stops firing. Credential injection already done (turn 25).

### 2026-06-28 — turn 27 (image era: v1.0.1077)
- Uploads this turn settled the two open questions from turn 26. (1) The in-process entrypoint
  ALREADY EXISTS and is public: `analysis.Analyse(root) (Output, error)` in analyse.go — already
  called by the analyser adapter's analyse_action.go. Nothing new to export; the "fully
  self-contained" dependency is discharged. (2) `github_source.go` showed the analyser fetches a
  repo via a TARBALL (`GET /repos/{o}/{r}/tarball/{ref}` → temp dir → SHA from the archive folder),
  not git — so go-git was the wrong call. SUPERSEDED the turn-26 go-git decision: reuse the existing
  `FetchToDir` instead (git stays out of the chassis entirely). User confirmed both: reuse FetchToDir;
  lift the fetcher into a neutral package.
- Built the self-contained in-process fetch+analyse action that replaces the cross-pod analyse step.
  All deliverables gofmt-clean (parse + format verified in sandbox; full build is user-side).
- NEW FILES (in /mnt/user-data/outputs, mirrored to their chassis paths):
  - `internal/reposource/github_source.go` — the fetcher LIFTED verbatim from
    internal/adapters/analyser/github_source.go (path-traversal guard in extractTarGz intact),
    package analyser→reposource, FILE path updated, framing neutralised + a lift note, and a
    `Fetcher` interface seam added.
  - `platform/orchestration/actions/analyse_repo_local_action.go` — `analyse_repo_local`
    (package actions). Resolves owner/repo/ref via resolveRAGConfigField (same keys as
    request_repo_analysis); pins the dominant code_symbols commit (read-only SELECT, best-effort,
    `pin_to_index_commit` flag); reads GITHUB_READ_TOKEN; `reposource.NewGitHubSource`→`FetchToDir`
    (NO os.RemoveAll — checkout persists for the pod); `analysis.Analyse(dir)`→Output; returns the
    Output fields at TOP LEVEL + commit_sha/owner/repo/ref as `repo_analysis`. Helpers
    pinToIndexCommit (local bool coercion — PRE-MERGE grep note for datahelpers.GetBoolField),
    indexCommitSHA, outputToMap.
  - `NNN_swap_analyse_repo_to_local.sql` — snapshot_agent then nested jsonb_set on
    default_config→workflow→steps→analyse_repo: action request_repo_analysis→analyse_repo_local,
    config += pin_to_index_commit:true, description reworded; next_step(lookup_symbols) +
    output_field(repo_analysis) preserved. Verify query + commented REVERT included.
  - `PATCH_lift_fetcher_and_register.md` — the 3 chassis edits: git rm the old github_source.go;
    adapter.go import + the one NewGitHubSource→reposource.NewGitHubSource change (analyse_action.go
    UNCHANGED — its local repoSource interface is still satisfied structurally); registry.go new
    `analyse_repo_local` entry (Category "code") + flow-comment update. Plus build/deploy→migration order.
- Confirmed against uploads before writing: registry.go registration pattern + that request_repo_analysis
  is Category "code" (line ~1057); adapter.go's single NewGitHubSource construction site; code_symbols
  columns (repo composed "owner/repo", commit_sha nullable) + resolveRAGConfigField in package actions;
  the diagnose-agent default_config analyse_repo step shape (steps is an OBJECT keyed by name, at
  default_config→workflow→steps).
- Next: user applies the patch + 2 new files → `go build ./...` → push/Actions/Backblaze (new image
  tag) → run the swap migration → re-trigger §6F (confirm by correlation_id). Expect assemble_bundle to
  clear and produce the first real bundle + verdict.

### 2026-06-29 — turn 28 (image era: v1.0.1077; diagnose_route fix pending redeploy)
- MILESTONE: with the swap migration applied (analyse_repo→analyse_repo_local), a hand-triggered
  §6F run went END-TO-END read-only: analyse_repo_local→lookup_symbols→load_runtime→assemble_bundle
  →verdict→route→emit→complete. repo_analysis carried a real {root,files,commit_sha}; assemble_bundle
  cleared the repo-root check; a verdict was produced. last_work_item/last_page untouched (read-only
  holds). §6F ticked. (analyse_repo_local ran in the spawned diagnose-agent Job 453ee814; its output
  came back via call_diagnoser.response, so the trail is readable without the Job pod logs.)
- BUG found (§6E): the run stopped at ITERATION 1 with UNVERIFIABLE/scope-not-narrowing. Root cause —
  diagnose_route never SEEDED the LoopState on the first iteration: it rehydrated from an ABSENT
  diagnose_state, leaving the zero value (Hypothesis="", Scope={}, PrevScopeSize=0). The narrowing
  guard is next.size() > prevSize+2; the model proposed a 3-item next_scope, so 3 > 0+2 tripped and it
  stopped before re-scoping. The trail's empty Hypothesis + null Scope.Symbols confirm the zero-value
  state. The VERDICT itself was sound: it abstained with two real Tier-2 citations (site_work_items
  needs_human_review on the index 'system-stats' section; agent_error_log content-regression "empty
  content" block) and asked precisely for the index page's page_components/pages content.
- FIX (code-only, NO migration): diagnose_route now seeds via diagnose.InitLoopState on the first
  iteration — hypothesis = the symptom (seed_hypothesis_field, default input_data.symptom), scope =
  the SAME seed assemble_bundle uses (seed_scope_field → code_results_field), so PrevScopeSize =
  seed.size()+1. REUSES InitLoopState + scopeFromCodeResults (no new logic, no duplicate helper). The
  three new optional config keys default to the existing field names, so the workflow needs no change.
  gofmt-clean.
- FLAGGED for the re-run (NOT fixed — re-run first, don't conflate): the model put TABLE names in
  next_scope (the code-symbol channel → Scope.Symbols, sliced as bodies). The index-page CONTENT it
  wants is a DB read → belongs in the verdict's data_requests (run read-only by load_runtime on
  loop-back) and/or Scope.Tables. If iter 2+ keeps asking for the same rows without receiving them,
  that channelling (likely a verdict-prompt tweak) is the next item.
- NEXT: rebuild+redeploy the chassis (diagnose_route fix) → re-trigger §6F (NEW correlation_id) →
  confirm §6E by reading the diagnose-agent row's route.iteration (≥ 2) BY correlation_id, not LIMIT 1.
- DELIVERABLE: platform/orchestration/actions/diagnose_route_action.go (patched, gofmt-clean).

### 2026-06-29 — turn 29 (seeding fix live; state-threading fix pending migration)
- BIG RESULT: with the seeding fix deployed, a §6F run (correlation 710d3b01, diagnose-agent orch
  8d488e01) ran the loop END-TO-END and CONFIRMED. ProcessingHistory shows the loop physically
  cycled FIVE times (analyse_repo_local→lookup_symbols→ then load_runtime→assemble_bundle→verdict→
  route ×5 →emit→complete). §6E loop-back CONFIRMED; the iteration-1 stop is gone. Read-only held.
- STATE-THREADING BUG found in the SAME run (structural). The final evidence_trail has ONE entry at
  Iteration 1 despite 5 passes, and collected_data_keys shows a "route" key but NO top-level
  "diagnose_state". diagnose_route reads its prior LoopState from state_field="diagnose_state", but
  its result lands under output_field "route" → the state is at route.diagnose_state. So it RE-SEEDS
  every iteration (my turn-28 seeding else-branch runs each pass, which is why PrevScopeSize=13 kept
  the narrowing guard from tripping and the loop kept going). Silent breakage: (1) max_iterations
  NEVER enforced (Iteration resets to 1 each pass — it stopped only because the model confirmed on
  pass 5); (2) evidence_trail truncated to the final iteration (the REFUTE/UNVERIFIABLE journey lost);
  (3) cross-iteration guards (SeenCitations/HypHistory/PrevScopeSize) reset each pass. Re-scope still
  worked because that flows through route.scope.Symbols (a separate, correctly-threaded field).
- FIX: state_field → route.diagnose_state (consistent with diagnose_emit reading route.status etc.).
  Operative LIVE fix is the migration (NNN_fix_diagnose_route_state_threading.sql) — NO rebuild, since
  the workflow sets state_field explicitly. Code DEFAULT also corrected (both the Defaults map and the
  GetStringField fallback) for the next build. Verified route's output persists across loop-back (the
  working re-scope via route.scope.Symbols already proves it), so route.diagnose_state will thread.
- §6G NOT evaluable yet, and this run does NOT pass it. §6G wants REFUTE→…→CONFIRM down to the
  resolveResultSpec coordinator cause. This run CONFIRMED the SYMPTOM hypothesis itself, citing a
  cta_url section block (index 'system-stats' needs_human_review) + a content-regression on a DIFFERENT
  page (tool-drop-rate-simulator) — tangential to the result-extraction cause, and exactly the move the
  verdict prompt's worked example says to REFUTE and follow upstream. Whether earlier passes did that is
  UNKNOWN (trail truncated). So: land the threading fix → re-run → read the FULL per-iteration trail →
  THEN judge premature-confirm vs genuine. Reserve judgement; keep the next_scope-vs-data_requests
  channelling watch (tables in next_scope) separate.
- NEXT: apply NNN_fix_diagnose_route_state_threading.sql (live, no rebuild) → re-trigger §6F (new
  correlation_id) → confirm trail_len == iteration and a non-converging run stops at iteration-cap →
  then read the full trail for §6G.
- DELIVERABLES: NNN_fix_diagnose_route_state_threading.sql; platform/orchestration/actions/diagnose_route_action.go (state_field default corrected, gofmt-clean).

### 2026-06-29 — turn 30 (state threading VERIFIED live; failsafe added; §6G now the focus)
- THREADING FIX VERIFIED. state_field = route.diagnose_state is live (confirmed in agent_definitions).
  Re-run correlation c72156d5 (diagnose-agent orch b25db7de; orchestrator orch f6ecada6) now threads
  the LoopState: route.diagnose_state has iteration:3, trail:[×3], seen_citations:{×4}, hyp_history:[×2],
  prev_scope_size:2. Emitted evidence_trail has 3 entries (trail_len == iteration), and the loop stopped
  at stopped_by=evidence-not-growing — a GUARD, not luck and not unbounded. §6E.1 satisfied: the guards +
  max_iterations cap are armed.
- FAILSAFE (the "just in case"). Established the loop is bounded by layers: 4 convergence guards +
  max_iterations:5 (loop-level, threading-dependent) AND timeout_seconds:1800 + fuel_budget:1000
  (engine-level, INDEPENDENT of the loop's bookkeeping — bound a runaway even if the loop guards were
  disarmed). A redundant counter adds nothing. Instead added a state-threading SELF-CHECK in diagnose_route
  (gofmt-clean): if it is about to re-seed but route.diagnose_state already exists (route has run before),
  it ABORTS loudly — next_step=emit, status UNVERIFIABLE, stopped_by="state-threading-error" — rather than
  silently re-seeding into a runaway. Catches the exact bug class we just fixed. Ships on the NEXT rebuild
  (bundled with the state_field code-default correction); the live system is already fixed via migration.
- §6G NOT passed — and now the full trail shows WHY (the threading fix gave us the per-iteration trail).
  The loop abstained UNVERIFIABLE×3 trying to verify "is the index content a stub" by querying section
  content: iter 1 re-scoped to save_page_sections/plan_sections (good — followed evidence to code);
  iter 2 issued a data_request against page_sections, which DOES NOT EXIST (SQLSTATE 42P01), correctly
  noted; iter 3 finally named real tables (page_components, pages) but evidence-not-growing fired (its
  citations were all previously-seen rows) and stopped the loop ONE iteration before those good queries
  would have run in load_runtime. It never REFUTED, so it never re-pointed upstream toward the build/
  coordinator path — it never reached resolveResultSpec.
- §6G WORK ITEMS (pick one, change one thing, re-run — do NOT conflate):
  1. evidence-not-growing is too eager w.r.t. in-flight data_requests: a verdict issuing a NEW unseen
     data_request is progress, but if its citations are all old the guard stops one iteration before the
     answer arrives (truncated this run). Candidate: count a new data_request as forward progress (thread
     a request-hash into guard memory). Structural — lives in guardAfter (loop.go); affects standalone Run.
  2. The model can't find the content table from the bundle schema (guessed page_sections; real candidates
     page_components/content_items). Surface content tables in the bundle schema, or seed them.
  3. The loop stays on the symptom hypothesis (abstains, never refutes → no upstream pivot). The raw-symptom
     seed may keep it in "verify the symptom" mode; a verdict-prompt nudge or mechanistic seed could help.
- NEXT: decide which §6G item to tackle (likely #1, the eager evidence-not-growing, since it cut off a
  promising lead); the failsafe + state_field code default ship on the next rebuild (no migration pending).
- DELIVERABLES: platform/orchestration/actions/diagnose_route_action.go (failsafe + state_field default, gofmt-clean).

### 2026-06-30 — turn 31 (§6G #1 fixed in engine: data_request counts as progress)
- Rebuild is live (pod agent-diagnose-agent-ebc7820c, chassis replicaset 6877555b4f). Run 0b76d9bc
  (diagnose-agent orch 01509525, orchestrator 1b10e5ad) threaded cleanly (3 iters, trail 3) and
  again stopped at evidence-not-growing — a clean SECOND reproduction of §6G work-item #1, sharper
  than before: the model CONVERGED on the query (iter 2 data_requests hit work_items 42P01 + slug
  42703 = wrong names; iter 3 formed a correct pages WHERE page_type='index' + page_components
  query) but evidence-not-growing fired on iter 3's stale citations and stopped ONE iteration
  before that good query ran in load_runtime. The good query never executed; never reached
  resolveResultSpec.
- FIX (engine, package diagnose, contextkit module): guardAfter now tracks issued read-only
  data_requests in a SeenRequests set that threads exactly like SeenCitations (added to LoopState,
  StepInput, StepDecision; seeded in InitLoopState and Run; threaded in Advance). evidence-not-
  growing and hypothesis-thrash now yield when the verdict issues a NEW (unseen) data_request — its
  answer arrives in the next gather, so a verdict that fixes its query is progress, not spinning. A
  RE-issue of the same query does NOT count (still trips); the iteration cap bounds the worst case.
  Files: advance.go, step.go, loop.go (gofmt-clean). Two regression tests: loop_datarequest_test.go
  (TestNewDataRequestDefersEvidenceNotGrowing — stale citation + new request keeps the loop alive to
  a later confirm; TestReissuedDataRequestStillTripsEvidenceNotGrowing — same query twice still
  trips). Existing guard tests verified safe (none attach data_requests, so newRequest stays false
  and they fire as before). NO migration, NO workflow/config change.
- DEPLOY: contextkit ENGINE change → rebuild contextkit + the chassis that imports diagnose; run
  `go test ./...` on the diagnose package first. Could NOT compile/test in sandbox (contextkit/
  internal/analysis dep absent) — gofmt parse-clean only.
- WHY #1 over a redundant something: it's the demonstrated blocker (twice), it's structural (the
  guard read SeenCitations but not the verdict's data_requests), and it routes around #2 (the
  model's table-name guessing) via the data_requests path — even with junk in next_scope, the
  data_requests run in load_runtime and feed the next bundle.
- STILL OPEN: #2 (model guesses content table — surface pages/page_components/content_items +
  columns in the bundle schema, or seed them); #3 (stays on symptom hypothesis, abstains not
  refutes — revisit after #1 lands). SEPARATE: trigger dropped site_id + correlation_id (empty in
  0b76d9bc envelope; earlier runs had e33263f4) → load_runtime not site-pinned; fix the envelope.
- NEXT: deploy #1 → re-run §6F (new correlation id) → read the full trail; expect the loop to run
  the pages/page_components query (iter 4) and reach a verdict instead of stopping early.
- DELIVERABLES: diagnose/advance.go, diagnose/step.go, diagnose/loop.go, diagnose/loop_datarequest_test.go.

---

### 2026-06-30 — run 0cd6e6a7 (#1 DEPLOYED but INERT; empty bundle schema is the real blocker)
- Deploy live (pod agent-diagnose-agent-7e0eb05f, chassis replicaset 7b76ff8dc4). Trigger now sends
  site_id=e33263f4 + a real correlation — the empty-site_id trigger gap is closed.
- Run correlation 0cd6e6a7 / diagnose-agent orch 0d1d4664 / orchestrator orch 115346dd. Loop ran 5
  iterations (vs 3 before) and stopped at evidence-not-growing on iter 5 (== the cap).
- FINDING 1 — the #1 fix is currently INERT. Every trail Verdict has DataRequests=null, yet the
  model DID issue data_requests (iters 2 & 5 cite `data_request error: relation "page_sections" does
  not exist`). So load_runtime RAN them (route.data_requests was populated by
  readOnlyDataRequestsFromWire), but the DOMAIN Verdict that reaches guardAfter has DataRequests=nil.
  Since the deployed advance.go ParseVerdictValue re-marshals -> ParseVerdict -> toVerdict, the only
  seam that can drop them is toVerdict: the chassis's pkg/diagnose/verdict_wire.go is an OLDER copy
  missing the DataRequests field/mapping (the /mnt/project copy maps them; two-copy drift). My engine
  fix is correct; it is simply never fed. FIX: sync verdict_wire.go into the chassis (stdlib-only, no
  import rewrite) — delivered as diagnose/verdict_wire.go. Confirm signal: the diagnose_route log line
  shows data_requests=N (>0) at route while the trail shows null.
- FINDING 2 (DOMINANT) — the bundle carries NO schema. assemble_bundle builds only ## Hypothesis,
  ## In-scope code, ## Runtime / DB evidence — there is no ## Schema section. Iter 5 needed_evidence
  says it plainly: "The schema section of the bundle contains no table definitions." So the model
  guesses the content table (page_sections -> 42P01) every iteration and can never form the right
  pages/page_components query. Without schema, even a fully-wired #1 cannot reach the content. This
  is work-item #2 and it gates §6G.
- The loop reaching iter 5 is a CONFOUND, not #1 working: the model re-quoted the same site_work_items
  rows with vs without a leading "[timestamp]" prefix (iter 1 "build/page_rerender ... tools-index"
  vs iter 3 "[2026-06-24T...] build/page_rerender ... tools-index"), so the citation keys differed and
  looked like new evidence — keeping newEvidence true past iter 3.
- OPEN QUESTION for #2: load_runtime's data_requests run against params.DB (clients_db, the
  orchestration DB) — page_sections 42P01 came from there. Need to confirm whether page CONTENT
  (pages/page_components) lives in clients_db or a per-site content DB; if the latter, the loop can't
  read content via data_requests at all (deeper issue). Asked the user for an information_schema
  listing before writing any schema-introspection SQL.
- NEXT: (a) user drops diagnose/verdict_wire.go into pkg/diagnose + rebuild (completes #1); (b) decide
  #2 placement — add a read-only information_schema introspection to load_runtime (it already has
  params.DB + QueryContext, mirroring LoadSiteForRebuildAction) and render a ## Schema section in the
  bundle, once we know the real content tables and which DB they live in.
- DELIVERABLE this turn: diagnose/verdict_wire.go (sync; stdlib-only).

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

- **Symbol bodies come from a git checkout the diagnose-agent makes itself (Option 3,
  turns 25-26).** The bundle needs in-scope source, but the analyser returns SPANS only,
  `code_symbols.content` is retrieval text (`composeSymbolContent` = kind+symbol+signature+doc+
  path), and the diagnose-agent pod has no checkout (the analyser clones in its own pod).
  Options weighed: (1) a `code_symbols` body column — REJECTED: forces a whole-repo Kafka
  payload analyser→indexer at index time and parks repo source in Postgres; (2) a stateful
  analyser serving slices — cleanest on responsibility (git in one place) but couples every
  iteration to a live analyser holding the checkout (per-iteration data/$ is negligible next to
  the LLM verdict; the real cost is a checkout cache w/ TTL/eviction/concurrency); (3) the
  diagnose-agent re-checks-out on demand — CHOSEN: git stays source of truth, nothing whole-repo
  in the DB, no per-iteration analyser coupling, checkout lives in the spawned agent's ephemeral
  pod (the loop already holds it open until the orchestration completes). Bundles are small
  (~30-100kB; `maxBodyChars`=60000), so data movement was never the deciding cost.
- **Fetch via the shared tarball fetcher, NOT go-git or a git binary (turn 27 — supersedes the
  turn-26 go-git choice).** The uploaded `github_source.go` showed the analyser already fetches a
  repo with ONE `GET /repos/{o}/{r}/tarball/{ref}` (extract to temp dir, parse the commit SHA from
  the archive's top-level folder) — no git at all. That beats go-git on the exact axis we cared
  about (git out of the chassis): no go-git dep, no `git` binary, just net/http + the read-only
  token. So reuse `FetchToDir` instead of cloning. User confirmed.
- **Lift the fetcher into a neutral package `internal/reposource` (turn 27).** `GitHubSource`/
  `NewGitHubSource`/`FetchToDir`/`extractTarGz` moved verbatim (path-traversal guard intact) out of
  `internal/adapters/analyser` so BOTH the analyser adapter AND the new diagnose action use it
  without an action→adapter import. Added a `Fetcher` interface as the seam. Analyser-side edits:
  delete the old `github_source.go`, repoint `adapter.go`'s one `NewGitHubSource` call to
  `reposource.NewGitHubSource` (+ import); `analyse_action.go` unchanged (its local `repoSource`
  interface is still satisfied structurally). User chose the lift over a direct action→adapter import.
- **Diagnose-agent FULLY self-contained: fetch once, analyse in-process (turns 26-27).** Not
  "fetch for bodies AND still call the analyser for the graph" (two fetches + a cross-pod hop) —
  fetch once and run `analysis` in-process for the call graph + spans, read bodies from that
  checkout. `analyse_repo` stops being a cross-pod call; `repoRoot` becomes a real local path.
  RESOLVED (turn 27): the in-process entrypoint ALREADY EXISTS and is public — `analysis.Analyse(root)
  (Output, error)` (analyse.go, already called by the analyser adapter's analyse_action.go). Nothing
  new to export. `lookup_symbols` still reads the `code_symbols` index for retrieval seeding —
  self-contained = no runtime analyser-adapter call, NOT independence from the shared index.
- **Fetch at the SAME commit the index was built on (turns 26-27).** The new action reads the
  DOMINANT `commit_sha` from `code_symbols` for the composed `owner/repo` label (one read-only
  SELECT, GROUP BY commit_sha ORDER BY COUNT(*) DESC) and fetches at that SHA, so the path:Symbol
  entries lookup seeds from the index resolve in the fetched tree. Best-effort: empty index / read
  error falls back to `ref` (HEAD). Gated by a `pin_to_index_commit` config flag (default true).
  Ordering note: this happens INSIDE analyse_repo (which runs before lookup), so the pin comes from
  the DB, not from lookup's output — keeps the workflow step order unchanged.
- **analyse_repo_local output SHAPE = analysis.Output at the TOP LEVEL + commit_sha/owner/repo/ref
  (turn 27).** So `repo_analysis.root` resolves to the local checkout and decoding `repo_analysis`
  as `analysis.Output` yields `.files` — exactly what diagnose_assemble_bundle (root + spans) and
  diagnose_route (call graph) read, UNCHANGED. NOT the analyser adapter's wrapped AnalyseResult
  ({..., output: Output}); the diagnose workflow never runs index_code_symbols (separate
  code-indexer agent, still on request_repo_analysis), so no shape conflict. owner+repo are
  included so the composing lookup (post repo-label patch) can read repo_analysis.owner/repo. The
  checkout is created ONCE (analyse_repo = start_step; loop returns to load_runtime) and is
  deliberately NOT cleaned up — it must outlive the action for every iteration's body reads;
  ephemeral pod teardown reclaims it.
- **diagnose_route SEEDS the LoopState on iteration 1 via InitLoopState (turn 28).** The chassis loop
  is workflow-driven, so diagnose_route owns first-iteration state init; relying on the zero-value
  LoopState left PrevScopeSize=0 (vs the intended seed.size()+1) and an empty hypothesis, which tripped
  the scope-must-narrow guard on the very first re-scope (next.size() > 0+2) and stopped the loop at
  iteration 1. Seed hypothesis = the symptom; seed scope = the same chain assemble_bundle uses
  (seed_scope → code_results). Reuses InitLoopState + scopeFromCodeResults; code-only, no migration.
- **diagnose_route reads its LoopState from route.diagnose_state, not a bare diagnose_state (turn 29).**
  An action's result lands under its output_field; route's is "route", so its prior state is at
  route.diagnose_state. Pointing state_field at the bare "diagnose_state" (which never exists at top
  level) made the loop re-seed every iteration — cap unenforced, trail truncated to the last pass, and
  the cross-iteration guards reset. Fixed via migration (operative, no rebuild) + the code default.
  This is the same convention diagnose_emit already uses (route.status/route.conclusion/...).
- **State-threading self-check failsafe over a redundant counter (turn 30).** The loop is already bounded
  by 4 guards + max_iterations (loop-level) and timeout_seconds + fuel_budget (engine-level, independent of
  the loop bookkeeping). A redundant iteration counter adds no robustness. The one residual risk is a
  threading REGRESSION silently disarming the cap+guards (the bug we hit). So diagnose_route now self-checks:
  if it re-seeds while route.diagnose_state already exists, it stops loudly (stopped_by=state-threading-error)
  rather than spinning. Engine timeout/fuel remain the threading-independent net.
- **A new data_request counts as forward progress in the spin guards (turn 31).** The
  evidence-must-grow (and no-thrash) guards measured progress only by NEW citations, so a verdict
  that fixed its query but couldn't yet cite the result looked identical to spinning and stopped the
  loop one iteration before the answer arrived. guardAfter now also tracks issued read-only
  data_requests (SeenRequests, threaded like SeenCitations); a NEW unseen request defers the stop, a
  re-issue does not, and the iteration cap still bounds the worst case. Chosen over loosening the
  guard to "any data_request skips" (which would neuter evidence-must-grow for most iterations) —
  precision keeps the anti-spin purpose intact.
- **GitHub read token scoped to the diagnoser via secretKeyRef, not passthrough (turn 25).**
  `spawn_actions.go` injects `GITHUB_READ_TOKEN` from `personae-platform-secrets` only when
  `isRepoCloningAgent(agentType)` (currently just `diagnose-agent`). secretKeyRef → the spawning
  pod never holds the token and no other agent type gets it; no RBAC change (that Secret is
  already referenced for DB/Anthropic creds). Read-only single-repo PAT, same one the analyser
  uses.

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
