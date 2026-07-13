-- ============================================================================
-- Agent Definitions: site-component-linker and component-template-fixer
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'site-component-linker',
             'Site Component Linker',
             'Links site_components rows to correct content_components from the style collection. Fixes NULL component_id that causes fallback rendering. Creates needs_rerender work item after linking.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.829', 'specialist',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":120,"steps":{"ensure_site_record":{"action":"ensure_site_record","config":{"store_brief_in_content_data":false},"next_step":"link_components","description":"Load site record","output_field":"site_record"},"link_components":{"action":"link_site_components","config":{"site_id":"site_record.site_id"},"next_step":"check_linked","description":"Link site_components to content_components from style collection","output_field":"link_result"},"check_linked":{"action":"conditional","config":{"condition":"link_result.linked > 0","then_step":"create_rerender","else_step":"complete"},"description":"Only create rerender item if we actually linked something"},"create_rerender":{"action":"query_database","config":{"query":"INSERT INTO site_work_items (site_id, source, domain, item_type, severity, summary, priority, handler_agent, status, created_by, spec) SELECT $1::uuid, ''side_effect'', ''build'', ''needs_rerender'', ''medium'', ''Rerender after component linkage fix'', 99, ''rerender-pages'', ''detected'', ''site-component-linker'', ''{\"refresh_site_components\": true}''::jsonb WHERE NOT EXISTS (SELECT 1 FROM site_work_items WHERE site_id = $1::uuid AND item_type = ''needs_rerender'' AND status IN (''detected'', ''triaged'', ''claimed''))","params":["site_record.site_id"]},"next_step":"complete","description":"Create needs_rerender work item","output_field":"rerender_created"},"complete":{"action":"complete_workflow","config":{"output_fields":["link_result"]},"description":"Complete"}}}}'::jsonb,
             '{"required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"link_result": "object with linked count"}}'::jsonb,
             '["maintenance", "components", "fix-agent"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'component-template-fixer',
             'Component Template Fixer',
             'Applies targeted fixes to site_components and page_components: CSS injection (nav flex), element removal (search icon), slot_name alignment. Routes on spec.fix_type.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.829', 'specialist',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":120,"steps":{"ensure_site_record":{"action":"ensure_site_record","config":{"store_brief_in_content_data":false},"next_step":"apply_fix","description":"Load site record","output_field":"site_record"},"apply_fix":{"action":"fix_component_template","config":{"site_id":"site_record.site_id"},"next_step":"check_needs_rerender","description":"Apply the fix specified in input_data.spec.fix_type","output_field":"fix_result"},"check_needs_rerender":{"action":"conditional","config":{"condition":"fix_result.fixed == true","then_step":"create_rerender","else_step":"complete"},"description":"Only rerender if something was actually changed"},"create_rerender":{"action":"query_database","config":{"query":"INSERT INTO site_work_items (site_id, source, domain, item_type, severity, summary, priority, handler_agent, status, created_by, spec) SELECT $1::uuid, ''side_effect'', ''build'', ''needs_rerender'', ''low'', ''Rerender after template fix'', 99, ''rerender-pages'', ''detected'', ''component-template-fixer'', ''{}''::jsonb WHERE NOT EXISTS (SELECT 1 FROM site_work_items WHERE site_id = $1::uuid AND item_type = ''needs_rerender'' AND status IN (''detected'', ''triaged'', ''claimed''))","params":["site_record.site_id"]},"next_step":"complete","description":"Create needs_rerender work item","output_field":"rerender_created"},"complete":{"action":"complete_workflow","config":{"output_fields":["fix_result"]},"description":"Complete"}}}}'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["fix_type", "slot_name", "pattern", "page_component_id"]}'::jsonb,
             '{"produces": {"fix_result": "object with fixed boolean and details"}}'::jsonb,
             '["maintenance", "components", "fix-agent", "css"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

SELECT type, display_name, status, agent_category
FROM agent_definitions
WHERE type IN ('site-component-linker', 'component-template-fixer')
  AND deleted_at IS NULL;


--

-- Add fix_type_field hint so the action knows to check spec.category
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_fix,config}',
        '{
            "site_id": "site_record.site_id",
            "fix_type_field": "category"
        }'::jsonb
                     )
WHERE type = 'component-template-fixer' AND deleted_at IS NULL;


UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        'source, domain, item_type',
        'source, pipeline, item_type'
                     )::jsonb
WHERE type = 'component-template-fixer';


