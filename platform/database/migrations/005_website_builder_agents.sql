-- Insert website builder specific agent definitions
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
-- Domain Analyst
('domain-analyst', 'Domain Analyst', 'Analyzes domains and determines appropriate website type', 'data-driven',
 '{
     "model": "claude-3-5-sonnet-20241022",
     "temperature": 0.3,
     "workflow": {
         "start_step": "analyze_domain",
         "steps": {
             "analyze_domain": {
                 "action": "validate_input",
                 "next_step": "research_industry"
             },
             "research_industry": {
                 "action": "call_agent",
                 "agent_type": "web-search",
                 "topic": "system.adapter.web.search",
                 "next_step": "categorize_business"
             },
             "categorize_business": {
                 "action": "llm_analyze",
                 "config": {"prompt_template": "domain_categorization"},
                 "next_step": "complete",
                 "store_memory": true
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }'),

-- Site Architect
('site-architect', 'Site Architect', 'Plans website structure and navigation', 'data-driven',
 '{
     "model": "claude-3-5-sonnet-20241022",
     "temperature": 0.5,
     "workflow": {
         "start_step": "analyze_requirements",
         "steps": {
             "analyze_requirements": {
                 "action": "validate_input",
                 "next_step": "determine_pages"
             },
             "determine_pages": {
                 "action": "llm_generate",
                 "config": {"template": "site_structure"},
                 "next_step": "pause_for_human"
             },
             "pause_for_human": {
                 "action": "pause_for_human_input",
                 "config": {"message": "Review and approve site structure"},
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }'),

-- HTML Developer
('html-developer', 'HTML Developer', 'Generates HTML/CSS/JS code for websites', 'code-driven',
 '{
     "model": "claude-3-5-sonnet-20241022",
     "temperature": 0.2,
     "workflow": {
         "start_step": "receive_specs",
         "steps": {
             "receive_specs": {
                 "action": "validate_input",
                 "next_step": "generate_template"
             },
             "generate_template": {
                 "action": "code_generate",
                 "config": {
                     "language": "html",
                     "framework": "vanilla",
                     "responsive": true
                 },
                 "next_step": "create_pages"
             },
             "create_pages": {
                 "action": "fan_out",
                 "config": {"per_page": true},
                 "next_step": "bundle_site"
             },
             "bundle_site": {
                 "action": "package_files",
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }'),

-- Visual Designer
('visual-designer', 'Visual Designer', 'Handles images, logos, and visual assets', 'adapter',
 '{
     "workflow": {
         "start_step": "analyze_brand",
         "steps": {
             "analyze_brand": {
                 "action": "validate_input",
                 "next_step": "search_images"
             },
             "search_images": {
                 "action": "call_agent",
                 "agent_type": "image-search",
                 "config": {"sources": ["unsplash", "pexels"]},
                 "next_step": "create_logo"
             },
             "create_logo": {
                 "action": "call_agent",
                 "agent_type": "image-generator",
                 "topic": "system.adapter.image.generate",
                 "next_step": "optimize_images"
             },
             "optimize_images": {
                 "action": "process_images",
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }'),

-- Site Publisher
('site-publisher', 'Site Publisher', 'Publishes websites to storage buckets', 'adapter',
 '{
     "workflow": {
         "start_step": "collect_assets",
         "steps": {
             "collect_assets": {
                 "action": "validate_input",
                 "next_step": "organize_files"
             },
             "organize_files": {
                 "action": "organize_structure",
                 "next_step": "upload_to_bucket"
             },
             "upload_to_bucket": {
                 "action": "s3_upload",
                 "config": {
                     "bucket": "${SITE_BUCKET}",
                     "public": true
                 },
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }'),

-- Website Builder Orchestrator
('website-builder', 'Website Builder', 'Orchestrates complete website creation', 'code-driven',
 '{
     "workflow": {
         "start_step": "validate_request",
         "steps": {
             "validate_request": {
                 "action": "validate_input",
                 "next_step": "spawn_agents"
             },
             "spawn_agents": {
                 "action": "spawn_group",
                 "config": {"group_type": "website-builder"},
                 "next_step": "analyze_domain"
             },
             "analyze_domain": {
                 "action": "call_agent",
                 "agent_type": "domain-analyst",
                 "next_step": "architect_site"
             },
             "architect_site": {
                 "action": "call_agent",
                 "agent_type": "site-architect",
                 "next_step": "gather_content"
             },
             "gather_content": {
                 "action": "fan_out",
                 "sub_tasks": [
                     {"agent_type": "content-researcher", "step_name": "research"},
                     {"agent_type": "visual-designer", "step_name": "visuals"}
                 ],
                 "next_step": "develop_site"
             },
             "develop_site": {
                 "action": "call_agent",
                 "agent_type": "html-developer",
                 "next_step": "publish_site"
             },
             "publish_site": {
                 "action": "call_agent",
                 "agent_type": "site-publisher",
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }')
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              updated_at = NOW();