-- 446 — tool-suggester: MARK the truncated descriptions 445 introduced
--
-- WHY. Migration 445 (bugs_open/275) traded a silent ROW cap for a silent
-- COLUMN cap: `left(description, 200)` cuts 50 of 74 descriptions with no
-- marker at all. The council's bug_historian seat objected (corr b684a399,
-- medium) and is right — that is 016b §9's shape, *"a hard cap that silently
-- discards its input's tail rewrites meaning, and the tail is whatever was
-- composed LAST"*, and it is the SAME failure mode as the bug 445 fixes: the
-- model answers plausibly either way and nothing looks wrong. A tool whose
-- distinguishing capability sits past character 200 now reads as generic, and
-- neither the model nor a human reading the row can tell it was cut.
--
-- This estate already marks truncation explicitly rather than silently —
-- `datahelpers.TruncateString` appends "..." under a tested ellipsis contract,
-- and `internal/adapters/webscrape/truncation.go` writes a literal
-- "[Content truncated ...]" banner. 445 followed neither.
--
-- THE FIX IS THE MARKER, NOT A BIGGER CAP, and that is a deliberate choice:
-- raising the cap to 300 costs +53% payload against 445's +22.4% and still cuts
-- 24 descriptions silently. The marker restores the SIGNAL at ~3% (a 13-char
-- suffix on the 50 rows that are actually cut), so the model can see that a
-- description continues and weigh it accordingly. Loss without a signal is the
-- defect; loss with one is a budget.
--
-- Cost: 50 marked rows x 13 chars = ~650 chars on a ~20,100-char payload.
--
-- Same shape as 445: scoped by id, pre-state gated, DO/RAISE verify, snapshot
-- first, rollback sidecar. Config-only, LIVE ON APPLY.
--
-- ROLLBACK: 446_tool_suggester_mark_truncated_descriptions_ROLLBACK.sql (restores
-- 445's unmarked left(description, 200)).

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '446_tool_suggester_mark_truncated_descriptions: pre-update');

DO $$
DECLARE q text; n int;
BEGIN
  -- exactly one live row for this type (checked because the "four agent types
  -- have TWO active rows" landmine makes an id-scoped UPDATE silently no-op on
  -- the shadowed one; tool-suggester is NOT one of those four — measured
  -- 2026-08-17, a single row at version 1 — and this gate keeps it that way)
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-suggester'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '446: expected exactly 1 live tool-suggester row, found % — a second active row would make the id-scoped UPDATE below a silent no-op', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}' INTO q
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';
  IF q IS NULL OR position('left(description, 200)' in q) = 0 THEN
    RAISE EXCEPTION '446: pre-state is not 445''s query — refusing to overwrite: %', q;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_library_tools,config,query}',
         to_jsonb('SELECT id::text, function, display_name, category, CASE WHEN length(description) > 200 THEN left(description, 200) || '' […truncated]'' ELSE description END AS description FROM content_components WHERE component_level = ''tool'' AND forked_from IS NULL AND is_active = true AND html_template != '''' AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) ORDER BY display_name'::text)
       ),
       updated_at = now()
 WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
   AND type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}' INTO q
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';
  IF position('[…truncated]' in q) = 0 THEN
    RAISE EXCEPTION '446: the truncation marker is absent — the whole point of this migration: %', q;
  END IF;
  IF q ~* 'LIMIT[[:space:]]+[0-9]+' THEN
    RAISE EXCEPTION '446: a row LIMIT reappeared (445 removed it): %', q;
  END IF;
  IF position('requires-backend' in q) = 0 OR position('deploy_config' in q) = 0 THEN
    RAISE EXCEPTION '446: 406''s requires-backend gate was lost: %', q;
  END IF;
END $$;

COMMIT;
