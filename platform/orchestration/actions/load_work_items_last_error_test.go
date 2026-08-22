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
	"time"

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
	return loadOneItemFull(t, attemptCount, errText, nil)
}

// loadOneItemCoded drives a row carrying BOTH halves of the typed channel.
// bugs_open/345: the loader reads retry_feedback->>'message' and
// retry_feedback->>'code', never site_work_items.error.
func loadOneItemCoded(t *testing.T, attemptCount int, errText, code interface{}) map[string]interface{} {
	t.Helper()
	return loadOneItemAll(t, attemptCount, errText, code, nil)
}

// loadOneItemFull additionally sets completed_at — the round-3 discriminator for
// a prior completed lifecycle (LANDMINES.md:7104).
func loadOneItemFull(t *testing.T, attemptCount int, errText, completedAt interface{}) map[string]interface{} {
	t.Helper()
	return loadOneItemAll(t, attemptCount, errText, nil, completedAt)
}

// loadOneItemAll is the one place the mocked column list lives. It is declared
// POSITIONALLY, which is why adding a column to the loader's SELECT breaks every
// test in this package at once — that breakage is the feature: a silent scan
// misalignment drops rows, and a dropped row is invisible at fleet scale
// (bugs_closed/078).
func loadOneItemAll(t *testing.T, attemptCount int, errText, code, completedAt interface{}) map[string]interface{} {
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
		"last_error", "last_error_code", "completed_at",
	}
	rows := sqlmock.NewRows(cols).
		AddRow(uuid.New(), siteID, "component_selector", "build", "needs_new_component",
			"medium", "Need component template", []byte(`{"section_type":"mortgages-repayment"}`), nil,
			50, "component-creator", "triaged", "needs_new_component:mortgages-repayment",
			nil, attemptCount, "auto",
			nil, nil, nil, errText, code, completedAt)

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

// Round 3 (corr 67b07528, prior_art_librarian HIGH): wi.error can be STALE.
// The success UPDATE writes status/result/completed_at and never touches error
// (LANDMINES.md:7104, measured 2026-08-08: status='complete' with a refusal
// still in error). A completed item hand-reset to 'triaged' would therefore
// show a PREVIOUS LIFECYCLE's failure to a fresh generation. completed_at
// survives such resets and no genuinely-failing item ever has one — so it is
// the discriminator, and a prior-lifecycle item must carry NOTHING.
func TestLoadWorkItems_PriorCompletedLifecycleCarriesNothing(t *testing.T) {
	item := loadOneItemFull(t, 0,
		"completion blocked: post-fix verification found the defect still present",
		time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC))

	if v, present := item["last_error"]; present {
		t.Errorf("last_error present (%v) on an item with a prior completed lifecycle — feeding a fresh generation a dead lifecycle's refusal is the round-3 hazard", v)
	}
}

// ── The typed channel (migration 561) ────────────────────────────────────────
//
// Everything above pins the GATE. These pin the SOURCE, which is the 2026-08-22
// change: the feedback no longer comes from `site_work_items.error`.
//
// Why it moved. `error` is a general-purpose annotation column and the prompt
// reading this key asserts a provenance it cannot support ("your previous
// output for this component was refused by validation"). [MEASURED 2026-08-22,
// live clients_db] TWENTY write sites across TEN files write that column,
// three of them the human operator HTTP path in admin/site_admin_handlers.go.
// Of 799 rows fleet-wide passing the gate above, 405 were human notes and only
// 11 were validation rejections; of the 17 that could reach a reader, 6 (35%)
// were misattributed — 3 token-cap truncations and 3 lane notes such as
// "HELD 2026-08-18 by the loanzy_uk_example_site lane: …".

// The load-bearing property, and the only one here that reads PRODUCTION SQL
// rather than a mocked row: the loader must select the typed channel and must
// NOT select wi.error. sqlmock matches the query text the action actually
// issues, so this is an assertion about the code's behaviour — not about a
// comment, and not a re-implementation of it.
func TestLoadWorkItems_ReadsTheTypedChannelNotTheErrorColumn(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(`retry_feedback`).WillReturnRows(sqlmock.NewRows([]string{}))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"site_id": siteID.String()},
		},
		StepConfig: models.Step{Config: map[string]interface{}{"site_id": "input_data.site_id"}},
	}

	if _, err := LoadWorkItemsAction(context.Background(), params); err != nil {
		t.Fatalf("LoadWorkItemsAction: %v", err)
	}
	// The expectation above only fires if the issued SQL names retry_feedback.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the loader did not read the typed channel: %v", err)
	}
}

// The code rides WITH the message, so the prompt can render the remedy that
// matches the failure class instead of asserting one class for all of them.
func TestLoadWorkItems_RetryCarriesTheFailureCode(t *testing.T) {
	const rejection = `component validation rejected for function="mortgages-repayment": ` +
		`field "currency_symbol" declares source "site_specs.locale.currency_symbol"`

	item := loadOneItemCoded(t, 1, rejection, "component_validation_rejected")

	if got, _ := item["last_error"].(string); got != rejection {
		t.Errorf("last_error = %q, want %q", got, rejection)
	}
	got, ok := item["last_error_code"].(string)
	if !ok {
		t.Fatal("last_error_code absent — the prompt cannot tell a validation refusal from a token-cap truncation, which is the 35% misattribution this change exists to end")
	}
	if got != "component_validation_rejected" {
		t.Errorf("last_error_code = %q, want the producer's own classification", got)
	}
}

// A message whose class is UNKNOWN must not arrive wearing a class. The key is
// absent rather than empty: a prompt gating on presence reads "supplied empty"
// as "supplied", which is how a truncation gets rendered under "your output was
// refused by validation".
func TestLoadWorkItems_UnclassifiedFailureCarriesNoCode(t *testing.T) {
	for _, code := range []interface{}{nil, "", "   "} {
		item := loadOneItemCoded(t, 1, "something failed", code)
		if _, present := item["last_error"]; !present {
			t.Fatal("last_error should still be carried — only the CODE is unknown")
		}
		if v, present := item["last_error_code"]; present {
			t.Errorf("last_error_code present (%q) for code %v — want the key absent", v, code)
		}
	}
}
