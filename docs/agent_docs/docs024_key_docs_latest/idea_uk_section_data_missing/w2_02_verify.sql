-- W2a step 2 (read-only): verify the declaration block now references the pair.
SELECT function, is_active, updated_at,
       (html_template LIKE '%rgba(255,255,255%') AS still_has_white_rgba,  -- expect f
       (html_template LIKE '%#ffffff%')          AS still_has_ffffff,      -- expect f
       substring(html_template from '--section-text:.{0,420}') AS declaration_region
FROM content_components
WHERE function = 'site-footer' AND is_active = true AND forked_from IS NULL;
