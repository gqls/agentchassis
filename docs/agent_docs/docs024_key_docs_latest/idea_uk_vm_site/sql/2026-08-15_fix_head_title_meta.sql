-- Fleet 'honest' pass: the HEAD surface (pages.title / pages.meta_description).
-- Exact substring replacement only, server-side: no global regex, no a/an rule,
-- nothing that can touch a character outside the replaced span (§X.56 lesson).
-- Every statement must report UPDATE 1 -- "every rule fired" is half the check.
\set ON_ERROR_STOP on
BEGIN;

-- 1. idea.uk / index  (meta)  "pushes back honestly." -> "pushes back."
UPDATE pages p SET meta_description = replace(p.meta_description,'pushes back honestly','pushes back'), updated_at = NOW()
FROM sites s WHERE s.id = p.site_id AND s.domain = 'idea.uk' AND p.name = 'index'
  AND p.meta_description LIKE '%pushes back honestly%';

-- 2. idea.uk / tool-funding-fit (meta)  "An honest steer" -> "A steer"
UPDATE pages p SET meta_description = replace(p.meta_description,'An honest steer','A steer'), updated_at = NOW()
FROM sites s WHERE s.id = p.site_id AND s.domain = 'idea.uk' AND p.name = 'tool-funding-fit'
  AND p.meta_description LIKE '%An honest steer%';

-- 3. idea.uk / guide-testing-it (TITLE)  minimal deletion: D-004 protects this
--    page's hand-authored copy, so remove the word and invent nothing.
UPDATE pages p SET title = replace(p.title,'honest experiments','experiments'), updated_at = NOW()
FROM sites s WHERE s.id = p.site_id AND s.domain = 'idea.uk' AND p.name = 'guide-testing-it'
  AND p.title LIKE '%honest experiments%';

-- 4. finetuning.uk / our-position-on-ai (TITLE)  "Our Honest Position" -> "Our Position"
UPDATE pages p SET title = replace(p.title,'Our Honest Position','Our Position'), updated_at = NOW()
FROM sites s WHERE s.id = p.site_id AND s.domain = 'finetuning.uk' AND p.name = 'our-position-on-ai'
  AND p.title LIKE '%Our Honest Position%';

-- 5. leopardessconsulting.co.uk / use-cases (meta)  show it, do not label it
UPDATE pages p SET meta_description = replace(p.meta_description,'each honestly labelled','each labelled for what it is'), updated_at = NOW()
FROM sites s WHERE s.id = p.site_id AND s.domain = 'leopardessconsulting.co.uk' AND p.name = 'use-cases'
  AND p.meta_description LIKE '%each honestly labelled%';

-- 6a. mortgagecalculator.co.uk / guide-first-time-buyer (meta, the CACHE)
UPDATE pages p SET meta_description = replace(p.meta_description,'An honest and comprehensive guide','A comprehensive guide'), updated_at = NOW()
FROM sites s WHERE s.id = p.site_id AND s.domain = 'mortgagecalculator.co.uk' AND p.name = 'guide-first-time-buyer'
  AND p.meta_description LIKE '%An honest and comprehensive guide%';

-- 6b. ...and the SOURCE. pages is a materialised cache; site_db_actions.go:1173
--     re-upserts meta_description = EXCLUDED.meta_description unconditionally,
--     so 6a alone regresses on the next plan sync. This is the only one of the
--     six whose CURRENT plan carries the string.
UPDATE site_plan_pages spp SET meta_description = replace(spp.meta_description,'An honest and comprehensive guide','A comprehensive guide')
FROM site_plans sp JOIN sites s ON s.id = sp.site_id
WHERE spp.plan_id = sp.id AND sp.is_current AND s.domain = 'mortgagecalculator.co.uk'
  AND spp.name = 'guide-first-time-buyer'
  AND spp.meta_description LIKE '%An honest and comprehensive guide%';

-- Assertion: nothing left behind, at BOTH layers. A bare SELECT cannot stop the
-- COMMIT (ON_ERROR_STOP ignores a non-empty result) -- it has to RAISE.
DO $$
DECLARE n_pages int; n_plan int; n_total int;
BEGIN
  SELECT count(*) INTO n_pages FROM sites s JOIN pages p ON p.site_id = s.id
    WHERE s.domain NOT LIKE 'pool-%' AND (p.title ~* '\yhonest' OR p.meta_description ~* '\yhonest');
  SELECT count(*) INTO n_plan FROM sites s
    JOIN site_plans sp ON sp.site_id = s.id AND sp.is_current
    JOIN site_plan_pages spp ON spp.plan_id = sp.id
    WHERE s.domain NOT LIKE 'pool-%' AND (spp.title ~* '\yhonest' OR spp.meta_description ~* '\yhonest');
  -- control: the denominator must be non-zero, or the zeros above are blind
  SELECT count(*) INTO n_total FROM sites s JOIN pages p ON p.site_id = s.id WHERE s.domain NOT LIKE 'pool-%';
  RAISE NOTICE 'remaining: pages=%, current-plan=% (control: % page rows scanned)', n_pages, n_plan, n_total;
  IF n_total = 0 THEN RAISE EXCEPTION 'control failed: scanned 0 page rows, the zero is blind'; END IF;
  IF n_pages <> 0 OR n_plan <> 0 THEN
    RAISE EXCEPTION 'left behind: % in pages, % in current plans', n_pages, n_plan;
  END IF;
END $$;

COMMIT;
