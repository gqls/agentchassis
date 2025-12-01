-- multi_page_support.sql
-- Adds a new multi-page site builder workflow (index + about + contact)

-- Step 1: Insert new agent group definition for multi-page sites
INSERT INTO agent_group_definitions (
    id,
    name,
    group_type,
    description,
    agent_configs,
    orchestration_workflow,
    version
) VALUES (
             gen_random_uuid(),
             'Multi-Page Site Builder',
             'multipage-site-builder',
             'Builds a 3-page site (index, about, contact) with landing page, generates content, and deploys to Git/B2.',
             '[
               {"role": "strategist", "agent_type": "site-strategist"},
               {"role": "architect", "agent_type": "landing-page-architect"},
               {"role": "writer", "agent_type": "content-writer"},
               {"role": "deployer", "agent_type": "site-deployer"}
             ]'::jsonb,
             '{
               "start_step": "spawn_strategist",
               "steps": {
                 "spawn_strategist": {
                   "action": "spawn_agent",
                   "config": {
                     "role": "strategist",
                     "agent_type": "site-strategist"
                   },
                   "next_step": "spawn_architect",
                   "description": "Spawn Site Strategist"
                 },
                 "spawn_architect": {
                   "action": "spawn_agent",
                   "config": {
                     "role": "architect",
                     "agent_type": "landing-page-architect"
                   },
                   "next_step": "spawn_writer",
                   "description": "Spawn Landing Page Architect"
                 },
                 "spawn_writer": {
                   "action": "spawn_agent",
                   "config": {
                     "role": "writer",
                     "agent_type": "content-writer"
                   },
                   "next_step": "spawn_deployer",
                   "description": "Spawn Content Writer"
                 },
                 "spawn_deployer": {
                   "action": "spawn_agent",
                   "config": {
                     "role": "deployer",
                     "agent_type": "site-deployer"
                   },
                   "next_step": "call_strategist",
                   "description": "Spawn Site Deployer"
                 },
                 "call_strategist": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "site-strategist",
                     "target_role": "strategist",
                     "timeout_seconds": 120
                   },
                   "next_step": "call_architect",
                   "description": "Get the Build Plan from the Strategist",
                   "output_field": "build_plan"
                 },
                 "call_architect": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "landing-page-architect",
                     "target_role": "architect",
                     "input_fields": ["build_plan", "input_data"],
                     "timeout_seconds": 120
                   },
                   "next_step": "call_writer",
                   "description": "Build the site template from Build Plan",
                   "output_field": "template_data"
                 },
                 "call_writer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "content-writer",
                     "target_role": "writer",
                     "input_fields": ["template_data", "build_plan", "input_data"],
                     "timeout_seconds": 300
                   },
                   "next_step": "wrap_multipage",
                   "description": "Generate content and assemble HTML",
                   "output_field": "final_html"
                 },
                 "wrap_multipage": {
                   "action": "wrap_multipage",
                   "config": {
                     "index_html_field": "final_html.assemble_html.final_html"
                   },
                   "next_step": "call_deployer",
                   "description": "Create about and contact pages, wrap into files map",
                   "output_field": "site_files"
                 },
                 "call_deployer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "site-deployer",
                     "target_role": "deployer",
                     "input_fields": ["site_files", "input_data"],
                     "timeout_seconds": 180
                   },
                   "next_step": "complete",
                   "description": "Deploy all pages to Git"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Multi-page site build complete"
                 }
               }
             }'::jsonb,
             1
         );


-- Step 2: Update site-deployer to support files_field (with backward compatibility)
-- The Go code falls back to content_field if files_field is not found,
-- so this won't break existing mvp-site-builder workflow

UPDATE agent_definitions
SET default_config = '{
  "processing_mode": "task",
  "timeout_seconds": 180,
  "workflow": {
    "start_step": "deploy_to_git",
    "steps": {
      "deploy_to_git": {
        "action": "git_commit",
        "config": {
          "repo_name": "sites",
          "domain_field": "domain",
          "files_field": "site_files.files",
          "content_field": "input_data.final_html.assemble_html.final_html",
          "commit_message": "Update site: {{.domain}}"
        },
        "next_step": "complete",
        "description": "Commit pages to Git repository"
      },
      "complete": {
        "action": "complete_workflow",
        "description": "Deployment complete"
      }
    }
  }
}'::jsonb
WHERE type = 'site-deployer';


-- Step 3: Verify the new group was created
SELECT
    group_type,
    name,
    description,
    orchestration_workflow->'steps'->'wrap_multipage' as wrap_step
FROM agent_group_definitions
WHERE group_type = 'multipage-site-builder';

-- Verify site-deployer config
SELECT
    type,
    default_config->'workflow'->'steps'->'deploy_to_git'->'config' as git_config
FROM agent_definitions
WHERE type = 'site-deployer';