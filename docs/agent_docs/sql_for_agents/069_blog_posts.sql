improvement loop runs empty_blog check
→ finds: blog page exists, 0 blog-post pages
→ creates: needs_blog_posts work item → blog-content-planner

dispatch loop picks it up
→ blog-content-planner:
      loads spec (identity, content_direction)
      checks existing posts (finds none)
      LLM plans 3-4 posts based on industry/audience
      sync_pages_to_db creates page records
      write_build_items creates needs_content_page items → page-build-handler
      creates needs_rerender for blog index

dispatch loop processes the post items
→ page-build-handler → page-content-writer (now reads specs)
      generates content in the spec''s voice
      saves to page_components, deploys to git

dispatch loop processes the rerender
  → blog index re-renders, article-grid finds the new posts



 -- ============================================================================
-- 1. Add read_site_spec step to page-content-writer
-- 2. Update the LLM prompt to reference spec data
-- 3. Block the blog article-grid issue
-- ============================================================================


-- ============================================================================
-- 1. Add load_site_specs step to page-content-writer
--    Insert between spawn_research_agent and prepare_link_context
-- ============================================================================

-- Change spawn_research_agent.next_step to load_site_specs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_research_agent,next_step}',
        '"load_site_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Add the load_site_specs step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_site_specs}',
        '{
            "action": "read_site_spec",
            "config": {
                "site_id": "input_data.site_record.site_id"
            },
            "next_step": "prepare_link_context",
            "error_step": "prepare_link_context",
            "description": "Load all site specs for content direction, identity, and design context",
            "output_field": "site_specs"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;


-- ============================================================================
-- 2. Update the generate_content prompt to reference site_specs
--    Add site_specs to input_fields and add spec sections to prompt
-- ============================================================================

-- Add site_specs to the input_fields array
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}',
        '["current_section", "render_context", "reviewed_brief", "current_page", "link_context", "site_plan", "site_specs"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Update the prompt template to include spec-driven content direction
-- We need to add the spec sections after the existing "Company Context" block
-- The prompt is very long so we use string replacement on a known anchor point
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
                    -- Find the section requirements header and insert spec context before it
                        '## Section Requirements',
                        '## Content Direction (from site spec — follow this closely)' || E'\n' ||
            '{{if .site_specs.specs.content_direction}}' || E'\n' ||
            'Voice: {{.site_specs.specs.content_direction.voice}}' || E'\n' ||
            'Emphasis: {{.site_specs.specs.content_direction.emphasis}}' || E'\n' ||
            'Avoid these phrases: {{.site_specs.specs.content_direction.avoid_phrases}}' || E'\n' ||
            'Social proof approach: {{.site_specs.specs.content_direction.social_proof_style}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            '{{if .site_specs.specs.identity.target_audience}}' || E'\n' ||
            'Target Audience: {{.site_specs.specs.identity.target_audience}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            '{{if .site_specs.specs.identity.key_differentiators}}' || E'\n' ||
            'Key Differentiators: {{.site_specs.specs.identity.key_differentiators}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            '{{if .site_specs.specs.design_intent.imagery_direction}}' || E'\n' ||
            'Imagery Direction: {{.site_specs.specs.design_intent.imagery_direction}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            E'\n' || '## Section Requirements'
                )
        )
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;


-- ============================================================================
-- 3. Create initial blog posts for finetuning.uk
--
-- The blog index page's article-grid lists blog-post pages.
-- No posts exist yet. Create 3-4 initial posts as work items.
-- page-build-handler will generate content using the spec's voice/direction.
-- After posts are deployed, re-render the blog index to pick them up.
-- ============================================================================

-- First, create the page records for the blog posts
INSERT INTO pages (site_id, name, slug, title, purpose, page_type, build_status, sort_order, in_header, in_footer)
VALUES
    ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'ai-for-uk-smes', '/blog/ai-for-uk-smes.html',
     'AI for UK SMEs: Where to Actually Start | FineTuning',
     'Practical guide for UK SMEs on where to start with AI — no hype, just actionable steps',
     'blog-post', 'planned', 20, false, false),
    ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'ai-hype-vs-reality', '/blog/ai-hype-vs-reality.html',
     'AI Hype vs Reality: What Actually Works in 2026 | FineTuning',
     'Honest assessment of which AI applications deliver real value and which are still overpromised',
     'blog-post', 'planned', 21, false, false),
    ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'custom-models-explained', '/blog/custom-models-explained.html',
     'Custom AI Models Explained: When Off-the-Shelf Isn''t Enough | FineTuning',
     'When and why a business needs a custom fine-tuned model vs using GPT/Claude directly',
     'blog-post', 'planned', 22, false, false),
    ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'automation-quick-wins', '/blog/automation-quick-wins.html',
     'Five AI Automation Quick Wins for Any Business | FineTuning',
     'Concrete examples of AI automations that pay for themselves within weeks',
     'blog-post', 'planned', 23, false, false)
    ON CONFLICT (site_id, name) DO NOTHING;

-- Create work items for each blog post
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    page_id, priority, handler_agent, status, created_by, spec
)
SELECT
    '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'human', 'build', 'needs_content_page', 'medium',
    'Write blog post: ' || p.title,
    p.id, 55, 'page-build-handler', 'triaged', 'human',
    jsonb_build_object(
            'page_name', p.name,
            'page_type', 'blog-post',
            'title', p.title,
            'purpose', p.purpose,
            'sections', jsonb_build_array('hero', 'article-body', 'call-to-action')
    )
FROM pages p
WHERE p.site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND p.page_type = 'blog-post'
  AND p.build_status = 'planned'
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.page_id = p.id AND wi.item_type = 'needs_content_page'
      AND wi.status IN ('triaged', 'claimed', 'detected')
);

-- Create a rerender item for the blog index (runs after posts are built)
-- Priority 60 so it runs after the posts (priority 55)
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    page_id, priority, handler_agent, status, created_by, spec
) VALUES (
             '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
             'human', 'build', 'needs_rerender', 'medium',
             'Re-render blog index after blog posts are created',
             (SELECT id FROM pages WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND name = 'blog'),
             60, 'rerender-pages', 'triaged', 'human',
             '{"refresh_site_components": false, "reason": "blog posts created, article-grid needs to pick them up"}'::jsonb
         );


-- ============================================================================
-- Verify
-- ============================================================================

-- Check the workflow chain
SELECT
    default_config->'workflow'->'steps'->'spawn_research_agent'->>'next_step' as after_spawn,
    default_config->'workflow'->'steps'->'load_site_specs'->>'next_step' as after_specs,
    default_config->'workflow'->'steps'->'load_site_specs'->>'error_step' as specs_error
FROM agent_definitions
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Check the prompt has the new section
SELECT
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' LIKE '%Content Direction%' as has_content_direction,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' LIKE '%site_specs%' as has_site_specs_ref
FROM agent_definitions
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Check input_fields includes site_specs
SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->'input_fields' as input_fields
FROM agent_definitions
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Check blog is blocked
SELECT p.name, p.build_status
FROM pages p
WHERE p.site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND p.name = 'blog';


----


-- ============================================================================
-- 1. Add read_site_spec step to page-content-writer
-- 2. Update the LLM prompt to reference spec data
-- 3. Block the blog article-grid issue
-- ============================================================================


-- ============================================================================
-- 1. Add load_site_specs step to page-content-writer
--    Insert between spawn_research_agent and prepare_link_context
-- ============================================================================

-- Change spawn_research_agent.next_step to load_site_specs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_research_agent,next_step}',
        '"load_site_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Add the load_site_specs step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_site_specs}',
        '{
            "action": "read_site_spec",
            "config": {
                "site_id": "input_data.site_record.site_id"
            },
            "next_step": "prepare_link_context",
            "error_step": "prepare_link_context",
            "description": "Load all site specs for content direction, identity, and design context",
            "output_field": "site_specs"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;


-- ============================================================================
-- 2. Update the generate_content prompt to reference site_specs
--    Add site_specs to input_fields and add spec sections to prompt
-- ============================================================================

-- Add site_specs to the input_fields array
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}',
        '["current_section", "render_context", "reviewed_brief", "current_page", "link_context", "site_plan", "site_specs"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Update the prompt template to include spec-driven content direction
-- We need to add the spec sections after the existing "Company Context" block
-- The prompt is very long so we use string replacement on a known anchor point
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
                    -- Find the section requirements header and insert spec context before it
                        '## Section Requirements',
                        '## Content Direction (from site spec — follow this closely)' || E'\n' ||
            '{{if .site_specs.specs.content_direction}}' || E'\n' ||
            'Voice: {{.site_specs.specs.content_direction.voice}}' || E'\n' ||
            'Emphasis: {{.site_specs.specs.content_direction.emphasis}}' || E'\n' ||
            'Avoid these phrases: {{.site_specs.specs.content_direction.avoid_phrases}}' || E'\n' ||
            'Social proof approach: {{.site_specs.specs.content_direction.social_proof_style}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            '{{if .site_specs.specs.identity.target_audience}}' || E'\n' ||
            'Target Audience: {{.site_specs.specs.identity.target_audience}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            '{{if .site_specs.specs.identity.key_differentiators}}' || E'\n' ||
            'Key Differentiators: {{.site_specs.specs.identity.key_differentiators}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            '{{if .site_specs.specs.design_intent.imagery_direction}}' || E'\n' ||
            'Imagery Direction: {{.site_specs.specs.design_intent.imagery_direction}}' || E'\n' ||
            '{{end}}' || E'\n' ||
            E'\n' || '## Section Requirements'
                )
        )
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;


-- ============================================================================
-- 3. Blog content is handled by blog-content-planner agent
--    (see blog_content_planner.sql)
--    Detection: empty_blog discovery check → creates needs_blog_posts item
--    Fix: blog-content-planner reads spec, plans posts, creates work items
-- ============================================================================


-- ============================================================================
-- Verify
-- ============================================================================

-- Check the workflow chain
SELECT
    default_config->'workflow'->'steps'->'spawn_research_agent'->>'next_step' as after_spawn,
    default_config->'workflow'->'steps'->'load_site_specs'->>'next_step' as after_specs,
    default_config->'workflow'->'steps'->'load_site_specs'->>'error_step' as specs_error
FROM agent_definitions
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Check the prompt has the new section
SELECT
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' LIKE '%Content Direction%' as has_content_direction,
    default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' LIKE '%site_specs%' as has_site_specs_ref
FROM agent_definitions
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Check input_fields includes site_specs
SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->'input_fields' as input_fields
FROM agent_definitions
WHERE type = 'page-content-writer' AND deleted_at IS NULL;

-- Check blog-content-planner exists
SELECT type, status FROM agent_definitions
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;