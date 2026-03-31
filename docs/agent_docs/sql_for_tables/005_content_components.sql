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

---

-- Update head components to link the external stylesheet
-- The webdesign-agent deploys CSS to /assets/css/styles.css

-- Update head-seo-standard component
UPDATE content_components
SET html_template = REPLACE(
        html_template,
        '</head>',
        '    <link rel="stylesheet" href="/assets/css/styles.css">
    </head>'
                    )
WHERE function_name = 'head-seo-standard'
  AND html_template NOT LIKE '%/assets/css/styles.css%';

-- Update "Document Head" component (simpler version)
UPDATE content_components
SET html_template = REPLACE(
        html_template,
        '</head>',
        '    <link rel="stylesheet" href="/assets/css/styles.css">
    </head>'
                    )
WHERE name = 'Document Head'
  AND html_template NOT LIKE '%/assets/css/styles.css%';

-- Verify the updates
SELECT name, function_name,
       CASE WHEN html_template LIKE '%/assets/css/styles.css%'
                THEN 'Has stylesheet link'
            ELSE 'Missing stylesheet link'
           END as stylesheet_status
FROM content_components
WHERE function_name IN ('head-seo-standard', 'head')
   OR name = 'Document Head';


---

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

--

-- head

UPDATE content_components
SET html_template = REPLACE(
        html_template,
        '</head>',
        '    <link rel="stylesheet" href="/assets/css/styles.css">
    </head>'
                    )
WHERE name = 'head-seo-standard'
  AND html_template NOT LIKE '%/assets/css/styles.css%';


---

-- use liced icons from cdn

-- Fix features component to use Lucide icons
-- The current template shows text labels like "zap", "shield" instead of icons

-- First, check what we have
SELECT name, function,
       SUBSTRING(html_template, 1, 200) as template_preview
FROM content_components
WHERE function = 'features' OR name LIKE '%feature%'
    LIMIT 5;

-- Update the features component to use Lucide icons from CDN
-- This adds the Lucide script and converts icon names to actual SVG icons
UPDATE content_components
SET html_template = '<section class="features-section" data-component="features">
        <div class="features-container">
            <h2>{{.title}}</h2>
            {{if .intro}}<p class="section-intro">{{.intro}}</p>{{end}}
            <div class="features-grid">
                {{range .features}}
                <div class="feature-item">
                    <div class="feature-icon" data-lucide="{{.icon}}"></div>
                    <h3>{{.title}}</h3>
                    <p>{{.description}}</p>
                </div>
                {{end}}
            </div>
        </div>
    </section>
<script src="https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"></script>
<script>
document.addEventListener("DOMContentLoaded", function() {
    if (typeof lucide !== "undefined") {
        lucide.createIcons();
    }
});
</script>
<style>
.features-section {
    padding: 5rem 2rem;
    background: var(--surface-color, #f8f9fa);
}
.features-container {
    max-width: 1200px;
    margin: 0 auto;
}
.features-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 1rem;
    color: var(--text-color, #333);
}
.section-intro {
    text-align: center;
    max-width: 700px;
    margin: 0 auto 3rem;
    color: var(--text-muted, #666);
    font-size: 1.1rem;
    line-height: 1.6;
}
.features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}
.feature-item {
    background: #fff;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
    transition: transform 0.2s, box-shadow 0.2s;
}
.feature-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.12);
}
.feature-icon {
    width: 48px;
    height: 48px;
    color: var(--accent-color, #0f3460);
    margin-bottom: 1rem;
}
.feature-icon svg {
    width: 100%;
    height: 100%;
}
.feature-item h3 {
    font-size: 1.25rem;
    margin-bottom: 0.75rem;
    color: var(--text-color, #333);
}
.feature-item p {
    color: var(--text-muted, #666);
    line-height: 1.6;
    margin: 0;
}
@media (max-width: 768px) {
    .features-section {
        padding: 3rem 1.5rem;
    }
    .features-grid {
        grid-template-columns: 1fr;
    }
}
</style>',
    updated_at = NOW()
WHERE function = 'features';

-- Verify
SELECT name, function,
       CASE WHEN html_template LIKE '%lucide%'
                THEN 'Has Lucide icons'
            ELSE 'No icons'
           END as icon_status
FROM content_components
WHERE function = 'features';


---

-- add stylesheet to head

-- Update head components to link the external stylesheet
-- The webdesign-agent deploys CSS to /assets/css/styles.css

-- Update head-seo-standard component (name = 'head-seo-standard', function = 'head')
UPDATE content_components
SET html_template = REPLACE(
        html_template,
        '</head>',
        '    <link rel="stylesheet" href="/assets/css/styles.css">
    </head>'
                    )
WHERE name = 'head-seo-standard'
  AND html_template NOT LIKE '%/assets/css/styles.css%';

-- Update "Document Head" component (simpler version)
UPDATE content_components
SET html_template = REPLACE(
        html_template,
        '</head>',
        '    <link rel="stylesheet" href="/assets/css/styles.css">
    </head>'
                    )
WHERE name = 'Document Head'
  AND html_template NOT LIKE '%/assets/css/styles.css%';

-- Verify the updates
SELECT name, function,
       CASE WHEN html_template LIKE '%/assets/css/styles.css%'
                THEN 'Has stylesheet link'
            ELSE 'Missing stylesheet link'
           END as stylesheet_status
FROM content_components
WHERE function = 'head'
   OR name IN ('head-seo-standard', 'Document Head');


----

-- Comprehensive Component Styling Fixes
-- Fixes: icons, centered containers, contact form, contact details consistency

-- ============================================================
-- 1. DIFFERENTIATORS COMPONENT - Add centered container styling
-- ============================================================
UPDATE content_components
SET html_template = '<section class="differentiators-section" data-component="differentiators">
        <div class="differentiators-container">
            <h2>{{.title}}</h2>
            <div class="differentiators-grid">
                {{range .differentiators}}
                <div class="differentiator-item">
                    <h3>{{.title}}</h3>
                    <p>{{.description}}</p>
                </div>
                {{end}}
            </div>
        </div>
    </section>
<style>
.differentiators-section {
    padding: 5rem 2rem;
    background: #fff;
}
.differentiators-container {
    max-width: 1200px;
    margin: 0 auto;
}
.differentiators-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 3rem;
    color: #1a1a2e;
}
.differentiators-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2rem;
}
.differentiator-item {
    padding: 2rem;
    background: #f8f9fa;
    border-radius: 8px;
    border-left: 4px solid #0f3460;
}
.differentiator-item h3 {
    font-size: 1.25rem;
    margin-bottom: 0.75rem;
    color: #1a1a2e;
}
.differentiator-item p {
    color: #555;
    line-height: 1.7;
    margin: 0;
}
@media (max-width: 768px) {
    .differentiators-section { padding: 3rem 1.5rem; }
    .differentiators-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'differentiators';

-- ============================================================
-- 2. SERVICES-GRID COMPONENT - Add centered container styling
-- ============================================================
UPDATE content_components
SET html_template = '<section class="services-grid-section" data-component="services-grid">
        <div class="services-container">
            <h2>{{.title}}</h2>
            <div class="services-grid">
                {{range .services}}
                <div class="service-item">
                    <h3>{{.name}}</h3>
                    <p>{{.description}}</p>
                    {{if .link}}<a href="{{.link}}" class="service-link">Learn more →</a>{{end}}
                </div>
                {{end}}
            </div>
        </div>
    </section>
<style>
.services-grid-section {
    padding: 5rem 2rem;
    background: #f8f9fa;
}
.services-container {
    max-width: 1200px;
    margin: 0 auto;
}
.services-grid-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 3rem;
    color: #1a1a2e;
}
.services-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
}
.service-item {
    background: #fff;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
    transition: transform 0.2s, box-shadow 0.2s;
}
.service-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.12);
}
.service-item h3 {
    font-size: 1.25rem;
    margin-bottom: 0.75rem;
    color: #1a1a2e;
}
.service-item p {
    color: #555;
    line-height: 1.7;
    margin-bottom: 1rem;
}
.service-link {
    color: #0f3460;
    text-decoration: none;
    font-weight: 500;
}
.service-link:hover {
    text-decoration: underline;
}
@media (max-width: 768px) {
    .services-grid-section { padding: 3rem 1.5rem; }
    .services-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'services-grid';

-- ============================================================
-- 3. SOCIAL-PROOF COMPONENT - Add centered container styling
-- ============================================================
UPDATE content_components
SET html_template = '<section class="social-proof-section" data-component="social-proof">
        <div class="social-proof-container">
            <h2>{{.title}}</h2>
            {{if .stats}}
            <div class="stats-grid">
                {{range .stats}}
                <div class="stat-item">
                    <span class="stat-number">{{.value}}</span>
                    <span class="stat-label">{{.label}}</span>
                </div>
                {{end}}
            </div>
            {{end}}
            {{if .client_logos}}
            <div class="logo-strip">
                {{range .client_logos}}
                <img src="{{.src}}" alt="{{.alt}}" class="client-logo">
                {{end}}
            </div>
            {{end}}
        </div>
    </section>
<style>
.social-proof-section {
    padding: 5rem 2rem;
    background: #1a1a2e;
    color: #fff;
}
.social-proof-container {
    max-width: 1200px;
    margin: 0 auto;
    text-align: center;
}
.social-proof-section h2 {
    font-size: clamp(1.5rem, 3vw, 2rem);
    margin-bottom: 3rem;
    color: #fff;
}
.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 2rem;
    margin-bottom: 3rem;
}
.stat-item {
    padding: 1.5rem;
}
.stat-number {
    display: block;
    font-size: clamp(2rem, 5vw, 3rem);
    font-weight: 700;
    color: #0f3460;
    margin-bottom: 0.5rem;
}
.stat-label {
    font-size: 0.95rem;
    color: rgba(255,255,255,0.8);
    line-height: 1.4;
}
.logo-strip {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
    gap: 2rem;
    opacity: 0.7;
}
.client-logo {
    height: 40px;
    width: auto;
    filter: brightness(0) invert(1);
}
@media (max-width: 768px) {
    .social-proof-section { padding: 3rem 1.5rem; }
    .stats-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>',
    updated_at = NOW()
WHERE function = 'social-proof';

-- ============================================================
-- 4. CTA COMPONENT - Add centered container styling
-- ============================================================
UPDATE content_components
SET html_template = '<section class="cta-section" data-component="call-to-action">
        <div class="cta-container">
            <h2>{{.title}}</h2>
            <p>{{.description}}</p>
            <a href="{{if .button_url}}{{.button_url}}{{else}}/contact.html{{end}}" class="btn btn-primary btn-large">{{if .button_text}}{{.button_text}}{{else}}Get Started{{end}}</a>
        </div>
    </section>
<style>
.cta-section {
    padding: 5rem 2rem;
    background: linear-gradient(135deg, #0f3460 0%, #1a1a2e 100%);
    text-align: center;
}
.cta-container {
    max-width: 800px;
    margin: 0 auto;
}
.cta-section h2 {
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    color: #fff;
    margin-bottom: 1rem;
}
.cta-section p {
    font-size: 1.1rem;
    color: rgba(255,255,255,0.9);
    margin-bottom: 2rem;
    line-height: 1.6;
}
.cta-section .btn {
    display: inline-block;
    padding: 1rem 2.5rem;
    background: #fff;
    color: #1a1a2e;
    text-decoration: none;
    border-radius: 4px;
    font-weight: 600;
    font-size: 1.1rem;
    transition: transform 0.2s, box-shadow 0.2s;
}
.cta-section .btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}
@media (max-width: 768px) {
    .cta-section { padding: 3rem 1.5rem; }
}
</style>',
    updated_at = NOW()
WHERE function = 'call-to-action';

-- ============================================================
-- 5. CONTACT-FORM COMPONENT - Better styling
-- ============================================================
UPDATE content_components
SET html_template = '<section class="contact-form-section" data-component="contact-form">
        <div class="form-container">
            <h2>{{if .title}}{{.title}}{{else}}Get In Touch{{end}}</h2>
            {{if .intro}}<p class="form-intro">{{.intro}}</p>{{end}}
            <form class="contact-form" action="/api/contact" method="POST">
                <div class="form-row">
                    <div class="form-group">
                        <label for="name">Name</label>
                        <input type="text" id="name" name="name" required placeholder="Your name">
                    </div>
                    <div class="form-group">
                        <label for="email">Email</label>
                        <input type="email" id="email" name="email" required placeholder="your@email.com">
                    </div>
                </div>
                <div class="form-group">
                    <label for="company">Company (Optional)</label>
                    <input type="text" id="company" name="company" placeholder="Your company">
                </div>
                <div class="form-group">
                    <label for="message">Message</label>
                    <textarea id="message" name="message" rows="5" required placeholder="Tell us about your project or challenge..."></textarea>
                </div>
                <button type="submit" class="btn btn-primary">Send Message</button>
            </form>
        </div>
    </section>
<style>
.contact-form-section {
    padding: 5rem 2rem;
    background: #fff;
}
.form-container {
    max-width: 700px;
    margin: 0 auto;
}
.contact-form-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.25rem);
    margin-bottom: 1rem;
    color: #1a1a2e;
}
.form-intro {
    text-align: center;
    color: #555;
    margin-bottom: 2rem;
    line-height: 1.6;
}
.contact-form {
    background: #f8f9fa;
    padding: 2.5rem;
    border-radius: 8px;
}
.form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
}
.form-group {
    margin-bottom: 1.5rem;
}
.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: #333;
}
.form-group input,
.form-group textarea {
    width: 100%;
    padding: 0.875rem 1rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
    font-family: inherit;
    transition: border-color 0.2s, box-shadow 0.2s;
    box-sizing: border-box;
}
.form-group input:focus,
.form-group textarea:focus {
    outline: none;
    border-color: #0f3460;
    box-shadow: 0 0 0 3px rgba(15, 52, 96, 0.1);
}
.form-group textarea {
    resize: vertical;
    min-height: 120px;
}
.contact-form .btn {
    display: block;
    width: 100%;
    padding: 1rem;
    background: #0f3460;
    color: #fff;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
}
.contact-form .btn:hover {
    background: #1a1a2e;
}
@media (max-width: 768px) {
    .contact-form-section { padding: 3rem 1.5rem; }
    .contact-form { padding: 1.5rem; }
    .form-row { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'contact-form';

-- ============================================================
-- 6. CONTACT-INFO COMPONENT - Use established contact details
-- ============================================================
UPDATE content_components
SET html_template = '<section class="contact-info-section" data-component="contact-info">
        <div class="contact-info-container">
            <h2>{{if .title}}{{.title}}{{else}}Contact Information{{end}}</h2>
            {{if .intro}}<p class="contact-intro">{{.intro}}</p>{{end}}
            <div class="contact-grid">
                <div class="contact-card">
                    <div class="contact-icon">✉</div>
                    <h3>Email</h3>
                    <a href="mailto:{{if .email}}{{.email}}{{else}}leopardess.consulting@contactforsales.com{{end}}">{{if .email}}{{.email}}{{else}}leopardess.consulting@contactforsales.com{{end}}</a>
                </div>
                <div class="contact-card">
                    <div class="contact-icon">☎</div>
                    <h3>Phone</h3>
                    <a href="tel:{{if .phone}}{{.phone}}{{else}}+447934524911{{end}}">{{if .phone_display}}{{.phone_display}}{{else}}+44 (0) 7934 524 911{{end}}</a>
                </div>
                <div class="contact-card">
                    <div class="contact-icon">⏰</div>
                    <h3>Hours</h3>
                    <p>{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm GMT{{end}}</p>
                </div>
            </div>
        </div>
    </section>
<style>
.contact-info-section {
    padding: 5rem 2rem;
    background: #f8f9fa;
}
.contact-info-container {
    max-width: 1000px;
    margin: 0 auto;
    text-align: center;
}
.contact-info-section h2 {
    font-size: clamp(1.75rem, 4vw, 2.25rem);
    margin-bottom: 1rem;
    color: #1a1a2e;
}
.contact-intro {
    color: #555;
    margin-bottom: 2.5rem;
    line-height: 1.6;
    max-width: 600px;
    margin-left: auto;
    margin-right: auto;
}
.contact-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 2rem;
}
.contact-card {
    background: #fff;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.contact-icon {
    font-size: 2rem;
    margin-bottom: 1rem;
}
.contact-card h3 {
    font-size: 1.1rem;
    margin-bottom: 0.5rem;
    color: #1a1a2e;
}
.contact-card a,
.contact-card p {
    color: #555;
    text-decoration: none;
    line-height: 1.5;
}
.contact-card a:hover {
    color: #0f3460;
}
@media (max-width: 768px) {
    .contact-info-section { padding: 3rem 1.5rem; }
    .contact-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'contact-info';

-- ============================================================
-- 7. FEATURES COMPONENT - With Lucide icons (CDN)
-- ============================================================
UPDATE content_components
SET html_template = '<section class="features-section" data-component="features">
        <div class="features-container">
            <h2>{{.title}}</h2>
            {{if .intro}}<p class="section-intro">{{.intro}}</p>{{end}}
            <div class="features-grid">
                {{range .features}}
                <div class="feature-item">
                    {{if .icon}}<i data-lucide="{{.icon}}" class="feature-icon"></i>{{end}}
                    <h3>{{.title}}</h3>
                    <p>{{.description}}</p>
                </div>
                {{end}}
            </div>
        </div>
    </section>
<script src="https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"></script>
<script>
document.addEventListener("DOMContentLoaded", function() {
    if (typeof lucide !== "undefined") { lucide.createIcons(); }
});
</script>
<style>
.features-section {
    padding: 5rem 2rem;
    background: #f8f9fa;
}
.features-container {
    max-width: 1200px;
    margin: 0 auto;
}
.features-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 1rem;
    color: #1a1a2e;
}
.section-intro {
    text-align: center;
    max-width: 700px;
    margin: 0 auto 3rem;
    color: #555;
    font-size: 1.1rem;
    line-height: 1.6;
}
.features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}
.feature-item {
    background: #fff;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
    transition: transform 0.2s, box-shadow 0.2s;
}
.feature-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.12);
}
.feature-icon {
    width: 48px;
    height: 48px;
    color: #0f3460;
    margin-bottom: 1rem;
}
.feature-item h3 {
    font-size: 1.25rem;
    margin-bottom: 0.75rem;
    color: #1a1a2e;
}
.feature-item p {
    color: #555;
    line-height: 1.7;
    margin: 0;
}
@media (max-width: 768px) {
    .features-section { padding: 3rem 1.5rem; }
    .features-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'features';

-- ============================================================
-- VERIFY UPDATES
-- ============================================================
SELECT function, name,
       CASE
           WHEN html_template LIKE '%max-width: 1200px%' OR html_template LIKE '%max-width: 1000px%' OR html_template LIKE '%max-width: 700px%' OR html_template LIKE '%max-width: 800px%'
               THEN 'Has centered container'
           ELSE 'Missing container'
           END as container_status,
       CASE
           WHEN html_template LIKE '%lucide%'
               THEN 'Has Lucide'
           ELSE '-'
           END as icons_status
FROM content_components
WHERE function IN ('differentiators', 'services-grid', 'social-proof', 'call-to-action', 'contact-form', 'contact-info', 'features');


---------------

-- Component CSS Variables Migration
--
-- Updates component templates to use CSS custom properties instead of hardcoded colors
-- This allows styles.css (from webdesign-agent) to control the color scheme
--
-- Pattern: var(--variable-name, fallback-value)
-- Fallback ensures components work even without styles.css loaded

-- ============================================================
-- Expected CSS Variables (defined by webdesign-agent in styles.css)
-- ============================================================
-- :root {
--   --color-primary: #1a1a2e;
--   --color-secondary: #16213e;
--   --color-accent: #0f3460;
--   --color-background: #ffffff;
--   --color-surface: #f8f9fa;
--   --color-text: #333333;
--   --color-text-muted: #555555;
--   --color-border: #e2e8f0;
--   --color-white: #ffffff;
--   --font-family: -apple-system, BlinkMacSystemFont, sans-serif;
--   --spacing-section: 5rem 2rem;
--   --container-max-width: 1200px;
-- }

-- ============================================================
-- 1. DIFFERENTIATORS COMPONENT
-- ============================================================
UPDATE content_components
SET html_template = '<section class="differentiators-section" data-component="differentiators">
        <div class="differentiators-container">
            <h2>{{.title}}</h2>
            <div class="differentiators-grid">
                {{range .differentiators}}
                <div class="differentiator-item">
                    <h3>{{.title}}</h3>
                    <p>{{.description}}</p>
                </div>
                {{end}}
            </div>
        </div>
    </section>
<style>
.differentiators-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background, #fff);
}
.differentiators-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.differentiators-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 3rem;
    color: var(--color-primary, #1a1a2e);
}
.differentiators-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2rem;
}
.differentiator-item {
    padding: 2rem;
    background: var(--color-surface, #f8f9fa);
    border-radius: 8px;
    border-left: 4px solid var(--color-accent, #0f3460);
}
.differentiator-item h3 {
    font-size: 1.25rem;
    margin-bottom: 0.75rem;
    color: var(--color-primary, #1a1a2e);
}
.differentiator-item p {
    color: var(--color-text-muted, #555);
    line-height: 1.7;
    margin: 0;
}
@media (max-width: 768px) {
    .differentiators-section { padding: 3rem 1.5rem; }
    .differentiators-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'differentiators';

-- ============================================================
-- 2. SERVICES-GRID COMPONENT
-- ============================================================
UPDATE content_components
SET html_template = '<section class="services-grid-section" data-component="services-grid">
        <div class="services-container">
            <h2>{{.title}}</h2>
            <div class="services-grid">
                {{range .services}}
                <div class="service-item">
                    <h3>{{.name}}</h3>
                    <p>{{.description}}</p>
                    {{if .link}}<a href="{{.link}}" class="service-link">Learn more →</a>{{end}}
                </div>
                {{end}}
            </div>
        </div>
    </section>
<style>
.services-grid-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-surface, #f8f9fa);
}
.services-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.services-grid-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 3rem;
    color: var(--color-primary, #1a1a2e);
}
.services-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
}
.service-item {
    background: var(--color-background, #fff);
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
    transition: transform 0.2s, box-shadow 0.2s;
}
.service-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.12);
}
.service-item h3 {
    font-size: 1.25rem;
    margin-bottom: 0.75rem;
    color: var(--color-primary, #1a1a2e);
}
.service-item p {
    color: var(--color-text-muted, #555);
    line-height: 1.7;
    margin-bottom: 1rem;
}
.service-link {
    color: var(--color-accent, #0f3460);
    text-decoration: none;
    font-weight: 500;
}
.service-link:hover {
    text-decoration: underline;
}
@media (max-width: 768px) {
    .services-grid-section { padding: 3rem 1.5rem; }
    .services-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'services-grid';

-- ============================================================
-- 3. SOCIAL-PROOF COMPONENT (dark background)
-- ============================================================
UPDATE content_components
SET html_template = '<section class="social-proof-section" data-component="social-proof">
        <div class="social-proof-container">
            <h2>{{.title}}</h2>
            {{if .stats}}
            <div class="stats-grid">
                {{range .stats}}
                <div class="stat-item">
                    <span class="stat-number">{{.value}}</span>
                    <span class="stat-label">{{.label}}</span>
                </div>
                {{end}}
            </div>
            {{end}}
            {{if .client_logos}}
            <div class="logo-strip">
                {{range .client_logos}}
                <img src="{{.src}}" alt="{{.alt}}" class="client-logo">
                {{end}}
            </div>
            {{end}}
        </div>
    </section>
<style>
.social-proof-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-primary, #1a1a2e);
    color: var(--color-white, #fff);
}
.social-proof-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
    text-align: center;
}
.social-proof-section h2 {
    font-size: clamp(1.5rem, 3vw, 2rem);
    margin-bottom: 3rem;
    color: var(--color-white, #fff);
}
.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 2rem;
    margin-bottom: 3rem;
}
.stat-item {
    padding: 1.5rem;
}
.stat-number {
    display: block;
    font-size: clamp(2rem, 5vw, 3rem);
    font-weight: 700;
    color: var(--color-accent, #0f3460);
    margin-bottom: 0.5rem;
}
.stat-label {
    font-size: 0.95rem;
    color: rgba(255,255,255,0.8);
    line-height: 1.4;
}
.logo-strip {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
    gap: 2rem;
    opacity: 0.7;
}
.client-logo {
    height: 40px;
    width: auto;
    filter: brightness(0) invert(1);
}
@media (max-width: 768px) {
    .social-proof-section { padding: 3rem 1.5rem; }
    .stats-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>',
    updated_at = NOW()
WHERE function = 'social-proof';

-- ============================================================
-- 4. CALL-TO-ACTION COMPONENT
-- ============================================================
UPDATE content_components
SET html_template = '<section class="cta-section" data-component="call-to-action">
        <div class="cta-container">
            <h2>{{.title}}</h2>
            <p>{{.description}}</p>
            <a href="{{if .button_url}}{{.button_url}}{{else}}/contact.html{{end}}" class="btn btn-primary btn-large">{{if .button_text}}{{.button_text}}{{else}}Get Started{{end}}</a>
        </div>
    </section>
<style>
.cta-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: linear-gradient(135deg, var(--color-accent, #0f3460) 0%, var(--color-primary, #1a1a2e) 100%);
    text-align: center;
}
.cta-container {
    max-width: 800px;
    margin: 0 auto;
}
.cta-section h2 {
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    color: var(--color-white, #fff);
    margin-bottom: 1rem;
}
.cta-section p {
    font-size: 1.1rem;
    color: rgba(255,255,255,0.9);
    margin-bottom: 2rem;
    line-height: 1.6;
}
.cta-section .btn {
    display: inline-block;
    padding: 1rem 2.5rem;
    background: var(--color-white, #fff);
    color: var(--color-primary, #1a1a2e);
    text-decoration: none;
    border-radius: 4px;
    font-weight: 600;
    font-size: 1.1rem;
    transition: transform 0.2s, box-shadow 0.2s;
}
.cta-section .btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}
@media (max-width: 768px) {
    .cta-section { padding: 3rem 1.5rem; }
}
</style>',
    updated_at = NOW()
WHERE function = 'call-to-action';

-- ============================================================
-- 5. CONTACT-FORM COMPONENT
-- ============================================================
UPDATE content_components
SET html_template = '<section class="contact-form-section" data-component="contact-form">
        <div class="form-container">
            <h2>{{if .title}}{{.title}}{{else}}Get In Touch{{end}}</h2>
            {{if .intro}}<p class="form-intro">{{.intro}}</p>{{end}}
            <form class="contact-form" action="/api/contact" method="POST">
                <div class="form-row">
                    <div class="form-group">
                        <label for="name">Name</label>
                        <input type="text" id="name" name="name" required placeholder="Your name">
                    </div>
                    <div class="form-group">
                        <label for="email">Email</label>
                        <input type="email" id="email" name="email" required placeholder="your@email.com">
                    </div>
                </div>
                <div class="form-group">
                    <label for="company">Company (Optional)</label>
                    <input type="text" id="company" name="company" placeholder="Your company">
                </div>
                <div class="form-group">
                    <label for="message">Message</label>
                    <textarea id="message" name="message" rows="5" required placeholder="Tell us about your project or challenge..."></textarea>
                </div>
                <button type="submit" class="btn btn-primary">Send Message</button>
            </form>
        </div>
    </section>
<style>
.contact-form-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background, #fff);
}
.form-container {
    max-width: 700px;
    margin: 0 auto;
}
.contact-form-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.25rem);
    margin-bottom: 1rem;
    color: var(--color-primary, #1a1a2e);
}
.form-intro {
    text-align: center;
    color: var(--color-text-muted, #555);
    margin-bottom: 2rem;
    line-height: 1.6;
}
.contact-form {
    background: var(--color-surface, #f8f9fa);
    padding: 2.5rem;
    border-radius: 8px;
}
.form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
}
.form-group {
    margin-bottom: 1.5rem;
}
.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: var(--color-text, #333);
}
.form-group input,
.form-group textarea {
    width: 100%;
    padding: 0.875rem 1rem;
    border: 1px solid var(--color-border, #ddd);
    border-radius: 4px;
    font-size: 1rem;
    font-family: inherit;
    transition: border-color 0.2s, box-shadow 0.2s;
    box-sizing: border-box;
}
.form-group input:focus,
.form-group textarea:focus {
    outline: none;
    border-color: var(--color-accent, #0f3460);
    box-shadow: 0 0 0 3px rgba(15, 52, 96, 0.1);
}
.form-group textarea {
    resize: vertical;
    min-height: 120px;
}
.contact-form .btn {
    display: block;
    width: 100%;
    padding: 1rem;
    background: var(--color-accent, #0f3460);
    color: var(--color-white, #fff);
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
}
.contact-form .btn:hover {
    background: var(--color-primary, #1a1a2e);
}
@media (max-width: 768px) {
    .contact-form-section { padding: 3rem 1.5rem; }
    .contact-form { padding: 1.5rem; }
    .form-row { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'contact-form';

-- ============================================================
-- 6. CONTACT-INFO COMPONENT
-- ============================================================
UPDATE content_components
SET html_template = '<section class="contact-info-section" data-component="contact-info">
        <div class="contact-info-container">
            <h2>{{if .title}}{{.title}}{{else}}Contact Information{{end}}</h2>
            {{if .intro}}<p class="contact-intro">{{.intro}}</p>{{end}}
            <div class="contact-grid">
                <div class="contact-card">
                    <div class="contact-icon">✉</div>
                    <h3>Email</h3>
                    <a href="mailto:{{if .email}}{{.email}}{{else}}info@example.com{{end}}">{{if .email}}{{.email}}{{else}}info@example.com{{end}}</a>
                </div>
                <div class="contact-card">
                    <div class="contact-icon">☎</div>
                    <h3>Phone</h3>
                    <a href="tel:{{if .phone}}{{.phone}}{{else}}+1234567890{{end}}">{{if .phone_display}}{{.phone_display}}{{else if .phone}}{{.phone}}{{else}}+1 (234) 567-890{{end}}</a>
                </div>
                <div class="contact-card">
                    <div class="contact-icon">⏰</div>
                    <h3>Hours</h3>
                    <p>{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}</p>
                </div>
            </div>
        </div>
    </section>
<style>
.contact-info-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-surface, #f8f9fa);
}
.contact-info-container {
    max-width: 1000px;
    margin: 0 auto;
    text-align: center;
}
.contact-info-section h2 {
    font-size: clamp(1.75rem, 4vw, 2.25rem);
    margin-bottom: 1rem;
    color: var(--color-primary, #1a1a2e);
}
.contact-intro {
    color: var(--color-text-muted, #555);
    margin-bottom: 2.5rem;
    line-height: 1.6;
    max-width: 600px;
    margin-left: auto;
    margin-right: auto;
}
.contact-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 2rem;
}
.contact-card {
    background: var(--color-background, #fff);
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.contact-icon {
    font-size: 2rem;
    margin-bottom: 1rem;
}
.contact-card h3 {
    font-size: 1.1rem;
    margin-bottom: 0.5rem;
    color: var(--color-primary, #1a1a2e);
}
.contact-card a,
.contact-card p {
    color: var(--color-text-muted, #555);
    text-decoration: none;
    line-height: 1.5;
}
.contact-card a:hover {
    color: var(--color-accent, #0f3460);
}
@media (max-width: 768px) {
    .contact-info-section { padding: 3rem 1.5rem; }
    .contact-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'contact-info';

-- ============================================================
-- 7. FEATURES COMPONENT
-- ============================================================
UPDATE content_components
SET html_template = '<section class="features-section" data-component="features">
        <div class="features-container">
            <h2>{{.title}}</h2>
            {{if .intro}}<p class="section-intro">{{.intro}}</p>{{end}}
            <div class="features-grid">
                {{range .features}}
                <div class="feature-item">
                    {{if .icon}}<i data-lucide="{{.icon}}" class="feature-icon"></i>{{end}}
                    <h3>{{.title}}</h3>
                    <p>{{.description}}</p>
                </div>
                {{end}}
            </div>
        </div>
    </section>
<script src="https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"></script>
<script>
document.addEventListener("DOMContentLoaded", function() {
    if (typeof lucide !== "undefined") { lucide.createIcons(); }
});
</script>
<style>
.features-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-surface, #f8f9fa);
}
.features-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.features-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 1rem;
    color: var(--color-primary, #1a1a2e);
}
.section-intro {
    text-align: center;
    max-width: 700px;
    margin: 0 auto 3rem;
    color: var(--color-text-muted, #555);
    font-size: 1.1rem;
    line-height: 1.6;
}
.features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}
.feature-item {
    background: var(--color-background, #fff);
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
    transition: transform 0.2s, box-shadow 0.2s;
}
.feature-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.12);
}
.feature-icon {
    width: 48px;
    height: 48px;
    color: var(--color-accent, #0f3460);
    margin-bottom: 1rem;
}
.feature-item h3 {
    font-size: 1.25rem;
    margin-bottom: 0.75rem;
    color: var(--color-primary, #1a1a2e);
}
.feature-item p {
    color: var(--color-text-muted, #555);
    line-height: 1.7;
    margin: 0;
}
@media (max-width: 768px) {
    .features-section { padding: 3rem 1.5rem; }
    .features-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'features';

-- ============================================================
-- 8. CASE-STUDIES-LIST COMPONENT
-- ============================================================
UPDATE content_components
SET html_template = '<section class="case-studies-section" data-component="case-studies-list">
        <div class="case-studies-container">
            <h2>{{if .title}}{{.title}}{{else}}Case Studies{{end}}</h2>
            <div class="case-studies-grid">
                {{range .case_studies}}
                <div class="case-study-item">
                    {{if .image}}<img src="{{.image}}" alt="{{.title}}" class="case-study-image">{{end}}
                    <h3>{{.title}}</h3>
                    {{if .client}}<p class="case-study-client">{{.client}}</p>{{end}}
                    <p>{{.description}}</p>
                    {{if .link}}<a href="{{.link}}" class="case-study-link">Read more</a>{{end}}
                </div>
                {{end}}
            </div>
        </div>
    </section>
<style>
.case-studies-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background, #fff);
}
.case-studies-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.case-studies-section h2 {
    text-align: center;
    font-size: clamp(1.75rem, 4vw, 2.5rem);
    margin-bottom: 3rem;
    color: var(--color-primary, #1a1a2e);
}
.case-studies-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2rem;
}
.case-study-item {
    background: var(--color-surface, #f8f9fa);
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
    transition: transform 0.2s, box-shadow 0.2s;
}
.case-study-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.12);
}
.case-study-image {
    width: 100%;
    height: 200px;
    object-fit: cover;
}
.case-study-item h3 {
    padding: 1.5rem 1.5rem 0.5rem;
    font-size: 1.25rem;
    color: var(--color-primary, #1a1a2e);
}
.case-study-client {
    padding: 0 1.5rem;
    font-size: 0.9rem;
    color: var(--color-accent, #0f3460);
    font-weight: 500;
    margin-bottom: 0.5rem;
}
.case-study-item p {
    padding: 0 1.5rem 1rem;
    color: var(--color-text-muted, #555);
    line-height: 1.6;
}
.case-study-link {
    display: block;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--color-border, #e2e8f0);
    color: var(--color-accent, #0f3460);
    text-decoration: none;
    font-weight: 500;
}
.case-study-link:hover {
    background: var(--color-background, #fff);
}
@media (max-width: 768px) {
    .case-studies-section { padding: 3rem 1.5rem; }
    .case-studies-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'case-studies-list';

-- ============================================================
-- VERIFY UPDATES
-- ============================================================
SELECT function, name,
       CASE
           WHEN html_template LIKE '%var(--color%'
               THEN 'Uses CSS variables'
           ELSE 'Hardcoded colors'
           END as css_status
FROM content_components
WHERE function IN ('differentiators', 'services-grid', 'social-proof', 'call-to-action',
                   'contact-form', 'contact-info', 'features', 'case-studies-list')
ORDER BY function;


-- Update footer templates to use FooterNavItems with fallback to NavItems
--
-- PREREQUISITE: Add to RenderContext in types.go:
--   FooterNavItems []NavItem `json:"footer_nav_items"`
--
-- Pattern change in templates:
--   BEFORE: {{range .NavItems}}<a href="{{.URL}}">{{.Label}}</a>{{end}}
--   AFTER:  {{if .FooterNavItems}}{{range .FooterNavItems}}<a href="{{.URL}}">{{.Label}}</a>{{end}}{{else if .NavItems}}{{range .NavItems}}<a href="{{.URL}}">{{.Label}}</a>{{end}}{{end}}

-- ============================================================
-- Step 1: Check which footers need updating
-- ============================================================
SELECT name, function,
       CASE
           WHEN html_template LIKE '%FooterNavItems%' THEN 'DONE'
           WHEN html_template LIKE '%NavItems%' THEN 'NEEDS UPDATE'
           ELSE 'NO NAV'
           END as status
FROM content_components
WHERE function LIKE '%footer%' OR name LIKE '%footer%'
ORDER BY name;

-- ============================================================
-- Step 2: Update templates using text replacement
-- ============================================================
-- This handles the common pattern where nav items are inside a <nav> tag

UPDATE content_components
SET html_template = regexp_replace(
        html_template,
        '\{\{range \.NavItems\}\}(.*?)\{\{end\}\}',
        '{{if .FooterNavItems}}{{range .FooterNavItems}}\1{{end}}{{else if .NavItems}}{{range .NavItems}}\1{{end}}{{end}}',
        'gs'
                    ),
    updated_at = NOW()
WHERE (function LIKE '%footer%' OR name LIKE '%footer%')
  AND html_template LIKE '%{{range .NavItems}}%'
  AND html_template NOT LIKE '%FooterNavItems%';

-- ============================================================
-- Step 3: Verify changes
-- ============================================================
SELECT name, function,
       CASE
           WHEN html_template LIKE '%FooterNavItems%' THEN 'UPDATED'
           WHEN html_template LIKE '%NavItems%' THEN 'STILL NEEDS WORK'
           ELSE 'NO NAV'
           END as status
FROM content_components
WHERE function LIKE '%footer%' OR name LIKE '%footer%'
ORDER BY name;


-- global vs local css

-- CSS Responsibility Barrier Implementation
--
-- PRINCIPLE: Global CSS handles all appearance (colors, fonts).
--            Component CSS handles only layout/structure.
--
-- Components should NOT re-declare colors on elements that global CSS styles.
-- Components should ONLY use color variables for:
--   - Section backgrounds (when different from page background)
--   - Borders (accent colors)
--   - Dark/inverted sections

-- ============================================================
-- COMPONENT CSS RULES:
-- ============================================================
-- DO:
--   - Use var(--color-surface) for section backgrounds
--   - Use var(--color-accent) for accent borders
--   - Use var(--color-primary) for dark section backgrounds
--   - Define grid, flexbox, positioning
--   - Define component-specific spacing (gaps, margins)
--
-- DO NOT:
--   - Re-declare color on h1, h2, h3, p, a (global handles this)
--   - Re-declare font-family (global handles this)
--   - Re-declare base font-size (global handles this)
--
-- EXCEPTION - Dark/inverted sections:
--   If a component has dark background, it MUST override text color
--   using var(--color-white) or explicit light color

-- ============================================================
-- 1. DIFFERENTIATORS - Layout only, no color overrides on text
-- ============================================================
UPDATE content_components
SET html_template = '<section class="differentiators-section" data-component="differentiators">
    <div class="differentiators-container">
        <h2>{{.title}}</h2>
        <div class="differentiators-grid">
            {{range .differentiators}}
            <div class="differentiator-item">
                <h3>{{.title}}</h3>
                <p>{{.description}}</p>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Layout only - colors inherited from global CSS */
.differentiators-section {
    padding: var(--spacing-section, 5rem 2rem);
}
.differentiators-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.differentiators-section h2 {
    text-align: center;
    margin-bottom: 3rem;
}
.differentiators-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}
.differentiator-item {
    padding: 2rem;
    background: var(--color-surface, #f8f9fa);
    border-radius: 8px;
    border-left: 4px solid var(--color-accent, #0f3460);
}
.differentiator-item h3 {
    margin-bottom: 0.75rem;
}
.differentiator-item p {
    margin: 0;
}
@media (max-width: 768px) {
    .differentiators-section { padding: 3rem 1.5rem; }
    .differentiators-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE name = 'differentiators' OR function = 'differentiators-section';

-- ============================================================
-- 2. SERVICES-GRID - Layout only
-- ============================================================
UPDATE content_components
SET html_template = '<section class="services-section" data-component="services-grid">
    <div class="services-container">
        <h2>{{.title}}</h2>
        {{if .subtitle}}<p class="services-subtitle">{{.subtitle}}</p>{{end}}
        <div class="services-grid">
            {{range .services}}
            <div class="service-item">
                <h3>{{.name}}</h3>
                <p>{{.description}}</p>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Layout only - colors inherited from global CSS */
.services-section {
    padding: var(--spacing-section, 5rem 2rem);
}
.services-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.services-section h2 {
    text-align: center;
    margin-bottom: 1rem;
}
.services-subtitle {
    text-align: center;
    max-width: 600px;
    margin: 0 auto 3rem;
}
.services-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
}
.service-item {
    padding: 2rem;
    background: var(--color-surface, #f8f9fa);
    border-radius: 8px;
    transition: transform 0.2s, box-shadow 0.2s;
}
.service-item:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.1);
}
.service-item h3 {
    margin-bottom: 1rem;
}
.service-item p {
    margin: 0;
}
@media (max-width: 768px) {
    .services-section { padding: 3rem 1.5rem; }
    .services-grid { grid-template-columns: 1fr; gap: 1.5rem; }
}
</style>',
    updated_at = NOW()
WHERE name = 'services-grid' OR function = 'services-grid';

-- ============================================================
-- 3. SOCIAL-PROOF (Dark section - MUST override text colors)
-- ============================================================
UPDATE content_components
SET html_template = '<section class="social-proof-section" data-component="social-proof">
    <div class="social-proof-container">
        <h2>{{.title}}</h2>
        <div class="testimonials-grid">
            {{range .testimonials}}
            <div class="testimonial-item">
                <blockquote>{{.quote}}</blockquote>
                <cite>
                    <strong>{{.author}}</strong>
                    {{if .role}}<span>{{.role}}</span>{{end}}
                </cite>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Dark section - MUST override text colors */
.social-proof-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-primary, #1a1a2e);
    color: var(--color-white, #fff);
}
.social-proof-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.social-proof-section h2 {
    text-align: center;
    margin-bottom: 3rem;
    color: var(--color-white, #fff);
}
.testimonials-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}
.testimonial-item {
    padding: 2rem;
    background: rgba(255,255,255,0.05);
    border-radius: 8px;
    border-left: 3px solid var(--color-accent, #0f3460);
}
.testimonial-item blockquote {
    font-size: 1.1rem;
    line-height: 1.7;
    margin: 0 0 1.5rem;
    font-style: italic;
    color: rgba(255,255,255,0.9);
}
.testimonial-item cite {
    display: block;
    font-style: normal;
}
.testimonial-item cite strong {
    display: block;
    color: var(--color-white, #fff);
}
.testimonial-item cite span {
    font-size: 0.9rem;
    color: rgba(255,255,255,0.7);
}
@media (max-width: 768px) {
    .social-proof-section { padding: 3rem 1.5rem; }
    .testimonials-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE name = 'social-proof' OR function = 'social_proof' OR function = 'testimonials';

-- ============================================================
-- 4. CALL-TO-ACTION (Dark section)
-- ============================================================
UPDATE content_components
SET html_template = '<section class="cta-section" data-component="call-to-action">
    <div class="cta-container">
        <h2>{{.title}}</h2>
        {{if .subtitle}}<p class="cta-subtitle">{{.subtitle}}</p>{{end}}
        <div class="cta-buttons">
            {{if .primary_button}}
            <a href="{{.primary_button.url}}" class="cta-btn cta-btn-primary">{{.primary_button.text}}</a>
            {{end}}
            {{if .secondary_button}}
            <a href="{{.secondary_button.url}}" class="cta-btn cta-btn-secondary">{{.secondary_button.text}}</a>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Dark section - MUST override text colors */
.cta-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-primary, #1a1a2e);
    color: var(--color-white, #fff);
    text-align: center;
}
.cta-container {
    max-width: 800px;
    margin: 0 auto;
}
.cta-section h2 {
    margin-bottom: 1rem;
    color: var(--color-white, #fff);
}
.cta-subtitle {
    margin-bottom: 2rem;
    color: rgba(255,255,255,0.85);
}
.cta-buttons {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
}
.cta-btn {
    display: inline-block;
    padding: 1rem 2rem;
    border-radius: 6px;
    text-decoration: none;
    font-weight: 600;
    transition: transform 0.2s, box-shadow 0.2s;
}
.cta-btn:hover {
    transform: translateY(-2px);
}
.cta-btn-primary {
    background: var(--color-white, #fff);
    color: var(--color-primary, #1a1a2e);
}
.cta-btn-secondary {
    background: transparent;
    border: 2px solid var(--color-white, #fff);
    color: var(--color-white, #fff);
}
@media (max-width: 768px) {
    .cta-section { padding: 3rem 1.5rem; }
    .cta-buttons { flex-direction: column; align-items: center; }
    .cta-btn { width: 100%; max-width: 280px; text-align: center; }
}
</style>',
    updated_at = NOW()
WHERE name = 'call-to-action' OR function = 'call_to_action' OR function = 'cta';

-- ============================================================
-- 5. FEATURES (Light section - layout only)
-- ============================================================
UPDATE content_components
SET html_template = '<section class="features-section" data-component="features">
    <div class="features-container">
        <h2>{{.title}}</h2>
        {{if .subtitle}}<p class="features-subtitle">{{.subtitle}}</p>{{end}}
        <div class="features-grid">
            {{range .features}}
            <div class="feature-item">
                {{if .icon}}<div class="feature-icon">{{.icon}}</div>{{end}}
                <h3>{{.title}}</h3>
                <p>{{.description}}</p>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Layout only - colors inherited from global CSS */
.features-section {
    padding: var(--spacing-section, 5rem 2rem);
}
.features-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.features-section h2 {
    text-align: center;
    margin-bottom: 1rem;
}
.features-subtitle {
    text-align: center;
    max-width: 600px;
    margin: 0 auto 3rem;
}
.features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
}
.feature-item {
    text-align: center;
    padding: 2rem;
}
.feature-icon {
    font-size: 2.5rem;
    margin-bottom: 1rem;
    color: var(--color-accent, #0f3460);
}
.feature-item h3 {
    margin-bottom: 1rem;
}
.feature-item p {
    margin: 0;
}
@media (max-width: 768px) {
    .features-section { padding: 3rem 1.5rem; }
    .features-grid { grid-template-columns: 1fr; gap: 1.5rem; }
}
</style>',
    updated_at = NOW()
WHERE name = 'features' OR function = 'features';

-- ============================================================
-- 6. CONTACT-FORM (Light section - layout only)
-- ============================================================
UPDATE content_components
SET html_template = '<section class="contact-form-section" data-component="contact-form">
    <div class="contact-form-container">
        <h2>{{.title}}</h2>
        {{if .subtitle}}<p class="contact-subtitle">{{.subtitle}}</p>{{end}}
        <form class="contact-form" action="{{.form_action}}" method="POST">
            <div class="form-group">
                <label for="name">Name</label>
                <input type="text" id="name" name="name" required>
            </div>
            <div class="form-group">
                <label for="email">Email</label>
                <input type="email" id="email" name="email" required>
            </div>
            <div class="form-group">
                <label for="message">Message</label>
                <textarea id="message" name="message" rows="5" required></textarea>
            </div>
            <button type="submit" class="form-submit">{{if .button_text}}{{.button_text}}{{else}}Send Message{{end}}</button>
        </form>
    </div>
</section>
<style>
/* Layout only - colors inherited from global CSS */
.contact-form-section {
    padding: var(--spacing-section, 5rem 2rem);
}
.contact-form-container {
    max-width: 600px;
    margin: 0 auto;
}
.contact-form-section h2 {
    text-align: center;
    margin-bottom: 1rem;
}
.contact-subtitle {
    text-align: center;
    margin-bottom: 2rem;
}
.contact-form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}
.form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}
.form-group label {
    font-weight: 500;
}
.form-group input,
.form-group textarea {
    padding: 0.75rem 1rem;
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 6px;
    font-family: inherit;
    font-size: 1rem;
    transition: border-color 0.2s, box-shadow 0.2s;
}
.form-group input:focus,
.form-group textarea:focus {
    outline: none;
    border-color: var(--color-accent, #0f3460);
    box-shadow: 0 0 0 3px rgba(15,52,96,0.1);
}
.form-submit {
    padding: 1rem 2rem;
    background: var(--color-accent, #0f3460);
    color: var(--color-white, #fff);
    border: none;
    border-radius: 6px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s, transform 0.2s;
}
.form-submit:hover {
    transform: translateY(-2px);
}
@media (max-width: 768px) {
    .contact-form-section { padding: 3rem 1.5rem; }
}
</style>',
    updated_at = NOW()
WHERE name = 'contact-form' OR function = 'contact-form';

-- ============================================================
-- 7. CASE-STUDIES (Light section - layout only)
-- ============================================================
UPDATE content_components
SET html_template = '<section class="case-studies-section" data-component="case-studies-list">
    <div class="case-studies-container">
        <h2>{{.title}}</h2>
        {{if .subtitle}}<p class="case-studies-subtitle">{{.subtitle}}</p>{{end}}
        <div class="case-studies-grid">
            {{range .case_studies}}
            <article class="case-study-item">
                <h3>{{.title}}</h3>
                <p class="case-study-client">{{.client}}</p>
                <p>{{.description}}</p>
                {{if .results}}<p class="case-study-results"><strong>Results:</strong> {{.results}}</p>{{end}}
            </article>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Layout only - colors inherited from global CSS */
.case-studies-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-surface, #f8f9fa);
}
.case-studies-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.case-studies-section h2 {
    text-align: center;
    margin-bottom: 1rem;
}
.case-studies-subtitle {
    text-align: center;
    max-width: 600px;
    margin: 0 auto 3rem;
}
.case-studies-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2rem;
}
.case-study-item {
    padding: 2rem;
    background: var(--color-background, #fff);
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}
.case-study-item h3 {
    margin-bottom: 0.5rem;
}
.case-study-client {
    font-size: 0.9rem;
    margin-bottom: 1rem;
    color: var(--color-text-muted, #666);
}
.case-study-results {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--color-border, #e2e8f0);
}
@media (max-width: 768px) {
    .case-studies-section { padding: 3rem 1.5rem; }
    .case-studies-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE name = 'case-studies-list' OR function = 'case-studies-list';

-- ============================================================
-- VERIFY
-- ============================================================
SELECT name, function,
       CASE
           WHEN html_template LIKE '%color: var(--color-primary%'
               AND html_template NOT LIKE '%background:%var(--color-primary%' THEN 'NEEDS REVIEW - sets text color'
           WHEN html_template LIKE '%Dark section%' THEN 'OK - Dark section'
           ELSE 'OK - Layout only'
           END as css_responsibility
FROM content_components
WHERE function IN ('differentiators-section', 'services-grid', 'social_proof',
                   'testimonials', 'call_to_action', 'cta', 'features',
                   'contact-form', 'case-studies-list')
ORDER BY name;

-- fix <no value> for eg social proof

-- =============================================================
-- Fix template field name mismatches causing <no value> and
-- empty headings
--
-- ROOT CAUSE:
-- LLM prompt tells model to return: headline, features[].name
-- Templates render: .title, .services[].name (or .title)
-- Input schemas tell LLM: section_title, services[].title
--
-- Three-way mismatch → LLM output doesn't match templates.
--
-- FIX: Align templates and schemas to match the standardized
-- LLM prompt output format:
--   - Section heading: headline (hero/services/CTA) or heading (contact/text)
--   - Feature/service arrays: features[].name, features[].description
--   - CTA: primary_cta, primary_cta_url
-- =============================================================


-- ============================================================
-- 1. services-grid
-- ============================================================
-- Template uses: {{.title}} for h2, {{range .services}} → {{.name}}
-- LLM prompt example returns: headline, features[].name
-- Schema says: section_title, services[].title
--
-- Fix: template to use headline, features, name
--       schema to match

-- Verify current state
SELECT name,
       html_template LIKE '%{{.title}}%' as has_dot_title,
       html_template LIKE '%{{range .services}}%' as has_range_services,
       html_template LIKE '%{{.name}}%' as has_dot_name,
       input_schema
FROM content_components WHERE name = 'services-grid';

-- Update template field names
UPDATE content_components
SET html_template = replace(
        replace(
                html_template,
                '<h2>{{.title}}</h2>',
                '<h2>{{.headline}}</h2>'
        ),
        '{{range .services}}',
        '{{range .features}}'
                    ),
    input_schema = '{"headline": "string", "subheadline": "string", "features": [{"name": "string", "description": "string"}]}'::jsonb,
    updated_at = now()
WHERE name = 'services-grid';

-- Verify
SELECT name,
       html_template LIKE '%{{.headline}}%' as has_headline,
       html_template LIKE '%{{range .features}}%' as has_range_features,
       html_template LIKE '%{{.name}}%' as has_dot_name,
       input_schema
FROM content_components WHERE name = 'services-grid';


-- ============================================================
-- 2. differentiators-section
-- ============================================================
-- Template uses: {{.title}} for h2, {{range .differentiators}} → {{.title}}, {{.description}}
-- LLM sees schema: section_title, differentiators[].title
--
-- Fix: h2 to use headline, range to use features, items to use name
-- Note: {{.title}} appears twice - once in h2 (top-level) and once
--       in range (item-level). We only change the h2 one and the
--       range item field.

-- Verify current state
SELECT name,
       html_template LIKE '%<h2>{{.title}}</h2>%' as has_h2_title,
       html_template LIKE '%{{range .differentiators}}%' as has_range_diff,
       input_schema
FROM content_components WHERE name = 'differentiators-section';

-- Update template
-- Step 1: Change h2 heading from .title to .headline
UPDATE content_components
SET html_template = replace(
        html_template,
        '<h2>{{.title}}</h2>',
        '<h2>{{.headline}}</h2>'
                    ),
    updated_at = now()
WHERE name = 'differentiators-section';

-- Step 2: Change range from .differentiators to .features
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{range .differentiators}}',
        '{{range .features}}'
                    ),
    updated_at = now()
WHERE name = 'differentiators-section';

-- Step 3: Inside the range, change item .title to .name
-- The h2 is already changed to .headline, so remaining .title
-- instances are inside the range block
UPDATE content_components
SET html_template = replace(
        html_template,
        '<h3>{{.title}}</h3>',
        '<h3>{{.name}}</h3>'
                    ),
    updated_at = now()
WHERE name = 'differentiators-section';

-- Step 4: Update schema
UPDATE content_components
SET input_schema = '{"headline": "string", "features": [{"name": "string", "description": "string"}]}'::jsonb,
    updated_at = now()
WHERE name = 'differentiators-section';

-- Verify
SELECT name,
       html_template LIKE '%{{.headline}}%' as has_headline,
       html_template LIKE '%{{range .features}}%' as has_range_features,
       html_template LIKE '%{{.name}}%' as has_dot_name,
       input_schema
FROM content_components WHERE name = 'differentiators-section';


-- ============================================================
-- 3. contact-form
-- ============================================================
-- Template uses: {{.title}} for h2
-- LLM prompt returns: heading (for contact/text sections)
-- Schema says: form_title
--
-- Fix: h2 to use heading

-- Verify current state
SELECT name,
       html_template LIKE '%<h2>{{.title}}</h2>%' as has_h2_title,
       input_schema
FROM content_components WHERE name = 'contact-form';

-- Update template
UPDATE content_components
SET html_template = replace(
        html_template,
        '<h2>{{.title}}</h2>',
        '<h2>{{.heading}}</h2>'
                    ),
    input_schema = '{"heading": "string", "form_description": "string", "submit_text": "string", "form_action": "string"}'::jsonb,
    updated_at = now()
WHERE name = 'contact-form';

-- Verify
SELECT name,
       html_template LIKE '%{{.heading}}%' as has_heading,
       input_schema
FROM content_components WHERE name = 'contact-form';


-- ============================================================
-- 4. Check for other templates using {{.title}} for h2
--    (to catch social-proof, call-to-action, etc.)
-- ============================================================
SELECT name, function, category, render_mode,
       html_template LIKE '%<h2>{{.title}}</h2>%' as has_h2_title
FROM content_components
WHERE html_template LIKE '%<h2>{{.title}}</h2>%'
  AND component_level = 'section'
ORDER BY name;


-- =============================================================
-- Fix remaining templates using <h2>{{.title}}</h2>
--
-- The LLM prompt returns "headline" for section headings.
-- These templates still use {{.title}} → renders empty.
--
-- Affected: call_to_action, social_proof, features,
--           testimonials, case-studies-list
-- Plus duplicate entries: "Call to Action", "Social Proof",
--           "Features Grid"
-- =============================================================

-- Preview: show all affected templates and their current state
SELECT name, function, category, render_mode,
       html_template LIKE '%<h2>{{.title}}</h2>%' as has_plain_title,
       html_template LIKE '%{{if .title}}%' as has_if_title
FROM content_components
WHERE html_template LIKE '%{{.title}}%'
  AND component_level = 'section'
ORDER BY name;

-- Fix: Replace <h2>{{.title}}</h2> with <h2>{{.headline}}</h2>
-- for all section-level components
--
-- NOTE: This only replaces the exact pattern <h2>{{.title}}</h2>
-- It does NOT replace {{.title}} inside {{range}} blocks
-- (where .title refers to an item title, not section heading)

UPDATE content_components
SET html_template = replace(
        html_template,
        '<h2>{{.title}}</h2>',
        '<h2>{{.headline}}</h2>'
                    ),
    updated_at = now()
WHERE html_template LIKE '%<h2>{{.title}}</h2>%'
  AND component_level = 'section';

-- Also handle the pattern with conditional:
-- {{if .title}}{{.title}}{{else}}Default{{end}}
-- This appears in contact-info template
-- We'll handle this separately per-component to preserve defaults

-- Verify what changed
SELECT name, function,
       html_template LIKE '%{{.headline}}%' as has_headline,
       html_template LIKE '%<h2>{{.title}}</h2>%' as still_has_title
FROM content_components
WHERE component_level = 'section'
  AND (html_template LIKE '%{{.headline}}%' OR html_template LIKE '%<h2>{{.title}}</h2>%')
ORDER BY name;


-- =============================================================
-- Also fix features template: range field name
-- The features template likely uses {{range .features}} with
-- items having {{.title}} — check and fix item-level .title
-- =============================================================

-- Check features templates for item-level field names
SELECT name, function,
       html_template LIKE '%{{range .features}}%' as has_range_features,
       html_template LIKE '%{{range .items}}%' as has_range_items,
    left(html_template, 600) as preview
FROM content_components
WHERE function = 'features' OR name IN ('features', 'Features Grid');

-- For features items, the LLM returns features[].name
-- If template uses {{.title}} inside range, fix to {{.name}}
-- (Run after checking the above query results)

-- UPDATE content_components
-- SET html_template = replace(
--         html_template,
--         '<h3>{{.title}}</h3>',
--         '<h3>{{.name}}</h3>'
--     ),
--     updated_at = now()
-- WHERE function = 'features'
--   AND html_template LIKE '%{{range %}%'
--   AND html_template LIKE '%<h3>{{.title}}</h3>%';


-- =============================================================
-- Update input_schema for the fixed components
-- Align schemas with what the LLM prompt examples return
-- =============================================================

-- call_to_action / Call to Action
UPDATE content_components
SET input_schema = '{"headline": "string", "subheadline": "string", "primary_cta": "string", "primary_cta_url": "string", "secondary_cta": "string", "secondary_cta_url": "string"}'::jsonb,
    updated_at = now()
WHERE function = 'call_to_action';

-- social_proof / Social Proof
UPDATE content_components
SET input_schema = '{"headline": "string", "testimonials": [{"quote": "string", "author": "string", "role": "string", "company": "string"}]}'::jsonb,
    updated_at = now()
WHERE function = 'social_proof';

-- features / Features Grid
UPDATE content_components
SET input_schema = '{"headline": "string", "subheadline": "string", "features": [{"name": "string", "description": "string", "icon": "string"}]}'::jsonb,
    updated_at = now()
WHERE function = 'features';

-- testimonials
UPDATE content_components
SET input_schema = '{"headline": "string", "testimonials": [{"quote": "string", "author": "string", "role": "string", "company": "string"}]}'::jsonb,
    updated_at = now()
WHERE function = 'testimonials';


-- other template changes

-- =============================================================
-- Fix remaining template field mismatches
--
-- call_to_action:
--   1. .subtitle → .subheadline
--   2. .primary_button.url/.text → .primary_cta / .primary_cta_url (flat)
--   3. .secondary_button.url/.text → .secondary_cta / .secondary_cta_url (flat)
--
-- features / Features Grid:
--   1. .subtitle → .subheadline
--   2. <h3>{{.title}}</h3> → <h3>{{.name}}</h3> (inside range)
--
-- case-studies-list:
--   1. {{.description}} → {{.summary}} (inside range)
--
-- social_proof / testimonials: Already aligned, no changes.
-- =============================================================


-- =============================================================
-- 1. call_to_action — Fix subtitle and CTA button structure
-- =============================================================

-- The CTA template currently uses nested .primary_button.url / .text
-- but the LLM returns flat: primary_cta, primary_cta_url
--
-- We need to rewrite the button HTML to use flat field access.
-- Also change .subtitle to .subheadline.
--
-- Both "call_to_action" and "Call to Action" have identical
-- templates, so we fix both with WHERE function = 'call_to_action'

-- First verify current state
SELECT name,
       html_template LIKE '%primary_button%' as has_nested_buttons,
       html_template LIKE '%subtitle%' as has_subtitle
FROM content_components
WHERE function = 'call_to_action';

-- Fix 1a: .subtitle → .subheadline
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{if .subtitle}}<p class="cta-subtitle">{{.subtitle}}</p>{{end}}',
        '{{if .subheadline}}<p class="cta-subtitle">{{.subheadline}}</p>{{end}}'
                    ),
    updated_at = now()
WHERE function = 'call_to_action';

-- Fix 1b: Replace nested primary_button with flat primary_cta
-- Old: {{if .primary_button}}
--        <a href="{{.primary_button.url}}" class="cta-btn cta-btn-primary">{{.primary_button.text}}</a>
--      {{end}}
-- New: {{if .primary_cta}}
--        <a href="{{.primary_cta_url}}" class="cta-btn cta-btn-primary">{{.primary_cta}}</a>
--      {{end}}

UPDATE content_components
SET html_template = replace(
        html_template,
        E'{{if .primary_button}}\n                <a href="{{.primary_button.url}}" class="cta-btn cta-btn-primary">{{.primary_button.text}}</a>',
        E'{{if .primary_cta}}\n                <a href="{{.primary_cta_url}}" class="cta-btn cta-btn-primary">{{.primary_cta}}</a>'
                    ),
    updated_at = now()
WHERE function = 'call_to_action';

-- Fix 1c: Replace nested secondary_button with flat secondary_cta
UPDATE content_components
SET html_template = replace(
        html_template,
        E'{{if .secondary_button}}\n                <a href="{{.secondary_button.url}}" class="cta-btn cta-btn-secondary">{{.secondary_button.text}}</a>',
        E'{{if .secondary_cta}}\n                <a href="{{.secondary_cta_url}}" class="cta-btn cta-btn-secondary">{{.secondary_cta}}</a>'
                    ),
    updated_at = now()
WHERE function = 'call_to_action';

-- Verify CTA fix
SELECT name,
       html_template LIKE '%primary_button%' as still_has_nested,
       html_template LIKE '%primary_cta%' as has_flat_cta,
       html_template LIKE '%subheadline%' as has_subheadline
FROM content_components
WHERE function = 'call_to_action';


-- =============================================================
-- 2. features / Features Grid — Fix subtitle and item .title
-- =============================================================

-- Fix 2a: .subtitle → .subheadline
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{if .subtitle}}<p class="features-subtitle">{{.subtitle}}</p>{{end}}',
        '{{if .subheadline}}<p class="features-subtitle">{{.subheadline}}</p>{{end}}'
                    ),
    updated_at = now()
WHERE function = 'features';

-- Fix 2b: Inside {{range .features}}, change .title → .name
-- This is the item-level title (not the h2 which is already .headline)
UPDATE content_components
SET html_template = replace(
        html_template,
        '<h3>{{.title}}</h3>',
        '<h3>{{.name}}</h3>'
                    ),
    updated_at = now()
WHERE function = 'features';

-- Verify features fix
SELECT name,
       html_template LIKE '%{{.name}}%' as has_dot_name,
       html_template LIKE '%<h3>{{.title}}</h3>%' as still_has_h3_title,
       html_template LIKE '%subheadline%' as has_subheadline
FROM content_components
WHERE function = 'features';


-- =============================================================
-- 3. case-studies-list — Fix .description → .summary
-- =============================================================

-- The case study items in the schema have "summary" but the
-- template renders {{.description}}
-- Also update the schema to be consistent

UPDATE content_components
SET html_template = replace(
        html_template,
        '<p>{{.description}}</p>',
        '<p>{{.summary}}</p>'
                    ),
    updated_at = now()
WHERE function = 'case-studies-list';

-- Also fix the subtitle pattern if present
UPDATE content_components
SET html_template = replace(
        html_template,
        '{{if .subtitle}}<p class="case-studies-subtitle">{{.subtitle}}</p>{{end}}',
        '{{if .subheadline}}<p class="case-studies-subtitle">{{.subheadline}}</p>{{end}}'
                    ),
    updated_at = now()
WHERE function = 'case-studies-list';

-- Update case-studies-list schema to match template field names
-- Old schema item fields: title, client, summary, link, image
-- Template uses: title ✓, client ✓, summary (now fixed) ✓, results (optional)
UPDATE content_components
SET input_schema = '{"headline": "string", "subheadline": "string", "case_studies": [{"title": "string", "client": "string", "summary": "string", "results": "string"}]}'::jsonb,
    updated_at = now()
WHERE function = 'case-studies-list';

-- Verify case-studies fix
SELECT name,
       html_template LIKE '%{{.summary}}%' as has_summary,
       html_template LIKE '%{{.description}}%' as still_has_description,
       input_schema
FROM content_components
WHERE function = 'case-studies-list';


-- =============================================================
-- 4. Duplicate cleanup check
--    Some components have two entries (e.g. "features" and
--    "Features Grid" both with function='features').
--    The WHERE function = 'xxx' catches both. Verify both fixed.
-- =============================================================

SELECT name, function,
       html_template LIKE '%primary_button%' as has_nested_btn,
       html_template LIKE '%<h3>{{.title}}</h3>%' as has_h3_title,
       html_template LIKE '%{{.subtitle}}%' as has_old_subtitle,
       html_template LIKE '%{{.description}}%' as has_old_description
FROM content_components
WHERE function IN ('call_to_action', 'features', 'case-studies-list', 'social_proof', 'testimonials')
ORDER BY function, name;


-- =============================================================
-- SUMMARY OF ALL CHANGES
-- =============================================================
--
-- call_to_action (2 rows):
--   ✓ .subtitle → .subheadline
--   ✓ .primary_button.url/.text → .primary_cta / .primary_cta_url
--   ✓ .secondary_button.url/.text → .secondary_cta / .secondary_cta_url
--   (Schema was already correct from file 51)
--
-- features (2 rows):
--   ✓ .subtitle → .subheadline
--   ✓ <h3>{{.title}}</h3> → <h3>{{.name}}</h3>
--   (Schema was already correct from file 51)
--
-- case-studies-list (1 row):
--   ✓ .subtitle → .subheadline
--   ✓ {{.description}} → {{.summary}}
--   ✓ Schema updated to match
--
-- social_proof (2 rows): No changes needed
-- testimonials (1 row): No changes needed


-- =============================================================
-- Fix CTA buttons - atomic replacements (no whitespace issues)
-- =============================================================

-- Replace the 6 individual template expressions:

-- 1. {{if .primary_button}} → {{if .primary_cta}}
UPDATE content_components
SET html_template = replace(html_template, '{{if .primary_button}}', '{{if .primary_cta}}'),
    updated_at = now()
WHERE function = 'call_to_action';

-- 2. {{.primary_button.url}} → {{.primary_cta_url}}
UPDATE content_components
SET html_template = replace(html_template, '{{.primary_button.url}}', '{{.primary_cta_url}}'),
    updated_at = now()
WHERE function = 'call_to_action';

-- 3. {{.primary_button.text}} → {{.primary_cta}}
UPDATE content_components
SET html_template = replace(html_template, '{{.primary_button.text}}', '{{.primary_cta}}'),
    updated_at = now()
WHERE function = 'call_to_action';

-- 4. {{if .secondary_button}} → {{if .secondary_cta}}
UPDATE content_components
SET html_template = replace(html_template, '{{if .secondary_button}}', '{{if .secondary_cta}}'),
    updated_at = now()
WHERE function = 'call_to_action';

-- 5. {{.secondary_button.url}} → {{.secondary_cta_url}}
UPDATE content_components
SET html_template = replace(html_template, '{{.secondary_button.url}}', '{{.secondary_cta_url}}'),
    updated_at = now()
WHERE function = 'call_to_action';

-- 6. {{.secondary_button.text}} → {{.secondary_cta}}
UPDATE content_components
SET html_template = replace(html_template, '{{.secondary_button.text}}', '{{.secondary_cta}}'),
    updated_at = now()
WHERE function = 'call_to_action';

-- Verify
SELECT name,
       html_template LIKE '%primary_button%' as still_has_nested,
       html_template LIKE '%primary_cta%' as has_flat_cta,
       html_template LIKE '%secondary_button%' as still_has_sec_nested,
       html_template LIKE '%secondary_cta%' as has_flat_sec
FROM content_components
WHERE function = 'call_to_action';


--- fix logo jpg/png
-- ============================================================================
-- FIX 5: Header templates — add logo image support
-- ============================================================================
-- Uses {{if .logo_url}} so it gracefully falls back to text when no logo image exists.
-- We update the logo <a> block and inject .logo-img CSS.

-- Fix 5a: header-professional-dark
UPDATE content_components
SET html_template = replace(
        html_template,
        '<a href="/index.html" class="logo">
                  <span class="logo-text">{{.logo_text}}</span>
              </a>',
        '<a href="/index.html" class="logo">
                  {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}
              </a>'
                    ),
    updated_at = NOW()
WHERE name = 'header-professional-dark';

-- Fix 5b: header-minimal-light
UPDATE content_components
SET html_template = replace(
        html_template,
        '<a href="/index.html" class="logo">
                  <span class="logo-text">{{.logo_text}}</span>
              </a>',
        '<a href="/index.html" class="logo">
                  {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}
              </a>'
                    ),
    updated_at = NOW()
WHERE name = 'header-minimal-light';

-- Fix 5c: header-bold-gradient
UPDATE content_components
SET html_template = replace(
        html_template,
        '<a href="/index.html" class="logo">
                  <span class="logo-text">{{.logo_text}}</span>
              </a>',
        '<a href="/index.html" class="logo">
                  {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}
              </a>'
                    ),
    updated_at = NOW()
WHERE name = 'header-bold-gradient';

-- Fix 5d: Add .logo-img CSS rule to all header components
-- Inject before .mobile-menu-toggle which exists in all header templates
UPDATE content_components
SET html_template = replace(
        html_template,
        '.mobile-menu-toggle {',
        '.logo-img {
          max-height: 40px;
          width: auto;
          display: block;
      }
      .mobile-menu-toggle {'
                    )
WHERE name IN ('header-professional-dark', 'header-minimal-light', 'header-bold-gradient')
  AND html_template NOT LIKE '%.logo-img%';

-- fix
-- Fix 5a: header-professional-dark and header-minimal-light
-- Both have the same pattern: just <span class="logo-text">
UPDATE content_components
SET html_template = replace(
        html_template,
        '<span class="logo-text">{{.logo_text}}</span>',
        '{{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}'
                    ),
    updated_at = NOW()
WHERE name IN ('header-professional-dark', 'header-minimal-light');

-- Fix 5b: header-bold-gradient
-- Has logo-icon + logo-text, wrap both in the else branch
UPDATE content_components
SET html_template = replace(
        html_template,
        '<span class="logo-icon">◆</span>
                  <span class="logo-text">{{.logo_text}}</span>',
        '{{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-icon">◆</span>
                  <span class="logo-text">{{.logo_text}}</span>{{end}}'
                    ),
    updated_at = NOW()
WHERE name = 'header-bold-gradient';

--

lucide icons

-- ============================================================================
-- Fix: Update features component templates to use Lucide icons via CDN
-- ============================================================================
-- Problem: Template renders icon names as plain text:
--   <div class="feature-icon">zap</div>
--
-- Fix: Use data-lucide attributes + Lucide CDN script
--   <div class="feature-icon"><i data-lucide="zap"></i></div>
--   + <script src="https://unpkg.com/lucide@0.460.0/dist/umd/lucide.min.js">
--
-- SQL-only fix, no Go rebuild needed.
-- There are TWO rows with function='features' — updates both.
--
-- Field names kept matching what the LLM currently generates:
--   headline, subheadline, features[].name, features[].description, features[].icon
-- ============================================================================

BEGIN;

UPDATE content_components
SET html_template = '<section class="features-section" data-component="features">
    <div class="features-container">
        <h2>{{.headline}}</h2>
        {{if .subheadline}}<p class="features-subtitle">{{.subheadline}}</p>{{end}}
        <div class="features-grid">
            {{range .features}}
            <div class="feature-item">
                {{if .icon}}<div class="feature-icon"><i data-lucide="{{.icon}}"></i></div>{{end}}
                <h3>{{.name}}</h3>
                <p>{{.description}}</p>
            </div>
            {{end}}
        </div>
    </div>
</section>
<script src="https://unpkg.com/lucide@0.460.0/dist/umd/lucide.min.js"></script>
<script>
document.addEventListener("DOMContentLoaded", function() {
    if (typeof lucide !== "undefined") {
        lucide.createIcons();
    }
});
</script>
<style>
/* Layout only - colors inherited from global CSS */
.features-section {
    padding: var(--spacing-section, 5rem 2rem);
}
.features-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.features-section h2 {
    text-align: center;
    margin-bottom: 1rem;
}
.features-subtitle {
    text-align: center;
    max-width: 600px;
    margin: 0 auto 3rem;
}
.features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
}
.feature-item {
    text-align: center;
    padding: 2rem;
}
.feature-icon {
    width: 48px;
    height: 48px;
    margin: 0 auto 1rem;
    color: var(--color-accent, #0f3460);
}
.feature-icon svg {
    width: 100%;
    height: 100%;
}
.feature-item h3 {
    margin-bottom: 1rem;
}
.feature-item p {
    margin: 0;
}
@media (max-width: 768px) {
    .features-section { padding: 3rem 1.5rem; }
    .features-grid { grid-template-columns: 1fr; gap: 1.5rem; }
}
</style>',
    updated_at = NOW()
WHERE function = 'features';

COMMIT;

-- Verify both rows updated
SELECT name, function,
       CASE WHEN html_template LIKE '%data-lucide%' THEN 'YES' ELSE 'NO' END AS has_lucide,
       CASE WHEN html_template LIKE '%lucide.createIcons%' THEN 'YES' ELSE 'NO' END AS has_script
FROM content_components
WHERE function = 'features';


-- ============================================================================
-- Fix: Lucide icons — site-wide CDN in head + data-lucide in features
-- ============================================================================
-- Two changes:
-- 1. head-seo-standard: Add Lucide CDN script + createIcons() before </head>
-- 2. features templates: Use <i data-lucide="{{.icon}}"> instead of {{.icon}}
--
-- This means ANY component can use data-lucide attributes and they'll work,
-- not just features. SQL-only, no Go rebuild.
-- ============================================================================

BEGIN;

-- -------------------------------------------------------
-- PART 1: Add Lucide to head-seo-standard
-- -------------------------------------------------------
-- Injects the CDN script before </head> if not already present.

UPDATE content_components
SET html_template = replace(
        html_template,
        '</head>',
        '    <!-- Lucide Icons -->
          <script src="https://unpkg.com/lucide@0.460.0/dist/umd/lucide.min.js"></script>
          <script>document.addEventListener("DOMContentLoaded", function() { if (typeof lucide !== "undefined") { lucide.createIcons(); } });</script>
      </head>'
                    ),
    updated_at = NOW()
WHERE name = 'head-seo-standard'
  AND html_template NOT LIKE '%lucide%';

-- -------------------------------------------------------
-- PART 2: Update features templates (both rows)
-- -------------------------------------------------------
-- Changes:
--   {{.icon}} text  →  <i data-lucide="{{.icon}}"></i>
--   .feature-icon CSS: font-size → width/height + svg sizing

UPDATE content_components
SET html_template = '<section class="features-section" data-component="features">
    <div class="features-container">
        <h2>{{.headline}}</h2>
        {{if .subheadline}}<p class="features-subtitle">{{.subheadline}}</p>{{end}}
        <div class="features-grid">
            {{range .features}}
            <div class="feature-item">
                {{if .icon}}<div class="feature-icon"><i data-lucide="{{.icon}}"></i></div>{{end}}
                <h3>{{.name}}</h3>
                <p>{{.description}}</p>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Layout only - colors inherited from global CSS */
.features-section {
    padding: var(--spacing-section, 5rem 2rem);
}
.features-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.features-section h2 {
    text-align: center;
    margin-bottom: 1rem;
}
.features-subtitle {
    text-align: center;
    max-width: 600px;
    margin: 0 auto 3rem;
}
.features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
}
.feature-item {
    text-align: center;
    padding: 2rem;
}
.feature-icon {
    width: 48px;
    height: 48px;
    margin: 0 auto 1rem;
    color: var(--color-accent, #0f3460);
}
.feature-icon svg {
    width: 100%;
    height: 100%;
}
.feature-item h3 {
    margin-bottom: 1rem;
}
.feature-item p {
    margin: 0;
}
@media (max-width: 768px) {
    .features-section { padding: 3rem 1.5rem; }
    .features-grid { grid-template-columns: 1fr; gap: 1.5rem; }
}
</style>',
    updated_at = NOW()
WHERE function = 'features';

COMMIT;

-- -------------------------------------------------------
-- Verify
-- -------------------------------------------------------
SELECT name, function,
       CASE WHEN html_template LIKE '%data-lucide%' THEN 'YES' ELSE 'NO' END AS has_lucide_attrs,
       CASE WHEN html_template LIKE '%lucide.createIcons%' THEN 'YES' ELSE 'NO' END AS has_init_script
FROM content_components
WHERE function IN ('features', 'head')
   OR name = 'head-seo-standard';


----


-- Fix 1: Update header-professional-dark to show BOTH logo image AND company name text
-- Current: either image OR text
-- Fix: show image with text beside it when logo_url is present

UPDATE content_components
SET html_template = E'<!-- HEADER SOURCE: component-db:header-professional-dark -->\n<header class="site-header site-header--dark">\n    <div class="header-container">\n        <a href="/index.html" class="logo">\n            {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{end}}\n            <span class="logo-text">{{.logo_text}}</span>\n        </a>\n        <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">\n            <span></span><span></span><span></span>\n        </button>\n        <nav class="main-nav" id="main-nav" role="navigation">\n            <ul>\n                {{if .nav_items_html}}{{.nav_items_html}}{{end}}\n            </ul>\n        </nav>\n        <a href="/contact.html" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>\n    </div>\n</header>\n<style>\n.site-header--dark {\n    background: #1a1a2e;\n    padding: 1rem 0;\n    position: sticky;\n    top: 0;\n    z-index: 1000;\n    box-shadow: 0 2px 10px rgba(0,0,0,0.1);\n}\n.header-container {\n    max-width: 1200px;\n    margin: 0 auto;\n    padding: 0 2rem;\n    display: flex;\n    align-items: center;\n    justify-content: space-between;\n    gap: 2rem;\n}\n.logo {\n    display: flex;\n    align-items: center;\n    gap: 0.75rem;\n    text-decoration: none;\n}\n.logo-text {\n    font-size: 1.25rem;\n    font-weight: 700;\n    color: #fff;\n}\n.logo-img {\n    max-height: 40px;\n    width: auto;\n    display: block;\n}\n.main-nav ul {\n    display: flex;\n    list-style: none;\n    margin: 0;\n    padding: 0;\n    gap: 2rem;\n}\n.main-nav a {\n    color: rgba(255,255,255,0.9);\n    text-decoration: none;\n    font-weight: 500;\n    padding: 0.5rem 0;\n    transition: color 0.2s;\n}\n.main-nav a:hover,\n.main-nav a.active {\n    color: #0f3460;\n}\n.header-cta {\n    background: #0f3460;\n    color: #fff;\n    padding: 0.6rem 1.25rem;\n    border-radius: 4px;\n    text-decoration: none;\n    font-weight: 500;\n    transition: opacity 0.2s;\n}\n.header-cta:hover {\n    opacity: 0.9;\n}\n.mobile-menu-toggle {\n    display: none;\n    background: none;\n    border: none;\n    cursor: pointer;\n    padding: 0.5rem;\n    flex-direction: column;\n    gap: 5px;\n}\n.mobile-menu-toggle span {\n    display: block;\n    width: 24px;\n    height: 2px;\n    background: #fff;\n    transition: transform 0.3s;\n}\n@media (max-width: 768px) {\n    .mobile-menu-toggle { display: flex; }\n    .main-nav {\n        position: absolute;\n        top: 100%;\n        left: 0;\n        right: 0;\n        background: #1a1a2e;\n        padding: 1rem 2rem;\n        display: none;\n        box-shadow: 0 4px 10px rgba(0,0,0,0.1);\n    }\n    .main-nav.active { display: block; }\n    .main-nav ul {\n        flex-direction: column;\n        gap: 0;\n    }\n    .main-nav a {\n        display: block;\n        padding: 0.75rem 0;\n        border-bottom: 1px solid rgba(255,255,255,0.1);\n    }\n    .header-cta { display: none; }\n}\n</style>\n<script>\ndocument.addEventListener("DOMContentLoaded", function() {\n    var toggle = document.querySelector(".mobile-menu-toggle");\n    var nav = document.querySelector(".main-nav");\n    if (toggle && nav) {\n        toggle.addEventListener("click", function() {\n            var expanded = toggle.getAttribute("aria-expanded") === "true";\n            toggle.setAttribute("aria-expanded", !expanded);\n            nav.classList.toggle("active");\n        });\n    }\n});\n</script>',
updated_at = NOW()
WHERE name = 'header-professional-dark';

-- Fix 2: Update social_proof to use hardcoded dark colors (not CSS variables)
-- The issue is CSS variable mismatch between head (--primary) and components (--color-primary)
-- For reliability, use explicit colors in dark sections

UPDATE content_components
SET html_template = E'<section class="social-proof-section" data-component="social-proof">\n    <div class="social-proof-container">\n        <h2>{{.headline}}</h2>\n        <div class="testimonials-grid">\n            {{range .testimonials}}\n            <div class="testimonial-item">\n                <blockquote>{{.quote}}</blockquote>\n                <cite>\n                    <strong>{{.author}}</strong>\n                    {{if .role}}<span>{{.role}}</span>{{end}}\n                </cite>\n            </div>\n            {{end}}\n        </div>\n    </div>\n</section>\n<style>\n/* Dark section with explicit colors for reliability */\n.social-proof-section {\n    padding: 5rem 2rem;\n    background: #1a1a2e;\n    color: #fff;\n}\n.social-proof-container {\n    max-width: 1200px;\n    margin: 0 auto;\n}\n.social-proof-section h2 {\n    text-align: center;\n    margin-bottom: 3rem;\n    color: #fff;\n}\n.testimonials-grid {\n    display: grid;\n    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));\n    gap: 2rem;\n}\n.testimonial-item {\n    padding: 2rem;\n    background: rgba(255,255,255,0.05);\n    border-radius: 8px;\n    border-left: 3px solid #0f3460;\n}\n.testimonial-item blockquote {\n    font-size: 1.1rem;\n    line-height: 1.7;\n    margin: 0 0 1.5rem;\n    font-style: italic;\n    color: rgba(255,255,255,0.9);\n}\n.testimonial-item cite {\n    display: block;\n    font-style: normal;\n}\n.testimonial-item cite strong {\n    display: block;\n    color: #fff;\n}\n.testimonial-item cite span {\n    font-size: 0.9rem;\n    color: rgba(255,255,255,0.7);\n}\n@media (max-width: 768px) {\n    .social-proof-section { padding: 3rem 1.5rem; }\n    .testimonials-grid { grid-template-columns: 1fr; }\n}\n</style>',
updated_at = NOW()
WHERE name = 'Social Proof';

-- Also update the testimonials component (duplicate of social_proof)
UPDATE content_components
SET html_template = E'<section class="social-proof-section" data-component="social-proof">\n    <div class="social-proof-container">\n        <h2>{{.headline}}</h2>\n        <div class="testimonials-grid">\n            {{range .testimonials}}\n            <div class="testimonial-item">\n                <blockquote>{{.quote}}</blockquote>\n                <cite>\n                    <strong>{{.author}}</strong>\n                    {{if .role}}<span>{{.role}}</span>{{end}}\n                </cite>\n            </div>\n            {{end}}\n        </div>\n    </div>\n</section>\n<style>\n/* Dark section with explicit colors for reliability */\n.social-proof-section {\n    padding: 5rem 2rem;\n    background: #1a1a2e;\n    color: #fff;\n}\n.social-proof-container {\n    max-width: 1200px;\n    margin: 0 auto;\n}\n.social-proof-section h2 {\n    text-align: center;\n    margin-bottom: 3rem;\n    color: #fff;\n}\n.testimonials-grid {\n    display: grid;\n    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));\n    gap: 2rem;\n}\n.testimonial-item {\n    padding: 2rem;\n    background: rgba(255,255,255,0.05);\n    border-radius: 8px;\n    border-left: 3px solid #0f3460;\n}\n.testimonial-item blockquote {\n    font-size: 1.1rem;\n    line-height: 1.7;\n    margin: 0 0 1.5rem;\n    font-style: italic;\n    color: rgba(255,255,255,0.9);\n}\n.testimonial-item cite {\n    display: block;\n    font-style: normal;\n}\n.testimonial-item cite strong {\n    display: block;\n    color: #fff;\n}\n.testimonial-item cite span {\n    font-size: 0.9rem;\n    color: rgba(255,255,255,0.7);\n}\n@media (max-width: 768px) {\n    .social-proof-section { padding: 3rem 1.5rem; }\n    .testimonials-grid { grid-template-columns: 1fr; }\n}\n</style>',
updated_at = NOW()
WHERE name = 'testimonials';


-- Fix 3: Add styling to about-content component
-- Currently missing <style> block for container layout

UPDATE content_components
SET html_template = E'<section class="about-content-section" data-component="about-content">\n    <div class="about-container">\n        {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}\n        <div class="about-text">\n            {{.content}}\n        </div>\n        {{if .highlights}}\n        <div class="about-highlights">\n            {{range .highlights}}\n            <div class="highlight-item">\n                <h3>{{.title}}</h3>\n                <p>{{.description}}</p>\n            </div>\n            {{end}}\n        </div>\n        {{end}}\n    </div>\n</section>\n<style>\n.about-content-section {\n    padding: 5rem 2rem;\n    background: #fff;\n}\n.about-container {\n    max-width: 1000px;\n    margin: 0 auto;\n}\n.about-content-section h2 {\n    font-size: clamp(1.75rem, 4vw, 2.5rem);\n    margin-bottom: 2rem;\n    color: #1a1a2e;\n    text-align: center;\n}\n.about-text {\n    font-size: 1.1rem;\n    line-height: 1.8;\n    color: #333;\n    margin-bottom: 3rem;\n}\n.about-text p {\n    margin-bottom: 1.5rem;\n}\n.about-highlights {\n    display: grid;\n    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));\n    gap: 2rem;\n    margin-top: 3rem;\n}\n.highlight-item {\n    padding: 1.5rem;\n    background: #f8f9fa;\n    border-radius: 8px;\n    border-left: 4px solid #0f3460;\n}\n.highlight-item h3 {\n    font-size: 1.25rem;\n    margin-bottom: 0.75rem;\n    color: #1a1a2e;\n}\n.highlight-item p {\n    color: #555;\n    line-height: 1.6;\n    margin: 0;\n}\n@media (max-width: 768px) {\n    .about-content-section { padding: 3rem 1.5rem; }\n    .about-highlights { grid-template-columns: 1fr; }\n}\n</style>',
updated_at = NOW()
WHERE name = 'about-content';


-- Fix 4: Add styling to leadership-team component
-- Currently missing <style> block for container layout

UPDATE content_components
SET html_template = E'<section class="team-section" data-component="leadership-team">\n    <div class="team-container">\n        {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}\n        {{if .section_intro}}<p class="section-intro">{{.section_intro}}</p>{{end}}\n        <div class="team-grid">\n            {{range .members}}\n            <div class="team-member">\n                {{if .photo}}<img src="{{.photo}}" alt="{{.name}}" class="member-photo">{{end}}\n                <h3>{{.name}}</h3>\n                <p class="member-title">{{.title}}</p>\n                {{if .bio}}<p class="member-bio">{{.bio}}</p>{{end}}\n            </div>\n            {{end}}\n        </div>\n    </div>\n</section>\n<style>\n.team-section {\n    padding: 5rem 2rem;\n    background: #f8f9fa;\n}\n.team-container {\n    max-width: 1200px;\n    margin: 0 auto;\n}\n.team-section h2 {\n    font-size: clamp(1.75rem, 4vw, 2.5rem);\n    margin-bottom: 1rem;\n    color: #1a1a2e;\n    text-align: center;\n}\n.section-intro {\n    text-align: center;\n    max-width: 700px;\n    margin: 0 auto 3rem;\n    color: #555;\n    line-height: 1.7;\n}\n.team-grid {\n    display: grid;\n    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));\n    gap: 2rem;\n}\n.team-member {\n    background: #fff;\n    padding: 2rem;\n    border-radius: 8px;\n    text-align: center;\n    box-shadow: 0 2px 8px rgba(0,0,0,0.06);\n}\n.member-photo {\n    width: 120px;\n    height: 120px;\n    border-radius: 50%;\n    object-fit: cover;\n    margin-bottom: 1.5rem;\n    background: #e0e0e0;\n}\n.team-member h3 {\n    font-size: 1.25rem;\n    margin-bottom: 0.5rem;\n    color: #1a1a2e;\n}\n.member-title {\n    color: #0f3460;\n    font-weight: 500;\n    margin-bottom: 1rem;\n    font-size: 0.95rem;\n}\n.member-bio {\n    color: #555;\n    line-height: 1.6;\n    font-size: 0.95rem;\n    text-align: left;\n}\n@media (max-width: 768px) {\n    .team-section { padding: 3rem 1.5rem; }\n    .team-grid { grid-template-columns: 1fr; }\n}\n</style>',
updated_at = NOW()
WHERE name = 'leadership-team';





-- =============================================================================
-- VERIFICATION QUERIES (run after updates)
-- =============================================================================

-- Verify header now shows both logo and text:
SELECT name,
       CASE
           WHEN html_template LIKE '%{{if .logo_url}}%logo-img%{{end}}%logo-text%' THEN 'FIXED: Shows both logo and text'
           WHEN html_template LIKE '%{{if .logo_url}}%{{else}}%{{end}}%' THEN 'OLD: Either/or pattern'
           ELSE 'CHECK MANUALLY'
           END as header_status
FROM content_components
WHERE name = 'header-professional-dark';

-- Verify social proof has explicit dark colors:
SELECT name,
       CASE
           WHEN html_template LIKE '%background: #1a1a2e%' THEN 'FIXED: Explicit dark colors'
           WHEN html_template LIKE '%var(--color-primary%' THEN 'OLD: Uses CSS variables (may have mismatch)'
           ELSE 'CHECK MANUALLY'
           END as social_proof_status
FROM content_components
WHERE name IN ('Social Proof', 'testimonials');

-- Verify about-content has styling:
SELECT name,
       CASE
           WHEN html_template LIKE '%<style>%.about-content-section%' THEN 'FIXED: Has styling'
           ELSE 'MISSING: No style block'
           END as about_status
FROM content_components
WHERE name = 'about-content';

-- Verify leadership-team has styling:
SELECT name,
       CASE
           WHEN html_template LIKE '%<style>%.team-section%' THEN 'FIXED: Has styling'
           ELSE 'MISSING: No style block'
           END as team_status
FROM content_components
WHERE name = 'leadership-team';


-- =============================================================================
-- NOTE: Logo Text Issue
-- =============================================================================
-- The user mentioned that AI-generated logos shouldn't have text in them because
-- the text gets garbled. This is controlled by the site-planner prompt that
-- generates the image_prompts.logo field.
--
-- To fix this, update the site-planner workflow prompt to explicitly say:
-- "Create a logo WITHOUT ANY TEXT - use only symbols, icons, or abstract shapes.
--  The company name will be displayed separately as HTML text next to the logo."
--
-- The site-planner is in agent_definitions WHERE agent_type = 'site-planner'.
-- Look in config->workflow->steps for the prompt_template that generates image_prompts.
--
-- =============================================================================


-- =============================================================================
-- SUMMARY OF CHANGES
-- =============================================================================
-- 1. header-professional-dark: Now shows logo image AND company name text together
-- 2. social_proof & testimonials: Use explicit #1a1a2e dark background instead of CSS variables
-- 3. about-content: Added full <style> block with centered container layout
-- 4. leadership-team: Added full <style> block with grid layout
-- 5. pageflow-builder: Changed image prompt handling from template syntax to input_mapping

-- granular editing
-- ============================================================================
-- 1. NEW COMPONENT: portfolio-showcase
-- Replaces social-proof/testimonials for sites where you don't have
-- client testimonials. Shows actual domains/sites built by the platform.
-- ============================================================================

INSERT INTO content_components (
    id, name, description, html_template, input_schema, function,
    display_name, category, semantic_tags, render_mode,
    component_level, is_active
) VALUES (
             gen_random_uuid(),
             'portfolio-showcase',
             'Showcase of sites and tools built by the platform. Honest portfolio — no fabricated testimonials, just the actual work. Each project shows domain, description, what it was built with, and build time.',
             E'<section class="portfolio-showcase-section" data-component="portfolio-showcase">\n    <div class="portfolio-container">\n        <h2>{{.headline}}</h2>\n        {{if .intro}}<p class="portfolio-intro">{{.intro}}</p>{{end}}\n        <div class="portfolio-grid">\n            {{range .projects}}\n            <div class="portfolio-item">\n                <div class="portfolio-item-header">\n                    <h3>{{.title}}</h3>\n                    {{if .live_url}}<a href="{{.live_url}}" class="portfolio-link" target="_blank" rel="noopener">Visit Site \\2192</a>{{end}}\n                </div>\n                {{if .domain}}<p class="portfolio-domain">{{.domain}}</p>{{end}}\n                <p class="portfolio-description">{{.description}}</p>\n                <div class="portfolio-meta">\n                    {{if .built_with}}<span class="portfolio-tag">{{.built_with}}</span>{{end}}\n                    {{if .build_time}}<span class="portfolio-tag portfolio-tag-time">{{.build_time}}</span>{{end}}\n                </div>\n            </div>\n            {{end}}\n        </div>\n    </div>\n</section>\n<style>\n/* Dark section to match the visual weight of the social-proof section it replaces */\n.portfolio-showcase-section {\n    padding: var(--spacing-section, 5rem 2rem);\n    background: var(--color-primary, #1a1a2e);\n    color: var(--color-white, #fff);\n}\n.portfolio-container {\n    max-width: var(--container-max-width, 1200px);\n    margin: 0 auto;\n}\n.portfolio-showcase-section h2 {\n    text-align: center;\n    margin-bottom: 1rem;\n    color: var(--color-white, #fff);\n}\n.portfolio-intro {\n    text-align: center;\n    max-width: 700px;\n    margin: 0 auto 3rem;\n    color: rgba(255,255,255,0.8);\n    line-height: 1.7;\n}\n.portfolio-grid {\n    display: grid;\n    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));\n    gap: 2rem;\n}\n.portfolio-item {\n    padding: 2rem;\n    background: rgba(255,255,255,0.05);\n    border-radius: 8px;\n    border-left: 3px solid var(--color-accent, #0f3460);\n    transition: transform 0.2s, box-shadow 0.2s;\n}\n.portfolio-item:hover {\n    transform: translateY(-2px);\n    box-shadow: 0 8px 24px rgba(0,0,0,0.2);\n}\n.portfolio-item-header {\n    display: flex;\n    justify-content: space-between;\n    align-items: flex-start;\n    gap: 1rem;\n    margin-bottom: 0.5rem;\n}\n.portfolio-item h3 {\n    margin: 0;\n    font-size: 1.2rem;\n    color: var(--color-white, #fff);\n}\n.portfolio-link {\n    color: var(--color-accent, #4da6ff);\n    text-decoration: none;\n    font-size: 0.85rem;\n    font-weight: 500;\n    white-space: nowrap;\n    transition: color 0.2s;\n}\n.portfolio-link:hover {\n    color: var(--color-white, #fff);\n}\n.portfolio-domain {\n    font-family: monospace;\n    font-size: 0.85rem;\n    color: rgba(255,255,255,0.5);\n    margin-bottom: 1rem;\n}\n.portfolio-description {\n    color: rgba(255,255,255,0.85);\n    line-height: 1.7;\n    margin-bottom: 1.5rem;\n}\n.portfolio-meta {\n    display: flex;\n    gap: 0.75rem;\n    flex-wrap: wrap;\n}\n.portfolio-tag {\n    display: inline-block;\n    padding: 0.25rem 0.75rem;\n    background: rgba(255,255,255,0.1);\n    border-radius: 4px;\n    font-size: 0.8rem;\n    color: rgba(255,255,255,0.7);\n}\n.portfolio-tag-time {\n    background: rgba(79, 166, 255, 0.15);\n    color: var(--color-accent, #4da6ff);\n}\n@media (max-width: 768px) {\n    .portfolio-showcase-section { padding: 3rem 1.5rem; }\n    .portfolio-grid { grid-template-columns: 1fr; }\n    .portfolio-item-header { flex-direction: column; gap: 0.5rem; }\n}\n</style>',
             '{"headline": "string", "intro": "string", "projects": [{"title": "string", "domain": "string", "description": "string", "built_with": "string", "build_time": "string", "live_url": "string"}]}',
             'portfolio-showcase',
             'Portfolio Showcase',
             'social-proof',
             '["portfolio", "showcase", "work", "projects", "built-with"]',
             'template',
             'section',
             true
         );


-- ============================================================================
-- 2. CONTENT DIRECTION COLUMN
-- Per-page instructions for content generation. Flows through to
-- content-writer prompt when present.
-- ============================================================================

ALTER TABLE pages ADD COLUMN IF NOT EXISTS content_direction JSONB;

COMMENT ON COLUMN pages.content_direction IS
    'Optional per-page content direction for rebuilds. Passed to content-writer prompt. '
    'Structure: { "instruction": "...", "format": "...", "examples": [...], "avoid": [...] }';


-- ============================================================================
-- 3. VERIFY
-- ============================================================================

SELECT id, name, display_name, function, category
FROM content_components
WHERE name = 'portfolio-showcase';

SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'pages' AND column_name = 'content_direction';


--

-- component naming standardisation

-- Component Naming Standardization
-- =================================
--
-- Contract:
--   1. content_components.function is the canonical identifier
--   2. Always kebab-case (hyphens): "social-proof" not "social_proof"
--   3. Must be unique across active components
--   4. data-component attribute in html_template MUST equal function
--   5. page_components.slot_name stores function value
--
-- Naming pattern:
--   General:       {purpose}            → hero, social-proof, call-to-action
--   Page-specific: {page}-{purpose}     → about-hero, services-hero, case-studies-hero
--   Site-level:    {slot}-{variant}     → header-professional-dark, footer-4-column
--
-- This migration:
--   Step 1: Fix underscore → hyphen in function column
--   Step 2: Fix shared function values (hero variants get unique functions)
--   Step 3: Sync data-component attributes in templates to match function
--   Step 4: Add DB constraint to enforce kebab-case and uniqueness
--   Step 5: Update page_components.slot_name to match new function values

-- ============================================================================
-- Step 1: Fix underscore functions → kebab-case
-- ============================================================================

-- social_proof → social-proof
UPDATE content_components
SET function = 'social-proof'
WHERE function = 'social_proof' AND is_active = true;

-- call_to_action → call-to-action
UPDATE content_components
SET function = 'call-to-action'
WHERE function = 'call_to_action' AND is_active = true;

-- featured_content → featured-content (if exists)
UPDATE content_components
SET function = 'featured-content'
WHERE function = 'featured_content' AND is_active = true;

-- Any other underscore functions we missed
UPDATE content_components
SET function = replace(function, '_', '-')
WHERE function LIKE '%\_%' AND is_active = true;

-- ============================================================================
-- Step 2: Fix shared hero functions — each variant gets unique function
-- ============================================================================

-- These already have the right data-component values in their templates,
-- they just need function to match.

UPDATE content_components
SET function = 'about-hero'
WHERE function = 'hero'
  AND name = 'About Page Hero'
  AND is_active = true;

UPDATE content_components
SET function = 'services-hero'
WHERE function = 'hero'
  AND name = 'Services Page Hero'
  AND is_active = true;

UPDATE content_components
SET function = 'contact-hero'
WHERE function = 'hero'
  AND name = 'Contact Page Hero'
  AND is_active = true;

UPDATE content_components
SET function = 'case-studies-hero'
WHERE function = 'hero'
  AND name = 'Case Studies Hero'
  AND is_active = true;

-- The generic hero (homepage hero) keeps function = 'hero'
-- Verify only one 'hero' remains
SELECT id, name, function
FROM content_components
WHERE function = 'hero' AND is_active = true;

-- ============================================================================
-- Step 3: Fix testimonials function → match its data-component
-- ============================================================================

-- The component named "Testimonials Section" has function "testimonials"
-- but its template has data-component="social-proof".
-- Two options:
--   a) Change function to "social-proof" (matches template)
--   b) Change data-component to "testimonials" (matches function)
--
-- Option a) is wrong because we already have social-proof components.
-- Option b) is right: this is a testimonials component, give it its own identity.

UPDATE content_components
SET html_template = replace(html_template, 'data-component="social-proof"', 'data-component="testimonials"')
WHERE function = 'testimonials'
  AND is_active = true
  AND html_template LIKE '%data-component="social-proof"%';

-- ============================================================================
-- Step 3b: Sync data-component attributes to match function for all components
-- ============================================================================

-- For any component where data-component doesn't match function,
-- update the template. This is safe because we've already standardized function.
DO $$
DECLARE
rec RECORD;
    old_attr TEXT;
    new_attr TEXT;
BEGIN
FOR rec IN
SELECT id, function, name,
       substring(html_template from 'data-component="([^"]+)"') as current_data_component,
       html_template
FROM content_components
WHERE is_active = true
  AND html_template LIKE '%data-component="%'
  AND substring(html_template from 'data-component="([^"]+)"') != function
    LOOP
        old_attr := 'data-component="' || rec.current_data_component || '"';
new_attr := 'data-component="' || rec.function || '"';

UPDATE content_components
SET html_template = replace(html_template, old_attr, new_attr)
WHERE id = rec.id;

RAISE NOTICE 'Fixed data-component: % → % (component: %)',
            rec.current_data_component, rec.function, rec.name;
END LOOP;
END $$;

-- ============================================================================
-- Step 4: Add DB constraints
-- ============================================================================

-- 4a: Kebab-case format check — only lowercase letters, digits, and hyphens
-- Allows empty string for legacy components without a function
ALTER TABLE content_components
DROP CONSTRAINT IF EXISTS chk_function_kebab_case;

ALTER TABLE content_components
    ADD CONSTRAINT chk_function_kebab_case
        CHECK (
            function = ''
                OR function ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$'
    );

-- 4b: Unique function across active components
-- Uses a partial unique index (only active components must be unique)
DROP INDEX IF EXISTS idx_content_components_unique_active_function;

CREATE UNIQUE INDEX idx_content_components_unique_active_function
    ON content_components (function)
    WHERE is_active = true AND function != '';

-- Note: site-level components (headers, footers, heads) use "site-header",
-- "site-footer", "head" as function but are differentiated by name/id.
-- If multiple active headers are needed, their function could be:
--   header-professional-dark, header-minimal-light, header-bold-gradient
-- rather than all sharing "site-header".
--
-- Check if we need to handle this:
SELECT function, count(*) as cnt
FROM content_components
WHERE is_active = true AND function != ''
GROUP BY function
HAVING count(*) > 1
ORDER BY cnt DESC;

-- If the above shows duplicates in site-header/site-footer/head,
-- update those to use their specific variant names:
UPDATE content_components
SET function = 'header-professional-dark'
WHERE function = 'site-header' AND name = 'Professional Dark Header' AND is_active = true;

UPDATE content_components
SET function = 'header-minimal-light'
WHERE function = 'site-header' AND name = 'Minimal Light Header' AND is_active = true;

UPDATE content_components
SET function = 'header-bold-gradient'
WHERE function = 'site-header' AND name = 'Bold Gradient Header' AND is_active = true;

UPDATE content_components
SET function = 'footer-4-column'
WHERE function = 'site-footer' AND name = '4-Column Footer' AND is_active = true;

UPDATE content_components
SET function = 'footer-standard'
WHERE function = 'site-footer' AND name = 'Standard Footer' AND is_active = true;

UPDATE content_components
SET function = 'footer-simple'
WHERE function = 'site-footer' AND name = 'Simple Footer' AND is_active = true;

UPDATE content_components
SET function = 'head-seo-standard'
WHERE function = 'head' AND name = 'Standard SEO Head' AND is_active = true;

-- ============================================================================
-- Step 5: Update page_components.slot_name to match new function values
-- ============================================================================

-- Fix underscore slot_names
UPDATE page_components
SET slot_name = replace(slot_name, '_', '-')
WHERE slot_name LIKE '%\_%';

-- Fix any slot_names that should match new function values
-- (This catches cases where slot_name was set from data-component
-- which was already correct, so most should already be fine)

-- ============================================================================
-- Step 6: Verification
-- ============================================================================

-- All active components: function should be kebab-case and unique
SELECT function, name,
       CASE WHEN function ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$' THEN 'OK' ELSE 'BAD' END as format_check,
       substring(html_template from 'data-component="([^"]+)"') as data_component,
       CASE
           WHEN html_template NOT LIKE '%data-component%' THEN 'N/A'
           WHEN substring(html_template from 'data-component="([^"]+)"') = function THEN 'MATCH'
           ELSE 'MISMATCH'
           END as attr_check
FROM content_components
WHERE is_active = true
ORDER BY function;

--

-- 1. Create use-cases-hero component (function: hero-use-cases)
INSERT INTO content_components (name, description, function, category, display_name,
                                html_template, input_schema, render_mode, component_level)
VALUES (
           'use-cases-hero',
           'Hero section for the use cases page',
           'hero-use-cases',
           'use-cases',
           'Use Cases Hero',
           '<section class="hero hero-use-cases" data-component="hero">
                 <div class="hero-content">
                     <h1>{{.headline}}</h1>
                     <p class="hero-subheadline">{{.subheadline}}</p>
                 </div>
             </section>
         <style>
         .hero-use-cases {
             min-height: 50vh;
             display: flex;
             align-items: center;
             justify-content: center;
             text-align: center;
             padding: 4rem 2rem;
             background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
         }
         .hero-use-cases .hero-content {
             max-width: 800px;
             margin: 0 auto;
             color: #fff;
         }
         .hero-use-cases h1 {
             font-size: clamp(1.75rem, 4vw, 2.75rem);
             font-weight: 700;
             margin-bottom: 1rem;
             line-height: 1.2;
             color: #fff;
         }
         .hero-use-cases .hero-subheadline {
             font-size: clamp(1rem, 2vw, 1.2rem);
             opacity: 0.9;
             line-height: 1.6;
             color: rgba(255,255,255,0.9);
         }
         </style>',
           '{"headline": "string", "subheadline": "string"}'::jsonb,
           'template',
           'section'
       );

-- 2. Create use-cases-list component (function: use-cases-list)
INSERT INTO content_components (name, description, function, category, display_name,
                                html_template, input_schema, render_mode, component_level)
VALUES (
           'use-cases-list',
           'Grid of use case items with title, client, summary and results',
           'use-cases-list',
           'use-cases',
           'Use Cases List',
           '<section class="use-cases-section" data-component="use-cases-list">
             <div class="use-cases-container">
                 <h2>{{.headline}}</h2>
                 {{if .subheadline}}<p class="use-cases-subtitle">{{.subheadline}}</p>{{end}}
                 <div class="use-cases-grid">
                     {{range .use_cases}}
                     <article class="use-case-item">
                         <h3>{{.title}}</h3>
                         <p class="use-case-client">{{.client}}</p>
                         <p>{{.summary}}</p>
                         {{if .results}}<p class="use-case-results"><strong>Results:</strong> {{.results}}</p>{{end}}
                     </article>
                     {{end}}
                 </div>
             </div>
         </section>
         <style>
         .use-cases-section {
             padding: var(--spacing-section, 5rem 2rem);
             background: var(--color-surface, #f8f9fa);
         }
         .use-cases-container {
             max-width: var(--container-max-width, 1200px);
             margin: 0 auto;
         }
         .use-cases-section h2 {
             text-align: center;
             margin-bottom: 1rem;
         }
         .use-cases-subtitle {
             text-align: center;
             max-width: 600px;
             margin: 0 auto 3rem;
         }
         .use-cases-grid {
             display: grid;
             grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
             gap: 2rem;
         }
         .use-case-item {
             padding: 2rem;
             background: var(--color-background, #fff);
             border-radius: 8px;
             box-shadow: 0 2px 8px rgba(0,0,0,0.06);
         }
         .use-case-item h3 {
             margin-bottom: 0.5rem;
         }
         .use-case-client {
             font-size: 0.9rem;
             margin-bottom: 1rem;
             color: var(--color-text-muted, #666);
         }
         .use-case-results {
             margin-top: 1rem;
             padding-top: 1rem;
             border-top: 1px solid var(--color-border, #e2e8f0);
         }
         @media (max-width: 768px) {
             .use-cases-section { padding: 3rem 1.5rem; }
             .use-cases-grid { grid-template-columns: 1fr; }
         }
         </style>',
           '{"headline": "string", "subheadline": "string", "use_cases": [{"title": "string", "client": "string", "results": "string", "summary": "string"}]}'::jsonb,
           'template',
           'section'
       );

-- 3. Update pages.sections to match the new function names
-- The plan says ["use-cases-hero", "use-cases-list", "call_to_action"]
-- but the function names are hero-use-cases and use-cases-list
-- Check what NormalizeComponentFunction does — it may handle this.
-- For now, update sections to use the function names directly:
UPDATE pages
SET sections = '["hero-use-cases", "use-cases-list", "call_to_action"]'::jsonb,
    build_status = 'needs_rebuild',
    updated_at = NOW()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
  AND name = 'use-cases';

-- 4. Verify
SELECT name, function, category FROM content_components
WHERE category = 'use-cases' ORDER BY function;


-- light and dark css components - styles.css has base default styles, inline can override
    -- also updated contract with hyphens not underscores

-- ============================================================================
-- 014 SECTION CONTEXT VARIABLE MIGRATION
--
-- Follows 042_component_naming_contract.md — all function references are
-- kebab-case as stored in content_components.function.
--
-- 1. Adds is_dark_section column to content_components
-- 2. Flags all dark-background section components
-- 3. Updates dark component templates with --section-* CSS variables
-- 4. Verification queries
--
-- Safe to re-run (IF NOT EXISTS, idempotent updates).
-- ============================================================================

-- ============================================================================
-- STEP 1: Add enforcement column
-- ============================================================================

ALTER TABLE content_components ADD COLUMN IF NOT EXISTS is_dark_section boolean DEFAULT false;

COMMENT ON COLUMN content_components.is_dark_section IS
  'True if component has dark background. MUST set --section-text, --section-text-muted, --section-heading, --section-surface, --section-border on container.';


-- ============================================================================
-- STEP 2: Flag ALL genuinely dark-background section components
-- NOTE: about-content, differentiators, leadership-team, contact-info are
-- NOT dark — they matched the #1a1a2e query because they use it as text color.
-- ============================================================================

UPDATE content_components
SET is_dark_section = true, updated_at = NOW()
WHERE function IN (
                   'social-proof',
                   'testimonials',
                   'call-to-action',
                   'hero',
                   'hero-about',
                   'hero-services',
                   'hero-contact',
                   'hero-case-studies',
                   'hero-use-cases',
                   'portfolio-showcase'
    )
  AND component_level = 'section';


-- ============================================================================
-- STEP 3: Update social-proof (function = 'social-proof')
-- ============================================================================

UPDATE content_components
SET html_template = '<section class="social-proof-section" data-component="social-proof">
    <div class="social-proof-container">
        <h2>{{.headline}}</h2>
        <div class="testimonials-grid">
            {{range .testimonials}}
            <div class="testimonial-item">
                <blockquote>{{.quote}}</blockquote>
                <cite>
                    <strong>{{.author}}</strong>
                    {{if .role}}<span>{{.role}}</span>{{end}}
                </cite>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
.social-proof-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-primary, #1a1a2e);
    color: var(--color-white, #fff);

    /* Dark section context — children adapt via --section-* variables */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
}
.social-proof-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.social-proof-section h2 {
    text-align: center;
    margin-bottom: 3rem;
}
.testimonials-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}
.testimonial-item {
    padding: 2rem;
    background: var(--section-surface);
    border-radius: 8px;
    border-left: 3px solid var(--section-border);
}
.testimonial-item blockquote {
    font-size: 1.1rem;
    line-height: 1.7;
    margin: 0 0 1.5rem;
}
.testimonial-item cite {
    display: block;
}
.testimonial-item cite strong {
    display: block;
}
.testimonial-item cite span {
    font-size: 0.9rem;
}
@media (max-width: 768px) {
    .social-proof-section { padding: 3rem 1.5rem; }
    .testimonials-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'social-proof';


-- ============================================================================
-- STEP 4: Update testimonials (function = 'testimonials')
-- data-component="testimonials" per naming contract
-- ============================================================================

UPDATE content_components
SET html_template = '<section class="social-proof-section" data-component="testimonials">
    <div class="social-proof-container">
        <h2>{{.headline}}</h2>
        <div class="testimonials-grid">
            {{range .testimonials}}
            <div class="testimonial-item">
                <blockquote>{{.quote}}</blockquote>
                <cite>
                    <strong>{{.author}}</strong>
                    {{if .role}}<span>{{.role}}</span>{{end}}
                </cite>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
.social-proof-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-primary, #1a1a2e);
    color: var(--color-white, #fff);

    /* Dark section context */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
}
.social-proof-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.social-proof-section h2 {
    text-align: center;
    margin-bottom: 3rem;
}
.testimonials-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}
.testimonial-item {
    padding: 2rem;
    background: var(--section-surface);
    border-radius: 8px;
    border-left: 3px solid var(--section-border);
}
.testimonial-item blockquote {
    font-size: 1.1rem;
    line-height: 1.7;
    margin: 0 0 1.5rem;
}
.testimonial-item cite {
    display: block;
}
.testimonial-item cite strong {
    display: block;
}
.testimonial-item cite span {
    font-size: 0.9rem;
}
@media (max-width: 768px) {
    .social-proof-section { padding: 3rem 1.5rem; }
    .testimonials-grid { grid-template-columns: 1fr; }
}
</style>',
    updated_at = NOW()
WHERE function = 'testimonials';


-- ============================================================================
-- STEP 5: Update call-to-action (function = 'call-to-action')
-- ============================================================================

UPDATE content_components
SET html_template = '<section class="cta-section" data-component="call-to-action">
    <div class="cta-container">
        <h2>{{.headline}}</h2>
        {{if .subheadline}}<p class="cta-subtitle">{{.subheadline}}</p>{{end}}
        <div class="cta-buttons">
            {{if .primary_cta}}
            <a href="{{.primary_cta_url}}" class="cta-btn cta-btn-primary">{{.primary_cta}}</a>
            {{end}}
            {{if .secondary_cta}}
            <a href="{{.secondary_cta_url}}" class="cta-btn cta-btn-secondary">{{.secondary_cta}}</a>
            {{end}}
        </div>
    </div>
</section>
<style>
.cta-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-primary, #1a1a2e);
    color: var(--color-white, #fff);
    text-align: center;

    /* Dark section context */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.85);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
}
.cta-container {
    max-width: 800px;
    margin: 0 auto;
}
.cta-section h2 {
    margin-bottom: 1rem;
}
.cta-subtitle {
    margin-bottom: 2rem;
}
.cta-buttons {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
}
.cta-btn {
    display: inline-block;
    padding: 1rem 2rem;
    border-radius: 6px;
    text-decoration: none;
    font-weight: 600;
    transition: transform 0.2s, box-shadow 0.2s;
}
.cta-btn:hover {
    transform: translateY(-2px);
}
.cta-btn-primary {
    background: var(--color-white, #fff);
    color: var(--color-primary, #1a1a2e);
}
.cta-btn-secondary {
    background: transparent;
    border: 2px solid var(--color-white, #fff);
    color: var(--color-white, #fff);
}
@media (max-width: 768px) {
    .cta-section { padding: 3rem 1.5rem; }
    .cta-buttons { flex-direction: column; align-items: center; }
    .cta-btn { width: 100%; max-width: 280px; text-align: center; }
}
</style>',
    updated_at = NOW()
WHERE function = 'call-to-action';


-- ============================================================================
-- STEP 6: Update hero variants
-- All data-component attributes match their function name exactly
-- ============================================================================

-- 6a. hero (function = 'hero', data-component="hero")
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

    /* Dark section context */
    --section-text: rgba(255,255,255,0.95);
    --section-text-muted: rgba(255,255,255,0.8);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.1);
    --section-border: rgba(255,255,255,0.3);
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
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.35rem);
    margin-bottom: 2rem;
    line-height: 1.6;
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
WHERE function = 'hero';

-- 6b. hero-about (data-component="hero-about")
UPDATE content_components
SET html_template = '<section class="hero hero-about" data-component="hero-about">
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

    /* Dark section context */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
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
}
.hero-about .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
</style>',
    updated_at = NOW()
WHERE function = 'hero-about';

-- 6c. hero-services (data-component="hero-services")
UPDATE content_components
SET html_template = '<section class="hero hero-services" data-component="hero-services">
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

    /* Dark section context */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
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
}
.hero-services .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
</style>',
    updated_at = NOW()
WHERE function = 'hero-services';

-- 6d. hero-contact (data-component="hero-contact")
UPDATE content_components
SET html_template = '<section class="hero hero-contact" data-component="hero-contact">
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

    /* Dark section context */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
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
}
.hero-contact .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
</style>',
    updated_at = NOW()
WHERE function = 'hero-contact';

-- 6e. hero-case-studies (data-component="hero-case-studies")
UPDATE content_components
SET html_template = '<section class="hero hero-case-studies" data-component="hero-case-studies">
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

    /* Dark section context */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
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
}
.hero-case-studies .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
</style>',
    updated_at = NOW()
WHERE function = 'hero-case-studies';

-- 6f. hero-use-cases (data-component="hero-use-cases")
UPDATE content_components
SET html_template = '<section class="hero hero-use-cases" data-component="hero-use-cases">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-use-cases {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);

    /* Dark section context */
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
}
.hero-use-cases .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: #fff;
}
.hero-use-cases h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
}
.hero-use-cases .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
</style>',
    updated_at = NOW()
WHERE function = 'hero-use-cases';


-- ============================================================================
-- STEP 7: portfolio-showcase
-- Not in project backup — flag only, template needs manual review
-- ============================================================================

-- NOTE: Check template with:
--   SELECT function, left(html_template, 500) FROM content_components
--   WHERE function = 'portfolio-showcase';
-- Then add --section-* variables to its container CSS.


-- ============================================================================
-- STEP 8: VERIFICATION
-- ============================================================================

-- 8a. All flagged dark components should have --section-text in template
SELECT name, function, is_dark_section,
       html_template LIKE '%--section-text%' as has_section_vars,
       CASE WHEN html_template LIKE '%--section-text%' THEN 'OK' ELSE 'NEEDS REVIEW' END as status
FROM content_components
WHERE is_dark_section = true
ORDER BY function, name;

-- 8b. Check for unflagged dark components (filtering out false positives)
SELECT name, function, is_dark_section, 'CHECK IF DARK' as warning
FROM content_components
WHERE is_dark_section = false
  AND component_level = 'section'
  AND (
    html_template LIKE '%background:%#1a1a2e%'
        OR html_template LIKE '%background: #1a1a2e%'
        OR html_template LIKE '%background: var(--color-primary%'
    )
  AND function NOT IN ('head', 'head-seo-standard', 'head-basic',
                       'site-header', 'header-professional-dark',
                       'header-minimal-light', 'header-bold-gradient')
ORDER BY function;

-- 8c. Verify data-component attributes match function names (naming contract)
SELECT name, function,
       CASE
           WHEN html_template LIKE '%data-component="' || function || '"%' THEN 'OK'
           ELSE 'MISMATCH'
           END as data_component_check
FROM content_components
WHERE is_dark_section = true
ORDER BY function;

--
-- departments grid

-- ============================================================================
-- 1. INSERT new departments-grid component into content_components
-- ============================================================================
-- This sits alongside the existing leadership-team component.
-- The planner chooses departments-grid when the site uses AI agent teams
-- (no real employees), and leadership-team when there are actual people.
--
-- Reuses the same CSS class names (.team-section, .team-member, .member-photo,
-- .member-title, .member-bio) so it gets the same visual treatment as
-- leadership-team without duplicate styles.

INSERT INTO content_components (
    name,
    description,
    html_template,
    input_schema,
    function,
    display_name,
    category,
    semantic_tags,
    render_mode,
    component_level,
    is_active
) VALUES (
             'departments-grid',
             'Grid of AI departments or functional teams with icon images',
             -- html_template:
             E'<section class="team-section" data-component="departments-grid">\n'
    '    <div class="team-container">\n'
    '        {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}\n'
    '        {{if .section_intro}}<p class="section-intro">{{.section_intro}}</p>{{end}}\n'
    '        <div class="team-grid">\n'
    '            {{range .departments}}\n'
    '            <div class="team-member">\n'
    '                {{if .icon}}<img src="{{.icon}}" alt="{{.name}} Department" class="member-photo">{{end}}\n'
    '                <h3>{{.name}}</h3>\n'
    '                {{if .subtitle}}<p class="member-title">{{.subtitle}}</p>{{end}}\n'
    '                {{if .description}}<p class="member-bio">{{.description}}</p>{{end}}\n'
    '            </div>\n'
    '            {{end}}\n'
    '        </div>\n'
    '    </div>\n'
    '</section>\n'
    '<style>\n'
    '.team-section {\n'
    '    padding: 5rem 2rem;\n'
    '    background: #f8f9fa;\n'
    '}\n'
    '.team-container {\n'
    '    max-width: 1200px;\n'
    '    margin: 0 auto;\n'
    '}\n'
    '.team-section h2 {\n'
    '    font-size: clamp(1.75rem, 4vw, 2.5rem);\n'
    '    margin-bottom: 1rem;\n'
    '    color: #1a1a2e;\n'
    '    text-align: center;\n'
    '}\n'
    '.section-intro {\n'
    '    text-align: center;\n'
    '    max-width: 700px;\n'
    '    margin: 0 auto 3rem;\n'
    '    color: #555;\n'
    '    line-height: 1.7;\n'
    '}\n'
    '.team-grid {\n'
    '    display: grid;\n'
    '    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));\n'
    '    gap: 2rem;\n'
    '}\n'
    '.team-member {\n'
    '    background: #fff;\n'
    '    padding: 2rem;\n'
    '    border-radius: 8px;\n'
    '    text-align: center;\n'
    '    box-shadow: 0 2px 8px rgba(0,0,0,0.06);\n'
    '}\n'
    '.member-photo {\n'
    '    width: 120px;\n'
    '    height: 120px;\n'
    '    border-radius: 50%;\n'
    '    object-fit: cover;\n'
    '    margin-bottom: 1.5rem;\n'
    '    background: #e0e0e0;\n'
    '}\n'
    '.team-member h3 {\n'
    '    font-size: 1.25rem;\n'
    '    margin-bottom: 0.5rem;\n'
    '    color: #1a1a2e;\n'
    '}\n'
    '.member-title {\n'
    '    color: #0f3460;\n'
    '    font-weight: 500;\n'
    '    margin-bottom: 1rem;\n'
    '    font-size: 0.95rem;\n'
    '}\n'
    '.member-bio {\n'
    '    color: #555;\n'
    '    line-height: 1.6;\n'
    '    font-size: 0.95rem;\n'
    '    text-align: left;\n'
    '}\n'
    '@media (max-width: 768px) {\n'
    '    .team-section { padding: 3rem 1.5rem; }\n'
    '    .team-grid { grid-template-columns: 1fr; }\n'
    '}\n'
    '</style>',
             -- input_schema:
             '{
                 "section_title": "string",
                 "section_intro": "string",
                 "departments": [{
                     "name": "string",
                     "subtitle": "string",
                     "description": "string",
                     "icon": "string"
                 }]
             }'::jsonb,
             -- function (used as slot_name):
             'departments-grid',
             -- display_name:
             'Departments Grid',
             -- category:
             'about',
             -- semantic_tags:
             '["departments", "teams", "ai-departments", "functional-teams", "agent-teams"]'::jsonb,
             -- render_mode:
             'template',
             -- component_level:
             'section',
             -- is_active:
             true
         );

----

-- hardcoded colours in heros in db that need fixing

-- 063b_hardcoded_colors_discovery.sql
--
-- Audit queries to run against clients_db to find components
-- with hardcoded colors that should be using CSS variables.
-- Run these BEFORE deploying 062 to understand the scope of work.
--
-- These are SELECT-only queries — no data changes.

-- ============================================================
-- 1. Find components with hardcoded hex colors in their CSS
--    (excludes colors inside CSS variable fallbacks like var(--x, #fff))
-- ============================================================
SELECT
    name,
    function,
    category,
    component_level,
    -- Count occurrences of hardcoded #hex patterns
    (LENGTH(html_template) - LENGTH(REPLACE(REPLACE(REPLACE(
                                                            html_template,
                                                            'var(--', ''), -- strip var() references first (crude but effective)
                                                    '#ffffff', ''),
                                            '#fff', ''
                                    ))) as approx_hardcoded_count,
    CASE
        WHEN html_template LIKE '%color: #%' AND html_template NOT LIKE '%var(--%'
            THEN 'ALL hardcoded (no vars at all)'
        WHEN html_template LIKE '%color: #%' AND html_template LIKE '%var(--%'
            THEN 'MIXED (some vars, some hardcoded)'
        WHEN html_template LIKE '%var(--%' AND html_template NOT LIKE '%color: #%'
            THEN 'CLEAN (vars only)'
        ELSE 'NO CSS colors found'
        END as css_status
FROM content_components
WHERE component_level = 'section'
ORDER BY css_status, category, function;

-- ============================================================
-- 2. Specific: find sections that hardcode text color
--    (these will fight with the inheritance model)
-- ============================================================
SELECT
    name,
    function,
    'hardcoded text color' as issue,
    SUBSTRING(html_template FROM 'color:\s*#[0-9a-fA-F]{3,8}') as found_pattern
FROM content_components
WHERE component_level = 'section'
  AND html_template ~ 'color:\s*#[0-9a-fA-F]{3,8}'
  -- Exclude colors that are inside var() fallbacks
  AND html_template !~ 'var\(--[^)]+,\s*#[0-9a-fA-F]{3,8}\)'
ORDER BY function;

-- ============================================================
-- 3. Find dark-section components (dark backgrounds)
--    Check if they have the --section-* variable contract
-- ============================================================
SELECT
    name,
    function,
    CASE
        WHEN html_template LIKE '%background:%#1a1a2e%'
            OR html_template LIKE '%background: #1a1a2e%' THEN 'hardcoded #1a1a2e'
        WHEN html_template LIKE '%background:%#0f172a%'
            OR html_template LIKE '%background: #0f172a%' THEN 'hardcoded #0f172a'
        WHEN html_template LIKE '%background: var(--color-primary%' THEN 'var(--color-primary)'
        WHEN html_template LIKE '%linear-gradient%1a1a2e%'
            OR html_template LIKE '%linear-gradient%16213e%' THEN 'gradient-dark'
        WHEN html_template LIKE '%linear-gradient(rgba(0,0,0%' THEN 'overlay-dark'
        ELSE 'other'
        END as dark_bg_type,
    CASE
        WHEN html_template LIKE '%--section-text%' THEN 'YES'
        ELSE 'MISSING'
        END as has_section_contract,
    CASE
        WHEN html_template LIKE '%--section-heading%' THEN 'YES'
        ELSE 'MISSING'
        END as has_section_heading,
    CASE
        WHEN html_template LIKE '%--section-surface%' THEN 'YES'
        ELSE 'MISSING'
        END as has_section_surface
FROM content_components
WHERE component_level = 'section'
  AND (
    html_template LIKE '%background:%#1a1a2e%'
        OR html_template LIKE '%background: #1a1a2e%'
        OR html_template LIKE '%background: var(--color-primary%'
        OR html_template LIKE '%background-color: var(--color-primary%'
        OR html_template LIKE '%linear-gradient%1a1a2e%'
        OR html_template LIKE '%linear-gradient%16213e%'
        OR html_template LIKE '%background:%#0f172a%'
        OR html_template LIKE '%background: #0f172a%'
        OR html_template LIKE '%linear-gradient(rgba(0,0,0%'
    )
ORDER BY has_section_contract, function;

-- ============================================================
-- 4. Find components that set color on text elements explicitly
--    (these override inheritance in dark sections)
-- ============================================================
SELECT
    name,
    function,
    CASE WHEN html_template ~ 'p\s*\{[^}]*color:' THEN 'p' ELSE '' END ||
    CASE WHEN html_template ~ 'h[1-6]\s*\{[^}]*color:' THEN ' h1-h6' ELSE '' END ||
    CASE WHEN html_template ~ 'blockquote\s*\{[^}]*color:' THEN ' blockquote' ELSE '' END ||
    CASE WHEN html_template ~ 'li\s*\{[^}]*color:' THEN ' li' ELSE '' END ||
    CASE WHEN html_template ~ 'strong\s*\{[^}]*color:' THEN ' strong' ELSE '' END
        as elements_with_forced_color
FROM content_components
WHERE component_level = 'section'
  AND (
    html_template ~ 'p\s*\{[^}]*color:'
    OR html_template ~ 'h[1-6]\s*\{[^}]*color:'
    OR html_template ~ 'blockquote\s*\{[^}]*color:'
    OR html_template ~ 'li\s*\{[^}]*color:'
    OR html_template ~ 'strong\s*\{[^}]*color:'
    )
ORDER BY function;

-- ============================================================
-- 5. Summary: overall CSS variable adoption
-- ============================================================
SELECT
    component_level,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE html_template LIKE '%var(--%') as uses_css_vars,
    COUNT(*) FILTER (WHERE html_template LIKE '%color: #%') as has_hardcoded_colors,
    COUNT(*) FILTER (WHERE html_template LIKE '%var(--%' AND html_template NOT LIKE '%color: #%') as fully_migrated,
    COUNT(*) FILTER (WHERE html_template LIKE '%--section-text%') as has_section_contract
FROM content_components
GROUP BY component_level
ORDER BY component_level;


----

-- backfill slot names

-- Migration 072: Backfill empty slot_names in page_components
--
-- Problem: Some page_components rows have NULL or empty slot_name.
-- This breaks section-editor's loadPageComponentBySlot which matches by slot_name.
--
-- Fix: pages.sections stores the planned section names as a JSON array
-- (e.g. ["contact-hero", "contact-form", "contact-info"]) in position order.
-- page_components.position maps 1:1 to this array (1-indexed).
-- Backfill slot_name from pages.sections[position - 1].
--
-- Only updates rows where slot_name is NULL or empty AND the position
-- maps to a valid index in the sections array.

BEGIN;

-- Backfill from pages.sections array
UPDATE page_components pc
SET slot_name = trim(both '"' from (p.sections->(pc.position - 1))::text)
    FROM pages p
WHERE pc.page_id = p.id
  AND (pc.slot_name IS NULL OR pc.slot_name = '')
  AND p.sections IS NOT NULL
  AND jsonb_array_length(p.sections) > 0
  AND pc.position > 0
  AND pc.position <= jsonb_array_length(p.sections);

-- Secondary backfill: use content_components.function for rows with component_id
-- but still no slot_name (e.g. if pages.sections was also empty)
UPDATE page_components pc
SET slot_name = cc.function
    FROM content_components cc
WHERE pc.component_id = cc.id
  AND (pc.slot_name IS NULL OR pc.slot_name = '')
  AND cc.function IS NOT NULL
  AND cc.function != '';

COMMIT;

-- Verify: any remaining empty slot_names?
SELECT COUNT(*) as empty_slots,
       COUNT(*) FILTER (WHERE pc.slot_name IS NOT NULL AND pc.slot_name != '') as filled_slots
FROM page_components pc;


----
-- rewrite to section level and add hitl classification
-- major rewrite

-- ============================================================================
-- Migration: content_components input_schema v2
--
-- Replaces flat type declarations with structured field definitions that
-- declare where each field's data comes from and what to do when it's missing.
--
-- The plan_sections action reads these schemas to determine which sections
-- can be generated (data available) vs which need human input.
--
-- Also deactivates legacy duplicate components (display-name versions).
-- ============================================================================

BEGIN;

-- ============================================================================
-- Part 1: Deactivate legacy duplicates
-- Keep the kebab-case production versions, deactivate the display-name ones
-- ============================================================================

UPDATE content_components SET is_active = false, updated_at = NOW()
WHERE name IN (
               'Call to Action',    -- duplicate of call_to_action
               'Features Grid',     -- duplicate of features
               'Site Footer',       -- duplicate of footer-standard/simple/4-column
               'Document Head',     -- duplicate of head-seo-standard
               'Site Header',       -- duplicate of header-professional-dark etc
               'Hero Section',      -- duplicate of hero
               'Social Proof'       -- duplicate of social_proof
    );

-- ============================================================================
-- Part 2: Heroes — LLM generates copy, images from site assets
-- ============================================================================

-- Main hero (with CTA and background image)
UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":        {"type": "text",  "source": "llm", "required": true},
        "subheadline":     {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "cta_text":        {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "cta_url":         {"type": "url",   "source": "pages.contact", "required": false, "on_missing": "use_fallback", "fallback": "/contact.html"},
        "secondary_cta":   {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "secondary_cta_url": {"type": "url", "source": "pages.services", "required": false, "on_missing": "use_fallback", "fallback": "/services.html"},
        "background_image": {"type": "image", "source": "site_assets.hero", "required": false, "on_missing": "use_fallback", "fallback": "/assets/images/hero.jpg"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'hero' AND is_active = true;

-- Simple heroes (just headline + subheadline)
UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":    {"type": "text", "source": "llm", "required": true},
        "subheadline": {"type": "text", "source": "llm", "required": false, "on_missing": "skip_field"}
    }
}'::jsonb, updated_at = NOW()
WHERE name IN ('about-hero', 'services-hero', 'contact-hero', 'case-studies-hero', 'use-cases-hero')
  AND is_active = true;

-- ============================================================================
-- Part 3: Content blocks — LLM generates everything
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {
        "section_title": {"type": "text", "source": "llm", "required": false, "on_missing": "skip_field"},
        "content":       {"type": "text", "source": "llm", "required": true},
        "highlights":    {"type": "array", "source": "llm", "required": false, "on_missing": "skip_field",
                          "items": {"title": "string", "description": "string"}}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'about-content' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "heading": {"type": "text", "source": "llm", "required": true},
        "content": {"type": "text", "source": "llm", "required": true}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'Generic Text Block' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":    {"type": "text",  "source": "llm", "required": true},
        "subheadline": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "features":    {"type": "array", "source": "llm", "required": true, "min_items": 2,
                        "items": {"name": "string", "icon": "string", "description": "string"}}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'features' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline": {"type": "text",  "source": "llm", "required": true},
        "features": {"type": "array", "source": "llm", "required": true, "min_items": 2,
                     "items": {"name": "string", "description": "string"}}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'differentiators-section' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":    {"type": "text",  "source": "llm", "required": true},
        "subheadline": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "features":    {"type": "array", "source": "llm", "required": true, "min_items": 2,
                        "items": {"name": "string", "description": "string"}}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'services-grid' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "section_title": {"type": "text",  "source": "llm", "required": true},
        "section_intro": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "departments":   {"type": "array", "source": "llm", "required": true, "min_items": 2,
                          "items": {"name": "string", "icon": "string", "subtitle": "string", "description": "string"}}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'departments-grid' AND is_active = true;

-- ============================================================================
-- Part 4: Real data required — needs factual information, cannot fabricate
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {
        "section_title": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "section_intro": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "members":       {"type": "array", "source": "site_specs.identity.team", "required": true,
                          "on_missing": "needs_human_review", "min_items": 1,
                          "items": {"name": "string", "title": "string", "bio": "string", "photo": "string"},
                          "missing_reason": "Team member names, titles, and bios are needed"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'leadership-team' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":     {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "testimonials": {"type": "array", "source": "site_specs.social_proof.testimonials", "required": true,
                         "on_missing": "skip_section", "min_items": 1,
                         "items": {"quote": "string", "author": "string", "role": "string", "company": "string"},
                         "missing_reason": "Customer testimonials with real names and quotes"}
    }
}'::jsonb, updated_at = NOW()
WHERE name IN ('social_proof', 'testimonials') AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":    {"type": "text",  "source": "llm", "required": true},
        "subheadline": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "case_studies": {"type": "array", "source": "site_specs.portfolio.case_studies", "required": true,
                         "on_missing": "needs_human_review", "min_items": 1,
                         "items": {"title": "string", "client": "string", "summary": "string", "results": "string"},
                         "missing_reason": "Real case study data — client names, project descriptions, outcomes"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'case-studies-list' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":    {"type": "text",  "source": "llm", "required": true},
        "subheadline": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "use_cases":   {"type": "array", "source": "site_specs.portfolio.use_cases", "required": true,
                        "on_missing": "needs_human_review", "min_items": 1,
                        "items": {"title": "string", "client": "string", "summary": "string", "results": "string"},
                        "missing_reason": "Real use case data — client names, descriptions, outcomes"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'use-cases-list' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline": {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "intro":    {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "projects": {"type": "array", "source": "site_specs.portfolio.projects", "required": true,
                     "on_missing": "needs_human_review", "min_items": 1,
                     "items": {"title": "string", "domain": "string", "description": "string",
                               "live_url": "string", "build_time": "string", "built_with": "string"},
                     "missing_reason": "Real project data — titles, URLs, descriptions"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'portfolio-showcase' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "section_title":   {"type": "text", "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Choose Your Plan"},
        "tier_1_name":     {"type": "text", "source": "site_specs.pricing.tiers[0].name", "required": true,
                            "on_missing": "needs_human_review", "missing_reason": "Pricing tier names and prices"},
        "tier_1_price":    {"type": "text", "source": "site_specs.pricing.tiers[0].price", "required": true, "on_missing": "needs_human_review"},
        "tier_1_features": {"type": "array", "source": "site_specs.pricing.tiers[0].features", "required": true, "on_missing": "needs_human_review"},
        "tier_1_cta":      {"type": "text", "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Get Started"},
        "tier_2_name":     {"type": "text", "source": "site_specs.pricing.tiers[1].name", "required": false, "on_missing": "skip_field"},
        "tier_2_price":    {"type": "text", "source": "site_specs.pricing.tiers[1].price", "required": false, "on_missing": "skip_field"},
        "tier_2_features": {"type": "array", "source": "site_specs.pricing.tiers[1].features", "required": false, "on_missing": "skip_field"},
        "tier_2_cta":      {"type": "text", "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Start Free Trial"},
        "tier_3_name":     {"type": "text", "source": "site_specs.pricing.tiers[2].name", "required": false, "on_missing": "skip_field"},
        "tier_3_price":    {"type": "text", "source": "site_specs.pricing.tiers[2].price", "required": false, "on_missing": "skip_field"},
        "tier_3_features": {"type": "array", "source": "site_specs.pricing.tiers[2].features", "required": false, "on_missing": "skip_field"},
        "tier_3_cta":      {"type": "text", "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Contact Sales"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'Pricing Tiers' AND is_active = true;

-- FAQ — LLM can generate plausible FAQs from site_specs.content_direction
UPDATE content_components SET input_schema = '{
    "fields": {
        "section_title": {"type": "text", "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Frequently Asked Questions"},
        "questions":     {"type": "array", "source": "llm", "required": true, "min_items": 3,
                          "items": {"question": "string", "answer": "string"},
                          "llm_guidance": "Generate FAQs based on the industry, services, and common objections for this type of business"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'FAQ Section' AND is_active = true;

-- ============================================================================
-- Part 5: Contact — needs real business data
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {
        "heading":          {"type": "text", "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Get in Touch"},
        "form_description": {"type": "text", "source": "llm", "required": false, "on_missing": "skip_field"},
        "form_action":      {"type": "url",  "source": "config.contact_form_action", "required": false, "on_missing": "use_fallback", "fallback": "#contact"},
        "submit_text":      {"type": "text", "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Send Message"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'contact-form' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "section_title": {"type": "text",  "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Contact Us"},
        "intro_text":    {"type": "text",  "source": "llm", "required": false, "on_missing": "skip_field"},
        "email":         {"type": "text",  "source": "site_specs.identity.email", "required": false,
                          "on_missing": "needs_human_review", "missing_reason": "Business contact email address"},
        "phone":         {"type": "text",  "source": "site_specs.identity.phone", "required": false, "on_missing": "skip_field"},
        "address":       {"type": "text",  "source": "site_specs.identity.address", "required": false, "on_missing": "skip_field"},
        "hours":         {"type": "text",  "source": "site_specs.identity.hours", "required": false, "on_missing": "skip_field"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'contact-info' AND is_active = true;

-- ============================================================================
-- Part 6: CTA — LLM copy, URLs resolved from pages
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {
        "headline":          {"type": "text", "source": "llm", "required": true},
        "subheadline":       {"type": "text", "source": "llm", "required": false, "on_missing": "skip_field"},
        "primary_cta":       {"type": "text", "source": "llm", "required": true},
        "primary_cta_url":   {"type": "url",  "source": "pages.contact", "required": false, "on_missing": "use_fallback", "fallback": "/contact.html"},
        "secondary_cta":     {"type": "text", "source": "llm", "required": false, "on_missing": "skip_field"},
        "secondary_cta_url": {"type": "url",  "source": "pages.services", "required": false, "on_missing": "use_fallback", "fallback": "/services.html"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'call_to_action' AND is_active = true;

-- ============================================================================
-- Part 7: Structural — rendered by rerender agent, not content writer
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {
        "title":            {"type": "text",  "source": "renderer", "required": true},
        "description":      {"type": "text",  "source": "renderer", "required": false, "on_missing": "skip_field"},
        "theme_css":        {"type": "text",  "source": "renderer", "required": false, "on_missing": "skip_field"},
        "font_url":         {"type": "text",  "source": "renderer", "required": false, "on_missing": "skip_field"},
        "canonical_url":    {"type": "text",  "source": "renderer", "required": false, "on_missing": "skip_field"},
        "primary_color":    {"type": "text",  "source": "config.color_scheme.primary", "required": false, "on_missing": "use_fallback", "fallback": "#1a1a2e"},
        "secondary_color":  {"type": "text",  "source": "config.color_scheme.secondary", "required": false, "on_missing": "use_fallback", "fallback": "#2d2d44"},
        "accent_color":     {"type": "text",  "source": "config.color_scheme.accent", "required": false, "on_missing": "use_fallback", "fallback": "#16a085"},
        "background_color": {"type": "text",  "source": "config.color_scheme.background", "required": false, "on_missing": "use_fallback", "fallback": "#ffffff"},
        "text_color":       {"type": "text",  "source": "config.color_scheme.text", "required": false, "on_missing": "use_fallback", "fallback": "#333333"},
        "structured_data":  {"type": "text",  "source": "renderer", "required": false, "on_missing": "skip_field"},
        "analytics_id":     {"type": "text",  "source": "config.analytics_id", "required": false, "on_missing": "skip_field"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'head-seo-standard' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {}
}'::jsonb, updated_at = NOW()
WHERE name = 'body-close' AND is_active = true;

-- Headers — rendered by rerender agent
UPDATE content_components SET input_schema = '{
    "fields": {
        "logo_text":      {"type": "text",  "source": "site_specs.identity.company_name", "required": true, "on_missing": "use_fallback", "fallback": "Company"},
        "logo_url":       {"type": "image", "source": "site_assets.logo", "required": false, "on_missing": "use_fallback", "fallback": "/assets/images/logo.png"},
        "nav_items":      {"type": "array", "source": "renderer.nav", "required": true},
        "cta_text":       {"type": "text",  "source": "llm", "required": false, "on_missing": "use_fallback", "fallback": "Get Started"},
        "cta_url":        {"type": "url",   "source": "pages.contact", "required": false, "on_missing": "use_fallback", "fallback": "/contact.html"},
        "primary_color":  {"type": "text",  "source": "config.color_scheme.primary", "required": false, "on_missing": "use_fallback", "fallback": "#1a1a2e"},
        "accent_color":   {"type": "text",  "source": "config.color_scheme.accent", "required": false, "on_missing": "use_fallback", "fallback": "#16a085"}
    }
}'::jsonb, updated_at = NOW()
WHERE name IN ('header-professional-dark', 'header-minimal-light', 'header-bold-gradient') AND is_active = true;

-- Footers — rendered by rerender agent
UPDATE content_components SET input_schema = '{
    "fields": {
        "company_name":   {"type": "text",  "source": "site_specs.identity.company_name", "required": true, "on_missing": "use_fallback", "fallback": "Company"},
        "tagline":        {"type": "text",  "source": "site_specs.identity.tagline", "required": false, "on_missing": "skip_field"},
        "contact_email":  {"type": "text",  "source": "site_specs.identity.email", "required": false, "on_missing": "skip_field"},
        "contact_phone":  {"type": "text",  "source": "site_specs.identity.phone", "required": false, "on_missing": "skip_field"},
        "nav_items":      {"type": "array", "source": "renderer.nav", "required": false},
        "copyright":      {"type": "text",  "source": "renderer", "required": false, "on_missing": "use_fallback", "fallback": "© 2026 Company. All rights reserved."}
    }
}'::jsonb, updated_at = NOW()
WHERE name IN ('footer-4-column', 'footer-simple', 'footer-standard') AND is_active = true;

-- ============================================================================
-- Part 8: Content/blog — dynamic data from DB queries
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {
        "section_title":    {"type": "text",    "source": "llm", "required": false, "on_missing": "skip_field"},
        "section_subtitle": {"type": "text",    "source": "llm", "required": false, "on_missing": "skip_field"},
        "articles":         {"type": "array",   "source": "query.blog_posts", "required": true, "on_missing": "skip_section",
                             "missing_reason": "No blog posts published yet"},
        "show_load_more":   {"type": "boolean", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": true},
        "load_more_text":   {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Load More"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'article_grid' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "brand_name":      {"type": "text",    "source": "site_specs.identity.company_name", "required": true, "on_missing": "use_fallback", "fallback": "Blog"},
        "categories":      {"type": "array",   "source": "query.blog_categories", "required": false, "on_missing": "skip_field"},
        "show_subscribe":  {"type": "boolean", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": false},
        "subscribe_text":  {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Subscribe"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'content_header' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "brand_name":               {"type": "text",    "source": "site_specs.identity.company_name", "required": true, "on_missing": "use_fallback", "fallback": "Blog"},
        "tagline":                  {"type": "text",    "source": "site_specs.identity.tagline", "required": false, "on_missing": "skip_field"},
        "categories":               {"type": "array",   "source": "query.blog_categories", "required": false, "on_missing": "skip_field"},
        "company_links":            {"type": "array",   "source": "renderer.nav", "required": false, "on_missing": "skip_field"},
        "social_links":             {"type": "array",   "source": "site_specs.identity.social_links", "required": false, "on_missing": "skip_field"},
        "legal_links":              {"type": "array",   "source": "renderer.legal_nav", "required": false, "on_missing": "skip_field"},
        "copyright":                {"type": "text",    "source": "renderer", "required": false, "on_missing": "use_fallback", "fallback": "© 2026"},
        "newsletter_title":         {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Stay Updated"},
        "newsletter_description":   {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Get the latest articles delivered to your inbox"},
        "email_placeholder":        {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Enter your email"},
        "show_newsletter":          {"type": "boolean", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": false}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'content_footer' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "category_name":     {"type": "text",  "source": "query.category", "required": true},
        "category_slug":     {"type": "text",  "source": "query.category", "required": true},
        "category_articles": {"type": "array", "source": "query.category_posts", "required": true, "on_missing": "skip_section"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'category_section' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "featured_title":    {"type": "text",  "source": "query.featured_post", "required": true, "on_missing": "skip_section"},
        "featured_excerpt":  {"type": "text",  "source": "query.featured_post", "required": false, "on_missing": "skip_field"},
        "featured_image":    {"type": "image", "source": "query.featured_post", "required": false, "on_missing": "skip_field"},
        "featured_author":   {"type": "text",  "source": "query.featured_post", "required": false, "on_missing": "skip_field"},
        "featured_date":     {"type": "text",  "source": "query.featured_post", "required": false, "on_missing": "skip_field"},
        "featured_category": {"type": "text",  "source": "query.featured_post", "required": false, "on_missing": "skip_field"},
        "featured_read_time": {"type": "text", "source": "query.featured_post", "required": false, "on_missing": "skip_field"},
        "read_more_text":    {"type": "text",  "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Read More"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'featured_article' AND is_active = true;

UPDATE content_components SET input_schema = '{
    "fields": {
        "show_popular":       {"type": "boolean", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": true},
        "popular_title":      {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Popular Articles"},
        "popular_articles":   {"type": "array",   "source": "query.popular_posts", "required": false, "on_missing": "skip_field"},
        "show_categories":    {"type": "boolean", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": true},
        "categories_title":   {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Categories"},
        "category_links":     {"type": "array",   "source": "query.blog_categories", "required": false, "on_missing": "skip_field"},
        "show_newsletter":    {"type": "boolean", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": false},
        "newsletter_title":   {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Newsletter"},
        "newsletter_description": {"type": "text", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Subscribe for updates"},
        "email_placeholder":  {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Enter your email"},
        "subscribe_button":   {"type": "text",    "source": "static", "required": false, "on_missing": "use_fallback", "fallback": "Subscribe"},
        "show_ad":            {"type": "boolean", "source": "static", "required": false, "on_missing": "use_fallback", "fallback": false},
        "ad_slot_id":         {"type": "text",    "source": "config.ad_slot_id", "required": false, "on_missing": "skip_field"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'content_sidebar' AND is_active = true;

-- ============================================================================
-- Part 9: Tools — self-contained, no external data needs
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {}
}'::jsonb, updated_at = NOW()
WHERE function LIKE 'tool-%' AND is_active = true;

-- ============================================================================
-- Part 10: Ads
-- ============================================================================

UPDATE content_components SET input_schema = '{
    "fields": {
        "ad_slot_id": {"type": "text", "source": "config.ad_slot_id", "required": true, "on_missing": "skip_section"}
    }
}'::jsonb, updated_at = NOW()
WHERE name = 'ad_zone_inline' AND is_active = true;

COMMIT;

-- ============================================================================
-- Verify
-- ============================================================================

-- Check all active components have the new format
SELECT name, function,
       input_schema->'fields' IS NOT NULL as has_v2_fields,
       jsonb_object_keys(COALESCE(input_schema->'fields', '{}'::jsonb)) as sample_field
FROM content_components
WHERE is_active = true
ORDER BY function, name
    LIMIT 20;

-- Check deactivated duplicates
SELECT name, function, is_active FROM content_components WHERE is_active = false ORDER BY function;

-- Count by source type
SELECT source, COUNT(*) as field_count
FROM (
         SELECT value->>'source' as source
         FROM content_components,
             jsonb_each(COALESCE(input_schema->'fields', '{}'::jsonb))
         WHERE is_active = true
     ) sub
GROUP BY source ORDER BY field_count DESC;

---
-- skip if missing
UPDATE content_components
SET input_schema = jsonb_set(
        input_schema,
        '{fields,nav_items,on_missing}',
        '"skip_section"'
                   )
WHERE function IN ('site-header')
  AND input_schema->'fields'->'nav_items' IS NOT NULL;

---
-- tools
-- ============================================================
-- Populate empty library tools with HTML templates
-- 5 of 6 empty tools (clip-path-builder source not provided)
-- ============================================================

-- 1. FAVICON GENERATOR
UPDATE content_components
SET html_template = $fav$<style>
    .tool-layout { display: grid; gap: var(--space-lg); }
    @media (min-width: 900px) { .tool-layout { grid-template-columns: 350px 1fr; } }
    .emoji-grid {
        display: grid; grid-template-columns: repeat(6, 1fr); gap: 0.5rem;
        max-height: 200px; overflow-y: auto;
        border: 1px solid var(--color-border, #eee); padding: 0.5rem; border-radius: 8px; margin-top: 1rem;
    }
    .emoji-btn {
        background: var(--color-surface, #f4f4f5); border: none; font-size: 1.5rem; cursor: pointer;
        border-radius: 4px; padding: 0.25rem; transition: background 0.2s;
    }
    .emoji-btn:hover { background: var(--color-border, #e4e4e7); transform: scale(1.1); }
    .preview-grid {
        display: flex; gap: 2rem; align-items: flex-end;
        background: var(--color-surface, #f4f4f5); padding: 2rem; border-radius: 8px; margin-top: 1rem;
    }
    .icon-preview { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; }
    .icon-box { background: #fff; border: 1px dashed var(--color-border, #ccc); display: flex; align-items: center; justify-content: center; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
    .browser-tab { background: var(--color-border, #e4e4e7); padding: 8px 16px; border-radius: 8px 8px 0 0; display: inline-flex; align-items: center; gap: 8px; font-size: 0.8rem; font-family: sans-serif; color: var(--color-text-muted, #555); width: 200px; }
    .guide-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card); }
    .code-output {
        background: #1e1e1e; color: #a9ff68; padding: 1rem;
        border-radius: 8px; font-family: var(--font-mono); font-size: 0.8rem;
        margin-top: 1rem; width: 100%; box-sizing: border-box;
    }
</style>

<main class="container" style="padding-top: var(--space-lg);">
    <h1 style="font-size: var(--text-h2);">Smart Favicon Generator</h1>

    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">The Caching Problem</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">
                    Browsers stubbornly cache favicons. If you update your file but don't see the change, try pressing <code>Ctrl + F5</code> (Windows) or <code>Cmd + Shift + R</code> (Mac) to force a refresh.
                </p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Implementation</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">
                    Place the generated <code>favicon.ico</code> in your site's root folder (where your index.html is). For best results, add the link tag below to your code.
                </p>
            </div>
        </div>
    </div>

    <div class="tool-layout">
        <aside>
            <div style="background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px;">
                <h3 style="font-size: 1rem; margin-bottom: 1rem;">1. Choose Icon</h3>
                <div style="display:flex; gap:0.5rem; margin-bottom:1rem;">
                    <button class="btn" id="btnEmoji" onclick="setMode('emoji')" style="flex:1; padding:0.5rem;">Emoji</button>
                    <button class="btn-outline" id="btnUpload" onclick="setMode('upload')" style="flex:1; padding:0.5rem;">Upload</button>
                </div>
                <div id="controlsEmoji">
                    <label style="display:block; margin-bottom:0.5rem; font-weight:600;">Selected Emoji</label>
                    <input type="text" id="emojiInput" value="🚀" style="font-size: 2rem; width: 100%; text-align: center; border: 1px solid var(--color-border, #ccc); border-radius: 4px; margin-bottom: 1rem;">
                    <label style="font-weight: 600; font-size: 0.85rem;">Quick Pick</label>
                    <div class="emoji-grid" id="emojiGrid"></div>
                    <label style="display:block; margin-top:1rem; margin-bottom:0.5rem; font-weight:600;">Size</label>
                    <input type="range" id="emojiSize" min="50" max="150" value="100" style="width:100%">
                </div>
                <div id="controlsUpload" style="display:none;">
                    <label style="display:block; margin-bottom:0.5rem; font-weight:600;">Upload Logo</label>
                    <input type="file" id="fileInput" accept="image/*">
                </div>
            </div>
            <button class="btn" onclick="downloadAll()" style="width:100%; margin-top: 1rem; background: var(--color-accent);">Download Icons</button>
        </aside>

        <section>
            <h3 style="font-size: 1rem;">Live Preview</h3>
            <div style="margin-top: 1rem;">
                <div class="browser-tab">
                    <img id="tabIcon" src="" style="width:16px; height:16px;">
                    <span>My Website</span>
                    <span style="margin-left: auto;">×</span>
                </div>
            </div>
            <div class="preview-grid">
                <div class="icon-preview">
                    <div class="icon-box" style="width: 32px; height: 32px;"><img id="prev32"></div>
                    <span style="font-size: 0.75rem;">32x32<br>(favicon.ico)</span>
                </div>
                <div class="icon-preview">
                    <div class="icon-box" style="width: 90px; height: 90px;"><img id="prev180" style="width: 90px; height: 90px;"></div>
                    <span style="font-size: 0.75rem;">180x180<br>(Touch Icon)</span>
                </div>
            </div>
            <h3 style="font-size: 1rem; margin-top: 2rem;">Installation Code</h3>
            <p style="font-size: 0.85rem; color: var(--color-text-muted, #666); margin-top: 0.5rem;">Paste this into your <code>&lt;head&gt;</code> tag:</p>
            <textarea class="code-output" readonly><link rel="icon" type="image/x-icon" href="/favicon.ico">
<link rel="apple-touch-icon" href="/apple-touch-icon.png"></textarea>
            <canvas id="renderCanvas" width="512" height="512" style="display:none;"></canvas>
        </section>
    </div>
</main>

<script>
    const canvas = document.getElementById('renderCanvas');
    const ctx = canvas.getContext('2d');
    const inputs = { emoji: document.getElementById('emojiInput'), size: document.getElementById('emojiSize'), file: document.getElementById('fileInput') };
    const previews = { tab: document.getElementById('tabIcon'), p32: document.getElementById('prev32'), p180: document.getElementById('prev180') };
    const emojiGrid = document.getElementById('emojiGrid');
    let mode = 'emoji';
    let uploadedImg = new Image();
    const emojis = ["🚀","⚡","🔥","✨","💎","❤️","✅","⭐","💡","🎉","🛠️","💻","🎨","📈","🔒","🌍","🌙","☀️","🌊","🌵","🍀","🍁","🍄","🍔","🍕","☕","🍺","🐱","🐶","🦊","🦁","🐸","🦄","💀","👻","🤖","👾","🎃","🏆","⚽","🏀","🎮","🎵","📷","💰","🛒","📦","✉️","📞","📅","📍","🚩","🏁","👁️","🧠","👋","👍"];
    emojis.forEach(e => {
        const btn = document.createElement('button');
        btn.className = 'emoji-btn'; btn.innerText = e;
        btn.onclick = () => { inputs.emoji.value = e; render(); };
        emojiGrid.appendChild(btn);
    });
    function setMode(newMode) {
        mode = newMode;
        document.getElementById('controlsEmoji').style.display = mode === 'emoji' ? 'block' : 'none';
        document.getElementById('controlsUpload').style.display = mode === 'upload' ? 'block' : 'none';
        if(mode === 'emoji') { document.getElementById('btnEmoji').className = 'btn'; document.getElementById('btnUpload').className = 'btn-outline'; }
        else { document.getElementById('btnEmoji').className = 'btn-outline'; document.getElementById('btnUpload').className = 'btn'; }
        render();
    }
    function render() {
        ctx.clearRect(0, 0, 512, 512);
        if (mode === 'emoji') {
            const fontSize = inputs.size.value * 3;
            ctx.font = fontSize + 'px sans-serif';
            ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
            ctx.fillText(inputs.emoji.value, 256, 256 + (fontSize * 0.1));
            updatePreviews();
        } else if (mode === 'upload' && uploadedImg.src) {
            const aspect = uploadedImg.width / uploadedImg.height;
            let drawW = 512; let drawH = 512;
            if (aspect > 1) { drawH = 512 / aspect; } else { drawW = 512 * aspect; }
            ctx.drawImage(uploadedImg, (512-drawW)/2, (512-drawH)/2, drawW, drawH);
            updatePreviews();
        }
    }
    function updatePreviews() {
        const dataUrl = canvas.toDataURL('image/png');
        previews.tab.src = dataUrl; previews.p32.src = dataUrl; previews.p180.src = dataUrl;
    }
    async function pngToIco(pngBlob) {
        const pngData = new Uint8Array(await pngBlob.arrayBuffer());
        const fileSize = pngData.length;
        const header = new Uint8Array([0, 0, 1, 0, 1, 0]);
        const entry = new Uint8Array(16);
        const view = new DataView(entry.buffer);
        entry[0] = 32; entry[1] = 32; entry[2] = 0; entry[3] = 0;
        view.setUint16(4, 1, true); view.setUint16(6, 32, true);
        view.setUint32(8, fileSize, true); view.setUint32(12, 22, true);
        const icoData = new Uint8Array(header.length + entry.length + fileSize);
        icoData.set(header, 0); icoData.set(entry, header.length); icoData.set(pngData, header.length + entry.length);
        return new Blob([icoData], { type: 'image/x-icon' });
    }
    async function downloadAll() {
        const c180 = document.createElement('canvas'); c180.width=180; c180.height=180;
        c180.getContext('2d').drawImage(canvas, 0,0,180,180);
        const linkPNG = document.createElement('a'); linkPNG.download = 'apple-touch-icon.png';
        linkPNG.href = c180.toDataURL('image/png'); linkPNG.click();
        const c32 = document.createElement('canvas'); c32.width=32; c32.height=32;
        c32.getContext('2d').drawImage(canvas, 0,0,32,32);
        c32.toBlob(async (blob) => {
            const icoBlob = await pngToIco(blob);
            const linkICO = document.createElement('a'); linkICO.download = 'favicon.ico';
            linkICO.href = URL.createObjectURL(icoBlob); linkICO.click();
        }, 'image/png');
    }
    inputs.emoji.addEventListener('input', render);
    inputs.size.addEventListener('input', render);
    inputs.file.addEventListener('change', (e) => { const file = e.target.files[0]; if(file) { const reader = new FileReader(); reader.onload = (evt) => { uploadedImg.onload = render; uploadedImg.src = evt.target.result; }; reader.readAsDataURL(file); } });
    render();
</script>$fav$,
    updated_at = NOW()
WHERE id = '0004ef64-88c4-47b5-8b25-a636bf352100'
  AND html_template = '';


-- 2. MEME GENERATOR
UPDATE content_components
SET html_template = $meme$<style>
    .tool-layout { display: grid; grid-template-columns: 1fr; gap: var(--space-lg); align-items: start; }
    @media (min-width: 900px) { .tool-layout { grid-template-columns: 350px 1fr; } }
    .canvas-container {
        background: var(--color-border, #eee); border: 2px dashed var(--color-border, #ccc);
        display: flex; align-items: center; justify-content: center;
        min-height: 400px; overflow: hidden; border-radius: var(--radius-md);
    }
    canvas { max-width: 100%; height: auto; box-shadow: var(--shadow-card); }
    .guide-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card); }
    .color-toggle { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
    .color-btn { width: 24px; height: 24px; border: 1px solid var(--color-border, #ccc); border-radius: 4px; cursor: pointer; }
    .color-btn.active { border: 2px solid var(--color-accent); transform: scale(1.1); }
</style>

<main class="container" style="padding-top: var(--space-lg);">
    <div style="margin-bottom: var(--space-lg);"><h1 style="font-size: var(--text-h2);">Meme Studio</h1></div>
    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Why the "Impact" Font?</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Memes traditionally use the <strong>Impact</strong> typeface because its thick, condensed letters are easy to read. Combined with a white fill and black outline (stroke), the text is visible on any background color.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Privacy First</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Just like our other tools, this runs entirely in your browser. Your photos are never uploaded to our servers.</p>
            </div>
        </div>
    </div>
    <div class="tool-layout">
        <aside>
            <div style="background: var(--color-surface); padding: var(--space-md); border-radius: var(--radius-md); border: 1px solid var(--color-border);">
                <h3 style="margin-bottom: 1rem; font-size: 1rem;">1. Upload Image</h3>
                <input type="file" id="imageInput" accept="image/*" style="width: 100%; margin-bottom: 1.5rem;">
                <h3 style="margin-bottom: 1rem; font-size: 1rem;">2. Add Text</h3>
                <div style="margin-bottom: 1rem;">
                    <div style="display:flex; justify-content:space-between; align-items:center;">
                        <label style="font-size: 0.85rem; font-weight: 600;">Top Text</label>
                        <div class="color-toggle">
                            <button class="color-btn active" style="background:#fff;" onclick="setColor('top', 'white', this)" title="White Text"></button>
                            <button class="color-btn" style="background:#000;" onclick="setColor('top', 'black', this)" title="Black Text"></button>
                        </div>
                    </div>
                    <input type="text" id="topText" placeholder="WHEN THE CODE" style="width: 100%; padding: 0.5rem; margin-top: 0.25rem;">
                </div>
                <div style="margin-bottom: 1.5rem;">
                    <div style="display:flex; justify-content:space-between; align-items:center;">
                        <label style="font-size: 0.85rem; font-weight: 600;">Bottom Text</label>
                        <div class="color-toggle">
                            <button class="color-btn active" style="background:#fff;" onclick="setColor('bottom', 'white', this)" title="White Text"></button>
                            <button class="color-btn" style="background:#000;" onclick="setColor('bottom', 'black', this)" title="Black Text"></button>
                        </div>
                    </div>
                    <input type="text" id="bottomText" placeholder="WORKS ON FIRST TRY" style="width: 100%; padding: 0.5rem; margin-top: 0.25rem;">
                </div>
                <div style="margin-bottom: 1.5rem;">
                    <label style="font-size: 0.85rem; font-weight: 600;">Text Size (px)</label>
                    <input type="range" id="textSize" min="20" max="100" value="40" style="width: 100%;">
                </div>
                <div style="display: flex; gap: 0.5rem;">
                    <button class="btn" onclick="resetAll()" style="flex:1; background: #666;">Reset</button>
                    <button class="btn" id="downloadBtn" style="flex:2; background: var(--color-accent);" disabled>Download Meme</button>
                </div>
            </div>
        </aside>
        <section>
            <div class="canvas-container"><canvas id="memeCanvas" width="600" height="400"></canvas></div>
            <p style="text-align: center; color: var(--color-text-muted, #666); font-size: 0.85rem; margin-top: 1rem;">Tip: Use uppercase for the classic meme look.</p>
        </section>
    </div>
</main>

<script>
    const canvas = document.getElementById('memeCanvas');
    const ctx = canvas.getContext('2d');
    const imageInput = document.getElementById('imageInput');
    const topText = document.getElementById('topText');
    const bottomText = document.getElementById('bottomText');
    const textSize = document.getElementById('textSize');
    const downloadBtn = document.getElementById('downloadBtn');
    let currentImage = null;
    let colors = { top: 'white', bottom: 'white' };
    function initCanvas() {
        ctx.fillStyle = "#ccc"; ctx.fillRect(0,0, canvas.width, canvas.height);
        ctx.fillStyle = "#666"; ctx.font = "20px sans-serif"; ctx.textAlign = "center";
        ctx.fillText("Upload an image to start", 300, 200);
    }
    initCanvas();
    imageInput.addEventListener('change', (e) => {
        const file = e.target.files[0];
        if(file) {
            const reader = new FileReader();
            reader.onload = (event) => {
                currentImage = new Image();
                currentImage.onload = () => {
                    const maxW = 800;
                    const scale = Math.min(1, maxW / currentImage.width);
                    canvas.width = currentImage.width * scale;
                    canvas.height = currentImage.height * scale;
                    draw(); downloadBtn.disabled = false;
                };
                currentImage.src = event.target.result;
            };
            reader.readAsDataURL(file);
        }
    });
    window.setColor = (pos, color, btn) => {
        colors[pos] = color;
        const parent = btn.parentElement;
        Array.from(parent.children).forEach(c => c.classList.remove('active'));
        btn.classList.add('active'); draw();
    };
    function draw() {
        if(!currentImage) return;
        ctx.clearRect(0,0, canvas.width, canvas.height);
        ctx.drawImage(currentImage, 0, 0, canvas.width, canvas.height);
        const size = textSize.value;
        ctx.font = '900 ' + size + 'px Impact, sans-serif';
        ctx.lineWidth = size / 25; ctx.textAlign = 'center';
        if(topText.value) {
            ctx.textBaseline = 'top'; ctx.fillStyle = colors.top;
            ctx.strokeStyle = colors.top === 'white' ? 'black' : 'white';
            ctx.fillText(topText.value.toUpperCase(), canvas.width/2, canvas.height*0.05);
            ctx.strokeText(topText.value.toUpperCase(), canvas.width/2, canvas.height*0.05);
        }
        if(bottomText.value) {
            ctx.textBaseline = 'bottom'; ctx.fillStyle = colors.bottom;
            ctx.strokeStyle = colors.bottom === 'white' ? 'black' : 'white';
            ctx.fillText(bottomText.value.toUpperCase(), canvas.width/2, canvas.height*0.95);
            ctx.strokeText(bottomText.value.toUpperCase(), canvas.width/2, canvas.height*0.95);
        }
    }
    window.resetAll = () => {
        currentImage = null; imageInput.value = ""; topText.value = ""; bottomText.value = "";
        textSize.value = 40; downloadBtn.disabled = true; colors = { top: 'white', bottom: 'white' };
        document.querySelectorAll('.color-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.color-toggle button:first-child').forEach(b => b.classList.add('active'));
        canvas.width = 600; canvas.height = 400; initCanvas();
    };
    downloadBtn.addEventListener('click', () => {
        const link = document.createElement('a'); link.download = 'meme.jpg';
        link.href = canvas.toDataURL('image/jpeg', 0.9); link.click();
    });
    topText.addEventListener('input', draw);
    bottomText.addEventListener('input', draw);
    textSize.addEventListener('input', draw);
</script>$meme$,
    updated_at = NOW()
WHERE id = '6ae53f32-be86-4c29-bc52-983c35d23b18'
  AND html_template = '';


-- 3. PROMPT ARCHITECT
UPDATE content_components
SET html_template = $prompt$<style>
    .tool-layout { display: grid; gap: var(--space-lg); }
    @media (min-width: 900px) { .tool-layout { grid-template-columns: 1fr 1fr; } }
    .param-group { background: var(--color-surface, #fff); padding: 1rem; border: 1px solid var(--color-border); border-radius: 8px; margin-bottom: 1rem; }
    .param-title { font-weight: 700; margin-bottom: 0.5rem; display: flex; justify-content: space-between; }
    .param-desc { font-size: 0.8rem; color: var(--color-text-muted, #666); margin-bottom: 0.8rem; }
    select, textarea { width: 100%; padding: 0.5rem; border: 1px solid var(--color-border, #ccc); border-radius: 4px; font-size: 0.9rem; }
    textarea { height: 100px; }
    .tag-cloud { display: flex; flex-wrap: wrap; gap: 0.5rem; }
    .tag { padding: 4px 10px; border: 1px solid var(--color-border, #ddd); border-radius: 20px; font-size: 0.8rem; cursor: pointer; transition: all 0.2s; }
    .tag.active { background: var(--color-accent); color: white; border-color: var(--color-accent); }
    .final-box { background: #1e1e1e; color: #a9ff68; padding: 1.5rem; border-radius: 8px; position: sticky; top: 20px; }
</style>

<main class="container" style="padding-top: var(--space-lg);">
    <h1 style="font-size: var(--text-h2);">AI Prompt Architect</h1>
    <div class="guide-box" style="background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card);">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Speak the Language of Physics</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">AI models mimic photography. To get good results, act like a Director of Photography. Don't just say "cool lighting." Ask for <strong>"Volumetric Lighting"</strong> (dusty beams of light) or <strong>"Rim Lighting"</strong> (glowing edges).</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Camera Lenses Matter</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);"><strong>Wide Angle (16mm)</strong> creates epic scale but distorts faces.<br><strong>Telephoto (85mm+)</strong> flattens faces and blurs backgrounds (Bokeh).</p>
            </div>
        </div>
    </div>
    <div class="tool-layout">
        <div>
            <div class="param-group">
                <div class="param-title">1. The Subject</div>
                <textarea id="coreSubject" placeholder="e.g. A futuristic cyberpunk detective standing in the rain..."></textarea>
            </div>
            <div class="param-group">
                <div class="param-title">2. Lighting</div>
                <p class="param-desc">How does the light interact with the scene?</p>
                <div class="tag-cloud" id="lightingTags">
                    <span class="tag" data-val="Volumetric lighting">Volumetric (God Rays)</span>
                    <span class="tag" data-val="Cinematic lighting">Cinematic</span>
                    <span class="tag" data-val="Bioluminescent">Bioluminescent</span>
                    <span class="tag" data-val="Rim lighting">Rim Lighting</span>
                    <span class="tag" data-val="Golden hour">Golden Hour</span>
                    <span class="tag" data-val="Softbox lighting">Softbox (Studio)</span>
                </div>
            </div>
            <div class="param-group">
                <div class="param-title">3. Camera & Lens</div>
                <p class="param-desc">Define the perspective and depth of field.</p>
                <div class="tag-cloud" id="cameraTags">
                    <span class="tag" data-val="Wide angle lens">Wide Angle (Epic)</span>
                    <span class="tag" data-val="85mm lens">85mm (Portrait)</span>
                    <span class="tag" data-val="Macro lens">Macro (Tiny Details)</span>
                    <span class="tag" data-val="Drone view">Drone View</span>
                    <span class="tag" data-val="Bokeh">Bokeh (Blurred BG)</span>
                    <span class="tag" data-val="GoPro footage">GoPro Fisheye</span>
                </div>
            </div>
            <div class="param-group">
                <div class="param-title">4. Engine & Style</div>
                <div class="tag-cloud" id="styleTags">
                    <span class="tag" data-val="Unreal Engine 5">Unreal Engine 5</span>
                    <span class="tag" data-val="Octane Render">Octane Render</span>
                    <span class="tag" data-val="Oil painting">Oil Painting</span>
                    <span class="tag" data-val="Synthwave">Synthwave</span>
                    <span class="tag" data-val="Ukiyo-e">Ukiyo-e (Japanese)</span>
                    <span class="tag" data-val="Pixar style">Pixar 3D</span>
                </div>
            </div>
            <div class="param-group">
                <div class="param-title">5. Midjourney Params</div>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                    <label>Aspect Ratio
                        <select id="arParam"><option value="">Square (1:1)</option><option value="--ar 16:9">Cinema (16:9)</option><option value="--ar 9:16">Phone (9:16)</option><option value="--ar 2:1">Ultra Wide (2:1)</option></select>
                    </label>
                    <label>Stylize (0-1000) <input type="number" id="sParam" value="250" step="50"></label>
                </div>
            </div>
        </div>
        <div>
            <div class="final-box">
                <h3 style="color: #fff; margin-bottom: 1rem; font-size: 1rem;">Final Prompt</h3>
                <div id="finalOutput" style="font-family: var(--font-mono); line-height: 1.6;">...</div>
                <button class="btn" onclick="copyPrompt()" style="width: 100%; margin-top: 1rem; background: #fff; color: #111;">Copy to Clipboard</button>
            </div>
        </div>
    </div>
</main>

<script>
    const state = { subject: "", lighting: [], camera: [], style: [], ar: "", stylize: "250" };
    const output = document.getElementById('finalOutput');
    function updateOutput() {
        let parts = [];
        if (state.subject) parts.push(state.subject);
        const mods = [...state.style, ...state.lighting, ...state.camera];
        if (mods.length > 0) parts.push(mods.join(", "));
        let params = "";
        if (state.ar) params += " " + state.ar;
        if (state.stylize) params += " --s " + state.stylize;
        output.innerText = parts.join(", ") + params;
    }
    function toggleTag(el, category) {
        const val = el.dataset.val;
        if (state[category].includes(val)) {
            state[category] = state[category].filter(item => item !== val);
            el.classList.remove('active');
        } else { state[category].push(val); el.classList.add('active'); }
        updateOutput();
    }
    document.getElementById('coreSubject').addEventListener('input', (e) => { state.subject = e.target.value; updateOutput(); });
    document.getElementById('lightingTags').addEventListener('click', (e) => { if(e.target.classList.contains('tag')) toggleTag(e.target, 'lighting'); });
    document.getElementById('cameraTags').addEventListener('click', (e) => { if(e.target.classList.contains('tag')) toggleTag(e.target, 'camera'); });
    document.getElementById('styleTags').addEventListener('click', (e) => { if(e.target.classList.contains('tag')) toggleTag(e.target, 'style'); });
    document.getElementById('arParam').addEventListener('change', (e) => { state.ar = e.target.value; updateOutput(); });
    document.getElementById('sParam').addEventListener('input', (e) => { state.stylize = e.target.value; updateOutput(); });
    function copyPrompt() { navigator.clipboard.writeText(output.innerText); alert("Copied!"); }
</script>$prompt$,
    updated_at = NOW()
WHERE id = '2c941ec2-b59e-4e0f-925d-0b4d05ce8959'
  AND html_template = '';


-- Verify
SELECT function, display_name, LENGTH(html_template) as template_len
FROM content_components
WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true
ORDER BY function;
-- ============================================================
-- Part 2: Bayesian Ranking + Background Remover
-- ============================================================

-- 4. BAYESIAN RANKING (bayes.js inlined)
UPDATE content_components
SET html_template = $bayes$<style>
    .tool-layout { display: grid; grid-template-columns: 1fr; gap: var(--space-lg); align-items: start; }
    @media (min-width: 900px) { .tool-layout { grid-template-columns: 350px 1fr; } }
    .product-comparison { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-md); margin-bottom: var(--space-lg); }
    .prod-card {
        background: var(--color-surface, #fff); border: 1px solid var(--color-border);
        padding: var(--space-md); border-radius: var(--radius-md); text-align: center;
        position: relative; transition: transform 0.2s, box-shadow 0.2s;
    }
    .prod-card.winner { border-color: #10b981; background: #ecfdf5; transform: scale(1.02); box-shadow: 0 10px 25px rgba(16, 185, 129, 0.15); z-index: 1; }
    .star-display { font-size: 1.5rem; color: #f59e0b; margin: 0.5rem 0; }
    .score-badge { background: #1e1e1e; color: #fff; padding: 0.25rem 0.5rem; border-radius: 4px; font-family: var(--font-mono); font-size: 0.8rem; }
    .formula-box {
        background: var(--color-surface); padding: 1rem; border-radius: var(--radius-sm);
        font-family: var(--font-mono); font-size: 0.85rem; margin-top: 1rem; color: var(--color-text-muted, #555); overflow-x: auto;
    }
</style>

<main class="container" style="padding-top: var(--space-lg);">
    <div style="margin-bottom: var(--space-lg);"><h1 style="font-size: var(--text-h2);">Bayesian Ranking Calculator</h1></div>
    <div style="background: var(--color-surface, white); border: 1px solid var(--color-border); border-radius: var(--radius-md); padding: var(--space-md); margin-bottom: var(--space-lg); box-shadow: var(--shadow-card);">
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: var(--space-lg);">
            <div>
                <h3 style="font-size: 1rem; margin-bottom: 0.5rem;">The "5-Star" Illusion</h3>
                <p style="font-size: 0.95rem; color: var(--color-text-muted, #555); line-height: 1.6;">Product A has one 5-star review. Product B has one hundred 4.8-star reviews. A simple average says A > B. This is statistically wrong.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">The Bayesian Solution</h3>
                <p style="font-size: 0.95rem; color: var(--color-text-muted, #555); line-height: 1.6;">We add "Confidence" to the score. We inject a small amount of dummy data (the global average) to pull outliers back to reality until they prove themselves with volume.</p>
            </div>
        </div>
    </div>
    <div class="tool-layout">
        <aside>
            <div style="background: var(--color-surface); padding: var(--space-md); border-radius: var(--radius-md); border: 1px solid var(--color-border);">
                <h3 style="margin-bottom: 1rem; font-size: 1rem;">1. Product A (The Outlier)</h3>
                <div style="margin-bottom: 1rem;"><label>Average Rating (0-5)</label><input type="number" id="ratingA" value="5.0" step="0.1" max="5"></div>
                <div style="margin-bottom: 1rem;"><label>Review Count</label><input type="number" id="countA" value="1"></div>
                <hr style="border: 0; border-top: 1px solid var(--color-border, #ccc); margin: 1.5rem 0;">
                <h3 style="margin-bottom: 1rem; font-size: 1rem;">2. Product B (The Veteran)</h3>
                <div style="margin-bottom: 1rem;"><label>Average Rating (0-5)</label><input type="number" id="ratingB" value="4.8" step="0.1" max="5"></div>
                <div style="margin-bottom: 1rem;"><label>Review Count</label><input type="number" id="countB" value="100"></div>
                <hr style="border: 0; border-top: 1px solid var(--color-border, #ccc); margin: 1.5rem 0;">
                <h3 style="margin-bottom: 1rem; font-size: 1rem;">3. The Constant (C)</h3>
                <p style="font-size: 0.8rem; margin-bottom: 0.5rem;">Confidence Level (Dummy Reviews)</p>
                <input type="range" id="confidence" min="1" max="50" value="10">
                <div style="text-align: right; font-size: 0.8rem; font-weight: 700;" id="confDisplay">C = 10</div>
            </div>
        </aside>
        <section>
            <h3 style="margin-bottom: 1rem;">Who Ranks Higher?</h3>
            <div class="product-comparison">
                <div class="prod-card" id="cardA">
                    <strong style="display:block; font-size: 1.1rem;">Product A</strong>
                    <div class="star-display">★★★★★</div>
                    <p style="font-size: 0.9rem; color: var(--color-text-muted, #666);"><span id="txtRatingA">5.0</span> stars<br>(<span id="txtCountA">1</span> reviews)</p>
                    <div style="margin-top: 1rem;"><span style="display:block; font-size:0.75rem; text-transform:uppercase; color:var(--color-text-muted, #888);">Bayesian Score</span><span class="score-badge" id="bayesA">--</span></div>
                </div>
                <div class="prod-card" id="cardB">
                    <strong style="display:block; font-size: 1.1rem;">Product B</strong>
                    <div class="star-display">★★★★☆</div>
                    <p style="font-size: 0.9rem; color: var(--color-text-muted, #666);"><span id="txtRatingB">4.8</span> stars<br>(<span id="txtCountB">100</span> reviews)</p>
                    <div style="margin-top: 1rem;"><span style="display:block; font-size:0.75rem; text-transform:uppercase; color:var(--color-text-muted, #888);">Bayesian Score</span><span class="score-badge" id="bayesB">--</span></div>
                </div>
            </div>
            <div class="formula-box">
                <strong>The Math:</strong><br>
                Score = ( (R × v) + (C × m) ) / (v + C)<br><br>
                R = Average Rating of Item<br>v = Vote Count of Item<br>m = Global Average (Assumed 3.5)<br>C = Confidence Constant (<span id="mathC">10</span>)
            </div>
        </section>
    </div>
</main>

<script>
    // Bayesian Average Engine (inlined from bayes.js)
    const inputs = {
        rA: document.getElementById('ratingA'), cA: document.getElementById('countA'),
        rB: document.getElementById('ratingB'), cB: document.getElementById('countB'),
        conf: document.getElementById('confidence')
    };
    const display = {
        cardA: document.getElementById('cardA'), cardB: document.getElementById('cardB'),
        bayesA: document.getElementById('bayesA'), bayesB: document.getElementById('bayesB'),
        txtRatingA: document.getElementById('txtRatingA'), txtCountA: document.getElementById('txtCountA'),
        txtRatingB: document.getElementById('txtRatingB'), txtCountB: document.getElementById('txtCountB'),
        confDisplay: document.getElementById('confDisplay'), mathC: document.getElementById('mathC')
    };
    const M = 3.5;
    function calculate() {
        const rA = parseFloat(inputs.rA.value); const vA = parseInt(inputs.cA.value);
        const rB = parseFloat(inputs.rB.value); const vB = parseInt(inputs.cB.value);
        const C = parseInt(inputs.conf.value);
        display.txtRatingA.innerText = rA; display.txtCountA.innerText = vA;
        display.txtRatingB.innerText = rB; display.txtCountB.innerText = vB;
        display.confDisplay.innerText = "C = " + C; display.mathC.innerText = C;
        const scoreA = ((rA * vA) + (M * C)) / (vA + C);
        const scoreB = ((rB * vB) + (M * C)) / (vB + C);
        display.bayesA.innerText = scoreA.toFixed(3);
        display.bayesB.innerText = scoreB.toFixed(3);
        display.cardA.classList.remove('winner'); display.cardB.classList.remove('winner');
        if (scoreA > scoreB) { display.cardA.classList.add('winner'); }
        else if (scoreB > scoreA) { display.cardB.classList.add('winner'); }
    }
    Object.values(inputs).forEach(input => input.addEventListener('input', calculate));
    calculate();
</script>$bayes$,
    updated_at = NOW()
WHERE id = 'c345a76a-2c46-42b7-be16-1e34b4b19594'
  AND html_template = '';


-- 5. BACKGROUND REMOVER
UPDATE content_components
SET html_template = $bgrem$<style>
    .tool-layout { display: grid; gap: var(--space-lg); }
    @media (min-width: 900px) { .tool-layout { grid-template-columns: 300px 1fr; } }
    .canvas-container {
        position: relative;
        background-image: linear-gradient(45deg, #ccc 25%, transparent 25%),
        linear-gradient(-45deg, #ccc 25%, transparent 25%),
        linear-gradient(45deg, transparent 75%, #ccc 75%),
        linear-gradient(-45deg, transparent 75%, #ccc 75%);
        background-size: 20px 20px;
        background-position: 0 0, 0 10px, 10px -10px, -10px 0px;
        background-color: #fff; border-radius: 8px; overflow: hidden;
        box-shadow: inset 0 0 20px rgba(0,0,0,0.1); min-height: 400px;
        display: flex; align-items: center; justify-content: center;
    }
    canvas { display: block; max-width: 100%; height: auto; }
    .tool-btn {
        display: flex; align-items: center; justify-content: center; gap: 0.5rem;
        width: 100%; padding: 0.75rem; border: 1px solid var(--color-border);
        background: var(--color-surface, #fff); border-radius: 6px; cursor: pointer; font-weight: 600; font-size: 0.9rem;
        margin-bottom: 0.5rem; transition: all 0.2s;
    }
    .tool-btn.active { background: var(--color-accent); color: white; border-color: var(--color-accent); }
    .tool-btn:hover:not(.active) { background: var(--color-surface, #f4f4f5); }
    .guide-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card); }
</style>

<main class="container" style="padding-top: var(--space-lg);">
    <h1 style="font-size: var(--text-h2);">Magic Background Eraser</h1>
    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Magic Wand (Auto)</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Click any color (like a white wall) to instantly remove it. Use the <strong>Tolerance</strong> slider to grab shadows and similar shades.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Manual Brush (Precision)</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">For tricky edges where colors are too similar, switch to the <strong>Eraser Brush</strong> to manually wipe away pixels.</p>
            </div>
        </div>
    </div>
    <div class="tool-layout">
        <aside>
            <div style="background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px;">
                <h3 style="font-size: 1rem; margin-bottom: 1rem;">1. Upload Image</h3>
                <input type="file" id="fileInput" accept="image/*" style="width: 100%; margin-bottom: 1.5rem;">
                <h3 style="font-size: 1rem; margin-bottom: 0.5rem;">2. Select Tool</h3>
                <button class="tool-btn active" id="btnWand" onclick="setTool('wand')"><span>🪄</span> Magic Wand (Auto)</button>
                <button class="tool-btn" id="btnBrush" onclick="setTool('brush')"><span>🧹</span> Manual Eraser</button>
                <div id="wandSettings" style="margin-top: 1rem;">
                    <label style="font-size: 0.8rem; font-weight: 600;">Wand Tolerance</label>
                    <div style="display: flex; align-items: center; gap: 10px;">
                        <input type="range" id="tolerance" min="0" max="100" value="30" style="flex:1;">
                        <span id="tolVal" style="font-size: 0.8rem; width: 30px;">30</span>
                    </div>
                </div>
                <div id="brushSettings" style="margin-top: 1rem; display: none;">
                    <label style="font-size: 0.8rem; font-weight: 600;">Brush Size</label>
                    <div style="display: flex; align-items: center; gap: 10px;">
                        <input type="range" id="brushSize" min="5" max="100" value="30" style="flex:1;">
                        <span id="brushVal" style="font-size: 0.8rem; width: 30px;">30px</span>
                    </div>
                </div>
                <hr style="border: 0; border-top: 1px solid var(--color-border, #eee); margin: 1.5rem 0;">
                <div style="display: flex; gap: 0.5rem; margin-bottom: 0.5rem;">
                    <button class="btn" onclick="undo()" id="btnUndo" disabled style="flex:1; background: #666;">⎌ Undo</button>
                    <button class="btn" onclick="restoreOriginal()" style="flex:1; background: #333;">Reset All</button>
                </div>
                <button class="btn" onclick="downloadPNG()" style="width: 100%; background: var(--color-accent);">Download PNG</button>
            </div>
        </aside>
        <section>
            <div class="canvas-container" id="canvasContainer"><canvas id="editorCanvas"></canvas></div>
            <p id="helperText" style="text-align: center; color: var(--color-text-muted, #666); font-size: 0.85rem; margin-top: 1rem;">Click a color to erase it.</p>
        </section>
    </div>
</main>

<script>
    const canvas = document.getElementById('editorCanvas');
    const ctx = canvas.getContext('2d', { willReadFrequently: true });
    const fileInput = document.getElementById('fileInput');
    const container = document.getElementById('canvasContainer');
    const toleranceInput = document.getElementById('tolerance');
    const brushSizeInput = document.getElementById('brushSize');
    const btnUndo = document.getElementById('btnUndo');
    const helperText = document.getElementById('helperText');
    let originalImg = new Image(); let history = []; let currentTool = 'wand'; let isDrawing = false;

    function setTool(tool) {
        currentTool = tool;
        document.getElementById('btnWand').classList.toggle('active', tool === 'wand');
        document.getElementById('btnBrush').classList.toggle('active', tool === 'brush');
        document.getElementById('wandSettings').style.display = tool === 'wand' ? 'block' : 'none';
        document.getElementById('brushSettings').style.display = tool === 'brush' ? 'block' : 'none';
        container.style.cursor = tool === 'wand' ? 'crosshair' : 'cell';
        helperText.innerText = tool === 'wand' ? "Click a color to auto-remove it." : "Click and drag to wipe away parts of the image.";
    }
    fileInput.addEventListener('change', (e) => {
        const file = e.target.files[0]; if(!file) return;
        const url = URL.createObjectURL(file);
        originalImg.onload = () => {
            let w = originalImg.width; let h = originalImg.height;
            if(w > 1000) { const scale = 1000/w; w = 1000; h = h * scale; }
            canvas.width = w; canvas.height = h;
            ctx.drawImage(originalImg, 0, 0, w, h); saveState();
        };
        originalImg.src = url; history = []; btnUndo.disabled = true;
    });
    function saveState() {
        if (history.length > 10) history.shift();
        history.push(ctx.getImageData(0, 0, canvas.width, canvas.height));
        btnUndo.disabled = false;
    }
    function undo() {
        if (history.length <= 1) return; history.pop();
        ctx.putImageData(history[history.length - 1], 0, 0);
        if(history.length === 1) btnUndo.disabled = true;
    }
    function restoreOriginal() {
        if(!originalImg.src) return; history = [];
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.drawImage(originalImg, 0, 0, canvas.width, canvas.height); saveState();
    }
    canvas.addEventListener('mousedown', (e) => {
        if (!originalImg.src) return; isDrawing = true;
        const pos = getPos(e);
        if (currentTool === 'wand') { removeBackgroundWand(pos.x, pos.y); saveState(); }
        else { eraseBrush(pos.x, pos.y); }
    });
    canvas.addEventListener('mousemove', (e) => { if (!isDrawing || currentTool !== 'brush') return; eraseBrush(getPos(e).x, getPos(e).y); });
    canvas.addEventListener('mouseup', () => { if (isDrawing && currentTool === 'brush') saveState(); isDrawing = false; });
    canvas.addEventListener('mouseleave', () => { if (isDrawing && currentTool === 'brush') saveState(); isDrawing = false; });
    function getPos(e) {
        const rect = canvas.getBoundingClientRect();
        return { x: Math.floor((e.clientX - rect.left) * (canvas.width / rect.width)), y: Math.floor((e.clientY - rect.top) * (canvas.height / rect.height)) };
    }
    function eraseBrush(x, y) {
        const size = parseInt(brushSizeInput.value);
        ctx.globalCompositeOperation = 'destination-out';
        ctx.beginPath(); ctx.arc(x, y, size / 2, 0, Math.PI * 2); ctx.fill();
        ctx.globalCompositeOperation = 'source-over';
    }
    function removeBackgroundWand(startX, startY) {
        const w = canvas.width; const h = canvas.height;
        const imageData = ctx.getImageData(0, 0, w, h); const data = imageData.data;
        const startIdx = (startY * w + startX) * 4;
        const r0 = data[startIdx], g0 = data[startIdx+1], b0 = data[startIdx+2], a0 = data[startIdx+3];
        if (a0 === 0) return;
        const tol = parseInt(toleranceInput.value);
        const visited = new Uint8Array(w * h); const stack = [startIdx];
        const matches = (idx) => {
            if (data[idx+3] === 0) return false;
            return (Math.abs(data[idx]-r0) + Math.abs(data[idx+1]-g0) + Math.abs(data[idx+2]-b0)) <= (tol * 3);
        };
        while(stack.length > 0) {
            const idx = stack.pop(); const pi = idx / 4;
            if (visited[pi]) continue; visited[pi] = 1; data[idx+3] = 0;
            const x = pi % w; const y = Math.floor(pi / w);
            if (x > 0 && matches(idx-4)) stack.push(idx-4);
            if (x < w-1 && matches(idx+4)) stack.push(idx+4);
            if (y > 0 && matches(idx-w*4)) stack.push(idx-w*4);
            if (y < h-1 && matches(idx+w*4)) stack.push(idx+w*4);
        }
        ctx.putImageData(imageData, 0, 0);
    }
    function downloadPNG() {
        const link = document.createElement('a'); link.download = 'erased-image.png';
        link.href = canvas.toDataURL('image/png'); link.click();
    }
    toleranceInput.addEventListener('input', () => document.getElementById('tolVal').innerText = toleranceInput.value);
    brushSizeInput.addEventListener('input', () => document.getElementById('brushVal').innerText = brushSizeInput.value + 'px');
</script>$bgrem$,
    updated_at = NOW()
WHERE id = 'bdd2990a-1cc3-47d4-8140-4560204e898c'
  AND html_template = '';


-- Verify all 5 updates
SELECT function, display_name, LENGTH(html_template) as template_len
FROM content_components
WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true
ORDER BY function;


---
-- tool routes and affinity fix

-- ============================================================
-- Route registration snippet (add to core-manager server setup)
-- ============================================================
-- Add after the assetAdminHandlers lines (~line 18879):
--
-- // Tool management
-- toolAdminHandlers := admin.NewToolAdminHandlers(clientsDB, appLogger)
-- siteGroup.GET("/:site_id/tools", toolAdminHandlers.HandleListTools)
-- siteGroup.DELETE("/:site_id/tools/:function", toolAdminHandlers.HandleRemoveTool)
-- siteGroup.POST("/:site_id/tools", toolAdminHandlers.HandleDeployTool)
--
-- // Library tools (not site-specific)
-- adminRoutes.GET("/tools/library", toolAdminHandlers.HandleListLibraryTools)
--
-- Where adminRoutes is the parent group that siteGroup belongs to.


-- ============================================================
-- Narrow password-entropy tool affinity
-- ============================================================
-- The password checker was deployed to 4 sites (including gaswholesalers)
-- because the library only had 2 tools with templates, giving the LLM
-- no real choice. Now that we have 7 tools, the LLM has better options.
-- But we should also narrow the semantic_tags so the matching logic
-- (and LLM context) makes it clear this is a niche tool.

UPDATE content_components
SET semantic_tags = '["calculator", "security", "password", "entropy", "privacy", "tech", "cybersecurity", "developer"]'::jsonb,
    description = 'Shannon entropy calculation, GPU crack time estimation, dictionary attack heuristic warning. Best suited for tech, cybersecurity, and developer-focused sites.',
    updated_at = NOW()
WHERE function = 'tool-password-entropy'
  AND forked_from IS NULL;

-- Also narrow the A/B test calculator — it's marketing/ecommerce specific
UPDATE content_components
SET semantic_tags = '["calculator", "statistics", "marketing", "ab-testing", "conversion", "ecommerce", "analytics"]'::jsonb,
    description = 'Z-score test for conversion rate differences with 95% confidence interval. Best suited for marketing, ecommerce, and analytics-focused sites.',
    updated_at = NOW()
WHERE function = 'tool-ab-test-calculator'
  AND forked_from IS NULL;


-- ============================================================
-- Verify: check which sites have password-entropy deployed
-- ============================================================
SELECT s.domain, cc.function, cc.display_name, p.url
FROM content_components cc
JOIN page_components pc ON pc.component_id = cc.id
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON s.id = p.site_id
WHERE cc.function = 'tool-password-entropy'
  AND cc.is_active = true
ORDER BY s.domain;

--

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
-- backfill with tags for new site_type categories
-- Migration: Add component selection metadata to content_components
-- Part of the component selector infrastructure (doc 028c)
--
-- This adds the columns that enable:
-- 1. section_type: abstract section purpose (hero, tool-grid, provocation-card)
--    decoupled from function (the specific template name)
-- 2. Selection metadata: suitable_site_types, suitable_page_types, content_shape,
--    visual_density — used by the component selector to score candidates
-- 3. Usage tracking: usage_count, avg_quality_score — feedback from deployments
--    and auditors, used in scoring
-- 4. Provenance: created_from — how this component entered the library
--
-- All columns are nullable with sensible defaults. Existing components continue
-- to work without any data in these columns. The selector falls back to
-- function-name matching when section_type is NULL.
--
-- Run the data backfill (next migration) after this to tag existing components.

-- ============================================================================
-- New columns
-- ============================================================================

-- section_type: the abstract section purpose this component implements
-- Multiple components can share a section_type (they're variants)
-- e.g. hero-split and hero-fullwidth both have section_type = 'hero'
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS section_type text;

-- suitable_site_types: which site types this component fits
-- JSONB array of strings, e.g. ["brochure", "saas", "interactive-platform"]
-- The selector scores higher when the site's type is in this array
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS suitable_site_types jsonb DEFAULT '[]'::jsonb;

-- suitable_page_types: which page types this component fits
-- JSONB array of strings, e.g. ["landing", "product-listing", "about"]
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS suitable_page_types jsonb DEFAULT '[]'::jsonb;

-- content_shape: what kind of content structure the component expects
-- Free text, e.g. "prose", "structured_list", "structured_card", "key_value_pairs"
-- Helps the selector match components to content that's available
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS content_shape text;

-- visual_density: how much content the component packs in
-- "low" (hero, CTA), "medium" (features grid), "high" (data table, tool grid)
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS visual_density text;

-- usage_count: how many times this component has been assigned to a page
-- Incremented by the selector when it picks this component
-- Higher usage = more battle-tested, weighted in scoring
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS usage_count integer DEFAULT 0;

-- avg_quality_score: average quality score from auditor feedback
-- NULL = unproven (new component, never audited)
-- 0.0-1.0 range, updated by auditors after each deployment
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS avg_quality_score float;

-- created_from: how this component entered the library
-- 'manual' = hand-crafted by a developer
-- 'generated' = created by the component-creator from a needs_new_component work item
-- 'adopted' = discovered during site adoption and stored
-- Useful for tracking quality by provenance and for UI filtering
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS created_from text DEFAULT 'manual';

-- ============================================================================
-- Indexes for selector queries
-- ============================================================================

-- Primary selector query: find components by section_type
CREATE INDEX IF NOT EXISTS idx_cc_section_type
    ON content_components (section_type)
    WHERE section_type IS NOT NULL
      AND is_active = true
      AND forked_from IS NULL;

-- Filter by suitable site types (GIN index for JSONB containment queries)
CREATE INDEX IF NOT EXISTS idx_cc_suitable_site_types
    ON content_components USING gin (suitable_site_types)
    WHERE is_active = true
      AND forked_from IS NULL;

-- Combined index for the most common selector query pattern
-- "find active library components for this section_type"
CREATE INDEX IF NOT EXISTS idx_cc_selector
    ON content_components (section_type, component_level)
    WHERE is_active = true
      AND forked_from IS NULL
      AND section_type IS NOT NULL;

-- ============================================================================
-- Constraints
-- ============================================================================

-- section_type should follow kebab-case like function does
ALTER TABLE content_components
    ADD CONSTRAINT chk_section_type_kebab_case
    CHECK (section_type IS NULL OR section_type ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$');

-- created_from should be one of the known provenance types
ALTER TABLE content_components
    ADD CONSTRAINT chk_created_from_valid
    CHECK (created_from IS NULL OR created_from IN ('manual', 'generated', 'adopted'));

-- visual_density should be one of the known levels
ALTER TABLE content_components
    ADD CONSTRAINT chk_visual_density_valid
    CHECK (visual_density IS NULL OR visual_density IN ('low', 'medium', 'high'));

-- usage_count should never be negative
ALTER TABLE content_components
    ADD CONSTRAINT chk_usage_count_non_negative
    CHECK (usage_count IS NULL OR usage_count >= 0);

-- avg_quality_score should be 0.0-1.0
ALTER TABLE content_components
    ADD CONSTRAINT chk_quality_score_range
    CHECK (avg_quality_score IS NULL OR (avg_quality_score >= 0.0 AND avg_quality_score <= 1.0));

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON COLUMN content_components.section_type IS 'Abstract section purpose (hero, tool-grid, provocation-card). Multiple components can share a section_type as variants. The selector picks the best variant for a given site.';
COMMENT ON COLUMN content_components.suitable_site_types IS 'JSONB array of site types this component fits. Selector scores higher for site_type matches. e.g. ["brochure", "interactive-platform"]';
COMMENT ON COLUMN content_components.suitable_page_types IS 'JSONB array of page types this component fits. e.g. ["landing", "about", "tool-index"]';
COMMENT ON COLUMN content_components.content_shape IS 'What kind of content structure the component expects: prose, structured_list, structured_card, key_value_pairs';
COMMENT ON COLUMN content_components.visual_density IS 'How much content the component packs in: low (hero, CTA), medium (features), high (data table)';
COMMENT ON COLUMN content_components.usage_count IS 'Times this component has been assigned to a page. Incremented by selector. Higher = more battle-tested.';
COMMENT ON COLUMN content_components.avg_quality_score IS 'Average quality score from auditor feedback. NULL = unproven. 0.0-1.0 range.';
COMMENT ON COLUMN content_components.created_from IS 'How this component entered the library: manual (hand-crafted), generated (component-creator), adopted (site adoption)';


                                                     -- Migration: Backfill section_type and suitable_site_types for existing components
-- Run AFTER migration_component_selection_metadata.sql
--
-- Strategy:
-- 1. For most existing components, section_type = function (hero → hero, social-proof → social-proof)
--    because we currently have one template per purpose
-- 2. For page-specific variants (about-hero, services-hero), section_type = the base purpose (hero)
--    so the selector can find all hero variants when a page needs a "hero"
-- 3. suitable_site_types set broadly for existing components since they were all built
--    for brochure/professional sites
-- 4. created_from = 'manual' for all existing components (they were hand-crafted)
--
-- This is idempotent — uses WHERE section_type IS NULL so re-running is safe.

-- ============================================================================
-- Section-level components: set section_type from function patterns
-- ============================================================================

-- Hero variants: all map to section_type 'hero'
UPDATE content_components SET
    section_type = 'hero',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio"]'::jsonb,
    suitable_page_types = '["landing", "index"]'::jsonb,
    content_shape = 'prose',
    visual_density = 'low',
    created_from = 'manual'
WHERE function IN ('hero', 'hero-split', 'hero-fullwidth', 'hero-minimal', 'hero-gradient')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Page-specific heroes: section_type matches the page purpose
UPDATE content_components SET
    section_type = 'hero',
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = CASE
        WHEN function LIKE 'about-%' THEN '["about"]'::jsonb
        WHEN function LIKE 'services-%' THEN '["services"]'::jsonb
        WHEN function LIKE 'contact-%' THEN '["contact"]'::jsonb
        WHEN function LIKE 'case-studies-%' THEN '["case-studies", "use-cases"]'::jsonb
        ELSE '[]'::jsonb
    END,
    content_shape = 'prose',
    visual_density = 'low',
    created_from = 'manual'
WHERE function LIKE '%-hero'
  AND component_level = 'section'
  AND section_type IS NULL;

-- Social proof / testimonials
UPDATE content_components SET
    section_type = 'social-proof',
    suitable_site_types = '["brochure", "saas", "landing-page"]'::jsonb,
    suitable_page_types = '["landing", "index", "about"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('social-proof', 'testimonials')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Call to action
UPDATE content_components SET
    section_type = 'call-to-action',
    suitable_site_types = '["brochure", "saas", "landing-page", "interactive-platform"]'::jsonb,
    suitable_page_types = '["landing", "index", "about", "services"]'::jsonb,
    content_shape = 'prose',
    visual_density = 'low',
    created_from = 'manual'
WHERE function IN ('call-to-action', 'cta')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Features / differentiators
UPDATE content_components SET
    section_type = 'features',
    suitable_site_types = '["brochure", "saas", "landing-page"]'::jsonb,
    suitable_page_types = '["landing", "index"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('features', 'differentiators', 'differentiators-section')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Services grid
UPDATE content_components SET
    section_type = 'services-grid',
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = '["services", "index"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('services-grid')
  AND component_level = 'section'
  AND section_type IS NULL;

-- About content
UPDATE content_components SET
    section_type = 'about-content',
    suitable_site_types = '["brochure", "saas", "portfolio"]'::jsonb,
    suitable_page_types = '["about"]'::jsonb,
    content_shape = 'prose',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('about-content')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Leadership / team
UPDATE content_components SET
    section_type = 'team',
    suitable_site_types = '["brochure"]'::jsonb,
    suitable_page_types = '["about"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('leadership-team', 'team')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Case studies / use cases
UPDATE content_components SET
    section_type = 'case-studies',
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = '["case-studies", "use-cases"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('case-studies-list', 'case-studies')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Contact sections
UPDATE content_components SET
    section_type = CASE
        WHEN function = 'contact-form' THEN 'contact-form'
        WHEN function = 'contact-info' THEN 'contact-info'
        ELSE 'contact'
    END,
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = '["contact"]'::jsonb,
    content_shape = 'structured_card',
    visual_density = 'low',
    created_from = 'manual'
WHERE function LIKE 'contact%'
  AND component_level = 'section'
  AND section_type IS NULL;

-- Content site components
UPDATE content_components SET
    section_type = CASE
        WHEN function = 'content_listing' THEN 'content-listing'
        WHEN function = 'category_listing' THEN 'category-listing'
        WHEN function = 'featured_content' THEN 'featured-content'
        WHEN function = 'sidebar' THEN 'sidebar'
        WHEN function = 'advertising' THEN 'advertising'
        ELSE function
    END,
    suitable_site_types = '["content-site", "blog"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'high',
    created_from = 'manual'
WHERE category = 'content-site'
  AND component_level = 'section'
  AND section_type IS NULL;

-- ============================================================================
-- Catch-all: any remaining section components get section_type = function
-- This handles components not explicitly mapped above
-- ============================================================================

UPDATE content_components SET
    section_type = function,
    suitable_site_types = '["brochure"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'section'
  AND section_type IS NULL
  AND function != ''
  AND is_active = true
  AND forked_from IS NULL;

-- ============================================================================
-- Header/footer/head components: section_type matches their role
-- These aren't selected by the section selector but tagging them
-- keeps the data consistent
-- ============================================================================

UPDATE content_components SET
    section_type = 'header',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio", "content-site"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'header'
  AND section_type IS NULL;

UPDATE content_components SET
    section_type = 'footer',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio", "content-site"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'footer'
  AND section_type IS NULL;

UPDATE content_components SET
    section_type = 'head',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio", "content-site"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'head'
  AND section_type IS NULL;

-- ============================================================================
-- Tool components: section_type = 'tool', keep their existing category
-- ============================================================================

UPDATE content_components SET
    section_type = 'tool',
    created_from = 'manual'
WHERE component_level = 'tool'
  AND section_type IS NULL;

-- ============================================================================
-- Verification query — run manually to check results
-- ============================================================================

-- SELECT section_type, function, suitable_site_types, created_from, component_level
-- FROM content_components
-- WHERE is_active = true AND forked_from IS NULL
-- ORDER BY component_level, section_type, function;

-- Check for any that were missed
-- SELECT function, component_level FROM content_components
-- WHERE section_type IS NULL AND is_active = true AND forked_from IS NULL;

---
-- create our own components backfill
-- Migration: Backfill section_type and suitable_site_types for existing components
-- Run AFTER migration_component_selection_metadata.sql
--
-- Strategy:
-- 1. For most existing components, section_type = function (hero → hero, social-proof → social-proof)
--    because we currently have one template per purpose
-- 2. For page-specific variants (about-hero, services-hero), section_type = the base purpose (hero)
--    so the selector can find all hero variants when a page needs a "hero"
-- 3. suitable_site_types set broadly for existing components since they were all built
--    for brochure/professional sites
-- 4. created_from = 'manual' for all existing components (they were hand-crafted)
--
-- This is idempotent — uses WHERE section_type IS NULL so re-running is safe.

-- ============================================================================
-- Section-level components: set section_type from function patterns
-- ============================================================================

-- Hero variants: all map to section_type 'hero'
UPDATE content_components SET
    section_type = 'hero',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio"]'::jsonb,
    suitable_page_types = '["landing", "index"]'::jsonb,
    content_shape = 'prose',
    visual_density = 'low',
    created_from = 'manual'
WHERE function IN ('hero', 'hero-split', 'hero-fullwidth', 'hero-minimal', 'hero-gradient')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Page-specific heroes: section_type matches the page purpose
UPDATE content_components SET
    section_type = 'hero',
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = CASE
        WHEN function LIKE 'about-%' THEN '["about"]'::jsonb
        WHEN function LIKE 'services-%' THEN '["services"]'::jsonb
        WHEN function LIKE 'contact-%' THEN '["contact"]'::jsonb
        WHEN function LIKE 'case-studies-%' THEN '["case-studies", "use-cases"]'::jsonb
        ELSE '[]'::jsonb
    END,
    content_shape = 'prose',
    visual_density = 'low',
    created_from = 'manual'
WHERE function LIKE '%-hero'
  AND component_level = 'section'
  AND section_type IS NULL;

-- Social proof / testimonials
UPDATE content_components SET
    section_type = 'social-proof',
    suitable_site_types = '["brochure", "saas", "landing-page"]'::jsonb,
    suitable_page_types = '["landing", "index", "about"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('social-proof', 'testimonials')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Call to action
UPDATE content_components SET
    section_type = 'call-to-action',
    suitable_site_types = '["brochure", "saas", "landing-page", "interactive-platform"]'::jsonb,
    suitable_page_types = '["landing", "index", "about", "services"]'::jsonb,
    content_shape = 'prose',
    visual_density = 'low',
    created_from = 'manual'
WHERE function IN ('call-to-action', 'cta')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Features / differentiators
UPDATE content_components SET
    section_type = 'features',
    suitable_site_types = '["brochure", "saas", "landing-page"]'::jsonb,
    suitable_page_types = '["landing", "index"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('features', 'differentiators', 'differentiators-section')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Services grid
UPDATE content_components SET
    section_type = 'services-grid',
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = '["services", "index"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('services-grid')
  AND component_level = 'section'
  AND section_type IS NULL;

-- About content
UPDATE content_components SET
    section_type = 'about-content',
    suitable_site_types = '["brochure", "saas", "portfolio"]'::jsonb,
    suitable_page_types = '["about"]'::jsonb,
    content_shape = 'prose',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('about-content')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Leadership / team
UPDATE content_components SET
    section_type = 'team',
    suitable_site_types = '["brochure"]'::jsonb,
    suitable_page_types = '["about"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('leadership-team', 'team')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Case studies / use cases
UPDATE content_components SET
    section_type = 'case-studies',
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = '["case-studies", "use-cases"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'medium',
    created_from = 'manual'
WHERE function IN ('case-studies-list', 'case-studies')
  AND component_level = 'section'
  AND section_type IS NULL;

-- Contact sections
UPDATE content_components SET
    section_type = CASE
        WHEN function = 'contact-form' THEN 'contact-form'
        WHEN function = 'contact-info' THEN 'contact-info'
        ELSE 'contact'
    END,
    suitable_site_types = '["brochure", "saas"]'::jsonb,
    suitable_page_types = '["contact"]'::jsonb,
    content_shape = 'structured_card',
    visual_density = 'low',
    created_from = 'manual'
WHERE function LIKE 'contact%'
  AND component_level = 'section'
  AND section_type IS NULL;

-- Content site components
UPDATE content_components SET
    section_type = CASE
        WHEN function = 'content_listing' THEN 'content-listing'
        WHEN function = 'category_listing' THEN 'category-listing'
        WHEN function = 'featured_content' THEN 'featured-content'
        WHEN function = 'sidebar' THEN 'sidebar'
        WHEN function = 'advertising' THEN 'advertising'
        ELSE function
    END,
    suitable_site_types = '["content-site", "blog"]'::jsonb,
    content_shape = 'structured_list',
    visual_density = 'high',
    created_from = 'manual'
WHERE category = 'content-site'
  AND component_level = 'section'
  AND section_type IS NULL;

-- ============================================================================
-- Catch-all: any remaining section components get section_type = function
-- This handles components not explicitly mapped above
-- ============================================================================

UPDATE content_components SET
    section_type = function,
    suitable_site_types = '["brochure"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'section'
  AND section_type IS NULL
  AND function != ''
  AND is_active = true
  AND forked_from IS NULL;

-- ============================================================================
-- Header/footer/head components: section_type matches their role
-- These aren't selected by the section selector but tagging them
-- keeps the data consistent
-- ============================================================================

UPDATE content_components SET
    section_type = 'header',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio", "content-site"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'header'
  AND section_type IS NULL;

UPDATE content_components SET
    section_type = 'footer',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio", "content-site"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'footer'
  AND section_type IS NULL;

UPDATE content_components SET
    section_type = 'head',
    suitable_site_types = '["brochure", "saas", "landing-page", "portfolio", "content-site"]'::jsonb,
    created_from = 'manual'
WHERE component_level = 'head'
  AND section_type IS NULL;

-- ============================================================================
-- Tool components: section_type = 'tool', keep their existing category
-- ============================================================================

UPDATE content_components SET
    section_type = 'tool',
    created_from = 'manual'
WHERE component_level = 'tool'
  AND section_type IS NULL;

-- ============================================================================
-- Verification query — run manually to check results
-- ============================================================================

-- SELECT section_type, function, suitable_site_types, created_from, component_level
-- FROM content_components
-- WHERE is_active = true AND forked_from IS NULL
-- ORDER BY component_level, section_type, function;

-- Check for any that were missed
-- SELECT function, component_level FROM content_components
-- WHERE section_type IS NULL AND is_active = true AND forked_from IS NULL;

---
                                                     --
                                                     clients_db=# -- Consolidate page-specific heroes under section_type = 'hero'
-- with suitable_page_types indicating which pages they're designed for.
--
-- Before: hero-about has section_type = 'hero-about', selector treats it as a separate type
-- After:  hero-about has section_type = 'hero', suitable_page_types = '["about"]'
--         The selector finds ALL hero variants when a page needs a 'hero' section,
--         and scores higher for variants that match the page type.
--
-- The function column stays unchanged — hero-about is still hero-about.
-- Only section_type and suitable_page_types change.

UPDATE content_components SET
    section_type = 'hero',
    suitable_page_types = '["about"]'::jsonb
WHERE function = 'hero-about'
  AND is_active = true
  AND forked_from IS NULL;

UPDATE content_components SET
    section_type = 'hero',
    suitable_page_types = '["services"]'::jsonb
WHERE function = 'hero-services'
  AND is_active = true
  AND forked_from IS NULL;

UPDATE content_components SET
    section_type = 'hero',
    suitable_page_types = '["contact"]'::jsonb
WHERE function = 'hero-contact'
  AND is_active = true
  AND forked_from IS NULL;

UPDATE content_components SET
    section_type = 'hero',
    suitable_page_types = '["case-studies", "use-cases"]'::jsonb
WHERE function = 'hero-case-studies'
  AND is_active = true
-- other page-specific heroes should score lower (wrong page_type)ariant)EEN 1 AND 3DND
UPDATE 1
UPDATE 1
UPDATE 1
UPDATE 1
UPDATE 1
     function      | section_type |      suitable_page_types      |                suitable_site_types
-------------------+--------------+-------------------------------+---------------------------------------------------
 hero              | hero         | ["landing", "index"]          | ["brochure", "saas", "landing-page", "portfolio"]
 hero-about        | hero         | ["about"]                     | ["brochure"]
 hero-case-studies | hero         | ["case-studies", "use-cases"] | ["brochure"]
 hero-contact      | hero         | ["contact"]                   | ["brochure"]
 hero-services     | hero         | ["services"]                  | ["brochure"]
 hero-use-cases    | hero         | ["use-cases", "case-studies"] | ["brochure"]
(6 rows)

     function      |        score
-------------------+---------------------
 hero-about        |                0.69
 hero-contact      |  0.5399999999999999
 hero-services     |  0.5399999999999999
 hero-use-cases    |  0.5399999999999999
 hero-case-studies |  0.5399999999999999
 hero              | 0.45999999999999996
(6 rows)

