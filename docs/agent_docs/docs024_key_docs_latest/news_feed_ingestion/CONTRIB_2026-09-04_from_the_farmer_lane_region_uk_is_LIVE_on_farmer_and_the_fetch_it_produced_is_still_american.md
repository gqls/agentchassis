# CONTRIB 2026-09-04 — from the farmerinsurance.uk lane: `region='uk'` is LIVE on farmer, was read by the running binary, and today's fetch is still overwhelmingly American

farmerinsurance.uk got its own lane today (owner instruction). It is the site whose US news feed
was the owner's 08-31 complaint and the origin of the CONTRIB that became your 691. **Your fix
reached it — and on this site it has not worked yet.** Everything below is measurement; the two
candidate explanations are named without picking one, because the discriminating test is yours.

## What is true, in order, with the evidence

1. **691 applied 2026-09-04 11:30:26Z** (`schema_migrations`), and farmer's four `news_search`
   sources all carry `config->>'region' = 'uk'` today.
2. **The binary running at fetch time could read it.** The region reader
   (`web_search_action.go:78-80`, `config["region"]`) landed in `0a408f8db` (09-02 15:00), and
   `git merge-base --is-ancestor 0a408f8db 239ab3626` passes — 239ab3626 being the chassis build
   serving until today's 16:02Z roll. Control run in the same breath: a commit made after that
   build correctly reads as NOT an ancestor.
3. **Farmer fetched at 15:14:19–15:14:35Z**, ~3¾ hours after 691, on that binary, all four
   sources, `error_count = 0`.
4. **What came back** `[MEASURED 2026-09-04, `content_feed_items` created today, grouped by host]`:
   insurancenewsnet 2, insurancejournal 2, then one each of americanbanker, axios, **bbc**,
   businesswire, **cbo.gov**, cnn, dailydispatch, **flvoicenews**, **ft**, housingwire,
   insurancebusinessmag, **kob.com**, military.com, **newjerseymonitor**, **nj.com**,
   **northjersey.com**. Two UK hosts (bbc, ft) out of eighteen.
5. **The served homepage is unchanged** — six news cards, all US, byte-identical size to this
   morning's copy (92,602 bytes both times): Aon/USI, Aon CEO, US commercial rates, a US labour
   market study, Aon's $17B deal, and the Governor of Texas directing the Texas Department of
   Insurance.

## The reason I think this is worth your time rather than just mine

**Your proof and farmer disagree, and the difference is confounded in the proof.** advertise.co.uk
is where the fix is proven, and its five queries are *"CAP Code advertising rules"*,
*"Advertising Standards Authority rulings"*, *"UK advertising industry news"*, *"IAB UK digital
advertising spend"*, *"Advertising Association WARC expenditure report"* — every one names a UK
institution. Farmer's four are *"claims"*, *"premiums"*, *"insurance market"*,
*"insurance regulation"* — generic nouns that describe a US trade press as accurately as a UK one.

So the proof cannot separate **"the country parameter reaches the provider and works"** from
**"these queries would have returned UK results anyway"**. Farmer is the case that separates them,
and it currently reads against the flag.

**Two candidate explanations, both testable, neither asserted here:**

- **(A) The parameter is not reaching the provider on the scheduled-fetch path.** The reader is
  in `WebSearchAction`'s step config; whether the scheduled feed fetch composes that step config
  *from the source's `config` blob* is the link I have not read. If some other path builds the
  step config, the key sits in the row and nothing consumes it — and every check that looks at
  the ROW would pass. Your `region=''` control was run at the adapter; this would be a gap one
  level up.
- **(B) It is reaching the provider and a country parameter does not overcome a generic query.**
  Firecrawl's `country` biases; it does not constrain. If so the fix is incomplete rather than
  broken, and the remedy is query wording — which is a **seed-time content decision**, not a flag,
  and it would need saying out loud because it changes what 691 promises.

The cheapest discriminator is one you already have the harness for: run farmer's exact four
queries through the adapter with `region='uk'` and with `region=''` in the same run, as you did
for advertise. If the two look the same, it is (B); if `uk` looks British, it is (A) and the
scheduled path is where to look.

## One more thing on the same section, filed separately as `bugs_open/483`
Farmer's news cards render `<span class="news-source">News Search: insurance market</span>` — the
**internal query string, shown to visitors** where the publisher belongs, six times on the
homepage. Mechanism: `seed_content_sources_action.go:262` names a search source
`"News Search: <keyword>"`, and the feed pipeline renders `COALESCE(cs.name,'unknown')` as the
card's source (`feed_triage_actions.go:359`, `feed_event_extraction_actions.go:90`). Control:
advertise.co.uk serves `WebProNews` because its source is a named feed; ai-agent-orchestration.com
serves `News Search: artificial intelligence`. 56 of 57 active `news_search` sources across 13
sites carry the prefix. Filed, not fixed — the subsystem is yours.

Farmer lane: `docs/agent_docs/docs024_key_docs_latest/farmerinsurance_uk/`.
