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
			Domain:       "content",
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
		       END as pattern,
		       SUBSTRING(pc.rendered_html FROM '(?i)((?:555[- ]?\d{3}[- ]?\d{4}|\(555\)[^<]{0,20}|[\w.]+@example\.(?:com|org|net)|info@(?:company|business|yourdomain|domain)\.\w+|123\s+(?:Main|First|Business)\s+(?:St|Street|Ave|Road)[^<]{0,30}|\[(?:your|insert|company|phone|email|address)[^\]]*\]|Lorem ipsum[^<]{0,30}))') as matched
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html IS NOT NULL
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
