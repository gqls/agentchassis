-- 473 — bug 184: make a page-rerender the mechanical repair for literal_markdown.
--
-- Two config changes on the live page-rerender row:
--   1. check_rerender_mode.condition gains OR spec.reason == 'literal_markdown',
--      so a literal_markdown work item (whose spec carries that reason as of the
--      184 fix-2 check change) takes the sections branch instead of the
--      assemble-only else branch.
--   2. rerender_sections.config.strip_literal_markdown = true — the opt-in flag
--      (default OFF in code) that makes RerenderPageSectionsAction pass each
--      stored section's content_data through
--      datahelpers.StripLiteralMarkdownFromContentData before it feeds both the
--      render context and the persisted mergedContent.
--
-- ORDERING: safe to apply BEFORE the image carrying the 184 fix-2 code —
--   - the extra OR clause references a spec.reason value nothing emits until
--     the re-routed check ships, so no dispatch can reach it early;
--   - strip_literal_markdown is unread by the old binary.
-- The check re-route itself ships only WITH the strip hook (one image), so
-- there is no window in which items dispatch to an unequipped handler.
--
-- Idempotent: the UPDATE is guarded on the condition not already carrying the
-- literal; re-running is a no-op. Verify is DO/RAISE (a SELECT-only verify
-- cannot stop the COMMIT — RFC_006 landmine).

BEGIN;

CREATE TABLE IF NOT EXISTS _backup_473_page_rerender AS
  SELECT id, type, default_config, now() AS backed_up_at
    FROM agent_definitions
   WHERE type = 'page-rerender' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,check_rerender_mode,config,condition}',
           to_jsonb(replace(
             default_config #>> '{workflow,steps,check_rerender_mode,config,condition}',
             'input_data.spec.reason == ''template_changed''',
             'input_data.spec.reason == ''template_changed'' OR input_data.spec.reason == ''literal_markdown'''
           ))
         ),
         '{workflow,steps,rerender_sections,config,strip_literal_markdown}',
         'true'::jsonb
       )
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
       NOT LIKE '%literal_markdown%'
   -- anchor: refuse to fire if the live literal this replace targets has moved
   AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
       LIKE '%input_data.spec.reason == ''template_changed''%';

DO $$
DECLARE cond text;
BEGIN
  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
    INTO cond
    FROM agent_definitions
   WHERE type = 'page-rerender' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cond IS NULL OR cond NOT LIKE '%literal_markdown%' THEN
    RAISE EXCEPTION '473 FAILED: check_rerender_mode.condition does not carry literal_markdown — the template_changed anchor has moved; read the live condition and re-anchor. Live: %', cond;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND (default_config #> '{workflow,steps,rerender_sections,config,strip_literal_markdown}')::text = 'true'
  ) THEN
    RAISE EXCEPTION '473 FAILED: rerender_sections.config.strip_literal_markdown is not true';
  END IF;

  -- the other four reasons must have survived the string surgery
  IF cond NOT LIKE '%image_landed%' OR cond NOT LIKE '%section_data_resolved%'
     OR cond NOT LIKE '%cta_links_stale%' OR cond NOT LIKE '%template_changed%' THEN
    RAISE EXCEPTION '473 FAILED: an existing reason clause was damaged. Live: %', cond;
  END IF;

  RAISE NOTICE '473 OK: condition = %', cond;
END $$;

COMMIT;
