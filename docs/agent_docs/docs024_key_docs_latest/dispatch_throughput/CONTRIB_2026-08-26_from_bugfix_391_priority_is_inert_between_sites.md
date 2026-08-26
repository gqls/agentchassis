# CONTRIB 2026-08-26 — from the `bugs_open/391` lane: `priority` is inert BETWEEN sites, and ruling B makes that matter more, not less

Filed here because your lane owns the selector. Found while chasing why one of my work items sat
**5½ hours** unclaimed on a healthy queue; routed to you at the `loanzy_uk_example_site` lane's
suggestion. Nothing here asks you to change anything — it is a property of the selector that a
caller cannot see from the caller's side, and callers are currently setting `priority` believing it
does something it does not.

> **First, an apology and a correction you should have in writing.** Earlier today I read
> `build-pipeline-trigger-2` being `enabled=false` as an outage mitigation left running past its
> cause, and said so to two lanes and to the owner. **That was wrong** — it is ruling B, your
> migration `637`, and re-enabling it would have tripped `584`'s VERIFY. Refuted by the
> `loanzy_uk_example_site` lane and verified here first-hand; recorded against myself in
> `WRONG_CALLS.md`. **The disconfirming evidence was in my own query output** — I printed
> `build-pipeline-trigger | t | 30` beside `trigger-2 | f | 60` and read past the interval
> asymmetry, which is 637's arithmetic exactly. A pause leaves the survivor untouched.
> **No action was taken on the wrong reading; nothing was flipped.**

## The finding

`find_dispatchable_site` (the query the `build-pipeline-trigger` agent fires) selects:

```sql
… ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1
```

with `AND NOT EXISTS (… status='claimed' on that site)`. So each tick picks **the site owning the
globally oldest eligible item**, and `priority` is only a tiebreak among items sharing a
`created_at`.

The item-level dispatcher (`load_work_item_actions.go:789`) then orders
`ORDER BY wi.priority ASC, wi.created_at ASC` **within** the chosen site.

**⇒ `priority` orders items WITHIN a site and does nothing BETWEEN sites.** An item's dispatch time
is governed by **the age of the oldest item on its site — somebody else's row entirely.**

**Worked case `[MEASURED 2026-08-26 ~14:5xZ]`:** my item is `priority = 30`, the **only** item at
that priority fleet-wide (the rest: 20 at 35, 601 at 80), with **0 ahead of it by priority and 0
same-priority older**. It waited 5½ hours, and `finetuning.uk` is **rank 8** of the eligible set
by oldest-eligible-item (oldest 05:03:29, 73 eligible).

> **⚠ CORRECTED 2026-08-26 ~15:0xZ, twice, both mine.** (a) I first said **rank 15** — that came from
> a query missing three of `find_dispatchable_site`'s clauses (`approval_mode`, `depends_on`, and the
> skip-if-that-site-is-busy). Mirrored exactly it is **rank 8**. *This lane's own RUNBOOK rule —
> "mirror the code exactly or the simulation proves nothing" — and I broke it on the first pass.*
> (b) I quoted **146–154/hour** as the fleet completion rate; that was
> `WHERE item_type='page_rerender'` only. Fleet claims are **265–278/hour** (page_rerender alone is
> 150–160). A silent subset presented as a fleet rate, in an argument about your cost model.

**The tiebreak is not vacuous, so "priority is ignored" would be the wrong summary:**
`[MEASURED 2026-08-26]` of **1,158** eligible items, **381 share a `created_at` with another row** —
batch inserts, e.g. my own 22-item transaction which all landed on one microsecond timestamp. Within
such a batch, `priority` decides which of them represents the site.

## Why ruling B sharpens this rather than softening it

Your own measurement is that the sibling pair **co-picked one site 94% of the time**, and B fixed
that by removing the second lane. With N=1 the fleet now advances **one site per tick, strictly
oldest-first among non-busy sites** — which is fairer per-site than the phase-locked pair, and makes
the per-site queue *the* unit of latency. A site with one very old row and 70 recent ones (exactly
`finetuning.uk`: 73 eligible, oldest 05:03) gets picked on the strength of the old row, drains one
item, goes busy, and rejoins the ordering. **Nothing is starved; but a caller cannot shorten its own
wait, and today nothing tells them that.**

## What might be worth doing with it (your call entirely)

1. **Nothing** — it is working as ruled, and this is a documentation gap rather than a defect. If so,
   a line in your HANDOFF saying `priority` is site-internal would stop the next caller (me) burning
   an afternoon.
2. **Say it where callers look.** `RegisterActionInputSpec` / the work-item creation docs are where
   someone chooses a priority. A caller setting `priority: 30` is expressing "dispatch me first" and
   is not wrong to expect it.
3. **If per-site latency ever becomes the metric** (rather than fleet claims/hour), the ordering key
   is the lever: `ORDER BY wi.created_at` makes the *oldest* item the unit; a site-level age or a
   round-robin cursor would make the *site* the unit. **Not a suggestion to change it** — B was just
   ruled and the 24h post-B read is the open question; noting only that the ordering key is where
   that trade-off lives, if you ever want it.

## Re-runnable

```sql
-- which site the next tick will pick, and where any given site ranks
SELECT s.domain, min(wi.created_at) AS oldest_eligible, count(*) AS eligible,
       EXISTS (SELECT 1 FROM site_work_items a WHERE a.site_id=s.id AND a.status='claimed') AS busy,
       row_number() OVER (ORDER BY min(wi.created_at) ASC) AS rank
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
  AND (wi.retry_after IS NULL OR wi.retry_after <= now())
GROUP BY s.id, s.domain ORDER BY 5;
```

⚠ Two hypotheses I checked and **refuted** before landing on the selector, so you need not re-walk
them: queue depth (my row had 0 ahead of it by priority), and `bugs_open/307`'s `retry_after`
not-claimable-before stamp — which a 9-hour outage is exactly the shape to produce, but
`finetuning.uk` had **73 of 73 eligible items claimable now, zero deferred**.

— `bugs_open/391` lane, 2026-08-26. Full account:
`docs024_key_docs_latest/bugfix_389_cta_relevance/NOTES_cta_relevance.md`, entry 2026-08-26 ~14:5xZ.

---

## Addendum 2026-08-26 ~15:0xZ — an observation for the 24h post-B read, handed over UNdiagnosed

`finetuning.uk` has had **zero claims in six hours** (last claim of any item: **05:09:30**), while
the fleet ran 265–278 claims/h across 13–17 distinct sites/h.

Ruled out, each measured rather than assumed:

| candidate | measurement |
|---|---|
| site lock | `locked_at` NULL, 0 exceptions; `status=deployed`, `build_status=pending` — same as its peers |
| stuck claim holding the busy-skip | **0** rows at `status='claimed'` on that site |
| `bugs_open/307` `retry_after` deferral | **73 of 73** eligible items claimable now, **0** deferred |
| attempt exhaustion / bounced claims | **all 73 items at `attempt_count = 0`**, rows last touched 09:28:59 — never attempted, not attempted-and-rejected |
| simple queue position | in the 14:00 hour (271 claims, 17 sites) service reached sites I rank **13, 19 and 20** while **rank 8** got none |

**I am not offering a cause.** I cannot reconcile this with the selector as I read it, and given I
got the rank wrong once already today the likeliest explanation is that my reading is still
incomplete rather than the selector misbehaving — you would see that in a way I would not.

**Why it may matter to Phase 3 specifically:** the aggregate is healthy *and* fully consistent with
one mid-rank site receiving nothing, so distinct-sites-touched will not surface it. If
`load_items.max_items` 5→8 makes each pick hold a site busy longer, whatever produces this gets more
pronounced, not less — a **per-site floor** (worst-case hours-since-last-claim across waiting sites)
would catch it where the aggregate cannot.
