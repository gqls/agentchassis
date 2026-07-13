-- W1b verify (read-only): the five cta pairs + trigger-bumped updated_at.
SELECT name, updated_at,
       substring(css_template from '--color-cta-bg:[^;]+')   AS cta_bg,
       substring(css_template from '--color-cta-text:[^;]+') AS cta_text
FROM layouts
WHERE name IN ('brochure-bold','affiliate-hub','media-grid','high-energy','docs-sidebar')
ORDER BY name;
-- Expect: #c2410c, #c2410c, #dc2626, #c4001d, #0369a1 respectively; cta_text unchanged (#ffffff).
