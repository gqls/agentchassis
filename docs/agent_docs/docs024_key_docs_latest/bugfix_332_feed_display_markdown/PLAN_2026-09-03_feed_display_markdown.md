# PLAN 2026-09-03 — bugs_open/332: feed markdown reaches visitors as literal text

**Lane opened** 2026-09-03 by the `332` session, resuming a bug that
`site_delivery_and_editor/HANDOFF_2026-09-03_boxingonline_owner_review_continue_here.md`
lists as **unstaffed** and that `bugfix_184` closed against ("nothing owed unless 332's
trigger fires"). It has fired.

---

## 1. What we are trying to do

Stop raw markdown from third-party news sources reaching visitors as literal text — on
**every** surface that displays it, not just the one that was fixed in August.

## 2. The problem, in plain terms first

`content_feed_items.source_summary` is our stored copy of a third-party article summary. It
legitimately contains the source's markdown. Anything that puts it in front of a visitor has
to remove the marker characters first, because the render pipe interprets nothing —
`# Heading` and `[text](url)` reach the page exactly as written.

In August one reader learned to do that. Two others never did, and the one that learned is
blind to more than half the cases. That is the whole bug.

## 3. Three findings [MEASURED 2026-09-03]

### 3.1 The strip is structurally blind, not absent

Every one of the 14 markdown occurrences live on the fleet today is an **unclosed**
half-pattern — `[text](https://…` with no closing paren, `**Text` with no closing `**`.
`datahelpers.MDLinkRe` requires the `)`; `MDBoldRe` requires the `**`.

The control is the zero: those same pages serve **no** ATX headings while 1,177 feed rows
carry them. The strip works. It works only on complete markdown.

### 3.2 The truncation that manufactures the half-patterns is ours

The bug file's 09-02 addendum marked this `[UNVERIFIED]`. It is
`internal/adapters/websearch/providers/firecrawl.go:143-150`:

```go
if len(snippet) > 200 { snippet = snippet[:197] + "..." }
```

Hardcoded, no config key, and a **byte** slice. 941 rows are exactly 200 chars long.
ScrapingBee and DuckDuckGo do not truncate. So the "faithful ingest record" is already only a
197-byte prefix — **we cut the link in half ourselves, then ask a complete-link regex to clean
it up.**

### 3.3 A display surface the bug file never names is serving raw — and it wins in the browser

`/data/news-archive.json`, public and 200 on boxingonline.ugg2.com, 20 items: **7 headings,
4 complete markdown links, 5 truncated links, 1 list marker, 1 image, 1 bold** — completely
unstripped. `loadNewsItems` (`render_news_section_action.go:367-390`) takes title and summary
raw; it only truncates.

The 7 headings are the decisive evidence: the server HTML has **zero**, the JSON has **seven**,
from the same query. Two readers of one column, one of them sanitising.

And every news page carries `<script src="/tools/assets/news-listing.js">` — a published 200 —
which fetches that JSON and runs `container.innerHTML = html` **unconditionally** on a
successful fetch. Its `hasServerRenderedItems` guard only covers the empty-feed and
fetch-failed branches. So for a JS-enabled visitor, what they read comes from the unstripped
JSON, and the server-side strip is cosmetic for that audience.

### The shape

| surface | reader | today |
|---|---|---|
| news page HTML | `queryresolve/news_items.go:419 projectNewsItems` | strips — blind to half-patterns |
| `/data/*.json` | `render_news_section_action.go:367 loadNewsItems` | **no strip** — 20 dirty items |
| `feed.xml` | `render_rss_feed_action.go:226 loadRSSItems` | **no strip** — 332's filed scope |
| triage / event-extraction LLM | `feed_triage_actions.go:356`, `feed_event_extraction_actions.go:86` | raw is correct — not a defect |

This is 016b §9 verbatim, and the bug file diagnosed itself correctly: *"one call site of a
shared judgement gets the rigorous fix; the sibling stays heuristic"*. There are two siblings.
`truncateSummary` and `truncateNewsSummary` are the same function copied into two packages,
and both byte-slice.

## 4. The design, and the four decisions inside it

Ranked by **what makes the bad state unrepresentable**, not by effort.

### 4.1 One narrow shared display projection — the only structural item

New `queryresolve/feed_display_text.go`: `FeedDisplayTitle(s)` and
`FeedDisplaySummary(s, maxBytes)`. All three display readers call it.

It removes a **class** — "a reader of `content_feed_items` renders display text without the
display discipline" — rather than a set of patterns. A fourth reader inherits it.

**Reuse, not invention.** The precedent is 24 hours old and on the same site:
`queryresolve/list_item_text.go` (bugs_open/425, 2026-09-02), written because *"two producers
derive the same listing for the same component, and a hand-copied rule is how a deliberate
split becomes accidental drift"*. Its header also records the council's `reuse_agent` seat
forcing a hand-rolled truncation onto `datahelpers.SafeCut`, *"the one truncation primitive in
this codebase since 2026-07-20"*. So `feedSummaryCut` delegates to `SafeCut`.

**DECISION — the projection is deliberately narrow.** The adversarial review pass attacked a
fuller unification and was right to. The three readers legitimately disagree on six things and
every one stays with the caller:

| divergence | why it must not move |
|---|---|
| escaping: `html.EscapeString` / `encoding/json` / `xml.Marshal` | unifying double-escapes two of three; `render_rss_feed_test.go:88` pins the XML case |
| truncation: 200 / 200 / 500 | RSS deliberately carries more |
| `" (Fuente: X)"` attribution | appended **after** the cut, so it can never be truncated away |
| ordering: relevance-first + per-tool cap vs chronological + URL dedup | a relevance-ordered RSS feed re-orders every rebuild and breaks reader-side dedup |
| dates: RFC1123Z / `"3d ago"` / long-form | the long form deliberately matches what the client produces after expanding the compact one |
| topics: joined string / escaped slice / absent | — |

`DISABLE_NEWS_MARKDOWN_STRIP` moves into the projection, so **one lever disarms all three
producers** — which is what 332's own fix candidate asks for and what today's arrangement
cannot deliver.

### 4.2 The vocabulary learns the truncated shapes — in two tiers

**Tier 1 — shared primitive, detection AND strip single-sourced. Exactly one new name.**

```go
MDLinkTruncatedRe = `(?:^|[\s(])\[[A-Za-z][^\]\n]{0,80}\]\((?:https?://|/)[^)\s]{0,200}$`
```

`MDLinkRe` with the closing `)` replaced by `$`, buying a **left word boundary** to pay for
the delimiter it lost. That trade is the design: `)` carried the discrimination, `$` plus a
boundary carries it now. `config[Debug](/api/v2/logs` does not fire — letter before the `[`.

> **DECISION — the strip must PRESERVE the truncation marker.** The adversarial pass caught a
> real defect in the first design. `[^)\s]{0,200}$` swallows Firecrawl's trailing `...`, so
> `…punched himself out and [lost in the ninth round](https://…-...` would have stripped to
> *"…punched himself out and lost in the ninth round"* — a grammatical, complete-looking
> sentence that **is not what the source said**, with nothing to tell the reader anything was
> cut. Today's output is ugly and honest; that fix would have been pretty and dishonest, on a
> paid customer's page, and **no check would ever have caught it** —
> `TestStripNeverInserts` only asserts length. So: the strip re-emits a trailing `...` when it
> removes a truncated tail that carried one, and a dedicated test pins it.

**DECISION — images: strip order only, NO new detection pattern.** `MDLinkRe` already fires on
the inner `[alt](url)` of all 93 letter-alt images, so a new `md_image` pattern name buys zero
detection and can only perturb `transformRouteSlot`'s routing and the exact-pattern-set test.
Instead run the image strip **before** the link strip so the stray `!` goes: `![alt](url)` →
`alt`, not `!alt`. `![](url)` — empty alt, **30 rows** — stays untouched, because the alt must
start with a letter; matching it would manufacture a blank, the one thing the council's guard
forbids.

**Tier 2 — feed-display strip only, in no detector.**

```go
mdFeedListMarkerRe  = `(?m)^[ \t]{0,3}[-*+][ \t]+`                          // 94 rows
mdFeedBracketTailRe = `(^|[\s(])\[([A-Za-z][^\]\n]{0,80})$`                 // 70 rows
mdFeedBoldTailRe    = `(^|[\s(\[])\*\*([A-Za-z][^*\n]{0,60} [^*\n]{6,60})$` // 15 rows
```

**DECISION — why tier 2 exists, and why unclosed bold moved into it.** The two review passes
disagreed about list markers and unclosed bold. Both objections are objections to putting them
in the *shared* primitive, and tier 2 answers both:

- **A list-marker DETECTION pattern would be actively harmful.** `ExtractAssertionText` yields
  one block per `<li>`, so `(?m)^` means *block start* on the `rendered_html` surface and
  *line start* on `content_data`. A `<li>` legitimately beginning with a hyphen would file a
  finding with no repair; the item terminates `wont_fix`/`failed`, both excluded from
  `idx_swi_dedup`, so the detector re-files it weekly — and `wont_fix` is excluded from both
  sides of the promoter ratio, so **the loop is invisible in the metric you would check**.
  Migration 499 says this in terms: *"a healthy-looking ratio here is not evidence that owned
  pages are being repaired."*
- **Unclosed bold must not reach the seven `content_data` strip seams.** `**kwargs` / `**args`
  and footnote glue (`Free delivery**Terms apply`) are live shapes on the developer-facing
  sites in the affected set. The left boundary kills both (`O(n**k` and `delivery**Terms` are
  letter-preceded) and the phrase guard kills `**args here`, but confining it to feed
  snippets — text this estate never authors — removes the residual from every seam where our
  own copy lives.
- `literal_markdown_test.go:47`'s `"prices from £99**"` survives untouched: `\*\*[A-Za-z]`
  needs a letter immediately after `**` and there is none. The brief predicted this would
  break; it does not.

**Precedent for strip ⊋ scan.** `rendered_html_code_spans.go:53-61` already establishes
conversion ⊊ detection. Its constraint is on **insertion**; tier 2 only **deletes**, and the
contract `Scan(Strip(x)) == ∅` requires only strip ⊇ scan, so the mirror is the safe
direction. Two property tests enforce it rather than asserting it.

### 4.3 The one council-blessed assertion that changes

`literal_markdown_test.go:166-168` pins `strip("![alt](url)") == "!alt"` (council `060bcc0a`
r5/r6, the blank-manufacture guard); `rerender_page_sections_action.go:1619-1620` repeats it in
prose. It becomes `"alt"`.

The assertion's stated **purpose** is that a bare image token keeps visible text. `alt` serves
that better: it still carries letters, so `TestStripToEmptyOnlyFromAlreadyEmptyInput` passes
unchanged in both clauses, and the visitor stops seeing a `!` that existed only because the
link strip ran before any image rule — never argued for, a leftover. The guard is not
weakened; the output is tightened while the guard holds. The test comment is rewritten to say
what the assertion is *for*, and the mirrored prose moves in the same commit.

### 4.4 Ingest — one site in, one routed

**IN — `feed_normalize_action.go:261-264`**, the 500-char cut on scraped `markdown_content`.
Ours, uncontested since 2026-03-27, and provably **inert today** (`scrape` = 472 rows with
empty summaries, because Strategy 1 sets `summary: ""`). The cheapest possible moment to close
a manufacturing site.

**ROUTED — `firecrawl.go:148`.** Not edited here. Four reasons:

1. **Irreversible, and the kill switch cannot reach it.** A display strip is reversible by
   construction — the raw record is intact and the switch restores it on the next render. An
   ingest strip writes the loss to disk. That converts a reversible transform into a permanent
   data edit, the exact posture the guardian objected to in `060bcc0a` r5.
2. **It hits the web vertical too.** `FirecrawlProvider.Search` serves `web` and `news` from
   one path. Other consumers of that snippet: `prepare_extraction_context.go` (concatenated
   into an extraction **LLM prompt**), `FilterSearchResultsAction` (`v3_site_actions.go:5606`
   matches `title+content+snippet+url` against `exclude_patterns`), `cacheSearchResults`.
3. **It makes the training corpus silently bimodal.** `cmd/reasoningset/extract.sql` rebuilds
   the triage model's input from the *current* column and states the assumption outright.
4. **A live lane holds the file.** `news_feed_ingestion` touched it 2026-09-02 (`0a408f8db`)
   and was committing at 16:31 today.

**And the benefit does not need a strip.** The 35.4%-of-budget figure is a *snippet-quality*
argument, and the instrument for a budget problem is the budget — the 200 is a hardcoded
literal with no config key. Raising the news-vertical cap costs no faithfulness, no corpus
bimodality, no consumer semantics. Handed over as a CONTRIB, per CLAUDE.md's 2026-07-29 §3
("a shared mechanism's OTHER consumers must be told, not merely measured").

One firecrawl change offered rather than taken: `snippet[:197]` is rune-unsafe and **2 rows
already carry U+FFFD**. `SafeCut` fixes it with no markdown and no semantics change.

### 4.5 The client JS — same defect family, different class, its own commit

`news-listing` and `latest-news` `js_content` insert `item.summary` into `innerHTML`
unescaped, and `loadNewsItems` does no escaping either.

**Exposure measured honestly:** 14 of 5,863 feed rows carry any HTML markup, **zero** carry
anything script-ish, **zero** of the 20 served items do. An exposure to close, **not** a live
vulnerability, and it must not be written up as one.

It is an **escaping** defect in `content_components.js_content` — a migration, live
immediately, no image — not a markdown defect in Go. Its own commit, its own `bugs_open/` file.

## 5. Explicitly not fixed, and where each goes

- **Scraped navigation as article text.** `- Tennis (W)`, `- NFL`, `- MLB` are ESPN's own
  cross-sport nav, arriving via `normalizeScrapeResults` Strategy 2. Tier 2 removes the `- `;
  the words stay nav. → `news_feed_ingestion` CONTRIB.
- **Relevance / vertical gating.** 332's own §4 parks it. The news lane.
- **Markdown tables.** Standing ruling, deliberately outside the 184 pattern set.
- **A post-strip summary quality floor.** Own `bugs_open/` file. Its contract — "no
  meaningless summary reaches a visitor" — has **no mechanical verifier**, where this change's
  ("no marker characters reach a visitor") is a regex over the served artefact. Shipping an
  unverifiable transform inside a verifiable one is how the verifiable one stops being
  trusted. It is also whole-value loss where stripping is a character subset, and "this is
  nav, not prose" is a relevance judgement — another lane's seam. The enabling fact, recorded
  so it is cheap to build later: `{{if .summary}}…{{end}}` means an empty summary renders
  **nothing**, cleanly.

## 6. Corrections to the originating brief

- The 09-02 addendum's `[UNVERIFIED]` — "which code truncates at ingest" — is **answered**:
  `firecrawl.go:143-150`, ours, 197 bytes plus an ellipsis.
- The addendum's §"When it stops being latent" watches for a **second RSS site**. Still one
  (relojistas.com, re-verified today). The trigger that actually fired was a different one,
  and the bug file says so; what neither the file nor the addendum names is the **JSON
  surface**, which is where most of the damage is.
- The bug file's fix candidate names two producers ("one lever disarms both"). There are
  **three**.
