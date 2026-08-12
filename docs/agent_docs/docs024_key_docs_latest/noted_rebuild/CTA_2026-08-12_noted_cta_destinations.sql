-- =============================================================================
-- noted.co.uk — resolve the CTA destinations the framework could not resolve
-- 2026-08-12. Owner decision: primary CTAs point at the live engine
-- (app.noted.co.uk); migration copy points at /migrate.html.
-- =============================================================================
--
-- WHY THIS FILE EXISTS
-- The build wrote CTA *text* ("Sign in", "See how it works") but no URLs, and both
-- templates gate each button on having BOTH:
--     hero:           {{if and .cta_text .cta_url}} / {{if and .secondary_cta .secondary_cta_url}}
--     call-to-action: {{if and .primary_cta .primary_cta_url}} / {{if and .secondary_cta .secondary_cta_url}}
-- (read from content_components.html_template, ids 23f95f00… and 0197e8d7…, not
-- from the work-item summaries). So every hero and call-to-action rendered ZERO
-- anchors and nothing on the site linked to the product. Six `unresolved_cta`
-- items sat in needs_human_review; their own fix note says "set the destination
-- manually", which is what this does.
--
-- WHAT IT DELIBERATELY DOES NOT DO
-- `migrate`'s primary CTA is "Save everything" — a LOCAL-DATA RESCUE, not a
-- sign-in. Its real destination is the `/legacy` page that PLAN §4 step 3 has not
-- built yet. Pointing it at app.noted.co.uk would misdescribe the action, and
-- pointing it at /legacy.html now would ship a 404 into a platform that actively
-- detects unbuilt internal links. So the two `migrate` PRIMARY urls are left
-- unset: the template's designed degraded state is "render no button", which is
-- honest, and those two work items stay open. Set them when /legacy exists.
--
-- MERGE, NEVER REPLACE: `content_data || jsonb_build_object(...)` keeps the copy
-- the framework wrote. Only destinations are added here — no wording is changed
-- (owner ruling 2026-08-06: the framework writes the content, not us).
--
-- ⚠ LANDMINE these keys live inside: a REGENERATION replaces content_data and
-- would DROP them, while a rerender merges (see memory `bugfix 238`). If the
-- copy is ever regenerated, re-run this file and re-render.
-- =============================================================================

BEGIN;

-- Sanity: fail loudly rather than update 0 rows if the shape moved.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE s.domain = 'noted.co.uk' AND pc.slot_name IN ('hero','call-to-action')
    AND p.name IN ('index','how-it-works','migrate');
  IF n <> 6 THEN
    RAISE EXCEPTION 'expected 6 hero/call-to-action components on index|how-it-works|migrate, found %', n;
  END IF;
END $$;

-- --- index -------------------------------------------------------------------
-- hero:           "Sign in"        + "See how it works"
-- call-to-action:  "Sign in"        + "See how it works"
UPDATE page_components pc SET content_data = pc.content_data || jsonb_build_object(
         'cta_url',           'https://app.noted.co.uk/',
         'secondary_cta_url', '/how-it-works.html')
FROM pages p, sites s
WHERE p.id = pc.page_id AND s.id = p.site_id
  AND s.domain = 'noted.co.uk' AND p.name = 'index' AND pc.slot_name = 'hero';

UPDATE page_components pc SET content_data = pc.content_data || jsonb_build_object(
         'primary_cta_url',   'https://app.noted.co.uk/',
         'secondary_cta_url', '/how-it-works.html')
FROM pages p, sites s
WHERE p.id = pc.page_id AND s.id = p.site_id
  AND s.domain = 'noted.co.uk' AND p.name = 'index' AND pc.slot_name = 'call-to-action';

-- --- how-it-works ------------------------------------------------------------
-- hero:           "Open Noted"     + "Bring your notes with you"   -> /migrate.html
-- call-to-action:  "Sign in"        + "Already have notes somewhere else? …" -> /migrate.html
UPDATE page_components pc SET content_data = pc.content_data || jsonb_build_object(
         'cta_url',           'https://app.noted.co.uk/',
         'secondary_cta_url', '/migrate.html')
FROM pages p, sites s
WHERE p.id = pc.page_id AND s.id = p.site_id
  AND s.domain = 'noted.co.uk' AND p.name = 'how-it-works' AND pc.slot_name = 'hero';

UPDATE page_components pc SET content_data = pc.content_data || jsonb_build_object(
         'primary_cta_url',   'https://app.noted.co.uk/',
         'secondary_cta_url', '/migrate.html')
FROM pages p, sites s
WHERE p.id = pc.page_id AND s.id = p.site_id
  AND s.domain = 'noted.co.uk' AND p.name = 'how-it-works' AND pc.slot_name = 'call-to-action';

-- --- migrate -----------------------------------------------------------------
-- SECONDARY ONLY ("See how it works"). Primary "Save everything" intentionally
-- left with no url until /legacy is built — see the header.
UPDATE page_components pc SET content_data = pc.content_data || jsonb_build_object(
         'secondary_cta_url', '/how-it-works.html')
FROM pages p, sites s
WHERE p.id = pc.page_id AND s.id = p.site_id
  AND s.domain = 'noted.co.uk' AND p.name = 'migrate' AND pc.slot_name IN ('hero','call-to-action');

-- --- close the four work items this fully resolves ---------------------------
-- migrate's two stay in needs_human_review: their primary is still unresolved.
--
-- ⚠ The join key here is `spec->>'page_name'`, NOT `page_id`. An earlier draft of
-- this file used page_id, which DOES NOT EXIST in this spec shape (the keys are
-- component / fix / missing / page_name / section_name / source) — so it would
-- have matched zero rows, updated nothing, and committed successfully. Hence the
-- row-count assertion below: a hand-written status change with no assertion is
-- indistinguishable from a no-op.
DO $$
DECLARE n int;
BEGIN
  UPDATE site_work_items wi
  SET status = 'complete', updated_at = now(),
      spec = wi.spec || jsonb_build_object(
        'resolution', 'Destinations set by hand 2026-08-12 (CTA_2026-08-12_noted_cta_destinations.sql). '
                   || 'Primary -> https://app.noted.co.uk/ ; migration copy -> /migrate.html. '
                   || 'No page-hub was needed: this product''s CTA target is a different origin.')
  FROM sites s
  WHERE s.id = wi.site_id AND s.domain = 'noted.co.uk'
    AND wi.item_type = 'unresolved_cta' AND wi.status = 'needs_human_review'
    AND wi.spec->>'page_name' IN ('index','how-it-works');
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 4 THEN
    RAISE EXCEPTION 'expected to close 4 unresolved_cta items (index x2, how-it-works x2), closed %', n;
  END IF;
  RAISE NOTICE 'closed % unresolved_cta items; migrate x2 deliberately left open', n;
END $$;

COMMIT;

-- =============================================================================
-- VERIFY — run after applying. Every row should show a url for each text it has,
-- except migrate's two primaries, which must remain absent BY DESIGN.
-- =============================================================================
SELECT p.name AS page, pc.slot_name AS slot,
       COALESCE(pc.content_data->>'cta_text', pc.content_data->>'primary_cta') AS primary_text,
       COALESCE(pc.content_data->>'cta_url',  pc.content_data->>'primary_cta_url', '(none — expected on migrate)') AS primary_url,
       pc.content_data->>'secondary_cta'     AS secondary_text,
       COALESCE(pc.content_data->>'secondary_cta_url','(none)') AS secondary_url
FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'noted.co.uk' AND pc.slot_name IN ('hero','call-to-action')
ORDER BY p.name, pc.position;

SELECT p.name AS page, wi.status, count(*)
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
LEFT JOIN pages p ON p.id = (wi.spec->>'page_id')::uuid
WHERE s.domain = 'noted.co.uk' AND wi.item_type = 'unresolved_cta'
GROUP BY 1,2 ORDER BY 1,2;
