-- page-content-writer — wire in the internal-link-resolver sub-agent.
--
-- Changes (all in default_config.workflow.steps):
--   1. NEW spawn_link_resolver  — spawn the resolver alongside research-agent.
--   2. spawn_research_agent.next_step: load_site_specs -> spawn_link_resolver.
--   3. NEW resolve_links        — call_agent once per page, AFTER
--      build_render_context, passing the section plan; error_step falls
--      through so a resolver failure cannot block the build.
--   4. NEW select_sections      — extract_fields with a FALLBACK CHAIN
--      (the proven map-of-arrays format, cf. research-agent extract_topic):
--      use the resolver's augmented sections, else the original plan.
--      This makes the wiring regression-safe: if the resolver is down or
--      returns nothing, behaviour is byte-identical to today.
--   5. build_render_context.next_step: process_sections_loop -> resolve_links.
--   6. process_sections_loop.config.iterate_over:
--      input_data.section_plan.sections_ready -> sections_for_render.sections_ready.
--
-- input_mapping uses "?" optional suffixes (003: input_mapping fields are
-- required by default) for everything except site_id — section_plan is absent
-- in some flows (e.g. page-rebuild's call), and the fallback chain preserves
-- today's behaviour there.
--
-- Apply AFTER internal_link_resolver_agent.sql (the call target must exist) and
-- after the batch image is live. Snapshot first via the house helper (v2_40:
-- do NOT hand-roll a backup table).

SELECT snapshot_agent('page-content-writer', 'linking: wire internal-link-resolver (spawn/call/select/iterate)');

UPDATE agent_definitions
SET default_config =
    jsonb_set(
      jsonb_set(
        jsonb_set(
          jsonb_set(
            jsonb_set(
              jsonb_set(
                default_config,
                '{workflow,steps,spawn_link_resolver}',
                $s1$
                {
                    "action": "spawn_agent",
                    "config": {
                        "role": "link_resolver",
                        "agent_type": "internal-link-resolver",
                        "await_response": true
                    },
                    "next_step": "load_site_specs",
                    "description": "Spawn internal link resolver agent",
                    "output_field": "link_resolver_info"
                }
                $s1$::jsonb,
                true),
              '{workflow,steps,spawn_research_agent,next_step}',
              '"spawn_link_resolver"'),
            '{workflow,steps,resolve_links}',
            $s2$
            {
                "action": "call_agent",
                "config": {
                    "agent_type": "internal-link-resolver",
                    "target_role": "link_resolver",
                    "input_mapping": {
                        "site_id": "input_data.site_record.site_id",
                        "page_type?": "input_data.current_page.page_type",
                        "page_name?": "input_data.current_page.name",
                        "sections?": "input_data.section_plan.sections_ready"
                    },
                    "timeout_seconds": 60
                },
                "next_step": "select_sections",
                "error_step": "select_sections",
                "description": "Resolve internal CTA destinations from real pages",
                "output_field": "resolved_links"
            }
            $s2$::jsonb,
            true),
          '{workflow,steps,select_sections}',
          $s3$
          {
              "action": "extract_fields",
              "config": {
                  "fields": {
                      "sections_ready": [
                          "resolved_links.sections_ready",
                          "input_data.section_plan.sections_ready"
                      ]
                  }
              },
              "next_step": "process_sections_loop",
              "description": "Use resolver-augmented sections, falling back to the original plan",
              "output_field": "sections_for_render"
          }
          $s3$::jsonb,
          true),
        '{workflow,steps,build_render_context,next_step}',
        '"resolve_links"'),
      '{workflow,steps,process_sections_loop,config,iterate_over}',
      '"sections_for_render.sections_ready"'),
    updated_at = now()
WHERE type = 'page-content-writer';

-- Verify: chain order, new steps present, loop repointed.
SELECT
    default_config->'workflow'->'steps'->'spawn_research_agent'->>'next_step'  AS research_next,   -- spawn_link_resolver
    default_config->'workflow'->'steps'->'spawn_link_resolver'->>'next_step'   AS resolver_spawn_next, -- load_site_specs
    default_config->'workflow'->'steps'->'build_render_context'->>'next_step'  AS render_ctx_next,  -- resolve_links
    default_config->'workflow'->'steps'->'resolve_links'->>'next_step'         AS resolve_next,     -- select_sections
    default_config->'workflow'->'steps'->'resolve_links'->>'error_step'        AS resolve_error,    -- select_sections
    default_config->'workflow'->'steps'->'select_sections'->>'next_step'       AS select_next,      -- process_sections_loop
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->>'iterate_over' AS loop_iterates  -- sections_for_render.sections_ready
FROM agent_definitions
WHERE type = 'page-content-writer';

-- Rollback: SELECT revert_agent('page-content-writer');
