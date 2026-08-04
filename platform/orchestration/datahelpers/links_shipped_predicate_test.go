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
	"os"
	"path/filepath"
	"runtime"
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

// TestMigration302CarriesTheCanonicalPredicateVerbatim is the bit-for-bit diff
// the council's architecture seat asked for on bugs_open/185 tranche 3, kept as
// a standing drift guard rather than run once in a terminal.
//
// Migration 302 surfaces `has_shipped` in build-site-planner's
// load_existing_pages query. A migration is SQL text in a file — it cannot CALL
// PageHasShippedPredicateFor — so nothing structural stops its hand-written
// predicate drifting from the canonical builder, and a fourth subtly-different
// spelling of "has this page shipped" is precisely the defect family
// bugs_open/185 exists to close. This test makes the two one: the migration
// must contain the builder's exact output, so editing either without the other
// goes red here.
//
// If migration 302's file is ever archived/moved, move this assertion onto its
// successor rather than deleting it — the LIVE row was written from that text,
// and this test is the only thing tying the row's predicate to the builder.
func TestMigration302CarriesTheCanonicalPredicateVerbatim(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	migPath := filepath.Join(repoRoot, "docs", "agent_docs", "sql_for_agents",
		"302_load_existing_pages_has_shipped.sql")
	raw, err := os.ReadFile(migPath)
	if err != nil {
		t.Fatalf("reading migration 302 (moved? update this test's path AND keep the assertion): %v", err)
	}
	mig := string(raw)

	want := PageHasShippedPredicateFor("p")
	if !strings.Contains(mig, want) {
		t.Errorf("migration 302 does not carry PageHasShippedPredicateFor(\"p\") verbatim.\n  builder: %s\nThe live build-site-planner row was written from this file; a drifted spelling here is a fourth definition of 'has this page shipped' (bugs_open/185).", want)
	}

	// Exactly once in the EXECUTABLE SQL — a second occurrence there would mean a
	// duplicated column. Comment lines are excluded deliberately: this file's
	// post-apply notes quote the predicate for an operator to paste, and that copy
	// SHOULD match the canonical spelling too.
	//
	// > The first version of this test counted over the whole file and went red the
	// > moment the doc comment was corrected to the canonical spelling — i.e. it
	// > punished the fix it exists to encourage. Caught by making that very edit.
	// > Scoping the count to executable lines keeps both properties: the SQL cannot
	// > gain a duplicate column, and every quoted copy anywhere in the file is still
	// > required to be the canonical spelling by the loop below.
	var sqlOnly []string
	for _, line := range strings.Split(mig, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			sqlOnly = append(sqlOnly, line)
		}
	}
	if n := strings.Count(strings.Join(sqlOnly, "\n"), want); n != 1 {
		t.Errorf("predicate appears %d times in migration 302's executable SQL, want exactly 1", n)
	}

	// No NEAR-MISS copy anywhere in the file, comments included. A predicate that
	// differs only in whitespace is the drift this guard exists to catch, and a
	// wrong copy sitting in an operator's paste-this block is how it spreads.
	loose := strings.NewReplacer(" ", "", "\t", "").Replace(want)
	for i, line := range strings.Split(mig, "\n") {
		squashed := strings.NewReplacer(" ", "", "\t", "").Replace(line)
		if strings.Contains(squashed, loose) && !strings.Contains(line, want) {
			t.Errorf("line %d carries a whitespace-drifted copy of the predicate — make it byte-identical to PageHasShippedPredicateFor(\"p\"):\n  %s", i+1, strings.TrimSpace(line))
		}
	}
}
