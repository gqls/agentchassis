# 027 — every news page on the platform renders zero news without JavaScript

**Filed 2026-07-19** (relojistas thread, at the owner's request). **Status: OPEN.**
Fleet-wide, affects **all three** sites that have a news page. Not a crash — the pages
return 200 and look fine in a browser. The defect is that the news itself exists only
client-side, so any consumer that does not execute JavaScript sees a news portal with no
news on it.

Filed here because there is no better home for it at present. Note it is arguably a
**design/architecture decision to revisit** rather than a coding mistake — the client-side
fetch was a reasonable way to get fresh news onto a static site. It is recorded as a bug
because its consequence is a real, measurable loss on live sites.

## Reproduction — all three affected sites

```bash
for u in https://relojistas.com/noticias/ https://gaswholesalers.com/news.html https://robot-hands.com/news.html; do
  curl -s "$u" | grep -c 'news-listing-loading'   # 1 = placeholder present
  curl -s "$u" | grep -ciE '<article|news-item'   # 0 = zero server-rendered items
done
```

Result, 2026-07-19: every one returns HTTP 200, placeholder present, **zero** server-rendered
news items. Confirmed live on all three.

## Mechanism

Deliberate and consistent, which is why it is a design question rather than an oversight:

1. `render_news_section_action.go:341-363` writes `data/latest-news.json` and
   `data/news-archive.json`. It is deterministic link rendering — no HTML is produced.
2. The two shared components that display news — `news-listing` and `latest-news`
   (`content_components`) — are both `render_mode='template'` with **empty `data_sources`**.
   Neither is server-rendered from the feed data.
3. `news-listing`'s template emits a placeholder plus
   `<script src="/tools/assets/news-listing.js"></script>`; that script `fetch()`es
   `/data/news-archive.json` and fills the list **after paint**.

So the pipeline ingests, triages, curates and publishes news correctly — and then hands the
last step to the browser.

## Who this actually costs

- **Search crawlers that do not execute JS** index a news page whose visible content is a
  loading message. On relojistas the loading message is also in English on a Spanish site
  (that half is `bugs_open/026`). For a site whose entire purpose is news, this is the
  content that matters most going unindexed.
- **Feed readers, previewers, and link-unfurlers** (Slack, WhatsApp, social cards) fetch
  HTML and do not run scripts — they see nothing.
- **relojistas specifically:** we have just proved the mission metric by reactivating a
  legacy RSS feed, and the measured traffic is *majority crawler* (Googlebot, meta-webindexer,
  Applebot — see the reactivation measurement in
  `docs024_key_docs_latest/traffic_probe/relojistas_rebuild_running_notes.md`). The audience
  that actually arrives is disproportionately the audience this defect blinds.

Worth stating the limit of the claim honestly: **Googlebot does render JavaScript**, so this
is not a total indexing loss for Google. It is a delay and a risk (render budget, deferred
second-pass indexing), and a total loss for the many consumers that never run JS at all.
`/feed.xml` and `/external.php?type=RSS2` are unaffected — they are server-rendered and
complete, which is part of why the reactivation worked.

## Fix candidates

1. **Server-render the news list at build time** (recommended). The data already exists in
   `content_feed_items` at render time — `render_news_section_action.go` is already reading
   it to write the JSON. Emit the item markup into the component's HTML in the same pass and
   keep the JSON + JS as a progressive-enhancement refresh. Highest fidelity, no loss of
   freshness, and it makes the "empty section" discovery check meaningful again instead of a
   known false positive (see 026).
2. **Hybrid: server-render the first N, lazy-load the rest.** Cheaper if archive pages are
   long; keeps the page small while making the important items indexable.
3. **Re-render on feed refresh.** The 6-hourly `content-feed-trigger` already runs; have it
   queue a rerender of news pages so the static HTML tracks the feed. More moving parts, and
   it couples page freshness to the deploy pipeline.
4. **Accept it, and compensate** — ensure `/feed.xml` is discoverable (`<link rel="alternate">`)
   and sitemapped. Not a fix, but the honest do-nothing option if the crawler cost is judged
   acceptable.

Option 1 is the structural fix and matches the platform convention of preferring structural
fixes over patches.

## How to verify a fix

The reproduction above must show **non-zero** server-rendered items with JavaScript never
executing — i.e. from `curl`, not from a browser. Trust the fetched HTML, not the rendered
page and not a `complete` status.

## Related

- `bugs_open/026` — the same `news-listing` component hardcodes an English placeholder and
  drops a required `<h1>`. Same component, and 026's fix pass is the natural place to do
  this one.
- The `empty_section` discovery check flags these sections as empty. Today that is a false
  positive. Under fix option 1 it would become a true signal again — a small extra reason to
  prefer it.

---

## Addendum 2026-07-19 (vetcomparison thread) — feasibility of fix option 1, measured

The owner asked how hard it would actually be to server-render the feed. I read the code
rather than estimating. **Option 1 is cheaper than this file implies**, for one reason nobody
had checked: the client JS is already progressive-enhancement-safe.

**1. There is no declarative binding engine to reuse — `data_sources` is dead metadata.**
`content_components.data_sources` is populated on four components and one carries
`render_mode='go_template'`. Both are read by **zero** lines of Go:

```bash
grep -rn "data_sources\|DataSources\|go_template" --include=*.go .   # → no matches, repo-wide
```

So "populate data_sources and let the template engine bind it" is not available — that engine
does not exist. Option 1 has to be implemented where the data already is. (Separately: those
four rows are misleading metadata that reads like a working feature. Worth a tidy-up.)

**2. The data is already in hand, in the right function, in one file.**
`RenderNewsSectionAction` (`platform/orchestration/actions/render_news_section_action.go`)
already does all the hard parts:
- `loadNewsItems` (:340) returns `[]newsJSONItem` — title, summary, url, published_at, source
  name, topics — straight from `content_feed_items` with no LLM in the path;
- it *already locates the component row*, joining `pages` → `page_components` →
  `content_components` on `cc.function = 'latest-news'` (:145-152), to read the headline out of
  `content_data`.

So the change is: render `[]newsJSONItem` → HTML, and write it to that row's `rendered_html`.
No new action, no new orchestration step, no schema change.

**3. The decisive finding: `latest-news.js` needs no change at all.**
The deployed script only overwrites the container when the fetch actually returned items, and
swallows failures:

```js
if (data.items && data.items.length > 0) { container.innerHTML = ... }   // else: leaves DOM alone
.catch(function() {});                                                   // failure keeps server HTML
```

That is exactly the hybrid contract this file's option 1 asks for, already implemented by
accident. Server-rendered markup survives an empty feed, a 404, and an offline B2 — and is
replaced by fresher items when the fetch succeeds. **No JS edit, no risk of double-rendering.**

**4. The markup is trivial to port.** The Go renderer must emit what the JS emits — one
`<article class="news-card">` per item (`news-card-title` + link, `news-card-summary`,
`news-card-meta` with `news-source` and `news-date`). ~40 lines with `html/template` escaping.
Use `html/template`, not string concatenation: these are third-party titles and summaries.

**Honest scope and unknowns:**
- **Two components, not one.** `latest-news` (homepage snippet) and `news-listing` (archive
  page). I verified the progressive-enhancement property for `latest-news.js` only —
  **`news-listing.js` is unverified**; check it before assuming the same. Fixing one and
  calling it done is the failure mode this platform keeps hitting.
- **Freshness moves to the deploy cadence.** `rendered_html` only reaches the live site on the
  next rerender+deploy, whereas the JSON deploys as a file. The hybrid is what saves this: HTML
  for crawlers, JS refresh for currency. Do not drop the JSON.
- **Locking is an open risk.** `page_components.lock_type='permanent'` and
  `component_write_guard.go` (`componentRegressionIssues`) both sit near this write path. A
  locked news component could reject or silently skip the write. Verify against a locked row.

**Effort:** one file, one render function, one UPDATE, plus tests — roughly half a day
including the `news-listing` half and a locked-component check. The expensive unknowns above
are checks, not design.

**How to verify** is unchanged and still the only thing that counts: `curl` the page, with JS
never executing, and count server-rendered `<article>` elements.
