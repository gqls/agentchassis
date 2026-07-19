-- 175_experience_context_component_js.sql — finish the job 174 started.
--
-- 174 surfaced `js_snippets` loaders into load_context. The new contract critic
-- then objected, correctly and repeatedly, that it STILL could not verify
-- Journeys D and E because "no gauntlet-interface / tool-arena-interface script
-- source is in context". It was right: component-owned JavaScript lives in
-- `content_components.js_content`, a DIFFERENT column from `js_snippets`, and
-- 174 surfaced only the latter.
--
--   js_snippets rows matching 'gauntlet-interface' ....... 0
--   content_components.js_content for gauntlet-interface .. 3909 bytes
--
-- So this is the same lying-by-omission defect a third time — component_level
-- filter (fix 3), js_snippets (174), and now js_content — which is itself the
-- argument for the critic's "an unverifiable pair is ITSELF an objection" rule:
-- that rule is what surfaced this gap instead of letting a fourth silent
-- approval through.
--
-- AND THE MISMATCH IT COULD NOT SEE IS REAL. Checked by hand against the live
-- asset: none of the six selectors the plan's Journey E assumes
-- (__timer, __response-input, __submit-btn, __result-score, __play-again-btn,
-- __start-btn) appear in gauntlet-interface's js_content — zero occurrences
-- each. The critic refused to approve something that was not merely
-- unverifiable but actually broken. Once this migration lands it will be able to
-- prove it rather than merely withhold approval.
--
-- Config-only: live on commit, no image roll. Seed 167 synced in-place.

BEGIN;

SELECT snapshot_agent('experience-planner', 'pre-update: 175 component js_content in context')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='experience-planner' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_context,config,query}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'load_context'->'config'->>'query',
           '  AS text',
           ' || '
           || ' E''\n\n## Component-owned JavaScript (content_components.js_content) — ALSO BINDING\n'' || '
           || ' E''Distinct from js_snippets above: this JS ships as /tools/assets/<function>.js and owns its component''''s DOM contract. If the plan names a selector or a computation for one of these components, it must appear in the source below. A selector the plan assumes and this source lacks does not exist.\n\n'' || '
           || ' COALESCE((SELECT string_agg(''### '' || cc.function || E'' (component-owned js)\n```javascript\n'' || left(cc.js_content, 8000) || CASE WHEN length(cc.js_content) > 8000 THEN E''\n/* … TRUNCATED at 8000 chars — ask for the rest rather than guessing … */'' ELSE '''' END || E''\n```\n'', E''\n'' ORDER BY cc.function) '
           || ' FROM (SELECT DISTINCT cc2.function, cc2.js_content FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc2 ON cc2.id=pc.component_id WHERE p.site_id = $1::uuid AND cc2.js_content IS NOT NULL AND cc2.js_content <> '''') cc), '
           || ' ''(no attached component carries its own js_content — any plan step assuming client-side behaviour from a component must say which script will provide it)'')'
           || '  AS text'
         )),
         false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'load_context'->'config'->>'query' INTO q
    FROM agent_definitions
   WHERE type='experience-planner'
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('js_content' in q) = 0 THEN
    RAISE EXCEPTION 'component js_content section not added';
  END IF;
  -- everything 174 and earlier added must survive
  IF position('js_snippets' in q) = 0 THEN
    RAISE EXCEPTION 'js_snippets section lost (174)';
  END IF;
  IF position('Attached components' in q) = 0 THEN
    RAISE EXCEPTION 'attached-components section lost';
  END IF;
  IF position('Open work items' in q) = 0 THEN
    RAISE EXCEPTION 'open-work-items section lost';
  END IF;
END $$;

COMMIT;

-- Rollback: restore the snapshot taken above.
