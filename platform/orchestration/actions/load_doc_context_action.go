// FILE: platform/orchestration/actions/load_doc_context_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. Requires migration 0NN_doc_plans_and_notes.sql.
//
// load_doc_context — the direct-by-key loader (maintenance_actions mould:
// params.DB + QueryContext, read-only). A fixer that knows its subject loads
// the current PLAN + latest-N NOTES BEFORE changing anything; a sibling step
// can hand doc_context to diagnose_assemble_bundle the same way
// runtime_evidence is handed in. Also extracts the machine-readable
// ```criteria fenced block from the PLAN body (tool-doc-header precedent:
// structured block parsed from a larger artifact) for the acceptance ladder.
//
// NOT an error when no PLAN exists — a new/undocumented subject is a normal
// state (has_plan=false); per house rules a 0-row result is information, not
// failure. rag_lookup remains the discovery route when the key is unknown.
//
// Registration (add to registry.go):
//   "load_doc_context": {
//       Handler:     LoadDocContextAction,
//       Category:    "documentation",
//       Description: "Load current PLAN + latest NOTES for a tool/pipeline subject",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadDocContextInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{"subject_type", "subject_key", "subject_key_field", "notes_limit"},
	Defaults: map[string]interface{}{
		"subject_key_field": "input_data.subject_key",
		"notes_limit":       10,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_doc_context", LoadDocContextInputSpec)
}

func LoadDocContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_doc_context"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config

	subjectType, subjectKey, err := docResolveSubject(config, params.CollectedData)
	if err != nil {
		return nil, fmt.Errorf("load_doc_context: %w", err)
	}
	notesLimit := datahelpers.GetIntField(config, "notes_limit", 10)

	// Current PLAN (0 rows = no plan yet, not an error).
	planBody := ""
	hasPlan := false
	err = params.DB.QueryRowContext(ctx, `
		SELECT body FROM doc_plans
		WHERE subject_type = $1 AND subject_key = $2 AND is_current`,
		subjectType, subjectKey).Scan(&planBody)
	if err == nil {
		hasPlan = true
	} else if !strings.Contains(err.Error(), "no rows") {
		return nil, fmt.Errorf("load_doc_context: plan query: %w", err)
	}

	// Latest-N NOTES, newest first.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT body, categories::text, created_at::text
		FROM doc_notes
		WHERE subject_type = $1 AND subject_key = $2
		ORDER BY created_at DESC
		LIMIT $3`,
		subjectType, subjectKey, notesLimit)
	if err != nil {
		return nil, fmt.Errorf("load_doc_context: notes query: %w", err)
	}
	defer rows.Close()

	notes := []map[string]interface{}{}
	for rows.Next() {
		var body, cats, createdAt string
		if err := rows.Scan(&body, &cats, &createdAt); err != nil {
			continue
		}
		notes = append(notes, map[string]interface{}{
			"body": body, "categories": cats, "created_at": createdAt,
		})
	}

	// Compose one prompt-ready text block (hypothesis-style self-containment,
	// mirroring diagnose_assemble_bundle's composed-bundle approach).
	var b strings.Builder
	fmt.Fprintf(&b, "## Travelling docs — %s %s\n\n", subjectType, subjectKey)
	if hasPlan {
		b.WriteString("### PLAN (current — read Deliberate decisions before changing anything)\n\n")
		b.WriteString(planBody)
		b.WriteString("\n\n")
	} else {
		b.WriteString("### PLAN\n\n(none recorded yet)\n\n")
	}
	if len(notes) > 0 {
		fmt.Fprintf(&b, "### NOTES (latest %d, newest first)\n\n", len(notes))
		for _, n := range notes {
			fmt.Fprintf(&b, "%s\n\n", n["body"])
		}
	} else {
		b.WriteString("### NOTES\n\n(none recorded yet)\n\n")
	}

	criteriaJSON := extractCriteriaBlock(planBody)

	logger.Info("load_doc_context: loaded",
		zap.String("subject_type", subjectType),
		zap.String("subject_key", subjectKey),
		zap.Bool("has_plan", hasPlan),
		zap.Int("notes", len(notes)),
		zap.Bool("has_criteria", criteriaJSON != ""))

	return map[string]interface{}{
		"doc_context":   b.String(),
		"has_plan":      hasPlan,
		"plan_body":     planBody,
		"notes":         notes,
		"note_count":    len(notes),
		"criteria_json": criteriaJSON, // "" when the PLAN has no ```criteria block
	}, nil
}

// extractCriteriaBlock returns the raw contents of the first ```criteria
// fenced block in a PLAN body ("" if absent). Convention: a JSON object under
// the "## Acceptance criteria" heading; consumers (contract-presence check,
// acceptance runner) parse it. Kept as a body-embedded block per the
// graduation rule — lift to a column only if volume demands.
func extractCriteriaBlock(body string) string {
	const fence = "```criteria"
	start := strings.Index(body, fence)
	if start < 0 {
		return ""
	}
	rest := body[start+len(fence):]
	end := strings.Index(rest, "```")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:end])
}
