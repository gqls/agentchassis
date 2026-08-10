-- FILE: docs/agent_docs/sql_for_agents/373_bugfix_214_canonicalise_orphaned_imagery_scope_refs_ROLLBACK.sql
--
-- Reverts 373. Restores each repaired row's PRE-REPAIR scope_ref.
--
-- The tuples below are the state measured 2026-08-10, BEFORE the forward
-- script ran. Reconcile them against the forward script's own RETURNING output
-- before applying — if a concurrent replan rewrote a plan in between, the ids
-- will differ and these UPDATEs will simply match nothing rather than damage
-- anything (they are keyed on id AND the current value).
--
-- Note what a rollback restores: the ORPHANED state, in which these rows name
-- a page no consumer can resolve and their generated assets are referenced by
-- nothing. That is the correct thing to restore if the forward script proves
-- wrong, but it is not a neutral state — see bugs_open/214.

\set ON_ERROR_STOP on

BEGIN;

-- gamesdesign.co.uk — about-index -> about (1 page hero + 4 section icons)
UPDATE site_plan_imagery SET scope_ref = 'about'
 WHERE scope = 'page' AND scope_ref = 'about-index' AND key = 'hero_about'
   AND plan_id = 'c96b501c-ca02-4a2d-bca6-7b963acbf1ce';

UPDATE site_plan_imagery SET scope_ref = 'about:2'
 WHERE scope = 'section' AND scope_ref = 'about-index:2'
   AND key IN ('icon_no_ads', 'icon_math_first', 'icon_browser_based', 'icon_practitioner')
   AND plan_id = 'c96b501c-ca02-4a2d-bca6-7b963acbf1ce';

-- gamesdesign.co.uk — contact-index -> contact
UPDATE site_plan_imagery SET scope_ref = 'contact'
 WHERE scope = 'page' AND scope_ref = 'contact-index' AND key = 'hero_contact'
   AND plan_id = 'c96b501c-ca02-4a2d-bca6-7b963acbf1ce';

-- fundamentallyai.com — news-index -> news
UPDATE site_plan_imagery spi SET scope_ref = 'news'
  FROM site_plans sp, sites si
 WHERE sp.id = spi.plan_id AND sp.is_current AND si.id = sp.site_id
   AND si.domain = 'fundamentallyai.com'
   AND spi.scope = 'page' AND spi.scope_ref = 'news-index' AND spi.key = 'hero_news';

-- mortgagecalculator.co.uk — about-index -> about, contact-index -> contact
UPDATE site_plan_imagery spi SET scope_ref = 'about'
  FROM site_plans sp, sites si
 WHERE sp.id = spi.plan_id AND sp.is_current AND si.id = sp.site_id
   AND si.domain = 'mortgagecalculator.co.uk'
   AND spi.scope = 'page' AND spi.scope_ref = 'about-index' AND spi.key = 'hero_about';

UPDATE site_plan_imagery spi SET scope_ref = 'contact'
  FROM site_plans sp, sites si
 WHERE sp.id = spi.plan_id AND sp.is_current AND si.id = sp.site_id
   AND si.domain = 'mortgagecalculator.co.uk'
   AND spi.scope = 'page' AND spi.scope_ref = 'contact-index' AND spi.key = 'hero_contact';

-- Confirm the pre-repair count is back.
DO $$
DECLARE
    remaining int;
BEGIN
    SELECT count(*) INTO remaining
      FROM site_plan_imagery spi
      JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
     WHERE spi.scope IN ('page', 'section')
       AND NOT EXISTS (SELECT 1 FROM pages p
                        WHERE p.site_id = sp.site_id
                          AND p.name = split_part(spi.scope_ref, ':', 1));
    RAISE NOTICE 'bugfix 214 rollback: % unresolvable imagery rows (expected 10 if the full forward repair is undone).', remaining;
END $$;

COMMIT;
