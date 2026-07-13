package actions

import "testing"

// The iteration number stamped on each persisted bundle decides whether the
// partial unique index (correlation_id, iteration) WHERE kind='bundle' stores
// five rows for a five-iteration run or collapses them into one. The failure is
// silent — an upsert, not an error — so it is worth pinning down.
func TestAssembleIteration(t *testing.T) {
	// diagnose_route encodes LoopState as JSON, so the iteration arrives as a
	// float64 through collected_data, not an int. Cover both.
	routeState := func(iter interface{}) map[string]interface{} {
		return map[string]interface{}{
			"route": map[string]interface{}{
				"diagnose_state": map[string]interface{}{"iteration": iter},
			},
		}
	}

	tests := []struct {
		name      string
		collected map[string]interface{}
		config    map[string]interface{}
		want      int
	}{
		{
			// The genuine first pass: diagnose_route has not run, so no state
			// exists anywhere. Must be 1, not 0 — the CHECK rejects 0.
			name:      "first pass, no route state",
			collected: map[string]interface{}{},
			want:      1,
		},
		{
			name:      "second pass reads prior state (float64, as JSON decodes it)",
			collected: routeState(float64(1)),
			want:      2,
		},
		{
			name:      "fifth pass",
			collected: routeState(float64(4)),
			want:      5,
		},
		{
			name:      "int survives too",
			collected: routeState(4),
			want:      5,
		},
		{
			// THE TRAP diagnose_route documents: LoopState lands under the route
			// step's output_field, so a bare "diagnose_state" never resolves. A
			// config pointing there yields 1 on every pass, and every bundle
			// upserts over iteration 1. Assert the misconfiguration's shape so
			// nobody "simplifies" the default to the bare name.
			name:      "bare diagnose_state does not resolve (collapses to 1)",
			collected: map[string]interface{}{"diagnose_state": map[string]interface{}{"iteration": float64(3)}},
			want:      1,
		},
		{
			// ...and that the same state IS reachable when the field path is
			// correct, proving the previous case fails on the path, not the data.
			name:      "explicit iteration_field override is honoured",
			collected: map[string]interface{}{"diagnose_state": map[string]interface{}{"iteration": float64(3)}},
			config:    map[string]interface{}{"iteration_field": "diagnose_state.iteration"},
			want:      4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := assembleIteration(tc.collected, tc.config); got != tc.want {
				t.Errorf("assembleIteration() = %d, want %d", got, tc.want)
			}
		})
	}
}
