# SUMMARY — 2026-08-24: closed, because the sorter was watched doing it on a real note

## What we're trying to do

Stop a queue from marking true findings as resolved. `required_fields_missing` items report a
page component missing text its own schema declares as required. A router sorts each one:
convert to a repair job, park for a human, or close. Every finding must end somewhere a person
or a pipeline can act on — and none must end up *looking* handled when it is not.

## Where we've come from

`bugs_closed/277` built that router in August with a sensible lookup: find the component among
those already published. That mirrored the only producer at the time. `bugs_closed/342` then
added a second producer whose stated purpose is reaching components that are *never* published
— so every finding it wrote was about something the router had been told not to look at, and
the router closed them all as imaginary. The 342 lane caught it on its last day, filed
`bugs_open/367`, and closed itself.

Yesterday we fixed it: not by widening the lookup (measurement showed that buys a crash, and
aims at a rebuild path where 28 of 31 jobs already fail), but by changing the rule — **a
disposer may close only on positive evidence of absence.** Everything else parks.

## What we've done since

The fix was proven at the route and against the whole population, but never *watched*. Today
closed that gap.

A new chassis build rolled at 15:39Z. The fix is database config, so a roll cannot carry or drop
it — but a re-run of the original seed would silently revert it, and that seed's own checks
would not notice. So the first action was the one-command `_VERIFY` sidecar built yesterday for
exactly this: three controls, all green.

Then the awkward finding: **nothing had happened.** No item of this type had been filed since
the fix, because the render-time producer only fires on a section edit. The closure criterion
could have sat unmet indefinitely while looking like patience. So we drove it — and it was a
repair, not a demonstration. The item the bug wrongly closed yesterday was re-checked (the
fields are still empty, the component still unpublished, untouched since 17 July), found to be
a false negative sitting in the "actioned" bucket, and re-opened.

The router picked it up in about 100 seconds and parked it.

## Where we are now

**`bugs_closed/367`.** Fixed, live, survived the roll, and observed on the real item. The two
orchestration records are each other's control — same item, same component, one day apart:

| when | route | target state | component | html |
|---|---|---|---|---|
| 08-23 17:09Z | `stale` → closed `complete` | — | *unresolved* | 0 |
| 08-24 16:08Z | `target_not_dispatchable` → parked | `pending` | `0a1498b3…` | 9,220 |

The parked row **holds its dedup key**, where the close had released it — the anti-churn
property as data rather than intent. The canary that drove it is a committed, guarded file: it
refuses if the finding is no longer true, if the key is taken, or if the item is not closed.

Four other documents that still called 367 a live open defect have been retracted, dated, with
the corrections marked rather than silently edited away.

## Where we're going

Nothing in this lane needs picking up. Two things are named and belong elsewhere:

**`bugs_open/375`** — of the three code paths that stamp work `complete`, the one this router
uses never consults the verifier framework built to stop false completions. That is *why* the
silent close was reachable, it is architecture-scope, and it is filed rather than smuggled into
a bug patch. Its blast radius is explicitly not established; the query to run is in the file.

**`bugs_open/333`** (owned by the 277 lane) — until owned-page routing is fixed, this
population can only ever park. The findings are visible and honest; they are not repaired, and
no document in this lane says otherwise.

One residual we could not close: the original seed `410` is a whole-config upsert whose verify
block never asserts the resolution predicate, so a hand re-run silently reverts this and passes
its own checks. It carries a header pointer, which is mitigation rather than a control. The
honest fix belongs to whoever next touches that file.
