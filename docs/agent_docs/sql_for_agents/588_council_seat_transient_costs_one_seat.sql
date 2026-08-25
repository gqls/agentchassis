-- 588_council_seat_transient_costs_one_seat.sql
--
-- ✅ APPLIED 2026-08-25 by hand, gate met, and RENAMED off `_HOLD` at that moment.
-- The rename is the point, not tidiness: inside the repo an applied `_HOLD` file is
-- indistinguishable from one still waiting, and the FILENAME is the only tell — a stale
-- `_HOLD` twin instructs the next reader to hold back something already live.
--
-- bugs_open/243-anthropic-cap. Council-Reviewed: 82f07fa6-1c42-46ad-bdf6-1d58892c44a7
--
-- WHAT: a council review seat whose LLM call errors costs ONE SEAT instead of the
-- whole round.
--
-- ⚠ AMENDED 2026-08-24: this file ORIGINALLY also cut the claude re-probe interval
-- from 3600s to 60s. That half was SPLIT OUT into
-- `596_claude_probe_interval_60s.sql` and APPLIED on the owner's instruction the
-- same day, because it has no dependency on the Go writer this file waits for, and
-- holding it here would have left a measured 60-minute fleet stall in place for no
-- reason. Do NOT re-add it: applying it twice is harmless but the ledger would then
-- carry two files claiming the same change, and 596 is the one that is recorded.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- ⚠ WAS _HOLD. The gate below WAS MET and this IS APPLIED — kept verbatim because it is the
--   record of what had to be true first, and because the same shape recurs.
--   Gate result 2026-08-25 on chassis v1.0.1337, BOTH replicas:
--     `step-error record capped at` = 1 (writer present)   <- the discriminating probe
--     `__step_errors`               = 1 (reader ALSO says this — NOT evidence)
--     known-present control         = 15, known-absent control = 0
--   Plus ancestry: `git merge-base --is-ancestor dbd865ee8 4c996e1b5` (the running binary's
--   own provenance stamp) => IN.
--
-- Why: repointing error_step makes a failed seat's result field merely ABSENT,
-- and diagnose_council_decide counts an absent field as an ABSTENTION. Its own
-- comment (diagnose_council_decide_action.go, the raw == nil branch) says why
-- that is wrong: "an abstention is a seat the relevance filter skipped, which is
-- information; an unreadable seat is an opinion we were owed and lost...
-- Conflating them would let a lost opinion read as a considered non-objection."
-- An `unreadable` seat DOWNGRADES an approval to revise; an abstention does not.
--
-- So applying this BEFORE the Go half is live does not merely fail to help — it
-- converts a LOUD failure (the round dies, you resubmit) into a SILENT one (the
-- round approves with a seat nobody heard from). That is a worse state than the
-- bug this fixes.
--
-- THE GATE, and run BOTH controls in the same breath — a grep that matches
-- everything proves nothing (LANDMINES: "a control that matches everything"):
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--   kubectl -n ai-persona-system exec $POD -- grep -ac __step_errors /proc/1/exe   # MUST be >= 1
--   kubectl -n ai-persona-system exec $POD -- grep -ac __step_errors_absent_control /proc/1/exe  # MUST be 0
-- Repeat for EVERY replica, not one: a partial roll leaves the old binary serving
-- some rounds. Then apply this file by hand and record it with --record-only.
-- ─────────────────────────────────────────────────────────────────────────────

BEGIN;

-- Convention this shop runs on and which the FIRST cut of this file omitted (caught on the
-- 2026-08-25 apply, not by the council — a reviewer reads a sketch, not executable SQL):
-- snapshot_agent opens every migration that touches agent_definitions, so the pre-change
-- config is recoverable without a git archaeology session.
SELECT snapshot_agent('council-gate', '588_council_seat_transient_costs_one_seat: pre-update');

-- Each seat's error_step becomes that seat's OWN next_step, so a failed seat
-- is skipped and the chain continues to the next reviewer.
--
-- The filter is on error_step='complete_invalid' with the two terminals named as
-- exceptions, NOT on a `review_%` name pattern. That is deliberate and it is the
-- editquality seat's objection on this round: a name filter would silently leave
-- a future gate_* seat routing to complete_invalid, which is the exact failure
-- this migration exists to close. Measured 2026-08-24: the 19 steps carrying
-- error_step='complete_invalid' are the 17 review_* seats plus exactly these two.
--
-- persist_submission and council_decide KEEP complete_invalid on purpose: if the
-- submission cannot be persisted there is nothing to review, and if aggregation
-- fails there is no verdict. Neither is a reviewer's opinion.
-- ⚠ REWRITTEN 2026-08-25. The first cut used
--     UPDATE agent_definitions ad ... FROM LATERAL jsonb_each(ad.default_config ...)
-- which Postgres refuses: "invalid reference to FROM-clause entry for table ad" — the UPDATE
-- target cannot be referenced from a LATERAL in its own FROM. It failed at apply, inside the
-- transaction, and rolled back with nothing changed. A CORRELATED SCALAR SUBQUERY may
-- reference `ad`, so the whole steps object is rebuilt in one pass instead.
-- The rule being applied is unchanged from the reviewed version; only the SQL that expresses
-- it. Note the council reviewed a SKETCH, and a sketch is not executable — a reviewer cannot
-- catch this class and should not be expected to.
UPDATE agent_definitions ad
SET default_config = jsonb_set(
      ad.default_config,
      '{workflow,steps}',
      (
        SELECT jsonb_object_agg(
                 s.k,
                 CASE
                   WHEN s.v->'config'->>'error_step' = 'complete_invalid'
                    AND s.k NOT IN ('persist_submission', 'council_decide')
                    AND COALESCE(s.v->>'next_step', '') <> ''
                   THEN jsonb_set(s.v, '{config,error_step}', to_jsonb(s.v->>'next_step'))
                   ELSE s.v
                 END)
        FROM jsonb_each(ad.default_config #> '{workflow,steps}') AS s(k, v)
      )
    ),
    updated_at = NOW()
WHERE ad.type = 'council-gate'
  AND ad.is_active
  AND COALESCE(ad.is_snapshot, false) = false
  AND ad.deleted_at IS NULL;

-- Verify INSIDE the transaction. A block of bare SELECTs cannot stop a COMMIT —
-- ON_ERROR_STOP ignores a non-empty result — so this must RAISE.
DO $$
DECLARE
  still_invalid int;
BEGIN
  SELECT count(*) INTO still_invalid
    FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k, v)
   WHERE type = 'council-gate' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND v->'config'->>'error_step' = 'complete_invalid'
     AND k NOT IN ('persist_submission', 'council_decide');
  IF still_invalid > 0 THEN
    RAISE EXCEPTION '% reviewer seat(s) still route error_step to complete_invalid', still_invalid;
  END IF;

  -- The two terminals must NOT have been swept up.
  IF (SELECT count(*) FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k, v)
       WHERE type = 'council-gate' AND is_active
         AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
         AND k IN ('persist_submission', 'council_decide')
         AND v->'config'->>'error_step' = 'complete_invalid') <> 2 THEN
    RAISE EXCEPTION 'persist_submission/council_decide must KEEP complete_invalid — one of them was rewritten';
  END IF;
END $$;

COMMIT;
