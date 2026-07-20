-- R4 (2026-07-20) — MatchMatrix tool built; CTA label↔URL pairs corrected.
--
-- CONTEXT
-- `/tools/matchmatrix/index.html` 404'd. It is now a real, hand-built interactive
-- tool (deployed to gqls/sites @ 0a6dc426, 38,144 B, 19/19 logic tests passing).
-- It computes F = (m·a·S)/(μ·n) and matches that against the FIVE grippers
-- actually held in `products` for this site, showing each manufacturer figure as
-- published and marking unpublished fields rather than inferring them.
--
-- It was hand-authored deliberately. `tool-generator` has NO fake-data rule
-- (`has_no_fake_data_rule=f`) and forbids fetch, so a data-backed tool routed
-- through it is structurally pushed to invent a gripper catalogue — /bugs_open/020.
--
-- THE REAL DEFECT THIS FIXES
-- `/tools/matchmatrix/index.html` had become a DEFAULT DUMPING GROUND for CTAs.
-- 20 components across 11 pages pointed at it, but only ~6 of their labels named
-- MatchMatrix. The rest said "Search the Gripper Catalog", "Browse the Learning
-- Center", "Open the Payload Calculator", "Request Integration Support" — each
-- with a real 200 destination sitting unused.
--
-- Secondary CTAs were worse: 20 of them, essentially ALL mismatched, and NONE of
-- them 404 — 14 labelled "Read the MatchMatrix Methodology" pointed at
-- /services.html while /matchmatrix-methodology.html served 200 the whole time.
-- Nothing flagged them precisely because they were not broken links.
--
-- This is /bugs_open/023 (label and URL are unrelated schema fields, nothing
-- pairs them) in a severe instance, and it is why this site has 20
-- `cta_names_unknown_destination` items parked in needs_human_review — detected,
-- never consumed (/bugs_open/033).
--
-- METHOD: every UPDATE below is keyed on the LABEL, not on the old URL, so the
-- destination follows what the button actually says. Repointing by URL alone
-- would have cemented the mismatches.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. PRIMARY CTAs — route by what the label says.
-- ---------------------------------------------------------------------------

-- 1a. Labels that genuinely name MatchMatrix. The URL was always right; the
--     page simply did not exist until now. These need no change beyond existing.
--     Listed explicitly so the audit trail shows they were considered:
--       about/call-to-action          "Run a MatchMatrix Query"
--       how-it-works/call-to-action   "Run a MatchMatrix Query"
--       how-to-specify/call-to-action "Open MatchMatrix"
--       index/call-to-action          "Run a MatchMatrix Query"
--       matchmatrix/hero              "Run MatchMatrix"
--       matchmatrix/call-to-action    "Run MatchMatrix"
--       product-detail/call-to-action "Run MatchMatrix on This Model"

-- 1b. Labels naming the GRIPPER CATALOG -> /gripper-catalog.html (live, 200).
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('cta_url', '/gripper-catalog.html')
    || jsonb_build_object('cta_target_title', 'Gripper Catalog | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'cta_text' IN ('Search the Gripper Catalog', 'Start with the Catalog');

UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('primary_cta_url', '/gripper-catalog.html')
    || jsonb_build_object('primary_cta_target_title', 'Gripper Catalog | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'primary_cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'primary_cta' IN ('Open the Gripper Catalog', 'Browse the Gripper Catalog');

-- 1c. Labels naming the PAYLOAD CALCULATOR -> its real tool page (live, 200).
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('cta_url', '/tools/gripper-payload-calculator/index.html')
    || jsonb_build_object('cta_target_title', 'Gripper Payload Calculator | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'cta_text' = 'Run Payload Calculation';

UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('primary_cta_url', '/tools/gripper-payload-calculator/index.html')
    || jsonb_build_object('primary_cta_target_title', 'Gripper Payload Calculator | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'primary_cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'primary_cta' = 'Open the Payload Calculator';

-- 1d. Labels naming the METHODOLOGY -> /matchmatrix-methodology.html (live, 200).
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('cta_url', '/matchmatrix-methodology.html')
    || jsonb_build_object('cta_target_title', 'MatchMatrix Methodology | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'cta_text' IN ('Explore the MatchMatrix Methodology', 'Read the Full Methodology');

-- 1e. Labels naming the SELECTION GUIDE -> /gripper-selection-guide.html (200).
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('cta_url', '/gripper-selection-guide.html')
    || jsonb_build_object('cta_target_title', 'Gripper Selection Guide | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'cta_text' = 'Start the Selection Guide';

-- 1f. Labels naming the LEARNING CENTER -> the hub R3 made canonical (200).
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('cta_url', '/learning-center-hub.html')
    || jsonb_build_object('cta_target_title', 'Learning Center | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'cta_text' = 'Browse the Learning Center';

UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('primary_cta_url', '/learning-center-hub.html')
    || jsonb_build_object('primary_cta_target_title', 'Learning Center | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'primary_cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'primary_cta' = 'Browse the Learning Center';

-- 1g. "Request Integration Support" is a contact action, not a tool -> /contact.html.
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('primary_cta_url', '/contact.html')
    || jsonb_build_object('primary_cta_target_title', 'Contact | Robot-Hands.com'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'primary_cta_url' = '/tools/matchmatrix/index.html'
  AND pc.content_data->>'primary_cta' = 'Request Integration Support';

-- ---------------------------------------------------------------------------
-- 2. SECONDARY CTAs — 20 of them, none 404, essentially all mispaired.
-- ---------------------------------------------------------------------------

-- 2a. Anything labelled as METHODOLOGY / "how we benchmark" / "how it works"
--     -> /matchmatrix-methodology.html. This is the 14 x /services.html plus the
--     5 x /tools/gripper-payload-calculator/index.html.
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('secondary_cta_url', '/matchmatrix-methodology.html'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data ? 'secondary_cta_url'
  AND pc.content_data->>'secondary_cta_url' IN
      ('/services.html', '/tools/gripper-payload-calculator/index.html')
  AND (pc.content_data->>'secondary_cta' ILIKE '%methodolog%'
    OR pc.content_data->>'secondary_cta' ILIKE '%how we benchmark%'
    OR pc.content_data->>'secondary_cta' ILIKE '%how matchmatrix works%');

-- 2b. "Browse the Gripper Catalog" as a secondary -> the catalog.
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('secondary_cta_url', '/gripper-catalog.html'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'secondary_cta' = 'Browse the Gripper Catalog';

-- 2c. "Run a MatchMatrix Query" as a secondary -> the tool that now exists.
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('secondary_cta_url', '/tools/matchmatrix/index.html'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'secondary_cta' ILIKE '%Run a MatchMatrix%';

-- ---------------------------------------------------------------------------
-- 3. TOOL LIST — drop the two cards advertising tools that do not exist.
--     tool-matchmatrix is now REAL and stays. tool-robot-payload-budget-
--     calculator has no page at all (owner ruling 2026-07-20: remove).
--     Both dead cards also carried an empty `image`, so they already rendered
--     worse than their three siblings.
-- ---------------------------------------------------------------------------
UPDATE page_components pc SET
  content_data = jsonb_set(
    pc.content_data, '{items}',
    (SELECT COALESCE(jsonb_agg(i), '[]'::jsonb)
       FROM jsonb_array_elements(pc.content_data->'items') AS i
      WHERE i->>'url' <> '/tools/robot-payload-budget-calculator/index.html')
  ),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data ? 'items'
  AND pc.content_data->'items' @> '[{"url":"/tools/robot-payload-budget-calculator/index.html"}]';

-- ---------------------------------------------------------------------------
-- 4. OVERSTATED COPY on the components touched above.
--     The site promised filtering "across the full catalog" over six actuation
--     technologies. There are FIVE grippers, and `products.specifications` holds
--     NO actuation-type field at all — so that filter had no basis in any data.
--     Corrected to what the index can actually support.
-- ---------------------------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('subheadline',
         'Compare gripper and end-effector hardware on the parameters that decide a '
         || 'specification — payload, jaw travel, gripping force and IP rating. '
         || 'MatchMatrix calculates the holding force your application actually needs, '
         || 'then tests it against every gripper in the index and shows each '
         || 'manufacturer figure exactly as published.'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND p.name = 'index'
  AND pc.content_data->>'subheadline' ILIKE '%soft-robotic%';

UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('subheadline',
         'Filter by payload, jaw travel, gripping force and IP rating across the '
         || 'grippers held in the index, then generate a shortlist you can defend '
         || 'in a design review — with every figure traceable to its manufacturer.'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'subheadline' ILIKE '%full catalog%';

UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('section_intro',
         'Parametric calculators and selection tools built around the decisions that '
         || 'actually matter in gripper specification — gripping force, payload '
         || 'margin, jaw travel and cycle time.')
    || jsonb_build_object('cta_supporting_text',
         'Run them in sequence or independently; MatchMatrix takes the force figure '
         || 'the calculators produce and tests it against the indexed grippers.'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND p.name = 'index'
  AND pc.content_data ? 'section_intro'
  AND pc.content_data->>'section_intro' ILIKE '%soft-robotic%';

-- ---------------------------------------------------------------------------
-- VERIFY (expect: 0 rows left pointing at a dead URL, 0 mispaired secondaries)
-- ---------------------------------------------------------------------------
\echo '--- primary CTAs still on the old dumping-ground URL, by label ---'
SELECT p.name, cc.name,
       COALESCE(pc.content_data->>'cta_text', pc.content_data->>'primary_cta') AS label,
       COALESCE(pc.content_data->>'cta_url',  pc.content_data->>'primary_cta_url') AS url
FROM page_components pc JOIN pages p ON p.id = pc.page_id
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = :'site'
  AND COALESCE(pc.content_data->>'cta_url', pc.content_data->>'primary_cta_url')
      = '/tools/matchmatrix/index.html'
ORDER BY p.name;

\echo '--- secondary CTAs labelled methodology but NOT pointing at it (expect 0) ---'
SELECT p.name, pc.content_data->>'secondary_cta', pc.content_data->>'secondary_cta_url'
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = :'site'
  AND pc.content_data->>'secondary_cta' ILIKE '%methodolog%'
  AND pc.content_data->>'secondary_cta_url' <> '/matchmatrix-methodology.html';

\echo '--- dead tool cards remaining (expect 0) ---'
SELECT p.name, jsonb_array_length(pc.content_data->'items') AS cards
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = :'site' AND pc.content_data ? 'items';

COMMIT;
