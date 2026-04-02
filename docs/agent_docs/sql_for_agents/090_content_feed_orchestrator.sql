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

--

-- ============================================================================
-- 027i — Content Feed Orchestrator: add source seeding + triage
-- ============================================================================
--
-- What this does:
--   1. Backs up the current content-feed-orchestrator definition
--   2. Updates the workflow to add:
--      - seed_sources step (first step — creates content_sources from spec)
--      - check_has_sources step (skip if no sources after seeding)
--      - run_triage step (call feed-triage agent to score ingested items)
--      - check_has_news step (skip git commit if nothing rendered)
--   3. Increases timeout to 900s (triage LLM call can take time)
--
-- New flow:
--   seed_sources → check_has_sources → dispatch_sources → run_triage
--     → render_news_json → check_has_news → commit_news → complete
--
-- Timing note: dispatch_sources fires off async feed-ingesters. They write
-- items with status='ingested' for NEXT run's triage. The run_triage step
-- scores items from PREVIOUS runs. render_news_json reads both 'relevant'
-- and 'ingested' items. Steady state: each run triages and renders items
-- from prior ingestion.
--
-- Prerequisites:
--   - seed_content_sources_action.go deployed to chassis
--   - Registry entry added for "seed_content_sources"
--
-- Revert:
--   UPDATE agent_definitions
--   SET default_config = (SELECT default_config
--       FROM agent_def_content_feed_orch_backup_20260402 LIMIT 1),
--       updated_at = NOW()
--   WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;
-- ============================================================================

-- Step 1: Backup
CREATE TABLE IF NOT EXISTS agent_def_content_feed_orch_backup_20260402 AS
SELECT * FROM agent_definitions
WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;

-- Step 2: Update the workflow
UPDATE agent_definitions
SET default_config = jsonb_build_object(
        'processing_mode', 'orchestrator',
        'timeout_seconds', 900,
        'workflow', jsonb_build_object(
                'start_step', 'seed_sources',
                'steps', jsonb_build_object(

                    -- Step 1: Ensure content_sources exist from classification spec
                        'seed_sources', jsonb_build_object(
                                'action', 'seed_content_sources',
                                'config', jsonb_build_object('site_id', 'input_data.site_id'),
                                'next_step', 'check_has_sources',
                                'description', 'Create content_sources from classification news_feed recommendation if none exist',
                                'output_field', 'seed_result'
                                        ),

                    -- Step 2: Skip entire pipeline if site has no sources
                    -- seed_result.has_sources is true when existing sources found OR new ones seeded
                        'check_has_sources', jsonb_build_object(
                                'action', 'evaluate_condition',
                                'config', jsonb_build_object(
                                        'condition_field', 'seed_result.has_sources',
                                        'conditions', jsonb_build_object('false', 'complete_no_sources'),
                                        'default', 'dispatch_sources'
                                          ),
                                'description', 'Skip if site has no content sources (not recommended for news)'
                                             ),

                    -- Early exit: no sources
                        'complete_no_sources', jsonb_build_object(
                                'action', 'complete_workflow',
                                'config', jsonb_build_object(
                                        'output_fields', jsonb_build_array('seed_result')
                                          ),
                                'description', 'No content sources — site not configured for news'
                                               ),

                    -- Step 3: Dispatch feed ingesters for due sources (async, fire-and-forget)
                        'dispatch_sources', jsonb_build_object(
                                'action', 'dispatch_feed_sources',
                                'config', jsonb_build_object('site_id', 'input_data.site_id'),
                                'next_step', 'run_triage',
                                'description', 'Load due sources and spawn feed-ingester per source',
                                'output_field', 'dispatch_result'
                                            ),

                    -- Step 4: Triage ingested items from prior runs
                        'run_triage', jsonb_build_object(
                                'action', 'call_agent',
                                'config', jsonb_build_object(
                                        'agent_type', 'feed-triage',
                                        'input_fields', jsonb_build_array('input_data'),
                                        'timeout_seconds', 300
                                          ),
                                'next_step', 'render_news_json',
                                'description', 'Score ingested items for relevance (items from prior ingestion runs)',
                                'output_field', 'triage_result'
                                      ),

                    -- Step 5: Render latest-news.json
                        'render_news_json', jsonb_build_object(
                                'action', 'render_news_section',
                                'config', jsonb_build_object(
                                        'site_id', 'input_data.site_id',
                                        'max_items', 6,
                                        'page_name', 'index',
                                        'max_age_hours', 72
                                          ),
                                'next_step', 'check_has_news',
                                'description', 'Produce latest-news JSON from relevant and recent items',
                                'output_field', 'news_render_result'
                                            ),

                    -- Step 6: Skip git commit if no items rendered
                        'check_has_news', jsonb_build_object(
                                'action', 'evaluate_condition',
                                'config', jsonb_build_object(
                                        'condition_field', 'news_render_result.item_count',
                                        'conditions', jsonb_build_object('0', 'complete'),
                                        'default', 'commit_news'
                                          ),
                                'description', 'Skip commit if no news items to display'
                                          ),

                    -- Step 7: Commit JSON file to git → S3 deploy
                        'commit_news', jsonb_build_object(
                                'action', 'git_commit',
                                'config', jsonb_build_object(
                                        'files_field', 'news_render_result.files',
                                        'domain_field', 'news_render_result.domain',
                                        'commit_message', 'Update latest news feed'
                                          ),
                                'next_step', 'complete',
                                'description', 'Commit latest-news.json to git repo for S3 deploy',
                                'output_field', 'news_commit_result'
                                       ),

                    -- Done
                        'complete', jsonb_build_object(
                                'action', 'complete_workflow',
                                'config', jsonb_build_object(
                                        'output_fields', jsonb_build_array(
                                                'seed_result', 'dispatch_result', 'triage_result',
                                                'news_render_result', 'news_commit_result'
                                                         )
                                          ),
                                'description', 'Feed orchestration complete'
                                    )
                         )
                    )
                     ),
    updated_at = NOW()
WHERE type = 'content-feed-orchestrator'
  AND deleted_at IS NULL;

-- Verify the update
SELECT type,
       default_config->'workflow'->'start_step' as start_step,
       default_config->'timeout_seconds' as timeout
FROM agent_definitions
WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;

