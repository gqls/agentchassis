-- 264_executable_checks_means_the_consumer_runs_it.sql
--
-- WHAT THIS CHANGES: the documented meaning of experience_patterns.executable_checks.
-- No data, no constraint, no column — one COMMENT, applied so the column does not
-- go on describing something the code stopped doing.
--
-- WHY. Migration 230 introduced the column as "Count of criteria checks a named
-- tier can actually execute", and the validator counted exactly that: any check
-- whose type SOME tier implements. But the register's only consumer,
-- verify_site_experience, evaluates Tier 2 and holds every Tier 4 check back —
-- so the stored number included precisely the checks that would never run.
--
-- That is not a rounding error. The column exists to back
-- experience_patterns_approved_needs_executable_check, the constraint that makes
-- "approved while asserting nothing" unrepresentable. A count inflated by checks
-- nothing executes is the vacuous-pass defect wearing a number, and it is the
-- defect this whole register exists to end.
--
-- It was found from two directions independently, which is why it is worth the
-- migration. The approval council's honesty seat read CC-001 and said
-- "executable_checks: 2 overstates coverage; only list_exists is unambiguously
-- executable today", and its deferral_honesty seat said the ledger was "ambiguous
-- exactly where honesty matters most". Reading the two code paths against each
-- other gave the same answer: the validator's counting and the consumer's tier
-- split were two copies of one rule, and they disagreed. They are now one
-- function (experienceNeedsBrowserReason), so they cannot drift again, and a
-- Tier 4 clause is carried as a DEFERRAL with its reason rather than counted or
-- dropped.
--
-- EFFECT ON EXISTING ROWS: none by this file. The stored counts are rewritten
-- when each entry next goes through write_experience_pattern (the 240 re-seed).
-- Measured before writing this, so the change is not a surprise: no entry falls
-- to zero — the smallest becomes 1 — and CC-001 goes 2 -> 1, which is the exact
-- number the council reached by reading the entry.
--
--   name                          | stored | tier-2 only
--   arrow-and-swipe-card-carousel |      5 |           3
--   count-up-stat-band            |      2 |           2
--   feed-driven-teaser-list       |      2 |           1
--   feed-promised-cta             |      2 |           2
--   hover-reveal-card-grid        |      2 |           1
--   illustrated-statement-block   |      1 |           1
--   scroll-snap-card-track        |      2 |           1
--   teaser-detail-deeplink        |      6 |           5
--   timed-remote-challenge-loop   |      6 |           5
--
-- ORDERING: comment-only, so it is safe in either order with the image. Apply it
-- WITH or AFTER the chassis carrying the new counting; applying it before simply
-- means the comment is aspirational for a few minutes.
--
-- APPLY (never via --apply, which takes every pending file including other
-- threads'):
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     -f - < 264_executable_checks_means_the_consumer_runs_it.sql
--   then: ./run-migrations.sh --record-only 264_...

BEGIN;

COMMENT ON COLUMN experience_patterns.executable_checks IS
  'Count of criteria checks the REGISTER''S OWN CONSUMER will actually run — today that is Tier 2 via verify_site_experience. NOT "a named tier implements the type": a Tier 4 clause is carried in deferred_checks with its reason, never counted here, because approval rests on this number (experience_patterns_approved_needs_executable_check) and an entry approved on checks nothing executes asserts nothing. Written by write_experience_pattern; the tier rule lives in one function, experienceNeedsBrowserReason, shared with the consumer so the two cannot drift.';

-- Guard: assert the post-condition inside the transaction. Proved to bite by
-- running it standalone against the database BEFORE this migration, where the
-- comment still read "a named tier can actually execute" and the DO block
-- raised. A guard never seen to fail is a guard nobody has tested.
DO $$
DECLARE
  c text;
BEGIN
  SELECT col_description('public.experience_patterns'::regclass,
                         (SELECT ordinal_position FROM information_schema.columns
                          WHERE table_schema='public' AND table_name='experience_patterns'
                            AND column_name='executable_checks')::int)
    INTO c;

  IF c IS NULL THEN
    RAISE EXCEPTION '264: executable_checks has no comment — did the column vanish?';
  END IF;
  IF c NOT LIKE '%REGISTER''S OWN CONSUMER%' THEN
    RAISE EXCEPTION '264: the column comment did not take (still: %)', left(c, 80);
  END IF;

  -- The constraint this column exists to serve must still be there. A comment
  -- change is harmless; a comment change that quietly outlives its constraint
  -- would be a column documenting a rule nothing enforces.
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'experience_patterns_approved_needs_executable_check'
      AND conrelid = 'public.experience_patterns'::regclass
  ) THEN
    RAISE EXCEPTION '264: experience_patterns_approved_needs_executable_check is gone — the number this comment describes no longer gates anything';
  END IF;
END $$;

COMMIT;

-- Verification (run after):
--   SELECT col_description('public.experience_patterns'::regclass, ordinal_position)
--   FROM information_schema.columns
--   WHERE table_name='experience_patterns' AND column_name='executable_checks';
--
-- Rollback (restores migration 230's wording verbatim):
--   COMMENT ON COLUMN experience_patterns.executable_checks IS
--     'Count of criteria checks a named tier can actually execute, written by write_experience_pattern. Approval requires at least one — see experience_patterns_approved_needs_executable_check.';
