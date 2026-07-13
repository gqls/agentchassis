-- ============================================================================
-- 004: Enhanced Site Classifier (v2)
-- ============================================================================
--
-- Changes site-classifier from a single Haiku LLM guess into a research-backed
-- orchestrator. The classifier now:
--   1. Prepares a research brief (Haiku — fast, cheap, handles domain-only input)
--   2. Spawns research-agent to investigate the domain via web search
--   3. Runs a Sonnet synthesis that produces both:
--      - Backward-compatible: site_type, recommended_builder, confidence
--      - New: domain_profile with business identity, tone, visual_direction,
--        image_guidance, and strategic analysis
--
-- Key design decisions:
--   - Domain-only input supported: objective may be absent or blank
--   - Not business-biased: handles content sites, news, ad-revenue models
--   - Strategic analysis: niche vs broad, SEO signals, revenue model
--   - Does NOT suggest pages or content structure (planner's job)
--   - Does NOT select style_collection (planner + webdesign-agent's job)
--   - DOES provide visual direction, tone, imagery guidance (design inputs)
--
-- The domain_profile is consumed downstream by:
--   - intake HITL review (human confirms or adjusts classification + profile)
--   - site-planner (reads profile to decide pages, components, content strategy)
--   - image-generator (reads image_guidance for hero/logo prompts)
--   - webdesign-agent (reads visual_direction for CSS decisions)
--   - page-content-writer (reads tone, identity for copy style)
--
-- Backward compatibility:
--   intake-orchestrator reads classification.response.result.site_type
--   and classification.response.result.recommended_builder — both still present.
--
-- Also updates intake-orchestrator call_classifier timeout (30s → 180s).
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. Update site-classifier: task → orchestrator with research
-- ============================================================================

UPDATE agent_definitions
SET
    description = 'Researches domain via web search, then classifies site type and builds a comprehensive domain profile. Handles domain-only input (no objective needed). Outputs backward-compatible site_type plus rich domain_profile for downstream agents.',
    agent_category = 'specialist',
    default_config = $config${
  "workflow": {
    "steps": {

      "prepare_research": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {
            "model": "claude-haiku-4-5",
            "provider": "anthropic",
            "max_tokens": 400,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": ["input_data"],
          "output_type": "json",
          "prompt_template": "You are preparing a research brief for a domain investigation. We need to understand what website should be built on this domain.\n\nDomain: {{.input_data.domain}}\nStated objective (may be blank): {{.input_data.objective}}\n\nYour job: create search queries that will help us understand this domain.\n\nConsider:\n- What does the domain name suggest? Multiple interpretations are fine.\n  e.g. 'finetuning.uk' could be AI model fine-tuning, car tuning, piano tuning\n  e.g. 'vonc.com' is ambiguous — could be an acronym, brand name, or abbreviation\n- If the domain is already live, we want to know what's there\n- If not, we want to find the best opportunities for this domain name\n- What industries or topics does this domain name fit naturally?\n\nReturn ONLY valid JSON:\n{\n  \"topic\": \"one sentence: what to investigate about this domain and what ambiguities to resolve\",\n  \"research_query\": \"3-8 word web search query\",\n  \"secondary_query\": \"alternative search angle if the domain is ambiguous\",\n  \"company\": \"inferred name from domain, or the domain itself\",\n  \"industry\": \"best guess, or 'unknown'\"\n}"
        },
        "next_step": "spawn_researcher",
        "description": "Build research brief from domain name — handles domain-only input",
        "output_field": "research_request"
      },

      "spawn_researcher": {
        "action": "spawn_agent",
        "config": {
          "role": "domain_researcher",
          "agent_type": "research-agent"
        },
        "next_step": "call_domain_research",
        "description": "Spawn research agent for domain investigation",
        "output_field": "researcher_agent"
      },

      "call_domain_research": {
        "action": "call_agent",
        "config": {
          "target_role": "domain_researcher",
          "input_mapping": {
            "current_section": "research_request.result"
          },
          "timeout_seconds": 120
        },
        "next_step": "classify_and_profile",
        "description": "Research agent investigates what this domain represents via web search",
        "output_field": "domain_research"
      },

      "classify_and_profile": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5",
            "provider": "anthropic",
            "max_tokens": 3000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": ["input_data", "domain_research"],
          "output_type": "json",
          "prompt_template": "You are a website strategist. Given a domain name and research findings, determine the best website to build on this domain and create a comprehensive profile to guide the build.\n\n## Domain\n{{.input_data.domain}}\n\n## Stated Objective\n{{.input_data.objective}}\n(If blank, you must determine the best use of this domain from research alone.)\n\n## Research Findings\n{{.domain_research.response.summary}}\n\nKey points:\n{{.domain_research.response.key_points}}\n\n## Your Task\n\nAnalyze this domain and produce a profile that downstream agents will use:\n- A SITE PLANNER will decide page structure and components from your profile\n- An IMAGE GENERATOR will create logo and hero images from your image_guidance\n- A WEB DESIGNER will create the visual design from your visual_direction\n- CONTENT WRITERS will write copy using your identity and tone guidance\n- A HUMAN will review and may adjust your recommendations\n\n## Strategic Thinking\n\nConsider these factors:\n\n1. **If the domain is already an established business/brand**: describe what it is and what the site should communicate.\n\n2. **If the domain is available/new/ambiguous**: recommend the BEST use of this domain name. Consider:\n   - What the name naturally suggests or could stand for\n   - Industry verticals where this name would be strong\n   - Revenue model: direct services, e-commerce, content/advertising, SaaS, lead generation\n   - SEO opportunity: is there a niche where this domain could rank well?\n   - Ambition: we want to build the best possible site for this domain. Not generic. Not safe. The most compelling, authoritative site this domain name could support.\n\n3. **Niche vs broad**: A tightly focused site is easier to rank and establish authority. A broader site has more long-term potential but needs more content. Recommend which approach suits this domain and why.\n\nReturn ONLY valid JSON:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"2-3 sentences: how you interpreted the domain, what the research revealed, and why you chose this direction\",\n  \"recommended_builder\": \"pageflow-builder\",\n  \"detected_industry\": \"specific industry name\",\n  \"detected_signals\": [\"signal1\", \"signal2\"],\n\n  \"identity\": {\n    \"name\": \"business or site name (may differ from domain — e.g. 'VONC' or 'Vonc Digital')\",\n    \"type\": \"specific description: 'AI model fine-tuning consultancy' not 'professional services'\",\n    \"industry\": \"primary industry or topic vertical\",\n    \"sub_industry\": \"specific niche\",\n    \"audience_primary\": \"who visits this site, with specifics\",\n    \"audience_secondary\": \"secondary audience if any\",\n    \"value_proposition\": \"core value or purpose in one sentence\",\n    \"sophistication\": \"technical|professional|casual|luxury|institutional|editorial\"\n  },\n\n  \"strategy\": {\n    \"revenue_model\": \"services|e-commerce|advertising|saas|lead-generation|subscription|affiliate\",\n    \"focus\": \"niche|broad|niche-expanding\",\n    \"focus_reasoning\": \"why this focus level suits the domain\",\n    \"competitive_position\": \"what makes this site's angle defensible or unique\",\n    \"seo_opportunity\": \"brief assessment of ranking potential in the chosen niche\",\n    \"growth_path\": \"how this site grows over time — what comes first, what comes later\"\n  },\n\n  \"tone\": {\n    \"voice\": \"description of ideal writing voice for this site\",\n    \"formality\": \"formal|professional|conversational|casual|playful|editorial\",\n    \"personality\": \"2-3 adjectives that capture the brand feel\"\n  },\n\n  \"visual_direction\": {\n    \"mood\": \"overall visual mood and feeling\",\n    \"color_signals\": [\"3-5 specific color directions with reasoning, e.g. 'deep indigo — conveys technical depth'\"],\n    \"imagery_style\": \"what imagery suits this site — be specific about style, not just subject\",\n    \"typography_feel\": \"what typography communicates the right feel\",\n    \"avoid\": [\"2-4 things to explicitly avoid in the design\"],\n    \"reference_sites\": [\"1-3 URLs of well-designed sites in this space, if known\"]\n  },\n\n  \"image_guidance\": {\n    \"hero\": \"Detailed scene/concept for hero image generation. Describe composition, colors, mood, subject matter. Example quality: 'Abstract flowing data streams converging into a refined diamond shape, deep indigo to electric blue gradient, subtle particle effects, wide cinematic aspect ratio'\",\n    \"logo\": \"Detailed description for logo generation: style, symbolism, color approach, what it should communicate at a glance\",\n    \"style_notes\": \"general image style guidance — what aesthetic all imagery should follow\"\n  }\n}\n\nRules:\n1. site_type MUST be exactly one of: landing, content, portfolio, brochure\n2. recommended_builder should be pageflow-builder unless you have strong reason otherwise\n3. Be SPECIFIC throughout. 'Modern and professional' is useless. 'Dark-themed interface with data visualization elements and monospace accents' is useful.\n4. image_guidance descriptions go DIRECTLY to an AI image generator — they must be detailed enough to produce striking, appropriate imagery. Describe scenes and concepts, not just adjectives.\n5. If research was thin, say so in reasoning, lower confidence, but still commit to a direction. A confident-but-wrong recommendation that a human can adjust is more useful than a vague hedge.\n6. Do NOT include page suggestions or site structure — the planner handles that.\n7. For open/ambiguous domains, make a BOLD recommendation. The human reviewer can adjust. Don't hedge with 'could be anything'.\n8. visual_direction.reference_sites should be real, well-known sites if possible."
        },
        "next_step": "complete",
        "description": "Synthesize research into classification and domain profile",
        "output_field": "classification_result"
      },

      "complete": {
        "action": "complete_workflow",
        "config": {
          "output": {
            "result": "classification_result.result"
          }
        },
        "description": "Return classification — result.site_type and result.recommended_builder for backward compat, plus result.identity, result.visual_direction etc. for new consumers"
      }
    },
    "start_step": "prepare_research"
  },
  "processing_mode": "orchestrator",
  "timeout_seconds": 180
}$config$::jsonb,
  updated_at = NOW()
WHERE type = 'site-classifier';

-- ============================================================================
-- 2. Update intake-orchestrator: increase call_classifier timeout
-- ============================================================================
-- The classifier now does web research (~30-60s) + LLM synthesis (~10s).
-- Old timeout was 30s which would expire mid-research.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_classifier,config,timeout_seconds}',
        '180'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'intake-orchestrator';

-- ============================================================================
-- 3. Verify
-- ============================================================================

-- Check the classifier is now an orchestrator with correct step count
SELECT type,
       default_config->>'processing_mode' as mode,
    default_config->'workflow'->>'start_step' as start_step,
    (SELECT count(*) FROM jsonb_object_keys(default_config->'workflow'->'steps')) as step_count,
    (default_config->'timeout_seconds')::text as timeout
FROM agent_definitions
WHERE type = 'site-classifier';

-- Check intake-orchestrator classifier timeout updated
SELECT type,
       default_config->'workflow'->'steps'->'call_classifier'->'config'->>'timeout_seconds' as classifier_timeout
FROM agent_definitions
WHERE type = 'intake-orchestrator';

COMMIT;

--

reset back to what it was

      -- ============================================================================
-- Revert: Undo 004 + 005 changes to existing agents
-- ============================================================================
-- Restores original state for agents modified by the incorrect 004/005 scripts.
-- Safe to run even if 004/005 were never applied (idempotent).
-- Leaves intake-orchestrator timeout at 180s (harmless).
-- ============================================================================

BEGIN;

-- 1. Revert site-classifier to original single-step Haiku task
UPDATE agent_definitions
SET
    description = 'Analyzes domain and objective to determine site type and recommend appropriate builder group',
    agent_category = NULL,
    default_config = $cls${"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return classification result"}, "classify_site": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "max_tokens": 1500, "api_key_env_var": "ANTHROPIC_API_KEY"}, "output_type": "json", "input_fields": ["input_data", "available_builders"], "output_field": "classification_result", "prompt_template": "Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.domain}}\n- Objective: {{.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Corporate sites, general business presence\n- Service providers, consultants, professional services\n- About/Services/Contact structure\n\nReturn ONLY valid JSON with this structure:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"brief explanation\",\n  \"recommended_builder\": \"builder-type\",\n  \"detected_industry\": \"industry name\",\n  \"detected_signals\": [\"signal1\", \"signal2\", ...]\n}"}, "next_step": "complete"}}, "start_step": "classify_site"}, "processing_mode": "task", "timeout_seconds": 30}$cls$::jsonb,
  updated_at = NOW()
WHERE type = 'site-classifier';

-- 2. Revert site-planner plan_site full config (prompt + input_fields)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config}',
        $pln${"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "max_tokens": 4000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "output_type": "json", "input_fields": ["input_data", "reviewed_brief", "available_components", "available_styles"], "prompt_template": "Plan a website for {{.input_data.domain}}.\n\n## Site Brief\n{{.reviewed_brief}}\n\n## Available Section Components\nThe following components are available in our component library. You MUST use ONLY these exact component names in the \"sections\" arrays:\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Task\nCreate a comprehensive site plan using ONLY the components listed above.\n\nReturn JSON in this format:\n```json\n{\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call_to_action\"]\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  }\n}\n```\n\nSTRICT RULES:\n1. ONLY use component names from the \"Available Section Components\" list above\n2. DO NOT invent new component names - if unsure, use \"hero\" for hero sections, \"features\" for feature lists, \"call_to_action\" for CTAs\n3. Use these standard mappings:\n   - For any hero/banner at page top: use \"hero\" or page-specific variants like \"contact-hero\", \"services-hero\", \"about-hero\"\n   - For feature lists: use \"features\"\n   - For service listings: use \"services-grid\"\n   - For testimonials/quotes: use \"testimonials\" or \"social_proof\"\n   - For calls to action: use \"call_to_action\"\n   - For contact forms: use \"contact-form\"\n   - For contact details: use \"contact-info\"\n   - For team sections: use \"leadership-team\"\n   - For case studies: use \"case-studies-list\"\n   - For about content: use \"about-content\"\n   - For differentiators/why-us: use \"differentiators-section\"\n\n4. Choose style_collection based on industry and tone from the brief\n5. Keep header navigation to 5-8 items maximum\n6. Always include: index (home) and contact pages\n\nIMAGE GENERATION (REQUIRED - DO NOT SKIP):\nYou MUST include needs_logo, needs_images, and image_prompts in your response.\n\n- Set needs_logo: true (always)\n- Set needs_images: true (always)\n- Provide image_prompts object with BOTH of these keys:\n  - \"logo\": A detailed 2-3 sentence prompt for logo generation. Describe the style (modern, classic, minimal), colors that match the brand, and any relevant imagery for the industry.\n  - \"hero_home\": A detailed 2-3 sentence prompt for the homepage hero background. Describe the mood (professional, energetic, calm), imagery type (abstract, photographic, geometric), and colors/atmosphere that fit the brand.\n\nExample image_prompts:\n{\n  \"logo\": \"A modern, minimal logo for a tech consulting company. Use clean geometric shapes with a navy blue and teal color palette. The design should convey innovation and trustworthiness.\",\n  \"hero_home\": \"A professional, abstract background with flowing gradients in deep navy and teal. Include subtle geometric patterns that suggest technology and connectivity. The mood should be confident and forward-thinking.\"\n}"}$pln$::jsonb
),
    updated_at = NOW()
WHERE type = 'site-planner';

-- 3. Revert pageflow-builder input_mappings
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_site_planner,config,input_mapping}',
        '{"input_data": "input_data", "site_record": "site_record", "reviewed_brief": "input_data.reviewed_brief"}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_site_design,config,input_mapping}',
        '{"domain": "site_record.domain", "site_id": "site_record.site_id"}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder';

-- 4. Revert site-work-orchestrator input_mappings
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_site_planner,config,input_mapping}',
        '{"input_data": "input_data", "site_record": "site_record", "reviewed_brief": "input_data.reviewed_brief"}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-work-orchestrator';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_site_design,config,input_mapping}',
        '{"domain": "site_record.domain", "site_id": "site_record.site_id"}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-work-orchestrator';

-- 5. Verify
SELECT type, default_config->>'processing_mode' as mode, agent_category
FROM agent_definitions WHERE type = 'site-classifier';

SELECT type, default_config->'workflow'->'steps'->'plan_site'->'config'->'input_fields' as inputs
FROM agent_definitions WHERE type = 'site-planner';

SELECT type,
       default_config->'workflow'->'steps'->'call_site_planner'->'config'->'input_mapping' as planner,
       default_config->'workflow'->'steps'->'apply_site_design'->'config'->'input_mapping' as design
FROM agent_definitions
WHERE type IN ('pageflow-builder', 'site-work-orchestrator')
ORDER BY type;

COMMIT;