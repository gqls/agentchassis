-- domain-research-classifier agent definition
-- Handler for: needs_domain_research work items
-- Pipeline position: first agent after seed_build_queue
--
-- Receives from dispatch loop:
--   input_data.site_id       — UUID of the site
--   input_data.domain        — domain name to research
--   input_data.objective     — optional objective hint
--   input_data.work_item_id  — the work item being processed
--
-- Outputs to site_specs:
--   aspect "identity"        — company name, tagline, industry, contact info
--   aspect "classification"  — site_type, confidence, detected_signals
--
-- Creates next work item:
--   needs_briefing → build-briefing-agent

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'domain-research-classifier',
             'Domain Research Classifier',
             'Researches a domain via web search and scrape, classifies site type, extracts identity signals, writes findings to site_specs, creates next work item.',
             'specialist',
             '{
                 "workflow": {
                     "steps": {

                         "search_domain": {
                             "action": "web_search",
                             "config": {
                                 "num_results": 8,
                                 "query_template": "{{.input_data.domain}} website company"
                             },
                             "next_step": "scrape_site",
                             "description": "Search for the domain to find its website and online presence",
                             "output_field": "search_results"
                         },

                         "scrape_site": {
                             "action": "scrape_web",
                             "config": {
                                 "url_field": "input_data.domain",
                                 "extract_mode": "text",
                                 "add_protocol": true,
                                 "max_pages": 3,
                                 "follow_links": ["about", "services", "contact", "team", "pricing"]
                             },
                             "next_step": "classify_and_extract",
                             "description": "Scrape the live site if it exists",
                             "output_field": "scraped_data"
                         },

                         "classify_and_extract": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5",
                                     "provider": "anthropic",
                                     "max_tokens": 3000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["input_data", "search_results", "scraped_data"],
                                 "output_format": "json",
                                 "prompt_template": "You are analyzing a domain for website creation.\n\nDomain: {{.input_data.domain}}\nObjective: {{if .input_data.objective}}{{.input_data.objective}}{{else}}Not specified — infer from domain name and research{{end}}\n\nSearch Results:\n{{.search_results}}\n\nScraped Website Content:\n{{.scraped_data}}\n\nAnalyze everything and return a JSON object with TWO sections:\n\n1. \"identity\" — what we know about this entity:\n```json\n{\n  \"company_name\": \"Best guess at company/brand name\",\n  \"tagline\": \"Tagline if found, null otherwise\",\n  \"industry\": \"Primary industry/sector\",\n  \"sub_industry\": \"More specific classification\",\n  \"about_summary\": \"2-3 sentence summary of what this company does\",\n  \"services\": [{\"name\": \"Service 1\", \"description\": \"Brief desc\"}],\n  \"contact\": {\n    \"email\": \"if found\",\n    \"phone\": \"if found\",\n    \"address\": \"if found\",\n    \"location\": \"city/region if found\"\n  },\n  \"has_existing_site\": true/false,\n  \"existing_site_quality\": \"good/adequate/poor/none\",\n  \"social_profiles\": {\"linkedin\": \"url\", \"twitter\": \"url\"},\n  \"key_people\": [{\"name\": \"Name\", \"role\": \"Role\"}],\n  \"unique_selling_points\": [\"USP 1\", \"USP 2\"],\n  \"target_audience\": \"Description of likely target audience\",\n  \"competitors_found\": [\"competitor1.com\", \"competitor2.com\"]\n}\n```\n\n2. \"classification\" — what kind of site to build:\n```json\n{\n  \"site_type\": \"brochure|landing|portfolio|content|ecommerce|tools\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Why this site type fits\",\n  \"recommended_builder\": \"pageflow-builder\",\n  \"page_count_estimate\": 4-8,\n  \"detected_signals\": [\"signal1\", \"signal2\"],\n  \"tone_suggestion\": \"professional|friendly|bold|technical|editorial\",\n  \"suggested_style\": \"professional-dark|modern-light|bold-creative|etc\"\n}\n```\n\nIMPORTANT:\n- If no existing site found, infer from the domain name\n- Be specific about industry — \"veterinary practice\" not just \"healthcare\"\n- List real competitors if search results reveal them\n- Choose site_type carefully based on the domain name hints and objective\n- For unknown domains, default to brochure with professional tone\n- recommended_builder should always be \"pageflow-builder\" for now\n\nReturn ONLY valid JSON with both identity and classification keys."
                             },
                             "next_step": "write_identity_spec",
                             "description": "LLM classifies site and extracts identity from research",
                             "output_field": "analysis"
                         },

                         "write_identity_spec": {
                             "action": "write_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "spec_data": "analysis.result.identity",
                                 "aspect": "identity",
                                 "source": "domain-research-classifier",
                                 "source_agent": "domain-research-classifier",
                                 "source_item_id": "input_data.work_item_id"
                             },
                             "next_step": "write_classification_spec",
                             "description": "Persist identity findings to site_specs",
                             "output_field": "identity_written"
                         },

                         "write_classification_spec": {
                             "action": "write_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "spec_data": "analysis.result.classification",
                                 "aspect": "classification",
                                 "source": "domain-research-classifier",
                                 "source_agent": "domain-research-classifier",
                                 "source_item_id": "input_data.work_item_id"
                             },
                             "next_step": "create_next_item",
                             "description": "Persist classification to site_specs",
                             "output_field": "classification_written"
                         },

                         "create_next_item": {
                             "action": "create_work_item",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_type": "needs_briefing",
                                 "handler_agent": "build-briefing-agent",
                                 "item_domain": "build",
                                 "severity": "high",
                                 "source": "domain-research-classifier",
                                 "summary": "Briefing needed after domain research",
                                 "item_key_prefix": "briefing",
                                 "priority": 10
                             },
                             "next_step": "complete",
                             "description": "Create the next work item for briefing stage",
                             "output_field": "next_item_created"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["analysis", "identity_written", "classification_written", "next_item_created"]
                             },
                             "description": "Research and classification complete"
                         }
                     },
                     "start_step": "search_domain"
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 180
             }'::jsonb,
             true,
             '["web-search", "scraping", "classification", "llm"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'analyst',
             'experimental',
             '["domain-research", "classification", "identity-extraction"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": ["site_id", "domain"], "optional": ["objective", "work_item_id"], "description": "Receives site_id + domain from dispatch loop. Optional objective from build_queue direction."}'::jsonb,
             '{"produces": {"identity_written": "site_spec write result for identity aspect", "classification_written": "site_spec write result for classification aspect", "next_item_created": "work item created for next pipeline stage"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();


---

-- reasoning about site strategy

-- 064b_patch_research_classifier_chain.sql
--
-- Change domain-research-classifier's create_next_item step to chain to
-- domain-strategist instead of build-briefing-agent.
--
-- Before: needs_domain_research → domain-research-classifier → needs_briefing → build-briefing-agent
-- After:  needs_domain_research → domain-research-classifier → needs_strategy → domain-strategist → needs_briefing → build-briefing-agent
--
-- The domain-research-classifier still does: search → scrape → extract identity + basic classification
-- The domain-strategist then does deeper reasoning about site strategy before briefing begins.
--
-- NOTE: The classifier's classify_and_extract LLM prompt still outputs a "classification" section.
-- That's fine — the strategist reads it as input and overwrites it with a richer version.
-- We keep the classifier's basic classification because:
--   a) It's useful input for the strategist (what did first-pass analysis suggest?)
--   b) If the strategist fails, at least we have a basic classification to fall back on

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_next_item,config}',
        '{
          "site_id": "input_data.site_id",
          "item_type": "needs_strategy",
          "handler_agent": "domain-strategist",
          "item_domain": "build",
          "severity": "high",
          "source": "domain-research-classifier",
          "summary": "Domain strategy analysis needed after research",
          "item_key_prefix": "strategy",
          "priority": 8
        }'::jsonb
                     )
WHERE type = 'domain-research-classifier';

-- Also update the step description for clarity
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_next_item,description}',
        '"Create work item for domain strategy analysis"'::jsonb
                     )
WHERE type = 'domain-research-classifier';


---


-- ============================================================================
-- 1. Update build-site-planner prompt to produce design_intent and content_direction
-- 2. Upgrade models: classifier + planner → opus-4-6, everything else → sonnet-4-6
--
-- Extended thinking (budget_tokens) requires a Go change to AnthropicClient.
-- That's a follow-up. For now, opus-4-6 without extended thinking is still
-- a significant upgrade for architectural decisions.
-- ============================================================================


-- ============================================================================
-- 1a. Update build-site-planner: add design_intent and content_direction
--     to the LLM prompt's JSON output format
-- ============================================================================

-- The prompt is a very long string. We'll use string replacement to add
-- the new fields to the JSON template section and rules.

-- Add design_intent and content_direction to the JSON output section
-- Find the closing of image_prompts and add after it
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
                replace(
                        replace(
                                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                            -- Add design_intent and content_direction to the JSON example
                                '  "image_prompts": {' || E'\n' || '    "logo": "Description for logo generation",' || E'\n' || '    "hero_home": "Description for home hero image"' || E'\n' || '  }' || E'\n' || '}',
                                '  "image_prompts": {' || E'\n' || '    "logo": "Description for logo generation",' || E'\n' || '    "hero_home": "Description for home hero image"' || E'\n' || '  },' || E'\n' ||
                '  "design_intent": {' || E'\n' ||
                '    "style_direction": "professional-dark or modern-light or bold-creative",' || E'\n' ||
                '    "colour_mood": "Description of colour feeling and why it fits the industry",' || E'\n' ||
                '    "typography_mood": "Description of font personality",' || E'\n' ||
                '    "imagery_direction": "What images should show",' || E'\n' ||
                '    "layout_preference": "Layout style description",' || E'\n' ||
                '    "avoid": ["Things to avoid in design"]' || E'\n' ||
                '  },' || E'\n' ||
                '  "content_direction": {' || E'\n' ||
                '    "voice": "How the site should sound",' || E'\n' ||
                '    "emphasis": "What to emphasise in content",' || E'\n' ||
                '    "avoid_phrases": ["Phrases to never use"],' || E'\n' ||
                '    "social_proof_style": "How to handle testimonials and proof",' || E'\n' ||
                '    "blog_strategy": "Content strategy for blog if applicable"' || E'\n' ||
                '  }' || E'\n' ||
                '}'
                        ),
                    -- Add rules 10 and 11 for the new fields
                        'Return ONLY valid JSON.',
                        '10. Include design_intent with explicit colour mood, typography direction, and layout preferences based on the industry and identity' || E'\n' ||
            '11. Include content_direction with voice, emphasis, and avoid_phrases tailored to the target audience and tone' || E'\n' || E'\n' ||
            'Return ONLY valid JSON.'
                )
        )
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL;


-- ============================================================================
-- 1b. Also update the domain-research-classifier to produce design_intent
--     and content_direction in its analysis output
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template',
                        'Return ONLY valid JSON with both identity and classification keys.',
                        'Also include a third key:\n\n3. "content_direction" — how content should be written:\n```json\n{\n  "voice": "How the brand should sound (e.g. ''experienced practitioners who cut through hype'')",\n  "tone": "professional|friendly|bold|technical|editorial|conversational",\n  "emphasis": "What to emphasise across all content",\n  "avoid_phrases": ["corporate jargon to avoid"],\n  "social_proof_style": "How to handle testimonials (e.g. ''company philosophy, not fake quotes'')",\n  "trust_signals": "What authority signals this industry needs"\n}\n```\n\n4. "design_intent" — visual direction:\n```json\n{\n  "style_direction": "professional-dark|modern-light|bold-creative",\n  "colour_mood": "Description of colour feeling and why",\n  "typography_mood": "Font personality description",\n  "imagery_direction": "What images should convey",\n  "layout_preference": "Layout style",\n  "avoid": ["Design elements to avoid"]\n}\n```\n\nReturn ONLY valid JSON with identity, classification, content_direction, and design_intent keys.'
                )
        )
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND deleted_at IS NULL;

-- Add write steps for the new aspects in domain-research-classifier
-- After write_classification_spec, add write_content_direction and write_design_intent

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,write_classification_spec,next_step}',
        '"write_content_direction_spec"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,write_content_direction_spec}',
        '{
            "action": "write_site_spec",
            "config": {
                "aspect": "content_direction",
                "source": "domain-research-classifier",
                "site_id": "input_data.site_id",
                "spec_data": "analysis.result.content_direction",
                "source_agent": "domain-research-classifier",
                "source_item_id": "input_data.work_item_id"
            },
            "next_step": "write_design_intent_spec",
            "error_step": "create_next_item",
            "description": "Persist content direction to site_specs",
            "output_field": "content_direction_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,write_design_intent_spec}',
        '{
            "action": "write_site_spec",
            "config": {
                "aspect": "design_intent",
                "source": "domain-research-classifier",
                "site_id": "input_data.site_id",
                "spec_data": "analysis.result.design_intent",
                "source_agent": "domain-research-classifier",
                "source_item_id": "input_data.work_item_id"
            },
            "next_step": "create_next_item",
            "error_step": "create_next_item",
            "description": "Persist design intent to site_specs",
            "output_field": "design_intent_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND deleted_at IS NULL;


-- ============================================================================
-- 2. Upgrade LLM models across all agents
--
-- Classifier + planner → claude-opus-4-6 (architectural decisions)
-- Everything else → claude-sonnet-4-6 (content, audit, tools)
--
-- Using aliases from model_aliases.go which resolve to full version strings.
-- ============================================================================

-- Classifier → opus-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,ai_service,model}',
        '"claude-opus-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND deleted_at IS NULL;

-- Build-site-planner → opus-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,ai_service,model}',
        '"claude-opus-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- Page-content-writer → sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,ai_service,model}',
        '"claude-sonnet-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Visual-design-auditor → sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_visual_llm_audit,config,ai_service,model}',
        '"claude-sonnet-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'visual-design-auditor' AND deleted_at IS NULL;

-- Content-quality-auditor → sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_content_llm_audit,config,ai_service,model}',
        '"claude-sonnet-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'content-quality-auditor' AND deleted_at IS NULL;

-- Site-review-agent → sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_strategic_review,config,ai_service,model}',
        '"claude-sonnet-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-review-agent' AND deleted_at IS NULL;

-- Tool-suggester → sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,config,ai_service,model}',
        '"claude-sonnet-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'tool-suggester' AND deleted_at IS NULL;

-- Tool-improver → sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,improve_tool,config,ai_service,model}',
        '"claude-sonnet-4-6"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'tool-improver' AND deleted_at IS NULL;


-- ============================================================================
-- Verify
-- ============================================================================

-- Check models
SELECT type,
       COALESCE(
               default_config->'workflow'->'steps'->'classify_and_extract'->'config'->'ai_service'->>'model',
        default_config->'workflow'->'steps'->'plan_site'->'config'->'ai_service'->>'model',
        default_config->'workflow'->'steps'->'run_visual_llm_audit'->'config'->'ai_service'->>'model',
        default_config->'workflow'->'steps'->'run_content_llm_audit'->'config'->'ai_service'->>'model',
        default_config->'workflow'->'steps'->'run_strategic_review'->'config'->'ai_service'->>'model',
        default_config->'workflow'->'steps'->'suggest_tools'->'config'->'ai_service'->>'model',
        default_config->'workflow'->'steps'->'improve_tool'->'config'->'ai_service'->>'model',
        'nested-in-loop'
    ) as model
FROM agent_definitions
WHERE type IN (
               'domain-research-classifier', 'build-site-planner',
               'visual-design-auditor', 'content-quality-auditor',
               'site-review-agent', 'tool-suggester', 'tool-improver'
    )
  AND deleted_at IS NULL
ORDER BY type;

-- Check classifier now writes 4 spec aspects
SELECT type,
       default_config->'workflow'->'steps'->'write_classification_spec'->>'next_step' as after_classification,
    default_config->'workflow'->'steps'->'write_content_direction_spec'->>'next_step' as after_content_dir,
    default_config->'workflow'->'steps'->'write_design_intent_spec'->>'next_step' as after_design_intent
FROM agent_definitions
WHERE type = 'domain-research-classifier' AND deleted_at IS NULL;

-- Check planner prompt contains design_intent
SELECT
    default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template' LIKE '%design_intent%' as has_design_intent,
    default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template' LIKE '%content_direction%' as has_content_direction
FROM agent_definitions
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

--

-- ============================================================================
-- Mission-Driven Site Support (Brief-Based)
-- ============================================================================
--
-- Two data formats travel through the pipeline:
--   mission/roadmap     — structured JSON for machine consumption
--                         (content writers, plan_sections, component selector)
--   mission_brief/roadmap_brief — plain text for LLM prompts
--                         (classifier, planner, any model can read these)
--
-- The trigger script writes both. The domain-submitter stores all four
-- as separate site_spec aspects. Prompts reference the briefs.
--
-- Changes:
--   1. domain-submitter: 4 new persist steps (mission, roadmap, mission_brief, roadmap_brief)
--   2. classifier prompt: add {{.input_data.mission_brief}} section
--   3. planner prompt: add {{.site_specs.specs.mission_brief}} and {{.site_specs.specs.roadmap_brief}} sections
-- ============================================================================


-- ============================================================================
-- 1. DOMAIN-SUBMITTER: Add persist steps for mission data
-- ============================================================================
-- Current flow: ensure_site_record → store_contact_info → store_submission_spec → create_research_item → complete
-- New flow:     ensure_site_record → store_contact_info → store_submission_spec → persist_mission → persist_mission_brief → persist_roadmap → persist_roadmap_brief → create_research_item → complete
--
-- Each persist step uses error_step to skip when the field isn't in input_data.
-- Normal domain submissions (no mission) pass through unchanged.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps}',
        default_config->'workflow'->'steps'
            || jsonb_build_object(
                                            'persist_mission', jsonb_build_object(
                            'action', 'write_site_spec',
                            'description', 'Persist structured mission data (skipped if absent)',
                            'config', jsonb_build_object(
                                    'aspect', 'mission',
                                    'source', 'domain-submitter',
                                    'site_id', 'site_record.site_id',
                                    'spec_data', 'input_data.mission',
                                    'source_agent', 'domain-submitter'
                                      ),
                            'next_step', 'persist_mission_brief',
                            'error_step', 'persist_mission_brief',
                            'output_field', 'mission_stored'
                                                               ),
                                            'persist_mission_brief', jsonb_build_object(
                                                    'action', 'write_site_spec',
                                                    'description', 'Persist mission brief text for LLM prompts (skipped if absent)',
                                                    'config', jsonb_build_object(
                                                            'aspect', 'mission_brief',
                                                            'source', 'domain-submitter',
                                                            'site_id', 'site_record.site_id',
                                                            'spec_data', 'input_data.mission_brief',
                                                            'source_agent', 'domain-submitter'
                                                              ),
                                                    'next_step', 'persist_roadmap',
                                                    'error_step', 'persist_roadmap',
                                                    'output_field', 'mission_brief_stored'
                                                                     ),
                                            'persist_roadmap', jsonb_build_object(
                                                    'action', 'write_site_spec',
                                                    'description', 'Persist structured roadmap data (skipped if absent)',
                                                    'config', jsonb_build_object(
                                                            'aspect', 'roadmap',
                                                            'source', 'domain-submitter',
                                                            'site_id', 'site_record.site_id',
                                                            'spec_data', 'input_data.roadmap',
                                                            'source_agent', 'domain-submitter'
                                                              ),
                                                    'next_step', 'persist_roadmap_brief',
                                                    'error_step', 'persist_roadmap_brief',
                                                    'output_field', 'roadmap_stored'
                                                               ),
                                            'persist_roadmap_brief', jsonb_build_object(
                                                    'action', 'write_site_spec',
                                                    'description', 'Persist roadmap brief text for LLM prompts (skipped if absent)',
                                                    'config', jsonb_build_object(
                                                            'aspect', 'roadmap_brief',
                                                            'source', 'domain-submitter',
                                                            'site_id', 'site_record.site_id',
                                                            'spec_data', 'input_data.roadmap_brief',
                                                            'source_agent', 'domain-submitter'
                                                              ),
                                                    'next_step', 'create_research_item',
                                                    'error_step', 'create_research_item',
                                                    'output_field', 'roadmap_brief_stored'
                                                                     )
               )
                     ),
    updated_at = NOW()
WHERE type = 'domain-submitter' AND is_active = true;

-- Wire store_submission_spec → persist_mission (instead of → create_research_item)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_submission_spec,next_step}',
        '"persist_mission"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-submitter' AND is_active = true;


-- ============================================================================
-- 2. CLASSIFIER: Add mission_brief to prompt
-- ============================================================================
-- The classifier receives input_data directly from the trigger.
-- mission_brief is a flat text string — no nested access needed.
-- When absent, the {{if}} block renders empty. Existing sites unaffected.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb(
                E'You are analyzing a domain for website creation.\n\nDomain: {{.input_data.domain}}\nObjective: {{if .input_data.objective}}{{.input_data.objective}}{{else}}Not specified — infer from domain name and research{{end}}\n\n{{if .input_data.mission_brief}}## Pre-Defined Mission\nThis site has a strategic mission provided by the owner. Use this as STRONG guidance for identity, tone, positioning, and design direction. The research below validates and supplements — the mission is the primary source.\n\n{{.input_data.mission_brief}}\n{{end}}\n{{if .input_data.roadmap_brief}}## Roadmap\n{{.input_data.roadmap_brief}}\n{{end}}\nSearch Results:\n{{.search_results}}\n\nScraped Website Content:\n{{.scraped_data}}\n\nAnalyze everything and return a JSON object with FOUR sections:\n\n1. \"identity\" — what we know about this entity:\n```json\n{\n  \"company_name\": \"Best guess at company/brand name\",\n  \"tagline\": \"Tagline if found, null otherwise\",\n  \"industry\": \"Primary industry/sector\",\n  \"sub_industry\": \"More specific classification\",\n  \"about_summary\": \"2-3 sentence summary of what this company does\",\n  \"services\": [{\"name\": \"Service 1\", \"description\": \"Brief desc\"}],\n  \"contact\": {\n    \"email\": \"if found\",\n    \"phone\": \"if found\",\n    \"address\": \"if found\",\n    \"location\": \"city/region if found\"\n  },\n  \"has_existing_site\": true/false,\n  \"existing_site_quality\": \"good/adequate/poor/none\",\n  \"social_profiles\": {\"linkedin\": \"url\", \"twitter\": \"url\"},\n  \"key_people\": [{\"name\": \"Name\", \"role\": \"Role\"}],\n  \"unique_selling_points\": [\"USP 1\", \"USP 2\"],\n  \"target_audience\": \"Description of likely target audience\",\n  \"competitors_found\": [\"competitor1.com\", \"competitor2.com\"]\n}\n```\n\n2. \"classification\" — what kind of site to build:\n```json\n{\n  \"site_type\": \"brochure|landing|portfolio|content|ecommerce|tools|interactive-platform\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Why this site type fits\",\n  \"recommended_builder\": \"pageflow-builder\",\n  \"page_count_estimate\": 4-8,\n  \"detected_signals\": [\"signal1\", \"signal2\"],\n  \"tone_suggestion\": \"professional|friendly|bold|technical|editorial|game-like|energetic\",\n  \"suggested_style\": \"professional-dark|modern-light|bold-creative|etc\"\n}\n```\n\n3. \"content_direction\" — how content should be written:\n```json\n{\n  \"voice\": \"How the brand should sound\",\n  \"tone\": \"professional|friendly|bold|technical|editorial|conversational|game-like\",\n  \"emphasis\": \"What to emphasise across all content\",\n  \"avoid_phrases\": [\"corporate jargon to avoid\"],\n  \"social_proof_style\": \"How to handle testimonials\",\n  \"trust_signals\": \"What authority signals this industry needs\"\n}\n```\n\n4. \"design_intent\" — visual direction:\n```json\n{\n  \"style_direction\": \"professional-dark|modern-light|bold-creative\",\n  \"colour_mood\": \"Description of colour feeling and why\",\n  \"typography_mood\": \"Font personality description\",\n  \"imagery_direction\": \"What images should convey\",\n  \"layout_preference\": \"Layout style\",\n  \"avoid\": [\"Design elements to avoid\"]\n}\n```\n\nIMPORTANT:\n- If a mission is provided above, derive identity/tone/design from it. Research validates.\n- If no mission, infer everything from domain name, research, and objective.\n- If no existing site found, infer from the domain name.\n- Be specific about industry — \"veterinary practice\" not just \"healthcare\".\n- For interactive-platform or novel site types, use those in site_type rather than forcing brochure.\n- recommended_builder should always be \"pageflow-builder\" for now.\n\nReturn ONLY valid JSON with identity, classification, content_direction, and design_intent keys.'
        )
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND is_active = true;


-- ============================================================================
-- 3. PLANNER: Add mission_brief and roadmap_brief to prompt
-- ============================================================================
-- The planner reads from site_specs. mission_brief and roadmap_brief are
-- flat text strings stored as their own aspects.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
                E'Plan a website for {{.input_data.domain}}.\n\n## Research Data\nIdentity: {{.site_specs.specs.identity}}\nClassification: {{.site_specs.specs.classification}}\n\n## Domain Strategy\n{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}\n\n## Briefing Answers\n{{.site_specs.specs.briefing}}\n\n{{if .site_specs.specs.mission_brief}}## Mission\n{{.site_specs.specs.mission_brief}}\n{{end}}\n{{if .site_specs.specs.roadmap_brief}}## Roadmap\nBuild ONLY what is in the current phase described below. When pages list section_types, use those in the sections arrays. Section types that do not match existing component names will be resolved by the component selector — output them as-is. Do NOT invent pages beyond what the current phase specifies.\n\n{{.site_specs.specs.roadmap_brief}}\n{{end}}\n## Available Section Components\nYou MUST use ONLY these exact component names in the \"sections\" arrays (unless the roadmap specifies section_types — those override):\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Canonical Page Types\n\nEvery page MUST have a page_type from this list:\n\n| page_type | Description | Use for |\n|-----------|-------------|----------|\n| `index` | Home page | Always exactly one |\n| `content` | Standard content page | About, services, contact, FAQ, etc |\n| `landing` | Conversion-focused page | Lead capture, specific offers |\n| `entity-directory` | Searchable directory of entities | Business listings, provider directories |\n| `entity-page` | Individual entity profile | Single business/provider detail page |\n| `tool` | Interactive tool or calculator | Cost calculators, comparison tools |\n| `blog-index` | Blog/news listing page | Article index, news feed |\n| `blog-post` | Individual blog article | Editorial content, guides |\n\nNot all page types have builders available yet. Plan the IDEAL site regardless — the build system handles which pages can be built now vs later.\n\n## Strategy Guidance\n\nIf a domain strategy is available above, use it as strong input:\n- The recommended site_type should guide the overall structure\n- The recommended page_types should inform which pages you plan\n- The revenue model should shape what conversion/lead-capture mechanisms you include\n- The tone should influence your style_collection choice\n\nYou have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.\n\nReturn JSON:\n```json\n{\n  \"site_type\": \"from the strategy, roadmap, or your own assessment\",\n  \"strategy_notes\": \"any notes on how you used or diverged from the strategy/roadmap\",\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"page_type\": \"index\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call-to-action\"]\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  },\n  \"design_intent\": {\n    \"style_direction\": \"professional-dark or modern-light or bold-creative\",\n    \"colour_mood\": \"Description of colour feeling and why it fits the industry\",\n    \"typography_mood\": \"Description of font personality\",\n    \"imagery_direction\": \"What images should show\",\n    \"layout_preference\": \"Layout style description\",\n    \"avoid\": [\"Things to avoid in design\"]\n  },\n  \"content_direction\": {\n    \"voice\": \"How the site should sound\",\n    \"emphasis\": \"What to emphasise in content\",\n    \"avoid_phrases\": [\"Phrases to never use\"],\n    \"social_proof_style\": \"How to handle testimonials and proof\",\n    \"blog_strategy\": \"Content strategy for blog if applicable\"\n  }\n}\n```\n\nRULES:\n1. Use component names from the Available Section Components list for sections arrays — UNLESS the roadmap specifies section_types, in which case use those\n2. Every page MUST have a page_type from the canonical list\n3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays\n4. Content and index pages need sections arrays populated\n5. Always include: index (home) and contact pages (unless the roadmap says otherwise)\n6. Keep header navigation to 5-8 items maximum\n7. Set needs_logo: true and needs_images: true (always)\n8. Provide detailed image_prompts for logo and hero_home\n9. Include design_intent with explicit colour mood, typography direction, and layout preferences\n10. Include content_direction with voice, emphasis, and avoid_phrases tailored to the target audience\n11. If the classification data includes content_features.news_feed.recommended = true, add \"latest-news\" to the homepage sections\n12. When a roadmap is present, the pages and section_types from the current phase take precedence over your own page planning\n\nReturn ONLY valid JSON.'
        )
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active = true;


-- ============================================================================
-- Verification
-- ============================================================================

-- Verify domain-submitter flow
SELECT key as step_name, value->>'next_step' as next_step, value->>'error_step' as error_step
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE type = 'domain-submitter' AND is_active = true
ORDER BY
    CASE key
    WHEN 'ensure_site_record' THEN 1
    WHEN 'store_contact_info' THEN 2
    WHEN 'store_submission_spec' THEN 3
    WHEN 'persist_mission' THEN 4
    WHEN 'persist_mission_brief' THEN 5
    WHEN 'persist_roadmap' THEN 6
    WHEN 'persist_roadmap_brief' THEN 7
    WHEN 'create_research_item' THEN 8
    WHEN 'complete' THEN 9
END;

-- Verify classifier prompt uses mission_brief (flat string, no nesting)
SELECT
    default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' LIKE '%mission_brief%' as uses_brief_not_structured,
    default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' NOT LIKE '%mission.mission%' as no_nested_access,
    default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' LIKE '%interactive-platform%' as has_interactive_type
FROM agent_definitions
WHERE type = 'domain-research-classifier' AND is_active = true;

-- Verify planner prompt uses roadmap_brief (flat string, no nesting)
SELECT
    default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template' LIKE '%roadmap_brief%' as uses_brief_not_structured,
    default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template' LIKE '%mission_brief%' as has_mission_brief,
    default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template' NOT LIKE '%roadmap.current_phase%' as no_nested_access
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true;

-- ============================================================================
-- Wire existing content into page-build-handler and content writer
-- ============================================================================
--
-- 1. page-build-handler: add load_existing_content step
-- 2. page-build-handler: pass existing_content to content writer
-- 3. page-content-writer: add existing_content to section loop prompt
--
-- ============================================================================

-- ── 1. Add load_existing_content step to page-build-handler ─────────────
-- Inserts between check_page_found and plan_sections

-- First: redirect check_page_found → load_existing_content (was → plan_sections)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_page_found,config,then_step}',
        '"load_existing_content"'
                     )
WHERE type = 'page-build-handler';

-- Then: add the new step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_content}',
        '{
            "action": "load_existing_content",
            "config": {
                "site_id": "site_record.site_id",
                "page_name": "page_record.name",
                "page_id": "page_record.id",
                "mode": "input_data.spec.mode"
            },
            "output_field": "existing_content",
            "next_step": "plan_sections",
            "error_step": "plan_sections",
            "description": "Load existing page content from adoption crawl if mode is recreate"
        }'
                     )
WHERE type = 'page-build-handler';

-- ── 2. Pass existing_content to content writer via input_mapping ────────
-- Add existing_content to the call_content_writer input_mapping
-- Uses ? suffix because non-adoption pages won't have this in collected_data
-- (per guidelines: item-type-specific fields in input_mapping must use ?)

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_content_writer,config,input_mapping,"existing_content?"}',
        '"existing_content"'
                     )
WHERE type = 'page-build-handler';

-- Also pass the mode from the work item spec
-- Uses ? suffix because non-adoption work items have no spec.mode
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_content_writer,config,input_mapping,"build_mode?"}',
        '"input_data.spec.mode"'
                     )
WHERE type = 'page-build-handler';

-- ── 3. Add existing_content to content writer prompt ────────────────────
-- The generate_content step needs existing_content in input_fields
-- and a conditional block in the prompt template

-- Add to input_fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}',
        '["current_section", "render_context", "reviewed_brief", "current_page", "link_context", "site_plan", "site_specs", "existing_content", "build_mode"]'
                     )
WHERE type = 'page-content-writer';

-- ============================================================================
-- NOTE: The prompt_template update below prepends a recreate-mode block.
-- It uses {{if .existing_content.has_existing}} to conditionally include
-- the adoption content. For non-adopted pages this block is skipped entirely.
--
-- Because the prompt is very long and we don't want to replace the whole thing,
-- we prepend the recreate block to the existing prompt by reading current value
-- and concatenating. But jsonb_set with string concat is awkward in SQL.
--
-- Simpler approach: update just the prompt_template with the full new value.
-- The existing prompt is preserved below with the recreate block added at the
-- end before the STRICT RULES section.
-- ============================================================================

-- We add the existing content block just before "## STRICT RULES"
-- by appending to the prompt template. This is the cleanest SQL approach.

-- First, let's verify the current prompt contains "## STRICT RULES"
-- then inject the recreate block before it.

DO $$
DECLARE
current_prompt text;
    new_prompt text;
    recreate_block text;
BEGIN
    -- Get current prompt
SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
INTO current_prompt
FROM agent_definitions
WHERE type = 'page-content-writer';

-- The recreate block to inject
recreate_block := E'\n\n{{if .existing_content}}{{if .existing_content.has_existing}}\n## EXISTING CONTENT — Recreate Mode\nThis page is being adopted from an existing site. Below is the original content from the page.\nYour task for this section: find the content relevant to the "{{.current_section.category}}" section and adapt it to fit the JSON schema above.\n\nPrioritise preserving the original meaning and information. Adapt the structure to match the required JSON field names.\nIf the existing content does not have material relevant to this section, generate fresh content as you normally would.\n\nOriginal page content:\n{{.existing_content.raw_markdown}}\n{{end}}{{end}}';

    -- Inject before ## STRICT RULES
    IF position('## STRICT RULES' in current_prompt) > 0 THEN
        new_prompt := replace(current_prompt, '## STRICT RULES', recreate_block || E'\n## STRICT RULES');
ELSE
        -- Fallback: append at end
        new_prompt := current_prompt || recreate_block;
END IF;

    -- Update the prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(new_prompt)
                     )
WHERE type = 'page-content-writer';

RAISE NOTICE 'Prompt updated. Recreate block injected before STRICT RULES.';
END $$;

--

-- ============================================================================
-- Enrich domain-research-classifier content_direction output
-- ============================================================================
-- The classifier already writes content_direction but produces a thin spec.
-- This updates the prompt to ask for the same rich structure that adoption
-- produces, adapted for prescriptive (what the site SHOULD sound like)
-- rather than descriptive (what it DOES sound like).
--
-- The write_site_spec action auto-formats it via FormatContentDirection.
-- ============================================================================

DO $outer$
DECLARE
prompt_text text;
BEGIN
    prompt_text := $prompt$You are analyzing a domain for website creation.

Domain: {{.input_data.domain}}
Objective: {{if .input_data.objective}}{{.input_data.objective}}{{else}}Not specified — infer from domain name and research{{end}}

{{if .input_data.mission_brief}}## Pre-Defined Mission
This site has a strategic mission provided by the owner. Use this as STRONG guidance for identity, tone, positioning, and design direction. The research below validates and supplements — the mission is the primary source.

{{.input_data.mission_brief}}
{{end}}
{{if .input_data.roadmap_brief}}## Roadmap
{{.input_data.roadmap_brief}}
{{end}}
Search Results:
{{.search_results}}

Scraped Website Content:
{{.scraped_data}}

Analyze everything and return a JSON object with FOUR sections. Be as specific and detailed as possible — vague guidance like "professional tone" is useless. Respond ONLY with valid JSON.

1. "identity" — what we know about this entity:
{
  "company_name": "Best guess at company/brand name",
  "tagline": "Tagline if found, null otherwise",
  "industry": "Primary industry/sector",
  "sub_industry": "More specific classification",
  "about_summary": "2-3 sentence summary of what this company does",
  "services": [{"name": "Service 1", "description": "Brief desc"}],
  "contact": {
    "email": "if found",
    "phone": "if found",
    "address": "if found",
    "location": "city/region if found"
  },
  "has_existing_site": true,
  "existing_site_quality": "good/adequate/poor/none",
  "social_profiles": {"linkedin": "url", "twitter": "url"},
  "key_people": [{"name": "Name", "role": "Role"}],
  "unique_selling_points": ["USP 1", "USP 2"],
  "target_audience": "Description of likely target audience",
  "competitors_found": ["competitor1.com", "competitor2.com"]
}

2. "classification" — what kind of site to build:
{
  "site_type": "brochure|landing|portfolio|content|ecommerce|tools|interactive-platform",
  "confidence": 0.0-1.0,
  "reasoning": "Why this site type fits",
  "recommended_builder": "pageflow-builder",
  "page_count_estimate": 5,
  "detected_signals": ["signal1", "signal2"],
  "tone_suggestion": "professional|friendly|bold|technical|editorial|game-like|energetic",
  "suggested_style": "professional-dark|modern-light|bold-creative|etc"
}

3. "content_direction" — DETAILED writing style guide for all content on this site:
{
  "voice": {
    "register": "Describe the overall register this site should use — e.g. 'Down-to-earth and matter-of-fact, avoids sales language'",
    "person": "Which grammatical person — e.g. 'Second person (you/your) for the reader, first plural (we) for the company'",
    "authority_level": "How the site should establish authority — e.g. 'States expertise confidently but never promises specific outcomes'",
    "emotional_tone": "The emotional quality — e.g. 'Reassuring without being patronising'",
    "formality": "Where on the formal-casual spectrum — e.g. 'Professional but not stiff. Uses contractions.'"
  },
  "sentence_style": {
    "average_length": "Short/medium/long for this industry",
    "structure_patterns": "How sentences should be constructed",
    "rhythm": "The cadence appropriate for this audience"
  },
  "persuasion_approach": {
    "method": "How the site should convince without hard selling",
    "trust_building": "How trust should be established for this industry",
    "social_proof_style": "How to handle testimonials and social proof"
  },
  "content_depth": {
    "thoroughness": "How deeply topics should be explored",
    "assumed_knowledge": "What the target audience already knows",
    "explanation_pattern": "How complex ideas should be introduced"
  },
  "writing_rules": [
    "At least 8 specific, actionable rules appropriate for this industry",
    "Example for finance: 'Never promise specific returns or savings'",
    "Example for health: 'Always recommend consulting a professional'",
    "Example for tech: 'Explain jargon on first use'"
  ],
  "compliance_rules": [
    "Industry-specific legal/regulatory rules — empty array if none apply",
    "Example for finance: 'Calculations are illustrative — do not constitute financial advice'",
    "Example for health: 'Information is not a substitute for professional medical advice'"
  ],
  "heading_style": {
    "format": "How headings should be phrased for this industry",
    "hierarchy": "How heading levels should be used"
  },
  "paragraph_style": {
    "typical_length": "Appropriate paragraph length for the audience",
    "structure": "How paragraphs should be organised"
  },
  "cta_style": {
    "approach": "How calls-to-action should be handled",
    "verb_choices": "What action verbs are appropriate"
  },
  "terminology": {
    "approach": "How domain-specific terms should be handled",
    "key_terms": ["List 5-10 key terms for this industry that content will need"]
  },
  "things_to_avoid": [
    "At least 6 specific things to avoid for this industry",
    "Example: 'Urgency language (limited time, act now)'",
    "Example: 'Unsubstantiated claims about results'"
  ],
  "things_to_emulate": [
    "At least 6 positive patterns appropriate for this industry",
    "Example: 'Lead with the user benefit, not the feature'"
  ],
  "example_phrases": {
    "characteristic": ["3-5 example phrases showing the right tone for this site"],
    "would_never_say": ["3-5 phrases this site should never use"]
  }
}

4. "design_intent" — visual direction:
{
  "style_direction": "professional-dark|modern-light|bold-creative",
  "colour_mood": "Description of colour feeling and why",
  "typography_mood": "Font personality description",
  "imagery_direction": "What images should convey",
  "layout_preference": "Layout style",
  "avoid": ["Design elements to avoid"]
}

IMPORTANT:
- If a mission is provided above, derive identity/tone/design from it. Research validates.
- If no mission, infer everything from domain name, research, and objective.
- If no existing site found, infer from the domain name and industry norms.
- Be specific about industry — "veterinary practice" not just "healthcare".
- content_direction.writing_rules must have at least 8 entries.
- content_direction.things_to_avoid and things_to_emulate must each have at least 6 entries.
- content_direction.compliance_rules should be populated for regulated industries (finance, health, legal, insurance) and empty [] for others.
- Every field in content_direction must be specific to THIS industry, not generic writing advice.
- For interactive-platform or novel site types, use those in site_type rather than forcing brochure.
- recommended_builder should always be "pageflow-builder" for now.

Return ONLY valid JSON with identity, classification, content_direction, and design_intent keys.$prompt$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb(prompt_text)
                     )
WHERE type = 'domain-research-classifier';

RAISE NOTICE 'Classifier prompt updated with rich content_direction structure.';
END $outer$;

-- Also bump max_tokens since content_direction is now much larger
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,ai_service,max_tokens}',
        '6000'
                     )
WHERE type = 'domain-research-classifier';

-- Update model to claude-sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,ai_service,model}',
        '"claude-sonnet-4-6"'
                     )
WHERE type = 'domain-research-classifier';

---
-- muddling with site specs

-- 1. Add read_site_specs step to classifier
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,read_site_specs}',
        jsonb_build_object(
                'action', 'read_site_spec',
                'description', 'Load site_specs (includes mission_brief and roadmap_brief if present)',
                'config', jsonb_build_object(
                        'site_id', 'input_data.site_id'
                          ),
                'next_step', 'classify_and_extract',
                'output_field', 'site_specs'
        )
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND is_active = true;

-- 2. Wire scrape_site → read_site_specs (instead of → classify_and_extract)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,scrape_site,next_step}',
        '"read_site_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND is_active = true;

-- 3. Add site_specs to the LLM step's input_fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,input_fields}',
        '["input_data", "search_results", "scraped_data", "site_specs"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND is_active = true;

-- 4. Fix prompt: input_data.mission_brief → site_specs.specs.mission_brief
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb(REPLACE(REPLACE(
                                 default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template',
                                 '.input_data.mission_brief',
                                 '.site_specs.specs.mission_brief'
                         ),
                         '.input_data.roadmap_brief',
                         '.site_specs.specs.roadmap_brief'
                 )::text)
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND is_active = true;

-- Verify the flow
SELECT key as step_name, value->>'next_step' as next_step
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE type = 'domain-research-classifier' AND is_active = true
ORDER BY
    CASE key
    WHEN 'search_domain' THEN 1
    WHEN 'scrape_site' THEN 2
    WHEN 'read_site_specs' THEN 3
    WHEN 'classify_and_extract' THEN 4
    WHEN 'write_identity_spec' THEN 5
    WHEN 'write_classification_spec' THEN 6
    WHEN 'write_content_direction_spec' THEN 7
    WHEN 'write_design_intent_spec' THEN 8
    WHEN 'create_next_item' THEN 9
    WHEN 'complete' THEN 10
END;
-- Should show: search_domain → scrape_site → read_site_specs → classify_and_extract → ...

-- Verify prompt references site_specs not input_data for briefs
SELECT
    default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' LIKE '%site_specs.specs.mission_brief%' as reads_from_site_specs,
    default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' NOT LIKE '%input_data.mission_brief%' as no_input_data_ref
FROM agent_definitions WHERE type = 'domain-research-classifier' AND is_active = true;
-- Expected: t, t

-- backup
e6ca8cca-398b-49f9-a435-337b1cc26c38 | domain-research-classifier | Domain Research Classifier | Researches a domain via web search and scrape, classifies site type, extracts identity signals, writes findings to site_specs, creates next work item. | specialist | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["analysis", "identity_written", "classification_written", "next_item_created"]}, "description": "Research and classification complete"}, "scrape_site": {"action": "scrape_web", "config": {"max_pages": 3, "url_field": "input_data.domain", "add_protocol": true, "extract_mode": "text", "follow_links": ["about", "services", "contact", "team", "pricing"]}, "next_step": "read_site_specs", "description": "Scrape the live site if it exists", "output_field": "scraped_data"}, "search_domain": {"action": "web_search", "config": {"num_results": 8, "query_template": "{{.input_data.domain}} website company"}, "next_step": "scrape_site", "description": "Search for the domain to find its website and online presence", "output_field": "search_results"}, "read_site_specs": {"action": "read_site_spec", "config": {"site_id": "input_data.site_id"}, "next_step": "classify_and_extract", "description": "Load site_specs (includes mission_brief and roadmap_brief if present)", "output_field": "site_specs"}, "create_next_item": {"action": "create_work_item", "config": {"source": "domain-research-classifier", "site_id": "input_data.site_id", "summary": "Domain strategy analysis needed after research", "priority": 8, "severity": "high", "item_type": "needs_strategy", "item_domain": "build", "handler_agent": "domain-strategist", "item_key_prefix": "strategy"}, "next_step": "complete", "description": "Create work item for domain strategy analysis", "output_field": "next_item_created"}, "write_identity_spec": {"action": "write_site_spec", "config": {"aspect": "identity", "source": "domain-research-classifier", "site_id": "input_data.site_id", "spec_data": "analysis.result.identity", "source_agent": "domain-research-classifier", "source_item_id": "input_data.work_item_id"}, "next_step": "write_classification_spec", "description": "Persist identity findings to site_specs", "output_field": "identity_written"}, "classify_and_extract": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 6000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["input_data", "search_results", "scraped_data", "site_specs"], "output_format": "json", "prompt_template": "You are analyzing a domain for website creation.\n\nDomain: {{.input_data.domain}}\nObjective: {{if .input_data.objective}}{{.input_data.objective}}{{else}}Not specified — infer from domain name and research{{end}}\n\n{{if .site_specs.specs.mission_brief}}## Pre-Defined Mission\nThis site has a strategic mission provided by the owner. Use this as STRONG guidance for identity, tone, positioning, and design direction. The research below validates and supplements — the mission is the primary source.\n\n{{.site_specs.specs.mission_brief.text}}\n{{end}}\n{{if .site_specs.specs.roadmap_brief}}## Roadmap\n{{.site_specs.specs.roadmap_brief.text}}\n{{end}}\nSearch Results:\n{{.search_results}}\n\nScraped Website Content:\n{{.scraped_data}}\n\nAnalyze everything and return a JSON object with FOUR sections. Be as specific and detailed as possible — vague guidance like \"professional tone\" is useless. Respond ONLY with valid JSON.\n\n1. \"identity\" — what we know about this entity:\n{\n  \"company_name\": \"Best guess at company/brand name\",\n  \"tagline\": \"Tagline if found, null otherwise\",\n  \"industry\": \"Primary industry/sector\",\n  \"sub_industry\": \"More specific classification\",\n  \"about_summary\": \"2-3 sentence summary of what this company does\",\n  \"services\": [{\"name\": \"Service 1\", \"description\": \"Brief desc\"}],\n  \"contact\": {\n    \"email\": \"if found\",\n    \"phone\": \"if found\",\n    \"address\": \"if found\",\n    \"location\": \"city/region if found\"\n  },\n  \"has_existing_site\": true,\n  \"existing_site_quality\": \"good/adequate/poor/none\",\n  \"social_profiles\": {\"linkedin\": \"url\", \"twitter\": \"url\"},\n  \"key_people\": [{\"name\": \"Name\", \"role\": \"Role\"}],\n  \"unique_selling_points\": [\"USP 1\", \"USP 2\"],\n  \"target_audience\": \"Description of likely target audience\",\n  \"competitors_found\": [\"competitor1.com\", \"competitor2.com\"]\n}\n\n2. \"classification\" — what kind of site to build:\n{\n  \"site_type\": \"brochure|landing|portfolio|content|ecommerce|tools|interactive-platform\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Why this site type fits\",\n  \"recommended_builder\": \"pageflow-builder\",\n  \"page_count_estimate\": 5,\n  \"detected_signals\": [\"signal1\", \"signal2\"],\n  \"tone_suggestion\": \"professional|friendly|bold|technical|editorial|game-like|energetic\",\n  \"suggested_style\": \"professional-dark|modern-light|bold-creative|etc\"\n}\n\n3. \"content_direction\" — DETAILED writing style guide for all content on this site:\n{\n  \"voice\": {\n    \"register\": \"Describe the overall register this site should use — e.g. 'Down-to-earth and matter-of-fact, avoids sales language'\",\n    \"person\": \"Which grammatical person — e.g. 'Second person (you/your) for the reader, first plural (we) for the company'\",\n    \"authority_level\": \"How the site should establish authority — e.g. 'States expertise confidently but never promises specific outcomes'\",\n    \"emotional_tone\": \"The emotional quality — e.g. 'Reassuring without being patronising'\",\n    \"formality\": \"Where on the formal-casual spectrum — e.g. 'Professional but not stiff. Uses contractions.'\"\n  },\n  \"sentence_style\": {\n    \"average_length\": \"Short/medium/long for this industry\",\n    \"structure_patterns\": \"How sentences should be constructed\",\n    \"rhythm\": \"The cadence appropriate for this audience\"\n  },\n  \"persuasion_approach\": {\n    \"method\": \"How the site should convince without hard selling\",\n    \"trust_building\": \"How trust should be established for this industry\",\n    \"social_proof_style\": \"How to handle testimonials and social proof\"\n  },\n  \"content_depth\": {\n    \"thoroughness\": \"How deeply topics should be explored\",\n    \"assumed_knowledge\": \"What the target audience already knows\",\n    \"explanation_pattern\": \"How complex ideas should be introduced\"\n  },\n  \"writing_rules\": [\n    \"At least 8 specific, actionable rules appropriate for this industry\",\n    \"Example for finance: 'Never promise specific returns or savings'\",\n    \"Example for health: 'Always recommend consulting a professional'\",\n    \"Example for tech: 'Explain jargon on first use'\"\n  ],\n  \"compliance_rules\": [\n    \"Industry-specific legal/regulatory rules — empty array if none apply\",\n    \"Example for finance: 'Calculations are illustrative — do not constitute financial advice'\",\n    \"Example for health: 'Information is not a substitute for professional medical advice'\"\n  ],\n  \"heading_style\": {\n    \"format\": \"How headings should be phrased for this industry\",\n    \"hierarchy\": \"How heading levels should be used\"\n  },\n  \"paragraph_style\": {\n    \"typical_length\": \"Appropriate paragraph length for the audience\",\n    \"structure\": \"How paragraphs should be organised\"\n  },\n  \"cta_style\": {\n    \"approach\": \"How calls-to-action should be handled\",\n    \"verb_choices\": \"What action verbs are appropriate\"\n  },\n  \"terminology\": {\n    \"approach\": \"How domain-specific terms should be handled\",\n    \"key_terms\": [\"List 5-10 key terms for this industry that content will need\"]\n  },\n  \"things_to_avoid\": [\n    \"At least 6 specific things to avoid for this industry\",\n    \"Example: 'Urgency language (limited time, act now)'\",\n    \"Example: 'Unsubstantiated claims about results'\"\n  ],\n  \"things_to_emulate\": [\n    \"At least 6 positive patterns appropriate for this industry\",\n    \"Example: 'Lead with the user benefit, not the feature'\"\n  ],\n  \"example_phrases\": {\n    \"characteristic\": [\"3-5 example phrases showing the right tone for this site\"],\n    \"would_never_say\": [\"3-5 phrases this site should never use\"]\n  }\n}\n\n4. \"design_intent\" — visual direction:\n{\n  \"style_direction\": \"professional-dark|modern-light|bold-creative\",\n  \"colour_mood\": \"Description of colour feeling and why\",\n  \"typography_mood\": \"Font personality description\",\n  \"imagery_direction\": \"What images should convey\",\n  \"layout_preference\": \"Layout style\",\n  \"avoid\": [\"Design elements to avoid\"]\n}\n\nIMPORTANT:\n- If a mission is provided above, derive identity/tone/design from it. Research validates.\n- If no mission, infer everything from domain name, research, and objective.\n- If no existing site found, infer from the domain name and industry norms.\n- Be specific about industry — \"veterinary practice\" not just \"healthcare\".\n- content_direction.writing_rules must have at least 8 entries.\n- content_direction.things_to_avoid and things_to_emulate must each have at least 6 entries.\n- content_direction.compliance_rules should be populated for regulated industries (finance, health, legal, insurance) and empty [] for others.\n- Every field in content_direction must be specific to THIS industry, not generic writing advice.\n- For interactive-platform or novel site types, use those in site_type rather than forcing brochure.\n- recommended_builder should always be \"pageflow-builder\" for now.\n\nReturn ONLY valid JSON with identity, classification, content_direction, and design_intent keys."}, "next_step": "write_identity_spec", "description": "LLM classifies site and extracts identity from research", "output_field": "analysis"}, "write_design_intent_spec": {"action": "write_site_spec", "config": {"aspect": "design_intent", "source": "domain-research-classifier", "site_id": "input_data.site_id", "spec_data": "analysis.result.design_intent", "error_step": "create_next_item", "source_agent": "domain-research-classifier", "source_item_id": "input_data.work_item_id"}, "next_step": "create_next_item", "description": "Persist design intent to site_specs", "output_field": "design_intent_written"}, "write_classification_spec": {"action": "write_site_spec", "config": {"aspect": "classification", "source": "domain-research-classifier", "site_id": "input_data.site_id", "spec_data": "analysis.result.classification", "source_agent": "domain-research-classifier", "source_item_id": "input_data.work_item_id"}, "next_step": "write_content_direction_spec", "description": "Persist classification to site_specs", "output_field": "classification_written"}, "write_content_direction_spec": {"action": "write_site_spec", "config": {"aspect": "content_direction", "source": "domain-research-classifier", "site_id": "input_data.site_id", "spec_data": "analysis.result.content_direction", "error_step": "create_next_item", "source_agent": "domain-research-classifier", "source_item_id": "input_data.work_item_id"}, "next_step": "write_design_intent_spec", "description": "Persist content direction to site_specs", "output_field": "content_direction_written"}}, "start_step": "search_domain"}, "processing_mode": "orchestrator", "timeout_seconds": 180} | t         | 2026-02-24 16:07:35.422594+00 | 2026-04-21 16:43:49.737291+00 |            | ["web-search", "scraping", "classification", "llm"] | docker.io/aqls/agent-chassis | v1.0.981  |         | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} | analyst        | experimental | ["domain-research", "classification", "identity-extraction"] | {}                     |           0 | f           | {"optional": ["objective", "work_item_id"], "required": ["site_id", "domain"], "description": "Receives site_id + domain from dispatch loop. Optional objective from build_queue direction."} | {"produces": {"identity_written": "site_spec write result for identity aspect", "next_item_created": "work item created for next pipeline stage", "classification_written": "site_spec write result for classification aspect"}} |                    0
(1 row)


-- change prompt to allow for adoption
-- 006_classifier_adoption_aware_prompt.sql
--
-- Make domain-research-classifier's classify_and_extract prompt adoption-aware.
--
-- Background
-- ----------
-- When a site is adopted via site-adoption-agent, adoption writes identity,
-- site_archetype, content_direction, design_intent, and design_reference
-- specs. Shortly after, validate_composition_inputs (first step in
-- site-design-planner's workflow) sees that classification is not present
-- and queues a needs_domain_research work item as a self-heal. The
-- domain-research-classifier then runs.
--
-- Observed on gamesdesign.co.uk (site_id available via lookup):
--   19:29  adoption wrote faithful identity + site_archetype + content_direction
--   20:14  classifier ran, web-searched the blank destination domain, invented
--          a consultancy identity, overwrote identity + content_direction +
--          design_intent with consultancy content. site_archetype was not
--          overwritten (only adoption writes that aspect).
--   21:31  build-site-planner wrote content_direction + design_intent a third
--          time. (Tracked separately as an open question — see doc 028.)
--
-- Root cause for the 20:14 overwrite: the classifier's prompt template surfaces
-- mission_brief and roadmap_brief from site_specs but nothing else. Adoption
-- output is in site_specs.specs.{identity,site_archetype,content_direction,
-- design_intent} but the prompt doesn't render it, so the LLM sees the
-- destination as a blank domain with only web search to work from.
--
-- What this patch does
-- --------------------
-- Extends the prompt_template so that when site_archetype exists in site_specs
-- (which only adoption writes), the prompt includes the adopted identity,
-- archetype, content_direction, and design_intent as ground truth, plus
-- explicit fidelity guidance: the first deployed version should stay close
-- to the adopted source, with minimal aspirational extension at this stage.
--
-- No workflow shape changes, no new actions. The read_site_specs step
-- already loads all aspects into site_specs.specs.<aspect>. The
-- execute_llm_prompt step already has site_specs in input_fields. Only the
-- prompt_template text changes.
--
-- Template uses the toJSON pipe (registered in datahelpers.RenderPromptTemplate)
-- to render nested specs inline.
--
-- Fidelity dial (deferred)
-- ------------------------
-- The prompt currently implements fidelity=high as the implicit default when
-- adoption is present. A future patch can accept an explicit fidelity value
-- in input_data (high / medium / low) and branch the guidance accordingly.
-- See doc 028 "Adoption as reference material".
--
-- Scope
-- -----
-- Only classify_and_extract.config.prompt_template is modified. All other
-- fields (ai_service, input_fields, output_format, step ordering) are
-- untouched. The rest of the agent_definitions row is untouched.
--
-- Rollback
-- --------
--   UPDATE agent_definitions
--   SET default_config = (
--       SELECT default_config FROM agent_definitions_backup_20260422
--       WHERE type = 'domain-research-classifier'
--   )
--   WHERE type = 'domain-research-classifier';
-- ----------------------------------------------------------------------------

BEGIN;

-- Safety: refuse to run if the dated backup is missing
DO $guard$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'agent_definitions_backup_20260422'
    ) THEN
        RAISE EXCEPTION 'agent_definitions_backup_20260422 not found — run 005_agent_definitions_backup_20260422.sql first';
END IF;
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions_backup_20260422
        WHERE type = 'domain-research-classifier'
    ) THEN
        RAISE EXCEPTION 'Backup row for domain-research-classifier not found in agent_definitions_backup_20260422';
END IF;
END
$guard$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb($prompt$You are analyzing a domain for website creation.

Domain: {{.input_data.domain}}
Objective: {{if .input_data.objective}}{{.input_data.objective}}{{else}}Not specified — infer from domain name and research{{end}}

{{if .site_specs.specs.mission_brief}}## Pre-Defined Mission
This site has a strategic mission provided by the owner. Use this as STRONG guidance for identity, tone, positioning, and design direction. The research below validates and supplements — the mission is the primary source.

{{.site_specs.specs.mission_brief.text}}
{{end}}
{{if .site_specs.specs.roadmap_brief}}## Roadmap
{{.site_specs.specs.roadmap_brief.text}}
{{end}}
{{if .site_specs.specs.site_archetype}}## Adoption Reference — STRONGEST signal

This site has been adopted from {{if and .site_specs.specs.identity .site_specs.specs.identity.adopted_from}}{{.site_specs.specs.identity.adopted_from}}{{else}}an existing source{{end}}. The adoption agent crawled the source and captured what this site IS right now. The first deployed version should closely resemble the adopted source. Finder/fixer agents can widen any aspirational gap over time; your job in this run is to produce a spec that is faithful to the adopted state, with only the minimal extension needed to fill fields the adopted specs do not cover.

Treat the following adopted specs as ground truth. Do NOT fabricate services, competitors, unique selling points, or positioning that contradict them. Do NOT rewrite design character that the adopted design_intent already describes.

### Adopted identity
{{.site_specs.specs.identity | toJSON}}

### Adopted archetype (what kind of site this is)
{{.site_specs.specs.site_archetype | toJSON}}
{{if .site_specs.specs.content_direction}}
### Adopted content direction (writing style)
{{.site_specs.specs.content_direction | toJSON}}
{{end}}{{if .site_specs.specs.design_intent}}
### Adopted design intent (visual direction)
{{.site_specs.specs.design_intent | toJSON}}
{{end}}

### Adoption fidelity guidance (follow precisely)
- classification.site_type MUST be consistent with site_archetype.content.primary_type and site_archetype.label. A site archetyped as a tools/utility platform with primary content "interactive tools and calculators" is site_type "tools" or "interactive-platform" — NOT "brochure".
- classification.tone_suggestion MUST be consistent with site_archetype.character.feel and the adopted content_direction voice.
- classification.suggested_style MUST be consistent with the adopted design_intent — specifically the palette character and dark/light scheme. If the adopted design is dark, suggested_style must be a dark variant.
- identity.has_existing_site MUST be true — adoption evidence is present. identity.existing_site_quality should reflect site_archetype.character.polish.
- identity.services — include ONLY what the adopted identity already contains or what site_archetype.purpose and content_model clearly imply. Do not invent services for a site whose archetype indicates no commercial services (e.g. revenue_model "none", content primarily tools or guides).
- identity.competitors_found — leave empty unless the adopted identity already has competitors listed.
- identity.unique_selling_points — derive from site_archetype.purpose and adopted content; do not invent.
- content_direction — use the adopted content_direction as the baseline. Preserve its voice, writing_rules, things_to_avoid, and example_phrases. You may fill fields it does not populate, but do not substitute different wording for fields that do exist.
- design_intent — preserve the adopted design_intent's palette reference values, typography, and dark/light scheme. You may refine guidance text or add character description, but do not propose different colours or a different scheme.
- Preserve adopted_from and adopted_at in the identity output.
{{end}}
Search Results:
{{.search_results}}

Scraped Website Content:
{{.scraped_data}}

Analyze everything and return a JSON object with FOUR sections. Be as specific and detailed as possible — vague guidance like "professional tone" is useless. Respond ONLY with valid JSON.

1. "identity" — what we know about this entity:
{
  "company_name": "Best guess at company/brand name",
  "tagline": "Tagline if found, null otherwise",
  "industry": "Primary industry/sector",
  "sub_industry": "More specific classification",
  "about_summary": "2-3 sentence summary of what this company does",
  "services": [{"name": "Service 1", "description": "Brief desc"}],
  "contact": {
    "email": "if found",
    "phone": "if found",
    "address": "if found",
    "location": "city/region if found"
  },
  "has_existing_site": true,
  "existing_site_quality": "good/adequate/poor/none",
  "social_profiles": {"linkedin": "url", "twitter": "url"},
  "key_people": [{"name": "Name", "role": "Role"}],
  "unique_selling_points": ["USP 1", "USP 2"],
  "target_audience": "Description of likely target audience",
  "competitors_found": ["competitor1.com", "competitor2.com"]
}

2. "classification" — what kind of site to build:
{
  "site_type": "brochure|landing|portfolio|content|ecommerce|tools|interactive-platform",
  "confidence": 0.0-1.0,
  "reasoning": "Why this site type fits",
  "recommended_builder": "pageflow-builder",
  "page_count_estimate": 5,
  "detected_signals": ["signal1", "signal2"],
  "tone_suggestion": "professional|friendly|bold|technical|editorial|game-like|energetic",
  "suggested_style": "professional-dark|modern-light|bold-creative|etc"
}

3. "content_direction" — DETAILED writing style guide for all content on this site:
{
  "voice": {
    "register": "Describe the overall register this site should use — e.g. 'Down-to-earth and matter-of-fact, avoids sales language'",
    "person": "Which grammatical person — e.g. 'Second person (you/your) for the reader, first plural (we) for the company'",
    "authority_level": "How the site should establish authority — e.g. 'States expertise confidently but never promises specific outcomes'",
    "emotional_tone": "The emotional quality — e.g. 'Reassuring without being patronising'",
    "formality": "Where on the formal-casual spectrum — e.g. 'Professional but not stiff. Uses contractions.'"
  },
  "sentence_style": {
    "average_length": "Short/medium/long for this industry",
    "structure_patterns": "How sentences should be constructed",
    "rhythm": "The cadence appropriate for this audience"
  },
  "persuasion_approach": {
    "method": "How the site should convince without hard selling",
    "trust_building": "How trust should be established for this industry",
    "social_proof_style": "How to handle testimonials and social proof"
  },
  "content_depth": {
    "thoroughness": "How deeply topics should be explored",
    "assumed_knowledge": "What the target audience already knows",
    "explanation_pattern": "How complex ideas should be introduced"
  },
  "writing_rules": [
    "At least 8 specific, actionable rules appropriate for this industry",
    "Example for finance: 'Never promise specific returns or savings'",
    "Example for health: 'Always recommend consulting a professional'",
    "Example for tech: 'Explain jargon on first use'"
  ],
  "compliance_rules": [
    "Industry-specific legal/regulatory rules — empty array if none apply",
    "Example for finance: 'Calculations are illustrative — do not constitute financial advice'",
    "Example for health: 'Information is not a substitute for professional medical advice'"
  ],
  "heading_style": {
    "format": "How headings should be phrased for this industry",
    "hierarchy": "How heading levels should be used"
  },
  "paragraph_style": {
    "typical_length": "Appropriate paragraph length for the audience",
    "structure": "How paragraphs should be organised"
  },
  "cta_style": {
    "approach": "How calls-to-action should be handled",
    "verb_choices": "What action verbs are appropriate"
  },
  "terminology": {
    "approach": "How domain-specific terms should be handled",
    "key_terms": ["List 5-10 key terms for this industry that content will need"]
  },
  "things_to_avoid": [
    "At least 6 specific things to avoid for this industry",
    "Example: 'Urgency language (limited time, act now)'",
    "Example: 'Unsubstantiated claims about results'"
  ],
  "things_to_emulate": [
    "At least 6 positive patterns appropriate for this industry",
    "Example: 'Lead with the user benefit, not the feature'"
  ],
  "example_phrases": {
    "characteristic": ["3-5 example phrases showing the right tone for this site"],
    "would_never_say": ["3-5 phrases this site should never use"]
  }
}

4. "design_intent" — visual direction:
{
  "style_direction": "professional-dark|modern-light|bold-creative",
  "colour_mood": "Description of colour feeling and why",
  "typography_mood": "Font personality description",
  "imagery_direction": "What images should convey",
  "layout_preference": "Layout style",
  "avoid": ["Design elements to avoid"]
}

IMPORTANT:
- If adoption reference is provided above, the Adoption fidelity guidance is MANDATORY. Adoption evidence outranks mission_brief, search_results, and scraped_data for identity, content_direction, and design_intent. Follow the fidelity rules precisely — this is the user's captured existing site, not a blank slate.
- If a mission is provided above (and no adoption), derive identity/tone/design from it. Research validates.
- If no mission and no adoption, infer everything from domain name, research, and objective.
- If no existing site found and no adoption, infer from the domain name and industry norms.
- Be specific about industry — "veterinary practice" not just "healthcare".
- content_direction.writing_rules must have at least 8 entries.
- content_direction.things_to_avoid and things_to_emulate must each have at least 6 entries.
- content_direction.compliance_rules should be populated for regulated industries (finance, health, legal, insurance) and empty [] for others.
- Every field in content_direction must be specific to THIS industry, not generic writing advice.
- For interactive-platform or novel site types, use those in site_type rather than forcing brochure.
- recommended_builder should always be "pageflow-builder" for now.

Return ONLY valid JSON with identity, classification, content_direction, and design_intent keys.
$prompt$::text)
)
WHERE type = 'domain-research-classifier'
  AND is_active = true;

-- Verify the change took (should return exactly 1 row)
SELECT
    type,
    is_active,
    length(default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template') AS prompt_len,
    substring(
            default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template'
        FROM 'Adoption Reference[^\n]*'
    ) AS adoption_block_marker
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

-- Defensive: verify that input_fields still includes site_specs (it should —
-- this patch doesn't touch it, but if a prior edit removed it, the new
-- prompt would render with empty site_specs references)
SELECT
    type,
    default_config->'workflow'->'steps'->'classify_and_extract'->'config'->'input_fields' AS input_fields
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

COMMIT;

-- ----------------------------------------------------------------------------
-- Post-apply notes
--
-- 1. No chassis rollout needed — agent_definitions are read live per invocation.
--    A classifier pod that is mid-run will finish on the old prompt; the next
--    invocation uses the new one.
--
-- 2. First test case: retrigger adoption on gamesdesign.co.uk (or any
--    previously-drifted site) after clearing its site_specs and pages. The
--    expected behaviour:
--      - Adoption runs, writes identity/site_archetype/content_direction/design_intent
--      - validate_composition_inputs queues a needs_domain_research item (same as before)
--      - domain-research-classifier runs, reads adoption specs, produces a
--        classification that matches the archetype (e.g. site_type=tools, not brochure)
--        and an identity that preserves adopted_from / has_existing_site=true /
--        doesn't invent consultancy services
--      - content_direction and design_intent produced by the classifier should
--        remain substantially consistent with what adoption wrote
--
-- 3. Observability: check llm_call_log after the next adoption. The prompt
--    length will be noticeably longer than the prior prompt when site_archetype
--    is present in site_specs. Output token counts per domain may also be
--    slightly higher because the LLM is asked to respect more structured input.
--
-- 4. Separately tracked (NOT in this patch):
--      - build-site-planner writing content_direction/design_intent a third time
--      - derive_content_direction's commercial-bias framing (senior copywriter etc.)
--      - analyze_site's hard-coded services[] schema
--      - Explicit fidelity input variable (high / medium / low) — currently
--        defaulted to "high when adoption present, unconstrained otherwise"
-- ----------------------------------------------------------------------------

-- enrich prompt - hardcodes values so not gonna use it


-- 008_classifier_emits_category_and_industry_tags.sql
--
-- Update the domain-research-classifier prompt to emit two new fields
-- on the classification spec:
--   category:      one of brochure|interactive|social|editorial|ecommerce|
--                  portfolio|hub|docs
--   industry_tags: array of 4-10 specific tags describing the site shape
--                  and vertical
--
-- These two fields are consumed by site-design-planner's layout matcher
-- (resolveLayoutByTags). Without them, tagSet is empty and the matcher
-- takes the "no classification tags" fallback path to brochure-formal —
-- which is what gamesdesign.co.uk hit.
--
-- Pairs with migration 007 which seeded tool-portal-dark and social-lobby
-- with industry_tags tuned to intersect the taxonomy this migration
-- introduces.
--
-- Strategy: surgical regexp_replace inside a DO block that first verifies
-- the current prompt matches the 006 state. This is safer than a full
-- prompt rewrite — if anyone edited the prompt between 006 and 008 the
-- migration aborts with a clear error rather than silently overwriting.
-- The previous migration (006) did full-rewrite; this one is targeted
-- because the change scope is genuinely narrow (one JSON schema block
-- plus one taxonomy guidance block plus one adoption-fidelity bullet).
--
-- ----------------------------------------------------------------------------
-- Rollback (undo this migration):
--   UPDATE agent_definitions
--   SET default_config = (
--       SELECT default_config FROM agent_definitions_backup_20260423
--       WHERE type = 'domain-research-classifier'
--   )
--   WHERE type = 'domain-research-classifier' AND is_active = true;
--
-- Then restart the chassis rollout so pods pick up the reverted prompt:
--   kubectl -n ai-persona-system rollout restart deployment/<chassis-deployment>
-- ----------------------------------------------------------------------------

BEGIN;

-- ---------------------------------------------------------------
-- 1. Back up the current (post-006) state before modifying.
-- ---------------------------------------------------------------

DROP TABLE IF EXISTS agent_definitions_backup_20260423;
CREATE TABLE agent_definitions_backup_20260423 AS
SELECT * FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

SELECT 'BACKUP' AS phase, COUNT(*) AS rows_backed_up
FROM agent_definitions_backup_20260423;


-- ---------------------------------------------------------------
-- 2. Apply the prompt update inside a guarded DO block.
-- ---------------------------------------------------------------

DO $migration$
DECLARE
old_prompt TEXT;
    step1_prompt TEXT;
    step2_prompt TEXT;
    new_prompt TEXT;
    expected_marker TEXT := '"suggested_style": "professional-dark|modern-light|bold-creative|etc"';
BEGIN
    -- Read the current prompt.
SELECT default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template'
INTO old_prompt
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

IF old_prompt IS NULL THEN
        RAISE EXCEPTION
            '008: classifier prompt not found. Is the domain-research-classifier agent_definition present and active?';
END IF;

    -- Pre-state sanity check: the 006 marker must be present. If not,
    -- something has changed the prompt and we should not run blind.
    IF position(expected_marker IN old_prompt) = 0 THEN
        RAISE EXCEPTION
            '008: expected the 006 prompt marker (%) but it is not present. Aborting — prompt is not in the state we can safely patch. Inspect the current prompt and update this migration before re-running.', expected_marker;
END IF;

    -- Edit 1: Replace the classification JSON schema block with the extended
    -- version that includes category and industry_tags.
    step1_prompt := regexp_replace(
        old_prompt,
        -- Match from `2. "classification"` through the closing `}` of the schema.
        -- [\s\S]*? is a non-greedy any-char match that works across newlines.
        -- The block contains no nested `{`/`}` so the first `}\n` we hit is
        -- the closing brace.
        '2\. "classification"[\s\S]*?\n\}\n',
        E'2. "classification" — what kind of site to build:\n'
        E'{\n'
        E'  "site_type": "brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social",\n'
        E'  "category": "brochure|interactive|social|editorial|ecommerce|portfolio|hub|docs",\n'
        E'  "industry_tags": ["specific-tag-1", "specific-tag-2", "4 to 10 tags total"],\n'
        E'  "confidence": 0.0-1.0,\n'
        E'  "reasoning": "Why this site type fits",\n'
        E'  "recommended_builder": "pageflow-builder",\n'
        E'  "page_count_estimate": 5,\n'
        E'  "detected_signals": ["signal1", "signal2"],\n'
        E'  "tone_suggestion": "professional|friendly|bold|technical|editorial|game-like|energetic",\n'
        E'  "suggested_style": "professional-dark|modern-light|bold-creative|etc"\n'
        E'}\n'
        E'\n'
        E'category and industry_tags are used by the layout matcher (site-design-planner) to pick a CSS layout from the library. The matcher scores layouts by how many of these tags overlap with each layout''s own industry_tags.\n'
        E'\n'
        E'category — the highest-level shape of the site. Pick ONE:\n'
        E'- brochure — traditional marketing/company site (services, about, contact)\n'
        E'- interactive — tool/utility platform, calculators, simulators, practitioner tools\n'
        E'- social — community platform, forum, social network, AI-seeded discussion\n'
        E'- editorial — news, magazine, blog, opinion, long-form content\n'
        E'- ecommerce — storefront, product catalogue, checkout\n'
        E'- portfolio — work showcase for a creative, studio, or individual\n'
        E'- hub — vertical information hub, regulatory resource, trade directory (non-commercial)\n'
        E'- docs — technical documentation, API reference, knowledge base\n'
        E'\n'
        E'industry_tags — an array of 4-10 specific tags describing the site. Include BOTH:\n'
        E'- site-shape tags: "interactive-platform", "tool-portal", "social-network", "provocation-platform", "magazine-grid", "affiliate-hub", "docs-sidebar", etc.\n'
        E'- vertical/domain tags: "game-design", "veterinary", "legal", "fitness", "fintech", "wellness", etc.\n'
        E'\n'
        E'Use hyphenated lowercase. Prefer specific over generic: "game-design" not "games", "ai-platform" not "ai".\n'
        E'\n'
        E'Existing library tags (non-exhaustive, use as inspiration): consultancy, law, finance, b2b, professional-services, tech-startup, fitness, creative-studios, design-agency, news, opinion, long-form-blog, video-platform, audio-library, podcast, developer-docs, api-reference, knowledge-base, wellness, lifestyle, boxing-gym, combat-sports, price-comparison, insurance-comparison, product-review, buyer-guide, independent-shop, single-purpose-calculator, api-playground, regulatory-information, trade-directory, interactive-platform, tools, tool-portal, developer-tools, calculators, utility-platform, game-design, game-development, technical-reference, design-tools, dark-utility, professional-dark, social, social-network, community-platform, provocation-platform, challenge-platform, game-social, room-based-platform, content-first-social, ai-platform, ai-curated-content, ai-seeded-community, creator-platform, archetype-platform.\n'
        E'\n'
        E'If no existing tag fits well, coin a new one using the same style. The library-growth audit flags unmatched tags for review.\n'
    );

    IF step1_prompt = old_prompt THEN
        RAISE EXCEPTION
            '008 step 1: regexp_replace for classification block had no effect. Check the regex against the current prompt.';
END IF;

    -- Edit 2: Add an Adoption fidelity bullet for category/industry_tags.
    -- Insert AFTER the existing suggested_style bullet, BEFORE the
    -- identity.has_existing_site bullet.
    step2_prompt := regexp_replace(
        step1_prompt,
        '(- classification\.suggested_style MUST be consistent with the adopted design_intent[^\n]*\n)',
        E'\\1'
        E'- classification.category and classification.industry_tags MUST reflect the adopted archetype. For an adopted tools/utility platform: category="interactive" and industry_tags include "interactive-platform" plus vertical tags drawn from site_archetype.industry (e.g. "game-design", "developer-tools"). For an adopted social/community platform: category="social" and industry_tags include "social-network" plus specific modality tags the archetype describes (e.g. "provocation-platform", "ai-seeded-community"). Prefer tags that intersect with existing library layouts before coining new ones.\n'
    );

    IF step2_prompt = step1_prompt THEN
        RAISE EXCEPTION
            '008 step 2: regexp_replace for adoption-fidelity bullet had no effect. The suggested_style bullet from 006 was not found.';
END IF;

    new_prompt := step2_prompt;

    -- Post-state sanity: both new field markers must be present.
    IF position('"category":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '008: post-edit prompt is missing "category": marker.';
END IF;
    IF position('"industry_tags":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '008: post-edit prompt is missing "industry_tags": marker.';
END IF;

    -- Apply.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb(new_prompt)
                     )
WHERE type = 'domain-research-classifier'
  AND is_active = true;

RAISE NOTICE '008: classifier prompt updated. Old length % chars; new length % chars.',
        length(old_prompt), length(new_prompt);
END
$migration$;


-- ---------------------------------------------------------------
-- 3. Verify — prompt length grew, new markers present.
-- ---------------------------------------------------------------

SELECT
    'AFTER' AS phase,
    type,
    length(default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template') AS prompt_len,
    CASE WHEN default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' LIKE '%"category":%'
    THEN 'YES' ELSE 'NO' END AS has_category_field,
    CASE WHEN default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' LIKE '%"industry_tags":%'
         THEN 'YES' ELSE 'NO' END AS has_industry_tags_field,
    CASE WHEN default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' LIKE '%classification.category and classification.industry_tags MUST reflect%'
         THEN 'YES' ELSE 'NO' END AS has_adoption_fidelity_bullet,
    CASE WHEN default_config->'workflow'->'steps'->'classify_and_extract'->'config'->>'prompt_template' LIKE '%Adoption Reference%'
         THEN 'YES' ELSE 'NO' END AS still_has_006_adoption_block
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;


-- input_fields unchanged? (The prompt relies on site_specs, search_results,
-- scraped_data — none of which should need changing for this edit.)
SELECT
    type,
    default_config->'workflow'->'steps'->'classify_and_extract'->'config'->'input_fields' AS input_fields
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

COMMIT;


-- ----------------------------------------------------------------------------
-- After applying: restart the chassis pods so they pick up the updated
-- agent_definition. The chassis caches agent configs in memory.
--
--   kubectl -n ai-persona-system rollout restart deployment/<chassis-deployment>
--
-- To test: queue a classifier re-run for any site and check the resulting
-- classification spec contains category and industry_tags fields:
--
--   SELECT data->>'category', data->'industry_tags'
--   FROM site_specs
--   WHERE aspect='classification' AND is_current=true
--   ORDER BY created_at DESC LIMIT 3;
--
-- The gamesdesign.co.uk remediation to exercise the new matcher path is
-- a separate deliberate step (see next migration/remediation SQL) —
-- invalidate resolved_composition AND queue needs_composition AND queue a
-- fresh needs_domain_research so the classifier re-runs with 008's prompt
-- and emits the fields the matcher expects. Doing those together avoids
-- the "spec superseded but install not re-queued" failure mode captured
-- in doc 028.
-- ----------------------------------------------------------------------------

-- 008_classifier_dynamic_taxonomy.sql
--
-- Wire the classifier to emit category + industry_tags on the classification
-- spec, using a DYNAMIC taxonomy read at run time from the layouts library
-- rather than hardcoded enum and tag lists in the prompt.
--
-- Three coordinated changes in one transaction:
--   1. Insert a new workflow step `read_layout_taxonomy` between the
--      existing `read_site_specs` step and `classify_and_extract`.
--   2. Update `read_site_specs.next_step` from `classify_and_extract`
--      to `read_layout_taxonomy`.
--   3. Update the prompt template to (a) declare `category` and
--      `industry_tags` on the classification JSON output, (b) render
--      the current category set and tag list from
--      `{{.layout_taxonomy.categories}}` and `{{.layout_taxonomy.industry_tags}}`
--      instead of hardcoding them, and (c) add an adoption-fidelity bullet
--      that ties the new tag fields to the adopted archetype.
--
-- DEPENDENCY: this migration requires the chassis to have a registered
-- action handler for `read_layout_taxonomy` (see
-- read_layout_taxonomy_action.go). If migrations land before the Go action
-- is deployed, the first classifier run after 008 will fail fast with an
-- unknown-action error — fail-loud is intentional; don't apply 008 until
-- the Go action is in the image the chassis is running.
--
-- ----------------------------------------------------------------------------
-- Rollback (undo this migration):
--   UPDATE agent_definitions
--   SET default_config = (
--       SELECT default_config FROM agent_definitions_backup_20260423
--       WHERE type = 'domain-research-classifier'
--   )
--   WHERE type = 'domain-research-classifier' AND is_active = true;
--
-- Then restart the chassis so pods pick up the reverted config:
--   kubectl -n ai-persona-system rollout restart deployment/<chassis-deployment>
-- ----------------------------------------------------------------------------

BEGIN;

-- ---------------------------------------------------------------
-- 1. Back up the current (post-006) state before modifying.
-- ---------------------------------------------------------------

DROP TABLE IF EXISTS agent_definitions_backup_20260423;
CREATE TABLE agent_definitions_backup_20260423 AS
SELECT * FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

SELECT 'BACKUP' AS phase, COUNT(*) AS rows_backed_up
FROM agent_definitions_backup_20260423;


-- ---------------------------------------------------------------
-- 2. Workflow + prompt changes inside a guarded DO block.
-- ---------------------------------------------------------------

DO $migration$
DECLARE
cfg jsonb;
    current_next_step text;
    old_prompt text;
    step1_prompt text;
    step2_prompt text;
    new_prompt text;
    prompt_marker text := '"suggested_style": "professional-dark|modern-light|bold-creative|etc"';
    taxonomy_marker text := '{{.layout_taxonomy.categories';
BEGIN
    -- --- Load current config --------------------------------------------------
SELECT default_config
INTO cfg
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

IF cfg IS NULL THEN
        RAISE EXCEPTION '008: classifier agent_definition not found or inactive.';
END IF;

    -- --- Pre-state checks -----------------------------------------------------
    -- read_site_specs must currently point at classify_and_extract.
    current_next_step := cfg #>> '{workflow,steps,read_site_specs,next_step}';
    IF current_next_step IS NULL THEN
        RAISE EXCEPTION '008: read_site_specs step not found in classifier workflow.';
END IF;
    IF current_next_step <> 'classify_and_extract' THEN
        RAISE EXCEPTION
            '008: expected read_site_specs.next_step = classify_and_extract, found %. Aborting — workflow is not in the state we can safely patch.',
            current_next_step;
END IF;

    -- The new step name must not already exist — if it does, something
    -- else has applied this (or a similar) change and we should not
    -- overwrite blind.
    IF cfg #> '{workflow,steps,read_layout_taxonomy}' IS NOT NULL THEN
        RAISE EXCEPTION
            '008: read_layout_taxonomy step already exists in classifier workflow. Aborting to avoid overwriting an existing definition.';
END IF;

    -- Prompt must contain the 006 marker.
    old_prompt := cfg #>> '{workflow,steps,classify_and_extract,config,prompt_template}';
    IF old_prompt IS NULL THEN
        RAISE EXCEPTION '008: classify_and_extract.config.prompt_template not found.';
END IF;
    IF position(prompt_marker IN old_prompt) = 0 THEN
        RAISE EXCEPTION
            '008: prompt is not in expected 006 state (marker not found: %). Aborting.',
            prompt_marker;
END IF;

    -- --- Workflow edit: insert the new step -----------------------------------
    cfg := jsonb_set(
        cfg,
        '{workflow,steps,read_layout_taxonomy}',
        jsonb_build_object(
            'action',       'read_layout_taxonomy',
            'config',       '{}'::jsonb,
            'next_step',    'classify_and_extract',
            'description',  'Load current categories and industry_tags from the layouts library so the prompt renders with the live taxonomy',
            'output_field', 'layout_taxonomy'
        ),
        true
    );

    -- --- Workflow edit: repoint read_site_specs's next_step -------------------
    cfg := jsonb_set(
        cfg,
        '{workflow,steps,read_site_specs,next_step}',
        to_jsonb('read_layout_taxonomy'::text),
        false
    );

    -- --- Prompt edit 1: replace classification schema + guidance --------------
    -- Match from `2. "classification"` through the closing `}\n` of its
    -- JSON schema block. The block contains no nested `{`/`}` so the
    -- non-greedy `[\s\S]*?` stops at the first `}\n`.
    step1_prompt := regexp_replace(
        old_prompt,
        '2\. "classification"[\s\S]*?\n\}\n',
        E'2. "classification" — what kind of site to build:\n'
        E'{\n'
        E'  "site_type": "brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social",\n'
        E'  "category": "<pick one from the Current library categories listed below>",\n'
        E'  "industry_tags": ["4 to 10 specific tags — see Current library tags below"],\n'
        E'  "confidence": 0.0-1.0,\n'
        E'  "reasoning": "Why this site type fits",\n'
        E'  "recommended_builder": "pageflow-builder",\n'
        E'  "page_count_estimate": 5,\n'
        E'  "detected_signals": ["signal1", "signal2"],\n'
        E'  "tone_suggestion": "professional|friendly|bold|technical|editorial|game-like|energetic",\n'
        E'  "suggested_style": "professional-dark|modern-light|bold-creative|etc"\n'
        E'}\n'
        E'\n'
        E'category and industry_tags are used by the layout matcher (site-design-planner) to pick a CSS layout from the library. The matcher scores layouts by how many of these tags overlap with each layout''s own industry_tags. Both fields must be populated.\n'
        E'\n'
        E'### category — the highest-level shape of the site. Pick ONE.\n'
        E'\n'
        E'Current library categories (read at this classifier run):\n'
        E'{{.layout_taxonomy.categories | toJSON}}\n'
        E'\n'
        E'Category meanings (stable vocabulary):\n'
        E'- brochure — traditional marketing/company site (services, about, contact)\n'
        E'- interactive — tool/utility platform, calculators, simulators, practitioner tools\n'
        E'- social — community platform, forum, social network, AI-seeded discussion\n'
        E'- editorial — news, magazine, blog, opinion, long-form content\n'
        E'- ecommerce — storefront, product catalogue, checkout\n'
        E'- portfolio — work showcase for a creative, studio, or individual\n'
        E'- hub — vertical information hub, regulatory resource, trade directory (non-commercial)\n'
        E'- docs — technical documentation, API reference, knowledge base\n'
        E'\n'
        E'If the library has a category not documented above, it was added after this documentation was last updated. Prefer a documented meaning when one fits; fall back to a library-current category when none of the documented meanings match.\n'
        E'\n'
        E'### industry_tags — 4 to 10 specific tags describing this site. Include BOTH:\n'
        E'- site-shape tags (e.g. "interactive-platform", "tool-portal", "social-network", "provocation-platform", "magazine-grid", "affiliate-hub", "docs-sidebar")\n'
        E'- vertical/domain tags (e.g. "game-design", "veterinary", "legal", "fitness", "fintech", "wellness")\n'
        E'\n'
        E'Use hyphenated lowercase. Prefer specific over generic: "game-design" not "games", "ai-platform" not "ai".\n'
        E'\n'
        E'Current library tags (match these when they describe this site — each match is an overlap point for the matcher):\n'
        E'{{.layout_taxonomy.industry_tags | toJSON}}\n'
        E'\n'
        E'The library currently has {{.layout_taxonomy.layout_count}} active layouts. If no existing tag fits this site well, coin a new one using the same style — an unmatched tag will trigger a library-growth review work item rather than silently fail.\n'
    );

    IF step1_prompt = old_prompt THEN
        RAISE EXCEPTION
            '008 prompt edit 1: regexp_replace for classification block had no effect. Check the regex against the current prompt.';
END IF;

    -- --- Prompt edit 2: add adoption-fidelity bullet for the new tag fields --
    -- Insert AFTER the existing suggested_style bullet.
    step2_prompt := regexp_replace(
        step1_prompt,
        '(- classification\.suggested_style MUST be consistent with the adopted design_intent[^\n]*\n)',
        E'\\1'
        E'- classification.category and classification.industry_tags MUST reflect the adopted archetype. For an adopted tools/utility platform: category="interactive" and industry_tags include "interactive-platform" plus vertical tags drawn from site_archetype.industry (e.g. "game-design", "developer-tools"). For an adopted social/community platform: category="social" and industry_tags include "social-network" plus specific modality tags the archetype describes (e.g. "provocation-platform", "ai-seeded-community"). Prefer tags that appear in the Current library tags list so the matcher can find a close layout rather than triggering a library-growth fallback.\n'
    );

    IF step2_prompt = step1_prompt THEN
        RAISE EXCEPTION
            '008 prompt edit 2: regexp_replace for adoption-fidelity bullet had no effect. The suggested_style bullet from 006 was not found.';
END IF;

    new_prompt := step2_prompt;

    -- --- Post-state checks ----------------------------------------------------
    IF position('"category":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '008: post-edit prompt is missing "category": marker.';
END IF;
    IF position('"industry_tags":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '008: post-edit prompt is missing "industry_tags": marker.';
END IF;
    IF position(taxonomy_marker IN new_prompt) = 0 THEN
        RAISE EXCEPTION
            '008: post-edit prompt is missing the dynamic taxonomy marker (%). The hardcoded replacement would have worked but the dynamic version did not.',
            taxonomy_marker;
END IF;

    -- --- Apply ----------------------------------------------------------------
    cfg := jsonb_set(
        cfg,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb(new_prompt),
        false
    );

UPDATE agent_definitions
SET default_config = cfg
WHERE type = 'domain-research-classifier'
  AND is_active = true;

RAISE NOTICE '008: classifier updated. Old prompt length % chars; new prompt length % chars.',
        length(old_prompt), length(new_prompt);
END
$migration$;


-- ---------------------------------------------------------------
-- 3. Verify — workflow wiring and prompt markers.
-- ---------------------------------------------------------------

SELECT
    'AFTER - workflow wiring' AS phase,
    default_config #>> '{workflow,steps,read_site_specs,next_step}'           AS read_site_specs_next_step,
    default_config #>> '{workflow,steps,read_layout_taxonomy,action}'         AS new_step_action,
    default_config #>> '{workflow,steps,read_layout_taxonomy,next_step}'      AS new_step_next_step,
    default_config #>> '{workflow,steps,read_layout_taxonomy,output_field}'   AS new_step_output_field
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

SELECT
    'AFTER - prompt markers' AS phase,
    length(default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}') AS prompt_len,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
        LIKE '%"category":%'                                  THEN 'YES' ELSE 'NO' END AS has_category_field,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%"industry_tags":%'                             THEN 'YES' ELSE 'NO' END AS has_industry_tags_field,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%{{.layout_taxonomy.categories%'                THEN 'YES' ELSE 'NO' END AS has_dynamic_categories,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%{{.layout_taxonomy.industry_tags%'             THEN 'YES' ELSE 'NO' END AS has_dynamic_tags,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%Adoption Reference%'                           THEN 'YES' ELSE 'NO' END AS still_has_006_adoption_block,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%classification.category and classification.industry_tags MUST reflect%'
                                                                       THEN 'YES' ELSE 'NO' END AS has_adoption_fidelity_bullet
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

COMMIT;


-- ----------------------------------------------------------------------------
-- After applying
-- ----------------------------------------------------------------------------
-- 1. Verify the Go action is present in the running chassis image. If not
--    yet deployed, hold here — do not queue classifier runs until the pod
--    image contains read_layout_taxonomy_action.go. The first classifier
--    run without the action will fail fast with "unknown action
--    'read_layout_taxonomy'" in orchestration_states.error.
--
-- 2. Restart the chassis so pods pick up the updated agent_definition:
--    kubectl -n ai-persona-system rollout restart deployment/<chassis-deployment>
--
-- 3. Test the action renders sensibly by triggering one classifier run
--    and checking the llm_call_log prompt_rendered for a site:
--
--    SELECT prompt_rendered
--    FROM llm_call_log
--    WHERE agent_type = 'domain-research-classifier'
--      AND step_name  = 'classify_and_extract'
--      AND created_at > NOW() - INTERVAL '10 minutes'
--    ORDER BY created_at DESC
--    LIMIT 1;
--
--    The rendered prompt should contain JSON arrays inline where the
--    taxonomy references are, e.g.
--      Current library categories (read at this classifier run):
--      ["brochure","interactive","social"]
--    etc. Empty arrays would indicate the action ran but the library is
--    empty or filtered everything out — check layouts.is_active values.
--
-- 4. Verify the resulting classification spec has the new fields:
--    SELECT data->>'category', data->'industry_tags'
--    FROM site_specs
--    WHERE aspect = 'classification' AND is_current = true
--    ORDER BY created_at DESC LIMIT 3;
--
-- The gamesdesign.co.uk remediation (invalidate resolved_composition AND
-- queue needs_composition AND queue a fresh classifier re-run so it picks
-- up 008's prompt) is a separate deliberate step — written as its own
-- remediation SQL once 008 is confirmed working. Doing all three re-queues
-- together avoids the "spec superseded but install not re-queued" failure
-- mode captured in doc 028.
-- ----------------------------------------------------------------------------

-- fix

-- 008_classifier_dynamic_taxonomy.sql
--
-- Wire the classifier to emit category + industry_tags on the classification
-- spec, using a DYNAMIC taxonomy read at run time from the layouts library
-- rather than hardcoded enum and tag lists in the prompt.
--
-- Three coordinated changes in one transaction:
--   1. Insert a new workflow step `read_layout_taxonomy` between the
--      existing `read_site_specs` step and `classify_and_extract`.
--   2. Update `read_site_specs.next_step` from `classify_and_extract`
--      to `read_layout_taxonomy`.
--   3. Update the prompt template to (a) declare `category` and
--      `industry_tags` on the classification JSON output, (b) render
--      the current category set and tag list from
--      `{{.layout_taxonomy.categories}}` and `{{.layout_taxonomy.industry_tags}}`
--      instead of hardcoding them, and (c) add an adoption-fidelity bullet
--      that ties the new tag fields to the adopted archetype.
--
-- DEPENDENCY: this migration requires the chassis to have a registered
-- action handler for `read_layout_taxonomy` (see
-- read_layout_taxonomy_action.go). If migrations land before the Go action
-- is deployed, the first classifier run after 008 will fail fast with an
-- unknown-action error — fail-loud is intentional; don't apply 008 until
-- the Go action is in the image the chassis is running.
--
-- REWRITE NOTE: an earlier version of this file used multi-line E-string
-- concatenation (E'...\n' on separate lines). That works in regular SQL
-- but fails in PL/pgSQL's expression parser, which doesn't perform the
-- adjacent-string-literal concatenation that plain SQL does. This version
-- uses nested dollar-quoted strings with real embedded newlines instead.
--
-- ----------------------------------------------------------------------------
-- Rollback (undo this migration):
--   UPDATE agent_definitions
--   SET default_config = (
--       SELECT default_config FROM agent_definitions_backup_20260423
--       WHERE type = 'domain-research-classifier'
--   )
--   WHERE type = 'domain-research-classifier' AND is_active = true;
--
-- Then restart the chassis so pods pick up the reverted config:
--   kubectl -n ai-persona-system rollout restart deployment/<chassis-deployment>
-- ----------------------------------------------------------------------------

BEGIN;

-- ---------------------------------------------------------------
-- 1. Back up the current (post-006) state before modifying.
-- ---------------------------------------------------------------

DROP TABLE IF EXISTS agent_definitions_backup_20260423;
CREATE TABLE agent_definitions_backup_20260423 AS
SELECT * FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

SELECT 'BACKUP' AS phase, COUNT(*) AS rows_backed_up
FROM agent_definitions_backup_20260423;


-- ---------------------------------------------------------------
-- 2. Workflow + prompt changes inside a guarded DO block.
-- ---------------------------------------------------------------

DO $migration$
DECLARE
cfg jsonb;
    current_next_step text;
    old_prompt text;
    step1_prompt text;
    step2_prompt text;
    new_prompt text;
    prompt_marker text := '"suggested_style": "professional-dark|modern-light|bold-creative|etc"';
    taxonomy_marker text := '{{.layout_taxonomy.categories';

    -- Replacement text for the classification JSON schema block.
    -- Dollar-quoted with actual embedded newlines — PL/pgSQL does NOT
    -- concatenate adjacent string literals the way plain SQL does, so
    -- this has to be a single literal.
    new_classification_block text := $newclass$2. "classification" — what kind of site to build:
{
  "site_type": "brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social",
  "category": "<pick one from the Current library categories listed below>",
  "industry_tags": ["4 to 10 specific tags — see Current library tags below"],
  "confidence": 0.0-1.0,
  "reasoning": "Why this site type fits",
  "recommended_builder": "pageflow-builder",
  "page_count_estimate": 5,
  "detected_signals": ["signal1", "signal2"],
  "tone_suggestion": "professional|friendly|bold|technical|editorial|game-like|energetic",
  "suggested_style": "professional-dark|modern-light|bold-creative|etc"
}

category and industry_tags are used by the layout matcher (site-design-planner) to pick a CSS layout from the library. The matcher scores layouts by how many of these tags overlap with each layout's own industry_tags. Both fields must be populated.

### category — the highest-level shape of the site. Pick ONE.

Current library categories (read at this classifier run):
{{.layout_taxonomy.categories | toJSON}}

Category meanings (stable vocabulary):
- brochure — traditional marketing/company site (services, about, contact)
- interactive — tool/utility platform, calculators, simulators, practitioner tools
- social — community platform, forum, social network, AI-seeded discussion
- editorial — news, magazine, blog, opinion, long-form content
- ecommerce — storefront, product catalogue, checkout
- portfolio — work showcase for a creative, studio, or individual
- hub — vertical information hub, regulatory resource, trade directory (non-commercial)
- docs — technical documentation, API reference, knowledge base

If the library has a category not documented above, it was added after this documentation was last updated. Prefer a documented meaning when one fits; fall back to a library-current category when none of the documented meanings match.

### industry_tags — 4 to 10 specific tags describing this site. Include BOTH:
- site-shape tags (e.g. "interactive-platform", "tool-portal", "social-network", "provocation-platform", "magazine-grid", "affiliate-hub", "docs-sidebar")
- vertical/domain tags (e.g. "game-design", "veterinary", "legal", "fitness", "fintech", "wellness")

Use hyphenated lowercase. Prefer specific over generic: "game-design" not "games", "ai-platform" not "ai".

Current library tags (match these when they describe this site — each match is an overlap point for the matcher):
{{.layout_taxonomy.industry_tags | toJSON}}

The library currently has {{.layout_taxonomy.layout_count}} active layouts. If no existing tag fits this site well, coin a new one using the same style — an unmatched tag will trigger a library-growth review work item rather than silently fail.
$newclass$;

    -- Replacement for edit 2 — prepended before the existing bullet.
    -- Uses \1 backreference inside a dollar-quoted string (which does
    -- not process \-escapes, so \1 stays literal \1 as regex needs).
    new_adoption_bullet text := E'\\1' || $bullet$- classification.category and classification.industry_tags MUST reflect the adopted archetype. For an adopted tools/utility platform: category="interactive" and industry_tags include "interactive-platform" plus vertical tags drawn from site_archetype.industry (e.g. "game-design", "developer-tools"). For an adopted social/community platform: category="social" and industry_tags include "social-network" plus specific modality tags the archetype describes (e.g. "provocation-platform", "ai-seeded-community"). Prefer tags that appear in the Current library tags list so the matcher can find a close layout rather than triggering a library-growth fallback.
$bullet$;
BEGIN
    -- --- Load current config --------------------------------------------------
    SELECT default_config
      INTO cfg
      FROM agent_definitions
     WHERE type = 'domain-research-classifier'
       AND is_active = true;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '008: classifier agent_definition not found or inactive.';
    END IF;

    -- --- Pre-state checks -----------------------------------------------------
    current_next_step := cfg #>> '{workflow,steps,read_site_specs,next_step}';
    IF current_next_step IS NULL THEN
        RAISE EXCEPTION '008: read_site_specs step not found in classifier workflow.';
    END IF;
    IF current_next_step <> 'classify_and_extract' THEN
        RAISE EXCEPTION
            '008: expected read_site_specs.next_step = classify_and_extract, found %. Aborting — workflow is not in the state we can safely patch.',
            current_next_step;
    END IF;

    IF cfg #> '{workflow,steps,read_layout_taxonomy}' IS NOT NULL THEN
        RAISE EXCEPTION
            '008: read_layout_taxonomy step already exists in classifier workflow. Aborting to avoid overwriting an existing definition.';
    END IF;

    old_prompt := cfg #>> '{workflow,steps,classify_and_extract,config,prompt_template}';
    IF old_prompt IS NULL THEN
        RAISE EXCEPTION '008: classify_and_extract.config.prompt_template not found.';
    END IF;
    IF position(prompt_marker IN old_prompt) = 0 THEN
        RAISE EXCEPTION
            '008: prompt is not in expected 006 state (marker not found: %). Aborting.',
            prompt_marker;
    END IF;

    -- --- Workflow edit: insert the new step -----------------------------------
    cfg := jsonb_set(
        cfg,
        '{workflow,steps,read_layout_taxonomy}',
        jsonb_build_object(
            'action',       'read_layout_taxonomy',
            'config',       '{}'::jsonb,
            'next_step',    'classify_and_extract',
            'description',  'Load current categories and industry_tags from the layouts library so the prompt renders with the live taxonomy',
            'output_field', 'layout_taxonomy'
        ),
        true
    );

    -- --- Workflow edit: repoint read_site_specs's next_step -------------------
    cfg := jsonb_set(
        cfg,
        '{workflow,steps,read_site_specs,next_step}',
        to_jsonb('read_layout_taxonomy'::text),
        false
    );

    -- --- Prompt edit 1: replace classification schema + guidance --------------
    -- Pattern matches from `2. "classification"` through the closing `}\n`
    -- of its JSON schema block. The block contains no nested `{`/`}` so the
    -- non-greedy `[\s\S]*?` stops at the first `}\n`.
    step1_prompt := regexp_replace(
        old_prompt,
        '2\. "classification"[\s\S]*?\n\}\n',
        new_classification_block
    );

    IF step1_prompt = old_prompt THEN
        RAISE EXCEPTION
            '008 prompt edit 1: regexp_replace for classification block had no effect. Check the regex against the current prompt.';
END IF;

    -- --- Prompt edit 2: add adoption-fidelity bullet for the new tag fields --
    -- Capture the existing suggested_style bullet in group 1 and prepend
    -- the new bullet so ordering reads: suggested_style bullet, THEN new bullet.
    -- Wait — we actually want the new bullet AFTER the existing, so we
    -- reference \1 first then append the new text. See new_adoption_bullet above.
    step2_prompt := regexp_replace(
        step1_prompt,
        '(- classification\.suggested_style MUST be consistent with the adopted design_intent[^\n]*\n)',
        new_adoption_bullet
    );

    IF step2_prompt = step1_prompt THEN
        RAISE EXCEPTION
            '008 prompt edit 2: regexp_replace for adoption-fidelity bullet had no effect. The suggested_style bullet from 006 was not found.';
END IF;

    new_prompt := step2_prompt;

    -- --- Post-state checks ----------------------------------------------------
    IF position('"category":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '008: post-edit prompt is missing "category": marker.';
END IF;
    IF position('"industry_tags":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '008: post-edit prompt is missing "industry_tags": marker.';
END IF;
    IF position(taxonomy_marker IN new_prompt) = 0 THEN
        RAISE EXCEPTION
            '008: post-edit prompt is missing the dynamic taxonomy marker (%). The edit should have inserted {{.layout_taxonomy.categories ...}} but did not.',
            taxonomy_marker;
END IF;

    -- --- Apply ----------------------------------------------------------------
    cfg := jsonb_set(
        cfg,
        '{workflow,steps,classify_and_extract,config,prompt_template}',
        to_jsonb(new_prompt),
        false
    );

UPDATE agent_definitions
SET default_config = cfg
WHERE type = 'domain-research-classifier'
  AND is_active = true;

RAISE NOTICE '008: classifier updated. Old prompt length % chars; new prompt length % chars.',
        length(old_prompt), length(new_prompt);
END
$migration$;


-- ---------------------------------------------------------------
-- 3. Verify — workflow wiring and prompt markers.
-- ---------------------------------------------------------------

SELECT
    'AFTER - workflow wiring' AS phase,
    default_config #>> '{workflow,steps,read_site_specs,next_step}'         AS read_site_specs_next_step,
    default_config #>> '{workflow,steps,read_layout_taxonomy,action}'       AS new_step_action,
    default_config #>> '{workflow,steps,read_layout_taxonomy,next_step}'    AS new_step_next_step,
    default_config #>> '{workflow,steps,read_layout_taxonomy,output_field}' AS new_step_output_field
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

SELECT
    'AFTER - prompt markers' AS phase,
    length(default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}') AS prompt_len,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
        LIKE '%"category":%'                                  THEN 'YES' ELSE 'NO' END AS has_category_field,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%"industry_tags":%'                             THEN 'YES' ELSE 'NO' END AS has_industry_tags_field,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%{{.layout_taxonomy.categories%'                THEN 'YES' ELSE 'NO' END AS has_dynamic_categories,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%{{.layout_taxonomy.industry_tags%'             THEN 'YES' ELSE 'NO' END AS has_dynamic_tags,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%Adoption Reference%'                           THEN 'YES' ELSE 'NO' END AS still_has_006_adoption_block,
    CASE WHEN default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
                 LIKE '%classification.category and classification.industry_tags MUST reflect%'
                                                                       THEN 'YES' ELSE 'NO' END AS has_adoption_fidelity_bullet
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND is_active = true;

COMMIT;


-- ----------------------------------------------------------------------------
-- After applying
-- ----------------------------------------------------------------------------
-- 1. Verify the Go action is present in the running chassis image. If not
--    yet deployed, hold here — do not queue classifier runs until the pod
--    image contains read_layout_taxonomy_action.go. The first classifier
--    run without the action will fail fast with "unknown action
--    'read_layout_taxonomy'" in orchestration_states.error.
--
-- 2. Restart the chassis so pods pick up the updated agent_definition:
--    kubectl -n ai-persona-system rollout restart deployment/<chassis-deployment>
--
-- 3. Test the action renders sensibly by triggering one classifier run
--    and checking the llm_call_log prompt_rendered for a site:
--
--    SELECT prompt_rendered
--    FROM llm_call_log
--    WHERE agent_type = 'domain-research-classifier'
--      AND step_name  = 'classify_and_extract'
--      AND created_at > NOW() - INTERVAL '10 minutes'
--    ORDER BY created_at DESC
--    LIMIT 1;
--
--    The rendered prompt should contain JSON arrays inline where the
--    taxonomy references are, e.g.
--      Current library categories (read at this classifier run):
--      ["brochure","interactive","social"]
--    Empty arrays indicate the action ran but the library is empty or
--    filtered everything out — check layouts.is_active values.
--
-- 4. Verify the resulting classification spec has the new fields:
--    SELECT data->>'category', data->'industry_tags'
--    FROM site_specs
--    WHERE aspect = 'classification' AND is_current = true
--    ORDER BY created_at DESC LIMIT 3;
--
-- The gamesdesign.co.uk remediation (invalidate resolved_composition AND
-- queue needs_composition AND queue a fresh classifier re-run so it picks
-- up 008's prompt) is a separate deliberate step — written as its own
-- remediation SQL once 008 is confirmed working. Doing all three re-queues
-- together avoids the "spec superseded but install not re-queued" failure
-- mode captured in doc 028.
-- ----------------------------------------------------------------------------