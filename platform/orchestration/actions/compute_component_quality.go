// FILE: platform/orchestration/actions/compute_component_quality_action.go
//
// ComputeComponentQualityAction calculates quality metrics for
// content_components templates and stores them in the quality_*
// columns. Can be invoked:
//   - Inline from store_generated_component after a new template is stored
//   - As a standalone action by component-quality-auditor to re-score
//     existing components periodically
//   - As a batch action to score ALL active components (config: scan_all: true)
//
// Quality is a composite score (0-100) reflecting contract compliance:
//   - Template closed properly (section tag balanced)
//   - data-component attribute present
//   - Template variables present for section-level components
//   - Schema fields match template variables (both directions)
//   - Schema field count reasonable (>= 1 for content components)
//
// The action does NOT modify the template itself — just scores it and
// flags issues. Regeneration is triggered separately by the auditor.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ComputeComponentQualityInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"component_id", "function", "scan_all", "below_score", "stale_days"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("compute_component_quality", ComputeComponentQualityInputSpec)
}

// ComponentQualityResult describes one component's computed quality.
type ComponentQualityResult struct {
	ComponentID           string   `json:"component_id"`
	Function              string   `json:"function"`
	TemplateVariableCount int      `json:"template_variable_count"`
	SchemaFieldCount      int      `json:"schema_field_count"`
	TemplateClosed        bool     `json:"template_closed"`
	SchemaTemplateSynced  bool     `json:"schema_template_synced"`
	HasDataComponent      bool     `json:"has_data_component"`
	QualityScore          int      `json:"quality_score"`
	QualityIssues         []string `json:"quality_issues"`
}

// ComputeComponentQualityAction scans one or many components and writes
// quality metrics to the DB.
//
// Config forms:
//
//	{ "component_id": "uuid" }           — score one component by id
//	{ "function": "hero" }                — score active component by function
//	{ "scan_all": true }                  — score every active component
//	{ "scan_all": true, "below_score": 50 } — re-score those below threshold
//	{ "scan_all": true, "stale_days": 7 } — re-score those not checked in N days
func ComputeComponentQualityAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "compute_component_quality"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		ComputeComponentQualityInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	componentID := inputs.Get("component_id")
	function := inputs.Get("function")
	scanAllRaw := inputs.GetRaw("scan_all")
	belowScoreRaw := inputs.GetRaw("below_score")
	staleDaysRaw := inputs.GetRaw("stale_days")

	scanAll, _ := scanAllRaw.(bool)

	// Build the query based on mode
	var rows *sql.Rows

	switch {
	case componentID != "":
		rows, err = params.DB.QueryContext(ctx, `
			SELECT id::text, COALESCE(function, ''), COALESCE(html_template, ''), COALESCE(input_schema::text, '{}'), component_level
			FROM content_components
			WHERE id = $1::uuid
		`, componentID)
	case function != "":
		rows, err = params.DB.QueryContext(ctx, `
			SELECT id::text, COALESCE(function, ''), COALESCE(html_template, ''), COALESCE(input_schema::text, '{}'), component_level
			FROM content_components
			WHERE function = $1 AND is_active = true AND forked_from IS NULL
			LIMIT 1
		`, function)
	case scanAll:
		query := `
			SELECT id::text, COALESCE(function, ''), COALESCE(html_template, ''), COALESCE(input_schema::text, '{}'), component_level
			FROM content_components
			WHERE is_active = true
		`
		var args []interface{}
		argN := 1
		if v, ok := belowScoreRaw.(float64); ok && v > 0 {
			query += fmt.Sprintf(" AND (quality_score IS NULL OR quality_score < $%d)", argN)
			args = append(args, int(v))
			argN++
		}
		if v, ok := staleDaysRaw.(float64); ok && v > 0 {
			query += fmt.Sprintf(" AND (quality_checked_at IS NULL OR quality_checked_at < NOW() - INTERVAL '%d days')", int(v))
		}
		query += " ORDER BY quality_checked_at ASC NULLS FIRST LIMIT 500"
		rows, err = params.DB.QueryContext(ctx, query, args...)
	default:
		return nil, fmt.Errorf("must provide component_id, function, or scan_all=true")
	}

	if err != nil {
		return nil, fmt.Errorf("query components: %w", err)
	}
	defer rows.Close()

	var results []ComponentQualityResult
	scored := 0
	errors := 0

	for rows.Next() {
		var (
			id, fn, tmpl, schema string
			level                sql.NullString
		)
		if err := rows.Scan(&id, &fn, &tmpl, &schema, &level); err != nil {
			logger.Warn("scan row failed", zap.Error(err))
			errors++
			continue
		}

		result := scoreComponent(id, fn, tmpl, schema, level.String)
		if err := persistQuality(ctx, params.DB, result); err != nil {
			logger.Warn("persist quality failed",
				zap.String("component_id", id),
				zap.Error(err))
			errors++
			continue
		}
		results = append(results, result)
		scored++
	}

	logger.Info("component quality scan complete",
		zap.Int("scored", scored),
		zap.Int("errors", errors),
		zap.Int("low_quality", countBelow(results, 60)))

	return map[string]interface{}{
		"status":      "complete",
		"scored":      scored,
		"errors":      errors,
		"low_quality": countBelow(results, 60),
		"results":     results,
	}, nil
}

// ---------------------------------------------------------------------------
// Scoring logic
// ---------------------------------------------------------------------------

var (
	tmplVarPattern    = regexp.MustCompile(`\{\{\s*\.([A-Za-z_][A-Za-z0-9_]*)`)
	dataCompAttrRegex = regexp.MustCompile(`data-component\s*=`)
	sectionCloseRegex = regexp.MustCompile(`</section\s*>`)
	sectionOpenRegex  = regexp.MustCompile(`<section\b`)
)

// scoreComponent computes quality metrics for one component. Pure function:
// same inputs → same outputs. No DB access.
func scoreComponent(componentID, function, template, schemaJSON, componentLevel string) ComponentQualityResult {
	issues := []string{}

	// 1. Template variables — extract {{.foo}} names
	tmplVars := extractTemplateVariables(template)
	tmplVarCount := len(tmplVars)

	// 2. Schema fields — parse input_schema JSON
	schemaFields := extractSchemaFields(schemaJSON)
	schemaFieldCount := len(schemaFields)

	// 3. Template closed — opening and closing <section> tags balance
	openCount := len(sectionOpenRegex.FindAllString(template, -1))
	closeCount := len(sectionCloseRegex.FindAllString(template, -1))
	templateClosed := openCount > 0 && openCount == closeCount

	// 4. data-component attribute present
	hasDataComponent := dataCompAttrRegex.MatchString(template)

	// 5. Template/schema sync — every {{.x}} has schema entry AND vice versa
	synced := true
	if tmplVarCount > 0 || schemaFieldCount > 0 {
		// Only check sync if at least one side has content
		for _, v := range tmplVars {
			if _, ok := schemaFields[v]; !ok {
				synced = false
				issues = append(issues, fmt.Sprintf("template var {{.%s}} has no schema entry", v))
				break
			}
		}
		if synced {
			for f := range schemaFields {
				found := false
				for _, v := range tmplVars {
					if v == f {
						found = true
						break
					}
				}
				if !found {
					synced = false
					issues = append(issues, fmt.Sprintf("schema field %q has no template variable", f))
					break
				}
			}
		}
	}

	// Compute composite score
	score := 100

	// Template not closed is a severe structural problem
	if !templateClosed {
		score -= 30
		issues = append(issues, "template not closed properly (unbalanced <section> tags)")
	}

	// Missing data-component breaks the rendering contract
	if !hasDataComponent {
		score -= 10
		issues = append(issues, "missing data-component attribute")
	}

	// Section-level components should have template variables for content.
	// Tool-level components often work with 0 variables (JS-driven).
	if componentLevel == "section" || componentLevel == "" {
		if tmplVarCount == 0 {
			score -= 50
			issues = append(issues, "section component has 0 template variables (content likely hardcoded)")
		}
		if schemaFieldCount == 0 && tmplVarCount > 0 {
			score -= 30
			issues = append(issues, fmt.Sprintf("template has %d variables but schema is empty", tmplVarCount))
		}
	}

	// Template/schema mismatch is the bug we keep hitting
	if !synced && (tmplVarCount > 0 || schemaFieldCount > 0) {
		score -= 20
	}

	if score < 0 {
		score = 0
	}

	return ComponentQualityResult{
		ComponentID:           componentID,
		Function:              function,
		TemplateVariableCount: tmplVarCount,
		SchemaFieldCount:      schemaFieldCount,
		TemplateClosed:        templateClosed,
		SchemaTemplateSynced:  synced,
		HasDataComponent:      hasDataComponent,
		QualityScore:          score,
		QualityIssues:         issues,
	}
}

// extractTemplateVariables returns a sorted, deduplicated list of
// Go-template variable names in the form {{.foo}}.
func extractTemplateVariables(template string) []string {
	matches := tmplVarPattern.FindAllStringSubmatch(template, -1)
	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) >= 2 {
			seen[m[1]] = true
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// extractSchemaFields parses input_schema JSON and returns the field names
// from input_schema.fields (if present).
//
// For Tier D array fields (type=array with an items sub-schema), the
// sub-schema's field names are ALSO added to the returned map. This is
// so the schema/template sync check accepts {{.title}}, {{.url}}, etc.
// references that appear inside a {{range .items}} block — those names
// are declared in the array's sub-schema, not at the top level.
//
// Without this, a Tier D template like
//
//	<div>{{range .items}}<h3>{{.title}}</h3>{{end}}</div>
//
// would fail validation because {{.title}} has no entry in fields (only
// fields.items.items.title exists). With this, .title is added to the
// returned map alongside .items, and the sync check passes.
func extractSchemaFields(schemaJSON string) map[string]bool {
	fields := make(map[string]bool)
	if schemaJSON == "" || schemaJSON == "{}" {
		return fields
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &parsed); err != nil {
		return fields
	}

	rawFields, ok := parsed["fields"].(map[string]interface{})
	if !ok {
		return fields
	}

	for name, def := range rawFields {
		fields[name] = true

		// If this is a Tier D array field, also include its sub-schema
		// field names. Shape:
		//   "items": {
		//     "type": "array",
		//     "items": {
		//       "title": { ... },
		//       "url":   { ... }
		//     }
		//   }
		defMap, defIsMap := def.(map[string]interface{})
		if !defIsMap {
			continue
		}
		fieldType, _ := defMap["type"].(string)
		if fieldType != "array" {
			continue
		}
		subSchema, subOk := defMap["items"].(map[string]interface{})
		if !subOk {
			continue
		}
		for subName := range subSchema {
			fields[subName] = true
		}
	}
	return fields
}

// persistQuality writes the scored result to content_components.
func persistQuality(ctx context.Context, db *sql.DB, r ComponentQualityResult) error {
	if _, err := uuid.Parse(r.ComponentID); err != nil {
		return fmt.Errorf("invalid component_id: %w", err)
	}

	issuesJSON, err := json.Marshal(r.QualityIssues)
	if err != nil {
		issuesJSON = []byte("[]")
	}

	_, err = db.ExecContext(ctx, `
		UPDATE content_components
		SET template_variable_count = $2,
		    schema_field_count      = $3,
		    template_closed         = $4,
		    schema_template_synced  = $5,
		    has_data_component      = $6,
		    quality_score           = $7,
		    quality_issues          = $8::jsonb,
		    quality_checked_at      = $9
		WHERE id = $1::uuid
	`,
		r.ComponentID,
		r.TemplateVariableCount,
		r.SchemaFieldCount,
		r.TemplateClosed,
		r.SchemaTemplateSynced,
		r.HasDataComponent,
		r.QualityScore,
		string(issuesJSON),
		time.Now(),
	)
	return err
}

func countBelow(results []ComponentQualityResult, threshold int) int {
	n := 0
	for _, r := range results {
		if r.QualityScore < threshold {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// Helper used by StoreGeneratedComponentAction to score a newly-stored component
// ---------------------------------------------------------------------------

// ScoreAndPersistComponent is an inline helper callable from other actions
// (e.g. store_generated_component) so newly-generated components get a
// quality score immediately without waiting for the periodic auditor.
func ScoreAndPersistComponent(
	ctx context.Context,
	db *sql.DB,
	componentID, function, template, schemaJSON, componentLevel string,
	logger *zap.Logger,
) ComponentQualityResult {
	result := scoreComponent(componentID, function, template, schemaJSON, componentLevel)
	if err := persistQuality(ctx, db, result); err != nil {
		logger.Warn("ScoreAndPersistComponent: persist failed",
			zap.String("component_id", componentID),
			zap.Error(err))
	} else {
		logger.Info("ScoreAndPersistComponent: scored",
			zap.String("component_id", componentID),
			zap.String("function", function),
			zap.Int("score", result.QualityScore),
			zap.Strings("issues", result.QualityIssues))
	}
	return result
}
