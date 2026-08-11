-- 384 — tool-acceptance: let the dispatcher name the page it means
-- (`url_field` on request_browser_run; owner decision 2026-08-11)
--
-- WHY. The Tier-4 page lookup resolves `pages.name IN (function,
-- 'tool-'||function)` — a name-guess. Nine active tool placements resolve NO
-- page under it (eight loancalculator.co.uk slugs that are different WORDS
-- from their component functions, e.g. tool-early-settlement on
-- tool-settlement-calculator; plus tool-loan-repayment which sits on `index`
-- and could never be fixed by any rename). The code has ALWAYS supported an
-- explicit URL and checks it FIRST (`tool_acceptance_actions.go:163-166`,
-- covered by `tool_acceptance_actions_test.go:377-380`); the live config
-- simply never set `url_field`. Evidence: staged_component_build NOTES
-- `## 2026-08-11 (parallel session)` and HANDOFF_2026-08-11 §3 item 4's
-- correction box.
--
-- WHAT. One key: steps.request_run.config.url_field =
-- "input_data.spec.page_url". Additive and inert — extraction of an absent
-- field yields "", and the name lookup runs exactly as today (the guard is
-- `if pageURL == ""`). Nothing changes for any existing producer; a work item
-- whose spec carries `page_url` resolves to that URL without a rename.
-- DB config: live immediately, no image, no roll. The running v1.0.1284
-- binary already carries the reading code (in the tree well before the
-- 08-11 09:23Z build).
--
-- PRODUCER HALF, deliberately NOT here: the due-sweep only raises items for
-- tools that already carry a PLAN, and every PLAN-carrying tool today
-- resolves by name — so no current producer needs to write page_url. The
-- first producers will be manual/driven runs at the loancalculator tools once
-- that lane's golden-derived PLANs exist. The manual wrapper
-- (tool_acceptance_run.sh, RUNBOOK §10) should gain an optional page_url
-- argument then — its owning session was mid-edit on 2026-08-11, so that edit
-- is left to them.
--
-- ROLLBACK: restore from the snapshot this file takes, or simply
--   UPDATE agent_definitions SET default_config = default_config #-
--   '{workflow,steps,request_run,config,url_field}' WHERE type='tool-acceptance-agent'
--   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

SELECT snapshot_agent('tool-acceptance-agent',
  '384_tool_acceptance_url_field: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,request_run,config,url_field}',
         '"input_data.spec.page_url"'
       ),
       updated_at = now()
 WHERE type = 'tool-acceptance-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  v text;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-acceptance-agent'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '384 guard: expected exactly 1 live tool-acceptance-agent row, found %', n;
  END IF;
  SELECT default_config#>>'{workflow,steps,request_run,config,url_field}' INTO v
    FROM agent_definitions
   WHERE type = 'tool-acceptance-agent'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF v IS DISTINCT FROM 'input_data.spec.page_url' THEN
    RAISE EXCEPTION '384 guard: url_field is %, want input_data.spec.page_url', COALESCE(v, '(null)');
  END IF;
END $$;

COMMIT;
