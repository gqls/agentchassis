# Runbook — diagnosis loop (testing the scaffold + adapters)

## Overview — what this is for

The wider system builds and operates marketing websites with autonomous agents.
When one of those builds goes wrong — a page silently fails to update, a rebuild
reports success but changes nothing — diagnosing it means reading the right code,
the right schema, and the right runtime evidence, then forming and *discarding*
hypotheses until the real cause is found. This conversation's work established two
things that shape the tool: first, that a single retrieval from a symptom
description **cannot** reach a cause that lives in shared infrastructure named
nothing like the symptom (measured: all retrieval methods scored zero on exactly
such a bug); and second, that the hardest, most valuable step in real debugging is
the willingness to abandon a confidently-stated wrong hypothesis when the evidence
breaks it — the step LLMs do worst by default.

The diagnosis loop is the response to both. It automates the debugging *motion* we
performed by hand throughout this work: form a hypothesis, gather scoped evidence
(read-only), judge whether that evidence confirms, refutes, or fails to settle the
hypothesis — and, crucially, when it refutes, re-scope by **following what the
evidence names** (the call graph, the runtime fault site) rather than re-searching
the symptom. It is read-only and human-gated: it emits a diagnosis with a full
evidence trail for a person to act on, and never changes code or triggers a run.
This runbook is how to build, run, and test the loop today, and what remains to
wire it to a live model. The safety and auditability live in deterministic code
(tested here); the reasoning quality lives in the model verdict (gated on a
real-bug evaluation before it is trusted).

---

**What this is (mechanism).** The diagnosis loop (design:
`contextkit/docs/DESIGN_diagnosis_loop.md`) is an agent loop that wraps the
read-only `cmd/bundle` gather around a cite-or-abstain verdict, re-scoping by
FOLLOWING runtime/call-graph evidence (not re-searching the symptom — the B4a
ceiling finding). This runbook is how to RUN and TEST the parts that exist today.

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

This section is the STANDALONE module (the dev/test harness). For building the
engine INTO the chassis (a different module), see §4c — that path has its own
import-rewrite + whole-package-copy procedure and surfaced four distinct build
errors worth doing in one pass.

```bash
cd <contextkit>            # the standalone module root
go build ./...             # whole module, incl. cmd/diagnose
go test ./internal/diagnose/ -v
```
Expect 26 passing tests: the happy-path confirm, refute-then-confirm FOLLOWING the
call graph, the no-citation→UNVERIFIABLE coercion, the four guards (iteration-cap,
scope-not-narrowing, evidence-not-growing, hypothesis-thrash), the pure `DecideStep`
guard/re-scope decisions (`step_test`), the chassis-facing `Advance` threaded across
iterations reproducing `Run()` with state round-trips (`advance_test`), the
verdict-wire parse incl. fail-safe unknown→UNVERIFIABLE (`verdict_wire_test`),
call-graph neighbourhood resolution (incl. dropping ubiquitous names), and the
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

## 4b. The chassis build — DRAFTED (the actions, the agent, the wiring)

As of 2026-06-17 the chassis side is drafted (in `chassis-drafts/`), grounded in
the real registry, action signatures, table schema, and a live agent row. The loop
runs as an AGENT, workflow-driven and observable — the verdict is its OWN step, not
buried in a monolith (the design decision that shaped this):

```
diagnose-agent workflow (default_config, processing_mode "task"):
  analyse_repo (request_repo_analysis)         REUSE — deployed analyser adapter
   → lookup_symbols (lookup_code_symbols)      REUSE — deployed, vector+trigram
   → load_runtime (diagnose_load_runtime)      NEW   — read-only SQL, real schema
   → assemble_bundle (diagnose_assemble_bundle) NEW  — hypothesis + bodies + runtime
   → verdict (execute_llm_prompt)              REUSE — its OWN observable step
   → route (diagnose_route)                    NEW   — guards + re-scope; loops back
        ├─ iterate → assemble_bundle                   to assemble_bundle or → emit
        └─ stop    → emit (diagnose_emit)       NEW   — read-only report formatter
   → complete (result_from diagnosis)
```

Files (all DRAFTS for your env; do not compile in the contextkit container):
- `pkg_diagnose_advance.go` (+ `_test.go`) — copy into `pkg/diagnose/` beside the
  engine. The chassis-facing Advance API; the 26 tests include the proof that
  Advance threaded across iterations equals Run().
- `diagnose_load_runtime_action.go` — read-only SQL over agent_error_log /
  site_work_items / orchestration_states (\d-verified columns). Returns
  `runtime_evidence`. The "completed-but-no-op" signal is surfaced from
  site_work_items status+error (the gamesdesign symptom).
- `diagnose_assemble_bundle_action.go` — composes the bundle (hypothesis at top, so
  the verdict reads a self-contained bundle). Scope fallback chain:
  `route.scope` (loop) → `input_data.seed_scope` → `code_lookup.code_results`.
  `readSymbolBody` is a STUB — wire it to slice start_line/end_line from
  `repo_analysis` (NOT a re-parse).
- `diagnose_route_action.go` — the per-iteration controller. Calls the engine's
  `Advance` once, returns a `next_step` override. It is a ROUTER: sets NO
  `output_field`, so its result lands under the step name `route` and is read as
  `route.scope` / `route.hypothesis` / `route.status` / `route.conclusion` /
  `route.evidence_trail` / `route.diagnose_state` (the conditional_route pattern,
  confirmed against coordinator.go `getNextStepFromResult`).
- `diagnose_emit_action.go` — thin read-only formatter. No DB write, no new table;
  the report returns via `complete_workflow` `result_from`.
- `registry_entries_diagnose.go` — Category "diagnose": load_runtime,
  assemble_bundle, route, emit. The verdict is `execute_llm_prompt`, NOT a diagnose
  action — no entry for it. (The earlier `diagnose_run` monolith was removed.)
- `NNN_seed_diagnose_agents.sql` — `diagnose-orchestrator` (wrapper, spawn→call→
  complete, processing_mode "orchestration") + `diagnose-agent` (the loop above,
  "task"). The workflow lives in **default_config** (verified against a live
  intake-orchestrator row — NOT the three NULL workflow columns).

Triggering: the existing generic-request envelope (kcat to
`system.agent.generic.requests`, `config.agent_type: diagnose-orchestrator`,
`input_data: {symptom, seed_scope, runtime_site, ...}`) — no new triggering code.
The orchestrator spawns the diagnose-agent worker pod (the wrapper pattern: the
loop does substantive in-chassis work so it must not run on the shared chassis
pods). A `trigger-diagnose.sh` mirroring `080c_trigger_adoption...` is the manual
entry point.

STILL TO WIRE before it compiles + runs (no design left): (1) `readSymbolBody`
span-slice; (2) paste `docs/PROMPT_diagnosis_verdict.md` into the migration's
verdict `prompt_template` placeholder (JSON-escaped); (3) confirm 5 chassis
mechanics noted in the migration footer (chiefly the `route` `next_step` override
under the step name vs a live `conditional_route`, and `assemble_bundle`
re-executing on the loop-back with `route.*` visible). Then §5.

## 4c. Porting the engine into the chassis — the build sequence (do this in one pass)

The engine is its own Go package developed in a standalone module; the chassis is a
different module. Moving `internal/diagnose/` → chassis `pkg/diagnose/` and adding
the actions surfaced FOUR build errors in the first real `make build-core-manager`,
each one only visible in the TARGET module (never in the source, where the file
exists and the path is right). They are predictable; do the whole sequence up front
rather than one build at a time.

**Step 1 — copy the WHOLE package as a unit, then diff the file list.** Do not
cherry-pick files; a sibling copied at an older version won't satisfy a newer file
(error 3 below). The validated set is 11 files: `loop.go`, `step.go`, `callgraph.go`,
`advance.go`, `gatherer.go`, `verdict_wire.go` + the five `*_test.go`
(`loop_test`, `step_test`, `advance_test`, `verdict_wire_test`, `adapters_test`).

```bash
cp <source>/internal/diagnose/*.go pkg/diagnose/
diff <(ls <source>/internal/diagnose) <(ls pkg/diagnose)   # MUST be empty
```

**Step 2 — rewrite the moved-package import path in every file that imports it.**
Only `callgraph.go`, `step.go`, `adapters_test.go` import the analysis package.
Error if skipped: `package contextkit/internal/analysis is not in std`.

```bash
MOD=$(grep '^module ' go.mod | awk '{print $2}')   # github.com/gqls/agentchassis
grep -rl '"contextkit/internal/analysis"' pkg/diagnose/ \
  | xargs sed -i "s#\"contextkit/internal/analysis\"#\"$MOD/internal/analysis\"#"
```

The chassis already has `internal/analysis/` (analyse.go + types.go); confirm its
structs match what `callgraph.go` reads (`Output.Files`, `FileInfo.Functions/.Path`,
`FuncDef.Name/.Calls`) — they were in sync this time, but verify, or field-access
errors follow the import resolving.

**Step 3 — build the package ALONE first (fast feedback before the binary).**

```bash
go build ./pkg/diagnose/    # resolves: import path (err A), missing step.go
                            # identifiers (err B: DecideStep/StepInput/NewCallGraphFromJSON),
                            # stale loop.go helper (err C: bestEffortConclusion)
go test  ./pkg/diagnose/    # 26 — proves BEHAVIOUR ported, not just compilation
```

The three engine-side errors (B, C) both trace to Step 1 being incomplete — a
missing file or a stale sibling. If Step 1 copied the whole dir as a unit, they
don't occur.

**Step 4 — the actions reference the REAL shared-package APIs, not assumed ones.**
The 4th error was `undefined: datahelpers.ExtractStringSlice` in
`diagnose_assemble_bundle_action.go` — a helper that does not exist. Before building
the actions, grep every shared-package call the new actions make and eyeball each
against the real package (see 016 §0/§9, and §4 prevention here):

```bash
grep -rhoE 'datahelpers\.[A-Za-z]+' platform/orchestration/actions/diagnose_*.go | sort -u
# each must appear as `func <Name>` under platform/orchestration/datahelpers/
```

Real extractors: `ExtractNestedField` (path→interface{}), `ExtractNestedFieldString/
Map/Int`, `GetStringField`, `GetIntField`, and for slices
`ExtractStringListHelper(val interface{}) []string` / `ToStringSlice([]interface{})`.
There is NO single path→[]string helper; compose:
`ExtractStringListHelper(ExtractNestedField(collected, path))`.

```bash
go build ./platform/orchestration/actions/   # the actions package alone
make build-core-manager                        # then the binary
```

None of these four were logic bugs — the engine passed its 26 tests throughout. They
are the cost of a cross-module move, and doing Steps 1–4 in order pays it once.

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

---

## 6. How this was reached — decisions and conclusions

This section records the choices and findings behind the loop, so the *why* is
not lost behind the *how*.

### The empirical finding that shaped the design (B4a)

Before the loop existed, the open question was whether better *retrieval*
(semantic embeddings over code) would let an agent find the right code from a bug
description. Two ground-truth tasks were measured against a clean, matched index
(both decisive symbols verified present, so a miss means "unreachable", not
"absent"):
- **skinner-box** (decisive symbol named for its mechanism): lexical recall 0.50,
  semantic 0.00. Semantic was pulled to symptom vocabulary (it ranked empty-/check-
  detector structs at the top) and lost the symbol lexical found. Embeddings did
  not earn their place even where lexical partly worked.
- **resultspec** (decisive symbols `resolveResultSpec` / `extractWorkflowResult`,
  taken from the REAL fix of the gamesdesign bug): lexical 0.00, semantic 0.00,
  fused 0.00. All three missed both, because the symptom's words ("page stale,
  silently discarded") and the mechanism's words ("result, workflow, extract")
  do not intersect.

**Conclusion:** when a cause lives in shared infrastructure named for its function,
not its failure mode, symptom-based code retrieval has a CEILING — and it is a
property of the whole category, not a lexical-vs-semantic gap. No embedding or
fusion closes a zero-overlap gap. (A secondary finding: naive RRF fusion can be
WORSE than lexical alone — a lone correct lexical hit, absent from the semantic
list, is demoted below symbols both lists agree on, even when that agreement is
among semantic's wrong matches.)

This is the lever the loop pulls: not better retrieval, but iterative re-scoping
that FOLLOWS runtime/call-graph evidence from the symptom code into the
infrastructure it calls — the move retrieval cannot make. Retrieval seeds only the
first scope; evidence steers every step after.

### Design and build choices

- **Separate runbook, not folded into the thin-slice runbook.** The loop is a
  distinct subproject; keeping it separate stops the two tangling. (This file.)
- **Build the adapters + entrypoint before writing the runbook.** A runbook for
  tools that don't exist is fiction; everything documented here was built and
  tested first, then written up.
- **The scaffold is deterministic; the verdict is the only model-dependent part.**
  Loop control, the four guards, the evidence trail, and re-scope are pure Go,
  tested without a model. The cite-or-abstain judgement is an interface, stubbed
  or scripted here, real on the chassis. This puts the SAFETY (guards, read-only,
  human-gated) in code that is verified, and isolates the part that needs a model.
- **Re-scope follows the call graph, with ubiquitous names dropped.** The analyser's
  `calls` are name-based (not type-resolved), so a callee name can map to many
  symbols; following ubiquitous names (`Run`, `String`, `New`, …) would explode the
  scope into noise — the symptom-vocabulary trap in call-graph form — so they are
  dropped at the source, with the narrowing guard as backstop.
- **The model emits strings, and the verdict-script IS the wire format.** The model
  returns human-legible JSON (`"REFUTED"`, `"runtime"`), not the scaffold's int
  enums. Because the script format equals the wire format, a scripted scenario is a
  faithful dry-run of the model path — the bytes you script are the bytes the model
  returns. `verdict_wire.go` + its tests are the seam that keeps the prompt's output
  and the scaffold in lockstep.
- **Fail safe at every join.** An unknown `outcome` parses to UNVERIFIABLE (never an
  accidental CONFIRM); a citation-less confirm/refute is coerced to UNVERIFIABLE;
  the guards stop a runaway or oscillating loop regardless of model output. A bad
  model response cannot make the loop conclude wrongly or run away.

### What is proven, and what is not

- **Proven here:** the scaffold and guards behave correctly (15 tests); the adapters
  work against real analysis data; the loop runs end-to-end and REPRODUCES the real
  gamesdesign diagnosis — including the two mid-course reversals (refuting "sections
  never reach save", then "the per-section token cap is the cause") before
  converging on `resolveResultSpec`, a symbol the symptom could never have retrieved,
  reached only by following the evidence.
- **NOT proven:** that the loop *reasons* well. The scripted run proves the plumbing
  carries a good reasoning path; it does not prove a live model would produce that
  path. That is the real-bug evaluation gate (§5): wire the model, run it on known
  bugs, and require it to reproduce the reversals and to abstain when the evidence
  genuinely doesn't settle the question — not merely confirm its first guess.

### A standing caution carried into the loop

Across this work, almost every wrong measurement came from the test instrument, not
the system under test — a wrong task string, a contaminated index, a stale shell
variable, a task description that leaked the answer's vocabulary. Each was caught
only by reading results skeptically rather than accepting them. The loop's verdict
step must apply that same suspicion to its OWN inputs — the bundle, its own reading
of a quote — which is harder than suspicion about the target code, and is the thing
to watch when the model verdict is wired and evaluated.
