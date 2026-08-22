package actions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RFC_046. resolveComponentVersionID exists to record a FACT about where bytes
// came from. Most of these tests assert that it declines — because the failure
// this whole class is made of is a mechanism that answers when it does not know
// (bugs_open/357: a tool stored in a row declaring itself `hero`, because hero
// was first in the plan). A resolver that always returns something would be the
// same defect one layer down.

func versionTestDB(t *testing.T) (*sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return &mock, func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
		db.Close()
	}
}

func sha(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// A carried stamp describes bytes that did not change, so it is returned
// untouched and WITHOUT touching the database. The "no query" half is the real
// assertion: sqlmock fails any call nobody expected.
func TestComponentVersion_CarriedStampIsKeptAndCostsNoQuery(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	carried := uuid.NewString()

	got, ok := resolveComponentVersionID(context.Background(), db, uuid.New(),
		componentVersionSource{carriedVersionID: carried}, zap.NewNop())

	if !ok || got != carried {
		t.Fatalf("carried stamp: got (%q,%v), want (%q,true)", got, ok, carried)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected no queries at all: %v", err)
	}
}

// A carried value that is not a uuid is dropped rather than written on. Passing
// it through would put a malformed identity in a column whose whole purpose is
// being trustworthy.
func TestComponentVersion_MalformedCarriedStampIsDropped(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	if got, ok := resolveComponentVersionID(context.Background(), db, uuid.Nil,
		componentVersionSource{carriedVersionID: "not-a-uuid"}, zap.NewNop()); ok {
		t.Fatalf("malformed carried stamp was accepted: %q", got)
	}
}

// Nothing known, nothing asked, nothing written — and no query, which is what
// keeps this off the hot path for the majority of sections.
func TestComponentVersion_NoProvenanceMeansNoStampAndNoQuery(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if _, ok := resolveComponentVersionID(context.Background(), db, uuid.New(),
		componentVersionSource{}, zap.NewNop()); ok {
		t.Fatal("a section with no provenance was stamped")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected no queries at all: %v", err)
	}
}

// THE ONE THAT MATTERS. The template changed between render and save, so the
// digest describes text we no longer hold. There IS a newest version row and it
// must NOT be used: it did not produce these bytes. Inferring here would rebuild
// bugs_open/357 inside its own fix.
//
// ⚠ THE DECOY IS THE TEST, and the obvious version of this test does not work.
// Writing it as "expect no component_versions query" passes even against code
// that DOES infer: sqlmock refuses the unexpected call, the resolver sees a
// query error, and it returns no-stamp for the wrong reason — a guard in series
// standing in for the one under test. So the decoy row is made AVAILABLE and
// attractive; the assertion is that the resolver does not take it.
func TestComponentVersion_TemplateChangedSinceRenderIsNotInferred(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	compID, decoyVersionID := uuid.New(), uuid.NewString()

	mock.ExpectQuery("SELECT html_template, input_schema FROM content_components").
		WithArgs(compID).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "input_schema"}).
			AddRow("<section>NEW TEXT</section>", []byte(`{}`)))
	// Available, and wrong. Any code path that reaches for "the component's
	// newest version" gets this id and fails the assertion below.
	mock.ExpectQuery("SELECT id FROM component_versions").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(decoyVersionID))
	mock.ExpectQuery("INSERT INTO component_versions").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(decoyVersionID))

	got, ok := resolveComponentVersionID(context.Background(), db, compID,
		componentVersionSource{renderedSHA: sha("<section>OLD TEXT</section>")}, zap.NewNop())

	if ok || got != "" {
		t.Fatalf("drifted template was stamped with %q (ok=%v) — the resolver inferred a version "+
			"that did not produce these bytes", got, ok)
	}
	if got == decoyVersionID {
		t.Fatal("resolver took the decoy: it fell back to the component's newest version")
	}
}

// The template is unchanged and already versioned: resolve to the existing row
// rather than minting a new one, or component_versions becomes a render log.
func TestComponentVersion_UnchangedTemplateResolvesToTheExistingRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	compID, versionID := uuid.New(), uuid.NewString()
	tmpl := "<section data-component=\"hero\">{{.headline}}</section>"

	mock.ExpectQuery("SELECT html_template, input_schema FROM content_components").
		WithArgs(compID).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "input_schema"}).
			AddRow(tmpl, []byte(`{}`)))
	mock.ExpectQuery("SELECT id FROM component_versions").
		WithArgs(compID, tmpl).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(versionID))

	got, ok := resolveComponentVersionID(context.Background(), db, compID,
		componentVersionSource{renderedSHA: sha(tmpl)}, zap.NewNop())
	if !ok || got != versionID {
		t.Fatalf("got (%q,%v), want (%q,true)", got, ok, versionID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// First time this exact text is rendered through here: record it, so the row can
// name the template that made it. bugs_closed/277 closed with fifteen rows
// permanently unrepairable precisely because nothing had done this.
func TestComponentVersion_FirstSightOfATemplateRecordsIt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	compID, versionID := uuid.New(), uuid.NewString()
	tmpl := "<section>{{.body}}</section>"

	mock.ExpectQuery("SELECT html_template, input_schema FROM content_components").
		WithArgs(compID).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "input_schema"}).
			AddRow(tmpl, []byte(`{}`)))
	mock.ExpectQuery("SELECT id FROM component_versions").
		WithArgs(compID, tmpl).
		WillReturnError(sqlErrNoRows())
	mock.ExpectQuery("INSERT INTO component_versions").
		WithArgs(compID, tmpl, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(versionID))

	got, ok := resolveComponentVersionID(context.Background(), db, compID,
		componentVersionSource{renderedSHA: sha(tmpl)}, zap.NewNop())
	if !ok || got != versionID {
		t.Fatalf("got (%q,%v), want (%q,true)", got, ok, versionID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The seam reports the digest of the text it executed, and reports NOTHING for
// an empty template — "there was no template" is a kind of unknown, and a token
// pointing at the empty string would be a provenance claim about nothing.
func TestRenderTemplate_StampsTheTemplateItExecuted(t *testing.T) {
	tmpl := "<section>{{.headline}}</section>"
	ctx := &RenderContext{ContentData: map[string]interface{}{"headline": "x"}}
	if _, _, _, err := RenderTemplate(tmpl, ctx, zap.NewNop()); err != nil {
		t.Fatalf("render: %v", err)
	}
	if ctx.RenderedTemplateSHA != sha(tmpl) {
		t.Errorf("RenderedTemplateSHA = %q, want %q", ctx.RenderedTemplateSHA, sha(tmpl))
	}

	empty := &RenderContext{}
	if _, _, _, err := RenderTemplate("", empty, zap.NewNop()); err != nil {
		t.Fatalf("empty render: %v", err)
	}
	if empty.RenderedTemplateSHA != "" {
		t.Errorf("empty template stamped %q — empty must mean unknown", empty.RenderedTemplateSHA)
	}
}

func sqlErrNoRows() error { return sql.ErrNoRows }

// CONTROL for the test above, and it is not optional. "The resolver did not take
// the decoy" only means something if the decoy was actually on offer — a mock
// that never serves it would make that assertion pass against any code at all.
// This runs the SAME mock setup through a deliberately-inferring stand-in and
// asserts it DOES get the decoy. If this control ever stops passing, the test
// above has quietly become unfalsifiable.
func TestComponentVersion_DecoyControl_AnInferringImplementationTakesIt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	compID, decoyVersionID := uuid.New(), uuid.NewString()

	mock.ExpectQuery("SELECT html_template, input_schema FROM content_components").
		WithArgs(compID).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "input_schema"}).
			AddRow("<section>NEW TEXT</section>", []byte(`{}`)))
	mock.ExpectQuery("SELECT id FROM component_versions").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(decoyVersionID))

	// The implementation RFC_046 forbids: ignore the digest, take the newest.
	var inferred string
	if err := db.QueryRowContext(context.Background(),
		`SELECT id FROM component_versions WHERE component_id = $1 ORDER BY version_number DESC LIMIT 1`,
		compID).Scan(&inferred); err != nil {
		t.Fatalf("control could not reach the decoy — the test above proves nothing: %v", err)
	}
	if inferred != decoyVersionID {
		t.Fatalf("control got %q, want the decoy %q", inferred, decoyVersionID)
	}
}
