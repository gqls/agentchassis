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
//
// Use this when the row should be filed under the EXECUTING step's identity.
// When the site carries its own provenance, use LogActionEntry — see the
// warning there, which is a council finding, not a style preference.
func LogActionError(ctx context.Context, params ActionParams, siteID, domain, action, code, severity, message string, contextPayload map[string]interface{}, logger *zap.Logger) bool {
	return LogActionEntry(ctx, params, agenterrors.Entry{
		SiteID:       siteID,
		Domain:       domain,
		Action:       action,
		ErrorCode:    code,
		Severity:     severity,
		ErrorMessage: message,
		Context:      contextPayload,
	}, logger)
}

// LogActionEntry persists one row, taking the caller's fields as authoritative
// and filling ONLY what the caller left zero from ActionParams. It is the door
// for a site that files under its OWN provenance rather than the executing
// step's — an origin/provenance struct, or hard-coded literals.
//
// ⚠ PROVENANCE IS NOT INTERCHANGEABLE, and this is a council finding rather
// than a preference. When consolidating the birth-path recorder in
// store_generated_component_action.go was first proposed, the edit-quality and
// guardian seats both objected that "a provenance-literal slip here would
// silently misfile birth-path rejections fleet-wide" — a row attributed to the
// wrong agent/step is worse than no row, because it is believed. Several
// converted sites deliberately file under a provenance that is NOT the running
// step (component_link_repair and the validate_page_content link recorder file
// under the ORIGIN of the content they repaired; store_generated_component
// files as "component-creator"/"store_component"). Set those fields explicitly
// here; never let them be inherited.
//
// Merge semantics, stated because they are load-bearing: a field the caller
// sets is used verbatim; a field left zero is filled from params (orchestration
// id, sender agent/pod, step name, work item id). That is what gives the nine
// historically orchestration_id-less sites their run join for free, while
// leaving every provenance literal exactly where its author put it.
func LogActionEntry(ctx context.Context, params ActionParams, entry agenterrors.Entry, logger *zap.Logger) bool {
	base := actionErrorEntry(params, entry.SiteID, entry.Domain)
	if entry.WorkItemID == "" {
		entry.WorkItemID = base.WorkItemID
	}
	if entry.OrchestrationID == "" {
		entry.OrchestrationID = base.OrchestrationID
	}
	if entry.AgentType == "" {
		entry.AgentType = base.AgentType
	}
	if entry.AgentID == "" {
		entry.AgentID = base.AgentID
	}
	if entry.PodName == "" {
		entry.PodName = base.PodName
	}
	if entry.StepName == "" {
		entry.StepName = base.StepName
	}
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
	return LogActionEntryFindings(ctx, params, agenterrors.Entry{
		SiteID: siteID,
		Domain: domain,
		Action: action,
	}, findings, logger)
}

// LogActionEntryFindings is LogActionFindings for a site that files under its
// OWN provenance — the findings form of LogActionEntry, with the same merge
// semantics and the same warning about never inheriting a provenance literal.
func LogActionEntryFindings(ctx context.Context, params ActionParams, base agenterrors.Entry, findings []agenterrors.Finding, logger *zap.Logger) (attempted, recorded int) {
	merged := actionErrorEntry(params, base.SiteID, base.Domain)
	if base.WorkItemID != "" {
		merged.WorkItemID = base.WorkItemID
	}
	if base.OrchestrationID != "" {
		merged.OrchestrationID = base.OrchestrationID
	}
	if base.AgentType != "" {
		merged.AgentType = base.AgentType
	}
	if base.AgentID != "" {
		merged.AgentID = base.AgentID
	}
	if base.PodName != "" {
		merged.PodName = base.PodName
	}
	if base.StepName != "" {
		merged.StepName = base.StepName
	}
	merged.Action = base.Action
	return agenterrors.RecordFindings(ctx, params.DB, logger, merged, findings)
}
