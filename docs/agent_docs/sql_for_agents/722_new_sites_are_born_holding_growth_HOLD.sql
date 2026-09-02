-- 722 — a NEW site is born holding growth (owner decision, 2026-09-02)
--
-- ⚠ _HOLD SUFFIX, DELIBERATE. Not applied, and held OUT of the migration runner
-- (SIDECAR_RE excludes _HOLD; it still lists it). Two reasons, both live:
--   1. The council round could NOT be sent — the kubeconfig token expired fleet-wide
--      mid-submission ("You must be logged in to the server (Unauthorized)"). The
--      trigger refused to report success, nothing was spent, and no correlation names
--      anything. RE-SUBMIT before applying; the submission file is unchanged at
--      <scratch>/growth_default_submission.json.
--   2. `run-migrations.sh --apply` takes EVERY pending file, so an unreviewed migration
--      sitting un-suffixed in this directory is one other session's routine apply away
--      from being live. The suffix is the only thing that holds it.
-- Remove the suffix ONLY after the council verdict is read.
--
-- improvement_loop lane. Evidence: bugs_open/447 §7 (this lane's contribution).
--
-- WHAT THE OWNER DECIDED, and it is narrow: a brand-new site should be born with
-- growth_posture='hold' and released by a human when it is ready. He was asked the
-- question in exactly that form ("what should a BRAND-NEW site's growth posture be?")
-- and chose "born hold until you release it" over "keep open, hold by hand" and over
-- a twin-pair-only variant.
--
-- ── WHY THIS IS A COLUMN DEFAULT AND NOT A LINE OF GO ────────────────────────────
--
-- A DEFAULT applies to every INSERT that does not name `settings`, whoever writes it:
-- upsertSite in site_db_actions.go, the SQL seeds that insert a site row directly
-- (436_tools_api_gripper_intake.sql does), and any hand-insert a future lane writes.
-- That is the same "door, not the producer" argument the growth door itself was moved
-- to writeWorkItem for (council corr 1e735fa2): a per-producer guard covers the
-- producers you found.
--
-- AND IT IS THE ONLY SHAPE THAT CANNOT RE-HOLD A RELEASED SITE. `upsertSite` is an
-- UPSERT — `ON CONFLICT (domain) DO UPDATE` — and `ensure_site_record` runs it on
-- EVERY improvement-loop pass. A Go change that stamped the posture would have to be
-- careful to write it on the insert arm only; a column default cannot fire on UPDATE
-- at all, so "the loop silently re-holds a site the owner released" is not a bug this
-- version can have. That is the whole reason for choosing it.
--
-- ── WHAT IT DOES NOT DO ─────────────────────────────────────────────────────────
--
-- It does not touch one existing row. A column default applies only to new rows, so
-- all 39 active sites [MEASURED 2026-09-02] keep the posture they have (unset = open).
-- The owner was asked about NEW sites and this changes only new sites. Anyone wanting
-- to hold an existing site sets the key by hand, as the gamedesign.uk lane did today.
--
-- ⚠ THE ONE WAY TO DEFEAT IT: an INSERT that names `settings` explicitly bypasses the
-- default entirely and the site is born open. [MEASURED 2026-09-02] no creation path
-- does today — the Go path names (domain, name, network_id, status) and the SQL seeds
-- name (id, domain, status). A future one might, silently. The parity test named in
-- the verify block is what would catch it; there is none today, and that is stated
-- rather than implied.
--
-- ⚠ SETTINGS MERGES SURVIVE THIS, CHECKED: every other writer of sites.settings uses
-- jsonb_set or `||` (4 of them), and the only non-merge is 625's `#-` deleting one
-- path. So a stamped posture is not wiped by the next settings write.

BEGIN;

ALTER TABLE sites
  ALTER COLUMN settings
  SET DEFAULT '{"maintenance_profile": {"growth_posture": "hold"}}'::jsonb;

-- ── VERIFY, as DO/RAISE and not a SELECT ────────────────────────────────────────
-- A verify block made of SELECTs cannot stop the COMMIT: ON_ERROR_STOP ignores a
-- non-empty result set (LANDMINES.md). These raise, and because the whole file is
-- wrapped in BEGIN/COMMIT, a RAISE here rolls the ALTER TABLE back with it — the
-- guard gates the mutation rather than reporting on one that already landed.
DO $$
DECLARE
  v_default text;
  v_reads_as text;
  v_changed_rows bigint;
BEGIN
  SELECT column_default INTO v_default
    FROM information_schema.columns
   WHERE table_name = 'sites' AND column_name = 'settings';

  IF v_default IS NULL THEN
    RAISE EXCEPTION '722: sites.settings has no default at all';
  END IF;

  -- STRUCTURAL, NOT A SUBSTRING (council round 1, editquality, medium). A
  -- LIKE '%growth_posture%' check passes on {"growth_posture":"hold"} (no nesting)
  -- and on {"maintenance":{"growth_posture":"hold"}} (wrong parent) — i.e. on
  -- exactly the wrong-depth defect it is supposed to catch, which this estate
  -- produces repeatedly. So resolve the default through THE READER'S OWN PATH,
  -- copied verbatim from datahelpers/growth_posture.go SiteGrowthPosture:
  --   SELECT COALESCE(settings->'maintenance_profile'->>'growth_posture','open')
  -- The default text is a quoted literal with a ::jsonb cast; strip the cast and
  -- the outer quotes to recover the literal, then ask the reader's question of it.
  v_reads_as := COALESCE(
    (btrim(split_part(v_default, '::', 1), $q$'$q$)::jsonb)
      -> 'maintenance_profile' ->> 'growth_posture',
    'open');

  IF v_reads_as <> 'hold' THEN
    RAISE EXCEPTION '722: a NEW site would read as %, not hold. The default is % — the key '
      'is at the wrong depth or under the wrong parent, and this migration would have '
      'shipped ZERO effect while reporting success.', v_reads_as, v_default;
  END IF;

  -- The disconfirming half: this migration must have changed NO existing row.
  --
  -- IF THIS ARM FIRES, READ IT BEFORE ASSUMING A BUG IN THIS MIGRATION (council
  -- round 1, guardian, low). It counts sites carrying ANY posture, and it cannot
  -- distinguish "the default leaked onto existing rows" (impossible — a column
  -- default never applies to an UPDATE) from "another lane set a posture by hand
  -- between this file being written and being applied", which is legitimate and
  -- expected: gamedesign.uk was set by hand on 2026-09-02 and is the 1 this
  -- threshold allows. A fire here almost certainly means the SECOND thing. Re-read
  -- the rows; if they were set deliberately, raise the threshold in a follow-up
  -- commit that says whose they are. Do NOT delete anyone's posture to make it pass.
  SELECT count(*) INTO v_changed_rows
    FROM sites
   WHERE settings->'maintenance_profile'->>'growth_posture' IS NOT NULL;

  IF v_changed_rows > 1 THEN
    RAISE EXCEPTION '722: % sites carry a growth_posture; this file expected at most 1 '
      '(gamedesign.uk, held by hand 2026-09-02). See the comment above this check — the '
      'likely cause is another lane holding a site deliberately, not a leak.', v_changed_rows;
  END IF;

  RAISE NOTICE '722 OK: a new site reads as "%"; % existing site(s) carry a posture, all set by hand.',
    v_reads_as, v_changed_rows;
END $$;

COMMIT;
