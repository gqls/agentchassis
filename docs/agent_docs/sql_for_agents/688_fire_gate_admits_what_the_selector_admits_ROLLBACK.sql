-- 688 ROLLBACK — restore the pre-688 fire-gate pre_query on BOTH trigger rows
--
-- Restores the exact old value (md5 200246f7ede3e33b14be2fc064efa7da, read live
-- 2026-09-02 and byte-verified against this file's literal before commit). Re-narrows
-- the gate to triaged/build/bare-lock — i.e. re-opens bugs_open/415's three
-- narrownesses; use only to back out 688 itself.
--
-- Same guards as the forward migration: both rows in one statement (LANDMINES
-- 2026-08-25 sibling parity), whole-value replacement (LANDMINES 2026-08-24
-- regexp_replace trap), rerun-safe, refuses on any third text.

BEGIN;

DO $rb$
DECLARE
  new_md5 constant text := '2ebd918b33b36d1b55014bbe60cc2dcb';  -- 688's text
  old_text constant text := $q$SELECT COUNT(*)::text as pending_sites
FROM sites s
WHERE s.locked_at IS NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status = 'triaged'
      AND wi.pipeline = 'build'
      AND wi.attempt_count < wi.max_attempts
      AND (wi.retry_after IS NULL OR wi.retry_after <= NOW())
)
HAVING COUNT(*) > 0$q$;
  n int;
  shared text;
BEGIN
  IF md5(old_text) <> '200246f7ede3e33b14be2fc064efa7da' THEN
    RAISE EXCEPTION '688 ROLLBACK: embedded old text does not hash to the anchor — this file was edited; do not apply';
  END IF;

  SELECT count(*) INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 2 THEN
    RAISE EXCEPTION '688 ROLLBACK preflight: % trigger rows, expected exactly 2', n;
  END IF;

  SELECT count(DISTINCT md5(coalesce(pre_query,''))) INTO n
    FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 1 THEN
    RAISE EXCEPTION '688 ROLLBACK preflight: trigger rows disagree on pre_query (% distinct) — reconcile first (LANDMINES sibling parity)', n;
  END IF;

  SELECT md5(coalesce(pre_query,'')) INTO shared
    FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';
  IF shared = md5(old_text) THEN
    RAISE NOTICE '688 ROLLBACK: pre_query already carries the pre-688 text — rerun is a no-op';
    RETURN;
  END IF;
  IF shared <> new_md5 THEN
    RAISE EXCEPTION '688 ROLLBACK preflight: live pre_query md5 % is not 688''s text — the gate drifted since 688; re-read it, never blind-replace', shared;
  END IF;

  UPDATE scheduled_tasks
     SET pre_query = old_text, updated_at = NOW()
   WHERE name LIKE 'build-pipeline-trigger%';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 2 THEN
    RAISE EXCEPTION '688 ROLLBACK: UPDATE touched % rows, expected exactly 2', n;
  END IF;
END $rb$;

DO $chk$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name LIKE 'build-pipeline-trigger%'
     AND md5(coalesce(pre_query,'')) = '200246f7ede3e33b14be2fc064efa7da';
  IF n <> 2 THEN
    RAISE EXCEPTION '688 ROLLBACK post-check: % of 2 rows carry the restored anchor text', n;
  END IF;
END $chk$;

COMMIT;
