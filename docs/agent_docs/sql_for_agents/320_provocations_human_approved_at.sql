-- 320_provocations_human_approved_at.sql
--
-- Adds `provocations.human_approved_at` so that GATE approval and HUMAN approval
-- stop being the same fact.
--
-- WHY (owner rulings 2026-08-09)
-- The owner reversed PLAN §10's no-human-approval decision: a human now approves a
-- provocation before it publishes. The table had one `status` column and no way to
-- record a second approval, so the human's consent was being carried implicitly by
-- `publish_on` — the scheduler's date was, in practice, the last gate before
-- publication.
--
-- That works and it is fragile in a specific way: it makes `schedule_provocations`
-- unschedulable. Put the scheduler on a cron and it silently re-automates the exact
-- step the owner just handed to a human, while looking like ordinary plumbing.
-- A separate column makes the two approvals distinguishable, so dating can go back
-- to being dating.
--
-- THE NEAR-MISS THAT PROMPTED IT, recorded because the shape recurs:
-- six model-written drafts arrived PRE-DATED, one dated the same day. Approving them
-- would have published a model's prose to a live site within six hours, under the
-- owner's name, with no human in the loop — in the same hour he ruled otherwise.
-- **A dated draft is a publish waiting for one status change.**
--
-- ORDERING — THIS HALF IS SAFE TO APPLY ALONE, AND THAT IS DELIBERATE
-- Nothing reads the column yet. `render_provocation_feed` still selects on
-- `status='approved' AND publish_on IS NOT NULL`; the matching Go change that adds
-- `AND human_approved_at IS NOT NULL` is inert until the next chassis roll. So:
--
--   * applied alone  -> no behaviour change at all (additive, nullable column)
--   * code rolls later -> anything WITHOUT the stamp stops publishing
--
-- The backfill below is what makes the second line safe. Every row that may
-- legitimately publish today is stamped now, so the roll cannot silently empty the
-- feed. Fail-closed is the right direction for the residue: an unstamped row stops
-- publishing rather than publishing unreviewed.
--
-- Idempotent; safe to re-run.

BEGIN;

ALTER TABLE provocations
  ADD COLUMN IF NOT EXISTS human_approved_at timestamptz,
  ADD COLUMN IF NOT EXISTS human_approved_by text;

COMMENT ON COLUMN provocations.human_approved_at IS
  'When a HUMAN approved this for publication — distinct from status=''approved'', which is the automated gate''s verdict. render_provocation_feed requires it (chassis > v1.0.1267). NULL means not yet approved by a person, and such a row never publishes.';
COMMENT ON COLUMN provocations.human_approved_by IS
  'Who approved it. Free text; the audit trail, not an access control.';

-- ---------------------------------------------------------------------------
-- Backfill 1: the owner's own writing.
-- These eight are his, written by hand, and have been live on the site for weeks.
-- Their human approval is not in question; it simply predates the column.
-- ---------------------------------------------------------------------------
UPDATE provocations
   SET human_approved_at = COALESCE(human_approved_at, created_at),
       human_approved_by = COALESCE(human_approved_by, 'owner (backfilled: authored by hand, published before the column existed)')
 WHERE domain = 'vonc.com' AND status = 'approved' AND source = 'human';

-- ---------------------------------------------------------------------------
-- Backfill 2: the six model-written provocations the owner read and approved
-- explicitly on 2026-08-09 ("those are good provocations to start us off").
-- This is a real approval, not a convenience: he was shown the title, the teaser
-- and the opening of each.
-- ---------------------------------------------------------------------------
UPDATE provocations
   SET human_approved_at = COALESCE(human_approved_at, now()),
       human_approved_by = COALESCE(human_approved_by, 'owner 2026-08-09 (reviewed titles + teasers + body openings in session)')
 WHERE domain = 'vonc.com' AND status = 'approved' AND source = 'llm'
   AND slug IN ('childhood-food-was-not-better','film-that-needs-explaining-has-failed',
                'live-music-is-worse','nobody-misses-pre-internet',
                'stopped-discovering-music-at-24','you-love-being-from-your-city');

-- Deliberately NOT backfilled: `calibration.vonc.com` (never publishable — the domain
-- is absent from `sites`), and anything `retired` or `draft`.

DO $$
DECLARE n_col int; n_unstamped int; n_stamped int; n_cal int;
BEGIN
  SELECT count(*) INTO n_col FROM information_schema.columns
   WHERE table_name='provocations' AND column_name IN ('human_approved_at','human_approved_by');
  IF n_col <> 2 THEN RAISE EXCEPTION 'expected both columns present, found %', n_col; END IF;

  -- THE ONE THAT MATTERS: after the Go half rolls, an approved+dated row without the
  -- stamp stops publishing. If any exists now, the roll would silently shrink the
  -- feed — so refuse here rather than discover it on the site.
  SELECT count(*) INTO n_unstamped FROM provocations
   WHERE domain <> 'calibration.vonc.com'
     AND status='approved' AND publish_on IS NOT NULL AND human_approved_at IS NULL;
  IF n_unstamped <> 0 THEN
    RAISE EXCEPTION 'REFUSING: % approved+dated row(s) carry no human_approved_at. Once the code half rolls they would silently stop publishing. Stamp them or explain them first.', n_unstamped;
  END IF;

  SELECT count(*) INTO n_stamped FROM provocations WHERE human_approved_at IS NOT NULL;
  IF n_stamped <> 14 THEN
    RAISE EXCEPTION 'expected 14 stamped rows (8 human + 6 owner-approved llm), found % — the backfill did not match what was measured', n_stamped;
  END IF;

  SELECT count(*) INTO n_cal FROM provocations
   WHERE domain='calibration.vonc.com' AND human_approved_at IS NOT NULL;
  IF n_cal <> 0 THEN RAISE EXCEPTION 'calibration rows were stamped (%) — they must never look publishable', n_cal; END IF;

  RAISE NOTICE 'human_approved_at added; % rows stamped; 0 approved+dated rows left unstamped', n_stamped;
END $$;

COMMIT;
