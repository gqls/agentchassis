// FILE: platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go
//
// Detection, fix transform and completion-time verifier for item_type
// "hardcoded_section_colors" (bugs_open/021 INSTANCE 2 — first verifier written
// against the widened VerifyTarget contract).
//
// Three pieces that must be read together, and ONE population query and ONE
// remit predicate shared between them:
//
//   - The CHECK sweeps unlocked page_components whose rendered_html carries a
//     hex background and a <style> block, then SPLITS that population by the
//     handler's own transform. The in-remit part becomes the dispatchable
//     hardcoded_section_colors item, counted honestly; the residue becomes a
//     capability_gap (see remit.go). The item stays a site-aggregate — locating
//     the defect needs VerifyTarget.SiteID, which is exactly why this item type
//     was unverifiable under the old spec-only contract (see verifiers.go).
//
//   - ReplaceHardcodedColors is the HANDLER's transform (fix_hardcoded_colors
//     action, agent color-variable-fixer), homed HERE so the handler, the check
//     and the verifier share one predicate. actions imports discovery_checks,
//     never the reverse, so this is the only package all three can reach it from.
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
// CHANGE 2026-07-26 (bugs_open/077): the check partitions. Until now only two of
// the three pieces honoured the remit — the verifier was scoped to the handler
// (34adb171c) while detection still asked the wider question, so on sites where
// the fixer's remit was EMPTY the check filed an item that no handler run could
// ever clear. Measured that day: 33 components across 9 sites matched, and on
// four of them (finetuning.uk 8, gaswholesalers.com 6, ai-agent-orchestration.com
// 4, dartsonline.com 1) the remit was provably zero. Those items parked as
// 'unresolved' — labelled "[unresolved after 2 attempts]", i.e. blaming a handler
// that was never able to succeed. Oldest 2026-04-08.
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
	population, err := hardcodedColourPopulation(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	if len(population) == 0 {
		return &CheckResult{}, nil
	}

	// The split that bugs_open/077 is about. inRemit is what a color-variable-fixer
	// run can actually clear; residue is what it provably cannot, no matter how
	// many times it is dispatched.
	inRemit, residue := PartitionByRemit(population, ReplaceHardcodedColors)

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":            "hardcoded_section_colors",
			"components_found": len(inRemit),
			"population":       len(population),
			"out_of_remit":     len(residue),
		}},
	}

	if len(inRemit) > 0 {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":            "hardcoded_section_colors",
			"components_found": len(inRemit),
			"population":       len(population),
			"out_of_remit":     len(residue),
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "design",
			ItemType: "hardcoded_section_colors",
			Severity: "medium",
			Summary: fmt.Sprintf(
				"Found %d components with hardcoded hex colors the colour fixer can replace with CSS variables",
				len(inRemit)),
			SpecJSON:     string(specJSON),
			Priority:     55,
			HandlerAgent: "color-variable-fixer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "hardcoded_section_colors",
			BatchID:      dctx.BatchID,
		})
	}

	if len(residue) > 0 {
		result.WorkItems = append(result.WorkItems, CapabilityGapItem(dctx, CapabilityGap{
			Check:         "hardcoded_section_colors",
			Pipeline:      "design",
			BuilderNeeded: "color-variable-fixer",
			GapKind:       GapHandlerRemit,
			// CORRECTED 2026-07-26 after measuring the residue rather than inferring it.
			// This used to read "rewrite the hex forms ReplaceHardcodedColors leaves
			// alone: light hexes, 3/4/8-digit forms, inline style attributes" — true as
			// a description and MISLEADING as a brief, because it frames the fix as
			// "match more patterns". On the first site measured (finetuning.uk) the
			// residue was 100% light SURFACE colour — #fff x6, #f8f9fa x5, #e0e0e0 x3,
			// tints #fef2f2/#ecfdf5, one 3-digit #eee — and not one dark fill missed on
			// a technicality. Widening the regexes alone would map every white card to
			// var(--color-primary) and repaint it in the brand colour.
			Capability: "the fixer has a TWO-WORD VOCABULARY: ReplaceHardcodedColors maps everything it " +
				"touches to var(--color-primary)/var(--color-secondary), which is only correct for DARK BRAND " +
				"FILLS — that, not its regexes, is what makes its remit narrow. The residue is light surface " +
				"colour and semantic tints, so widening the patterns without widening the TARGETS would repaint " +
				"white cards in the brand colour. What is wanted is a colour->variable mapping derived from the " +
				"site's own palette: style_collections.color_palette already names background and background_alt, " +
				"and the platform already emits --color-background/--color-surface/--color-surface-alt/" +
				"--color-border/--color-white. Deriving the map from data is also what makes widening safe. " +
				"Open questions, not assumptions: whether a bare #fff is worth variabilising; whether semantic " +
				"tints need palette entries that do not yet exist; whether inline style=\"\" attributes come " +
				"into scope at all.",
			Population:   len(population),
			Residue:      len(residue),
			Examples:     residue,
			CodePointers: hardcodedColourCodePointers,
		}))
	}

	return result, nil
}

// hardcodedColourCodePointers scope the eventual build for whoever picks the gap
// up (the feature designer's spec gate asks for these alongside owner approval).
var hardcodedColourCodePointers = []map[string]string{
	{
		"path": "platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go",
		"why":  "replaceCSSColors holds the three regexes that define the remit; widening happens here",
	},
	{
		"path": "platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors_test.go",
		"why":  "TestReplaceHardcodedColorsRemit pins the CURRENT remit — its false cases are the ones a widening must flip, deliberately",
	},
	{
		"path": "platform/orchestration/actions/fix_harcoded_colours_action.go",
		"why":  "the handler that applies the transform; it must keep calling the shared copy, not fork one",
	},
}

// hardcodedColourPopulationSQL is the DETECTOR's predicate, written once. The
// check and the verifier both run it, so the population they reason about cannot
// drift apart — the failure that made the 2026-07-25 council manufacture an
// objection against two queries that were in fact byte-identical, and the one
// that would matter for real if they ever stopped being.
//
// Includes the locked_at exemption: a human-locked component must not block
// completion, and the check would not re-file it either.
const hardcodedColourPopulationSQL = `
	SELECT p.name, COALESCE(pc.slot_name, ''), COALESCE(pc.rendered_html, '')
	FROM page_components pc
	JOIN pages p ON pc.page_id = p.id
	WHERE p.site_id = $1
	  AND pc.locked_at IS NULL
	  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
	  AND pc.rendered_html LIKE '%<style%'
`

// hardcodedColourPopulation returns every component matching the detector's
// predicate, carrying the rendered_html so the caller can apply the handler's
// remit in Go. Key is "page/slot" — the identity used in verifier Detail lines
// and in capability_gap examples.
func hardcodedColourPopulation(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]RemitCandidate, error) {
	rows, err := db.QueryContext(ctx, hardcodedColourPopulationSQL, siteID)
	if err != nil {
		return nil, fmt.Errorf("hardcoded colour population query failed: %w", err)
	}
	defer rows.Close()

	var out []RemitCandidate
	for rows.Next() {
		var page, slot, html string
		if err := rows.Scan(&page, &slot, &html); err != nil {
			return nil, err
		}
		out = append(out, RemitCandidate{Key: page + "/" + slot, Body: html})
	}
	return out, rows.Err()
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

// VerifyHardcodedSectionColorsResolved re-checks, at completion time, that the
// color-variable-fixer's transform has nothing left to do on this site. The
// item is a site-level aggregate (spec carries only a count), so the target's
// SiteID is the whole location — this was the flagged first verifier for the
// widened VerifyTarget contract.
func VerifyHardcodedSectionColorsResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	if target.SiteID == uuid.Nil {
		return VerifyResult{}, fmt.Errorf("hardcoded_section_colors is site-scoped and the target carries no site_id")
	}

	// The DETECTOR's population, from the same constant the check runs, including
	// the locked_at exemption. The handler-remit filter happens in Go below. Zero
	// rows is unambiguous here, unlike the single-target verifiers' missing-row
	// case (bugs_open/032): an aggregate item's defect is "components matching the
	// sweep exist", so an empty sweep IS the defect gone.
	candidates, err := hardcodedColourPopulation(ctx, db, target.SiteID)
	if err != nil {
		return VerifyResult{}, err
	}

	return hardcodedSectionColoursVerdict(candidates), nil
}

// hardcodedSectionColoursVerdict applies the HANDLER's remit to the detector's
// population: a candidate still counts only if ReplaceHardcodedColors would
// change it. The same PartitionByRemit the check now files against, so a
// completion cannot be judged by a different rule than the one that filed the
// item. Pure — unit tested.
func hardcodedSectionColoursVerdict(candidates []RemitCandidate) VerifyResult {
	inRemit, _ := PartitionByRemit(candidates, ReplaceHardcodedColors)
	if len(inRemit) > 0 {
		return VerifyResult{
			Resolved: false,
			Detail: fmt.Sprintf(
				"%d component(s) still carry colours the fixer's own transform would replace (first: %s)",
				len(inRemit), inRemit[0].Key),
		}
	}
	return VerifyResult{
		Resolved: true,
		Detail: "no unlocked component carries a colour within the fixer's remit " +
			"(out-of-remit hexes — light, 3-digit, inline style attributes — may legitimately remain)",
	}
}
