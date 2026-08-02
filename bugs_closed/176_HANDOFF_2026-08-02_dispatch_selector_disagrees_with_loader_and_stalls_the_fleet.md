# 176 — the dispatch selector and the item loader disagree about "dispatchable"; one blocked row stalls the whole fleet, silently

**Filed AND closed 2026-08-02** (session "bugfix 19"). Found while verifying
`bugs_closed/154`'s fix; fixed by `sql_for_agents/285`, live and proven at the
artefact the same hour. Filed as a closed case because the mechanism is durable
and non-obvious, not because it needed a queue.

Workstream docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_154_work_item_routing_columns/`
(NOTES + README, 2026-08-02 entries). Register: **WDS-002**. Landmine appended
(footprint `find_dispatchable_site` / `LoadWorkItemsAction` / `depends_on`).

## Symptom

Zero work items claimed anywhere in the fleet for **89 minutes** (08:06→09:36Z),
and again for **68 minutes** (09:41→10:14Z), while:

- `build-pipeline-trigger` fired every 120s as configured,
- `build-dispatch-loop` ran 16 times in 40 minutes,
- **every one of those runs finished `COMPLETED` with `error` NULL**,
- and 366 work items sat `triaged` across 17 unlocked sites.

Nothing logged a failure. There is no error state anywhere in this bug.

## Root cause

Two queries decide "is this dispatchable" and they are not the same query.

**SELECTOR** — `build-pipeline-trigger.find_dispatchable_site` — picks the SITE:

```
status IN ('triaged','approved') AND attempt_count < max_attempts
AND no 'claimed' item on the site
```

**LOADER** — `LoadWorkItemsAction`, `platform/orchestration/actions/load_work_item_actions.go`
(query built ~line 624), inside `build-dispatch-loop` — picks the ITEMS on that
site. Same three, **plus two**:

```
AND (COALESCE(approval_mode,'auto') = 'auto' OR status = 'approved')
AND (depends_on IS NULL OR NOT EXISTS (unresolved dependency))
```

So the selector can hand the loop a site whose only "eligible" item the loader
refuses. The loop loads **zero** items, reports `has_items:false` with
**`rows_dropped:0`**, notifies the scheduler, and completes successfully having
done nothing. **No claim is made**, so the `NOT EXISTS ... status='claimed'`
clause never excludes the site — it is still eligible next tick, and is picked
again. The queue does not advance and nothing records a failure.

`rows_dropped:0` is the discriminator: it does not mean "nothing was dropped
after loading", it means the SQL never matched a row that the other query calls
eligible.

## The blocking row

```
93f2a3b7  content_rewrite  triaged  page-build-handler  priority 110
          site robot-hands.com  created 2026-07-31 12:27:28  attempt 0
          depends_on = {0733a7a4}

0733a7a4  needs_content_page  status = needs_human_review   <- never clears
```

`needs_human_review` is terminal for automation: nothing moves it to
`complete`/`verified` without a person, so `93f2a3b7` is permanently unloadable.
Fleet-wide it was **1 blocked row out of 366** selector-eligible items across 17
sites — and it happened to be at the head of the queue, so it stalled everything
behind it.

## Interaction with `sql_for_agents/284` — read this before changing either

284 (same day, owner-directed) replaced the selector's lowest-UUID ordering with
oldest-waiting-first. **That change is correct and stays. It also converts this
defect from intermittent to permanent**, and the two are only safe together:

- Under lowest-UUID, a blocked site held the head only while it happened to sort
  lowest, and released it when a lower-UUID site gained work.
- Under oldest-waiting-first, the key is `created_at` — it never changes and only
  ages — so **an unloadable item, once at the head, is at the head for ever.**

Anyone reasoning about either migration alone will get its blast radius wrong.

Timeline, which shows both regimes failing the same way:

```
08:03–08:06  robot-hands drains its last 5 CLAIMABLE items
08:06–09:36  ZERO claims fleet-wide (89 min) — pre-284 lowest-UUID selector picks
             robot-hands every tick; every dispatch loop loads 0
09:36        284 live → gamesdesign (starved 3d10h) served, 5 claims
09:41–10:14  ZERO claims again (68 min) — robot-hands back at the head, by AGE,
             and now permanently
10:14        285 live
10:16–10:22  relojistas 5, vetcomparison 2, webdesign 1 — exact FIFO order
```

## Fix

`sql_for_agents/285` — give the selector the loader's two clauses verbatim, so
"this site has dispatchable work" means the same thing in both places. Config
only; no image dependency; live on commit. Pre-flight count 1 (md5-guarded),
`snapshot_agent`, `UPDATE 1`, verified.

**Safety argument is structural, not statistical: strictly narrowing.** Every
site this removes is a site where the loop would have loaded 0 items and claimed
nothing. No site that would have been served loses service; the only behaviour
removed is the wasted pick that blocks everyone behind it.

## How it was verified

At the loop's own output, not at run status — every run was `COMPLETED`
throughout, before and after, which is exactly why status is useless here:

```sql
SELECT initial_request_data->'input_data'->>'domain',
       collected_data->'load_items'->>'item_count'
FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
ORDER BY created_at DESC LIMIT 8;
```

```
robot-hands.com   0   10:03, 10:06, 10:08, 10:11, 10:13   <- pre-285
relojistas.com    5   10:16                                <- post-285
vetcomparison.uk  5   10:19
webdesign.co.uk   1   10:21
```

**Negative control**, chosen so that a fix which "worked" by merely dispatching
more could not pass it: `93f2a3b7` must remain `triaged`, unclaimed,
`attempt_count 0`. It is genuinely undispatchable, and the point of the fix is
that the queue stops **waiting** for it — not that it gets dispatched. Confirmed
after the fix.

## Left open deliberately

- **The loader's dependency subquery is site-scoped** (`WHERE site_id = $1`), so a
  `depends_on` pointing at another site's item can never resolve and blocks its
  item for ever. 285 copies this faithfully on purpose — the selector's job is to
  AGREE with the loader, not to be independently correct. Fixing the cross-site
  case means changing the Go loader, and both queries must then move together.
- **`93f2a3b7` itself.** `0733a7a4` needs a human; that is work-item triage, not a
  dispatch defect. It unblocks the moment that item reaches complete/verified.
- **The selector still ignores `sites.locked_at`** while the scheduler's
  `pre_query` gates on it, and the pre_query is `triaged`-only while the selector
  takes `triaged|approved`. Currently inert — measured 2026-08-02: all 366
  eligible items were `triaged`/`build` on unlocked sites, so neither asymmetry
  can bite today. `[UNMEASURED]` whether it ever has.
- The loader's optional `item_pipeline`/`handler_agent` filters are **not**
  mirrored into the selector: the live `load_items` config carries neither, so
  mirroring would invent a filter the loader does not apply and re-introduce the
  dead `domain='build'` clause that migration 067 removed.

## Transferable pattern (also added to 016b §9)

**When two queries must agree about the same predicate, the one you are reading
is not necessarily the one that decided.** A dispatcher, a claimer, a validator
and a detector each carry their own copy of "is this eligible", and drift between
them produces no error — it produces silent no-ops that look like idleness. The
tell is a component that runs on schedule, completes cleanly, and changes
nothing. Diagnose from what the component *decided* (its `collected_data`), never
from its status.

Corollary logged in `WRONG_CALLS.md`: I twice recorded these stalls as
"comparable to known behaviour, not yet outside it". A range assembled from
unexplained instances of the same symptom is not a baseline, and a second
occurrence is evidence a thing is systematic — the opposite of the reassurance I
read into it.
