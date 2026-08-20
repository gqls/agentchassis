// FILE: platform/orchestration/actions/create_tool_component_regenerate_test.go
//
// bugs_open/331: create_tool_component had a CREATE path and no REPLACE path.
// The per-item `replace_existing` input turns the already_exists short-circuit
// into an in-place regeneration of the incumbent (CTS-009's convention).
//
// Arms pinned here (sqlmock, ordered walks, same scaffolding as the adopt and
// fork tests — adoptTestToolHTML / adoptTestParams):
//
//	A. flag ABSENT + incumbent ⇒ already_exists, the walk ENDS at the probe
//	   (zero writes). Pins the byte-identical default; the probe stays the
//	   per-site throttle.
//	B. flag TRUE (JSON bool, and the quoted "true" a hand-written spec may
//	   carry) + incumbent ⇒ read incumbent → version snapshot → one tx:
//	   placements locked FOR UPDATE → content_components UPDATE pinned to the
//	   new html and the INCUMBENT id → page_components UPDATE pinned to the new
//	   html → COMMIT. Result: regenerated:true, component_id == incumbent, the
//	   placement's page id/url, no already_exists.
//	   Mutation run (2026-08-19): delete the page_components UPDATE in the arm
//	   → ExpectationsWereMet fails ("expectation not met: UPDATE page_components")
//	   → restored.
//	C. flag TRUE + incumbent with NO live writable placement ⇒ ROLLBACK and a
//	   typed refusal; the content_components UPDATE is never reached (the
//	   assertion on the message is what proves it: an unexpected UPDATE would
//	   surface as "update component", not "refused").
//	E. flag TRUE + incumbent placed on >1 site ⇒ the shared fence (285) refuses
//	   BEFORE the transaction; no Begin, no UPDATE.
//	F. flag TRUE + a HOLLOW incoming template (valid doc header, balanced tags,
//	   ZERO visible text — the ab-test hollow-shell shape, and byte-for-byte
//	   what adoptTestToolHTML used to be) ⇒ refused BEFORE the snapshot; the
//	   only write is the agent_error_log rejection row. Mutation run
//	   (2026-08-20): gate deleted ⇒ the walk reaches MAX(version_number)
//	   unexpectedly ⇒ restored.
//	G. flag TRUE + an incumbent with substantial visible text (>=200 chars) and
//	   an incoming template keeping <50% of it ⇒ the RELATIVE half refuses the
//	   same way (evaluateSectionShrink, shared axis, shared config key).
//	D. flag TRUE + NO incumbent ⇒ exactly today's greenfield walk
//	   (expectPreamble + the bare page INSERT); the flag does nothing when
//	   there is nothing to replace.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

const regenIncumbentID = "0c9e6e2c-9a4e-4e85-9281-78f470a99a91"

func TestCreateToolComponent_ReplaceAbsentKeepsTheAlreadyExistsShortCircuit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(regenIncumbentID))
	// Nothing else: the walk ends here.

	p := adoptTestParams(db, nil)
	res, err := CreateToolComponentAction(context.Background(), p)
	if err != nil {
		t.Fatalf("expected the already_exists short-circuit, got error: %v", err)
	}
	m := res.(map[string]interface{})
	if m["already_exists"] != true || m["component_id"] != regenIncumbentID {
		t.Fatalf("expected already_exists for %s, got %v", regenIncumbentID, m)
	}
	if _, has := m["regenerated"]; has {
		t.Fatalf("flag absent must not regenerate: %v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func expectRegenerateWalk(mock sqlmock.Sqlmock, pageID string) {
	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(regenIncumbentID))
	// incumbent read
	mock.ExpectQuery(`COALESCE\(html_template, ''\), COALESCE\(input_schema::text, '\{\}'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"html", "schema"}).AddRow("<div>old</div>", "{}"))
	// version snapshot (best-effort)
	mock.ExpectQuery(`MAX\(version_number\)`).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))
	mock.ExpectExec("INSERT INTO component_versions").
		WithArgs(regenIncumbentID, 3, "<div>old</div>", "{}", sqlmock.AnyArg(), "tool-generator:replace_existing", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// the shared-template fence (bugs_open/285): one page, one site
	mock.ExpectQuery(`count\(DISTINCT pc.page_id\), count\(DISTINCT p.site_id\)`).
		WithArgs(regenIncumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"pages", "sites"}).AddRow(1, 1))
	// the transaction
	mock.ExpectBegin()
	mock.ExpectQuery("FOR UPDATE OF pc").
		WithArgs(regenIncumbentID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_id", "name", "url", "had_html"}).
			AddRow("3a56d6dc-0000-0000-0000-000000000001", pageID, "tool-aspect-ratio", "/tools/aspect-ratio/index.html", true))
	mock.ExpectExec("UPDATE content_components").
		WithArgs(adoptTestToolHTML, "Aspect Ratio Calculator", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), regenIncumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE page_components").
		WithArgs(adoptTestToolHTML, `{"3a56d6dc-0000-0000-0000-000000000001"}`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func TestCreateToolComponent_ReplaceRegeneratesTheIncumbentInPlace(t *testing.T) {
	for _, flag := range []interface{}{true, "true"} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		pageID := uuid.New().String()
		expectRegenerateWalk(mock, pageID)

		p := adoptTestParams(db, nil)
		p.CollectedData["replace_existing"] = flag
		res, err := CreateToolComponentAction(context.Background(), p)
		if err != nil {
			t.Fatalf("flag=%v: expected in-place regeneration, got error: %v", flag, err)
		}
		m := res.(map[string]interface{})
		if m["regenerated"] != true || m["component_id"] != regenIncumbentID || m["page_id"] != pageID {
			t.Fatalf("flag=%v: unexpected result map: %v", flag, m)
		}
		if _, has := m["already_exists"]; has {
			t.Fatalf("flag=%v: a regeneration must not report already_exists: %v", flag, m)
		}
		if m["page_url"] != "/tools/aspect-ratio/index.html" || m["function"] != "tool-aspect-ratio" {
			t.Fatalf("flag=%v: downstream keys (page_url/function) wrong: %v", flag, m)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("flag=%v: walk: %v", flag, err)
		}
		db.Close()
	}
}

func TestCreateToolComponent_ReplaceRefusesWhenNoLivePlacement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(regenIncumbentID))
	mock.ExpectQuery(`COALESCE\(html_template, ''\)`).
		WillReturnRows(sqlmock.NewRows([]string{"html", "schema"}).AddRow("<div>old</div>", "{}"))
	mock.ExpectQuery(`MAX\(version_number\)`).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))
	mock.ExpectExec("INSERT INTO component_versions").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`count\(DISTINCT pc.page_id\)`).
		WillReturnRows(sqlmock.NewRows([]string{"pages", "sites"}).AddRow(1, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("FOR UPDATE OF pc").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_id", "name", "url", "had_html"})) // no live slot
	mock.ExpectRollback()

	p := adoptTestParams(db, nil)
	p.CollectedData["replace_existing"] = true
	_, err = CreateToolComponentAction(context.Background(), p)
	if err == nil || !strings.Contains(err.Error(), "replace_existing refused") {
		t.Fatalf("expected the typed refusal, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func TestCreateToolComponent_ReplaceWithNoIncumbentIsTheGreenfieldPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectPreamble(mock) // probe EMPTY → creation walk, exactly as the adopt tests pin it
	mock.ExpectExec("INSERT INTO pages").
		WillReturnError(errDuplicatePagesKey{})
	mock.ExpectExec("DELETE FROM content_components").
		WillReturnResult(sqlmock.NewResult(0, 1))

	p := adoptTestParams(db, nil)
	p.CollectedData["replace_existing"] = true
	_, err = CreateToolComponentAction(context.Background(), p)
	if err == nil || !strings.Contains(err.Error(), "failed to create tool page") {
		t.Fatalf("expected today's bare-insert path (stopped at the induced collision), got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("walk: %v", err)
	}
}

// E. The incumbent is placed on MORE THAN ONE site: a site-scoped regeneration
// must not rewrite a row other sites serve. Refused by the shared fence BEFORE
// the transaction — no Begin, no UPDATE.
func TestCreateToolComponent_ReplaceRefusesAnIncumbentServedByAnotherSite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(regenIncumbentID))
	mock.ExpectQuery(`COALESCE\(html_template, ''\)`).
		WillReturnRows(sqlmock.NewRows([]string{"html", "schema"}).AddRow("<div>old</div>", "{}"))
	mock.ExpectQuery(`MAX\(version_number\)`).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))
	mock.ExpectExec("INSERT INTO component_versions").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`count\(DISTINCT pc.page_id\)`).
		WillReturnRows(sqlmock.NewRows([]string{"pages", "sites"}).AddRow(2, 2)) // shared row

	p := adoptTestParams(db, nil)
	p.CollectedData["replace_existing"] = true
	_, err = CreateToolComponentAction(context.Background(), p)
	if err == nil || !strings.Contains(err.Error(), "across 2 sites") {
		t.Fatalf("expected the fence refusal, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("walk: %v", err)
	}
}

// hollowToolHTML passes both birth gates and has ZERO visible text — the
// ab-test hollow-shell shape (bugs_closed/286 "related finding": 13,284 chars
// of markup, zero visible chars, served blank). This is what adoptTestToolHTML
// was before the non-hollow gate existed; it is preserved here as the fixture
// that must FAIL arm F for as long as the gate lives.
const hollowToolHTML = `<div class="tool-container"><script>
/* === tool-doc ===
name: Aspect Ratio Calculator
=== /tool-doc === */
function calc() { return 1; }
</script></div>`

func TestCreateToolComponent_ReplaceRefusesAHollowRegeneration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(regenIncumbentID))
	mock.ExpectQuery(`COALESCE\(html_template, ''\)`).
		WillReturnRows(sqlmock.NewRows([]string{"html", "schema"}).AddRow("<div><p>A working tool with real visible words in it.</p></div>", "{}"))
	// The ONLY write after the refusal decision: the rejection row. No
	// MAX(version_number), no snapshot, no Begin — the walk ends here, which is
	// what "refused BEFORE the snapshot" means in mock terms.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(0, 1))

	p := adoptTestParams(db, nil)
	p.CollectedData["replace_existing"] = true
	p.CollectedData["html_content"] = hollowToolHTML
	_, err = CreateToolComponentAction(context.Background(), p)
	if err == nil || !strings.Contains(err.Error(), "ZERO visible text") {
		t.Fatalf("expected the hollow refusal, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func TestCreateToolComponent_ReplaceRefusesAShrinkPastTheFloor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// An incumbent above minShrinkGuardVisibleChars (200) so the relative half
	// is armed; the incoming fixture (adoptTestToolHTML) keeps ~30 visible
	// chars, far under the default 50% floor.
	incumbent := "<div><p>" + strings.Repeat("real prose ", 30) + "</p></div>" // ~330 visible chars

	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("webdesign.co.uk"))
	mock.ExpectQuery("FROM content_components cc").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(regenIncumbentID))
	mock.ExpectQuery(`COALESCE\(html_template, ''\)`).
		WillReturnRows(sqlmock.NewRows([]string{"html", "schema"}).AddRow(incumbent, "{}"))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(0, 1))

	p := adoptTestParams(db, nil)
	p.CollectedData["replace_existing"] = true
	_, err = CreateToolComponentAction(context.Background(), p)
	if err == nil || !strings.Contains(err.Error(), "visible text would fall") {
		t.Fatalf("expected the shrink refusal, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("walk: %v", err)
	}
}
