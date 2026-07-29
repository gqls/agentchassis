// FILE: platform/orchestration/actions/validate_page_content_claims_test.go
//
// Severity contract for the claims gate (validate_page_content check 8):
// banned claims are blockers (known falsehoods, human-ruled), unregistered
// numbers are errors (route to human review, never block outright — the
// extractor has false positives by design). The scan logic itself is covered
// by the benchmark corpus in datahelpers/claims_test.go.

package actions

import (
	"go.uber.org/zap"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const claimsGateTestEB = `{
  "audit_doc": "docs/leopardessconsulting/AUDIT_verified_facts.md",
  "facts": [
    {"id": "C1-records-verified", "claim": "2,767 records verified", "value": 2767,
     "kind": "metric", "source": {"sql": "SELECT 1"}, "verified_at": "2026-07-16", "tolerance": "exact"}
  ],
  "banned_claims": [
    {"pattern": "eight (functional |business )?(departments|functions|areas)", "reason": "L0 audit U1"}
  ]
}`

func TestClaimsGateSeverities(t *testing.T) {
	eb, err := datahelpers.ParseEvidenceBase([]byte(claimsGateTestEB))
	if err != nil || eb == nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}

	blocks := datahelpers.ExtractAssertionText(
		`<p>We span eight departments and serve 45 clients.</p>`)

	banned := checkBannedClaims(blocks, eb, true, "test-site", zap.NewNop())
	if len(banned) != 1 || banned[0].Severity != "blocker" || banned[0].Category != "claims" {
		t.Errorf("banned claim must be a blocker in category claims, got %+v", banned)
	}

	numbers := checkUnregisteredNumbers(blocks, eb, datahelpers.ClaimSurface{})
	if len(numbers) != 1 || numbers[0].Severity != "error" || numbers[0].Category != "claims" {
		t.Errorf("unregistered number must be an error in category claims, got %+v", numbers)
	}
	if numbers[0].Value != "45" {
		t.Errorf("expected the fabricated '45' to flag, got %+v", numbers)
	}
}
