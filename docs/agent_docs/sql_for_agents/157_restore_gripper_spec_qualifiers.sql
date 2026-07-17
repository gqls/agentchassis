-- 157: restore the spec qualifiers the first working refresher run stripped
--
-- Context (2026-07-17, empty_sections_loop_integrity Session 10). The
-- refresh_product_specs fix (v1.0.1128) made the refresher work for the first
-- time: 4/5 products refreshed, every value literally present on its source
-- page, no fabrication, no key added or lost. But it DEGRADED five values by
-- replacing hand-verified text with the barer literal from the page's value
-- cell — spec tables split meaning across label and value ("Stroke per jaw |
-- 6 mm"), so the model extracted "6 mm" where the human had recorded "6 mm per
-- jaw". For a parallel gripper that silently HALVES the stated stroke.
--
-- Not yet served to users: pages are rendered artifacts and robot-hands still
-- serves the good values; the next rebuild would have shipped the weaker ones.
--
-- Fixed at source in the same session: the merge step now refuses to trade a
-- richer value for a barer restatement of itself (specValueIsRestatement, with
-- these exact pairs as test cases). A genuine change ("30 N" -> "45 N") and an
-- enrichment ("11 kg" -> "11 kg (24.3 lb)") still land. This migration repairs
-- the rows the pre-guard run already wrote.
--
-- verified_date is deliberately LEFT at 2026-07-17: the pages really were
-- re-verified that day and the figures really do match. Only the human's
-- qualifying wording is being restored, not the verification claim.
--
-- Verify after applying:
--   SELECT name, specifications->>'stroke', specifications->>'payload',
--          specifications->>'voltage', specifications->>'interface'
--   FROM products WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
--     AND category='gripper' ORDER BY name;
--   -- expect: "6 mm per jaw", "0.15 kg (recommended workpiece weight)",
--   --         "24 V DC", "10 mm per jaw", "I/O (IO-Link option)"

BEGIN;

-- Schunk EGP 40-N-S-B: stroke, payload, voltage
UPDATE products
SET specifications = specifications
        || jsonb_build_object(
             'stroke',  '6 mm per jaw',
             'payload', '0.15 kg (recommended workpiece weight)',
             'voltage', '24 V DC'
           ),
    updated_at = NOW()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND name = 'Schunk EGP 40-N-S-B';

-- Zimmer Group GEP5010IO-00-A: stroke, interface
UPDATE products
SET specifications = specifications
        || jsonb_build_object(
             'stroke',    '10 mm per jaw',
             'interface', 'I/O (IO-Link option)'
           ),
    updated_at = NOW()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND name = 'Zimmer Group GEP5010IO-00-A';

COMMIT;
