// FILE: platform/orchestration/actions/site_component_divergence_test.go
//
// Pins for the bugs_open/226 divergence mechanism. The mechanism's promise
// rests on three properties no compiler enforces:
//
//   1. the classification branches (stamped-and-matching is the ONLY
//      machine_made verdict; everything unstamped stays advisory);
//   2. the artefact digest is stamped in the SAME statement that stores the
//      bytes (a stamp written anywhere else can describe different bytes);
//   3. the Go-side vocabulary and the DB trigger's CASE (sql_for_agents/344)
//      agree — they are two hand-maintained copies of one judgement, which is
//      exactly the drift class chrome_render_inputs_contract_test.go beside
//      this file exists to catch for the 117 stamp.

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

func classifyWithRows(t *testing.T, stamped interface{}, current string, byteCount int) *siteComponentArtefactVerdict {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"rendered_html_digest", "md5", "length"}).
		AddRow(stamped, current, byteCount)
	mock.ExpectQuery("SELECT rendered_html_digest, md5\\(rendered_html\\), length\\(rendered_html\\)").
		WillReturnRows(rows)

	v, cErr := classifySiteComponentArtefact(context.Background(), db, uuid.New(), "footer")
	if cErr != nil {
		t.Fatalf("classify: %v", cErr)
	}
	return v
}

func TestClassifySiteComponentArtefactBranches(t *testing.T) {
	const d = "0123456789abcdef0123456789abcdef"
	const other = "ffffffffffffffffffffffffffffffff"

	if v := classifyWithRows(t, d, d, 4096); v.State != artefactMachineMade {
		t.Errorf("stamped+matching = %q, want machine_made", v.State)
	}
	if v := classifyWithRows(t, d, other, 4096); v.State != artefactHandPatched {
		t.Errorf("stamped+mismatched = %q, want hand_patched", v.State)
	} else if v.StampedDigest != d || v.CurrentDigest != other {
		t.Errorf("hand_patched verdict must carry both digests for the work item; got %q / %q",
			v.StampedDigest, v.CurrentDigest)
	}
	if v := classifyWithRows(t, nil, other, 4096); v.State != artefactUnstamped {
		t.Errorf("NULL stamp = %q, want unstamped — pre-344 rows must stay advisory, never accused", v.State)
	}
	if v := classifyWithRows(t, "", other, 4096); v.State != artefactUnstamped {
		t.Errorf("empty-string stamp = %q, want unstamped", v.State)
	}
}

func TestClassifySiteComponentArtefactEmptySlot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT rendered_html_digest").WillReturnError(sql.ErrNoRows)

	v, cErr := classifySiteComponentArtefact(context.Background(), db, uuid.New(), "head")
	if cErr != nil {
		t.Fatalf("ErrNoRows must classify, not error: %v", cErr)
	}
	if v.State != artefactEmpty {
		t.Errorf("no row = %q, want empty", v.State)
	}
}

// The digest must be stamped in the SAME statement that stores the bytes —
// md5($1) beside rendered_html = $1, inside the one guarded UPDATE. Anchored
// at the store statement (not grepped bare), same pattern as the fingerprint
// contract tests: a comment or another statement cannot satisfy it.
func TestChromeStoreStampsDigestInSameStatement(t *testing.T) {
	src, err := os.ReadFile("render_site_components_action.go")
	if err != nil {
		t.Fatalf("cannot read render_site_components_action.go: %v", err)
	}

	stmt := regexp.MustCompile(`(?s)UPDATE site_components AS sc\s*\n\s*SET rendered_html = \$1, build_status = 'rendered'.*?WHERE sc\.site_id`)
	m := stmt.Find(src)
	if m == nil {
		t.Fatal("cannot locate the guarded chrome store statement — if it moved, move this pin with it")
	}
	if !strings.Contains(string(m), "rendered_html_digest = md5($1)") {
		t.Error("the chrome store statement no longer stamps rendered_html_digest = md5($1) beside the bytes — " +
			"the bugs_open/226 hand-patch detector is disarmed (a stamp in any OTHER statement can describe different bytes)")
	}
}

// The Go vocabulary and the 344 trigger's CASE are two copies of one
// judgement. This pin fails when either side renames a verdict or the trigger
// stops judging by md5 — the drift would otherwise surface as work items that
// disagree with the ledger they cite.
func TestDivergenceVocabularyMatchesMigration344(t *testing.T) {
	mig, err := os.ReadFile("../../../docs/agent_docs/sql_for_agents/344_site_component_history_divergence_guard.sql")
	if err != nil {
		t.Fatalf("cannot read migration 344 (moved?): %v", err)
	}
	s := string(mig)

	for _, verdict := range []siteComponentArtefactState{artefactUnstamped, artefactMachineMade, artefactHandPatched} {
		if !strings.Contains(s, "'"+string(verdict)+"'") {
			t.Errorf("migration 344 does not carry verdict %q — Go and trigger vocabularies have drifted", verdict)
		}
	}
	if !strings.Contains(s, "OLD.rendered_html_digest = md5(OLD.rendered_html)") {
		t.Error("migration 344's trigger no longer judges by md5 of the OLD bytes — its verdicts and classifySiteComponentArtefact's are no longer the same judgement")
	}
}
