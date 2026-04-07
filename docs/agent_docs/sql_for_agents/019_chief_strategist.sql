-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================

-- Add deduplication rules before the OUTPUT FORMAT section -- too many social proofs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_build_plan,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'generate_build_plan'->'config'->>'prompt_template',
                        'OUTPUT FORMAT (valid JSON only):',
                        E'COMPONENT PLACEMENT RULES:\n- testimonials/social-proof: Use on ONE page only (usually index). Never repeat across pages.\n- team-grid: Use on ONE page only (usually about).\n- cta-banner/cta-split: Acceptable on most pages but vary the headline/text.\n- services-grid/services-list: Use on index (summary) OR services (detail), not both with identical content.\n- hero sections: Every page should have a hero, but each page needs a DIFFERENT hero variant.\n- faq-accordion: Use on ONE page only.\n- contact-form: Use on ONE page only (contact).\n- Each page should have a distinct purpose with unique sections. If two pages look similar, merge them.\n\nOUTPUT FORMAT (valid JSON only):'
                )
        ),
        updated_at = NOW()
            WHERE type = 'chief-strategist';


-- Verify the prompt now contains the new rules
SELECT
    type,
    CASE WHEN default_config->'workflow'->'steps'->'generate_build_plan'->'config'->>'prompt_template'
        LIKE '%COMPONENT PLACEMENT RULES%'
    THEN 'YES - rules present'
  ELSE 'NO - rules missing'
END AS has_dedup_rules
FROM agent_definitions
WHERE type = 'chief-strategist';


    -- ============================================================================
-- 029b_planner_add_news_listing_component.sql
--
-- Adds news-listing to the available component types list in both
-- chief-strategist agent definitions (V1 and V2).
--
-- This is a minimal change — just makes the planner aware the component
-- exists. The actual creation of news pages for existing sites is handled
-- by the discovery → content-gap-planner path.
-- ============================================================================

-- Both V1 and V2 have the same component list line in their prompt_template.
-- The line in the JSON-escaped prompt reads:
--   - category-grid, content-feed, search-bar
-- We change it to:
--   - category-grid, content-feed, news-listing, search-bar

UPDATE agent_definitions
SET default_config = regexp_replace(
        default_config::text,
        'category-grid, content-feed, search-bar',
        'category-grid, content-feed, news-listing, search-bar',
        'g'  -- replace all occurrences (prompt appears once per version, but safe)
                     )::jsonb,
updated_at = NOW()
WHERE type = 'chief-strategist'
  AND deleted_at IS NULL
  AND default_config::text LIKE '%category-grid, content-feed, search-bar%';

-- ============================================================================
-- Verify
-- ============================================================================
-- SELECT type, version,
--        default_config::text LIKE '%news-listing%' as has_news_listing
-- FROM agent_definitions
-- WHERE type = 'chief-strategist' AND deleted_at IS NULL;