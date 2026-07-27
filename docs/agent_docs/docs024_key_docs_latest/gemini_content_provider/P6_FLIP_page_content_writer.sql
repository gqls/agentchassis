-- FILE: P6_FLIP_page_content_writer.sql
--
-- Flip page-content-writer from anthropic/claude-sonnet-4-6 to
-- gemini/gemini-pro-latest. Owner chose the pro tier (quality / provider
-- diversity, 2026-07-27). Context: bugs_open/107, workstream PLAN P6.
--
-- SAFE TO RUN ONLY IF the chassis pod carries the 107 fix. Verified on
-- v1.0.1173 (pod agent-chassis-5f85dff548-8d2tq, 5 positive markers + 2 negative
-- controls, 2026-07-27). Against an older chassis this reproduces the 07-24
-- starvation.
--
-- RUN IT:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db \
--     < docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/P6_FLIP_page_content_writer.sql
--
-- ROLLBACK: re-run with the two values swapped back to
-- anthropic / claude-sonnet-4-6 / ANTHROPIC_API_KEY, or restore ai_service from
-- bak_agent_definitions_pcw_20260727 (already created 2026-07-27).
--
-- THREE THINGS THIS SCRIPT GETS RIGHT THAT AN OBVIOUS VERSION DOES NOT:
--
-- 1. The step is NOT top-level. It is nested inside the loop's sub_workflow:
--    workflow -> steps -> process_sections_loop -> config -> sub_workflow ->
--    steps -> generate_content -> config -> ai_service.
--    The shorter path you would guess returns NULL with no error.
--
-- 2. It MERGES with `||` rather than replacing. `max_tokens: 8000` lives inside
--    the same ai_service block; a wholesale jsonb_set replace drops it and the
--    client falls back to 2048 — a 4x cut to the writer's budget, invisible in
--    the diff, surfacing later as truncated sections.
--
-- 3. It is guarded on updated_at. Several threads write this row (the
--    architecture-review re-seed touched it at 13:44:56 on 2026-07-27).
--    UPDATE 0 means someone else wrote it — RE-READ, do not retry blind.
--    And UPDATE 1 is NOT proof the value landed: jsonb_set on a missing parent
--    is a silent no-op that still reports 1. Hence the verification block.

\set ON_ERROR_STOP on

BEGIN;

-- Backup (idempotent; already created on 2026-07-27, recreated here so the
-- script is self-contained if run later).
DROP TABLE IF EXISTS bak_agent_definitions_pcw_20260727_p6;
CREATE TABLE bak_agent_definitions_pcw_20260727_p6 AS
SELECT * FROM agent_definitions WHERE type = 'page-content-writer';

-- Read the current state. CHECK THIS OUTPUT before trusting the update below:
-- provider should be anthropic, and max_tokens should be 8000.
\echo '--- BEFORE ---'
SELECT a.updated_at,
       v->'config'->'ai_service'                                  AS ai_service_before,
       (v->'config'->>'prompt_template') LIKE '%Voice & Style%'    AS style_block_intact
FROM agent_definitions a,
     jsonb_each(a.default_config->'workflow'->'steps'->'process_sections_loop'
                ->'config'->'sub_workflow'->'steps') AS e(k,v)
WHERE a.type='page-content-writer' AND a.is_active
  AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND e.k='generate_content';

-- The flip. Guarded on the updated_at read at 13:44:56Z on 2026-07-27; if
-- another session has written the row since, this updates 0 rows and the
-- verification block below will raise.
UPDATE agent_definitions a
SET default_config = jsonb_set(
      a.default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,ai_service}',
      (a.default_config->'workflow'->'steps'->'process_sections_loop'->'config'
        ->'sub_workflow'->'steps'->'generate_content'->'config'->'ai_service')
      || '{"provider":"gemini","model":"gemini-pro-latest","api_key_env_var":"GEMINI_API_KEY"}'::jsonb),
    updated_at = now()
WHERE a.type='page-content-writer' AND a.is_active
  AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND a.updated_at = '2026-07-27 13:44:56.343485+00';

\echo '--- AFTER ---'
SELECT v->'config'->'ai_service'                                  AS ai_service_now,
       (v->'config'->>'prompt_template') LIKE '%Voice & Style%'    AS style_block_intact,
       length(v->'config'->>'prompt_template')                     AS tmpl_chars
FROM agent_definitions a,
     jsonb_each(a.default_config->'workflow'->'steps'->'process_sections_loop'
                ->'config'->'sub_workflow'->'steps') AS e(k,v)
WHERE a.type='page-content-writer' AND a.is_active
  AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND e.k='generate_content';

-- Fail LOUD rather than committing a half-applied or no-op change. Asserts the
-- three things that can each go wrong silently: provider not switched, the
-- 8000-token budget dropped by a replace, the style prompt lost.
DO $$
DECLARE svc jsonb; style boolean;
BEGIN
  SELECT v->'config'->'ai_service',
         (v->'config'->>'prompt_template') LIKE '%Voice & Style%'
    INTO svc, style
  FROM agent_definitions a,
       jsonb_each(a.default_config->'workflow'->'steps'->'process_sections_loop'
                  ->'config'->'sub_workflow'->'steps') AS e(k,v)
  WHERE a.type='page-content-writer' AND a.is_active
    AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND e.k='generate_content';

  IF svc->>'provider' IS DISTINCT FROM 'gemini' THEN
    RAISE EXCEPTION 'provider is % — the update did not land (guard failed on updated_at, or the jsonb path is wrong). ROLLING BACK.', svc->>'provider';
  END IF;
  IF svc->>'model' IS DISTINCT FROM 'gemini-pro-latest' THEN
    RAISE EXCEPTION 'model is % not gemini-pro-latest. ROLLING BACK.', svc->>'model';
  END IF;
  IF (svc->>'max_tokens')::int IS DISTINCT FROM 8000 THEN
    RAISE EXCEPTION 'max_tokens is % not 8000 — a sibling key was dropped, which would quietly cut the writer to the 2048 default. ROLLING BACK.', svc->>'max_tokens';
  END IF;
  IF NOT style THEN
    RAISE EXCEPTION 'the Voice & Style block is gone from prompt_template. ROLLING BACK.';
  END IF;
  RAISE NOTICE 'OK: gemini/gemini-pro-latest, max_tokens 8000 preserved, style block intact.';
END $$;

COMMIT;

-- NEXT (do not skip): rebuild ONE page and READ the copy before any site-wide
-- rewrite. `complete` is not proof the work happened — read the artefact.
-- Compare against the Claude baseline snapshot referenced in the brochure
-- workstream (about_copy_before.txt) and check: em dashes, filler words,
-- fact-first openings, and that the page's own story survived.
-- Watch for a *TruncatedError naming thinking, which would mean the 8192 reserve
-- is too small for the writer's real (context-loaded) prompt — raise
-- thinking_reserve_tokens in the same ai_service block and re-run.
