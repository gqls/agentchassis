-- 499_robot_hands_compressed_air_facts.sql
-- Substrate for 498: five metrics on the energy cost of compressed air, every
-- quote taken VERBATIM from ENERGY STAR / US DOE, "Determine the Cost of
-- Compressed Air for Your Plant", extracted with pdftotext in-session
-- 2026-08-21. Primary source deliberately, not the compressor-vendor pages that
-- restate these figures second-hand.
--
-- Note on the pair rh-air-hp-in / rh-air-hp-out: both come from ONE sentence
-- (7-8 hp electrical to deliver 1 hp at the air motor). They are registered as
-- two facts because the chart plots them against each other, and each carries
-- the same full sentence as its quote so neither can be read out of context.
-- The 7 is the LOW end of the stated 7-8 range: the conservative end favours the
-- side the piece argues against, which is the direction an honest estimate
-- should err.
--
-- ROLLBACK: supersede the current row and restore the prior one.

\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE cur jsonb;
BEGIN
  SELECT data INTO cur FROM site_specs
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current;
  IF cur IS NULL THEN RAISE EXCEPTION 'no current evidence_base row'; END IF;
  IF cur::text LIKE '%rh-air-efficiency%' THEN
    RAISE EXCEPTION 'compressed-air facts already registered - refusing double apply';
  END IF;
END $$;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
  SELECT * FROM site_specs WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current;

UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT cur.site_id, cur.aspect,
  jsonb_set(cur.data, '{facts}', COALESCE(cur.data->'facts','[]'::jsonb) || $af$[
 {
  "id": "rh-air-share-typical",
  "kind": "metric",
  "claim": "Share of a typical industrial facility's electricity consumed generating compressed air",
  "value": 10,
  "unit": "percent",
  "tolerance": "exact",
  "context_terms": [
   "compressed air",
   "electricity"
  ],
  "writer_line": "about {value}% of a typical plant's electricity goes on generating compressed air",
  "verified_at": "2026-08-21",
  "source": {
   "citation": {
    "publisher": "ENERGY STAR / US Department of Energy",
    "title": "Determine the Cost of Compressed Air for Your Plant",
    "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
    "quote": "A recent survey by the U.S. Department of Energy showed that for a typical industrial facility, approximately 10% of the electricity consumed is for generating compressed air.",
    "accessed": "2026-08-21"
   }
  }
 },
 {
  "id": "rh-air-share-high",
  "kind": "metric",
  "claim": "Share of electricity compressed air may account for at some facilities",
  "value": 30,
  "unit": "percent",
  "tolerance": "exact",
  "context_terms": [
   "compressed air",
   "some facilities"
  ],
  "writer_line": "at some plants it is {value}% or more",
  "verified_at": "2026-08-21",
  "source": {
   "citation": {
    "publisher": "ENERGY STAR / US Department of Energy",
    "title": "Determine the Cost of Compressed Air for Your Plant",
    "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
    "quote": "For some facilities, compressed air generation may account for 30% or more of the electricity consumed.",
    "accessed": "2026-08-21"
   }
  }
 },
 {
  "id": "rh-air-efficiency",
  "kind": "metric",
  "claim": "Overall efficiency of a typical compressed air system, at the low end of the stated range",
  "value": 10,
  "unit": "percent",
  "tolerance": "exact",
  "context_terms": [
   "efficiency",
   "compressed air system"
  ],
  "writer_line": "overall efficiency can be as low as {value}%",
  "verified_at": "2026-08-21",
  "source": {
   "citation": {
    "publisher": "ENERGY STAR / US Department of Energy",
    "title": "Determine the Cost of Compressed Air for Your Plant",
    "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
    "quote": "The overall efficiency of a typical compressed air system can be as low as 10-15%.",
    "accessed": "2026-08-21"
   }
  }
 },
 {
  "id": "rh-air-hp-in",
  "kind": "metric",
  "claim": "Electrical horsepower supplied to the compressor to operate a 1 hp air motor at 100 psig, at the low end of the stated range",
  "value": 7,
  "unit": "hp",
  "tolerance": "exact",
  "context_terms": [
   "air motor",
   "horsepower",
   "compressor"
  ],
  "writer_line": "{value}-8 hp of electrical power to run a 1 hp air motor",
  "verified_at": "2026-08-21",
  "source": {
   "citation": {
    "publisher": "ENERGY STAR / US Department of Energy",
    "title": "Determine the Cost of Compressed Air for Your Plant",
    "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
    "quote": "For example, to operate a 1 hp air motor at 100 psig, approximately 7-8 hp of electrical power is supplied to the air compressor.",
    "accessed": "2026-08-21"
   }
  }
 },
 {
  "id": "rh-air-hp-out",
  "kind": "metric",
  "claim": "Air-motor horsepower delivered for that electrical input, per the same worked example",
  "value": 1,
  "unit": "hp",
  "tolerance": "exact",
  "context_terms": [
   "air motor",
   "horsepower"
  ],
  "writer_line": "delivering {value} hp at the tool",
  "verified_at": "2026-08-21",
  "source": {
   "citation": {
    "publisher": "ENERGY STAR / US Department of Energy",
    "title": "Determine the Cost of Compressed Air for Your Plant",
    "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
    "quote": "For example, to operate a 1 hp air motor at 100 psig, approximately 7-8 hp of electrical power is supplied to the air compressor.",
    "accessed": "2026-08-21"
   }
  }
 }
]$af$::jsonb),
  cur.source, 'news_editorial_features-lane', 'news_editorial_features-lane', true, cur.pinned,
  'Compressed-air energy figures registered 2026-08-21 for the electric-vs-pneumatic editorial feature; quotes verbatim from the DOE/ENERGY STAR PDF via pdftotext'
FROM _cur cur;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs, jsonb_array_elements(data->'facts') f
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current
     AND f->>'id' IN ('rh-air-share-typical','rh-air-share-high','rh-air-efficiency','rh-air-hp-in','rh-air-hp-out');
  IF n <> 5 THEN RAISE EXCEPTION 'expected 5 compressed-air facts, got %', n; END IF;
  SELECT count(*) INTO n FROM site_specs, jsonb_array_elements(data->'facts') f
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current
     AND f->>'id' LIKE 'rh-air-%' AND f->'source'->'citation'->>'quote' IS NULL;
  IF n > 0 THEN RAISE EXCEPTION '% compressed-air fact(s) missing a verbatim quote', n; END IF;
END $$;

COMMIT;
