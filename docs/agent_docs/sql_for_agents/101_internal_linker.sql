-- ============================================================================
-- FIX 2: Internal-linker agent definition
-- No unique constraint on type, so check existence first.
-- ============================================================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'internal-linker' AND deleted_at IS NULL) THEN
        -- Update existing
UPDATE agent_definitions
SET display_name = 'Internal Linker',
    description = 'Finds existing pages that should contextually link to an orphaned sub-page. Loads the target page and candidate pages with content samples, uses LLM to pick natural link placements, creates content_rewrite items for page-build-handler.',
    default_config = '{
                "workflow": {
                    "start_step": "ensure_site_record",
                    "processing_mode": "orchestrator",
                    "timeout_seconds": 300,
                    "steps": {
                        "ensure_site_record": {
                            "action": "ensure_site_record",
                            "config": {},
                            "next_step": "load_target_page",
                            "output_field": "site_record"
                        },
                        "load_target_page": {
                            "action": "query_database",
                            "config": {
                                "query": "SELECT p.id::text as page_id, p.name, p.url, p.title, p.page_type, LEFT(string_agg(pc.rendered_html, '' ''), 500) as content_sample FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL WHERE p.site_id = $1 AND p.name = $2 AND p.status = ''active'' GROUP BY p.id, p.name, p.url, p.title, p.page_type LIMIT 1",
                                "params": ["site_record.site_id", "input_data.spec.page_name"],
                                "output_format": "object"
                            },
                            "next_step": "check_target_found",
                            "output_field": "target_page",
                            "description": "Load the orphaned page details and content sample"
                        },
                        "check_target_found": {
                            "action": "conditional",
                            "config": {
                                "condition": "target_page.page_id != null",
                                "then_step": "load_candidate_pages",
                                "else_step": "complete_not_found"
                            }
                        },
                        "load_candidate_pages": {
                            "action": "query_database",
                            "config": {
                                "query": "SELECT p.name, p.url, p.title, p.page_type, LEFT(string_agg(pc.rendered_html, '' ''), 800) as content_sample FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL WHERE p.site_id = $1 AND p.name != $2 AND p.status = ''active'' AND p.page_type IN (''content'', ''service'', ''landing'', ''tool'') GROUP BY p.name, p.url, p.title, p.page_type HAVING COUNT(pc.id) > 0 ORDER BY p.name LIMIT 15",
                                "params": ["site_record.site_id", "input_data.spec.page_name"],
                                "output_format": "array"
                            },
                            "next_step": "check_candidates",
                            "output_field": "candidate_pages",
                            "description": "Load other site pages with content samples"
                        },
                        "check_candidates": {
                            "action": "conditional",
                            "config": {
                                "condition": "candidate_pages.count > 0",
                                "then_step": "load_specs",
                                "else_step": "complete_no_candidates"
                            }
                        },
                        "load_specs": {
                            "action": "read_site_spec",
                            "config": { "site_id": "site_record.site_id" },
                            "next_step": "plan_links",
                            "error_step": "plan_links",
                            "output_field": "site_specs"
                        },
                        "plan_links": {
                            "action": "execute_llm_prompt",
                            "config": {
                                "ai_service": {
                                    "model": "claude-haiku-4-5",
                                    "provider": "anthropic",
                                    "max_tokens": 2000,
                                    "api_key_env_var": "ANTHROPIC_API_KEY"
                                },
                                "input_fields": ["target_page", "candidate_pages", "site_record", "site_specs"],
                                "output_format": "json",
                                "prompt_template": "You are adding internal links to a website. A sub-page exists but no other page links to it, making it undiscoverable.\n\n## Target Page (needs inbound links)\nName: {{.target_page.name}}\nURL: {{.target_page.url}}\nTitle: {{.target_page.title}}\nType: {{.target_page.page_type}}\nContent preview: {{.target_page.content_sample}}\n\n## Site Context\nDomain: {{.site_record.domain}}\n{{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}}\nAudience: {{.site_specs.specs.identity.target_audience}}{{end}}\n\n## Candidate Pages (could link TO the target)\n{{range .candidate_pages}}### {{.name}} ({{.url}})\nTitle: {{.title}}\nContent: {{.content_sample}}\n\n{{end}}\n\n## Task\nPick 1-3 candidate pages where a contextual link to {{.target_page.url}} would be natural and useful for readers. Do NOT force links — if only one page is a good fit, return one.\n\nFor each link placement, explain WHERE in the page content the link fits and WHAT anchor text to use.\n\nReturn ONLY valid JSON:\n{\n  \"links\": [\n    {\n      \"source_page\": \"page-name\",\n      \"anchor_text\": \"natural anchor text\",\n      \"context\": \"Brief description of where in the page this link belongs and why it is relevant\",\n      \"guidance\": \"Rewrite instruction for the content writer: where to place the link and how to integrate it naturally\"\n    }\n  ],\n  \"reasoning\": \"Why these pages were chosen\"\n}"
                            },
                            "next_step": "check_has_links",
                            "error_step": "complete_error",
                            "output_field": "link_plan",
                            "description": "LLM decides which pages should link to the target"
                        },
                        "check_has_links": {
                            "action": "conditional",
                            "config": {
                                "condition": "link_plan.result.links != null",
                                "then_step": "create_items_loop",
                                "else_step": "complete"
                            }
                        },
                        "create_items_loop": {
                            "action": "loop",
                            "config": {
                                "items_field": "link_plan.result.links",
                                "item_variable": "current_link",
                                "max_iterations": 5,
                                "continue_on_error": true,
                                "sub_workflow": {
                                    "start_step": "create_rewrite_item",
                                    "steps": {
                                        "create_rewrite_item": {
                                            "action": "create_work_item",
                                            "config": {
                                                "site_id": "site_record.site_id",
                                                "source": "internal-linker",
                                                "item_type": "content_rewrite",
                                                "item_domain": "build",
                                                "severity": "low",
                                                "priority": 90,
                                                "handler_agent": "page-build-handler",
                                                "summary": "current_link.guidance",
                                                "item_key_prefix": "internal_link",
                                                "spec_data": {
                                                    "page_name": "current_link.source_page",
                                                    "suggestion": "current_link.guidance",
                                                    "link_target_url": "target_page.url",
                                                    "link_target_title": "target_page.title",
                                                    "anchor_text": "current_link.anchor_text",
                                                    "source": "internal-linker"
                                                }
                                            },
                                            "next_step": "done",
                                            "output_field": "item_created"
                                        },
                                        "done": {
                                            "action": "loop_complete"
                                        }
                                    }
                                }
                            },
                            "next_step": "complete",
                            "output_field": "items_created"
                        },
                        "complete": {
                            "action": "complete_workflow",
                            "config": { "output_fields": ["link_plan", "items_created"] }
                        },
                        "complete_not_found": {
                            "action": "complete_workflow",
                            "config": { "status": "skipped", "reason": "Target page not found", "output_fields": ["target_page"] }
                        },
                        "complete_no_candidates": {
                            "action": "complete_workflow",
                            "config": { "status": "skipped", "reason": "No candidate pages with content to link from", "output_fields": ["target_page"] }
                        },
                        "complete_error": {
                            "action": "complete_workflow",
                            "config": { "output_fields": ["link_plan"], "success_message": "LLM link planning failed" }
                        }
                    }
                }
            }'::jsonb,
            is_active = true,
            image_tag = 'v1.0.954',
            input_contract = '{"required": ["site_id"], "optional": ["domain", "work_item_id"], "description": "Receives site_id. Loads target page from work item spec, finds candidate pages, creates content_rewrite items."}'::jsonb,
            output_contract = '{"produces": {"link_plan": "LLM plan with source pages and link guidance", "items_created": "content_rewrite work items for page-build-handler"}}'::jsonb,
            idle_timeout_seconds = 120,
            updated_at = NOW()
WHERE type = 'internal-linker';

RAISE NOTICE 'internal-linker: UPDATED existing definition';
ELSE
        -- Insert new
        INSERT INTO agent_definitions (
            type, display_name, description, category,
            default_config, is_active, capabilities,
            image_repository, image_tag, resources, topics, health_config,
            env_vars, version, delegation_preferences,
            agent_category, status, domain_tags,
            input_contract, output_contract, idle_timeout_seconds
        ) VALUES (
            'internal-linker',
            'Internal Linker',
            'Finds existing pages that should contextually link to an orphaned sub-page. Loads the target page and candidate pages with content samples, uses LLM to pick natural link placements, creates content_rewrite items for page-build-handler.',
            'specialist',
            '{
                "workflow": {
                    "start_step": "ensure_site_record",
                    "processing_mode": "orchestrator",
                    "timeout_seconds": 300,
                    "steps": {
                        "ensure_site_record": {
                            "action": "ensure_site_record",
                            "config": {},
                            "next_step": "load_target_page",
                            "output_field": "site_record"
                        },
                        "load_target_page": {
                            "action": "query_database",
                            "config": {
                                "query": "SELECT p.id::text as page_id, p.name, p.url, p.title, p.page_type, LEFT(string_agg(pc.rendered_html, '' ''), 500) as content_sample FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL WHERE p.site_id = $1 AND p.name = $2 AND p.status = ''active'' GROUP BY p.id, p.name, p.url, p.title, p.page_type LIMIT 1",
                                "params": ["site_record.site_id", "input_data.spec.page_name"],
                                "output_format": "object"
                            },
                            "next_step": "check_target_found",
                            "output_field": "target_page",
                            "description": "Load the orphaned page details and content sample"
                        },
                        "check_target_found": {
                            "action": "conditional",
                            "config": {
                                "condition": "target_page.page_id != null",
                                "then_step": "load_candidate_pages",
                                "else_step": "complete_not_found"
                            }
                        },
                        "load_candidate_pages": {
                            "action": "query_database",
                            "config": {
                                "query": "SELECT p.name, p.url, p.title, p.page_type, LEFT(string_agg(pc.rendered_html, '' ''), 800) as content_sample FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL WHERE p.site_id = $1 AND p.name != $2 AND p.status = ''active'' AND p.page_type IN (''content'', ''service'', ''landing'', ''tool'') GROUP BY p.name, p.url, p.title, p.page_type HAVING COUNT(pc.id) > 0 ORDER BY p.name LIMIT 15",
                                "params": ["site_record.site_id", "input_data.spec.page_name"],
                                "output_format": "array"
                            },
                            "next_step": "check_candidates",
                            "output_field": "candidate_pages",
                            "description": "Load other site pages with content samples"
                        },
                        "check_candidates": {
                            "action": "conditional",
                            "config": {
                                "condition": "candidate_pages.count > 0",
                                "then_step": "load_specs",
                                "else_step": "complete_no_candidates"
                            }
                        },
                        "load_specs": {
                            "action": "read_site_spec",
                            "config": { "site_id": "site_record.site_id" },
                            "next_step": "plan_links",
                            "error_step": "plan_links",
                            "output_field": "site_specs"
                        },
                        "plan_links": {
                            "action": "execute_llm_prompt",
                            "config": {
                                "ai_service": {
                                    "model": "claude-haiku-4-5",
                                    "provider": "anthropic",
                                    "max_tokens": 2000,
                                    "api_key_env_var": "ANTHROPIC_API_KEY"
                                },
                                "input_fields": ["target_page", "candidate_pages", "site_record", "site_specs"],
                                "output_format": "json",
                                "prompt_template": "You are adding internal links to a website. A sub-page exists but no other page links to it, making it undiscoverable.\n\n## Target Page (needs inbound links)\nName: {{.target_page.name}}\nURL: {{.target_page.url}}\nTitle: {{.target_page.title}}\nType: {{.target_page.page_type}}\nContent preview: {{.target_page.content_sample}}\n\n## Site Context\nDomain: {{.site_record.domain}}\n{{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}}\nAudience: {{.site_specs.specs.identity.target_audience}}{{end}}\n\n## Candidate Pages (could link TO the target)\n{{range .candidate_pages}}### {{.name}} ({{.url}})\nTitle: {{.title}}\nContent: {{.content_sample}}\n\n{{end}}\n\n## Task\nPick 1-3 candidate pages where a contextual link to {{.target_page.url}} would be natural and useful for readers. Do NOT force links — if only one page is a good fit, return one.\n\nFor each link placement, explain WHERE in the page content the link fits and WHAT anchor text to use.\n\nReturn ONLY valid JSON:\n{\n  \"links\": [\n    {\n      \"source_page\": \"page-name\",\n      \"anchor_text\": \"natural anchor text\",\n      \"context\": \"Brief description of where in the page this link belongs and why it is relevant\",\n      \"guidance\": \"Rewrite instruction for the content writer: where to place the link and how to integrate it naturally\"\n    }\n  ],\n  \"reasoning\": \"Why these pages were chosen\"\n}"
                            },
                            "next_step": "check_has_links",
                            "error_step": "complete_error",
                            "output_field": "link_plan",
                            "description": "LLM decides which pages should link to the target"
                        },
                        "check_has_links": {
                            "action": "conditional",
                            "config": {
                                "condition": "link_plan.result.links != null",
                                "then_step": "create_items_loop",
                                "else_step": "complete"
                            }
                        },
                        "create_items_loop": {
                            "action": "loop",
                            "config": {
                                "items_field": "link_plan.result.links",
                                "item_variable": "current_link",
                                "max_iterations": 5,
                                "continue_on_error": true,
                                "sub_workflow": {
                                    "start_step": "create_rewrite_item",
                                    "steps": {
                                        "create_rewrite_item": {
                                            "action": "create_work_item",
                                            "config": {
                                                "site_id": "site_record.site_id",
                                                "source": "internal-linker",
                                                "item_type": "content_rewrite",
                                                "item_domain": "build",
                                                "severity": "low",
                                                "priority": 90,
                                                "handler_agent": "page-build-handler",
                                                "summary": "current_link.guidance",
                                                "item_key_prefix": "internal_link",
                                                "spec_data": {
                                                    "page_name": "current_link.source_page",
                                                    "suggestion": "current_link.guidance",
                                                    "link_target_url": "target_page.url",
                                                    "link_target_title": "target_page.title",
                                                    "anchor_text": "current_link.anchor_text",
                                                    "source": "internal-linker"
                                                }
                                            },
                                            "next_step": "done",
                                            "output_field": "item_created"
                                        },
                                        "done": {
                                            "action": "loop_complete"
                                        }
                                    }
                                }
                            },
                            "next_step": "complete",
                            "output_field": "items_created"
                        },
                        "complete": {
                            "action": "complete_workflow",
                            "config": { "output_fields": ["link_plan", "items_created"] }
                        },
                        "complete_not_found": {
                            "action": "complete_workflow",
                            "config": { "status": "skipped", "reason": "Target page not found", "output_fields": ["target_page"] }
                        },
                        "complete_no_candidates": {
                            "action": "complete_workflow",
                            "config": { "status": "skipped", "reason": "No candidate pages with content to link from", "output_fields": ["target_page"] }
                        },
                        "complete_error": {
                            "action": "complete_workflow",
                            "config": { "output_fields": ["link_plan"], "success_message": "LLM link planning failed" }
                        }
                    }
                }
            }'::jsonb,
            true,
            '["internal-linking", "content-structure"]'::jsonb,
            'docker.io/aqls/agent-chassis', 'v1.0.954',
            '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
            '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
            '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
            '[]'::jsonb,
            1,
            '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
            'specialist', 'active',
            '["linking", "content-structure", "navigation"]'::jsonb,
            '{"required": ["site_id"], "optional": ["domain", "work_item_id"], "description": "Receives site_id. Loads target page from work item spec, finds candidate pages, creates content_rewrite items."}'::jsonb,
            '{"produces": {"link_plan": "LLM plan with source pages and link guidance", "items_created": "content_rewrite work items for page-build-handler"}}'::jsonb,
            120
        );

        RAISE NOTICE 'internal-linker: INSERTED new definition';
END IF;
END $$;