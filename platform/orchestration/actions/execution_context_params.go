// FILE: platform/orchestration/actions/execution_context_params.go
//
// The `$ctx.` parameter namespace: a way for workflow SQL to bind the identity
// of the RUN THAT IS EXECUTING IT, rather than a value out of collected_data.
//
// Why this exists (bugs_open/124). A dispatch loop claims a queued row and runs
// it. Afterwards there is no way back from the row to the run: the row carries
// whatever correlation its CREATOR minted, while every artefact the run writes
// is keyed on the run's own envelope correlation
// (params.ExecutionContext.CorrelationID — see diagnose_assemble_bundle_action.go).
// For `needs_diagnosis` that made the documented "item, bundles and doc_notes
// all join on one key" property false for every loop-dispatched item, and
// silently: `spec.correlation_id` still LOOKED like a key, it just named nothing.
//
// The fix a queue needs is for the claim itself to stamp the claiming run's id
// onto the row — one atomic UPDATE, no second write to go missing. But
// `query_database` could only bind values it could find in collected_data, and
// a run's own correlation is not in collected_data: it lives in the execution
// context. So every lane that wanted this had to grow a bespoke Go action.
//
// This is the generic seam instead. It is deliberately NOT diagnose-specific:
// "which run picked this row up" is a question every queue-driven workflow has.
//
//	'params', jsonb_build_array('$ctx.correlation_id')
//	'query',  'UPDATE q SET claimed_by_run = $1 WHERE id = ... RETURNING id'
//
// Contract, matching the existing collected_data params exactly:
//   - only paths with the `$ctx.` prefix take this branch; every other path
//     resolves out of collected_data as before. Additive by construction — no
//     existing path can start with `$`, so no existing workflow changes shape.
//   - an unknown key is an ERROR, not an empty string. A silently-empty bind is
//     how a stamp column fills with '' and looks populated.
//   - an EMPTY value is an error too, for the same reason. If a field is
//     legitimately absent (a root orchestration has no parent), do not bind it:
//     there is no "optional" mode on purpose.

package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// executionContextParamPrefix is the reserved namespace. Chosen over a bare
// dotted path (`execution.correlation_id`) because a step's output_field could
// legitimately be named `execution` and would then shadow it; no collected_data
// key can begin with '$'.
const executionContextParamPrefix = "$ctx."

// executionContextParam resolves a `$ctx.<field>` parameter path against the
// run's execution context.
//
// Returns isCtx=false for any path outside the namespace, so callers fall
// through to their normal resolution. When isCtx is true the error is
// authoritative: a bad key or an empty value fails the step rather than binding
// a placeholder.
func executionContextParam(execCtx *types.ExecutionContext, path string) (value string, isCtx bool, err error) {
	if !strings.HasPrefix(path, executionContextParamPrefix) {
		return "", false, nil
	}
	field := strings.TrimPrefix(path, executionContextParamPrefix)

	if execCtx == nil {
		return "", true, fmt.Errorf("query param path %q needs the execution context, which is nil", path)
	}

	available := map[string]string{
		"correlation_id":          execCtx.CorrelationID,
		"orchestration_id":        execCtx.OrchestrationID,
		"parent_orchestration_id": execCtx.ParentOrchestrationID,
		"orchestration_name":      execCtx.OrchestrationName,
		"client_id":               execCtx.ClientID,
		"request_id":              execCtx.RequestID,
		"step_name":               execCtx.StepName,
		"group_id":                execCtx.GroupID,
	}

	v, ok := available[field]
	if !ok {
		return "", true, fmt.Errorf("query param path %q: unknown execution-context field %q (known: %s)",
			path, field, strings.Join(sortedKeys(available), ", "))
	}
	if v == "" {
		return "", true, fmt.Errorf("query param path %q resolved to an empty %s — bind it only where the run actually carries one",
			path, field)
	}
	return v, true, nil
}

// sortedKeys keeps the "known fields" half of an error message stable, so the
// same mistake reads the same way twice.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
