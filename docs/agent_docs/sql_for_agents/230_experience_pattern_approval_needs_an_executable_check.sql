-- 230_experience_pattern_approval_needs_an_executable_check.sql
--
-- Closes the one gap three council seats raised independently about the
-- experience register's substrate (council correlation
-- bbdd2c5e-1b9d-4179-a31c-a8a5c3c3bf32, APPROVED with advisory objections;
-- bug_historian and guardian each rated it, and the submission's own risk #5
-- named it):
--
--   "an entry can defer every hard check and still reach 'approved' with
--    criteria that assert nothing substantive"
--
-- The write-time validator deliberately treats a check the platform cannot
-- execute as DEFERRED rather than an error — dropping it would delete the
-- clause from the record, which is how a pattern comes to look fully checked
-- when its most important rule is not checked at all. But deferral must not be
-- free: an entry consisting ENTIRELY of deferred checks asserts nothing, and
-- before this migration nothing distinguished it from a real one at any status
-- the schema could report.
--
-- The intended answer in the submission was "a minimum-executable-checks rule
-- at the APPROVAL step". That is a rule an operator (or a future step) must
-- REMEMBER, which is a schema defect wearing a documentation costume. So it is
-- a constraint instead: an entry with no executable check CANNOT be stored as
-- approved or proven. The bad state is unrepresentable rather than merely
-- discouraged, and it holds no matter which path later writes the status.
--
-- ORDERING — unlike migration 218, this one is safe to apply before the image
-- that populates the new columns, and deliberately so:
--   * both columns have defaults, so an older binary's INSERT still succeeds;
--   * the constraint TIGHTENS (it never widens what is accepted), so it cannot
--     create the split-contract shape of 184 where the DB accepts what code
--     rejects;
--   * it constrains only 'approved' and 'proven', and no path can currently
--     write either — write_experience_pattern always writes 'draft'.
-- A tightening constraint on an unreachable state has no window in which it can
-- bite the running fleet.

BEGIN;

-- The validator's own accounting, stored on the row so that the approval
-- question is answerable from the register alone, without re-running the
-- validator or reading a log.
ALTER TABLE experience_patterns
  ADD COLUMN IF NOT EXISTS executable_checks integer NOT NULL DEFAULT 0;

ALTER TABLE experience_patterns
  ADD COLUMN IF NOT EXISTS deferred_checks jsonb NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN experience_patterns.executable_checks IS
  'Count of criteria checks a named tier can actually execute, written by write_experience_pattern. Approval requires at least one — see experience_patterns_approved_needs_executable_check.';
COMMENT ON COLUMN experience_patterns.deferred_checks IS
  'Checks beyond the platform today, each with its reason. Carried, reported, and NEVER counted as a pass.';

ALTER TABLE experience_patterns
  DROP CONSTRAINT IF EXISTS experience_patterns_approved_needs_executable_check;
ALTER TABLE experience_patterns
  ADD CONSTRAINT experience_patterns_approved_needs_executable_check
  CHECK (status = 'draft' OR executable_checks > 0);

-- Guard: assert the post-conditions inside the transaction, and prove the
-- constraint actually refuses the state it exists to refuse rather than merely
-- existing. A constraint that has never been seen to bite is a comment.
DO $guard$
DECLARE
    refused boolean := false;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                    WHERE table_name = 'experience_patterns' AND column_name = 'executable_checks') THEN
        RAISE EXCEPTION '230: executable_checks column not added';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                    WHERE table_name = 'experience_patterns' AND column_name = 'deferred_checks') THEN
        RAISE EXCEPTION '230: deferred_checks column not added';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint
                    WHERE conname = 'experience_patterns_approved_needs_executable_check') THEN
        RAISE EXCEPTION '230: the approval constraint was not created';
    END IF;

    -- Falsification, inside a subtransaction that is always undone: an
    -- all-deferred entry must be refusable as 'approved'.
    BEGIN
        INSERT INTO experience_patterns (name, kind, display_name, contract, status, executable_checks)
        VALUES ('--guard-230-probe', 'component-contract', 'guard probe',
                '{"triggers":[]}'::jsonb, 'approved', 0);
    EXCEPTION WHEN check_violation THEN
        refused := true;
    END;
    IF NOT refused THEN
        RAISE EXCEPTION '230: the constraint did NOT refuse an approved entry with zero executable checks — it does not do what it exists to do';
    END IF;

    -- And the converse: it must not refuse a legitimate approval.
    BEGIN
        INSERT INTO experience_patterns (name, kind, display_name, contract, status, executable_checks)
        VALUES ('--guard-230-probe2', 'component-contract', 'guard probe',
                '{"triggers":[]}'::jsonb, 'approved', 1);
        DELETE FROM experience_patterns WHERE name = '--guard-230-probe2';
    EXCEPTION WHEN check_violation THEN
        RAISE EXCEPTION '230: the constraint refuses a VALID approval — it is too strict';
    END;

    IF EXISTS (SELECT 1 FROM experience_patterns WHERE name LIKE '--guard-230-probe%') THEN
        RAISE EXCEPTION '230: a guard probe row survived';
    END IF;
END
$guard$;

COMMIT;

-- Verify
SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint
WHERE conname = 'experience_patterns_approved_needs_executable_check';
SELECT column_name, data_type, column_default FROM information_schema.columns
WHERE table_name = 'experience_patterns' AND column_name IN ('executable_checks','deferred_checks')
ORDER BY column_name;
SELECT count(*) AS patterns_total, count(*) FILTER (WHERE status <> 'draft') AS non_draft
FROM experience_patterns;

-- Rollback recipe (hand-run):
--   ALTER TABLE experience_patterns DROP CONSTRAINT IF EXISTS experience_patterns_approved_needs_executable_check;
--   ALTER TABLE experience_patterns DROP COLUMN IF EXISTS executable_checks;
--   ALTER TABLE experience_patterns DROP COLUMN IF EXISTS deferred_checks;
