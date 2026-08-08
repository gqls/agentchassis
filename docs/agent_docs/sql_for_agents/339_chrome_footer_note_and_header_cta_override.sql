-- 339_chrome_footer_note_and_header_cta_override.sql
-- "Any oufe rerender should not break the site" (owner directive, 2026-08-08).
--
-- oufe.com carried two hand-patches that lived ONLY in the stored
-- site_components artefact, and the chrome re-render of 2026-07-31 19:21
-- silently reverted BOTH (live on the wire until this file's propagation ran):
--   1. the footer honesty note (fallibility disclosure; the object mig 268's
--      header warned a refresh would delete — it did);
--   2. the header CTA rewrite (FIX_2026-07-26: "Get Started"->/contact.html
--      replaced by "Read the cases"->/cases/index.html, on a site whose brief
--      forbids implying a purchase).
-- A stored artefact is one legitimate rebuild from reset. This file moves both
-- into the template+config path (STY-050 mechanism, worked example
-- SQL_2026-08-02d_seam1_footer_compliance_carrier.sql) so ANY rebuild —
-- including the bugs_open/117 baseline wave — REPRODUCES them.
--
-- SAFETY, measured 2026-08-08 before authoring:
--   · footer-theme-chrome is the footer slot of 16 sites; header-theme-chrome
--     the header slot of 15. Both new blocks are {{if}}-gated on config keys
--     with 0 fleet-wide hits (content_components.input_schema, html_template,
--     site_specs all searched for footer_note / header_cta_url /
--     header_cta_label — the three existing footer_note/footer-note hits are
--     unrelated page components: two tool pages and webdesign.co.uk's own
--     footer, none of them chrome on any site pointing at these components).
--     An unset site renders BYTE-IDENTICALLY — proven by
--     platform/orchestration/actions/chrome_note_and_cta_override_test.go
--     (7 tests: unset / empty-string / partial-config identity, set-site
--     render, coexistence with compliance_lines, override-when-default-absent)
--     whose old-template constants are md5-identical to the live rows and
--     whose new-template constants are md5-identical to the literals below
--     (this file is GENERATED from those constants).
--   · The header override deliberately fires only when BOTH url and label are
--     set, and it bypasses chromeLinks.Allows (which vets the default cta_url)
--     — correct-or-absent is the operator's duty for that key, stated in the
--     schema description.
--   · site_config is operator-owned; update_site_spec_from_item merges
--     per-field, so the in-place jsonb_set below cannot be wholesale-reverted
--     by a pipeline writer.
--   · Chrome is a stored artefact (bugs_open/117): this file alone changes
--     NOTHING served until the site's chrome re-renders. Propagation for oufe
--     is section D (run separately). Other 15/16 sites: no observable change
--     now or at their next rebuild (gated out), except that their next rebuild
--     stamps render_inputs (the 117 wave does this once per site regardless).
--
-- Apply: kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- Then record: ./scripts/migration/run-migrations.sh --record-only <file> --note "..."

BEGIN;

-- A. Footer: gated footer_note band + input_schema field (wrapped shape).
--    Drift guard: md5 of the template as read 2026-08-08; the DO block turns
--    0-rows into an abort (a bare UPDATE cannot stop the COMMIT).
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
    </div>{{if .footer_note}}
    <div class="footer-note">
        <p>{{.footer_note}}</p>
        <style>.footer-note { max-width: 1200px; margin: 2rem auto 0; padding: 1.5rem 2rem; border-top: 1px solid rgba(255,255,255,0.15); } .footer-note p { color: var(--color-footer-text, var(--color-text-muted)); font-size: 0.85rem; margin: 0; }</style>
    </div>{{end}}
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
  input_schema = jsonb_set(input_schema, '{fields,footer_note}',
    '{"type": "text", "source": "config.chrome.footer_note", "description": "Per-site plain-text disclosure band rendered between the footer link columns and the footer-bottom bar, CSS inside the gate. Unset/empty sites render byte-identically (proven: chrome_note_and_cta_override_test.go). Plain text only - rendered inside a <p>. STY-052 (third consumer of the STY-050 mechanism); value lives in site_specs aspect site_config under chrome.footer_note. Exists because stored-artefact-only chrome patches are deleted by every rebuild (oufe honesty note, 2026-07-31)."}'::jsonb),
  updated_at = now()
WHERE id = 'e6347680-4c7c-448b-8cfc-1cea509159d1' AND is_active
  AND md5(html_template) = 'eea3fb6911cacc97f56a98ba8d68bba6';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = 'e6347680-4c7c-448b-8cfc-1cea509159d1'
    AND md5(html_template) = '79f57d24f21b05269013386aab28abd1'
    AND input_schema #>> '{fields,footer_note,source}' = 'config.chrome.footer_note'
    AND input_schema #>> '{fields,compliance_lines,source}' = 'config.chrome.compliance_lines';
  IF n <> 1 THEN
    RAISE EXCEPTION 'footer-theme-chrome update did not land (drift guard hit, template bytes differ, or a schema key was lost) — aborting; re-read the live row';
  END IF;
END $$;

-- B. Header: CTA override branch + input_schema fields (this component's
--    schema is FLAT — no "fields" wrapper; keep its shape).
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
        {{if and .header_cta_url .header_cta_label}}<a href="{{.header_cta_url}}" class="header-cta">{{.header_cta_label}}</a>{{else}}{{if .cta_url}}<a href="{{.cta_url}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>{{end}}{{end}}
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
  input_schema = jsonb_set(jsonb_set(COALESCE(input_schema, '{}'::jsonb),
    '{header_cta_url}',
    '{"type": "url", "source": "config.chrome.header_cta_url", "required": false, "on_missing": "skip_field", "description": "Per-site override of the header CTA target; fires ONLY when chrome.header_cta_label is also set (partial config renders the default byte-identically - proven: chrome_note_and_cta_override_test.go). NOTE: bypasses the chromeLinks.Allows vetting applied to the default cta_url - correct-or-absent is the operator''s duty here. STY-053; value lives in site_specs aspect site_config under chrome.header_cta_url."}'::jsonb),
    '{header_cta_label}',
    '{"type": "text", "source": "config.chrome.header_cta_label", "required": false, "on_missing": "skip_field", "description": "Anchor text for the header CTA override; fires only together with chrome.header_cta_url. STY-053; value lives in site_specs aspect site_config under chrome.header_cta_label."}'::jsonb),
  updated_at = now()
WHERE id = '58fde68f-9190-4e5e-b6a5-ea21cf27a9af' AND is_active
  AND md5(html_template) = '0aae8077d9be27df8fef428b54561396';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = '58fde68f-9190-4e5e-b6a5-ea21cf27a9af'
    AND md5(html_template) = 'de23eed5bf2056999eeb7d906184f746'
    AND input_schema #>> '{header_cta_url,source}' = 'config.chrome.header_cta_url'
    AND input_schema #>> '{header_cta_label,source}' = 'config.chrome.header_cta_label'
    AND input_schema #>> '{gtm_container_id,source}' = 'config.analytics.gtm_container_id';
  IF n <> 1 THEN
    RAISE EXCEPTION 'header-theme-chrome update did not land (drift guard hit, template bytes differ, or a schema key was lost) — aborting; re-read the live row';
  END IF;
END $$;

-- C. oufe's values. In-place per-field merge on the CURRENT site_config row
--    (the house pattern update_site_spec_from_item uses); oufe's existing
--    analytics.gtm_container_id must survive — the verify asserts it.
--    footer_note is the owner-approved wording
--    (oufe DRAFT_disclaimer_for_owner_approval.md §A, deployed 2026-07-26,
--    deleted by the 2026-07-31 chrome rebuild).
UPDATE site_specs SET
  data = jsonb_set(data, '{chrome}', COALESCE(data->'chrome', '{}'::jsonb) || jsonb_build_object(
    'footer_note', 'OUFE publishes educational analysis of financial and legal mechanism. We make mistakes, and some of what is here is assembled with AI assistance that can invent convincing detail. Check anything that matters against the primary source. Nothing here is investment advice or a recommendation.'::text,
    'header_cta_url', '/cases/index.html',
    'header_cta_label', 'Read the cases'
  )),
  updated_at = now()
WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39'
  AND aspect = 'site_config' AND is_current;

DO $$
DECLARE d jsonb;
BEGIN
  SELECT data INTO d FROM site_specs
  WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect = 'site_config' AND is_current;
  IF d IS NULL
     OR d #>> '{chrome,footer_note}' NOT LIKE 'OUFE publishes educational analysis%'
     OR d #>> '{chrome,header_cta_url}' <> '/cases/index.html'
     OR d #>> '{chrome,header_cta_label}' <> 'Read the cases'
     OR d #>> '{analytics,gtm_container_id}' <> 'GTM-PQ3WCTBD' THEN
    RAISE EXCEPTION 'oufe site_config merge did not land, or clobbered a sibling key — aborting';
  END IF;
END $$;

COMMIT;

-- =========================================================================
-- D. PROPAGATION for oufe — run SEPARATELY, after: (1) the transaction above
--    is committed, (2) no chassis pod (re)started in the last ~300s, (3) the
--    open-work-items check on oufe is clean. Shape mirrors the proven lendzy
--    seam-1 item (needs_rerender / rerender-pages / refresh_site_components).
--    NOTE oufe has 4 rebuild_policy='owned' pages (privacy, disclaimer, both
--    tools) which a site-wide rerender SKIPS silently — refresh their chrome
--    individually afterwards, and verify by counting deployed pages on the
--    wire, never by orchestration status.
--
-- INSERT INTO site_work_items
--   (site_id, source, item_type, severity, summary, spec, priority,
--    handler_agent, status, created_by, item_key)
-- VALUES (
--   'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39',
--   'oufe_lane',
--   'needs_rerender',
--   'medium',
--   'Re-render chrome + reassemble pages: pick up the footer_note + header CTA override carriers (mig 339)',
--   jsonb_build_object(
--     'reason', 'chrome_config_carriers_mig_339',
--     'refresh_site_components', true
--   ),
--   90,
--   'rerender-pages',
--   'triaged',
--   'claude-session-oufe-rerender-safety-2026-08-08',
--   'oufe-mig339-chrome-carriers-2026-08-08'
-- );
--
-- VERIFY on the WIRE after the item completes and the owned pages are
-- refreshed (all 9 pages):
--   curl -s https://oufe.com/<page> | grep -c 'OUFE publishes educational analysis'   -> 1
--   curl -s https://oufe.com/<page> | grep -c 'class="header-cta">Read the cases'     -> 1
--   curl -s https://oufe.com/<page> | grep -c 'class="header-cta">Get Started'        -> 0
-- and the 117 stamp: SELECT count(*) FILTER (WHERE render_inputs IS NOT NULL)
--   FROM site_components WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39';
