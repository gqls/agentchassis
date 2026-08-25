// FILE: platform/orchestration/actions/discovery_checks/verify_required_fields_missing.go
//
// The completion verifier for `required_fields_missing` — the first entry taken
// off verifier_coverage_test.go's own catMechanical backlog, which says of itself:
// "These SHOULD get verifiers — this is the actionable backlog, not an excuse list."
//
// WHAT IT RE-CHECKS, in plain terms. The finding says: this component instance
// declares schema-required fields sourced from the LLM, and content_data never
// received them, so the template renders them as empty strings. That predicate is
// re-runnable against live rows, which is what makes the type mechanical. This
// verifier re-resolves the component and asks missingRequiredValueFields again —
// the DETECTOR'S OWN function, not a re-implementation of it, so the two cannot
// drift into disagreeing about what "required" and "empty" mean.
//
// ── THE ONE DECISION THAT MATTERS: WHICH AXIS RESOLVES THE COMPONENT ──
//
// It resolves on the LIFECYCLE axis — COALESCE(build_status,'pending') <> 'removed'
// — and NOT on build_status = 'deployed', even though the detector scans
// 'deployed'. That asymmetry is deliberate and it is the whole lesson of
// bugs_closed/367.
//
// That bug: the router's classifier resolved the offending component with
// `WHERE pc.build_status = 'deployed'`, mirroring the detector's own filter. A
// finding about a NON-deployed component therefore resolved nothing, fell through
// to the `stale` arm, and was CLOSED `complete` with no error — a true finding
// scored as a success, and quieter than the failure it replaced. Migration 574
// fixed it by changing the RULE rather than the filter: resolution moved to the
// lifecycle axis, spelled to match pageComponentNotRemovedSQL.
//
// A verifier that mirrored the detector's `deployed` filter would rebuild that
// exact defect one layer up, and worse: it would return Resolved:true for a
// component that is merely pending, i.e. it would certify a defect as fixed
// because it could not see it. The question a verifier asks is "is the defect
// gone?", and a component sitting at `pending` with empty content_data will render
// empty the moment it deploys. The defect is not gone; it is not yet visible.
//
// ── EVERY "RESOLVED" ARM RESTS ON A POSITIVE FACT ──
//
// RFC_017's ruling is that "I could not check" must never read as "I checked and
// it is fixed". So no arm below returns Resolved:true because a lookup came back
// empty. Each one names a live row:
//
//   - the PAGE is gone            → nothing renders this slot. Positive.
//   - the slot is SUPPRESSED      → the page declares it excluded. Positive.
//   - the component is REMOVED    → the assembly-excluded tombstone. Positive.
//   - the component is LOCKED     → the accept-as-is resolution (CQ-023). Positive.
//   - it is a RUNTIME-FILL slot   → deliberately empty at build time; the detector
//                                   skips these, so it would no longer file this.
//   - the predicate returns EMPTY → the fields are populated. Positive.
//
// A page that resolves with NO row at that slot is the one case that is an
// absence rather than a fact, and it is treated as resolved for a stated reason:
// the page exists and nothing occupies the slot, so nothing can render an empty
// field there. That is the same evidence migration 574 accepts for `stale`.
// Anything that could not be READ returns an error, which is fail-CLOSED.
//
// ── WHY IT DECLARES Grades, AND WHY THE SHAPE IS POSITIVE ──
//
// bugs_open/213: a verifier registered for an item_type grades EVERY row carrying
// that name, including rows filed by a producer who meant something else — and
// because the verifier answers its own question correctly, the mismatch reads as a
// clean pass. Measured there: 11 of 11 second-producer items closed complete,
// untouched.
//
// This type has TWO producers as of 2026-08-23 (CQ-023, and the ruling of
// 2026-08-02 §1 that a shared item_type must name its producer set):
//
//  1. check_required_fields_missing.go — POST-DEPLOY, scans stored component
//     instances, writes spec.component_id / page_id / page_name / slot_name /
//     component_function / missing_fields / reason.
//  2. work_items_common.go emitRequiredFieldsMissing (bugs_closed/342) —
//     RENDER-TIME, files about components that by construction are NOT stored yet,
//     and writes page_name / slot_name / missing_fields but NONE of component_id,
//     page_id, component_function or reason [MEASURED 2026-08-23: 62 of 62 vs 0 of 3].
//
// Producer 2's items are the exact hazard: this verifier resolves a STORED
// component, and for a finding about one that was never stored the resolution
// would come back empty and the "no row at that slot" arm above would certify it
// resolved. So Grades declares a POSITIVE shape — "the finding names a component
// instance that existed when it was filed" — rather than a blocklist of producers,
// which is what bugs_open/213 asks for and what keeps a well-shaped item from an
// UNKNOWN third producer gradable.
//
// ⚠ Disclaiming BLOCKS completion (runRegisteredVerifier's out_of_scope arm), it
// does not wave the item through. That is correct and it is not free: producer 2's
// items would stop completing on any arm that ARMS this verifier. No arm is armed
// (livespec.UnarmedVerifiedCompleters), so this is inert today — see the
// registration note at the foot of this file.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// requiredFieldsComponentNotRemovedSQL is the lifecycle predicate, spelled to
// match actions.pageComponentNotRemovedSQL and migration 574's rule.
//
// ⚠ Deliberately NOT `build_status = 'deployed'`. See the header — mirroring the
// detector's filter here is bugs_closed/367 rebuilt one layer up.
const requiredFieldsComponentNotRemovedSQL = "COALESCE(pc.build_status, 'pending') <> 'removed'"

// ── NOT REGISTERED YET, AND THAT IS THE POINT OF THIS BLOCK ──────────────────
//
// The one line that arms this file is:
//
//	func init() {
//		RegisterVerifierWithPolicy("required_fields_missing", VerifyRequiredFieldsMissingResolved,
//			VerifierPolicy{Grades: gradesRequiredFieldsMissing})
//	}
//
// It is deliberately absent, because REGISTERING A VERIFIER IS NOT A ONE-LINE
// CHANGE and this file is where the next author will look. Writing that init()
// with nothing else fails FIVE build guards, each naming a real prerequisite —
// which is the estate working, not an obstacle. Measured by doing it, 2026-08-25:
//
//  1. TestClaimTimeoutExclusionCoversBothCompletionGates — add "required_fields_missing"
//     to livespec.ClaimedItemTimeoutExclusions AND ship a migration amending the LIVE
//     scheduled task's pre_query.
//
//     ⚠⚠ CORRECTED 2026-08-25 19:20Z — THE ORDER BELOW WAS WRONG AND WOULD HAVE BROKEN
//     THE BUILD. This block used to say the exclusions entry "MUST COME FIRST". It must
//     not, and the reason is that the lockstep fails in BOTH directions:
//         excluded ⇔ (verifier) OR (noChangeGates) OR (a REFUSING acceptancePredicateGates entry)
//     `required_fields_missing` is in NONE of the three today, so adding it to the
//     exclusions slice on its own trips the REVERSE arm — "declared excluded but NO gate
//     can grade it", the bugs_open/006 §C churn. Verified 2026-08-25 by reading all three
//     rosters; raised by the bugs_open/395 lane, which needs the same migration for
//     `content_rewrite` and hit the constraint from its own side.
//     ⚠ Note the third roster (gate 1c, acceptancePredicateGates) did not exist when this
//     file was written — 395 added it at 69479bcf6, 13:41 the same day.
//
//     THE CORRECT ORDER, and each half is placed for a different reason:
//       (a) APPLY THE MIGRATION FIRST. It is about the LIVE object, not the build, and it
//           is what stops the sweep completing past the verifier. It opens a window where
//           the live clause holds a type the Go slice does not, so --live-declaration-drift
//           will fire; that is noisy and SAFE (the sweep merely skips the type). The
//           reverse window — Go first — is the one with the actual hole in it.
//       (b) THEN, IN ONE COMMIT, the RegisterVerifier call AND the ClaimedItemTimeoutExclusions
//           entry. Either alone breaks the build, in opposite directions.
//
//     ⚠ AND IF SOMEBODY ELSE IS ADDING A TYPE TOO: livespec.Declarations pins the live
//     pre_query to ClaimedItemTimeoutExclusionClause()'s rendering of the Go slice with a
//     FragmentMatch Min:1/Max:1. Two migrations each written against today's 14-type clause
//     will NOT compose — whichever applies second must render the MERGED slice, or the
//     drift auditor fires on a correct-looking change. bugs_open/395 owes `content_rewrite`
//     on the same clause and has agreed to anchor its amendment on this one's tail.
//
//     The original reason this step leads, unchanged and still true:
//     the claimed-item-timeout sweep writes site_work_items directly, so until the
//     live clause excludes this type the sweep can auto-complete an item straight
//     past the verifier below — bugs_closed/317, reintroduced for this type by the
//     very act of adding its guard. The Go declaration alone changes nothing in
//     production.
//  2. TestOnlyTheOptedInVerifierCarriesAScopeTest — license the Grades opt-in with
//     an archive-inclusive producer-shape measurement. Already done, and it is in the
//     optedIn entry: 191 rows lifetime (site_work_items UNION site_work_items_archive,
//     2026-08-25), exactly TWO spec shapes — 188 carrying component_id+page_id+
//     page_name+slot_name+reason (the post-deploy detector), 3 without component_id
//     (the render-time producer, bugs_closed/342). The discriminator is the SPEC
//     SHAPE, never item_type and never created_by.
//  3. TestEveryItemTypeIsVerifiedOrAnAcknowledgedGap — remove the type from
//     itemTypesWithoutVerifiers, or the gap list is lying.
//  4. TestEveryPagesQueryingCheckDeclaresItsLifecyclePosture — declare this file's
//     posture. Already done: PostureObserves.
//  5. TestUnarmedCompletersOfVerifiedTypesAreAcknowledged (bugs_open/375 candidate 4)
//     — CQ-023's router has THREE unarmed `complete` arms for this type, and each
//     must either be armed or carry a recorded reason. ⚠ Arming close_converted is
//     the one CQ-023 warns fail-closes a live route; arm close_stale and close_resolved
//     first, and read that router's close paths before either.
//
// WHY IT IS NOT DONE HERE. Step 1 has to edit platform/livespec/livespec.go, and on
// 2026-08-25 another session had four hunks of in-flight, non-compiling work in that
// file. A pathspec commit of a shared file takes the other session's half-written
// work as a passenger (LANDMINES.md), so the honest move was to stop at the file
// boundary rather than ship somebody else's unfinished rename inside a verifier
// commit. The sequence above is the whole of what is left.

// gradesRequiredFieldsMissing reports whether this predicate speaks for an item.
//
// POSITIVE shape match on what the predicate resolves: a finding that names a
// stored component instance, on a page, in a slot. See the header for why a
// producer list would be the wrong instrument.
func gradesRequiredFieldsMissing(target VerifyTarget) (bool, string) {
	pageName, _ := target.Spec["page_name"].(string)
	slotName, _ := target.Spec["slot_name"].(string)
	componentID, _ := target.Spec["component_id"].(string)

	if pageName == "" || slotName == "" {
		return false, "this finding carries no page_name/slot_name, so there is no slot to re-resolve; " +
			"the predicate re-checks a stored component instance at (page, slot)"
	}
	if componentID == "" {
		return false, "this finding names no component_id, so it was not filed about a STORED component " +
			"instance — the render-time producer (work_items_common.go emitRequiredFieldsMissing, " +
			"bugs_closed/342) files about components that are not stored yet. Re-resolving a slot for such " +
			"an item would find nothing and wrongly read that absence as a repair (bugs_open/213)"
	}
	if _, ok := target.Spec["missing_fields"]; !ok {
		return false, "this finding lists no missing_fields, so there is no claim to re-check"
	}
	return true, ""
}

// VerifyRequiredFieldsMissingResolved re-runs the detector's predicate.
//
// Resolution is by (site_id, page name, slot_name) and NEVER by spec.component_id,
// which is the key CQ-023 records as unstable across rerenders: a stored id
// under-resolves against live rows (measured there — one conversion's own spec id
// was already dead while (page, slot) found the live component). component_id is
// used only by Grades above, as evidence of PROVENANCE, never as a lookup key.
func VerifyRequiredFieldsMissingResolved(ctx context.Context, db *sql.DB, target VerifyTarget,
	logger *zap.Logger) (VerifyResult, error) {

	pageName, _ := target.Spec["page_name"].(string)
	slotName, _ := target.Spec["slot_name"].(string)

	var pageID string
	var suppressed bool
	err := db.QueryRowContext(ctx, `
		SELECT p.id::text,
		       COALESCE(p.suppressed_sections, '[]'::jsonb) ? $3
		  FROM pages p
		 WHERE p.site_id = $1 AND p.name = $2
	`, target.SiteID, pageName, slotName).Scan(&pageID, &suppressed)
	if err == sql.ErrNoRows {
		return VerifyResult{Resolved: true,
			Detail: fmt.Sprintf("page %q no longer exists on this site, so nothing renders slot %q", pageName, slotName)}, nil
	}
	if err != nil {
		return VerifyResult{}, fmt.Errorf("resolve page %q: %w", pageName, err)
	}
	if suppressed {
		return VerifyResult{Resolved: true,
			Detail: fmt.Sprintf("page %q declares slot %q in suppressed_sections, so it is not rendered", pageName, slotName)}, nil
	}

	var schemaText, contentText string
	var locked, runtimeFill bool
	err = db.QueryRowContext(ctx, `
		SELECT cc.input_schema::text,
		       COALESCE(pc.content_data::text, '{}'),
		       pc.locked_at IS NOT NULL,
		       COALESCE(pc.rendered_html, '') LIKE '%data-runtime-fill%'
		  FROM page_components pc
		  JOIN content_components cc ON cc.id = pc.component_id
		 WHERE pc.page_id = $1::uuid
		   AND COALESCE(pc.slot_name, '') = $2
		   AND `+requiredFieldsComponentNotRemovedSQL+`
		   AND cc.input_schema IS NOT NULL
		 ORDER BY pc.position
		 LIMIT 1
	`, pageID, slotName).Scan(&schemaText, &contentText, &locked, &runtimeFill)
	if err == sql.ErrNoRows {
		return VerifyResult{Resolved: true,
			Detail: fmt.Sprintf("page %q exists and carries no live component at slot %q (removed, retired, or "+
				"schema-less), so no template renders an empty required field there", pageName, slotName)}, nil
	}
	if err != nil {
		return VerifyResult{}, fmt.Errorf("resolve component at (%s, %s): %w", pageName, slotName, err)
	}
	if locked {
		return VerifyResult{Resolved: true,
			Detail: fmt.Sprintf("the component at (%s, %s) is locked — the accept-as-is resolution (CQ-023); "+
				"the detector skips locked rows and would not re-file this", pageName, slotName)}, nil
	}
	if runtimeFill {
		return VerifyResult{Resolved: true,
			Detail: fmt.Sprintf("the component at (%s, %s) is a runtime-fill slot: deliberately empty at build "+
				"time, filled browser-side. The detector skips these", pageName, slotName)}, nil
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaText), &schema); err != nil {
		// Unparseable schema is NOT "resolved" — it is unreadable, which RFC_017
		// says must fail closed rather than certify.
		return VerifyResult{}, fmt.Errorf("component at (%s, %s) has an unparseable input_schema: %w", pageName, slotName, err)
	}
	fields, ok, fromLegacy := datahelpers.SchemaContentFields(schema)
	// ⚠ len(fields) == 0 IS A REFUSAL, NOT A CLEAN BILL — and this is the one line
	// where this predicate could have inverted silently.
	//
	// LANDMINES.md, "When your predicate is 'the CONFIG declares X and the DATA
	// lacks X', an unreadable config computes to HEALTHY": no schema → no required
	// fields → the "missing" list is empty → the slot reads as FILLED. The entry's
	// own footprint is this type's detector, and its warning is precisely about
	// copying the filing half's behaviour into a resolving one: *"there `continue`
	// means 'do not file' and here it must mean 'do not count as observed'."*
	//
	// The asymmetry is real and it is not theoretical. SchemaContentFields returns
	// (map, true, false) for a v2 schema whose `fields` object is EMPTY
	// (component_schema_fields.go:78 — the `fields` key is present, so the type
	// assertion succeeds and it returns early). For the DETECTOR that is harmless:
	// zero required fields means nothing to file. For a VERIFIER the identical
	// arithmetic means "nothing is missing, so the defect is repaired" — so a
	// component whose schema had its field set emptied, which is the silent-loss
	// class bugs_open/012 and /021 are about, would be CERTIFIED as fixed by the
	// very guard added to catch that. Under RFC_017 the honest answer is that
	// verification could not run.
	//
	// Caught by the council's REVISE round on corr c8ed18c1: the reuse_agent seat
	// pointed at this landmine (via a symbol, findResolvedRequiredFields, that no
	// longer exists in the tree) and the substance landed even though the pointer
	// did not.
	if !ok || len(fields) == 0 {
		return VerifyResult{}, fmt.Errorf("component at (%s, %s) declares no readable content fields "+
			"(readable=%v, declared_fields=%d), so the predicate cannot be re-run — an empty declaration "+
			"computes to \"nothing missing\" and must not be read as a repair", pageName, slotName, ok, len(fields))
	}
	if fromLegacy {
		datahelpers.WarnLegacyDialect(logger, "verify_required_fields_missing", pageName+"/"+slotName)
	}
	var contentData map[string]interface{}
	if err := json.Unmarshal([]byte(contentText), &contentData); err != nil {
		contentData = map[string]interface{}{}
	}

	// THE DETECTOR'S OWN FUNCTION. Re-implementing it here would let the two drift
	// into disagreeing about what "required" and "empty" mean, which is the
	// lockstep class this estate keeps filing bugs about.
	missing := missingRequiredValueFields(fields, contentData)
	if len(missing) == 0 {
		return VerifyResult{Resolved: true,
			Detail: fmt.Sprintf("every schema-required llm-sourced field on the component at (%s, %s) is now "+
				"populated in content_data", pageName, slotName)}, nil
	}
	return VerifyResult{Resolved: false,
		Detail: fmt.Sprintf("the component at (%s, %s) is STILL missing %d schema-required field(s): %v",
			pageName, slotName, len(missing), missing)}, nil
}
