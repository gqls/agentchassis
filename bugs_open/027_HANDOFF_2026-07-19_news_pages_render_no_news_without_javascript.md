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
- > **VERIFIED, and the two halves are NOT alike — do not generalise finding 3.** I first wrote
  > that `news-listing.js` was unverified and to check before assuming it matched. Checked it:
  > **it does not.** Where `latest-news.js` leaves the DOM alone when it has nothing, the
  > listing script *overwrites the container in both failure modes*:
  > ```js
  > if (!data.items || data.items.length === 0) {
  >   container.innerHTML = "<p class=\"news-listing-empty\">No news items available yet...</p>"; }
  > .catch(function(err) {
  >   container.innerHTML = "<p class=\"news-listing-empty\">Unable to load news...</p>"; });
  > ```
  > So on the archive page, server-rendered items would be **destroyed** by the very script
  > meant to enhance them — on an empty feed, on a 404, on any B2 blip. The homepage half needs
  > no JS change; **the archive half needs the JS fixed first, or it is a regression, not a fix.**
  > The asset is byte-identical across robot-hands, relojistas and gaswholesalers (one md5), so
  > one edit covers the fleet. This is exactly the fix-one-branch-and-call-it-done failure this
  > platform keeps hitting, and it was one `curl` away from shipping.
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

---

## Addendum 2026-07-20 (relojistas thread) — STATUS: interim fix LIVE in v1.0.1140; proper fix designed, not yet built

An implementation of fix option 1 was committed (`1005e1af2`: migration 178 JS guard +
`persistNewsSectionHTML` string-injection into `rendered_html`) and rode another thread's
sweep build into production (`v1.0.1140`, pod-verified 2026-07-20 ~18:00Z). **It is the
WRONG mechanism and is scheduled for removal:**

- The council gate returned REVISE (correlation `4b91237a`): render_guardian showed a
  scoped rerender regenerates `rendered_html` from `html_template` + `content_data`, so
  injected news that lives in neither is silently wiped — recreating this bug
  intermittently.
- A guidelines check confirmed it violates 003's source-of-truth contract verbatim ("HTML
  patching was rejected as an edit mechanism").

**Interim behaviour to expect:** on each `render_news_section` run, news pages' components
gain server-rendered `<article>` markup in the DB, which reaches live pages on their next
deploy and may later vanish again on a scoped rerender. So this bug's symptom will
INTERMITTENTLY appear fixed. Do not close 027 on a `curl` showing articles while the
injection mechanism is the thing producing them.

**The contract-compliant design (agreed, not yet built):** declare the items as
`query.latest_news` / `query.news_archive` sources in the two components' `input_schema`;
add those resolvers to `queryresolve` over `content_feed_items`; render items in the
`html_template` (`{{if .items}}{{range …}}` — nil guard mandatory); deliver via the
existing `page_rerender` / `section_data_resolved` light path. Then news lives in
`content_data`, rerenders REFRESH it instead of wiping it, and the injection machinery is
deleted. Migration 178 (the JS guard) stays — it is correct under either mechanism.

**Close criteria updated:** 027 closes when the query-source route is live in an image,
the injection call sites are REMOVED, and a from-curl check shows `<article>` elements that
survive a scoped rerender of the same page.
