-- SQL_2026-08-03_induce_180.sql — induce bugs_open/180's fix on the one page it damages.
--
-- WHY REASON-LESS, DELIBERATELY. The landmine "a reason-less page_rerender re-staples STORED
-- html, so a template fix never lands" is TRUE and is exactly what is wanted here: the stored
-- page_components.rendered_html still HOLDS the correct anchor — it is the DEPLOYED string
-- that lost it, because repairOutboundPageLinks (LNK-023) repairs the assembled page on the
-- way out and leaves the DB copy alone. So re-stapling the stored (correct) HTML and running
-- the FIXED outbound repair over it is the narrowest possible induction, and it exercises
-- precisely the changed code path (rerender_single_page_action.go:223).
--
-- A reason of 'section_data_resolved' would instead REGENERATE the tool's sections from
-- content_data through a template — a much larger blast radius on a live tool page, and it
-- would not discriminate: it would rewrite the very bytes whose survival is the test.
--
-- PREDICTION, written BEFORE the run (so the result can disconfirm it):
--   curl the page -> `q.link` count goes 0 -> >0, and the served JS reads
--   ' <a href="' + q.link + '" ...>See guide section</a>.</p>' again.
BEGIN;
INSERT INTO site_work_items
    (site_id, page_id, source, pipeline, item_type, severity, summary, spec,
     priority, handler_agent, status, created_by, item_key)
VALUES
    ('72b9e3a6-872f-4528-a6d6-7f205ea60f4d', '39906f75-6f14-4991-ad9c-afde972e433a',
     'bugfix_180', 'build', 'page_rerender', 'medium',
     'Re-render tool-cma-obligation-checker — induce bugs_open/180 fix (LNK-029): the outbound repair must stop deleting the JS-built anchor',
     jsonb_build_object('domain','vetcomparison.uk',
                        'page_id','39906f75-6f14-4991-ad9c-afde972e433a',
                        'page_name','tool-cma-obligation-checker'),
     40, 'page-rerender', 'triaged', 'bugfix-136-sibling-link-repair-lane',
     'page_rerender_tool-cma-obligation-checker_180_induction')
ON CONFLICT DO NOTHING;
COMMIT;
