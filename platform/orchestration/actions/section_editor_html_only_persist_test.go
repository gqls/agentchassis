// Pins the invariant the rendered_html_transform edit type's safety argument
// rests on (council corr b72a4029 r2, editquality medium): calling
// updatePageComponentAfterEdit with a NIL content_data map must take the
// html-only UPDATE branch — the executed statement must not mention
// content_data AT ALL. If a refactor ever made nil mean "SET content_data =
// NULL", every transform edit would destroy the ported components' 215-byte
// provenance payload while reporting success, and no other test would notice:
// the action result carries only the html. The claim "the transform is
// structurally unable to write content_data" is exactly as strong as this
// test — which is why the assertion is a NEGATIVE match on the captured SQL
// text, not an args-arity check (an inline `content_data = NULL` binds no
// extra arg and would sail past arity).

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// captureSQLMatcher accepts every statement and records it, so the test can
// assert on the SQL text directly — including the negative half sqlmock's
// presence-regex matching cannot express.
type captureSQLMatcher struct{ queries *[]string }

func (m captureSQLMatcher) Match(expectedSQL, actualSQL string) error {
	*m.queries = append(*m.queries, actualSQL)
	return nil
}

func TestUpdatePageComponentAfterEdit_NilContentDataNeverTouchesContentData(t *testing.T) {
	var queries []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureSQLMatcher{queries: &queries}))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

	if err := updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<p>new</p>", nil); err != nil {
		t.Fatalf("updatePageComponentAfterEdit(nil content_data): %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected exactly 1 statement, got %d", len(queries))
	}
	sql := queries[0]
	if !strings.Contains(sql, "rendered_html") {
		t.Errorf("html-only branch must set rendered_html; got: %s", sql)
	}
	if strings.Contains(sql, "content_data") {
		t.Errorf("NIL content_data must be UNTOUCHED, but the statement names it — "+
			"the rendered_html_transform safety argument is void: %s", sql)
	}
}

// The positive control: with a NON-nil map the statement MUST write
// content_data — proving the negative assertion above discriminates the two
// branches rather than passing on any UPDATE shape.
func TestUpdatePageComponentAfterEdit_WithContentDataWritesIt(t *testing.T) {
	var queries []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureSQLMatcher{queries: &queries}))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

	if err := updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<p>new</p>",
		map[string]interface{}{"body": "x"}); err != nil {
		t.Fatalf("updatePageComponentAfterEdit(with content_data): %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected exactly 1 statement, got %d", len(queries))
	}
	if !strings.Contains(queries[0], "content_data") {
		t.Errorf("non-nil content_data must be written; got: %s", queries[0])
	}
}
