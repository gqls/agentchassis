# RFC 001 — at-least-once delivery, completion-time dedupe, DB-driven retry

**Status: RATIFIED (keep-live), owner 2026-07-25 — against the live evidence
in §7. Moves to IMPLEMENTED when the induced-fault campaign and week-later
stats land.**

An honest preamble: this RFC documents a redesign that is **already running in
production**, not one awaiting a green light. The code was built and committed
(`fd122fbec`, 2026-07-25) while the council gate reviewed it as a point fix;
the gate's guardian vetoed twice, holding that a change of this shape needs an
architecture-review track — which did not exist. The owner agreed and created
the track (see `PROCESS_architecture_review.md`); before the guardian's
contained alternative could be executed, another session's routine fleet build
(v1.0.1159, deployed 10:36 UTC) carried the committed code live — the shared
branch working exactly as documented ("committed code rides anyone's next
build"). So this, the track's first RFC, is written from the unusual position
of having live evidence instead of a proposal. It exists so the decision is
made deliberately even though the deployment was not.

## 1. Problem + evidence

`bugs_open/003`: parents strand 30–90 min (or forever, pre-F1) when a spawned
child's request or response is lost. ~68 reaper kills per 2 days; 365 expired
awaited requests with `retry_version=0` (never retried once) at design time.
Three mechanisms, each verified at HEAD with file:line in the bug file:

- **(a)** `kafka.Consumer.Consume()` committed the offset BEFORE returning the
  message — at-most-once wearing an at-least-once comment. Pod death
  annihilated whatever was in flight.
- **(b)** dedupe recorded RECEIPT, not COMPLETION — a redelivered message
  whose first copy died mid-work read as a duplicate.
- **(c)** timeout retry lived only in process-local `time.Sleep` goroutines;
  nothing consumed `awaited_requests.status='expired'`.

Plus two premise defects that made retries structurally unable to heal:
`retry_version` parsed only for responses in `FromHeaders`; and
`RecordMessageProcessing`'s ON CONFLICT named the PK while the real UNIQUE
constraint excludes `retry_version`. Full detail: bug file, "2026-07-25 build
record".

## 2. Design (as implemented, `fd122fbec`)

- **Consume**: `FetchMessage` → process → `CommitMessages`, commit
  unconditional after `processMessage` returns (ratified D3 — parent drives
  retry; a poisoned message must not head-of-line-block). `Consume()` deleted.
- **Dedupe**: two-phase — claim `'processing'` with a lease
  (`PROCESSED_MESSAGES_LEASE_SECONDS`, default 900) via
  `ON CONFLICT ON CONSTRAINT processed_messages_unique`, `MarkMessageComplete`
  as a defer. Duplicate ⇔ `retry_version >= incoming` AND (`'complete'` OR a
  live lease held by a DIFFERENT pod). Empty `request_id` is loud
  (`DEDUPE_SKIPPED_NO_REQUEST_ID`), not a silent bypass.
- **Retry**: both the fast-path timer and the 1-min ticker funnel through
  atomic claims (`'expired'→'retrying'`, `FOR UPDATE … SKIP LOCKED`, joined on
  `orchestration_states.status='AWAITING_RESPONSES'`, 60-min window, 5-min
  stale-claim reclaim) — exactly one actor drives any expiry, across pods and
  restarts. `retry_version>=3` marks `'error'` BEFORE routing (wedge-proof).
- **Schema** (migration 205, applied 08:40 UTC, 8/8 VERIFY): `'retrying'`
  status; `processed_messages.status` DEFAULT `'complete'` +
  `lease_expires_at` — the DEFAULT preserves old-binary semantics, which is
  what makes image-first rollback safe.

## 3. Alternatives considered

- **Guardian round 1 — split: retry driver first, consume/dedupe later.**
  Verified mechanically viable at `fd122fbec^`: pre-edit `HasProcessedMessage`
  matches `retry_version` exactly, and the caller swallows the resulting
  unique-violation error and processes anyway. Rejected because it works via
  an accidental fail-open (ERROR log per legitimate retry, dedupe row stale at
  v0) and leaves mechanism (a) unfixed — recovery waits a full step timeout
  instead of immediate redelivery.
- **Guardian round 2 — ship only the `context.go` header fix.** Same
  fail-open dependency, narrower still; everything else deferred to this
  track. This was the owner's chosen path (2026-07-25) — overtaken by the
  ride-along deploy before it could be executed.
- **A standalone retry service.** Rejected: re-driving a retry requires the
  coordinator's own resend machinery (execCtx rebuild, produce, re-arm); a
  sidecar duplicates orchestrator internals — the reuse seat would rightly
  object.
- **Feature-flagging the consume order.** Rejected: an env-gated choice
  between two commit orderings doubles the code paths in the exact function
  whose simplicity is the safety property.

## 4. Blast radius, named (mechanically derived)

`go list -deps` over all 19 `cmd/` targets: `platform/agentbase` +
`platform/messaging` link into **agent-chassis only**. 13 services import
`platform/kafka` for Producer/FetchMessage/CommitMessages — surfaces
untouched. `Consume()` has **zero remaining callers including test files**
(repo-wide grep; `go vet`, which compiles `_test.go`, exits 0). Behaviour
changes in exactly two binaries: **agent-chassis** and **git-adapter**.

## 5. Rollout — actual state, and the staged plan had it been deliberate

Actual: migration first (by design), then both binaries at once via the
v1.0.1159 ride-along (not by design). The deliberate plan this RFC would have
carried: stage A git-adapter (small canary), stage B agent-chassis, induced
faults at each stage. That order remains the template for future RFCs.

## 6. Rollback plan

Redeploy the previous image (v1.0.1156-class). The schema tolerates old
binaries in both directions: old code inserts get `status='complete'`
(receipt-time semantics preserved); `'retrying'` rows are swept to
`'cancelled'` by the 60-min cleanup sweep; `205_…_ROLLBACK.sql` exists if the
schema must follow (image first, schema second). No data migration required
either way.

## 7. Acceptance evidence

Deploy-verification (DONE, on the running pods): created literals present
(`RETRY_TICKER_CLAIMED`, `DEDUPE_CLAIM_LOST`, `MARK_COMPLETE_FAILED` — 3/3),
removed literal `Consume() called` greps 0.

First ~4.5 h live (10:36–15:05 UTC, 2026-07-25):
- 175 awaited requests created; **19 retried by the new driver; 7 recovered
  end-to-end to `processed`** — each a certain silent loss under the old code;
- 6 exhausted the retry cap and were marked `'error'` loudly (previously:
  silent 90-min strandings);
- **0 rows stuck `'retrying'`; 0 expired leases; 806/806 dedupe claims
  completed**; no panics; only 1 request expired in the trailing 90 min
  (vs a steady multi-per-hour bleed before).

Still owed (retires the RFC to IMPLEMENTED):
- induced-fault campaign: mid-orchestration chassis-pod delete
  (`retry_version` must increment via a surviving/new pod ~1–2 min after
  `timeout_at`); child kill mid-handler (lease-expiry redelivery, exactly one
  applied response);
- leopardess `ai-readiness-quiz` end-to-end repro (§6 of the bug file);
- liveness restart re-test (outstanding from F4);
- week-later stats: reaper `stale AWAITING_RESPONSES` ≈ 0; expired-never-
  retried population stops growing.

## 8. Open follow-ons (not blocking)

- Guarantee `request_id` on every inbound path (this design makes absence
  loud; it does not fix it).
- Delete `RunSimpleNotUsed` outright (editquality, twice).
- Config tweaks ship separately from migrations (conceded process rule).
- Response consumer-group race blocks replicas≥2 — separate design
  (`ANALYSIS_chassis_response_consumer_group_race.md`).

## Decision record

- 2026-07-25, owner: architecture-review track created; guardian path chosen
  (revert to context.go-only + review) — **overtaken by events** (v1.0.1159
  ride-along deploy) before execution.
- 2026-07-25, owner: **KEEP LIVE — RFC ratified** against the §7 evidence
  (19 retries / 7 end-to-end recoveries / 0 stuck claims in the first 4.5 h).
  No `Council-Reviewed` trailer applies anywhere in this arc — the gate never
  approved; the trailer is earned, not assumed. The induced-fault campaign
  runs as this RFC's acceptance evidence.
