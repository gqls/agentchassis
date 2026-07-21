package actions

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// The dormant-agents checker's pure functions are safety-relevant: an unstable
// item_key would flood site_work_items on every sweep, and an off-by-one age
// floor would flag freshly-seeded agents that simply have not run yet (the
// evidence-researcher false-positive the bug warns about). Test the real
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
	a := dormantAgent{Type: "nav-link-fixer", SampleStep: "fix_nav", UniqueSteps: 3, ActiveRows: 1, AgeDays: 145.2, FirstCreated: time.Unix(1740000000, 0)}
	s := dormantSpecJSON(a, 14)
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("spec not valid JSON: %v", err)
	}
	for _, k := range []string{"agent_type", "sample_step", "unique_steps", "age_days", "first_created", "method", "caveat", "source"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("spec missing key %q: %s", k, s)
		}
	}
	if m["agent_type"] != "nav-link-fixer" {
		t.Fatalf("agent_type wrong: %v", m["agent_type"])
	}
	// The method note must state owner_agent_type is NOT used — the trap this bug warns about.
	if ms, _ := m["method"].(string); !strings.Contains(ms, "owner_agent_type") {
		t.Fatalf("method note must name the owner_agent_type trap: %s", ms)
	}
}

// The report is the discoverability artifact — it must state the retention
// caveat (so a reader never treats "never observed" as "never ran ever") and
// must show every group so nothing is hidden.
func TestRenderDormantAgentsHonestAndComplete(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	oldest := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	stats := dormantStats{ActiveWithWorkflow: 155, Measurable: 123, BlindSpot: 32}
	past := []dormantAgent{{Type: "nav-link-fixer", SampleStep: "fix_nav", UniqueSteps: 3, ActiveRows: 1, AgeDays: 145}}
	under := []dormantAgent{{Type: "feature-implementer", SampleStep: "implement", UniqueSteps: 2, ActiveRows: 1, AgeDays: 3.9}}

	// dry run
	r := renderDormantAgents(now, 14, stats, past, under, nil, oldest, 0, 0, 0, 10, 0, true)
	for _, want := range []string{"DRY RUN", "nav-link-fixer", "feature-implementer", "never observed", "2026-05-28", "owner_agent_type"} {
		if !strings.Contains(r, want) {
			t.Fatalf("report missing %q:\n%s", want, r)
		}
	}
	if strings.Contains(r, "## Bookkeeping") {
		t.Fatal("dry run must not print live bookkeeping")
	}

	// live run
	r2 := renderDormantAgents(now, 14, stats, past, under, nil, oldest, 1, 0, 0, 10, 2, false)
	if !strings.Contains(r2, "## Bookkeeping") {
		t.Fatalf("live run must print bookkeeping:\n%s", r2)
	}
	if !strings.Contains(r2, "closed as resolved 2") {
		t.Fatalf("live run must report close-out count:\n%s", r2)
	}
}

// A type with more than one active row is the is_active-hygiene shadowing case;
// the report must surface it rather than silently collapse it.
func TestRenderDormantAgentsFlagsDuplicateActiveRows(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	past := []dormantAgent{{Type: "chief-strategist", SampleStep: "strategise", UniqueSteps: 1, ActiveRows: 2, AgeDays: 245}}
	r := renderDormantAgents(now, 14, dormantStats{}, past, nil, nil, time.Time{}, 0, 0, 0, 10, 0, true)
	if !strings.Contains(r, "2 active rows") {
		t.Fatalf("report must flag duplicate active rows:\n%s", r)
	}
}
