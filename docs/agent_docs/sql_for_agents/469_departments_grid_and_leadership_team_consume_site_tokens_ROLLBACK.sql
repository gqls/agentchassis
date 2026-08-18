-- 469_departments_grid_and_leadership_team_consume_site_tokens_ROLLBACK.sql
--
-- Restores the BYTE-EXACT pre-469 html_template for departments-grid and
-- leadership-team from the migration_backups rows 469 wrote. Not a reverse-regex,
-- so it cannot drift from what was actually replaced.
--
-- ⚠ THE STATE THIS RETURNS TO IS A MEASURED 24-FAILURE STATE on
-- ai-agent-orchestration.com: #E6EDF3 on #FFFFFF at 1.18:1 (member/department
-- headings) and #E6EDF3 on #F8F9FA at 1.12:1 (section headings and the stray
-- intro paragraph), across index/departments-grid, about/departments-grid and
-- about/leadership-team. Roll back only to isolate a worse regression elsewhere,
-- and say which.
--
-- ⚠ IT ALSO RE-BREAKS NOTHING ON THE TWO LIGHT SITES, because nothing there was
-- broken: finetuning.uk and leopardessconsulting.co.uk pass every element of this
-- block both before and after 469. Rolling back to "fix" a light site is therefore
-- always the wrong diagnosis — measure first.
--
-- Templates only. Placements already re-rendered keep the tokenised html until
-- re-rendered again, so a rollback is NOT complete until those pages re-render.

BEGIN;

UPDATE content_components cc
SET html_template = (b.old_value->>'html_template'),
    updated_at = now()
FROM migration_backups b
WHERE b.migration_name = '469_departments_grid_and_leadership_team_consume_site_tokens'
  AND b.target_table = 'content_components'
  AND b.target_id = cc.id::text;

DO $$
DECLARE
  restored  int;
  tokens_left int;
BEGIN
  SELECT count(*) INTO restored FROM content_components cc
   JOIN migration_backups b
     ON b.target_id = cc.id::text
    AND b.migration_name = '469_departments_grid_and_leadership_team_consume_site_tokens'
   WHERE cc.html_template = (b.old_value->>'html_template');
  IF restored <> 2 THEN
    RAISE EXCEPTION 'rollback 469: expected 2 rows byte-identical to their backup, found %', restored;
  END IF;

  -- No tokenised declaration may survive in the two rows.
  SELECT count(*) INTO tokens_left FROM content_components
   WHERE id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b','c5af72e8-73ff-4dfe-bf88-54f7fa3978e1')
     AND html_template ~ '(background|color):\s*var\(--color-';
  IF tokens_left <> 0 THEN
    RAISE EXCEPTION 'rollback 469: % row(s) still carry a tokenised declaration', tokens_left;
  END IF;

  RAISE NOTICE 'rollback 469 OK: both templates restored byte-exact. THE 24 FAILURES ARE BACK once the pages re-render.';
END $$;

COMMIT;
