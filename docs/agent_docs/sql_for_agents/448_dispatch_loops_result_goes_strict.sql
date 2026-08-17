-- 448 — diagnose-dispatch-loop + report-dispatch-loop: mark_complete's `result`
-- and `work_item_id` go STRICT (`!`), and mark_complete gains an error route
--
-- bugs_open/287 (spawn_record slug), Half 2 / Migration A. These two agents are
-- NOT loops (top-level steps, no sub_workflow) — 287 §6a's own 090 work item is
-- a live instance of the defect on diagnose-dispatch-loop: the item completed
-- with the SPAWN RECORD as its result while the verdict sat in the child's row.
--
-- MECHANISM (verified at HEAD 2026-08-17, bugfix_287_spawn_record/PLAN):
-- `"result": "handler_result"` is a DOTLESS mapping. ExtractActionInputs'
-- whole-tree search runs before the single-segment mapping (Strategy 4) and
-- resolves the FIELD NAME `result` to `handler_spawned.result` — the spawn-ack
-- payload — every time (RESOLVER_MAPPING_BYPASSED, RFC_029 §10.5/10.6). The `!`
-- marker (RFC_029 §9 D3, CTS-060, parser live in the fleet since v1.0.1303) is
-- the sanctioned remedy: explicit resolution only, loud failure on absence.
--
-- MEASURED before writing (2026-08-17, live DB): 3/3 retained COMPLETED
-- diagnose-dispatch-loop orchestrations hold the full reply envelope at exactly
-- `collected_data.handler_result` (call_handler's un-suffixed output_field), so
-- the strict mapping resolves on the happy path. report-dispatch-loop had zero
-- retained rows in the window (no traffic) — same config shape, same reasoning.
--
-- `work_item_id` goes strict for the same reason: it resolves explicitly today
-- (`claimed.work_item_id`, dotted, Strategy 0); `!` only forbids the silent
-- whole-tree fallback, whose winner would be a DIFFERENT item's id (the
-- wrong-item-completion door).
--
-- `error_step: mark_failed` is added so a strict miss fails THIS item (via
-- fail_work_item) instead of failing the orchestration and stranding the item —
-- the adversarial-review finding G1 on this bug's plan. Config-level error_step
-- is the live pattern on these agents (their call_handler already carries one).
--
-- PRE-APPLY CHECKS (both run 2026-08-17; re-run before applying):
--   (1) two-active-rows trap (LANDMINES): both types carry exactly ONE active,
--       non-snapshot, non-deleted row (version 1):
--         SELECT type, count(*) FROM agent_definitions
--          WHERE type IN ('diagnose-dispatch-loop','report-dispatch-loop')
--            AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
--          GROUP BY type;   -- expect 1 and 1
--   (2) the running chassis carries the `!` parser: stamp is v1.0.1305+
--       (binary probe with a known-present AND known-absent sha; RFC_029 §10.6
--       probed 6a782274b on 2026-08-17).
--
-- Idempotent: the UPDATE is fenced on the un-marked `result` key still being
-- present; the doc_notes INSERT is fenced too (the 417 re-run lesson —
-- "idempotent" that covers only the UPDATE is not idempotent).
-- snapshot_agent is the TWO-ARG overload (writes agent_definitions_backup).

BEGIN;

SELECT snapshot_agent('diagnose-dispatch-loop',
    '448_dispatch_loops_result_goes_strict.sql: pre-update');
SELECT snapshot_agent('report-dispatch-loop',
    '448_dispatch_loops_result_goes_strict.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             (default_config
                #- '{workflow,steps,mark_complete,config,result}'
                #- '{workflow,steps,mark_complete,config,work_item_id}'),
             '{workflow,steps,mark_complete,config,result!}',
             COALESCE(default_config #> '{workflow,steps,mark_complete,config,result}',
                      '"handler_result"'::jsonb),
             true),
           '{workflow,steps,mark_complete,config,work_item_id!}',
           COALESCE(default_config #> '{workflow,steps,mark_complete,config,work_item_id}',
                    '"claimed.work_item_id"'::jsonb),
           true),
         '{workflow,steps,mark_complete,config,error_step}',
         '"mark_failed"'::jsonb,
         true),
       updated_at = now()
 WHERE type IN ('diagnose-dispatch-loop','report-dispatch-loop')
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,mark_complete,config}' ? 'result';

-- Verify (DO/RAISE — a SELECT verify cannot stop the COMMIT): each of the two
-- live rows carries result!/work_item_id!/error_step and neither old spelling.
DO $$
DECLARE r record; cfg jsonb; n int := 0;
BEGIN
    FOR r IN SELECT type, default_config FROM agent_definitions
              WHERE type IN ('diagnose-dispatch-loop','report-dispatch-loop')
                AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    LOOP
        n := n + 1;
        cfg := r.default_config #> '{workflow,steps,mark_complete,config}';
        IF cfg IS NULL THEN
            RAISE EXCEPTION '448: % has no mark_complete config', r.type;
        END IF;
        IF NOT (cfg ? 'result!') OR (cfg ? 'result') THEN
            RAISE EXCEPTION '448: % result spelling wrong after update — config is %', r.type, cfg;
        END IF;
        IF NOT (cfg ? 'work_item_id!') OR (cfg ? 'work_item_id') THEN
            RAISE EXCEPTION '448: % work_item_id spelling wrong after update — config is %', r.type, cfg;
        END IF;
        IF cfg->>'result!' <> 'handler_result' THEN
            RAISE EXCEPTION '448: % result! value unexpected: %', r.type, cfg->>'result!';
        END IF;
        IF cfg->>'error_step' <> 'mark_failed' THEN
            RAISE EXCEPTION '448: % error_step missing/unexpected: %', r.type, cfg->>'error_step';
        END IF;
    END LOOP;
    IF n <> 2 THEN
        RAISE EXCEPTION '448: expected 2 live rows, saw %', n;
    END IF;
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
SELECT 'pipeline', 'build',
$note$## 448: diagnose/report-dispatch-loop mark_complete goes STRICT — bugs_open/287 (spawn_record), Migration A
`result!: handler_result` + `work_item_id!: claimed.work_item_id` + `error_step: mark_failed` on both agents' mark_complete. The `!` marker (RFC_029 §9 D3) means explicit resolution only: the item's stored result is now the handler's reply envelope or a loud per-item failure — never the whole-tree search's guess (which deterministically returned the SPAWN RECORD, `handler_spawned.result`). Until the sibling migration 452 (HOLD) is applied after the next chassis roll, build-dispatch-loop still has the defect — read verdicts from orchestration_states, not the item, for that agent.
Categories: migration$note$,
'["migration"]'::jsonb, 'agent', 'bugfix-287-spawn-record-lane'
WHERE NOT EXISTS (
    SELECT 1 FROM doc_notes
    WHERE body LIKE '## 448: diagnose/report-dispatch-loop mark_complete goes STRICT%'
);

COMMIT;
