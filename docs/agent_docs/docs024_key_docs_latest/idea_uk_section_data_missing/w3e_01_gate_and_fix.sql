-- W3e: about-content + brief-explanation — the last two idea.uk gate components.
-- Gate first (read-only), then the two guarded updates. Counts derived mechanically
-- from the w3e_00 dumps. position() where needles carry literal '%'.

-- ── GATE ──
SELECT function,
 (length(html_template) - length(replace(html_template,'#1a1a2e','')))/length('#1a1a2e') AS c_1a1a2e,          -- about-content: expect 2
 (html_template LIKE '%background: #fff;%')                                              AS n_bg_fff,          -- about-content: t
 (html_template LIKE '%color: #333;%')                                                   AS n_333,             -- about-content: t
 (html_template LIKE '%background: #f8f9fa;%')                                           AS n_f8f9fa,          -- about-content: t
 (html_template LIKE '%border-left: 4px solid #0f3460;%')                                AS n_borderleft,      -- about-content: t
 (html_template LIKE '%color: #555;%')                                                   AS n_555,             -- about-content: t
 (length(html_template) - length(replace(html_template,'rgba(255,255,255','')))/length('rgba(255,255,255')     AS white_rgba_count, -- brief-explanation: expect 4
 (html_template LIKE '%--section-heading: #ffffff;%')                                    AS n_heading,         -- brief-explanation: t
 position('background: radial-gradient(ellipse at 60% 40%, rgba(var(--color-primary, #7c3aed), 0.12) 0%, transparent 70%);' in html_template) > 0 AS n_radial, -- brief-explanation: t
 (html_template LIKE '%box-shadow: 0 0 0 3px rgba(124,58,237,0.25);%')                   AS n_ring             -- brief-explanation: t
FROM content_components
WHERE function IN ('about-content','brief-explanation')
  AND is_active = true AND forked_from IS NULL
ORDER BY function;

-- ── FIX 1: about-content — literals → core palette vars ──
UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(html_template,
  'background: #fff;',
  'background: var(--color-background);'),
  'color: #1a1a2e;',
  'color: var(--color-heading, var(--color-text));'),
  'color: #333;',
  'color: var(--color-text);'),
  'background: #f8f9fa;',
  'background: var(--color-surface);'),
  'border-left: 4px solid #0f3460;',
  'border-left: 4px solid var(--color-accent);'),
  'color: #555;',
  'color: var(--color-text-muted);'),
 updated_at = now()
WHERE function = 'about-content' AND is_active = true AND forked_from IS NULL
  AND html_template LIKE '%background: #fff;%'
RETURNING function,
          (html_template ~ '#[0-9a-fA-F]{3,8}')     AS still_has_any_hex,   -- expect f
          (html_template LIKE '%var(--color-%')     AS has_palette_vars;    -- expect t

-- ── FIX 2: brief-explanation — dark declarations → ambient pass-through; fix the
--            invalid rgba(var(),α) glow (renders for the first time) and the violet ring ──
UPDATE content_components
SET html_template =
 replace(replace(replace(replace(replace(replace(replace(html_template,
  '--section-text: rgba(255,255,255,0.9);',
  '--section-text: var(--color-text);'),
  '--section-text-muted: rgba(255,255,255,0.7);',
  '--section-text-muted: var(--color-text-muted);'),
  '--section-heading: #ffffff;',
  '--section-heading: var(--color-heading, var(--color-text));'),
  '--section-surface: rgba(255,255,255,0.05);',
  '--section-surface: var(--color-surface);'),
  '--section-border: rgba(255,255,255,0.2);',
  '--section-border: var(--color-border);'),
  'background: radial-gradient(ellipse at 60% 40%, rgba(var(--color-primary, #7c3aed), 0.12) 0%, transparent 70%);',
  'background: radial-gradient(ellipse at 60% 40%, color-mix(in srgb, var(--color-primary) 12%, transparent) 0%, transparent 70%);'),
  'box-shadow: 0 0 0 3px rgba(124,58,237,0.25);',
  'box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 25%, transparent);'),
 updated_at = now()
WHERE function = 'brief-explanation' AND is_active = true AND forked_from IS NULL
  AND html_template LIKE '%--section-heading: #ffffff;%'
RETURNING function,
          (html_template LIKE '%rgba(255,255,255%')      AS still_has_white_rgba,  -- expect f
          (html_template LIKE '%rgba(var(--color-primary%') AS still_has_bad_rgba, -- expect f
          (html_template LIKE '%rgba(124,58,237%')        AS still_has_violet,     -- expect f
          (html_template LIKE '%--section-text: var(--color-text);%') AS has_passthrough; -- expect t
-- Expect: gate row values as annotated, then UPDATE 1 + UPDATE 1.
