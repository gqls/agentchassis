-- ============================================================================
-- CSS & JS SNIPPET SYSTEM WITH SEMANTIC LABELLING
-- ============================================================================

-- Drop existing if needed (comment out for production)
-- DROP TABLE IF EXISTS css_snippets CASCADE;
-- DROP TABLE IF EXISTS js_snippets CASCADE;
-- DROP TABLE IF EXISTS theme_tags CASCADE;

-- ============================================================================
-- THEME TAGS - Semantic labelling system
-- ============================================================================
CREATE TABLE IF NOT EXISTS theme_tags (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
name TEXT NOT NULL UNIQUE,
category TEXT NOT NULL,  -- 'mood', 'industry', 'style', 'audience'
description TEXT,
related_tags TEXT[],     -- tags that pair well with this one
created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert semantic tags
INSERT INTO theme_tags (name, category, description, related_tags) VALUES
-- Mood tags
('energetic', 'mood', 'High energy, dynamic, action-oriented', ARRAY['bold', 'modern', 'youthful']),
('calm', 'mood', 'Peaceful, serene, relaxed', ARRAY['minimal', 'professional', 'trustworthy']),
('urgent', 'mood', 'Creates sense of immediacy and action', ARRAY['bold', 'energetic', 'conversion-focused']),
('warm', 'mood', 'Friendly, approachable, comforting', ARRAY['personal', 'trustworthy', 'community']),
('cool', 'mood', 'Professional, detached, sophisticated', ARRAY['modern', 'minimal', 'corporate']),
('playful', 'mood', 'Fun, lighthearted, creative', ARRAY['youthful', 'colorful', 'creative']),

-- Style tags
('bold', 'style', 'Strong contrasts, heavy fonts, impactful', ARRAY['energetic', 'modern', 'conversion-focused']),
('minimal', 'style', 'Clean, lots of whitespace, simple', ARRAY['calm', 'professional', 'modern']),
('luxurious', 'style', 'Premium, elegant, high-end', ARRAY['professional', 'trustworthy', 'corporate']),
('organic', 'style', 'Natural, flowing, earthy', ARRAY['warm', 'sustainable', 'health']),
('geometric', 'style', 'Sharp angles, structured, precise', ARRAY['modern', 'tech', 'corporate']),
('retro', 'style', 'Vintage, nostalgic, classic', ARRAY['warm', 'creative', 'personal']),

-- Industry-adjacent tags (semantic, not specific)
('tech', 'industry', 'Technology, software, digital', ARRAY['modern', 'cool', 'minimal']),
('health', 'industry', 'Healthcare, wellness, fitness', ARRAY['calm', 'trustworthy', 'organic']),
('finance', 'industry', 'Banking, insurance, investments', ARRAY['trustworthy', 'professional', 'corporate']),
('creative', 'industry', 'Design, art, entertainment', ARRAY['playful', 'bold', 'colorful']),
('ecommerce', 'industry', 'Online retail, shopping', ARRAY['conversion-focused', 'trustworthy', 'modern']),
('saas', 'industry', 'Software as a service', ARRAY['tech', 'modern', 'professional']),

-- Audience tags
('youthful', 'audience', 'Targeting younger demographics', ARRAY['energetic', 'playful', 'modern']),
('professional', 'audience', 'B2B, enterprise, business', ARRAY['trustworthy', 'corporate', 'minimal']),
('consumer', 'audience', 'B2C, general public', ARRAY['warm', 'accessible', 'conversion-focused']),
('premium', 'audience', 'High-end, luxury market', ARRAY['luxurious', 'minimal', 'trustworthy']),

-- Functional tags
('conversion-focused', 'functional', 'Optimized for conversions', ARRAY['bold', 'urgent', 'ecommerce']),
('content-heavy', 'functional', 'Lots of text, articles, blogs', ARRAY['minimal', 'calm', 'readable']),
('visual-heavy', 'functional', 'Image/video focused', ARRAY['bold', 'modern', 'creative']),
('trustworthy', 'functional', 'Builds credibility and trust', ARRAY['professional', 'calm', 'corporate']),

-- Color mood tags
('colorful', 'color', 'Vibrant, multi-colored palette', ARRAY['playful', 'energetic', 'creative']),
('monochrome', 'color', 'Single color family, sophisticated', ARRAY['minimal', 'luxurious', 'professional']),
('dark-mode', 'color', 'Dark backgrounds, light text', ARRAY['tech', 'modern', 'bold']),
('light-mode', 'color', 'Light backgrounds, dark text', ARRAY['minimal', 'calm', 'professional']),
('high-contrast', 'color', 'Strong color contrasts', ARRAY['bold', 'conversion-focused', 'accessible'])
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- CSS THEMES - Updated with semantic tags
-- ============================================================================
ALTER TABLE css_themes ADD COLUMN IF NOT EXISTS semantic_tags TEXT[];
ALTER TABLE css_themes ADD COLUMN IF NOT EXISTS color_palette JSONB;
ALTER TABLE css_themes ADD COLUMN IF NOT EXISTS typography JSONB;

-- Update existing themes with semantic tags
UPDATE css_themes SET semantic_tags = ARRAY['professional', 'trustworthy', 'minimal', 'light-mode'] WHERE name = 'default';
UPDATE css_themes SET semantic_tags = ARRAY['energetic', 'bold', 'urgent', 'high-contrast'] WHERE name = 'boxing';
UPDATE css_themes SET semantic_tags = ARRAY['warm', 'organic', 'friendly', 'light-mode'] WHERE name = 'bakery';
UPDATE css_themes SET semantic_tags = ARRAY['tech', 'modern', 'dark-mode', 'bold'] WHERE name = 'tech';
UPDATE css_themes SET semantic_tags = ARRAY['professional', 'trustworthy', 'finance', 'corporate'] WHERE name = 'professional-dark';

-- Insert additional semantic themes
INSERT INTO css_themes (name, display_name, description, category, semantic_tags, css_content)
VALUES
(
'calm-minimal',
'Calm Minimal',
'Clean, spacious design with soft colors for content-focused sites',
'minimal',
ARRAY['calm', 'minimal', 'light-mode', 'content-heavy', 'professional'],
':root {
--color-primary: #6366f1;
--color-primary-hover: #4f46e5;
--color-primary-text: #ffffff;
--color-secondary: #a5b4fc;
--color-secondary-hover: #818cf8;
--color-secondary-text: #1e1b4b;
--color-accent: #f472b6;

    --color-text: #374151;
    --color-text-muted: #6b7280;
    --color-heading: #111827;
    --color-background: #fafafa;
    --color-border: #e5e7eb;

    --color-header-bg: #ffffff;
    --color-header-text: #111827;
    --color-hero-title: #111827;
    --color-hero-subtitle: #4b5563;
    --color-card-bg: #ffffff;
    --color-cta-bg: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    --color-cta-text: #ffffff;
    --color-footer-bg: #f9fafb;
    --color-footer-text: #374151;

    --border-radius: 0.75rem;
    --shadow: 0 1px 2px rgba(0,0,0,0.05);
    --shadow-lg: 0 4px 6px rgba(0,0,0,0.07);
}

body {
font-family: "Inter", -apple-system, sans-serif;
letter-spacing: -0.01em;
}

.section { padding: 6rem 1rem; }
.container { max-width: 1000px; }'
),
(
'bold-conversion',
'Bold Conversion',
'High-contrast, action-oriented design for landing pages',
'conversion',
ARRAY['bold', 'conversion-focused', 'urgent', 'high-contrast', 'ecommerce'],
':root {
--color-primary: #f97316;
--color-primary-hover: #ea580c;
--color-primary-text: #ffffff;
--color-secondary: #0ea5e9;
--color-secondary-hover: #0284c7;
--color-secondary-text: #ffffff;
--color-accent: #22c55e;

    --color-text: #18181b;
    --color-text-muted: #52525b;
    --color-heading: #09090b;
    --color-background: #ffffff;
    --color-border: #d4d4d8;

    --color-header-bg: #09090b;
    --color-header-text: #ffffff;
    --color-hero-title: #09090b;
    --color-hero-subtitle: #3f3f46;
    --color-card-bg: #fafafa;
    --color-cta-bg: linear-gradient(135deg, #f97316 0%, #dc2626 100%);
    --color-cta-text: #ffffff;
    --color-footer-bg: #18181b;
    --color-footer-text: #d4d4d8;

    --border-radius: 0.5rem;
    --shadow: 0 4px 12px rgba(0,0,0,0.15);
    --shadow-lg: 0 8px 24px rgba(0,0,0,0.2);
}

body {
font-family: "Inter", -apple-system, sans-serif;
font-weight: 500;
}

.button {
font-weight: 700;
text-transform: uppercase;
letter-spacing: 0.05em;
}

.hero__title {
font-size: 4rem;
font-weight: 900;
}'
),
(
'warm-friendly',
'Warm Friendly',
'Approachable, personal design for service businesses',
'warm',
ARRAY['warm', 'personal', 'trustworthy', 'consumer', 'light-mode'],
':root {
--color-primary: #059669;
--color-primary-hover: #047857;
--color-primary-text: #ffffff;
--color-secondary: #fbbf24;
--color-secondary-hover: #f59e0b;
--color-secondary-text: #1c1917;
--color-accent: #f472b6;

    --color-text: #44403c;
    --color-text-muted: #78716c;
    --color-heading: #1c1917;
    --color-background: #fffbeb;
    --color-border: #e7e5e4;

    --color-header-bg: #ffffff;
    --color-header-text: #1c1917;
    --color-hero-title: #1c1917;
    --color-hero-subtitle: #57534e;
    --color-card-bg: #ffffff;
    --color-cta-bg: linear-gradient(135deg, #059669 0%, #10b981 100%);
    --color-cta-text: #ffffff;
    --color-footer-bg: #fef3c7;
    --color-footer-text: #44403c;

    --border-radius: 1rem;
    --shadow: 0 2px 8px rgba(0,0,0,0.08);
    --shadow-lg: 0 4px 16px rgba(0,0,0,0.1);
}

body {
font-family: "Nunito", "Segoe UI", sans-serif;
}

.section__title {
font-weight: 700;
}

.card {
border-radius: 1.5rem;
}'
),
(
'dark-modern',
'Dark Modern',
'Sleek dark theme for tech and SaaS products',
'dark',
ARRAY['tech', 'modern', 'dark-mode', 'saas', 'cool'],
':root {
--color-primary: #6366f1;
--color-primary-hover: #4f46e5;
--color-primary-text: #ffffff;
--color-secondary: #22d3ee;
--color-secondary-hover: #06b6d4;
--color-secondary-text: #0f172a;
--color-accent: #f472b6;

    --color-text: #cbd5e1;
    --color-text-muted: #94a3b8;
    --color-heading: #f1f5f9;
    --color-background: #0f172a;
    --color-border: #334155;

    --color-header-bg: #020617;
    --color-header-text: #f1f5f9;
    --color-hero-title: #f8fafc;
    --color-hero-subtitle: #cbd5e1;
    --color-card-bg: #1e293b;
    --color-cta-bg: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
    --color-cta-text: #ffffff;
    --color-footer-bg: #020617;
    --color-footer-text: #94a3b8;

    --border-radius: 0.75rem;
    --shadow: 0 0 0 1px rgba(148, 163, 184, 0.1);
    --shadow-lg: 0 0 30px rgba(99, 102, 241, 0.15);
}

body {
font-family: "Inter", -apple-system, sans-serif;
}

.card {
border: 1px solid var(--color-border);
}'
),
(
'premium-elegant',
'Premium Elegant',
'Luxurious, refined design for high-end products',
'premium',
ARRAY['luxurious', 'premium', 'minimal', 'professional', 'monochrome'],
':root {
--color-primary: #0f172a;
--color-primary-hover: #1e293b;
--color-primary-text: #ffffff;
--color-secondary: #d4af37;
--color-secondary-hover: #b8972f;
--color-secondary-text: #0f172a;
--color-accent: #d4af37;

    --color-text: #334155;
    --color-text-muted: #64748b;
    --color-heading: #0f172a;
    --color-background: #fafaf9;
    --color-border: #e2e8f0;

    --color-header-bg: #0f172a;
    --color-header-text: #fafaf9;
    --color-hero-title: #0f172a;
    --color-hero-subtitle: #475569;
    --color-card-bg: #ffffff;
    --color-cta-bg: #0f172a;
    --color-cta-text: #fafaf9;
    --color-footer-bg: #0f172a;
    --color-footer-text: #e2e8f0;

    --border-radius: 0;
    --shadow: none;
    --shadow-lg: 0 25px 50px rgba(0,0,0,0.1);
}

body {
font-family: "Cormorant Garamond", Georgia, serif;
letter-spacing: 0.02em;
}

.button {
border-radius: 0;
letter-spacing: 0.15em;
text-transform: uppercase;
font-size: 0.875rem;
}

.section__title {
font-weight: 400;
letter-spacing: 0.1em;
text-transform: uppercase;
}'
)
ON CONFLICT (name) DO UPDATE SET
semantic_tags = EXCLUDED.semantic_tags,
css_content = EXCLUDED.css_content;

-- ============================================================================
-- CSS SNIPPETS - Reusable CSS effects and patterns
-- ============================================================================
CREATE TABLE IF NOT EXISTS css_snippets (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
name TEXT NOT NULL,
function TEXT NOT NULL,           -- semantic function: "hover-lift", "fade-in", "gradient-bg"
category TEXT NOT NULL,           -- "animation", "effect", "pattern", "utility", "interaction"
semantic_tags TEXT[],             -- tags for matching
css_content TEXT NOT NULL,
selector_prefix TEXT DEFAULT '',  -- e.g., ".card" to scope the CSS
dependencies TEXT[],              -- other snippets this requires
description TEXT,
created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_css_snippets_function ON css_snippets(function);
CREATE INDEX IF NOT EXISTS idx_css_snippets_category ON css_snippets(category);
CREATE INDEX IF NOT EXISTS idx_css_snippets_tags ON css_snippets USING GIN(semantic_tags);

-- Insert CSS snippets
INSERT INTO css_snippets (name, function, category, semantic_tags, css_content, description) VALUES
-- Hover effects
('Lift on Hover', 'hover-lift', 'interaction',
ARRAY['modern', 'subtle', 'professional'],
'.hover-lift { transition: transform 0.3s ease, box-shadow 0.3s ease; }
.hover-lift:hover { transform: translateY(-4px); box-shadow: 0 12px 24px rgba(0,0,0,0.15); }',
'Subtle lift effect on hover with shadow'),

('Scale on Hover', 'hover-scale', 'interaction',
ARRAY['energetic', 'bold', 'playful'],
'.hover-scale { transition: transform 0.2s ease; }
.hover-scale:hover { transform: scale(1.05); }',
'Slight scale up on hover'),

('Glow on Hover', 'hover-glow', 'interaction',
ARRAY['tech', 'modern', 'dark-mode'],
'.hover-glow { transition: box-shadow 0.3s ease; }
.hover-glow:hover { box-shadow: 0 0 20px var(--color-primary, #6366f1); }',
'Glowing effect on hover'),

-- Animations
('Fade In Up', 'animation-fade-in-up', 'animation',
ARRAY['subtle', 'professional', 'modern'],
'@keyframes fadeInUp {
from { opacity: 0; transform: translateY(20px); }
to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up { animation: fadeInUp 0.6s ease-out forwards; }
.animate-fade-in-up-delay-1 { animation-delay: 0.1s; }
.animate-fade-in-up-delay-2 { animation-delay: 0.2s; }
.animate-fade-in-up-delay-3 { animation-delay: 0.3s; }',
'Elements fade in from below'),

('Slide In Left', 'animation-slide-in-left', 'animation',
ARRAY['energetic', 'bold', 'dynamic'],
'@keyframes slideInLeft {
from { opacity: 0; transform: translateX(-30px); }
to { opacity: 1; transform: translateX(0); }
}
.animate-slide-in-left { animation: slideInLeft 0.5s ease-out forwards; }',
'Elements slide in from the left'),

('Pulse', 'animation-pulse', 'animation',
ARRAY['urgent', 'conversion-focused', 'attention'],
'@keyframes pulse {
0%, 100% { transform: scale(1); }
50% { transform: scale(1.05); }
}
.animate-pulse { animation: pulse 2s ease-in-out infinite; }',
'Gentle pulsing effect for CTAs'),

('Gradient Shift', 'animation-gradient', 'animation',
ARRAY['modern', 'creative', 'bold'],
'@keyframes gradientShift {
0% { background-position: 0% 50%; }
50% { background-position: 100% 50%; }
100% { background-position: 0% 50%; }
}
.animate-gradient {
background: linear-gradient(-45deg, var(--color-primary), var(--color-secondary), var(--color-accent, var(--color-primary)));
background-size: 400% 400%;
animation: gradientShift 8s ease infinite;
}',
'Animated gradient background'),

-- Visual effects
('Glass Morphism', 'effect-glass', 'effect',
ARRAY['modern', 'tech', 'premium'],
'.glass {
background: rgba(255, 255, 255, 0.1);
backdrop-filter: blur(10px);
-webkit-backdrop-filter: blur(10px);
border: 1px solid rgba(255, 255, 255, 0.2);
}',
'Frosted glass effect'),

('Gradient Text', 'effect-gradient-text', 'effect',
ARRAY['modern', 'bold', 'creative'],
'.gradient-text {
background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
-webkit-background-clip: text;
-webkit-text-fill-color: transparent;
background-clip: text;
}',
'Gradient fill for text'),

('Shadow Soft', 'effect-shadow-soft', 'effect',
ARRAY['calm', 'minimal', 'professional'],
'.shadow-soft { box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
.shadow-soft-lg { box-shadow: 0 8px 40px rgba(0,0,0,0.12); }',
'Soft, diffused shadows'),

('Shadow Hard', 'effect-shadow-hard', 'effect',
ARRAY['bold', 'retro', 'playful'],
'.shadow-hard { box-shadow: 8px 8px 0 var(--color-primary); }',
'Hard-edge retro shadow'),

-- Patterns
('Dot Pattern', 'pattern-dots', 'pattern',
ARRAY['subtle', 'professional', 'minimal'],
'.pattern-dots {
background-image: radial-gradient(circle, var(--color-border) 1px, transparent 1px);
background-size: 20px 20px;
}',
'Subtle dot pattern background'),

('Grid Pattern', 'pattern-grid', 'pattern',
ARRAY['tech', 'modern', 'geometric'],
'.pattern-grid {
background-image: linear-gradient(var(--color-border) 1px, transparent 1px),
linear-gradient(90deg, var(--color-border) 1px, transparent 1px);
background-size: 40px 40px;
}',
'Grid line pattern'),

-- Utilities
('Smooth Scroll', 'utility-smooth-scroll', 'utility',
ARRAY['professional', 'modern'],
'html { scroll-behavior: smooth; }
@media (prefers-reduced-motion: reduce) { html { scroll-behavior: auto; } }',
'Smooth scrolling with accessibility support'),

('Focus Visible', 'utility-focus', 'utility',
ARRAY['accessible', 'professional'],
':focus-visible {
outline: 2px solid var(--color-primary);
outline-offset: 2px;
}
:focus:not(:focus-visible) { outline: none; }',
'Accessible focus states'),

('Selection Style', 'utility-selection', 'utility',
ARRAY['professional', 'branded'],
'::selection {
background: var(--color-primary);
color: var(--color-primary-text);
}',
'Branded text selection')

ON CONFLICT DO NOTHING;

-- ============================================================================
-- JS SNIPPETS - Reusable JavaScript interactions
-- ============================================================================
CREATE TABLE IF NOT EXISTS js_snippets (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
name TEXT NOT NULL,
function TEXT NOT NULL,           -- semantic function: "mobile-nav", "smooth-scroll"
category TEXT NOT NULL,           -- "navigation", "animation", "form", "utility"
semantic_tags TEXT[],
js_content TEXT NOT NULL,
trigger TEXT DEFAULT 'DOMContentLoaded',  -- when to run
dependencies TEXT[],              -- other scripts needed
description TEXT,
created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_js_snippets_function ON js_snippets(function);
CREATE INDEX IF NOT EXISTS idx_js_snippets_category ON js_snippets(category);
CREATE INDEX IF NOT EXISTS idx_js_snippets_tags ON js_snippets USING GIN(semantic_tags);

-- Insert JS snippets
INSERT INTO js_snippets (name, function, category, semantic_tags, trigger, js_content, description) VALUES
-- Navigation
('Mobile Nav Toggle', 'nav-mobile-toggle', 'navigation',
ARRAY['responsive', 'mobile'],
'DOMContentLoaded',
'document.addEventListener("DOMContentLoaded", function() {
const toggle = document.querySelector(".nav-toggle");
const menu = document.querySelector(".site-header__menu");
if (toggle && menu) {
toggle.addEventListener("click", function() {
menu.classList.toggle("is-open");
toggle.setAttribute("aria-expanded", menu.classList.contains("is-open"));
});
}
});',
'Toggle mobile navigation menu'),

('Smooth Scroll Links', 'nav-smooth-scroll', 'navigation',
ARRAY['professional', 'modern'],
'DOMContentLoaded',
'document.querySelectorAll(''a[href^="#"]'').forEach(anchor => {
anchor.addEventListener("click", function(e) {
const target = document.querySelector(this.getAttribute("href"));
if (target) {
e.preventDefault();
target.scrollIntoView({ behavior: "smooth", block: "start" });
}
});
});',
'Smooth scroll to anchor links'),

('Active Nav on Scroll', 'nav-scroll-spy', 'navigation',
ARRAY['professional', 'modern'],
'DOMContentLoaded',
'const sections = document.querySelectorAll("section[id]");
const navLinks = document.querySelectorAll(".site-header__link");

function updateActiveNav() {
const scrollPos = window.scrollY + 100;
sections.forEach(section => {
if (scrollPos >= section.offsetTop && scrollPos < section.offsetTop + section.offsetHeight) {
navLinks.forEach(link => {
link.classList.remove("is-active");
if (link.getAttribute("href") === "#" + section.id) {
link.classList.add("is-active");
}
});
}
});
}

window.addEventListener("scroll", updateActiveNav);
updateActiveNav();',
'Highlight active nav item based on scroll position'),

-- Animations
('Animate on Scroll', 'animation-on-scroll', 'animation',
ARRAY['modern', 'dynamic'],
'DOMContentLoaded',
'const animateOnScroll = () => {
const elements = document.querySelectorAll("[data-animate]");
const observer = new IntersectionObserver((entries) => {
entries.forEach(entry => {
if (entry.isIntersecting) {
entry.target.classList.add("is-visible");
observer.unobserve(entry.target);
}
});
}, { threshold: 0.1 });

elements.forEach(el => observer.observe(el));
};
animateOnScroll();',
'Trigger animations when elements scroll into view'),

('Counter Animation', 'animation-counter', 'animation',
ARRAY['energetic', 'conversion-focused'],
'DOMContentLoaded',
'function animateCounter(el) {
const target = parseInt(el.dataset.target);
const duration = 2000;
const start = performance.now();

function update(now) {
const progress = Math.min((now - start) / duration, 1);
el.textContent = Math.floor(progress * target).toLocaleString();
if (progress < 1) requestAnimationFrame(update);
}
requestAnimationFrame(update);
}

const observer = new IntersectionObserver((entries) => {
entries.forEach(entry => {
if (entry.isIntersecting) {
animateCounter(entry.target);
observer.unobserve(entry.target);
}
});
});

document.querySelectorAll("[data-counter]").forEach(el => observer.observe(el));',
'Animate numbers counting up'),

-- Interactions
('FAQ Accordion', 'interaction-accordion', 'interaction',
ARRAY['professional', 'faq'],
'DOMContentLoaded',
'document.querySelectorAll(".faq-item").forEach(item => {
const summary = item.querySelector("summary");
if (summary) {
summary.addEventListener("click", () => {
// Close others in same group
const parent = item.closest(".faq-list");
if (parent) {
parent.querySelectorAll(".faq-item[open]").forEach(other => {
if (other !== item) other.removeAttribute("open");
});
}
});
}
});',
'Single-open accordion for FAQ sections'),

('Copy to Clipboard', 'interaction-copy', 'interaction',
ARRAY['tech', 'utility'],
'DOMContentLoaded',
'document.querySelectorAll("[data-copy]").forEach(btn => {
btn.addEventListener("click", async () => {
const text = btn.dataset.copy || btn.textContent;
await navigator.clipboard.writeText(text);
btn.classList.add("is-copied");
setTimeout(() => btn.classList.remove("is-copied"), 2000);
});
});',
'Copy text to clipboard with feedback'),

-- Forms
('Form Validation', 'form-validation', 'form',
ARRAY['professional', 'conversion-focused'],
'DOMContentLoaded',
'document.querySelectorAll("form[data-validate]").forEach(form => {
form.addEventListener("submit", function(e) {
let valid = true;
form.querySelectorAll("[required]").forEach(field => {
if (!field.value.trim()) {
valid = false;
field.classList.add("is-invalid");
} else {
field.classList.remove("is-invalid");
}
});
if (!valid) e.preventDefault();
});
});',
'Basic client-side form validation'),

-- Utility
('Lazy Load Images', 'utility-lazy-load', 'utility',
ARRAY['performance', 'modern'],
'DOMContentLoaded',
'if ("IntersectionObserver" in window) {
const imageObserver = new IntersectionObserver((entries) => {
entries.forEach(entry => {
if (entry.isIntersecting) {
const img = entry.target;
img.src = img.dataset.src;
img.classList.remove("lazy");
imageObserver.unobserve(img);
}
});
});
document.querySelectorAll("img.lazy").forEach(img => imageObserver.observe(img));
}',
'Lazy load images for performance'),

('Dark Mode Toggle', 'utility-dark-mode', 'utility',
ARRAY['modern', 'tech'],
'DOMContentLoaded',
'const toggle = document.querySelector("[data-theme-toggle]");
const prefersDark = window.matchMedia("(prefers-color-scheme: dark)");

function setTheme(dark) {
document.documentElement.setAttribute("data-theme", dark ? "dark" : "light");
localStorage.setItem("theme", dark ? "dark" : "light");
}

if (toggle) {
toggle.addEventListener("click", () => {
const isDark = document.documentElement.getAttribute("data-theme") === "dark";
setTheme(!isDark);
});
}

// Init
const saved = localStorage.getItem("theme");
setTheme(saved ? saved === "dark" : prefersDark.matches);',
'Toggle between light and dark themes')

ON CONFLICT DO NOTHING;

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Function to find themes by semantic tags
CREATE OR REPLACE FUNCTION find_themes_by_tags(search_tags TEXT[], min_match INT DEFAULT 2)
RETURNS TABLE(
theme_name TEXT,
display_name TEXT,
match_count INT,
matched_tags TEXT[],
semantic_tags TEXT[]
) AS $$
BEGIN
RETURN QUERY
SELECT
ct.name,
ct.display_name,
(SELECT COUNT(*)::INT FROM unnest(ct.semantic_tags) t WHERE t = ANY(search_tags)) as match_count,
ARRAY(SELECT unnest(ct.semantic_tags) INTERSECT SELECT unnest(search_tags)) as matched_tags,
ct.semantic_tags
FROM css_themes ct
WHERE ct.semantic_tags && search_tags
AND (SELECT COUNT(*) FROM unnest(ct.semantic_tags) t WHERE t = ANY(search_tags)) >= min_match
ORDER BY match_count DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to find CSS snippets by tags
CREATE OR REPLACE FUNCTION find_css_snippets_by_tags(search_tags TEXT[])
RETURNS TABLE(
snippet_name TEXT,
function TEXT,
css_content TEXT,
match_count INT
) AS $$
BEGIN
RETURN QUERY
SELECT
cs.name,
cs.function,
cs.css_content,
(SELECT COUNT(*)::INT FROM unnest(cs.semantic_tags) t WHERE t = ANY(search_tags)) as match_count
FROM css_snippets cs
WHERE cs.semantic_tags && search_tags
ORDER BY match_count DESC;
END;
$$ LANGUAGE plpgsql;

-- Example usage:
-- SELECT * FROM find_themes_by_tags(ARRAY['modern', 'tech', 'conversion-focused']);
-- SELECT * FROM find_css_snippets_by_tags(ARRAY['modern', 'hover']);