package actions

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// bugs_open/327 — a PARTIAL write to `content_direction` must not shrink the brief.
//
// WHY THIS DRIVES THE WHOLE ACTION THROUGH A MOCK DB rather than testing
// FormatContentDirection directly. The defect was never in the formatter: handed a map,
// it renders that map faithfully. It was in WHICH MAP THE ACTION HANDED IT — the incoming
// partial, before the deep merge. A unit test of the formatter alone passes identically in
// the broken and the fixed world, which is why the bug survived from 2026-04-18 to
// 2026-08-19 with three live sites serving a fragment of their brief and every write
// logging success. So the assertion is made on the JSON the action actually inserts.

// The INSERT's `data` argument is captured with `captureArg`, which already exists in
// this package (`tool_acceptance_convergence_test.go`) because sqlmock exposes no public
// accessor for recorded args and the matcher is the only hook. Reused rather than
// duplicated.

// existingBrief is the state a real site is in: a full brief, with a `formatted`
// rendering of every key, as the classifier leaves it.
func existingBrief() map[string]interface{} {
	spec := map[string]interface{}{
		"voice":               "Direct and technical.",
		"things_to_avoid":     []interface{}{"the word 'seamless'", "urgency language"},
		"writing_rules":       []interface{}{"active voice", "short paragraphs"},
		"heading_style":       map[string]interface{}{"case": "sentence"},
		"terminology":         map[string]interface{}{"preferred": "agent"},
		"example_phrases":     []interface{}{"Agents fail in isolation."},
		"persuasion_approach": "Evidence before claim.",
	}
	spec["formatted"] = datahelpers.FormatContentDirection(spec)
	return spec
}

// writeContentDirection runs the real action against a mock DB and returns the `data`
// JSONB it inserted.
func writeContentDirection(t *testing.T, current, partial map[string]interface{}) map[string]interface{} {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, oldID := uuid.New(), uuid.New()
	currentJSON, _ := json.Marshal(current)
	var inserted string

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"id", "data"}).AddRow(oldID.String(), currentJSON))
	mock.ExpectExec("UPDATE site_specs").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO site_specs").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), captureArg{&inserted},
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	mock.ExpectCommit()

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"site_id": siteID.String(),
				"spec":    partial,
			},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id":      "input_data.site_id",
			"spec_data":    "input_data.spec",
			"aspect":       "content_direction",
			"source":       "test",
			"source_agent": "test-agent",
		}},
	}

	if _, err := WriteSiteSpecAction(context.Background(), params); err != nil {
		t.Fatalf("WriteSiteSpecAction: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
	if inserted == "" {
		t.Fatal("no data argument was captured from the INSERT")
	}

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(inserted), &out); err != nil {
		t.Fatalf("inserted data is not JSON: %v", err)
	}
	return out
}

// TestPartialWriteKeepsEveryLabelInTheBrief is the test whose absence let 327 ship.
func TestPartialWriteKeepsEveryLabelInTheBrief(t *testing.T) {
	// The realistic shape of the write that caused the damage: a small, careful
	// correction touching two keys, over a document holding seven.
	out := writeContentDirection(t, existingBrief(), map[string]interface{}{
		"voice":         "Direct, technical, and warmer than before.",
		"blog_strategy": "Two posts a week.",
	})

	formatted, _ := out["formatted"].(string)
	if formatted == "" {
		t.Fatal("no formatted field was written at all")
	}

	var missing []string
	for key, val := range out {
		if key == "formatted" || val == nil {
			continue
		}
		if !strings.Contains(formatted, datahelpers.HumaniseKey(key)+":") {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Errorf("these merged keys never reach the writer's brief: %v\nbrief was:\n%s",
			missing, formatted)
	}

	// The correction itself must be in there too. A brief that kept every old key by
	// ignoring the new write would satisfy the check above and be just as wrong.
	if !strings.Contains(formatted, "warmer than before") {
		t.Error("the incoming correction is absent from the brief")
	}
	if strings.Contains(formatted, "Direct and technical.") {
		t.Error("the superseded value survived in the brief")
	}
}

// TestBriefIsDeterministic covers the second defect in 327: `range` over a Go map is
// randomised, so an identical spec used to render differently on every call — which made
// a diff of two briefs useless for telling whether a correction had landed.
func TestBriefIsDeterministic(t *testing.T) {
	spec := existingBrief()
	first := datahelpers.FormatContentDirection(spec)
	for i := 0; i < 50; i++ {
		if got := datahelpers.FormatContentDirection(spec); got != first {
			t.Fatalf("rendering is not stable across calls (iteration %d)\nfirst:\n%s\ngot:\n%s",
				i, first, got)
		}
	}
	// One unsorted level is enough to destabilise the whole output, so assert the
	// nested-map case explicitly rather than trusting the top level to cover it.
	nested := map[string]interface{}{"terminology": map[string]interface{}{
		"preferred": "agent", "avoid": "bot", "context": "orchestration", "notes": "n/a",
	}}
	base := datahelpers.FormatContentDirection(nested)
	for i := 0; i < 50; i++ {
		if got := datahelpers.FormatContentDirection(nested); got != base {
			t.Fatalf("nested rendering is not stable (iteration %d)\nfirst:\n%s\ngot:\n%s",
				i, base, got)
		}
	}
}

// TestEmptyValuesAreNotWrittenAsLabels keeps the fix honest in the other direction: an
// empty key legitimately renders to nothing, so the drop check above must not be
// satisfiable by emitting bare labels for empty values.
func TestEmptyValuesAreNotWrittenAsLabels(t *testing.T) {
	out := datahelpers.FormatContentDirection(map[string]interface{}{
		"voice":            "Direct.",
		"compliance_rules": []interface{}{},
		"empty_note":       "",
	})
	if strings.Contains(out, "Compliance rules:") || strings.Contains(out, "Empty note:") {
		t.Errorf("empty values must not appear as labels:\n%s", out)
	}
	if !strings.Contains(out, "Voice:") {
		t.Errorf("a non-empty value must appear:\n%s", out)
	}
}

// TestBriefDoesNotNestItsOwnPreviousRendering is the control for the load-bearing
// assumption behind rendering from `merged` rather than from the partial, and it was
// added because a council reviewer graded that assumption "asserted, not shown" at high
// severity — correctly, on the evidence available: the skip it depends on lives in
// FormatContentDirection (`if key == "formatted" { continue }`) and no test named it.
//
// The hazard is real if that skip ever goes: `merged` carries the PREVIOUS `formatted`
// string at call time (the partial has no such key, so the merge keeps the old one), so a
// formatter that walked it would embed the whole previous brief inside the new one, on
// every write, compounding. That would be worse than the data loss it replaces.
//
// Asserting "each label appears exactly once" catches it precisely: a nested render
// repeats every label the previous rendering contained.
func TestBriefDoesNotNestItsOwnPreviousRendering(t *testing.T) {
	out := writeContentDirection(t, existingBrief(), map[string]interface{}{
		"voice": "Direct, technical, and warmer than before.",
	})
	formatted, _ := out["formatted"].(string)
	if formatted == "" {
		t.Fatal("no formatted field was written")
	}
	for _, label := range []string{"Voice:", "Things to avoid:", "Writing rules:",
		"Heading style:", "Terminology:", "Example phrases:", "Persuasion approach:"} {
		if n := strings.Count(formatted, label); n != 1 {
			t.Errorf("label %q appears %d times, want exactly 1 — the previous rendering "+
				"has been embedded in the new one\nbrief was:\n%s", label, n, formatted)
		}
	}
	// The stale rendering must not survive verbatim either: the old `Voice:` line is the
	// cheapest single witness of a nested blob.
	if strings.Contains(formatted, "Voice: Direct and technical.") {
		t.Error("the previous rendering's Voice line survives inside the new brief")
	}
}

// TestBriefIsDeterministicAtEveryDepth answers the second council objection: sorting was
// read as covering "only two levels". It is applied wherever a map is rendered, and
// FormatSpecValue recurses into itself, so depth is covered by construction — but by
// construction is not by test. MEASURED 2026-08-19: live content_direction data is at most
// two levels deep (0 three-level maps, 0 arrays-of-objects across all 25 current specs),
// so this fixture is deliberately DEEPER than anything in production.
func TestBriefIsDeterministicAtEveryDepth(t *testing.T) {
	deep := map[string]interface{}{
		"terminology": map[string]interface{}{
			"preferred": "agent",
			"avoid":     "bot",
			"per_audience": map[string]interface{}{
				"engineers": "agent",
				"buyers":    "assistant",
				"press":     "system",
				"nested_again": map[string]interface{}{
					"z": "last", "a": "first", "m": "middle",
				},
			},
		},
	}
	first := datahelpers.FormatContentDirection(deep)
	for i := 0; i < 50; i++ {
		if got := datahelpers.FormatContentDirection(deep); got != first {
			t.Fatalf("four-level rendering is not stable (iteration %d)\nfirst:\n%s\ngot:\n%s",
				i, first, got)
		}
	}
	// And the deepest values must actually be rendered, or stability would be trivial.
	for _, want := range []string{"first", "middle", "last"} {
		if !strings.Contains(first, want) {
			t.Errorf("deepest level is not rendered at all: %q missing from\n%s", want, first)
		}
	}
}

// TestNoAspectBranchDerivesBeforeTheMerge is the guard a council reviewer asked for
// (bug_historian, round 2): the fix closes the three aspect branches that exist TODAY, but
// the shape — a per-aspect field derived from the pre-merge `specMap` — stays reachable by
// any future branch, and this corpus has closed cases (086, 093, 042) that are exactly
// "the underlying rule stayed generic and got hit again on the untouched branch". Their
// point about a bug-file note being the unread verdict is well made, so this is a test.
//
// ⚠ IT SCANS SOURCE, so it is deliberately built to survive the known trap that a
// source-scanning test makes COMMENTS load-bearing: comment lines are stripped before
// anything is counted, which is why the long explanatory comments added around both call
// sites do not affect it.
func TestNoAspectBranchDerivesBeforeTheMerge(t *testing.T) {
	raw, err := os.ReadFile("site_spec_actions.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	var code []string
	for _, line := range strings.Split(string(raw), "\n") {
		if t := strings.TrimSpace(line); strings.HasPrefix(t, "//") {
			continue // comments must not be able to satisfy or break this test
		}
		code = append(code, line)
	}
	src := strings.Join(code, "\n")

	// 1. The aspect branches are enumerated. A fourth one must fail here and be
	//    accounted for deliberately — that is the whole point of the assertion.
	got := regexp.MustCompile(`if aspect == "([a-z_]+)"`).FindAllStringSubmatch(src, -1)
	var aspects []string
	for _, m := range got {
		aspects = append(aspects, m[1])
	}
	want := []string{"identity", "content_direction"}
	if !reflect.DeepEqual(aspects, want) {
		t.Errorf("aspect branches in WriteSiteSpecAction are %v, want %v.\n"+
			"A new branch is fine — but if it DERIVES a field (a formatted/summary/normalised "+
			"value) it must do so from `merged`, never from the incoming `specMap`, or it "+
			"reintroduces bugs_open/327. Account for it here.", aspects, want)
	}

	// 2. Ordering: every derivation must sit after the merge. Positions, not prose.
	merge := strings.Index(src, "merged := siteSpecDeepMerge(")
	format := strings.Index(src, "FormatContentDirection(")
	normalise := strings.Index(src, "normaliseServicesField(")
	if merge < 0 || format < 0 || normalise < 0 {
		t.Fatalf("could not locate the call sites (merge=%d format=%d normalise=%d) — "+
			"if one was renamed, update this test rather than deleting it", merge, format, normalise)
	}
	if format < merge {
		t.Error("FormatContentDirection is called BEFORE siteSpecDeepMerge — this is " +
			"bugs_open/327 exactly: the brief would describe only the incoming partial")
	}
	if normalise < merge {
		t.Error("normaliseServicesField is called BEFORE siteSpecDeepMerge, so identity " +
			"normalisation would run on the partial rather than the merged document")
	}
}
