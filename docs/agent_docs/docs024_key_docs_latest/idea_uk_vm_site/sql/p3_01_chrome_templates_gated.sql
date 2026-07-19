-- p3_01_chrome_templates_gated.sql — idea.uk: make the site navigable (bugs_open/018)
--
-- WHY. The chrome renderer (render_site_components_action.go:222-262) hands templates a
-- FIXED vocabulary and never reads the component's input_schema (:530; grep input_schema
-- in that file returns nothing). idea.uk's two per-site fork components declared field
-- names absent from that map (nav_link_1_url…, col1_link1_url…), so every one resolved to
-- "" — and with ZERO {{if}} gates in either template, each empty value rendered as a
-- visible dead href="". 30 of 34 anchors on every page. Diagnosis: /bugs_open/018.
--
-- WHAT THIS DOES. Rewrites both templates against the vocabulary the renderer actually
-- supplies (categories / company_links / legal_links / social_links / cta_url / cta_text /
-- logo_url / company_name / tagline / copyright / show_subscribe) and GATES every anchor,
-- so an unresolvable destination renders nothing rather than a broken link (invariant
-- LNK-005). CSS is preserved verbatim except one line: the footer grid goes from
-- "2fr 1fr 1fr 1fr" to "2fr 1fr 1fr" because the 3 hardcoded link columns become 2
-- data-driven ones. input_schema is rewritten to match the templates, because the
-- component-creator contract (093_component_creator.sql:1185) is that template and schema
-- are two views of the same thing — leaving the old schema would be exactly the drift that
-- caused this.
--
-- SCOPE. These two components are used by ONE site each (idea.uk). Verified fleet-wide:
-- no other domain has any empty-href chrome component. This changes no other site.
--
-- VERIFIED BEFORE APPLYING. Both templates rendered through text/template with
-- Option("missingkey=zero") — the same engine as executeGoTemplate (call_agent.go:1150) —
-- against live-shaped data. Result: 0 empty href/src with full data, 0 with logo_url unset,
-- and 0 when the renderer supplies NOTHING AT ALL (only the literal "/" logo link survives).
--
-- ROLLBACK. Originals saved beside this file:
--   p3_01_BACKUP_site-header_template_pre_2026-07-19.html
--   p3_01_BACKUP_site-footer_template_pre_2026-07-19.html
--
-- NOT A COMPLETE FIX ON ITS OWN. Go changes are inert until an image rolls; this is DB, so
-- it is live at the next render. But the deployed pages have chrome BAKED IN, so the pages
-- must be re-rendered and re-deployed before idea.uk actually looks different.

BEGIN;

-- Fail loudly rather than silently updating 0 rows.
DO $guard$
BEGIN
  IF (SELECT count(*) FROM content_components
       WHERE id IN ('f420f3fa-43a2-4a2f-b2e1-39770d45b494',
                    '4238e467-25a6-4174-bee0-6fce914398c8')) <> 2 THEN
    RAISE EXCEPTION 'expected both idea.uk chrome components to exist';
  END IF;
END
$guard$;

UPDATE content_components SET html_template = $tpl$
<style>
  .site-header-section {
    position: sticky;
    top: 0;
    z-index: 1000;
    background-color: var(--color-header-bg, var(--color-background));
    color: var(--color-header-text, var(--color-text));
    box-shadow: 0 1px 4px rgba(0,0,0,0.08);
    width: 100%;
  }
  .site-header-section .header-inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 64px;
  }
  .site-header-section .header-logo {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    text-decoration: none;
    flex-shrink: 0;
  }
  .site-header-section .header-logo-img {
    height: 36px;
    width: auto;
    display: block;
  }
  .site-header-section .header-logo-name {
    font-size: 1.2rem;
    font-weight: 700;
    color: var(--color-header-text, var(--color-heading));
    letter-spacing: -0.01em;
    white-space: nowrap;
  }
  .site-header-section nav.header-nav {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }
  .site-header-section .nav-link {
    display: inline-flex;
    align-items: center;
    padding: 0.5rem 0.85rem;
    font-size: 0.95rem;
    font-weight: 500;
    color: var(--color-header-text, var(--color-text));
    text-decoration: none;
    border-radius: var(--border-radius, 6px);
    transition: background 0.15s, color 0.15s;
    min-height: 44px;
    white-space: nowrap;
  }
  .site-header-section .nav-link:hover,
  .site-header-section .nav-link:focus-visible {
    background: rgba(0,0,0,0.06);
    outline: none;
  }
  .site-header-section .header-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
  }
  .site-header-section .btn-secondary {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem 1.1rem;
    min-height: 44px;
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--color-header-text, var(--color-text));
    background: transparent;
    border: 1.5px solid currentColor;
    border-radius: var(--border-radius, 6px);
    text-decoration: none;
    cursor: pointer;
    transition: background 0.15s, color 0.15s;
    white-space: nowrap;
  }
  .site-header-section .btn-secondary:hover,
  .site-header-section .btn-secondary:focus-visible {
    background: rgba(0,0,0,0.06);
    outline: none;
  }
  .site-header-section .btn-primary {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem 1.25rem;
    min-height: 44px;
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--color-primary-text, #fff);
    background: var(--color-primary);
    border: none;
    border-radius: var(--border-radius, 6px);
    text-decoration: none;
    cursor: pointer;
    transition: background 0.15s;
    white-space: nowrap;
  }
  .site-header-section .btn-primary:hover,
  .site-header-section .btn-primary:focus-visible {
    background: var(--color-primary-hover, var(--color-primary));
    outline: none;
  }
  .site-header-section .hamburger {
    display: none;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    gap: 5px;
    width: 44px;
    height: 44px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    border-radius: var(--border-radius, 6px);
  }
  .site-header-section .hamburger span {
    display: block;
    width: 22px;
    height: 2px;
    background: var(--color-header-text, var(--color-text));
    border-radius: 2px;
    transition: transform 0.2s, opacity 0.2s;
  }
  .site-header-section .hamburger[aria-expanded="true"] span:nth-child(1) {
    transform: translateY(7px) rotate(45deg);
  }
  .site-header-section .hamburger[aria-expanded="true"] span:nth-child(2) {
    opacity: 0;
  }
  .site-header-section .hamburger[aria-expanded="true"] span:nth-child(3) {
    transform: translateY(-7px) rotate(-45deg);
  }
  .site-header-section .mobile-menu {
    display: none;
    flex-direction: column;
    background: var(--color-header-bg, var(--color-background));
    border-top: 1px solid var(--color-border, rgba(0,0,0,0.1));
    padding: 1rem 2rem 1.5rem;
  }
  .site-header-section .mobile-menu.is-open {
    display: flex;
  }
  .site-header-section .mobile-menu .nav-link {
    padding: 0.65rem 0;
    border-bottom: 1px solid var(--color-border, rgba(0,0,0,0.07));
    border-radius: 0;
  }
  .site-header-section .mobile-menu .nav-link:last-of-type {
    border-bottom: none;
  }
  .site-header-section .mobile-actions {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-top: 1rem;
  }
  .site-header-section .mobile-actions .btn-secondary,
  .site-header-section .mobile-actions .btn-primary {
    width: 100%;
    justify-content: center;
  }
  @media (max-width: 768px) {
    .site-header-section nav.header-nav,
    .site-header-section .header-actions {
      display: none;
    }
    .site-header-section .hamburger {
      display: flex;
    }
  }
</style>
<section class="site-header-section" data-component="site-header">
  <div class="header-inner">
    <a href="/" class="header-logo" aria-label="{{.company_name}} home">
      {{if .logo_url}}<img class="header-logo-img" src="{{.logo_url}}" alt="{{.company_name}} logo" />{{end}}
      <span class="header-logo-name">{{.company_name}}</span>
    </a>
    <nav class="header-nav" aria-label="Main navigation">
      {{range .categories}}<a href="{{.url}}" class="nav-link">{{.label}}</a>
      {{end}}
    </nav>
    <div class="header-actions">
      {{if .cta_url}}<a href="{{.cta_url}}" class="btn-primary">{{if .cta_text}}{{.cta_text}}{{else}}Get started{{end}}</a>{{end}}
    </div>
    <button class="hamburger" aria-expanded="false" aria-controls="mobile-menu" aria-label="Open menu">
      <span></span>
      <span></span>
      <span></span>
    </button>
  </div>
  <div class="mobile-menu" id="mobile-menu" role="navigation" aria-label="Main navigation">
    {{range .categories}}<a href="{{.url}}" class="nav-link">{{.label}}</a>
    {{end}}
    <div class="mobile-actions">
      {{if .cta_url}}<a href="{{.cta_url}}" class="btn-primary">{{if .cta_text}}{{.cta_text}}{{else}}Get started{{end}}</a>{{end}}
    </div>
  </div>
</section>
<script src="/tools/assets/site-header.js"></script>
$tpl$,
input_schema = $sch${
  "fields": {
    "company_name":  {"type": "text",      "source": "renderer", "required": true,
                      "llm_guidance": "Site/brand name. Supplied by render_site_components_action from sites.company_name."},
    "logo_url":      {"type": "image_url", "source": "renderer", "required": false,
                      "llm_guidance": "Logo image URL from sites.logo_url. Gated: no value renders no <img>."},
    "categories":    {"type": "list",      "source": "renderer.nav", "required": false,
                      "llm_guidance": "Primary nav items [{label,url,name,slug}] from site_nav_items via GetNavItems(NavGroupPrimary)."},
    "cta_url":       {"type": "text",      "source": "renderer", "required": false,
                      "llm_guidance": "Header CTA destination, validated against real pages by the renderer. Gated: empty renders no button."},
    "cta_text":      {"type": "text",      "source": "renderer", "required": false,
                      "llm_guidance": "Header CTA label. Falls back to 'Get started' in-template."}
  }
}$sch$,
updated_at = now()
WHERE id = 'f420f3fa-43a2-4a2f-b2e1-39770d45b494';

UPDATE content_components SET html_template = $tpl$
<style>
.site-footer-section {
  --section-text: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 90%, transparent);
  --section-text-muted: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 70%, transparent);
  --section-heading: var(--color-footer-text, var(--color-heading));
  --section-surface: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 5%, transparent);
  --section-border: color-mix(in srgb, var(--color-footer-text, var(--color-text)) 20%, transparent);
  background: var(--color-footer-bg, var(--color-surface));
  color: var(--section-text);
  padding: 4rem 2rem 2rem;
}
.site-footer-section * {
  box-sizing: border-box;
}
.site-footer-section .footer-inner {
  max-width: var(--container-max-width, 1200px);
  margin: 0 auto;
}
.site-footer-section .footer-top {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 3rem;
  padding-bottom: 3rem;
  border-bottom: 1px solid var(--section-border);
}
.site-footer-section .footer-brand .footer-logo {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--section-heading);
  text-decoration: none;
  display: inline-block;
  margin-bottom: 1rem;
}
.site-footer-section .footer-brand .footer-tagline {
  font-size: 0.95rem;
  color: var(--section-text-muted);
  line-height: 1.6;
  margin-bottom: 1.5rem;
  max-width: 280px;
}
.site-footer-section .footer-social {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.site-footer-section .footer-social a {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--section-surface);
  border: 1px solid var(--section-border);
  color: var(--section-text);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 600;
  transition: background 0.2s, border-color 0.2s;
  min-width: 44px;
  min-height: 44px;
}
.site-footer-section .footer-social a:hover {
  background: var(--color-primary, #4f46e5);
  border-color: var(--color-primary, #4f46e5);
}
.site-footer-section .footer-col h4 {
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--section-heading);
  margin: 0 0 1.25rem;
}
.site-footer-section .footer-col ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}
.site-footer-section .footer-col ul li a {
  color: var(--section-text-muted);
  text-decoration: none;
  font-size: 0.9rem;
  transition: color 0.2s;
  display: inline-block;
  min-height: 44px;
  line-height: 44px;
}
.site-footer-section .footer-col ul li a:hover {
  color: var(--section-text);
}
.site-footer-section .footer-newsletter {
  margin-top: 2.5rem;
  padding: 2rem;
  background: var(--section-surface);
  border: 1px solid var(--section-border);
  border-radius: var(--border-radius, 8px);
}
.site-footer-section .footer-newsletter h4 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--section-heading);
  margin: 0 0 0.5rem;
}
.site-footer-section .footer-newsletter p {
  font-size: 0.875rem;
  color: var(--section-text-muted);
  margin: 0 0 1rem;
}
.site-footer-section .newsletter-form {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.site-footer-section .newsletter-form input[type="email"] {
  flex: 1;
  min-width: 200px;
  padding: 0.65rem 1rem;
  border-radius: var(--border-radius, 8px);
  border: 1px solid var(--section-border);
  background: var(--section-surface);
  color: var(--section-text);
  font-size: 0.9rem;
  outline: none;
  min-height: 44px;
}
.site-footer-section .newsletter-form input[type="email"]::placeholder {
  color: var(--section-text-muted);
}
.site-footer-section .newsletter-form input[type="email"]:focus {
  border-color: var(--color-primary, #4f46e5);
}
.site-footer-section .newsletter-form button {
  padding: 0.65rem 1.25rem;
  background: var(--color-primary, #4f46e5);
  color: var(--color-primary-text, #fff);
  border: none;
  border-radius: var(--border-radius, 8px);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  min-height: 44px;
  transition: background 0.2s;
  white-space: nowrap;
}
.site-footer-section .newsletter-form button:hover {
  background: var(--color-primary-hover, #4338ca);
}
.site-footer-section .footer-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
  padding-top: 2rem;
  margin-top: 2rem;
}
.site-footer-section .footer-copyright {
  font-size: 0.85rem;
  color: var(--section-text-muted);
}
.site-footer-section .footer-legal {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
}
.site-footer-section .footer-legal a {
  font-size: 0.85rem;
  color: var(--section-text-muted);
  text-decoration: none;
  transition: color 0.2s;
  min-height: 44px;
  display: inline-flex;
  align-items: center;
}
.site-footer-section .footer-legal a:hover {
  color: var(--section-text);
}
.site-footer-section .newsletter-success {
  display: none;
  font-size: 0.875rem;
  color: var(--section-text);
  margin-top: 0.5rem;
}
@media (max-width: 768px) {
  .site-footer-section .footer-top {
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }
  .site-footer-section .footer-brand {
    grid-column: 1 / -1;
  }
  .site-footer-section .footer-bottom {
    flex-direction: column;
    align-items: flex-start;
  }
  .site-footer-section .footer-newsletter {
    margin-top: 0;
  }
}
@media (max-width: 480px) {
  .site-footer-section .footer-top {
    grid-template-columns: 1fr;
  }
  .site-footer-section .newsletter-form {
    flex-direction: column;
  }
  .site-footer-section .newsletter-form input[type="email"],
  .site-footer-section .newsletter-form button {
    width: 100%;
  }
}
</style>
<section class="site-footer-section" data-component="site-footer">
  <div class="footer-inner">
    <div class="footer-top">
      <div class="footer-brand">
        <a href="/" class="footer-logo">{{.company_name}}</a>
        {{if .tagline}}<p class="footer-tagline">{{.tagline}}</p>{{end}}
        {{if .social_links}}<nav class="footer-social" aria-label="Social links">
          {{range .social_links}}<a href="{{.url}}" rel="noopener noreferrer">{{.name}}</a>
          {{end}}
        </nav>{{end}}
      </div>
      {{if .categories}}<div class="footer-col">
        <h4>Site</h4>
        <ul>
          {{range .categories}}<li><a href="{{.url}}">{{.label}}</a></li>
          {{end}}
        </ul>
      </div>{{end}}
      {{if .company_links}}<div class="footer-col">
        <h4>Company</h4>
        <ul>
          {{range .company_links}}<li><a href="{{.url}}">{{.name}}</a></li>
          {{end}}
        </ul>
      </div>{{end}}
    </div>
    {{if .show_subscribe}}<div class="footer-newsletter">
      <h4>{{.newsletter_title}}</h4>
      <p>{{.newsletter_description}}</p>
      <form class="newsletter-form" aria-label="Newsletter signup" novalidate>
        <input type="email" name="email" placeholder="{{.email_placeholder}}" aria-label="Email address" required />
        <button type="submit">{{.subscribe_text}}</button>
      </form>
      <p class="newsletter-success" role="status" aria-live="polite"></p>
    </div>{{end}}
    <div class="footer-bottom">
      <p class="footer-copyright">{{.copyright}}</p>
      {{if .legal_links}}<nav class="footer-legal" aria-label="Legal">
        {{range .legal_links}}<a href="{{.url}}">{{.name}}</a>
        {{end}}
      </nav>{{end}}
    </div>
  </div>
</section>
<script src="/tools/assets/site-footer.js"></script>
$tpl$,
input_schema = $sch${
  "fields": {
    "company_name":           {"type": "text", "source": "renderer", "required": true},
    "tagline":                {"type": "text", "source": "renderer", "required": false},
    "copyright":              {"type": "text", "source": "renderer", "required": false},
    "categories":             {"type": "list", "source": "renderer.nav", "required": false,
                               "llm_guidance": "Primary nav items for the footer Site column."},
    "company_links":          {"type": "list", "source": "renderer", "required": false,
                               "llm_guidance": "[{name,url}] About/Contact/Careers, built by the renderer from footer nav items."},
    "legal_links":            {"type": "list", "source": "renderer", "required": false,
                               "llm_guidance": "[{name,url}] from the legal nav group — real pages only, never fabricated."},
    "social_links":           {"type": "list", "source": "renderer", "required": false,
                               "llm_guidance": "[{name,url}]. Renderer currently supplies an empty list, so the block is gated out."},
    "show_subscribe":         {"type": "boolean", "source": "renderer", "required": false,
                               "llm_guidance": "Gates the newsletter block. False by default: there is no newsletter backend, so the form would be a dead control (cf. bugs_open/017)."},
    "newsletter_title":       {"type": "text", "source": "renderer", "required": false},
    "newsletter_description": {"type": "text", "source": "renderer", "required": false},
    "email_placeholder":      {"type": "text", "source": "renderer", "required": false},
    "subscribe_text":         {"type": "text", "source": "renderer", "required": false}
  }
}$sch$,
updated_at = now()
WHERE id = '4238e467-25a6-4174-bee0-6fce914398c8';

-- The logo asset exists and serves 200, but sites.logo_url is empty, so the header
-- rendered src="" (now gated, so it would simply render no logo image at all).
UPDATE sites SET logo_url = '/assets/images/logo.jpg', logo_text = 'idea.uk'
WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a' AND coalesce(logo_url,'') = '';

COMMIT;

-- Post-apply checks.
SELECT name,
       length(html_template)                       AS tpl_bytes,
       (length(html_template) - length(replace(html_template,'{{if',''))) / 4 AS gates,
       html_template LIKE '%href=""%'              AS literal_empty_href
FROM content_components
WHERE id IN ('f420f3fa-43a2-4a2f-b2e1-39770d45b494','4238e467-25a6-4174-bee0-6fce914398c8');

SELECT domain, logo_url, logo_text FROM sites WHERE id='1244516d-014d-421c-88c6-090bb1e9552a';
