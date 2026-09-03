-- 736 — orchestration_status_vocabulary.RUNNING.written_by names a function that
--       no longer exists, as of bugs_open/329's fix (b55f837ef, register WFA-025).
--
-- WHY THIS EXISTS. Migration 466 made orchestration_states.status a single source
-- of truth and seeded a `written_by` column recording, for each status, the Go
-- symbol that writes it. RUNNING's row named
-- `StateRepository.ClearExecutingStep (state.go:1428)`. That function is DELETED:
-- 329's fix folds the clear into the claim, so RUNNING is now written inside
-- StateRepository.ClaimStaleOrchestration's mutator, under the version CAS.
--
-- HOW IT WAS FOUND, because the class matters more than this row. Two council
-- seats (guardian [low], prior_art_librarian [medium], corr 3beb3f54) objected
-- that deleting an exported method rested on a repo-wide grep and asked what a
-- caller invisible to the compiler would look like. `go build ./...` is clean and
-- there is no reflection on StateRepository — but grepping OUTSIDE .go found this:
-- a live DATABASE row, seeded by a migration, naming the deleted symbol. It could
-- never break a build, and nothing would ever have flagged it.
--
-- ⚠ THE ROW IS DOCUMENTATION, NOT A DISPATCH TARGET. Nothing reads `written_by` to
-- decide anything — the reaper and database-cleanup read is_terminal/is_pausable,
-- which this migration does not touch. So this is a correctness fix to a reference
-- an engineer will trust, not a behavioural change. It is applied rather than left
-- because a stale `written_by` is exactly the "seed is not the system" trap in
-- reverse: here the live row IS the record, and it is now wrong.
--
-- SAFE TO APPLY BEFORE OR AFTER the chassis roll: it describes code, and it is
-- true the moment b55f837ef is committed (it is), regardless of which image runs.

BEGIN;

UPDATE orchestration_status_vocabulary
   SET written_by = 'StateRepository.ClaimStaleOrchestration (state.go, the EXECUTING_STEP arm of the claim)',
       notes      = 'The inter-step gap. Written inside the stale-takeover CLAIM, under the version CAS '
                    || '(bugs_open/329, register WFA-025) — previously by ClearExecutingStep, which was a '
                    || 'separate unguarded read-modify-write and is deleted. Also the status a healthy row '
                    || 'occupies for milliseconds between steps; bugs_closed/294.',
       updated_at = now()
 WHERE status = 'RUNNING';

-- Prove the UPDATE did what it says. A verify block of bare SELECTs cannot stop a
-- COMMIT (ON_ERROR_STOP ignores a non-empty result) — so RAISE, and make the
-- failure induceable by changing the literal below.
DO $do$
DECLARE got text;
BEGIN
  SELECT written_by INTO got FROM orchestration_status_vocabulary WHERE status = 'RUNNING';
  IF got IS NULL THEN
    RAISE EXCEPTION '736: no RUNNING row in orchestration_status_vocabulary — migration 466 not applied?';
  END IF;
  IF got LIKE '%ClearExecutingStep%' THEN
    RAISE EXCEPTION '736: RUNNING.written_by still names the deleted ClearExecutingStep: %', got;
  END IF;
  IF got NOT LIKE '%ClaimStaleOrchestration%' THEN
    RAISE EXCEPTION '736: RUNNING.written_by did not take the new value, got: %', got;
  END IF;
  RAISE NOTICE '736 OK: RUNNING.written_by = %', got;
END
$do$;

COMMIT;
