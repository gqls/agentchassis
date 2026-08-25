package discovery_checks

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The verifier is deliberately UNREGISTERED (see the file header for the
// five-step sequence). These exercise it directly, so it is not the
// "helper with no callers that looks like a finished refactor" shape — and so
// that whoever performs the registration inherits a proven predicate rather than
// one nobody has run.

// ⚠ THE REGRESSION GUARD THAT MATTERS MOST. The component lookup must resolve on
// the LIFECYCLE axis, not on build_status = 'deployed'. Mirroring the detector's
// filter is bugs_closed/367 rebuilt one layer up: a finding about a non-deployed
// component would resolve nothing and be certified as fixed.
//
// It is the EXPECTATION that carries the predicate, not the returned rows —
// sqlmock hands back whatever the test queued regardless of the statement, so a
// values-only assertion proves the plumbing and nothing about the query.
const wantLifecyclePredicate = `COALESCE\(pc.build_status, 'pending'\) <> 'removed'`

func rfmTarget(spec map[string]interface{}) VerifyTarget {
	return VerifyTarget{ItemID: uuid.New(), SiteID: uuid.New(), ItemType: "required_fields_missing", Spec: spec}
}

func fullSpec() map[string]interface{} {
	return map[string]interface{}{
		"page_name": "about", "slot_name": "hero", "component_id": "c-1",
		"page_id": "p-1", "component_function": "hero", "reason": "…",
		"missing_fields": []interface{}{"product_name"},
	}
}

func expectPage(mock sqlmock.Sqlmock, suppressed bool) {
	mock.ExpectQuery(`SELECT p.id::text`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "suppressed"}).AddRow(uuid.New().String(), suppressed))
}

func expectComponent(mock sqlmock.Sqlmock, schema, content string, locked, runtimeFill bool) {
	mock.ExpectQuery(wantLifecyclePredicate).
		WillReturnRows(sqlmock.NewRows([]string{"input_schema", "content_data", "locked", "runtime_fill"}).
			AddRow(schema, content, locked, runtimeFill))
}

// bugs_open/213: the scope test must disclaim what this predicate cannot resolve.
func TestGradesRequiredFieldsMissing(t *testing.T) {
	if speaks, why := gradesRequiredFieldsMissing(rfmTarget(fullSpec())); !speaks {
		t.Errorf("the post-deploy detector's own shape must be graded, got %q", why)
	}

	// The render-time producer (bugs_closed/342) writes page_name+slot_name but no
	// component_id, about components that are not stored. Re-resolving a slot for
	// one of those finds nothing, and the "no component at that slot" arm would
	// read that absence as a repair.
	noComponent := fullSpec()
	delete(noComponent, "component_id")
	speaks, why := gradesRequiredFieldsMissing(rfmTarget(noComponent))
	if speaks {
		t.Error("an item with no component_id was NOT filed about a stored component — grading it " +
			"reproduces bugs_open/213 exactly")
	}
	if !strings.Contains(why, "component_id") {
		t.Errorf("the disclaimer must say what is missing; an operator reads it off a blocked item. got %q", why)
	}

	for _, drop := range []string{"page_name", "slot_name", "missing_fields"} {
		s := fullSpec()
		delete(s, drop)
		if speaks, _ := gradesRequiredFieldsMissing(rfmTarget(s)); speaks {
			t.Errorf("an item with no %s must be disclaimed — there is nothing to re-resolve or re-check", drop)
		}
	}
}

func TestVerifyRequiredFieldsMissing_PredicateStillFailingBlocks(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectPage(mock, false)
	expectComponent(mock,
		`{"fields":{"product_name":{"source":"llm","required":true}}}`, `{}`, false, false)

	got, err := VerifyRequiredFieldsMissingResolved(context.Background(), db, rfmTarget(fullSpec()), zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if got.Resolved {
		t.Fatal("a component whose required llm field is still absent has NOT been repaired")
	}
	if !strings.Contains(got.Detail, "product_name") {
		t.Errorf("the detail must name the field that is still missing, got %q", got.Detail)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// THE NEGATIVE CONTROL, without which the test above would pass for a verifier
// that refuses everything.
func TestVerifyRequiredFieldsMissing_PopulatedResolves(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectPage(mock, false)
	expectComponent(mock,
		`{"fields":{"product_name":{"source":"llm","required":true}}}`,
		`{"product_name":"Gripper 3000"}`, false, false)

	got, err := VerifyRequiredFieldsMissingResolved(context.Background(), db, rfmTarget(fullSpec()), zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Resolved {
		t.Fatalf("a populated required field IS the repair, got %q", got.Detail)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// Every "resolved" arm must rest on a positive fact, never on a failed lookup.
func TestVerifyRequiredFieldsMissing_PositiveAbsenceArms(t *testing.T) {
	cases := []struct {
		name  string
		setup func(sqlmock.Sqlmock)
		want  string
	}{
		{"page gone", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT p.id::text`).WillReturnError(nil).
				WillReturnRows(sqlmock.NewRows([]string{"id", "suppressed"}))
		}, "no longer exists"},
		{"slot suppressed", func(m sqlmock.Sqlmock) { expectPage(m, true) }, "suppressed_sections"},
		{"no live component at the slot", func(m sqlmock.Sqlmock) {
			expectPage(m, false)
			m.ExpectQuery(wantLifecyclePredicate).
				WillReturnRows(sqlmock.NewRows([]string{"input_schema", "content_data", "locked", "runtime_fill"}))
		}, "no live component"},
		{"component locked (accept-as-is)", func(m sqlmock.Sqlmock) {
			expectPage(m, false)
			expectComponent(m, `{"fields":{}}`, `{}`, true, false)
		}, "locked"},
		{"runtime-fill slot", func(m sqlmock.Sqlmock) {
			expectPage(m, false)
			expectComponent(m, `{"fields":{}}`, `{}`, false, true)
		}, "runtime-fill"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			c.setup(mock)
			got, err := VerifyRequiredFieldsMissingResolved(context.Background(), db, rfmTarget(fullSpec()), zap.NewNop())
			if err != nil {
				t.Fatalf("this arm must resolve, not error: %v", err)
			}
			if !got.Resolved {
				t.Fatalf("expected resolved, got %q", got.Detail)
			}
			if !strings.Contains(got.Detail, c.want) {
				t.Errorf("detail must name the positive fact (%q), got %q", c.want, got.Detail)
			}
		})
	}
}

// RFC_017: "I could not check" must never read as "I checked and it is fixed".
func TestVerifyRequiredFieldsMissing_UnreadableFailsClosed(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectPage(mock, false)
	expectComponent(mock, `{not json`, `{}`, false, false)

	got, err := VerifyRequiredFieldsMissingResolved(context.Background(), db, rfmTarget(fullSpec()), zap.NewNop())
	if err == nil {
		t.Fatalf("an unparseable schema must ERROR (fail-closed under RFC_017), got resolved=%v %q",
			got.Resolved, got.Detail)
	}
	if got.Resolved {
		t.Error("a failed verification must never report Resolved:true")
	}
}
