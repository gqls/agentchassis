# NOTES — bugfix 003 (spawn loses child response)

Append-only, newest at the bottom. Missteps are the point.

---

**2026-07-20 (session 1, "bugfix 003").** Research pass over bugs_open/003.
All §3d/§4 code citations re-verified at HEAD — nothing had been fixed by
other threads. Live scale: 70 reaper kills / 2 days (spawn_ingester 27,
spawn_diagnoser 4); 24 EXECUTING_STEP zombies, oldest 2026-05-28.

Found the third root cause: `handleRequestTimeout` is driven only by
`time.Sleep` goroutines spawned at request-send time (coordinator.go:1816,
:2117, :2962); `TimeoutMonitor` (helpers.go) has zero constructors —
dead code; nothing consumes `awaited_requests.status='expired'`. Chassis pod
born 07:35Z same day = restarts are routine timer-killers. Filed the claim to
the diagnosis loop BEFORE writing it into the bug file (corr
`d971e8c2-0c41-4251-b46f-705b471f5dc1`).

**Misstep (caught by owner):** wrote "a lost child response has no rescue path
at all" — true at the awaited-request layer, overstated at the system level.
`claimed-item-timeout` (scheduled_tasks, 120s) DOES retry work-item-backed
flows: evidence-based auto-complete, else reset to `triaged` after 40 min
claim, capped by max_attempts. Verified its full pre_query. Limits: 70–130 min
latency, whole-item redo, nothing for flows without a work item or with their
dispatch loop disabled (`diagnose-pipeline-trigger` enabled=f). Recorded in
bug file as 2026-07-20b nuance.

**Misstep (caught by audit check before applying):** F1 draft said >3h with
"no legitimate step runs 3h" — audit showed one healthy 3.72h stint
(check_health, 2026-06-28). Applied >4h instead. Second draft bug caught in
the same pass: `'...' || current_step` NULLs the whole error string when
current_step is NULL; COALESCE added.

F1 APPLIED LIVE 12:43:26Z; reaper fired 12:43:35Z; all 24 zombies reaped
(verified: 0 rows >4h; 24 rows carry the new error, 11 of them wedged at step
`complete` — worth a look when F2 is built: why do orchestrations wedge at the
completion step?). Mirrored to 020_scheduled_tasks.sql, committed 539768695.

Dial-error evidence for the network correction gathered (12h window, 8 pods):
all three brokers, 4 of 5 nodes, broker-0 dominant, 10–52 errs/pod/12h; filed
as bugs_open/040 (commit 722d84ade). The 2026-07-15 broker-2-only signature
did NOT reproduce.

**Own diagnosis run became a specimen:** the diagnose-agent orchestration for
`d971e8c2` wrote bundle iteration 1 at 11:43Z then wedged at step `route`
(EXECUTING_STEP) — the exact §4.3 class it was filed to grade. No verdict as
of 14:00Z. It will be reaped by the new F1 clause ~15:43Z if still stalled.
A graded verdict on the structural claim is therefore still OUTSTANDING —
the claim rests on direct code reading (all cited functions opened, zero
constructor hits for TimeoutMonitor, zero consumers of status='expired').

Hardcoded-200 health endpoints confirmed (`cmd/agent-chassis/main.go:141–150`)
— probes on spawned Jobs point at them; `platform/health/server.go` has a real
Checkers-based server nobody wired in. Owner directed: fix the health checks
now, summary doc written (SUMMARY_2026-07-20).
