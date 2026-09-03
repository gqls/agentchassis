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

---

## 2026-09-03 — PICKED UP (unowned) and RE-VERIFIED. The bug has INVERTED: intake has STOPPED, the served damage is LIVE, and nothing we did caused either

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_400_news_goto_urls/`. Unowned confirmed —
only the filing lane's own docs mention it and its §Who disclaims it (*"This lane consumed it and
is not fixing it"*), no commit subject addresses it, and no live session is named for it.

### 1. The intake has stopped — with a demand control, because a zero is worthless without one

```sql
SELECT created_at::date, count(*) AS all_new_items,
       count(*) FILTER (WHERE source_url LIKE 'https://www.google.com/goto%') AS goto
FROM content_feed_items WHERE created_at > now() - interval '12 days' GROUP BY 1 ORDER BY 1 DESC;
```

`[MEASURED 2026-09-03]` — the feed is **healthy and busy** throughout, and the goto form stops dead:

| day | all new items | goto |
|---|---|---|
| 09-03 → 08-29 | 166, 289, 238, 173, 134, 304 | **0, 0, 0, 0, 0, 0** |
| 08-28 → 08-23 | 322, 361, 352, 207, 142, 145 | 212, 315, 218, 84, 14, 43 |

**Six days, ~1,300 new items, zero occurrences.** The demand control is the whole point: without it
this reads as "fixed" and with it, it reads as "the shape stopped arriving".

### 2. ⚠ WE DID NOT CAUSE IT, AND IT CAN COME BACK SILENTLY

**The same sources are still producing.** Per `source_id`, the eight biggest goto producers have
each ingested **6–44 items since it stopped** with `last_any` of 2026-09-02, and **zero** goto rows.
So this is not "the sources went quiet" and not a config change — **the provider's response shape
changed upstream**. Nothing in this tree changed (`git log` over `internal/adapters/websearch/` and
the three feed actions for 08-26→08-31 shows only `201236b2a`, an unrelated phase-lock fix).

**Consequence, and it is the reason to still fix this:** an intake defect that stopped for reasons
outside our control is exactly the kind that returns, and **we have no guard and no detector** —
it would resume silently and the first sign would be another lane noticing links on a page.

### 3. The served damage IS live, right now, and no fix candidate in this file addresses it

`[MEASURED 2026-09-03, curl'd]` — `idea.uk/data/latest-news.json` serves **2 of 6**;
`mortgagecalculator.co.uk/data/latest-news.json` serves **1 of 6**. Stored: **1,378 rows across 11
sites**, newest 2026-08-28.

**This file's three candidates all address INTAKE.** Candidate 3 (filter at render) is correctly
rejected as hiding the symptom — but a **repair of the stored rows is not a render filter**, and it
is missing from the file. As written, all three candidates could ship and every one of those 1,378
rows would still serve a `google.com/goto` link for as long as it stays in the window.

### 4. Two facts that change the fix shape, both established first-hand

**(a) "Decode it" is NOT AVAILABLE.** Candidate 1 offers *"either decode it or follow it once"*. The
token is **opaque** — 250 chars, and base64-decoding at every padding yields no printable substring
(it is a protobuf/encrypted blob, not an encoded URL). Verified on a live row. **Only the follow arm
exists**, so the candidate is one option, not two.

**(b) The redirect STILL RESOLVES, so the backlog is RECOVERABLE — and one hop is enough.**
A single request, no `-L`, no body:

```bash
curl -s -o /dev/null -w "%{http_code} %{redirect_url}" -A "Mozilla/5.0" "<goto url>"
```

`[MEASURED 2026-09-03]` **3 of 3** returned **302** with the real publisher in `Location`:
`hpcwire.com/…`, `fortune.com/2026/08/27/anthropic-…`, `nature.com/articles/s44360-026-00190-2`.

⚠ **Do NOT follow to the final page.** Following with `-L` returned **403** — the publisher blocks
our agent — *after* the correct target had already been captured. So a fix that requires a 200
would wrongly discard a recoverable row. **Read `Location` from the 302 and stop.** That also means
no publisher is ever fetched: one hop, to Google, no body.

### 5. Revised fix shape (supersedes §Fix candidates for the intake half, and adds the missing half)

1. **Unwrap at the bridge** (`feed_normalize_action.go`) rather than at the provider. Same
   door-closing as candidate 1 — it is upstream of `source_url` and therefore of the dedup index —
   but it catches **every** provider, and the shape is provider-behaviour, not scrapingbee-specific.
   Resolve via one non-following request; on any failure keep the goto URL rather than dropping the
   item (a working link beats no item), and **count the failure** so the arm cannot go quiet.
2. **A detector, because the recurrence is the real risk.** The daily-check convention: count rows
   matching the goto form in the last window and fail on non-zero, with a demand control (total new
   rows > 0) so a dead feed cannot read as clean. Without this, a silent resumption is invisible.
3. **Repair the 1,378 stored rows** — one-hop resolve, `UPDATE content_feed_items SET source_url`.
   ⚠ **`idx_cfi_dedup` is a partial UNIQUE index on `source_url`**, so a repair can collide with an
   existing direct-URL row for the same story — which is this file's own `[UNMEASURED]` duplicate
   question, arriving as a constraint on the repair. Measure the collision set BEFORE updating, and
   decide merge-vs-skip; do not let the UPDATE discover it.

### 6. What is NOT done, stated plainly

Nothing is committed. No code has been written. The above is diagnosis and design only, handed off
deliberately rather than half-implemented — the fix is in council scope (`platform/`, `internal/`)
and wants its own round. The lane's `HANDOFF_2026-09-03_start_here.md` is the entry point.
