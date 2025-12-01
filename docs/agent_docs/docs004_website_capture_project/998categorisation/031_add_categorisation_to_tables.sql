-- 001_add_categorization.sql
-- Adds category, status, and domain_tags to agent tables for better organization
-- Run this against the clients_db database

BEGIN;

-- ============================================================================
-- PART 1: Schema changes for agent_group_definitions
-- ============================================================================

-- Add category column (what the agent group DOES, domain-agnostic)
-- Categories: builder, analyzer, collector, transformer, evaluator, researcher, workflow, monitor
ALTER TABLE agent_group_definitions
    ADD COLUMN IF NOT EXISTS category TEXT;

-- Add status column
-- Status: active, experimental, deprecated, demo, template
ALTER TABLE agent_group_definitions
    ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'experimental';

-- Add domain_tags for flexible domain tagging
-- Examples: ["website", "marketing"], ["healthcare", "fitness"], ["finance", "valuation"]
ALTER TABLE agent_group_definitions
    ADD COLUMN IF NOT EXISTS domain_tags JSONB DEFAULT '[]'::jsonb;

-- Add constraints
ALTER TABLE agent_group_definitions
    ADD CONSTRAINT check_agd_category
        CHECK (category IS NULL OR category IN (
                                                'builder',      -- Creates/generates artifacts (websites, documents, reports)
                                                'analyzer',     -- Examines data, produces insights (valuation, assessment, code review)
                                                'collector',    -- Gathers data from sources (web scraping, API aggregation, surveys)
                                                'transformer',  -- Converts between formats (data migration, format conversion)
                                                'evaluator',    -- Scores, ranks, or values items (lead scoring, risk assessment)
                                                'researcher',   -- Deep investigation (market research, competitive analysis)
                                                'workflow',     -- Multi-step business processes (approvals, intake, HITL)
                                                'monitor'       -- Watches and alerts (health tracking, system monitoring)
            ));

ALTER TABLE agent_group_definitions
    ADD CONSTRAINT check_agd_status
        CHECK (status IN ('active', 'experimental', 'deprecated', 'demo', 'template'));

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_agd_category ON agent_group_definitions(category);
CREATE INDEX IF NOT EXISTS idx_agd_status ON agent_group_definitions(status);
CREATE INDEX IF NOT EXISTS idx_agd_domain_tags ON agent_group_definitions USING GIN(domain_tags);


-- ============================================================================
-- PART 2: Schema changes for agent_definitions
-- ============================================================================

-- Add category column for individual agents
ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS agent_category TEXT;

-- Add status column
ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'experimental';

-- Add domain_tags
ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS domain_tags JSONB DEFAULT '[]'::jsonb;

-- Add constraints
ALTER TABLE agent_definitions
    ADD CONSTRAINT check_ad_category
        CHECK (agent_category IS NULL OR agent_category IN (
                                                            'strategist',   -- Planning, decision-making
                                                            'executor',     -- Performs tasks, builds things
                                                            'analyst',      -- Examines, evaluates, reports
                                                            'integrator',   -- Connects to external systems
                                                            'coordinator',  -- Orchestrates other agents
                                                            'specialist'    -- Domain-specific expertise
            ));

ALTER TABLE agent_definitions
    ADD CONSTRAINT check_ad_status
        CHECK (status IN ('active', 'experimental', 'deprecated', 'demo', 'template'));

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_ad_category ON agent_definitions(agent_category);
CREATE INDEX IF NOT EXISTS idx_ad_status ON agent_definitions(status);
CREATE INDEX IF NOT EXISTS idx_ad_domain_tags ON agent_definitions USING GIN(domain_tags);


-- ============================================================================
-- PART 3: Categorize existing agent_group_definitions
-- ============================================================================

-- Website builders
UPDATE agent_group_definitions SET
                                   category = 'builder',
                                   status = 'deprecated',
                                   domain_tags = '["website"]'::jsonb
WHERE group_type IN (
    'website-builder',
    'simple-website-builder',
    'multi-section-website-builder',
    'website-builder-with-images'
    );

-- MVP site builder - our current working one
UPDATE agent_group_definitions SET
                                   category = 'builder',
                                   status = 'active',
                                   domain_tags = '["website", "landing-page"]'::jsonb
WHERE group_type = 'mvp-site-builder';

-- Landing page and content builders
UPDATE agent_group_definitions SET
                                   category = 'builder',
                                   status = 'experimental',
                                   domain_tags = '["website", "landing-page"]'::jsonb
WHERE group_type IN (
    'landing-page-builder',
    'content-site-builder',
    'brand-focused-site-builder'
    );

-- Robot hands demos
UPDATE agent_group_definitions SET
                                   category = 'builder',
                                   status = 'demo',
                                   domain_tags = '["website", "demo"]'::jsonb
WHERE group_type IN (
    'robot-hands-website',
    'robot-hands-complete-website'
    );

-- Analyzers
UPDATE agent_group_definitions SET
                                   category = 'analyzer',
                                   status = 'experimental',
                                   domain_tags = '["website", "seo"]'::jsonb
WHERE group_type = 'website-analyzer';

-- Collectors
UPDATE agent_group_definitions SET
                                   category = 'collector',
                                   status = 'experimental',
                                   domain_tags = '["web-scraping"]'::jsonb
WHERE group_type = 'parallel-scraper';

-- Workflows
UPDATE agent_group_definitions SET
                                   category = 'workflow',
                                   status = 'experimental',
                                   domain_tags = '["intake"]'::jsonb
WHERE group_type = 'intake-orchestrator';

UPDATE agent_group_definitions SET
                                   category = 'workflow',
                                   status = 'experimental',
                                   domain_tags = '["approval", "hitl"]'::jsonb
WHERE group_type = 'content-approval-hitl';

-- Generators
UPDATE agent_group_definitions SET
                                   category = 'builder',
                                   status = 'demo',
                                   domain_tags = '["messaging"]'::jsonb
WHERE group_type = 'welcome-message-generator';


-- ============================================================================
-- PART 4: Categorize existing agent_definitions
-- ============================================================================

-- Strategist agents
UPDATE agent_definitions SET
                             agent_category = 'strategist',
                             status = 'active',
                             domain_tags = '["website"]'::jsonb
WHERE type IN ('chief-strategist', 'site-strategist');

-- Architect/executor agents
UPDATE agent_definitions SET
                             agent_category = 'executor',
                             status = 'active',
                             domain_tags = '["website"]'::jsonb
WHERE type IN ('site-component-architect', 'landing-page-architect', 'html-assembler');

-- Content agents
UPDATE agent_definitions SET
                             agent_category = 'executor',
                             status = 'active',
                             domain_tags = '["content", "website"]'::jsonb
WHERE type IN ('content-creator', 'content-writer');

-- Deployer agents
UPDATE agent_definitions SET
                             agent_category = 'integrator',
                             status = 'active',
                             domain_tags = '["deployment", "git"]'::jsonb
WHERE type IN ('deployer-agent', 'site-deployer');

-- Generic/orchestrator
UPDATE agent_definitions SET
                             agent_category = 'coordinator',
                             status = 'active',
                             domain_tags = '[]'::jsonb
WHERE type = 'generic';


-- ============================================================================
-- PART 5: Helper views for querying
-- ============================================================================

-- View: Active workflows by category
CREATE OR REPLACE VIEW v_active_workflows AS
SELECT
    group_type,
    name,
    category,
    domain_tags,
    description,
    usage_count,
    updated_at
FROM agent_group_definitions
WHERE status = 'active'
ORDER BY category, group_type;

-- View: All workflows with status
CREATE OR REPLACE VIEW v_all_workflows AS
SELECT
    group_type,
    name,
    category,
    status,
    domain_tags,
    LEFT(description, 80) as description_preview,
    usage_count,
    created_at::date as created
FROM agent_group_definitions
ORDER BY
    CASE status
    WHEN 'active' THEN 1
    WHEN 'experimental' THEN 2
    WHEN 'demo' THEN 3
    WHEN 'template' THEN 4
    WHEN 'deprecated' THEN 5
END,
    category,
    group_type;

-- View: Agents by category
CREATE OR REPLACE VIEW v_agents_by_category AS
SELECT
    type,
    display_name,
    agent_category,
    status,
    domain_tags,
    LEFT(description, 60) as description_preview
FROM agent_definitions
WHERE deleted_at IS NULL
ORDER BY agent_category, type;

COMMIT;


-- ============================================================================
-- VERIFICATION QUERIES (run these after the migration)
-- ============================================================================

-- Check workflow categorization
-- SELECT * FROM v_all_workflows;

-- Find uncategorized workflows
-- SELECT group_type, name FROM agent_group_definitions WHERE category IS NULL;

-- Find uncategorized agents
-- SELECT type, display_name FROM agent_definitions WHERE agent_category IS NULL AND deleted_at IS NULL;

-- Query by domain
-- SELECT * FROM agent_group_definitions WHERE domain_tags ? 'website' AND status = 'active';

-- Query active analyzers
-- SELECT * FROM agent_group_definitions WHERE category = 'analyzer' AND status = 'active';