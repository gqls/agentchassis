-- 630 — tool-cta's inverted button paints a HARD-CODED WHITE; converge it on the
--       palette slot its sibling already uses, so the CTA button has ONE face
--
-- bugs_open/398, round 2. Raised by the council's editquality seat against 619's
-- submission (correlation f0591cb2): the seat objected that --color-cta-bg-ink
-- was grounded on a surface the ink does not sit on. Investigating that turned up
-- a real divergence the round-1 plan had not measured.
--
-- THE DIVERGENCE [MEASURED 2026-08-25, content_components.html_template]
--   call-to-action  .cta-btn-primary       background: var(--color-cta-text, var(--color-primary-text))
--   tool-cta        .tool-cta-btn-primary  background: var(--color-white, #fff)
--
-- Both are the SAME UI object: the inverted button that sits inside the CTA band.
-- One reads the palette; the other hard-codes white. That is the bugs_open/113
-- layout-literal shape, and it is not cosmetic — it makes the button's face
-- unpredictable from the palette, so nothing can derive a correct ink for it.
--
-- WHY THIS IS THE FIX RATHER THAN GROUNDING THE INK ON BOTH FACES. Grounding on
-- both was the obvious answer and it is WRONG, not merely expensive: **7 live
-- themes** pair a LIGHT band with a NEAR-BLACK ink — `cta_bg #e9e2d3` /
-- `cta_text #1a1a1a` (theme-noted-co-uk, -idea-uk, -lendzy-co-uk, -loanzy-uk,
-- -mortgagecalculator-co-uk, -remortgagecalculator-uk, -webdesign-co-uk)
-- [MEASURED 2026-08-25]. There the two faces are #ffffff and #1a1a1a, and NO
-- single colour clears 4.5 against both; legibleInkFor would fall through to its
-- terminal branch and emit black at a worst ratio of ~1.2 — a second invisible
-- button, in the opposite direction from the one 619 repairs. Two faces needed
-- either two tokens or one converged face. This converges the face.
--
-- EFFECT TODAY, before the chassis roll that emits --color-cta-bg-ink:
--   * cta_text = #ffffff (most themes, all 10 gradient ones): NO VISIBLE CHANGE —
--     var(--color-cta-text) resolves to the same white the literal hard-coded.
--   * the 7 light-band themes: the button face becomes #1a1a1a and the ink falls
--     back through to `var(--color-cta-bg)` = #e9e2d3, i.e. cream on near-black.
--     That is legible (≈12:1) and is an improvement on today's cream-on-white.
--
-- NOT FANNED OUT HERE. 631 fans out the hero repair only; every page carrying a
-- CTA button waits for one fan-out after the chassis roll, so it is re-rendered
-- once with both halves live rather than twice.
--
-- Rollback: 630_..._ROLLBACK.sql

BEGIN;

UPDATE content_components
   SET html_template = replace(
           html_template,
           E'background: var(--color-white, #fff);\n    color: var(--color-cta-bg-ink',
           E'background: var(--color-cta-text, var(--color-primary-text));\n    color: var(--color-cta-bg-ink'),
       updated_at = now()
 WHERE name = 'tool-cta'
   AND html_template LIKE E'%background: var(--color-white, #fff);\n    color: var(--color-cta-bg-ink%';

DO $$
DECLARE n_converged int; n_literal int; n_sibling int;
BEGIN
    SELECT count(*) INTO n_converged FROM content_components
     WHERE name = 'tool-cta'
       AND html_template LIKE E'%background: var(--color-cta-text, var(--color-primary-text));\n    color: var(--color-cta-bg-ink%';
    IF n_converged <> 1 THEN
        RAISE EXCEPTION '630: tool-cta button face converged on % row(s), want 1 — has 619 been applied? this migration anchors on the ink line it wrote', n_converged;
    END IF;

    -- The literal must be gone from the BUTTON. It may legitimately survive
    -- elsewhere in the template, so this is scoped to the anchored pair.
    SELECT count(*) INTO n_literal FROM content_components
     WHERE name = 'tool-cta'
       AND html_template LIKE E'%background: var(--color-white, #fff);\n    color: var(--color-cta-bg-ink%';
    IF n_literal <> 0 THEN
        RAISE EXCEPTION '630: % tool-cta row(s) still hard-code the button face', n_literal;
    END IF;

    -- And the two siblings must now agree, which is the whole point.
    SELECT count(*) INTO n_sibling FROM content_components
     WHERE name IN ('call-to-action', 'tool-cta')
       AND html_template LIKE '%background: var(--color-cta-text, var(--color-primary-text));%'
       AND html_template LIKE '%var(--color-cta-bg-ink, var(--color-cta-bg))%';
    IF n_sibling <> 2 THEN
        RAISE EXCEPTION '630: % of 2 CTA buttons paint the palette face AND read the ink companion, want 2', n_sibling;
    END IF;

    RAISE NOTICE '630 OK: both CTA buttons now paint var(--color-cta-text, var(--color-primary-text)) and read --color-cta-bg-ink';
END $$;

COMMIT;
