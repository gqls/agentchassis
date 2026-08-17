-- 445 — tool-suggester: show the LLM the WHOLE library, and pay for it out of
-- `description` rather than out of coverage (bugs_open/275)
--
-- WHY. `load_library_tools` ends `ORDER BY display_name LIMIT 30`. Measured
-- 2026-08-16 the library holds **74** masters, so the model deciding which
-- tools a site should have was shown **30 of 74** — and which 44 were hidden
-- was an accident of the alphabet, not a judgement of relevance. It is a
-- SILENT cap: an LLM returns plausible suggestions whether it saw 30 rows or
-- 74, so nothing downstream ever looks wrong. (Filed at 68 masters / 38 hidden
-- on 2026-08-14; the library grows, so the damage grows with it. Migration 406
-- flagged this defect in its own header and deliberately did not touch it.)
--
-- The gate 406 added barely narrows anything and is NOT the cause: exactly
-- **1 of 74** masters carries `requires-backend`, and 3 of 40 sites have the
-- capability. The cap is doing all of the hiding.
--
-- WHAT THE REAL CONSTRAINT IS — measured, not assumed. The bug's candidate 1
-- says "if the real constraint is prompt size, cap by TOKENS at the prompt
-- assembly, not by row count in the dark". The data says which knob:
--   description  29,832 chars across 74 rows  <-- 80% of the payload
--   id            2,664
--   display_name  2,100
--   function      1,828
--   category        785
-- description: median 374, mean 403, max 2,526; 50 of 74 exceed 200 chars.
--
--   TODAY     30 rows (41% of library), description uncapped -> 16,421 chars
--   THIS      74 rows (100%),           description <= 200   -> 20,376 chars
--   (74 rows at description <= 300 would be 25,146.)
--
-- So: the WHOLE library for +24% payload (~4,000 chars, ~1k tokens). 200 was
-- checked for MEANING as well as size — the first 200 chars of the longest
-- descriptions still say what the tool IS ("The Arena is Spark's competitive
-- mode, v1 as a fully self-contained client-side experience..."), which is
-- what a relevance judgement needs.
--
-- The prompt renders `- {{.display_name}} ({{.function}}, id: {{.id}}): {{.description}}`.
-- `category` is SELECTED and never RENDERED (785 chars of dead payload) —
-- deliberately left alone: dropping a column a future consumer might read is
-- scope creep with non-zero risk for a 2% saving.
--
-- 406'S GATE IS PRESERVED BYTE-FOR-BYTE. The bug warns explicitly not to
-- restore an older ungated sketch; the only differences from the live text are
-- `left(description, 200) AS description` and the removal of `LIMIT 30`.
--
-- Consumers of load_library_tools's output: the `suggest_tools` LLM step of
-- tool-suggester, nothing else (output_field `library_tools`, read only inside
-- this workflow) — re-verified 2026-08-16, same finding as 406's header.
--
-- Config-only: no image dependency, LIVE ON APPLY. Scoped by id with a
-- pre-state gate, which is the shape 406's own council round asked future
-- agent_definitions migrations to adopt (406 addendum, 2026-08-14).
--
-- ROLLBACK: 445_tool_suggester_whole_library_ROLLBACK.sql, or restore the
-- snapshot this file takes
-- (snapshot_agent note '445_tool_suggester_whole_library: pre-update').

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '445_tool_suggester_whole_library: pre-update');

-- Pre-state gate: refuse unless the row is EXACTLY the one this file was
-- written against, still carrying 406's gated query with the LIMIT 30 intact.
-- If another session has already changed this step, abort rather than clobber.
DO $$
DECLARE q text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
     AND type = 'tool-suggester'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '445: expected exactly 1 live tool-suggester row at the pinned id, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}' INTO q
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';

  IF q IS NULL OR position('requires-backend' in q) = 0 THEN
    RAISE EXCEPTION '445: pre-state does not carry 406''s requires-backend gate — refusing to overwrite: %', q;
  END IF;
  IF position('LIMIT 30' in q) = 0 THEN
    RAISE EXCEPTION '445: pre-state has no LIMIT 30 — someone has already changed this step; re-read it before applying: %', q;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_library_tools,config,query}',
         to_jsonb('SELECT id::text, function, display_name, category, left(description, 200) AS description FROM content_components WHERE component_level = ''tool'' AND forked_from IS NULL AND is_active = true AND html_template != '''' AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) ORDER BY display_name'::text)
       ),
       updated_at = now()
 WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
   AND type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Post-state verify. DO/RAISE, never a bare SELECT: ON_ERROR_STOP ignores a
-- non-empty result, so a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE q text; p jsonb;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}',
         default_config#>'{workflow,steps,load_library_tools,config,params}'
    INTO q, p
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';

  -- the cap is gone
  IF q ~* 'LIMIT[[:space:]]+[0-9]+' THEN
    RAISE EXCEPTION '445: a LIMIT survives in the query: %', q;
  END IF;
  -- description is bounded at the source instead
  IF position('left(description, 200)' in q) = 0 THEN
    RAISE EXCEPTION '445: description is not bounded — the payload argument depended on it: %', q;
  END IF;
  -- 406's gate is intact (the thing the bug warns about losing)
  IF position('requires-backend' in q) = 0 OR position('deploy_config' in q) = 0 THEN
    RAISE EXCEPTION '445: 406''s requires-backend gate was lost: %', q;
  END IF;
  -- the $1 binding 406 added is untouched
  IF p IS NULL OR jsonb_array_length(p) <> 1 OR p->>0 <> 'input_data.site_id' THEN
    RAISE EXCEPTION '445: load_library_tools params wrong: %', p;
  END IF;
END $$;

COMMIT;
