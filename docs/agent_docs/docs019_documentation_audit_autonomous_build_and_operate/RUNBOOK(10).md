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
- [ ] §6C build + register the four diagnose actions (Category 'diagnose') + deploy
- [ ] §6C VERIFY route.scope→assemble_bundle field shape (route.scope.symbols vs ExtractStringListHelper)
- [ ] §6D analyser path verified (`code_symbols` populated for the chassis)
- [ ] §6E workflow loop-back confirmed
- [ ] §6F one hand-triggered read-only pass
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

### 6C. Fix the existing diagnose-agent workflow (it's `diagnose_route`, not `diagnose_run`) — [ ]  ← do this next

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

NOTE — engine-copy divergence: the chassis `pkg/diagnose` is BEHIND the contextkit
engine; its `Verdict` has no `DataRequests` field and there is no `DataRequest` type
(the build failed on these). So `diagnose_route` reads `data_requests` straight from
the RAW verdict wire (`verdict.result`) via `readOnlyDataRequestsFromWire` rather than
a typed field, and lints there with `diagnose.IsReadOnlySQL` (which IS in the chassis
copy). This is robust either way; if you later sync the chassis engine, the wire path
still works. (The parse-time lint that `verdict_wire.toVerdict` does in contextkit is
therefore replaced, in the chassis, by the route-layer lint above.)

STILL VERIFY (separate, unchanged): `diagnose_route` writes `EncodeScope(NextScope)` —
an OBJECT `{symbols, ...}` — at `route.scope`, but `diagnose_assemble_bundle` reads it
with `ExtractStringListHelper` (a LIST). Confirm the helper unwraps `.symbols`, or set
`loop_scope_field` to `route.scope.symbols`, else CODE re-scope won't flow on loop-back.
Not covered by the engine tests (they exercise `Advance`/`DecideStep`, not the actions).

### 6D. Analyser path verification in production — [ ]

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

### 6E. Confirm the loop re-scopes (route → assemble_bundle) — [ ]  (after 6C)

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

### 6F. Trigger once by hand — [ ]

SKELETON — fill the envelope from the 080c generic-request shape (correlation_id,
responses topic, target agent type = the `type` from 6C):
```bash
#!/usr/bin/env bash
# trigger-diagnose.sh — one hand-triggered READ-ONLY diagnosis pass
set -euo pipefail
SITE="${1:-gamesdesign.co.uk}"
SYMPTOM="${2:-index page completed but content is a stub}"
AGENT_TYPE="<the type you seeded in 6C>"
IMAGE="<your kafka client image>"
REQUESTS_TOPIC="<the orchestrator requests topic>"
PAYLOAD=$(printf '{"correlation_id":"diag-%s","target_agent_type":"%s","input_data":{"site":"%s","symptom":"%s"}}' \
  "$(date +%s)" "$AGENT_TYPE" "$SITE" "$SYMPTOM")
kubectl -n kafka run kafka-produce --rm -i --restart=Never --image="$IMAGE" -- \
  bash -lc "echo '$PAYLOAD' | <produce to \$REQUESTS_TOPIC>"
```
Confirm it wrote nothing:
```bash
$PSQL -x -c "SELECT orchestration_id, status, current_step, error FROM orchestration_states ORDER BY created_at DESC LIMIT 1;"
$PSQL -c "SELECT max(created_at) AS last_work_item FROM site_work_items;"   # unchanged across the run
$PSQL -c "SELECT max(created_at) AS last_page      FROM pages;"             # unchanged across the run
```
*DoD:* one full pass produces a bundle + a verdict, writes nothing, triggers nothing.

### 6G. The eval gate — the real test — [ ]

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
