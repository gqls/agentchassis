package actions

import (
	"strings"
	"testing"
	"time"
)

// The awareness surface must be deterministic and honest: an empty section
// must read as "no activity", never as "not checked". renderDigest is the
// real rendering path — tested directly.

func TestRenderDigestEmptyWindowIsExplicit(t *testing.T) {
	out := renderDigest(24, time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC), nil, nil, nil)
	for _, want := range []string{
		"last 24h",
		"No fix-loop runs in this window.",
		"No diagnosis artifacts written in this window.",
		"No agent definitions were changed in this window.",
		"no model in this path",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("digest missing %q:\n%s", want, out)
		}
	}
}

func TestRenderDigestSections(t *testing.T) {
	runs := []digestRun{
		{AgentType: "fix-implementer", OrchID: "d05d126e-22ff-4853-987c-606ca4ce6269",
			Status: "COMPLETED", Terminal: "complete", CreatedAt: "2026-07-13 12:11",
			GateNote: "gate PASS", PRNote: "PR https://github.com/gqls/agentchassis/pull/1"},
		{AgentType: "fix-proposer", OrchID: "df4ffbb6-aaaa-bbbb-cccc-000000000000",
			Status: "COMPLETED", Terminal: "complete", CreatedAt: "2026-07-12 19:43"},
	}
	corrs := []digestCorrelation{
		{CorrelationID: "11111111-e2e2-4a1b-9c3d-000000000001",
			Kinds: "council_report:1, fix_plan:1", Decision: "approved", DecidedBy: "all reviewers approve"},
	}
	snaps := []digestSnapshot{
		{AgentType: "fix-implementer", Reason: "pre-update: raise sketch_to_files max_tokens", TakenAt: "2026-07-13 10:02"},
	}
	out := renderDigest(24, time.Date(2026, 7, 13, 13, 0, 0, 0, time.UTC), runs, corrs, snaps)

	for _, want := range []string{
		"Runs (2)",
		"gate PASS",
		"PR https://github.com/gqls/agentchassis/pull/1",
		"`11111111`: council_report:1, fix_plan:1 — latest council: **approved** (all reviewers approve)",
		"Agent config changes (1)",
		"raise sketch_to_files max_tokens",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("digest missing %q:\n%s", want, out)
		}
	}
}

func TestPqStringArray(t *testing.T) {
	got := pqStringArray([]string{"a", "b-c"})
	if got != `{"a","b-c"}` {
		t.Fatalf("pqStringArray: %s", got)
	}
}
