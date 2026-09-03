-- 730_planner_plans_the_launch_posts.sql
--
-- The planner REFUSES to plan blog-post pages, deferring to an editorial pass that DOES NOT
-- EXIST — measured by the gamedesign.uk session (CONTRIB in bugs_open/444, commit 7343ecb01):
-- 3 of 32 plan_site runs in 30 days wrote the deferral into their own strategy_notes
-- (designblog.co.uk 2026-09-02 16:10:51Z "posts are created editorially"; seotools.co.uk +3min;
-- gamedesign.uk twice), while every live article page on the estate is an ordinary PLANNED page
-- (webdesign.co.uk 52, dartsonline 23, finetuning 22, gamesdesign 13 — census framed so a
-- non-plan source would have falsified it; none did; needs_content_page only BUILDS planned
-- pages). There is NO per-site lever: gamedesign's mission v3 said in plain words "The site
-- launches with real articles… A page that lists articles must list articles", VERIFIED to
-- reach the rendered prompt, and the planner planned zero anyway. So the fix lands here, in the
-- prompt row. OWNER RULING 2026-09-03 ("5: both") sanctions the producer work; for articles the
-- producer EXISTS (planner + writer) and the gap is this invocation refusal — the cheap arm.
--
-- The edit: ONE anchored insertion — new rule 20 after rule 19, before "Return ONLY valid
-- JSON." It names the two observed deferral phrasings, states the measured fact (no later
-- pass), instructs planning 3-6 launch posts on REAL subjects with the shape every working
-- article page has, forbids subject-copying from examples (the quoted-exemplar trap is why
-- this ships as a RULE, not a JSON example — an example subject would ship verbatim onto
-- wrong verticals), and states the honest alternative (no articles wanted -> no hub either).
-- It COMPOSES with 720's rule-3 gate: the gate mechanically holds a hub with no children;
-- rule 20 steers the planner to the right side of that fork (plan the children) instead of
-- the silent third path (defer and ship prose).
--
-- Same shared row two lanes edit: anchored replace with exact-count guards, snapshot,
-- DO/RAISE verify, byte-exact-roundtrip ROLLBACK (proven locally), dry-run before submit.
-- Apply: psql -f THIS FILE ONLY.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '730 REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner', '730_planner_plans_the_launch_posts.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs int; ends int; elses int; vars int;
  anchor_a text := $A730$right to be all prose.

Return ONLY valid JSON.$A730$;
  repl_a  text := $R730$right to be all prose.

20. THERE IS NO LATER EDITORIAL PASS. Nothing outside this plan creates blog-post pages: a deferral such as "posts are created editorially" or "satisfied by the blog infrastructure" plans an EMPTY articles hub that rule 3 will hold back, and no later system fills it — every live article page on the estate was planned as an ordinary page in a plan like this one. So when your architecture includes a blog-index, an articles hub, or any section-index whose children are articles, plan the INDIVIDUAL launch posts in THIS plan: three to six blog-post pages, each on a REAL subject drawn from the briefing, strategy or vertical landscape for THIS site (never a placeholder, never a subject copied from an example), each with a populated sections array carrying a per-post "subject" (the empty-sections allowance in rule 3 is for pages other systems render, not for launch posts), in_header false, nav_order 200 or higher — the shape of every working article page on the estate. If the brief genuinely wants no articles at launch, plan no articles hub either. Deferring the posts while planning the hub is the one wrong answer.

Return ONLY valid JSON.$R730$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '730: prompt_template not found'; END IF;

  n := (length(tpl) - length(replace(tpl, anchor_a, ''))) / length(anchor_a);
  IF n <> 1 THEN RAISE EXCEPTION '730: anchor found % times, expected 1', n; END IF;

  ifs   := (length(tpl) - length(replace(tpl, '{{if ', ''))) / length('{{if ');
  ends  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses := (length(tpl) - length(replace(tpl, '{{else}}', ''))) / length('{{else}}');
  vars  := (length(tpl) - length(replace(tpl, '{{.', ''))) / length('{{.');

  newtpl := replace(tpl, anchor_a, repl_a);

  IF length(newtpl) <> length(tpl) + (length(repl_a) - length(anchor_a)) THEN
    RAISE EXCEPTION '730: unexpected length delta';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) / length('{{.') <> vars
     OR (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends
     OR (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses THEN
    RAISE EXCEPTION '730: template variable or if/else/end balance changed';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '730: updated % rows, expected exactly 1', n; END IF;
END $do$;

DO $$
DECLARE tpl text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  n := (length(tpl) - length(replace(tpl, 'THERE IS NO LATER EDITORIAL PASS', '')))
       / length('THERE IS NO LATER EDITORIAL PASS');
  IF n <> 1 THEN RAISE EXCEPTION '730 VERIFY: rule 20 not present exactly once (found %)', n; END IF;
  n := (length(tpl) - length(replace(tpl, 'Return ONLY valid JSON.', '')))
       / length('Return ONLY valid JSON.');
  IF n <> 1 THEN RAISE EXCEPTION '730 VERIFY: closing JSON directive count wrong (%)', n; END IF;
  IF position('19. MATCH STRUCTURE TO PROMISE' in tpl) = 0 THEN
    RAISE EXCEPTION '730 VERIFY: rule 19 damaged';
  END IF;
  RAISE NOTICE '730 OK: the planner now plans the launch posts.';
END $$;

COMMIT;
