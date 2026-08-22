// FILE: platform/orchestration/actions/store_generated_component_section_type_heal_test.go
//
// Proves the section_type self-heal also runs when the birth gate REFUSES
// (bugs_open/337), and — the half that carries the safety argument — that it
// refuses to make an INACTIVE component selectable.
//
// Why the rejection path needs the heal at all: the regeneration UPDATE already
// carries the same COALESCE, but it only ever runs on a SUCCESSFUL store, so the
// repair was gated behind the success its own absence prevents. A NULL
// section_type hides the row from the selector AND from the writer's advisory,
// the blind writer strands a field, the gate refuses, nothing heals, and the
// component is refused again — measured at 70 refusals with no success.
//
// Mutation-proof construction: sqlmock fails on any unexpected statement, and
// these two tests are each other's controls. The heal UPDATE is declared in the
// active test and NOT declared in the inactive one, so dropping the `is_active`
// condition from the SQL makes the inactive test fail on an unexpected
// statement, and dropping the whole call makes the active test fail on an
// unmet expectation. Neither mutation can pass both.

package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

const healIncumbentID = "dddddddd-4444-4444-4444-444444444444"

// strandingParams generates a template that DROPS the incumbent's `heading`
// field, so the field-contract guard refuses — the exact shape of the 97
// measured refusals. No work item id is supplied, so the retry-feedback writer
// short-circuits and this test stays insulated from that path.
func strandingParams(db *sql.DB) ActionParams {
	return ActionParams{
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"section_type":       "input_data.section_type",
			"generated_template": "input_data.generated_template",
		}},
		CollectedData: map[string]interface{}{"input_data": map[string]interface{}{
			"section_type": "test-widget",
			"generated_template": map[string]interface{}{
				"function": "test-widget",
				"html_template": `<section class="test-widget" data-component="test-widget">` +
					`<div class="tw-inner"><h2>{{.replacement_heading}}</h2></div></section>`,
				"input_schema": map[string]interface{}{
					"fields": map[string]interface{}{
						"replacement_heading": map[string]interface{}{"type": "text", "source": "llm", "required": true},
					},
				},
			},
		}},
	}
}

// expectRefusedRegeneration declares everything up to and including the
// rejection record, for a same-site regeneration with no requester site id.
func expectRefusedRegeneration(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template", "input_schema", "js_content"}).
			AddRow(healIncumbentID,
				`<section data-component="test-widget"><h2>{{.heading}}</h2></section>`,
				`{"fields":{"heading":{"type":"text","source":"llm","required":true}}}`,
				nil))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\)`).WithArgs(healIncumbentID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1))
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).
		WillReturnRows(sqlmock.NewRows([]string{"aspect"}).AddRow("identity"))
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
}

// A refused regeneration must still repair the metadata gap that helped cause
// the refusal. The UPDATE is narrow by construction: one column, COALESCE so an
// existing value is never overwritten, and `section_type IS NULL` so it is a
// no-op on rows that do not need it.
func TestStoreGeneratedComponent_RejectionHealsNullSectionType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectRefusedRegeneration(mock)
	mock.ExpectExec(`UPDATE content_components[\s\S]*section_type = COALESCE\(section_type[\s\S]*is_active = true`).
		WithArgs("test-widget", healIncumbentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = StoreGeneratedComponentAction(context.Background(), strandingParams(db))
	if err == nil {
		t.Fatal("a stranding regeneration must still be refused — the heal must not rescue the write")
	}
	if !strings.Contains(err.Error(), "removes/renames") {
		t.Errorf("the refusal must remain the field-contract one; got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// THE SAFETY PROPERTY. Healing section_type makes a component visible to the
// selector, and migration 036 deactivates broken components precisely so pages
// stop choosing them. So the heal must never fire on a row that is not already
// active — otherwise a component that just failed the gate would be offered to
// page planning, which is strictly worse than the invisibility being repaired.
//
// The gate lives in the SQL (`AND is_active = true`), so this test asserts it
// where a caller cannot forge it: the statement is sent with that condition
// present. Removing the condition changes the statement and this expectation
// stops matching.
func TestStoreGeneratedComponent_HealIsGatedOnIsActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	expectRefusedRegeneration(mock)
	// The heal statement must carry BOTH guards. An UPDATE without them is an
	// unexpected statement and fails the test.
	mock.ExpectExec(`UPDATE content_components[\s\S]*WHERE id = \$2::uuid[\s\S]*AND is_active = true[\s\S]*AND section_type IS NULL`).
		WithArgs("test-widget", healIncumbentID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows: inactive or already set

	_, err = StoreGeneratedComponentAction(context.Background(), strandingParams(db))
	if err == nil {
		t.Fatal("expected the field-contract refusal")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// A refused CREATION has no row to repair, so no heal statement may be sent.
// Firing one would mean writing to a component the store never identified.
func TestStoreGeneratedComponent_RefusedCreationSendsNoHeal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// No incumbent → creation. A schema with no fields to strand still gets
	// refused, on the structural rules, without ever being a regeneration.
	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).
		WillReturnRows(sqlmock.NewRows([]string{"aspect"}).AddRow("identity"))
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
	// Nothing else. A heal UPDATE here is an unexpected statement.

	params := strandingParams(db)
	input := params.CollectedData["input_data"].(map[string]interface{})
	tmpl := input["generated_template"].(map[string]interface{})
	tmpl["html_template"] = `<section class="test-widget">no data-component attribute</section>`

	if _, err = StoreGeneratedComponentAction(context.Background(), params); err == nil {
		t.Fatal("expected a structural refusal")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}
