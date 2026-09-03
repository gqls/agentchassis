-- gamedesign.uk — enqueue the briefing -> site_plan rebuild chain that the strategist's
-- B2 refresh-safety gate deliberately withheld (2026-09-03 ~08:40Z).
--
-- WHY: 082 re-submission (corr aab87c0c, brief v2, 2026-09-02 20:10:59Z) ran classifier ->
-- vertical research -> strategist; the strategist's `gate_next_item` step
-- ("site_state.is_deployed == true" -> complete, else create_next_item; description: "a deployed
-- site's strategy refresh must NOT enqueue the briefing->site-plan rebuild chain") completed at
-- 20:26:45 WITHOUT filing needs_briefing, because gamedesign.uk is deployed. So 082 on a live site
-- is a STRATEGY REFRESH, not a rebuild. The owner has ruled the design and copy must change
-- (bugs_open/446), strategy v2 is current (20:26:40), so the chain is enqueued here on purpose,
-- in the strategist's own shape (17:23 row: source domain-strategist, build, high, prio 10,
-- handler build-briefing-agent, key briefing_gamedesign.uk, spec {}). The prior row with that key is
-- terminal (complete), so idx_swi_dedup allows it, exactly as strategy_gamedesign.uk was allowed.
-- Downstream: build-briefing-agent -> needs_site_plan -> build-site-planner (post-mig-718) ->
-- reconcile_site_plan diffs against the 4 realised pages -> composition/design/pages/rerender.
-- growth_posture='hold' stays on: evaluate_tools/add_tool file as records.
\set ON_ERROR_STOP on
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'domain-strategist', 'build', 'needs_briefing', 'high',
  'Briefing needed after domain strategy v2 (owner-directed rebuild, bugs_open/446; the deployed-site gate withheld this on purpose and it is enqueued on purpose)',
  '{"reason":"owner_directed_rebuild_brief_v2","strategy_spec_at":"2026-09-02T20:26:40Z","withheld_by":"domain-strategist gate_next_item (is_deployed)","lane":"gamedesign_uk_rebuild"}'::jsonb,
  10, 'build-briefing-agent', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-03', 'briefing_gamedesign.uk')
RETURNING id, item_type, status, created_at::timestamp(0);
