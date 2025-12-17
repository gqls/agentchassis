UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "develop",
                "steps": {
                    "develop": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-sonnet-4-5-20250514",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "site_architecture", "site_content"],
                            "prompt_template": "You are an expert HTML/CSS developer. Create a complete, production-ready webpage based on:\n\nSite Architecture: {{.site_architecture}}\nContent: {{.site_content}}\n\nCreate a single HTML file with:\n- Embedded CSS in a <style> tag\n- Responsive design (mobile-first)\n- Modern, clean aesthetic\n- Semantic HTML5 elements\n- The color scheme and typography from the architecture\n- All content properly placed\n\nReturn ONLY the complete HTML document, no explanation. Start with <!DOCTYPE html>."
                        },
                        "output_field": "html_result",
                        "next_step": "complete",
                        "description": "Develop HTML/CSS"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return developed HTML"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'html-developer';

remove the top-level fields that are overriding the step config:
default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = default_config - 'ai_service' - 'prompt_template' - 'processing_mode'
WHERE type IN ('site-architect', 'content-creator', 'html-developer');

--

claude-haiku-4-5-20251001

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "develop",
                "steps": {
                    "develop": {
                        "action": "execute_llm_prompt",
                        "config": {
                            "ai_service": {
                                "model": "claude-sonnet-4-5-20250514",
                                "provider": "anthropic",
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            },
                            "input_fields": ["input_data", "site_architecture", "site_content"],
                            "prompt_template": "You are an expert HTML/CSS developer. Create a complete, production-ready webpage based on:\n\nSite Architecture: {{.site_architecture}}\nContent: {{.site_content}}\n\nCreate a single HTML file with:\n- Embedded CSS in a <style> tag\n- Responsive design (mobile-first)\n- Modern, clean aesthetic\n- Semantic HTML5 elements\n- The color scheme and typography from the architecture\n- All content properly placed\n\nReturn ONLY the complete HTML document, no explanation. Start with <!DOCTYPE html>."
                        },
                        "output_field": "html_result",
                        "next_step": "complete",
                        "description": "Develop HTML/CSS"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return developed HTML"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'html-developer';

UPDATE agent_definitions
SET
    input_contract = '{
        "required": ["input_data", "site_architecture", "site_content"],
        "expects": {
            "site_architecture": "object",
            "site_content": "object"
        }
    }'::jsonb,
    output_contract = '{
        "produces": "final_html",
        "format": {
            "type": "object",
            "properties": {
                "html": "string",
                "css": "string"
            },
            "description": "Developed HTML and CSS for the site"
        }
    }'::jsonb
WHERE type = 'html-developer';


-- ============================================================================
-- FIX: html-developer - Add max_tokens and improve prompt to handle large inputs
-- ============================================================================

-- Fix 1: Add explicit max_tokens and improve the prompt template
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,develop,config}',
            jsonb_build_object(
                    'ai_service', jsonb_build_object(
                    'model', 'claude-sonnet-4-5-20250514',
                    'provider', 'anthropic',
                    'api_key_env_var', 'ANTHROPIC_API_KEY',
                    'max_tokens', 16000  -- Explicit token limit for Sonnet
                                  ),
                    'input_fields', ARRAY['input_data', 'site_architecture', 'site_content'],
                    'prompt_template', 'You are an expert HTML/CSS developer creating a production-ready single-page website.

IMPORTANT: Return ONLY the complete HTML document. Start with <!DOCTYPE html> and end with </html>. No explanations, no markdown.

Site Details:
Domain: {{.input_data.domain}}
Type: {{.site_architecture.architecture_result.result}}

Create a modern, responsive HTML5 page with:
- Semantic HTML5 structure
- Embedded CSS in <style> tag (mobile-first, modern design)
- The color scheme and typography from the architecture
- All sections from the site content properly structured
- Professional, clean aesthetic
- Responsive design that works on all devices

Content to include:
{{.site_content.content_result.result}}

Return the complete HTML file now:'
            )
                     )
WHERE type = 'html-developer';

--

-- ============================================================================
-- IMPROVED: html-developer with better error handling and smaller prompts
-- ============================================================================

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
        "workflow": {
            "start_step": "develop",
            "steps": {
                "develop": {
                    "action": "execute_llm_prompt",
                    "config": {
                        "ai_service": {
                            "model": "claude-sonnet-4-5-20250514",
                            "provider": "anthropic",
                            "api_key_env_var": "ANTHROPIC_API_KEY",
                            "max_tokens": 16000
                        },
                        "input_fields": ["input_data", "site_architecture", "site_content"],
                        "prompt_template": "Create a complete, production-ready HTML5 webpage.\n\nDomain: {{.input_data.domain}}\n\nIMPORTANT INSTRUCTIONS:\n1. Return ONLY the HTML code\n2. Start with <!DOCTYPE html>\n3. Include all CSS in a <style> tag in the <head>\n4. Make it responsive (mobile-first)\n5. Use modern, clean design\n6. Include all content sections\n\nBased on the architecture and content provided, generate the complete HTML now.\n\nNOTE: The architecture and content details are in the context. Use them to inform your design choices but DO NOT reproduce large JSON blocks in your output."
                    },
                    "output_field": "html_result",
                    "next_step": "complete",
                    "description": "Develop HTML/CSS"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Return developed HTML"
                }
            }
        },
        "storage": {
            "type": "s3",
            "enabled": true,
            "auto_store": true,
            "bucket_env": "ASSETS_BUCKET",
            "public_access": true
        },
        "processing_mode": "task",
        "timeout_seconds": 300
    }'::jsonb
WHERE type = 'html-developer';



to use actions - not just llm
-- ============================================================================
-- CORRECTED: html-developer using proper HTML action architecture
-- Uses generate_html, process_html, validate_html actions instead of raw LLM
-- ============================================================================

UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "generate_html",
        "steps": {
            "generate_html": {
                "action": "generate_html",
                "description": "Generate HTML using intelligent context gathering",
                "next_step": "process_html",
                "output_field": "html_generation"
            },
            "process_html": {
                "action": "process_html",
                "description": "Process and enhance HTML (meta tags, responsive, etc.)",
                "next_step": "validate_html",
                "output_field": "html_processing"
            },
            "validate_html": {
                "action": "validate_html",
                "description": "Validate HTML structure and semantics",
                "next_step": "complete",
                "output_field": "html_validation"
            },
            "complete": {
                "action": "complete_workflow",
                "description": "Return complete HTML"
            }
        }
    },
    "storage": {
        "type": "s3",
        "enabled": true,
        "auto_store": true,
        "bucket_env": "ASSETS_BUCKET",
        "public_access": true
    },
    "processing_mode": "task",
    "timeout_seconds": 300
}'::jsonb,
    image_tag = 'v1.0.509',
    updated_at = now()
WHERE type = 'html-developer';

-- ============================================================================
-- How the HTML actions work (from html_actions.go):
-- ============================================================================
--
-- GenerateHTMLAction automatically extracts from CollectedData:
--   - analyze_domain → domain analysis
--   - architect_site → site structure
--   - create_content → content
--   - input_data → business info
--
-- ProcessHTMLAction enhances HTML:
--   - Ensures proper structure (html, head, body)
--   - Adds meta tags (charset, viewport, description)
--   - Ensures responsive design
--   - Optimizes images
--   - Minifies CSS/JS
--
-- ValidateHTMLAction checks:
--   - Required elements (html, head, body, title)
--   - Meta tags (charset, viewport)
--   - Images (src, alt)
--   - Links (href)
--   - Accessibility
--
-- ============================================================================

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps' as workflow_steps,
    image_tag
FROM agent_definitions
WHERE type = 'html-developer';

===
with passing input params properly
-- ============================================================================
-- CORRECT: html-developer workflow using input_fields configuration
-- No hardcoded field names - workflow defines what data to pass
-- ============================================================================

UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "generate_html",
        "steps": {
            "generate_html": {
                "action": "generate_html",
                "config": {
                    "input_fields": ["input_data", "site_architecture", "site_content", "domain_analysis"],
                    "generation_type": "full",
                    "max_tokens": 16000,
                     "ai_service": {
                            "model": "claude-haiku-4-5-20251001",
                            "provider": "anthropic",
                            "max_tokens": 16000,
                            "api_key_env_var": "ANTHROPIC_API_KEY"
                          }
                },
                "description": "Generate HTML from configurable input fields",
                "next_step": "process_html",
                "output_field": "html_generation"
            },
            "process_html": {
                "action": "process_html",
                "config": {
                    "input_fields": ["html_generation"]
                },
                "description": "Process and enhance HTML",
                "next_step": "validate_html",
                "output_field": "html_processing"
            },
            "validate_html": {
                "action": "validate_html",
                "config": {
                    "input_fields": ["html_processing"]
                },
                "description": "Validate HTML structure",
                "next_step": "complete",
                "output_field": "html_validation"
            },
            "complete": {
                "action": "complete_workflow",
                "description": "Return complete HTML"
            }
        }
    },
    "storage": {
        "type": "s3",
        "enabled": true,
        "auto_store": true,
        "bucket_env": "ASSETS_BUCKET",
        "public_access": true
    },
    "processing_mode": "task",
    "timeout_seconds": 300
}'::jsonb,
    image_tag = 'v1.0.511',
    updated_at = now()
WHERE type = 'html-developer';

-- ============================================================================
-- HOW THIS WORKS (properly configurable):
-- ============================================================================
--
-- 1. The workflow specifies input_fields:
--    "input_fields": ["input_data", "site_architecture", "site_content"]
--
-- 2. The action extracts those fields from CollectedData:
--    context["input_data"] = CollectedData["input_data"]
--    context["site_architecture"] = CollectedData["site_architecture"]
--    context["site_content"] = CollectedData["site_content"]
--
-- 3. Different workflows can pass different fields:
--
--    Example A (full pipeline):
--    "input_fields": ["input_data", "domain_analysis", "site_architecture", "site_content"]
--
--    Example B (minimal):
--    "input_fields": ["input_data"]
--
--    Example C (custom names):
--    "input_fields": ["business_info", "page_structure", "page_content"]
--
-- The action works with whatever fields are specified - NO HARDCODING!
--
-- ============================================================================

-- Example: Different workflow with different field names
-- This shows the flexibility of using input_fields

-- CREATE OR REPLACE FUNCTION example_custom_workflow() RETURNS void AS $$
-- BEGIN
--   -- Imagine this workflow has different step names:
--   -- - "analyze_business" instead of "analyze_domain"
--   -- - "design_layout" instead of "architect_site"
--   -- - "write_copy" instead of "create_content"
--
--   -- The html-developer can still work because we configure input_fields:
--   INSERT INTO agent_definitions (type, default_config) VALUES (
--     'custom-html-developer',
--     '{
--       "workflow": {
--         "steps": {
--           "generate_html": {
--             "action": "generate_html",
--             "config": {
--               "input_fields": ["input_data", "analyze_business", "design_layout", "write_copy"]
--             }
--           }
--         }
--       }
--     }'::jsonb
--   );
-- END;
-- $$ LANGUAGE plpgsql;

-- Verify the configuration
SELECT
    type,
    default_config->'workflow'->'steps'->'generate_html'->'config'->>'input_fields' as input_fields_config,
    default_config->'workflow'->'steps'->'generate_html'->'config'->>'generation_type' as generation_type,
    image_tag
FROM agent_definitions
WHERE type = 'html-developer';

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "generate_html",
                "steps": {
                    "generate_html": {
                        "action": "generate_html",
                        "config": {
                            "input_fields": [
                                "input_data",
                                "site_architecture",
                                "site_content",
                                "domain_analysis"
                            ],
                            "generation_type": "full",
                            "max_tokens": 16000,
                            "ai_service": {
                                "model": "claude-haiku-4-5-20251001",
                                "provider": "anthropic",
                                "max_tokens": 16000,
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            }
                        },
                        "description": "Generate HTML from configurable input fields",
                        "next_step": "process_html",
                        "output_field": "html_generation"
                    },
                    "process_html": {
                        "action": "process_html",
                        "config": {
                            "input_fields": ["html_generation"]
                        },
                        "description": "Process and enhance HTML",
                        "next_step": "validate_html",
                        "output_field": "html_processing"
                    },
                    "validate_html": {
                        "action": "validate_html",
                        "config": {
                            "input_fields": ["html_processing"]
                        },
                        "description": "Validate HTML structure",
                        "next_step": "complete",
                        "output_field": "html_validation"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return complete HTML"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'html-developer';


--

UPDATE agent_definitions
SET default_workflow = jsonb_set(
        default_workflow,
        '{workflow,steps,generate_html,config,input_fields}',
        '["input_data", "page_content", "site_content", "current_page",
           "site_architecture", "domain_analysis"]'::jsonb
                       )
WHERE agent_type = 'html-developer';

-- Update html-developer workflow to accept both multipage and single-page contexts
-- Adds page_content and current_page to input_fields so it works in loop contexts

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "generate_html",
                "steps": {
                    "generate_html": {
                        "action": "generate_html",
                        "config": {
                            "input_fields": [
                                "input_data",
                                "page_content",
                                "current_page",
                                "site_architecture",
                                "site_content",
                                "domain_analysis"
                            ],
                            "generation_type": "full",
                            "ai_service": {
                                "model": "claude-haiku-4-5-20251001",
                                "provider": "anthropic",
                                "max_tokens": 16000,
                                "api_key_env_var": "ANTHROPIC_API_KEY"
                            }
                        },
                        "description": "Generate HTML from configurable input fields",
                        "next_step": "process_html",
                        "output_field": "html_generation"
                    },
                    "process_html": {
                        "action": "process_html",
                        "config": {
                            "input_fields": ["html_generation"]
                        },
                        "description": "Process and enhance HTML",
                        "next_step": "validate_html",
                        "output_field": "html_processing"
                    },
                    "validate_html": {
                        "action": "validate_html",
                        "config": {
                            "input_fields": ["html_processing"]
                        },
                        "description": "Validate HTML structure",
                        "next_step": "complete",
                        "output_field": "html_validation"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Return complete HTML"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'html-developer';

-- Verify the update
SELECT
    type,
    jsonb_pretty(default_config->'workflow'->'steps'->'generate_html'->'config'->'input_fields') as input_fields
FROM agent_definitions
WHERE type = 'html-developer';


---

-- ============================================================================
-- Add output_type to Agent Configs
-- This tells ai_actions.go whether to append JSON output instructions
-- ============================================================================

-- 6. HTML DEVELOPER - Outputs HTML (NOT JSON)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_html,config,output_type}',
        '"html"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'html-developer'
  AND is_active = true;

