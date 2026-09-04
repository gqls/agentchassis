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

---

# ADDENDUM 2026-09-03 — FIXED at the producer; what shipped, what did not, and the surface this file never named

**Added by the `332` lane**, which picked this up after
`site_delivery_and_editor/HANDOFF_2026-09-03_boxingonline_owner_review_continue_here.md`
listed it **unstaffed** and `bugfix_184` closed with "nothing owed unless 332's trigger
fires". It had fired. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_332_feed_display_markdown/`.

**Status: fix COMMITTED, not yet live** — Go, so inert until a chassis roll. This file stays
OPEN until the served artefacts are re-measured after that roll.

## 1. The 09-02 addendum's `[UNVERIFIED]` is ANSWERED, and the answer is that the half-patterns are ours

`internal/adapters/websearch/providers/firecrawl.go:143-150`:

```go
if len(snippet) > 200 { snippet = snippet[:197] + "..." }
```

Hardcoded, no config key, and a **byte** slice. 941 rows sit at exactly 200 characters. So we
cut the link in half ourselves and then asked a complete-link regex to clean it up.

**Not edited by this lane.** Routed to `news_feed_ingestion`, who own that file, with the
measurements — and who **fixed the producer side themselves the same day** (`6f0a246de`): the
cut now backs off a genuine link opening, and is rune-safe. They independently declined the
strip-before-truncate for the reasons this lane gave: it writes an irreversible loss to a
record `DISABLE_NEWS_MARKDOWN_STRIP` cannot undo, on a path shared with web search, feeding
`cmd/reasoningset`'s training corpus. Two lanes reached that seam from opposite ends and
agreed. **Their fix is inert until the web-search-adapter image rolls, which is the owner's
call, not theirs.**

## 2. THE SURFACE THIS FILE NEVER NAMED, and it is the one that wins in the browser

This file's §"The mechanism" lists **three** readers of `source_summary`. There are **four**,
and the missing one carried most of the damage:

`render_news_section_action.go:367-390 loadNewsItems` writes `/data/latest-news.json` and
`/data/news-archive.json` — **public** — from `r.Title` and `r.Summary` **raw**. No strip, no
escaping, only a truncate.

Measured at the artefact 2026-09-03, boxingonline.ugg2.com's archive JSON (200, 11,074 bytes,
20 items): **7 ATX headings, 4 complete markdown links, 5 truncated links, 1 list marker, 1
image, 1 bold**. The server-rendered HTML of the **same query** carried **zero** headings.

**And every news page overwrites itself with it.** `<script src="/tools/assets/news-listing.js">`
is a published 200 on all five affected hosts; it fetches the archive JSON and runs
`container.innerHTML = html` **unconditionally** on a successful fetch. So for a JS-enabled
visitor the August fix was **cosmetic**. Full trap, and why a `grep` of the page cannot find
it: LANDMINES, *"The served news page HTML is OVERWRITTEN in the browser"*.

## 3. Surface 1's status, corrected precisely

The 09-02 addendum proposed re-marking surface 1 from `**Fixed.**` to *"fixed for complete
links; blind to pre-truncated ones"*. That is right, and the control makes it sharper: the
affected pages carry **zero** ATX headings while 1,177 feed rows have them. **The strip runs
and succeeds. It was blind, not broken** — `MDLinkRe` requires the closing `)`, `MDBoldRe` the
closing `**`, and every one of the 14 occurrences live across five sites was an unclosed
half-pattern.

## 4. IT WAS NEVER ONE SITE

Five news pages served literal markdown, each verified with its own per-host 404 control:

| host | `](http` | `![` | `**A` |
|---|---|---|---|
| boxingonline.ugg2.com | 5 | 1 | 1 |
| fundamentallyai.com | 3 | 0 | 0 |
| robot-hands.com | 2 | 0 | 0 |
| ai-agent-orchestration.com | 2 | 1 | 0 |
| idea.uk | 2 | 1 | 0 |

Eleven sites carry dirty feed rows. All damage is `source_type='news_search'`; **`rss` (834
rows) carries ZERO markdown**, which is exactly why relojistas measured clean in August and
this file read as latent.

## 5. What shipped

1. **One display projection** — `queryresolve/feed_display_text.go`
   (`FeedDisplayTitle`/`FeedDisplaySummary`), called by all three display readers. It kills
   the duplicated `truncateSummary`/`truncateNewsSummary` pair (identical bodies in two
   packages, both byte-slicing; 2 rows already carry U+FFFD) in favour of
   `datahelpers.SafeCut`. **`DISABLE_NEWS_MARKDOWN_STRIP` moved into it, so ONE LEVER NOW
   DISARMS ALL THREE PRODUCERS** — this file's own fix candidate asked for that and the old
   arrangement could not deliver it. Proven, not asserted:
   `TestLoadRSSItemsHonoursTheKillSwitch`.
2. **The vocabulary learned the truncated shapes** — one new detection pattern
   (`MDLinkTruncatedRe`), an image strip-order change with no new pattern name, and a **tier 2**
   feed-display strip (list markers, bracket tails, bold tails) that enters no detector.
3. **The strip re-emits the truncation marker.** This is the finding no check could have made:
   deleting the `...` along with the URL turns a severed fragment into a grammatical sentence
   the source never wrote. `TestStripNeverInserts` asserts only length; the result scans clean
   by construction.
4. `feed_normalize_action.go`'s 500-char scrape cut now strips first (inert today, guarded
   anyway).
5. `sweep_site_defects.sh` §1.4 gained `/data/*.json` and `feed.xml` arms — **the check that
   would have caught this in August and did not, because it read only the page**.

## 6. THE RSS SURFACE — this file's original scope — is fixed, and the artefact CANNOT prove it

`loadRSSItems` now calls the projection for both title and description, with the
`(Fuente: X)` attribution appended **after** the cut so it can never be truncated away.

**But relojistas is still the only `rss_feed` site and its rows still carry zero markdown**, so
a clean `feed.xml` after the roll is a **no-regression control, not evidence**. It read clean
before. The evidence is `TestLoadRSSItemsStripsLiteralMarkdown`, the unit test this file's own
fix candidate asked for. The signal to actually watch on the live feed is the **opposite**
direction: **a drop in `<item>` count, or any empty `<description>`, means the strip emptied a
live feed.** Pre-fix baseline for that comparison: **30 items, 24,437 bytes, 0 markdown, 0
empty** (2026-09-03 — and note it is 30, not the 25 this lane's plan first assumed).

## 7. The re-review trigger in §"When it stops being latent" never fired, and should be retired

It watches for a **second** site enabling `rss_feed`. Still one. The trigger was aimed at the
surface with the least traffic and no detector, and the damage arrived on the two surfaces the
file assumed were safe. Kept here as a record of a reasonable watch pointed the wrong way.

## 8. Still open, and what closes this file

- **The roll.** Go changes are inert. Re-measure the five-host table (expect zero), both
  `/data/*.json` files (20 dirty items today), and relojistas' `feed.xml` for the
  item-count/empty-description signal.
- **⚠ EVERY SERVED CHECK ABOVE EXPIRES THE MOMENT THIS ROLLS, and it must not be read as
  evidence about the stored data afterwards** (the `news_feed_ingestion` lane's catch,
  2026-09-03, found on their own migration-746 verification plan before it was found here).
  A clean JSON post-roll means **the strip ran** — nothing more. A table full of raw markdown
  and a spotless one produce byte-identical served output once the projection sits in front of
  them, so the check cannot come out otherwise. That is an undisconfirmable measurement wearing
  the clothes of *"judge at the served artefact"*, which is exactly the rule that walks you
  into it. **Two questions, two instruments:** *is the visitor seeing junk?* → the served
  surface. *Is ingestion clean?* → `content_feed_items`, never the surface. And note the
  corollary of the switch now living in the projection: flipping
  `DISABLE_NEWS_MARKDOWN_STRIP` fills yesterday's "verified clean" surfaces with junk that was
  in the table all along.
- **The self-heal premise, which is falsifiable — WITH A DEMAND CONTROL, corrected 2026-09-03.**
  All 9 affected `page_components` were rewritten within 19 hours, three within the hour
  [MEASURED 2026-09-03 16:20Z], so a producer-side fix should repair every page unaided within
  about a day. The falsifier as first written — *"re-run the five-host census and expect zero"* —
  was **incomplete**: a page can read zero because that day's feed carried no markdown, which
  proves nothing about self-healing. **Pair it:** the page must read zero WHILE the site's own
  feed rows still carry the shape in the column. Zero on both is not a better result, it means
  the census is broken. If the page does not reach zero while the column is dirty, the
  self-heal premise broke, not the fix.
- **Council `803f0d81-02be-4bb6-9e65-363439ff87ba`** — submitted, verdict owed and to be read.
- Two things routed out rather than folded in: **`bugs_open/472`** (the same JSON inserted into
  `innerHTML` unescaped — 14/5,863 rows carry markup, none executable; an exposure, not a
  vulnerability) and **`bugs_open/473`** (a stripped summary can still be ESPN's navigation
  menu — a clean `](http` count on that page is NOT evidence this is fixed).


---

# ADDENDUM 2026-09-04 — VERIFIED LIVE AND WORKING, plus two defects the verification found

**Added by the `332` lane after the 2026-09-03 22:06:58Z chassis roll.** Lane docs and the
cold-start doc: `docs024_key_docs_latest/bugfix_332_feed_display_markdown/HANDOFF_2026-09-04_continue_here.md`.

## 1. The fix is LIVE and WORKS — proven by capability, not by commit

The image tag (`v1.0.1360`) also appears in commits predating this lane, which is the same-tag
stale-cache shape, and a 40-sha binary probe was **Terminated** — so its empty output is not
evidence and was not read as any.

What settles it is the demand control:

| site | feed column (7d) | re-rendered | page |
|---|---|---|---|
| **dartsonline.com** | **dirty** — 2 tail-links | **10:40Z**, post-roll | **CLEAN** |
| gaswholesalers / relojistas / webdesign | clean | post-roll | clean — **proves nothing** |

Dirty column + post-roll re-render + clean page. boxingonline.ugg2.com is **5 → 0**.

## 2. Two defects found by that verification — fixed in `adef5d481`, NOT YET ROLLED

**(a) A truncated IMAGE survives.** `![alt](url…` falls through every rule: `mdImageStripRe`
needs the closing paren, `mdLinkTruncatedStripRe`'s left boundary `(^|[\s(])` **rejects the
preceding `!`** — which is exactly the image marker — and `mdFeedImageTailRe` requires no `]`.
Found on idea.uk, re-rendered **10:49Z**, i.e. *after* the roll: that timing is what made it a
gap rather than a stale binary. It is the §8 falsifier of the previous addendum firing as
designed.

**(b) ⚠ `MDLinkTruncatedRe` WAS NEVER WIRED into `LiteralMarkdownPatterns`.** Declared,
exported, documented, and cited to council `803f0d81` as *"detection AND strip single-sourced"*.
It was not. **The scan has been blind to truncated links** — this bug's own defect — since the
change shipped, so a page serving one scanned clean. No wrong repair was dispatched (the strip
worked throughout), but the claim to the council was false.

Nothing could have caught it: every property test routes through that one function, and
`TestStripThenScanFindsNothing` passed **vacuously** — a fixpoint holds trivially when `Scan`
cannot see the pattern. **A round-trip property cannot detect a missing arm.** → `WRONG_CALLS`.

## 3. What closes this file

The roll carrying `adef5d481`, then the three-part check in §4 of the handoff (five hosts, the
demand control, and the `rerendered_since_roll=1 AND still_dirty=1` falsifier). Nothing else.
`bugs_open/472` and `bugs_open/473` are separate and owned elsewhere.
