-- DRAFT DDL — experience_patterns + site_experiences (experience register, P2)
-- STATUS: DESIGN ARTIFACT ONLY. Not a migration. Do NOT apply.
-- The real migration ships in the Phase-2 council-gate change-set together with
-- the subject_type addition (see design/subject_type_addition.md) and the first
-- consumer wiring (PLAN §3.7.4 — no dead stock).

-- ============================================================================
-- experience_patterns — the register: base experience patterns, site-agnostic.
-- Mirrors the tool precedent: this row is the selectable library entry;
-- a travelling doc (subject_type='experience-pattern', subject_key=name)
-- carries provenance/direction/notes/council trail.
-- ============================================================================
CREATE TABLE experience_patterns (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
        -- kebab-case. Immutable once status='approved' (recommended rule:
        -- supersede under a new name rather than rename — RekeyTravellingDocs
        -- is wired for tool renames only; immutability sidesteps the class).
    kind text NOT NULL CHECK (kind IN ('component-contract','micro-journey')),
    display_name text NOT NULL,
    aka jsonb NOT NULL DEFAULT '[]'::jsonb,
        -- industry playbook names, e.g. ["master-detail","drill-down"]
    description text,

    -- Selection axes (columns so the component-selector scoring shape
    -- transfers; all GIN-indexed below)
    primitives jsonb NOT NULL DEFAULT '[]'::jsonb,
        -- level-1 verbs, e.g. ["reveal","navigate"] (design/taxonomy_seed.md)
    section_types jsonb NOT NULL DEFAULT '[]'::jsonb,
        -- component families involved, e.g. ["hero-card-carousel"]
    destination_roles jsonb NOT NULL DEFAULT '[]'::jsonb,
        -- e.g. [{"role":"entity-page","of":"card.entity"},{"role":"tool","of":"topic"}]
        -- roles reuse the existing page_type/role vocabulary; NEVER concrete URLs
    funnel_stage text CHECK (funnel_stage IN ('awareness','consideration','conversion')),
        -- vocabulary adopted from the superseded site_flows/flow_pages schema
    suitable_site_types jsonb NOT NULL DEFAULT '[]'::jsonb,

    -- The contract: every trigger names an OBSERVABLE outcome (gauntlet
    -- hollow-placeholder rule); navigating triggers name a destination role.
    contract jsonb NOT NULL,
        -- [{"control_role":"card-image","primitive":"navigate",
        --   "outcome":"detail page for the card's entity loads",
        --   "destination_role":{"role":"entity-page","of":"card.entity"}},
        --  {"control_role":"read-more","primitive":"reveal",
        --   "outcome":"summary expands in place; control label toggles"}]

    -- Acceptance side (design/criteria_template_schema_v1.md):
    binding_schema jsonb NOT NULL DEFAULT '{}'::jsonb,
        -- declared placeholders a fork must bind, e.g.
        -- {"card_selector":{"type":"selector"},"detail_page":{"type":"page","role":"entity-page"}}
    criteria_template jsonb,
        -- v1 template; write-time validated: parses AND placeholder-closure
        -- against binding_schema (no undeclared placeholders, no unused
        -- declared ones without a waiver)

    status text NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','approved','proven')),
        -- draft = visible, UNSELECTABLE; approved = pattern council pass;
        -- proven = first live green run of bound criteria (evidence upgrade,
        -- set by the checker, never by review)

    -- Provenance
    source text,               -- 'harvest' | 'authored' | migration id
    source_agent text,
    harvested_from text,       -- e.g. 'vonc.com provocations journey'
    created_by text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_exp_patterns_kind ON experience_patterns (kind) WHERE status <> 'draft';
CREATE INDEX idx_exp_patterns_section_types ON experience_patterns USING gin (section_types);
CREATE INDEX idx_exp_patterns_dest_roles ON experience_patterns USING gin (destination_roles);
CREATE INDEX idx_exp_patterns_site_types ON experience_patterns USING gin (suitable_site_types);

-- ============================================================================
-- site_experiences — per-site forks/bindings of register entries.
-- The base entry is invariant; ALL concrete values live here (bug-045 lesson:
-- static fallbacks that re-apply per render cannot be overridden — bindings
-- are data).
-- ============================================================================
CREATE TABLE site_experiences (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id uuid NOT NULL,
    pattern_name text NOT NULL REFERENCES experience_patterns(name),
    instance_key text NOT NULL DEFAULT 'default',
        -- a site can host the same pattern twice (carousel on home AND on
        -- blog); (site_id, pattern_name, instance_key) is the identity
    bindings jsonb NOT NULL DEFAULT '{}'::jsonb,
        -- placeholder -> concrete value; pages bound by page id + url snapshot,
        -- e.g. {"detail_page":{"page_id":"…","url":"/provocations/x/"},
        --       "card_selector":".carousel-card"}
    status text NOT NULL DEFAULT 'proposed'
        CHECK (status IN ('proposed','bound','verified','broken')),
        -- proposed = pattern selected (experience_brief written), roles not
        --   yet all resolved; bound = bind-time validation passed; verified =
        --   bound criteria green on the live site; broken = a previously
        --   verified check now fails (detector-set, like build_status)
    built_from_plan_version uuid,
        -- site_plans.id the binding was resolved against, so binding staleness
        -- is detectable exactly the way page staleness is (bug 038/040 shape)
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (site_id, pattern_name, instance_key)
);

CREATE INDEX idx_site_experiences_site ON site_experiences (site_id, status);

-- ----------------------------------------------------------------------------
-- Open P2 decisions (record verdicts in PLAN before the migration is written):
-- 1. FK from bindings' page references: bindings are jsonb so no FK — decide
--    whether a trigger/constraint or the reconcile guard owns dangling-page
--    detection (leaning: reconcile guard, consistent with needs_page flow).
-- 2. usage_count / avg_quality_score columns (content_components parity):
--    defer until a consumer exists — bugs_open/060 showed usage_count dead on
--    content_components; do not add dead columns.
-- 3. Whether site_experiences.status='broken' emits a work item and which
--    handler owns it (candidate: the experience loop's needs_experience_replan
--    once T5.2 routing exists; Tier-2-only until then).
-- ----------------------------------------------------------------------------
