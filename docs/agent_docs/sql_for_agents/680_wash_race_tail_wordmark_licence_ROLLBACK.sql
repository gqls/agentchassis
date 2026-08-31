-- ROLLBACK for 680. Strips the appended clause by its own literal, restoring the row's
-- original wording. Note this RESTORES a live wordmark licence on boxingonline.com's
-- logo prompt — only run it if the wash itself is the problem.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE v_n int;
BEGIN
  UPDATE site_plan_imagery
     SET prompt = replace(prompt,
           ' — OWNER RULING 2026-08-31 (migration 680, race tail of 670): render a text-free mark with no lettering or words of any kind; any earlier wording in this prompt that permits or presupposes a wordmark is void. The brand name is set in HTML beside the logo, never rendered in the image.',
           '')
   WHERE id = 'b56182fa-cdfe-4b9a-b1c8-606ea9fa39ea'
     AND prompt LIKE '%migration 680, race tail of 670%';
  GET DIAGNOSTICS v_n = ROW_COUNT;
  IF v_n <> 1 THEN
    RAISE EXCEPTION '680 rollback: expected exactly 1 row, updated %', v_n;
  END IF;
END $$;
COMMIT;
