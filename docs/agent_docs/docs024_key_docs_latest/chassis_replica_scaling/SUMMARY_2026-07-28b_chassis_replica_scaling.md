# SUMMARY — work-item parallelisation / chassis replica scaling — MILESTONE CLOSE-OUT, 2026-07-28 (end of day)

*The definitive record of the milestone, written at thread close on the owner's
request. Supersedes nothing — the earlier `SUMMARY_2026-07-28` file is the
mid-afternoon state and stays as part of the series; this is the complete one.*

## What we're trying to do

Make the platform able to run many pieces of work at the same time. Until this
morning, every job the fleet dispatched — page builds, re-renders, review
councils, diagnoses — passed through a single one-at-a-time queue in the
chassis: one job executed while everything behind it waited. A slow job held
up every site's work; a stuck one froze dispatch entirely until someone
restarted the pod. The owner's stated destination is thousands of domains,
which needs roughly two-and-a-half orders of magnitude more throughput than
one queue can ever give. The owner's question that opened the thread —
"parallelise per domain? per area of code? some other way?" — was answered:
**per orchestration**, because that is the platform's real unit of state,
ordering, and failure, and everything narrower (the dependency graph between
work items, component write locks, atomic claims) was already enforced at
that grain. Per-site ordering turned out to be three lines of convenience
SQL, not a requirement; per-code-area is not a unit the system records at all.

## Where we've come from

The design — "Kafka delivers, Postgres decides" — was written on 20 July and
sat unbuilt. Its premise: keep Kafka as the letterbox, make the database the
to-do list; consume in milliseconds, execute from a claim-worker pool, and
per-orchestration ordering becomes structural instead of accidental. Two
blockers had been retired from other directions before today: the ownership
discard that made multiple chassis replicas unsafe was replaced by a takeover
(the bug-075 thread), leaving exactly one unsafe reset — deliberately left
for this build to fix first. Prior rulings honoured throughout: Kafka
partition scaling rejected (~125 partitions to reach target rates — wrong
tool); more lanes rejected (lanes multiply serial queues); a second
Deployment rejected (response-history replay per pod). This morning the owner
approved building the whole programme, stage by stage, with a written
read-out at each boundary and chassis rolls gated on a quiet lane.

## What we've done — all of it today, all council-reviewed, all live

**The morning baseline that defined the problem.** The same five cheap page
re-renders that completed 5/5 sequentially (6–8 s each) went **0/5 when fired
concurrently — twice, the second time on a stable pod** — failing at exactly
the 12-minute retry budget. Chasing that produced the day's first deep
mechanism: an await fails when its callee's queue plus the response lane's
delay exceeds the 3-minute timeout, and the retry re-queues **at the back**,
so once the inequality holds it holds forever — a treadmill. The kill shot
was watching the git-adapter answer a request successfully in 3.0 seconds and
the chassis process that success **five seconds after** the await's final
timeout had already failed the job.

**CS-1 — the claim-stealing guard** (commit `afbd005f9`, corr `a45f59af`,
APPROVED). CLAIM_RECOVERY could reset a claim made milliseconds earlier by a
live actor, running a step's side effects twice — the known blocker for any
concurrent response processing. The staleness predicate now lives inside the
reset's own UPDATE (atomic, no read-then-write race); recovery of genuinely
dead claims is untouched. Its first council run was killed thirty seconds in
by another thread's deploy — the "a roll costs an in-flight council" landmine
firing on the very fix being reviewed — and the resubmission passed. Both
branches were later **proven live by induced test**: a fixture claim held
fresh was HELD (14:58:20, still owned by its holder afterwards), a fixture
claim gone stale was RECOVERED (14:58:33, re-claimed by the real pod).

**CS-2 — thin ingest and the claim-worker pool** (commit `28b1a0305` plus
`924df47d0`/`31f2fc1e2`, corr `9f0499b9`, APPROVED on round 3 of an
unusually good council trail). Messages become `chassis_intake_events` rows
in milliseconds; four workers per pod claim serialisation keys through a
claims-table CAS (double-claim structurally unrepresentable) and run the
existing processMessage unchanged; crash recovery is lease expiry, takeover
reset, and the existing dedupe. The council's three rounds each made the
change better: round 1's guardian demanded proof the prerequisites were done
rather than planned; round 2 demanded checkable artifacts rather than
citations — so the mutable-state audit now lives in the database
(`doc_notes`, subject_key `mutable-state-audit-cs2-9f0499b9`), the blast
radius is measured (exactly one binary compiles agentbase; 181 agent
definitions run through it), the backpressure cap got a real measurement
(5,000 ≈ four days of total fleet dispatch), and the DB pool ceiling the
audit itself uncovered (hard-coded 4 connections) became a knob. Enablement
was two-step (dark roll, verify, then flag) and the decisive structural
proof came at once: four intake events **started within 0.9 seconds of each
other** — four orchestrations genuinely executing simultaneously — with the
fifth queuing exactly until a worker freed.

**CS-3a — no more response replay** (commit `f4d24252f`, corr `f4e425dc`,
APPROVED). Found while verifying CS-2: every chassis restart made the
response lane re-read the ENTIRE reply history under its fresh per-pod
group — measured at ~49 messages/minute against a 12,000-deep topic, i.e.
**two to three hours of response deafness per restart, growing daily**. By
its shape and numbers this was most of what bug 029 had been calling the
"post-roll degraded window". One additive constructor
(`NewConsumerFromLatest`) and an env gate later, a fresh pod's response
group is born at the head of the topic — verified live: LAG 0 at birth, and
the day's final burst ran clean on a pod barely two minutes old, inside what
used to be the dead zone. One honest wrong call on the way, logged in
WRONG_CALLS: the flag was first flipped while the running binary predated
the code — a silent no-op that cost one restart — despite the pre-flip
pod-grep gate having been written down twice that same hour. The gate was
applied properly ever after, and caught real trouble again later (a Go
const turns out to be a vacuous grep marker; log-message literals are the
real discriminators).

**Phase 3 — replies through the pool** (config flip of code approved inside
CS-2). Verified by response-kind intake events flowing and an end-to-end
control completing. With both queues fixed, the **council-in-its-own-pod
wrapper (Candidate 4) completed its first attempt** — the same path that
failed "timed out after 3 retries" the day before and roughly half of all
historical attempts. The non-response class it was blocked on appears to have
been the response-loss family all along; the flip of the default stays with
bug 096's owner after more runs.

**Phase 4 — two replicas** (owner-consented mid-afternoon). The intake layer
changed the maths since the plan was written: the claims CAS already
coordinates workers across pods, and the intake table's uniqueness
constraint deduplicates the response broadcast between per-pod groups — so
scaling was one `kubectl scale`, reversible, no new Go. The originally
planned shared-response-group change is now an efficiency refinement, not a
correctness prerequisite. Side effect worth saying twice: **a deploy no
longer freezes dispatch** — one pod keeps serving while the other restarts.

**CS-2d — the pool's first operational defect, found by operating it**
(commit `369f486f3`, APPROVED round 4 under the CS-2 correlation). Ninety
minutes after enablement, saturation testing exposed a hole: work abandoned
by a dying pod sat `running` forever, because the candidate scan only looked
at `pending` — so the takeover reset that recovers dead holders' work was
unreachable precisely for the case it existed to serve. Two real dispatches
from 13:53 were stuck exactly that way. A one-line widen (live holders stay
excluded by the lease check) fixed it, and the proof was the casualties
themselves: **both orphans re-ran and completed at 16:35 — two and a half
hours late instead of never** — recovered cross-pod, with exactly two
takeover-reset log lines to show for it.

**The git-adapter queues same-site deploys** (owner-ruled; commit
`7dc876795`, corr `bf2bef0a`, APPROVED). The last burst failure standing was
GitHub's 422 "not a fast forward" when concurrent commits to the shared
sites repository raced the ref update. The fix makes GitHub's own CAS the
queue: the loser re-reads the winner's head and rebuilds its commit on top,
up to four attempts; blobs are content-addressed and created once; real
errors still fail loudly and immediately. Proven on both branches: 16/16
concurrent deploys live, and — since no natural race fired — an induced
fake-GitHub test that returns the verbatim 422 while moving the branch head,
asserting the retry rebuilds on the winner's base. This also absorbs a human
push landing in the same window, half of bug 120's collision class.

**Handed to their owners, in writing**: the treadmill mechanism and the
replay finding to bug 029 (whose own author then retired the spawn-race
framing on the strength of it); the fan-out requirement with sizing numbers
to the dispatch-gate lane (their `LIMIT 1`-site-per-tick is now the fleet's
throughput ceiling); the watchdog predicate to the dispatch-queue
workstream (key on the orchestration row, so it survives the cutover — and
note a 4-hour reaper does exist, which is bookkeeping, not alerting); the
Candidate 4 result to bug 096; the machine half of bug 120 closed with the
workflow half explicitly left open.

## Where we are now

Live configuration: chassis `v1.0.1190`, **two replicas**, four workers
each, `CHASSIS_INTAKE_MODE=worker_pool_all`, `CHASSIS_DB_MAX_OPEN_CONNS=12`,
`CHASSIS_RESPONSES_START_AT=latest`; git-adapter `v1.0.1187` on both
replicas. Every change council-approved (six submissions, ten rounds, five
approvals — one review died on the fleet's AI budget cap and passed on
resubmission when the owner restored it), every deploy verified against the
running pod, every failing branch exercised — induced where nature declined
to cooperate.

The scoreboard, one line: **nought-of-five in twelve minutes at breakfast;
five-of-five in twenty seconds, on two pods, by tea** — and ten-at-once
completed everything that genuinely reached the platform. Operationally the
fleet gained three things beyond throughput: deploys no longer stop the
world; the "things fail for twenty minutes after a deploy" folklore has a
measured mechanism and a shipped fix; and abandoned work is now recovered
rather than silently lost.

## Where we're going

Nothing on this programme blocks anything. Remaining, in priority order for
whoever picks it up:
1. **Dispatch-gate fan-out** (their lane, requirement + sizing delivered):
   release N sites per tick — the current fleet ceiling.
2. **Adapter tier**: the eight sequential-by-design adapters are the next
   serialisation layer; the await-timeout-versus-queue-depth interaction is
   documented in bug 029 for a deliberate decision (longer timeouts vs
   smarter retry).
3. **Efficiency polish on this programme**: the shared response group (cuts
   duplicate fetch at replicas>1), worker-count tuning gated on the
   still-unmeasured orchestration-state write amplification, the hygiene
   list (dead consumer groups, lane retirement decision, dropping the dead
   `orchestration_requests` table).
4. **Candidate 4 flip** (bug 096's owner): more wrapper runs now that
   councils return real verdicts, then the default can move.

---

### The record, precisely (for the ledger)

| Deliverable | Commit(s) | Council corr | Live as |
|---|---|---|---|
| CS-1 staleness guard | `afbd005f9` (+test repair `2da1c5f50`, `d0cda1e39`) | `a45f59af…` APPROVED | chassis ≥ v1.0.1184 |
| CS-2 intake + pool (+2b pool knob, +2c backpressure) | `28b1a0305`, `924df47d0`, `31f2fc1e2` | `9f0499b9…` APPROVED r3 | flag on since ~10:56Z |
| CS-2d orphan recovery | `369f486f3` | `9f0499b9…` APPROVED r4 | chassis v1.0.1190 |
| CS-3a seed-to-latest | `f4d24252f` | `f4e425dc…` APPROVED | flag on since ~11:23Z |
| Phase 3 responses-via-pool | (within CS-2) | (within CS-2) | flag on since ~14:56Z |
| Phase 4 two replicas | config only | owner consent | since ~15:08Z |
| Git-adapter ref-CAS retry | `7dc876795` (+induced test `1602dcd95`) | `bf2bef0a…` APPROVED | git-adapter v1.0.1187 |
| Migration | `249_chassis_intake_events.sql` applied + recorded | — | clients_db |
| Audit artifact | `doc_notes` id `701bce70…` | — | queryable |

Key documents: this directory's NOTES (the technical log, including every
correction and both wrong calls), `README_where_we_are.md` (the plain-prose
history), `WRONG_CALLS.md` (the env-flip-before-binary entry),
`bugs_open/029` and `/120` and `/096` (the contributions), and
`bugfix_029_dispatch_gate/CONTRIB_2026-07-28_fan_out_requirement_from_parallelisation.md`
(the handoff).
