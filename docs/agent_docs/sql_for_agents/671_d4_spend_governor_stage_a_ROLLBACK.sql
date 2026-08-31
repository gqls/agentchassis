-- 671_d4_spend_governor_stage_a_ROLLBACK.sql — removes stage A entirely.
-- Guarded: refuses while the governor is switched ON (a live governor must be disabled
-- deliberately first — UPDATE governor_config SET enabled=false — never yanked by rollback).
-- Stage B (the Go claim-step check) fails OPEN on these objects being absent only if it was
-- built that way; if stage B has shipped, roll IT back first.

BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public'
                   AND table_name='governor_config') THEN
    RAISE EXCEPTION '671 ROLLBACK REFUSED: governor_config does not exist — 671 is not applied.';
  END IF;
  IF EXISTS (SELECT 1 FROM governor_config WHERE id=1 AND enabled) THEN
    RAISE EXCEPTION '671 ROLLBACK REFUSED: governor is ENABLED. Disable it deliberately first.';
  END IF;
END $$;

DELETE FROM scheduled_tasks WHERE name = 'spend-governor-state';
DROP VIEW governor_spend_mtd;
DROP TABLE governor_state;
DROP TABLE governor_config;
DROP TABLE governor_work_class_map;
DROP TABLE governor_model_prices;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM information_schema.tables WHERE table_schema='public'
    AND table_name LIKE 'governor\_%';
  IF n <> 0 THEN RAISE EXCEPTION '671 ROLLBACK VERIFY: % governor_ tables remain', n; END IF;
  PERFORM 1 FROM scheduled_tasks WHERE name='spend-governor-state';
  IF FOUND THEN RAISE EXCEPTION '671 ROLLBACK VERIFY: task row remains'; END IF;
  RAISE NOTICE '671 ROLLBACK OK: stage A removed entirely.';
END $$;

COMMIT;
