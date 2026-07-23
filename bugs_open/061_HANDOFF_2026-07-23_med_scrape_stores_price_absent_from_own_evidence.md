# 061 — med scrape can store a price that does not appear in its own retained evidence

**Filed 2026-07-23** while verifying the med-pipeline revival (vetcomparison workstream,
session "bugfix 054"). **Status: OPEN — mechanism not yet diagnosed.** Small blast radius
today (2 of 10 fresh rows), but the defect class is exactly what the provenance policy
exists to prevent: a published figure whose retained evidence does not contain it.

## One-line

`MedScrapePricesAction` stored two Advocate prices (£17.95, £29.95) whose own evidence
markdown — captured in the same scrape — contains only £17.**75** / £29.**75** (+£0.20
each). The export's provenance gate (v1.0.1151) checks provenance *presence* (url +
capture date), not parse *fidelity*, so both rows published to
`vetcomparison.uk/data/medicine-prices.json` (commit `a52fbf0`).

## Evidence (live DB, 2026-07-23)

- Run: orchestration `8e2eaa07` (worker) / `5717ab5d` (parent), both COMPLETED; 10 fresh
  `med_price_snapshots`, 19 `med_scrape_evidence` rows.
- The two rows: listing `retailer_product_name='Advocate'`
  (`https://www.petdrugsonline.co.uk/advocate`), snapshots
  `(Large Cats 80mg/8mg (4kg-8kg), 17.95)` and `(Small Dogs 10kg Pack of 3, 29.95)`,
  `collection_method='scrape'`, `raw_data={"markdown_length": 9293}`.
- Their evidence row (same listing, same run): `variants_found=2, prices_stored=2`,
  `length(markdown_content)=9256`. Neither `17.95` nor `29.95` appears in it; £-fragments
  present include `£17.75`, `£29.75Save`, `£27.69`, `£25.61`.
- The other 8 fresh rows all verify: each price appears verbatim in its own evidence
  markdown (query in RUNBOOK §Med retailer pipeline; per-row LIKE against
  `to_char(price,'FM999999D00')`).

## Hypotheses — UNVERIFIED, in rough likelihood order

1. **[INFERRED] Was-price / RRP capture:** the page shows sale prices (`£29.75Save…`);
   the parser may have matched a struck-through "was £29.95" that the markdown
   conversion dropped or reformatted (the +£0.20 symmetry on both variants suggests a
   systematic was-vs-now offset, not noise).
2. **[UNVERIFIED] Markdown divergence:** evidence stores 9,256 chars vs the parser's
   `markdown_length` 9,293 — 37 chars differ. If evidence is cleaned/truncated after
   parsing, the parser may have seen figures the evidence no longer holds. (If so, the
   evidence write is the defect — evidence must be byte-what-the-parser-saw.)
3. **[UNVERIFIED] Parser strategy cross-match:** `parseMedPriceVariants` has multiple
   retailer strategies; a variant-block regex may pair a size label with a price from an
   adjacent block.

## Fix candidates

1. **Parse-fidelity guard at the scrape write** (preferred; same fail-closed shape as the
   export gate): before INSERTing a snapshot, require the price literal to appear in the
   evidence markdown being stored for that listing; on miss, skip + count + Warn (never
   silent). Cheap: one `strings.Contains` per variant.
2. Evidence byte-fidelity: persist exactly the markdown the parser consumed (fixes
   hypothesis 2 whichever way it resolves).
3. Diagnose first via the 090 loop if the mechanism resists a quick read of
   `parseMedPriceVariants` (`vet_med_price_scrape_action.go:497+`, pdo strategy ~:460).

## How to verify a fix

Re-scrape the Advocate listing (`UPDATE ... last_triggered_at=NULL` on med-scrape-prices;
it is stalest-first, or temporarily deactivate other listings); assert every stored
snapshot's price appears in its own evidence markdown; induce the failing branch with a
synthetic markdown lacking the price (the guard must skip + count, not store).

## Related

- vetcomparison RUNBOOK §Med retailer pipeline (verification queries; leak-path notes).
- `bugs_closed/054` (the envelope defect whose revival work surfaced this).
- Same LISTING-quality family, different defect: legacy category-page listings
  ("Cat Tick", "Horse" — created 2026-04-02 by url_discovery) publish
  cheapest-in-category prices under non-product names. Provenanced but semantically weak;
  hygiene pass = owner call (deactivate + path-filter in discovery; note the export does
  NOT filter on `med_retailer_listings.is_active`, only `med_retailers.is_active`, so
  deactivating a listing alone does not stop its recent snapshots exporting for 14 days).
