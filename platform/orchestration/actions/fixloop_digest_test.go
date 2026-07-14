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
	out := renderDigest(24, time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC), nil, nil, digestImmune{}, nil)
	for _, want := range []string{
		"last 24h",
		"No fix-loop runs in this window.",
		"No diagnosis artifacts written in this window.",
		"Diagnosis queue empty — nothing is waiting on a dispatch decision.",
		"No silent-failure findings open or closed in this window.",
		"No standing capability gaps.",
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
	out := renderDigest(24, time.Date(2026, 7, 13, 13, 0, 0, 0, time.UTC), runs, corrs, digestImmune{}, snaps)

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

// Parked escalations are decisions waiting on the owner: the queue must show
// EVERY open item each digest (not just in-window ones), and closed silent
// findings must be visibly CLOSED, never silently dropped.
func TestRenderDigestEscalationChannel(t *testing.T) {
	immune := digestImmune{
		TriageSweeps: 2,
		SilentSweeps: 3,
		Escalations: []digestEscalation{
			{ItemKey: "triage-diag:silent_failure:fd86fec2c4da", Status: "awaiting_diagnosis",
				Summary: "silent failure on 2 site(s)", CreatedAt: "2026-07-14 16:00", New: true},
			{ItemKey: "triage-diag:needs_new_component:c4ad0be8a0f2", Status: "awaiting_diagnosis",
				Summary: "handler \"component-creator\" fails", CreatedAt: "2026-07-14 14:53", New: false},
		},
		SilentOpen: []digestSilentFinding{
			{ItemKey: "silent:nav_linked_never_built:5fe8785b", Summary: "4 page(s) on dartsonline.com"},
		},
		SilentClosed: []digestSilentFinding{
			{ItemKey: "silent:nav_linked_never_built:1244516d", Summary: "2 page(s) on idea.uk"},
		},
		Gaps: []capabilityGap{{Builder: "shop-builder", Count: 3, Sites: 2}},
	}
	out := renderDigest(24, time.Date(2026, 7, 14, 17, 0, 0, 0, time.UTC), nil, nil, immune, nil)

	for _, want := range []string{
		"Escalation channel — immune system → loop (2 open, 1 new)",
		"Sweeps this window: 2 triage, 3 silent-check",
		"NEW 2026-07-14 16:00  `triage-diag:silent_failure:fd86fec2c4da` [awaiting_diagnosis]",
		"- 2026-07-14 14:53  `triage-diag:needs_new_component:c4ad0be8a0f2`",
		"(1 open, 1 closed this window)",
		"OPEN `silent:nav_linked_never_built:5fe8785b`",
		"CLOSED `silent:nav_linked_never_built:1244516d`",
		"**shop-builder** needed — 3 item(s) across 2 site(s).",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("digest missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Diagnosis queue empty") {
		t.Fatal("non-empty queue must not render the empty-state line")
	}
}

func TestPqStringArray(t *testing.T) {
	got := pqStringArray([]string{"a", "b-c"})
	if got != `{"a","b-c"}` {
		t.Fatalf("pqStringArray: %s", got)
	}
}
