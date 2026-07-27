// FILE: platform/orchestration/actions/discovery_checks/check_palette_contrast.go
//
// features_open/026 Phase 2b. Asks one question of a site's palette: are the
// foreground/background pairings a stylesheet will emit legible?
//
// WHY THIS EXISTS
// ---------------
// On 2026-07-27 fundamentallyai.com served card headings at 1.21:1 — near-white
// text on white cards — across five pages for three days, plus section eyebrows
// at 1.11:1. ~101 WCAG-AA failures, measured in headless Chromium. Every page's
// build_status said `deployed`; not one of the platform's ~50 discovery checks
// said anything, and the owner found it on his phone.
//
// None of them could have. Every existing check reads an INPUT — a component
// template, a token vocabulary, an href — and asks whether that input looks
// right. The palette was valid. The layout was valid. Every component correctly
// used `--color-heading` and `--color-text-muted`. The defect lived only in the
// COMPOSITION. `check_forced_text_colors` comes closest and still cannot see
// it: it detects dark-on-dark inside a `<style>` block, and every colour here
// arrived through a CSS variable defined in the stylesheet.
//
// WHY IT READS `palettes.colours` AND NOT THE DESIGN INTENT
// --------------------------------------------------------
// This is the load-bearing design decision, and the obvious choice is wrong.
//
// The site's *intent* lives in `site_specs.design_intent.palette.reference_values`
// and it was CORRECT throughout the incident — a coherent dark palette with a
// light primary. Auditing intent would have reported a clean site while the
// live pages were unreadable. The whole defect was intent ≠ artefact.
//
// `palettes.colours`, reached via `site_specs.resolved_composition.palette_id`,
// holds the *composed* palette the renderer actually emitted, including the
// specialised slots derived by fillDarkSchemeSpecialisedSlots. On the morning of
// the incident that row held `card_bg: #ffffff`; it now holds `#132239`. That is
// the vantage point that would have caught it, and it is still DB-only —
// no browser, no HTTP, microseconds.
//
// WHAT IT STILL CANNOT SEE, stated so nobody reads a green result as more than
// it is: a component that hard-codes an ink over a themed fill. That is family 3
// of 026's three, it is invisible to any palette-level check by construction,
// and it produced a real regression the same day — repairing the palette flipped
// `primary` from near-black to light blue and two components hard-coding `#fff`
// over it went from 17:1 to 2.32:1. Only scripts/render_audit.py catches that.
//
// WHY THE FINDING IS A capability_gap AND NOT A DISPATCH
// -----------------------------------------------------
// There is no palette-repair handler, and inventing one would re-create
// bugs_closed/077 — a check that filed items at an agent which had never
// existed, so every one was blocked at claim time. Repainting a brand is an
// authoring decision. The honest output is a roadmap row a human reads, not a
// dispatch to nobody.
package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/colour"
)

func init() { Register(&PaletteContrastCheck{}) }

type PaletteContrastCheck struct{}

func (c *PaletteContrastCheck) Name() string { return "palette_contrast" }

func (c *PaletteContrastCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	slots, paletteName, err := loadComposedPalette(dctx)
	if err != nil {
		return nil, err
	}
	// No resolved composition yet is not a finding. A site that has never had a
	// stylesheet rendered has nothing to be wrong about, and filing against it
	// would put noise in front of the same humans this check needs to be
	// believed by.
	if slots == nil {
		dctx.Logger.Debug("palette_contrast: site has no resolved palette yet, skipping")
		return &CheckResult{}, nil
	}

	failures := colour.AuditPalette(slots)
	if len(failures) == 0 {
		return &CheckResult{}, nil
	}

	findings := make([]map[string]interface{}, 0, len(failures))
	worst := failures[0] // AuditPalette sorts worst-first
	for _, f := range failures {
		findings = append(findings, map[string]interface{}{
			"check":           "palette_contrast",
			"role":            f.Pair.Role,
			"foreground_slot": f.Pair.ForegroundSlot,
			"background_slot": f.Pair.BackgroundSlot,
			"foreground":      f.Pair.Foreground,
			"background":      f.Pair.Background,
			"ratio":           roundTo2(f.Ratio),
			"required":        f.Need,
		})
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":       "palette_contrast",
		"palette":     paletteName,
		"failures":    findings,
		"remediation": "Adjust the palette slots named above, or derive the specialised slots from the scheme. Repainting a brand is an authoring decision, which is why this is a capability_gap and not a dispatch.",
		"cannot_see":  "A component that hard-codes an ink over a themed fill is invisible to this check by construction (features_open/026 family 3). Use scripts/render_audit.py for that.",
	})

	return &CheckResult{
		Findings: findings,
		WorkItems: []WorkItemSpec{{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "capability_gap",
			Severity: severityFor(worst.Ratio),
			Summary: fmt.Sprintf(
				"Palette emits %d unreadable pairing(s); worst is %s at %.2f:1 (needs %.1f:1) — %s on %s",
				len(failures), worst.Pair.Role, worst.Ratio, worst.Need,
				worst.Pair.Foreground, worst.Pair.Background),
			SpecJSON:     string(specJSON),
			Priority:     20,
			HandlerAgent: "", // deliberately none — see the file header
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "palette_contrast",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// severityFor grades by how far below AA the worst pairing sits. Text at 1.2:1
// is not "a bit low", it is invisible, and a check that reports every failure at
// the same severity makes the reader do the triage the check should have done.
func severityFor(ratio float64) string {
	switch {
	case ratio < 2.0:
		return "high" // effectively invisible
	case ratio < 3.0:
		return "medium"
	default:
		return "low" // readable but below AA
	}
}

func roundTo2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

// loadComposedPalette reads the palette the renderer actually emitted, via the
// site's resolved composition. Returns (nil, "", nil) when the site has no
// resolved composition or no palette row — both are "nothing to judge yet"
// rather than errors.
func loadComposedPalette(dctx DiscoveryCheckContext) (colour.PaletteSlots, string, error) {
	var raw []byte
	var name string
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT p.colours, COALESCE(p.name, '')
		  FROM site_specs ss
		  JOIN palettes p
		    ON p.id = (ss.data->>'palette_id')::uuid
		 WHERE ss.site_id = $1
		   AND ss.aspect  = 'resolved_composition'
		   AND ss.is_current
		 LIMIT 1`, dctx.SiteID).Scan(&raw, &name)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("load composed palette: %w", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, "", fmt.Errorf("parse palettes.colours for %q: %w", name, err)
	}

	slots := make(colour.PaletteSlots, len(m))
	for k, v := range m {
		s, ok := v.(string)
		if !ok || s == "" {
			continue
		}
		// Slots holding rgba()/var()/gradients are skipped rather than
		// reported: they are legitimate values this check has no opinion
		// about, and AuditPalette's own missing-slot rules decide whether
		// their absence matters. Silently judging an rgba() on its opaque
		// colour would over-report contrast, which is the one direction a
		// legibility check must never fail in.
		if !strings.HasPrefix(s, "#") {
			continue
		}
		slots[k] = s
	}
	return slots, name, nil
}
