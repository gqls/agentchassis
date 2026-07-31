// FILE: platform/orchestration/actions/discovery_checks/check_dead_controls.go
//
// Detects DEAD CONTROLS in deployed rendered HTML: anchors styled as
// links/CTAs whose href is a no-op ("#", "#!", javascript:void...) — the
// control renders, the visitor clicks, nothing happens. Guard rail 3 of the
// experience loop (RUNBOOK T2.3). The vonc gauntlet ("Enter the Gauntlet" /
// "Preview Today's Challenge", both href="#", live for weeks) is the proof:
// ClassifyLinkScope files "#" under the anchor scope, so phantom_internal_links
// (correctly) never looked at it, and nothing else did.
//
// Deliberate boundaries:
//   - href="" belongs to phantom_internal_links (empty_internal_href) — not
//     re-flagged here, one owner per class.
//   - <button> with no handler is NOT judged statically (JS binds at runtime);
//     the post-hydration equivalent lives in the Tier-4 browser tier.
//   - Runtime-fill shells (data-runtime-fill) are exempt — their placeholder
//     hrefs are hydrated client-side (provocation-card / lobby-grid). Exemption
//     is per ANCHOR (datahelpers.RuntimeFillSpans), so a component holding a
//     shell alongside static links still has those static links judged.
//   - page_components only, NOT site_components: chrome nav toggles use
//     href="#" + JS legitimately, and chrome has its own fixers. Content
//     surfaces are where a dead CTA is a broken promise.
//
// Routing: needs_human_review with NO handler (first landing is
// detect-and-surface — a dead control can mean "wire the destination",
// "build the missing page", or "remove the mock"; picking a fixer
// automatically would guess).
//
// Registration: automatic via init(). Enable by adding "dead_controls" to a
// discovery agent's checks array AFTER the image carrying this file is live.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&DeadControlsCheck{}) }

type DeadControlsCheck struct{}

func (c *DeadControlsCheck) Name() string { return "dead_controls" }

type deadControlFinding struct {
	PageName    string
	PageID      string
	SlotName    string
	Href        string
	Text        string
	Occurrences int
}

func (c *DeadControlsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, p.id::text, COALESCE(pc.slot_name, ''), pc.rendered_html
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND p.status = 'active'
		  AND pc.build_status = 'deployed'
		  AND pc.rendered_html IS NOT NULL
		  AND pc.rendered_html <> ''
	`, dctx.SiteID)
	// Liveness is the COMPONENT's deploy state (pc.build_status='deployed' +
	// rendered_html), NOT the page-level p.build_status. A page routinely serves
	// its deployed artefact on the CDN while p.build_status has drifted to
	// 'needs_rebuild' (bugs_open/049/052/053: ~34 fleet pages serve 200 as
	// needs_rebuild). Filtering on p.build_status='deployed' therefore blinded
	// this check to exactly those live pages — including the vonc gauntlet, the
	// case named in this file's header: it serves two dead href="#" CTAs while
	// p.build_status='needs_rebuild', so the detector never saw its own proof.
	if err != nil {
		return nil, fmt.Errorf("dead_controls page_components query failed: %w", err)
	}
	defer rows.Close()

	type dcKey struct{ page, pageID, slot, href, text string }
	counts := make(map[dcKey]int)
	for rows.Next() {
		var pageName, pageID, slot, html string
		if err := rows.Scan(&pageName, &pageID, &slot, &html); err != nil {
			continue
		}
		// Exempt the anchors INSIDE a hydrating subtree, not every anchor in a
		// component that happens to contain one (bugs_open/137). Where the marker
		// is on the component's own root — the common case — this returns exactly
		// what the old whole-component skip did.
		for _, a := range datahelpers.DeadControlAnchorsOutsideRuntimeFill(html) {
			text := a.Text
			if len(text) > 80 {
				text = text[:80]
			}
			counts[dcKey{pageName, pageID, slot, a.Href, text}]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dead_controls row scan failed: %w", err)
	}

	if len(counts) == 0 {
		return &CheckResult{}, nil
	}

	findings := make([]deadControlFinding, 0, len(counts))
	for k, n := range counts {
		findings = append(findings, deadControlFinding{
			PageName: k.page, PageID: k.pageID, SlotName: k.slot,
			Href: k.href, Text: k.text, Occurrences: n,
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].PageName != findings[j].PageName {
			return findings[i].PageName < findings[j].PageName
		}
		if findings[i].SlotName != findings[j].SlotName {
			return findings[i].SlotName < findings[j].SlotName
		}
		return findings[i].Text < findings[j].Text
	})

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":    "dead_controls",
			"count":    len(findings),
			"findings": findings,
		}},
	}

	for _, f := range findings {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":       "dead_controls",
			"page_name":   f.PageName,
			"page_id":     f.PageID,
			"slot_name":   f.SlotName,
			"href":        f.Href,
			"link_text":   f.Text,
			"occurrences": f.Occurrences,
			"fix": "An interactive control on a deployed page goes nowhere (no-op href). " +
				"Decide: wire the real destination, build the missing target, or remove " +
				"the mock. A not-yet feature must be absent or labelled coming-soon — " +
				"never simulated.",
		})

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(f.PageID); perr == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:    dctx.SiteID,
			PageID:    pageIDPtr,
			Source:    "discovery",
			Pipeline:  "build",
			ItemType:  "dead_control",
			Severity:  "high",
			Summary:   fmt.Sprintf("Dead control on %s (%s): %q -> %s", f.PageName, f.SlotName, f.Text, f.Href),
			SpecJSON:  string(specJSON),
			Priority:  35,
			Status:    "needs_human_review",
			CreatedBy: dctx.AgentType,
			ItemKey:   fmt.Sprintf("dead_control:%s:%s:%s", f.PageName, f.SlotName, f.Text),
			BatchID:   dctx.BatchID,
		})
	}

	dctx.Logger.Warn("dead_controls: found no-op interactive controls on deployed pages",
		zap.Int("count", len(findings)),
		zap.String("site_id", dctx.SiteID.String()))

	return result, nil
}
