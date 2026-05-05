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


---


-- ============================================================================
-- Migration 031: Site Plan as Declarative Artefact (Phase 1, doc 030)
-- ============================================================================
-- Introduces the site-plan domain as a separate concern from site_specs.
--
-- Four plan-domain tables, all row-shaped for scale:
--
--   site_plans            — version anchor; one current row per site
--   site_plan_pages       — per-page structural row
--   site_plan_sections    — per-section structural row, carrying resolved
--                           component / palette / layout IDs for HTML
--                           provenance attributes downstream
--   site_plan_directives  — per-directive row at site / page / section
--                           scope; the storage form of design / content /
--                           voice / structural guidance. HITL locks live
--                           on the row via locked_at / locked_by.
--
-- Plus two nullable columns on existing tables for drift detection and
-- reconciler scheduling:
--
--   pages.built_from_plan_version    — links built page to the plan that
--                                      produced it. Reconciler flags pages
--                                      whose value lags the current plan.id.
--   sites.last_reconciled_at         — scheduled reconciler skips sites
--                                      whose value is fresher than the tick
--                                      interval.
--
-- Why a separate schema and not site_specs aspects: doc 030, Q1.
-- Anticipated scale (1000+ pages, 10k+ products) makes a row-per-thing
-- table the right shape; conflating operational plan rows into the
-- strategic site_specs table forces every reader to filter by aspect
-- prefix and obscures the conceptual boundary.
--
-- Why row-per-directive and not a JSONB blob per partial: doc 030 (after
-- discussion). At anticipated scale the plan carries thousands of
-- directives across site / page / section scopes. JSONB blobs force every
-- read to load and traverse the whole thing; per-row storage allows
-- surgical queries (give me design directives for this page), surgical
-- HITL edits, and lock transfer at a meaningful granularity.
--
-- Versioning pattern mirrors site_specs: per-key UNIQUE index gated on
-- is_current = true, history rows kept with superseded_at set.
--
-- Lock transfer: HITL-locked directives have locked_at IS NOT NULL.
-- The write_site_plan action, when writing a fresh plan, looks up the
-- previous current plan's locked directives by stable composite key
-- (scope, scope_ref, category, subject) and transfers the lock onto the
-- equivalent new directive row. No reader other than write_site_plan
-- needs to understand this; every consumer reads
-- site_plan_directives.locked_at uniformly.
--
-- All additions are nullable / new-table; no backfill required. Safe to
-- apply in production with no agent downtime.
-- ============================================================================


-- ============================================================================
-- 1. site_plans — version anchor (one current plan per site)
-- ============================================================================

CREATE TABLE IF NOT EXISTS site_plans (
                                          id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID        NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    is_current      BOOLEAN     NOT NULL DEFAULT true,
    source_agent    TEXT,                                     -- e.g. 'build-site-planner'
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    superseded_at   TIMESTAMPTZ,
    created_by      TEXT        NOT NULL DEFAULT 'system',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE UNIQUE INDEX IF NOT EXISTS idx_site_plans_current
    ON site_plans (site_id) WHERE is_current = true;

CREATE INDEX IF NOT EXISTS idx_site_plans_history
    ON site_plans (site_id, created_at DESC);

CREATE TRIGGER trg_site_plans_updated_at
    BEFORE UPDATE ON site_plans
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ============================================================================
-- 2. site_plan_pages — row per planned page
-- ============================================================================
-- Canonical name (post-canonicalisation) and canonical url (post-helper)
-- are committed here at plan-write time. The reconciler diffs this set
-- against the pages table and emits needs_page:<name> work items for the
-- delta.
--
-- parent_section is the structural signal the role validator uses to
-- detect section indexes regardless of LLM-supplied role.

CREATE TABLE IF NOT EXISTS site_plan_pages (
                                               id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id          UUID        NOT NULL REFERENCES site_plans(id) ON DELETE CASCADE,
    name             TEXT        NOT NULL,                    -- canonical name
    role             TEXT        NOT NULL,                    -- 'tool' | 'guide' | 'section_index' | 'content' | etc.
    slug             TEXT        NOT NULL,                    -- raw stem before canonicalisation
    url              TEXT        NOT NULL,                    -- canonical url from CanonicalisePage
    parent_section   TEXT,                                    -- nullable; structural parent for nested pages
    title            TEXT,
    meta_description TEXT,
    nav_label        TEXT,
    in_header        BOOLEAN     NOT NULL DEFAULT true,
    in_footer        BOOLEAN     NOT NULL DEFAULT true,
    nav_order        INTEGER,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Each canonical name appears once per plan. Phase 0's site_work_items
-- dedup index has the same shape; the two enforce consistency at
-- different layers.
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
-- 3. site_plan_sections — row per planned section within a page
-- ============================================================================
-- Naming note: there are existing actions called `plan_sections` and
-- `save_page_sections` in the registry, but they operate on per-section
-- *content data* at render time (the title text, body copy, image refs
-- for an instance of a section). `site_plan_sections` is the structural
-- plan: "page X has a hero section at position 0, a features section at
-- position 1, ..." with no content. The two concepts share a noun and
-- nothing else.
--
-- Stable composite key for lock transfer across plan rebuilds:
--   (site_id via plan_id → site_plans.site_id, page_name, ordering)
--
-- Carries resolved IDs (component_version_id, palette_id, layout_id,
-- typography_set_id) so HTML emission downstream can attach data-* attrs:
--
--   <section data-component="hero"
--            data-component-version-id="..."
--            data-section-id="..."
--            data-palette-id="..."
--            data-plan-id="...">
--
-- These resolved IDs are written nullable here; resolution happens
-- downstream of plan write (likely by site-design-planner or a follow-up
-- step). For Phase 1's first cut write_site_plan leaves them NULL.

CREATE TABLE IF NOT EXISTS site_plan_sections (
                                                  id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id               UUID        NOT NULL REFERENCES site_plans(id) ON DELETE CASCADE,
    page_name             TEXT        NOT NULL,               -- denormalised from site_plan_pages.name
    ordering              INTEGER     NOT NULL,               -- 0-based position in the page's section list
    component_name        TEXT        NOT NULL,               -- 'hero', 'features', etc — for readability + data-component attr

-- Resolved provenance IDs — nullable, filled by downstream resolver
    component_version_id  UUID,                               -- which version of the component template
    palette_id            UUID,                               -- which palette resolved for this section
    layout_id             UUID,                               -- which layout
    typography_set_id     UUID,                               -- which typography set

    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Stable composite for section identity within a plan; the same key shape
-- (page_name, ordering) is what scope_ref encodes for section-scoped
-- directives, and what write_site_plan uses for lock transfer.
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_plan_sections_key
    ON site_plan_sections (plan_id, page_name, ordering);

-- Page-level reads ("give me all sections for this page") are common.
CREATE INDEX IF NOT EXISTS idx_site_plan_sections_page
    ON site_plan_sections (plan_id, page_name);


-- ============================================================================
-- 4. site_plan_directives — row per directive at site / page / section scope
-- ============================================================================
-- The storage form of design / content / voice / structural guidance.
-- Brief-rendering is handled by a Go helper that reads the cascade
-- (site → page → section) for a consumer and emits prompt-ready text;
-- consumers never read directives directly.
--
-- scope_ref encoding:
--   site scope    : NULL
--   page scope    : page name (canonical, e.g. 'tool-ttk-calculator')
--   section scope : '<page_name>:<ordering>' (e.g. 'tool-ttk-calculator:0')
--
-- Cardinality (single-value override vs multi-value accumulation) is not
-- stored on the row; the brief renderer has a small lookup table mapping
-- subjects like 'writing_rules' → multi, 'voice.register' → single.
-- See doc 030 §"Directive cascade and brief assembly".
--
-- Lock transfer: when write_site_plan writes a new plan, it inspects the
-- previous plan's locked directives and copies locked_at / locked_by
-- onto the equivalent new row, matched by
-- (scope, scope_ref, category, subject, ordering) under the same site.

CREATE TABLE IF NOT EXISTS site_plan_directives (
                                                    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id         UUID        NOT NULL REFERENCES site_plans(id) ON DELETE CASCADE,
    scope           TEXT        NOT NULL CHECK (scope IN ('site', 'page', 'section')),
    scope_ref       TEXT,                                     -- NULL for site; page_name; '<page_name>:<ordering>'
    category        TEXT        NOT NULL,                     -- 'design' | 'content' | 'voice' | 'style' | 'structural'
    subject         TEXT        NOT NULL,                     -- 'palette.character', 'voice.register', 'writing_rules', ...
    directive       TEXT        NOT NULL,                     -- the actual guidance text (paragraph or bullet)
    ordering        INTEGER     NOT NULL DEFAULT 0,           -- preserves LLM order for multi-cardinality subjects
    source          TEXT        NOT NULL DEFAULT 'llm',       -- 'llm' | 'classifier' | 'adoption' | 'manual'

-- HITL lock columns (pattern from page_components / site_components)
    locked_at       TIMESTAMPTZ,
    locked_by       TEXT,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source_agent    TEXT,
    created_by      TEXT        NOT NULL DEFAULT 'system'
    );

-- Brief renderer's primary read shape: "give me all directives at this
-- scope for this scope_ref". The cascade walks site (scope_ref NULL),
-- then page (scope_ref = page name), then section (scope_ref = composite).
CREATE INDEX IF NOT EXISTS idx_site_plan_directives_scope
    ON site_plan_directives (plan_id, scope, scope_ref);

-- Tighter filter when the consumer wants only one category (design only,
-- voice only).
CREATE INDEX IF NOT EXISTS idx_site_plan_directives_category
    ON site_plan_directives (plan_id, scope, scope_ref, category);

-- Lock-transfer query on plan rebuild: write_site_plan looks at the
-- previous plan's locked rows.
CREATE INDEX IF NOT EXISTS idx_site_plan_directives_locked
    ON site_plan_directives (plan_id) WHERE locked_at IS NOT NULL;


-- ============================================================================
-- 5. pages.built_from_plan_version — drift detection
-- ============================================================================
-- Set by the page-build-handler when a page is built. Reconciler uses
-- this to find pages whose value != the site's current plan id and emit
-- rebuild work items.
--
-- Nullable: existing pages (built before Phase 1) have NULL here.
-- Reconciler treats NULL as "needs first build under new plan."
--
-- No FK to site_plans because the referenced plan may eventually be
-- hard-deleted (audit retention policy is out of scope here). The
-- reconciler handles missing-plan-id gracefully by treating it as drift.

ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS built_from_plan_version UUID;

CREATE INDEX IF NOT EXISTS idx_pages_plan_version
    ON pages (site_id, built_from_plan_version);


-- ============================================================================
-- 6. sites.last_reconciled_at — scheduled reconciler tick
-- ============================================================================
-- Updated by reconcile_site_plan when it finishes. Scheduled tick walks
-- sites where last_reconciled_at < (now - interval) to avoid re-scanning
-- recently-reconciled sites.
--
-- Nullable: pre-existing sites have NULL, which the scheduled tick
-- treats as "never reconciled" — first tick picks them up.

ALTER TABLE sites
    ADD COLUMN IF NOT EXISTS last_reconciled_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_sites_reconcile_due
    ON sites (last_reconciled_at NULLS FIRST)
    WHERE status = 'active';