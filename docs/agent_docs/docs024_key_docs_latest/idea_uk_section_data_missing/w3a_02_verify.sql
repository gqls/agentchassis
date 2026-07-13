-- W3a step 2 (read-only): the converted regions.
SELECT function, updated_at,
       substr(html_template, position('.cta-section {' in html_template), 620)   AS section_region,
       substr(html_template, position('.cta-btn-primary' in html_template), 240) AS button_region
FROM content_components
WHERE function = 'call-to-action' AND is_active = true AND forked_from IS NULL;
