// FILE: platform/orchestration/actions/finding_code_roster_test.go
//
// bugs_open/358. ONE roster of agent_error_log finding codes, read from the
// registry, replacing the hand-maintained copies that used to live inside two
// individual tests.
//
// WHY. Two tests in this package each carried their own hard-coded list of
// "codes the estate already writes", to assert a new code collided with none of
// them:
//
//	discovery_checks_error_log_test.go       — nine codes
//	save_sections_content_data_links_test.go — eight codes, plus prefix-disjointness
//
// Both were correct when written and both had gone stale: neither contained
// RESOLVER_CONFLICTING_CANDIDATES (the largest population in the table),
// PLAN_SECTION_NAME_DROPPED, or CONTENT_KEY_LOSS. Two hand-maintained rosters
// that must stay identical is the drift class this estate keeps filing bugs
// about, and it is why 099_SYNC_gate_roster.py exists for the council seats.
//
// The tests' STATED REASONS are unchanged and still right — a shared code makes
// "which path caught this" unanswerable, and prefix-disjointness is a real
// property because the estate has live LIKE queries on
// `tool_crosslink_not_emitted%` and `component_validation_%`. Only the source of
// truth moves. The registry is checked against the LIVE TABLE daily
// (config-key-audit --finding-codes), so unlike a list in a test file it cannot
// quietly stop describing the estate.
package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const findingCodeRegistryRelPath = "docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json"

// findingCodeRoster is every code the registry declares. Read fresh from the
// file rather than embedded: a copy compiled into the test binary would be the
// third hand-maintained roster, which is the thing being retired.
func findingCodeRoster(t *testing.T) []string {
	t.Helper()
	path := filepath.Join(moduleRoot(t), findingCodeRegistryRelPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("finding-code registry unreadable at %s: %v\n"+
			"This roster is the single source of truth for code distinctness (bugs_open/358); "+
			"a missing file must fail loudly, never silently pass a collision.", path, err)
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("finding-code registry is not a JSON object: %v", err)
	}
	codes := make([]string, 0, len(entries))
	for code := range entries {
		if strings.HasPrefix(code, "_") {
			continue // "_doc"
		}
		codes = append(codes, code)
	}
	// VACUITY GUARD. An empty roster makes every distinctness assertion below
	// pass trivially — the precise failure this file exists to prevent, one
	// level up. The registry has carried dozens of codes since it was seeded, so
	// a near-empty read is an instrument failure, not a small estate.
	if len(codes) < 10 {
		t.Fatalf("finding-code registry yielded only %d codes — refusing to certify distinctness "+
			"against an empty roster, which would pass for any code whatsoever", len(codes))
	}
	return codes
}

// findingCodeRosterExcept is the roster minus the code under test, which is the
// form both callers want: "every OTHER code the estate writes".
func findingCodeRosterExcept(t *testing.T, self string) []string {
	t.Helper()
	all := findingCodeRoster(t)
	out := make([]string, 0, len(all))
	for _, c := range all {
		if c != self {
			out = append(out, c)
		}
	}
	return out
}

// The code under test must itself be declared. Without this, a code could pass
// every distinctness check below simply by not being in the registry — the
// roster would then be measuring the wrong population, and an unregistered code
// would look MORE compliant than a registered one.
func assertFindingCodeIsRegistered(t *testing.T, code string) {
	t.Helper()
	for _, c := range findingCodeRoster(t) {
		if c == code {
			return
		}
	}
	t.Errorf("error code %q is not declared in %s — every code this package writes needs a "+
		"disposition (consumed / instrumented / human-evidence / operational), or `unruled` if "+
		"the decision is genuinely open. bugs_open/358.", code, findingCodeRegistryRelPath)
}

// findingCodeDistinctnessProblems is the PURE predicate — the whole of the
// judgement, with no file access and no *testing.T. It is split out so its
// discrimination can be mutation-proved against a fixture roster
// (TestDistinctnessPredicateDiscriminates below) rather than by editing the
// shipped registry, which is a shared file several sessions have open at once
// (WRONG_CALLS.md 2026-08-22: a session mutated a shared file in place to prove
// a guard, and another session committed it mid-window).
//
// Prefix-disjointness is not a style rule. save_sections_content_data_links_test
// measured its justification: the estate has two live LIKE queries
// (`tool_crosslink_not_emitted%`, `component_validation_%`), so a code sharing a
// prefix with another is silently swept into someone else's population.
func findingCodeDistinctnessProblems(code string, roster []string) []string {
	var problems []string
	for _, other := range roster {
		if other == code {
			continue // the code's own registry entry is not a collision with itself
		}
		if strings.HasPrefix(code, other) || strings.HasPrefix(other, code) {
			problems = append(problems, "error code "+code+" shares a prefix with live code "+
				other+" — a LIKE query would catch both")
		}
	}
	return problems
}

// assertFindingCodeDistinct is the shared assertion both repointed tests now
// call: registered, AND distinct from every other code the estate writes.
func assertFindingCodeDistinct(t *testing.T, code string) {
	t.Helper()
	assertFindingCodeIsRegistered(t, code)
	for _, p := range findingCodeDistinctnessProblems(code, findingCodeRoster(t)) {
		t.Error(p)
	}
}

// The predicate must be able to come out BOTH ways, or repointing the two tests
// at it would have quietly replaced two stale-but-working checks with one that
// always passes — the exact shape bugs_open/358 is about.
func TestDistinctnessPredicateDiscriminates(t *testing.T) {
	// A real prefix relation, in the direction that bit the estate: a live
	// `component_validation_%` query would sweep both of these into one
	// population.
	if got := findingCodeDistinctnessProblems("component_validation",
		[]string{"component_validation_rejected", "UNKNOWN"}); len(got) != 1 {
		t.Fatalf("a code that is a prefix of a live one must be caught; got %v", got)
	}
	// ...and the other direction.
	if got := findingCodeDistinctnessProblems("component_validation_rejected_extra",
		[]string{"component_validation_rejected", "UNKNOWN"}); len(got) != 1 {
		t.Fatalf("a code that EXTENDS a live one must be caught too; got %v", got)
	}

	// THE CONTROL, and it is the load-bearing half. Codes that share a stem but
	// where neither is a prefix of the other are the estate's whole naming
	// convention — firing here would ban it, and a predicate that flags
	// everything is as useless as one that flags nothing.
	if got := findingCodeDistinctnessProblems("CONTENT_LINK_REPAIR_DETAIL",
		[]string{"CONTENT_LINK_REPAIR_SKIPPED", "CONTENT_DATA_LINK_AUDIT", "UNKNOWN"}); len(got) != 0 {
		t.Fatalf("sibling codes sharing a stem must PASS; got %v", got)
	}
	// A code compared against a roster that contains its own entry must not
	// report itself — the roster read from the registry always does contain it.
	if got := findingCodeDistinctnessProblems("CONTENT_DATA_LINK_AUDIT",
		[]string{"CONTENT_DATA_LINK_AUDIT", "UNKNOWN"}); len(got) != 0 {
		t.Fatalf("a code must not collide with its own registry entry; got %v", got)
	}
}

// The roster's own health, asserted once for all codes rather than once per new
// code against a frozen snapshot of the others — which is what the two
// hand-maintained lists could never do.
func TestFindingCodeRosterIsMutuallyDistinct(t *testing.T) {
	codes := findingCodeRoster(t)
	for i, a := range codes {
		for _, b := range codes[i+1:] {
			if a == b {
				t.Errorf("registry declares %q twice", a)
			}
			if strings.HasPrefix(b, a) {
				t.Errorf("%q is a prefix of %q — a LIKE query on either catches both", a, b)
			}
			if strings.HasPrefix(a, b) {
				t.Errorf("%q is a prefix of %q — a LIKE query on either catches both", b, a)
			}
		}
	}
}

// Every code this package declares as a constant must be in the registry. This
// is the SOURCE-SIDE EARLY WARNING, and it is deliberately NOT the guarantee:
// it can only see codes reachable from a constant in this package, so it is
// blind to the ~7 sites that pass a variable, to the positional-argument
// callers of LogActionError, to the pgx family in internal/agents/, to the SQL
// writers, and to anything arriving from agent_definitions config.
//
// The guarantee lives in `config-key-audit --finding-codes`, which reads
// SELECT DISTINCT error_code from the live table and is blind to none of those.
// Anything this test misses is caught there within a day of the code first
// firing. Its value is immediacy at commit time, not coverage — do not read a
// pass here as "every code is declared".
func TestPackageErrorCodeConstantsAreRegistered(t *testing.T) {
	for _, code := range []string{
		validationDetailErrorCode,
		linkRepairErrorCode,
		validationWarningErrorCode,
		contentDataEnvelopeErrorCode,
		discoveryCheckErrorCode,
		planRefusalErrorCode,
		sectionDedupErrorCode,
		contentDataLinkErrorCode,
		contentDataRegressionErrorCode,
		claimsFloorErrorCode,
		deployStampRefusedErrorCode,
	} {
		assertFindingCodeIsRegistered(t, code)
	}
}
