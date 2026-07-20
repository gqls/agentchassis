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

**2026-07-20 (later, same session).** Health-check fix BUILT:
`cmd/agent-chassis/health.go` (healthState + broker prober; liveness fails only
after 300s continuous all-broker unreachability, env-tunable; readiness =
started + broker reachable within ~63s; boot grace via lastKafkaOK seeded at
start), main.go wired, probe stanzas added to the chassis base deployment.
`go build` + `gofmt` clean. Council-submitted with FORCE=1 (files are cmd/ +
deployments/, outside the ^(platform|internal|pkg)/ scope regex — code, not
docs, so the override is the intended use). SUBMISSION_CORR
`3a18a1a4-f3f1-429f-933c-2ef9e2e1833b`; verdict pending at time of writing.
INERT until image roll either way.

F2/F3 premise double-checks (owner asked before go-ahead) — ALL PASS:
- `cleanupExpiredAwaitedRequests` ticker IS started (`coordinator.go:109`) —
  unlike TimeoutMonitor, this extension point is live.
- Live `awaited_requests` CHECK = waiting/processing/processed/expired/
  cancelled/error — **no 'retrying'**; F2 needs migration `180` (next free
  number in `docs/agent_docs/sql_for_agents/`; watch bugs_open/007's
  migrations-ledger trap).
- `processMessage` returns void; success branch at agent.go:849 is where
  `MarkMessageComplete` goes.
- `handleProcessingError` sends an error RESPONSE to the parent → in-process
  failures already retry via the parent's handleRecoverableError; so F3's
  commit must be unconditional after processMessage returns (redelivering
  locally would double-execute and fight the parent's retry). D3 confirmed.
- Diagnosis-loop verdict on the third root cause: NOT obtained — the run
  wedged at `route` (EXECUTING_STEP, 140 min stale, bundle iter 1 only),
  itself a 003-class casualty; will be reaped by the new F1 clause. Intake
  item closed as cancelled with a note. Claim rests on direct code reading.
- Owner's "scheduled_tasks has retry machinery" point: verified & reconciled
  (see 2026-07-20b nuance in the bug file) — work-item layer only.

**2026-07-20 (later still).** Council round 1 on the health fix: **REVISE**
(reuse_agent objection decisive; guardian + debug_historian + prior_art also
objected; 8 abstained). The objections were fair and improved the change:

- **Misstep (mine): built the prober as a chassis-local file** when
  `platform/health` already existed as the natural shared home — exactly the
  reuse-vs-build discipline the seat enforces. Round 2 moves it to
  `platform/health/kafka_reachability.go`, bridged to the existing
  `Checkers` machinery via `Checker()`; chassis-local health.go deleted.
- **Misstep (mine): round-1 sketch under-showed the ctx move** — editquality
  read it as a second unlinked context. The runbook literally warns "the
  sketch is the only view of your code reviewers get"; re-learned it anyway.
- The survey the reuse seat asked for found the pattern is WORSE than 003
  recorded: git adapter (:855), browserrunner (:189), thunder, analyser,
  webscrape, imagegenerator, auth-service ALL serve hardcoded health JSON.
  All named as follow-up adopters in round 2.
- Guardian's blast-radius check (their own SQL): 165 agent types, 6 pipelines
  on this image. Restart-storm answer firmed up: liveness restarts are
  IN-PLACE kubelet restarts with capped exponential backoff — not evictions,
  PDBs inapplicable, no scheduling churn.
- debug_historian: post-rollout verification now an explicit plan step —
  curl /health expecting `kafka_last_ok_seconds_ago` (old binary returns bare
  "OK" — a discriminating literal) + strings-grep with positive control.

Round 2 resubmitted on the same trail (`RESUBMIT_CORR=3a18a1a4`), run orch
`e4da4360`. No FORCE needed this round — the main file is now `platform/`.
Build + gofmt + vet clean after the refactor. Verdict pending.
