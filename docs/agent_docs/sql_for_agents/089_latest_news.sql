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

---

-- updates

-- 026c — Latest News Component (JS-based)
--
-- The latest-news component fetches /data/latest-news.json client-side.
-- News updates are decoupled from page rerenders — only the JSON file changes.
-- The component HTML is set once at build time and remains stable.

INSERT INTO content_components (
    name, display_name, description, function, category,
    component_level, render_mode, semantic_tags,
    html_template, input_schema, is_active
) VALUES (
             'Latest News Feed',
             'Latest News',
             'Client-side rendered news section. Fetches /data/latest-news.json and renders cards. Content writer provides the headline; items are populated from the feed pipeline.',
             'latest-news',
             'content',
             'section',
             'template',
             '["news", "feed", "dynamic", "freshness"]'::jsonb,
             '<!-- latest-news component -->
         <section data-component="latest-news" class="latest-news-section section-padding">
           <div class="container">
             <h2 class="section-heading" id="news-headline">{{.headline}}</h2>
             {{if .subheadline}}<p class="section-subheadline" id="news-subheadline">{{.subheadline}}</p>{{end}}
             <div class="news-grid" id="news-container">
               <noscript>
                 <p class="news-empty">Enable JavaScript to see the latest news{{if .insights_url}}, or visit
                 <a href="{{.insights_url}}">our insights page</a>{{end}}.</p>
               </noscript>
             </div>
             <div id="news-footer"></div>
           </div>
         </section>
         <script>
         (function() {
           fetch("/data/latest-news.json")
             .then(function(r) { return r.json(); })
             .then(function(data) {
               if (data.headline) {
                 document.getElementById("news-headline").textContent = data.headline;
               }
               var sub = document.getElementById("news-subheadline");
               if (sub && data.subheadline) {
                 sub.textContent = data.subheadline;
               }
               var container = document.getElementById("news-container");
               if (data.items && data.items.length > 0) {
                 container.innerHTML = data.items.map(function(item) {
                   var html = "<article class=\"news-card\"><div class=\"news-card-content\">";
                   html += "<h3 class=\"news-card-title\"><a href=\"" + item.url + "\" target=\"_blank\" rel=\"noopener noreferrer\">" + item.title + "</a></h3>";
                   if (item.summary) {
                     html += "<p class=\"news-card-summary\">" + item.summary + "</p>";
                   }
                   html += "<div class=\"news-card-meta\">";
                   if (item.source) {
                     html += "<span class=\"news-source\">" + item.source + "</span>";
                   }
                   if (item.date) {
                     html += "<time class=\"news-date\">" + item.date + "</time>";
                   }
                   html += "</div></div></article>";
                   return html;
                 }).join("");
               }
               if (data.insights_url) {
                 document.getElementById("news-footer").innerHTML =
                   "<div class=\"news-section-footer\"><a href=\"" + data.insights_url +
                   "\" class=\"news-more-link\">" + (data.insights_label || "More insights &rarr;") +
                   "</a></div>";
               }
             })
             .catch(function() {});
         })();
         </script>',
             '{
                 "fields": {
                     "headline": {
                         "type": "text",
                         "source": "llm",
                         "required": true,
                         "default": "Latest News",
                         "description": "Section heading — generated by content writer, tailored to site identity"
                     },
                     "subheadline": {
                         "type": "text",
                         "source": "llm",
                         "required": false,
                         "description": "Optional subheading — can be updated by triage rewriter with contextual text"
                     },
                     "insights_url": {
                         "type": "text",
                         "source": "query.pages",
                         "required": false,
                         "description": "URL to insights listing page (auto-populated when insights section exists)"
                     }
                 },
                 "render_action": "render_news_section",
                 "data_source": "/data/latest-news.json",
                 "refresh_interval": "6h",
                 "notes": "Items loaded client-side from JSON. Headline from content writer stored in content_data. JSON file updated by render_news_section + git_commit. No page rerender needed for news updates."
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
