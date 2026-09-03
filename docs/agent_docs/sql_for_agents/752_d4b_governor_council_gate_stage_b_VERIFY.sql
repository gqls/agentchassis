-- 752_d4b_governor_council_gate_stage_b_VERIFY.sql — the DAILY parity check the guardian asked
-- for on corr dc6d2a54. Not a migration; run it beside 657's and 584's:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--     -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/752_d4b_governor_council_gate_stage_b_VERIFY.sql
--
-- WHAT IT GUARDS. 099_SYNC_gate_roster.py --apply regenerates the council-gate row's whole
-- workflow from the fix-proposer mirror; the four D4b steps live nowhere in that mirror, so a
-- regeneration silently UNGOVERNS the council and the row reads as perfectly healthy. So does
-- any hand edit that repoints start_step. This check reads the live row, not the migration.
-- ⚠ FAILS BY DESIGN before 752 is applied — that failure is its mutation proof. Do not "fix" it
-- by widening; apply 752 (after the owner's level answer) or accept the red.
-- Exit 0 = the gate is in place and wired; a RAISE names what moved.

DO $$
DECLARE cfg jsonb; q text; n int; adm boolean; lvl int;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF cfg IS NULL THEN RAISE EXCEPTION '752 VERIFY 1/6: no live council-gate row'; END IF;

  -- 1. the governor gate IS the start step (the 099-regeneration failure mode)
  IF cfg#>>'{workflow,start_step}' <> 'gate_spend_governor' THEN
    RAISE EXCEPTION '752 VERIFY 1/6: council-gate start_step is % — the spend-governor gate is NOT the entry point (099 --apply or a hand edit removed it; the council is UNGOVERNED)', cfg#>>'{workflow,start_step}';
  END IF;

  -- 2. all four steps present and wired to each other and to the old start step
  IF cfg#>'{workflow,steps,gate_spend_governor}' IS NULL OR cfg#>'{workflow,steps,route_spend_governor}' IS NULL
     OR cfg#>'{workflow,steps,note_withheld}' IS NULL OR cfg#>'{workflow,steps,complete_withheld}' IS NULL THEN
    RAISE EXCEPTION '752 VERIFY 2/6: one or more D4b steps missing from council-gate';
  END IF;
  IF cfg#>>'{workflow,steps,gate_spend_governor,next_step}' <> 'route_spend_governor'
     OR cfg#>>'{workflow,steps,route_spend_governor,config,then_step}' <> 'load_schema_hint'
     OR cfg#>>'{workflow,steps,route_spend_governor,config,else_step}' <> 'note_withheld'
     OR cfg#>>'{workflow,steps,note_withheld,next_step}' <> 'complete_withheld' THEN
    RAISE EXCEPTION '752 VERIFY 2/6: D4b steps present but re-wired';
  END IF;

  -- 3. fail-open is still declared (an unreadable governor must RUN the review)
  IF cfg#>>'{workflow,steps,gate_spend_governor,error_step}' <> 'load_schema_hint' THEN
    RAISE EXCEPTION '752 VERIFY 3/6: gate_spend_governor.error_step is % — fail-open lost', COALESCE(cfg#>>'{workflow,steps,gate_spend_governor,error_step}','<absent>');
  END IF;

  -- 4. the gate still asks the ONE predicate, and binds the run's own correlation
  q := cfg#>>'{workflow,steps,gate_spend_governor,config,query}';
  IF position('governor_admits_agent(''council-gate'')' in q) = 0 THEN
    RAISE EXCEPTION '752 VERIFY 4/6: the gate query no longer calls governor_admits_agent(''council-gate'')';
  END IF;
  IF cfg#>>'{workflow,steps,gate_spend_governor,config,params,0}' <> '$ctx.correlation_id' THEN
    RAISE EXCEPTION '752 VERIFY 4/6: the gate no longer binds $ctx.correlation_id';
  END IF;

  -- 5. the predicate exists and the gate SQL still executes against the live schema
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits_agent') THEN
    RAISE EXCEPTION '752 VERIFY 5/6: governor_admits_agent() is gone — 751 rolled back under a live caller';
  END IF;
  EXECUTE 'SELECT admitted, shed_level FROM (' || replace(q, '$1', '''752-daily-verify''') || ') g' INTO adm, lvl;
  IF adm IS DISTINCT FROM governor_admits_agent('council-gate') THEN
    RAISE EXCEPTION '752 VERIFY 5/6: gate SQL and governor_admits_agent disagree at the live level';
  END IF;

  -- 6. council-gate is still mapped (an unmapped agent is ADMITTED forever — the gate would be a no-op)
  SELECT count(*) INTO n FROM governor_agent_class_map WHERE agent_type='council-gate';
  IF n <> 1 THEN
    RAISE EXCEPTION '752 VERIFY 6/6: council-gate has % rows in governor_agent_class_map — the gate step is live but can never refuse', n;
  END IF;

  RAISE NOTICE '752 VERIFY OK: council-gate enters at gate_spend_governor, wired, fail-open, asks governor_admits_agent, binds $ctx.correlation_id, mapped (class=%, admitted now=%, level=%).',
    (SELECT class FROM governor_agent_class_map WHERE agent_type='council-gate'), adm, lvl;
END $$;
