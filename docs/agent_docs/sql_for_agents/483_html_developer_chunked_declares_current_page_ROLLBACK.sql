-- 483 ROLLBACK — remove input_fields from html-developer-chunked's three
-- generate_html steps, restoring the buildContextSmart fallback path.
--
-- Only safe BEFORE the ensureCoreFields gate (RFC_029 §10.13 step 3) has rolled:
-- after the gate, removing the declaration means the prompt loses its page line.
-- The guard below does not know whether the gate is live — the operator must.

BEGIN;

DO $$
DECLARE
    k text;
    cfg jsonb;
BEGIN
    FOREACH k IN ARRAY ARRAY['generate_structure','generate_styles','generate_content'] LOOP
        SELECT default_config #> ARRAY['workflow','steps',k,'config'] INTO cfg
          FROM agent_definitions
         WHERE type = 'html-developer-chunked' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
        IF cfg IS NULL THEN
            RAISE EXCEPTION '483 ROLLBACK: no live html-developer-chunked.% — nothing to roll back', k;
        END IF;
        IF cfg->'input_fields' IS DISTINCT FROM '["input_data","site_architecture","site_content","domain_analysis","current_page"]'::jsonb THEN
            RAISE EXCEPTION '483 ROLLBACK: %.input_fields is % — not the list 483 wrote; another lane has changed it, do not remove', k, cfg->'input_fields';
        END IF;
    END LOOP;
END $$;

UPDATE agent_definitions
   SET default_config = (((default_config
           #- '{workflow,steps,generate_structure,config,input_fields}')
           #- '{workflow,steps,generate_styles,config,input_fields}')
           #- '{workflow,steps,generate_content,config,input_fields}'),
       updated_at = NOW()
 WHERE type = 'html-developer-chunked'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    k text;
    cfg jsonb;
BEGIN
    FOREACH k IN ARRAY ARRAY['generate_structure','generate_styles','generate_content'] LOOP
        SELECT default_config #> ARRAY['workflow','steps',k,'config'] INTO cfg
          FROM agent_definitions
         WHERE type = 'html-developer-chunked' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
        IF cfg ? 'input_fields' THEN
            RAISE EXCEPTION '483 ROLLBACK VERIFY: %.input_fields still present: %', k, cfg->'input_fields';
        END IF;
        IF cfg->>'output_type' <> 'html' THEN
            RAISE EXCEPTION '483 ROLLBACK VERIFY: %''s other keys did not survive: %', k, cfg::text;
        END IF;
    END LOOP;
    RAISE NOTICE '483 ROLLBACK OK: input_fields removed from the three steps; other keys intact';
END $$;

COMMIT;
