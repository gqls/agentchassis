# Register — new:resilience-self-heal

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

3 concepts, consolidated from 2 raw extractions across units U12, plus 1 added post-freeze (RSH-003, 2026-07-28).

### RSH-001 — Dual-signal self-heal on missing spec dependency
- **status:** deployed
- **status-evidence:** "validate_composition_inputs both loud-logs AND queues a recovery work item on miss... the two-strike rule marks the item unresolved."
- **what:** A general resilience pattern for a Go action depending on a spec aspect that may not yet exist: emit a loud error log AND queue a recovery work item in the same failure path — the log is a durable dashboard signal, and the queued item is a genuine self-heal mechanism, since if it later runs successfully the originally-dependent item auto-redispatches. Repeated failures of the same recovery item accumulate via a two-strike rule into a terminal `unresolved` state rather than retrying forever.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"1. Status Summary", #"5. Decisions Made This Session" (docs024_archives unit)
- **relations:** site-design-planner Choice B scope; work-item two-strike/wont_fix pattern; composition resolver orphan-rows policy (RSH-002, same composition subsystem)
- **verify-later:** `validate_composition_inputs_action.go` implementation; two-strike rule location in the dispatch loop

### RSH-002 — Composition resolver orphan-rows policy
- **status:** aspirational
- **status-evidence:** "If install fails, those rows become orphans... we extend the existing database-cleanup scheduled task to sweep them. Draft SQL in draft_composition_orphan_cleanup.sql."
- **what:** Because the palette/typography_set resolvers each commit in their own transaction before `install_site_composition` runs, a failed install leaves orphaned rows behind. The accepted design tolerates this — low-cost orphans are allowed to occur and are swept up periodically by an extension to the existing `database-cleanup` scheduled task, rather than adding cross-resolver rollback/transaction coordination.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Orphan policy" (docs024_archives unit)
- **relations:** composition resolution architecture; dual-signal self-heal (RSH-001, same source unit/subsystem)
- **verify-later:** `database-cleanup` scheduled task pre_query — confirm the orphan-sweep CTE was actually merged in

### RSH-003 — Retry-as-replay: an awaited request is re-sent, never rebuilt
- **status:** deployed (chassis v1.0.1193, migration 263; landed 2026-07-28)
- **status-evidence:** `awaited_requests.request_payload jsonb` live on clients_db; `RETRY_PAYLOAD_UNAVAILABLE`/`RETRY_SELF_ADDRESSED`/`MISROUTED_REQUEST` present in the v1.0.1193 binary and the deleted `is_retry` stub absent from it (present in v1.0.1192 — positive control).
- **what:** The platform-wide contract for retrying a timed-out awaited request. A sending action (`spawn_agent`, `call_agent`, and `spawn_remote_agent` via the same helper) records the exact `producer.Produce` arguments under the reserved result key `types.RetryPayloadKey`; `createAwaitedRequest` lifts it onto the awaited request and `InsertAwaitedRequest` persists it to `awaited_requests.request_payload` (deliberately `json:"-"` on the struct so it never enters the hot `orchestration_states.awaited_requests` JSONB). On timeout `handleRecoverableError` **replays** that message via `RetryPayload.ReplayRequest`, which re-stamps only `retry_version`, `message_id` and `timestamp` and regenerates the Kafka headers from the replayed message so headers and body cannot disagree. **The invariant: a retry differs from the original in those three fields and nothing else.** With no recorded payload the coordinator REFUSES to retry rather than synthesising one. **The landmine this exists to prevent:** the previous code rebuilt the retry from the *awaiting* orchestration's own state, so every retry carried the PARENT's `orchestration_id`, an empty body and `Action:"execute"` — the receiver resolved the parent's row, saw `AWAITING_RESPONSES`, declined the work and logged success (`bugs_closed/129`). Measured before the fix: 430 of 430 retried requests in 14 days took that path, 294 exhausted the budget.
- **open review question:** any NEW action that sends an awaited request must set `types.RetryPayloadKey` in its result or it gets **no retries at all** — a deliberate loud failure, countable via `SELECT step_name, count(*) FROM awaited_requests WHERE request_payload IS NULL AND sent_at > now() - interval '2 hours' GROUP BY 1`. Covered today because `call_agent`/`spawn_agent` are the only awaited senders seeded fleet-wide; that is a measured fact about `agent_definitions`, not a structural guarantee, so re-measure before assuming it still holds.
- **sources:** `platform/orchestration/types/retry_payload.go`; `platform/orchestration/coordinator.go` (`handleRecoverableError`, `handleOrchestrationStatus`, `createAwaitedRequest`); `platform/orchestration/state.go` (`awaitedRequestColumns`, `InsertAwaitedRequest`); `docs/agent_docs/sql_for_agents/263_awaited_requests_request_payload.sql`; `bugs_closed/129_HANDOFF_2026-07-28_spawned_child_adopts_parents_orchestration_row_and_silently_declines_the_work.md`
- **relations:** the awaited-request timeout driver (`bugs_closed/003` F2 fast path + durable ticker) is what *calls* this; the adapter re-execute branch in `handleRecoverableError` is the deliberately-untouched alternative (rejected for agents because re-executing `spawn_agent` spawns a second pod — the `bugs_open/124` double-dispatch class); spawn→call handshake failures (`agent_error_log` "timed out after 3 retries")
- **verify-later:** whether the `MISROUTED_REQUEST` child-side guard ever fires in practice; if it does, some sender other than the two covered ones is setting identity wrongly
