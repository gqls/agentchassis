// FILE: platform/orchestration/actions/discovery_checks/check_component_template_corrupted.go
//
// Discovery check: component templates that are baked rendered output.
//
// Detects active content_components USED ON THIS SITE'S PAGES whose
// html_template is saved rendered output instead of a template: either a
// literal "<no value>" is baked in (the fingerprint of a render round-trip —
// Go's missingkey rendering saved back as the template), or the template has
// no {{...}} variables at all while its input_schema declares content fields.
// Such components render empty text and src="" no matter what the resolver
// supplies — the 2026-07-10 robot-hands finding (content-block-about,
// product-specs, info-card-grid; 11 active components fleet-wide).
//
// This is the BRIDGE the system was missing: compute_component_quality
// already flags exactly this ("section component has 0 template variables")
// but nothing turned the flag into action — the system-stats precedent was
// queued manually. This check emits needs_component_regeneration items in the
// same shape as that precedent, routed to component-creator.
//
// Cross-site guard: components are fleet-shared (one function, one active
// component), so before emitting, the check skips any component that already
// has an OPEN regeneration item from ANY site — the per-site dedup index
// cannot catch that.

package discovery_checks

import (
	"encoding/json"

	"go.uber.org/zap"
)

func init() { Register(&ComponentTemplateCorruptedCheck{}) }

type ComponentTemplateCorruptedCheck struct{}

func (c *ComponentTemplateCorruptedCheck) Name() string { return "component_template_corrupted" }

// Regenerations are LLM-heavy; cap emissions per site per pass.
const maxTemplateRegensPerPass = 5

func (c *ComponentTemplateCorruptedCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT cc.id, cc.name, cc.function,
		       COALESCE(cc.quality_score, 0),
		       COALESCE(cc.quality_issues::text, '[]'),
		       position('<no value>' IN cc.html_template) > 0 AS has_no_value
		  FROM page_components pc
		  JOIN pages p ON p.id = pc.page_id
		  JOIN content_components cc ON cc.id = pc.component_id
		 WHERE p.site_id = $1
		   AND cc.is_active = true
		   AND cc.html_template IS NOT NULL
		   AND (
		         cc.html_template LIKE '%<no value>%'
		      OR (position('{{' IN cc.html_template) = 0
		          AND jsonb_typeof(cc.input_schema->'fields') = 'object'
		          AND cc.input_schema->'fields' <> '{}'::jsonb)
		   )
		 ORDER BY cc.name
	`, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type corrupted struct {
		id, name, function, qualityIssues string
		qualityScore                      int
		hasNoValue                        bool
	}
	var found []corrupted
	for rows.Next() {
		var r corrupted
		if err := rows.Scan(&r.id, &r.name, &r.function, &r.qualityScore, &r.qualityIssues, &r.hasNoValue); err != nil {
			continue
		}
		found = append(found, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}
	emitted := 0
	for _, r := range found {
		if emitted >= maxTemplateRegensPerPass {
			dctx.Logger.Info("component_template_corrupted: per-pass cap reached",
				zap.Int("cap", maxTemplateRegensPerPass),
				zap.Int("found_total", len(found)))
			break
		}

		// Cross-site guard: skip if ANY site already has an open regeneration
		// item for this component (per-site dedup can't see other sites).
		var open int
		if err := dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT count(*) FROM site_work_items
			 WHERE item_type = 'needs_component_regeneration'
			   AND spec->>'component_id' = $1
			   AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved')
		`, r.id).Scan(&open); err != nil {
			dctx.Logger.Warn("component_template_corrupted: open-item guard query failed",
				zap.String("component", r.name), zap.Error(err))
			continue
		}
		if open > 0 {
			continue
		}

		reason := "html_template has no {{...}} variables but the schema declares content fields"
		if r.hasNoValue {
			reason = "literal <no value> baked into html_template (rendered output saved as template)"
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":         "component_template_corrupted",
			"component":     r.name,
			"function":      r.function,
			"component_id":  r.id,
			"quality_score": r.qualityScore,
			"reason":        reason,
		})

		spec, err := json.Marshal(map[string]interface{}{
			"function":       r.function,
			"component_id":   r.id,
			"section_type":   r.function,
			"quality_score":  r.qualityScore,
			"quality_issues": r.qualityIssues,
			"reason":         reason,
		})
		if err != nil {
			continue
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "needs_component_regeneration",
			Severity:     "high",
			Summary:      "Regenerate " + r.name + " — " + reason,
			SpecJSON:     string(spec),
			Priority:     5,
			HandlerAgent: "component-creator",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "component_template_corrupted:" + r.function,
			BatchID:      dctx.BatchID,
		})
		emitted++
	}

	if emitted > 0 {
		dctx.Logger.Info("component_template_corrupted: emitted regeneration items",
			zap.Int("count", emitted),
			zap.Int("found_total", len(found)))
	}
	return result, nil
}
