# CONTRIB from the loanzy lane, 2026-08-31 — OWNER ASK: UK news default for .uk/.co.uk (a region flag). The capability is ABSENT, not undriven.

The owner, reviewing farmerinsurance.uk: *"The news is from America. I'd like it to be UK
news for all .co.uk and .uk sites, perhaps as a flag with a UK default."*

Routed here after the news_editorial_features lane correctly declined it (their lane is
editorial FEATURE PAGES, not feed ingestion — "the news FEED subsystem has its own lanes",
naming this one first). **Their verification, so you start from measurement (all
[MEASURED 2026-08-31] by news_editorial, quoted with attribution):**

1. Farmer's four `content_sources` carry bare configs — `{"query": "claims",
   "num_results": 10}` etc. Nothing in the row scopes region.
2. **Fleet-wide, the capability does not exist**: across all 48 `news_search` configs,
   `jsonb_object_keys` yields num_results(48), query(48), max_items(20), feed_url(10),
   model(9), hours_lookback(9), provider(9), search_tools(9), prompt_template(9),
   keywords(8), time_range(5), url(1), scrape_config(1) — **zero region/country/locale/
   gl/hl keys anywhere**. Not "a flag nobody set"; it must be built end to end.
3. **The seam is `web_search_action.go`** — it already reads exactly this class of
   per-source parameter from the config blob (num_results :58, time_range :68, documented
   :38-40). A region key belongs beside them; time_range is the precedent that the blob is
   where per-source search parameters live.
4. **The TLD default belongs at SEED time** (`seed_content_sources_action.go`) so the
   default is visible in the row a human inspects. Open design question they flagged:
   check whether the estate already derives anything from TLD before inventing a second
   derivation.
5. **Scope caution (theirs, and right):** a TLD-keyed default changes live content for
   every existing .uk/.co.uk site with news_search sources — count the affected population
   (one query, content_sources join sites on domain suffix) BEFORE it ships, not after.

Two adjacent defects for the same fix window: the acceptance council's site-review AND
content-quality seats each filed (held, record-mode) that farmer's news section links to
**malformed Google redirect URLs** rather than real content — verdicts at `deferred` on
site `99cae989-2413-430d-b026-59dfeeb638c0` (queue query: loanzy_uk_example_site/RUNBOOK;
release is a human verb). And bugs_open/316's bootstrap catch-22 is the standing third leg.

The owner's word is the mandate for the FEATURE; the design is yours. Full review:
loanzy_uk_example_site/OWNER_REVIEW_2026-08-31_farmerinsurance_first_review_and_routing.md §3.
