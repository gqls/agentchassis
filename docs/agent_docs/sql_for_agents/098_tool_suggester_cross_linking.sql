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

-- 071: Add cross-linking to tool-suggester
-- Adds related_pages to LLM output schema + create_cross_links workflow step
-- Go action create_tool_cross_link_items is already deployed and registered.
--
-- Changes:
--   1. create_items_loop.next_step: "complete" → "create_cross_links"
--   2. New step: create_cross_links (calls create_tool_cross_link_items action)
--   3. Prompt: add related_pages field to JSON schema + instruction 6
--
-- Safe to re-run (idempotent jsonb_set operations).

BEGIN;

-- Step 1: Add the create_cross_links step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_cross_links}',
        '{
            "action": "create_tool_cross_link_items",
            "config": {
                "input_fields": ["input_data"]
            },
            "next_step": "complete",
            "description": "Create content_rewrite work items to cross-link tools from related pages",
            "output_field": "cross_links_created"
        }'::jsonb
                     )
WHERE type = 'tool-suggester';

-- Step 2: Point create_items_loop to create_cross_links instead of complete
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_items_loop,next_step}',
        '"create_cross_links"'::jsonb
                     )
WHERE type = 'tool-suggester';

-- Step 3: Update the prompt to include related_pages in JSON schema and instructions
-- We replace the complexity line in the schema to add related_pages after it,
-- and add instruction 6 before the rejected_tools section.
--
-- The prompt is stored as a deeply escaped string inside jsonb. We use
-- replace() on the text representation of the prompt_template value.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,config,prompt_template}',
        to_jsonb(
                replace(
                        replace(
                                default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template',
                            -- Add related_pages field after complexity in the JSON schema
                                '"complexity": "simple"',
                                '"complexity": "simple",\n      "related_pages": ["page-name-1", "page-name-2"]'
                        ),
                    -- Add instruction 6 before the rejected_tools instruction
                        'Also list tools from the library that you considered but rejected',
                        '6. For each suggestion, include related_pages: a list of 1-3 existing page names (from the Existing Pages list above) whose topic connects naturally to the tool. These pages will get a contextual reference added to them.\n\nAlso list tools from the library that you considered but rejected'
                )
        )
                     )
WHERE type = 'tool-suggester';

-- Step 4: Also add cross_links_created to the complete step output_fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config,output_fields}',
        '["evaluation", "items_created", "cross_links_created"]'::jsonb
                     )
WHERE type = 'tool-suggester';

COMMIT;

-- Verification queries:
-- 1. Check structural changes
SELECT
    default_config->'workflow'->'steps'->'create_items_loop'->>'next_step' as items_loop_next,
    default_config->'workflow'->'steps'->'create_cross_links'->>'action' as cross_links_action,
    default_config->'workflow'->'steps'->'create_cross_links'->>'next_step' as cross_links_next
FROM agent_definitions WHERE type = 'tool-suggester';
-- Expected: create_cross_links, create_tool_cross_link_items, complete

-- 2. Check prompt has related_pages
SELECT
    default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template' LIKE '%related_pages%' as has_related_pages
FROM agent_definitions WHERE type = 'tool-suggester';
-- Expected: true

-- 071: Add cross-linking to tool-suggester
-- Adds related_pages to LLM output schema + create_cross_links workflow step
-- Go action create_tool_cross_link_items is already deployed and registered.
--
-- Changes:
--   1. create_items_loop.next_step: "complete" → "create_cross_links"
--   2. New step: create_cross_links (calls create_tool_cross_link_items action)
--   3. Prompt: add related_pages field to JSON schema + instruction 6
--   4. complete step: add cross_links_created to output_fields
--
-- Safe to re-run (checks before modifying).

BEGIN;

-- Step 1: Add the create_cross_links step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_cross_links}',
        '{
            "action": "create_tool_cross_link_items",
            "config": {
                "input_fields": ["input_data"]
            },
            "next_step": "complete",
            "description": "Create content_rewrite work items to cross-link tools from related pages",
            "output_field": "cross_links_created"
        }'::jsonb
                     )
WHERE type = 'tool-suggester';

-- Step 2: Point create_items_loop to create_cross_links instead of complete
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_items_loop,next_step}',
        '"create_cross_links"'::jsonb
                     )
WHERE type = 'tool-suggester';

-- Step 3: Update the prompt — uses DO block to handle escaping correctly
DO $$
DECLARE
v_prompt text;
    v_new_prompt text;
    v_changed boolean := false;
BEGIN
    -- Extract the current prompt
SELECT default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template'
INTO v_prompt
FROM agent_definitions WHERE type = 'tool-suggester';

IF v_prompt IS NULL THEN
        RAISE NOTICE '071: tool-suggester prompt_template not found!';
        RETURN;
END IF;

    -- Check if already applied
    IF v_prompt LIKE '%related_pages%' THEN
        RAISE NOTICE '071: related_pages already in prompt, skipping prompt update';
        RETURN;
END IF;

    -- Replace 1: Add related_pages to JSON schema after complexity line
    -- The prompt stores JSON examples with escaped quotes: \"complexity\": \"simple\"
    IF v_prompt LIKE '%\"complexity\": \"simple\"%' THEN
        -- Unescaped quotes (unlikely but handle it)
        v_new_prompt := replace(
            v_prompt,
            '"complexity": "simple"',
            '"complexity": "simple",' || E'\n' || '      "related_pages": ["page-name-1", "page-name-2"]'
        );
        v_changed := true;
        RAISE NOTICE '071: Matched unescaped quotes form';
    ELSIF v_prompt LIKE E'%\\"complexity\\": \\"simple\\"%' THEN
        -- Escaped quotes (expected form from Go template strings)
        v_new_prompt := replace(
            v_prompt,
            E'\\\"complexity\\\": \\\"simple\\\"',
            E'\\\"complexity\\\": \\\"simple\\\",' || E'\\n      \\\"related_pages\\\": [\\\"page-name-1\\\", \\\"page-name-2\\\"]'
        );
        v_changed := true;
        RAISE NOTICE '071: Matched escaped quotes form';
ELSE
        RAISE NOTICE '071: WARNING — could not match complexity pattern in prompt. Showing context:';
        -- Show what's around "complexity" for debugging
        RAISE NOTICE '071: Prompt substring: %', substring(v_prompt from position('complexity' in v_prompt) - 20 for 100);
        RETURN;
END IF;

    -- Replace 2: Add instruction 6 before "Also list tools"
    IF v_new_prompt LIKE '%Also list tools from the library%' THEN
        v_new_prompt := replace(
            v_new_prompt,
            'Also list tools from the library that you considered but rejected',
            '6. For each suggestion, include related_pages: a list of 1-3 existing page names (from the Existing Pages list above) whose topic connects naturally to the tool. These pages will get a contextual reference added to them.' || E'\\n\\n' || 'Also list tools from the library that you considered but rejected'
        );
        RAISE NOTICE '071: Added instruction 6';
ELSE
        RAISE NOTICE '071: WARNING — could not find "Also list tools" in prompt';
END IF;

    -- Apply the updated prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,suggest_tools,config,prompt_template}',
        to_jsonb(v_new_prompt)
                     )
WHERE type = 'tool-suggester';

RAISE NOTICE '071: Prompt updated successfully';
END $$;

-- Step 4: Add cross_links_created to complete step output_fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config,output_fields}',
        '["evaluation", "items_created", "cross_links_created"]'::jsonb
                     )
WHERE type = 'tool-suggester';

COMMIT;

-- ============================================================================
-- Verification
-- ============================================================================

-- 1. Check structural changes
SELECT
    default_config->'workflow'->'steps'->'create_items_loop'->>'next_step' as items_loop_next,
    default_config->'workflow'->'steps'->'create_cross_links'->>'action' as cross_links_action,
    default_config->'workflow'->'steps'->'create_cross_links'->>'next_step' as cross_links_next
FROM agent_definitions WHERE type = 'tool-suggester';
-- Expected: create_cross_links, create_tool_cross_link_items, complete

-- 2. Check prompt has related_pages
SELECT
    default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template' LIKE '%related_pages%' as has_related_pages
FROM agent_definitions WHERE type = 'tool-suggester';
-- Expected: true

-- 3. Show a snippet of the prompt around related_pages to confirm formatting
SELECT substring(
               default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template'
    FROM position('related_pages' in (default_config->'workflow'->'steps'->'suggest_tools'->'config'->>'prompt_template')) - 30
    FOR 120
) as prompt_context
FROM agent_definitions WHERE type = 'tool-suggester';