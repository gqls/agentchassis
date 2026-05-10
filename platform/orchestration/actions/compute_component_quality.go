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
	"text/template"
	"text/template/parse"
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
	dataCompAttrRegex = regexp.MustCompile(`data-component\s*=`)
	sectionCloseRegex = regexp.MustCompile(`</section\s*>`)
	sectionOpenRegex  = regexp.MustCompile(`<section\b`)
)

// scoreComponent computes quality metrics for one component. Pure function:
// same inputs → same outputs. No DB access.
func scoreComponent(componentID, function, template, schemaJSON, componentLevel string) ComponentQualityResult {
	issues := []string{}

	// 1. Template variables — extract {{.foo}} and {{$.foo}} names
	tmplVars := extractTemplateVariables(template)
	tmplVarCount := len(tmplVars)

	// 2. Schema fields — parse input_schema JSON.
	//    schemaFields:    top-level fields (input_schema.fields.X keys)
	//    subSchemaFields: array sub-schema fields (input_schema.fields.X.items keys
	//                     where X has type=array)
	//
	// Top-level fields are required to appear as {{.X}} in the template.
	// Sub-schema fields are valid {{.X}} targets but are not required to
	// appear (the LLM declares a catalog and may use a subset).
	schemaFields := extractSchemaFields(schemaJSON)
	subSchemaFields := extractSubSchemaFields(schemaJSON)
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
		// Direction 1: every template var must be declared somewhere
		// (top-level OR in a sub-schema). Strict — a {{.X}} in the
		// template that has no schema declaration anywhere is a bug.
		for _, v := range tmplVars {
			_, inTopLevel := schemaFields[v]
			_, inSubSchema := subSchemaFields[v]
			if !inTopLevel && !inSubSchema {
				synced = false
				issues = append(issues, fmt.Sprintf("template var {{.%s}} has no schema entry", v))
				break
			}
		}
		// Direction 2: every TOP-LEVEL schema field must appear as a
		// template var. Sub-schema fields are NOT required here — the
		// LLM may declare nav_label in an array's items catalog and
		// choose not to render it in this particular template.
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
// top-level field names referenced in a Go-template body.
//
// Implementation: parses the template with Go's text/template package
// and walks the parse tree, collecting every FieldNode (which is what
// every .X reference becomes after parsing — direct access {{.X}},
// pipeline arg {{foo .X}}, range/with/if pipe {{range .X}}, and
// root-context {{$.X}} all parse to a FieldNode whose first ident
// is "X").
//
// Why parse instead of regex: the previous regex-based extractor had
// to be updated each time a new construct turned up (the original missed
// {{range .X}}, then {{$.X}}, then {{with .X}}, etc.). Each miss caused
// a silent schema/template sync failure. The Go parser, by definition,
// recognises every form the template language supports, so this stays
// correct as the language evolves.
//
// On parse failure: returns empty slice. The caller's structural check
// (TemplateClosed) will catch the unparseable-template case via the
// section-tag-balance heuristic; we don't fail loudly here because
// scoreComponent is also used for already-stored historical components
// that may pre-date current expectations.
//
// Funcs registered: only "placeholder" and "toJSON" are present in any
// agent prompt's funcMap, but the LLM's html_template output uses pure
// {{.X}}/{{range .X}}/{{end}} syntax — no helper calls. Registering
// no-op stubs costs little and makes the parser tolerant of the rare
// case where an LLM mistakenly emits a helper call in its output.
func extractTemplateVariables(tmplStr string) []string {
	if tmplStr == "" {
		return nil
	}
	// Recover from any unexpected nil-deref during parse-tree traversal.
	// Returning an empty slice on panic degrades gracefully — the sync
	// check will fall back to whatever schemaFields contains, and the
	// validator's structural checks (template_closed, has_data_component)
	// still run independently.
	var result []string
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = nil
			}
		}()
		funcs := template.FuncMap{
			"placeholder": func(string) string { return "" },
			"toJSON":      func(interface{}) string { return "" },
			"rangeStart":  func(string) string { return "" },
			"rangeEnd":    func() string { return "" },
		}
		trees, err := parse.Parse("component", tmplStr, "", "", funcs)
		if err != nil {
			return
		}
		seen := map[string]bool{}
		for _, t := range trees {
			if t == nil {
				continue
			}
			walkTemplateNode(t.Root, seen)
		}
		result = make([]string, 0, len(seen))
		for name := range seen {
			result = append(result, name)
		}
		sort.Strings(result)
	}()
	return result
}

// walkTemplateNode recursively descends a parse.Node, collecting every
// top-level field name into seen.
//
// Each typed-pointer case guards against a typed-nil interface value
// (interface holds a non-nil type descriptor but a nil pointer; the
// `n == nil` check at the top doesn't catch this case). RangeNode.List
// and similar can legitimately be nil when the parser leaves a branch
// empty, and the interface-conversion target type-switch matches the
// nil pointer rather than falling through.
func walkTemplateNode(n parse.Node, seen map[string]bool) {
	if n == nil {
		return
	}
	switch x := n.(type) {
	case *parse.ListNode:
		if x == nil {
			return
		}
		for _, c := range x.Nodes {
			walkTemplateNode(c, seen)
		}
	case *parse.ActionNode:
		if x == nil {
			return
		}
		walkTemplatePipe(x.Pipe, seen)
	case *parse.RangeNode:
		if x == nil {
			return
		}
		walkTemplatePipe(x.Pipe, seen)
		walkTemplateNode(x.List, seen)
		walkTemplateNode(x.ElseList, seen)
	case *parse.IfNode:
		if x == nil {
			return
		}
		walkTemplatePipe(x.Pipe, seen)
		walkTemplateNode(x.List, seen)
		walkTemplateNode(x.ElseList, seen)
	case *parse.WithNode:
		if x == nil {
			return
		}
		walkTemplatePipe(x.Pipe, seen)
		walkTemplateNode(x.List, seen)
		walkTemplateNode(x.ElseList, seen)
	case *parse.TemplateNode:
		if x == nil {
			return
		}
		walkTemplatePipe(x.Pipe, seen)
	}
	// Other node types (TextNode, CommentNode, etc.) carry no field refs.
}

// walkTemplatePipe walks a pipeline (the part after {{), which is a
// sequence of commands; each command has a list of args; each arg may
// itself be a node containing field refs.
func walkTemplatePipe(p *parse.PipeNode, seen map[string]bool) {
	if p == nil {
		return
	}
	for _, cmd := range p.Cmds {
		if cmd == nil {
			continue
		}
		for _, arg := range cmd.Args {
			if arg == nil {
				continue
			}
			switch a := arg.(type) {
			case *parse.FieldNode:
				if a == nil {
					continue
				}
				if len(a.Ident) > 0 {
					seen[a.Ident[0]] = true
				}
			case *parse.PipeNode:
				if a == nil {
					continue
				}
				walkTemplatePipe(a, seen)
			case *parse.ChainNode:
				if a == nil {
					continue
				}
				if len(a.Field) > 0 {
					seen[a.Field[0]] = true
				}
			}
			// VariableNode ($x), IdentifierNode (function names),
			// StringNode, NumberNode, etc. don't reference top-level
			// data fields.
		}
	}
}

// extractSchemaFields parses input_schema JSON and returns the field names
// declared at the TOP LEVEL of input_schema.fields.
//
// For Tier D array fields, the array's sub-schema field names (the keys
// under fields.X.items) are NOT included by this function. Use
// extractSubSchemaFields for those. The two are kept separate because
// they have different roles in the template/schema sync check:
//   - Top-level fields must appear as {{.X}} in the template (direction 2)
//   - Sub-schema fields appear as {{.X}} INSIDE a {{range}} block, but
//     don't have to appear at all — the LLM may declare nav_label as an
//     available field and choose not to render it in this template.
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

	for name := range rawFields {
		fields[name] = true
	}
	return fields
}

// extractSubSchemaFields returns field names declared inside any Tier D
// array field's sub-schema. These are the names that appear as
// {{.X}} INSIDE a {{range}} block in the template — title, url, etc.
//
// Sub-schema fields are valid template-var targets (so a {{.title}} in
// the template doesn't fail "no schema entry"), but they are NOT
// required to appear in the template. The LLM may declare a sub-schema
// catalog of available fields and use only some of them.
//
// Shape recognised:
//
//	"items": {
//	  "type": "array",
//	  "items": {
//	    "title": { "type": "text" },
//	    "url":   { "type": "url" }
//	  }
//	}
func extractSubSchemaFields(schemaJSON string) map[string]bool {
	subFields := make(map[string]bool)
	if schemaJSON == "" || schemaJSON == "{}" {
		return subFields
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &parsed); err != nil {
		return subFields
	}

	rawFields, ok := parsed["fields"].(map[string]interface{})
	if !ok {
		return subFields
	}

	for _, def := range rawFields {
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
			subFields[subName] = true
		}
	}
	return subFields
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
