// FILE: platform/orchestration/actions/discovery_checks/check_placeholder_contact.go
//
// CHANGE: Added pc.locked_at IS NULL to skip locked components.
//
// ---------------------------------------------------------------------------
// 2026-08-02, bugs_open/140 — THIS CHECK DID NOT KNOW THE PLACEHOLDERS OUR OWN
// COMPONENT LIBRARY SHIPPED.
//
// The check is named for this defect and raises work items titled "Fabricated
// contact info on page X". Its pattern set was written from the generic
// placeholder conventions — 555 numbers, example.com, "123 Main St", Lorem
// ipsum, John Doe — and was never reconciled against the literals the platform's
// own `contact-info` component substituted when a site's datum was absent:
//
//	{{if .phone}}…{{else}}+1234567890{{end}}                  <- tel: href
//	{{if .phone_display}}…{{else}}+1 (234) 567-890{{end}}
//	{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}
//
// Measured over every unlocked page_components row fleet-wide, 2026-08-02:
// the nine existing patterns COMBINED matched 1 row; the fabricated hours
// matched 8 and the dummy phone 1, and neither had a pattern. So the detector
// was blind to 8 of the 9 live fabrications, and the 8 it missed came from our
// own library. vetcomparison.uk served `tel:+1234567890` and the invented hours
// on the wire that morning.
//
// Migration 287 removed those fallbacks at source (the component's input_schema
// already declared "on_missing": "skip_field" — the template simply disobeyed
// it). The patterns below are the backstop for a REINTRODUCTION.
//
// WHY A LITERAL ROSTER AND NOT "rendered but absent from content_data".
// That join looks like the roster-free test and it is UNSOUND here. RenderContext
// carries top-level `Email` and `Phone` fields (component_library.go, json tags
// "email"/"phone"), and contextToInterfaceMap derives the scalar half of the
// template contract from those tags — so a component can legitimately render a
// phone that its own content_data does not hold, because the value came from site
// identity. idea.uk is exactly that shape. A content_data join would flag it as
// fabricated. Do not "improve" this check into that test without first proving
// every path a contact datum can arrive by; an over-firing guard gets switched
// off, and then it protects nothing (the calibration lesson in
// component_write_guard.go's header).
//
// COST OF A FALSE POSITIVE, stated: a business that genuinely publishes
// "Monday – Friday, 9am – 6pm" in exactly that form, en dash and all, raises one
// work item for a human to dismiss. That is the same trade the 555 patterns
// already make, and it is cheap. The alternative — matching the SHAPE of business
// hours — would flag every site that states real ones.
//
// THE ROSTER DRIFTS BY CONSTRUCTION, and a script is what stops it: a new
// component with a new invented default is invisible here until someone adds the
// literal. scripts/check_placeholder_fallbacks.py reads the LIVE component
// library and flags any active template whose {{else}} branch asserts a fact
// rather than a label, so the source of new entries is measured rather than
// remembered. Run it after any component seed.
// ---------------------------------------------------------------------------

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&PlaceholderContactCheck{}) }

type PlaceholderContactCheck struct{}

func (c *PlaceholderContactCheck) Name() string { return "placeholder_contact" }

func (c *PlaceholderContactCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	placeholders, err := findPlaceholderContact(dctx)
	if err != nil {
		return nil, err
	}
	if len(placeholders) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":    "placeholder_contact",
			"count":    len(placeholders),
			"findings": placeholders,
		}},
	}

	// Group by page — one work item per affected page
	pageFindings := map[string][]placeholderContactFinding{}
	for _, f := range placeholders {
		pageFindings[f.PageID] = append(pageFindings[f.PageID], f)
	}

	for pageID, findings := range pageFindings {
		patterns := make([]string, len(findings))
		for i, f := range findings {
			patterns[i] = f.Pattern
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "placeholder_contact",
			"page_id":   pageID,
			"page_name": findings[0].PageName,
			"patterns":  patterns,
			"findings":  findings,
		})

		var pageIDPtr *uuid.UUID
		if parsed, err := uuid.Parse(pageID); err == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "placeholder_contact",
			Severity:     "high",
			Summary:      fmt.Sprintf("Fabricated contact info on page %s (%d patterns)", findings[0].PageName, len(findings)),
			SpecJSON:     string(specJSON),
			Priority:     30,
			HandlerAgent: "page-content-writer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("placeholder_contact:%s", pageID),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

type placeholderContactFinding struct {
	PageID      string `json:"page_id"`
	PageName    string `json:"page_name"`
	Position    int    `json:"position"`
	Pattern     string `json:"pattern"`
	MatchedText string `json:"matched_text"`
}

func findPlaceholderContact(dctx DiscoveryCheckContext) ([]placeholderContactFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.page_id, p.name, pc.position,
		       CASE
		           WHEN pc.rendered_html ~* '555[- ]?\d{3}[- ]?\d{4}' THEN 'fake_phone_555'
		           WHEN pc.rendered_html ~* '\(555\)' THEN 'fake_phone_555'
		           WHEN pc.rendered_html ~* '@example\.(com|org|net)' THEN 'example_email'
		           WHEN pc.rendered_html ~* 'info@(company|business|yourdomain|domain)\.' THEN 'generic_email'
		           WHEN pc.rendered_html ~* '123\s+(Main|First|Business)\s+(St|Street|Ave|Road)' THEN 'fake_address'
		           WHEN pc.rendered_html ~* '\[(?:your|insert|company|phone|email|address)' THEN 'bracket_placeholder'
		           WHEN pc.rendered_html ~* 'Lorem ipsum' THEN 'lorem_ipsum'
		           WHEN pc.rendered_html ~* '\+1[- ]?\(0{3}\)' THEN 'fake_phone_000'
		           WHEN pc.rendered_html ~* '(?:John|Jane)\s+(?:Doe|Smith)\s' THEN 'placeholder_name'
		           -- bugs_open/140: the dummies our OWN component library shipped.
		           WHEN pc.rendered_html ~* '\+?1[- ]?\(?234\)?[- ]?567[- ]?890' THEN 'library_dummy_phone'
		           WHEN pc.rendered_html ~* 'Monday\s*[-–—]\s*Friday,?\s*9am\s*[-–—]\s*6pm' THEN 'library_fabricated_hours'
		       END as pattern,
		       SUBSTRING(pc.rendered_html FROM '(?i)((?:555[- ]?\d{3}[- ]?\d{4}|\(555\)[^<]{0,20}|[\w.]+@example\.(?:com|org|net)|info@(?:company|business|yourdomain|domain)\.\w+|123\s+(?:Main|First|Business)\s+(?:St|Street|Ave|Road)[^<]{0,30}|\[(?:your|insert|company|phone|email|address)[^\]]*\]|Lorem ipsum[^<]{0,30}|\+?1[- ]?\(?234\)?[- ]?567[- ]?890|Monday\s*[-–—]\s*Friday,?\s*9am\s*[-–—]\s*6pm))') as matched
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html IS NOT NULL
		  AND pc.locked_at IS NULL
		  AND (
		      pc.rendered_html ~* '555[- ]?\d{3}[- ]?\d{4}'
		      OR pc.rendered_html ~* '\(555\)'
		      OR pc.rendered_html ~* '@example\.(com|org|net)'
		      OR pc.rendered_html ~* 'info@(company|business|yourdomain|domain)\.'
		      OR pc.rendered_html ~* '123\s+(Main|First|Business)\s+(St|Street|Ave|Road)'
		      OR pc.rendered_html ~* '\[(?:your|insert|company|phone|email|address)'
		      OR pc.rendered_html ~* 'Lorem ipsum'
		      OR pc.rendered_html ~* '\+1[- ]?\(0{3}\)'
		      OR pc.rendered_html ~* '(?:John|Jane)\s+(?:Doe|Smith)\s'
		      OR pc.rendered_html ~* '\+?1[- ]?\(?234\)?[- ]?567[- ]?890'
		      OR pc.rendered_html ~* 'Monday\s*[-–—]\s*Friday,?\s*9am\s*[-–—]\s*6pm'
		  )
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("placeholder_contact query failed: %w", err)
	}
	defer rows.Close()

	var findings []placeholderContactFinding
	for rows.Next() {
		var f placeholderContactFinding
		var pattern, matched sql.NullString
		if err := rows.Scan(&f.PageID, &f.PageName, &f.Position, &pattern, &matched); err != nil {
			dctx.Logger.Warn("Failed to scan placeholder contact", zap.Error(err))
			continue
		}
		if pattern.Valid {
			f.Pattern = pattern.String
		}
		if matched.Valid {
			f.MatchedText = matched.String
		}
		findings = append(findings, f)
	}
	return findings, nil
}
