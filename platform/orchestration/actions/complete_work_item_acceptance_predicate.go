// FILE: platform/orchestration/actions/complete_work_item_acceptance_predicate.go
//
// Completion gate 1c: before an item is stamped 'complete', re-evaluate the item's
// OWN stated acceptance criterion — the machine-checkable half of it — against the
// page as it stands now.
//
// WHAT AN ACCEPTANCE PREDICATE IS, in plain terms, because nothing else on this
// path deals in one. A work item is one recorded defect on one page. The producer
// that filed it also wrote, in prose, what would make the page right
// (`spec.acceptance_test`). Since 2026-08-24 one producer additionally writes a
// small STRUCTURED condition beside that prose (`spec.acceptance_predicate`, register
// CLM-024) — "the meta description must state X before any count" — which
// `verify_acceptance_predicates` PROVED FALSE of the page at the moment the finding
// was made, and which it refuses to store otherwise.
//
// WHY THIS EXISTS (bugs_open/395). `complete` means "the handler reported success",
// and until now nothing on that path read the criterion. So a handler that rebuilds
// a page and changes something OTHER than what the finding asked for closes the item
// exactly as a correct repair would. That is not a theory: on this producer's FIRST
// live run, webdesign.co.uk's index finding was dispatched to page-build-handler,
// which rebuilt and deployed the page (commit ee88ba3c, 2026-08-24 22:25Z) and closed
// the item `complete`; the page was rebuilt again on 2026-08-25; and its own stored
// predicate STILL REFUTES, because the meta description it names never changed.
// Pinned by TestTheFirstLiveEmittedPredicatesStillRefuteAfterTheFix.
//
// ⚠ THIS IS NOT gate 1b (complete_work_item_no_change.go) AND NOT ITS DUPLICATE.
// 1b refuses a completion whose handler reports it changed NOTHING. Here the handler
// changed something real — a rebuild and a deploy, with a commit sha — so 1b is right
// not to fire. The two are complements: 1b catches "the handler did nothing", 1c
// catches "the handler did something else".
//
// ⚠ AND THE REASON GIVEN FOR ITS PLACEMENT IN bugs_open/395 §4 IS THE WRONG ONE,
// recorded here as a correction rather than repeated. That candidate says a verifier
// cannot do this because "verifyBeforeComplete's VerifyTarget carries the SPEC, not
// the RESULT". True, and it is gate 1b's argument — 1b needs the handler's reply.
// Gate 1c needs the SPEC and the current page row, both of which a verifier has. The
// real reasons it is a gate are: GetVerifier is ONE verifier per item_type, a scarce
// shared slot on a type (`content_rewrite`) many producers file into; and the gates
// here compose, so a third opt-in one costs nothing to the types that have not asked
// for it.
//
// WHERE IT SITS: between gate 1b and gate 2. Gate 1c grades the item's OWN stated
// criterion; gate 2 grades the TYPE's generic predicate. If the item's own criterion
// is still false, producing a generic `verified` that contradicts the item's own
// terms serves nobody. It is also the cheaper of the two for most types — one
// page-metadata read, no HTTP, no browser, no page fetch, which is the standing
// objection recorded three times in verifier_coverage_test.go against every option
// that computes a rendered property.
//
// OPT-IN PER item_type WITH THE UNSAFE DEFAULT OFF (owner ruling 2026-08-02 §2), and
// THREE-VALUED for the same reason unreadableOutcome is: the zero value must not be a
// policy. A type absent from acceptancePredicateGates takes a map miss and this file
// changes nothing about it — byte-identical to today. An item whose spec carries no
// predicate is likewise untouched, which is nearly all of them.
//
// EVERY EVALUATED PREDICATE LEAVES A VERDICT ON THE ROW, INCLUDING `holds`, and that
// is the instrument this whole slice rests on rather than decoration. Without a
// recorded `holds`, a gate that permits is indistinguishable from a gate that never
// ran — which is exactly the shape of the residual the record-only roster entry
// carries (see acceptancePredicateGates). Recording it turns "this gate has only ever
// seen failures" from a worry into a query.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
)

// predicateGateOutcome declares, per item_type, what a STILL-REFUTING predicate
// means for that type at completion time.
//
// WHY IT IS PER-TYPE AND THREE-VALUED. `predicateRecords` and `predicateRefuses` are
// unsafe in different directions: refusing is new blocking authority on a shared
// completion path, and recording is the false green bugs_open/395 was filed about.
// Neither may be an author-time silent default, which is what the third value buys —
// exactly the shape complete_work_item_no_change.go's unreadableOutcome settled on
// after bugs_open/302.
type predicateGateOutcome int

const (
	// predicateUndeclared is the zero value and is NOT a policy: the roster test
	// fails an entry carrying it, and at runtime it records rather than blocks, so
	// an entry written by somebody who never read this comment cannot start
	// refusing completions by accident.
	predicateUndeclared predicateGateOutcome = iota

	// predicateRecords stamps the verdict on the row and lets the completion
	// proceed. The item still closes green; what changes is that the estate can
	// now COUNT the false greens instead of finding them by reading pages.
	predicateRecords

	// predicateRefuses blocks the completion and routes the item into the attempt
	// machinery, as for a persisting defect. Requires RefusalWhy, and — because
	// the claimed-item-timeout sweep writes site_work_items directly and NO
	// completion gate runs for it — requires the type to be declared in
	// livespec.ClaimedItemTimeoutExclusions. That second requirement is a BUILD
	// FAILURE, not a memo: see TestClaimTimeoutExclusionCoversBothCompletionGates.
	predicateRefuses
)

// acceptancePredicateRule declares, for ONE item_type, that its items' predicates
// may be graded at completion, and what a refutation means.
type acceptancePredicateRule struct {
	// Why the type opts in: what licenses re-grading this type's own criterion at
	// completion. Surfaced to whoever reads a graded item.
	Why string

	// OnRefuted is the type's declaration. The zero value records; the roster test
	// refuses to let an entry ship without stating one.
	OnRefuted predicateGateOutcome

	// RefusalWhy is the measurement licensing predicateRefuses, and is required for
	// it. A SEPARATE field from Why on purpose: Why licenses "this type's criterion
	// is worth re-grading", which is a claim about the producer. This licenses
	// "a still-refuting criterion here cannot be a repair", which is a claim about
	// what this type's handlers actually do — a different question with different
	// evidence, and the one that has to survive a blocked completion.
	RefusalWhy string

	// PromotionOwes, when non-empty, DECLARES what is still missing before this
	// entry may move from predicateRecords to predicateRefuses. It is the honest
	// statement of a record-only entry's debt — the LicenceVoided shape from
	// noChangeGates, pointed forward instead of backward — so a record-only arm
	// cannot quietly become permanent by nobody remembering it was meant to be
	// temporary. The roster test requires it on every recording entry.
	PromotionOwes string
}

// acceptancePredicateGates is the opt-in roster. Absent item_type → this file is
// inert for it.
//
// ── WHY content_rewrite AND WHY IT RECORDS RATHER THAN REFUSES ──────────────
// It is the type the only predicate-writing producer files into, and the type the
// worked case in bugs_open/395 sits on. It records, in this first cut, for three
// reasons, of which the first is decisive:
//
//  1. THERE IS NO NEGATIVE CONTROL. [MEASURED 2026-08-25] all three live predicates
//     refute, and no row exists anywhere where a predicate is SATISFIED after its
//     fix. A gate that has only ever seen failures cannot be distinguished from one
//     that refuses everything (bugs_open/395 §6, which states this is not optional).
//     The control exists at the gate — a live predicate against a satisfying string
//     returns `holds` and permits, TestAPredicateThatHoldsPermitsTheCompletion — and
//     that is a unit, not a run. The two must not be confused for one another.
//  2. Refusing needs a SECOND live change on an object this lane does not own: a
//     migration amending the claimed-item-timeout sweep's live pre_query, on a type
//     carrying 1,637 completions all-history [MEASURED 2026-08-25, site_work_items
//     UNION site_work_items_archive].
//  3. Recording produces the blast-radius census bugs_open/395 §5 calls the first
//     job, on a queryable surface (result->'_verification'->'acceptance_predicate'),
//     which today can only be estimated by reading pages.
//
// ⚠ THE COST OF THAT CHOICE, NAMED: this is a third instance of CLM-023's residual —
// an enforcement arm proven by mutation-tested units and never fired in production.
// Do not quote a clean run as evidence that the refusal arm works. What stops it
// becoming permanent is PromotionOwes below plus the build failure the promotion
// itself triggers.
var acceptancePredicateGates = map[string]acceptancePredicateRule{
	"content_rewrite": {
		Why: "the offer-analyser is the one producer that writes a machine-checkable half of its own " +
			"acceptance test (CLM-024), and it files into this type; on its first live run the index " +
			"finding was rebuilt, deployed and closed `complete` while its own predicate still refuted " +
			"(bugs_open/395, webdesign.co.uk 2026-08-24 22:25Z, commit ee88ba3c)",
		OnRefuted: predicateRecords,
		PromotionOwes: "a NEGATIVE CONTROL that is a live row, not a unit: one item whose predicate is " +
			"SATISFIED after its fix and which completed green through this gate. [MEASURED 2026-08-25] " +
			"all three live predicates refute and no such row exists, so promoting today would arm a " +
			"refusal that has never been observed to permit anything. Promotion ALSO requires adding " +
			"'content_rewrite' to livespec.ClaimedItemTimeoutExclusions and shipping the migration that " +
			"amends the live pre_query — the claimed-item-timeout sweep writes the row directly and no " +
			"completion gate runs for it, so without that the refusal is bypassed by a timeout " +
			"(bugs_closed/317). The lockstep test makes that half a build failure; this sentence is the " +
			"half no test can assert.",
	},
}

// acceptancePredicateGateFor is the roster lookup, indirected through a package
// variable so a test can drive the REFUSING arm end to end without mutating the
// roster map itself.
//
// That is not a stylistic preference, and it is the same reason verifierLookup
// exists one file along. The roster today holds no refusing entry — deliberately,
// there is no live negative control — so the ONLY way to prove that a refusal
// actually reaches the caller and returns mayComplete=false is a fixture. Mutating
// acceptancePredicateGates directly to supply one would be visible to
// TestClaimTimeoutExclusionCoversBothCompletionGates, which reads that map and
// asserts a refusing entry implies a claim-timeout exclusion: a test adding a
// synthetic refusing type would break that guard, or worse, pass while it happened
// to run first. Indirecting here keeps the fixture local and leaves the roster the
// single source of truth for every other reader.
//
// Production NEVER re-points this: TestAcceptancePredicateLookupIsNotASwitchInProduction
// asserts there is one assignment in the package's non-test source, the declaration
// below.
var acceptancePredicateGateFor = func(itemType string) (acceptancePredicateRule, bool) {
	rule, ok := acceptancePredicateGates[itemType]
	return rule, ok
}

// acceptancePredicateGateOutcome is gate 1c's verdict.
type acceptancePredicateGateOutcome int

const (
	// predicateGatePass — not this gate's business: the type has not opted in, or
	// the item carries no predicate, or the predicate is satisfied.
	predicateGatePass acceptancePredicateGateOutcome = iota

	// predicateGateRecorded — a predicate was evaluated and the verdict is worth
	// stamping on the row. Completion proceeds. Covers BOTH a satisfied predicate
	// (the negative control's evidence) and a still-refuting one on a recording
	// type (the false green, now counted).
	predicateGateRecorded

	// predicateGateBlocked — a predicate still refutes and the type declares that
	// this cannot be a repair.
	predicateGateBlocked

	// predicateGateBlind — the type opted in, the item carries a predicate, and
	// this gate COULD NOT EVALUATE IT. Completion proceeds and the caller records
	// it loudly; see acceptancePredicateNote for why this is not silent.
	predicateGateBlind
)

// acceptancePredicateNote is the "opted in but could not grade it" observation,
// carried out to the caller rather than recorded where it is detected — the same
// shape, and for the same reason, as noChangeAbstention: the item_type travels WITH
// the observation instead of the caller re-reading it from the database and creating
// two sources of truth for one fact inside a single completion.
type acceptancePredicateNote struct {
	ItemType string
	Detail   string
}

// gradeAcceptancePredicate is gate 1c.
//
// Returns the payload to merge into result._verification (nil when this gate has
// nothing to say) and the outcome.
//
// THE ORDER OF THE EARLY EXITS IS THE OPT-IN GUARANTEE, not stylistic. The roster
// lookup is FIRST, so a type that has not opted in never reaches the spec decode, a
// database read, or a log line — it is byte-identical to today's behaviour, which is
// what "the unsafe default is OFF" has to mean to be worth anything.
func gradeAcceptancePredicate(ctx context.Context, db *sql.DB, itemType string, specJSON []byte,
	siteID uuid.UUID, logger *zap.Logger) (map[string]interface{}, acceptancePredicateGateOutcome) {

	rule, opted := acceptancePredicateGateFor(itemType)
	if !opted {
		return nil, predicateGatePass
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		// Deliberately NOT a block, and deliberately not silent either. An
		// unparseable spec is gate 2's problem to fail closed on if a verifier is
		// registered (runRegisteredVerifier says why); making it newly blocking HERE
		// would be a behaviour change beyond what opting into predicate grading
		// asks for. It is still this gate going blind, so it is recorded.
		return blindPayload(itemType, "the item's spec did not decode, so its predicate could not be read: "+err.Error()),
			predicateGateBlind
	}

	stored, ok := spec[acceptancePredicateKey].(map[string]interface{})
	if !ok || len(stored) == 0 {
		// The common case by a wide margin, and NOT a gap: most items of this type
		// come from producers that write no predicate at all. Nothing to grade, so
		// nothing to say — recording a note here would bury the real signal under
		// one row per completion.
		return nil, predicateGatePass
	}

	page := predicatePageName(stored, spec)
	if page == "" {
		return blindPayload(itemType, "the stored predicate names no page and the item's spec carries none, so there is nothing to evaluate it against"),
			predicateGateBlind
	}

	// ⚠ THE SAME QUERY THE EMIT SIDE USED, deliberately, and this is load-bearing:
	// the model authored the predicate against the page list it was shown, so
	// evaluating over a DIFFERENT population would let a predicate be graded against
	// a page that was not on the surface, or reported unevaluable for one that was.
	// One function, so the two ends of a predicate's life cannot answer the same
	// question differently — the reason revalidateVoiceTells re-runs the emit side's
	// own scanner rather than restating its predicate.
	subjects, err := loadAcceptancePredicateSubjects(ctx, db, siteID)
	if err != nil {
		return blindPayload(itemType, "the page surface could not be loaded, so no predicate could be evaluated: "+err.Error()),
			predicateGateBlind
	}

	subject, found := subjects[page]
	if !found {
		// Phrased on the emit side's own rule (two council seats, corr ef482d1c): an
		// EMPTY subject set must blame this gate, not the named page. No page is on
		// the surface in that case, and sending the next reader to the model when
		// the fault is the query's is how a gate goes silently inert while looking
		// like it is working.
		detail := fmt.Sprintf("page %q is not on this site's active surface, so the item's predicate could not be evaluated", page)
		if len(subjects) == 0 {
			detail = "no pages were loaded for this site at all, so NO predicate could be evaluated — this is a fault in this gate's page query or its site_id, not in the predicate"
		}
		return blindPayload(itemType, detail), predicateGateBlind
	}

	// ⚠ predicateForEvaluation, NEVER the stored map. The stored form carries the
	// emission provenance keys, and EvaluateAcceptancePredicate enforces a closed key
	// set — so evaluating it verbatim returns `inapplicable` for EVERY live
	// predicate, and this gate would read as permanently blind while appearing to
	// run. See emissionProvenanceKeys.
	verdict, reason := EvaluateAcceptancePredicate(predicateForEvaluation(stored), subject)

	payload := map[string]interface{}{
		"item_type": itemType,
		"page":      page,
		"verdict":   string(verdict),
		"detail":    reason,
		// The verdict this predicate carried when it was WRITTEN, echoed beside the
		// verdict it has NOW. A refutation that was already there is the false green;
		// the two together are what make that readable without a second query.
		"verdict_at_emission": stored[emissionVerdictKey],
	}

	switch verdict {
	case PredicateHolds:
		// The permit arm, and it is RECORDED rather than passed silently: this is
		// the only evidence that distinguishes a gate that grades and permits from
		// one that never ran. See the file header.
		payload["outcome"] = "permitted"
		return payload, predicateGateRecorded

	case PredicateInapplicable:
		// A stored predicate was evaluable at emission BY CONSTRUCTION — the emit
		// gate refuses to store one that is not. So it being unevaluable now means
		// the page changed shape, the vocabulary moved, or this gate has gone blind,
		// and the third wears the same face as the first two.
		return blindPayload(itemType, "the item's stored predicate is no longer evaluable, although it was at emission: "+reason),
			predicateGateBlind
	}

	// PredicateRefutes. What it MEANS is the type's own declaration, not this
	// function's to assume.
	payload["why_opted_in"] = rule.Why
	if rule.OnRefuted == predicateRefuses {
		payload["outcome"] = "blocked"
		payload["refusal_why"] = rule.RefusalWhy
		logger.Warn("gradeAcceptancePredicate: the item's own acceptance criterion still refutes — blocking completion",
			zap.String("item_type", itemType), zap.String("page", page), zap.String("detail", reason))
		return payload, predicateGateBlocked
	}

	// predicateRecords, and predicateUndeclared which the roster test forbids and
	// the runtime treats as recording — the safety property that an entry added
	// without reading the declaration cannot start blocking completions.
	payload["outcome"] = "recorded_only"
	payload["promotion_owes"] = rule.PromotionOwes
	logger.Warn("gradeAcceptancePredicate: the item's own acceptance criterion still refutes, and this type only RECORDS — completing anyway",
		zap.String("item_type", itemType), zap.String("page", page), zap.String("detail", reason),
		zap.String("remedy", "this is the false green of bugs_open/395, now counted: query result->'_verification'->'acceptance_predicate'"))
	return payload, predicateGateRecorded
}

// blindPayload is the shape recorded when this gate could not grade what it was
// asked to grade. Built here rather than at four call sites so every cause reaches
// the row wearing the same key names and a census over them is one query.
func blindPayload(itemType, detail string) map[string]interface{} {
	return map[string]interface{}{
		"item_type": itemType,
		"verdict":   string(PredicateInapplicable),
		"outcome":   "not_evaluable",
		"detail":    detail,
	}
}

// recordAcceptancePredicateBlindSpot persists a gate-1c blind spot to
// agent_error_log, for the same reason recordUnknownNoChangeShape does: a zap.Warn
// lives in an ephemeral pod log that does not survive a rollout, so a guard going
// blind would leave no queryable trace.
//
// ⚠ AND THIS ONE IS THE OBJECTION TWO COUNCIL SEATS RAISED AGAINST THE EMIT SIDE
// (corr ef482d1c, editquality + debug_historian), answered in the same shape: a page
// query that matches nothing makes EVERY predicate unevaluable, and "the gate found
// nothing to grade" is byte-identical to the acceptable outcome "no item carried a
// predicate". A measurement that the hazard is not live today is not a guard against
// it, so the blind case is made LOUD rather than left to a count that reads clean.
//
// Severity 'warning' — the completion itself is legitimate under the conservative
// rule; what needs attention is that this gate could not judge what it was handed.
// Best-effort by design: failing to record must never block a completion the gate
// has already allowed.
//
// ⚠ DO NOT PUT THE WORDS `output`, `response`, `content`, `text`, `body`, `raw`,
// `prompt`, `completion` OR `partial` ANYWHERE IN THE zap.Warn BELOW — not in a
// string, not in a comment inside it. `pattern-check.py`'s `logged-model-output`
// rule runs PAYLOAD_NAME over the raw statement text for six lines from the log
// sink, skipping only up to the first comma, so it cannot tell a payload being
// LOGGED from an ordinary English word being TYPED. This call logs no model data
// at all (an id, a type, a gate-authored detail, a fixed remedy) and drew the
// finding twice: once for a remedy ending "NOT the model's output", and then
// AGAIN for the comment explaining that — which is why the explanation lives up
// here, outside the six-line window, instead of beside the line it describes.
// Recorded in LANDMINES.md so the rule can be narrowed rather than dodged.
func recordAcceptancePredicateBlindSpot(ctx context.Context, params ActionParams, itemID uuid.UUID,
	note acceptancePredicateNote, logger *zap.Logger) {

	logger.Warn("gradeAcceptancePredicate: the item's stored acceptance predicate could not be evaluated — completing, but this gate could not judge it",
		zap.String("item_id", itemID.String()),
		zap.String("item_type", note.ItemType),
		zap.String("detail", note.Detail),
		zap.String("remedy", "if this fires on every item of the type, suspect the page surface query in loadAcceptancePredicateSubjects or the emission-provenance strip in predicateForEvaluation, rather than what the model wrote"))

	if params.DB == nil {
		return
	}

	LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
		WorkItemID:   itemID.String(),
		Action:       "complete_work_item",
		ErrorMessage: "acceptance-predicate gate could not evaluate the item's own stored criterion for item_type '" + note.ItemType + "' — item completed ungraded by this gate: " + note.Detail,
		ErrorCode:    "ACCEPTANCE_PREDICATE_NOT_EVALUABLE",
		Severity:     "warning",
		Context: map[string]interface{}{
			"item_type": note.ItemType,
			"guard":     "gradeAcceptancePredicate",
			"remedy": "a stored predicate was evaluable at emission by construction, so this is a change in the page, " +
				"the vocabulary, or this gate. If EVERY item of the type reports it, the fault is here: check " +
				"loadAcceptancePredicateSubjects' surface query and predicateForEvaluation's strip (bugs_open/395, CLM-024)",
		},
	}, logger)
}
