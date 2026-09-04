-- 755_d4b_gate_composes_the_withheld_body_only_when_withholding.sql — a misleading artefact
-- of the council gate (752/754), found by the 2026-09-04 11:21Z incident.
--
-- WHAT HAPPENED. The Anthropic account's prepaid credit ran out at 11:21:12Z; every LLM call
-- fleet-wide failed with HTTP 400 "Your credit balance is too low" until 11:56:48Z. Council
-- reviewer seats returned nothing readable, diagnose_council_decide errored, and six council
-- runs ended at complete_invalid. A peer lane, reading one of those runs' collected_data, found
-- `governor.body = "spend-governor: council-gate run … WITHHELD at shed level 0 …"` beside
-- `admitted: true` and reported — reasonably — that the gate was withholding at level 0. It was
-- not: the trail on every run reads gate → route → load_schema_hint → … → council_decide. The
-- gate had ADMITTED them. But the gate's query composed the WITHHELD text unconditionally, so
-- every admitted run carried a sentence saying it had been withheld. That text misled a
-- careful reader and cost an emergency disarm (12:07:26Z) of a gate that was innocent.
--
-- THE FIX: the body is NULL when admitted. note_withheld (the only reader) runs only on the
-- else branch, so a NULL body on the admit path is never read. One token of logic; the
-- remedy text also stops telling a reader to wait for the level to "drop" when it is 0.
-- GUARD: md5-pins the post-754 gate query (77336786e5c169402211ee886d998f60); refuses on
-- replay. VERIFY: PREPARE (the binding), L0 → body IS NULL, forced L3 → body present and
-- naming the correlation and level; row otherwise untouched. Every column of governor_state
-- restored as found. Rollback: 755_..._ROLLBACK.sql.

BEGIN;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF q IS NULL THEN RAISE EXCEPTION '755 REFUSED: no gate_spend_governor step — 752 not applied.'; END IF;
  IF position('CASE WHEN COALESCE(governor_admits_agent' in q) > 0 THEN RAISE EXCEPTION '755 REFUSED: body is already conditional — already applied (replay).'; END IF;
  IF md5(q) <> '77336786e5c169402211ee886d998f60' THEN
    RAISE EXCEPTION '755 REFUSED: gate query md5 % is not the post-754 text this was written against — drifted, investigate.', md5(q);
  END IF;
  PERFORM snapshot_agent('council-gate', '755 pre-apply: withheld body composed only when withholding');
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,gate_spend_governor,config,query}',
  to_jsonb($Q$SELECT COALESCE(governor_admits_agent('council-gate'), true) AS admitted, COALESCE(gs.shed_level, 0) AS shed_level, CASE WHEN COALESCE(governor_admits_agent('council-gate'), true) THEN NULL ELSE format('spend-governor: council-gate run for submission %s WITHHELD at shed level %s (%s%% of budget spent) - NOT queued; do not retry; re-trigger when governor_state.shed_level is below council-gate''s threshold in governor_agent_class_map (or when the budget is raised). RFC_065.', $1::text, COALESCE(gs.shed_level, 0), COALESCE(round(100*gs.mtd_usd/NULLIF(gc.monthly_budget_usd,0)), 0)) END AS body FROM (SELECT 1) always_one_row LEFT JOIN governor_state gs ON gs.id=1 LEFT JOIN governor_config gc ON gc.id=1$Q$::text))
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE q text; n int; adm boolean; lvl int; body text;
        s_level int; s_month text; s_mtd numeric; s_tok bigint; s_at timestamptz; c_enabled boolean;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('CASE WHEN COALESCE(governor_admits_agent' in q) = 0 THEN RAISE EXCEPTION '755 VERIFY: the conditional body did not land'; END IF;
  n := (length(q) - length(replace(q, '$1::text', ''))) / length('$1::text');
  IF n <> 1 THEN RAISE EXCEPTION '755 VERIFY: $1::text appears % times, expected 1', n; END IF;

  -- the BINDING (754's lesson): PREPARE with an unspecified parameter type
  EXECUTE 'PREPARE p755 AS ' || q;
  EXECUTE 'DEALLOCATE p755';

  SELECT shed_level, month, mtd_usd, unpriced_io_tokens, computed_at INTO s_level, s_month, s_mtd, s_tok, s_at FROM governor_state WHERE id=1;
  SELECT enabled INTO c_enabled FROM governor_config WHERE id=1;
  UPDATE governor_config SET enabled = true WHERE id=1;

  -- L0: admitted, and the body is NULL — nothing in the blob can read as a withholding
  UPDATE governor_state SET shed_level = 0 WHERE id=1;
  EXECUTE 'SELECT admitted, body FROM (' || replace(q, '$1::text', '''755-verify-corr''') || ') g' INTO adm, body;
  IF adm IS DISTINCT FROM true OR body IS NOT NULL THEN
    RAISE EXCEPTION '755 VERIFY: at L0 expected admitted=true, body NULL; got admitted=% body=%', adm, left(coalesce(body,'<null>'),80);
  END IF;
  -- L3: refused, and the body names the correlation and the level
  UPDATE governor_state SET shed_level = 3 WHERE id=1;
  EXECUTE 'SELECT admitted, shed_level, body FROM (' || replace(q, '$1::text', '''755-verify-corr''') || ') g' INTO adm, lvl, body;
  IF adm IS DISTINCT FROM false OR lvl <> 3 OR body IS NULL OR position('755-verify-corr' in body) = 0 OR position('WITHHELD at shed level 3' in body) = 0 THEN
    RAISE EXCEPTION '755 VERIFY: at L3 expected admitted=false and a body naming the correlation and level 3; got admitted=% body=%', adm, left(coalesce(body,'<null>'),120);
  END IF;
  -- still exactly one row with a missing state row (752's fail-open)
  DELETE FROM governor_state WHERE id=1;
  EXECUTE 'SELECT count(*), bool_and(admitted) FROM (' || replace(q, '$1::text', '''755-verify-corr''') || ') g' INTO n, adm;
  IF n <> 1 OR adm IS DISTINCT FROM true THEN RAISE EXCEPTION '755 VERIFY: with governor_state missing: % row(s), admitted=% (need 1, true)', n, adm; END IF;
  INSERT INTO governor_state (id, shed_level, month, mtd_usd, unpriced_io_tokens, computed_at) VALUES (1, s_level, s_month, s_mtd, s_tok, s_at);
  UPDATE governor_config SET enabled = c_enabled WHERE id=1;

  -- row otherwise untouched
  SELECT count(*) INTO n FROM agent_definitions, jsonb_object_keys(default_config#>'{workflow,steps}') k
   WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 48 THEN RAISE EXCEPTION '755 VERIFY: step count % (expected 48)', n; END IF;
  RAISE NOTICE '755 OK: withheld body composed only when withholding (L0 body NULL; L3 body names corr + level); PREPARE clean; fail-open on a missing row kept; governor_state restored column-for-column.';
END $$;

COMMIT;
