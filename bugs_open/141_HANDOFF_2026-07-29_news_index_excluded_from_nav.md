# 141 — a news-index page at its canonical URL can never enter the nav

**Filed:** 2026-07-29, by the webdesign.co.uk thread (session 4), which watched it
fire live while waiting for the handoff's promised "the nav row will reappear by
itself".
**Severity:** medium — a deployed, in-header page is permanently invisible in
site navigation, silently, with every green status along the way.
**Class:** structural (shared classification helper), one-word fix.
**Fix:** committed with this file (`isSectionIndexType` + regression tests).
**OPEN until** the fix is live on a rolled image and the webdesign.co.uk News
nav row survives a `populate_nav_tables` run.

---

## The symptom

webdesign.co.uk's news page (`page_type='news-index'`, `/news/index.html`,
`in_header=true`) deployed 2026-07-29 07:44Z. A `nav_drift` item completed
07:52Z: `populate_nav_tables` ran, deleted and rebuilt `site_nav_items` — and
the News row was still absent. Every status green, nav wrong.

## The mechanism, watched not inferred

`classifyPagesForNav` (`populate_nav_tables_action.go:339`) skips any page whose
URL sits under a child prefix (`/tools/`, `/blog/`, `/news/`, …) unless
`isSectionIndexType(page_type)` exempts it as the section's parent:

```go
if isChildPageURL(page.URL) && !isSectionIndexType(page.PageType) {
```

`isSectionIndexType` (`v3_site_actions.go:4700`) admitted `blog-index`,
`entity-directory`, `section-index` — **not `news-index`**. So a news-index page
at `/news/index.html` is classified as its own child and skipped, permanently.

Live evidence, nav-updater pod `agent-nav-updater-f466efeb-pjkp4`, 07:52:09Z:

```
"msg":"classifyPagesForNav: skipping child page","step_name":"refresh_nav_tables",
"name":"news","url":"/news/index.html"
```

## Why nobody hit it before

The three other deployed news-index pages dodge the prefix list, not the bug:
gaswholesalers.com and robot-hands.com carry `/news.html` (the non-canonical
shape `bugs_open/080` exists to eliminate), and relojistas.com carries
`/noticias/index.html` (`/noticias/` is not in the prefix list). The
canonicalisation work of 015/080 makes `/news/index.html` the shape every future
news page gets — i.e. the fleet was converging on exactly the URL this helper
mishandles. webdesign.co.uk is the first site to combine the canonical URL with
a nav-flagged news page.

Census (2026-07-29, unfiltered baseline then the wrongly-excluded set): the
child-prefix skip currently excludes 185 pages fleet-wide — 107 `tool`,
60 `blog-post`, 17 `guide` (all correctly), and **exactly one wrongly:
webdesign.co.uk `/news/index.html`, the only `news-index` under a child
prefix.**

## The fix (committed with this file)

Add `news-index` to `isSectionIndexType`. Call-site effects, all checked:

1. `populate_nav_tables_action.go:339` — the fix: canonical news pages enter nav.
2. `populate_nav_tables_action.go:419` — `navPriorityTier`: news-index pages
   rise from tier 3 (by name) to tier 2 (typed hub). Matters only when a header
   is over `max_header_items` (8); no current site with a news-index page is.
3. `sectionStemOf` (`v3_site_actions.go:4713`) — for `/news/index.html` the
   `/index.html` fallback already treated it as an index: **no change**. For
   flat `/news.html` pages, "news" now registers as a section stem, so the
   planner's Pass C drops an LLM re-proposal of a colliding flat news page —
   which is a small *repair* of the 015/080 duplicate-news-page family, not a
   regression.
4. `v3_site_actions.go:4995` (Pass C read) — same as 3.

Regression tests: `populate_nav_tables_classify_test.go` — canonical URL enters
primary, flat URL still enters primary, tool/blog-post children still skipped.

## How to verify live (the closing bar)

1. Pod-grep the rolled image for the marker: the new comment string
   `news-index belongs here` in the binary is NOT greppable (comments compile
   out) — grep the TEST expectation instead is equally vacuous. Use behaviour:
2. File a `nav_drift` item for webdesign.co.uk (shape in
   `webdesign_couk/SQL_p19`, change the item_key suffix), let nav-updater run, then:
   `SELECT label, url FROM site_nav_items ni JOIN sites s ON s.id=ni.site_id
    WHERE s.domain='webdesign.co.uk' AND ni.url='/news/index.html';`
   — one row, and the served header on any page gains News after the re-render
   fan-out.
3. Negative control: `/tools/touch-target/index.html` must still be absent from
   `site_nav_items`.

## Council verdict — APPROVED round 1 (corr `e0a52a70`, 08:13:34Z)

"approved with 1 advisory objection(s) — none high-severity"; 8 seats abstained
(relevance-gated), 0 unreadable. The code commit `fc7c05c21` predates the
verdict (committed-not-held, per practice), so no trailer rides it; this file is
the join. Advisories, each answered with a query rather than carried:

- **debug_historian (medium): closing needs pod-level verification, not
  `go test ok`.** Agreed and already this file's closing bar (§ How to verify
  live) — behaviour-based, because a one-word switch case has no discriminating
  binary string.
- **debug_historian (low): the ai-agent-orchestration.com check was deferred to
  "a reviewer" instead of run.** Run 08:2xZ: its news page is
  `page_type='content'`, `in_header=false` — it does not take the new branch.
- **guardian (low): tier/stem call sites inherit the change untested.**
  Quantified 08:2xZ against every news-index site: robot-hands 7 candidates for
  8 slots (order may shift, membership cannot); relojistas and webdesign 5 each
  (inert); **gaswholesalers 12 candidates for 8 slots — there the promotion is
  LIVE behaviour change**: on its next nav rebuild News is guaranteed a header
  slot and the last tier-3 content page drops. That is the truncation design
  operating on correct classification (the site's own `in_header=true` flag says
  News belongs there), recorded here so the change in its header is traceable to
  this fix rather than mistaken for drift.

## The transferable pattern (016b §9 candidate)

A "typed index" family that grows by enumeration (`blog-index`,
`entity-directory`, `section-index`, now `news-index`) fails silently for each
type someone forgets, and the failure surfaces far away (nav, planner
validation) with no error. The census that found the blast radius is one query;
run it before assuming the next typed index is handled.
