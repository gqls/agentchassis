// FILE: platform/livespec/unarmed_completers.go
//
// Split out of livespec.go deliberately, not for tidiness: on 2026-08-25 another
// session had an in-flight rename in that file, and a pathspec commit of a shared
// file takes the other session's HALF-WRITTEN work as a passenger (LANDMINES.md,
// "a pathspec commit still takes a SAME-FILE passenger"). A declaration with its
// own file cannot do that to anybody.

package livespec

// ── Unarmed completers of `complete` via update_work_item_status (bugs_open/375) ──
//
// WHAT THE PROBLEM IS, in plain terms. A work item is one recorded defect. A
// VERIFIER re-runs that defect's own predicate immediately before the item is
// stamped `complete`, and refuses the stamp if the defect is still there. The
// platform has THREE writers of `complete` and they do not agree about consulting
// it: CompleteWorkItemAction always does; the claimed-item-timeout sweep never does
// (held off a type by ClaimedItemTimeoutExclusions above, bugs_closed/317); and
// UpdateWorkItemStatusAction does so only when the STEP sets
// `verify_before_complete: true` (WII-030). No step sets it today.
//
// THE FAILURE THIS DECLARATION EXISTS TO STOP. verifier_coverage_test.go maintains
// a backlog of item types that OUGHT to get verifiers and calls it, in its own
// words, "the actionable backlog, not an excuse list". Somebody will work that
// backlog, register a verifier, watch the coverage guard go green — and protect
// nothing, because the steps that complete their type never ask. The runtime record
// added in WII-030 catches that, but only AFTER a live item has already completed
// unverified. This declaration is the half that catches it at authoring time.
//
// THE CONTRACT, both directions, and which half checks which:
//
//	declared  ⇔  (a live update_work_item_status step whose status is `complete`
//	              and which does NOT set verify_before_complete)
//
//   - GO, at build time (unarmed_completer_lockstep_test.go): every entry whose
//     ItemType HAS a registered verifier must carry a non-empty Acknowledged
//     reason — somebody looked at that arm and decided, on the record, not to arm
//     it. An entry naming a type with no verifier needs only its Why.
//   - THE LIVE HALF, which Go cannot do (cmd/config-key-audit
//     --unarmed-verified-completers): the declared set must EQUAL the live set.
//     Go cannot read agent_definitions, so without that arm this list goes stale by
//     ADDITION — a new agent completing through this writer would be invisible,
//     which is precisely the criticism the council levelled at the first cut of
//     verifier_coverage_test.go ("a guard against 'someone must remember' that
//     itself relies on someone remembering"). DO NOT rely on the Go half alone.
//
// ⚠ ItemType IS A DECLARATION, NOT A DERIVATION. Which item types an agent
// completes lives in site_work_items.handler_agent, not in its workflow config, so
// no Go test and no config audit can derive it. The author of an entry is asserting
// it. Re-measure with the census in
// docs024_key_docs_latest/bugfix_375_completion_verifier_gap/RUNBOOK_completion_verifier_gap.md
// — and UNION site_work_items_archive, because the live table is a rolling window
// and over it alone this very census reported 5 types where there are 7.

// UnarmedCompleter is one live workflow step that stamps `complete` through
// update_work_item_status without consulting the item type's verifier.
type UnarmedCompleter struct {
	Agent    string // agent_definitions.type
	Step     string // the step name within that agent's workflow
	ItemType string // the item_type this arm completes — DECLARED, see the ⚠ above
	Why      string // why this arm exists at all

	// Acknowledged is required, and required to be a REASON, once ItemType has a
	// registered verifier: at that moment this arm is actively skipping a guard,
	// and the estate's rule is that such a gap is a decision on the record rather
	// than an accident nobody can see. Empty is correct while no verifier exists.
	Acknowledged string
}

// UnarmedVerifiedCompleters is the declaration. Sourced from the live census
// 2026-08-25; refresh query in the RUNBOOK named above.
var UnarmedVerifiedCompleters = []UnarmedCompleter{
	{
		Agent: "image-build-handler", Step: "mark_work_item_complete", ItemType: "needs_imagery",
		Why: "the image builder closes its own item on success; the highest-volume type on this writer " +
			"(565 rows / 469 completions all-history as of 2026-08-25, live UNION archive)",
	},
	{
		Agent: "image-source-unsatisfiable-handler", Step: "close_complete", ItemType: "image_source_unsatisfiable",
		Why: "closes a flag-only finding. ⚠ its detector files HandlerAgent:\"\" at needs_human_review, so " +
			"nothing routes to this agent — 17 rows / 2 completions all-history, and arming this arm would " +
			"arm a step almost nothing reaches (bugs_open/033 CONTRIB 2026-08-25 records the same shape for image_url_404)",
	},
	{
		Agent: "image-url-404-handler", Step: "close_complete", ItemType: "image_url_404",
		Why: "same flag-only shape: the detector sets HandlerAgent:\"\" deliberately (check_image_url_404.go, " +
			"\"repairing a stale reference means removing or repointing it, which no image generator can decide\"). " +
			"3 rows ever reached this agent, by hand, and it escalated all three back to needs_human_review",
	},
	{
		Agent: "required-fields-missing-handler", Step: "close_converted", ItemType: "required_fields_missing",
		Why: "CQ-023's router: closes the original after filing a repair as a follow-on item",
	},
	{
		Agent: "required-fields-missing-handler", Step: "close_resolved", ItemType: "required_fields_missing",
		Why: "CQ-023's router: closes on positive evidence that the required fields are now populated",
	},
	{
		Agent: "required-fields-missing-handler", Step: "close_stale", ItemType: "required_fields_missing",
		Why: "CQ-023's router: closes on POSITIVE evidence of absence only, since migration 574 " +
			"(bugs_closed/367 — before it, a finding about a non-deployed component fell here and closed green)",
	},
}
