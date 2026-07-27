package agentenv

import "testing"

// TestProviderKeyEnvNeverCarriesALiteralValue pins the security property, not the
// contents. A SecretKeyRef means the credential is resolved by the kubelet and
// never enters the spawner's memory, its logs, or the pod spec it writes. A
// regression to `Value:` would still work at runtime — the agent would find its
// key — and would silently print every provider credential into any dump of a
// Job spec. That is the failure this test exists to catch, because it is the one
// that does not announce itself.
func TestProviderKeyEnvNeverCarriesALiteralValue(t *testing.T) {
	env := ProviderKeyEnv()
	if len(env) == 0 {
		t.Fatal("no provider keys at all — both spawners would produce agents that cannot call any model")
	}
	for _, e := range env {
		if e.Value != "" {
			t.Errorf("%s carries a literal value; provider keys must be SecretKeyRef only", e.Name)
		}
		if e.ValueFrom == nil || e.ValueFrom.SecretKeyRef == nil {
			t.Fatalf("%s has no SecretKeyRef", e.Name)
		}
		if got := e.ValueFrom.SecretKeyRef.Name; got != defaultSecretName {
			t.Errorf("%s reads Secret %q, want %q", e.Name, got, defaultSecretName)
		}
		// The env var name and the Secret key are identical by convention, and
		// an agent definition names a provider by its env var alone — so a
		// mismatch here resolves to an empty string at os.Getenv and fails at
		// client construction, which is exactly bugs_open/112's shape.
		if e.ValueFrom.SecretKeyRef.Key != e.Name {
			t.Errorf("%s reads Secret key %q; name and key must match", e.Name, e.ValueFrom.SecretKeyRef.Key)
		}
	}
}

// TestGeminiIsPresent is the bugs_open/112 regression guard. page-content-writer
// was switched to Gemini in the live database while no spawner granted
// GEMINI_API_KEY, so the writer failed at client construction inside its spawned
// pod and — generate_content having no error_step — took the whole page build
// with it, every section of every page.
//
// Named explicitly rather than asserted by count: a count test passes when
// someone swaps one provider for another, which is the same outage again.
func TestGeminiIsPresent(t *testing.T) {
	for _, n := range ProviderKeyNames() {
		if n == "GEMINI_API_KEY" {
			return
		}
	}
	t.Fatal("GEMINI_API_KEY missing from the spawned-pod allow-list — this is bugs_open/112 recurring")
}

// TestProviderKeyNamesIsACopy guards the single-source-of-truth property. Both
// spawners read this list; if a caller could mutate the returned slice it would
// change what the OTHER spawner grants, which is precisely the cross-spawner
// drift this package was created to end.
func TestProviderKeyNamesIsACopy(t *testing.T) {
	first := ProviderKeyNames()
	if len(first) == 0 {
		t.Fatal("empty allow-list")
	}
	original := first[0]
	first[0] = "MUTATED"

	if second := ProviderKeyNames(); second[0] != original {
		t.Fatalf("ProviderKeyNames leaks its backing array: a caller's write changed it to %q", second[0])
	}
}
