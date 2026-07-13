-- W2a step 1: de-hardcode the active generated site-footer.
-- Five literal dark declarations -> references to the footer pair (alphas preserved:
-- 90/70/heading-full/5/20 via color-mix); background fallback #1a1a2e -> var(--color-surface).
-- Exact-string nested replace() (needles byte-for-byte from the Check 2a dump) — no regex.
-- Guarded on a pre-state marker => idempotent. updated_at bumped (column confirmed in \d).
UPDATE content_components
SET html_template =
  replace(replace(replace(replace(replace(replace(html_template,
    '--section-text: rgba(255,255,255,0.9);',
    '--section-text: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 90%, transparent);'),
    '--section-text-muted: rgba(255,255,255,0.7);',
    '--section-text-muted: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 70%, transparent);'),
    '--section-heading: #ffffff;',
    '--section-heading: var(--color-footer-text, var(--color-heading));'),
    '--section-surface: rgba(255,255,255,0.05);',
    '--section-surface: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 5%, transparent);'),
    '--section-border: rgba(255,255,255,0.2);',
    '--section-border: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 20%, transparent);'),
    'background: var(--color-footer-bg, #1a1a2e);',
    'background: var(--color-footer-bg, var(--color-surface));'),
    updated_at = now()
WHERE function = 'site-footer'
  AND is_active = true
  AND forked_from IS NULL
  AND html_template LIKE '%--section-heading: #ffffff;%'
RETURNING function,
          (html_template LIKE '%rgba(255,255,255%')  AS still_has_white_rgba,  -- expect f
          (html_template LIKE '%color-mix%')          AS has_color_mix,          -- expect t
          (html_template LIKE '%var(--color-footer-text%') AS refs_footer_text;  -- expect t
-- Expect: UPDATE 1.
