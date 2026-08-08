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
// RETRACTION (2026-08-08): the review-queue sweep can now CLOSE an item of this
// type when a re-scan of the page finds no tells — see revalidateVoiceTells in
// the actions package. That is not the auto-rewrite §3c defers and this file's
// `fix` text forbids: retraction never edits copy, it only stops asserting a
// finding that the current page no longer supports. The scan it re-runs is
// ScanVoiceTells below, shared with this check so the two ends of an item's
// life cannot answer the same question differently.
//
// Skips locked components (human-pinned) by precedent. Skips site_components
// chrome: nav labels and footer link lists are not prose.
//
// Registration: automatic via init() -> Register(&VoiceTellsCheck{}).
// Enable by adding "voice_tells" to a discovery agent's checks array.

package discovery_checks

import (
	"context"
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
	gate, err := LoadVoiceGate(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	if gate == nil {
		return &CheckResult{}, nil // site not opted in
	}

	scans, err := ScanVoiceTells(dctx.Ctx, dctx.DB, dctx.SiteID, "", gate, dctx.Logger)
	if err != nil {
		return nil, err
	}

	// ScanVoiceTells reports every page it examined, including the clean ones, so
	// the revalidator can tell "examined and clean" from "examined nothing". Only
	// the pages that actually tripped the gate become work items — filtering here
	// keeps this check's emitted output exactly what it was before the scan was
	// shared.
	var affected []*VoicePageScan
	total := 0
	for _, s := range scans {
		if len(s.Findings) == 0 {
			continue
		}
		affected = append(affected, s)
		total += len(s.Findings)
	}
	if len(affected) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check": "voice_tells", "count": total, "pages": len(affected),
		}},
	}
	for _, agg := range affected {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "voice_tells",
			"page_id":   agg.PageID,
			"page_name": agg.PageName,
			"findings":  agg.Findings,
			"fix": "Human review. The register fix is the site's approved rewrite path " +
				"(page-build-handler + the stored rewrite prompt in spec.suggestion), " +
				"followed by re-review — never an unreviewed auto-rewrite.",
		})
		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(agg.PageID); perr == nil {
			pageIDPtr = &parsed
		}
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "voice_tells",
			Severity:     "medium", // style is softer than truth — never high
			Summary:      fmt.Sprintf("Voice tells on %s: %d finding(s) read machine-written", agg.PageName, len(agg.Findings)),
			SpecJSON:     string(specJSON),
			Priority:     40, // behind claims (25/35): truth outranks register
			HandlerAgent: "", // HITL-terminal in this phase (spec §3c)
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("voice:%s", agg.PageID),
			BatchID:      dctx.BatchID,
		})
	}

	dctx.Logger.Info("voice_tells: findings",
		zap.Int("finding_count", total),
		zap.Int("pages_affected", len(affected)),
		zap.String("site_id", dctx.SiteID.String()))
	return result, nil
}

// VoicePageScan is one page's voice scan. It carries the findings AND the two
// counts that say how the scan reached them, because the review-queue
// revalidator has to distinguish three states a bare `len(Findings) == 0`
// collapses into one:
//
//   - examined some components, found no tells      -> the prose was fixed
//   - examined NOTHING (page gone, unbuilt, empty)  -> we cannot answer
//   - the only components are human-LOCKED          -> we cannot answer
//
// Only the first licenses closing a live work item. The other two look
// identical from the findings list alone, which is exactly the no-op case that
// reads like a success.
type VoicePageScan struct {
	PageID   string
	PageName string
	Findings []map[string]interface{}
	// ComponentsExamined counts components actually passed to ScanVoice.
	ComponentsExamined int
	// ComponentsSkippedLocked counts human-pinned components, which the emit
	// side has always skipped and this scan still does — counted rather than
	// silently dropped so a page whose only tells sit in a locked component
	// cannot be read as "the prose was fixed".
	ComponentsSkippedLocked int
}

// ScanVoiceTells runs the site's voice gate over its live rendered components
// and returns one VoicePageScan per page examined, INCLUDING pages with no
// findings.
//
// It is shared by the emit side (VoiceTellsCheck.Run, pageIDFilter "") and the
// review-queue revalidator (one page). That sharing is the point: the two ends
// of a voice_tells item's life must not be able to answer "does this copy read
// machine-written?" differently, which is what two copies of this query and
// these thresholds would eventually do. Same precedent as needs_page, whose
// revalidator sits beside the resolver its emit-side guards use.
func ScanVoiceTells(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageIDFilter string, gate *datahelpers.VoiceGate, logger *zap.Logger) ([]*VoicePageScan, error) {
	query := `
		SELECT pc.page_id::text, p.name, COALESCE(p.page_type,''), p.url,
		       COALESCE(pc.slot_name, ''), pc.rendered_html, pc.locked_at
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND p.status IN ('active','deployed')
		  AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''`
	args := []interface{}{siteID}
	if strings.TrimSpace(pageIDFilter) != "" {
		query += ` AND pc.page_id = $2::uuid`
		args = append(args, strings.TrimSpace(pageIDFilter))
	}
	query += ` ORDER BY p.name, pc.position`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("voice_tells page query failed: %w", err)
	}
	defer rows.Close()

	byPage := map[string]*VoicePageScan{}
	var order []*VoicePageScan

	for rows.Next() {
		var pageID, pageName, pageType, url, slotName, html string
		var lockedAt sql.NullTime
		if err := rows.Scan(&pageID, &pageName, &pageType, &url, &slotName, &html, &lockedAt); err != nil {
			logger.Warn("voice_tells: scan error", zap.Error(err))
			continue
		}
		agg, ok := byPage[pageID]
		if !ok {
			agg = &VoicePageScan{PageID: pageID, PageName: pageName}
			byPage[pageID] = agg
			order = append(order, agg)
		}
		if lockedAt.Valid {
			agg.ComponentsSkippedLocked++
			continue
		}
		agg.ComponentsExamined++

		longForm := pageType == "blog-post" || strings.HasPrefix(url, "/guides/") || strings.HasPrefix(url, "/blog/")
		for _, f := range gate.ScanVoice(datahelpers.ExtractAssertionText(html), longForm) {
			agg.Findings = append(agg.Findings, map[string]interface{}{
				"check": f.Check, "slot_name": slotName, "matched": f.Matched,
				"reason": f.Reason, "snippet": f.Snippet,
				"value": f.Value, "threshold": f.Threshold, "occurrences": f.Occurrences,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("voice_tells iteration failed: %w", err)
	}
	return order, nil
}

// LoadVoiceGate reads the site's current voice spec and compiles its
// voice_gate. nil,nil = not opted in (mirrors loadEvidenceBaseForCheck).
//
// Exported for the review-queue revalidator: a site that is not opted in cannot
// have its parked voice items re-judged at all, and that has to be answerable
// from the `actions` package.
func LoadVoiceGate(ctx context.Context, db *sql.DB, siteID uuid.UUID) (*datahelpers.VoiceGate, error) {
	var data []byte
	err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'voice' AND is_current = true
	`, siteID).Scan(&data)
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
