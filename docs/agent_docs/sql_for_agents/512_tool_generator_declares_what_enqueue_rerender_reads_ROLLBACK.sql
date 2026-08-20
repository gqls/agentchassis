-- ROLLBACK for 512 — remove the input_fields declaration from tool-generator's
-- enqueue_rerender step, returning `reason` and `component_id` to the whole-tree
-- search.
--
-- WHEN YOU WOULD RUN THIS: if the declaration turns out to starve a field the
-- action actually consumes. The observable would be a tool-birth rerender that
-- stops being enqueued at all (site_id or domain no longer reaching the action),
-- NOT a change of render mode — reason was already inert by value, so its loss
-- cannot be the symptom. Check `create_rerender_items` results on tool-generator
-- runs before blaming this file.
--
-- NOTE it does NOT restore the conflict rows deliberately; it restores the
-- search, and the rows follow because the search is what writes them.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','enqueue_rerender','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '512 ROLLBACK: no live tool-generator enqueue_rerender step to roll back';
    END IF;
    IF NOT (cfg ? 'input_fields') THEN
        RAISE EXCEPTION '512 ROLLBACK: enqueue_rerender carries no input_fields — 512 is not applied, or has already been rolled back';
    END IF;
    -- Refuse to delete a list that is NOT the one 512 wrote: another session may
    -- have declared something wider on purpose.
    IF cfg->'input_fields' <> '["site_id","domain"]'::jsonb THEN
        RAISE EXCEPTION '512 ROLLBACK: input_fields is %, not 512''s ["site_id","domain"] — someone else owns this list; do not remove it', cfg->'input_fields';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,enqueue_rerender,config,input_fields}',
       updated_at = NOW()
 WHERE type = 'tool-generator'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','enqueue_rerender','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'input_fields' THEN
        RAISE EXCEPTION '512 ROLLBACK VERIFY: input_fields is still present: %', cfg->'input_fields';
    END IF;
    IF cfg->>'site_id' IS DISTINCT FROM 'site_record.site_id'
       OR cfg->>'domain' IS DISTINCT FROM 'input_data.domain' THEN
        RAISE EXCEPTION '512 ROLLBACK VERIFY: the removal took a neighbouring key with it: %', cfg::text;
    END IF;
    RAISE NOTICE '512 ROLLBACK OK: input_fields removed; the step''s other six keys intact';
END $$;

COMMIT;
