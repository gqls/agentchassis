// FILE: platform/orchestration/actions/write_audit_findings_field_capability.go
//
// ROUTING RULE 3b — a finding is not routed at a handler that cannot WRITE the
// field its own acceptance criterion names.
//
// ── WHY THIS EXISTS ─────────────────────────────────────────────────────────
// `bugs_open/395` §9 and `bugs_open/320` §5. The worked case, twice, nine days
// apart, on the same page:
//
//	An auditor finds that webdesign.co.uk's index meta description leads with a
//	catalogue count when it should lead with the offer. It files a
//	`content_rewrite` item routed at `page-build-handler`. That handler rebuilds
//	the page, deploys it, reports success truthfully, and the item closes
//	`complete` — while `pages.meta_description` is byte-identical, because
//	NOTHING IN page-build-handler CAN WRITE THAT COLUMN. The estate records a
//	green completion for work that was structurally impossible.
//
// `bugs_open/320` §9 already stated the rule in prose — *"Do not file
// content_rewrite items for them"* — on 2026-08-19. A different producer broke
// it on 2026-08-24. Neither lane was careless: **nothing enforces prose.** This
// file is that sentence with an enforcer attached, which is the only form that
// survives the next producer who never reads the bug file.
//
// ── WHY capability_gap AND NOT A REFUSAL ────────────────────────────────────
// Dropping the finding would replace a false green with silence, and silence is
// what `bugs_open/320` was filed about (56% of live pages had no description and
// nothing repaired it). So this takes the EXISTING shape Rule 3 already uses for
// `noHandlerCategories` (`bugs_open/323`, `bugs_closed/077`): status `deferred`,
// empty `handler_agent`, low severity, priority 200. The finding stays on the
// record as the demand signal for a capability the estate does not have, and it
// is not dispatchable, so no handler spends attempts on the impossible.
//
// ── WHY IT IS A POST-ROUTING GUARD, NOT A BRANCH IN THE ROUTER ──────────────
// `classifyFindingRoute` decides the handler in several places (Rules 3, 4 and
// the category table), and a new category added later would silently bypass a
// branch placed inside it. Wrapping the router means EVERY route passes through
// this check, including ones not yet written — the property that makes the bad
// state unrepresentable rather than merely absent today.
//
// ⚠ THE HONEST LIMIT, stated rather than buried: this can only fire where the
// finding names its field MECHANICALLY, i.e. where `verify_acceptance_predicates`
// (CLM-024) attached a predicate. A prose-only `acceptance_test` is undecidable
// and passes through untouched. So this is a partial guard and must not be read
// as covering the class — the complement (refusing to MINT a predicate over an
// unwritable field) belongs in the emit gate and is the `vigilant_designer_offer_analysis`
// lane's half of the agreed split.
//
// ── BLAST RADIUS, because this is a SHARED entry point ──────────────────────
// Council 021cb965 (`guidelines` seat) asked for the real number rather than the
// historical hit count, and it is right to: `write_audit_findings` serves every
// audit-source agent, so this guard changes routing for ALL of them going
// forward, not only the producer whose three rows motivated it.
//
//	[MEASURED 2026-08-25] 7 live agents reach this action:
//	brief-fidelity-auditor, content-quality-auditor, council-gate, fix-proposer,
//	offer-analyser, site-review-agent, visual-design-auditor.
//	SELECT type FROM agent_definitions WHERE is_active
//	  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
//	  AND default_config::text LIKE '%write_audit_findings%';
//
// What bounds the change is NOT the producer set but the predicate: only a
// finding carrying a CLM-024 `acceptance_predicate` over a ROSTERED field can be
// re-routed, and today only one of those seven emits predicates at all.
//
// ── THE CONTRACT ON THE EXPORTED HELPER (council 021cb965, `architecture` seat)
// The seat's objection is correct on its own terms: `HandlerCanWriteField` has an
// agreed cross-lane consumer from day one, so RFC_022's opt-in exception (which
// requires ZERO live consumers) does not cover it. Taken as the seat's stated
// minimum — a written contract note — rather than deferred:
//
//   - OWNER: this file. The roster is single-source; a second definition of the
//     same map is the drift this helper exists to prevent, and
//     TestThePageFieldWriterRosterIsDefinedExactlyOnce makes that a build failure
//     rather than a request in a comment.
//   - CALLERS OWE: nothing at call time — `known=false` means NOT MEASURED and
//     must not be read as either capability or incapacity.
//   - THIS FILE OWES: that every entry carries a re-checkable census and the date
//     it was taken, and that the staleness re-check is MECHANICAL, not prose.
//     ⚠ It is currently prose, and the `constitution` seat named that precisely:
//     *"the exact enforcement gap 320 §9 was filed about, now reproduced one
//     layer down."* That is the sharpest objection in the round and it is
//     conceded. The fix is the WII-031 treatment — a live-drift audit that reads
//     `agent_definitions` and fails when a rostered handler has GAINED a step
//     that writes the column. Named follow-on, not built here; until it exists
//     this file's guarantee rests on a human re-running the census below.
//
// ⚠⚠ AND WHEN SOMEBODY BUILDS THAT AUDIT: IT MUST RESOLVE STEPS THROUGH THE
// ACTION REGISTRY, NEVER BY GREPPING CONFIG FOR THE COLUMN NAME. An audit that
// asks *"does this handler's config mention `meta_description`?"* would run
// GREEN while being wrong about two thirds of the writers — see the measurement
// on the roster entry below. That would be this file's own blind spot reproduced
// inside the very check built to detect its staleness, which is the third time
// today this shape has appeared. `RFC_057` §3 is the fuller argument.
package actions

import (
	"fmt"
	"strings"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// pageFieldWriteRule declares, for ONE page-level field that an acceptance
// predicate is allowed to name, which handler agents can actually write it.
type pageFieldWriteRule struct {
	// WritableBy is the per-handler verdict for this field, and it must be TOTAL
	// over routableHandlers: every handler classifyFindingRoute can name carries
	// an explicit true or false, with its measurement in Why.
	//
	// ⚠ CHANGED 2026-08-31 (bugs_open/395 §9, owner-ruled). This was previously
	// a set of handlers that CAN write, so an absent handler read as "cannot" —
	// a silent default, and it is what let the `title` entry ship a false claim:
	// content-gap-planner was simply never considered, and absence answered for
	// it. A missing handler is now NOT MEASURED (HandlerCanWriteField returns
	// known=false) and the finding routes exactly as it did before this rule
	// existed, which is the safe direction; TestPageFieldWritersIsTotalOverThe-
	// RoutableHandlers makes the omission loud rather than silent.
	//
	// An EMPTY map is therefore no longer a legal value — it means "no handler
	// measured", not "no handler can write".
	WritableBy map[string]bool

	// Why is the measurement that licenses the entry: the enumeration of the
	// column's writers. A reviewer must be able to check it without re-deriving
	// it, and a reader must be able to tell a measured claim from an assumed one.
	Why string

	// Measured is the date the writer census was taken. Required, because a
	// NEGATIVE capability claim goes stale BY ADDITION — the day a handler gains
	// a step that writes this column, this entry starts refusing findings that
	// have become fixable, and nothing here can notice. Re-check with:
	//
	//	git log --since=<Measured> --diff-filter=AM -- platform/orchestration/actions
	//
	// (owner ruling 2026-08-22: a count of things carries the date it was counted).
	Measured string
}

// routableHandlers is the universe this roster must answer for: every handler
// agent classifyFindingRoute can put on a finding. It is the whole reason the
// roster can make a NEGATIVE claim at all — "no handler can write X" is only
// meaningful against a named, closed set of handlers.
//
// SOURCE OF TRUTH is classifyFindingRoute in write_audit_findings_action.go:
// its `HandlerAgent:` literals (Rules 1–6) plus every value in the designRouting
// map. TestRoutableHandlersMatchesTheRouter scans that file and fails when the
// two drift, so a new route cannot silently arrive unmeasured.
//
// ⚠ IT IS A LIST OF HANDLERS, NOT OF AGENTS. Anything not reachable as a
// finding's handler_agent is out of scope no matter what it can write —
// meta-description-backfiller writes the column freely and is absent here,
// because no finding is ever routed at it.
//
// [MEASURED 2026-08-31] the eight below are the complete set; the literals were
// byte-identical at rule 3b's ship (f4aa19ae7) and at this commit.
var routableHandlers = []string{
	// classifyFindingRoute, direct literals
	"page-build-handler",
	"copy-editor",
	"content-gap-planner",
	"spec-updater",
	// designRouting values, reached via Rule 1's `HandlerAgent: handler`
	"webdesign-agent",
	"component-template-fixer",
	"site-component-linker",
	"css-patch-agent",
}

// pageFieldWriters is the roster. A field ABSENT from this map is not this
// rule's business and routes exactly as before — the opt-in default, with the
// unsafe side (silently routing an impossible finding) requiring a deliberate
// entry to close rather than a deliberate entry to open.
//
// ⚠ The two entries below are the whole predicate vocabulary as of 2026-08-25
// (`acceptancePredicateFields` in verify_acceptance_predicates_action.go admits
// `meta_description` and `title` and nothing else). If that vocabulary widens,
// a new field arrives here UNMEASURED and correctly falls through — which is the
// safe direction, but it is also the direction that lets this guard quietly stop
// covering the class. `TestPageFieldWritersCoversThePredicateVocabulary` fails
// when the two drift apart, so the widening cannot be silent.
var pageFieldWriters = map[string]pageFieldWriteRule{
	"meta_description": {
		// TOTAL over routableHandlers. All false — measured per handler, through
		// each one's own spawn closure, not just its own step list.
		WritableBy: map[string]bool{
			"page-build-handler":       false,
			"copy-editor":              false,
			"content-gap-planner":      false,
			"spec-updater":             false,
			"webdesign-agent":          false,
			"component-template-fixer": false,
			"site-component-linker":    false,
			"css-patch-agent":          false,
		},
		// ⚠ The council's `debugging` seat objected that this claim was "asserted
		// from a private code read, not independently checkable by SQL". Fair.
		// Here is the checkable form — but NOT the obvious one, and the difference
		// is the whole lesson:
		//
		// ⚠⚠ DO NOT SEARCH FOR THE COLUMN NAME. An agent can write this column
		// without ever naming it: `upsertPage` (site_db_actions.go:1235) writes it
		// and is reached through the ACTION `sync_pages_to_db`. [MEASURED 2026-08-25,
		// by the vigilant_designer_offer_analysis lane, RFC_057] THREE live agents
		// carry a `sync_pages_to_db` step and exactly ONE names `meta_description`
		// anywhere in its config — so **2 of 3 writers are invisible to a
		// column-name search**. Which columns an action touches is a GO fact, not a
		// config fact. This is also why the roster is DECLARED rather than derived:
		// a derivation over config text cannot see two thirds of its own population.
		//
		// SEARCH FOR THE ACTION NAMES INSTEAD, whole-config so nesting cannot hide
		// one (a `workflow.steps` walk misses steps inside a loop's sub_workflow and
		// returns a confident zero — register WII-031; it missed
		// meta-description-backfiller on this very question):
		//
		//	SELECT type FROM agent_definitions
		//	 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
		//	   AND (default_config::text LIKE '%"sync_pages_to_db"%'
		//	     OR default_config::text LIKE '%"save_page_meta_description"%'
		//	     OR default_config::text LIKE '%"apply_adoption_plan"%');
		//
		// [MEASURED 2026-08-25] returns build-site-planner, meta-description-backfiller,
		// pageflow-builder, site-adoption-agent, site-work-orchestrator (+ council-gate
		// and fix-proposer, which merely quote the names in prompt/footprint text).
		// **page-build-handler and page-content-writer are NOT in it.**
		//
		// That the instrument OVER-reports is the right direction here and is the
		// reason to trust the answer: this entry makes a NEGATIVE claim, and a search
		// that yields false positives yielding NOTHING for page-build-handler is
		// strong evidence, where an under-reporting search yielding nothing would be
		// none at all.
		Why: "[MEASURED 2026-08-25] every writer of pages.meta_description is create-or-fill-blank " +
			"except one: site_db_actions.go:1235 and apply_adoption_plan_action.go:84 are both " +
			"COALESCE(NULLIF(EXCLUDED,''), pages.meta_description), so a non-empty value survives them; " +
			"the one UPDATE that CAN overwrite is save_page_meta_description_action.go:211, and it is " +
			"itself guarded TWICE — [CORRECTED 2026-08-31: this read 'the only UNCONDITIONAL UPDATE', " +
			"which was wrong] its WHERE clause carries ($3::bool OR COALESCE(meta_description,'')=''), " +
			"where $3 is the opt-in overwrite_existing config field, DEFAULT FALSE, so a caller that " +
			"does not set it gets a no-op reported as a refusal rather than a write; and it is " +
			"reachable from " +
			"exactly ONE live agent (meta-description-backfiller) whose scheduled pre_query selects " +
			"COALESCE(p.meta_description,'')='' — empty values only. page-build-handler, where content " +
			"findings are routed, has NO step that touches the column. bugs_open/395 §9, bugs_open/320 §5",
		Measured: "2026-08-25",
	},
	"title": {
		// TOTAL over routableHandlers. content-gap-planner is TRUE — see Why.
		WritableBy: map[string]bool{
			"content-gap-planner":      true,
			"page-build-handler":       false,
			"copy-editor":              false,
			"spec-updater":             false,
			"webdesign-agent":          false,
			"component-template-fixer": false,
			"site-component-linker":    false,
			"css-patch-agent":          false,
		},
		// ⚠ CORRECTED 2026-08-31. This entry previously claimed `WritableBy: {}`
		// — nobody can write it — licensed by: "pages.title has one UPDATE writer,
		// apply_gap_plan_action.go:652, which is reached from the gap-plan path and
		// from NO AUDIT-ROUTED HANDLER". Both halves of that clause name the SAME
		// AGENT, and the claim was false on the day it shipped (bugs_open/395 §9).
		// It was never wrong in production: 46 firings to date, none displacing
		// content-gap-planner. Latent, not harmless.
		Why: "[MEASURED 2026-08-31] pages.title has FIVE UPDATE writers, not one. " +
			"(1) apply_gap_plan_action.go:652, a bare unconditional `UPDATE pages SET title = $3, " +
			"sections = $4::jsonb` — and it IS reachable from an audit-routed handler: " +
			"content-gap-planner is named by classifyFindingRoute at write_audit_findings_action.go:696 " +
			"(Rule 5) and :712 (Rule 6), carries `apply_gap_plan` as a live workflow step (resolved " +
			"nesting-safe via jsonb_path_query_array($.**.action), never a workflow.steps walk), and " +
			"completes that route 989 times [live 0 / ARCHIVE 989, measured 2026-09-02]. ⚠ THE LIVE TABLE " +
			"ALONE RETURNS 0 and that is not a refutation — completing a row ARCHIVES it, so the successes " +
			"are all in site_work_items_archive by construction; a council reviewer's own check saw the 0. " +
			"Query both. Hence WritableBy true. " +
			"(2)-(5) UpsertPageForRole's Refresh list -> updatePageColumns emits a bare " +
			"`UPDATE pages SET title = $n` from create_report_page_action.go:178, " +
			"deploy_tool_action.go:464 and :636, and create_tool_component_action.go:653; that helper " +
			"was born 2026-08-02, THREE WEEKS BEFORE the original census, so this was an omission at " +
			"authoring and not staleness. [MEASURED 2026-09-02, council 76231f57 editquality objection] " +
			"THEIR OWNING AGENTS ARE report-builder, tool-deployer AND tool-generator — the complete set " +
			"carrying create_report_page / create_tool_component / deploy_tool_to_site — and that set is " +
			"DISJOINT from routableHandlers AND from every spawn closure of it (page-content-writer, " +
			"page-rerender, internal-link-resolver, research-agent, site-asset-renderer). So writers " +
			"(2)-(5) change no verdict in this roster: no finding can be routed at an agent that reaches " +
			"them. Recorded because 'they are not routable handlers' was ASSERTED in the first submission " +
			"and a reviewer was right to ask which handlers they were. " +
			"The other seven handlers are false, each measured through its own spawn closure: " +
			"page-build-handler reaches page-content-writer, page-rerender, internal-link-resolver and " +
			"research-agent, none of which writes the column (bugs_open/395 §9e); webdesign-agent " +
			"reaches site-asset-renderer and its update_site_content writes sites.content_data; " +
			"copy-editor, spec-updater, component-template-fixer, site-component-linker and " +
			"css-patch-agent carry no page-column writer at all. bugs_open/395 §9",
		Measured: "2026-08-31",
	},
}

// HandlerCanWriteField answers, for one handler agent and one page-level field,
// whether that handler can actually write it — returning the measurement that
// licenses the answer alongside it, so a caller can put evidence in front of a
// reader rather than a bare boolean.
//
// `known` is false when the field is not in the roster OR when the field is
// rostered but this HANDLER carries no verdict in it: NOT MEASURED either way,
// so no caller may treat it as either capability or incapacity.
//
// ⚠ The second arm was added 2026-08-31 and is the fix for bugs_open/395 §9.
// Before it, `rule.WritableBy[handler]` returned Go's zero value for a handler
// nobody had considered, and the caller could not tell that apart from a
// measured "no" — so an unconsidered handler read as PROVEN INCAPABLE. That is
// how the `title` entry shipped a false claim about content-gap-planner. The
// roster is now total over routableHandlers and a gap is loud; this arm is what
// makes the failure SAFE in the moment it happens rather than silent.
//
// ⚠ IT IS EXPORTED FOR ONE REASON: the emit gate (CLM-024) is INTENDED to call it
// too, so a predicate over an unwritable field would be stamped at source with the
// same verdict this routing seam acts on.
//
// ⚠⚠ BUT THAT SECOND CALLER DOES NOT EXIST YET, and this comment said it did.
// [MEASURED 2026-09-02, forced by council 76231f57's guardian objection] there are
// ZERO Go references to `field_writable` anywhere in the repo, ZERO live work items
// carry the stamp (0 of 57 rows holding a predicate), and `HandlerCanWriteField` has
// EXACTLY ONE production call site — withUnwritableFieldGuard, in this file, 30 lines
// below. RFC_057 §5 says so in as many words ("not a decision on my emit-side stamp,
// which is unbuilt"), and this comment — and the risks section of the council
// submission that shipped the totality fix — both described the seam as live anyway.
// The habit of repeating a doc's framing without checking it is the exact failure this
// whole file exists to guard against, committed in the file's own header.
// So: ONE caller today. The export is for a caller that is planned, not present. Two hand-maintained answers to "can this handler
// write this field" is precisely the drift class the council reviews for — and
// the failure would be silent, because each side would look internally correct.
// One roster, two callers. Do not copy this map.
func HandlerCanWriteField(handlerAgent, field string) (canWrite, known bool, why string) {
	rule, ok := pageFieldWriters[strings.ToLower(strings.TrimSpace(field))]
	if !ok {
		return false, false, ""
	}
	verdict, measured := rule.WritableBy[strings.TrimSpace(handlerAgent)]
	if !measured {
		// This handler was never considered for this field. NOT "cannot write".
		return false, false, ""
	}
	return verdict, true, rule.Why
}

// predicateTargetField returns the page-level field this finding's own
// acceptance criterion names, lower-cased, or "" when it names none.
//
// ⚠ IT READS THE REJECTED PREDICATE TOO, AND THAT IS LOAD-BEARING. The emit gate
// moves a predicate it refuses into `acceptance_predicate_rejected` rather than
// deleting it. If that gate ever refuses a predicate over an unwritable field,
// reading only the live key would mean the two guards BLIND EACH OTHER: theirs
// fires first, the field disappears from where this one looks, the finding routes
// at the incapable handler exactly as before, and both report success. Reading
// both makes the two orders compose instead of cancelling.
//
// ⚠ AND THE TWO KEYS HAVE DIFFERENT SHAPES, which is the trap. The live value IS
// the predicate (`{"field": "meta_description", …}`); the rejected value is an
// `AcceptancePredicateRejection` WRAPPER (`{"verdict":…, "reason":…, "predicate":
// {…}}`, verify_acceptance_predicates_action.go:277-281), so the field is one
// level down. Reading `["field"]` off the rejection finds nothing, returns "",
// and this guard silently declines to fire on exactly the population it was
// written for — a blind spot with no error and no log line.
func predicateTargetField(f auditFinding) string {
	candidates := []map[string]interface{}{f.AcceptancePredicate}
	if inner, ok := f.AcceptancePredicateRejected["predicate"].(map[string]interface{}); ok {
		candidates = append(candidates, inner)
	}
	// Defensive: also accept a flat rejection, in case the wrapper is ever
	// dropped. Costs one map lookup and removes a shape dependency.
	candidates = append(candidates, f.AcceptancePredicateRejected)

	for _, p := range candidates {
		if len(p) == 0 {
			continue
		}
		if v, ok := p["field"].(string); ok {
			if s := strings.ToLower(strings.TrimSpace(v)); s != "" {
				return s
			}
		}
	}
	return ""
}

// withUnwritableFieldGuard converts a routed finding into a capability_gap when
// the handler it was routed at cannot write the field the finding's own
// criterion names. Every other finding is returned byte-identical.
func withUnwritableFieldGuard(c classifiedFinding, f auditFinding, auditSource string) classifiedFinding {
	// Nothing was routed at a handler — Rule 3's capability_gap, the unknown
	// category fallback, or a record-only row. Not this rule's business.
	if strings.TrimSpace(c.HandlerAgent) == "" {
		return c
	}

	field := predicateTargetField(f)
	if field == "" {
		return c // prose-only criterion: undecidable here, by design
	}

	canWrite, known, why := HandlerCanWriteField(c.HandlerAgent, field)
	if !known || canWrite {
		return c
	}

	spec := c.Spec
	if spec == nil {
		spec = map[string]interface{}{}
	}
	spec["page_name"] = c.PageName
	spec["finding_severity"] = c.Severity
	spec["gap_kind"] = checks.GapHandlerRemit
	spec["unwritable_field"] = field
	spec["would_have_routed_at"] = c.HandlerAgent
	spec["builder_needed"] = fmt.Sprintf(
		"a handler that can WRITE pages.%s on a named page in response to a finding — "+
			"the capability exists in save_page_meta_description_action.go but no work-item-driven "+
			"agent reaches it, and arming a standing path that rewrites published copy is an OWNER "+
			"decision withheld on 2026-08-21 (bugs_open/320 §15)", field)
	spec["capability"] = fmt.Sprintf(
		"%s finding from %s names pages.%s, which %q cannot write. %s",
		c.ItemType, auditSource, field, c.HandlerAgent, why)
	spec["not_dispatchable"] = "status 'deferred' + empty handler_agent — deliberate. Promoting this " +
		"row dispatches work the named handler is structurally incapable of doing; it would rebuild " +
		"the page, report success, and close green with the field unchanged (bugs_open/395 §1)"

	return classifiedFinding{
		ItemType:     "capability_gap",
		HandlerAgent: "",
		Severity:     "low",
		Priority:     200,
		Status:       "deferred",
		PageID:       nil,
		PageName:     c.PageName,
		Summary: fmt.Sprintf("no handler can write pages.%s (%s): %s",
			field, auditSource, f.Description),
		Spec: spec,
		// Dedup on the FIELD, not the page or the finding: what is missing is a
		// capability, however many pages report it. Same reasoning as Rule 3's
		// dedup on the category.
		DedupKey: fmt.Sprintf("capability_gap:no_writer_for_page_field:%s", field),
	}
}
