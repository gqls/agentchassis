# 061 — med scrape can store a price that does not appear in its own retained evidence

**Filed 2026-07-23** while verifying the med-pipeline revival (vetcomparison workstream,
session "bugfix 054"). **Status: OPEN — mechanism CONFIRMED 2026-07-24 (session "bugfix 061
med scrape"), fix BUILT + tested, data remediated; stays open until the fix is live in a
rolled image.** The defect class is exactly what the provenance policy exists to prevent:
a published figure whose retained evidence does not contain it.

> **CORRECTED 2026-07-24 — blast radius was 212 rows, not 2.** The "2 of 10 fresh rows"
> below was true of that run only. A same-run-evidence fidelity sweep over the whole table
> found **8 fabricated rows in the live 14-day export window** (Advocate ×2 — the filed
> pair, live on the site — plus Vetoryl ×3, Corvental-D ×1, Doxybactin ×2) and **204 more
> from the pipeline's April era** (2026-04-07→11), 79 of them at exactly **£17.48 — the
> LLM prompt's worked-example price, echoed back**. All 212 quarantined and deleted; see
> §Remediation.

## Mechanism — CONFIRMED 2026-07-24 (direct evidence, no inference)

**The LLM fallback fabricates price tables, and they are stored blind as
`collection_method='scrape'`.** When the regex parser finds 0 variants,
`MedScrapePricesAction` hands a **1,500-char truncated** window (`md[:1500]`) of the
product section to a local Mistral (`llmExtractPriceVariants`). The fallback gate checked
the FULL product section for a `£`, not the window the LLM would actually see — so on the
Advocate **category page** (regex rightly found nothing) the gate passed on a *delivery
banner's* £49, while the window itself contained **no product prices at all**. Shown
priceless text and asked to extract prices, the model invented complete variants — sizes,
prices, TVPs — and `storeMedPriceSnapshots` stored them with no check against the evidence.

Smoking guns in `llm_call_log` (provider='ollama', prompts + responses retained;
timestamps match the snapshots to the second):

| call id | when | returned |
|---|---|---|
| `449ba06e` | 07-23 12:25:49 | Advocate: `17.95` / `29.95` + invented TVPs 23.60/48.90 and a size ("Small Dogs 10kg Pack of 3") that appears **nowhere in its prompt** |
| `b3f17928` | 07-23 18:00:21 | Vetoryl: all three strengths invented (59.87 / 120.99 / 204.87 + TVPs) |
| `d7e6df03` | 07-24 06:14:35 | Corvental-D: **echoed the prompt's worked example verbatim** (`100ml, 17.48, tvp 0`) |
| `f224d42c` | 07-24 12:03:14 | Doxybactin: both pack sizes invented (9.48 / 23.68) |

The prompt for `449ba06e` (retained in `prompt_rendered`) ends mid-token at "[Advocate Sp" —
the 1,500-char cut — confirming the LLM never saw £17.75/£29.75 or any product price.

## Hypotheses — RESOLVED 2026-07-24

1. ~~**[INFERRED] Was-price / RRP capture**~~ → **REFUTED.** The LLM never saw the real
   prices at all; the +£0.20 symmetry was coincidence (hallucination from model priors).
   What caught it: reading the retained `prompt_rendered` for call `449ba06e`.
2. ~~**[UNVERIFIED] Markdown divergence** (9,293 vs 9,256)~~ → **REFUTED — measurement
   artifact.** Go `len()` counts bytes, Postgres `length()` counts characters; `£` is two
   bytes in UTF-8. `octet_length(markdown_content) = 9293` exactly. **Evidence IS
   byte-what-the-parser-saw.**
3. ~~**[UNVERIFIED] Parser strategy cross-match**~~ → not the mechanism here (regex found
   nothing, correctly — category page). The shipped guard covers this class anyway.

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

## Fix — BUILT 2026-07-24, unit-tested, awaiting image roll

All in `platform/orchestration/actions/vet_med_price_scrape_action.go` (+ new
`vet_med_price_scrape_action_test.go`); fix candidate 1 from the original filing, plus
provenance and gate hardening the diagnosis exposed:

1. **Parse-fidelity guard** (`medFilterVariantsByEvidence` / `medPriceLiteralInMarkdown`):
   a variant is stored only if its price literal appears in the markdown retained as
   evidence — comma-insensitive, non-digit neighbours (so £117.95 can't vouch for 17.95);
   whole-pound/short renderings only count £-prefixed (so "17ml" can't vouch for £17.00).
   Price miss → drop + Warn + `fidelity_skipped` count in the action result. TVP miss →
   TVP zeroed, variant kept. Applies to both regex and LLM paths (regex passes by
   construction). Evidence rows now record pre-guard `variants_found` vs `prices_stored`,
   so a guard drop is visible in the DB row itself.
2. **Honest provenance**: LLM-extracted rows now store `collection_method='scrape_llm'`
   (was: indistinguishable `'scrape'`). Checked before changing: nothing filters on the
   column (free text; MV and export ignore it).
3. **Gate on the window the LLM actually sees**: the fallback now fires only if the
   truncated window contains `£` (was: the full product section — the Advocate trigger).
4. **Prompt hardening**: "only extract prices that appear verbatim … never estimate,
   recall, or compute"; the worked example is labelled "format example only — never copy
   its values" (the Corvental-D echo proves the example gets copied).

Unit tests (17 cases) include the induced fault — the exact Advocate regression shape
(markdown holds £17.75/£29.75Save, variants claim 17.95/29.95 → both dropped) — and the
whole-pound/embedded-digit/comma edge cases. Green against `git archive HEAD` + changed
files overlaid.

Code committed `ca2cd7535` (2026-07-24). Council gate submission:
`SUBMISSION_CORR=7cf73cc1-1f14-4ca0-9feb-2a90fa4bfd83` (run orch `a8e0c003`), verdict
pending at time of writing — check:
`SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
correlation_id='7cf73cc1-1f14-4ca0-9feb-2a90fa4bfd83' AND kind='council_report' ORDER BY created_at;`

## Remediation — DONE 2026-07-24 (before the next export, which was ~07-25 12:31)

All 212 fabricated rows copied to **`business_intel.med_price_snapshots_quarantine_061`**
(full rows + `quarantined_at` + `reason`; reversible), then deleted; MV refreshed:

- Window 8 (llm_call_log-confirmed): snapshot ids 3989, 3990 (Advocate 17.95/29.95),
  3995–3997 (Vetoryl 59.87/120.99/204.87), 4020 (Corvental-D 17.48), 4024, 4025
  (Doxybactin 9.48/23.68).
- April-era 204 (2026-04-07→11; price absent from same-run evidence, gap ≤120s): includes
  the £17.48 example-echo epidemic (79 rows; NOTE 19 *other* 17.48 rows are genuine —
  17.48 appears in their evidence — so never purge by price value, only by evidence check).
- 33 listings' `med_retailer_listings.last_price` carried fabricated values (30 × 17.48);
  reset to NULL where no surviving snapshot backs the value (nothing in Go reads
  last_price today — write-only column — but it was wrong data).
- Verified after: full-table fidelity sweep returns **0 PRICE_ABSENT**; MV contains none
  of the 8 window values. The live medicine-prices.json still carries the two Advocate
  fabrications until the next `med-export-json` run (48h cadence) rebuilds it clean.

Sweep query: vetcomparison `RUNBOOK_vetcomparison.md` §Med retailer pipeline
(fidelity sweep). **Interim until the image rolls**: `med-scrape-prices` runs 6-hourly and
can still fabricate — re-run the sweep and quarantine-delete any new PRICE_ABSENT rows.

## How to verify the fix live (post-roll)

1. Pod-grep a string the change CREATED, not one it uses: `fidelity guard dropped variant`
   + a positive control.
2. Re-scrape the Advocate listing (stalest-first; `UPDATE ... last_scraped_at=NULL` on the
   listing) and assert: any stored snapshot's price appears in its own evidence markdown;
   the category-page case yields skip + Warn + `fidelity_skipped` > 0, not stored rows.
3. Confirm new LLM-path rows (if any legitimate ones occur) carry
   `collection_method='scrape_llm'`.

## Related

- vetcomparison RUNBOOK §Med retailer pipeline (verification queries; leak-path notes).
- `bugs_closed/054` (the envelope defect whose revival work surfaced this).
- Same LISTING-quality family, different defect: legacy category-page listings
  ("Cat Tick", "Horse" — created 2026-04-02 by url_discovery) publish
  cheapest-in-category prices under non-product names. Provenanced but semantically weak;
  hygiene pass = owner call (deactivate + path-filter in discovery; note the export does
  NOT filter on `med_retailer_listings.is_active`, only `med_retailers.is_active`, so
  deactivating a listing alone does not stop its recent snapshots exporting for 14 days).
