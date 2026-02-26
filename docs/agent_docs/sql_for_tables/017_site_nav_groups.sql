clients_db=# SELECT g.group_key, g.group_label, i.label, i.url, i.position
             FROM site_nav_groups g
                      JOIN site_nav_items i ON i.group_id = g.id
             WHERE g.site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
             ORDER BY g.group_key, i.position;
group_key |   group_label   |    label     |        url         | position
-----------+-----------------+--------------+--------------------+----------
 legal     | Legal           | Privacy      | /privacy.html      |        0
 legal     | Legal           | Terms        | /terms.html        |        1
 primary   | Main Navigation | Home         | /index.html        |        0
 primary   | Main Navigation | About        | /about.html        |        1
 primary   | Main Navigation | Services     | /services.html     |        2
 primary   | Main Navigation | Use Cases    | /use-cases.html    |        3
 primary   | Main Navigation | Case Studies | /case-studies.html |        4
 primary   | Main Navigation | Contact      | /contact.html      |        5
(8 rows)


--

-- 061_fix_nav_templates_global.sql
--
-- Fix all header/footer/nav content_component templates that use anchor-style
-- links (href="#{{.slug}}") instead of proper page URLs (href="{{.url}}").
--
-- This is a one-time fix for existing templates. The render context already
-- provides both .slug and .url for each nav item — the bug is purely in
-- the template data.

-- Step 1: Fix all header/footer/nav templates
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(html_template, 'href="#{{.slug}}"', 'href="{{.url}}"'),
        'href="#{{.name}}"', 'href="{{.url}}"'
                    ),
    updated_at = now()
WHERE (function IN ('header', 'footer')
    OR function LIKE 'header-%'
    OR function LIKE 'footer-%'
    OR function LIKE 'nav-%')
  AND (html_template LIKE '%href="#{{.slug}}"%'
    OR  html_template LIKE '%href="#{{.name}}"%');

-- Step 2: Verify no anchor-style links remain
SELECT id, function, name
FROM content_components
WHERE (function LIKE 'header%' OR function LIKE 'footer%' OR function LIKE 'nav%')
  AND (html_template LIKE '%href="#{{.%');

-- Should return 0 rows. If any remain, they use a pattern we didn't catch.