# PLAN — bugfix 247: delete the dead `AgentConfigLoader.LoadAgentConfig` cache, `processRequest`, `selectWorkflowOLD`

Bug: `bugs_open/247_HANDOFF_2026-08-10_dead_agent_config_loader_holds_an_unmutexed_cache_keyed_on_agent_type.md`.
Filed by the `bugfix_239` lane, unowned as of 2026-08-11 (`scripts/who-owns.py 247`: no owning
workstream identified; the only commits touching the file are the original filing and an
unrelated 239 fix that happens to touch the same source file). No workstream directory
existed for it before this one.

## Why this bug, why now

`who-owns.py` scan across the open, recently-filed bug range (240-253) on 2026-08-11 found
247 to be the only one both (a) unowned and (b) independently corroborated as still valid:
`architecture_review/RFC_023` cites `selectWorkflowOLD` as "already one, dead, recorded in
bugs_open/247" while making an unrelated argument — i.e. a second thread independently
re-derived the same dead-code fact after the bug was filed, which is as strong a validity
signal as this estate produces without re-running the diagnosis loop.

This is a **self-evidencing** fix in the sense the CLAUDE.md "diagnosis before debugging"
section describes: the claim is "these three functions have zero callers", which is fully
checkable by grep, and the fix (delete them) either compiles clean or doesn't. It is
**not** a case needing the 090 diagnosis loop — there is no non-obvious mechanism, no
cross-cutting root cause, nothing to diagnose. What it does need, because it touches
`platform/`, is the advisory council gate (see RUNBOOK).

## The fix, as planned (prepared by a fable-model agent, general-purpose subagent, read-only)

Full plan text preserved in `NOTES_dead_agent_config_loader.md` (first entry). Summary:

1. **Rationale.** Two risks, both already named in the bug file: (a) `AgentConfigLoader.cache`
   is a bare unmutexed map — a data race the day it gets a second caller, currently latent
   only because nothing calls it; (b) the whole cluster (`AgentConfigLoader.LoadAgentConfig`,
   `processRequest`, `selectWorkflowOLD`) is a misleading signpost — a session tracing "how
   does the chassis pick a workflow?" can lose real time reasoning about code that never
   runs, which is exactly what happened to the `239` lane before it filed this bug.
2. **Scope, surgically bounded.** `AgentConfigLoader` the *type* is NOT dead — confirmed live
   via `internal/agents/contentcreator/agent.go` (`LoadFromDatabase`, `GetDefaultConfig`).
   Only the unmutexed-cache method `LoadAgentConfig` and its exclusive caller chain die.
   In `processor.go`: `processRequest`, `selectWorkflowOLD` and its private helper cluster
   (`determineWorkflowModeOLD`, `isComplexRequest`, `getDefaultTaskWorkflow`,
   `determineWorkflowMode`), the `configLoader` field + its constructor wiring, and the
   now-unused `config` package import. In `agent_config_loader.go`: `LoadAgentConfig` and
   the `cache` field/init (and optionally the always-empty `db *sql.DB` field — included,
   same class, same struct, compile-verified safe).
3. **What stays untouched.** The live workflow path (`loadAgentDefinition` :361,
   `selectWorkflow` :1107, `getDefaultOrchestrationWorkflow` :1454,
   `convertToWorkflowPlan` :427, `executeWorkflow` :2022), every live `AgentConfigLoader`
   method, and `platform/agentbase/agent.go`'s unrelated `processRequests` (plural).
4. **Tests.** None reference any of the deleted symbols — verified by grep over
   `--include=*_test.go`. No test file changes.
5. **Architecture-scope judgment: not architecture scope.** Per the 2026-07-29 owner ruling,
   architecture review is triggered by a change to what a *shared mechanism guarantees* for
   a *live* caller. This deletion changes zero live callers' observable behaviour (compiled
   output for every live path is provably identical) and has zero consumers to notify — the
   entire evidence base of the fix is that nothing calls it. It reduces shared surface
   (removes a second, hard-coded answer to "what is this agent's workflow?") rather than
   adding one. Still routed through the advisory council gate per the platform-code norm,
   because that norm applies regardless of RFC-scope.
6. **Verification.** Fresh greps immediately before editing (the plan's caller-count checks
   must be re-run at edit time, not trusted from the plan — shared tree), then
   `go build ./...`, `go vet`, `go test ./platform/messaging/... ./platform/config/...
   ./internal/agents/contentcreator/...`, and `go test -race ./platform/messaging/...`.

Out of scope, noted for a follow-up bug rather than widening this commit: `sendSuccessResponseOLD`
(processor.go, currently found near :2097) and `sendErrorResponseOLD` (near :2242) are the
same dead-`OLD`-twin class with zero callers, not named by 247. `LoadFromJSON` on
`AgentConfigLoader` also has zero live callers today but is a harmless one-line wrapper, not
part of the race/signpost defect — left in place.

## Decision log

- 2026-08-11: picked up 247 over the more heavily-trafficked 240s/250s range specifically
  because those are visibly owned by active named lanes (`loanandmortgagecalculator_couk`,
  `staged_component_build`, `bugfix_203_phantom_cta_cleanup`, `bugfix_246_shared_pool_ownership`)
  per `who-owns.py`, and 247 was the one substantive OPEN item in range with no owner.
