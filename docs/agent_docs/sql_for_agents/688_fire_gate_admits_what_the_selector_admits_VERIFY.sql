-- 688 VERIFY — the fire gate is a strict widening of the selector's admission
--
-- One-shot, read-only, rerunnable any time. Asserts the post-688 state on the live
-- rows; exits non-zero (RAISE) on any failure, prints one PASS notice otherwise.
--
-- MUTATION-PROVED 2026-09-02: run against the pre-688 text this FAILS on assertion 2
-- ('approved' missing) — the pre-apply run expecting failure is part of applying 688.
--
-- What it pins, and what it deliberately does not:
--   * parity across BOTH trigger rows (the disabled sibling is the rollback path);
--     PARITY, not a text pin — same axis as 584 VERIFY 1/7, so future gate edits that
--     update both rows stay green here too... except on 688's own clauses:
--   * 'approved' PRESENT, pipeline ABSENT, lock_except_item_ids PRESENT, attempt/retry
--     arms PRESENT. These five are bugs_open/415's three narrownesses plus the two
--     shared arms; a future edit that re-narrows the gate should fail here loudly.
--   * It does NOT assert approval_mode/depends_on/busy-skip are absent — those are
--     selector-side narrowings the gate deliberately omits (wider is safe); adding
--     them to the gate would be wrong but is a review matter, not a VERIFY matter.

DO $vfy$
DECLARE n int; q text;
BEGIN
  -- 1. row set + parity
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 2 THEN
    RAISE EXCEPTION '688 VERIFY 1: % trigger rows, expected 2 (enabled + disabled sibling)', n;
  END IF;
  SELECT count(DISTINCT md5(coalesce(pre_query,''))) INTO n
    FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF n <> 1 THEN
    RAISE EXCEPTION '688 VERIFY 1: trigger rows carry % distinct pre_query values — a by-name UPDATE missed the sibling (LANDMINES 2026-08-25)', n;
  END IF;

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';

  -- 2. narrowness 1 closed: approved admitted
  IF q NOT LIKE '%''approved''%' THEN
    RAISE EXCEPTION '688 VERIFY 2: ''approved'' missing — an approved-only backlog would not fire the trigger (bugs_open/415 narrowness 1)';
  END IF;

  -- 3. narrowness 2 closed: no pipeline filter (selector and loader have none)
  IF q LIKE '%pipeline%' THEN
    RAISE EXCEPTION '688 VERIFY 3: pipeline filter present — a non-build backlog would not fire the trigger (bugs_open/415 narrowness 2)';
  END IF;

  -- 4. narrowness 3 closed: lock-exception arm, cross-site spelling
  IF q NOT LIKE '%lock_except_item_ids%' THEN
    RAISE EXCEPTION '688 VERIFY 4: lock-exception arm missing — a lock-excepted-only backlog would not fire the trigger (bugs_open/415 narrowness 3)';
  END IF;

  -- 5. the shared arms stayed (widening, not deletion)
  IF q NOT LIKE '%attempt_count < wi.max_attempts%' THEN
    RAISE EXCEPTION '688 VERIFY 5: attempt_count arm missing — the gate widened past the selector''s shared arms';
  END IF;
  IF q NOT LIKE '%retry_after IS NULL OR wi.retry_after <= NOW()%' THEN
    RAISE EXCEPTION '688 VERIFY 5: retry_after arm missing or respelled — the gate widened past the selector''s shared arms';
  END IF;

  RAISE NOTICE '688 VERIFY PASS: gate is a strict widening of the selector admission on both rows';
END $vfy$;
