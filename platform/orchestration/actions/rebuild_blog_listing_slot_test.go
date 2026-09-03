package actions

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Tests for bugs_open/457 — rebuild_blog_listing appended an orphan
// page_components row on every run and then hard-failed on migration 316's
// byte-identical guard.
//
// WHY THE DECISION IS A PURE FUNCTION, AND WHY THAT IS THE POINT. The defect was
// never in the INSERT statement; it was that the caller asked the wrong
// question. findBlogListingSlot returned (name, componentID) and the caller read
// a nil id as "this slot is free", when only ONE of the four resolution paths
// had looked in page_components at all. Extracting decideBlogListingWrite makes
// the whole table — including every case that must NOT write — assertable with
// no database, so the cases that matter most are the cheapest to pin.
//
// ⚠ THE TRAP THESE TESTS EXIST TO HOLD SHUT. Binding a real component_id on the
// INSERT (which the fix also does, so rows can be attributed) stops the new row
// colliding with the NULL-component_id rows already present, because
// uq_page_components_no_byte_identical_duplicate is NULLS NOT DISTINCT. That
// constraint is the ONLY thing currently reporting the duplication. So a change
// that binds the id without deciding the write in Go first would convert a loud
// failure into a silent seventh append — strictly worse than the bug. TestNeverInsertsWhenTheSlotIsOccupied
// is the guard on that, and it is the one to keep if any of these are ever
// trimmed.

func occupied(origin blogSlotOrigin, name string, occupants int) blogListingSlot {
	s := blogListingSlot{Name: name, Origin: origin, Occupants: occupants, OccupancyKnown: true}
	if occupants == 1 {
		s.Existing = uuid.New()
	}
	return s
}

// The motivating case: boxingonline.com /articles/index.html, where the plan
// names no listing slot, strategy 2b guessed `generic-text-block` (a slot the
// page uses for PROSE), and seven rows now sit in it.
//
// Mutation that must kill it: order the switch so Origin is tested before
// Occupants, or drop the Occupants > 1 arm.
func TestBlogListingRefusesToWriteIntoAnOccupiedGuessedSlot(t *testing.T) {
	op, reason := decideBlogListingWrite(occupied(slotOriginPlanFallback, "generic-text-block", 7))
	if op != opRefuseAmbiguous {
		t.Fatalf("seven occupants must refuse, got op %v (%s)", op, reason)
	}
}

// The root cause, pinned. Before the fix the caller branched on "did strategy 1
// return an id", so a slot resolved by any other strategy took the INSERT arm
// even when a row was sitting in it. The decision must depend on OCCUPANCY, not
// on which strategy found the name.
//
// Mutation that must kill it: restore the old `existingComponentID != uuid.Nil`
// proxy, i.e. make the update arm conditional on Origin == slotOriginExistingRow.
func TestBlogListingUpdatesTheSingleOccupantWhateverStrategyNamedTheSlot(t *testing.T) {
	for _, origin := range []blogSlotOrigin{
		slotOriginExistingRow, slotOriginPlanListing, slotOriginPlanFallback, slotOriginDefault,
	} {
		op, reason := decideBlogListingWrite(occupied(origin, "blog-listing", 1))
		if op != opUpdate {
			t.Errorf("origin %s with exactly one occupant must UPDATE it, got %v (%s)", origin, op, reason)
		}
	}
}

// The trap guard. No combination of origin and occupancy may insert into a slot
// that already holds rows — that is what makes the migration-316 interaction
// safe rather than merely quiet.
//
// Mutation that must kill it: bind component_id on the INSERT without the
// occupancy check, i.e. let any non-existing-row origin fall through to opInsert.
func TestBlogListingNeverInsertsWhenTheSlotIsOccupied(t *testing.T) {
	for _, origin := range []blogSlotOrigin{
		slotOriginExistingRow, slotOriginPlanListing, slotOriginPlanFallback, slotOriginDefault,
	} {
		for _, occupants := range []int{1, 2, 6, 7} {
			op, _ := decideBlogListingWrite(occupied(origin, "generic-text-block", occupants))
			if op == opInsert {
				t.Errorf("origin %s with %d occupants must never INSERT", origin, occupants)
			}
		}
	}
}

// A guessed slot name is by construction NOT a listing slot — strategy 1 or 2a
// would have matched if it were. Creating a listing there is a silent
// structural edit to someone else's page, repeated once per run, which is how
// six orphans accumulated in two days.
//
// Mutation that must kill it: let strategy 2b's name grant write authority again.
func TestBlogListingRefusesToInventASlotFromAGuess(t *testing.T) {
	op, reason := decideBlogListingWrite(occupied(slotOriginPlanFallback, "generic-text-block", 0))
	if op != opRefuseNoSlotAuthority {
		t.Fatalf("a guessed slot with no occupants must refuse, got %v (%s)", op, reason)
	}
	if !strings.Contains(reason, "generic-text-block") {
		t.Errorf("the refusal must name the slot it declined, got %q", reason)
	}
}

// Same rule one step along: a page that declares no sections at all has not
// named a listing slot either, so `blog-listing` here is a default, not a plan.
//
// Mutation that must kill it: treat slotOriginDefault as authorising a write.
func TestBlogListingRefusesToInventASlotFromTheDefault(t *testing.T) {
	op, _ := decideBlogListingWrite(occupied(slotOriginDefault, "blog-listing", 0))
	if op != opRefuseNoSlotAuthority {
		t.Fatalf("the default slot name on a plan-less page must refuse, got %v", op)
	}
}

// The positive control, so the refusals above cannot be satisfied by a function
// that refuses everything. When the page's own plan names a listing-class slot
// and nothing occupies it, creating it is the correct behaviour.
//
// Mutation that must kill it: over-tighten the refusal to cover 2a as well,
// which would leave a planned-but-unbuilt listing permanently unbuilt.
func TestBlogListingCreatesTheSlotThePlanDeclares(t *testing.T) {
	slot := occupied(slotOriginPlanListing, "blog-listing", 0)
	slot.PlanPos = 2
	op, reason := decideBlogListingWrite(slot)
	if op != opInsert {
		t.Fatalf("a plan-declared empty listing slot must be created, got %v (%s)", op, reason)
	}
}

// An error reading occupancy is not an empty slot. This is the swallow class:
// `occupants = 0` on error would make a failed lookup indistinguishable from a
// free slot, and the append would come straight back.
//
// Mutation that must kill it: default OccupancyKnown to true, or set
// occupants = 0 in findBlogListingSlot's error branch.
func TestBlogListingOccupancyErrorDoesNotReadAsAnEmptySlot(t *testing.T) {
	op, reason := decideBlogListingWrite(blogListingSlot{
		Name: "blog-listing", Origin: slotOriginPlanListing, OccupancyKnown: false,
	})
	if op != opRefuseUnknown {
		t.Fatalf("an unreadable occupancy must refuse, got %v (%s)", op, reason)
	}
}

// The refusal must never be an error return, and this is the cheapest test here
// for the largest consequence. rebuild_blog_listing is an unconditional step of
// the rerender-pages workflow, sitting before create_rerender_items, and that
// workflow declares no error_step. An error return therefore aborts the run and
// creates NONE of the page rerenders — the 18-page outage bugs_open/457 was
// filed for. A refusal that returns an error reproduces the outage under a new
// name.
//
// Mutation that must kill it: `return nil, fmt.Errorf(...)` on any refusal path.
func TestBlogListingRefusalPathsReturnNoError(t *testing.T) {
	src := readActionSource(t, "rebuild_blog_listing_action.go")

	// ⚠ Scope the window FIRST. The obvious form of this test — an `(?s)` regex
	// from the log line to `return nil, fmt.Errorf` — is worthless: `.*?` runs to
	// the end of the FILE, so it matches an unrelated error return hundreds of
	// lines later and reports a defect that is not there. (It did, on the first
	// run of this test.) Cut the block, then assert inside it.
	start := strings.Index(src, "refusing to write the blog listing")
	if start < 0 {
		t.Fatal("could not find the refusal branch — if it was renamed, move this pin with it")
	}
	end := strings.Index(src[start:], "}, nil")
	if end < 0 {
		t.Fatal("the refusal branch does not end in a result return")
	}
	block := src[start : start+end]

	if !strings.Contains(block, "return map[string]interface{}{") {
		t.Error("the refusal must return a result map so the rerender-pages chain continues to create_rerender_items")
	}
	if strings.Contains(block, "fmt.Errorf") {
		t.Error("the refusal returns an error — that aborts rerender-pages before create_rerender_items (bugs_open/457)")
	}
}

// The INSERT must bind component_id, so a row it creates is attributable,
// re-renderable by component, and visible to every component-keyed query.
//
// ⚠ This assertion reads the INSERT's own COLUMN LIST, deliberately. A window
// scan around the statement does not work here and quietly returns the wrong
// answer: this file logs `zap.String("component_id", ...)` a few lines below the
// statement, so "is component_id near the INSERT?" was true even while the
// column was absent — it cleared the one writer on the estate that was actually
// violating the rule.
//
// Mutation that must kill it: drop component_id from the INSERT's column list.
func TestBlogListingInsertBindsComponentID(t *testing.T) {
	src := readActionSource(t, "rebuild_blog_listing_action.go")
	m := regexp.MustCompile(`INSERT INTO page_components \(([^)]*)\)`).FindStringSubmatch(src)
	if m == nil {
		t.Fatal("could not find the INSERT INTO page_components column list")
	}
	if !strings.Contains(m[1], "component_id") {
		t.Errorf("the listing INSERT must bind component_id; column list was %q", m[1])
	}
	if regexp.MustCompile(`VALUES \(\$1, \$2, 3,`).MatchString(src) {
		t.Error("position is hard-coded to 3 again — it must come from the discovered slot")
	}
}

// The second defect fixed in the same change, and the one the council's
// editquality seat correctly noted had no test (round 1, corr 13273c8c, medium).
//
// `function = 'content-listing'` is not unique: [MEASURED 2026-09-03] two active
// rows, the shared `content-listing` and a per-site fork
// `content-listing-guides-boxingonline-com` created a day earlier. With
// `ORDER BY created_at DESC LIMIT 1` the fork won, so one site's template became
// every site's. They are byte-identical today, which is why nothing broke — the
// defect is that the NEXT divergence would propagate silently.
//
// ⚠ WHY THIS IS A SOURCE ASSERTION AND NOT A BEHAVIOURAL ONE, stated rather than
// hidden: the preference is expressed as an ORDER BY and resolved by Postgres
// under a LIMIT 1. A sqlmock test cannot exercise it — the mock returns whatever
// rows the test hands it, in the order the test chose, so it would assert the
// fixture rather than the query. Driving it honestly needs a real database with
// two seeded rows. Until that exists this pins the clause itself, which at least
// fails when someone "fixes" it back to newest-wins.
//
// Mutation that must kill it: restore `ORDER BY created_at DESC`.
func TestBlogListingPrefersTheCanonicalListingComponentOverTheNewest(t *testing.T) {
	src := readActionSource(t, "rebuild_blog_listing_action.go")

	start := strings.Index(src, "func loadContentListingTemplate")
	if start < 0 {
		t.Fatal("loadContentListingTemplate not found — if it was renamed, move this pin with it")
	}
	block := src[start:]
	if end := strings.Index(block, "\nfunc "); end > 0 {
		block = block[:end]
	}

	if !strings.Contains(block, "ORDER BY (name = function) DESC") {
		t.Error("the content-listing lookup must prefer the CANONICAL row (name = function) — " +
			"function='content-listing' is not unique, and ordering by recency alone let one site's " +
			"fork become every site's listing template (bugs_open/457)")
	}
	if regexp.MustCompile(`ORDER BY\s+created_at DESC`).MatchString(block) {
		t.Error("ordering by created_at DESC alone is the defect: the newest row wins, and the newest " +
			"row is whichever site forked the component most recently")
	}
	// The ambiguity must be visible, not silently resolved — RFC_034 makes several
	// forks per function the intended future, so refusing is wrong and staying
	// quiet is worse.
	if !strings.Contains(block, "count(*) OVER ()") {
		t.Error("the lookup must count its candidates so an ambiguity can be logged rather than silently resolved")
	}
}

func readActionSource(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}
