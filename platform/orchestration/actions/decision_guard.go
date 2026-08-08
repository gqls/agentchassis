// FILE: platform/orchestration/actions/decision_guard.go
//
// Decision records (RFC_015): a decision row in doc_notes (categories ?
// 'decision') may carry a fenced ```covers block naming the pages/slots it
// protects. CheckDecisionCoverage answers "is this write touching a
// decision-covered slot, and which decisions cover it?" — the citation gate
// at the write seams refuses a covered write unless the work item NAMES one
// of the covering decisions (acknowledges_decision / supersedes_decision).
//
// Allow change, forbid regression: any agent may change a covered slot by
// citing the decision (visible, auditable); no agent may change what it did
// not know existed. This deliberately mirrors the lock gate's skip-result
// semantics (bugs_open/058) — a refusal is a successful skip, never a
// failing orchestration, because the state is one only a citing item (or a
// superseding decision) can change.
//
// SELF-SCOPING AND INERT BY DEFAULT: sites with no decision rows, and slots
// no ```covers block names, are entirely unaffected. That is the
// additive-and-inert shape of the 2026-07-29 owner ruling.
//
// covers block shape, inside the decision row's body:
//
//	```covers
//	{"pages": ["index"], "slots": ["brief-explanation"]}
//	```
//
// pages/slots match exactly; "*" matches anything; an absent/empty slots
// list means every slot on the named pages.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// DecisionCoverage names one decision that covers a write target.
type DecisionCoverage struct {
	Key    string // doc_notes.subject_key, e.g. "D-004-guide-copy-hand-authored"
	NoteID string // doc_notes.id, for the refusal message
}

var fencedBlockRe = regexp.MustCompile("(?s)```([a-z_]+)\\s*\\n(.*?)```")

// ExtractFencedBlock returns the body of the first fenced block with the
// given tag (e.g. "covers", "guard") from a decision row's prose body, or "".
func ExtractFencedBlock(body, tag string) string {
	for _, m := range fencedBlockRe.FindAllStringSubmatch(body, -1) {
		if m[1] == tag {
			return strings.TrimSpace(m[2])
		}
	}
	return ""
}

type decisionCovers struct {
	Pages []string `json:"pages"`
	Slots []string `json:"slots"`
}

func matchesList(list []string, value string) bool {
	for _, v := range list {
		if v == "*" || strings.EqualFold(v, value) {
			return true
		}
	}
	return false
}

// CheckDecisionCoverage returns the active decisions whose ```covers block
// matches the given page (and slot, when the block names slots). A row with
// no covers block covers nothing — prose-only decisions steer and document
// but never gate. Errors are returned so callers can decide their own
// fail-open/fail-closed posture; the section-editor gate fails OPEN with a
// warning (matching the lock gate's posture at the same seam).
func CheckDecisionCoverage(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, slotName string) ([]DecisionCoverage, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, subject_key, body
		FROM doc_notes
		WHERE site_id = $1
		  AND categories ? 'decision'
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var covered []DecisionCoverage
	for rows.Next() {
		var id, key, body string
		if err := rows.Scan(&id, &key, &body); err != nil {
			return nil, err
		}
		block := ExtractFencedBlock(body, "covers")
		if block == "" {
			continue
		}
		var c decisionCovers
		if json.Unmarshal([]byte(block), &c) != nil || len(c.Pages) == 0 {
			continue // malformed covers block gates nothing — it still steers as prose
		}
		if !matchesList(c.Pages, pageName) {
			continue
		}
		if len(c.Slots) > 0 && !matchesList(c.Slots, slotName) {
			continue
		}
		covered = append(covered, DecisionCoverage{Key: key, NoteID: id})
	}
	return covered, rows.Err()
}

// CitationSatisfies reports whether a work item's citation (the
// acknowledges_decision / supersedes_decision input, possibly a
// comma-separated list) names at least one of the covering decisions.
func CitationSatisfies(citation string, covered []DecisionCoverage) bool {
	if strings.TrimSpace(citation) == "" {
		return false
	}
	for _, part := range strings.Split(citation, ",") {
		part = strings.TrimSpace(part)
		for _, c := range covered {
			if strings.EqualFold(part, c.Key) {
				return true
			}
		}
	}
	return false
}

// CoveredKeys renders the covering decision keys for refusal messages.
func CoveredKeys(covered []DecisionCoverage) string {
	keys := make([]string, len(covered))
	for i, c := range covered {
		keys[i] = c.Key
	}
	return strings.Join(keys, ", ")
}
