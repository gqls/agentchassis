-- W3e rollback: inverse replaces for both components (belt-and-braces: the .bak files).
UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(html_template,
  'background: var(--color-background);', 'background: #fff;'),
  'color: var(--color-heading, var(--color-text));', 'color: #1a1a2e;'),
  'color: var(--color-text);', 'color: #333;'),
  'background: var(--color-surface);', 'background: #f8f9fa;'),
  'border-left: 4px solid var(--color-accent);', 'border-left: 4px solid #0f3460;'),
  'color: var(--color-text-muted);', 'color: #555;'),
 updated_at = now()
WHERE function = 'about-content' AND is_active = true
  AND html_template LIKE '%var(--color-background);%'
RETURNING function;

UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(replace(html_template,
  '--section-text: var(--color-text);', '--section-text: rgba(255,255,255,0.9);'),
  '--section-text-muted: var(--color-text-muted);', '--section-text-muted: rgba(255,255,255,0.7);'),
  '--section-heading: var(--color-heading, var(--color-text));', '--section-heading: #ffffff;'),
  '--section-surface: var(--color-surface);', '--section-surface: rgba(255,255,255,0.05);'),
  '--section-border: var(--color-border);', '--section-border: rgba(255,255,255,0.2);'),
  'background: radial-gradient(ellipse at 60% 40%, color-mix(in srgb, var(--color-primary) 12%, transparent) 0%, transparent 70%);',
  'background: radial-gradient(ellipse at 60% 40%, rgba(var(--color-primary, #7c3aed), 0.12) 0%, transparent 70%);'),
  'box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 25%, transparent);',
  'box-shadow: 0 0 0 3px rgba(124,58,237,0.25);'),
 updated_at = now()
WHERE function = 'brief-explanation' AND is_active = true
  AND html_template LIKE '%--section-text: var(--color-text);%'
RETURNING function;
