-- W3a rollback: inverse of the ten replaces (belt-and-braces: the .bak from the shell step).
UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(html_template,
  'background: var(--color-cta-bg, var(--color-primary));',
  'background: var(--color-primary, #1a1a2e);'),
  'color: var(--color-cta-text, var(--color-primary-text));',
  'color: var(--color-white, #fff);'),
  '--section-text: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 90%, transparent);',
  '--section-text: rgba(255,255,255,0.9);'),
  '--section-text-muted: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 85%, transparent);',
  '--section-text-muted: rgba(255,255,255,0.85);'),
  '--section-heading: var(--color-cta-text, var(--color-primary-text));',
  '--section-heading: #ffffff;'),
  '--section-surface: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 5%, transparent);',
  '--section-surface: rgba(255,255,255,0.05);'),
  '--section-border: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 20%, transparent);',
  '--section-border: rgba(255,255,255,0.2);'),
  'background: var(--color-cta-text, var(--color-primary-text));',
  'background: var(--color-white, #fff);'),
  'color: var(--color-cta-bg, var(--color-primary));',
  'color: var(--color-primary, #1a1a2e);'),
  'border: 2px solid var(--color-cta-text, var(--color-primary-text));',
  'border: 2px solid var(--color-white, #fff);'),
 updated_at = now()
WHERE function = 'call-to-action' AND is_active = true
  AND html_template LIKE '%--color-cta-bg%'
RETURNING function, (html_template LIKE '%--section-heading: #ffffff;%') AS restored;  -- expect t
