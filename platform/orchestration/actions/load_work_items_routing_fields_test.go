// FILE: platform/orchestration/actions/load_work_items_routing_fields_test.go
//
// bugs_open/154 — a first-class column on site_work_items was invisible to
// every handler.
//
// LoadWorkItemsAction built current_item from a SELECT that omitted
// site_work_items.component_id, so the only path a dispatcher could reference
// was current_item.spec.component_id — a copy the creating agent had to
// remember to duplicate into the spec JSONB. tool-auditor populated the column
// and not the blob, so its improve_tool items reached tool-improver with
// input_data.component_id unresolved and died at load_tool's query_database.
// Items from three other creators, which wrote the blob, ran clean. Measured
// 2026-07-31: 4 of 4 tool-auditor items column-only, 235 rows spec-only.
//
// The tests below pin the three properties that make ONE dispatcher path
// correct for both populations, plus the two things the fix must NOT do.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestSetRoutingField_ColumnFirstThenSpec covers the resolution order itself.
func TestSetRoutingField_ColumnFirstThenSpec(t *testing.T) {
	const colVal = "11111111-1111-1111-1111-111111111111"
	const specVal = "22222222-2222-2222-2222-222222222222"

	cases := []struct {
		name   string
		column string
		spec   map[string]interface{}
		want   string
		absent bool
	}{
		{
			// The tool-auditor population: the creator used the schema
			// properly, and that is exactly what made the item undispatchable.
			name:   "column set, spec missing — the bugs_open/154 case",
			column: colVal,
			spec:   map[string]interface{}{},
			want:   colVal,
		},
		{
			// The 235-row majority. Must keep working unchanged.
			name:   "column empty, spec set — the established convention",
			column: "",
			spec:   map[string]interface{}{"component_id": specVal},
			want:   specVal,
		},
		{
			// The column is the authority: it is the typed, indexed, FK-shaped
			// source, and spec is a hand-maintained copy that can go stale.
			name:   "both set — the column wins",
			column: colVal,
			spec:   map[string]interface{}{"component_id": specVal},
			want:   colVal,
		},
		{
			// The key must stay ABSENT, not become "". The dispatch mapping
			// marks it optional ("component_id?"), and an optional path that
			// RESOLVES is forwarded while one that is MISSING is skipped —
			// so materialising "" turns "not supplied" into "supplied empty"
			// for handlers that gate on presence (create_rerender_items).
			name:   "neither set — key absent, NOT empty string",
			column: "",
			spec:   map[string]interface{}{},
			absent: true,
		},
		{
			// A non-string under the same name is not an id; refuse it rather
			// than forward something a uuid parse will choke on downstream.
			name:   "spec holds a non-string — refused, key absent",
			column: "",
			spec:   map[string]interface{}{"component_id": 42},
			absent: true,
		},
		{
			name:   "spec holds an empty string — treated as absent",
			column: "",
			spec:   map[string]interface{}{"component_id": ""},
			absent: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := map[string]interface{}{"spec": tc.spec}
			setRoutingField(item, "component_id", tc.column, zap.NewNop())

			got, present := item["component_id"]
			if tc.absent {
				if present {
					t.Errorf("component_id should be ABSENT, got %#v", got)
				}
				return
			}
			if !present {
				t.Fatal("component_id absent; the dispatch mapping would skip it and the handler would die on a nil param")
			}
			if got != tc.want {
				t.Errorf("component_id = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSetRoutingField_NeverMutatesSpec is the guard on the design choice.
//
// Backfilling the resolved value INTO spec was the other candidate fix and it
// is unsafe: rerender-pages reads input_data.spec.component_id, and
// create_rerender_items gates
//
//	scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""
//
// on it — so a write into spec could silently flip a site-wide rerender into a
// component-scoped one. Top-level exposure leaves every spec.* reader reading
// exactly what it reads today. This test fails if anyone "tidies" that up.
func TestSetRoutingField_NeverMutatesSpec(t *testing.T) {
	spec := map[string]interface{}{"reason": "section_data_resolved"}
	item := map[string]interface{}{"spec": spec}

	setRoutingField(item, "component_id", "33333333-3333-3333-3333-333333333333", zap.NewNop())

	if _, leaked := spec["component_id"]; leaked {
		t.Error("component_id was written into spec — this can flip create_rerender_items " +
			"from a site-wide rerender to a component-scoped one; expose top-level only")
	}
	if len(spec) != 1 {
		t.Errorf("spec was mutated: %#v", spec)
	}
}

// TestSetRoutingField_ToleratesAbsentSpec — spec is nullable in the schema and
// unmarshals to a non-map when it holds a JSON scalar, so this must not panic.
func TestSetRoutingField_ToleratesAbsentSpec(t *testing.T) {
	for _, item := range []map[string]interface{}{
		{},
		{"spec": nil},
		{"spec": "not-a-map"},
	} {
		setRoutingField(item, "component_id", "", zap.NewNop())
		if _, present := item["component_id"]; present {
			t.Errorf("no column and no usable spec must leave the key absent: %#v", item)
		}
	}
	// The column still resolves when spec is unusable.
	item := map[string]interface{}{"spec": "not-a-map"}
	setRoutingField(item, "component_id", "44444444-4444-4444-4444-444444444444", zap.NewNop())
	if item["component_id"] != "44444444-4444-4444-4444-444444444444" {
		t.Errorf("column value must resolve regardless of spec's shape: %#v", item)
	}
}

// TestLoadWorkItems_ExposesRoutingColumns is the end-to-end regression.
//
// It exercises the real SELECT and the real Scan, so it fails if the two ever
// drift out of alignment — the failure mode that matters here, because a
// mis-ordered scan drops the row through the `continue` and the site is
// re-selected forever with nothing to dispatch (the bugs_closed/078 livelock).
func TestLoadWorkItems_ExposesRoutingColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	componentID := uuid.New()
	specOnlyComponent := uuid.New()

	cols := []string{
		"id", "site_id", "source", "pipeline", "item_type",
		"severity", "summary", "spec", "page_id",
		"priority", "handler_agent", "status", "item_key",
		"batch_id", "attempt_count", "approval_mode",
		"component_id", "entity_id", "affected_url",
	}

	rows := sqlmock.NewRows(cols).
		// The tool-auditor shape: column set, spec carries no component_id.
		AddRow(uuid.New(), siteID, "audit", "build", "improve_tool",
			"medium", "audit fix", []byte(`{"reason":"audit"}`), nil,
			60, "tool-improver", "triaged", "audit_fix_example.com",
			nil, 0, "auto",
			componentID, nil, nil).
		// The majority shape: column NULL, spec carries it.
		AddRow(uuid.New(), siteID, "acceptance", "build", "improve_tool",
			"medium", "acceptance fail", []byte(`{"component_id":"`+specOnlyComponent.String()+`"}`), nil,
			60, "tool-improver", "triaged", "acceptance_fail:x",
			nil, 0, "auto",
			nil, nil, nil).
		// Neither: the key must not appear at all.
		AddRow(uuid.New(), siteID, "generic", "build", "needs_rerender",
			"low", "rerender", []byte(`{}`), nil,
			60, "rerender-pages", "triaged", "rerender:x",
			nil, 0, "auto",
			nil, nil, nil)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"site_id": siteID.String()},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": "input_data.site_id",
		}},
	}

	out, err := LoadWorkItemsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("LoadWorkItemsAction: %v", err)
	}

	result := out.(map[string]interface{})
	items := result["items"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("loaded %d items, want 3 — a scan misalignment drops rows silently", len(items))
	}

	first := items[0].(map[string]interface{})
	if first["component_id"] != componentID.String() {
		t.Errorf("column-only item: component_id = %v, want %v — this is the bugs_open/154 failure",
			first["component_id"], componentID.String())
	}

	second := items[1].(map[string]interface{})
	if second["component_id"] != specOnlyComponent.String() {
		t.Errorf("spec-only item: component_id = %v, want %v — the fallback regressed the 235-row majority",
			second["component_id"], specOnlyComponent.String())
	}

	third := items[2].(map[string]interface{})
	if got, present := third["component_id"]; present {
		t.Errorf("item with neither source must leave component_id absent, got %#v", got)
	}

	// page_id stays column-only and untouched: measured 2026-07-31, 218 rows
	// have a NULL column and a spec.page_id, every one of which would newly
	// gain current_item.page_id if this fix were extended to it for symmetry.
	if _, present := first["page_id"]; present {
		t.Error("page_id must remain column-only — extending the spec fallback to it " +
			"widens what reaches 218 items' handlers without editing them")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
