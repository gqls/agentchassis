# 101 — `scrape_web` silently ignores four config keys, so two workflows describe a crawl they never perform

**Filed** 2026-07-26 by session "bugfix 061" (vetcomparison workstream).
**Status** OPEN. Unowned. Low severity, **high misleading-power** — it caused a false claim in a
commit message within an hour of being read (`WRONG_CALLS.md`, 2026-07-26).

---

## Symptom

`vet-practice-verifier`'s `scrape_website` step is configured like this:

```json
{ "action": "scrape_web",
  "config": { "max_pages": 3, "extract_mode": "text",
              "url_field": "business_record.business.website_url",
              "fallback_url_field": "search_results.results.0.url",
              "follow_links": ["fees","prices","about","team","contact","services"] } }
```

It reads as: fetch up to three pages, follow the fees/prices/about/team/contact/services links,
extract text, and if the practice has no website fall back to the top search result.

**It fetches the home page. Once. That is all it has ever done.**

## Root cause

`WebscrapeAction` (`platform/orchestration/actions/webscrape_actions.go:27-147`) reads exactly
five config keys — `url_field`, `url`, `action`, `upload_results`, `scrape_config` — resolves a
single URL, and dispatches one request to the webscrape adapter. The other four keys are never
read by anything:

```
$ grep -rn "follow_links" --include=*.go .          # no hits
$ grep -rn "extract_mode\|fallback_url_field" --include=*.go .   # no hits
$ grep -rn "max_pages" --include=*.go . | grep -i webscrape      # no hits
```

Unknown config keys are silently ignored, so an aspirational or stale key is **indistinguishable
by inspection from a live one**. The config reads as a specification of behaviour while being
evidence of nothing.

## Blast radius — two live agents

```sql
SELECT ad.type, e.k AS step,
       (v->'config' ? 'max_pages') AS max_pages, (v->'config' ? 'follow_links') AS follow_links,
       (v->'config' ? 'extract_mode') AS extract_mode, (v->'config' ? 'fallback_url_field') AS fallback_url
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND v->>'action' = 'scrape_web'
  AND (v->'config' ?| array['max_pages','follow_links','extract_mode','fallback_url_field']);
```
| type | step | max_pages | follow_links | extract_mode | fallback_url |
|---|---|---|---|---|---|
| `domain-research-classifier` | `scrape_site` | t | t | t | f |
| `vet-practice-verifier` | `scrape_website` | t | t | t | t |

`domain-research-classifier` is **not this workstream's** — whoever owns it is classifying domains
on home-page text alone while its config says otherwise.

## Consequences

1. **Measured cost, vetcomparison** (read-only probe, n=100, deterministic sample). Company-number
   extraction against the home page only: **22/100 (22%)**. Reading legal/terms pages as the config
   implies: **30/100 (30%)** — the 8 extra hits were on `/terms` (5), `/privacy` (2) and
   `/terms-and-conditions` (1). Note the configured `follow_links` list
   (`fees, prices, about, team, contact, services`) contains **no legal page**, so even if it were
   honoured it would not capture these; the list itself is wrong for its purpose.

   > ⚠️ **READ THIS BEFORE IMPLEMENTING candidate 2 — it may not be sufficient.**
   > `[UNSETTLED]` The probe fetched raw HTML. Production goes through Firecrawl, and
   > `FirecrawlScrapingProvider.Scrape` sets `onlyMainContent := false` then adds the key **only
   > when true** (`firecrawl.go:77-111`) — so a caller passing no `scrape_config` (the vet verifier
   > passes none) has the key **omitted**, and Firecrawl applies its own default. **If that default
   > strips nav/footer, production sees less than the probe did and adding page fetches will not
   > help**, because company numbers live in footers.
   > Evidence is ambiguous: of 2,452 stored Firecrawl samples
   > (`med_scrape_evidence.markdown_content`), **75% retain footer nav text** (suggesting footers
   > survive) but **0 contain company-registration text** (equally explained by those being
   > retailer product pages).
   > **Settle it first, in one run:** do a single real verification and read what came back. If
   > extraction is dropping footers, the fix is `only_main_content: false` reaching the payload —
   > note the current code cannot express that, since false means "omit".
2. **`fallback_url_field` is a silent dead path.** "No website → use the top search result" never
   fires. Moot today (all 3,419 `businesses` rows carry a `website_url`) but it is a trap for
   anyone who assumes the fallback protects them.
3. **It actively misleads readers.** Documented instance: a session (this one) read the config,
   concluded that widening `follow_links` was a free config-only win, and wrote that into NOTES,
   README, a PLAN correction and commit `096276f90` before catching it.

## Fix candidates — ordered by what closes the door

1. **Reject unknown config keys** for registered actions (validate step config against a per-action
   allow-list at seed/def-update time, fail loudly). Makes the bad state unrepresentable and fixes
   the whole class, not these four keys. Biggest change; the only one that stops the next instance.
2. **Implement the keys** in `WebscrapeAction` — honour `max_pages`/`follow_links`/`extract_mode`,
   and `fallback_url_field` when the primary resolves empty. Delivers the 16%→28% and makes the
   config true. Go change; **bundle with `bugs_open/100`**, same step, same roll.
3. **Delete the inert keys** from both agent definitions. Config-only, live immediately, honest —
   but loses the stated intent and leaves the silent-ignore class in place for the next author.

Prefer 1 + 2 together. 3 alone is a tidy-up that removes the evidence of the problem rather than
the problem — though it is a reasonable immediate step if 2 is not scheduled soon.

## How to verify a fix

- For candidate 2, do **not** trust a green run. Pick a practice whose number is on `/terms` and
  absent from the home page (`Ark Veterinary Centre`, `05886364`, verified 2026-07-26) — a
  **negative control before the fix and a positive after**, on the same practice.
- For candidate 1, the check is that a definition carrying a bogus key is *refused*. Seed one and
  watch it fail; a silent accept is the bug restating itself.

## Related

- `bugs_open/100` — the provenance defect on the same step; fix in the same council round.
- `WRONG_CALLS.md` 2026-07-26 — "I recommended a config change without checking the config key is
  read", with the one-command check.
- Same family as the fleet's inert-layer entries: **the artefact exists, so the capability is
  assumed.**
