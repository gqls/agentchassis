-- Insert generic fallback component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Generic Text Block',
'Fallback component for any unmatched section',
'generic-text-block',
'<section id="{{.ComponentID}}" class="section section--generic">
  <div class="container">
    <h2 class="section__title">{{.heading}}</h2>
    <div class="section__content">{{.content}}</div>
  </div>
</section>',
    '{
      "heading": "Section Heading",
      "content": "Your content goes here. This is a generic text block that can be used for any section."
    }'::jsonb,
    NOW(),
    NOW()
);

-- HEAD component with base CSS
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Document Head',
'HTML head with meta tags, base CSS and optional theme',
'head',
'<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{{.description}}">
  <title>{{.title}}</title>
  <style>
    /* CSS Reset & Base Styles */
    * { margin: 0; padding: 0; box-sizing: border-box; }

    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      line-height: 1.6;
      color: var(--color-text);
      background-color: var(--color-background);
    }
    
    /* Layout */
    .container { 
      max-width: 1200px; 
      margin: 0 auto; 
      padding: 0 1rem; 
    }
    .container--narrow { max-width: 800px; }
    
    /* Grid System */
    .grid { 
      display: grid; 
      gap: 2rem; 
    }
    .grid--2 { grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); }
    .grid--3 { grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); }
    .grid--4 { grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); }
    
    /* Sections */
    .section { padding: 4rem 1rem; }
    .section__title { 
      font-size: 2.5rem; 
      margin-bottom: 1rem;
      color: var(--color-heading);
    }
    .section__title--center { text-align: center; }
    .section__content { font-size: 1.125rem; }
    
    /* Cards */
    .card { 
      padding: 2rem; 
      background: var(--color-card-bg);
      border-radius: var(--border-radius);
      box-shadow: var(--shadow);
    }
    
    /* Buttons */
    .button {
      display: inline-block;
      padding: 0.75rem 1.5rem;
      border: none;
      border-radius: var(--border-radius);
      font-size: 1rem;
      font-weight: 600;
      text-decoration: none;
      cursor: pointer;
      transition: all 0.2s;
    }
    .button--primary {
      background: var(--color-primary);
      color: var(--color-primary-text);
    }
    .button--primary:hover {
      background: var(--color-primary-hover);
    }
    .button--secondary {
      background: var(--color-secondary);
      color: var(--color-secondary-text);
    }
    .button--secondary:hover {
      background: var(--color-secondary-hover);
    }
    .button--primary-inverse {
      background: var(--color-background);
      color: var(--color-primary);
    }
    .button--secondary-inverse {
      background: transparent;
      color: var(--color-background);
      border: 2px solid var(--color-background);
    }
    .button--large { 
      padding: 1rem 2rem; 
      font-size: 1.125rem; 
    }
    .button--small { 
      padding: 0.5rem 1rem; 
      font-size: 0.875rem; 
    }
    .button--full-width { 
      display: block; 
      width: 100%; 
      text-align: center; 
    }
    
    /* Header */
    .site-header {
      background: var(--color-header-bg);
      color: var(--color-header-text);
      position: sticky;
      top: 0;
      z-index: 1000;
      box-shadow: var(--shadow);
    }
    .site-header__nav {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1rem;
    }
    .site-header__brand {
      font-size: 1.5rem;
      font-weight: bold;
    }
    .site-header__menu {
      display: flex;
      gap: 2rem;
      list-style: none;
    }
    .site-header__link {
      color: var(--color-header-text);
      text-decoration: none;
    }
    .site-header__link:hover {
      opacity: 0.8;
    }
    
    /* Hero */
    .hero {
      text-align: center;
      padding: 4rem 1rem;
    }
    .hero__title {
      font-size: 3rem;
      margin-bottom: 1.5rem;
      color: var(--color-hero-title);
    }
    .hero__subtitle {
      font-size: 1.5rem;
      margin-bottom: 2rem;
      color: var(--color-hero-subtitle);
    }
    .hero__actions {
      display: flex;
      gap: 1rem;
      justify-content: center;
      flex-wrap: wrap;
    }
    
    /* Features */
    .feature {
      text-align: center;
    }
    .feature__icon {
      font-size: 3rem;
      margin-bottom: 1rem;
    }
    .feature__title {
      font-size: 1.25rem;
      margin-bottom: 0.5rem;
      color: var(--color-heading);
    }
    .feature__description {
      color: var(--color-text-muted);
    }
    
    /* Pricing */
    .pricing-tier {
      position: relative;
    }
    .pricing-tier--featured {
      border: 3px solid var(--color-primary);
      transform: scale(1.05);
    }
    .pricing-tier__badge {
      position: absolute;
      top: -1rem;
      right: 1rem;
      background: var(--color-primary);
      color: var(--color-primary-text);
      padding: 0.25rem 1rem;
      border-radius: var(--border-radius);
      font-size: 0.875rem;
      font-weight: bold;
    }
    .pricing-tier__name {
      font-size: 1.5rem;
      margin-bottom: 1rem;
    }
    .pricing-tier__price {
      font-size: 2.5rem;
      font-weight: bold;
      margin-bottom: 1.5rem;
      color: var(--color-primary);
    }
    .pricing-tier__features {
      list-style: none;
      margin-bottom: 2rem;
    }
    .pricing-tier__feature {
      padding: 0.5rem 0;
      border-bottom: 1px solid var(--color-border);
    }
    
    /* Testimonials */
    .stat-highlight {
      text-align: center;
      margin-bottom: 3rem;
    }
    .stat-highlight__number {
      font-size: 3rem;
      font-weight: bold;
      color: var(--color-primary);
    }
    .stat-highlight__label {
      font-size: 1.25rem;
      color: var(--color-text-muted);
    }
    .testimonial {
      text-align: center;
    }
    .testimonial__rating {
      color: var(--color-accent);
      font-size: 1.25rem;
      margin-bottom: 1rem;
    }
    .testimonial__text {
      font-style: italic;
      margin-bottom: 1rem;
    }
    .testimonial__author {
      font-weight: bold;
      color: var(--color-text-muted);
    }
    
    /* FAQ */
    .faq-item {
      padding: 1.5rem;
      background: var(--color-card-bg);
      border-radius: var(--border-radius);
      margin-bottom: 1rem;
      cursor: pointer;
    }
    .faq-item__question {
      font-size: 1.125rem;
      font-weight: 600;
      cursor: pointer;
    }
    .faq-item__answer {
      margin-top: 1rem;
      color: var(--color-text-muted);
    }
    
    /* CTA Section */
    .section--cta {
      background: var(--color-cta-bg);
      color: var(--color-cta-text);
    }
    .cta {
      text-align: center;
    }
    .cta__title {
      font-size: 2.5rem;
      margin-bottom: 1rem;
      color: var(--color-cta-text);
    }
    .cta__subtitle {
      font-size: 1.25rem;
      margin-bottom: 2rem;
    }
    .cta__actions {
      display: flex;
      gap: 1rem;
      justify-content: center;
      flex-wrap: wrap;
    }
    
    /* Footer */
    .site-footer {
      background: var(--color-footer-bg);
      color: var(--color-footer-text);
      padding: 3rem 1rem 1rem;
    }
    .site-footer__brand {
      font-size: 1.5rem;
      font-weight: bold;
      margin-bottom: 0.5rem;
    }
    .site-footer__tagline {
      opacity: 0.8;
    }
    .site-footer__heading {
      font-size: 1.125rem;
      margin-bottom: 1rem;
    }
    .site-footer__links {
      list-style: none;
    }
    .site-footer__links li {
      margin-bottom: 0.5rem;
    }
    .site-footer__link {
      color: var(--color-footer-text);
      text-decoration: none;
      opacity: 0.8;
    }
    .site-footer__link:hover {
      opacity: 1;
    }
    .site-footer__bottom {
      margin-top: 3rem;
      padding-top: 2rem;
      border-top: 1px solid var(--color-border);
      text-align: center;
      opacity: 0.6;
    }
    
    /* Responsive */
    @media (max-width: 768px) {
      .site-header__menu { display: none; }
      .hero__title { font-size: 2rem; }
      .hero__subtitle { font-size: 1.25rem; }
      .section__title { font-size: 2rem; }
      .grid--4 { grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); }
    }
  </style>
  <link rel="stylesheet" href="/themes/{{.theme}}.css">
</head>',
    '{
      "title": "Your Site Title",
      "description": "Your site description for SEO",
      "theme": "default"
    }'::jsonb,
    NOW(),
    NOW()
);

-- Header component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Site Header',
'Main navigation header',
'header',
'<header id="{{.ComponentID}}" class="site-header">
  <nav class="site-header__nav container">
    <div class="site-header__brand">{{.brand_name}}</div>
    <ul class="site-header__menu">
      <li class="site-header__menu-item"><a href="#hero" class="site-header__link">Home</a></li>
      <li class="site-header__menu-item"><a href="#features" class="site-header__link">Features</a></li>
      <li class="site-header__menu-item"><a href="#pricing" class="site-header__link">Pricing</a></li>
      <li class="site-header__menu-item"><a href="#faq" class="site-header__link">FAQ</a></li>
    </ul>
    <a href="#call_to_action" class="button button--primary button--small">{{.cta_text}}</a>
  </nav>
</header>',
    '{
      "brand_name": "Your Brand",
      "cta_text": "Get Started"
    }'::jsonb,
    NOW(),
    NOW()
);

-- Hero component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Hero Section',
'Main hero banner with headline and CTAs',
'hero',
'<section id="{{.ComponentID}}" class="section section--hero">
  <div class="container hero">
    <h1 class="hero__title">{{.headline}}</h1>
    <p class="hero__subtitle">{{.subheadline}}</p>
    <div class="hero__actions">
      <a href="#pricing" class="button button--primary button--large">{{.primary_cta}}</a>
      <a href="#features" class="button button--secondary button--large">{{.secondary_cta}}</a>
    </div>
  </div>
</section>',
    '{
      "headline": "Transform Your Business Today",
      "subheadline": "The ultimate solution for modern teams",
      "primary_cta": "Get Started Now",
      "secondary_cta": "Learn More"
    }'::jsonb,
    NOW(),
    NOW()
);

-- Social Proof component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Social Proof',
'Testimonials and statistics',
'social_proof',
'<section id="{{.ComponentID}}" class="section section--social-proof">
  <div class="container">
    <div class="stat-highlight">
      <div class="stat-highlight__number">{{.stat_number}}</div>
      <div class="stat-highlight__label">{{.stat_label}}</div>
    </div>
    <div class="testimonials grid grid--3">
      <div class="testimonial card">
        <div class="testimonial__rating">★★★★★</div>
        <p class="testimonial__text">"{{.testimonial_1}}"</p>
        <p class="testimonial__author">{{.author_1}}</p>
      </div>
      <div class="testimonial card">
        <div class="testimonial__rating">★★★★★</div>
        <p class="testimonial__text">"{{.testimonial_2}}"</p>
        <p class="testimonial__author">{{.author_2}}</p>
      </div>
      <div class="testimonial card">
        <div class="testimonial__rating">★★★★★</div>
        <p class="testimonial__text">"{{.testimonial_3}}"</p>
        <p class="testimonial__author">{{.author_3}}</p>
      </div>
    </div>
  </div>
</section>',
    '{
      "stat_number": "50,000+",
      "stat_label": "Happy Customers",
      "testimonial_1": "This product changed everything for our team.",
      "author_1": "Jane Smith, CEO",
      "testimonial_2": "Best decision we made this year.",
      "author_2": "John Doe, Founder",
      "testimonial_3": "Incredible results in just 30 days.",
      "author_3": "Sarah Johnson, Manager"
    }'::jsonb,
    NOW(),
    NOW()
);

-- Features component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Features Grid',
'Grid of product/service features',
'features',
'<section id="{{.ComponentID}}" class="section section--features">
  <div class="container">
    <h2 class="section__title section__title--center">{{.section_title}}</h2>
    <div class="features grid grid--4">
      <div class="feature">
        <div class="feature__icon">{{.icon_1}}</div>
        <h3 class="feature__title">{{.feature_1_title}}</h3>
        <p class="feature__description">{{.feature_1_description}}</p>
      </div>
      <div class="feature">
        <div class="feature__icon">{{.icon_2}}</div>
        <h3 class="feature__title">{{.feature_2_title}}</h3>
        <p class="feature__description">{{.feature_2_description}}</p>
      </div>
      <div class="feature">
        <div class="feature__icon">{{.icon_3}}</div>
        <h3 class="feature__title">{{.feature_3_title}}</h3>
        <p class="feature__description">{{.feature_3_description}}</p>
      </div>
      <div class="feature">
        <div class="feature__icon">{{.icon_4}}</div>
        <h3 class="feature__title">{{.feature_4_title}}</h3>
        <p class="feature__description">{{.feature_4_description}}</p>
      </div>
    </div>
  </div>
</section>',
    '{
      "section_title": "Why Choose Us",
      "icon_1": "✓",
      "feature_1_title": "Fast & Reliable",
      "feature_1_description": "Lightning-fast performance you can count on",
      "icon_2": "🔒",
      "feature_2_title": "Secure",
      "feature_2_description": "Bank-level security for your peace of mind",
      "icon_3": "📱",
      "feature_3_title": "Mobile Ready",
      "feature_3_description": "Works perfectly on any device",
      "icon_4": "💪",
      "feature_4_title": "24/7 Support",
      "feature_4_description": "Always here when you need us"
    }'::jsonb,
    NOW(),
    NOW()
);

-- Pricing component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Pricing Tiers',
'Pricing comparison table',
'pricing',
'<section id="{{.ComponentID}}" class="section section--pricing">
  <div class="container">
    <h2 class="section__title section__title--center">{{.section_title}}</h2>
    <div class="pricing-tiers grid grid--3">
      <div class="pricing-tier card">
        <h3 class="pricing-tier__name">{{.tier_1_name}}</h3>
        <div class="pricing-tier__price">{{.tier_1_price}}</div>
        <ul class="pricing-tier__features">
          <li class="pricing-tier__feature">{{.tier_1_feature_1}}</li>
          <li class="pricing-tier__feature">{{.tier_1_feature_2}}</li>
          <li class="pricing-tier__feature">{{.tier_1_feature_3}}</li>
        </ul>
        <a href="#" class="button button--secondary button--full-width">{{.tier_1_cta}}</a>
      </div>
      <div class="pricing-tier pricing-tier--featured card">
        <div class="pricing-tier__badge">Popular</div>
        <h3 class="pricing-tier__name">{{.tier_2_name}}</h3>
        <div class="pricing-tier__price">{{.tier_2_price}}</div>
        <ul class="pricing-tier__features">
          <li class="pricing-tier__feature">{{.tier_2_feature_1}}</li>
          <li class="pricing-tier__feature">{{.tier_2_feature_2}}</li>
          <li class="pricing-tier__feature">{{.tier_2_feature_3}}</li>
          <li class="pricing-tier__feature">{{.tier_2_feature_4}}</li>
        </ul>
        <a href="#" class="button button--primary button--full-width">{{.tier_2_cta}}</a>
      </div>
      <div class="pricing-tier card">
        <h3 class="pricing-tier__name">{{.tier_3_name}}</h3>
        <div class="pricing-tier__price">{{.tier_3_price}}</div>
        <ul class="pricing-tier__features">
          <li class="pricing-tier__feature">{{.tier_3_feature_1}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_2}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_3}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_4}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_5}}</li>
        </ul>
        <a href="#" class="button button--secondary button--full-width">{{.tier_3_cta}}</a>
      </div>
    </div>
  </div>
</section>',
    '{
      "section_title": "Choose Your Plan",
      "tier_1_name": "Basic",
      "tier_1_price": "$29/mo",
      "tier_1_feature_1": "Basic features",
      "tier_1_feature_2": "Up to 10 users",
      "tier_1_feature_3": "Email support",
      "tier_1_cta": "Get Started",
      "tier_2_name": "Pro",
      "tier_2_price": "$79/mo",
      "tier_2_feature_1": "All Basic features",
      "tier_2_feature_2": "Up to 50 users",
      "tier_2_feature_3": "Priority support",
      "tier_2_feature_4": "Advanced analytics",
      "tier_2_cta": "Start Free Trial",
      "tier_3_name": "Enterprise",
      "tier_3_price": "$199/mo",
      "tier_3_feature_1": "All Pro features",
      "tier_3_feature_2": "Unlimited users",
      "tier_3_feature_3": "24/7 phone support",
      "tier_3_feature_4": "Custom integrations",
      "tier_3_feature_5": "Dedicated account manager",
      "tier_3_cta": "Contact Sales"
    }'::jsonb,
    NOW(),
    NOW()
);

-- FAQ component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'FAQ Section',
'Frequently asked questions accordion',
'faq',
'<section id="{{.ComponentID}}" class="section section--faq">
  <div class="container container--narrow">
    <h2 class="section__title section__title--center">{{.section_title}}</h2>
    <div class="faq-list">
      <details class="faq-item">
        <summary class="faq-item__question">{{.question_1}}</summary>
        <p class="faq-item__answer">{{.answer_1}}</p>
      </details>
      <details class="faq-item">
        <summary class="faq-item__question">{{.question_2}}</summary>
        <p class="faq-item__answer">{{.answer_2}}</p>
      </details>
      <details class="faq-item">
        <summary class="faq-item__question">{{.question_3}}</summary>
        <p class="faq-item__answer">{{.answer_3}}</p>
      </details>
      <details class="faq-item">
        <summary class="faq-item__question">{{.question_4}}</summary>
        <p class="faq-item__answer">{{.answer_4}}</p>
      </details>
    </div>
  </div>
</section>',
    '{
      "section_title": "Frequently Asked Questions",
      "question_1": "How does it work?",
      "answer_1": "Our platform is designed to be intuitive and easy to use. Simply sign up, configure your settings, and start seeing results.",
      "question_2": "What payment methods do you accept?",
      "answer_2": "We accept all major credit cards, PayPal, and bank transfers for annual plans.",
      "question_3": "Can I cancel anytime?",
      "answer_3": "Yes! You can cancel your subscription at any time with no penalties or fees.",
      "question_4": "Do you offer customer support?",
      "answer_4": "Absolutely! We offer 24/7 support via email, chat, and phone for all our customers."
    }'::jsonb,
    NOW(),
    NOW()
);

-- Call to Action component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Call to Action',
'Final conversion section',
'call_to_action',
'<section id="{{.ComponentID}}" class="section section--cta">
  <div class="container container--narrow cta">
    <h2 class="cta__title">{{.headline}}</h2>
    <p class="cta__subtitle">{{.subheadline}}</p>
    <div class="cta__actions">
      <a href="#" class="button button--primary-inverse button--large">{{.primary_button}}</a>
      <a href="#" class="button button--secondary-inverse button--large">{{.secondary_button}}</a>
    </div>
  </div>
</section>',
    '{
      "headline": "Ready to Get Started?",
      "subheadline": "Join thousands of satisfied customers today",
      "primary_button": "Start Free Trial",
      "secondary_button": "Schedule Demo"
    }'::jsonb,
    NOW(),
    NOW()
);

-- Footer component
INSERT INTO content_components (id, name, description, "function", html_template, input_schema, created_at, updated_at)
VALUES (
gen_random_uuid(),
'Site Footer',
'Footer with links and copyright',
'footer',
'<footer id="{{.ComponentID}}" class="site-footer">
  <div class="container">
    <div class="site-footer__content grid grid--4">
      <div class="site-footer__col">
        <h3 class="site-footer__brand">{{.brand_name}}</h3>
        <p class="site-footer__tagline">{{.tagline}}</p>
      </div>
      <div class="site-footer__col">
        <h4 class="site-footer__heading">Product</h4>
        <ul class="site-footer__links">
          <li><a href="#features" class="site-footer__link">Features</a></li>
          <li><a href="#pricing" class="site-footer__link">Pricing</a></li>
          <li><a href="#" class="site-footer__link">Updates</a></li>
        </ul>
      </div>
      <div class="site-footer__col">
        <h4 class="site-footer__heading">Company</h4>
        <ul class="site-footer__links">
          <li><a href="#" class="site-footer__link">About</a></li>
          <li><a href="#" class="site-footer__link">Blog</a></li>
          <li><a href="#" class="site-footer__link">Careers</a></li>
        </ul>
      </div>
      <div class="site-footer__col">
        <h4 class="site-footer__heading">Legal</h4>
        <ul class="site-footer__links">
          <li><a href="#" class="site-footer__link">Privacy</a></li>
          <li><a href="#" class="site-footer__link">Terms</a></li>
          <li><a href="#" class="site-footer__link">Contact</a></li>
        </ul>
      </div>
    </div>
    <div class="site-footer__bottom">
      <p class="site-footer__copyright">{{.copyright}}</p>
    </div>
  </div>
</footer>',
    '{
      "brand_name": "Your Brand",
      "tagline": "Making the world a better place",
      "copyright": "© 2024 Your Company. All rights reserved."
    }'::jsonb,
    NOW(),
    NOW()
);


====
====

Now the architect can:

Determine which theme to use (from domain/objective analysis)
Query css_themes table: SELECT css_content FROM css_themes WHERE name = 'boxing'
Inject the CSS into the HEAD component's {{.theme_css}} placeholder
Output complete, self-contained HTML

This way all CSS is versioned, deployable, and can be improved over time!

-- Create CSS themes table
CREATE TABLE IF NOT EXISTS css_themes (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
name TEXT NOT NULL UNIQUE,
display_name TEXT NOT NULL,
description TEXT,
category TEXT, -- 'professional', 'aggressive', 'warm', 'modern', etc.
css_content TEXT NOT NULL,
version INTEGER DEFAULT 1,
is_active BOOLEAN DEFAULT true,
created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_css_themes_name ON css_themes(name);
CREATE INDEX idx_css_themes_category ON css_themes(category);

-- Insert default theme
INSERT INTO css_themes (name, display_name, description, category, css_content)
VALUES (
'default',
'Default Professional',
'Neutral professional theme suitable for most business sites',
'professional',
':root {
--color-primary: #3b82f6;
--color-primary-hover: #2563eb;
--color-primary-text: #ffffff;
--color-secondary: #64748b;
--color-secondary-hover: #475569;
--color-secondary-text: #ffffff;
--color-accent: #fbbf24;

--color-text: #1e293b;
--color-text-muted: #64748b;
--color-heading: #0f172a;
--color-background: #ffffff;
--color-border: #e2e8f0;

--color-header-bg: #1e293b;
--color-header-text: #ffffff;
--color-hero-title: #0f172a;
--color-hero-subtitle: #475569;
--color-card-bg: #ffffff;
--color-cta-bg: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
--color-cta-text: #ffffff;
--color-footer-bg: #1e293b;
--color-footer-text: #ffffff;

--border-radius: 0.5rem;
--shadow: 0 1px 3px rgba(0,0,0,0.1);
}'
);

-- Insert boxing theme
INSERT INTO css_themes (name, display_name, description, category, css_content)
VALUES (
'boxing',
'Boxing Energy',
'Aggressive, high-energy theme for sports and competitive industries',
'aggressive',
':root {
--color-primary: #dc2626;
--color-primary-hover: #b91c1c;
--color-primary-text: #ffffff;
--color-secondary: #fbbf24;
--color-secondary-hover: #f59e0b;
--color-secondary-text: #000000;
--color-accent: #fbbf24;

--color-text: #1c1917;
--color-text-muted: #57534e;
--color-heading: #000000;
--color-background: #fafaf9;
--color-border: #d6d3d1;

--color-header-bg: #000000;
--color-header-text: #ffffff;
--color-hero-title: #dc2626;
--color-hero-subtitle: #57534e;
--color-card-bg: #ffffff;
--color-cta-bg: linear-gradient(135deg, #dc2626 0%, #991b1b 100%);
--color-cta-text: #ffffff;
--color-footer-bg: #1c1917;
--color-footer-text: #ffffff;

--border-radius: 0.25rem;
--shadow: 0 4px 6px rgba(0,0,0,0.2);
}

body {
font-family: "Impact", "Arial Black", sans-serif;
letter-spacing: 0.5px;
}

.hero__title {
text-transform: uppercase;
letter-spacing: 2px;
}'
);

-- Insert bakery theme
INSERT INTO css_themes (name, display_name, description, category, css_content)
VALUES (
'bakery',
'Warm Bakery',
'Warm, inviting theme for food, hospitality, and lifestyle brands',
'warm',
':root {
--color-primary: #ea580c;
--color-primary-hover: #c2410c;
--color-primary-text: #ffffff;
--color-secondary: #78350f;
--color-secondary-hover: #451a03;
--color-secondary-text: #ffffff;
--color-accent: #fbbf24;

--color-text: #44403c;
--color-text-muted: #78716c;
--color-heading: #292524;
--color-background: #fef3c7;
--color-border: #fde68a;

--color-header-bg: #78350f;
--color-header-text: #fef3c7;
--color-hero-title: #78350f;
--color-hero-subtitle: #78716c;
--color-card-bg: #fffbeb;
--color-cta-bg: linear-gradient(135deg, #ea580c 0%, #c2410c 100%);
--color-cta-text: #ffffff;
--color-footer-bg: #44403c;
--color-footer-text: #fef3c7;

--border-radius: 1rem;
--shadow: 0 4px 8px rgba(120, 53, 15, 0.15);
}

body {
font-family: "Georgia", "Times New Roman", serif;
}

.section__title {
font-family: "Comic Sans MS", cursive;
}'
);

-- Insert tech theme
INSERT INTO css_themes (name, display_name, description, category, css_content)
VALUES (
'tech',
'Modern Tech',
'Dark, sleek theme for technology and SaaS companies',
'modern',
':root {
--color-primary: #8b5cf6;
--color-primary-hover: #7c3aed;
--color-primary-text: #ffffff;
--color-secondary: #06b6d4;
--color-secondary-hover: #0891b2;
--color-secondary-text: #ffffff;
--color-accent: #ec4899;

--color-text: #e2e8f0;
--color-text-muted: #94a3b8;
--color-heading: #f1f5f9;
--color-background: #0f172a;
--color-border: #1e293b;

--color-header-bg: #020617;
--color-header-text: #f1f5f9;
--color-hero-title: #f1f5f9;
--color-hero-subtitle: #cbd5e1;
--color-card-bg: #1e293b;
--color-cta-bg: linear-gradient(135deg, #8b5cf6 0%, #ec4899 100%);
--color-cta-text: #ffffff;
--color-footer-bg: #020617;
--color-footer-text: #cbd5e1;

--border-radius: 0.75rem;
--shadow: 0 0 20px rgba(139, 92, 246, 0.3);
}

body {
font-family: "Inter", -apple-system, sans-serif;
}'
);

-- Insert law/finance theme
INSERT INTO css_themes (name, display_name, description, category, css_content)
VALUES (
'professional-dark',
'Professional Dark',
'Conservative, trustworthy theme for law, finance, and professional services',
'professional',
':root {
--color-primary: #1e40af;
--color-primary-hover: #1e3a8a;
--color-primary-text: #ffffff;
--color-secondary: #475569;
--color-secondary-hover: #334155;
--color-secondary-text: #ffffff;
--color-accent: #d97706;

--color-text: #1e293b;
--color-text-muted: #64748b;
--color-heading: #0f172a;
--color-background: #f8fafc;
--color-border: #cbd5e1;

--color-header-bg: #0f172a;
--color-header-text: #f1f5f9;
--color-hero-title: #1e40af;
--color-hero-subtitle: #475569;
--color-card-bg: #ffffff;
--color-cta-bg: linear-gradient(135deg, #1e40af 0%, #1e3a8a 100%);
--color-cta-text: #ffffff;
--color-footer-bg: #0f172a;
--color-footer-text: #cbd5e1;

--border-radius: 0.25rem;
--shadow: 0 2px 4px rgba(0,0,0,0.1);
}

body {
font-family: "Merriweather", "Georgia", serif;
}

.section__title {
font-weight: 700;
letter-spacing: -0.5px;
}'
);

