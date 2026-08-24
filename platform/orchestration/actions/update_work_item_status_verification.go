// FILE: platform/orchestration/actions/update_work_item_status_verification.go
//
// The registered-verifier gate for the SECOND writer of `complete`
// (bugs_open/375).
//
// WHAT THE PROBLEM IS, in plain terms. A work item is one recorded defect on one
// site. A VERIFIER is a per-item_type re-check that runs immediately before the
// item is stamped `complete`: it re-runs the defect's own predicate and refuses
// the completion if the defect is still there. Its whole purpose is to stop a
// handler saga that reports success without having touched anything
// (verifier_coverage_test.go; the class of bugs_open/017, one level up).
//
// The platform has TWO code paths that stamp `complete`. CompleteWorkItemAction
// asks the verifier (complete_work_item_verification.go).
// UpdateWorkItemStatusAction never has — `GetVerifier` was called from exactly
// one place in the tree. So which guard a completion gets depends on which
// action a DB-configured workflow happened to name, and nothing said so.
//
// THE BLAST RADIUS, MEASURED rather than assumed [MEASURED 2026-08-24, live DB]:
// of 200 live agent definitions, 6 name `update_work_item_status` across 22
// steps, and 4 of them reach `complete` across 6 arms — image-build-handler,
// image-source-unsatisfiable-handler, image-url-404-handler and
// required-fields-missing-handler (3 arms). Those four handle SEVEN item types
// (needs_imagery, required_fields_missing, needs_hero_image,
// unfulfilled_hero_variant, needs_logo, image_url_404,
// image_source_unsatisfiable; 578 completions all-history) and NONE of the seven
// has a registered verifier. So no verifier is bypassed today: the defect is
// LATENT. (Controlled: the same 13-type registered list run without the handler
// filter returns 11 of 13 with rows, under handlers disjoint from those four, so
// the zero is not a mis-spelled IN list. Re-run both from
// RUNBOOK_completion_verifier_gap.md — a census does not go wrong, it goes stale
// by addition.)
//
// ⚠ THE TYPE AND COMPLETION COUNTS ABOVE MUST BE READ FROM
// `site_work_items` UNION `site_work_items_archive`, AND THE FIRST VERSION OF THIS
// COMMENT DID NOT. `site_work_items` is a ROLLING WINDOW — the archiver moves
// terminal rows out — so a census over it alone answers "recently", not
// "all-history". Corrected after the council's prior_art_librarian seat objected on
// exactly that ground (corr 7a6add95, medium): the live table alone reported FIVE
// types and 134 completions; the union reports SEVEN and 578. The two extra types
// (unfulfilled_hero_variant, image_url_404) had been completed entirely into the
// archive and were invisible. The CONCLUSION survived — all seven are unverified,
// so the zero intersection holds — but it survived by luck, not by the measurement.
//
// WHY IT IS STILL A BUG. verifier_coverage_test.go maintains the list of types
// that ought to get verifiers and calls it, in its own words, "the actionable
// backlog, not an excuse list". Two of the five are on it. Whoever works that
// backlog will register a verifier, watch the coverage test go green, and protect
// nothing — the trap is set for a named person.
//
// And it is signposted with the WRONG warning, which is the part that decided this
// design. Register entry CQ-023 tells that person: "a verifier later registered
// for required_fields_missing would fail-closed the `converted` arm's completion".
// That is FALSE today, and false BECAUSE of this bug — `close_converted` is an
// update_work_item_status step, so registering the verifier would not fail-close
// anything, it would do nothing at all. They are braced for one wrong outcome and
// would get a different one, silently.
//
// WHY OPT-IN, AND WHY PER STEP (owner ruling 2026-08-02 §2: new authority on a
// shared seam ships as a field whose unsafe default is OFF, so the decision is
// visible to a reviewer of the CALLER). Making the consult automatic would make
// CQ-023's sentence come true — it would break a live route as a side effect of
// arming a guard nobody asked for. Arming is per STEP rather than per item_type
// because CQ-023 shows the decision is per close-path: one router's `converted`
// and `stale` arms want different answers, and the reviewer who can tell them
// apart is the one reading that agent's config.
//
// RFC_022's narrowing (owner ruling 2026-08-11) therefore applies, with the
// consumers ENUMERATED rather than asserted — "asserting it without the query is
// itself the objection". All three conditions hold: it is opt-in; the unsafe side
// (today's behaviour) is the default; and zero live consumers name the key — the
// census above is the enumeration, and no seed sets `verify_before_complete`.
//
// WHY GATE 2 ONLY, and not verifyBeforeComplete wholesale. That function runs two
// gates. Gate 1b (the no-change gate) reads the HANDLER'S OWN REPLY PAYLOAD;
// UpdateWorkItemStatusAction has no such payload — it has step config — so handing
// gate 1b something else would grade the wrong evidence, which is precisely the
// error complete_work_item_no_change.go's own header records about re-reading the
// result column. Gate 1b's roster holds one type today (dark_section_audit) and
// none of our five, so calling it would be INERT — and "inert today" is the exact
// reasoning this bug is about, so it is not good enough.
//
// THE SECOND HALF, and why it is not decoration. An opt-in safety field with an
// unsafe default rots unexercised unless something tells the next person to arm
// it — a cost the owner named explicitly on 2026-07-29 §2. So when this arm runs
// UNARMED on an item whose type DOES have a registered verifier, the completion
// proceeds exactly as before and the bypass is RECORDED ON THE ROW
// (result._verification.status = "verifier_not_consulted"). No completion outcome
// changes, so there is no liveness risk; it needs no hand-maintained roster, so it
// cannot go stale by addition; and it lands on a queryable surface rather than a
// pod log that does not survive a roll — the same remedy recordUnknownVerdict and
// `completion_skipped` already use. It is ONE write per completion event, never on
// a cadence, so it does not touch the WII-018 stale-reaper trap.

package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// verifierLookup is checks.GetVerifier, indirected through a package variable so
// a test can drive this gate WITHOUT registering anything into the process-wide
// verifier registry.
//
// That is not a stylistic preference. The registry has no removal, and it is read
// by a lockstep guard asserting a contract across the whole process:
// TestClaimTimeoutExclusionCoversBothCompletionGates requires that every item_type
// with a verifier is also declared in livespec.ClaimedItemTimeoutExclusions,
// because a THIRD writer — the claimed-item-timeout sweep — completes rows
// directly and neither gate runs for it (bugs_closed/317). A test that registers a
// synthetic type and leaves it there breaks that guard, which is exactly what the
// first draft of this change did. Indirecting here keeps the fixture local and
// leaves the real registry the single source of truth for every other reader.
//
// Production NEVER re-points this: TestVerifierLookupIsNotASwitchInProduction
// asserts there is one assignment in the package's non-test source, the
// declaration below.
var verifierLookup = checks.GetVerifier

// updateStatusVerifyConfigKey is the step-config key that arms the gate. Named
// here rather than spelled inline because the failure messages, the tests and the
// documentation that tells the next person how to arm it must all say the same
// word.
const updateStatusVerifyConfigKey = "verify_before_complete"

// verifyBeforeUpdateStatusComplete is completion gate 2 for
// UpdateWorkItemStatusAction's `complete` arm.
//
// Returns the payload to embed at result._verification (nil when there is nothing
// to say) and whether the completion may proceed. It NEVER blocks unless `armed`.
//
// One row read serves both arms deliberately: the unarmed arm has to know the
// item's type in order to notice that a verifier existed, which is the whole point
// of it, and reading the row twice would be a second definition of the same
// question.
func verifyBeforeUpdateStatusComplete(ctx context.Context, db *sql.DB, itemID uuid.UUID,
	armed bool, logger *zap.Logger) (map[string]interface{}, bool) {

	row, err := loadWorkItemVerifyRow(ctx, db, itemID)
	if err != nil {
		// Same policy as verifyBeforeComplete's: the completion UPDATE will no-op
		// or fail on its own, and refusing here would turn an unreadable row into a
		// spent attempt.
		return nil, true
	}

	verifier, policy := verifierLookup(row.ItemType)
	if verifier == nil {
		// No verifier for this type — nothing is being bypassed, so nothing is
		// recorded. This is every live item of the six configured arms today.
		return nil, true
	}

	if !armed {
		// The trap, sprung — and now audible. Complete exactly as before, but say
		// on the row that a guard existed and was not consulted.
		logger.Warn("UpdateWorkItemStatusAction: a verifier is registered for this item_type and this step did not consult it",
			zap.String("item_id", itemID.String()),
			zap.String("item_type", row.ItemType),
			zap.String("remedy", "set "+updateStatusVerifyConfigKey+": true on this step's config, after reading that item type's close paths (bugs_open/375, CQ-023)"))
		return map[string]interface{}{
			"status":    "verifier_not_consulted",
			"item_type": row.ItemType,
			"writer":    "update_work_item_status",
			"detail": "a verifier is registered for this item_type, but this completion was written by " +
				"update_work_item_status with " + updateStatusVerifyConfigKey + " unset, so the defect was not re-checked",
			"remedy": "arm " + updateStatusVerifyConfigKey + ": true on this step, or move the completion to " +
				"complete_work_item; read that item type's close paths first (bugs_open/375, register CQ-023)",
		}, true
	}

	return runRegisteredVerifier(ctx, db, itemID, row.ItemType, row.SpecJSON, row.SiteID, row.PageID,
		verifier, policy, logger)
}
