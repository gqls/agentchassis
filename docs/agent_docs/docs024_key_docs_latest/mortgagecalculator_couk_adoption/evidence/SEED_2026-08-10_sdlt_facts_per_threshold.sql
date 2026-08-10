-- A1 (PLAN_2026-08-09 Piece 4(i)) -- one SDLT fact per threshold and per rate.
--
-- 4 facts -> 13. No schema change: every `value` stays scalar, and each band
-- edge now carries its OWN verbatim GOV.UK quote, so an oracle can read a band
-- table off the register instead of parsing prose out of a `claim` string.
--
-- Every quote was lifted programmatically from the visible text produced by the
-- REAL Go extractor (datahelpers.VisibleTextFromHTML) and re-checked with
-- datahelpers.QuoteFoundInText against the live pages before this file was
-- written; the check was induced red first (a quote reading 126,000 instead of
-- 125,000 reports NOTFOUND and refuses to emit). Nothing here was retyped.
--
-- Retired: `sdlt-standard-bands` (value 12, bands in prose) and
-- `sdlt-ftb-nil-rate` (value 300000, two rules in one claim) -- both are now
-- covered by granular facts. Checked before retiring: zero references in
-- doc_plans, site_work_items or page_components; only this lane's own docs.
-- Kept by id: `sdlt-ftb-relief-cap` and `sdlt-additional-surcharge` (already
-- one number each, and named in the plan's induced-drift test).
-- NEW: `sdlt-additional-surcharge-floor` (40000, cited to the higher-rates
-- guidance) -- registered because the 2026-08-10 rebuild DROPPED the tool's
-- 40,000 floor, which was true law absent from the register.

BEGIN;

UPDATE site_specs SET is_current = false
 WHERE site_id = '62b5978e-4271-4589-8e00-4baebfc0447c' AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, pinned)
VALUES (
  '62b5978e-4271-4589-8e00-4baebfc0447c',
  'evidence_base',
  $spec${
  "facts": [
    {
      "id": "sdlt-standard-nil-band-upper",
      "kind": "metric",
      "unit": "GBP",
      "claim": "No SDLT is due on the portion of the price up to £125,000 (standard residential rates)",
      "value": 125000,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "Up to £125,000 Zero",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "Standard stamp duty is nothing on the first £{value} of the price"
    },
    {
      "id": "sdlt-standard-rate-125k-250k",
      "kind": "metric",
      "unit": "percent",
      "claim": "2% applies to the portion of the price from £125,001 to £250,000",
      "value": 2,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "The next £125,000 (the portion from £125,001 to £250,000) 2%",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "{value}% is charged on the portion from £125,001 to £250,000"
    },
    {
      "id": "sdlt-standard-band-250k-upper",
      "kind": "metric",
      "unit": "GBP",
      "claim": "The 2% band ends at £250,000 — above it the 5% band begins",
      "value": 250000,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "The next £125,000 (the portion from £125,001 to £250,000) 2%",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365
    },
    {
      "id": "sdlt-standard-rate-250k-925k",
      "kind": "metric",
      "unit": "percent",
      "claim": "5% applies to the portion of the price from £250,001 to £925,000",
      "value": 5,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "The next £675,000 (the portion from £250,001 to £925,000) 5%",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "{value}% is charged on the portion from £250,001 to £925,000"
    },
    {
      "id": "sdlt-standard-band-925k-upper",
      "kind": "metric",
      "unit": "GBP",
      "claim": "The 5% band ends at £925,000 — above it the 10% band begins",
      "value": 925000,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "The next £675,000 (the portion from £250,001 to £925,000) 5%",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365
    },
    {
      "id": "sdlt-standard-rate-925k-1500k",
      "kind": "metric",
      "unit": "percent",
      "claim": "10% applies to the portion of the price from £925,001 to £1.5 million",
      "value": 10,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "The next £575,000 (the portion from £925,001 to £1.5 million) 10%",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "{value}% is charged on the portion from £925,001 to £1.5 million"
    },
    {
      "id": "sdlt-standard-band-1500k-upper",
      "kind": "metric",
      "unit": "GBP",
      "claim": "The 10% band ends at £1.5 million — above it the top rate applies",
      "value": 1500000,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "The next £575,000 (the portion from £925,001 to £1.5 million) 10%",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365
    },
    {
      "id": "sdlt-standard-top-rate",
      "kind": "metric",
      "unit": "percent",
      "claim": "12% applies to the portion of the price above £1.5 million",
      "value": 12,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "The remaining amount (the portion above £1.5 million) 12%",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "{value}% is charged on the portion above £1.5 million"
    },
    {
      "id": "sdlt-ftb-nil-band-upper",
      "kind": "metric",
      "unit": "GBP",
      "claim": "First-time buyers pay no SDLT on the portion of the price up to £300,000",
      "value": 300000,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "no SDLT up to £300,000",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "First-time buyers pay no stamp duty up to £{value}"
    },
    {
      "id": "sdlt-ftb-rate-300k-500k",
      "kind": "metric",
      "unit": "percent",
      "claim": "First-time buyers pay 5% on the portion of the price from £300,001 to £500,000",
      "value": 5,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "5% SDLT on the portion from £300,001 to £500,000",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "First-time buyers pay {value}% on the portion from £300,001 to £500,000"
    },
    {
      "id": "sdlt-ftb-relief-cap",
      "kind": "metric",
      "unit": "GBP",
      "claim": "First-time buyer relief cannot be claimed at all if the price is over £500,000 — standard rates apply to the whole purchase",
      "value": 500000,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "If the price is over £500,000, you cannot claim the relief.",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "Above £{value} first-time buyer relief disappears entirely — the standard bands apply to the whole price"
    },
    {
      "id": "sdlt-additional-surcharge",
      "kind": "metric",
      "unit": "percent",
      "claim": "Buying an additional residential property usually adds 5 percentage points on top of each SDLT band",
      "value": 5,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
          "quote": "You’ll usually have to pay 5% on top of SDLT rates if buying a new residential property means you’ll own more than one.",
          "title": "Stamp Duty Land Tax: Residential property rates",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "An additional residential property usually costs {value} percentage points on top of each stamp duty band"
    },
    {
      "id": "sdlt-additional-surcharge-floor",
      "kind": "metric",
      "unit": "GBP",
      "claim": "The higher (additional property) rates apply to purchases of £40,000 or more",
      "value": 40000,
      "source": {
        "citation": {
          "url": "https://www.gov.uk/guidance/stamp-duty-land-tax-buying-an-additional-residential-property",
          "quote": "You must pay the higher Stamp Duty Land Tax (SDLT) rates when you buy a residential property (or a part of one) for £40,000 or more",
          "title": "Stamp Duty Land Tax: buying an additional residential property",
          "accessed": "2026-08-10",
          "published": "",
          "publisher": "GOV.UK"
        }
      },
      "verified_at": "2026-08-10",
      "staleness_days": 365,
      "writer_line": "The additional-property surcharge applies to purchases of £{value} or more"
    }
  ],
  "writer_block": "NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):\n- Standard stamp duty is nothing on the first £125,000 of the price\n- 2% is charged on the portion from £125,001 to £250,000\n- 5% is charged on the portion from £250,001 to £925,000\n- 10% is charged on the portion from £925,001 to £1.5 million\n- 12% is charged on the portion above £1.5 million\n- First-time buyers pay no stamp duty up to £300,000\n- First-time buyers pay 5% on the portion from £300,001 to £500,000\n- Above £500,000 first-time buyer relief disappears entirely — the standard bands apply to the whole price\n- An additional residential property usually costs 5 percentage points on top of each stamp duty band\n- The additional-property surcharge applies to purchases of £40,000 or more",
  "governing_rule": "This site's calculators and copy assert UK tax and product rules (SDLT bands, reliefs, surcharges). Every rate, band or threshold stated anywhere on the site must trace to a fact below, each citing its official source verbatim. A calculator is a claim about legislation: when a fact here changes or its citation is lost, every tool that encodes it is suspect until re-verified. Where a rule is not in this register, pages must not assert it — silence is publishable, a guessed rate is not.",
  "writer_block_managed": true
}$spec$::jsonb,
  'adoption',
  'mortgagecalculator-adoption-lane: A1 per-threshold SDLT facts',
  true,
  true
);

DO $$
DECLARE n_facts int; n_cited int; n_current int; n_ids int;
BEGIN
  SELECT jsonb_array_length(data->'facts'),
         (SELECT count(*) FROM jsonb_array_elements(data->'facts') f
           WHERE f->'source'->'citation'->>'quote' <> ''),
         (SELECT count(DISTINCT f->>'id') FROM jsonb_array_elements(data->'facts') f)
    INTO n_facts, n_cited, n_ids
    FROM site_specs
   WHERE site_id = '62b5978e-4271-4589-8e00-4baebfc0447c' AND aspect = 'evidence_base' AND is_current;

  SELECT count(*) INTO n_current FROM site_specs
   WHERE site_id = '62b5978e-4271-4589-8e00-4baebfc0447c' AND aspect = 'evidence_base' AND is_current;

  IF n_facts <> 13 THEN
    RAISE EXCEPTION 'expected 13 facts, found %', n_facts;
  END IF;
  IF n_cited <> 13 THEN
    RAISE EXCEPTION 'expected every fact to carry a quote, only % do', n_cited;
  END IF;
  IF n_ids <> 13 THEN
    RAISE EXCEPTION 'fact ids are not unique: % distinct of 13', n_ids;
  END IF;
  IF n_current <> 1 THEN
    RAISE EXCEPTION 'expected exactly one current evidence_base row, found %', n_current;
  END IF;
END $$;

COMMIT;

