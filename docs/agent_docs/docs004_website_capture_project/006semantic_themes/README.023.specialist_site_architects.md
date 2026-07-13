-- ============================================================================
-- SPECIALIST ARCHITECT SYSTEM
-- ============================================================================

-- ============================================================================
-- 1. UPDATE BRIEFING AGENT - Now detects site_type
-- ============================================================================

UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "analyze_and_brief",
"steps": {
"analyze_and_brief": {
"action": "execute_llm_prompt",
"config": {
"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY",
"max_tokens": 3000
},
"input_fields": ["input_data"],
"output_field": "structured_brief",
"prompt_template": "Analyze this website request and create a comprehensive brief.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Model (if specified): {{.input_data.model}}\n\nFirst, determine the SITE TYPE based on the domain and objective:\n\n**landing** - Conversion-focused sites:\n- Product/service sales pages\n- SaaS landing pages\n- Lead generation\n- App download pages\n- Event registration\n- Single clear CTA goal\n\n**content** - Content/publishing sites:\n- News, blogs, magazines\n- Content aggregation\n- Ad-revenue models\n- SEO/traffic focused\n- Multiple articles/posts\n- Category navigation\n\n**portfolio** - Showcase sites:\n- Creative portfolios\n- Agency showcases\n- Case study focused\n- Gallery/visual heavy\n- Client testimonials\n\n**directory** - Listing sites:\n- Business directories\n- Job boards\n- Marketplace listings\n- Search/filter focused\n\nAnalyze the domain name and objective to determine site_type, then create a full brief.\n\nReturn JSON:\n{\n  \"site_type\": \"landing|content|portfolio|directory\",\n  \"site_type_confidence\": 0.0-1.0,\n  \"site_type_reasoning\": \"why this classification\",\n  \"analysis\": {\n    \"industry\": \"detected industry/niche\",\n    \"industry_confidence\": 0.0-1.0,\n    \"domain_interpretation\": \"what the domain name suggests\"\n  },\n  \"audience\": {\n    \"primary\": \"primary target audience\",\n    \"secondary\": \"secondary audience if applicable\",\n    \"demographics\": [\"age range\", \"profession\"],\n    \"psychographics\": [\"values\", \"motivations\", \"pain points\"]\n  },\n  \"brand\": {\n    \"tone\": \"professional|casual|technical|friendly|authoritative|playful\",\n    \"personality\": [\"trait1\", \"trait2\", \"trait3\"],\n    \"voice_examples\": [\"example phrase in brand voice\"]\n  },\n  \"messaging\": {\n    \"value_proposition\": \"core value proposition\",\n    \"key_messages\": [\"message1\", \"message2\", \"message3\"],\n    \"usps\": [\"unique selling point 1\", \"usp 2\"],\n    \"proof_points\": [\"credibility element 1\", \"element 2\"]\n  },\n  \"structure\": {\n    \"recommended_sections\": [\"section1\", \"section2\"],\n    \"priority_sections\": [\"most important sections\"],\n    \"optional_sections\": [\"sections that could be added\"]\n  },\n  \"theme\": {\n    \"recommended\": \"theme name\",\n    \"semantic_tags\": [\"tag1\", \"tag2\", \"tag3\"],\n    \"color_mood\": \"color feeling description\",\n    \"style_notes\": \"specific style recommendations\"\n  },\n  \"content_guidelines\": {\n    \"headline_style\": \"guidance for headlines\",\n    \"cta_style\": \"guidance for calls to action\",\n    \"avoid\": [\"things to avoid\"],\n    \"emphasize\": [\"things to emphasize\"]\n  },\n  \"monetization\": {\n    \"model\": \"subscription|ads|sales|leads|freemium\",\n    \"ad_zones\": [\"if ads: recommended ad placements\"]\n  }\n}\n\nReturn ONLY valid JSON."
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return the structured brief with site_type"
}
}
},
"processing_mode": "task",
"timeout_seconds": 60
}'::jsonb
WHERE type = 'briefing-agent';

-- ============================================================================
-- 2. RENAME CURRENT ARCHITECT → LANDING-PAGE-ARCHITECT
-- ============================================================================

-- First, create the new landing-page-architect as a copy
INSERT INTO agent_definitions (
id, type, display_name, description, category,
default_config, is_active, capabilities,
image_repository, image_tag, resources, topics, health_config
)
SELECT
gen_random_uuid(),
'landing-page-architect',
'Landing Page Architect',
'Assembles conversion-focused landing pages from component library (PAS, AIDA, etc.)',
category,
default_config,
true,
ARRAY['build', 'assemble', 'database', 'landing-page', 'conversion'],
image_repository, image_tag, resources, topics, health_config
FROM agent_definitions
WHERE type = 'site-component-architect'
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
description = EXCLUDED.description,
capabilities = EXCLUDED.capabilities,
updated_at = now();

-- ============================================================================
-- 3. CREATE CONTENT-SITE-ARCHITECT
-- ============================================================================

INSERT INTO agent_definitions (
id, type, display_name, description, category,
default_config, is_active, capabilities,
image_repository, image_tag, resources, topics, health_config
)
VALUES (
gen_random_uuid(),
'content-site-architect',
'Content Site Architect',
'Assembles content/publishing sites with article grids, category nav, and ad zones',
'data-driven',
'{
"workflow": {
"start_step": "assemble_content_site",
"steps": {
"assemble_content_site": {
"action": "assemble_from_library",
"config": {
"site_type": "content",
"input_fields": ["build_plan_data", "brief_data"],
"default_sections": ["header", "featured_article", "article_grid", "sidebar", "category_nav", "footer"],
"component_category": "content-site"
},
"next_step": "complete",
"description": "Assemble content site template from component library"
},
"complete": {
"action": "complete_workflow",
"description": "Return the content site template"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180
}'::jsonb,
true,
ARRAY['build', 'assemble', 'database', 'content-site', 'publishing'],
'docker.io/aqls/agent-chassis',
'v1.0.476',
'{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
'{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
'{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
)
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
updated_at = now();

-- ============================================================================
-- 4. CREATE PORTFOLIO-ARCHITECT (for future use)
-- ============================================================================

INSERT INTO agent_definitions (
id, type, display_name, description, category,
default_config, is_active, capabilities,
image_repository, image_tag, resources, topics, health_config
)
VALUES (
gen_random_uuid(),
'portfolio-architect',
'Portfolio Site Architect',
'Assembles portfolio/showcase sites with galleries, case studies, and visual layouts',
'data-driven',
'{
"workflow": {
"start_step": "assemble_portfolio_site",
"steps": {
"assemble_portfolio_site": {
"action": "assemble_from_library",
"config": {
"site_type": "portfolio",
"input_fields": ["build_plan_data", "brief_data"],
"default_sections": ["header", "hero_visual", "work_grid", "case_study", "client_logos", "about", "contact", "footer"],
"component_category": "portfolio-site"
},
"next_step": "complete",
"description": "Assemble portfolio site template from component library"
},
"complete": {
"action": "complete_workflow",
"description": "Return the portfolio site template"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180
}'::jsonb,
true,
ARRAY['build', 'assemble', 'database', 'portfolio-site', 'showcase'],
'docker.io/aqls/agent-chassis',
'v1.0.476',
'{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
'{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
'{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
)
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
updated_at = now();

-- ============================================================================
-- 5. CONTENT SITE COMPONENTS
-- ============================================================================

-- Add category for content site components
INSERT INTO content_components (name, display_name, function, category, semantic_tags, html_template, content_requirements, sort_order)
VALUES
-- Header for content sites (with category nav)
(
'content_header',
'Content Site Header',
'header',
'content-site',
ARRAY['content', 'publishing', 'news'],
'<header class="site-header site-header--content">
<div class="container">
<nav class="site-header__nav">
<a href="/" class="site-header__brand">{{.brand_name}}</a>
<ul class="site-header__categories">
{{range .categories}}
<li><a href="#{{.slug}}" class="site-header__category">{{.name}}</a></li>
{{end}}
</ul>
<div class="site-header__actions">
<button class="site-header__search-toggle" aria-label="Search">
<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
</button>
{{if .show_subscribe}}
<a href="#subscribe" class="button button--small button--primary">{{.subscribe_text}}</a>
{{end}}
</div>
</nav>
</div>
  </header>',
  '{"brand_name": "string", "categories": "array", "show_subscribe": "boolean", "subscribe_text": "string"}',
  10
),
-- Featured Article Hero
(
  'featured_article',
  'Featured Article Hero',
  'featured_content',
  'content-site',
  ARRAY['content', 'hero', 'featured'],
  '<section class="section section--featured">
    <div class="container">
      <article class="featured-article">
        <div class="featured-article__image">
          <img src="{{.featured_image}}" alt="{{.featured_title}}" loading="lazy">
          <span class="featured-article__category">{{.featured_category}}</span>
        </div>
        <div class="featured-article__content">
          <h1 class="featured-article__title">{{.featured_title}}</h1>
          <p class="featured-article__excerpt">{{.featured_excerpt}}</p>
          <div class="featured-article__meta">
            <span class="featured-article__author">{{.featured_author}}</span>
            <span class="featured-article__date">{{.featured_date}}</span>
            <span class="featured-article__read-time">{{.featured_read_time}}</span>
          </div>
          <a href="#" class="button button--primary">{{.read_more_text}}</a>
        </div>
      </article>
    </div>
  </section>',
  '{"featured_image": "string", "featured_title": "string", "featured_excerpt": "string", "featured_category": "string", "featured_author": "string", "featured_date": "string", "featured_read_time": "string", "read_more_text": "string"}',
  20
),
-- Article Grid
(
  'article_grid',
  'Article Grid',
  'content_listing',
  'content-site',
  ARRAY['content', 'grid', 'articles'],
  '<section class="section section--articles">
    <div class="container">
      <div class="section__header">
        <h2 class="section__title">{{.section_title}}</h2>
        {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
      </div>
      <div class="article-grid grid grid--3">
        {{range .articles}}
        <article class="article-card hover-lift">
          <div class="article-card__image">
            <img src="{{.image}}" alt="{{.title}}" loading="lazy">
            <span class="article-card__category">{{.category}}</span>
          </div>
          <div class="article-card__content">
            <h3 class="article-card__title">{{.title}}</h3>
            <p class="article-card__excerpt">{{.excerpt}}</p>
            <div class="article-card__meta">
              <span class="article-card__date">{{.date}}</span>
              <span class="article-card__read-time">{{.read_time}}</span>
            </div>
          </div>
        </article>
        {{end}}
      </div>
      {{if .show_load_more}}
      <div class="section__actions">
        <button class="button button--secondary">{{.load_more_text}}</button>
      </div>
      {{end}}
    </div>
  </section>',
  '{"section_title": "string", "section_subtitle": "string", "articles": "array", "show_load_more": "boolean", "load_more_text": "string"}',
  30
),
-- Sidebar with Ad Zone
(
  'content_sidebar',
  'Content Sidebar',
  'sidebar',
  'content-site',
  ARRAY['content', 'sidebar', 'ads'],
  '<aside class="sidebar">
    {{if .show_newsletter}}
    <div class="sidebar__widget sidebar__widget--newsletter">
      <h3 class="sidebar__title">{{.newsletter_title}}</h3>
      <p class="sidebar__text">{{.newsletter_description}}</p>
      <form class="newsletter-form">
        <input type="email" placeholder="{{.email_placeholder}}" class="newsletter-form__input" required>
        <button type="submit" class="button button--primary button--full-width">{{.subscribe_button}}</button>
      </form>
    </div>
    {{end}}

    {{if .show_popular}}
    <div class="sidebar__widget sidebar__widget--popular">
      <h3 class="sidebar__title">{{.popular_title}}</h3>
      <ul class="popular-list">
        {{range .popular_articles}}
        <li class="popular-list__item">
          <a href="#" class="popular-list__link">
            <span class="popular-list__number">{{.rank}}</span>
            <span class="popular-list__title">{{.title}}</span>
          </a>
        </li>
        {{end}}
      </ul>
    </div>
    {{end}}
    
    {{if .show_ad}}
    <div class="sidebar__widget sidebar__widget--ad">
      <div class="ad-zone ad-zone--sidebar" data-ad-slot="{{.ad_slot_id}}">
        <span class="ad-zone__label">Advertisement</span>
        <!-- Ad content inserted here -->
      </div>
    </div>
    {{end}}
    
    {{if .show_categories}}
    <div class="sidebar__widget sidebar__widget--categories">
      <h3 class="sidebar__title">{{.categories_title}}</h3>
      <ul class="category-list">
        {{range .category_links}}
        <li><a href="#{{.slug}}" class="category-list__link">{{.name}} <span class="category-list__count">({{.count}})</span></a></li>
        {{end}}
      </ul>
    </div>
    {{end}}
  </aside>',
  '{"show_newsletter": "boolean", "newsletter_title": "string", "newsletter_description": "string", "email_placeholder": "string", "subscribe_button": "string", "show_popular": "boolean", "popular_title": "string", "popular_articles": "array", "show_ad": "boolean", "ad_slot_id": "string", "show_categories": "boolean", "categories_title": "string", "category_links": "array"}',
  40
),
-- In-content Ad Zone
(
  'ad_zone_inline',
  'Inline Ad Zone',
  'advertising',
  'content-site',
  ARRAY['ads', 'monetization'],
  '<div class="ad-zone ad-zone--inline" data-ad-slot="{{.ad_slot_id}}">
    <span class="ad-zone__label">Advertisement</span>
    <!-- Ad content inserted here -->
  </div>',
  '{"ad_slot_id": "string"}',
  50
),
-- Category Section
(
  'category_section',
  'Category Section',
  'category_listing',
  'content-site',
  ARRAY['content', 'category', 'navigation'],
  '<section class="section section--category" id="{{.category_slug}}">
    <div class="container">
      <div class="section__header section__header--with-link">
        <h2 class="section__title">{{.category_name}}</h2>
        <a href="#" class="section__link">View all {{.category_name}} →</a>
      </div>
      <div class="article-grid grid grid--4">
        {{range .category_articles}}
        <article class="article-card article-card--compact hover-lift">
          <div class="article-card__image">
            <img src="{{.image}}" alt="{{.title}}" loading="lazy">
          </div>
          <div class="article-card__content">
            <h3 class="article-card__title">{{.title}}</h3>
            <span class="article-card__date">{{.date}}</span>
          </div>
        </article>
        {{end}}
      </div>
    </div>
  </section>',
  '{"category_slug": "string", "category_name": "string", "category_articles": "array"}',
  60
),
-- Content Site Footer
(
  'content_footer',
  'Content Site Footer',
  'footer',
  'content-site',
  ARRAY['content', 'footer'],
  '<footer class="site-footer site-footer--content">
    <div class="container">
      <div class="site-footer__grid grid grid--4">
        <div class="site-footer__about">
          <h3 class="site-footer__brand">{{.brand_name}}</h3>
          <p class="site-footer__tagline">{{.tagline}}</p>
          <div class="site-footer__social">
            {{range .social_links}}
            <a href="{{.url}}" class="site-footer__social-link" aria-label="{{.platform}}">{{.icon}}</a>
            {{end}}
          </div>
        </div>
        <div class="site-footer__links">
          <h4 class="site-footer__heading">Categories</h4>
          <ul class="site-footer__list">
            {{range .categories}}
            <li><a href="#{{.slug}}" class="site-footer__link">{{.name}}</a></li>
            {{end}}
          </ul>
        </div>
        <div class="site-footer__links">
          <h4 class="site-footer__heading">Company</h4>
          <ul class="site-footer__list">
            {{range .company_links}}
            <li><a href="{{.url}}" class="site-footer__link">{{.name}}</a></li>
            {{end}}
          </ul>
        </div>
        <div class="site-footer__newsletter">
          <h4 class="site-footer__heading">{{.newsletter_title}}</h4>
          <p class="site-footer__text">{{.newsletter_description}}</p>
          <form class="newsletter-form newsletter-form--footer">
            <input type="email" placeholder="{{.email_placeholder}}" class="newsletter-form__input">
            <button type="submit" class="button button--primary">→</button>
          </form>
        </div>
      </div>
      <div class="site-footer__bottom">
        <p>{{.copyright}}</p>
        <nav class="site-footer__legal">
          {{range .legal_links}}
          <a href="{{.url}}">{{.name}}</a>
          {{end}}
        </nav>
      </div>
    </div>
  </footer>',
  '{"brand_name": "string", "tagline": "string", "social_links": "array", "categories": "array", "company_links": "array", "newsletter_title": "string", "newsletter_description": "string", "email_placeholder": "string", "copyright": "string", "legal_links": "array"}',
  100
)
ON CONFLICT (name) DO UPDATE SET
  html_template = EXCLUDED.html_template,
  content_requirements = EXCLUDED.content_requirements,
  updated_at = now();

-- ============================================================================
-- 6. CONTENT SITE CSS THEME
-- ============================================================================

INSERT INTO css_themes (name, display_name, description, category, semantic_tags, css_content)
VALUES (
'content-modern',
'Modern Content',
'Clean, readable theme optimized for content sites and publishing',
'content',
ARRAY['content', 'publishing', 'readable', 'light-mode', 'minimal'],
':root {
--color-primary: #2563eb;
--color-primary-hover: #1d4ed8;
--color-primary-text: #ffffff;
--color-secondary: #64748b;
--color-secondary-hover: #475569;
--color-secondary-text: #ffffff;
--color-accent: #f59e0b;

    --color-text: #1e293b;
    --color-text-muted: #64748b;
    --color-heading: #0f172a;
    --color-background: #ffffff;
    --color-surface: #f8fafc;
    --color-border: #e2e8f0;

    --color-header-bg: #ffffff;
    --color-header-text: #0f172a;
    --color-card-bg: #ffffff;
    --color-footer-bg: #0f172a;
    --color-footer-text: #e2e8f0;

    --border-radius: 0.5rem;
    --shadow: 0 1px 3px rgba(0,0,0,0.1);
    --shadow-lg: 0 4px 12px rgba(0,0,0,0.1);
    
    --font-body: "Inter", -apple-system, sans-serif;
    --font-heading: "Inter", -apple-system, sans-serif;
    --font-size-base: 1rem;
    --line-height-body: 1.7;
    --line-height-heading: 1.3;
}

body {
font-family: var(--font-body);
font-size: var(--font-size-base);
line-height: var(--line-height-body);
}

/* Content site specific */
.site-header--content {
border-bottom: 1px solid var(--color-border);
box-shadow: none;
}

.site-header__categories {
display: flex;
gap: 1.5rem;
list-style: none;
}

.site-header__category {
color: var(--color-text-muted);
text-decoration: none;
font-weight: 500;
transition: color 0.2s;
}

.site-header__category:hover {
color: var(--color-primary);
}

.featured-article {
display: grid;
grid-template-columns: 1.5fr 1fr;
gap: 3rem;
align-items: center;
}

.featured-article__image {
position: relative;
border-radius: var(--border-radius);
overflow: hidden;
}

.featured-article__image img {
width: 100%;
height: 400px;
object-fit: cover;
}

.featured-article__category {
position: absolute;
top: 1rem;
left: 1rem;
background: var(--color-primary);
color: white;
padding: 0.25rem 0.75rem;
border-radius: 2rem;
font-size: 0.75rem;
font-weight: 600;
text-transform: uppercase;
}

.featured-article__title {
font-size: 2.5rem;
line-height: var(--line-height-heading);
margin-bottom: 1rem;
}

.featured-article__excerpt {
font-size: 1.125rem;
color: var(--color-text-muted);
margin-bottom: 1.5rem;
}

.featured-article__meta {
display: flex;
gap: 1rem;
color: var(--color-text-muted);
font-size: 0.875rem;
margin-bottom: 1.5rem;
}

.article-card {
background: var(--color-card-bg);
border-radius: var(--border-radius);
overflow: hidden;
box-shadow: var(--shadow);
}

.article-card__image {
position: relative;
}

.article-card__image img {
width: 100%;
height: 200px;
object-fit: cover;
}

.article-card__category {
position: absolute;
bottom: 0.75rem;
left: 0.75rem;
background: var(--color-primary);
color: white;
padding: 0.125rem 0.5rem;
border-radius: 2rem;
font-size: 0.625rem;
font-weight: 600;
text-transform: uppercase;
}

.article-card__content {
padding: 1.25rem;
}

.article-card__title {
font-size: 1.125rem;
line-height: var(--line-height-heading);
margin-bottom: 0.5rem;
}

.article-card__excerpt {
color: var(--color-text-muted);
font-size: 0.875rem;
margin-bottom: 0.75rem;
display: -webkit-box;
-webkit-line-clamp: 2;
-webkit-box-orient: vertical;
overflow: hidden;
}

.article-card__meta {
display: flex;
gap: 1rem;
color: var(--color-text-muted);
font-size: 0.75rem;
}

.article-card--compact .article-card__image img {
height: 140px;
}

.article-card--compact .article-card__content {
padding: 1rem;
}

.article-card--compact .article-card__title {
font-size: 1rem;
}

/* Sidebar */
.sidebar {
display: flex;
flex-direction: column;
gap: 2rem;
}

.sidebar__widget {
background: var(--color-surface);
padding: 1.5rem;
border-radius: var(--border-radius);
}

.sidebar__title {
font-size: 1rem;
margin-bottom: 1rem;
padding-bottom: 0.75rem;
border-bottom: 2px solid var(--color-primary);
}

.popular-list {
list-style: none;
}

.popular-list__item {
padding: 0.75rem 0;
border-bottom: 1px solid var(--color-border);
}

.popular-list__item:last-child {
border-bottom: none;
}

.popular-list__link {
display: flex;
gap: 1rem;
align-items: flex-start;
text-decoration: none;
color: var(--color-text);
}

.popular-list__number {
font-size: 1.5rem;
font-weight: 700;
color: var(--color-primary);
line-height: 1;
}

.popular-list__title {
font-size: 0.875rem;
line-height: 1.4;
}

.category-list {
list-style: none;
}

.category-list__link {
display: flex;
justify-content: space-between;
padding: 0.5rem 0;
text-decoration: none;
color: var(--color-text);
}

.category-list__count {
color: var(--color-text-muted);
}

/* Ad zones */
.ad-zone {
background: var(--color-surface);
border: 1px dashed var(--color-border);
border-radius: var(--border-radius);
padding: 1rem;
text-align: center;
min-height: 250px;
display: flex;
align-items: center;
justify-content: center;
}

.ad-zone__label {
font-size: 0.625rem;
text-transform: uppercase;
color: var(--color-text-muted);
letter-spacing: 0.1em;
}

.ad-zone--inline {
margin: 2rem 0;
min-height: 100px;
}

/* Newsletter form */
.newsletter-form {
display: flex;
flex-direction: column;
gap: 0.75rem;
}

.newsletter-form__input {
padding: 0.75rem 1rem;
border: 1px solid var(--color-border);
border-radius: var(--border-radius);
font-size: 0.875rem;
}

.newsletter-form--footer {
flex-direction: row;
}

.newsletter-form--footer .newsletter-form__input {
flex: 1;
}

/* Section header with link */
.section__header--with-link {
display: flex;
justify-content: space-between;
align-items: center;
margin-bottom: 2rem;
}

.section__link {
color: var(--color-primary);
text-decoration: none;
font-weight: 500;
}

@media (max-width: 768px) {
.featured-article {
grid-template-columns: 1fr;
}

    .featured-article__title {
      font-size: 1.75rem;
    }
    
    .site-header__categories {
      display: none;
    }
}'
)
ON CONFLICT (name) DO UPDATE SET
css_content = EXCLUDED.css_content,
semantic_tags = EXCLUDED.semantic_tags,
updated_at = now();

-- ============================================================================
-- 7. UPDATED WORKFLOW WITH CONDITIONAL ARCHITECT ROUTING
-- ============================================================================
-- Uses conditional_call_agent to route to the appropriate architect in one step

UPDATE agent_group_definitions
SET
updated_at = now(),
agent_configs = '[
{"role": "briefer", "agent_type": "briefing-agent"},
{"role": "chief_strategist", "agent_type": "chief-strategist"},
{"role": "architect", "agent_type": "landing-page-architect"},
{"role": "content_creator", "agent_type": "content-creator"},
{"role": "html_assembler", "agent_type": "html-assembler"},
{"role": "deployer", "agent_type": "deployer-agent"}
]'::jsonb,
orchestration_workflow = '{
"start_step": "spawn_briefer",
"steps": {
"spawn_briefer": {
"action": "spawn_agent",
"config": {"role": "briefer", "agent_type": "briefing-agent"},
"next_step": "spawn_strategist",
"description": "Spawn Briefing Agent"
},
"spawn_strategist": {
"action": "spawn_agent",
"config": {"role": "chief_strategist", "agent_type": "chief-strategist"},
"next_step": "spawn_content_creator",
"description": "Spawn Chief Strategist"
},
"spawn_content_creator": {
"action": "spawn_agent",
"config": {"role": "content_creator", "agent_type": "content-creator"},
"next_step": "spawn_assembler",
"description": "Spawn Content Creator"
},
"spawn_assembler": {
"action": "spawn_agent",
"config": {"role": "html_assembler", "agent_type": "html-assembler"},
"next_step": "spawn_deployer",
"description": "Spawn HTML Assembler"
},
"spawn_deployer": {
"action": "spawn_agent",
"config": {"role": "deployer", "agent_type": "deployer-agent"},
"next_step": "call_briefer",
"description": "Spawn Deployer"
},
"call_briefer": {
"action": "call_agent",
"description": "Get structured brief with site_type detection",
"config": {
"agent_type": "briefing-agent",
"target_role": "briefer",
"timeout_seconds": 60
},
"output_field": "brief_data",
"next_step": "call_strategist"
},
"call_strategist": {
"action": "call_agent",
"description": "Get the Build Plan from the Strategist",
"config": {
"agent_type": "chief-strategist",
"target_role": "chief_strategist",
"input_fields": ["brief_data", "input_data"],
"timeout_seconds": 120
},
"output_field": "build_plan_data",
"next_step": "call_architect"
},
"call_architect": {
"action": "conditional_call_agent",
"description": "Route to and call appropriate architect based on site_type",
"config": {
"field_path": "brief_data.structured_brief.result.site_type",
"agent_mapping": {
"landing": "landing-page-architect",
"content": "content-site-architect",
"portfolio": "portfolio-architect"
},
"default_agent": "landing-page-architect",
"input_fields": ["build_plan_data", "brief_data", "input_data"],
"timeout_seconds": 120
},
"output_field": "template_data",
"next_step": "call_content_creator"
},
"call_content_creator": {
"action": "call_agent",
"description": "Generate content JSON",
"config": {
"agent_type": "content-creator",
"target_role": "content_creator",
"input_fields": ["template_data", "build_plan_data", "brief_data", "input_data"],
"timeout_seconds": 300
},
"output_field": "content_data",
"next_step": "call_assembler"
},
"call_assembler": {
"action": "call_agent",
"description": "Assemble final HTML with CSS and JS",
"config": {
"agent_type": "html-assembler",
"target_role": "html_assembler",
"input_fields": ["content_data", "template_data", "brief_data", "input_data"],
"timeout_seconds": 120
},
"output_field": "final_html_data",
"next_step": "call_deployer"
},
"call_deployer": {
"action": "call_agent",
"description": "Push the final site to Git",
"config": {
"agent_type": "deployer-agent",
"target_role": "deployer",
"input_fields": ["final_html_data", "input_data"],
"timeout_seconds": 180
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Site build is complete."
}
}
}'::jsonb,
description = '[MVP v3] Multi-architect workflow: Routes to specialist architects (landing/content/portfolio) based on site_type'
WHERE group_type = 'mvp-site-builder';


===

updated:

-- ============================================================================
-- SPECIALIST ARCHITECT SYSTEM
-- ============================================================================

-- ============================================================================
-- 1. UPDATE BRIEFING AGENT - Now detects site_type
-- ============================================================================

UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "analyze_and_brief",
"steps": {
"analyze_and_brief": {
"action": "execute_llm_prompt",
"config": {
"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY",
"max_tokens": 3000
},
"input_fields": ["input_data"],
"output_field": "structured_brief",
"prompt_template": "Analyze this website request and create a comprehensive brief.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Model (if specified): {{.input_data.model}}\n\nFirst, determine the SITE TYPE based on the domain and objective:\n\n**landing** - Conversion-focused sites:\n- Product/service sales pages\n- SaaS landing pages\n- Lead generation\n- App download pages\n- Event registration\n- Single clear CTA goal\n\n**content** - Content/publishing sites:\n- News, blogs, magazines\n- Content aggregation\n- Ad-revenue models\n- SEO/traffic focused\n- Multiple articles/posts\n- Category navigation\n\n**portfolio** - Showcase sites:\n- Creative portfolios\n- Agency showcases\n- Case study focused\n- Gallery/visual heavy\n- Client testimonials\n\n**directory** - Listing sites:\n- Business directories\n- Job boards\n- Marketplace listings\n- Search/filter focused\n\nAnalyze the domain name and objective to determine site_type, then create a full brief.\n\nReturn JSON:\n{\n  \"site_type\": \"landing|content|portfolio|directory\",\n  \"site_type_confidence\": 0.0-1.0,\n  \"site_type_reasoning\": \"why this classification\",\n  \"analysis\": {\n    \"industry\": \"detected industry/niche\",\n    \"industry_confidence\": 0.0-1.0,\n    \"domain_interpretation\": \"what the domain name suggests\"\n  },\n  \"audience\": {\n    \"primary\": \"primary target audience\",\n    \"secondary\": \"secondary audience if applicable\",\n    \"demographics\": [\"age range\", \"profession\"],\n    \"psychographics\": [\"values\", \"motivations\", \"pain points\"]\n  },\n  \"brand\": {\n    \"tone\": \"professional|casual|technical|friendly|authoritative|playful\",\n    \"personality\": [\"trait1\", \"trait2\", \"trait3\"],\n    \"voice_examples\": [\"example phrase in brand voice\"]\n  },\n  \"messaging\": {\n    \"value_proposition\": \"core value proposition\",\n    \"key_messages\": [\"message1\", \"message2\", \"message3\"],\n    \"usps\": [\"unique selling point 1\", \"usp 2\"],\n    \"proof_points\": [\"credibility element 1\", \"element 2\"]\n  },\n  \"structure\": {\n    \"recommended_sections\": [\"section1\", \"section2\"],\n    \"priority_sections\": [\"most important sections\"],\n    \"optional_sections\": [\"sections that could be added\"]\n  },\n  \"theme\": {\n    \"recommended\": \"theme name\",\n    \"semantic_tags\": [\"tag1\", \"tag2\", \"tag3\"],\n    \"color_mood\": \"color feeling description\",\n    \"style_notes\": \"specific style recommendations\"\n  },\n  \"content_guidelines\": {\n    \"headline_style\": \"guidance for headlines\",\n    \"cta_style\": \"guidance for calls to action\",\n    \"avoid\": [\"things to avoid\"],\n    \"emphasize\": [\"things to emphasize\"]\n  },\n  \"monetization\": {\n    \"model\": \"subscription|ads|sales|leads|freemium\",\n    \"ad_zones\": [\"if ads: recommended ad placements\"]\n  }\n}\n\nReturn ONLY valid JSON."
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return the structured brief with site_type"
}
}
},
"processing_mode": "task",
"timeout_seconds": 60
}'::jsonb
WHERE type = 'briefing-agent';

-- ============================================================================
-- 2. RENAME CURRENT ARCHITECT → LANDING-PAGE-ARCHITECT
-- ============================================================================

-- First, create the new landing-page-architect as a copy
INSERT INTO agent_definitions (
id, type, display_name, description, category,
default_config, is_active, capabilities,
image_repository, image_tag, resources, topics, health_config
)
SELECT
gen_random_uuid(),
'landing-page-architect',
'Landing Page Architect',
'Assembles conversion-focused landing pages from component library (PAS, AIDA, etc.)',
category,
default_config,
true,
ARRAY['build', 'assemble', 'database', 'landing-page', 'conversion'],
image_repository, image_tag, resources, topics, health_config
FROM agent_definitions
WHERE type = 'site-component-architect'
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
description = EXCLUDED.description,
capabilities = EXCLUDED.capabilities,
updated_at = now();

-- ============================================================================
-- 3. CREATE CONTENT-SITE-ARCHITECT
-- ============================================================================

INSERT INTO agent_definitions (
id, type, display_name, description, category,
default_config, is_active, capabilities,
image_repository, image_tag, resources, topics, health_config
)
VALUES (
gen_random_uuid(),
'content-site-architect',
'Content Site Architect',
'Assembles content/publishing sites with article grids, category nav, and ad zones',
'data-driven',
'{
"workflow": {
"start_step": "assemble_content_site",
"steps": {
"assemble_content_site": {
"action": "assemble_from_library",
"config": {
"site_type": "content",
"input_fields": ["build_plan_data", "brief_data"],
"default_sections": ["header", "featured_article", "article_grid", "sidebar", "category_nav", "footer"],
"component_category": "content-site"
},
"next_step": "complete",
"description": "Assemble content site template from component library"
},
"complete": {
"action": "complete_workflow",
"description": "Return the content site template"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180
}'::jsonb,
true,
ARRAY['build', 'assemble', 'database', 'content-site', 'publishing'],
'docker.io/aqls/agent-chassis',
'v1.0.476',
'{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
'{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
'{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
)
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
updated_at = now();

-- ============================================================================
-- 4. CREATE PORTFOLIO-ARCHITECT (for future use)
-- ============================================================================

INSERT INTO agent_definitions (
id, type, display_name, description, category,
default_config, is_active, capabilities,
image_repository, image_tag, resources, topics, health_config
)
VALUES (
gen_random_uuid(),
'portfolio-architect',
'Portfolio Site Architect',
'Assembles portfolio/showcase sites with galleries, case studies, and visual layouts',
'data-driven',
'{
"workflow": {
"start_step": "assemble_portfolio_site",
"steps": {
"assemble_portfolio_site": {
"action": "assemble_from_library",
"config": {
"site_type": "portfolio",
"input_fields": ["build_plan_data", "brief_data"],
"default_sections": ["header", "hero_visual", "work_grid", "case_study", "client_logos", "about", "contact", "footer"],
"component_category": "portfolio-site"
},
"next_step": "complete",
"description": "Assemble portfolio site template from component library"
},
"complete": {
"action": "complete_workflow",
"description": "Return the portfolio site template"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180
}'::jsonb,
true,
ARRAY['build', 'assemble', 'database', 'portfolio-site', 'showcase'],
'docker.io/aqls/agent-chassis',
'v1.0.476',
'{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
'{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
'{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
)
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
updated_at = now();

-- ============================================================================
-- 5. CONTENT SITE COMPONENTS
-- ============================================================================

-- Add category for content site components
INSERT INTO content_components (name, display_name, function, category, semantic_tags, html_template, content_requirements, sort_order)
VALUES
-- Header for content sites (with category nav)
(
'content_header',
'Content Site Header',
'header',
'content-site',
ARRAY['content', 'publishing', 'news'],
'<header class="site-header site-header--content">
<div class="container">
<nav class="site-header__nav">
<a href="/" class="site-header__brand">{{.brand_name}}</a>
<ul class="site-header__categories">
{{range .categories}}
<li><a href="#{{.slug}}" class="site-header__category">{{.name}}</a></li>
{{end}}
</ul>
<div class="site-header__actions">
<button class="site-header__search-toggle" aria-label="Search">
<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
</button>
{{if .show_subscribe}}
<a href="#subscribe" class="button button--small button--primary">{{.subscribe_text}}</a>
{{end}}
</div>
</nav>
</div>
  </header>',
  '{"brand_name": "string", "categories": "array", "show_subscribe": "boolean", "subscribe_text": "string"}',
  10
),
-- Featured Article Hero
(
  'featured_article',
  'Featured Article Hero',
  'featured_content',
  'content-site',
  ARRAY['content', 'hero', 'featured'],
  '<section class="section section--featured">
    <div class="container">
      <article class="featured-article">
        <div class="featured-article__image">
          <img src="{{.featured_image}}" alt="{{.featured_title}}" loading="lazy">
          <span class="featured-article__category">{{.featured_category}}</span>
        </div>
        <div class="featured-article__content">
          <h1 class="featured-article__title">{{.featured_title}}</h1>
          <p class="featured-article__excerpt">{{.featured_excerpt}}</p>
          <div class="featured-article__meta">
            <span class="featured-article__author">{{.featured_author}}</span>
            <span class="featured-article__date">{{.featured_date}}</span>
            <span class="featured-article__read-time">{{.featured_read_time}}</span>
          </div>
          <a href="#" class="button button--primary">{{.read_more_text}}</a>
        </div>
      </article>
    </div>
  </section>',
  '{"featured_image": "string", "featured_title": "string", "featured_excerpt": "string", "featured_category": "string", "featured_author": "string", "featured_date": "string", "featured_read_time": "string", "read_more_text": "string"}',
  20
),
-- Article Grid
(
  'article_grid',
  'Article Grid',
  'content_listing',
  'content-site',
  ARRAY['content', 'grid', 'articles'],
  '<section class="section section--articles">
    <div class="container">
      <div class="section__header">
        <h2 class="section__title">{{.section_title}}</h2>
        {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
      </div>
      <div class="article-grid grid grid--3">
        {{range .articles}}
        <article class="article-card hover-lift">
          <div class="article-card__image">
            <img src="{{.image}}" alt="{{.title}}" loading="lazy">
            <span class="article-card__category">{{.category}}</span>
          </div>
          <div class="article-card__content">
            <h3 class="article-card__title">{{.title}}</h3>
            <p class="article-card__excerpt">{{.excerpt}}</p>
            <div class="article-card__meta">
              <span class="article-card__date">{{.date}}</span>
              <span class="article-card__read-time">{{.read_time}}</span>
            </div>
          </div>
        </article>
        {{end}}
      </div>
      {{if .show_load_more}}
      <div class="section__actions">
        <button class="button button--secondary">{{.load_more_text}}</button>
      </div>
      {{end}}
    </div>
  </section>',
  '{"section_title": "string", "section_subtitle": "string", "articles": "array", "show_load_more": "boolean", "load_more_text": "string"}',
  30
),
-- Sidebar with Ad Zone
(
  'content_sidebar',
  'Content Sidebar',
  'sidebar',
  'content-site',
  ARRAY['content', 'sidebar', 'ads'],
  '<aside class="sidebar">
    {{if .show_newsletter}}
    <div class="sidebar__widget sidebar__widget--newsletter">
      <h3 class="sidebar__title">{{.newsletter_title}}</h3>
      <p class="sidebar__text">{{.newsletter_description}}</p>
      <form class="newsletter-form">
        <input type="email" placeholder="{{.email_placeholder}}" class="newsletter-form__input" required>
        <button type="submit" class="button button--primary button--full-width">{{.subscribe_button}}</button>
      </form>
    </div>
    {{end}}

    {{if .show_popular}}
    <div class="sidebar__widget sidebar__widget--popular">
      <h3 class="sidebar__title">{{.popular_title}}</h3>
      <ul class="popular-list">
        {{range .popular_articles}}
        <li class="popular-list__item">
          <a href="#" class="popular-list__link">
            <span class="popular-list__number">{{.rank}}</span>
            <span class="popular-list__title">{{.title}}</span>
          </a>
        </li>
        {{end}}
      </ul>
    </div>
    {{end}}
    
    {{if .show_ad}}
    <div class="sidebar__widget sidebar__widget--ad">
      <div class="ad-zone ad-zone--sidebar" data-ad-slot="{{.ad_slot_id}}">
        <span class="ad-zone__label">Advertisement</span>
        <!-- Ad content inserted here -->
      </div>
    </div>
    {{end}}
    
    {{if .show_categories}}
    <div class="sidebar__widget sidebar__widget--categories">
      <h3 class="sidebar__title">{{.categories_title}}</h3>
      <ul class="category-list">
        {{range .category_links}}
        <li><a href="#{{.slug}}" class="category-list__link">{{.name}} <span class="category-list__count">({{.count}})</span></a></li>
        {{end}}
      </ul>
    </div>
    {{end}}
  </aside>',
  '{"show_newsletter": "boolean", "newsletter_title": "string", "newsletter_description": "string", "email_placeholder": "string", "subscribe_button": "string", "show_popular": "boolean", "popular_title": "string", "popular_articles": "array", "show_ad": "boolean", "ad_slot_id": "string", "show_categories": "boolean", "categories_title": "string", "category_links": "array"}',
  40
),
-- In-content Ad Zone
(
  'ad_zone_inline',
  'Inline Ad Zone',
  'advertising',
  'content-site',
  ARRAY['ads', 'monetization'],
  '<div class="ad-zone ad-zone--inline" data-ad-slot="{{.ad_slot_id}}">
    <span class="ad-zone__label">Advertisement</span>
    <!-- Ad content inserted here -->
  </div>',
  '{"ad_slot_id": "string"}',
  50
),
-- Category Section
(
  'category_section',
  'Category Section',
  'category_listing',
  'content-site',
  ARRAY['content', 'category', 'navigation'],
  '<section class="section section--category" id="{{.category_slug}}">
    <div class="container">
      <div class="section__header section__header--with-link">
        <h2 class="section__title">{{.category_name}}</h2>
        <a href="#" class="section__link">View all {{.category_name}} →</a>
      </div>
      <div class="article-grid grid grid--4">
        {{range .category_articles}}
        <article class="article-card article-card--compact hover-lift">
          <div class="article-card__image">
            <img src="{{.image}}" alt="{{.title}}" loading="lazy">
          </div>
          <div class="article-card__content">
            <h3 class="article-card__title">{{.title}}</h3>
            <span class="article-card__date">{{.date}}</span>
          </div>
        </article>
        {{end}}
      </div>
    </div>
  </section>',
  '{"category_slug": "string", "category_name": "string", "category_articles": "array"}',
  60
),
-- Content Site Footer
(
  'content_footer',
  'Content Site Footer',
  'footer',
  'content-site',
  ARRAY['content', 'footer'],
  '<footer class="site-footer site-footer--content">
    <div class="container">
      <div class="site-footer__grid grid grid--4">
        <div class="site-footer__about">
          <h3 class="site-footer__brand">{{.brand_name}}</h3>
          <p class="site-footer__tagline">{{.tagline}}</p>
          <div class="site-footer__social">
            {{range .social_links}}
            <a href="{{.url}}" class="site-footer__social-link" aria-label="{{.platform}}">{{.icon}}</a>
            {{end}}
          </div>
        </div>
        <div class="site-footer__links">
          <h4 class="site-footer__heading">Categories</h4>
          <ul class="site-footer__list">
            {{range .categories}}
            <li><a href="#{{.slug}}" class="site-footer__link">{{.name}}</a></li>
            {{end}}
          </ul>
        </div>
        <div class="site-footer__links">
          <h4 class="site-footer__heading">Company</h4>
          <ul class="site-footer__list">
            {{range .company_links}}
            <li><a href="{{.url}}" class="site-footer__link">{{.name}}</a></li>
            {{end}}
          </ul>
        </div>
        <div class="site-footer__newsletter">
          <h4 class="site-footer__heading">{{.newsletter_title}}</h4>
          <p class="site-footer__text">{{.newsletter_description}}</p>
          <form class="newsletter-form newsletter-form--footer">
            <input type="email" placeholder="{{.email_placeholder}}" class="newsletter-form__input">
            <button type="submit" class="button button--primary">→</button>
          </form>
        </div>
      </div>
      <div class="site-footer__bottom">
        <p>{{.copyright}}</p>
        <nav class="site-footer__legal">
          {{range .legal_links}}
          <a href="{{.url}}">{{.name}}</a>
          {{end}}
        </nav>
      </div>
    </div>
  </footer>',
  '{"brand_name": "string", "tagline": "string", "social_links": "array", "categories": "array", "company_links": "array", "newsletter_title": "string", "newsletter_description": "string", "email_placeholder": "string", "copyright": "string", "legal_links": "array"}',
  100
)
ON CONFLICT (name) DO UPDATE SET
  html_template = EXCLUDED.html_template,
  content_requirements = EXCLUDED.content_requirements,
  updated_at = now();

-- ============================================================================
-- 6. CONTENT SITE CSS THEME
-- ============================================================================

INSERT INTO css_themes (name, display_name, description, category, semantic_tags, css_content)
VALUES (
'content-modern',
'Modern Content',
'Clean, readable theme optimized for content sites and publishing',
'content',
ARRAY['content', 'publishing', 'readable', 'light-mode', 'minimal'],
':root {
--color-primary: #2563eb;
--color-primary-hover: #1d4ed8;
--color-primary-text: #ffffff;
--color-secondary: #64748b;
--color-secondary-hover: #475569;
--color-secondary-text: #ffffff;
--color-accent: #f59e0b;

    --color-text: #1e293b;
    --color-text-muted: #64748b;
    --color-heading: #0f172a;
    --color-background: #ffffff;
    --color-surface: #f8fafc;
    --color-border: #e2e8f0;

    --color-header-bg: #ffffff;
    --color-header-text: #0f172a;
    --color-card-bg: #ffffff;
    --color-footer-bg: #0f172a;
    --color-footer-text: #e2e8f0;

    --border-radius: 0.5rem;
    --shadow: 0 1px 3px rgba(0,0,0,0.1);
    --shadow-lg: 0 4px 12px rgba(0,0,0,0.1);
    
    --font-body: "Inter", -apple-system, sans-serif;
    --font-heading: "Inter", -apple-system, sans-serif;
    --font-size-base: 1rem;
    --line-height-body: 1.7;
    --line-height-heading: 1.3;
}

body {
font-family: var(--font-body);
font-size: var(--font-size-base);
line-height: var(--line-height-body);
}

/* Content site specific */
.site-header--content {
border-bottom: 1px solid var(--color-border);
box-shadow: none;
}

.site-header__categories {
display: flex;
gap: 1.5rem;
list-style: none;
}

.site-header__category {
color: var(--color-text-muted);
text-decoration: none;
font-weight: 500;
transition: color 0.2s;
}

.site-header__category:hover {
color: var(--color-primary);
}

.featured-article {
display: grid;
grid-template-columns: 1.5fr 1fr;
gap: 3rem;
align-items: center;
}

.featured-article__image {
position: relative;
border-radius: var(--border-radius);
overflow: hidden;
}

.featured-article__image img {
width: 100%;
height: 400px;
object-fit: cover;
}

.featured-article__category {
position: absolute;
top: 1rem;
left: 1rem;
background: var(--color-primary);
color: white;
padding: 0.25rem 0.75rem;
border-radius: 2rem;
font-size: 0.75rem;
font-weight: 600;
text-transform: uppercase;
}

.featured-article__title {
font-size: 2.5rem;
line-height: var(--line-height-heading);
margin-bottom: 1rem;
}

.featured-article__excerpt {
font-size: 1.125rem;
color: var(--color-text-muted);
margin-bottom: 1.5rem;
}

.featured-article__meta {
display: flex;
gap: 1rem;
color: var(--color-text-muted);
font-size: 0.875rem;
margin-bottom: 1.5rem;
}

.article-card {
background: var(--color-card-bg);
border-radius: var(--border-radius);
overflow: hidden;
box-shadow: var(--shadow);
}

.article-card__image {
position: relative;
}

.article-card__image img {
width: 100%;
height: 200px;
object-fit: cover;
}

.article-card__category {
position: absolute;
bottom: 0.75rem;
left: 0.75rem;
background: var(--color-primary);
color: white;
padding: 0.125rem 0.5rem;
border-radius: 2rem;
font-size: 0.625rem;
font-weight: 600;
text-transform: uppercase;
}

.article-card__content {
padding: 1.25rem;
}

.article-card__title {
font-size: 1.125rem;
line-height: var(--line-height-heading);
margin-bottom: 0.5rem;
}

.article-card__excerpt {
color: var(--color-text-muted);
font-size: 0.875rem;
margin-bottom: 0.75rem;
display: -webkit-box;
-webkit-line-clamp: 2;
-webkit-box-orient: vertical;
overflow: hidden;
}

.article-card__meta {
display: flex;
gap: 1rem;
color: var(--color-text-muted);
font-size: 0.75rem;
}

.article-card--compact .article-card__image img {
height: 140px;
}

.article-card--compact .article-card__content {
padding: 1rem;
}

.article-card--compact .article-card__title {
font-size: 1rem;
}

/* Sidebar */
.sidebar {
display: flex;
flex-direction: column;
gap: 2rem;
}

.sidebar__widget {
background: var(--color-surface);
padding: 1.5rem;
border-radius: var(--border-radius);
}

.sidebar__title {
font-size: 1rem;
margin-bottom: 1rem;
padding-bottom: 0.75rem;
border-bottom: 2px solid var(--color-primary);
}

.popular-list {
list-style: none;
}

.popular-list__item {
padding: 0.75rem 0;
border-bottom: 1px solid var(--color-border);
}

.popular-list__item:last-child {
border-bottom: none;
}

.popular-list__link {
display: flex;
gap: 1rem;
align-items: flex-start;
text-decoration: none;
color: var(--color-text);
}

.popular-list__number {
font-size: 1.5rem;
font-weight: 700;
color: var(--color-primary);
line-height: 1;
}

.popular-list__title {
font-size: 0.875rem;
line-height: 1.4;
}

.category-list {
list-style: none;
}

.category-list__link {
display: flex;
justify-content: space-between;
padding: 0.5rem 0;
text-decoration: none;
color: var(--color-text);
}

.category-list__count {
color: var(--color-text-muted);
}

/* Ad zones */
.ad-zone {
background: var(--color-surface);
border: 1px dashed var(--color-border);
border-radius: var(--border-radius);
padding: 1rem;
text-align: center;
min-height: 250px;
display: flex;
align-items: center;
justify-content: center;
}

.ad-zone__label {
font-size: 0.625rem;
text-transform: uppercase;
color: var(--color-text-muted);
letter-spacing: 0.1em;
}

.ad-zone--inline {
margin: 2rem 0;
min-height: 100px;
}

/* Newsletter form */
.newsletter-form {
display: flex;
flex-direction: column;
gap: 0.75rem;
}

.newsletter-form__input {
padding: 0.75rem 1rem;
border: 1px solid var(--color-border);
border-radius: var(--border-radius);
font-size: 0.875rem;
}

.newsletter-form--footer {
flex-direction: row;
}

.newsletter-form--footer .newsletter-form__input {
flex: 1;
}

/* Section header with link */
.section__header--with-link {
display: flex;
justify-content: space-between;
align-items: center;
margin-bottom: 2rem;
}

.section__link {
color: var(--color-primary);
text-decoration: none;
font-weight: 500;
}

@media (max-width: 768px) {
.featured-article {
grid-template-columns: 1fr;
}

    .featured-article__title {
      font-size: 1.75rem;
    }
    
    .site-header__categories {
      display: none;
    }
}'
)
ON CONFLICT (name) DO UPDATE SET
css_content = EXCLUDED.css_content,
semantic_tags = EXCLUDED.semantic_tags,
updated_at = now();

-- ============================================================================
-- 7. UPDATED WORKFLOW WITH CONDITIONAL ARCHITECT ROUTING
-- ============================================================================
-- Uses conditional_call_agent to route to the appropriate architect in one step

UPDATE agent_group_definitions
SET
updated_at = now(),
agent_configs = '[
{"role": "briefer", "agent_type": "briefing-agent"},
{"role": "chief_strategist", "agent_type": "chief-strategist"},
{"role": "architect", "agent_type": "landing-page-architect"},
{"role": "content_creator", "agent_type": "content-creator"},
{"role": "html_assembler", "agent_type": "html-assembler"},
{"role": "deployer", "agent_type": "deployer-agent"}
]'::jsonb,
orchestration_workflow = '{
"start_step": "spawn_briefer",
"steps": {
"spawn_briefer": {
"action": "spawn_agent",
"config": {"role": "briefer", "agent_type": "briefing-agent"},
"next_step": "spawn_strategist",
"description": "Spawn Briefing Agent"
},
"spawn_strategist": {
"action": "spawn_agent",
"config": {"role": "chief_strategist", "agent_type": "chief-strategist"},
"next_step": "spawn_content_creator",
"description": "Spawn Chief Strategist"
},
"spawn_content_creator": {
"action": "spawn_agent",
"config": {"role": "content_creator", "agent_type": "content-creator"},
"next_step": "spawn_assembler",
"description": "Spawn Content Creator"
},
"spawn_assembler": {
"action": "spawn_agent",
"config": {"role": "html_assembler", "agent_type": "html-assembler"},
"next_step": "spawn_deployer",
"description": "Spawn HTML Assembler"
},
"spawn_deployer": {
"action": "spawn_agent",
"config": {"role": "deployer", "agent_type": "deployer-agent"},
"next_step": "call_briefer",
"description": "Spawn Deployer"
},
"call_briefer": {
"action": "call_agent",
"description": "Get structured brief with site_type detection",
"config": {
"agent_type": "briefing-agent",
"target_role": "briefer",
"timeout_seconds": 60
},
"output_field": "brief_data",
"next_step": "call_strategist"
},
"call_strategist": {
"action": "call_agent",
"description": "Get the Build Plan from the Strategist",
"config": {
"agent_type": "chief-strategist",
"target_role": "chief_strategist",
"input_fields": ["brief_data", "input_data"],
"timeout_seconds": 120
},
"output_field": "build_plan_data",
"next_step": "call_architect"
},
"call_architect": {
"action": "conditional_call_agent",
"description": "Route to and call appropriate architect based on site_type",
"config": {
"field_path": "brief_data.structured_brief.result.site_type",
"agent_mapping": {
"landing": "landing-page-architect",
"content": "content-site-architect",
"portfolio": "portfolio-architect"
},
"default_agent": "landing-page-architect",
"input_fields": ["build_plan_data", "brief_data", "input_data"],
"timeout_seconds": 120
},
"output_field": "template_data",
"next_step": "call_content_creator"
},
"call_content_creator": {
"action": "call_agent",
"description": "Generate content JSON",
"config": {
"agent_type": "content-creator",
"target_role": "content_creator",
"input_fields": ["template_data", "build_plan_data", "brief_data", "input_data"],
"timeout_seconds": 300
},
"output_field": "content_data",
"next_step": "call_assembler"
},
"call_assembler": {
"action": "call_agent",
"description": "Assemble final HTML with CSS and JS",
"config": {
"agent_type": "html-assembler",
"target_role": "html_assembler",
"input_fields": ["content_data", "template_data", "brief_data", "input_data"],
"timeout_seconds": 120
},
"output_field": "final_html_data",
"next_step": "call_deployer"
},
"call_deployer": {
"action": "call_agent",
"description": "Push the final site to Git",
"config": {
"agent_type": "deployer-agent",
"target_role": "deployer",
"input_fields": ["final_html_data", "input_data"],
"timeout_seconds": 180
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Site build is complete."
}
}
}'::jsonb,
description = '[MVP v3] Multi-architect workflow: Routes to specialist architects (landing/content/portfolio) based on site_type'
WHERE group_type = 'mvp-site-builder';