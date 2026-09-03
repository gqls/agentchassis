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
	"strings"

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

	// THE WRITE DOOR VALIDATES THE FACTS DECLARATION IT ACCEPTS (bugs_open/288
	// defect A). Until 2026-08-24 nothing did: bugs_open/288, its lane PLAN and
	// its council submission all stated that "validator rule P11 refuses a
	// malformed declaration where it is written", and P11 lives in
	// ValidateExperienceCriteria (experience_criteria.go), whose only production
	// caller is write_experience_pattern_action.go — the EXPERIENCE-PATTERN
	// register. Tool PLANs come through here, and here did no validation of any
	// kind, so P11 had never once seen a tool fence.
	//
	// Deliberately NOT ValidateExperienceCriteria: that validator judges a whole
	// experience-criteria template and would reject a tool PLAN for absent
	// `checks`/`binding_schema`, which are not this document's business. It
	// shares the part that matters instead — criteriaFactsFromValue, whose own
	// comment says it exists "so the two cannot disagree on what a well-formed
	// declaration is". This is its third caller, not a third spelling of the rule.
	//
	// Scoped to tool subjects whose fence actually mentions facts, so every
	// other PLAN write is byte-identical to before. Refusing rather than
	// logging, because the whole defect being closed is that a declaration
	// nobody can read looked exactly like a document that declared nothing.
	// The fence is extracted ONCE and shared by both rules below. Two
	// extractions would be two chances to disagree about which ```criteria
	// block is the fence, and extractCriteriaBlock has a documented landmine of
	// its own (prose naming the fence in backticks hijacks the FIRST match).
	toolCriteria := ""
	if subjectType == "tool" {
		toolCriteria = extractCriteriaBlock(body)
		if factsKeyMentioned(toolCriteria) {
			if _, issues := parseCriteriaFacts(toolCriteria); len(issues) > 0 {
				return nil, fmt.Errorf(
					"write_doc_plan: refusing PLAN for tool %q — its criteria fence declares facts that cannot be read: %s",
					subjectKey, strings.Join(issues, "; "))
			}
		}
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

	// THE DOOR SEES A BLIND FENCE BEING BORN (bugs_open/449).
	//
	// This action is the ONLY production writer of a PLAN body — verified by
	// enumerating every Go INSERT/UPDATE against doc_plans — and exactly three
	// live agents reach it (tool-generator, experience-planner,
	// experience-register-writer). So it is the one place that sees every
	// GENERATED fence at the moment it is written. (Operator scripts write
	// doc_plans over psql and bypass this entirely; that is the right way round
	// — the hand-installed fences are the ones that already assert values — but
	// it means this is not a total guarantee and must not be described as one.)
	//
	// Measured 2026-09-03: tool-generator had written 186 current fences, 115
	// asserting no expected value of any kind, newest that same day. Nothing
	// anywhere said so, so the only way to know was to run the census by hand.
	//
	// RECORDED, NOT REFUSED, and that is a considered choice rather than
	// timidity. A tool with NO PLAN is inert at BOTH tiers — Tier 2 writes a
	// needs_criteria note and stops, Tier 4 emits nothing — which is strictly
	// worse than a weak fence. Refusing becomes correct only once the authoring
	// prompts can satisfy the rule; until then a refusal here would trade a
	// blind check for no check at all.
	//
	// The trigger is read off the FENCE, never off a classifier: a check
	// carrying a `fill` or `select` step has declared that the tool takes input,
	// so a document that drives inputs and then asserts nothing about what came
	// out is incomplete on its own terms, with no outside judgement about what
	// kind of tool this is. 55 of the 186 met it on the day this was written.
	//
	// AFTER the commit, deliberately: a note about a PLAN that failed to write
	// would be a record of a document that does not exist. And non-fatal in both
	// directions — an unreadable fence yields Parsed=false and no note (Tier 2
	// already reports that case separately as criteria_unparseable, and a
	// second, weaker report of it would be noise), while a failed note INSERT is
	// logged and never returned: the PLAN is the caller's deliverable and an
	// observation about it must not cost the document.
	if subjectType == "tool" {
		if a := summariseCriteriaValueAssertions(toolCriteria); a.DrivesButAssertsNothing() {
			author := sourceAgent
			if author == "" {
				author = "an unidentified caller"
			}
			noteBody := fmt.Sprintf(`## Fence asserts no value — %s
Observed: the criteria fence just written for this tool DRIVES its inputs (a fill or select step) and then asserts no expected value anywhere — no computed_values expect_values, no interaction expect.text_matches. A Tier-4 PASS on this fence means the page loaded and something appeared once the inputs were filled; it says NOTHING about whether any number the tool printed is correct.
Root cause: the fence-authoring prompt enumerates a closed vocabulary of liveness checks and never names computed_values, so the correctness check is never a candidate (bugs_open/449 §3). Written by %s.
Fix: none applied — this is a record, not a refusal. A tool with no PLAN at all is inert at BOTH tiers, so a blind fence is still better than none and the document is accepted as written. To close it, assert at least one worked example whose expected value did not come from this tool's own output.
Verified: n/a — an observation about the document, made at the write door
Categories: fence_asserts_no_value`, subjectKey, author)
			if _, nerr := insertDocNote(ctx, params.DB, "tool", subjectKey, "", noteBody,
				`["fence_asserts_no_value"]`, "write-doc-plan", sourceAgent, "", createdBy); nerr != nil {
				logger.Warn("write_doc_plan: fence_asserts_no_value note insert failed",
					zap.String("subject_key", subjectKey), zap.Error(nerr))
			} else {
				logger.Info("write_doc_plan: fence drives inputs but asserts no value (bugs_open/449)",
					zap.String("subject_key", subjectKey), zap.String("author", sourceAgent))
			}
		}
	}

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
	// Membership in validDocSubjectTypes (doc_subjects_common.go), the single
	// Go-side source kept in lockstep with the doc_plans/doc_notes
	// subject_type CHECK constraints. A value the DB accepts but this gate
	// rejects — or vice versa — is a split contract (bugs_open/064); move
	// both together.
	if !isValidDocSubjectType(subjectType) {
		return "", "", fmt.Errorf("subject_type must be one of %s, got %q", docSubjectTypesQuoted(), subjectType)
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
