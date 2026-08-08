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
//
// TWO HALVES, AND ONLY ONE OF THEM IS SAFE TO INHERIT (2026-08-08). The merge
// this file performs was originally one undifferentiated fill-what-is-zero pass.
// Four council seats — bug_historian, guardian, architecture, and the RFC_012
// lane's own risks block — independently named the same gap: a provenance the
// caller DELIBERATELY omits is safe, but one it FORGETS was filled silently
// from the running step, with no error, no warning, and no test in the package
// that would catch it. So the merge is now split explicitly:
//
//   - the JOIN half (orchestration_id, agent_id, pod_name, work_item_id) is
//     inherited whenever the caller leaves it zero. Nobody can get this wrong:
//     these columns answer "which run/pod produced this row", and the running
//     step is always the right answer. This is what gave the nine historically
//     orchestration_id-less sites their run join for free.
//
//   - the PROVENANCE half (agent_type, step_name) is NEVER inherited unless the
//     caller says so, by choosing LogActionEntryInheritingProvenance over
//     LogActionEntry. Unsafe default OFF, declared at the call site, per OWNER
//     RULING 2026-08-02 (RFC_010 §2): "when a seam's widest branch is licensed
//     by 'callers must all be X', make X a field with the unsafe default OFF …
//     a comment is not a control on a tree this many sessions share."
//     BOTH columns, and the symmetry was earned rather than designed in: the
//     first version inherited step_name whenever agent_type was named, and two
//     seats of the approving council round (bug_historian medium, editquality)
//     objected that a mixed row — agent from the caller, step from the runtime —
//     is "a narrower recurrence of the exact defect being fixed". They were
//     right. An empty step_name asserts nothing; a borrowed one asserts
//     something nobody claimed.
//
// Why a named door rather than a bool on agenterrors.Entry (the shape the
// RFC_012 handoff sketched): Entry is the ROW, and agenterrors.Write — the
// package that owns the type — would ignore such a field entirely. An inert
// field on a shared leaf type is its own trap, and orchestration.AgentErrorEntry
// is an alias of Entry, so agentbase, messaging and the coordinator would all
// see a knob that does nothing for them. A function name costs nothing, is
// visible to a reviewer reading only the caller, and makes the census a grep.
//
// WHY THE UNATTRIBUTED ROW STILL LANDS. The tempting alternative is to refuse
// the write, and component_write_guard.go's "a row that misattributes the
// writer is worse than no row" appears to license it. But this table is the
// only sink that survives an awaited step, so a refusal silently destroys a
// finding — and a sentinel does not misattribute, it declares ignorance. So the
// row lands carrying agent_type = unattributedAgentType, the running step's
// identity demoted into context where it asserts nothing, and an Error log
// naming the caller. It is queryable as a standing detector:
//
//	SELECT action, error_code, count(*) FROM agent_error_log
//	WHERE agent_type = 'unattributed' GROUP BY 1, 2;
//
// Live at the time of writing: 0 rows, and 0 rows carrying agent_type '' or
// 'unknown' either — so this path fires for a MISTAKE, never for normal
// traffic.

package actions

import (
	"context"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// unattributedAgentType is what a row carries when its writer named no
// provenance and did not declare inheritance.
//
// It is deliberately a non-empty literal: agent_type is NOT NULL on
// agent_error_log, and the best-effort writer swallows a constraint violation
// as a warn — so an empty string does not produce a bad row, it produces NO
// row, which is the worst outcome of the three. It is also deliberately not
// "generic" or "unknown": both are live values with real traffic ("generic" has
// 559 rows across 25 distinct step_names, the widest spread of any agent_type
// on the table, precisely because it is what an unresolved sender falls back
// to), so a detector keyed on either could never separate a provenance mistake
// from ordinary noise.
const unattributedAgentType = "unattributed"

// actionJoinIdentity builds the JOIN half of an agent_error_log row from
// ActionParams — the columns that answer "which run, which pod" and never
// "whose finding is this". Safe to inherit unconditionally.
func actionJoinIdentity(params ActionParams, siteID, domain string) agenterrors.Entry {
	entry := agenterrors.Entry{
		SiteID: siteID,
		Domain: domain,
	}
	if params.ExecutionContext != nil {
		entry.OrchestrationID = params.ExecutionContext.OrchestrationID
		entry.AgentID = params.ExecutionContext.Sender.AgentID
		entry.PodName = params.ExecutionContext.Sender.PodName
	}
	// The item id most actions carry, when they carry one.
	entry.WorkItemID = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id")
	return entry
}

// runningStepProvenance resolves the PROVENANCE half — the identity of the step
// currently executing. The context wins over the params fallback when it can
// answer, because a spawned child's params can lag its sender.
//
// The agent-type half is NOT this function's own ladder any more, and that is
// the point of the change. It asks types.ExecutionContext.ResolvedAgentType,
// which is the same ladder coordinator.determineOwnerAgentType climbs to fill
// orchestration_states.owner_agent_type. Before 2026-08-08 this function read
// Sender.AgentType directly — one rung short — so a row filed under "the
// running step" recorded the dispatch-path sender ('generic') while the
// orchestration row for the same run recorded the real agent. Two
// implementations of one question, and they disagreed in production.
//
// The tail below the context is deliberately this package's own and is NOT in
// the shared method: params.AgentType is state.OwnerAgentType, i.e. the durable
// output of that same ladder recorded when the orchestration was created. It is
// the right floor HERE and the wrong one in the coordinator, which is why the
// shared method stops at the context and returns "".
func runningStepProvenance(params ActionParams) (agentType, stepName string) {
	agentType, stepName = params.AgentType, params.CurrentStep
	if params.ExecutionContext != nil {
		if resolved := params.ExecutionContext.ResolvedAgentType(); resolved != "" {
			agentType = resolved
		}
		if params.ExecutionContext.StepName != "" {
			stepName = params.ExecutionContext.StepName
		}
	}
	return agentType, stepName
}

// inheritJoinIdentity fills the join half of entry from base, leaving anything
// the caller set alone. Provenance is deliberately absent from this function:
// that is the whole point of the split.
func inheritJoinIdentity(entry *agenterrors.Entry, base agenterrors.Entry) {
	if entry.WorkItemID == "" {
		entry.WorkItemID = base.WorkItemID
	}
	if entry.OrchestrationID == "" {
		entry.OrchestrationID = base.OrchestrationID
	}
	if entry.AgentID == "" {
		entry.AgentID = base.AgentID
	}
	if entry.PodName == "" {
		entry.PodName = base.PodName
	}
}

// withUnattributedDiagnostics returns a COPY of the caller's context payload
// carrying the diagnostics for an unattributed row. A copy, not a mutation:
// the map belongs to the caller (and, in the findings path, is shared across
// every finding in the batch), so writing into it would leak this row's
// bookkeeping into the caller's own state.
func withUnattributedDiagnostics(payload map[string]interface{}, agentType, stepName string) map[string]interface{} {
	annotated := make(map[string]interface{}, len(payload)+3)
	for k, v := range payload {
		annotated[k] = v
	}
	annotated["provenance_missing"] = true
	// Demoted, not asserted: this is who was RUNNING, which is not the same
	// claim as who the finding belongs to.
	if agentType != "" {
		annotated["provenance_running_agent_type"] = agentType
	}
	if stepName != "" {
		annotated["provenance_running_step_name"] = stepName
	}
	annotated["remedy"] = "this row's writer named no agent_type: set Entry.AgentType explicitly, or call LogActionEntryInheritingProvenance if the running step IS the right provenance"
	return annotated
}

// resolveProvenance applies the merge to entry in place. inheritProvenance is
// the caller's declaration, not a guess: false means an unnamed agent_type is a
// mistake and is marked as one.
func resolveProvenance(params ActionParams, entry *agenterrors.Entry, inheritProvenance bool, logger *zap.Logger) {
	inheritJoinIdentity(entry, actionJoinIdentity(params, entry.SiteID, entry.Domain))
	runningAgentType, runningStepName := runningStepProvenance(params)

	if inheritProvenance {
		if entry.AgentType == "" {
			entry.AgentType = runningAgentType
		}
		if entry.StepName == "" {
			entry.StepName = runningStepName
		}
		return
	}

	if entry.AgentType != "" {
		// The caller named its provenance, so it owns this row — INCLUDING
		// step_name, which is deliberately NOT filled from the running step.
		//
		// This branch used to inherit step_name, and two council seats
		// (bug_historian medium, editquality) objected in the same round that it
		// was "a narrower recurrence of the exact defect being fixed": a mixed
		// row, agent from the caller and step from the runtime, is a claim
		// neither of them made. An EMPTY step_name asserts nothing, which is the
		// property that matters — the defect was false attribution, not absence.
		// The only two sites of that shape (prepare_link_context,
		// diagnose_persist_fix_plan) now declare inheritance instead, so no
		// row's contents changed.
		if entry.StepName == "" && logger != nil {
			logger.Warn("agent_error_log row names an agent_type but no step_name — recorded with step_name empty rather than borrowing the running step's",
				zap.String("agent_type", entry.AgentType),
				zap.String("action", entry.Action),
				zap.String("running_step_name", runningStepName),
				zap.String("remedy", "set Entry.StepName, or call LogActionEntryInheritingProvenance if the running step IS the right step"))
		}
		return
	}

	// Neither named nor declared. Land the row — it may be the only surviving
	// copy of a finding — but never let it claim an identity it did not earn.
	entry.AgentType = unattributedAgentType
	entry.StepName = ""
	entry.Context = withUnattributedDiagnostics(entry.Context, runningAgentType, runningStepName)
	if logger != nil {
		logger.Error("agent_error_log row written with NO provenance — recorded as unattributed rather than credited to the running step",
			zap.String("action", entry.Action),
			zap.String("error_code", entry.ErrorCode),
			zap.String("running_agent_type", runningAgentType),
			zap.String("running_step_name", runningStepName),
			zap.String("remedy", "set Entry.AgentType at the call site, or call LogActionEntryInheritingProvenance if the running step IS the right provenance"))
	}
}

// LogActionError persists one row from inside an action, filed under the
// EXECUTING step's identity — provenance inheritance is this helper's whole
// contract, so it declares it on the caller's behalf.
//
// When the site carries its OWN provenance, use LogActionEntry — see the
// warning there, which is a council finding, not a style preference.
func LogActionError(ctx context.Context, params ActionParams, siteID, domain, action, code, severity, message string, contextPayload map[string]interface{}, logger *zap.Logger) bool {
	return LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
		SiteID:       siteID,
		Domain:       domain,
		Action:       action,
		ErrorCode:    code,
		Severity:     severity,
		ErrorMessage: message,
		Context:      contextPayload,
	}, logger)
}

// LogActionEntry persists one row for a site that files under its OWN
// provenance — an origin/provenance struct, or hard-coded literals. The
// caller's fields are authoritative; the JOIN half is filled from ActionParams
// where the caller left it zero.
//
// ⚠ PROVENANCE IS NOT INHERITED HERE, and this is a council finding rather than
// a preference. When consolidating the birth-path recorder in
// store_generated_component_action.go was first proposed, the edit-quality and
// guardian seats both objected that "a provenance-literal slip here would
// silently misfile birth-path rejections fleet-wide" — a row attributed to the
// wrong agent/step is worse than no row, because it is believed. Several sites
// deliberately file under a provenance that is NOT the running step
// (component_link_repair and the validate_page_content link recorder file under
// the ORIGIN of the content they repaired; store_generated_component files as
// "component-creator"/"store_component").
//
// So: name AgentType, or call LogActionEntryInheritingProvenance. An Entry that
// does neither still lands, marked unattributed — see the file header for why
// that beats both silent inheritance and a refused write.
func LogActionEntry(ctx context.Context, params ActionParams, entry agenterrors.Entry, logger *zap.Logger) bool {
	resolveProvenance(params, &entry, false, logger)
	return agenterrors.Write(ctx, params.DB, logger, entry)
}

// LogActionEntryInheritingProvenance is LogActionEntry for a site where the
// RUNNING STEP genuinely is the right provenance — the declared opt-in to the
// inheritance LogActionEntry refuses to perform silently.
//
// Use it when the row belongs to the step that is executing and the site still
// needs the Entry door for the rest of the shape (an explicit work_item_id or
// site_id that is NOT the running step's, most often). Reach for it
// deliberately: `grep -rn LogActionEntryInheritingProvenance` is the census of
// every site trusting the running step's identity, and a reviewer of the caller
// can see the decision without opening this file.
func LogActionEntryInheritingProvenance(ctx context.Context, params ActionParams, entry agenterrors.Entry, logger *zap.Logger) bool {
	resolveProvenance(params, &entry, true, logger)
	return agenterrors.Write(ctx, params.DB, logger, entry)
}

// LogActionFindings persists a set of findings an action computed and needs to
// survive an await — one row each, identity shared, filed under the EXECUTING
// step's identity. Call BEFORE the dispatch, so a failed send cannot unrecord
// what the action found. Put the difference between the returns in your audit
// rather than letting a lost row read as a recorded one:
//
//	attempted, recorded := LogActionFindings(ctx, params, siteID, domain, action, findings, logger)
//	audit["conditions_recorded"] = recorded
//	if attempted != recorded { audit["conditions_lost"] = attempted - recorded }
func LogActionFindings(ctx context.Context, params ActionParams, siteID, domain, action string, findings []agenterrors.Finding, logger *zap.Logger) (attempted, recorded int) {
	return LogActionEntryFindingsInheritingProvenance(ctx, params, agenterrors.Entry{
		SiteID: siteID,
		Domain: domain,
		Action: action,
	}, findings, logger)
}

// LogActionEntryFindings is LogActionFindings for a site that files under its
// OWN provenance — the findings form of LogActionEntry, with the same merge
// semantics and the same refusal to inherit a provenance the caller did not
// name.
func LogActionEntryFindings(ctx context.Context, params ActionParams, base agenterrors.Entry, findings []agenterrors.Finding, logger *zap.Logger) (attempted, recorded int) {
	return recordFindings(ctx, params, base, findings, false, logger)
}

// LogActionEntryFindingsInheritingProvenance is the findings form of
// LogActionEntryInheritingProvenance: the declared opt-in for a batch that
// belongs to the running step.
func LogActionEntryFindingsInheritingProvenance(ctx context.Context, params ActionParams, base agenterrors.Entry, findings []agenterrors.Finding, logger *zap.Logger) (attempted, recorded int) {
	return recordFindings(ctx, params, base, findings, true, logger)
}

// recordFindings resolves the shared identity once and delegates the per-row
// write. The unattributed diagnostics need per-finding handling: RecordFindings
// REPLACES the base Context with each Finding's own, so a diagnostic left on
// the base would be silently dropped for every row.
func recordFindings(ctx context.Context, params ActionParams, base agenterrors.Entry, findings []agenterrors.Finding, inheritProvenance bool, logger *zap.Logger) (attempted, recorded int) {
	resolveProvenance(params, &base, inheritProvenance, logger)
	if base.AgentType == unattributedAgentType {
		runningAgentType, runningStepName := runningStepProvenance(params)
		annotated := make([]agenterrors.Finding, len(findings))
		for i, f := range findings {
			f.Context = withUnattributedDiagnostics(f.Context, runningAgentType, runningStepName)
			annotated[i] = f
		}
		findings = annotated
	}
	return agenterrors.RecordFindings(ctx, params.DB, logger, base, findings)
}
