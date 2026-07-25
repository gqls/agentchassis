# PLAN — bugfix 003 (spawn loses child response)

Started 2026-07-20 ("bugfix 003" thread). The full evidence base and fix plan
live in `bugs_open/003_HANDOFF_spawn_lost_child_response.md` (2026-07-20
research pass) — this file records the *decisions and their reasons*, and any
corrections to the plan as it meets reality. Point at the bug file, don't fork
it.

## Decisions

- **D1 (2026-07-20): F1 threshold >4h, not the drafted >3h.**
  `orchestration_state_audit` (7.5 weeks) showed exactly one healthy
  EXECUTING_STEP stint over 3h (3.72h, a `check_health` step, 2026-06-28) and
  none over 4h. 4h has zero historical false positives and still drains every
  zombie (all were 18h+). APPLIED LIVE 12:43Z, verified: 24 reaped on first
  firing.
- **D2 (2026-07-20): F2 drives retries from the DB, not from re-armed
  in-memory timers.** Rebuilding sleeping goroutines on startup would keep the
  guarantee process-local (dies again on the next restart mid-window). The
  per-pod 1-minute `cleanupExpiredAwaitedRequests` ticker already exists in
  every chassis pod; extending it with an atomic claim
  (`status='expired'→'retrying'`) makes any surviving pod the rescuer.
- **D3 (2026-07-20): F3's offset commit happens after `processMessage`
  returns, unconditionally** (the server.go semantics, not client.go's
  no-commit-on-error). In-process handler errors already route through
  `handleProcessingError`; the guarantee being bought is against pod death.
  Not committing on handler error would poison-loop deterministic failures.
- **D4 (2026-07-20): F3 ships only together with the completion-time dedupe**
  (two-phase claim on `processed_messages`). Either alone is a no-op or worse
  — 003 §4.4a-bis explains why; do not let these separate in review.
- **D5 (2026-07-20): network layer split out** as `bugs_open/040` — the
  2026-07-15 single-route signature no longer reproduces; platform fixes must
  tolerate flakes rather than wait on infra.

## Corrections to the originating brief (bugs_open/003)

- §3b's broker-2/one-node claim: superseded 2026-07-20 (see 040). Corrected in
  place in the bug file.
- §4.1's "add a liveness probe": probes already exist on spawned Jobs; the
  endpoints are hardcoded 200s (`cmd/agent-chassis/main.go:141–150`). The fix
  is honest endpoints, not new probes. Corrected in place.
- "No rescue path at all" (research pass, first version): overstated at the
  system level — the owner pointed out `scheduled_tasks` retry machinery;
  verified: `claimed-item-timeout` retries *work-item-backed* flows at 70–130
  min latency, whole-item redo, finite attempts. The awaited-request layer has
  none. Nuance recorded in the bug file 2026-07-20b.

## Phasing

1. ✅ F1 reaper EXECUTING_STEP clause — config, applied + verified 2026-07-20.
2. ▶ F4 (health part): honest `/health` + `/ready` + consumer reachability
   signal + chassis Deployment probe stanzas. Go + kustomize; inert until
   image roll. Council-gate before commit.
3. F2 + F3 together, one image roll, council-gated. Includes the
   `processed_messages` migration and the `awaited_requests` status CHECK
   addition (`'retrying'`).
4. F4 (rollout part): preStop, grace 60s, replicas 2, idle_timeout_seconds
   config for diagnose-agent/image-generator.
5. Verification per the bug file's list (deliberate mid-orchestration roll,
   §6 repro, discriminating pod-grep, week-later reaper stats).

---

> **CORRECTED 2026-07-25 (build session):** three premises in the F2/F3 design
> above were wrong and would have made retries a no-op if shipped as written —
> (1) `FromHeaders` parsed `retry_version` only for responses, so a resent
> request arrived at the child as v0 and was dedupe-dropped; (2)
> `RecordMessageProcessing`'s ON CONFLICT named the PK (includes
> retry_version), so a v1 insert errored instead of taking over — the claim
> targets `ON CONFLICT ON CONSTRAINT processed_messages_unique`; (3)
> `GetAwaitedRequest` admitted `'waiting'` only and expiry stamps
> `processed_at`, so a late response to a retried request was
> DUPLICATE_SKIPPED — claims set `processed_at=NULL`, getter widened to
> `('waiting','retrying')`. Full detail: bug file 2026-07-25 build record.
> **Also: the migration is 205, not "180/199"** — numbering raced twice with
> other workstreams between design and build. Built & committed `fd122fbec`;
> migration LIVE; awaiting the agent-chassis + git-adapter roll.
