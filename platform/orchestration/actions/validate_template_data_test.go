// FILE: platform/orchestration/actions/validate_template_data_test.go
//
// bugs_open/453, council corr 54abc24b (bug_historian, medium). validateTemplateData
// is the runtime's ONLY cross-check over the input_fields <-> template-context pair,
// and it was asking about a key the extractor never writes: it split a dotted entry
// on parts[0] while ExtractFields stores under the LAST segment.
//
// The test drives the REAL extractor rather than a hand-built map, so it cannot pass
// on a fixture that agrees with the old bug. Log output is asserted through a zap
// observer, because "reported as missing" IS the behaviour — the function returns
// nothing and changes no control flow.
package actions

import (
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestValidateTemplateDataAsksForTheKeyTheExtractorActuallyWrote(t *testing.T) {
	collected := map[string]interface{}{
		"reviewed_brief": map[string]interface{}{"company_name": "Acme"},
	}
	// One dotted entry, extracted for real.
	const field = "reviewed_brief.company_name"
	templateData := datahelpers.ExtractFields(collected, []string{field}, zap.NewNop())

	// Precondition, asserted rather than assumed: the extractor stored the LEAF.
	// If this ever changes, the rest of the test is meaningless and should fail here.
	if _, ok := templateData["company_name"]; !ok {
		t.Fatalf("precondition: ExtractFields must store the leaf; got keys %v", getTemplateDataKeys(templateData))
	}
	if _, ok := templateData["reviewed_brief"]; ok {
		t.Fatalf("precondition: the HEAD segment must not be written by a dotted entry; keys %v",
			getTemplateDataKeys(templateData))
	}

	core, logs := observer.New(zap.WarnLevel)
	validateTemplateData(templateData, map[string]interface{}{
		"input_fields": []interface{}{field},
	}, zap.New(core))

	if n := logs.FilterMessageSnippet("TEMPLATE DATA VALIDATION FAILED").Len(); n != 0 {
		t.Errorf("a dotted entry that extracted SUCCESSFULLY was reported as missing (%d error log(s)) — "+
			"the validator is splitting on the first segment while ExtractFields stores the last", n)
	}
}

// The other direction: a genuine absence must still be reported, or the fix above
// would have been achieved by making the check blind.
func TestValidateTemplateDataStillReportsAGenuineAbsence(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	validateTemplateData(map[string]interface{}{"present": 1}, map[string]interface{}{
		"input_fields": []interface{}{"present", "nowhere.to_be_found"},
	}, zap.New(core))

	entries := logs.FilterMessageSnippet("TEMPLATE DATA VALIDATION FAILED").All()
	if len(entries) != 1 {
		t.Fatalf("a missing field must still be reported; got %d error log(s)", len(entries))
	}
	// Read it through ContextMap rather than asserting on Field.Interface:
	// zap.Strings stores an ArrayMarshaler, not a []string, so a type assertion
	// there yields an empty slice and the assertion passes vacuously in the
	// direction that matters.
	got := fmt.Sprint(entries[0].ContextMap()["missing_fields"])
	if !strings.Contains(got, "nowhere.to_be_found") {
		t.Errorf("missing_fields = %s, want it to name the absent field", got)
	}
	if strings.Contains(got, "present") {
		t.Errorf("missing_fields = %s, must not name the field that WAS supplied", got)
	}
}
