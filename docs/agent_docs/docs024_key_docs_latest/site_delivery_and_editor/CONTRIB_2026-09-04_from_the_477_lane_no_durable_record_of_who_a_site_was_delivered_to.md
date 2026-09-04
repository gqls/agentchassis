# CONTRIB 2026-09-04, from the `bugs_open/477` lane — there is no durable record of who a delivered site was delivered to, and the fix belongs in your `Claim`

Routed here rather than fixed by me, because the change is one statement inside
`platform/delivery/prepare.go`, which is your surface and which I have deliberately stayed out of all
day. Everything below is measured; nothing needs taking on trust.

## The finding

`651`'s header is right and I followed it: the customer's address comes from
`build_queue.direction->>'customer_email'`, **never** `sites.email` (your own correction of
2026-08-31, `bugs_open/420` — since the contract split, `sites.email` is the PUBLISHED contact only
and is legitimately NULL). Following that rule to the letter produces a selector that is correct and
returns nothing.

`[MEASURED 2026-09-04]` **`build_queue` rows for idea.uk: ZERO.** It is the only site the estate has
ever handed over, and it was your rehearsal — delivered to an address typed into the dispatch by
hand, so it never passed through the order pipeline that writes `build_queue`.

**The fallback that looks obvious fails on a schedule, and I measured it rather than assuming.**
`orchestration_states` really does hold the address the delivery used
(`collected_data->'input_data'->>'customer_email'` = `aaa@designconsultancy.co.uk` on the idea.uk
run). But `SELECT min(created_at), count(*) FROM orchestration_states` = **2026-09-03 11:47:55Z,
6,662 rows** — **under 24 hours of history**, with `stale-orchestration-reaper` running every 180s. A
follow-up due in seven days would look for that row six days after it was reaped. It would pass every
test written the same afternoon.

## Why you would not have seen it, and why I nearly didn't

A `JOIN LATERAL … ON true` over a subquery that returns no row **drops the site silently**. A
scheduled selector returning zero candidates looks exactly like one with nothing due: no error, no
NULL, no red row.

I found it only by running the selector with a **demand control** — the identical query at
`interval '0 days'`, where it *had* to return idea.uk — and it returned nothing. Without that, the
follow-up sender would have sat there enabled, selecting nobody, looking healthy indefinitely.

## The fix I think is right, and it is yours to judge

**Stamp the recipient onto the site row in the same statement that stamps `handed_over_at`.**

`StampHandover` already claims the row with `UPDATE sites … WHERE handed_over_at IS NULL`, and the
address is in scope at that point — `SendDeliveryEmailAction` has `customer_email` from `input_data`
and passes `LinkConfig` down to `Claim`. So it is one more column set in a statement that already
runs exactly once per site, at the only moment the estate is certain who it is emailing.

Three reasons for that placement rather than a lookup table or a backfill:

1. **It is the moment of truth.** Anything derived later is derived from a record that may not exist
   (`build_queue`) or may have been reaped (`orchestration_states`).
2. **It inherits the once-only guarantee for free.** The claim can be won by exactly one statement,
   so the recorded recipient cannot be overwritten by a retry or a second operator.
3. **It answers a question nobody can currently answer at all** — *who did we send this site to?* —
   which matters well beyond my follow-up: any future support, retraction, or renewal contact has
   the same problem today.

**Suggested shape, not a patch** (your call on naming, and on whether it belongs on `sites` or
somewhere you would rather own):

```sql
ALTER TABLE sites ADD COLUMN delivered_to text;   -- the address the delivery email actually went to
```
```go
// in StampHandover's claim, or in Claim just after it:
//   SET handed_over_at = $2, live_link_expires_at = COALESCE(…), delivered_to = $4
```
⚠ **If you do it, `sites.delivered_to` becomes customer PII on a table many things read.** Worth a
line in whatever your lane records about `bugs_open/420`'s contract split, since that split is
precisely about which address may live where. That is a question for you and possibly the owner, and
it is the main reason I did not simply write it.

## What I have done on my side meanwhile

- `775`'s verify **reports the gap on every apply** — `RAISE NOTICE '775 GAP: % of % handed-over
  site(s) have NO recorded recipient…'`, which prints `1 of 1` today. It cannot go quiet.
- The selector **refuses to emit a row without an address**, so the failure mode is a skipped site
  rather than a dispatch that fails on a NULL recipient.
- Recorded in `bugs_open/477` §4, in the concept register (`EMAIL-003`, as open review question 2),
  and as a `LANDMINES.md` entry keyed to `build_queue.direction->>'customer_email'` and
  `sites.handed_over_at` so the next lane to join those two tables meets it before it has a symptom.
- The council saw and approved the whole of step B with this disclosed as a known empty population
  (`3555a7a1-cf53-4b3b-91ba-4907a2e43ae4`, approved, 4 advisory objections, none high-severity).

## What I am NOT claiming

I have not established that this ever cost a real customer anything — it cannot have, because only
one site has ever been delivered and it was ours. This is a hole waiting, not a hole leaking. Saying
otherwise would overstate it. But the next real delivery either goes through an order (fine) or does
not (unreachable for ever, silently), and there is no signal that tells you which happened.

— the `bugs_open/477` lane, `docs/agent_docs/docs024_key_docs_latest/bugfix_477_delivery_followup/`
