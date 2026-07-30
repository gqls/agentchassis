// FILE: platform/orchestration/actions/diagnose_artifact_count.go
//
// One counter for every bounded loop in the diagnosis/fix machinery.
//
// The pattern is the same wherever an LLM loop has to stop: write a durable
// artefact for this round, then count the artefacts of that kind for THIS RUN and
// compare against a cap. Sourcing the counter from the artefact rather than from
// workflow state is deliberate — the workflow's collected_data field is overwritten
// on each pass through the loop, so it cannot hold a count across iterations.
//
// Extracted 2026-07-30 in answer to the council's reuse seat on bugs_open/099
// candidate 2. Before extraction there were three hand-written copies of this
// query: two in diagnose_council_decide_action.go (the revise round counter and the
// reframe-once bookkeeping) and a third about to be added for the plan-repair loop.
// The seat's objection was precise — the risk is not the duplication itself but
// that "if the council-decide counter's scoping or fail-closed semantics are later
// fixed, this new copy will silently not get the fix". One function, three callers,
// so a scoping fix lands everywhere at once.
//
// SCOPING IS LOAD-BEARING, and it is why this takes an orchestration id at all.
// diagnose_council_decide_action.go records the reason: the correlation belongs to
// the DIAGNOSIS and accumulates artefacts across re-runs of the proposer, so a
// count scoped by correlation alone starts a fresh run part-way through its budget
// (bugs_open/033's round-counting scope bug, FIX-033). Every caller must pass the
// run id.
//
// LANDMINE, preserved rather than fixed here: orchID is interface{} because the
// existing callers pass nullIfEmpty(...), which yields a NIL for an empty id — and
// `orchestration_id = $2` NEVER matches a NULL, so such a call counts 0 for ever
// and 0 reads as "first round". This extraction is a pure refactor and deliberately
// does NOT change that behaviour for its existing callers; a caller that needs the
// count to be trustworthy must assert a non-empty id itself, as
// planValidationRefusal does. See LANDMINES.md.
package actions

import (
	"context"
	"database/sql"
	"fmt"
)

// countRunArtifacts counts diagnosis_artifacts rows of one kind for a single loop
// run, optionally narrowed to one metadata key/value.
//
// Pass metaKey == "" for no metadata filter. The metadata comparison is on
// metadata->>key, so the value is always compared as text.
//
// Returns an error rather than a zero on failure, because every caller must decide
// for itself what a counting failure means — and in this machinery the answer is
// always "fail closed", never "assume round one".
func countRunArtifacts(
	ctx context.Context, db *sql.DB,
	correlationID string, orchestrationID interface{},
	kind string, metaKey, metaValue string,
) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("countRunArtifacts: nil database")
	}

	query := `SELECT count(*) FROM diagnosis_artifacts
		 WHERE correlation_id = $1 AND orchestration_id = $2 AND kind = $3`
	args := []interface{}{correlationID, orchestrationID, kind}
	if metaKey != "" {
		query += ` AND metadata->>$4 = $5`
		args = append(args, metaKey, metaValue)
	}

	var n int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
