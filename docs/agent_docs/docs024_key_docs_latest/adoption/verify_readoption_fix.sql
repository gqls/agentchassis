-- verify_readoption_fix.sql
-- Staged checks for the gamesdesign re-adoption
-- (correlation 01534a57-ca61-4d1f-9337-4a663b85de16, 2026-06-05 08:18Z).
--
-- The convergence fix engages ONLY if BOTH:
--   (1) the build pass sees NO current plan (adoption_locked first-plan branch), AND
--   (2) migration (a) load_existing_pages carry-fields + v3 (b) are deployed.
-- Site id changes/reuses across runs; everything resolves via domain.

-- ===== GATE — run during the crawl/LLM window, BEFORE the build step ==========
-- (G1) Is the OLD plan still current? Adoption upserts the site (same id) and
--      does not touch site_plans, so 77d88a60 likely remains is_current. If a
--      row returns, load_existing_pages computes adoption_locked=FALSE and the
--      reconcile (and the whole fix) will NOT engage.
SELECT id, is_current, created_at
FROM site_plans
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND is_current = true;

-- (G2) ONLY if (a)+(b) are deployed: retire the current plan to make this a
--      true first pass. DO NOT run this if (b) is not deployed — clearing the
--      plan without the carry fix lets the union clobber adopted sections.
-- UPDATE site_plans SET is_current = false
-- WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--   AND is_current = true;

-- ===== DURING BUILD — planner pod logs =======================================
-- Confirm reconcile actually ran (not a no-op):
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 \
--     | grep 'reconciled with adoption-locked pages'
-- Expect unioned_in > 0 (adopted pages unioned); dropped_collision > 0 if the
-- LLM proposed a bare sibling. All-zero == no-op (G1 not met, or existing_pages
-- not reaching validate). Dedup line, if a sibling was proposed:
--   ... | grep 'dropped page duplicating an adopted item topic'

-- ===== POST-BUILD — after sync ===============================================
-- (P1) CLOBBER CHECK: did adopted pages keep their sections after sync?
--      Empty sections here on pages that should have content == the carry fix
--      (a)+(b) is not in effect.
SELECT name, page_type, build_status, sections
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND (name LIKE 'guide-%' OR page_type = 'guide')
ORDER BY name;

-- (P2) SIBLING + PLAN CHECK: guides present in the current plan, no bare names.
SELECT spp.name, spp.role
FROM site_plan_pages spp
JOIN site_plans sp ON sp.id = spp.plan_id AND sp.is_current = true
WHERE sp.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND (spp.name LIKE 'guide-%'
       OR spp.name IN ('economy-basics','fairness-in-rng','p2p-architecture','rng-design','skinner-box'))
ORDER BY spp.name;
-- Expect: guide-* rows present; NO bare names.

-- (P3) Hub check (separate planner gap — may still be empty even with the fix):
SELECT name, page_type, sections
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND name IN ('index','tools-index','guides-index','games-index')
ORDER BY name;
