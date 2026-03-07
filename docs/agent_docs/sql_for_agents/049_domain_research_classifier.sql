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

