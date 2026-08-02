// FILE: platform/orchestration/actions/prune_floor.go
//
// A shared FLOOR for reconciliation prunes — the "delete everything this run did
// not re-write" pattern (bugs_closed/135).
//
// The pattern itself is correct and load-bearing: a writer re-upserts what it saw
// and then deletes the remainder, so rows for things that no longer exist do not
// accumulate. It is catastrophic in exactly one case — the run did NOT see the
// whole corpus — and nothing at any of the call sites checked that it had. From
// 016b §9: "the identical DELETE+INSERT rebuild pattern destroyed a working A*
// pathfinding game; recurred independently on a second site months later." The
// index case is worse than the site case, because an emptied index answers
// "absent" confidently rather than visibly breaking (bugs_closed/108).
//
// WHAT THIS FILE IS: the DECISION RULE and its refusal text, as pure functions of
// counts. It is deliberately free of SQL and of any one table's shape, because the
// measurement is per-table and the rule is not. Callers supply cohorts —
// (label, confirmed-by-this-run, stored) triples — and get a verdict plus a
// refusal sentence that names its own remedy.
//
// WHY COHORTS, not one total. A single whole-corpus ratio is not sensitive enough:
// a run that re-confirms 95% of the rows can still have dropped 100% of one class
// (one symbol kind, one nav group, one page's links), and the total hides it
// completely. So the caller partitions: one cohort per class it can lose
// independently, plus — where it has one — a whole-corpus cohort in a DIFFERENT
// unit (e.g. distinct paths seen, rather than rows written), which is the signal
// that says "this run did not see the corpus" rather than "this run wrote less".
//
// WHY A RESOLVABLE FLOOR. A large legitimate deletion is a real event (a big
// refactor merged upstream, a site's nav genuinely halved). A guard with no exit
// is a defect in a guard's costume — the council's `guardian` seat, round 5 of the
// 7ba5b8c4 trail. So the ratio is step config, 0 disables it entirely, and the
// refusal text states that remedy rather than leaving the operator to find it.
//
// CONSUMERS (all four reconciliation deletes are now guarded — bugs_closed/165
// closed the list this comment used to hold open):
//   - index_code_symbols            code_symbols_actions.go        (bugs_closed/135)
//   - save_page_sections            save_sections_prune_floor.go   (165 site A)
//   - populate_nav_tables           nav_prune_floor.go             (165 site B)
//   - extract_and_sync_links        link_registry_prune_floor.go   (165 site C)
//
// Each supplies its OWN cohorts. That is the load-bearing part of the design and
// the one thing a new consumer must not copy: the rule is shared, the partition
// is not. Both sites that measured before choosing found the partition their own
// bug file had proposed was wrong on the data — A's per-slot_name (998 of 1,009
// groups hold one row, so every legitimate single-section removal scores 0%) and
// B's per-nav-group (the classifier legitimately RE-HOMES a page between groups,
// so a re-homing reads as a 100% loss of one class). Measure, then partition.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// defaultPruneFloorRatio is the fraction of a cohort's STORED rows that this run
// must have re-confirmed before the cohort may be pruned. 0.5 is the value the
// 7ba5b8c4 council trail settled on: high enough that a half-blind run is caught,
// low enough that ordinary churn (and any refactor short of halving a class)
// passes untouched.
const defaultPruneFloorRatio = 0.5

// pruneCohort is one independently-losable slice of the corpus.
//
//	Confirmed — rows this run re-wrote (i.e. rows the prune will KEEP)
//	Stored    — rows present for the cohort now (Confirmed + what the prune would DELETE)
//
// Both counts MUST be measured after the run's writes and with the exact
// complement of the DELETE's own predicate, or the guard is judging a different
// population from the one being deleted.
type pruneCohort struct {
	Label     string
	Confirmed int
	Stored    int
}

// ratio is the share of the cohort this run re-confirmed. An empty cohort has
// nothing to lose, so it reads as fully confirmed rather than as 0/0 — a new
// class appearing for the first time must never be able to refuse a prune.
func (c pruneCohort) ratio() float64 {
	if c.Stored <= 0 {
		return 1
	}
	return float64(c.Confirmed) / float64(c.Stored)
}

// cohortVerdict is a cohort with its ratio and the floor's judgement on it.
type cohortVerdict struct {
	pruneCohort
	Ratio float64
	Below bool
}

// pruneFloorVerdict is the whole decision. Allowed is the only field a caller
// must act on; the rest exists so the refusal can be REPORTED rather than merely
// taken (a refusal nobody can see is the 034/076 shape: "no error anywhere"
// usually means no error surface).
type pruneFloorVerdict struct {
	Floor    float64         // the floor actually applied, after clamping
	Asked    float64         // what the caller passed, before clamping
	Clamped  bool            // Asked was outside (0,1] and was corrected
	Disabled bool            // floor <= 0: the guard was switched off by config
	Allowed  bool            // may the prune proceed?
	Cohorts  []cohortVerdict // every cohort, worst ratio first
	Failing  []cohortVerdict // the subset below the floor, worst first
}

// evaluatePruneFloor judges a set of cohorts against a floor.
//
// The prune is refused when ANY cohort is below the floor. It is all-or-nothing
// on purpose: a partial prune ("delete the kinds that passed") is harder to
// reason about, and the loss it avoids is small — a refused prune only retains
// rows that a later healthy run will prune anyway, because the delete is defined
// against the CURRENT commit and not against this run. A refusal is therefore
// self-healing; a wrong delete is not.
//
// Boundary: a cohort exactly AT the floor passes (Below is ratio < floor), so a
// floor of 1.0 means "every stored row must be re-confirmed" and never refuses a
// run that saw everything.
func evaluatePruneFloor(floor float64, cohorts []pruneCohort) pruneFloorVerdict {
	v := pruneFloorVerdict{Asked: floor, Floor: floor, Allowed: true}

	if floor <= 0 {
		// Explicitly disabled. Allowed, but recorded as disabled so that "no
		// refusal" can never be mistaken for "the guard looked and was content".
		v.Floor = 0
		v.Disabled = true
		v.Clamped = floor < 0
		return v
	}
	if floor > 1 {
		// A ratio above 1 is unsatisfiable — it would refuse every prune for
		// ever, which reads exactly like a working guard. Clamp and say so.
		v.Floor = 1
		v.Clamped = true
	}

	for _, c := range cohorts {
		r := c.ratio()
		cv := cohortVerdict{pruneCohort: c, Ratio: r, Below: r < v.Floor}
		v.Cohorts = append(v.Cohorts, cv)
		if cv.Below {
			v.Failing = append(v.Failing, cv)
			v.Allowed = false
		}
	}

	sort.SliceStable(v.Cohorts, func(i, j int) bool { return v.Cohorts[i].Ratio < v.Cohorts[j].Ratio })
	sort.SliceStable(v.Failing, func(i, j int) bool { return v.Failing[i].Ratio < v.Failing[j].Ratio })
	return v
}

// Reason is the human sentence for the decision — the text that goes in the log,
// in the durable note and in the action's own result. op names the operation
// ("index_code_symbols: prune"), subject names what was being pruned
// ("gqls/agentchassis at a1b2c3d"), configKey names the knob that resolves it.
//
// It always states the numbers and always states the remedy. A refusal a reader
// cannot act on gets overridden by someone deleting the guard.
//
// aftermath IS REQUIRED, AND IT IS THE ONE CLAUSE THAT IS NOT A PROPERTY OF THE
// FLOOR. It says what the caller did with the rows it declined to prune, and the
// four consumers genuinely differ:
//
//   - 135 (index_code_symbols) SKIPS ONLY THE PRUNE. The run's own writes stand,
//     the unconfirmed rows are kept, and a later run that sees the whole corpus
//     does tidy them up.
//   - A, B and C REFUSE THE WHOLE OPERATION — delete and insert are one
//     transaction, so refusing half would double the rows. Nothing is written,
//     nothing is tidied later, and the artefact already stored still stands.
//
// This used to be a fixed clause carrying 135's aftermath ("the rows this run
// did not confirm are retained and a later run that sees the whole corpus will
// prune them"), which made the sentence FALSE for three of the four consumers:
// it told those operators to wait for a clean-up that will never come. Observed
// live on oufe.com, 2026-08-02, in the site B induction (bugs_closed/165). It is a
// required parameter rather than an optional one precisely so a fifth consumer
// cannot inherit somebody else's aftermath by saying nothing — the failure mode
// here is a default that is silently wrong, not a caller who forgets.
//
// An EMPTY aftermath degrades to the shorter sentence rather than to a default,
// because "NOTHING was deleted." on its own is true for every possible consumer
// and a borrowed clause is not.
func (v pruneFloorVerdict) Reason(op, subject, configKey, aftermath string) string {
	switch {
	case v.Disabled:
		return fmt.Sprintf("%s: floor DISABLED by config (%s=0) for %s — pruning whatever this run did not re-confirm, unchecked",
			op, configKey, subject)
	case v.Allowed:
		return fmt.Sprintf("%s: floor cleared for %s (%s=%.2f); %s", op, subject, configKey, v.Floor, v.cohortList(v.Cohorts))
	default:
		tail := ""
		if aftermath != "" {
			tail = "; " + aftermath
		}
		return fmt.Sprintf("%s: REFUSED for %s — this run re-confirmed too little of what is stored (%s=%.2f): %s. "+
			"NOTHING was deleted%s. "+
			"If the shrinkage is genuine, re-run with %s lowered (0 disables the floor entirely).",
			op, subject, configKey, v.Floor, v.cohortList(v.Failing), tail, configKey)
	}
}

// cohortList renders cohorts as "kind=func 38%% (1160 of 3048)", worst first and
// capped, so a refusal naming twenty cohorts stays readable in a log line.
func (v pruneFloorVerdict) cohortList(cs []cohortVerdict) string {
	const max = 6
	if len(cs) == 0 {
		return "no cohorts measured (nothing stored to protect)"
	}
	parts := make([]string, 0, max+1)
	for i, c := range cs {
		if i == max {
			parts = append(parts, fmt.Sprintf("… and %d more", len(cs)-max))
			break
		}
		parts = append(parts, fmt.Sprintf("%s %.0f%% (%d of %d)", c.Label, 100*c.Ratio, c.Confirmed, c.Stored))
	}
	return strings.Join(parts, ", ")
}

// Detail renders every cohort for the action result, so the numbers behind a
// decision survive in orchestration_states rather than only in a log line the
// retention window eats. Reported on PASS as well as on refusal: candidate (3) of
// bugs_closed/135 — "pruned: 4000" as a bare success counter is the alarm presented
// as output, and the fix is to publish the denominator beside it.
func (v pruneFloorVerdict) Detail() []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(v.Cohorts))
	for _, c := range v.Cohorts {
		out = append(out, map[string]interface{}{
			"cohort":    c.Label,
			"confirmed": c.Confirmed,
			"stored":    c.Stored,
			"ratio":     round2(c.Ratio),
			"below":     c.Below,
		})
	}
	return out
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

// ---------------------------------------------------------------------------
// The parts every consumer needs and none of which are table-specific.
//
// Added 2026-07-31 for bugs_closed/165 sites B and C. Site A
// (save_sections_prune_floor.go) invented both of these inline, and B and C
// were each about to spell them again. Three spellings of one durable surface
// is the exact drift class this council reviews for, so they move here.
//
// ADDITIVE AND INERT: nothing reaches either function until a caller names it,
// and no existing consumer's behaviour changes. Under the owner ruling of
// 2026-07-29 §1 that makes this a normal council-gate change rather than
// architecture scope — what the shared mechanism GUARANTEES is untouched; only
// its surface area grows.
// ---------------------------------------------------------------------------

// pruneFloorStatus is the value reported under "completeness_status". A bare
// count is ambiguous between "nothing to prune" and "we refused to prune"
// (bugs_closed/165 says this in terms), so the status is always emitted alongside
// the numbers rather than left to be inferred from them.
const (
	pruneStatusPassed   = "passed"
	pruneStatusDisabled = "floor_disabled"
	pruneStatusRefused  = "refused"
)

// pruneFloorDetail renders the standard reporting block for an action result.
//
// Emitted on a PASS as well as on a refusal: candidate (3) of bugs_closed/135 —
// "pruned: 4000" as a bare success counter is the alarm presented as output, and
// the fix is to publish the denominator beside it. `extra` carries the caller's
// own raw numbers (its projected inserts, its stored rows), which belong in the
// same block so a reader has the whole measurement in one place.
func pruneFloorDetail(v pruneFloorVerdict, fromConfig bool, reason string, extra map[string]interface{}) map[string]interface{} {
	status := pruneStatusPassed
	switch {
	case !v.Allowed:
		status = pruneStatusRefused
	case v.Disabled:
		status = pruneStatusDisabled
	}

	d := map[string]interface{}{
		"completeness_floor":       v.Floor,
		"completeness_from_config": fromConfig,
		"completeness_status":      status,
		"completeness_reason":      reason,
		"completeness_cohorts":     v.Detail(),
	}
	for k, val := range extra {
		d[k] = val
	}
	return d
}

// pruneRefusal describes one refusal for the durable surface below. Kept as a
// struct because the call sites differ in what they can name (a page, a whole
// site) and a positional signature of six strings is how the wrong one gets
// passed silently.
type pruneRefusal struct {
	SiteID   uuid.UUID
	PageID   *uuid.UUID // nil for a site-scoped writer with no single page at fault
	Source   string     // the action, e.g. "populate_nav_tables"
	Pipeline string     // the pipeline the work item belongs to, e.g. "build"
	ItemType string     // the work item type, e.g. "prune_refused_incomplete"
	ItemKey  string     // dedup key — per subject, so repeated refusals collapse to one open item
	Subject  string     // what was being pruned, for the summary line
	Summary  string     // the one-line human summary
	Reason   string     // the floor's own sentence, numbers included
	Fix      string     // what a human should do about it
}

// emitPruneRefusalWorkItem is the DURABLE surface for a refusal by a
// SITE-SCOPED writer. A refusal nobody can see is the 034/076 shape — "no error
// anywhere" usually means no error surface — and a log line dies with the
// retention window.
//
// site_work_items rather than 135's doc_notes, because a site-scoped writer has
// the columns site_work_items requires (site_id, pipeline) in hand; 135 fell
// back to doc_notes only because a repo-wide indexer has no site.
//
// Routing mirrors the lock-blocked family: status 'needs_human_review' with NO
// handler_agent, because whether a corpus genuinely shrank by half is a human
// judgement (accept it and lower the floor for that step, or find the writer
// that truncated) and no automated handler can make it.
//
// recurrenceExpected IS SET, and the landmine on that flag says to state which
// of the two cases this is rather than leaving the reader to guess. This is NOT
// a detected defect handed to a handler that keeps failing — there is no handler,
// so "the fix is not working" has no referent. Each refusal is a fresh event
// about a fresh run. Without the flag, insertWorkItem's anti-churn heuristics
// apply: a refusal arriving within 3h of a terminal predecessor is DROPPED
// ENTIRELY, and the third is born `unresolved` — terminal, and off the very
// queue this row exists to reach. Dedup is NOT waived (idx_swi_dedup still
// refuses a second OPEN item for the same site+item_key), so a site refusing
// every rebuild still collapses to one open item.
//
// Best-effort by design: every failure is logged and callers discard the error
// with an explicit `_ =`. It is RETURNED only so a test can observe it — an
// unobservable guard is how the vacuous-mock landmine got written.
func emitPruneRefusalWorkItem(ctx context.Context, db *sql.DB, r pruneRefusal, logger *zap.Logger) error {
	if db == nil {
		return nil
	}

	spec := map[string]interface{}{
		"subject": r.Subject,
		"source":  r.Source,
		"reason":  r.Reason,
		"fix":     r.Fix,
	}
	specJSON, _ := json.Marshal(spec)

	summary := r.Summary
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn(r.Source+": begin tx failed for prune refusal item",
			zap.String("subject", r.Subject), zap.Error(err))
		return err
	}
	if _, err := insertWorkItem(ctx, tx, workItem{
		siteID:    r.SiteID,
		pageID:    r.PageID,
		source:    r.Source,
		pipeline:  r.Pipeline,
		itemType:  r.ItemType,
		severity:  "high",
		summary:   summary,
		spec:      string(specJSON),
		priority:  20,
		status:    "needs_human_review",
		createdBy: r.Source,
		itemKey:   r.ItemKey,
		// See the note above: no handler_agent, so a repeat is a new event, not
		// a failed fix.
		recurrenceExpected: true,
	}, logger); err != nil {
		_ = tx.Rollback()
		logger.Warn(r.Source+": could not record the prune refusal as a work item",
			zap.String("subject", r.Subject), zap.Error(err))
		return err
	}
	if err := tx.Commit(); err != nil {
		logger.Warn(r.Source+": commit failed for prune refusal item",
			zap.String("subject", r.Subject), zap.Error(err))
		return err
	}
	return nil
}

// pruneFloorFromConfig reads the floor from step config, tolerating the three
// shapes a JSON-backed config actually arrives in (float64 from JSON, int from a
// hand-built map, string from a seed that quoted the value). An unparseable value
// falls back to the default rather than to 0: a typo must never silently DISABLE
// a destructive-operation guard, which is the failure mode where the wrong
// outcome looks exactly like the right one.
func pruneFloorFromConfig(config map[string]interface{}, key string, def float64) (float64, bool) {
	raw, present := config[key]
	if !present || raw == nil {
		return def, false
	}
	switch t := raw.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return def, false
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, true
		}
		return def, false
	default:
		return def, false
	}
}
