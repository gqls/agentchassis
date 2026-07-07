# HANDOFF — Diagnosis→Fix Loop → fresh context

STATUS: DISCUSSION PHASE — this file updates as the originating chat's
critical discussion takes positions; at cutover it is the complete start
state. (Travelling-docs pattern: runbook = plan, notes = history, this =
the session-level travelling doc.)

## 1. Read first
RUNBOOK_diagnosis_fix_loop.md — the task statement (written for a newcomer),
what exists, the F0–F3 plan, boundaries, and the open questions Q-A…Q-H with
whatever positions the discussion has recorded by the time you read this.
Then NOTES_running_fixloop.md for the discussion log and decisions.

## 2. File manifest for the new project folder

MUST-HAVE (working docs):
- RUNBOOK_diagnosis_fix_loop.md, NOTES_running_fixloop.md, this handoff.
- RUNBOOK_code_retrieval_route.md — the closed diagnosis-route record (what
  the loop does today, its guards, its incidents).
- RUNBOOK_31_.md — contextkit build (slicing, bundle assembly, cmd/bundle).
- RUNBOOK_design_diagnosis_loop_7_.md — the scaffold-testing lineage + the
  founding retrieval/abandonment insights.
- Guidelines 000_documentation_index, 001_development_guide,
  003_contracts_and_standards (the guidelines agent's own corpus, later).
- RUNBOOK_builder_route.md — the pipeline map (guardian seed material) +
  boundary awareness.
- The tools chat's runbook + running notes (arriving next turn) — Q-F
  coordination.

CODE CONTEXT — generate with contextkit itself (dogfooding; regenerate the
analysis first per RUNBOOK_31, or reuse a recent /tmp/analysis_repo.json):

```bash
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "F0 of the diagnosis→fix loop: make each iteration's bundle durably fetchable and add per-task running notes written per iteration/step; design the documented intake route. Understand how the loop assembles, emits and completes today before proposing the persistence/egress point." \
  -scope pkg/diagnose/loop.go \
  -scope pkg/diagnose/step.go \
  -scope pkg/diagnose/advance.go \
  -scope pkg/diagnose/verdict_wire.go \
  -scope pkg/diagnose/sqlguard.go \
  -scope platform/orchestration/actions/diagnose_route_action.go \
  -scope platform/orchestration/actions/diagnose_assemble_bundle_action.go \
  -scope platform/orchestration/actions/diagnose_emit_action.go \
  -scope platform/orchestration/actions/diagnose_load_runtime_action.go \
  -scope platform/orchestration/actions/workflow_actions.go:CompleteWorkflowAction \
  -include platform/orchestration/actions/registry.go \
  > BUNDLE_fixloop_F0.md
```
Adjust -scope paths if the diagnose package lives elsewhere in the tree;
add the tools chat's persist_diagnosis_note action file once their docs land.
(The uploaded prototype sources — loop/step/gatherer/docselect/verdict_wire/
main — are the /mnt/project mirror; the chassis copies are authoritative.)

ALSO CARRY: example_bundle.txt (the invocation pattern), and the diagnose
trigger script pattern from RUNBOOK_code_retrieval_route.md §7E.

## 3. Inherited gotchas (diagnose-relevant subset)
- Loop core is READ-ONLY by contract; sqlguard allowlists reads — keep it so;
  the F1 write surface is a separate agent with isolated credentials (the
  spawn token-gate pattern exists in spawn_actions.go).
- Result contracts: result_from/output_fields both live post-Option-A; the
  response size guard exists (max_response_bytes) — bundle egress via
  completion payloads is therefore BOUNDED; don't ship megabytes, persist
  and reference.
- agent-type (hyphen) pod label; is_active gates spawn; seeds must copy image
  columns from a live donor; check body.status not just header status.
- Schema before SQL; snapshot_agent before agent_definitions updates; 0 rows
  isn't decisive until the query is checked; explicit git refs, never HEAD.

## 4. Opening move for the new chat
Paste the constitution + this trio + the manifest docs + BUNDLE_fixloop_F0.md.
First action: settle Q-A/Q-B from the recorded discussion positions, then
slice F0.1 with pre-registered success criteria.
