# Diagnosis loop — chassis integration design

How the diagnosis loop becomes a running chassis agent: triggering, file
placement, the chassis/standalone split, and the action wiring. Grounded in the
chassis pattern (GlobalActionRegistry + ActionDefinition{Handler}; agents are
workflows of steps that call registered actions; LLM calls go through the
existing `execute_llm_prompt` action).

---

## 1. The shape it takes in the chassis

The loop is NOT a new top-level runtime — it is **an agent** like page-build-handler
or the content-writer: a workflow of steps that call registered actions. The
deterministic scaffold (`pkg/diagnose`) is the engine; a thin set of **actions**
expose its steps to the workflow; the verdict step is the EXISTING
`execute_llm_prompt` action pointed at the verdict prompt.

Concretely, three things live in the chassis:
1. **`pkg/diagnose/`** — the engine (the package you already copied there): loop
   control, guards, evidence trail, call-graph re-scope, wire parser. Pure Go, no
   model. This is reused verbatim from the standalone module.
2. **A handful of actions** (in `platform/orchestration/actions/`) that wrap the
   engine's externally-driven steps — gather, verdict, re-scope, emit — so a
   workflow can drive the loop. These are NEW but THIN (they call into
   `pkg/diagnose`).
3. **An agent definition** (a row in `agent_definitions`, like every other agent)
   whose workflow wires those actions into the loop, and whose verdict step holds
   the prompt.

Why an agent and not a CLI: the chassis already gives agents the things the loop
needs — Kafka request/response topics (triggering, results), the
`execute_llm_prompt` action (the verdict, with credit/timeout handling), DB access
for evidence (the gather), and the spawn/await machinery. Re-implementing those in
a CLI would duplicate the chassis. The standalone `cmd/diagnose` stays as the
DEV/TEST harness (scripted verdicts, dry-bundle) — it does not run in production.

---

## 2. How it is triggered

Three natural triggers, in increasing autonomy — start with the first:

**(a) On demand, by a human or another agent (the first target).** A message to the
diagnose agent's requests topic with `{symptom, seed_scope, runtime_site}`. This is
the analogue of running `cmd/diagnose` by hand — a developer (or a support
workflow) says "diagnose this". Lowest risk: a person asked for it, a person reads
the result.

**(b) On a build/▸rebuild failure, via the existing work-item machinery.** The
chassis already records failures (the `agent_error_log` rows, the `site_work_items`
status). A scheduler or a post-failure step can enqueue a diagnose request when a
work item fails or a rebuild completes-but-no-op (exactly the gamesdesign symptom).
The diagnose agent then runs read-only and posts a diagnosis for triage. Still
human-gated: it produces a report, not a fix.

**(c) Proactive (later, cautiously).** A scheduled sweep that diagnoses recurring
error signatures. Only after (a)/(b) have proven the loop reasons well on known
bugs — and still emitting reports, never acting.

Triggering is just a Kafka message to the agent's requests topic in every case;
the difference is WHO sends it. Start with (a); it needs no new triggering code,
only the agent.

---

## 3. File placement — chassis vs standalone

The rule: **the engine and the prompt live in the chassis (they run in
production); the dev/test harness and the experiment artefacts stay standalone.**

### Lives in the chassis (runs in production)

| File / artefact | Chassis location | Why |
|---|---|---|
| the engine (`loop.go`, `callgraph.go`, `gatherer.go`, `verdict_wire.go`) | `pkg/diagnose/` (already there) | the deterministic core the actions call |
| `loop_test.go` + the wire/adapter tests | `pkg/diagnose/` | the safety check travels WITH the code it guards |
| the verdict prompt | `agent_definitions` row (the diagnose agent's verdict step), sourced from `PROMPT_diagnosis_verdict.md` | prompts live in agent definitions, like every other agent's |
| the new actions (gather/verdict/rescope/emit) | `platform/orchestration/actions/` | where all actions live; registered in `GlobalActionRegistry` |
| the agent definition (workflow) | `agent_definitions` (a migration/seed, like other agents) | the workflow that drives the loop |

### Stays standalone (dev/test only, NOT in the chassis)

| File / artefact | Where | Why |
|---|---|---|
| `cmd/diagnose/main.go` | the standalone module | the dev harness (scripted verdicts, dry-bundle); not a production entrypoint |
| `groundtruth_targets.json` + the B4a artefacts (`lex.json`, `sem.json`, `chassis*.json`, `embeddings.json`) | standalone | experiment inputs/outputs; not runtime code |
| `RUNBOOK_design_diagnosis_loop.md`, `DESIGN_*.md` | docs (either repo) | documentation, not code |

### The gather adapter is the one piece that changes shape

`gatherer.go` currently shells out to `cmd/bundle` (a subprocess) — fine
standalone, wrong in the chassis (an agent should not shell out to a Go binary).
In the chassis, the gather becomes **chassis actions** doing the same read-only
work `cmd/bundle` orchestrates: the analyser read, the `\d`/row/runtime DB reads
(the chassis already has DB access and the dbcontext-equivalent queries), and the
assembler compose. So `BundleGatherer` is replaced by a `ChassisGatherer`
implementing the same `Gatherer` interface — the engine is unchanged (it only
knows the interface); only the implementation swaps subprocess → in-process
actions. This is the main porting task, and it is isolated behind the interface
the scaffold already defines.

---

## 4. The actions (thin wrappers over the engine)

New actions, registered in `GlobalActionRegistry`, **Category: "diagnose"** (a new
category alongside the existing `analysis`/`code`). Each is thin — it marshals
workflow data in, calls `pkg/diagnose`, marshals the result out:

- **`diagnose_gather`** — assemble the read-only bundle for the current scope
  (analyser + schema + runtime + compose). The `ChassisGatherer` work, exposed as
  an action so the workflow step can call it. Reuses the existing DB-read and
  assembler code paths.
- **`diagnose_verdict`** — this is NOT a new action. It is the EXISTING
  `execute_llm_prompt` action, configured with the verdict prompt and
  `output_format: json`, fed the hypothesis + the gathered bundle. Its JSON output
  is the wire format `verdict_wire.go` already parses.
- **`diagnose_step`** — apply one engine iteration: take the verdict, run the
  guards, compute the next scope (call-graph follow), decide continue/stop. Calls
  `pkg/diagnose` (the guard + re-scope logic, already tested). Returns
  continue-with-new-scope OR a terminal result.
- **`diagnose_emit`** — write the final diagnosis + evidence trail (to a topic, a
  table, or a work-item note) for a human. Never a fix.

The LOOP itself (the `for` over iterations) is expressed as the workflow's own
`loop` construct (the chassis already has `loop` / `conditional` actions — see the
registry), OR kept inside a single `diagnose_run` action that calls
`diagnose.Run(...)` with a `Verdicter` that itself invokes `execute_llm_prompt`.
The second is simpler and keeps the convergence guards in the tested Go engine
rather than re-expressed as workflow conditionals — **prefer `diagnose_run`
wrapping the engine, with the verdict as an injected `execute_llm_prompt` call.**
(This matches the standing guidance: keep complexity in Go actions, keep workflows
thin; do not re-implement the loop's guards as SQL/workflow branches.)

---

## 5. The workflow (agent definition), minimal shape

```
diagnose-agent workflow:
  start: gather
  steps:
    gather:   diagnose_run            # wraps pkg/diagnose.Run; verdict = execute_llm_prompt(verdict_prompt)
              config:
                verdict_prompt_ref: <the PROMPT_diagnosis_verdict.md text>
                max_iterations: 5
                seed_from: input_data           # {symptom, seed_scope, runtime_site}
              output_field: diagnosis
              next_step: emit
    emit:     diagnose_emit            # diagnosis + evidence trail to the caller / a triage table
              next_step: complete
    complete: complete_workflow
              config: {result_from: diagnosis}   # the singular-key flatten (post-gamesdesign-fix)
```

Note the `result_from` on complete — the gamesdesign bug this loop was built to
diagnose was caused by the OLD result-extraction mishandling the singular key; the
fix (`resolveResultSpec`) makes `result_from` the correct, current way to return a
flattened result. So the diagnose agent uses the fixed contract.

---

## 6. Build order (chassis side)

1. **Port the gatherer** — `ChassisGatherer` implementing `Gatherer` via in-process
   read-only actions (analyser read + DB reads + assemble), replacing the
   subprocess `BundleGatherer`. The one real porting task; isolated behind the
   interface.
2. **Wrap the engine as `diagnose_run`** + `diagnose_emit` actions; register them
   (`Category: "diagnose"`). `diagnose_verdict` = `execute_llm_prompt` with the
   prompt — no new action.
3. **Seed the agent definition** (workflow above + the verdict prompt) as a
   migration, like other agents.
4. **Wire a trigger (a)** — a requests topic the agent consumes; test by sending a
   known symptom by hand.
5. **The real-bug evaluation gate** (from the runbook §5): run the live agent on
   the gamesdesign bug and the 016 §9 catalogue; require it to reproduce the
   mid-course REVERSALS and to abstain when the evidence doesn't settle it — not
   just confirm a first guess. Only after this passes is (b)/(c) triggering
   considered.

`pkg/diagnose` (engine + tests) is already in the chassis and unchanged; the work
is the gatherer port, the thin action wrappers, the agent definition, and the
evaluation. The standalone module remains the dev/test harness.

---

## 7. What stays read-only and human-gated (unchanged in the chassis)

The boundaries the engine enforces do not relax in the chassis: the gather is
read-only (analyser + `\d`/SELECT/log-reads + compose — no writes, no triggered
runs); the verdict only judges and cites; the loop emits a diagnosis + evidence
trail for a human. The chassis adds Kafka/DB/LLM plumbing around that core, not new
authority. A diagnose agent that could change code or trigger a rebuild would be a
different, more dangerous thing — explicitly out of scope.
