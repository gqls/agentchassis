// FILE: platform/orchestration/actions/work_item_retraction.go
//
// The SHARED half of audit-path retraction: a producer that files work items
// being given the authority to CLOSE them again from its own later
// re-observation, on the discovery path (WII-009's seam, reached from outside
// the discovery-check framework by an action rather than a check_*.go).
//
// WHY IT IS SHARED, AND WHO ASKED FOR IT. WII-016's council round
// (corr a43b63d6-da35-4136-9471-88ec6ace799a, `architecture` seat, advisory,
// low) recorded that write_render_audit_findings was the SECOND producer to
// hand-roll the same three moves inline — build the still-failing set BEFORE
// the filing filters, load the open rows, close what is absent through
// resolveWorkItems — and that "a third should extract a shared helper rather
// than copy-paste it". `dark_section_audit` (bugs_open/213 D1 half two) is the
// third. This is that helper.
//
// WHAT IS SHARED IS THE PART THAT IS EASY TO GET WRONG, not the easy part:
//
//  1. AN UNAVAILABLE OBSERVATION MUST BE INERT, NEVER WIDE. Every caller has
//     some way of failing to observe anything at all — an adapter too old to
//     send pages_audited, an LLM reply whose shape was not recognised. A
//     retraction rule of the form "close what is absent" reads a failed
//     observation as "everything is fixed" and closes the lot. So `observed`
//     is a REQUIRED argument and false means do nothing, and each caller must
//     say in its own words what makes an observation available.
//
//  2. THE READ MUST FULLY DRAIN BEFORE THE FIRST WRITE. A database/sql Tx is
//     ONE connection: retracting while the candidate cursor is still streaming
//     deadlocks the transaction against itself. loadAuditRetractionCandidates
//     therefore returns a slice, never rows, and closes its cursor before
//     returning — the caller cannot get this wrong by accident.
//
//  3. THE PARK MUST NOT DRAIN SILENTLY. `deferred` is NOT in
//     workItemClosedStatuses, so a retraction closes PARKED items too — a
//     stated decision (WII-016), not a side effect, because a retraction is
//     evidence-stamped and the park exists to prevent UNGRADED completions.
//     The parked count is returned separately so the draining is visible.
//
// WHAT IS DELIBERATELY NOT SHARED IS THE JUDGEMENT: which rows a run's
// observation covers, and what counts as "the defect is gone". Those differ
// completely between callers — the render audit knows the identities of the
// pages it measured and can retract on ONE silence, because a browser
// measured a contrast ratio; the design audit cannot resolve its own
// spec.page_name to a page at all and retracts on THREE, because its silence
// is an LLM's. Both live in the caller's `decide` function, next to the
// evidence that licenses them.
//
// ⚠ THE TRAP EVERY ADOPTER INHERITS (WII-016's, restated because it is the one
// that destroys data): THE STILL-FAILING SET MUST BE BUILT BEFORE EVERY FILTER
// THAT DECIDES WHAT TO FILE. A finding dropped by a lock, a cap or a dedup
// check was MEASURED AND IS STILL FAILING; it simply could not be acted on.
// Building the set from the items a run FILED reads "not filed" as "fixed" and
// closes live defects. That set belongs to the caller, and this helper cannot
// check it for you.

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"go.uber.org/zap"
)

// auditRetractionCandidate is one OPEN row of the retracting producer's item_type.
//
// Spec and Result are carried raw because the two live callers need different
// things from them and neither wants a second query per row: the design-audit
// path reads spec.audit_source to keep one producer from closing another's
// findings, and reads result.retraction to carry its silence streak.
type auditRetractionCandidate struct {
	ID     uuid.UUID
	Key    string
	Status string
	Spec   []byte
	Result []byte
}

// retractionVerdict is the caller's per-row decision. The three values are
// deliberately distinct rather than a bool: "this run did not look at it" and
// "this run looked and it is still broken" are the same ACTION but different
// FACTS, and a caller that cannot tell them apart is one measurement away from
// closing rows it never observed.
type retractionVerdict int

const (
	// retractionOutOfScope — this run's observation does not cover this row.
	// Leave it alone; it is not evidence of anything.
	retractionOutOfScope retractionVerdict = iota
	// retractionStillFailing — observed, and the defect is still present.
	retractionStillFailing
	// retractionResolved — observed, and the defect is gone. Close it.
	retractionResolved
)

// loadAuditRetractionCandidates reads every OPEN row of one item_type for one site,
// fully, and closes its cursor before returning. See point (2) in the header:
// the slice return is the guard, not a style preference.
//
// The status predicate is workItemClosedStatuses, so `failed`, `unresolved`
// and `deferred` rows ARE candidates — that is RFC_010 Decision 2's ruling and
// the whole point of retraction (a row that says "we gave up" is exactly the
// one a fresh positive observation should close).
//
// itemType is INTERPOLATED rather than bound, matching sqlInList directly
// above it in the same query: every caller passes a compile-time constant from
// its own roster, never user input. Same obligation as sqlInList's — callers
// must supply already-safe const values.
func loadAuditRetractionCandidates(
	ctx context.Context,
	tx *sql.Tx,
	siteID uuid.UUID,
	itemType string,
) ([]auditRetractionCandidate, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
		SELECT id,
		       COALESCE(item_key, ''),
		       status,
		       COALESCE(spec,   '{}'::jsonb)::text,
		       COALESCE(result, '{}'::jsonb)::text
		FROM site_work_items
		WHERE site_id   = $1
		  AND item_type = '%s'
		  AND status NOT IN (%s)
	`, itemType, sqlInList(workItemClosedStatuses)), siteID)
	if err != nil {
		return nil, fmt.Errorf("load open %s items: %w", itemType, err)
	}
	defer rows.Close()

	var out []auditRetractionCandidate
	for rows.Next() {
		var c auditRetractionCandidate
		if sErr := rows.Scan(&c.ID, &c.Key, &c.Status, &c.Spec, &c.Result); sErr != nil {
			return nil, fmt.Errorf("scan open %s item: %w", itemType, sErr)
		}
		out = append(out, c)
	}
	if rErr := rows.Err(); rErr != nil {
		return nil, fmt.Errorf("read open %s items: %w", itemType, rErr)
	}
	return out, nil
}

// retractResolvedAuditFindings closes the candidates the caller's `decide` reports
// as positively observed to be fixed, and only those.
//
// `observed` is the inert guard of header point (1): false means this run made
// no usable observation, and NOTHING is closed. It is a separate argument from
// the candidate list on purpose — an empty candidate list and an unavailable
// observation both produce zero retractions today, but they mean opposite
// things, and a caller that conflates them will eventually pass an empty
// still-failing set built from a failed parse.
//
// `decide` returns the verdict and, for retractionResolved, the REASON that
// will be stamped into result.reason. An empty reason is refused rather than
// defaulted: resolveWorkItems already treats a blank reason as a bug (a row
// that closed itself with no stated cause is indistinguishable later from one
// a human closed by hand), and failing here names the caller.
func retractResolvedAuditFindings(
	ctx context.Context,
	tx *sql.Tx,
	siteID uuid.UUID,
	checkName string,
	batchID uuid.UUID,
	itemType string,
	observed bool,
	candidates []auditRetractionCandidate,
	decide func(auditRetractionCandidate) (retractionVerdict, string),
	logger *zap.Logger,
) (retracted int, parked int, err error) {
	if !observed || len(candidates) == 0 {
		return 0, 0, nil
	}

	done := map[string]bool{}
	for _, c := range candidates {
		if c.Key == "" || done[c.Key] {
			continue
		}
		verdict, reason := decide(c)
		if verdict != retractionResolved {
			continue
		}
		if reason == "" {
			return 0, 0, fmt.Errorf("retract %s/%s: %s returned retractionResolved with an empty reason",
				itemType, c.Key, checkName)
		}
		done[c.Key] = true

		n, rErr := resolveWorkItems(ctx, tx, siteID, checkName, batchID, checks.ResolvedFinding{
			ItemType: itemType,
			ItemKey:  c.Key,
			Reason:   reason,
		}, logger)
		if rErr != nil {
			return 0, 0, rErr
		}
		retracted += n
		// A row that was `deferred` was PARKED, and the park draining unnoticed
		// is exactly what this counter exists to make visible.
		if n > 0 && c.Status == "deferred" {
			parked += n
		}
	}

	if retracted > 0 {
		logger.Info("Retracted work items no longer reproducing",
			zap.String("check", checkName),
			zap.String("item_type", itemType),
			zap.Int("retracted", retracted),
			zap.Int("parked", parked),
			zap.Int("candidates", len(candidates)))
	}
	return retracted, parked, nil
}
