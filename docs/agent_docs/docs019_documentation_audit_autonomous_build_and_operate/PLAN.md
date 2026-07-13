# PLAN — diagnosis-loop workstream

The diagnosis loop is an AGENT on the chassis (a workflow of steps calling
registered actions): gather → verdict (its own `execute_llm_prompt`) → route →
[loop | emit]. Gather reuses existing live actions (`request_repo_analysis`,
`lookup_code_symbols`, `execute_llm_prompt`) plus four new diagnose actions
(`diagnose_load_runtime`, `diagnose_assemble_bundle`, `diagnose_route`,
`diagnose_emit`). The workflow lives in `default_config`, not the NULL workflow
columns. See `NOTES_running_synthesis.md` for state and `RUNBOOK.md` for how-to.

---

## DONE

- Engine built + tested standalone in `contextkit/internal/diagnose/` (loop, step,
  callgraph, gatherer, verdict_wire, advance, docselect, queryselect, sqlguard).
- Verdict prompt (`PROMPT_diagnosis_verdict.md`): cite-or-abstain, worked
  gamesdesign reversal, the data_requests read-only rule.
- Chassis integration designed + drafted; the engine partially ported to
  `pkg/diagnose` and the four actions present in `platform/orchestration/actions/`.
- Three-guard model-SQL feature: Guard 1 (prompt), Guard 2 (`IsReadOnlySQL`),
  Guard 3 (read-only transaction, confirmed on the chassis side).
- **This thread:** `bundle_diagnosis_loop.sh`; `analysis.ReadSymbolBody`
  (tested + byte-identical to `cmd/assembler`); `diagnose_assemble_bundle` merged
  to use it (stub removed).

## IN FLIGHT (this thread's tail — small, mostly your-env)

1. Place `symbolbody.go` in both `internal/analysis` copies; `go test`.
2. Build the merged action in the chassis (`go build ./platform/orchestration/actions/`).
3. Collapse `cmd/assembler` onto `analysis.ReadSymbolBody`; confirm output unchanged.
4. Port `sqlguard.go` (+ test) into chassis `pkg/diagnose`; re-run the bundle and
   confirm `IsReadOnlySQL` resolves.

## GATED (the real remaining work — order unchanged from the handoff)

1. **Data-request EXECUTION wiring** in `diagnose_load_runtime` (deploy-side,
   untestable in the sandbox): after the fixed runtime reads, run each
   lint-surviving `DataRequest` under `params.DB.BeginTx(ctx,
   &sql.TxOptions{ReadOnly:true})` + `SET LOCAL statement_timeout`, format rows
   into runtime_evidence, `defer tx.Rollback()`. Set the domain `SchemaTables` so
   the bundle carries the `\d` digest. Harness equivalent needs a GRANT-based
   read-only role (`diagnose_ro`, SELECT-only — NOT `ALTER ROLE … SET
   default_transaction_read_only`, unreliable under pgbouncer transaction pooling).
2. **Confirm chassis mechanics** in the migration footer (diagnose_route
   `next_step` override under the step name; assemble re-executes on the
   route→assemble loop-back with `route.*` visible; `{{.bundle.bundle}}` via
   input_fields; `default_config` is the workflow column [confirmed]; call_agent
   child-result key).
3. **Analyser path verification in production** (the upstream blocker): smoke →
   first index (watch awaited-reply nesting: data vs body) → verify
   `code_symbols WHERE repo='gqls/agentchassis'` → B4a-on-live. The loop's gather
   reuses `request_repo_analysis` + `lookup_code_symbols`, so they must work.
4. **Build + deploy the diagnose actions + migration; trigger once by hand**
   (mirror the 080c generic-request envelope).
5. **THE EVAL GATE.** Run the LIVE loop (real model verdicter, not a script) on the
   gamesdesign bug + the 016 §9 catalogue; it MUST reproduce the mid-course
   REVERSALS and ABSTAIN when unsettled, not confirm first guesses. No automatic
   triggering until this passes. Scaffold correct ≠ reasons well.

---

## Discipline (carry these — load-bearing, and this project has been bitten)

- Reuse before recreate; alter existing functions/structs/actions before new ones.
- Schema-before-SQL: `\d` first, always. A 0-row result is not decisive until the
  query itself is checked.
- Structural fix over quick patch. Read REAL code/signatures/schema — don't assume
  (assumed helpers + stale copies have bitten repeatedly).
- Copy a Go package across a module boundary as a WHOLE DIR + diff the file LISTS;
  grep every shared-package call against the real package before building.
- Untested code is a liability — behaviour-test, don't eyeball.
- Thin workflows; complexity in Go action code. Keep workflow variable names in
  sync with what actions read. No sub-workflows in SQL — spawn sub-agents. Every
  agent is an orchestrator; agents reply to the CALLER's responses topic.
- text + CHECK not enums; version + previous_version_id; deleted_at soft-delete.
- No `logger.Debug` (won't show). British English. Flag any variable/signature
  change explicitly.

---

## CHANGELOG

- **2026-06-24** — Added `bundle_diagnosis_loop.sh`. Implemented + tested
  `analysis.ReadSymbolBody` (proven byte-identical to `cmd/assembler`). Merged
  `diagnose_assemble_bundle_action.go` to decode the typed `analysis.Output` and
  call the shared slicer; removed the `readSymbolBody` stub. Established: engine at
  `contextkit/internal/diagnose`, chassis `pkg/diagnose` is a partial port (missing
  `sqlguard.go`); `cmd/bundle` is a wrapper, slicer is `cmd/assembler`; the two
  `analyse.go` copies have drifted (types appear shared). Opened follow-ups:
  place `symbolbody.go` in both analysis copies, build the chassis action, collapse
  `cmd/assembler`, port `sqlguard.go`.
