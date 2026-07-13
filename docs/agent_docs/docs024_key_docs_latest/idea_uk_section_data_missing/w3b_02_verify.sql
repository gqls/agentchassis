-- W3b step 2 (read-only): the three converted regions.
SELECT function, updated_at,
       substr(html_template, position('style="' in html_template), 560)             AS inline_branches,
       substr(html_template, position('.hero {' in html_template), 560)             AS hero_block,
       substr(html_template, position('.hero .btn-primary' in html_template), 340)  AS buttons
FROM content_components
WHERE function = 'hero' AND is_active = true AND forked_from IS NULL;
