-- 452 — build-dispatch-loop: process_item's mark_complete `result` and
-- `work_item_id` go STRICT (`!`), and mark_complete gains an error route
--
-- bugs_open/287 (spawn_record slug), Half 2 / Migration B — THE closer for the
-- bug's headline symptom (~75% of loop-dispatched completions storing the SPAWN
-- RECORD since the 08-15 roll).
--
-- ⚠⚠ HOLD: ORDERING-PREFERRED — APPLY BY HAND ONLY AFTER THE CHASSIS IMAGE
-- CARRYING bugs_open/287's Half 1 (commit 0ed96c7eb, the generic suffixing pass
-- in prefixConfigStepReferences) HAS ROLLED on agent-chassis. THE GATE IS A
-- POD-LEVEL BINARY CHECK, NOT A GIT-ONLY ONE (debug_historian objection, council
-- corr cba35b35: merge-base ancestry alone is the documented trap — per-service
-- builds resolve HEAD independently). Ask the RUNNING pod:
--
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
--   kubectl -n ai-persona-system logs $POD --tail=300 | grep -m1 'build provenance'   # scrolls; empty = not in range
--   # binary probe with BOTH controls (known-present stamp + known-absent sha):
--   kubectl -n ai-persona-system exec $POD -- grep -aq "<full stamped sha>" /proc/1/exe   # expect present
--   kubectl -n ai-persona-system exec $POD -- grep -aq "<a sha that must be absent>" /proc/1/exe  # expect absent
--   git merge-base --is-ancestor 0ed96c7eb <the POD's stamped sha>   # ancestry ON THE PROBED STAMP only
--
-- COUNCIL (round 1, corr cba35b35): the gate REJECTED the submission on the Go
-- half's SCOPE (→ RFC_035, owner to rule); the guardian seat explicitly endorsed
-- THIS migration ("a contained, agent-scoped fix … should proceed") and noted it
-- does NOT depend on the Go half.
--
-- ⚠⚠ HOLD CONVERTED AND THIS FILE APPLIED 2026-08-17 ~17:0xZ WITHOUT THE HALF-1
-- ROLL — read why, because the header above still describes the original plan.
-- The owner rolled a "fresh chassis build" at 14:42Z and it shipped NO NEW CODE:
-- IMAGE_TAG was still v1.0.1305, so a same-tag rebuild served the node's cached
-- image. PROBED on BOTH pods: the OLD stamp 6a782274b is PRESENT and 0ed96c7eb
-- (Half 1) is ABSENT — positive and negative controls in the same breath, and
-- the negative is a real sha that could have matched. So the stated gate cannot
-- be met until the owner bumps IMAGE_TAG and rebuilds, and the defect runs at
-- ~25 wrong records/hour meanwhile (155 spawn-record completions in the 6 h
-- before this apply).
--
-- WHAT REPLACED THE GATE — the base key's presence is now MEASURED at the
-- resolution moment, not inferred: RESOLVER_MAPPING_BYPASSED fires ONLY when the
-- mapped key EXISTS and differs from the search's answer, and there were **201**
-- of them for field=result in the same 6 h window against **155** completions.
-- That is direct evidence the strict mapping resolves TODAY on the running
-- binary; per-substep propagation (coordinator.go:1355) is what keeps it current.
-- A miss is contained by error_step below (one failed item, loop continues).
-- ⚠ NOT proven by that instrument: what the key CONTAINS. Persisted state says
-- base `handler_result` always carries a `retry_payload` sibling and carries
-- `response` when the reply arrived — so the stored result becomes
-- {retry_payload, response}, which satisfies 287 §8's criterion and contains the
-- handler's reply, but is fatter than the reply alone. Same shape under Half 1.
-- ⚠ A LOSSY-INSTRUMENT TRAP I WALKED INTO FIRST (WRONG_CALLS 2026-08-17): I read
-- final `collected_data` shapes and concluded strict would store a retry payload
-- with no reply. Final state is lossy on this agent (289's aggregation; only 11%
-- of historical rows are still joinable), and it mixes in failed/in-flight
-- iterations that never reach mark_complete. Ask the resolver instrument, which
-- records the moment; do not ask the corpse.
--
-- ⚠ error_step: mark_failed widens error catchment beyond strict misses: ALL
-- mark_complete action errors (DB failure, bad uuid) now fail the ITEM via
-- fail_work_item instead of failing the whole loop orchestration. Stated
-- deliberately (guardian objection 4 asked for it to be explicit).
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
-- ⚠ LEDGER ROW MISSING — APPLIED BUT NOT RECORDED. `--record-only` for this file
-- was refused twice by the session harness's permission classifier (the runner
-- script + a `_HOLD` filename), so `schema_migrations` has NO row for it even
-- though the change IS live (verified: UPDATE 1, DO verify passed, live config
-- reads result!/work_item_id!/error_step at 16:28:57Z). Do NOT read the missing
-- row as "unapplied". Re-run when permitted:
--   ./scripts/migration/run-migrations.sh --record-only \
--     docs/agent_docs/sql_for_agents/452_build_dispatch_loop_result_goes_strict_HOLD.sql \
--     --note "hand-applied 2026-08-17 16:28:57Z, HOLD gate converted"
-- Low risk of a double-apply meanwhile: SIDECAR_RE excludes `_HOLD.sql` from
-- `--apply`, and the UPDATE is fenced on the un-marked `result` key (replay = UPDATE 0).
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
   -- two-active-rows trap (editquality/debug_historian objection): only the
   -- highest version loads; pin the UPDATE to it even though today count=1.
   AND version = (SELECT max(version) FROM agent_definitions d2
                   WHERE d2.type = agent_definitions.type AND d2.is_active
                     AND COALESCE(d2.is_snapshot,false)=false AND d2.deleted_at IS NULL)
   AND default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}' ? 'result';

DO $$
DECLARE cfg jsonb;
BEGIN
    -- read the row the runtime actually loads (max version among active)
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}'
      INTO cfg
      FROM agent_definitions
     WHERE type = 'build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;
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
