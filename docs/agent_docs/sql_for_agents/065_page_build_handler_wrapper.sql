-- The full page-build-handler workflow after all changes:
--
-- ensure_site_record
-- → load_page_record          (Go action — loads page from DB)
-- → check_page_found        (conditional — page exists?)
-- → plan_sections         (Go action — resolves data sources per section)
-- → check_has_ready_sections (conditional — any sections ready?)
-- → spawn_content_writer
-- → call_content_writer  (receives section_plan with ready sections + resolved data)
-- → check_content_produced (conditional — writer produced HTML?)
-- → validate_content     (Go action — placeholders, templates, contamination)
-- → save_sections      (reads validation_result.clean_html)
-- → update_status
-- → spawn_rerender_agent → deploy_page → complete
--
-- Error paths:
--   check_page_found (no page)        → complete_error
--   check_has_ready_sections (none)   → complete_error (deferred items created)
--   check_content_produced (skipped)  → complete_error
--   validate_content (blockers)       → mark_needs_review → complete_error

-- ============================================================================
-- page-build-handler: Wrapper handler for page-content-writer
--
-- Solves the specialist vs handler problem: page-content-writer generates
-- content but doesn't persist it. This handler spawns page-content-writer,
-- captures its output, then runs save_page_sections + update_page_status
-- to persist content to page_components.
--
-- No new Go code needed — uses existing actions in a workflow.
--
-- Pre-flight check (Step 0):
--   Existing actions used: ensure_site_record, spawn_agent, call_agent,
--   save_page_sections, update_page_status, page-rerender (agent),
--   complete_workflow. All exist and are registered.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'page-build-handler',
             'Page Build Handler',
             'Wrapper handler for page-content-writer. Spawns the specialist, captures output, persists sections to page_components via save_page_sections, updates page status, assembles and deploys via page-rerender. Dispatch-loop compatible.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.831', 'specialist',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"ensure_site_record":{"action":"ensure_site_record","config":{"store_brief_in_content_data":false},"next_step":"spawn_content_writer","description":"Load site record for context","output_field":"site_record"},"spawn_content_writer":{"action":"spawn_agent","config":{"role":"content_writer","agent_type":"page-content-writer"},"next_step":"call_content_writer","description":"Spawn page-content-writer specialist","output_field":"writer_agent"},"call_content_writer":{"action":"call_agent","config":{"target_role":"content_writer","input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain","current_page":"input_data.spec","site_record":"site_record","site_plan":"site_record.content_data","reviewed_brief?":"site_record.content_data.reviewed_brief"},"timeout_seconds":300},"next_step":"save_sections","error_step":"complete_error","description":"Call page-content-writer to generate content","output_field":"page_content"},"save_sections":{"action":"save_page_sections","config":{"html_field":"page_content.response.page_html","sections_metadata_field":"page_content.response.sections_metadata","page_name_field":"input_data.spec.page_name","site_id_field":"site_record.site_id"},"next_step":"update_status","error_step":"complete_error","description":"Persist generated sections to page_components","output_field":"sections_saved"},"update_status":{"action":"update_page_status","config":{"status":"deployed","page_name_field":"input_data.spec.page_name","site_id_field":"site_record.site_id"},"next_step":"spawn_rerender_agent","description":"Mark page as deployed","output_field":"status_updated"},"spawn_rerender_agent":{"action":"spawn_agent","config":{"role":"page_renderer","agent_type":"page-rerender"},"next_step":"deploy_page","description":"Spawn page-rerender for assembly and deploy","output_field":"rerender_agent"},"deploy_page":{"action":"call_agent","config":{"target_role":"page_renderer","input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain","page_id":"sections_saved.page_id"},"timeout_seconds":120},"next_step":"complete","error_step":"complete_error","description":"Assemble page from stored components and deploy to git","output_field":"deploy_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["sections_saved","deploy_result"]},"description":"Page build complete"},"complete_error":{"action":"complete_workflow","config":{"output_fields":["page_content","sections_saved"],"success_message":"Page build completed with errors"},"description":"Page build completed with errors"}}}}'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["page_name", "page_id", "sections"]}'::jsonb,
             '{"produces": {"sections_saved": "save result with page_id and section count", "deploy_result": "git commit result"}}'::jsonb,
             '["build", "content", "handler"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

-- Update existing needs_content_page items to use page-build-handler
UPDATE site_work_items
SET handler_agent = 'page-build-handler'
WHERE item_type = 'needs_content_page'
  AND handler_agent = 'page-content-writer'
  AND status IN ('detected', 'triaged', 'wont_fix');

-- Create a new blog content work item for finetuning.uk
-- (previous one was wont_fix because page-content-writer didn't persist)
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    page_id, priority, handler_agent, status, created_by, spec
) VALUES (
             '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
             'human', 'build', 'needs_content_page', 'high',
             'Generate content for blog page - currently empty',
             (SELECT id FROM pages WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND name = 'blog'),
             50, 'page-build-handler', 'triaged', 'human',
             '{"page_name": "blog", "sections": ["hero", "article-grid", "call-to-action"]}'::jsonb
         );

-- Verify
SELECT type, display_name, status
FROM agent_definitions WHERE type = 'page-build-handler' AND deleted_at IS NULL;

SELECT item_type, handler_agent, status
FROM site_work_items
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND item_type = 'needs_content_page'
ORDER BY created_at DESC LIMIT 2;

--

-- Add a check after call_content_writer
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_content_writer,next_step}',
        '"check_content_produced"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Add the conditional check
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_content_produced}',
        '{
            "action": "conditional",
            "config": {
                "condition": "page_content.response.skipped != true AND page_content.response.page_body != ",
                "then_step": "save_sections",
                "else_step": "complete_error"
            },
            "description": "Check if content writer produced content or skipped"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Update complete_error to be informative
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete_error,config}',
        '{"output_fields": ["page_content", "site_record"], "success_message": "Content writer skipped — page has no sections defined"}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Verify
SELECT
    default_config->'workflow'->'steps'->'call_content_writer'->>'next_step' as after_writer,
    default_config->'workflow'->'steps'->'check_content_produced' IS NOT NULL as has_guard
FROM agent_definitions
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- more

-- 1. Add guard after content writer (prevents crash)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_content_writer,next_step}',
        '"check_content_produced"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_content_produced}',
        '{
            "action": "conditional",
            "config": {
                "condition": "page_content.response.skipped != true",
                "then_step": "save_sections",
                "else_step": "complete_error"
            },
            "description": "Check if content writer produced content or skipped"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 2. Add page record lookup BEFORE calling content writer
-- Insert between ensure_site_record and spawn_content_writer
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ensure_site_record,next_step}',
        '"load_page_record"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_record}',
        '{
            "action": "query_database",
            "config": {
                "query": "SELECT id, name, title, page_type, sections::text, url, purpose FROM pages WHERE site_id = $1 AND name = $2 LIMIT 1",
                "params": ["site_record.site_id", "input_data.page_name"],
                "output_format": "single_row"
            },
            "next_step": "spawn_content_writer",
            "error_step": "spawn_content_writer",
            "description": "Load page record from DB to get sections, title, page_type",
            "output_field": "page_record"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 3. Update the content writer call to pass page_record as current_page
-- so it gets sections from the DB, not from the audit finding spec
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_content_writer,config,input_mapping,current_page}',
        '"page_record"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 4. Also update deploy_page to get page_id from page_record
-- (currently looks at sections_saved.page_id which might not exist)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_page,config,input_mapping,page_id}',
        '"page_record.id"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Verify the chain
SELECT
    default_config->'workflow'->'steps'->'ensure_site_record'->>'next_step' as s1,
    default_config->'workflow'->'steps'->'load_page_record'->>'next_step' as s2,
    default_config->'workflow'->'steps'->'call_content_writer'->>'next_step' as s3,
    default_config->'workflow'->'steps'->'check_content_produced'->>'next_step' as s4,
    default_config->'workflow'->'steps'->'call_content_writer'->'config'->'input_mapping'->>'current_page' as writer_page_source
FROM agent_definitions
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Expected: s1=load_page_record, s2=spawn_content_writer, s3=check_content_produced, s4=save_sections
-- writer_page_source=page_record

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_record,config,query}',
        '"SELECT id, name, title, page_type, sections::text, url FROM pages WHERE site_id = $1 AND name = $2 LIMIT 1"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

--
-- remove sql from workflow definition

-- ============================================================================
-- Update page-build-handler to use load_page_record Go action
-- instead of inline query_database SQL.
--
-- Run AFTER deploying load_page_record_action.go + registry entry.
-- ============================================================================

-- Replace the load_page_record step (was query_database with inline SQL)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_record}',
        '{
            "action": "load_page_record",
            "config": {
                "site_id": "site_record.site_id",
                "page_name": "input_data.page_name"
            },
            "next_step": "check_page_found",
            "description": "Load page record from DB — sections, title, page_type",
            "output_field": "page_record"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Add guard: if page not found, go to complete_error (not crash)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_page_found}',
        '{
            "action": "conditional",
            "config": {
                "condition": "page_record.found == true",
                "then_step": "spawn_content_writer",
                "else_step": "complete_error"
            },
            "description": "Check if page exists in DB — audit findings for new pages will skip here"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Update content writer input_mapping to use page_record instead of input_data.spec
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_content_writer,config,input_mapping,current_page}',
        '"page_record"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Update deploy_page to use page_record.id instead of sections_saved.page_id
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_page,config,input_mapping,page_id}',
        '"page_record.id"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Verify the full chain
SELECT
    default_config->'workflow'->'steps'->'ensure_site_record'->>'next_step' as s1_next,
    default_config->'workflow'->'steps'->'load_page_record'->>'action' as s2_action,
    default_config->'workflow'->'steps'->'load_page_record'->>'next_step' as s2_next,
    default_config->'workflow'->'steps'->'check_page_found'->'config'->>'then_step' as s3_then,
    default_config->'workflow'->'steps'->'check_page_found'->'config'->>'else_step' as s3_else,
    default_config->'workflow'->'steps'->'call_content_writer'->'config'->'input_mapping'->>'current_page' as writer_page_src,
    default_config->'workflow'->'steps'->'deploy_page'->'config'->'input_mapping'->>'page_id' as deploy_page_id_src
FROM agent_definitions
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Expected:
-- s1_next = load_page_record
-- s2_action = load_page_record (Go action, not query_database)
-- s2_next = check_page_found
-- s3_then = spawn_content_writer
-- s3_else = complete_error
-- writer_page_src = page_record
-- deploy_page_id_src = page_record.id


---

-- ============================================================================
-- Wire plan_sections into page-build-handler
--
-- Flow after this change:
--
--   ensure_site_record
--     → load_page_record
--       → check_page_found
--         → plan_sections              ← NEW
--           → check_has_ready_sections ← NEW
--             → spawn_content_writer
--               → call_content_writer (receives section_plan.sections_ready)
--                 → validate_content
--                   → save_sections → update_status → deploy
--
-- The content writer now receives only sections that have data available.
-- Deferred sections become needs_human_review work items.
-- Skipped sections are silently omitted.
--
-- Run AFTER deploying plan_sections_action.go + registry entry
-- and the input_schema v2 migration.
-- ============================================================================

-- 1. Add plan_sections step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_sections}',
        '{
            "action": "plan_sections",
            "config": {
                "site_id": "site_record.site_id",
                "sections": "page_record.sections",
                "page_name": "page_record.name",
                "domain": "site_record.domain",
                "work_item_id": "input_data.work_item_id"
            },
            "next_step": "check_has_ready_sections",
            "error_step": "complete_error",
            "description": "Check data availability for each section, triage readiness",
            "output_field": "section_plan"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 2. Add check_has_ready_sections guard
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_has_ready_sections}',
        '{
            "action": "conditional",
            "config": {
                "condition": "section_plan.ready_count > 0",
                "then_step": "spawn_content_writer",
                "else_step": "complete_error"
            },
            "description": "If no sections are ready (all deferred or skipped), skip content generation"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 3. Route check_page_found → plan_sections (was → spawn_content_writer)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_page_found,config,then_step}',
        '"plan_sections"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 4. Update content writer input_mapping to pass section_plan
--    The content writer receives:
--      current_page: page_record (for page-level context)
--      section_plan: the ready sections with resolved data and LLM field lists
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_content_writer,config,input_mapping,section_plan}',
        '"section_plan"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 5. Verify the full chain
SELECT
    default_config->'workflow'->'steps'->'ensure_site_record'->>'next_step' as s1,
    default_config->'workflow'->'steps'->'load_page_record'->>'next_step' as s2,
    default_config->'workflow'->'steps'->'check_page_found'->'config'->>'then_step' as s3_then,
    default_config->'workflow'->'steps'->'plan_sections'->>'next_step' as s4,
    default_config->'workflow'->'steps'->'check_has_ready_sections'->'config'->>'then_step' as s5_then,
    default_config->'workflow'->'steps'->'check_has_ready_sections'->'config'->>'else_step' as s5_else,
    default_config->'workflow'->'steps'->'call_content_writer'->'config'->'input_mapping'->>'section_plan' as writer_gets_plan,
    default_config->'workflow'->'steps'->'validate_content'->>'next_step' as after_validate
FROM agent_definitions
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Expected:
-- s1 = load_page_record
-- s2 = check_page_found
-- s3_then = plan_sections
-- s4 = check_has_ready_sections
-- s5_then = spawn_content_writer
-- s5_else = complete_error
-- writer_gets_plan = section_plan
-- after_validate = save_sections



-- plan_sections          → prevents fabrication (don't ask for what you can't fill)
-- ↓
-- content writer         → generates HTML for ready sections only
-- ↓
-- validate_page_content  → catches LLM mistakes (placeholders, templates, contamination)
-- ↓
-- mark_needs_review      → routes failures to HITL (needs the status_override patch)
-- ↓
-- save_sections          → only reached if both checks pass

-- more of the same

-- ============================================================================
-- Wire validate_content into page-build-handler
--
-- This runs AFTER add_plan_sections_step.sql has already been applied.
-- The chain at that point is:
--
--   ... → check_content_produced → save_sections → ...
--
-- After this change:
--
--   ... → check_content_produced → validate_content → save_sections → ...
--                                       ↓ (blockers)
--                                  mark_needs_review → complete_error
--
-- validate_content catches LLM output problems:
--   - Placeholder text ("To be added", "Lorem ipsum")
--   - Unrendered templates ({{.field}})
--   - Cross-site contamination (wrong domain)
--
-- plan_sections (already wired) handles missing data upstream.
-- This is the safety net for problems that happen DURING generation.
--
-- Run AFTER:
--   1. add_plan_sections_step.sql
--   2. Deploying validate_page_content_action.go
--   3. Deploying fail_work_item status_override patch
-- ============================================================================

-- 1. Add validate_content step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,validate_content}',
        '{
            "action": "validate_page_content",
            "config": {
                "html_field": "page_content.response.page_html",
                "domain": "site_record.domain"
            },
            "next_step": "save_sections",
            "error_step": "mark_needs_review",
            "description": "Check for placeholders, unrendered templates, cross-site contamination",
            "output_field": "validation_result"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 2. Route check_content_produced → validate_content (was → save_sections)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_content_produced,config,then_step}',
        '"validate_content"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 3. Update save_sections to read cleaned HTML from validation_result
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,save_sections,config,html_field}',
        '"validation_result.clean_html"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 4. Add mark_needs_review step (uses fail_work_item with status_override)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,mark_needs_review}',
        '{
            "action": "fail_work_item",
            "config": {
                "work_item_id": "input_data.work_item_id",
                "error_message": "Content validation failed — needs human review",
                "status_override": "needs_human_review"
            },
            "next_step": "complete_error",
            "description": "Mark work item for human review when content has placeholder or contamination issues",
            "output_field": "review_result"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- 5. Verify
SELECT
    default_config->'workflow'->'steps'->'check_content_produced'->'config'->>'then_step' as after_check_content,
    default_config->'workflow'->'steps'->'validate_content'->>'next_step' as after_validate,
    default_config->'workflow'->'steps'->'validate_content'->>'error_step' as validate_error,
    default_config->'workflow'->'steps'->'save_sections'->'config'->>'html_field' as save_reads_from,
    default_config->'workflow'->'steps'->'mark_needs_review'->'config'->>'status_override' as review_status
FROM agent_definitions
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Expected:
-- after_check_content = validate_content
-- after_validate = save_sections
-- validate_error = mark_needs_review
-- save_reads_from = validation_result.clean_html
-- review_status = needs_human_review

---

-- timeout issues
-- Check current timeout
SELECT default_config->'workflow'->'steps'->'call_content_writer'->'config'->>'timeout_seconds' as current_timeout
FROM agent_definitions WHERE type = 'page-build-handler';

-- Increase call_content_writer timeout from 300 to 900
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"timeout_seconds": 300}, "next_step": "check_content_produced"',
        '"timeout_seconds": 900}, "next_step": "check_content_produced"'
                     )::jsonb
WHERE type = 'page-build-handler';

-- Verify
SELECT default_config->'workflow'->'steps'->'call_content_writer'->'config'->>'timeout_seconds' as new_timeout
FROM agent_definitions WHERE type = 'page-build-handler';

--

-- Increase page-build-handler workflow timeout to match
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,timeout_seconds}',
        '1800'
                     )
WHERE type = 'page-build-handler';

-- Verify
SELECT default_config->'workflow'->'steps'->'call_content_writer'->'config'->>'timeout_seconds' as step_timeout,
    default_config->'workflow'->>'timeout_seconds' as workflow_timeout
FROM agent_definitions WHERE type = 'page-build-handler';