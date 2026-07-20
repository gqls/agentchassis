-- R4b (2026-07-20) — LIVE FABRICATED STATISTICS on robot-hands.com/about.html.
--
-- Found while fixing the CTA pairing. The about page's stat block was serving:
--
--   "Gripper Models Indexed"   : "1,200+"      -> the index holds FIVE
--   "Actuation Technologies"   : "6"           -> `products.specifications` has
--                                                 NO actuation-type field at all,
--                                                 so this figure has no source
--   "MatchMatrix Queries Run"  : "Tracked & Published"
--                                              -> nothing tracks or publishes it
--
-- Verified live before the fix:
--   curl -s https://robot-hands.com/about.html \
--     | grep -oE '1,200\+|>6<|Gripper Models Indexed|Actuation Technologies'
--   -> 1,200+ / >6< / both labels, all present
--
-- Ground truth:
--   SELECT count(*) FROM products
--    WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND category='gripper';
--   -> 5   (Schunk EGP 40-N-S-B, OnRobot 2FG7, Robotiq 2F-85,
--           Zimmer Group GEP5010IO-00-A, Festo EHPS-20-A-LK)
--
-- This is the same class as the vetcomparison fabrication (997 invented price
-- rows, /bugs_open/020) — unsourced figures served to the public — but it did NOT
-- arrive via the tool-recreation path. It was written straight into
-- page_components.content_data as ordinary page copy, which means the audited-
-- content protections discussed in 020 fix candidate (4) would not have caught
-- it. Recorded in NOTES + WRONG_CALLS.
--
-- Replacements below are all checkable against the DB: 5 gripper models, 5
-- manufacturers, 4 parameters the tool actually compares (payload, jaw travel,
-- gripping force, IP rating).
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. The three fabricated statistics -> figures with a source.
-- ---------------------------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('stat_1_label', 'Gripper Models Indexed')
    || jsonb_build_object('stat_1_value', '5')
    || jsonb_build_object('stat_2_label', 'Manufacturers Covered')
    || jsonb_build_object('stat_2_value', '5')
    || jsonb_build_object('stat_3_label', 'Parameters Compared')
    || jsonb_build_object('stat_3_value', '4'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'stat_1_value' = '1,200+';

-- ---------------------------------------------------------------------------
-- 2. The body copy makes the same unsupported six-technology claim.
-- ---------------------------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('body_text',
         'Robot-Hands.com is a specification platform for automation engineers, '
      || 'system integrators and manufacturing engineers who need to compare '
      || 'robotic grippers against a single consistent benchmark, rather than '
      || 'reconciling incompatible datasheets from half a dozen vendor sites. '
      || 'The index is deliberately small and fully sourced: every gripper listed '
      || 'carries the figures its manufacturer publishes, normalised so they can '
      || 'be compared like for like, and any parameter a manufacturer does not '
      || 'publish is shown as unpublished rather than estimated. MatchMatrix '
      || 'calculates the holding force an application genuinely requires and tests '
      || 'each indexed gripper against it, showing the working. The methodology is '
      || 'published in full and the sources are named. No preferred vendors, no '
      || 'sponsored placements, no black-box rankings.'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'body_text' ILIKE '%soft-robotic%'
  AND p.name = 'about';

-- ---------------------------------------------------------------------------
-- 3. This component keys its CTA on `cta_label`, not `cta_text`, so the R4
--    pairing pass missed it. Label says methodology; send it to the methodology.
-- ---------------------------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('cta_url', '/matchmatrix-methodology.html')
    || jsonb_build_object('cta_target_title', 'MatchMatrix Methodology | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'cta_label' ILIKE '%methodolog%'
  AND pc.content_data->>'cta_url' = '/tools/matchmatrix/index.html';

-- ---------------------------------------------------------------------------
-- VERIFY
-- ---------------------------------------------------------------------------
\echo '--- any remaining unsourced volume claims (expect 0) ---'
SELECT p.name, e.k, e.v
FROM page_components pc JOIN pages p ON p.id = pc.page_id,
LATERAL jsonb_each_text(pc.content_data) AS e(k,v)
WHERE p.site_id = :'site'
  AND (e.v ~ '[0-9],[0-9]{3}\+' OR e.v ILIKE '%1,200%')
ORDER BY p.name;

\echo '--- about stat block now ---'
SELECT pc.content_data->>'stat_1_label' AS l1, pc.content_data->>'stat_1_value' AS v1,
       pc.content_data->>'stat_2_label' AS l2, pc.content_data->>'stat_2_value' AS v2,
       pc.content_data->>'stat_3_label' AS l3, pc.content_data->>'stat_3_value' AS v3
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = :'site' AND p.name = 'about'
  AND pc.content_data ? 'stat_1_value';

COMMIT;
