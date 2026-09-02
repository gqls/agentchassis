-- 675_d4_governor_admits_function_ROLLBACK.sql — drops the canonical predicate + view.
-- Refuses while ANY live agent config references governor_admits (674 applied, or the
-- honour_spend_governor flags present — the Go call sites emit governor_admits() only when
-- a step config carries the flag, so the config census covers the binary too). Roll 674
-- back first; then this is safe at any governor state (the function is pure read).

BEGIN;

DO $$
DECLARE n int;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits') THEN
    RAISE EXCEPTION '675 ROLLBACK REFUSED: governor_admits() does not exist — 675 not applied.';
  END IF;
  SELECT count(*) INTO n FROM agent_definitions
  WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND (default_config::text LIKE '%governor_admits%'
         OR default_config::text LIKE '%honour_spend_governor%');
  IF n <> 0 THEN
    RAISE EXCEPTION '675 ROLLBACK REFUSED: % live agent config(s) still reference the predicate/flag — roll 674 back first.', n;
  END IF;
END $$;

DROP VIEW governor_withheld_now;
DROP FUNCTION governor_admits(text);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits') THEN
    RAISE EXCEPTION '675 ROLLBACK VERIFY: function survived';
  END IF;
  RAISE NOTICE '675 ROLLBACK OK: predicate + view removed; stage-A tables untouched.';
END $$;

COMMIT;
