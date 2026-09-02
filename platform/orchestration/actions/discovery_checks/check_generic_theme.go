package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"strings"
)

func init() { Register(&GenericThemeCheck{}) }

type GenericThemeCheck struct{}

func (c *GenericThemeCheck) Name() string { return "generic_theme" }

func (c *GenericThemeCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	finding, err := findGenericTheme(dctx)
	if err != nil {
		return nil, err
	}
	if finding == nil {
		return &CheckResult{}, nil
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":   "generic_theme",
		"finding": finding,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   "generic_theme",
			"finding": finding,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "generic_theme",
			Severity:     "medium",
			Summary:      "Site using default theme — no industry-specific styling",
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "webdesign-agent",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "generic_theme",
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

type genericThemeFinding struct {
	HasWebdesignSpec bool   `json:"has_webdesign_spec"`
	HasIdentitySpec  bool   `json:"has_identity_spec"`
	CSSLength        int    `json:"css_length"`
	UsesDefaultColor bool   `json:"uses_default_color"`
	Detail           string `json:"detail"`
	// SuppressedByThemeKit records that the palette-taste arms fired but were
	// deliberately not raised because the site sits on an applied theme kit.
	SuppressedByThemeKit bool `json:"suppressed_by_theme_kit,omitempty"`
}

func findGenericTheme(dctx DiscoveryCheckContext) (*genericThemeFinding, error) {
	// No-collection sites are owned by missing_style_collection, which emits
	// the composition pair (needs_composition -> needs_design). Skip them here
	// so generic_theme doesn't also queue a bare webdesign-agent run that would
	// race ahead of composition. generic_theme handles only the "collection
	// exists but the theme looks default/bland" case.
	var hasCollection bool
	if err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT style_collection_id IS NOT NULL FROM sites WHERE id = $1
	`, dctx.SiteID).Scan(&hasCollection); err != nil {
		return nil, err
	}
	if !hasCollection {
		return nil, nil
	}

	// Does this site sit on a theme kit whose palette was actually applied?
	// Used BELOW, after the finding is computed — not as an early return.
	//
	// The suppression is narrow on purpose. Re-dispatching webdesign-agent at
	// a themed site re-rolls its palette (the colour-churn mechanism this
	// file's own history documents), so the "looks default / no webdesign
	// spec" arms must not fire. But the CSSLength == 0 arm is a different
	// fact — the head component has NO CSS at all — and no other discovery
	// check reads head CSS, so suppressing that arm would leave a themed site
	// silently unstyled with nothing able to notice. An early return did
	// exactly that.
	//
	// Keyed on `applied.palette`, not on the row existing: a fill_gaps apply
	// that skipped the palette (site already had its own design_intent)
	// changed no colours, so there is nothing for this check to be deferring
	// to. The join to theme_kits also requires the kit still be active — a
	// deactivated kit must not exempt its sites for ever.
	var themeKitAppliedPalette bool
	if err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT EXISTS (
		         SELECT 1
		           FROM site_specs ss
		           JOIN theme_kits tk ON tk.id::text = ss.data->>'theme_kit_id'
		          WHERE ss.site_id = $1
		            AND ss.aspect = 'theme_kit_adoption'
		            AND ss.is_current = true
		            AND tk.is_active = true
		            AND COALESCE((ss.data->'applied'->>'palette')::boolean, false)
		       )
	`, dctx.SiteID).Scan(&themeKitAppliedPalette); err != nil {
		return nil, err
	}

	finding := &genericThemeFinding{}

	// Check for webdesign spec. Two storage conventions exist:
	//   - site_specs aspect='webdesign' — this check's original contract,
	//     which NO code path has ever written (0 rows fleet-wide, 2026-07-17);
	//   - sites.content_data.color_scheme — what the webdesign-agent's
	//     update_site step (update_site_content, merge=true,
	//     content_field=design_spec.result) actually stores.
	// Testing only the former made this check fire on every themed site
	// every discovery pass, each time dispatching webdesign-agent — whose
	// analyze_design LLM re-rolls the palette, drifting site colours run
	// over run (robot-hands R1, 2026-07-17: four CSS rewrites in a day,
	// one rolled a light background onto a dark site).
	// ONE predicate, not two sequential probes: two independent "does a spec
	// exist" paths in different tables are a standing invitation to diverge
	// (council-gate objection, round 3). The site_specs arm stays so the
	// original contract remains satisfiable if a writer ever appears; it is
	// an OR term here, not a separate query with its own fallback branch.
	// The content_data arm tests the VALUE, not key presence — a crashed or
	// empty run can leave color_scheme as null/{}, and reading that as
	// "spec exists" would suppress the finding for a site with no usable
	// spec: the mirror-image false negative of the bug being fixed.
	var hasWebdesignSpec bool
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT EXISTS (
		         SELECT 1 FROM site_specs
		          WHERE site_id = $1 AND aspect = 'webdesign' AND is_current = true
		       )
		    OR EXISTS (
		         SELECT 1 FROM sites
		          WHERE id = $1
		            AND jsonb_typeof(content_data->'color_scheme') = 'object'
		            AND content_data->'color_scheme' <> '{}'::jsonb
		       )
	`, dctx.SiteID).Scan(&hasWebdesignSpec)
	finding.HasWebdesignSpec = hasWebdesignSpec

	// Check for identity spec (needed by webdesign)
	var identityCount int
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM site_specs 
		WHERE site_id = $1 AND aspect = 'identity' AND is_current = true
	`, dctx.SiteID).Scan(&identityCount)
	finding.HasIdentitySpec = identityCount > 0

	// Check actual CSS for default values
	var cssHTML sql.NullString
	dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT rendered_html FROM site_components
		WHERE site_id = $1 AND slot_name = 'head'
	`, dctx.SiteID).Scan(&cssHTML)

	if cssHTML.Valid {
		finding.CSSLength = len(cssHTML.String)
		// Default fallback primary color from base stylesheet
		if strings.Contains(cssHTML.String, "--color-primary: #2c3e50") ||
			strings.Contains(cssHTML.String, "--color-primary: #333") {
			finding.UsesDefaultColor = true
		}
	}

	// Determine if this is a problem
	isGeneric := false
	suppressedByThemeKit := false
	if !finding.HasWebdesignSpec {
		finding.Detail = "No webdesign spec in site_specs or sites.content_data.color_scheme — agent never produced themed CSS"
		isGeneric = true
		suppressedByThemeKit = themeKitAppliedPalette
	} else if finding.UsesDefaultColor {
		finding.Detail = "CSS uses default fallback colours — webdesign may have had no identity context"
		isGeneric = true
		suppressedByThemeKit = themeKitAppliedPalette
	} else if finding.CSSLength == 0 {
		// NOT suppressible by a theme kit. "This site has no CSS at all" is
		// not a palette-taste judgement the kit already answered, and no other
		// discovery check looks at head CSS — staying silent here would leave
		// a themed site unstyled with nothing able to notice.
		finding.Detail = "Head component has no CSS content"
		isGeneric = true
	}

	if !isGeneric {
		return nil, nil
	}
	if suppressedByThemeKit {
		// Recorded, not merely skipped: a suppression that leaves no trace is
		// indistinguishable from a check that never ran.
		finding.Detail += " — SUPPRESSED: site is on a theme kit whose palette was applied; re-dispatching webdesign-agent would re-roll those colours. Change the kit (apply_theme_kit mode=reapply) or raise this by hand."
		finding.SuppressedByThemeKit = true
		return nil, nil
	}

	return finding, nil
}
