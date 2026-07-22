-- ============================================================================
-- MIGRATION 191: raise the diagnose-agent pod resource REQUESTS
-- ============================================================================
-- WHY (bugs_open/043 "diagnosis runs hang at the route step"): diagnose-agent is
-- spawned as a per-run Kubernetes Job whose pod resources come from THIS row's
-- `resources` column (spawn_actions.go parseResourceSpec → the Job pod spec).
-- Until now diagnose-agent carried NO explicit resources, so it inherited the
-- agent_definitions table DEFAULT — requests cpu 100m / memory 256Mi. Diagnose
-- runs are the fleet's heaviest single job (clone + analyse the whole repo, then
-- up to 5×[gather → LLM verdict → route], 100–535 s in ONE ephemeral pod), and
-- under burst load the diagnosis pods were dying mid-run and stranding their
-- orchestrations (an instance of the bugs_open/003 spawn-loss / at-most-once-
-- consume class; the 4h EXECUTING_STEP reaper is what FAILs them). A pod is an
-- eviction candidate under node pressure precisely when it exceeds its REQUESTS,
-- so the low 100m/256Mi request made diagnose-agent — a long-lived, analysis-
-- heavy pod — a disproportionately likely victim.
--
-- This raises the REQUESTS toward its real footprint so it is far less likely to
-- be evicted (limits unchanged — memory peak is ~150Mi, well under the 1Gi cap;
-- OOM was ruled out). It does NOT change the table default (111 agents share it);
-- it sets diagnose-agent's row EXPLICITLY.
--
-- This is NOT the root fix — that is bugs_open/003's at-least-once / retry work
-- (F2/F3), which re-drives a lost run instead of stranding it. This is the
-- eviction-exposure mitigation recommended in the 043 diagnosis.
--
-- LIVE ON APPLY: the spawner reads this row at spawn time, so the new requests
-- take effect on the NEXT diagnose run — no image roll, no chassis restart.
-- Idempotent: re-running sets the same value.
-- ============================================================================

BEGIN;

UPDATE agent_definitions
SET resources = jsonb_build_object(
        'requests', jsonb_build_object('cpu', '250m', 'memory', '512Mi'),
        'limits',   jsonb_build_object('cpu', '500m', 'memory', '1Gi')
    ),
    updated_at = now()
WHERE type = 'diagnose-agent'
  AND deleted_at IS NULL
  AND COALESCE(is_snapshot, false) = false;

COMMIT;

-- Verify (expect requests cpu 250m / memory 512Mi):
--   SELECT type, resources->'requests' AS requests, resources->'limits' AS limits
--   FROM agent_definitions
--   WHERE type='diagnose-agent' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
