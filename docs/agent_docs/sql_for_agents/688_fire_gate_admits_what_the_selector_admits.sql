-- 688 — bugs_open/415: the fire gate admits what the selector admits (gate ⊇ selector)
--
-- WHY
-- ---
-- scheduled_tasks.build-pipeline-trigger.pre_query decides WHETHER the trigger fires;
-- the selector it gates (find_dispatchable_site, post-657) decides WHICH site. The gate
-- was narrower than the selector in THREE independent ways (bugs_open/415, third
-- narrowness confirmed 2026-09-02):
--   1. wi.status = 'triaged' only          — selector admits ('triaged','approved')
--   2. wi.pipeline = 'build' present       — selector and loader have NO pipeline filter
--   3. s.locked_at IS NULL bare            — selector honours the lock-EXCEPTION arm
-- So a backlog that is approved-only, non-build-pipeline, or entirely on a lock-excepted
-- site is dispatchable by the selector and loadable by the loader, and the trigger never
-- fires to ask. No error anywhere; the damage is an absence (the 413 meter-blindness
-- class). Theoretical at today's volume ([MEASURED 2026-09-02] 102 eligible rows, all
-- triaged/build; 0 lock-exception entries) but the door is reachable — end-of-backlog is
-- exactly where the throughput work drives — and this drift class has bitten three times
-- on this seam (078→285, 413→657, this).
--
-- WHAT THIS DOES (chosen fix: bugs_open/415 candidate 1; owner not blocked, 2026-09-02)
-- --------------------------------------------------------------------------------------
-- Replaces the WHOLE pre_query value on BOTH trigger rows with a strict widening:
--   * status IN ('triaged', 'approved')
--   * pipeline filter REMOVED
--   * lock-exception arm added in the selector's own CROSS-SITE spelling
--     (⚠ NOT the per-site Go fragment — work_items_common.go:851-870 explains why the
--     two spellings must stay different; the per-site one binds a $1 that does not
--     exist here)
--   * attempt_count / retry_after arms kept (selector has both)
--
-- DELIBERATELY NOT ADDED: approval_mode, depends_on, busy-skip. Those are selector-side
-- narrowings. A gate may be WIDER than its selector — a spare fire is one cheap no-op
-- tick (the selector returns 0 rows, check_has_site ends the run, ~1.3ms gate cost
-- measured 2026-09-02) — but it must never be narrower. Do NOT "fix" the gate by adding
-- them: that re-creates exactly the drift this migration closes.
--
-- BOTH ROWS IN ONE STATEMENT (LANDMINES 2026-08-25 sibling parity): the disabled
-- build-pipeline-trigger-2 row is ruling B's rollback path; a by-name UPDATE on just the
-- enabled row desyncs it silently. ROW_COUNT is asserted = 2. The 584 daily VERIFY's
-- assertion 1/7 pins PARITY across rows, not text, so this both-rows update keeps it
-- green and no lockstep edit is owed (584_..._VERIFY.sql:24-33, read 2026-09-02).
--
-- WHOLE-VALUE REPLACEMENT, md5-preflighted (LANDMINES 2026-08-24: regexp_replace with
-- 'n' over multi-line text silently replaces nothing and still reports UPDATE 1).
-- Anchor: md5(pre_query) = 200246f7ede3e33b14be2fc064efa7da on both rows, read live
-- 2026-09-02. New value md5: 2ebd918b33b36d1b55014bbe60cc2dcb.
--
-- RERUN-SAFE: a replay finds the new text already in place, raises a NOTICE and no-ops;
-- the post-check still holds. Any OTHER text refuses — the gate drifted beneath this
-- file; re-read it, never blind-replace.
--
-- DB config is live the moment this applies. The change only makes the trigger fire
-- MORE (gate ⊇ selector), so no measurement-window coordination is owed (unlike 657);
-- the dispatch_throughput lane is pinged at commit + apply regardless.
--
-- Verify: 688_fire_gate_admits_what_the_selector_admits_VERIFY.sql (mutation-proved:
-- fails against the pre-fix text). Rollback: 688_..._ROLLBACK.sql (restores the exact
-- old value, anchor 200246f7..., both rows).

BEGIN;

DO $mig$
DECLARE
  old_md5 constant text := '200246f7ede3e33b14be2fc064efa7da';
  new_text constant text := $q$SELECT COUNT(*)::text as pending_sites
FROM sites s
WHERE EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))
      AND wi.status IN ('triaged', 'approved')
      AND wi.attempt_count < wi.max_attempts
      AND (wi.retry_after IS NULL OR wi.retry_after <= NOW())
)
HAVING COUNT(*) > 0$q$;
  n int;
  shared text;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 2 THEN
    RAISE EXCEPTION '688 preflight: % build-pipeline-trigger rows, expected exactly 2 (enabled + disabled sibling) — the row set changed beneath this migration', n;
  END IF;

  SELECT count(DISTINCT md5(coalesce(pre_query,''))) INTO n
    FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 1 THEN
    RAISE EXCEPTION '688 preflight: trigger rows disagree on pre_query (% distinct) — reconcile before widening (LANDMINES 2026-08-25, sibling parity)', n;
  END IF;

  SELECT md5(coalesce(pre_query,'')) INTO shared
    FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';
  IF shared = md5(new_text) THEN
    RAISE NOTICE '688: pre_query already carries this migration''s text — rerun is a no-op';
    RETURN;
  END IF;
  IF shared <> old_md5 THEN
    RAISE EXCEPTION '688 preflight: live pre_query md5 % is neither the 2026-09-02 anchor % nor this migration''s own text — the gate drifted; re-read it, never blind-replace', shared, old_md5;
  END IF;

  UPDATE scheduled_tasks
     SET pre_query = new_text, updated_at = NOW()
   WHERE name LIKE 'build-pipeline-trigger%';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 2 THEN
    RAISE EXCEPTION '688: UPDATE touched % rows, expected exactly 2 (both siblings — LANDMINES sibling parity)', n;
  END IF;
END $mig$;

-- Post-check, asserted so the COMMIT cannot succeed on a partial or drifted apply
-- (a block of SELECTs cannot stop a COMMIT — DO/RAISE can).
DO $chk$
DECLARE n int; q text;
BEGIN
  SELECT count(DISTINCT md5(coalesce(pre_query,''))) INTO n
    FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 1 THEN
    RAISE EXCEPTION '688 post-check: siblings desynced (% distinct pre_query values)', n;
  END IF;
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';
  IF q NOT LIKE '%''approved''%' THEN
    RAISE EXCEPTION '688 post-check: ''approved'' missing from the gate (narrowness 1 still open)';
  END IF;
  IF q LIKE '%pipeline%' THEN
    RAISE EXCEPTION '688 post-check: pipeline filter still present (narrowness 2 still open)';
  END IF;
  IF q NOT LIKE '%lock_except_item_ids%' THEN
    RAISE EXCEPTION '688 post-check: lock-exception arm missing (narrowness 3 still open)';
  END IF;
  IF q NOT LIKE '%attempt_count < wi.max_attempts%' OR q NOT LIKE '%retry_after%' THEN
    RAISE EXCEPTION '688 post-check: attempt/retry arms missing — the gate widened past the selector''s shared arms';
  END IF;
END $chk$;

COMMIT;
