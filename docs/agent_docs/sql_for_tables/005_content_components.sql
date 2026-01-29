clients_db=# \d content_components;
Table "public.content_components"
Column      |           Type           | Collation | Nullable |          Default
------------------+--------------------------+-----------+----------+----------------------------
 id               | uuid                     |           | not null | gen_random_uuid()
 name             | text                     |           | not null |
 description      | text                     |           |          |
 html_template    | text                     |           | not null |
 input_schema     | jsonb                    |           |          |
 function         | text                     |           | not null | 'generic-text-block'::text
 created_at       | timestamp with time zone |           |          | now()
 updated_at       | timestamp with time zone |           |          | now()
 display_name     | text                     |           |          |
 category         | character varying(255)   |           |          |
 semantic_tags    | jsonb                    |           |          |
 sort_order       | integer                  |           |          |
 render_mode      | text                     |           |          | 'template'::text
 agent_type       | text                     |           |          |
 agent_workflow   | text                     |           |          |
 data_sources     | text[]                   |           |          |
 child_components | jsonb                    |           |          |
 component_level  | text                     |           |          | 'section'::text
 is_active        | boolean                  |           |          | true
Indexes:
    "content_components_pkey" PRIMARY KEY, btree (id)
    "content_components_name_key" UNIQUE CONSTRAINT, btree (name)
    "idx_components_agent_type" btree (agent_type) WHERE agent_type IS NOT NULL
    "idx_components_level" btree (component_level)
    "idx_components_render_mode" btree (render_mode)
    "idx_content_components_category" btree (category)
    "idx_content_components_function" btree (function)
Referenced by:
    TABLE "page_components" CONSTRAINT "page_components_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_footer_component_id_fkey" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_footer_fk" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_component_id_fkey" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_fk" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_home_component_id_fkey" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_home_fk" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)


-- set active for older records
clients_db=# UPDATE content_components SET is_active = true WHERE is_active IS NULL;


UPDATE content_components
    SET display_name = COALESCE(display_name, name),
        category = COALESCE(category, '')
    WHERE display_name IS NULL OR category IS NULL;

-- And optionally add NOT NULL constraints with defaults:

ALTER TABLE content_components
    ALTER COLUMN display_name SET DEFAULT '',
ALTER COLUMN category SET DEFAULT '';

      --

      add missing components

-- ==============================================================================
-- ISSUE 30b: Add missing components to content_components table
-- ==============================================================================
-- The page requests sections: hero, features, social_proof, call_to_action
-- But these components don't exist in the database, causing stub creation.
--
-- This script adds the missing components with proper templates and input schemas.
-- ==============================================================================

-- 1. Hero Section
INSERT INTO content_components (name, function, display_name, description, html_template, input_schema, category)
VALUES (
    'hero',
    'hero',
    'Hero Section',
    'Main hero banner with compelling headline, subheadline, and call-to-action buttons for homepage and landing pages.',
    '<section id="{{.ComponentID}}" class="section section--hero">
  <div class="container hero">
    <h1 class="hero__title">{{.headline}}</h1>
    <p class="hero__subtitle">{{.subheadline}}</p>
    <div class="hero__actions">
      <a href="{{.primary_cta_url}}" class="button button--primary button--large">{{.primary_cta}}</a>
      {{if .secondary_cta}}<a href="{{.secondary_cta_url}}" class="button button--secondary button--large">{{.secondary_cta}}</a>{{end}}
    </div>
  </div>
</section>
<style>
.section--hero {
  padding: 6rem 2rem;
  background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 100%);
  color: white;
  text-align: center;
}
.hero__title {
  font-size: 3rem;
  margin-bottom: 1rem;
  font-weight: 700;
}
.hero__subtitle {
  font-size: 1.25rem;
  opacity: 0.9;
  max-width: 600px;
  margin: 0 auto 2rem;
}
.hero__actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}
.button {
  padding: 0.75rem 1.5rem;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 600;
  transition: transform 0.2s, box-shadow 0.2s;
}
.button--primary {
  background: var(--accent-color, #0f3460);
  color: white;
}
.button--secondary {
  background: transparent;
  border: 2px solid white;
  color: white;
}
.button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}
@media (max-width: 768px) {
  .hero__title { font-size: 2rem; }
  .hero__subtitle { font-size: 1rem; }
}
</style>',
    '{
  "type": "object",
  "required": ["headline", "subheadline", "primary_cta"],
  "properties": {
    "headline": {"type": "string", "description": "Main headline - bold, attention-grabbing"},
    "subheadline": {"type": "string", "description": "Supporting text that expands on the headline"},
    "primary_cta": {"type": "string", "description": "Primary call-to-action button text"},
    "primary_cta_url": {"type": "string", "description": "URL for primary CTA", "default": "#contact"},
    "secondary_cta": {"type": "string", "description": "Optional secondary CTA text"},
    "secondary_cta_url": {"type": "string", "description": "URL for secondary CTA", "default": "#features"}
  }
}'::jsonb,
    'hero'
)
ON CONFLICT (name) DO UPDATE SET
    function = EXCLUDED.function,
                                display_name = EXCLUDED.display_name,
                                description = EXCLUDED.description,
                                html_template = EXCLUDED.html_template,
                                input_schema = EXCLUDED.input_schema,
                                category = EXCLUDED.category,
                                updated_at = NOW();

-- 2. Features Grid
INSERT INTO content_components (name, function, display_name, description, html_template, input_schema, category)
VALUES (
           'features',
           'features',
           'Features Grid',
           'A grid layout showcasing key features or services with icons, titles, and descriptions.',
           '<section id="{{.ComponentID}}" class="section section--features">
         <div class="container">
           <h2 class="section__title section__title--center">{{.section_title}}</h2>
           {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
           <div class="features grid grid--3">
             {{range .features}}
             <div class="feature card">
               <div class="feature__icon">{{.icon}}</div>
               <h3 class="feature__title">{{.title}}</h3>
               <p class="feature__description">{{.description}}</p>
             </div>
             {{end}}
           </div>
         </div>
       </section>
       <style>
       .section--features {
         padding: 5rem 2rem;
         background: var(--background-color, #ffffff);
       }
       .section__title--center {
         text-align: center;
         margin-bottom: 0.5rem;
       }
       .section__subtitle {
         text-align: center;
         color: #666;
         max-width: 600px;
         margin: 0 auto 3rem;
       }
       .grid--3 {
         display: grid;
         grid-template-columns: repeat(3, 1fr);
         gap: 2rem;
       }
       .feature.card {
         padding: 2rem;
         background: white;
         border-radius: 8px;
         box-shadow: 0 2px 10px rgba(0,0,0,0.05);
         text-align: center;
         transition: transform 0.2s, box-shadow 0.2s;
       }
       .feature.card:hover {
         transform: translateY(-4px);
         box-shadow: 0 8px 25px rgba(0,0,0,0.1);
       }
       .feature__icon {
         font-size: 2.5rem;
         margin-bottom: 1rem;
         color: var(--accent-color, #0f3460);
       }
       .feature__title {
         font-size: 1.25rem;
         margin-bottom: 0.75rem;
       }
       .feature__description {
         color: #666;
         line-height: 1.6;
       }
       @media (max-width: 768px) {
         .grid--3 { grid-template-columns: 1fr; }
       }
       </style>',
           '{
         "type": "object",
         "required": ["section_title", "features"],
         "properties": {
           "section_title": {"type": "string", "description": "Section heading"},
           "section_subtitle": {"type": "string", "description": "Optional supporting text"},
           "features": {
             "type": "array",
             "description": "Array of feature items (3-6 recommended)",
             "items": {
               "type": "object",
               "properties": {
                 "icon": {"type": "string", "description": "Emoji or icon character"},
                 "title": {"type": "string", "description": "Feature name"},
                 "description": {"type": "string", "description": "Brief description"}
               }
             }
           }
         }
       }'::jsonb,
           'features'
       )
    ON CONFLICT (name) DO UPDATE SET
    function = EXCLUDED.function,
                              display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              category = EXCLUDED.category,
                              updated_at = NOW();

-- 3. Social Proof / Testimonials
INSERT INTO content_components (name, function, display_name, description, html_template, input_schema, category)
VALUES (
           'social_proof',
           'social_proof',
           'Social Proof',
           'Testimonials, statistics, or trust indicators that build credibility.',
           '<section id="{{.ComponentID}}" class="section section--social-proof">
         <div class="container">
           {{if .stat_number}}
           <div class="stat-highlight">
             <div class="stat-highlight__number">{{.stat_number}}</div>
             <div class="stat-highlight__label">{{.stat_label}}</div>
           </div>
           {{end}}
           <div class="testimonials grid grid--3">
             {{range .testimonials}}
             <div class="testimonial card">
               <div class="testimonial__rating">★★★★★</div>
               <p class="testimonial__text">"{{.quote}}"</p>
               <p class="testimonial__author">{{.author}}</p>
               {{if .role}}<p class="testimonial__role">{{.role}}</p>{{end}}
             </div>
             {{end}}
           </div>
         </div>
       </section>
       <style>
       .section--social-proof {
         padding: 5rem 2rem;
         background: #f8f9fa;
       }
       .stat-highlight {
         text-align: center;
         margin-bottom: 3rem;
       }
       .stat-highlight__number {
         font-size: 4rem;
         font-weight: 700;
         color: var(--primary-color, #1a1a2e);
       }
       .stat-highlight__label {
         font-size: 1.25rem;
         color: #666;
       }
       .testimonial.card {
         padding: 2rem;
         background: white;
         border-radius: 8px;
         box-shadow: 0 2px 10px rgba(0,0,0,0.05);
       }
       .testimonial__rating {
         color: #ffc107;
         margin-bottom: 1rem;
       }
       .testimonial__text {
         font-style: italic;
         color: #333;
         line-height: 1.6;
         margin-bottom: 1rem;
       }
       .testimonial__author {
         font-weight: 600;
         margin-bottom: 0.25rem;
       }
       .testimonial__role {
         color: #666;
         font-size: 0.9rem;
       }
       @media (max-width: 768px) {
         .grid--3 { grid-template-columns: 1fr; }
         .stat-highlight__number { font-size: 3rem; }
       }
       </style>',
           '{
         "type": "object",
         "properties": {
           "stat_number": {"type": "string", "description": "Key statistic (e.g., 100+, 5000+)"},
           "stat_label": {"type": "string", "description": "What the stat represents (e.g., Satisfied Clients)"},
           "testimonials": {
             "type": "array",
             "description": "Array of testimonial items (2-3 recommended)",
             "items": {
               "type": "object",
               "properties": {
                 "quote": {"type": "string", "description": "The testimonial text"},
                 "author": {"type": "string", "description": "Person name"},
                 "role": {"type": "string", "description": "Job title or company"}
               }
             }
           }
         }
       }'::jsonb,
           'social-proof'
       )
    ON CONFLICT (name) DO UPDATE SET
    function = EXCLUDED.function,
                              display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              category = EXCLUDED.category,
                              updated_at = NOW();

-- 4. Call to Action
INSERT INTO content_components (name, function, display_name, description, html_template, input_schema, category)
VALUES (
           'call_to_action',
           'call_to_action',
           'Call to Action',
           'A prominent CTA section encouraging visitors to take the next step.',
           '<section id="{{.ComponentID}}" class="section section--cta">
         <div class="container cta">
           <h2 class="cta__title">{{.headline}}</h2>
           <p class="cta__text">{{.subheadline}}</p>
           <div class="cta__actions">
             <a href="{{.cta_url}}" class="button button--primary button--large">{{.cta_text}}</a>
             {{if .secondary_text}}<p class="cta__secondary">{{.secondary_text}}</p>{{end}}
           </div>
         </div>
       </section>
       <style>
       .section--cta {
         padding: 5rem 2rem;
         background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 100%);
         color: white;
         text-align: center;
       }
       .cta__title {
         font-size: 2.5rem;
         margin-bottom: 1rem;
       }
       .cta__text {
         font-size: 1.25rem;
         opacity: 0.9;
         max-width: 600px;
         margin: 0 auto 2rem;
       }
       .cta__actions .button {
         padding: 1rem 2rem;
         font-size: 1.1rem;
       }
       .cta__secondary {
         margin-top: 1rem;
         opacity: 0.7;
         font-size: 0.9rem;
       }
       @media (max-width: 768px) {
         .cta__title { font-size: 1.75rem; }
         .cta__text { font-size: 1rem; }
       }
       </style>',
           '{
         "type": "object",
         "required": ["headline", "cta_text"],
         "properties": {
           "headline": {"type": "string", "description": "Compelling headline (e.g., Ready to Get Started?)"},
           "subheadline": {"type": "string", "description": "Supporting text explaining the value"},
           "cta_text": {"type": "string", "description": "Button text (e.g., Contact Us, Get Started)"},
           "cta_url": {"type": "string", "description": "URL for the CTA button", "default": "#contact"},
           "secondary_text": {"type": "string", "description": "Optional text below button (e.g., No credit card required)"}
         }
       }'::jsonb,
           'cta'
       )
    ON CONFLICT (name) DO UPDATE SET
    function = EXCLUDED.function,
                              display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              category = EXCLUDED.category,
                              updated_at = NOW();

-- Verify the inserts
SELECT name, function, display_name,
       CASE WHEN html_template IS NOT NULL THEN 'yes' ELSE 'no' END as has_template,
       CASE WHEN input_schema IS NOT NULL THEN 'yes' ELSE 'no' END as has_schema
FROM content_components
WHERE name IN ('hero', 'features', 'social_proof', 'call_to_action')
ORDER BY name;

--
-- adding some images to hero components
UPDATE content_components
SET
    html_template = '<section class="hero" data-component="hero"{{if .background_image}} style="background-image: url(''{{.background_image}}'');"{{end}}>
        <div class="hero-overlay"></div>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .cta_text}}<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
            {{if .secondary_cta}}<a href="{{if .secondary_cta_url}}{{.secondary_cta_url}}{{else}}#features{{end}}" class="btn btn-secondary">{{.secondary_cta}}</a>{{end}}
        </div>
    </section>',
    input_schema = '{"headline": "string", "subheadline": "string", "cta_text": "string", "cta_url": "string", "secondary_cta": "string", "secondary_cta_url": "string", "background_image": "string (optional URL)"}'
WHERE name = 'hero';

-- again adds gradient as fallback

-- Update hero component to support background images
UPDATE content_components
SET
    html_template = '<section class="hero" data-component="hero"{{if .background_image}} style="background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url(''{{.background_image}}''); background-size: cover; background-position: center;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .cta_text}}<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
            {{if .secondary_cta}}<a href="{{if .secondary_cta_url}}{{.secondary_cta_url}}{{else}}#features{{end}}" class="btn btn-secondary">{{.secondary_cta}}</a>{{end}}
        </div>
    </section>',
    input_schema = '{"headline": "string", "subheadline": "string", "cta_text": "string", "cta_url": "string", "secondary_cta": "string", "secondary_cta_url": "string", "background_image": "string (optional URL)"}',
    updated_at = NOW()
WHERE name = 'hero';

--
-- add some images

-- Update hero template with CSS gradient fallback when no image
UPDATE content_components
SET
    html_template = '<section class="hero" data-component="hero" style="{{if .background_image}}background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url(''{{.background_image}}''); background-size: cover; background-position: center;{{else}}background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%);{{end}}">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .cta_text}}<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
            {{if .secondary_cta}}<a href="{{if .secondary_cta_url}}{{.secondary_cta_url}}{{else}}#features{{end}}" class="btn btn-secondary">{{.secondary_cta}}</a>{{end}}
        </div>
    </section>',
    input_schema = '{"headline": "string", "subheadline": "string", "cta_text": "string", "cta_url": "string", "secondary_cta": "string", "secondary_cta_url": "string", "background_image": "string (optional URL)"}',
    updated_at = NOW()
WHERE name = 'hero';

--

-- the two together
UPDATE content_components
SET
    html_template = '<section class="hero" data-component="hero" style="{{if .background_image}}background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url(''{{.background_image}}''); background-size: cover; background-position: center;{{else}}background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%);{{end}}">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .cta_text}}<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
            {{if .secondary_cta}}<a href="{{if .secondary_cta_url}}{{.secondary_cta_url}}{{else}}#features{{end}}" class="btn btn-secondary">{{.secondary_cta}}</a>{{end}}
        </div>
    </section>',
    input_schema = '{"headline": "string", "subheadline": "string", "cta_text": "string", "cta_url": "string", "secondary_cta": "string", "secondary_cta_url": "string", "background_image": "string (optional URL)"}',
    updated_at = NOW()
WHERE name = 'hero';

-- latest?
UPDATE content_components
SET
    html_template = '<section class="hero" data-component="hero" style="{{if .background_image}}background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url(''{{.background_image}}''); background-size: cover; background-position: center;{{else}}background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%);{{end}}">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .cta_text}}<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
            {{if .secondary_cta}}<a href="{{if .secondary_cta_url}}{{.secondary_cta_url}}{{else}}#features{{end}}" class="btn btn-secondary">{{.secondary_cta}}</a>{{end}}
        </div>
    </section>',
    input_schema = '{"headline": "string", "subheadline": "string", "cta_text": "string", "cta_url": "string", "secondary_cta": "string", "secondary_cta_url": "string", "background_image": "string (optional URL)"}',
    updated_at = NOW()
WHERE name = 'hero';

-- template format add .
-- Find and fix footer templates that use {{logo_text}} instead of {{.logo_text}}
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(html_template, '{{logo_text}}', '{{.logo_text}}'),
                '{{company_name}}', '{{.company_name}}'
        ),
        '{{tagline}}', '{{.tagline}}'
                    )
WHERE html_template LIKE '%{{logo_text}}%'
   OR html_template LIKE '%{{company_name}}%'
   OR html_template LIKE '%{{tagline}}%';

---------------

-- templating fixes

-- Convert Handlebars {{#each services}} to Go template {{range .services}}
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(
                        REPLACE(
                                REPLACE(html_template,
                                        '{{#each services}}', '{{range .services}}'),
                                '{{/each}}', '{{end}}'),
                        '{{this.slug}}', '{{.slug}}'),
                '{{this.name}}', '{{.name}}'),
        '{{#each nav_items}}', '{{range .nav_items}}'
                    )
WHERE html_template LIKE '%{{#each%}'
   OR html_template LIKE '%{{/each%}}';


=======================
-------------------------

-- ============================================================================
-- FIX: Footer template using Handlebars syntax that isn't processed
-- ============================================================================
--
-- PROBLEM:
--   Footer template uses {{#each services}}...{{/each}} Handlebars syntax
--   Go's template engine uses {{range .services}}...{{end}} syntax
--   The renderEachBlocks function only handles nav_items, not services
--
-- SOLUTION OPTIONS:
--   A) Update template to use Go syntax {{range .services}}
--   B) Update template to use pre-rendered {{.services_html}}
--   C) Remove dynamic services list (use static or nav_items for footer)
--
-- ============================================================================

-- First, let's see what the current footer template looks like
SELECT id, name, html_template
FROM content_components
WHERE name LIKE '%footer%' OR function = 'site-footer'
    LIMIT 5;

-- ============================================================================
-- OPTION A: Convert to Go template syntax (if using executeGoTemplate)
-- ============================================================================
-- This requires the services data to be in the context as a slice of maps
-- with "name" and "slug" fields

UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(html_template,
                        '{{#each services}}',
                        '{{range .services}}'
                ),
                '{{/each}}',
                '{{end}}'
        ),
        '{{this.',
        '{{.'
                    )
WHERE name LIKE '%footer%'
  AND html_template LIKE '%{{#each services}}%';

-- ============================================================================
-- OPTION B: Use pre-rendered nav_items_html for quick links AND services
-- ============================================================================
-- This is simpler - just use navigation for both sections

UPDATE content_components
SET html_template = REPLACE(
        html_template,
        '<ul>
                    {{#each services}}
                    <li><a href="/services.html#{{this.slug}}">{{this.name}}</a></li>
                    {{/each}}
                </ul>',
        '<ul>{{.nav_items_html}}</ul>'
                    )
WHERE name LIKE '%footer%'
  AND html_template LIKE '%{{#each services}}%';

-- ============================================================================
-- OPTION C: Replace Handlebars with static placeholder message
-- ============================================================================
-- Useful for debugging or if services aren't critical
-- didn't do this
/*UPDATE content_components
SET html_template = REPLACE(
        html_template,
        '{{#each services}}
                    <li><a href="/services.html#{{this.slug}}">{{this.name}}</a></li>
                    {{/each}}',
        '<!-- Services will be populated dynamically -->'
                    )
WHERE name LIKE '%footer%'
  AND html_template LIKE '%{{#each services}}%';

-- ============================================================================
-- OPTION D: Update to use nav_items for Quick Links section too
-- ============================================================================

-- Check if quick links section is also empty
SELECT id, name,
       CASE WHEN html_template LIKE '%Quick Links%' THEN 'Has Quick Links' ELSE 'No Quick Links' END as has_quick_links,
       CASE WHEN html_template LIKE '%{{#each%' THEN 'Has Handlebars' ELSE 'No Handlebars' END as has_handlebars
FROM content_components
WHERE name LIKE '%footer%';

-- ============================================================================
-- RECOMMENDED: Full footer template replacement
-- ============================================================================
-- This replaces the entire footer template with one that uses Go template
-- syntax and includes proper fallbacks

-- First backup the old template
-- INSERT INTO component_versions (component_id, html_template, version_note, created_at)
-- SELECT id, html_template, 'Backup before Handlebars fix', NOW()
-- FROM content_components WHERE name = 'footer-4-column';

-- Then update with fixed template
UPDATE content_components
SET html_template = E'<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand">
            <h3>{{.logo_text}}</h3>
            <p>{{.tagline}}</p>
        </div>
        <div class="footer-links">
            <h4>Quick Links</h4>
            <ul>{{.nav_items_html}}</ul>
        </div>
        <div class="footer-services">
            <h4>Services</h4>
            <ul>{{.nav_items_html}}</ul>
        </div>
        <div class="footer-contact">
            <h4>Contact</h4>
            {{if .email}}<p><a href="mailto:{{.email}}">{{.email}}</a></p>{{end}}
            {{if .phone}}<p>{{.phone}}</p>{{end}}
        </div>
    </div>
    <div class="footer-bottom">
        <div class="footer-bottom-container">
            <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>
            <div class="footer-legal">
                <a href="/privacy.html">Privacy Policy</a>
                <a href="/terms.html">Terms of Service</a>
            </div>
        </div>
    </div>
</footer>
<style>
.site-footer {
    background: {{if .primary_color}}{{.primary_color}}{{else}}#1a1a2e{{end}};
    color: rgba(255,255,255,0.9);
    padding: 4rem 0 0;
    margin-top: auto;
}
.footer-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: grid;
    grid-template-columns: 2fr 1fr 1fr 1fr;
    gap: 3rem;
}
.footer-brand h3 {
    color: #fff;
    margin: 0 0 0.75rem;
    font-size: 1.25rem;
}
.footer-brand p {
    color: rgba(255,255,255,0.7);
    margin: 0 0 1.5rem;
    line-height: 1.6;
}
.footer-links h4,
.footer-services h4,
.footer-contact h4 {
    color: #fff;
    margin: 0 0 1rem;
    font-size: 1rem;
    font-weight: 600;
}
.footer-links ul,
.footer-services ul {
    list-style: none;
    padding: 0;
    margin: 0;
}
.footer-links li,
.footer-services li {
    margin-bottom: 0.5rem;
}
.footer-links a,
.footer-services a,
.footer-contact a {
    color: rgba(255,255,255,0.7);
    text-decoration: none;
    transition: color 0.2s;
}
.footer-links a:hover,
.footer-services a:hover,
.footer-contact a:hover {
    color: #fff;
}
.footer-contact p {
    margin: 0 0 0.5rem;
    color: rgba(255,255,255,0.7);
}
.footer-bottom {
    margin-top: 3rem;
    padding: 1.5rem 0;
    border-top: 1px solid rgba(255,255,255,0.1);
}
.footer-bottom-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.footer-bottom p {
    margin: 0;
    color: rgba(255,255,255,0.5);
    font-size: 0.9rem;
}
.footer-legal {
    display: flex;
    gap: 2rem;
}
.footer-legal a {
    color: rgba(255,255,255,0.5);
    text-decoration: none;
    font-size: 0.9rem;
}
.footer-legal a:hover {
    color: rgba(255,255,255,0.8);
}
@media (max-width: 768px) {
    .footer-container {
        grid-template-columns: 1fr;
        gap: 2rem;
    }
    .footer-bottom-container {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
    }
}
</style>'
WHERE name = 'footer-4-column';

-- Verify the update
SELECT name, LENGTH(html_template) as template_length,
       CASE WHEN html_template LIKE '%{{#each%' THEN 'STILL HAS HANDLEBARS' ELSE 'OK' END as status
FROM content_components
WHERE name LIKE '%footer%';*/

=======
-------

-- ============================================================================
-- FIX: footer-4-column template - Convert Handlebars to Go template syntax
-- ============================================================================
--
-- PROBLEMS:
--   1. {{#each services}}...{{/each}} - Handlebars, not supported
--   2. {{#if social_links}}...{{/if}} - Handlebars, not supported
--   3. {{#if contact_email}}...{{/if}} - Handlebars, not supported
--   4. Mixed placeholder styles (some with dot, some without)
--
-- SOLUTION:
--   Convert to Go template syntax or use pre-rendered HTML placeholders
--
-- ============================================================================

-- Update footer-4-column to use consistent Go template syntax
UPDATE content_components
SET html_template = E'<!-- FOOTER SOURCE: component-db:footer-4-column -->
<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand">
            <h3>{{.logo_text}}</h3>
            <p>{{.tagline}}</p>
        </div>
        <div class="footer-links">
            <h4>Quick Links</h4>
            <ul>
                {{.nav_items_html}}
            </ul>
        </div>
        <div class="footer-services">
            <h4>Services</h4>
            <ul>
                {{.nav_items_html}}
            </ul>
        </div>
        <div class="footer-contact">
            <h4>Contact</h4>
            {{if .email}}<p><a href="mailto:{{.email}}">{{.email}}</a></p>{{end}}
            {{if .phone}}<p><a href="tel:{{.phone}}">{{.phone}}</a></p>{{end}}
        </div>
    </div>
    <div class="footer-bottom">
        <div class="footer-bottom-container">
            <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>
            <div class="footer-legal">
                <a href="/privacy.html">Privacy Policy</a>
                <a href="/terms.html">Terms of Service</a>
            </div>
        </div>
    </div>
</footer>
<style>
.site-footer {
    background: {{.primary_color}};
    color: rgba(255,255,255,0.9);
    padding: 4rem 0 0;
    margin-top: auto;
}
.footer-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: grid;
    grid-template-columns: 2fr 1fr 1fr 1fr;
    gap: 3rem;
}
.footer-brand h3 {
    color: #fff;
    margin: 0 0 0.75rem;
    font-size: 1.25rem;
}
.footer-brand p {
    color: rgba(255,255,255,0.7);
    margin: 0 0 1.5rem;
    line-height: 1.6;
}
.social-links {
    display: flex;
    gap: 1rem;
}
.social-links a {
    color: rgba(255,255,255,0.7);
    transition: color 0.2s;
}
.social-links a:hover {
    color: {{.accent_color}};
}
.footer-links h4,
.footer-services h4,
.footer-contact h4 {
    color: #fff;
    margin: 0 0 1rem;
    font-size: 1rem;
    font-weight: 600;
}
.footer-links ul,
.footer-services ul {
    list-style: none;
    padding: 0;
    margin: 0;
}
.footer-links li,
.footer-services li {
    margin-bottom: 0.5rem;
}
.footer-links a,
.footer-services a,
.footer-contact a {
    color: rgba(255,255,255,0.7);
    text-decoration: none;
    transition: color 0.2s;
}
.footer-links a:hover,
.footer-services a:hover,
.footer-contact a:hover {
    color: #fff;
}
.footer-contact p {
    margin: 0 0 0.5rem;
    color: rgba(255,255,255,0.7);
}
.footer-bottom {
    margin-top: 3rem;
    padding: 1.5rem 0;
    border-top: 1px solid rgba(255,255,255,0.1);
}
.footer-bottom-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.footer-bottom p {
    margin: 0;
    color: rgba(255,255,255,0.5);
    font-size: 0.9rem;
}
.footer-legal {
    display: flex;
    gap: 2rem;
}
.footer-legal a {
    color: rgba(255,255,255,0.5);
    text-decoration: none;
    font-size: 0.9rem;
}
.footer-legal a:hover {
    color: rgba(255,255,255,0.8);
}
@media (max-width: 768px) {
    .footer-container {
        grid-template-columns: 1fr;
        gap: 2rem;
    }
    .footer-bottom-container {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
    }
}
</style>'
WHERE name = 'footer-4-column';

-- Also fix footer-simple which has Handlebars {{#if}}
UPDATE content_components
SET html_template = E'<!-- FOOTER SOURCE: component-db:footer-simple -->
<footer class="site-footer site-footer--simple">
    <div class="footer-container">
        <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>
        <nav class="footer-nav">
            <a href="/privacy.html">Privacy</a>
            <a href="/terms.html">Terms</a>
            {{if .email}}<a href="mailto:{{.email}}">Contact</a>{{end}}
        </nav>
    </div>
</footer>
<style>
.site-footer--simple {
    background: #f8f9fa;
    padding: 2rem 0;
    margin-top: auto;
    border-top: 1px solid #e9ecef;
}
.site-footer--simple .footer-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.site-footer--simple p {
    margin: 0;
    color: #6c757d;
    font-size: 0.9rem;
}
.footer-nav {
    display: flex;
    gap: 2rem;
}
.footer-nav a {
    color: #6c757d;
    text-decoration: none;
    font-size: 0.9rem;
}
.footer-nav a:hover {
    color: {{.primary_color}};
}
@media (max-width: 768px) {
    .site-footer--simple .footer-container {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
    }
}
</style>'
WHERE name = 'footer-simple';

-- Fix footer-standard placeholder inconsistencies (missing dots)
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(
                        REPLACE(html_template,
                                '{{nav_items_html}}', '{{.nav_items_html}}'
                        ),
                        '{{contact_email}}', '{{.email}}'
                ),
                '{{primary_color}}', '{{.primary_color}}'
        ),
        '{{accent_color}}', '{{.accent_color}}'
                    )
WHERE name = 'footer-standard';

-- Verify the updates
SELECT name,
       CASE WHEN html_template LIKE '%{{#each%' THEN 'HAS_HANDLEBARS_EACH' ELSE 'OK' END as each_status,
       CASE WHEN html_template LIKE '%{{#if%' THEN 'HAS_HANDLEBARS_IF' ELSE 'OK' END as if_status,
       LENGTH(html_template) as template_length
FROM content_components
WHERE name LIKE '%footer%';



--------- header fix

-- ============================================================================
-- FIX: Header templates - Convert Handlebars {{#if}} to Go template {{if}}
-- ============================================================================
--
-- PROBLEM:
--   header-professional-dark and header-minimal-light use:
--   {{#if logo_url}}...{{/if}} - Handlebars syntax
--
-- SOLUTION:
--   Convert to Go template syntax: {{if .logo_url}}...{{end}}
--
-- ============================================================================

-- First, let's see the full templates
SELECT name, html_template
FROM content_components
WHERE name IN ('header-professional-dark', 'header-minimal-light');

-- Fix header-professional-dark
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(
                        REPLACE(
                                REPLACE(html_template,
                                        '{{#if logo_url}}', '{{if .logo_url}}'
                                ),
                                '{{/if}}', '{{end}}'
                        ),
                        '{{logo_url}}', '{{.logo_url}}'
                ),
                '{{nav_items_html}}', '{{.nav_items_html}}'
        ),
        '{{cta_url}}', '{{.cta_url}}'
                    )
WHERE name = 'header-professional-dark';

-- Fix header-minimal-light
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(
                        REPLACE(
                                REPLACE(html_template,
                                        '{{#if logo_url}}', '{{if .logo_url}}'
                                ),
                                '{{/if}}', '{{end}}'
                        ),
                        '{{logo_url}}', '{{.logo_url}}'
                ),
                '{{nav_items_html}}', '{{.nav_items_html}}'
        ),
        '{{cta_url}}', '{{.cta_url}}'
                    )
WHERE name = 'header-minimal-light';

-- Fix header-bold-gradient (missing dots on placeholders)
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(html_template,
                        '{{nav_items_html}}', '{{.nav_items_html}}'
                ),
                '{{cta_url}}', '{{.cta_url}}'
        ),
        '{{cta_text}}', '{{.cta_text}}'
                    )
WHERE name = 'header-bold-gradient';

-- Verify the updates
SELECT name,
       CASE WHEN html_template LIKE '%{{#if%' THEN 'HAS_HANDLEBARS_IF' ELSE 'OK' END as if_status,
       CASE WHEN html_template LIKE '%{{nav_items_html}}%' THEN 'MISSING_DOT'
            WHEN html_template LIKE '%{{.nav_items_html}}%' THEN 'OK'
            ELSE 'NO_NAV' END as nav_placeholder_status
FROM content_components
WHERE name LIKE '%header%';

--

-- template fixes

-- ============================================================================
-- COMPREHENSIVE TEMPLATE FIX: Headers and Footers
-- ============================================================================
--
-- PROBLEMS FOUND:
-- 1. Headers have mixed Handlebars/Go syntax: {{#if cta_text}} vs {{if .logo_url}}
-- 2. Footers showing <no value> - variables not in context
-- 3. Some placeholders missing dots: {{nav_items_html}} vs {{.nav_items_html}}
--
-- ============================================================================

-- First, let's see current state of templates
SELECT name,
       CASE WHEN html_template LIKE '%{{#if%' THEN 'HAS_HANDLEBARS_IF' ELSE 'OK' END as handlebars_if,
       CASE WHEN html_template LIKE '%{{#each%' THEN 'HAS_HANDLEBARS_EACH' ELSE 'OK' END as handlebars_each,
       CASE WHEN html_template LIKE '%{{nav_items_html}}%' THEN 'MISSING_DOT' ELSE 'OK' END as missing_dot
FROM content_components
WHERE name LIKE '%header%' OR name LIKE '%footer%';

-- ============================================================================
-- FIX header-professional-dark - Full replacement
-- ============================================================================
UPDATE content_components
SET html_template = E'<!-- HEADER SOURCE: component-db:header-professional-dark -->
<header class="site-header site-header--dark">
    <div class="header-container">
        <a href="/index.html" class="logo">
            <span class="logo-text">{{.logo_text}}</span>
        </a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">
            <span></span><span></span><span></span>
        </button>
        <nav class="main-nav" id="main-nav" role="navigation">
            <ul>
                {{.nav_items_html}}
            </ul>
        </nav>
        <a href="/contact.html" class="header-cta">{{.cta_text}}</a>
    </div>
</header>
<style>
.site-header--dark {
    background: {{.primary_color}};
    padding: 1rem 0;
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}
.header-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 2rem;
}
.logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    text-decoration: none;
}
.logo-text {
    font-size: 1.25rem;
    font-weight: 700;
    color: #fff;
}
.main-nav ul {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 2rem;
}
.main-nav a {
    color: rgba(255,255,255,0.9);
    text-decoration: none;
    font-weight: 500;
    padding: 0.5rem 0;
    transition: color 0.2s;
}
.main-nav a:hover,
.main-nav a.active {
    color: {{.accent_color}};
}
.header-cta {
    background: {{.accent_color}};
    color: #fff;
    padding: 0.6rem 1.25rem;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 500;
    transition: opacity 0.2s;
}
.header-cta:hover {
    opacity: 0.9;
}
.mobile-menu-toggle {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    flex-direction: column;
    gap: 5px;
}
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: #fff;
    transition: transform 0.3s;
}
@media (max-width: 768px) {
    .mobile-menu-toggle { display: flex; }
    .main-nav {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: {{.primary_color}};
        padding: 1rem 2rem;
        display: none;
        box-shadow: 0 4px 10px rgba(0,0,0,0.1);
    }
    .main-nav.active { display: block; }
    .main-nav ul {
        flex-direction: column;
        gap: 0;
    }
    .main-nav a {
        display: block;
        padding: 0.75rem 0;
        border-bottom: 1px solid rgba(255,255,255,0.1);
    }
    .header-cta { display: none; }
}
</style>
<script>
document.addEventListener("DOMContentLoaded", function() {
    var toggle = document.querySelector(".mobile-menu-toggle");
    var nav = document.querySelector(".main-nav");
    if (toggle && nav) {
        toggle.addEventListener("click", function() {
            var expanded = toggle.getAttribute("aria-expanded") === "true";
            toggle.setAttribute("aria-expanded", !expanded);
            nav.classList.toggle("active");
        });
    }
});
</script>'
WHERE name = 'header-professional-dark';

-- ============================================================================
-- FIX header-minimal-light - Full replacement
-- ============================================================================
UPDATE content_components
SET html_template = E'<!-- HEADER SOURCE: component-db:header-minimal-light -->
<header class="site-header site-header--light">
    <div class="header-container">
        <a href="/index.html" class="logo">
            <span class="logo-text">{{.logo_text}}</span>
        </a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu">
            <span></span><span></span><span></span>
        </button>
        <nav class="main-nav" id="main-nav" role="navigation">
            <ul>
                {{.nav_items_html}}
            </ul>
        </nav>
    </div>
</header>
<style>
.site-header--light {
    background: #ffffff;
    padding: 1rem 0;
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 10px rgba(0,0,0,0.05);
}
.header-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.logo {
    text-decoration: none;
}
.logo-text {
    font-size: 1.25rem;
    font-weight: 700;
    color: {{.primary_color}};
}
.main-nav ul {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 2rem;
}
.main-nav a {
    color: #333;
    text-decoration: none;
    font-weight: 500;
    transition: color 0.2s;
}
.main-nav a:hover,
.main-nav a.active {
    color: {{.primary_color}};
}
.mobile-menu-toggle {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    flex-direction: column;
    gap: 5px;
}
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: #333;
}
@media (max-width: 768px) {
    .mobile-menu-toggle { display: flex; }
    .main-nav {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: #fff;
        padding: 1rem 2rem;
        display: none;
        box-shadow: 0 4px 10px rgba(0,0,0,0.1);
    }
    .main-nav.active { display: block; }
    .main-nav ul {
        flex-direction: column;
        gap: 0;
    }
    .main-nav a {
        display: block;
        padding: 0.75rem 0;
        border-bottom: 1px solid #eee;
    }
}
</style>
<script>
document.addEventListener("DOMContentLoaded", function() {
    var toggle = document.querySelector(".mobile-menu-toggle");
    var nav = document.querySelector(".main-nav");
    if (toggle && nav) {
        toggle.addEventListener("click", function() {
            nav.classList.toggle("active");
        });
    }
});
</script>'
WHERE name = 'header-minimal-light';

-- ============================================================================
-- FIX header-bold-gradient - Ensure dots on placeholders
-- ============================================================================
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(
                        REPLACE(html_template,
                                '{{nav_items_html}}', '{{.nav_items_html}}'
                        ),
                        '{{cta_url}}', '{{.cta_url}}'
                ),
                '{{cta_text}}', '{{.cta_text}}'
        ),
        '{{primary_color}}', '{{.primary_color}}'
                    )
WHERE name = 'header-bold-gradient';

-- ============================================================================
-- FIX footer-4-column - Full replacement with working template
-- ============================================================================
UPDATE content_components
SET html_template = E'<!-- FOOTER SOURCE: component-db:footer-4-column -->
<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand">
            <h3>{{.logo_text}}</h3>
            <p>{{.tagline}}</p>
        </div>
        <div class="footer-links">
            <h4>Quick Links</h4>
            <ul>
                {{.nav_items_html}}
            </ul>
        </div>
        <div class="footer-services">
            <h4>Our Services</h4>
            <ul>
                {{.nav_items_html}}
            </ul>
        </div>
        <div class="footer-contact">
            <h4>Contact</h4>
            <p><a href="mailto:{{.email}}">{{.email}}</a></p>
            <p>{{.phone}}</p>
        </div>
    </div>
    <div class="footer-bottom">
        <div class="footer-bottom-container">
            <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>
            <div class="footer-legal">
                <a href="/privacy.html">Privacy Policy</a>
                <a href="/terms.html">Terms of Service</a>
            </div>
        </div>
    </div>
</footer>
<style>
.site-footer {
    background: {{.primary_color}};
    color: rgba(255,255,255,0.9);
    padding: 4rem 0 0;
    margin-top: auto;
}
.footer-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: grid;
    grid-template-columns: 2fr 1fr 1fr 1fr;
    gap: 3rem;
}
.footer-brand h3 {
    color: #fff;
    margin: 0 0 0.75rem;
    font-size: 1.25rem;
}
.footer-brand p {
    color: rgba(255,255,255,0.7);
    margin: 0 0 1.5rem;
    line-height: 1.6;
}
.footer-links h4,
.footer-services h4,
.footer-contact h4 {
    color: #fff;
    margin: 0 0 1rem;
    font-size: 1rem;
    font-weight: 600;
}
.footer-links ul,
.footer-services ul {
    list-style: none;
    padding: 0;
    margin: 0;
}
.footer-links li,
.footer-services li {
    margin-bottom: 0.5rem;
}
.footer-links a,
.footer-services a,
.footer-contact a {
    color: rgba(255,255,255,0.7);
    text-decoration: none;
    transition: color 0.2s;
}
.footer-links a:hover,
.footer-services a:hover,
.footer-contact a:hover {
    color: #fff;
}
.footer-contact p {
    margin: 0 0 0.5rem;
    color: rgba(255,255,255,0.7);
}
.footer-bottom {
    margin-top: 3rem;
    padding: 1.5rem 0;
    border-top: 1px solid rgba(255,255,255,0.1);
}
.footer-bottom-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.footer-bottom p {
    margin: 0;
    color: rgba(255,255,255,0.5);
    font-size: 0.9rem;
}
.footer-legal {
    display: flex;
    gap: 2rem;
}
.footer-legal a {
    color: rgba(255,255,255,0.5);
    text-decoration: none;
    font-size: 0.9rem;
}
.footer-legal a:hover {
    color: rgba(255,255,255,0.8);
}
@media (max-width: 768px) {
    .footer-container {
        grid-template-columns: 1fr;
        gap: 2rem;
    }
    .footer-bottom-container {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
    }
}
</style>'
WHERE name = 'footer-4-column';

-- ============================================================================
-- VERIFY: Check all templates are clean
-- ============================================================================
SELECT name,
       CASE WHEN html_template LIKE '%{{#%' THEN 'STILL_HAS_HANDLEBARS' ELSE 'OK' END as handlebars_check,
       CASE WHEN html_template LIKE '%{{nav_items_html}}%' AND html_template NOT LIKE '%{{.nav_items_html}}%'
                THEN 'MISSING_DOT' ELSE 'OK' END as dot_check,
       LENGTH(html_template) as length
FROM content_components
WHERE name LIKE '%header%' OR name LIKE '%footer%';

-- ============================================================================
-- COMPREHENSIVE TEMPLATE FIX: Headers and Footers
-- ============================================================================
--
-- PROBLEMS FOUND:
-- 1. Headers have mixed Handlebars/Go syntax: {{#if cta_text}} vs {{if .logo_url}}
-- 2. Footers showing <no value> - variables not in context
-- 3. Some placeholders missing dots: {{nav_items_html}} vs {{.nav_items_html}}
--
-- ============================================================================

-- First, let's see current state of templates
SELECT name,
       CASE WHEN html_template LIKE '%{{#if%' THEN 'HAS_HANDLEBARS_IF' ELSE 'OK' END as handlebars_if,
       CASE WHEN html_template LIKE '%{{#each%' THEN 'HAS_HANDLEBARS_EACH' ELSE 'OK' END as handlebars_each,
       CASE WHEN html_template LIKE '%{{nav_items_html}}%' THEN 'MISSING_DOT' ELSE 'OK' END as missing_dot
FROM content_components
WHERE name LIKE '%header%' OR name LIKE '%footer%';

-- ============================================================================
-- FIX header-professional-dark - Full replacement
-- ============================================================================
UPDATE content_components
SET html_template = E'<!-- HEADER SOURCE: component-db:header-professional-dark -->
<header class="site-header site-header--dark">
    <div class="header-container">
        <a href="/index.html" class="logo">
            <span class="logo-text">{{.logo_text}}</span>
        </a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">
            <span></span><span></span><span></span>
        </button>
        <nav class="main-nav" id="main-nav" role="navigation">
            <ul>
                {{.nav_items_html}}
            </ul>
        </nav>
        <a href="/contact.html" class="header-cta">{{.cta_text}}</a>
    </div>
</header>
<style>
.site-header--dark {
    background: {{.primary_color}};
    padding: 1rem 0;
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}
.header-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 2rem;
}
.logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    text-decoration: none;
}
.logo-text {
    font-size: 1.25rem;
    font-weight: 700;
    color: #fff;
}
.main-nav ul {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 2rem;
}
.main-nav a {
    color: rgba(255,255,255,0.9);
    text-decoration: none;
    font-weight: 500;
    padding: 0.5rem 0;
    transition: color 0.2s;
}
.main-nav a:hover,
.main-nav a.active {
    color: {{.accent_color}};
}
.header-cta {
    background: {{.accent_color}};
    color: #fff;
    padding: 0.6rem 1.25rem;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 500;
    transition: opacity 0.2s;
}
.header-cta:hover {
    opacity: 0.9;
}
.mobile-menu-toggle {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    flex-direction: column;
    gap: 5px;
}
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: #fff;
    transition: transform 0.3s;
}
@media (max-width: 768px) {
    .mobile-menu-toggle { display: flex; }
    .main-nav {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: {{.primary_color}};
        padding: 1rem 2rem;
        display: none;
        box-shadow: 0 4px 10px rgba(0,0,0,0.1);
    }
    .main-nav.active { display: block; }
    .main-nav ul {
        flex-direction: column;
        gap: 0;
    }
    .main-nav a {
        display: block;
        padding: 0.75rem 0;
        border-bottom: 1px solid rgba(255,255,255,0.1);
    }
    .header-cta { display: none; }
}
</style>
<script>
document.addEventListener("DOMContentLoaded", function() {
    var toggle = document.querySelector(".mobile-menu-toggle");
    var nav = document.querySelector(".main-nav");
    if (toggle && nav) {
        toggle.addEventListener("click", function() {
            var expanded = toggle.getAttribute("aria-expanded") === "true";
            toggle.setAttribute("aria-expanded", !expanded);
            nav.classList.toggle("active");
        });
    }
});
</script>'
WHERE name = 'header-professional-dark';

-- ============================================================================
-- FIX header-minimal-light - Full replacement
-- ============================================================================
UPDATE content_components
SET html_template = E'<!-- HEADER SOURCE: component-db:header-minimal-light -->
<header class="site-header site-header--light">
    <div class="header-container">
        <a href="/index.html" class="logo">
            <span class="logo-text">{{.logo_text}}</span>
        </a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu">
            <span></span><span></span><span></span>
        </button>
        <nav class="main-nav" id="main-nav" role="navigation">
            <ul>
                {{.nav_items_html}}
            </ul>
        </nav>
    </div>
</header>
<style>
.site-header--light {
    background: #ffffff;
    padding: 1rem 0;
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 10px rgba(0,0,0,0.05);
}
.header-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.logo {
    text-decoration: none;
}
.logo-text {
    font-size: 1.25rem;
    font-weight: 700;
    color: {{.primary_color}};
}
.main-nav ul {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 2rem;
}
.main-nav a {
    color: #333;
    text-decoration: none;
    font-weight: 500;
    transition: color 0.2s;
}
.main-nav a:hover,
.main-nav a.active {
    color: {{.primary_color}};
}
.mobile-menu-toggle {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    flex-direction: column;
    gap: 5px;
}
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: #333;
}
@media (max-width: 768px) {
    .mobile-menu-toggle { display: flex; }
    .main-nav {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: #fff;
        padding: 1rem 2rem;
        display: none;
        box-shadow: 0 4px 10px rgba(0,0,0,0.1);
    }
    .main-nav.active { display: block; }
    .main-nav ul {
        flex-direction: column;
        gap: 0;
    }
    .main-nav a {
        display: block;
        padding: 0.75rem 0;
        border-bottom: 1px solid #eee;
    }
}
</style>
<script>
document.addEventListener("DOMContentLoaded", function() {
    var toggle = document.querySelector(".mobile-menu-toggle");
    var nav = document.querySelector(".main-nav");
    if (toggle && nav) {
        toggle.addEventListener("click", function() {
            nav.classList.toggle("active");
        });
    }
});
</script>'
WHERE name = 'header-minimal-light';

-- ============================================================================
-- FIX header-bold-gradient - Ensure dots on placeholders
-- ============================================================================
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(
                        REPLACE(html_template,
                                '{{nav_items_html}}', '{{.nav_items_html}}'
                        ),
                        '{{cta_url}}', '{{.cta_url}}'
                ),
                '{{cta_text}}', '{{.cta_text}}'
        ),
        '{{primary_color}}', '{{.primary_color}}'
                    )
WHERE name = 'header-bold-gradient';

-- ============================================================================
-- FIX footer-4-column - Full replacement with working template
-- ============================================================================
UPDATE content_components
SET html_template = E'<!-- FOOTER SOURCE: component-db:footer-4-column -->
<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand">
            <h3>{{.logo_text}}</h3>
            <p>{{.tagline}}</p>
        </div>
        <div class="footer-links">
            <h4>Quick Links</h4>
            <ul>
                {{.nav_items_html}}
            </ul>
        </div>
        <div class="footer-services">
            <h4>Our Services</h4>
            <ul>
                {{.nav_items_html}}
            </ul>
        </div>
        <div class="footer-contact">
            <h4>Contact</h4>
            <p><a href="mailto:{{.email}}">{{.email}}</a></p>
            <p>{{.phone}}</p>
        </div>
    </div>
    <div class="footer-bottom">
        <div class="footer-bottom-container">
            <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>
            <div class="footer-legal">
                <a href="/privacy.html">Privacy Policy</a>
                <a href="/terms.html">Terms of Service</a>
            </div>
        </div>
    </div>
</footer>
<style>
.site-footer {
    background: {{.primary_color}};
    color: rgba(255,255,255,0.9);
    padding: 4rem 0 0;
    margin-top: auto;
}
.footer-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: grid;
    grid-template-columns: 2fr 1fr 1fr 1fr;
    gap: 3rem;
}
.footer-brand h3 {
    color: #fff;
    margin: 0 0 0.75rem;
    font-size: 1.25rem;
}
.footer-brand p {
    color: rgba(255,255,255,0.7);
    margin: 0 0 1.5rem;
    line-height: 1.6;
}
.footer-links h4,
.footer-services h4,
.footer-contact h4 {
    color: #fff;
    margin: 0 0 1rem;
    font-size: 1rem;
    font-weight: 600;
}
.footer-links ul,
.footer-services ul {
    list-style: none;
    padding: 0;
    margin: 0;
}
.footer-links li,
.footer-services li {
    margin-bottom: 0.5rem;
}
.footer-links a,
.footer-services a,
.footer-contact a {
    color: rgba(255,255,255,0.7);
    text-decoration: none;
    transition: color 0.2s;
}
.footer-links a:hover,
.footer-services a:hover,
.footer-contact a:hover {
    color: #fff;
}
.footer-contact p {
    margin: 0 0 0.5rem;
    color: rgba(255,255,255,0.7);
}
.footer-bottom {
    margin-top: 3rem;
    padding: 1.5rem 0;
    border-top: 1px solid rgba(255,255,255,0.1);
}
.footer-bottom-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.footer-bottom p {
    margin: 0;
    color: rgba(255,255,255,0.5);
    font-size: 0.9rem;
}
.footer-legal {
    display: flex;
    gap: 2rem;
}
.footer-legal a {
    color: rgba(255,255,255,0.5);
    text-decoration: none;
    font-size: 0.9rem;
}
.footer-legal a:hover {
    color: rgba(255,255,255,0.8);
}
@media (max-width: 768px) {
    .footer-container {
        grid-template-columns: 1fr;
        gap: 2rem;
    }
    .footer-bottom-container {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
    }
}
</style>'
WHERE name = 'footer-4-column';

-- ============================================================================
-- VERIFY: Check all templates are clean
-- ============================================================================
SELECT name,
       CASE WHEN html_template LIKE '%{{#%' THEN 'STILL_HAS_HANDLEBARS' ELSE 'OK' END as handlebars_check,
       CASE WHEN html_template LIKE '%{{nav_items_html}}%' AND html_template NOT LIKE '%{{.nav_items_html}}%'
                THEN 'MISSING_DOT' ELSE 'OK' END as dot_check,
       LENGTH(html_template) as length
FROM content_components
WHERE name LIKE '%header%' OR name LIKE '%footer%';


-- template paths
-- ============================================================================
-- Fix footer and header templates to use correct variable names
-- ============================================================================
-- The Go code will generate:
--   nav_items_html  - For navigation links
--   services_html   - For services list
--
-- Both are pre-rendered HTML strings like:
--   <li><a href="/page.html">Label</a></li>
-- ============================================================================

-- First, check current state
SELECT name,
       CASE WHEN html_template LIKE '%{{.nav_items_html}}%' THEN 'has_nav_items_html'
            WHEN html_template LIKE '%{{nav_items_html}}%' THEN 'missing_dot'
            ELSE 'not_found' END as nav_check,
       CASE WHEN html_template LIKE '%{{.services_html}}%' THEN 'has_services_html'
            WHEN html_template LIKE '%{{services_html}}%' THEN 'missing_dot'
            ELSE 'not_found' END as services_check
FROM content_components
WHERE name LIKE '%header%' OR name LIKE '%footer%';

-- Fix header-professional-dark: wrap nav_items_html in if block
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{.nav_items_html}}',
        '{{if .nav_items_html}}{{.nav_items_html}}{{end}}'
                    )
WHERE name = 'header-professional-dark'
  AND html_template LIKE '%{{.nav_items_html}}%'
  AND html_template NOT LIKE '%{{if .nav_items_html}}%';

-- Fix header-minimal-light
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{.nav_items_html}}',
        '{{if .nav_items_html}}{{.nav_items_html}}{{end}}'
                    )
WHERE name = 'header-minimal-light'
  AND html_template LIKE '%{{.nav_items_html}}%'
  AND html_template NOT LIKE '%{{if .nav_items_html}}%';

-- Fix header-bold-gradient
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{.nav_items_html}}',
        '{{if .nav_items_html}}{{.nav_items_html}}{{end}}'
                    )
WHERE name = 'header-bold-gradient'
  AND html_template LIKE '%{{.nav_items_html}}%'
  AND html_template NOT LIKE '%{{if .nav_items_html}}%';

-- Fix footer-4-column nav_items_html
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{.nav_items_html}}',
        '{{if .nav_items_html}}{{.nav_items_html}}{{end}}'
                    )
WHERE name = 'footer-4-column'
  AND html_template LIKE '%{{.nav_items_html}}%'
  AND html_template NOT LIKE '%{{if .nav_items_html}}%';

-- Fix footer-4-column services - need to add services_html variable
-- First check what's in the services section
SELECT name,
       substring(html_template from 'footer-services.{0,300}') as services_section_preview
FROM content_components
WHERE name = 'footer-4-column';

-- The footer services section might be using nav_items_html incorrectly
-- or might have some other placeholder. Let's see the actual content first.

-- After the Go fix is applied, the templates will work with:
-- {{if .nav_items_html}}{{.nav_items_html}}{{end}} - renders nav links or nothing
-- {{if .services_html}}{{.services_html}}{{end}} - renders services or nothing

---

small fixes

-- ============================================================================
-- FIX 1: Footer template - use services_html for "Our Services" section
-- ============================================================================
-- Currently both Quick Links AND Our Services use nav_items_html
-- We need the services section to use services_html instead

-- First, check current footer template structure
SELECT name,
       substring(html_template from 'footer-links.{0,150}') as links_section,
       substring(html_template from 'footer-services.{0,150}') as services_section
FROM content_components
WHERE name = 'footer-4-column';

-- The footer-4-column template likely has both sections using {{.nav_items_html}}
-- We need to change the services section to use {{.services_html}}

-- Update footer-4-column: Change services section to use services_html
-- This is tricky because we need to target only the second occurrence
-- Let's do it by matching the specific div class

UPDATE content_components
SET html_template = regexp_replace(
        html_template,
        '(<div class="footer-services">[\s\S]*?<ul>)\s*\{\{\.nav_items_html\}\}\s*(</ul>)',
        E'\\1\n                {{if .services_html}}{{.services_html}}{{else}}{{.nav_items_html}}{{end}}\n            \\2',
        'g'
                    )
WHERE name = 'footer-4-column';

-- Verify the change
SELECT name,
       CASE WHEN html_template LIKE '%footer-services%services_html%' THEN 'FIXED' ELSE 'NOT_FIXED' END as services_section_status
FROM content_components
WHERE name = 'footer-4-column';


-- ============================================================================
-- FIX 2: Fix nav_labels in pages table
-- ============================================================================
-- Current labels are too long: "Insights & Blog | Leopardess Consulting"
-- Should be short: "Insights", "Careers", etc.

-- First, see current state
SELECT name, nav_label, title, url
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
ORDER BY nav_order;

-- Fix the nav_labels to be short and clean
UPDATE pages SET nav_label = 'Home' WHERE name = 'index'
                                      AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'About' WHERE name = 'about'
                                       AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'Services' WHERE name = 'services'
                                          AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'Use Cases' WHERE name = 'use-cases'
                                           AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'Case Studies' WHERE name = 'case-studies'
                                              AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'Contact' WHERE name = 'contact'
                                         AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'Insights' WHERE name = 'insights'
                                          AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'Careers' WHERE name = 'careers'
                                         AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

-- Verify the changes
SELECT name, nav_label, nav_order
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
ORDER BY nav_order;

-- Also fix duplicate "Use Cases" - case-studies should have different label
-- Already done above with 'Case Studies'


-- ============================================================================
-- FIX 3: Diagnose why services aren't being extracted from reviewed_brief
-- ============================================================================

-- Check if services exist in the reviewed_brief
SELECT
    orchestration_id,
    -- Check if services exist at different paths
    collected_data->'reviewed_brief'->'services' IS NOT NULL as services_at_top,
    collected_data->'reviewed_brief'->'response'->'services' IS NOT NULL as services_in_response,
    -- Get the actual services if present
    jsonb_array_length(COALESCE(
            collected_data->'reviewed_brief'->'services',
            collected_data->'reviewed_brief'->'response'->'services',
            '[]'::jsonb
                       )) as services_count,
    -- Preview the services
    jsonb_pretty(COALESCE(
            collected_data->'reviewed_brief'->'services',
            collected_data->'reviewed_brief'->'response'->'services'
                 )) as services_preview
FROM orchestration_states
WHERE correlation_id = '0663d7af-4aa8-499b-95b7-d2c212bdf99f'
  AND orchestration_name LIKE '%pageflow%'
ORDER BY created_at DESC
    LIMIT 1;

-- Check the sites table content_data for services
SELECT
    domain,
    content_data->'services' IS NOT NULL as has_services,
    jsonb_array_length(COALESCE(content_data->'services', '[]'::jsonb)) as services_count,
    jsonb_pretty(content_data->'services') as services
FROM sites
WHERE domain = 'leopardessconsulting.co.uk';

-- Check reviewed_brief structure in intake-orchestrator
SELECT
    orchestration_id,
    orchestration_name,
    jsonb_typeof(collected_data->'reviewed_brief') as reviewed_brief_type,
    collected_data->'reviewed_brief' ? 'response' as has_response_wrapper,
    collected_data->'reviewed_brief' ? 'services' as has_services_direct,
    collected_data->'reviewed_brief'->'response' ? 'services' as has_services_in_response
FROM orchestration_states
WHERE correlation_id = '0663d7af-4aa8-499b-95b7-d2c212bdf99f'
  AND collected_data ? 'reviewed_brief'
ORDER BY created_at
    LIMIT 3;


-- ============================================================================
-- FIX 4: Add cta_text to render context (for header CTA button)
-- ============================================================================
-- The header CTA is empty: <a href="/contact.html" class="header-cta"></a>
-- Need to check if cta_text is in reviewed_brief

SELECT
    collected_data->'reviewed_brief'->'response'->>'cta_text' as cta_text_in_response,
    collected_data->'reviewed_brief'->>'cta_text' as cta_text_direct
FROM orchestration_states
WHERE correlation_id = '0663d7af-4aa8-499b-95b7-d2c212bdf99f'
  AND collected_data ? 'reviewed_brief'
    LIMIT 1;

-- Check sites.content_data for cta_text
SELECT
    content_data->>'cta_text' as cta_text,
    content_data->>'cta_url' as cta_url
FROM sites
WHERE domain = 'leopardessconsulting.co.uk';


-- ============================================================================
-- FIX 5: Ensure header template uses cta_text properly
-- ============================================================================
-- Check what the header template expects for CTA

SELECT name,
       substring(html_template from 'header-cta.{0,100}') as cta_section
FROM content_components
WHERE name = 'header-professional-dark';

-- The CTA button should show text like "Get Started" or "Contact Us"
-- Update header to use cta_text with fallback
UPDATE content_components
SET html_template = replace(
        html_template,
        '<a href="/contact.html" class="header-cta"></a>',
        '<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>'
                    )
WHERE name = 'header-professional-dark'
  AND html_template LIKE '%<a href="/contact.html" class="header-cta"></a>%';

-- Also fix other header variants
UPDATE content_components
SET html_template = replace(
        html_template,
        '<a href="/contact.html" class="header-cta"></a>',
        '<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>'
                    )
WHERE name IN ('header-minimal-light', 'header-bold-gradient')
  AND html_template LIKE '%<a href="/contact.html" class="header-cta"></a>%';


-- ============================================================================
-- VERIFICATION: Check all fixes applied
-- ============================================================================
SELECT name,
       CASE WHEN html_template LIKE '%services_html%' THEN 'YES' ELSE 'NO' END as has_services_html,
       CASE WHEN html_template LIKE '%cta_text%' THEN 'YES' ELSE 'NO' END as has_cta_text,
       CASE WHEN html_template LIKE '%Get Started%' THEN 'YES' ELSE 'NO' END as has_cta_fallback
FROM content_components
WHERE name LIKE '%header%' OR name LIKE '%footer%';

UPDATE content_components
SET html_template = replace(
        html_template,
        'footer-services">
                <h4>Our Services</h4>
                <ul>
                    {{if .nav_items_html}}{{.nav_items_html}}{{end}}',
        'footer-services">
                <h4>Our Services</h4>
                <ul>
                    {{if .services_html}}{{.services_html}}{{end}}'
                    )
WHERE name = 'footer-4-column';

--

-- services.html and footer and header updates
-- ============================================================================
-- Clean fixes for leopardessconsulting.co.uk
-- Run each statement separately to avoid parsing issues
-- ============================================================================

-- 1. Add cta_text and cta_url to sites.content_data
UPDATE sites
SET content_data = COALESCE(content_data, '{}'::jsonb) ||
                   '{"cta_text": "Get Started", "cta_url": "/contact.html"}'::jsonb
WHERE domain = 'leopardessconsulting.co.uk';

-- 2. Add services to sites.content_data
UPDATE sites
SET content_data = content_data ||
                   '{"services": ["Automated Website Solutions", "Multi-Agent Systems", "Digital Transformation Strategy", "Web Consultancy"]}'::jsonb
WHERE domain = 'leopardessconsulting.co.uk';

-- 3. Verify the content_data update
SELECT
    content_data->>'cta_text' as cta_text,
    content_data->>'cta_url' as cta_url,
    jsonb_array_length(content_data->'services') as services_count
FROM sites
WHERE domain = 'leopardessconsulting.co.uk';

-- 4. Fix remaining nav_labels (privacy and terms)
UPDATE pages SET nav_label = 'Privacy'
WHERE name = 'privacy'
  AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

UPDATE pages SET nav_label = 'Terms'
WHERE name = 'terms'
  AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

-- 5. Verify nav_labels
SELECT name, nav_label, nav_order
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
ORDER BY nav_order;

-- 6. Add CTA fallback to header templates
UPDATE content_components
SET html_template = replace(
        html_template,
        '>{{.cta_text}}</a>',
        '>{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>'
                    )
WHERE name = 'header-professional-dark'
  AND html_template LIKE '%>{{.cta_text}}</a>%';

UPDATE content_components
SET html_template = replace(
        html_template,
        '>{{.cta_text}}</a>',
        '>{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>'
                    )
WHERE name = 'header-minimal-light'
  AND html_template LIKE '%>{{.cta_text}}</a>%';

UPDATE content_components
SET html_template = replace(
        html_template,
        '>{{.cta_text}}</a>',
        '>{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>'
                    )
WHERE name = 'header-bold-gradient'
  AND html_template LIKE '%>{{.cta_text}}</a>%';

-- 7. Verify header CTA fallbacks
SELECT name,
       CASE WHEN html_template LIKE '%Get Started%' THEN 'HAS_FALLBACK' ELSE 'NO_FALLBACK' END as cta_status
FROM content_components
WHERE name LIKE 'header-%';

-- 8. Verify footer services_html is in place (from previous update)
SELECT name,
       CASE WHEN html_template LIKE '%services_html%' THEN 'YES' ELSE 'NO' END as has_services_html
FROM content_components
WHERE name = 'footer-4-column';

-- add image to hero

UPDATE content_components
SET html_template = '<section class="hero" data-component="hero"{{if .hero_home_url}} style="background: linear-gradient(135deg, rgba(26,26,46,0.8) 0%, rgba(22,33,62,0.75) 50%, rgba(15,52,96,0.7) 100%), url(''{{.hero_home_url}}'') center/cover no-repeat;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .primary_cta_text}}<a href="{{.primary_cta_url | default "/contact.html"}}" class="btn btn-primary">{{.primary_cta_text}}</a>{{end}}
            {{if .secondary_cta_text}}<a href="{{.secondary_cta_url | default "/services.html"}}" class="btn btn-secondary">{{.secondary_cta_text}}</a>{{end}}
        </div>
    </section>',
    updated_at = NOW()
WHERE function = 'hero' AND (category IS NULL OR category = '');

-- Verify
SELECT name, function,
       CASE WHEN html_template LIKE '%hero_home_url%' THEN 'YES' ELSE 'NO' END as has_image_support
FROM content_components WHERE function = 'hero';

-- adding images to hero

-- ============================================================
-- PART 2: Update hero template to use background image URL
-- ============================================================

-- Update the main hero component (home page)
UPDATE content_components
SET html_template = '<section class="hero" data-component="hero"{{if .hero_url}} style="background: linear-gradient(135deg, rgba(26,26,46,0.8) 0%, rgba(22,33,62,0.75) 50%, rgba(15,52,96,0.7) 100%), url(''{{.hero_url}}'') center/cover no-repeat;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .primary_cta_text}}<a href="{{.primary_cta_url | default "/contact.html"}}" class="btn btn-primary">{{.primary_cta_text}}</a>{{end}}
            {{if .secondary_cta_text}}<a href="{{.secondary_cta_url | default "/services.html"}}" class="btn btn-secondary">{{.secondary_cta_text}}</a>{{end}}
        </div>
    </section>',
    updated_at = NOW()
WHERE function = 'hero' AND (category IS NULL OR category = '');

-- Verify the workflow flow
SELECT
    type,
    default_config->'workflow'->'steps'->'store_hero_asset'->'next_step' as "1_store_next",
    default_config->'workflow'->'steps'->'deploy_hero_image'->'next_step' as "2_deploy_next"
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Expected output:
-- 1_store_next: "deploy_hero_image"
-- 2_deploy_next: "select_style_collection"

-- Verify hero template
SELECT name, function,
       CASE WHEN html_template LIKE '%hero_url%' THEN 'YES' ELSE 'NO' END as has_image_support
FROM content_components WHERE function = 'hero';

--

UPDATE content_components
SET html_template = regexp_replace(
        html_template,
        'style="background: linear-gradient\([^"]+\)"',
        'style="background: {{if .hero_url}}linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.5)), url(''{{.hero_url}}'') center/cover no-repeat{{else}}linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%){{end}}"'
                    )
WHERE name = 'hero' AND category = 'hero';

-- Verify
SELECT name,
       CASE WHEN html_template LIKE '%hero_url%' THEN 'YES' ELSE 'NO' END as has_hero_url,
       substring(html_template, 1, 500) as template_start
FROM content_components
WHERE name = 'hero' AND category = 'hero';

-- verification
name | has_hero_url |                                                                                                                                                                                          template_start
------+--------------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 hero | NO           | <section class="hero" data-component="hero" style="{{if .background_image}}background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('{{.background_image}}'); background-size: cover; background-position: center;{{else}}background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%);{{end}}">+
      |              |         <div class="hero-content">                                                                                                                                                                                                                                                                                                                                                               +
      |              |             <h1>{{.headline}}</h1>                                                                                                                                                                                                                                                                                                                                                               +
      |              |             <p class="hero-subheadline">{{.s
(1 row)

-- fixing images hero
     -- Fix the template to support hero_url
UPDATE content_components
SET html_template = replace(
        replace(
                html_template,
                '{{if .background_image}}',
                '{{if or .hero_url .background_image}}'
        ),
        E'url(\'{{.background_image}}\')',
    E'url(\'{{or .hero_url .background_image}}\')'
)
WHERE name = 'hero' AND category = 'hero';

-- Verify
SELECT substring(html_template, 1, 600) FROM content_components WHERE name = 'hero' AND category = 'hero';

-- Update hero template to check hero_url as well as background_image
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{if .background_image}}',
        '{{if or .hero_url .background_image}}'
                    )
WHERE name = 'hero' AND category = 'hero';

-- Then update the url reference
UPDATE content_components
SET html_template = replace(
        html_template,
        E'url(\'{{.background_image}}\')',
    E'url(\'{{or .hero_url .background_image}}\')'
)
WHERE name = 'hero' AND category = 'hero';

-- Verify
SELECT substring(html_template, 1, 400) FROM content_components WHERE name = 'hero';


--

-- css to hero

-- Add CSS to hero components
-- This adds inline <style> blocks to match the header/footer pattern

-- Update the main "hero" component (id: 23f95f00-f293-466e-b43a-81791ea0fc6c)
UPDATE content_components
SET html_template = '<section class="hero" data-component="hero" style="{{if or .hero_url .background_image}}background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url(''{{or .hero_url .background_image}}''); background-size: cover; background-position: center;{{else}}background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%);{{end}}">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .cta_text}}<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
            {{if .secondary_cta}}<a href="{{if .secondary_cta_url}}{{.secondary_cta_url}}{{else}}#features{{end}}" class="btn btn-secondary">{{.secondary_cta}}</a>{{end}}
        </div>
    </section>
<style>
.hero {
    min-height: 70vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    position: relative;
}
.hero-content {
    max-width: 900px;
    margin: 0 auto;
    color: #fff;
    z-index: 1;
}
.hero h1 {
    font-size: clamp(2rem, 5vw, 3.5rem);
    font-weight: 700;
    margin-bottom: 1.5rem;
    line-height: 1.2;
    color: #fff;
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.35rem);
    margin-bottom: 2rem;
    opacity: 0.95;
    line-height: 1.6;
    color: rgba(255,255,255,0.95);
}
.hero .btn {
    display: inline-block;
    padding: 0.875rem 2rem;
    margin: 0.5rem;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 600;
    font-size: 1rem;
    transition: all 0.2s ease;
}
.hero .btn-primary {
    background: var(--accent-color, #0f3460);
    color: #fff;
    border: 2px solid var(--accent-color, #0f3460);
}
.hero .btn-primary:hover {
    background: transparent;
    color: #fff;
}
.hero .btn-secondary {
    background: transparent;
    color: #fff;
    border: 2px solid rgba(255,255,255,0.8);
}
.hero .btn-secondary:hover {
    background: rgba(255,255,255,0.1);
}
@media (max-width: 768px) {
    .hero {
        min-height: 60vh;
        padding: 3rem 1.5rem;
    }
    .hero .btn {
        display: block;
        width: 100%;
        max-width: 280px;
        margin: 0.5rem auto;
    }
}
</style>',
    updated_at = NOW()
WHERE name = 'hero';

-- Update "Hero Section" (id: ad64fada-3e73-493d-b906-bf32517031f0)
UPDATE content_components
SET html_template = '<section class="hero" data-component="hero"{{if .hero_url}} style="background: linear-gradient(135deg, rgba(26,26,46,0.8) 0%, rgba(22,33,62,0.75) 50%, rgba(15,52,96,0.7) 100%), url(''{{.hero_url}}'') center/cover no-repeat;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .primary_cta_text}}<a href="{{.primary_cta_url | default "/contact.html"}}" class="btn btn-primary">{{.primary_cta_text}}</a>{{end}}
            {{if .secondary_cta_text}}<a href="{{.secondary_cta_url | default "/services.html"}}" class="btn btn-secondary">{{.secondary_cta_text}}</a>{{end}}
        </div>
    </section>
<style>
.hero {
    min-height: 70vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    position: relative;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.hero-content {
    max-width: 900px;
    margin: 0 auto;
    color: #fff;
    z-index: 1;
}
.hero h1 {
    font-size: clamp(2rem, 5vw, 3.5rem);
    font-weight: 700;
    margin-bottom: 1.5rem;
    line-height: 1.2;
    color: #fff;
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.35rem);
    margin-bottom: 2rem;
    opacity: 0.95;
    line-height: 1.6;
    color: rgba(255,255,255,0.95);
}
.hero .btn {
    display: inline-block;
    padding: 0.875rem 2rem;
    margin: 0.5rem;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 600;
    font-size: 1rem;
    transition: all 0.2s ease;
}
.hero .btn-primary {
    background: var(--accent-color, #0f3460);
    color: #fff;
    border: 2px solid var(--accent-color, #0f3460);
}
.hero .btn-primary:hover {
    background: transparent;
    color: #fff;
}
.hero .btn-secondary {
    background: transparent;
    color: #fff;
    border: 2px solid rgba(255,255,255,0.8);
}
.hero .btn-secondary:hover {
    background: rgba(255,255,255,0.1);
}
@media (max-width: 768px) {
    .hero {
        min-height: 60vh;
        padding: 3rem 1.5rem;
    }
    .hero .btn {
        display: block;
        width: 100%;
        max-width: 280px;
        margin: 0.5rem auto;
    }
}
</style>',
    updated_at = NOW()
WHERE name = 'Hero Section';

-- Update page-specific hero variants (about, services, contact, case-studies)
-- These don't have background images, just gradient backgrounds

UPDATE content_components
SET html_template = '<section class="hero hero-about" data-component="about-hero">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-about {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.hero-about .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: #fff;
}
.hero-about h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
    color: #fff;
}
.hero-about .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    opacity: 0.9;
    line-height: 1.6;
    color: rgba(255,255,255,0.9);
}
</style>',
    updated_at = NOW()
WHERE name = 'about-hero';

UPDATE content_components
SET html_template = '<section class="hero hero-services" data-component="services-hero">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-services {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.hero-services .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: #fff;
}
.hero-services h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
    color: #fff;
}
.hero-services .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    opacity: 0.9;
    line-height: 1.6;
    color: rgba(255,255,255,0.9);
}
</style>',
    updated_at = NOW()
WHERE name = 'services-hero';

UPDATE content_components
SET html_template = '<section class="hero hero-contact" data-component="contact-hero">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-contact {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.hero-contact .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: #fff;
}
.hero-contact h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
    color: #fff;
}
.hero-contact .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    opacity: 0.9;
    line-height: 1.6;
    color: rgba(255,255,255,0.9);
}
</style>',
    updated_at = NOW()
WHERE name = 'contact-hero';

UPDATE content_components
SET html_template = '<section class="hero hero-case-studies" data-component="case-studies-hero">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-case-studies {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.hero-case-studies .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: #fff;
}
.hero-case-studies h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
    color: #fff;
}
.hero-case-studies .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    opacity: 0.9;
    line-height: 1.6;
    color: rgba(255,255,255,0.9);
}
</style>',
    updated_at = NOW()
WHERE name = 'case-studies-hero';

-- Verify updates
SELECT name,
       CASE WHEN html_template LIKE '%<style>%' THEN 'Has CSS' ELSE 'No CSS' END as has_css,
       updated_at
FROM content_components
WHERE function = 'hero';


