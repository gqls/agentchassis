// FILE: platform/orchestration/actions/write_audit_findings_input_spec_test.go
//
// Pins bugs_open/264's candidate 2: audit_source is Required with no Default,
// so a config that fails to resolve it now fails ExtractActionInputs loudly
// instead of silently landing under "design-audit". A test that only checked
// the four live configs would pass on the OLD (Optional+Defaulted) shape too —
// the property worth pinning is that an UNRESOLVABLE audit_source is now a
// hard extraction error, not a review round's claim about it.

package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// TestWriteAuditFindingsInputSpec_UnresolvableAuditSourceErrors reproduces the
// exact pre-fix shape (a plain string with no dot and no matching top-level
// collected_data key, e.g. "content-quality-audit") and asserts extraction now
// fails naming audit_source, rather than silently defaulting to "design-audit".
func TestWriteAuditFindingsInputSpec_UnresolvableAuditSourceErrors(t *testing.T) {
	_, err := datahelpers.ExtractActionInputs(
		// site_id resolves fine — isolates the failure to audit_source alone.
		map[string]interface{}{"site_record": map[string]interface{}{"site_id": "11111111-1111-1111-1111-111111111111"}},
		map[string]interface{}{
			"site_id":      "site_record.site_id",
			"audit_source": "content-quality-audit", // unresolvable: no dot, no matching top-level key
		},
		WriteAuditFindingsInputSpec, zap.NewNop())

	if err == nil {
		t.Fatal("ExtractActionInputs returned no error for an unresolvable audit_source; " +
			"bugs_open/264's whole point is that this must now fail loudly, not silently default to design-audit")
	}
	if !strings.Contains(err.Error(), "audit_source") {
		t.Fatalf("error %q does not name audit_source as the missing field", err.Error())
	}
}

// TestWriteAuditFindingsInputSpec_ResolvableAuditSourceStillWorks is the
// no-op-case companion: the Required change must not break a config that
// resolves correctly (the shape migration 399 gives all four live auditors).
func TestWriteAuditFindingsInputSpec_ResolvableAuditSourceStillWorks(t *testing.T) {
	in, err := datahelpers.ExtractActionInputs(
		map[string]interface{}{
			"site_record":          map[string]interface{}{"site_id": "11111111-1111-1111-1111-111111111111"},
			"audit_source_literal": map[string]interface{}{"audit_source": "content-quality-audit"},
		},
		map[string]interface{}{
			"site_id":      "site_record.site_id",
			"audit_source": "audit_source_literal.audit_source",
		},
		WriteAuditFindingsInputSpec, zap.NewNop())

	if err != nil {
		t.Fatalf("ExtractActionInputs failed on a resolvable audit_source dot-path: %v", err)
	}
	if got := in.Get("audit_source"); got != "content-quality-audit" {
		t.Fatalf("audit_source = %q, want %q", got, "content-quality-audit")
	}
}
