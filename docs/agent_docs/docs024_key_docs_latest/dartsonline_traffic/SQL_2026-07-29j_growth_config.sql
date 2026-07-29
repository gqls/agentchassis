-- growth_config for dartsonline (owner decisions D3 + D5, 2026-07-29).
-- D3: backfill burst then ~4-5 articles/week. Building an ALREADY-PLANNED page is not
--     budget-gated (the budget governs page CREATION), so the nine guides drain freely;
--     this raises the cap for the NEW analysis posts that follow.
-- D5: tool-heavy, roughly 1 tool per 6 articles/guides. `content_tools_ratio` is read by
--     nothing yet — the reader is the check_missing_tools change described in PLAN
--     P-tools. Recorded here deliberately as the config surface that change will read,
--     so the policy and its future reader live in one place. Marked UNREAD so nobody
--     mistakes it for live behaviour (see WRONG_CALLS: a dead key looks exactly like a
--     live one).
-- Key names verified against page_growth_budget.go:82-86.
BEGIN;
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
VALUES ('5fe8785b-223d-41a3-88ee-c07187622381', 'growth_config',
  jsonb_build_object(
    'initial_target', 12,
    'weekly_content_pages_max', 3,
    'weekly_blog_posts_max', 5,
    'weekly_structural_pages_max', 3,
    'absolute_max', 60,
    'content_tools_ratio', 6,
    'content_tools_ratio_note', 'Owner decision D5 2026-07-29: roughly one interactive tool per six articles/guides. NOT YET READ BY ANY CODE — the intended reader is discovery_checks/check_missing_tools.go, which today evaluates tool need on a 7/30-day timer with no reference to how much content the site has.'
  ),
  'authored', 'dartsonline-traffic-workstream', true, 'dartsonline-traffic-workstream',
  'D3 cadence + D5 tool ratio. weekly_blog_posts_max raised 2->5 for the analysis lane.')
ON CONFLICT DO NOTHING;
COMMIT;
SELECT aspect, data->>'weekly_blog_posts_max' AS blog_max, data->>'content_tools_ratio' AS tool_ratio
FROM site_specs WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='growth_config' AND is_current=true;
