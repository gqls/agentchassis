-- 752_d4b_governor_council_gate_stage_b_HOLD.sql — D4b STAGE B, CONFIG-ONLY. HELD.
--
-- WHY HELD (two reasons, both owner-shaped). (1) Applying this IS arming: the moment it applies,
-- council-gate runs are refused whenever governor_admits_agent('council-gate') is false — today
-- that means shed level 3 (the 'research' seed), i.e. at 95% of the monthly budget. The
-- architecture seat on corr dc6d2a54 and this lane both said: do not arm until the OWNER has
-- confirmed the level (L3 = last / L1 = first is a cultural choice, not a config default; asked
-- in README_where_we_are three times, unanswered as of writing). (2) Its own lighter council
-- round (editquality, dc6d2a54: "a sketch is not an edit"). Apply by hand ONLY after both:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--     -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/752_d4b_governor_council_gate_stage_b_HOLD.sql
-- then drop the _HOLD suffix and `--record-only --note` it (bugs_closed/150 lifecycle).
--
-- WHAT IT IS. The guardian's advisory on dc6d2a54, adopted: instead of a Go gate at
-- processor.go executeWorkflow (the ONE admission point every agent in the fleet passes
-- through), the spend-governor check is a LEADING STEP INSIDE council-gate's OWN workflow:
--   gate_spend_governor  (query_database)  SELECT governor_admits_agent('council-gate') …
--   route_spend_governor (conditional)     admitted → load_schema_hint (the old start step)
--                                           else     → note_withheld
--   note_withheld        (append_doc_note) durable record: subject_key spend-governor,
--                                           categories [spend-governor, withheld-run]
--   complete_withheld    (complete_workflow) terminal, with a success_message a polling
--                                           session will READ: withheld, not queued, do not retry
-- No Go, no roll, nothing shared touched, opt-in by construction (only a workflow carrying the
-- step is governed), and THE ORCHESTRATION ROW IS THE OBSERVABLE: the 097 runbook's existing
-- `… WHERE collected_data->'input_data'->>'fix_correlation_id' = '<corr>'` query returns
-- current_step = complete_withheld. That dissolves two other seats' objections: bug_historian's
-- (a failed withheld-row write dropped the run with no trace — here the row exists BEFORE the
-- decision) and reuse_agent's (a third audit table — governor_withheld_runs — when
-- agent_error_log already exists; so this migration DROPS governor_withheld_runs and its view).
--
-- FAIL-OPEN, at the step level: gate_spend_governor carries error_step = load_schema_hint, so
-- an unreadable governor (or a $ctx binding failure) RUNS THE REVIEW rather than refusing it —
-- the same posture as the selector, the loader and the claim backstop. The engine resolves
-- error_step from the step-level field first, then config (coordinator.go:3672-3675).
--
-- GUARD: the live council-gate row's WHOLE workflow md5 must be the text this was written
-- against (8dd74a5b042a7376a1e26fbf5db6ba00, version 2, start_step load_schema_hint, 44 steps,
-- none of the four new names present). Any drift REFUSES — investigate, do not overwrite.
-- ⚠ 099_SYNC_gate_roster.py --apply regenerates this row and would silently REMOVE these
-- steps (guardian, dc6d2a54). It is suspended (LANDMINES, migration 377); the standing guard is
-- 752_..._VERIFY.sql, run daily beside 657's — it FAILS BY DESIGN until this applies.
-- Rollback: 752_..._HOLD_ROLLBACK.sql (removes the four steps, restores start_step, RECREATES
-- governor_withheld_runs so 751's own rollback still finds what it expects).

BEGIN;

-- ---------------------------------------------------------------- refusal first
DO $$
DECLARE m text; ss text; n int; v int;
BEGIN
  SELECT md5(default_config#>>'{workflow}'), default_config#>>'{workflow,start_step}', version,
         (SELECT count(*) FROM jsonb_object_keys(default_config#>'{workflow,steps}'))
    INTO m, ss, v, n
  FROM agent_definitions
  WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF m IS NULL THEN RAISE EXCEPTION '752 REFUSED: no live council-gate row.'; END IF;
  IF ss = 'gate_spend_governor' THEN
    RAISE EXCEPTION '752 REFUSED: start_step is already gate_spend_governor — already applied (replay).';
  END IF;
  IF m <> '8dd74a5b042a7376a1e26fbf5db6ba00' THEN
    RAISE EXCEPTION '752 REFUSED: council-gate workflow md5 % is not the text this was written against (8dd74a5b…, v2, 44 steps) — drifted, investigate before overwriting.', m;
  END IF;
  IF ss <> 'load_schema_hint' OR n <> 44 THEN
    RAISE EXCEPTION '752 REFUSED: expected start_step load_schema_hint and 44 steps, found % / %', ss, n;
  END IF;
  SELECT count(*) INTO n FROM agent_definitions, jsonb_object_keys(default_config#>'{workflow,steps}') k
  WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND k IN ('gate_spend_governor','route_spend_governor','note_withheld','complete_withheld');
  IF n <> 0 THEN RAISE EXCEPTION '752 REFUSED: % of the four new step names already exist on the row.', n; END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits_agent') THEN
    RAISE EXCEPTION '752 REFUSED: governor_admits_agent() missing — 751 (stage A) not applied.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='governor_withheld_runs') THEN
    RAISE EXCEPTION '752 REFUSED: governor_withheld_runs absent — 751 not applied or this already ran.';
  END IF;
  -- One live row only: the UPDATE below is by type, and a duplicate active row would be silently half-updated.
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '752 REFUSED: % active council-gate rows, expected exactly 1.', n; END IF;

  PERFORM snapshot_agent('council-gate', '752 pre-apply: D4b stage B gate step (RFC_065)');
END $$;

-- ---------------------------------------------------------------- the four steps + the new start step
UPDATE agent_definitions
SET default_config = jsonb_set(
  jsonb_set(default_config, '{workflow,steps}',
    (default_config#>'{workflow,steps}') || $J$
    {
      "gate_spend_governor": {
        "action": "query_database",
        "description": "D4b (RFC_065, AGOV-013): ask the spend governor whether a council-gate run is admitted right now, and compose the withheld note in the same query so nothing else has to know the level. $ctx.correlation_id is THIS run's envelope correlation — the key the 097 runbook finds a submission by. error_step = FAIL-OPEN: an unreadable governor runs the review.",
        "config": {
          "query": "SELECT governor_admits_agent('council-gate') AS admitted, gs.shed_level, format('spend-governor: council-gate run for submission %s WITHHELD at shed level %s (%s%% of budget spent) - NOT queued; do not retry; re-trigger when governor_state.shed_level drops. RFC_065.', $1, gs.shed_level, round(100*gs.mtd_usd/NULLIF(gc.monthly_budget_usd,0))) AS body FROM governor_state gs, governor_config gc WHERE gs.id=1 AND gc.id=1",
          "params": ["$ctx.correlation_id"],
          "output_format": "object"
        },
        "output_field": "governor",
        "next_step": "route_spend_governor",
        "error_step": "load_schema_hint"
      },
      "route_spend_governor": {
        "action": "conditional",
        "description": "D4b: admitted → the review proceeds exactly as before (load_schema_hint was the start step); withheld → record and stop. Nothing in the review path below this step changed.",
        "config": {
          "condition": "governor.admitted == true",
          "then_step": "load_schema_hint",
          "else_step": "note_withheld"
        }
      },
      "note_withheld": {
        "action": "append_doc_note",
        "description": "D4b: the DURABLE record that a council run was withheld — orchestration rows evaporate in ~24h, doc_notes do not. Same channel the governor's level-change alarm uses (subject_key spend-governor). A note failure must not hide the outcome: route on to the terminal regardless.",
        "config": {
          "subject_type": "pipeline",
          "subject_key": "spend-governor",
          "note_body_field": "governor.body",
          "note_categories": ["spend-governor", "withheld-run"],
          "note_source": "council-gate",
          "created_by": "council-gate",
          "error_step": "complete_withheld"
        },
        "output_field": "withheld_note",
        "next_step": "complete_withheld"
      },
      "complete_withheld": {
        "action": "complete_workflow",
        "description": "D4b terminal: the council did NOT run. This step name is what the 097 runbook's find-your-run query shows as current_step — read it as WITHHELD, not queued.",
        "config": {
          "output_fields": ["governor"],
          "success_message": "council gate did NOT run: WITHHELD by the spend governor (RFC_065) — not queued, do not retry; re-trigger when governor_state.shed_level drops below council-gate's class threshold (governor_agent_class_map)"
        }
      }
    }
    $J$::jsonb),
  '{workflow,start_step}', '"gate_spend_governor"'::jsonb)
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ---------------------------------------------------------------- reuse_agent + architecture (dc6d2a54): the table is not needed
DROP VIEW governor_withheld_runs_recent;
DROP TABLE governor_withheld_runs;

-- ---------------------------------------------------------------- verify (DO/RAISE)
DO $$
DECLARE cfg jsonb; q text; n int; adm boolean; lvl int; body text; saved_level int; saved_enabled boolean;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;

  -- shape
  IF cfg#>>'{workflow,start_step}' <> 'gate_spend_governor' THEN RAISE EXCEPTION '752 VERIFY: start_step is %', cfg#>>'{workflow,start_step}'; END IF;
  SELECT count(*) INTO n FROM jsonb_object_keys(cfg#>'{workflow,steps}');
  IF n <> 48 THEN RAISE EXCEPTION '752 VERIFY: expected 48 steps, found %', n; END IF;
  IF cfg#>>'{workflow,steps,gate_spend_governor,action}' <> 'query_database'
     OR cfg#>>'{workflow,steps,gate_spend_governor,next_step}' <> 'route_spend_governor'
     OR cfg#>>'{workflow,steps,gate_spend_governor,error_step}' <> 'load_schema_hint'
     OR cfg#>>'{workflow,steps,gate_spend_governor,output_field}' <> 'governor' THEN
    RAISE EXCEPTION '752 VERIFY: gate_spend_governor is mis-wired (action/next/error/output)';
  END IF;
  IF cfg#>>'{workflow,steps,route_spend_governor,config,condition}' <> 'governor.admitted == true'
     OR cfg#>>'{workflow,steps,route_spend_governor,config,then_step}' <> 'load_schema_hint'
     OR cfg#>>'{workflow,steps,route_spend_governor,config,else_step}' <> 'note_withheld' THEN
    RAISE EXCEPTION '752 VERIFY: route_spend_governor is mis-wired';
  END IF;
  IF cfg#>>'{workflow,steps,note_withheld,next_step}' <> 'complete_withheld'
     OR cfg#>>'{workflow,steps,note_withheld,config,subject_key}' <> 'spend-governor'
     OR cfg#>>'{workflow,steps,note_withheld,config,note_body_field}' <> 'governor.body' THEN
    RAISE EXCEPTION '752 VERIFY: note_withheld is mis-wired';
  END IF;
  IF cfg#>>'{workflow,steps,complete_withheld,action}' <> 'complete_workflow' THEN
    RAISE EXCEPTION '752 VERIFY: complete_withheld is not a terminal';
  END IF;
  -- the old start step is untouched and still points where it did
  IF cfg#>>'{workflow,steps,load_schema_hint,action}' <> 'query_database'
     OR cfg#>>'{workflow,steps,load_schema_hint,next_step}' <> 'persist_submission' THEN
    RAISE EXCEPTION '752 VERIFY: load_schema_hint was disturbed';
  END IF;
  IF jsonb_typeof(cfg#>'{workflow,steps,gate_spend_governor,config,params}') <> 'array'
     OR cfg#>>'{workflow,steps,gate_spend_governor,config,params,0}' <> '$ctx.correlation_id' THEN
    RAISE EXCEPTION '752 VERIFY: the gate does not bind $ctx.correlation_id';
  END IF;

  -- the gate's SQL RUNS and DISCRIMINATES (a literal stands in for the $ctx binding, which is runtime Go)
  q := replace(cfg#>>'{workflow,steps,gate_spend_governor,config,query}', '$1', '''752-verify-corr''');
  SELECT shed_level INTO saved_level FROM governor_state WHERE id=1;
  SELECT enabled INTO saved_enabled FROM governor_config WHERE id=1;
  EXECUTE 'SELECT admitted, shed_level, body FROM (' || q || ') g' INTO adm, lvl, body;
  IF adm IS DISTINCT FROM governor_admits_agent('council-gate') THEN
    RAISE EXCEPTION '752 VERIFY: gate query disagrees with governor_admits_agent at the live level';
  END IF;
  UPDATE governor_config SET enabled = true WHERE id=1;
  UPDATE governor_state SET shed_level = 3 WHERE id=1;
  EXECUTE 'SELECT admitted, shed_level, body FROM (' || q || ') g' INTO adm, lvl, body;
  IF adm IS DISTINCT FROM false OR lvl <> 3 OR position('WITHHELD at shed level 3' in body) = 0
     OR position('752-verify-corr' in body) = 0 THEN
    RAISE EXCEPTION '752 VERIFY: at L3 the gate should refuse and say so; got admitted=% level=% body=%', adm, lvl, left(body,120);
  END IF;
  UPDATE governor_state SET shed_level = 0 WHERE id=1;
  EXECUTE 'SELECT admitted FROM (' || q || ') g' INTO adm;
  IF adm IS DISTINCT FROM true THEN RAISE EXCEPTION '752 VERIFY: at L0 the gate should admit'; END IF;
  UPDATE governor_state SET shed_level = saved_level WHERE id=1;
  UPDATE governor_config SET enabled = saved_enabled WHERE id=1;

  -- fleet negative control: no OTHER live agent row carries the gate step
  SELECT count(*) INTO n FROM agent_definitions
  WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND type <> 'council-gate' AND default_config#>'{workflow,steps,gate_spend_governor}' IS NOT NULL;
  IF n <> 0 THEN RAISE EXCEPTION '752 VERIFY: % other agent rows carry gate_spend_governor', n; END IF;

  -- the table is gone, the predicate is not
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='governor_withheld_runs') THEN
    RAISE EXCEPTION '752 VERIFY: governor_withheld_runs still exists';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits_agent') THEN
    RAISE EXCEPTION '752 VERIFY: governor_admits_agent vanished';
  END IF;

  RAISE NOTICE '752 OK: council-gate starts at gate_spend_governor (48 steps, load_schema_hint untouched); gate SQL runs and flips at L3 with the correlation id in the body; fail-open error_step set; no other row carries it; governor_withheld_runs dropped. Council runs are now refused whenever governor_admits_agent(council-gate) is false — currently at shed level 3.';
END $$;

COMMIT;
