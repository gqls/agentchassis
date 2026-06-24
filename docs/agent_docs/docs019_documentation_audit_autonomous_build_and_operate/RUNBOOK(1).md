# RUNBOOK — diagnosis loop: symbol-body slicing, bundle assembly, and what remains

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
`*ana` is the loaded analysis Output (the `*analysis.Output` from `loadAnalysis`;
`ReadSymbolBody` takes the value, hence `*ana`). `start/end` from `locateSymbol`
(for the header) and the body from `ReadSymbolBody` agree because both pick the
first match. Once nothing else calls `readLines`, delete it — keep `locateSymbol`
as the header-metadata lookup. Do NOT make `ReadSymbolBody` return `kind/start/end`:
the chassis action doesn't need them, so keep its contract as "give me the body".

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
diff is empty. If you also added a `Type.Method` scope, that line flips from
"symbol not found — skipped" to a real body; that is the only intended change.

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

For the diagnosis-loop RUN procedure itself (seed / scripted / stub / live-gather
runs, the verdict script, the flags), see the existing
`RUNBOOK_design_diagnosis_loop*.md` — not repeated here.

---

## 6. Completing the whole task (what remains)

The work in §1–4 is plumbing; it does not by itself produce a loop that reasons.
The path from here, in order, each with a definition of done (DoD):

**A. Close the in-flight tail (§1–4).** symbolbody in both analysis copies +
tests green; the merged action builds in `$CHASSIS`; `cmd/assembler` collapsed
onto the shared slicer (bundle diff unchanged); `sqlguard.go` ported and resolving
in the bundle. *DoD:* `go build ./...` and `go test ./...` clean in both modules;
`bundle_diagnosis_loop.sh` resolves all seven engine scopes + the four actions.

**B. Data-request EXECUTION wiring in `diagnose_load_runtime`** (deploy-side;
cannot be tested in the sandbox). The verdict may emit `data_requests` (vetted
SELECTs that survive the `IsReadOnlySQL` lint); today they are surfaced but not
run automatically. After the fixed runtime reads, execute each surviving request
under `params.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly:true})` + `SET LOCAL
statement_timeout` (the chassis `query_timeout=120` is only an outer bound),
format rows into `runtime_evidence`, `defer tx.Rollback()`. Set the domain
`SchemaTables` so the bundle carries the `\d` digest. The harness equivalent
needs a GRANT-based read-only role (`diagnose_ro`, SELECT-only — NOT `ALTER ROLE
… SET default_transaction_read_only`, unreliable under pgbouncer transaction
pooling). *DoD:* a hypothesis needing a query gets its rows automatically in the
next iteration's bundle, under a read-only tx proven by a refused write probe.

**C. Confirm the chassis workflow mechanics** (the migration footer). The loop is
WORKFLOW-driven (gather → verdict → route → [loop | emit]); verify: `diagnose_route`
`next_step` override is read under the STEP NAME (vs a live `conditional_route`);
`diagnose_assemble_bundle` re-executes on the route→assemble loop-back with
`route.*` visible; `{{.bundle.bundle}}` surfaces via input_fields; `default_config`
is the workflow column (confirmed); the `call_agent` child-result key is correct.
*DoD:* a scripted loop-back actually re-scopes and re-gathers — not just emits once.

**D. Analyser path verification in production** (the upstream blocker). The loop's
gather REUSES `request_repo_analysis` + `lookup_code_symbols`, so the
request → adapter response → index path must work end to end: smoke the analyser
adapter (deployed as uk_001); first index of `gqls/agentchassis`; verify
`code_symbols WHERE repo='gqls/agentchassis'`; re-run the B4a finding on live data.
Watch the awaited-reply nesting (data vs body — confirm, don't guess). *DoD:*
`lookup_code_symbols` returns real chassis symbols and the bundle's code section
is populated from the live index.

**E. Seed agents + register actions + paste the prompt + build/deploy.** Two
agent-definition rows (`default_config`-shaped); registry entries (Category
`diagnose`) for the four new actions; the verdict prompt
(`PROMPT_diagnosis_verdict.md`) pasted into the `execute_llm_prompt`
`prompt_template` (JSON-escaped); build + deploy the diagnose actions + migration.
*DoD:* the agent exists, the actions are registered, and a hand-trigger reaches the
workflow.

**F. Trigger once by hand** (`trigger-diagnose.sh`, mirroring the 080c
generic-request envelope). *DoD:* one full read-only pass runs end to end against a
real bug (gamesdesign), producing a bundle and a verdict, writing nothing and
triggering nothing.

**G. THE EVAL GATE — the real test.** Run the LIVE loop (real model verdicter, NOT
the script) on the gamesdesign bug + the 016 §9 catalogue. It must reproduce the
mid-course REVERSALS (REFUTE "sections never reach save" → REFUTE "token cap" →
CONFIRM `resolveResultSpec`) and ABSTAIN when the evidence doesn't settle the
question — not confirm first guesses. *DoD:* on the fixture the loop falsifies the
wrong hypotheses and lands on `resolveResultSpec` by FOLLOWING evidence, and on
under-determined cases it abstains. Only after this passes should any non-manual
triggering be considered. Scaffold correct ≠ reasons well.
