// FILE: platform/orchestration/actions/save_page_meta_description_refusal_item.go
//
// Files the durable record of a COPY-GATE refusal of a meta description.
//
// Its own file, not an addition to save_page_meta_description_action.go, for the
// reason given in platform/livespec/unarmed_completers.go: a pathspec commit of a
// shared file takes another session's half-written work as a passenger, and that
// file is one this lane and the 320/338 lanes all touch.
//
// ── WHAT THE PROBLEM WAS (bugs_open/442) ─────────────────────────────────────
//
// `save_page_meta_description` runs the site's voice gate and the banned-claims
// sweep before it writes. When one fires it REFUSES — correct behaviour — and
// returns `(map, nil)`: a NIL error. So the step succeeds, the loop continues,
// the orchestration COMPLETEs, the scheduled task stamps a clean run, and the
// page stays blank on an hourly schedule. The refusal was one `logger.Warn` plus
// a field in `collected_data` that nothing asserts on and that is unreadable
// after ~26 hours.
//
// ── WHY THE OBVIOUS FIX WAS NOT THE FIX, AND THE NUMBER THAT DECIDED IT ──────
//
// The obvious version of "make it loud" is a `needs_human_review` item with no
// handler, which is how the estate's other flag-only findings are filed. That is
// not loud. [MEASURED 2026-09-03, site_work_items UNION site_work_items_archive]:
//
//	items WITH a handler_agent   56,315   83% complete
//	items with NO handler         6,699   17% complete,  989 parked
//
// and `voice_tells` — the queue this refusal would naturally have joined — is
// 69 rows, **every one of them handler-less**: 3 complete, 66 parked, nothing
// filed since 2026-08-27. The graveyard is not a fact about busy humans. It is
// what filing without an actor looks like.
//
// So the item is filed AT AN ACTOR. `meta-description-repair` (migration 729)
// re-asks the writer for the sentence WITH the refusal reason quoted back at it,
// and saves through this same action, so the same gates judge the retry. A
// second refusal fails the item into the estate's ordinary attempt machinery
// with both attempts and both reasons on the record — which is a far better
// thing to hand a person than "this page has no description".
//
// ── WHAT IS DELIBERATELY NOT FILED, AND WHY EACH ─────────────────────────────
//
//   - The four cheap reasons (`empty_candidate`, `candidate_looks_internal`,
//     `candidate_too_long`, `already_has_description`). Nothing was published and
//     nobody needs to judge anything: either there was no usable text or the page
//     already has a description. bugs_open/442 §5 candidate 1 says the same.
//
//   - `voice_gate_unreadable`. This is the tempting one and it is WRONG to file
//     here. It means the gate could not be LOADED — an infrastructure fault, not
//     a judgement about the copy. Routing it to a rewrite handler asks the wrong
//     actor: the retry would produce a new sentence, the gate would still be
//     unreadable, and the item would churn against its own dedup key for as long
//     as the fault lasted. It stays a `logger.Warn`, it is named in the operator
//     message (migration 728), and it is recorded as a residual in bugs_open/442
//     rather than silently folded in here.
//
// ── SHIPS ARMED, NO OPT-IN FIELD, AND THAT IS A DECISION ─────────────────────
//
// The 2026-08-02 §2 ruling puts new authority on a SHARED seam behind an opt-in
// field defaulting OFF. This is not that: the authority is exercised by one
// action, on its own refusal path, and no other caller can reach it. Per the
// RFC_022 narrowing of 2026-08-11 an opt-in field nothing names would be the
// inert-by-construction shape that ruling exists to discourage, and the owner has
// ruled against default-OFF switches that rot unexercised. It also adds no key to
// this action's RFC_022 optional-key budget, which stays at 5.

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// metaDescriptionRefusalHandler is the agent that repairs a refused description.
// Seeded by migration 729. If it is ever removed, writeWorkItem's registration
// probe DEMOTES the item to `deferred` rather than filing something unroutable
// (load_work_item_actions.go) — so a missing handler degrades to a parked row,
// never to a dispatcher livelock (bugs_open/078).
const metaDescriptionRefusalHandler = "meta-description-repair"

// metaDescriptionRefusalItemType is the work item type. New type, one producer.
const metaDescriptionRefusalItemType = "meta_description_refused"

// metaDescriptionRefusalIsLoud says whether a refusal reason earns a work item.
//
// TRUE for the two COPY JUDGEMENTS only. Everything else is either "nothing was
// published and nothing is wrong" or, for voice_gate_unreadable, a fault a
// rewrite cannot fix — see the file header, which gives the reason for each.
//
// Written as an explicit allow-list rather than "not one of the cheap four", so
// an EIGHTH reason added later is silent by default rather than loud by default.
// That is the correct direction for a filing path: a new reason nobody has
// classified should not start filing work items on its own.
func metaDescriptionRefusalIsLoud(reason string) bool {
	switch reason {
	case "voice_tell", "banned_claim":
		return true
	default:
		return false
	}
}

// metaDescriptionRefusalItemKey is the dedup key. PAGE-SCOPED on purpose: keyed
// on the site alone, one page's open refusal would hold the slot against every
// other page on that site, which is the trap deadURLControlItemKey documents.
//
// Deliberately NOT keyed on the reason or the candidate text. The defect is "this
// page's description was refused and the page is blank"; a rewrite that trips a
// DIFFERENT rule is the same defect still unfixed, and keying on the reason would
// file a second row for it every hour.
func metaDescriptionRefusalItemKey(pageID uuid.UUID) string {
	return fmt.Sprintf("%s:%s", metaDescriptionRefusalItemType, pageID.String())
}

// fileMetaDescriptionRefusal records a copy-gate refusal as a work item.
//
// NEVER returns an error and never changes what the caller returns: the refusal
// itself is correct behaviour that has already been decided by the time this
// runs, and a bookkeeping failure must not turn a correct refusal into a failed
// step. Every exit logs. This mirrors emitSectionDeadControlItem's posture.
func fileMetaDescriptionRefusal(
	ctx context.Context,
	params ActionParams,
	config map[string]interface{},
	candidate, reason, detail string,
	logger *zap.Logger,
) {
	if !metaDescriptionRefusalIsLoud(reason) {
		return
	}

	siteID := resolveMetaDescriptionSiteID(params)
	if siteID == uuid.Nil {
		logger.Warn("meta description refused, but no site_id resolvable — not filed",
			zap.String("reason", reason))
		return
	}

	// Resolved HERE rather than before the gate, so this cannot change the
	// action's own behaviour: today a gate refusal returns before the page id is
	// ever resolved, and moving that resolution earlier would make a page-id
	// config fault error out on candidates that are refused anyway.
	pageID, err := resolveMetaDescriptionPageID(ctx, params, config, logger)
	if err != nil || pageID == uuid.Nil {
		logger.Warn("meta description refused, but no page id resolvable — not filed",
			zap.String("reason", reason), zap.Error(err))
		return
	}

	spec := map[string]interface{}{
		"page_id":           pageID.String(),
		"refused_candidate": candidate,
		"reason":            reason,
		"detail":            detail,
		"source":            "save_page_meta_description",
		"fix": "The site's copy gates refused this meta description, so the page " +
			"was left with none and the hourly backfill will offer the same kind " +
			"of sentence again. Rewrite it so it does not trip the rule named in " +
			"`detail`, keeping it one sentence, 110-150 characters and at most 20 " +
			"words. Do not raise the site's voice thresholds to make it pass: they " +
			"are per-site and would stop gating that site's PAGES too " +
			"(bugs_open/338).",
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		logger.Warn("meta description refused: could not marshal spec — not filed",
			zap.String("reason", reason), zap.Error(err))
		return
	}

	summary := fmt.Sprintf("Meta description refused by the %s gate — page left blank: %s",
		reason, detail)
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("meta description refused: begin tx failed — not filed", zap.Error(err))
		return
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		pageID:       &pageID,
		source:       "save_page_meta_description",
		pipeline:     "build",
		itemType:     metaDescriptionRefusalItemType,
		severity:     "medium",
		summary:      summary,
		spec:         string(specJSON),
		priority:     40,
		handlerAgent: metaDescriptionRefusalHandler,
		status:       "triaged",
		createdBy:    "save_page_meta_description",
		itemKey:      metaDescriptionRefusalItemKey(pageID),
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("meta description refused: insert work item failed — refusal stays a log line",
			zap.String("reason", reason), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("meta description refused: commit failed — refusal stays a log line",
			zap.String("reason", reason), zap.Error(err))
		return
	}

	// `inserted` false is the ORDINARY case on the second and later hours: the
	// dedup key already holds an open row for this page. Logged at Info either
	// way so the two are distinguishable in a pod log, because "filed" and
	// "already filed" look identical from the outside otherwise.
	logger.Info("meta description refused: work item recorded",
		zap.String("page_id", pageID.String()),
		zap.String("reason", reason),
		zap.String("handler", metaDescriptionRefusalHandler),
		zap.Bool("inserted", inserted))
}
