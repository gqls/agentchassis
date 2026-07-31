// FILE: platform/orchestration/actions/discovery_checks/handler_coverage_test.go
//
// Build-enforced guard: a discovery check may not route work at a handler agent
// that does not exist.
//
// WHY THIS EXISTS
// ---------------
// bugs_open/077 is about a detector whose predicate is wider than its handler's
// remit. Its degenerate case is worse and was found while fixing it: a detector
// routing at a handler with NO remit at all, because the agent was never
// registered. `claim_work_item_action.go:126-168` marks such an item 'blocked'
// with "Handler agent not registered" — after it has occupied a dedup slot,
// counted in the backlog, and (at check_component_standards' priority 3) sorted
// near the front of the queue.
//
// Two were live in this package on 2026-07-26, and NEITHER was noticed by anything:
//
//	forced-text-color-fixer   check_forced_text_colors      1 live match waiting
//	site-metadata-fixer       check_component_standards     0 live matches, priority 3
//
// Both are now filed as capability_gap items instead (remit.go). This guard is
// what stops the third one being written. It is deliberately modelled on
// verifier_coverage_test.go — same package, same two-halves shape, same
// growth-only refresh rule — because a second, differently-shaped list is the
// drift this whole class is about.
//
// THE SENSOR / RATCHET SPLIT, and its honest limit
// -----------------------------------------------
// The SENSOR reads this package's source, so it needs no refresh: a check written
// tomorrow with a new HandlerAgent literal fails the build immediately. The
// RATCHET is knownHandlerAgents, a hand-refreshed snapshot of live
// agent_definitions — it can go stale in one direction only, by an agent being
// DELETED after being listed here. That direction is not silent: the item goes
// 'blocked' at claim with an explicit error, which is the loud failure the
// deleted-agent case deserves and the unwritten-agent case never got.
//
// Refresh by UNION, never replacement (the rule liveItemTypes learned the hard
// way — a type can vanish from a live query while its producer is still deployed):
//
//	SELECT DISTINCT type FROM agent_definitions WHERE deleted_at IS NULL ORDER BY 1;

package discovery_checks

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// handlerAgentLiteralRe matches `HandlerAgent: "foo"` in a check's WorkItemSpec.
// The empty string is matched too and handled explicitly below — "" is the
// canonical spelling of a deliberate no-handler route (needs_human_review items,
// migration 217), not an omission.
var handlerAgentLiteralRe = regexp.MustCompile(`HandlerAgent:\s*"([a-z0-9-]*)"`)

// handlerAgentComputedRe matches `HandlerAgent: <expression>` — a route this test
// cannot read from source. There are none today; a new one must be declared in
// computedHandlerAgentSites or the guard has a silent hole where it claims cover.
var handlerAgentComputedRe = regexp.MustCompile(`HandlerAgent:\s*([^"\s][^,\n]*)`)

// handlerAgentConstRe matches a package-level string constant a check then uses
// as its HandlerAgent, e.g. `navLinkHandlerAgent = "nav-link-fixer"`. Resolving
// these keeps a check honest without forcing it to repeat a literal it also needs
// elsewhere — the alternative is two copies of one agent name, which is the drift
// this file exists to prevent, in miniature.
var handlerAgentConstRe = regexp.MustCompile(`(?m)^\s*(\w+)\s*=\s*"([a-z0-9-]+)"`)

// handlerAgentIdentRe matches `HandlerAgent: someIdentifier` — resolvable against
// the constants above, or else a genuine runtime route.
var handlerAgentIdentRe = regexp.MustCompile(`HandlerAgent:\s*([A-Za-z_]\w*)\s*,`)

// computedHandlerAgentSites are check files that pick handler_agent at RUNTIME
// from a routing table, so no source scan can say which route a given finding
// takes. Each must be named here or the guard would claim cover it does not have.
//
// Both entries route from a function whose returns are string LITERALS in the
// same file, so the literal scan picks the agent names up anyway and classifies
// them — adding a third destination to either therefore still fails this guard
// until it is classified, which is the behaviour we want. (Same reasoning as
// computedItemTypeSites' note on check_directory.go.)
var computedHandlerAgentSites = map[string]string{
	"check_phantom_internal_links.go":   "routeBySurface() — per-surface routing table (page-build-handler / rerender-pages / …)",
	"check_unfulfilled_image_prompt.go": "classifyPromptKey() — per-prompt-purpose routing table (image-build-handler / asset-deployer)",
}

// knownHandlerAgents is every agent type a check may route at. Snapshot of live
// agent_definitions taken 2026-07-26, filtered to the types this package
// actually names — listing all ~170 fleet agents here would make the guard's
// denominator meaningless and hide the thing it measures.
var knownHandlerAgents = map[string]bool{
	"asset-deployer":            true,
	"blog-content-planner":      true,
	"color-variable-fixer":      true,
	"component-creator":         true,
	"component-template-fixer":  true,
	"content-feed-orchestrator": true,
	"content-gap-planner":       true,
	"deduplicate-sections":      true, // seeded 2026-07-31 by 269_*.sql; verified live BEFORE this line

	"image-build-handler":     true,
	"internal-linker":         true,
	"nav-link-fixer":          true,
	"nav-updater":             true,
	"page-build-handler":      true,
	"page-content-writer":     true,
	"page-rerender":           true,
	"rerender-pages":          true,
	"site-component-linker":   true,
	"site-design-planner":     true,
	"tool-acceptance-agent":   true,
	"tool-auditor":            true,
	"tool-improver":           true,
	"tool-recreation-handler": true,
	"tool-suggester":          true,
	"webdesign-agent":         true,
}

// handlerGap describes a handler a check WANTS and the platform does not have.
//
// Being in this map is not an excuse — it is a promise that the check does not
// dispatch to the missing agent. The test below enforces exactly that: a declared
// gap whose check still files a dispatchable item is a build failure, because a
// declaration that changes no behaviour is just a comment.
type handlerGap struct {
	// Why the agent does not exist and what it would take to make it.
	Why string
	// The check that wants it. Must file a capability_gap naming this agent as
	// builder_needed instead of routing work at it.
	Check string
}

// handlersNotYetBuilt is the declared-gap escape hatch. EMPTY TODAY — both
// instances that existed on 2026-07-26 were converted to capability_gap items
// rather than declared, because declaring a dead route leaves the dead route.
//
// It is kept, empty, because the honest case for an entry exists: a check written
// against an agent that is genuinely coming, in a branch, this week. If you add
// one, the entry is the thing that has to be deleted when the agent lands — so
// write Why as an instruction to your successor, not as an excuse to yourself.
var handlersNotYetBuilt = map[string]handlerGap{}

// TestEveryCheckHandlerAgentExistsOrIsADeclaredGap is the load-bearing assertion.
func TestEveryCheckHandlerAgentExistsOrIsADeclaredGap(t *testing.T) {
	seen, err := scanHandlerAgentLiterals(t)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	assertHandlerAgentsClassified(t, seen, knownHandlerAgents, handlersNotYetBuilt)
}

// scanHandlerAgentLiterals reads the package source and returns each non-empty
// HandlerAgent literal mapped to the file it came from.
func scanHandlerAgentLiterals(t *testing.T) (map[string]string, error) {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		return nil, err
	}

	seen := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		text := string(src)

		for _, m := range handlerAgentLiteralRe.FindAllStringSubmatch(text, -1) {
			if m[1] == "" {
				continue // deliberate no-handler route; see the regex comment
			}
			seen[m[1]] = f
		}

		// Constants declared in this file, so `HandlerAgent: someConst` resolves.
		consts := map[string]string{}
		for _, m := range handlerAgentConstRe.FindAllStringSubmatch(text, -1) {
			consts[m[1]] = m[2]
		}

		for _, m := range handlerAgentComputedRe.FindAllStringSubmatch(text, -1) {
			expr := strings.TrimSpace(m[1])
			if expr == "" {
				continue
			}
			// An identifier naming a constant in this file is not "computed" — it
			// is a literal one indirection away, and must be checked like one.
			if id := handlerAgentIdentRe.FindStringSubmatch("HandlerAgent: " + expr + ","); id != nil {
				if value, ok := consts[id[1]]; ok {
					seen[value] = f
					continue
				}
			}
			if _, declared := computedHandlerAgentSites[f]; !declared {
				t.Errorf("%s computes HandlerAgent at runtime (%s) but is not in computedHandlerAgentSites.\n"+
					"The source scan cannot see what it routes at, so this guard would claim cover it does not have.", f, expr)
			}
		}
	}
	return seen, nil
}

// assertHandlerAgentsClassified is the verdict, taking its inputs as parameters so
// the guard can be exercised against a deliberately broken world — see
// TestHandlerCoverageGuardFailsOnAnUnknownAgent. A guard that has never been
// observed to fail is not known to work: bugs_open/076's lockstep test validated
// every entry against a marker its own doc comment contained, and only a bogus
// entry exposed it.
func assertHandlerAgentsClassified(t *testing.T, seen map[string]string, known map[string]bool, gaps map[string]handlerGap) {
	t.Helper()

	if len(seen) == 0 {
		t.Fatal("scanned 0 handler agents — the regex or the layout changed, and this guard is now vacuous")
	}

	// A gap that is also a known agent means the list is lying.
	for agent, gap := range gaps {
		if known[agent] {
			t.Errorf("handler %q is declared not-yet-built (%s) but IS in knownHandlerAgents — delete the gap entry", agent, gap.Why)
		}
		if gap.Why == "" || gap.Check == "" {
			t.Errorf("handler gap %q must name the check that wants it and why it does not exist", agent)
		}
	}

	for agent, file := range seen {
		if known[agent] {
			continue
		}
		if gap, declared := gaps[agent]; declared {
			// A declared gap must not still be dispatched to. The check that wants
			// it files a capability_gap; routing work at it anyway is the defect
			// the declaration claims to have handled.
			t.Errorf("handler %q is a declared gap (%s) yet %s still routes work at it — "+
				"file a capability_gap naming it as builder_needed instead (see remit.go)", agent, gap.Why, file)
			continue
		}
		t.Errorf("check in %s routes work at handler agent %q, which is not a known agent "+
			"and is not a declared gap.\n"+
			"Every item it files will be marked 'blocked' at claim time "+
			"(claim_work_item_action.go:126-168) after occupying a dedup slot.\n"+
			"Either seed the agent and add it to knownHandlerAgents, or file a capability_gap "+
			"instead of routing at it (remit.go). This is how forced-text-color-fixer and "+
			"site-metadata-fixer went unnoticed.", file, agent)
	}
}

// TestHandlerCoverageGuardFailsOnAnUnknownAgent proves the guard is not vacuous by
// running it against a world it must reject. If this ever passes silently, the
// assertion above has stopped asserting.
func TestHandlerCoverageGuardFailsOnAnUnknownAgent(t *testing.T) {
	cases := []struct {
		name string
		seen map[string]string
		want string
	}{
		{
			name: "agent that does not exist anywhere",
			seen: map[string]string{"totally-invented-fixer": "check_bogus.go"},
			want: "not a known agent",
		},
		{
			name: "declared gap still being dispatched to",
			seen: map[string]string{"not-built-yet-fixer": "check_bogus.go"},
			want: "still routes work at it",
		},
		{
			name: "empty scan is vacuous, not clean",
			seen: map[string]string{},
			want: "vacuous",
		},
	}
	gaps := map[string]handlerGap{
		"not-built-yet-fixer": {Why: "deliberately fake, for this test", Check: "check_bogus"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			probe := &testing.T{}
			// t.Fatal in the vacuous case unwinds the goroutine it runs on, so the
			// probe must be driven on its own.
			done := make(chan struct{})
			go func() {
				defer close(done)
				assertHandlerAgentsClassified(probe, tc.seen, knownHandlerAgents, gaps)
			}()
			<-done
			if !probe.Failed() {
				t.Errorf("guard accepted %q — it is not asserting what it claims to", tc.name)
			}
		})
	}
}

// TestHandlerCoverageIsReported prints the picture so a reader sees the shape
// rather than only violations. Never fails.
func TestHandlerCoverageIsReported(t *testing.T) {
	seen, err := scanHandlerAgentLiterals(t)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	agents := make([]string, 0, len(seen))
	for a := range seen {
		agents = append(agents, a)
	}
	sort.Strings(agents)
	t.Logf("checks route at %d distinct handler agents: %v", len(agents), agents)
	t.Logf("declared not-yet-built handlers: %d", len(handlersNotYetBuilt))
}
