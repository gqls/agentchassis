-- 731_rule20_says_dormant_not_nonexistent.sql
--
-- CORRECTS 730 within the hour, on its own evidence-author's retraction (gamedesign.uk,
-- verified first-hand by them + the 427 lane): rule 20's flat "THERE IS NO LATER EDITORIAL
-- PASS" is TOO STRONG as a statement of fact. A blog-post producer EXISTS and is wired
-- (platform/orchestration/actions/create_blog_posts_action.go, registered registry.go:720,
-- named by exactly one live agent definition: blog-content-planner) — it is DORMANT, not
-- absent: 10 LLM calls all-history, first 2026-04-03, last 2026-04-24, none in four months
-- (llm_call_log — the instrument with real memory; orchestration_states is a rolling ~24h
-- window and cannot support all-history claims). The planner's "satisfied by the blog
-- infrastructure" named a real, wired, non-running mechanism: undriven, not hallucinated.
-- Rule 20's STEERING survives unchanged (no pass RUNS; deferring posts ships empty hubs —
-- the measured outcome on three sites). What must not stand is a live prompt asserting an
-- absolute that becomes false the moment anyone revives blog-content-planner. The fix:
-- say it RUNS not that it EXISTS, with the dormancy DATED so the claim is mechanically
-- re-checkable (the estate's count-carries-its-date rule, applied to a prompt).
-- ⚠ 730's ROLLBACK anchors on the original rule-20 text and will REFUSE after this applies
-- (correct: resolve from the snapshots, not by force). Apply: psql -f THIS FILE ONLY.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '731 REFUSED: expected exactly 1 active row, found %', n; END IF;
  PERFORM snapshot_agent('build-site-planner', '731_rule20_says_dormant_not_nonexistent.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  anchor_a text := $A731F$20. THERE IS NO LATER EDITORIAL PASS. Nothing outside this plan creates blog-post pages: a deferral$A731F$;
  repl_a  text := $RA731F$20. NO LATER EDITORIAL PASS RUNS — do not defer posts to one. The one mechanism that could create posts later (blog-content-planner, via create_blog_posts) has been DORMANT since 2026-04-24 (10 LLM calls all-history, none since; measured 2026-09-03), so a deferral$RA731F$;
  anchor_b text := $B731F$plans an EMPTY articles hub that rule 3 will hold back, and no later system fills it — every$B731F$;
  repl_b  text := $RB731F$plans an EMPTY articles hub that rule 3 will hold back, and in practice nothing fills it — every$RB731F$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '731F: prompt_template not found'; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_a, ''))) / length(anchor_a);
  IF n <> 1 THEN RAISE EXCEPTION '731F: anchor A found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_b, ''))) / length(anchor_b);
  IF n <> 1 THEN RAISE EXCEPTION '731F: anchor B found % times, expected 1', n; END IF;
  newtpl := replace(replace(tpl, anchor_a, repl_a), anchor_b, repl_b);
  IF length(newtpl) <> length(tpl) + (length(repl_a)-length(anchor_a)) + (length(repl_b)-length(anchor_b)) THEN
    RAISE EXCEPTION '731F: unexpected length delta';
  END IF;
  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '731F: updated % rows, expected exactly 1', n; END IF;
END $do$;

DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('THERE IS NO LATER EDITORIAL PASS' in tpl) > 0 THEN
    RAISE EXCEPTION '731 VERIFY: the false absolute still present';
  END IF;
  IF position('NO LATER EDITORIAL PASS RUNS' in tpl) = 0
     OR position('DORMANT since 2026-04-24' in tpl) = 0 THEN
    RAISE EXCEPTION '731 VERIFY: dated-dormancy form missing';
  END IF;
  RAISE NOTICE '731 OK: rule 20 now states dormancy with a date, not a false absolute.';
END $$;

COMMIT;
