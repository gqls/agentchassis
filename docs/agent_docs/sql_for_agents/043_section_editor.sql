-- =============================================================================
-- SECTION-EDITOR AGENT
-- =============================================================================
-- Granular editing of individual page sections without full rebuild pipeline.
-- Always updates content_data first, then re-renders from template + DB context.
-- This ensures edits survive future re-renders (nav updates, theme changes, etc).
--
-- Supports:
--   content_edit:    Update content_data fields (merge or full replace),
--                    re-render component template with full site context from DB
--   component_swap:  Replace component template, re-render with existing content_data
--
-- Input contract:
--   Required: domain (or site_id), edit_type
--   Target:   page_component_id OR (page_name + slot_name)
--   Edit-specific:
--     content_edit:    field_updates (merge) OR content_data (replace)
--     component_swap:  new_component_function
--
-- Example trigger messages:
--
-- Field update (merge into existing content_data):
-- {
--   "domain": "leopardessconsulting.co.uk",
--   "page_name": "index",
--   "slot_name": "hero",
--   "edit_type": "content_edit",
--   "field_updates": { "headline": "Strategic Consulting for Growth" }
-- }
--
-- Full content replacement (e.g. rewrite a case study):
-- {
--   "domain": "leopardessconsulting.co.uk",
--   "page_name": "use-cases",
--   "slot_name": "case-studies-list",
--   "edit_type": "content_edit",
--   "content_data": {
--     "section_title": "How We Help",
--     "cases": [
--       {"title": "Digital Transformation", "description": "..."},
--       {"title": "Process Optimisation", "description": "..."}
--     ]
--   }
-- }
--
-- Component swap:
-- {
--   "domain": "leopardessconsulting.co.uk",
--   "page_name": "index",
--   "slot_name": "social-proof",
--   "edit_type": "component_swap",
--   "new_component_function": "testimonials-grid"
-- }
-- =============================================================================


INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    version,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'section-editor',
             'Section Editor Agent',
             'Performs granular edits to individual page sections. Updates content_data (source of truth) then re-renders from template with full site context. Edits survive future re-renders.',
             'specialist',
             jsonb_build_object(
                     'workflow', jsonb_build_object(
                     'start_step', 'ensure_site_record',
                     'processing_mode', 'orchestrator',
                     'timeout_seconds', 300,
                     'steps', jsonb_build_object(

                         -- Step 1: Load site record from domain/site_id
                             'ensure_site_record', jsonb_build_object(
                                     'action', 'ensure_site_record',
                                     'config', '{}'::jsonb,
                                     'next_step', 'spawn_deployer',
                                     'description', 'Load site record from database',
                                     'output_field', 'site_record'
                                                   ),

                         -- Step 2: Spawn deployer for later use
                             'spawn_deployer', jsonb_build_object(
                                     'action', 'spawn_agent',
                                     'config', jsonb_build_object(
                                             'role', 'deployer',
                                             'agent_type', 'deployer-agent'
                                               ),
                                     'next_step', 'load_edit_context',
                                     'description', 'Spawn deployer agent',
                                     'output_field', 'deployer_agent'
                                               ),

                         -- Step 3: Load target section + component template + page info
                             'load_edit_context', jsonb_build_object(
                                     'action', 'load_edit_context',
                                     'config', jsonb_build_object(
                                             'input_fields', jsonb_build_array(
                                                     'site_id', 'page_component_id', 'page_name', 'slot_name', 'domain'
                                                             )
                                               ),
                                     'next_step', 'apply_edit',
                                     'description', 'Load page component, component template, and page info',
                                     'output_field', 'edit_context'
                                                  ),

                         -- Step 4: Apply the edit (content_edit or component_swap)
                         --         Updates content_data, re-renders from template, reassembles page
                             'apply_edit', jsonb_build_object(
                                     'action', 'apply_section_edit',
                                     'config', jsonb_build_object(
                                             'input_fields', jsonb_build_array(
                                                     'edit_type', 'field_updates', 'content_data',
                                                     'new_component_function', 'page_component_id'
                                                             )
                                               ),
                                     'next_step', 'deploy_page',
                                     'description', 'Apply edit to content_data, re-render, reassemble page',
                                     'output_field', 'edit_result'
                                           ),

                         -- Step 5: Commit updated page to git
                             'deploy_page', jsonb_build_object(
                                     'action', 'git_commit',
                                     'config', jsonb_build_object(
                                             'domain_field', 'edit_result.domain',
                                             'content_field', 'edit_result.html',
                                             'filename_field', 'edit_result.filename',
                                             'commit_message', 'Section edit via section-editor'
                                               ),
                                     'next_step', 'update_page_status',
                                     'description', 'Commit updated page HTML to git',
                                     'output_field', 'git_result'
                                            ),

                         -- Step 6: Mark page as deployed
                             'update_page_status', jsonb_build_object(
                                     'action', 'update_page_status',
                                     'config', jsonb_build_object(
                                             'status', 'deployed',
                                             'page_id_field', 'edit_result.page_id',
                                             'commit_from', 'git_result.commit_sha'
                                               ),
                                     'next_step', 'trigger_deploy',
                                     'description', 'Update page build_status to deployed',
                                     'output_field', 'status_updated'
                                                   ),

                         -- Step 7: Trigger Cloudflare deploy
                             'trigger_deploy', jsonb_build_object(
                                     'action', 'call_agent',
                                     'config', jsonb_build_object(
                                             'agent_type', 'deployer-agent',
                                             'target_role', 'deployer',
                                             'input_mapping', jsonb_build_object(
                                                     'site_record', 'site_record'
                                                              ),
                                             'timeout_seconds', 120
                                               ),
                                     'next_step', 'complete',
                                     'description', 'Trigger Cloudflare deployment',
                                     'output_field', 'deploy_result'
                                               ),

                         -- Step 8: Done
                             'complete', jsonb_build_object(
                                     'action', 'complete_workflow',
                                     'config', jsonb_build_object(
                                             'output_fields', jsonb_build_array('edit_result', 'git_result', 'deploy_result')
                                               ),
                                     'description', 'Section edit complete'
                                         )
                              )
                                 )
             ),
             true,   -- is_active
             '[]',   -- capabilities
             1,      -- version
             'specialist',
             'experimental',
             '["maintenance", "editing", "granular"]'::jsonb,
             '{}',   -- briefing_questionnaire
             0,
             false,
             jsonb_build_object(
                     'required', jsonb_build_array('domain', 'edit_type'),
                     'optional', jsonb_build_array(
                             'site_id', 'page_component_id', 'page_name', 'slot_name',
                             'field_updates', 'content_data', 'new_component_function'
                                 ),
                     'description', 'Provide domain + edit_type + target (page_component_id or page_name+slot_name) + edit params'
             ),
             jsonb_build_object(
                     'produces', jsonb_build_object(
                     'edit_result', 'Reassembled page HTML + metadata',
                     'git_result', 'Git commit result',
                     'deploy_result', 'Cloudflare deployment result'
                                 )
             )
         );

-- the mapping is now (note ? is optional):
{
  "domain": "input_data.domain",           // required — always present
  "edit_type": "input_data.edit_type",     // required — always present
  "page_name?": "input_data.page_name",    // optional — may target by ID instead
  "slot_name?": "input_data.slot_name",    // optional
  "field_updates?": "input_data.field_updates",           // only for content_edit merge
  "content_data?": "input_data.content_data",             // only for content_edit replace
  "new_component_function?": "input_data.new_component_function",  // only for component_swap
  "page_component_id?": "input_data.page_component_id"   // alternative targeting
}


-- =============================================================================
-- TRIGGER SCRIPTS (for manual testing)
-- =============================================================================
--
-- Content edit - merge field_updates into existing content_data:
-- cat <<'EOF' | kubectl -n kafka exec -i kafka-0 -- kafka-console-producer.sh \
--   --broker-list localhost:9092 --topic system.intake
-- {
--   "type": "section-editor",
--   "data": {
--     "domain": "leopardessconsulting.co.uk",
--     "page_name": "index",
--     "slot_name": "hero",
--     "edit_type": "content_edit",
--     "field_updates": {
--       "headline": "Strategic Consulting for Growth"
--     }
--   }
-- }
-- EOF
--
-- Content edit - full replacement (rewrite case studies):
-- cat <<'EOF' | kubectl -n kafka exec -i kafka-0 -- kafka-console-producer.sh \
--   --broker-list localhost:9092 --topic system.intake
-- {
--   "type": "section-editor",
--   "data": {
--     "domain": "leopardessconsulting.co.uk",
--     "page_name": "use-cases",
--     "slot_name": "case-studies-list",
--     "edit_type": "content_edit",
--     "content_data": {
--       "section_title": "How We Help",
--       "section_subtitle": "Real results for real businesses",
--       "cases": [
--         {
--           "title": "Digital Transformation",
--           "description": "Helping companies modernise their technology and processes",
--           "outcome": "Streamlined operations across the organisation"
--         }
--       ]
--     }
--   }
-- }
-- EOF
--
-- Component swap:
-- cat <<'EOF' | kubectl -n kafka exec -i kafka-0 -- kafka-console-producer.sh \
--   --broker-list localhost:9092 --topic system.intake
-- {
--   "type": "section-editor",
--   "data": {
--     "domain": "leopardessconsulting.co.uk",
--     "page_name": "index",
--     "slot_name": "social-proof",
--     "edit_type": "component_swap",
--     "new_component_function": "testimonials-grid"
--   }
-- }
-- EOF


                                                       ---


-- Fix section-editor agent definition: rename content_data → replacement_content_data
--
-- Root cause: ExtractActionInputs has a backward-compat nested lookup that checks
-- site_record.<field_name> for every optional field. The field name "content_data"
-- collides with site_record.content_data (the site plan from ensure_site_record).
-- When the apply_edit step listed "content_data" in input_fields, the nested lookup
-- found the site plan and used it as a full content replacement — overwriting the
-- hero's content_data with the site plan instead of merging field_updates.
--
-- Fix: rename to "replacement_content_data" — no site_record.replacement_content_data exists.
-- The trigger script's input_mapping already translates:
--   "replacement_content_data?": "input_data.content_data"
-- so external callers still use the intuitive "content_data" field name.

-- Update the workflow: apply_edit step input_fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_edit,config,input_fields}',
        '["edit_type", "field_updates", "replacement_content_data", "new_component_function", "page_component_id"]'::jsonb
                     ),
-- Update the input_contract: rename content_data to replacement_content_data in optional list
    input_contract = jsonb_set(
            input_contract,
            '{optional}',
            '["site_id", "page_component_id", "page_name", "slot_name", "field_updates", "replacement_content_data", "new_component_function"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'section-editor' AND is_active = true;

-- Verify
SELECT
    default_config->'workflow'->'steps'->'apply_edit'->'config'->'input_fields' as apply_edit_input_fields,
    input_contract->'optional' as optional_contract
FROM agent_definitions
WHERE type = 'section-editor' AND is_active = true;

-- Also clear the contaminated hero content_data
UPDATE page_components
SET content_data = NULL
WHERE id = 'a41df8a0-7607-48a8-8462-722fb2d1c1b2';