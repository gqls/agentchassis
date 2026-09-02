// FILE: platform/orchestration/actions/unrendered_imagery_markers_lockstep_test.go
//
// Lockstep contract: discovery_checks.InteractiveStructuralMarkers mirrors this
// package's private interactiveStructuralMarkers (save_page_sections_action.go)
// — the markers bugs_open/357's machinery uses to recognise a stored
// interactive fragment, which check_unrendered_page_imagery reuses to keep a
// 357-shaped hero row out of its "deliverable" state.
//
// It is a MIRROR, not a single source, because single-sourcing requires moving
// the private list down the dependency graph (discovery_checks cannot import
// actions — actions imports it), and that edit belongs to the 357 lane whose
// live machinery reads the original. This test is what makes the mirror safe in
// the meantime: either list changing alone breaks the build, in this package,
// where both are visible. If the 357 lane moves the list into discovery_checks,
// delete this test and the private copy together.
package actions

import (
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestInteractiveStructuralMarkersLockstep(t *testing.T) {
	if len(checks.InteractiveStructuralMarkers) != len(interactiveStructuralMarkers) {
		t.Fatalf("marker lists differ in length: discovery_checks has %d, actions has %d — "+
			"a fragment shape one recognises and the other does not misclassifies pages in "+
			"check_unrendered_page_imagery",
			len(checks.InteractiveStructuralMarkers), len(interactiveStructuralMarkers))
	}
	for i, m := range interactiveStructuralMarkers {
		if checks.InteractiveStructuralMarkers[i] != m {
			t.Errorf("marker %d: discovery_checks %q != actions %q",
				i, checks.InteractiveStructuralMarkers[i], m)
		}
	}
}
