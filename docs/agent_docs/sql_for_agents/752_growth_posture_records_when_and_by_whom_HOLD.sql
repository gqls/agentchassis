-- 752 — a growth hold records WHEN it was set and BY WHOM (improvement_loop lane, 2026-09-03)
--
-- WHY. Migration 722 makes a new site born holding growth (owner ruling 2026-09-02), and
-- its own header names the residual it leaves: "nothing reports 'site held for longer than
-- N days', so a site nobody releases stops growing silently". The daily CronJob
-- growth-posture-hold-check (deployments/kustomize/services/growth-posture-hold-check/) is
-- that report — and a report of "longer than N days" needs a clock.
--
-- [MEASURED 2026-09-03 17:0xZ] the hold record has none. 722's trigger stamps ONLY
-- `growth_posture`. Of the two hand-held sites, gamedesign.uk carries
-- `growth_posture_reason` + `growth_posture_set_by` (with the date inside the set_by
-- STRING) and apis.uk carries `growth_posture` alone. Two lanes, two shapes, zero
-- timestamps. `sites.updated_at` cannot stand in for one: ensure_site_record UPSERTs every
-- site on every improvement-loop pass (~50/day fleet-wide), so it says when the loop last
-- visited, not when the hold began.
--
-- WHAT. The trigger FUNCTION is replaced (CREATE OR REPLACE, same name; the INSERT-only
-- TRIGGER itself is untouched and not recreated). When — and only when — it stamps the
-- posture, it now stamps three more keys beside it:
--
--     growth_posture_set_at  = now()   (to_jsonb(now()): ISO-8601 with zone, the same
--                                       shape `last_audit.at` already uses on this object)
--     growth_posture_set_by  = 'trg_sites_born_holding_growth'
--     growth_posture_reason  = 'born held — migration 722 (owner ruling 2026-09-02: a
--                               new site is released by a human)'
--
-- The key names are the gamedesign.uk lane's hand-hold shape — the only shape in use that
-- records anything — so a born hold and a documented hand hold read alike; `set_at` is the
-- one addition. The hand recipe in register WDS-020 gains the same three lines.
--
-- WHAT IT DOES NOT DO — no backfill. The two hand-held rows belong to other lanes and keep
-- exactly what they wrote. A `set_at` invented here for them would be a guess wearing a
-- timestamp. The CronJob reports a hold with no set_at as "age unknown" and bounds it below
-- by the first day the check itself saw the hold (read back from its own doc_notes rows),
-- which is honest and needs no write to anybody's site.
--
-- 722's rules survive unchanged and each keeps a verify arm below:
--   * ABSENT means "nobody said"; a STATED value is kept in either direction, and a stated
--     value gets NO record from this trigger (arm 3) — the record must only accompany a
--     stamp this trigger made, or "set_by = trigger" would be a lie on a seed's own row.
--   * jsonb_set only creates the LAST path element — the parent is materialised first.
--   * INSERT-only, no UPDATE arm (arm 5 — the shape 722's arm 6 found discriminating).
--   * No existing row moves — and arm 4 MEASURES that before/after inside the block instead
--     of hard-coding a threshold: 722's "at most 1 site carries a posture" was true for a
--     few hours and read 2 by the afternoon, which is exactly how a verify arm rots.
--
-- ALREADY-APPLIED GUARD, because 722 taught this lane the cost of its absence: 722 was
-- applied by hand and not recorded, and its fully idempotent body (CREATE OR REPLACE +
-- DROP TRIGGER IF EXISTS) left it reading as innocently pending for the runner's probe all
-- day. This file RAISEs 'already applied' when the live function body already carries
-- growth_posture_set_at, so the runner marks it LIKELY ALREADY APPLIED and skips it rather
-- than replaying. Record it in the same motion as applying it:
--   ./scripts/migration/run-migrations.sh --record-only 752_growth_posture_records_when_and_by_whom.sql --note '<what you verified>'
--
-- ROLLBACK: 752_growth_posture_records_when_and_by_whom_ROLLBACK.sql restores 722's body.
--
-- _HOLD: held back from the runner while the council verdict is pending (the runner LISTS a
-- _HOLD file and never applies it). Apply by hand once approved, record in the same motion,
-- and drop the suffix as the record of application — naming BOTH paths on the commit:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < <this file>
--   git mv <this file> 752_growth_posture_records_when_and_by_whom.sql
--   ./scripts/migration/run-migrations.sh --record-only 752_growth_posture_records_when_and_by_whom.sql --note '<what you verified>'

BEGIN;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_proc
     WHERE proname = 'sites_born_holding_growth'
       AND prosrc LIKE '%growth_posture_set_at%'
  ) THEN
    RAISE EXCEPTION '752 already applied: sites_born_holding_growth() already stamps growth_posture_set_at';
  END IF;
END $$;

CREATE OR REPLACE FUNCTION sites_born_holding_growth() RETURNS trigger AS $fn$
BEGIN
  -- ABSENT means "nobody said" and becomes hold. A row that names the key KEEPS what it
  -- names, in either direction (722). A stated value also gets NO record from here.
  IF NEW.settings IS NULL
     OR NEW.settings->'maintenance_profile'->>'growth_posture' IS NULL THEN

    -- Materialise the parent first: jsonb_set only creates the LAST path element (722,
    -- and migration 291's header before it).
    NEW.settings := COALESCE(NEW.settings, '{}'::jsonb);
    IF NEW.settings->'maintenance_profile' IS NULL THEN
      NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile}', '{}'::jsonb, true);
    END IF;
    NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile,growth_posture}', '"hold"'::jsonb, true);

    -- 752: the hold records when and by whom, beside the posture it stamped.
    NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile,growth_posture_set_at}',
                              to_jsonb(now()), true);
    NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile,growth_posture_set_by}',
                              '"trg_sites_born_holding_growth"'::jsonb, true);
    NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile,growth_posture_reason}',
                              to_jsonb('born held — migration 722 (owner ruling 2026-09-02: a new site is released by a human)'::text), true);
  END IF;

  RETURN NEW;
END;
$fn$ LANGUAGE plpgsql;

DO $$
DECLARE
  v_before bigint; v_after bigint;
  v_id_a uuid; v_id_c uuid;
  v_posture text; v_set_by text; v_set_at timestamptz; v_reason text;
  v_c_posture text; v_c_set_at text;
  v_upsert_posture text;
BEGIN
  SELECT count(*) INTO v_before FROM sites
   WHERE settings->'maintenance_profile'->>'growth_posture' IS NOT NULL;

  -- ARM 1: a born-held site carries all four keys, set_by names the trigger, set_at is now.
  -- The ::timestamptz cast IS the reader's arm — the CronJob parses set_at back as a
  -- timestamp, and a value that does not cast aborts this block with 22007.
  INSERT INTO sites (domain, name, network_id, status)
  VALUES ('zz-752-probe-a.invalid', 'zz-a',
          (SELECT network_id FROM sites WHERE network_id IS NOT NULL LIMIT 1), 'active')
  RETURNING id INTO v_id_a;

  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture', 'open'),
         settings->'maintenance_profile'->>'growth_posture_set_by',
         (settings->'maintenance_profile'->>'growth_posture_set_at')::timestamptz,
         settings->'maintenance_profile'->>'growth_posture_reason'
    INTO v_posture, v_set_by, v_set_at, v_reason
    FROM sites WHERE id = v_id_a;

  IF v_posture <> 'hold' THEN
    RAISE EXCEPTION '752 ARM 1: posture reads %, expected hold — 722 regressed', v_posture;
  END IF;
  IF v_set_by IS DISTINCT FROM 'trg_sites_born_holding_growth' THEN
    RAISE EXCEPTION '752 ARM 1: set_by reads %, expected the trigger''s own name', v_set_by;
  END IF;
  IF v_set_at IS NULL OR abs(extract(epoch FROM now() - v_set_at)) > 60 THEN
    RAISE EXCEPTION '752 ARM 1: set_at reads %, expected within 60s of now()', v_set_at;
  END IF;
  IF v_reason IS NULL OR v_reason NOT LIKE 'born held%' THEN
    RAISE EXCEPTION '752 ARM 1: reason reads %, expected the born-held reason', v_reason;
  END IF;

  -- ARM 3 (numbered as in 722): an explicit opt-out keeps its posture AND gets no record.
  INSERT INTO sites (domain, name, network_id, status, settings)
  VALUES ('zz-752-probe-c.invalid', 'zz-c',
          (SELECT network_id FROM sites WHERE network_id IS NOT NULL LIMIT 1), 'active',
          '{"maintenance_profile": {"growth_posture": "open"}}'::jsonb)
  RETURNING id INTO v_id_c;

  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture', 'open'),
         settings->'maintenance_profile'->>'growth_posture_set_at'
    INTO v_c_posture, v_c_set_at
    FROM sites WHERE id = v_id_c;
  IF v_c_posture <> 'open' THEN
    RAISE EXCEPTION '752 ARM 3: a stated opt-out was overridden (reads %)', v_c_posture;
  END IF;
  IF v_c_set_at IS NOT NULL THEN
    RAISE EXCEPTION '752 ARM 3: a STATED posture was given a set_at (%) — the record must only accompany a stamp this trigger made', v_c_set_at;
  END IF;

  -- ARM 4: no existing row moved. Measured before/after, not thresholded.
  SELECT count(*) INTO v_after FROM sites
   WHERE settings->'maintenance_profile'->>'growth_posture' IS NOT NULL
     AND domain NOT LIKE 'zz-752-probe-%';
  IF v_after <> v_before THEN
    RAISE EXCEPTION '752 ARM 4: postures on existing rows went % -> %', v_before, v_after;
  END IF;

  -- ARM 5: INSERT-only survives (722 arm 6's shape). Strip the key from probe A, UPSERT it
  -- the way ensure_site_record does, and it must stay open — if the trigger has acquired
  -- an UPDATE arm, every unset site would be held within one sweep.
  UPDATE sites SET settings = settings #- '{maintenance_profile,growth_posture}'
   WHERE id = v_id_a;

  INSERT INTO sites (domain, name, network_id, status)
  VALUES ('zz-752-probe-a.invalid', 'zz-a',
          (SELECT network_id FROM sites WHERE id = v_id_a), 'active')
  ON CONFLICT (domain) DO UPDATE SET updated_at = now();

  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture', 'open')
    INTO v_upsert_posture FROM sites WHERE id = v_id_a;
  IF v_upsert_posture <> 'open' THEN
    RAISE EXCEPTION '752 ARM 5: a routine UPSERT re-held an unset site (reads %) — the trigger has an UPDATE arm', v_upsert_posture;
  END IF;

  DELETE FROM sites WHERE domain LIKE 'zz-752-probe-%';

  RAISE NOTICE '752 OK: born-held probe set_at=% set_by=%; opt-out kept open with no record; existing postures %->%; unset-stays-open under upsert=%',
    v_set_at, v_set_by, v_before, v_after, v_upsert_posture;
END $$;

COMMIT;
