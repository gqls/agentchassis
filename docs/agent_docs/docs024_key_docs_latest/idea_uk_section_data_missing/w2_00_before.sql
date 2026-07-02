-- W2a step 0 (read-only): schema + needle presence + site-head placement + W3 value prep.
\d content_components

-- 0.1 Footer needle check: all six needles the fix replaces must be present exactly once-ish;
--     white_rgba_count tells us whether the five declarations are the ONLY white rgba() uses.
SELECT function, is_active,
       (html_template LIKE '%--section-text: rgba(255,255,255,0.9);%')       AS n1,
       (html_template LIKE '%--section-text-muted: rgba(255,255,255,0.7);%') AS n2,
       (html_template LIKE '%--section-heading: #ffffff;%')                  AS n3,
       (html_template LIKE '%--section-surface: rgba(255,255,255,0.05);%')   AS n4,
       (html_template LIKE '%--section-border: rgba(255,255,255,0.2);%')     AS n5,
       (html_template LIKE '%background: var(--color-footer-bg, #1a1a2e);%') AS n6,
       (length(html_template) - length(replace(html_template, 'rgba(255,255,255', '')))
         / length('rgba(255,255,255') AS white_rgba_count   -- expect 4 (n1,n2,n4,n5)
FROM content_components
WHERE function = 'site-footer' AND is_active = true AND forked_from IS NULL;

-- 0.2 Is site-head placed anywhere as a page section? (It is unreachable as chrome —
--     RenderHead looks up function 'head'; site-head is component_level='section'.)
SELECT 'pages.sections' AS source, count(*) AS refs
FROM pages p, jsonb_array_elements_text(p.sections) AS sec(name)
WHERE sec.name = 'site-head'
UNION ALL
SELECT 'page_components', count(*)
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE cc.function = 'site-head';

-- 0.3 W3 prep: the primary pair values on the two portal layouts (hero gradient text
--     will consume --color-primary-text per the reuse rule).
SELECT name, scheme,
       substring(css_template from '--color-primary:[^;]+')      AS primary_val,
       substring(css_template from '--color-primary-text:[^;]+') AS primary_text_val
FROM layouts
WHERE name IN ('tool-portal-light','tool-portal-dark');

-- 0.4 Blast-radius context: how many active sites sit on each layout (footer change is
--     library-wide but inert until each site's footer re-renders).
SELECT l.name AS layout, l.scheme, count(s.id) AS active_sites
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes ct ON ct.id = sc.css_theme_id
JOIN layouts l ON l.id = ct.layout_id
WHERE s.status = 'active'
GROUP BY l.name, l.scheme
ORDER BY active_sites DESC;
