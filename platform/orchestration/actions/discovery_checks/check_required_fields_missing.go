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
// Routed since 2026-08-15 (owner ruling, bugs_open/277): items are emitted at
// triaged for required-fields-missing-handler (seed 410), a pure-SQL router
// that classifies each finding against live DB state and converts, closes, or
// parks it per class. This supersedes the original flag-only design — its
// reasoning ("the honest resolutions — give the site a data source, or remove
// the component — are human decisions") survives inside the router: the two
// classes that genuinely need a human (blob components whose regeneration
// would replace served HTML, and owned/tool pages with no plan) are parked
// back to needs_human_review WITH the classification and the safe options,
// holding the dedup key so re-raises cannot churn.
//
// NOTE: input_schema is normally the v2 `fields` wrapper. A legacy JSON-Schema
// (`properties`+`required[]`) component is read via datahelpers.SchemaContentFields
// too — this audit is the post-deploy companion to the render gate, so it must
// not fail open on the same dialect the gate now handles (bugs_open/026).

package discovery_checks

import (
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

// requiredFieldsHandlerAgent is the router seeded by
// 410_required_fields_missing_router.sql (bugs_open/277, owner ruling
// 2026-08-15). It classifies each item by (page_name, slot_name) against live
// DB state and converts, closes, or parks it — this check no longer decides.
// Declared in block form so handler_coverage_test.go's const resolver sees it.
const (
	requiredFieldsHandlerAgent = "required-fields-missing-handler"
)

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
			dctx.Logger.Info("required_fields_missing: per-pass cap reached",
				zap.Int("cap", maxRequiredFieldFlagsPerPass))
			return result, nil
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
			// Routed (owner ruling 2026-08-15, bugs_open/277): born triaged
			// so the dispatch loop hands it to the router, which decides
			// convert vs close vs park per row. Seeded by 410; the roster
			// entry in handler_coverage_test.go is the contract.
			HandlerAgent: requiredFieldsHandlerAgent,
			Status:       "triaged",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("required_fields_missing:%s:%s", pageID, slotName),
			BatchID:      dctx.BatchID,
		})
		emitted++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if emitted > 0 {
		dctx.Logger.Info("required_fields_missing: flagged components with missing required fields",
			zap.Int("count", emitted))
	}
	return result, nil
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
			// Still owned by image_source_unsatisfiable. Left UNCHANGED by the
			// 2026-08-11 widening below, deliberately: adding image-typed rows
			// to this queue is a volume change nobody has measured, and the
			// 238 class is the type:"url" one that neither check could see.
			continue
		}
		source, _ := def["source"].(string)
		// CORRECTED 2026-08-11 (bugs_open/238): the old comment here read
		// "render-time sources (query.*, site_assets.*, pages.*) are not baked
		// into content_data", and for site_assets.* that is FALSE. plan_sections
		// resolves those into resolvedData, and RenderComponentAction's
		// merge_with overlay persists resolvedData INTO content_data — that is
		// PBP-014, and it is what makes no-LLM re-rendering possible at all. So a
		// site_assets.* field absent from a deployed row is exactly the
		// observable this check exists to report, and it was skipped on a
		// premise the platform contradicts.
		//
		// site_specs.* / pages.* / query.* stay skipped in this slice: their
		// flag volume is unmeasured, and the six link fields bugs_open/238 also
		// lost are {{if}}-gated (milder harm, authored degradation). Widening to
		// them is a separate change with its own census.
		//
		// WHAT THIS DOES TO THE SIBLING CHECK, stated because the landmine on
		// this family requires it: `image_source_unsatisfiable` was widened in
		// the same commit, from a type predicate to the `site_assets.` source
		// prefix. The two now overlap on `site_assets.*` fields that are not
		// image-typed, and that overlap is deliberate — they answer different
		// questions. That check asks "can this component's declared source EVER
		// be satisfied on this site" (a cause, per component); this one asks "is
		// this DEPLOYED ROW missing the value now" (the damage, per row). On
		// bugs_open/238 both were true and each names a different remedy.
		if source != "" && source != "llm" && !strings.HasPrefix(source, "site_assets.") {
			continue
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
