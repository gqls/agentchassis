-- 452 — build-dispatch-loop: process_item's mark_complete `result` and
-- `work_item_id` go STRICT (`!`), and mark_complete gains an error route
--
-- bugs_open/287 (spawn_record slug), Half 2 / Migration B — THE closer for the
-- bug's headline symptom (~75% of loop-dispatched completions storing the SPAWN
-- RECORD since the 08-15 roll).
--
-- ⚠⚠ HOLD: ORDERING-PREFERRED — APPLY BY HAND ONLY AFTER THE CHASSIS IMAGE
-- CARRYING bugs_open/287's Half 1 (the generic suffixing pass in
-- prefixConfigStepReferences — commit subject "287 …generic reference-shaped
-- suffixing…") HAS ROLLED on agent-chassis. Verify:
--
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the Half 1 commit> <the stamped sha>
--
-- WHY THE ORDER (and why it is "preferred", not load-bearing like 417's):
-- with Half 1 live, loop expansion rewrites `result!: handler_result` to
-- `result!: handler_result_N`, so the strict mapping resolves the reply's own
-- iteration-suffixed key DIRECTLY. Without Half 1 it resolves the base
-- `handler_result` — which IS present and current at mark_complete time
-- (setLoopVariable -> propagateIterationOutputs runs before every substep,
-- coordinator.go:1355; the 279 RESOLVER_MAPPING_BYPASSED rows/day prove the
-- mapped key exists — 287 §11's correction of §9 fact 2/§10) — but only via the
-- propagation side-channel, whose silent early-returns (missing loop_metadata /
-- item key) would then surface as strict failures. Suffix-first removes that
-- coupling and matches the RFC_029 lane's published fix-then-ratchet sequencing.
-- The `!` PARSER itself has been live since v1.0.1303; that is not the gate here.
--
-- MEASURED before writing (2026-08-17, live DB; queries in
-- bugfix_287_spawn_record/RUNBOOK): 176 conflict rows/day for field=result,
-- winner always `handler_spawned.result`; 279 bypass rows/day on the dotless
-- `handler_result` mapping; item census 4 spawn-record vs 1 own-envelope
-- completions in the 08:00Z hour today. After this flip those go to zero while
-- the loop keeps running, or fail LOUDLY per item — never a silent guess.
--
-- `error_step: mark_failed` (adversarial-review finding G1): without it a
-- strict miss fails the WHOLE loop orchestration (mark_complete has no error
-- route; continue_on_error=false) and strands the item in 'claimed', which no
-- code path reclaims. mark_failed is an in-loop substep; expansion prefixes
-- config-level error_step correctly (error_step_loop_expansion_test.go).
--
-- PRE-APPLY CHECKS (run 2026-08-17; re-run before applying):
--   (1) two-active-rows trap: build-dispatch-loop has exactly ONE active,
--       non-snapshot, non-deleted row (version 1):
--         SELECT count(*) FROM agent_definitions
--          WHERE type='build-dispatch-loop' AND is_active
--            AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;  -- expect 1
--   (2) the binary gate above (Half 1 ancestor of the agent-chassis stamp).
--   (3) no orchestration dispatch within ~300s of a chassis pod (re)start.
--
-- AFTER APPLYING, VERIFY AT THE INSTRUMENT (bugfix_287_spawn_record/RUNBOOK):
-- RESOLVER_% rows with context->>'field'='result' for build-dispatch-loop -> 0
-- WHILE the loop has traffic; item census spawn_record -> 0 while own_envelope
-- rises. work_item_id/current_page conflict rows are NOT this bug's metric.
--
-- Idempotent: UPDATE fenced on the un-marked `result` key; doc_notes fenced.
-- snapshot_agent is the TWO-ARG overload (writes agent_definitions_backup).

BEGIN;

SELECT snapshot_agent('build-dispatch-loop',
    '452_build_dispatch_loop_result_goes_strict_HOLD.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             (default_config
                #- '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,result}'
                #- '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,work_item_id}'),
             '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,result!}',
             COALESCE(default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,result}',
                      '"handler_result"'::jsonb),
             true),
           '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,work_item_id!}',
           COALESCE(default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,work_item_id}',
                    '"current_item.id"'::jsonb),
           true),
         '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,error_step}',
         '"mark_failed"'::jsonb,
         true),
       updated_at = now()
 WHERE type = 'build-dispatch-loop'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}' ? 'result';

DO $$
DECLARE cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}'
      INTO cfg
      FROM agent_definitions
     WHERE type = 'build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '452: no live build-dispatch-loop mark_complete config';
    END IF;
    IF NOT (cfg ? 'result!') OR (cfg ? 'result') THEN
        RAISE EXCEPTION '452: result spelling wrong after update — config is %', cfg;
    END IF;
    IF NOT (cfg ? 'work_item_id!') OR (cfg ? 'work_item_id') THEN
        RAISE EXCEPTION '452: work_item_id spelling wrong after update — config is %', cfg;
    END IF;
    IF cfg->>'result!' <> 'handler_result' THEN
        RAISE EXCEPTION '452: result! value unexpected: %', cfg->>'result!';
    END IF;
    IF cfg->>'work_item_id!' <> 'current_item.id' THEN
        RAISE EXCEPTION '452: work_item_id! value unexpected: %', cfg->>'work_item_id!';
    END IF;
    IF cfg->>'error_step' <> 'mark_failed' THEN
        RAISE EXCEPTION '452: error_step missing/unexpected: %', cfg->>'error_step';
    END IF;
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
SELECT 'pipeline', 'build',
$note$## 452: build-dispatch-loop mark_complete goes STRICT — bugs_open/287 (spawn_record), Migration B
`result!: handler_result` + `work_item_id!: current_item.id` + `error_step: mark_failed` on process_item's mark_complete. With Half 1 (generic loop-expansion suffixing) live, the strict mapping expands to the iteration-suffixed reply key and resolves it directly; the whole-tree search — whose deterministic winner was the SPAWN RECORD (`handler_spawned.result`, 176x/day) — never runs for these fields. A genuinely absent reply now fails THAT item loudly (error route: mark_failed) instead of storing a foreign payload. Watch it land: RESOLVER_% rows field='result' for build-dispatch-loop -> 0 while the loop has traffic; site_work_items spawn-record census -> 0.
Categories: migration$note$,
'["migration"]'::jsonb, 'agent', 'bugfix-287-spawn-record-lane'
WHERE NOT EXISTS (
    SELECT 1 FROM doc_notes
    WHERE body LIKE '## 452: build-dispatch-loop mark_complete goes STRICT%'
);

COMMIT;
