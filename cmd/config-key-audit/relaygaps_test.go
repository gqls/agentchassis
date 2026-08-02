// FILE: cmd/config-key-audit/relaygaps_test.go
//
// bugs_open/174. These fixtures are the live diagnose-dispatch-loop shape,
// reduced: a claim_item whose RETURNING projects some spec keys, a call_handler
// whose input_mapping forwards some of those, and a callee declaring an
// input_contract that names more than either.
//
// The pre-fix fixture is the config as it actually stood on 2026-08-02, so the
// first test is a real regression test rather than an invented one. It is the
// same proof run against the live snapshot: --relay-gaps exits 1 on the pre-fix
// export and 0 on the post-fix export, and the finding names both keys in BOTH
// categories.

package main

import "testing"

// The two allow-lists in series, as they were before migration 285: nine spec
// keys projected, ten keys forwarded, and a callee declaring eleven.
const preFixRelayExport = `[
	{"type": "diagnose-dispatch-loop", "workflow": {"start_step": "claim_item", "steps": {
		"claim_item": {"action": "query_database", "output_field": "claimed", "config": {
			"query": "UPDATE site_work_items SET status='diagnosing' WHERE id = (SELECT id FROM site_work_items LIMIT 1) RETURNING id::text AS work_item_id, handler_agent, spec->>'symptom' AS symptom, spec->>'owner' AS owner, spec->>'ref' AS ref, spec->>'site_id' AS target_site_id, spec->>'correlation_id' AS correlation_id"
		}},
		"call_handler": {"action": "call_agent", "config": {
			"target_role": "handler",
			"input_mapping": {
				"symptom": "claimed.symptom",
				"owner?": "claimed.owner",
				"ref?": "claimed.ref",
				"site_id?": "claimed.target_site_id",
				"correlation_id?": "claimed.correlation_id",
				"work_item_id": "claimed.work_item_id"
			}
		}}
	}}},
	{"type": "diagnose-orchestrator",
	 "input_contract": {"required": ["symptom"], "optional": ["owner","ref","site_id","correlation_id","seed_scope","runtime_page"]},
	 "workflow": {"start_step": "go", "steps": {"go": {"action": "complete_workflow"}}}}
]`

func decodeOrFail(t *testing.T, raw string) []contractAgent {
	t.Helper()
	agents, failed, err := decodeContractAgents([]byte(raw))
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}
	if failed != 0 {
		t.Fatalf("expected 0 undecodable rows, got %d", failed)
	}
	return agents
}

// THE BUG. seed_scope and runtime_page are declared by the callee, projected by
// nothing, and forwarded by nothing.
//
// not_projected is the half the 174 ticket missed. Its fix candidate 1 named
// only the input_mapping, which would have added an entry sourced from
// claimed.seed_scope — a path the RETURNING clause never produces — and the key
// being optional, ResolveInputMapping would have dropped it in silence a second
// time. Asserting the two lists separately is what makes that visible.
func TestFindRelayGaps_ReportsKeysNeitherProjectedNorForwarded(t *testing.T) {
	report := findRelayGaps(decodeOrFail(t, preFixRelayExport))

	if len(report.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(report.Findings), report.Findings)
	}
	f := report.Findings[0]
	if f.Caller != "diagnose-dispatch-loop" || f.Callee != "diagnose-orchestrator" {
		t.Errorf("wrong relay: %+v", f)
	}
	assertSet(t, "not_forwarded", f.NotForwarded, []string{"runtime_page", "seed_scope"})
	assertSet(t, "not_projected", f.NotProjected, []string{"runtime_page", "seed_scope"})

	// work_item_id is the loop's own bookkeeping — declared LoopInternal, and it
	// must not be reported as a gap in the other direction just because the
	// callee's contract does not name it.
	for _, k := range f.NotForwarded {
		if k == "work_item_id" {
			t.Errorf("work_item_id reported as a gap; it is declared LoopInternal")
		}
	}
	if len(report.Unmatched) != 0 {
		t.Errorf("registry entry did not match the fixture: %v", report.Unmatched)
	}
}

// A key wired in the input_mapping but sourced from a column the claim query
// never projects. It reads as fixed and resolves to nothing — precisely the
// state a mapping-only fix would have produced, which is why this is its own
// category rather than folded into "not forwarded".
func TestFindRelayGaps_ReportsAMappingThatResolvesToNothing(t *testing.T) {
	const halfFixed = `[
		{"type": "diagnose-dispatch-loop", "workflow": {"start_step": "claim_item", "steps": {
			"claim_item": {"action": "query_database", "output_field": "claimed", "config": {
				"query": "UPDATE x SET y=1 RETURNING id::text AS work_item_id, handler_agent, spec->>'symptom' AS symptom"
			}},
			"call_handler": {"action": "call_agent", "config": {"target_role": "handler", "input_mapping": {
				"symptom": "claimed.symptom",
				"seed_scope?": "claimed.seed_scope",
				"work_item_id": "claimed.work_item_id"
			}}}
		}}},
		{"type": "diagnose-orchestrator",
		 "input_contract": {"required": ["symptom"], "optional": ["seed_scope"]},
		 "workflow": {"start_step": "go", "steps": {"go": {"action": "complete_workflow"}}}}
	]`

	report := findRelayGaps(decodeOrFail(t, halfFixed))
	if len(report.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(report.Findings), report.Findings)
	}
	f := report.Findings[0]
	assertSet(t, "maps_to_nothing", f.MapsToNothing, []string{"seed_scope -> claimed.seed_scope"})
	// It IS forwarded — the mapping names it — so the gap is not "not forwarded".
	assertSet(t, "not_forwarded", f.NotForwarded, nil)
}

// The fixed shape: both lists carry the keys, and the check passes. Without this
// the tests above are satisfied by a detector that always fires.
func TestFindRelayGaps_PassesOnTheFixedShape(t *testing.T) {
	const fixed = `[
		{"type": "diagnose-dispatch-loop", "workflow": {"start_step": "claim_item", "steps": {
			"claim_item": {"action": "query_database", "output_field": "claimed", "config": {
				"query": "UPDATE x SET y=1 RETURNING id::text AS work_item_id, handler_agent, spec->>'symptom' AS symptom, spec->'seed_scope' AS seed_scope, spec->>'runtime_page' AS runtime_page"
			}},
			"call_handler": {"action": "call_agent", "config": {"target_role": "handler", "input_mapping": {
				"symptom": "claimed.symptom",
				"seed_scope?": "claimed.seed_scope",
				"runtime_page?": "claimed.runtime_page",
				"work_item_id": "claimed.work_item_id"
			}}}
		}}},
		{"type": "diagnose-orchestrator",
		 "input_contract": {"required": ["symptom"], "optional": ["seed_scope","runtime_page"]},
		 "workflow": {"start_step": "go", "steps": {"go": {"action": "complete_workflow"}}}}
	]`

	report := findRelayGaps(decodeOrFail(t, fixed))
	if len(report.Findings) != 0 {
		t.Fatalf("expected 0 findings on the fixed shape, got %+v", report.Findings)
	}
	if len(report.Unmatched) != 0 {
		t.Errorf("unexpected unmatched entries: %v", report.Unmatched)
	}
	// The registered relay must NOT also appear as uncovered — that would make
	// every clean run print a nag for the one thing that IS checked.
	for _, u := range report.Uncovered {
		if u.Caller == "diagnose-dispatch-loop" {
			t.Errorf("the registered relay was reported as uncovered: %+v", u)
		}
	}
}

// A registry entry that matches nothing is an assertion that silently stopped
// running — the same failure this tool exists to catch, one level up. It must be
// LOUD (it exits 1 alongside real findings), never a clean pass.
//
// This is not hypothetical: the first live run matched nothing, because
// validation.WalkSteps qualifies paths ("steps.call_handler") and the registry
// names steps as an author writes them. Only the unmatched list caught it.
func TestFindRelayGaps_UnmatchedRegistryEntryIsReported(t *testing.T) {
	const callerMissing = `[
		{"type": "diagnose-orchestrator",
		 "input_contract": {"required": ["symptom"], "optional": ["seed_scope"]},
		 "workflow": {"start_step": "go", "steps": {"go": {"action": "complete_workflow"}}}}
	]`
	report := findRelayGaps(decodeOrFail(t, callerMissing))
	if len(report.Unmatched) != 1 {
		t.Fatalf("expected the missing caller to be reported as unmatched, got %v", report.Unmatched)
	}
	if len(report.Findings) != 0 {
		t.Errorf("a relay that could not be checked must not also produce findings: %+v", report.Findings)
	}
}

// A callee with no input_contract is "nothing to check against", not "accepts
// nothing" — roughly half the live fleet has no contract, and treating that as
// an empty declaration would make every such relay pass vacuously.
func TestFindRelayGaps_CalleeWithoutAContractIsUnmatchedNotClean(t *testing.T) {
	const noContract = `[
		{"type": "diagnose-dispatch-loop", "workflow": {"start_step": "claim_item", "steps": {
			"claim_item": {"action": "query_database", "output_field": "claimed", "config": {"query": "UPDATE x SET y=1 RETURNING id::text AS work_item_id"}},
			"call_handler": {"action": "call_agent", "config": {"target_role": "handler", "input_mapping": {"work_item_id": "claimed.work_item_id"}}}
		}}},
		{"type": "diagnose-orchestrator", "workflow": {"start_step": "go", "steps": {"go": {"action": "complete_workflow"}}}}
	]`
	report := findRelayGaps(decodeOrFail(t, noContract))
	if len(report.Unmatched) != 1 {
		t.Fatalf("a callee with no contract must be UNMATCHED (nothing to check), got %v", report.Unmatched)
	}
}

// The discovery half. A dispatcher-shaped relay the registry does not name must
// be reported, so the registry cannot silently fall behind the fleet.
func TestFindRelayGaps_DiscoversAnUnregisteredDispatcher(t *testing.T) {
	const newLoop = `[
		{"type": "some-new-dispatch-loop", "workflow": {"start_step": "claim", "steps": {
			"claim": {"action": "query_database", "output_field": "claimed", "config": {"query": "UPDATE x SET y=1 RETURNING id::text AS work_item_id, spec->>'domain' AS domain"}},
			"call_handler": {"action": "call_agent", "config": {"target_role": "handler", "input_mapping": {
				"domain": "claimed.domain", "work_item_id": "claimed.work_item_id"
			}}}
		}}}
	]`
	report := findRelayGaps(decodeOrFail(t, newLoop))
	found := false
	for _, u := range report.Uncovered {
		if u.Caller == "some-new-dispatch-loop" {
			found = true
		}
	}
	if !found {
		t.Errorf("a new dispatcher-shaped relay was not discovered: %+v", report.Uncovered)
	}
}

// projectedAliases must read a bare RETURNING column as its own alias.
// `handler_agent` is exactly this in the live claim query, and treating it as
// unprojected would invent a permanent false finding on the real config.
func TestProjectedAliases_BareColumnCountsAsProjected(t *testing.T) {
	got := projectedAliases("UPDATE x SET y=1 RETURNING id::text AS work_item_id, handler_agent, spec->>'symptom' AS symptom")
	for _, want := range []string{"work_item_id", "handler_agent", "symptom"} {
		if !got[want] {
			t.Errorf("alias %q not detected in RETURNING clause; got %v", want, got)
		}
	}
	if got["site_work_items"] {
		t.Errorf("a table name before RETURNING leaked into the alias set: %v", got)
	}
}

func assertSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %v, want %v", label, got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: got %v, want %v", label, got, want)
			return
		}
	}
}
