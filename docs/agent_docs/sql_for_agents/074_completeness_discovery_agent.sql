-- ============================================================================
-- Update completeness-discovery-agent with new integrity checks
--
-- Adds four new checks to the existing empty_sections check:
--   - cross_site_contamination: other site's company name in rendered HTML
--   - unrendered_templates: raw {{.field}} syntax in stored HTML
--   - missing_style_collection: site has no style_collection_id
--   - deactivated_site_components: site_components → inactive content_components
--
-- All are algorithmic — no LLM budget.
--
-- Run AFTER deploying integrity_checks.go (registers the checks via init()).
-- ============================================================================

UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "ensure_site_record",
        "processing_mode": "task",
        "timeout_seconds": 120,
        "steps": {
            "ensure_site_record": {
                "action": "ensure_site_record",
                "config": { "input_fields": ["site_id", "domain"] },
                "next_step": "run_structural_checks",
                "description": "Load site record from domain or site_id",
                "output_field": "site_record"
            },
            "run_structural_checks": {
                "action": "run_discovery_checks",
                "config": {
                    "site_id": "site_record.site_id",
                    "check_domain": "structural",
                    "checks": [
                        "missing_style_collection",
                        "deactivated_site_components",
                        "unrendered_templates",
                        "cross_site_contamination"
                    ]
                },
                "next_step": "run_content_checks",
                "description": "Run structural integrity checks — style collection, component linkage, template rendering, contamination",
                "output_field": "structural_result"
            },
            "run_content_checks": {
                "action": "run_discovery_checks",
                "config": {
                    "site_id": "site_record.site_id",
                    "check_domain": "content",
                    "checks": [
                        "empty_sections"
                    ]
                },
                "next_step": "complete",
                "description": "Run content completeness checks — empty deployed sections",
                "output_field": "content_result"
            },
            "complete": {
                "action": "complete_workflow",
                "config": { "output_fields": ["structural_result", "content_result"] },
                "description": "Completeness discovery complete"
            }
        }
    }
}'::jsonb,
description = 'Scans sites for structural integrity and content completeness issues: cross-site contamination, unrendered templates, missing style collections, deactivated components, empty sections. All checks algorithmic — no LLM budget needed.',
updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Verify
SELECT type, description,
       default_config->'workflow'->'steps'->'run_structural_checks'->'config'->'checks' as structural_checks,
       default_config->'workflow'->'steps'->'run_content_checks'->'config'->'checks' as content_checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- several updates

-- ============================================================================
-- Wire integrity checks into the correct discovery agents
--
-- Per 009b improvement loop flow:
--   Step 3: design-discovery-agent  → structural linkage checks
--   Step 4: completeness-discovery-agent → content integrity checks
--
-- The improvement loop runs these in order:
--   quality-discovery (step 2) → design-discovery (step 3) → completeness-discovery (step 4)
-- so structural fixes (missing style, stale components) are found before
-- content checks (contamination, unrendered templates) run.
--
-- All checks produce status: 'detected'. The improvement loop's triage step (7)
-- promotes them to 'triaged' for dispatch.
--
-- Run AFTER deploying integrity_checks.go (registers checks via init()).
-- ============================================================================

-- ── 1. Update design-discovery-agent ──
-- Add: missing_style_collection, deactivated_site_components,
--       stale_site_components, shared_css_theme

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["undeployed_assets", "missing_css", "duplicate_palette", "hardcoded_section_colors", "forced_text_colors", "missing_tools", "missing_style_collection", "deactivated_site_components", "stale_site_components", "shared_css_theme"]'::jsonb
                     ),
    description = 'Scans sites for design-domain issues: undeployed assets, missing CSS, colour problems, missing style collections, deactivated components, stale header/footer renders, shared style collections. All algorithmic — no LLM budget.',
    updated_at = NOW()
WHERE type = 'design-discovery-agent' AND deleted_at IS NULL;

-- ── 2. Update completeness-discovery-agent ──
-- Add: cross_site_contamination, unrendered_templates
-- Keep: empty_sections (existing)

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["empty_sections", "cross_site_contamination", "unrendered_templates"]'::jsonb
                     ),
    description = 'Scans sites for content completeness and integrity: empty sections, cross-site company name contamination, unrendered Go template syntax. All algorithmic — no LLM budget.',
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- ── 3. Verify both agents ──

SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks,
       substring(description, 1, 100) as desc
FROM agent_definitions
WHERE type IN ('design-discovery-agent', 'completeness-discovery-agent')
  AND deleted_at IS NULL
ORDER BY type;

-- Expected:
-- completeness-discovery-agent | ["empty_sections", "cross_site_contamination", "unrendered_templates"]
-- design-discovery-agent       | ["undeployed_assets", "missing_css", ..., "shared_css_theme"]



-- ============================================================================
-- Add missing_structure check to completeness-discovery-agent
--
-- This check runs alongside empty_sections. It flags deployed pages
-- where rendered_header or rendered_footer is NULL.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["empty_sections", "missing_structure"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- 1. Add empty_blog and missing_structure to completeness-discovery-agent
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["empty_sections", "missing_structure", "empty_blog"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- 2. Create immediate work items for existing sites with empty blogs
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT
    p.site_id,
    'discovery', 'build', 'needs_blog_posts', 'medium',
    'Blog page exists but no posts — needs content planning for ' || s.domain,
    jsonb_build_object(
            'check', 'empty_blog',
            'blog_page_id', p.id,
            'blog_page_name', p.name,
            'post_count', 0
    ),
    50, 'blog-content-planner', 'triaged', 'admin',
    'empty_blog:' || s.id
FROM pages p
         JOIN sites s ON s.id = p.site_id
WHERE (p.name = 'blog' OR p.page_type = 'blog-index')
  AND p.status = 'active'
  AND NOT EXISTS (
    SELECT 1 FROM pages child
    WHERE child.site_id = p.site_id
      AND child.page_type = 'blog-post'
)
    ON CONFLICT DO NOTHING;

-- 3. Check what was created
SELECT s.domain, wi.item_type, wi.handler_agent, wi.status
FROM site_work_items wi
         JOIN sites s ON s.id = wi.site_id
WHERE wi.handler_agent = 'blog-content-planner';
---
-- add orphan pages to workflow

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["empty_sections", "missing_structure", "empty_blog", "orphan_pages"]'
                     )
WHERE type = 'completeness-discovery-agent';

---

-- 026d — Enable news feed discovery checks
--
-- Adds the four news feed checks to the completeness-discovery-agent's
-- checks config. Run after deploying the chassis with check_news_feed.go.

-- First, check current checks list
SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as current_checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

-- Add news feed checks to the existing list
-- This preserves existing checks and appends the new ones
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (
            SELECT jsonb_agg(DISTINCT val)
            FROM (
                     -- Existing checks
                     SELECT val
                     FROM jsonb_array_elements(
                                  default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
                          ) val
                     UNION
                     -- New news feed checks
                     SELECT val FROM jsonb_array_elements(
                                             '["missing_news_sources", "missing_news_section", "stale_news_section", "all_sources_erroring"]'::jsonb
                                     ) val
                 ) combined
        )
                     ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL
    RETURNING type,
    default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as updated_checks;


-- ============================================================================
-- 029_news_listing_page.sql
--
-- Registers the news-listing component, CSS, and related infrastructure
-- for the dedicated /news.html listing page (Tier 2 of news expansion).
-- ============================================================================

-- ============================================================================
-- 1. Register news-listing content_component
-- ============================================================================
-- This component renders on the /news.html page. It loads data/news-archive.json
-- via fetch() and displays a list of news items with source attribution,
-- dates, and topic tags.
--
-- The template follows the same client-side-fetch pattern as latest-news:
-- at build time, items are empty; the feed orchestrator populates the JSON.

INSERT INTO content_components (
    name, display_name, description, function, category,
    component_level, is_active, render_mode,
    section_type, content_shape, visual_density,
    suitable_site_types, suitable_page_types,
    html_template, input_schema
) VALUES (
             'news-listing',
             'News Listing Page',
             'Full news listing page that loads items from /data/news-archive.json. Shows 20 items with source, date, topic tags, and external links. Companion to the latest-news homepage snippet.',
             'news-listing',
             'content',
             'section',
             true,
             'template',
             'news-listing',
             'mixed',
             'high',
             '["corporate", "content", "tools"]'::jsonb,
             '["news-index"]'::jsonb,
             -- HTML template: loads JSON and renders a news list
             '<section class="news-listing-section" id="news-listing">
           <div class="news-listing-container">
             <div class="news-listing-header">
               <h1 class="news-listing-title">{{.headline}}</h1>
               {{if .subheadline}}<p class="news-listing-subtitle">{{.subheadline}}</p>{{end}}
             </div>
             <div class="news-listing-items" id="news-listing-items">
               <p class="news-listing-loading">Loading latest news...</p>
             </div>
             <div class="news-listing-footer" id="news-listing-footer" style="display:none;">
               <p class="news-listing-count" id="news-listing-count"></p>
               <p class="news-listing-updated" id="news-listing-updated"></p>
             </div>
           </div>
           <script>
           (function() {
             fetch("/data/news-archive.json")
               .then(function(r) { return r.json(); })
               .then(function(data) {
                 var container = document.getElementById("news-listing-items");
                 var footer = document.getElementById("news-listing-footer");
                 if (!data.items || data.items.length === 0) {
                   container.innerHTML = "<p class=\"news-listing-empty\">No news items available yet. Check back soon.</p>";
                   return;
                 }
                 var html = "";
                 data.items.forEach(function(item) {
                   html += "<article class=\"news-list-item\">";
                   html += "<div class=\"news-list-item-content\">";
                   html += "<h3 class=\"news-list-item-title\"><a href=\"" + item.url + "\" target=\"_blank\" rel=\"noopener noreferrer\">" + item.title + "</a></h3>";
                   if (item.summary) {
                     html += "<p class=\"news-list-item-summary\">" + item.summary + "</p>";
                   }
                   html += "<div class=\"news-list-item-meta\">";
                   if (item.source) {
                     html += "<span class=\"news-list-item-source\">" + item.source + "</span>";
                   }
                   if (item.date) {
                     html += "<span class=\"news-list-item-date\">" + item.date + "</span>";
                   }
                   html += "</div>";
                   if (item.topics) {
                     html += "<div class=\"news-list-item-topics\">";
                     item.topics.split(", ").forEach(function(tag) {
                       html += "<span class=\"news-list-tag\">" + tag + "</span>";
                     });
                     html += "</div>";
                   }
                   html += "</div>";
                   html += "</article>";
                 });
                 container.innerHTML = html;
                 footer.style.display = "block";
                 var countEl = document.getElementById("news-listing-count");
                 if (data.items_total && data.items_total > data.item_count) {
                   countEl.textContent = "Showing " + data.item_count + " of " + data.items_total + " items";
                 } else {
                   countEl.textContent = data.item_count + " items";
                 }
                 var updatedEl = document.getElementById("news-listing-updated");
                 if (data.updated_at) {
                   var d = new Date(data.updated_at);
                   updatedEl.textContent = "Last updated: " + d.toLocaleDateString() + " " + d.toLocaleTimeString();
                 }
               })
               .catch(function(err) {
                 var container = document.getElementById("news-listing-items");
                 container.innerHTML = "<p class=\"news-listing-empty\">Unable to load news. Please try again later.</p>";
               });
           })();
           </script>
         </section>',
             -- input_schema: what the content writer provides
             '{
               "type": "object",
               "properties": {
                 "headline": {
                   "type": "string",
                   "description": "Page headline for the news listing",
                   "source": "llm"
                 },
                 "subheadline": {
                   "type": "string",
                   "description": "Optional subtitle for context",
                   "source": "llm"
                 }
               },
               "required": ["headline"]
             }'::jsonb
         )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              suitable_page_types = EXCLUDED.suitable_page_types,
                              updated_at = NOW();

-- ============================================================================
-- 2. CSS snippet for news listing page
-- ============================================================================
-- Applied automatically by render_css_from_spec when a page has the
-- news-listing component (via applies_to && $1::jsonb query).

INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES (
                                                                                                'News Listing Page',
                                                                                                'Styles for the /news.html listing page — news-listing component. Full-width list layout with source attribution, topic tags, and metadata.',
                                                                                                '/* News Listing Page */
                                                                                            .news-listing-section {
                                                                                              padding: 3rem 1.5rem;
                                                                                              max-width: 900px;
                                                                                              margin: 0 auto;
                                                                                            }
                                                                                            .news-listing-header {
                                                                                              margin-bottom: 2rem;
                                                                                              border-bottom: 2px solid var(--color-border, #e5e7eb);
                                                                                              padding-bottom: 1rem;
                                                                                            }
                                                                                            .news-listing-title {
                                                                                              font-size: 2rem;
                                                                                              font-weight: 700;
                                                                                              margin: 0 0 0.5rem 0;
                                                                                              color: var(--color-heading, #1a1a2e);
                                                                                            }
                                                                                            .news-listing-subtitle {
                                                                                              font-size: 1.1rem;
                                                                                              color: var(--color-text-muted, #6b7280);
                                                                                              margin: 0;
                                                                                            }
                                                                                            .news-listing-items {
                                                                                              display: flex;
                                                                                              flex-direction: column;
                                                                                              gap: 0;
                                                                                            }
                                                                                            .news-list-item {
                                                                                              padding: 1.25rem 0;
                                                                                              border-bottom: 1px solid var(--color-border, #e5e7eb);
                                                                                            }
                                                                                            .news-list-item:last-child {
                                                                                              border-bottom: none;
                                                                                            }
                                                                                            .news-list-item-title {
                                                                                              font-size: 1.15rem;
                                                                                              font-weight: 600;
                                                                                              margin: 0 0 0.4rem 0;
                                                                                              line-height: 1.4;
                                                                                            }
                                                                                            .news-list-item-title a {
                                                                                              color: var(--color-heading, #1a1a2e);
                                                                                              text-decoration: none;
                                                                                              transition: color 0.15s ease;
                                                                                            }
                                                                                            .news-list-item-title a:hover {
                                                                                              color: var(--color-primary, #2563eb);
                                                                                            }
                                                                                            .news-list-item-summary {
                                                                                              font-size: 0.95rem;
                                                                                              color: var(--color-text, #374151);
                                                                                              margin: 0 0 0.5rem 0;
                                                                                              line-height: 1.6;
                                                                                            }
                                                                                            .news-list-item-meta {
                                                                                              display: flex;
                                                                                              gap: 1rem;
                                                                                              font-size: 0.8rem;
                                                                                              color: var(--color-text-muted, #9ca3af);
                                                                                              margin-bottom: 0.3rem;
                                                                                            }
                                                                                            .news-list-item-source {
                                                                                              font-weight: 500;
                                                                                            }
                                                                                            .news-list-item-topics {
                                                                                              display: flex;
                                                                                              flex-wrap: wrap;
                                                                                              gap: 0.4rem;
                                                                                              margin-top: 0.3rem;
                                                                                            }
                                                                                            .news-list-tag {
                                                                                              display: inline-block;
                                                                                              background: var(--color-surface, #f3f4f6);
                                                                                              color: var(--color-text-muted, #6b7280);
                                                                                              padding: 0.15rem 0.5rem;
                                                                                              border-radius: 3px;
                                                                                              font-size: 0.75rem;
                                                                                              font-weight: 500;
                                                                                            }
                                                                                            .news-listing-footer {
                                                                                              margin-top: 1.5rem;
                                                                                              padding-top: 1rem;
                                                                                              border-top: 1px solid var(--color-border, #e5e7eb);
                                                                                              display: flex;
                                                                                              justify-content: space-between;
                                                                                              align-items: center;
                                                                                            }
                                                                                            .news-listing-count,
                                                                                            .news-listing-updated {
                                                                                              font-size: 0.8rem;
                                                                                              color: var(--color-text-muted, #9ca3af);
                                                                                              margin: 0;
                                                                                            }
                                                                                            .news-listing-loading,
                                                                                            .news-listing-empty {
                                                                                              text-align: center;
                                                                                              padding: 2rem;
                                                                                              color: var(--color-text-muted, #9ca3af);
                                                                                              font-size: 0.95rem;
                                                                                            }
                                                                                            @media (max-width: 640px) {
                                                                                              .news-listing-section {
                                                                                                padding: 2rem 1rem;
                                                                                              }
                                                                                              .news-listing-title {
                                                                                                font-size: 1.5rem;
                                                                                              }
                                                                                              .news-listing-footer {
                                                                                                flex-direction: column;
                                                                                                gap: 0.5rem;
                                                                                              }
                                                                                            }',
                                                                                                '["news", "listing", "archive", "feed"]'::jsonb,
                                                                                                '["news-listing"]'::jsonb
                                                                                            )
    ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
                              applies_to = EXCLUDED.applies_to;

-- ============================================================================
-- 3. Update chief-strategist prompt (Rule for news page)
-- ============================================================================
-- This can't be done via SQL alone — it requires updating the prompt_template
-- in the chief-strategist agent definition. The prompt change is documented
-- here as text, to be applied as a str_replace to the agent_definitions row.
--
-- ADD to the chief-strategist prompt_template, in the "STEP 2" section
-- after the component types list:
--
-- Rule: News Page
-- If the site's classification spec has content_features.news_feed.separate_page = true,
-- include a "News" or "Insights" page with:
-- - name: "news"
-- - title: "News & Insights | {Brand}"
-- - page_type: "news-index"
-- - components: [{"type": "news-listing", "priority": "high"}]
-- - In the sitemap: {"label": "News", "page": "news", "url": "/news.html", "in_header": true, "in_footer": true}
--
-- This is applied programmatically below:

-- NOTE: The prompt template in the chief-strategist contains the planner rules.
-- Rather than modifying the enormous prompt via SQL (error-prone), we add a
-- separate post-planning step that checks the spec and creates the news page
-- if needed. This is handled by the content-gap-planner when it receives a
-- "missing_news_page" work item from discovery.
--
-- The planner already has "content-feed" as an available component type, but
-- doesn't know about "news-listing". The discovery → gap-planner → build path
-- is more maintainable than embedding news logic in the planner prompt.

-- ============================================================================
-- 4. Discovery check SQL — enable missing_news_page check
-- ============================================================================
-- The Go check is in check_news_feed.go (see separate file).
-- Register it in the completeness-discovery-agent's checks array:

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (
            SELECT jsonb_agg(DISTINCT val)
            FROM (
                     SELECT val FROM jsonb_array_elements_text(
                                             default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
                                     ) AS val
                     UNION
                     SELECT 'missing_news_page'
                 ) sub
        )
                     ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent'
  AND deleted_at IS NULL;

-- ============================================================================
-- 5. Verify
-- ============================================================================
-- After running:
--
-- SELECT name, function, component_level, suitable_page_types
-- FROM content_components WHERE function = 'news-listing';
--
-- SELECT name, applies_to FROM css_snippets WHERE name = 'News Listing Page';
--
-- SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
-- FROM agent_definitions WHERE type = 'completeness-discovery-agent';


-- Add unresolved_sections to completeness-discovery-agent
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["all_sources_erroring", "empty_blog", "empty_sections", "missing_news_section", "missing_news_sources", "missing_structure", "orphan_pages", "stale_news_section", "unresolved_sections"]'::jsonb
                     )
WHERE type = 'completeness-discovery-agent';

---

-- Add both missing checks back
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (
            SELECT jsonb_agg(DISTINCT val)
            FROM (
                     SELECT val FROM jsonb_array_elements_text(
                                             default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
                                     ) AS val
                     UNION
                     SELECT 'missing_news_page'
                     UNION
                     SELECT 'unlinked_page_components'
                 ) sub
        )
                     ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent'
  AND deleted_at IS NULL;

-- Verify
SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions
WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

