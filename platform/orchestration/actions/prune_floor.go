// FILE: platform/orchestration/actions/prune_floor.go
//
// A shared FLOOR for reconciliation prunes — the "delete everything this run did
// not re-write" pattern (bugs_open/135).
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
// CURRENT CONSUMER: index_code_symbols (code_symbols_actions.go).
// KNOWN CANDIDATES, deliberately NOT converted here (each is another lane's live
// territory and each needs its own cohort choice measured, not assumed):
// populate_nav_tables_action.go:147 (site_nav_items/groups, whole-site DELETE),
// site_db_actions.go:1474 (link_registry per source page),
// save_page_sections_action.go:532 (agent-writable page_components per page).
// Naming them here is the point of putting the rule in its own file: the next
// thread that touches one of those does not have to re-derive the argument.

package actions

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
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
func (v pruneFloorVerdict) Reason(op, subject, configKey string) string {
	switch {
	case v.Disabled:
		return fmt.Sprintf("%s: floor DISABLED by config (%s=0) for %s — pruning whatever this run did not re-confirm, unchecked",
			op, configKey, subject)
	case v.Allowed:
		return fmt.Sprintf("%s: floor cleared for %s (%s=%.2f); %s", op, subject, configKey, v.Floor, v.cohortList(v.Cohorts))
	default:
		return fmt.Sprintf("%s: REFUSED for %s — this run re-confirmed too little of what is stored (%s=%.2f): %s. "+
			"NOTHING was deleted; the rows this run did not confirm are retained and a later run that sees the whole corpus will prune them. "+
			"If the shrinkage is genuine, re-run with %s lowered (0 disables the floor entirely).",
			op, subject, configKey, v.Floor, v.cohortList(v.Failing), configKey)
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
// bugs_open/135 — "pruned: 4000" as a bare success counter is the alarm presented
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
