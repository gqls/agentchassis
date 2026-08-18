-- 461 — corrects 460's fixer query: pages has NO filename column
--       (bugs_open/283 §13; caught minutes after 460 applied, before any
--        fixer run consumed it)
--
-- 460 is ledger-recorded and is NOT edited (the 458 norm). Its Change 2
-- embedded `p.filename` inside the create_rerender query STRING — invisible to
-- 460's own probe run, because the inner query is DATA to the UPDATE and only
-- becomes SQL when the fixer's query_database step executes it. The first real
-- template fix would have errored at that step. Caught by running the SAME
-- query shape directly for the canary's completion item, where Postgres named
-- the column at parse time.
--
-- `filename` is DERIVED in Go (datahelpers.PageDeployFilename(url, name)) and
-- measured 2026-08-18: NO page-rerender step reads input_data.spec.filename —
-- getPageInfo recomputes it from the page row on the deploy path. So the fix
-- is to DROP the key from the spec, not to fake the derivation in SQL.
--
-- Guarded on 460's post-image (the query must still contain p.filename), so
-- re-application is a no-op and a concurrent correction aborts loudly.
BEGIN;

SELECT snapshot_agent('component-template-fixer', '461 pre-image: drop p.filename from create_rerender');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,create_rerender,config,query}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,create_rerender,config,query}',
        ',''filename'',p.filename',
        ''
      ))
    ),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_rerender,config,query}' LIKE '%p.filename%';

DO $$
DECLARE q text;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}'
    INTO q FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF q LIKE '%filename%' THEN
    RAISE EXCEPTION '461: p.filename still present in create_rerender query: %', left(q,200);
  END IF;
  IF q NOT LIKE '%template_changed%' OR q NOT LIKE '%page_rerender%' THEN
    RAISE EXCEPTION '461: 460''s intent was lost while editing: %', left(q,200);
  END IF;
  -- The corrected inner query must PARSE. PREPARE compiles without executing;
  -- this is the check 460 lacked and the reason this file exists.
  EXECUTE 'PREPARE chk461 AS ' || q;
  DEALLOCATE chk461;
END $$;

COMMIT;
