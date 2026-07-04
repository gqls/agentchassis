// FILE: platform/orchestration/actions/persist_diagnosis_note_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. Requires migration 0NN_doc_plans_and_notes.sql.
//
// persist_diagnosis_note — config-gated step placed AFTER diagnose_emit in the
// diagnosis agent's workflow. diagnose_emit stays read-only (its own design
// comment: persisting is a separate decision needing its own schema — this is
// that schema). Writes ONE doc_notes row from the shaped report when, and only
// when, the run's subject is explicit in input_data — SKIP, DON'T GUESS: a
// mis-filed note poisons a subject's history. UNVERIFIABLE runs are still
// persisted, tagged unconfirmed-diagnosis — dead ends prevent retries.
// Reuses insertDocNote (append_doc_note_action.go) and docResolveSubject
// (write_doc_plan_action.go); the subject_type here comes from input_data,
// not static step config, so it is resolved locally.
//
// Registration (add to registry.go):
//   "persist_diagnosis_note": {
//       Handler:     PersistDiagnosisNoteAction,
//       Category:    "documentation",
//       Description: "Persist a diagnosis report as a NOTES entry when the subject is explicit",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// diagnose_route is a router, so diagnose_emit's inputs land under "route";
// emit's own output lands under its output_field "diagnosis" (see the
// diagnosis migration) — these defaults read the emitted report.
var PersistDiagnosisNoteInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"diag_status_field", "diag_summary_field", "diag_conclusion_field",
		"diag_stopped_by_field", "diag_trail_field",
		"diag_subject_type_field", "diag_subject_key_field",
		"diag_symptom_field", "diag_site_id_field", "source_item_id_field",
	},
	Defaults: map[string]interface{}{
		"diag_status_field":       "diagnosis.status",
		"diag_summary_field":      "diagnosis.summary",
		"diag_conclusion_field":   "diagnosis.conclusion",
		"diag_stopped_by_field":   "diagnosis.stopped_by",
		"diag_trail_field":        "diagnosis.evidence_trail",
		"diag_subject_type_field": "input_data.subject_type",
		"diag_subject_key_field":  "input_data.subject_key",
		"diag_symptom_field":      "input_data.symptom",
		"diag_site_id_field":      "input_data.site_id",
		"source_item_id_field":    "input_data.item_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("persist_diagnosis_note", PersistDiagnosisNoteInputSpec)
}

func PersistDiagnosisNoteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "persist_diagnosis_note"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config
	cd := params.CollectedData
	get := func(key, def string) string {
		return datahelpers.ExtractNestedFieldString(cd, datahelpers.GetStringField(config, key, def))
	}

	// Subject gate — skip (success, persisted=false), never guess.
	subjectType := get("diag_subject_type_field", "input_data.subject_type")
	subjectKey := get("diag_subject_key_field", "input_data.subject_key")
	if (subjectType != "tool" && subjectType != "pipeline") || subjectKey == "" {
		logger.Info("persist_diagnosis_note: no explicit subject — skipping (do not guess)",
			zap.String("subject_type", subjectType), zap.String("subject_key", subjectKey))
		return map[string]interface{}{"persisted": false, "reason": "no explicit subject"}, nil
	}

	status := get("diag_status_field", "diagnosis.status")
	conclusion := get("diag_conclusion_field", "diagnosis.conclusion")
	if status == "" && conclusion == "" {
		logger.Info("persist_diagnosis_note: no diagnosis in collected data — skipping")
		return map[string]interface{}{"persisted": false, "reason": "no diagnosis report found"}, nil
	}
	stoppedBy := get("diag_stopped_by_field", "diagnosis.stopped_by")
	symptom := get("diag_symptom_field", "input_data.symptom")

	trailLen := 0
	if trail := datahelpers.ExtractNestedField(cd,
		datahelpers.GetStringField(config, "diag_trail_field", "diagnosis.evidence_trail")); trail != nil {
		if arr, ok := trail.([]interface{}); ok {
			trailLen = len(arr)
		}
	}

	// Import-shaped entry (RUNBOOK §3 format).
	confirmed := status == "CONFIRMED"
	rootCause := conclusion
	if !confirmed {
		rootCause = fmt.Sprintf("unconfirmed (%s) — best-effort trail; recorded as a dead end. %s", stoppedBy, conclusion)
	}
	body := fmt.Sprintf(
		"## %s — diagnosis: %s\nObserved: %s\nRoot cause: %s\nVerified: evidence trail, %d iteration(s); status %s\nCategories: diagnosis%s",
		time.Now().Format("2006-01-02"),
		firstNonEmptyDoc(symptom, subjectKey),
		firstNonEmptyDoc(symptom, "(no seed symptom recorded)"),
		rootCause,
		trailLen,
		status,
		map[bool]string{true: "", false: ", unconfirmed-diagnosis"}[confirmed],
	)

	categories := []string{"diagnosis"}
	if !confirmed {
		categories = append(categories, "unconfirmed-diagnosis")
	}
	categoriesJSON := docCategoriesJSONFromList(categories)

	siteID := get("diag_site_id_field", "input_data.site_id")
	sourceItemID := get("source_item_id_field", "input_data.item_id")
	sourceAgent := ""
	if params.ExecutionContext != nil {
		sourceAgent = params.Headers["agent_type"]
	}

	noteID, err := insertDocNote(ctx, params.DB, subjectType, subjectKey,
		siteID, body, categoriesJSON, "diagnosis-loop", sourceAgent, sourceItemID, sourceAgent)
	if err != nil {
		return nil, fmt.Errorf("persist_diagnosis_note: %w", err)
	}

	logger.Info("persist_diagnosis_note: persisted",
		zap.String("subject_type", subjectType),
		zap.String("subject_key", subjectKey),
		zap.String("note_id", noteID),
		zap.Bool("confirmed", confirmed))

	return map[string]interface{}{
		"persisted":    true,
		"note_id":      noteID,
		"subject_type": subjectType,
		"subject_key":  subjectKey,
		"confirmed":    confirmed,
	}, nil
}

// docCategoriesJSONFromList marshals a []string to a jsonb-ready string.
func docCategoriesJSONFromList(cats []string) string {
	out := "["
	for i, c := range cats {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("%q", c)
	}
	return out + "]"
}

// firstNonEmptyDoc returns the first non-empty string.
func firstNonEmptyDoc(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
