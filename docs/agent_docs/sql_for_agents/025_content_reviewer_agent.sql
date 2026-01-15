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

---

-- FIX: content-reviewer template syntax
-- ======================================
-- Issues:
-- 1. Missing dot prefixes on variable references
-- 2. Handlebars syntax ({{#each}}, {{this.}}) needs Go template syntax

-- Fix auto_eval_content step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,auto_eval_content,config,prompt_template}',
        '"Review this page content for quality and accuracy.\n\n## Page\nName: {{.current_page.name}}\nTitle: {{.current_page.title}}\n\n## Company Brief\nCompany: {{.reviewed_brief.company_name}}\nIndustry: {{.reviewed_brief.industry}}\nTone: {{.reviewed_brief.tone}}\nServices: {{.reviewed_brief.services}}\n\n## Content to Review\n{{range .page_content.sections}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}\n\n## Evaluation Criteria\n1. Accuracy: Does content match the brief? No invented claims?\n2. Completeness: Are all sections filled in properly?\n3. Quality: Professional tone? No placeholder text?\n4. Brand Alignment: Matches company voice and values?\n5. Technical: Valid HTML? Proper structure?\n\n## Return JSON:\n```json\n{\n  \"approved\": true/false,\n  \"overall_score\": 0.0-1.0,\n  \"issues\": [\n    {\n      \"section\": \"hero\",\n      \"severity\": \"error|warning|info\",\n      \"issue\": \"Description of the issue\",\n      \"suggestion\": \"How to fix it\"\n    }\n  ],\n  \"strengths\": [\"Good point 1\", \"Good point 2\"],\n  \"summary\": \"Brief overall assessment\"\n}\n```\n\nApprove if:\n- No errors (warnings are OK)\n- Score >= 0.7\n- No placeholder text detected\n- Content matches brief"'
                     )
WHERE type = 'content-reviewer';

-- Verify content-reviewer fix
SELECT 'content-reviewer' as agent,
       'auto_eval_content' as step,
       CASE
           WHEN default_config->'workflow'->'steps'->'auto_eval_content'->'config'->>'prompt_template' LIKE '%{{.current_page.%'
           AND default_config->'workflow'->'steps'->'auto_eval_content'->'config'->>'prompt_template' LIKE '%{{range .page_content.sections}}%'
    AND default_config->'workflow'->'steps'->'auto_eval_content'->'config'->>'prompt_template' LIKE '%{{.component_name}}%'
    THEN 'FIXED'
    ELSE 'NEEDS FIX'
END as status
FROM agent_definitions
WHERE type = 'content-reviewer';


    -- ============================================
-- 3. content-reviewer: auto_eval_content step
-- ============================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,auto_eval_content,config,ai_service,model}',
        '"claude-sonnet-4-5-20250929"'
                     ),
    updated_at = NOW()
WHERE type = 'content-reviewer';

--

current snapshot
 0a58cf02-4a9e-4900-9c7e-207622f12247 | content-reviewer | Content Reviewer | Reviews page content for quality, accuracy, and brand alignment. Supports HITL mode (human review with edits) and auto-eval mode (LLM review with auto-approve or flag). | specialist | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output": {"edits": "review_result.edits", "issues": "review_result.issues", "content": "review_result.content", "approved": "review_result.approved", "review_mode": "review_result.review_mode", "reviewed_at": "review_result.reviewed_at", "reviewed_by": "review_result.reviewed_by"}}}, "auto_eval_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250929", "provider": "anthropic", "max_tokens": 1500, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["current_page", "page_content", "reviewed_brief"], "output_format": "json", "prompt_template": "Review this page content for quality and accuracy.\n\n## Page\nName: {{.current_page.name}}\nTitle: {{.current_page.title}}\n\n## Company Brief\nCompany: {{.reviewed_brief.company_name}}\nIndustry: {{.reviewed_brief.industry}}\nTone: {{.reviewed_brief.tone}}\nServices: {{.reviewed_brief.services}}\n\n## Content to Review\n{{range .page_content.sections}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}\n\n## Evaluation Criteria\n1. Accuracy: Does content match the brief? No invented claims?\n2. Completeness: Are all sections filled in properly?\n3. Quality: Professional tone? No placeholder text?\n4. Brand Alignment: Matches company voice and values?\n5. Technical: Valid HTML? Proper structure?\n\n## Return JSON:\n```json\n{\n  \"approved\": true/false,\n  \"overall_score\": 0.0-1.0,\n  \"issues\": [\n    {\n      \"section\": \"hero\",\n      \"severity\": \"error|warning|info\",\n      \"issue\": \"Description of the issue\",\n      \"suggestion\": \"How to fix it\"\n    }\n  ],\n  \"strengths\": [\"Good point 1\", \"Good point 2\"],\n  \"summary\": \"Brief overall assessment\"\n}\n```\n\nApprove if:\n- No errors (warnings are OK)\n- Score >= 0.7\n- No placeholder text detected\n- Content matches brief"}, "next_step": "check_auto_approval", "description": "LLM evaluates content quality and accuracy", "output_field": "eval_result"}, "escalate_to_human": {"action": "request_human_input", "config": {"message": "Auto-review found issues with {{current_page.name}} - human review required", "editable": true, "ui_config": {"title": "Content Review - Issues Found", "description": "Auto-review flagged issues. Please review and fix.", "show_issues": true, "issues_field": "eval_result.issues"}, "request_type": "review", "timeout_seconds": 3600, "notification_topic": "system.notifications.ui"}, "next_step": "process_escalation_response", "description": "Escalate to human - auto-eval found issues", "output_field": "escalation_response"}, "check_auto_approval": {"action": "conditional", "config": {"condition": "eval_result.approved == true AND eval_result.overall_score >= 0.7", "else_step": "escalate_to_human", "then_step": "finalize_auto_result"}, "description": "Check if auto-eval passed or needs human review"}, "prepare_hitl_review": {"action": "prepare_review_data", "config": {"include_fields": ["current_page", "page_content", "reviewed_brief"], "format_for_display": true}, "next_step": "request_human_review", "description": "Prepare content for human review interface", "output_field": "review_data"}, "finalize_auto_result": {"action": "build_review_result", "config": {"approved": true, "reviewer": "eval-agent", "eval_score": "eval_result.overall_score", "review_mode": "auto-eval"}, "next_step": "update_component_status", "description": "Build result from successful auto-eval", "output_field": "review_result"}, "finalize_hitl_result": {"action": "build_review_result", "config": {"edits_field": "processed_response.edits", "review_mode": "hitl", "approved_field": "processed_response.approved", "reviewer_field": "processed_response.responded_by"}, "next_step": "update_component_status", "description": "Build final result from HITL review", "output_field": "review_result"}, "request_human_review": {"action": "request_human_input", "config": {"message": "Review page content for {{current_page.name}}", "editable": true, "ui_config": {"title": "Content Review", "show_diff": true, "description": "Review and edit page content before publishing", "allow_comments": true}, "data_field": "review_data", "request_type": "review", "stop_on_cancel": false, "timeout_seconds": 3600, "notification_topic": "system.notifications.ui"}, "next_step": "process_human_response", "description": "Send to HITL for human review and editing", "output_field": "human_response"}, "determine_review_mode": {"action": "conditional", "config": {"condition": "input_data.review_mode == 'hitl' OR input_data.require_human_review == true", "else_step": "auto_eval_content", "then_step": "prepare_hitl_review"}, "description": "Check whether to use HITL or auto-eval"}, "process_human_response": {"action": "process_human_input_response", "config": {"extract_edits": true}, "next_step": "finalize_hitl_result", "description": "Process the human reviewer's response", "output_field": "processed_response"}, "update_component_status": {"action": "update_page_components_status", "config": {"status": "approved", "page_from": "current_page", "reviewed_at_field": "review_result.reviewed_at", "reviewed_by_field": "review_result.reviewed_by"}, "next_step": "complete", "description": "Update page_components build_status and review info", "output_field": "status_updated"}, "finalize_escalation_result": {"action": "build_review_result", "config": {"review_mode": "escalated", "auto_eval_issues": "eval_result.issues"}, "next_step": "update_component_status", "description": "Build result from escalated HITL review", "output_field": "review_result"}, "process_escalation_response": {"action": "process_human_input_response", "config": {"extract_edits": true}, "next_step": "finalize_escalation_result", "description": "Process response from escalated review", "output_field": "escalation_processed"}}, "start_step": "determine_review_mode"}, "processing_mode": "task", "timeout_seconds": 600} | t         | 2025-12-22 17:47:42.958031+00 | 2026-01-04 15:53:02.823108+00 |            | ["content-review", "quality-assurance", "hitl", "auto-eval"] | docker.io/aqls/agent-chassis | v1.0.625  |         | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | active | []          | {}                     |           0 | f           | {"expects": {"current_page": "object with page name and title", "page_content": "object with sections[] containing rendered HTML", "reviewed_brief": "object with company info for validation"}, "required": ["current_page", "page_content"]} | {"produces": {"edits": "object - any modifications made", "issues": "array - issues found (if not approved)", "content": "object - final content (possibly edited)", "approved": "boolean - whether content passed review", "review_mode": "string - hitl or auto-eval", "reviewed_at": "timestamp", "reviewed_by": "string - user email or eval-agent"}}
(1 row)


-- Issue 35: Fix content-reviewer prompt template to use correct section data path
--
-- Problem: The prompt expects page_content.sections but page-content-writer
-- outputs sections in different locations depending on the workflow stage
--
-- This SQL updates ONLY the prompt_template, keeping all other config the same

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,auto_eval_content,config,prompt_template}',
        '"Review this page content for quality and accuracy.\n\n## Page\nName: {{.current_page.name}}\nTitle: {{.current_page.title}}\n\n## Company Brief\nCompany: {{.reviewed_brief.company_name}}\nIndustry: {{.reviewed_brief.industry}}\nTone: {{.reviewed_brief.tone}}\nServices: {{.reviewed_brief.services}}\n\n## Content to Review\n{{if .page_content.sections}}{{range .page_content.sections}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}{{else if .page_content.process_sections_loop_complete}}{{range .page_content.process_sections_loop_complete.results}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}{{else if .page_content.processed_sections}}{{range .page_content.processed_sections.results}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}{{else}}\n### Full Page HTML\n{{.page_content.compile_page.page_html}}\n{{end}}\n\n## Evaluation Criteria\n1. Accuracy: Does content match the brief? No invented claims?\n2. Completeness: Are all sections filled in properly?\n3. Quality: Professional tone? No placeholder text?\n4. Brand Alignment: Matches company voice and values?\n5. Technical: Valid HTML? Proper structure?\n\n## Return JSON:\n```json\n{\n  \"approved\": true/false,\n  \"overall_score\": 0.0-1.0,\n  \"issues\": [\n    {\n      \"section\": \"hero\",\n      \"severity\": \"error|warning|info\",\n      \"issue\": \"Description of the issue\",\n      \"suggestion\": \"How to fix it\"\n    }\n  ],\n  \"strengths\": [\"Good point 1\", \"Good point 2\"],\n  \"summary\": \"Brief overall assessment\"\n}\n```\n\nApprove if:\n- No errors (warnings are OK)\n- Score >= 0.7\n- No placeholder text detected\n- Content matches brief"'
                     ),
    updated_at = NOW()
WHERE type = 'content-reviewer';

-- Verify the update
SELECT type,
    LEFT(default_config->'workflow'->'steps'->'auto_eval_content'->'config'->>'prompt_template', 300) as prompt_preview
FROM agent_definitions
WHERE type = 'content-reviewer';


-- Lower the auto-approval threshold for content-reviewer
-- This will cause more content to be auto-approved without human review
--
-- Current: approved == true AND overall_score >= 0.7
-- New:     approved == true AND overall_score >= 0.5
--
-- DO NOT RUN until HITL flow is tested

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_auto_approval,config,condition}',
        '"eval_result.approved == true AND eval_result.overall_score >= 0.5"'::jsonb
             )
WHERE type = 'content-reviewer';

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'check_auto_approval'->'config'->'condition' as approval_condition
FROM agent_definitions
WHERE type = 'content-reviewer';

-- Alternative: Even more aggressive - approve if no errors regardless of score
-- UPDATE agent_definitions
-- SET config = jsonb_set(
--     config,
--     '{workflow,steps,check_auto_approval,config,condition}',
--     '"eval_result.approved == true"'::jsonb
-- )
-- WHERE agent_type = 'content-reviewer';

---

-- ============================================================================
-- Update content-reviewer conditional to handle missing review_mode gracefully
--
-- The current conditional checks for input_data.review_mode but this field
-- is not always passed by callers. The behavior (defaulting to auto-eval)
-- is correct, but the warning logs are noisy.
--
-- This update changes the conditional to check for the fields at multiple
-- paths and explicitly handle the "not specified" case.
-- ============================================================================

-- Option 1: Update the conditional to check both paths
-- The conditional currently is:
--   "condition": "input_data.review_mode == 'hitl' OR input_data.require_human_review == true"
--
-- Change to also check without input_data prefix:
--   "condition": "(input_data.review_mode == 'hitl' OR review_mode == 'hitl') OR (input_data.require_human_review == true OR require_human_review == true)"

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,determine_review_mode,config,condition}',
        '"(input_data.review_mode == ''hitl'' OR review_mode == ''hitl'') OR (input_data.require_human_review == true OR require_human_review == true)"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'content-reviewer';

-- Also update the input_contract to document that review_mode is optional
UPDATE agent_definitions
SET input_contract = '{
    "expects": {
        "current_page": "object with name, title - the page being reviewed",
        "page_content": "object with sections[] containing rendered content",
        "reviewed_brief": "object with company info for context",
        "review_mode": "string (optional) - ''hitl'' for human review, ''auto'' for LLM review. Defaults to auto.",
        "require_human_review": "boolean (optional) - if true, forces HITL mode. Defaults to false."
    },
    "required": ["current_page", "page_content"],
    "optional": ["reviewed_brief", "review_mode", "require_human_review"],
    "defaults": {
        "review_mode": "auto",
        "require_human_review": false
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'content-reviewer';

-- Verify the changes
SELECT
    type,
    default_config->'workflow'->'steps'->'determine_review_mode'->'config'->>'condition' as condition,
    input_contract
FROM agent_definitions
WHERE type = 'content-reviewer';


-- ============================================================================
-- Alternative: Add a "set_defaults" step before the conditional
-- This is a more robust approach that explicitly sets default values
-- ============================================================================

-- This would require adding a new step:
/*
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,set_review_defaults}',
    '{
        "action": "set_defaults",
        "config": {
            "defaults": {
                "review_mode": "auto",
                "require_human_review": false
            },
            "paths": [
                {"field": "review_mode", "try_paths": ["input_data.review_mode", "review_mode"]},
                {"field": "require_human_review", "try_paths": ["input_data.require_human_review", "require_human_review"]}
            ]
        },
        "next_step": "determine_review_mode",
        "description": "Set default values for optional review configuration"
    }'::jsonb
),
updated_at = NOW()
WHERE type = 'content-reviewer';

-- Then update the start_step
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,start_step}',
    '"set_review_defaults"'::jsonb
),
updated_at = NOW()
WHERE type = 'content-reviewer';
*/