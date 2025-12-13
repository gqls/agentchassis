-- Multi-Track Sitemap Architecture Schema
-- Designed to support 1-N flows, but optimized for single-flow sites

-- ============================================================================
-- FLOWS: Define user journeys/narratives within a site
-- ============================================================================
CREATE TABLE site_flows (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Site association
                            orchestration_id UUID NOT NULL, -- links to the site build orchestration
                            domain TEXT NOT NULL,

    -- Flow identity
                            flow_name TEXT NOT NULL, -- e.g. "primary_conversion", "thought_leadership", "partner_track"
                            is_primary BOOLEAN DEFAULT true, -- for single-flow sites, this is always true

    -- Audience definition
                            audience_segment TEXT NOT NULL, -- e.g. "c_suite_executives", "technical_evaluators"
                            audience_description TEXT, -- detailed persona info

    -- Narrative structure
                            narrative_arc JSONB NOT NULL, -- stages with voice/tone parameters per stage
    /*
    Example structure:
    {
        "stage_1": {
            "name": "awareness",
            "objective": "establish_expertise",
            "voice_formality": 0.6,
            "technical_depth": 0.4,
            "sales_pressure": 0.1,
            "pacing": "moderate"
        },
        "stage_2": {
            "name": "consideration",
            "objective": "build_trust",
            "voice_formality": 0.7,
            "technical_depth": 0.6,
            "sales_pressure": 0.3
        },
        "stage_3": {
            "name": "conversion",
            "objective": "drive_action",
            "voice_formality": 0.8,
            "technical_depth": 0.5,
            "sales_pressure": 0.7
        }
    }
    */

    -- Entry and exit
                            entry_points TEXT[], -- where users enter this flow: ["organic_search", "linkedin_ad", "referral"]
                            success_metric TEXT, -- what conversion looks like: "newsletter_signup", "demo_request", "purchase"

    -- Metadata
                            created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                            updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                            version INTEGER DEFAULT 1,
                            is_active BOOLEAN DEFAULT true
);

CREATE INDEX idx_site_flows_orchestration ON site_flows(orchestration_id);
CREATE INDEX idx_site_flows_domain ON site_flows(domain);
CREATE INDEX idx_site_flows_primary ON site_flows(is_primary) WHERE is_primary = true;

-- ============================================================================
-- FLOW PAGES: Pages within a specific flow with stage-specific context
-- ============================================================================
CREATE TABLE flow_pages (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Flow association
                            flow_id UUID NOT NULL REFERENCES site_flows(id) ON DELETE CASCADE,

    -- Page identity
                            page_path TEXT NOT NULL, -- e.g. "index.html", "services/consulting.html"
                            page_title TEXT,

    -- Position in narrative
                            stage_in_narrative TEXT NOT NULL, -- maps to narrative_arc stage keys: "stage_1", "stage_2", etc
                            sequence_order INTEGER NOT NULL, -- order within the flow (1, 2, 3...)

    -- Context overrides (inherits from flow, can override)
                            context_overrides JSONB DEFAULT '{}',
    /*
    Example overrides:
    {
        "voice_formality": 0.9,  // higher than flow baseline
        "technical_depth": 0.8,
        "urgency": 0.6,
        "data_density": 0.7,
        "emotional_appeal": 0.3
    }
    */

    -- Page structure
                            page_archetype TEXT, -- e.g. "long_form_sales", "testimonial_grid", "feature_comparison"
                            components JSONB, -- component tree for this page

    -- Metadata
                            created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                            updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

                            UNIQUE(flow_id, page_path)
);

CREATE INDEX idx_flow_pages_flow ON flow_pages(flow_id);
CREATE INDEX idx_flow_pages_stage ON flow_pages(stage_in_narrative);
CREATE INDEX idx_flow_pages_sequence ON flow_pages(flow_id, sequence_order);

-- ============================================================================
-- PAGE TRANSITIONS: How users move through flows
-- ============================================================================
CREATE TABLE page_transitions (
                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                  from_page_id UUID NOT NULL REFERENCES flow_pages(id) ON DELETE CASCADE,
                                  to_page_id UUID NOT NULL REFERENCES flow_pages(id) ON DELETE CASCADE,

    -- Transition type
                                  transition_type TEXT NOT NULL, -- "next_in_flow", "alternate_path", "cross_flow", "exit_flow"

    -- For A/B testing and optimization
                                  weight DECIMAL DEFAULT 1.0, -- probability of this transition
                                  conversion_rate DECIMAL, -- how often this transition leads to success

    -- Metadata
                                  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

                                  UNIQUE(from_page_id, to_page_id)
);

CREATE INDEX idx_page_transitions_from ON page_transitions(from_page_id);
CREATE INDEX idx_page_transitions_to ON page_transitions(to_page_id);

-- ============================================================================
-- BRAND DNA: Site-level invariants (what never changes)
-- ============================================================================
CREATE TABLE site_brand_dna (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                orchestration_id UUID NOT NULL UNIQUE,
                                domain TEXT NOT NULL,

    -- Visual identity
                                theme_name TEXT NOT NULL, -- links to css_themes
                                color_palette JSONB,
                                typography JSONB,

    -- Core messaging (invariants)
                                core_message TEXT NOT NULL, -- one sentence: "AI orchestration made pragmatic"
                                core_values TEXT[], -- ["transparency", "expertise", "results"]

    -- Voice boundaries
                                voice_parameters JSONB NOT NULL,
    /*
    Example:
    {
        "formality_range": [0.4, 1.0],      // can vary from casual to very formal
        "technical_depth_range": [0.3, 0.9],
        "sales_pressure_range": [0.1, 0.8],
        "forbidden_phrases": ["revolutionary", "cutting-edge"],
        "required_elements": ["data-driven", "proven results"]
    }
    */

    -- Company facts
                                company_info JSONB,

                                created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_brand_dna_orchestration ON site_brand_dna(orchestration_id);
CREATE INDEX idx_brand_dna_domain ON site_brand_dna(domain);

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Get primary flow for a site (for single-flow sites)
CREATE OR REPLACE FUNCTION get_primary_flow(p_orchestration_id UUID)
RETURNS TABLE (
    flow_id UUID,
    flow_name TEXT,
    narrative_arc JSONB
) AS $$
BEGIN
RETURN QUERY
SELECT id, site_flows.flow_name, site_flows.narrative_arc
FROM site_flows
WHERE orchestration_id = p_orchestration_id
  AND is_primary = true
  AND is_active = true
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Get all pages in flow order
CREATE OR REPLACE FUNCTION get_flow_pages_ordered(p_flow_id UUID)
RETURNS TABLE (
    page_id UUID,
    page_path TEXT,
    stage TEXT,
    sequence INTEGER
) AS $$
BEGIN
RETURN QUERY
SELECT id, flow_pages.page_path, stage_in_narrative, sequence_order
FROM flow_pages
WHERE flow_id = p_flow_id
ORDER BY sequence_order ASC;
END;
$$ LANGUAGE plpgsql;

-- Get context for a specific page (merges flow + page overrides)
CREATE OR REPLACE FUNCTION get_page_context(p_page_id UUID)
RETURNS JSONB AS $$
DECLARE
v_flow_context JSONB;
    v_page_overrides JSONB;
    v_stage_context JSONB;
    v_result JSONB;
BEGIN
    -- Get flow narrative and page details
SELECT
    sf.narrative_arc,
    fp.context_overrides,
    sf.narrative_arc->fp.stage_in_narrative
INTO v_flow_context, v_page_overrides, v_stage_context
FROM flow_pages fp
         JOIN site_flows sf ON fp.flow_id = sf.id
WHERE fp.id = p_page_id;

-- Merge: flow stage context + page overrides
-- Page overrides take precedence
v_result := v_stage_context || v_page_overrides;

RETURN v_result;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- MIGRATION HELPER: Create default single-flow structure for existing sites
-- ============================================================================
CREATE OR REPLACE FUNCTION create_default_flow(
    p_orchestration_id UUID,
    p_domain TEXT,
    p_audience TEXT DEFAULT 'primary_audience',
    p_flow_name TEXT DEFAULT 'primary_conversion'
)
RETURNS UUID AS $$
DECLARE
v_flow_id UUID;
    v_default_narrative JSONB;
BEGIN
    -- Default 3-stage narrative (awareness → consideration → conversion)
    v_default_narrative := '{
        "stage_1": {
            "name": "awareness",
            "objective": "capture_attention",
            "voice_formality": 0.6,
            "technical_depth": 0.5,
            "sales_pressure": 0.2,
            "pacing": "engaging"
        },
        "stage_2": {
            "name": "consideration",
            "objective": "build_trust",
            "voice_formality": 0.7,
            "technical_depth": 0.6,
            "sales_pressure": 0.4,
            "pacing": "informative"
        },
        "stage_3": {
            "name": "conversion",
            "objective": "drive_action",
            "voice_formality": 0.7,
            "technical_depth": 0.5,
            "sales_pressure": 0.7,
            "pacing": "urgent"
        }
    }'::JSONB;

INSERT INTO site_flows (
    orchestration_id,
    domain,
    flow_name,
    is_primary,
    audience_segment,
    narrative_arc,
    entry_points,
    success_metric
) VALUES (
             p_orchestration_id,
             p_domain,
             p_flow_name,
             true,
             p_audience,
             v_default_narrative,
             ARRAY['organic_search', 'direct_traffic'],
             'primary_cta_completion'
         )
    RETURNING id INTO v_flow_id;

RETURN v_flow_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE site_flows IS 'User journeys/narratives within a site. Single-flow sites have one primary flow.';
COMMENT ON TABLE flow_pages IS 'Pages within a flow with stage-specific context and overrides.';
COMMENT ON TABLE page_transitions IS 'Tracks how users move between pages for optimization.';
COMMENT ON TABLE site_brand_dna IS 'Site-level invariants that never change across flows.';
COMMENT ON COLUMN site_flows.is_primary IS 'For single-flow sites, always true. For multi-flow, one must be primary.';
COMMENT ON COLUMN site_flows.narrative_arc IS 'Stages with voice/tone parameters. Keys are stage_1, stage_2, etc.';
COMMENT ON COLUMN flow_pages.context_overrides IS 'Page-specific overrides that take precedence over flow baseline.';
COMMENT ON COLUMN site_brand_dna.voice_parameters IS 'Defines allowed variance ranges for voice parameters.';