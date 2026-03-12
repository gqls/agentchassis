-- ============================================================================
-- css-patch-agent: applies targeted CSS fixes from audit findings
--
-- Unlike webdesign-agent (which regenerates everything from scratch), this
-- agent reads the current stylesheet, gets the LLM to apply a specific fix,
-- and deploys the patched CSS. No full regeneration.
--
-- Handles: spacing fixes, responsive fixes, layout fixes, variable fixes
-- Does NOT handle: full theme changes, colour palette redesigns (webdesign-agent)
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category, status,
    image_repository, image_tag,
    resources, default_config, input_contract, output_contract,
    domain_tags, agent_category, idle_timeout_seconds
) VALUES (
             'css-patch-agent',
             'CSS Patch Agent',
             'Applies targeted CSS fixes from audit findings. Reads the current stylesheet, uses LLM to generate a specific patch, and deploys. Does not regenerate the full theme.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.860',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "processing_mode": "orchestrator",
                     "timeout_seconds": 180,
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "load_current_css",
                             "output_field": "site_record"
                         },
                         "load_current_css": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT ct.id::text as theme_id, ct.css_content, ct.name as theme_name FROM css_themes ct JOIN style_collections sc ON sc.css_theme_id = ct.id JOIN sites s ON s.style_collection_id = sc.id WHERE s.id = $1::uuid",
                                 "params": ["site_record.site_id"],
                                 "output_format": "object"
                             },
                             "next_step": "check_has_css",
                             "error_step": "complete_no_css",
                             "description": "Load current CSS from css_themes via style_collection",
                             "output_field": "current_css"
                         },
                         "check_has_css": {
                             "action": "conditional",
                             "config": {
                                 "condition": "current_css.css_content != null",
                                 "then_step": "plan_css_fix",
                                 "else_step": "complete_no_css"
                             },
                             "description": "Skip if no CSS found (style_collection not linked)"
                         },
                         "plan_css_fix": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 8000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_record", "current_css", "input_data"],
                                 "output_format": "json",
                                 "prompt_template": "You are a CSS expert applying a targeted fix to an existing stylesheet.\n\n## Site\nDomain: {{.site_record.domain}}\n\n## Audit Finding\nCategory: {{.input_data.spec.category}}\nDescription: {{.input_data.spec.description}}\nSuggestion: {{.input_data.spec.suggestion}}\n{{if .input_data.spec.affected_component}}Affected component: {{.input_data.spec.affected_component}}{{end}}\n{{if .input_data.spec.page_name}}Page: {{.input_data.spec.page_name}}{{end}}\n\n## Current Stylesheet\n```css\n{{.current_css.css_content}}\n```\n\n## Instructions\nApply ONLY the fix described in the audit finding. Do not redesign or refactor unrelated CSS. Make the minimum change needed.\n\nFor spacing fixes: ensure consistent use of CSS variables or shared classes.\nFor responsive fixes: add or correct media queries.\nFor layout fixes: fix the specific layout issue described.\n\nReturn valid JSON:\n{\n  \"patched_css\": \"... the complete updated stylesheet with the fix applied ...\",\n  \"changes_summary\": \"Brief description of what was changed\",\n  \"lines_changed\": 3,\n  \"css_added\": \"... only the new/modified CSS rules (for logging) ...\"\n}"
                             },
                             "next_step": "save_css_to_db",
                             "error_step": "complete_error",
                             "description": "LLM applies targeted CSS fix",
                             "output_field": "css_fix"
                         },
                         "save_css_to_db": {
                             "action": "query_database",
                             "config": {
                                 "query": "UPDATE css_themes SET css_content = $2, updated_at = NOW(), version = version + 1 WHERE id = $1::uuid RETURNING id::text, version",
                                 "params": ["current_css.theme_id", "css_fix.result.patched_css"],
                                 "output_format": "object"
                             },
                             "next_step": "deploy_css",
                             "error_step": "complete_error",
                             "description": "Write patched CSS back to css_themes",
                             "output_field": "css_saved"
                         },
                         "deploy_css": {
                             "action": "git_commit",
                             "config": {
                                 "file_path": "assets/css/styles.css",
                                 "domain_field": "site_record.domain",
                                 "content_field": "css_fix.result.patched_css",
                                 "commit_message": "CSS fix: {{.input_data.spec.category}} — {{.css_fix.result.changes_summary}}"
                             },
                             "next_step": "complete",
                             "error_step": "complete_error",
                             "description": "Deploy patched CSS to git",
                             "output_field": "css_deployed"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["css_fix", "css_deployed"]
                             },
                             "description": "CSS fix applied and deployed"
                         },
                         "complete_no_css": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": [],
                                 "success_message": "No CSS theme linked to site — cannot patch"
                             },
                             "description": "Skip — no stylesheet found"
                         },
                         "complete_error": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["css_fix"],
                                 "success_message": "CSS fix failed"
                             },
                             "description": "Error path"
                         }
                     }
                 }
             }'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["spec"]}'::jsonb,
             '{"produces": {"css_fix": "LLM fix plan with patched CSS", "css_deployed": "git commit result"}}'::jsonb,
             '["css", "maintenance", "fix-agent"]'::jsonb,
             'specialist',
             120
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    display_name = EXCLUDED.display_name,
    image_tag = EXCLUDED.image_tag,
    idle_timeout_seconds = EXCLUDED.idle_timeout_seconds,
    updated_at = NOW();

-- ============================================================================
-- Reclassify needs_design_review items by category
-- ============================================================================

-- Spacing and responsive items → css-patch-agent
UPDATE site_work_items
SET handler_agent = 'css-patch-agent'
WHERE item_type = 'needs_design_review'
  AND handler_agent = 'webdesign-agent'
  AND spec->>'category' IN ('spacing', 'responsive', 'layout')
  AND status IN ('triaged', 'failed');

-- Colour items → color-variable-fixer (if any were missed)
UPDATE site_work_items
SET handler_agent = 'color-variable-fixer',
    item_type = 'hardcoded_section_colors'
WHERE item_type = 'needs_design_review'
  AND handler_agent = 'webdesign-agent'
  AND spec->>'category' = 'colour'
  AND status IN ('triaged', 'failed');

-- Reset failed items so they can retry with correct handler
UPDATE site_work_items
SET status = 'triaged', attempt_count = 0,
    claimed_by = NULL, claimed_at = NULL, error = NULL
WHERE item_type = 'needs_design_review'
  AND status = 'failed';

-- Also update the write_audit_findings classification so future audits
-- route correctly. This is for reference — the actual mapping is in Go code.
-- spacing/responsive/layout → css-patch-agent
-- colour → color-variable-fixer
-- full redesign → webdesign-agent (only for generic_theme items)

-- Verify the reclassification
SELECT wi.item_type, wi.handler_agent, wi.status,
       wi.spec->>'category' as category, COUNT(*)
FROM site_work_items wi
WHERE wi.item_type IN ('needs_design_review', 'hardcoded_section_colors')
  AND wi.domain = 'build'
  AND wi.status NOT IN ('complete', 'blocked')
GROUP BY wi.item_type, wi.handler_agent, wi.status, wi.spec->>'category'
ORDER BY wi.handler_agent, wi.item_type;