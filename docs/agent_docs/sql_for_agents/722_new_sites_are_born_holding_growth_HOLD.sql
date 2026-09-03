-- 722 — a NEW site is born holding growth (owner decision, 2026-09-02)
--
-- improvement_loop lane. Evidence: bugs_open/447 §7.
--
-- ⚠ _HOLD SUFFIX, DELIBERATE. Not applied, and held OUT of the migration runner
-- (SIDECAR_RE excludes _HOLD; it still lists it), because `run-migrations.sh --apply`
-- takes EVERY pending file and this must not go live before its verdict is read.
-- Drop the suffix ONLY after reading an APPROVED verdict on trail
-- 070347dd-c410-4cf2-b5e6-8c87e568a792.
--
-- WHAT THE OWNER DECIDED, narrowly: a brand-new site should be born with
-- growth_posture='hold' and released by a human when it is ready. He was asked in
-- exactly that form and chose it over "keep open, hold by hand" and over a
-- twin-pair-only variant.
--
-- ── WHY A TRIGGER AND NOT A COLUMN DEFAULT (rounds 1-2 got this WRONG) ──────────
--
-- Rounds 1 and 2 of this migration set the `sites.settings` COLUMN DEFAULT, on the
-- stated ground that "no creation path names `settings` on insert". THAT CLAIM WAS
-- FALSE. It was drawn from two citations; the council's prior_art_librarian seat
-- objected that a coverage claim needs a sweep, and the sweep found **15 site-creation
-- call sites, of which TWO name `settings`**:
--
--   * docs024_key_docs_latest/gamedesign_uk_rebuild/SEED_2026-09-02_...sql — which
--     created oxenunity.com THE SAME DAY, and is the SEED shape CLAUDE.md tells every
--     lane to use ("Seed the site row and its specs (SEED_*.sql)").
--   * sql_for_tables/005_content_components.sql — the system.internal row.
--
-- A column default is bypassed entirely by an INSERT that names the column, so the
-- default version would have shipped a hold that the estate's OWN DOCUMENTED creation
-- path silently defeats. DEMONSTRATED, not reasoned: with the default applied and no
-- trigger, an INSERT naming `settings` reads back as **open**.
--
-- A BEFORE INSERT trigger reads NEW.settings whatever produced it, so all 15 paths
-- are covered — the same "door, not the producer" argument that moved the growth door
-- itself into writeWorkItem (council corr 1e735fa2).
--
-- ── BEFORE INSERT ONLY, AND WHAT THAT DOES *NOT* GUARANTEE ────────────────
--
-- `ensure_site_record` UPSERTs the site row on EVERY improvement-loop pass (~50/day
-- fleet-wide). A trigger that also fired on UPDATE would silently re-hold every site
-- the owner had released, for ever, and nothing would report it. INSERT-only makes
-- that unrepresentable rather than merely avoided. Arm 5 of the verify block induces
-- an UPSERT of an existing site and fails if its posture moves.
--
--
-- ⚠ BUT "BEFORE INSERT, THEREFORE EXISTING ROWS ARE SAFE" IS FALSE IN THE PRESENCE
-- OF AN UPSERT, and rounds 1-3 of this migration said it was. The council's guardian
-- seat raised it (HIGH, round 3) and it is right: Postgres fires a BEFORE INSERT row
-- trigger for EVERY row a statement PROPOSES, before the conflict is detected, and
-- `EXCLUDED` is that POST-TRIGGER row. So an UPSERT whose DO UPDATE branch writes
-- `settings = EXCLUDED.settings` carries this trigger's stamp onto a row that
-- already existed.
--
-- INDUCED, both branches, 2026-09-03, with this trigger installed:
--   ON CONFLICT (domain) DO UPDATE SET updated_at = now()           -> cookly.uk stays OPEN
--   ON CONFLICT (domain) DO UPDATE SET settings = EXCLUDED.settings -> lampenkap.com becomes HELD
-- The second is a RELEASED site being silently re-held.
--
-- IT IS NOT REACHABLE TODAY, and that is a measurement rather than a hope.
-- [MEASURED 2026-09-03] the four UPSERTs on `sites` set `updated_at = now()`
-- (upsertSite), `email = COALESCE(sites.email, EXCLUDED.email)` (one seed), or
-- DO NOTHING (two). None writes `settings`. And pg_trigger carried ZERO
-- non-internal triggers on `sites` before this one, so there is no ordering
-- question either.
--
-- THE STANDING RULE THIS LEAVES, one reasonable edit away from being broken:
-- **no `DO UPDATE SET` on `sites` may write `settings` from `EXCLUDED`.** Merge into
-- the existing row instead — `settings = sites.settings || <your jsonb>` never
-- carries the stamp. Recorded in LANDMINES.md with the induction recipe, because
-- ARM 5 TESTS ONLY TODAY'S CLAUSE and will keep passing after someone widens it.
-- ── ABSENT MEANS "NOBODY SAID"; A STATED VALUE IS KEPT ─────────────────────────
--
-- The trigger stamps only when the key is ABSENT. A row that names it keeps what it
-- names, in either direction — so `"growth_posture": "open"` in a seed is an explicit,
-- greppable opt-out, and this trigger will not override it. That is the difference
-- between a default nobody can see and a decision somebody wrote down.
--
-- ⚠ jsonb_set ONLY CREATES THE LAST PATH ELEMENT. Writing
-- {maintenance_profile,growth_posture} into a settings value with no
-- maintenance_profile object SILENTLY NO-OPS and returns the input unchanged —
-- migration 291's header records the same trap for record_audit_pass. DEMONSTRATED:
-- the obvious one-line version leaves `settings = {}` and the row reads **open**. The
-- parent is materialised first, deliberately, in two steps.
--
-- ⚠ THE VERIFY BLOCK IS THE READER'S OWN SQL, NOT A SHAPE TEST. Round 1's verify was
-- `column_default LIKE '%growth_posture%'`, which passes on {"growth_posture":"hold"}
-- and on {"maintenance":{"growth_posture":"hold"}} — i.e. on exactly the wrong-depth
-- defect it existed to catch. Every arm below now reads the row back through
-- `COALESCE(settings->'maintenance_profile'->>'growth_posture','open')`, copied from
-- datahelpers/growth_posture.go SiteGrowthPosture.
--
-- ── WHAT IT DOES NOT DO ────────────────────────────────────────────────────────
--
-- It touches no existing row: a BEFORE INSERT trigger cannot fire on rows that already
-- exist. All 39 active sites [MEASURED 2026-09-02] keep the posture they have (unset =
-- open). Arm 4 aborts if that turns out false.
--
-- RESIDUAL, named not hidden: nothing reports "site held for longer than N days", so a
-- site nobody releases stops growing silently. Owed by the improvement_loop lane and
-- recorded in its handoff.

BEGIN;

-- The default stays as it was. A column default cannot do this job — see the header.
ALTER TABLE sites ALTER COLUMN settings SET DEFAULT '{}'::jsonb;

CREATE OR REPLACE FUNCTION sites_born_holding_growth() RETURNS trigger AS $fn$
BEGIN
  -- ABSENT means "nobody said" and becomes hold. A row that names the key KEEPS what
  -- it names, in either direction — so `"growth_posture": "open"` in a seed is an
  -- explicit, greppable opt-out and this trigger will not override it. That is the
  -- difference between a default nobody can see and a decision somebody wrote down.
  IF NEW.settings IS NULL
     OR NEW.settings->'maintenance_profile'->>'growth_posture' IS NULL THEN

    -- ⚠ MATERIALISE THE PARENT FIRST. jsonb_set only creates the LAST element of the
    -- path: writing {maintenance_profile,growth_posture} into a settings value that
    -- has no maintenance_profile object SILENTLY NO-OPS and returns the input
    -- unchanged. That exact trap is recorded in migration 291's header for
    -- record_audit_pass, which had to do the same thing. Two steps, deliberately.
    NEW.settings := COALESCE(NEW.settings, '{}'::jsonb);
    IF NEW.settings->'maintenance_profile' IS NULL THEN
      NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile}', '{}'::jsonb, true);
    END IF;
    NEW.settings := jsonb_set(NEW.settings, '{maintenance_profile,growth_posture}', '"hold"'::jsonb, true);
  END IF;

  RETURN NEW;
END;
$fn$ LANGUAGE plpgsql;

-- BEFORE INSERT ONLY. See the header for what this does NOT guarantee under an
-- UPSERT whose DO UPDATE branch writes settings from EXCLUDED. ensure_site_record UPSERTs
-- the site row on EVERY improvement-loop pass (~50/day fleet-wide), so a trigger that
-- also fired on UPDATE would silently re-hold every site the owner had released,
-- for ever. INSERT-only makes "the loop re-holds a released site" unrepresentable
-- rather than merely avoided.
DROP TRIGGER IF EXISTS trg_sites_born_holding_growth ON sites;
CREATE TRIGGER trg_sites_born_holding_growth
  BEFORE INSERT ON sites
  FOR EACH ROW EXECUTE FUNCTION sites_born_holding_growth();

DO $$
DECLARE
  v_id_a uuid; v_id_b uuid; v_id_c uuid;
  v_a text; v_b text; v_c text; v_untouched bigint; v_existing_posture text; v_unset_posture text;
BEGIN
  -- ARM 1: the Go path — INSERT naming no settings at all.
  INSERT INTO sites (domain, name, network_id, status)
  VALUES ('zz-722-probe-a.invalid', 'zz-a',
          (SELECT network_id FROM sites WHERE network_id IS NOT NULL LIMIT 1), 'active')
  RETURNING id INTO v_id_a;

  -- ARM 2: the SEED path that NAMES settings — the case a column default could not
  -- reach, and the case that actually exists (oxenunity.com, seeded 2026-09-02).
  INSERT INTO sites (domain, name, network_id, status, settings)
  VALUES ('zz-722-probe-b.invalid', 'zz-b',
          (SELECT network_id FROM sites WHERE network_id IS NOT NULL LIMIT 1), 'active',
          '{"managed_by": "hand", "seeded_by": "probe"}'::jsonb)
  RETURNING id INTO v_id_b;

  -- ARM 3: an EXPLICIT opt-out must survive untouched.
  INSERT INTO sites (domain, name, network_id, status, settings)
  VALUES ('zz-722-probe-c.invalid', 'zz-c',
          (SELECT network_id FROM sites WHERE network_id IS NOT NULL LIMIT 1), 'active',
          '{"maintenance_profile": {"growth_posture": "open"}}'::jsonb)
  RETURNING id INTO v_id_c;

  -- Read all three back through THE READER'S OWN SQL (datahelpers/growth_posture.go,
  -- SiteGrowthPosture), not through the shape they were written with.
  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture','open') INTO v_a FROM sites WHERE id=v_id_a;
  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture','open') INTO v_b FROM sites WHERE id=v_id_b;
  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture','open') INTO v_c FROM sites WHERE id=v_id_c;

  IF v_a <> 'hold' THEN RAISE EXCEPTION '722 ARM 1 (no settings column): reads %, expected hold', v_a; END IF;
  IF v_b <> 'hold' THEN RAISE EXCEPTION '722 ARM 2 (seed NAMES settings): reads %, expected hold — this is the arm a column DEFAULT fails', v_b; END IF;
  IF v_c <> 'open' THEN RAISE EXCEPTION '722 ARM 3 (explicit opt-out): reads %, expected open — the trigger must not override a stated decision', v_c; END IF;

  -- ARM 2 must also have PRESERVED the seed's own keys.
  IF (SELECT settings->>'managed_by' FROM sites WHERE id=v_id_b) IS DISTINCT FROM 'hand' THEN
    RAISE EXCEPTION '722 ARM 2: the trigger destroyed the seed''s own settings keys';
  END IF;

  -- ARM 4, THE DISCONFIRMING ONE: no existing row moved, and an UPSERT of an existing
  -- domain must not re-hold it.
  SELECT count(*) INTO v_untouched FROM sites
   WHERE settings->'maintenance_profile'->>'growth_posture' IS NOT NULL
     AND domain NOT LIKE 'zz-722-probe-%';
  IF v_untouched > 1 THEN
    RAISE EXCEPTION '722 ARM 4: % existing sites carry a posture, expected at most 1 (gamedesign.uk, by hand)', v_untouched;
  END IF;

  -- ARM 5, PROBE-SCOPED. Council round 4, debug_historian (HIGH): the first version
  -- of this arm UPSERTed cookly.uk — a REAL production site — and the cleanup below
  -- only matches 'zz-722-probe-%', so on COMMIT it would have permanently bumped that
  -- site's updated_at. That is not a cosmetic leak: `improvement-sweep`'s pre_query is
  -- ORDER BY s.updated_at ASC NULLS FIRST LIMIT 1, so bumping it sends the site to the
  -- BACK of the improvement loop's own rotation. A verify block would have silently
  -- degraded the subsystem this migration belongs to. It never showed up in testing
  -- because every test run was rolled back; only an APPLY would have committed it.
  --
  -- The probe-scoped version also tests the RIGHT scenario. What matters is not "an
  -- already-open site stays open" but "a site the owner RELEASED is not re-held by the
  -- next routine UPSERT" — so release it first, then upsert it with upsertSite's REAL
  -- clause (site_db_actions.go:1111, `ON CONFLICT (domain) DO UPDATE SET updated_at =
  -- NOW()`), and assert the release survived.
  UPDATE sites
     SET settings = jsonb_set(settings, '{maintenance_profile,growth_posture}', '"open"'::jsonb, true)
   WHERE id = v_id_a;

  INSERT INTO sites (domain, name, network_id, status)
  VALUES ('zz-722-probe-a.invalid', 'zz-a',
          (SELECT network_id FROM sites WHERE id = v_id_a), 'active')
  ON CONFLICT (domain) DO UPDATE SET updated_at = now();

  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture','open') INTO v_existing_posture
    FROM sites WHERE id = v_id_a;
  IF v_existing_posture <> 'open' THEN
    RAISE EXCEPTION '722 ARM 5: a RELEASED site was re-held (reads %) by a routine UPSERT '
      '— the trigger fired on the UPDATE arm', v_existing_posture;
  END IF;

  -- ARM 6, AND IT IS THE ONE THAT ACTUALLY DISCRIMINATES. Arm 5 above is necessary but
  -- NOT sufficient, and I only found that by mutating: changing the trigger to
  -- `BEFORE INSERT OR UPDATE` leaves arm 5 PASSING, because a released site carries the
  -- key with value 'open' and this trigger only stamps when the key is ABSENT.
  --
  -- The population an UPDATE-firing trigger would actually damage is the 38 sites that
  -- carry NO posture key at all [MEASURED 2026-09-02] — they would be silently held on
  -- their next routine UPSERT, i.e. within ~15 minutes, fleet-wide. So arm 6 builds
  -- exactly that shape: a site with the key REMOVED, upserted the way ensure_site_record
  -- does. It fails if the trigger has acquired an UPDATE arm.
  UPDATE sites SET settings = settings #- '{maintenance_profile,growth_posture}'
   WHERE id = v_id_b;

  INSERT INTO sites (domain, name, network_id, status)
  VALUES ('zz-722-probe-b.invalid', 'zz-b',
          (SELECT network_id FROM sites WHERE id = v_id_b), 'active')
  ON CONFLICT (domain) DO UPDATE SET updated_at = now();

  SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture','open') INTO v_unset_posture
    FROM sites WHERE id = v_id_b;
  IF v_unset_posture <> 'open' THEN
    RAISE EXCEPTION '722 ARM 6: a site with NO posture key was stamped % by a routine UPSERT '
      '— the trigger has an UPDATE arm and would hold all 38 unset sites within one sweep',
      v_unset_posture;
  END IF;

  DELETE FROM sites WHERE domain LIKE 'zz-722-probe-%';

  RAISE NOTICE '722 OK: no-settings=%, seed-names-settings=%, explicit-opt-out=%, released-survives-upsert=%, unset-stays-open=%, existing rows carrying a posture=%',
    v_a, v_b, v_c, v_existing_posture, v_unset_posture, v_untouched;
END $$;

COMMIT;
