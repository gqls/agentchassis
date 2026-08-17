-- 450_tool_auditor_review_items_park_at_needs_human_review.sql
--
-- bugs_open/291: tool-auditor's create_review_item step (create_work_item action)
-- files item_type='needs_human_review' at handler_agent='hitl-review' — an agent
-- that has NEVER existed (documented 2026-04-19 as "a convention, not a registered
-- agent"; the planned handler was never built) — and carries NO status key, so
-- create_work_item_action.go:208-211 births every row at the dispatchable default
-- 'triaged'. The dispatch loop claims each row and ClaimWorkItemAction's
-- handler-not-registered branch flips it to status='blocked',
-- error='Handler agent not registered: hitl-review', where feasibility-recheck can
-- never promote it (its predicate requires the handler to exist in
-- agent_definitions). Measured 2026-08-17: 14 rows, 2 sites, growing daily
-- (5 on 08-16), each holding an idx_swi_dedup slot audit_review_<page_id> so the
-- auditor's LATER findings for the same page are silently dropped.
--
-- THE FIX HERE: add "status": "needs_human_review" to the step config — the HITL
-- parking idiom (migration 217). Parked rows are invisible to dispatch (claim and
-- the loader both take only status IN ('triaged','approved')), and their designed
-- exit already exists: the admin confirm endpoint
-- (internal/core-manager/admin/confirm_work_item_handler.go) turns a confirmed
-- spec.check='tool_auditor' review item into an 'improve_tool' follow-up.
--
-- ⚠ DELIBERATELY NOT touching handler_agent in this migration. The LIVE binary's
-- create_work_item action REFUSES an empty handler_agent
-- (create_work_item_action.go:184-187, "handler_agent config is required"), and
-- the enclosing loop runs continue_on_error:true — so flipping the handler to the
-- canonical '' TODAY would make every review-item filing hard-error and be
-- silently swallowed: every finding lost with NO row at all, strictly worse than
-- the bug. The parked 'hitl-review' value is inert at this status (the exact shape
-- resolve_composition_layout_action.go has run safely with since April). The flip
-- to '' is STAGED in docs024_key_docs_latest/bugfix_291_hitl_review_phantom_handler/
-- (STAGED_tool_auditor_review_handler_to_empty.sql) and moves into this directory
-- ONLY after the fleet roll carrying the relaxed validation is provenance-verified.
--
-- Config-only: no image dependency, LIVE ON APPLY.
-- Rollback: 450_tool_auditor_review_items_park_at_needs_human_review_ROLLBACK.sql
--   (removes the status key at the same path, gated on it being present).

BEGIN;

SELECT snapshot_agent('tool-auditor',
  '450_tool_auditor_review_items_park_at_needs_human_review: pre-update');

-- Pre-state gate: exactly the shape this file was written against, at the pinned id.
DO $$
DECLARE
  n int;
  h text;
  has_status boolean;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
     AND type = 'tool-auditor'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '450: expected exactly 1 live tool-auditor row at the pinned id, found %', n;
  END IF;

  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}',
         (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}') ? 'status'
    INTO h, has_status
    FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';

  IF h IS DISTINCT FROM 'hitl-review' OR has_status THEN
    RAISE EXCEPTION '450: pre-state has moved (handler_agent=%, has status key=%) — someone changed this step; re-read it before applying', h, has_status;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,status}',
         '"needs_human_review"'),
       updated_at = now()
 WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
   AND type = 'tool-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Post-state verify. DO/RAISE, never a bare SELECT: ON_ERROR_STOP ignores a
-- non-empty result, so a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
  s text;
  h text;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,status}',
         default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}'
    INTO s, h
    FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';

  IF s IS DISTINCT FROM 'needs_human_review' THEN
    RAISE EXCEPTION '450: post-state status is % — expected needs_human_review', s;
  END IF;
  IF h IS DISTINCT FROM 'hitl-review' THEN
    RAISE EXCEPTION '450: post-state handler_agent is % — this file must not touch it (see header)', h;
  END IF;
END $$;

COMMIT;
