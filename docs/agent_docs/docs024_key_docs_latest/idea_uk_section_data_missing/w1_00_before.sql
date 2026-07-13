-- W1 step 0 (read-only): schema check + before-state + library sweep.
-- Schema first, per convention (also shows whether layouts has an updated_at column;
-- w1_01 deliberately does not touch it — add "updated_at = now()" there if present and wanted).
\d layouts

-- Before-state: confirm the anchor line exists and the CTA pair is absent on tool-portal-light.
SELECT name, scheme,
       substring(css_template from '--color-footer-text:[^;]+;') AS anchor_line,
       (css_template LIKE '%--color-cta-bg%')   AS has_cta_bg,    -- expect f
       (css_template LIKE '%--color-cta-text%') AS has_cta_text   -- expect f
FROM layouts
WHERE name = 'tool-portal-light' AND is_active = true;

-- Sweep (the second half of W1): every layout's cta pair values + hero pair presence.
-- Paste this back; the contrast arithmetic gets done off-cluster.
SELECT name, scheme,
       substring(css_template from '--color-cta-bg:[^;]+')    AS cta_bg,
       substring(css_template from '--color-cta-text:[^;]+')  AS cta_text,
       (css_template LIKE '%--color-hero-title%')    AS has_hero_title,
       (css_template LIKE '%--color-hero-subtitle%') AS has_hero_subtitle
FROM layouts
WHERE is_active = true
ORDER BY scheme NULLS LAST, name;
