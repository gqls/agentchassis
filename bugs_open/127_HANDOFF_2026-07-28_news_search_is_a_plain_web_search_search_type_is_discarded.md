# 127 — every "news search" feed is a plain WEB search; `search_type` is discarded at the provider interface

**Filed 2026-07-28** from `webdesign.co.uk`. **Taken by the bugs-sweep thread 2026-07-28.**
Affects **every site in the fleet with a `news_search` content source** — measured today on
webdesign.co.uk, robot-hands.com, relojistas.com, gaswholesalers.com and
ai-agent-orchestration.com all use this path.

> ## STATUS 2026-07-28 — fix candidates 1+3 COMMITTED (`723a10259`); INERT until BOTH images roll
>
> `SearchProvider.Search` now takes the `SearchOptions` that was always defined and never
> constructed, so the parameter can no longer be dropped — the single call site must pass
> it. Per provider: **ScrapingBee** sends `search_type=news` + `tbs` and parses
> `news_results`; **Firecrawl** (live primary) sends `sources:["news"]` + `tbs` and parses
> `data.news`; **DuckDuckGo** has no news vertical on the html endpoint and now declines
> with `ErrUnsupportedSearchType` (candidate 3) — the adapter falls through, and if every
> provider declines the request errors instead of serving web results labelled news.
> `time_range` flows from `source_config` → `FetchNewsSearchAction` → `WebSearchAction` →
> adapter payload → provider date filters (`qdr:*` / `df`). Provider dates are normalised
> to RFC3339 at the boundary (Firecrawl news dates are relative text like "3 months ago");
> `WriteFeedItemsAction` parses RFC3339 only, so this is what lets `source_published_at`
> stop being NULL. 14 regression tests in `providers/search_options_test.go`,
> `websearch/search_options_test.go`, `actions/web_search_options_test.go`.
>
> **Why this stays OPEN — the fix is not live:**
> 1. **Both images must roll**: `web-search-adapter` (the interface + providers) AND
>    `agent-chassis` (the `time_range` plumbing). The adapter alone delivers the core fix;
>    the chassis part only adds the optional recency window.
> 2. **Induced verification owed** (recipe below, unchanged): after the roll, fire a fetch
>    with a recent-event query and assert `content_feed_items.source_published_at` is
>    populated and recent. Adapter pod-grep: `"unsupported search type"` present (new),
>    `"all %d providers failed after retries"` absent (replaced by a variant naming the
>    search_type — a marker that flips both ways).
> 3. **One documented-not-witnessed risk**: the `news_results` field names and `tbs`
>    forwarding for ScrapingBee come from its public docs (fetched 2026-07-28), not from a
>    live call with a real key. If wrong, the news parse yields zero results and errors
>    loudly rather than mislabelling — but check the fallbacks list in the first live run.
>
> Council verdict `a7ae8ce8-ef40-4503-be8a-972ebe1b0973`: **APPROVED round 1** (10:46Z,
> `unreadable: 0`, plan summary matched verbatim). The fix commit `723a10259` predates the
> verdict, so the trailer rides the verdict-record commit. The seats' three containment
> asks were answered by grep after approval, all clean: `SearchProvider` has no implementer
> or caller outside the adapter package + its tests (guardian's tips-to-veto condition
> refuted); nothing parses the old "providers failed" error string; no existing
> date-normalisation helper in `datahelpers/`/`pkg/` for the normaliser to have duplicated
> (reuse seat). debug_historian's ask — pod-grep the new symbols, don't trust the induced
> fetch alone — is folded into item 2 above; prior_art's ask — re-check provider
> availability at roll time, not from an earlier log — is done as part of the roll.

---

## Symptom

News feeds ingest successfully, report success, and return **evergreen reference
pages and SEO listicles instead of news**. Measured on webdesign.co.uk across two
ingestions (2026-07-27 19:49 and 2026-07-28 07:50), 10 items per source:

| query | what came back |
|---|---|
| `web platform standards W3C WHATWG specification` | `HTML5 specification`, `The web standards model — MDN`, `What is the difference between the W3C and the WHATWG? : r/javascript` — **evergreen documentation** |
| `design agency acquisition merger industry report` | `Design Agencies Market Research Report 2034`, `Creative Agency Market Trends & Forecast Data 2035`, `Merge: Trusted M&A Advisor to Buy and Sell Agencies` — **report vendors and advisory-firm marketing, dated years in the future** |
| `typeface release design system open source` | `28 Best Free Fonts for Modern UI Design in 2026`, `Favourite free (open source) software? : r/typography` (twice) — **listicles and forum threads** |
| `UK web design industry` (since retired) | 9 of 10 were `Top Web Design Agencies in the UK 2026` ranking listicles |

Nothing is dated. Nothing is recent. Several results are literally market-forecast
pages for 2034–2035. That is the signature of a **web** search ranking by authority
and SEO, not a news index ranking by recency.

## Root cause — the parameter cannot reach the provider

The chain is intact right up to the last hop, which is what makes it convincing
from the inside:

1. `FetchNewsSearchAction` (`platform/orchestration/actions/feed_fetch_async_actions.go:158`)
   **explicitly forces** it:
   ```go
   params.StepConfig.Config["search_type"] = "news"
   ```
   and its own doc comment at :119 says *"3. Forces search_type to 'news'"*.
2. `WebSearchAction` (`web_search_action.go:62`) reads it and puts it in the Kafka
   payload (`:161`, `:217`).
3. The adapter unmarshals it — `internal/adapters/websearch/adapter.go:46`:
   ```go
   SearchType string `json:"search_type,omitempty"` // web, news, images
   ```
   and **logs it** at `:199`.
4. Then it calls the provider — `adapter.go:371`:
   ```go
   results, err := provider.Search(attemptCtx, query, numResults)
   ```
   and the interface (`providers/provider.go:7`) is:
   ```go
   type SearchProvider interface {
       Search(ctx context.Context, query string, numResults int) ([]SearchResult, error)
       Name() string
       IsAvailable() bool
   }
   ```

**The interface takes only `query` and `numResults`. There is no parameter for
`search_type` to travel in.** It is unmarshalled, logged, and dropped.

Confirmed at the other end — no provider references it at all:

```
grep -c "search_type\|SearchType" internal/adapters/websearch/providers/*.go
  duckduckgo.go:0   scrapingbee.go:0   firecrawl.go:0
```

### The options struct exists and is orphaned

`providers/provider.go:22` defines exactly what is needed, and **nothing constructs
it** (`grep "SearchOptions{"` → no hits anywhere):

```go
type SearchOptions struct {
    SearchType string   // web, news, images
    Language   string
    Region     string   // us, uk, etc.
    TimeRange  string   // day, week, month, year   <-- the recency control, also dead
    SafeSearch bool
    Domains    []string
}
```

So **`TimeRange` is dead too** — there is no recency constraint on any feed
anywhere, which is the direct explanation for 2034-dated forecast pages arriving in
a "news" feed.

## Why this survived so long

**A dead config key looks exactly like a live one.** Here it looked *better* than
live: a named action forces it, a doc comment states the guarantee, and the adapter
logs the value it received. Every artefact a reader would check says "news". The
only place the truth is visible is a function signature one hop further on.

It is also invisible from the data: the feed reports `success: true`, items arrive,
counts look healthy (10 per source), and `duplicate_of` is null. Only reading the
titles reveals it, and only if you know what news is supposed to look like.

## Fix candidates — ordered by what closes the door

1. **Widen the interface to carry the options struct that already exists.**
   `Search(ctx, query string, numResults int, opts SearchOptions)` — then each
   provider maps `SearchType`/`TimeRange` onto its own API (DuckDuckGo has news
   and time-range parameters; Firecrawl and ScrapingBee both expose search
   controls). This makes the bad state unrepresentable: the parameter can no
   longer be silently dropped because the call site must pass it. Three providers
   to update, all small.
2. **Same, but as a variadic option** (`Search(ctx, query, n, opts ...SearchOption)`)
   if the churn on existing callers matters — same effect, no signature break.
3. **Fail loudly instead of silently.** If `search_type != "web"` and no provider
   supports it, return an error rather than serving web results as news. Does not
   fix anything, but converts a silent wrong-content bug into a visible one, and
   is worth doing alongside 1 regardless.
4. **Route news to a real news source instead.** `api_news` (`FetchLLMNewsAction`)
   already exists, supports `hours_lookback` (default 12), and genuinely searches
   for recent material via xAI/OpenAI/Perplexity. For any site that needs actual
   news this is available **today with no code change** — but note it costs LLM
   calls per fetch and the webdesign.co.uk spec deliberately set
   `source_types: ["news_search"]` to avoid an unchosen xAI source.

**Not a fix: rewording the queries.** Two rounds of that on webdesign.co.uk
(SQL_p9 → SQL_p14) moved the *bias* — retiring agency-ranking bait was worth doing
on its own compliance grounds — but a web search cannot be turned into a news
index by phrasing. The 2034 forecast pages arrived under a query that named
acquisitions and mergers.

## How to verify a fix

Induce it, do not infer it:

1. Pick a source and set its query to something with an unambiguous recent event.
2. **Before:** results include undated evergreen pages and material dated years
   ahead. **After:** results carry `published_at` and cluster in the last days.
3. Assert on `content_feed_items.source_published_at` being populated and recent —
   it is currently **NULL on every row** for webdesign.co.uk, which is itself a
   usable regression check.

Do **not** verify by checking that the feed ingests, that counts are 10, or that
the run reports success. All three are true today and always have been.

## Related

- `bugs_open/116` — the link checks that have never run. Same family: machinery
  that exists, reports success, and does not do the thing.
- Standing note *"grep the config key before calling it a win — unknown config keys
  are silently ignored, so a dead key looks exactly like a live one."* This is the
  sharpest instance yet, because the key is not merely unknown to the consumer —
  it is **known, typed, logged, and structurally unable to arrive**.
- `docs/agent_docs/docs024_key_docs_latest/webdesign_couk/NOTES_webdesign_couk.md`,
  entries 2026-07-27/28, for the two rounds of query retuning and the measured
  before/after titles.
