-- ============================================================================
-- EXAMPLE: 20-Page Website Builder with Batched Generation
-- Demonstrates how to handle large multi-page sites without hitting limits
-- ============================================================================

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
    version
) VALUES (
             'multipage-website-builder',
             'Multi-Page Website Builder',
             'Builds large websites (20+ pages) using batched generation to avoid token limits',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "analyze_requirements",
                     "steps": {
                         "analyze_requirements": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 4000
                                 },
                                 "input_fields": ["input_data.domain", "input_data.page_list"],
                                 "prompt_template": "Analyze requirements for a {{.input_data.domain}} website with these pages: {{.input_data.page_list}}.\n\nCreate a site architecture including:\n- Navigation structure\n- Shared design elements\n- Color scheme and typography\n- Content themes for each page\n\nReturn as JSON."
                             },
                             "output_field": "site_architecture",
                             "next_step": "generate_shared_styles",
                             "description": "Analyze site requirements and create architecture"
                         },

                         "generate_shared_styles": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-haiku-4-5-20251001",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 8000
                                 },
                                 "input_fields": ["site_architecture"],
                                 "prompt_template": "Generate shared CSS for entire site based on:\n{{.site_architecture.result}}\n\nInclude:\n- CSS reset and base styles\n- Typography system\n- Color variables\n- Responsive layout utilities\n- Navigation styles\n- Footer styles\n\nReturn ONLY CSS (no HTML, no markdown)."
                             },
                             "output_field": "shared_styles",
                             "next_step": "generate_batch_1",
                             "description": "Generate shared CSS for all pages"
                         },

                         "generate_batch_1": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 16000
                                 },
                                 "input_fields": ["site_architecture", "input_data"],
                                 "prompt_template": "Generate complete HTML for pages 1-4 of {{.input_data.domain}}:\n\n1. index.html (Home)\n2. about.html (About Us)\n3. services.html (Services)\n4. team.html (Team)\n\nArchitecture: {{.site_architecture.result}}\n\nFor each page, return complete HTML from <!DOCTYPE> to </html>.\n\nRETURN AS JSON:\n{\n  \"index.html\": \"<!DOCTYPE html>...\",\n  \"about.html\": \"<!DOCTYPE html>...\",\n  \"services.html\": \"<!DOCTYPE html>...\",\n  \"team.html\": \"<!DOCTYPE html>...\"\n}\n\nIMPORTANT: DO NOT include shared CSS - it will be injected automatically."
                             },
                             "output_field": "batch_1_pages",
                             "next_step": "generate_batch_2",
                             "description": "Generate pages 1-4"
                         },

                         "generate_batch_2": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 16000
                                 },
                                 "input_fields": ["site_architecture", "input_data"],
                                 "prompt_template": "Generate complete HTML for pages 5-8:\n\n5. products.html (Products)\n6. pricing.html (Pricing)\n7. features.html (Features)\n8. testimonials.html (Testimonials)\n\nReturn as JSON map of filename to HTML."
                             },
                             "output_field": "batch_2_pages",
                             "next_step": "generate_batch_3",
                             "description": "Generate pages 5-8"
                         },

                         "generate_batch_3": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 16000
                                 },
                                 "input_fields": ["site_architecture", "input_data"],
                                 "prompt_template": "Generate complete HTML for pages 9-12:\n\n9. case-studies.html (Case Studies)\n10. blog.html (Blog)\n11. resources.html (Resources)\n12. faq.html (FAQ)\n\nReturn as JSON map."
                             },
                             "output_field": "batch_3_pages",
                             "next_step": "generate_batch_4",
                             "description": "Generate pages 9-12"
                         },

                         "generate_batch_4": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 16000
                                 },
                                 "input_fields": ["site_architecture", "input_data"],
                                 "prompt_template": "Generate complete HTML for pages 13-16:\n\n13. support.html (Support)\n14. contact.html (Contact)\n15. careers.html (Careers)\n16. partners.html (Partners)\n\nReturn as JSON map."
                             },
                             "output_field": "batch_4_pages",
                             "next_step": "generate_batch_5",
                             "description": "Generate pages 13-16"
                         },

                         "generate_batch_5": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 16000
                                 },
                                 "input_fields": ["site_architecture", "input_data"],
                                 "prompt_template": "Generate complete HTML for pages 17-20:\n\n17. press.html (Press)\n18. legal.html (Legal)\n19. privacy.html (Privacy Policy)\n20. sitemap.html (Sitemap)\n\nReturn as JSON map."
                             },
                             "output_field": "batch_5_pages",
                             "next_step": "assemble_multipage_site",
                             "description": "Generate pages 17-20"
                         },

                         "assemble_multipage_site": {
                             "action": "assemble_multipage_site",
                             "config": {
                                 "index_html_field": "batch_1_pages.index.html",
                                 "batch_fields": [
                                     "batch_1_pages",
                                     "batch_2_pages",
                                     "batch_3_pages",
                                     "batch_4_pages",
                                     "batch_5_pages"
                                 ],
                                 "shared_styles_field": "shared_styles.result",
                                 "navigation_field": "site_architecture.navigation",
                                 "stream_to_s3": true
                             },
                             "output_field": "assembled_site",
                             "next_step": "complete",
                             "description": "Assemble all pages with shared styles and navigation"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Site assembly complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600
             }'::jsonb,
             true,
             '["orchestration", "website-builder", "multi-page", "batched-generation"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.509',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             1
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       updated_at = now();

-- ============================================================================
-- Example Input Data
-- ============================================================================

-- To trigger this workflow:
-- {
--   "domain": "techcorp.com",
--   "page_list": "index, about, services, team, products, pricing, features, testimonials, case-studies, blog, resources, faq, support, contact, careers, partners, press, legal, privacy, sitemap"
-- }

-- Expected Output:
-- {
--   "assembled_site": {
--     "stored_files": {
--       "index.html": "s3://bucket/multipage-sites/orch-id/timestamp/index.html",
--       "about.html": "s3://bucket/multipage-sites/orch-id/timestamp/about.html",
--       ... (18 more files)
--     },
--     "page_count": 20,
--     "total_bytes": 1234567,
--     "mode": "streamed_to_s3"
--   }
-- }

-- ============================================================================
-- Token Budget Analysis
-- ============================================================================

-- Step 1: analyze_requirements     = 4,000 tokens
-- Step 2: generate_shared_styles   = 8,000 tokens
-- Step 3: generate_batch_1 (4 pgs) = 16,000 tokens
-- Step 4: generate_batch_2 (4 pgs) = 16,000 tokens
-- Step 5: generate_batch_3 (4 pgs) = 16,000 tokens
-- Step 6: generate_batch_4 (4 pgs) = 16,000 tokens
-- Step 7: generate_batch_5 (4 pgs) = 16,000 tokens
-- Step 8: assemble_multipage_site  = No LLM call
-- ------------------------------------------------
-- Total LLM tokens:                = 92,000 tokens
-- Total LLM calls:                 = 7 calls
-- Average per page:                = 4,600 tokens (efficient!)
--
-- Compare to naive single-call approach:
-- - Would need ~200,000 tokens (IMPOSSIBLE)
-- - Would fail or return empty
-- - No way to debug which page failed
--
-- Benefits of batched approach:
-- ✓ Stays well within limits
-- ✓ Each batch is independently debuggable
-- ✓ Can adjust batch sizes if needed
-- ✓ Streaming prevents memory issues
-- ✓ Scales to 50+ pages by adding more batches