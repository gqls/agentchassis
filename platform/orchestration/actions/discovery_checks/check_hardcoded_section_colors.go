// FILE: platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go
//
// Detection, fix transform and completion-time verifier for item_type
// "hardcoded_section_colors" (bugs_open/021 INSTANCE 2 — first verifier written
// against the widened VerifyTarget contract).
//
// Three pieces that must be read together:
//
//   - The CHECK counts unlocked page_components whose rendered_html carries a
//     hex background and a <style> block, and files ONE site-aggregate work
//     item. The spec carries only a count — locating the defect needs
//     VerifyTarget.SiteID, which is exactly why this item type was unverifiable
//     under the old spec-only contract (see verifiers.go).
//
//   - ReplaceHardcodedColors is the HANDLER's transform (fix_hardcoded_colors
//     action, agent color-variable-fixer), homed HERE so the handler and the
//     verifier share one predicate. actions imports discovery_checks, never the
//     reverse, so this is the only package both can reach it from.
//
//   - The VERIFIER re-checks at completion time that the handler's transform is
//     at a FIXED POINT over the detector's population. Deliberately NOT the
//     detector's predicate: the detector matches ANY hex background (light,
//     3/4/8-digit, inline style="" attributes included), while the handler only
//     rewrites dark 6-digit hexes and two-colour Ndeg gradients inside <style>
//     blocks. A verifier re-running the detector's predicate would mark
//     correctly-handled items unresolved and strand them in 'failed' — the trap
//     that got the page_rerender verifier written, tested and HELD on
//     2026-07-20 (WRONG_CALLS.md: read the HANDLER's remit, not the detector's
//     predicate). "The transform changes nothing" IS the handler's remit, by
//     construction, so the two cannot drift.
//
// CHANGE: Added pc.locked_at IS NULL to skip locked components.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() {
	Register(&HardcodedSectionColorsCheck{})
	RegisterVerifier("hardcoded_section_colors", VerifyHardcodedSectionColorsResolved)
}

type HardcodedSectionColorsCheck struct{}

func (c *HardcodedSectionColorsCheck) Name() string { return "hardcoded_section_colors" }

func (c *HardcodedSectionColorsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	count, err := countHardcodedColorComponents(dctx)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return &CheckResult{}, nil
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":            "hardcoded_section_colors",
		"components_found": count,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":            "hardcoded_section_colors",
			"components_found": count,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "design",
			ItemType:     "hardcoded_section_colors",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Found %d components with hardcoded hex colors in inline styles instead of CSS variables", count),
			SpecJSON:     string(specJSON),
			Priority:     55,
			HandlerAgent: "color-variable-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "hardcoded_section_colors",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

func countHardcodedColorComponents(dctx DiscoveryCheckContext) (int, error) {
	var count int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
		  AND pc.rendered_html LIKE '%<style%'
	`, dctx.SiteID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("hardcoded color count query failed: %w", err)
	}
	return count, nil
}

// ============================================================================
// The handler's transform (moved from fix_harcoded_colours_action.go)
// ============================================================================

// ReplaceHardcodedColors finds hardcoded hex colors in CSS background
// declarations within <style> blocks and replaces them with CSS variables.
// It is THE transform the fix_hardcoded_colors action applies; the verifier
// below uses it as the remit predicate, so keep it the single copy — do not
// fork a private one back into package actions.
//
// Only operates inside <style>...</style> blocks — does not touch inline
// style="" attributes or HTML content.
//
// Replacement strategy:
//   - background with single dark hex     → var(--color-primary)
//   - background-color with dark hex      → var(--color-primary)
//   - gradient with two dark hex values   → gradient with var(--color-primary), var(--color-secondary)
//   - color: #fff/#ffffff/white in dark   → var(--color-white) (already works but normalises)
func ReplaceHardcodedColors(html string) string {
	// Find <style> blocks and only do replacements within them
	styleBlockRe := regexp.MustCompile(`(?is)(<style[^>]*>)(.*?)(</style>)`)

	return styleBlockRe.ReplaceAllStringFunc(html, func(block string) string {
		matches := styleBlockRe.FindStringSubmatch(block)
		if len(matches) < 4 {
			return block
		}
		openTag := matches[1]
		cssContent := matches[2]
		closeTag := matches[3]

		newCSS := replaceCSSColors(cssContent)
		return openTag + newCSS + closeTag
	})
}

// replaceCSSColors does the actual CSS color replacement within a style block.
func replaceCSSColors(css string) string {
	result := css

	// Pattern: linear-gradient(Xdeg, #hex1, #hex2) → linear-gradient(Xdeg, var(--color-primary), var(--color-secondary))
	gradientRe := regexp.MustCompile(
		`(linear-gradient\s*\(\s*\d+deg\s*,\s*)` + // gradient opening with angle
			`#[0-9a-fA-F]{3,8}` + // first color
			`(\s*,\s*)` + // comma
			`#[0-9a-fA-F]{3,8}` + // second color
			`(\s*\))`, // close paren
	)
	result = gradientRe.ReplaceAllString(result,
		`${1}var(--color-primary)${2}var(--color-secondary)${3}`)

	// Pattern: linear-gradient(rgba(0,0,0,X), rgba(0,0,0,Y)) — overlay gradients, leave alone
	// These are opacity overlays on hero images, not brand colors

	// Pattern: background: #hex (single dark color, not white/light)
	// Only replace dark colors (first digit 0-4 in hex = dark)
	bgSingleRe := regexp.MustCompile(
		`(background\s*:\s*)#([0-4][0-9a-fA-F]{5})(\s*[;}\n])`,
	)
	result = bgSingleRe.ReplaceAllString(result, `${1}var(--color-primary)${3}`)

	// Pattern: background-color: #hex (dark)
	bgColorRe := regexp.MustCompile(
		`(background-color\s*:\s*)#([0-4][0-9a-fA-F]{5})(\s*[;}\n])`,
	)
	result = bgColorRe.ReplaceAllString(result, `${1}var(--color-primary)${3}`)

	return result
}

// ============================================================================
// Completion-time verifier (item_type "hardcoded_section_colors")
// ============================================================================

// hardcodedColourCandidate is one component from the DETECTOR's population,
// carried to the verdict with enough identity for a useful Detail line.
type hardcodedColourCandidate struct {
	Page string
	Slot string
	HTML string
}

// VerifyHardcodedSectionColorsResolved re-checks, at completion time, that the
// color-variable-fixer's transform has nothing left to do on this site. The
// item is a site-level aggregate (spec carries only a count), so the target's
// SiteID is the whole location — this was the flagged first verifier for the
// widened VerifyTarget contract.
func VerifyHardcodedSectionColorsResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	if target.SiteID == uuid.Nil {
		return VerifyResult{}, fmt.Errorf("hardcoded_section_colors is site-scoped and the target carries no site_id")
	}

	// The DETECTOR's population — mirrors countHardcodedColorComponents' WHERE,
	// including the locked_at exemption (a human-locked component must not block
	// completion; the check would not re-file it either). The handler-remit
	// filter happens in Go below. Zero rows is unambiguous here, unlike the
	// single-target verifiers' missing-row case (bugs_open/032): an aggregate
	// item's defect is "components matching the sweep exist", so an empty sweep
	// IS the defect gone.
	rows, err := db.QueryContext(ctx, `
		SELECT p.name, COALESCE(pc.slot_name, ''), COALESCE(pc.rendered_html, '')
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
		  AND pc.rendered_html LIKE '%<style%'
	`, target.SiteID)
	if err != nil {
		return VerifyResult{}, err
	}
	defer rows.Close()

	var candidates []hardcodedColourCandidate
	for rows.Next() {
		var c hardcodedColourCandidate
		if err := rows.Scan(&c.Page, &c.Slot, &c.HTML); err != nil {
			return VerifyResult{}, err
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return VerifyResult{}, err
	}

	return hardcodedSectionColoursVerdict(candidates), nil
}

// hardcodedSectionColoursVerdict applies the HANDLER's remit to the detector's
// population: a candidate still counts only if ReplaceHardcodedColors would
// change it. Pure — unit tested.
func hardcodedSectionColoursVerdict(candidates []hardcodedColourCandidate) VerifyResult {
	remaining := 0
	example := ""
	for _, c := range candidates {
		if ReplaceHardcodedColors(c.HTML) != c.HTML {
			remaining++
			if example == "" {
				example = c.Page + "/" + c.Slot
			}
		}
	}
	if remaining > 0 {
		return VerifyResult{
			Resolved: false,
			Detail: fmt.Sprintf(
				"%d component(s) still carry colours the fixer's own transform would replace (first: %s)",
				remaining, example),
		}
	}
	return VerifyResult{
		Resolved: true,
		Detail: "no unlocked component carries a colour within the fixer's remit " +
			"(out-of-remit hexes — light, 3-digit, inline style attributes — may legitimately remain)",
	}
}
