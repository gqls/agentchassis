# 316 — the news-feed cap serves the alphabet, not the schedule: ranks 1–5 are 0% late, ranks 6–9 are always late, and the queue is 2.1× oversubscribed

**Filed 2026-08-19** by the `bugfix_275_silent_row_caps` lane. **Fourth instance of `bugs_open/275`'s
class, and the first one where the remedy is ORDERING, not payload bounding.** It also **refines a
claim this lane wrote into register LCO-009**: that a work-queue cap means *"coverage is eventual, not
a defect"*. Coverage is indeed eventual. It is also **systematically unfair, and the unfairness lands
hardest on the site that asked to be refreshed most often.**

## The defect

`content-feed-trigger.find_news_sites` ends:

```sql
... WHERE <site is news-feed enabled AND has a deployed page AND has a source due now>
ORDER BY s.domain LIMIT 5
```

`ORDER BY s.domain` — **alphabetical, and stable**. The runs are 6-hourly. So each run takes the first
five *alphabetically* of whatever is due, and the same names win every time they are in contention.

## Measured 2026-08-19 09:03Z (live `clients_db`)

**Every run hits the cap.** Over the whole retained window of `orchestration_states`, `find_news_sites`
returned **exactly 5 of 5 on five consecutive runs** (08-18 08:32, 14:32, 20:32; 08-19 02:33, 08:33).
The sibling `model-directory-trigger` returned 4 against a cap of 12 on all four of its runs in the
same window — the negative control, so this is not an artefact of how the census counts.

**The lateness boundary falls exactly at the cap boundary.** Each site's own configured
`content_sources.fetch_interval` is the yardstick:

| alpha rank | domain | its cadence | overdue by | % of its OWN cycle |
|---|---|---|---|---|
| 1 | ai-agent-orchestration.com | 6 h | — | **0** |
| 2 | dartsonline.com | 4 h | — | **0** |
| 3 | fundamentallyai.com | 6 h | — | **0** |
| 4 | gaswholesalers.com | 6 h | — | **0** |
| 5 | mortgagecalculator.co.uk | 6 h | — | **0** |
| 6 | **relojistas.com** | **3 h** | **3 h 30 m** | **117%** |
| 7 | robot-hands.com | 6 h | 29 m | 8% |
| 8 | vetcomparison.uk | 6 h | 26 m | 7% |
| 9 | webdesign.co.uk | 6 h | 24 m | 7% |

Five sites at zero, four sites late, and the split is **precisely** ranks 1–5 versus 6–9. Nothing about
these sites differs except their initial letter. Last fetch times say the same thing twice over: ranks
1–5 were all served by the 08:33Z run (08:34–08:42), ranks 6–9 by the 02:33Z run (02:36–02:42).

**`relojistas.com` is the worst-hit because it asked for the most.** A 3-hour cadence means it comes due
twice per 6-hourly window; sitting one place past the cut, it waits for a window where four of the five
alphabetically-earlier sites happen not to be due. By the time the 14:32Z run reaches it, it will be
~9 hours late on a 3-hour schedule.

## ⚠ The cap is not the whole story — the queue is genuinely undersized

Do not fix only the ordering and expect everyone to be on time:

| | per day |
|---|---|
| site-fetches **demanded** by the configured cadences (Σ 86400/interval) | **42** |
| slots **supplied** (4 runs × cap 5) | **20** |
| **oversubscription** | **2.10×** |

**Removing the cap entirely does not close it either:** 4 runs × 9 eligible sites = 36 slots against 42
demanded, still 1.17×. Roughly 9–10 site-fetches come due per 6-hour window against 5 slots, which is
exactly the alternation the fetch timestamps show — top five, then bottom four.

So there are two separable defects, and they want different fixes:

1. **Unfairness** — who absorbs the shortfall is decided by the alphabet. Fix with ordering.
2. **Capacity** — there is a real shortfall regardless of who absorbs it. Fix with the cap size or the
   run frequency, and only after deciding whether the configured cadences are what anyone actually wants.

## Fix candidates, ordered by what closes the door

1. **`ORDER BY min_next_fetch_at NULLS FIRST` (oldest-due first) instead of `ORDER BY s.domain`.** This
   is the standard work-queue ordering and it makes the cap harmless *as a fairness matter*: whoever has
   waited longest goes first, so lateness spreads evenly instead of concentrating on four fixed names.
   One config change, DB-live on apply, no roll. ⚠ The current query has no `next_fetch_at` in scope —
   it lives in the `EXISTS` subquery — so this needs the source join lifted, not just the ORDER BY
   swapped. **The sibling `model-directory-trigger` already uses `ORDER BY random()`**, which is the
   cheap version of the same idea and is why it has never shown this pattern.
2. **Then size the queue deliberately** — raise the cap to ≥10 (covers a full window's due set), or run
   more often, or lengthen the cadences if 3-hourly was aspirational. **This is an owner decision, not a
   defect fix**: it trades LLM/ingest spend against freshness, and the arithmetic above is the input.
3. **Do NOT just raise the cap and leave the ordering.** A bigger alphabetical cut still starves the
   same tail the moment demand exceeds supply again — and demand grows with every new news-feed site.

## How to verify a fix

- **The disconfirming pair is already established**, which is what makes this cheap to check: today
  ranks 1–5 are at 0% and ranks 6–9 are all late. After an ordering fix, re-run the lateness query and
  **the overdue set must not correlate with alphabetical rank.** If the same four names are still late,
  the fix did not land.
- Watch `relojistas.com` specifically — it is the sentinel, because its 3-hour cadence makes it the
  first to show starvation and the first to show relief.
- Capacity is verified by the arithmetic, not by a run: Σ(86400/interval) ≤ runs_per_day × cap.

## What this does NOT claim

Nothing here says feeds are broken or pages are stale to a reader — the sites are being refreshed, just
later than configured. The harm is schedule adherence, and the file states it in units of each site's
own cadence rather than in absolute hours for that reason.

## Filing basis (owner ruling 2026-07-31)

**No `090` run; substitution stated plainly**, on the same basis as `bugs_open/298`: no new mechanism is
asserted. The cap is `bugs_open/275`'s, council-approved and registered as **LCO-009**; everything above
is arithmetic on live rows, every figure reproducible by one query, and the central claim carries its
own disconfirming arm (a control agent that never hits its cap, and a lateness split that lands exactly
on the cap boundary rather than anywhere else). Grepped `bugs_open/` and `bugs_closed/` before filing —
`find_news_sites` appears only in `275`'s census, never as a defect in its own right.

## Related

`bugs_open/275` (the class; its §2026-08-18 evening entry has the `collected_data` census method that
found this) · register **LCO-009** — **its "expect the WARN to fire on work-queue steps; that is not a
false positive" note is vindicated here, and its "eventual coverage is not a defect" gloss is what this
file narrows** · `bugs_open/298` / `bugs_open/313` (the other live instances, both on `internal-linker`)
· `bugs_closed/297`.

---

# UPDATE 2026-08-22 — still valid, WORSE, and two of the three fix candidates above need correcting

Added by the `bugfix_316_news_feed_ordering` lane (docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_316_news_feed_ordering/`). Ownership checked first:
`scripts/who-owns.py 316` names `bugfix_275_silent_row_caps`, which **closed 2026-08-19 10:30Z** and whose
handoff lists this ticket among *"three tickets that stand on their own — they are not this lane and
nobody owns them"*. Taken on that basis.

## Still live

`content-feed-trigger.find_news_sites` still ends **`ORDER BY s.domain LIMIT 5`**, byte-identical to the
text quoted above. [MEASURED 2026-08-22 09:52Z]

## The starvation has compounded — the sentinel has moved

All five retained runs returned **5 of 5** (at cap), reproducing the central claim on a fresh window:

| run (UTC) | domains returned |
|---|---|
| 08-22 08:38 | dartsonline, gaswholesalers, mortgagecalculator, relojistas, vetcomparison |
| 08-22 02:38 | ai-agent-orchestration, dartsonline, fundamentallyai, relojistas, robot-hands |
| 08-21 20:37 | dartsonline, gaswholesalers, mortgagecalculator, relojistas, vetcomparison |
| 08-21 14:37 | ai-agent-orchestration, dartsonline, fundamentallyai, relojistas, robot-hands |
| 08-21 08:37 | ai-agent-orchestration, dartsonline, fundamentallyai, gaswholesalers, mortgagecalculator |

**`webdesign.co.uk` (alphabetical rank 9, last) appears in ZERO of the five.** Last fetched
**2026-08-21 02:45Z** — over **31 hours** on a **6-hour** cadence, i.e. **419% of its own cycle**. This
file recorded that same site at **7%** three days ago.

It was eligible by the trigger's **own** predicate at the runs it missed, not merely by a looser
paraphrase — `news_feed.recommended = true`, **128** deployed pages, and **5 of 5** active sources due at
both 08-21 14:37Z and 08-22 08:38Z. Control in the same query: `ai-agent-orchestration.com` reads **0**
sources due at those same instants, so the check discriminates.

⚠ **The sentinel this file nominates, `relojistas.com`, is currently a POOR one** — it appears in 4 of 5
runs and is not overdue. At alphabetical rank 6 it sits just past the old cut and wins whenever fewer
than five earlier sites are due. That is the same mechanism, not a refutation: the alphabet does not fix
*who* is late, it fixes *that the latest-lettered contender loses*. **Watch `webdesign.co.uk`** — it is
the site the ordering currently starves, and the one whose relief will be unambiguous.

## ⚠ CORRECTION to fix candidate 1 — `NULLS FIRST` as written creates a permanent squatter

The eligibility predicate has two arms: `NOT EXISTS (any active content_sources)` **OR**
`EXISTS (a source due now)`. A site matching the **first** arm has no active sources, so its
`min(next_fetch_at)` is NULL **and it is permanently eligible** — no fetch can advance a timestamp it does
not have. Under `ORDER BY min_next_fetch_at NULLS FIRST` such a site would win **every run for ever**:
deterministic starvation of everyone else, which is worse than the alphabet.

[MEASURED 2026-08-22] Zero sites are in that state today (all nine carry 1–9 active sources), so this is
latent, not live — but the arm is deliberate and the fix must answer it rather than delete it. Second,
smaller trap in the same expression: a source with `next_fetch_at IS NULL` has **never been fetched** and
is maximally overdue, but SQL `min()` skips NULLs, so a bare `min(cs.next_fetch_at)` hides it.

The ordering fix is still right, and it is not even a new convention — the platform's own Go layer
already orders this exact work by `next_fetch_at ASC NULLS FIRST` at
`platform/orchestration/actions/dispatch_feed_sources_action.go:101` and
`platform/orchestration/actions/feed_actions.go:1016`. Those select **sources within a site**; the
**site**-selection query, which lives in config rather than Go, is the one layer that skipped it.

## ⚠ CORRECTION to fix candidate 2 — there are TWO caps, in series, and raising one does nothing

`process_sites`, the `loop` step that consumes `news_sites.rows`, carries **`max_iterations: 5`**.
[MEASURED 2026-08-22: `loop_cap|query_cap` = `5|5`.]

So *"raise the cap to ≥10"* as written would change throughput by **nothing** — the query returns 10 rows
and the loop processes the first 5 and stops. And it would be worse than inert: the cap-hit census this
class is measured by (`jsonb_array_length` of the step's own output) would go from 5-of-5 to 10-of-10 and
**stop reporting a cap hit**, so the instrument would report relief that never happened.

Supply is `runs/day × min(query LIMIT, loop max_iterations)`. **Both literals must move together or
neither moves.** The capacity arithmetic itself reproduces exactly on today's rows — 9 eligible sites,
**42** fetches/day demanded against **20** supplied, **2.10×** — and remains, as this file says, an owner
spend decision. Each loop iteration spawns and calls a `content-feed-orchestrator` per site (600 s
timeout), so the cap is a real spend gate.

## The class, censused fleet-wide

Every `query_database` step in live config carrying a `LIMIT`, classified: ~19 are the `LIMIT 1`
fetch-one/claim-one idiom, correctly ordered; two put the `LIMIT` inside a subquery (outer result one
row); `build-pipeline-trigger.find_dispatchable_site` is correct FIFO; `model-directory-trigger`
(`random()`, 12) never binds at 3–4 rows. **`find_news_sites` is the only live member of the dangerous
shape.**

The near-miss is the instructive one. `meta-description-backfiller.load_pages_missing_meta` is
`ORDER BY p.name LIMIT 25` — alphabetical *and* capped — and is **not** this defect, because a page that
gains a meta description **leaves** its candidate set. That forces the distinction, which sharpens the
narrowing this file made to register **LCO-009**:

> A cap on a set **replenished by the clock** (rows never leave; they acquire a later due-time) starves
> the tail **permanently** under a static `ORDER BY`. A cap on a set that is **consumed** (a row leaves
> once served) is only a batching delay — and there, "coverage is eventual" is true.

The row count alone cannot tell the two apart, which is precisely why LCO-009's WARN cannot.

---

# FIXED (ordering half) 2026-08-22 10:54Z — migration `554`, live on apply. **STILL OPEN** for the capacity half.

`ORDER BY s.domain LIMIT 5` → `ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5`, with the source
aggregate lifted into a derived column because `next_fetch_at` is only in scope inside the `EXISTS`
subquery. **The eligibility predicate is byte-for-byte unchanged**, so the blast radius is exactly "who
goes first" — the pre-fix and post-fix queries return the same five sites at the same instant.

`docs/agent_docs/sql_for_agents/554_news_feed_trigger_orders_by_the_schedule_not_the_alphabet.sql`
(+`_ROLLBACK`), ledger-recorded. Council **APPROVED**, corr `e6e8b923-f614-4a1e-97d8-bf40fb5e3cc3`.
Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_316_news_feed_ordering/`.

**Verified at the artefact.** The query read back out of the live agent row and `EXECUTE`d as the runtime
will run it now returns `webdesign.co.uk` **first** — the site that had been absent from 5 of 5 runs.

**And the class now has a detector**, so this shape is findable without a human reading a query:
`cmd/config-key-audit --capped-schedule-ordering`, daily CronJob `capped-schedule-ordering-check`,
register **SCH-027**. Its demand control is the point: **1 finding over 194 live agents before the
migration, 0 after** — same binary, same command, only the config changed. Both runs are committed in
the lane as `CONTROL_prefix_…` / `CONTROL_postfix_…`. ⚠ The Go half is **inert until the owner's next
fleet roll**, and its image has never been built, so do not apply the CronJob overlay before a
`build-`+`push-` (an ImagePullBackOff here reports as a Job still RUNNING).

## ⚠ WHY THIS FILE STAYS IN `bugs_open/`

Two reasons, and the second is the one that matters:

1. **The capacity half is untouched and is an owner decision.** 42 site-fetches/day demanded against 20
   supplied (2.10×), re-derived from live rows on 2026-08-22. Ordering decides *who* absorbs a shortfall;
   it cannot create a slot.
2. **The relief itself is not yet observed.** Everything above verifies the *query*. The trigger is
   6-hourly and the next run after the fix is **14:37Z**; nothing has re-fetched yet. Per this estate's
   own bar, a fix is closed when it is **fixed AND live AND proven at the artefact**, and the artefact
   here is a fetch that has not happened.

## How the next session should verify the relief (and how NOT to)

Run the lateness query in the lane RUNBOOK after 14:37Z and check the **bug file's own disconfirming
test**: *the overdue set must no longer correlate with alphabetical rank.*

- **Watch `webdesign.co.uk`, not `relojistas.com`.** This file nominates relojistas as the sentinel; on
  2026-08-22 it was served in 4 of 5 runs and was not overdue at all, so it would have shown nothing.
  `webdesign.co.uk` is the site the alphabet was actually starving.
- ⚠ **Do not read "everyone is on time" as success — it would mean you measured wrong.** At 2.10×
  oversubscription somebody is always late. The claim to test is that **lateness ROTATES**. If nobody is
  late, suspect the measurement before believing the fix.
- ⚠ **Do not read one good run as the fix working either.** Five slots, nine sites: after the most
  overdue site is served, the next-most-overdue goes next. It is the *pattern over several runs* that
  distinguishes a fair queue from a lucky one.

---

# CAPACITY HALF DONE TOO — 2026-08-22, migration `556`. **Both halves now fixed; file stays OPEN pending observation.**

**OWNER DECISION 2026-08-22: "increase the capacity with both caps together."** Done —
`docs/agent_docs/sql_for_agents/556_news_feed_capacity_both_caps_to_10.sql` (+`_ROLLBACK`), applied and
ledger-recorded. Council **APPROVED unanimously** (corr `2cfe6fbd-c7da-4f63-ba22-9883305c38df`,
*"all reviewers approve"*, 10 reviewers).

Both literals moved in **one** migration with a verify block asserting both — the query `LIMIT 5 → 10`
and `process_sites.max_iterations 5 → 10`. Fix candidate 2 in the original file said only the first.

## ⚠ THIS FILE'S HEADLINE ARITHMETIC IS A POOL FIGURE, AND THE POOL FRAMING MISLEADS

*"42 demanded vs 20 supplied, 2.10× · removing the cap entirely still leaves 36 vs 42, 1.17×"* is
correct as arithmetic and wrong as a guide to what a cap can buy. **Per site, no cap can help at all
beyond 4 fetches/day**, because the trigger fires 4×/day
(`scheduled_tasks.content-feed-refresh`, `interval_seconds = 21600`). [MEASURED 2026-08-22]

| | wants/day | ceiling at a 6-hourly trigger | after `556` |
|---|---|---|---|
| 7 sites (6 h cadence) | 4 | 4 | **fully satisfied** |
| `dartsonline.com` (4 h) | 6 | 4 | capped **by frequency** |
| `relojistas.com` (3 h) | 8 | 4 | capped **by frequency** |

So `556` takes **seven of nine sites from "served roughly every other run" to "on their configured
cadence"**, and the entire residual — 6 fetches/day — belongs to two sites whose cadences are shorter
than the trigger interval. **That is a trigger-frequency or cadence decision and it remains open.** The
1.17× figure above is really "two sites want more runs than exist", not "the cap is still slightly small".

**Cost, measured:** 20 → 36 site-refreshes/day (+80%); `feed-triage` LLM spend ~78k → ~140k tokens/day.

## ⚠ A THIRD CAP, on a different axis, one source from binding

`DispatchFeedSourcesAction` reads `max_dispatches`, default **10**, bounding SOURCES **per site**;
`dispatch_sources` sets only `site_id`, so the default applies. The busiest eligible site has **9 active
sources**. It does not bind today and **one more source on that site makes it bite silently.** Different
axis, not fixed here, recorded so nobody rediscovers it from a symptom.

## Why this file is STILL OPEN

**Nothing has been observed yet.** Both changes are verified at the config and at the query — the
installed query now returns 5 rows against a cap of 10 (before `556`, every run returned exactly 5 of 5),
and the class detector reports 0 findings over 194 live agents. But **no trigger run has happened since
either migration.** Next run **14:37Z**.

To close, after a run or two:
1. `webdesign.co.uk` refreshed — its `last_fetched_at` moves off 2026-08-21 02:45Z;
2. the overdue set no longer correlates with alphabetical rank (this file's own disconfirming test);
3. the seven 6-hour-cadence sites sit at ≤ 100% of their own cycle;
4. `relojistas.com` and `dartsonline.com` remain late — **expected, and NOT a failure of these fixes.**
   If they are on time, suspect the measurement.

⚠ **"Everyone is on time" is still the wrong success criterion**, for a narrower reason than before: it
is now achievable for seven sites and impossible for two.

---

## CONTRIBUTION 2026-08-22 from the `bugs_open/318` lane — `capped-schedule-ordering-check` exists everywhere except the cluster, and the next release will create it

Not a request and not a criticism — a fact your lane will want before the next roll,
found while auditing release coverage for `318`.

**`capped-schedule-ordering-check` has an overlay pinned at `v1.0.1324`, a dockerfile,
`build-`/`push-`/`deploy-` targets, and membership of both `RELEASE_IMAGES` and
`AGENT_DEPLOY_SERVICES` (`d56fd6b11`) — and `kubectl -n ai-persona-system get cronjob`
does not list it at all** `[MEASURED 2026-08-22]`. It was scaffolded and never applied,
which is consistent with your commit message; recording it because two things about it
have changed since.

1. **It had never been built at any tag, and that broke the next release.** It was one of
   three images declared in `RELEASE_IMAGES` and reached by no build target, so
   `push-backend` — which loops that list with `docker push … || exit 1` — would have
   aborted the release *after* building 22 images and *before* `deploy-core`, deploying
   nothing at all. Fixed in `95757b6c2` by deriving `build-backend` from `RELEASE_IMAGES`,
   so every release image is now built by construction. **Your image now builds**: proven
   at a scratch tag (`make build-capped-schedule-ordering-check IMAGE_TAG=scratch-318-probe`
   → image exported, then removed), which mattered because it had no prior evidence its
   dockerfile worked.
2. **So the next `make release` will CREATE the CronJob**, because it is in
   `AGENT_DEPLOY_SERVICES` and `deploy-agents` applies the overlay. If that is what you
   want, nothing to do — it will arrive built from committed `HEAD` at the release's
   pinned commit, which is the good version of switching it on. **If it is NOT ready to
   run daily yet, take it out of `AGENT_DEPLOY_SERVICES` now** (leaving it in
   `RELEASE_IMAGES` is harmless — the image gets built and pushed and nothing applies it).

Also worth knowing: **the next release must run at ≥ `v1.0.1325`.** `v1.0.1324` is
contaminated — `commit-sha-exposure-check:v1.0.1324` was hand-built from an unpinned
commit, and a same-tag re-push serves the node's stale cached image.

Nothing here needs a reply. The release-coverage gate that found it is register
**BLD-026**; its report names any service in this shape.
