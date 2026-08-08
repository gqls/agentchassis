-- 339_chrome_footer_note_and_header_cta_override_ROLLBACK.sql
-- HAND-RUN SIDECAR (uppercase suffix: excluded from the runner's --apply).
-- Restores both chrome templates to their pre-339 bytes, removes the two/one
-- schema keys, and removes oufe's three chrome.* config values. Written at the
-- debug_historian seat's round-2 request (council trail 5c18ccaa): the forward
-- file's DO/RAISE guards abort a partial apply, but a post-COMMIT retreat had
-- no documented path. The restore bytes are the SAME pinned constants the
-- forward migration was generated from (chrome_note_and_cta_override_test.go).
-- NOTE rollback alone changes nothing served (bugs_open/117, stored artefact):
-- follow with a needs_rerender item (refresh_site_components: true) as in the
-- forward file's section D, or the sites keep serving the note/CTA from the
-- stored chrome.

BEGIN;

UPDATE content_components SET
  html_template = $tpl$<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand">
            <h3>{{.logo_text}}</h3>
            {{if .tagline}}<p>{{.tagline}}</p>{{end}}
        </div>
        {{if .quick_links_html}}<div class="footer-links">
            <h4>Quick Links</h4>
            <ul>
                {{.quick_links_html}}
            </ul>
        </div>{{end}}
        {{if .services_html}}<div class="footer-services">
            <h4>Explore</h4>
            <ul>
                {{.services_html}}
            </ul>
        </div>{{end}}
        {{if or .email .phone}}<div class="footer-contact">
            <h4>Contact</h4>
            {{if .email}}<p><a href="mailto:{{.email}}">{{.email}}</a></p>{{end}}
            {{if .phone}}<p>{{.phone}}</p>{{end}}
        </div>{{end}}
    </div>
    <div class="footer-bottom">
        <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>{{if .compliance_lines}}
        <div class="footer-compliance">{{range .compliance_lines}}<p>{{.}}</p>{{end}}</div>
        <style>.footer-compliance { margin-top: 0.5rem; } .footer-compliance p { color: var(--color-footer-text, var(--color-text-muted)); font-size: 0.85rem; margin: 0.25rem 0; }</style>{{end}}
        {{if .legal_links}}<div class="footer-legal">
            {{range .legal_links}}<a href="{{.url}}">{{.name}}</a>{{end}}
        </div>{{end}}
    </div>
</footer>
<style>
/* Theme-owned chrome — var()-based gaps only; the layout styles .site-footer,
   .footer-container and .footer-bottom. */
.footer-brand p { color: var(--color-footer-text, var(--color-text-muted)); margin: 0.5rem 0 0; }
.footer-legal { display: flex; gap: 1rem; flex-wrap: wrap; justify-content: center; margin-top: 0.5rem; }
.footer-legal a { color: var(--color-footer-text, var(--color-text-muted)); }
.footer-legal a:hover { color: var(--color-accent); }
</style>$tpl$,
  input_schema = input_schema #- '{fields,footer_note}',
  updated_at = now()
WHERE id = 'e6347680-4c7c-448b-8cfc-1cea509159d1' AND is_active
  AND md5(html_template) = '79f57d24f21b05269013386aab28abd1';

UPDATE content_components SET
  html_template = $tpl${{if .gtm_container_id}}<!-- Google Tag Manager (noscript) -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id={{.gtm_container_id}}"
height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
<!-- End Google Tag Manager (noscript) -->{{end}}
<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">
            {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}
        </a>
        <nav class="main-nav">
            <ul>
                {{if .nav_items_html}}{{.nav_items_html}}{{end}}
            </ul>
        </nav>
        {{if .cta_url}}<a href="{{.cta_url}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>{{end}}
        <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">
            <span></span><span></span><span></span>
        </button>
    </div>
</header>
<style>
/* Theme-owned chrome: every colour is a CSS variable resolved by the site
   stylesheet. Only gaps the layout does not style are covered here. */
.header-cta {
    background: var(--color-cta-bg, var(--color-accent));
    color: var(--color-cta-text, var(--color-primary-text));
    padding: 0.5rem 1.1rem;
    border-radius: var(--radius, 4px);
    text-decoration: none;
    font-weight: 600;
    font-size: 0.9rem;
    white-space: nowrap;
}
.header-cta:hover { filter: brightness(1.1); }
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: var(--color-header-text, var(--color-text));
    margin: 5px 0;
}
@media (max-width: 768px) {
    .main-nav.is-open {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: var(--color-header-bg, var(--color-surface));
        padding: 1rem;
        border-bottom: 1px solid var(--color-border);
    }
    .main-nav.is-open ul { flex-direction: column; }
    .main-nav.is-open a { display: block; padding: 0.75rem 0; }
    .header-cta { display: none; }
}
</style>
<script>
document.addEventListener("DOMContentLoaded", function() {
    var toggle = document.querySelector(".mobile-menu-toggle");
    var nav = document.querySelector(".main-nav");
    if (toggle && nav) {
        toggle.addEventListener("click", function() {
            var open = nav.classList.toggle("is-open");
            toggle.setAttribute("aria-expanded", open ? "true" : "false");
        });
    }
});
</script>$tpl$,
  input_schema = (input_schema - 'header_cta_url') - 'header_cta_label',
  updated_at = now()
WHERE id = '58fde68f-9190-4e5e-b6a5-ea21cf27a9af' AND is_active
  AND md5(html_template) = 'de23eed5bf2056999eeb7d906184f746';

UPDATE site_specs SET
  data = jsonb_set(data, '{chrome}',
    ((data->'chrome') - 'footer_note' - 'header_cta_url' - 'header_cta_label')),
  updated_at = now()
WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39'
  AND aspect = 'site_config' AND is_current
  AND data ? 'chrome';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE (id = 'e6347680-4c7c-448b-8cfc-1cea509159d1' AND md5(html_template) = 'eea3fb6911cacc97f56a98ba8d68bba6'
         AND NOT (input_schema #> '{fields}') ? 'footer_note')
     OR (id = '58fde68f-9190-4e5e-b6a5-ea21cf27a9af' AND md5(html_template) = '0aae8077d9be27df8fef428b54561396'
         AND NOT input_schema ? 'header_cta_url');
  IF n <> 2 THEN
    RAISE EXCEPTION 'rollback did not restore both templates (drift guard hit — a later session may have edited them; re-read the live rows) — aborting';
  END IF;
END $$;

COMMIT;
