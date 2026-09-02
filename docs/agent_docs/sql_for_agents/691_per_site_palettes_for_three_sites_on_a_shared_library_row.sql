-- 691_per_site_palettes_for_three_sites_on_a_shared_library_row.sql
--
-- RFC_059's "smaller fix", per the owner's decision 2026-09-02: withdraw the
-- structural render-merge pin, and instead make the two stores AGREE where they
-- disagree. Three sites are served correct colours ONLY because the LLM overlay
-- carries design_intent.palette.reference_values through at render time; their
-- composition row (css_themes.palette_id) points at the SHARED library palette
-- `professional-dark`, whose values they do not use.
--
-- WHY THAT MATTERS EVEN THOUGH THE SITES LOOK RIGHT: any render path that ever
-- skips or short-circuits the overlay serves them the generic default instead.
-- RFC_059's own proposed fix would have done exactly that, and two of these
-- three were the RFC's nominated canaries — the fix would have repainted the
-- sites it was watching to prove itself safe. The composition row should hold
-- the truth on its own.
--
-- THIS IS APPEARANCE-NEUTRAL BY CONSTRUCTION. The colours written below are the
-- values these sites ALREADY SERVE, read from the live stylesheets on
-- 2026-09-02 before writing this file:
--   finetuning.uk       --color-primary #1A1A2E  --color-background #F5F3EF
--   gaswholesalers.com  --color-primary #1A1A2E  --color-background #F4F1EB
--   cv1.co.uk           --color-primary #1a2e44  --color-background #f5f7fa
-- and they match each site's reference_values exactly. After this, the palette
-- row says what the browser already shows.
--
-- IT ALSO ENDS A PALETTE-PHILOSOPHY VIOLATION. resolve_composition_pallette_
-- action.go's own header states it plainly: "Palettes are ALWAYS site-specific
-- ... Library reuse of palettes would be lying — 'this site's brand colour is
-- #c8102e' is a statement about one specific site, not a reusable template."
-- These three share one row. `professional-dark` keeps existing as a library
-- seed; it simply stops being any site's live palette. Same clone-and-repoint
-- shape as DES-053.
--
-- ⚠ loanzy.uk IS DELIBERATELY EXCLUDED. It is the fourth divergent site, but a
-- different case: it serves #2A9D8F / #F8FAF9, which matches NEITHER its palette
-- row (#1B4F72 / #F8FAFC) NOR its reference_values (#1A7F6E / #F8F9FA) — three
-- different values. That is the render overlay having produced its own, which
-- under the owner's 2026-09-02 ruling is legitimate authority, not a mismatch to
-- reconcile. Choosing among the three is a judgement about that site and belongs
-- to its lane, not to this migration.

BEGIN;

DO $$
DECLARE
    r            record;
    new_pal_id   uuid;
    moved        int := 0;
BEGIN
    FOR r IN
        SELECT s.id            AS site_id,
               s.domain        AS domain,
               ct.id           AS css_theme_id,
               p.id            AS old_palette_id,
               p.name          AS old_palette_name,
               p.category      AS category,
               p.industry_tags AS industry_tags,
               ss.data->'palette'->'reference_values' AS ref
          FROM sites s
          JOIN style_collections sc ON sc.id = s.style_collection_id
          JOIN css_themes        ct ON ct.id = sc.css_theme_id
          JOIN palettes          p  ON p.id  = ct.palette_id
          JOIN site_specs        ss ON ss.site_id = s.id
                                   AND ss.aspect = 'design_intent'
                                   AND ss.is_current
         WHERE s.domain IN ('cv1.co.uk','finetuning.uk','gaswholesalers.com')
           AND p.name = 'professional-dark'
           AND jsonb_typeof(ss.data->'palette'->'reference_values') = 'object'
    LOOP
        -- Refuse rather than half-do it: a site whose reference_values lack the
        -- two slots we verified against the live stylesheet is not the case this
        -- migration measured, and must not be repointed on a guess.
        IF (r.ref->>'primary') IS NULL OR (r.ref->>'background') IS NULL THEN
            RAISE EXCEPTION '% has reference_values without primary/background — not the measured case, refusing', r.domain;
        END IF;

        INSERT INTO palettes (
            name, display_name, description, colours,
            category, industry_tags,
            is_active, origin, needs_review,
            forked_from_palette_id, source_site_id, source_domain, forked_at
        ) VALUES (
            'palette-' || replace(replace(r.domain, '.', '-'), '_', '-') || '-' || left(replace(r.site_id::text, '-', ''), 8),
            'Palette for ' || r.domain,
            'Per-site palette carrying the values this site already serves; '
              || 'split off the shared library row ' || r.old_palette_name
              || ' so the composition row no longer depends on the render overlay to be correct (691).',
            r.ref,
            r.category, r.industry_tags,
            true, 'site_split', false,
            r.old_palette_id, r.site_id, r.domain, now()
        )
        RETURNING id INTO new_pal_id;

        UPDATE css_themes SET palette_id = new_pal_id, updated_at = now()
         WHERE id = r.css_theme_id;

        moved := moved + 1;
        RAISE NOTICE '691: % -> own palette % (was %)', r.domain, new_pal_id, r.old_palette_name;
    END LOOP;

    IF moved <> 3 THEN
        RAISE EXCEPTION '691 expected to move exactly 3 sites, moved % — the population changed since it was measured (2026-09-02); re-measure before applying', moved;
    END IF;
END $$;

-- Prove it: every one of the three now has its own palette whose core slots
-- equal its design_intent reference_values. A row that fails this is the
-- migration having done nothing useful, so fail the transaction.
DO $$
DECLARE bad int;
BEGIN
    SELECT count(*) INTO bad
      FROM sites s
      JOIN style_collections sc ON sc.id = s.style_collection_id
      JOIN css_themes        ct ON ct.id = sc.css_theme_id
      JOIN palettes          p  ON p.id  = ct.palette_id
      JOIN site_specs        ss ON ss.site_id = s.id
                               AND ss.aspect = 'design_intent'
                               AND ss.is_current
     WHERE s.domain IN ('cv1.co.uk','finetuning.uk','gaswholesalers.com')
       AND (p.source_site_id IS DISTINCT FROM s.id
            OR p.colours->>'primary'    IS DISTINCT FROM ss.data->'palette'->'reference_values'->>'primary'
            OR p.colours->>'background' IS DISTINCT FROM ss.data->'palette'->'reference_values'->>'background');
    IF bad > 0 THEN
        RAISE EXCEPTION '691 verify: % of 3 sites still disagree with their reference_values', bad;
    END IF;
END $$;

COMMIT;
