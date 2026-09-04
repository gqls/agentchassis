# BUG 483 — every search-fed news card shows visitors the internal query string where the publisher's name belongs

**Filed** 2026-09-04 by the `farmerinsurance_uk` lane.
**Status** OPEN. Live on the served homepage of at least two sites; 13 sites carry the ingredient.
**Owner note:** the news feed subsystem is `bugfix_316_news_feed_ordering`'s. This is filed, not
fixed, and a CONTRIB points at it from their directory.

## 1. Symptom, at the artefact

`farmerinsurance.uk/index.html` renders, six times:

```html
<span class="news-source">News Search: insurance market</span>
```

That span sits where a publisher's name belongs, beside the article's headline and date. The
article behind it is from Insurance Journal; the visitor is shown the site's own internal search
keyword instead. `[MEASURED 2026-09-04, curl, with a 404 control on an invented path.]`

**The disconfirming control, which is what makes this a mechanism and not a farmer quirk:**

| site | what the news-source span serves | its source kind |
|---|---|---|
| `advertise.co.uk` | `WebProNews` — a real publisher | named feed source |
| `ai-agent-orchestration.com` | `News Search: artificial intelligence` | `news_search` |
| `farmerinsurance.uk` | `News Search: insurance market` | `news_search` |

So the renderer is not broken — it prints what it is given, and for a named feed source that is a
publisher. (`boxingonline.com` could not be resolved from this workstation, so it is NOT counted
either way.)

## 2. Mechanism — two correct-looking halves that compose into an internal string on a public page

1. `platform/orchestration/actions/seed_content_sources_action.go:262` names a search source
   after its keyword: `name := fmt.Sprintf("News Search: %s", keyword)`. Reasonable — it is what
   an operator inspecting `content_sources` wants to see.
2. The feed pipeline resolves a card's source label from that same column:
   `COALESCE(cs.name, 'unknown') as source_name`
   (`feed_triage_actions.go:359`, and identically `feed_event_extraction_actions.go:90`).
   Also reasonable — for an RSS source the name IS the publisher.

Neither half is wrong on its own. Composed, a field whose audience is an operator is rendered to
the public, and the string it carries is the query we ran.

## 3. Population `[MEASURED 2026-09-04]`

```sql
SELECT count(*) FILTER (WHERE name LIKE 'News Search:%') AS named_after_query,
       count(*) AS active_news_search_sources,
       count(DISTINCT site_id) AS sites
FROM content_sources WHERE source_type='news_search' AND is_active;
-- 56 of 57, across 13 sites
```

Every one of those 56 is a string a visitor may be shown. How many sites actually RENDER one
today is not established here — two are confirmed by curl (farmer, ai-agent-orchestration),
advertise.co.uk is confirmed NOT to, and the remaining ten are unchecked. **Do not quote "13
sites are showing this" from this file** — 13 is the population that carries the ingredient.

## 4. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Give the card the article's own publisher.** The fetched item knows its host
   (`prnewswire.com`, `insurancejournal.com`); deriving a display name from it removes the
   operator field from the public surface entirely, so no naming convention can leak again.
2. **Split the columns**: keep `content_sources.name` as the operator label and add a
   `display_name` the renderer reads, defaulted from the publisher where one exists and NULL —
   rendering nothing — where it does not. Narrower, but it leaves two fields that can drift.
3. **Stop putting "News Search: " in the name.** Cheapest and worst: it makes the leak less
   obviously an artefact ("insurance market" beside a headline still is not a source) and leaves
   the composition intact for the next field.
4. Suppress the span when the source is `news_search`. Honest, loses attribution that a reader
   arguably wants, and does nothing about (1)'s real gap.

**(1) is the fix.** (2) is the containable version if the fetched item turns out not to carry a
usable publisher for every provider.

## 5. Adjacent, same site, same section — do not conflate
- The **region defect** (owner finding 08-31): the same four sources are regionless
  ("insurance market", "insurance regulation", "claims", "premiums"), so a UK farm site's
  homepage carries US corporate M&A news. That is the news lane's existing CONTRIB and a
  different fix.
- **Unverified numbers**: farmer's `claims_unverified` item names three unregistered numbers on
  the homepage — 84%, 65%, 75% — and all three come from a US trade-press news snippet in this
  same section. The news feed is importing unattributable statistics into the claims layer.
- The **malformed Google redirect URLs** the council seats filed against this section
  `[MEASURED 2026-09-04]` are **fixed** — every news link on farmer's homepage now points at a
  real publisher host.

## 6. How to verify a fix
Curl a site whose news comes from `news_search` and assert the `news-source` span contains no
`News Search:` prefix and no bare query string, with `advertise.co.uk` (a named feed source) as
the positive control that a real publisher still renders.

---

## 7. APPENDED 2026-09-04 — a uniqueness constraint that rules candidate 3 out, from the `feed lane`

The `news_feed_ingestion` lane supplied a fact that changes the ordering in §4, and it is the kind
that makes a fix look green while doing nothing:

> the source name is derived from the keyword and **`idx_cs_site_name` is UNIQUE on
> `(site_id, name)` with `ON CONFLICT DO NOTHING`.** So editing `vertical_keywords` alone changes
> nothing for the 57 existing sources and verifies green — retuning is DELETE plus re-insert.

Consequences for this bug:

- **Candidate 3 ("stop putting `News Search: ` in the name") is now firmly the wrong fix**, not
  merely the cheapest-and-worst. Changing the naming in
  `seed_content_sources_action.go:262` cannot rename an existing source: the insert conflicts on
  `(site_id, name)` and does nothing. It would apply only to sources seeded afterwards, so the
  56 live ones keep their prefix while a census of the CODE reads as fixed. A rename applied
  deliberately (UPDATE, not re-seed) would work but re-points a column other machinery keys on.
- **Candidates 1 and 2 are unaffected and stay in that order.** Deriving the card's publisher from
  the item's own `source_url` host (1) touches no unique key at all, and is — in the feed lane's
  words — "what a reader actually wants". A separate `display_name` column (2) is the containable
  version if some provider's items turn out to carry no usable host.

Recorded with credit rather than merged into §4 silently, because the constraint is theirs and a
later reader should be able to see which half of this file came from the subsystem's owner.
