-- 397: thunder_config.provision_wait_timeout_seconds — stop destroying the box
--      we just paid for, and make the deadline tunable without a build
--
-- Closes bugs_open/258 defect 2.
--
-- THE DEFECT. ProvisionAction's waitTimeout was a hardcoded 5 minutes with no
-- config path — not in thunder_config, not an env var. An a6000 does not boot
-- that fast (measured twice on 2026-08-12: 4m39s and 4m49s in, still STARTING),
-- so WaitForRunning hit its deadline, the compensating cleanup fired, and the
-- instance was DELETED at the moment it was probably about to become useful.
-- The caller got an error and paid for ~5 minutes of nothing. The compensation
-- itself was correct and worked exactly as designed; the deadline it compensated
-- for was wrong.
--
-- WHY A COLUMN AND NOT A BIGGER CONSTANT. Boot time is a vendor property, not
-- ours to guess, and it will change without telling us. A constant means a code
-- change, an image build and a whole-fleet release every time Thunder's boot
-- time drifts. This is live config: change it and the next provision uses it.
--
-- ⚠⚠ THE COUPLING THAT MAKES THIS DANGEROUS TO RAISE CARELESSLY ⚠⚠
--
--   adapter wait timeout   MUST STAY BELOW   the dispatching step's timeout_seconds
--
-- The gpu-provisioner workflow's `dispatch_provision` step awaits the response
-- for `timeout_seconds` (600 as of 2026-08-13):
--   SELECT default_config->'workflow'->'steps'->'dispatch_provision'->'config'->>'timeout_seconds'
--   FROM agent_definitions WHERE type='gpu-provisioner' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--
-- If the adapter's wait EXCEEDS that await, a slow-but-SUCCESSFUL provision is
-- the worst case, not the best: the await expires first, the chassis retry
-- driver re-executes the step, the second attempt is refused by the
-- bugs_open/259 claim guard (correctly — no duplicate box), and the workflow
-- reports FAILED while a real, running, billed instance sits there with a row.
-- Nobody is watching that box because the workflow that asked for it believes
-- it failed.
--
-- So the default here is 540s (9 minutes): comfortably above the 5 minutes that
-- demonstrably was not enough, and comfortably below the 600s await, leaving
-- ~60s for the create, keypair, secret and INSERT either side. **To go higher,
-- raise the STEP's timeout_seconds FIRST**, then this — in that order, so the
-- invariant never inverts even briefly. The CHECK below bounds this column at
-- 1800s, which is NOT a safe value on its own: it is a bound on absurdity, not
-- permission to use it.
--
-- WHY THIS IS ONLY SAFE NOW. Raising the wait makes the handler block longer,
-- and a longer block used to mean MORE duplicate billing — the 258/259
-- interaction that made this fix unsafe to apply first. bugs_open/259's claim
-- guard (migration 396) is live as of thunder-adapter v1.0.1295, so a
-- re-dispatch can no longer build a second box. Do not revert 396 and leave
-- this raised.
--
-- Rollback recipe (do not run as part of this file):
--   ALTER TABLE thunder_config DROP COLUMN provision_wait_timeout_seconds;
--   -- the adapter falls back to its compiled-in default if the column is absent

BEGIN;

ALTER TABLE thunder_config
    ADD COLUMN IF NOT EXISTS provision_wait_timeout_seconds integer NOT NULL DEFAULT 540
        CONSTRAINT thunder_config_provision_wait_bounds
        CHECK (provision_wait_timeout_seconds BETWEEN 60 AND 1800);

COMMENT ON COLUMN thunder_config.provision_wait_timeout_seconds IS
  'How long ProvisionAction waits for a new instance to reach RUNNING before compensating (deleting it). Was a hardcoded 5 min, which an a6000 cannot meet — bugs_open/258 defect 2. MUST STAY BELOW the gpu-provisioner dispatch_provision step timeout_seconds (600s): if it exceeds it, a slow SUCCESSFUL provision leaves a live billed box while the workflow reports FAILED. To raise above ~540, raise the step timeout FIRST.';

-- Verify (RFC_006: a block of SELECTs cannot stop a COMMIT — DO/RAISE can),
-- and INDUCE the constraint rather than trusting the catalogue.
DO $$
DECLARE
  v integer;
BEGIN
  SELECT provision_wait_timeout_seconds INTO v FROM thunder_config LIMIT 1;
  IF v IS NULL THEN
    RAISE EXCEPTION '397: column missing or thunder_config has no row';
  END IF;
  IF v > 600 THEN
    RAISE EXCEPTION '397: default % exceeds the 600s dispatch_provision await — see the coupling note in this file', v;
  END IF;

  -- the bound must actually refuse an absurd value
  BEGIN
    UPDATE thunder_config SET provision_wait_timeout_seconds = 5;
    RAISE EXCEPTION '397: the CHECK accepted 5 seconds — the bound does not work';
  EXCEPTION
    WHEN check_violation THEN
      NULL;  -- expected
  END;
END $$;

COMMIT;
