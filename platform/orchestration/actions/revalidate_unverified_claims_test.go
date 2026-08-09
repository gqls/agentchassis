// FILE: platform/orchestration/actions/revalidate_unverified_claims_test.go
//
// The verdict ladder for claims_unverified retraction. Every arm is exercised
// without a database, which is why unverifiedClaimsVerdict is split from the
// query glue.
//
// The property under test is not "does it return the right string" but the one
// that makes a wrong answer expensive: `resolved` CLOSES a live human-review
// item about an unsupported factual claim, so it must be reachable ONLY from
// "components were read, and they were clean". Every state where the scan read
// nothing, or read only part of the page, has to come back non-terminal.
package actions

import (
	"context"
	"strings"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

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
				PageID: "p1", PageName: "about", ComponentsExamined: 4,
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
				Findings: []checks.ClaimFinding{finding("banned_claim")},
			},
			want:        revalidationStillHolds,
			reasonHas:   "still carries",
			explanation: "ladder order: a page that still asserts a banned claim must not be reported as unreadable",
		},
		{
			name:        "clean, but part of the page was locked and unread",
			scan:        &checks.ClaimsPageScan{PageID: "p1", PageName: "about", ComponentsExamined: 2, ComponentsSkippedLocked: 1},
			want:        revalidationUnknown,
			reasonHas:   "were not read",
			explanation: "the reported claims may live in the pinned components, so closing would assert something unchecked",
		},
		{
			name:        "clean, whole page read",
			scan:        &checks.ClaimsPageScan{PageID: "p1", PageName: "about", ComponentsExamined: 3},
			want:        revalidationResolved,
			explanation: "the only state that licenses closing the item",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := unverifiedClaimsVerdict("p1", tc.scan)
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
		{PageID: "p1"},
		{PageID: "p1", ComponentsSkippedLocked: 1},
		{PageID: "p1", ComponentsSkippedLocked: 9},
	}
	for _, scan := range unreadable {
		got := unverifiedClaimsVerdict("p1", scan)
		if got.Verdict == revalidationResolved {
			t.Errorf("scan %+v produced `resolved` without examining a component — "+
				"that closes a live human-review item on the strength of having read nothing", scan)
		}
	}
}

// A clean page must NOT be reported as still holding: the sweep exists to retract
// findings that stopped being true, and a revalidator that can only say
// still_holds is the armed-but-inert shape this lane keeps finding.
func TestUnverifiedClaimsActuallyRetracts(t *testing.T) {
	got := unverifiedClaimsVerdict("p1", &checks.ClaimsPageScan{PageID: "p1", PageName: "about", ComponentsExamined: 5})
	if got.Verdict != revalidationResolved {
		t.Fatalf("a fully-read, claim-free page returned %q; if this cannot resolve, registering it "+
			"only adds scan cost and the type stays parked for ever", got.Verdict)
	}
	if got.Evidence["components_examined"] != 5 {
		t.Errorf("evidence should record how much was read, got %v — the count is what makes the "+
			"close auditable after the fact", got.Evidence["components_examined"])
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
		ItemType: "claims_unverified",
		ItemKey:  "claims:site_components",
		Spec:     map[string]interface{}{"check": "unverified_claims", "surface": "site_components"},
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
	got := unverifiedClaimsVerdict("p1", &checks.ClaimsPageScan{
		PageID: "p1", PageName: "about", ComponentsExamined: 2,
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
