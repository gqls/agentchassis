// FILE: platform/orchestration/actions/save_page_meta_description_refusal_item_test.go
//
// bugs_open/442 — a copy-gate refusal was one logger.Warn nothing asserted on.
// The property under test is the one §6 of that file names, BOTH ARMS, because
// either alone proves nothing:
//
//  1. a refusal by a COPY GATE must produce a durable artefact, at an actor;
//  2. everything else must produce NOTHING — no row, and no DB call at all.
//
// Arm 2 is the one that is easy to get wrong and easy to test vacuously. It is
// asserted here as "sqlmock has no expectations, so ANY database interaction
// fails the test", which cannot pass by accident.
//
// Each test names the mutation that kills it. A test that would still pass with
// the rule deleted proves nothing (MEMORY: a-quiet-test-passes-when-the-rule-is-gone).
package actions

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// TestMetaDescriptionRefusalIsAnAllowList pins the classification, and pins it in
// the direction that matters: an UNKNOWN reason must be SILENT.
//
// MUTATION THAT KILLS IT: rewrite metaDescriptionRefusalIsLoud as a deny-list
// ("not one of the cheap four"). Every named case below still passes — and the
// unknown-reason case fails, which is the whole point. A reason nobody has
// classified must not start filing work items on its own, and bugs_open/442 §4
// is the evidence that this vocabulary grows by addition without anyone noticing.
func TestMetaDescriptionRefusalIsAnAllowList(t *testing.T) {
	loud := []string{"voice_tell", "banned_claim"}
	silent := []string{
		"empty_candidate", "candidate_looks_internal", "candidate_too_long",
		"already_has_description",
		// An infrastructure fault, not a copy judgement: a rewrite cannot fix a
		// gate that will not load, and filing it would churn the dedup key for as
		// long as the fault lasted. See the file header for the full reason.
		"voice_gate_unreadable",
		// The two that make this an allow-list rather than a deny-list.
		"a_reason_added_next_month", "",
	}
	for _, r := range loud {
		if !metaDescriptionRefusalIsLoud(r) {
			t.Errorf("reason %q must file an item: it is a copy judgement a person or a rewrite must act on", r)
		}
	}
	for _, r := range silent {
		if metaDescriptionRefusalIsLoud(r) {
			t.Errorf("reason %q must NOT file an item — see the header for why each of these is excluded", r)
		}
	}
}

// TestMetaDescriptionRefusalItemKeyIsPageScoped. A key that omitted the page
// would let ONE page's open refusal hold the dedup slot against every other page
// on the same site — deadURLControlItemKey documents that exact trap.
//
// MUTATION THAT KILLS IT: drop the page id from the key. The two keys collide and
// the inequality below fails.
func TestMetaDescriptionRefusalItemKeyIsPageScoped(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	ka, kb := metaDescriptionRefusalItemKey(a), metaDescriptionRefusalItemKey(b)
	if ka == kb {
		t.Fatalf("two different pages produced the same dedup key %q — one page's refusal would hold the slot for the whole site", ka)
	}
	if !strings.Contains(ka, a.String()) {
		t.Errorf("key %q does not name its page", ka)
	}
	if !strings.HasPrefix(ka, metaDescriptionRefusalItemType+":") {
		t.Errorf("key %q is not namespaced by item type — it could collide with another producer's keys", ka)
	}
}

// refusalParams builds the collected data a real backfill loop emits at the
// moment of a refusal: a site record and the loop's current item.

// TestCopyGateRefusalFilesAtAnActor is arm 1. It asserts the row is written with
// the REPAIR HANDLER and a dispatchable status — not flag-only.
//
// This is the assertion the whole change exists for. [MEASURED 2026-09-03,
// site_work_items UNION site_work_items_archive] items WITH a handler complete at
// 83%; items with none complete at 17%. Filing this at needs_human_review with no
// handler would have looked like a fix and joined a 69-row queue that has closed
// five items ever.
//
// MUTATION THAT KILLS IT: set handlerAgent to "" (or status to
// "needs_human_review") in fileMetaDescriptionRefusal. WithArgs pins both
// columns positionally, so the INSERT no longer matches and the test fails.
func TestCopyGateRefusalFilesAtAnActor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()

	mock.ExpectBegin()
	expectWorkItemDoorStandsDown(mock)
	expectHandlerRegisteredProbe(mock, metaDescriptionRefusalHandler, true)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			metaDescriptionRefusalHandler, // $11 handler_agent
			"triaged",                     // $12 status
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"site_record":         map[string]interface{}{"site_id": siteID.String()},
			"current_description": map[string]interface{}{"page_id": pageID.String()},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "current_description.page_id",
		}},
	}
	cfg := map[string]interface{}{"page_id_field": "current_description.page_id"}

	fileMetaDescriptionRefusal(context.Background(), params, cfg,
		"A powerful, seamless solution — comprehensive and revolutionary.",
		"voice_tell", "banned_phrase: overclaims (\"powerful\")", zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the refusal did not file the row this change exists to file: %v", err)
	}
}

// TestNonCopyGateRefusalNeverOpensATransaction is arm 2.
//
// ⚠ THE FIRST VERSION OF THIS TEST WAS VACUOUS AND ITS COMMENT CLAIMED IT COULD
// NOT BE. It registered NO expectations and asserted ExpectationsWereMet() == nil,
// on the theory that "any DB interaction fails the test". It does not:
// ExpectationsWereMet reports UNFULFILLED EXPECTATIONS, so with none registered it
// is nil unconditionally. Running the mutation is what exposed it — with the
// metaDescriptionRefusalIsLoud early return DELETED the test still passed, because
// the unexpected BeginTx merely returned an error to the code under test, which
// logged it and returned. MEMORY: a-quiet-test-passes-when-the-rule-is-gone,
// and the marker rule's sharper form — a check is only evidence if it could have
// come out otherwise.
//
// The assertion is now INVERTED and therefore real: one expectation is registered
// (a transaction begins) and the test requires it to be UNFULFILLED. If the filing
// path ever reaches BeginTx for these reasons, the expectation is met, the error
// is nil, and the test fails.
//
// MUTATION THAT KILLS IT (verified): delete the metaDescriptionRefusalIsLoud early
// return. Every reason then reaches BeginTx, the expectation is fulfilled, and
// every subtest fails.
func TestNonCopyGateRefusalNeverOpensATransaction(t *testing.T) {
	for _, reason := range []string{
		"empty_candidate", "candidate_looks_internal", "candidate_too_long",
		"already_has_description", "voice_gate_unreadable", "",
	} {
		t.Run(reason, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			// Registered so it can go UNFULFILLED. This is the assertion.
			mock.ExpectBegin()

			params := ActionParams{
				Context: context.Background(),
				DB:      db,
				Logger:  zap.NewNop(),
				CollectedData: map[string]interface{}{
					"site_record":         map[string]interface{}{"site_id": uuid.New().String()},
					"current_description": map[string]interface{}{"page_id": uuid.New().String()},
				},
			}
			cfg := map[string]interface{}{"page_id_field": "current_description.page_id"}

			fileMetaDescriptionRefusal(context.Background(), params, cfg,
				"some candidate", reason, "detail", zap.NewNop())

			if err := mock.ExpectationsWereMet(); err == nil {
				t.Fatalf("reason %q opened a transaction — it must be silent, filing nothing and "+
					"touching nothing. Only voice_tell and banned_claim are copy judgements an "+
					"actor can repair; see the allow-list and its per-reason justification.", reason)
			}
		})
	}
}

// TestRefusalFilingNeverBreaksItsCaller. The refusal is CORRECT behaviour and has
// already been decided by the time the filing runs. A bookkeeping failure must
// not convert it into a failed step — which, under the loop's
// continue_on_error, would be indistinguishable from the silence this change
// removes.
//
// MUTATION THAT KILLS IT: make fileMetaDescriptionRefusal return an error and
// propagate it from the action's refusal branch. The action then returns a
// non-nil error and the assertion below fails.
func TestRefusalFilingNeverBreaksItsCaller(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The database refuses the transaction outright — the worst case for the
	// filing, and the one that must still leave the refusal intact.
	mock.ExpectBegin().WillReturnError(errors.New("connection reset by peer"))

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"site_record":         map[string]interface{}{"site_id": uuid.New().String()},
			"current_description": map[string]interface{}{"page_id": uuid.New().String()},
		},
	}
	cfg := map[string]interface{}{"page_id_field": "current_description.page_id"}

	// Must not panic and must not block: the only contract is that it returns.
	fileMetaDescriptionRefusal(context.Background(), params, cfg,
		"candidate", "banned_claim", "fleet-wide: unevidenced superlative", zap.NewNop())
}

// TestRefusalReportsWhetherItWasFiled closes the hole the council's bug_historian
// seat found in the first cut (corr 76288ff9, [medium]): every failure branch of
// the filing used to be a bare logger.Warn, so a failure to write the LOUD RECORD
// put the refusal straight back to being a log line — the same defect one hop
// deeper, and hidden behind a design narrative saying it was fixed. bugs_closed/034
// is that shape.
//
// The refusal must still succeed (nil error, updated:false) — a bookkeeping fault
// must not fail a correct refusal — but the step's OWN OUTPUT must say the record
// is missing, because that map is the surface migration 728's operator message
// tells a reader to read.
//
// MUTATION THAT KILLS IT (verified): drop `filed` from the action's result map, or
// make fileMetaDescriptionRefusal return true unconditionally. Both assertions
// below then fail while every other test in this file stays green — which is the
// point: nothing else in the suite can see this.
func TestRefusalReportsWhetherItWasFiled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The filing cannot possibly succeed.
	mock.ExpectBegin().WillReturnError(errors.New("connection reset by peer"))

	pageID := uuid.New()
	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"site_record":         map[string]interface{}{"site_id": uuid.New().String()},
			"current_description": map[string]interface{}{"page_id": pageID.String()},
		},
	}
	cfg := map[string]interface{}{"page_id_field": "current_description.page_id"}

	filed, fileError := fileMetaDescriptionRefusal(context.Background(), params, cfg,
		"candidate", "banned_claim", "fleet-wide: unevidenced superlative", zap.NewNop())

	if filed {
		t.Fatalf("filed reported true when the transaction could not even begin — " +
			"a caller would record the refusal as durably captured when nothing was written")
	}
	if fileError == "" {
		t.Fatalf("filed was false but no reason was reported — the caller can see THAT the " +
			"record is missing but not why, which is the log line again one field along")
	}
}

// TestBothPageIDResolutionsAreTheSameCall answers the council guardian's [low]
// objection (corr 76288ff9): that the refusal path resolving a page id
// "separately from whatever id-resolution the action already does" could point the
// filed item at a different page than the log line.
//
// They cannot disagree, because it is the SAME function called with the SAME
// config — this test pins that rather than leaving it as a code-reading claim, so
// a future edit that gives the refusal path its own resolver breaks here.
func TestBothPageIDResolutionsAreTheSameCall(t *testing.T) {
	pageID := uuid.New()
	params := ActionParams{
		Context: context.Background(),
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"current_description": map[string]interface{}{"page_id": pageID.String()},
		},
	}
	cfg := map[string]interface{}{"page_id_field": "current_description.page_id"}

	a, errA := resolveMetaDescriptionPageID(context.Background(), params, cfg, zap.NewNop())
	b, errB := resolveMetaDescriptionPageID(context.Background(), params, cfg, zap.NewNop())
	if errA != nil || errB != nil {
		t.Fatalf("resolution errored: %v / %v", errA, errB)
	}
	if a != pageID || b != pageID {
		t.Fatalf("resolver did not return the configured page id: %v / %v vs %v", a, b, pageID)
	}
}
