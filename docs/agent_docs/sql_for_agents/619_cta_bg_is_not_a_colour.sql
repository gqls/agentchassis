-- 619 — `cta_bg` may hold a GRADIENT, so no component may use it in a <color> position
--
-- bugs_open/398. Filed by the finetuning_uk_service lane 2026-08-25 from an owner
-- report ("a couple of the pages have no hero images which has meant that the copy
-- is also unreadable. e.g. services.html").
--
-- WHAT IS WRONG
-- -------------
-- `cta_bg` is the one palette slot whose contract admits a gradient as well as a
-- colour — derive_brand_head_assets_action.go:275 documents that shape as expected,
-- and 10 of the fleet's css_themes hold one [MEASURED 2026-08-25]. Five component
-- rows substitute that token into positions that require a <color>:
--
--   * 3 hero bands: `background: linear-gradient(135deg, var(--color-cta-bg) 0%,
--     color-mix(in srgb, var(--color-cta-bg) 82%, ...) 100%)` — the token lands in
--     a colour-STOP and inside color-mix().
--   * 2 CTA buttons: `color: var(--color-cta-bg...)`.
--
-- A gradient in a <color> position makes the declaration INVALID AT COMPUTED-VALUE
-- TIME. CSS then discards it and falls back to inherit/initial — NOT, and this is
-- the trap, to the earlier valid declaration in the same rule. So the hero's own
-- safety net (`background: var(--color-cta-bg, var(--color-primary));` on the line
-- above) is defeated by the very line meant to enhance it, and the button's ink
-- falls back to the band's INHERITED white.
--
-- MEASURED ON THE LIVE ESTATE 2026-08-25, scripts/render_audit.py (VIZ-010):
--   finetuning.uk/services.html   .hero-content H1  1.11:1 (needs 3.0)
--   finetuning.uk/about.html      .hero-content H1  1.11:1
--   finetuning.uk/{services,about,your-own-model}.html  A.cta-btn  1.00:1 (needs 4.5)
--   robot-hands.com/about.html                          A.cta-btn  1.00:1
-- Control, same component family, SOLID cta_bg (#e9e2d3): noted.co.uk/contact.html
-- takes the identical non-imaged band branch and has ZERO hero failures. The
-- variable is the token's TYPE, nothing else.
--
-- WHY THE ESTATE DID NOT ALREADY FIX IT: it detected it. 11 contrast_failure rows
-- were filed for finetuning.uk on 2026-08-10 — including the 1.11:1 H1s — with the
-- invented `H1.H1` selector (bugs_closed/352), sat `deferred` for two weeks, and
-- were cancelled wholesale by migration 587 on 2026-08-24 without repair.
--
-- WHAT THIS DOES
-- --------------
-- 1. Heroes: DELETE the second `background:` declaration. The first one is already
--    there, is valid for a gradient AND for a colour, and is what these three
--    components fall back to. The deleted line was cosmetic (an 82% color-mix
--    toward the ink); nothing else depends on it. Note the sibling heroes
--    `hero`, `use-cases-hero` and `case-studies-hero` carry the SAME idiom against
--    `--color-primary`, which is always a colour — they are correct and are
--    deliberately untouched. That is the in-repo control for this diagnosis.
-- 2. Buttons: point the ink at `--color-cta-bg-ink`, the renderer-owned legible-ink
--    companion added alongside this in palette_specialised_slots.go. Written in the
--    mandatory two-level form `var(--ink, var(--raw))` — the measured contract of
--    that mechanism (inkPolicy.enabled's comment: 46 consumer rows, 0 bare) — so
--    this half is a NO-OP until the chassis roll emits the token, and degrades to
--    exactly today's behaviour if the ink kill-switch is ever thrown.
--
-- ⚠ NOT DONE HERE, DELIBERATELY: the first gradient STOP is not a usable ink. On 6
-- of the 10 gradient themes it scores under AA against the button's white face
-- (#3b82f6 -> 3.68, #059669 -> 3.82, #8b5cf6 -> 4.28) [MEASURED 2026-08-25]. The
-- stop is the SOURCE for the derivation; legibleInkFor re-tints it. Anyone
-- "simplifying" this migration to write the stop directly reintroduces the defect
-- on half the fleet and it will look fine on finetuning, whose #1e40af scores 8.8.
--
-- AFTER APPLYING: the served pages still hold the OLD <style> block. Fire a
-- page_rerender for the affected pages — it re-assembles from content_data plus the
-- template, so the CSS updates and the COPY IS NOT REGENERATED (which is what makes
-- this compatible with the finetuning lane's standing copy hold).
--
-- Rollback: 619_cta_bg_is_not_a_colour_ROLLBACK.sql

BEGIN;

-- 1. The three hero bands ---------------------------------------------------
UPDATE content_components
   SET html_template = replace(
           html_template,
           E'\n    background: linear-gradient(135deg, var(--color-cta-bg, var(--color-primary)) 0%, color-mix(in srgb, var(--color-cta-bg, var(--color-primary)) 82%, var(--color-cta-text, var(--color-primary-text))) 100%);',
           ''),
       updated_at = now()
 WHERE name IN ('about-hero', 'contact-hero', 'services-hero')
   AND html_template LIKE '%background: linear-gradient(135deg, var(--color-cta-bg%';

-- 2a. call-to-action --------------------------------------------------------
UPDATE content_components
   SET html_template = replace(
           html_template,
           'color: var(--color-cta-bg, var(--color-primary));',
           'color: var(--color-cta-bg-ink, var(--color-cta-bg));'),
       updated_at = now()
 WHERE name = 'call-to-action'
   AND html_template LIKE '%color: var(--color-cta-bg, var(--color-primary));%';

-- 2b. tool-cta. Anchored on the preceding background declaration so the needle
--     cannot match anything but the inverted button's ink.
UPDATE content_components
   SET html_template = replace(
           html_template,
           E'background: var(--color-white, #fff);\n    color: var(--color-cta-bg);',
           E'background: var(--color-white, #fff);\n    color: var(--color-cta-bg-ink, var(--color-cta-bg));'),
       updated_at = now()
 WHERE name = 'tool-cta'
   AND html_template LIKE E'%background: var(--color-white, #fff);\n    color: var(--color-cta-bg);%';

-- 3. VERIFY. A DO/RAISE, not a block of SELECTs: ON_ERROR_STOP does not treat a
--    non-empty result as an error, so a SELECT-only verify block cannot stop the
--    COMMIT. INDUCED 2026-08-25 against pre-migration state: the stop arm RAISEd
--    ('4 component(s) ...' — which is how the 4th, unshipped row below was found
--    at all, my by-name census having encoded its own answer). Arms 2 and 4 read
--    2 and 0 pre-migration, i.e. both would raise. Arm 3 reads 3 both before and
--    after: it is a post-condition against OVER-deletion, and cannot fire early.
DO $$
DECLARE
    n_band2   int;
    n_mix     int;
    n_ctacol  int;
    n_heroes  int;
    n_inked   int;
BEGIN
    -- No component may still put cta_bg in a colour-STOP position.
    SELECT count(*) INTO n_band2
      FROM content_components
     WHERE html_template LIKE '%linear-gradient(135deg, var(--color-cta-bg%';
    IF n_band2 <> 0 THEN
        RAISE EXCEPTION '619: % component(s) still substitute --color-cta-bg into a gradient stop', n_band2;
    END IF;

    -- The color-mix() family is PINNED, not cleared, and the difference is
    -- deliberate. One row uses `background: color-mix(in srgb, var(--color-cta-bg)
    -- N%, transparent)` — tool-password-entropy_pre_037, three occurrences. It is
    -- the same invalid-at-computed-value-time fault, but (a) its harm is a MISSING
    -- PANEL TINT rather than unreadable text, and (b) it has ZERO page_components
    -- uses [MEASURED 2026-08-25], so nothing serves it. Repairing it properly needs
    -- a solid FILL companion, which is a different token from the legible INK this
    -- change adds and which would have no live consumer — the accumulating-opt-in-
    -- surface shape RFC_022 exists to discourage. So it is recorded in bugs_open/398
    -- as a known remainder and pinned here: if the count moves, a NEW component has
    -- taken the same wrong turn and this migration refuses rather than shrugging.
    SELECT count(*) INTO n_mix
      FROM content_components
     WHERE html_template LIKE '%color-mix(in srgb, var(--color-cta-bg%';
    IF n_mix <> 1 THEN
        RAISE EXCEPTION '619: % component(s) use --color-cta-bg inside color-mix(), want exactly the 1 known unshipped row (tool-password-entropy_pre_037)', n_mix;
    END IF;

    SELECT count(*) INTO n_ctacol
      FROM content_components
     WHERE html_template ~ 'color:\s*var\(--color-cta-bg[,)]';
    IF n_ctacol <> 0 THEN
        RAISE EXCEPTION '619: % component(s) still use --color-cta-bg as a colour: value', n_ctacol;
    END IF;

    -- The heroes must KEEP their valid single declaration; deleting both would
    -- leave an unpainted band, which is the failure this repair exists to end.
    SELECT count(*) INTO n_heroes
      FROM content_components
     WHERE name IN ('about-hero', 'contact-hero', 'services-hero')
       AND html_template LIKE '%background: var(--color-cta-bg, var(--color-primary));%';
    IF n_heroes <> 3 THEN
        RAISE EXCEPTION '619: % of 3 heroes retain the valid background declaration, want 3', n_heroes;
    END IF;

    SELECT count(*) INTO n_inked
      FROM content_components
     WHERE html_template LIKE '%var(--color-cta-bg-ink, var(--color-cta-bg))%';
    IF n_inked <> 2 THEN
        RAISE EXCEPTION '619: % of 2 CTA buttons point at --color-cta-bg-ink, want 2', n_inked;
    END IF;

    RAISE NOTICE '619 OK: 3 hero bands cleaned, 2 CTA button inks repointed, 0 <color>-position uses of --color-cta-bg remain';
END $$;

COMMIT;
