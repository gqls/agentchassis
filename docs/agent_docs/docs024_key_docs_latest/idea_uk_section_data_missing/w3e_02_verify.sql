-- W3e verify (read-only).
SELECT function, updated_at,
       substr(html_template, position('-section {' in html_template) - 24, 460) AS root_region
FROM content_components
WHERE function IN ('about-content','brief-explanation')
  AND is_active = true AND forked_from IS NULL
ORDER BY function;
