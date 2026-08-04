# PLAN — `bugs_open/156`, the prevention half: nothing stops a page being saved with content-identical duplicate slots

**Opened 2026-08-04.** Lane takes `156`'s **candidate 1** — a dedup guard at the save
choke point. The detection half is not ours (the `151` lane shipped
`check_content_duplication` + `remove_duplicate_page_sections` on 2026-07-31, deliberately
inert); the data instance was fixed on 2026-07-30. **What was left, and unowned, is that
nothing stops the state being created.**

## Validity, re-measured today before claiming it

Filed 2026-07-30; four days is long enough on this tree that the premise had to be re-taken
rather than inherited.

| claim in the bug file | re-measured 2026-08-04 | verdict |
|---|---|---|
| no unique index on `page_components` beyond the pkey | `\d page_components` — 9 indexes, only `page_components_pkey` unique | **STANDS** |
| no guard in the save compares sections to each other | read all seven refusal/record blocks in `save_page_sections_action.go` — every one compares incoming against **existing rows** or against a floor; none compares the incoming set to itself | **STANDS** |
| `content_hash` never written | `grep -rn "content_hash" --include=*.go` — no writer | **STANDS** |
| 17 duplicate `(page_id, slot_name)` groups, 11 legitimate, 6 content-identical (vonc) | **12 groups today, 0 content-identical.** The 6 vonc ones are gone (fixed 07-30); 11 legitimate remain; 1 is the NULL-`content_data` pair on finetuning.uk | **CHANGED — and it matters** |

The last row is the one to read twice. **Live damage today is nil.** This lane is therefore
not repairing anything; it is closing a gap that produced one measured incident and can
produce another, and the guard must be judged on whether it could ever destroy a live
section — not on how much it cleans up (it cleans up nothing).

## The constraint that shapes the whole design

A unique index on `(page_id, slot_name)` is the obvious fix and is **wrong** — 11 live pages
legitimately repeat a slot with different content. **The discriminator is content identity,
not slot repetition.** This is already a LANDMINE entry with the census query attached.

## The rule we are shipping, and why it is narrower than the bug file's own candidate

The bug file proposes collapsing on `(slot_name, md5(content_data))`. We are **not** shipping
that key, and the correction is the first real decision of this lane:

- **It over-collapses on NULL `content_data`.** finetuning.uk/our-position-on-ai has two rows
  with NULL `content_data` and that shape is live today. Under the file's key every
  NULL-content pair is "identical" whatever their HTML says — so the guard would delete a
  live section on exactly the shape the census already flags as a footnote.
- **It over-collapses on same-content/different-HTML.** `rendered_html` is what serves. Two
  entries carrying the same structured content through different markup are not
  interchangeable.

**Shipped rule:** collapse entry N into an earlier kept entry only when **every value the
INSERT would bind is equal, `position` excluded** — `slot_name`, `rendered_html`,
`component_id` (after the insert loop's parse-else-NULL normalisation) and `content_data`
(after its nil/empty→NULL normalisation). Under that rule the collapsed row would have been
**indistinguishable from its survivor in the database**, so nothing representable can be
lost. `content_brief` is a function of page purpose + slot name, so slot equality subsumes
it; `build_status` is a constant; `position` is renumbered by the loop regardless.

This still catches the whole of the recorded incident: vonc's six pairs matched on
`rendered_html`, `content_data` **and** `component_id`.

## Decisions, with reasons

1. **A sibling file, not surgery on the action.** `save_sections_dedup.go` + test, one call
   site. That is this action's established shape (link repair, claims guard, shrink guard,
   prune floor, metadata source are all sibling files) and it keeps the footprint in a file
   many sessions touch down to ~10 lines.
2. **Placement: immediately after the "sections reaching save" diagnostic, before every
   guard.** Not cosmetic — a doubled list makes four downstream measurements lie:
   the content-regression guard's `newTextLen` is doubled (so a page truly cut to 13% of its
   text reads as 26% and passes the 25% floor), the completeness floor's numerator is doubled
   (a save truncated 6→2 but doubled to 4 scores 67% and passes a 0.5 floor), the claims
   record double-counts, and — the one nobody had noticed — **the locked-slot path
   manufactures a duplicate of locked copy**: the first copy of a locked slot consumes the
   lock and is discarded, the second falls through and is INSERTed beside the locked row.
   Keeping the diagnostic ahead of the collapse preserves the true arrival count in the log.
3. **Parity with the repair: the plan guard.** `remove_duplicate_page_sections` refuses to
   delete a repetition the effective plan source specifies (council trail `da3f2d9b`,
   bug_historian seat, owner decision 2026-07-31). A save-time guard that silently collapses
   what the repair explicitly refuses to delete makes the two halves disagree about the same
   question — the exact drift class this platform's council reviews for. So the guard calls
   the same `datahelpers.PlanSpecifiedSectionCounts`, with the same per-slot accounting, and
   never takes a slot below its planned count. Called **lazily** — only once a duplicate
   group has actually been found — so the normal path costs no query.
4. **Direction of failure is the opposite of the repair's.** The repair fails **closed**: an
   unreadable plan store aborts, because it is about to DELETE. A collapse guard's
   conservative direction is **not collapsing** — on any plan-read error it returns the input
   unchanged and logs. Both mean "do nothing destructive"; they differ because the default
   action differs. Refusing the whole save would add a new failure mode to the fleet's
   busiest save path, which is `bugs_closed/073`'s defect.
5. **It records, it never refuses.** The save proceeds. Rationale: the collapse is by
   construction lossless, so there is nothing for a human to adjudicate before the page
   ships; and the producer is unknown, so a refusal would break builds for a cause nobody can
   yet fix.
6. **The durable record is half the value.** `156` says `[UNRECOVERABLE] the producer of that
   12-entry list is not identifiable from retained data` — `collected_data` is pruned at ~24h
   and the orchestration rows had aged out. So the guard writes `agent_error_log` (the claims
   guard's shape) carrying what that hunt lacked: which extraction path built the list, the
   metadata field and its origin, the step name, the driving work item, and the **adjacency
   signature** — `1,1,2,2,3,3` vs `1,2,3,1,2,3`, the distinction that ruled out the
   concurrent-save race and is the first thing the next investigator will want.
7. **Candidate 3 (`populate content_hash`) is deliberately NOT taken.** Seven Go call sites
   INSERT into `page_components`; only this one would populate the column, so `content_hash
   IS NULL` would read as "not a duplicate" for every row the other six wrote. A Go-marshal
   hash also would not equal a hash of the jsonb `::text` read back, so the column would carry
   a value nothing else can recompute — a third definition of section identity. The honest
   shape, if it is ever wanted, is a DB-side generated column `md5(content_data::text)`,
   which covers all writers at once. Recorded here as a decision, not an omission.

## Where the identity rule lives, and why it does not fork `SectionIdentityKey`

`datahelpers.SectionIdentityKey` is the shared definition used by both the detector and the
repair. Its own contract says the blob must be `content_data::text` **read from the jsonb
column**, and that a Go-remarshalled blob is not canonical. Our guard runs pre-persist, where
`content_data` is still a Go map — so calling it would violate the precondition its comment
spells out, and the architecture seat wrote a scope boundary into that comment in council
round 2.

So the guard keys locally, and the relationship is stated as a **provable subset** rather than
left as two rules that might drift: equal `json.Marshal` output implies identical documents,
which implies identical jsonb text after persist — therefore **everything this guard collapses
the detector would also have flagged**. The guard can only under-collapse relative to the
detector, never over-collapse. A cross-reference goes into `section_text.go` where the next
person to widen that key will read it.

## What "done" means here

- Guard shipped with tests that can fail (each has a named mutation).
- Council gate run on the platform change.
- `bugs_open/156` updated with the corrected key, the plan-guard parity decision and the
  candidate-3 decision.
- **The bug stays OPEN until the fix is LIVE** — the bar is fixed AND live, and a chassis roll
  is somebody else's. Post-roll verification recipe in the RUNBOOK.

## What this lane does NOT do

- It does not find the producer. `156` candidate 4 needs a fresh reproduction; the record this
  guard writes is what makes that tractable the first time it fires.
- It does not touch the `151` lane's inert detector or its enable switch.
- It does not repair anything: there is nothing to repair (0 content-identical groups today).
