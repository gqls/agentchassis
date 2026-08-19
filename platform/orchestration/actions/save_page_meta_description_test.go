// FILE: platform/orchestration/actions/save_page_meta_description_test.go
//
// bugs_open/320 — `pages.meta_description` had no writer that could repair an
// existing page, and the route that looked like one (`content_rewrite` →
// `page-build-handler`) COMPLETED without writing the column. So the property
// under test here is not merely "it writes"; it is that **a refusal is
// distinguishable from a write**, which is the specific failure this action
// exists to end.
//
// Each test is chosen to FAIL against a named wrong implementation rather than
// merely pass against the right one. The mutation that kills each is named in
// its comment — a test that would still pass with the rule deleted proves
// nothing (MEMORY: a-quiet-test-passes-when-the-rule-is-gone).
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// metaDescGuardClause is the WHERE fragment that carries the overwrite policy.
// The tests match on THIS rather than on "UPDATE pages", and that is a
// correction, not a flourish:
//
//   > **The first version of this file matched `regexp.QuoteMeta("UPDATE pages")`
//   > and its comment claimed that deleting the guard clause would kill the test.
//   > It did not.** Running the mutation is what exposed it — with the clause
//   > deleted the action still passed three args, `WithArgs(..., false)` still
//   > matched, and all five tests stayed green. sqlmock does not care that `$3`
//   > became unreferenced, so the policy could have been silently moved out of
//   > the SQL and into nothing at all.
//
// Matching the clause makes the test fail on exactly the edit that would make
// `overwrite_existing` decorative. MEMORY: a-mutation-that-passes-may-have-hit-
// a-guard-in-series, and mutate-the-code-to-prove-the-guard — a green suite is
// not evidence until the mutation has actually been run.
const metaDescGuardClause = `($3::bool OR COALESCE(meta_description, '') = '')`

// TestSavePageMetaDescription_WritesWhenBlank is the happy path, and it asserts
// the SQL actually carries the overwrite policy in its WHERE clause.
//
// MUTATION THAT KILLS IT (verified by running it): delete the guard clause from
// the UPDATE. The ExpectQuery pattern is the clause itself, so the expectation
// no longer matches. Verified in the other direction too — with the clause
// present the suite is green.
func TestSavePageMetaDescription_WritesWhenBlank(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(metaDescGuardClause)).
		WithArgs(pageID, "A short, human description of the page.", false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pageID))

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id":          pageID.String(),
			"meta_description": "A short, human description of the page.",
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "page_id",
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["updated"] != true {
		t.Fatalf("expected updated=true, got %#v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestSavePageMetaDescription_DefaultDoesNotOverwrite proves the opt-in field's
// default is the SAFE side (owner ruling 2026-08-02 §2). The action must pass
// overwrite=false when the caller says nothing.
//
// MUTATION THAT KILLS IT: change the default in
// `GetBoolField(config,"overwrite_existing", false)` to true — WithArgs then
// sees true and the expectation fails. This is the test that would have caught
// a "sensible default" flipped by a later editor.
func TestSavePageMetaDescription_DefaultDoesNotOverwrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	// The row is NOT returned: this models a page that already has a description,
	// so the guarded WHERE matches nothing.
	mock.ExpectQuery(regexp.QuoteMeta(metaDescGuardClause)).
		WithArgs(pageID, "Replacement copy.", false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id":          pageID.String(),
			"meta_description": "Replacement copy.",
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "page_id",
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["updated"] != false {
		t.Fatalf("expected updated=false when the page already has copy, got %#v", res)
	}
	if res["reason"] != "already_has_description" {
		t.Fatalf("a refusal must SAY it was a refusal; got reason=%v", res["reason"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestSavePageMetaDescription_OverwriteOptIn is the other arm: a gate tested only
// on its refusing side is indistinguishable from one that always refuses (a trap
// this estate has paid for — see the bugfix_284 lane's psql guard).
//
// MUTATION THAT KILLS IT: ignore the config key and hardcode false.
func TestSavePageMetaDescription_OverwriteOptIn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(metaDescGuardClause)).
		WithArgs(pageID, "Replacement copy.", true).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pageID))

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id":          pageID.String(),
			"meta_description": "Replacement copy.",
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field":      "page_id",
			"overwrite_existing": true,
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.(map[string]interface{})["updated"] != true {
		t.Fatalf("opt-in overwrite must write; got %#v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestSavePageMetaDescription_RefusesABrief proves the bugs_closed/103 guard is
// actually CALLED, not merely imported. A backfill that republishes build briefs
// as public copy would recreate the exact defect 103 closed — 1,206 characters
// of generator instructions under a Google result.
//
// MUTATION THAT KILLS IT: delete the MetaDescriptionLooksInternal branch. The
// action would then attempt the UPDATE, and sqlmock fails on an unexpected query.
func TestSavePageMetaDescription_RefusesABrief(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// No ExpectQuery at all: any database call is a failure.

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id": uuid.New().String(),
			// A real brief marker from the 103 census.
			"meta_description": "Build an interactive calculator: no fetch calls, no backend.",
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "page_id",
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["updated"] != false || res["reason"] != "candidate_looks_internal" {
		t.Fatalf("a brief must be refused by name; got %#v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestSavePageMetaDescription_EmptyCandidateWritesNothing closes the loop that
// created bugs_open/320 in the first place: a blank must never reach the column.
//
// MUTATION THAT KILLS IT: remove the `candidate == ""` early return — the action
// would issue an UPDATE writing '' and sqlmock fails on the unexpected query.
func TestSavePageMetaDescription_EmptyCandidateWritesNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id":          uuid.New().String(),
			"meta_description": "   ",
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "page_id",
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.(map[string]interface{})["reason"] != "empty_candidate" {
		t.Fatalf("expected empty_candidate refusal, got %#v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// ── THE COPY GATES (owner requirement 2026-08-19) ───────────────────────────
//
// The owner waived the read-first review pass on condition that the summaries
// "go through the copy guidance and checks so they don't sound like AI". The
// checks REPLACE the review, so a test suite that never exercises them would be
// certifying the wrong thing entirely.

// TestSavePageMetaDescription_VoiceGateRefusesABannedPhrase proves the site's own
// voice rules are actually CONSULTED and are BLOCKING — not loaded and ignored.
//
// MUTATION THAT KILLS IT (verified by running it): delete the
// metaDescriptionFailsCopyGates call from the action. The UPDATE then fires and
// sqlmock reports an unexpected query, because no ExpectQuery("UPDATE") is set.
func TestSavePageMetaDescription_VoiceGateRefusesABannedPhrase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	// A voice gate banning the single most AI-sounding construction there is.
	mock.ExpectQuery(regexp.QuoteMeta("aspect = 'voice'")).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(
			`{"voice_gate":{"enabled":true,"banned_phrases":[{"pattern":"in today's fast-paced world","reason":"AI boilerplate opener"}]}}`)))
	// No UPDATE expectation: any write is a failure.

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id":          uuid.New().String(),
			"site_record":      map[string]interface{}{"site_id": siteID.String()},
			"meta_description": "In today's fast-paced world, our solutions empower your business.",
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "page_id",
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["updated"] != false || res["reason"] != "voice_tell" {
		t.Fatalf("a banned phrase must be refused as voice_tell; got %#v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestSavePageMetaDescription_CleanCopyPassesTheGates is the other arm, and it is
// the one that matters most: a gate tested only on its REFUSING side is
// indistinguishable from one that refuses everything, which here would mean the
// whole backfill silently writes nothing while reporting named refusals.
//
// MUTATION THAT KILLS IT: make metaDescriptionFailsCopyGates return a constant
// non-empty reason.
func TestSavePageMetaDescription_CleanCopyPassesTheGates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	clean := "Work out what a loan really costs before you apply, with the rate and term you were actually offered."

	mock.ExpectQuery(regexp.QuoteMeta("aspect = 'voice'")).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(
			`{"voice_gate":{"enabled":true,"banned_phrases":[{"pattern":"in today's fast-paced world","reason":"AI boilerplate opener"}]}}`)))
	mock.ExpectQuery(regexp.QuoteMeta("aspect = 'evidence_base'")).
		WithArgs(siteID).
		WillReturnError(sql.ErrNoRows) // site has no register — fleet-wide arm still runs
	mock.ExpectQuery(regexp.QuoteMeta(metaDescGuardClause)).
		WithArgs(pageID, clean, false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pageID))

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id":          pageID.String(),
			"site_record":      map[string]interface{}{"site_id": siteID.String()},
			"meta_description": clean,
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "page_id",
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.(map[string]interface{})["updated"] != true {
		t.Fatalf("clean copy must pass the gates and be written; got %#v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestSavePageMetaDescription_UnreadableGateIsNotAPass pins the failure direction.
// If the voice spec query errors, we must NOT publish — "I could not check" is
// not "it is fine". This is the shape the estate keeps paying for elsewhere: a
// swallowed error that reads as a clean result.
//
// MUTATION THAT KILLS IT: return ("","") instead of ("voice_gate_unreadable",…)
// on the load error — the action would then write and sqlmock would see an
// unexpected UPDATE.
func TestSavePageMetaDescription_UnreadableGateIsNotAPass(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("aspect = 'voice'")).
		WithArgs(siteID).
		WillReturnError(fmt.Errorf("connection reset"))

	params := ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"page_id":          uuid.New().String(),
			"site_record":      map[string]interface{}{"site_id": siteID.String()},
			"meta_description": "A perfectly reasonable sentence about the page.",
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"page_id_field": "page_id",
		}},
	}

	out, err := SavePageMetaDescriptionAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["updated"] != false || res["reason"] != "voice_gate_unreadable" {
		t.Fatalf("an unreadable gate must refuse, not pass; got %#v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
