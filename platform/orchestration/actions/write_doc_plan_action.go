// FILE: platform/orchestration/actions/write_doc_plan_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. Requires migration 0NN_doc_plans_and_notes.sql.
//
// write_doc_plan — supersede-pattern write of a travelling PLAN doc
// (PLAN_travelling_docs.md rev 4). One transaction: flip the current row for
// (subject_type, subject_key) to is_current=false + superseded_at, insert the
// new body as current. Mirrors the site_specs supersede pattern re-keyed to
// the subject. Library-level: called at first creation of a tool function
// (create_tool_component) or on intent change; NOT on deploy_tool_to_site
// forks. The caller owns subject validity (e.g. the function exists) — this
// action owns the write.
//
// Registration (add to registry.go):
//   "write_doc_plan": {
//       Handler:     WriteDocPlanAction,
//       Category:    "documentation",
//       Description: "Supersede-write the current PLAN doc for a tool/pipeline subject",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// Field names are prefixed/specific (dev guide §3) to avoid nested-lookup
// collisions; none are content_data/status/domain/site_id.
var WriteDocPlanInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"subject_type", "subject_key", "subject_key_field",
		"plan_body_field", "plan_source", "plan_notes_field",
		"source_item_id_field", "created_by",
	},
	Defaults: map[string]interface{}{
		"plan_body_field":      "doc_plan_body",
		"subject_key_field":    "input_data.subject_key",
		"source_item_id_field": "input_data.item_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_doc_plan", WriteDocPlanInputSpec)
}

func WriteDocPlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_doc_plan"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config

	subjectType, subjectKey, err := docResolveSubject(config, params.CollectedData)
	if err != nil {
		return nil, fmt.Errorf("write_doc_plan: %w", err)
	}

	bodyField := datahelpers.GetStringField(config, "plan_body_field", "doc_plan_body")
	body := datahelpers.ExtractNestedFieldString(params.CollectedData, bodyField)
	if body == "" {
		return nil, fmt.Errorf("write_doc_plan: empty PLAN body at %q", bodyField)
	}

	source := datahelpers.GetStringField(config, "plan_source", "")
	sourceAgent := ""
	orchID := ""
	if params.ExecutionContext != nil {
		sourceAgent = params.Headers["agent_type"]
		orchID = params.ExecutionContext.OrchestrationID
	}
	if source == "" {
		source = sourceAgent
	}
	planNotes := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "plan_notes_field", ""))
	sourceItemID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "source_item_id_field", "input_data.item_id"))
	createdBy := datahelpers.GetStringField(config, "created_by", sourceAgent)

	// One tx: supersede then insert (partial unique index enforces one current).
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("write_doc_plan: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck — no-op after Commit

	res, err := tx.ExecContext(ctx, `
		UPDATE doc_plans
		SET is_current = false, superseded_at = now(), updated_at = now()
		WHERE subject_type = $1 AND subject_key = $2 AND is_current`,
		subjectType, subjectKey)
	if err != nil {
		return nil, fmt.Errorf("write_doc_plan: supersede: %w", err)
	}
	superseded, _ := res.RowsAffected()

	var planID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO doc_plans (subject_type, subject_key, body, source, source_agent,
		                       source_item_id, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,'')::uuid, NULLIF($7,''), NULLIF($8,''))
		RETURNING id`,
		subjectType, subjectKey, body, nullIfEmpty(source), nullIfEmpty(sourceAgent),
		sourceItemID, planNotes, createdBy).Scan(&planID)
	if err != nil {
		return nil, fmt.Errorf("write_doc_plan: insert: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("write_doc_plan: commit: %w", err)
	}

	logger.Info("write_doc_plan: PLAN written",
		zap.String("subject_type", subjectType),
		zap.String("subject_key", subjectKey),
		zap.String("plan_id", planID),
		zap.Int64("superseded", superseded),
		zap.String("orchestration_id", orchID))

	return map[string]interface{}{
		"plan_id":      planID,
		"subject_type": subjectType,
		"subject_key":  subjectKey,
		"superseded":   superseded > 0,
	}, nil
}

// docResolveSubject resolves (subject_type, subject_key) from step config /
// collected data: subject_type is static per step (direct config only);
// subject_key is direct config or a field path. Shared by the doc actions.
func docResolveSubject(config map[string]interface{}, collected map[string]interface{}) (string, string, error) {
	subjectType := datahelpers.GetStringField(config, "subject_type", "")
	// Kept in lockstep with the doc_plans/doc_notes subject_type CHECK
	// constraint (migration 163 added 'experience' for the experience loop).
	// A value the DB accepts but this gate rejects — or vice versa — is a
	// split contract; move both together.
	if subjectType != "tool" && subjectType != "pipeline" && subjectType != "experience" {
		return "", "", fmt.Errorf("subject_type must be 'tool', 'pipeline' or 'experience', got %q", subjectType)
	}
	subjectKey := datahelpers.GetStringField(config, "subject_key", "")
	if subjectKey == "" {
		keyField := datahelpers.GetStringField(config, "subject_key_field", "input_data.subject_key")
		subjectKey = datahelpers.ExtractNestedFieldString(collected, keyField)
	}
	if subjectKey == "" {
		return "", "", fmt.Errorf("no subject_key (direct config or subject_key_field)")
	}
	return subjectType, subjectKey, nil
}
