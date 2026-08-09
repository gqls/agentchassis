// Tests for the bugs_open/229 page-side divergence guard. Pins, in order:
//   1. classify returns only what the destructive statements will actually
//      remove, with both digests carried for the work item;
//   2. every render/save writer stamps rendered_html_digest = md5(html) in
//      the SAME statement as the bytes (anchored at each statement — a
//      comment or a neighbouring statement cannot satisfy the pin);
//   3. adopt_verbatim does NOT stamp — ported bytes are not reproducible
//      from content_data, and a stamp there would disarm the detector;
//   4. the Go vocabulary matches migration 357's trigger CASE exactly.

package actions

import (
	"context"
	"database/sql"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestClassifyPageComponentArtefactsReturnsMismatchedRows(t *testing.T) {
	const stamped = "0123456789abcdef0123456789abcdef"
	const current = "ffffffffffffffffffffffffffffffff"

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	compID, siteID := uuid.New(), uuid.New()
	rows := sqlmock.NewRows([]string{"id", "site_id", "slot", "position", "stamped", "current", "length", "owned"}).
		AddRow(compID, siteID, "hero", 2, stamped, current, 4096, true)
	mock.ExpectQuery("SELECT pc.id, p.site_id").WillReturnRows(rows)

	got, cErr := classifyPageComponentArtefacts(context.Background(), db, uuid.New())
	if cErr != nil {
		t.Fatalf("classify: %v", cErr)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 divergent row, got %d", len(got))
	}
	d := got[0]
	if d.ComponentID != compID || d.SiteID != siteID || d.Position != 2 || !d.OwnedPage {
		t.Errorf("row identity lost in scan: %+v", d)
	}
	if d.StampedDigest != stamped || d.CurrentDigest != current {
		t.Errorf("verdict must carry both digests for the work item; got %q / %q",
			d.StampedDigest, d.CurrentDigest)
	}
}

// The classify SELECT must use the SAME agent-writable predicate as the
// destructive statements: a locked row survives the rebuild, and counting it
// as destroyed is a false accusation. Anchored at the query in the classify
// function itself.
func TestClassifyPageComponentsUsesWritablePredicate(t *testing.T) {
	src, err := os.ReadFile("page_component_divergence.go")
	if err != nil {
		t.Fatalf("cannot read page_component_divergence.go: %v", err)
	}
	stmt := regexp.MustCompile(`(?s)SELECT pc\.id, p\.site_id.*?FROM page_components pc.*?pageComponentAgentWritableSQL\("pc\."\)`)
	if stmt.Find(src) == nil {
		t.Error("classifyPageComponentArtefacts no longer filters on pageComponentAgentWritableSQL — " +
			"it will count lock-surviving rows as destroyed (false alarms) or drift from the DELETE's scope")
	}
}

func TestClassifyPageComponentArtefactsEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT pc.id, p.site_id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "site_id", "slot", "position", "stamped", "current", "length", "owned"}))

	got, cErr := classifyPageComponentArtefacts(context.Background(), db, uuid.New())
	if cErr != nil {
		t.Fatalf("classify on empty page: %v", cErr)
	}
	if len(got) != 0 {
		t.Errorf("empty page must classify to an empty slice, got %d rows", len(got))
	}
}

func TestReadBackPageDivergenceFromLedger(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	rows := sqlmock.NewRows([]string{"component_id", "site_id", "slot", "position", "stamped", "current", "length", "owned"}).
		AddRow(uuid.Nil, siteID, "hero", 3, "aaaa", "bbbb", 1024, false)
	mock.ExpectQuery("SELECT COALESCE\\(component_id").WillReturnRows(rows)

	got, rErr := readBackPageDivergenceFromLedger(context.Background(), db, uuid.New())
	if rErr != nil {
		t.Fatalf("read-back: %v", rErr)
	}
	if len(got) != 1 || got[0].SiteID != siteID || got[0].Position != 3 {
		t.Fatalf("ledger read-back lost the row: %+v", got)
	}
}

func TestPageDivergenceItemKeyShape(t *testing.T) {
	pageID := uuid.MustParse("12345678-9abc-def0-1234-56789abcdef0")
	key := pageDivergenceItemKey(pageID, 4, "0123456789abcdefdeadbeef")
	want := "page_divergence_overwritten:page_component:12345678:4:0123456789ab"
	if key != want {
		t.Errorf("item key = %q, want %q — the digest fragment is the per-patch dedup axis (chrome round 1)", key, want)
	}
}

// Same-statement stamp pins, one per stamped writer. Each is anchored at its
// own destructive/insert statement so a stamp anywhere else cannot satisfy it.
func TestPageWritersStampDigestInSameStatement(t *testing.T) {
	cases := []struct {
		file   string
		anchor string
		stamp  string
		reason string
	}{
		{
			file:   "save_page_sections_action.go",
			anchor: `(?s)INSERT INTO page_components \(page_id, position, rendered_html, rendered_html_digest,.*?'deployed'\)`,
			stamp:  "md5($3)",
			reason: "the full-save INSERT is the page render path",
		},
		{
			file:   "rebuild_blog_listing_action.go",
			anchor: `(?s)UPDATE page_components\s*\n\s*SET rendered_html = \$1, rendered_html_digest = md5\(\$1\), content_data`,
			stamp:  "md5($1)",
			reason: "the listing refresh UPDATE arm",
		},
		{
			file:   "rebuild_blog_listing_action.go",
			anchor: `(?s)INSERT INTO page_components \(page_id, slot_name, position, rendered_html, rendered_html_digest.*?VALUES \(\$1, \$2, 3, \$3, md5\(\$3\)`,
			stamp:  "md5($3)",
			reason: "the listing INSERT arm",
		},
		{
			file:   "section_editor_actions.go",
			anchor: `(?s)SET rendered_html = \$2,\s*\n\s*rendered_html_digest = md5\(\$2\),\s*\n\s*content_data`,
			stamp:  "md5($2)",
			reason: "the section editor renders the edited content",
		},
		{
			file:   "create_report_page_action.go",
			anchor: `(?s)SET rendered_html = \$1, rendered_html_digest = md5\(\$1\), content_data`,
			stamp:  "md5($1)",
			reason: "the dossier render UPDATE arm",
		},
		{
			file:   "create_report_page_action.go",
			anchor: `(?s)rendered_html, rendered_html_digest, content_data, build_status\)\s*\n\s*VALUES \(\$1, \$2, 0, \$3, \$4, md5\(\$4\)`,
			stamp:  "md5($4)",
			reason: "the dossier render INSERT arm",
		},
	}
	for _, c := range cases {
		src, err := os.ReadFile(c.file)
		if err != nil {
			t.Fatalf("cannot read %s: %v", c.file, err)
		}
		m := regexp.MustCompile(c.anchor).Find(src)
		if m == nil {
			t.Errorf("%s: cannot locate the stamped statement (%s) — if it moved, move this pin with it",
				c.file, c.reason)
			continue
		}
		if !strings.Contains(string(m), c.stamp) {
			t.Errorf("%s: statement no longer stamps %s beside the bytes (%s) — the bugs_open/229 detector is disarmed there",
				c.file, c.stamp, c.reason)
		}
	}
}

// adopt_verbatim must NOT stamp: ported bytes are not reproducible from
// content_data. A stamp beside them would declare recoverable what is not —
// the declaring-a-key-silences-your-own-detector class.
func TestAdoptVerbatimDoesNotStamp(t *testing.T) {
	src, err := os.ReadFile("adopt_verbatim.go")
	if err != nil {
		t.Fatalf("cannot read adopt_verbatim.go: %v", err)
	}
	if strings.Contains(string(src), "rendered_html_digest") {
		t.Error("adopt_verbatim writes rendered_html_digest — ported bytes are NOT reproducible from " +
			"content_data, so stamping them silences the bugs_open/229 detector for adopted pages")
	}
}

// The Go side's md5 comparison vocabulary must match migration 357's trigger
// CASE — two halves of one judgement (the 016b both-halves landmine class).
func TestDivergenceVocabularyMatchesMigration357(t *testing.T) {
	src, err := os.ReadFile("../../../docs/agent_docs/sql_for_agents/357_page_component_artefact_archive.sql")
	if err != nil {
		t.Skipf("migration file not reachable from test dir: %v", err)
	}
	s := string(src)
	for _, needle := range []string{
		"WHEN OLD.rendered_html_digest IS NULL THEN 'unstamped'",
		"WHEN OLD.rendered_html_digest = md5(OLD.rendered_html) THEN 'machine_made'",
		"ELSE 'hand_patched'",
		"'artefact_archive_trigger'",
	} {
		if !strings.Contains(s, needle) {
			t.Errorf("migration 357 no longer contains %q — the DB-side verdict has drifted from the Go side's", needle)
		}
	}
	goSrc, err := os.ReadFile("page_component_divergence.go")
	if err != nil {
		t.Fatalf("cannot read page_component_divergence.go: %v", err)
	}
	if !strings.Contains(string(goSrc), "rendered_html_digest <> md5(pc.rendered_html)") {
		t.Error("classify no longer compares stamp to md5(rendered_html) — drifted from the trigger's judgement")
	}
}

func TestClassifyPageComponentsPropagatesError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT pc.id, p.site_id").WillReturnError(sql.ErrConnDone)

	if _, cErr := classifyPageComponentArtefacts(context.Background(), db, uuid.New()); cErr == nil {
		t.Error("classify must surface query errors — callers use the error to switch to the ledger read-back")
	}
}
