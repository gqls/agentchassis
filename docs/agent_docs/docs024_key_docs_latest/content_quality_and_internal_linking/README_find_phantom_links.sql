the first confirms routing, the third shows exactly what it catches on the current build


clients_db=# -- i. Do the two distinct handler agents exist? Routing depends on it.
--    nav-link-fixer (site_component surface) is referenced by check_broken_nav_links;
--    internal-link-resolver (page_component surface) is the agent we're about to build.
SELECT type, display_name, status, is_active
FROM agent_definitions
WHERE type IN ('nav-link-fixer','internal-link-resolver')
ORDER BY type;

-- ii. Is site_components populated for this site? (the header/footer substrate;
--     pages.rendered_header was empty, so the header literals live here or nowhere)
SELECT slot_name,
       (rendered_html IS NOT NULL AND rendered_html <> '') AS has_html,
       length(coalesce(rendered_html,'')) AS len
FROM site_components
WHERE site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
ORDER BY slot_name;

-- iii. Dry-run the detection itself (same SQL the check runs), so we see the
--      findings before any work items are created.
WITH site AS (SELECT id FROM sites WHERE domain='gamesdesign.co.uk'),
    real_pages AS (
SELECT DISTINCT regexp_replace(regexp_replace(url,'index\.html$',''),'/+$','') AS norm_url
FROM pages WHERE site_id = (SELECT id FROM site)
    ),
    raw_links AS (
SELECT 'page_component'::text AS surface, p.name AS page_name,
    pc.page_id::text AS page_id, COALESCE(pc.slot_name,'') AS slot_name,
    (regexp_matches(pc.rendered_html,'href="([^"]*)"','g'))[1] AS href
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM site) AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
UNION ALL
SELECT 'site_component'::text, '', '', COALESCE(sc.slot_name,''),
    (regexp_matches(sc.rendered_html,'href="([^"]*)"','g'))[1]
FROM site_components sc
WHERE sc.site_id = (SELECT id FROM site) AND sc.rendered_html IS NOT NULL AND sc.rendered_html <> ''
    ),
    classified AS (
SELECT surface, page_name, page_id, slot_name, href,
    CASE WHEN href='' THEN 'empty'
    WHEN href LIKE '/%' AND href NOT LIKE '//%'
    AND split_part(split_part(href,'#',1),'?',1) ~ '\.html$' THEN 'page_link'
    ELSE 'skip' END AS kind,
    regexp_replace(regexp_replace(split_part(split_part(href,'#',1),'?',1),'index\.html$',''),'/+$','') AS norm_href
FROM raw_links
    )
ORDER BY c.surface, c.page_name, c.href;name, c.href, issue_typeULL)ntom_internal_link' END AS issue_type,
      type      |  display_name  |    status    | is_active
----------------+----------------+--------------+-----------
 nav-link-fixer | Nav Link Fixer | experimental | t
(1 row)

 slot_name | has_html | len
-----------+----------+------
 footer    | t        | 3623
 head      | t        | 8009
 header    | t        | 3710
(3 rows)

    surface     |       page_name        | slot_name  |      href      |      issue_type       | occurrences
----------------+------------------------+------------+----------------+-----------------------+-------------
 page_component | about-index            | hero       | /contact.html  | phantom_internal_link |           1
 page_component | about-index            | hero       | /services.html | phantom_internal_link |           1
 page_component | contact-index          | hero       | /contact.html  | phantom_internal_link |           1
 page_component | contact-index          | hero       | /services.html | phantom_internal_link |           1
 page_component | games-index            | game-list  |                | empty_internal_href   |           1
 page_component | games-index            | hero       | /contact.html  | phantom_internal_link |           1
 page_component | games-index            | hero       | /services.html | phantom_internal_link |           1
 page_component | guide-economy-basics   | hero       | /contact.html  | phantom_internal_link |           1
 page_component | guide-economy-basics   | hero       | /services.html | phantom_internal_link |           1
 page_component | guide-fairness-in-rng  | hero       | /contact.html  | phantom_internal_link |           1
 page_component | guide-fairness-in-rng  | hero       | /services.html | phantom_internal_link |           1
 page_component | guide-p2p-architecture | hero       | /contact.html  | phantom_internal_link |           1
 page_component | guide-p2p-architecture | hero       | /services.html | phantom_internal_link |           1
 page_component | guide-rng-design       | hero       | /contact.html  | phantom_internal_link |           1
 page_component | guide-rng-design       | hero       | /services.html | phantom_internal_link |           1
 page_component | guides-index           | guide-list |                | empty_internal_href   |           1
 page_component | guides-index           | hero       | /contact.html  | phantom_internal_link |           1
 page_component | guides-index           | hero       | /services.html | phantom_internal_link |           1
 page_component | guide-skinner-box      | hero       | /contact.html  | phantom_internal_link |           1
 page_component | guide-skinner-box      | hero       | /services.html | phantom_internal_link |           1
 page_component | index                  | game-list  |                | empty_internal_href   |           1
 page_component | index                  | guide-list |                | empty_internal_href   |           1
 page_component | index                  | tool-list  |                | empty_internal_href   |           1
 page_component | index                  | hero       | /contact.html  | phantom_internal_link |           1
 page_component | index                  | hero       | /services.html | phantom_internal_link |           1
 page_component | tools-index            | tool-list  |                | empty_internal_href   |           1
 page_component | tools-index            | hero       | /contact.html  | phantom_internal_link |           1
 page_component | tools-index            | hero       | /services.html | phantom_internal_link |           1
 site_component |                        | header     | /contact.html  | phantom_internal_link |           1
 site_component |                        | footer     | /privacy.html  | phantom_internal_link |           1
 site_component |                        | footer     | /terms.html    | phantom_internal_link |           1
(31 rows)
