-- 253_oufe_chart_and_footer_readability.sql
-- RECORD ONLY — these statements were applied live on 2026-07-28 and are kept
-- here so the change is reviewable and repeatable, not to be re-run blindly.
--
-- The owner sent back a screenshot of the new capital-structure chart with the
-- eyebrow and the chart's own title unreadable. Three separate defects, all
-- previously WRITTEN DOWN by this workstream and none previously CHECKED FOR:
--
--  1. eyebrow 1.23 — evidence-chart styles it `var(--color-primary)`, and oufe's
--     --color-primary (#1B2A3B) is identical to its surface colour. That is the
--     headline defect of bugs_open/122, landing on a component added this week.
--     Patched on OUFE'S STORED INSTANCE, not the shared component: fundamentallyai
--     also uses evidence-chart and its palette is correct there (--color-primary
--     #86ADDE, light on dark), so changing the component would alter a working
--     site to fix a broken palette on this one.
--
--  2. chart title 1.29 / caption 2.85 — light --color-text on --color-card-bg
--     #ffffff, a WHITE card on a dark navy site. bugs_open/122 called this
--     "latent ... the same invisibility bug waiting". Fixed in the site
--     stylesheet (gqls/sites e9770b171): card-bg is now #1B2A3B, giving 11.32
--     and 5.12. Only 1 usage in the sheet, so the blast radius is nil.
--
--  3. THE FIX BROKE THE BARS, and no contrast tool caught it. The default bar
--     fill is var(--color-primary) — now exactly the card colour — so two of
--     three bars became invisible. Bars carry no text, and contrastscan skips
--     elements with no text content. That is register concept VIZ-011 (chart
--     furniture is a graphical object needing 3.0) written the same day and not
--     implemented in the tool meant to enforce it. Caught by SCREENSHOTTING the
--     result of the fix rather than re-running the check that had passed.
--
-- Also: the footer note rendered at ~560px because `max-width: 80ch` at 0.85rem
-- capped it far short of the 1200px container, so it read as belonging to the
-- first grid column. Now full width with a rule above it.
--
-- Measured against the card #1B2A3B: --color-text-muted 5.12, slate #6B7F96 3.54.

-- 1 + 3: oufe's stored evidence-chart instance (locked permanent, so no re-render
--        reverts these).
UPDATE page_components pc
SET rendered_html = replace(replace(replace(pc.rendered_html,
      'color: var(--color-primary, #1e40af);', 'color: var(--color-accent, #c49a3c);'),
      'background: var(--color-primary, #1e40af);', 'background: var(--color-text-muted, #8a9bae);'),
      'background: var(--color-secondary, rgba(127, 127, 127, 0.6));', 'background: #6B7F96;')
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39'
  AND pc.slot_name = 'evidence-chart';

-- 2 is in the site stylesheet, not the DB: gqls/sites oufe.com/assets/css/styles.css
--   --color-card-bg: #ffffff -> #1B2A3B   (commit e9770b171)

-- Footer note: full container width, banded.
UPDATE site_components
SET rendered_html = replace(replace(rendered_html,
      'margin: 0; max-width: 80ch; }', 'margin: 0; max-width: none; }'),
      '.footer-note { max-width: 1200px; margin: 0 auto; padding: 0 2rem 1.5rem; }',
      '.footer-note { max-width: 1200px; margin: 2rem auto 0; padding: 1.5rem 2rem; border-top: 1px solid rgba(255,255,255,0.15); }')
WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND slot_name = 'footer';

-- VERIFIED AFTER APPLYING: 14 measurable pairs all pass (worst 4.78), and a
-- screenshot confirms all three bars and every label are legible. The numbers
-- alone were not sufficient and are not sufficient next time either.
