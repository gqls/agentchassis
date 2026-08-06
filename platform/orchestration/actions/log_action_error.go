// FILE: platform/orchestration/actions/log_action_error.go
//
// The actions-package door onto the ONE agent_error_log writer
// (agenterrors.Write — RFC_012 option B, owner ruling 2026-08-06).
//
// Until the writer moved to a leaf package, every action that needed a durable
// row hand-copied the INSERT — nineteen copies whose column lists drifted to
// 8/9/10/11/13 columns, with orchestration_id absent from nine of them (a row
// that cannot be joined to its run) and three incompatible null-handling
// disciplines. This helper retires the copy class: it resolves the identity
// columns from ActionParams the way the best of those copies did
// (retract_page_deployment's, proven live 2026-08-05) and delegates the write.
//
// For findings that must survive an AWAIT — the RFC_012 class — call
// LogActionFindings BEFORE the dispatch. THIS TABLE IS THE ONLY SINK THAT
// SURVIVES AN AWAITED STEP: the collected_data sibling key was refuted live
// (park loads fresh state; RFC_012 addendum 2). Both helpers are best-effort
// and COUNT their failures rather than swallowing them.

package actions

import (
	"context"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// actionErrorEntry builds the identity half of an agent_error_log row from
// ActionParams. Fields the caller sets on the returned Entry win; this only
// fills what the params can answer.
func actionErrorEntry(params ActionParams, siteID, domain string) agenterrors.Entry {
	entry := agenterrors.Entry{
		SiteID:    siteID,
		Domain:    domain,
		AgentType: params.AgentType,
		StepName:  params.CurrentStep,
	}
	if params.ExecutionContext != nil {
		entry.OrchestrationID = params.ExecutionContext.OrchestrationID
		if params.ExecutionContext.Sender.AgentType != "" {
			entry.AgentType = params.ExecutionContext.Sender.AgentType
		}
		entry.AgentID = params.ExecutionContext.Sender.AgentID
		entry.PodName = params.ExecutionContext.Sender.PodName
		if params.ExecutionContext.StepName != "" {
			entry.StepName = params.ExecutionContext.StepName
		}
	}
	// The item id most actions carry, when they carry one.
	entry.WorkItemID = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id")
	return entry
}

// LogActionError persists one row from inside an action, identity resolved
// from params. Best-effort: returns whether the row landed; a failure must
// never change the disposition the caller has already decided.
func LogActionError(ctx context.Context, params ActionParams, siteID, domain, action, code, severity, message string, contextPayload map[string]interface{}, logger *zap.Logger) bool {
	entry := actionErrorEntry(params, siteID, domain)
	entry.Action = action
	entry.ErrorCode = code
	entry.Severity = severity
	entry.ErrorMessage = message
	entry.Context = contextPayload
	return agenterrors.Write(ctx, params.DB, logger, entry)
}

// LogActionFindings persists a set of findings an action computed and needs to
// survive an await — one row each, identity shared. Call BEFORE the dispatch,
// so a failed send cannot unrecord what the action found. Put the difference
// between the returns in your audit rather than letting a lost row read as a
// recorded one:
//
//	attempted, recorded := LogActionFindings(ctx, params, siteID, domain, action, findings, logger)
//	audit["conditions_recorded"] = recorded
//	if attempted != recorded { audit["conditions_lost"] = attempted - recorded }
func LogActionFindings(ctx context.Context, params ActionParams, siteID, domain, action string, findings []agenterrors.Finding, logger *zap.Logger) (attempted, recorded int) {
	entry := actionErrorEntry(params, siteID, domain)
	entry.Action = action
	return agenterrors.RecordFindings(ctx, params.DB, logger, entry, findings)
}
