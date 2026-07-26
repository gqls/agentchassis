// FILE: platform/orchestration/actions/truncation_guard.go
//
// Polices the CONSUMER half of the tolerate-a-truncation contract (bugs_open/076).
//
// ExecuteLLMPromptAction may keep a response that the model cut at max_tokens,
// stamping __truncated beside the result instead of failing the step. That trade
// is only sound while a consumer READS the marker: the step now succeeds, so
// without a reader nothing downstream can tell a complete answer from a fragment,
// and bugs_closed/019 would have traded a loud void for a silent half-answer.
//
// Tolerance was already opt-in per step (tolerate_truncation, default false), so
// the PRODUCER half fails closed correctly. The consumer half was enforced by
// nothing at all — every guard to date was hand-built at whichever call site had
// just been burned (diagnose_council_decide after bugs_closed/019,
// verify_report_prose after the gripper dossier). A step could therefore opt into
// tolerance with no guarded consumer anywhere in its workflow and ship a
// confident fragment, and no test, seed check or report would say so.
//
// This file makes that claim checkable in one place.
package actions

import (
	"sort"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// truncationAwareActions names every action whose Go code reads the __truncated
// marker and degrades on it. An entry here is a CLAIM ABOUT THAT CODE, not a
// label: adding a name without a reader silently re-opens bugs_open/076 for
// every workflow containing that action. TestTruncationAwareActionsReadTheMarker
// holds the two in lockstep by scanning the package for the reader.
//
// The value is the mechanism, so a reviewer can check the claim without grepping.
var truncationAwareActions = map[string]string{
	"diagnose_council_decide": "markerFieldFor -> review.Degraded; an approve cannot stand beside an unreadable seat (bugs_closed/019)",
	"verify_report_prose":     "truncationMarkerField -> errors \"refusing to publish a fragment\" before the page is written (gripper dossier, 8e8b55818)",
}

// acceptsTruncatedConfigKey lets a step declare in CONFIG that its action handles
// a partial input, for actions generic enough that the Go registry above cannot
// speak for them. It is deliberately a per-step declaration rather than a second
// registry: two hand-maintained lists that must agree is the drift class this
// platform keeps paying for (see the council gate's 099 roster mirror).
const acceptsTruncatedConfigKey = "accepts_truncated"

// findTruncationAwareConsumer reports whether this workflow contains any step
// that would read a __truncated marker, and names the first such step.
//
// LIMIT, stated plainly because it bounds what a pass here proves: this asks
// whether the workflow has A guarded consumer, NOT whether the truncated field
// reaches it. Deciding the latter needs reference-detection across prompt
// templates and config paths, and measured against live config on 2026-07-26 a
// textual scan matched 28 "consumers" in the council agents alone, nearly all of
// them spurious — a gate built on it would fail closed on sound workflows.
//
// So this is a FLOOR, not a proof: it cannot confirm the fragment is guarded, but
// it does refuse the case the bug is actually about — a step opting into
// tolerance in a workflow where nothing anywhere reads the marker.
func findTruncationAwareConsumer(steps map[string]models.Step, producingSteps ...string) (string, bool) {
	skip := make(map[string]bool, len(producingSteps))
	for _, name := range producingSteps {
		if name != "" {
			skip[name] = true
		}
	}

	// Sorted so the named consumer is stable across runs: this string lands in
	// logs that get compared between orchestrations.
	names := make([]string, 0, len(steps))
	for name := range steps {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		// A step cannot certify its own truncation, so the producer is excluded
		// even if it carries accepts_truncated.
		if skip[name] {
			continue
		}
		step := steps[name]
		if _, aware := truncationAwareActions[step.Action]; aware {
			return name, true
		}
		if datahelpers.GetBoolField(step.Config, acceptsTruncatedConfigKey, false) {
			return name, true
		}
	}
	return "", false
}
