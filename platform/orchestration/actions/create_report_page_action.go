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
	"database/sql"
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
	lockedPreserved := false
	switch {
	case lookupErr == nil:
		var res sql.Result
		res, err = params.DB.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $1, content_data = $2::jsonb,
			    build_status = 'approved', updated_at = NOW()
			WHERE id = $3 AND `+pageComponentAgentWritableSQL(""),
			sectionHTML, string(contentJSON), existingPC)
		if err == nil {
			if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
				// Human-locked instance row: the stored dossier stands and this
				// re-run's fresh copy is discarded (bugs_open/058). Downstream
				// validation gets the artefact that will actually serve.
				lockedPreserved = true
				lockedBy, lockType := "", ""
				if st, lockErr := CheckComponentLock(ctx, params.DB, existingPC, logger); lockErr == nil && st != nil {
					lockedBy, lockType = st.LockedBy, st.LockType
				}
				logger.Warn("create_report_page: instance row is human-locked — stored report kept (bugs_open/058)",
					zap.String("page_component_id", existingPC.String()),
					zap.String("locked_by", lockedBy))
				emitLockBlockedChangeItem(ctx, params.DB, siteID, &pageID, &existingPC,
					pageName, reportComponentFunction, lockedBy, lockType,
					"overwrite", "create_report_page", logger)
				if scanErr := params.DB.QueryRowContext(ctx,
					`SELECT rendered_html FROM page_components WHERE id = $1`,
					existingPC).Scan(&sectionHTML); scanErr != nil {
					logger.Warn("create_report_page: could not load stored locked report",
						zap.Error(scanErr))
				}
			}
		}
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
	result := map[string]interface{}{
		"page_id":       pageID.String(),
		"page_name":     pageName,
		"page_url":      pageURL,
		"request_id":    requestID.String(),
		"site_id":       siteID.String(),
		"domain":        siteDomain,
		"rendered_html": sectionHTML, // for the validate_page_content step
	}
	if lockedPreserved {
		result["locked_preserved"] = true
	}
	return result, nil
}

// ============================================================================
// Deterministic report rendering. All scoring-derived text is escaped; prose
// sections were verified upstream and are inserted as HTML.
//
// NO Go template engine is used here or in report_charts.go — every byte is
// built with strings.Builder and fmt.Fprintf. This is deliberate and load
// bearing: text/template renders a missing field as empty with no error
// (missingkey=zero), which on this page would silently understate a report
// whose entire value is "every number traces or the run fails". String
// building has no such mode — a missing figure is either an explicit
// "Not published by manufacturer" or a type-checked zero we test for. Keep it
// that way; introducing a template here reopens that hole (council 7ed137d1,
// bug_historian: the platform's most recent unpatched root cause).
// ============================================================================

// reportDossierCSS is inlined into the rendered section, and has to be.
//
// rerender_single_page CONCATENATES the stored page_components.rendered_html —
// it does not render templates and it does not collect component stylesheets,
// so a class this renderer emits is styled only if the SITE stylesheet already
// defines it. It does not: checked against robot-hands.com's site_specs on
// 2026-07-25, zero occurrences of any report-* class. Without this block the
// dossier ships to a paying visitor as unstyled text — the same failure as
// bugs_open/027 (content hero unstyled where no style guide defined its
// classes), on a page that IS the deliverable.
//
// Scoped under .report-dossier so it cannot leak into site chrome, and every
// colour rides a site CSS variable with a plain fallback, so the report picks
// up each site's palette without this file knowing any site's colours.
const reportDossierCSS = `<style>
.report-dossier .report-disclaimer{border-left:4px solid var(--color-accent,#b45309);background:var(--color-surface-alt,#fff8ed);padding:1rem 1.25rem;margin:0 0 2rem;font-size:.95rem;line-height:1.6}
.report-dossier h2{margin:2.25rem 0 .75rem;font-size:1.25rem}
.report-dossier .report-request-echo{margin:0 0 2rem}
.report-dossier .report-formulas{margin:0 0 1.5rem}
.report-dossier .criteria-table{width:100%;border-collapse:collapse;margin:0 0 1rem}
.report-dossier .criteria-table th,.report-dossier .criteria-table td{text-align:left;padding:.5rem .75rem;border-bottom:1px solid var(--color-border,#e2e8f0);vertical-align:top}
.report-dossier .criteria-table th{width:40%;font-weight:600;color:var(--color-text-muted,#475569)}
.report-dossier .formula-list{list-style:none;padding:0;margin:0}
.report-dossier .formula-list li{margin:0 0 .5rem}
.report-dossier .formula-list code{display:block;padding:.6rem .8rem;background:var(--color-surface-alt,#f1f5f9);border-radius:4px;font-size:.9rem;overflow-x:auto;white-space:pre-wrap;word-break:break-word}
.report-dossier .report-summary-line{font-size:1.05rem;padding:.9rem 1.1rem;background:var(--color-surface-alt,#f1f5f9);border-radius:4px;margin:1.5rem 0}
.report-dossier .report-chart{margin:1.5rem 0;overflow-x:auto}
.report-dossier .report-chart-omissions{font-size:.9rem;color:var(--color-text-muted,#475569);margin:0 0 1.5rem}
.report-dossier .report-prose{margin:0 0 1.5rem;line-height:1.65}
.report-dossier .report-cards{margin:2rem 0}
.report-dossier .match-card{border:1px solid var(--color-border,#e2e8f0);border-left-width:4px;border-radius:6px;padding:1rem 1.25rem;margin:0 0 1rem}
.report-dossier .match-head{display:flex;flex-wrap:wrap;align-items:baseline;gap:.5rem;margin:0 0 .25rem}
.report-dossier .match-name{font-weight:600;font-size:1.05rem}
.report-dossier .match-maker{font-size:.85rem;color:var(--color-text-muted,#475569);margin:0 0 .75rem}
.report-dossier .verdict-badge{display:inline-block;padding:.15rem .6rem;border-radius:999px;font-size:.75rem;font-weight:600;letter-spacing:.02em}
.report-dossier .verdict-match{border-left-color:#16a34a}
.report-dossier .verdict-match .verdict-badge{background:#dcfce7;color:#166534}
.report-dossier .verdict-marginal{border-left-color:#ca8a04}
.report-dossier .verdict-marginal .verdict-badge{background:#fef9c3;color:#854d0e}
.report-dossier .verdict-no-match{border-left-color:#dc2626}
.report-dossier .verdict-no-match .verdict-badge{background:#fee2e2;color:#991b1b}
.report-dossier .verdict-insufficient-data{border-left-color:#94a3b8}
.report-dossier .verdict-insufficient-data .verdict-badge{background:#e2e8f0;color:#334155}
.report-dossier .crit-flag{font-size:.7rem;font-weight:700;padding:.1rem .4rem;border-radius:3px}
.report-dossier .crit-pass{background:#dcfce7;color:#166534}
.report-dossier .crit-fail{background:#fee2e2;color:#991b1b}
.report-dossier .crit-none{color:var(--color-text-muted,#64748b);font-style:italic}
.report-dossier .conflict-note,.report-dossier .tech-note{font-size:.85rem;color:var(--color-text-muted,#475569);margin:.75rem 0 0;padding-left:.75rem;border-left:2px solid var(--color-border,#e2e8f0)}
.report-dossier .report-provenance{margin:2.5rem 0 0;padding-top:1.25rem;border-top:1px solid var(--color-border,#e2e8f0);font-size:.85rem;color:var(--color-text-muted,#475569)}
.report-dossier .report-provenance a{color:inherit}
@media (max-width:640px){.report-dossier .criteria-table th{width:auto}}
</style>`

func renderReportSection(scoring, prose map[string]interface{}) (string, error) {
	esc := html.EscapeString
	var b strings.Builder

	b.WriteString(`<section class="section report-dossier" data-component="report-dossier"><div class="container">`)
	b.WriteString(reportDossierCSS)
	b.WriteString(`<h1 class="section__title">Gripper Selection &amp; Integration Dossier</h1>`)

	// Proximate disclaimer. Conspicuous and next to the deliverable, not
	// relegated to a footer or a legal page: this report is machine-generated,
	// it reasons only over the figures manufacturers publish, and it is
	// information rather than engineering sign-off. Required by the council's
	// compliance seat for this page type (correlation 7ed137d1).
	b.WriteString(`<p class="report-disclaimer" role="note">` +
		`<strong>Read this first.</strong> This dossier is generated automatically ` +
		`from manufacturer-published specifications, using the physics shown below. ` +
		`It is information to speed up your selection — not engineering advice and not a ` +
		`sign-off. Figures are only as current as the sources cited at the foot of this ` +
		`page, unpublished figures are marked as such and never estimated, and no ` +
		`calculation here can see your actual part, surface finish or duty cycle. ` +
		`Verify every figure against the manufacturer's datasheet and validate on your ` +
		`own parts before you buy.</p>`)

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
	svg, omitted := renderHeadroomChart(cands)
	if svg != "" {
		b.WriteString(`<div class="report-chart">` + svg + `</div>`)
	}
	// Name what the chart could not plot. A missing bar is honest only if the
	// reader can see it is missing and why — otherwise a figure lost upstream
	// is indistinguishable from one the manufacturer never published.
	if len(omitted) > 0 {
		b.WriteString(`<p class="report-chart-omissions">Not plotted, because no ` +
			`comparable capacity figure is published for them: `)
		for i, name := range omitted {
			if i > 0 {
				b.WriteString(`, `)
			}
			b.WriteString(esc(name))
		}
		b.WriteString(`. Their full assessment is in the cards below.</p>`)
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
