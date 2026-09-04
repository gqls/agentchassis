-- 766 — point `copywriter-directory-discovery` at the researcher that can actually serve it, and
-- re-enable it. Follows 765 (which creates `copywriter-directory-researcher` and leaves it inert).
-- Owner instruction 2026-09-04: add the copywriter kind to the global register.
--
-- WHY THE TASK WAS DISABLED: it targeted `directory-researcher`, whose extraction prompt is the
-- AI-MODELS one ("You are extracting ATOMIC, CITABLE claims about AI MODELS…"). Three runs on
-- 2026-09-03/04 did search and scrape correctly and extracted ZERO candidates — the model obeyed the
-- prompt it was given. Disabled 2026-09-04 by the portfolio_positioning lane; this file is the repair.
--
-- THE QUERY STAYS UNDER 200 CHARACTERS. `extractSearchQuery` (web_search_action.go) drops any resolved
-- query of >=200 chars as a presumed LLM refusal message and then reports "search query not found",
-- which reads as a config fault and is not one. The current 165-char query is kept as-is; the
-- pre-shortening 444-char original stays parked under `research_query_previous_over_200` and the guard
-- below refuses if anyone has re-lengthened it. LANDMINES carries this trap twice.
--
-- FIRST RUN IS CHEAP AND SHOULD BE READ: one web_search, three scrapes, one LLM call. Read the result
-- before trusting the register:
--   SELECT status, current_step, collected_data->'registration'
--     FROM orchestration_states WHERE owner_agent_type='copywriter-directory-researcher'
--     ORDER BY created_at DESC LIMIT 1;
--   SELECT slug, name, status FROM directory_entities WHERE kind='copywriter' ORDER BY created_at;
-- Expect ORGANISATIONS ONLY. A named individual, a sector or a listicle title in that list is the
-- prompt's hard rule failing and is worth a prompt revision before the weekly cadence takes over.

BEGIN;

DO $g$
DECLARE t record;
BEGIN
  SELECT * INTO t FROM scheduled_tasks WHERE name='copywriter-directory-discovery';
  IF NOT FOUND THEN RAISE EXCEPTION '766 REFUSED: scheduled task copywriter-directory-discovery not found'; END IF;
  IF t.target_agent_type <> 'directory-researcher' THEN
    RAISE EXCEPTION '766 REFUSED: task targets % — expected directory-researcher (has someone already retargeted it?)', t.target_agent_type; END IF;
  IF length(t.input_data->>'research_query') >= 200 THEN
    RAISE EXCEPTION '766 REFUSED: research_query is % chars — web_search silently drops >=200 (LANDMINES)', length(t.input_data->>'research_query'); END IF;
  IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type='copywriter-directory-researcher'
                   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) THEN
    RAISE EXCEPTION '766 REFUSED: copywriter-directory-researcher does not exist — apply 765 first'; END IF;
END $g$;

UPDATE scheduled_tasks
   SET target_agent_type = 'copywriter-directory-researcher',
       enabled = true,
       last_triggered_at = NULL,          -- fire on the scheduler's next tick
       last_completed_at = NULL,
       description = description || ' — RETARGETED 2026-09-04 (migration 766) to copywriter-directory-researcher; the previous target extracts AI models only, which is why three runs returned zero candidates.',
       updated_at = NOW()
 WHERE name='copywriter-directory-discovery';

DO $v$
DECLARE t record; n int;
BEGIN
  SELECT * INTO t FROM scheduled_tasks WHERE name='copywriter-directory-discovery';
  IF t.target_agent_type <> 'copywriter-directory-researcher' THEN RAISE EXCEPTION '766 VERIFY: retarget did not take'; END IF;
  IF NOT t.enabled THEN RAISE EXCEPTION '766 VERIFY: task is not enabled'; END IF;
  IF t.last_triggered_at IS NOT NULL THEN RAISE EXCEPTION '766 VERIFY: last_triggered_at should be NULL so the scheduler fires it'; END IF;
  IF length(t.input_data->>'research_query') >= 200 THEN RAISE EXCEPTION '766 VERIFY: query length regressed to %', length(t.input_data->>'research_query'); END IF;
  SELECT count(*) INTO n FROM scheduled_tasks WHERE target_agent_type='copywriter-directory-researcher' AND enabled;
  IF n <> 1 THEN RAISE EXCEPTION '766 VERIFY: expected exactly 1 enabled task on the new agent, found %', n; END IF;
END $v$;

COMMIT;
