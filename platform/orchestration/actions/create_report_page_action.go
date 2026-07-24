// FILE: platform/orchestration/actions/create_report_page_action.go
//
// CreateReportPageAction composes and persists one gripper-dossier report
// page (gripper dossier pilot — DESIGN_2026-07-24_gripper_dossier_pilot.md
// §3 A3). Adapted from the page-creation half of deploy_tool_action.go with
// deliberate deltas:
//
//   - NO component fork: one shared, site-owned 'report-dossier' template
//     component (seed 205) serves every report; the per-request content
//     lives on the page_components INSTANCE row (rendered_html +
//     content_data) — rerender_single_page concatenates stored
//     rendered_html, it does not render templates, so the final HTML is
//     produced HERE in Go.
//   - NO CTA component (bugs_open/023: the resolver is label-blind);
//     provenance links are plain fixed-label anchors.
//   - Page is invisible to all generic machinery: page_type='report' (no
//     listing generator keys on it), in_header=false + in_footer=false
//     (populate_nav_tables skips it entirely), build_status='pending'
//     (write_build_items targets 'planned'), rebuild_policy='owned'.
//   - The page_components upsert is keyed by (page_id, slot_name) lookup —
//     NEVER by a remembered page_components.id (ids are not stable across
//     re-renders; robot-hands lane landmine).
//
// The report renders exclusively from: score_grippers output (deterministic
// physics over verified published figures) and the verified prose sections
// (already gated by verify_report_prose). A defensive guard refuses any
// prose containing a <script tag — report pages are fully static.
//
// Config / data inputs:
//   - site_id       (required)
//   - request_id    (required) — island UUID; becomes the page slug
//   - scoring_field (config, required) — dotted path to score_grippers output
//   - prose_field   (config, required) — dotted path to the prose object

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CreateReportPageInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id", "request_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_report_page", CreateReportPageInputSpec)
}

const reportComponentFunction = "report-dossier"

func CreateReportPageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "create_report_page"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, CreateReportPageInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	requestID, err := uuid.Parse(inputs.Get("request_id"))
	if err != nil {
		// The request id IS the public slug; a malformed one must fail loudly,
		// never be slugified into something guessable.
		return nil, fmt.Errorf("invalid request_id (must be the island UUID): %w", err)
	}

	config := params.StepConfig.Config
	scoringField, _ := config["scoring_field"].(string)
	proseField, _ := config["prose_field"].(string)
	if scoringField == "" || proseField == "" {
		return nil, fmt.Errorf("create_report_page requires scoring_field and prose_field in step config")
	}
	scoring, ok := datahelpers.ExtractNestedField(params.CollectedData, scoringField).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("scoring_field %q did not resolve to an object", scoringField)
	}
	prose, ok := datahelpers.ExtractNestedField(params.CollectedData, proseField).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("prose_field %q did not resolve to an object", proseField)
	}
	for k, v := range prose {
		if s, ok := v.(string); ok && strings.Contains(strings.ToLower(s), "<script") {
			return nil, fmt.Errorf("prose section %s contains a <script tag — report pages are static", k)
		}
	}

	// --- Shared template component (seed 205) ---
	var componentID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		SELECT id FROM content_components
		WHERE function = $1 AND is_active = true
		ORDER BY created_at DESC LIMIT 1`, reportComponentFunction).Scan(&componentID)
	if err != nil {
		return nil, fmt.Errorf("report-dossier component not found (seed 205 applied?): %w", err)
	}

	var siteDomain string
	if err := params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&siteDomain); err != nil {
		return nil, fmt.Errorf("loading site domain: %w", err)
	}

	sectionHTML, err := renderReportSection(scoring, prose)
	if err != nil {
		return nil, err
	}

	pageName := "report-" + requestID.String()
	pageURL := "/reports/" + requestID.String() + ".html"
	pageTitle := "Gripper Selection & Integration Dossier"
	generatedAt := time.Now().UTC().Format(time.RFC3339)

	contentData := map[string]interface{}{
		"schema":       "report-dossier.v1",
		"request_id":   requestID.String(),
		"scoring":      scoring,
		"prose":        prose,
		"generated_at": generatedAt,
		"generator":    "report-builder/1",
	}
	contentJSON, err := json.Marshal(contentData)
	if err != nil {
		return nil, fmt.Errorf("marshalling report content_data: %w", err)
	}

	// --- Page row (idempotent re-run: ON CONFLICT updates in place) ---
	var pageID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO pages (
			site_id, name, url, title, page_type,
			nav_label, nav_order, in_header, in_footer,
			meta_description, sections, build_status, status, rebuild_policy
		) VALUES (
			$1, $2, $3, $4, 'report',
			NULL, 999, false, false,
			'Application-specific gripper selection dossier.', $5::jsonb,
			'pending', 'active', 'owned'
		)
		ON CONFLICT (site_id, name) DO UPDATE SET
			url = EXCLUDED.url, title = EXCLUDED.title,
			sections = EXCLUDED.sections, updated_at = NOW()
		RETURNING id`,
		siteID, pageName, pageURL, pageTitle,
		fmt.Sprintf(`["%s"]`, reportComponentFunction),
	).Scan(&pageID)
	if err != nil {
		return nil, fmt.Errorf("creating report page: %w", err)
	}

	// --- Instance row, keyed by (page_id, slot_name) lookup ---
	var existingPC uuid.UUID
	lookupErr := params.DB.QueryRowContext(ctx, `
		SELECT id FROM page_components WHERE page_id = $1 AND slot_name = $2`,
		pageID, reportComponentFunction).Scan(&existingPC)
	switch {
	case lookupErr == nil:
		_, err = params.DB.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $1, content_data = $2::jsonb,
			    build_status = 'approved', updated_at = NOW()
			WHERE id = $3`, sectionHTML, string(contentJSON), existingPC)
	default:
		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO page_components
			    (page_id, component_id, position, slot_name,
			     rendered_html, content_data, build_status)
			VALUES ($1, $2, 0, $3, $4, $5::jsonb, 'approved')`,
			pageID, componentID, reportComponentFunction, sectionHTML, string(contentJSON))
	}
	if err != nil {
		return nil, fmt.Errorf("persisting report page_component: %w", err)
	}

	logger.Info("create_report_page: page composed",
		zap.String("page_id", pageID.String()), zap.String("url", pageURL),
		zap.Int("html_bytes", len(sectionHTML)))
	return map[string]interface{}{
		"page_id":       pageID.String(),
		"page_name":     pageName,
		"page_url":      pageURL,
		"request_id":    requestID.String(),
		"site_id":       siteID.String(),
		"domain":        siteDomain,
		"rendered_html": sectionHTML, // for the validate_page_content step
	}, nil
}

// ============================================================================
// Deterministic report rendering. All scoring-derived text is escaped; prose
// sections were verified upstream and are inserted as HTML.
// ============================================================================

func renderReportSection(scoring, prose map[string]interface{}) (string, error) {
	esc := html.EscapeString
	var b strings.Builder

	b.WriteString(`<section class="section report-dossier" data-component="report-dossier"><div class="container">`)
	b.WriteString(`<h1 class="section__title">Gripper Selection &amp; Integration Dossier</h1>`)

	// Request echo panel.
	if req, ok := scoring["request"].(map[string]interface{}); ok {
		b.WriteString(`<div class="report-request-echo"><h2>Your application</h2><table class="criteria-table"><tbody>`)
		for _, row := range [][2]string{
			{"Workpiece mass", numStr(req["mass_kg"]) + " kg"},
			{"Required opening / part size", numStr(req["travel_mm"]) + " mm"},
			{"Acceleration", numStr(req["accel_ms2"]) + " m/s²"},
			{"Safety factor", numStr(req["safety_factor"])},
			{"Surface", strVal(req["material"]) + " (μ " + numStr(req["mu"]) + ")"},
			{"Gripping surfaces", numStr(req["surfaces_n"])},
		} {
			fmt.Fprintf(&b, `<tr><th scope="row">%s</th><td>%s</td></tr>`, esc(row[0]), esc(row[1]))
		}
		if ipMin := numStr(req["ip_min"]); ipMin != "" && ipMin != "0" {
			fmt.Fprintf(&b, `<tr><th scope="row">Required protection</th><td>IP%s</td></tr>`, esc(ipMin))
		}
		b.WriteString(`</tbody></table></div>`)
	}

	// Requirements + printed formulas.
	if reqs, ok := scoring["requirements"].(map[string]interface{}); ok {
		b.WriteString(`<div class="report-formulas"><h2>Requirement, by technology class</h2>`)
		if formulas, ok := reqs["formulas"].([]interface{}); ok {
			b.WriteString(`<ul class="formula-list">`)
			for _, f := range formulas {
				fmt.Fprintf(&b, `<li><code>%s</code></li>`, esc(strVal(f)))
			}
			b.WriteString(`</ul>`)
		} else if formulas, ok := reqs["formulas"].([]string); ok {
			b.WriteString(`<ul class="formula-list">`)
			for _, f := range formulas {
				fmt.Fprintf(&b, `<li><code>%s</code></li>`, esc(f))
			}
			b.WriteString(`</ul>`)
		}
		b.WriteString(`</div>`)
	}

	// Verdict summary sentence (deterministic, from scoring).
	fmt.Fprintf(&b, `<p class="report-summary-line"><strong>%s</strong></p>`, esc(strVal(scoring["summary_sentence"])))

	// Headroom chart from re-decoded assessments (never drawn from prose).
	var cands []assessment
	if raw, err := json.Marshal(scoring["candidates"]); err == nil {
		_ = json.Unmarshal(raw, &cands)
	}
	if len(cands) == 0 {
		return "", fmt.Errorf("scoring output carries no candidates — refusing to render an empty report")
	}
	if svg := renderHeadroomChart(cands); svg != "" {
		b.WriteString(`<div class="report-chart">` + svg + `</div>`)
	}

	// Prose sections (verified upstream by verify_report_prose).
	for _, sec := range []struct{ key, heading string }{
		{"summary_html", "Engineering summary"},
		{"candidates_html", "Candidate assessment"},
		{"integration_html", "Integration notes"},
		{"vendor_questions_html", "Questions to put to the vendor"},
	} {
		body := strVal(prose[sec.key])
		if body == "" {
			return "", fmt.Errorf("prose section %s is empty at render time", sec.key)
		}
		fmt.Fprintf(&b, `<div class="report-prose"><h2>%s</h2>%s</div>`, esc(sec.heading), body)
	}

	// Candidate cards, deterministic from scoring.
	b.WriteString(`<div class="report-cards"><h2>Every candidate, scored</h2>`)
	for _, a := range cands {
		fmt.Fprintf(&b, `<div class="match-card verdict-%s"><div class="match-head"><span class="match-name">%s</span><span class="verdict-badge">%s</span></div>`,
			esc(strings.ToLower(strings.ReplaceAll(a.Verdict, " ", "-"))), esc(a.Name), esc(a.Verdict))
		fmt.Fprintf(&b, `<div class="match-maker">%s · %s</div><table class="criteria-table"><tbody>`, esc(a.Maker), esc(a.TechLabel))
		for _, r := range a.Criteria {
			val := `<span class="crit-none">Not published by manufacturer</span>`
			if r.Text != nil {
				val = esc(*r.Text)
			}
			flag := ""
			switch r.Flag {
			case "pass":
				flag = ` <span class="crit-flag crit-pass">PASS</span>`
			case "fail":
				flag = ` <span class="crit-flag crit-fail">FAIL</span>`
			}
			note := ""
			if r.Note != "" {
				note = ` <span class="crit-none">` + esc(r.Note) + `</span>`
			}
			fmt.Fprintf(&b, `<tr><th scope="row">%s</th><td>%s%s%s</td></tr>`, esc(r.Label), val, flag, note)
		}
		b.WriteString(`</tbody></table>`)
		if a.Conflict != "" {
			fmt.Fprintf(&b, `<p class="conflict-note">%s</p>`, esc(a.Conflict))
		}
		if a.TechNote != "" {
			fmt.Fprintf(&b, `<p class="tech-note">%s</p>`, esc(a.TechNote))
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)

	// Provenance footer: plain fixed-label anchors, no CTA components.
	b.WriteString(`<div class="report-provenance"><h2>Sources</h2><p>Every figure above is either computed from your inputs (formulas shown) or quoted from the manufacturer's published specification. Specs as published; verify with the vendor before purchase.</p><ul>`)
	for _, a := range cands {
		if a.SourceURL == "" {
			continue
		}
		fmt.Fprintf(&b, `<li>%s — <a href="%s" rel="nofollow noopener">manufacturer specification</a> (verified %s)</li>`,
			esc(a.Name), esc(a.SourceURL), esc(a.VerifiedDate))
	}
	b.WriteString(`</ul></div></div></section>`)

	return b.String(), nil
}

func strVal(v interface{}) string {
	s, _ := v.(string)
	return s
}

func numStr(v interface{}) string {
	switch n := v.(type) {
	case float64:
		return trimNum(n)
	case int:
		return fmt.Sprintf("%d", n)
	case string:
		return n
	default:
		return ""
	}
}
