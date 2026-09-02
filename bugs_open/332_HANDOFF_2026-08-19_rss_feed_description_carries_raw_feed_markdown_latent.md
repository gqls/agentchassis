# 332 — `render_rss_feed` emits `content_feed_items.source_summary` raw, so feed markdown reaches RSS `<description>` — LATENT today (1 RSS site, 0 served defects)

**Filed** 2026-08-19 by the bugfix_184 lane, on the council's `bug_historian` advisory in
round 6 of `060bcc0a` ("naming the gap in a code comment is not a tracked follow-up").
**Severity: LOW. Status: latent, measured clean in production today.** Filed so the exposure
has an owner and a re-review trigger, not because a visitor has seen it.

## The mechanism (plain terms first)

`content_feed_items.source_summary` is the raw ingest record of a third-party article
summary. It legitimately carries the source's markdown — measured fleet-wide 2026-08-19:
of 2,382 `relevant`/`ingested` rows in the last 30 days, **574 carry ATX headings
(`# …`), 203 carry `[text](url)` links, 49 carry `**bold**`**.

Three surfaces read that column:
1. the news **page** resolver (`queryresolve/news_items.go projectNewsItems`) — now strips
   literal markdown at the producer (commit `f3939f27d`, live v1.0.1315; kill switch
   `DISABLE_NEWS_MARKDOWN_STRIP` in `f6d632291`, inert until the next roll). **Fixed.**
2. `feed_triage_actions.go:356` — feeds it to the triage LLM as INPUT. Raw is correct there;
   markdown is data, not display. **Not a defect.**
3. `render_rss_feed_action.go:228 loadRSSItems` — copies it (truncated to 500 chars) into the
   RSS item `<description>` via `xml.Marshal`. XML escaping keeps the feed well-formed; the
   markdown MARKER CHARACTERS pass through as text, so a reader would show `# Heading` /
   `[text](url)` literally. **This is the gap.** It is the 016b §9 shape "one call site of a
   shared judgement gets the rigorous fix; the sibling stays heuristic".

## Why it is LATENT, measured (2026-08-19 ~21:20Z)

- `sites.deploy_config->'rss_feed'` is set on **exactly one site**: `relojistas.com`
  (`enabled: true`, self_url `https://relojistas.com/external.php?type=RSS2`).
  `render_rss_feed` returns early ("rss_feed not enabled in deploy_config") for every
  other site; `feed.xml` 404s on dartsonline.com, mortgagecalculator.co.uk,
  gaswholesalers.com, fundamentallyai.com (checked).
- relojistas.com's live feed (both URLs, 200, 20,179 bytes): **25 items, 0 descriptions
  matching heading / md_link / bold**. Its own feed rows: **0 headings, 0 md links** of 79
  rows in 30 days — Spanish watch-press sources do not emit markdown.
- So the 574-row fleet figure is entirely on sites that publish no RSS. **Zero served
  defects today.** [MEASURED, with the disconfirming result stated: a non-zero count on
  relojistas' feed or rows would have made this a live bug.]

## When it stops being latent (the re-review triggers)

- a second site enables `rss_feed` in `deploy_config` (any of the sites whose feed rows
  carry markdown — dartsonline.com, fundamentallyai.com, ai-agent-orchestration.com are
  the heaviest by the 574 census), or
- relojistas.com adds a content_source that emits markdown summaries.

Check: `SELECT domain FROM sites WHERE deploy_config->'rss_feed'->>'enabled'='true';` —
more than one row = re-measure the feed(s) with the §Scope regexes on `<description>`.

## Fix candidate (unbuilt — deliberately NOT ridden in on 184 round 6)

One line in `loadRSSItems`: pass `title.String` and `summary.String` through
`datahelpers.StripLiteralMarkdown(s, !datahelpers.HTMLMarkupRe.MatchString(s))` before
`truncateNewsSummary` (strip BEFORE truncate — a link cut mid-URL is a half-pattern nothing
can match), under the SAME `DISABLE_NEWS_MARKDOWN_STRIP` switch so one lever disarms both
producers. Why it was not done in round 6: the RSS surface has **no detector** (the
`literal_markdown` check scans `page_components`, not feed.xml), so a strip there would be
an unverifiable mutation with no artefact check behind it; and widening a REVISE round's
scope is how the objection surface grows. Whoever builds it should add a feed.xml scan
(curl + the four regexes over `<description>`) to the verification, and a unit test on
`loadRSSItems`' projection mirroring `TestProjectNewsItemsStripsLiteralMarkdown`.

## Relations

`bugs_closed/184` (llm_markdown slug — the page-surface bug this is the sibling of);
CQ-019 (register); `docs024_key_docs_latest/bugfix_184_literal_markdown/` NOTES 2026-08-19
~21:00Z; council `060bcc0a` round 6 (`bug_historian`, advisory LOW).

---

# ADDENDUM 2026-09-02 — NO LONGER LATENT, and surface 1's "Fixed" is conditionally FALSE

**Added by the boxingonline.com session** (first paid customer build), from the owner's second
review. **Two changes to this file's status, both measured 2026-09-02, queries inline.**

## 1. The re-review trigger has fired — but on the NEWS PAGE, not on RSS

This file's §"When it stops being latent" watches for a second site enabling `rss_feed`. That is
not what happened. **The defect is live on surface 1 — the news page — which this file records as
`**Fixed.**`** (`queryresolve/news_items.go projectNewsItems`, commit `f3939f27d`, live
v1.0.1315).

Served now at `https://boxingonline.ugg2.com/news/index.html`, a **paid customer site**:

- **5** occurrences of literal `](http` rendered as page text
- e.g. `- Tennis (W)\n- [NLL (Lacrosse)](https://www.espn.com/boxin...`
- e.g. `Itauma (14-1, 12 KOs) ultimately punched himself out and [lost in the ninth round](https://sports.yahoo.com/boxing/live/moses-itauma-vs-filip-hrgovic-live-results-round-by-round-updates-...`
- `NLL (Lacrosse)` and `MLB` are **ESPN's own cross-sport navigation**, captured as article text.

Present in `page_components.content_data` AND `rendered_html` for the `news-listing` slot, so it
is baked in, not a render-time artefact. The kill switch is **unset** on all pods checked
(`DISABLE_NEWS_MARKDOWN_STRIP` empty on both `agent-chassis` replicas and `core-manager`), so the
strip was enabled and ran.

## 2. WHY the strip does not catch these — this file predicted it, in the wrong section

`content_feed_items.source_summary` **is stored already truncated**, at ≤200 chars with a
trailing ellipsis. Truncating a `[text](url)` link mid-URL leaves `[text](url` — **an unclosed
half-pattern that a complete-link regex cannot match.**

```sql
SELECT count(*) FILTER (WHERE source_summary LIKE '%](http%')  AS with_md_link,
       count(*) FILTER (WHERE source_summary ~ '\]\([^)]*$')   AS unclosed,
       count(*)                                                 AS total
  FROM content_feed_items WHERE created_at > now() - interval '30 days';
-- with_md_link = 518 · unclosed = 282 · total = 5779
```

**282 of 518 markdown links reaching the strip — 54% — are already incomplete before it runs.**

This file's own fix candidate for surface 3 states the hazard exactly: *"strip BEFORE truncate — a
link cut mid-URL is a half-pattern nothing"*. **The same hazard already applies upstream of
surface 1**, where the truncation happens at ingest rather than at render, so no ordering change
inside `projectNewsItems` can reach it. The strip is correct for complete links and structurally
blind to truncated ones — which is why surface 1 measured clean when it was written and serves
dirty now.

`[UNVERIFIED]` **which** code truncates at ingest — I measured the stored shape (≤200 chars,
trailing ellipsis, 54% unclosed) and did not read the writer. Naming it is the first job.
The feed lane's own pointer for the adjacent extraction question:
`feed_normalize_action.go:192-194` ("Get content - prefer markdown, then HTML") takes
`markdown_content` straight from the firecrawl adapter — supplied by the news_editorial_features
session, who verified my page measurements independently and declined the work as out of lane.

## 3. Status change proposed

**LATENT → LIVE, severity LOW → the owner has seen it on a paid deliverable.** Surface 1 should be
re-marked from `**Fixed.**` to *fixed for complete links; blind to pre-truncated ones*, because a
future reader will otherwise treat the news page as a solved surface. Note this is the
`a-pass-from-a-blind-check-outlives-the-blindness` shape: the 08-19 measurement was honest and
correct, and the population it measured has since changed.

## 4. Two things NOT part of this bug, recorded so they are not folded in

- **Relevance/vertical gating** — 12–13 UFC/MMA mentions on a boxing site. Different seam
  (`feed_news_recommendation_action.go` territory per the news lane), not a markdown defect.
- **Verification trap, corrected in both directions.** `boxingonline.com` is a parked catch-all
  (invented path → **200**, real page → 114 bytes). The SERVING host is
  `boxingonline.ugg2.com` (`sites.publish_target='b2worker'`), which **404s an invented path
  correctly**. So status codes ARE usable — **probe the slug, never the customer domain.**
  `probe-page-url.sh` reports CONTROL-FAIL if handed the customer domain, and that verdict is
  about the argument, not about the site.
