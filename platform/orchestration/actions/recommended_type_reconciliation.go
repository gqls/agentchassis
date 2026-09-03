// FILE: platform/orchestration/actions/recommended_type_reconciliation.go
//
// bugs_open/428 — the strategy-to-plan reconciliation: ONE answer to "did every
// page_type this site's strategy recommended actually reach the validated plan,
// and if not, WHICH STAGE removed it and does the reason hold up?"
//
// THE INVARIANT THIS ENFORCES. A page_type named in the strategy's
// `recommended_page_types` either appears in the final validated plan, or a
// durable, structured record names the stage that removed it. Nothing else in
// the build path asserts that today: the strategy is rendered into the planner's
// prompt and never read again by code, and `strategy_notes` — the field migration
// 687 obliges the planner to justify each omission in — has NO Go reader
// anywhere. `write_site_plan_action.go` inserts `site_plans` with
// (site_id, source_agent, created_by) only, so `site_plans.notes` is never
// written either. The obligation was discharged into a channel nothing reads.
//
// WHY IT READS BOTH SIDES OF VALIDATION, WHICH IS THE LOAD-BEARING PART.
// A recommended type can be missing from the final plan for THREE different
// reasons that are indistinguishable if you only look at the output:
//
//  1. the planner never proposed it            → planner_omitted
//  2. the planner proposed it and one of this action's own identity/truncation
//     passes deleted it                        → dropped_in_validation
//  3. a source gate (444's listing arm, 450's tool arm) held it, having already
//     filed its own capability_gap             → held_by_gate
//
// Class 2 is not hypothetical and is the reason this reads three snapshots
// rather than one. [MEASURED 2026-09-03, gamedesign.uk re-plan, orchestration
// correlation 9fe9660e-7272-4f51-b968-2ff769738086]: `plan_site` emitted 9 pages
// including five `blog-post` pages on real subjects; `validate_plan` returned 4,
// with `blog-post` absent entirely. `capability_gaps_emitted: 0`, no
// `agent_error_log` row, no error, no non-zero status. The five went to
// reconcilePlanWithRealised's Pass C, whose `slugOf` takes the FIRST path
// segment — so a legitimate CHILD of a realised section index
// (/articles/<slug>.html → "articles") is indistinguishable from a flat page
// colliding with that index (/articles/index.html → "articles") and is dropped.
// The count lands in `counts.DroppedCollision`, which is logged and nowhere
// else. A comparison made against this action's OUTPUT would have reported
// "blog-post planned" and seen nothing wrong, because the type WAS planned —
// and then deleted downstream of the planner. That defect is filed separately by
// the gamedesign.uk lane; this file is the detector that makes its whole class
// visible without knowing which pass is at fault.
//
// WHY A DEFERRAL IS NOT AUTOMATICALLY A DEFECT — the discriminating case.
// The planner holds a licensed final say (migration 687 kept it deliberately),
// so "the planner left a recommended type out" is not by itself wrong. What made
// bugs_open/428's residual a defect is narrower: the planner deferred a type to a
// mechanism THAT IS NOT RUNNING. gamedesign.uk's own words, post-687 and
// compliant with 687's per-type obligation: "The blog-post type is satisfied by
// the blog infrastructure; individual posts are not planned as static pages
// here." The mechanism it names is real and wired — `blog-content-planner` via
// `create_blog_posts` — and has not run since 2026-04-24 (bugs_open/460). So the
// obligation was met with a reason that is false as applied, and nothing checked
// it. This file checks it: a deferral to a producer this estate can see RUNNING
// is recorded as sound and files no work item; a deferral to one that is not is
// the finding.
//
// POLICY: FAIL OPEN, AND RECORD ONLY. This never changes the plan, never drops
// or adds a page, and never fails the step. Every error path logs and returns.
// The work items it files are RFC_056 `filing_mode: record` verdicts — parked by
// construction, visible in the admin Review Queue bugs_open/428 built, and
// dispatchable by nobody. That is deliberate and is the same ruling that shaped
// the rest of this bug: an opinion about what a site OUGHT to have must not
// auto-dispatch a page rewrite (bugs_closed/238, finetuning.uk).
//
// ⚠ SHIPS ARMED, WITH A KILL SWITCH. No step-config key: `validate_site_plan`
// already carries nine optional keys and has no ActionInputSpec, so a tenth
// would be invisible to WFA-013's optional-key budget while making this
// detector's arming yet another thing a migration has to remember (the estate's
// own landmine: a detector sat off for nine days after its blocker cleared).
// `DISABLE_RECOMMENDED_TYPE_RECONCILIATION` disarms it fleet-wide with no build,
// following the same convention as the owned-page door in
// load_work_item_actions.go — a lever to pull in anger, not a default-OFF switch
// that rots unexercised.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// recommendedTypeDormantAfter is how long a registered producer may go without
// running before a deferral to it stops counting as sound.
//
// 30 days, and the number is doing less work than it looks: the case this
// exists for is four months stale, and a producer that ran this month is
// unambiguously running. Anything between "a fortnight" and "a quarter" would
// classify every case on file identically. It is a const rather than a config
// key for the reason in the file header — arming surface is the cost here, not
// tuning.
const recommendedTypeDormantAfter = 30 * 24 * time.Hour

// recommendedTypeReconciliationKillSwitch disarms this whole file with no build.
const recommendedTypeReconciliationKillSwitch = "DISABLE_RECOMMENDED_TYPE_RECONCILIATION"

// omissionClass says which stage removed a recommended page_type. The value is
// written into spec.omission_class and is the field a fleet census groups on —
// "the planner declined it" and "we deleted it after the planner proposed it"
// are different bugs with different owners, and before this they produced
// identical evidence.
type omissionClass string

const (
	// omissionPlannerOmitted: the type is absent from the planner's own output.
	omissionPlannerOmitted omissionClass = "planner_omitted"
	// omissionDroppedInValidation: the planner proposed it and this action's
	// identity/truncation passes removed it before any gate ran.
	omissionDroppedInValidation omissionClass = "dropped_in_validation"
	// omissionHeldByGate: a source gate held it, and has filed its own
	// capability_gap naming the enablement. Recorded, never re-filed.
	omissionHeldByGate omissionClass = "held_by_gate"
)

// Finding codes. Declared in
// docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json
// — an undeclared code fails findingcodes_scan_test.go by design (bugs_open/358).
const (
	findingRecommendedTypeReconciled     = "RECOMMENDED_TYPE_RECONCILED"
	findingRecommendedTypePlannerOmitted = "RECOMMENDED_TYPE_PLANNER_OMITTED"
	findingRecommendedTypeDropped        = "RECOMMENDED_TYPE_DROPPED_IN_VALIDATION"
	findingRecommendedTypeHeldByGate     = "RECOMMENDED_TYPE_HELD_BY_GATE"
	findingRecommendedTypeDeferredLive   = "RECOMMENDED_TYPE_DEFERRED_TO_LIVE_PRODUCER"
)

// externalProducer names the mechanism OUTSIDE a site plan that can create pages
// of a page_type — the thing a planner is implicitly pointing at when it defers
// a type "to the blog infrastructure".
//
// ⚠ THIS NAMES THE PRODUCER OF THE TYPE, NOT EVERY WRITER OF SUCH A PAGE, and
// the two have come apart in practice. The bugs_open/450 lane measured 36 writes
// across six seotools tool-page shells in one hour, none of them from
// `tool-deployer`: they came from `needs_content_page` → `page-build-handler`,
// minted by `rerender_single_page_action` and `tool-generator`. So a LIVE verdict
// here answers "can this type still arrive by its own route?" and says nothing
// whatever about whether generic writers are also reaching such pages. Do not
// widen the reading; a second question needs a second instrument.
type externalProducer struct {
	// AgentType is the agent_definitions.type whose runs are the liveness signal.
	AgentType string
	// DrivenBy names what dispatches it, so a dormant verdict has somewhere to go.
	DrivenBy string
	// Note is what a reader of a finding needs in order to act.
	Note string
}

// externalPageTypeProducers is deliberately SMALL and deliberately not derived
// from builderForPageType.
//
// builderForPageType answers "which handler BUILDS a page of this type once it
// is in the plan". This map answers a different question — "what could create
// such a page WITHOUT the planner" — and the two disagree on exactly the entries
// that matter. `blog-post` routes to page-build-handler there (it builds pages
// already planned) while its only non-planner producer is `blog-content-planner`;
// `tool` has NO builder there at all yet has a live external producer. Deriving
// one from the other would collapse that distinction and answer the wrong
// question in both directions. An entry here is a claim that a page of this type
// can appear on a site that never planned one — add one only with the row that
// proves it.
var externalPageTypeProducers = map[string]externalProducer{
	"blog-post": {
		AgentType: "blog-content-planner",
		DrivenBy:  "discovery check empty_blog (check_empty_blog.go) files needs_blog_posts at this handler",
		Note:      "creates blog-post page rows directly via create_blog_posts; bugs_open/460 records that it ran 14 times to 2026-04-24 and has been silent since",
	},
	"tool": {
		AgentType: "tool-deployer",
		DrivenBy:  "the design rotation, on its own schedule — not from the site plan",
		Note:      "creates its own tool page rows under its own names; bugs_open/450 §7 measured that nothing reads PLANNED tool pages to decide what to build",
	},
}

// producerLiveness is what agent_run_stats can honestly say about a producer.
//
// agent_run_stats is the instrument with a durable memory — upserted at every
// orchestration START on the resolved agent type and NOT pruned — which is why
// it is used here in preference to the two rolling windows that have misled
// three sessions on this exact question in one week: `orchestration_states`
// spans ~24 hours, and `site_work_items` archives closed rows out of the table
// you queried (all 14 of blog-content-planner's driver items are archive-only,
// so the live table says the producer never ran). `llm_call_log` reaches further
// back but has a live deletion function (cleanup_old_llm_logs) in pg_proc, so an
// all-history read from it can silently become a windowed one.
//
// ⚠ IT IS FORWARD-ONLY, and this struct carries TrackingSince so no reader can
// forget that. A producer that last ran BEFORE tracking began has no row, and
// "never ran" and "ran before we were counting" are the same evidence — which is
// exactly blog-content-planner's position today. Every finding this file writes
// states the window rather than asserting the stronger claim.
type producerLiveness struct {
	Registered    bool
	HasRow        bool
	RunCount      int64
	LastRan       time.Time
	TrackingSince time.Time
	// Verdict is one of: unregistered, unreadable, never_since_tracking,
	// dormant, live.
	Verdict string
}

// Live reports whether a deferral to this producer is currently sound. Only a
// row with a recent run qualifies: every other state, including "we cannot
// tell", counts against the deferral, because the cost of a wrong "live" is a
// site that ships without the pages and nothing recorded, while the cost of a
// wrong "not live" is one parked record row a person reads.
func (l producerLiveness) Live() bool { return l.Verdict == "live" }

// recommendedType is one entry of the strategy's recommended_page_types.
type recommendedType struct {
	PageType  string
	Reasoning string
}

// plannerTypeDecision is the planner's own structured statement about one
// recommended type it did not plan.
//
// Read from the optional `page_type_decisions` array in the plan JSON. NOTHING
// about presence is taken from it — whether a type is in the plan is computed
// from `pages`, here, every time. The planner's word is used for one thing only:
// WHICH PRODUCER it thinks will supply the type, so the claim can be checked
// against whether that producer runs. This split is the whole design: the model
// states intent, the framework verifies the fact. Absent field = the check still
// runs, with stated_reason falling back to a substring read of strategy_notes.
type plannerTypeDecision struct {
	PageType   string
	Decision   string
	Reason     string
	DeferredTo string
}

// recommendedTypeOmission is one recommended page_type that did not reach the
// final plan, with everything a reader needs to act on it.
type recommendedTypeOmission struct {
	PageType      string
	Class         omissionClass
	Reasoning     string   // the STRATEGY's reason for recommending it
	ProposedPages []string // page names that carried the type before it was removed
	StatedReason  string   // the planner's own reason, structured or scraped
	ReasonSource  string   // "page_type_decisions" | "strategy_notes" | "none"
	ClaimedTo     string   // the producer the planner named, if any
	Producer      externalProducer
	Liveness      producerLiveness
}

// recommendedTypeReconciliation is what one reconciliation did.
//
// Returned rather than kept internal so a test can assert a NEGATIVE. The two
// cases this file most needs to prove — a gate hold files no second row, and a
// deferral to a LIVE producer files none either — are cases where nothing
// happens, and "nothing happened because the rule held" is indistinguishable
// from "nothing happened because the code never got there" if the only evidence
// is an absent mock expectation (an unexpected Begin makes sqlmock error, this
// file logs and carries on, and the test passes for the wrong reason). Asserting
// on Omissions and GapsFiled makes the negative real.
type recommendedTypeReconciliation struct {
	// Skipped names the fail-open branch taken, or "" if the check ran.
	Skipped       string
	Recommended   []recommendedType
	Omissions     []recommendedTypeOmission
	Substitutions []string
	// PresentButFewer: recommended types still in the plan whose page count fell.
	// The blind spot of a type-level check, kept countable rather than implied.
	PresentButFewer map[string]interface{}
	GapsFiled       int
	// FindingsAttempted/FindingsRecorded mirror LogActionFindings, so a lost
	// durable row never reads as a written one.
	FindingsAttempted int
	FindingsRecorded  int
}

// reconcileRecommendedPageTypes is the whole of this file's entry point. It is
// called at the very end of ValidateSitePlanAction, after both source gates, so
// `final` is the page set the plan writer will actually receive.
//
// proposed = the planner's own output, snapshotted before any pass ran.
// preGate  = the page set after this action's identity/truncation/component
// passes and before the tool and listing gates.
// final    = what survives everything.
func reconcileRecommendedPageTypes(ctx context.Context, params ActionParams, plan map[string]interface{}, proposed, preGate, final []planPageView) recommendedTypeReconciliation {
	logger := params.Logger
	if logger == nil {
		return recommendedTypeReconciliation{Skipped: "no_logger"}
	}
	if os.Getenv(recommendedTypeReconciliationKillSwitch) != "" {
		logger.Info("recommended_type_reconciliation: disarmed by kill switch — no reconciliation performed",
			zap.String("env", recommendedTypeReconciliationKillSwitch))
		return recommendedTypeReconciliation{Skipped: "kill_switch"}
	}
	if params.DB == nil {
		logger.Warn("recommended_type_reconciliation: no DB — skipped (fail-open)")
		return recommendedTypeReconciliation{Skipped: "no_db"}
	}

	recommended := recommendedTypesFrom(params.CollectedData)
	if len(recommended) == 0 {
		// Not a defect: many sites have no strategy aspect, or one that names no
		// page types. Logged rather than recorded — the audit row below exists to
		// make a ZERO-omission run distinguishable from a run that never
		// reconciled, and writing one here would blur that line in the other
		// direction (a site with no recommendations is not a reconciled site).
		logger.Info("recommended_type_reconciliation: strategy names no recommended_page_types — nothing to reconcile")
		return recommendedTypeReconciliation{Skipped: "no_recommendations"}
	}

	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		logger.Warn("recommended_type_reconciliation: no parseable site id — skipped (fail-open)",
			zap.String("site_id", siteIDStr))
		return recommendedTypeReconciliation{Skipped: "no_site_id", Recommended: recommended}
	}

	proposedTypes := pageTypeIndex(proposed)
	preGateTypes := pageTypeIndex(preGate)
	finalTypes := pageTypeIndex(final)
	decisions := plannerTypeDecisionsFrom(plan)
	notes := strings.ToLower(planStringField(plan, "strategy_notes"))

	liveness := map[string]producerLiveness{}
	var omissions []recommendedTypeOmission
	var substitutions []string
	shrink := map[string]interface{}{}

	for _, rec := range recommended {
		if n := len(finalTypes[rec.PageType]); n > 0 {
			// THE STATED BLIND SPOT, made measurable rather than merely admitted.
			// This check is TYPE-level: it asks whether the type reached the plan,
			// not whether every proposed page of that type did. So on an
			// ESTABLISHED site — where Pass A's union restores realised pages the
			// LLM omitted — a type keeps its presence even when new children of
			// it were dropped, and this loop goes quiet. That asymmetry is
			// precisely why bugs_open/463's defect hid from 2026-05-21: an
			// established site restores its children and looks healthy, and only a
			// site with none (gamedesign.uk) shows the drop at type level.
			// Recording the shrink here does not turn it into a finding — the
			// preserve/union machinery legitimately reshapes page sets on every
			// re-plan, so a per-page rule would be noise — but it puts the number
			// where a fleet census can find it, so the population this check
			// cannot see is countable instead of hypothetical.
			if proposedN := len(proposedTypes[rec.PageType]); proposedN > n {
				shrink[rec.PageType] = map[string]interface{}{"proposed": proposedN, "final": n}
			}
			continue
		}
		// A section-index-family sibling standing in for the recommended type is
		// a substitution, not an omission — gamedesign.uk planned a section-index
		// "articles-index" where the strategy said blog-index, and said so. The
		// family test is datahelpers.IsSectionIndexRole, the estate's ONE
		// definition of that vocabulary (the architecture seat has already ruled
		// against a second copy of it), so this cannot drift from the listing
		// gate's own notion of the family. Recorded in the audit row's context,
		// deliberately not as a finding: it is a fact worth being able to census,
		// not a defect.
		if datahelpers.IsSectionIndexRole(rec.PageType) && anySectionIndexPresent(finalTypes) {
			substitutions = append(substitutions, rec.PageType)
			continue
		}

		om := recommendedTypeOmission{
			PageType:  rec.PageType,
			Reasoning: rec.Reasoning,
		}
		switch {
		case len(preGateTypes[rec.PageType]) > 0:
			om.Class = omissionHeldByGate
			om.ProposedPages = preGateTypes[rec.PageType]
		case len(proposedTypes[rec.PageType]) > 0:
			om.Class = omissionDroppedInValidation
			om.ProposedPages = proposedTypes[rec.PageType]
		default:
			om.Class = omissionPlannerOmitted
		}

		if d, ok := decisions[rec.PageType]; ok {
			om.StatedReason = d.Reason
			om.ClaimedTo = d.DeferredTo
			om.ReasonSource = "page_type_decisions"
		} else if notes != "" && strings.Contains(notes, rec.PageType) {
			// 687's obligation, discharged unstructured. All six post-687
			// plan_site calls named every omitted type this way, so the fallback
			// is the common path until the structured field ships — but it can
			// only establish that the type was MENTIONED, never what was claimed
			// about it, which is precisely the gap the structured field closes.
			om.StatedReason = "named in strategy_notes (unstructured — reason not machine-readable)"
			om.ReasonSource = "strategy_notes"
		} else {
			om.ReasonSource = "none"
		}

		// Liveness is asked of the producer the PLANNER named if it named one,
		// otherwise of the registry's producer for the type. The planner's claim
		// is checked in preference to ours precisely because the claim is the
		// thing under test.
		producerKey := om.ClaimedTo
		if p, ok := externalPageTypeProducers[rec.PageType]; ok {
			om.Producer = p
			if producerKey == "" {
				producerKey = p.AgentType
			}
		}
		if producerKey != "" {
			if l, seen := liveness[producerKey]; seen {
				om.Liveness = l
			} else {
				l = readProducerLiveness(ctx, params.DB, producerKey, logger)
				liveness[producerKey] = l
				om.Liveness = l
			}
		} else {
			om.Liveness = producerLiveness{Verdict: "unregistered"}
		}

		omissions = append(omissions, om)
	}

	sort.Slice(omissions, func(i, j int) bool { return omissions[i].PageType < omissions[j].PageType })

	findings := []agenterrors.Finding{recommendedTypeAuditFinding(recommended, omissions, substitutions, shrink)}
	for _, om := range omissions {
		findings = append(findings, om.finding())
	}
	attempted, recorded := LogActionFindings(ctx, params, siteID.String(), "", "validate_plan", findings, logger)
	warnUnrecordedDrops(attempted, recorded, logger)

	filed := 0
	for _, om := range omissions {
		if !om.actionable() {
			continue
		}
		if fileRecommendedTypeGap(ctx, params, siteID, om) {
			filed++
		}
	}

	logger.Info("recommended_type_reconciliation: complete",
		zap.Int("recommended", len(recommended)),
		zap.Int("omitted", len(omissions)),
		zap.Int("family_substitutions", len(substitutions)),
		zap.Int("gap_items_filed", filed))

	return recommendedTypeReconciliation{
		Recommended:       recommended,
		Omissions:         omissions,
		Substitutions:     substitutions,
		PresentButFewer:   shrink,
		GapsFiled:         filed,
		FindingsAttempted: attempted,
		FindingsRecorded:  recorded,
	}
}

// actionable reports whether this omission earns a parked work item.
//
// The two silences are deliberate. A type HELD BY A GATE already has a
// capability_gap from that gate naming the enablement — filing a second row for
// the same fact is the duplicate-producer shape this estate keeps paying for. A
// type DEFERRED TO A LIVE PRODUCER is a sound decision: the pages can still
// arrive, and a row saying otherwise would be a false positive that trains
// readers to ignore the queue. Everything else is recorded for a person.
func (om recommendedTypeOmission) actionable() bool {
	if om.Class == omissionHeldByGate {
		return false
	}
	if om.Class == omissionPlannerOmitted && om.Liveness.Live() {
		return false
	}
	return true
}

// builderNeeded is the slug a capability_gap consumer groups on
// (diagnose_triage's roadmap grouping and fixloop_digest both read
// spec.builder_needed). It names WHAT IS MISSING, which for a page the planner
// proposed and validation deleted is not a builder at all.
func (om recommendedTypeOmission) builderNeeded() string {
	if om.Class == omissionDroppedInValidation {
		return "plan_page_identity"
	}
	if om.Producer.AgentType != "" {
		return om.Producer.AgentType
	}
	if om.ClaimedTo != "" {
		return om.ClaimedTo
	}
	return "planner_must_plan:" + om.PageType
}

func (om recommendedTypeOmission) code() string {
	switch om.Class {
	case omissionDroppedInValidation:
		return findingRecommendedTypeDropped
	case omissionHeldByGate:
		return findingRecommendedTypeHeldByGate
	default:
		if om.Liveness.Live() {
			return findingRecommendedTypeDeferredLive
		}
		return findingRecommendedTypePlannerOmitted
	}
}

func (om recommendedTypeOmission) severity() string {
	switch om.code() {
	case findingRecommendedTypeHeldByGate, findingRecommendedTypeDeferredLive:
		return "info"
	default:
		return "warning"
	}
}

func (om recommendedTypeOmission) message() string {
	switch om.Class {
	case omissionDroppedInValidation:
		return fmt.Sprintf("strategy-recommended page_type %q was PLANNED (%s) and removed during validation before any source gate ran",
			om.PageType, strings.Join(om.ProposedPages, ", "))
	case omissionHeldByGate:
		return fmt.Sprintf("strategy-recommended page_type %q was planned (%s) and held by a source gate, which filed its own capability_gap",
			om.PageType, strings.Join(om.ProposedPages, ", "))
	default:
		if om.Liveness.Live() {
			return fmt.Sprintf("strategy-recommended page_type %q was not planned; deferred to %q, which is running (%d runs, last %s)",
				om.PageType, om.livenessSubject(), om.Liveness.RunCount, om.Liveness.LastRan.Format(time.RFC3339))
		}
		return fmt.Sprintf("strategy-recommended page_type %q was not planned and no running producer can supply it (%s: %s)",
			om.PageType, om.livenessSubject(), om.Liveness.Verdict)
	}
}

func (om recommendedTypeOmission) livenessSubject() string {
	if om.ClaimedTo != "" {
		return om.ClaimedTo
	}
	if om.Producer.AgentType != "" {
		return om.Producer.AgentType
	}
	return "no registered producer"
}

// finding builds the durable row.
//
// ⚠ THE ERROR CODE IS NAMED AS A FILE-SCOPE CONST IN EACH ARM, and this switch
// exists ONLY for that. The obvious one-liner — `ErrorCode: om.code()` — is a
// CallExpr, which findingcodes_scan_test.go cannot resolve, and it says so
// rather than guessing: "this site is INVISIBLE to the scan". A code the source
// scan cannot see is one only the daily live-table check can catch, and then
// only after it has already fired in production — which is bugs_open/358's whole
// case (LINK_CONTEXT_UNAVAILABLE reached the live table past exactly this blind
// spot). Keeping code() for the switch's own logic and the tests, while naming
// the consts literally here, satisfies both readers. Do not collapse this back.
func (om recommendedTypeOmission) finding() agenterrors.Finding {
	f := agenterrors.Finding{
		Severity: om.severity(),
		Message:  om.message(),
		Context:  om.context(),
	}
	switch om.code() {
	case findingRecommendedTypeDropped:
		f.ErrorCode = findingRecommendedTypeDropped
	case findingRecommendedTypeHeldByGate:
		f.ErrorCode = findingRecommendedTypeHeldByGate
	case findingRecommendedTypeDeferredLive:
		f.ErrorCode = findingRecommendedTypeDeferredLive
	default:
		f.ErrorCode = findingRecommendedTypePlannerOmitted
	}
	return f
}

func (om recommendedTypeOmission) context() map[string]interface{} {
	c := map[string]interface{}{
		"page_type":          om.PageType,
		"omission_class":     string(om.Class),
		"strategy_reasoning": truncateForRecord(om.Reasoning, 400),
		"stated_reason":      truncateForRecord(om.StatedReason, 400),
		"reason_source":      om.ReasonSource,
		"producer_verdict":   om.Liveness.Verdict,
		"builder_needed":     om.builderNeeded(),
	}
	if len(om.ProposedPages) > 0 {
		c["proposed_pages"] = om.ProposedPages
	}
	if om.ClaimedTo != "" {
		c["claimed_producer"] = om.ClaimedTo
	}
	if om.Producer.AgentType != "" {
		c["registered_producer"] = om.Producer.AgentType
		c["producer_driven_by"] = om.Producer.DrivenBy
	}
	if om.Liveness.HasRow {
		c["producer_run_count"] = om.Liveness.RunCount
		c["producer_last_ran"] = om.Liveness.LastRan.Format(time.RFC3339)
	}
	if !om.Liveness.TrackingSince.IsZero() {
		// Stated on every row that consults liveness, because agent_run_stats is
		// forward-only: without this a "never ran" reads as an all-history claim
		// it cannot support.
		c["liveness_tracking_since"] = om.Liveness.TrackingSince.Format(time.RFC3339)
	}
	return c
}

// recommendedTypeAuditFinding is the row every reconciled run writes, clean ones
// included.
//
// It exists so that "no omissions" and "the reconciliation never ran" stop being
// the same evidence — the failure mode this whole file is about, one level up.
// `info` severity is sanctioned for exactly this shape by agenterrors.Finding's
// own contract.
func recommendedTypeAuditFinding(recommended []recommendedType, omissions []recommendedTypeOmission, substitutions []string, shrink map[string]interface{}) agenterrors.Finding {
	types := make([]string, 0, len(recommended))
	for _, r := range recommended {
		types = append(types, r.PageType)
	}
	omitted := make([]string, 0, len(omissions))
	byClass := map[string]int{}
	for _, om := range omissions {
		omitted = append(omitted, om.PageType)
		byClass[string(om.Class)]++
	}
	ctx := map[string]interface{}{
		"recommended_types": types,
		"recommended_count": len(recommended),
		"omitted_count":     len(omissions),
		"by_class":          byClass,
	}
	if len(omitted) > 0 {
		ctx["omitted_types"] = omitted
	}
	if len(substitutions) > 0 {
		ctx["family_substitutions"] = substitutions
	}
	if len(shrink) > 0 {
		// Types still PRESENT whose page count fell between the planner's
		// proposal and the final plan — the population this type-level check
		// cannot flag, kept countable. See the loop for why it is not a finding.
		ctx["present_but_fewer_pages"] = shrink
	}
	return agenterrors.Finding{
		ErrorCode: findingRecommendedTypeReconciled,
		Severity:  "info",
		Message: fmt.Sprintf("reconciled %d strategy-recommended page_type(s) against the validated plan: %d absent, %d satisfied by a section-index sibling",
			len(recommended), len(omissions), len(substitutions)),
		Context: ctx,
	}
}

// fileRecommendedTypeGap files ONE parked capability_gap for an omitted type,
// through the SHARED work-item writer.
//
// SHAPE, and why each half is what it is:
//   - item_type `capability_gap` REUSED, not invented. This is the third
//     producer of that type (after builderForPageType's unavailable-builder arm
//     and the two source gates), which under the owner ruling of 2026-08-02 §1
//     needs no RFC provided the producer set and the shared item_key shape are
//     named in the concept-register entry — they are, under BLD-030.
//   - item_key `recommended_type_gap:<page_type>` — per site, per type, ONE row
//     however many pages carried it. Deliberately NOT the gates'
//     `capability_gap:<type>:<page>` key: theirs is per PAGE and answers "this
//     page cannot be built", this one is per TYPE and answers "this site has no
//     pages of a type its own strategy asked for". Co-dedupping them would make
//     one silently swallow the other.
//   - handler_agent EMPTY and status `deferred` — the bugs_closed/078/291
//     convention; a named-but-nonexistent handler livelocks the dispatcher.
//   - filing_mode `record` (RFC_056) so it appears in the admin Review Queue
//     this bug already built, and so no promoter can ever dispatch it.
//   - and NO routed_handler / routed_status, which is the honest part: there is
//     no automatic repair to release. HandleReleaseRecordVerdict requires both
//     fields non-empty, so its WHERE clause refuses these rows by construction
//     rather than by intention.
//
// Best-effort: a failed write must not change a plan this action has already
// decided on, but it must not be silent either.
func fileRecommendedTypeGap(ctx context.Context, params ActionParams, siteID uuid.UUID, om recommendedTypeOmission) bool {
	spec := map[string]interface{}{
		"gap_kind":       "recommended_page_type_absent",
		"page_type":      om.PageType,
		"omission_class": string(om.Class),
		"builder_needed": om.builderNeeded(),
		"reason":         om.message(),
		"filing_mode":    "record",
		"not_dispatchable": "status 'deferred' + empty handler_agent — deliberate (bugs_closed/078/291 convention). " +
			"This row carries NO routed_handler/routed_status because there is no automatic repair for it: " +
			"HandleReleaseRecordVerdict therefore refuses it by construction, and that is the intended behaviour, not a gap.",
		"deferred_reason": "filing_mode=record (RFC_056): a verdict about what this site's strategy asked for and did not get. " +
			"An opinion about a site's shape must not auto-dispatch a page rewrite (bugs_closed/238).",
		"needs_human": "read the omission_class: 'dropped_in_validation' is a platform defect on this plan, " +
			"'planner_omitted' is a planning decision whose stated producer is not running.",
	}
	for k, v := range om.context() {
		if _, taken := spec[k]; !taken {
			spec[k] = v
		}
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		params.Logger.Error("recommended_type_reconciliation: spec marshal failed",
			zap.String("page_type", om.PageType), zap.Error(err))
		return false
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		params.Logger.Error("recommended_type_reconciliation: gap tx open failed",
			zap.String("page_type", om.PageType), zap.Error(err))
		return false
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "validate_site_plan",
		pipeline:     "build",
		itemType:     "capability_gap",
		severity:     "low",
		summary:      fmt.Sprintf("[verdict, not dispatched] Strategy recommends page_type '%s' and the plan has none — %s", om.PageType, om.Class),
		spec:         string(specJSON),
		priority:     200,
		handlerAgent: "",
		status:       "deferred",
		createdBy:    "validate_site_plan",
		itemKey:      "recommended_type_gap:" + om.PageType,
	}, params.Logger)
	if err != nil {
		_ = tx.Rollback()
		params.Logger.Error("recommended_type_reconciliation: gap insert failed",
			zap.String("page_type", om.PageType), zap.Error(err))
		return false
	}
	if err := tx.Commit(); err != nil {
		params.Logger.Error("recommended_type_reconciliation: gap commit failed",
			zap.String("page_type", om.PageType), zap.Error(err))
		return false
	}
	if !inserted {
		params.Logger.Info("recommended_type_reconciliation: gap already on file (dedup)",
			zap.String("page_type", om.PageType))
	}
	return inserted
}

// readProducerLiveness asks agent_run_stats what it can honestly say about one
// producer, and returns the tracking window with every answer.
//
// The LEFT JOIN against a one-row aggregate is what makes "no row for this
// agent" and "the table is empty" distinguishable: the first returns a row with
// NULL run stats and a real tracking_since, the second returns a row with both
// NULL, and only the second is genuinely unreadable.
func readProducerLiveness(ctx context.Context, db *sql.DB, agentType string, logger *zap.Logger) producerLiveness {
	l := producerLiveness{Registered: true, Verdict: "unreadable"}
	var runCount sql.NullInt64
	var lastRan, trackingSince sql.NullTime
	err := db.QueryRowContext(ctx, `
		SELECT s.run_count, s.last_ran_at, t.tracking_since
		FROM (SELECT min(first_ran_at) AS tracking_since FROM agent_run_stats) t
		LEFT JOIN agent_run_stats s ON s.agent_type = $1
	`, agentType).Scan(&runCount, &lastRan, &trackingSince)
	if err != nil {
		logger.Warn("recommended_type_reconciliation: producer liveness read failed — verdict 'unreadable'",
			zap.String("agent_type", agentType), zap.Error(err))
		return l
	}
	if trackingSince.Valid {
		l.TrackingSince = trackingSince.Time
	}
	if !runCount.Valid {
		// No row: the producer has not run since tracking began. NOT "never ran"
		// — agent_run_stats is forward-only and blog-content-planner's real last
		// run (2026-04-24) predates it, which is why the finding always carries
		// liveness_tracking_since alongside this verdict.
		l.Verdict = "never_since_tracking"
		return l
	}
	l.HasRow = true
	l.RunCount = runCount.Int64
	if lastRan.Valid {
		l.LastRan = lastRan.Time
	}
	if time.Since(l.LastRan) > recommendedTypeDormantAfter {
		l.Verdict = "dormant"
		return l
	}
	l.Verdict = "live"
	return l
}

// recommendedTypesFrom reads the strategy's recommended_page_types out of the
// collected data.
//
// The path is `site_specs.specs.strategy.recommended_page_types` — the
// read_specs step's all-aspects output, where specs[aspect] is the entire
// site_specs.data object. This is the SAME data the planner's prompt renders,
// which is the point: the check reads the planner's own input rather than a
// second copy that could drift from it.
func recommendedTypesFrom(collected map[string]interface{}) []recommendedType {
	raw := datahelpers.ExtractNestedField(collected, "site_specs.specs.strategy.recommended_page_types")
	entries, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	out := make([]recommendedType, 0, len(entries))
	for _, e := range entries {
		var pt, reasoning string
		switch v := e.(type) {
		case string:
			pt = v
		case map[string]interface{}:
			pt, _ = v["page_type"].(string)
			reasoning, _ = v["reasoning"].(string)
		}
		pt = normalizePageType(strings.TrimSpace(pt))
		if pt == "" || seen[pt] {
			continue
		}
		seen[pt] = true
		out = append(out, recommendedType{PageType: pt, Reasoning: reasoning})
	}
	return out
}

// plannerTypeDecisionsFrom reads the optional structured decision list.
// Tolerant by design: an absent or malformed field degrades to no decisions,
// never to an error, because the presence computation does not depend on it.
func plannerTypeDecisionsFrom(plan map[string]interface{}) map[string]plannerTypeDecision {
	out := map[string]plannerTypeDecision{}
	entries, ok := plan["page_type_decisions"].([]interface{})
	if !ok {
		return out
	}
	for _, e := range entries {
		m, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		pt, _ := m["page_type"].(string)
		pt = normalizePageType(strings.TrimSpace(pt))
		if pt == "" {
			continue
		}
		d := plannerTypeDecision{PageType: pt}
		d.Decision, _ = m["decision"].(string)
		d.Reason, _ = m["reason"].(string)
		d.DeferredTo, _ = m["deferred_to"].(string)
		d.DeferredTo = strings.TrimSpace(d.DeferredTo)
		out[pt] = d
	}
	return out
}

// pageTypeIndex maps each page_type present in a page set to the names carrying
// it. Empty-typed pages are ignored rather than defaulted: a default here would
// invent a presence the plan never stated.
func pageTypeIndex(views []planPageView) map[string][]string {
	idx := map[string][]string{}
	for _, v := range views {
		if v.Role == "" {
			continue
		}
		idx[v.Role] = append(idx[v.Role], v.Name)
	}
	return idx
}

// anySectionIndexPresent reports whether the plan carries any page of the
// section-index family.
func anySectionIndexPresent(types map[string][]string) bool {
	for role := range types {
		if datahelpers.IsSectionIndexRole(role) {
			return true
		}
	}
	return false
}

// planStringField reads a top-level string field off the plan map.
func planStringField(plan map[string]interface{}, key string) string {
	s, _ := plan[key].(string)
	return s
}

// truncateForRecord bounds a free-text field going into a durable row. The
// strategy's per-type reasoning runs to several hundred words and the whole of
// it is already in site_specs; what belongs on the finding is enough to
// recognise which recommendation this was.
func truncateForRecord(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// planPageViewsOf snapshots a page slice as the shared plan-page view. Taken at
// three points in ValidateSitePlanAction; each snapshot must be a COPY, because
// the passes downstream mutate the maps in place.
func planPageViewsOf(pages []interface{}) []planPageView {
	views := make([]planPageView, 0, len(pages))
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			views = append(views, pageViewFromMap(pm))
		}
	}
	return views
}
