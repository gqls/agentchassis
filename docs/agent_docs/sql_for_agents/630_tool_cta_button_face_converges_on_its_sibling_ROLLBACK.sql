-- ROLLBACK for 630. Restores tool-cta's hard-coded white button face, i.e.
-- restores the divergence that makes the button's face underivable from the
-- palette. Re-render affected pages afterwards or the served CSS will not match.
BEGIN;
UPDATE content_components
   SET html_template = replace(
           html_template,
           E'background: var(--color-cta-text, var(--color-primary-text));\n    color: var(--color-cta-bg-ink',
           E'background: var(--color-white, #fff);\n    color: var(--color-cta-bg-ink'),
       updated_at = now()
 WHERE name = 'tool-cta';
COMMIT;
