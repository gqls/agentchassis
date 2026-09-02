-- gamedesign.uk — HOLD the tool-suggester's plants (2026-09-02 ~21:40Z).
-- The improvement loop's design-discovery-agent ran evaluate_tools; tool-suggester wrote a `tools`
-- spec and filed 8 add_tool items at 20:02Z — four DEPLOY the sibling gamesdesign.co.uk's existing
-- library tools by name (combat-balance-comparator, economy-flow-modeller, xp-curve-designer,
-- damage-formula-designer), two generate new ones, two more queued (stat-budget-allocator,
-- loot-table-balancer — also the sibling's). tool-deployer then planted 12 `planned` pages +
-- companion guides + nav_drift + "add tool reference to index" rewrites. The brief (v1 AND v2)
-- says the site "does not publish calculators, simulators, tool pages or a guide library";
-- positioning (GD2) says tools live on the sibling. The brief-fidelity auditor recorded the
-- violation at 20:04 in record mode and dispatched nothing. bugs_open/447 (to file).
-- REVERSIBLE: pages archived (never built, never deployed, never linked), items cancelled with a
-- reason; the generated components stay in the library for the owner to reinstate if wanted.
\set ON_ERROR_STOP on
BEGIN;

UPDATE site_work_items
   SET status='cancelled',
       result = COALESCE(result,'{}'::jsonb) || '{"cancelled_by":"gamedesign_uk_rebuild lane","cancelled_at":"2026-09-02T21:40:00Z","reason":"tool-suggester plant contradicts the brief (no tool pages) and positioning (tools live on gamesdesign.co.uk); four are the sibling tools by name. bugs_open/447"}'::jsonb
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414'
   AND status IN ('triaged','detected','claimed')
   AND created_at > '2026-09-02 20:02'
   AND (created_by IN ('tool-suggester','tool-deployer') OR source IN ('tool-suggester','tool-deployer'));

UPDATE pages SET status='archived', updated_at=now()
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414'
   AND name LIKE 'tool-%' AND build_status='planned' AND deployed_at IS NULL
   AND created_at > '2026-09-02 20:07';

COMMIT;

SELECT 'cancelled items' AS what, count(*) FROM site_work_items WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND status='cancelled' AND result->>'cancelled_at'='2026-09-02T21:40:00Z'
UNION ALL SELECT 'archived tool pages', count(*) FROM pages WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND name LIKE 'tool-%' AND status='archived'
UNION ALL SELECT 'still-open tool items', count(*) FROM site_work_items WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND status IN ('triaged','claimed','detected') AND (created_by IN ('tool-suggester','tool-deployer') OR source IN ('tool-suggester','tool-deployer'));
