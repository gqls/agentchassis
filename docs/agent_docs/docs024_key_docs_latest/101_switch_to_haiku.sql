-- ============================================================================
-- 077: Switch all LLM agents to haiku to reduce token costs
-- Run the RESTORE section when ready to switch back
-- ============================================================================

-- CURRENT STATE (recorded 2026-04-10):
--
-- OPUS (switch to haiku):
--   build-site-planner        claude-opus-4-6
--   chief-strategist (x2)     claude-opus-4-6
--   tool-recreation-handler   claude-opus-4-6  (stage 2 only, stage 1 is sonnet)
--
-- SONNET (switch to haiku):
--   blog-content-planner      claude-sonnet-4-6
--   briefing-agent            claude-sonnet-4-6
--   build-briefing-agent      claude-sonnet-4-6
--   component-creator         claude-sonnet-4-6 (x2 refs)
--   content-creator           claude-sonnet-4-6 (1 of 2 refs, other already haiku)
--   content-gap-planner       claude-sonnet-4-6
--   content-quality-auditor   claude-sonnet-4-6
--   content-reviewer          claude-sonnet-4-6
--   css-patch-agent           claude-sonnet-4-6
--   domain-research-classifier claude-sonnet-4-6
--   domain-strategist         claude-sonnet-4-6
--   feed-triage               claude-sonnet-4-6 (x2 refs)
--   page-content-writer       claude-sonnet-4-6
--   reasoning                 claude-sonnet-4-6 (x2 refs)
--   research-agent            claude-sonnet-4-6 (1 of 2 refs, other already haiku)
--   researcher                claude-sonnet-4-6 (x2 refs)
--   site-adoption-agent       claude-sonnet-4-6 (x6 refs)
--   site-architect            claude-sonnet-4-6
--   site-classifier           claude-sonnet-4-6
--   site-planner              claude-sonnet-4-6
--   site-review-agent         claude-sonnet-4-6
--   site-scraper              claude-sonnet-4-6
--   tool-auditor              claude-sonnet-4-6
--   tool-generator            claude-sonnet-4-6
--   tool-improver             claude-sonnet-4-6
--   tool-recreation-handler   claude-sonnet-4-6 (stage 1)
--   tool-suggester            claude-sonnet-4-6
--   visual-design-auditor     claude-sonnet-4-6
--   webdesign-agent           claude-sonnet-4-6
--
-- ALREADY HAIKU (no change needed):
--   brand-designer, ch-llm-reviewer, content-creator (1 ref),
--   content-creator-about, content-creator-contact (x2),
--   content-creator-cta (x2), content-creator-features (x2),
--   content-creator-hero, content-creator-hero-without-research,
--   content-creator-testimonials (x2), content-researcher,
--   content_researcher, content-writer, copywriter (x2),
--   domain-analyst, html-developer, research-agent (1 ref),
--   simple-content-writer-with-approval, site-strategist,
--   vet-batch-processor, vet-practice-verifier, website-builder
--
-- NON-CLAUDE (no change):
--   image-generator           sdxl

-- ============================================================================
-- SWITCH TO HAIKU
-- ============================================================================

-- Switch all sonnet references to haiku
UPDATE agent_definitions
SET default_config = replace(default_config::text, 'claude-sonnet-4-6', 'claude-haiku-4-5')::jsonb
WHERE is_active = true
  AND default_config::text LIKE '%claude-sonnet-4-6%';

-- Switch all opus references to haiku
UPDATE agent_definitions
SET default_config = replace(default_config::text, 'claude-opus-4-6', 'claude-haiku-4-5')::jsonb
WHERE is_active = true
  AND default_config::text LIKE '%claude-opus-4-6%';

-- Verify
SELECT type, regexp_matches(default_config::text, '"model":\s*"([^"]+)"', 'g') as model
FROM agent_definitions
WHERE is_active = true
  AND default_config::text LIKE '%"model"%'
  AND default_config::text NOT LIKE '%haiku%'
  AND default_config::text NOT LIKE '%sdxl%'
ORDER BY type;
-- Should return 0 rows (everything is haiku or sdxl)


-- ============================================================================
-- RESTORE (run when ready to switch back)
-- ============================================================================
-- Copy everything below this line into a separate file for later use.

-- Restore opus agents
UPDATE agent_definitions
SET default_config = replace(default_config::text, 'claude-haiku-4-5', 'claude-opus-4-6')::jsonb
WHERE type IN ('build-site-planner')
  AND is_active = true;
--
-- -- chief-strategist has 2 rows
UPDATE agent_definitions
SET default_config = replace(default_config::text, 'claude-haiku-4-5', 'claude-opus-4-6')::jsonb
WHERE type = 'chief-strategist'
  AND is_active = true;
--
-- -- tool-recreation-handler stage 2 (opus) — tricky, it has both sonnet and opus
-- -- The recreate_tool step uses opus, analyze_tool uses sonnet
-- -- After restoring sonnet below, then target just the recreate step:
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,recreate_tool,config,ai_service,model}',
    '"claude-opus-4-6"'::jsonb
)
WHERE type = 'tool-recreation-handler' AND is_active = true;
--
-- -- Restore sonnet agents (everything that was sonnet but NOT the opus ones above)
UPDATE agent_definitions
SET default_config = replace(default_config::text, 'claude-haiku-4-5', 'claude-sonnet-4-6')::jsonb
WHERE type IN (
    'blog-content-planner', 'briefing-agent', 'build-briefing-agent',
    'component-creator', 'content-gap-planner', 'content-quality-auditor',
    'content-reviewer', 'css-patch-agent', 'domain-research-classifier',
    'domain-strategist', 'feed-triage', 'page-content-writer',
    'reasoning', 'researcher', 'site-adoption-agent',
    'site-architect', 'site-classifier', 'site-planner',
    'site-review-agent', 'site-scraper', 'tool-auditor',
    'tool-generator', 'tool-improver', 'tool-recreation-handler',
    'tool-suggester', 'visual-design-auditor', 'webdesign-agent'
) AND is_active = true;
--
-- -- Fix mixed agents: research-agent and content-creator have both haiku and sonnet refs
-- -- The replace above will have switched ALL their refs to sonnet
-- -- Need to restore the haiku ones:
-- -- research-agent: search step is haiku, synthesis step is sonnet (already correct after bulk restore)
-- -- content-creator: one version is haiku, one is sonnet (two rows with same type)
-- -- These may need manual checking after restore.
