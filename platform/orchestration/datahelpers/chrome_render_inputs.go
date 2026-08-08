// FILE: platform/orchestration/datahelpers/chrome_render_inputs.go
//
// The chrome render-inputs fingerprint: ONE SQL expression that names every
// store site chrome is rendered from and digests each one, so "has anything
// chrome renders from changed since this row was stored?" is answerable
// without re-running the render (bugs_open/117).
//
// Two callers, and the split is the reason this lives here rather than beside
// either of them:
//
//   - render_site_components_action.go stamps the expression's value onto
//     site_components.render_inputs in the SAME guarded UPDATE that stores
//     rendered_html — the stamp records what the world looked like when this
//     chrome was rendered.
//   - discovery_checks/check_integrity.go (stale_site_components) recomputes
//     the same expression against the live rows and compares with the stamp.
//     A difference means a re-render would consume different inputs.
//
// discovery_checks cannot import package actions (the dependency runs the
// other way — asserted by site_component_lock_guard_test.go), so a Go-side
// shared implementation is impossible and two hand-maintained SQL copies are
// the drift class this repo keeps paying for (bugs_open/185 is the same story
// for the deployment predicate). Both callers embedding this ONE string is
// what makes stamp and recompute agree by construction.
//
// The fingerprint is deliberately a map of NAMED per-input digests rather
// than one opaque hash, so a mismatch can say WHICH input moved. Comparison
// is whole-object jsonb equality (`render_inputs IS DISTINCT FROM (...)`);
// bumping ChromeRenderInputsVersion therefore forces a deliberate fleet-wide
// restamp (every stored stamp mismatches on the `v` key).
//
// What is covered, and what is not (each boundary measured 2026-08-08, in the
// 117 lane's PLAN D6–D8):
//
//   template     content_components.html_template + input_schema of the row's
//                component_id. Hashing the VALUE, not updated_at, is the point:
//                fix_harcoded_colours_action.go writes html_template with no
//                timestamp, and a timestamp bump that changes nothing is not
//                staleness.
//   identity     sites identity columns (loadSiteDataFull's vocabulary).
//   style        style_collections.color_palette + css_themes.css_content.
//   specs        site_specs aspects site_config / identity / design_intent —
//                exactly the list resolveConfigPath hard-codes
//                (plan_sections_action.go), pinned to it by test. Covers the
//                config.* and site_specs.identity.* schema-fill sources live
//                chrome components declare (GTM container ids, colour schemes,
//                compliance lines).
//   nav          active site_nav_items with group, position and per-item
//                target fetchability (the deployment predicate below). NOT the
//                Go-side degradation branches of GetNavItems — both callers
//                share this same approximation, so it stays self-consistent.
//   services     the footer services column's exact source query
//                (buildServicesHTML), label+url of the capped six.
//   brand_assets active assets with purpose logo / sprite_sheet (the head
//                slot's sprite gate and the identity logo fallback).
//   plan_logo    the current plan's site-scope logo imagery joined to its
//                active asset (loadSiteDataFull's override).
//   year         date_part('year', now()) — the copyright year is a real
//                render input; the cost is one deliberate fleet restamp every
//                1 January.
//
//   NOT covered: the pages-table nav fallback for sites with no nav rows
//   (0 of 19 chrome sites today); the header-CTA fallback ranking over
//   interactive pages/hubs (the harmful direction — a dead CTA — is caught by
//   markStaleChromeLinkSlot, bugs_open/191); site_specs aspects beyond the
//   resolver's three; site_assets.* purposes beyond logo/sprite.
//
// The expression correlates on a site_components row aliased `sc` — callers
// MUST alias the table (`site_components AS sc` / `FROM site_components sc`).
// A bare, un-aliased embedding would let the subqueries' own site_id columns
// (pages, assets, site_specs all have one) capture the reference and the
// expression would compare every site with itself, silently.

package datahelpers

import "fmt"

// ChromeRenderInputsVersion is the fingerprint schema version. Bump it when
// the input set below changes; every stored stamp then mismatches on `v` and
// the fleet restamps through the normal detect→rebuild pipe. That restamp is
// a deliberate cost, not an accident — say so in the commit that bumps it.
const ChromeRenderInputsVersion = 1

// AgentWritableSQLFor is the canonical "may automation write this row?"
// predicate for the lock columns page_components and site_components share
// (locked_at, locked_by, lock_type, lock_expires_at — bugs 058/069). alias is
// the table alias INCLUDING the trailing dot ("sc.") or "" when the statement
// has no alias. It moved here from actions/lock_helpers.go (which now
// delegates) the day discovery_checks needed it — the move
// content_data_envelope_guard.go's header predicted for pure predicates.
func AgentWritableSQLFor(alias string) string {
	return fmt.Sprintf(
		"(%[1]slocked_at IS NULL OR (%[1]slock_type = 'timed' AND %[1]slock_expires_at IS NOT NULL AND %[1]slock_expires_at < NOW()))",
		alias)
}

// ChromeRenderInputsSQL is the fingerprint expression. It yields a jsonb
// value; every aggregate carries an explicit ORDER BY (determinism was
// verified against the live fleet before this shipped — two runs, identical
// bytes, 57 rows). chr(31)/chr(30) are the field/row separators: control
// characters that cannot appear in the values being joined.
var ChromeRenderInputsSQL = fmt.Sprintf(`jsonb_build_object(
  'v', %d,
  'year', date_part('year', now())::int,
  'template', (SELECT md5(cc.html_template || chr(31) || COALESCE(cc.input_schema::text, ''))
               FROM content_components cc WHERE cc.id = sc.component_id),
  'identity', (SELECT md5(concat_ws(chr(31), s.domain, COALESCE(s.name, ''), COALESCE(s.company_name, ''),
                COALESCE(s.tagline, ''), COALESCE(s.email, ''), COALESCE(s.phone, ''),
                COALESCE(s.contact_address, ''), COALESCE(s.logo_text, ''), COALESCE(s.logo_url, '')))
               FROM sites s WHERE s.id = sc.site_id),
  'style', (SELECT md5(concat_ws(chr(31), COALESCE(scol.color_palette::text, ''), COALESCE(ct.css_content, '')))
            FROM sites s
            LEFT JOIN style_collections scol ON s.style_collection_id = scol.id
            LEFT JOIN css_themes ct ON scol.css_theme_id = ct.id
            WHERE s.id = sc.site_id),
  'specs', (SELECT COALESCE(md5(string_agg(ss.aspect || chr(31) || ss.data::text, chr(30) ORDER BY ss.aspect)), '')
            FROM site_specs ss
            WHERE ss.site_id = sc.site_id AND ss.is_current
              AND ss.aspect IN ('site_config', 'identity', 'design_intent')),
  'nav', (SELECT COALESCE(md5(string_agg(concat_ws(chr(31), ng.group_type, ni.position::text, ni.label, ni.url,
              EXISTS (SELECT 1 FROM pages p WHERE p.site_id = ni.site_id AND p.url = ni.url
                      AND p.status NOT IN ('deleted', 'archived')
                      AND NOT (%s))::text
            ), chr(30) ORDER BY ng.position, ni.position, ni.url)), '')
          FROM site_nav_items ni JOIN site_nav_groups ng ON ni.group_id = ng.id
          WHERE ni.site_id = sc.site_id AND ni.status = 'active'),
  'services', (SELECT COALESCE(md5(string_agg(t.row_text, chr(30) ORDER BY t.ord, t.name)), '')
               FROM (SELECT concat_ws(chr(31), COALESCE(p.nav_label, p.title, p.name),
                            COALESCE(p.url, '/' || p.name || '.html')) AS row_text,
                            COALESCE(p.nav_order, 99) AS ord, p.name
                     FROM pages p
                     WHERE p.site_id = sc.site_id
                       AND p.status IN ('deployed', 'active')
                       AND NOT (%s)
                       AND p.name NOT IN ('index', 'about', 'contact', 'privacy', 'terms', 'cookies', '404', 'sitemap', 'faq', 'careers', 'insights', 'blog', 'news')
                       AND p.name <> 'services'
                       AND (p.in_header = true OR p.in_footer = true)
                     ORDER BY COALESCE(p.nav_order, 99), p.name
                     LIMIT 6) t),
  'brand_assets', (SELECT COALESCE(md5(string_agg(a.asset_key || chr(31) || a.purpose, chr(30) ORDER BY a.asset_key, a.purpose)), '')
                   FROM assets a
                   WHERE a.site_id = sc.site_id AND a.status = 'active'
                     AND a.purpose IN ('logo', 'sprite_sheet')),
  'plan_logo', (SELECT COALESCE(md5(string_agg(a.asset_key || chr(31) || a.purpose, chr(30) ORDER BY spi.ordering)), '')
                FROM site_plan_imagery spi
                JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
                JOIN assets a ON a.site_id = sp.site_id AND a.asset_key = spi.key AND a.status = 'active'
                WHERE sp.site_id = sc.site_id AND spi.scope = 'site' AND spi.kind = 'logo')
)`,
	ChromeRenderInputsVersion,
	NeverDeployedPagePredicateFor("p"),
	NeverDeployedPagePredicateFor("p"),
)
