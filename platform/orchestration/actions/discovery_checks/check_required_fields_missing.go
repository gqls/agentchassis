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
// NOTE: input_schema uses a `fields` wrapper, NOT JSON-Schema `properties`.
// Querying `properties` returns nothing and will mislead you.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

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
		fields, ok := schema["fields"].(map[string]interface{})
		if !ok {
			continue
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
