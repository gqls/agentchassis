// Patch to design_actions.go — make the new code's signature observable
// from collected_data. Adds three fields to the result map that are ONLY
// set by the new code: pages_loaded, built_components_total,
// planned_components_total. Their presence/absence in collected_data
// then conclusively tells us whether the new code ran.
//
// Apply by replacing the equivalent section in LoadSiteForDesignAction
// of design_actions.go. Logic unchanged; only the result map gains 3 keys.

// ─────────────────────────────────────────────────────────────────
// Replace the existing "Loud warning when the fallback fires" block
// and everything up to "funcSlice := make([]string, 0, ..." with this:
// ─────────────────────────────────────────────────────────────────

	// Loud warning when the fallback fires. Previously this was a silent
	// no-op which masked the load_site_for_design bug for months — empty
	// allComponents meant every webdesign-agent run got the same 5-item
	// hardcoded list as the "real" component inventory, no matter what the
	// site actually contained. If this fires on an established site (one
	// with page_components rows), something is broken upstream.
	usingFallback := false
	if len(allComponents) == 0 {
		usingFallback = true
		params.Logger.Warn("LoadSiteForDesignAction: NO COMPONENTS FOUND — using hardcoded 5-item fallback list. "+
			"For a site with built pages this indicates page_components is empty or the query is broken.",
			zap.String("site_id", id.String()),
			zap.Bool("queried_pages", includePages),
			zap.Int("pages_loaded", pagesLoaded),
			zap.Int("built_components_total", builtTotal),
			zap.Int("planned_components_total", plannedTotal))
		for _, f := range []string{"hero", "services-grid", "differentiators", "social-proof", "call-to-action"} {
			allComponents[f] = true
		}
	}

	funcSlice := make([]string, 0, len(allComponents))
	for f := range allComponents {
		funcSlice = append(funcSlice, f)
	}
	result["all_component_functions"] = funcSlice

	// Surface the new code's signature into collected_data so external
	// observers (the orchestration_states row) can confirm which code
	// path ran. These fields are absent in the old design_actions.go,
	// so seeing them in collected_data is proof the new code is live.
	result["pages_loaded"]             = pagesLoaded
	result["built_components_total"]   = builtTotal
	result["planned_components_total"] = plannedTotal
	result["used_fallback_components"] = usingFallback
