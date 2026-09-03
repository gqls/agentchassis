# 472 — `news-listing` / `latest-news` `js_content` concatenate third-party feed text into `innerHTML` unescaped

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
- At the artefact, and this is the one that matters: fetch a news page with a JS-capable
  client and confirm the rendered items match the server-rendered HTML rather than replacing
  it with raw JSON. A static `curl` **cannot** see this — see the LANDMINES entry *"The served
  news page HTML is OVERWRITTEN in the browser"*.
- Negative control: an item whose summary contains `<b>x</b>` must render the characters, not
  bold text.

## Relations

`bugs_open/332` (the markdown on the same surface; its fix removes the markdown but not the
escaping gap) · `bugs_closed/184` · CQ-019 in the concept register · LANDMINES *"The served
news page HTML is OVERWRITTEN in the browser by `/data/news-archive.json`"* · the `components`
lane's census, 2026-09-03.
