// FILE: platform/orchestration/actions/tool_writer_link_repair_362_test.go
//
// bugs_open/362. The two tool writers (create_tool_component_action.go,
// deploy_tool_action.go) persisted rendered_html without the dead-link repair
// their three siblings call. The wiring was deferred on 2026-08-02 because the
// repair then destroyed JavaScript that BUILDS anchors (bugs_open/180: against
// '<a href="' + x + '">' the href capture reads empty and the "repair" deletes
// a working link from a running program), unblocked the same day by LNK-029's
// span-aware repair, and left unwired for twenty days by two documents that
// went on saying "wait".
//
// The first test is the landmine's own prescribed probe, run DESPITE LNK-029:
// the fix is the reason wiring is allowed, not a reason to trust it unprobed.
// Tool markup is exactly where JS-built anchors live, so the probe uses the
// REAL seam over the 180 shape and demands byte-identical script bytes — while
// ALSO demanding a genuinely dead markup anchor beside it IS repaired, because
// a byte-identical-only probe cannot tell "the repair spared the script" from
// "the repair did nothing at all".
//
// The second test holds both call sites to the wiring; it is the mutation
// target (delete either call in a scratch copy and it must fail naming the
// file).
package actions

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The 180 shape, verbatim in spirit: a script that assembles an anchor from
// pieces, inside otherwise ordinary tool markup.
const jsBuiltAnchorScript = `<script>
var links = items.map(function (q) {
  return '<a href="' + q.link + '" class="result-cta-primary">' + q.label + '</a>';
});
resultBox.innerHTML = links.join('');
</script>`

// TestToolWriterRepairSparesJSBuiltAnchorsAndStillRepairsMarkup is two-armed on
// one input: the script span must come back byte-identical AND the dead markup
// anchor beside it must be unlinked (text kept). A clean-input-only probe would
// pass identically on a seam that skips everything.
func TestToolWriterRepairSparesJSBuiltAnchorsAndStillRepairsMarkup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(linkablePageStatusPredicate)).
		WillReturnRows(pageIndexRows("/index.html", "/tools/real-tool/index.html"))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	in := `<div class="tool-container"><p>` +
		`<a href="/tools/real-tool/index.html">a real page</a> and ` +
		`<a href="/tools/does-not-exist.html">a phantom page</a></p>` +
		jsBuiltAnchorScript + `</div>`

	got := repairComponentHTMLBeforePersist(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, uuid.New(),
		"webdesign.co.uk", "tool-probe", "/tools/probe/index.html",
		"create_tool_component", in, zap.NewNop())

	if !strings.Contains(got, jsBuiltAnchorScript) {
		t.Errorf("the script span was NOT byte-identical — the repair touched JavaScript that builds anchors, which is bugs_open/180 reborn:\n%q", got)
	}
	if strings.Contains(got, `href="/tools/does-not-exist.html"`) {
		t.Errorf("the phantom markup link survived — the repair is not actually reaching markup: %q", got)
	}
	if !strings.Contains(got, "a phantom page") {
		t.Errorf("unlinking dropped the anchor text, which is content: %q", got)
	}
	if !strings.Contains(got, `href="/tools/real-tool/index.html"`) {
		t.Errorf("a valid markup link was disturbed: %q", got)
	}
}

// TestToolWriterRepairIsFullyInertOnCleanToolBytes is the zero arm: JS-built
// anchors plus only-valid markup must come back byte-identical as a WHOLE, and
// silently (the negative asserted the non-vacuous way the clean-component test
// in component_link_repair_test.go documents: register the forbidden INSERT and
// require ExpectationsWereMet to fail naming it).
func TestToolWriterRepairIsFullyInertOnCleanToolBytes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(linkablePageStatusPredicate)).
		WillReturnRows(pageIndexRows("/index.html"))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1)) // must stay UNMATCHED

	in := `<div><a href="/index.html">home</a>` + jsBuiltAnchorScript + `</div>`
	got := repairComponentHTMLBeforePersist(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, uuid.New(),
		"webdesign.co.uk", "tool-probe", "/tools/probe/index.html",
		"deploy_tool_to_site", in, zap.NewNop())

	if got != in {
		t.Errorf("clean tool bytes were perturbed:\n got %q\nwant %q", got, in)
	}
	if err := mock.ExpectationsWereMet(); err == nil {
		t.Fatal("a clean component wrote an agent_error_log row — a no-op repair is not an event")
	}
}

// TestToolWritersCallTheRepairBeforeTheirPersist scans both writer files and
// requires the repair call to appear BEFORE the page_components INSERT.
// Comment lines are skipped; finding either marker zero times is a loud
// failure (a broken scan must not pass silently — the 336/339 test pattern).
func TestToolWritersCallTheRepairBeforeTheirPersist(t *testing.T) {
	for _, file := range []string{
		"create_tool_component_action.go",
		"deploy_tool_action.go",
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("cannot read %s: %v", file, err)
		}
		repairLine, insertLine := -1, -1
		for i, l := range strings.Split(string(src), "\n") {
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if repairLine == -1 && strings.Contains(l, "repairComponentHTMLBeforePersist(") {
				repairLine = i + 1
			}
			if insertLine == -1 && strings.Contains(l, "INSERT INTO page_components") {
				insertLine = i + 1
			}
		}
		if repairLine == -1 {
			t.Errorf("%s: no call to repairComponentHTMLBeforePersist — the bugs_open/362 wiring is gone; this writer persists rendered_html unrepaired again", file)
			continue
		}
		if insertLine == -1 {
			t.Errorf("%s: no page_components INSERT found — the scan is anchored on nothing and this test is asserting nothing", file)
			continue
		}
		if repairLine > insertLine {
			t.Errorf("%s: the repair call (line %d) comes AFTER the persist (line %d) — what is persisted is not what was repaired", file, repairLine, insertLine)
		}
	}
}
