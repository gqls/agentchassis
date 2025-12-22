-- ===========================================================================
-- CONTENT REVIEWER AGENT
-- File: 047_content_reviewer_agent.sql
-- ===========================================================================
-- Reviews content - supports both HITL (human) and auto-eval (LLM) modes.
-- Uses request_human_input action for HITL, execute_llm_prompt for auto.
-- Follows patterns from hitl_actions.go and hitl_request_human_input.go
-- ===========================================================================

BEGIN;

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    is_active,
    status,
    version,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    input_contract,
    output_contract,
    default_config
) VALUES (
             'content-reviewer',
             'Content Reviewer',
             'Reviews page content for quality, accuracy, and brand alignment. Supports HITL mode (human review with edits) and auto-eval mode (LLM review with auto-approve or flag).',
             'specialist',
             true,
             'active',
             1,
             '["content-review", "quality-assurance", "hitl", "auto-eval"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.575',
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "error": "system.errors.{type}",
                 "process": "system.agent.{type}.process",
                 "response": "system.responses.{type}"
             }'::jsonb,
             '{
                 "port": 8080,
                 "liveness_path": "/health",
                 "readiness_path": "/ready",
                 "initial_delay_seconds": 15
             }'::jsonb,
             -- Input contract
             '{
                 "expects": {
                     "current_page": "object with page name and title",
                     "page_content": "object with sections[] containing rendered HTML",
                     "reviewed_brief": "object with company info for validation"
                 },
                 "required": ["current_page", "page_content"]
             }'::jsonb,
             -- Output contract
             '{
                 "produces": {
                     "approved": "boolean - whether content passed review",
                     "review_mode": "string - hitl or auto-eval",
                     "reviewed_by": "string - user email or eval-agent",
                     "reviewed_at": "timestamp",
                     "edits": "object - any modifications made",
                     "issues": "array - issues found (if not approved)",
                     "content": "object - final content (possibly edited)"
                 }
             }'::jsonb,
             -- Workflow - supports both HITL and auto-eval modes
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "determine_review_mode",
                     "steps": {
                         "determine_review_mode": {
                             "action": "conditional",
                             "description": "Check whether to use HITL or auto-eval",
                             "config": {
                                 "condition": "input_data.review_mode == ''hitl'' OR input_data.require_human_review == true",
                                 "then_step": "prepare_hitl_review",
                                 "else_step": "auto_eval_content"
                             }
                         },

                         "prepare_hitl_review": {
                             "action": "prepare_review_data",
                             "description": "Prepare content for human review interface",
                             "config": {
                                 "include_fields": ["current_page", "page_content", "reviewed_brief"],
                                 "format_for_display": true
                             },
                             "next_step": "request_human_review",
                             "output_field": "review_data"
                         },

                         "request_human_review": {
                             "action": "request_human_input",
                             "description": "Send to HITL for human review and editing",
                             "config": {
                                 "request_type": "review",
                                 "notification_topic": "system.notifications.ui",
                                 "editable": true,
                                 "data_field": "review_data",
                                 "message": "Review page content for {{current_page.name}}",
                                 "ui_config": {
                                     "title": "Content Review",
                                     "description": "Review and edit page content before publishing",
                                     "show_diff": true,
                                     "allow_comments": true
                                 },
                                 "timeout_seconds": 3600,
                                 "stop_on_cancel": false
                             },
                             "next_step": "process_human_response",
                             "output_field": "human_response"
                         },

                         "process_human_response": {
                             "action": "process_human_input_response",
                             "description": "Process the human reviewer''s response",
                             "config": {
                                 "extract_edits": true
                             },
                             "next_step": "finalize_hitl_result",
                             "output_field": "processed_response"
                         },

                         "finalize_hitl_result": {
                             "action": "build_review_result",
                             "description": "Build final result from HITL review",
                             "config": {
                                 "review_mode": "hitl",
                                 "approved_field": "processed_response.approved",
                                 "edits_field": "processed_response.edits",
                                 "reviewer_field": "processed_response.responded_by"
                             },
                             "next_step": "update_component_status",
                             "output_field": "review_result"
                         },

                         "auto_eval_content": {
                             "action": "execute_llm_prompt",
                             "description": "LLM evaluates content quality and accuracy",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5-20250514",
                                     "max_tokens": 1500,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["current_page", "page_content", "reviewed_brief"],
                                 "output_format": "json",
                                 "prompt_template": "Review this page content for quality and accuracy.\n\n## Page\nName: {{current_page.name}}\nTitle: {{current_page.title}}\n\n## Company Brief\nCompany: {{reviewed_brief.company_name}}\nIndustry: {{reviewed_brief.industry}}\nTone: {{reviewed_brief.tone}}\nServices: {{reviewed_brief.services}}\n\n## Content to Review\n{{#each page_content.sections}}\n### Section: {{this.component_name}}\n{{this.rendered_html}}\n{{/each}}\n\n## Evaluation Criteria\n1. Accuracy: Does content match the brief? No invented claims?\n2. Completeness: Are all sections filled in properly?\n3. Quality: Professional tone? No placeholder text?\n4. Brand Alignment: Matches company voice and values?\n5. Technical: Valid HTML? Proper structure?\n\n## Return JSON:\n```json\n{\n  \"approved\": true/false,\n  \"overall_score\": 0.0-1.0,\n  \"issues\": [\n    {\n      \"section\": \"hero\",\n      \"severity\": \"error|warning|info\",\n      \"issue\": \"Description of the issue\",\n      \"suggestion\": \"How to fix it\"\n    }\n  ],\n  \"strengths\": [\"Good point 1\", \"Good point 2\"],\n  \"summary\": \"Brief overall assessment\"\n}\n```\n\nApprove if:\n- No errors (warnings are OK)\n- Score >= 0.7\n- No placeholder text detected\n- Content matches brief"
                             },
                             "next_step": "check_auto_approval",
                             "output_field": "eval_result"
                         },

                         "check_auto_approval": {
                             "action": "conditional",
                             "description": "Check if auto-eval passed or needs human review",
                             "config": {
                                 "condition": "eval_result.approved == true AND eval_result.overall_score >= 0.7",
                                 "then_step": "finalize_auto_result",
                                 "else_step": "escalate_to_human"
                             }
                         },

                         "escalate_to_human": {
                             "action": "request_human_input",
                             "description": "Escalate to human - auto-eval found issues",
                             "config": {
                                 "request_type": "review",
                                 "notification_topic": "system.notifications.ui",
                                 "editable": true,
                                 "message": "Auto-review found issues with {{current_page.name}} - human review required",
                                 "ui_config": {
                                     "title": "Content Review - Issues Found",
                                     "description": "Auto-review flagged issues. Please review and fix.",
                                     "show_issues": true,
                                     "issues_field": "eval_result.issues"
                                 },
                                 "timeout_seconds": 3600
                             },
                             "next_step": "process_escalation_response",
                             "output_field": "escalation_response"
                         },

                         "process_escalation_response": {
                             "action": "process_human_input_response",
                             "description": "Process response from escalated review",
                             "config": {
                                 "extract_edits": true
                             },
                             "next_step": "finalize_escalation_result",
                             "output_field": "escalation_processed"
                         },

                         "finalize_escalation_result": {
                             "action": "build_review_result",
                             "description": "Build result from escalated HITL review",
                             "config": {
                                 "review_mode": "escalated",
                                 "auto_eval_issues": "eval_result.issues"
                             },
                             "next_step": "update_component_status",
                             "output_field": "review_result"
                         },

                         "finalize_auto_result": {
                             "action": "build_review_result",
                             "description": "Build result from successful auto-eval",
                             "config": {
                                 "review_mode": "auto-eval",
                                 "approved": true,
                                 "reviewer": "eval-agent",
                                 "eval_score": "eval_result.overall_score"
                             },
                             "next_step": "update_component_status",
                             "output_field": "review_result"
                         },

                         "update_component_status": {
                             "action": "update_page_components_status",
                             "description": "Update page_components build_status and review info",
                             "config": {
                                 "page_from": "current_page",
                                 "status": "approved",
                                 "reviewed_at_field": "review_result.reviewed_at",
                                 "reviewed_by_field": "review_result.reviewed_by"
                             },
                             "next_step": "complete",
                             "output_field": "status_updated"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output": {
                                     "approved": "review_result.approved",
                                     "review_mode": "review_result.review_mode",
                                     "reviewed_by": "review_result.reviewed_by",
                                     "reviewed_at": "review_result.reviewed_at",
                                     "content": "review_result.content",
                                     "edits": "review_result.edits",
                                     "issues": "review_result.issues"
                                 }
                             }
                         }
                     }
                 }
             }'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              version = EXCLUDED.version,
                              default_config = EXCLUDED.default_config,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = now();

COMMIT;