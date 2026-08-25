-- 625_aiao_owned_tool_pages_contrast_ROLLBACK.sql
--
-- Restores the byte-exact pre-625 `rendered_html` for the three owned tool page
-- components AND the pre-625 `html_template` for the two aiao tool templates, from
-- the migration_backups rows 625 wrote.
--
-- ⚠ THE STATE THIS RETURNS TO IS 9 MEASURED CONTRAST FAILURES:
--   /tools/agent-complexity-estimator.html  6  (h2 1.04, four .ace-legend at 1.00, .estimate-btn 1.04)
--   /tools/password-entropy.html            2  (#666666 on #0D1117 = 3.30, needs 4.5)
--   /tools/tool-llm-cost-calculator.html    1  (.calc-btn 1.04)
-- Roll back only to isolate a worse regression elsewhere, and say which.
--
-- ⚠ THE DATABASE IS NOT THE SERVED PAGE. 625's effect reached the live site only via a
-- second step — an ASSEMBLE-mode re-render through `refresh_owned_page_chrome.sh`,
-- which flips the page to 'generic', re-assembles the STORED html with fresh chrome,
-- and flips it back under an EXIT trap. **This rollback restores the database only.**
-- The old CSS will not return to the live pages until that same script is run again
-- (any marker will do; the served check is what tells you it landed). Until then the
-- site and the database disagree, and the SITE is the one visitors see.
--
-- ⚠ DO NOT "SIMPLIFY" THAT STEP BY LEAVING THE PAGES ON 'generic'. They hold their
-- whole tool — calculator markup and `<script>` — in one component; the composition
-- loop commits freshly-written HTML to the deploying repo one step BEFORE
-- `save_page_sections` refuses, so a generic build ships prose over the calculator.
-- Verify `rebuild_policy='owned'` on all three after ANY run of that script.

BEGIN;

UPDATE page_components pc
SET rendered_html = (b.old_value->>'rendered_html'), updated_at = now()
FROM migration_backups b
WHERE b.migration_name='625_aiao_owned_tool_pages_contrast'
  AND b.target_table='page_components' AND b.target_id = pc.id::text;

UPDATE content_components cc
SET html_template = (b.old_value->>'html_template'), updated_at = now()
FROM migration_backups b
WHERE b.migration_name='625_aiao_owned_tool_pages_contrast'
  AND b.target_table='content_components' AND b.target_id = cc.id::text;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components pc JOIN migration_backups b
    ON b.target_id=pc.id::text AND b.migration_name='625_aiao_owned_tool_pages_contrast'
   WHERE b.target_table='page_components' AND pc.rendered_html = (b.old_value->>'rendered_html');
  IF n <> 3 THEN RAISE EXCEPTION 'rollback 625: % of 3 artefacts byte-identical to backup', n; END IF;

  SELECT count(*) INTO n FROM content_components cc JOIN migration_backups b
    ON b.target_id=cc.id::text AND b.migration_name='625_aiao_owned_tool_pages_contrast'
   WHERE b.target_table='content_components' AND cc.html_template = (b.old_value->>'html_template');
  IF n <> 2 THEN RAISE EXCEPTION 'rollback 625: % of 2 templates byte-identical to backup', n; END IF;

  SELECT count(*) INTO n FROM pages
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
     AND name IN ('agent-complexity-estimator','password-entropy','tool-llm-cost-calculator')
     AND rebuild_policy = 'owned';
  IF n <> 3 THEN RAISE EXCEPTION 'rollback 625: only % of 3 tool pages are still rebuild_policy=owned — restore that FIRST', n; END IF;

  RAISE NOTICE 'rollback 625 OK: 3 artefacts + 2 templates restored byte-exact, all 3 pages still owned. THE LIVE PAGES ARE UNCHANGED until refresh_owned_page_chrome.sh is re-run.';
END $$;

COMMIT;
