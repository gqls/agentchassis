-- 457_tool_auditor_review_handler_to_empty.sql — bugs_open/291, Phase 3.
--
-- ORDERING GATE SATISFIED 2026-08-17 17:1x — this file was staged outside the
-- migration runner's directory until the binary carrying the relaxed
-- create_work_item validation was PROVEN live, because applying it early does not
-- error loudly: it makes every tool-auditor review-item filing hard-error inside a
-- continue_on_error loop, i.e. every finding silently lost.
--
-- The proof, taken before moving this file in (all three, one sitting):
--   1. DIGEST match — local `aqls/agent-chassis:v1.0.1307` RepoDigest
--      sha256:8339bdbd7999… == the running pods' imageID. (The 08-17 afternoon
--      "fresh build" FAILED exactly this check: same tag v1.0.1305, node served its
--      cached image, 203 commits inert fleet-wide. See LANDMINES.)
--   2. BINARY probe with both controls discriminating — `DISABLE_UNREGISTERED_HANDLER_DEMOTION`
--      PRESENT (1); positive control "Handler agent not registered: " present (2);
--      negative control ZZZ_NOT_A_REAL_SYMBOL_291 absent (0).
--   3. ANCESTRY, with a control — image OCI revision a6d1c53c068a5df421479cc9e8801f251f80d539;
--      `git merge-base --is-ancestor` says YES for both c8400e452 (guard + relaxed
--      validation) and f629f4530 (kill-switch), and correctly NO for a commit made
--      after the build.
--
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
