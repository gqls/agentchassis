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
