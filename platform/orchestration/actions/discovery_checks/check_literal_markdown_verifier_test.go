// FILE: platform/orchestration/actions/discovery_checks/check_literal_markdown_verifier_test.go
//
// Tests for VerifyLiteralMarkdownResolved — bugs_open/201 SYMPTOM 2's guard.
//
// WHAT THESE ARE FOR, and why the third case is the important one. The bug is a
// handler saga reporting success having written nothing, and complete_work_item
// stamping 'complete' on its word. A verifier only helps if it says NO when the
// defect survives — so the load-bearing assertion is the UNRESOLVED case, not the
// happy path. The happy path passing proves nothing on its own: a verifier that
// always returned Resolved would pass it too.
//
// The zero-rows case is tested because a verifier that answers "no markdown found"
// on a page whose components have been LOST would stamp 'complete' over
// bugs_closed/194's damage class (31 of 106 components NULL on one live site). Both
// states present identically as "nothing to scan"; the only honest answer is to
// refuse to certify.

package discovery_checks

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func literalMarkdownComponentRows(rows ...[3]string) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"slot_name", "rendered_html", "content_data"})
	for _, row := range rows {
		r.AddRow(row[0], row[1], row[2])
	}
	return r
}

func literalMarkdownVerifyTarget() VerifyTarget {
	pageID := uuid.New()
	return VerifyTarget{
		ItemID:   uuid.New(),
		SiteID:   uuid.New(),
		PageID:   &pageID,
		ItemType: "literal_markdown",
		Spec:     map[string]interface{}{"page_name": "how-pricing-works"},
	}
}

// THE LOAD-BEARING CASE: the defect survives, so completion must be refused.
// This is the exact live shape from gaswholesalers.com on 2026-08-06 — an item
// that read status='complete' while `**Decision Engine**` was still in content_data.
func TestVerifyLiteralMarkdownResolved_RefusesWhenMarkdownSurvives(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").WillReturnRows(literalMarkdownComponentRows(
		[3]string{"hero", "<h1>Pricing</h1>", `{"subheadline":"**Decision Engine** pricing"}`},
	))

	res, err := VerifyLiteralMarkdownResolved(context.Background(), db, literalMarkdownVerifyTarget(), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier returned error: %v", err)
	}
	if res.Resolved {
		t.Fatalf("Resolved=true while `**Decision Engine**` is still in content_data — this is exactly the "+
			"symptom-2 stamp the verifier exists to prevent. Detail: %q", res.Detail)
	}
	if !strings.Contains(res.Detail, "hero") {
		t.Errorf("Detail should name the offending slot so a human can act on it; got %q", res.Detail)
	}
}

// The clean case. Weak on its own — see the file header — but it must not report
// a false NEGATIVE, which would strand a correctly-repaired item in 'failed'.
func TestVerifyLiteralMarkdownResolved_PassesWhenClean(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").WillReturnRows(literalMarkdownComponentRows(
		[3]string{"hero", "<h1>Pricing</h1><p>Clear, honest pricing.</p>", `{"subheadline":"Decision engine pricing"}`},
		[3]string{"pricing", "<p>Per unit.</p>", `{"body":"No hidden fees."}`},
	))

	res, err := VerifyLiteralMarkdownResolved(context.Background(), db, literalMarkdownVerifyTarget(), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier returned error: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("Resolved=false on clean content — a correctly-repaired item would burn its attempts "+
			"and strand in 'failed' (the page_rerender trap). Detail: %q", res.Detail)
	}
}

// A zero-row page must NOT be certified. "No markdown found" and "the components
// are gone" are the same observation; only one of them means repaired.
func TestVerifyLiteralMarkdownResolved_RefusesToCertifyAnEmptyPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").WillReturnRows(literalMarkdownComponentRows())

	res, err := VerifyLiteralMarkdownResolved(context.Background(), db, literalMarkdownVerifyTarget(), zap.NewNop())
	if err == nil {
		t.Fatalf("expected an error (cannot verify), got Resolved=%v Detail=%q — certifying a page with no "+
			"scannable components stamps 'complete' over bugs_closed/194's content-loss class",
			res.Resolved, res.Detail)
	}
	if res.Resolved {
		t.Errorf("Resolved must stay false alongside the error; got true")
	}
}

// An item with no page_id cannot be located, so the verifier must say so rather
// than scan the wrong thing or silently pass.
func TestVerifyLiteralMarkdownResolved_RefusesWithoutPageID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	target := literalMarkdownVerifyTarget()
	target.PageID = nil

	if _, err := VerifyLiteralMarkdownResolved(context.Background(), db, target, zap.NewNop()); err == nil {
		t.Fatal("expected an error when the item carries no page_id")
	}
}

// The detector and the verifier must apply the SAME predicate — that is verifiers.go's
// stated contract, and the reason scanComponentRowForMarkdown was extracted rather than
// copied. This pins the shared helper against the guarded patterns the check documents,
// including the two false-positive guards that must NOT fire.
func TestScanComponentRowForMarkdown_MatchesTheDocumentedPredicate(t *testing.T) {
	cases := []struct {
		name, slot, html, contentJSON string
		wantFindings                  bool
	}{
		{"bold in content_data", "hero", "", `{"sub":"**Decision Engine**"}`, true},
		{"heading in content_data", "body", "", `{"sub":"# Pricing"}`, true},
		{"clean", "hero", "<p>All good.</p>", `{"sub":"Decision engine"}`, false},
		// The check's own false-positive discipline: arithmetic is not bold, and a
		// markup-bearing value is not a text-typed field so its backticks are code.
		{"arithmetic is not bold", "body", "", `{"sub":"3 * 4 = 12 and 2**10"}`, false},
		{"code span suppressed inside markup", "body", "", `{"sub":"<p>use ` + "`npm i`" + ` here</p>"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := scanComponentRowForMarkdown(tc.slot, tc.html, tc.contentJSON, zap.NewNop())
			if (len(got) > 0) != tc.wantFindings {
				t.Fatalf("findings=%d, want any=%v — the verifier and the detector share this "+
					"function, so a change here silently changes both", len(got), tc.wantFindings)
			}
		})
	}
}
