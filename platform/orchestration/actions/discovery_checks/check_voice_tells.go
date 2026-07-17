// FILE: platform/orchestration/actions/discovery_checks/check_voice_tells.go
//
// Post-deploy voice audit (SPEC_voice_tells_check §3a): flags copy that reads
// machine-written, per site, against that site's voice_gate config on the
// `voice` spec aspect. Companion to check_unverified_claims — claims answers
// "is it true?", this answers "does it read like a person wrote it?".
//
// Motivating case (2026-07-17): a framework rewrite of the leopardess services
// page produced a staccato triad headline, "Not a demo. Not a proof of
// concept." strawman framing, and 40-word packed sentences — every one of
// which this check's deterministic signals catch. A human caught it that day;
// this check is what catches it the next time.
//
// Opt-in: sites whose `voice` spec has no enabled `voice_gate` block are
// skipped entirely. Long-form pages (blog-post page_type, /guides/ urls) use
// relaxed thresholds — essay rhythm differs from landing copy.
//
// Routing: style is softer than truth — severity is ALWAYS medium, findings
// terminate at human review (item_key 'voice:<page_id>'), and there is no
// automated handler. The optional auto-rewrite mode in the spec (§3c) is a
// later phase, off by default, and NOT implemented here.
//
// Skips locked components (human-pinned) by precedent. Skips site_components
// chrome: nav labels and footer link lists are not prose.
//
// Registration: automatic via init() -> Register(&VoiceTellsCheck{}).
// Enable by adding "voice_tells" to a discovery agent's checks array.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&VoiceTellsCheck{}) }

type VoiceTellsCheck struct{}

func (c *VoiceTellsCheck) Name() string { return "voice_tells" }

func (c *VoiceTellsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	gate, err := loadVoiceGateForCheck(dctx)
	if err != nil {
		return nil, err
	}
	if gate == nil {
		return &CheckResult{}, nil // site not opted in
	}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.page_id::text, p.name, COALESCE(p.page_type,''), p.url,
		       COALESCE(pc.slot_name, ''), pc.rendered_html
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND p.status IN ('active','deployed')
		  AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
		  AND pc.locked_at IS NULL
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("voice_tells page query failed: %w", err)
	}
	defer rows.Close()

	type pageAgg struct {
		pageID, pageName string
		findings         []map[string]interface{}
	}
	byPage := map[string]*pageAgg{}
	var order []string

	for rows.Next() {
		var pageID, pageName, pageType, url, slotName, html string
		if err := rows.Scan(&pageID, &pageName, &pageType, &url, &slotName, &html); err != nil {
			dctx.Logger.Warn("voice_tells: scan error", zap.Error(err))
			continue
		}
		longForm := pageType == "blog-post" || strings.HasPrefix(url, "/guides/") || strings.HasPrefix(url, "/blog/")
		fs := gate.ScanVoice(datahelpers.ExtractAssertionText(html), longForm)
		if len(fs) == 0 {
			continue
		}
		agg, ok := byPage[pageID]
		if !ok {
			agg = &pageAgg{pageID: pageID, pageName: pageName}
			byPage[pageID] = agg
			order = append(order, pageID)
		}
		for _, f := range fs {
			agg.findings = append(agg.findings, map[string]interface{}{
				"check": f.Check, "slot_name": slotName, "matched": f.Matched,
				"reason": f.Reason, "snippet": f.Snippet,
				"value": f.Value, "threshold": f.Threshold, "occurrences": f.Occurrences,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("voice_tells iteration failed: %w", err)
	}
	if len(order) == 0 {
		return &CheckResult{}, nil
	}

	total := 0
	for _, id := range order {
		total += len(byPage[id].findings)
	}
	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check": "voice_tells", "count": total, "pages": len(order),
		}},
	}
	for _, id := range order {
		agg := byPage[id]
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "voice_tells",
			"page_id":   agg.pageID,
			"page_name": agg.pageName,
			"findings":  agg.findings,
			"fix": "Human review. The register fix is the site's approved rewrite path " +
				"(page-build-handler + the stored rewrite prompt in spec.suggestion), " +
				"followed by re-review — never an unreviewed auto-rewrite.",
		})
		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(agg.pageID); perr == nil {
			pageIDPtr = &parsed
		}
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "voice_tells",
			Severity:     "medium", // style is softer than truth — never high
			Summary:      fmt.Sprintf("Voice tells on %s: %d finding(s) read machine-written", agg.pageName, len(agg.findings)),
			SpecJSON:     string(specJSON),
			Priority:     40, // behind claims (25/35): truth outranks register
			HandlerAgent: "", // HITL-terminal in this phase (spec §3c)
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("voice:%s", agg.pageID),
			BatchID:      dctx.BatchID,
		})
	}

	dctx.Logger.Info("voice_tells: findings",
		zap.Int("finding_count", total),
		zap.Int("pages_affected", len(order)),
		zap.String("site_id", dctx.SiteID.String()))
	return result, nil
}

// loadVoiceGateForCheck reads the site's current voice spec and compiles its
// voice_gate. nil,nil = not opted in (mirrors loadEvidenceBaseForCheck).
func loadVoiceGateForCheck(dctx DiscoveryCheckContext) (*datahelpers.VoiceGate, error) {
	var data []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'voice' AND is_current = true
	`, dctx.SiteID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("voice_tells voice spec query failed: %w", err)
	}
	gate, err := datahelpers.ParseVoiceGate(data)
	if err != nil {
		return nil, fmt.Errorf("voice_tells voice_gate parse failed: %w", err)
	}
	return gate, nil // nil when no enabled voice_gate — opt-in contract
}
