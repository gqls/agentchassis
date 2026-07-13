-- W3a step 1: convert the active call-to-action to the CTA pair (contract rule 3).
-- Gate already passed (w3_00: ten needles t, color_white_count=4, white_rgba_count=4).
-- Ten exact-string replaces; n2's needle appears twice (section root + .cta-btn-secondary)
-- and BOTH occurrences want the same replacement. Buttons become the inverse pair
-- (bg = cta-text, label = cta-bg) so a dark band keeps a light button and a light band
-- gets a dark one. No literal fallbacks — every layout defines both pair vars (Check 4c + W1).
UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(html_template,
  'background: var(--color-primary, #1a1a2e);',
  'background: var(--color-cta-bg, var(--color-primary));'),
  'color: var(--color-white, #fff);',
  'color: var(--color-cta-text, var(--color-primary-text));'),
  '--section-text: rgba(255,255,255,0.9);',
  '--section-text: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 90%, transparent);'),
  '--section-text-muted: rgba(255,255,255,0.85);',
  '--section-text-muted: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 85%, transparent);'),
  '--section-heading: #ffffff;',
  '--section-heading: var(--color-cta-text, var(--color-primary-text));'),
  '--section-surface: rgba(255,255,255,0.05);',
  '--section-surface: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 5%, transparent);'),
  '--section-border: rgba(255,255,255,0.2);',
  '--section-border: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 20%, transparent);'),
  'background: var(--color-white, #fff);',
  'background: var(--color-cta-text, var(--color-primary-text));'),
  'color: var(--color-primary, #1a1a2e);',
  'color: var(--color-cta-bg, var(--color-primary));'),
  'border: 2px solid var(--color-white, #fff);',
  'border: 2px solid var(--color-cta-text, var(--color-primary-text));'),
 updated_at = now()
WHERE function = 'call-to-action'
  AND is_active = true
  AND forked_from IS NULL
  AND html_template LIKE '%--section-heading: #ffffff;%'
RETURNING function,
          (html_template LIKE '%--color-white%')     AS still_has_color_white,  -- expect f
          (html_template LIKE '%#fff%')              AS still_has_fff,          -- expect f
          (html_template LIKE '%rgba(255,255,255%')  AS still_has_white_rgba,   -- expect f
          (html_template LIKE '%--color-cta-bg%')    AS has_cta_bg;             -- expect t
-- Expect: UPDATE 1.
