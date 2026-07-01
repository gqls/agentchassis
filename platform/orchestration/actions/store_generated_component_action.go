// FILE: platform/orchestration/actions/store_generated_component_action.go
//
// StoreGeneratedComponentAction stores an LLM-generated component template
// into the content_components table with full selection metadata.
//
// Used by the component-creator handler agent after execute_llm_prompt
// generates the html_template and input_schema.
//
// Registration:
//   "store_generated_component": {
//       Handler:     StoreGeneratedComponentAction,
//       Category:    "site",
//       Description: "Store a generated component template in the component library",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "store_component": {
//       "action": "store_generated_component",
//       "config": {
//           "section_type": "input_data.section_type",
//           "site_type": "input_data.site_type",
//           "generated_template": "generate_template"
//       },
//       "next_step": "complete",
//       "output_field": "stored_component"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var StoreGeneratedComponentInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"section_type"},
	Optional:   []string{"site_type", "page_context", "description", "design_direction", "generated_template"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("store_generated_component", StoreGeneratedComponentInputSpec)
}

func StoreGeneratedComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "store_generated_component"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		StoreGeneratedComponentInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	sectionType := inputs.Get("section_type")
	siteType := inputs.Get("site_type")
	pageContext := inputs.Get("page_context")
	description := inputs.Get("description")

	// Work item source (optional): the originating work item's `source`
	// field, if this action was triggered by one. Used as change_source
	// on component_versions rows so history can be traced back to the
	// audit/triage/manual action that caused the change. Empty string if
	// no work item (e.g. direct programmatic invocation) — the snapshot
	// helper writes NULL in that case.
	workItemSource := datahelpers.ExtractNestedFieldString(
		params.CollectedData, "input_data.source",
	)

	// The LLM output is in collected_data under the output_field of the generate step.
	// extract the generated template — look for the LLM result which contains
	// html_template and input_schema as structured output.
	generatedRaw := inputs.GetRaw("generated_template")

	htmlTemplate, inputSchemaJSON, functionName, isDark, err := parseGeneratedTemplate(generatedRaw, sectionType, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated template: %w", err)
	}

	if htmlTemplate == "" {
		return nil, fmt.Errorf("generated template is empty for section_type %q", sectionType)
	}

	// Separate inline <script> blocks into js_content.
	// The template keeps a <script src="/tools/assets/{function}.js"> reference.
	// Components without inline JS are unaffected (jsContent will be empty).
	var jsContent string
	htmlTemplate, jsContent = separateInlineJS(htmlTemplate, functionName)

	if jsContent != "" {
		logger.Info("store_generated_component: extracted inline JS to js_content",
			zap.String("function", functionName),
			zap.Int("js_length", len(jsContent)),
			zap.Int("template_length", len(htmlTemplate)))
	}

	// ── Validate template quality ───────────────────────────────────────
	// Reject templates that are clearly broken — CSS-only output, truncated
	// by token limit, or missing input_schema. Without these checks, broken
	// components enter the DB and silently cause every page using this
	// section type to render empty/CSS-only content.

	// Check 1: Template must contain HTML structure (section or div),
	// not just a <style> block.
	templateLower := strings.ToLower(htmlTemplate)
	if !strings.Contains(templateLower, "<section") && !strings.Contains(templateLower, "<div") {
		return nil, fmt.Errorf(
			"generated template for %q has no HTML structure (<section> or <div>) — likely CSS-only or truncated output",
			sectionType)
	}

	// Check 2: Unclosed <style> tags indicate token-limit truncation.
	styleOpens := strings.Count(templateLower, "<style")
	styleCloses := strings.Count(templateLower, "</style>")
	if styleOpens > styleCloses {
		return nil, fmt.Errorf(
			"generated template for %q has %d unclosed <style> tag(s) — likely truncated by token limit",
			sectionType, styleOpens-styleCloses)
	}

	// Check 3: Empty input_schema means the component has no content fields.
	// It can't accept LLM-generated content, so every page using it will
	// render the raw template with no substitution.
	if inputSchemaJSON == "{}" || inputSchemaJSON == "" || inputSchemaJSON == `{"fields":{}}` {
		logger.Warn("store_generated_component: empty input_schema — component has no content fields",
			zap.String("section_type", sectionType),
			zap.String("function", functionName))
		return nil, fmt.Errorf(
			"generated template for %q has empty input_schema — no content fields defined, page builds would produce empty sections",
			sectionType)
	}

	// Build suitable_site_types from the site_type that triggered the creation
	suitableSiteTypes := []string{}
	if siteType != "" {
		suitableSiteTypes = append(suitableSiteTypes, siteType)
	}
	suitableSiteTypesJSON, _ := json.Marshal(suitableSiteTypes)

	// Build suitable_page_types from page context if available
	suitablePageTypes := []string{}
	if pageContext != "" {
		suitablePageTypes = append(suitablePageTypes, pageContext)
	}
	suitablePageTypesJSON, _ := json.Marshal(suitablePageTypes)

	// Build display name from function
	displayName := datahelpers.FunctionToDisplayName(functionName)

	// Determine category from site_type
	category := "custom"
	if siteType != "" {
		category = siteType
	}

	logger.Info("store_generated_component: storing component",
		zap.String("section_type", sectionType),
		zap.String("function", functionName),
		zap.String("category", category),
		zap.Int("template_length", len(htmlTemplate)),
		zap.Bool("is_dark", isDark))

	// ── Check for existing component (regeneration vs creation) ─────────
	// If a component with this function name already exists, we're in the
	// regeneration path: the new template will REPLACE the existing one,
	// with the old state snapshotted to component_versions first. If not,
	// we're creating a new component.
	//
	// Either way, Layer 1 validation below MUST pass before we touch the
	// DB. An existing broken component is NOT grounds to silently accept
	// another broken template.
	var (
		existingID      string
		existingHTML    string
		existingSchema  string
		existingJS      sql.NullString
		existingVersion int // max version_number from component_versions; 0 if none
		isRegeneration  bool
	)
	err = params.DB.QueryRowContext(ctx, `
		SELECT id::text,
		       COALESCE(html_template, ''),
		       COALESCE(input_schema::text, '{}'),
		       js_content
		FROM content_components
		WHERE function = $1 AND forked_from IS NULL
		ORDER BY is_active DESC, updated_at DESC
		LIMIT 1
	`, functionName).Scan(&existingID, &existingHTML, &existingSchema, &existingJS)

	// Note on the is_active filter (changed 2026-05-06):
	// Previously this query had `AND is_active = true`, which meant
	// regeneration of a deactivated row would fall through to the
	// creation branch and hit the unique-on-name constraint when the
	// old row had name == function. Removing the filter and ordering
	// by `is_active DESC, updated_at DESC` preserves the previous
	// behaviour when active rows exist (the active row sorts first)
	// and fixes the regeneration path when only an inactive row
	// exists.
	//
	// Reactivation: the UPDATE branch below sets is_active = true
	// unconditionally. A regenerated template that passes all
	// pre-store quality gates is by definition healthy, and the
	// most common reason for a component being inactive is "broken
	// template, awaiting regeneration" (migration 036 set 42 rows
	// inactive on this basis). Resurrecting an operator-deactivated
	// component is acceptable here: if the operator wanted it gone
	// permanently, they'd delete it, not deactivate it.

	if err == nil {
		isRegeneration = true
		// Find the latest version_number so the snapshot we write gets
		// MAX+1. Unique index (component_id, version_number) enforces
		// monotonic numbering.
		if err := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(version_number), 0)
			FROM component_versions
			WHERE component_id = $1::uuid
		`, existingID).Scan(&existingVersion); err != nil {
			// Non-fatal: if the query fails we default to 0 and the
			// snapshot INSERT will use version_number=1. Log for visibility.
			logger.Warn("store_generated_component: could not read current max version_number, defaulting to 0",
				zap.String("component_id", existingID),
				zap.Error(err))
			existingVersion = 0
		}
		logger.Info("store_generated_component: regeneration — existing component found",
			zap.String("function", functionName),
			zap.String("existing_id", existingID),
			zap.Int("current_max_version", existingVersion))
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing component: %w", err)
	}

	// ── Layer 1 pre-store validation ────────────────────────────────────
	// Before inserting, run the same scoring logic used after insert. This
	// catches templates that pass the structural Check 1/2/3 above but have
	// deeper problems: zero template variables despite a populated schema,
	// schema-template field mismatch, malformed placeholder syntax.
	// Rejecting here prevents broken components entering the DB and being
	// used on pages before the quality auditor catches them.
	//
	// The scoring is a pure function (no DB access) so running it twice —
	// once here, once after INSERT — has no side effects on the first call.
	schemaJSONStr := string(inputSchemaJSON)
	preStoreScore := scoreComponent("", functionName, htmlTemplate, schemaJSONStr, "section")

	// Reject on structural problems that make the component unusable.
	// These are the same conditions that produced quality_score=30 on
	// provocation-feed and archetype-combinations (2026-04-17).
	blockingIssues := []string{}
	if !preStoreScore.TemplateClosed {
		blockingIssues = append(blockingIssues, "template not closed properly")
	}
	if !preStoreScore.HasDataComponent {
		blockingIssues = append(blockingIssues, "missing data-component attribute")
	}
	if preStoreScore.TemplateVariableCount == 0 && preStoreScore.SchemaFieldCount > 0 {
		blockingIssues = append(blockingIssues, fmt.Sprintf(
			"template has 0 {{.var}} placeholders but schema declares %d fields — content would be unreachable",
			preStoreScore.SchemaFieldCount))
	}
	if !preStoreScore.SchemaTemplateSynced && preStoreScore.TemplateVariableCount > 0 {
		blockingIssues = append(blockingIssues, "template variables and schema fields do not match")
	}

	// Substantive template with no placeholders at all. Catches the case
	// the prior conditions miss: both template AND schema are populated
	// with HTML/CSS but no {{placeholder "..."}} tokens exist. Such a
	// template can render only static markup — no LLM content can be
	// injected, no static fallbacks can be substituted. The threshold
	// (500 chars) excludes legitimately tiny utility components like
	// dividers or spacers.
	const substantiveTemplateThreshold = 500
	if preStoreScore.TemplateVariableCount == 0 && len(htmlTemplate) > substantiveTemplateThreshold {
		blockingIssues = append(blockingIssues, fmt.Sprintf(
			"template is %d chars but has 0 {{placeholder \"...\"}} tokens — no content path exists",
			len(htmlTemplate)))
	}

	// Literal "<no value>" strings in the template are Go text/template
	// render artifacts: the template was rendered against an empty data
	// context (or with default missing-key handling), the unresolved
	// variables produced "<no value>", and that output was stored back
	// as the template. Such a template is permanently broken — the
	// placeholders are gone and can't be restored without regeneration.
	if strings.Contains(htmlTemplate, "<no value>") {
		blockingIssues = append(blockingIssues, fmt.Sprintf(
			"template contains %d '<no value>' artifacts — Go template render output mistakenly stored as source",
			strings.Count(htmlTemplate, "<no value>")))
	}

	// Regeneration must not break the field-name contract that existing
	// dependents' content_data is keyed on. content_data is written to
	// match the component's input_schema field names at build time;
	// renaming or removing a retained field strands that stored content —
	// the renamed placeholder no longer matches the stored key, and
	// RenderTemplate silently strips the unmatched placeholder to "", so
	// the section renders empty with no error (this is the fdd92ad4
	// system-stats failure: e.g. stat_1_number→stat1_value, eyebrow→
	// eyebrow_label renamed in place while dependents' content_data stayed
	// on the old keys). Adding new fields is fine; dropping/renaming an
	// existing one is the damage, so block it here rather than overwrite
	// the shared row and silently empty every dependent. Intentional
	// field-set changes to a shared component must go through a deliberate
	// migration, not an LLM regeneration side effect.
	//
	// This compares old-schema fields to new-schema fields as a proxy for
	// "what dependents have" (content_data is written to match the schema).
	// If exactness is ever needed, swap to querying the affected
	// page_components for the union of their content_data keys.
	if isRegeneration {
		oldFields := schemaFieldSet(existingSchema)
		newFields := schemaFieldSet(schemaJSONStr)
		var stranded []string
		for name := range oldFields {
			if !newFields[name] {
				stranded = append(stranded, name)
			}
		}
		if len(stranded) > 0 {
			sort.Strings(stranded) // deterministic message ordering
			blockingIssues = append(blockingIssues, fmt.Sprintf(
				"regeneration removes/renames %d existing schema field(s) (%s) that dependents' content_data is keyed on — overwriting would strand stored content and render those sections empty; preserve these field names or migrate dependents explicitly",
				len(stranded), strings.Join(stranded, ", ")))
		}
	}

	if len(blockingIssues) > 0 {
		// Persist a structured rejection record to agent_error_log. This
		// makes validation failures queryable across the system — we can
		// run analytics on "which fields does the LLM most often forget
		// to declare?" or "which functions repeatedly fail Direction 2?"
		// without trawling kubectl logs.
		//
		// The act of writing here is best-effort (best handled inside the
		// helper). The next return is the actual rejection.
		recordValidationRejection(
			ctx, params.DB, logger, params,
			functionName, sectionType,
			htmlTemplate, schemaJSONStr,
			preStoreScore, blockingIssues,
		)

		logger.Warn("store_generated_component: rejecting low-quality template",
			zap.String("function", functionName),
			zap.String("section_type", sectionType),
			zap.Int("pre_store_score", preStoreScore.QualityScore),
			zap.Int("template_variable_count", preStoreScore.TemplateVariableCount),
			zap.Int("schema_field_count", preStoreScore.SchemaFieldCount),
			zap.Strings("blocking_issues", blockingIssues),
			zap.Strings("all_issues", preStoreScore.QualityIssues),
		)
		return nil, fmt.Errorf(
			"generated template for %q rejected by pre-store validation: %s",
			sectionType, strings.Join(blockingIssues, "; "))
	}

	logger.Info("store_generated_component: pre-store validation passed",
		zap.String("function", functionName),
		zap.Int("pre_store_score", preStoreScore.QualityScore),
		zap.Int("template_variable_count", preStoreScore.TemplateVariableCount),
		zap.Int("schema_field_count", preStoreScore.SchemaFieldCount),
	)

	// ── Write to DB: UPDATE (regeneration) or INSERT (creation) ─────────
	// Both paths end with scoring + markPagesForRebuild so those stay below.
	var componentID string
	var status string
	var regenPagesMarked int64
	var newVersion int // populated on regeneration; 0 for creation

	if isRegeneration {
		// Snapshot current state to component_versions BEFORE the UPDATE.
		// Best-effort: a failed snapshot logs Warn but does not block the
		// UPDATE — losing history is recoverable; leaving a broken template
		// in place is not.
		newVersion = existingVersion + 1
		snapshotErr := snapshotComponentVersion(
			ctx, params.DB, existingID, newVersion,
			existingHTML, existingSchema,
			nullStringToGo(existingJS),
			"Regenerated by component-creator",
			"component-creator:regen",
			workItemSource,
			logger,
		)
		if snapshotErr != nil {
			logger.Warn("store_generated_component: version snapshot failed, continuing with UPDATE",
				zap.String("component_id", existingID),
				zap.Int("intended_version", newVersion),
				zap.Error(snapshotErr))
			// newVersion still advances — even if we couldn't write the
			// snapshot, we don't want to overwrite version N later and
			// create a gap. The post-UPDATE return reflects what we tried.
		}

		// UPDATE in place: preserves component_id so all foreign key
		// references (page_components, site_components, link_registry,
		// etc.) keep resolving without any relink step.
		result, err := params.DB.ExecContext(ctx, `
			UPDATE content_components
			SET html_template   = $1,
			    input_schema    = $2::jsonb,
			    js_content      = $3,
			    is_dark_section = $4,
			    render_mode     = $5,
			    is_active       = true,
			    updated_at      = NOW()
			WHERE id = $6::uuid
		`,
			htmlTemplate,
			inputSchemaJSON,
			nullIfEmpty(jsContent),
			isDark,
			deriveRenderMode(inputSchemaJSON), // derived from schema, not hardcoded
			existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing component during regeneration: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			return nil, fmt.Errorf("regeneration UPDATE affected %d rows (expected 1) for component %s",
				rowsAffected, existingID)
		}
		componentID = existingID
		status = "regenerated"

		// Mark dependent page_components pending so the rerender pipeline
		// rebuilds them against the new template. We do NOT overwrite
		// rendered_html — that field holds the last-good render per page
		// and needs per-page variable substitution to regenerate correctly.
		var affectedSiteIDs []string
		regenPagesMarked, affectedSiteIDs = markPagesPendingRebuild(ctx, params.DB, existingID, logger)

		// Raise one needs_rerender work item per affected site so the
		// rerender-pages handler actually regenerates. Without this the
		// build_status=pending flag is informational only — nothing
		// downstream scans page_components for pending rows.
		//
		// The dedup index (site_id, item_key) with item_key scoped by
		// component_id prevents duplicates if the same regen runs twice.
		rerenderItemsCreated := 0
		for _, siteID := range affectedSiteIDs {
			if created := createRerenderWorkItem(
				ctx, params.DB, siteID, existingID, functionName, workItemSource, logger,
			); created {
				rerenderItemsCreated++
			}
		}

		logger.Info("store_generated_component: component regenerated",
			zap.String("component_id", componentID),
			zap.String("function", functionName),
			zap.Int("previous_version", existingVersion),
			zap.Int("new_version", newVersion),
			zap.Int64("pages_marked_rebuild", regenPagesMarked),
			zap.Int("affected_sites", len(affectedSiteIDs)),
			zap.Int("rerender_items_created", rerenderItemsCreated))
	} else {
		// Creation path — unchanged INSERT block.
		err = params.DB.QueryRowContext(ctx, `
			INSERT INTO content_components (
				name, display_name, function, category, component_level,
				section_type, suitable_site_types, suitable_page_types,
				description, html_template, js_content, input_schema,
				is_dark_section, render_mode, created_from, is_active,
				usage_count, avg_quality_score,
				semantic_tags
			) VALUES (
				$1, $2, $3, $4, 'section',
				$5, $6::jsonb, $7::jsonb,
				$8, $9, $10, $11::jsonb,
				$12, $13, 'generated', true,
				0, NULL,
				$14::jsonb
			)
			RETURNING id::text
		`,
			functionName,                      // $1 name
			displayName,                       // $2 display_name
			functionName,                      // $3 function
			category,                          // $4 category
			sectionType,                       // $5 section_type
			string(suitableSiteTypesJSON),     // $6 suitable_site_types
			string(suitablePageTypesJSON),     // $7 suitable_page_types
			description,                       // $8 description
			htmlTemplate,                      // $9 html_template (JS extracted)
			nullIfEmpty(jsContent),            // $10 js_content (NULL if no JS)
			inputSchemaJSON,                   // $11 input_schema
			isDark,                            // $12 is_dark_section
			deriveRenderMode(inputSchemaJSON), // $13 render_mode (derived from schema, not hardcoded)
			datahelpers.BuildSemanticTags(sectionType, siteType), // $14 semantic_tags
		).Scan(&componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert component: %w", err)
		}
		status = "created"

		// Snapshot version 1 so history is complete from creation onward.
		// Best-effort — a snapshot failure here doesn't undo the INSERT.
		if err := snapshotComponentVersion(
			ctx, params.DB, componentID, 1,
			htmlTemplate, inputSchemaJSON,
			jsContent,
			"Initial version — created by component-creator",
			"component-creator:create",
			workItemSource,
			logger,
		); err != nil {
			logger.Warn("store_generated_component: initial version snapshot failed, continuing",
				zap.String("component_id", componentID),
				zap.Error(err))
		} else {
			newVersion = 1
		}

		logger.Info("store_generated_component: component created",
			zap.String("component_id", componentID),
			zap.String("function", functionName),
			zap.String("section_type", sectionType))
	}

	// Score the resulting row (both paths). Persists to content_components.
	qualityResult := ScoreAndPersistComponent(
		ctx, params.DB,
		componentID, functionName, htmlTemplate, schemaJSONStr, "section",
		logger,
	)

	// Mark pages that were waiting for this section_type as needs_rebuild.
	// When plan_sections can't find a component, the page stays deployed
	// with a gap. Now that the component exists (or has been regenerated),
	// those pages should rebuild.
	markPagesForRebuild(ctx, params.DB, sectionType, logger)

	response := map[string]interface{}{
		"component_id":   componentID,
		"function":       functionName,
		"section_type":   sectionType,
		"display_name":   displayName,
		"category":       category,
		"status":         status,
		"template_size":  len(htmlTemplate),
		"has_js":         jsContent != "",
		"js_size":        len(jsContent),
		"quality_score":  qualityResult.QualityScore,
		"quality_issues": qualityResult.QualityIssues,
	}
	if isRegeneration {
		response["previous_version"] = existingVersion
		response["new_version"] = newVersion
		response["pages_marked_rebuild"] = regenPagesMarked
	} else if newVersion > 0 {
		response["new_version"] = newVersion
	}
	return response, nil
}

// parseGeneratedTemplate extracts html_template, input_schema, and metadata
// from the LLM output. The LLM is instructed to return JSON with these fields,
// but we handle various formats defensively.
func parseGeneratedTemplate(raw interface{}, sectionType string, logger *zap.Logger) (
	htmlTemplate string, inputSchemaJSON string, functionName string, isDark bool, err error,
) {
	if raw == nil {
		return "", "{}", "", false, fmt.Errorf("generated template is nil")
	}

	// The LLM output might be:
	// 1. A map with "result" containing the JSON string (from execute_llm_prompt)
	// 2. A map with html_template/input_schema directly
	// 3. A string containing JSON (possibly wrapped in markdown code blocks)

	var data map[string]interface{}

	switch v := raw.(type) {
	case map[string]interface{}:
		// Check for execute_llm_prompt's "result" wrapper
		if result, ok := v["result"]; ok {
			switch r := result.(type) {
			case string:
				data = parseJSONStringToMap(r, logger)
			case map[string]interface{}:
				data = r
			}
		} else {
			data = v
		}
	case string:
		data = parseJSONStringToMap(v, logger)
	default:
		return "", "{}", "", false, fmt.Errorf("unexpected type for generated template: %T", raw)
	}

	// Extract html_template
	if ht, ok := data["html_template"].(string); ok {
		htmlTemplate = strings.TrimSpace(ht)
	}

	// Safety check: if htmlTemplate still looks like a JSON blob wrapping an
	// html_template field, extract the actual HTML. This catches cases where
	// json.Unmarshal failed and the fallback stored the entire JSON string.
	htmlTemplate = unwrapJSONBlobIfNeeded(htmlTemplate, logger)

	// Extract input_schema
	inputSchemaJSON = "{}"
	if schema, ok := data["input_schema"]; ok {
		if schemaMap, ok := schema.(map[string]interface{}); ok {
			schemaBytes, _ := json.Marshal(schemaMap)
			inputSchemaJSON = string(schemaBytes)
		} else if schemaStr, ok := schema.(string); ok {
			inputSchemaJSON = schemaStr
		}
	}

	// Extract or derive function name
	if fn, ok := data["function"].(string); ok && fn != "" {
		functionName = fn
	} else {
		// Derive from section_type — prefix with a category hint
		functionName = sectionType
	}

	// Validate kebab-case
	functionName = datahelpers.NormaliseToKebab(functionName)

	// Extract is_dark_section
	if dark, ok := data["is_dark_section"].(bool); ok {
		isDark = dark
	}

	return htmlTemplate, inputSchemaJSON, functionName, isDark, nil
}

// parseJSONStringToMap takes a raw string (possibly with markdown code blocks)
// and tries to parse it as a JSON map. If standard parsing fails, it falls back
// to field-level extraction from broken JSON. If the string is not JSON at all,
// it treats it as raw HTML.
//
// Uses datahelpers.StripCodeFences (shared with content_search, create_tool, etc.)
// and datahelpers.SafeUnmarshalString for safe parsing.
func parseJSONStringToMap(s string, logger *zap.Logger) map[string]interface{} {
	// Strip markdown code blocks using the shared helper from datahelpers
	cleaned := datahelpers.StripCodeFences(s)

	// Try standard JSON parse using the shared SafeUnmarshalString
	var data map[string]interface{}
	if datahelpers.SafeUnmarshalString(cleaned, &data) {
		logger.Info("store_generated_component: parsed LLM output as JSON",
			zap.Int("fields", len(data)))
		return data
	}

	// JSON parse failed — try field-level extraction from broken JSON.
	// LLMs often produce JSON with unescaped characters in HTML/SVG content
	// that breaks json.Unmarshal, but the structure is still recoverable.
	if strings.Contains(cleaned, `"html_template"`) {
		logger.Info("store_generated_component: json.Unmarshal failed, attempting field extraction",
			zap.Int("length", len(cleaned)),
			zap.String("first_80", truncateStr(cleaned, 80)))

		result := map[string]interface{}{}

		if ht, ok := extractJSONStringField(cleaned, "html_template"); ok {
			result["html_template"] = ht
			logger.Info("store_generated_component: extracted html_template from broken JSON",
				zap.Int("length", len(ht)),
				zap.String("first_40", truncateStr(ht, 40)))
		}
		if fn, ok := extractJSONStringField(cleaned, "function"); ok {
			result["function"] = fn
		}
		// is_dark_section is a bool, not a string — check with simple contains
		if strings.Contains(cleaned, `"is_dark_section": true`) || strings.Contains(cleaned, `"is_dark_section":true`) {
			result["is_dark_section"] = true
		}

		if _, hasHTML := result["html_template"]; hasHTML {
			return result
		}
	}

	// Not JSON at all — treat as raw HTML
	logger.Info("store_generated_component: treating as raw HTML",
		zap.Int("length", len(cleaned)),
		zap.String("first_50", truncateStr(cleaned, 50)))
	return map[string]interface{}{
		"html_template": cleaned,
	}
}

// extractJSONStringField extracts the value of a string field from a JSON-like
// string, even when json.Unmarshal fails due to unescaped characters elsewhere.
// It manually scans the JSON string value, handling standard escape sequences.
func extractJSONStringField(s, fieldName string) (string, bool) {
	key := `"` + fieldName + `"`
	keyIdx := strings.Index(s, key)
	if keyIdx == -1 {
		return "", false
	}

	// Find the colon after the key
	rest := s[keyIdx+len(key):]
	colonIdx := strings.IndexByte(rest, ':')
	if colonIdx == -1 {
		return "", false
	}
	rest = rest[colonIdx+1:]

	// Skip whitespace
	rest = strings.TrimLeft(rest, " \t\n\r")

	// Must start with a quote
	if len(rest) == 0 || rest[0] != '"' {
		return "", false
	}
	rest = rest[1:] // skip opening quote

	// Scan for the closing unescaped quote, handling escape sequences
	var result strings.Builder
	result.Grow(len(rest))
	i := 0
	for i < len(rest) {
		if rest[i] == '\\' && i+1 < len(rest) {
			// JSON escape sequence
			switch rest[i+1] {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case '/':
				result.WriteByte('/')
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case 'b':
				result.WriteByte('\b')
			case 'f':
				result.WriteByte('\f')
			default:
				// Unknown escape — preserve as-is
				result.WriteByte(rest[i])
				result.WriteByte(rest[i+1])
			}
			i += 2
		} else if rest[i] == '"' {
			// Unescaped closing quote — end of field value
			return result.String(), true
		} else {
			result.WriteByte(rest[i])
			i++
		}
	}

	// No closing quote found. The JSON is severely broken, but we likely have
	// the content up to the end of the string. Return what we collected.
	extracted := result.String()
	if len(extracted) > 0 {
		return strings.TrimSpace(extracted), true
	}
	return "", false
}

// unwrapJSONBlobIfNeeded checks if a string looks like it's still a JSON blob
// wrapping an html_template field, and extracts the actual HTML if so.
// This is a safety net that catches any case where earlier parsing failed.
func unwrapJSONBlobIfNeeded(s string, logger *zap.Logger) string {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "{") {
		return s // Not a JSON blob
	}
	if !strings.Contains(trimmed, `"html_template"`) {
		return s // Doesn't contain the field
	}

	logger.Info("store_generated_component: unwrapping JSON blob from html_template",
		zap.Int("length", len(trimmed)),
		zap.String("first_60", truncateStr(trimmed, 60)))

	// Try standard JSON parse first using shared SafeUnmarshalString
	var wrapper map[string]interface{}
	if datahelpers.SafeUnmarshalString(trimmed, &wrapper) {
		if ht, ok := wrapper["html_template"].(string); ok && strings.TrimSpace(ht) != "" {
			return strings.TrimSpace(ht)
		}
	}

	// Standard parse failed — use field extraction
	if ht, ok := extractJSONStringField(trimmed, "html_template"); ok && ht != "" {
		return ht
	}

	return s // Couldn't unwrap — return as-is
}

// markPagesForRebuild finds deployed pages whose sections array references
// the given section_type and marks them for rebuild. This closes the loop
// when a component is created after plan_sections already ran and deferred
// the section.
func markPagesForRebuild(ctx context.Context, db *sql.DB, sectionType string, logger *zap.Logger) {
	res, err := db.ExecContext(ctx, `
		UPDATE pages SET build_status = 'needs_rebuild', updated_at = NOW()
		WHERE status = 'active'
		  AND build_status = 'deployed'
		  AND EXISTS (
		      SELECT 1 FROM jsonb_array_elements_text(sections) sec
		      WHERE sec = $1
		  )
	`, sectionType)
	if err != nil {
		logger.Warn("store_generated_component: failed to mark pages for rebuild",
			zap.String("section_type", sectionType),
			zap.Error(err))
		return
	}
	if rows, _ := res.RowsAffected(); rows > 0 {
		logger.Info("store_generated_component: marked pages for rebuild",
			zap.String("section_type", sectionType),
			zap.Int64("pages_marked", rows))
	}
}

// truncateStr returns the first n characters of s, appending "..." if truncated.
// Named truncateStr to avoid conflict with any future stdlib truncate.
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// separateInlineJS extracts inline <script> blocks from an HTML template,
// stores them separately, and replaces with a <script src> reference.
//
// Only extracts <script> tags WITHOUT attributes (inline JS).
// Leaves <script src="...">, <script type="module">, etc. untouched.
//
// Multiple inline script blocks are combined into a single JS content string.
// If no inline scripts found, returns the template unchanged and empty jsContent.
func separateInlineJS(htmlTemplate, functionName string) (cleanHTML, jsContent string) {
	// Match <script> tags with no attributes — these contain inline JS.
	// (?s) enables dot-matches-newline.
	re := regexp.MustCompile(`(?s)<script\s*>(.*?)</script>`)

	var jsBlocks []string
	hasInlineJS := false

	cleanHTML = re.ReplaceAllStringFunc(htmlTemplate, func(match string) string {
		// Safety check: skip if somehow a src= tag matched
		trimmed := strings.TrimSpace(match)
		if len(trimmed) > 20 && strings.Contains(trimmed[:20], "src=") {
			return match
		}

		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		js := strings.TrimSpace(submatch[1])
		if js == "" {
			return "" // empty script tag, just remove
		}

		jsBlocks = append(jsBlocks, js)
		hasInlineJS = true
		return "" // remove the inline script block from HTML
	})

	if !hasInlineJS {
		return htmlTemplate, ""
	}

	// Combine all JS blocks
	jsContent = strings.Join(jsBlocks, "\n\n")

	// Add the <script src> reference after </section>
	scriptRef := fmt.Sprintf(`<script src="/tools/assets/%s.js"></script>`, functionName)

	if idx := strings.LastIndex(cleanHTML, "</section>"); idx >= 0 {
		insertAt := idx + len("</section>")
		cleanHTML = cleanHTML[:insertAt] + "\n" + scriptRef + cleanHTML[insertAt:]
	} else {
		cleanHTML = cleanHTML + "\n" + scriptRef
	}

	// Clean up double blank lines left by removed script blocks
	for strings.Contains(cleanHTML, "\n\n\n") {
		cleanHTML = strings.ReplaceAll(cleanHTML, "\n\n\n", "\n\n")
	}

	return cleanHTML, jsContent
}

// ---------------------------------------------------------------------------
// Regeneration helpers
// ---------------------------------------------------------------------------

// snapshotComponentVersion writes the pre-update state of a content_component
// into component_versions so we have history for rollback, diffing, or
// allowing other sites to opt back to an earlier version.
//
// Uses live schema columns: version_number, change_description, changed_by,
// change_source, plus html_template, input_schema, css_template.
// css_template is left NULL for now — the section components store
// everything (CSS + HTML) in html_template and the separate css_template
// column isn't used by the current generator.
//
// changedBy identifies the agent/principal making the change
// (e.g. "component-creator:regen", "tool-improver:auto").
// changeSource identifies the triggering work item or event
// (e.g. the work item's source field, "manual_regen_after_prompt_fix",
// "component-quality-auditor"). May be "" if there is no originating work
// item — the column is nullable, and the helper writes NULL in that case.
//
// The caller is expected to precompute versionNumber as
// MAX(version_number)+1 for this component_id. The unique index
// (component_id, version_number) will reject duplicates, so if two concurrent
// regenerations both compute the same MAX+1, one will fail here and the
// caller logs+continues per best-effort policy.
//
// Returns non-nil error on failure. The caller treats snapshot as
// best-effort: an error here should be logged but not block the UPDATE.
func snapshotComponentVersion(
	ctx context.Context,
	db *sql.DB,
	componentID string,
	versionNumber int,
	htmlTemplate string,
	inputSchemaJSON string,
	jsContent string, // currently not stored in component_versions; passed for future compatibility
	changeDescription string,
	changedBy string,
	changeSource string,
	logger *zap.Logger,
) error {
	_ = jsContent // reserved for when component_versions grows a js_content column

	var changeSourceArg interface{}
	if changeSource == "" {
		changeSourceArg = nil
	} else {
		changeSourceArg = changeSource
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO component_versions (
			component_id, version_number,
			html_template, input_schema,
			change_description, changed_by, change_source,
			created_at
		) VALUES (
			$1::uuid, $2,
			$3, $4::jsonb,
			$5, $6, $7,
			NOW()
		)
	`,
		componentID,
		versionNumber,
		htmlTemplate,
		inputSchemaJSON,
		changeDescription,
		changedBy,
		changeSourceArg,
	)
	if err != nil {
		return fmt.Errorf("insert component_versions (component=%s, version=%d): %w",
			componentID, versionNumber, err)
	}

	logger.Info("store_generated_component: version snapshot written",
		zap.String("component_id", componentID),
		zap.Int("version_number", versionNumber),
		zap.String("changed_by", changedBy),
		zap.String("change_source", changeSource))
	return nil
}

// markPagesPendingRebuild flips build_status to 'pending' for all
// page_components that use this component_id. Does NOT touch rendered_html
// (that's per-page content; the rerender pipeline will regenerate it with
// variable substitution). Returns the count of pages marked AND the
// distinct site_ids they belong to, so the caller can raise one
// needs_rerender work item per affected site.
//
// Failures are logged but not returned — the UPDATE to content_components
// has already succeeded, and page rebuild eligibility is something the
// auditor can re-check independently. Returns (0, nil) on failure.
func markPagesPendingRebuild(
	ctx context.Context,
	db *sql.DB,
	componentID string,
	logger *zap.Logger,
) (pagesMarked int64, affectedSiteIDs []string) {
	// First, collect the distinct site_ids that will be affected, via the
	// join from page_components → pages. Do this BEFORE the UPDATE so we
	// have a stable set of sites to raise rerender items for.
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.site_id::text
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.component_id = $1::uuid
	`, componentID)
	if err != nil {
		logger.Warn("markPagesPendingRebuild: failed to enumerate affected sites",
			zap.String("component_id", componentID),
			zap.Error(err))
		// Still try the UPDATE — pages can be marked pending even if we
		// can't raise rerender items; the auditor will catch them later.
	} else {
		for rows.Next() {
			var sid string
			if scanErr := rows.Scan(&sid); scanErr == nil && sid != "" {
				affectedSiteIDs = append(affectedSiteIDs, sid)
			}
		}
		rows.Close()
	}

	result, err := db.ExecContext(ctx, `
		UPDATE page_components
		SET build_status = 'pending', updated_at = NOW()
		WHERE component_id = $1::uuid
	`, componentID)
	if err != nil {
		logger.Warn("markPagesPendingRebuild: UPDATE failed, page_components left in current state",
			zap.String("component_id", componentID),
			zap.Error(err))
		return 0, affectedSiteIDs
	}
	pagesMarked, _ = result.RowsAffected()
	if pagesMarked > 0 {
		logger.Info("markPagesPendingRebuild: flagged pages for rebuild",
			zap.String("component_id", componentID),
			zap.Int64("pages", pagesMarked),
			zap.Int("sites", len(affectedSiteIDs)))
	}
	return pagesMarked, affectedSiteIDs
}

// nullStringToGo converts a sql.NullString to a plain string, treating
// NULL as "". Keeps call sites readable when we don't care about the
// null-vs-empty distinction.
func nullStringToGo(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

// createRerenderWorkItem inserts a needs_rerender work item for one site,
// scoped to a specific regenerated component. The rerender-pages handler
// picks up these items and rebuilds affected pages.
//
// The item_key is component-scoped (component_regen_rerender:<uuid>) so
// that:
//   - multiple concurrent regens of DIFFERENT components produce
//     distinct work items even within the same site
//   - repeat regens of the SAME component collide with the dedup index
//     idx_swi_dedup (site_id, item_key) and the INSERT is a no-op
//     (excluding completed/failed rows — see index WHERE clause)
//
// Returns true if a row was actually inserted, false if the dedup
// ON CONFLICT path was taken or an error was logged.
//
// Errors are logged but never propagate — an orphaned pending page is
// recoverable by the auditor, and blocking the whole regen on a work
// item write would waste the UPDATE we just completed.
//
// workItemSource is the originating work item's source field, used as
// the `source` column so this synthetic rerender item can be traced back
// to whatever caused the regen. Empty string is safe — `source` has a
// NOT NULL constraint in site_work_items, so we substitute a default
// of "component-creator" when the caller passes "".
func createRerenderWorkItem(
	ctx context.Context,
	db *sql.DB,
	siteID string,
	componentID string,
	functionName string,
	workItemSource string,
	logger *zap.Logger,
) bool {
	sourceField := workItemSource
	if sourceField == "" {
		sourceField = "component-creator"
	}

	itemKey := fmt.Sprintf("component_regen_rerender:%s", componentID)
	summary := fmt.Sprintf("Re-render pages after %s regeneration", functionName)
	// reason=section_data_resolved so the per-page rerender items (created by
	// rerender-pages' create_rerender_items, once it propagates spec.reason)
	// drive a section re-render of this component's dependents rather than an
	// assemble-only re-ship. component_id above is what scopes it to those
	// dependents.
	specJSON := fmt.Sprintf(
		`{"component_id": %q, "function": %q, "reason": "section_data_resolved", "refresh_site_components": false}`,
		componentID, functionName,
	)

	// Insert with a guard that mirrors the dedup index's WHERE clause.
	// The guard on the INSERT is redundant given the unique index but
	// makes the intent readable and avoids a unique-violation error
	// path that would need error-sniffing.
	result, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity,
			summary, priority, handler_agent, status, created_by,
			spec, item_key
		)
		SELECT $1::uuid, $2, 'build', 'needs_rerender', 'medium',
		       $3, 99, 'rerender-pages', 'triaged', 'store_generated_component',
		       $4::jsonb, $5
		WHERE NOT EXISTS (
			SELECT 1 FROM site_work_items
			WHERE site_id = $1::uuid
			  AND item_key = $5
			  AND status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed')
		)
	`,
		siteID,
		sourceField,
		summary,
		specJSON,
		itemKey,
	)
	if err != nil {
		logger.Warn("createRerenderWorkItem: INSERT failed, site will rely on auditor to catch pending pages",
			zap.String("site_id", siteID),
			zap.String("component_id", componentID),
			zap.Error(err))
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Info("createRerenderWorkItem: rerender item already pending for this component/site (dedup)",
			zap.String("site_id", siteID),
			zap.String("component_id", componentID),
			zap.String("item_key", itemKey))
		return false
	}

	logger.Info("createRerenderWorkItem: raised needs_rerender work item",
		zap.String("site_id", siteID),
		zap.String("component_id", componentID),
		zap.String("function", functionName),
		zap.String("item_key", itemKey),
		zap.String("source", sourceField))
	return true
}

// ---------------------------------------------------------------------------
// Validation rejection logging
// ---------------------------------------------------------------------------

// orphanSchemaFieldPattern extracts the field name from a Direction-2
// sync issue like:  schema field "card_link_label" has no template variable
var orphanSchemaFieldPattern = regexp.MustCompile(`^schema field "([^"]+)" has no template variable$`)

// unknownTemplateVarPattern extracts the field name from a Direction-1
// sync issue like:  template var {{.cta_link}} has no schema entry
var unknownTemplateVarPattern = regexp.MustCompile(`^template var \{\{\.([^}]+)\}\} has no schema entry$`)

// recordValidationRejection writes a structured row to agent_error_log
// when pre-store validation rejects an LLM-generated component. This
// gives us a queryable trail of which fields the LLM keeps getting
// wrong, separable from the rest of the chassis log noise.
//
// Best-effort: failures inside this helper are logged at warn level but
// do not affect the caller's return path. The action still returns the
// same rejection error to its caller.
//
// Severity mapping:
//   - "warning"  — Direction-2 bookkeeping mismatch (schema declares a
//     field the template doesn't use, or vice versa). The
//     LLM produced something structurally well-formed but
//     failed list-reconciliation. Common, addressable.
//   - "error"    — Structural failures (template not closed, missing
//     data-component, "<no value>" artifacts, 0-placeholder
//     substantive template). These indicate the LLM
//     produced something broken at a deeper level.
func recordValidationRejection(
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
	params ActionParams,
	functionName string,
	sectionType string,
	htmlTemplate string,
	schemaJSON string,
	score ComponentQualityResult,
	blockingIssues []string,
) {
	if db == nil {
		return
	}

	// Classify issues into orphan-field (bookkeeping) vs other.
	orphanSchemaFields := []string{}
	unknownTemplateVars := []string{}
	otherIssues := []string{}
	for _, issue := range score.QualityIssues {
		if m := orphanSchemaFieldPattern.FindStringSubmatch(issue); len(m) > 1 {
			orphanSchemaFields = append(orphanSchemaFields, m[1])
			continue
		}
		if m := unknownTemplateVarPattern.FindStringSubmatch(issue); len(m) > 1 {
			unknownTemplateVars = append(unknownTemplateVars, m[1])
			continue
		}
		otherIssues = append(otherIssues, issue)
	}

	// Severity: bookkeeping-only failures are warning; anything else is error.
	severity := "warning"
	if len(otherIssues) > 0 || len(unknownTemplateVars) > 0 {
		severity = "error"
	}
	// Severity error if any structural blockers are present (not closed,
	// missing data-component, no-value artifacts, 0-var substantive).
	for _, b := range blockingIssues {
		if strings.Contains(b, "not closed") ||
			strings.Contains(b, "missing data-component") ||
			strings.Contains(b, "<no value>") ||
			strings.Contains(b, "0 {{.var}} placeholders") ||
			strings.Contains(b, "0 {{placeholder") {
			severity = "error"
			break
		}
	}

	// Pull context from the action params.
	workItemID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id")
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id")
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain")

	contextPayload := map[string]interface{}{
		"function":                functionName,
		"section_type":            sectionType,
		"pre_store_score":         score.QualityScore,
		"template_variable_count": score.TemplateVariableCount,
		"schema_field_count":      score.SchemaFieldCount,
		"schema_template_synced":  score.SchemaTemplateSynced,
		"template_closed":         score.TemplateClosed,
		"has_data_component":      score.HasDataComponent,
		"template_len":            len(htmlTemplate),
		"schema_len":              len(schemaJSON),
		"orphan_schema_fields":    orphanSchemaFields,
		"unknown_template_vars":   unknownTemplateVars,
		"other_issues":            otherIssues,
		"blocking_issues":         blockingIssues,
		"all_issues":              score.QualityIssues,
	}
	contextJSON, _ := json.Marshal(contextPayload)
	if contextJSON == nil {
		contextJSON = []byte("{}")
	}

	errorMessage := fmt.Sprintf(
		"component validation rejected for function=%q section_type=%q: %s",
		functionName, sectionType, strings.Join(blockingIssues, "; "))

	errorCode := "component_validation_rejected"
	if len(orphanSchemaFields) > 0 && len(otherIssues) == 0 && len(unknownTemplateVars) == 0 {
		errorCode = "component_validation_orphan_schema_field"
	}
	if len(unknownTemplateVars) > 0 {
		errorCode = "component_validation_unknown_template_var"
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO agent_error_log (
			site_id, domain, work_item_id, orchestration_id,
			agent_type, agent_id, pod_name, step_name, action,
			error_message, error_code, severity, context
		) VALUES (
			NULLIF($1, '')::uuid, $2, NULLIF($3, '')::uuid, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13::jsonb
		)
	`,
		siteID,
		domain,
		workItemID,
		params.ExecutionContext.OrchestrationID,
		"component-creator",
		params.ExecutionContext.Sender.AgentID,
		params.ExecutionContext.Sender.PodName,
		"store_component",
		"store_generated_component",
		errorMessage,
		errorCode,
		severity,
		string(contextJSON),
	)
	if err != nil {
		logger.Warn("recordValidationRejection: failed to write to agent_error_log",
			zap.Error(err),
			zap.String("function", functionName))
	}
}

// deriveRenderMode inspects a JSON-encoded input_schema and returns "agent"
// if any field has source="llm", otherwise "template".
//
// This ensures render_mode is always consistent with the schema rather than
// being hardcoded at creation time. The page-content-writer workflow's
// check_render_mode conditional routes sections to LLM generation when
// render_mode == "agent"; without this derivation every component would
// permanently take the template-only path regardless of its content needs.
//
// Called by both the INSERT (creation) and UPDATE (regeneration) paths in
// StoreGeneratedComponentAction so the value is always up to date.
func deriveRenderMode(inputSchemaJSON string) string {
	if inputSchemaJSON == "" || inputSchemaJSON == "{}" {
		return "template"
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(inputSchemaJSON), &schema); err != nil {
		return "template"
	}

	fields, ok := schema["fields"].(map[string]interface{})
	if !ok {
		return "template"
	}

	for _, v := range fields {
		fieldDef, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if source, ok := fieldDef["source"].(string); ok && source == "llm" {
			return "agent"
		}
	}

	return "template"
}

// schemaFieldSet returns the set of field names declared under
// input_schema.fields. Empty or invalid schema → empty set. Mirrors the
// fields-parse used by deriveRenderMode.
func schemaFieldSet(inputSchemaJSON string) map[string]bool {
	out := map[string]bool{}
	if inputSchemaJSON == "" || inputSchemaJSON == "{}" {
		return out
	}
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(inputSchemaJSON), &schema); err != nil {
		return out
	}
	fields, ok := schema["fields"].(map[string]interface{})
	if !ok {
		return out
	}
	for name := range fields {
		out[name] = true
	}
	return out
}
