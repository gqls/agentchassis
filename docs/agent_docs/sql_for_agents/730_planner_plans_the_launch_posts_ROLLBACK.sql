-- 730_planner_plans_the_launch_posts_ROLLBACK.sql
-- Anchored reversal; refuses unless rule 20 present exactly once. Roundtrip proven byte-exact.

BEGIN;

DO $$
BEGIN
  PERFORM snapshot_agent('build-site-planner', '730_ROLLBACK: pre-rollback');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  anchor_a text := $R730$right to be all prose.

20. THERE IS NO LATER EDITORIAL PASS. Nothing outside this plan creates blog-post pages: a deferral such as "posts are created editorially" or "satisfied by the blog infrastructure" plans an EMPTY articles hub that rule 3 will hold back, and no later system fills it — every live article page on the estate was planned as an ordinary page in a plan like this one. So when your architecture includes a blog-index, an articles hub, or any section-index whose children are articles, plan the INDIVIDUAL launch posts in THIS plan: three to six blog-post pages, each on a REAL subject drawn from the briefing, strategy or vertical landscape for THIS site (never a placeholder, never a subject copied from an example), each with a populated sections array carrying a per-post "subject" (the empty-sections allowance in rule 3 is for pages other systems render, not for launch posts), in_header false, nav_order 200 or higher — the shape of every working article page on the estate. If the brief genuinely wants no articles at launch, plan no articles hub either. Deferring the posts while planning the hub is the one wrong answer.

Return ONLY valid JSON.$R730$;
  repl_a  text := $A730$right to be all prose.

Return ONLY valid JSON.$A730$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  n := (length(tpl) - length(replace(tpl, anchor_a, ''))) / length(anchor_a);
  IF n <> 1 THEN RAISE EXCEPTION '730R: rule-20 block found % times, expected 1 (later edit overlapped? resolve from snapshot)', n; END IF;
  newtpl := replace(tpl, anchor_a, repl_a);
  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '730R: updated % rows', n; END IF;
END $do$;

DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('THERE IS NO LATER EDITORIAL PASS' in tpl) > 0 THEN
    RAISE EXCEPTION '730R VERIFY: rule 20 still present';
  END IF;
  RAISE NOTICE '730R OK: rule 20 removed.';
END $$;

COMMIT;
