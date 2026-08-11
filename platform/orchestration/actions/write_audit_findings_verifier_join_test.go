// FILE: platform/orchestration/actions/write_audit_findings_verifier_join_test.go
//
// Guards the join bugs_open/213 measured the cost of: a design-audit category
// routed onto an item_type whose registered verifier grades a DIFFERENT predicate.
//
// The failure this prevents is silent by construction. Nothing errors, nothing is
// mis-written, and the verifier answers its own question correctly — it simply
// answers it about the wrong item, returns Resolved:true, and the item closes
// 'complete' with its defect untouched. On the one route where this happened,
// 11 of 11 design-audit items closed clean and not one ever failed to close, while
// every item that DID fail to close belonged to the producer who wrote the verifier.
//
// WHY THE ASSERTION IS SHAPED LIKE THIS. The bug file's own §6 warns: "do not grade
// this by re-running the verifier — it will pass again, for the same correct
// reason." So this test never executes a verifier's predicate against anything. It
// asks a structural question instead — for every design-audit route, either no
// verifier claims that item_type, or the one that does has a scope test AND that
// scope test disclaims a design-audit-shaped item.
//
// At the HEAD that filed bugs_open/213 this test FAILS on dark_section: the route
// pointed at hardcoded_section_colors, a verifier was registered for it, and no
// scope test existed to consult.

package actions

import (
	"testing"

	"github.com/google/uuid"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// designAuditShapedTarget builds the item shape produced by
// WriteAuditFindingsAction, modelled on the gamesdesign.co.uk row that bugs_open/213
// documents at §3 — a named section's defect plus its own stated pass condition,
// which is precisely what a site-aggregate predicate cannot grade.
func designAuditShapedTarget(itemType string) checks.VerifyTarget {
	return checks.VerifyTarget{
		ItemID:   uuid.New(),
		SiteID:   uuid.New(),
		ItemType: itemType,
		Spec: map[string]interface{}{
			"audit_source": "design-audit",
			"category":     "dark_section",
			"page_name":    "index",
			"description": "system-stats-section uses var(--color-primary, #1a1a2e) as its background, " +
				"but the palette defines --color-primary as #00bcd4 (cyan).",
			"acceptance_test": ".system-stats-section computed background-color is a dark colour " +
				"(luminance < 0.1), not the #00bcd4 cyan primary",
		},
	}
}

// TestDesignAuditRoutesAreNotGradedBySomeoneElsesVerifier is the regression guard
// for bugs_open/213 over the whole designItemTypes channel, not just the one route
// that was caught. A future category added to that map lands here automatically.
func TestDesignAuditRoutesAreNotGradedBySomeoneElsesVerifier(t *testing.T) {
	for category, itemType := range designItemTypes {
		verifier, policy := checks.GetVerifier(itemType)
		if verifier == nil {
			// No verifier claims this item_type, so there is no predicate to
			// mistake for this producer's. Completion proceeds unverified, which
			// is a declared gap (verifier_coverage_test.go), not this bug.
			continue
		}

		if policy.Grades == nil {
			t.Errorf("design-audit category %q files item_type %q, which HAS a registered "+
				"verifier and NO scope test.\n"+
				"That verifier will grade this producer's items against its own predicate, "+
				"answer its own question correctly, and close them 'complete' untouched — "+
				"bugs_open/213, measured at 11 of 11 on the dark_section route.\n"+
				"Fix by giving this category its own item_type, or by giving that verifier a "+
				"VerifierPolicy.Grades that disclaims an item of this shape.",
				category, itemType)
			continue
		}

		if speaks, _ := policy.Grades(designAuditShapedTarget(itemType)); speaks {
			t.Errorf("design-audit category %q files item_type %q, whose verifier CLAIMS to "+
				"grade a design-audit-shaped item.\n"+
				"If that is genuinely true the verifier must be able to evaluate "+
				"spec.acceptance_test, which nothing on the completion path reads today "+
				"(bugs_open/213 §5.2). If it is not true, the scope test is too wide.",
				category, itemType)
		}
	}
}

// TestHardcodedColourVerifierGradesItsOwnAggregate is the other direction, and it
// is the one that stops the fix over-correcting: a scope test that disclaimed
// everything would pass the test above while quietly disabling verification for the
// producer who wrote the verifier. That would trade a silent false PASS for a silent
// false BLOCK — both are "the grader cannot see this item", one just fails louder.
func TestHardcodedColourVerifierGradesItsOwnAggregate(t *testing.T) {
	_, policy := checks.GetVerifier("hardcoded_section_colors")
	if policy.Grades == nil {
		t.Fatal("hardcoded_section_colors has no scope test; the bugs_open/213 fix is unwired")
	}

	// The discovery check's own aggregate shape. spec.check has been written by
	// every version of that check since its first commit (62a79c8ac), and all 10
	// live rows from it carry the key — measured 2026-08-10, against 0 of 11 from
	// the design audit.
	own := checks.VerifyTarget{
		ItemID:   uuid.New(),
		SiteID:   uuid.New(),
		ItemType: "hardcoded_section_colors",
		Spec: map[string]interface{}{
			"check":            "hardcoded_section_colors",
			"components_found": 3,
		},
	}
	if speaks, why := policy.Grades(own); !speaks {
		t.Errorf("the verifier disclaimed the check's OWN aggregate item — every real "+
			"detection would now block instead of being graded. why=%q", why)
	}
}

// TestOnlyTheOptedInVerifierCarriesAScopeTest answers the council guardian seat's
// medium objection on the round that approved this change (corr c9c7c83f): the scope
// gate sits inside `verifyBeforeComplete`, which is the single choke point EVERY
// work-item completion passes, across every pipeline — and the claim that a nil
// `Grades` leaves those completions byte-identical was asserted by reading the diff,
// not asserted by a test. Both of the change's other tests target only the one route
// being fixed, so nothing covered the shared branch itself.
//
// This is the cheap half of that: prove the opt-in is actually opt-IN. If a later
// edit gives another item_type a scope test — deliberately or by copying a
// registration line — that type's completions start passing through a predicate
// nobody reviewed for it, which is this bug's own shape reintroduced by the fix for
// it. Exactly one type is licensed to carry one today.
//
// It does NOT prove the nil branch cannot panic; `if policy.Grades != nil` is a plain
// nil check on a func field and there is no target.Spec access before it, so there is
// no type assertion to slip. What it proves is that the branch is not TAKEN for the
// other ten.
func TestOnlyTheOptedInVerifierCarriesAScopeTest(t *testing.T) {
	const optedIn = "hardcoded_section_colors"

	for _, itemType := range checks.RegisteredVerifierItemTypes() {
		_, policy := checks.GetVerifier(itemType)
		switch {
		case itemType == optedIn && policy.Grades == nil:
			t.Errorf("%s is the one type that must carry a scope test and does not — "+
				"the bugs_open/213 fix is unwired", itemType)
		case itemType != optedIn && policy.Grades != nil:
			t.Errorf("%s has acquired a scope test. Every completion of this item_type now "+
				"passes through a predicate that was reviewed for a DIFFERENT type — which is "+
				"bugs_open/213's own defect, reintroduced by its fix. If this is deliberate, "+
				"the scope test needs its own measurement of which producers file this type "+
				"(spec->>'audit_source', NOT item_type, NOT created_by) before it can be trusted.",
				itemType)
		}
	}
}
