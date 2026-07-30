-- 274_adoption_forward_fidelity_to_spawned_agent.sql
--
-- Forward `input_data.fidelity` across the adoption spawn boundary.
--
-- THE DEFECT
-- ----------
-- `082_submit_domain_unified.sh` accepts `--fidelity locked|high|medium|low|new`
-- and writes it into the submission's `input_data`. For an ADOPT submission the
-- entry agent is `site-adoption-orchestrator`, which does spawn→call→complete:
-- it spawns `site-adoption-agent` as its own pod and hands it work through
-- `call_agent`'s `input_mapping`.
--
-- `input_mapping` is an ALLOW-LIST, not a passthrough. From
-- `platform/orchestration/input_contracts/input_mapping.go`:
--
--     // InputMapping defines explicit source paths for input fields
--     // Key = destination field name (what child receives)
--     // Value = source path in CollectedData (where to get it)
--
-- The live `call_adopter` mapping enumerates exactly four fields:
--
--     "url?"                 -> input_data.url
--     "domain?"              -> input_data.domain
--     "target_url?"          -> input_data.target_url
--     "destination_domain?"  -> input_data.destination_domain
--
-- `fidelity` is not among them, so it reaches the ORCHESTRATOR and is dropped at
-- the spawn. `apply_adoption_plan` runs inside the spawned AGENT, so
-- `adoptionFidelity(params.CollectedData)` — which reads `input_data.fidelity` —
-- sees nothing and returns "". Every adoption therefore takes the recreate path
-- regardless of what the operator asked for.
--
-- WHY IT LOOKS LIKE SUCCESS, WHICH IS THE WHOLE PROBLEM
-- ----------------------------------------------------
-- Nothing errors. The flag is accepted by the script, recorded in the
-- submission, visible in `orchestration_states.collected_data` for the
-- orchestrator, and then quietly absent one hop later. The run completes, pages
-- are created, work items are queued — all of it the recreate behaviour. An
-- operator who asked to PRESERVE a site gets it REBUILT, and the only evidence
-- is the absence of a key in a child payload nobody inspects.
--
-- This was invisible until 2026-07-30 because nothing consumed `fidelity`: the
-- value was inert at both ends, so a missing hop between them changed nothing.
-- The moment `apply_adoption_plan` gained a real `locked` branch
-- (`adopt_verbatim.go`, concept register ADO-037), the missing hop became the
-- difference between preserving a site and overwriting it.
--
-- THE CHANGE
-- ----------
-- One new optional key in the `call_adopter` input_mapping. `?` marks it
-- optional, matching the four existing keys, so a submission without
-- `--fidelity` behaves exactly as before — the child simply does not receive
-- the field, `adoptionFidelity` returns "", and the recreate path runs as it
-- always has. Nothing else changes.
--
-- ORDERING: config only. The consuming code (`fidelity=locked` in
-- `apply_adoption_plan`, `deploy_mode='verbatim'` in `rerender_single_page`) is
-- already live — verified on v1.0.1211, both replicas, by grepping a string the
-- change added plus a positive control. So this migration completes a path whose
-- far end already exists; applying it cannot half-enable anything.
--
-- VERIFY AFTER APPLYING (the mapping, then the behaviour):
--   SELECT jsonb_pretty(default_config #> '{workflow,steps,call_adopter,config,input_mapping}')
--   FROM agent_definitions WHERE type='site-adoption-orchestrator' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- then, on the next locked adoption, the spawned agent's own row must carry it:
--   SELECT collected_data->'input_data'->>'fidelity'
--   FROM orchestration_states WHERE owner_agent_type='site-adoption-agent'
--   ORDER BY created_at DESC LIMIT 1;    -- expect 'locked', not NULL
--
-- ROLLBACK: the snapshot below is taken first and holds the pre-change mapping;
--   restore with
--   UPDATE agent_definitions a SET default_config = b.default_config
--   FROM agent_definitions_backup b
--   WHERE a.type='site-adoption-orchestrator' AND b.type=a.type
--     AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
--     AND b.snapshot_taken_at = (SELECT max(snapshot_taken_at) FROM agent_definitions_backup
--                                WHERE type='site-adoption-orchestrator');
--
-- LANDMINES honoured here (docs024_key_docs_latest/LANDMINES.md):
--   * `snapshot_agent(text,text)` writes to `agent_definitions_backup`; the
--     one-arg form writes an `is_snapshot=true` row into `agent_definitions`.
--     Checking the wrong table makes a good snapshot look like a no-op — so the
--     assertion below reads `agent_definitions_backup` and checks it holds the
--     PRE-change value, not merely that a row exists.
--   * `jsonb_set(..., create_if_missing := false)` on a wrong path is a SILENT
--     no-op. Here the key is new by design so `true` is correct, which means a
--     typo in the PATH would silently create a useless key instead of erroring —
--     hence the post-write assertion that the resolved mapping contains the new
--     entry, and the row-count check.

\set ON_ERROR_STOP on

BEGIN;

-- 1. Snapshot first (two-arg form -> agent_definitions_backup).
SELECT snapshot_agent(
    'site-adoption-orchestrator',
    'pre 274: forward input_data.fidelity across the adoption spawn boundary'
) AS snapshot_id;

-- 2. Assert the snapshot holds the PRE-change mapping. If this is already true,
--    an earlier run applied the change and this file is a no-op re-run.
DO $$
DECLARE
    snap_has_fidelity boolean;
BEGIN
    SELECT (default_config #>> '{workflow,steps,call_adopter,config,input_mapping}') LIKE '%fidelity%'
      INTO snap_has_fidelity
    FROM agent_definitions_backup
    WHERE type = 'site-adoption-orchestrator'
    ORDER BY snapshot_taken_at DESC
    LIMIT 1;

    IF snap_has_fidelity IS NULL THEN
        RAISE EXCEPTION '274: no snapshot row found for site-adoption-orchestrator — refusing to change a live agent with no rollback point';
    END IF;
    IF snap_has_fidelity THEN
        RAISE NOTICE '274: snapshot already contains fidelity — this appears to be a re-run; continuing (idempotent)';
    END IF;
END $$;

-- 3. Add the one optional mapping key.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_adopter,config,input_mapping,fidelity?}',
        '"input_data.fidelity"'::jsonb,
        true   -- new key by design
    ),
    updated_at = NOW()
WHERE type = 'site-adoption-orchestrator'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- 4. Assert exactly one live row changed AND the key actually landed at the
--    intended path. A mistyped path would have created a stray key with
--    create_if_missing=true and reported success.
DO $$
DECLARE
    rows_live   integer;
    has_key     boolean;
    maps_to     text;
BEGIN
    SELECT count(*),
           bool_and((default_config #> '{workflow,steps,call_adopter,config,input_mapping}') ? 'fidelity?'),
           min(default_config #>> '{workflow,steps,call_adopter,config,input_mapping,fidelity?}')
      INTO rows_live, has_key, maps_to
    FROM agent_definitions
    WHERE type = 'site-adoption-orchestrator'
      AND is_active
      AND COALESCE(is_snapshot, false) = false
      AND deleted_at IS NULL;

    IF rows_live <> 1 THEN
        RAISE EXCEPTION '274: expected exactly 1 live site-adoption-orchestrator row, found %', rows_live;
    END IF;
    IF NOT has_key THEN
        RAISE EXCEPTION '274: fidelity? key did not land in call_adopter.input_mapping — check the jsonb path';
    END IF;
    IF maps_to <> 'input_data.fidelity' THEN
        RAISE EXCEPTION '274: fidelity? maps to %, expected input_data.fidelity', maps_to;
    END IF;

    RAISE NOTICE '274: OK — call_adopter.input_mapping now forwards fidelity? -> input_data.fidelity';
END $$;

COMMIT;
