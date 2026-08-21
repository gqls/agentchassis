-- 548 — bugs_open/305: point `render_section` at the copy gate's OUTPUT, because
-- the repair was being made and then thrown away.
--
-- ⚠ HELD (`_HOLD`): the live build `bac189921` does NOT carry the `result` key —
-- that change is newer than it. Apply only after a roll that includes it.
-- Numbered 548 because 518 was taken by another session between drafting and
-- writing; the directory moved from 516 to 547 in one afternoon, so CHECK THE
-- NUMBER AT WRITE TIME, not at plan time.
--
-- ⚠ APPLY WITH THE IMAGE THAT CARRIES THE `result` KEY (commit after 0eea9e597).
-- Applied against an older binary, `copy_gate.result` does not exist and
-- render_section would find no content — a loud failure, not a silent one, but a
-- failure. Probe the capability first, on every replica, with an absent control:
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
--   kubectl -n ai-persona-system exec $POD -- grep -ac 'copy_gate' /proc/1/exe   # >0
--   kubectl -n ai-persona-system exec $POD -- grep -ac 'copy_gatez' /proc/1/exe  # 0
--
-- WHAT WAS WRONG, measured at the ARTEFACT (2026-08-21, live):
-- `remortgagecalculator.uk/mortgage-lenders`, orchestration completed 18:38:38,
-- components saved 18:38:37. The gate reported, for two sections,
-- `status: repaired, hits_before: 1, hits_after: 0` — and the stored
-- `content_data.subheadline` came back BYTE-IDENTICAL to the pre-repair
-- `generated_content.result.subheadline`, still carrying
-- "…using your own numbers rather than a lender's advertised rate."
--
-- The renderer never saw the patch. The action mutated the writer's content map
-- in place, which is correct in memory and does not survive the step boundary:
-- `saveStepResultWithRetry` reloads a fresh state and copies only the CURRENT
-- step's own stepName/output_field, so an in-place edit to the PREVIOUS step's
-- output is dropped. (The same mechanism drops the `__copy_gate` page counter,
-- which is why the budget is per section.)
--
-- ⚠ THE SHAPE OF THE FAILURE IS THE POINT: an honest marker over a page that
-- never changed. `hits_after: 0` was true of the map the action held and false
-- of the page. Nothing errored, the orchestration COMPLETED, and only reading
-- the stored artefact showed it.
--
-- THE FIX: the action now also returns the (patched) content as its own
-- `result`, which is the mechanism that demonstrably persists — `copy_gate_0`,
-- `copy_gate_1` … all reach the durable row — and this migration points
-- `content_from` at it. Exactly the shape `generated_content.result` already
-- had, so the loop's reference-suffixing handles it unchanged.
--
-- Anchored on the current value and needle-gated: a re-run is a 0-row no-op, and
-- if another session has re-pointed this step the UPDATE does nothing rather
-- than stealing their edge.

BEGIN;

SELECT snapshot_agent('page-content-writer',
                      '518_render_section_reads_the_repaired_content.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,content_from}',
         '"copy_gate.result"'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,action}' = 'render_component'
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,action}' = 'rewrite_negations'
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,content_from}' = 'generated_content.result';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,content_from}' = 'copy_gate.result';
  IF n <> 1 THEN
    RAISE EXCEPTION '548 FAILED: expected render_section.content_from = copy_gate.result on exactly 1 row, got %', n;
  END IF;
  RAISE NOTICE '548 OK: render_section now reads the repaired content';
END $$;

COMMIT;
