-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================

-- content-reviewer
-- Reviews content, may involve HITL
UPDATE agent_definitions
SET input_contract = '{
    "required": ["current_page", "page_content"],
    "optional": ["reviewed_brief"]
}'::jsonb,
    output_contract = '{
    "produces": ["review_result", "approved", "feedback"]
}'::jsonb
WHERE type = 'content-reviewer';

----

-- add review code

-- Update Content-Reviewer to validate links and emails
--
-- Adds a validation step that runs BEFORE review mode determination.
-- Validation issues are passed to both auto-eval and HITL review.
--
-- Flow:
--   validate_content → determine_review_mode → (auto-eval OR hitl) → complete
--
-- Validation checks:
--   1. Internal links - must point to existing pages in the site
--   2. Email addresses - must match site's contact_email (warns on mismatch)

-- First, add the new validate_content step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,validate_content}',
        '{
            "action": "validate_page_content",
            "config": {
                "html_field": "page_content.response.page_html",
                "site_id_field": "input_data.site_record.site_id",
                "check_internal_links": true,
                "check_emails": true,
                "pages_from": "input_data.site_plan.pages"
            },
            "description": "Check for broken links and incorrect contact info",
            "next_step": "determine_review_mode",
            "output_field": "validation_result"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'content-reviewer';

-- Update start_step to validate_content
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,start_step}',
        '"validate_content"'
                     )
WHERE type = 'content-reviewer';

-- Update auto_eval_content to include validation issues in its prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,auto_eval_content,config,input_fields}',
        '["current_page", "page_content", "reviewed_brief", "validation_result"]'::jsonb
                     )
WHERE type = 'content-reviewer';

-- Update prepare_hitl_review to include validation issues
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,prepare_hitl_review,config,include_fields}',
        '["current_page", "page_content", "reviewed_brief", "validation_result"]'::jsonb
                     )
WHERE type = 'content-reviewer';

-- Update escalate_to_human to show validation issues
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,escalate_to_human,config}',
        '{
            "message": "Content review needed for {{current_page.name}}. Please review and fix.",
            "show_issues": true,
            "issues_field": "eval_result.issues",
            "validation_issues_field": "validation_result.issues",
            "request_type": "review",
            "timeout_seconds": 3600,
            "notification_topic": "system.notifications.ui"
        }'::jsonb
                     )
WHERE type = 'content-reviewer';

-- Add conditional to check validation before auto-eval approval
-- If validation has errors, always escalate to human
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_auto_approval,config,condition}',
        '"eval_result.approved == true AND eval_result.overall_score >= 0.7 AND (validation_result.error_count == 0 OR validation_result.error_count == null)"'
                     )
WHERE type = 'content-reviewer';

-- Verify the changes
SELECT type, version,
       default_config->'workflow'->'start_step' as start_step,
       jsonb_object_keys(default_config->'workflow'->'steps') as step_names
FROM agent_definitions
WHERE type = 'content-reviewer';

-- Show the new validation step
SELECT type,
       jsonb_pretty(default_config->'workflow'->'steps'->'validate_content') as validate_step
FROM agent_definitions
WHERE type = 'content-reviewer';

-- ============================================================
-- Also update the auto_eval_content LLM prompt to mention validation
-- ============================================================
-- Note: This assumes auto_eval_content uses execute_llm_prompt
-- The prompt should tell the LLM about any validation issues found

-- Check current auto_eval step config
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'auto_eval_content')
FROM agent_definitions WHERE type = 'content-reviewer';

