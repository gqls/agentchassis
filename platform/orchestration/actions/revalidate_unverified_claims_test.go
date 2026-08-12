// FILE: platform/orchestration/actions/revalidate_unverified_claims_test.go
//
// The verdict ladder for claims_unverified retraction. Every arm is exercised
// without a database, which is why unverifiedClaimsVerdict is split from the
// query glue.
//
// The property under test is not "does it return the right string" but the one
// that makes a wrong answer expensive: `resolved` CLOSES a live human-review
// item about an unsupported factual claim, so it must be reachable ONLY from
// "components were read, they were clean, AND the page itself has changed since
// the finding was filed". Every state where the scan read nothing, read only
// part of the page, or cannot show the page moved, has to come back
// non-terminal.
package actions

import (
	"context"
	"strings"
	"testing"
	"time"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// The finding was filed at filedAt. changedAfter is a component edit made since;
// changedBefore is one made before it, which must NOT license a close — that is
// the whole point of the owner's 2026-08-09 copy-changed gate.
var (
	filedAt       = time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	changedAfter  = filedAt.Add(48 * time.Hour)
	changedBefore = filedAt.Add(-48 * time.Hour)
)

// The CLAIM-GRANULAR gate's fixtures (council round 5, `compliance` HIGH).
//
// flaggedHero is what the work item recorded at filing time: this page asserted
// "90,790" in the `hero` slot. heroCleaned is that slot AFTER the copy was fixed
// — read, and the cited text gone. heroUntouched is the case the gate exists for:
// the page stopped tripping the check (so the scan is clean) while the words the
// finding cited are still sitting there, which means the STANDARD moved.
//
// Every test below that expects to reach the owner's timestamp gate must pass
// heroCleaned, or it would refuse at the claim gate instead and pass for the
// wrong reason — the failure mode that makes a green suite meaningless.
var flaggedHero = []flaggedClaim{{Check: "unregistered_number", Slot: "hero", Matched: "90,790"}}

func heroCleaned() map[string]string {
	return map[string]string{"hero": "<p>We help small firms adopt AI.</p>"}
}
func heroUntouched() map[string]string {
	return map[string]string{"hero": "<p>Trusted by 90,790 customers.</p>"}
}

func TestUnverifiedClaimsVerdictLadder(t *testing.T) {
	finding := func(name string) checks.ClaimFinding {
		return checks.ClaimFinding{Check: name}
	}

	cases := []struct {
		name        string
		scan        *checks.ClaimsPageScan
		want        string
		reasonHas   string
		explanation string
	}{
		{
			name:        "page absent from the re-scan",
			scan:        nil,
			want:        revalidationUnknown,
			reasonHas:   "not evidence the claims were removed",
			explanation: "a deleted or unbuilt page is not proof the copy was corrected",
		},
		{
			name:        "every component locked, nothing read",
			scan:        &checks.ClaimsPageScan{PageID: "p1", ComponentsExamined: 0, ComponentsSkippedLocked: 3},
			want:        revalidationUnknown,
			reasonHas:   "examined nothing",
			explanation: "empty findings here means the audit read nothing at all",
		},
		{
			name: "claims still present",
			scan: &checks.ClaimsPageScan{
				PageID: "p1", PageName: "about", ComponentsExamined: 4, NewestComponentUpdate: changedAfter,
				Findings: []checks.ClaimFinding{finding("banned_claim"), finding("unregistered_number")},
			},
			want:        revalidationStillHolds,
			reasonHas:   "still carries",
			explanation: "the finding is unchanged, so the item stays open",
		},
		{
			name: "claims present AND some components locked - still_holds wins",
			scan: &checks.ClaimsPageScan{
				PageID: "p1", PageName: "about", ComponentsExamined: 2, ComponentsSkippedLocked: 1,
				NewestComponentUpdate: changedAfter,
				Findings:              []checks.ClaimFinding{finding("banned_claim")},
			},
			want:        revalidationStillHolds,
			reasonHas:   "still carries",
			explanation: "ladder order: a page that still asserts a banned claim must not be reported as unreadable",
		},
		{
			name: "clean, but part of the page was locked and unread",
			scan: &checks.ClaimsPageScan{PageID: "p1", PageName: "about", ComponentsExamined: 2,
				ComponentsSkippedLocked: 1, NewestComponentUpdate: changedAfter,
				ExaminedTextBySlot: heroCleaned()},
			want:        revalidationUnknown,
			reasonHas:   "were not read",
			explanation: "the reported claims may live in the pinned components, so closing would assert something unchecked",
		},
		{
			name: "clean and whole page read, but the COPY never changed - the register moved",
			scan: &checks.ClaimsPageScan{PageID: "p1", PageName: "about", ComponentsExamined: 3,
				NewestComponentUpdate: changedBefore, ExaminedTextBySlot: heroCleaned()},
			want:        revalidationUnknown,
			reasonHas:   "register moved, not the page",
			explanation: "OWNER RULING 2026-08-09: a register edit alone must never retract a factual-claim finding",
		},
		{
			name: "clean, whole page read, and the copy changed since filing",
			scan: &checks.ClaimsPageScan{PageID: "p1", PageName: "about", ComponentsExamined: 3,
				NewestComponentUpdate: changedAfter, ExaminedTextBySlot: heroCleaned()},
			want:        revalidationResolved,
			explanation: "the only state that licenses closing the item",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := unverifiedClaimsVerdict("p1", filedAt, flaggedHero, tc.scan)
			if got.Verdict != tc.want {
				t.Errorf("verdict = %q, want %q (%s)", got.Verdict, tc.want, tc.explanation)
			}
			if tc.reasonHas != "" && !strings.Contains(got.Reason, tc.reasonHas) {
				t.Errorf("reason %q does not contain %q — the reason is what a human reads in the\n"+
					"work item, so a verdict with a reason that no longer explains it is a defect", got.Reason, tc.reasonHas)
			}
			if got.Reason == "" {
				t.Error("every verdict must carry a reason; an unexplained close is unreviewable")
			}
		})
	}
}

// The load-bearing assertion, stated as a property rather than per-case: NOTHING
// reaches `resolved` unless components were actually examined. If a future edit
// reorders the ladder or adds an arm that returns resolved on an unread page,
// this fails even though every individual case above might still pass.
func TestUnverifiedClaimsNeverResolvesWithoutReadingSomething(t *testing.T) {
	unreadable := []*checks.ClaimsPageScan{
		nil,
		{PageID: "p1", NewestComponentUpdate: changedAfter, ExaminedTextBySlot: heroCleaned()},
		{PageID: "p1", ComponentsSkippedLocked: 1, NewestComponentUpdate: changedAfter, ExaminedTextBySlot: heroCleaned()},
		{PageID: "p1", ComponentsSkippedLocked: 9, NewestComponentUpdate: changedAfter, ExaminedTextBySlot: heroCleaned()},
	}
	for _, scan := range unreadable {
		got := unverifiedClaimsVerdict("p1", filedAt, flaggedHero, scan)
		if got.Verdict == revalidationResolved {
			t.Errorf("scan %+v produced `resolved` without examining a component — "+
				"that closes a live human-review item on the strength of having read nothing", scan)
		}
	}
}

// THE OWNER'S GATE, stated as a property so a future edit cannot quietly drop it
// (OWNER RULING 2026-08-09, answering the council's `compliance` seat across two
// rounds). The register is DATA: adding a fact makes a previously unregistered
// number verifiable and the page stops tripping the check with the copy
// untouched. `resolved` must be unreachable unless an EXAMINED component was
// edited after the finding was filed.
//
// The zero-value case is the one most likely to regress: a page whose components
// carry no timestamp must refuse, never close.
func TestUnverifiedClaimsNeverResolvesWhenOnlyTheRegisterMoved(t *testing.T) {
	cases := []struct {
		name    string
		newest  time.Time
		filedAt time.Time
	}{
		{"component edited before the finding was filed", changedBefore, filedAt},
		{"component edited at exactly the filing instant", filedAt, filedAt},
		{"no timestamp at all on any examined component", time.Time{}, filedAt},
		{"filing date unknown, page edited long ago", changedBefore, time.Time{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := unverifiedClaimsVerdict("p1", tc.filedAt, flaggedHero, &checks.ClaimsPageScan{
				PageID: "p1", PageName: "about", ComponentsExamined: 3,
				NewestComponentUpdate: tc.newest, ExaminedTextBySlot: heroCleaned(),
			})
			if got.Verdict == revalidationResolved {
				t.Errorf("a clean page whose copy cannot be shown to have moved returned `resolved` — "+
					"that is a register edit closing a factual-claim review row, which the owner ruling of "+
					"2026-08-09 forbids (newest=%v filed=%v)", tc.newest, tc.filedAt)
			}
		})
	}
}

// A clean page whose copy DID change must retract: the sweep exists to withdraw
// findings that stopped being true, and a revalidator that can only say
// still_holds is the armed-but-inert shape this lane keeps finding. The gate
// above must narrow the close, not abolish it.
func TestUnverifiedClaimsActuallyRetracts(t *testing.T) {
	got := unverifiedClaimsVerdict("p1", filedAt, flaggedHero, &checks.ClaimsPageScan{
		PageID: "p1", PageName: "about", ComponentsExamined: 5, NewestComponentUpdate: changedAfter,
		ExaminedTextBySlot: heroCleaned(),
	})
	if got.Verdict != revalidationResolved {
		t.Fatalf("a fully-read, claim-free page whose copy changed after filing returned %q; if this "+
			"cannot resolve, registering it only adds scan cost and the type stays parked for ever", got.Verdict)
	}
	if got.Evidence["components_examined"] != 5 {
		t.Errorf("evidence should record how much was read, got %v — the count is what makes the "+
			"close auditable after the fact", got.Evidence["components_examined"])
	}
	// Both timestamps travel with the close so the gate's decision is auditable
	// later without re-deriving it from a table whose rows may have moved on.
	if got.Evidence["item_filed_at"] != filedAt || got.Evidence["newest_component_update"] != changedAfter {
		t.Errorf("evidence must carry both sides of the copy-changed comparison, got filed=%v newest=%v",
			got.Evidence["item_filed_at"], got.Evidence["newest_component_update"])
	}
}

// The spec-missing arm returns before touching the database, so a nil *sql.DB is
// safe here and proves the early return really is early.
//
// This is also the arm the grouped site-chrome item lands on: its item_key is the
// literal `claims:site_components` and its spec carries `surface`, not `page_id`.
// No such item is open today, so this test is the only thing exercising it.
func TestUnverifiedClaimsRefusesAnItemWithNoPageID(t *testing.T) {
	got := revalidateUnverifiedClaims(context.Background(), nil, parkedReviewItem{
		ItemType:  "claims_unverified",
		ItemKey:   "claims:site_components",
		Spec:      map[string]interface{}{"check": "unverified_claims", "surface": "site_components"},
		CreatedAt: filedAt,
	}, nil)
	if got.Verdict != revalidationUnknown {
		t.Errorf("verdict = %q, want %q", got.Verdict, revalidationUnknown)
	}
	if !strings.Contains(got.Reason, "item_key") {
		t.Errorf("reason should say the item_key is deliberately not parsed, got %q", got.Reason)
	}
}

// The by_check evidence map is what a reviewer reads to see WHICH claim classes
// are still live on the page. A still_holds that cannot say what it found is not
// reviewable — and this is the map that would silently empty if ClaimFinding's
// Check field were ever renamed.
func TestUnverifiedClaimsStillHoldsNamesTheCheckClasses(t *testing.T) {
	got := unverifiedClaimsVerdict("p1", filedAt, flaggedHero, &checks.ClaimsPageScan{
		PageID: "p1", PageName: "about", ComponentsExamined: 2, NewestComponentUpdate: changedAfter,
		Findings: []checks.ClaimFinding{
			{Check: "banned_claim"}, {Check: "banned_claim"}, {Check: "unregistered_number"},
		},
	})
	byCheck, ok := got.Evidence["by_check"].(map[string]int)
	if !ok {
		t.Fatalf("by_check evidence missing or wrong type: %#v", got.Evidence["by_check"])
	}
	if byCheck["banned_claim"] != 2 || byCheck["unregistered_number"] != 1 {
		t.Errorf("by_check = %v, want banned_claim=2 unregistered_number=1", byCheck)
	}
}

// THE CLAIM-GRANULAR GATE, stated as a property (council round 5, `compliance`
// HIGH, raised in rounds 3, 4 and 5). The owner's gate above proves THE PAGE
// MOVED; it cannot prove THE FLAGGED CLAIM WAS ADDRESSED, and an edit to an
// unrelated slot satisfies it. So: a clean scan whose cited text is still sitting
// in the slot it was cited from must never close, however recently the page was
// touched.
//
// changedAfter is passed deliberately in every case — the owner's gate is
// SATISFIED here. If this test ever passes because the timestamp refused instead,
// it would be asserting nothing at all.
func TestUnverifiedClaimsNeverResolvesWhenTheFlaggedTextIsStillThere(t *testing.T) {
	got := unverifiedClaimsVerdict("p1", filedAt, flaggedHero, &checks.ClaimsPageScan{
		PageID: "p1", PageName: "about", ComponentsExamined: 3,
		NewestComponentUpdate: changedAfter, ExaminedTextBySlot: heroUntouched(),
	})
	if got.Verdict == revalidationResolved {
		t.Fatalf("closed a factual-claim item while the text it cited (%q) is still in its slot — "+
			"the page stopped tripping the check because the STANDARD moved, which is exactly what "+
			"the compliance seat objected to in three successive council rounds", flaggedHero[0].Matched)
	}
	if got.Verdict != revalidationStillHolds {
		t.Errorf("verdict = %q, want %q: the words are still on the page, so the finding still holds",
			got.Verdict, revalidationStillHolds)
	}
	if !strings.Contains(got.Reason, "STILL in the component") {
		t.Errorf("reason must say which side of the comparison failed, got %q", got.Reason)
	}
	if got.Evidence["flagged_texts"] != 1 {
		t.Errorf("evidence must record how many cited texts were checked, got %v", got.Evidence["flagged_texts"])
	}
}

// Case-insensitivity is not cosmetic: a rerender that changes only the casing of
// surrounding markup must not read as the claim having been removed.
func TestUnverifiedClaimsFlaggedTextMatchIsCaseInsensitive(t *testing.T) {
	flagged := []flaggedClaim{{Check: "banned_claim", Slot: "hero", Matched: "Independently Verified"}}
	got := unverifiedClaimsVerdict("p1", filedAt, flagged, &checks.ClaimsPageScan{
		PageID: "p1", PageName: "about", ComponentsExamined: 3, NewestComponentUpdate: changedAfter,
		ExaminedTextBySlot: map[string]string{"hero": "<p>our results are independently verified</p>"},
	})
	if got.Verdict == revalidationResolved {
		t.Error("a casing difference was read as the claim having been removed")
	}
}

// "I could not look" must not be spelled the same as "it is gone". A slot absent
// from the examined set means the component was deleted, renamed or human-locked —
// none of which is evidence the claim was withdrawn.
func TestUnverifiedClaimsRefusesWhenTheFlaggedSlotWasNotExamined(t *testing.T) {
	for _, tc := range []struct {
		name  string
		slots map[string]string
	}{
		{"the slot is absent entirely", map[string]string{"footer": "unrelated copy"}},
		{"nothing was recorded at all", map[string]string{}},
		{"nil map", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := unverifiedClaimsVerdict("p1", filedAt, flaggedHero, &checks.ClaimsPageScan{
				PageID: "p1", PageName: "about", ComponentsExamined: 3,
				NewestComponentUpdate: changedAfter, ExaminedTextBySlot: tc.slots,
			})
			if got.Verdict != revalidationUnknown {
				t.Errorf("verdict = %q, want %q: the cited slot was never read, so its claim "+
					"cannot be confirmed gone", got.Verdict, revalidationUnknown)
			}
			if !strings.Contains(got.Reason, "could not be re-checked") {
				t.Errorf("reason should say the check could not be made, got %q", got.Reason)
			}
		})
	}
}

// An item carrying no finding text at all cannot be judged claim-granularly, and
// the safe reading of "no recorded claim" is refusal — not a free close.
func TestUnverifiedClaimsRefusesAnItemWithNoRecordedFindingText(t *testing.T) {
	for _, tc := range []struct {
		name    string
		flagged []flaggedClaim
	}{
		{"no findings recorded", nil},
		{"a finding with empty matched text", []flaggedClaim{{Check: "banned_claim", Slot: "hero"}}},
		{"matched is whitespace only", []flaggedClaim{{Check: "banned_claim", Slot: "hero", Matched: "   "}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := unverifiedClaimsVerdict("p1", filedAt, tc.flagged, &checks.ClaimsPageScan{
				PageID: "p1", PageName: "about", ComponentsExamined: 3,
				NewestComponentUpdate: changedAfter, ExaminedTextBySlot: heroCleaned(),
			})
			if got.Verdict == revalidationResolved {
				t.Errorf("closed an item whose recorded claim text could not be checked (%s)", tc.name)
			}
		})
	}
}

// flaggedClaimsFromSpec reads the producer's own shape. A spec whose findings[]
// is missing, the wrong type, or holds junk must degrade to "nothing to check"
// rather than panicking inside the sweep — one malformed row must not take the
// whole scheduled run down with it.
func TestFlaggedClaimsFromSpecToleratesJunk(t *testing.T) {
	cases := []map[string]interface{}{
		{},
		{"findings": nil},
		{"findings": "not an array"},
		{"findings": []interface{}{"not an object", 42, nil}},
	}
	for i, spec := range cases {
		if got := flaggedClaimsFromSpec(spec); len(got) != 0 {
			t.Errorf("case %d: expected no usable claims, got %+v", i, got)
		}
	}
	full := map[string]interface{}{"findings": []interface{}{
		map[string]interface{}{"check": "banned_claim", "slot_name": "hero", "matched": "independently verified"},
		map[string]interface{}{"check": "unregistered_number", "slot_name": "stats", "matched": "90,790"},
	}}
	got := flaggedClaimsFromSpec(full)
	if len(got) != 2 || got[0].Matched != "independently verified" || got[1].Slot != "stats" {
		t.Errorf("real spec shape misread: %+v", got)
	}
}

// PREDICATE PARITY, and the trap inside it (council round 6, `reuse_agent`).
// claimStillOnPage now searches datahelpers.ExtractAssertionText, the same
// extraction the emit side scans. That is right for rendered_html — but
// ExaminedTextBySlot also carries content_data, which is JSON, not HTML. If the
// extractor dropped non-HTML text, a claim living in stored content would read as
// GONE, which GRANTS closure: the unsafe direction, and invisible without this test.
func TestClaimStillOnPageSeesClaimsInNonHTMLStoredContent(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{"plain prose, no markup at all", "Trusted by 90,790 customers across the UK."},
		{"stored content_data JSON", `{"headline":"Trusted by 90,790 customers","cta":"Book"}`},
		{"html and json concatenated, as the scan stores them", `<p>Our reach</p>{"stat":"90,790"}`},
		{"claim inside ordinary markup", `<section><h2>Reach</h2><p>Trusted by 90,790 customers.</p></section>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scan := &checks.ClaimsPageScan{PageID: "p1", ExaminedTextBySlot: map[string]string{"hero": tc.text}}
			present, judgeable := claimStillOnPage(scan, flaggedHero[0])
			if !judgeable {
				t.Fatalf("the slot was examined, so the question is answerable; got judgeable=false")
			}
			if !present {
				t.Errorf("claim %q not found in %q — a claim the extractor cannot see reads as REMOVED, "+
					"which closes a live factual-claim item", flaggedHero[0].Matched, tc.text)
			}
		})
	}
}

// The other half of parity: text that exists only in MARKUP is not an assertion
// and must not hold an item open. This is what searching raw html got wrong.
func TestClaimStillOnPageIgnoresMarkupOnlyMatches(t *testing.T) {
	scan := &checks.ClaimsPageScan{PageID: "p1", ExaminedTextBySlot: map[string]string{
		"hero": `<div class="grid-90,790" data-count="90,790"><p>We help small firms adopt AI.</p></div>`,
	}}
	present, judgeable := claimStillOnPage(scan, flaggedHero[0])
	if !judgeable {
		t.Fatal("slot was examined; expected an answerable question")
	}
	if present {
		t.Error("a token appearing only in a class name and a data attribute was read as a live claim — " +
			"no reader sees it, and holding the item open on that basis is a false refusal")
	}
}

// THE SHARED-LOADER CONTRACT, made mechanical (council round 6, `editquality` and
// `guardian` MEDIUM, independently).
//
// loadParkedReviewItems serves ALL covered item types, so a column added to its
// SELECT without a matching rows.Scan destination is a runtime scan error that
// fails the sweep for EVERY revalidator at once — not a compile error, and
// invisible to any caller. The two halves were previously joined only by a
// warning comment, in a change whose own submission argued that a comment is not
// a control on a shared tree. This is that control.
func TestParkedReviewItemSelectAndScanAgree(t *testing.T) {
	var it parkedReviewItem
	var specJSON []byte
	dests := parkedReviewItemScanDests(&it, &specJSON)

	if len(parkedReviewItemColumns) != len(dests) {
		t.Fatalf("loadParkedReviewItems SELECT has %d column(s) but rows.Scan takes %d destination(s): "+
			"%v.\nThese are ONE contract. A mismatch is `sql: expected %d destination arguments in Scan, "+
			"not %d` at runtime, and it breaks the sweep for every covered item_type, not just yours.",
			len(parkedReviewItemColumns), len(dests), parkedReviewItemColumns, len(parkedReviewItemColumns), len(dests))
	}
	if len(parkedReviewItemColumns) == 0 {
		t.Fatal("the SELECT list is empty; the loader would select nothing and every revalidator would idle")
	}
	for i, d := range dests {
		if d == nil {
			t.Errorf("scan destination %d (%q) is nil — Scan would fail at runtime",
				i, parkedReviewItemColumns[i])
		}
	}
	// created_at is the newest member and the one the claims_unverified gates
	// depend on; naming it here means dropping it fails a test rather than
	// silently sending every item to the zero-value `unknown` arm for ever.
	var found bool
	for _, c := range parkedReviewItemColumns {
		if c == "created_at" {
			found = true
		}
	}
	if !found {
		t.Error("created_at is absent from the SELECT: filedAt would arrive as the zero value and the " +
			"copy-changed gate could never reach `resolved` — the exact defect editquality predicted in round 3")
	}
}
