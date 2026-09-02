-- 722 ROLLBACK — restore the previous sites.settings default
--
-- Written because the council's debug_historian seat noted (round 1, low) that the
-- house style for sql_for_agents is a separate, checked rollback file, and that a
-- future operator reverting a change should run one rather than reconstruct it from
-- a plan document.
--
-- WHAT REVERTING DOES AND DOES NOT DO. It restores the default only. Any site
-- ALREADY CREATED while 722 was in force keeps `growth_posture='hold'` on its own
-- row — a column default is not retroactive in either direction. That is correct:
-- those sites were deliberately born held, and silently opening them would dispatch
-- growth into sites nobody has reviewed, which is the failure 722 exists to prevent.
-- To open one, release it the normal way (the recipe is stamped on its held rows).
--
-- The query at the foot lists exactly which sites that is, so reverting is an
-- informed act rather than a blind one.

BEGIN;

ALTER TABLE sites
  ALTER COLUMN settings
  SET DEFAULT '{}'::jsonb;

DO $$
DECLARE
  v_default text;
  v_born_held bigint;
BEGIN
  SELECT column_default INTO v_default
    FROM information_schema.columns
   WHERE table_name = 'sites' AND column_name = 'settings';

  IF v_default IS NULL OR v_default NOT LIKE '{}%' THEN
    RAISE EXCEPTION '722 ROLLBACK: default is % — expected the empty object', COALESCE(v_default, '(null)');
  END IF;

  SELECT count(*) INTO v_born_held
    FROM sites
   WHERE settings->'maintenance_profile'->>'growth_posture' = 'hold';

  RAISE NOTICE '722 ROLLBACK OK: default restored. % site(s) still hold growth on their own row '
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
