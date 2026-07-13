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

For the diagnosis-loop RUN procedure itself (seed / scripted / stub / live-gather
runs, the verdict script, the flags), see the existing
`RUNBOOK_design_diagnosis_loop*.md` — not repeated here.

---

## 6. Completing the whole task (what remains)

The work in §1–4 is plumbing; it does not by itself produce a loop that reasons.
Below is the path from here, in order, each with code/scripts and a definition of
done (DoD). Several steps are deploy- or cluster-side and cannot be run in the
sandbox; those carry a CODE SKETCH / SKELETON marker — confirm signatures and the
envelope against your env before running. **Schema-before-SQL: run `\d <table>`
before every query.** The column names used below were taken from the current
`\d` dump; confirm against live. A shared psql handle for the snippets:
```bash
export PSQL="kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db"
```

### A. Close the in-flight tail (§1–4)

One consolidated verification:
```bash
# slicer present + green in BOTH analysis copies
cd "$CK"      && go test ./internal/analysis -run ReadSymbolBody
cd "$CHASSIS" && go test ./internal/analysis -run ReadSymbolBody
# chassis action + the two CLIs build
cd "$CHASSIS" && go build ./platform/orchestration/actions/
cd "$CK"      && go build ./cmd/assembler ./cmd/bundle
# sqlguard ported + green
cd "$CHASSIS" && go test ./pkg/diagnose -run IsReadOnlySQL
# the loop bundle resolves EVERY scope (no skips)
cd "$CK" && bash "$CK/bundle_diagnosis_loop.sh"        # writes /tmp/bundle_diagnosis_loop.md
grep -F "symbol not found" /tmp/bundle_diagnosis_loop.md \
  && echo "FAIL: a scope was skipped" || echo "OK: all scopes resolved"
```
*DoD:* every command green; no "symbol not found" line in the bundle.

### B. Data-request EXECUTION wiring in `diagnose_load_runtime`

First confirm the read-only substrate is real (runnable now) — a write inside a
`READ ONLY` transaction must be refused:
```bash
$PSQL -v ON_ERROR_STOP=1 -c "BEGIN READ ONLY; DELETE FROM site_work_items WHERE false; COMMIT;" \
  && echo "UNEXPECTED: the write was allowed" \
  || echo "good: the read-only tx refused the write"
```

CODE SKETCH — insert into `diagnose_load_runtime` after the fixed runtime reads,
modelled on `maintenance_actions.go` (`params.DB` + `QueryContext`). `params.DB`
is `*sql.DB` (confirmed 2026-06-17); pgbouncer `pool_mode = transaction`. Confirm
the `diagnose.DataRequest` field names and the `IsReadOnlySQL` import path in your
env before building.
```go
// Run each lint-surviving data request READ-ONLY; append rows to runtime_evidence.
// statement_timeout bounds a runaway scan (the chassis query_timeout=120 is only
// the outer bound). Reads only; defer Rollback; never commits.
func runDataRequests(ctx context.Context, db *sql.DB, reqs []diagnose.DataRequest, into *strings.Builder) {
	for _, r := range reqs {
		if err := diagnose.IsReadOnlySQL(r.SQL); err != nil { // Guard 2 (defence in depth)
			fmt.Fprintf(into, "\n> data_request skipped (lint): %v\n> %s\n", err, r.Why)
			continue
		}
		tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		if err != nil {
			fmt.Fprintf(into, "\n> data_request tx error: %v\n", err)
			continue
		}
		func() {
			defer tx.Rollback() // always; this path never commits
			if _, err := tx.ExecContext(ctx, "SET LOCAL statement_timeout = '15s'"); err != nil {
				fmt.Fprintf(into, "\n> statement_timeout error: %v\n", err)
				return
			}
			rows, err := tx.QueryContext(ctx, r.SQL)
			if err != nil {
				fmt.Fprintf(into, "\n> data_request error: %v\n> %s\n", err, r.Why)
				return
			}
			defer rows.Close()
			fmt.Fprintf(into, "\n### data_request — %s\n\n```\n", r.Why)
			into.WriteString(formatRows(rows)) // existing row->text helper (maintenance_actions style)
			into.WriteString("```\n")
		}()
	}
}
```
Also set the domain `SchemaTables` so the bundle carries the `\d` digest (the
existing `-schema-tables` path). The runtime reads themselves query real columns,
e.g.:
```sql
-- confirm first: \d agent_error_log
SELECT occurred_at, step_name, action, error_code, error_message
FROM agent_error_log
WHERE orchestration_id = $1 OR site_id = $2
ORDER BY occurred_at DESC LIMIT 50;
```

Harness read-only role (GRANT-based; the harness `-psql` points at this — NOT
`ALTER ROLE … SET default_transaction_read_only`, which is unreliable under
pgbouncer transaction pooling):
```sql
CREATE ROLE diagnose_ro LOGIN PASSWORD '...';
GRANT CONNECT ON DATABASE clients_db TO diagnose_ro;
GRANT USAGE ON SCHEMA public TO diagnose_ro;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO diagnose_ro;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO diagnose_ro;
```
*DoD:* the probe refuses the write; a hypothesis that needs a query gets its rows
automatically into the next iteration's `runtime_evidence`, under a read-only tx.

### C. Confirm the chassis workflow mechanics

The workflow lives in `default_config`. Inspect the diagnose agents and their
steps:
```bash
$PSQL -c "\d agent_definitions"
$PSQL -c "SELECT type, display_name, category, is_active FROM agent_definitions WHERE category='diagnose' ORDER BY type;"
# inspect the workflow JSON shape (find the steps key, then list step names)
AGENT_TYPE="<the diagnose agent type from the row above>"
$PSQL -Atc "SELECT default_config FROM agent_definitions WHERE type='$AGENT_TYPE';" | jq 'keys'
```
Then run a scripted pass and confirm the route→assemble loop-back actually
re-scopes (vs emitting once) by reading the step path:
```bash
$PSQL -c "\d orchestration_states"
$PSQL -x -c "SELECT orchestration_id, status, current_step, execution_path, error
             FROM orchestration_states
             WHERE workflow_plan::text LIKE '%diagnose%'
             ORDER BY created_at DESC LIMIT 1;"
# the loop-back is real if execution_path shows: assemble -> verdict -> route -> assemble
```
Live logs (Info+ only; no logger.Debug):
```bash
APP="<diagnose agent app/deployment label>"
kubectl -n ai-persona-system logs -l "app=$APP" --tail=200 \
  | grep -E "diagnose_route|diagnose_assemble_bundle|next_step|scope"
```
*DoD:* a scripted loop-back re-scopes and re-gathers — `execution_path` shows a
second `assemble` after `route`, not a single emit.

### D. Analyser path verification in production

The loop's gather REUSES `request_repo_analysis` + `lookup_code_symbols`, so the
request → adapter response → index path must work end to end:
```bash
kubectl -n ai-persona-system get pods | grep -Ei "analys|uk_001"
# verify code_symbols was populated for the chassis repo
$PSQL -c "\d code_symbols"     # confirm the columns (repo, path, symbol, signature, ...) first
$PSQL -c "SELECT count(*) FROM code_symbols WHERE repo='gqls/agentchassis';"
$PSQL -c "SELECT path, symbol FROM code_symbols WHERE repo='gqls/agentchassis' LIMIT 20;"
# watch the awaited reply land — inspect the nesting (data vs body), do not guess
$PSQL -x -c "SELECT request_id, step_name, target_agent_type, status, sent_at, processed_at
             FROM awaited_requests ORDER BY sent_at DESC LIMIT 5;"
```
*DoD:* count > 0 for `gqls/agentchassis`; `lookup_code_symbols` returns real chassis
symbols; the bundle's code section is populated from the live index (re-run B4a).

### E. Seed agents + register actions + paste the prompt + build/deploy

```bash
# apply the seed migration (the NNN_seed_diagnose_agents.sql)
$PSQL -v ON_ERROR_STOP=1 -f "$CHASSIS/<path>/NNN_seed_diagnose_agents.sql"
# verify the two agent rows exist and carry a workflow in default_config
$PSQL -c "SELECT type, category, is_active, (default_config ? 'workflow') AS has_workflow
          FROM agent_definitions WHERE category='diagnose';"
```
Register the four action handlers in Go the way the existing actions are
(`RegisterActionInputSpec` in each action's `init()` + the handler map in
`registry.go`, Category `diagnose`), then confirm and deploy:
```bash
cd "$CHASSIS"
grep -RnE "diagnose_(load_runtime|assemble_bundle|route|emit)" platform/orchestration/actions/registry.go
# paste the verdict prompt (PROMPT_diagnosis_verdict.md) into the execute_llm_prompt
# prompt_template in the migration (JSON-escaped) before applying it.
go build ./... && go test ./...
git add -A && git commit -m "diagnose: seed agents, register actions, verdict prompt" && git push
# GitHub Actions -> Backblaze S3
```
*DoD:* two diagnose agent rows with `has_workflow = t`; the four actions appear in
`registry.go`; build/test green; image released.

### F. Trigger once by hand

SKELETON — fill the envelope from the 080c generic-request shape (correlation_id,
responses topic, target agent type); this is a skeleton, not a verified payload:
```bash
#!/usr/bin/env bash
# trigger-diagnose.sh — one hand-triggered READ-ONLY diagnosis pass
set -euo pipefail
SITE="${1:-gamesdesign.co.uk}"
SYMPTOM="${2:-index page completed but content is a stub}"
AGENT_TYPE="<the diagnose agent type>"
IMAGE="<your kafka client image>"
REQUESTS_TOPIC="<the orchestrator requests topic>"
PAYLOAD=$(printf '{"correlation_id":"diag-%s","target_agent_type":"%s","input_data":{"site":"%s","symptom":"%s"}}' \
  "$(date +%s)" "$AGENT_TYPE" "$SITE" "$SYMPTOM")
kubectl -n kafka run kafka-produce --rm -i --restart=Never --image="$IMAGE" -- \
  bash -lc "echo '$PAYLOAD' | <produce to \$REQUESTS_TOPIC>"
```
Watch it run, and confirm it wrote nothing:
```bash
$PSQL -x -c "SELECT orchestration_id, status, current_step, error FROM orchestration_states ORDER BY created_at DESC LIMIT 1;"
APP="<diagnose agent app/deployment label>"
kubectl -n ai-persona-system logs -l "app=$APP" --tail=300 | grep -E "verdict|bundle|emit"
# the run is read-only: these high-water marks must NOT move across the run
$PSQL -c "SELECT max(created_at) AS last_work_item FROM site_work_items;"
$PSQL -c "SELECT max(created_at) AS last_page FROM pages;"
```
*DoD:* one full pass produces a bundle and a verdict, writes nothing, triggers
nothing (the high-water marks are unchanged; no new orchestration was spawned by
the diagnosis).

### G. The eval gate — the real test

Run the LIVE loop (real model verdicter, NOT the script) on the gamesdesign
fixture and the 016 §9 catalogue, and assert the REASONING, not just completion:
```text
gamesdesign fixture — seed hypothesis: "sections never reach save"
expected trajectory the loop MUST reproduce (by following evidence, not re-searching):
  REFUTE  "sections never reach save"   evidence: pages/page_components show the row written
  REFUTE  "token cap"                   evidence: no truncation in the call
  CONFIRM resolveResultSpec             evidence: singular output_field collapses the page to a stub
```
Run it via the deployed agent (F) or `cmd/diagnose` live, capturing each
iteration's verdict outcome and next scope:
```bash
# example (harness form) — adjust to your live verdicter wiring:
cd "$CK"
go run ./cmd/diagnose \
  -analysis /tmp/chassis_clean.json -root "$CHASSIS" \
  -constitution "$CK/thin_slice_constitution.md" \
  -seed-hypothesis "sections never reach save" \
  -seed-scope platform/orchestration/actions/save_page_sections_action.go:SavePageSectionsAction \
  -psql "$PSQL"
# capture the per-iteration outcome (CONFIRM/REFUTE/UNVERIFIABLE) + the re-scoped target
```
Pass criteria:
- on the fixture, the loop FALSIFIES the two wrong hypotheses and lands on
  `resolveResultSpec` by FOLLOWING the call graph / queries — not by re-searching
  the symptom;
- on under-determined catalogue items it ABSTAINS (UNVERIFIABLE) rather than
  confirming a first guess.

Only after this passes should any non-manual triggering be considered. Scaffold
correct ≠ reasons well.
