# 400 — the news pipeline stores and SERVES Google redirect URLs as article links

**Filed:** 2026-08-25 · **By:** `idea_uk_vm_site` lane (session "idea.uk") · **Status:** OPEN —
damage measured at two served artefacts; mechanism located to the component, not read to the line.
**Severity:** low-medium — links function (via Google's redirect) but name no publisher, add a
tracking hop, and undermine the "real articles with real links" property the `news_search` source
type exists to provide. Fleet-wide.
**Class:** provider data passed through unnormalised.

## Damage, measured `[MEASURED 2026-08-25]`

- `https://idea.uk/data/latest-news.json` (16:31Z): **3 of 6** item `url`s are
  `https://www.google.com/goto?url=CAES…` — an opaque token, not the publisher.
- `https://mortgagecalculator.co.uk/data/latest-news.json`: **1** such link served.
  `fundamentallyai.com` and `relojistas.com`: 0 served today.
- Stored: idea.uk **9 of 12** `content_feed_items.source_url` in that form;
  fundamentallyai **33 of 372** `relevant` rows. So the form both enters and survives triage.

## Mechanism, located (and where locating stopped)

`news_search` sources are fetched by `fetch_news_search`
(`platform/orchestration/actions/feed_fetch_async_actions.go`) via the websearch adapter; the
Google-news provider is `internal/adapters/websearch/providers/scrapingbee.go` (`search_type=news`,
parses `news_results`). `feed_normalize_action.go` bridges results into
`content_feed_items.source_url`; `render_news_section_action.go` serves `source_url` as-is.
**No `goto` handling exists anywhere in the tree** (`grep -rn 'goto' internal/adapters/websearch
platform/orchestration/actions` → nothing URL-related). So the redirect form arrives in the
provider's response data and nothing unwraps it.

**Declared first-hand substitute (2026-07-31 ruling):** the damage is verified at the artefacts
and in the DB first-hand; the mechanism sentence above is `[INFERRED]` from the grep absence plus
the data shape — the exact provider field was not traced line-by-line. A `090` run would earn the
citation if the fixing thread wants it; the fix does not depend on which upstream field carries it.

## Consequences beyond cosmetics

- Dedup keys on `source_url` (partial unique index `idx_cfi_dedup`,
  `feed_actions.go:896-905`) — one story fetched once as a goto link and once direct is **two
  rows**, so the 6-item snippet can carry a duplicate story. `[UNMEASURED]` whether it has
  happened yet.
- The redirect target is Google's, not ours — link rot/geo behaviour is out of our control, and a
  reader hovering sees `google.com`, on sites whose value proposition includes source honesty.

## Fix candidates, ordered by what closes the door

1. **Unwrap at the provider** (`scrapingbee.go`): when a result URL matches the goto form, either
   decode it or follow it once (bounded HEAD/GET, no body) and store the final URL. Closes the
   door for every consumer, including dedup.
2. **Normalise at the bridge** (`feed_normalize_action.go`): same unwrap, catches every provider.
   Equivalent door-closing; slightly later, so raw rows still enter logs.
3. Filter/resolve at render — hides the symptom, leaves dedup broken. Do not.

## How to verify a fix

While fetches continue post-roll:
`SELECT count(*) FROM content_feed_items WHERE source_url LIKE 'https://www.google.com/goto%' AND created_at > '<fix date>';` → 0,
with a demand control (total new rows > 0); and the served JSONs carry publisher hosts only.

## Who

The news-feed machinery's lane (adjacent: `bugs_open/316`,
`scripts/initial_messages/100_news_feed_ingester/`). This lane consumed it and is not fixing it.
