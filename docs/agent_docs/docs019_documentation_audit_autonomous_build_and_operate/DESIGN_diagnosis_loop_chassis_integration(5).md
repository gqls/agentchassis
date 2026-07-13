# Diagnosis loop — chassis integration design

> **UPDATE 2026-06-24 (corrected) — the BUILT design is `diagnose_route`, not
> `diagnose_run`; there is no `diagnose_run` action.** The actions that actually
> exist are `diagnose_load_runtime`, `diagnose_assemble_bundle`, `diagnose_route`,
> `diagnose_emit` (+ the engine in `pkg/diagnose`), implementing the workflow-driven
> loop the 06-17 note described: gather → verdict (`execute_llm_prompt`) →
> `diagnose_route` (engine guards + call-graph re-scope, then a `next_step` override)
> → back to `assemble_bundle` | `emit`. So the §4–§6 `diagnose_run` recommendation
> below is the ABANDONED path — read it as `diagnose_route` with the verdict as a
> separate step. (An earlier revision of this banner wrongly said diagnose_run was
> chosen; the action files show otherwise.) The loop works because the coordinator
> honours a `next_step` key in an action's result (`getNextStepFromResult`, as
> `conditional_route` uses) — confirm that in your tree; `diagnose_route` itself is
> verified to compile against the engine (Advance / LoopState / ParseVerdictValue /
> Encode* all exist in `advance.go`, `Outcome.String()` in `loop.go`).

## Status (2026-06-24)

- **Engine** (`pkg/diagnose`): the Run / Advance / DecideStep core + convergence
  guards + verdict-wire parser exist and are tested (15 tests). Reused verbatim.
  `diagnose_route` depends only on symbols that exist here (checked).
- **Actions**: `diagnose_load_runtime`, `diagnose_assemble_bundle`, `diagnose_route`,
  `diagnose_emit` are all DRAFTED. `request_repo_analysis` + `lookup_code_symbols`
  are reused as-is. There is NO `diagnose_run` (and none is needed).
- **Agents**: `diagnose-agent` is seeded but doubly wrong — its workflow names the
  non-existent `diagnose_run` AND sits in `orchestration_workflow` (json), which the
  loader does not read. Fix = `NNN_fix_diagnose_agent_workflow.sql`: rewrites it to
  the `diagnose_route` workflow in `default_config`, verdict prompt inline, with an
  `ai_service` block. `diagnose-orchestrator` is correct (spawn+call); the same
  migration moves it to `default_config`.
- **Verdict prompt**: INLINE in the verdict `execute_llm_prompt` step
  (`PROMPT_diagnosis_verdict.md`). `diagnose-verdict-v1` was a prompt-registry ref
  under the abandoned `diagnose_run` design — not used by the built loop.
- **Known gap**: the model's `data_requests` and `runtime_site` re-gather are parsed
  and lint-checked (`verdict_wire.go`) but NOT wired into the loop — `diagnose_route`
  does not forward them and the loop-back returns to `assemble_bundle`, not
  `load_runtime`. Code-following re-scope (the call-graph) works; runtime-following
  re-gather is an enhancement (route must forward them + `gather_step: load_runtime`).
- **Gate**: the real-bug evaluation (§6 build-order step 5) has not run.

How the diagnosis loop becomes a running chassis agent: triggering, file
placement, the chassis/standalone split, and the action wiring. Grounded in the
chassis pattern (GlobalActionRegistry + ActionDefinition{Handler}; agents are
workflows of steps that call registered actions; LLM calls go through the
existing `execute_llm_prompt` action).

---

## 0. STEP ZERO — what already exists (the pre-flight, per 001 dev guide)

The dev guide's first rule: prove no existing action/agent covers the need before
creating one. Searching `registry.go` for the gather sub-steps changed this design
materially — **most of the gather is already chassis capability**:

| Gather sub-step | Existing chassis action(s) | Verdict |
|---|---|---|
| analyse a repo at a ref | **`request_repo_analysis`** (Category `code`) — "Ask the analyser adapter to parse a repo at ref; awaits the symbol output" | REUSE — this is the analyser adapter we deployed |
| retrieve relevant code symbols | **`lookup_code_symbols`** (`storage`) — "Retrieve relevant code symbols from code_symbols (vector, trigram fallback)" | REUSE — this is the resolve_targets equivalent, already live, lexical+vector |
| (re)index symbols | **`index_code_symbols`** (`storage`) | REUSE if a fresh index is needed |
| load DB context for a site/page | **`load_site_for_rebuild`** (`site`) — "Load site context… from DB"; **`load_edit_context`** | REUSE / MODEL ON — existing read-only DB-gather actions |
| validate/inspect schema | **`validate_schema`** | partial — covers validation, not arbitrary `\d`; the runtime/schema read is the gap |
| assemble/compose the bundle | `assemble_page`, `assemble_multipage_site`, `build_render_context` exist but are PAGE/SITE assemblers, NOT a code-context bundle composer | **GAP** — no code-bundle assembler in the chassis (the standalone `cmd/assembler` has no chassis equivalent) |
| the verdict (LLM, json out) | **`execute_llm_prompt`** (`llm`) — the same action the content-writer uses | REUSE — the verdict is this action + the verdict prompt |

So the honest finding (answering "do we have an assembler?"): **no code-context
assembler exists in the chassis** — `assemble_*` are page/site builders, a
different thing. But the two pieces that looked hardest to port — code analysis and
symbol retrieval — are ALREADY chassis actions (`request_repo_analysis`,
`lookup_code_symbols`), because the analyser adapter is deployed. So the
`ChassisGatherer` is mostly WIRING EXISTING ACTIONS, with one genuinely new piece:
a code-context assembler step (and arbitrary schema/runtime reads if the bundle
needs them beyond what `load_site_for_rebuild` returns).

This is exactly the reuse the dev guide demands, found by searching first.

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

## 2. How it is triggered (the on-demand path, with the real envelope)

The first trigger is on-demand, and the example trigger
(`080c_trigger_adoption_separate_domains_orchestrator_sh`) gives the EXACT
envelope to mirror — same topic, headers, and responses-topic shape every chassis
agent uses. A diagnose request is a kcat publish to `system.agent.generic.requests`
with:

```
headers: correlation_id, request_id, message_id, orchestration_id,
         orchestration_name=diagnose-<ts>, step_name=start, client_id,
         message_type=request, action=orchestrate,
         from_agent_type=user, from_agent_id=cli,
         responses_topic=system.agent.generic.responses
body: {"action":"orchestrate",
       "config":{"agent_type":"diagnose-orchestrator"},
       "input_data":{"symptom":"…","seed_scope":["path/file.go:Symbol"],
                     "runtime_site":"gamesdesign.co.uk","runtime_page":"index"}}
```

This needs NO new triggering code — it is the existing generic-request envelope
with `agent_type: diagnose-orchestrator` and a diagnosis-shaped `input_data`. A
person (or a workflow) sends it; the result comes back on `responses_topic`; a
person reads it. Lowest risk, and the natural first iteration.

**Orchestrator vs long-running agent (your open question).** The guidelines say
**every agent is an orchestrator** and the orchestrator **spawns a dedicated pod**
for the worker (the example: `site-adoption-orchestrator → spawns
site-adoption-agent`). Follow that pattern: a thin `diagnose-orchestrator` receives
the request and spawns a `diagnose-agent` worker pod that runs the loop, exactly as
adoption does. Do NOT stand up a bespoke long-running service — that would
duplicate the chassis's spawn/await/topic machinery the dev guide is built around,
and contradict "every agent is an orchestrator". The agent-chassis worker pod IS
the long-running context for the duration of one diagnosis; it exits when done.

Later triggers — (b) on a build/rebuild failure via the existing
`site_work_items`/`agent_error_log` machinery, (c) a proactive scheduled sweep —
are the SAME envelope sent by a different caller (a post-failure step, the
scheduler). They need no new agent, only a new sender, and are gated on the
real-bug evaluation passing first.

### Contract compliance (the standing rules this must obey)

- **Respond to the caller's responses topic, not your own.** The worker
  (`diagnose-agent`) replies to the orchestrator's responses topic (its parent),
  and the orchestrator replies to the original `responses_topic` in the request
  header — never to its own responses topic. (The standing rule; the envelope
  above carries `responses_topic` for exactly this.)
- **One orchestrator owns the workflow; sub-work is a spawned sub-agent, not a
  sub-workflow in SQL.** If the diagnosis ever needs a research-style sub-step,
  spawn a sub-agent with its own workflow — keep the loop's workflow thin and the
  complexity in the Go engine (the guards stay in `pkg/diagnose`, not re-expressed
  as workflow conditionals).
- **Keep workflow variable names in sync with what the actions expect**, and reuse
  the existing `code`/`storage`/`llm` actions found in §0 rather than re-creating
  them.

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

### The gather changes shape — but it is mostly WIRING, not a port

`gatherer.go` currently shells out to `cmd/bundle` (a subprocess) — fine
standalone, wrong in the chassis (an agent should not shell out to a Go binary).
But STEP ZERO (§0) showed the heavy parts are already chassis actions:
`request_repo_analysis` (analyse), `lookup_code_symbols` (retrieve),
`load_site_for_rebuild` (DB context). So the `ChassisGatherer` implementing the
`Gatherer` interface mostly CALLS THESE EXISTING ACTIONS in sequence; the only new
piece is a **code-context assembler step** (compose the retrieved symbols + schema
+ runtime into the bundle text the verdict reads) and any arbitrary schema/runtime
reads beyond what `load_site_for_rebuild` returns. The engine is unchanged (it
knows only the `Gatherer` interface); the implementation swaps subprocess → a
sequence of existing actions + one new compose step. This is far less than a full
port — the analyser and retrieval, which looked like the hard parts, are done.

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

1. **Build the `ChassisGatherer`** — mostly WIRING existing actions (`§0`):
   `request_repo_analysis` → `lookup_code_symbols` → `load_site_for_rebuild`, then
   the ONE new piece, a code-context assembler step that composes the retrieved
   symbols + schema + runtime into the bundle text. Implements the `Gatherer`
   interface; the engine is unchanged. (The analyser + retrieval, the parts that
   looked hardest, already exist and are live — confirmed in §0.)
2. **Wrap the engine as `diagnose_run`** + `diagnose_emit` actions; register them
   (`Category: "diagnose"`). `diagnose_verdict` = `execute_llm_prompt` with the
   verdict prompt — no new action.
3. **Seed the agent definitions** — a thin `diagnose-orchestrator` (spawns the
   worker) and the `diagnose-agent` worker (the loop workflow + the verdict
   prompt) — as a migration, like other agents.
4. **Wire trigger (a)** — the generic-request envelope from §2 with
   `agent_type: diagnose-orchestrator`; test by sending a known symptom by hand
   (a `trigger-diagnose.sh` mirroring the example trigger's envelope).
5. **The real-bug evaluation gate** (runbook §5): run the live agent on the
   gamesdesign bug and the 016 §9 catalogue; require it to reproduce the
   mid-course REVERSALS and to abstain when the evidence doesn't settle it — not
   just confirm a first guess. Only after this passes is (b)/(c) triggering
   considered.

`pkg/diagnose` (engine + tests) is already in the chassis and unchanged; the work
is the gatherer WIRING (+ one compose step), the thin action wrappers, the two
agent definitions, and the evaluation. The standalone module remains the dev/test
harness.

---

## 7. What stays read-only and human-gated (unchanged in the chassis)

The boundaries the engine enforces do not relax in the chassis: the gather is
read-only (analyser + `\d`/SELECT/log-reads + compose — no writes, no triggered
runs); the verdict only judges and cites; the loop emits a diagnosis + evidence
trail for a human. The chassis adds Kafka/DB/LLM plumbing around that core, not new
authority. A diagnose agent that could change code or trigger a rebuild would be a
different, more dangerous thing — explicitly out of scope.
