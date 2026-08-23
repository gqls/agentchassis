// FILE: platform/orchestration/actions/tool_cross_link_items_test.go
//
// bugs_open/029. The defect these tests stand against: a tool page's URL was
// CONSTRUCTED from the tool's function name at suggestion time, and matched no
// page on any of the three shapes this platform actually produces.
//
// The emitter itself needs a live DB (it reads pages/site_work_items), so what
// is unit-testable is the boundary around it: the shapes related_pages arrives
// in, and the build_status predicate that decides whether a link is safe to
// write now or must wait behind the page's build.

package actions

import "testing"

func TestRelatedPagesFromSpec(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want []string
	}{
		{
			// The shape it actually arrives in: spec jsonb decoded into
			// map[string]interface{} by the work item loader.
			name: "decoded jsonb array",
			in:   []interface{}{"services", "capabilities"},
			want: []string{"services", "capabilities"},
		},
		{
			name: "already a string slice",
			in:   []string{"services"},
			want: []string{"services"},
		},
		{
			name: "json-encoded string",
			in:   `["services","about"]`,
			want: []string{"services", "about"},
		},
		{
			// A suggestion with no related_pages is normal, not an error:
			// the emitter logs and does nothing.
			name: "absent",
			in:   nil,
			want: nil,
		},
		{
			name: "wrong type is not a panic",
			in:   42,
			want: nil,
		},
		{
			// Non-string members are dropped rather than stringified — a
			// number here would resolve against no page anyway.
			name: "mixed members",
			in:   []interface{}{"services", 7, "", "about"},
			want: []string{"services", "about"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := relatedPagesFromSpec(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v (len %d), want %v (len %d)", got, len(got), tc.want, len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("index %d: got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestToolPageLive(t *testing.T) {
	// The vocabulary is exactly these three (checked against pages.build_status
	// fleet-wide 2026-07-25: deployed 363, needs_rebuild 31, planned 26).
	// needs_rebuild counts as live: the page was deployed and is queued for a
	// refresh, so the link resolves today. planned does NOT: linking to it is
	// the bug.
	if !toolPageLive("deployed") {
		t.Error("deployed must count as live")
	}
	if !toolPageLive("needs_rebuild") {
		t.Error("needs_rebuild must count as live — the page is served while it waits")
	}
	if toolPageLive("planned") {
		t.Error("planned must NOT count as live — that is the 404 this bug is about")
	}
	if toolPageLive("") {
		t.Error("an unreadable build_status must not be treated as live")
	}
}

// TestCrossLinkEmitDecision pins Guard 2's decision table — bugs_open/353.
//
// This exists because the defect it stands against lived in a branch NO unit
// test could reach: the decision was inline in a DB-dependent function, so the
// only tests possible were of its inputs (`toolPageLive`, `relatedPagesFromSpec`
// above), and those passed throughout the 19 days the guard was silently
// withholding every new tool's cross-links. Pinning inputs is not pinning a
// guard; the decision is extracted so it can be CALLED.
//
// The two rows that matter are the last two: identical except for the caller's
// promise, and they must differ. If they ever agree, either the opt-in has been
// made the default (unsafe — the owner's 2026-08-02 ruling says the unsafe side
// defaults OFF) or it has become inert (the 353 defect is back).
func TestCrossLinkEmitDecision(t *testing.T) {
	cases := []struct {
		name                  string
		pageLive              bool
		gateItemFound         bool
		buildEnqueuedByCaller bool
		want                  crossLinkDecision
	}{
		{"a served page needs no gate at all", true, false, false, crossLinkEmitUngated},
		{"served, and the promise is irrelevant", true, false, true, crossLinkEmitUngated},
		{"not live but a build item exists: gate on it", false, true, false, crossLinkEmitGated},
		{"a gate item outranks the caller's promise — depends_on is stricter", false, true, true, crossLinkEmitGated},
		{"THE 353 CASE: not live, no gate item, no promise -> withhold", false, false, false, crossLinkWithhold},
		{"THE 353 FIX: not live, no gate item, caller owns the build -> emit", false, false, true, crossLinkEmitUngated},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := crossLinkEmitDecision(tc.pageLive, tc.gateItemFound, tc.buildEnqueuedByCaller); got != tc.want {
				t.Errorf("crossLinkEmitDecision(live=%v, gate=%v, promise=%v) = %v, want %v",
					tc.pageLive, tc.gateItemFound, tc.buildEnqueuedByCaller, got, tc.want)
			}
		})
	}
}

// TestCrossLinkOptInDefaultsToWithhold is the one assertion that cannot be made
// by the table above: that a caller which says NOTHING gets the safe branch. A
// zero-valued request is exactly what a new caller written by someone who never
// read this file produces, and it must not be granted the permissive arm.
func TestCrossLinkOptInDefaultsToWithhold(t *testing.T) {
	var req toolCrossLinkRequest // zero value: the forgetful caller
	if req.pageBuildIsEnqueuedByThisWorkflow {
		t.Fatal("the opt-in must default to false — the unsafe side is the default per the 2026-08-02 shared-seam ruling")
	}
	if got := crossLinkEmitDecision(false, false, req.pageBuildIsEnqueuedByThisWorkflow); got != crossLinkWithhold {
		t.Errorf("a zero-valued request on an unbuilt page must withhold, got %v", got)
	}
}
