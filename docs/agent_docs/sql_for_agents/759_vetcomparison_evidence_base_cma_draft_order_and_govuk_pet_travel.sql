-- 759_vetcomparison_evidence_base_cma_draft_order_and_govuk_pet_travel.sql
--
-- Gives vetcomparison.uk its evidence register: 21 facts (13 CITED with a URL and a
-- verbatim quote verified through the production matcher, 8 ATTESTED from a primary document
-- the matcher structurally cannot read), plus 6 banned-claim patterns and the `veterinary`
-- citation-code preset. OWNER RULING D1, 2026-09-03: "we'd want a register for vetcomparison".
--
-- The register is UNOWNED and this build was invited: the vetcomparison lane confirmed it never
-- started one and handed over its primary-source transcription
-- (docs024_key_docs_latest/vetcomparison/ATTESTATION_2026-09-03_cma_draft_order_and_govuk_pet_travel.md,
-- and NOTES_vetcomparison.md 2026-09-03). Method: lendzy_co_uk/RUNBOOK_lendzy_co_uk.md §8 + §8b/§8c/§8e/§8f.
--
-- ─── THREE LIVE ERRORS THIS PASS FOUND, recorded as FOUR `corrects_site_citation` records ────
--
-- (three distinct errors; the caps error is recorded on BOTH cap facts, so the verify block
-- below asserts 4 records, not 3 - the numbers differ on purpose and each is checked.)
--
-- Every lane that has run this method has found errors in its own site's live copy (lendzy 2,
-- loanzy 1, loancalculator 2, loancash 3). This is now five for five.
--
-- (1) **THE FINAL REPORT DATE IS WRONG BY ~16 MONTHS. NEW, found by this pass.**
--     /guides/cma-compliance/ and /guides/cma-market-investigation/ both say the CMA "published
--     its final report in November 2024". The CMA's own case-page timetable says "24 March 2026
--     Final report published", and its consultation page says "In March 2026 we published the
--     final report". November 2024 is when the Inquiry Chair spoke at the BVA Congress - also on
--     the case page, which is the likely origin of the slip.
--
-- (2) **THE £21 / £12.50 PRESCRIPTION FEE CAPS ARE SERVED AS SETTLED AND ARE PLACEHOLDERS.**
--     Known to the vetcomparison lane since 2026-08-24 and unfixed; confirmed here by reading the
--     primary document rather than inheriting the claim. The draft Order defines them in square
--     brackets: "'Initial Primary Prescription Fee Cap' means [£21 inclusive of VAT. This will be
--     adjusted for inflation between the date of the Final Decision Report and the latest monthly
--     CPI figure available before the Order is made]." Seven served pages state them as settled.
--
-- (3) **"36 SERVICE CATEGORIES" IS 36 SERVICES IN 5 CATEGORIES. NEW, found by this pass.**
--     Draft Schedule 1's own column heading is "Service, product, treatment or procedure (36
--     total)", grouped into five numbered categories (12+6+6+9+3 = 36, checked). The number is
--     right; the noun is not. Live on at least eight served pages.
--
-- **NO COPY IS TOUCHED HERE.** The 695/699/738 precedent and the owner's content freeze: a
-- register records what is true, it does not rewrite served pages. Repairs are the owner's call.
--
-- ─── WHY 8 FACTS CARRY NO CITATION, WHICH IS A DESIGN DECISION AND NOT A SHORTCUT ─────────
--
-- **The CMA's primary sources are PDFs, and the daily citation re-checker cannot read a PDF.**
-- Measured 2026-09-03 through the production matcher (`go run ./cmd/fcaquotecheck`) against
-- the draft substantive Order: HTTP 200, raw=392,144 bytes, and `VisibleTextFromHTML` yields
-- 296,699 characters of extracted noise - in which EVERY quote returns false, including
-- "Compliance Date", which is certainly in the document. The absent control returned false too,
-- so at a PDF the check discriminates NOTHING.
--
-- A `source.citation` on a PDF would therefore classify as `citation_lost` drift EVERY DAY, for
-- ever, and that false alarm is indistinguishable from a real one (RUNBOOK §8 step 4). So those
-- facts carry `source.attested_by` instead, which `refresh_evidence_base_action.go` handles with
-- a ~180-day staleness NUDGE and never fetches (verified at the code: the re-fetch arm is gated
-- on `if _, has := src["citation"]; has`, :576). Each such fact also carries `source_document`
-- (the PDF URL, for a human) and `no_citation_because` (the measurement above), so the absence is
-- legible rather than looking like an omission.
--
-- ⚠ **THIS IS A FOURTH SIGNATURE FOR THE UNREGISTRABLE-HOST CENSUS (RUNBOOK §8g).** The three
-- known ones are a Cloudflare challenge page, UA-differential serving, and founding-name slugs.
-- This one is different in kind: the HOST is fine, the DOCUMENT is right, the fetch returns 200
-- and a large "visible text" - it is the EXTRACTOR that cannot read the format. A host-acceptance
-- check would pass it. Only §8 step 4 - the probe THROUGH the production matcher - catches it.
--
-- ─── WHAT THIS REGISTER ARMS, MEASURED, INCLUDING WHERE IT ARMS NOTHING ──────────────────────
--
-- **banned_claims: LIVE and proven.** All 6 patterns compile in the exact consumer form
-- ("(?i)"+p, claims.go:468 - a non-compiling pattern silently falls back to QuoteMeta and is
-- armed, counted and inert), all 6 FIRE on their own positive control through the production
-- scanner, and all 6 return ZERO hits across this site's 23 served pages. The negation guard
-- suppressed 0, so the zero is the site's, not the guard's. The inherited finance patterns were
-- NOT adopted: the vetcomparison lane warned they false-positive here in reverse (this site
-- legitimately says it publishes no prices yet), so every pattern below is written for this site.
--
-- **the daily citation re-check: LIVE** on the 13 cited facts (3 CMA case/consultation pages,
-- 10 gov.uk pet-travel facts), each quote verified through the production matcher on 2026-09-03
-- with an absent control returning false in the same run.
--
-- ⚠ **the number scan: ARMED BUT CURRENTLY UNEXERCISED, and I am stating that rather than
-- implying coverage.** `ScanUnregisteredNumbers` is high-precision: it fires only on a number in
-- a business-claim context (`businessClaimContextRe`). Measured against this site's own pages
-- with a DEMAND CONTROL (the same register with `facts` emptied): both produce 0 findings on all
-- seven non-editorial pages, so that zero was vacuous. The facts DO work where the scan reaches -
-- control flags "The threshold is 15 first opinion practice sites", the full register supports it,
-- and an unrelated "We have 4,000 clients" stays flagged in BOTH (so the register is not blanket-
-- supporting). But the sentences carrying £21/£12.50/36 live on `guide`, `blog-post` and `tool`
-- pages, which `editorialPageTypes` gates OFF by design. **This register does not promise number
-- checking inside the guides.**
--
-- ⚠ `pinned` IS SET AND IS NOT PROTECTION. `write_site_spec` ignores `pinned` and drops it, and the
-- daily refresher supersede-and-reinserts under ITS OWN `created_by` - so this file's ROLLBACK
-- guard expires at the first refresher pass, by design (RUNBOOK §8d, §8e).
--
-- ⚠ `kind:"policy"` is NOT a canonical kind (metric|capability|entity|attestation) and resolves
-- silently to `metric`; `UnrecognisedKinds()` will report it. That is a PRE-EXISTING fleet
-- convention - lendzy, loancash and loancalculator all use it across 39 facts - and this register
-- follows it deliberately so cross-register queries stay consistent. Named here so it is a known
-- condition rather than a discovery.
--
-- ⚠ **EVERY CMA FACT IS DRAFT AND CARRIES `draft_status`.** The statutory deadline for making the
-- substantive Order is 23 September 2026. THE DAY IT IS MADE, ALL OF IT NEEDS RE-VERIFICATION -
-- the bracketed figures become real numbers and the compliance dates stop being "[X March 2027]".
--
-- Deliberate omissions, at the vetcomparison lane's instruction: third-party practice prices
-- (Vet Home Certs' £99/£110 publish under the claim-listing licence with consent snapshotted in
-- business_intel and a source URL + observed_at per row - they are the practice's claims, not the
-- site's), and the site's deliberate silences (no ownership/independence claim, no unclaimed
-- practice prices, the OV-qualification claim held unpublished on purpose). A register that
-- "completes" those publishes something a person chose not to publish.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/bugfix_414_planted_marker_as_claim/
-- Rollback: 759_..._ROLLBACK.sql

BEGIN;

-- Guard 1 (RUNBOOK §8e): bind the UUID to the DOMAIN. Without this a mistyped site_id silently
-- populates ANOTHER site's register and every verify count below still passes, being scoped to
-- the same wrong id.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND domain = 'vetcomparison.uk') THEN
    RAISE EXCEPTION '759 ABORT: site_id does not resolve to vetcomparison.uk';
  END IF;
END $$;

-- Guard 2: refuse if a register already exists - do not silently supersede another author's work.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF n <> 0 THEN
    RAISE EXCEPTION '759 ABORT: vetcomparison already has % current evidence_base row(s) - read them before writing', n;
  END IF;
END $$;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
VALUES (
  '72b9e3a6-872f-4528-a6d6-7f205ea60f4d',
  'evidence_base',
  '{"facts":[{"id":"CMA-FINAL-REPORT-2026-03-24","kind":"policy","claim":"The CMA published the final report of its market investigation into veterinary services for household pets on 24 March 2026.","source":{"citation":{"url":"https://www.gov.uk/cma-cases/veterinary-services-market-for-pets-review","quote":"24 March 2026 Final report published","title":"Veterinary services market for pets review - GOV.UK","accessed":"2026-09-03","publisher":"Competition and Markets Authority"}},"verified_at":"2026-09-03","writer_line":"the CMA''s final report, published 24 March 2026","corrects_site_citation":"LIVE ERROR. /guides/cma-compliance/index.html and /guides/cma-market-investigation/index.html both say the CMA ''published its final report in November 2024''. That is wrong by about sixteen months. November 2024 is when the Inquiry Chair gave a speech to the BVA Congress (also on the case page); the final report is 24 March 2026. Found 2026-09-03 by the register pass."},{"id":"CMA-STATUTORY-DEADLINE-2026-09-23","kind":"policy","claim":"The CMA''s statutory deadline for implementing remedial action in the veterinary services market investigation is 23 September 2026.","source":{"citation":{"url":"https://www.gov.uk/cma-cases/veterinary-services-market-for-pets-review","quote":"23 September 2026 Statutory deadline for implementing remedial action","title":"Veterinary services market for pets review - GOV.UK","accessed":"2026-09-03","publisher":"Competition and Markets Authority"}},"verified_at":"2026-09-03","writer_line":"the 23 September 2026 statutory deadline for implementing remedial action"},{"id":"CMA-DRAFT-ORDER-CONSULTATION-2026-07-21","kind":"policy","claim":"The CMA published the draft substantive Order and Undertakings for formal consultation on 21 July 2026.","source":{"citation":{"url":"https://www.gov.uk/cma-cases/veterinary-services-market-for-pets-review","quote":"21 July 2026 Consultation on draft substantive Order and Undertakings published.","title":"Veterinary services market for pets review - GOV.UK","accessed":"2026-09-03","publisher":"Competition and Markets Authority"}},"verified_at":"2026-09-03","writer_line":"the draft substantive Order, published for consultation on 21 July 2026"},{"id":"GOVUK-AHC-VALIDITY-10-DAYS","kind":"policy","value":10,"context_terms":["animal health certificate","certificate","days","entry into the eu","eu"],"claim":"An animal health certificate is valid for 10 days after the date of issue for entry into the EU.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/getting-an-animal-health-certificate","quote":"10 days for entry into the EU","title":"Taking your pet dog, cat or ferret abroad: Getting an animal health certificate - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the 10-day validity of an animal health certificate for entry into the EU"},{"id":"GOVUK-AHC-ONWARD-6-MONTHS","kind":"policy","value":6,"context_terms":["onward travel","months","animal health certificate","certificate"],"claim":"An animal health certificate is valid for 6 months for onward travel within the EU after entry.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/getting-an-animal-health-certificate","quote":"6 months for onward travel within the EU after you enter the EU","title":"Taking your pet dog, cat or ferret abroad: Getting an animal health certificate - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the 6-month onward-travel validity of an animal health certificate"},{"id":"GOVUK-AHC-REENTRY-6-MONTHS","kind":"policy","value":6,"context_terms":["re-entry","great britain","months","certificate"],"claim":"An animal health certificate is valid for 6 months for re-entry to Great Britain.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/getting-an-animal-health-certificate","quote":"6 months for re-entry to Great Britain","title":"Taking your pet dog, cat or ferret abroad: Getting an animal health certificate - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the 6-month re-entry validity of an animal health certificate"},{"id":"GOVUK-AHC-MAX-5-PETS","kind":"policy","value":5,"context_terms":["pets","animal health certificate","certificate","add up to"],"claim":"Up to 5 pets may be added to one animal health certificate.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/getting-an-animal-health-certificate","quote":"You can add up to 5 pets to an animal health certificate.","title":"Taking your pet dog, cat or ferret abroad: Getting an animal health certificate - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the limit of 5 pets on one animal health certificate"},{"id":"GOVUK-RABIES-MIN-AGE-12-WEEKS","kind":"policy","value":12,"context_terms":["weeks","rabies","vaccinat","old"],"claim":"A pet must be at least 12 weeks old before it can be vaccinated against rabies for EU travel.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/travelling-to-an-eu-country-or-northern-ireland","quote":"Your vet needs proof that your pet is at least 12 weeks old before vaccinating them.","title":"Taking your pet dog, cat or ferret abroad: Travelling to an EU country - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the minimum age of 12 weeks before a rabies vaccination"},{"id":"GOVUK-RABIES-WAIT-21-DAYS","kind":"policy","value":21,"context_terms":["days","rabies","vaccination","wait"],"claim":"A pet owner must wait at least 21 full days after the first rabies vaccination (or the last of the first course) before travelling.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/travelling-to-an-eu-country-or-northern-ireland","quote":"This will be at least 21 full days after the first vaccination","title":"Taking your pet dog, cat or ferret abroad: Travelling to an EU country - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the 21-full-day wait after a first rabies vaccination"},{"id":"GOVUK-TAPEWORM-PRAZIQUANTEL","kind":"policy","claim":"Tapeworm treatment for dogs travelling to Finland, Ireland, Malta, Northern Ireland or Norway must contain praziquantel or an equivalent proven effective against Echinococcus multilocularis.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/tapeworm-treatment-for-dogs","quote":"contain praziquantel or an equivalent proven to be effective against the Echinococcus multilocularis tapeworm","title":"Taking your pet dog, cat or ferret abroad: Tapeworm treatment for dogs - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the praziquantel requirement for tapeworm treatment"},{"id":"GOVUK-TAPEWORM-MIN-24-HOURS","kind":"policy","value":24,"context_terms":["hours","tapeworm","treatment","arrive"],"claim":"Tapeworm treatment must be given no less than 24 hours before arrival.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/tapeworm-treatment-for-dogs","quote":"no less than 24 hours before you arrive","title":"Taking your pet dog, cat or ferret abroad: Tapeworm treatment for dogs - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the 24-hour minimum before arrival for tapeworm treatment"},{"id":"GOVUK-TAPEWORM-MAX-5-DAYS","kind":"policy","value":5,"context_terms":["days","tapeworm","treatment","arrive"],"claim":"Tapeworm treatment must be given no more than 5 days before arrival.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/tapeworm-treatment-for-dogs","quote":"no more than 5 days (120 hours) before you arrive","title":"Taking your pet dog, cat or ferret abroad: Tapeworm treatment for dogs - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the 5-day maximum before arrival for tapeworm treatment"},{"id":"GOVUK-TAPEWORM-MAX-120-HOURS","kind":"policy","value":120,"context_terms":["hours","tapeworm","treatment","arrive"],"claim":"The 5-day maximum for tapeworm treatment before arrival is equivalently expressed as 120 hours.","source":{"citation":{"url":"https://www.gov.uk/taking-your-pet-abroad/tapeworm-treatment-for-dogs","quote":"no more than 5 days (120 hours) before you arrive","title":"Taking your pet dog, cat or ferret abroad: Tapeworm treatment for dogs - GOV.UK","accessed":"2026-09-03","publisher":"GOV.UK"}},"verified_at":"2026-09-03","writer_line":"the 120-hour maximum before arrival for tapeworm treatment"},{"id":"CMA-DRAFT-PRIMARY-PRESCRIPTION-FEE-CAP-21","kind":"policy","value":21,"context_terms":["prescription","fee","cap","written prescription"],"draft_status":"DRAFT — NOT YET LAW. Re-verify the day the substantive Order is made (statutory deadline 23 September 2026).","claim":"The draft Order PROPOSES an Initial Primary Prescription Fee Cap of £21 inclusive of VAT — as a bracketed placeholder, to be adjusted for inflation between the date of the Final Decision Report and the latest monthly CPI figure available before the Order is made. It is not a settled figure.","source":{"attested_by":"414 register-programme lane, 2026-09-03, by reading the primary document: the draft Veterinary Services Market Investigation Order 2026 (392,144 bytes, fetched 2026-09-03 from https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-substantive-orderpdf), pdftotext -layout, definitions section. Verbatim: \"''Initial Primary Prescription Fee Cap'' means [£21 inclusive of VAT. This will be adjusted for inflation between the date of the Final Decision Report and the latest monthly CPI figure available before the Order is made].\" Corroborates the vetcomparison lane''s ATTESTATION_2026-09-03_cma_draft_order_and_govuk_pet_travel.md §4."},"source_document":"https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-substantive-orderpdf","no_citation_because":"The primary source is a PDF. The daily citation re-checker extracts visible text from HTML, so a citation here would fail the quote match EVERY DAY and read as citation_lost drift for ever. Measured 2026-09-03 through the production matcher (cmd/fcaquotecheck): the PDF returns HTTP 200 and 296,699 chars of extracted text, and EVERY quote returns false — including \"Compliance Date\", which is certainly in the document. attested_by instead gets a ~180-day staleness nudge and no false drift.","verified_at":"2026-09-03","writer_line":"the £21 primary prescription fee cap PROPOSED by the draft Order (not yet settled)","corrects_site_citation":"LIVE ERROR, known to the vetcomparison lane since 2026-08-24 and still unfixed. Seven served pages state the £21 cap as a settled figure — /guides/cma-compliance/, /guides/independent-strategy/, /guides/tool-cma-obligation-checker-guide, /guides/tool-cma-transparency-checker-pet-owner-guide, /guides/tool-pet-treatment-cost-estimator-guide, /guides/tool-practice-price-transparency-benchmarker-guide and /how-it-works.html. Recommended honest phrasing (owner''s call, content freeze applies): ''the draft Order proposes a cap of £21, to be finalised — with an inflation adjustment — when the Order is made.''"},{"id":"CMA-DRAFT-ADDITIONAL-PRESCRIPTION-FEE-CAP-12-50","kind":"policy","value":12.5,"context_terms":["additional","medicine","prescription","cap","fee"],"draft_status":"DRAFT — NOT YET LAW. Re-verify the day the substantive Order is made (statutory deadline 23 September 2026).","claim":"The draft Order PROPOSES an Initial Additional Prescription Fee Cap of £12.50 inclusive of VAT — as a bracketed placeholder, to be adjusted for inflation before the Order is made. It is not a settled figure.","source":{"attested_by":"414 register-programme lane, 2026-09-03, by reading the primary document (draft substantive Order, pdftotext -layout, definitions section). Verbatim: \"''Initial Additional Prescription Fee Cap'' means [£12.50 inclusive of VAT. This will be adjusted for inflation between the date of the Final Decision Report and the latest monthly CPI figure available before the Order is made].\""},"source_document":"https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-substantive-orderpdf","no_citation_because":"PDF primary source — see CMA-DRAFT-PRIMARY-PRESCRIPTION-FEE-CAP-21.","verified_at":"2026-09-03","writer_line":"the £12.50 additional-medicine prescription fee cap PROPOSED by the draft Order (not yet settled)","corrects_site_citation":"LIVE ERROR, same seven pages and same shape as the £21 cap: stated as settled when the draft carries it as a bracketed, inflation-adjustable placeholder."},{"id":"CMA-DRAFT-LARGE-BUSINESS-THRESHOLD-15","kind":"policy","value":15,"context_terms":["large","small","veterinary business","fop","first opinion","threshold","ooh"],"draft_status":"DRAFT — NOT YET LAW. Re-verify the day the substantive Order is made.","claim":"Under the draft Order a ''Large Veterinary Business'' is one with 15 or more FOPs and/or OOH Centres; a ''Small Veterinary Business'' has fewer than 15.","source":{"attested_by":"414 register-programme lane, 2026-09-03, from the draft substantive Order (pdftotext -layout), size definitions. Verbatim: \"''Large Veterinary Business'' means a Veterinary Business with 15 or more FOPs and/or OOH Centres.\" / \"''Small Veterinary Business'' means a Veterinary Business with fewer than 15 FOPs and/or OOH Centres.\" Independently transcribed by the vetcomparison lane in ATTESTATION_2026-09-03 §2 and agreeing verbatim."},"source_document":"https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-substantive-orderpdf","no_citation_because":"PDF primary source — see CMA-DRAFT-PRIMARY-PRESCRIPTION-FEE-CAP-21.","verified_at":"2026-09-03","writer_line":"the 15-site threshold between Large and Small Veterinary Businesses"},{"id":"CMA-DRAFT-PRICE-LIST-SERVICES-36","kind":"policy","value":36,"context_terms":["service","services","price list","categories","standard price list"],"draft_status":"DRAFT — NOT YET LAW. Re-verify the day the substantive Order is made.","claim":"Draft Schedule 1 (Price List Schedule) defines 36 services, products, treatments or procedures in total, grouped into 5 numbered categories: Consultation and preventative care (12), Prescription, dispensing and administration (6), Surgeries and treatments (6), Diagnostics and laboratory tests (9), End-of-life care (3).","source":{"attested_by":"414 register-programme lane, 2026-09-03, from draft Schedule 1 (101,304 bytes, pdftotext -layout). Verbatim column heading: \"Service, product, treatment or procedure (36 total)\". The five category counts 12+6+6+9+3 sum to 36, checked."},"source_document":"https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-schedule-1-price-list-schedulepdf","no_citation_because":"PDF primary source — see CMA-DRAFT-PRIMARY-PRESCRIPTION-FEE-CAP-21.","verified_at":"2026-09-03","writer_line":"the 36 services the draft price list schedule defines, across 5 categories","corrects_site_citation":"IMPRECISION, live on at least eight served pages (/about.html, /how-it-works.html x3, /guides/cma-compliance/, /guides/cma-market-investigation/, /guides/independent-strategy/, /guides/tool-cma-obligation-checker-guide, /guides/tool-cma-transparency-checker-pet-owner-guide, /guides/tool-pet-treatment-cost-estimator-guide): the site says ''36 service categories''. The schedule defines 36 SERVICES grouped into 5 CATEGORIES. The number 36 is right; the noun it attaches to is not. Found 2026-09-03 by the register pass."},{"id":"CMA-DRAFT-PRICE-LIST-CATEGORIES-5","kind":"policy","value":5,"context_terms":["categories","price list","grouped","schedule"],"draft_status":"DRAFT — NOT YET LAW. Re-verify the day the substantive Order is made.","claim":"Draft Schedule 1 groups its 36 defined services into 5 numbered categories.","source":{"attested_by":"414 register-programme lane, 2026-09-03, from draft Schedule 1 (pdftotext -layout): five numbered category rows, 1. Consultation and preventative care through 5. End-of-life care."},"source_document":"https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-schedule-1-price-list-schedulepdf","no_citation_because":"PDF primary source — see CMA-DRAFT-PRIMARY-PRESCRIPTION-FEE-CAP-21.","verified_at":"2026-09-03","writer_line":"the 5 categories the draft price list schedule groups its services into"},{"id":"VC-DIRECTORY-LISTINGS-60","kind":"metric","value":60,"context_terms":["listings","directory","listed","practices"],"claim":"The published practice directory carried 60 listings when counted on 2026-09-03.","source":{"sql":"SELECT count(*) FROM business_intel.vet_practice_details;","attested_by":"414 register-programme lane, 2026-09-03: counted at the served artefact — https://vetcomparison.uk/directory/index.html renders the string ''60 listings'' and carries 60 ''Visit website'' links, 51 of them Vet Home Certs branches and 9 other practices. The underlying business_intel.vet_practice_details table holds 2,825 rows, which is the SEARCHABLE population, not the LISTED one; the two numbers answer different questions and must not be substituted for each other."},"verified_at":"2026-09-03","writer_line":"the 60 listings currently published in the practice directory","goes_stale_by_addition":"This is a COUNT and it will drift upward as listings are added, while continuing to read as current. The daily refresher does not re-prove it (only source.citation facts are re-fetched). Re-count at the served page before quoting it, and note that the site''s own directory page carries the figure as static text — the claims scan has already flagged that (claims_unverified, medium): ''Replace with a dynamically generated count or remove the static figure if it cannot be verified and kept current.''"},{"id":"VC-CLAIMING-IS-FREE","kind":"capability","claim":"Claiming a practice listing on VetComparison.uk is free.","source":{"attested_by":"414 register-programme lane, 2026-09-03. True by construction — the site operates no charging mechanism for claims. Served wording, /entities/practice.html: ''Claiming is free, and it lets you correct the details we hold, confirm who owns the practice, and add the standard price list the CMA remedies require larger practices to publish.'' Also /guides/independent-strategy/ (''Claiming a listing is free.'') and /index.html (''free of charge'')."},"verified_at":"2026-09-03","writer_line":"claiming a listing is free"},{"id":"VC-NO-INVENTED-FIGURES","kind":"capability","claim":"VetComparison.uk does not invent figures and will not publish a price it cannot attribute to a source.","source":{"attested_by":"414 register-programme lane, 2026-09-03. This is the site''s own published promise and it is the promise this register exists to make mechanically checkable: the site was remediated on 2026-07-14 for fabricated prices, a false ''proprietary data'' notice and a fabricated CMA quote (LEGAL_2026-07-15_vetcomparison_factual_record.md). Served wording, /entities/practice.html and /how-it-works.html: ''We do not invent figures and we will not publish a price we cannot attribute to a source.'' The banned_claims set below is the enforcement half; this fact is the promise half."},"verified_at":"2026-09-03","writer_line":"we do not invent figures and will not publish a price we cannot attribute to a source"}],"citation_code_presets":["veterinary"],"banned_claims":[{"pattern":"\\bproprietary\\s+(data|dataset|database|research|figures|pricing|information)\\b","must_fire":"Our pricing comes from proprietary data gathered across the sector.","severity":"blocker","why":"THIS SITE''S ORIGINAL SIN. Remediated 2026-07-14 for a false ''proprietary data'' notice alongside fabricated prices and a fabricated CMA quote (LEGAL_2026-07-15_vetcomparison_factual_record.md). The vetcomparison lane''s own words: ''that exact phrase is this site''s original sin.'' The site holds no proprietary dataset; every figure is either a practice''s own published price or an attributed public source."},{"pattern":"\\b(the )?CMA (said|says|stated|states|confirms|confirmed|told us|has confirmed)\\b","must_fire":"The CMA said that every practice must publish a price list by December.","severity":"blocker","why":"The 2026-07-14 remediation included a FABRICATED CMA QUOTE. The CMA''s positions are quotable only from its published documents, with a URL and a verbatim quote. A ''the CMA says'' construction is how an unsourced paraphrase acquires the authority of a regulator."},{"pattern":"\\b(we|vetcomparison(\\.uk)?) (are|is) (fully |completely |truly |genuinely )?independent\\b","must_fire":"We are independent and take no money from practices.","severity":"blocker","why":"A DELIBERATE ABSENCE, not an oversight. No ownership/independence claim is published: the is_independent data is known-contradictory and P2 is unfinished (vetcomparison lane, handover 2026-09-03). A register that ''completes'' this would publish something a person chose not to publish. NOTE the site legitimately discusses PRACTICE independence (/guides/independent-strategy/) — this pattern is scoped to the site asserting it of ITSELF."},{"pattern":"\\b(every|all|complete(ly)?|comprehensive) (uk )?(vet|veterinary) (practices?|surgeries|sites?)\\b[^.]{0,40}\\b(listed|covered|included|indexed)\\b","must_fire":"Every UK veterinary practice is listed and covered in our directory.","severity":"blocker","why":"The directory is not complete and must never claim to be — it carried 60 published listings on 2026-09-03 against a searchable population of 2,825 rows, and neither number is the whole of the UK. A completeness claim on a comparison directory is the claim a pet owner would most reasonably rely on."},{"pattern":"\\b(fully|independently|externally|properly|all) (verified|audited|fact.?checked|confirmed)\\b","must_fire":"All prices are independently verified before publication.","severity":"blocker","why":"The site''s honest position is that it ATTRIBUTES and DATES figures, not that it verifies them. The claims scan has already flagged the served sentence ''Practice details are collected from practices'' own websites and verified before listing'' as unsupported (claims_unverified, high). Verification is a stronger promise than attribution and nothing performs it."},{"pattern":"\\b(average|typical|median|usual|going) (price|cost|fee|charge|rate)s? (is|are|of|for|comes? to)\\b[^.]{0,30}£","must_fire":"The average price for a consultation is £45 across the UK.","severity":"blocker","why":"FABRICATED PRICES were half of the 2026-07-14 remediation, and an averaged figure is the shape an invented price takes on a comparison site — it reads as derived rather than asserted. The site''s own guide says it ''do[es] not fill gaps with averages dressed up as facts''. Attributed third-party prices (Vet Home Certs £99/£110, published under the claim-listing licence with consent snapshotted in business_intel) are NOT this shape and must not be caught: they are named, dated and attributed to the practice."}]}'::jsonb,
  'manual',
  NULL,
  'bugfix_414 register-programme lane (migration 759)',
  true,
  true,
  'Vetcomparison evidence register under owner ruling D1 (2026-09-03). 21 facts: 13 CITED (3 CMA case/consultation + 10 gov.uk pet-travel), each quote verified through datahelpers.QuoteFoundInText via cmd/fcaquotecheck on 2026-09-03 with an absent control returning false in the same run; 8 ATTESTED from the CMA draft Order and draft Schedule 1, which are PDFs the production matcher structurally cannot read (measured: every quote false, including one certainly present, and the control false too - so a citation there would read as citation_lost drift every day for ever). 4 facts carry corrects_site_citation: the final report is 24 March 2026 not November 2024 (NEW), the £21/£12.50 caps are bracketed placeholders served as settled (known since 08-24, unfixed), and "36 service categories" is 36 services in 5 categories (NEW). 6 banned_claims, all compiled in the exact consumer form, all firing on their own control, all 0 hits over 23 served pages with 0 suppressed by the negation guard. Copy not touched - owner decision.'
);

DO $$
DECLARE nfacts int; ncited int; nattested int; ncorr int; nban int; ndraft int; nwrong int; ndomain int;
BEGIN
  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims')
    INTO nfacts, nban
    FROM site_specs WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;

  IF nfacts <> 21 THEN
    RAISE EXCEPTION '759 VERIFY: expected 21 facts, found %', nfacts;
  END IF;
  IF nban <> 6 THEN
    RAISE EXCEPTION '759 VERIFY: expected 6 banned_claims patterns, found %', nban;
  END IF;

  -- a cited fact must carry BOTH a url and a substantive quote, or the daily re-check has
  -- nothing to re-prove and the citation is decoration.
  SELECT count(*) INTO ncited FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->'source'->'citation'->>'url' LIKE 'https://%'
     AND length(f->'source'->'citation'->>'quote') > 20;
  IF ncited <> 13 THEN
    RAISE EXCEPTION '759 VERIFY: expected 13 facts with a https URL and a substantive quote, found %', ncited;
  END IF;

  -- an attested fact must say WHY it has no citation, or the absence reads as an omission.
  SELECT count(*) INTO nattested FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->'source' ? 'attested_by'
     AND NOT (f->'source' ? 'citation');
  IF nattested <> 8 THEN
    RAISE EXCEPTION '759 VERIFY: expected 8 attested-only facts, found %', nattested;
  END IF;

  -- every CMA fact sourced from the draft Order must be tagged draft-vs-final, because the
  -- substantive Order is due 23 September 2026 and the day it is made these all change.
  SELECT count(*) INTO ndraft FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->>'id' LIKE 'CMA-DRAFT-%' AND f ? 'draft_status';
  IF ndraft <> 5 THEN
    RAISE EXCEPTION '759 VERIFY: expected 5 CMA-DRAFT-* facts each tagged draft_status, found %', ndraft;
  END IF;

  SELECT count(*) INTO ncorr FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f ? 'corrects_site_citation';
  IF ncorr <> 4 THEN
    RAISE EXCEPTION '759 VERIFY: expected 4 facts recording a corrected site citation, found %', ncorr;
  END IF;

  -- the three findings must be the ones actually recorded, not merely three of something
  SELECT count(*) INTO nwrong FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->>'corrects_site_citation' LIKE '%November 2024%';
  IF nwrong <> 1 THEN
    RAISE EXCEPTION '759 VERIFY: the November-2024 final-report error is not recorded (found %)', nwrong;
  END IF;

  -- CONTROL (RUNBOOK §8e): read the domain back through a join on sites, never the id typed above,
  -- and prove this migration wrote to no other site.
  SELECT count(*) INTO ndomain FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE ss.created_by = 'bugfix_414 register-programme lane (migration 759)' AND s.domain <> 'vetcomparison.uk';
  IF ndomain <> 0 THEN
    RAISE EXCEPTION '759 VERIFY: this migration wrote to % row(s) on a site that is NOT vetcomparison.uk', ndomain;
  END IF;

  RAISE NOTICE '759 OK: vetcomparison evidence register created - % facts (% cited, % attested-from-PDF), % correcting a live misstatement, % banned-claim patterns, veterinary citation preset', nfacts, ncited, nattested, ncorr, nban;
END $$;

COMMIT;
