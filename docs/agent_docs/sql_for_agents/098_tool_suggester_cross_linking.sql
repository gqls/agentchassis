-- Migration 071: Add tool cross-linking to tool-suggester
--
-- Two changes:
-- 1. Update suggest_tools prompt to include related_pages in output schema
-- 2. Add create_cross_links step after create_items_loop
--
-- The cross-link step calls create_tool_cross_link_items (Go action) which
-- reads the suggestions array and creates content_rewrite work items for
-- each related page. page-build-handler then weaves tool references into
-- existing page content.

-- ═══════════════════════════════════════════════════════════════════════
-- Part 1: Update the LLM prompt to request related_pages per suggestion
-- ═══════════════════════════════════════════════════════════════════════

-- We need to update the JSON output schema in the prompt.
-- The prompt is in: default_config -> workflow -> steps -> suggest_tools -> config -> prompt_template
--
-- We add to the suggestion object:
--   "related_pages": ["page-name-1", "page-name-2"]
-- And add instruction text about what related_pages means.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,config,prompt_template}',
        to_jsonb(
                regexp_replace(
                    -- Get current prompt
                        default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template',
                    -- Find the complexity field in the JSON schema
                        '"complexity": "simple"',
                    -- Replace with complexity + related_pages
                        '"complexity": "simple",\n      "related_pages": ["page-name-1", "page-name-2"]',
                        'g'
                )
        )
                     )
WHERE type = 'tool-suggester';

-- Now add the instruction about related_pages to the prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,config,prompt_template}',
        to_jsonb(
                regexp_replace(
                        default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template',
                        '6. Also list tools from the library',
                        '6. For each suggestion, include related_pages: a list of 1-3 existing page names (from the pages list above) where a contextual reference to the tool would help visitors. Choose pages whose topic naturally connects to what the tool does. Do NOT include contact, privacy, or terms pages. Do NOT include the tool page itself.\n\n7. Also list tools from the library',
                        'g'
                )
        )
                     )
WHERE type = 'tool-suggester';

-- ═══════════════════════════════════════════════════════════════════════
-- Part 2: Add create_cross_links step after create_items_loop
-- ═══════════════════════════════════════════════════════════════════════

-- Current flow: ... -> create_items_loop -> complete
-- New flow:     ... -> create_items_loop -> create_cross_links -> complete

-- Step 1: Change create_items_loop's next_step from "complete" to "create_cross_links"
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_items_loop,next_step}',
        '"create_cross_links"'
                     )
WHERE type = 'tool-suggester';

-- Step 2: Add the create_cross_links step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_cross_links}',
        '{
            "action": "create_tool_cross_link_items",
            "config": {
                "site_id": "site_record.site_id",
                "suggestions": "evaluation.result.suggestions"
            },
            "next_step": "complete",
            "error_step": "complete",
            "description": "Create content_rewrite items to cross-link tools from related pages",
            "output_field": "cross_link_result"
        }'
                     )
WHERE type = 'tool-suggester';

-- ═══════════════════════════════════════════════════════════════════════
-- Verify
-- ═══════════════════════════════════════════════════════════════════════

-- Check the prompt includes related_pages
SELECT default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template' LIKE '%related_pages%' as prompt_has_related_pages
FROM agent_definitions WHERE type = 'tool-suggester';

-- Check the workflow flow
SELECT
    default_config->'workflow'->'steps'->'create_items_loop'->>'next_step' as items_loop_next,
    default_config->'workflow'->'steps'->'create_cross_links'->>'action' as cross_links_action,
    default_config->'workflow'->'steps'->'create_cross_links'->>'next_step' as cross_links_next
FROM agent_definitions WHERE type = 'tool-suggester';