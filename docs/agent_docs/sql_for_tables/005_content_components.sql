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



