-- multi_page_support.sql
-- Adds a new multi-page site builder workflow (index + about + contact)

-- Step 1: Insert new agent group definition for multi-page sites
-- old --
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


--
updated added html builder (version 2)

https://claude.ai/chat/a36b6fe1-efa1-4d53-b30e-768ab6c9bf68
        -- old --
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
    {"role": "html_assembler", "agent_type": "html-assembler"},
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
        "next_step": "spawn_html_assembler",
        "description": "Spawn Content Writer"
      },
      "spawn_html_assembler": {
        "action": "spawn_agent",
        "config": {
          "role": "html_assembler",
          "agent_type": "html-assembler"
        },
        "next_step": "spawn_deployer",
        "description": "Spawn HTML Assembler"
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
        "next_step": "call_html_assembler",
        "description": "Generate content JSON",
        "output_field": "content_data"
      },
      "call_html_assembler": {
        "action": "call_agent",
        "config": {
          "agent_type": "html-assembler",
          "target_role": "html_assembler",
          "input_fields": ["template_data", "content_data", "input_data"],
          "timeout_seconds": 120
        },
        "next_step": "wrap_multipage",
        "description": "Assemble final HTML from template and content",
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
  2
);


--

better - wrap is not a local action here:
       -- Updated multipage-site-builder workflow
-- Uses multipage-wrapper as an agent instead of local action

-- First, run multipage_wrapper_agent.sql to create the agent

-- Then update the workflow:
UPDATE agent_group_definitions
SET
    agent_configs = '[
    {"role": "strategist", "agent_type": "site-strategist"},
    {"role": "architect", "agent_type": "landing-page-architect"},
    {"role": "writer", "agent_type": "content-writer"},
    {"role": "html_assembler", "agent_type": "html-assembler"},
    {"role": "multipage_wrapper", "agent_type": "multipage-wrapper"},
    {"role": "deployer", "agent_type": "site-deployer"}
  ]'::jsonb,
  orchestration_workflow = '{
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
        "next_step": "spawn_html_assembler",
        "description": "Spawn Content Writer"
      },
      "spawn_html_assembler": {
        "action": "spawn_agent",
        "config": {
          "role": "html_assembler",
          "agent_type": "html-assembler"
        },
        "next_step": "spawn_multipage_wrapper",
        "description": "Spawn HTML Assembler"
      },
      "spawn_multipage_wrapper": {
        "action": "spawn_agent",
        "config": {
          "role": "multipage_wrapper",
          "agent_type": "multipage-wrapper"
        },
        "next_step": "spawn_deployer",
        "description": "Spawn Multi-Page Wrapper"
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
        "next_step": "call_html_assembler",
        "description": "Generate content JSON",
        "output_field": "content_data"
      },
      "call_html_assembler": {
        "action": "call_agent",
        "config": {
          "agent_type": "html-assembler",
          "target_role": "html_assembler",
          "input_fields": ["template_data", "content_data", "input_data"],
          "timeout_seconds": 120
        },
        "next_step": "call_multipage_wrapper",
        "description": "Assemble final HTML from template and content",
        "output_field": "final_html"
      },
      "call_multipage_wrapper": {
        "action": "call_agent",
        "config": {
          "agent_type": "multipage-wrapper",
          "target_role": "multipage_wrapper",
          "input_fields": ["final_html", "input_data"],
          "timeout_seconds": 60
        },
        "next_step": "call_deployer",
        "description": "Create about and contact pages",
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
  }'::jsonb
WHERE group_type = 'multipage-site-builder';

-- Verify
SELECT
    group_type,
    name,
    jsonb_array_length(agent_configs) as agent_count,
    orchestration_workflow->'steps'->'spawn_multipage_wrapper' as wrapper_spawn,
    orchestration_workflow->'steps'->'call_multipage_wrapper' as wrapper_call
FROM agent_group_definitions
WHERE group_type = 'multipage-site-builder';


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


-- Fix the deployer files_field to match actual data path
-- The wrap_multipage local action stores its result under the step name

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_to_git,config,files_field}',
        '"input_data.site_files.wrap_multipage.files"'::jsonb
                     )
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


-- above update code as an insert --
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
               {"role": "html_assembler", "agent_type": "html-assembler"},
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
        "next_step": "spawn_html_assembler",
        "description": "Spawn Content Writer"
      },
      "spawn_html_assembler": {
        "action": "spawn_agent",
        "config": {
          "role": "html_assembler",
          "agent_type": "html-assembler"
        },
        "next_step": "spawn_multipage_wrapper",
        "description": "Spawn HTML Assembler"
      },
      "spawn_multipage_wrapper": {
        "action": "spawn_agent",
        "config": {
          "role": "multipage_wrapper",
          "agent_type": "multipage-wrapper"
        },
        "next_step": "spawn_deployer",
        "description": "Spawn Multi-Page Wrapper"
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
        "next_step": "call_html_assembler",
        "description": "Generate content JSON",
        "output_field": "content_data"
      },
      "call_html_assembler": {
        "action": "call_agent",
        "config": {
          "agent_type": "html-assembler",
          "target_role": "html_assembler",
          "input_fields": ["template_data", "content_data", "input_data"],
          "timeout_seconds": 120
        },
        "next_step": "call_multipage_wrapper",
        "description": "Assemble final HTML from template and content",
        "output_field": "final_html"
      },
      "call_multipage_wrapper": {
        "action": "call_agent",
        "config": {
          "agent_type": "multipage-wrapper",
          "target_role": "multipage_wrapper",
          "input_fields": ["final_html", "input_data"],
          "timeout_seconds": 60
        },
        "next_step": "call_deployer",
        "description": "Create about and contact pages",
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


---

some more themes

INSERT INTO css_themes (
    name,
    display_name,
    description,
    category,
    semantic_tags,
    css_content
) VALUES
(
    'modern-engineering-clean',
    'Modern Engineering',
    'A precise, architectural design using cool grays, deep blues, and structural layouts. Professional and authoritative.',
    'modern',
    '{professional,clean,corporate,saas,trust}',
    ':root {
        /* Palette: Precision & Trust */
        --color-primary: #0f172a;        /* Slate 900 */
        --color-primary-hover: #334155;  /* Slate 700 */
        --color-primary-text: #ffffff;

        --color-secondary: #0ea5e9;      /* Sky 500 - used sparingly for active states */
        --color-secondary-hover: #0284c7;
        --color-secondary-text: #ffffff;

        --color-accent: #64748b;         /* Slate 500 - subtle accent */

        --color-text: #334155;           /* Slate 700 - softer than black */
        --color-text-muted: #64748b;     /* Slate 500 */
        --color-heading: #020617;        /* Slate 950 */

        --color-background: #ffffff;
        --color-background-alt: #f8fafc; /* Slate 50 */

        --color-border: #e2e8f0;         /* Slate 200 */

        /* Specialized Areas */
        --color-header-bg: rgba(255, 255, 255, 0.9);
        --color-header-text: #0f172a;

        --color-hero-title: #0f172a;
        --color-hero-subtitle: #475569;

        --color-card-bg: #ffffff;

        /* Gradient is subtle, almost metallic */
        --color-cta-bg: #0f172a;
        --color-cta-text: #ffffff;

        --color-footer-bg: #f8fafc;
        --color-footer-text: #475569;

        /* Design Tokens */
        --border-radius: 6px; /* Tighter radius for precision look */
        --shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
        --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
        --font-sans: "Inter", system-ui, -apple-system, sans-serif;
    }

    body {
        font-family: var(--font-sans);
        background-color: var(--color-background);
        color: var(--color-text);
        line-height: 1.6;
        -webkit-font-smoothing: antialiased;
    }

    /* Subtle backdrop blur for a modern glass feel on headers */
    header {
        backdrop-filter: blur(8px);
        border-bottom: 1px solid var(--color-border);
    }

    /* Cards are clean, bordered, minimal shadow */
    .card {
        border: 1px solid var(--color-border);
        border-radius: var(--border-radius);
        background: var(--color-card-bg);
        transition: transform 0.2s ease, box-shadow 0.2s ease;
    }

    .card:hover {
        transform: translateY(-2px);
        box-shadow: var(--shadow-lg);
        border-color: var(--color-secondary);
    }

    /* Buttons are solid, geometric */
    .button {
        font-weight: 500;
        letter-spacing: -0.01em;
        border-radius: var(--border-radius);
    }
    '
),
(
    'soft-editorial',
    'Soft Editorial',
    'A gentle, smart aesthetic with warmer tones and serif typography. Uses whitespace to create a premium, thoughtful feel.',
    'elegant',
    '{subtle,gentle,premium,blog,publishing,agency}',
    ':root {
        /* Palette: Organic & Calm */
        --color-primary: #4338ca;        /* Indigo 700 - muted */
        --color-primary-hover: #3730a3;
        --color-primary-text: #ffffff;

        --color-secondary: #e0e7ff;      /* Indigo 100 */
        --color-secondary-hover: #c7d2fe;
        --color-secondary-text: #312e81;

        --color-accent: #f59e0b;         /* Amber - mainly for small highlights */

        --color-text: #292524;           /* Warm Grey/Stone 800 */
        --color-text-muted: #57534e;     /* Stone 600 */
        --color-heading: #1c1917;        /* Stone 900 */

        /* The background is not pure white, it is "paper" */
        --color-background: #fafaf9;     /* Stone 50 */
        --color-background-alt: #f5f5f4; /* Stone 100 */

        --color-border: #e7e5e4;         /* Stone 200 */

        /* Specialized Areas */
        --color-header-bg: #fafaf9;
        --color-header-text: #1c1917;

        --color-hero-title: #1c1917;
        --color-hero-subtitle: #44403c;

        --color-card-bg: #ffffff;

        --color-cta-bg: #4338ca;
        --color-cta-text: #ffffff;

        --color-footer-bg: #e7e5e4;
        --color-footer-text: #44403c;

        /* Design Tokens */
        --border-radius: 12px; /* Softer, friendlier corners */
        --shadow: 0 4px 6px -1px rgb(0 0 0 / 0.05), 0 2px 4px -2px rgb(0 0 0 / 0.05); /* Very diffused */
        --shadow-lg: 0 20px 25px -5px rgb(0 0 0 / 0.05), 0 8px 10px -6px rgb(0 0 0 / 0.01);

        --font-display: "Merriweather", "Georgia", serif;
        --font-body: "Lato", system-ui, sans-serif;
    }

    body {
        font-family: var(--font-body);
        background-color: var(--color-background);
        color: var(--color-text);
        line-height: 1.7; /* Relaxed reading experience */
    }

    h1, h2, h3, h4, .hero-title {
        font-family: var(--font-display);
        font-weight: 700;
        letter-spacing: -0.02em;
    }

    /* Header is minimal, no border, just floats */
    header {
        background: transparent;
        padding-top: 1rem;
        padding-bottom: 1rem;
    }

    /* Cards are soft, elevating gently */
    .card {
        border: 1px solid rgba(0,0,0,0.03); /* Almost invisible border */
        border-radius: var(--border-radius);
        background: var(--color-card-bg);
        box-shadow: var(--shadow);
    }

    .hero-section {
        /* A gentle fade overlay instead of a block color */
        background: linear-gradient(to bottom, transparent 0%, rgba(67, 56, 202, 0.03) 100%);
    }

    .button {
        border-radius: 50px; /* Pill shapes are friendlier */
        padding-left: 2rem;
        padding-right: 2rem;
        font-family: var(--font-body);
    }
    '
);


-- Insert a new agent definition for multipage-wrapper
-- This agent simply executes the wrap_multipage action

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
             'multipage-wrapper',
             'Multi-Page Site Wrapper',
             'Wraps single-page site into multi-page structure (index, about, contact)',
             'data-driven',
             '{
               "processing_mode": "task",
               "timeout_seconds": 30,
               "workflow": {
                 "start_step": "wrap_multipage",
                 "steps": {
                   "wrap_multipage": {
                     "action": "wrap_multipage",
                     "config": {
                       "index_html_field": "input_data.final_html.assemble_html.final_html"
                     },
                     "next_step": "complete",
                     "description": "Create about and contact pages"
                   },
                   "complete": {
                     "action": "complete_workflow",
                     "description": "Return files map"
                   }
                 }
               }
             }'::jsonb,
             true,
             '["data-transformation", "html", "multipage"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.484',
             '{
               "requests": {"cpu": "100m", "memory": "256Mi"},
               "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
               "process": "system.agent.{type}.process",
               "response": "system.responses.{type}",
               "error": "system.errors.{type}"
             }'::jsonb,
             '{
               "port": 8080,
               "liveness_path": "/health",
               "readiness_path": "/ready",
               "initial_delay_seconds": 15
             }'::jsonb,
             '[]'::jsonb,
             1,
             '{
               "prefer_delegation": true,
               "fallback_to_self": true
             }'::jsonb
         );