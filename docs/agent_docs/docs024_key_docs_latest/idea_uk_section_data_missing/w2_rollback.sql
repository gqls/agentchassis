-- W2a rollback (only if needed): inverse of the six replaces, guarded on post-state.
-- Belt-and-braces: the .bak file from the shell step holds the full pre-change template.
UPDATE content_components
SET html_template =
  replace(replace(replace(replace(replace(replace(html_template,
    '--section-text: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 90%, transparent);',
    '--section-text: rgba(255,255,255,0.9);'),
    '--section-text-muted: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 70%, transparent);',
    '--section-text-muted: rgba(255,255,255,0.7);'),
    '--section-heading: var(--color-footer-text, var(--color-heading));',
    '--section-heading: #ffffff;'),
    '--section-surface: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 5%, transparent);',
    '--section-surface: rgba(255,255,255,0.05);'),
    '--section-border: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 20%, transparent);',
    '--section-border: rgba(255,255,255,0.2);'),
    'background: var(--color-footer-bg, var(--color-surface));',
    'background: var(--color-footer-bg, #1a1a2e);'),
    updated_at = now()
WHERE function = 'site-footer'
  AND is_active = true
  AND html_template LIKE '%color-mix%'
RETURNING function, (html_template LIKE '%--section-heading: #ffffff;%') AS restored;  -- expect t
