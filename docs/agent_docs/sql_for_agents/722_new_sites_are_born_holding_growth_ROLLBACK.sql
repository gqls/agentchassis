-- 722 ROLLBACK — remove the born-holding trigger
--
-- Written because the council's debug_historian seat noted (round 1, low) that the
-- house style for sql_for_agents is a separate, checked rollback file, and that a
-- future operator reverting a change should run one rather than reconstruct it from
-- a plan document.
--
-- WHAT REVERTING DOES AND DOES NOT DO. It removes the trigger only. Any site
-- ALREADY CREATED while 722 was in force keeps `growth_posture='hold'` on its own
-- row — a column default is not retroactive in either direction. That is correct:
-- those sites were deliberately born held, and silently opening them would dispatch
-- growth into sites nobody has reviewed, which is the failure 722 exists to prevent.
-- To open one, release it the normal way (the recipe is stamped on its held rows).
--
-- The query at the foot lists exactly which sites that is, so reverting is an
-- informed act rather than a blind one.

BEGIN;

DROP TRIGGER IF EXISTS trg_sites_born_holding_growth ON sites;
DROP FUNCTION IF EXISTS sites_born_holding_growth();

DO $$
DECLARE
  v_born_held bigint;
BEGIN
  IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_sites_born_holding_growth') THEN
    RAISE EXCEPTION '722 ROLLBACK: the trigger is still present';
  END IF;

  SELECT count(*) INTO v_born_held
    FROM sites
   WHERE settings->'maintenance_profile'->>'growth_posture' = 'hold';

  RAISE NOTICE '722 ROLLBACK OK: trigger removed. % site(s) still hold growth on their own row '
    'and are UNAFFECTED by this rollback — release each one deliberately if that is what you want.',
    v_born_held;
END $$;

COMMIT;

-- Which sites still hold after the rollback, and when they were born:
--   SELECT domain, created_at::date,
--          settings->'maintenance_profile'->>'growth_posture' AS posture
--     FROM sites
--    WHERE settings->'maintenance_profile'->>'growth_posture' IS NOT NULL
--    ORDER BY created_at;
