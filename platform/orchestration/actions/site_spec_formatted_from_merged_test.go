package actions

import (
	"context"
	"encoding/json"
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
