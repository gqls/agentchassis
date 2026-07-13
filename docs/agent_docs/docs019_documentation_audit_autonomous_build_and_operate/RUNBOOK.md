# RUNBOOK — symbol-body slicing + diagnose_assemble_bundle wiring

Concrete build/test/apply steps for the work in flight. For the diagnosis-loop
RUN procedure (seed/scripted/stub/live-gather runs, the verdict script, the flags)
see the existing `RUNBOOK_design_diagnosis_loop*.md` — not repeated here.

Environment (sandbox parity): `apt-get install -y golang-go`; `export
PATH=$PATH:/usr/lib/go-1.22/bin GOCACHE=/tmp/gocache GOPATH=/tmp/gopath`.
k8s ns `ai-persona-system` (DB pod `postgres-clients-0`, db `clients_db`, user
`clients_user`, pgbouncer 6432 transaction pooling); Kafka ns `kafka`
(`personae-kafka-cluster-*`). Deploy: GitHub → GitHub Actions → Backblaze S3.

---

## 1. `analysis.ReadSymbolBody` — place + test

The slicer is `symbolbody.go` (package `analysis`). It must sit in BOTH analysis
copies — they are separate module copies:
- `contextkit/internal/analysis/symbolbody.go`
- `<chassis>/internal/analysis/symbolbody.go`  (the action imports this one)

Test it in each module:
```bash
go test ./internal/analysis -run ReadSymbolBody -v
```

Standalone sandbox recipe (what was used to prove it here — no module needed):
```bash
D=/tmp/anatest; mkdir -p "$D"; cd "$D"; go mod init anatest
cp <path>/internal/analysis/analyse.go analyse.go
cp <path>/internal/analysis/types.go   types.go
cp <path>/symbolbody.go                symbolbody.go
cp <path>/symbolbody_test.go           symbolbody_test.go
go vet . && go test . -run ReadSymbolBody -v
```
Expect: `PASS`. The test covers a plain func (body from the `func` line to `}`,
NO doc comment), a type, receiver-qualified method disambiguation
(`Type.Method`), whole-file (`path` with no `:Symbol`), and clean errors (not
panics) for an unknown path/symbol.

Convention (do not change without re-diffing against `cmd/assembler`): the body is
file lines `[StartLine, EndLine]` INCLUSIVE, 1-indexed, exactly as the analyser
records them; the trailing newline is owned by the caller's render template.

Ground-truth check (optional, ties the helper to the real assembler output):
analyse a tree, `ReadSymbolBody(root, out, "path.go:Sym")`, and diff against the
`` ```go `` block the assembler wrote for that symbol — identical modulo the
template's trailing newline.

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
go build ./platform/orchestration/actions/
go vet   ./platform/orchestration/actions/
```
Prereq for the build to succeed: §1 must have placed `symbolbody.go` in the
chassis `internal/analysis` (so `analysis.ReadSymbolBody` exists), and that
package's `types.go` must define the fields the helper uses.

Behaviour to expect once deployed: on each iteration the action reads the in-scope
symbols' bodies (instead of erroring), composes hypothesis + code + runtime into
`bundle`, and the verdict step reads a self-contained bundle. A symbol the
analysis does not know is logged (`could not read body`) and skipped, not fatal.

---

## 3. Collapse `cmd/assembler` onto the shared slicer (reuse close-out)

`cmd/assembler` currently has its own inline body-slicer (it produced the verified
bundle). Point it at `analysis.ReadSymbolBody` so there is ONE implementation.
Before collapsing, confirm the assembler's symbol resolution matches `spanOf`:
- method-name collisions (two receivers, same method name, one file);
- whole-file scope entries;
- any "first match wins" assumptions.
Re-run a known bundle and diff the output before/after — it must be unchanged.

---

## 4. Port `sqlguard.go` into the chassis (Guard 2)

`IsReadOnlySQL` exists only at `contextkit/internal/diagnose/sqlguard.go`; the
chassis `pkg/diagnose` port is missing it. If the chassis verdict path lints
model-written data-requests (it should — Guard 2 in the three-guard model), copy:
- `<chassis>/pkg/diagnose/sqlguard.go`
- `<chassis>/pkg/diagnose/sqlguard_test.go`
then `go test ./pkg/diagnose -run IsReadOnlySQL`. Remember this is defence in
depth, NOT the safety boundary — the read-only transaction/role around execution
is the real guarantee (see the file header).

After porting, re-run the bundle script (§5) and confirm `sqlguard.go:IsReadOnlySQL`
now resolves in the bundle (it was the one scope that silently dropped).

---

## 5. The diagnosis-loop bundle script

`bundle_diagnosis_loop.sh` drives `cmd/bundle` to assemble context for working ON
the loop. Confirm before running: `pkg/diagnose` + the four actions are committed
and re-analysed into `chassis_clean.json`; `-doc`/`-constitution` paths resolve
from the contextkit working dir. The contextkit ALT block (root = the contextkit
checkout, `internal/diagnose/*`) resolves every engine symbol without the chassis
port. Regenerate the analysis when code moves:
```bash
go run ./cmd/analyser <root> -exclude docs/ -exclude _archive/ > /tmp/chassis_clean.json
```

---

## 6. Gated items (carried — see PLAN.md)

Data-request EXECUTION wiring in `diagnose_load_runtime` (read-only `BeginTx` +
`SET LOCAL statement_timeout`); analyser path verification in production;
build/deploy + one hand-triggered run; and the EVAL GATE. None are unblocked by
this thread's work alone.
