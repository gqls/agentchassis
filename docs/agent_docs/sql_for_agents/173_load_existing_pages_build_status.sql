-- 173_load_existing_pages_build_status.sql
--
-- bugs_open/001 — re-planning a site silently discards its built pages'
-- composition. Half of that fix; the other half is in Go.
--
-- WHY. reconcilePlanWithRealised (v3_site_actions.go) is the only thing that
-- carries a realised page's composition into a re-plan. It used to force-preserve
-- ONLY pages with adoption_locked = true. With the minimal 054 variant above,
-- adoption_locked is true only while the site has NO current plan — i.e. uniquely
-- the first plan after adoption. Every subsequent re-plan therefore saw an empty
-- preserved set, the convergence no-op'd, and the plan became the raw LLM
-- proposal: a built page re-proposed under the same name was silently
-- re-composed, and a built page the LLM omitted was dropped outright. Proven on
-- idea.uk 2026-07-14 (plan 32be2797 -> ff03bdef): four built pages regressed,
-- two of which re-rendered and re-deployed the regressed artefact.
--
-- The Go side now preserves pages that are adoption_locked OR built. "Built" is
-- build_status = 'deployed' — the same test decideEmit uses in
-- reconcile_site_plan_action.go. This migration surfaces that column so the Go
-- side can see it.
--
-- ORDERING IS FREE. This migration is inert on its own (the current chassis
-- reads only adoption_locked and ignores extra columns), and the Go change
-- degrades to the old adoption-locked-only behaviour when build_status is absent
-- (realisedPageIsBuilt returns false on a missing column). So DB-first and
-- image-first are both safe; the fix is complete once both have landed.
--
-- ALSO FIXED HERE: the live query selected `p.sections, p.meta_description,
-- p.nav_order` TWICE — a duplicated tail left by the carry-fields migration
-- landing on top of 054. Harmless (the row map just overwrites) but misleading
-- to read, so it is deduplicated in passing.
--
-- NOTE the query below is written against the LIVE query as read from
-- agent_definitions on 2026-07-19, which already carries the sections /
-- meta_description / nav_order columns. It is NOT built on the 054 block in
-- 053_build_site_planner.sql, which predates them.

BEGIN;

SELECT snapshot_agent('build-site-planner',
                      'load_existing_pages surfaces build_status for the built-page preservation set (bugs_open/001)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages,config,query}',
        to_jsonb($Q$
SELECT p.name, p.page_type, p.url, p.title, p.nav_label, p.in_header, p.in_footer,
       p.sections, p.meta_description, p.nav_order, p.build_status,
       CASE
         WHEN NOT EXISTS (
             SELECT 1 FROM site_plans sp
             WHERE sp.site_id = p.site_id AND sp.is_current = true
         ) THEN true
         ELSE false
       END AS adoption_locked
FROM pages p
WHERE p.site_id = $1 AND p.status = 'active'
ORDER BY p.name
$Q$::text),
        true
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

COMMIT;

-- ---------------------------------------------------------------------------
-- Verification
-- ---------------------------------------------------------------------------
-- 1. The step query now carries build_status exactly once, and sections /
--    meta_description / nav_order exactly once each:
--    SELECT default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query'
--    FROM agent_definitions WHERE type='build-site-planner' AND is_active=true;
--
-- 2. The query still runs (it is the planner's own step query — check it against
--    a real site before trusting a re-plan):
--    SELECT p.name, p.build_status FROM pages p
--    WHERE p.site_id = '<site>' AND p.status = 'active' ORDER BY p.name;
--
-- 3. Functional, once the chassis image carrying the Go half is rolled: re-plan a
--    site with built pages and check the validate step's log line
--    "ValidateSitePlanAction: reconciled with realised pages" — snapped_sections
--    and/or unioned_in should be non-zero where before the whole convergence
--    no-op'd, and every built page's pages.sections must be unchanged afterwards.
