-- 664 — point finetuning.uk's eight orphaned hero images at the pages they were made for
--
-- bugs_open/412 §9. The 2026-08-26 canary generated nine hero images, deployed all nine, and wired
-- exactly ONE (`careers`). The other eight have been sitting in the public bucket ever since,
-- resolving fine and referenced by nothing.
--
-- ⚠ THIS IS THE CHEAP REMEDY, NOT THE STRUCTURAL FIX. The structural defect — nothing joins a
-- deployed asset to its page component, so the build path delivers no image — is still open as
-- `bugs_open/412` candidate 1 (hoist the wiring to deploy time). This file only rescues the nine
-- images already paid for. **If imagery is generated for these pages again, it will orphan again.**
--
-- ⚠ IT REGENERATES NO COPY, and that matters: the owner handed copy quality back to
-- `copy_quality_two_stage` on 2026-08-30, so this lane must not reopen it. Setting content_data and
-- re-rendering with `spec.reason='template_changed'` routes to `rerender_sections` (no LLM). The
-- fan-out is filed separately after this, deliberately, so the two halves can be verified apart.
--
-- EVIDENCE, re-measured the morning this ran rather than carried from the 08-30 note (the whole
-- reason §9 exists is that a delivery figure was quoted while its pipeline was still draining):
--   * all 8 target images resolve HTTP 200, 62,227-137,925 B [MEASURED 2026-08-31]
--   * an invented path in the same family 404s, so the check can fail and did not
--   * all 9 hero components are image-capable (migration 649 fixed the last two)
--   * 8 of 9 carry no hero_url, and several pages have re-rendered since 08-26 and still do not,
--     which is what says nothing auto-wires it
--
-- Rollback: 664_..._ROLLBACK.sql

BEGIN;

UPDATE page_components pc
   SET content_data = jsonb_set(COALESCE(pc.content_data, '{}'::jsonb), '{hero_url}',
                                to_jsonb('/assets/images/content-hero-' || p.name || '.jpg')),
       updated_at = now()
  FROM pages p, sites s, content_components cc
 WHERE p.id = pc.page_id
   AND s.id = p.site_id
   AND cc.id = pc.component_id
   AND s.domain = 'finetuning.uk'
   AND cc.name LIKE '%hero%'
   AND cc.html_template LIKE '%hero_url%'                      -- refuse a component that cannot render it (412 §7)
   AND p.name IN ('about','approach','case-studies','contact','model-approach-selector',
                  'services','tool-ai-readiness-checker','use-cases')
   AND COALESCE(NULLIF(pc.content_data->>'hero_url',''), NULL) IS NULL;  -- never overwrite an existing choice

DO $$
DECLARE n_wired int; n_missing int; n_bad_shape int; n_other_sites int;
BEGIN
    SELECT count(*) INTO n_wired
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
      JOIN content_components cc ON cc.id=pc.component_id
     WHERE s.domain='finetuning.uk' AND cc.name LIKE '%hero%'
       AND p.name IN ('about','approach','careers','case-studies','contact',
                      'model-approach-selector','services','tool-ai-readiness-checker','use-cases')
       AND NULLIF(pc.content_data->>'hero_url','') IS NOT NULL;
    IF n_wired <> 9 THEN
        RAISE EXCEPTION '664: % of 9 hero components carry a hero_url, want 9 (8 wired here + careers already)', n_wired;
    END IF;

    -- Every value must match the deterministic path, or a typo has pointed a page at a 404.
    SELECT count(*) INTO n_bad_shape
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
      JOIN content_components cc ON cc.id=pc.component_id
     WHERE s.domain='finetuning.uk' AND cc.name LIKE '%hero%'
       AND NULLIF(pc.content_data->>'hero_url','') IS NOT NULL
       AND pc.content_data->>'hero_url' <> '/assets/images/content-hero-' || p.name || '.jpg';
    IF n_bad_shape <> 0 THEN
        RAISE EXCEPTION '664: % hero_url value(s) do not match /assets/images/content-hero-<page>.jpg', n_bad_shape;
    END IF;

    -- Nothing outside this site may have been touched.
    SELECT count(*) INTO n_other_sites
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain <> 'finetuning.uk' AND pc.updated_at > now() - interval '1 minute';
    IF n_other_sites <> 0 THEN
        RAISE EXCEPTION '664: % component(s) on OTHER sites were modified — the join was too loose', n_other_sites;
    END IF;

    SELECT count(*) INTO n_missing
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
      JOIN content_components cc ON cc.id=pc.component_id
     WHERE s.domain='finetuning.uk' AND cc.name LIKE '%hero%'
       AND cc.html_template NOT LIKE '%hero_url%'
       AND NULLIF(pc.content_data->>'hero_url','') IS NOT NULL;
    IF n_missing <> 0 THEN
        RAISE EXCEPTION '664: % image-INCAPABLE component(s) were given a hero_url — that is an orphan by construction', n_missing;
    END IF;

    RAISE NOTICE '664 OK: 9 of 9 hero components carry a deterministic hero_url; no other site touched; no incapable component wired';
END $$;

COMMIT;
