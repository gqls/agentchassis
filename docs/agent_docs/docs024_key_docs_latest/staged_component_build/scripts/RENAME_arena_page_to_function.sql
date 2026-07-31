-- Close the last BROKEN-A case: make vonc.com's arena page resolvable by the
-- acceptance lookup, which is `name IN (function, 'tool-' || function)` scoped by
-- site_id AND status='active' (tool_acceptance_actions.go:140-146).
--
-- The component's function is `tool-arena-interface`; the page was named `tool-arena`,
-- which matches neither that nor `tool-tool-arena-interface`, so `request_browser_run`
-- hard-errored with "no deployed page URL". THE COMPONENT IS NOT AN ORPHAN — it is live,
-- deployed and serving at /tools/arena/index.html.
--
-- ⚠ TWO ROWS, NOT ONE, AND THIS IS THE POINT OF THIS FILE.
-- `check_sectionless_pages` (discovery_checks/check_sectionless_pages.go:118) joins
-- `site_plan_pages spp ON spp.name = p.name`. Renaming pages.name ALONE would
-- desynchronise that join and this page would SILENTLY LEAVE that detector's population
-- — it currently qualifies (0 sections) and is currently reported (work item 559cb636,
-- still `unresolved`). Losing a detection while "fixing" a naming defect is the exact
-- class this lane exists to prevent, so both name-side rows move together.
--
-- Measured before applying, all zero or accounted for:
--   pages.name='tool-arena-interface' on vonc.com ....... 0  (no collision)
--   site_plan_sections.page_name='tool-arena' .......... 0  (no sections to re-key)
--   site_plan_imagery.scope_ref='tool-arena' ........... 0  (imagery keys on scope_ref)
--   pages.status ....................................... active  (the lookup requires it)
--   page_components .................................... key on page_id, not name
--   the served filename ................................ pages.url, NOT pages.name
--   nav text ........................................... nav_label='Arena', title='The Arena'
--   site_plan_pages.slug / .url ........................ 'arena' / '/tools/arena/index.html'
--                                                        — `name` is not the URL source
--
-- Scoped by ID, never by name, so a concurrent rename cannot make this hit the wrong row.

UPDATE pages
   SET name = 'tool-arena-interface', updated_at = now()
 WHERE id = 'd2c8a925-1dca-44b6-a866-4561905a87a8' AND name = 'tool-arena';

UPDATE site_plan_pages
   SET name = 'tool-arena-interface'
 WHERE id = 'dd578268-44db-4d8d-9e13-a040e50d1868' AND name = 'tool-arena';
