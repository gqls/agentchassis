-- ============================================================================
-- Migration 031: Site Plan as Declarative Artefact (Phase 1, doc 030)
-- ============================================================================
-- Introduces the site-plan domain as a separate concern from site_specs.
--
--   site_plans         — one current plan per site, with versioned history
--   site_plan_pages    — row-per-page declaration; canonical name + role + url
--   site_plan_partials — design_direction, content_strategy, lazy page briefs
--
-- Plus two nullable column adds for drift detection and reconciler scheduling:
--
--   pages.built_from_plan_version   — links a built page to the plan that
--                                     produced it; reconciler flags pages
--                                     whose value lags the current plan.id
--   sites.last_reconciled_at        — scheduled reconciler skips sites whose
--                                     value is fresher than the tick interval
--
-- Why a separate schema and not site_specs aspects: doc 030, Q1. Anticipated
-- scale (1000+ pages, 10k+ products) makes a row-per-page table the right
-- shape; mixing operational plan rows into the strategic site_specs table
-- forces every reader to filter and obscures the conceptual boundary.
--
-- Versioning pattern mirrors site_specs: per-key UNIQUE index gated on
-- is_current = true, history rows kept with superseded_at set.
--
-- All additions are nullable / new-table; no backfill required. Safe to
-- apply in production with no agent downtime.
-- ============================================================================


-- ============================================================================
-- 1. site_plans — one current plan per site
-- ============================================================================
-- Each plan-builder run writes a new row, marks the previous one superseded.
-- The plan rows themselves carry no page content — that's site_plan_pages.
-- This table is the version anchor that pages.built_from_plan_version points
-- at and that reconciler reads to decide what work items to emit.

CREATE TABLE IF NOT EXISTS site_plans (
                                          id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID        NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    is_current      BOOLEAN     NOT NULL DEFAULT true,
    source_agent    TEXT,                                     -- e.g. 'site-planner'
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    superseded_at   TIMESTAMPTZ,
    created_by      TEXT        NOT NULL DEFAULT 'system',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- One current plan per site; older plans kept for audit / drift inspection.
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_plans_current
    ON site_plans (site_id) WHERE is_current = true;

-- History lookup ordered by recency for a given site.
CREATE INDEX IF NOT EXISTS idx_site_plans_history
    ON site_plans (site_id, created_at DESC);

CREATE TRIGGER trg_site_plans_updated_at
    BEFORE UPDATE ON site_plans
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ============================================================================
-- 2. site_plan_pages — row-per-page declaration of the planned site
-- ============================================================================
-- One row per page in the plan. Canonical name (post-canonicalisation) and
-- canonical url (post-helper, post-role-validator) are committed here at
-- plan-write time. The reconciler diffs this set against pages and emits
-- needs_page:<name> work items for the delta.
--
-- parent_section is the structural signal the role validator uses to detect
-- section indexes regardless of LLM-supplied role. If 'tools' appears as a
-- name AND any other row in the same plan has parent_section = 'tools',
-- then 'tools' is a section_index, not a content page.

CREATE TABLE IF NOT EXISTS site_plan_pages (
                                               id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id         UUID        NOT NULL REFERENCES site_plans(id) ON DELETE CASCADE,
    name            TEXT        NOT NULL,                     -- canonical name
    role            TEXT        NOT NULL,                     -- 'tool' | 'guide' | 'section_index' | 'content' | etc.
    slug            TEXT        NOT NULL,                     -- raw stem before canonicalisation
    url             TEXT        NOT NULL,                     -- canonical url from URL helper
    parent_section  TEXT,                                     -- nullable; structural parent
    in_header       BOOLEAN     NOT NULL DEFAULT true,
    in_footer       BOOLEAN     NOT NULL DEFAULT true,
    nav_order       INTEGER,
    page_data       JSONB       NOT NULL DEFAULT '{}'::jsonb, -- title, sections, meta_description, etc.
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Each canonical name appears once per plan. Phase 0's site_work_items
-- dedup index has the same shape; the two enforce consistency at different
-- layers.
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_plan_pages_name
    ON site_plan_pages (plan_id, name);

-- Reconciler reads pages for a plan in role order; this index supports that.
CREATE INDEX IF NOT EXISTS idx_site_plan_pages_role
    ON site_plan_pages (plan_id, role);

-- parent_section lookups for the role validator and for nav rendering.
CREATE INDEX IF NOT EXISTS idx_site_plan_pages_parent_section
    ON site_plan_pages (plan_id, parent_section)
    WHERE parent_section IS NOT NULL;


-- ============================================================================
-- 3. site_plan_partials — design direction, content strategy, page briefs
-- ============================================================================
-- Eager partials (design_direction, content_strategy) are written by
-- plan-builder during the three-call cascade. Lazy partials (page_brief:<name>)
-- are written by build_page_brief on demand from page-build-handler.
--
-- partial_type is a free-form text key. Conventions:
--   'design_direction'      — eager, one per plan
--   'content_strategy'      — eager, one per plan
--   'page_brief:<name>'     — lazy, one per plan per page that has been built

CREATE TABLE IF NOT EXISTS site_plan_partials (
                                                  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id         UUID        NOT NULL REFERENCES site_plans(id) ON DELETE CASCADE,
    partial_type    TEXT        NOT NULL,
    data            JSONB       NOT NULL,
    is_current      BOOLEAN     NOT NULL DEFAULT true,
    source_agent    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    superseded_at   TIMESTAMPTZ,
    created_by      TEXT        NOT NULL DEFAULT 'system',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- One current partial per (plan, partial_type); older versions kept.
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_plan_partials_current
    ON site_plan_partials (plan_id, partial_type) WHERE is_current = true;

-- History lookup for a given partial.
CREATE INDEX IF NOT EXISTS idx_site_plan_partials_history
    ON site_plan_partials (plan_id, partial_type, created_at DESC);

CREATE TRIGGER trg_site_plan_partials_updated_at
    BEFORE UPDATE ON site_plan_partials
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ============================================================================
-- 4. pages.built_from_plan_version — drift detection
-- ============================================================================
-- Set by the page-build-handler when a page is built. Reconciler uses this
-- to find pages whose value != the site's current plan id and emit rebuild
-- work items.
--
-- Nullable: existing pages (built before Phase 1) have NULL here. Reconciler
-- treats NULL as "needs first build under new plan."
--
-- No FK to site_plans because the referenced plan may eventually be hard-
-- deleted (audit retention policy is out of scope here). The reconciler
-- handles missing-plan-id gracefully by treating it as drift.

ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS built_from_plan_version UUID;

CREATE INDEX IF NOT EXISTS idx_pages_plan_version
    ON pages (site_id, built_from_plan_version);


-- ============================================================================
-- 5. sites.last_reconciled_at — scheduled reconciler tick
-- ============================================================================
-- Updated by reconcile_site_plan when it finishes. Scheduled tick walks
-- sites where last_reconciled_at < (now - interval) to avoid re-scanning
-- recently-reconciled sites.
--
-- Nullable: pre-existing sites have NULL, which the scheduled tick treats
-- as "never reconciled" — first tick picks them up.

ALTER TABLE sites
    ADD COLUMN IF NOT EXISTS last_reconciled_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_sites_reconcile_due
    ON sites (last_reconciled_at NULLS FIRST)
    WHERE status = 'active';