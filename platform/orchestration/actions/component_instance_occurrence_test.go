// FILE: platform/orchestration/actions/component_instance_occurrence_test.go
//
// bugs_closed/283 / RFC_032 step 3 — the two single-section render paths stamped
// a CONSTANT occurrence 0 on every instance, so any per-section render (a build,
// a content_rewrite, a section edit) re-collided every multi-instance page it
// touched. Reproduced at the served artefact on 2026-08-24:
// gaswholesalers.com/pricing-transparency.html and vetcomparison.uk/how-it-works.html
// each carried id="c-generic-text-block" TWICE.
//
// Every test below names the mutation that kills it. A test that cannot fail is
// not evidence, and the whole defect class here is "a check that passes while
// blind" — so the controls matter as much as the assertions.
package actions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// loopConfig builds the config a loop-expanded render step actually carries.
// The index is float64 on purpose: it round-trips through orchestration_states
// JSONB between plan time and execution, and that is the shape that arrives.
func loopConfig(loopName string, idx float64) map[string]interface{} {
	return map[string]interface{}{
		"loop_name":       loopName,
		"loop_item_index": idx,
		"loop_var_name":   "current_section",
	}
}

// loopItems builds CollectedData holding one sectionPlanItem-shaped entry per
// function, under the real key format.
func loopItems(loopName string, functions ...string) map[string]interface{} {
	c := map[string]interface{}{}
	for i, fn := range functions {
		c[datahelpers.LoopItemKey(loopName, i)] = map[string]interface{}{
			"name":     fn,
			"function": fn,
			"component": map[string]interface{}{
				"function": fn,
				"name":     strings.ToUpper(fn), // decoy display name — must never be read
			},
		}
	}
	return c
}

func boundToken(rc *RenderContext) string {
	s, _ := rc.ContentData[InstanceContentKey].(string)
	return s
}

// The defect itself. Three prior sections, two of them the same function as
// ours, so we are the THIRD instance and must take occurrence 2 ("-3").
//
// MUTATION THAT KILLS IT: revert DeriveAndBindInstanceToken's loop arm to a
// constant `occ := 0` — i.e. the code that shipped before this change. Token
// becomes the bare "c-generic-text-block" and the assertion fails.
func TestDeriveAndBind_loopCountsSameFunctionPredecessors(t *testing.T) {
	rc := &RenderContext{}
	p := PlacementFromLoopStep(
		loopConfig("process_sections_loop", 3),
		loopItems("process_sections_loop", "generic-text-block", "hero", "generic-text-block"))

	DeriveAndBindInstanceToken(context.Background(), nil, rc, "generic-text-block", p, zap.NewNop())

	if got, want := boundToken(rc), "c-generic-text-block-3"; got != want {
		t.Fatalf("third instance of a repeated component must take occurrence 2: got %q want %q", got, want)
	}
}

// The derived occurrence must agree with the canonical whole-page rule, INCLUDING
// its case/whitespace key equality — live data carries both "FAQ " and "faq".
//
// MUTATION THAT KILLS IT: drop the lower/trim from instanceFunctionKey (or from
// InstanceCounter.Next). The two derivations then disagree on the mixed-case
// entries and the loop below fails at index 2.
func TestDeriveAndBind_agreesWithInstanceCounter(t *testing.T) {
	functions := []string{"faq", "hero", "FAQ ", "generic-text-block", " faq"}
	canonical := InstanceTokensForPage(functions)

	for i, fn := range functions {
		rc := &RenderContext{}
		p := PlacementFromLoopStep(
			loopConfig("process_sections_loop", float64(i)),
			loopItems("process_sections_loop", functions...))
		DeriveAndBindInstanceToken(context.Background(), nil, rc, fn, p, zap.NewNop())

		if got := boundToken(rc); got != canonical[i] {
			t.Fatalf("index %d (%q): single-section derivation %q disagrees with the canonical walk %q — "+
				"two paths rendering the same page must agree on every token",
				i, fn, got, canonical[i])
		}
	}
}

// The reader of the loop-expansion contract. Three things are pinned at once:
// the float64 index, the item-key format, and the refusal to read the DISPLAY
// name off the component map.
//
// MUTATION THAT KILLS IT (any of three): change datahelpers.LoopItemKey's format
// so the items are not found (count drops to 0); read config["loop_item_index"]
// with a bare .(int) assertion (idx becomes 0, no predecessors); or make
// functionOfLoopItem fall back to component["name"] (the decoy is upper-cased,
// so it stops matching and the count drops).
func TestPlacementFromLoopStep_readsTheExpanderContract(t *testing.T) {
	p := PlacementFromLoopStep(
		loopConfig("process_sections_loop", 4),
		loopItems("process_sections_loop", "faq", "faq", "hero", "faq", "faq"))

	if !p.LoopKnown {
		t.Fatal("loop context present but LoopKnown false — the fallback would silently take over")
	}
	if len(p.PriorFunctions) != 4 {
		t.Fatalf("expected the 4 items before index 4, got %d (%v)", len(p.PriorFunctions), p.PriorFunctions)
	}
	for i, want := range []string{"faq", "faq", "hero", "faq"} {
		if p.PriorFunctions[i] != want {
			t.Fatalf("prior %d: got %q want %q — a DISPLAY name here would neither match the "+
				"canonical walk nor be stable", i, p.PriorFunctions[i], want)
		}
	}
}

// An index that arrives as json.Number (decoders configured with UseNumber) must
// not read as absent — absent means occurrence 0 on every instance, which is
// indistinguishable from the defect.
//
// MUTATION THAT KILLS IT: delete the jsonNumberInt arm from placementInt.
func TestPlacementFromLoopStep_acceptsJSONNumberIndex(t *testing.T) {
	cfg := map[string]interface{}{
		"loop_name":       "process_sections_loop",
		"loop_item_index": json.Number("2"),
	}
	p := PlacementFromLoopStep(cfg, loopItems("process_sections_loop", "faq", "faq", "faq"))
	if len(p.PriorFunctions) != 2 {
		t.Fatalf("json.Number index must be read: got %d priors, want 2", len(p.PriorFunctions))
	}
}

// The derivation is an INPUT-IMPROVER, never a gate. With no context at all it
// must bind occurrence 0 — exactly what shipped before — and it must still bind.
//
// MUTATION THAT KILLS IT: make DeriveAndBindInstanceToken return early without
// calling BindInstanceToken when it cannot derive. The token goes empty, and an
// empty {{.InstanceID}} renders id="" — the Half B defect class.
func TestDeriveAndBind_noContextBindsOccurrenceZeroAndStillBinds(t *testing.T) {
	rc := &RenderContext{}
	DeriveAndBindInstanceToken(context.Background(), nil, rc, "faq", SectionPlacement{}, zap.NewNop())

	if got, want := boundToken(rc), "c-faq"; got != want {
		t.Fatalf("no placement context must fall back to occurrence 0 (today's behaviour): got %q want %q", got, want)
	}
}

// A nil logger must not turn a diagnostic into a panic on a live render path.
//
// MUTATION THAT KILLS IT: use `logger` directly instead of logf(logger).
func TestDeriveAndBind_toleratesNilLogger(t *testing.T) {
	rc := &RenderContext{}
	DeriveAndBindInstanceToken(context.Background(), nil, rc, "faq", SectionPlacement{}, nil)
	if boundToken(rc) != "c-faq" {
		t.Fatal("nil logger must not change the binding")
	}
}

// A failed lookup must not fail the render. The count is an improvement to an
// input; refusing on it would turn a diagnostic into an outage.
//
// MUTATION THAT KILLS IT: make DeriveAndBindInstanceToken propagate
// storedPredecessorCount's error (change its signature to return error and have
// the call sites act on it) — or skip the bind on the error branch.
func TestDeriveAndBind_storedLookupErrorFallsBackAndStillBinds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("count\\(\\*\\)").WillReturnError(errors.New("boom"))

	rc := &RenderContext{}
	p := SectionPlacement{PageID: "11111111-1111-1111-1111-111111111111", Position: 3, RowID: "22222222-2222-2222-2222-222222222222"}
	DeriveAndBindInstanceToken(context.Background(), db, rc, "faq", p, zap.NewNop())

	if got, want := boundToken(rc), "c-faq"; got != want {
		t.Fatalf("a lookup error must degrade to occurrence 0, not fail the render: got %q want %q", got, want)
	}
}

// The editor path's count: position-exact, with the (position, id) tie arm that
// matches loadStoredSections' ORDER BY.
//
// MUTATION THAT KILLS IT: drop the tie arm from the withTie query (the argument
// count no longer matches and the expectation fails), or drop the
// build_status <> 'removed' filter (the regexp expectation no longer matches).
func TestStoredPredecessorCount_positionExactWithTieBreak(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	page := "11111111-1111-1111-1111-111111111111"
	row := "22222222-2222-2222-2222-222222222222"

	mock.ExpectQuery(`build_status IS DISTINCT FROM 'removed'`).
		WithArgs(page, "generic-text-block", 3, row).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	n, err := storedPredecessorCount(context.Background(), db,
		"generic-text-block", SectionPlacement{PageID: page, Position: 3, RowID: row})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("got %d predecessors, want 1", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the query shape changed: %v", err)
	}
}

// Without a row id there is no tie to break, and the query must drop to a strict
// `position <` with THREE arguments — passing a fourth would error at the driver.
//
// MUTATION THAT KILLS IT: always use the withTie query. sqlmock then sees 4 args
// against a 3-arg expectation and fails.
func TestStoredPredecessorCount_withoutRowIDUsesStrictPosition(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	page := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery(`pc.position < \$3`).
		WithArgs(page, "faq", 2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	if _, err := storedPredecessorCount(context.Background(), db, "faq",
		SectionPlacement{PageID: page, Position: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the no-tie query shape changed: %v", err)
	}
}

// PlacementFromStoredRow must read a position that arrived as float64 through
// workflow state. A float64 read as absent yields Position 0, which fails the
// `Position > 0` guard and silently drops the whole editor path back to
// occurrence 0 — the defect, wearing the fallback's clothes.
//
// MUTATION THAT KILLS IT: read pcData["position"] with a bare .(int) assertion.
func TestPlacementFromStoredRow_acceptsFloatPosition(t *testing.T) {
	p := PlacementFromStoredRow(map[string]interface{}{
		"id":       "22222222-2222-2222-2222-222222222222",
		"page_id":  "11111111-1111-1111-1111-111111111111",
		"position": float64(4),
	})
	if p.Position != 4 {
		t.Fatalf("a float64 position must be read as 4, got %d — this is how the editor path "+
			"silently reverts to occurrence 0", p.Position)
	}
	if p.PageID == "" || p.RowID == "" {
		t.Fatal("page id and row id must both be carried; without them the count cannot be taken")
	}
}

// The canonical walk's ordering must be deterministic, or the two derivations can
// disagree on a page with a (position) tie.
//
// Asserted at the QUERY, not by scanning the source: a source-scan test makes the
// file's own COMMENTS load-bearing, and this change adds a comment that mentions
// the very ordering it would search for — so the needle would match the prose and
// pass vacuously even if the SQL were reverted (LANDMINES: "a source-scanning test
// makes comments load-bearing — first occurrence wins").
//
// MUTATION THAT KILLS IT: revert loadStoredSections to `ORDER BY position ASC`.
func TestLoadStoredSections_ordersByPositionThenID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	page := uuid.New()
	mock.ExpectQuery(`ORDER BY position ASC, id ASC`).
		WithArgs(page).
		WillReturnRows(sqlmock.NewRows([]string{
			"component_id", "slot_name", "content_data", "rendered_html", "position", "component_version_id",
		}))

	if _, err := loadStoredSections(context.Background(), db, page, zap.NewNop()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("loadStoredSections no longer orders by (position, id): %v", err)
	}
}
