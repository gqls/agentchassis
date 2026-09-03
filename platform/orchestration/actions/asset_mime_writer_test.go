package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The framework half of bugs_open/433.
//
// The bug was not one writer forgetting a column; it was that FOUR writers each
// decided independently whether to record what they had stored, and three chose
// not to. 1,023 of 1,418 rows carried no mime_type at all [MEASURED 2026-09-03],
// and the writers were only identifiable by censusing the table and reasoning
// backwards from asset_key shapes. This test makes the next writer declare
// itself instead.
//
// THE RULE: a statement that writes assets.url must also write assets.mime_type.
// Binding them is what makes the row self-consistent — url says where the
// artefact is, mime_type says what it is, and the writer that publishes the
// bytes is the only code in a position to know both. Writing NULL is an
// acceptable answer (NULLIF(...,'') is how the writers here spell "I could not
// identify these bytes"); saying nothing at all is not, because an absent column
// is indistinguishable from a writer that never considered the question.
//
// WHAT IT DELIBERATELY DOES NOT ASSERT: that the value is non-NULL. A source
// scanner cannot judge that, and a NULL written by a writer that sniffed and
// failed is exactly right. Value honesty is a data question, not a source one.

var assetStatementRe = regexp.MustCompile(`(?s)(INSERT INTO assets\s*\((?:[^)]*)\)|UPDATE assets\b.*?(?:WHERE|RETURNING|` + "`" + `))`)

// assetWriterExemptions maps a file to the reason its assets write legitimately
// omits mime_type. EMPTY BY DESIGN — after bugs_open/433 every writer complies,
// so the ban is available rather than a baseline of tolerated violations. An
// entry here is a decision someone made and must say why.
var assetWriterExemptions = map[string]string{}

func TestEveryAssetsWriterThatSetsURLAlsoSetsMimeType(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	checked := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, stmt := range assetWriteStatements(string(src)) {
			if !mentionsColumn(stmt, "url") {
				continue
			}
			checked++
			if mentionsColumn(stmt, "mime_type") {
				continue
			}
			if reason, ok := assetWriterExemptions[f]; ok {
				t.Logf("%s: exempt — %s", f, reason)
				continue
			}
			t.Errorf("%s writes assets.url without assets.mime_type.\n"+
				"A row whose url says where the artefact is and says nothing about what it is "+
				"cannot be trusted by any reader (bugs_open/433: 1,023 of 1,418 rows were empty "+
				"because three writers each independently chose not to record it).\n"+
				"Bind it from the BYTES you are publishing — storage.SniffImageExtAndMIME — and "+
				"write NULLIF(..., '') so an unidentifiable image records NULL rather than a "+
				"confident guess. If omitting it is genuinely right here, add an entry to "+
				"assetWriterExemptions saying why.\nStatement: %s", f, squash(stmt))
		}
	}
	// The empty-set trap: a classifier that matches nothing passes silently.
	if checked == 0 {
		t.Fatal("found no assets writes at all — the statement matcher has stopped biting, so this test is asserting nothing")
	}
}

// The mutation guard. Without it, gutting assetStatementRe makes the test above
// pass on a tree full of violations, and a green ratchet reads as a clean tree.
// Precedent: TestUnguardedRuleFiresOnASyntheticProducer in asset_lock_guard_test.go.
func TestAssetsWriterMimeRuleFiresOnASyntheticWriter(t *testing.T) {
	bad := "_, _ = db.Exec(`INSERT INTO assets (id, site_id, url, origin_type) VALUES ($1,$2,$3,'generated')`)"
	stmts := assetWriteStatements(bad)
	if len(stmts) != 1 {
		t.Fatalf("the matcher must find the synthetic writer, found %d", len(stmts))
	}
	if !mentionsColumn(stmts[0], "url") {
		t.Error("the synthetic writer sets url and the rule cannot see it")
	}
	if mentionsColumn(stmts[0], "mime_type") {
		t.Error("the synthetic writer sets no mime_type, yet the rule believes it does")
	}

	good := "_, _ = db.Exec(`INSERT INTO assets (id, url, mime_type) VALUES ($1,$2,NULLIF($3,''))`)"
	gs := assetWriteStatements(good)
	if len(gs) != 1 || !mentionsColumn(gs[0], "mime_type") {
		t.Error("a compliant writer must be acquitted, or the rule is unsatisfiable")
	}

	neither := "_, _ = db.Exec(`UPDATE assets SET status = 'archived' WHERE id = $1`)"
	for _, s := range assetWriteStatements(neither) {
		if mentionsColumn(s, "url") {
			t.Error("a statement touching neither column must not be flagged")
		}
	}
}

// ⚠ Reads the STATEMENT, never a window around it. A line-window scan is the
// obvious implementation and it is wrong in both directions here: this package
// logs zap.String("mime_type", ...) and zap.String("component_id", ...) near
// several of these statements, so "is the column mentioned nearby?" returns true
// while the column is absent. That exact mistake cleared the one page_components
// writer that was actually violating its rule (see the 457 lane's WRONG_CALLS
// entry of 2026-09-03).
func assetWriteStatements(src string) []string {
	var out []string
	for _, m := range assetStatementRe.FindAllString(src, -1) {
		out = append(out, m)
	}
	return out
}

func mentionsColumn(stmt, col string) bool {
	// Word-boundary match so `url` does not match `origin_url`, and
	// `mime_type` is not satisfied by a comment mentioning it.
	stripped := regexp.MustCompile(`(?m)--.*$`).ReplaceAllString(stmt, "")
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(col) + `\b`).MatchString(stripped)
}

func squash(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
