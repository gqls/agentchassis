-- SQL_2026-08-02d_seam1_footer_compliance_carrier.sql
-- Seam 1 (HANDOFF_2026-08-02 priority 1): chrome-level carrier for every-page
-- invariants. The shared footer-theme-chrome component gains a
-- {{if .compliance_lines}}-gated block; the value arrives per site through the
-- existing input_schema gap-fill (render_site_components_action.go:756-819 →
-- sourceResolver → config.chrome.compliance_lines → resolveConfigPath, which
-- searches site_specs aspects site_config/identity/design_intent).
-- SECOND consumer of the registered STY-050 mechanism (concept register,
-- styling-render-pipeline: "Per-site chrome config via a gated input_schema
-- field") — mechanism is prior art (GTM on idea.uk, 2026-07-30; analytics_id
-- on head-seo-standard since 2026-05-13); the vocabulary KEY is what is new,
-- registered as STY-051.
--
-- SAFETY, measured before authoring (2026-08-02 night):
--   · footer-theme-chrome is the footer slot of ALL 14 live sites (LANDMINES:690
--     class). The block is gated and inline on the copyright line, its CSS
--     INSIDE the gate, so an unset site renders BYTE-IDENTICALLY — proven by
--     platform/orchestration/actions/footer_compliance_lines_test.go (4 tests,
--     incl. empty-array and wrong-type containment) whose old-template constant
--     is md5-identical to the live row and whose new-template constant is
--     md5-identical to the literal below.
--   · `compliance_lines` collides with nothing: 0 hits fleet-wide in
--     content_components.input_schema, html_template, and current site_specs.
--   · site_config is operator-owned (no pipeline writer rewrites it wholesale;
--     update_site_spec_from_item MERGES per-field, carrying other keys forward).
--   · Chrome is a stored artefact (bugs_open/117): this UPDATE alone changes
--     NOTHING on any site until that site's chrome re-renders — the propagation
--     item at the bottom is what makes lendzy pick it up.
--
-- Apply: kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- (Section C is deliberately NOT part of the transaction — dispatch discipline:
--  no dispatch within ~300s of a chassis pod restart; check first.)

BEGIN;

-- A. Gated compliance block + input_schema on the shared footer component.
--    Drift guard: md5 of the template as read 2026-08-02; refuses if any other
--    session has edited the row since. The DO block turns 0-rows into an abort
--    (a bare UPDATE cannot stop the COMMIT — a-one-off-deletion-is-not-a-class-fix
--    landmine).
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
  input_schema = '{"fields": {"compliance_lines": {"source": "config.chrome.compliance_lines", "type": "array", "description": "Every-page compliance/mission lines: an ARRAY of plain-text strings rendered as a gated block in the footer chrome on every page. Unset/empty sites render byte-identically (proven: footer_compliance_lines_test.go). A non-array value degrades the whole template to the regex fallback renderer - array of strings ONLY. STY-051 (second consumer of the STY-050 mechanism); value lives in site_specs aspect site_config under chrome.compliance_lines."}}}'::jsonb,
  updated_at = now()
WHERE id = 'e6347680-4c7c-448b-8cfc-1cea509159d1' AND is_active
  AND md5(html_template) = '95801d67e0ee29988e0c704f4abe971f';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = 'e6347680-4c7c-448b-8cfc-1cea509159d1'
    AND md5(html_template) = 'eea3fb6911cacc97f56a98ba8d68bba6'
    AND input_schema #>> '{fields,compliance_lines,source}' = 'config.chrome.compliance_lines';
  IF n <> 1 THEN
    RAISE EXCEPTION 'footer-theme-chrome update did not land (drift guard hit, or template bytes differ) — aborting; re-read the live row';
  END IF;
END $$;

-- B. Seed lendzy's every-page lines (mission MISSION_2026-08-02_lendzy_shadow:
--    "Every page's chrome states plainly that the site does not lend, never
--    will, takes no applications, does no lead generation, and is independent
--    of and not affiliated with the Financial Conduct Authority.").
--    lendzy has NO current site_config row (measured), so INSERT, not merge.
INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes)
VALUES (
  '8ff093d5-1f19-453b-9439-a10379bbcd76',
  'site_config',
  '{"chrome": {"compliance_lines": [
     "Lendzy does not lend, never will, takes no applications and does no lead generation.",
     "Lendzy is independent of, and not affiliated with, the Financial Conduct Authority."
   ]}}'::jsonb,
  'operator',
  'claude-session-portfolio-positioning-2026-08-02',
  'Every-page compliance lines (mission: the independence line is load-bearing and must appear on every page). Carried by the chrome-compliance-lines seam: footer-theme-chrome input_schema -> site_specs.site_config.chrome.compliance_lines.'
);

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
  WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect = 'site_config' AND is_current
    AND jsonb_array_length(data #> '{chrome,compliance_lines}') = 2;
  IF n <> 1 THEN
    RAISE EXCEPTION 'lendzy site_config seed did not land — aborting';
  END IF;
END $$;

COMMIT;

-- =========================================================================
-- C. PROPAGATION — run SEPARATELY, after: (1) the transaction above is
--    committed, (2) no chassis pod (re)started in the last ~300s.
--    Shape mirrors today's proven lendzy item (needs_rerender /
--    rerender-pages / refresh_site_components) which re-rendered all three
--    chrome slots at 16:04Z and reassembled the pages.
--
-- INSERT INTO site_work_items
--   (site_id, source, item_type, severity, summary, spec, priority,
--    handler_agent, status, created_by, item_key)
-- VALUES (
--   '8ff093d5-1f19-453b-9439-a10379bbcd76',
--   'portfolio_positioning',
--   'needs_rerender',
--   'medium',
--   'Re-render chrome + reassemble pages: pick up the footer compliance-lines carrier (seam 1)',
--   jsonb_build_object(
--     'reason', 'chrome_compliance_lines_seam',
--     'refresh_site_components', true
--   ),
--   90,
--   'rerender-pages',
--   'triaged',
--   'claude-session-portfolio-positioning-2026-08-02',
--   'lendzy-seam1-compliance-chrome-2026-08-02'
-- );
--
-- VERIFY after the item completes, on the BUILT ARTEFACT (gqls/sites
-- origin/master), never on page_components: every page's footer carries both
-- lines; strip <script>/<style> before any census (handoff command notes).
