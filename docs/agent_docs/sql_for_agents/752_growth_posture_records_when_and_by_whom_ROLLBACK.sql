-- 752 ROLLBACK — restore migration 722's trigger function body (posture only, no record)
--
-- Reverting removes the set_at / set_by / reason STAMPING from future inserts. Rows already
-- stamped keep their record — those keys are inert data that nothing refuses on, and the
-- growth-posture-hold-check reads them if present and copes if absent. The trigger itself
-- is untouched by both 752 and this file; 722's rollback is the one that drops it.

BEGIN;

CREATE OR REPLACE FUNCTION sites_born_holding_growth() RETURNS trigger AS $fn$
BEGIN
  IF NEW.settings IS NULL
     OR NEW.settings->'maintenance_profile'->>'growth_posture' IS NULL THEN
    NEW.settings := COALESCE(NEW.settings, '{}'::jsonb);
    IF NEW.settings->'maintenance_profile' IS NULL THEN
      NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile}', '{}'::jsonb, true);
    END IF;
    NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile,growth_posture}', '"hold"'::jsonb, true);
  END IF;
  RETURN NEW;
END;
$fn$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'sites_born_holding_growth'
                AND prosrc LIKE '%growth_posture_set_at%') THEN
    RAISE EXCEPTION '752 ROLLBACK: the function still stamps growth_posture_set_at';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_sites_born_holding_growth') THEN
    RAISE EXCEPTION '752 ROLLBACK: the 722 trigger is missing — this rollback expected to leave it in place';
  END IF;
  RAISE NOTICE '752 ROLLBACK OK: 722 body restored; the trigger still stamps posture only';
END $$;

COMMIT;
