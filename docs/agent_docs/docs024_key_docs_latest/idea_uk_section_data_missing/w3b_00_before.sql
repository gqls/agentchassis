-- W3b step 0 (read-only): hero needle gate. position()>0 is used instead of LIKE because
-- several needles contain literal '%' (LIKE wildcards). Counts confirm full coverage.
SELECT function, is_active,
 position('background-position: center;{{else}}' in html_template) > 0 AS g1_image_branch,
 position('background: linear-gradient(135deg, var(--color-primary, #1a1a2e) 0%, var(--color-secondary, #16213e) 50%, var(--color-accent, #0f3460) 100%);{{end}}' in html_template) > 0 AS g2_gradient,
 position('--section-text: rgba(255,255,255,0.95);' in html_template) > 0  AS g3,
 position('--section-text-muted: rgba(255,255,255,0.8);' in html_template) > 0 AS g4,
 position('--section-heading: #ffffff;' in html_template) > 0              AS g5,
 position('--section-surface: rgba(255,255,255,0.1);' in html_template) > 0 AS g6,
 position('--section-border: rgba(255,255,255,0.3);' in html_template) > 0  AS g7,
 position(E'    margin: 0 auto;\n    color: #fff;' in html_template) > 0    AS g8_content,
 position(E'    background: var(--color-accent, #0f3460);\n    color: #fff;\n    border: 2px solid var(--color-accent, #0f3460);' in html_template) > 0 AS g9_btn_primary,
 position(E'    background: transparent;\n    color: #fff;\n}' in html_template) > 0 AS g10_hover,
 position(E'    background: transparent;\n    color: #fff;\n    border: 2px solid rgba(255,255,255,0.8);' in html_template) > 0 AS g11_btn_secondary,
 position('background: rgba(255,255,255,0.1);' in html_template) > 0       AS g12_hover_bg,
 (length(html_template) - length(replace(html_template,'color: #fff;','')))/length('color: #fff;')            AS color_fff_count,   -- expect 4
 (length(html_template) - length(replace(html_template,'rgba(255,255,255','')))/length('rgba(255,255,255')    AS white_rgba_count,  -- expect 7
 (length(html_template) - length(replace(html_template,'#0f3460','')))/length('#0f3460')                       AS accent_hex_count   -- expect 2
FROM content_components
WHERE function = 'hero' AND is_active = true AND forked_from IS NULL;
-- GATE: all g1..g12 true AND counts 4 / 7 / 2. Anything else = template drifted; re-dump first.
