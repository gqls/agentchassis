-- W2a step 2 CORRECTED (read-only). Previous version failed: Postgres ARE quantifier
-- bounds max out at 255, so .{0,420} is an invalid repetition count. substr+position
-- has no such limit.
SELECT function, is_active, updated_at,
       (html_template LIKE '%rgba(255,255,255%') AS still_has_white_rgba,  -- expect f
       (html_template LIKE '%#ffffff%')          AS still_has_ffffff,      -- expect f
       substr(html_template, position('--section-text:' in html_template), 460) AS declaration_region
FROM content_components
WHERE function = 'site-footer' AND is_active = true AND forked_from IS NULL;
