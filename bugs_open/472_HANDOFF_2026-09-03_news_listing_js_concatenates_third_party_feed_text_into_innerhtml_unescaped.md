# 472 — `news-listing` / `latest-news` `js_content` concatenate third-party feed text into `innerHTML` unescaped

> ## ✅ FIXED IN THE LIBRARY 2026-09-04 16:28:35Z — migration 758 applied by hand
>
> Council `17a61f16` **APPROVED** (round 3). Both components now build DOM nodes and assign
> `textContent`, with `innerHTML` kept for clearing only, and hrefs routed through
> `safeExternalHref` / `safeInternalHref`.
>
> **Verified at the rows immediately after apply:** `still_concatenates=f`,
> `innerHTML = html` absent, `textContent` helper present, both href helpers present, and the
> `^\/(?!\/)` relative-href arm present — on both components.
>
> **Preconditions re-checked on the day, not inherited:** exactly 2 rows, **0 forked**, both
> still on the old form and untouched since July/August. Rehearsed under `BEGIN/ROLLBACK`
> against live state first (2 UPDATEs, 1 row each, all five post-conditions true), then applied.
>
> **Structural check on the applied scripts** (a character scanner tracking string, comment and
> regex-literal state): both **balanced** — curly/paren/bracket all 0 — both end on `})();`,
> and **neither contains its own `$js$` delimiter**, which is the failure that would truncate a
> script mid-function and still apply successfully. ⚠ **This is NOT a parse**: it cannot see a
> missing comma, a bad member expression, or a reserved word used as an identifier.
>
> **⚠ STILL OUTSTANDING — the browser check, and it is the load-bearing one.** The published
> `/tools/assets/*.js` do NOT change on apply; they republish on each site's **next render**.
> Verified 16:29Z: idea.uk and ai-agent-orchestration.com still serve the OLD 3,587-byte asset.
> So nothing has yet been demonstrated in a browser, and the failure mode that matters — the
> new script throwing and leaving an empty list, because `container.innerHTML = ""` runs before
> the loop — is only visible there. See §Verify.
>
> **No ledger row needed, and do not try:** `run-migrations.sh --record-only` **refuses** a
> `_HOLD.sql`, correctly — it is classed as an uppercase-suffixed sidecar the runner never
> applies, so it can never be replayed and a ledger row would be meaningless.
>
> **Rollback if the browser check fails:** `758_..._ROLLBACK.sql`, which restores both scripts
> verbatim and states plainly that running it re-opens this defect.

**Filed** 2026-09-03 by the `bugs_open/332` lane, found while fixing the markdown that reaches
the same surface. **Severity: LOW. Status: an exposure to close, NOT a live vulnerability —
and it must not be written up as one.**

## The mechanism, in plain terms first

Two components in the shared library build their HTML by joining strings together and then
hand the result to the browser as markup:

```js
html += "<p class=\"news-list-item-summary\">" + item.summary + "</p>";
...
container.innerHTML = html;
```

`item.summary` comes from `/data/news-archive.json`, which `loadNewsItems`
(`render_news_section_action.go:367-390`) writes straight from
`content_feed_items.source_summary` — third-party text scraped from other people's websites.
Nothing on that path escapes it. The server-rendered path *does* escape
(`projectNewsItems` calls `html.EscapeString`); the JSON path never has.

So if a feed summary ever contained HTML, the browser would treat it as markup rather than
text.

## Why it is an exposure and not an incident — measured, 2026-09-03

```sql
SELECT count(*) FILTER (WHERE source_summary ~ '<[A-Za-z/!]')                          AS has_markup,
       count(*) FILTER (WHERE source_summary ~* '<script|onerror=|onload=|javascript:') AS script_ish,
       count(*)                                                                         AS total
  FROM content_feed_items WHERE created_at > now() - interval '30 days';
-- 14 | 0 | 5863
```

**14 of 5,863** rows carry any HTML markup at all. **Zero** carry anything executable. **Zero**
of the 20 items served in boxingonline's archive JSON do. So there is nothing to exploit today
and no evidence anyone has tried.

**⚠ THE REASON THE NUMBER IS LOW IS A MECHANISM, AND IT CAN CHANGE WITHOUT ANYONE NOTICING.**
The RSS ingest path calls `stripHTML` (`feed_actions.go:248`); the web-search path does not.
That asymmetry is where the 14 come from, and **a new ingest source inherits whichever path it
is wired to**. So a future non-zero here is this same bug arriving, not a new one — do not
re-diagnose it.

## Scope — the whole component library was censused, and it is these two only

The `components` lane ran the check across every active component at this lane's request
(2026-09-03). **Quote the split, never the total**: a raw count of `innerHTML` reports 23
components and 21 of them are fine.

| | count | shape |
|---|---|---|
| **DEFECT** | **2** | fetch JSON, accumulate with `html +=`, assign, no escaping — `news-listing` (13 accumulations), `latest-news` (6) |
| SAFE — the best pattern in the library | 12 | the directory/tracker family (adoption-tracker, model-directory, protocol-tracker, the mortgage-lender / health-insurer / savings-provider directories and their `-listing` siblings). `innerHTML` used **only to clear** (`container.innerHTML = ""`); all data through `textContent` via an element helper. Zero concatenation. |
| SAFE — the fallback pattern | 1 | `webdesign-couk-header` genuinely fetches (`/search.json`) and concatenates, but has a complete `esc()` covering `& < > " '` applied to **every** interpolated value. Checked specifically for a gap, since a partial escape helper is this bug's usual shape. There isn't one. |
| SAFE — not data-backed | 8 | calculators interpolating locally computed numbers, or the user's own browser input through an `escHtml` built on `createTextNode`. No fetch, no DB text. |

They also checked a surface a script-level review would miss: whether any `js_content` carries
Go template placeholders, which would bake DB text into the script body rather than fetching
it. One component has one (`contact-block`) and it has **no `innerHTML` at all**.

## Two different counts, and they answer different questions

**[ADDED 2026-09-03, correcting a conflation of mine.]** I quoted "five sites" for most of this
lane. That is the count of hosts currently **serving the published JS asset**. The count of
sites **bound to the component** — which is what a `js_content` fix actually changes — is
**ten** for `news-listing` and **eight** for `latest-news`:

```sql
SELECT cc.function, count(DISTINCT pc.component_id) AS distinct_components_in_use,
       count(DISTINCT p.site_id) AS sites_using
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
 WHERE cc.function IN ('news-listing','latest-news') GROUP BY 1;
-- 2026-09-03: latest-news 1 component / 8 sites · news-listing 1 component / 10 sites
```

Both numbers are true. **Use the binding count for blast radius and the asset count for "who is
affected today"** — I had the smaller one to hand because I had spent the day probing served
assets, and never asked whether it answered the question I was putting it to.

That query also settles the question the council's guardian seat raised and I had never asked:
`content_components` carries `forked_from`, and **103 of 561 rows across 67 distinct functions
are forked**. Had either of these two been forked per site, a `WHERE function = '...'` UPDATE
would have patched one site and left the defect live on the rest. Both are shared and single-row
— and migration 758's verify block now **asserts** that rather than trusting it.

## Where it is live

`/tools/assets/news-listing.js` is a published **200** on boxingonline.ugg2.com, idea.uk,
ai-agent-orchestration.com, robot-hands.com and fundamentallyai.com.
`/tools/assets/latest-news.js` is 200 on idea.uk, ai-agent-orchestration.com and
robot-hands.com, and **404** on the other two — so the homepage arm is live on some sites and
not others and you cannot infer one from the other. Each verified against a
`/tools/assets/zzz-not-real.js` 404 control on the same host.

## Fix candidate — copy the library's own majority pattern, do not add an escape helper

The `components` lane's recommendation, adopted: **build the elements and assign
`textContent`, keeping `innerHTML` for clearing only** — the shape the 12 directory components
already use.

Preferred over adding an `esc()` because it **removes the class rather than filtering it**: an
escape helper regresses the moment someone adds a fourth interpolated field and forgets to
wrap it, which is precisely how this defect arrives. Where markup structure inside an item is
genuinely needed, `webdesign-couk-header`'s `esc()` is the in-library precedent to copy
verbatim rather than re-derive.

It is a `content_components.js_content` migration — **DB config, live the moment it applies,
no image and no roll**. That is why it is filed separately from `bugs_open/332`'s Go work: it
is an **escaping** defect, not a markdown one, and the two want different reviews.

## Verify

- Before: `curl -s <host>/tools/assets/news-listing.js | grep -c 'innerHTML = html'` → 1.
- After: that returns 0, and `grep -c 'textContent'` is non-zero.
- At the artefact, and this is the one that matters. A static `curl` **cannot** see it — the
  script replaces the server HTML on load (LANDMINES, *"The served news page HTML is
  OVERWRITTEN in the browser"*). **But it is not unverifiable, and an earlier draft of this file
  implied it was:** `browser-runner-adapter` is a running service (277 completed
  `acceptance_run` items, 2026-09-03), and `run_checks` reads the **live DOM after settle**.
  Three checks, taken **before and after in the same session** — a baseline from hours earlier
  lets another lane's rerender in as a confound:
  1. `selector_count` `article.news-list-item` — >0 proves the script ran and built the list.
  2. `selector_exists` `a.news-more-link[href^="/"]` — **the only possible instrument for this
     behaviour**: the served HTML of three affected homepages contains six matches for
     `news-more-link` and **zero rendered anchors** — all six are CSS rules, so the link exists
     nowhere but the client-built DOM.
  3. `has_visible_area` `#news-listing-items` — not redundant with (1);
     `run_checks_action.go:617` records three tools that measured 1146x0 while
     `selector_exists` passed all three.
  ⚠ `selector_exists`/`selector_count` pass on count > 0 and **fail on zero, with no
  expect-zero form** — so "there must be no `a.news-more-link[href='#']`" is inexpressible.
  Every assertion must be the presence of the RIGHT state, which is why (2) is `[href^="/"]`.
  Capture the screenshot of the resolving link: it is the one artefact that outlives the
  session, and nothing else in this chain — not the verify block, not the council, not either
  peer review — could see that behaviour.
- Negative control: an item whose summary contains `<b>x</b>` must render the characters, not
  bold text.

## Relations

`bugs_open/332` (the markdown on the same surface; its fix removes the markdown but not the
escaping gap) · `bugs_closed/184` · CQ-019 in the concept register · LANDMINES *"The served
news page HTML is OVERWRITTEN in the browser by `/data/news-archive.json`"* · the `components`
lane's census, 2026-09-03.
