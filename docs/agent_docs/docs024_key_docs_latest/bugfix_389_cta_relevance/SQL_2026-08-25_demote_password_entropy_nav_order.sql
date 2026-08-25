-- Owner decision 2, 2026-08-25: demote the fossil nav_order that makes password-entropy
-- the #1 CTA candidate on three sites (bugs_open/391).
--
-- WHY 900 and not 200: 200 is the sites' ordinary tool value, so it would tie with the
-- relevant tools and fall to the alphabetical tiebreak — "password-entropy" sorts ahead of
-- every "tool-*" name, so it would STILL win on ai-agent-orchestration and finetuning.
-- 900 is unambiguous and states the intent.
--
-- SCOPE: exactly the three rows carrying the March fossil. Verified before and after.
-- This does NOT retire the tool (owner decision 1) — that is a sequenced operation with a
-- measured blast radius of 91 component references + 1 footer + 3 tools.html listings.
BEGIN;

\echo '--- BEFORE ---'
SELECT s.domain, p.name, p.nav_order, p.in_header
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.name='password-entropy' ORDER BY s.domain;

UPDATE pages p
SET nav_order = 900, updated_at = now()
FROM sites s
WHERE s.id = p.site_id
  AND p.name = 'password-entropy'
  AND p.nav_order = 1
  AND s.domain IN ('ai-agent-orchestration.com','finetuning.uk','leopardessconsulting.co.uk');

\echo '--- AFTER (expect 3 rows at 900) ---'
SELECT s.domain, p.name, p.nav_order, p.in_header
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.name='password-entropy' ORDER BY s.domain;

-- Guard: abort if we did not move exactly three rows.
DO $$
DECLARE moved int;
BEGIN
  SELECT count(*) INTO moved FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE p.name='password-entropy' AND p.nav_order=900;
  IF moved <> 3 THEN
    RAISE EXCEPTION 'expected 3 rows at nav_order=900, found %', moved;
  END IF;
END $$;

\echo '--- NEW rank-1 CTA target per affected site (the point of the change) ---'
WITH cands AS (
  SELECT s.domain, p.name, COALESCE(p.nav_order,100) AS nav_order,
         row_number() OVER (PARTITION BY s.id ORDER BY COALESCE(p.nav_order,100), p.name) AS rank
  FROM pages p JOIN sites s ON s.id=p.site_id
  WHERE p.page_type IN ('tool','game') AND p.status IN ('active','deployed')
    AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') = 'planned')
)
SELECT domain, name AS new_primary_cta_target, nav_order FROM cands
WHERE rank=1 AND domain IN ('ai-agent-orchestration.com','finetuning.uk','leopardessconsulting.co.uk')
ORDER BY domain;

COMMIT;
