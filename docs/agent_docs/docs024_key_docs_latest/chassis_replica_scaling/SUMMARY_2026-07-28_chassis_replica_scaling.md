# SUMMARY — chassis replica scaling / work-item parallelisation, 2026-07-28

**What we're trying to do.** Make the platform able to run many pieces of work
at the same time. Until today every job the fleet dispatched went through a
single file-of-one queue in the chassis: one job executed while everything
behind it waited, and a slow job — a nine-minute review council, a stuck
deploy — held up every site's work at once. The goal the owner set is
thousands of domains, which needs hundreds of times more throughput than a
single queue can give.

**Where we've come from.** The design was written on 20 July ("Kafka
delivers, Postgres decides") and sat unbuilt for a week while its two
blocking questions were answered from other directions: the response-discard
that made two chassis copies unsafe was fixed by the bug-075 thread, and the
one remaining unsafe reset was left, deliberately, for this build to do
first. This morning the owner chose the parallelisation axis — per
orchestration, not per domain or per code area — and approved building the
whole programme.

**What we've done (all of it today, all council-approved, all live).** A
safety guard so two workers can never double-process one reply (CS-1). The
main change: messages become database rows on arrival and a pool of four
workers per pod executes them, with per-job ordering enforced by an atomic
claims table (CS-2, plus a backpressure cap and a connection-pool knob the
review demanded, plus a same-day fix — CS-2d — for an orphaned-work hole we
found by operating it). Replies now flow through the same pool (Phase 3).
The chassis runs two replicas (Phase 4, owner-consented) — which also means
a deploy no longer freezes dispatch, because one pod keeps serving while the
other restarts. Two adjacent defects found and fixed along the way: every
chassis restart used to replay the entire reply history — two to three hours
of deafness that explains most of the fleet's "things fail after a deploy"
folklore — now started at the head instead; and the git adapter now queues
same-site deploys at GitHub's own consistency check instead of failing the
loser (owner-ruled).

**Where we are now.** The same five-page test that failed nought-out-of-five
with twelve minutes of retrying this morning runs five-for-five in twenty
seconds, across two pods; ten-at-once completed everything that reached the
platform. Work that a dying pod abandons is recovered and re-run rather than
lost — proven live with two real orphans this afternoon. The remaining known
ceilings, in order: the dispatch gate still releases one site per tick (the
requirement and sizing are handed to its owning lane), the adapters are
one-at-a-time services, and the await-timeout-versus-queue-depth treadmill
mechanism is documented in bug 029 for its owner.

**Where we're going.** Nothing is owed to the programme itself except
efficiency refinements: the shared response group (CS-3 proper, cuts
duplicate fetch at replicas>1), worker-count tuning gated on the
write-amplification measurement, and the hygiene list (dead consumer groups,
retiring the extra lanes, dropping the dead intake-shaped table). The next
real throughput step belongs to the dispatch-gate lane (release N sites per
tick); after that, the adapter tier. The wrapper that runs councils in their
own pods completed its first post-fix attempt, so the council-off-the-lane
idea is live again for its owner to finish.
