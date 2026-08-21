-- ═══════════════════════════════════════════════════════════════════════════
-- ⛔ SUPERSEDED 2026-08-21 — the migration this rolls back was never applied
-- (retired in favour of 515_page_build_handler_plan_sections_declares_page_type).
-- See the SUPERSEDED banner on 514's main file for the full account. This
-- sidecar is kept only so the pair stays together under one number; it has
-- nothing to roll back and nothing should ever invoke it.
-- ═══════════════════════════════════════════════════════════════════════════

-- ROLLBACK for 514 — remove the page_type wire from page-build-handler's
-- plan_sections step, returning it to the whole-tree search.
--
-- WHEN YOU WOULD RUN THIS: if page_record.page_type turns out NOT to be what
-- plan_sections wants (e.g. a future refactor renames the field, or a site's
-- page_record legitimately lacks page_type where the search would have found
-- a valid value elsewhere). The observable would be the component selector's
-- relevance scoring regressing for pages where page_record.page_type is
-- absent but some other tree location has a real one — check
-- collected_data->'page_record'->>'page_type' on a failing run before
-- assuming this file is the cause.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','plan_sections','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '514 ROLLBACK: no live page-build-handler plan_sections step to roll back';
    END IF;
    IF NOT (cfg ? 'page_type') THEN
        RAISE EXCEPTION '514 ROLLBACK: plan_sections carries no page_type — 514 is not applied, or has already been rolled back';
    END IF;
    -- Refuse to delete a value that is NOT the one 514 wrote: another session
    -- may have re-pointed it on purpose.
    IF cfg->>'page_type' IS DISTINCT FROM 'page_record.page_type' THEN
        RAISE EXCEPTION '514 ROLLBACK: page_type is %, not 514''s page_record.page_type — someone else owns this wire; do not remove it', cfg->'page_type';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,plan_sections,config,page_type}',
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','plan_sections','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'page_type' THEN
        RAISE EXCEPTION '514 ROLLBACK VERIFY: page_type is still present: %', cfg->'page_type';
    END IF;
    IF cfg->>'page_name' IS DISTINCT FROM 'page_record.name' THEN
        RAISE EXCEPTION '514 ROLLBACK VERIFY: the removal took a neighbouring key with it: %', cfg::text;
    END IF;
    RAISE NOTICE '514 ROLLBACK OK: page_type wire removed; the step''s other five keys intact';
END $$;

COMMIT;
