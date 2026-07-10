package actions

import (
	"strings"
	"testing"
)

// The structural gate between the LLM's plan and the artifact F1.1b will act
// on. Shape only — correctness belongs to the council (F2) and the human PR.
func TestValidateFixPlan(t *testing.T) {
	good := fixPlan{
		Summary: "add the missing mark_no_sections step and ground nav in build_status",
		Edits: []fixPlanEdit{
			{File: "platform/orchestration/actions/populate_nav_tables_action.go",
				Symbol: "loadPagesForNav", Operation: "modify",
				Rationale: "nav selects on status, never build_status (diagnosis citation 4)",
				Sketch:    "add AND build_status = 'deployed' to the WHERE clause"},
			{File: "agent_definitions:page-build-handler", Symbol: "check_has_ready_sections",
				Operation: "config_change",
				Rationale: "sectionless pages route to a success terminal",
				Sketch:    "else_step -> mark_no_sections (fail_work_item, needs_human_review)"},
		},
		GroundedIn: []string{"Content writer skipped — page has no sections defined"},
	}
	if p := validateFixPlan(good, 8); len(p) != 0 {
		t.Fatalf("good plan must validate, got: %v", p)
	}

	cases := []struct {
		name   string
		mutate func(fixPlan) fixPlan
		expect string
	}{
		{"no edits", func(p fixPlan) fixPlan { p.Edits = nil; return p }, "no edits"},
		{"ungrounded", func(p fixPlan) fixPlan { p.GroundedIn = nil; return p }, "grounded_in is empty"},
		{"path traversal", func(p fixPlan) fixPlan { p.Edits[0].File = "../../etc/passwd"; return p }, "traversal"},
		{"absolute path", func(p fixPlan) fixPlan { p.Edits[0].File = "/etc/passwd"; return p }, "traversal"},
		{"rogue operation", func(p fixPlan) fixPlan { p.Edits[0].Operation = "execute"; return p }, "allowlist"},
		{"edit cap", func(p fixPlan) fixPlan {
			for i := 0; i < 9; i++ {
				p.Edits = append(p.Edits, p.Edits[0])
			}
			return p
		}, "architecture change"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// deep-ish copy: reslice edits so mutations don't leak between cases
			cp := good
			cp.Edits = append([]fixPlanEdit{}, good.Edits...)
			cp.GroundedIn = append([]string{}, good.GroundedIn...)
			problems := validateFixPlan(tc.mutate(cp), 8)
			if len(problems) == 0 {
				t.Fatal("expected validation failure")
			}
			if !strings.Contains(strings.Join(problems, "; "), tc.expect) {
				t.Fatalf("want %q named, got %v", tc.expect, problems)
			}
		})
	}
}
