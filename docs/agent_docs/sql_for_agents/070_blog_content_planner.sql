-- ============================================================================
-- 1. Fix empty_sections handler_agent: page-content-writer → page-build-handler
--    This is a Go change in check_empty_sections.go:
--    HandlerAgent: "page-build-handler" (was "page-content-writer")
--
--    Note: can't fix via SQL since it's hardcoded in Go. Noting for next build.
-- ============================================================================


-- ============================================================================
-- 2. blog-content-planner agent
--
-- Receives: site_id, domain
-- Does:
--   1. Loads site specs (identity, content_direction)
--   2. Checks existing blog-post pages
--   3. Asks LLM to plan 3-4 posts based on industry/services/audience
--   4. Creates page records for each post
--   5. Creates needs_content_page work items for page-build-handler
--   6. Creates needs_rerender for blog index
--
-- No new Go code. Uses existing actions:
--   ensure_site_record, read_site_spec, query_database,
--   execute_llm_prompt, create_pages_from_plan (if exists),
--   create_work_item, complete_workflow
--
-- Since we don't have a "create multiple pages" action, we use
-- execute_llm_prompt to plan the posts, then a Go action to create
-- the page records and work items. We DO have write_build_items which
-- creates work items from a site plan. We can use that if we structure
-- the LLM output to match the expected format.
--
-- Simpler approach: the LLM plans posts, a single query_database step
-- creates the page records, and create_work_item steps create the items.
-- But that requires knowing the count in advance.
--
-- Simplest approach: the agent produces a plan, a NEW Go action
-- create_blog_posts_from_plan handles creation of pages + work items.
--
-- BUT: we want to avoid new Go code where possible. Let's use the
-- pattern of LLM produces JSON → write_build_items action (which
-- already creates page records and work items from a plan).
--
-- The blog planner's LLM output should match what write_build_items
-- expects: an array of pages with sections.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'blog-content-planner',
             'Blog Content Planner',
             'Plans initial blog posts for a site based on its industry, services, and target audience. Reads site specs, checks existing posts, asks LLM to plan 3-4 relevant posts, creates page records and work items for page-build-handler.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.837', 'specialist',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":300,"steps":{
                 "ensure_site_record":{
                     "action":"ensure_site_record",
                     "config":{"store_brief_in_content_data":false},
                     "next_step":"load_specs",
                     "output_field":"site_record"
                 },
                 "load_specs":{
                     "action":"read_site_spec",
                     "config":{"site_id":"site_record.site_id"},
                     "next_step":"check_existing_posts",
                     "output_field":"site_specs"
                 },
                 "check_existing_posts":{
                     "action":"query_database",
                     "config":{
                         "query":"SELECT name, title, slug, build_status FROM pages WHERE site_id = $1 AND page_type = ''blog-post'' ORDER BY sort_order",
                         "params":["site_record.site_id"],
                         "output_format":"array"
                     },
                     "next_step":"plan_posts",
                     "output_field":"existing_posts"
                 },
                 "plan_posts":{
                     "action":"execute_llm_prompt",
                     "config":{
                         "ai_service":{
                             "model":"claude-sonnet-4-6",
                             "provider":"anthropic",
                             "max_tokens":3000,
                             "api_key_env_var":"ANTHROPIC_API_KEY"
                         },
                         "input_fields":["site_record","site_specs","existing_posts"],
                         "output_format":"json",
                         "prompt_template":"You are planning initial blog posts for a website.\n\n## Site\nDomain: {{.site_record.domain}}\nCompany: {{if .site_specs.specs.identity.company_name}}{{.site_specs.specs.identity.company_name}}{{end}}\nIndustry: {{if .site_specs.specs.identity.industry}}{{.site_specs.specs.identity.industry}}{{end}}\nServices: {{if .site_specs.specs.identity.services}}{{.site_specs.specs.identity.services}}{{end}}\nTarget Audience: {{if .site_specs.specs.identity.target_audience}}{{.site_specs.specs.identity.target_audience}}{{end}}\n\n## Content Direction\n{{if .site_specs.specs.content_direction}}Voice: {{.site_specs.specs.content_direction.voice}}\nEmphasis: {{.site_specs.specs.content_direction.emphasis}}\nAvoid: {{.site_specs.specs.content_direction.avoid_phrases}}{{end}}\n\n## Existing Blog Posts (do not duplicate these)\n{{if .existing_posts}}{{range .existing_posts}}- {{.title}} ({{.name}}){{end}}{{else}}None yet — this is the initial set.{{end}}\n\nPlan 3-4 blog posts that would be genuinely useful to the target audience and demonstrate the company''s expertise. Each post should serve a different purpose:\n- One educational/guide post (helps the audience understand something)\n- One opinion/insight post (shows thought leadership)\n- One practical/how-to post (gives actionable advice)\n- Optionally: one industry news/trends post\n\nFor each post, choose a specific topic relevant to this company''s industry and services. Do NOT write generic AI/tech posts unless the company is in AI/tech.\n\nReturn ONLY valid JSON:\n{\n  \"posts\": [\n    {\n      \"name\": \"kebab-case-slug\",\n      \"title\": \"Full Post Title | Company Name\",\n      \"purpose\": \"What this post covers and why it helps the audience\",\n      \"page_type\": \"blog-post\",\n      \"sections\": [\"hero\", \"article-body\", \"call-to-action\"]\n    }\n  ]\n}"
                     },
                     "next_step":"create_post_pages",
                     "output_field":"blog_plan"
                 },
                 "create_post_pages":{
                     "action":"sync_pages_to_db",
                     "config":{
                         "input_fields":["site_record","blog_plan"],
                         "pages_field":"blog_plan.result.posts",
                         "page_type_default":"blog-post",
                         "slug_prefix":"/blog/",
                         "in_header":false,
                         "in_footer":false,
                         "start_sort_order":20
                     },
                     "next_step":"create_build_items",
                     "error_step":"create_build_items_fallback",
                     "output_field":"pages_created"
                 },
                 "create_build_items":{
                     "action":"write_build_items",
                     "config":{
                         "site_id":"site_record.site_id",
                         "site_plan":"blog_plan.result",
                         "handler_agent":"page-build-handler",
                         "base_priority":55,
                         "item_domain":"build"
                     },
                     "next_step":"create_rerender",
                     "error_step":"complete",
                     "output_field":"build_items"
                 },
                 "create_build_items_fallback":{
                     "action":"complete_workflow",
                     "config":{
                         "output_fields":["blog_plan","existing_posts"],
                         "success_message":"Blog posts planned but page creation needs manual sync"
                     }
                 },
                 "create_rerender":{
                     "action":"create_work_item",
                     "config":{
                         "site_id":"site_record.site_id",
                         "item_type":"needs_rerender",
                         "handler_agent":"rerender-pages",
                         "item_domain":"build",
                         "source":"blog-content-planner",
                         "priority":60,
                         "severity":"medium",
                         "summary":"Re-render blog index after new posts created",
                         "item_key_prefix":"rerender_blog"
                     },
                     "next_step":"complete",
                     "output_field":"rerender_item"
                 },
                 "complete":{
                     "action":"complete_workflow",
                     "config":{"output_fields":["blog_plan","pages_created","build_items","rerender_item"]}
                 }
             }}}'::jsonb,
             '{"required": ["site_id", "domain"]}'::jsonb,
             '{"produces": {"blog_plan": "LLM-planned posts", "build_items": "work items created for each post"}}'::jsonb,
             '["content", "blog", "planning"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();


-- ============================================================================
-- 3. Add detection for empty blog index
--
-- The content-quality-auditor's check_empty_pages step already finds pages
-- with no content. But the blog index HAS content (hero + article-grid + CTA)
-- — the article-grid is just empty because there are no posts.
--
-- Add a specific structural check: blog page exists but no blog-post pages.
-- This goes in the improvement loop as a discovery check.
--
-- For now, we can detect this via content-quality-auditor by adding the
-- blog-post count to the data it sends to the LLM. But the more reliable
-- approach is a Go discovery check.
--
-- Go change needed (next build): Add to check_component_standards.go:
--
--   type BlogEmptyCheck struct{}
--   func (c *BlogEmptyCheck) Name() string { return "empty_blog" }
--   func (c *BlogEmptyCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
--       // Check if a blog/blog-index page exists
--       var blogPageID string
--       err := dctx.DB.QueryRowContext(dctx.Ctx, `
--           SELECT id::text FROM pages
--           WHERE site_id = $1 AND (page_type = 'blog-index' OR name = 'blog')
--             AND build_status IN ('deployed', 'planned')
--       `, dctx.SiteID).Scan(&blogPageID)
--       if err != nil { return &CheckResult{}, nil } // no blog page, nothing to check
--
--       // Count blog-post pages
--       var postCount int
--       dctx.DB.QueryRowContext(dctx.Ctx, `
--           SELECT COUNT(*) FROM pages
--           WHERE site_id = $1 AND page_type = 'blog-post'
--             AND build_status IN ('deployed', 'planned')
--       `, dctx.SiteID).Scan(&postCount)
--
--       if postCount > 0 { return &CheckResult{}, nil } // has posts, fine
--
--       // Blog page exists but no posts
--       pageID, _ := uuid.Parse(blogPageID)
--       return &CheckResult{
--           Findings: []map[string]interface{}{{
--               "check": "empty_blog",
--               "blog_page_id": blogPageID,
--               "post_count": 0,
--           }},
--           WorkItems: []WorkItemSpec{{
--               SiteID:       dctx.SiteID,
--               Source:       "discovery",
--               Domain:       "content",
--               ItemType:     "needs_blog_posts",
--               Severity:     "medium",
--               Summary:      "Blog page exists but no blog posts — needs initial content",
--               PageID:       &pageID,
--               Priority:     50,
--               HandlerAgent: "blog-content-planner",
--               Status:       "detected",
--               CreatedBy:    dctx.AgentType,
--               ItemKey:      fmt.Sprintf("empty_blog:%s", dctx.SiteID),
--               BatchID:      dctx.BatchID,
--           }},
--       }, nil
--   }
--
-- Register in init(): Register(&BlogEmptyCheck{})
-- ============================================================================


-- ============================================================================
-- 4. Also fix the empty_sections handler_agent routing
--    Go change (next build): in check_empty_sections.go
--    Change: HandlerAgent: "page-content-writer"
--    To:     HandlerAgent: "page-build-handler"
-- ============================================================================


-- ============================================================================
-- Verify
-- ============================================================================
SELECT type, display_name, status
FROM agent_definitions
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

--

-- Fix check_existing_posts query: use url not slug
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_existing_posts,config,query}',
        '"SELECT name, title, url, build_status FROM pages WHERE site_id = $1 AND page_type = ''blog-post'' ORDER BY nav_order"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Verify
SELECT default_config->'workflow'->'steps'->'check_existing_posts'->'config'->>'query'
FROM agent_definitions WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

--

-- Fix 1: Change plan_posts output_field to site_plan
-- so sync_pages_to_db finds the pages where it expects them
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_posts,output_field}',
        '"site_plan"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Fix 2: Update the LLM prompt to return "pages" not "posts"
-- (sync_pages_to_db looks for .pages array)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_posts,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'plan_posts'->'config'->>'prompt_template',
                        '"posts": [',
                        '"pages": ['
                )
        )
                     ),
    updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Fix 3: Remove unsupported pages_field from sync_pages_to_db config
-- Use simpler config that the action already supports
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_post_pages,config}',
        '{"input_fields": ["site_record", "site_plan"]}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Fix 4: Also update write_build_items to use site_plan (already correct path)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_build_items,config}',
        '{"site_id": "site_record.site_id", "site_plan": "site_plan"}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Verify the chain
SELECT
    default_config->'workflow'->'steps'->'plan_posts'->>'output_field' as plan_output,
    default_config->'workflow'->'steps'->'plan_posts'->'config'->>'prompt_template' LIKE '%"pages"%' as uses_pages_key,
    default_config->'workflow'->'steps'->'create_post_pages'->'config' as sync_config
FROM agent_definitions
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;


----

-- Simplify blog-content-planner: replace sync_pages_to_db + write_build_items
-- with single create_blog_posts action

-- Replace create_post_pages step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_post_pages}',
        '{
            "action": "create_blog_posts",
            "config": {
                "site_id": "site_record.site_id",
                "plan_field": "site_plan.result"
            },
            "next_step": "complete",
            "error_step": "complete",
            "description": "Create page records and work items for each planned blog post",
            "output_field": "posts_created"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Remove the steps we no longer need (create_build_items, create_build_items_fallback, create_rerender)
-- by pointing plan_posts directly to create_post_pages
-- (already correct — plan_posts.next_step is create_post_pages)

-- Remove unused steps from the workflow
UPDATE agent_definitions
SET default_config = default_config
    #- '{workflow,steps,create_build_items}'
    #- '{workflow,steps,create_build_items_fallback}'
    #- '{workflow,steps,create_rerender}',
updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Update complete to include posts_created
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config,output_fields}',
        '["site_plan", "existing_posts", "posts_created"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

-- Verify simplified chain: ensure → load_specs → check_existing → plan_posts → create_post_pages → complete
SELECT
    default_config->'workflow'->'steps'->'plan_posts'->>'next_step' as after_plan,
    default_config->'workflow'->'steps'->'create_post_pages'->>'next_step' as after_create,
    default_config->'workflow'->'steps'->'create_post_pages'->'config' as create_config
FROM agent_definitions
WHERE type = 'blog-content-planner' AND deleted_at IS NULL;

