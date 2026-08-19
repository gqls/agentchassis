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
