-- 396: thunder_provision_claims — make a duplicate billable GPU unrepresentable
--
-- Closes bugs_open/259_HANDOFF_2026-08-12_one_provision_request_builds_several_billable_gpus.md
-- (resolve 259 by SLUG — a concurrent session filed a different 259 the same day).
--
-- WHAT ACTUALLY HAPPENS (corrected 2026-08-12; the bug file's original
-- mechanism was wrong and is corrected in place):
--   The dispatch_provision step carries timeout_seconds: 600. When that await
--   expires, the chassis retry driver (coordinator.go retryExpiredAwaitedRequest,
--   budget RetryVersion < 3) RE-EXECUTES the step. Each re-execution runs
--   DispatchThunderProvisionAction afresh, which mints a NEW request_id
--   (thunder_provision_dispatch.go:99) and publishes a NEW provision message.
--   The adapter's provision action has no idempotency, so every attempt builds
--   another billable GPU — up to FOUR per logical request, bounded only by the
--   retry budget.
--
--   Measured, orchestration 8c5bf926 / correlation 23c9bc6a:
--     4 awaited_requests rows, 4 DISTINCT request_ids, 1 orchestration_id,
--     each sent ~1s after the previous row's timeout_at, ending FAILED.
--   It is NOT Kafka redelivery of one message: redelivery replays identical
--   bytes, so the request_id would have been constant. It is not.
--
-- WHY correlation_id IS THE KEY:
--   Across those four attempts the ONLY stable identifier is correlation_id.
--   request_id is fresh per attempt, so keying on it would never fire — which
--   is precisely the trap, because request_id is what the dispatch code makes
--   look canonical. thunder_instances cannot serve as the dedup surface either:
--   its row is written only AFTER the box is up, so a failed attempt leaves
--   nothing behind to collide with.
--
-- WHY A SEPARATE TABLE, not a column on thunder_instances:
--   A pre-create claim row in thunder_instances would carry no real vendor id,
--   and reconcile_thunder_instances classifies any live row whose
--   thunder_instance_id is absent at the vendor as a ghost_row
--   (thunder_reconcile_action.go:204-219) — so every in-flight provision would
--   file a spurious orphan-sweep finding. That sweep is FTW-042, shipped and
--   council-approved on 2026-08-09; this migration does not disturb it.
--
-- SECOND USE: this table is also the durable record of a FAILED provision that
--   bugs_open/258 defect 3 says does not exist today (no thunder_instances row,
--   no agent_error_log row — a failed provision is currently unauditable once
--   pod logs rotate).
--
-- CONSUMER NOTICE (owner ruling 2026-07-29 §3 — tell the consumers, do not
--   merely measure them): the sole producer of provision_instance is
--   gpu-provisioner, via dispatch_thunder_provision. Its guarantee CHANGES:
--   a second provision_instance under a correlation that has already attempted
--   one is now REFUSED rather than served. That is the point — the old
--   behaviour was to build another GPU — but a workflow that relied on the
--   retry driver to re-provision after a transient failure will now fail fast
--   instead of silently spending money. No other agent_definitions row
--   dispatches this action (verified by grep over the action registry).
--
-- Rollback recipe (do not run as part of this file):
--   DROP TABLE IF EXISTS thunder_provision_claims;

BEGIN;

CREATE TABLE IF NOT EXISTS thunder_provision_claims (
    -- The idempotency key. One row per logical provision request, taken
    -- BEFORE the vendor call, so a duplicate is impossible rather than merely
    -- detected after the money is spent.
    correlation_id      text PRIMARY KEY,

    -- Audit: who asked, and under what run.
    orchestration_id    uuid,
    first_request_id    text,          -- request_id of the winning attempt
    training_run_id     uuid,
    requested_by        text,

    -- How many times a provision was ATTEMPTED under this correlation,
    -- including the refused ones. This is the number that says how hard the
    -- retry driver leaned on the door.
    attempts            integer NOT NULL DEFAULT 1 CHECK (attempts >= 1),

    -- Outcome. 'claimed' is the pre-vendor state; a row stuck in it means the
    -- adapter died between claim and create (fails closed — a later attempt is
    -- refused, and the box, if any, is caught by the FTW-042 orphan sweep).
    status              text NOT NULL DEFAULT 'claimed'
                        CHECK (status IN ('claimed','created','succeeded','failed')),

    -- Vendor identifier, written as soon as CreateInstance returns, so a
    -- crash after that point still leaves the box attributable.
    thunder_instance_id text,
    -- thunder_instances.id, on success.
    provisioning_id     uuid,
    last_error          text,

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE thunder_provision_claims IS
  'Idempotency + audit for provision_instance, keyed on correlation_id (bugs_open/259). The chassis retry driver re-dispatches an expired await with a FRESH request_id, so correlation_id is the only identifier stable across attempts. The row is taken before the vendor call: a duplicate provision is unrepresentable, not merely detected. Also the durable record of failed provisions (258 defect 3).';

COMMENT ON COLUMN thunder_provision_claims.attempts IS
  'Total provision attempts seen for this correlation, including refused ones. > 1 means the retry driver fired; the refusals are the bug staying fixed.';

-- Find the loud cases (retry driver leaning on a correlation) without a scan.
CREATE INDEX IF NOT EXISTS idx_thunder_provision_claims_repeat
    ON thunder_provision_claims (attempts, created_at)
    WHERE attempts > 1;

-- Join back to a training run when auditing spend.
CREATE INDEX IF NOT EXISTS idx_thunder_provision_claims_training_run
    ON thunder_provision_claims (training_run_id)
    WHERE training_run_id IS NOT NULL;

-- updated_at maintenance: reuse the trigger function thunder_instances already
-- uses, so there is one such function for this subsystem, not two.
DROP TRIGGER IF EXISTS trg_thunder_provision_claims_updated_at ON thunder_provision_claims;
CREATE TRIGGER trg_thunder_provision_claims_updated_at
    BEFORE UPDATE ON thunder_provision_claims
    FOR EACH ROW EXECUTE FUNCTION thunder_set_updated_at();

-- Verify (RFC_006 lesson: a block of SELECTs cannot stop a COMMIT — DO/RAISE can)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_name = 'thunder_provision_claims' AND constraint_type = 'PRIMARY KEY'
  ) THEN
    RAISE EXCEPTION '396: the PK on correlation_id IS the idempotency — it is missing';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_thunder_provision_claims_updated_at' AND NOT tgisinternal
  ) THEN
    RAISE EXCEPTION '396: updated_at trigger missing';
  END IF;

  -- Prove the constraint actually refuses a duplicate, rather than trusting
  -- that a PK named in the catalogue behaves like one. Induce it.
  BEGIN
    INSERT INTO thunder_provision_claims (correlation_id) VALUES ('__396_probe__');
    INSERT INTO thunder_provision_claims (correlation_id) VALUES ('__396_probe__');
    RAISE EXCEPTION '396: a duplicate correlation_id was ACCEPTED — the dedup does not work';
  EXCEPTION
    WHEN unique_violation THEN
      NULL;  -- expected: this is the whole point of the table
  END;
  DELETE FROM thunder_provision_claims WHERE correlation_id = '__396_probe__';
END $$;

COMMIT;
