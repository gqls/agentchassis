-- ============================================================================
-- Migration: Phase 2G Step 1 — site_plan_imagery table
-- ============================================================================
--
-- Adds a new sibling table to site_plan_directives for structured imagery
-- requirements at site / page / section scope. Forms the storage layer for
-- the imagery_plan extension to the planner.
--
-- Same scope/scope_ref/locking pattern as site_plan_directives, but with
-- structured columns appropriate for image generation (kind, prompt,
-- style_hints, constraints) rather than free-text directive.
--
-- This is PURE DDL with NO BEHAVIOUR CHANGE. The table is empty until
-- step 2 (write_site_plan extension) and step 3 (planner prompt extension)
-- start writing to it.
--
-- Sequencing:
--   Step 1 ← you are here (additive DDL, dormant)
--   Step 2: write_site_plan Go extension to populate the table
--   Step 3: planner prompt extension to emit the imagery block
--   Step 4: check_unfulfilled_imagery_plan discovery check
--   Step 5: image-build-handler accepts the new spec shape
--
-- Reference: PLAN_imagery_phase_2g.md
-- ============================================================================

BEGIN;

-- ── 1. Create the table ──

CREATE TABLE IF NOT EXISTS site_plan_imagery (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id       uuid NOT NULL REFERENCES site_plans(id) ON DELETE CASCADE,

    -- Scope: where this imagery requirement applies.
    -- site:     applies across the site (logo, brand-hero)
    -- page:     applies to a specific page (hero_about, illustration on /tools)
    -- section:  applies to a specific section instance on a specific page
    scope         text NOT NULL,
    scope_ref     text,

    -- Key: the asset_key.
    -- For site scope: simple name like 'logo' or 'hero_canonical'.
    -- For page scope: simple name like 'hero_about'.
    -- For section scope: short name like 'icon_precision'.
    -- The discovery check computes the namespaced asset_key from
    -- (scope, scope_ref, key) at work-item emission time.
    key           text NOT NULL,

    -- Kind: categorical type of image. Drives downstream generation choices
    -- (cfg_scale, negative_prompt, style_preset — Phase 2H).
    -- product is deliberately excluded — product imagery comes from the
    -- affiliate_products resolver, not the planner.
    kind          text NOT NULL,

    -- The actual image prompt. Required for all rows.
    prompt        text NOT NULL,

    -- Optional structured hints. Cascade with site_plan_directives'
    -- imagery_direction at generation time (additive, not replacing).
    style_hints   jsonb,
    constraints   jsonb,

    -- Preserves LLM declaration order for multi-imagery within the same
    -- (scope, scope_ref).
    ordering      int NOT NULL DEFAULT 0,

    -- Provenance, matching site_plan_directives.
    source        text NOT NULL DEFAULT 'llm',

    -- HITL locking, matching site_plan_directives.
    -- A locked row's prompt survives plan rebuilds via lock-transfer
    -- in write_site_plan.
    locked_at     timestamptz,
    locked_by     text,

    created_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT chk_scope CHECK (
        scope IN ('site', 'page', 'section')
    ),

    -- Enumerated kind. New kinds added via migration. Update this constraint
    -- (and the planner prompt) together.
    CONSTRAINT chk_kind CHECK (
        kind IN ('logo', 'hero', 'illustration', 'icon', 'infographic')
    ),

    -- scope_ref must be consistent with scope.
    -- site:    scope_ref MUST be NULL
    -- page:    scope_ref MUST be a page_name (non-null)
    -- section: scope_ref MUST be '<page_name>:<ordering>' (contains a colon)
    CONSTRAINT chk_scope_ref_consistency CHECK (
        (scope = 'site' AND scope_ref IS NULL)
        OR (scope = 'page' AND scope_ref IS NOT NULL AND scope_ref NOT LIKE '%:%')
        OR (scope = 'section' AND scope_ref IS NOT NULL AND scope_ref LIKE '%:%')
    ),

    -- Source values enumerated for clarity.
    CONSTRAINT chk_source CHECK (
        source IN ('llm', 'classifier', 'manual', 'adoption')
    )
);

-- ── 2. Indexes ──

-- Uniqueness: one imagery row per (plan, scope, scope_ref, key).
-- COALESCE handles the NULL scope_ref case (Postgres treats NULLs as
-- distinct in unique indexes without this; we want a single row per logical
-- target, so coerce NULL to empty string).
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_plan_imagery_unique
    ON site_plan_imagery (plan_id, scope, COALESCE(scope_ref, ''), key);

-- Cascade read pattern: site → page → section.
-- The brief renderer queries by (plan_id, scope, scope_ref) — matches this
-- index for fast lookups.
CREATE INDEX IF NOT EXISTS idx_site_plan_imagery_plan
    ON site_plan_imagery (plan_id, scope, scope_ref);

-- Lock-transfer queries: find all locked rows from previous current plan.
-- Partial index keeps it small.
CREATE INDEX IF NOT EXISTS idx_site_plan_imagery_locks
    ON site_plan_imagery (plan_id)
    WHERE locked_at IS NOT NULL;

-- ── 3. Comments for future readers ──

COMMENT ON TABLE site_plan_imagery IS
    'Structured imagery requirements at site/page/section scope. Sibling to '
    'site_plan_directives. Phase 2G — see PLAN_imagery_phase_2g.md';

COMMENT ON COLUMN site_plan_imagery.scope IS
    'Where this imagery applies: site | page | section. Determines what '
    'scope_ref means.';

COMMENT ON COLUMN site_plan_imagery.scope_ref IS
    'NULL for scope=site. page_name for scope=page. <page_name>:<ordering> '
    'for scope=section. Constraint chk_scope_ref_consistency enforces.';

COMMENT ON COLUMN site_plan_imagery.key IS
    'The asset_key for this image at its scope. The discovery check '
    'namespaces it appropriately (e.g. scope=page + scope_ref=about + '
    'key=illustration → asset_key=page.about.illustration).';

COMMENT ON COLUMN site_plan_imagery.kind IS
    'Categorical image kind. Drives generation behaviour in Phase 2H. '
    'Enum: logo, hero, illustration, icon, infographic. Product images '
    'come from affiliate_products resolver and are NOT in this enum.';

COMMENT ON COLUMN site_plan_imagery.style_hints IS
    'Optional JSONB. Cascades ADDITIVELY with site_plan_directives '
    'imagery_direction at generation time. Example: '
    '{"medium": "line drawing", "mood": "warm"}.';

COMMENT ON COLUMN site_plan_imagery.constraints IS
    'Optional JSONB. Generation constraints (aspect ratio, transparency). '
    'Example: {"aspect": "1:1", "transparent_background": true}.';

COMMENT ON COLUMN site_plan_imagery.locked_at IS
    'HITL lock. Mirrors site_plan_directives. Locked rows survive plan '
    'rebuild via lock-transfer in write_site_plan.';

-- ── 4. Sanity check ──
-- The migration assumes site_plans exists (Phase 1). Fail loudly if not.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'site_plans' AND table_schema = 'public'
    ) THEN
        RAISE EXCEPTION
            'site_plans table does not exist; Phase 1 plan-domain migration '
            'must run first. Phase 2G depends on site_plans being present.';
    END IF;
END $$;

COMMIT;

-- ============================================================================
-- Rollback (if needed):
--
--   DROP TABLE site_plan_imagery;
--
-- No data preserved (additive DDL, table was empty after migration).
-- ============================================================================

-- Verification queries (run manually after deploy):
--
--   -- Table exists
--   \d site_plan_imagery
--
--   -- Indexes present
--   SELECT indexname FROM pg_indexes
--   WHERE tablename = 'site_plan_imagery'
--   ORDER BY indexname;
--
--   -- Constraints active
--   SELECT conname, contype FROM pg_constraint
--   WHERE conrelid = 'site_plan_imagery'::regclass
--   ORDER BY contype, conname;
--
--   -- Table is empty (expected — populated by write_site_plan from step 2 on)
--   SELECT COUNT(*) FROM site_plan_imagery;
