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
// ⚠ PHASE 1 IS THE GO SIDE ONLY. Today the guards compare Go against these
// declarations. NOTHING here is compared against the live database yet — that is
// the phase-2 auditor (bugs_open/363), and until it ships a declaration can drift
// from the live object silently. Declarations that cannot be exercised at all
// until then are marked PhaseLiveAudit and counted, so "declared but never read"
// is visible rather than implied.
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
	// PhaseGoSide: a phase-1 Go test compares the Go vocabulary to this Declaration.
	PhaseGoSide Phase = iota
	// PhaseLiveAudit: nothing can check this until the phase-2 live auditor exists.
	// It is INERT. It is here so the phase-2 probe is written down where the rest of
	// the contract lives, not so anyone can believe it is guarded.
	PhaseLiveAudit
)

// DeferredDeclarations is the number of Declarations that are inert until the
// phase-2 auditor lands. It is asserted by livespec_test, so ADDING an inert
// declaration forces this number up and makes the gap impossible to grow quietly.
//
// This exists because of a council objection (bug_historian, round 2): a field
// that is accepted but never read is indistinguishable from one that works, and
// this platform has been bitten by that shape before.
const DeferredDeclarations = 1

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
// longer exists. Correcting the live prose is phase-2 work; until then the live
// object misdescribes itself and this Go declaration is the accurate one.
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
		// ⚠ INERT UNTIL PHASE 2 — nothing runs ProbeSQL today. It is declared here
		// because it is the one piece of this contract that a body-text comparison
		// structurally cannot cover: the live TRIGGER SET has already outgrown the
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
}

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
