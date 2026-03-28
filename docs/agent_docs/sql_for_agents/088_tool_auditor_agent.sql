-- ============================================================
-- tool-auditor agent definition
-- ============================================================
-- LLM-based code review of deployed tools (Tier 2 quality check).
-- Receives audit_tool work items from check_tool_health discovery.
-- Reads the full HTML/CSS/JS, sends to LLM for reasoning,
-- creates improve_tool items for issues it's confident about
-- and needs_human_review items for uncertain findings.
--
-- Flow:
--   check_tool_health (structural Tier 1) passes tools that aren't broken
--   → creates audit_tool work item
--   → tool-auditor loads the HTML and site context
--   → LLM reviews the code for bugs, mobile issues, UX, accessibility
--   → creates improve_tool items per finding
--
-- The LLM prompt is the core value — it reasons through the code
-- like a code review, checking interaction logic, layout behaviour,
-- edge cases, and accessibility. This catches things regex can't.

INSERT INTO agent_definitions (
    type, display_name, description, category, status,
    image_repository, image_tag,
    resources, default_config, input_contract, output_contract,
    domain_tags, agent_category, idle_timeout_seconds
) VALUES (
             'tool-auditor',
             'Tool Auditor',
             'LLM-based code review of deployed interactive tools. Reads full HTML/CSS/JS source, reasons through logic and layout, identifies bugs, mobile issues, UX problems, and accessibility gaps. Creates improve_tool items for fixable issues.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.914',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             $cfg${
        "workflow": {
            "start_step": "ensure_site_record",
             "processing_mode": "orchestrator",
             "timeout_seconds": 300,
             "steps": {
                "ensure_site_record": {
                    "action": "ensure_site_record",
             "config": {},
             "next_step": "load_tool",
             "output_field": "site_record"
                },
             "load_tool": {
                    "action": "query_database",
             "config": {
                        "query": "SELECT cc.id::text AS component_id, cc.function, cc.display_name, cc.html_template, COALESCE(pc.rendered_html, '') AS rendered_html, cc.description, p.id::text AS page_id, p.name AS page_name, p.url AS page_url, COALESCE(p.build_status, '') AS build_status FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1",
             "params": ["input_data.component_id"],
             "output_format": "object"
                    },
             "output_field": "tool_data",
             "next_step": "check_tool_found",
             "description": "Load full tool source and page context"
                },
             "check_tool_found": {
                    "action": "conditional",
             "config": {
                        "condition": "tool_data.component_id != null",
             "then_step": "load_site_context",
             "else_step": "complete_not_found"
                    }
                },
             "complete_not_found": {
                    "action": "complete_workflow",
             "config": {
                        "output_fields": ["tool_data"],
             "status": "skipped",
             "reason": "Tool component not found"
                    }
                },
             "load_site_context": {
                    "action": "read_site_spec",
             "config": {
                        "site_id": "site_record.site_id"
                    },
             "output_field": "site_specs",
             "next_step": "llm_audit",
             "error_step": "llm_audit",
             "description": "Load site context for industry-aware review"
                },
             "llm_audit": {
                    "action": "execute_llm_prompt",
             "config": {
                        "ai_service": {
                            "provider": "anthropic",
             "model": "claude-sonnet-4-6",
             "max_tokens": 4000,
             "api_key_env_var": "ANTHROPIC_API_KEY"
                        },
             "input_fields": ["tool_data", "site_record", "site_specs"],
             "output_format": "json",
             "prompt_template": "You are a senior frontend developer conducting a code review of an interactive tool deployed on a website. Your job is to find real bugs, usability problems, and mobile rendering issues by reasoning through the code.\n\n## Tool Under Review\nName: {{.tool_data.display_name}}\nFunction: {{.tool_data.function}}\nPage: {{.tool_data.page_url}}\n{{if .tool_data.description}}Purpose: {{.tool_data.description}}{{end}}\n\n## Site Context\nDomain: {{.site_record.domain}}\n{{if .site_specs.specs.identity}}Industry: {{.site_specs.specs.identity.industry}}\nAudience: {{.site_specs.specs.identity.target_audience}}{{end}}\n\n## Full Source Code\n{{.tool_data.html_template}}\n\n## What to Check\n\n### 1. JavaScript Logic Bugs\n- Do event listeners reference elements that exist in the HTML?\n- Are variables initialised before use?\n- Do calculations handle edge cases (zero, negative, empty input, NaN)?\n- Is there a division by zero risk?\n- Do DOM queries (getElementById, querySelector) match actual IDs/classes?\n- Are there race conditions in async code?\n\n### 2. Mobile & Touch\n- Will the layout work at 375px wide? Look at grid/flex definitions.\n- Are interactive targets (buttons, inputs, sliders) at least 44px tall?\n- Can all controls be reached by scrolling? Is anything clipped or overflowing?\n- Do any hover-only interactions have no touch equivalent?\n\n### 3. UX & Interaction\n- Is it obvious what the user should do first?\n- Is there feedback when the user takes an action (visual change, result update)?\n- Do all interactive elements have visible labels?\n- Is the copy/download functionality working (correct DOM references)?\n\n### 4. CSS & Styling\n- Are colours hardcoded as hex instead of using var(--color-*)?\n- Will the tool look broken without the site's global.css loaded?\n- Are there any !important declarations that could fight the site theme?\n\n### 5. Accessibility\n- Do form inputs have associated labels (label[for], aria-label, or wrapping label)?\n- Are there any images missing alt text?\n- Can the tool be operated with keyboard only (tab order, enter to submit)?\n\n### 6. Self-Containment\n- Does the tool reference external APIs, CDNs, or fetch calls?\n- Does it depend on variables/functions defined outside the template?\n- Are element IDs unique enough to avoid conflicts if multiple tools share a page?\n\n## Output Format\n\nReturn ONLY valid JSON. For each issue found, include:\n- category: one of 'js_bug', 'mobile', 'ux', 'css', 'accessibility', 'dependency'\n- severity: 'high' (broken functionality), 'medium' (degraded experience), 'low' (polish)\n- confidence: 'certain' (you can trace the bug in the code), 'likely' (strong evidence but not 100%), 'possible' (worth checking but you might be wrong)\n- description: what the problem is\n- fix_suggestion: how to fix it (be specific — reference line context, element IDs, CSS rules)\n- affected_element: which HTML element or JS function is involved\n\nIf the tool has NO issues, return an empty findings array with a brief summary.\n\n{\n  \"tool_function\": \"the-function-name\",\n  \"summary\": \"Brief overall assessment (1-2 sentences)\",\n  \"quality_score\": 1-10,\n  \"findings\": [\n    {\n      \"category\": \"mobile\",\n      \"severity\": \"medium\",\n      \"confidence\": \"certain\",\n      \"description\": \"Grid layout uses fixed 350px sidebar with no breakpoint below 900px\",\n      \"fix_suggestion\": \"Add @media (max-width: 600px) to stack sidebar below main content\",\n      \"affected_element\": \".tool-layout grid-template-columns\"\n    }\n  ]\n}"
                    },
             "output_field": "audit_result",
             "next_step": "check_has_findings",
             "error_step": "complete_error",
             "description": "LLM reviews the tool source code"
                },
             "check_has_findings": {
                    "action": "conditional",
             "config": {
                        "condition": "audit_result.result.findings != null",
             "then_step": "create_items_loop",
             "else_step": "complete"
                    }
                },
             "create_items_loop": {
                    "action": "loop",
             "config": {
                        "items_field": "audit_result.result.findings",
             "item_variable": "current_finding",
             "mode": "sequential",
             "max_iterations": 10,
             "continue_on_error": true,
             "sub_workflow": {
                            "start_step": "check_confidence",
             "steps": {
                                "check_confidence": {
                                    "action": "conditional",
             "config": {
                                        "condition": "current_finding.confidence != possible",
             "then_step": "create_improve_item",
             "else_step": "create_review_item"
                                    },
             "description": "Route certain/likely to auto-fix, possible to HITL"
                                },
             "create_improve_item": {
                                    "action": "create_work_item",
             "config": {
                                        "site_id": "site_record.site_id",
             "item_type": "improve_tool",
             "handler_agent": "tool-improver",
             "item_domain": "build",
             "severity": "current_finding.severity",
             "source": "tool-auditor",
             "summary": "current_finding.description",
             "priority": 60,
             "item_key_prefix": "audit_fix",
             "spec_data": {
                                            "component_id": "tool_data.component_id",
             "issue": "current_finding.description",
             "fix_suggestion": "current_finding.fix_suggestion",
             "category": "current_finding.category",
             "check": "tool_auditor",
             "page_id": "tool_data.page_id",
             "page_name": "tool_data.page_name"
                                        }
                                    },
             "output_field": "item_created",
             "next_step": "done"
                                },
             "create_review_item": {
                                    "action": "create_work_item",
             "config": {
                                        "site_id": "site_record.site_id",
             "item_type": "needs_human_review",
             "handler_agent": "hitl-review",
             "item_domain": "build",
             "severity": "low",
             "source": "tool-auditor",
             "summary": "current_finding.description",
             "priority": 100,
             "item_key_prefix": "audit_review",
             "spec_data": {
                                            "component_id": "tool_data.component_id",
             "tool_function": "tool_data.function",
             "issue": "current_finding.description",
             "fix_suggestion": "current_finding.fix_suggestion",
             "category": "current_finding.category",
             "confidence": "current_finding.confidence",
             "check": "tool_auditor"
                                        }
                                    },
             "output_field": "item_created",
             "next_step": "done"
                                },
             "done": {
                                    "action": "loop_complete",
             "description": "Finding processed"
                                }
                            }
                        }
                    },
             "output_field": "items_created",
             "next_step": "complete"
                },
             "complete": {
                    "action": "complete_workflow",
             "config": {
                        "output_fields": ["audit_result", "items_created"]
                    }
                },
             "complete_error": {
                    "action": "complete_workflow",
             "config": {
                        "output_fields": ["audit_result"],
             "success_message": "Audit LLM call failed"
                    }
                }
            }
        }
    }$cfg$::jsonb,
             '{"required": ["site_id", "domain", "component_id"], "optional": ["work_item_id"]}'::jsonb,
             '{"produces": {"audit_result": "LLM findings with quality_score", "items_created": "improve_tool and review items"}}'::jsonb,
             '["tools", "quality", "audit", "llm"]'::jsonb,
             'specialist',
             120
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    display_name = EXCLUDED.display_name,
    image_tag = EXCLUDED.image_tag,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();

-- Verify
SELECT type, display_name, status
FROM agent_definitions
WHERE type = 'tool-auditor' AND deleted_at IS NULL;

