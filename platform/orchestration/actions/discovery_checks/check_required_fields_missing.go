// FILE: platform/orchestration/actions/discovery_checks/check_required_fields_missing.go
//
// Discovery check: deployed component instances whose schema-required value
// fields are absent or empty in content_data.
//
// The defect this catches (robot-hands gripper-detail, found 2026-07-14): a
// product-details instance whose content_data held 49 keys of chrome and
// boilerplate (sku_label, add_to_cart_label, size_option_1..4, nav_items…)
// while EVERY value field the template renders — product_name, product_price,
// feature_1..4, product_sku — was simply absent, despite the component schema
// declaring them {"source":"llm","required":true}. The template's
// {{.product_name}} resolved to "", shipping a fully-furnished e-commerce
// shell that sells nothing. sectionHasVisibleContent kept it (static label
// text counts as text), and empty_sections only fires on empty-heading /
// near-empty markup — so nothing owned this failure mode.
//
// Scope: only fields sourced from the LLM (source "llm" or no source) are
// checked. query.* / site_assets.* / pages.* fields resolve at render time,
// not in content_data — dartsonline's product-grid renders 14 real cards from
// query.affiliate_products with no product rows in content_data, and must not
// be flagged. Missing render-time data is owned by needs_section_data /
// on_missing enforcement in plan_sections, and image fields are owned by
// image_source_unsatisfiable.
//
// Flag-only: no handler agent, emitted at needs_human_review. page-build-
// handler cannot fix these (it rebuilds from spec sections; the live cases sit
// on entity pages whose sections list is empty), and the honest resolutions —
// give the site a data source, or remove the component — are human decisions.
//
// NOTE: input_schema is normally the v2 `fields` wrapper. A legacy JSON-Schema
// (`properties`+`required[]`) component is read via datahelpers.SchemaContentFields
// too — this audit is the post-deploy companion to the render gate, so it must
// not fail open on the same dialect the gate now handles (bugs_open/026).
//
// Retraction (2026-08-03): this check can now CLOSE required_fields_missing
// items it positively re-observes as filled, via CheckResult.Resolved. Second
// adopter of the RFC_010 seam, after check_empty_sections. Read
// findResolvedRequiredFields' comment before changing it — the refusals are
// load-bearing, and this check's flag-only design makes retraction the ONLY
// mechanism that can ever close one of its items (no handler agent, so
// CompleteWorkItemAction is never reached).

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&RequiredFieldsMissingCheck{}) }

type RequiredFieldsMissingCheck struct{}

func (c *RequiredFieldsMissingCheck) Name() string { return "required_fields_missing" }

// maxRequiredFieldFlagsPerPass bounds noise on badly-shaped sites; remaining
// gaps surface on later passes once earlier flags are resolved.
const maxRequiredFieldFlagsPerPass = 25

func (c *RequiredFieldsMissingCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.id, pc.page_id, p.name, COALESCE(pc.slot_name, ''),
		       COALESCE(cc.function, pc.slot_name, 'unknown'),
		       cc.input_schema::text, COALESCE(pc.content_data::text, '{}'),
		       COALESCE(pc.rendered_html, '') LIKE '%data-runtime-fill%'
		  FROM page_components pc
		  JOIN pages p ON p.id = pc.page_id
		  JOIN content_components cc ON cc.id = pc.component_id
		 WHERE p.site_id = $1
		   AND pc.build_status = 'deployed'
		   AND pc.locked_at IS NULL
		   AND cc.input_schema IS NOT NULL
		   AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')
		   AND COALESCE(cc.function, '') NOT IN ('header', 'footer', 'head-seo')
		   AND NOT (COALESCE(p.suppressed_sections, '[]'::jsonb) ? COALESCE(pc.slot_name, ''))
		 ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("required_fields_missing query failed: %w", err)
	}
	defer rows.Close()

	result := &CheckResult{}
	emitted := 0

	for rows.Next() {
		var componentID, pageID, pageName, slotName, function, schemaText, contentText string
		var isRuntimeFill bool
		if err := rows.Scan(&componentID, &pageID, &pageName, &slotName, &function,
			&schemaText, &contentText, &isRuntimeFill); err != nil {
			dctx.Logger.Warn("required_fields_missing: scan failed", zap.Error(err))
			continue
		}
		if isRuntimeFill {
			// Deliberately empty at build time — a browser-side loader fills it.
			continue
		}

		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(schemaText), &schema); err != nil {
			continue
		}
		fields, ok, fromLegacy := datahelpers.SchemaContentFields(schema)
		if !ok {
			continue
		}
		if fromLegacy {
			datahelpers.WarnLegacyDialect(dctx.Logger, "check_required_fields_missing", function)
		}
		var contentData map[string]interface{}
		if err := json.Unmarshal([]byte(contentText), &contentData); err != nil {
			contentData = map[string]interface{}{}
		}

		missing := missingRequiredValueFields(fields, contentData)
		if len(missing) == 0 {
			continue
		}

		if emitted >= maxRequiredFieldFlagsPerPass {
			// BREAK, NOT RETURN — and the difference is the whole trap.
			//
			// `return result, nil` here was correct while this check could only
			// FILE: the cap is about noise, so stopping is the right answer. It
			// became a bug the moment the check could also RETRACT, because it
			// skips the retraction below on exactly the sites that hit the cap —
			// the badly-shaped ones with the most stale items. It is green in
			// every test that stays under 25 findings, which is all of them.
			//
			// This is the same shape as the `len(sections) == 0` early return
			// documented in check_empty_sections.go, wearing a different hat: not
			// at the top of the function and not fired by an empty result, so
			// grepping for the documented form would not have found it. Any
			// monotonic check adopting Resolved must be read for EVERY early exit
			// between its scan and its retraction, not just the leading guard.
			dctx.Logger.Info("required_fields_missing: per-pass cap reached — filing stops here, retraction still runs",
				zap.Int("cap", maxRequiredFieldFlagsPerPass))
			break
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":              "required_fields_missing",
			"page":               pageName,
			"slot_name":          slotName,
			"component_function": function,
			"missing_fields":     missing,
		})

		spec, err := json.Marshal(map[string]interface{}{
			"check":              "required_fields_missing",
			"component_id":       componentID,
			"page_id":            pageID,
			"page_name":          pageName,
			"slot_name":          slotName,
			"component_function": function,
			"missing_fields":     missing,
			"reason":             "schema declares these fields required with source llm, but content_data never received them; the template renders them as empty strings",
		})
		if err != nil {
			continue
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "required_fields_missing",
			Severity: "medium",
			Summary: fmt.Sprintf("Component '%s' on page %s is missing %d schema-required value field(s): %s",
				function, pageName, len(missing), strings.Join(missing, ", ")),
			SpecJSON: string(spec),
			Priority: 140,
			// Flag-only: no handler, and needs_human_review keeps it out of
			// the dispatch loop while the dedup key holds it open.
			HandlerAgent: "",
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("required_fields_missing:%s:%s", pageID, slotName),
			BatchID:      dctx.BatchID,
		})
		emitted++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Closed before the retraction query so this check never holds two pooled
	// connections at once. sql.Rows.Close is idempotent, so the defer above is
	// still correct and still covers every early return.
	rows.Close()

	if emitted > 0 {
		dctx.Logger.Info("required_fields_missing: flagged components with missing required fields",
			zap.Int("count", emitted))
	}

	// ── Retraction: say what was positively observed FILLED ─────────────────
	//
	// A FAILED OBSERVATION MUST NEVER SUPPRESS A FINDING. Warn and carry on with
	// an empty Resolved rather than returning the error: the runner's `continue`
	// on a check error would drop this run's INSERTS too, so propagating a
	// retraction fault would trade a missed closure for a missed defect.
	// Degrading to "did not retract" leaves every item open, the safe direction.
	resolved, rerr := findResolvedRequiredFields(dctx)
	if rerr != nil {
		dctx.Logger.Warn("required_fields_missing: retraction observation failed — findings unaffected, nothing retracted",
			zap.Error(rerr))
	} else {
		result.Resolved = append(result.Resolved, resolved...)
	}

	return result, nil
}

// ============================================================================
// Retraction (RFC_010 Decision 1 — CheckResult.Resolved)
// ============================================================================

// findResolvedRequiredFields names the open required_fields_missing findings
// this run has POSITIVELY RE-OBSERVED to be filled. Second adopter of the
// RFC_010 seam, after check_empty_sections.
//
// ── WHY THIS TYPE PARTICULARLY NEEDS IT ────────────────────────────────────
//
// This check is FLAG-ONLY: no handler agent, items born needs_human_review.
// So CompleteWorkItemAction — the only other thing in the estate that closes a
// work item — is never reached for one of these, and until this function
// existed a required_fields_missing item could not close by ANY mechanism
// except a human hand on the database. Measured 2026-08-03: 70 rows ever, 59
// open, the oldest raised 2026-07-14. Six of the 59 name components whose
// content_data has since been filled in and are simply stale.
//
// ── IT ENUMERATES FROM THE ITEM SIDE, AND THAT IS THE SAFETY PROPERTY ───────
//
// The tempting implementation is "close every open item this run did not
// re-file". That is retraction by ABSENCE, which the seam forbids: a check that
// errored, or was blinded by a bad predicate, returns an empty finding set
// indistinguishable from a healthy site. Here it is worse than usual, because
// this check has a PER-PASS CAP (25): on a capped site the absence rule would
// close every item beyond the 25th precisely because the check stopped looking.
//
// So this walks the slots that required_fields_missing items actually name for
// this site and re-runs the predicate against what is in each one NOW. Every
// retraction is a statement about a row that was read.
//
// ── THE REFUSALS, each a real population measured against the live queue ────
//
//   - NO DEPLOYED COMPONENT IN THE SLOT → no retraction. Absence is equally
//     "fixed by removing the component" and "a rebuild silently deleted it",
//     this platform's most repeated failure (bugs_open/012, /021, /032).
//     3 of the 59 open items sit here.
//   - NO READABLE SCHEMA → no retraction, and this is the refusal unique to
//     this check. The predicate is `required fields declared by the schema that
//     are absent from content_data`, so a component whose input_schema is NULL,
//     unparseable, or carries no recognised field wrapper yields NO REQUIRED
//     FIELDS AT ALL — and a naive reading of that is "nothing missing, healthy".
//     It is the exact inverse: the observation could not be made. Treating an
//     unreadable schema as health would close an item every time a component's
//     schema was dropped, which is the same class of silent loss as the missing
//     component above. The filing half `continue`s on all three conditions and
//     that is correct there (no schema, nothing to assert); the same `continue`
//     on this side must never reach the healthy counter.
//   - RUNTIME-FILL SHELL → no retraction. The filing half skips these because a
//     browser-side loader supplies the content, so content_data being empty is
//     not a defect. That makes it a reason NOT TO FILE; it is not evidence the
//     fields arrived. This side has observed nothing about them, so it refuses.
//     Deliberately NARROWER than the negation of the filing predicate — every
//     refusal here errs toward leaving the item open.
//   - MIXED SLOT → no retraction. bugs_open/156: (page_id, slot_name) is not
//     unique. If any component in the slot still lacks a required field the
//     finding may still be true of that one, so the whole slot is left alone.
//   - STILL MISSING → no retraction, obviously. 50 of the 59.
//
// The invariant that keeps this readable: `healthy` is incremented ONLY where
// missingRequiredValueFields has actually been run and returned nothing. Every
// other path falls through to the `healthy != deployed` gate, so a new refusal
// is added by NOT counting rather than by remembering to veto.
//
// ── ONE PREDICATE, NOT TWO ─────────────────────────────────────────────────
//
// It re-runs missingRequiredValueFields — the same pure function the filing
// half uses, already unit tested. A second copy of "what counts as missing"
// drifting against the first is the defect class this lane keeps paying for.
//
// ── IT CARRIES NO STATUS VOCABULARY, DELIBERATELY ──────────────────────────
//
// No `status NOT IN (...)` filter. resolveWorkItems owns that predicate
// (workItemClosedStatuses); this function owns the observation. This package
// already holds two hand-rolled copies of the status list and they already
// disagree (check_truncated_component.go:163 includes 'cancelled',
// check_component_template_corrupted.go:141 omits it); a third, in a function
// whose job is CLOSING items, is how a list drift becomes a wrongly-closed
// defect. Reading a few already-closed rows costs a no-op UPDATE.
func findResolvedRequiredFields(dctx DiscoveryCheckContext) ([]ResolvedFinding, error) {
	// page_id is compared AS TEXT rather than cast to uuid: a cast would raise
	// on a malformed spec value and take the whole check down with it, and the
	// set is bounded at a few dozen keys per site.
	//
	// ⚠ THE COALESCE IS LOAD-BEARING HERE, NOT DEFENSIVE — and in the opposite
	// direction from check_empty_sections, which is why it is worth stating.
	// That check's filing half sets WorkItemSpec.PageID, so its column is always
	// populated and the COALESCE only pins a preference. THIS check's filing
	// half (above) never sets PageID: measured 2026-08-03, all 70
	// required_fields_missing rows ever written have a NULL page_id column and a
	// populated spec->>'page_id', 0 disagreements. So the spec fallback is the
	// arm that actually fires today, and the column arm is the one that costs
	// nothing until the filing half is fixed. Reading the column first is still
	// right: it is the first-class one, and a future filing change that populates
	// it would otherwise silently stop this from retracting, with no signal.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT k.item_key, pc.id, cc.input_schema::text,
		       COALESCE(pc.content_data::text, '{}'),
		       COALESCE(pc.rendered_html, '') LIKE '%data-runtime-fill%'
		FROM (
		    SELECT DISTINCT item_key,
		           COALESCE(page_id::text, spec->>'page_id') AS page_id,
		           COALESCE(spec->>'slot_name','')           AS slot_name
		      FROM site_work_items
		     WHERE site_id = $1
		       AND item_type = 'required_fields_missing'
		       AND item_key IS NOT NULL
		) k
		LEFT JOIN page_components pc
		       ON pc.page_id::text = k.page_id
		      AND COALESCE(pc.slot_name, '') = k.slot_name
		      AND pc.build_status = 'deployed'
		LEFT JOIN content_components cc ON cc.id = pc.component_id
		ORDER BY k.item_key
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("required_fields_missing retraction query failed: %w", err)
	}
	defer rows.Close()

	type slotObservation struct {
		deployed   int
		healthy    int
		detail     string
		unreadable int
	}
	observed := map[string]*slotObservation{}
	var order []string

	for rows.Next() {
		var itemKey string
		var componentID, schemaText sql.NullString
		var contentText string
		var isRuntimeFill bool
		if err := rows.Scan(&itemKey, &componentID, &schemaText, &contentText, &isRuntimeFill); err != nil {
			dctx.Logger.Warn("required_fields_missing: failed to scan retraction row", zap.Error(err))
			continue
		}
		obs, seen := observed[itemKey]
		if !seen {
			obs = &slotObservation{}
			observed[itemKey] = obs
			order = append(order, itemKey)
		}
		if !componentID.Valid {
			continue // LEFT JOIN miss: the slot holds no deployed component
		}
		obs.deployed++

		if isRuntimeFill {
			obs.unreadable++
			continue // a loader fills this client-side; content_data says nothing
		}
		if !schemaText.Valid {
			obs.unreadable++
			continue // no input_schema: the predicate cannot be re-run
		}
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(schemaText.String), &schema); err != nil {
			obs.unreadable++
			continue
		}
		fields, ok, _ := datahelpers.SchemaContentFields(schema)
		if !ok {
			obs.unreadable++
			continue // no recognised field wrapper: no required set to check against
		}
		var contentData map[string]interface{}
		if err := json.Unmarshal([]byte(contentText), &contentData); err != nil {
			obs.unreadable++
			continue
		}

		// The SAME pure predicate the filing half runs, three lines above.
		if missing := missingRequiredValueFields(fields, contentData); len(missing) == 0 {
			obs.healthy++
			obs.detail = fmt.Sprintf("%d schema field(s) checked", len(fields))
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("required_fields_missing retraction scan failed: %w", err)
	}

	var out []ResolvedFinding
	for _, itemKey := range order {
		obs := observed[itemKey]
		if obs.deployed == 0 || obs.healthy != obs.deployed {
			if obs.unreadable > 0 {
				dctx.Logger.Debug("required_fields_missing: refusing to retract an unobservable slot",
					zap.String("item_key", itemKey),
					zap.Int("deployed", obs.deployed),
					zap.Int("unreadable", obs.unreadable))
			}
			continue
		}
		reason := fmt.Sprintf(
			"re-observed filled: all %d deployed component(s) in this slot now carry every schema-required LLM value field (%s)",
			obs.deployed, obs.detail)
		// ItemKey, never AllOfType. The unit observed is one slot on one page,
		// exactly the unit the key addresses.
		out = append(out, ResolvedFinding{
			ItemType: "required_fields_missing",
			ItemKey:  itemKey,
			Reason:   reason,
		})
	}
	return out, nil
}

// missingRequiredValueFields returns the schema field names that are declared
// required, are LLM-sourced value fields (not images, not render-time
// sources), and are absent-or-empty in contentData. Sorted for stable
// summaries and dedup keys. Pure — unit tested.
func missingRequiredValueFields(fields, contentData map[string]interface{}) []string {
	var missing []string
	for name, defRaw := range fields {
		def, ok := defRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if !fieldFlagTrue(def["required"]) {
			continue
		}
		fieldType, _ := def["type"].(string)
		if fieldType == "image" || fieldType == "image_url" {
			continue // owned by image_source_unsatisfiable
		}
		source, _ := def["source"].(string)
		if source != "" && source != "llm" {
			continue // render-time sources (query.*, site_assets.*, pages.*) are not baked into content_data
		}
		if valueIsEmpty(contentData[name]) {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

// fieldFlagTrue accepts the bool/string encodings of `required` seen in
// stored schemas (true and "true").
func fieldFlagTrue(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(t, "true")
	}
	return false
}

// valueIsEmpty reports whether a content_data value is missing for rendering
// purposes: absent, nil, whitespace-only string, or empty array/object.
// Numbers and booleans are never empty (0 and false are legitimate values).
func valueIsEmpty(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(t) == ""
	case []interface{}:
		return len(t) == 0
	case map[string]interface{}:
		return len(t) == 0
	}
	return false
}
