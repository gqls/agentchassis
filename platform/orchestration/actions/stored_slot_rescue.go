// FILE: platform/orchestration/actions/stored_slot_rescue.go
//
// The STORED-IDENTITY half of section-name acceptance (bugs_open/204).
//
// WHAT THIS IS FOR. Four surfaces resolve a proposed section name against the
// component catalogue and DROP what they cannot resolve. On a decomposed site a
// page's sections are POSITIONAL slot names — `prose-0`, `tool-1` — which are no
// component's name or function under any spelling, so the catalogue can never
// resolve them and every one is dropped. Measured 2026-08-21: of the 140 section
// drops recorded fleet-wide since the durable record shipped on 08-16, **140 are
// positional slot names** across 41 pages. On 2026-08-20 one replan emptied
// `pages.sections` on 41 of 45 live pages this way.
//
// THE JUDGEMENT, and why it is narrow on purpose. This asks exactly one question,
// per page: *does this page already carry a slot under this name?* If it does, the
// name is not junk — it is the page's own record of its composition, and deleting
// it destroys the only thing that says what the page is made of (`page_components`
// keeps serving, so nothing looks wrong until the next rebuild builds an empty
// page over a live one).
//
// ⚠ THIS IS NOT A WIDENING OF `componentNameResolver`, AND MUST NOT BECOME ONE.
// LANDMINES.md ("Widening a planner's component MENU changes nothing on its own")
// forbids that explicitly: three of the resolver's four call sites are the gap
// planner, whose menu PLAN-049 records as *deliberately* un-widened, so anything
// added to the resolver's valid set would hand that path an authority an
// owner-facing decision withheld. The distinction here is structural, not a
// promise in a comment:
//   - the resolver's query, signature and valid set are untouched;
//   - the judgement is keyed on (page, slot) from that page's OWN realised rows,
//     so it can only ever preserve a name the page already carries — it cannot
//     authorise placing any component anywhere it is not already placed;
//   - it runs only AFTER the resolver and its menu union have both missed, so it
//     can never rebind or shadow a name the catalogue would have resolved.
// `TestStoredSlotRescue_IsScopedToTheProposedPage` is what makes the second bullet
// a property rather than a claim.
//
// LAZY BY DESIGN. The site's rows are read at most once per run, and only on the
// FIRST resolver miss. A site whose section names are honest component functions
// never misses, so it issues no extra query at all — the arm is inert on the
// undecomposed estate by construction, not by configuration.
//
// FAILING TOWARD KEEPING. If the read fails, the verdict is `slotUnknown` and the
// caller KEEPS the entry rather than dropping it. A transient database error must
// not be able to reproduce the incident this file exists to prevent. The costs are
// not symmetric: a junk name surviving into a plan is deferred loudly one step
// later by `plan_sections` (which resolves stored identity itself and files an
// actionable item), whereas an emptied decomposed page is recoverable only from a
// snapshot somebody thought to take.

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// slotVerdict is deliberately three-valued. A boolean would collapse "this page
// does not carry that slot" (drop is correct) into "I could not find out" (drop
// is the incident), which is the whole failure mode.
type slotVerdict int

const (
	// slotNotStored — the page carries no slot under this name. The caller's
	// existing drop is correct and proceeds unchanged.
	slotNotStored slotVerdict = iota
	// slotStored — the page already carries this slot. Keep the entry verbatim.
	slotStored
	// slotUnknown — the stored rows could not be read. Keep the entry, loudly.
	slotUnknown
)

// storedSlotRescue is one run's view of a site's stored slot names. Not safe for
// concurrent use; each call site builds its own and uses it within one action.
type storedSlotRescue struct {
	db     *sql.DB
	siteID uuid.UUID
	logger *zap.Logger

	loaded bool
	failed bool
	set    map[string]map[string]bool

	// keptPages records what was rescued, for the durable summary. A count alone
	// could not tell a reader WHICH pages were decomposed.
	keptPages map[string][]string
	kept      int
}

// storedSlotRescueFor builds a rescue from the site id as it arrives in collected
// data — a string that may be absent or malformed. It degrades to nil (i.e. to
// today's behaviour exactly) rather than erroring, because a run with no site
// identity is one this arm has nothing to say about, not one that should fail.
func storedSlotRescueFor(db *sql.DB, siteIDStr string, logger *zap.Logger) *storedSlotRescue {
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		if logger != nil && siteIDStr != "" {
			logger.Warn("stored slot rescue disabled: site id is not a uuid — a decomposed page's positional slot names will be dropped as before",
				zap.String("site_id", siteIDStr))
		}
		return nil
	}
	return newStoredSlotRescue(db, siteID, logger)
}

// newStoredSlotRescue returns nil when it has nothing to work with — no database,
// or no site identity. A nil rescue answers slotNotStored for everything, which is
// exactly today's behaviour, so a caller never needs to nil-check it.
func newStoredSlotRescue(db *sql.DB, siteID uuid.UUID, logger *zap.Logger) *storedSlotRescue {
	if db == nil || siteID == uuid.Nil {
		return nil
	}
	return &storedSlotRescue{db: db, siteID: siteID, logger: logger, keptPages: map[string][]string{}}
}

// verdict answers whether pageName already carries a slot called sectionName.
//
// Call it ONLY after the component resolver has missed. Calling it earlier would
// let a stored slot shadow a name the catalogue resolves, which is a different
// (and unreviewed) behaviour — pinned by TestStoredSlotRescue_RunsOnlyAfterAMiss.
func (r *storedSlotRescue) verdict(ctx context.Context, pageName, sectionName string) slotVerdict {
	if r == nil || pageName == "" || sectionName == "" {
		return slotNotStored
	}
	if !r.loaded {
		r.load(ctx)
	}
	if r.failed {
		return slotUnknown
	}
	if r.set[pageName][sectionName] {
		r.kept++
		r.keptPages[pageName] = append(r.keptPages[pageName], sectionName)
		return slotStored
	}
	return slotNotStored
}

func (r *storedSlotRescue) load(ctx context.Context) {
	r.loaded = true
	rows, err := datahelpers.LoadPageSlotRowsForSite(ctx, r.db, r.siteID)
	if err != nil {
		r.failed = true
		if r.logger != nil {
			// Loud, because the consequence of this failure is that section names
			// survive that would otherwise have been dropped — a reader seeing
			// unexpected names in a plan needs this line to explain them.
			r.logger.Warn("stored slot identities could not be read — section names that resolve to no component are being KEPT rather than dropped for this run, because a transient read failure must not be able to empty a decomposed page",
				zap.String("site_id", r.siteID.String()),
				zap.Error(err))
		}
		return
	}
	r.set = datahelpers.SlotNameSet(rows)
	if r.logger != nil {
		r.logger.Info("loaded stored slot identities for section-name rescue",
			zap.String("site_id", r.siteID.String()),
			zap.Int("pages_with_slots", len(r.set)))
	}
}

// keptCount is the observable that distinguishes "the rescue fired" from "no name
// needed rescuing this run". Without it those two are indistinguishable, which is
// the `resolvedViaMenu` lesson from bugs_open/282's round.
func (r *storedSlotRescue) keptCount() int {
	if r == nil {
		return 0
	}
	return r.kept
}

// readFailed reports whether the run is keeping names because it could not read,
// rather than because it recognised them. A caller reporting a clean run must not
// conflate the two.
func (r *storedSlotRescue) readFailed() bool {
	return r != nil && r.failed
}

// keptFinding builds the ONE durable row per run that says what was rescued.
//
// One row, not one per keep: a decomposed site rescues 40-70 names in a single
// replan, and 70 rows would bury the drops that still matter. A keep is the
// CORRECT outcome — the record exists so a later session can prove the arm fired
// on real data (a post-fix zero in the drop table is indistinguishable from a
// blind detector without it), not because anything is wrong.
func (r *storedSlotRescue) keptFinding() []agenterrors.Finding {
	if r == nil || r.kept == 0 {
		return nil
	}
	return []agenterrors.Finding{{
		ErrorCode: "PLAN_SECTION_NAME_KEPT_BY_STORED_SLOT",
		Severity:  "warning",
		Message: fmt.Sprintf("%d proposed section name(s) across %d page(s) resolved to no active component but ARE stored slot names on the page proposed for — kept instead of dropped (bugs_open/204)",
			r.kept, len(r.keptPages)),
		Context: map[string]interface{}{
			"kept_count":  r.kept,
			"kept_pages":  r.keptPages,
			"explanation": "these are positional slot names (prose-0, tool-1) on a decomposed page; the component each one is lives on page_components.component_id, not in the component catalogue. Dropping them destroys the only record of the page's composition.",
			"remedy":      "none — this is the corrective arm working. If you are here because a plan carries a name you did not expect, check kept_pages: the name is one the page already serves.",
		},
	}}
}
