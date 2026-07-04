// FILE: platform/orchestration/actions/append_doc_note_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. Requires migration 0NN_doc_plans_and_notes.sql.
//
// append_doc_note — one INSERT into doc_notes (no read-modify-write; concurrent
// appends are safe). Called as the LAST step of the fix agents (tool-improver /
// update_component_html / component-template-fixer), by workflow-altering
// migrations (pipeline notes), and via persist_diagnosis_note. Uses
// docResolveSubject from write_doc_plan_action.go and nullIfEmpty from the
// package helpers (rag_actions.go).
//
// Registration (add to registry.go):
//   "append_doc_note": {
//       Handler:     AppendDocNoteAction,
//       Category:    "documentation",
//       Description: "Append one NOTES entry (row) for a tool/pipeline subject",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var AppendDocNoteInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"subject_type", "subject_key", "subject_key_field",
		"note_body_field", "note_categories", "note_categories_field",
		"note_site_id_field", "note_source", "source_item_id_field", "created_by",
	},
	Defaults: map[string]interface{}{
		"note_body_field":       "doc_note_body",
		"subject_key_field":     "input_data.subject_key",
		"note_categories_field": "doc_note_categories",
		"note_site_id_field":    "input_data.site_id",
		"source_item_id_field":  "input_data.item_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("append_doc_note", AppendDocNoteInputSpec)
}

func AppendDocNoteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "append_doc_note"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config

	subjectType, subjectKey, err := docResolveSubject(config, params.CollectedData)
	if err != nil {
		return nil, fmt.Errorf("append_doc_note: %w", err)
	}

	bodyField := datahelpers.GetStringField(config, "note_body_field", "doc_note_body")
	body := datahelpers.ExtractNestedFieldString(params.CollectedData, bodyField)
	if body == "" {
		return nil, fmt.Errorf("append_doc_note: empty note body at %q", bodyField)
	}

	categoriesJSON := docCategoriesJSON(config, params.CollectedData)

	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "note_site_id_field", "input_data.site_id"))
	sourceItemID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "source_item_id_field", "input_data.item_id"))

	sourceAgent := ""
	if params.ExecutionContext != nil {
		sourceAgent = params.Headers["agent_type"]
	}
	source := datahelpers.GetStringField(config, "note_source", sourceAgent)
	createdBy := datahelpers.GetStringField(config, "created_by", sourceAgent)

	noteID, err := insertDocNote(ctx, params.DB, subjectType, subjectKey,
		siteID, body, categoriesJSON, source, sourceAgent, sourceItemID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("append_doc_note: %w", err)
	}

	logger.Info("append_doc_note: entry appended",
		zap.String("subject_type", subjectType),
		zap.String("subject_key", subjectKey),
		zap.String("note_id", noteID),
		zap.String("categories", categoriesJSON))

	return map[string]interface{}{
		"note_id":      noteID,
		"subject_type": subjectType,
		"subject_key":  subjectKey,
		"categories":   categoriesJSON,
	}, nil
}

// insertDocNote is the single INSERT shared with persist_diagnosis_note.
func insertDocNote(ctx context.Context, db *sql.DB, subjectType, subjectKey,
	siteID, body, categoriesJSON, source, sourceAgent, sourceItemID, createdBy string) (string, error) {
	var id string
	err := db.QueryRowContext(ctx, `
		INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories,
		                       source, source_agent, source_item_id, created_by)
		VALUES ($1, $2, NULLIF($3,'')::uuid, $4, $5::jsonb,
		        NULLIF($6,''), NULLIF($7,''), NULLIF($8,'')::uuid, NULLIF($9,''))
		RETURNING id`,
		subjectType, subjectKey, siteID, body, categoriesJSON,
		source, sourceAgent, sourceItemID, createdBy).Scan(&id)
	return id, err
}

// docCategoriesJSON resolves categories from a direct config list or a
// collected-data field ([]interface{} or []string), returning a jsonb-ready
// string; defaults to "[]".
func docCategoriesJSON(config map[string]interface{}, collected map[string]interface{}) string {
	var raw interface{}
	if direct, ok := config["note_categories"]; ok && direct != nil {
		raw = direct
	} else {
		field := datahelpers.GetStringField(config, "note_categories_field", "doc_note_categories")
		raw = datahelpers.ExtractNestedField(collected, field)
	}
	if raw == nil {
		return "[]"
	}
	out := []string{}
	switch v := raw.(type) {
	case []string:
		out = v
	case []interface{}:
		for _, it := range v {
			if s, ok := it.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	case string:
		if v != "" {
			out = append(out, v)
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(b)
}
