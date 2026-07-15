# Factual record — vetcomparison.uk data review and remediation

**Prepared:** 15 July 2026
**Status:** Contemporaneous internal factual record. This is not legal advice and makes no
admissions or legal conclusions; it records what was found, when, and what was done. It is
written so that a solicitor can be briefed from it quickly. Facts below are stated with their
evidence source. Where something is unverified it is marked as such.

---

## 1. Background

`vetcomparison.uk` is a static website operated by us, served from the `gqls/sites` GitHub
repository via an automated deploy to object storage behind Cloudflare. Its purpose is to help
UK pet owners compare veterinary prices, anticipating the price-transparency remedies arising
from the Competition and Markets Authority (CMA) market investigation into veterinary services
for household pets.

A predecessor prototype had earlier been built against `vetcomparison.co.uk`, a domain we do
not own. The present site was added to the repository and published on **2 February 2026**
(repo commit `1ec1eda8`, "add vetcomparison.uk").

## 2. What was published (2 February 2026 – date of remediation)

The published site contained the following. All of it originated from a prototype dataset; an
internal review (below) concluded the figures could not be traced to any source.

1. **A price directory of 3,124 named, real UK veterinary practices** (`/data/vet-full-index.json`
   and per-postcode/per-name variants), each entry carrying a specific consultation fee
   (`consult_price`) and prescription fee (`rx_price`), and in the per-postcode files a
   `services` block with further specific prices (e.g. "Booster Vac (Dog)", "Urgent
   Consultation"). Practice names, postcodes and websites referred to identifiable real
   businesses.
2. **A medicine price comparison tool** covering 155 medicine entries with specific prices
   attributed to four named online pharmacies (Pet Drugs Online, Animed Direct, VioVet,
   Hyperdrug) and an invented "typical vet price" per medicine (`vet_price_est`).
3. **A data-rights notice on every record**: `"© VetComparison.uk - Proprietary Data. Do not
   scrape."`
4. **Three guide pages**, including one containing a quotation presented as from the CMA's
   provisional findings ("We anticipate a structural shift…") and one asserting that a named
   practice ("Taylor Vets, Glasgow") charged £33 for writing a prescription. Guides also stated
   a £16 prescription cap (the final figure is £21) and other figures presented as CMA
   findings without citation.

## 3. Discovery

On **14 July 2026**, during a planned rebuild of the site, an internal data review found that
the published price data was not genuine. Evidence, all reproducible from the repository and
production database:

- **Statistical fingerprint inconsistent with real-world data.** Across 3,124 practices there
  were only 46 distinct consultation fees; 692 practices (22.2%) carried the identical fee of
  £48; all prices were integers. Genuine scraped price data held elsewhere in our systems does
  not exhibit this pattern.
- **No provenance.** No entry in the published dataset carried a source URL, capture date, or
  evidence reference. By contrast, our genuine price-collection pipeline stores a source URL,
  content hash and screenshot for every observation; 100% of its rows carry a source URL.
- **Internal inconsistency with our verified data.** The published medicine data attributed
  prices to "animeddirect.co.uk"; our verified retailer records show the correct domain is
  `animed.co.uk` (a correction recorded in our retailer seed data). The published dataset
  therefore did not come from our working collection pipeline.
- **Lineage.** A generated SQL seed file (`007_seed_vets.sql`, "Seed data import for vet
  practices from E.json and C.json — Generated automatically") shows the dataset derives from
  prototype JSON files, i.e. from content generation, not from collection.

Conclusion recorded at the time: the published prices were fabricated (machine-generated
placeholder data), attributed to real businesses, and published with a proprietary-data claim
that was itself unfounded.

## 4. Related contamination of the internal database

The same prototype dataset had been partially imported into the production database
(`business_intel.business_prices`), where **997 price rows tagged `source='seed_import'`** were
attached to 251 practice records, 235 of which are real practices verified by our collection
pipeline. These rows carried no source URL. They had not been re-published from the database
(the relevant export task `med-export-json` was and is disabled; a practice-price exporter was
never built), but they were live in the sense of being flagged `is_current=true`.

Sixteen practice records whose existence was never independently verified are tagged
`verification_status='seed_import'` and are excluded from all publication.

## 5. Remediation

| Date | Action | Evidence |
|---|---|---|
| 14 Jul 2026 | Fabrication identified and quantified; decision taken to remove all published prices immediately and rebuild on sourced data only | This record; session analysis |
| 15 Jul 2026 | Site stripped: all price data files deleted; medicine comparison tool deleted; the three guide pages removed; the "proprietary data / do not scrape" notice removed; directory reduced to factual fields only (name, location, postcode, website) | Commit `fbc0b929` on branch `strip-vetcomparison`, repo `gqls/sites` |
| 15 Jul 2026 | Directory regenerated from our verified dataset: 2,579 practices confirmed by our verification pipeline (each with website and postcode), replacing the prototype list of uncertain provenance | Same commit; export query against `business_intel.businesses` (`verification_status='verified'`) |
| 15 Jul 2026 | Database quarantine: all 997 fabricated price rows set `is_current=false` (retained, not deleted, for audit); verified afterwards — 0 fabricated rows remain current; the 803 genuine, source-attributed rows are unaffected | Production `UPDATE`, verified counts 997→0 |
| 15 Jul 2026 | Confirmed no automated process can republish price data: all med/vet export and collection scheduled tasks are disabled; no site-builder record exists for the domain | `scheduled_tasks` query; `sites` table query |
| 15 Jul 2026 | **Deployed to the live domain.** The strip was pushed to `master` (commit `92526ccd`) and the deploy workflow published it. Verified live: the public directory JSON now serves only factual fields with no prices; the removed medicine calculator, medicine data and guide pages return HTTP 404. The fabricated dataset is no longer publicly accessible. | Repo `gqls/sites` commit `92526ccd`; live checks of `vetcomparison.uk` (directory JSON, `/assets/js/calc.js`, `/data/medicine-index.json`, `/guides/cma-compliance.html`) |
| 15 Jul 2026 | **Guide pages republished with sourced content** at the same three URLs, replacing the removed versions. Rewritten against the CMA final report (24 Mar 2026) and CMA guidance, both verified same day; each page carries a "last reviewed" date, a status banner noting the Order is not yet made, and a sources list. Audited before publication against the removed content: none of the earlier unsourced figures, the fabricated quotation, or the named-practice claim recur; the only monetary figures published are the CMA's own (£21 / £12.50 prescription cap, £500 estimate threshold). Verified live. | Repo `gqls/sites` commit `f18eb395`; live checks of all three `/guides/` URLs (HTTP 200, new content, old figures absent) |

The replacement page states plainly that prices were removed because they could not be traced
to a source, and that price comparison will return only on the basis of sourced, dated,
attributed data.

## 6. Regulatory context (grounded 15 July 2026)

Verified directly against the CMA case page and CMA guidance on 15 July 2026:

- The CMA published its **final report** in the veterinary services market investigation on
  **24 March 2026**. ([case page](https://www.gov.uk/cma-cases/veterinary-services-market-for-pets-review))
- **No remedies Order has yet been made.** The statutory deadline for the Order is
  **23 September 2026**. A consultation on the draft **Funding** Order (RCVS levy) opened
  30 June 2026 and closes **30 July 2026**; the consultation on the **substantive** remedies
  Order was scheduled for July 2026 publication and had not appeared as at 15 July 2026.
- Once the Order is made, veterinary businesses must publish **standard price lists for a
  defined set of services** (a fixed 36-item list in 5 categories; no free text; VAT-inclusive;
  six standardised pet categories), online at most one click from the homepage — large groups
  within 3 months (~December 2026), smaller businesses within 6 months (~March 2027) — plus a
  parasiticides price list, ownership disclosure (6 months, all businesses), written estimates
  for treatment ≥£500, and a written prescription fee cap of **£21** (£12.50 per additional
  medicine). ([CMA guidance](https://www.gov.uk/guidance/what-veterinary-businesses-and-vets-need-to-do-following-the-cmas-final-vets-report))
- Practices must submit key information to the **RCVS "Find a Vet"** platform (within 12
  months, ~September 2027), and the RCVS will share that data with **approved third parties,
  "who might use it to offer a comparison site (subject to certain restrictions)"** (CMA
  guidance, quoted). The final report treats third-party comparison services, including those
  gathering data by web scraping, as legitimate, subject to the general-law requirement that
  information not be presented in a misleading or unfair manner (Final report Part B,
  ¶3.320–3.321), and names an existing comparison operator (VetHelpDirect, ¶3.205).
- These remedies bind veterinary businesses and the RCVS. They do not impose obligations on
  comparison websites beyond the general consumer-law standard above; the "no paid or promoted
  rankings" restriction (Part B ¶3.209(a)) is a condition of receiving the RCVS data feed as
  an approved third party.

## 7. Publication policy adopted (from 14 July 2026)

1. **No price is published unless it can be traced to a source**: a source URL and capture
   date, with underlying evidence (page content hash, screenshot) retained internally.
2. **Per-practice prices are published with attribution and a dated source link**, or with the
   practice's own consent through a claimed listing; otherwise only **aggregate statistics**
   (e.g. area medians) are published.
3. **No data-rights claims are made over third-party facts.** The earlier proprietary-data
   notice has been removed and will not be reinstated in that form.
4. **Regulatory figures are stated only from primary sources** (CMA publications), with the
   distinction between proposed and final requirements preserved.
5. Practices can request correction or removal via a published contact route; the earlier
   correction mechanism is retained on claimed listings.

## 8. Open items

- Update this record if the substantive draft Order consultation opens (expected July 2026);
  the final Order may adjust the 36-item list, weight bands and caveat wording.
- Data-quality (not a liability item): a small number of verified practice records hold a
  scraped page title in the name field rather than a clean practice name (e.g. an entry reading
  "26 Vets in Birmingham - Compare Prices & …"). These are real records with a cosmetic
  capture flaw, to be cleaned in the rebuild; they carry no fabricated prices.
- Obtain a solicitor's review of (a) this record, (b) the database-right position on
  republishing practice price lists at scale (not addressed by the CMA report; genuinely
  untested), before per-practice price publication resumes at scale.
- Unresolved detail for schema work: Part B ¶3.88 describes six pet categories while the CMA
  guidance table merges "cat, small dog: <10kg" into one band — resolve against the final
  Order before freezing the comparison schema.

## Sources

- CMA case page: https://www.gov.uk/cma-cases/veterinary-services-market-for-pets-review (accessed 15 Jul 2026)
- CMA final decision report, Part B (24 Mar 2026): https://assets.publishing.service.gov.uk/media/69c2ad8d13f1436476e44382/Final_decision_report_-_Part_B_24.3.26.pdf
- CMA guidance for veterinary businesses: https://www.gov.uk/guidance/what-veterinary-businesses-and-vets-need-to-do-following-the-cmas-final-vets-report (accessed 15 Jul 2026)
- Repository evidence: `gqls/sites` commits `1ec1eda8` (2 Feb 2026, publication) and `fbc0b929` (15 Jul 2026, remediation)
- Database evidence: `business_intel.business_prices` (`source='seed_import'` rows, quarantined 15 Jul 2026); `business_intel.businesses` verification statuses
