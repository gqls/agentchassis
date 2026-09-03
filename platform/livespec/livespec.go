// FILE: platform/livespec/livespec.go
//
// livespec is the DECLARATION of what a live database object should contain —
// kept in a file that is ALLOWED TO CHANGE.
//
// WHY THIS PACKAGE EXISTS (bugs_open/363, council b3676918 APPROVED round 2).
// Several Go guards used to assert a property of a live DB object — a scheduled
// task's pre_query, a trigger function body, a CHECK constraint, an
// agent_definitions workflow — by reading the migration file that once created
// it. That cannot work, and not by accident:
//
//   - A migration is APPEND-ONLY HISTORY. run-migrations.sh records a checksum of
//     the FILE into schema_migrations, so editing an applied migration makes that
//     record a lie. The file is frozen the day it applies.
//   - The LIVE object is the accumulation of every migration since, plus lawful
//     direct edits. It keeps moving.
//
// So a guard reading a migration asserts something the checksum rule has already
// made impossible ("migration 357 no longer contains X" cannot become true), and
// it is blind to every later edit. Measured: the claimed-item-timeout exclusion
// clause was edited by migrations 322, 331 and 374 under filenames the old guard's
// glob could not match, and migration 220 itself was edited NINE times after it
// applied — for most of its life it WAS the mutable declaration. The checksum
// convention took that role away and nothing replaced it, which is why migration
// 482 ended up hiding its declaration in a PROSE COMMENT for a test to parse.
//
// This package is the replacement. It is a leaf: it imports nothing from the rest
// of the tree, so any guard may import it without a cycle.
//
// BOTH LEGS ARE NOW LIVE (phase 2 deployed 2026-08-23, CronJob
// live-declaration-drift-check at 07:00 UTC; first run 07:00:08Z read 5 objects,
// exit 0). Go guards compare Go against these declarations; the daily auditor
// (`config-key-audit --live-declaration-drift`) compares these declarations
// against the live objects. An entry checked ONLY by the auditor is marked
// PhaseLiveAudit and counted, so "declared but never read" stays visible.
//
// ⚠ A declaration is only as good as its coverage: an object with no entry here
// has no live tie at all, however well its Go guard reads.
package livespec

import (
	"fmt"
	"strconv"
	"strings"
)

// MatchMode says how a live probe result is compared to a Declaration.
type MatchMode int

const (
	// FragmentMatch: every Fragment's occurrence bounds must hold in the probe text.
	FragmentMatch MatchMode = iota
	// CountEqual: the probe returns a single integer rendered as text, which must
	// equal ExpectCount. The probe is text-typed because every probe in this
	// registry returns one text column; the comparator parses it (see CompareCount)
	// rather than relying on the driver's typing.
	CountEqual
)

// Phase records whether a Declaration can be exercised today.
type Phase int

const (
	// ⚠ READ THE NAMES CAREFULLY — they say WHO CHECKS, not WHETHER anything does.
	// Both values are checked as of 2026-08-23; neither is inert. A reader who
	// assumes "GoSide = checked, LiveAudit = pending" has it backwards, and the
	// bugs_open/375 lane lost time to exactly that reading (and a council seat lost
	// an objection to it) while the comment below still said LiveAudit was INERT.

	// PhaseGoSide: a Go unit test compares the Go vocabulary to this Declaration.
	// The daily auditor ALSO probes it — every Declaration is live-checked.
	PhaseGoSide Phase = iota
	// PhaseLiveAudit: checked ONLY by the daily auditor, never by a Go test, because
	// asserting it needs a database and `go test` has none. Not inert — merely
	// invisible to `go test`, which is why LiveAuditOnlyDeclarations counts it.
	PhaseLiveAudit
)

// LiveAuditOnlyDeclarations is the number of Declarations that NO Go test reads —
// they are checked solely by the daily auditor, because a unit test has no
// database. Asserted by livespec_test, so adding one forces this number up and the
// set cannot grow quietly.
//
// ~~DeferredDeclarations~~ renamed 2026-08-23: while phase 2 was unbuilt these
// entries were INERT, and the old name said so. Phase 2 is deployed and they are
// now checked daily — leaving a constant called "Deferred" would have been this
// lane's own defect, a written statement outliving its truth.
//
// It exists because of a council objection (bug_historian, round 2): a field that
// is accepted but never read is indistinguishable from one that works.
// ⚠ 8 -> 10 on 2026-08-26 (bugs_open/404): the re-render reason count and the
// component-template-fixer query are both live-only. The count reads a live
// column and the fixer entry reads a live query body; neither is something a
// unit test with no database can check, which is what this Phase means.
const LiveAuditOnlyDeclarations = 10

// MaxDeclarations is a growth boundary (council: architecture, round 1). livespec
// is a registry of guarded live objects, not a general config store; if it sprawls
// past this, that is a scope decision a human should make rather than a drift.
const MaxDeclarations = 24

// Fragment is one required (or forbidden) piece of a live object's text.
//
// The zero value is deliberately NOT "forbidden": Forbidden is its own field, so
// a Fragment written as {Text: "x"} means "must appear at least once", which is
// what an author intends by default.
type Fragment struct {
	Text      string
	Min       int  // minimum occurrences; 0 means no minimum
	Max       int  // maximum occurrences; <= 0 means unbounded
	Forbidden bool // must not appear at all (Min/Max ignored)
}

// Declaration is one guarded live database object.
type Declaration struct {
	Key         string // stable identifier, used by guards and by the phase-2 auditor
	Kind        string // scheduled_task | trigger_fn | trigger_bindings | constraint | workflow
	ProbeSQL    string // how the phase-2 auditor reads the LIVE object; one SELECT
	Mode        MatchMode
	Fragments   []Fragment
	ExpectCount int    // CountEqual only
	Phase       Phase  // whether anything checks this today
	Provenance  string // which migrations shaped the live object, and why it looks like this
}

// ClaimedItemTimeoutExclusions is the item_type exclusion list of the
// claimed-item-timeout sweep.
//
// THE CONTRACT: excluded ⇔ (the type has a registered verifier, gate 2) OR (the
// type has a noChangeGates entry, gate 1b). The sweep auto-completes a timed-out
// claim by writing site_work_items directly, so NEITHER completion gate runs; this
// list is the only protection. bugs_closed/317 exists because the contract was
// once stated as "the LOCKSTEP TWIN of the RegisterVerifier() calls" — gate 2
// alone — which silently excluded nothing for a gate-1b-only type.
//
// ⚠ The LIVE pre_query STILL CARRIES that superseded sentence in its own comment,
// and still names TestRegisteredVerifiersMatchClaimTimeoutExclusion, a test that no
// longer exists. This Go declaration is the accurate one.
//
// ⚠ AND THE DAILY AUDITOR CANNOT CATCH IT. Phase 2 shipped 2026-08-23 and this is
// the residual it does not close: the auditor compares the live CLAUSE to the
// declared fragments, and the clause matches — it is the PROSE ABOVE the clause
// that lies. Correcting it needs a migration on the live column, which belongs to
// whichever lane next edits it (bugs_open/341, bugs_closed/307 own it today). Do
// not read a clean live-declaration-drift run as evidence this sentence is fine.
var ClaimedItemTimeoutExclusions = []string{
	"truncated_component",
	"hardcoded_section_colors",
	"empty_section",
	"orphan_element_refs",
	"content_duplication",
	"page_canonical_collision",
	"dead_fragment_link",
	"literal_markdown",
	"unbuilt_internal_link",
	"revenue_shape_cta",
	"missing_conversion_path",
	"decision_regression",
	"needs_brand_head_assets",
	"dark_section_audit",
	// bugs_open/375 / WII-032. ⚠ APPENDED AT THE END DELIBERATELY — position is
	// load-bearing: ClaimedItemTimeoutExclusionClause() renders in slice order and the
	// live-drift Declaration is a FragmentMatch Min:1/Max:1 on that exact string. Migration
	// 634 appended this type at the end of the live clause, so anywhere but last renders a
	// string that can never match, and every build-time test stays green while the daily
	// auditor fires for ever (the lockstep is set-based; the round-trip test is order-blind).
	"required_fields_missing",
}

// WorkItemRetryNotPendingAliased is the cooldown predicate as workItemRetryNotPendingSQL("wi")
// renders it. The boundary is NON-STRICT on purpose: with "<" an item would be
// claimable a moment before it is completable, and the disagreement would only
// surface as a race.
const WorkItemRetryNotPendingAliased = "(wi.retry_after IS NULL OR wi.retry_after <= NOW())"

// ClaimedItemTimeoutExclusionClause renders the exclusion list as the SQL fragment
// the live pre_query carries. One renderer, so the Go list and the SQL spelling
// cannot drift into two vocabularies.
func ClaimedItemTimeoutExclusionClause() string {
	quoted := make([]string, 0, len(ClaimedItemTimeoutExclusions))
	for _, itemType := range ClaimedItemTimeoutExclusions {
		quoted = append(quoted, "'"+itemType+"'")
	}
	return "item_type NOT IN (" + strings.Join(quoted, ", ") + ")"
}

// Declarations is the registry. Keyed on the LIVE OBJECT, never on the migration
// that last touched it: migration 506 writes TWO live objects (a scheduled_tasks
// row and an agent_definitions row), so a registry keyed on files would reproduce
// inside the fix the very defect it exists to remove.
var Declarations = []Declaration{
	{
		Key:      "scheduled_task.claimed-item-timeout.exclusions",
		Kind:     "scheduled_task",
		ProbeSQL: "SELECT pre_query FROM scheduled_tasks WHERE name = 'claimed-item-timeout'",
		Mode:     FragmentMatch,
		Phase:    PhaseGoSide,
		Fragments: []Fragment{
			{Text: ClaimedItemTimeoutExclusionClause(), Min: 1, Max: 1},
		},
		Provenance: "220 (edited 9x while it was still the declaration), then 322/331/374 " +
			"amended the clause under names the old glob could not match; 482 froze the list; 524 " +
			"added the cooldown stamp to the same column.",
	},
	{
		Key:      "scheduled_task.build-pipeline-trigger.retry_cooldown",
		Kind:     "scheduled_task",
		ProbeSQL: "SELECT pre_query FROM scheduled_tasks WHERE name = 'build-pipeline-trigger'",
		Mode:     FragmentMatch,
		Phase:    PhaseGoSide,
		Fragments: []Fragment{
			{Text: WorkItemRetryNotPendingAliased, Min: 1},
			{Text: "retry_after < NOW()", Forbidden: true},
		},
		Provenance: "506 wrote this row and an agent_definitions query in one migration.",
	},
	{
		Key:      "trigger_fn.site_component_history_archive",
		Kind:     "trigger_fn",
		ProbeSQL: "SELECT pg_get_functiondef('site_component_history_archive'::regproc)",
		Mode:     FragmentMatch,
		Phase:    PhaseGoSide,
		Fragments: []Fragment{
			{Text: "'unstamped'", Min: 1},
			{Text: "'machine_made'", Min: 1},
			{Text: "'hand_patched'", Min: 1},
			{Text: "OLD.rendered_html_digest = md5(OLD.rendered_html)", Min: 1},
		},
		Provenance: "344. The md5 clause is the judgement classifySiteComponentArtefact mirrors.",
	},
	{
		Key:      "trigger_fn.page_component_artefact_archive",
		Kind:     "trigger_fn",
		ProbeSQL: "SELECT pg_get_functiondef('page_component_artefact_archive'::regproc)",
		Mode:     FragmentMatch,
		Phase:    PhaseGoSide,
		Fragments: []Fragment{
			{Text: "WHEN OLD.rendered_html_digest IS NULL THEN 'unstamped'", Min: 1},
			{Text: "WHEN OLD.rendered_html_digest = md5(OLD.rendered_html) THEN 'machine_made'", Min: 1},
			{Text: "ELSE 'hand_patched'", Min: 1},
			{Text: "'artefact_archive_trigger'", Min: 1},
		},
		Provenance: "357.",
	},
	{
		// CHECKED DAILY BY THE AUDITOR ONLY — no Go test reads this one, because a
		// unit test has no database. It is the one piece of this contract that a
		// body-text comparison structurally cannot cover: the live TRIGGER SET has
		// already outgrown the
		// migration that declares it (357 declares _upd and _del; 552 added
		// _content_archive_upd), so the function body can match perfectly while the
		// set of firing conditions has changed underneath it.
		Key:         "trigger_bindings.page_component_artefact_archive",
		Kind:        "trigger_bindings",
		Mode:        CountEqual,
		ExpectCount: 3,
		Phase:       PhaseLiveAudit,
		ProbeSQL: "SELECT count(*)::text FROM pg_trigger t JOIN pg_proc p ON p.oid = t.tgfoid " +
			"WHERE NOT t.tgisinternal AND p.proname = 'page_component_artefact_archive'",
		Provenance: "357 declares 2 bindings; 552 (bugs_open/355's lane) added a third. Measured 3 live 2026-08-22.",
	},
	// ── THE THREE GUARDS THAT KEEP A REPO-SIDE CHECK AND OWED A LIVE TIE ────────
	// Added 2026-08-23. livespec_test's allow-list said "Live tie is phase 2" for
	// doc_subjects_common_test.go, links_shipped_predicate_test.go and
	// v3_render_slot_name_test.go — and phase 2 shipped without covering any of
	// them, so the reasons were promising something that did not exist. Their Go
	// guards stay as they are (each checks a real REPO property); these entries are
	// the live half those reasons promised. Auditor-only: no Go test reads them.
	{
		Key:      "constraint.doc_plans_subject_type_check",
		Kind:     "constraint",
		ProbeSQL: "SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = 'doc_plans_subject_type_check'",
		Mode:     FragmentMatch,
		Phase:    PhaseLiveAudit,
		Fragments: []Fragment{
			{Text: "'tool'::text", Min: 1}, {Text: "'pipeline'::text", Min: 1},
			{Text: "'experience'::text", Min: 1}, {Text: "'action'::text", Min: 1},
			{Text: "'experience-pattern'::text", Min: 1}, {Text: "'component'::text", Min: 1},
		},
		Provenance: "doc_plans deliberately carries a NARROWER vocabulary than doc_notes — a landmine " +
			"has no shared-contract shape to put in a plan. 6 values measured live 2026-08-22.",
	},
	{
		// THE PAIRED COUNT — this is what makes the declaration above able to fail in
		// BOTH directions. FragmentMatch asserts each declared value is PRESENT, so on
		// its own it catches a value being REMOVED from the live constraint and is
		// blind to one being ADDED: every declared fragment is still there, and the
		// clean run reads identically. Measured 2026-08-25 (bugs_open/363).
		//
		// No per-fragment Max can close that — a Max bounds one value's occurrences,
		// never the SIZE OF THE SET, and a newly added value is in nobody's fragment
		// list. Only a count assertion sees it. That is the same reason
		// trigger_bindings.page_component_artefact_archive is CountEqual: 357 declared
		// two bindings and 552 added a third, which is this bug's founding example.
		Key:         "constraint.doc_plans_subject_type_check.value_count",
		Kind:        "constraint",
		Mode:        CountEqual,
		ExpectCount: 6,
		Phase:       PhaseLiveAudit,
		ProbeSQL: "SELECT ((length(pg_get_constraintdef(oid)) - length(replace(pg_get_constraintdef(oid), '::text', ''))) " +
			"/ length('::text'))::text FROM pg_constraint WHERE conname = 'doc_plans_subject_type_check'",
		Provenance: "counts '::text' casts in the constraint body, one per value (the subject_type " +
			"column itself is not cast, so there is no off-by-one). Measured live 2026-08-26: 6. The " +
			"simulated-addition control was run on the doc_notes twin (8 -> 9), proving the probe form " +
			"moves when a value is added; it was not re-run separately here.",
	},
	{
		Key:      "constraint.doc_notes_subject_type_check",
		Kind:     "constraint",
		ProbeSQL: "SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = 'doc_notes_subject_type_check'",
		Mode:     FragmentMatch,
		Phase:    PhaseLiveAudit,
		Fragments: []Fragment{
			{Text: "'tool'::text", Min: 1}, {Text: "'pipeline'::text", Min: 1},
			{Text: "'experience'::text", Min: 1}, {Text: "'action'::text", Min: 1},
			{Text: "'experience-pattern'::text", Min: 1}, {Text: "'component'::text", Min: 1},
			{Text: "'decision'::text", Min: 1},
			// 'landmine' is the load-bearing one: rebuilding this constraint from
			// doc_plans' array — the natural way to make the two agree — DROPS it and
			// orphans the live landmine corpus. Migration 273 refuses to run without
			// it for exactly this reason; this is the standing form of that refusal.
			{Text: "'landmine'::text", Min: 1},
		},
		Provenance: "8 values measured live 2026-08-22 (doc_plans' 6 plus landmine and decision).",
	},
	{
		// THE PAIRED COUNT for doc_notes. See the doc_plans twin above for why a
		// FragmentMatch declaration cannot see an ADDED value on its own.
		//
		// This is the one where addition is the dangerous direction. doc_notes is the
		// wider vocabulary and the two constraints are deliberately NOT identical, so
		// "make them agree" is a standing temptation that has already been guarded
		// against once (migration 273 refuses to rebuild this constraint without
		// 'landmine'). A value quietly appearing here is exactly how the two drift.
		Key:         "constraint.doc_notes_subject_type_check.value_count",
		Kind:        "constraint",
		Mode:        CountEqual,
		ExpectCount: 8,
		Phase:       PhaseLiveAudit,
		ProbeSQL: "SELECT ((length(pg_get_constraintdef(oid)) - length(replace(pg_get_constraintdef(oid), '::text', ''))) " +
			"/ length('::text'))::text FROM pg_constraint WHERE conname = 'doc_notes_subject_type_check'",
		Provenance: "counts '::text' casts, one per value. Measured live 2026-08-26: 8 — and 9 when one " +
			"extra value is simulated into the same body, which is the disconfirming control that makes " +
			"this probe evidence rather than a number that agrees with us.",
	},
	{
		Key:  "workflow.build-site-planner.load_existing_pages",
		Kind: "workflow",
		ProbeSQL: "SELECT default_config::text FROM agent_definitions WHERE type = 'build-site-planner' " +
			"AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL",
		Mode:  FragmentMatch,
		Phase: PhaseLiveAudit,
		Fragments: []Fragment{
			{Text: "NOT (p.deployed_at IS NULL AND COALESCE(p.build_status, '') <> 'deployed')", Min: 1, Max: 1},
		},
		Provenance: "datahelpers.PageHasShippedPredicateFor(alias p), written into this row by migration 302. " +
			"Measured live 2026-08-22: exactly one verbatim occurrence.",
	},
	{
		Key:  "workflow.page-content-writer.slot_name_from",
		Kind: "workflow",
		ProbeSQL: "SELECT default_config::text FROM agent_definitions WHERE type = 'page-content-writer' " +
			"AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL",
		Mode:  FragmentMatch,
		Phase: PhaseLiveAudit,
		Fragments: []Fragment{
			// BOTH render steps must set it; one is a half-wired workflow that looks fine.
			//
			// Max added 2026-08-25 to close this entry's half of the gain-blindness:
			// with Min alone, a THIRD render step appearing live is invisible. Exactly
			// 2 is the live truth, and a legitimate third step SHOULD stop the auditor
			// and make someone update this line — that is drift detection working, not
			// a false positive.
			{Text: "\"slot_name_from\": \"current_section.name\"", Min: 2, Max: 2},
		},
		Provenance: "seed 023 sets it on render_section and render_from_template. The SEED is not the " +
			"system — this is the live half. Measured live 2026-08-22: 2 occurrences.",
	},
	{
		// Contributed by the bugs_open/333 lane, landed here rather than in a file of
		// its own: LiveAuditOnlyDeclarations is a COUNTED invariant, so two sessions
		// appending to it independently is precisely the collision that kept this
		// package dirty — and two lanes blocked — for two days. 333 stands down.
		//
		// CountEqual on an EXACT PATH, not a fragment of default_config::text: the
		// door is a single boolean, and a substring search for it would also match
		// the key appearing under any other step, or inside a comment. Note the
		// probe deliberately avoids jsonb_path_exists' `?` operator — a bare `?` in
		// SQL handed to a Go driver can be read as a bind placeholder.
		Key:         "workflow.page-build-handler.refuse_owned_page",
		Kind:        "workflow",
		Mode:        CountEqual,
		ExpectCount: 1,
		Phase:       PhaseLiveAudit,
		ProbeSQL: "SELECT count(*)::text FROM agent_definitions WHERE type = 'page-build-handler' " +
			"AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL " +
			"AND default_config #>> '{workflow,steps,load_page_record,config,refuse_owned_page}' = 'true'",
		Provenance: "the WII-028 owned-page door reads this key; migration 488 wrote it. Its Go tie is " +
			"work_item_owned_page_door_test.go, which quotes the frozen path as a literal rather than " +
			"reading 488's file. Measured live 2026-08-25: 1 row, and 0 for a deliberately bogus path.",
	},
	{
		// bugs_open/404. The gate is the OTHER reader of the vocabulary
		// RerenderSectionReasons defines; this is what ties the two together
		// across the Go/DB boundary that neither side can cross.
		Key:  "workflow.page-rerender.check_rerender_mode.reasons",
		Kind: "workflow",
		ProbeSQL: "SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}' " +
			"FROM agent_definitions WHERE type = 'page-rerender' AND is_active " +
			"AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL",
		Mode:      FragmentMatch,
		Phase:     PhaseGoSide,
		Fragments: []Fragment{{Text: CheckRerenderModeConditionClause(), Min: 1, Max: 1}},
		Provenance: "REB-001 wrote the three-reason condition; migration 460 appended " +
			"template_changed and 473 appended literal_markdown, both on 2026-08-18, neither " +
			"touching the Go reader — which is bugs_open/404. The renderer's output was compared " +
			"against the live text on 2026-08-26 and is byte-identical.",
	},
	{
		// ⚠ THE PAIRED COUNT, AND IT IS NOT DECORATION.
		//
		// The whole-clause fragment above is NOT self-bounding. The exclusions
		// clause can be, because it ends at a closing paren; this condition has no
		// terminator at all, so a SIXTH reason APPENDED to the live gate leaves the
		// declared five-value prefix present and the Min:1/Max:1 fragment GREEN.
		// A fragment sees loss and mutation; only a count sees ADDITION — which is
		// precisely the direction bugs_open/404 drifted in, twice, in one day.
		Key:         "workflow.page-rerender.check_rerender_mode.reasons.value_count",
		Kind:        "workflow",
		Mode:        CountEqual,
		ExpectCount: len(RerenderSectionReasons),
		Phase:       PhaseLiveAudit,
		ProbeSQL: "SELECT ((length(c) - length(replace(c, 'input_data.spec.reason ==', ''))) " +
			"/ length('input_data.spec.reason =='))::text FROM (SELECT default_config #>> " +
			"'{workflow,steps,check_rerender_mode,config,condition}' AS c FROM agent_definitions " +
			"WHERE type = 'page-rerender' AND is_active AND COALESCE(is_snapshot, false) = false " +
			"AND deleted_at IS NULL) t",
		Provenance: "counts occurrences of 'input_data.spec.reason ==', one per value. ExpectCount " +
			"derives from RerenderSectionReasons, so live-versus-list parity IS the assertion rather " +
			"than a number someone has to remember to bump. Measured live 2026-08-26: 5.",
	},
	{
		// bugs_open/404 candidate 3, and the reason it is declared rather than
		// merely fixed: this is the THIRD copy of "which reasons mean re-resolve",
		// written in raw SQL inside a workflow config where nothing was watching it.
		Key:  "workflow.component-template-fixer.create_rerender",
		Kind: "workflow",
		ProbeSQL: "SELECT default_config #>> '{workflow,steps,create_rerender,config,query}' " +
			"FROM agent_definitions WHERE type = 'component-template-fixer' AND is_active " +
			"AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL",
		Mode:  FragmentMatch,
		Phase: PhaseLiveAudit,
		Fragments: []Fragment{
			{Text: "'reason','template_changed'", Min: 1, Max: 1},
			{Text: "p.rebuild_policy IS DISTINCT FROM 'owned'", Min: 1},
			{Text: "p.status = 'active'", Min: 1},
		},
		Provenance: "migration 460 rewrote this step. The LIVE row has since drifted from 460's " +
			"text — it gained the rebuild_policy clause and LOST the filename spec key (measured " +
			"2026-08-26) — which is exactly why this is declared against the live object rather " +
			"than against the file. The p.status fragment is live only after migration 655; before " +
			"that the fixer filed re-renders for archived pages, re-publishing retired ones and " +
			"making a retraction self-undoing (bugs_open/098's mechanism; 16 of 60 live tool-cta " +
			"instances sat on archived pages, measured 2026-08-25).",
	},
	{
		// bugs_open/437. The live prompt is where a schema becomes an
		// INSTRUCTION, and the instruction was wrong: built from a flat list of
		// element names, the exemplar rendered mechanism-flow's nested
		// `steps[].branches` as `"branches": "..."`, the writer produced the
		// string it was shown, and the render gate refused the page — 119 times
		// across six sites in the fortnight to 2026-09-02.
		//
		// Declared here because the fix has a Go half and a DB half that deploy
		// independently, and the DB half is the one nothing else watches: the Go
		// side's own tests pass whether or not the live row was ever migrated.
		Key:  "workflow.page-content-writer.prompt_item_shape",
		Kind: "workflow",
		// #>> the prompt path rather than default_config::text: this returns the
		// template UNESCAPED, so the fragments below are the template's own
		// spelling — byte-identical to the migration's replacement text and to
		// what the Go test renders. Probing the ::text form would force a
		// JSON-escaped restatement of a quote-heavy fragment, i.e. a third
		// spelling to keep in step.
		ProbeSQL: "SELECT default_config #>> '{workflow,steps,process_sections_loop,config," +
			"sub_workflow,steps,generate_content,config,prompt_template}' FROM agent_definitions " +
			"WHERE type = 'page-content-writer' AND is_active " +
			"AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL",
		Mode:  FragmentMatch,
		Phase: PhaseGoSide,
		Fragments: []Fragment{
			{Text: WriterPromptNestedExemplar, Min: 1, Max: 1},
			{Text: WriterPromptItemNotesTail, Min: 1, Max: 1},
			{Text: WriterPromptFlatExemplarPre437, Forbidden: true},
		},
		Provenance: "migration 724 wrote both sites; seed 023 carries the pre-437 spelling and is " +
			"history, not the system. The Go tie is writer_prompt_item_shape_437_test.go, which " +
			"renders these fragments through the real datahelpers.RenderPromptTemplate in both " +
			"deploy states. Note that path runs under text/template's DEFAULT missingkey (invalid), " +
			"not missingkey=zero: an absent value_shape is still falsy in {{if}}, which is what makes " +
			"the two halves order-free, but a bare print of either new key would emit a literal " +
			"<no value> into the prompt — hence both keys appear only inside their guards.",
	},
}

// The page-content-writer prompt's rendering of a field spec's element shape
// (bugs_open/437). These are the exact template directives migration 724 writes
// into the live row; the Go render test quotes them from here rather than
// spelling them a second time, so a change to the live contract has one place to
// be made and one place to be read.
const (
	// WriterPromptNestedExemplar is the Output Format exemplar's opening: prefer
	// the nested skeleton the planner computed, fall back to the flat item-name
	// rendering, and only then to a scalar.
	WriterPromptNestedExemplar = `"{{$f.name}}": {{if $f.value_shape}}{{$f.value_shape}}{{else if $f.item_fields}}`
	// WriterPromptItemNotesTail appends the per-property shape sentences to the
	// "What To Write" field list.
	WriterPromptItemNotesTail = `{{if .item_notes}}{{range $n := .item_notes}} {{$n}}{{end}}{{end}}`
	// WriterPromptFlatExemplarPre437 is the spelling this fix removes — the one
	// that rendered `"branches": "..."` and told the writer a nested array was a
	// string. Declared Forbidden so a revert is caught by the daily auditor
	// rather than by six sites' builds failing again.
	WriterPromptFlatExemplarPre437 = `"{{$f.name}}": {{if $f.item_fields}}`
)

// Get returns the Declaration with the given key.
func Get(key string) (Declaration, bool) {
	for _, d := range Declarations {
		if d.Key == key {
			return d, true
		}
	}
	return Declaration{}, false
}

// MustGet is Get for callers that have a compile-time-known key; a miss is a
// programming error, not a runtime condition.
func MustGet(key string) Declaration {
	d, ok := Get(key)
	if !ok {
		panic("livespec: no declaration named " + key)
	}
	return d
}

// HasFragment reports whether the Declaration requires the given text.
//
// It answers "does the DECLARATION demand this", not "does the live object contain
// it" — that second question needs the phase-2 auditor and a database.
func (d Declaration) HasFragment(text string) bool {
	for _, f := range d.Fragments {
		if f.Text == text && !f.Forbidden {
			return true
		}
	}
	return false
}

// Forbids reports whether the Declaration forbids the given text outright.
func (d Declaration) Forbids(text string) bool {
	for _, f := range d.Fragments {
		if f.Text == text && f.Forbidden {
			return true
		}
	}
	return false
}

// CompareFragments checks live probe text against a FragmentMatch Declaration and
// returns one problem string per violated Fragment. Phase-2 entry point; exported
// and unit-testable now so the comparator is exercised before it is ever wired to
// a database.
func (d Declaration) CompareFragments(live string) []string {
	var problems []string
	for _, f := range d.Fragments {
		n := strings.Count(live, f.Text)
		switch {
		case f.Forbidden && n > 0:
			problems = append(problems, fmt.Sprintf("%s: forbidden fragment %q appears %d time(s)", d.Key, f.Text, n))
		case f.Forbidden:
			// absent, as required
		case n < f.Min:
			problems = append(problems, fmt.Sprintf("%s: fragment %q appears %d time(s), want at least %d", d.Key, f.Text, n, f.Min))
		case f.Max > 0 && n > f.Max:
			problems = append(problems, fmt.Sprintf("%s: fragment %q appears %d time(s), want at most %d", d.Key, f.Text, n, f.Max))
		}
	}
	return problems
}

// CompareCount checks a CountEqual Declaration against the probe's text result.
//
// The probe renders its count as TEXT so every probe in the registry returns one
// text column; parsing here rather than trusting the driver's typing is what keeps
// that uniform (council: editquality, round 2 — a text probe matched against an int
// field is a comparison that can silently never fire).
func (d Declaration) CompareCount(live string) []string {
	got, err := strconv.Atoi(strings.TrimSpace(live))
	if err != nil {
		return []string{fmt.Sprintf("%s: probe returned %q, which is not an integer", d.Key, live)}
	}
	if got != d.ExpectCount {
		return []string{fmt.Sprintf("%s: live count is %d, declared %d", d.Key, got, d.ExpectCount)}
	}
	return nil
}
