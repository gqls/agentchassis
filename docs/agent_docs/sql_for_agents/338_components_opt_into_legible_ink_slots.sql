-- 338_components_opt_into_legible_ink_slots.sql
--
-- bugs_open/122 — the CONFIG half of the ink-slots fix. The engine half (register
-- VIZ-014, buildLegibleInkDefaults) is live and pod-proven on v1.0.1266; it emits
-- --color-accent-text / --color-primary-ink / --color-accent-ink from the
-- renderer's :root block as step 12, and NOTHING CONSUMES THEM YET. This file is
-- what makes a visitor's page change.
--
-- Council: APPROVED round 2, correlation c4d9c841-3658-4742-85b5-961e062ecad2
-- (submission SUBMISSION_2026-08-06b_ink_slots_round2.json, edit 7).
--
-- OWNING PIPELINE: brochure_component_library owns content_components.html_template
-- and the layouts theme half. Named here at the edit, per the guardian's objection.
--
-- SAFE BY FALLBACK. Every replacement is var(<new>, <EXACTLY what is there today>),
-- so a page rendered against a stylesheet that predates the engine change resolves
-- the fallback and renders byte-identically. There is no flag day.
--
-- ============================================================================
-- WHY THIS FILE DOES NOT MATCH THE APPROVED SKETCH, AND WHY THAT IS THE POINT
-- ============================================================================
-- The approved sketch used a bare global replace() per component. Measured against
-- the live rows on 2026-08-08, that would have been WRONG on three of the four
-- components and a visible regression on one:
--
--   component               needle occurrences   intended targets
--   case-studies-grid        2                    2   <- sketch correct
--   system-stats             5                    1   (.stats-eyebrow)
--   image-hover-card-grid    2                    1   (__eyebrow)
--   tool-list                6                    2   (.tl-eyebrow, .tl-card-link)
--
-- replace() is GLOBAL WITHIN THE STRING, and the sketch's gate counted ROWS, so it
-- could not see this. In tool-list the four non-targets include
--   .tl-card-icon { background: var(--color-primary); }
--   .tl-cta-btn   { background: var(--color-primary); }
-- i.e. repointing a BACKGROUND to an ink colour — the exact inversion this whole
-- bug is about. In system-stats two of the five are `background:` and one is an
-- `outline:`; in image-hover-card-grid the second is an `outline:`.
--
-- So every needle below is RULE-SCOPED: it carries enough surrounding declaration
-- text to name one rule, and each DO block asserts the occurrence count it expects.
-- If a template moves under this migration the count changes and the RAISE fires.
--
-- The layouts half also did not match. The sketch assumed `a { color: var(--color-accent); }`
-- on one line; that form exists in 0 of 18 layouts. Four of the five use a
-- multi-line rule and tool-portal-light uses a different single-line form, so they
-- are two separate needles below.
--
-- ============================================================================
-- BACKUP — TAKE THIS FIRST, OUTSIDE THE TRANSACTION
-- ============================================================================
-- \copy (SELECT id,name,html_template FROM content_components WHERE is_active AND forked_from IS NULL AND name IN ('case-studies-grid','system-stats','image-hover-card-grid','tool-list')) TO 'backup_338_content_components.tsv'
-- \copy (SELECT id,name,css_template FROM layouts WHERE is_active AND name IN ('high-energy','tool-portal-dark','brochure-formal','technical-precise','tool-portal-light')) TO 'backup_338_layouts.tsv'
--
-- Rollback is at the foot of this file.

BEGIN;

-- ============================================================================
-- 1. case-studies-grid — .csg-cta-btn and .csg-filter-btn.active
--    Both fills are var(--color-accent,…) but both ink with --color-primary-text,
--    which is the ink for a PRIMARY fill. Wrong slot; the value is fine.
--    Closes: finetuning.uk x2.
--    Both occurrences ARE targets (verified 2026-08-08: L107 .csg-filter-btn.active,
--    L258 .csg-cta-btn, both `color:`), so the plain needle is safe here.
-- ============================================================================
DO $$
DECLARE
    needle   text := 'var(--color-primary-text, #fff)';
    repl     text := 'var(--color-accent-text, var(--color-primary-text, #fff))';
    hits     int;
    rows_hit int;
BEGIN
    SELECT (length(html_template) - length(replace(html_template, needle, '')))
           / length(needle)
      INTO hits
      FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'case-studies-grid';

    IF hits IS NULL THEN
        RAISE EXCEPTION '338/case-studies-grid: no active unforked row — STOP, do not widen the match';
    END IF;
    IF hits <> 2 THEN
        RAISE EXCEPTION '338/case-studies-grid: expected 2 occurrences of the needle, found %. '
                        'The template moved under this migration; re-read it rather than widening.', hits;
    END IF;

    UPDATE content_components
       SET html_template = replace(html_template, needle, repl),
           updated_at    = now()
     WHERE is_active AND forked_from IS NULL AND name = 'case-studies-grid';
    GET DIAGNOSTICS rows_hit = ROW_COUNT;
    IF rows_hit <> 1 THEN
        RAISE EXCEPTION '338/case-studies-grid: expected 1 row updated, got %', rows_hit;
    END IF;

    PERFORM 1 FROM content_components
      WHERE is_active AND forked_from IS NULL AND name = 'case-studies-grid'
        AND position('var(--color-accent-text,' in html_template) > 0;
    IF NOT FOUND THEN
        RAISE EXCEPTION '338/case-studies-grid: post-state missing --color-accent-text; rolling back';
    END IF;
END $$;

-- ============================================================================
-- 2. system-stats — .stats-eyebrow ONLY (1 of 5 accent uses in this template).
--    The other four: .stat-card::after background, .stat-suffix color,
--    .stats-cta background, .stats-cta:focus outline. None are this defect.
--    Closes: gamesdesign.co.uk x1, vonc.com x1.
--    NB: this does NOT touch the .system-stats-section --section-* literals —
--    that is bugs_open/212, a different class with a different repair.
-- ============================================================================
DO $$
DECLARE
    needle   text := E'    text-transform: uppercase;\n    color: var(--color-accent, #7dd3fc);\n    margin-bottom: 1rem;';
    repl     text := E'    text-transform: uppercase;\n    color: var(--color-accent-ink, var(--color-accent, #7dd3fc));\n    margin-bottom: 1rem;';
    hits     int;
    rows_hit int;
BEGIN
    SELECT (length(html_template) - length(replace(html_template, needle, '')))
           / length(needle)
      INTO hits
      FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'system-stats';

    IF hits IS NULL THEN
        RAISE EXCEPTION '338/system-stats: no active unforked row — STOP';
    END IF;
    IF hits <> 1 THEN
        RAISE EXCEPTION '338/system-stats: expected exactly 1 .stats-eyebrow rule match, found %. '
                        'A global replace here would repaint two backgrounds and an outline — STOP.', hits;
    END IF;

    UPDATE content_components
       SET html_template = replace(html_template, needle, repl),
           updated_at    = now()
     WHERE is_active AND forked_from IS NULL AND name = 'system-stats';
    GET DIAGNOSTICS rows_hit = ROW_COUNT;
    IF rows_hit <> 1 THEN
        RAISE EXCEPTION '338/system-stats: expected 1 row updated, got %', rows_hit;
    END IF;

    -- post-state, and the NEGATIVE control: the four non-targets must be untouched.
    PERFORM 1 FROM content_components
      WHERE is_active AND forked_from IS NULL AND name = 'system-stats'
        AND position('var(--color-accent-ink,' in html_template) > 0;
    IF NOT FOUND THEN
        RAISE EXCEPTION '338/system-stats: post-state missing --color-accent-ink; rolling back';
    END IF;

    SELECT (length(html_template) - length(replace(html_template, 'var(--color-accent, #7dd3fc)', '')))
           / length('var(--color-accent, #7dd3fc)')
      INTO hits
      FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'system-stats';
    IF hits <> 4 THEN
        RAISE EXCEPTION '338/system-stats: expected 4 untouched accent uses to remain, found % — '
                        'the replace over-applied; rolling back', hits;
    END IF;
END $$;

-- ============================================================================
-- 3. image-hover-card-grid — __eyebrow ONLY (the other bare --color-primary is
--    an `outline:` on __card:focus-visible).
--    Closes: dartsonline.co.uk x1.
-- ============================================================================
DO $$
DECLARE
    needle   text := E'    text-transform: uppercase;\n    color: var(--color-primary);\n    margin-bottom: 0.75rem;';
    repl     text := E'    text-transform: uppercase;\n    color: var(--color-primary-ink, var(--color-primary));\n    margin-bottom: 0.75rem;';
    hits     int;
    rows_hit int;
BEGIN
    SELECT (length(html_template) - length(replace(html_template, needle, '')))
           / length(needle)
      INTO hits
      FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'image-hover-card-grid';

    IF hits IS NULL THEN
        RAISE EXCEPTION '338/image-hover-card-grid: no active unforked row — STOP';
    END IF;
    IF hits <> 1 THEN
        RAISE EXCEPTION '338/image-hover-card-grid: expected exactly 1 __eyebrow match, found % — STOP', hits;
    END IF;

    UPDATE content_components
       SET html_template = replace(html_template, needle, repl),
           updated_at    = now()
     WHERE is_active AND forked_from IS NULL AND name = 'image-hover-card-grid';
    GET DIAGNOSTICS rows_hit = ROW_COUNT;
    IF rows_hit <> 1 THEN
        RAISE EXCEPTION '338/image-hover-card-grid: expected 1 row updated, got %', rows_hit;
    END IF;

    PERFORM 1 FROM content_components
      WHERE is_active AND forked_from IS NULL AND name = 'image-hover-card-grid'
        AND position('var(--color-primary-ink,' in html_template) > 0;
    IF NOT FOUND THEN
        RAISE EXCEPTION '338/image-hover-card-grid: post-state missing --color-primary-ink; rolling back';
    END IF;

    -- the outline on __card:focus-visible must survive
    SELECT (length(html_template) - length(replace(html_template, 'outline: 3px solid var(--color-primary)', '')))
           / length('outline: 3px solid var(--color-primary)')
      INTO hits
      FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'image-hover-card-grid';
    IF hits <> 1 THEN
        RAISE EXCEPTION '338/image-hover-card-grid: the focus-visible outline was altered (% left) — rolling back', hits;
    END IF;
END $$;

-- ============================================================================
-- 4. tool-list — .tl-eyebrow and .tl-card-link ONLY (2 of 6 bare --color-primary).
--    The four non-targets: .tl-card-icon background, .tl-cta-btn background, and
--    two :hover rules where it sits inside var(--color-primary-hover, …).
--    Closes: robot-hands.com x2.
--    This template is one-rule-per-line, so the needles are single-line.
-- ============================================================================
DO $$
DECLARE
    n_eyebrow text := 'text-transform: uppercase; color: var(--color-primary); margin-bottom: 0.75rem;';
    r_eyebrow text := 'text-transform: uppercase; color: var(--color-primary-ink, var(--color-primary)); margin-bottom: 0.75rem;';
    n_link    text := 'font-weight: 600; color: var(--color-primary); text-decoration: none;';
    r_link    text := 'font-weight: 600; color: var(--color-primary-ink, var(--color-primary)); text-decoration: none;';
    hits      int;
    rows_hit  int;
BEGIN
    SELECT (length(html_template) - length(replace(html_template, n_eyebrow, ''))) / length(n_eyebrow)
      INTO hits FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'tool-list';
    IF hits IS NULL THEN
        RAISE EXCEPTION '338/tool-list: no active unforked row — STOP';
    END IF;
    IF hits <> 1 THEN
        RAISE EXCEPTION '338/tool-list: expected exactly 1 .tl-eyebrow match, found % — STOP', hits;
    END IF;

    SELECT (length(html_template) - length(replace(html_template, n_link, ''))) / length(n_link)
      INTO hits FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'tool-list';
    IF hits <> 1 THEN
        RAISE EXCEPTION '338/tool-list: expected exactly 1 .tl-card-link match, found % — STOP', hits;
    END IF;

    UPDATE content_components
       SET html_template = replace(replace(html_template, n_eyebrow, r_eyebrow), n_link, r_link),
           updated_at    = now()
     WHERE is_active AND forked_from IS NULL AND name = 'tool-list';
    GET DIAGNOSTICS rows_hit = ROW_COUNT;
    IF rows_hit <> 1 THEN
        RAISE EXCEPTION '338/tool-list: expected 1 row updated, got %', rows_hit;
    END IF;

    SELECT (length(html_template) - length(replace(html_template, 'var(--color-primary-ink,', '')))
           / length('var(--color-primary-ink,')
      INTO hits FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'tool-list';
    IF hits <> 2 THEN
        RAISE EXCEPTION '338/tool-list: expected 2 --color-primary-ink references, found % — rolling back', hits;
    END IF;

    -- NEGATIVE CONTROL: the two backgrounds must be untouched.
    SELECT (length(html_template) - length(replace(html_template, 'background: var(--color-primary);', '')))
           / length('background: var(--color-primary);')
      INTO hits FROM content_components
     WHERE is_active AND forked_from IS NULL AND name = 'tool-list';
    IF hits <> 2 THEN
        RAISE EXCEPTION '338/tool-list: expected 2 untouched background: var(--color-primary) rules, found % — '
                        'an ink replacement hit a FILL; rolling back', hits;
    END IF;
END $$;

-- ============================================================================
-- 5. layouts — the base `a {}` rule, 4 layouts with the multi-line form.
--    17 of 18 layouts colour an ink with var(--color-primary); 5 colour the base
--    link with accent. These 4 share one byte sequence.
--    Closes: gaswholesalers.com x6 (all six are base links).
-- ============================================================================
DO $$
DECLARE
    needle   text := E'a {\n  color: var(--color-accent);';
    repl     text := E'a {\n  color: var(--color-accent-ink, var(--color-accent));';
    targets  text[] := ARRAY['high-energy','tool-portal-dark','brochure-formal','technical-precise'];
    hits     int;
    rows_hit int;
BEGIN
    SELECT count(*) INTO hits
      FROM layouts
     WHERE is_active AND name = ANY(targets)
       AND position(needle in css_template) > 0;
    IF hits <> 4 THEN
        RAISE EXCEPTION '338/layouts: expected the multi-line anchor rule in all 4 layouts, found in %. '
                        'A layout was re-themed under this migration; re-read before widening.', hits;
    END IF;

    UPDATE layouts
       SET css_template = replace(css_template, needle, repl)
     WHERE is_active AND name = ANY(targets);
    GET DIAGNOSTICS rows_hit = ROW_COUNT;
    IF rows_hit <> 4 THEN
        RAISE EXCEPTION '338/layouts: expected 4 rows updated, got %', rows_hit;
    END IF;

    SELECT count(*) INTO hits
      FROM layouts
     WHERE is_active AND name = ANY(targets)
       AND position('var(--color-accent-ink,' in css_template) > 0;
    IF hits <> 4 THEN
        RAISE EXCEPTION '338/layouts: post-state present in only % of 4; rolling back', hits;
    END IF;
END $$;

-- ============================================================================
-- 6. layouts — tool-portal-light, which uses the SINGLE-LINE form.
--    Split out because its byte sequence differs; the sketch's one-line needle
--    matched 0 of 18 and this is the only layout even close to it.
-- ============================================================================
DO $$
DECLARE
    needle   text := 'a { color: var(--color-accent); text-decoration: none;';
    repl     text := 'a { color: var(--color-accent-ink, var(--color-accent)); text-decoration: none;';
    hits     int;
    rows_hit int;
BEGIN
    SELECT (length(css_template) - length(replace(css_template, needle, ''))) / length(needle)
      INTO hits FROM layouts WHERE is_active AND name = 'tool-portal-light';
    IF hits IS NULL THEN
        RAISE EXCEPTION '338/tool-portal-light: no active row — STOP';
    END IF;
    IF hits <> 1 THEN
        RAISE EXCEPTION '338/tool-portal-light: expected exactly 1 base link rule, found % — STOP', hits;
    END IF;

    UPDATE layouts
       SET css_template = replace(css_template, needle, repl)
     WHERE is_active AND name = 'tool-portal-light';
    GET DIAGNOSTICS rows_hit = ROW_COUNT;
    IF rows_hit <> 1 THEN
        RAISE EXCEPTION '338/tool-portal-light: expected 1 row updated, got %', rows_hit;
    END IF;

    PERFORM 1 FROM layouts
     WHERE is_active AND name = 'tool-portal-light'
       AND position('var(--color-accent-ink,' in css_template) > 0;
    IF NOT FOUND THEN
        RAISE EXCEPTION '338/tool-portal-light: post-state missing; rolling back';
    END IF;
END $$;

COMMIT;

-- ============================================================================
-- PROPAGATION — READ THIS, THE MIGRATION IS NOT THE FIX
-- ============================================================================
-- The UPDATEs above change the SOURCE. A visitor sees nothing until the affected
-- pages re-render. This was the council's open `editquality` MEDIUM objection and
-- listing the rows is NOT enough — they must be enqueued.
--
-- The affected placements (measured 2026-08-06: 16 across <=4 sites each):
--
--   SELECT s.domain, p.url, cc.name
--     FROM page_components pc
--     JOIN pages p            ON p.id = pc.page_id
--     JOIN sites s            ON s.id = p.site_id
--     JOIN content_components cc ON cc.id = pc.component_id
--    WHERE cc.name IN ('case-studies-grid','system-stats','image-hover-card-grid','tool-list')
--    ORDER BY s.domain, p.url;
--
-- Layout changes are wider: every site on the 5 layouts re-renders its stylesheet.
--
-- VERIFICATION IS THE SERVED PAGE, never this file's row counts and never a
-- 'complete' work item (see bugs_open/213 for why that is not idle advice — a
-- work item on exactly this defect class closed `complete` having written nothing).
--   python3 scripts/render_audit.py <urls> > after_$(date +%F).txt
-- and diff PER SELECTOR against BASELINE_2026-08-06_render_audit.txt.
--
-- ============================================================================
-- ROLLBACK — inverse replace(), same shape. Uncomment to run.
-- ============================================================================
-- BEGIN;
-- UPDATE content_components SET html_template = replace(html_template,
--          'var(--color-accent-text, var(--color-primary-text, #fff))',
--          'var(--color-primary-text, #fff)')
--  WHERE is_active AND forked_from IS NULL AND name = 'case-studies-grid';
-- UPDATE content_components SET html_template = replace(html_template,
--          'var(--color-accent-ink, var(--color-accent, #7dd3fc))',
--          'var(--color-accent, #7dd3fc)')
--  WHERE is_active AND forked_from IS NULL AND name = 'system-stats';
-- UPDATE content_components SET html_template = replace(html_template,
--          'var(--color-primary-ink, var(--color-primary))',
--          'var(--color-primary)')
--  WHERE is_active AND forked_from IS NULL AND name IN ('image-hover-card-grid','tool-list');
-- UPDATE layouts SET css_template = replace(css_template,
--          'var(--color-accent-ink, var(--color-accent))',
--          'var(--color-accent)')
--  WHERE is_active AND name IN ('high-energy','tool-portal-dark','brochure-formal',
--                               'technical-precise','tool-portal-light');
-- COMMIT;
-- (and backup_338_*.tsv above are the belt-and-braces copy)
