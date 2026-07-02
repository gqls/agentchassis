-- W1 step 2 (read-only): verify a single clean insertion and show the region.
SELECT name,
       (length(css_template) - length(replace(css_template, '--color-cta-bg', '')))
         / length('--color-cta-bg')   AS cta_bg_occurrences,    -- expect 1
       (length(css_template) - length(replace(css_template, '--color-cta-text', '')))
         / length('--color-cta-text') AS cta_text_occurrences,  -- expect 1
       substring(css_template from '--color-footer-text.{0,230}') AS insertion_region
FROM layouts
WHERE name = 'tool-portal-light';
