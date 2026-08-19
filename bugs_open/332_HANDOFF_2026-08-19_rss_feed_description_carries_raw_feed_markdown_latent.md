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
