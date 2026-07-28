# 101 — `scrape_web` silently ignores four config keys, so two workflows describe a crawl they never perform

> ## STATUS 2026-07-28 — candidates 1 + 2 COMMITTED (`2ebabf2ca`); INERT until TWO images roll
>
> Taken by the "bugsearch" thread, bundled with `bugs_open/100` as this file instructs.
> Council submitted: `SUBMISSION_CORR=f4cf0aab-5a08-4475-91ea-fa831cff323c`.
>
> **The `[UNSETTLED]` box below is SETTLED, and the answer was a third bug.**
> `FirecrawlScrapingProvider.Scrape` read `only_main_content` and then emitted the key
> **only when true**, so `only_main_content: false` was inexpressible and Firecrawl's
> documented default (`true` — strips headers, navs, footers) applied instead. The
> `/crawl` path in the same file was always correct; they now agree. **Three live steps
> ask for `false` and have been getting the opposite**: `site-scraper/scrape_site`,
> `site-adoption-agent/fetch_primary_css`, `website-capture-firecrawl/scrape_main_page`.
> So the probe's 22%→30% was never reachable by adding page fetches: the probe fetched
> **raw HTML**, production fetched **main-content-only**, and company numbers live in
> footers. `[UNVERIFIED]` how much of the gap this accounts for — not measured.
>
> **Candidate 2 (implement the keys):** `fallback_url_field` and `extract_mode` are
> implemented. `max_pages`/`follow_links` map onto the dialect the adapter already
> supports (`Crawl` reads `limit` and `include_paths`) rather than a second one being
> invented — and because `/scrape` fetches exactly one page, a single-page step
> carrying them now **warns** instead of silently fetching one page and reporting
> success.
>
> **Candidate 1 (the class fix):** `ActionInputSpec` gains `ConfigKeys`, and the
> workflow validator reports step-config keys the action does not read. **Opt-in per
> action** — measured 811 distinct (action,key) pairs over 228 live actions, so a
> fleet-wide allow-list would be a guess at scale and an over-strict validator is worse
> than the inert key it chases. `StrictConfig` turns the warning into a refusal once a
> contract is known complete; `scrape_web` is **not strict yet**, because both live
> definitions would have to be corrected first and `domain-research-classifier` still
> has no owner.
>
> **A FIFTH inert key, found by the new audit on its first run, not by reading this
> file:** `add_protocol`, on `domain-research-classifier/scrape_site`. It is a near-miss
> typo — the code reads `add_protocol_if_missing`, and that line belongs to a
> **different action**. A bare domain reaching the adapter is a failed fetch, so it was
> not cosmetic. Implemented. This is the argument for the detector in one line: four
> keys were found by a careful reader, the fifth needed a machine.
>
> **Candidate 3 (delete the keys) was NOT done, and is now unnecessary** — the keys are
> either implemented or reported, so nothing false is left in place to delete. The
> "delete by (action, key), never by key" trap recorded below is unchanged and still
> applies to anyone doing fleet-wide config edits.
>
> **Still open until:** the chassis image rolls (and the web-scrape-adapter image for
> the `only_main_content` half). New: `./scripts/audit-config-keys.sh`.

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

## Post-roll triage 2026-07-27 (~15:55 UTC) — unchanged; both definitions still carry the inert keys

Fleet rolled to **v1.0.1174** (`2026-07-27T15:11:15Z`). No fix written, so nothing
moved. The blast-radius query above re-run live returns the identical two rows:

| type | step | max_pages | follow_links | extract_mode | fallback_url |
|---|---|---|---|---|---|
| `vet-practice-verifier` | `scrape_website` | t | t | t | t |
| `domain-research-classifier` | `scrape_site` | t | t | t | f |

`domain-research-classifier` is still classifying domains on home-page text alone
while its config says otherwise, and still has no identified owner.

**Why this one is worth more attention than its severity line suggests.** It has
already produced one false claim that reached a commit message within an hour
(`WRONG_CALLS.md` 2026-07-26), and the failure mode is not "a feature is missing" —
it is that **a dead config key is indistinguishable by inspection from a live one**,
so the config *reads as evidence* while being evidence of nothing. That is a
misleading-power problem, and misleading-power compounds: every future session that
reads this step's config gets the same wrong impression, and the cheap check
(`grep -rn "<key>" --include=*.go .`) is one nobody thinks to run precisely because
the key is sitting there looking implemented.

**Candidate 3 is a genuine quick win and is being under-sold by its own entry.**
Deleting the four inert keys from both definitions is config-only, live immediately,
under an hour, and needs no image window. The file rightly notes it "removes the
evidence of the problem rather than the problem" — but the *misleading* half is the
half that has actually cost something so far, and candidate 2 (implementing the keys)
is gated behind a Go window that `bugs_open/100` shows is not imminent. Deleting them
now and implementing later is strictly better than leaving a false specification in
place for however long that window takes. **Caveat:** do not delete silently — leave
the intent in the bug file and a comment in the seed, or the next author re-adds them.

**Do not skip the `[UNSETTLED]` box above before implementing candidate 2.** The
Firecrawl `onlyMainContent` question is unresolved and could make the whole
22%→30% gain evaporate; settling it costs one real verification run and must come
first.

**Bundling still holds:** same step, same roll as `bugs_open/100`, and together they
gate `vetcomparison` P1 — see the post-roll block in `100` for the sequencing
argument.

## Related

- `bugs_open/100` — the provenance defect on the same step; fix in the same council round.
- `WRONG_CALLS.md` 2026-07-26 — "I recommended a config change without checking the config key is
  read", with the one-command check.
- Same family as the fleet's inert-layer entries: **the artefact exists, so the capability is
  assumed.**

---

## ⚠️ 2026-07-27 — CANDIDATE 3 HAS A TRAP: the keys are inert PER ACTION, not per name

Contributed by the bug-sweep thread while re-grounding this case. **Nothing above
is wrong** — but the diagnostic greps in "Root cause" are filtered in a way that
makes all four keys look universally dead, and one of them is not.

The file's own check for `max_pages` is:

```
$ grep -rn "max_pages" --include=*.go . | grep -i webscrape      # no hits
```

That is a true statement about `webscrape`. Re-run it **unfiltered** and the
picture changes:

```
$ grep -rn --include=*.go "max_pages" . | grep -v '^./docs/'
platform/orchestration/actions/select_representative_content_action.go:35   Optional: []string{"max_pages", ...}
platform/orchestration/actions/select_representative_content_action.go:54   if mp, ok := config["max_pages"].(float64); ok {
platform/orchestration/actions/v3_site_actions.go:2963                      if mp, ok := config["max_pages"].(float64); ok {
platform/orchestration/actions/v3_site_actions.go:5323                      logger.Warn("validate: preserved pages exceed max_pages; ...")
```

**`max_pages` is a live, load-bearing key — for two OTHER actions.** And those
actions are in use right now:

| action | type | step | `max_pages` | live? |
|---|---|---|---|---|
| `scrape_web` | `domain-research-classifier` | `scrape_site` | 3 | **inert** (this bug) |
| `scrape_web` | `vet-practice-verifier` | `scrape_website` | 3 | **inert** (this bug) |
| `select_representative_content` | `site-adoption-agent` | `select_content` | 3 | **LIVE** — read at `select_representative_content_action.go:54` |
| `validate_site_plan` | `build-site-planner` | `validate_plan` | 80 | **LIVE** — read at `v3_site_actions.go:2963` |
| `validate_site_plan` | `site-planner` | `validate_plan` | 20 | **LIVE** — same |

**So candidate 3 must delete by (action, key), never by key.** A fleet-wide
cleanup written the obvious way —

```sql
-- DO NOT DO THIS
UPDATE agent_definitions SET default_config = ... WHERE config ? 'max_pages'
```

— strips a live page cap from three steps, including `build-site-planner`'s 80,
where the code's own warning shows the value doing real work: *"validate:
preserved pages exceed max_pages; keeping all preserved, dropping all net-new"*.
Silently uncapping a site planner is a considerably worse bug than the one this
file is about.

The other three keys **are** dead everywhere, checked unfiltered across the whole
tree (`docs/` excluded, which holds vendored copies):

```
$ grep -rn --include=*.go "follow_links\|extract_mode\|fallback_url_field" . | grep -v '^./docs/'
(no output)
```

This is the same shape as the recorded landmine *"a narrow filter defines the
conclusion"*: a grep filtered by a term taken from the question answers
confidently about a small world and never reveals that the world was small. The
filter was `webscrape`, and it was correct for the claim being made — the risk
only appears when the next reader carries the conclusion **"these four keys are
inert"** into a fleet-wide edit, which is exactly what candidate 3 is.

**Unchanged by this note:** candidate 3 is still not recommended over 1+2, for the
reason the file already gives — it removes the evidence rather than the problem.
And `domain-research-classifier` still has no identified owner, so half of
candidate 3's blast radius is someone else's agent.
