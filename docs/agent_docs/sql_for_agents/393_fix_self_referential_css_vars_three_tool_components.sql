-- 393: make the self-referential CSS custom-property fallbacks OPERATIVE on
-- three mortgagecalculator.co.uk tool components (equity-release, stamp-duty,
-- rate-forecaster).
--
-- THE DEFECT: the tool generator sometimes writes its theme bridge as
--     .tool-page { --primary-color: var(--primary-color, #0b2545); ... }
-- A custom property that references ITSELF is a dependency cycle (a self-loop),
-- and per css-variables-1 s3 a cycle makes the property invalid at
-- computed-value time — the fallback CANNOT rescue its own cycle, and the
-- subtree is poisoned even when :root defines the property. Measured in
-- headless Chromium 2026-08-11: the primary button computes
-- background transparent, white label on the pale panel = 1.05:1 contrast
-- (WCAG AA floor 4.5:1). Found by the tool-acceptance vision pass on
-- equity-release (vision_finding 2026-08-11 15:35:58Z); the same idiom is on
-- stamp-duty and rate-forecaster. tool-simple writes the literal directly and
-- measures 15.54:1 — it is the healthy sibling this fix converges on.
--
-- THE FIX: replace every `--x: var(--x, <literal>)` with `--x: <literal>` in
-- the three components' rendered_html (content_data is NULL for tool
-- components; rendered_html is the stored source). Literals, not a re-bridge,
-- because the site's own --primary-color is #b59230 gold, whose pairing with
-- white text is the 2.95:1 AA failure the staged_component_build lane has
-- already handed this lane as an open palette decision — inheriting it would
-- trade a 1.05:1 failure for a 2.95:1 one. If that palette decision later
-- fixes the token, a proper two-name bridge (e.g.
-- --primary-color: var(--color-cta-bg, #0b2545)) becomes safe.
--
-- Owner direction 2026-08-11 (in chat, mortgagecalculator adoption lane):
-- "fix the button in this lane". Precedent for the shape: migration 382
-- (contrast fixed at the stored source, redeploy through the framework,
-- verify at the artefact).
--
-- AFTER APPLYING: the three pages must be redeployed (assemble-only
-- single-page deploy, RUNBOOK s10b) — this migration edits the STORED html;
-- the bucket copy is unchanged until the deploy runs.
--
-- ROLLBACK: 393_fix_self_referential_css_vars_three_tool_components_ROLLBACK.sql
-- (restores rendered_html from migration_backups rows written below).

BEGIN;

-- 1. Backup the exact rows being changed.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '393_fix_self_referential_css_vars_three_tool_components',
       'page_components', c.id::text,
       jsonb_build_object('rendered_html', c.rendered_html),
       'pre-393 html for ' || p.name
FROM page_components c JOIN pages p ON p.id = c.page_id
WHERE p.site_id = '62b5978e-4271-4589-8e00-4baebfc0447c'
  AND p.name IN ('tool-equity-release', 'tool-stamp-duty', 'tool-rate-forecaster');

-- 2. Make the fallbacks operative. The backreference \1 in the PATTERN is the
--    self-reference assertion: a legitimate two-name bridge
--    var(--other, #x) does not match and is left alone.
UPDATE page_components c
SET rendered_html = regexp_replace(
        c.rendered_html,
        '(--[a-z-]+): var\(\1, (#[0-9a-fA-F]{3,8})\)',
        '\1: \2', 'g'),
    updated_at = now()
FROM pages p
WHERE p.id = c.page_id
  AND p.site_id = '62b5978e-4271-4589-8e00-4baebfc0447c'
  AND p.name IN ('tool-equity-release', 'tool-stamp-duty', 'tool-rate-forecaster');

-- 3. Verification that can actually stop the COMMIT (DO/RAISE, not bare
--    SELECTs — a non-empty result does not stop ON_ERROR_STOP).
DO $$
DECLARE
    remaining int;
    changed   int;
    backed_up int;
    simple_touched int;
BEGIN
    -- every self-reference in the three targets is gone
    SELECT count(*) INTO remaining
    FROM page_components c JOIN pages p ON p.id = c.page_id
    WHERE p.site_id = '62b5978e-4271-4589-8e00-4baebfc0447c'
      AND p.name IN ('tool-equity-release', 'tool-stamp-duty', 'tool-rate-forecaster')
      AND c.rendered_html ~ '(--[a-z-]+): var\(\1,';
    IF remaining > 0 THEN
        RAISE EXCEPTION '393: % component(s) still carry a self-referential var()', remaining;
    END IF;

    -- all three rows were actually rewritten this transaction
    SELECT count(*) INTO changed
    FROM page_components c JOIN pages p ON p.id = c.page_id
    WHERE p.site_id = '62b5978e-4271-4589-8e00-4baebfc0447c'
      AND p.name IN ('tool-equity-release', 'tool-stamp-duty', 'tool-rate-forecaster')
      AND c.rendered_html LIKE '%--primary-color: #%';
    IF changed <> 3 THEN
        RAISE EXCEPTION '393: expected 3 components with a literal --primary-color, found %', changed;
    END IF;

    -- the backups exist before the change is committed
    SELECT count(*) INTO backed_up
    FROM migration_backups
    WHERE migration_name = '393_fix_self_referential_css_vars_three_tool_components';
    IF backed_up < 3 THEN
        RAISE EXCEPTION '393: only % backup row(s) written', backed_up;
    END IF;

    -- no-op control: tool-simple (the healthy literal sibling) must be untouched
    SELECT count(*) INTO simple_touched
    FROM page_components c JOIN pages p ON p.id = c.page_id
    WHERE p.site_id = '62b5978e-4271-4589-8e00-4baebfc0447c'
      AND p.name = 'tool-simple'
      AND c.updated_at > now() - interval '1 minute';
    IF simple_touched > 0 THEN
        RAISE EXCEPTION '393: tool-simple was modified — the pattern is wider than the defect';
    END IF;
END $$;

COMMIT;
