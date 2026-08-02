// FILE: platform/orchestration/datahelpers/single_owner_test.go
//
// RFC 006 / bugs_closed/150. ListSingleOwnerActions is the seam between a spec
// declaration and the offline detector that enforces it. The one property worth
// pinning here is the one that is easy to get wrong and impossible to notice:
// the listing must NOT be gated on checksConfig(), the way ListDeclaredConfigKeys
// is. Coupling them would make SingleOwner silently inert for any action that
// has not adopted the unrelated config-key mechanism — a declaration that reads
// as enforced and enforces nothing, which is WRONG_CALLS.md 2026-07-28's shape.
package datahelpers

import "testing"

func TestListSingleOwnerActionsDoesNotRequireConfigCheckingOptIn(t *testing.T) {
	const name = "test_single_owner_without_config_optin"

	// No ConfigKeys, no CheckConfig — i.e. this action has NOT opted into
	// unknown-config-key detection. Its single-ownership must still be listed.
	RegisterActionInputSpec(name, ActionInputSpec{
		Required:    []string{"site_id"},
		SingleOwner: true,
	})

	found := false
	for _, a := range ListSingleOwnerActions() {
		if a == name {
			found = true
		}
	}
	if !found {
		t.Errorf("%s declares SingleOwner but is absent from ListSingleOwnerActions(); "+
			"the listing has been coupled to config-key opt-in, which makes the declaration inert", name)
	}
}

func TestListSingleOwnerActionsExcludesUndeclaredActions(t *testing.T) {
	const name = "test_ordinary_shared_action"

	RegisterActionInputSpec(name, ActionInputSpec{
		Required:   []string{"site_id"},
		ConfigKeys: []string{"whatever"},
		// SingleOwner deliberately unset: the zero value must mean "shared is
		// fine", so that adding this field cannot make every existing action
		// single-owner by default.
	})

	for _, a := range ListSingleOwnerActions() {
		if a == name {
			t.Errorf("%s does not declare SingleOwner but was listed; the default has inverted "+
				"and the detector would fire on every ordinary shared action", name)
		}
	}
}

func TestListSingleOwnerActionsIsSorted(t *testing.T) {
	RegisterActionInputSpec("test_zzz_single_owner", ActionInputSpec{SingleOwner: true})
	RegisterActionInputSpec("test_aaa_single_owner", ActionInputSpec{SingleOwner: true})

	list := ListSingleOwnerActions()
	for i := 1; i < len(list); i++ {
		if list[i-1] > list[i] {
			t.Fatalf("ListSingleOwnerActions() is not sorted at %d: %v — an unstable order makes "+
				"the detector's output churn between runs and defeats diffing two reports", i, list)
		}
	}
}
