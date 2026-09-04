-- FILE: docs/agent_docs/sql_for_agents/774_sites_followup_sent_at_is_the_follow_up_emails_once_only_claim.sql
--
-- bugs_open/477 step B, part 1 of 2: the column the follow-up email is claimed on.
--
-- WHY A COLUMN AND NOT A WORK ITEM. The follow-up is a scheduled thing that
-- emails customers, and the failure that matters is not "it never went" but "it
-- went every night". delivery-email-sender is safe because the handover stamp
-- claims the row: `UPDATE sites SET handed_over_at = … WHERE handed_over_at IS
-- NULL` can be won by exactly one statement. A scheduled sender has no
-- equivalent unless one is built, and this column is it. platform/delivery's
-- ClaimFollowup claims on `followup_sent_at IS NULL` in the same way, so
-- at-most-once is a property of the UPDATE rather than of anybody remembering.
--
-- WHAT IT IS NOT. It is not a send LOG and must not be read as one: it records
-- that a follow-up was CLAIMED. The stamp deliberately stands even if the send
-- then fails, because for a chase email not sending beats sending twice. The
-- send itself is visible in the orchestration run and the mail host, and if a
-- reliable per-send record is ever wanted that is a separate table, not a
-- second meaning bolted onto this column.
--
-- NULLABLE, NO DEFAULT, NO BACKFILL, and that is the whole change: every
-- existing row reads "no follow-up claimed yet", which is true of all 60 of them
-- (1 handed over, 0 confirmed, measured 2026-09-04 11:43Z). Applying this sends
-- nothing and changes no behaviour: nothing writes the column until the action
-- ships in an image, and nothing calls the action until 775 seeds an agent that
-- is deliberately DISABLED.
--
-- Council: submitted with the step B round. Migrations are in gate scope.
-- Rollback sidecar: 774_..._ROLLBACK.sql (drops the column).

BEGIN;

ALTER TABLE sites ADD COLUMN IF NOT EXISTS followup_sent_at timestamptz;

COMMENT ON COLUMN sites.followup_sent_at IS
  'bugs_open/477: when the post-delivery follow-up email was CLAIMED for this site. '
  'The at-most-once gate for that email (delivery.ClaimFollowup claims on IS NULL, '
  'exactly as handed_over_at claims the delivery). Not a send log: the stamp stands '
  'even if the send then fails, because for a chase email not sending beats sending twice.';

-- VERIFY — a DO block, not a SELECT. A verify block made of SELECTs cannot stop
-- the COMMIT: ON_ERROR_STOP ignores a non-empty result set, so the migration
-- reports and commits anyway. RAISE is what actually aborts.
DO $$
DECLARE
  -- v_ prefixes are load-bearing: a PL/pgSQL variable named is_nullable is
  -- AMBIGUOUS against information_schema.columns' own column of that name, and
  -- Postgres refuses the whole block. Caught by dry-running this file before
  -- applying it, which is the only reason it is not a failed migration.
  v_col_type text;
  v_is_nullable text;
  v_col_default text;
BEGIN
  SELECT data_type, is_nullable, column_default
    INTO v_col_type, v_is_nullable, v_col_default
    FROM information_schema.columns
   WHERE table_schema = 'public' AND table_name = 'sites' AND column_name = 'followup_sent_at';

  IF v_col_type IS NULL THEN
    RAISE EXCEPTION '774 FAILED: sites.followup_sent_at does not exist after the ALTER';
  END IF;
  IF v_col_type <> 'timestamp with time zone' THEN
    RAISE EXCEPTION '774 FAILED: sites.followup_sent_at is %, want timestamp with time zone', v_col_type;
  END IF;
  -- NOT NULL or a DEFAULT would both be wrong, and wrong in the dangerous
  -- direction: a default of now() would mark every site as already followed up,
  -- and NOT NULL would force a backfill that means the same thing.
  IF v_is_nullable <> 'YES' THEN
    RAISE EXCEPTION '774 FAILED: sites.followup_sent_at is NOT NULL; every existing row would need a value that would read as "already sent"';
  END IF;
  IF v_col_default IS NOT NULL THEN
    RAISE EXCEPTION '774 FAILED: sites.followup_sent_at has default %, which would stamp existing rows', v_col_default;
  END IF;

  -- The population must be untouched: this migration claims nothing.
  IF (SELECT count(*) FROM sites WHERE followup_sent_at IS NOT NULL) <> 0 THEN
    RAISE EXCEPTION '774 FAILED: % site(s) already carry followup_sent_at; this migration must not stamp any',
      (SELECT count(*) FROM sites WHERE followup_sent_at IS NOT NULL);
  END IF;

  RAISE NOTICE '774 OK: sites.followup_sent_at added, nullable, no default, 0 rows stamped';
END $$;

COMMIT;
