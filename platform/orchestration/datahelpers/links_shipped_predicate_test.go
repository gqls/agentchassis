// FILE: platform/orchestration/datahelpers/links_shipped_predicate_test.go
//
// bugs_open/185 — every query that asked `build_status = 'deployed'` meant "is this
// page live" and got the wrong answer for 28 of them.
//
// The alias is why the judgement kept being re-typed by hand: most consumers alias
// their `pages` table, the bare constant could not be dropped into their WHERE
// clause, so each one wrote the expression out again — and a hand-written copy is
// where `= 'deployed'` creeps back in. These tests pin the two properties that make
// the aliased builders safe to reach for: the forms cannot drift from each other,
// and neither may name a build_status value other than 'deployed'.
package datahelpers

import (
	"strings"
	"testing"
)

// TestAliasedPredicateMatchesTheBareOne is the anti-drift property. The bare
// constant is DERIVED from the builder, so this cannot fail by accident — it
// fails if someone re-introduces a hand-written literal for either form.
func TestAliasedPredicateMatchesTheBareOne(t *testing.T) {
	if got := NeverDeployedPagePredicateFor(""); got != NeverDeployedPagePredicate {
		t.Errorf("bare builder and constant disagree:\n  builder:  %s\n  constant: %s", got, NeverDeployedPagePredicate)
	}
	aliased := NeverDeployedPagePredicateFor("p")
	if aliased != strings.ReplaceAll(
		strings.ReplaceAll(NeverDeployedPagePredicate, "deployed_at", "p.deployed_at"),
		"COALESCE(build_status", "COALESCE(p.build_status") {
		t.Errorf("aliased form is not the bare form with the alias applied: %s", aliased)
	}
}

// TestShippedPredicateIsTheNegationAndIsParenthesised guards the shape callers
// depend on. It is dropped into a WHERE clause beside other conjuncts, and the
// predicate is itself an AND — unparenthesised, `x AND NOT a AND b` binds
// differently from what every call site means, and the resulting query still runs.
func TestShippedPredicateIsTheNegationAndIsParenthesised(t *testing.T) {
	got := PageHasShippedPredicateFor("p")
	if !strings.HasPrefix(got, "NOT (") || !strings.HasSuffix(got, ")") {
		t.Errorf("predicate must be a parenthesised negation, got %q", got)
	}
	if !strings.Contains(got, NeverDeployedPagePredicateFor("p")) {
		t.Errorf("predicate is not the negation of the aliased never-deployed form, got %q", got)
	}
}

// TestNeitherFormNamesAnotherBuildStatus is the one with live rows behind it, and
// it is deliberately a copy of the rule links_deployment_test.go already applies to
// the bare constant — because the builders are what new call sites will reach for,
// and a rule enforced on the old spelling only is not enforced.
//
// Naming a status is the failure mode: `needs_rebuild` singled out produced a
// 34-page false-positive class for the nav lane, and `= 'deployed'` alone hides 28
// live pages from every detector that uses it (bugs_open/185). `deployed_at IS NULL`
// is what carries both cases correctly.
func TestNeitherFormNamesAnotherBuildStatus(t *testing.T) {
	for _, form := range []string{
		NeverDeployedPagePredicateFor(""),
		NeverDeployedPagePredicateFor("p"),
		PageHasShippedPredicateFor("p"),
	} {
		for _, forbidden := range []string{"needs_rebuild", "planned", "pending"} {
			if strings.Contains(form, forbidden) {
				t.Errorf("predicate names build_status %q — that is the false-positive class this exists to avoid: %s", forbidden, form)
			}
		}
		if !strings.Contains(form, "deployed_at IS NULL") {
			t.Errorf("predicate does not key on deployed_at, so it cannot see a shipped needs_rebuild page: %s", form)
		}
	}
}

// TestEmptyAliasProducesNoStrayDot catches the obvious builder bug, which would
// produce `.deployed_at IS NULL` — valid-looking, and a syntax error only at
// execution time, i.e. inside whichever check runs next in production.
func TestEmptyAliasProducesNoStrayDot(t *testing.T) {
	if strings.HasPrefix(NeverDeployedPagePredicateFor(""), ".") {
		t.Error("empty alias produced a leading dot")
	}
	if strings.Contains(NeverDeployedPagePredicateFor(""), "(.") {
		t.Error("empty alias produced a stray dot inside COALESCE")
	}
}
