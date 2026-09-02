-- 689_theme_kits.sql
--
-- Phase 1 of the theme-kit system (owner-approved plan, session "themes",
-- /home/ant/.claude/plans/please-think-hard-about-starry-locket.md).
--
-- WHAT THIS IS: a named, listable, forkable REGISTRY that bundles existing
-- library rows (layouts/palettes/typography_sets/content_components chrome)
-- into one selectable "look", plus a page_archetypes table replacing the
-- hardcoded defaultSectionsForPage Go switch in apply_gap_plan_action.go.
--
-- NAMING: `theme_kits`, deliberately NOT `themes` — `css_themes`, `theme_id`,
-- `needs_theme_review` and `forked_from_theme_id` already mean "one site's CSS
-- composition record" throughout this codebase (see css_themes below). A
-- second table called `themes` would make every `theme_id` in the tree
-- ambiguous. "Theme" stays the user-facing word; `theme_kits` is the
-- internal noun. Register entry: docs026_concept_register/register/
-- design-composition.md, cross-referenced against DES-003/DES-013/DES-042.
--
-- COUPLING MODEL: applying a kit MATERIALIZES defaults into a site's own rows
-- (site_specs, design_intent) — it is never a live FK the site stays bound
-- to. No render or resolve path outside this migration's own consumers reads
-- `theme_kits` directly; a site that adopts a kit can diverge on any field
-- immediately afterward with nothing checking "is this site themed".
--
-- LINEAGE COLUMNS mirror migration 025's idiom on palettes/layouts/
-- typography_sets exactly (origin/needs_review/forked_from_*_id/
-- source_site_id/source_domain/forked_at) so a kit can itself be forked from
-- another kit or from a site the same way those tables already work.

BEGIN;

CREATE TABLE IF NOT EXISTS theme_kits (
  id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name                 text UNIQUE NOT NULL,
  display_name         text NOT NULL,
  description          text,
  category             text,
  industry_tags        text[] NOT NULL DEFAULT '{}',

  layout_id            uuid REFERENCES layouts(id)            ON DELETE SET NULL,
  palette_id           uuid REFERENCES palettes(id)            ON DELETE SET NULL,  -- a TEMPLATE row; the site gets its own fork on apply, same as every other composition path
  typography_set_id    uuid REFERENCES typography_sets(id)     ON DELETE SET NULL,
  header_component_id  uuid REFERENCES content_components(id)  ON DELETE SET NULL,
  footer_component_id  uuid REFERENCES content_components(id)  ON DELETE SET NULL,

  -- Open slot for dimensions not built in Phase 1 (nav pattern, voice
  -- preset — see the plan §7). Deliberately jsonb, not new columns/tables,
  -- until a second kit needs to SHARE one, per the plan's own reasoning.
  extras               jsonb NOT NULL DEFAULT '{}',

  is_active                boolean NOT NULL DEFAULT true,
  needs_review             boolean NOT NULL DEFAULT false,
  origin                   text NOT NULL DEFAULT 'seed',   -- seed | site_fork | reference (Phase 5, not built yet)
  forked_from_theme_kit_id uuid REFERENCES theme_kits(id) ON DELETE SET NULL,
  source_site_id           uuid REFERENCES sites(id) ON DELETE SET NULL,
  source_domain            text,
  forked_at                timestamptz,

  created_by  text NOT NULL DEFAULT 'system',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_theme_kits_active ON theme_kits(is_active, needs_review) WHERE is_active;
CREATE INDEX IF NOT EXISTS idx_theme_kits_industry ON theme_kits USING gin(industry_tags);
CREATE INDEX IF NOT EXISTS idx_theme_kits_source_site ON theme_kits(source_site_id) WHERE source_site_id IS NOT NULL;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_theme_kits_updated_at') THEN
    CREATE TRIGGER trg_theme_kits_updated_at BEFORE UPDATE ON theme_kits
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END $$;

-- ── page_archetypes ─────────────────────────────────────────────────────
--
-- Replaces the hardcoded defaultSectionsForPage Go switch
-- (apply_gap_plan_action.go:995-1042) as a reusable, forkable, THREE-WAY
-- SCOPED table: a site-specific row beats a theme-kit row beats a fleet row
-- (both null). This is what lets a site declare its own durable
-- page-structure convention WITHOUT adopting any theme kit at all.
--
-- `sections` stores content_components.function names, matching exactly
-- what site_plan_sections.component_name and the planner's own vocabulary
-- already use — a concrete component id would be wrong here, since it
-- would pin a shared archetype to one specific site's fork.
--
-- match_kind mirrors the switch's own two-level structure: page_type checks
-- (exact) outrank page_name checks (bugs_open/015 — type is not localised,
-- a name can be). 'page_name_contains' exists because the switch's faq/
-- pricing cases use `key == X || strings.Contains(key, X)`, not equality.
CREATE TABLE IF NOT EXISTS page_archetypes (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  theme_kit_id     uuid REFERENCES theme_kits(id) ON DELETE CASCADE,
  site_id          uuid REFERENCES sites(id)      ON DELETE CASCADE,
  CHECK (NOT (theme_kit_id IS NOT NULL AND site_id IS NOT NULL)),

  match_kind       text NOT NULL CHECK (match_kind IN
                     ('page_type','page_name','page_name_contains','page_name_suffix','default')),
  match_value      text NOT NULL DEFAULT '',
  sections         jsonb NOT NULL,
  description      text,
  is_active        boolean NOT NULL DEFAULT true,
  origin           text NOT NULL DEFAULT 'seed',
  forked_from_archetype_id uuid REFERENCES page_archetypes(id) ON DELETE SET NULL,
  created_by       text NOT NULL DEFAULT 'system',
  created_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (theme_kit_id, site_id, match_kind, match_value)
);

CREATE INDEX IF NOT EXISTS idx_page_archetypes_theme_kit ON page_archetypes(theme_kit_id) WHERE theme_kit_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_page_archetypes_site ON page_archetypes(site_id) WHERE site_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_page_archetypes_fleet ON page_archetypes(match_kind, match_value)
  WHERE theme_kit_id IS NULL AND site_id IS NULL AND is_active;

-- ── site_specs aspect used by apply_theme_kit_action.go ────────────────
--
-- No schema needed — 'theme_kit_adoption' is a plain site_specs aspect,
-- append-only via the existing supersede-then-insert pattern. Documented
-- here for a reader of this migration:
--   {"theme_kit_id","theme_kit_name","applied_at","mode","applied":{...},"skipped":{...}}

-- ══════════════════════════════════════════════════════════════════════
-- SEED: fleet-scope page_archetypes — a verified, direct port of
-- defaultSectionsForPage's cases, PLUS one new case (section-index).
-- ══════════════════════════════════════════════════════════════════════
--
-- section-index is a genuine gap, not present in the Go switch: real
-- section-index pages (homegarden.uk's month pages) fall through to the
-- bare `default` today, missing structured slots entirely. [MEASURED
-- 2026-09-02, via calendar session + this session's own page_components
-- query]: most section-index pages currently render hero + generic-text-
-- block only; a few with real listings add content-listing. Seeded here
-- as hero + content-listing rather than perpetuating the generic-text
-- pattern, which is really the bare fallback bleeding through, not a
-- deliberate structure. Credit: calendar session.

INSERT INTO page_archetypes (theme_kit_id, site_id, match_kind, match_value, sections, description, origin)
VALUES
  (NULL, NULL, 'page_type', 'news-index', '["hero","news-listing","call-to-action"]'::jsonb,
    'Ported from defaultSectionsForPage (apply_gap_plan_action.go), 2026-09-02', 'seed'),
  (NULL, NULL, 'page_type', 'entity-directory', '["hero","directory-listing"]'::jsonb,
    'Ported from defaultSectionsForPage. directory-listing resolves query.business_directory at plan time (bugs_open/206) — a real per-site list, not fabricated content.', 'seed'),
  (NULL, NULL, 'page_type', 'section-index', '["hero","content-listing"]'::jsonb,
    'NEW, not in the Go switch — a real gap. See migration header. Credit: calendar session, 2026-09-02.', 'seed'),
  (NULL, NULL, 'page_name', 'faq', '["hero","faq","call-to-action"]'::jsonb,
    'Ported from defaultSectionsForPage (equality half of the faq case).', 'seed'),
  (NULL, NULL, 'page_name_contains', 'faq', '["hero","faq","call-to-action"]'::jsonb,
    'Ported from defaultSectionsForPage (contains half of the faq case).', 'seed'),
  (NULL, NULL, 'page_name', 'contact', '["contact-hero","contact-form","contact-info"]'::jsonb,
    'Ported from defaultSectionsForPage.', 'seed'),
  (NULL, NULL, 'page_name', 'pricing', '["hero","pricing","faq","call-to-action"]'::jsonb,
    'Ported from defaultSectionsForPage (equality half of the pricing case).', 'seed'),
  (NULL, NULL, 'page_name_contains', 'pricing', '["hero","pricing","faq","call-to-action"]'::jsonb,
    'Ported from defaultSectionsForPage (contains half of the pricing case).', 'seed'),
  (NULL, NULL, 'page_name', 'about', '["hero-about","about-content","call-to-action"]'::jsonb,
    'Ported from defaultSectionsForPage.', 'seed'),
  (NULL, NULL, 'page_name_suffix', 'guides-index', '["hero","guide-list"]'::jsonb,
    'Ported from defaultSectionsForPage. Verified live fleet pattern (bugs_open/206): mortgagecalculator.co.uk, idea.uk, gamesdesign.co.uk, relojistas.com.', 'seed'),
  (NULL, NULL, 'page_name', 'guide-index', '["hero","guide-list"]'::jsonb,
    'Ported from defaultSectionsForPage (the equality alternate in the guides-index case).', 'seed'),
  (NULL, NULL, 'page_name_suffix', 'tools-index', '["hero","tool-list"]'::jsonb,
    'Ported from defaultSectionsForPage. Sibling pattern (gamesdesign.co.uk, robot-hands.com, finetuning.uk, ai-agent-orchestration.com).', 'seed'),
  (NULL, NULL, 'page_name', 'tool-index', '["hero","tool-list"]'::jsonb,
    'Ported from defaultSectionsForPage (the equality alternate in the tools-index case).', 'seed'),
  (NULL, NULL, 'default', '', '["hero","generic-text-block","call-to-action"]'::jsonb,
    'Ported from defaultSectionsForPage''s bare default.', 'seed')
ON CONFLICT (theme_kit_id, site_id, match_kind, match_value) DO NOTHING;

-- ══════════════════════════════════════════════════════════════════════
-- SEED: 4 theme kits, built entirely from rows that already exist —
-- no new visual design in this phase, plumbing only.
--
-- Chrome pin IDs are HARDCODED, not resolved by function-name subquery.
-- [MEASURED 2026-09-02, credit: components session] content_components
-- .function is NOT unique after the canonical-row predicate — a naive
-- "SELECT id ... WHERE function='site-header'" would silently pick
-- whichever row wins an accidental alphabetical tiebreak. Verified
-- directly: chromePinEligibleSQL (component_library.go:334) requires
-- component_level IN ('site','header','footer','head') — 'site-header'/
-- 'site-footer' are component_level='section' and are NOT chrome-eligible
-- at all despite the matching name; 'header-theme-chrome'/
-- 'footer-theme-chrome' (component_level='site') are the only actually-
-- eligible rows for these functions, verified via a direct query before
-- writing this seed.
-- ══════════════════════════════════════════════════════════════════════

-- header-theme-chrome = 58fde68f-9190-4e5e-b6a5-ea21cf27a9af
-- footer-theme-chrome = e6347680-4c7c-448b-8cfc-1cea509159d1
-- Both verified chrome-eligible (component_level='site', is_active) directly
-- against the live DB on 2026-09-02, per the header note above.

INSERT INTO theme_kits (name, display_name, description, category, industry_tags,
                         layout_id, palette_id, typography_set_id,
                         header_component_id, footer_component_id, origin)
VALUES
  ('brochure-formal-classic', 'Brochure — Formal Classic',
   'A formal brochure layout with the default professional palette and modern sans typography.',
   'brochure', ARRAY['professional','services'],
   (SELECT id FROM layouts WHERE name = 'brochure-formal'),
   (SELECT id FROM palettes WHERE name = 'default'),
   (SELECT id FROM typography_sets WHERE name = 'sans-modern'),
   '58fde68f-9190-4e5e-b6a5-ea21cf27a9af'::uuid,
   'e6347680-4c7c-448b-8cfc-1cea509159d1'::uuid,
   'seed'),
  ('docs-technical', 'Docs — Technical Reference',
   'A sidebar-navigated docs layout with an engineering-clean palette and monospace-accented typography.',
   'docs', ARRAY['documentation','technical','saas'],
   (SELECT id FROM layouts WHERE name = 'docs-sidebar'),
   (SELECT id FROM palettes WHERE name = 'modern-engineering-clean'),
   (SELECT id FROM typography_sets WHERE name = 'mono-technical'),
   '58fde68f-9190-4e5e-b6a5-ea21cf27a9af'::uuid,
   'e6347680-4c7c-448b-8cfc-1cea509159d1'::uuid,
   'seed'),
  ('soft-editorial', 'Soft — Editorial',
   'A light editorial layout with a calm-minimal palette and serif-editorial typography.',
   'editorial', ARRAY['editorial','content','magazine'],
   (SELECT id FROM layouts WHERE name = 'soft-editorial'),
   (SELECT id FROM palettes WHERE name = 'calm-minimal'),
   (SELECT id FROM typography_sets WHERE name = 'serif-editorial'),
   '58fde68f-9190-4e5e-b6a5-ea21cf27a9af'::uuid,
   'e6347680-4c7c-448b-8cfc-1cea509159d1'::uuid,
   'seed'),
  ('tool-portal-light', 'Tool Portal — Light',
   'A light, interactive tool-portal layout with a modern-content palette and friendly sans typography.',
   'tool', ARRAY['interactive','calculator','tool-portal'],
   (SELECT id FROM layouts WHERE name = 'tool-portal-light'),
   (SELECT id FROM palettes WHERE name = 'content-modern'),
   (SELECT id FROM typography_sets WHERE name = 'sans-friendly'),
   '58fde68f-9190-4e5e-b6a5-ea21cf27a9af'::uuid,
   'e6347680-4c7c-448b-8cfc-1cea509159d1'::uuid,
   'seed')
ON CONFLICT (name) DO NOTHING;

DO $$
DECLARE missing_fk int;
BEGIN
  -- Guard: refuse silently-NULL FKs from a subquery typo (name mismatch)
  -- rather than ship a kit that references nothing.
  SELECT count(*) INTO missing_fk FROM theme_kits
   WHERE origin = 'seed' AND created_at > now() - interval '5 minutes'
     AND (layout_id IS NULL OR palette_id IS NULL OR typography_set_id IS NULL);
  IF missing_fk > 0 THEN
    RAISE EXCEPTION 'theme_kits seed: % row(s) have a NULL layout/palette/typography FK — a name subquery did not match', missing_fk;
  END IF;
END $$;

COMMIT;
