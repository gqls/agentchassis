-- ============================================================================
-- CSS & JS SNIPPET SYSTEM WITH SEMANTIC LABELLING
-- ============================================================================

-- Drop existing if needed (comment out for production)
DROP TABLE IF EXISTS css_snippets CASCADE;
DROP TABLE IF EXISTS js_snippets CASCADE;
DROP TABLE IF EXISTS theme_tags CASCADE;

-- ============================================================================
-- CSS/JS SNIPPET SYSTEM WITH SEMANTIC TAGGING
-- ============================================================================

-- ============================================================================
-- THEME TAGS - Semantic categories for matching
-- ============================================================================

CREATE TABLE IF NOT EXISTS theme_tags (
                                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    category VARCHAR(30) NOT NULL, -- 'mood', 'style', 'industry', 'audience', 'functional', 'color'
    description TEXT,
    related_tags JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP DEFAULT now()
    );

-- Insert semantic tags
INSERT INTO theme_tags (name, category, description, related_tags) VALUES
-- Mood tags
('energetic', 'mood', 'High energy, dynamic, action-oriented', '["bold", "modern", "youthful"]'::jsonb),
('calm', 'mood', 'Peaceful, serene, relaxed', '["minimal", "professional", "trustworthy"]'::jsonb),
('urgent', 'mood', 'Creates sense of immediacy and action', '["bold", "energetic", "conversion-focused"]'::jsonb),
('warm', 'mood', 'Friendly, approachable, comforting', '["personal", "trustworthy", "community"]'::jsonb),
('cool', 'mood', 'Professional, detached, sophisticated', '["modern", "minimal", "corporate"]'::jsonb),
('playful', 'mood', 'Fun, lighthearted, creative', '["youthful", "colorful", "creative"]'::jsonb),

-- Style tags
('bold', 'style', 'Strong contrasts, heavy fonts, impactful', '["energetic", "modern", "conversion-focused"]'::jsonb),
('minimal', 'style', 'Clean, lots of whitespace, simple', '["calm", "professional", "modern"]'::jsonb),
('luxurious', 'style', 'Premium, elegant, high-end', '["professional", "trustworthy", "corporate"]'::jsonb),
('organic', 'style', 'Natural, flowing, earthy', '["warm", "sustainable", "health"]'::jsonb),
('geometric', 'style', 'Sharp angles, structured, precise', '["modern", "tech", "corporate"]'::jsonb),
('retro', 'style', 'Vintage, nostalgic, classic', '["warm", "creative", "personal"]'::jsonb),

-- Industry-adjacent tags (semantic, not specific)
('tech', 'industry', 'Technology, software, digital', '["modern", "cool", "minimal"]'::jsonb),
('health', 'industry', 'Healthcare, wellness, fitness', '["calm", "trustworthy", "organic"]'::jsonb),
('finance', 'industry', 'Banking, insurance, investments', '["trustworthy", "professional", "corporate"]'::jsonb),
('creative', 'industry', 'Design, art, entertainment', '["playful", "bold", "colorful"]'::jsonb),
('ecommerce', 'industry', 'Online retail, shopping', '["conversion-focused", "trustworthy", "modern"]'::jsonb),
('saas', 'industry', 'Software as a service', '["tech", "modern", "professional"]'::jsonb),

-- Audience tags
('youthful', 'audience', 'Targeting younger demographics', '["energetic", "playful", "modern"]'::jsonb),
('professional', 'audience', 'B2B, enterprise, business', '["trustworthy", "corporate", "minimal"]'::jsonb),
('consumer', 'audience', 'B2C, general public', '["warm", "accessible", "conversion-focused"]'::jsonb),
('premium', 'audience', 'High-end, luxury market', '["luxurious", "minimal", "trustworthy"]'::jsonb),

-- Functional tags
('conversion-focused', 'functional', 'Optimized for conversions', '["bold", "urgent", "ecommerce"]'::jsonb),
('content-heavy', 'functional', 'Lots of text, articles, blogs', '["minimal", "calm", "readable"]'::jsonb),
('visual-heavy', 'functional', 'Image/video focused', '["bold", "modern", "creative"]'::jsonb),
('trustworthy', 'functional', 'Builds credibility and trust', '["professional", "calm", "corporate"]'::jsonb),

-- Color mood tags
('colorful', 'color', 'Vibrant, multi-colored palette', '["playful", "energetic", "creative"]'::jsonb),
('monochrome', 'color', 'Single color family, sophisticated', '["minimal", "luxurious", "professional"]'::jsonb),
('dark-mode', 'color', 'Dark backgrounds, light text', '["tech", "modern", "bold"]'::jsonb),
('light-mode', 'color', 'Light backgrounds, dark text', '["minimal", "calm", "professional"]'::jsonb),
('high-contrast', 'color', 'Strong color contrasts', '["bold", "conversion-focused", "accessible"]'::jsonb)
    ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- CSS THEMES - Add semantic_tags column if not exists
-- ============================================================================

ALTER TABLE css_themes ADD COLUMN IF NOT EXISTS semantic_tags JSONB DEFAULT '[]'::jsonb;

-- Update existing themes with semantic tags
UPDATE css_themes SET semantic_tags = '["professional", "trustworthy", "minimal", "light-mode"]'::jsonb WHERE name = 'default';
UPDATE css_themes SET semantic_tags = '["energetic", "bold", "urgent", "high-contrast"]'::jsonb WHERE name = 'boxing';
UPDATE css_themes SET semantic_tags = '["warm", "organic", "friendly", "light-mode"]'::jsonb WHERE name = 'bakery';
UPDATE css_themes SET semantic_tags = '["tech", "modern", "dark-mode", "bold"]'::jsonb WHERE name = 'tech';
UPDATE css_themes SET semantic_tags = '["professional", "trustworthy", "finance", "corporate"]'::jsonb WHERE name = 'professional-dark';

-- ============================================================================
-- NEW THEMES WITH SEMANTIC TAGS
-- ============================================================================

-- Clean/Minimal theme
INSERT INTO css_themes (name, description, css_content, semantic_tags)
VALUES (
           'clean-minimal',
           'Clean, minimal design with lots of whitespace',
           ':root {
           --primary: #2563eb;
           --primary-dark: #1d4ed8;
           --accent: #0ea5e9;
           --background: #ffffff;
           --surface: #f8fafc;
           --text: #1e293b;
           --text-light: #64748b;
           --border: #e2e8f0;
           --radius: 0.5rem;
           --shadow: 0 1px 3px rgba(0,0,0,0.1);
           --font-main: "Inter", -apple-system, BlinkMacSystemFont, sans-serif;
           --font-heading: "Inter", -apple-system, BlinkMacSystemFont, sans-serif;
           --line-height: 1.7;
           --max-width: 1200px;
           --section-padding: 5rem 1.5rem;
         }

         body {
           font-family: var(--font-main);
           line-height: var(--line-height);
           color: var(--text);
           background: var(--background);
         }

         .container { max-width: var(--max-width); margin: 0 auto; padding: 0 1.5rem; }
         section { padding: var(--section-padding); }

         h1, h2, h3, h4 {
           font-family: var(--font-heading);
           font-weight: 600;
           line-height: 1.2;
           color: var(--text);
         }

         h1 { font-size: 3rem; }
         h2 { font-size: 2.25rem; }
         h3 { font-size: 1.5rem; }

         .btn {
           display: inline-block;
           padding: 0.875rem 2rem;
           border-radius: var(--radius);
           font-weight: 500;
           text-decoration: none;
           transition: all 0.2s ease;
         }

         .btn-primary {
           background: var(--primary);
           color: white;
         }

         .btn-primary:hover {
           background: var(--primary-dark);
         }',
           '["calm", "minimal", "light-mode", "content-heavy", "professional"]'::jsonb
       )
    ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
                              semantic_tags = EXCLUDED.semantic_tags;

-- Bold/Conversion theme
INSERT INTO css_themes (name, description, css_content, semantic_tags)
VALUES (
           'bold-conversion',
           'High-contrast, conversion-focused design',
           ':root {
           --primary: #dc2626;
           --primary-dark: #b91c1c;
           --accent: #fbbf24;
           --background: #0f172a;
           --surface: #1e293b;
           --text: #f8fafc;
           --text-light: #94a3b8;
           --border: #334155;
           --radius: 0.75rem;
           --shadow: 0 4px 20px rgba(0,0,0,0.3);
           --font-main: "Inter", -apple-system, sans-serif;
           --font-heading: "Inter", -apple-system, sans-serif;
           --line-height: 1.6;
           --max-width: 1200px;
           --section-padding: 4rem 1.5rem;
         }

         body {
           font-family: var(--font-main);
           line-height: var(--line-height);
           color: var(--text);
           background: var(--background);
         }

         h1, h2, h3, h4 {
           font-family: var(--font-heading);
           font-weight: 800;
           line-height: 1.1;
           text-transform: uppercase;
         }

         h1 { font-size: 3.5rem; letter-spacing: -0.02em; }
         h2 { font-size: 2.5rem; }

         .btn {
           display: inline-block;
           padding: 1rem 2.5rem;
           border-radius: var(--radius);
           font-weight: 700;
           text-transform: uppercase;
           letter-spacing: 0.05em;
           text-decoration: none;
           transition: all 0.2s ease;
         }

         .btn-primary {
           background: var(--primary);
           color: white;
           box-shadow: 0 4px 15px rgba(220, 38, 38, 0.4);
         }

         .btn-primary:hover {
           background: var(--primary-dark);
           transform: translateY(-2px);
           box-shadow: 0 6px 20px rgba(220, 38, 38, 0.5);
         }

         .highlight { color: var(--accent); }',
           '["bold", "conversion-focused", "urgent", "high-contrast", "ecommerce"]'::jsonb
       )
    ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
                              semantic_tags = EXCLUDED.semantic_tags;

-- Warm/Friendly theme
INSERT INTO css_themes (name, description, css_content, semantic_tags)
VALUES (
           'warm-friendly',
           'Warm, approachable design for community/personal brands',
           ':root {
           --primary: #ea580c;
           --primary-dark: #c2410c;
           --accent: #65a30d;
           --background: #fffbeb;
           --surface: #fef3c7;
           --text: #451a03;
           --text-light: #78350f;
           --border: #fcd34d;
           --radius: 1rem;
           --shadow: 0 2px 10px rgba(234, 88, 12, 0.1);
           --font-main: "Nunito", -apple-system, sans-serif;
           --font-heading: "Nunito", -apple-system, sans-serif;
           --line-height: 1.7;
           --max-width: 1100px;
           --section-padding: 4rem 1.5rem;
         }

         body {
           font-family: var(--font-main);
           line-height: var(--line-height);
           color: var(--text);
           background: var(--background);
         }

         h1, h2, h3, h4 {
           font-family: var(--font-heading);
           font-weight: 700;
           line-height: 1.3;
         }

         h1 { font-size: 2.75rem; }
         h2 { font-size: 2rem; }

         .btn {
           display: inline-block;
           padding: 0.875rem 2rem;
           border-radius: var(--radius);
           font-weight: 600;
           text-decoration: none;
           transition: all 0.2s ease;
         }

         .btn-primary {
           background: var(--primary);
           color: white;
         }

         .btn-primary:hover {
           background: var(--primary-dark);
         }',
           '["warm", "personal", "trustworthy", "consumer", "light-mode"]'::jsonb
       )
    ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
                              semantic_tags = EXCLUDED.semantic_tags;

-- Tech/SaaS theme
INSERT INTO css_themes (name, description, css_content, semantic_tags)
VALUES (
           'tech-saas',
           'Modern tech/SaaS design with gradients',
           ':root {
           --primary: #8b5cf6;
           --primary-dark: #7c3aed;
           --accent: #06b6d4;
           --background: #0f0f23;
           --surface: #1a1a2e;
           --text: #e2e8f0;
           --text-light: #94a3b8;
           --border: #2d2d44;
           --radius: 0.75rem;
           --shadow: 0 4px 20px rgba(139, 92, 246, 0.2);
           --font-main: "Inter", -apple-system, sans-serif;
           --font-heading: "Inter", -apple-system, sans-serif;
           --line-height: 1.6;
           --max-width: 1200px;
           --section-padding: 5rem 1.5rem;
           --gradient: linear-gradient(135deg, var(--primary), var(--accent));
         }

         body {
           font-family: var(--font-main);
           line-height: var(--line-height);
           color: var(--text);
           background: var(--background);
         }

         h1, h2, h3, h4 {
           font-family: var(--font-heading);
           font-weight: 700;
           line-height: 1.2;
         }

         h1 { font-size: 3.5rem; }
         h2 { font-size: 2.5rem; }

         .gradient-text {
           background: var(--gradient);
           -webkit-background-clip: text;
           -webkit-text-fill-color: transparent;
           background-clip: text;
         }

         .btn {
           display: inline-block;
           padding: 1rem 2rem;
           border-radius: var(--radius);
           font-weight: 600;
           text-decoration: none;
           transition: all 0.3s ease;
         }

         .btn-primary {
           background: var(--gradient);
           color: white;
         }

         .btn-primary:hover {
           transform: translateY(-2px);
           box-shadow: var(--shadow);
         }',
           '["tech", "modern", "dark-mode", "saas", "cool"]'::jsonb
       )
    ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
                              semantic_tags = EXCLUDED.semantic_tags;

-- Luxury/Premium theme
INSERT INTO css_themes (name, description, css_content, semantic_tags)
VALUES (
           'luxury-premium',
           'Elegant, premium design for high-end brands',
           ':root {
           --primary: #a16207;
           --primary-dark: #854d0e;
           --accent: #d4af37;
           --background: #fafaf9;
           --surface: #f5f5f4;
           --text: #1c1917;
           --text-light: #57534e;
           --border: #d6d3d1;
           --radius: 0;
           --shadow: 0 1px 2px rgba(0,0,0,0.05);
           --font-main: "Cormorant Garamond", Georgia, serif;
           --font-heading: "Cormorant Garamond", Georgia, serif;
           --line-height: 1.8;
           --max-width: 1000px;
           --section-padding: 6rem 2rem;
         }

         body {
           font-family: var(--font-main);
           font-size: 1.125rem;
           line-height: var(--line-height);
           color: var(--text);
           background: var(--background);
         }

         h1, h2, h3, h4 {
           font-family: var(--font-heading);
           font-weight: 400;
           line-height: 1.2;
           letter-spacing: 0.05em;
         }

         h1 { font-size: 3.5rem; }
         h2 { font-size: 2.5rem; }

         .btn {
           display: inline-block;
           padding: 1rem 2.5rem;
           border: 1px solid var(--text);
           font-weight: 500;
           letter-spacing: 0.1em;
           text-transform: uppercase;
           text-decoration: none;
           transition: all 0.3s ease;
         }

         .btn-primary {
           background: var(--text);
           color: var(--background);
         }

         .btn-primary:hover {
           background: transparent;
           color: var(--text);
         }',
           '["luxurious", "premium", "minimal", "professional", "monochrome"]'::jsonb
       )
    ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
                              semantic_tags = EXCLUDED.semantic_tags;

-- ============================================================================
-- CSS SNIPPETS - Reusable CSS for components
-- ============================================================================

CREATE TABLE IF NOT EXISTS css_snippets (
                                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    css_content TEXT NOT NULL,
    semantic_tags JSONB DEFAULT '[]'::jsonb,
    applies_to JSONB DEFAULT '[]'::jsonb, -- component functions this applies to
    created_at TIMESTAMP DEFAULT now()
    );

-- Hover effects
INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES
                                                                                         ('hover-lift', 'Subtle lift on hover',
                                                                                          '.hover-lift { transition: transform 0.2s ease, box-shadow 0.2s ease; }
                                                                                          .hover-lift:hover { transform: translateY(-4px); box-shadow: 0 10px 25px rgba(0,0,0,0.1); }',
                                                                                          '["modern", "subtle", "professional"]'::jsonb,
                                                                                          '["card", "feature", "testimonial"]'::jsonb),

                                                                                         ('hover-scale', 'Scale up on hover',
                                                                                          '.hover-scale { transition: transform 0.2s ease; }
                                                                                          .hover-scale:hover { transform: scale(1.05); }',
                                                                                          '["energetic", "bold", "playful"]'::jsonb,
                                                                                          '["card", "image", "button"]'::jsonb),

                                                                                         ('hover-glow', 'Glow effect on hover',
                                                                                          '.hover-glow { transition: box-shadow 0.3s ease; }
                                                                                          .hover-glow:hover { box-shadow: 0 0 20px var(--primary, rgba(99, 102, 241, 0.5)); }',
                                                                                          '["tech", "modern", "dark-mode"]'::jsonb,
                                                                                          '["button", "card", "cta"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET css_content = EXCLUDED.css_content, semantic_tags = EXCLUDED.semantic_tags;

-- Border effects
INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES
                                                                                         ('border-gradient', 'Gradient border effect',
                                                                                          '.border-gradient {
                                                                                            border: 2px solid transparent;
                                                                                            background: linear-gradient(var(--background), var(--background)) padding-box,
                                                                                                        linear-gradient(135deg, var(--primary), var(--accent, var(--primary))) border-box;
                                                                                          }',
                                                                                          '["modern", "tech", "premium"]'::jsonb,
                                                                                          '["card", "feature", "cta"]'::jsonb),

                                                                                         ('border-animate', 'Animated border on hover',
                                                                                          '.border-animate {
                                                                                            position: relative;
                                                                                            overflow: hidden;
                                                                                          }
                                                                                          .border-animate::after {
                                                                                            content: "";
                                                                                            position: absolute;
                                                                                            bottom: 0;
                                                                                            left: 0;
                                                                                            width: 0;
                                                                                            height: 2px;
                                                                                            background: var(--primary);
                                                                                            transition: width 0.3s ease;
                                                                                          }
                                                                                          .border-animate:hover::after { width: 100%; }',
                                                                                          '["modern", "bold", "creative"]'::jsonb,
                                                                                          '["link", "nav-item", "card"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET css_content = EXCLUDED.css_content, semantic_tags = EXCLUDED.semantic_tags;

-- Animation snippets
INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES
                                                                                         ('fade-in-up', 'Fade in from below animation',
                                                                                          '@keyframes fadeInUp {
                                                                                            from { opacity: 0; transform: translateY(20px); }
                                                                                            to { opacity: 1; transform: translateY(0); }
                                                                                          }
                                                                                          .fade-in-up { animation: fadeInUp 0.6s ease forwards; }',
                                                                                          '["subtle", "professional", "modern"]'::jsonb,
                                                                                          '["hero", "section", "card"]'::jsonb),

                                                                                         ('pulse-attention', 'Pulsing attention animation',
                                                                                          '@keyframes pulseAttention {
                                                                                            0%, 100% { transform: scale(1); }
                                                                                            50% { transform: scale(1.05); }
                                                                                          }
                                                                                          .pulse-attention { animation: pulseAttention 2s ease-in-out infinite; }',
                                                                                          '["energetic", "bold", "dynamic"]'::jsonb,
                                                                                          '["cta", "button", "badge"]'::jsonb),

                                                                                         ('shake-attention', 'Shake animation for attention',
                                                                                          '@keyframes shake {
                                                                                            0%, 100% { transform: translateX(0); }
                                                                                            25% { transform: translateX(-5px); }
                                                                                            75% { transform: translateX(5px); }
                                                                                          }
                                                                                          .shake-attention:hover { animation: shake 0.5s ease; }',
                                                                                          '["urgent", "conversion-focused", "attention"]'::jsonb,
                                                                                          '["cta", "button", "alert"]'::jsonb),

                                                                                         ('float', 'Gentle floating animation',
                                                                                          '@keyframes float {
                                                                                            0%, 100% { transform: translateY(0); }
                                                                                            50% { transform: translateY(-10px); }
                                                                                          }
                                                                                          .float { animation: float 3s ease-in-out infinite; }',
                                                                                          '["modern", "creative", "bold"]'::jsonb,
                                                                                          '["hero-image", "icon", "decoration"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET css_content = EXCLUDED.css_content, semantic_tags = EXCLUDED.semantic_tags;

-- Button styles
INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES
                                                                                         ('btn-glass', 'Glassmorphism button style',
                                                                                          '.btn-glass {
                                                                                            background: rgba(255, 255, 255, 0.1);
                                                                                            backdrop-filter: blur(10px);
                                                                                            border: 1px solid rgba(255, 255, 255, 0.2);
                                                                                          }
                                                                                          .btn-glass:hover {
                                                                                            background: rgba(255, 255, 255, 0.2);
                                                                                          }',
                                                                                          '["modern", "tech", "premium"]'::jsonb,
                                                                                          '["button", "cta"]'::jsonb),

                                                                                         ('btn-outline-fill', 'Outline button that fills on hover',
                                                                                          '.btn-outline-fill {
                                                                                            background: transparent;
                                                                                            border: 2px solid var(--primary);
                                                                                            color: var(--primary);
                                                                                            transition: all 0.3s ease;
                                                                                          }
                                                                                          .btn-outline-fill:hover {
                                                                                            background: var(--primary);
                                                                                            color: white;
                                                                                          }',
                                                                                          '["modern", "bold", "creative"]'::jsonb,
                                                                                          '["button", "cta"]'::jsonb),

                                                                                         ('btn-minimal', 'Minimal underline button',
                                                                                          '.btn-minimal {
                                                                                            background: none;
                                                                                            border: none;
                                                                                            padding: 0.5rem 0;
                                                                                            border-bottom: 1px solid currentColor;
                                                                                            transition: opacity 0.2s ease;
                                                                                          }
                                                                                          .btn-minimal:hover { opacity: 0.7; }',
                                                                                          '["calm", "minimal", "professional"]'::jsonb,
                                                                                          '["link", "button"]'::jsonb),

                                                                                         ('btn-3d', '3D push button effect',
                                                                                          '.btn-3d {
                                                                                            box-shadow: 0 4px 0 var(--primary-dark, #1d4ed8);
                                                                                            transform: translateY(0);
                                                                                            transition: all 0.1s ease;
                                                                                          }
                                                                                          .btn-3d:hover {
                                                                                            transform: translateY(2px);
                                                                                            box-shadow: 0 2px 0 var(--primary-dark, #1d4ed8);
                                                                                          }',
                                                                                          '["bold", "retro", "playful"]'::jsonb,
                                                                                          '["button", "cta"]'::jsonb),

                                                                                         ('btn-icon', 'Button with icon alignment',
                                                                                          '.btn-icon {
                                                                                            display: inline-flex;
                                                                                            align-items: center;
                                                                                            gap: 0.5rem;
                                                                                          }
                                                                                          .btn-icon svg { width: 1.25em; height: 1.25em; }',
                                                                                          '["subtle", "professional", "minimal"]'::jsonb,
                                                                                          '["button", "cta", "link"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET css_content = EXCLUDED.css_content, semantic_tags = EXCLUDED.semantic_tags;

-- Card styles
INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES
                                                                                         ('card-glass', 'Glassmorphism card',
                                                                                          '.card-glass {
                                                                                            background: rgba(255, 255, 255, 0.05);
                                                                                            backdrop-filter: blur(10px);
                                                                                            border: 1px solid rgba(255, 255, 255, 0.1);
                                                                                            border-radius: var(--radius, 1rem);
                                                                                          }',
                                                                                          '["tech", "modern", "geometric"]'::jsonb,
                                                                                          '["card", "feature", "testimonial"]'::jsonb),

                                                                                         ('card-bordered', 'Simple bordered card',
                                                                                          '.card-bordered {
                                                                                            border: 1px solid var(--border);
                                                                                            border-radius: var(--radius, 0.5rem);
                                                                                            padding: 1.5rem;
                                                                                          }',
                                                                                          '["professional", "modern"]'::jsonb,
                                                                                          '["card", "feature"]'::jsonb),

                                                                                         ('card-shadow', 'Card with shadow depth',
                                                                                          '.card-shadow {
                                                                                            background: var(--surface, white);
                                                                                            border-radius: var(--radius, 0.75rem);
                                                                                            box-shadow: 0 4px 20px rgba(0,0,0,0.08);
                                                                                            padding: 2rem;
                                                                                          }',
                                                                                          '["accessible", "professional"]'::jsonb,
                                                                                          '["card", "feature", "testimonial"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET css_content = EXCLUDED.css_content, semantic_tags = EXCLUDED.semantic_tags;

-- Form styles
INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES
    ('input-modern', 'Modern input field styling',
     '.input-modern {
       width: 100%;
       padding: 0.875rem 1rem;
       border: 1px solid var(--border);
       border-radius: var(--radius, 0.5rem);
       background: var(--surface, white);
       color: var(--text);
       transition: border-color 0.2s ease, box-shadow 0.2s ease;
     }
     .input-modern:focus {
       outline: none;
       border-color: var(--primary);
       box-shadow: 0 0 0 3px rgba(var(--primary-rgb, 37, 99, 235), 0.1);
     }',
     '["professional", "branded"]'::jsonb,
     '["form", "input", "newsletter"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET css_content = EXCLUDED.css_content, semantic_tags = EXCLUDED.semantic_tags;

-- Layout snippets
INSERT INTO css_snippets (name, description, css_content, semantic_tags, applies_to) VALUES
    ('responsive-grid', 'Responsive grid layout',
     '.grid-responsive {
       display: grid;
       grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
       gap: 2rem;
     }',
     '["responsive", "mobile"]'::jsonb,
     '["features", "cards", "gallery"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET css_content = EXCLUDED.css_content, semantic_tags = EXCLUDED.semantic_tags;

-- ============================================================================
-- JS SNIPPETS - Reusable JavaScript for components
-- ============================================================================

CREATE TABLE IF NOT EXISTS js_snippets (
                                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    js_content TEXT NOT NULL,
    semantic_tags JSONB DEFAULT '[]'::jsonb,
    applies_to JSONB DEFAULT '[]'::jsonb,
    dependencies JSONB DEFAULT '[]'::jsonb, -- external libraries needed
    created_at TIMESTAMP DEFAULT now()
    );

-- Scroll animations
INSERT INTO js_snippets (name, description, js_content, semantic_tags, applies_to) VALUES
                                                                                       ('scroll-reveal', 'Reveal elements on scroll',
                                                                                        'document.addEventListener("DOMContentLoaded", function() {
                                                                                          const observer = new IntersectionObserver((entries) => {
                                                                                            entries.forEach(entry => {
                                                                                              if (entry.isIntersecting) {
                                                                                                entry.target.classList.add("revealed");
                                                                                                observer.unobserve(entry.target);
                                                                                              }
                                                                                            });
                                                                                          }, { threshold: 0.1 });

                                                                                          document.querySelectorAll(".reveal-on-scroll").forEach(el => observer.observe(el));
                                                                                        });',
                                                                                        '["professional", "modern"]'::jsonb,
                                                                                        '["section", "card", "feature"]'::jsonb),

                                                                                       ('smooth-scroll', 'Smooth scroll to anchors',
                                                                                        'document.querySelectorAll(''a[href^="#"]'').forEach(anchor => {
                                                                                          anchor.addEventListener("click", function(e) {
                                                                                            e.preventDefault();
                                                                                            const target = document.querySelector(this.getAttribute("href"));
                                                                                            if (target) {
                                                                                              target.scrollIntoView({ behavior: "smooth", block: "start" });
                                                                                            }
                                                                                          });
                                                                                        });',
                                                                                        '["professional", "modern"]'::jsonb,
                                                                                        '["navigation", "link"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET js_content = EXCLUDED.js_content, semantic_tags = EXCLUDED.semantic_tags;

-- Interactive elements
INSERT INTO js_snippets (name, description, js_content, semantic_tags, applies_to) VALUES
                                                                                       ('counter-animate', 'Animate numbers counting up',
                                                                                        'function animateCounter(el) {
                                                                                          const target = parseInt(el.dataset.target);
                                                                                          const duration = parseInt(el.dataset.duration) || 2000;
                                                                                          const start = 0;
                                                                                          const startTime = performance.now();

                                                                                          function update(currentTime) {
                                                                                            const elapsed = currentTime - startTime;
                                                                                            const progress = Math.min(elapsed / duration, 1);
                                                                                            el.textContent = Math.floor(progress * target).toLocaleString();
                                                                                            if (progress < 1) requestAnimationFrame(update);
                                                                                          }
                                                                                          requestAnimationFrame(update);
                                                                                        }

                                                                                        const counterObserver = new IntersectionObserver((entries) => {
                                                                                          entries.forEach(entry => {
                                                                                            if (entry.isIntersecting) {
                                                                                              animateCounter(entry.target);
                                                                                              counterObserver.unobserve(entry.target);
                                                                                            }
                                                                                          });
                                                                                        }, { threshold: 0.5 });

                                                                                        document.querySelectorAll(".counter").forEach(el => counterObserver.observe(el));',
                                                                                        '["modern", "dynamic"]'::jsonb,
                                                                                        '["stats", "numbers", "social-proof"]'::jsonb),

                                                                                       ('typing-effect', 'Typewriter text effect',
                                                                                        'function typeWriter(el) {
                                                                                          const text = el.dataset.text;
                                                                                          const speed = parseInt(el.dataset.speed) || 50;
                                                                                          let i = 0;
                                                                                          el.textContent = "";

                                                                                          function type() {
                                                                                            if (i < text.length) {
                                                                                              el.textContent += text.charAt(i);
                                                                                              i++;
                                                                                              setTimeout(type, speed);
                                                                                            }
                                                                                          }
                                                                                          type();
                                                                                        }

                                                                                        document.querySelectorAll(".typing").forEach(el => {
                                                                                          const observer = new IntersectionObserver((entries) => {
                                                                                            if (entries[0].isIntersecting) {
                                                                                              typeWriter(el);
                                                                                              observer.disconnect();
                                                                                            }
                                                                                          });
                                                                                          observer.observe(el);
                                                                                        });',
                                                                                        '["energetic", "conversion-focused"]'::jsonb,
                                                                                        '["hero", "headline"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET js_content = EXCLUDED.js_content, semantic_tags = EXCLUDED.semantic_tags;

-- UI interactions
INSERT INTO js_snippets (name, description, js_content, semantic_tags, applies_to) VALUES
                                                                                       ('accordion', 'Accordion/FAQ toggle',
                                                                                        'document.querySelectorAll(".accordion-trigger").forEach(trigger => {
                                                                                          trigger.addEventListener("click", function() {
                                                                                            const content = this.nextElementSibling;
                                                                                            const isOpen = content.style.maxHeight;

                                                                                            // Close all others
                                                                                            document.querySelectorAll(".accordion-content").forEach(c => {
                                                                                              c.style.maxHeight = null;
                                                                                              c.previousElementSibling.classList.remove("active");
                                                                                            });

                                                                                            // Toggle current
                                                                                            if (!isOpen) {
                                                                                              content.style.maxHeight = content.scrollHeight + "px";
                                                                                              this.classList.add("active");
                                                                                            }
                                                                                          });
                                                                                        });',
                                                                                        '["professional", "faq"]'::jsonb,
                                                                                        '["faq", "accordion"]'::jsonb),

                                                                                       ('copy-to-clipboard', 'Copy text to clipboard',
                                                                                        'document.querySelectorAll(".copy-btn").forEach(btn => {
                                                                                          btn.addEventListener("click", async function() {
                                                                                            const text = this.dataset.copy || this.previousElementSibling.textContent;
                                                                                            await navigator.clipboard.writeText(text);
                                                                                            const original = this.textContent;
                                                                                            this.textContent = "Copied!";
                                                                                            setTimeout(() => this.textContent = original, 2000);
                                                                                          });
                                                                                        });',
                                                                                        '["tech", "utility"]'::jsonb,
                                                                                        '["code", "share"]'::jsonb),

                                                                                       ('form-validation', 'Basic form validation',
                                                                                        'document.querySelectorAll("form[data-validate]").forEach(form => {
                                                                                          form.addEventListener("submit", function(e) {
                                                                                            let valid = true;
                                                                                            this.querySelectorAll("[required]").forEach(field => {
                                                                                              if (!field.value.trim()) {
                                                                                                valid = false;
                                                                                                field.classList.add("error");
                                                                                              } else {
                                                                                                field.classList.remove("error");
                                                                                              }
                                                                                            });

                                                                                            const emailFields = this.querySelectorAll(''[type="email"]'');
                                                                                            emailFields.forEach(field => {
                                                                                              if (field.value && !field.value.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/)) {
                                                                                                valid = false;
                                                                                                field.classList.add("error");
                                                                                              }
                                                                                            });

                                                                                            if (!valid) e.preventDefault();
                                                                                          });
                                                                                        });',
                                                                                        '["professional", "conversion-focused"]'::jsonb,
                                                                                        '["form", "newsletter", "contact"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET js_content = EXCLUDED.js_content, semantic_tags = EXCLUDED.semantic_tags;

-- Performance/UX
INSERT INTO js_snippets (name, description, js_content, semantic_tags, applies_to) VALUES
                                                                                       ('lazy-load-images', 'Lazy load images',
                                                                                        'if ("loading" in HTMLImageElement.prototype) {
                                                                                          document.querySelectorAll(''img[loading="lazy"]'').forEach(img => {
                                                                                            img.src = img.dataset.src;
                                                                                          });
                                                                                        } else {
                                                                                          const imageObserver = new IntersectionObserver((entries) => {
                                                                                            entries.forEach(entry => {
                                                                                              if (entry.isIntersecting) {
                                                                                                const img = entry.target;
                                                                                                img.src = img.dataset.src;
                                                                                                imageObserver.unobserve(img);
                                                                                              }
                                                                                            });
                                                                                          });
                                                                                          document.querySelectorAll(''img[data-src]'').forEach(img => imageObserver.observe(img));
                                                                                        }',
                                                                                        '["performance", "modern"]'::jsonb,
                                                                                        '["image", "gallery"]'::jsonb),

                                                                                       ('mobile-menu-toggle', 'Mobile menu toggle',
                                                                                        'const menuToggle = document.querySelector(".menu-toggle");
                                                                                        const mobileMenu = document.querySelector(".mobile-menu");

                                                                                        if (menuToggle && mobileMenu) {
                                                                                          menuToggle.addEventListener("click", () => {
                                                                                            mobileMenu.classList.toggle("open");
                                                                                            menuToggle.classList.toggle("active");
                                                                                            document.body.classList.toggle("menu-open");
                                                                                          });
                                                                                        }',
                                                                                        '["modern", "tech"]'::jsonb,
                                                                                        '["navigation", "header"]'::jsonb)
    ON CONFLICT (name) DO UPDATE SET js_content = EXCLUDED.js_content, semantic_tags = EXCLUDED.semantic_tags;

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Find themes by semantic tags (returns themes matching ANY of the tags)
CREATE OR REPLACE FUNCTION find_themes_by_tags(search_tags TEXT[])
RETURNS TABLE (
  theme_name VARCHAR,
  theme_description TEXT,
  match_count INT,
  matched_tags TEXT[]
) AS $$
BEGIN
RETURN QUERY
SELECT
    ct.name,
    ct.description,
    (SELECT COUNT(*)::INT FROM jsonb_array_elements_text(ct.semantic_tags) t WHERE t = ANY(search_tags)),
    ARRAY(SELECT t FROM jsonb_array_elements_text(ct.semantic_tags) t WHERE t = ANY(search_tags))
FROM css_themes ct
WHERE ct.semantic_tags ?| search_tags
ORDER BY (SELECT COUNT(*) FROM jsonb_array_elements_text(ct.semantic_tags) t WHERE t = ANY(search_tags)) DESC;
END;
$$ LANGUAGE plpgsql;

-- Find CSS snippets by semantic tags
CREATE OR REPLACE FUNCTION find_css_snippets_by_tags(search_tags TEXT[])
RETURNS TABLE (
  snippet_name VARCHAR,
  snippet_description TEXT,
  css_content TEXT,
  match_count INT
) AS $$
BEGIN
RETURN QUERY
SELECT
    cs.name,
    cs.description,
    cs.css_content,
    (SELECT COUNT(*)::INT FROM jsonb_array_elements_text(cs.semantic_tags) t WHERE t = ANY(search_tags))
FROM css_snippets cs
WHERE cs.semantic_tags ?| search_tags
ORDER BY (SELECT COUNT(*) FROM jsonb_array_elements_text(cs.semantic_tags) t WHERE t = ANY(search_tags)) DESC;
END;
$$ LANGUAGE plpgsql;

-- Find JS snippets by semantic tags
CREATE OR REPLACE FUNCTION find_js_snippets_by_tags(search_tags TEXT[])
RETURNS TABLE (
  snippet_name VARCHAR,
  snippet_description TEXT,
  js_content TEXT,
  match_count INT
) AS $$
BEGIN
RETURN QUERY
SELECT
    js.name,
    js.description,
    js.js_content,
    (SELECT COUNT(*)::INT FROM jsonb_array_elements_text(js.semantic_tags) t WHERE t = ANY(search_tags))
FROM js_snippets js
WHERE js.semantic_tags ?| search_tags
ORDER BY (SELECT COUNT(*) FROM jsonb_array_elements_text(js.semantic_tags) t WHERE t = ANY(search_tags)) DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- USAGE EXAMPLES
-- ============================================================================
-- SELECT * FROM find_themes_by_tags(ARRAY['modern', 'tech', 'conversion-focused']);
-- SELECT * FROM find_css_snippets_by_tags(ARRAY['modern', 'hover']);
-- SELECT * FROM find_js_snippets_by_tags(ARRAY['professional', 'modern']);

