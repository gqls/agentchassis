-- 490 — internal-linker: make check_candidates satisfiable, uncap the candidate
--       set, and arm the branch against re-drift (bugs_open/313 + bugs_open/298)
--
-- WHY. `load_candidate_pages` declares `output_format: "array"`, so the action
-- returns a bare slice with NO `count` key (database_actions.go:129); the very
-- next step tests `candidate_pages.count > 0`, a path that cannot resolve against
-- an array (conditional_branch_action.go:397-412 — strategy 5 requires a map), so
-- the numeric arm returns false (:275-284) and EVERY run since the agent's
-- creation (2026-04-12) has exited at `complete_no_candidates` — including a
-- 2026-08-18 run holding fifteen candidates in collected_data. `plan_links`, the
-- only LLM step, has NEVER executed: llm_call_log for agent_type='internal-linker'
-- has zero rows in all history (re-verified 2026-08-19). 57 needs_internal_links
-- work items read `complete` across four months with no link ever planned.
-- Diagnosis loop: CONFIRMED, first iteration
-- (RUN_CORRELATION_ID=c4aa3559-86b1-4356-a28b-c71dfa661465, 2026-08-18).
--
-- WHAT THIS DOES — four edits, one row, and why each:
--   1. load_candidate_pages.output_format: "array" -> "object". The producer then
--      emits rows/count/columns and `candidate_pages.count > 0` resolves — the
--      same pairing the working load_target_page/check_target_found uses two
--      steps earlier in this same workflow.
--   2. load_candidate_pages.query: drop `LIMIT 15` (bugs_open/298 — the cut was
--      ORDER BY p.name, alphabetical; 8 of 26 sites exceed 15 candidates today,
--      worst 69) and MARK the 800-char truncation the query already performs
--      (' […truncated]', migration 446's worked remedy — an unmarked cut is the
--      same defect one level down). The dominant column is already bounded, so
--      the worst site costs ~60kB of prompt — bounded payload instead of bounded
--      coverage, the shape council-approved for bugs_open/275 (corr b684a399).
--      ORDER BY p.name stays: with no cap it is deterministic presentation, not
--      a cut.
--   3. plan_links.prompt_template: `{{range .candidate_pages}}` ->
--      `{{range .candidate_pages.rows}}`. THE MANDATORY SECOND HALF of edit 1:
--      ranging the new map would yield columns/count/rows values in key order —
--      a broken prompt instead of a dead branch (bugs_open/313 §fix-candidate-1).
--   4. check_candidates.config.fail_on_non_numeric: true — opts THIS routing
--      condition into the loud-failure guard shipped alongside this migration
--      (conditional_branch_action.go; register WFA-019). Unknown config keys are
--      ignored at execution (bugs_open/101), so this is inert until the chassis
--      rolls with that change, and inert afterwards unless the path stops
--      resolving again — a pure tripwire against re-drift.
--
-- The seed 101_internal_linker.sql carries the defective pairing and is applied
-- history — deliberately NOT edited; this migration supersedes it.
--
-- Scoped by id, pre-state gated (exact-text equality), DO/RAISE verify with
-- controls AND a PREPARE parse-check on the rewritten query (a LIKE check proves
-- a needle, not that the SQL parses — bugs_open/314 §9's council round), snapshot
-- first, rollback sidecar. Config-only, LIVE ON APPLY, no roll needed for edits
-- 1-3. Council: submitted alongside the Go halves (in scope via platform/).
--
-- ROLLBACK: 490_internal_linker_candidates_object_uncapped_fail_loud_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('internal-linker',
  '490_internal_linker_candidates_object_uncapped_fail_loud: pre-update');

DO $$
DECLARE n int; ofmt text; q text; cond text; tmpl text; flag text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'internal-linker'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '490: expected exactly 1 live internal-linker row, found % — a second active row would make the id-scoped UPDATE a silent no-op', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_candidate_pages,config,output_format}',
         default_config#>>'{workflow,steps,load_candidate_pages,config,query}',
         default_config#>>'{workflow,steps,check_candidates,config,condition}',
         default_config#>>'{workflow,steps,plan_links,config,prompt_template}',
         default_config#>>'{workflow,steps,check_candidates,config,fail_on_non_numeric}'
    INTO ofmt, q, cond, tmpl, flag
    FROM agent_definitions WHERE id = '93cffe67-baf4-4fb1-bec9-ba546fb24a54';

  IF ofmt IS NULL THEN
    RAISE EXCEPTION '490: pinned row 93cffe67 not found or has no load_candidate_pages step — wrong row?';
  END IF;
  IF ofmt IS DISTINCT FROM 'array' THEN
    RAISE EXCEPTION '490: pre-state output_format is %, expected ''array'' — already applied, or changed under me; refusing', ofmt;
  END IF;
  IF q IS DISTINCT FROM $oldq$SELECT p.name, p.url, p.title, p.page_type, LEFT(string_agg(pc.rendered_html, ' '), 800) as content_sample FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL WHERE p.site_id = $1 AND p.name != $2 AND p.status = 'active' AND p.page_type IN ('content', 'service', 'landing', 'tool') GROUP BY p.name, p.url, p.title, p.page_type HAVING COUNT(pc.id) > 0 ORDER BY p.name LIMIT 15$oldq$ THEN
    RAISE EXCEPTION '490: load_candidate_pages.query is not the text this migration was written against — refusing to overwrite someone else''s change';
  END IF;
  IF cond IS DISTINCT FROM 'candidate_pages.count > 0' THEN
    RAISE EXCEPTION '490: check_candidates condition is % — changed under me, refusing', cond;
  END IF;
  IF position('{{range .candidate_pages}}' in tmpl) = 0 THEN
    RAISE EXCEPTION '490: plan_links template does not contain the bare {{range .candidate_pages}} — already applied or changed; refusing';
  END IF;
  IF position('.candidate_pages.rows' in tmpl) > 0 THEN
    RAISE EXCEPTION '490: template already references .candidate_pages.rows — double-apply; refusing';
  END IF;
  IF (length(tmpl) - length(replace(tmpl, '{{range', ''))) / 7 <> 1 THEN
    RAISE EXCEPTION '490: expected exactly ONE {{range in the template — the single replace() below would be unsafe; refusing';
  END IF;
  IF flag IS NOT NULL THEN
    RAISE EXCEPTION '490: fail_on_non_numeric already present (%) on check_candidates — double-apply; refusing', flag;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config =
       jsonb_set(
         jsonb_set(
           jsonb_set(
             jsonb_set(
               default_config,
               '{workflow,steps,load_candidate_pages,config,output_format}',
               to_jsonb('object'::text)
             ),
             '{workflow,steps,load_candidate_pages,config,query}',
             to_jsonb($newq$SELECT p.name, p.url, p.title, p.page_type, CASE WHEN length(string_agg(pc.rendered_html, ' ')) > 800 THEN LEFT(string_agg(pc.rendered_html, ' '), 800) || ' […truncated]' ELSE string_agg(pc.rendered_html, ' ') END as content_sample FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL WHERE p.site_id = $1 AND p.name != $2 AND p.status = 'active' AND p.page_type IN ('content', 'service', 'landing', 'tool') GROUP BY p.name, p.url, p.title, p.page_type HAVING COUNT(pc.id) > 0 ORDER BY p.name$newq$::text)
           ),
           '{workflow,steps,plan_links,config,prompt_template}',
           to_jsonb(
             replace(
               default_config#>>'{workflow,steps,plan_links,config,prompt_template}',
               '{{range .candidate_pages}}',
               '{{range .candidate_pages.rows}}'
             )
           )
         ),
         '{workflow,steps,check_candidates,config,fail_on_non_numeric}',
         'true'::jsonb
       ),
       updated_at = now()
 WHERE id = '93cffe67-baf4-4fb1-bec9-ba546fb24a54'
   AND type = 'internal-linker'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE ofmt text; q text; tmpl text; cond text; then_s text; else_s text; flag text; tq text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_candidate_pages,config,output_format}',
         default_config#>>'{workflow,steps,load_candidate_pages,config,query}',
         default_config#>>'{workflow,steps,plan_links,config,prompt_template}',
         default_config#>>'{workflow,steps,check_candidates,config,condition}',
         default_config#>>'{workflow,steps,check_candidates,config,then_step}',
         default_config#>>'{workflow,steps,check_candidates,config,else_step}',
         default_config#>>'{workflow,steps,check_candidates,config,fail_on_non_numeric}',
         default_config#>>'{workflow,steps,load_target_page,config,query}'
    INTO ofmt, q, tmpl, cond, then_s, else_s, flag, tq
    FROM agent_definitions WHERE id = '93cffe67-baf4-4fb1-bec9-ba546fb24a54';

  IF ofmt IS DISTINCT FROM 'object' THEN
    RAISE EXCEPTION '490: output_format is % after the update, expected object', ofmt;
  END IF;
  IF q ~* 'LIMIT[[:space:]]+[0-9]+' THEN
    RAISE EXCEPTION '490: a row LIMIT survives on load_candidate_pages after the update';
  END IF;
  IF position(' […truncated]' in q) = 0 THEN
    RAISE EXCEPTION '490: the truncation marker is absent from the rewritten query';
  END IF;
  IF position('HAVING COUNT(pc.id) > 0' in q) = 0 OR position('ORDER BY p.name' in q) = 0 THEN
    RAISE EXCEPTION '490: the rewritten query lost a load-bearing clause (HAVING / ORDER BY)';
  END IF;
  -- the rewritten query must PARSE — a position() check proves a needle, not SQL
  EXECUTE 'PREPARE _mig490_parse_check AS ' || q;
  EXECUTE 'DEALLOCATE _mig490_parse_check';
  IF position('{{range .candidate_pages.rows}}' in tmpl) = 0 THEN
    RAISE EXCEPTION '490: template does not range over .candidate_pages.rows after the update';
  END IF;
  IF position('{{range .candidate_pages}}' in tmpl) > 0 THEN
    RAISE EXCEPTION '490: the bare {{range .candidate_pages}} survives — both forms would be live at once';
  END IF;
  IF flag IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '490: fail_on_non_numeric is % after the update, expected true', flag;
  END IF;
  -- controls: what this migration must NOT have disturbed
  IF cond IS DISTINCT FROM 'candidate_pages.count > 0'
     OR then_s IS DISTINCT FROM 'load_specs'
     OR else_s IS DISTINCT FROM 'complete_no_candidates' THEN
    RAISE EXCEPTION '490: check_candidates routing changed (cond=%, then=%, else=%) — this migration must not touch it', cond, then_s, else_s;
  END IF;
  IF position('## Candidate Pages' in tmpl) = 0
     OR position('{{.target_page.url}}' in tmpl) = 0
     OR position('Return ONLY valid JSON:' in tmpl) = 0 THEN
    RAISE EXCEPTION '490: the template lost a load-bearing section — the replace() hit more than it should have';
  END IF;
  IF tq IS NULL OR position('LIMIT 1' in tq) = 0 THEN
    RAISE EXCEPTION '490: load_target_page.query disturbed — its LIMIT 1 is correct and must stay';
  END IF;
END $$;

COMMIT;
