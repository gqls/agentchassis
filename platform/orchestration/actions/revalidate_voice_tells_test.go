// FILE: platform/orchestration/actions/revalidate_voice_tells_test.go
//
// The verdict ladder for voice_tells retraction. Every arm is exercised without
// a database, which is why voiceTellsVerdict is split from the query glue.
//
// The property under test is not "does it return the right string" but the one
// that makes a wrong answer expensive: `resolved` CLOSES a live human-review
// item, so it must be reachable ONLY from "components were read, and they were
// clean". Every state where the scan read nothing, or read only part of the
// page, has to come back non-terminal.
package actions

import (
	"context"
	"strings"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestVoiceTellsVerdictLadder(t *testing.T) {
	finding := func(name string) map[string]interface{} {
		return map[string]interface{}{"check": name}
	}

	cases := []struct {
		name        string
		scan        *checks.VoicePageScan
		want        string
		reasonHas   string
		explanation string
	}{
		{
			name:        "page absent from the re-scan",
			scan:        nil,
			want:        revalidationUnknown,
			reasonHas:   "not evidence the prose was fixed",
			explanation: "a deleted or unserved page is not proof the copy was rewritten",
		},
		{
			name:        "every component locked, nothing read",
			scan:        &checks.VoicePageScan{PageID: "p1", ComponentsExamined: 0, ComponentsSkippedLocked: 3},
			want:        revalidationUnknown,
			reasonHas:   "read nothing",
			explanation: "empty findings here means the scan examined nothing at all",
		},
		{
			name: "tells still present",
			scan: &checks.VoicePageScan{
				PageID: "p1", PageName: "services", ComponentsExamined: 4,
				Findings: []map[string]interface{}{finding("strawman"), finding("long_sentences")},
			},
			want:        revalidationStillHolds,
			reasonHas:   "still trips",
			explanation: "the finding is unchanged, so the item stays open",
		},
		{
			name: "tells present AND some components locked - still_holds wins",
			scan: &checks.VoicePageScan{
				PageID: "p1", PageName: "services", ComponentsExamined: 2, ComponentsSkippedLocked: 1,
				Findings: []map[string]interface{}{finding("strawman")},
			},
			want:        revalidationStillHolds,
			reasonHas:   "still trips",
			explanation: "ladder order: a page that still trips must not be reported as unreadable",
		},
		{
			name:        "clean, but part of the page was locked and unread",
			scan:        &checks.VoicePageScan{PageID: "p1", PageName: "services", ComponentsExamined: 2, ComponentsSkippedLocked: 1},
			want:        revalidationUnknown,
			reasonHas:   "were not read",
			explanation: "the reported tells may live in the pinned components, so closing would assert something unchecked",
		},
		{
			name:        "clean, whole page read",
			scan:        &checks.VoicePageScan{PageID: "p1", PageName: "services", ComponentsExamined: 3},
			want:        revalidationResolved,
			explanation: "the only state that licenses closing the item",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := voiceTellsVerdict("p1", tc.scan)
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
func TestVoiceTellsNeverResolvesWithoutReadingSomething(t *testing.T) {
	unreadable := []*checks.VoicePageScan{
		nil,
		{PageID: "p1"},
		{PageID: "p1", ComponentsSkippedLocked: 1},
		{PageID: "p1", ComponentsSkippedLocked: 9},
	}
	for _, scan := range unreadable {
		got := voiceTellsVerdict("p1", scan)
		if got.Verdict == revalidationResolved {
			t.Errorf("scan %+v produced `resolved` without examining a component — "+
				"that closes a live human-review item on the strength of having read nothing", scan)
		}
	}
}

// A clean page must NOT be reported as still holding: the sweep exists to retract
// findings that stopped being true, and a revalidator that can only say
// still_holds is the armed-but-inert shape this lane keeps finding.
func TestVoiceTellsActuallyRetracts(t *testing.T) {
	got := voiceTellsVerdict("p1", &checks.VoicePageScan{PageID: "p1", PageName: "services", ComponentsExamined: 5})
	if got.Verdict != revalidationResolved {
		t.Fatalf("a fully-read, tell-free page returned %q; if this cannot resolve, registering it "+
			"only adds scan cost and the type stays parked for ever", got.Verdict)
	}
	if got.Evidence["components_examined"] != 5 {
		t.Errorf("evidence should record how much was read, got %v — the count is what makes the "+
			"close auditable after the fact", got.Evidence["components_examined"])
	}
}

// The spec-missing arm returns before touching the database, so a nil *sql.DB is
// safe here and proves the early return really is early.
func TestVoiceTellsRefusesAnItemWithNoPageID(t *testing.T) {
	got := revalidateVoiceTells(context.Background(), nil, parkedReviewItem{
		ItemType: "voice_tells",
		ItemKey:  "voice:2c106994-b19a-4818-8a2b-4ef475ebd77f",
		Spec:     map[string]interface{}{"check": "voice_tells"},
	}, nil)
	if got.Verdict != revalidationUnknown {
		t.Errorf("verdict = %q, want %q", got.Verdict, revalidationUnknown)
	}
	// The item_key carries a parseable page id. Reading it would make the verdict
	// depend on a prefix convention rather than the field the producer writes —
	// the mistake §0b of this lane's handoff spent a session refuting.
	if !strings.Contains(got.Reason, "item_key") {
		t.Errorf("reason should say the item_key is deliberately not parsed, got %q", got.Reason)
	}
}
