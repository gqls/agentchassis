// FILE: platform/orchestration/actions/discovery_checks/check_palette_contrast_test.go
//
// The fixtures are live palettes, taken from `palettes.colours` on 2026-07-27,
// so these tests assert against defects that exist on production sites rather
// than against invented ones. The pair-level arithmetic is tested in
// platform/colour; what is tested here is the part this file owns — reading the
// right row, skipping the right values, and grading severity.

package discovery_checks

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// fundamentallyai.com's palette as it SHIPPED on the morning of 2026-07-27 —
// the state the owner found on his phone.
const brokenColoursJSON = `{"text":"#E4EAF2","accent":"#C8902A","border":"#1B2D47",
 "cta_bg":"#1a365d","card_bg":"#ffffff","primary":"#0E1B2E","surface":"#111E33",
 "cta_text":"#ffffff","secondary":"#1A2E48","background":"#080E1C",
 "text_muted":"#7E91A8","primary_text":"#ffffff",
 "footer_text":"rgba(232,237,243,0.88)"}`

// The same palette after the stylesheet was regenerated.
const repairedColoursJSON = `{"text":"#E4EAF2","accent":"#C8902A","border":"#1B2D47",
 "cta_bg":"#101E33","card_bg":"#132239","primary":"#86ADDE","surface":"#111E33",
 "cta_text":"#E8EDF3","secondary":"#4A6C99","background":"#080E1C",
 "text_muted":"#7E91A8","primary_text":"#071019",
 "footer_text":"rgba(232,237,243,0.88)"}`

func ctxWithPalette(t *testing.T, coloursJSON string, rows bool) (DiscoveryCheckContext, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	q := mock.ExpectQuery("SELECT p.colours").WithArgs(sqlmock.AnyArg())
	if rows {
		q.WillReturnRows(sqlmock.NewRows([]string{"colours", "name"}).
			AddRow([]byte(coloursJSON), "palette-under-test"))
	} else {
		q.WillReturnRows(sqlmock.NewRows([]string{"colours", "name"}))
	}
	return DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(),
		Pipeline: "design", AgentType: "design-discovery-agent",
		BatchID: uuid.New(), Logger: zap.NewNop(),
	}, func() { db.Close() }
}

func TestPaletteContrastFiresOnTheShippedDefect(t *testing.T) {
	dctx, done := ctxWithPalette(t, brokenColoursJSON, true)
	defer done()

	res, err := (&PaletteContrastCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}
	// The three the owner could see.
	if len(res.Findings) != 3 {
		t.Errorf("want 3 findings, got %d: %+v", len(res.Findings), res.Findings)
	}

	wi := res.WorkItems[0]
	// A capability_gap, NOT a dispatch: there is no palette-repair handler and
	// inventing one recreates bugs_closed/077. If someone later gives this a
	// HandlerAgent, this assertion is where they should have to think about it.
	if wi.ItemType != "capability_gap" {
		t.Errorf("item_type = %q, want capability_gap", wi.ItemType)
	}
	if wi.HandlerAgent != "" {
		t.Errorf("handler_agent = %q, want empty — routing this at an agent that "+
			"cannot act on it is bugs_closed/077's exact mechanism", wi.HandlerAgent)
	}
	// 1.11:1 is invisible, not "a bit low". A check that grades every failure
	// the same makes the reader do the triage.
	if wi.Severity != "high" {
		t.Errorf("severity = %q, want high (worst pairing is 1.11:1)", wi.Severity)
	}
}

func TestPaletteContrastIsSilentOnTheRepairedPalette(t *testing.T) {
	dctx, done := ctxWithPalette(t, repairedColoursJSON, true)
	defer done()

	res, err := (&PaletteContrastCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Findings) != 0 {
		t.Errorf("the repaired palette must produce nothing; got %d finding(s), %d item(s): %+v",
			len(res.Findings), len(res.WorkItems), res.Findings)
	}
}

func TestPaletteContrastSkipsASiteWithNoResolvedPalette(t *testing.T) {
	dctx, done := ctxWithPalette(t, "", false)
	defer done()

	res, err := (&PaletteContrastCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("a site with no resolved composition must not error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("a site that has never had a stylesheet rendered has nothing to be wrong about; got %d item(s)", len(res.WorkItems))
	}
}

// TestPaletteContrastSkipsNonHexValues pins the one direction a legibility
// check must never fail in. `footer_text` is a live rgba() value; judging it on
// its opaque colour would OVER-report contrast, i.e. call a failing pair
// passing. Skipping is the conservative choice and is deliberate.
func TestPaletteContrastSkipsNonHexValues(t *testing.T) {
	dctx, done := ctxWithPalette(t, brokenColoursJSON, true)
	defer done()

	res, _ := (&PaletteContrastCheck{}).Run(dctx)
	for _, f := range res.Findings {
		for _, k := range []string{"foreground", "background"} {
			if v, _ := f[k].(string); len(v) > 0 && v[0] != '#' {
				t.Errorf("finding reports a non-hex %s %q — rgba()/var() values must be skipped, "+
					"not judged on their opaque colour", k, v)
			}
		}
	}
}

func TestSeverityGrading(t *testing.T) {
	for _, tc := range []struct {
		ratio float64
		want  string
	}{
		{1.11, "high"},   // invisible — the live eyebrow
		{2.32, "medium"}, // the regression the palette repair introduced
		{3.23, "low"},    // readable but below AA — the live card body
	} {
		if got := severityFor(tc.ratio); got != tc.want {
			t.Errorf("severityFor(%.2f) = %q, want %q", tc.ratio, got, tc.want)
		}
	}
}
