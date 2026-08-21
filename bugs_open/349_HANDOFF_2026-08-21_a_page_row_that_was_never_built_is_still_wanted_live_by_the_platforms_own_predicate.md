# 349 — a page row that was never built is still "wanted live" by the platform's own predicate, and serves 404

**Filed 2026-08-21** by the `news_editorial_features` lane, after the
`dartsonline_traffic` lane surfaced one instance on robot-hands.com while testing
its new `unfetchable` audit state. **Status: OPEN, UNOWNED.**

> **On the 090 loop, stated plainly because the 2026-07-31 owner ruling requires it.**
> This is a cross-cutting structural claim, so it went to the diagnosis loop
> **before** any root cause is asserted here. Intake
> `854f482f-b999-402e-bcd5-604c2ac23999`, run
> `8c4f8a67-abfb-452a-8b21-0e7e59d2adc3`. **§4 deliberately contains no root
> cause** — only the mechanism as read from source and the open questions.

## 1. The one-paragraph version

`PageWantedLivePredicateFor` (`platform/orchestration/datahelpers/links.go:361-367`)
is the platform's own definition of "a page this site wants live". It returns
**`status = 'active'` and nothing else** — no `build_status` term, no
`deployed_at` term. So a page row that was planned and **never built** satisfies
it, is enumerated by every consumer as a live page, and serves **404** to anyone
who follows a link to it.

## 2. The measurement, and the split that matters

```sql
SELECT count(*) AS rows_total, count(DISTINCT site_id) AS sites
  FROM pages WHERE status='active' AND build_status='planned' AND deployed_at IS NULL;
--  42 | 14        [MEASURED 2026-08-21]
```

**42 is the wrong number to quote, and the split is the finding.** Two of those
sites have *no deployed pages at all* — they are sites planned and not yet built,
where a `planned` row is correct and expected:

| class | rows | sites |
|---|---|---|
| **ORPHAN on a live site** — the defect | **20** | **12** |
| whole site never built — legitimate | 22 | 2 (`adversecreditmortgage.co.uk` 19, `pool-ai-agents.internal` 3) |

The tell that forced the split: `adversecreditmortgage.co.uk`'s **`/index.html`**
is in the raw list. A site whose own homepage is unbuilt is not suffering this
defect; it simply has not been built. Quoting 42 would have inflated the bug by
more than double and sent someone to "fix" two sites that are fine.

Affected live sites (orphans): webdesign.co.uk 3 · `pool-energy-utilities.internal` 3 ·
vetcomparison.uk 2 · dartsonline.com 2 · idea.uk 2 · mortgagecalculator.co.uk 2 ·
leopardessconsulting.co.uk 1 · gamesdesign.co.uk 1 · loanzy.uk 1 · vonc.com 1 ·
robot-hands.com 1 · gaswholesalers.com 1.

**They really do 404.** 7 of 7 sampled at random across 5 sites, fetched
2026-08-21:

```
404 gaswholesalers.com/guides/tool-supplier-comparison-calculator-guide.html
404 mortgagecalculator.co.uk/scorecard-simulator.html
404 mortgagecalculator.co.uk/guides/market-structure/index.html
404 robot-hands.com/gripper-technology-comparison.html
404 vetcomparison.uk/tools/compliance-deadline-calculator/index.html
404 vetcomparison.uk/entities/practice.html
404 leopardessconsulting.co.uk/case-study-automated-intelligence-pipeline.html
```

**Not a stale backlog — it is still accruing.** Oldest row `2026-06-22`, newest
**`2026-08-21`** (today). Page types spread across `content` 21, `blog-post` 8,
`landing` 4, `entity-page` 3, `section-index` 2, `tool` 2, `guide` 1,
`mortgage-lenders` 1 — so it is not one producer's shape.

## 3. Why this is worth a bug rather than a tidy-up

The rows are not inert. Because they satisfy the live predicate, they are
**visible to consumers that enumerate live pages** — the same helper is used to
pick a site's news-index page for the homepage `insights_url`
(`render_news_section_action.go`), and `InboundLinkSurfaces`
(`links.go:369+`) names the census contract for "what links to this page".
A never-built row can therefore be selected, linked to, and counted as coverage.

It is also **exactly invisible to the checks most likely to look**: a page-level
audit that enumerates `build_status='deployed'` never sees it, and one that
enumerates `status='active'` sees it and gets a 404 body. The
`dartsonline_traffic` lane found the first instance only because it added an
explicit `unfetchable` third state to its census — before that, a 404 either read
as "still carrying the marker" or was silently counted.

## 4. What is NOT established here

**No root cause.** Read from source, the mechanism is only that the predicate
omits `build_status`. What is unknown, and what the 090 run was asked:

- what is supposed to advance a `planned` row to built, and why these did not;
- whether consumers of the live predicate are expected to check `build_status`
  themselves (in which case the defect is at ~N call sites, not in the helper);
- whether the two populations in §2 have the same cause or only the same shape.

**Do not assume the fix is to add `build_status='deployed'` to the predicate.**
That would change what every consumer sees in one edit, and at least one consumer
(`check_orphan_pages`) exists precisely to find pages that are *not* reachable —
narrowing the predicate could blind it. That is a shared-seam change and belongs
in architecture review, not in a patch.

## 5. Candidates, unranked, none costed

1. Narrow the predicate (see the warning above — measure every call site first).
2. Leave the predicate and add a discovery check that raises `active` +
   `planned` + no `deployed_at` on a site that has deployed pages, using §2's
   split so it never fires on an unbuilt site.
3. Fix the producer(s) — needs the 090 verdict first.

## 6. Related, and how they differ

- `bugs_open/315` — `deployed_at` stamped *without* publishing. The inverse: a row
  that claims deployment it did not get. Here there is no stamp at all.
- `bugs_open/348` — a render refusal reports `complete`. That would leave
  `build_status='deployed'`, not `planned`.
- `bugs_open/320`/`339` — same lane pairing found this; different column.
