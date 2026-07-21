// FILE: platform/orchestration/actions/discovery_checks/check_truncated_component.go
//
// Discovery check: active component templates left TRUNCATED by a cut-off
// generation — an unterminated <script>/<style>/<section>/<div>/<fieldset>.
//
// bugs_open/046. bugs_closed/012 fixed the CAUSE of component truncation (the
// write guard, stop_reason decoding, migration 168's token raise) and restored
// ONE casualty by hand. Nobody swept for the rest: 9 damaged components (8 tool,
// 1 section) across 6 domains were still active, still deployed, serving
// unterminated JavaScript to live visitors — and invisible to every check we
// had. rendered_html carried the same cut as the template, so there was no
// template-vs-render disagreement to notice; Tier-4 acceptance passed them (a
// dead <script> fails no declared layout criterion). This check is the missing
// sweep: it would have caught all 8 tools the day they landed.
//
// SIGNAL — tag imbalance only, deliberately narrower than toolTemplateValid.
// A template with an opening <script> and no matching </script> serves broken
// JavaScript; that imbalance is the exact, unambiguous truncation signature.
// Calibrated against the full live population 2026-07-21: the 5-pair imbalance
// catches EXACTLY the 9 census rows fleet-wide, 0 over-fire. toolTemplateValid
// (plan_sections) additionally rejects a template that does not end on a closed
// tag — correct for a load-time drop against the 19 healthy tools that all end
// cleanly, but as a fleet-wide *sweep* the ends-mid-token heuristic flags 36
// legitimate templates. A discovery item is a human's queue entry, not a silent
// schema drop, so the sweep must be high-precision. Imbalance alone is.
//
// truncationTagPairs MIRRORS actions.balancedPairs (component_write_guard.go),
// the canonical list. It cannot be imported: package actions imports
// discovery_checks (registry wiring), so the reverse would be an import cycle.
// truncationTagPairsMirrorGuard (the test) fails if the two lists drift.
//
// ROUTING — detect-and-surface: needs_human_review, NO handler. The remedy
// varies per casualty and none of it is safe to automate here:
//   - restore from an intact prior version (grip-force had one — cheap, no LLM);
//   - regenerate (7 of 8 have no intact version) — but tool recreation can
//     FABRICATE data (bugs_open/020 shipped an invented practice directory
//     live), so auto-routing this class to a regenerator is unsafe;
//   - remove the component.
// The spec carries intact_version_available / intact_version_number so whoever
// works the item can tell restore from regenerate at a glance. Same choice, and
// same reasons, as check_dead_controls.
//
// DELIVERY — restoring or regenerating the TEMPLATE does not by itself change
// the live page; that needs a re-render, and the re-render delivery pipeline has
// its own open defect (bugs_open/024). This check surfaces the source damage; it
// does not own delivery.
//
// Registration: automatic via init(). Enable by adding "truncated_component" to
// a discovery agent's checks array AFTER the image carrying this file is live.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"
)

func init() {
	Register(&TruncatedComponentCheck{})
	RegisterVerifier("truncated_component", VerifyTruncatedComponentResolved)
}

type TruncatedComponentCheck struct{}

func (c *TruncatedComponentCheck) Name() string { return "truncated_component" }

// truncationTagPairs are the open/close tokens whose balance is checked. MIRRORS
// actions.balancedPairs — keep them identical (truncationTagPairsMirrorGuard
// enforces it). See the file header for why it is mirrored, not imported.
var truncationTagPairs = []struct{ open, close string }{
	{"<script", "</script>"},
	{"<style", "</style>"},
	{"<section", "</section>"},
	{"<div", "</div>"},
	{"<fieldset", "</fieldset>"},
}

// truncationMinLen matches the census floor: templates below it are treated as
// intentional stubs, never truncation casualties (mirrors toolTemplateValid /
// sectionTemplateValid, and the census query's `length >= 100`).
const truncationMinLen = 100

// unterminatedTagPairs returns the open tokens (e.g. "<script") that appear more
// often than their closing token in html — the truncation signature. Empty
// result means every paired tag is balanced. Pure: no DB, directly unit-tested.
//
// Case-folded so <SCRIPT> cannot slip past a lowercase token. Returned in the
// fixed truncationTagPairs order for stable messages.
func unterminatedTagPairs(html string) []string {
	if len(html) < truncationMinLen {
		return nil
	}
	folded := strings.ToLower(html)
	var bad []string
	for _, pair := range truncationTagPairs {
		if strings.Count(folded, pair.open) > strings.Count(folded, pair.close) {
			bad = append(bad, pair.open)
		}
	}
	return bad
}

func (c *TruncatedComponentCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Components used on THIS site's pages. Components are fleet-shared by
	// function, so scope is via the page join (no direct site_id on
	// content_components) — same shape as check_component_template_corrupted.
	// No build_status filter: a component on a not-yet-deployed page is still a
	// truncated source that will serve broken markup when it renders.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT cc.id, cc.name, cc.function, cc.component_level, cc.html_template
		  FROM page_components pc
		  JOIN pages p ON p.id = pc.page_id
		  JOIN content_components cc ON cc.id = pc.component_id
		 WHERE p.site_id = $1
		   AND cc.is_active = true
		   AND cc.html_template IS NOT NULL
		   AND length(cc.html_template) >= $2
		 ORDER BY cc.name
	`, dctx.SiteID, truncationMinLen)
	if err != nil {
		return nil, fmt.Errorf("truncated_component query failed: %w", err)
	}
	defer rows.Close()

	type damaged struct {
		id, name, function, level string
		unterminated              []string
	}
	var found []damaged
	for rows.Next() {
		var id, name, function, level, html string
		if err := rows.Scan(&id, &name, &function, &level, &html); err != nil {
			continue
		}
		if bad := unterminatedTagPairs(html); len(bad) > 0 {
			found = append(found, damaged{id, name, function, level, bad})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("truncated_component row scan failed: %w", err)
	}
	if len(found) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}
	for _, d := range found {
		// Cross-site guard: components are fleet-shared, so skip if ANY site
		// already has an open item for this component_id (the per-site dedup
		// index cannot see other sites' items). Mirrors
		// check_component_template_corrupted.
		var open int
		if err := dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT count(*) FROM site_work_items
			 WHERE item_type = 'truncated_component'
			   AND spec->>'component_id' = $1
			   AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled')
		`, d.id).Scan(&open); err != nil {
			dctx.Logger.Warn("truncated_component: open-item guard query failed",
				zap.String("component", d.name), zap.Error(err))
			continue
		}
		if open > 0 {
			continue
		}

		intactVer, hasIntact := c.newestIntactVersion(dctx.Ctx, dctx.DB, d.id)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":                    "truncated_component",
			"component":                d.name,
			"function":                 d.function,
			"component_level":          d.level,
			"component_id":             d.id,
			"unterminated":             d.unterminated,
			"intact_version_available": hasIntact,
		})

		specMap := map[string]interface{}{
			"check":                    "truncated_component",
			"component":                d.name,
			"function":                 d.function,
			"component_level":          d.level,
			"component_id":             d.id,
			"unterminated":             d.unterminated,
			"intact_version_available": hasIntact,
			"fix": "This component's html_template is a cut-off generation: " +
				strings.Join(d.unterminated, ", ") + " opened but never closed. It " +
				"serves broken markup (unterminated <script> stops the tool's " +
				"JavaScript and swallows the page tail as script text). Remedy: if " +
				"an intact prior version exists, restore it (cheap, no LLM); else " +
				"regenerate the component; else remove it. Do NOT blindly re-run " +
				"tool recreation on a data-backed tool — it can fabricate records " +
				"(bugs_open/020). Restoring the template fixes the source; the live " +
				"page updates only on the next re-render (bugs_open/024).",
		}
		if hasIntact {
			specMap["intact_version_number"] = intactVer
		}
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			continue
		}

		summary := fmt.Sprintf("Truncated component %s (%s) serves unterminated %s",
			d.name, d.level, strings.Join(d.unterminated, ", "))
		if hasIntact {
			summary += fmt.Sprintf(" — intact v%d available to restore", intactVer)
		} else {
			summary += " — no intact prior version (needs regeneration)"
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:    dctx.SiteID,
			Source:    "discovery",
			Pipeline:  "build",
			ItemType:  "truncated_component",
			Severity:  "high",
			Summary:   summary,
			SpecJSON:  string(specJSON),
			Priority:  35,
			Status:    "needs_human_review",
			CreatedBy: dctx.AgentType,
			ItemKey:   fmt.Sprintf("truncated_component:%s", d.id),
			BatchID:   dctx.BatchID,
		})
	}

	sort.Slice(result.WorkItems, func(i, j int) bool {
		return result.WorkItems[i].Summary < result.WorkItems[j].Summary
	})

	if len(result.WorkItems) > 0 {
		dctx.Logger.Warn("truncated_component: found components serving unterminated markup",
			zap.Int("count", len(result.WorkItems)),
			zap.String("site_id", dctx.SiteID.String()))
	}
	return result, nil
}

// VerifyTruncatedComponentResolved re-checks, at completion time, that the
// component named in the item is no longer truncated — re-using the SAME
// predicate the detection used. Resolved when the component is deactivated (no
// longer served) or its current html_template balances every paired tag.
//
// component_id here is content_components.id (what the check writes into the
// spec). A missing row is reported as an error, never Resolved: a content
// component is normally deactivated (is_active=false), not deleted, so an absent
// row is ambiguous — genuine removal or a silent delete — and this platform has
// paid repeatedly for reading that ambiguity as success (bugs_open/012, /032).
// The gate fails OPEN on verifier error (complete_work_item_verification.go), so
// this records an honest unknown rather than a false green.
func VerifyTruncatedComponentResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	componentID, _ := target.Spec["component_id"].(string)
	if componentID == "" {
		return VerifyResult{}, fmt.Errorf("truncated_component spec has no component_id")
	}

	var html sql.NullString
	var isActive sql.NullBool
	err := db.QueryRowContext(ctx,
		`SELECT html_template, is_active FROM content_components WHERE id = $1`, componentID,
	).Scan(&html, &isActive)
	if err == sql.ErrNoRows {
		return VerifyResult{}, fmt.Errorf(
			"cannot verify: component %s no longer exists (removed or silently deleted — indistinguishable here)",
			componentID)
	}
	if err != nil {
		return VerifyResult{}, err
	}

	if isActive.Valid && !isActive.Bool {
		return VerifyResult{Resolved: true, Detail: "component is deactivated — no longer served"}, nil
	}
	if bad := unterminatedTagPairs(html.String); len(bad) > 0 {
		return VerifyResult{Resolved: false,
			Detail: "still truncated: unterminated " + strings.Join(bad, ", ")}, nil
	}
	return VerifyResult{Resolved: true, Detail: "html_template balances every paired tag"}, nil
}

// newestIntactVersion returns the highest component_versions.version_number for
// componentID whose html_template is structurally whole (every paired tag
// balanced), and whether one exists. It lets the surfaced item distinguish a
// cheap restore from a costly regeneration. Uses the same predicate as the
// detection, so "intact" here means exactly "not truncated".
func (c *TruncatedComponentCheck) newestIntactVersion(ctx context.Context, db *sql.DB, componentID string) (int, bool) {
	rows, err := db.QueryContext(ctx, `
		SELECT version_number, html_template
		  FROM component_versions
		 WHERE component_id = $1 AND html_template IS NOT NULL
		 ORDER BY version_number DESC
	`, componentID)
	if err != nil {
		return 0, false
	}
	defer rows.Close()
	for rows.Next() {
		var ver int
		var html string
		if err := rows.Scan(&ver, &html); err != nil {
			continue
		}
		if len(unterminatedTagPairs(html)) == 0 && len(html) >= truncationMinLen {
			return ver, true
		}
	}
	return 0, false
}
