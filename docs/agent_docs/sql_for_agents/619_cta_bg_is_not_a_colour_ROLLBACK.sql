-- ROLLBACK for 619. Restores the exact prior bytes, i.e. restores the DEFECT
-- (white-on-cream heroes and 1.00:1 CTA buttons on the 10 gradient-cta_bg themes).
-- Only reach for this if the repair itself breaks something worse; re-render the
-- affected pages afterwards or the served CSS will not match the templates.
BEGIN;

UPDATE content_components
   SET html_template = replace(
           html_template,
           E'    background: var(--color-cta-bg, var(--color-primary));\n',
           E'    background: var(--color-cta-bg, var(--color-primary));\n    background: linear-gradient(135deg, var(--color-cta-bg, var(--color-primary)) 0%, color-mix(in srgb, var(--color-cta-bg, var(--color-primary)) 82%, var(--color-cta-text, var(--color-primary-text))) 100%);\n'),
       updated_at = now()
 WHERE name IN ('about-hero', 'contact-hero', 'services-hero')
   AND html_template NOT LIKE '%linear-gradient(135deg, var(--color-cta-bg%';

UPDATE content_components
   SET html_template = replace(
           html_template,
           'color: var(--color-cta-bg-ink, var(--color-cta-bg));',
           'color: var(--color-cta-bg, var(--color-primary));'),
       updated_at = now()
 WHERE name = 'call-to-action';

UPDATE content_components
   SET html_template = replace(
           html_template,
           'color: var(--color-cta-bg-ink, var(--color-cta-bg));',
           'color: var(--color-cta-bg);'),
       updated_at = now()
 WHERE name = 'tool-cta';

COMMIT;
