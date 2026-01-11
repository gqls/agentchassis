https://claude.ai/chat/c75db769-ab82-45f5-a4ed-8083bedc09ce
058_one_page_website

Good, I have enough information. Now I understand the structure. Let me create the configurations for a complete robot-hands.com website with:

Homepage with hero section and image
About page
Contact page

I need to:

Create an agent group definition with a workflow
Define agents for different sections
Use the image adapter for images

Let me start by designing the workflow first, then I'll create the configurations step by step.
Based on what I've seen:

We have a working robot-hands-website group already
We need to expand it to include about and contact pages
We should use content creators for each section
Use image generator for images
Build HTML pages with the content

Let me think about the workflow:

Spawn hero writer agent
Generate hero content
Spawn image creator agent
Generate hero image
Spawn about writer agent
Generate about content
Spawn contact writer agent
Generate contact content
Assemble homepage HTML
Assemble about page HTML
Assemble contact page HTML
Complete

Actually, I should be more efficient - spawn all agents first, then generate content, then assemble.
Let me create a cleaner workflow structure.

agent group definition for the complete website, then create the individual agent definitions that are needed.
The workflow should:

Spawn all agents first (hero_writer, image_creator, about_writer, contact_writer)
Generate content for each section (hero, image, about, contact)
Assemble the pages (homepage with hero+image, about page, contact page)
Complete

Let me create this step by step.


--

ALTER TABLE agent_group_definitions
ADD CONSTRAINT agent_group_unique_group_type_version UNIQUE (group_type, version);


-- Agent Group Definition for Robot Hands Complete Website
-- This orchestrates creation of homepage (hero + image), about page, and contact page

INSERT INTO agent_group_definitions (
id,
name,
group_type,
agent_configs,
orchestration_workflow,
version
) VALUES (
gen_random_uuid(),
'Robot Hands Complete Website',
'robot-hands-complete-website',

    -- Agent configurations: define roles and agent types needed
    '[
        {
            "role": "hero_writer",
            "agent_type": "content-creator-hero-without-research"
        },
        {
            "role": "image_creator",
            "agent_type": "image-generator"
        },
        {
            "role": "about_writer",
            "agent_type": "content-creator-about"
        },
        {
            "role": "contact_writer",
            "agent_type": "content-creator-contact"
        }
    ]'::jsonb,
    
    -- Orchestration workflow: defines the sequence of operations
    '{
        "start_step": "spawn_hero_writer",
        "steps": {
            
            "spawn_hero_writer": {
                "action": "spawn_agent",
                "description": "Spawn hero content writer",
                "config": {
                    "role": "hero_writer",
                    "agent_type": "content-creator-hero-without-research"
                },
                "next_step": "spawn_image_creator"
            },
            
            "spawn_image_creator": {
                "action": "spawn_agent",
                "description": "Spawn image generator",
                "config": {
                    "role": "image_creator",
                    "agent_type": "image-generator"
                },
                "next_step": "spawn_about_writer"
            },
            
            "spawn_about_writer": {
                "action": "spawn_agent",
                "description": "Spawn about page writer",
                "config": {
                    "role": "about_writer",
                    "agent_type": "content-creator-about"
                },
                "next_step": "spawn_contact_writer"
            },
            
            "spawn_contact_writer": {
                "action": "spawn_agent",
                "description": "Spawn contact page writer",
                "config": {
                    "role": "contact_writer",
                    "agent_type": "content-creator-contact"
                },
                "next_step": "generate_hero"
            },
            
            "generate_hero": {
                "action": "call_agent",
                "description": "Generate hero section content",
                "config": {
                    "agent_type": "content-creator-hero-without-research",
                    "target_role": "hero_writer",
                    "prompt": "Write a compelling hero section for {{.business_name}}, a company specializing in {{.business_type}}. Focus on precision robotics and automation. Include a powerful headline and engaging subheadline."
                },
                "next_step": "generate_hero_image"
            },
            
            "generate_hero_image": {
                "action": "call_agent",
                "description": "Generate hero background image",
                "config": {
                    "agent_type": "image-generator",
                    "target_role": "image_creator",
                    "prompt": "Professional photograph of precision robotic hands assembling electronic components in a modern factory setting, dramatic lighting with subtle blue tones, photorealistic, 8k, wide angle view",
                    "width": 1920,
                    "height": 1080
                },
                "next_step": "generate_about"
            },
            
            "generate_about": {
                "action": "call_agent",
                "description": "Generate about page content",
                "config": {
                    "agent_type": "content-creator-about",
                    "target_role": "about_writer",
                    "prompt": "Write an about page for {{.business_name}}, explaining that this website was created by AI agents working together. Include information about the agent orchestration system, how different agents collaborated to create content, generate images, and assemble the pages. Make it informative and interesting."
                },
                "next_step": "generate_contact"
            },
            
            "generate_contact": {
                "action": "call_agent",
                "description": "Generate contact page content",
                "config": {
                    "agent_type": "content-creator-contact",
                    "target_role": "contact_writer",
                    "prompt": "Write a contact page for {{.business_name}}. Include friendly text encouraging visitors to reach out, and placeholder contact information (email: contact@robot-hands.com, phone: +1-555-ROBOTS, address: Innovation District, Tech City)."
                },
                "next_step": "assemble_homepage"
            },
            
            "assemble_homepage": {
                "action": "aggregate_webpage",
                "description": "Assemble homepage HTML with hero section and image",
                "config": {
                    "wrapper": {
                        "html_head": "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>Robot Hands - Precision Robotics & Automation</title>\n<style>\n* { margin: 0; padding: 0; box-sizing: border-box; }\nbody { font-family: -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif; line-height: 1.6; color: #333; }\nnav { background: #1a1a2e; padding: 1rem 2rem; position: fixed; top: 0; width: 100%; z-index: 1000; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }\nnav ul { list-style: none; display: flex; gap: 2rem; align-items: center; }\nnav a { color: #fff; text-decoration: none; font-weight: 500; transition: color 0.3s; }\nnav a:hover { color: #4a9eff; }\n.hero-section { min-height: 100vh; display: flex; align-items: center; justify-content: center; position: relative; padding: 6rem 2rem 4rem; background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); color: white; text-align: center; }\n.hero-section::before { content: ''; position: absolute; top: 0; left: 0; right: 0; bottom: 0; background-size: cover; background-position: center; opacity: 0.3; z-index: 0; }\n.hero-content { position: relative; z-index: 1; max-width: 900px; }\n.hero-section h1 { font-size: 3.5rem; font-weight: 800; margin-bottom: 1.5rem; line-height: 1.2; }\n.hero-section p { font-size: 1.5rem; margin-bottom: 2rem; opacity: 0.9; }\n.cta-button { display: inline-block; padding: 1rem 2.5rem; background: #4a9eff; color: white; text-decoration: none; border-radius: 50px; font-weight: 600; font-size: 1.1rem; transition: all 0.3s; box-shadow: 0 4px 15px rgba(74, 158, 255, 0.4); }\n.cta-button:hover { background: #3a8eef; transform: translateY(-2px); box-shadow: 0 6px 20px rgba(74, 158, 255, 0.6); }\nfooter { background: #1a1a2e; color: white; text-align: center; padding: 2rem; margin-top: 4rem; }\n</style>\n</head>\n<body>\n<nav>\n<ul>\n<li><a href=\"index.html\">Home</a></li>\n<li><a href=\"about.html\">About</a></li>\n<li><a href=\"contact.html\">Contact</a></li>\n</ul>\n</nav>",
                        "html_foot": "<footer>\n<p>&copy; 2025 Robot Hands. Created by <a href=\"https://ai-agent-orchestration.com\" >AI Agent Orchestration</a>.</p>\n</footer>\n</body>\n</html>"
                    },
                    "section_order": ["generate_hero"],
                    "response_fields": ["generate_hero", "generate_hero_image"],
                    "add_section_tags": true,
                    "page_name": "index.html"
                },
                "next_step": "assemble_about_page"
            },
            
            "assemble_about_page": {
                "action": "aggregate_webpage",
                "description": "Assemble about page HTML",
                "config": {
                    "wrapper": {
                        "html_head": "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>About - Robot Hands</title>\n<style>\n* { margin: 0; padding: 0; box-sizing: border-box; }\nbody { font-family: -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif; line-height: 1.6; color: #333; background: #f5f7fa; }\nnav { background: #1a1a2e; padding: 1rem 2rem; position: fixed; top: 0; width: 100%; z-index: 1000; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }\nnav ul { list-style: none; display: flex; gap: 2rem; align-items: center; }\nnav a { color: #fff; text-decoration: none; font-weight: 500; transition: color 0.3s; }\nnav a:hover { color: #4a9eff; }\n.content { max-width: 900px; margin: 6rem auto 4rem; padding: 3rem; background: white; border-radius: 10px; box-shadow: 0 2px 20px rgba(0,0,0,0.1); }\n.content h1 { font-size: 2.5rem; color: #1a1a2e; margin-bottom: 2rem; }\n.content p { font-size: 1.1rem; margin-bottom: 1.5rem; line-height: 1.8; }\nfooter { background: #1a1a2e; color: white; text-align: center; padding: 2rem; margin-top: 4rem; }\n</style>\n</head>\n<body>\n<nav>\n<ul>\n<li><a href=\"index.html\">Home</a></li>\n<li><a href=\"about.html\">About</a></li>\n<li><a href=\"contact.html\">Contact</a></li>\n</ul>\n</nav>",
                        "html_foot": "<footer>\n<p>&copy; 2025 Robot Hands. Created by <a href=\"https://ai-agent-orchestration.com\" >AI Agent Orchestration</a>.</p>\n</footer>\n</body>\n</html>"
                    },
                    "section_order": ["generate_about"],
                    "response_fields": ["generate_about"],
                    "add_section_tags": false,
                    "page_name": "about.html"
                },
                "next_step": "assemble_contact_page"
            },
            
            "assemble_contact_page": {
                "action": "aggregate_webpage",
                "description": "Assemble contact page HTML",
                "config": {
                    "wrapper": {
                        "html_head": "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>Contact - Robot Hands</title>\n<style>\n* { margin: 0; padding: 0; box-sizing: border-box; }\nbody { font-family: -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif; line-height: 1.6; color: #333; background: #f5f7fa; }\nnav { background: #1a1a2e; padding: 1rem 2rem; position: fixed; top: 0; width: 100%; z-index: 1000; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }\nnav ul { list-style: none; display: flex; gap: 2rem; align-items: center; }\nnav a { color: #fff; text-decoration: none; font-weight: 500; transition: color 0.3s; }\nnav a:hover { color: #4a9eff; }\n.content { max-width: 900px; margin: 6rem auto 4rem; padding: 3rem; background: white; border-radius: 10px; box-shadow: 0 2px 20px rgba(0,0,0,0.1); }\n.content h1 { font-size: 2.5rem; color: #1a1a2e; margin-bottom: 2rem; }\n.content p { font-size: 1.1rem; margin-bottom: 1.5rem; line-height: 1.8; }\n.contact-info { background: #f5f7fa; padding: 2rem; border-radius: 8px; margin-top: 2rem; }\n.contact-info p { margin-bottom: 1rem; }\nfooter { background: #1a1a2e; color: white; text-align: center; padding: 2rem; margin-top: 4rem; }\n</style>\n</head>\n<body>\n<nav>\n<ul>\n<li><a href=\"index.html\">Home</a></li>\n<li><a href=\"about.html\">About</a></li>\n<li><a href=\"contact.html\">Contact</a></li>\n</ul>\n</nav>",
                        "html_foot": "<footer>\n<p>&copy; 2025 Robot Hands. Created by <a href=\"https://ai-agent-orchestration.com\" >AI Agent Orchestration</a>.</p>\n</footer>\n</body>\n</html>"
                    },
                    "section_order": ["generate_contact"],
                    "response_fields": ["generate_contact"],
                    "add_section_tags": false,
                    "page_name": "contact.html"
                },
                "next_step": "complete"
            },
            
            "complete": {
                "action": "complete_workflow",
                "description": "Return complete website with all pages"
            }
        }
    }'::jsonb,
    
    1
)
ON CONFLICT (group_type, version)
DO UPDATE SET
name = EXCLUDED.name,
agent_configs = EXCLUDED.agent_configs,
orchestration_workflow = EXCLUDED.orchestration_workflow,
updated_at = NOW();
    

--

-- Agent Definition: Content Creator for Contact Pages
-- This agent specializes in writing contact page content

INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
capabilities,
image_repository,
image_tag,
resources,
topics,
health_config,
env_vars,
version,
delegation_preferences
) VALUES (
gen_random_uuid(),
'content-creator-contact',
'Contact Page Writer',
'Specialized in writing welcoming and effective contact page content',
'data-driven',

    -- Default configuration with workflow
    '{
        "processing_mode": "task",
        "ai_service": {
            "provider": "anthropic",
            "model": "claude-haiku-4-5-20251001",
            "api_key_env_var": "ANTHROPIC_API_KEY"
        },
        "max_tokens": 1500,
        "temperature": 0.7,
        "workflow": {
            "start_step": "generate_contact_content",
            "steps": {
                "generate_contact_content": {
                    "action": "execute_llm_prompt",
                    "description": "Generate contact page content",
                    "config": {
                        "input_fields": ["input_data"],
                        "prompt_template": "{{.prompt}}\n\nCreate welcoming and clear content. Structure it with an introductory paragraph followed by contact details in a clean format. The contact details can be focused around how to contact the ai agents, or at least we can invent a customer contact (ai) agent. (There are not many humans involved in this venture)"
                    },
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Return contact page content"
                }
            }
        }
    }'::jsonb,
    
    true,  -- is_active
    
    '["content", "contact", "communication", "customer-engagement"]'::jsonb,
    
    'docker.io/aqls/agent-chassis',
    'v1.0.407',
    
    '{
        "requests": {
            "cpu": "100m",
            "memory": "256Mi"
        },
        "limits": {
            "cpu": "500m",
            "memory": "1Gi"
        }
    }'::jsonb,
    
    '{
        "process": "system.agent.content-creator-contact.process",
        "response": "system.responses.content-creator-contact",
        "error": "system.errors.content-creator-contact"
    }'::jsonb,
    
    '{
        "port": 8080,
        "liveness_path": "/health",
        "readiness_path": "/ready",
        "initial_delay_seconds": 30
    }'::jsonb,
    
    '[]'::jsonb,  -- env_vars
    
    2,  -- version
    
    '{
        "prefer_delegation": true,
        "fallback_to_self": true
    }'::jsonb
)
ON CONFLICT (type, version)
DO UPDATE SET
display_name = EXCLUDED.display_name,
description = EXCLUDED.description,
default_config = EXCLUDED.default_config,
capabilities = EXCLUDED.capabilities,
updated_at = NOW();


-- Agent Definition: Content Creator for About Pages
-- This agent specializes in writing about page content

INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
capabilities,
image_repository,
image_tag,
resources,
topics,
health_config,
env_vars,
version,
delegation_preferences
) VALUES (
gen_random_uuid(),
'content-creator-about',
'About Page Writer',
'Specialized in writing compelling about page content that tells the story of a business or explains a concept',
'data-driven',

    -- Default configuration with workflow
    '{
        "processing_mode": "task",
        "ai_service": {
            "provider": "anthropic",
            "model": "claude-haiku-4-5-20251001",
            "api_key_env_var": "ANTHROPIC_API_KEY"
        },
        "max_tokens": 2000,
        "temperature": 0.7,
        "workflow": {
            "start_step": "generate_about_content",
            "steps": {
                "generate_about_content": {
                    "action": "execute_llm_prompt",
                    "description": "Generate about page content",
                    "config": {
                        "input_fields": ["input_data"],
                        "prompt_template": "{{.prompt}}\n\nProvide well-structured content suitable for an about page. Use clear paragraphs and maintain a professional yet approachable tone. As a small addendum add that the site is created by ai agents and may be for sale."
                    },
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Return about page content"
                }
            }
        }
    }'::jsonb,
    
    true,  -- is_active
    
    '["content", "about", "storytelling", "brand-narrative"]'::jsonb,
    
    'docker.io/aqls/agent-chassis',
    'v1.0.407',
    
    '{
        "requests": {
            "cpu": "100m",
            "memory": "256Mi"
        },
        "limits": {
            "cpu": "500m",
            "memory": "1Gi"
        }
    }'::jsonb,
    
    '{
        "process": "system.agent.content-creator-about.process",
        "response": "system.responses.content-creator-about",
        "error": "system.errors.content-creator-about"
    }'::jsonb,
    
    '{
        "port": 8080,
        "liveness_path": "/health",
        "readiness_path": "/ready",
        "initial_delay_seconds": 30
    }'::jsonb,
    
    '[]'::jsonb,  -- env_vars
    
    1,  -- version
    
    '{
        "prefer_delegation": true,
        "fallback_to_self": true
    }'::jsonb
)
ON CONFLICT (type, version)
DO UPDATE SET
display_name = EXCLUDED.display_name,
description = EXCLUDED.description,
default_config = EXCLUDED.default_config,
capabilities = EXCLUDED.capabilities,
updated_at = NOW();