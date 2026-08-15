-- Deliver the head-data fix: assemble-only rerender for the six pages whose
-- pages.title / pages.meta_description changed.
--
-- spec deliberately carries NO "reason": check_rerender_mode's ELSE branch is
-- assemble-stored-sections + current chrome + deploy. section_data_resolved
-- would risk RUNBOOK TRAP 1b (escalation to the LLM writer) on the D-004
-- hand-authored guide page.
-- page_id goes in the COLUMN as well as the spec, and filename is included --
-- LANDMINES.md:267, where omitting either burns 3 attempts looking like a flaky
-- handler rather than a malformed item.
\set ON_ERROR_STOP on
BEGIN;

INSERT INTO site_work_items (
  site_id, page_id, source, item_type, severity, summary, spec,
  affected_url, handler_agent, status, created_by, priority, item_key, triaged_at
)
SELECT
  p.site_id, p.id,
  'fleet_copy_quality', 'page_rerender', 'low',
  'Re-assemble ' || p.name || ' — head data changed (title/meta description)',
  jsonb_build_object(
    'domain',    s.domain,
    'page_id',   p.id::text,
    'page_name', p.name,
    'filename',  ltrim(p.url, '/')
  ),
  p.url, 'page-rerender', 'triaged', 'claude-ideauk-headmeta-20260815', 70,
  'page_rerender_headmeta_' || p.id::text, NOW()
FROM sites s JOIN pages p ON p.site_id = s.id
WHERE (s.domain, p.name) IN (
  ('idea.uk','index'),
  ('idea.uk','tool-funding-fit'),
  ('idea.uk','guide-testing-it'),
  ('finetuning.uk','our-position-on-ai'),
  ('leopardessconsulting.co.uk','use-cases'),
  ('mortgagecalculator.co.uk','guide-first-time-buyer')
);

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by = 'claude-ideauk-headmeta-20260815' AND status = 'triaged';
  RAISE NOTICE 'dispatched: % items', n;
  IF n <> 6 THEN RAISE EXCEPTION 'expected 6 rerender items, got %', n; END IF;
END $$;

COMMIT;
