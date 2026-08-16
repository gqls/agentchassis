// FILE: platform/orchestration/actions/create_tool_component_adopt_test.go
//
// bugs_open/286: create_tool_component had NO attach-to-existing-page path —
// a bare INSERT INTO pages with no ON CONFLICT and no pre-lookup, so a
// same-URL replacement of an existing tool page (the ported-tool rebuild
// route) died on pages_site_id_name_key, and the error branch deleted the
// just-created component (090 diagnosis CONFIRMED, corr 3050effc). The
// opt-in `adopt_existing_page` step-config key (default OFF, the 2026-08-02
// §2 shape) routes page identity through resolveToolPageIdentity +
// UpsertPageForRole — deploy_tool's machinery — instead.
//
// The three arms pinned here:
//
//	A. flag ON + a LIVE page already holding the role: the action attaches to
//	   THAT page (the link insert targets the existing row's id) and never
//	   writes the pages table.
//	B. flag ABSENT: byte-for-byte the old path — bare INSERT, collision
//	   surfaces as "failed to create tool page", component cleaned up.
//	C. flag ON + NO existing page: the canonical page is created; a later
//	   failure still removes the page this call created.
//
// The guard "an ADOPTED page is never deleted by cleanup" is a NEGATIVE and a
// mock cannot assert one (an unexpected DELETE returns an error the
// fire-and-forget cleanup swallows, and ExpectationsWereMet only checks the
// expected side). It is pinned differentially: C proves cleanup deletes a
// page the call CREATED under the flag; A's expectation list contains no
// pages DELETE; the three lines between them are the review surface.
package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// A fixture that passes both birth gates: the tool-doc header sentinels and
// componentTemplateValid's balanced-tag / ends-cleanly predicate.
const adoptTestToolHTML = `<div class="tool-container"><script>
/* === tool-doc ===
name: Aspect Ratio Calculator
=== /tool-doc === */
function calc() { return 1; }
</script></div>`

func adoptTestParams(db *sql.DB, adopt interface{}) ActionParams {
	cfg := map[string]interface{}{}
	if adopt != nil {
		cfg["adopt_existing_page"] = adopt
	}
	return ActionParams{
		DB:     db,
		Logger: zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{
			Action:   "process",
			StepName: "save_tool",
		},
		StepConfig: models.Step{Config: cfg},
		CollectedData: map[string]interface{}{
			"site_id":      "6b49db8e-d447-4467-8277-4f3018af9897",
			"html_content": adoptTestToolHTML,
			"function":     "tool-aspect-ratio",
			"display_name": "Aspect Ratio Calculator",
		},
	}
}

// expectPreamble mocks everything up to and including the component INSERT:
// site-domain load, the already-exists probe (empty), the component write,
// and the CanonicalisePage url_shape read.
func expectPreamble(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO content_components").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("url_shape").
		WillReturnRows(sqlmock.NewRows([]string{"shape"}))
}

func TestCreateToolComponent_AdoptAttachesToTheLiveSameRolePage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	existingPageID := uuid.New()

	expectPreamble(mock)
	// The adopt arm's own url_shape read (resolveToolPageIdentity argument).
	mock.ExpectQuery("url_shape").
		WillReturnRows(sqlmock.NewRows([]string{"shape"}))
	// The live row keeps its stored identity (bugs_open/080's rule).
	mock.ExpectQuery("SELECT name, url FROM pages").
		WithArgs(sqlmock.AnyArg(), "aspect-ratio", "tool-aspect-ratio").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}).
			AddRow("tool-aspect-ratio", "/tools/aspect-ratio/index.html"))
	// UpsertPageForRole: the name is taken, so DO NOTHING returns no row...
	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	// ...and the conflict branch finds a SHIPPED page already holding the
	// tool role -> PageRoleRefreshed with an empty Refresh set: no page write.
	mock.ExpectBegin()
	mock.ExpectQuery("FOR UPDATE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status", "status", "shipped"}).
			AddRow(existingPageID.String(), "tool", "deployed", "active", true))
	mock.ExpectCommit()
	// The link insert must target the EXISTING page's id. Induced failure so
	// the test ends at the seam under test.
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(existingPageID.String(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)
	// Cleanup removes the component this call created — and nothing else: an
	// adopted page is not this call's to delete (see file comment on why the
	// negative is pinned by Test C's differential rather than the mock).
	mock.ExpectExec("DELETE FROM content_components").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, actErr := CreateToolComponentAction(context.Background(), adoptTestParams(db, true))
	if actErr == nil || !strings.Contains(actErr.Error(), "failed to link component to page") {
		t.Fatalf("want the induced link error, got: %v", actErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateToolComponent_FlagAbsentKeepsTheBareInsertAndItsCollision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectPreamble(mock)
	// The old path, unchanged: a bare INSERT that hits the unique key.
	mock.ExpectExec("INSERT INTO pages").
		WillReturnError(errDuplicatePagesKey{})
	mock.ExpectExec("DELETE FROM content_components").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, actErr := CreateToolComponentAction(context.Background(), adoptTestParams(db, nil))
	if actErr == nil || !strings.Contains(actErr.Error(), "failed to create tool page") {
		t.Fatalf("want the collision surfaced as 'failed to create tool page', got: %v", actErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateToolComponent_AdoptStillCleansUpAPageItCreated(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	createdPageID := uuid.New()

	expectPreamble(mock)
	mock.ExpectQuery("url_shape").
		WillReturnRows(sqlmock.NewRows([]string{"shape"}))
	// No existing row under either name shape: canonical identity is minted.
	mock.ExpectQuery("SELECT name, url FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url"}))
	// UpsertPageForRole creates the page.
	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(createdPageID.String()))
	// Link failure: cleanup must delete BOTH the page this call created and
	// the component — the created/adopted asymmetry's other half.
	mock.ExpectExec("INSERT INTO page_components").
		WillReturnError(sql.ErrConnDone)
	mock.ExpectExec("DELETE FROM pages").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM content_components").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, actErr := CreateToolComponentAction(context.Background(), adoptTestParams(db, true))
	if actErr == nil || !strings.Contains(actErr.Error(), "failed to link component to page") {
		t.Fatalf("want the induced link error, got: %v", actErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// errDuplicatePagesKey mimics the pq unique-violation surface enough for the
// error path under test (the action only wraps and returns it).
type errDuplicatePagesKey struct{}

func (errDuplicatePagesKey) Error() string {
	return `duplicate key value violates unique constraint "pages_site_id_name_key"`
}
