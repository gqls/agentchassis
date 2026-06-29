# RUNBOOK — diagnosis loop: symbol-body slicing, bundle assembly, and what remains

_Created: 2026-06-24 · Last updated: 2026-06-24_

## What this is (high-level)

This project builds **contextkit** — a tool that assembles a tightly-scoped slice
of a codebase (the relevant source in full, its call-graph neighbourhood as
signatures, the database schema, and authored guidance) into one paste-ready
"bundle" for a single task — and, on top of it, a **diagnosis loop**. The
diagnosis loop is an AI agent that investigates a software bug strictly
READ-ONLY: it forms a hypothesis, gathers scoped evidence (code bodies plus
read-only database rows), issues a verdict that must either CITE the evidence or
ABSTAIN, then re-scopes by FOLLOWING that evidence — the call graph for code,
vetted queries for data — rather than re-searching the original symptom. It never
edits code, never runs a build, and never triggers the system it inspects; a
human acts on its findings. The hard part LLMs do worst is falsification —
abandoning a wrong hypothesis — so the loop is judged on whether it reproduces
mid-course reversals and abstains when the evidence is inconclusive. It is
developed against, and dogfooded on, the **agent-chassis** repository (a
Go/Postgres/Kafka system that runs agents which plan and build multipage
websites). This runbook covers the build/test/apply steps for the pieces
currently in flight — the shared symbol-body slicer and the bundle-assembly
action — and, in §6, the full outline of what remains to reach a working loop.

## Status checklist

Sandbox-verified code is ticked; everything that must run in your env is not.

- [x] §1 `ReadSymbolBody` written + unit-tested (standalone)
- [ ] §1 placed in BOTH `internal/analysis` copies + `go test` green in-module
- [x] §2 `diagnose_assemble_bundle` merged (gofmt-clean)
- [ ] §2 builds in the chassis (`go build ./platform/orchestration/actions/`)
- [x] §3 `cmd/assembler` collapse merged + verified byte-identical
- [ ] §3 dropped into `$CK/cmd/assembler` + rebuilt in your tree
- [ ] §4 `sqlguard.go` (+test) ported to chassis `pkg/diagnose`
- [ ] §5 bundle script resolves every scope (no "symbol not found")
- [x] §6B `diagnose_load_runtime` data-request execution merged (gofmt-clean)
- [ ] §6B built + the read-only probe run
- [x] §6B `diagnose_ro` role migration written (`NNN_create_diagnose_ro_role.sql`)
- [ ] §6B role migration applied
- [x] §6C diagnose pair seeded; design is `diagnose_route` (no `diagnose_run`); diagnostician dropped
- [x] §6C workflow-fix migration WRITTEN (`NNN_fix_diagnose_agent_workflow.sql`) — rewrites to diagnose_route in default_config + inline prompt + ai_service
- [x] §6C agent workflow-fix migration APPLIED (has_route=t, run_loop gone, gather_step=load_runtime)
- [x] §6C orchestrator workflow RESTORED (applied; UPDATE 1)
- [x] §6C coordinator `next_step` override CONFIRMED (coordinator.go:1093 getNextStepFromResult)
- [x] §6C data_requests wiring MERGED into diagnose_route + diagnose_load_runtime (gofmt-clean); SELECT-only confirmed at 3 layers
- [x] §6C build + register the four diagnose actions (Category 'diagnose') + deploy — DONE 2026-06-29 (image also carries analyse_repo_local + lifted internal/reposource)
- [x] §6C VERIFY route.scope→assemble_bundle field shape — DONE: loop_scope_field = route.scope.Symbols (matches diagnose_route's EncodeScope + assemble's ExtractStringListHelper)
- [x] §6D analyser path verified — `code_symbols` populated for the chassis (436 rows, gqls/agentchassis)
- [x] §6E workflow loop-back confirmed — run 710d3b01 cycled the loop 5× and CONFIRMED (seeding fix live)
- [ ] §6E.1 state-threading fix (state_field → route.diagnose_state) APPLIED + re-run shows trail_len == iteration and cap enforced
- [x] §6F one hand-triggered read-only pass — DONE 2026-06-29 (loop ran analyse_repo_local→…→emit→complete, read-only; correlation 7edf33b4)
- [ ] §6G eval gate passed (reversals reproduced; abstains when unsettled)

## Where the files live (cmd/* vs the chassis)

The `cmd/*` tools — `assembler`, `bundle`, `dbcontext`, `analyser` — belong to the
contextkit module and live under `$CK` (`$CK/cmd/assembler/main.go`, and so on).
They do NOT go into the chassis. The deployed chassis diagnosis AGENT does not
shell out to them: it runs the registered actions (`diagnose_load_runtime`,
`diagnose_assemble_bundle`, `diagnose_route`, `diagnose_emit`), and the assemble
action slices bodies by calling `analysis.ReadSymbolBody` in the SHARED
`internal/analysis` package directly. So, concretely:
- `assembler_main.go` (this turn's collapse) → `$CK/cmd/assembler/main.go` (contextkit).
- `symbolbody.go` → BOTH `$CK/internal/analysis/` AND `$CHASSIS/internal/analysis/`
  (separate module copies of the shared package).
- the chassis gains the four `*_action.go` files + `pkg/diagnose` (engine) + the
  `internal/analysis` slicer — but no new `cmd/*`.

The `cmd/*` tools are the dev/harness side: the bundle script and the standalone
`cmd/diagnose` loop runner (§7). They are how you exercise the loop without the
cluster; the deployed agent uses the actions.

## Environment

Sandbox parity: `apt-get install -y golang-go`; then
`export PATH=$PATH:/usr/lib/go-1.22/bin GOCACHE=/tmp/gocache GOPATH=/tmp/gopath`.

Exported paths used by every script below — set once per shell:
```bash
export CHASSIS=~/projects/agentchassis
export CK=$CHASSIS/docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit
```
`$CHASSIS` is the chassis module root (`github.com/gqls/agentchassis`); `$CK` is
the contextkit module (its own `go.mod`, module `contextkit`) nested under it.
The `cmd/*` tools and the contextkit `internal/analysis` + `internal/diagnose`
live under `$CK`; the chassis actions and the chassis `internal/analysis` +
`pkg/diagnose` live under `$CHASSIS`.

Cluster: k8s ns `ai-persona-system` (DB pod `postgres-clients-0`, db `clients_db`,
user `clients_user`, via pgbouncer 6432 transaction pooling); Kafka ns `kafka`
(`personae-kafka-cluster-*`). Deploy: GitHub → GitHub Actions → Backblaze S3.

---

## 1. `analysis.ReadSymbolBody` — place + test

The slicer is `symbolbody.go` (package `analysis`). It must sit in BOTH analysis
copies — they are separate module copies:
- `$CK/internal/analysis/symbolbody.go`
- `$CHASSIS/internal/analysis/symbolbody.go`   (the chassis action imports this one)

Test in each module:
```bash
cd "$CK"      && go test ./internal/analysis -run ReadSymbolBody -v
cd "$CHASSIS" && go test ./internal/analysis -run ReadSymbolBody -v
```

Standalone sandbox recipe (no module needed — this is what proved it):
```bash
D=/tmp/anatest; rm -rf "$D"; mkdir -p "$D"; cd "$D"; go mod init anatest
cp "$CK/internal/analysis/analyse.go"          analyse.go
cp "$CK/internal/analysis/types.go"            types.go
cp "$CK/internal/analysis/symbolbody.go"       symbolbody.go
cp "$CK/internal/analysis/symbolbody_test.go"  symbolbody_test.go
go vet . && go test . -run ReadSymbolBody -v
```
Expect `PASS`. It covers a plain func (body from the `func` line to `}`, NO doc
comment), a type, receiver-qualified method disambiguation (`Type.Method`),
whole-file (`path` with no `:Symbol`), and clean errors (not panics) on an unknown
path/symbol.

Convention (do not change without re-diffing against `cmd/assembler`): the body
is file lines `[StartLine, EndLine]` INCLUSIVE, 1-indexed, exactly as the analyser
records them; the trailing newline is owned by the caller's render template.

---

## 2. `diagnose_assemble_bundle_action.go` — the merge

What changed (the `readSymbolBody` stub is gone):
- imports `encoding/json` and `github.com/gqls/agentchassis/internal/analysis`;
- after the `repoRoot` check, decodes `repo_analysis` (from `collected_data`) into
  a typed `analysis.Output` via `decodeAnalysisOutput` (map → JSON → Output, with
  string/[]byte fallbacks);
- the scope loop calls `analysis.ReadSymbolBody(repoRoot, anaOut, sym)`;
- dropped the now-redundant `_ = analysisField`.

Build in your env (cannot compile in the sandbox — datahelpers/zap/ActionParams):
```bash
cd "$CHASSIS"
go build ./platform/orchestration/actions/
go vet   ./platform/orchestration/actions/
```
Prereq: §1 placed `symbolbody.go` in `$CHASSIS/internal/analysis`, and that
package's `types.go` defines the fields the helper uses. The two `analyse.go`
copies have drifted (contextkit has `AnalyseWithExclude` + the `name(N).go`
skip; the chassis copy is older) — confirm the chassis `types.go` matches before
relying on the build.

Behaviour once deployed: each iteration reads the in-scope symbols' bodies
(instead of erroring), composes hypothesis + code + runtime into `bundle`, and the
verdict step reads a self-contained bundle. A symbol the analysis does not know is
logged (`could not read body`) and skipped, not fatal.

---

## 3. Collapse `cmd/assembler` onto the shared slicer (reuse close-out)

**Why.** `analysis.ReadSymbolBody` and `cmd/assembler` currently slice symbol
bodies with TWO separate implementations. The assembler's is `splitScope` +
`locateSymbol` + `readLines` (in `$CK/cmd/assembler/main.go`); the shared one,
`ReadSymbolBody`, is exactly those three combined, and the diagnosis-loop action
already uses it. Two copies of one convention is the drift this project keeps
getting bitten by — collapse to one source of truth.

**The two, side by side (so the collapse is demonstrably safe):**
- `splitScope` and the shared `splitSymbol` are identical — split on the LAST
  colon.
- `locateSymbol(fi, name)` matches by BARE name, functions then types, FIRST
  match wins, with no receiver-qualified form. `ReadSymbolBody`'s `spanOf` does
  the same for bare names AND additionally resolves `Type.Method`. So `spanOf` is
  a strict SUPERSET: identical for every name the assembler resolves today, and it
  only *adds* the `Type.Method` case the assembler currently reports as "symbol
  not found". No regression is possible from the swap.
- `readLines(root, rel, start, end)` reads the file, splits on `\n`, returns
  `lines[start-1:end]` (1-indexed inclusive) — the same convention as
  `ReadSymbolBody`/`sliceLines`, which is why the bundle diff came out
  byte-identical. The only extra is `readLines`' `end == -1` "to end of file"
  sentinel, used by the whole-file branch and never by the named-symbol branch;
  `ReadSymbolBody` covers whole-file by returning the entire file for a `path`
  with no `:Symbol`, so the sentinel isn't needed.

**The change (minimal, no contract change).** In `cmd/assembler/main.go`, the
named-symbol branch keeps `locateSymbol` for the HEADER (it needs `kind`, `start`,
`end` to print ``"### path — `sym` (kind, lines start–end)"``), and routes only
the BODY read through the shared slicer:
```go
// named-symbol branch, currently:
start, end, kind, found := locateSymbol(fi, sym)
...
if full {
    src, err := readLines(*root, path, start, end)   // <-- the body read
    ...
}

// after — slice via the shared implementation:
if full {
    src, err := analysis.ReadSymbolBody(*root, *ana, sc)  // sc is the full "path:sym"
    ...
}
```
`an` is the loaded analysis Output (the `*analysis.Output` from `loadAnalysis`;
`ReadSymbolBody` takes the value, hence `*an`). `start/end` from `locateSymbol`
(for the header) and the body from `ReadSymbolBody` agree because both pick the
first match. The whole-file branch's body read converts the same way —
`analysis.ReadSymbolBody(*root, *an, path)` (a `path` with no `:Symbol` returns
the whole file). With both reads routed through the shared slicer, `readLines` is
now unused — delete it; keep `locateSymbol` as the header-metadata lookup. Do NOT
make `ReadSymbolBody` return `kind/start/end`: the chassis action doesn't need
them, so keep its contract as "give me the body".

This collapse is DONE and verified: the merged assembler was built against the
analysis package and its output diffed against the pre-collapse binary across a
bare-name func, a whole-file, and a `Type.Method` scope — byte-identical.

**Verify (must be byte-identical before/after).** Build a known bundle with the
pre-collapse binary, apply the change, rebuild the SAME bundle, diff:
```bash
cd "$CK"
FLAGS=(-analysis /tmp/chassis_clean.json -root "$CHASSIS" \
       -constitution "$CK/thin_slice_constitution.md" \
       -task "collapse check" -step debug \
       -scope pkg/diagnose/advance.go:Advance \
       -scope platform/orchestration/actions/diagnose_assemble_bundle_action.go)
go run ./cmd/bundle "${FLAGS[@]}" -out /tmp/bundle_before.md     # pre-collapse
# ...apply the cmd/assembler edit, then:
go run ./cmd/bundle "${FLAGS[@]}" -out /tmp/bundle_after.md      # post-collapse
diff /tmp/bundle_before.md /tmp/bundle_after.md && echo "unchanged — collapse safe"
```
The header lines `(kind, lines a–b)` come from `locateSymbol` (unchanged); the
``` ```go ``` blocks come from `ReadSymbolBody` (identical convention) — so the
diff is empty. Note the named-symbol branch still calls `locateSymbol` for the
header and the `found` gate, so a `Type.Method` scope is still reported as
"symbol not found — skipped" (both binaries behave identically) — `ReadSymbolBody`'s
receiver-qualified support is not reached in the assembler. If you ever want the
assembler to accept `Type.Method`, relax `locateSymbol` too; that is out of scope
for the collapse.

---

## 4. Port `sqlguard.go` into the chassis (Guard 2)

`IsReadOnlySQL` exists only at `$CK/internal/diagnose/sqlguard.go`; the chassis
`pkg/diagnose` port is missing it — that is why a bundle scope on
`sqlguard.go:IsReadOnlySQL` silently dropped. If the chassis verdict path lints
model-written data-requests (it should — Guard 2 of the three-guard model), copy
the file and its test, then test:
```bash
cp "$CK/internal/diagnose/sqlguard.go"      "$CHASSIS/pkg/diagnose/sqlguard.go"
cp "$CK/internal/diagnose/sqlguard_test.go" "$CHASSIS/pkg/diagnose/sqlguard_test.go"
cd "$CHASSIS" && go test ./pkg/diagnose -run IsReadOnlySQL -v
```
(Defence in depth, NOT the safety boundary — the read-only transaction/role around
execution is the real guarantee; see the file header.)

**Confirm it now resolves in a bundle.** Re-analyse the chassis so the new file is
indexed, build a focused bundle scoping the symbol, and check the output:
```bash
cd "$CK"

# 1) re-analyse the chassis so pkg/diagnose/sqlguard.go enters the index
go run ./cmd/analyser "$CHASSIS" -exclude docs/ -exclude _archive/ > /tmp/chassis_clean.json

# 2) build a bundle scoping just the symbol (focused check; bundle_diagnosis_loop.sh
#    already lists it too)
go run ./cmd/bundle \
  -analysis /tmp/chassis_clean.json -root "$CHASSIS" \
  -constitution "$CK/thin_slice_constitution.md" \
  -task "confirm sqlguard resolves" -step debug \
  -scope pkg/diagnose/sqlguard.go:IsReadOnlySQL \
  -out /tmp/bundle_sqlguard_check.md

# 3) CONFIRM: the symbol's section is present, and NO scope was skipped
grep -F "### pkg/diagnose/sqlguard.go" /tmp/bundle_sqlguard_check.md \
  && echo "resolved: the IsReadOnlySQL section is in the bundle" \
  || echo "NOT resolved — see below"
grep -F "symbol not found" /tmp/bundle_sqlguard_check.md \
  && echo "WARN: a scope was skipped (the analysis did not index the file)" \
  || echo "no skipped-scope lines — good"
```
The first grep matches the assembler's symbol header
(``### path — `Sym` (kind, lines a–b)``); its presence means `locateSymbol` found
`IsReadOnlySQL` in the indexed `pkg/diagnose/sqlguard.go` and `ReadSymbolBody`
sliced the body. The assembler prints `> scope "...": symbol not found ... —
skipped` when a scope misses; that line must be ABSENT here. If you see it,
the analyser didn't index the file — re-check step 1 and that the copy landed
under `$CHASSIS/pkg/diagnose/`.

---

## 5. The diagnosis-loop bundle script

`bundle_diagnosis_loop.sh` drives `cmd/bundle` to assemble context for working ON
the loop (it already uses `$CK`). Confirm before running: `pkg/diagnose` + the
four diagnose actions are committed and re-analysed into `chassis_clean.json`;
`-doc`/`-constitution` paths resolve from `$CK`. The contextkit ALT block in the
script (root = `$CK`, `internal/diagnose/*`) resolves every engine symbol without
the chassis port. Regenerate the analysis whenever code moves:
```bash
cd "$CK" && go run ./cmd/analyser "$CHASSIS" -exclude docs/ -exclude _archive/ > /tmp/chassis_clean.json
```

The loop's RUN procedure (build/test, scripted, stub, live-gather, the flags, and
wiring the real model verdicter) is inlined in §7 below, so this runbook is
self-contained.

---

## 6. Completing the whole task (what remains)

> ## ▶ CURRENT POSITION — 2026-06-29 (read this first)
>
> **§6F + §6E loop-back DONE. The seeding fix is live and the loop ran end-to-end, read-only,
> to a CONFIRMED verdict.** Run correlation `710d3b01-…` (diagnose-agent orch `8d488e01`) cycled
> the loop **five times** (`analyse_repo_local → lookup_symbols →` then `load_runtime →
> assemble_bundle → verdict → route` ×5 `→ emit → complete`) and finished `CONFIRMED` with three
> cited rows. So loop-back + re-scope work; the iteration-1 stop is gone.
>
> **But the same run exposed a STATE-THREADING bug (fixed — migration below, no rebuild).**
> The final `evidence_trail` has only **one** entry at `Iteration: 1` despite five physical
> passes, and `collected_data_keys` shows a `route` key but **no top-level `diagnose_state`**.
> `diagnose_route` reads its prior `LoopState` from `state_field`, but its result lands under
> `output_field "route"` — so the state is at `route.diagnose_state`, not the bare
> `diagnose_state` the config pointed at. It re-seeded EVERY iteration, which silently broke:
> (1) **`max_iterations` never enforced** (Iteration resets to 1 each pass — it stopped only
> because the model confirmed); (2) **trail truncated to the last iteration** (the REFUTE/
> UNVERIFIABLE journey is lost); (3) **cross-iteration guards reset** each pass. Re-scope still
> worked because that flows through `route.scope.Symbols` (a separate, correctly-threaded field).
> FIX: `state_field` → `route.diagnose_state` (consistent with how `diagnose_emit` reads
> `route.status` etc.). **Operative fix is the migration — no rebuild** (the workflow sets
> `state_field` explicitly); the code default is corrected for the next build.
>
> **NEXT — apply the migration, then re-run and read the FULL trail:**
> ```bash
> $PSQL -f NNN_fix_diagnose_route_state_threading.sql
> $PSQL -c "SELECT default_config #>> '{workflow,steps,route,config,state_field}' AS state_field
>           FROM agent_definitions WHERE type='diagnose-agent';"   # expect: route.diagnose_state
> # Re-trigger §6F with a NEW correlation_id; capture it; then:
> $PSQL -c "SELECT collected_data #>> '{route,iteration}'                          AS iteration,
>                  jsonb_array_length(collected_data #> '{route,evidence_trail}')   AS trail_len,
>                  collected_data #>> '{route,stopped_by}'                          AS stopped_by,
>                  collected_data #>> '{emit,status}'                               AS status
>           FROM orchestration_states WHERE correlation_id = '<the-new-id>' ORDER BY created_at;"
> ```
> After the fix `trail_len` should equal `iteration` (one entry per pass), and a non-converging
> run should stop at `stopped_by = iteration-cap` instead of running on.
>
> **§6G (eval gate) is NOT yet evaluable — and do not read this run as passing it.** §6G wants
> the loop to REFUTE→…→CONFIRM down to the `resolveResultSpec` coordinator cause. This run
> CONFIRMED the *symptom* hypothesis itself, citing a `cta_url` section block and a
> content-regression on a DIFFERENT page (`tool-drop-rate-simulator`) — tangential to the
> result-extraction cause the fixture expects, and exactly the move the verdict prompt's worked
> example says to REFUTE and follow upstream. Whether the loop did that in earlier passes is
> UNKNOWN because the trail was truncated. So: land the threading fix → re-run → read the full
> per-iteration trail → THEN judge whether it reaches the real cause or confirms prematurely
> (which would point at the seed-hypothesis/scope or the verdict prompt). Reserve judgement until
> the trail is whole; don't conflate this with the still-open `next_scope`-vs-`data_requests`
> channelling watch (tables in `next_scope`).
>
> **Do NOT re-trigger §6D** (code-indexer, done — 436 rows).






The work in §1–4 is plumbing; it does not by itself produce a loop that reasons.
Below is the path, reordered so the **current blocker is up front: the diagnose
agent and its workflow are not seeded** (6C). Two queries confirm that — and they
are not at fault, the columns are real, the rows simply do not exist yet:
```
SELECT type FROM agent_definitions WHERE category='diagnose';            -- (0 rows)
SELECT 1 FROM orchestration_states WHERE workflow_plan::text LIKE '%diagnose%'; -- (0 rows)
```
So nothing diagnose-related can run until 6C is done. Each step has a definition
of done (DoD). Steps marked CODE SKETCH / SKELETON are unbuilt against
signatures/envelopes to confirm in your env. **Schema-before-SQL: `\d <table>`
before every query.** A shared psql handle:
```bash
export PSQL="kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db"
```

### 6A. Close the in-flight tail (§1–4) — [ ]

```bash
cd "$CK"      && go test ./internal/analysis -run ReadSymbolBody
cd "$CHASSIS" && go test ./internal/analysis -run ReadSymbolBody
cd "$CHASSIS" && go build ./platform/orchestration/actions/
cd "$CK"      && go build ./cmd/assembler ./cmd/bundle
cd "$CHASSIS" && go test ./pkg/diagnose -run IsReadOnlySQL      # after §4 port
cd "$CK" && bash "$CK/bundle_diagnosis_loop.sh"
grep -F "symbol not found" /tmp/bundle_diagnosis_loop.md \
  && echo "FAIL: a scope was skipped" || echo "OK: all scopes resolved"
```
*DoD:* every command green; no "symbol not found" line in the bundle.

### 6B. Data-request execution in `diagnose_load_runtime` — [x] code merged, [ ] built/applied

This is now MERGED into `diagnose_load_runtime_action.go` (`runDataRequests` +
`dataRequestsFromCollected` + `formatRowsText`): on a loop-back it reads the prior
verdict's `data_requests` from `route.data_requests`, runs each in a READ ONLY
transaction with `SET LOCAL statement_timeout`, and appends the rows to
`runtime_evidence`. It imports `pkg/diagnose` for `IsReadOnlySQL` (Guard 2), so it
needs §4 (sqlguard ported to `pkg/diagnose`) to build. Confirm the wire keys
(`sql`/`why`) against `PROMPT_diagnosis_verdict.md`, and grep for an existing
generic row formatter before keeping `formatRowsText`.

Confirm the read-only substrate (Guard 3) — a write inside a `READ ONLY`
transaction must be refused. Target `site_flows` (empty; `WHERE false` touches
nothing — the lowest-risk possible probe; the point is only that the engine
refuses a DELETE inside `BEGIN READ ONLY`):
```bash
$PSQL -c "BEGIN READ ONLY; DELETE FROM site_flows WHERE false; COMMIT;" \
  && echo "UNEXPECTED: the write was allowed" \
  || echo "good: the read-only tx refused the write"
# expected: ERROR: cannot execute DELETE in a read-only transaction  -> "good: ..."
```

The harness read-only role is a migration: `NNN_create_diagnose_ro_role.sql`
(renumber to your next sequence number; run it as a role with CREATEROLE; password
set from a secret, not in the file). Verify after applying:
```bash
$PSQL -c "SELECT has_table_privilege('diagnose_ro','site_flows','SELECT') AS can_read,
                 has_table_privilege('diagnose_ro','site_flows','DELETE') AS can_delete;"
# can_read = t, can_delete = f
```
*DoD:* the probe refuses the write; the action builds; the role reads but cannot write.

### 6C. Fix the existing diagnose-agent workflow (it's `diagnose_route`, not `diagnose_run`) — [x] DONE

You already have the pair: `diagnose-agent` (worker) + `diagnose-orchestrator`
(wrapper), seeded 2026-06-20. Do NOT seed a new one (the `diagnostician` draft is
superseded). The BUILT design is the workflow-driven loop, NOT a `diagnose_run`
action — there is no `diagnose_run`. The drafted actions are `diagnose_load_runtime`,
`diagnose_assemble_bundle`, `diagnose_route`, `diagnose_emit`, and `diagnose_route`
is verified to compile against the engine (`Advance`/`LoopState`/`ParseVerdictValue`/
`Encode*` exist in `advance.go`, `Outcome.String()` in `loop.go`). The loop is:
`analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict
(execute_llm_prompt) → route (diagnose_route) → [loop back to assemble_bundle |
emit] → complete`, where `diagnose_route` runs the engine guards + call-graph
re-scope and overrides `next_step`.

The seeded `diagnose-agent` is wrong twice: its workflow names the non-existent
`diagnose_run`, and it sits in `orchestration_workflow` (the loader reads
`default_config`).

**1. Apply the workflow-fix migration.** `NNN_fix_diagnose_agent_workflow.sql`
rewrites `diagnose-agent`'s workflow to the `diagnose_route` shape in
`default_config` (verdict prompt inline, `ai_service` block) and moves the
orchestrator:
```bash
$PSQL -v ON_ERROR_STOP=1 -f "$CHASSIS/<migrations path>/NNN_fix_diagnose_agent_workflow.sql"
$PSQL -c "SELECT type,
                 (default_config ? 'workflow') AS has_wf,
                 (default_config -> 'workflow' -> 'steps' ? 'route') AS has_route,
                 (default_config -> 'workflow' -> 'steps' ? 'run_loop') AS still_run_loop,
                 (default_config ->> 'processing_mode') AS mode,
                 (orchestration_workflow IS NULL) AS orch_null
          FROM agent_definitions WHERE type IN ('diagnose-agent','diagnose-orchestrator');"
# diagnose-agent: has_wf=t, has_route=t, still_run_loop=f, mode=orchestrator, orch_null=t
```
The agent rewrite landed (verified: `has_route=t`, `still_run_loop=f`). The
ORCHESTRATOR move did NOT — it shows `has_wf=f` (its workflow was nulled out of
`orchestration_workflow` without landing in `default_config`). Restore it with
`NNN_restore_diagnose_orchestrator_workflow.sql` (re-seeds the spawn→call→complete
workflow explicitly, since the source column is now NULL; raises the
`call_diagnoser` timeout to 1800 to match the agent loop):
```bash
$PSQL -v ON_ERROR_STOP=1 -f "$CHASSIS/<migrations path>/NNN_restore_diagnose_orchestrator_workflow.sql"
$PSQL -c "SELECT type, (default_config ? 'workflow') AS has_wf,
                 (default_config -> 'workflow' -> 'steps' ? 'spawn_diagnoser') AS has_spawn
          FROM agent_definitions WHERE type='diagnose-orchestrator';"  # both t
```
Every migration that touches `agent_definitions` snapshots the row first via
`SELECT snapshot_agent('<type>', '<reason>')` (standing rule — the fix/restore
migrations already do). Do NOT apply the older
`NNN_move_diagnose_workflow_to_default_config.sql` (bannered superseded).

**2. Build + register the four actions.** `request_repo_analysis`,
`lookup_code_symbols`, `complete_workflow`, `execute_llm_prompt` EXIST. The four
diagnose actions are DRAFTED (this work) — finish `diagnose_load_runtime` and
`diagnose_assemble_bundle` per §1–4, then build + register all four (Category
"diagnose"):
```bash
cd "$CHASSIS"
grep -RnE "diagnose_(load_runtime|assemble_bundle|route|emit)" platform/orchestration/actions/registry.go
go build ./... && go test ./...
```

**3. Confirm the `next_step` override mechanism.** The whole loop depends on the
coordinator honouring a `next_step` key in an action's result (the `diagnose_route`
comment cites `coordinator.go getNextStepFromResult`, as `conditional_route` uses).
Confirm it before trusting the loop:
```bash
grep -Rn "getNextStepFromResult\|next_step" "$CHASSIS"/platform/orchestration/coordinator*.go | head
```

Then deploy: `git add -A && git commit -m "diagnose: fix workflow (diagnose_route, default_config, inline prompt+ai_service); register actions" && git push`.
*DoD:* `diagnose-agent` shows `has_route=t`/`still_run_loop=f` in `default_config`;
the four actions registered; the coordinator honours `next_step`; build/test green;
image released.

**The data_requests channel — now wired (was dormant from a wiring gap, not by
design).** Re-checking the substrate: `sqlguard.go` is explicit that the chassis runs
model SQL under a **read-only transaction** (`db.BeginTx{ReadOnly:true}`), which is the
real read-only guarantee — so model-written `data_requests` are safe in the chassis.
The catalogue (`queryselect.go`, "the model never writes SQL") is the *standalone's*
mechanism, where `dbcontext -rows` only gets read-only safety from the role. So
`runDataRequests` in `diagnose_load_runtime` is the chassis analogue of the catalogue
and SHOULD be live. It was dormant only because three things weren't wired; the fix
migration + `PATCH_wire_data_requests.md` wire them:
(1) `gather_step: load_runtime` so the loop returns to the SQL-runner (done in the
migration); (2) `diagnose_route` forwards `verdict.DataRequests` into
`route.data_requests` (it has the parsed verdict in hand before `Advance`, so the
engine not carrying `DataRequests` through `Scope` is no obstacle — code re-scope and
data re-gather are separate channels); (3) `diagnose_load_runtime` reads
`route.data_requests` by default. Flow: `load_runtime` (runs the prior verdict's
requests + re-gathers runtime) → `assemble_bundle` → `verdict` (may emit new requests)
→ `route` (forwards them) → `load_runtime` … → `emit`.

SELECT-only is enforced at THREE layers (confirmed in the code): (1) the verdict
prompt §7 instructs a single read-only `SELECT`/`WITH … SELECT` only; (2) the model's
text is filtered through `diagnose.IsReadOnlySQL` TWICE — first at the route layer
(`diagnose_route` reads `data_requests` from the verdict wire and drops non-read-only
ones) and again in `runDataRequests` before execution; (3) the read-only transaction
(`db.BeginTx{ReadOnly:true}` + `statement_timeout`) is the real backstop, rejecting
any write (incl. data-modifying CTEs) regardless of the lint.

How `diagnose_route` reads `data_requests`: the model's `data_requests` are wire data
(the verdict's JSON at `verdict.result`), so `diagnose_route` reads them from the wire
via `readOnlyDataRequestsFromWire` and keeps only the read-only ones
(`diagnose.IsReadOnlySQL`) at the route boundary — independent of the engine's `Verdict`
type. `load_runtime` then re-lints and the read-only transaction is the backstop.
Reading from the wire rather than a typed `Verdict.DataRequests` field keeps the action
decoupled from the engine package; the two are equivalent, and the lint is what matters.

CONFIRMED ISSUE + FIX (was "STILL VERIFY"): `diagnose_route` writes
`EncodeScope(NextScope)` at `route.scope`. `EncodeScope` is `json.Marshal` of the
`Scope` struct, and that struct has NO json tags, so the object's keys are the Go
field names: `{"Symbols":[...],"Tables":[...],"RuntimeSite":"...",...}`. But
`diagnose_assemble_bundle` reads `loop_scope_field "route.scope"` through
`ExtractStringListHelper` (a `[]string` reader), which coerces that OBJECT to empty —
so on every loop-back the scope fell through the fallback chain (seed_scope →
code_results) and NEVER advanced. First-pass diagnosis still worked (iteration 1 uses
code_results); only the re-scope (iterations 2+) was silently inert, which would defeat
§6E and the §6G eval. Not covered by the engine tests (they exercise `Advance`/
`DecideStep`, not the actions). **Fix = migration
`NNN_fix_assemble_bundle_loop_scope_field.sql`:** point `loop_scope_field` at the list
inside the object — `route.scope.Symbols` (capital S; `ExtractNestedField` traverses 3
levels — precedent: `input_data.section_plan.sections_ready`, `site_specs.identity.team`).
Config-only, no rebuild. Apply before §6E/§6G; the §6F first pass runs without it.

### 6D. Analyser path verification in production — [x] index populated (436 rows); lookup-in-loop verified at 6F

The loop's gather REUSES `request_repo_analysis` + `lookup_code_symbols`, so the
request → adapter response → index path must work end to end:
```bash
kubectl -n ai-persona-system get pods | grep -Ei "analys|uk_001"
$PSQL -c "\d code_symbols"     # confirm columns (repo, path, symbol, signature, ...) first
$PSQL -c "SELECT count(*) FROM code_symbols WHERE repo='gqls/agentchassis';"
$PSQL -c "SELECT path, symbol FROM code_symbols WHERE repo='gqls/agentchassis' LIMIT 20;"
$PSQL -x -c "SELECT request_id, step_name, target_agent_type, status, sent_at, processed_at
             FROM awaited_requests ORDER BY sent_at DESC LIMIT 5;"  # inspect reply nesting (data vs body)
```
*DoD:* count > 0 for `gqls/agentchassis`; `lookup_code_symbols` returns real chassis
symbols; the bundle's code section is populated from the live index.

**Adapter availability — RESOLVED 2026-06-24 (secret name/key mismatch).** The
analyser-adapter pod sat `Pending`/`CreateContainerConfigError`, looping with the image
already present. `kubectl -n ai-persona-system get pod <pod> -o
jsonpath='{.status.containerStatuses[0].state.waiting.message}'` named it exactly:
`secret "analyser-github-read" not found`. Root cause: the deployment's env mapped
`secretKeyRef{name: analyser-github-read, key: token}`, but Terraform creates no such
secret — the read-only PAT lives in `personae-platform-secrets` under key
`GITHUB_READ_TOKEN` (`main.tf`, `var.github_read_token`). The deployment was pointed at
a secret/key that never existed. Fix: point the env at the real secret+key with
`secretKeyRef`, NOT `envFrom`:
```yaml
- name: GITHUB_READ_TOKEN
  valueFrom:
    secretKeyRef:
      name: personae-platform-secrets
      key: GITHUB_READ_TOKEN
```
`envFrom` on `personae-platform-secrets` was the wrong tool — it injects EVERY key (the
read-write `GITHUB_TOKEN`, the DB passwords, `JWT_SECRET_KEY`), the opposite of the
read-only intent; `secretKeyRef` exposes only the one key. After re-applying the overlay
and deleting the stuck pod, the adapter came up `1/1 Running` and
`kubectl -n ai-persona-system get secret personae-platform-secrets -o
jsonpath='{.data.GITHUB_READ_TOKEN}' | base64 -d | head -c4` returned a real token.
Note: `GITHUB_API_BASE` is a plain env (defaults to `https://api.github.com`), never a
secret, and was never the cause. Tidy-ups left: the `analyser-adapter.yaml` config comment
still says "via envFrom / analyser-github-read Secret" — correct it to
`secretKeyRef → personae-platform-secrets/GITHUB_READ_TOKEN`; and ensure `variables.tf`
declares `github_read_token` (the renamed var) and tfvars sets it.

**Populating `code_symbols` (prerequisite — written by `index_code_symbols`, only
read by the loop).** `code_symbols` is WRITTEN solely by the `index_code_symbols`
storage action and only READ by the loop's `lookup_code_symbols`. The diagnose agent
does NOT index — its gather is `request_repo_analysis → lookup_code_symbols` (a read).
So a diagnosis run never populates the table; an empty `code_symbols` (count 0, no
repos) just means nothing has indexed yet — expected on a fresh production with no runs.

The writer is a SEPARATE agent — **`code-indexer`** (seeded 2026-06-12) — whose
workflow is `request_repo_analysis → await analyser → index_code_symbols`. Confirm it's
present, then hand-trigger it for the chassis repo:
```bash
$PSQL -c "SELECT type, version, status FROM agent_definitions WHERE type='code-indexer';"
# trigger it with the SAME 080c envelope shape as §6F, but:
#   target_agent_type = code-indexer
#   input_data        = {"owner":"gqls","repo":"agentchassis","ref":"HEAD"}
```
Concrete trigger (run where kubectl can reach the `kafka` namespace). The envelope's
`action: orchestrate` + `config.agent_type` is the generic entry point; the cosmetic
labels were corrected from a classifier copy, the kcat produce is verbatim:
```bash
set -euo pipefail

TARGET_AGENT_TYPE='code-indexer'
OWNER='gqls'
REPO='agentchassis'
REF='HEAD'
LANGUAGE='go'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "========================================="
echo "Manual code-indexer trigger"
echo "========================================="
echo "  Target Agent Type: $TARGET_AGENT_TYPE"
echo "  Owner: $OWNER   Repo: $REPO   Ref: $REF"
echo "  Timestamp: $TIMESTAMP"
echo "  SAVE THESE IDs:"
echo "    CORRELATION_ID=$CORRELATION_ID"
echo "    ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo "========================================="

kubectl -n kafka run -i --rm "kcat-code-indexer-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-code-indexer-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":{"owner":"$OWNER","repo":"$REPO","ref":"$REF","language":"$LANGUAGE"}}
JSON

echo "Triggered code-indexer via the generic entry point. Tail by correlation:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID'"
echo "  kubectl -n ai-persona-system logs -f -l agent_type=code-indexer --tail=500 | grep '$CORRELATION_ID'"
echo "Watch the indexer's actions/markers:"
echo "  kubectl -n ai-persona-system logs -l agent_type=code-indexer --tail=500 | grep '$CORRELATION_ID' | grep -E 'request_repo_analysis|index_code_symbols|analysed repo|analyse response sent'"
echo "Orchestration state:"
echo "  psql -c \"SELECT status, current_step, substring(COALESCE(error,''),1,300) AS err FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;\""
```
Watch both logs: the adapter (`analysed repo` file_count → `analyse response sent`) and
the indexer (`index_code_symbols` counts). Then re-run the §6D count — it climbs, and the
`uq_code_symbols_identity (repo, path, symbol)` constraint makes re-runs UPSERT (safe to
repeat). Label convention: `index_code_symbols` composes `repo` as the `owner/repo` form
(`gqls/agentchassis`) from the analyser reply, so `WHERE repo='gqls/agentchassis'` is the
right filter (no trigger-supplied label to mismatch).

Standing caveat: if `index_code_symbols` errors `no analysis output at
repo_analysis.output`, inspect the awaited reply's nesting in collected_data (payload key
`data` in git/thunder vs `body` in the canonical type) and adjust the index step's
`analysis_field`/`commit_field`/`repo_field` — don't conclude the index failed from a 0
before checking the reply shape and the `repo` label.

Once `code_symbols` has chassis rows, `lookup_symbols` in the diagnose loop returns a real
seed scope; until then a diagnosis trigger should pass `input_data.seed_scope` so
`assemble_bundle` has a starting scope (it errors "no scope" otherwise).

**RESULT — 2026-06-26: indexed, 436 rows.** The §6D trigger ran end to end: analyser
`analyse response sent` (status complete → `system.agent.generic.responses`), the
code-indexer's `request_analysis` awaited request `processed`, `index_symbols` wrote rows,
and `SELECT count(*) FROM code_symbols WHERE repo='gqls/agentchassis'` = 436 (sample paths
look right: `cmd/.../main.go`, `internal/adapters/...`). So `lookup_symbols` now has a live
index to seed from. Two benign observations from that run, not blockers:
- The analyser logs `Failed to fetch message from Kafka: context deadline exceeded` on
  `system.adapter.analyser.requests` every ~10s AFTER the analyse succeeded — that's the
  idle consumer-poll timeout being logged at error level (the topic exists and the request
  was served), not a fault. Cosmetic: idle fetch timeouts shouldn't be ERROR.
- `awaited_requests.target_agent_type='unknown'` for `request_analysis` is expected — the
  analyser is addressed by a fixed TOPIC, not a target agent type (agent calls like
  `feed-ingester` show a type; topic-addressed adapter calls don't).
Quick confirm the indexer orchestration reached a terminal status (the snapshot caught it
mid-run at `EXECUTING_STEP`/`index_symbols`):
```bash
$PSQL -c "SELECT status, current_step FROM orchestration_states WHERE correlation_id='<the run>'::uuid;"  # expect a terminal status
```

### 6E. Confirm the loop re-scopes (route → load_runtime → assemble_bundle) — [x] DONE 2026-06-29 (5× loop-back, CONFIRMED). Follow-up: state-threading fix (NNN_fix_diagnose_route_state_threading.sql) to restore cap/trail/guards

The loop IS workflow-driven: `diagnose_route` overrides `next_step` back to
`assemble_bundle` each iteration (or to `emit` to stop). So re-scoping shows up in
the orchestration step path (repeated `assemble_bundle` → `verdict` → `route`) and
in the emitted evidence trail. Confirm both:
```bash
$PSQL -x -c "SELECT orchestration_id, status, current_step, error
             FROM orchestration_states
             WHERE workflow_plan::text LIKE '%diagnose%'
             ORDER BY created_at DESC LIMIT 1;"
APP="<diagnose-agent app/deployment label>"
kubectl -n ai-persona-system logs -l "app=$APP" --tail=300 \
  | grep -E "iteration|REFUTED|CONFIRMED|next_scope|stopped by"
```
*DoD:* the logs/evidence trail show multiple iterations with at least one re-scope
(a REFUTED→next_scope move), not a single pass.

### 6F. Trigger once by hand — [x] DONE 2026-06-29 (end-to-end read-only pass; outcome UNVERIFIABLE/scope-not-narrowing at iter 1 → see §6E)

Trigger format = the same 080c generic-request envelope as the §6D code-indexer trigger,
with `config.agent_type` = the diagnosis ENTRY POINT, **`diagnose-orchestrator`** (it spawns
a dedicated diagnose-agent pod and forwards the result, keeping the loop off shared pods).
The pass is READ-ONLY — the loop gathers evidence and proposes a verdict; it never writes or
triggers a build.

**Step 1 — get the ids + confirm the agent types** (fill the trigger vars from these). At
least one runtime selector is needed; `site_id` is the reliable one (`load_runtime`'s
`site_work_items` query — where the "completed but stub" evidence lives — only runs when
`site_id` is set):
```bash
# a) entry-point agent + that its workflow loaded. Both should be active, has_wf=t.
#    Target the ORCHESTRATOR; diagnose-agent is the worker it spawns:
$PSQL -c "SELECT type, status, (default_config ? 'workflow') AS has_wf
          FROM agent_definitions WHERE type IN ('diagnose-orchestrator','diagnose-agent');"

# b) gamesdesign site_id — use whichever domain<->site table your schema has (the stub
#    may not have logged an error, so the sites lookup is the reliable one; agent_error_log
#    carries both columns only if the run errored):
$PSQL -c "SELECT id, domain FROM sites WHERE domain ILIKE '%gamesdesign%';"                          # if a sites table exists
$PSQL -c "SELECT DISTINCT site_id, domain FROM agent_error_log WHERE domain ILIKE '%gamesdesign%';"  # if it logged errors

# c) (optional) a recent gamesdesign orchestration to point runtime evidence at precisely:
$PSQL -c "SELECT correlation_id, status, current_step, updated_at
          FROM orchestration_states WHERE site_id = '<site_id from b>'::uuid
          ORDER BY updated_at DESC LIMIT 10;"
```

**Step 2 — trigger** (populated with what we know: `gqls/agentchassis@HEAD` is the repo to
analyse — the one we just indexed; the gamesdesign stub is the symptom; `gamesdesign.co.uk`
is the runtime site. Fill `SITE_ID`/`RUN_CORRELATION` from step 1, or leave them empty —
`load_runtime` tolerates empty via NULL filters and falls back to the domain/`runtime_site`):
```bash
set -euo pipefail

TARGET_AGENT_TYPE='diagnose-orchestrator'
OWNER='gqls'; REPO='agentchassis'; REF='HEAD'
SYMPTOM='index page completed but content is a stub'
RUNTIME_SITE='gamesdesign.co.uk'
SITE_ID=''            # <- from step 1b (optional; empty skips the site_work_items filter)
RUN_CORRELATION=''    # <- from step 1c (optional; empty filters orchestration_states by site/domain)
# To steer iteration 1 explicitly, add  "seed_scope":["path:Symbol", ...]  to input_data
# below; omitted here so lookup_symbols seeds from the symptom (the index is populated).

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'

echo "SAVE: CORRELATION_ID=$CORRELATION_ID  ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run -i --rm "kcat-diagnose-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-diagnose-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":{"owner":"$OWNER","repo":"$REPO","ref":"$REF","symptom":"$SYMPTOM","runtime_site":"$RUNTIME_SITE","site_id":"$SITE_ID","correlation_id":"$RUN_CORRELATION"}}
JSON

echo "Tail by correlation:"
echo "  kubectl -n ai-persona-system logs -f -l agent_type=diagnose-agent --tail=500 | grep '$CORRELATION_ID'"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis    --tail=500 | grep '$CORRELATION_ID'"
```
`runtime_site` → domain is CONFIRMED in `diagnose_load_runtime` (`domain_field` default
`input_data.runtime_site`), so passing `runtime_site=gamesdesign.co.uk` populates the
`agent_error_log` domain filter; `site_id` is still what reaches `site_work_items`. If the
runtime bundle is still empty, check the `site_id`/`runtime_site` values you sent resolve a
row before concluding no evidence exists — a 0 here is a selector miss, not an absence.

**Step 3 — confirm it wrote nothing** (the whole point of a READ-ONLY pass):
```bash
$PSQL -x -c "SELECT orchestration_id, status, current_step, error FROM orchestration_states ORDER BY created_at DESC LIMIT 1;"
$PSQL -c "SELECT max(created_at) AS last_work_item FROM site_work_items;"   # unchanged across the run
$PSQL -c "SELECT max(created_at) AS last_page      FROM pages;"             # unchanged across the run
```
*DoD:* one full pass produces a bundle + a verdict, writes nothing, triggers nothing.

**RUN 1 — 2026-06-26 (correlation `03c4604a-…`) — FAILED at `assemble_bundle`, read-only
held.** Both rows FAILED: orchestrator `c43e3295` (`call_diagnoser`, CHILD_ORCHESTRATION_FAILED)
and the spawned agent `d2f85ef5` (`assemble_bundle`), error:
`diagnose_assemble_bundle: no scope (tried "route.scope", "input_data.seed_scope", then code_results)`.
Nothing written (site_work_items 06-25 / pages 06-22 unchanged), so READ-ONLY held. The failure
is iteration-1 seeding, NOT the loop-back bug: `route.scope` empty (route not run yet) and
`seed_scope` empty (none passed) are expected, but `code_lookup.code_results` — the iter-1 seed
from `lookup_code_symbols` — was ALSO empty. The lookup RAN (we reached the next step) but yielded
no usable `{path,symbol}` rows. `loop_scope_field` is irrelevant here (it only bites on loop-back).

Root-cause candidates for the empty `code_lookup.code_results`, fastest first:
```bash
# (B) OUTPUT-KEY mismatch (deterministic; most likely). assemble_bundle reads
#     code_results_field = code_lookup.code_results. Confirm lookup_code_symbols
#     actually RETURNS that key with {path,symbol} elements — grep its return:
grep -n "func .*[Ll]ookup.*Code\|return map\[string\]interface{}{\|\"code_results\"\|\"results\"\|\"matches\"\|\"path\"\|\"symbol\"\|repo_field\|\"repo\"" \
  "$CHASSIS"/platform/orchestration/actions/code_symbols_actions.go
#   If the key is not "code_results" (e.g. it returns the list directly at code_lookup,
#   or under "results"), fix the workflow's code_results_field to match — config-only.

# (C) EMBEDDINGS null -> vector search returns nothing, trigram (prose symptom vs Go
#     identifiers) also finds little. Were embeddings computed at index time?
$PSQL -c "SELECT count(*) AS total, count(embedding) AS with_embedding FROM code_symbols WHERE repo='gqls/agentchassis';"
#   with_embedding = 0 -> re-index with the embedding service reachable, or the lookup
#   must lean on trigram (which won't match a prose symptom well).

# (A) REPO filter mismatch: lookup filters WHERE repo = repo_analysis.repo. The index
#     stored repo='gqls/agentchassis'. Confirm request_repo_analysis returns the SAME
#     label (owner/repo), not bare 'agentchassis' or a URL. (Check the analyser reply
#     shape / repo_analysis.repo; SQL proves the label exists:)
$PSQL -c "SELECT DISTINCT repo FROM code_symbols WHERE repo ILIKE '%agentchassis%';"
```

Unblock to test the rest of the loop NOW (verdict → route → emit) while the lookup is fixed:
pass an explicit `seed_scope` so `assemble_bundle` uses it instead of the empty `code_results`.
For a mechanics smoke any real symbols work; for the §6G eval we want the lookup seeding from
the symptom, so prefer fixing the lookup. Candidate symbols:
```bash
$PSQL -c "SELECT path, symbol FROM code_symbols WHERE repo='gqls/agentchassis'
          AND (path ILIKE '%coordinator%' OR path ILIKE '%saga%' OR symbol ILIKE '%result%'
               OR symbol ILIKE '%extract%') ORDER BY path LIMIT 30;"
# then add to the §6F trigger body, e.g.:  "seed_scope":["platform/.../result_spec.go:resolveResultSpec"]
```
gamesdesign ids confirmed: `sites` → `e33263f4-74f8-494f-b191-546845dbbddf`; note `agent_error_log`
carries ~23 site_ids for the domain (per-build instances), so `runtime_site=gamesdesign.co.uk`
(domain) is the better runtime selector here than any single `site_id`.

*DoD status:* read-only ✓; bundle+verdict ✗ (blocked at seeding). Re-running unchanged fails
identically — fix the lookup seed (or pass `seed_scope`) first.

**ROOT CAUSE CONFIRMED (2026-06-26) — repo-label asymmetry.** Ruled out (B) and (C): the
lookup returns `code_results` with `{path,symbol}` (code_symbols_actions.go:111/99-100) and
all 436 rows have embeddings. The fault is the repo filter. `index_code_symbols` COMPOSES the
label `owner/repo` (`repo_analysis.owner + "/" + repo_analysis.repo` → `gqls/agentchassis`,
lines 146-154), but `lookup_code_symbols` does NOT compose (line 59) and the workflow set
`lookup_symbols.repo_field = "repo_analysis.repo"` (bare `agentchassis`). So the lookup queried
`WHERE repo='agentchassis'` against rows under `gqls/agentchassis` → 0 hits → empty
`code_results` → "no scope". Run 2 (`960ec6bf`) almost certainly failed the same way — the
`LIMIT 1` `COMPLETED` row is a red herring (as `8b5744fc` was in Run 1); confirm by correlation.
Fixes:
- TEST NOW (config-only, no rebuild): `NNN_fix_lookup_repo_label_workaround.sql` — drop
  `repo_field`, set literal `config.repo="gqls/agentchassis"` on the lookup step. Repo-specific,
  temporary; unblocks the whole loop/eval immediately.
- STRUCTURAL (rebuild): `PATCH_code_symbols_shared_repo_label.md` — one
  `resolveCodeRepoLabel` used by index AND lookup so they can't diverge; then drop the bare
  `repo_field`/literal so lookup composes `owner/repo` for ANY repo (the workaround's REVERT
  block). Ship, revert the workaround, re-trigger.
Apply `NNN_fix_assemble_bundle_loop_scope_field.sql` too (separate; the re-scope/loop-back fix).

STATUS 2026-06-26: workaround APPLIED + verified (lookup_config = {repo:gqls/agentchassis,
top_k:12, query_field:input_data.symptom}, no repo_field). loop_scope_field migration
confirmed live (Run 2 error read `route.scope.Symbols`). Structural patch APPLIED into
`code_symbols_actions.go` (gofmt-clean) — build+deploy, then run the workaround REVERT block
so the lookup composes owner/repo for any repo. With the workaround live, the NEXT §6F trigger
should clear seeding.

NEXT BLOCKER — symbol bodies (Run 3, 2026-06-26): seeding cleared (deployed), then
assemble_bundle failed `repo root not found at "repo_analysis.root"`. Architectural, not a
config tweak: the analyser returns SPANS not bodies; code_symbols.content is retrieval text
(composeSymbolContent = kind+symbol+signature+doc+path), NOT bodies; the diagnose-agent pod
has no checkout (the analyser clones in its OWN pod). DECISION (options weighed across two
turns): OPTION 3 — the diagnose-agent re-checks-out the repo on demand. Git stays the source
of truth; nothing whole-repo in the DB; no per-iteration analyser coupling. (Option 2 — a
stateful analyser serving slices — was cleanest on responsibility but coupled every iteration
to a live analyser; option 1 — bodies in the DB — meant whole-repo Kafka payloads at index
time.) DONE this turn: spawn_actions.go injects GITHUB_READ_TOKEN via secretKeyRef
(personae-platform-secrets) gated by isRepoCloningAgent -> ONLY diagnose-agent pods get the
token; the spawner never holds it; no RBAC change (that Secret is already referenced for the
DB/ANTHROPIC creds, so the spawned pod's SA can already read it). DECIDED for option 3, REFINED turn 27: (a) **fetch via the analyser's existing TARBALL fetcher
`FetchToDir`, NOT go-git/git** — `github_source.go` does one `GET /repos/{o}/{r}/tarball/{ref}`,
extracts to a temp dir, parses the SHA from the archive folder; git stays out of the chassis
entirely (supersedes the turn-26 go-git choice); (b) **lift that fetcher into a neutral package
`internal/reposource`** so both the analyser adapter and the new action use it without an
action→adapter import; (c) **fully self-contained** — fetch once + run `analysis.Analyse(dir)`
in-process for the call graph + spans, read bodies from that checkout, drop the cross-pod analyse
call (`repoRoot` becomes a real local path). The in-process entrypoint ALREADY EXISTS and is
public — `analysis.Analyse(root) (Output, error)` (analyse.go, already used by the analyser
adapter), so nothing new to export; (d) pin the **dominant `code_symbols.commit_sha`** for the
repo (read-only SELECT inside the action, best-effort, `pin_to_index_commit` flag) so seeded
path:Symbol entries resolve in the fetched tree. The action returns the Output fields at TOP LEVEL
+ commit_sha/owner/repo/ref as `repo_analysis`, so assemble_bundle (root + spans) and route (call
graph) read it unchanged. `lookup_symbols` still reads the `code_symbols` index for retrieval
seeding (self-contained = no runtime analyser call, NOT index-independent).
BUILT (turn 27, gofmt-clean, awaiting your build+deploy): `internal/reposource/github_source.go`
(lifted), `platform/orchestration/actions/analyse_repo_local_action.go` (the new action),
`NNN_swap_analyse_repo_to_local.sql` (the workflow swap), `PATCH_lift_fetcher_and_register.md` (the
3 chassis edits). Credential injection already DONE (spawn_actions.go, turn 25).

DEPLOY SEQUENCE (run in order; you run all builds/SQL):
```text
# 1. Apply PATCH_lift_fetcher_and_register.md:
#    - git rm internal/adapters/analyser/github_source.go
#    - adapter.go: add import internal/reposource + change the one NewGitHubSource(...) call to reposource.NewGitHubSource(...)
#    - registry.go: add the "analyse_repo_local" entry (Category "code") after request_repo_analysis
# 2. Drop in the two new files (internal/reposource/github_source.go, platform/orchestration/actions/analyse_repo_local_action.go)
# 3. Build + deploy:
go build ./...            # expect clean (deps: internal/analysis, datahelpers, zap, net/http)
git add -A && git commit -m "diagnose: in-process analyse_repo_local (tarball fetch + in-process analyse); lift fetcher to internal/reposource"
git push                 # GitHub Actions -> Backblaze; note the new image tag
# 4. THEN run the workflow swap (needs the new action present):
psql "$DBURL" -f NNN_swap_analyse_repo_to_local.sql
#    verify: action=analyse_repo_local, pin=true, output_field=repo_analysis, next_step=lookup_symbols
# 5. Re-trigger §6F (same 080c envelope) and confirm by correlation_id (NEVER LIMIT 1).
#    Expect: assemble_bundle clears "repo root not found"; first real bundle + verdict produced.
```


### 6G. The eval gate — the real test — [ ]  (BLOCKED: needs the state-threading fix live so the FULL per-iteration trail is readable; run 710d3b01 CONFIRMED the symptom on tangential citations, not the resolveResultSpec cause — reserve judgement until the trail is whole)

Run the LIVE loop (real model verdicter, NOT the script) on the gamesdesign
fixture and the 016 §9 catalogue, and assert the REASONING:
```text
gamesdesign fixture — seed hypothesis: "sections never reach save"
trajectory the loop MUST reproduce (by following evidence, not re-searching):
  REFUTE  "sections never reach save"   evidence: pages/page_components show the row written
  REFUTE  "token cap"                   evidence: no truncation in the call
  CONFIRM resolveResultSpec             evidence: singular output_field collapses the page to a stub
```
Pass criteria: on the fixture it FALSIFIES the two wrong hypotheses and lands on
`resolveResultSpec` by FOLLOWING the call graph / queries; on under-determined
catalogue items it ABSTAINS (UNVERIFIABLE) rather than confirming a first guess.
Only after this passes should any non-manual triggering be considered. Scaffold
correct ≠ reasons well.

---

## 7. Running the loop (self-contained)

The standalone `cmd/diagnose` harness wires the engine together for dev/test runs
WITHOUT the cluster or a model. This is the procedure formerly kept in the design
runbook, inlined here. All paths use `$CK` / `$CHASSIS` from the Environment
section.

### 7.0 Build + unit tests (the scaffold + adapters)

```bash
cd "$CK"
go build ./...                       # whole module, incl. cmd/diagnose
go test ./internal/diagnose/ -v      # expect the engine suite green
```
The engine tests are the SAFETY check: the guards (iteration-cap, scope-not-
narrowing, evidence-not-growing, hypothesis-thrash), the no-citation→UNVERIFIABLE
coercion, the unknown-outcome→UNVERIFIABLE fail-safe, the call-graph re-scope, and
the gatherer's scope→bundle-flags translation are behaviour-tested, not eyeballed.

### 7.1 Scripted end-to-end run (no model, no cluster) — proves the loop reasons

The verdict script IS the model wire format (schema:
`PROMPT_diagnosis_verdict.md`), so it is a faithful stand-in. `-dry-bundle` makes
`cmd/bundle` write the command it WOULD run instead of needing the cluster. The
reference scenario is the gamesdesign bug (two reversals → `resolveResultSpec`):
```bash
cat > /tmp/verdicts.json << 'EOF'
[
  {"outcome":"REFUTED",
   "citations":[{"tier":"runtime","where":"agent_error_log","quote":"content regression blocked: new content ~3k vs ~13k existing","fresh":"2026-06-14"}],
   "revised_hypothesis":"the regeneration is short upstream in the content writer",
   "next_scope":["plan_sections_action.go:PlanSectionsAction"],
   "runtime_site":"page-build-handler"},
  {"outcome":"REFUTED",
   "citations":[{"tier":"static","where":"content_writer default_config","quote":"max_tokens 2000 per-section loop"}],
   "revised_hypothesis":"the writer output is silently discarded at result extraction",
   "next_scope":["coordinator.go:extractWorkflowResult"]},
  {"outcome":"CONFIRMED",
   "citations":[{"tier":"static","where":"result_spec.go:resolveResultSpec","quote":"singular output_field ignored; only plural output_fields honoured"}]}
]
EOF

cd "$CK"
go run ./cmd/analyser "$CHASSIS" -exclude docs/ -exclude _archive/ > /tmp/chassis_clean.json
go run ./cmd/diagnose \
  -analysis /tmp/chassis_clean.json -root "$CHASSIS" -constitution "$CK/thin_slice_constitution.md" \
  -seed-hypothesis "page rebuild reports success but the live page stays stale; sections never reach save_page_sections" \
  -seed-scope "platform/orchestration/actions/save_page_sections_action.go:SavePageSectionsAction" \
  -callgraph /tmp/chassis_clean.json \
  -verdict-script /tmp/verdicts.json \
  -dry-bundle
```
Expect: a CONFIRMED diagnosis naming `resolveResultSpec`, `stopped by: confirmed`,
3 iterations, and an evidence trail showing the scope evolving
save → plan_sections → coordinator across the two reversals (the per-iteration
`cmd/bundle` commands are written to `/tmp/diag_bundle_N.md`). The point: iter 1
REFUTES the seed on runtime evidence (falsification), the re-scope FOLLOWS the
named symbol (not the symptom), and it converges on a symbol the symptom could
never have retrieved.

### 7.2 Stub run (no script) — gather + guards only

```bash
cd "$CK"
go run ./cmd/diagnose \
  -analysis /tmp/chassis_clean.json -root "$CHASSIS" -constitution "$CK/thin_slice_constitution.md" \
  -seed-hypothesis "…symptom…" -seed-scope "path/file.go:Symbol" \
  -callgraph /tmp/chassis_clean.json -dry-bundle
```
Expect: `stopped by: iteration-cap`, no CONFIRMED, a stderr note that no model is
wired. The correct model-less behaviour — it never fabricates a verdict.

### 7.3 Live gather (real bundle, still no model)

Drop `-dry-bundle` and add `-psql` so the loop assembles REAL bundles each
iteration (read-only). Verdict is still the stub or a script:
```bash
cd "$CK"
go run ./cmd/diagnose \
  -analysis /tmp/chassis_clean.json -root "$CHASSIS" -constitution "$CK/thin_slice_constitution.md" \
  -psql "$PSQL" \
  -seed-hypothesis "…" -seed-scope "path/file.go:Symbol" -seed-tables page_components,pages,site_work_items \
  -runtime-site gamesdesign.co.uk -runtime-page index \
  -callgraph /tmp/chassis_clean.json -verdict-script /tmp/verdicts.json
```
Inspect the bundles in the temp dir to confirm the gather is sound before a model
is ever involved.

### 7.4 Flags

```
REQUIRED: -analysis -root -constitution -seed-hypothesis -seed-scope
DB:       -psql (omit => bundle skips DB gather), -seed-tables, -runtime-site, -runtime-page, -capabilities
LOOP:     -callgraph FILE (re-scope by following calls), -max-iter N (default 5), -no-follow
VERDICT:  -verdict-script FILE (scripted), or omit for the abstaining stub
TESTING:  -dry-bundle (bundle writes the command it would run; no cluster)
```

### 7.5 Wiring the real model verdicter (chassis-side)

The model verdict step is the cite-or-abstain contract in
`PROMPT_diagnosis_verdict.md`, pasted into the agent's `execute_llm_prompt`
`prompt_template` (6C). Per iteration the loop passes the current hypothesis + the
assembled bundle as the user content; the model returns ONE JSON object in the
wire format; `diagnose.ParseVerdict` turns it into the domain `Verdict` — that seam
is checked by `verdict_wire_test.go` (keep the prompt's output schema and
`verdict_wire.go` in lockstep). Fail-safes already in place: an unknown `outcome`
parses to UNVERIFIABLE (never an accidental CONFIRM); a citation-less confirm/refute
is coerced to UNVERIFIABLE; the guards stop a runaway/oscillating loop regardless
of what the model emits. Because the script format equals the wire format, every
scripted scenario in §7.1 is a faithful dry-run of the model path.
