// FILE: platform/orchestration/actions/work_item_routability_single_definition_test.go
//
// OWNER RULING 2026-08-17, breaking the tie on council round c22998e8 (bugs_closed/284),
// where two seats objected in OPPOSITE directions about the same edit:
//   * reuse_agent (medium): a THIRD rendering of "does this work item name a handler
//     that exists?" — discovery_checks/remit.go's HandlerStepConfig — was left coupled
//     to the other two by a PROSE COMMENT only.
//   * guardian (medium): touching claim_work_item_action.go at all exceeded the bug.
// The ruling was to unify. This change does it in the way that also honours the second
// objection: the renderer moved DOWN into discovery_checks (import direction — `actions`
// imports `discovery_checks`, never the reverse), and `workItemHandlerRegisteredSQL`
// stays as a thin delegation, so every existing call site in `actions`, the claim path
// included, is BYTE-IDENTICAL.
//
// WHAT THIS TEST IS FOR. Unifying three copies is worth nothing if a fourth can appear
// tomorrow; the estate has already paid for that lesson once (the idx_swi_dedup ↔
// workItemTerminalStatuses landmine). So the guard is structural: within
// platform/orchestration, the registration predicate may be written in exactly ONE
// place, and any new inline copy fails the build.
//
// ⚠ WHAT IS DELIBERATELY OUT OF SCOPE, named so nobody thinks it was missed. The tree
// holds TWO further renderings of the same SQL in
// internal/core-manager/admin/agent_handlers.go (:101 "check if type already exists"
// before CREATING an agent, and :724 "verify agent exists" before recreating topics).
// They are NOT unified and should not be: they ask a different QUESTION that happens to
// share a predicate — admin CRUD existence, not work-item routability. Binding them to
// HandlerRegisteredSQL would mean that narrowing routability later (say, requiring
// is_active) would silently change what the admin API accepts and what it 404s. Same
// text, different contract; the scope of this guard is the contract, not the text.
//
// (Discovered while implementing the ruling: the objection said "a third copy" and the
// tree actually held FIVE renderings. Counting them was one grep — see bugs_closed/284.)

package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// TestRoutabilityPredicateIsDelegatedNotCopied proves the `actions` helper is a
// delegation rather than a second copy that happens to agree today. It compares
// rendered output for several expression shapes, so a re-inlined body with a subtle
// difference (a missing alias, a narrowed clause) fails here rather than in production.
func TestRoutabilityPredicateIsDelegatedNotCopied(t *testing.T) {
	for _, expr := range []string{"$1", "wi.handler_agent", "ad2.type", "'page-build-handler'"} {
		got := workItemHandlerRegisteredSQL(expr)
		want := checks.HandlerRegisteredSQL(expr)
		if got != want {
			t.Errorf("actions and discovery_checks render DIFFERENT registration predicates for %q:\n"+
				"  actions:          %s\n  discovery_checks: %s\n"+
				"They must be one definition (owner ruling 2026-08-17, bugs_closed/284) — "+
				"workItemHandlerRegisteredSQL should delegate, not re-implement.", expr, got, want)
		}
		if !strings.Contains(got, "agent_definitions") || !strings.Contains(got, "deleted_at IS NULL") {
			t.Errorf("rendered predicate lost its substance for %q: %s", expr, got)
		}
	}
}

// inlineRegistrationCheck matches the registration predicate written out by hand.
var inlineRegistrationCheck = regexp.MustCompile(`(?i)EXISTS\s*\(\s*SELECT\s+1\s+FROM\s+agent_definitions`)

// TestRegistrationPredicateHasExactlyOneDefinition scans platform/orchestration for
// hand-written copies. Exactly one is licensed: the renderer itself.
func TestRegistrationPredicateHasExactlyOneDefinition(t *testing.T) {
	const owner = "platform/orchestration/actions/discovery_checks/remit.go"
	root := moduleRoot(t)

	var found []string
	err := filepath.WalkDir(filepath.Join(root, "platform", "orchestration"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		for i, line := range strings.Split(stripLineComments(string(src)), "\n") {
			if inlineRegistrationCheck.MatchString(line) {
				found = append(found, rel+":"+strconv.Itoa(i+1))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	if len(found) != 1 || !strings.HasPrefix(found[0], owner) {
		t.Errorf("the agent-registration predicate must be written in exactly ONE place "+
			"(%s, rendered by HandlerRegisteredSQL). Found %d: %v\n"+
			"A hand-written copy drifts from the claim path the moment either is narrowed, and "+
			"nothing joins them up — that is bugs_closed/284's own defect one level down. Call "+
			"the renderer instead. (core-manager's admin CRUD copies are deliberately outside "+
			"this scope — see this file's header.)", owner, len(found), found)
	}
}

// TestInlineRegistrationPatternStillBites keeps the scanner honest: a pattern that
// silently stops matching would pass forever on an empty result set.
func TestInlineRegistrationPatternStillBites(t *testing.T) {
	for _, bad := range []string{
		`"SELECT EXISTS(SELECT 1 FROM agent_definitions WHERE type = $1 AND deleted_at IS NULL)",`,
		`return "EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = " + expr`,
		`EXISTS( SELECT 1 FROM agent_definitions`,
	} {
		if !inlineRegistrationCheck.MatchString(bad) {
			t.Errorf("scanner no longer matches an inline copy — the guard is disarmed: %q", bad)
		}
	}
	for _, good := range []string{
		`return checks.HandlerRegisteredSQL(handlerExpr)`,
		`EXISTS (SELECT 1 FROM site_work_items WHERE site_id = $1)`,
		`FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config)`,
	} {
		if inlineRegistrationCheck.MatchString(good) {
			t.Errorf("scanner false-positives on %q", good)
		}
	}
}
