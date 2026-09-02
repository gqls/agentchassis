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

// contentHub is now the shared positional-candidate type (bugs_open/436): the
// supply SQL, the ranking and the eligibility filter live in
// datahelpers/cta_positional.go so the cta_rank_anomaly discovery check can
// compute the same rank-1 the writers do without a hand-mirrored copy.
type contentHub = datahelpers.CTAPositionalCandidate

// areasExcludedFromCTA names the utility areas a FRESH CTA pick must never land
// in. It governs the POSITIONAL PICK ONLY — datahelpers.RankCTAPositionalCandidates.
// The set itself moved to datahelpers.CTAExcludedAreas (bugs_open/436) so the
// discovery checks stop carrying a hand-mirrored copy; this alias keeps every
// in-package reference and test meaning exactly what it always did.
//
// It NO LONGER governs the label match (bugs_open/308 Phase B): that supply is
// datahelpers.LoadCTALabelUniverse, which deliberately offers utility pages so
// a button whose copy says "Contact our supply team" can reach the contact
// page. candidatesFromHubs, which applied this set to the label-match supply,
// was deleted with that change rather than left callerless.
// Judging an already-STORED destination with it was bugs_open/248's clobber
// (slug cta_recompute_clobbers_authored_contact_links): "never newly SEND a
// generated CTA to contact" is a sound default; "never TRUST an existing link to
// contact" is a different and much stronger claim that happens to reuse the same
// set. See storedCTADestinationIsAuthored for the deliberate asymmetry, and
// resolve_internal_links_authored_destination_test.go for the test that pins it.
var areasExcludedFromCTA = datahelpers.CTAExcludedAreas

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
	// LABEL-MATCH supply is the SHARED universe (bugs_open/308 Phase B), not the
	// positional pick's hub/tool lists: the detector that files the repair and
	// the writers that perform it must answer "which pages may this label name?"
	// from one place, or the repair cannot reach what the check suggested. The
	// positional pick above is untouched and still refuses every utility area.
	//
	// A load failure is FATAL here, matching the two loaders above. Degrading to
	// an empty universe would silently disable every label match for the page —
	// the build would succeed, each CTA would take the positional pick, and the
	// only trace would be a Warn nobody reads.
	candidates, err := datahelpers.LoadCTALabelUniverse(ctx, params.DB, siteID)
	if err != nil {
		return nil, err
	}

	// Existing published label per slot, if this page has been built before —
	// nil/empty for a brand-new page, which is fine: there is nothing yet to
	// match against, so every section falls through to today's positional pick
	// exactly as before. Loaded once per page, not per section.
	existingLabels := loadExistingSectionContentData(ctx, params, siteID, pageName, logger)

	// stamp_cta_destination_guidance (OPT-IN, unsafe default OFF — the owner
	// ruling of 2026-08-02: new authority on a shared seam ships as a field a
	// reviewer of the CALLER can see; single live consumer today is
	// internal-link-resolver). When armed, each resolved CTA destination is
	// appended to the paired LABEL field's llm_field_specs description, which
	// the page-content-writer prompt already renders — so the writer is told
	// what the button's destination IS instead of inventing one. Measured
	// before this shipped (bugs_open/299): the *_target_title VALUE reached 0
	// of 182 sampled prompts; only the guidance sentence naming the field did.
	// Mis-typed values fail OFF, matching recordDeadURLControls' semantics.
	stampGuidance, _ := params.StepConfig.Config[ctaDestinationGuidanceConfigKey].(bool)

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
		setCTAField(resolved, existing, fields[0], primary, validPages, function, sectionName, "primary", &unresolved,
			existingLabelFor(existing, labelFieldOf[fields[0]]), candidates, pageName)
		setCTAField(resolved, existing, fields[1], secondary, validPages, function, sectionName, "secondary", &unresolved,
			existingLabelFor(existing, labelFieldOf[fields[1]]), candidates, pageName)
		section["resolved_data"] = resolved
		if stampGuidance {
			stampCTADestinationGuidance(section, labelFieldOf[fields[0]], resolved, fields[0])
			stampCTADestinationGuidance(section, labelFieldOf[fields[1]], resolved, fields[1])
		}
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
//
// `stored` is the slot's currently-persisted content_data (nil for a page that
// has never been built, which reads as "" everywhere below and degrades to
// exactly the old behaviour). It exists for the authored-destination branch —
// bugs_open/248 — and gives this writer the same shape as its sibling,
// rerender_page_sections' applyCTARecompute.
// ctaDestinationGuidanceConfigKey arms the destination stamp on the label
// field's llm_field_specs description (see the read site above). Unset or a
// non-bool value ⇒ pre-existing behaviour, byte for byte.
const ctaDestinationGuidanceConfigKey = "stamp_cta_destination_guidance"

// stampCTADestinationGuidance appends "Destination (fixed): <title>…" to the
// llm_field_specs entry for the CTA's LABEL field, once the URL field has a
// resolved companion title (page or non-page — the same gap produced both the
// bugs_closed/268 family and bugs_open/299's phone-dialling button). The
// specs pipe (plan_sections → llm_field_specs[].description → the writer
// prompt's per-field description) already exists; this only puts the datum on
// it. No-op when the label field is unknown (schema not attached in transit),
// the title is absent, or the section carries no specs — degrading to exactly
// today's behaviour, never blocking resolution.
func stampCTADestinationGuidance(section map[string]interface{}, labelField string, resolved map[string]interface{}, urlField string) {
	if labelField == "" || urlField == "" {
		return
	}
	title, _ := resolved[ctaTargetTitleField(urlField)].(string)
	if title == "" {
		return
	}
	specs, ok := section["llm_field_specs"].([]interface{})
	if !ok {
		return
	}
	for _, raw := range specs {
		spec, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if name, _ := spec["name"].(string); name != labelField {
			continue
		}
		desc, _ := spec["description"].(string)
		spec["description"] = strings.TrimSpace(desc +
			" Destination (fixed): " + title +
			". Write this CTA's text to name or clearly promise this destination; never promise a different one.")
		return
	}
}

func setCTAField(resolved, stored map[string]interface{}, field string, target contentHub, validPages datahelpers.PageURLSet,
	function, sectionName, slot string, unresolved *[]map[string]interface{},
	existingLabel string, candidates []datahelpers.LabelMatchCandidate, pageName string) {
	if field == "" {
		return // single-URL component — no field in this slot, nothing to resolve or report
	}
	// Carry the section's existing mint record forward before deciding anything
	// (bugs_open/308 Phase A). Both persist paths merge SHALLOWLY and the record
	// is a nested map, so a resolved_data stamping only THIS field would replace
	// the stored record wholesale and drop the sibling slot's stamp — after which
	// that slot reads as authored and freezes. Four of the six ctaFieldNames
	// components have two slots, so it is the common case.
	//
	// Deliberately INSIDE this function rather than once per section in the
	// caller's loop: a loop-level call can be deleted without any test noticing
	// (proven by mutation — removing it left every test green), whereas here it
	// is on the path of every unit test that exercises a keep. SeedCTAMinted only
	// fills entries not already present, so calling it once per field is
	// idempotent and order-independent.
	datahelpers.SeedCTAMinted(resolved, stored)
	if existingLabel != "" {
		// An AMBIGUOUS label (bugs_open/308: two different pages tie on every
		// ranking key that carries signal) reports !ok, so control falls to the
		// keeps below and the stored value stands. That is the safe direction:
		// the alternative is writing a destination chosen by alphabetical order.
		// BestLabelMatchForPage, not BestLabelMatch: a label naming the page it
		// SITS ON names nothing (bugs_open/308, the 2026-08-23 hand audit — 12%
		// of the widening's writes were self-links). This path has the page's
		// name but not its URL, which is the same key rank() has always used.
		if match, ok, _ := datahelpers.BestLabelMatchForPage(existingLabel, candidates, pageName, ""); ok && validPages.Contains(match.URL) {
			resolved[field] = match.URL
			datahelpers.SetCTAMinted(resolved, field, match.URL)
			if match.Title != "" {
				resolved[ctaTargetTitleField(field)] = match.Title
			}
			return
		}
	}
	if storedCTADestinationIsAuthored(stored, field, validPages) {
		storedURL, _ := stored[field].(string)
		// bugs_open/248, BUILD-path half. This writer had no keep branch at all,
		// so an authored /contact.html died on the next full regeneration even
		// once the rerender path stopped clobbering it — the recompute and the
		// rebuild are two writers of one field and only one of them was fixed.
		//
		// WRITTEN, not skipped, and that is the whole point: a fresh
		// content_data need not carry the old url, and plan_sections' stored
		// carry misses on non-deployed rows, conflicted duplicate slots and
		// mismatched slot names. Returning silently would leave the field to
		// whatever else happened to populate resolved_data — which for a gated
		// template means no button at all. A branch added to keep a link must
		// not be able to delete it.
		//
		// Positioned AFTER the label match on purpose (bugs_open/248's own
		// verification bar #2): a fabricated contact url whose label names a
		// real page is 203's defect and is still repaired. Inert for every
		// non-utility stored value — those are re-derived exactly as before.
		//
		// NO MINT STAMP HERE, and its absence is load-bearing rather than an
		// omission (bugs_open/308 Phase A): reaching this branch PROVES the
		// stamp does not cover storedURL, because storedCTADestinationIsAuthored
		// returns false as soon as it does. A conditional re-stamp would be dead
		// code, and an unconditional one would assert that the resolver minted a
		// value it is the entire point of this branch to say a person authored.
		resolved[field] = storedURL
		if title, _ := stored[ctaTargetTitleField(field)].(string); title != "" {
			resolved[ctaTargetTitleField(field)] = title
		}
		return
	}
	if storedURL, _ := stored[field].(string); storedURL != "" &&
		ctaExcludedDestination(storedURL) && validPages.Contains(storedURL) {
		// A MINTED utility destination (bugs_open/308 Phase B). Reaching this
		// line proves the mint record COVERS storedURL — the branch above
		// returns for a valid utility url it does not cover — so this is the
		// resolver's own earlier output, not a person's.
		//
		// It is kept anyway, and the reason is not provenance but what the
		// alternative is: the only branch below is the POSITIONAL pick, which
		// cannot produce a utility destination and would therefore replace a
		// working contact button with an unrelated tool page — bugs_open/248's
		// damage, arriving through 308's fix. Control only reaches here when the
		// label match declined (generic or ambiguous copy), so there is no
		// evidence for a move; and if the copy DOES name another page, the label
		// match above already won and this line is never reached.
		//
		// The invariant, stated positively and symmetric with rank()'s:
		// THE POSITIONAL PICK MAY NEITHER CHOOSE NOR DISPLACE A UTILITY
		// DESTINATION. Only a confident label match moves one.
		resolved[field] = storedURL
		if title, _ := stored[ctaTargetTitleField(field)].(string); title != "" {
			resolved[ctaTargetTitleField(field)] = title
		}
		return
	}
	if storedURL, _ := stored[field].(string); datahelpers.IsAuthoredNonPageCTADestination(storedURL) {
		// bugs_open/299, the NON-PAGE half of the same trap: tel:/mailto:/
		// external/named-fragment destinations fail validPages.Contains by
		// construction, so no resolver path can have produced one — it was
		// authored — and no earlier keep could see one, so it fell to the
		// positional pick and a phone button became a tool link. DISJOINT
		// from the 248 branch above (that one REQUIRES validPages
		// membership); no url can satisfy both predicates.
		//
		// WRITTEN, for the branch above's reasons — and the write is also
		// the URI repair: the fleet's live tel: values carry spaces and
		// parens RFC 3966 forbids, and NormalizeTelHref fixes the ones that
		// are unambiguous. The ones it refuses (the collapsed-trunk
		// "+440…") are kept RAW: inventing digits is a human's call, and
		// check_cta_nonpage files them for one.
		//
		// The companion title is COMPUTED, not carried: it is what lets the
		// content writer write copy FOR this destination ("a phone call to
		// …") instead of inventing a destination the copy then promises —
		// which is exactly how bugs_open/299's button was written.
		kept := storedURL
		if norm, ok := datahelpers.NormalizeTelHref(storedURL); ok {
			kept = norm
		}
		resolved[field] = kept
		if title := datahelpers.DescribeCTADestination(storedURL); title != "" {
			resolved[ctaTargetTitleField(field)] = title
		}
		return
	}
	if target.URL != "" && validPages.Contains(target.URL) {
		resolved[field] = target.URL
		datahelpers.SetCTAMinted(resolved, field, target.URL)
		if title := targetTitle(target); title != "" {
			resolved[ctaTargetTitleField(field)] = title
		}
		return
	}
	// UNRESOLVED FALLTHROUGH — the one branch that writes no url, and therefore
	// the one place a mint stamp can be lost (bugs_open/308 Phase A).
	//
	// plan_sections' PBP-039 carry may already have left a previously-minted url
	// in resolved[field]; this branch does not touch it, so that value is what
	// gets persisted. The plan→save funnel REPLACES content_data (RFC_042 §2,
	// DELETE+INSERT), so the previous generation's stamp does NOT survive the
	// persist on its own — and an unstamped valid utility url reads as authored
	// next cycle and is frozen for ever by the keep above. That is this bug's
	// own failure mode reintroduced through its fix, so the stamp is re-asserted
	// here.
	//
	// The fix belongs HERE and not in carryStored: setCTAField is handed `stored`
	// as a fresh page_components read (loadExistingSectionContentData at the call
	// site), not the carry's output, so the record is already in hand — and
	// PBP-039's carry, whose register entry says in terms "do not remove it, do
	// not reorder it", is left untouched.
	//
	// No re-stamp is needed HERE: the SeedCTAMinted at the top of this function
	// has already carried the stored record forward, and a carried value equals
	// the stored one by construction (plan_sections' carry reads the same row),
	// so the seeded entry covers it. A second, value-guarded re-stamp at this
	// branch was written first and then DELETED as redundant — mutation showed
	// removing it changed no test, while no-opping the seed fails both this
	// branch's test and the sibling-slot one. Two guards in series where one is
	// load-bearing is how an unexercised branch ships.
	*unresolved = append(*unresolved, map[string]interface{}{
		"section":   sectionName,
		"component": function,
		"field":     field,
		"slot":      slot,
	})
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
//
// The map is keyed by slot_name and the loop below is last-row-wins, so the
// filter and the ORDER BY are load-bearing rather than tidiness. 11 pages
// fleet-wide legitimately carry a duplicate slot_name, and this read now also
// decides whether an authored destination is KEPT (setCTAField's
// storedCTADestinationIsAuthored branch, bugs_open/248) — so an unordered read
// made that decision nondeterministic between two rows, and an unfiltered one
// let a section someone had REMOVED drive it. build_status IS DISTINCT FROM,
// not !=, because the column is nullable: != would silently drop every
// NULL-status row, which is the worse failure (same reasoning as
// rerender_page_sections' loadStoredSections).
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
		  AND `+datahelpers.NotRemoved("pc")+`
		ORDER BY pc.position ASC
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
// hubs, each group by nav_order; excluding the page itself (by name),
// utility/legal destinations, and pages whose row opts out of CTA targethood
// (`pages.eligible_as_cta_target = false`, bugs_open/436). Zero-value
// contentHub => no sensible target. pageType is carried for a future
// intent-aware (LLM) upgrade; v2 does not branch on it. Sites with no
// interactive pages behave exactly as v1.
//
// The ordering and every filter live in datahelpers.RankCTAPositionalCandidates
// so the cta_rank_anomaly discovery check computes the same rank-1 this
// function hands to all three of its callers — the build-time resolver, the
// rerender recompute, and the site header fallback (whose output is never
// persisted, which is why the eligibility filter binds HERE and not in a
// loader's WHERE clause).
func chooseCTATargets(pageType, pageName string, interactive, hubs []contentHub) (contentHub, contentHub) {
	ordered := append(datahelpers.RankCTAPositionalCandidates(pageName, interactive),
		datahelpers.RankCTAPositionalCandidates(pageName, hubs)...)

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
// never point at (contact, legal, about...). Shared as
// datahelpers.CTAExcludedDestination (bugs_open/436); this wrapper keeps the
// in-package call sites and tests unchanged.
func ctaExcludedDestination(url string) bool {
	return datahelpers.CTAExcludedDestination(url)
}

// storedCTADestinationIsAuthored reports whether an ALREADY-STORED CTA url must
// be treated as authored — written by a person or by content authoring, not by
// this resolver — and therefore kept rather than recomputed away.
//
// There is no provenance field on content_data: bugs_open/248 (slug
// cta_recompute_clobbers_authored_contact_links) established that a fabricated
// and an authored /contact.html are byte-identical, so no amount of care at the
// call site can tell them apart by value. Provenance is derived from the
// resolver's own constraints instead.
//
// THE INVARIANT: no resolver path can produce a utility-area destination.
//   - the positional pick cannot: chooseCTATargets' rank() drops every
//     candidate whose URL shape is excluded, so it can never be chosen;
//   - the label match cannot: candidatesFromHubs applies the same URL-shape
//     filter before matching, so BestLabelMatch can never return one.
//
// A stored url that is BOTH a valid page AND in a utility area was therefore
// not written by setCTAField or applyCTARecompute. Both keep it.
//
// Validity is load-bearing, not decoration. An INVALID /contact.html is
// bugs_open/203's phantom fallback on a site with no contact page, and
// replacing that is the repair, not the bug — which is why a url absent from
// validPages returns false here and falls through to the positional pick.
//
// WHAT WOULD BREAK THIS, and it can be broken at a distance: letting
// utility-area pages into the resolver's candidate supply — widening
// loadContentHubs/loadInteractivePages, or removing candidatesFromHubs' filter,
// or removing rank()'s excluded-area test. Any of those lets the resolver mint
// the very urls this predicate declares un-mintable, and both keep branches
// would then freeze the resolver's own output for ever. If you widen the
// candidate set to utility pages deliberately (bugs_closed/023's authored-intent
// forward direction), this predicate needs REAL recorded provenance first.
//
// ── bugs_open/308, PHASE A (2026-08-22): the provenance is now RECORDED, and
// the shape test is no longer the whole answer. ────────────────────────────
//
// The paragraph above describes the state this function was written in and is
// kept because it is still exactly right about the POSITIONAL route and about
// schema fallbacks. What has changed is that "the resolver cannot have written
// this" is no longer inferred from the resolver's constraints — it is read from
// a record the resolver writes (datahelpers.CTAMintedCovers / CTAMintedKey).
//
// Behaviour is UNCHANGED until the candidate set widens: with no stamps in the
// database and candidatesFromHubs still filtering, the resolver never mints a
// utility url, so the new conjunct never fires and this returns what it always
// did. It is what makes the widening (Phase B) safe rather than a regression:
// once a label CAN resolve to /contact.html, a stamped one is re-derivable and
// an unstamped one is still the person's.
//
// ── bugs_open/308, PHASE B (2026-08-23): THE WIDENING HAS LANDED, so the two
// paragraphs above are now HISTORY, not description. ───────────────────────
//
// `candidatesFromHubs` is DELETED. The label match is supplied by
// datahelpers.LoadCTALabelUniverse, which offers every page on the site —
// utility areas included — so **"no resolver path can produce a utility-area
// destination" is FALSE from this commit onward**, by design and by owner
// ruling (2026-08-18: "keep a provenance record", then widen). The bullet list
// above is retained verbatim because it is the reason the record had to exist
// first; do not read it as a live invariant.
//
// What is still true, and is the invariant a reader should hold instead:
//
//	THE POSITIONAL PICK MAY NEITHER CHOOSE NOR DISPLACE A UTILITY DESTINATION.
//	Only a confident label match puts one there or moves it away.
//
// rank() enforces the first half (unchanged since 2026-07-14). The keep
// branches in setCTAField and applyCTARecompute enforce the second, and BOTH
// had to change for it: each previously let a MINTED utility destination — a
// state that could not exist before this commit — fall through to the
// positional pick the moment its label went generic, which is bugs_open/248's
// clobber arriving through 308's own fix.
//
// This predicate keeps its exact meaning ("a person put this here"), and it is
// still the thing that stops a recompute overwriting a real contact button. It
// simply no longer carries the whole weight: the keeps below it now hold a
// valid utility destination whatever its provenance, and the provenance record
// decides which of them writes it, not whether it survives.
//
// THE SIGNATURE IS THE ENFORCEMENT. It takes the stored map and the field name
// rather than a bare url so that no caller can reach the utility-area shape
// test without also handing over the map the mint check reads. A call site
// holding only a url string does not compile — which is the compile-time form
// of the owner's 2026-08-02 ruling that a comment is not a control on a tree
// this many sessions share.
func storedCTADestinationIsAuthored(stored map[string]interface{}, field string, validPages datahelpers.PageURLSet) bool {
	url, _ := stored[field].(string)
	if url == "" || !ctaExcludedDestination(url) || !validPages.Contains(url) {
		return false
	}
	// Recorded, not derived: a utility url this resolver minted is its own
	// output and must stay re-derivable, or the keep freezes it for ever.
	return !datahelpers.CTAMintedCovers(stored, field, url)
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
	hubs, err := datahelpers.LoadCTAPositionalCandidates(ctx, params.DB, siteID, datahelpers.CTAPositionalHubsSQL)
	if err != nil {
		return nil, fmt.Errorf("loadContentHubs: %w", err)
	}
	return hubs, nil
}

// loadInteractivePages returns the site's tool/game pages — the destinations a
// CTA should prefer over a content hub when both exist ("Enter the Gauntlet"
// should land on the Gauntlet, not a section index).
func loadInteractivePages(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) ([]contentHub, error) {
	pages, err := datahelpers.LoadCTAPositionalCandidates(ctx, params.DB, siteID, datahelpers.CTAPositionalInteractiveSQL)
	if err != nil {
		return nil, fmt.Errorf("loadInteractivePages: %w", err)
	}
	return pages, nil
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
// Shared as datahelpers.FirstPathSegment (bugs_open/436).
func firstPathSegment(url string) string {
	return datahelpers.FirstPathSegment(url)
}
