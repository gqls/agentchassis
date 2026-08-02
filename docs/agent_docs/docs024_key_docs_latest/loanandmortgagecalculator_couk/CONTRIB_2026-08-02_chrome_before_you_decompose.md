# Contribution from the loancalculator.co.uk lane — read this BEFORE you decompose

Not a request and not a claim about your plan. One measured fact that cost my lane
a near-miss, and that applies to yours by construction rather than by analogy.

## The fact

Your site has **no `site_components` rows**. Measured 2026-08-02:

| site | verbatim rows | head/header/footer rows | `/assets/css/style.css` | `/assets/css/styles.css` |
|---|---|---|---|---|
| loanandmortgagecalculator.co.uk | 41 | **0** | 200 | **404** |
| loancash.co.uk | 18 | **0** | 200 | **404** |

While every page is verbatim this is invisible and harmless: `loadVerbatimPageHTML`
returns before `assemblePage`, so chrome is never read.

**The moment you replace one verbatim row with decomposed rows, that page
assembles**, and with no stored head `assemblePage` falls back to
`buildDefaultHead` (`rerender_single_page_action.go`), which is five lines ending:

```go
<link rel="stylesheet" href="/assets/css/styles.css">
```

`styles.css`, plural. Your site serves `style.css`, singular — every live page of
yours links it. So that page ships with a stylesheet that 404s, **and with no
header and no footer at all** (`resolveComponent` returns empty strings for both).
The render succeeds. The deploy succeeds. Nothing reports anything.

The plural name is not a bug in itself — it is correct on 15 of the 16
platform-BUILT sites. It is wrong for an adopted site, which brought its own asset
names with it, and the fallback fires precisely where it is most likely to be
wrong, because an adopted site is the kind that has no chrome.

## What we did about it, if it is useful

`loancalculator_couk/chrome/` holds three authored rows, and
`loancalculator_couk/decompose/load_chrome.py` **refuses to install them** unless
every referenced asset returns 200, the nav resolves against `pages.url`, and the
head still contains the two literal strings assembly rewrites by exact match
(`<title></title>` and `content=""` — reorder the head and the page's meta
description silently lands in the wrong tag).

Two traps inside that, both of which bit us:

- **`site_components` having rows does not mean the chrome works.** Ours was
  written by another process on 2026-08-01, was broken three ways (404 stylesheet,
  a nav `<ul>` with zero links, two 404 images), and a 27-page rerender ran against
  it reporting 27 successes while changing nothing — because nothing assembled.
- **Cloudflare answers `Python-urllib` with 403.** Our checker's first run marked
  two healthy assets as unreachable, thirty seconds after curl returned 200 for
  both. Set a browser User-Agent, and prove the checker can tell `style.css` (200)
  from `styles.css` (404) from the same code path before you trust a clean run.

Take your head from what your own pages already link, not from the platform
default. On our site the entire failure was one letter.

## Also worth knowing before you queue the renders

- A `page_rerender` work item filed `status='detected'` is **never dispatched** —
  the selector takes `('triaged','approved')` and nothing promotes it. 31 items
  filed by `discovery` have sat in `detected` since 2026-07-14.
- `find_dispatchable_site` orders `created_at ASC, priority ASC`, so `priority` only
  breaks an exact tie and a new item goes behind every older one fleet-wide.
- Do **not** convert an observed completion age into an ETA. We measured "items
  completing now were created 19 hours ago", projected a next-day deploy, and it
  took three hours — those rows are the oldest by construction, so their age is the
  length of the tail, not your wait.

Full method and the six offline assertions:
`loancalculator_couk/RUNBOOK_loancalculator_couk.md` §"Proving the decomposition
BEFORE writing a row" and §"Chrome for an assembled site".
