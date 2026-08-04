-- SEED — bugs_open/173 live induction, BOTH branches
--
-- 173's own bar: "give a loop two substeps, make the tolerant one fail and
-- confirm the iteration is skipped while the orchestration continues; then make
-- the strict one fail in the same loop and confirm the orchestration FAILS.
-- Both branches, or the flag is untested in the direction that matters."
--
-- Two agents, because tolerance is static config: each isolates ONE override
-- direction, and in each the LOOP-level flag is set to the OPPOSITE of the
-- substep's. That is the whole point — if expansion still stamped the loop's
-- value over the substep's, each run would produce the other run's outcome.
--
--   test-173-tolerant-substep : loop UNSET (strict)   + substep TRUE   -> expect COMPLETED
--   test-173-strict-substep   : loop TRUE  (tolerant) + substep FALSE  -> expect FAILED
--
-- The induced fault is `SELECT 1/0` through query_database: deterministic,
-- read-only, and it cannot touch any production row.
--
-- TEMPORARY. Both rows are deleted by the DELETE at the foot of this file once
-- the induction has been witnessed — see the runbook.

INSERT INTO agent_definitions (type, display_name, description, category, default_config, is_active)
VALUES
('test-173-tolerant-substep',
 'BUG 173 induction — tolerant substep inside a STRICT loop',
 'Temporary. Loop carries no continue_on_error; the failing substep declares continue_on_error=true. Expect the iteration to be skipped and the orchestration to COMPLETE.',
 'experimental',
 '{
   "workflow": {
     "start_step": "make_items",
     "processing_mode": "orchestrator",
     "timeout_seconds": 180,
     "steps": {
       "make_items": {
         "action": "query_database",
         "description": "Two items, so the skip must advance the iteration rather than end the loop",
         "config": {"query": "SELECT 1 AS n UNION ALL SELECT 2", "output_format": "array"},
         "output_field": "items",
         "next_step": "run_loop"
       },
       "run_loop": {
         "action": "loop",
         "description": "NO loop-level continue_on_error — the loop is strict",
         "config": {
           "items_field": "items",
           "item_variable": "row",
           "sub_workflow": {
             "start_step": "boom",
             "steps": {
               "boom": {
                 "action": "query_database",
                 "description": "Induced fault; declares its OWN tolerance against a strict loop",
                 "config": {"query": "SELECT 1/0", "output_format": "array", "continue_on_error": true},
                 "output_field": "boom_out",
                 "next_step": "after"
               },
               "after": {
                 "action": "query_database",
                 "description": "Never reached on a skipped iteration",
                 "config": {"query": "SELECT 1 AS ok", "output_format": "array"},
                 "output_field": "after_out"
               }
             }
           }
         },
         "output_field": "loop_out",
         "next_step": "complete"
       },
       "complete": {
         "action": "complete_workflow",
         "description": "Reached only if the skip worked",
         "config": {"output_fields": ["loop_out"], "success_message": "173 tolerant-substep induction complete"}
       }
     }
   }
 }'::jsonb,
 true),
('test-173-strict-substep',
 'BUG 173 induction — strict substep inside a TOLERANT loop',
 'Temporary. Loop carries continue_on_error=true; the failing substep declares continue_on_error=false. Expect the orchestration to FAIL despite the tolerant loop.',
 'experimental',
 '{
   "workflow": {
     "start_step": "make_items",
     "processing_mode": "orchestrator",
     "timeout_seconds": 180,
     "steps": {
       "make_items": {
         "action": "query_database",
         "config": {"query": "SELECT 1 AS n UNION ALL SELECT 2", "output_format": "array"},
         "output_field": "items",
         "next_step": "run_loop"
       },
       "run_loop": {
         "action": "loop",
         "description": "Loop-level continue_on_error TRUE — every substep would be tolerant before this fix",
         "config": {
           "items_field": "items",
           "item_variable": "row",
           "continue_on_error": true,
           "sub_workflow": {
             "start_step": "boom",
             "steps": {
               "boom": {
                 "action": "query_database",
                 "description": "Induced fault; opts OUT of the tolerant loop",
                 "config": {"query": "SELECT 1/0", "output_format": "array", "continue_on_error": false},
                 "output_field": "boom_out",
                 "next_step": "after"
               },
               "after": {
                 "action": "query_database",
                 "config": {"query": "SELECT 1 AS ok", "output_format": "array"},
                 "output_field": "after_out"
               }
             }
           }
         },
         "output_field": "loop_out",
         "next_step": "complete"
       },
       "complete": {
         "action": "complete_workflow",
         "description": "Must NOT be reached",
         "config": {"output_fields": ["loop_out"], "success_message": "173 strict-substep induction — should never print"}
       }
     }
   }
 }'::jsonb,
 true);

-- CLEANUP, run after the induction is witnessed:
-- DELETE FROM agent_definitions WHERE type IN ('test-173-tolerant-substep','test-173-strict-substep');
