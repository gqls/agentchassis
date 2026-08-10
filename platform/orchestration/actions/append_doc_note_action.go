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
		// note_body_suffix_field appends a MECHANICALLY-COMPOSED qualifier to the
		// body — a field written by an action, never by a model (bugs_open/223).
		// OPT-IN with the unsafe default OFF, per the owner ruling of 2026-08-02:
		// absent from a step's config, this action behaves byte-identically, so no
		// existing consumer changes and adoption is visible at each call site
		// rather than asserted in a comment.
		"note_body_suffix_field",
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

// applyBodySuffix appends a mechanically-composed qualifier to a note body.
//
// WHY THIS IS A FUNCTION AND NOT A PROMPT INSTRUCTION (bugs_open/223). The
// landmine-verifier's product is a doc_notes row of prose, read months later by
// sessions and by council seats, and its status is not parsed by anything. When a
// verdict rests on checks the code index could not answer, the ONLY way a future
// reader can tell is if the row says so — and a model asked to include that
// caveat will include it most of the time, which is exactly the property that
// makes the failures expensive: 3 of 4 runs on identical 0-row input already
// abstained correctly, and the entry was degraded by the fourth. A suffix composed
// in Go cannot be softened, contradicted or dropped by the model whose verdict it
// qualifies.
//
// An EMPTY resolution is stated in the row rather than skipped: the qualifier's
// absence is the very thing the key exists to make visible, so a misconfigured
// field must be loud in the artefact and not merely in a log nobody reads.
func applyBodySuffix(body, suffix, fieldName string) string {
	if fieldName == "" {
		return body
	}
	if suffix == "" {
		return body + "\n\n(note suffix field '" + fieldName + "' resolved EMPTY — the mechanical qualifier is MISSING, not withheld: treat this note as unqualified.)"
	}
	return body + "\n\n" + suffix
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
	// The suffix is applied AFTER the empty-body refusal above, so a configured
	// suffix can never make an empty verdict look like a written one.
	suffixField := datahelpers.GetStringField(config, "note_body_suffix_field", "")
	body = applyBodySuffix(body, datahelpers.ExtractNestedFieldString(params.CollectedData, suffixField), suffixField)

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
