// FILE: platform/orchestration/actions/stored_slot_rescue_test.go
//
// bugs_closed/204 — validate_plan (and two apply_gap_plan arms) DELETED a section
// name that is a positional slot the page already carries. These pin the fix and,
// just as importantly, the four things it must NOT do:
//
//   1. it must not rescue a name on a page that does not carry it — the scoping
//      property that makes this different from the resolver widening LANDMINES
//      forbids. Without this test that difference is only a claim in a comment;
//   2. it must not fire when the catalogue resolves the name, and must issue no
//      query at all in that case — so the arm is inert on the undecomposed estate
//      by construction rather than by configuration;
//   3. it must not stop dropping names that really are junk — the checker's actual
//      job, which 140 of 140 recorded drops were never about;
//   4. it must not rewrite what it keeps: object-form entries keep their RFC_016
//      facts, and a positional name is NOT collapsed onto its component function.

package actions

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// validateSlotParams builds a validate_plan run over one decomposed page whose
// sections are positional slot names. siteID travels the way the live workflow
// supplies it, at site_record.site_id.
func validateSlotParams(siteID string, sections []interface{}) ActionParams {
	return ActionParams{
		Context: context.Background(),
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID},
			"llm_plan": map[string]interface{}{
				"pages": []interface{}{
					map[string]interface{}{
						"name": "guide-how-loans-are-calculated", "page_type": "guide",
						"sections": sections,
					},
				},
			},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"plan_field":          "llm_plan",
			"validate_components": true,
		}},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "66666666-6666-6666-6666-666666666666",
			StepName:        "validate_plan",
		},
	}
}

// slotIdentityRows is what LoadPageSlotRowsForSite returns: page, slot, component.
func slotIdentityRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"page_name", "slot_name", "component_id"})
}

// THE BUG, INVERTED. Pre-fix this run produced sections=[] for the page and two
// PLAN_SECTION_NAME_DROPPED rows; on 2026-08-20 the same shape emptied 41 of 45
// live pages on one site.
func TestValidatePlan_PositionalSlotNamesAreKeptNotDropped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	componentA := uuid.New().String()

	expectResolverQueries(mock)
	// The rescue's ONE read, on the first miss.
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "").
		WillReturnRows(slotIdentityRows().
			AddRow("guide-how-loans-are-calculated", "prose-0", componentA).
			AddRow("guide-how-loans-are-calculated", "prose-1", componentA))

	params := validateSlotParams(siteID.String(), []interface{}{"prose-0", "prose-1"})
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := sectionsOfFirstPage(t, out)
	want := []string{"prose-0", "prose-1"}
	if len(got) != len(want) {
		t.Fatalf("sections = %v, want %v — the pre-fix result was an empty list", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			// Naming the failure precisely: rewriting prose-0 to its component's
			// function would collapse prose-0 and prose-1 onto one name, which is
			// what the positional naming exists to prevent.
			t.Errorf("section %d = %q, want %q kept VERBATIM", i, got[i], want[i])
		}
	}
}

// The regression guard for the checker's real job. A name that is neither a
// component nor a stored slot is still junk and must still go.
func TestValidatePlan_UnstoredUnresolvableNameIsStillDropped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	expectResolverQueries(mock)
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "").
		WillReturnRows(slotIdentityRows().
			AddRow("guide-how-loans-are-calculated", "prose-0", uuid.New().String()))

	params := validateSlotParams(siteID.String(), []interface{}{"prose-0", "prose-99"})
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := sectionsOfFirstPage(t, out)
	if len(got) != 1 || got[0] != "prose-0" {
		t.Fatalf("sections = %v, want exactly [prose-0]: prose-99 is on no page and must still be dropped", got)
	}
}

// THE SCOPING PROPERTY. This is the test that makes "this is not a resolver
// widening" a fact about the code rather than a claim in a comment: a slot stored
// on ANOTHER page grants nothing here.
func TestStoredSlotRescue_IsScopedToTheProposedPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	expectResolverQueries(mock)
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "").
		WillReturnRows(slotIdentityRows().
			// prose-0 exists on the SITE, but on a different page.
			AddRow("some-other-page", "prose-0", uuid.New().String()))

	params := validateSlotParams(siteID.String(), []interface{}{"prose-0"})
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sectionsOfFirstPage(t, out); len(got) != 0 {
		t.Fatalf("sections = %v, want empty: a slot stored on another page must NOT rescue this one — "+
			"a site-scoped keep would be exactly the resolver widening PLAN-050's landmine forbids", got)
	}
}

// INERT ON THE UNDECOMPOSED ESTATE. A plan whose names all resolve must issue no
// slot query at all. sqlmock's ExpectationsWereMet is what turns "we think it is
// lazy" into an assertion — an unexpected query fails the run.
func TestStoredSlotRescue_IssuesNoQueryWhenEveryNameResolves(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	expectResolverQueries(mock) // and deliberately NO page_components expectation

	params := validateSlotParams(siteID.String(), []interface{}{"hero", "faq"})
	params.DB = db

	if _, err := ValidateSitePlanAction(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet/unexpected expectations: %v", err)
	}
}

// FAILING TOWARD KEEPING. A transient read failure must not be able to do what
// the bug did. The disconfirming result is an empty section list.
func TestStoredSlotRescue_ReadFailureKeepsRatherThanDrops(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	expectResolverQueries(mock)
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "").
		WillReturnError(errors.New("connection reset"))

	params := validateSlotParams(siteID.String(), []interface{}{"prose-0"})
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("a slot read failure must not fail the plan: %v", err)
	}
	if got := sectionsOfFirstPage(t, out); len(got) != 1 || got[0] != "prose-0" {
		t.Fatalf("sections = %v, want [prose-0] kept: a transient read failure must not empty a decomposed page", got)
	}
}

// The keep must not silently swallow the RFC_016 fact assignments travelling
// inside an object-form entry — the drop happens BEFORE the object→string split,
// which is why those facts died with the entry pre-fix.
func TestValidatePlan_KeptObjectEntryRetainsItsFacts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	expectResolverQueries(mock)
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "").
		WillReturnRows(slotIdentityRows().
			AddRow("guide-how-loans-are-calculated", "prose-0", uuid.New().String()))

	params := validateSlotParams(siteID.String(), []interface{}{
		map[string]interface{}{
			"name":  "prose-0",
			"facts": []interface{}{"sdlt-standard-nil-band-upper", "boe-base-rate"},
		},
	})
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	plan := out.(map[string]interface{})
	page := plan["pages"].([]interface{})[0].(map[string]interface{})

	if got := sectionsOfFirstPage(t, out); len(got) != 1 || got[0] != "prose-0" {
		t.Fatalf("sections = %v, want [prose-0]", got)
	}
	facts, ok := page["section_facts"].([]interface{})
	if !ok || len(facts) != 1 {
		t.Fatalf("section_facts = %v (%T), want one aligned entry — the object→string split must have run on the KEPT entry", page["section_facts"], page["section_facts"])
	}
	assigned, ok := facts[0].([]interface{})
	if !ok || len(assigned) != 2 {
		t.Fatalf("fact assignment = %v, want the two ids the planner attached", facts[0])
	}
}

// ── the unit-level judgement ────────────────────────────────────────────────

func TestStoredSlotRescue_NilRescueIsTodaysBehaviour(t *testing.T) {
	// A run with no db or no site identity must degrade to dropping exactly as
	// before, and must not panic on a nil receiver.
	var r *storedSlotRescue
	if v := r.verdict(context.Background(), "index", "prose-0"); v != slotNotStored {
		t.Errorf("nil rescue verdict = %v, want slotNotStored", v)
	}
	if r.keptCount() != 0 || r.readFailed() {
		t.Errorf("nil rescue must report nothing kept and no failure")
	}
	if f := r.keptFinding(); f != nil {
		t.Errorf("nil rescue must file no finding, got %v", f)
	}
	if got := storedSlotRescueFor(nil, "not-a-uuid", zap.NewNop()); got != nil {
		t.Errorf("a malformed site id must disable the rescue, not enable it with uuid.Nil")
	}
}

func TestStoredSlotRescue_CleanRunFilesNoKeptFinding(t *testing.T) {
	// A negative a mock cannot assert: keptFinding returns nil when nothing was
	// rescued, so a site with no decomposed pages adds no rows at all.
	r := &storedSlotRescue{keptPages: map[string][]string{}}
	if f := r.keptFinding(); f != nil {
		t.Errorf("a run that rescued nothing must file no finding, got %v", f)
	}
}

func TestStoredSlotRescue_ReadFailureIsLoudAndSaysWhatItDid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	mock.ExpectQuery("FROM page_components pc").WillReturnError(errors.New("boom"))

	core, recorded := observer.New(zapcore.WarnLevel)
	r := newStoredSlotRescue(db, siteID, zap.New(core))
	if v := r.verdict(context.Background(), "index", "prose-0"); v != slotUnknown {
		t.Fatalf("verdict on a failed read = %v, want slotUnknown", v)
	}
	if !r.readFailed() {
		t.Error("readFailed must distinguish 'kept because unreadable' from 'kept because recognised'")
	}
	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly one warning, got %d", len(entries))
	}
	// The line must say what it DID, not merely that something failed — a reader
	// meeting an unexpected name in a plan needs this to explain it.
	if !strings.Contains(entries[0].Message, "KEPT rather than dropped") {
		t.Errorf("the warning does not say what the failure caused: %q", entries[0].Message)
	}
}

// ── the council's medium objection, answered in code ────────────────────────
//
// Council corr f73f4eeb (APPROVED, bug_historian, medium): "slotUnknown collapses
// 'DB read failed' into the same keep-path as 'legitimately stored' ... it silently
// absorbs an infrastructure fault into an apparently-clean validation pass."
//
// The logs did already distinguish the two, but the objection landed where it
// mattered most and I had missed it: the DURABLE record did not. A run that kept
// every name because the database was unreachable filed NO row at all, and so read
// exactly like a clean pass — the silent-absorb shape this lane exists to remove,
// reproduced one level up. These pin the fix.

func TestStoredSlotRescue_ReadFailureFilesItsOwnDurableRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM page_components pc").WillReturnError(errors.New("connection reset"))

	r := newStoredSlotRescue(db, uuid.New(), zap.NewNop())
	for _, name := range []string{"prose-0", "prose-1", "tool-1"} {
		if v := r.verdict(context.Background(), "guide", name); v != slotUnknown {
			t.Fatalf("verdict for %q = %v, want slotUnknown", name, v)
		}
	}

	f := r.keptFinding()
	if len(f) != 1 {
		t.Fatalf("a failed read must file exactly one finding, got %d", len(f))
	}
	if f[0].ErrorCode != "PLAN_SECTION_STORED_SLOT_READ_FAILED" {
		t.Errorf("error code = %q; it must NOT share PLAN_SECTION_NAME_KEPT_BY_STORED_SLOT, "+
			"or a query for 'the rescue worked' silently counts runs where it did not", f[0].ErrorCode)
	}
	if got := f[0].Context["kept_without_checking"]; got != 3 {
		t.Errorf("kept_without_checking = %v, want 3 — the count is what says how much of this plan is unverified", got)
	}
	if r.keptCount() != 0 {
		t.Errorf("keptCount = %d, want 0: nothing was RESCUED, it was kept unchecked", r.keptCount())
	}
}

func TestStoredSlotRescue_CleanRunFilesOnlyTheRescueRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(
		slotIdentityRows().AddRow("guide", "prose-0", uuid.New().String()))

	r := newStoredSlotRescue(db, uuid.New(), zap.NewNop())
	if v := r.verdict(context.Background(), "guide", "prose-0"); v != slotStored {
		t.Fatalf("verdict = %v, want slotStored", v)
	}

	f := r.keptFinding()
	if len(f) != 1 || f[0].ErrorCode != "PLAN_SECTION_NAME_KEPT_BY_STORED_SLOT" {
		t.Fatalf("a successful run must file exactly the rescue row, got %+v", f)
	}
	// The disconfirming direction: a read-failure row on a run where the read
	// succeeded would make the failure signal useless by crying wolf.
	for _, x := range f {
		if x.ErrorCode == "PLAN_SECTION_STORED_SLOT_READ_FAILED" {
			t.Error("a successful read must not file a read-failure row")
		}
	}
}
