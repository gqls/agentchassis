// FILE: platform/orchestration/datahelpers/links_lifecycle_test.go
package datahelpers

import "testing"

// TestPageWantedLivePredicateForms pins both spellings of the lifecycle axis.
// The aliased and bare forms are one function so they cannot drift — the same
// property NeverDeployedPagePredicateFor already guarantees for the build axis,
// and the reason the estate kept hand-typing `status = 'active'` before this
// member existed (098 debt 4).
func TestPageWantedLivePredicateForms(t *testing.T) {
	if got := PageWantedLivePredicateFor("p"); got != "p.status = 'active'" {
		t.Errorf("aliased form = %q, want p.status = 'active'", got)
	}
	if got := PageWantedLivePredicateFor(""); got != "status = 'active'" {
		t.Errorf("bare form = %q, want status = 'active'", got)
	}
}
