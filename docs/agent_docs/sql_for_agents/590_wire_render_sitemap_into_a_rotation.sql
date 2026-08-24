-- 590 — WIRE render_sitemap INTO SOMETHING THAT ACTUALLY RUNS.
--
-- WHY. The council reviewed the change that built `render_sitemap` and returned
-- REVISE at high severity (correlation 8a004aab-be85-4d6d-bdb1-4fb114f1d64b) on
-- one objection, which was right:
--
--   "The whole rationale for this change is that a working generator which
--    nothing invokes is 'not a mechanism'... A registered-but-uncalled action
--    reproduces the diagnosed defect in a new form."
--
-- The action has been registered and callable since a596fa84b (2026-08-21) and
-- **nothing has ever called it**. This migration is the caller.
--
-- THE BEFORE-FIGURE, re-measured 2026-08-24 by fetching every live site and
-- READING THE BODY (a status code is not evidence — see the two traps below):
-- **8 of 28** live sites serve a sitemap of ours. That is the number this is
-- judged against, and re-measuring it after is the proof.
--
--   ⚠ TWO sites answer 200 on /sitemap.xml with something that is NOT ours:
--     adversecreditmortgage.co.uk returns the PARKING PROVIDER's file (171 bytes,
--     one <loc> for /lander, no matching `pages` row), and noted.co.uk returns
--     27,414 bytes of text/html — its own homepage, served for any path. A third,
--     webdesign.uk, 302s to webdesign.co.uk. A status-code census counts all
--     three as successes. The mechanical test used instead: extract every <loc>
--     and match its path against that site's `pages` rows. Ours score 17/18 to
--     98/98; the parking file scores 0/1.
--
-- WHY A ROTATION AND NOT THE PAGE-DEPLOY PATH. Both are wanted eventually (a new
-- page and a retracted page are different events). The rotation goes first
-- because it is the one that can reach the 20 sites that have no sitemap AT ALL
-- — wiring the deploy path only helps a site that is rebuilt — and because the
-- cost is bounded. Measured 2026-08-24: 735 listable pages across 28 live sites,
-- average 26, maximum 135. The probe is a GET per URL, so one site per tick is
-- ~26 requests; the deploy path would re-probe a whole site on every page change
-- (135 requests for webdesign.co.uk, every time). The deploy-path half is left
-- for a second, separate change, once this one is proven at the artefact.
--
-- WHAT THIS DOES NOT DO. It does not touch the 20 sites' page content, and it
-- cannot publish an empty sitemap: `render_sitemap` returns rendered=false /
-- url_count=0 when a site has no listable URLs or has opted out, and the
-- conditional below routes that to `complete` without a commit. An empty sitemap
-- tells a crawler the site has no pages, which is worse than none.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. The agent. Four steps, and deliberately the SAME SHAPE as the RSS trio in
--    content-feed-orchestrator (render → conditional → git_commit), because
--    render_sitemap was written to return render_rss_feed's {files, domain}
--    contract precisely so this would work unchanged.
--
--    NOTE there is no `ensure_site_record` step, unlike render-audit-agent. That
--    action can CREATE a site row, and a sweep must never bring a site into
--    existence as a side effect of refreshing its sitemap. The rotation's
--    pre_query already yields site_id, which the scheduler merges into
--    input_data (cmd/scheduler/main.go:223-229).
-- ---------------------------------------------------------------------------
INSERT INTO agent_definitions (type, display_name, category, description, default_config, is_active)
SELECT
  'sitemap-refresh',
  'Sitemap refresh',
  'specialist',
  'Regenerates and commits sitemap.xml for one site. Driven by the sitemap-refresh-rotation scheduled task, one site per tick.',
  jsonb_build_object(
    'input_schema', jsonb_build_object('required', jsonb_build_array('site_id')),
    'workflow', jsonb_build_object(
      'start_step', 'render_sitemap',
      'steps', jsonb_build_object(

        'render_sitemap', jsonb_build_object(
          'action', 'render_sitemap',
          'description', 'Render sitemap.xml from active, indexable, deployed pages. ON by default; a site opts out with deploy_config.sitemap.enabled=false. Probes every URL and lists only 2xx.',
          'config', jsonb_build_object('site_id', 'input_data.site_id'),
          'output_field', 'sitemap_render_result',
          'next_step', 'check_has_urls'),

        'check_has_urls', jsonb_build_object(
          'action', 'evaluate_condition',
          'description', 'url_count is 0 in BOTH no-op cases — the site opted out, or nothing was listable. Either way there is nothing to commit, and publishing an empty sitemap actively misinforms a crawler.',
          'config', jsonb_build_object(
            'condition_field', 'sitemap_render_result.url_count',
            'conditions', jsonb_build_object('0', 'complete'),
            'default', 'commit_sitemap')),

        -- ⚠ DO NOT ADD 'repo_name' TO THIS CONFIG. Its ABSENCE is what makes the
        -- step correct for all 28 sites. resolveGitRepoNameDB (helpers.go:232)
        -- tries explicit config['repo_name'] FIRST, then site_record, then
        -- sites.github_repo by domain, then 'sites'. 4 of the 28 live sites are
        -- vm-sites (idea.uk, noted.co.uk, relojistas.com, webdesign.uk as of
        -- 2026-08-24) and the other 24 have an empty github_repo. Omitting the key
        -- lets each site resolve its own repo; setting it to 'sites' would send
        -- those 4 to the WRONG repo, and LANDMINES records exactly that failure
        -- for the hand-run script that hardcodes it — kcat exits 0, the adapter
        -- logs no error, GitHub shows the commit, and the served file never
        -- changes. Raised by the council's guardian seat, corr 8a004aab.
        'commit_sitemap', jsonb_build_object(
          'action', 'git_commit',
          'description', 'Commit sitemap.xml. Reaches B2 sites as well as git-hosted ones: resolveGitRepoNameDB (helpers.go:236-246) reads sites.github_repo and defaults to `sites` when empty, which is what every B2 domain has.',
          'config', jsonb_build_object(
            'files_field', 'sitemap_render_result.files',
            'domain_field', 'sitemap_render_result.domain',
            'commit_message', 'Update sitemap.xml'),
          'output_field', 'sitemap_commit_result',
          'next_step', 'complete'),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Sitemap refresh complete.',
          'config', jsonb_build_object(
            'multiple_output_fields', jsonb_build_array('sitemap_render_result', 'sitemap_commit_result')))
      ))),
  true
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
   WHERE type='sitemap-refresh' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

-- ---------------------------------------------------------------------------
-- 2. The clock. Rotation over sites, stamped in site_discovery_rotation — the
--    same table render-audit-agent uses, which is already precedent for a
--    non-discovery agent_type living there.
--
--    ⚠ site-discovery-staleness-check reads this table. Checked 2026-08-24: its
--    coverage query CROSS JOINs its own three DISCOVERY_AGENTS and LEFT JOINs on
--    agent_type, and its findings loop iterates that same list, so a fourth
--    agent_type CANNOT create a false finding. It appears only in the
--    informational `stamps_last_24h` line of the report.
--
--    locked_at IS NULL is load-bearing, not boilerplate. adversecreditmortgage.co.uk
--    is locked by an owner HALT (2026-08-18) with its queued items held; a
--    sitemap commit is a deploy, and this sweep must not drive one against a
--    halt. Same idiom as build-pipeline-trigger.
--
--    Interval and threshold: 1800s ticks, one site per tick, refreshing a site
--    whose stamp is older than 3 days. First sweep covers all 28 in ~14 hours;
--    steady state is ~9 sites/day due, i.e. ~245 probe GETs/day fleet-wide.
-- ---------------------------------------------------------------------------
INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds, fire_message)
SELECT
  'sitemap-refresh-rotation',
  'Regenerate and commit sitemap.xml for one due site per tick. Answers the council REVISE on 8a004aab: render_sitemap had no caller.',
  1800,
  'sitemap-refresh',
  'system.agent.scheduled.requests',
  '{}'::jsonb,
  'sitemap-refresh',
  1,
  $PQ$
WITH due AS (
  SELECT s.id AS sid, s.domain
  FROM sites s
  LEFT JOIN site_discovery_rotation r
    ON r.site_id = s.id AND r.agent_type = 'sitemap-refresh'
  WHERE s.status IN ('active', 'deployed')
    AND s.locked_at IS NULL
    AND COALESCE(r.last_selected_at, '-infinity'::timestamptz) < now() - interval '3 days'
    AND NOT EXISTS (
      SELECT 1 FROM site_work_items wi
      WHERE wi.site_id = s.id AND wi.status = 'claimed' AND wi.pipeline = 'build')
  ORDER BY r.last_selected_at ASC NULLS FIRST, s.id
  LIMIT 1
), stamped AS (
  INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
  SELECT sid, 'sitemap-refresh', now() FROM due
  ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at = EXCLUDED.last_selected_at
)
SELECT sid::text AS site_id, domain FROM due
$PQ$,
  true,
  900,
  true
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='sitemap-refresh-rotation');

-- ---------------------------------------------------------------------------
-- 3. Verify. DO/RAISE, not bare SELECTs: ON_ERROR_STOP does not abort a COMMIT
--    on a non-empty result set, so a verify block of SELECTs verifies nothing.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    n_steps int; n_task int; v_action text; v_default text; v_files text; v_target text; v_enabled boolean;
BEGIN
    SELECT count(*) INTO n_steps FROM agent_definitions,
         LATERAL jsonb_object_keys(default_config #> '{workflow,steps}')
     WHERE type='sitemap-refresh' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_steps <> 4 THEN RAISE EXCEPTION '590: expected 4 steps on sitemap-refresh, found %', n_steps; END IF;

    -- The action name is the whole point of the change: a typo here is the
    -- registered-but-uncalled defect all over again, and nothing would fail
    -- until the first tick.
    SELECT default_config #>> '{workflow,steps,render_sitemap,action}' INTO v_action
      FROM agent_definitions WHERE type='sitemap-refresh' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_action IS DISTINCT FROM 'render_sitemap' THEN
      RAISE EXCEPTION '590: start step calls %, not render_sitemap', v_action;
    END IF;

    -- If the conditional's default does not reach the commit step, the sweep
    -- renders a perfect sitemap every 30 minutes and publishes none of them.
    SELECT default_config #>> '{workflow,steps,check_has_urls,config,default}' INTO v_default
      FROM agent_definitions WHERE type='sitemap-refresh' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_default IS DISTINCT FROM 'commit_sitemap' THEN
      RAISE EXCEPTION '590: conditional default is %, so nothing is ever committed', v_default;
    END IF;

    SELECT default_config #>> '{workflow,steps,commit_sitemap,config,files_field}' INTO v_files
      FROM agent_definitions WHERE type='sitemap-refresh' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_files IS DISTINCT FROM 'sitemap_render_result.files' THEN
      RAISE EXCEPTION '590: commit reads files from %, which is not what render_sitemap writes', v_files;
    END IF;

    SELECT count(*), max(target_agent_type), bool_or(enabled) INTO n_task, v_target, v_enabled
      FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';
    IF n_task <> 1 THEN RAISE EXCEPTION '590: expected 1 rotation task, found %', n_task; END IF;
    IF v_target IS DISTINCT FROM 'sitemap-refresh' THEN
      RAISE EXCEPTION '590: rotation targets %, which has no workflow', v_target;
    END IF;
    IF NOT v_enabled THEN RAISE EXCEPTION '590: rotation task is disabled — it would never fire'; END IF;

    -- The lock guard protects an owner HALT. Assert it is present rather than
    -- trusting that it was typed.
    PERFORM 1 FROM scheduled_tasks
      WHERE name='sitemap-refresh-rotation' AND pre_query LIKE '%locked_at IS NULL%';
    IF NOT FOUND THEN
      RAISE EXCEPTION '590: pre_query does not exclude locked sites — it could deploy against an owner halt';
    END IF;

    RAISE NOTICE '590 OK — sitemap-refresh seeded (4 steps) and sitemap-refresh-rotation enabled at 1800s.';
END $$;

COMMIT;
