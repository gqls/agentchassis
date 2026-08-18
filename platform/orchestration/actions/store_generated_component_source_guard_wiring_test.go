// FILE: platform/orchestration/actions/store_generated_component_source_guard_wiring_test.go
//
// Proves the source-vocabulary guard is actually INVOKED from
// StoreGeneratedComponentAction — not merely correct as a pure function
// (bugs_open/309; council fdb032c6, the objection three seats raised
// independently: a mutation deleting the wiring line would pass a suite that
// only tests the pure function). Same mutation-proof construction as the
// shared-component fence test: sqlmock fails the test on any unexpected
// statement, so a refusal must happen BEFORE the content_components INSERT —
// delete the wiring and the flow reaches an INSERT no expectation covers.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// sourceGuardParams builds a generation that passes every gate UPSTREAM of the
// source guard — real HTML structure, balanced tags, data-component attribute,
// template variables exactly matching the schema fields — so the only refusal
// left to fire is the one under test. (A refusal from an earlier guard would
// make this test pass for the wrong reason; the message assertion below pins
// which guard fired.)
func sourceGuardParams(db *sql.DB, schemaFields string) ActionParams {
	return ActionParams{
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"section_type":       "input_data.section_type",
			"generated_template": "input_data.generated_template",
		}},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"section_type": "test-widget",
				"generated_template": map[string]interface{}{
					"function": "test-widget",
					"html_template": `<section class="test-widget" data-component="test-widget">` +
						`<div class="tw-inner"><h2>{{.heading}}</h2>` +
						`{{if .link_url}}<a href="{{.link_url}}">{{.link_label}}</a>{{end}}` +
						`</div></section>`,
					"input_schema": map[string]interface{}{
						"fields": mustUnmarshalMap(schemaFields),
					},
				},
			},
		},
	}
}

func mustUnmarshalMap(s string) map[string]interface{} {
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		panic(err)
	}
	return m
}

// The phantom-aspect declaration (the exact bugs_open/309 shape) must be
// refused, and refused BY THE SOURCE GUARD — the error names the aspect.
func TestStoreGeneratedComponent_SourceGuardIsWired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1. Existing-component lookup: none (creation path).
	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").WillReturnError(sql.ErrNoRows)
	// 2. The guard's aspect-vocabulary read — 'blog' deliberately absent.
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).
		WillReturnRows(sqlmock.NewRows([]string{"aspect"}).AddRow("identity").AddRow("contact"))
	// 3. The rejection record. No content_components INSERT is expected:
	//    reaching one is the mutation this test exists to catch.
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))

	params := sourceGuardParams(db, `{
		"heading":    {"type": "text", "source": "llm", "required": true},
		"link_url":   {"type": "url",  "source": "site_specs.blog.post1_url", "required": true},
		"link_label": {"type": "text", "source": "static", "fallback": "Read more", "required": false}
	}`)

	_, err = StoreGeneratedComponentAction(context.Background(), params)
	if err == nil {
		t.Fatal("phantom-aspect schema must be REFUSED at store; got nil error")
	}
	if !strings.Contains(err.Error(), `no site carries a site_specs aspect named "blog"`) {
		t.Errorf("refusal must come from the source guard and name the phantom aspect; got %q", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}

// When the aspect-vocabulary read FAILS, the aspect half goes quiet (fail
// open) — but the skip is recorded durably, and the DB-free halves still
// gate: an unknown query name must still refuse the store.
func TestStoreGeneratedComponent_AspectReadFailureIsRecordedAndFailOpen(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT id::text`).WithArgs("test-widget").WillReturnError(sql.ErrNoRows)
	// Aspect read fails — the guard must not block on the aspect half...
	mock.ExpectQuery(`SELECT DISTINCT aspect FROM site_specs`).
		WillReturnError(sql.ErrConnDone)
	// ...the skip is recorded durably (SOURCE_GUARD_ASPECT_SET_UNAVAILABLE)...
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))
	// ...and the query-name half still refuses, recording the rejection.
	mock.ExpectExec(`INSERT INTO agent_error_log`).WillReturnResult(sqlmock.NewResult(1, 1))

	params := sourceGuardParams(db, `{
		"heading":    {"type": "text",  "source": "llm", "required": true},
		"link_url":   {"type": "url",   "source": "site_specs.blog.post1_url", "required": true},
		"link_label": {"type": "array", "source": "query.no_such_query", "required": false}
	}`)

	_, err = StoreGeneratedComponentAction(context.Background(), params)
	if err == nil {
		t.Fatal("unknown query name must still be refused when the aspect read fails; got nil error")
	}
	if strings.Contains(err.Error(), `site_specs aspect named "blog"`) {
		t.Errorf("aspect half must be SKIPPED on a failed vocabulary read; got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "no such query is registered") {
		t.Errorf("refusal must come from the query-name half; got %q", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock expectations: %v", err)
	}
}
