# NOTES — running synthesis (contextkit diagnosis loop, continuation thread)

Update this EVERY turn: append a dated entry under TURN LOG, fold anything
load-bearing into DECISIONS or RISKS, and refresh the STATE DIGEST so a fresh
chat can resume from the top. Memory is OFF; this file is the durable record.
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

---

## DECISIONS (with rationale)

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
