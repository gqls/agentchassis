-- W1 step 1: add the CTA pair to tool-portal-light.
-- Anchor = the existing --color-footer-text line (verified present, Check 2b); \1 preserves
-- it byte-for-byte, so no whitespace guessing. The NOT LIKE guard makes this idempotent
-- (a re-run is a 0-row no-op). regexp_replace without flags edits the first occurrence only.
--
-- Values: neutral light band + dark ink, mirroring tool-portal-dark's neutral #1e1e1e band
-- (its CTA is an elevated neutral, not an accent). Accent-band alternative in keeping with
-- soft-editorial: swap "#e9e2d3" -> "#9b4020" and "#1a1a1a" -> "#faf8f3".
-- Either way a site's palette can override via its cta_bg / cta_text specialised slots.
UPDATE layouts
SET css_template = regexp_replace(
      css_template,
      '(--color-footer-text:[^;]+;)',
      E'\\1\n  --color-cta-bg:         {{palette "cta_bg"         "#e9e2d3"}};\n  --color-cta-text:       {{palette "cta_text"       "#1a1a1a"}};'
    )
WHERE name = 'tool-portal-light'
  AND is_active = true
  AND css_template NOT LIKE '%--color-cta-bg%'
RETURNING name,
          substring(css_template from '--color-cta-bg:[^;]+')   AS inserted_cta_bg,
          substring(css_template from '--color-cta-text:[^;]+') AS inserted_cta_text;
-- Expect: UPDATE 1, with both inserted_* columns populated.
