-- 269_deduplicate_sections_handler.sql
--
-- The handler for `content_duplication` items raised by
-- check_content_duplication (owner ruling 2026-07-31: the deterministic half
-- only — see gauntlet_dead_cta/SUMMARY_2026-07-31).
--
-- ORDERING, AND WHY THIS SEED IS SAFE TO APPLY BEFORE THE IMAGE ROLLS
-- ------------------------------------------------------------------
-- `remove_duplicate_page_sections` is a NEW action. Go changes are inert until
-- an image is rebuilt and rolled, so until then this agent would fail at runtime
-- with an unknown action.
--
-- That is safe here because NOTHING CAN DISPATCH TO IT YET. A discovery check
-- only runs when a discovery agent's workflow config names it, and
-- `content_duplication` is named by no agent's config. So the chain
-- check -> item -> this handler cannot start until someone deliberately opts the
-- check in, which must happen AFTER the roll.
--
-- The seed is applied first anyway, deliberately, because
-- handler_coverage_test.go's ratchet is a snapshot of LIVE agent_definitions:
-- listing an agent there before it exists would make the guard assert something
-- false, and that guard exists precisely because two checks once routed at
-- handlers that were never registered (bugs_open/077).
--
-- ORDER OF OPERATIONS, in full:
--   1. this seed                                    (now — inert)
--   2. add "deduplicate-sections" to knownHandlerAgents  (now — truthful)
--   3. chassis image carrying remove_duplicate_page_sections rolls
--   4. pod-grep the running binary for the action name + a control
--   5. THEN add "content_duplication" to a discovery agent's check list
--
-- WHY THE RERENDER IS A QUEUED ITEM AND NOT AN INLINE STEP
-- -------------------------------------------------------
-- Deleting rows changes what the page should render, so the served page is stale
-- until it is reassembled. Doing the rerender inline would couple a deterministic
-- DB transaction to a multi-agent render+deploy that can fail on its own; a
-- half-succeeded pair would leave the page correct in the database and doubled on
-- the site, which is the exact divergence class this codebase keeps being bitten
-- by. Queuing a `page_rerender` item keeps each half independently retryable, and
-- mirrors what the hand-fix of bugs_open/156 did (DELETEs, then an assemble-only
-- rerender).

BEGIN;

-- image_tag matches its live peers (color-variable-fixer, page-rerender were on
-- v1.0.1212 when this was written) so this agent participates in the normal roll
-- rather than pinning itself to `latest`, which is the column default and is not
-- what any other agent uses. The tag it starts on does NOT contain
-- remove_duplicate_page_sections — see the ordering note above; that is expected
-- and safe because nothing can dispatch here yet.
INSERT INTO agent_definitions (type, display_name, description, category, image_tag, is_active, default_config, created_at, updated_at)
VALUES (
  'deduplicate-sections',
  'Deduplicate Sections',
  'Removes content-identical duplicate sections from one page and queues an assemble-only rerender. Deterministic: no LLM, no rewriting. The near-duplicate case that needs judgement is never routed here — see check_content_duplication.',
  'specialist',
  'v1.0.1212',
  true,
  jsonb_build_object(
    'processing_mode', 'orchestrator',
    'timeout_seconds', 300,
    'workflow', jsonb_build_object(
      'start_step', 'remove_duplicates',
      'steps', jsonb_build_object(
        'remove_duplicates', jsonb_build_object(
          'action', 'remove_duplicate_page_sections',
          'description', 'Delete content-identical duplicate sections on this page and renumber positions. Re-derives the victims from current content; never trusts the detection-time list.',
          'config', jsonb_build_object('page_id', 'input_data.page_id'),
          'output_field', 'dedupe_result',
          'next_step', 'queue_rerender'
        ),
        'queue_rerender', jsonb_build_object(
          'action', 'create_work_item',
          'description', 'Queue an assemble-only rerender so the served page stops showing the removed sections',
          'config', jsonb_build_object(
            'site_id', 'input_data.site_id',
            'source', 'deduplicate-sections',
            'item_type', 'page_rerender',
            'item_domain', 'build',
            'handler_agent', 'page-rerender',
            'severity', 'medium',
            'priority', 10,
            'summary', 'Re-assemble and deploy after removing duplicate sections',
            'item_key_prefix', 'dedupe_rerender',
            'spec', jsonb_build_object('page_id', 'input_data.page_id')
          ),
          'output_field', 'rerender_item',
          'next_step', 'complete'
        ),
        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Duplicate sections removed; rerender queued',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('dedupe_result', 'rerender_item')
          )
        )
      )
    )
  ),
  now(), now()
)
ON CONFLICT (type, version) DO NOTHING;

-- Guard: a silent no-op INSERT is indistinguishable from success.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'deduplicate-sections' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 live deduplicate-sections row, found %', n;
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'deduplicate-sections' AND deleted_at IS NULL
       AND default_config->'workflow'->'steps'->'remove_duplicates'->>'action'
           = 'remove_duplicate_page_sections'
  ) THEN
    RAISE EXCEPTION 'deduplicate-sections is seeded but does not call remove_duplicate_page_sections';
  END IF;
END $$;

COMMIT;

-- Verify (expect one row, and the action name spelled exactly as registered):
--   SELECT type, default_config->'workflow'->>'start_step' AS start,
--          default_config->'workflow'->'steps'->'remove_duplicates'->>'action' AS action
--   FROM agent_definitions WHERE type='deduplicate-sections' AND deleted_at IS NULL;
