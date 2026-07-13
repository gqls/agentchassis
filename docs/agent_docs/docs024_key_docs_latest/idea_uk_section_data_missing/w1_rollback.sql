-- W1 rollback (only if needed): removes exactly the two inserted lines, value-agnostic
-- (matches the leading newline + two-space indent + labels w1_01 wrote, whatever hex was used).
-- Belt-and-braces: the full pre-change css_template backup file from the shell step also exists.
UPDATE layouts
SET css_template = regexp_replace(
      css_template,
      E'\n  --color-cta-bg:[^;]+;\n  --color-cta-text:[^;]+;',
      ''
    )
WHERE name = 'tool-portal-light'
RETURNING name, (css_template LIKE '%--color-cta-bg%') AS still_has_cta_bg;  -- expect f
