-- 559_case_studies_grid_optional_scroll_snap_carousel_ROLLBACK.sql
--
-- Restores the byte-exact pre-559 `case-studies-grid` template from the
-- migration_backups row 559 wrote, and clears the opt-in flag from the two
-- ai-agent-orchestration.com placements.
--
-- ⚠ ORDER MATTERS AND THE FLAG MUST GO TOO. Restoring the template alone leaves
-- `carousel_enabled: true` sitting in `content_data` on two placements, where it is
-- inert but wrong — and if 559 is ever re-applied it would silently switch those
-- pages back on without anyone choosing it. Both halves, one transaction.
--
-- Templates only re-render on demand: placements keep the carousel html until they
-- are re-rendered, so a rollback is NOT complete until those two pages re-render
-- (page-scoped `template_changed`, RUNBOOK R8).
--
-- ⚠ NOTHING ON THE OTHER TWO SITES CHANGES, in either direction. finetuning.uk and
-- leopardessconsulting.co.uk never had the flag set, so they render identically
-- before 559, after 559, and after this rollback. If one of them looks different,
-- this file is not the cause and rolling it back will not help.

BEGIN;

UPDATE content_components cc
SET html_template = (b.old_value->>'html_template'),
    updated_at = now()
FROM migration_backups b
WHERE b.migration_name = '559_case_studies_grid_optional_scroll_snap_carousel'
  AND b.target_table = 'content_components'
  AND b.target_id = cc.id::text;

UPDATE page_components pc
SET content_data = pc.content_data - 'carousel_enabled',
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND pc.component_id = '3f946437-1dc7-4164-987d-620933589076'
  AND pc.content_data ? 'carousel_enabled';

DO $$
DECLARE restored int; flagged int;
BEGIN
  SELECT count(*) INTO restored FROM content_components cc
   JOIN migration_backups b ON b.target_id = cc.id::text
    AND b.migration_name='559_case_studies_grid_optional_scroll_snap_carousel'
   WHERE cc.html_template = (b.old_value->>'html_template');
  IF restored <> 1 THEN
    RAISE EXCEPTION 'rollback 559: template not byte-identical to its backup (found %)', restored;
  END IF;

  SELECT count(*) INTO flagged FROM page_components
   WHERE component_id='3f946437-1dc7-4164-987d-620933589076'
     AND content_data ? 'carousel_enabled';
  IF flagged <> 0 THEN
    RAISE EXCEPTION 'rollback 559: % placement(s) still carry carousel_enabled', flagged;
  END IF;

  RAISE NOTICE 'rollback 559 OK: template restored byte-exact and the opt-in cleared. Re-render the two pages to propagate.';
END $$;

COMMIT;
