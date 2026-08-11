// FILE: platform/orchestration/actions/resolve_internal_links_action.go
//
// ResolveInternalLinksAction is the core action of the internal-link-resolver
// agent. It augments a page's CTA-bearing sections (the ctaFieldNames set) with
// intent-appropriate internal link destinations resolved from the REAL pages —
// never a hardcoded or fabricated target — writing them into each section's
// resolved_data so the existing render path (render_component's merge_with:
// current_section.resolved_data) picks them up with no render-loop change.
//
// Contract:
//   in : site_id (required), page_type, page_name, and a `sections` config PATH
//        pointing at section_plan.sections_ready
//   out: { "sections_ready": [...augmented...],
//          "unresolved":     [ {section, component, field, slot} ... ] }
// The caller (page-content-writer) iterates the returned sections_ready; an
// unresolved entry feeds the build-time unresolved_cta signal.
//
// v1 rule (deterministic, generic): primary/secondary CTA point at the site's
// top content hubs (page_type='section-index', by nav_order, excluding
// about/contact/legal, skipping the page's own hub). Real, validated, never a
// phantom; absent hub -> field left unset (gated template renders no button) and
// reported unresolved. The agent boundary lets this be upgraded (LLM intent-
// matching: a guide -> its related tool) without changing callers — page_type is
// carried for that future use.
//
// Field names differ by component — see ctaFieldNames for the covered set
// and each component's primary/secondary url field names.
//
// Input handling follows 003/001 contracts: site_id/page_type/page_name are
// scalars via ExtractActionInputs; `sections` is a complex (array) value whose
// config holds a PATH, so it is resolved directly from collected_data. `sections`
// is deliberately NOT an InputSpec field — that name collides with the
// pages.sections column reachable through the current_page nested-source loop.
//
// Registration (action_registry.go):
//   "resolve_internal_links": { Handler: ResolveInternalLinksAction,
//       Category: "content", IsLocal: true }

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ResolveInternalLinksInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"page_type", "page_name"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("resolve_internal_links", ResolveInternalLinksInputSpec)
}

type contentHub struct {
	Name     string
	Title    string
	URL      string
	Area     string
	NavOrder int
}

var areasExcludedFromCTA = map[string]bool{
	"about": true, "contact": true, "privacy": true, "terms": true, "legal": true,
}

// ctaFieldNames maps a CTA component to its primary/secondary url field names.
// An empty second entry means the component has a single CTA url field.
//
// STAGED ROLLOUT (council trail 2525f980): this map is now an OVERRIDE on the
// schema-derived pairing in datahelpers/ctafields.go, which currently runs
// OBSERVE-ONLY (the delta between the two is logged below; the map still
// decides every write). Stage 2 inverts the precedence after a real build's
// delta log has been reviewed, via a further council round — see the doc_notes
// rows under subject_keys resolve_internal_links / rerender_page_sections.
//
// This set is shared by BOTH writers of CTA destinations:
//   - build time: this action (setCTAField)
//   - repair time: rerender_page_sections' cta_links_stale recompute
//     (applyCTARecompute)
//
// The misdirected_cta discovery check scans anchors on EVERY component, so a
// button-bearing component missing from this set is detectable but not
// repairable — its findings can only escalate to human review. A component
// added here must ALSO have its url fields' schema source flipped to
// "renderer" (migrations 091, 098): a site_specs.*/pages.*/static source is
// re-resolved into resolved_data on every render and merges last, so no
// recompute or content edit can win against it.
var ctaFieldNames = map[string][2]string{
	"hero":                   {"cta_url", "secondary_cta_url"},
	"call-to-action":         {"primary_cta_url", "secondary_cta_url"},
	"archetype-grid":         {"cta_url", ""},
	"archetype-combinations": {"cta_primary_url", "cta_secondary_url"},
	"gauntlet-cta":           {"cta_primary_url", "cta_secondary_url"},
	"content-block-about":    {"cta_url", ""},
}

func ResolveInternalLinksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "resolve_internal_links"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, ResolveInternalLinksInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	pageType := inputs.Get("page_type")
	pageName := inputs.Get("page_name")

	// `sections` config value is a PATH (e.g. "input_data.section_plan.sections_ready").
	// Resolve it directly from collected_data: it is a complex array value (not a
	// scalar ExtractActionInputs could return) and keeping it out of the InputSpec
	// avoids the current_page.sections field-name collision.
	var sections []interface{}
	if sectionsPath, ok := params.StepConfig.Config["sections"].(string); ok && sectionsPath != "" {
		if raw := datahelpers.ExtractNestedField(params.CollectedData, sectionsPath); raw != nil {
			sections, _ = raw.([]interface{})
		}
	}

	hubs, err := loadContentHubs(ctx, params, siteID, logger)
	if err != nil {
		return nil, err
	}
	interactive, err := loadInteractivePages(ctx, params, siteID, logger)
	if err != nil {
		return nil, err
	}
	validPages := loadResolverPageSet(ctx, params, siteID, logger)
	primary, secondary := chooseCTATargets(pageType, pageName, interactive, hubs)
	candidates := candidatesFromHubs(interactive, hubs)

	// Existing published label per slot, if this page has been built before —
	// nil/empty for a brand-new page, which is fine: there is nothing yet to
	// match against, so every section falls through to today's positional pick
	// exactly as before. Loaded once per page, not per section.
	existingLabels := loadExistingSectionContentData(ctx, params, siteID, pageName, logger)

	var unresolved []map[string]interface{}
	for _, raw := range sections {
		section, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		function := sectionComponentFunction(section)
		fields, isCTA := ctaFieldNames[function]
		// STAGE 1 (observe-only, trail 2525f980): derive the CTA pairing from
		// the component's own schema and log the delta vs the map, plus a Warn
		// per url field the derivation cannot see. ctaFieldNames still decides
		// every write this stage.
		var labelFieldOf map[string]string
		if schema := sectionInputSchema(section); schema != nil {
			derived := datahelpers.DeriveCTAURLFields(schema)
			if d := ctaDerivationDelta(fields, isCTA, derived); d != "" {
				logger.Info("resolve_internal_links: cta derivation delta (observe-only)",
					zap.String("component", function), zap.String("delta", d))
			}
			for _, f := range datahelpers.UncoveredCTAURLFields(schema) {
				logger.Warn("resolve_internal_links: uncovered cta url field",
					zap.String("component", function), zap.String("field", f))
			}
			labelFieldOf = make(map[string]string, len(derived))
			for _, cf := range derived {
				labelFieldOf[cf.URLField] = cf.LabelField
			}
		}
		if !isCTA {
			continue
		}
		sectionName := stringOrEmpty(section["name"])
		resolved := sectionResolvedData(section)
		existing := existingLabels[sectionName]
		setCTAField(resolved, fields[0], primary, validPages, function, sectionName, "primary", &unresolved,
			existingLabelFor(existing, labelFieldOf[fields[0]]), candidates)
		setCTAField(resolved, fields[1], secondary, validPages, function, sectionName, "secondary", &unresolved,
			existingLabelFor(existing, labelFieldOf[fields[1]]), candidates)
		section["resolved_data"] = resolved
	}

	logger.Info("resolve_internal_links: augmented CTA sections",
		zap.String("page_type", pageType),
		zap.String("page_name", pageName),
		zap.Int("section_count", len(sections)),
		zap.Int("hub_count", len(hubs)),
		zap.Int("interactive_count", len(interactive)),
		zap.Int("unresolved", len(unresolved)))

	// Build-time detectability: a correctly-dropped CTA leaves no fingerprint in
	// deployed HTML (the gated template renders no button), so the absence is
	// only knowable HERE. Emit one HITL work item per affected section,
	// mirroring createDeferredItems (same dedup + status semantics). Rebuilding
	// cannot fix "no hub exists", so these go to review, not to a handler.
	if len(unresolved) > 0 {
		emitUnresolvedCTAItems(ctx, params, siteID, pageName, unresolved, logger)
	}

	return map[string]interface{}{
		"sections_ready": sections,
		"unresolved":     unresolved,
	}, nil
}

// emitUnresolvedCTAItems inserts an unresolved_cta work item per affected
// section (needs_human_review, deduped on item_key — ON CONFLICT DO NOTHING,
// same as createDeferredItems). Failures are logged, never returned: the
// signal must not block the build.
func emitUnresolvedCTAItems(ctx context.Context, params ActionParams, siteID uuid.UUID,
	pageName string, unresolved []map[string]interface{}, logger *zap.Logger) {

	pageKey := pageName
	if pageKey == "" {
		pageKey = "unknown-page"
	}

	// Group by section: one item per section, listing its missing fields.
	type sectionMiss struct {
		component string
		fields    []string
	}
	bySection := make(map[string]*sectionMiss)
	order := make([]string, 0, len(unresolved))
	for _, u := range unresolved {
		section := stringOrEmpty(u["section"])
		if _, seen := bySection[section]; !seen {
			bySection[section] = &sectionMiss{component: stringOrEmpty(u["component"])}
			order = append(order, section)
		}
		bySection[section].fields = append(bySection[section].fields, stringOrEmpty(u["field"]))
	}

	for _, section := range order {
		miss := bySection[section]

		spec := map[string]interface{}{
			"page_name":    pageName,
			"section_name": section,
			"component":    miss.component,
			"missing":      miss.fields,
			"source":       "resolve_internal_links",
			"fix": "No real page exists to serve as this CTA's destination " +
				"(no eligible content hub). The gated template renders no button. " +
				"Add/activate a section-index hub, or set the destination manually.",
		}
		specJSON, _ := json.Marshal(spec)

		summary := fmt.Sprintf("Unresolved CTA on %s ('%s'): no real-page destination for %s",
			pageName, section, strings.Join(miss.fields, ", "))
		if len(summary) > 250 {
			summary = summary[:247] + "..."
		}

		itemKey := fmt.Sprintf("unresolved_cta_%s_%s_%s",
			pageKey, sanitiseSectionKey(section), siteID)

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, priority, status, created_by, item_key
			) VALUES ($1, 'internal-link-resolver', 'build', 'unresolved_cta', 'low', $2,
					  $3::jsonb, 30, 'needs_human_review', 'resolve_internal_links', $4)
			ON CONFLICT DO NOTHING
		`, siteID, summary, string(specJSON), itemKey)
		if err != nil {
			logger.Warn("resolve_internal_links: failed to insert unresolved_cta item",
				zap.String("section", section), zap.Error(err))
		} else {
			logger.Info("resolve_internal_links: unresolved_cta item created",
				zap.String("page", pageName), zap.String("section", section),
				zap.Strings("missing", miss.fields))
		}
	}
}

// setCTAField writes a validated url into resolved_data, or records it
// unresolved. Alongside the url it writes the target's title under
// "<field-minus-_url>_target_title" (cta_url -> cta_target_title,
// primary_cta_url -> primary_cta_target_title) so the content writer can
// write CTA copy FOR the actual destination instead of guessing one.
//
// If this slot already has a published label (existingLabel, non-empty only
// when the page has been built before — see loadExistingSectionContentData),
// a real candidate whose name/title/nav_label matches that label's own
// distinctive tokens is preferred over the generic positional pick. This is
// the bugs_open/203 follow-on: chooseCTATargets alone cannot tell "Run the
// Risk Checker" from any other button on the same page, so two labelled
// slots on one page could otherwise both receive the site's top-ranked hubs
// in nav-order regardless of what either button claims to do. A generic
// label ("Get Started") or one matching no candidate falls through to
// today's positional behaviour unchanged.
func setCTAField(resolved map[string]interface{}, field string, target contentHub, validPages datahelpers.PageURLSet,
	function, sectionName, slot string, unresolved *[]map[string]interface{},
	existingLabel string, candidates []datahelpers.LabelMatchCandidate) {
	if field == "" {
		return // single-URL component — no field in this slot, nothing to resolve or report
	}
	if existingLabel != "" {
		if match, ok := datahelpers.BestLabelMatch(existingLabel, candidates); ok && validPages.Contains(match.URL) {
			resolved[field] = match.URL
			if match.Title != "" {
				resolved[ctaTargetTitleField(field)] = match.Title
			}
			return
		}
	}
	if target.URL != "" && validPages.Contains(target.URL) {
		resolved[field] = target.URL
		if title := targetTitle(target); title != "" {
			resolved[ctaTargetTitleField(field)] = title
		}
		return
	}
	*unresolved = append(*unresolved, map[string]interface{}{
		"section":   sectionName,
		"component": function,
		"field":     field,
		"slot":      slot,
	})
}

// candidatesFromHubs converts the ranked interactive/hub contentHub lists
// into label-match candidates. Both lists are already filtered (excluded
// areas, the page's own URL) by chooseCTATargets' own rank() — this reuses
// that filtering rather than re-deriving it, so the candidate set label
// matching searches is always a subset of what positional choice would have
// picked from.
func candidatesFromHubs(interactive, hubs []contentHub) []datahelpers.LabelMatchCandidate {
	var out []datahelpers.LabelMatchCandidate
	add := func(list []contentHub, isInteractive bool) {
		for _, h := range list {
			if c, ok := datahelpers.NewLabelMatchCandidate(h.Name, h.Name, h.Title, h.URL, isInteractive, ""); ok {
				out = append(out, c)
			}
		}
	}
	add(interactive, true)
	add(hubs, false)
	return out
}

// existingLabelFor reads a string field from a possibly-nil content_data map,
// returning "" (never matches, falls through to positional pick) for a
// missing/non-string/empty value or an unknown label field name.
func existingLabelFor(contentData map[string]interface{}, labelField string) string {
	if contentData == nil || labelField == "" {
		return ""
	}
	s, _ := contentData[labelField].(string)
	return s
}

// loadExistingSectionContentData returns this page's CURRENTLY PERSISTED
// content_data per slot, keyed by slot_name — the published label a build is
// about to overwrite, if the page already exists. Returns an empty map (not
// an error) for a brand-new page or a load failure: an existing label is an
// enhancement to today's positional pick, never a requirement, so a miss here
// must degrade to exactly today's behaviour rather than block resolution.
func loadExistingSectionContentData(ctx context.Context, params ActionParams, siteID uuid.UUID, pageName string, logger *zap.Logger) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	if pageName == "" {
		return out
	}
	rows, err := params.DB.QueryContext(ctx, `
		SELECT COALESCE(pc.slot_name, ''), pc.content_data
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1 AND p.name = $2
	`, siteID, pageName)
	if err != nil {
		logger.Warn("resolve_internal_links: loadExistingSectionContentData query failed, degrading to no existing labels", zap.Error(err))
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var slotName string
		var cdJSON []byte
		if err := rows.Scan(&slotName, &cdJSON); err != nil {
			continue
		}
		if len(cdJSON) == 0 {
			continue
		}
		var cd map[string]interface{}
		if err := json.Unmarshal(cdJSON, &cd); err != nil {
			continue
		}
		out[slotName] = cd
	}
	if err := rows.Err(); err != nil {
		logger.Warn("resolve_internal_links: loadExistingSectionContentData iteration failed", zap.Error(err))
	}
	return out
}

// ctaTargetTitleField derives the companion title field for a CTA url field.
func ctaTargetTitleField(urlField string) string {
	return strings.TrimSuffix(urlField, "_url") + "_target_title"
}

// targetTitle prefers the page's human title over its internal name.
func targetTitle(t contentHub) string {
	if t.Title != "" {
		return t.Title
	}
	return t.Name
}

// chooseCTATargets — v2: interactive pages (tool/game) first, then content
// hubs, each group by nav_order; excluding the page itself (by name) and
// utility/legal destinations. Zero-value contentHub => no sensible target.
// pageType is carried for a future intent-aware (LLM) upgrade; v2 does not
// branch on it. Sites with no interactive pages behave exactly as v1.
func chooseCTATargets(pageType, pageName string, interactive, hubs []contentHub) (contentHub, contentHub) {
	rank := func(candidates []contentHub) []contentHub {
		ordered := make([]contentHub, 0, len(candidates))
		for _, h := range candidates {
			if areasExcludedFromCTA[h.Area] || ctaExcludedDestination(h.URL) {
				continue
			}
			if pageName != "" && h.Name == pageName { // don't point a page's hero at itself
				continue
			}
			ordered = append(ordered, h)
		}
		sort.SliceStable(ordered, func(i, j int) bool {
			if ordered[i].NavOrder != ordered[j].NavOrder {
				return ordered[i].NavOrder < ordered[j].NavOrder
			}
			return ordered[i].Name < ordered[j].Name
		})
		return ordered
	}

	ordered := append(rank(interactive), rank(hubs)...)

	var primary, secondary contentHub
	if len(ordered) > 0 {
		primary = ordered[0]
	}
	if len(ordered) > 1 {
		secondary = ordered[1]
	}
	return primary, secondary
}

// ctaExcludedDestination reports whether a URL lands in an area a CTA should
// never point at (contact, legal, about...). firstPathSegment cannot express
// this for top-level pages ("/contact.html" has no second slash), so the test
// normalises first and strips the .html suffix: "/contact.html" -> "contact",
// "/legal/privacy.html" -> "legal".
func ctaExcludedDestination(url string) bool {
	p := strings.TrimPrefix(datahelpers.NormalizePagePath(url), "/")
	if i := strings.IndexByte(p, '/'); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimSuffix(p, ".html")
	return areasExcludedFromCTA[p]
}

// sectionInputSchema extracts the component's input_schema from a
// sections_ready entry. plan_sections attaches the full component row
// (Component: comp.Raw), whose input_schema is a JSON string after a DB
// round-trip or an already-parsed map in-process — ParseInputSchemaValue
// handles both.
func sectionInputSchema(section map[string]interface{}) map[string]interface{} {
	comp, ok := section["component"].(map[string]interface{})
	if !ok {
		return nil
	}
	return datahelpers.ParseInputSchemaValue(comp["input_schema"])
}

// ctaDerivationDelta describes how the schema-derived CTA field set differs
// from the hardcoded map's entry for this component. Empty string = identical
// coverage (nothing to observe). Observe-only: consumed by a log line, never
// by control flow.
func ctaDerivationDelta(mapped [2]string, isCTA bool, derived []datahelpers.CTAField) string {
	mappedSet := make(map[string]bool, 2)
	if isCTA {
		for _, f := range mapped {
			if f != "" {
				mappedSet[f] = true
			}
		}
	}
	var extra, missing []string
	seen := make(map[string]bool, len(derived))
	for _, cf := range derived {
		seen[cf.URLField] = true
		if !mappedSet[cf.URLField] {
			extra = append(extra, cf.URLField+"("+cf.Source+")")
		}
	}
	for f := range mappedSet {
		if !seen[f] {
			missing = append(missing, f)
		}
	}
	if len(extra) == 0 && len(missing) == 0 {
		return ""
	}
	sort.Strings(extra)
	sort.Strings(missing)
	var parts []string
	if len(extra) > 0 {
		parts = append(parts, "derived-not-mapped: "+strings.Join(extra, ","))
	}
	if len(missing) > 0 {
		parts = append(parts, "mapped-not-derived: "+strings.Join(missing, ","))
	}
	return strings.Join(parts, "; ")
}

func sectionComponentFunction(section map[string]interface{}) string {
	if comp, ok := section["component"].(map[string]interface{}); ok {
		if fn := stringOrEmpty(comp["function"]); fn != "" {
			return fn
		}
		if nm := stringOrEmpty(comp["name"]); nm != "" {
			return nm
		}
	}
	// Fall back to the section name (often equals the component function).
	return stringOrEmpty(section["name"])
}

func sectionResolvedData(section map[string]interface{}) map[string]interface{} {
	if rd, ok := section["resolved_data"].(map[string]interface{}); ok && rd != nil {
		return rd
	}
	return map[string]interface{}{}
}

// THE BUILD-AXIS ARM ON THE THREE LOADERS BELOW (added 2026-08-09).
//
// All three used to carry only the LIFECYCLE arm — `status IN ('active',
// 'deployed')`, or NOT IN ('deleted','archived') — which answers "does the
// platform still want this page served?" and says nothing about whether it ever
// WAS. A page row is `active` from the moment the planner creates it, so a
// planned-but-never-built page was a valid CTA target and a valid member of the
// resolver's page set. datahelpers' own doc says the two axes are separate on
// purpose and that a caller must "pair this with whichever build-axis arm YOUR
// question needs"; these three were missing theirs.
//
// Measured on fundamentallyai.com 2026-08-08: the cost-calculator guide served
// `<a href="/platform-log/index.html">Platform Log</a>` for 18 days while that
// page was `status='active'`, `build_status='planned'`, `deployed_at IS NULL` —
// a live 404 emitted BY the resolver, because loadResolverPageSet counted it as
// a real page. The same day's planning pass created three more rows in exactly
// that state (/tools/tools/index.html and two duplicates), so the population is
// not a one-off.
//
// The obligation is the one queryresolve's FetchablePageEligibilitySQL already
// applies to every page LISTING — "a listing must never advertise a page that
// would 404" (bugs_open/052). Link resolution had the identical obligation and
// no predicate at all.
//
// WHY PageMayBeLinkedPredicateFor AND NOT PageHasShippedPredicateFor. The
// stricter sibling was tried first and MEASURED against live HTTP fleet-wide
// (2026-08-09): it would have dropped 39 pages from link candidacy, and **11 of
// them serve HTTP 200** — nine mortgagecalculator.co.uk pages, idea.uk's
// ab-test-calculator, webdesign.co.uk's llm-cost-calculator, nearly all
// `build_status='needs_rebuild'`, i.e. built once, still serving, never stamped
// with deployed_at. Delisting a working page is "worse than the bug"
// (bugs_open/052 addendum), so the shipped-predicate is the wrong tool for this
// question even though it is right for "may I mutate this".
//
// The narrower floor — never BUILT (`planned` + no deployed_at) — was measured
// the same way: 22 pages fleet-wide, **all 22 return 404, none returns 200**.
// That is the disconfirming test this change rests on, and it could have come
// out the other way, because the broader predicate did.
func loadContentHubs(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) ([]contentHub, error) {
	return loadCTACandidatePages(ctx, params, siteID, logger, "loadContentHubs", `
		SELECT name, COALESCE(title, name), url, COALESCE(nav_order, 100)
		FROM pages
		WHERE site_id = $1
		  AND page_type = 'section-index'
		  AND status IN ('active', 'deployed')
		  AND `+datahelpers.PageMayBeLinkedPredicateFor("")+`
	`)
}

// loadInteractivePages returns the site's tool/game pages — the destinations a
// CTA should prefer over a content hub when both exist ("Enter the Gauntlet"
// should land on the Gauntlet, not a section index).
func loadInteractivePages(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) ([]contentHub, error) {
	return loadCTACandidatePages(ctx, params, siteID, logger, "loadInteractivePages", `
		SELECT name, COALESCE(title, name), url, COALESCE(nav_order, 100)
		FROM pages
		WHERE site_id = $1
		  AND page_type IN ('tool', 'game')
		  AND status IN ('active', 'deployed')
		  AND `+datahelpers.PageMayBeLinkedPredicateFor("")+`
	`)
}

func loadCTACandidatePages(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger, caller, query string) ([]contentHub, error) {
	rows, err := params.DB.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, fmt.Errorf("%s query failed: %w", caller, err)
	}
	defer rows.Close()

	var hubs []contentHub
	for rows.Next() {
		var h contentHub
		if err := rows.Scan(&h.Name, &h.Title, &h.URL, &h.NavOrder); err != nil {
			logger.Warn(caller+": scan error", zap.Error(err))
			continue
		}
		h.Area = firstPathSegment(h.URL)
		hubs = append(hubs, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s iteration failed: %w", caller, err)
	}
	return hubs, nil
}

// loadResolverPageSet is the PAGE-CONTENT link-target set: every page the
// platform intends to serve, status floor only, with NO deployment predicate.
// That looseness is deliberate, not an oversight — a batch build resolves page
// A's CTA before page B deploys minutes later, and a content CTA is re-resolved
// on every page render and rerender, so a target that is merely not-yet-shipped
// is corrected by the next pass rather than frozen.
//
// CHROME MAY NOT USE THIS SET (bugs_open/191). Chrome ships on every page,
// renders once behind an idempotence gate, and has no repair pass, so the same
// looseness there is a site-wide 404 nothing later fixes. Chrome link targets go
// through LoadChromeLinkPolicy (chrome_link_policy.go). The caller allow-list
// for this function is enforced by chrome_link_policy_test.go.
func loadResolverPageSet(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) datahelpers.PageURLSet {
	rows, err := params.DB.QueryContext(ctx, `
		SELECT url FROM pages WHERE site_id = $1 AND status NOT IN ('deleted', 'archived')
		  AND `+datahelpers.PageMayBeLinkedPredicateFor("")+`
	`, siteID)
	if err != nil {
		logger.Warn("resolve_internal_links: page set load failed", zap.Error(err))
		return datahelpers.NewPageURLSet(nil)
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			continue
		}
		urls = append(urls, u)
	}
	return datahelpers.NewPageURLSet(urls)
}

// firstPathSegment("/tools/index.html") -> "tools"; "/index.html" -> "".
func firstPathSegment(url string) string {
	trimmed := strings.TrimPrefix(url, "/")
	if i := strings.IndexByte(trimmed, '/'); i >= 0 {
		return trimmed[:i]
	}
	return ""
}
