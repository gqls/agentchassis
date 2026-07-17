// FILE: platform/orchestration/datahelpers/travelling_docs_rekey.go
//
// RekeyTravellingDocs moves a subject's travelling docs (doc_plans + doc_notes)
// from one subject_key to another, inside the caller's transaction. Guard rail
// 2 of the experience loop: a rename that does not move the docs strands the
// PLAN — and its acceptance criteria — under a key nothing sweeps. The vonc
// arena (page renamed to tool-arena while docs stayed at tool-arena-interface,
// invisible to the acceptance ladder for three days) is the proof.
//
// Refuses when BOTH keys hold a current plan: the caller must supersede one
// first. Silently merging would violate idx_doc_plans_current and, worse,
// silently pick a winner.

package datahelpers

import (
	"context"
	"database/sql"
	"fmt"
)

// DocRekeyer is the slice of *sql.Tx / *sql.DB this helper needs.
type DocRekeyer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// RekeyTravellingDocs returns how many plan and note rows moved.
func RekeyTravellingDocs(ctx context.Context, q DocRekeyer, subjectType, oldKey, newKey string) (plansMoved, notesMoved int64, err error) {
	if subjectType == "" || oldKey == "" || newKey == "" || oldKey == newKey {
		return 0, 0, fmt.Errorf("rekey: subject_type plus distinct old/new keys required (type=%q old=%q new=%q)",
			subjectType, oldKey, newKey)
	}

	var oldCurrent, newCurrent int
	err = q.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE subject_key = $2),
			COUNT(*) FILTER (WHERE subject_key = $3)
		FROM doc_plans
		WHERE subject_type = $1 AND is_current
	`, subjectType, oldKey, newKey).Scan(&oldCurrent, &newCurrent)
	if err != nil {
		return 0, 0, fmt.Errorf("rekey: current-plan probe: %w", err)
	}
	if oldCurrent > 0 && newCurrent > 0 {
		return 0, 0, fmt.Errorf("rekey: both %q and %q hold a current %s plan — supersede one first",
			oldKey, newKey, subjectType)
	}

	res, err := q.ExecContext(ctx, `
		UPDATE doc_plans SET subject_key = $3, updated_at = NOW()
		WHERE subject_type = $1 AND subject_key = $2
	`, subjectType, oldKey, newKey)
	if err != nil {
		return 0, 0, fmt.Errorf("rekey doc_plans: %w", err)
	}
	plansMoved, _ = res.RowsAffected()

	res, err = q.ExecContext(ctx, `
		UPDATE doc_notes SET subject_key = $3
		WHERE subject_type = $1 AND subject_key = $2
	`, subjectType, oldKey, newKey)
	if err != nil {
		return plansMoved, 0, fmt.Errorf("rekey doc_notes: %w", err)
	}
	notesMoved, _ = res.RowsAffected()

	return plansMoved, notesMoved, nil
}
