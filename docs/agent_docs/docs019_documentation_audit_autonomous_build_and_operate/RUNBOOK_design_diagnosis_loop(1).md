# Runbook — diagnosis loop (testing the scaffold + adapters)

**What this is.** The diagnosis loop (design: `contextkit/docs/DESIGN_diagnosis_loop.md`)
is an agent loop that wraps the read-only `cmd/bundle` gather around a
cite-or-abstain verdict, re-scoping by FOLLOWING runtime/call-graph evidence
(not re-searching the symptom — the B4a ceiling finding). This runbook is how to
RUN and TEST the parts that exist today.

**What exists (built + unit-tested in `internal/diagnose/`):**
- the deterministic **scaffold** (`loop.go`) — loop control, the four convergence
  guards, the evidence trail, re-scope-by-call-graph;
- the **BundleGatherer** (`gatherer.go`) — shells out to `cmd/bundle` (read-only);
- the **AnalysisCallGraph** (`callgraph.go`) — re-scope by following the
  analyser's `calls`, dropping ubiquitous names;
- the **`cmd/diagnose`** entrypoint wiring them together.

**What does NOT exist yet (chassis-side follow-on):**
- the **real Verdicter** — the LLM cite-or-abstain step (DESIGN §2). It needs a
  model; it can't run in the contextkit module alone. Until it's wired, the loop
  uses either a SCRIPTED verdict file (for testing a known reasoning path) or a
  STUB that abstains every iteration (so a model-less run never fabricates a
  conclusion). The runbook's "real-bug evaluation" (last section) is gated on
  this.

---

## 0. Build + unit tests (verifies the scaffold and adapters)

```bash
cd <contextkit>            # the module root
go build ./...             # whole module, incl. cmd/diagnose
go test ./internal/diagnose/ -v
```
Expect 11 passing tests: the happy-path confirm, refute-then-confirm FOLLOWING the
call graph, the no-citation→UNVERIFIABLE coercion, the four guards (iteration-cap,
scope-not-narrowing, evidence-not-growing, hypothesis-thrash), call-graph
neighbourhood resolution (incl. dropping ubiquitous names), and the
gatherer's scope→bundle-flags translation (with + without `-psql`).

These tests are the SAFETY check: the guards are what stop the loop spinning,
fabricating, or wandering, so they are behaviour-tested, not eyeballed.

---

## 1. Scripted end-to-end run (no model, no cluster) — proves the loop reasons through a path

This drives the loop through a KNOWN reasoning path via a verdict script, with
`-dry-bundle` so `cmd/bundle` writes the command it WOULD run instead of needing
a cluster. Use it to confirm the loop reproduces a real diagnosis — including the
mid-course REVERSALS, which are the behaviour that matters (DESIGN §0).

**The reference scenario is the real gamesdesign bug** (two reversals →
`resolveResultSpec`). The verdict script is the MODEL WIRE FORMAT — the same JSON
the real model emits (schema: `docs/PROMPT_diagnosis_verdict.md`), so the script
is a faithful stand-in for the model. `outcome` ∈ CONFIRMED|REFUTED|UNVERIFIABLE;
`tier` ∈ static|state|runtime:

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

# an analysis to back the call graph (any chassis/contextkit analysis works):
go run ./cmd/analyser ~/projects/agentchassis -exclude docs/ -exclude _archive/ > /tmp/chassis_clean.json

go run ./cmd/diagnose \
  -analysis /tmp/chassis_clean.json -root ~/projects/agentchassis -constitution <path>/thin_slice_constitution.md \
  -seed-hypothesis "page rebuild reports success but the live page stays stale; sections never reach save_page_sections" \
  -seed-scope "platform/orchestration/actions/save_page_sections_action.go:SavePageSectionsAction" \
  -callgraph /tmp/chassis_clean.json \
  -verdict-script /tmp/verdicts.json \
  -dry-bundle
```
Expect: a CONFIRMED diagnosis naming `resolveResultSpec`, `stopped by: confirmed`,
3 iterations, and an EVIDENCE TRAIL showing the scope evolving
save → plan_sections → coordinator across the two reversals. The per-iteration
`cmd/bundle` commands are written to `/tmp/diag_bundle_N.md` (the gather audit).

What to look for (this is the point of the loop):
- iter 1 REFUTES the seed hypothesis on RUNTIME evidence — the falsification move;
- the re-scope FOLLOWS the named symbol (call-graph), it does not re-search the
  symptom;
- it converges to a symbol the SYMPTOM could never have retrieved (the B4a
  ceiling case) — reached by FOLLOWING, the loop's whole reason to exist.

## 2. Stub run (no script) — exercises gather + guards only

Without `-verdict-script`, the stub verdicter abstains every iteration, so the
loop runs the gather + guards and stops at the cap WITHOUT concluding (it never
fabricates a verdict with no model):
```bash
go run ./cmd/diagnose \
  -analysis /tmp/chassis_clean.json -root ~/projects/agentchassis -constitution <path>/thin_slice_constitution.md \
  -seed-hypothesis "…symptom…" -seed-scope "path/file.go:Symbol" \
  -callgraph /tmp/chassis_clean.json -dry-bundle
```
Expect: `stopped by: iteration-cap`, no CONFIRMED, a NOTE on stderr that no model
is wired. This is the correct model-less behaviour — useful for checking the
plumbing against your real analysis without a script.

## 3. Live gather (real bundle, still no model)

Drop `-dry-bundle` and add `-psql` to make the loop assemble REAL bundles each
iteration (runs `cmd/bundle` for real — read-only). The verdict is still the stub
or a script (no model yet), so this checks that real bundles assemble across the
loop's evolving scope:
```bash
go run ./cmd/diagnose \
  -analysis /tmp/chassis_clean.json -root ~/projects/agentchassis -constitution <path>/thin_slice_constitution.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -seed-hypothesis "…" -seed-scope "path/file.go:Symbol" -seed-tables page_components,pages,site_work_items \
  -runtime-site gamesdesign.co.uk -runtime-page index \
  -callgraph /tmp/chassis_clean.json -verdict-script /tmp/verdicts.json
```
The bundles are written to the temp dir; inspect them to confirm the gather is
sound before the model is ever involved.

---

## 4. Flags

```
REQUIRED: -analysis -root -constitution -seed-hypothesis -seed-scope
DB:       -psql (omit => bundle skips DB gather), -seed-tables, -runtime-site, -runtime-page, -capabilities
LOOP:     -callgraph FILE (re-scope by following calls), -max-iter N (default 5), -no-follow
VERDICT:  -verdict-script FILE (scripted), or omit for the abstaining stub
TESTING:  -dry-bundle (bundle writes the command it would run; no cluster)
```

---

## 4a. Wiring the real model Verdicter (chassis-side)

The model verdict step is `docs/PROMPT_diagnosis_verdict.md` — the cite-or-abstain
contract (DESIGN §2) as the actual prompt. To wire it on the chassis:
- the agent definition holds that prompt; per iteration the loop passes the current
  hypothesis + the assembled bundle as the user content;
- the model returns ONE JSON object in the wire format (string `outcome`/`tier`);
- `diagnose.ParseVerdict` turns that JSON into the domain `Verdict` the scaffold
  consumes — this is the SEAM, and `verdict_wire_test.go` is its check. If the
  prompt's output schema and `verdict_wire.go` drift, those tests fail; keep them
  in lockstep.
- because the script format IS the wire format, every scripted scenario in §1 is a
  faithful dry-run of the model path: the bytes you script are the bytes the model
  returns.

Fail-safes already in place (so a bad model response can't do harm): an unknown
`outcome` string parses to UNVERIFIABLE (never an accidental CONFIRM); a
citation-less confirm/refute is coerced to UNVERIFIABLE by the scaffold; the
guards stop a runaway/oscillating loop regardless of what the model emits.

## 5. The real-bug evaluation gate (chassis-side, before trusting it unsupervised)

The scaffold being correct ≠ the loop diagnosing well. Before the loop is trusted
to narrow scope unsupervised, the REAL Verdicter (the model) must be wired and the
loop run against KNOWN bugs, and it must:
- reproduce the diagnosis (e.g. gamesdesign → `resolveResultSpec`), AND
- reproduce the mid-course REVERSALS (refute "sections never reach save", refute
  "token cap") rather than confirming the first guess — the move LLMs do worst
  and the one this whole design exists to force (DESIGN §0, §2).
- abstain (UNVERIFIABLE) when the bundle genuinely doesn't settle it, naming the
  missing evidence — not fabricate a confident wrong answer.

Test set: this gamesdesign bug (full trail in `RUNBOOK_gamesdesign_silent_norebuild_bug.md`),
and the silent-no-op catalogue in 016 §9. A loop that confirms the first guess on
every known bug is the failure mode, not the success — judge it on the reversals.

**A standing caution (from the B4a sessions):** the loop must apply the
cite-or-abstain skepticism to its OWN inputs too. Almost every wrong number in B4a
came from a faulty measurement setup (wrong task, contaminated index, stale var,
leaky task string), each caught by reading results skeptically. The loop's verdict
prompt should treat its own bundle/ground-truth with the same suspicion — harder
than skepticism about the target code, and the thing to watch when wiring §2.
