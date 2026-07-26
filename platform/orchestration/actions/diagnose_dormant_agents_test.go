package actions

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// The dormant-agents checker's pure functions are safety-relevant: an unstable
// item_key would flood site_work_items on every sweep, and an off-by-one age
// floor would flag freshly-seeded agents that have not run yet. Test the real
// functions.

func TestDormantItemKeyStableAndPerAgent(t *testing.T) {
	if dormantItemKey("feature-implementer") != dormantItemKey("feature-implementer") {
		t.Fatal("same agent must yield the same key")
	}
	if dormantItemKey("feature-implementer") == dormantItemKey("nav-link-fixer") {
		t.Fatal("different agents must yield different keys")
	}
	if !strings.HasPrefix(dormantItemKey("nav-link-fixer"), "dormant:") {
		t.Fatalf("key not readable/prefixed: %s", dormantItemKey("nav-link-fixer"))
	}
}

// The age floor is the guard against flagging a fresh seed. Boundary: an agent
// exactly at the floor is PAST it (eligible); one hair under is not.
func TestDormantPartitionAgeFloorBoundary(t *testing.T) {
	agents := []dormantAgent{
		{Type: "brand-new", AgeDays: 1.0},
		{Type: "just-under", AgeDays: 13.9},
		{Type: "exactly-floor", AgeDays: 14.0},
		{Type: "legacy", AgeDays: 300.0},
	}
	past, under := dormantPartition(agents, 14.0)
	if len(past) != 2 {
		t.Fatalf("expected 2 past-floor (exactly-floor, legacy), got %d: %+v", len(past), past)
	}
	if len(under) != 2 {
		t.Fatalf("expected 2 under-floor (brand-new, just-under), got %d: %+v", len(under), under)
	}
	pastTypes := map[string]bool{}
	for _, a := range past {
		pastTypes[a.Type] = true
	}
	if !pastTypes["exactly-floor"] {
		t.Fatal("an agent exactly at the floor must be eligible to emit (>=, not >)")
	}
	if pastTypes["just-under"] {
		t.Fatal("an agent under the floor must not be emitted")
	}
}

// A cap must surface the freshest genuinely-unused capabilities first, not bury
// them under a decade of legacy rows.
func TestDormantEmitOrderYoungestFirst(t *testing.T) {
	past := []dormantAgent{
		{Type: "legacy", AgeDays: 300},
		{Type: "recent", AgeDays: 20},
		{Type: "older", AgeDays: 120},
	}
	ordered := dormantEmitOrder(past)
	if ordered[0].Type != "recent" || ordered[2].Type != "legacy" {
		t.Fatalf("expected youngest-first (recent, older, legacy); got %s, %s, %s",
			ordered[0].Type, ordered[1].Type, ordered[2].Type)
	}
	// must not mutate the input
	if past[0].Type != "legacy" {
		t.Fatal("dormantEmitOrder must not mutate its input slice")
	}
}

func TestDormantSpecJSONShape(t *testing.T) {
	a := dormantAgent{Type: "nav-link-fixer", ActiveRows: 1, AgeDays: 145.2, FirstCreated: time.Unix(1740000000, 0)}
	s := dormantSpecJSON(a, 14)
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("spec not valid JSON: %v", err)
	}
	for _, k := range []string{"agent_type", "active_rows", "age_days", "first_created", "method", "caveat", "source"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("spec missing key %q: %s", k, s)
		}
	}
	if m["agent_type"] != "nav-link-fixer" {
		t.Fatalf("agent_type wrong: %v", m["agent_type"])
	}
	// The method note must name the durable substrate and disclaim owner_agent_type / orchestration_states.
	ms, _ := m["method"].(string)
	if !strings.Contains(ms, "agent_run_stats") || !strings.Contains(ms, "owner_agent_type") {
		t.Fatalf("method note must name agent_run_stats and disclaim owner_agent_type: %s", ms)
	}
}

// The report is the discoverability artifact — it must state the tracking window
// (so a reader never treats "never run" as "never ran ever") and show every
// group so nothing is hidden.
func TestRenderDormantAgentsHonestAndComplete(t *testing.T) {
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	trackingSince := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC) // ~27d window
	stats := dormantStats{ActiveWithWorkflow: 160, Ran: 130}
	past := []dormantAgent{{Type: "nav-link-fixer", ActiveRows: 1, AgeDays: 145}}
	under := []dormantAgent{{Type: "feature-implementer", ActiveRows: 1, AgeDays: 3.9}}

	// dry run
	r := renderDormantAgents(now, 14, stats, past, under, trackingSince, 27.0, true, 0, 0, 0, 10, 0, true)
	for _, want := range []string{"DRY RUN", "nav-link-fixer", "feature-implementer", "never run", "agent_run_stats", "2026-07-24"} {
		if !strings.Contains(r, want) {
			t.Fatalf("report missing %q:\n%s", want, r)
		}
	}
	if strings.Contains(r, "## Bookkeeping") {
		t.Fatal("dry run must not print live bookkeeping")
	}

	// live run, window sufficient
	r2 := renderDormantAgents(now, 14, stats, past, under, trackingSince, 27.0, true, 1, 0, 0, 10, 2, false)
	if !strings.Contains(r2, "## Bookkeeping") {
		t.Fatalf("live run must print bookkeeping:\n%s", r2)
	}
	if !strings.Contains(r2, "closed as resolved 2") {
		t.Fatalf("live run must report close-out count:\n%s", r2)
	}
}

// The window guard is the correctness fix for the forward-only cold start: a
// live sweep whose tracking window is shorter than the age floor must emit
// NOTHING and say so loudly, so flipping dry_run off right after deploy cannot
// flood false positives.
func TestRenderDormantAgentsWindowTooShortBanner(t *testing.T) {
	now := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	trackingSince := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC) // ~2d window
	past := []dormantAgent{{Type: "fix-proposer", ActiveRows: 1, AgeDays: 145}}

	// live sweep, window (2d) < floor (14d): must show the banner and emit nothing.
	r := renderDormantAgents(now, 14, dormantStats{}, past, nil, trackingSince, 2.0, false, 0, 0, 0, 10, 0, false)
	if !strings.Contains(r, "WINDOW TOO SHORT") {
		t.Fatalf("live sweep with insufficient window must show the guard banner:\n%s", r)
	}
	if !strings.Contains(r, "agent_run_stats") {
		t.Fatal("banner must name the durable substrate (agent_run_stats)")
	}
	// window sufficient: no banner.
	r2 := renderDormantAgents(now, 14, dormantStats{}, past, nil, trackingSince, 30.0, true, 0, 0, 0, 10, 0, false)
	if strings.Contains(r2, "WINDOW TOO SHORT") {
		t.Fatal("a sufficient window must NOT show the too-short banner")
	}
}

// The empty-table cold start must be reported as such, not as a fleet outage.
func TestRenderDormantAgentsEmptyTrackingTable(t *testing.T) {
	now := time.Date(2026, 7, 24, 15, 0, 0, 0, time.UTC)
	past := []dormantAgent{{Type: "fix-proposer", ActiveRows: 1, AgeDays: 145}}
	r := renderDormantAgents(now, 14, dormantStats{ActiveWithWorkflow: 160, Ran: 0}, past, nil, time.Time{}, 0.0, false, 0, 0, 0, 10, 0, true)
	if !strings.Contains(r, "No runs recorded yet") {
		t.Fatalf("empty tracking table must be reported as the cold start:\n%s", r)
	}
}

// A type with more than one active row is the is_active-hygiene shadowing case;
// the report must surface it rather than silently collapse it.
func TestRenderDormantAgentsFlagsDuplicateActiveRows(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	trackingSince := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	past := []dormantAgent{{Type: "chief-strategist", ActiveRows: 2, AgeDays: 245}}
	r := renderDormantAgents(now, 14, dormantStats{}, past, nil, trackingSince, 53.0, true, 0, 0, 0, 10, 0, true)
	if !strings.Contains(r, "2 active rows") {
		t.Fatalf("report must flag duplicate active rows:\n%s", r)
	}
}
