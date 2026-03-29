-- 026_stage2_deploy.sql
--
-- Run after Stage 1 chassis is deployed.
-- Based on actual schema inspection of gaswholesalers.com site.

-- =========================================================================
-- 1. CSS snippet for latest-news component
-- =========================================================================
-- css_snippets table exists (19 rows). Adding the news grid CSS.

INSERT INTO css_snippets (name, function, category, semantic_tags, css_content, description) VALUES
    (
        'Latest News Grid',
        'latest-news',
        'component',
        ARRAY['news', 'feed', 'dynamic', 'cards', 'grid'],
        '/* Latest News Section */
    .latest-news-section {
        padding: var(--section-padding, 4rem 0);
    }

    .latest-news-section .section-heading {
        margin-bottom: 0.5rem;
    }

    .latest-news-section .section-subheadline {
        color: var(--color-text-muted, #718096);
        margin-bottom: 2rem;
    }

    .news-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
        gap: 1.5rem;
    }

    .news-card {
        background: var(--color-surface, #f7fafc);
        border: 1px solid var(--color-border, #e2e8f0);
        border-radius: 8px;
        padding: 1.25rem;
        transition: transform 0.2s ease, box-shadow 0.2s ease;
    }

    .news-card:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    }

    .news-card-title {
        font-size: 1.05rem;
        font-weight: 600;
        line-height: 1.4;
        margin: 0 0 0.5rem;
    }

    .news-card-title a {
        color: var(--color-text, #2d3748);
        text-decoration: none;
    }

    .news-card-title a:hover {
        color: var(--color-primary, #3182ce);
    }

    .news-card-summary {
        font-size: 0.9rem;
        color: var(--color-text-muted, #718096);
        line-height: 1.5;
        margin: 0 0 0.75rem;
    }

    .news-card-meta {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-size: 0.8rem;
        color: var(--color-text-muted, #718096);
    }

    .news-source {
        font-weight: 500;
    }

    .news-date {
        white-space: nowrap;
    }

    .news-empty {
        text-align: center;
        color: var(--color-text-muted, #718096);
        padding: 2rem;
        font-style: italic;
    }

    .news-section-footer {
        text-align: center;
        margin-top: 2rem;
    }

    .news-more-link {
        color: var(--color-primary, #3182ce);
        text-decoration: none;
        font-weight: 500;
        font-size: 1rem;
    }

    .news-more-link:hover {
        text-decoration: underline;
    }

    @media (max-width: 640px) {
        .news-grid {
            grid-template-columns: 1fr;
            gap: 1rem;
        }
        .news-card {
            padding: 1rem;
        }
    }

    @media (min-width: 641px) and (max-width: 1024px) {
        .news-grid {
            grid-template-columns: repeat(2, 1fr);
        }
    }',
        'Responsive card grid for latest-news. CSS variables from site theme. 3 col desktop, 2 tablet, 1 mobile.'
    ) ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
    description = EXCLUDED.description,
    semantic_tags = EXCLUDED.semantic_tags,
    updated_at = NOW();


-- =========================================================================
-- 2. Enrich classification spec with news_feed recommendation
-- =========================================================================
-- Current classification for gaswholesalers has: site_type, confidence,
-- suggested_style, tone_suggestion, detected_signals, etc.
-- No content_features yet. Deep merge adds it.

-- Supersede current classification
UPDATE site_specs SET is_current = false, superseded_at = NOW()
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'classification' AND is_current = true;

-- Insert enriched version with news_feed
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current)
SELECT
    '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'classification',
    data || '{
        "content_features": {
            "news_feed": {
                "recommended": true,
                "reason": "Gas wholesale markets change daily — price, supply, and regulatory news adds SEO freshness and return visits",
                "vertical_keywords": ["wholesale gas prices", "natural gas market", "gas supply", "LNG", "energy regulation", "oil prices"],
                "source_types": ["rss", "news_search", "api_news"]
            }
        }
    }'::jsonb,
    'enrichment',
    'evaluate_news_feed',
    true
FROM site_specs
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'classification' AND is_current = false
  AND superseded_at IS NOT NULL
ORDER BY superseded_at DESC
    LIMIT 1;

-- Verify
SELECT data->'content_features'->'news_feed'->>'recommended' as news_recommended,
    data->'content_features'->'news_feed'->>'reason' as reason
FROM site_specs
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'classification' AND is_current = true;


-- =========================================================================
-- 3. Add latest-news to the site_plan sections
-- =========================================================================
-- Current homepage sections: hero, features, services-grid,
-- differentiators-section, social_proof, call_to_action
-- Insert latest-news before call_to_action.

-- Supersede current site_plan
UPDATE site_specs SET is_current = false, superseded_at = NOW()
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'site_plan' AND is_current = true;

-- Insert updated site_plan with latest-news in homepage sections
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current)
SELECT
    '5fe15466-4e2e-4ff2-981e-98c1b7074002',
    'site_plan',
    -- Replace the index page sections array to include latest-news
    jsonb_set(
            data,
            '{pages,0,sections}',
            '["hero", "features", "services-grid", "differentiators-section", "social_proof", "latest-news", "call_to_action"]'::jsonb
    ),
    'enrichment',
    'news-section-addition',
    true
FROM site_specs
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'site_plan' AND is_current = false
  AND superseded_at IS NOT NULL
ORDER BY superseded_at DESC
    LIMIT 1;

-- Verify homepage now has latest-news
SELECT data->'pages'->0->>'name' as page_name,
    data->'pages'->0->'sections' as sections
FROM site_specs
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'site_plan' AND is_current = true;


-- =========================================================================
-- 4. Add latest-news page_component to homepage
-- =========================================================================
-- Current homepage has positions 1-6 (hero through call-to-action).
-- Shift call-to-action from position 6 to 7, insert latest-news at 6.

-- Get the latest-news component ID
-- (created by 026c_latest_news_component.sql in stage 1)

-- Shift call-to-action down
UPDATE page_components
SET position = 7, updated_at = NOW()
WHERE page_id = (
    SELECT p.id FROM pages p
    WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
      AND p.name = 'index'
)
  AND position = 6;

-- Insert latest-news at position 6
INSERT INTO page_components (
    page_id, component_id, position, slot_name,
    rendered_html, content_data, build_status
)
SELECT
    p.id,
    cc.id,
    6,
    'latest-news',
    -- Server-rendered shell with JS fetch — headline from content_data
    cc.html_template,
    '{"headline": "Energy Market News", "subheadline": "Latest developments in wholesale gas and energy markets"}'::jsonb,
    'deployed'
FROM pages p
         CROSS JOIN content_components cc
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'index'
  AND cc.function = 'latest-news'
  AND cc.is_active = true;

-- Verify new section order
SELECT pc.position, pc.slot_name, cc.function, cc.display_name,
       length(pc.rendered_html) as html_length,
       pc.build_status
FROM page_components pc
         LEFT JOIN content_components cc ON cc.id = pc.component_id
         JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'index'
ORDER BY pc.position;


-- =========================================================================
-- 5. Enable news feed discovery checks
-- =========================================================================

-- Check current checks list
SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as current_checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Add news feed checks (preserves existing)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (
            SELECT jsonb_agg(DISTINCT val)
            FROM (
                     SELECT val
                     FROM jsonb_array_elements(
                                  default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
                          ) val
                     UNION
                     SELECT val FROM jsonb_array_elements(
                                             '["missing_news_sources", "missing_news_section", "stale_news_section", "all_sources_erroring"]'::jsonb
                                     ) val
                 ) combined
        )
                     ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL
    RETURNING type,
    default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as updated_checks;


-- =========================================================================
-- 6. Planner prompt update
-- =========================================================================
-- The build-site-planner's plan_site step has the LLM prompt with RULES 1-11.
-- Classification data flows via {{.site_specs.specs.classification}} so the
-- LLM already sees content_features.news_feed when present.
-- Append rule 12 for news sections.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
                (default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template')
                    || E'\n12. If the classification data includes content_features.news_feed.recommended = true, add "latest-news" to the homepage (index) sections array. Place it after the main content sections and before "call-to-action". The latest-news component is data-driven — its items come from a news feed database, not the content writer. The content writer will only generate a headline for this section. Do NOT include latest-news if news_feed.recommended is absent or false.'
        )
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- Verify rule 12 was appended
SELECT substring(
               default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
    length(default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template') - 200
) as prompt_tail
FROM agent_definitions
WHERE type = 'build-site-planner' AND deleted_at IS NULL;


-- =========================================================================
-- 7. Final verification
-- =========================================================================

-- CSS snippet
SELECT name, function FROM css_snippets WHERE function = 'latest-news';

-- Component template
SELECT function, display_name FROM content_components WHERE function = 'latest-news';

-- Classification has news_feed
SELECT data->'content_features'->'news_feed'->>'recommended' as news_rec
FROM site_specs
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'classification' AND is_current = true;

-- site_plan has latest-news in homepage
SELECT data->'pages'->0->'sections' as index_sections
FROM site_specs
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND aspect = 'site_plan' AND is_current = true;

-- Homepage has 7 sections (was 6)
SELECT pc.position, cc.function
FROM page_components pc
         LEFT JOIN content_components cc ON cc.id = pc.component_id
         JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'index'
ORDER BY pc.position;

-- Discovery checks enabled
SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Feed items ready for triage
SELECT status, COUNT(*) FROM content_feed_items
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
GROUP BY status;

-- Orchestrator has render + commit steps
SELECT
    default_config->'workflow'->'steps'->'render_news_json' IS NOT NULL as has_render,
    default_config->'workflow'->'steps'->'commit_news'->'action' as commit_action
FROM agent_definitions WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;