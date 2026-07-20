# SUMMARY — bugfix 003 (spawn loses child response) — 2026-07-20

*Written to be read aloud. Current state only; chronology lives in NOTES and
README_where_we_are.*

## What we're trying to do

Stop the platform silently losing work when a parent orchestration spawns a
child agent and waits for its reply. Today a lost reply strands the parent for
30–90 minutes until a database sweep declares it failed — or, in one whole
class of cases, strands it forever. This has been eating image generation, page
builds, business-intel ingestion and even the diagnosis loop itself.

## Where we've come from

Bug 003 was filed on 2026-07-15 with one network cause (certain machines
couldn't reach one Kafka broker) and, added by later threads, a second code
cause: the message loop tells Kafka "handled it" *before* handling it, so any
pod restart destroys whatever was in flight. This thread re-verified all of
that against today's code and live system, and found a third cause and several
wrong assumptions.

## What we've done

- **Found the third root cause.** The timeout-and-retry machinery everyone
  assumed would rescue a lost reply *does exist and is correct* — but it runs
  as a small timer inside the memory of whichever pod sent the request. When
  that pod restarts (which happens on every deploy), every pending timer dies
  with it, and nothing rebuilds them. So the safety net exists only as long as
  nothing goes wrong, which is exactly when you don't need a safety net.
- **Found that the health checks are decorative.** The spawned worker pods DO
  have Kubernetes health probes — but the endpoints they call are **hardcoded
  to answer "OK" unconditionally** (`/health` and `/ready` in
  `cmd/agent-chassis/main.go` return 200 no matter what). A pod that cannot
  reach Kafka at all reports itself healthy indefinitely, so Kubernetes never
  restarts it. The probes were theatre. This is being fixed now: the endpoints
  are being wired to real checks (database reachable, Kafka reachable,
  consumers actually running).
- **Corrected the network story.** The "one bad node, one bad broker" pattern
  from July 15 no longer reproduces; the dial failures are now low-grade,
  intermittent, and hit all three brokers from at least four of five nodes.
  Filed separately as bug 040 — it's a cluster-infrastructure problem, and the
  platform must tolerate it rather than wait for it to be fixed.
- **Shipped the first fix (F1).** The overnight sweep that fails stuck
  orchestrations had a blind spot: one whole status (`EXECUTING_STEP`) was
  never swept, which is how 24 zombie orchestrations — one 53 days old — had
  accumulated. The sweep now covers it (threshold >4 hours, chosen from 7½
  weeks of audit history showing nothing healthy ever ran that long). Applied
  live at 12:43Z today; all 24 zombies drained on the first firing.
- **Verified the owner's point about scheduled-task retries.** There IS retry
  machinery in `scheduled_tasks` — but it operates one layer up (work items),
  retries only work-item-backed flows, takes 70–130 minutes, and redoes whole
  items from scratch. Recorded in the bug file as a nuance, not a refutation.

## Where we are now

F1 is live and verified. The remaining fixes are designed and double-checked
but not yet implemented: F2 (drive retries from the database so they survive
restarts), F3 (commit Kafka offsets only after processing, and record message
completion rather than receipt — these two must ship together), F4 (the honest
health checks, now in progress, plus fail-fast self-restart and graceful
drain). A diagnosis-loop run was filed to independently grade the third root
cause before we asserted it; that run itself appears to have been eaten by the
very bug it was grading — which is not the verdict we asked for, but is
evidence of a kind.

## Where we're going

Implement the health checks now (code inert until an image roll). Then F2+F3
together in one reviewed image roll — that is the change that makes deploys
and crashes survivable, and it unlocks safe continuous deploy later. Then the
rollout-survivability config (probes on the main deployment, drain grace,
second replica). Bug 040 (the network flakiness itself) stays open for an
infrastructure investigation on its own track.
