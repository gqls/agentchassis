// FILE: platform/orchestration/actions/discovery_checks/check_section_source_drift_test.go
//
// ── THE MUTATION TABLE (bugs_open/469) ──────────────────────────────────────
//
// A mock's own bookkeeping cannot assert a negative: sqlmock will happily agree
// that a guard "worked" when the guard was never reached. So every safety
// property below was proven by MUTATING the source and watching a named test
// fail — and then restoring it. Re-run this table when you change the check;
// heed the standing warning that a mutation which PASSES usually hit a second
// guard in series rather than proving the first one redundant.
//
//	mutation applied to check_section_source_drift.go        test that must FAIL
//	---------------------------------------------------------------------------
//	sectionMultisetDiff -> set difference (skip the budget)  TestSectionMultisetDiff
//	drop `f.Receipt = &rc` in retraction()                   TestAuthorityWonCarriesAReceipt
//	subtract the RAW cache instead of mergedToday in `lost`  TestLockedRowIsNeverReportedLost
//	retract when the page is absent from cacheSections       TestVanishedPageIsLeftOpen
//	accept a spec whose item_key is not this check's         TestForeignShapeIsDisclaimed
//	drop the authorityReadProductive demand control          TestNoAuthorityTodayNeedsADemandControl
//	make appendRetractions inert on the zero-finding site    TestCloserRunsWhenNothingIsFlagged
//	  (the semantic of placing it after the loop's early return)
//	severity hard-coded to "medium" (or to "high")           TestSeverityFollowsWouldDropPresent
//	summary stops naming would_drop                          TestSummaryNamesWhatTheBuildWillDestroy
//	add a second ResolvedFinding{ literal to the file        TestOneRetractionConstructor
//	change a constant without its field-site string          TestItemTypeConstantsMatchTheStrings
//
// All eleven were applied by script and observed to FAIL on 2026-09-03, each
// with the source restored and diffed byte-identical afterwards, before this
// file was considered done. Nothing here is a guard nobody watched fire.

package discovery_checks

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestOrderedListsEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"identical", []string{"hero", "cta"}, []string{"hero", "cta"}, true},
		{"different length", []string{"hero"}, []string{"hero", "cta"}, false},
		{"different order (layout matters)", []string{"hero", "cta"}, []string{"cta", "hero"}, false},
		{"different content", []string{"hero", "specs"}, []string{"hero", "cta"}, false},
		{"both empty", []string{}, []string{}, true},
		// The exact robot-hands product-detail drift that motivated this check:
		// authoritative table still held the old e-commerce layout while
		// pages.sections had been swapped to the spec sheet.
		{
			"product-detail drift",
			[]string{"product-hero", "product-specs", "call-to-action"},
			[]string{"gripper-spec-sheet", "call-to-action"},
			false,
		},
	}
	for _, tc := range cases {
		if got := orderedListsEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: orderedListsEqual(%v,%v)=%v want %v", tc.name, tc.a, tc.b, got, tc.want)
		}
	}
}

// TestSectionMultisetDiff pins the property that makes the loss test correct.
//
// A SET difference passes every case here except the repeated one — and the
// repeated one is not hypothetical: a live site's plan names generic-text-block
// six times (measured 2026-09-03). Under a set difference, dropping one of six
// reports "nothing lost", which is the under-report this whole bug is about.
func TestSectionMultisetDiff(t *testing.T) {
	cases := []struct {
		name       string
		have, want []string
		expect     []string
	}{
		{"nothing dropped", []string{"a", "b"}, []string{"a", "b"}, []string{}},
		{"one dropped", []string{"a", "b", "c"}, []string{"a", "c"}, []string{"b"}},
		{"all dropped", []string{"a"}, []string{}, []string{"a"}},
		{"order preserved", []string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},
		// THE ONE A SET DIFFERENCE GETS WRONG.
		{"repeated: six down to five", []string{"g", "g", "g", "g", "g", "g"}, []string{"g", "g", "g", "g", "g"}, []string{"g"}},
		{"repeated: two down to one", []string{"g", "x", "g"}, []string{"g", "x"}, []string{"g"}},
		{"repeated: none dropped", []string{"g", "g"}, []string{"g", "g"}, []string{}},
		// A pure reorder loses nothing — the multiset is unchanged.
		{"reorder loses nothing", []string{"a", "b"}, []string{"b", "a"}, []string{}},
	}
	for _, tc := range cases {
		got := sectionMultisetDiff(tc.have, tc.want)
		if !orderedListsEqual(got, tc.expect) {
			t.Errorf("%s: sectionMultisetDiff(%v,%v)=%v want %v", tc.name, tc.have, tc.want, got, tc.expect)
		}
	}
}

// TestClassifyDriftResolution walks the real 2026-09-03 backlog, because the
// cases that matter are the ones where `direction` and `lost` DISAGREE.
func TestClassifyDriftResolution(t *testing.T) {
	cases := []struct {
		name          string
		frozenAuth    []string
		frozenCache   []string
		rawToday      []string
		mergedToday   []string
		wantDirection string
		wantLost      []string
		wantGained    []string
		wantReordered bool
	}{
		{
			// robot-hands.com/gripper-catalog — REAL DAMAGE.
			name:          "authority deleted a section the cache held",
			frozenAuth:    []string{"hero", "generic-text-block", "info-card-grid", "call-to-action"},
			frozenCache:   []string{"hero", "generic-text-block", "gripper-spec-sheet", "info-card-grid", "call-to-action"},
			rawToday:      []string{"hero", "generic-text-block", "info-card-grid", "call-to-action"},
			mergedToday:   []string{"hero", "generic-text-block", "info-card-grid", "call-to-action"},
			wantDirection: "authority_won",
			wantLost:      []string{"gripper-spec-sheet"},
			wantGained:    []string{},
		},
		{
			// oufe.com/contact — SAME DIRECTION, NOTHING LOST. This is the case
			// that proves `direction` cannot be the predicate: the cache had
			// dropped a section and the authority put it back.
			name:          "authority RESTORED a section the cache had dropped",
			frozenAuth:    []string{"hero-contact", "contact-info", "contact-form"},
			frozenCache:   []string{"hero-contact", "contact-form"},
			rawToday:      []string{"hero-contact", "contact-info", "contact-form"},
			mergedToday:   []string{"hero-contact", "contact-info", "contact-form"},
			wantDirection: "authority_won",
			wantLost:      []string{},
			wantGained:    []string{"contact-info"},
		},
		{
			// idea.uk/guides-index — a swap. Lost AND gained; the owning lane
			// later judged it a benign rename. The receipt is what let them.
			name:          "a swap files a receipt and records what replaced it",
			frozenAuth:    []string{"hero", "content-listing"},
			frozenCache:   []string{"hero", "guide-list"},
			rawToday:      []string{"hero", "content-listing"},
			mergedToday:   []string{"hero", "content-listing"},
			wantDirection: "authority_won",
			wantLost:      []string{"guide-list"},
			wantGained:    []string{"content-listing"},
		},
		{
			// boxingonline.com/tool-fight-calendar after migration 750 — the
			// benign close this whole mechanism exists to allow.
			name:          "the plan was corrected to match the cache",
			frozenAuth:    []string{"hero-tool", "generic-text-block", "advertising"},
			frozenCache:   []string{"hero-tool", "event-list"},
			rawToday:      []string{"hero-tool", "event-list"},
			mergedToday:   []string{"hero-tool", "event-list"},
			wantDirection: "cache_held",
			wantLost:      []string{},
			wantGained:    []string{},
		},
		{
			name:          "neither list: a third state",
			frozenAuth:    []string{"hero", "a"},
			frozenCache:   []string{"hero", "b"},
			rawToday:      []string{"hero", "c"},
			mergedToday:   []string{"hero", "c"},
			wantDirection: "third_list",
			wantLost:      []string{"b"},
			wantGained:    []string{"c"},
		},
		{
			name:          "a pure reorder loses nothing but is recorded",
			frozenAuth:    []string{"a", "b"},
			frozenCache:   []string{"b", "a"},
			rawToday:      []string{"a", "b"},
			mergedToday:   []string{"a", "b"},
			wantDirection: "authority_won",
			wantLost:      []string{},
			wantGained:    []string{},
			wantReordered: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyDriftResolution(driftItemSnapshot{
				PageName:      "p",
				Authoritative: tc.frozenAuth,
				PagesSections: tc.frozenCache,
			}, nil, tc.rawToday, tc.mergedToday, "site_plan_sections")

			if got.direction != tc.wantDirection {
				t.Errorf("direction = %q, want %q", got.direction, tc.wantDirection)
			}
			if !orderedListsEqual(got.lost, tc.wantLost) {
				t.Errorf("lost = %v, want %v", got.lost, tc.wantLost)
			}
			if !orderedListsEqual(got.gained, tc.wantGained) {
				t.Errorf("gained = %v, want %v", got.gained, tc.wantGained)
			}
			if got.reordered != tc.wantReordered {
				t.Errorf("reordered = %v, want %v", got.reordered, tc.wantReordered)
			}
		})
	}
}

// TestAuthorityWonCarriesAReceipt is the safety property in one assertion: a
// retraction of a LOSSY resolution must carry the record of what was lost.
func TestAuthorityWonCarriesAReceipt(t *testing.T) {
	res := classifyDriftResolution(driftItemSnapshot{
		ItemKey:       "section_source_drift:gripper-catalog",
		ItemID:        "item-1",
		PageName:      "gripper-catalog",
		AuthSource:    "site_plan_sections",
		Authoritative: []string{"hero", "info-card-grid"},
		PagesSections: []string{"hero", "gripper-spec-sheet", "info-card-grid"},
	}, nil, []string{"hero", "info-card-grid"}, []string{"hero", "info-card-grid"}, "site_plan_sections")

	f := res.retraction(DiscoveryCheckContext{Logger: zap.NewNop()}, map[string]int{"gripper-spec-sheet": 24})

	if f.Receipt == nil {
		t.Fatal("a resolution that DESTROYED a section retracted with no receipt — this is the bug")
	}
	if f.Receipt.ItemType != "section_composition_lost" {
		t.Errorf("receipt item type = %q", f.Receipt.ItemType)
	}
	if f.Receipt.Severity != "high" {
		t.Errorf("severity = %q, want high — 24 archived rows is evidence it had rendered", f.Receipt.Severity)
	}
	if !f.Receipt.RecurrenceExpected {
		t.Error("a second loss on the same page is a second EVENT; without RecurrenceExpected the anti-churn brake drops the re-file")
	}
	if f.Receipt.HandlerAgent != "" {
		t.Errorf("handler = %q, want none — a machine cannot tell an intended removal from this bug's completion", f.Receipt.HandlerAgent)
	}
	// The evidence must be COPIED, not pointed at: the drift row is closed in
	// the same transaction and site_work_items is a rolling window.
	for _, key := range []string{"pages_sections_at_filing", "authoritative_at_filing", "lost_sections", "agreed_list_today"} {
		if !strings.Contains(f.Receipt.SpecJSON, key) {
			t.Errorf("receipt spec is missing %q — it must copy its evidence, not point at a row that will be archived", key)
		}
	}
	if !strings.Contains(f.Receipt.SpecJSON, "gripper-spec-sheet") {
		t.Error("the receipt does not name what was lost")
	}
	if f.Evidence["direction"] != "authority_won" {
		t.Errorf("evidence direction = %v", f.Evidence["direction"])
	}
}

// TestCacheHeldRetractsWithoutReceipt: the benign close must stay cheap, or the
// mechanism becomes a receipt generator and the receipts stop meaning anything.
func TestCacheHeldRetractsWithoutReceipt(t *testing.T) {
	res := classifyDriftResolution(driftItemSnapshot{
		ItemKey:       "section_source_drift:tool-fight-calendar",
		PageName:      "tool-fight-calendar",
		Authoritative: []string{"hero-tool", "generic-text-block", "advertising"},
		PagesSections: []string{"hero-tool", "event-list"},
	}, nil, []string{"hero-tool", "event-list"}, []string{"hero-tool", "event-list"}, "site_plan_sections")

	f := res.retraction(DiscoveryCheckContext{Logger: zap.NewNop()}, nil)
	if f.Receipt != nil {
		t.Fatalf("the cache's own list won and nothing was lost, yet a loss receipt was filed: %s", f.Receipt.Summary)
	}
	if f.Evidence["direction"] != "cache_held" {
		t.Errorf("direction = %v, want cache_held", f.Evidence["direction"])
	}
	if f.ItemKey != "section_source_drift:tool-fight-calendar" {
		t.Errorf("item key = %q", f.ItemKey)
	}
}

// TestLockedRowIsNeverReportedLost. A human-locked row the plan does not name
// sits in the frozen cache and NOT in the frozen authority, so it lands in
// `cacheOnly`. It is still on the page. Subtracting the RAW cache instead of the
// MERGED list would report it lost — and would file a false destruction receipt
// on every locked page on the estate.
func TestLockedRowIsNeverReportedLost(t *testing.T) {
	res := classifyDriftResolution(driftItemSnapshot{
		PageName:      "contact",
		Authoritative: []string{"hero", "contact-form"},
		PagesSections: []string{"hero", "contact-form", "chat-input-box"}, // locked row in the cache
	}, nil,
		[]string{"hero", "contact-form"},                   // raw cache today: the plan's list
		[]string{"hero", "contact-form", "chat-input-box"}, // merged: the lock is still on the page
		"site_plan_sections")

	if len(res.lost) != 0 {
		t.Fatalf("lost = %v — a locked row that is STILL ON THE PAGE was reported destroyed", res.lost)
	}
	f := res.retraction(DiscoveryCheckContext{Logger: zap.NewNop()}, nil)
	if f.Receipt != nil {
		t.Fatal("a false loss receipt was filed for a locked row that never went anywhere")
	}
}

// ─── Run-level tests ────────────────────────────────────────────────────────

type driftMockPlan struct {
	planRows   [][2]string       // page_name, component_name
	aspectJSON string            // "" => ErrNoRows
	cache      map[string]string // page name -> sections JSON
	locked     [][]driver.Value  // rows for LockedPageSlotsSQL
	slots      [][2]string       // page name, slot name (live page_components)
	openItems  [][4]string       // id, item_key, created_at, spec JSON
}

// expectDriftQueries wires the six reads Run performs, in order. The two
// page_components reads are disambiguated on their SELECT lists, not on their
// shared FROM clause — matching "FROM page_components pc" alone binds the first
// expectation to whichever runs first and silently mis-pairs them.
func expectDriftQueries(mock sqlmock.Sqlmock, siteID uuid.UUID, p driftMockPlan) {
	planRows := sqlmock.NewRows([]string{"page_name", "component_name"})
	for _, r := range p.planRows {
		planRows.AddRow(r[0], r[1])
	}
	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID).WillReturnRows(planRows)

	if p.aspectJSON == "" {
		mock.ExpectQuery("SELECT data FROM site_specs").WithArgs(siteID).WillReturnError(sql.ErrNoRows)
	} else {
		mock.ExpectQuery("SELECT data FROM site_specs").WithArgs(siteID).
			WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(p.aspectJSON)))
	}

	cacheRows := sqlmock.NewRows([]string{"id", "name", "sections"})
	for name, secs := range p.cache {
		cacheRows.AddRow(uuid.New(), name, secs)
	}
	mock.ExpectQuery("FROM pages").WithArgs(siteID).WillReturnRows(cacheRows)

	lockedRows := sqlmock.NewRows([]string{"id", "name", "slot_name", "position", "component_id", "function", "cname", "lock_type", "locked_by"})
	for _, r := range p.locked {
		lockedRows.AddRow(r...)
	}
	mock.ExpectQuery(`COALESCE\(pc.lock_type`).WithArgs(siteID, "").WillReturnRows(lockedRows)

	slotRows := sqlmock.NewRows([]string{"name", "slot_name", "component_id"})
	for _, s := range p.slots {
		slotRows.AddRow(s[0], s[1], "")
	}
	mock.ExpectQuery(`SELECT p.name, COALESCE\(pc.slot_name`).WithArgs(siteID, "").WillReturnRows(slotRows)

	itemRows := sqlmock.NewRows([]string{"id", "item_key", "created_at", "spec"})
	for _, it := range p.openItems {
		itemRows.AddRow(it[0], it[1], it[2], it[3])
	}
	mock.ExpectQuery("FROM site_work_items").WithArgs(siteID, "section_source_drift").WillReturnRows(itemRows)
}

func runDriftCheck(t *testing.T, p driftMockPlan) (*CheckResult, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	siteID := uuid.New()
	expectDriftQueries(mock, siteID, p)
	res, err := (&SectionSourceDriftCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		db.Close()
		t.Fatalf("run: %v", err)
	}
	return res, mock, func() { db.Close() }
}

// bugs_open/285 — the loader merges human-locked live rows into the list and
// into pages.sections, so a page whose plan omits a locked section is NOT
// drifted, whether its cache pre-dates the fix (raw plan) or post-dates it
// (plan + locked). A genuine one-store edit still is.
func TestSectionSourceDrift_LockedLiveRowIsNotDrift(t *testing.T) {
	base := driftMockPlan{
		planRows: [][2]string{
			{"contact", "hero"}, {"contact", "contact-info"},
			{"about", "hero-about"}, {"about", "faq"},
		},
		// contact carries a locked chat box at position 3 the plan does not name.
		locked: [][]driver.Value{{"r1", "contact", "chat-input-box", 3, "cid", "chat-input-box", "chat-input-box", "permanent", "lane"}},
	}
	run := func(t *testing.T, cache map[string]string, wantFindings int) {
		t.Helper()
		p := base
		p.cache = cache
		res, mock, done := runDriftCheck(t, p)
		defer done()
		if got := len(res.Findings); got != wantFindings {
			t.Errorf("findings = %d (%v), want %d", got, res.Findings, wantFindings)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	}
	t.Run("cache pre-dates the fix (raw plan) → not drift", func(t *testing.T) {
		run(t, map[string]string{"contact": `["hero","contact-info"]`, "about": `["hero-about","faq"]`}, 0)
	})
	t.Run("cache written by the fixed loader (plan + locked) → not drift", func(t *testing.T) {
		run(t, map[string]string{"contact": `["hero","contact-info","chat-input-box"]`, "about": `["hero-about","faq"]`}, 0)
	})
	t.Run("a genuine one-store edit still flags", func(t *testing.T) {
		run(t, map[string]string{"contact": `["hero","contact-info","chat-input-box"]`, "about": `["gripper-spec-sheet","faq"]`}, 1)
	})
}

// TestCloserRunsWhenNothingIsFlagged is the landmine test. The site with ZERO
// findings is the ONLY site an early return fires on — and it is precisely the
// site whose stale items need closing. Moving appendRetractions below the filing
// loop, or behind the per-pass cap, makes the closer inert exactly here.
func TestCloserRunsWhenNothingIsFlagged(t *testing.T) {
	res, mock, done := runDriftCheck(t, driftMockPlan{
		planRows: [][2]string{{"about", "hero"}, {"about", "faq"}},
		cache:    map[string]string{"about": `["hero","faq"]`}, // AGREES: zero findings
		slots:    [][2]string{{"about", "hero"}, {"about", "faq"}},
		openItems: [][4]string{{
			"item-1", "section_source_drift:about", "2026-07-28T00:00:00Z",
			`{"page_name":"about","authoritative_source":"site_plan_sections",` +
				`"authoritative":["hero","faq"],"pages_sections":["hero","spec-sheet","faq"]}`,
		}},
	})
	defer done()

	if len(res.Findings) != 0 {
		t.Fatalf("expected zero findings on an agreeing site, got %d", len(res.Findings))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("the closer did not run on a site with nothing to file — Resolved = %d", len(res.Resolved))
	}
	if res.Resolved[0].Receipt == nil {
		t.Fatal("spec-sheet was in the frozen cache, is not on the page today, and no receipt was filed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// TestVanishedPageIsLeftOpen — absence is not an observation. A page row that
// has gone (deleted, or its sections emptied) tells us nothing about whether the
// divergence resolved, so the item stays open.
func TestVanishedPageIsLeftOpen(t *testing.T) {
	res, _, done := runDriftCheck(t, driftMockPlan{
		planRows: [][2]string{{"about", "hero"}},
		cache:    map[string]string{"about": `["hero"]`}, // the ITEM's page is not here
		slots:    [][2]string{{"about", "hero"}},
		openItems: [][4]string{{
			"item-1", "section_source_drift:gone-page", "2026-07-28T00:00:00Z",
			`{"page_name":"gone-page","authoritative":["hero"],"pages_sections":["hero","x"]}`,
		}},
	})
	defer done()
	if len(res.Resolved) != 0 {
		t.Fatalf("retracted an item about a page that was never read: %+v", res.Resolved)
	}
}

// TestForeignShapeIsDisclaimed — item_type is not a predicate (the GradesFunc
// principle, bugs_open/213). A row filed under this type by something else, or
// with a malformed spec, is left alone rather than judged by a predicate that
// does not describe it.
func TestForeignShapeIsDisclaimed(t *testing.T) {
	cases := []struct {
		name string
		item [4]string
	}{
		{"key does not follow this check's contract", [4]string{
			"i1", "some_other_producer:about", "t",
			`{"page_name":"about","authoritative":["hero"],"pages_sections":["hero","x"]}`}},
		{"spec is not JSON", [4]string{"i2", "section_source_drift:about", "t", `not json`}},
		{"authoritative is missing", [4]string{"i3", "section_source_drift:about", "t",
			`{"page_name":"about","pages_sections":["hero","x"]}`}},
		{"authoritative is not an array", [4]string{"i4", "section_source_drift:about", "t",
			`{"page_name":"about","authoritative":"hero","pages_sections":["hero","x"]}`}},
		{"page_name is empty", [4]string{"i5", "section_source_drift:", "t",
			`{"page_name":"","authoritative":["hero"],"pages_sections":["hero","x"]}`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, _, done := runDriftCheck(t, driftMockPlan{
				planRows:  [][2]string{{"about", "hero"}},
				cache:     map[string]string{"about": `["hero"]`},
				slots:     [][2]string{{"about", "hero"}},
				openItems: [][4]string{tc.item},
			})
			defer done()
			if len(res.Resolved) != 0 {
				t.Fatalf("closed an item this check's predicate does not speak for: %+v", res.Resolved)
			}
		})
	}
}

// TestNoAuthorityTodayNeedsADemandControl. "No higher source names this page"
// is a valid reason to close — the finding's premise is gone. But it is only an
// OBSERVATION if the authority read actually worked. A read that returned
// nothing for the entire site is indistinguishable from a mid-replan window,
// and closing on it would retract every drift item on the site at once.
func TestNoAuthorityTodayNeedsADemandControl(t *testing.T) {
	item := [4]string{
		"item-1", "section_source_drift:about", "2026-07-28T00:00:00Z",
		`{"page_name":"about","authoritative":["hero","faq"],"pages_sections":["hero"]}`,
	}

	t.Run("no rows for the whole site → control FAILS → left open", func(t *testing.T) {
		res, _, done := runDriftCheck(t, driftMockPlan{
			planRows:  nil, // tier 1 empty
			cache:     map[string]string{"about": `["hero"]`},
			slots:     [][2]string{{"about", "hero"}},
			openItems: [][4]string{item},
		})
		defer done()
		if len(res.Resolved) != 0 {
			t.Fatalf("closed on an authority read that produced nothing at all: %+v", res.Resolved)
		}
	})

	t.Run("rows for a SIBLING page → control HOLDS → closed", func(t *testing.T) {
		res, _, done := runDriftCheck(t, driftMockPlan{
			planRows:  [][2]string{{"other", "hero"}}, // the read worked; it just has no row for 'about'
			cache:     map[string]string{"about": `["hero"]`, "other": `["hero"]`},
			slots:     [][2]string{{"about", "hero"}},
			openItems: [][4]string{item},
		})
		defer done()
		if len(res.Resolved) != 1 {
			t.Fatalf("the authority read demonstrably worked and named no higher source for this page; Resolved = %d", len(res.Resolved))
		}
		if res.Resolved[0].Evidence["serving_source"] != "pages.sections" {
			t.Errorf("serving source = %v, want pages.sections", res.Resolved[0].Evidence["serving_source"])
		}
	})
}

// TestSeverityFollowsWouldDropPresent — the detection-time grade, which is the
// only part of this change that acts BEFORE the loss.
func TestSeverityFollowsWouldDropPresent(t *testing.T) {
	t.Run("the sync will drop a section the page carries → high", func(t *testing.T) {
		res, _, done := runDriftCheck(t, driftMockPlan{
			planRows: [][2]string{{"about", "hero"}},
			cache:    map[string]string{"about": `["hero","spec-sheet"]`},
			slots:    [][2]string{{"about", "hero"}, {"about", "spec-sheet"}},
		})
		defer done()
		if len(res.WorkItems) != 1 {
			t.Fatalf("want 1 item, got %d", len(res.WorkItems))
		}
		if res.WorkItems[0].Severity != "high" {
			t.Errorf("severity = %q, want high — the next build destroys a section the page carries", res.WorkItems[0].Severity)
		}
	})
	t.Run("the sync only ADDS → medium", func(t *testing.T) {
		res, _, done := runDriftCheck(t, driftMockPlan{
			planRows: [][2]string{{"about", "hero"}, {"about", "faq"}},
			cache:    map[string]string{"about": `["hero"]`},
			slots:    [][2]string{{"about", "hero"}},
		})
		defer done()
		if len(res.WorkItems) != 1 {
			t.Fatalf("want 1 item, got %d", len(res.WorkItems))
		}
		if res.WorkItems[0].Severity != "medium" {
			t.Errorf("severity = %q, want medium — nothing is lost by an addition", res.WorkItems[0].Severity)
		}
	})
	t.Run("drops a name the page does NOT carry → medium", func(t *testing.T) {
		res, _, done := runDriftCheck(t, driftMockPlan{
			planRows: [][2]string{{"about", "hero"}},
			cache:    map[string]string{"about": `["hero","ghost"]`},
			slots:    [][2]string{{"about", "hero"}}, // no `ghost` row on the page
		})
		defer done()
		if res.WorkItems[0].Severity != "medium" {
			t.Errorf("severity = %q, want medium — the cache named a section the page never had", res.WorkItems[0].Severity)
		}
	})
}

// TestSummaryNamesWhatTheBuildWillDestroy. The old summary printed two lists and
// left the reader to diff them; six items sat undifferentiated for up to 37 days.
func TestSummaryNamesWhatTheBuildWillDestroy(t *testing.T) {
	res, _, done := runDriftCheck(t, driftMockPlan{
		planRows: [][2]string{{"about", "hero"}},
		cache:    map[string]string{"about": `["hero","spec-sheet"]`},
		slots:    [][2]string{{"about", "hero"}, {"about", "spec-sheet"}},
	})
	defer done()
	s := res.WorkItems[0].Summary
	if !strings.Contains(s, "DROP") || !strings.Contains(s, "spec-sheet") {
		t.Errorf("summary does not name what the next build destroys: %q", s)
	}
	if !strings.Contains(res.WorkItems[0].SpecJSON, "would_drop_present") {
		t.Error("spec omits would_drop_present — the retraction pass reads it to grade the receipt")
	}
}

// ─── Source pins ────────────────────────────────────────────────────────────

// TestOneRetractionConstructor is the structural half of the pairing.
//
// Go cannot make an in-package struct literal uncompilable, so "it will not
// build" is not available. What IS available is refusing a second construction
// site: the receipt is attached by the same function that computes `lost`, and
// a future author who writes a second ResolvedFinding here — without the
// coupling — fails this test rather than passing review.
func TestOneRetractionConstructor(t *testing.T) {
	src, err := os.ReadFile("check_section_source_drift.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	body := string(src)

	if n := len(regexp.MustCompile(`ResolvedFinding\{`).FindAllString(body, -1)); n != 1 {
		t.Errorf("found %d ResolvedFinding constructions, want exactly 1 (driftResolution.retraction).\n"+
			"A second one can retract WITHOUT the receipt that makes the retraction honest — which is\n"+
			"bugs_open/469 reintroduced. Route it through retraction() instead.", n)
	}
	if n := len(regexp.MustCompile(`"section_composition_lost"`).FindAllString(body, -1)); n != 2 {
		t.Errorf("found %d occurrences of the receipt's item type, want exactly 2 (the const and the one field site).\n"+
			"A third means a second filing site, and the receipt must only ever be filed as part of a retraction.", n)
	}
}

// TestItemTypeConstantsMatchTheStrings. The field sites spell their strings so
// verifier_coverage_test.go's source scan can see them; the constants serve the
// keys and queries. This is what stops the two drifting.
func TestItemTypeConstantsMatchTheStrings(t *testing.T) {
	if sectionDriftItemType != "section_source_drift" {
		t.Errorf("sectionDriftItemType = %q — the constant and the field site have drifted", sectionDriftItemType)
	}
	if sectionCompositionLostItemType != "section_composition_lost" {
		t.Errorf("sectionCompositionLostItemType = %q — the constant and the field site have drifted", sectionCompositionLostItemType)
	}
	if got := sectionDriftItemKey("about"); got != "section_source_drift:about" {
		t.Errorf("item key = %q; the retraction pass matches on this exact shape", got)
	}
}

// TestSectionCompositionLostKeyDistinguishesDistinctLosses — an identical repeat
// dedups, a different loss does not. Gripper-spec-sheet's real history (added,
// destroyed, rescued by hand, destroyed again) needs the second event to file.
func TestSectionCompositionLostKeyDistinguishesDistinctLosses(t *testing.T) {
	a := sectionCompositionLostKey("p", []string{"x"})
	b := sectionCompositionLostKey("p", []string{"x"})
	c := sectionCompositionLostKey("p", []string{"x", "y"})
	d := sectionCompositionLostKey("q", []string{"x"})
	if a != b {
		t.Error("the same loss on the same page must dedup")
	}
	if a == c {
		t.Error("losing [x] and losing [x,y] must be distinct findings")
	}
	if a == d {
		t.Error("the same loss on different pages must be distinct findings")
	}
	if !strings.HasPrefix(a, "section_composition_lost:p:") {
		t.Errorf("key %q does not carry its type prefix — workItemKey's contract is that the prefix equals the item_type", a)
	}
}
