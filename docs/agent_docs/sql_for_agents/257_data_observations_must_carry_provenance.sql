-- 257 — business_intel.data_observations: an observation that cannot say where it
--       came from cannot be stored.
--
-- bugs_open/100, fix candidate 1: "make the unsourced state unrepresentable at the
-- write". The Go half (provenance taken from the fetch record rather than from the
-- model's own output) is necessary but not sufficient — it leaves the old failure
-- mode reachable by any future caller that omits it, which is exactly how 2,970
-- rows came to be stored with no source, in silence, over four months.
--
-- WHY A CHECK AND NOT "NOT NULL"
-- source_type is ALREADY declared NOT NULL, and every one of those 2,970 rows
-- satisfied it — because the Go read produced an empty STRING, not a NULL. The
-- constraint that was already there did not fire once. NOT NULL on source_url would
-- be the same non-event; the empty string is the bad value, so the empty string is
-- what has to be refused.
--
-- WHY "NOT VALID"
-- It enforces on every INSERT and UPDATE from the moment it is applied, but does not
-- re-check the existing rows. That is deliberate: the 2,970 historical rows are
-- genuinely unsourced and genuinely unpublishable under our own rule, and deleting
-- or back-filling them would either destroy evidence or invent provenance — the
-- second being precisely the defect this bug is about. They stay, refused by the
-- publishing rule rather than hidden by a data fix.
--
-- BLAST RADIUS — checked before writing this, not assumed.
-- One writer in the entire tree inserts into this table:
-- platform/orchestration/actions/business_intel_actions.go (StoreBusinessVerificationAction),
-- which as of this change always takes provenance from the fetch record and logs
-- loudly when it cannot. There is no other insert path to break.
--
-- SEQUENCING: apply this AFTER the chassis image carrying the Go fix is live. Applied
-- before, it would refuse writes the running binary still cannot satisfy — turning a
-- silent data-quality defect into a hard failure of vet verification. The Go change is
-- what makes this constraint satisfiable; this constraint is what keeps it true.

BEGIN;

ALTER TABLE business_intel.data_observations
    DROP CONSTRAINT IF EXISTS data_observations_provenance_not_empty;

ALTER TABLE business_intel.data_observations
    ADD CONSTRAINT data_observations_provenance_not_empty
    CHECK (
        COALESCE(source_url, '')  <> ''
        AND COALESCE(source_type, '') <> ''
    )
    NOT VALID;

COMMENT ON CONSTRAINT data_observations_provenance_not_empty
    ON business_intel.data_observations IS
    'bugs_open/100: provenance is recorded by the component that performed the fetch, '
    'never reported by the model that read the result. NOT VALID so the 2,970 '
    'pre-fix rows are left as they are — unsourced and unpublishable — rather than '
    'back-filled with invented provenance.';

COMMIT;

-- VERIFY (expect: the constraint exists and is NOT VALID; historical rows untouched)
--
--   SELECT conname, convalidated
--   FROM pg_constraint
--   WHERE conrelid = 'business_intel.data_observations'::regclass
--     AND conname = 'data_observations_provenance_not_empty';
--   -- data_observations_provenance_not_empty | f      <- f = NOT VALID, as intended
--
-- NEGATIVE CONTROL — the constraint is only meaningful if it actually refuses.
-- Without this step "quiet" and "not enforcing" look identical:
--
--   BEGIN;
--   INSERT INTO business_intel.data_observations (business_id, source_type, raw_data)
--   VALUES (NULL, '', '{}'::jsonb);
--   -- expect: ERROR ... violates check constraint "data_observations_provenance_not_empty"
--   ROLLBACK;
