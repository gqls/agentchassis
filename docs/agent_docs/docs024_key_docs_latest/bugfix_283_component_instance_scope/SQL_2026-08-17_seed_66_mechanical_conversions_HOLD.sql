-- SQL_2026-08-17_seed_66_mechanical_conversions_HOLD.sql
--
-- ⚠ HOLD — DO NOT APPLY UNTIL THE CANARY (item 38efde3b, tool-css-unit-converter)
-- HAS COMPLETED END-TO-END: converted, rerendered, served page diffed. The whole
-- point of a canary is that this file waits for it. Lives in the lane dir, NOT in
-- sql_for_agents/, precisely so no migration runner ever sweeps it up.
--
-- WHAT IT DOES. One instance_scope_conversion work item per ELIGIBLE COMPONENT
-- ROW — the 66 mechanical rows of bugs_open/283 / RFC_034 (owner ruling
-- 2026-08-17: hybrid, LMC first, through the framework). Eligibility is DERIVED
-- at apply time by the same predicate as the corpus census, never from a pasted
-- id list that goes stale:
--
--   * active, binds by getElementById, placed on a live page   (the 91)
--   * MINUS the judged pool (function list below — the 25: script work needed,
--     the converter's gate would refuse them anyway; excluding them here just
--     avoids 25 no-op dispatches)
--   * MINUS templates already referencing {{.InstanceID}} (idempotency)
--   * MINUS rows with an open item under this dedup key (the partial unique
--     index enforces it; ON CONFLICT DO NOTHING makes re-application safe)
--
-- The per-row item_key 'instance-scope:<first 8 of id>' keeps FORKED rows
-- distinct (4 functions carry 2-4 rows each; keying by function would collapse
-- them — RFC_034 §1's trap).
--
-- Note the gate double-covers the judged exclusion: if the function list below
-- drifts (a template's script is rewritten between census and apply), the
-- converter itself refuses with action:"needs_script_scoping" and writes
-- nothing. The list is an efficiency, not a safety mechanism.
--
-- ⚠ THE ELIGIBLE COUNT IS A MOVING TARGET — that is WHY eligibility is derived here
-- rather than pasted. Dry-run 2026-08-17 evening: 69 rows (not the morning's 66) —
-- the corpus grew by three DURING THE DAY (tool-markdown-tables, tool-html-minifier,
-- tool-review-council-simulator, all newly placed on live pages). Re-run the dry-run
-- count below at apply time and expect it to have moved again; the number that must
-- NOT move is the judged-pool exclusion, which the converter's gate re-checks per row
-- regardless.
--
-- DRY-RUN the eligible count first:
--   SELECT count(DISTINCT c.id) FROM content_components c
--     JOIN page_components pc ON pc.component_id=c.id
--     JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
--   WHERE c.is_active AND c.html_template ~ 'getElementById'
--     AND c.html_template NOT LIKE '%{{.InstanceID}}%'
--     AND c.function NOT IN ( /* the list below */ );
--
-- VERIFY after applying:
--   SELECT count(*) FROM site_work_items WHERE item_type='instance_scope_conversion'
--     AND status='triaged';                      -- expect the dry-run count minus any already open
--   -- and watch the fleet convert:
--   SELECT count(*) FROM content_components WHERE is_active
--     AND html_template LIKE '%{{.InstanceID}}%';

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  priority, handler_agent, status, created_by, spec, item_key, batch_id
)
SELECT DISTINCT ON (c.id)
  s.id,
  'manual', 'build', 'instance_scope_conversion', 'low',
  'bugs_open/283 batch (RFC_034): deterministic instance-scope conversion of ' || c.function
    || ' (row ' || left(c.id::text, 8) || ') — mechanical pool; gate refuses if the script is unscoped',
  35,
  'component-template-fixer', 'triaged', '283-lane-batch-seed',
  jsonb_build_object(
    'fix_type', 'scope_component_instance',
    'component_id', c.id::text,
    'category', 'seam',
    'note', 'batch of the 66 mechanical rows, RFC_034 DECIDED; canary 38efde3b proved the pipeline'
  ),
  'instance-scope:' || left(c.id::text, 8),
  gen_random_uuid()
FROM content_components c
JOIN page_components pc ON pc.component_id = c.id
JOIN pages p  ON p.id = pc.page_id
JOIN sites s  ON s.id = p.site_id
WHERE c.is_active
  AND c.html_template ~ 'getElementById'
  AND c.html_template NOT LIKE '%{{.InstanceID}}%'
  AND c.function NOT IN (
    -- the judged pool, RFC_034 §3a (23 LMC calculators + 2 tools); the
    -- converter's gate re-checks this per row regardless
    'loans-application-tracker','loans-car-finance-calculator','loans-compare-loans',
    'loans-consolidation','loans-credit-health-check','loans-damage-checker',
    'loans-interest-rate-stress-test','loans-loan-vs-savings','loans-overpayment-calculator',
    'loans-settlement-calculator','loans-standard-calc',
    'mortgages-affordability','mortgages-bridging-loan','mortgages-equity-release',
    'mortgages-fact-finder','mortgages-fee-analyser','mortgages-investor',
    'mortgages-overpayment','mortgages-portfolio','mortgages-rate-forecaster',
    'mortgages-repayment','mortgages-simple','mortgages-stamp-duty',
    'tool-archetype-clash-calculator','tool-bayesian-ranking'
  )
ORDER BY c.id, s.id
ON CONFLICT DO NOTHING;
