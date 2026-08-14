// FILE: platform/orchestration/actions/revalidate_review_queue_action.go
//
// bugs_open/033 — the human-review queue has no drain. 370 items sit at
// needs_human_review (2026-07-25); not one has ever been actioned through the
// admin surface, and the surface has been visible and reachable since
// 2026-07-20. The owner ruled it a QUEUE, not a bin (2026-07-22), and ruled the
// approach: "split it — auto-drain what can be, queue the rest" (2026-07-20).
// This action is the auto-drain half.
//
// THE PROBLEM IT SOLVES IS NOT THE BACKLOG SIZE. It is that 321 of the 370
// parked items describe a page that has been REDEPLOYED since the item was
// filed, and nothing re-checks them — so a ghost and a live finding are
// byte-identical in site_work_items. Measured example: leopardessconsulting.co.uk
// /how-we-work carried two unresolved_cta items parked 2026-07-10 saying the
// hero and call-to-action had no destination for cta_url/secondary_cta_url; the
// page redeployed 2026-07-18 with every one of those fields populated. A human
// working that queue spends their attention on findings about page states that
// stopped existing months ago.
//
// WHAT IT DOES
//
//	For each parked item it looks up a REVALIDATOR by item_type and re-evaluates
//	the ORIGINAL finding against currently-deployed state. Three verdicts:
//
//	  resolved    — the condition is provably no longer true of the live page.
//	                Close it: status='complete', resolution_path='auto:revalidated',
//	                evidence into result.revalidation.
//	  still_holds — the condition is provably still true. NO status change; the
//	                item is stamped so a human reading the queue can see it was
//	                re-confirmed, and when.
//	  unknown     — cannot be determined (no deployed component, component
//	                renders from something other than content_data). Stamped with
//	                the reason; left alone.
//
//	SINCE 2026-08-06 the sweep only LOADS types it has a revalidator for, so
//	"no revalidator for the type" is no longer one of unknown's causes — an
//	unknown now always means a revalidator looked and could not tell. The
//	unjudgeable rows are counted instead, in full, into uncovered_types.
//
// WHY AUTO-CLOSING IS SAFE, AND WHY IT IS ALLOWED TO BE WRONG
//
//	Every terminal status is excluded from the idx_swi_dedup predicate, so
//	closing an item RELEASES ITS DEDUP KEY. If a 'resolved' verdict is wrong,
//	the producing check raises it again, fresh and correctly dated. A wrong
//	close costs one re-raise; it cannot lose a finding. That is what makes this
//	a reversible sweep rather than a bulk delete.
//
//	> QUALIFIED 2026-07-25 after the council gate (corr ccba9c51, bug_historian,
//	> medium) pressed on this claim — correctly: it was asserted, not measured,
//	> and it is the whole safety case. Verified since, and it does NOT hold
//	> unconditionally.
//	>
//	> All three producers re-insert with ON CONFLICT DO NOTHING on a
//	> deterministic item_key, so a terminal row does not block a re-raise
//	> (resolve_internal_links_action.go:257 for unresolved_cta; plan_sections'
//	> createDeferredItems for needs_section_data; RunDiscoveryChecks ->
//	> insertWorkItem for required_fields_missing). What does NOT hold is
//	> "the check will run again": all three fire on a PAGE BUILD or a discovery
//	> pass over that site, not on a timer. A page that is never rebuilt never
//	> re-raises, and a wrong close on such a page IS a silent loss.
//	>
//	> Also worth knowing before trusting the re-raise: across the platform's
//	> whole history only 8 items of these three types have EVER reached a
//	> terminal status (7 needs_section_data + 1 required_fields_missing; zero of
//	> the 70 unresolved_cta rows). So the re-raise path has essentially never
//	> been exercised for these types. Treat it as reasoned, not observed.
//	>
//	> The mitigation that does hold unconditionally is the audit trail: every
//	> close records the exact fields it judged populated in result.revalidation
//	> and stamps resolution_path='auto:revalidated', so a wrong close is
//	> individually identifiable and reversible by SQL (see the RUNBOOK) whether
//	> or not anything re-raises.
//
//	The asymmetry is deliberate: 'resolved' demands POSITIVE evidence that the
//	finding no longer holds. Every ambiguity resolves to 'unknown' and stays
//	queued. Guessing in the other direction would be the same failure the whole
//	bug is about.
//
// WHAT IT DELIBERATELY DOES NOT TOUCH
//   - updated_at, on anything it does not close. reconcile_superseded_reviews
//     computes "parked since" as GREATEST(created_at, updated_at); bumping
//     updated_at on every sweep would push that forward and hide genuinely
//     superseded pairs from it. The sweep timestamp lives in
//     result.revalidation.at instead. (General updated_at maintenance is
//     bugs_open/035's.)
//   - cta_names_unknown_destination. That check is mid-flight in the
//     cta_link_integrity workstream / bugs_open/023, which already knows 18 of
//     them are false positives of the excluded-area branch. Two threads on one
//     check is what that costs.
//   - The 78 items carrying an error and a machine handler_agent — failures
//     parked by FailWorkItemAction's status_override branch, which does not
//     increment attempt_count so they neither retry nor age out. Real defect,
//     open owner decision (033 D2), not this sweep's.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RevalidateReviewQueueInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{"site_id", "item_type", "max_items", "dry_run"},
	Defaults: map[string]interface{}{
		"max_items": 50,
		"dry_run":   true, // report-only unless a caller explicitly opts in
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("revalidate_review_queue", RevalidateReviewQueueInputSpec)
}

// Verdicts.
const (
	revalidationResolved   = "resolved"
	revalidationStillHolds = "still_holds"
	revalidationUnknown    = "unknown"
)

// parkedReviewItem is one row of the needs_human_review queue.
type parkedReviewItem struct {
	ID       string
	SiteID   uuid.UUID
	ItemType string
	ItemKey  string
	Spec     map[string]interface{}
	// CreatedAt is when the finding was filed. Added 2026-08-09 for the
	// claims_unverified copy-changed gate (OWNER RULING, below): a revalidator
	// that must prove the PAGE moved — not merely that the standard moved —
	// needs the date the claim was made to compare against. Other revalidators
	// ignore it; it is one more column from a row this query already reads.
	CreatedAt time.Time
}

// parkedReviewItemColumns IS the SELECT list of loadParkedReviewItems, in the
// order rows.Scan expects, and parkedReviewItemScanDests returns exactly one
// destination per entry.
//
// WHY THESE EXIST RATHER THAN A COMMENT (council round 6, `editquality` and
// `guardian` MEDIUM, independently). loadParkedReviewItems is the SHARED loader
// for every covered item_type — six as of 2026-08-12 — so a column added to the
// SELECT without a matching destination is a `sql: expected N destination
// arguments in Scan, not M` at RUNTIME, and it fails the whole sweep for EVERY
// revalidator at once, not just the one whose type prompted the edit. It is not
// a compile error and no caller can see the risk.
//
// The two halves used to be joined only by a warning comment. Both seats pointed
// out that the submission proposing that comment argued, three paragraphs
// earlier, that "a comment is not a control on a tree this many sessions share"
// (owner ruling 2026-08-02 §2) — and they were right that we had exempted
// ourselves from our own rule. Splitting the contract into two adjacent,
// countable values makes it mechanically checkable:
// TestParkedReviewItemSelectAndScanAgree fails the moment they diverge.
//
// ⚠ ADD TO BOTH, IN THE SAME POSITION. The test catches a COUNT mismatch, which
// is the failure that reaches production; it cannot catch two columns swapped
// into each other's positions when their Go types happen to be compatible.
var parkedReviewItemColumns = []string{
	"id::text",
	"site_id",
	"item_type",
	"COALESCE(item_key, '')",
	"COALESCE(spec, '{}'::jsonb)",
	"created_at",
}

func parkedReviewItemScanDests(it *parkedReviewItem, specJSON *[]byte) []interface{} {
	return []interface{}{&it.ID, &it.SiteID, &it.ItemType, &it.ItemKey, specJSON, &it.CreatedAt}
}

// revalidationVerdict is one revalidator's answer about one parked item.
//
// ARM NAMES WHICH RUNG OF THE LADDER DECIDED, and it exists because the verdict
// alone cannot say how far the decision got (2026-08-13).
//
// A revalidator is a ladder: several arms return the same `unknown` for
// completely different reasons, and the arms that matter most — the gates that
// stand between a clean scan and a closed item — sit at the BOTTOM. So the three
// counters (`resolved` / `still_holds` / `unknown`) cannot distinguish "the gate
// approved" from "the ladder stopped six rungs above the gate", and on
// claims_unverified those two states were confused for two days running. From
// the 2026-08-13 sweep: 21 items decided, every gate counter 0, and the true
// cause was that a clean scan happened 0 times in 21 — the gates were never
// reached at all. Reading it required LIKE-matching the prose of `reason`, which
// is the same fragility this lane's own landmine warns about (a grep proves
// absence only for the spelling it searches): reword a reason and every such
// query goes silently to zero.
//
// So the arm is a STABLE KEY, not prose. `GROUP BY arm` answers "where does the
// ladder stop?" mechanically, where before it took a LIKE over the reason text.
// Reaching is the observation; refusing is only one of its two outcomes, and a
// counter of refusals can never see the other one.
//
// ⚠ IT IS A SNAPSHOT, NOT A HISTORY, AND THE DISTINCTION IS THIS LANE'S OWN
// LANDMINE — walked into by the change that added the field. `result.revalidation`
// is REWRITTEN in full on every sweep (see applyRevalidation), so the arm on a
// work item is the rung that decided it on the MOST RECENT sweep and nothing
// else. It cannot answer "has a gate ever been reached": an item that reached one
// last week and now stops at scan_still_trips has had that arm overwritten. The
// landmine is filed as "a per-row revalidation stamp is LAST-WRITE-WINS, so
// GROUP BY it reads exactly like a run log and is not one" — filed by this lane
// two days before this field was written, and the council's debug_historian seat
// caught the same overclaim in the submission that shipped it.
//
// The per-RUN record is the other surface: each sweep writes its whole decision
// set to the step output, `collected_data->'sweep'->'items'` in
// orchestration_states, one row per run and never overwritten. That is where a
// rate or a trend has to come from. [MEASURED 2026-08-14: exactly ONE sweep row
// is retained (08-13 08:44:39Z), so the surface is structurally right and
// effectively empty — it becomes a history only as runs accumulate.]
//
// ⚠ IT IS AN OBSERVATION, NOT A CONTROL. Nothing branches on Arm and nothing may
// start to: an instrument that changes the outcome it measures cannot then be
// used as evidence about that outcome. Its whole value is that it is inert.
//
// COVERAGE IS PARTIAL AND SAYS SO. Only revalidateUnverifiedClaims sets it
// today. An empty Arm is therefore NOT recorded as absent — the sweep writes
// `unreported:<item_type>` instead (see the record built in Run), because "this
// revalidator has not been instrumented" and "no arm was recorded" must not
// share a spelling. That is the same rule the ladders themselves are built on,
// applied to the instrument that watches them.
type revalidationVerdict struct {
	Verdict  string                 `json:"verdict"`
	Reason   string                 `json:"reason"`
	Arm      string                 `json:"arm,omitempty"`
	Evidence map[string]interface{} `json:"evidence,omitempty"`
}

// armUnreportedPrefix marks a verdict from a revalidator that does not yet name
// its arms. It is deliberately ugly and deliberately not empty: a NULL in the
// data would be read as "no arm" when it means "nobody wrote one", and the whole
// point of the field is to stop those two being the same query result.
const armUnreportedPrefix = "unreported:"

// recordedArm is the arm as WRITTEN to the item, which is not always the arm as
// returned: an uninstrumented revalidator returns "" and is recorded as
// `unreported:<item_type>`.
//
// It is a function rather than three lines inside Run for the same reason
// unverifiedClaimsVerdict is split from its query glue — the rule is then
// testable without a database, and a rule that is only stated in a comment is
// not enforced by anything.
func recordedArm(v revalidationVerdict, itemType string) string {
	if v.Arm == "" {
		return armUnreportedPrefix + itemType
	}
	return v.Arm
}

// reviewRevalidator re-evaluates one parked item against currently-deployed state.
type reviewRevalidator func(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict

// reviewRevalidators maps item_type → revalidator. An item_type absent from this
// map yields 'unknown' and is counted in uncovered_types, so the coverage gap is
// reported rather than silently read as "nothing to drain".
//
// The three v1 entries reduce to the same question — are the fields this item
// names now populated on the currently-deployed component? — differing only in
// which spec key carries the field list.
//
// needs_page (added 2026-08-03, bugs_open/187) asks a different one: was the
// PAGE built? It was the largest uncovered type in the queue and the only one
// with a producer-side defect behind it — 28 items minted for pages that
// resolved no sections, parked for ever because nothing drained the type. Its
// implementation lives in page_section_satisfiability.go, beside the resolver
// the emit-side guards use, so the two ends of an item's life cannot answer the
// same question differently.
var reviewRevalidators = map[string]reviewRevalidator{
	// spec.missing: ["cta_url", "secondary_cta_url"]
	"unresolved_cta": revalidateNamedFields("missing"),
	// spec.missing_fields: ["headline"]
	"required_fields_missing": revalidateNamedFields("missing_fields"),
	// spec.missing: [{"field":"email","source":"site_specs.identity.email"}, …]
	"needs_section_data": revalidateNamedFields("missing"),
	// spec.page_name: "tungsten-guide"
	"needs_page": revalidateNeedsPage,
	// spec.page_id: "2c106994-…" — asks a third question: does the page still
	// trip the site's own voice gate? Its implementation re-runs the emit side's
	// scanner (discovery_checks.ScanVoiceTells) rather than restating the
	// predicate, for the same reason needs_page sits beside its resolver.
	// Added 2026-08-08: 25 items, nothing had ever closed one — see
	// revalidate_voice_tells.go for the closer census that established that.
	"voice_tells": revalidateVoiceTells,
	// spec.page_id: "2c106994-…" — asks whether the page still asserts something
	// the site's evidence_base register does not support. Re-runs the emit side's
	// scanner (discovery_checks.ScanDeployedClaims) rather than restating the
	// predicate, for the same reason voice_tells and needs_page do.
	// Added 2026-08-09: 23 items across 7 sites, ONE producer, nothing had ever
	// closed one — see revalidate_unverified_claims.go for the closer census that
	// established that, and for the correction to this lane's earlier "two
	// producers" claim.
	"claims_unverified": revalidateUnverifiedClaims,
}

// coveredItemTypes is the selection's source of truth, derived from
// reviewRevalidators so the two CANNOT disagree: registering a revalidator widens
// what the sweep loads, in the same edit, with nothing else to remember.
//
// Sorted, so the generated SQL is deterministic — Go map order is randomised per
// run, and an IN list that reshuffles every pass defeats plan caching and makes
// the query unassertable in a test.
//
// Safe to interpolate via sqlInList (which does no escaping): these are Go map
// keys — compile-time constants, never user input.
func coveredItemTypes() []string {
	out := make([]string, 0, len(reviewRevalidators))
	for t := range reviewRevalidators {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// validateTypeFilter normalises an operator-supplied item_type and REFUSES one the
// sweep has no revalidator for, rather than letting it return an empty pass.
//
// Before the selection carried a covered-type filter, asking for an uncovered type
// got you a batch of `unknown` stamps — visibly nothing useful. Now it would match
// no rows at all, and "scanned 0" reads identically to "the queue is drained". The
// operator asked a question this action cannot answer; say so, and name the set
// that CAN be answered, so the refusal is not a dead end.
//
// Pure, so both directions are testable without a database — the reason it is a
// function and not four lines inline.
func validateTypeFilter(typeFilter string) (string, error) {
	tf := strings.TrimSpace(typeFilter)
	if tf == "" {
		return "", nil
	}
	if _, ok := reviewRevalidators[tf]; !ok {
		return "", fmt.Errorf(
			"item_type %q has no revalidator, so this sweep cannot judge it; covered types are %v",
			tf, coveredItemTypes())
	}
	return tf, nil
}

// reportUncoveredBacklog counts the parked rows this sweep cannot judge, by type.
//
// It replaces counting them one row at a time inside the pass, which was wrong in
// both directions: it spent the cap on rows that could never resolve, and it
// reported only the uncovered rows that happened to fall INSIDE the cap — so the
// coverage gap looked smaller exactly when the backlog was worst.
//
// Read-only, and deliberately not allowed to fail the sweep: a broken count is a
// reporting problem, not a reason to stop draining the queue.
func reportUncoveredBacklog(ctx context.Context, db *sql.DB) (map[string]int, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT item_type, count(*)
		FROM site_work_items
		WHERE status IN (%s) AND item_type NOT IN (%s)
		GROUP BY item_type`,
		sqlInList(workItemRevalidatableStatuses), sqlInList(coveredItemTypes())))
	if err != nil {
		return nil, fmt.Errorf("count uncovered backlog: %w", err)
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var itemType string
		var n int
		if err := rows.Scan(&itemType, &n); err != nil {
			return nil, fmt.Errorf("scan uncovered backlog: %w", err)
		}
		out[itemType] = n
	}
	return out, rows.Err()
}

func RevalidateReviewQueueAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "revalidate_review_queue"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("revalidate_review_queue: no DB handle")
	}

	config := params.StepConfig.Config
	maxItems := datahelpers.GetIntField(config, "max_items", 50)
	if maxItems <= 0 {
		maxItems = 50
	}
	dryRun := true
	if d, ok := config["dry_run"].(bool); ok {
		dryRun = d
	}
	siteFilter, _ := config["site_id"].(string)
	typeFilter, _ := config["item_type"].(string)

	items, err := loadParkedReviewItems(ctx, params.DB, siteFilter, typeFilter, maxItems)
	if err != nil {
		return nil, err
	}

	// The coverage gap is now COUNTED, not accumulated from whatever fell inside
	// the cap. A failure here must not stop the drain — it is reporting.
	uncovered, uErr := reportUncoveredBacklog(ctx, params.DB)
	if uErr != nil {
		logger.Warn("could not count the uncovered backlog; the sweep continues without it",
			zap.Error(uErr))
		uncovered = map[string]int{}
	}

	sweptAt := time.Now().UTC().Format(time.RFC3339)
	counts := map[string]int{revalidationResolved: 0, revalidationStillHolds: 0, revalidationUnknown: 0}
	closed := 0
	reports := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		revalidator, covered := reviewRevalidators[item.ItemType]
		var verdict revalidationVerdict
		if !covered {
			// UNREACHABLE BY CONSTRUCTION — the selection's IN list is derived from
			// this same map. Kept, and made loud, because the two agreeing is the
			// property the whole change rests on: if this ever fires, the filter and
			// the registry have drifted apart and the sweep is judging blind.
			logger.Error("selected an item_type with no revalidator — the selection filter and "+
				"reviewRevalidators have drifted; this should be impossible",
				zap.String("item_type", item.ItemType), zap.String("item_id", item.ID))
			verdict = revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("no revalidator registered for item_type %q", item.ItemType),
			}
		} else {
			verdict = revalidator(ctx, params.DB, item, logger)
		}

		// An uninstrumented revalidator names itself rather than writing a NULL that
		// a reader would take for "no arm was reached". See revalidationVerdict.Arm.
		arm := recordedArm(verdict, item.ItemType)

		record := map[string]interface{}{
			"verdict":   verdict.Verdict,
			"reason":    verdict.Reason,
			"arm":       arm,
			"at":        sweptAt,
			"item_type": item.ItemType,
		}
		if verdict.Evidence != nil {
			record["evidence"] = verdict.Evidence
		}

		counts[verdict.Verdict]++
		reports = append(reports, map[string]interface{}{
			"item_id":   item.ID,
			"item_key":  item.ItemKey,
			"item_type": item.ItemType,
			"verdict":   verdict.Verdict,
			"reason":    verdict.Reason,
			"arm":       arm,
			"evidence":  verdict.Evidence,
		})

		if verdict.Verdict == revalidationResolved {
			logger.Info("review item auto-resolved — the finding no longer holds on the deployed page",
				zap.String("item_id", item.ID),
				zap.String("item_key", item.ItemKey),
				zap.String("item_type", item.ItemType),
				zap.String("reason", verdict.Reason),
				zap.Bool("dry_run", dryRun))
		}
		if dryRun {
			continue
		}
		didClose, wErr := applyRevalidation(ctx, params.DB, item, verdict, record)
		if wErr != nil {
			logger.Warn("could not persist revalidation; item left as-is for the next sweep",
				zap.String("item_id", item.ID), zap.Error(wErr))
			continue
		}
		if didClose {
			closed++
		}
	}

	// A FULL BATCH MEANS JUDGEABLE WORK WAS LEFT BEHIND, and silence about it is
	// how 64 rows went unreached for a fortnight. The payload has always carried
	// capped_at, but nobody compares two numbers in a JSON blob nobody reads — so
	// say it, once, at WARN, in the words of the consequence.
	capBinding := len(items) >= maxItems
	if capBinding {
		logger.Warn("revalidate_review_queue: the pass filled its cap — judgeable rows were left unexamined "+
			"and the oldest will be re-selected first next run; raise max_items or shorten the queue",
			zap.Int("scanned", len(items)), zap.Int("capped_at", maxItems))
	}

	uncoveredTotal := 0
	for _, n := range uncovered {
		uncoveredTotal += n
	}

	logger.Info("revalidate_review_queue: sweep done",
		zap.Int("scanned", len(items)),
		zap.Int("resolved", counts[revalidationResolved]),
		zap.Int("still_holds", counts[revalidationStillHolds]),
		zap.Int("unknown", counts[revalidationUnknown]),
		zap.Int("closed", closed),
		zap.Bool("cap_binding", capBinding),
		zap.Int("uncovered_backlog", uncoveredTotal),
		zap.Bool("dry_run", dryRun))

	return map[string]interface{}{
		"success":     true,
		"scanned":     len(items),
		"resolved":    counts[revalidationResolved],
		"still_holds": counts[revalidationStillHolds],
		"unknown":     counts[revalidationUnknown],
		"closed":      closed,
		"dry_run":     dryRun,
		"capped_at":   maxItems,
		"cap_binding": capBinding,
		// UNCOVERED_TYPES CHANGED MEANING 2026-08-06: it used to count the
		// unjudgeable rows that happened to fall inside the cap; it now counts every
		// parked row this sweep cannot judge, whether or not it was loaded. A reader
		// comparing runs across that date is comparing two different measurements —
		// the new one is larger, and it is the true one.
		"uncovered_types":   uncovered,
		"uncovered_backlog": uncoveredTotal,
		"items":             reports,
	}, nil
}

// loadParkedReviewItems reads the queue, oldest first — the oldest items are the
// ones most likely to be describing a page state that no longer exists.
func loadParkedReviewItems(ctx context.Context, db *sql.DB, siteFilter, typeFilter string, maxItems int) ([]parkedReviewItem, error) {
	// GATE 1 OF 3. The two write-time CAS guards in recordRevalidation must carry
	// the same list or a widened selection reports rows it then fails to update —
	// see workItemRevalidatableStatuses for why `unresolved`/`failed` are in it.
	//
	// THE TYPE FILTER IS WHAT MAKES THE CAP MEAN ANYTHING (2026-08-06). Without it
	// this selected the oldest N parked rows of ANY type, and the types with no
	// revalidator return `unknown`, which is deliberately NON-TERMINAL: they stay
	// parked, stay oldest, and are re-selected every run. Measured that day: 396 of
	// the oldest 500 were unjudgeable, so only ~104 head slots ever turned over and
	// 64 judgeable rows sat permanently beyond the cap — never reached, not reached
	// slowly. Filtering here means the cap bounds JUDGEABLE work, and a backlog
	// growing in some unrelated type can no longer push this sweep's own work out
	// of the batch. The uncovered rows are still reported, by reportUncoveredBacklog,
	// which counts ALL of them rather than the subset that fell inside the cap.
	// ⚠ THIS SELECT AND THE rows.Scan BELOW ARE ONE CONTRACT. That used to be
	// asserted only by this comment; it is now asserted by parkedReviewItemColumns
	// and parkedReviewItemScanDests, which a test compares. See their doc comment.
	query := fmt.Sprintf(`
		SELECT %s
		FROM site_work_items
		WHERE status IN (%s) AND item_type IN (%s)`,
		strings.Join(parkedReviewItemColumns, ", "),
		sqlInList(workItemRevalidatableStatuses), sqlInList(coveredItemTypes()))
	args := []interface{}{}
	if strings.TrimSpace(siteFilter) != "" {
		siteID, err := uuid.Parse(strings.TrimSpace(siteFilter))
		if err != nil {
			return nil, fmt.Errorf("invalid site_id %q: %w", siteFilter, err)
		}
		args = append(args, siteID)
		query += fmt.Sprintf(" AND site_id = $%d", len(args))
	}
	tf, tfErr := validateTypeFilter(typeFilter)
	if tfErr != nil {
		return nil, tfErr
	}
	if tf != "" {
		args = append(args, tf)
		query += fmt.Sprintf(" AND item_type = $%d", len(args))
	}
	args = append(args, maxItems)
	query += fmt.Sprintf(" ORDER BY created_at ASC LIMIT $%d", len(args))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load parked review items: %w", err)
	}
	defer rows.Close()

	var items []parkedReviewItem
	for rows.Next() {
		var it parkedReviewItem
		var specJSON []byte
		if err := rows.Scan(parkedReviewItemScanDests(&it, &specJSON)...); err != nil {
			return nil, fmt.Errorf("scan parked review item: %w", err)
		}
		if len(specJSON) > 0 {
			_ = json.Unmarshal(specJSON, &it.Spec)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// applyRevalidation writes one verdict. The status guard is what makes the sweep
// safe to run alongside a human working the same queue: if the item left
// needs_human_review between the read and the write, nothing happens.
//
// Note what is NOT set on the non-closing path: updated_at. See the file header.
func applyRevalidation(ctx context.Context, db *sql.DB, item parkedReviewItem,
	verdict revalidationVerdict, record map[string]interface{}) (bool, error) {

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return false, fmt.Errorf("marshal revalidation record: %w", err)
	}

	// GATE 2 OF 3 — record-only. CAS on the status so a row that moved underneath
	// the sweep is not written to; the list must match gate 1's.
	if verdict.Verdict != revalidationResolved {
		_, err = db.ExecContext(ctx, fmt.Sprintf(`
			UPDATE site_work_items
			SET result = COALESCE(result, '{}'::jsonb)
			          || jsonb_build_object('revalidation', $2::jsonb)
			WHERE id = $1::uuid AND status IN (%s)
		`, sqlInList(workItemRevalidatableStatuses)), item.ID, string(recordJSON))
		return false, err
	}

	// GATE 3 OF 3 — the close. Same list again: this is the one whose silent
	// disagreement with gate 1 would look like "the sweep found nothing to do".
	res, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE site_work_items
		SET status          = 'complete',
		    completed_at    = NOW(),
		    updated_at      = NOW(),
		    resolution_path = 'auto:revalidated',
		    result = COALESCE(result, '{}'::jsonb)
		          || jsonb_build_object('revalidation', $2::jsonb)
		WHERE id = $1::uuid AND status IN (%s)
	`, sqlInList(workItemRevalidatableStatuses)), item.ID, string(recordJSON))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// revalidateNamedFields builds the revalidator shared by every "these fields are
// missing" finding. specKey names the spec entry holding the field list.
//
// It keys the component lookup on (page_name, slot) and NEVER on
// spec.component_id: page_components.id is not stable across re-renders, so a
// stale component_id reads as "the section is gone" when the section is in fact
// still there under a new row id. Measured 2026-07-25: keyed on component_id,
// 30 of 30 parked needs_section_data items and 11 of 45 required_fields_missing
// items resolve to a component that no longer exists.
func revalidateNamedFields(specKey string) reviewRevalidator {
	return func(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict {
		fields := extractMissingFieldNames(item.Spec[specKey])
		if len(fields) == 0 {
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("spec.%s names no fields, so there is nothing to re-check", specKey),
			}
		}

		pageName := specString(item.Spec, "page_name", "page")
		slotName := specString(item.Spec, "slot_name", "section_name", "component_function")
		if pageName == "" || slotName == "" {
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  "spec names no page and/or no section, so the finding cannot be located on the live site",
			}
		}

		row, err := loadPageComponentBySlotRO(ctx, db, item.SiteID, pageName, NormalizeComponentFunction(slotName))
		if err == sql.ErrNoRows {
			// The section is not on the deployed page. That MIGHT mean the
			// finding is moot, but it might equally be a lookup miss — so it is
			// not positive evidence and the item stays queued.
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("no deployed component matches %s/%s", pageName, slotName),
			}
		}
		if err != nil {
			logger.Warn("revalidate: component lookup failed",
				zap.String("item_id", item.ID), zap.Error(err))
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("component lookup failed for %s/%s", pageName, slotName),
			}
		}
		if len(row.ContentData) == 0 {
			// The component renders from a template, a DERIVED source or a
			// static fallback rather than content_data. content_data cannot
			// answer the question, so we do not pretend it can.
			return revalidationVerdict{
				Verdict: revalidationUnknown,
				Reason:  fmt.Sprintf("component %s/%s carries no content_data; it renders from another source", pageName, row.SlotName),
				Evidence: map[string]interface{}{
					"page_name":         pageName,
					"slot_name":         row.SlotName,
					"page_component_id": row.ID.String(),
					"build_status":      row.BuildStatus,
				},
			}
		}

		perField := map[string]bool{}
		allPopulated := true
		for _, f := range fields {
			populated := fieldPopulated(row.ContentData[f])
			perField[f] = populated
			if !populated {
				allPopulated = false
			}
		}

		evidence := map[string]interface{}{
			"page_name":         pageName,
			"slot_name":         row.SlotName,
			"page_component_id": row.ID.String(),
			"build_status":      row.BuildStatus,
			"fields":            perField,
		}
		if allPopulated {
			return revalidationVerdict{
				Verdict:  revalidationResolved,
				Reason:   fmt.Sprintf("every field this item reports missing (%s) is populated on the deployed component", strings.Join(fields, ", ")),
				Evidence: evidence,
			}
		}
		return revalidationVerdict{
			Verdict:  revalidationStillHolds,
			Reason:   "at least one reported-missing field is still empty on the deployed component",
			Evidence: evidence,
		}
	}
}

// extractMissingFieldNames reads a spec's missing-field list in either shape the
// producers emit: a plain array of names (unresolved_cta, required_fields_missing)
// or an array of objects carrying "field" (needs_section_data). Anything else
// yields no names, which the caller reports as 'unknown' rather than guessing.
func extractMissingFieldNames(raw interface{}) []string {
	list, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var names []string
	for _, entry := range list {
		switch v := entry.(type) {
		case string:
			if name := strings.TrimSpace(v); name != "" {
				names = append(names, name)
			}
		case map[string]interface{}:
			if name, _ := v["field"].(string); strings.TrimSpace(name) != "" {
				names = append(names, strings.TrimSpace(name))
			}
		}
	}
	return names
}

// specString returns the first non-empty string value among the given spec keys.
func specString(spec map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := spec[k].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// fieldPopulated reports whether a content_data value counts as supplied.
// Absent, null, blank string, empty list and empty map are all "still missing" —
// the templates render every one of them as nothing, which is the condition the
// original findings describe.
func fieldPopulated(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(t) != ""
	case []interface{}:
		return len(t) > 0
	case map[string]interface{}:
		return len(t) > 0
	case bool:
		return t
	default:
		return true // numbers and anything else concrete count as supplied
	}
}
