-- FILE: 681_offer_analyser_producer_register_gate_HOLD.sql
--
-- WIRES the producer register gate into offer-analyser, between the cardinal
-- attribution gate and the write.
--
-- ⚠⚠ _HOLD, AND THE HOLD IS THE WHOLE POINT. Apply this ONLY after an image
-- carrying commit f7156fb54 (or later) has rolled. `repair_ordering_register`
-- is a Go action; a seed naming an unregistered action is a FATAL run, and this
-- migration also repoints `write_offer_ordering.spec_data` at the new step's
-- output — so applying it early does not merely add an inert step, it breaks
-- the ordering write outright.
--
-- ⚠ HOW TO PROVE THE CODE IS LIVE — AND A TAG DOES NOT PROVE IT (council round 2,
-- debug_historian, medium: the precondition named a check without saying what
-- would satisfy it). A same-tag rebuild serves the node's cached binary, so
-- IMAGE_TAG and `kubectl get deploy -o jsonpath=...image` can both read new over
-- a stale process. Two acceptable proofs, in order of preference:
--
--  1. The service's own provenance line, then an ancestry test:
--       kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--         | grep -m1 'build provenance'
--       git merge-base --is-ancestor f7156fb54 <the sha that line reports>
--     ⚠ It is a STARTUP line and it SCROLLS — on a busy agent-chassis it is out
--     of reach within hours. An empty grep means "not in range", NOT "unstamped".
--
--  2. PROBE THE RUNNING BINARY FOR A SYMBOL THIS CHANGE INTRODUCED, with BOTH
--     controls in the same breath (a probe with no controls proves nothing —
--     a grep that matches everything and one that matches nothing look identical
--     when you only read the exit code):
--       POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis \
--              -o jsonpath='{.items[0].metadata.name}')
--       # MUST be present (introduced by f7156fb54):
--       kubectl -n ai-persona-system exec $POD -- grep -ac 'repair_ordering_register' /proc/1/exe
--       # MUST be absent (never existed — the negative control):
--       kubectl -n ai-persona-system exec $POD -- grep -ac 'repair_ordering_regsiter' /proc/1/exe
--     Never `strings` (absent from the debian-slim images; behind the customary
--     2>/dev/null its failure is indistinguishable from "not stamped"). Both traps
--     are in LANDMINES.md.
--
-- ⚠ AND PROBE THE POD THAT WILL RUN THIS AGENT. `-l app=<subsystem>` can select a
-- pod of a different service (one image, every label), and `logs deploy/X` reads
-- one pod of N. Check every replica before concluding.
--
-- WHY THE STEP GOES HERE AND NOT BEFORE THE CARDINAL GATE. The cardinal gate
-- DROPS points that assert a quantity absent from the field they cite. Repairing
-- the phrasing of a point that is about to be dropped spends a model call on
-- copy nobody will read, and worse, it would repair the text the drop record
-- quotes — so the audit trail would show a violation that no longer matches the
-- point it names. Cardinals first, register second.
--
-- ⚠ TWO SIMILAR-BUT-DIFFERENT LITERALS APPEAR BELOW. DO NOT SWAP THEM (council
-- round 2, editquality, low). They are an INPUT READ and an OUTPUT WRITE and they
-- belong to different steps:
--
--   'ordering_checked.object'           <- what THIS step READS.  It is the
--                                          CARDINAL gate's output_field + .object.
--   'ordering_register_checked.object'  <- what the WRITE reads.  It is THIS
--                                          step's own output_field + .object.
--
-- Swapping them silently bypasses one gate or the other while every guard below
-- still passes, which is why the verify block asserts the read side against the
-- predecessor's declared output_field rather than against a literal.
--
-- ⚠ THE CHAIN IS THREADED THROUGH `object`, AND THE REPO SQL IS STALE ON THIS.
-- Migration 408 shows `write_offer_ordering.spec_data = offer_analysis.result.ordering`.
-- The LIVE row says `ordering_checked.object` — it was repointed when the cardinal
-- gate went in, and reading the seed rather than the live row would have produced
-- a migration that silently bypassed the cardinal gate as well as this one.
-- [VERIFIED 2026-08-31 at the live agent_definitions row.]
--
-- Register: CQ-xxx (see the concept register entry filed with this change).
-- Council: submission correlation 4054f4d9-cd75-4b9c-8b8c-b7b86f11de1e.

BEGIN;

-- ── Snapshot, before anything ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS bak_ad_offeranalyser_20260831 AS
SELECT * FROM agent_definitions
 WHERE type = 'offer-analyser' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── Guards: abort if the shape this migration assumes has moved ──────────────
-- Written as DO/RAISE rather than as SELECTs. ⚠ A verify block made of SELECTs
-- CANNOT stop the COMMIT: ON_ERROR_STOP ignores a non-empty result set, so a
-- guard that "returns the bad rows" reports the fault and commits it anyway.
DO $$
DECLARE
  v_cfg     jsonb;
  v_spec    text;
  v_next    text;
  v_nrows   int;
BEGIN
  SELECT count(*) INTO v_nrows FROM agent_definitions
   WHERE type = 'offer-analyser' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF v_nrows <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 live offer-analyser definition, found %', v_nrows;
  END IF;

  SELECT default_config->'workflow'->'steps' INTO v_cfg FROM agent_definitions
   WHERE type = 'offer-analyser' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF v_cfg ? 'repair_ordering_register' THEN
    RAISE EXCEPTION 'step repair_ordering_register already exists — another session has applied this, or an equivalent. Read the live row before re-applying.';
  END IF;

  IF NOT (v_cfg ? 'verify_ordering_cardinals') THEN
    RAISE EXCEPTION 'verify_ordering_cardinals is absent — the chain this migration threads into is not there';
  END IF;

  v_next := v_cfg->'verify_ordering_cardinals'->>'next_step';
  IF v_next IS DISTINCT FROM 'write_offer_ordering' THEN
    RAISE EXCEPTION 'verify_ordering_cardinals.next_step is % (expected write_offer_ordering) — the chain has changed; re-derive the insertion point', v_next;
  END IF;

  v_spec := v_cfg->'write_offer_ordering'->'config'->>'spec_data';
  IF v_spec IS DISTINCT FROM 'ordering_checked.object' THEN
    RAISE EXCEPTION 'write_offer_ordering.spec_data is % (expected ordering_checked.object) — repointing it blindly would bypass a gate; re-derive', v_spec;
  END IF;
END $$;

-- ── 1. Insert the step ───────────────────────────────────────────────────────
-- ⚠ jsonb_set IS STRICT: any NULL argument returns NULL for the WHOLE document,
-- silently. The documented way this bites on agent_definitions.default_config is
-- a correlated subquery in the value position that finds no row — the config is
-- then NULLed and the agent is dead with no error. EVERY value below is a LITERAL
-- (jsonb_build_object / a quoted ::jsonb), no subquery anywhere in this file, so
-- that path is closed by construction rather than by care. Keep it that way: if a
-- future edit needs a looked-up value, COALESCE it and assert non-NULL before the
-- jsonb_set, or the failure is invisible until the next run of the agent.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,repair_ordering_register}',
         jsonb_build_object(
           'action', 'repair_ordering_register',
           'description',
             'H1c. Every lead_with point is scanned against BANNED_REGISTER_v1 (the owner''s banned words AND the shape family) '
             || 'before it is persisted, and a violating point is RESTATED by one judged model call. The producer minted these at '
             || '23-24% for six days and migration 667''s wash was refilled within the hour, so the gate is a producer property '
             || 'rather than a corpus state. It NEVER drops a point and NEVER fails the run: an unrepairable point keeps its '
             || 'original text and is recorded under register_repairs, because what makes the producer''s rate measurable is the '
             || 'record, not the refusal — and that rate is the only evidence this works. Runs AFTER the cardinal gate so it does '
             || 'not spend a call repairing a point that is about to be dropped. NO error_step by design, same as the write it guards.',
           'config', jsonb_build_object(
             'object_field',       'ordering_checked.object',
             'items_key',          'lead_with',
             'text_key',           'point',
             'differentiated_key', 'differentiated',
             'record_key',         'register_repairs'
           ),
           'output_field', 'ordering_register_checked',
           'next_step',    'write_offer_ordering'
         ),
         true)
 WHERE type = 'offer-analyser' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 2. Route the cardinal gate into it ───────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,verify_ordering_cardinals,next_step}',
         '"repair_ordering_register"'::jsonb,
         false)
 WHERE type = 'offer-analyser' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 3. Point the write at the repaired object ────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,write_offer_ordering,config,spec_data}',
         '"ordering_register_checked.object"'::jsonb,
         false)
 WHERE type = 'offer-analyser' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── Verify: DO/RAISE, so a failure aborts rather than prints ────────────────
DO $$
DECLARE
  v_cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO v_cfg FROM agent_definitions
   WHERE type = 'offer-analyser' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF v_cfg->'verify_ordering_cardinals'->>'next_step' <> 'repair_ordering_register' THEN
    RAISE EXCEPTION 'cardinal gate does not route into the new step';
  END IF;
  IF v_cfg->'repair_ordering_register'->>'next_step' <> 'write_offer_ordering' THEN
    RAISE EXCEPTION 'new step does not route into the write';
  END IF;
  IF v_cfg->'write_offer_ordering'->'config'->>'spec_data' <> 'ordering_register_checked.object' THEN
    RAISE EXCEPTION 'the write does not consume the repaired object';
  END IF;
  -- ⚠ THE CHAIN MUST NOT HAVE A HOLE: the new step reads what the cardinal gate
  -- WROTE. If those two names ever disagree the run does not fail — the action
  -- resolves nothing, errors on a missing object, and the ordering write is lost.
  IF v_cfg->'repair_ordering_register'->'config'->>'object_field'
     <> (v_cfg->'verify_ordering_cardinals'->>'output_field') || '.object' THEN
    RAISE EXCEPTION 'the new step does not read the cardinal gate''s own output (% vs %)',
      v_cfg->'repair_ordering_register'->'config'->>'object_field',
      v_cfg->'verify_ordering_cardinals'->>'output_field';
  END IF;

  RAISE NOTICE 'offer-analyser: producer register gate wired — cardinals -> repair_ordering_register -> write';
END $$;

COMMIT;
