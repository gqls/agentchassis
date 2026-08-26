// Pins the bugs_open/399 write-time CTA label/destination audit.
//
// EVERY TEST NAMES THE MUTATION IT KILLS. The estate's standing lesson is that a
// guard whose deletion leaves the suite green is not a guard, and the sibling
// that motivated this file (setCTAField's SeedCTAMinted call) was proven by
// mutation to be deletable with every test still passing.
package actions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// heroSchema is a real-shaped hero input_schema: a url field whose paired label
// is the bare-stem form (cta_url → cta_text), which is one of THREE pairings
// live on the fleet and the reason this pass reads the schema instead of
// manipulating the key name.
// Shape and field specs copied from a live content_components.input_schema row
// for function='hero' (read 2026-08-26), not composed: a fixture I invent
// exercises the rule I imagined rather than the one production uses, and the
// v2 "fields" wrapper is exactly what a hand-written flat map would miss.
func heroSchema() map[string]interface{} {
	f := func(t, src string) map[string]interface{} {
		return map[string]interface{}{
			"type": t, "source": src, "required": false, "on_missing": "skip_field",
		}
	}
	return map[string]interface{}{
		"fields": map[string]interface{}{
			"headline":          f("text", "llm"),
			"cta_text":          f("text", "llm"),
			"cta_url":           f("url", "renderer"),
			"secondary_cta":     f("text", "llm"),
			"secondary_cta_url": f("url", "renderer"),
		},
	}
}

func auditCandidates(t *testing.T) []datahelpers.LabelMatchCandidate {
	t.Helper()
	rows := []struct{ id, name, title, url, nav string }{
		{"1", "news-index", "Darts News | Darts Online", "/news/index.html", "News"},
		{"2", "brands-index", "All Brands | Darts Online", "/brands/index.html", "Brands"},
		{"3", "shaft-length", "Shaft Length Guide", "/guides/shaft-length.html", "Shaft Length"},
	}
	out := make([]datahelpers.LabelMatchCandidate, 0, len(rows))
	for _, r := range rows {
		c, ok := datahelpers.NewLabelMatchCandidate(r.id, r.name, r.title, r.url, false, r.nav)
		if !ok {
			t.Fatalf("fixture candidate %q has no distinctive tokens", r.name)
		}
		out = append(out, c)
	}
	return out
}

func sectionWith(content map[string]interface{}) []SectionData {
	return []SectionData{{ComponentName: "hero", ComponentID: "c1", ContentData: content}}
}

func schemaMap() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{"c1": heroSchema()}
}

// TestAuditSectionCTALabelsCatchesTheFiledDefect is the DISCONFIRMING CASE the
// bug demands. It is the live row from dartsonline.com news-index/hero, still
// present and re-minted on 2026-08-25 20:58:09Z, seventeen minutes after the
// bug was filed.
//
// MUTATION: make JudgeCTALabel return CTALabelAgrees unconditionally, or delete
// the DeriveCTAURLFields loop body. This fails.
func TestAuditSectionCTALabelsCatchesTheFiledDefect(t *testing.T) {
	sections := sectionWith(map[string]interface{}{
		"cta_text":         "Catch up on this week's darts news",
		"cta_url":          "/brands/index.html",
		"cta_target_title": "All Brands | Darts Online",
	})

	got := auditSectionCTALabels(sections, schemaMap(), auditCandidates(t), "shaft-length", "")
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1 — the audit is blind to the defect it was written for", len(got))
	}
	f := got[0]
	if f.Verdict != "contradicts" {
		t.Errorf("verdict = %q, want contradicts", f.Verdict)
	}
	if f.NamedURL != "/news/index.html" {
		t.Errorf("named %q, want /news/index.html — the record must say what the copy DOES name, "+
			"or a human cannot act on it", f.NamedURL)
	}
	// Both sides must survive into the record: bugs_open/391 needs the
	// destination, cta_target_content_pass needs the label.
	if f.Label == "" || f.Destination == "" || f.TargetTitle == "" {
		t.Errorf("record dropped a side: label=%q destination=%q target_title=%q",
			f.Label, f.Destination, f.TargetTitle)
	}
}

// TestAuditSectionCTALabelsIsSilentOnAgreement is the false-positive guard.
// On the live fleet most pairs agree, so a pass that fired on them would file
// hundreds of records a day and be switched off within a week.
//
// MUTATION: drop the `if j.Verdict == CTALabelAgrees { continue }` arm. This fails.
func TestAuditSectionCTALabelsIsSilentOnAgreement(t *testing.T) {
	sections := sectionWith(map[string]interface{}{
		"cta_text":         "Catch up on this week's darts news",
		"cta_url":          "/news/index.html",
		"cta_target_title": "Darts News | Darts Online",
	})
	if got := auditSectionCTALabels(sections, schemaMap(), auditCandidates(t), "shaft-length", ""); len(got) != 0 {
		t.Fatalf("findings = %d, want 0 — a correct button was recorded: %+v", len(got), got)
	}
}

// TestAuditSectionCTALabelsSkipsUnresolvedDestinations pins the SCOPE test.
// A CTA-bearing component the resolver never covered carries no *_target_title;
// judging it would report a contradiction against a destination no resolver
// chose, and would silently widen this pass beyond the population the bug is
// about.
//
// MUTATION: delete `|| title == ""` from the skip condition. This fails.
func TestAuditSectionCTALabelsSkipsUnresolvedDestinations(t *testing.T) {
	sections := sectionWith(map[string]interface{}{
		"cta_text": "Catch up on this week's darts news",
		"cta_url":  "/brands/index.html",
		// no cta_target_title: setCTAField never ran for this component
	})
	if got := auditSectionCTALabels(sections, schemaMap(), auditCandidates(t), "shaft-length", ""); len(got) != 0 {
		t.Fatalf("findings = %d, want 0 — the pass judged a destination the resolver never wrote", len(got))
	}
}

// TestAuditSectionCTALabelsLeavesNonPageDestinationsToTheirOwnCheck pins the
// caller-side scope filter. A tel:/mailto: button is bugs_open/299's class and
// already has a live check (check_cta_nonpage) asking this same question through
// this same predicate. Judging it here would double-file it.
//
// MUTATION: delete the ClassifyLinkScope guard. This fails.
func TestAuditSectionCTALabelsLeavesNonPageDestinationsToTheirOwnCheck(t *testing.T) {
	sections := sectionWith(map[string]interface{}{
		"cta_text":         "Catch up on this week's darts news",
		"cta_url":          "tel:+447934524911",
		"cta_target_title": "a phone call to +44 (0) 7934 524 911",
	})
	if got := auditSectionCTALabels(sections, schemaMap(), auditCandidates(t), "shaft-length", ""); len(got) != 0 {
		t.Fatalf("findings = %d, want 0 — a tel: button was double-filed against check_cta_nonpage", len(got))
	}
}

// TestAuditSectionCTALabelsRecordsAmbiguityWithoutConvicting pins RFC_047's
// owner ruling at this seam: a button whose copy names two pages equally well is
// RECORDED as undecidable and never counted as a contradiction. Merging the two
// would misreport the size of the actionable half.
//
// MUTATION: count ambiguous findings as contradictions in countCTALabelFindings.
// This fails.
func TestAuditSectionCTALabelsRecordsAmbiguityWithoutConvicting(t *testing.T) {
	a, _ := datahelpers.NewLabelMatchCandidate("1", "alpha-report", "Annual Report", "/alpha.html", false, "")
	b, _ := datahelpers.NewLabelMatchCandidate("2", "beta-report", "Annual Report", "/beta.html", false, "")

	sections := sectionWith(map[string]interface{}{
		"cta_text":         "Read the Annual Report",
		"cta_url":          "/somewhere-else.html",
		"cta_target_title": "Somewhere Else",
	})
	got := auditSectionCTALabels(sections, schemaMap(), []datahelpers.LabelMatchCandidate{a, b}, "index", "")
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1 — an undecidable button must still be recorded, "+
			"or RFC_047 §10's stated gap stays open", len(got))
	}
	if !got[0].Ambiguous {
		t.Error("finding not marked ambiguous")
	}
	if got[0].NamedURL != "" {
		t.Errorf("an ambiguous finding named a page (%q) — that is the guess the ruling forbids", got[0].NamedURL)
	}
	contradicts, ambiguous := countCTALabelFindings(got)
	if contradicts != 0 || ambiguous != 1 {
		t.Errorf("counts = (%d contradicts, %d ambiguous), want (0, 1)", contradicts, ambiguous)
	}
}

// TestAuditSectionCTALabelsNeedsTheSchema pins that the label field is read from
// the component's declared schema and never guessed from the key name. Measured
// 2026-08-26, the three live pairings are cta_target_title→cta_text,
// secondary_cta_target_title→secondary_cta and cta_primary_target_title→
// cta_primary_label — so any string rule is wrong on two of them.
//
// MUTATION: fall back to a derived label name when the schema is missing.
// This fails.
func TestAuditSectionCTALabelsNeedsTheSchema(t *testing.T) {
	sections := sectionWith(map[string]interface{}{
		"cta_text":         "Catch up on this week's darts news",
		"cta_url":          "/brands/index.html",
		"cta_target_title": "All Brands | Darts Online",
	})
	// No schema for c1.
	if got := auditSectionCTALabels(sections, map[string]map[string]interface{}{}, auditCandidates(t), "shaft-length", ""); len(got) != 0 {
		t.Fatalf("findings = %d, want 0 — the pass invented a label field name", len(got))
	}
}

// TestAuditSectionCTALabelsHandlesTheBareStemPairing proves the schema-derived
// pairing actually resolves the awkward live case: secondary_cta_url pairs with
// the BARE stem secondary_cta, not secondary_cta_text.
//
// MUTATION: restrict DeriveCTAURLFields' sibling suffixes to _text/_label.
// This fails.
func TestAuditSectionCTALabelsHandlesTheBareStemPairing(t *testing.T) {
	sections := sectionWith(map[string]interface{}{
		"secondary_cta":              "Browse all brands we cover",
		"secondary_cta_url":          "/news/index.html",
		"secondary_cta_target_title": "Darts News | Darts Online",
	})
	got := auditSectionCTALabels(sections, schemaMap(), auditCandidates(t), "shaft-length", "")
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1 — the bare-stem label pairing was not resolved", len(got))
	}
	if got[0].LabelField != "secondary_cta" {
		t.Errorf("label field = %q, want secondary_cta", got[0].LabelField)
	}
	if got[0].NamedURL != "/brands/index.html" {
		t.Errorf("named %q, want /brands/index.html", got[0].NamedURL)
	}
}

// TestCTALabelFindingSerialisesBothSides pins that the durable record can
// actually carry the contradiction. A record that drops a side is a log line.
func TestCTALabelFindingSerialisesBothSides(t *testing.T) {
	raw, err := json.Marshal([]ctaLabelFinding{{
		Slot: "hero", Component: "hero", URLField: "cta_url", LabelField: "cta_text",
		Label: "Catch up on this week's darts news", Destination: "/brands/index.html",
		TargetTitle: "All Brands | Darts Online", NamedURL: "/news/index.html",
		NamedTitle: "Darts News | Darts Online", Verdict: "contradicts",
	}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{"cta_text", "/brands/index.html", "All Brands", "/news/index.html", "contradicts"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("record does not carry %q: %s", want, raw)
		}
	}
}

// TestCTALabelAuditIsWiredIntoTheSavePath is the WIRING SCAN. The pure seam
// above can be perfect while nothing calls it — which is how a guard ships
// inert. Source-scanned rather than asserted through a mock because the call
// sits inside SavePageSectionsAction's guard chain, which needs a live DB.
//
// MUTATION: delete the auditCTALabelAgreement call from
// save_page_sections_action.go. This fails.
func TestCTALabelAuditIsWiredIntoTheSavePath(t *testing.T) {
	src := readSourceFile(t, "save_page_sections_action.go")
	if !strings.Contains(src, "auditCTALabelAgreement(ctx, params, siteID,") {
		t.Fatal("save_page_sections_action.go no longer calls auditCTALabelAgreement — " +
			"the audit is inert, and every other test in this file would still pass")
	}
}

// TestCTALabelAuditAsksTheSharedQuestion is the ANTI-DRIFT scan, and it is the
// one that enforces RFC_047 §9 mechanically rather than by comment. The detector
// and this audit must ask ONE question; a private comparison re-inlined here is
// the drift that bugs_open/203's extraction and bugs_open/308 both exist to stop.
//
// MUTATION: replace the JudgeCTALabel call with a direct BestLabelMatchForPage
// or a label-vs-target_title token comparison. This fails.
func TestCTALabelAuditAsksTheSharedQuestion(t *testing.T) {
	audit := readSourceFile(t, "cta_label_audit.go")
	if !strings.Contains(audit, "datahelpers.JudgeCTALabel(") {
		t.Error("cta_label_audit.go does not call datahelpers.JudgeCTALabel")
	}
	for _, forbidden := range []string{"BestLabelMatch(", "BestLabelMatchForPage(", "LabelTokens("} {
		if strings.Contains(audit, forbidden) {
			t.Errorf("cta_label_audit.go calls %s directly — that is a second definition of "+
				"'this button is misdirected' beside the detector's (RFC_047 §9)", forbidden)
		}
	}
	detector := readSourceFile(t, "discovery_checks/check_misdirected_cta.go")
	if !strings.Contains(detector, "datahelpers.JudgeCTALabel(") {
		t.Error("check_misdirected_cta.go no longer routes through datahelpers.JudgeCTALabel — " +
			"the detector and the audit have drifted apart")
	}
}

// TestAuditCTALabelAgreementCannotFailTheSave is the objection the council's
// guardian and bug_historian seats both raised (corr e9bda035): the plan
// ASSERTED "the pass can never fail a save" and never showed it. This shows it.
//
// SavePageSectionsAction's guard chain is the single save seam six orchestration
// pipelines share, and is the write path implicated in this estate's documented
// silent-content-loss family. An instrument that can abort it is a worse defect
// than the one it measures — and "it only reads" is not a defence: a nil map, a
// surprise type in content_data, or a malformed input_schema panics exactly as
// loudly as a write would.
//
// MUTATION: delete the `defer func() { recover() }()` at the top of
// auditCTALabelAgreement. This test panics and fails.
func TestAuditCTALabelAgreementCannotFailTheSave(t *testing.T) {
	// A section whose content_data holds the WRONG TYPE under every key the pass
	// reads, plus a nil-valued entry — the shapes a malformed LLM response and a
	// hand-edited row actually produce.
	hostile := []SectionData{{
		ComponentName: "hero",
		ComponentID:   "c1",
		ContentData: map[string]interface{}{
			"cta_text":         []interface{}{"not", "a", "string"},
			"cta_url":          map[string]interface{}{"nested": true},
			"cta_target_title": nil,
		},
	}}

	// The pure seam must survive it and simply decline to judge: every read is a
	// comma-ok assertion, so a wrong type reads as "" and is skipped.
	got := auditSectionCTALabels(hostile, schemaMap(), auditCandidates(t), "index", "")
	if len(got) != 0 {
		t.Errorf("hostile content_data produced %d finding(s); want 0 — the pass judged values it could not read: %+v", len(got), got)
	}

	// And a nil sections slice, a nil schema map and nil candidates must all be
	// inert rather than a nil-map panic.
	for name, call := range map[string]func() []ctaLabelFinding{
		"nil sections": func() []ctaLabelFinding {
			return auditSectionCTALabels(nil, schemaMap(), auditCandidates(t), "index", "")
		},
		"nil schemas":    func() []ctaLabelFinding { return auditSectionCTALabels(hostile, nil, auditCandidates(t), "index", "") },
		"nil candidates": func() []ctaLabelFinding { return auditSectionCTALabels(hostile, schemaMap(), nil, "index", "") },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("auditSectionCTALabels panicked on %s: %v — this would abort a save", name, r)
				}
			}()
			if n := len(call()); n != 0 {
				t.Errorf("%s produced %d finding(s), want 0", name, n)
			}
		})
	}

	// The WIRED entry point on a nil DB must simply return. ⚠ This arm does NOT
	// exercise the recover — params.DB is nil, so the function returns at its
	// first guard before any logger or query call. Proving the recover needs a
	// mock DB; that is TestCTALabelAuditContainsItsOwnPanic below, and the first
	// draft of this file got that wrong.
	auditCTALabelAgreement(context.Background(), ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{ctaLabelAuditConfigKey: true}},
		Logger:     zap.NewNop(),
	}, uuid.New(), "example.com", "index", "/index.html", hostile, zap.NewNop())
}

// TestCTALabelAuditContainsItsOwnPanic proves the recover is present and doing
// work, by driving a real panic THROUGH the entry point.
//
// ⚠ THE FIRST VERSION OF THIS TEST WAS A FALSE MUTATION CLAIM, and it is worth
// recording where it will be read. It passed a nil logger and asserted "delete
// the recover and this fails" — but with params.DB nil the function returns at
// its first guard, long before any logger call, so no panic was ever driven and
// deleting the recover left the suite GREEN. A mutation that passes has usually
// hit a guard in series; here it never reached the code under test at all.
//
// This version supplies a mock DB so the early guards pass, then fails the
// label-universe query so control reaches logger.Warn with a nil logger — a
// genuine panic, inside the region the recover covers.
//
// MUTATION: delete the `defer func(){ recover() }()` at the top of
// auditCTALabelAgreement. This test then panics and fails. VERIFIED by running it.
func TestCTALabelAuditContainsItsOwnPanic(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// Any query the pass issues fails, which routes to the logger.Warn branch.
	mock.ExpectQuery(".*").WillReturnError(errors.New("induced: universe unavailable"))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic ESCAPED auditCTALabelAgreement: %v — SavePageSectionsAction would abort, "+
				"and this pass is an instrument that must never be able to fail a save", r)
		}
	}()

	// nil logger: the Warn on the failure path dereferences it. The recover is
	// the only thing between that and the save.
	auditCTALabelAgreement(context.Background(), ActionParams{
		DB:         db,
		StepConfig: models.Step{Config: map[string]interface{}{ctaLabelAuditConfigKey: true}},
	}, uuid.New(), "example.com", "index", "/index.html",
		[]SectionData{{ComponentName: "hero", ComponentID: "c1",
			ContentData: map[string]interface{}{"cta_text": "x"}}}, nil)
}
