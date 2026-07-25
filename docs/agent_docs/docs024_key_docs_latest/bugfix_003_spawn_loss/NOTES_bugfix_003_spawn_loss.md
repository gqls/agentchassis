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

**2026-07-20 ~15:30Z. Round 2 VOIDED by bugs_open/019's class, not by a
verdict:** run `e4da4360` died at `review_editquality` with `error` empty; the
real cause was in `collected_data.__step_error` (the 036-documented hiding
place): `execute_llm_prompt: response truncated: stop_reason=max_tokens
(output_tokens=8000)`. The 019 tolerance fix (migration 177) is live but its
Go half is inert until the next image roll. **Misstep (mine): the round-2
submission was ~2× round 1's length** — reviewer output scales with submission
size, and the first seat blew the 8000 cap. Round 3 = same plan, submission
halved, resubmitted on the same trail (run orch `518c399b`). Watcher now also
breaks on `complete_invalid` — round 2's watcher only matched FAILED and sat
silent for its full 45 min on a dead run (Monitor coverage lesson).

**Bug-number collision:** another thread filed a different `040` today
(failed-page-build). Recorded in `bugs_closed/README.md` duplicate-numbers
table + a header note in our file; cite ours as **040-kafka-dial**.

**2026-07-20 ~19:00Z. The health fix WENT LIVE — swept into another session's
commit, before its verdict.** Owner deployed v1.0.1140. Sequence:
`bca5d8255` ("v1.0.1140 - sweep. includes imagegenerator changes,
coordinator.go amend.") took `cmd/agent-chassis/main.go`,
`platform/health/kafka_reachability.go` and the base `deployment.yaml` along
with its own work — the CLAUDE.md landmine, from the other side this time
(my WIP was the passenger). Nothing was lost; forward-only holds.

**Live verification (against pods, not git):**
- chassis pod `agent-chassis-5567d99bd6-5snzn` and spawned Job pod
  `agent-build-dispatch-loop-06bd9258-k5jxq` both return
  `{"kafka_last_ok_seconds_ago":N,"status":"ok"}`; `/ready` → `{"status":"ready"}`.
- Discriminating grep: `kafka_last_ok_seconds_ago` → 1 in the pod binary;
  positive control `^READY$` (the old hardcoded literal) → 0. Both directions.
- Probes attached on the live Deployment (liveness /health 60/15/5/4,
  readiness /ready 10/10/3/3).
- **Fleet restarts: 0 across 44 pods.** The guardian's restart-storm worry has
  not materialised — though nothing has yet been *sustainedly* unreachable, so
  this is absence of the trigger, not proof of the guard [UNMEASURED].

**Round 3 verdict: REVISE** (decided by editquality; abstained 6). Movement:
**reuse_agent flipped to APPROVE** — the refactor answered it — as did
debug_historian, llm_reliability, guidelines, constitution, mission,
diagnosis_guardian. Remaining objections, and what the live system says:
- editquality/edit-3: "do spawned Jobs really inherit?" — **partly right, and
  my wording was loose.** They inherit the *endpoint honesty* from the shared
  image, NOT the probe stanzas (theirs come from the spawner and always
  existed). Verified: the spawned pod above runs v1.0.1140 with probe path
  `/health` and serves real JSON.
- guardian/edit-2: "does an overlay already set probes?" — **no**;
  `grep -rn 'livenessProbe|readinessProbe' deployments/.../agent-chassis/overlays/`
  returns nothing. No conflict.
- editquality/edit-2 + guardian/edit-1: the ctx relocation is asserted, not
  shown. Fair — a sketch limitation, not a code defect; the built binary
  compiles with one root ctx.
- prior_art/edit-1: **this one lands, against me.** See WRONG_CALLS.md
  2026-07-20 — `health.NewServer` has ZERO callers, so the `Checkers`
  "machinery" I claimed to bridge into is itself dead. Placement still right;
  the reuse *claim* was overstated, and I never re-grepped it because it was
  the claim that flattered the plan.

**Consequence for the record: NO `Council-Reviewed` trailer for this change.**
Three rounds, no APPROVED verdict; the trailer is earned, not assumed.

**Also spotted, NOT touched (owner call):** `bca5d8255` committed a **93 MB
`agent-chassis` binary** into the repo root (`git ls-files agent-chassis`
confirms it is tracked). That is another session's commit and possibly
deliberate, so flagged rather than removed.

---

## 2026-07-25 — F2+F3 built, committed `fd122fbec`, migration 205 live; awaiting roll

**What landed.** The full F2+F3 pass from the approved plan
(`~/.claude/plans/bugs-open-003-...md`): migration 205 trio + seed mirrors
(001/007), 8 Go edits (context.go retry_version unconditional; state.go
two-phase dedupe + atomic retry claims; coordinator.go fast-path/ticker funnel;
agent.go + processor.go + git adapter consume ports; consumer.go `Consume()`
DELETED). Quick wins: idle_timeout 600/900 live via the migration;
`terminationGracePeriodSeconds: 60` + preStop committed `171063cb8` (rides the
roll). Migration applied 08:40 UTC, **all 8 VERIFY checks pass** — 159,142
processed_messages rows defaulted to `'complete'`, both partial indexes
present, cleanup function carries the retrying sweep.

**Council:** submission `b896fc22-05d9-4c61-9852-cb1e494de872` dispatched
09:37; run `6bc842f4` was EXECUTING_STEP (review_tooling_provenance) within
minutes — no 29-min queue this time. Committed WITHOUT waiting per the
holding-work-uncommitted-for-a-verdict rule; NO trailer (earned only on
APPROVED).

**Missteps and frictions this session, for the record:**
- **The scratchpad did not survive the session boundary.** The council
  submission JSON, the overlay, and the processor patch splits were all gone at
  resume; rebuilt from the working tree + plan. Anything that must survive a
  session belongs in the repo or the DB, not `/tmp`.
- **processor.go same-file passenger:** the file carried another session's
  uncommitted bugs_open/060 hunk (@1709, calls `discovery.RecordAgentRun` which
  is NOT in HEAD — it would not even compile). Committed via the reverse-apply
  dance: split `git diff` by hunk → `git apply -R` theirs → pathspec commit
  mine → re-apply theirs. Window kept to seconds by building the verification
  overlay BEFORE the dance. Post-restore grep confirmed their WIP back in the
  tree.
- **Pattern-check untouched-twin flag on context.go: verified FALSE ALARM.**
  `FromRequestHeaders`/`FromResponseHeaders` take typed structs and copy
  RetryVersion/RetryCount directly (context.go:359/:401) — the defect only ever
  lived in the string-map `FromHeaders`. Checked before dismissing, not after.
- **Pre-existing test breakage at HEAD** (`orchestration_test.go:171`,
  NewSagaCoordinator signature) — another session's; not touched, per the
  shared-tree rule. Everything else passes; compile of all four real trees
  clean against current HEAD.

**Still owed (plan §E):** image roll of agent-chassis AND git-adapter (bump
IMAGE_TAG), discriminating pod-greps (`RETRY_TICKER_CLAIMED`,
`DEDUPE_CLAIM_LOST`, `MARK_COMPLETE_FAILED`, `DEDUPE_SKIPPED_NO_REQUEST_ID`,
`TIMEOUT_FAST_PATH_CLAIM_FAILED`; `Consume() called` must grep 0), the
mid-orchestration roll test, the child-kill lease test, the leopardess
`ai-readiness-quiz` repro, the liveness restart re-test, and week-later reaper
stats.

## 2026-07-25 (later) — council round 1 REJECTED (guardian veto); D4 wrong call caught; round 2 submitted

**Verdict (corr b896fc22, ~7 min after dispatch — no queue):** REJECTED, hard
veto from guardian. 7 seats approved (reuse, guidelines, tooling_provenance,
diagnosis_guardian, constitution, mission, prior_art), editquality +
debug_historian objected, 6 abstained. Guardian's case: 7-file cross-package
rewrite of the fleet's delivery guarantee = architecture-change territory;
named alternative = split the roll (stage 1: mig 205 + context.go + a
ticker-only retry driver; stage 2: the at-least-once/dedupe rewrite via its
own review). debug_historian's HIGH: the submission stated no post-roll
pod-verification procedure (it exists in PLAN/bug file — a submission gap, not
a work gap). **Images NOT rolled; parked pending the verdict.**

**The veto caught a real error of mine:** the submission claimed "ratified D4:
F2+F3 ship together". D4 verbatim binds **F3's two halves** (offset-commit +
completion-dedupe), not F2-to-F3 — the one-roll scope was the owner's choice,
not a ruling. Logged in WRONG_CALLS.md; memory corrected. Checking the
pre-edit code showed the guardian's stage 1 would in fact FUNCTION — pre-edit
HasProcessedMessage matches retry_version exactly, so a v1 resend passes, and
the caller swallows RecordMessageProcessing's unique-violation error and
processes anyway — but through an accidental fail-open (ERROR log per retry,
dedupe row stale at v0), and it leaves restart-annihilation unfixed.

**Round 2 submitted** (same trail corr, run f995d99b, orchestration name
council-gate-095112): edits unchanged; rationale corrects the D4 claim,
names blast radius per-service (agentbase+messaging link ONLY into
agent-chassis; 13 kafka importers use untouched surfaces; two images change
behaviour), replaces the Consume-deletion grep claim with the compile proof,
quotes the pod-verification + induced-fault procedure verbatim, concedes the
idle_timeout piggyback and the missing pre-DDL dump, and engages the split on
the merits (viable-by-accident, plus the practical fact that the change is
already committed on the shared branch, so a true split now requires a revert
commit — that churn decision goes to the owner if the council still prefers
it). If round 2 rejects: owner decision packet, no roll.

## 2026-07-25 (evening) — kill test run: F2 machinery PASSES, and exposes bugs_open/075

**Test:** chassis pod `7hpxp` deleted 15:25:06 with three organic
`generic-orchestrate-0725-1524` orchestrations in flight (caught by a DB
watcher; monitor tracked transitions).

**PASS evidence for F2/F3:** orphaned request `6779e7a6` expired 15:27:29;
**vet-intel's ticker claimed it at 15:28:01** (`RETRY_TICKER_CLAIMED` — first
live firing of the ticker path, and cross-SERVICE: any surviving coordinator
pod is the rescuer, exactly the design). It drove the adapter-action branch
correctly (deploy_page → re-execute with full payload). The new pod started
clean; grace-60/preStop were live on the deployment.

**What the test exposed (the real find):** the rescued orchestrations then
LOOPED — deploy_page re-executed every ~3 min, git adapter succeeding each
cycle in ~4 s (real GitHub commit each time), response produced, then
DISCARDED by the ownership check (`processing_node` = the dead pod,
coordinator.go:269–277; under at-least-once the discard commits the offset —
permanent). Cap never trips: adapter branch creates a fresh rv=0 request and
does RetryCount[step]=rv+1 (assignment of 1, not increment). Reaper blind:
every cycle refreshes last_activity. Filed **bugs_open/075** (3 root causes,
4 fix candidates, cross-refs the replica-race ANALYSIS doc); pattern added to
016b §9.

**Containment & outcome:** `processing_node=''` for the two loopers + the
frozen `update_status` casualty (`026e9fab`); next cycle's responses applied
and **both orchestrations COMPLETED 15:46:37** — zero work lost. `026e9fab`
left for F1's >4h reaper (nothing re-drives an EXECUTING_STEP orphan; noted
in 075). Also found pre-existing orphans (3 tool-auditor pods, 1 July-13 row)
— left for their workstreams, listed in 075.

**Misstep for the record:** my first read of the loop was "the git adapter is
wedged — possibly MY consume-port regression" (its log silence looked
damning). Wrong twice: the silent pod simply doesn't own the partition (two
replicas, one partition), and per-message logs are Debug-suppressed. The
group-lag check (lag 0) and the ACTIVE pod's logs cleared the adapter in two
queries — check the consumer group before accusing the consumer.
