-- STAGED_tool_auditor_review_handler_to_empty.sql — bugs_open/291, Phase 3.
--
-- ⚠⚠ DO NOT MOVE THIS FILE INTO sql_for_agents/ OR APPLY IT UNTIL THE ORDERING
-- GATE BELOW PASSES. It is staged HERE, outside the migration runner's
-- directory, precisely so another session's `run-migrations.sh --apply` cannot
-- pick it up early. Applied against the OLD binary it does not error loudly —
-- it makes every tool-auditor review-item filing hard-error inside a
-- continue_on_error loop, i.e. EVERY FINDING SILENTLY LOST, worse than the bug.
--
-- ORDERING GATE (the 442-style precondition; all three must hold):
--   1. The chassis roll carrying commit c8400e452 ("291 phase 2: the write door
--      demotes...") is LIVE on agent-chassis. Verify per SERVICE at the artefact:
--        kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--          | grep -m1 'build provenance'      # startup line; scrolls — see RUNBOOK
--        git merge-base --is-ancestor c8400e452 <stamped-sha> && echo SHIPPED
--      (fallback: known-value binary probe with a positive AND negative control —
--      LANDMINES.) That roll relaxes create_work_item's empty-handler validation
--      (create_work_item_action.go): empty handler is legal for parked items.
--   2. Migration 450 is applied (the status key exists — the gate below asserts it).
--   3. Re-check the next free migration number in sql_for_agents/ (450/451 were
--      this lane's; the directory moves daily), rename this file to that number,
--      move it in, apply by hand, then --record-only. Write the matching
--      _ROLLBACK.sql (jsonb_set the path back to '"hitl-review"').
--
-- WHAT IT DOES: flips tool-auditor's create_review_item.config.handler_agent
-- from the inert phantom 'hitl-review' to the canonical '' (migration 217 idiom).
-- With '', CHECK swi_no_handlerless_promotable (443) also structurally refuses
-- any promotion of a parked review item into the dispatch queue — including the
-- admin Retry path (site_admin_handlers.go), which until this flip recycles a
-- parked review item back to blocked (non-corrupting, but noisy).

BEGIN;

SELECT snapshot_agent('tool-auditor',
  'STAGED_tool_auditor_review_handler_to_empty (bugs_open/291 phase 3): pre-update');

DO $$
DECLARE
  n int;
  h text;
  s text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
     AND type = 'tool-auditor'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '291p3: expected exactly 1 live tool-auditor row at the pinned id, found %', n;
  END IF;

  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}',
         default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,status}'
    INTO h, s
    FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';

  IF s IS DISTINCT FROM 'needs_human_review' THEN
    RAISE EXCEPTION '291p3: status key is % — migration 450 must be applied first; refusing', s;
  END IF;
  IF h IS DISTINCT FROM 'hitl-review' THEN
    RAISE EXCEPTION '291p3: handler_agent is % — pre-state has moved; re-read before applying', h;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}',
         '""'),
       updated_at = now()
 WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
   AND type = 'tool-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  h text;
  s text;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}',
         default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,status}'
    INTO h, s
    FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';
  IF h IS DISTINCT FROM '' OR s IS DISTINCT FROM 'needs_human_review' THEN
    RAISE EXCEPTION '291p3: post-state wrong (handler=%, status=%)', h, s;
  END IF;
END $$;

-- Sweep the parked ROWS still carrying the phantom (measured 2026-08-17: exactly
-- one — a needs_new_layout_candidate from site-design-planner, 2026-08-12, which
-- also proves resolve_composition_layout_action.go is live and reachable; the
-- Go flip to '' in c8400e452 stops new ones at the same roll this file waits
-- for). Inert where they sit, but admin Retry would recycle them to blocked;
-- with '' CHECK 443 refuses the Retry instead. Scoped to PARKED rows only.
UPDATE site_work_items
   SET handler_agent = '',
       result = COALESCE(result, '{}'::jsonb) || jsonb_build_object('handler_291',
         jsonb_build_object('swept_at', now()::text, 'from_handler', 'hitl-review')),
       updated_at = now()
 WHERE handler_agent = 'hitl-review'
   AND status = 'needs_human_review';

DO $$
DECLARE
  leftover int;
BEGIN
  SELECT count(*) INTO leftover FROM site_work_items WHERE handler_agent = 'hitl-review';
  IF leftover > 0 THEN
    RAISE EXCEPTION '291p3: % row(s) still carry hitl-review at a NON-parked status — do not widen this sweep blind; read them first', leftover;
  END IF;
END $$;

COMMIT;

-- After applying: the very next tool-auditor run's review items must arrive at
-- status='needs_human_review', handler_agent='' (RUNBOOK "Prove the config fix
-- at the artefact"). Then update bugs_open/291 + the WDS-018 register entry and
-- close the bug if the guard is also live (fixed-AND-live bar).
