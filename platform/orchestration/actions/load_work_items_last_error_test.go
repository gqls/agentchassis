// FILE: platform/orchestration/actions/load_work_items_last_error_test.go
//
// bugs_open/345 — a rejected component was regenerated from IDENTICAL inputs,
// so the writer never learned why it had been refused and every retry
// reproduced the same rejection.
//
// Measured 2026-08-21 before the fix: 99 `component_validation_rejected` rows
// across 3 sites, and EVERY work item with more than one rejection had exactly
// ONE distinct rejection reason — up to 52 rejections on a single item, which
// is ~17 generations inside each of its 3 dispatch attempts. The retry could
// not succeed, because `generate_template`'s inputs (`input_data`,
// `site_record`, `site_specs`, `existing_component`) carried nothing about the
// previous failure.
//
// The fix hands the previous attempt's failure text to the handler as
// `current_item.last_error`. These tests pin the three properties that make
// that safe, and the one that makes it useful.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// loadOneItem drives the real LoadWorkItemsAction over a single mocked row and
// returns the item map it built. Deliberately calls production code rather than
// re-implementing the row→map step: a test that MIRRORS the construction cannot
// catch drift in it (WRONG_CALLS, 2026-08-19).
func loadOneItem(t *testing.T, attemptCount int, errText interface{}) map[string]interface{} {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	cols := []string{
		"id", "site_id", "source", "pipeline", "item_type",
		"severity", "summary", "spec", "page_id",
		"priority", "handler_agent", "status", "item_key",
		"batch_id", "attempt_count", "approval_mode",
		"component_id", "entity_id", "affected_url",
		"error",
	}
	rows := sqlmock.NewRows(cols).
		AddRow(uuid.New(), siteID, "component_selector", "build", "needs_new_component",
			"medium", "Need component template", []byte(`{"section_type":"mortgages-repayment"}`), nil,
			50, "component-creator", "triaged", "needs_new_component:mortgages-repayment",
			nil, attemptCount, "auto",
			nil, nil, nil, errText)

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
	items := out.(map[string]interface{})["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("loaded %d items, want 1 — a scan misalignment drops rows silently", len(items))
	}
	return items[0].(map[string]interface{})
}

// The point of the whole fix: a retry can see why its predecessor was refused.
func TestLoadWorkItems_RetryCarriesThePreviousFailure(t *testing.T) {
	const rejection = `generated template for "mortgages-repayment" rejected by pre-store ` +
		`validation: field "currency_symbol" declares source "site_specs.locale.currency_symbol" ` +
		`but no site carries a site_specs aspect named "locale"`

	item := loadOneItem(t, 1, rejection)

	got, ok := item["last_error"].(string)
	if !ok {
		t.Fatal("last_error absent on a retry — the writer is regenerating blind, which is bugs_open/345 itself")
	}
	if got != rejection {
		t.Errorf("last_error was altered.\n got: %q\nwant: %q", got, rejection)
	}
}

// A genuinely FRESH item must be byte-identical to pre-345 behaviour. A fresh
// item is one whose error column is NULL — no INSERT path writes it — and the
// prompt block is `{{if}}`-guarded, so an absent key renders nothing.
func TestLoadWorkItems_FreshItemCarriesNothing(t *testing.T) {
	item := loadOneItem(t, 0, nil)

	if v, present := item["last_error"]; present {
		t.Errorf("last_error present on a fresh item (%v) — a first generation must be unchanged", v)
	}
}

// An item with a recorded failure carries it EVEN AT attempt_count == 0.
// This was the opposite before council round 1 (corr 67b07528), whose own
// verification query showed the 52-rejection item was 52 DISTINCT
// orchestrations while attempt_count reached only 3: dispatches happen without
// consuming an attempt (the ladder's transient release writes error and
// deliberately does not count), so an attempt_count gate hides the failure
// text from exactly the re-dispatches that need it. The error column is the
// truthful signal that a previous run failed; attempt_count is not.
func TestLoadWorkItems_UncountedRedispatchStillCarriesTheFailure(t *testing.T) {
	const rejection = "component validation rejected: regeneration removes/renames 18 existing schema field(s)"

	item := loadOneItem(t, 0, rejection)

	got, ok := item["last_error"].(string)
	if !ok {
		t.Fatal("last_error absent on an uncounted re-dispatch — the 49-of-52 population regenerates blind again")
	}
	if got != rejection {
		t.Errorf("last_error altered.\n got: %q\nwant: %q", got, rejection)
	}
}

// NULL and empty must both stay absent, not become "". A handler that gates on
// presence would otherwise read "supplied empty" as "supplied".
func TestLoadWorkItems_NoFailureTextStaysAbsent(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  interface{}
	}{
		{"NULL column", nil},
		{"empty string", ""},
		{"whitespace only", "   \n\t "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := loadOneItem(t, 2, tc.val)
			if v, present := item["last_error"]; present {
				t.Errorf("last_error present (%q) for %s — want the key absent", v, tc.name)
			}
		})
	}
}

// The text is partly quoted from the rejected artefact — for the source check it
// echoes ~60 aspect names — and it is heading back into a prompt. Bound it.
func TestLoadWorkItems_PreviousFailureIsCapped(t *testing.T) {
	huge := strings.Repeat("x", 5000)

	item := loadOneItem(t, 1, huge)

	got := item["last_error"].(string)
	if len(got) >= len(huge) {
		t.Fatalf("last_error not capped: %d chars passed through unbounded", len(got))
	}
	if !strings.HasSuffix(got, "…[truncated]") {
		t.Error("a capped value must SAY it was capped, or the reader treats a severed message as the whole reason")
	}
	if !strings.HasPrefix(got, strings.Repeat("x", 2000)) {
		t.Error("the cap must keep the START of the message — the guard names the offending field first")
	}
}
