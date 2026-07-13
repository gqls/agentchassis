// FILE: internal/adapters/thunder/provision_action_test.go
//
// Unit tests for the pure helpers in provision_action.go.
//
// Full happy-path / error-path tests of Execute would require mocking
// thunderAPI, secretManager, and *sql.DB (via sqlmock or similar). That's
// a substantial test suite — deferring to a follow-up file once we've
// confirmed the action works in production against the real Thunder API.

package thunder

import (
	"testing"
)

func TestDeriveInstanceType(t *testing.T) {
	cases := []struct {
		gpu     string
		numGPUs int
		want    string
	}{
		{"a100", 1, "a100_1"},
		{"a100", 2, "a100_2"},
		{"h100", 1, "h100_1"},
		{"t4", 4, "t4_4"},
		{"", 1, "unknown_1"},  // missing GPU still gets a label
		{"a100", 0, "a100_1"}, // zero count defaults to 1
	}
	for _, tc := range cases {
		got := deriveInstanceType(tc.gpu, tc.numGPUs)
		if got != tc.want {
			t.Errorf("deriveInstanceType(%q, %d) = %q, want %q",
				tc.gpu, tc.numGPUs, got, tc.want)
		}
	}
}

func TestNullableUUID(t *testing.T) {
	cases := []struct {
		in        string
		wantValid bool
	}{
		{"", false},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"abc123", false},
	}
	for _, tc := range cases {
		got := nullableUUID(tc.in)
		if got.Valid != tc.wantValid {
			t.Errorf("nullableUUID(%q).Valid = %v, want %v",
				tc.in, got.Valid, tc.wantValid)
		}
		if got.Valid && got.String != tc.in {
			t.Errorf("nullableUUID(%q).String = %q, want %q",
				tc.in, got.String, tc.in)
		}
	}
}
