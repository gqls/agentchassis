-- W3b rollback: inverse of the twelve replaces (belt-and-braces: the .bak file).
UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(html_template,
  'background-position: center; --hero-ink: #fff;{{else}}',
  'background-position: center;{{else}}'),
  '--hero-ink: var(--color-primary-text); background: var(--color-primary); background: linear-gradient(135deg, var(--color-primary) 0%, color-mix(in srgb, var(--color-primary) 85%, var(--color-primary-text)) 100%);{{end}}',
  'background: linear-gradient(135deg, var(--color-primary, #1a1a2e) 0%, var(--color-secondary, #16213e) 50%, var(--color-accent, #0f3460) 100%);{{end}}'),
  '--section-text: color-mix(in srgb, var(--hero-ink) 95%, transparent);',
  '--section-text: rgba(255,255,255,0.95);'),
  '--section-text-muted: color-mix(in srgb, var(--hero-ink) 80%, transparent);',
  '--section-text-muted: rgba(255,255,255,0.8);'),
  '--section-heading: var(--hero-ink);',
  '--section-heading: #ffffff;'),
  '--section-surface: color-mix(in srgb, var(--hero-ink) 10%, transparent);',
  '--section-surface: rgba(255,255,255,0.1);'),
  '--section-border: color-mix(in srgb, var(--hero-ink) 30%, transparent);',
  '--section-border: rgba(255,255,255,0.3);'),
  E'    margin: 0 auto;\n    color: var(--hero-ink);',
  E'    margin: 0 auto;\n    color: #fff;'),
  E'    background: var(--hero-ink);\n    color: var(--color-primary);\n    border: 2px solid var(--hero-ink);',
  E'    background: var(--color-accent, #0f3460);\n    color: #fff;\n    border: 2px solid var(--color-accent, #0f3460);'),
  E'    background: transparent;\n    color: var(--hero-ink);\n}',
  E'    background: transparent;\n    color: #fff;\n}'),
  E'    background: transparent;\n    color: var(--hero-ink);\n    border: 2px solid color-mix(in srgb, var(--hero-ink) 80%, transparent);',
  E'    background: transparent;\n    color: #fff;\n    border: 2px solid rgba(255,255,255,0.8);'),
  'background: color-mix(in srgb, var(--hero-ink) 10%, transparent);',
  'background: rgba(255,255,255,0.1);'),
 updated_at = now()
WHERE function = 'hero' AND is_active = true
  AND html_template LIKE '%--hero-ink%'
RETURNING function, (html_template LIKE '%--section-heading: #ffffff;%') AS restored;  -- expect t
