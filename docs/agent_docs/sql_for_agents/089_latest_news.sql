-- 026c — Latest News Component + Discovery Checks
--
-- Creates the latest-news content component and supporting SQL.
-- Run against clients_db after the content_sources table exists.

-- ---------------------------------------------------------------------------
-- 1. latest-news component template
-- ---------------------------------------------------------------------------
-- Follows the same pattern as blog-listing: data-driven, rendered by a
-- dedicated Go action (render_news_section), not by the LLM content writer.
-- The template uses CSS classes — styling comes from the site's CSS theme.

INSERT INTO content_components (
    name, display_name, description, function, category,
    component_level, render_mode, semantic_tags,
    html_template, input_schema, is_active
) VALUES (
             'Latest News Feed',
             'Latest News',
             'Displays recent news items relevant to the site vertical. Links out to original sources. Data loaded from content_feed_items by render_news_section action.',
             'latest-news',
             'content',
             'section',
             'template',
             '["news", "feed", "dynamic", "freshness"]'::jsonb,
             '<!-- latest-news component -->
         <section data-component="latest-news" class="latest-news-section section-padding">
           <div class="container">
             <h2 class="section-heading">{{.headline}}</h2>
             {{if .subheadline}}<p class="section-subheadline">{{.subheadline}}</p>{{end}}
             <div class="news-grid">
               {{range .news_items}}
               <article class="news-card">
                 <div class="news-card-content">
                   <h3 class="news-card-title">
                     <a href="{{.source_url}}" target="_blank" rel="noopener noreferrer">{{.source_title}}</a>
                   </h3>
                   {{if .source_summary}}<p class="news-card-summary">{{.source_summary}}</p>{{end}}
                   <div class="news-card-meta">
                     {{if .source_name}}<span class="news-source">{{.source_name}}</span>{{end}}
                     {{if .published_display}}<time class="news-date">{{.published_display}}</time>{{end}}
                   </div>
                 </div>
               </article>
               {{end}}
             </div>
             {{if not .news_items}}
             <p class="news-empty">News updates coming soon.</p>
             {{end}}
           </div>
         </section>',
             '{
                 "fields": {
                     "headline": {
                         "type": "text",
                         "source": "llm",
                         "required": true,
                         "default": "Latest News",
                         "description": "Section heading — can be customised per site"
                     },
                     "subheadline": {
                         "type": "text",
                         "source": "llm",
                         "required": false,
                         "description": "Optional subheading"
                     },
                     "news_items": {
                         "type": "array",
                         "source": "query.content_feed_items",
                         "required": false,
                         "description": "Populated by render_news_section action at render time",
                         "on_missing": "use_fallback",
                         "fallback": []
                     }
                 },
                 "render_action": "render_news_section",
                 "refresh_interval": "6h",
                 "notes": "Data populated by render_news_section Go action, not by LLM. Template rendering only — no JS, static HTML."
             }'::jsonb,
             true
         ) ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    semantic_tags = EXCLUDED.semantic_tags,
    updated_at = NOW();

-- Verify
SELECT id, function, display_name, render_mode, component_level
FROM content_components
WHERE function = 'latest-news';
