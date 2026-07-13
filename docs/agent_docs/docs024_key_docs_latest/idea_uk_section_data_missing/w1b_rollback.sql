-- W1b rollback: inverse hex swaps (same anchoring + guards).
UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#c2410c', E'\\1#f97316')
WHERE name = 'brochure-bold' AND css_template LIKE '%#c2410c%' RETURNING name;
UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#c2410c', E'\\1#ea580c')
WHERE name = 'affiliate-hub' AND css_template LIKE '%#c2410c%' RETURNING name;
UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#dc2626', E'\\1#ef4444')
WHERE name = 'media-grid' AND css_template LIKE '%#dc2626%' RETURNING name;
UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#c4001d', E'\\1#ff1744')
WHERE name = 'high-energy' AND css_template LIKE '%#c4001d%' RETURNING name;
UPDATE layouts SET css_template = regexp_replace(css_template, '(--color-cta-bg:[^;]*)#0369a1', E'\\1#0284c7')
WHERE name = 'docs-sidebar' AND css_template LIKE '%#0369a1%' RETURNING name;
