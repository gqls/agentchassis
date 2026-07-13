-- 064_domain_strategist.sql
--
-- domain-strategist agent definition
-- Handler for: needs_strategy work items
-- Pipeline position: after domain-research-classifier, before build-briefing-agent
--
-- Single responsibility: determine the STRATEGY for a domain.
-- Does NOT design page architecture — that's the planner's job.
--
-- Outputs strategic guidance that the planner reads as input:
--   - What kind of site (canonical site_type)
--   - What revenue model
--   - What content strategy
--   - What kinds of pages the strategy calls for (page_type recommendations)
--   - Tone and positioning
--
-- The planner has final say on actual pages, URLs, sections, nav structure.
-- It reads the strategy and may agree, adjust, or override.
--
-- Writes to site_specs:
--   aspect "strategy" — full reasoning and recommendations
--
-- Does NOT overwrite "classification" — that's the researcher's output.
-- The planner reads both and synthesises.
--
-- Creates next work item:
--   needs_briefing → build-briefing-agent

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics,
    health_config, env_vars, version,
    delegation_preferences, agent_category, status,
    domain_tags, briefing_questionnaire,
    usage_count, is_snapshot, input_contract, output_contract
) VALUES (
             'domain-strategist',
             'Domain Strategist',
             'Determines optimal site strategy for a domain. Analyzes domain value, revenue models, competitive positioning. Outputs strategic guidance — site type, revenue model, content strategy, recommended page types. Does not design page architecture (that is the planner responsibility).',
             'specialist',
             '{
                 "workflow": {
                     "start_step": "read_specs",
                     "steps": {

                         "read_specs": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "aspect": "all"
                             },
                             "next_step": "analyze_strategy",
                             "description": "Load identity and basic classification from domain research",
                             "output_field": "site_specs"
                         },

                         "analyze_strategy": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["input_data", "site_specs"],
                                 "output_format": "json",
                                 "prompt_template": "You are a domain strategist. Your job is to determine the best website strategy for a domain name. You determine WHAT kind of site to build and WHY. You do NOT design the page architecture — a separate planner agent handles that.\n\nDomain: {{.input_data.domain}}\n\n## Research Data\n{{.site_specs}}\n\n## Your Task\n\nThink carefully about this domain. Work through each section below IN ORDER.\n\n### 1. Domain Name Analysis\n\nClassify the domain name:\n- **company_brand**: A specific business name (acmeplumbing.com, smithandjoneslaw.co.uk). The site should represent THAT business.\n- **generic_industry**: A generic industry or service term (gaswholesalers.com, londonplumbers.co.uk). Much higher value as an authority/directory — pretending to BE a single gas wholesaler wastes the domain.\n- **geographic_service**: A location + service combination (manchesterelectricians.com). Ideal for a local directory or lead generation.\n- **product_category**: A product type (standingdesks.com, bestcoffeegrinders.com). Good for reviews, comparisons, or affiliate content.\n- **ambiguous**: Could be either a brand or a generic term. Explain why and make a judgment call.\n\n### 2. Search Intent\n\nWho would search for these keywords?\n- What are they looking for?\n- Commercial intent (ready to buy/hire) or informational (researching)?\n- What are the likely high-value search terms?\n\n### 3. Revenue Model Assessment\n\nRate each model for THIS domain: strong_fit / possible / poor_fit.\n\n- **lead_generation**: Capture enquiries, sell leads to businesses in the industry\n- **affiliate**: Link to businesses with affiliate programs, earn commission\n- **display_advertising**: Build traffic via content/SEO, monetise with ads\n- **sponsored_listings**: Businesses pay for premium placement in a directory\n- **direct_business**: Domain represents an actual business, revenue from the business itself\n- **saas_tools**: Provide a useful tool, monetise via premium features or leads\n\nPick ONE primary and up to TWO secondary models.\n\n### 4. Competitive Positioning\n\nBased on the research:\n- Who currently ranks for these keywords?\n- What kind of sites are they?\n- What gap exists that this domain could fill?\n\n### 5. Site Type Recommendation\n\nChoose EXACTLY ONE from this canonical list:\n\n| Site Type | When to use |\n|-----------|-------------|\n| `brochure` | Domain represents a specific business — showcase their services/products |\n| `authority-portal` | Generic industry domain — be THE resource with directory + editorial content |\n| `local-directory` | Geographic + service domain — local service directory with listings |\n| `review-site` | Product category domain — reviews, comparisons, affiliate content |\n| `content-hub` | Informational domain — blog/magazine style, articles as primary content |\n| `landing-page` | Specific product/offer — single high-conversion page |\n| `portfolio` | Creative/agency domain — showcase of work/projects |\n\n### 6. Page Type Recommendations\n\nBased on the strategy, recommend which PAGE TYPES the planner should consider. These are recommendations, not a page plan — the planner decides the actual pages.\n\nChoose from this canonical list:\n\n| Page Type | Description |\n|-----------|-------------|\n| `content` | Standard content page (about, services, contact, etc) |\n| `index` | Home page |\n| `landing` | Conversion-focused page |\n| `entity-directory` | Searchable/filterable directory of entities |\n| `entity-page` | Individual entity profile page |\n| `tool` | Interactive tool or calculator |\n| `blog-index` | Blog/news listing page |\n| `blog-post` | Individual blog article |\n\nFor each recommended page type, explain WHY the strategy calls for it.\n\n### 7. Content & Tone\n\nWhat tone suits this site? What kind of content draws the target audience?\n\nReturn your analysis as JSON:\n```json\n{\n  \"domain_type\": \"company_brand|generic_industry|geographic_service|product_category|ambiguous\",\n  \"domain_type_reasoning\": \"why this classification\",\n  \"search_intent\": {\n    \"primary_intent\": \"commercial|informational|navigational\",\n    \"likely_searches\": [\"search term 1\", \"search term 2\"],\n    \"high_value_terms\": [\"term with commercial value\"]\n  },\n  \"revenue_models\": {\n    \"lead_generation\": {\"fit\": \"strong_fit|possible|poor_fit\", \"reasoning\": \"...\"},\n    \"affiliate\": {\"fit\": \"...\", \"reasoning\": \"...\"},\n    \"display_advertising\": {\"fit\": \"...\", \"reasoning\": \"...\"},\n    \"sponsored_listings\": {\"fit\": \"...\", \"reasoning\": \"...\"},\n    \"direct_business\": {\"fit\": \"...\", \"reasoning\": \"...\"},\n    \"saas_tools\": {\"fit\": \"...\", \"reasoning\": \"...\"},\n    \"primary_model\": \"one of the above keys\",\n    \"secondary_models\": [\"key1\"]\n  },\n  \"competitive_position\": {\n    \"current_landscape\": \"brief description\",\n    \"gap_opportunity\": \"what gap this site fills\",\n    \"defensible_moat\": \"why competitors cant easily replicate\"\n  },\n  \"site_type\": \"one of the canonical site types\",\n  \"site_type_reasoning\": \"why this site type was chosen\",\n  \"recommended_page_types\": [\n    {\"page_type\": \"entity-directory\", \"reasoning\": \"core to the authority portal strategy\"},\n    {\"page_type\": \"content\", \"reasoning\": \"about page, contact, services overview\"},\n    {\"page_type\": \"blog-index\", \"reasoning\": \"editorial content for SEO traffic\"}\n  ],\n  \"tone\": \"professional|friendly|authoritative|editorial|technical|bold\",\n  \"content_strategy\": \"what content draws visitors and keeps them coming back\",\n  \"growth_path\": \"how the site scales over time\",\n  \"value_proposition\": \"one sentence describing what this site offers visitors\"\n}\n```\n\nReturn ONLY valid JSON."
                             },
                             "next_step": "write_strategy_spec",
                             "description": "Deep reasoning about domain value, revenue model, site type, and strategic recommendations",
                             "output_field": "strategy_analysis"
                         },

                         "write_strategy_spec": {
                             "action": "write_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "spec_data": "strategy_analysis.result",
                                 "aspect": "strategy",
                                 "source": "domain-strategist",
                                 "source_agent": "domain-strategist",
                                 "source_item_id": "input_data.work_item_id"
                             },
                             "next_step": "create_next_item",
                             "description": "Persist strategy reasoning to site_specs (own aspect, does not overwrite classification)",
                             "output_field": "strategy_written"
                         },

                         "create_next_item": {
                             "action": "create_work_item",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "item_type": "needs_briefing",
                                 "handler_agent": "build-briefing-agent",
                                 "item_domain": "build",
                                 "severity": "high",
                                 "source": "domain-strategist",
                                 "summary": "Briefing needed after domain strategy complete",
                                 "item_key_prefix": "briefing",
                                 "priority": 10
                             },
                             "next_step": "complete",
                             "description": "Chain to briefing agent",
                             "output_field": "next_item_created"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["strategy_analysis", "strategy_written", "next_item_created"]
                             },
                             "description": "Strategy complete — planner will read this via site_specs"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 240
             }'::jsonb,
             true,
             '["classification", "strategy", "llm"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.811',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'analyst',
             'experimental',
             '["strategy", "classification"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": ["site_id"], "optional": ["domain", "work_item_id"], "description": "Receives site_id from dispatch loop. Reads identity and basic classification from site_specs."}'::jsonb,
             '{"produces": {"strategy_written": "site_spec for strategy aspect with site_type, revenue model, page type recommendations"}}'::jsonb
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

