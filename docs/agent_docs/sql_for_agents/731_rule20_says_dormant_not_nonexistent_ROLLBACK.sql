-- 731_rule20_says_dormant_not_nonexistent_ROLLBACK.sql
-- Restores 730's original wording (which 730's own ROLLBACK then matches again).

BEGIN;
DO $$ BEGIN PERFORM snapshot_agent('build-site-planner', '731_ROLLBACK: pre-rollback'); END $$;
DO $do$
DECLARE
  tpl text; newtpl text; n int;
  anchor_a text := $A731R$20. NO LATER EDITORIAL PASS RUNS — do not defer posts to one. The one mechanism that could create posts later (blog-content-planner, via create_blog_posts) has been DORMANT since 2026-04-24 (10 LLM calls all-history, none since; measured 2026-09-03), so a deferral$A731R$;
  repl_a  text := $RA731R$20. THERE IS NO LATER EDITORIAL PASS. Nothing outside this plan creates blog-post pages: a deferral$RA731R$;
  anchor_b text := $B731R$plans an EMPTY articles hub that rule 3 will hold back, and in practice nothing fills it — every$B731R$;
  repl_b  text := $RB731R$plans an EMPTY articles hub that rule 3 will hold back, and no later system fills it — every$RB731R$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '731R: prompt_template not found'; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_a, ''))) / length(anchor_a);
  IF n <> 1 THEN RAISE EXCEPTION '731R: anchor A found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_b, ''))) / length(anchor_b);
  IF n <> 1 THEN RAISE EXCEPTION '731R: anchor B found % times, expected 1', n; END IF;
  newtpl := replace(replace(tpl, anchor_a, repl_a), anchor_b, repl_b);
  IF length(newtpl) <> length(tpl) + (length(repl_a)-length(anchor_a)) + (length(repl_b)-length(anchor_b)) THEN
    RAISE EXCEPTION '731R: unexpected length delta';
  END IF;
  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '731R: updated % rows, expected exactly 1', n; END IF;
END $do$;
DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('THERE IS NO LATER EDITORIAL PASS' in tpl) = 0 THEN
    RAISE EXCEPTION '731R VERIFY: original wording not restored';
  END IF;
  RAISE NOTICE '731R OK.';
END $$;
COMMIT;
