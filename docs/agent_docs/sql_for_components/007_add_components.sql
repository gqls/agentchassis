-- ============================================================================
-- Add missing content components that the site-planner LLM tends to generate
-- These are common section types needed for brochure/corporate websites
-- ============================================================================

-- Check component_level column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'content_components'
                   AND column_name = 'component_level') THEN
ALTER TABLE content_components ADD COLUMN component_level TEXT DEFAULT 'section';
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'content_components'
                   AND column_name = 'category') THEN
ALTER TABLE content_components ADD COLUMN category TEXT DEFAULT 'general';
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'content_components'
                   AND column_name = 'display_name') THEN
ALTER TABLE content_components ADD COLUMN display_name TEXT;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'content_components'
                   AND column_name = 'semantic_tags') THEN
ALTER TABLE content_components ADD COLUMN semantic_tags JSONB DEFAULT '[]'::jsonb;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'content_components'
                   AND column_name = 'is_active') THEN
ALTER TABLE content_components ADD COLUMN is_active BOOLEAN DEFAULT true;
END IF;
END $$;

-- ============================================================================
-- Hero sections (various page types)
-- ============================================================================

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'hero',
           'Hero Section',
           'Primary hero section with headline, subheadline and CTA',
           '<section class="hero" data-component="hero">
               <div class="hero-content">
                   <h1>{{.headline}}</h1>
                   <p class="hero-subheadline">{{.subheadline}}</p>
                   {{if .cta_text}}<a href="{{.cta_url}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
               </div>
           </section>',
           '{"headline": "string", "subheadline": "string", "cta_text": "string", "cta_url": "string"}'::jsonb,
           'hero',
           'section',
           'general',
           '["hero", "header", "landing"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'contact-hero',
           'Contact Page Hero',
           'Hero section for contact pages',
           '<section class="hero hero-contact" data-component="contact-hero">
               <div class="hero-content">
                   <h1>{{.headline}}</h1>
                   <p class="hero-subheadline">{{.subheadline}}</p>
               </div>
           </section>',
           '{"headline": "string", "subheadline": "string"}'::jsonb,
           'hero',
           'section',
           'contact',
           '["hero", "contact"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'services-hero',
           'Services Page Hero',
           'Hero section for services pages',
           '<section class="hero hero-services" data-component="services-hero">
               <div class="hero-content">
                   <h1>{{.headline}}</h1>
                   <p class="hero-subheadline">{{.subheadline}}</p>
               </div>
           </section>',
           '{"headline": "string", "subheadline": "string"}'::jsonb,
           'hero',
           'section',
           'services',
           '["hero", "services"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'about-hero',
           'About Page Hero',
           'Hero section for about pages',
           '<section class="hero hero-about" data-component="about-hero">
               <div class="hero-content">
                   <h1>{{.headline}}</h1>
                   <p class="hero-subheadline">{{.subheadline}}</p>
               </div>
           </section>',
           '{"headline": "string", "subheadline": "string"}'::jsonb,
           'hero',
           'section',
           'about',
           '["hero", "about"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'case-studies-hero',
           'Case Studies Hero',
           'Hero section for case studies/use cases pages',
           '<section class="hero hero-case-studies" data-component="case-studies-hero">
               <div class="hero-content">
                   <h1>{{.headline}}</h1>
                   <p class="hero-subheadline">{{.subheadline}}</p>
               </div>
           </section>',
           '{"headline": "string", "subheadline": "string"}'::jsonb,
           'hero',
           'section',
           'case-studies',
           '["hero", "case-studies", "portfolio"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              is_active = true;

-- ============================================================================
-- Contact page components
-- ============================================================================

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'contact-form',
           'Contact Form',
           'Standard contact form with name, email, message fields',
           '<section class="contact-form-section" data-component="contact-form">
               <div class="form-container">
                   {{if .form_title}}<h2>{{.form_title}}</h2>{{end}}
                   {{if .form_description}}<p>{{.form_description}}</p>{{end}}
                   <form class="contact-form" action="{{.form_action}}" method="POST">
                       <div class="form-group">
                           <label for="name">Name</label>
                           <input type="text" id="name" name="name" required placeholder="Your name">
                       </div>
                       <div class="form-group">
                           <label for="email">Email</label>
                           <input type="email" id="email" name="email" required placeholder="your@email.com">
                       </div>
                       <div class="form-group">
                           <label for="message">Message</label>
                           <textarea id="message" name="message" rows="5" required placeholder="How can we help?"></textarea>
                       </div>
                       <button type="submit" class="btn btn-primary">{{.submit_text}}</button>
                   </form>
               </div>
           </section>',
           '{"form_title": "string", "form_description": "string", "form_action": "string", "submit_text": "string"}'::jsonb,
           'contact-form',
           'section',
           'contact',
           '["contact", "form", "lead-capture"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'contact-info',
           'Contact Information',
           'Contact details including address, phone, email',
           '<section class="contact-info-section" data-component="contact-info">
               <div class="contact-details">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   {{if .intro_text}}<p>{{.intro_text}}</p>{{end}}
                   <div class="contact-items">
                       {{if .email}}<div class="contact-item"><strong>Email:</strong> <a href="mailto:{{.email}}">{{.email}}</a></div>{{end}}
                       {{if .phone}}<div class="contact-item"><strong>Phone:</strong> <a href="tel:{{.phone}}">{{.phone}}</a></div>{{end}}
                       {{if .address}}<div class="contact-item"><strong>Address:</strong> {{.address}}</div>{{end}}
                       {{if .hours}}<div class="contact-item"><strong>Hours:</strong> {{.hours}}</div>{{end}}
                   </div>
               </div>
           </section>',
           '{"section_title": "string", "intro_text": "string", "email": "string", "phone": "string", "address": "string", "hours": "string"}'::jsonb,
           'contact-info',
           'section',
           'contact',
           '["contact", "info", "details"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

-- ============================================================================
-- Features and services components
-- ============================================================================

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'features',
           'Features Grid',
           'Grid of feature items with icons and descriptions',
           '<section class="features-section" data-component="features">
               <div class="features-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   {{if .section_intro}}<p class="section-intro">{{.section_intro}}</p>{{end}}
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
           </section>',
           '{"section_title": "string", "section_intro": "string", "features": [{"icon": "string", "title": "string", "description": "string"}]}'::jsonb,
           'features',
           'section',
           'general',
           '["features", "benefits", "grid"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'services-grid',
           'Services Grid',
           'Grid of services offered',
           '<section class="services-grid-section" data-component="services-grid">
               <div class="services-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   <div class="services-grid">
                       {{range .services}}
                       <div class="service-item">
                           <h3>{{.title}}</h3>
                           <p>{{.description}}</p>
                           {{if .link}}<a href="{{.link}}" class="service-link">Learn more</a>{{end}}
                       </div>
                       {{end}}
                   </div>
               </div>
           </section>',
           '{"section_title": "string", "services": [{"title": "string", "description": "string", "link": "string"}]}'::jsonb,
           'services-grid',
           'section',
           'services',
           '["services", "offerings", "grid"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

-- ============================================================================
-- Social proof and testimonials
-- ============================================================================

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'social_proof',
           'Social Proof Section',
           'Logos, stats, or trust indicators',
           '<section class="social-proof-section" data-component="social-proof">
               <div class="social-proof-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   {{if .stats}}
                   <div class="stats-grid">
                       {{range .stats}}
                       <div class="stat-item">
                           <span class="stat-number">{{.number}}</span>
                           <span class="stat-label">{{.label}}</span>
                       </div>
                       {{end}}
                   </div>
                   {{end}}
                   {{if .logos}}
                   <div class="logo-strip">
                       {{range .logos}}
                       <img src="{{.src}}" alt="{{.alt}}" class="client-logo">
                       {{end}}
                   </div>
                   {{end}}
               </div>
           </section>',
           '{"section_title": "string", "stats": [{"number": "string", "label": "string"}], "logos": [{"src": "string", "alt": "string"}]}'::jsonb,
           'social-proof',
           'section',
           'general',
           '["social-proof", "trust", "stats", "logos"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'testimonials',
           'Testimonials Section',
           'Customer testimonials and quotes',
           '<section class="testimonials-section" data-component="testimonials">
               <div class="testimonials-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   <div class="testimonials-grid">
                       {{range .testimonials}}
                       <div class="testimonial-item">
                           <blockquote>{{.quote}}</blockquote>
                           <cite>
                               <strong>{{.name}}</strong>
                               {{if .title}}<span>{{.title}}</span>{{end}}
                               {{if .company}}<span>{{.company}}</span>{{end}}
                           </cite>
                       </div>
                       {{end}}
                   </div>
               </div>
           </section>',
           '{"section_title": "string", "testimonials": [{"quote": "string", "name": "string", "title": "string", "company": "string"}]}'::jsonb,
           'testimonials',
           'section',
           'general',
           '["testimonials", "quotes", "social-proof"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

-- ============================================================================
-- Call to action
-- ============================================================================

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'call_to_action',
           'Call to Action',
           'CTA section with headline and button',
           '<section class="cta-section" data-component="call-to-action">
               <div class="cta-container">
                   <h2>{{.headline}}</h2>
                   {{if .subheadline}}<p>{{.subheadline}}</p>{{end}}
                   <a href="{{.cta_url}}" class="btn btn-primary btn-large">{{.cta_text}}</a>
               </div>
           </section>',
           '{"headline": "string", "subheadline": "string", "cta_text": "string", "cta_url": "string"}'::jsonb,
           'call-to-action',
           'section',
           'general',
           '["cta", "conversion", "action"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

-- ============================================================================
-- About page components
-- ============================================================================

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'about-content',
           'About Content',
           'Main about page content section',
           '<section class="about-content-section" data-component="about-content">
               <div class="about-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   <div class="about-text">
                       {{.content}}
                   </div>
                   {{if .highlights}}
                   <div class="about-highlights">
                       {{range .highlights}}
                       <div class="highlight-item">
                           <h3>{{.title}}</h3>
                           <p>{{.description}}</p>
                       </div>
                       {{end}}
                   </div>
                   {{end}}
               </div>
           </section>',
           '{"section_title": "string", "content": "string", "highlights": [{"title": "string", "description": "string"}]}'::jsonb,
           'about-content',
           'section',
           'about',
           '["about", "company", "story"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'leadership-team',
           'Leadership Team',
           'Team members grid',
           '<section class="team-section" data-component="leadership-team">
               <div class="team-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   {{if .section_intro}}<p class="section-intro">{{.section_intro}}</p>{{end}}
                   <div class="team-grid">
                       {{range .members}}
                       <div class="team-member">
                           {{if .photo}}<img src="{{.photo}}" alt="{{.name}}" class="member-photo">{{end}}
                           <h3>{{.name}}</h3>
                           <p class="member-title">{{.title}}</p>
                           {{if .bio}}<p class="member-bio">{{.bio}}</p>{{end}}
                       </div>
                       {{end}}
                   </div>
               </div>
           </section>',
           '{"section_title": "string", "section_intro": "string", "members": [{"name": "string", "title": "string", "bio": "string", "photo": "string"}]}'::jsonb,
           'leadership-team',
           'section',
           'about',
           '["team", "leadership", "people"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'differentiators-section',
           'Differentiators',
           'What makes us different section',
           '<section class="differentiators-section" data-component="differentiators">
               <div class="differentiators-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   <div class="differentiators-grid">
                       {{range .differentiators}}
                       <div class="differentiator-item">
                           <h3>{{.title}}</h3>
                           <p>{{.description}}</p>
                       </div>
                       {{end}}
                   </div>
               </div>
           </section>',
           '{"section_title": "string", "differentiators": [{"title": "string", "description": "string"}]}'::jsonb,
           'differentiators',
           'section',
           'about',
           '["differentiators", "unique", "why-us"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

-- ============================================================================
-- Case studies / Portfolio
-- ============================================================================

INSERT INTO content_components (name, display_name, description, html_template, input_schema, "function", component_level, category, semantic_tags, is_active)
VALUES (
           'case-studies-list',
           'Case Studies List',
           'Grid or list of case studies',
           '<section class="case-studies-section" data-component="case-studies-list">
               <div class="case-studies-container">
                   {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
                   <div class="case-studies-grid">
                       {{range .case_studies}}
                       <div class="case-study-item">
                           {{if .image}}<img src="{{.image}}" alt="{{.title}}" class="case-study-image">{{end}}
                           <h3>{{.title}}</h3>
                           <p class="case-study-client">{{.client}}</p>
                           <p>{{.summary}}</p>
                           {{if .link}}<a href="{{.link}}" class="case-study-link">Read more</a>{{end}}
                       </div>
                       {{end}}
                   </div>
               </div>
           </section>',
           '{"section_title": "string", "case_studies": [{"title": "string", "client": "string", "summary": "string", "image": "string", "link": "string"}]}'::jsonb,
           'case-studies-list',
           'section',
           'case-studies',
           '["case-studies", "portfolio", "work"]'::jsonb,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              is_active = true;

-- ============================================================================
-- Create unique constraint on name if it doesn't exist
-- ============================================================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'content_components_name_key'
    ) THEN
ALTER TABLE content_components ADD CONSTRAINT content_components_name_key UNIQUE (name);
END IF;
EXCEPTION
    WHEN duplicate_table THEN NULL;
END $$;

-- Verify components were added
SELECT name, display_name, category, "function"
FROM content_components
WHERE name IN ('hero', 'contact-hero', 'contact-form', 'contact-info', 'features', 'services-grid', 'social_proof', 'call_to_action', 'about-content', 'leadership-team', 'differentiators-section', 'case-studies-list', 'services-hero', 'about-hero', 'case-studies-hero', 'testimonials')
ORDER BY category, name;