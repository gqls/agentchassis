# PLAN — bugs_open/326, making the front door re-usable

Design, phasing, decisions **and their reasons**. Corrections to the originating
brief live here, marked as corrections — never silently edited away.

## The brief, and the correction it needs

`bugs_open/326` says re-submitting a domain reports COMPLETED and queues nothing,
and blames `create_work_item` deduping on `item_key` "in ANY status".

> **CORRECTION 2026-08-23 — the filed root cause is wrong, and fixing what it names
> would fix nothing.** `idx_swi_dedup` excludes terminal statuses, `complete` and
> `cancelled` among them; `writeWorkItem`'s `ON CONFLICT` names that same predicate.
> The real mechanism is the anti-churn brake ABOVE the insert. Evidence, queries and
> the negative control: `NOTES_326_retry_the_front_door.md`. The correction is also
> written into `bugs_open/326` itself, and the loanzy.uk lane (which filed it) has
> recorded the matching correction in its own docs rather than having it forked here.

## The decision that shaped everything else

**Do not fix the five build-chain steps. Fix the door they go through.**

The brake was designed for a DETECTED DEFECT and is wrong for an ACTION REQUEST.
The estate already knows this — `bugs_closed/024` established it and
`work_item_recurrence_test.go` still states it — and already has the lever
(`recurrenceExpected`). 024 fixed the Go call sites it touched and stopped there,
and two years later the same defect took a customer build.

So the question was not "which steps need the flag" but **"why did the right
conclusion fail to propagate, and what stops that happening a third time?"** Three
answers, and the plan is one edit for each:

1. **The door could lose work at all.** Both arms now defer instead of dropping or
   burying. This protects the 36 Go call sites nobody will reclassify and every
   author who forgets the flag — a weaker guarantee than the flag gives, but
   unconditional.
2. **The outcome was unreadable.** `deduped:true` meant two opposite things. It now
   means one, and `deferred`/`retry_after`/`prior_attempts` say the rest.
3. **Adoption was unmeasured.** A census makes an undeclared classification a
   standing finding, so the next author does not have to remember.

## Decisions, with reasons

**Deferral over deletion of the brake.** The brake does real work — it stops a
detector/fixer churn loop. The defect was never that it slows a key down; it was
that it slowed it down by destroying the request. `retry_after` expresses "not yet"
without expressing "never", and it is live machinery: one renderer
(`workItemRetryNotPendingSQL`), honoured at the claim, the dispatch loader and the
completion verifier. Nothing new was invented.

**`maxBackoffMinutes` (720) reused, not a new constant.** Its own comment records
that it is bounded to stay inside the stale-reaper's 48h ceiling on a triaged row.
A fresh constant would carry the number without the reason, and the next person to
"tidy" it would not know what they were breaking. A test pins the relationship.

**The within-cycle deferral is the window REMAINDER, not a flat interval.** An
arrival at 2h59m waits a minute, so the total quiet period per key is exactly what
it always was. A flat 3h would have silently tripled it for late arrivals; three
boundary cases pin it.

**Kill switch armed, not opt-in.** The owner has ruled against default-OFF switches
that rot unexercised. `DISABLE_ANTI_CHURN_DEFERRAL` is a redeploy-free disarm lever
back to exactly the former behaviour, and it is tested on both arms — an untested
lever is a lever nobody can safely pull.

**`on_dedup` opt-in with the unsafe default OFF.** New authority on a shared seam
ships as a field a reviewer of the caller can see, not a doc comment (owner ruling
2026-08-02 §2). Per RFC_022 that shape is not architecture-scope, and the consumers
were enumerated rather than asserted: zero live definitions branch on this action's
output.

**Migration scope narrowed to five steps (owner ruling, 2026-08-23).** A draft swept
13 and found, in doing so, exactly why not to: `claims-auditor.request_claims_review`
NEEDS the counter, because its revalidator-close loop writes `complete` into the
two-strike window by design. The other 16 keyed steps go to their own lanes via the
census, named.

**Rejected: the bug's candidate 1, "put the attempt in the key."** It defeats the one
arm that works, and that arm is the bug's own required negative control — a
concurrent duplicate submission must still dedup. Two operators submitting one domain
at once would otherwise each mint a key nobody holds and both cascades would
interleave over one site.

**Not fixed, deliberately:** the 635 existing two-strike `unresolved` rows (draining
that landfill is RFC_010 / `bugs_open/033` D2's open owner decision, and overruling it
from inside this patch is what the failure-ladder header warns against); the other 16
undeclared steps; `bugs_open/327` (the trigger script's own silence); `bugs_open/333`'s
policy door, which edits the same function and deserves its own round.

## Phasing, and why this order

1. **Migration 572** — config, live on apply, closes the customer path **today**
   against the current binary. Not redundant with the Go half: without it a
   re-submission would be *deferred up to 3h* rather than dropped, which is better
   and still wrong for a retry.
2. **Go + tests** — council submission, commit, `make build-agent-chassis` (which
   builds committed HEAD, so commit first).
3. **Roll**, then verify at the artefact.
4. **Migration 573 (`_HOLD`)** by hand, after the roll. This one is a genuine
   ordering constraint and the file names the mechanism: StrictConfig plus
   per-message `ValidateWorkflow` means applying it early fails *every*
   domain-submitter run, not once.

No ordering constraint is claimed for the Go half. HEAD is shared and any session's
roll ships it (owner ruling 2026-07-29 retired that exemption); pretending otherwise
would be a claim I cannot support.
