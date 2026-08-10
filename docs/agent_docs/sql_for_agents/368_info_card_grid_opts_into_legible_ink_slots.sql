-- ============================================================================
-- 368 — info-card-grid opts into the legible ink slots (bug 122 follow-up)
-- ============================================================================
-- Written 2026-08-10 (bugfix_122_contrast_ink_slots lane).
--
-- WHY. The 2026-08-10 re-audit (AFTER_2026-08-10_render_audit.txt) found the
-- sub-shape A defect re-materialised on dartsonline.com: the lane that removed
-- image-hover-card-grid (whose one baseline failure closed by removal) moved its
-- placements onto info-card-grid, which was never in migration 338's coverage.
-- Six new firm failures on dartsonline index:
--     .info-card-grid__eyebrow      1.14:1   (#1A1F2E on #0F1219, page background)
--     .info-card-grid__card-link    1.06:1 x5 (#1A1F2E on #1E2436, card-bg = surface)
--
-- MECHANISM — identical to 338 §4 (tool-list, .tl-eyebrow/.tl-card-link, which
-- closed robot-hands x2, verified in the same re-audit). info-card-grid hardcodes
-- nothing: it uses `color: var(--color-primary)` as a FOREGROUND on the page
-- grounds (`var(--color-background)` section, `var(--color-card-bg,
-- var(--color-surface))` cards). A site whose primary is a dark fill colour gets
-- an invisible ink. `--color-primary-ink` is the renderer-computed legible
-- companion (VIZ-014, live since v1.0.1266), already present in the served
-- stylesheets of the 11 sites the CSS half re-rendered 08-09/08-10.
--
-- MEASURED BEFORE WRITING (the 07-28 ruling: measure, do not ask the reviewer to):
--   * Placements: 27 placements across 14 sites (query in the lane's NOTES,
--     2026-08-10 — pages.sections is an array of PLAIN STRINGS, not objects).
--     Consumers: ai-agent-orchestration, cookly, dartsonline, finetuning,
--     gaswholesalers, idea.uk, lendzy, leopardess, mortgagecalculator, oufe,
--     relojistas, robot-hands, vetcomparison, webdesign.co.uk, webdesign.uk.
--   * On sites where primary is already legible on the grounds, legibleInkFor
--     returns primary UNCHANGED, so --color-primary-ink == --color-primary and
--     this change is a byte-level no-op after re-render. The 08-10 audit shows
--     oufe, relojistas, vetcomparison and webdesign.co.uk homepages (which all
--     place this component) have zero info-card-grid failures.
--   * On sites whose stylesheet lacks the ink slot (not in the 11: e.g. cookly,
--     lendzy), the var() FALLBACK keeps today's behaviour exactly.
--   * dartsonline predicted result (served CSS slots, WCAG formula):
--     eyebrow 1.14 -> 16.73:1, card-link 1.06 -> 13.78:1.
--
-- TARGETS — exactly 2 of the template's 6 `var(--color-primary)` occurrences:
--     L24   .info-card-grid__eyebrow    color:    <- target
--     L118  .info-card-grid__card-link  color:    <- target
--     L127  :hover  var(--color-primary-hover, var(--color-primary))  non-target
--     L132  :focus-visible outline                                    non-target
--     L270  background: var(--color-primary)                          non-target
--                (repointing a BACKGROUND to an ink is the inversion this bug is about)
--     L274  outline                                                   non-target
-- replace() is global within the string, so both needles below are RULE-SCOPED
-- (three declarations each, real 4-space indentation, LF line ends) and every
-- count is asserted. Template is one-declaration-per-line (unlike tool-list's
-- one-rule-per-line), hence multi-line needles.
--
-- INERT UNTIL RE-RENDER: rendered_html is baked per page; this row change ships
-- to a visitor only when a page re-renders. The lane re-renders dartsonline
-- index + brands deliberately; the other 25 placements pick it up as a no-op
-- (or a fix) whenever their own lanes next render them.
--
-- ============================================================================
-- BACKUP — TAKE THIS FIRST, OUTSIDE THE TRANSACTION
-- ============================================================================
-- kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--   -c "COPY (SELECT id,name,html_template FROM content_components WHERE is_active AND forked_from IS NULL AND name='info-card-grid') TO STDOUT" \
--   > .../scratchpad/backups/backup_368_info_card_grid.tsv
--
-- Rollback is at the foot of this file.

BEGIN;

DO $$
DECLARE
    n_eyebrow  text := E'    text-transform: uppercase;\n    color: var(--color-primary);\n    margin-bottom: 0.75rem;';
    r_eyebrow  text := E'    text-transform: uppercase;\n    color: var(--color-primary-ink, var(--color-primary));\n    margin-bottom: 0.75rem;';
    n_cardlink text := E'    font-weight: 600;\n    color: var(--color-primary);\n    text-decoration: none;';
    r_cardlink text := E'    font-weight: 600;\n    color: var(--color-primary-ink, var(--color-primary));\n    text-decoration: none;';
    tpl        text;
    c          int;
BEGIN
    -- IDEMPOTENCY. Both replacements nest their own needle, so a second run
    -- would double-wrap rather than no-op. schema_migrations should prevent
    -- that; this makes it impossible rather than merely unlikely.
    PERFORM 1 FROM content_components
      WHERE is_active AND forked_from IS NULL AND name = 'info-card-grid'
        AND position('var(--color-primary-ink,' in html_template) > 0;
    IF FOUND THEN
        RAISE EXCEPTION '368/info-card-grid: already applied (var(--color-primary-ink,… present). '
                        'Re-running would double-wrap. STOP.';
    END IF;

    SELECT html_template INTO tpl
      FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'info-card-grid';

    IF tpl IS NULL THEN
        RAISE EXCEPTION '368/info-card-grid: no active unforked row — STOP, do not widen the match';
    END IF;

    -- Pre-conditions: each rule-scoped needle exactly once; exactly 2 bare
    -- `color: var(--color-primary);` declarations in the whole template.
    c := (length(tpl) - length(replace(tpl, n_eyebrow, ''))) / length(n_eyebrow);
    IF c <> 1 THEN
        RAISE EXCEPTION '368/info-card-grid: eyebrow needle count % (expected 1) — template moved, STOP', c;
    END IF;
    c := (length(tpl) - length(replace(tpl, n_cardlink, ''))) / length(n_cardlink);
    IF c <> 1 THEN
        RAISE EXCEPTION '368/info-card-grid: card-link needle count % (expected 1) — template moved, STOP', c;
    END IF;
    c := (length(tpl) - length(replace(tpl, 'color: var(--color-primary);', ''))) / length('color: var(--color-primary);');
    IF c <> 2 THEN
        RAISE EXCEPTION '368/info-card-grid: bare color:primary count % (expected 2) — template moved, STOP', c;
    END IF;

    UPDATE content_components
       SET html_template = replace(replace(html_template, n_eyebrow, r_eyebrow), n_cardlink, r_cardlink),
           updated_at    = now()
     WHERE is_active AND forked_from IS NULL AND name = 'info-card-grid';

    -- Post-conditions: targets gone, non-targets intact.
    SELECT html_template INTO tpl
      FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'info-card-grid';

    c := (length(tpl) - length(replace(tpl, 'color: var(--color-primary);', ''))) / length('color: var(--color-primary);');
    IF c <> 0 THEN
        RAISE EXCEPTION '368/info-card-grid: % bare color:primary left after replace (expected 0)', c;
    END IF;
    c := (length(tpl) - length(replace(tpl, 'background: var(--color-primary);', ''))) / length('background: var(--color-primary);');
    IF c <> 1 THEN
        RAISE EXCEPTION '368/info-card-grid: background:primary count % after replace (expected 1 intact)', c;
    END IF;
    c := (length(tpl) - length(replace(tpl, 'outline: 2px solid var(--color-primary);', ''))) / length('outline: 2px solid var(--color-primary);');
    IF c <> 2 THEN
        RAISE EXCEPTION '368/info-card-grid: outline count % after replace (expected 2 intact)', c;
    END IF;
    c := (length(tpl) - length(replace(tpl, 'var(--color-primary-hover, var(--color-primary))', ''))) / length('var(--color-primary-hover, var(--color-primary))');
    IF c <> 1 THEN
        RAISE EXCEPTION '368/info-card-grid: hover-nested count % after replace (expected 1 intact)', c;
    END IF;

    RAISE NOTICE '368/info-card-grid: 2 foreground uses repointed to --color-primary-ink; 4 non-targets verified intact';
END $$;

COMMIT;

-- ============================================================================
-- ROLLBACK
-- ============================================================================
-- Either restore html_template from backup_368_info_card_grid.tsv, or unwrap:
--
-- UPDATE content_components
--    SET html_template = replace(html_template,
--        'color: var(--color-primary-ink, var(--color-primary));',
--        'color: var(--color-primary);'),
--        updated_at = now()
--  WHERE is_active AND forked_from IS NULL AND name = 'info-card-grid';
--
-- (Safe: the wrapped form occurs exactly twice, both ours — assert 2 before, 0 after.)
-- Pages re-rendered in between carry the old template until re-rendered again.
