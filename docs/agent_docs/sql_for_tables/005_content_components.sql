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

