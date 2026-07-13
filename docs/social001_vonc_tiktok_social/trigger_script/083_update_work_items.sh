clients_db=# -- Get the page_id and filename first
SELECT id, name, url FROM pages
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
ORDER BY name;

clients_db=# SELECT id, name, url FROM pages
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
ORDER BY name;
                  id                  |            name            |                   url
--------------------------------------+----------------------------+-----------------------------------------
 a28abcd7-186b-4a33-9b89-5d7bfd727012 | about                      | /about.html
 2d0fd96a-59ca-4941-9e32-331f0f15314d | archetypes                 | /archetypes.html
 56f049fb-3ffe-49ad-b5fa-f6a87edfcb26 | contact                    | /contact.html
 b4d24f8e-fccd-49df-9dad-aa56a0b20a68 | index                      | /index.html
 f204e18f-49a9-4dc0-8457-571a9deaeb65 | provocation                | /blog/provocation.html
 e4b3b195-919f-45ad-854e-201d3e846ea8 | provocations-index         | /provocations/index.html
 f1bc679f-5c48-46e8-9bb5-76cb8cf99ca5 | tool-archetype-taster-quiz | /tools/archetype-taster-quiz/index.html
 ecb637c1-845f-46bf-b174-9c92a43f9586 | tool-gauntlet              | /tools/gauntlet/index.html
(8 rows)


-- Then insert with the correct spec
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, pipeline
) VALUES (
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74',
    'manual', 'page_rerender', 'medium', 'Manual rerender of <page>',
    jsonb_build_object(
        'domain',    'vonc.com',
        'page_id',   '<uuid from pages table>',
        'page_name', '<name>',
        'filename',  '<url without leading slash>'
    ),
    50, 'page-rerender', 'triaged', 'manual',
    'manual-rerender-<page>-' || gen_random_uuid(),
    'build'
);