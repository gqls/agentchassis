-- ============================================================================
-- BLOCK A: DB migrations for the work-item build pipeline (Phase 0)
--
-- No Go changes needed. Can be applied immediately.
-- Run against clients_db.
--
-- Contents:
--   1. site_specs table
--   2. build_queue table
--   3. page_component_history table
--   4. approval_mode column on site_work_items
--   5. page_spec column on pages
--   6. Drop content_snapshot / schema_snapshot from page_components
--
-- NOTE on item 6: The Go struct PageComponentInstance has a SchemaSnapshot
-- field (omitempty). After applying this migration, remove that struct field
-- from the Go code. It's not actively used by any action — just a leftover
-- from an earlier design that this migration replaces.
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. site_specs
--
-- Stores all site specification data (identity, strategy, tone, visual_direction,
-- design, image_guidance, structure, marketing, adoption_source).
-- Each row is a complete record for one aspect. write_site_spec deep-merges
-- partial updates before inserting, so every row is self-contained.
-- History via is_current flag — old versions remain for rollback/audit.
-- ============================================================================

CREATE TABLE IF NOT EXISTS site_specs (
                                          id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    aspect          text NOT NULL,
    data            jsonb NOT NULL,

    -- Provenance
    source          text NOT NULL,       -- 'classifier', 'adoption', 'hitl',
-- 'planner', 'improvement', 'seed', 'manual',
-- 'rollback', 'fork', 'recovery'
    source_agent    text,                -- agent type that wrote this
    source_item_id  uuid,                -- work item that caused this write
    notes           text,                -- human-readable reason

-- Currency
    is_current      boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    superseded_at   timestamptz,
    created_by      text NOT NULL
    );

-- One current spec per site per aspect (enforced)
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_specs_current
    ON site_specs (site_id, aspect)
    WHERE is_current = true;

-- Fast lookup: all current specs for a site
CREATE INDEX IF NOT EXISTS idx_site_specs_lookup
    ON site_specs (site_id)
    WHERE is_current = true;

-- History queries: all versions of a given aspect, newest first
CREATE INDEX IF NOT EXISTS idx_site_specs_history
    ON site_specs (site_id, aspect, created_at DESC);

-- Find specs written by a particular work item
CREATE INDEX IF NOT EXISTS idx_site_specs_source_item
    ON site_specs (source_item_id)
    WHERE source_item_id IS NOT NULL;


-- ============================================================================
-- 2. build_queue
--
-- Domain queue for seeding new sites into the work-item pipeline.
-- seed_build_queue action reads from here, creates site records and
-- initial work items based on the direction field.
-- ============================================================================

CREATE TABLE IF NOT EXISTS build_queue (
                                           id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    domain          text UNIQUE NOT NULL,
    direction       jsonb,               -- null, {objective}, {adopt_from}, {fork_from}, {brief_complete: true, ...}
    status          text NOT NULL DEFAULT 'queued',
    batch_id        uuid,
    priority        integer DEFAULT 100,
    created_at      timestamptz DEFAULT now(),
    updated_at      timestamptz DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_build_queue_status
    ON build_queue (status, priority)
    WHERE status = 'queued';


-- ============================================================================
-- 3. page_component_history
--
-- Tracks previous content_data values for page_components.
-- Before any content_data write, the current value is copied here.
-- Each row is a complete snapshot (not a diff).
-- Replaces the unused content_snapshot/schema_snapshot columns.
-- ============================================================================

CREATE TABLE IF NOT EXISTS page_component_history (
                                                      id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    component_id    uuid REFERENCES page_components(id) ON DELETE SET NULL,
    page_id         uuid NOT NULL REFERENCES pages(id),
    site_id         uuid NOT NULL REFERENCES sites(id),
    content_data    jsonb NOT NULL,

    source          text NOT NULL,       -- 'content-writer', 'section-editor', 'rollback', etc.
    source_item_id  uuid,                -- work item that triggered the change
    created_at      timestamptz NOT NULL DEFAULT now()
    );

-- Recent history per component (for rollback)
CREATE INDEX IF NOT EXISTS idx_pch_component
    ON page_component_history (component_id, created_at DESC);

-- Recent history per site (for site-wide audit)
CREATE INDEX IF NOT EXISTS idx_pch_site
    ON page_component_history (site_id, created_at DESC);

-- Find history entries from a specific work item
CREATE INDEX IF NOT EXISTS idx_pch_source_item
    ON page_component_history (source_item_id)
    WHERE source_item_id IS NOT NULL;


-- ============================================================================
-- 4. approval_mode on site_work_items
--
-- Controls whether items auto-dispatch or require human/eval approval.
-- Values: 'auto' (default), 'hitl', 'eval'
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'site_work_items' AND column_name = 'approval_mode'
    ) THEN
ALTER TABLE site_work_items ADD COLUMN approval_mode text DEFAULT 'auto';
END IF;
END $$;


-- ============================================================================
-- 5. page_spec on pages
--
-- JSONB column for page-level specification data: content hints, existing
-- content (for adoption), content_direction for rewrites. Nullable — most
-- pages won't have one until adoption or improvement agents write specs.
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'pages' AND column_name = 'page_spec'
    ) THEN
ALTER TABLE pages ADD COLUMN page_spec jsonb;
END IF;
END $$;


-- ============================================================================
-- 6. Drop content_snapshot / schema_snapshot from page_components
--
-- These are unused. page_component_history replaces them.
-- The Go struct has SchemaSnapshot (omitempty) — remove it after this runs.
--
-- v_section_schema_status references both columns. Not used in Go code
-- but useful as a diagnostic view, so recreate it without the snapshot columns.
-- ============================================================================

DROP VIEW IF EXISTS v_section_schema_status;
ALTER TABLE page_components DROP COLUMN IF EXISTS content_snapshot;
ALTER TABLE page_components DROP COLUMN IF EXISTS schema_snapshot;

CREATE VIEW v_section_schema_status AS
SELECT pc.id AS page_component_id,
       pc.page_id,
       p.name AS page_name,
       s.domain,
       cc.name AS component_name,
       cc.function AS component_function,
       COALESCE(pc.schema_mode, s.schema_mode, 'flexible'::text) AS effective_schema_mode,
       pc.locked_at,
       pc.locked_by,
       pc.build_status,
       pc.reviewed_at
FROM page_components pc
         JOIN pages p ON pc.page_id = p.id
         JOIN sites s ON p.site_id = s.id
         LEFT JOIN content_components cc ON pc.component_id = cc.id;


COMMIT;

-- ============================================================================
-- Verification queries (run after migration to confirm)
-- ============================================================================

-- SELECT count(*) FROM information_schema.tables WHERE table_name IN ('site_specs', 'build_queue', 'page_component_history');
-- Expected: 3

-- SELECT column_name FROM information_schema.columns WHERE table_name = 'site_work_items' AND column_name = 'approval_mode';
-- Expected: 1 row

-- SELECT column_name FROM information_schema.columns WHERE table_name = 'pages' AND column_name = 'page_spec';
-- Expected: 1 row

-- SELECT column_name FROM information_schema.columns WHERE table_name = 'page_components' AND column_name IN ('content_snapshot', 'schema_snapshot');
-- Expected: 0 rows

--

----

-- ============================================================================
-- Backfill site_specs for finetuning.uk from content_data
-- ============================================================================

-- Identity aspect
INSERT INTO site_specs (site_id, aspect, data, source, created_by)
SELECT s.id, 'identity', jsonb_build_object(
        'company_name', COALESCE(s.company_name, 'FineTuning'),
        'tagline', COALESCE(s.tagline, 'AI for the Rest of Us'),
        'email', COALESCE(s.email, ''),
        'phone', COALESCE(s.phone, ''),
        'tone', COALESCE(s.content_data->'response'->>'tone', 'conversational, direct, anti-corporate'),
        'target_audience', COALESCE(s.content_data->'response'->>'target_audience', 'UK SMEs looking for practical AI solutions'),
        'industry', 'AI consulting / technology services'
                         ), 'backfill', 'migration'
FROM sites s
WHERE s.id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND NOT EXISTS (
    SELECT 1 FROM site_specs ss WHERE ss.site_id = s.id AND ss.aspect = 'identity' AND ss.is_current = true
);

-- Design intent aspect
INSERT INTO site_specs (site_id, aspect, data, source, created_by)
SELECT s.id, 'design_intent', jsonb_build_object(
        'style_direction', 'professional-dark',
        'colour_mood', 'dark navy with blue accents — tech, trust, sophistication',
        'typography_mood', 'serif headings (Merriweather) with clean body — authoritative but approachable',
        'layout_preference', 'spacious sections, clear hierarchy, prominent CTAs'
                              ), 'backfill', 'migration'
FROM sites s
WHERE s.id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND NOT EXISTS (
    SELECT 1 FROM site_specs ss WHERE ss.site_id = s.id AND ss.aspect = 'design_intent' AND ss.is_current = true
);

-- Content direction aspect
INSERT INTO site_specs (site_id, aspect, data, source, created_by)
VALUES (
           '1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'content_direction',
           '{"voice": "Experienced practitioners who cut through AI hype — no buzzwords, no nonsense", "emphasis": "practical results, honest advice, UK SME focus", "avoid_phrases": ["synergy", "cutting-edge", "world-class", "leverage", "disrupt"], "social_proof_style": "company commitments and philosophy, not fabricated testimonials"}'::jsonb,
           'backfill', 'migration'
       );

-- ============================================================================
-- Backfill site_specs for gaswholesalers.com
-- ============================================================================

INSERT INTO site_specs (site_id, aspect, data, source, created_by)
SELECT s.id, 'identity', jsonb_build_object(
        'company_name', COALESCE(s.company_name, 'Gas Wholesalers'),
        'tagline', COALESCE(s.tagline, 'Wholesale Gas Supply Solutions'),
        'email', COALESCE(s.email, ''),
        'phone', COALESCE(s.phone, ''),
        'tone', 'professional, direct, trustworthy',
        'target_audience', 'commercial fuel buyers, fleet managers, facility managers, energy procurement managers',
        'industry', 'energy / wholesale fuel distribution'
                         ), 'backfill', 'migration'
FROM sites s
WHERE s.id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND NOT EXISTS (
    SELECT 1 FROM site_specs ss WHERE ss.site_id = s.id AND ss.aspect = 'identity' AND ss.is_current = true
);

INSERT INTO site_specs (site_id, aspect, data, source, created_by)
SELECT s.id, 'design_intent', jsonb_build_object(
        'style_direction', 'professional-dark',
        'colour_mood', 'dark blue and orange — energy, petroleum, industrial strength',
        'typography_mood', 'clean sans-serif — modern corporate, not stuffy',
        'imagery_direction', 'industrial facilities, fuel infrastructure, fleet vehicles',
        'layout_preference', 'spacious sections, clear hierarchy, prominent CTAs'
                              ), 'backfill', 'migration'
FROM sites s
WHERE s.id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND NOT EXISTS (
    SELECT 1 FROM site_specs ss WHERE ss.site_id = s.id AND ss.aspect = 'design_intent' AND ss.is_current = true
);

INSERT INTO site_specs (site_id, aspect, data, source, created_by)
VALUES (
           '5fe15466-4e2e-4ff2-981e-98c1b7074002', 'content_direction',
           '{"voice": "Experienced industry insiders, not salespeople", "emphasis": "reliability, transparency, no-nonsense service", "avoid_phrases": ["synergy", "cutting-edge", "world-class", "solutions provider"], "social_proof_style": "company commitments rather than fabricated testimonials"}'::jsonb,
           'backfill', 'migration'
       );

-- Strategy aspect for gaswholesalers (from the content strategy framework)
INSERT INTO site_specs (site_id, aspect, data, source, created_by)
VALUES (
           '5fe15466-4e2e-4ff2-981e-98c1b7074002', 'strategy',
           '{"visitor_type": "B2B buyers — facility managers, business owners, energy procurement managers", "primary_intent": "find and compare wholesale gas suppliers, understand pricing, negotiate better contracts", "satisfaction_condition": "identified 2-3 suppliers to contact, understand current market pricing, know what to look for in a contract", "monetisation": "lead generation — quote request forms to energy brokers, lead value £10-60 depending on qualification", "recurring_value": "updated pricing page with wholesale market rates, supplier directory", "trust_threshold": "high — commercial energy contracts are high-value B2B decisions"}'::jsonb,
           'backfill', 'migration'
       );

-- Verify
SELECT site_id, aspect, source, LEFT(data::text, 80) as preview
FROM site_specs
WHERE site_id IN ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', '5fe15466-4e2e-4ff2-981e-98c1b7074002')
  AND is_current = true
ORDER BY site_id, aspect;