-- 483 — html-developer-chunked: its three generate_html steps DECLARE what they
--       read, so the ensureCoreFields gate (RFC_029 §10.13 step 3) cannot affect
--       them. CONFIG HALF, applied BEFORE the Go gate by design.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- At the end of every field extraction, a safety net (`ensureCoreFields`,
-- unified_extractor.go) tops up six values whether or not the caller asked for
-- them: domain, objective, model, current_page, current_section, render_context.
-- The first three are consumed undeclared by 55 steps across 31 agent types and
-- are NOT being touched. The last three have ONE live consumer that depends on
-- the injection rather than on its own request list: this agent.
--
-- `generate_html` (html_actions.go) builds its prompt from `buildContextSmart`,
-- which hands `input_fields` to ExtractFields — or, when `input_fields` is absent,
-- a fixed fallback list that does NOT include current_page. The prompt builder
-- then reads context["current_page"] to write a "Page: <name>" line. So today,
-- for this agent, that line exists ONLY because the safety net injected the page.
--
-- ============================================================================
-- THE RULE
-- ============================================================================
-- OWNER RULING 2026-08-18 (RFC_029 §10.13, sequence A step 3): gate the three
-- page-ish injections to requested-only — and the config edit for the one
-- dependent consumer lands FIRST, because config is live immediately and the
-- Go gate rides the next roll. Order is load-bearing: gate-then-config would
-- leave a window where the prompt silently loses its page line.
--
-- ============================================================================
-- HOW THIS CASE MEASURES AGAINST IT
-- ============================================================================
-- [MEASURED 2026-08-19] html-developer-chunked: exactly 1 live definition; 3
-- generate_html steps (generate_structure / generate_styles / generate_content);
-- NONE carries input_fields; 0 orchestrations ALL TIME under owner_agent_type;
-- 0 live steps name it as agent_type. It is dormant. That makes this edit cheap
-- AND makes it exactly the kind of thing that rots silently: a dormant agent
-- that depends on an injection nobody declared is the first thing to break the
-- day someone wakes it, long after the gate shipped and everyone forgot why.
-- Declaring the list now costs nothing and closes that door permanently.
--
-- The list written here is the fallback `buildContextSmart` uses today
-- (input_data, site_architecture, site_content, domain_analysis) PLUS
-- current_page — so the four it already extracted keep being extracted, and
-- the one it was getting by injection is now requested explicitly. Net change to
-- what reaches the prompt: NONE today; the difference is only that after the
-- gate it STILL gets current_page, where without this file it would not.
--
-- ROLLBACK: 483_html_developer_chunked_declares_current_page_ROLLBACK.sql removes
-- input_fields from the three steps and asserts absence.

BEGIN;

-- GUARD: refuse unless the live rows are the ones this file was written against.
DO $$
DECLARE
    n int;
    k text;
    step jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'html-developer-chunked' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '483: expected exactly 1 live html-developer-chunked row, found %', n;
    END IF;

    FOREACH k IN ARRAY ARRAY['generate_structure','generate_styles','generate_content'] LOOP
        SELECT default_config #> ARRAY['workflow','steps',k] INTO step
          FROM agent_definitions
         WHERE type = 'html-developer-chunked' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
        IF step IS NULL THEN
            RAISE EXCEPTION '483: html-developer-chunked has no % step — the workflow has been restructured since 2026-08-19; re-derive this migration', k;
        END IF;
        IF step->>'action' <> 'generate_html' THEN
            RAISE EXCEPTION '483: % runs %, not generate_html — the declaration would be read by nothing', k, step->>'action';
        END IF;
        IF step->'config' ? 'input_fields' THEN
            RAISE EXCEPTION '483: % ALREADY carries input_fields (%) — another session has applied this or an equivalent; do not overwrite it', k, step->'config'->'input_fields';
        END IF;
    END LOOP;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(jsonb_set(jsonb_set(default_config,
           '{workflow,steps,generate_structure,config,input_fields}',
           '["input_data","site_architecture","site_content","domain_analysis","current_page"]'::jsonb, true),
           '{workflow,steps,generate_styles,config,input_fields}',
           '["input_data","site_architecture","site_content","domain_analysis","current_page"]'::jsonb, true),
           '{workflow,steps,generate_content,config,input_fields}',
           '["input_data","site_architecture","site_content","domain_analysis","current_page"]'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'html-developer-chunked'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

-- VERIFY — a DO block that RAISEs, not a SELECT (ON_ERROR_STOP cannot stop a
-- COMMIT on a non-empty result set; LANDMINES / RFC_006).
DO $$
DECLARE
    k text;
    cfg jsonb;
    described text;
BEGIN
    FOREACH k IN ARRAY ARRAY['generate_structure','generate_styles','generate_content'] LOOP
        SELECT default_config #> ARRAY['workflow','steps',k,'config'] INTO cfg
          FROM agent_definitions
         WHERE type = 'html-developer-chunked' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        IF NOT (cfg->'input_fields' @> '["current_page"]'::jsonb) THEN
            RAISE EXCEPTION '483 VERIFY: %.input_fields does not contain current_page after the update: %', k, cfg->'input_fields';
        END IF;
        IF jsonb_array_length(cfg->'input_fields') <> 5 THEN
            RAISE EXCEPTION '483 VERIFY: %.input_fields has % entries, want 5', k, jsonb_array_length(cfg->'input_fields');
        END IF;
        -- The pre-existing keys must survive (jsonb_set is surgical; assert it).
        IF cfg->>'output_type' <> 'html' OR cfg->>'generation_type' IS NULL OR (cfg->>'max_tokens')::int IS NULL THEN
            RAISE EXCEPTION '483 VERIFY: %''s pre-existing config keys did not survive: %', k, cfg::text;
        END IF;
    END LOOP;

    -- NEGATIVE CONTROL in the same transaction: no OTHER live agent's generate_html
    -- step may have acquired this exact 5-entry list (an UPDATE with a wider WHERE
    -- would pass the assertions above identically).
    SELECT string_agg(ad.type || '.' || s.key, ', ') INTO described
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
       AND s.value->>'action' = 'generate_html'
       AND s.value->'config'->'input_fields' = '["input_data","site_architecture","site_content","domain_analysis","current_page"]'::jsonb
       AND ad.type <> 'html-developer-chunked';
    IF described IS NOT NULL THEN
        RAISE EXCEPTION '483 VERIFY: the declaration leaked to steps it was not meant for: %', described;
    END IF;

    RAISE NOTICE '483 OK: html-developer-chunked''s three generate_html steps now declare current_page; no other step acquired the list';
END $$;

COMMIT;
