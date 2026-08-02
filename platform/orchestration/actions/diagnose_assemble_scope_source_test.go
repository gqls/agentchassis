// FILE: platform/orchestration/actions/diagnose_assemble_scope_source_test.go
//
// bugs_open/174 — a diagnosis aimed at chosen symbols silently ran against
// whatever the code search happened to return, because `seed_scope` was dropped
// in transit by diagnose-dispatch-loop's input_mapping allow-list. The scope
// fallback chain then did exactly what it was designed to do and supplied a
// different, plausible scope. No error, no warning, a normal-looking bundle.
//
// Two halves are under test here, and they fail for DIFFERENT reasons:
//
//   1. TYPE. Fixing the allow-list is not sufficient. `seed_scope` lives in the
//      work item's jsonb `spec`, and QueryDatabaseAction stringifies every []byte
//      a column scan returns — so the value arrives as the STRING `["a","b"]`.
//      ExtractStringListHelper used to return nil for a string, which means the
//      "fixed" path would have dropped the seed a third time, in a new place,
//      just as silently. These tests pin the string arm, so a future tidy-up of
//      the helper cannot re-open the hole.
//
//   2. PROVENANCE. The action cannot tell "no seed was given" from "the seed was
//      confiscated in transit" and must not pretend to — but it CAN say which
//      arm it took. `scope_source` is asserted on all three arms.
//
// Every test asserts the EFFECT (which symbols the verdicter actually sees), not
// merely that a field is set — 174's own verify section makes that distinction,
// because field-present and scope-used are exactly what the fallback chain pulls
// apart.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// scopeChainRun drives the action with the PRODUCTION field names, so the three
// arms of the fallback chain stay distinct: route.scope, then
// input_data.seed_scope, then code_lookup.code_results.
//
// It deliberately does NOT reuse bodyCapRun, which pins scope_field to "scope" —
// that collapses the loop-back arm and the seed arm onto one key, and a test run
// through it would pass no matter which of the two the action read.
func scopeChainRun(t *testing.T, collected map[string]interface{}) map[string]interface{} {
	t.Helper()
	out, err := DiagnoseAssembleBundleAction(context.Background(), ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig: models.Step{Config: map[string]interface{}{
			"max_body_chars": 60000,
			"persist_bundle": false, // no DB in a unit test; egress is not under test
		}},
	})
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", out)
	}
	return m
}

// seedFixture reuses bodyCapFixture's throwaway checkout (aaa_big.go,
// zzz_small.go) and adds a populated code_results, so the code-search fallback
// is ARMED in every test below. That is what makes "the seed was used" a real
// claim: without a live fallback to lose to, a passing seed test would also pass
// against the bug.
func seedFixture(t *testing.T) map[string]interface{} {
	t.Helper()
	collected, _, _ := bodyCapFixture(t, 3, 3)
	collected["code_lookup"] = map[string]interface{}{
		"code_results": []interface{}{
			map[string]interface{}{"path": "zzz_small.go", "symbol": "Target"},
		},
	}
	return collected
}

// runWithSeed drives the action with a seed scope of the given SHAPE, leaving
// "scope" (the loop-back arm) unset so the seed arm is the one under test.
func runWithSeed(t *testing.T, seed interface{}) map[string]interface{} {
	t.Helper()
	collected := seedFixture(t)
	collected["input_data"] = map[string]interface{}{"seed_scope": seed}
	return scopeChainRun(t, collected)
}

// THE BUG, in the shape production actually delivers it: a jsonb column that has
// been through QueryDatabaseAction is a STRING holding a JSON array.
//
// Pre-fix this returned nil from ExtractStringListHelper, the chain fell through
// to code_results, and the bundle rendered zzz_small.go — a real, plausible,
// entirely wrong scope. Proved by mutation: revert the string arm of
// ExtractStringListHelper and this test fails on scope_source ("code_results")
// and on the body assertion, in that order.
func TestScopeSource_SeedArrivingAsJSONStringIsUsed(t *testing.T) {
	res := runWithSeed(t, `["aaa_big.go:Target"]`)

	if got := res["scope_source"]; got != "seed" {
		t.Fatalf("scope_source = %v, want \"seed\" — a JSON-array STRING is the shape a jsonb spec key has after query_database stringifies it; returning nil here is bugs_open/174 re-opened", got)
	}
	section := inScopeSection(res["bundle"].(string))
	if !strings.Contains(section, "aaa_big.go:Target") {
		t.Errorf("in-scope code does not contain the SEEDED symbol; got:\n%s", section)
	}
	// The negative half. Field-present is not scope-used, and the fallback is what
	// makes them come apart — so assert the fallback's symbol is ABSENT.
	if strings.Contains(section, "zzz_small.go:Target") {
		t.Errorf("in-scope code contains the code_results FALLBACK symbol — the seed was recorded but not used:\n%s", section)
	}
}

// The already-decoded shapes must keep working. A []interface{} is what a seed
// passed straight through a Kafka envelope looks like.
func TestScopeSource_SeedAsDecodedListStillWorks(t *testing.T) {
	for name, seed := range map[string]interface{}{
		"[]interface{}": []interface{}{"aaa_big.go:Target"},
		"[]string":      []string{"aaa_big.go:Target"},
		"[]byte":        []byte(`["aaa_big.go:Target"]`),
	} {
		t.Run(name, func(t *testing.T) {
			res := runWithSeed(t, seed)
			if got := res["scope_source"]; got != "seed" {
				t.Fatalf("scope_source = %v, want \"seed\"", got)
			}
			if !strings.Contains(inScopeSection(res["bundle"].(string)), "aaa_big.go:Target") {
				t.Errorf("seeded symbol missing from in-scope code")
			}
		})
	}
}

// THE NEGATIVE CONTROL 174 asks for by name: an intake with no seed must still
// work and still fall through to code_results as it does today. A fix that made
// the seed arm greedy — matching an empty or malformed value — would break every
// diagnosis that legitimately passes no scope, which is most of them.
func TestScopeSource_NoSeedFallsThroughToCodeResults(t *testing.T) {
	for name, seed := range map[string]interface{}{
		"absent":             nil,
		"empty JSON array":   `[]`,
		"not a JSON array":   `aaa_big.go:Target`,
		"JSON object":        `{"path":"aaa_big.go"}`,
		"empty decoded list": []interface{}{},
	} {
		t.Run(name, func(t *testing.T) {
			collected := seedFixture(t)
			if seed != nil {
				collected["input_data"] = map[string]interface{}{"seed_scope": seed}
			}
			res := scopeChainRun(t, collected)

			if got := res["scope_source"]; got != "code_results" {
				t.Fatalf("scope_source = %v, want \"code_results\" — a value this helper cannot read as a list must fall through, not be half-honoured", got)
			}
			if !strings.Contains(inScopeSection(res["bundle"].(string)), "zzz_small.go:Target") {
				t.Errorf("code_results fallback did not supply the scope")
			}
		})
	}
}

// The loop-back arm keeps priority over the seed. This is the arm that carries a
// revised scope on iteration 2+, and a fix that let a stale seed outrank it would
// pin the loop to iteration 1's scope for ever.
func TestScopeSource_LoopScopeOutranksSeed(t *testing.T) {
	collected := seedFixture(t)
	collected["input_data"] = map[string]interface{}{"seed_scope": `["zzz_small.go:Target"]`}
	collected["route"] = map[string]interface{}{"scope": []interface{}{"aaa_big.go:Target"}}
	res := scopeChainRun(t, collected)

	if got := res["scope_source"]; got != "route" {
		t.Fatalf("scope_source = %v, want \"route\"", got)
	}
	if strings.Contains(inScopeSection(res["bundle"].(string)), "zzz_small.go:Target") {
		t.Errorf("the seed overrode the loop's revised scope")
	}
}

// The bundle note fires on the ambiguous arm and ONLY there.
//
// The second half is the byte-identity control this file inherits from 164: a
// diagnosis whose scope was chosen must produce the same bundle text as before
// this change, or every existing diagnosis's baseline has moved and no future
// comparison against an archived bundle means anything.
func TestScopeSource_BundleNoteOnlyOnTheAmbiguousArm(t *testing.T) {
	const marker = "**This scope was NOT chosen — it is the code-search fallback.**"

	fallback := scopeChainRun(t, seedFixture(t))
	if !strings.Contains(fallback["bundle"].(string), marker) {
		t.Errorf("code_results arm did not render the fallback note — the arm that cannot distinguish a missing seed from a confiscated one is the one that must say so")
	}

	seeded := runWithSeed(t, `["aaa_big.go:Target"]`)
	if strings.Contains(seeded["bundle"].(string), marker) {
		t.Errorf("seed arm rendered the fallback note; the scope WAS chosen")
	}
	routedCollected := seedFixture(t)
	routedCollected["route"] = map[string]interface{}{"scope": []interface{}{"aaa_big.go:Target"}}
	routed := scopeChainRun(t, routedCollected)
	if strings.Contains(routed["bundle"].(string), marker) {
		t.Errorf("route arm rendered the fallback note; the scope WAS chosen")
	}
}
