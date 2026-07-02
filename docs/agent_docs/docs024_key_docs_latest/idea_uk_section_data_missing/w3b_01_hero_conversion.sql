-- W3b step 1: convert the active hero to the ink model (call = option (c)).
-- Image branch keeps overlay + sets --hero-ink:#fff (structural-dark exception).
-- No-image branch: layered solid + single-hue gradient mixing 15% TOWARD the ink
-- (depth on dark and light primaries; bounded contrast cost; solid layer = color-mix fallback).
-- Buttons: primary = inverse pair (ink bg, primary label); secondary from ink mixes.
-- Multi-line needles (E'' with \n) disambiguate the four different 'color: #fff;' sites.
UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(html_template,
  'background-position: center;{{else}}',
  'background-position: center; --hero-ink: #fff;{{else}}'),
  'background: linear-gradient(135deg, var(--color-primary, #1a1a2e) 0%, var(--color-secondary, #16213e) 50%, var(--color-accent, #0f3460) 100%);{{end}}',
  '--hero-ink: var(--color-primary-text); background: var(--color-primary); background: linear-gradient(135deg, var(--color-primary) 0%, color-mix(in srgb, var(--color-primary) 85%, var(--color-primary-text)) 100%);{{end}}'),
  '--section-text: rgba(255,255,255,0.95);',
  '--section-text: color-mix(in srgb, var(--hero-ink) 95%, transparent);'),
  '--section-text-muted: rgba(255,255,255,0.8);',
  '--section-text-muted: color-mix(in srgb, var(--hero-ink) 80%, transparent);'),
  '--section-heading: #ffffff;',
  '--section-heading: var(--hero-ink);'),
  '--section-surface: rgba(255,255,255,0.1);',
  '--section-surface: color-mix(in srgb, var(--hero-ink) 10%, transparent);'),
  '--section-border: rgba(255,255,255,0.3);',
  '--section-border: color-mix(in srgb, var(--hero-ink) 30%, transparent);'),
  E'    margin: 0 auto;\n    color: #fff;',
  E'    margin: 0 auto;\n    color: var(--hero-ink);'),
  E'    background: var(--color-accent, #0f3460);\n    color: #fff;\n    border: 2px solid var(--color-accent, #0f3460);',
  E'    background: var(--hero-ink);\n    color: var(--color-primary);\n    border: 2px solid var(--hero-ink);'),
  E'    background: transparent;\n    color: #fff;\n}',
  E'    background: transparent;\n    color: var(--hero-ink);\n}'),
  E'    background: transparent;\n    color: #fff;\n    border: 2px solid rgba(255,255,255,0.8);',
  E'    background: transparent;\n    color: var(--hero-ink);\n    border: 2px solid color-mix(in srgb, var(--hero-ink) 80%, transparent);'),
  'background: rgba(255,255,255,0.1);',
  'background: color-mix(in srgb, var(--hero-ink) 10%, transparent);'),
 updated_at = now()
WHERE function = 'hero'
  AND is_active = true
  AND forked_from IS NULL
  AND html_template LIKE '%--section-heading: #ffffff;%'
RETURNING function,
          (position('color: #fff;' in html_template) > 0)        AS still_has_color_fff,   -- expect f
          (html_template LIKE '%rgba(255,255,255%')               AS still_has_white_rgba,  -- expect f
          (html_template LIKE '%#0f3460%')                        AS still_has_accent_hex,  -- expect f
          (html_template LIKE '%--hero-ink%')                     AS has_hero_ink,          -- expect t
          (html_template LIKE '%color-mix%')                      AS has_color_mix;         -- expect t
-- Expect: UPDATE 1. (--hero-ink: #fff remains in the IMAGE branch by design.)
