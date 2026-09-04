// FILE: platform/orchestration/actions/save_sections_layer2_provenance_test.go
//
// bugs_open/357 / RFC_046 — THE STAMP MUST DESCRIBE THE BYTES THAT ARE STORED,
// and Layer 2 is where those two can come apart.
//
// Layer 2 exists to stop a rebuild blanking an interactive tool: when the fresh
// composition produces a non-interactive section for a slot whose stored row is
// interactive, it SPLICES the stored bytes back in. It has always been the reason
// the 22 mislabelled tool pages still serve their tools.
//
// But the splice replaces the section's HTML while leaving the rest of the
// section alone — including the RenderedTemplateSHA the fresh render just
// attached, which describes the bytes the splice DISCARDED. While the compile hop
// was dropping that digest (the severed carrier this lane fixed on 2026-08-23)
// the mistake was invisible: the field was always empty. Deliver it, and the
// resolver matches it happily against the hero component and writes the hero
// version onto a whole interactive tool — a confident, checkable, WRONG answer.
// The lane's own standard calls that worse than no stamp.
//
// So un-severing the carrier without this hygiene would have converted "no stamp"
// into "false stamp" on precisely the pathological rows. That is why these two
// tests exist, and why they are a PAIR: one asserts the splice yields no stamp,
// and the other proves the stamp it declined was genuinely on offer. "It did not
// take the decoy" means nothing if the decoy was never served.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// The stored tool: interactive markup, exactly the shape sectionHTMLIsInteractive
// and interactiveHTMLSQL both recognise.
const layer2ToolHTML = `<div class="tool-page"><canvas id="sim"></canvas>` +
	`<script>function run(){return 1}</script></div>`

// The fresh rebuild's hero band for the same slot: prose, no interactivity. This
// is what the composition produces for a tool page whose plan lists `hero` first,
// and it is what Layer 2 refuses to overwrite the tool with.
const layer2HeroHTML = `<section class="hero" data-component="hero"><div class="container">` +
	`<h1>Mortgage repayment</h1><p>Work out what a mortgage costs each month.</p></div></section>`

// decoyHeroTemplate is a real component template. Its digest is what a fresh hero
// render reports, so a resolver that is handed that digest WILL find this
// component and stamp it — which is the whole point of using it as the decoy.
const decoyHeroTemplate = `<section class="hero" data-component="hero"><div class="container">` +
	`<h1>{{.headline}}</h1><p>{{.subheadline}}</p></div></section>`

// layer2PreloadWith stages one stored interactive row for the Layer 2 preload.
// storedStamp is deliberately EMPTY in these tests: that is the live state of all
// 22 rows in bugs_open/357's population, and it is also what makes the mutation
// detectable — a carried stamp would short-circuit resolveComponentVersionID
// before it ever looked at the digest, hiding the very field under test.
// The stored content_data is NULL on purpose, and not only for tidiness: when the
// splice hands a section stored content_data, countSectionsWithContentData stops
// returning zero and the save SKIPS its content_data regression count — which
// desynchronises the shared harness's ordered expectations and fails every later
// assertion for a reason that has nothing to do with provenance. NULL is also the
// live shape of the nine original bugs_open/357 rows.
func layer2PreloadWith(slot, html, storedStamp string) *sqlmock.Rows {
	return layer2PreloadWithIdentity(slot, html, storedStamp, "")
}

// layer2PreloadWithIdentity also stages the stored row's component. ⚠ The column
// list here must match the preload query's exactly: a short row makes rows.Scan
// fail, the Layer 2 loop logs and skips, and the splice NEVER RUNS — at which
// point the assertions below pass while testing nothing. That happened while
// phase 2 added component_id to the query, and only the re-append case (which
// asserts a row COUNT, so it cannot be satisfied by the splice not running)
// noticed. Keep an assertion in this file that dies when the carry is skipped.
func layer2PreloadWithIdentity(slot, html, storedStamp, storedComponentID string) *sqlmock.Rows {
	return layer2PreloadWithFunction(slot, html, storedStamp, storedComponentID, "")
}

func layer2PreloadWithFunction(slot, html, storedStamp, storedComponentID, fn string) *sqlmock.Rows {
	// is_active true: the stored component exists and is live, which is the state
	// every one of these fixtures is about. layer2PreloadWithInactiveComponent
	// stages the other case.
	return layer2PreloadRows().AddRow(slot, html, nil, storedStamp, storedComponentID, fn, true)
}

// layer2PreloadWithInactiveComponent stages a stored row whose component_id no
// longer resolves to an active component — the case reappendedComponentID must
// refuse, because a dangling id is worse than none on the re-render path.
func layer2PreloadWithInactiveComponent(slot, html, storedComponentID, fn string) *sqlmock.Rows {
	return layer2PreloadRows().AddRow(slot, html, nil, "", storedComponentID, fn, false)
}

// TestAdoptCarriedProvenance_ClearsTheDiscardedDigest is the DIRECT pin, and it
// exists because the action-level tests below cannot do this job.
//
// Measured while writing them: removing the clearing line and re-running the
// spliced-tool test PASSED. The resolver swallows its own query errors and
// returns "no stamp" when a lookup fails, so a section carrying a stale digest
// and a section carrying none both end up writing nil — a guard in series
// standing in for the property under test. The action-level tests still earn
// their place (they pin the observable outcome and prove the decoy is reachable),
// but this is the one that fails when the decision changes.
//
// MUTATION THAT MUST BREAK IT: delete `s.RenderedTemplateSHA = ""` from
// adoptCarriedProvenance.
func TestAdoptCarriedProvenance_ClearsTheDiscardedDigest(t *testing.T) {
	const freshDigest = "d1e2f3a4b5c60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90"

	t.Run("stored row has no stamp — the section becomes honestly unknown", func(t *testing.T) {
		s := SectionData{RenderedTemplateSHA: freshDigest, ComponentVersionID: "abandoned"}
		adoptCarriedProvenance(&s, "")
		if s.RenderedTemplateSHA != "" {
			t.Errorf("RenderedTemplateSHA = %q, want empty. The digest describes the bytes the splice "+
				"DISCARDED; left in place it resolves and stamps the wrong component's version onto a "+
				"whole interactive tool (bugs_open/357).", s.RenderedTemplateSHA)
		}
		if s.ComponentVersionID != "" {
			t.Errorf("ComponentVersionID = %q, want empty — an unstamped stored row must not leave the "+
				"section claiming a stamp it did not inherit", s.ComponentVersionID)
		}
	})

	t.Run("stored row has a stamp — the section inherits it", func(t *testing.T) {
		s := SectionData{RenderedTemplateSHA: freshDigest}
		adoptCarriedProvenance(&s, "11111111-2222-3333-4444-555555555555")
		if s.RenderedTemplateSHA != "" {
			t.Errorf("RenderedTemplateSHA = %q, want empty", s.RenderedTemplateSHA)
		}
		if s.ComponentVersionID != "11111111-2222-3333-4444-555555555555" {
			t.Errorf("ComponentVersionID = %q — the stored stamp describes these exact bytes and must "+
				"travel with them", s.ComponentVersionID)
		}
	})
}

// TestSplice_UsesAdoptCarriedProvenance keeps the seam wired to the decision
// above. Without it, someone re-inlining the two lines — or deleting them —
// changes production while every test here still passes, because the unit test
// only knows about the helper.
//
// MUTATION THAT MUST BREAK IT: replace the adoptCarriedProvenance call in the
// Layer 2 splice with the raw assignments, or remove it.
func TestSplice_UsesAdoptCarriedProvenance(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)
	fd, ok := funcs["SavePageSectionsAction"]
	if !ok {
		t.Fatal("CONTROL FAILED: SavePageSectionsAction not found — the scan cannot see its target")
	}
	if !callsNamed(fd, "adoptCarriedProvenance") {
		t.Error("SavePageSectionsAction no longer calls adoptCarriedProvenance.\n" +
			"Layer 2's splice hands a section bytes it did not render; the section's provenance must " +
			"be made to describe those bytes at that moment. If this moved somewhere else, point this " +
			"test at the new home — do not delete it: the property is untestable at the action level " +
			"(the resolver's error handling masks it), which is the whole reason the helper exists.")
	}
}

// TestLayer2_SplicedToolIsNotStampedWithTheDiscardedRender is the assertion.
//
// The stored tool is spliced over the fresh hero band. The section still carries
// the hero component_id (nothing in this phase changes that — the identity fix is
// a separate round), so the INSERT does reach the resolver. What it must NOT
// carry is the fresh render's digest, because that digest describes the hero band
// that was just thrown away.
//
// The assertion is partly an ABSENCE OF EXPECTATIONS: no content_components read
// and no component_versions read are registered here, so if the resolver goes
// looking — which it only does when a digest survives — sqlmock fails the call
// and the save reports it. The stamp bind is pinned to nil in the same breath.
//
// MUTATION THAT MUST BREAK IT: delete `sections[matchedIdx].RenderedTemplateSHA = ""`
// from the splice arm. The digest survives, the resolver issues the
// content_components query nobody expected, and this test goes red.
func TestLayer2_SplicedToolIsNotStampedWithTheDiscardedRender(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID, heroID := uuid.New(), uuid.New(), uuid.New()

	expectSaveSlotReadsPreloading(mock, siteID, pageID, "tool-repayment", lockedRowSet(), 1, 0, 1,
		layer2PreloadWith("hero", layer2ToolHTML, ""))

	// position 1, slot "hero", and the stamp bind ($8) must be nil — unknown
	// provenance written as NULL, which is the honest state for bytes this save
	// did not render.
	// ⚠ The HTML bind is pinned to the TOOL, not AnyArg. That is what makes this
	// test die if the splice stops running for an unrelated reason — a short
	// preload row, a renamed column — instead of passing while asserting nothing
	// about a splice that never happened. It nearly did exactly that.
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, 1, layer2ToolHTML, "hero",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	meta := map[string]interface{}{
		"rendered_html":         layer2HeroHTML,
		"stored_slot_name":      "hero",
		"component_name":        "hero",
		"component_id":          heroID.String(),
		"rendered_template_sha": sha(decoyHeroTemplate),
	}

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-repayment", []interface{}{meta}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := saveResult(t, out)["sections_saved"]; got != 1 {
		t.Errorf("sections_saved = %v, want 1", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet or unexpected sqlmock calls — the resolver most likely went looking for a "+
			"version using the discarded render's digest: %v", err)
	}
}

// TestLayer2_TheDecoyStampIsGenuinelyOnOffer is the control, and without it the
// test above is worth nothing: a nil stamp proves the hygiene only if the same
// digest, in the same fixture, WOULD otherwise have resolved to something.
//
// Identical setup, one difference: the incoming section is itself interactive, so
// Layer 2 takes its "rebuild reproduced an interactive section — keep it" arm and
// splices nothing. Nothing was discarded, the digest still describes the stored
// bytes, and the resolver is expected to find the component, mint the version row
// and stamp it.
func TestLayer2_TheDecoyStampIsGenuinelyOnOffer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID, heroID := uuid.New(), uuid.New(), uuid.New()
	decoyVersionID := uuid.New()

	expectSaveSlotReadsPreloading(mock, siteID, pageID, "tool-repayment", lockedRowSet(), 1, 0, 1,
		layer2PreloadWith("hero", layer2ToolHTML, ""))

	// The resolver's two reads, now genuinely reached: the component's current
	// template (whose digest equals the one the section carries), then the
	// version row keyed by that exact template text.
	mock.ExpectQuery("SELECT html_template, input_schema FROM content_components").
		WithArgs(heroID).
		WillReturnRows(sqlmock.NewRows([]string{"html_template", "input_schema"}).
			AddRow(decoyHeroTemplate, []byte(`{}`)))
	mock.ExpectQuery("SELECT id FROM component_versions").
		WithArgs(heroID, decoyHeroTemplate).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(decoyVersionID.String()))

	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, 1, sqlmock.AnyArg(), "hero",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), decoyVersionID.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	meta := map[string]interface{}{
		// Interactive incoming content: no splice, so the digest keeps describing
		// the bytes actually stored.
		"rendered_html":         layer2ToolHTML,
		"stored_slot_name":      "hero",
		"component_name":        "hero",
		"component_id":          heroID.String(),
		"rendered_template_sha": sha(decoyHeroTemplate),
	}

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-repayment", []interface{}{meta}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := saveResult(t, out)["sections_saved"]; got != 1 {
		t.Errorf("sections_saved = %v, want 1", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the decoy stamp was NOT reachable in this fixture, which makes the paired test's "+
			"nil result meaningless rather than meaningful: %v", err)
	}
}

// TestLayer2_ReappendedToolCarriesTheStoredStamp pins the other carry arm. When
// the incoming composition drops the slot entirely, Layer 2 re-appends the stored
// tool — and the row it appends must carry the stored row's own provenance rather
// than acquiring the plan's or losing what it had.
//
// This arm cannot reach the database today: the INSERT resolves a version only
// when the section also has a component_id, and a re-appended section has none.
// The test therefore asserts at the SECTION level, through the exported behaviour
// it can see — the row is written with no component and no stamp — and exists so
// that the two arms cannot silently disagree once the identity round lands.
func TestLayer2_ReappendedToolCarriesTheStoredStamp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()

	expectSaveSlotReadsPreloading(mock, siteID, pageID, "tool-repayment", lockedRowSet(), 1, 0, 1,
		layer2PreloadWith("tool-calculator", layer2ToolHTML, ""))

	// The incoming section goes in first and is not what this test is about, so
	// it is left unconstrained; over-binding it only buys failures about other
	// columns. The re-appended tool at position 2 is the assertion.
	mock.ExpectExec("INSERT INTO page_components").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, 2, sqlmock.AnyArg(), "tool-calculator",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-repayment", []interface{}{
			map[string]interface{}{
				"rendered_html":    layer2HeroHTML,
				"stored_slot_name": "prose-0",
				"component_name":   "prose-0",
			},
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := saveResult(t, out)["sections_saved"]; got != 2 {
		t.Errorf("sections_saved = %v, want 2 (the incoming section plus the re-appended tool)", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}

}
