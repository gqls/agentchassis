// FILE: platform/orchestration/actions/adopt_fragment_section_test.go
//
// RFC_046 phase 2 / bugs_open/357 — a fragment is typed TRUTHFULLY or not at all.
//
// The bug: a tool page arrives as one fragment with no <section>, is stored as a
// single section, and has its identity invented from POSITION in the page plan —
// so 22 live rows (12 of them born on 2026-08-23 alone) declare themselves the
// shared `hero` while storing a whole interactive tool.
//
// The tests are built as PAIRS, because the interesting assertion is a refusal and
// a refusal is only meaningful if the thing refused was genuinely available. Every
// "it did not bind hero" case has a sibling proving hero WAS on offer in exactly
// that fixture — with the flag off, the same section binds hero, which is both the
// decoy control and the proof that the default is byte-identical to today.
package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The tool fragment as it actually arrives: no <section>, no data-component,
// nothing about the bytes saying what they are.
const adoptToolFragment = `<div class="tool-page"><h1>Time-To-Kill Calculator</h1>` +
	`<input id="dps"><canvas id="out"></canvas><script>function ttk(){return 1}</script></div>`

// fallbackSection is what saveSectionsExtractFromHTML produces for that fragment
// AFTER enrichSectionsWithPlannedNames has given it the plan's first slot name.
// The slot name is `hero` and it STAYS `hero` — that is the design, not an
// oversight: renaming it is what arms the carry-forward landmine.
func fallbackSection() []SectionData {
	return []SectionData{{
		ComponentName:   "hero",
		HTML:            adoptToolFragment,
		Position:        1,
		FallbackAdopted: true,
	}}
}

func expectAdoptedFragmentLookup(mock sqlmock.Sqlmock, id string) {
	mock.ExpectQuery("SELECT id::text, html_template FROM content_components").
		WithArgs(adoptedFragmentFunction).
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template"}).AddRow(id, "{{.body}}"))
}

// TestAdopt_ArmedFragmentIsTypedTruthfullyAndNeverByItsSlotName is the assertion.
//
// The section's slot name is `hero`, and hero is a real, resolvable component —
// the sibling test below proves the binding is on offer. Armed, the fragment is
// bound to the component that provably produces its bytes instead, gains
// `content_data.body`, and earns a provenance stamp. The hero lookup is NOT
// expected here, so if the resolution path ran at all sqlmock would fail it.
//
// MUTATION THAT MUST BREAK IT: delete the `adoptFragments && FallbackAdopted`
// branch from enrichSectionsWithComponentIDs. The section falls through to the
// name resolution, binds hero, and both assertions fail.
func TestAdopt_ArmedFragmentIsTypedTruthfullyAndNeverByItsSlotName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	adoptedID := uuid.NewString()
	expectAdoptedFragmentLookup(mock, adoptedID)

	sections := fallbackSection()
	enrichSectionsWithComponentIDs(context.Background(), db, sections, zap.NewNop(), true)

	s := sections[0]
	if s.ComponentID != adoptedID {
		t.Errorf("ComponentID = %q, want the adopted-fragment component %q — a fragment must be typed by "+
			"what provably produces its bytes, never by the slot name its page plan supplied", s.ComponentID, adoptedID)
	}
	if got, _ := s.ContentData["body"].(string); got != adoptToolFragment {
		t.Errorf("content_data.body was not set to the fragment (got %d bytes, want %d) — without it the row "+
			"is not regenerable, which is the whole point of typing it", len(got), len(adoptToolFragment))
	}
	if s.RenderedTemplateSHA == "" {
		t.Error("no provenance stamp: adoption renders the template, so the digest is EARNED here and the " +
			"existing resolver would stamp the row at the INSERT")
	}
	if s.ComponentName != "hero" {
		t.Errorf("ComponentName = %q, want \"hero\" UNCHANGED. Renaming the slot is what makes the next "+
			"rebuild's carry-forward miss and re-append the tool beside a fresh section (LANDMINES) — this "+
			"change must never touch it", s.ComponentName)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet or unexpected queries — the name-resolution path most likely ran: %v", err)
	}
}

// TestAdopt_DisarmedIsExactlyTodaysBehaviour is the decoy control AND the
// default-OFF proof, which is why it is one test and not two.
//
// Same section, flag off. The name resolution runs and binds `hero` — today's
// behaviour, and the defect. Its value here is that it proves hero was reachable
// in the fixture above: "it did not bind hero" means nothing if hero was never
// on offer.
//
// MUTATION THAT MUST BREAK IT: change the flag's default, or gate the adoption on
// something other than the flag. This test then binds the adopted component and
// the assertion fails — which is the estate's 2026-08-02 rule (unsafe default OFF)
// expressed as a test rather than as a comment.
func TestAdopt_DisarmedIsExactlyTodaysBehaviour(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	heroID := uuid.NewString()
	mock.ExpectQuery("SELECT id::text FROM content_components").
		WithArgs("hero").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(heroID))

	sections := fallbackSection()
	enrichSectionsWithComponentIDs(context.Background(), db, sections, zap.NewNop(), false)

	if sections[0].ComponentID != heroID {
		t.Errorf("ComponentID = %q, want hero %q. With the flag OFF this must be byte-identical to today — "+
			"and if hero were NOT resolvable in this fixture, the paired test's refusal would be vacuous",
			sections[0].ComponentID, heroID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the hero binding was not actually on offer, which makes the paired test meaningless: %v", err)
	}
}

// TestAdopt_RefusesWhenTheTemplateWouldNotReproduceTheBytes pins the proof step.
// Adoption is a claim that a component produces these bytes; the only thing that
// makes it a fact rather than an assertion is rendering the template and checking.
// A wrapping template must therefore be refused, leaving the row unidentified —
// honestly unknown beats confidently wrong.
//
// MUTATION THAT MUST BREAK IT: drop the `rendered != s.HTML` check.
func TestAdopt_RefusesWhenTheTemplateWouldNotReproduceTheBytes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id::text, html_template FROM content_components").
		WithArgs(adoptedFragmentFunction).
		WillReturnRows(sqlmock.NewRows([]string{"id", "html_template"}).
			// Someone "improved" the identity template by wrapping it.
			AddRow(uuid.NewString(), `<section class="adopted">{{.body}}</section>`))

	s := SectionData{ComponentName: "hero", HTML: adoptToolFragment, FallbackAdopted: true}
	if adoptFragmentSection(context.Background(), db, &s, zap.NewNop()) {
		t.Fatal("adopted despite the template not reproducing the bytes — the row would then claim a " +
			"component that regenerates something else, which is worse than claiming nothing")
	}
	if s.ComponentID != "" || s.ContentData != nil || s.RenderedTemplateSHA != "" {
		t.Errorf("a refused adoption must change NOTHING: id=%q content_data=%v sha=%q",
			s.ComponentID, s.ContentData, s.RenderedTemplateSHA)
	}
	_ = mock.ExpectationsWereMet()
}

// TestAdopt_UnseededEstateLeavesTheRowUnidentifiedRatherThanGuessing: the seed
// migration is HELD, so "armed but not seeded" is a real state and must degrade
// safely — to the honest unknown RFC_046 asks for, never back to the guess.
//
// MUTATION THAT MUST BREAK IT: make the lookup failure fall through to the name
// resolution instead of `continue`.
func TestAdopt_UnseededEstateLeavesTheRowUnidentifiedRatherThanGuessing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id::text, html_template FROM content_components").
		WithArgs(adoptedFragmentFunction).
		WillReturnError(sql.ErrNoRows)
	// Deliberately NOT expected: the hero lookup. Its absence is the assertion —
	// a failed adoption must not become a successful guess.

	sections := fallbackSection()
	enrichSectionsWithComponentIDs(context.Background(), db, sections, zap.NewNop(), true)

	if sections[0].ComponentID != "" {
		t.Errorf("ComponentID = %q, want empty — with no component to adopt onto, the row must stay "+
			"unidentified rather than falling back to its positional name", sections[0].ComponentID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected query — the name-resolution fallback ran after a failed adoption: %v", err)
	}
}

// TestAdopt_ASectionThatDeclaresItselfIsLeftAlone. A fragment carrying
// data-component is not identity-unknown: the attribute is evidence about the
// bytes, and RFC_046 §4 keeps that inference precisely because it is evidence
// rather than position. Adoption must not swallow it.
//
// MUTATION THAT MUST BREAK IT: drop the data-component check from the adoption
// branch.
func TestAdopt_ASectionThatDeclaresItselfIsLeftAlone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	declaredID := uuid.NewString()
	mock.ExpectQuery("SELECT id::text FROM content_components").
		WithArgs("faq-block").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(declaredID))

	sections := []SectionData{{
		ComponentName:   "hero",
		HTML:            `<div class="tool-page" data-component="faq-block"><p>declared</p></div>`,
		Position:        1,
		FallbackAdopted: true,
	}}
	enrichSectionsWithComponentIDs(context.Background(), db, sections, zap.NewNop(), true)

	if sections[0].ComponentID != declaredID {
		t.Errorf("ComponentID = %q, want the self-declared component %q — the attribute is evidence about "+
			"these bytes and outranks both the plan's name and adoption", sections[0].ComponentID, declaredID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations — the self-declaration was not honoured: %v", err)
	}
}

// TestCarriedIdentity_IsOffByDefault pins the second half of the flag. The Layer 2
// carry arms must state the identity of the bytes they hand over only when armed;
// disarmed they must state nothing, which is today's behaviour exactly.
//
// MUTATION THAT MUST BREAK IT: return storedComponentID unconditionally.
func TestCarriedIdentity_IsOffByDefaultAndNarrowedToAdoptedFragments(t *testing.T) {
	if got := carriedIdentity(false, "some-component", adoptedFragmentFunction); got != "" {
		t.Errorf("carriedIdentity(disarmed) = %q, want empty — the default must change nothing", got)
	}
	if got := carriedIdentity(true, "some-component", adoptedFragmentFunction); got != "some-component" {
		t.Errorf("carriedIdentity(armed, adopted) = %q, want the stored component: an adopted row's identity "+
			"must survive a rebuild or the next one re-mints the plan's identity over it", got)
	}
	// The narrowing the council asked for, as an assertion rather than a promise:
	// a legitimately-typed component is NOT re-typed by the carry.
	if got := carriedIdentity(true, "some-component", "hero"); got != "" {
		t.Errorf("carriedIdentity(armed, hero) = %q, want empty. Carrying identity for every interactive "+
			"section is broader than the diagnosed bug and would silently keep a legitimately-typed "+
			"component at its OLD identity when a plan intended to swap it — three council seats made "+
			"that point independently", got)
	}
	if got := carriedIdentity(true, "", adoptedFragmentFunction); got != "" {
		t.Errorf("carriedIdentity with no stored component = %q, want empty", got)
	}
}

// TestAdopt_UnidentifiedRowStillWritesItsBytes answers the council's GATING
// objection (bug_historian, high) rather than arguing with it.
//
// The objection: this change creates a NEW population — rows with component_id
// NULL where today they would carry a wrong-but-present component — and the
// estate has a documented case of exactly that shape, bugs_closed/039
// "section_naming_a_missing_component_renders_an_empty_stub", where a section
// pointing at no component silently rendered empty. The first submission asserted
// "the page serves identically either way" and never verified it.
//
// This pins the half that lives at the seam being changed: a section with no
// component still reaches the INSERT with its bytes intact and a NULL component
// bind — it is not dropped, not blanked, not skipped.
//
// ⚠ WHAT THIS DOES NOT COVER, stated rather than implied: the RERENDER path.
// resolveComponent (rerender_page_sections_action.go:377) falls through to the
// slot-NAME map when component_id is empty, so an unadopted row named `hero`
// resolves to the hero component there and the fresh-render entry re-emits
// hero's id — which re-binds it on the next save. Adoption normally succeeds (the
// seeded template is the identity function, so the round trip only fails if
// someone edits it), but that residual is real, it is named in the bug file, and
// it is why this is a reduction of the mint rather than its elimination.
func TestAdopt_UnidentifiedRowStillWritesItsBytes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	expectSaveSlotReads(mock, siteID, pageID, "tool-ttk-calculator", lockedRowSet(), 1, 0, 1)

	// The row the objection is about: bytes present, component NULL (bind 5), and
	// no provenance (bind 8). Pinning the HTML bind is what makes this a real
	// assertion — a blanked or dropped section cannot satisfy it.
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, 1, adoptToolFragment, "hero",
			nil, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-ttk-calculator", []interface{}{
			map[string]interface{}{
				"rendered_html":    adoptToolFragment,
				"stored_slot_name": "hero",
				"component_name":   "hero",
				// component_id deliberately ABSENT — the unidentified row.
			},
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := saveResult(t, out)["sections_saved"]; got != 1 {
		t.Errorf("sections_saved = %v, want 1 — a section with no component must still be SAVED. If this "+
			"is 0 the row was dropped, which is bugs_closed/039's shape and the objection's exact worry", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the unidentified row did not reach the INSERT with its bytes intact: %v", err)
	}
}
