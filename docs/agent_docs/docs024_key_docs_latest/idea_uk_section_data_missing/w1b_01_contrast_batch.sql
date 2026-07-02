-- W1b: same-hue darker cta_bg fallbacks on the five layouts failing 4.5 with white text.
-- Anchored inside the --color-cta-bg line via \1 backreference (whitespace-safe, W1 pattern);
-- each statement guarded on its own old hex => idempotent. Trigger from W2b now bumps
-- updated_at automatically on layouts. Zero live impact: none of the 7 sites uses these five.
UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#f97316', E'\\1#c2410c')
WHERE name = 'brochure-bold' AND css_template LIKE '%--color-cta-bg%' AND css_template LIKE '%#f97316%'
RETURNING name, substring(css_template from '--color-cta-bg:[^;]+') AS cta_bg;

UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#ea580c', E'\\1#c2410c')
WHERE name = 'affiliate-hub' AND css_template LIKE '%--color-cta-bg%' AND css_template LIKE '%#ea580c%'
RETURNING name, substring(css_template from '--color-cta-bg:[^;]+') AS cta_bg;

UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#ef4444', E'\\1#dc2626')
WHERE name = 'media-grid' AND css_template LIKE '%--color-cta-bg%' AND css_template LIKE '%#ef4444%'
RETURNING name, substring(css_template from '--color-cta-bg:[^;]+') AS cta_bg;

UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#ff1744', E'\\1#c4001d')
WHERE name = 'high-energy' AND css_template LIKE '%--color-cta-bg%' AND css_template LIKE '%#ff1744%'
RETURNING name, substring(css_template from '--color-cta-bg:[^;]+') AS cta_bg;

UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#0284c7', E'\\1#0369a1')
WHERE name = 'docs-sidebar' AND css_template LIKE '%--color-cta-bg%' AND css_template LIKE '%#0284c7%'
RETURNING name, substring(css_template from '--color-cta-bg:[^;]+') AS cta_bg;
-- Expect: five separate "UPDATE 1" results, each RETURNING the new hex.
