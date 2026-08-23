// FILE: platform/orchestration/actions/render_seam_absent_required_test.go
//
// bugs_open/342. Go's missingkey=zero renders a field the content never supplied
// as EMPTY WITH NO ERROR; page assembly then drops the visually-empty section,
// so the content does not arrive broken — it does not arrive at all. That is the
// mechanism behind the fleet-wide blanking of article bodies
// (bugs_closed/004/005), and until 2026-08-21 the gate that catches it ran at
// TWO of the fifteen render call sites.
//
// The seam now applies the SAME rule for every caller. These tests assert the
// report exists, and — the half that matters — that it stays SILENT in the three
// cases where firing would be wrong.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func schemaRequiring(fields ...string) map[string]interface{} {
	f := map[string]interface{}{}
	for _, name := range fields {
		f[name] = map[string]interface{}{"source": "llm", "required": true, "type": "text"}
	}
	return map[string]interface{}{"fields": f}
}

// absentRequiredReport renders and returns the fields named by the
// absent-required Error line, plus whether that line fired at all.
func absentRequiredReport(t *testing.T, ctx *RenderContext, tmpl string) ([]string, bool) {
	t.Helper()
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	if _, _, _, err := RenderTemplate(tmpl, ctx, logger); err != nil {
		t.Fatalf("fixture must render: %v", err)
	}

	for _, entry := range logs.All() {
		if !strings.Contains(entry.Message, "REQUIRED content field(s) absent") {
			continue
		}
		if entry.Level != zapcore.ErrorLevel {
			t.Errorf("the absent-required report is %v, want Error — a required field is a stated "+
				"contract and the section is about to be dropped; Warn is what it already was", entry.Level)
		}
		// ContextMap(), not a type assertion on Field.Interface: zap.Strings
		// stores an ArrayMarshaler, not a []string, so asserting the concrete
		// type silently yields nil — which reads exactly like "the report named
		// no fields". The first version of this helper did that and the test
		// caught it, which is the only reason it is written this way.
		raw, ok := entry.ContextMap()["absent_required_fields"]
		if !ok {
			return nil, true
		}
		var got []string
		switch v := raw.(type) {
		case []string:
			got = v
		case []interface{}:
			for _, x := range v {
				if str, ok := x.(string); ok {
					got = append(got, str)
				}
			}
		}
		return got, true
	}
	return nil, false
}

// The defect itself: a required field the content never supplied.
func TestAbsentRequiredFieldIsReported(t *testing.T) {
	ctx := &RenderContext{
		InputSchema: schemaRequiring("headline", "body"),
		ContentData: map[string]interface{}{"headline": "Present"},
	}

	got, fired := absentRequiredReport(t, ctx, `<h1>{{.headline}}</h1><div>{{.body}}</div>`)
	if !fired {
		t.Fatal("an absent REQUIRED field rendered silently — this is bugs_open/342 and the whole " +
			"point of carrying the schema on the context")
	}
	if strings.Join(got, ",") != "body" {
		t.Errorf("reported fields = %v, want [body] — the report must name the absent one and only it", got)
	}
}

// CONTROL 1 — complete content must be silent. A report that fires on a healthy
// render is noise, and noise on every page is how a real signal gets filtered
// out by whoever reads the logs.
func TestCompleteContentIsSilent(t *testing.T) {
	ctx := &RenderContext{
		InputSchema: schemaRequiring("headline", "body"),
		ContentData: map[string]interface{}{"headline": "Present", "body": "Also present"},
	}
	if _, fired := absentRequiredReport(t, ctx, `<h1>{{.headline}}</h1><div>{{.body}}</div>`); fired {
		t.Error("CONTROL FAILED: complete content produced an absent-required report")
	}
}

// CONTROL 2 — an OPTIONAL absent field must be silent here. It is already
// covered, at Warn, by the empty-placeholder report; escalating it would erase
// the distinction this change exists to draw.
func TestAbsentOptionalFieldIsNotReportedAsRequired(t *testing.T) {
	ctx := &RenderContext{
		InputSchema: map[string]interface{}{"fields": map[string]interface{}{
			"headline": map[string]interface{}{"source": "llm", "required": true, "type": "text"},
			"footnote": map[string]interface{}{"source": "llm", "type": "text"},
		}},
		ContentData: map[string]interface{}{"headline": "Present"},
	}
	if got, fired := absentRequiredReport(t, ctx, `<h1>{{.headline}}</h1><small>{{.footnote}}</small>`); fired {
		t.Errorf("CONTROL FAILED: an absent OPTIONAL field was reported as required: %v", got)
	}
}

// CONTROL 3 — NO SCHEMA means UNKNOWN, not valid. Thirteen of the fifteen call
// sites pass nil today; the seam must then behave exactly as it did before, and
// this test is the difference between "fail-open, stated" and "fail-open,
// discovered later".
func TestNoSchemaMeansNoReport(t *testing.T) {
	ctx := &RenderContext{ContentData: map[string]interface{}{}}
	if _, fired := absentRequiredReport(t, ctx, `<h1>{{.headline}}</h1>`); fired {
		t.Error("CONTROL FAILED: a nil schema produced a report — the seam must not invent a contract")
	}
}

// A NON-LLM required field is the resolver's business, not the writer's: the
// resolver fills it, and its absence is reported by the fill path with far more
// context than the seam has.
func TestRequiredNonLLMFieldIsNotTheSeamsBusiness(t *testing.T) {
	ctx := &RenderContext{
		InputSchema: map[string]interface{}{"fields": map[string]interface{}{
			"news_items": map[string]interface{}{"source": "query", "required": true, "type": "array"},
		}},
		ContentData: map[string]interface{}{},
	}
	if got, fired := absentRequiredReport(t, ctx, `<ul>{{range .news_items}}<li>x</li>{{end}}</ul>`); fired {
		t.Errorf("CONTROL FAILED: a resolver-sourced field was reported at the seam: %v", got)
	}
}

// THE SEAM AND THE PRE-RENDER GATE DISAGREE, AND BOTH ARE RIGHT — pinned here
// with the reason, because the obvious "fix" is to make one match the other and
// that would break whichever one you changed.
//
// They ask different questions. The pre-render gate asks "did the WRITER supply
// this required field?" and refuses on that. The seam asks "will it RENDER
// EMPTY?", because 342's damage is empty -> page assembly drops the section ->
// the content vanishes. contextToInterfaceMap supplies fleet defaults (cta_text
// falls back to "Get Started"), so a field the writer never produced can still
// render something — the section survives, and there is nothing for the seam to
// report.
//
// So the seam's report is a SUBSET of the gate's. This test found that
// divergence rather than assuming it: its first version asserted the two lists
// were equal and failed, naming cta_text.
func TestSeamReportsASubsetOfThePreRenderGateAndSaysWhy(t *testing.T) {
	schema := schemaRequiring("headline", "body", "cta_text")
	content := map[string]interface{}{"headline": "Present", "body": "   "} // blank counts as absent

	gate := missingRequiredLLMFields(schema, content)
	if strings.Join(gate, ",") != "body,cta_text" {
		t.Fatalf("fixture drifted: the pre-render gate reports %v, want [body cta_text]", gate)
	}

	ctx := &RenderContext{InputSchema: schema, ContentData: content}
	got, fired := absentRequiredReport(t, ctx, `<h1>{{.headline}}</h1><div>{{.body}}</div><a>{{.cta_text}}</a>`)
	if !fired {
		t.Fatal("the seam stayed silent on a required field that renders empty")
	}

	// SUBSET, and specifically: body yes (nothing defaults it), cta_text no
	// (contextToInterfaceMap defaults it, so it renders "Get Started").
	if strings.Join(got, ",") != "body" {
		t.Errorf("seam reports %v, want [body] only. cta_text must NOT appear: the render context "+
			"defaults it, so it does not render empty and the section is not dropped. If this now "+
			"reports cta_text, either the default was removed (then the gate and the seam agree "+
			"again and this test should say so) or the seam has started judging writer output "+
			"instead of render data (then it is duplicating the pre-render gate).", got)
	}
	for _, f := range got {
		found := false
		for _, g := range gate {
			if f == g {
				found = true
			}
		}
		if !found {
			t.Errorf("seam reported %q, which the pre-render gate does not — the seam must never be "+
				"a SUPERSET: it would then be refusing-by-log on content the gate considers complete", f)
		}
	}
}

// ── The escalation half (council bb7f5d0e round 1, bug_historian, GATING) ──
//
// The first version of this fix reported at Error and stopped. This estate has an
// owner ruling that a named log is NOT escalation (bugs_open/054, 2026-07-22),
// earned on exactly this shape: bugs_open/018 shipped the observability half for
// dead controls, thirty of them shipped anyway, and the ruling was "make it MEAN
// something". So the seam PUBLISHES its finding for a caller that can act.

// The seam must publish, not merely log — a caller cannot escalate what it cannot
// read, and parsing your own log lines is not a channel.
func TestSeamPublishesAbsentRequiredForCallersToAct(t *testing.T) {
	ctx := &RenderContext{
		InputSchema: schemaRequiring("headline", "body"),
		ContentData: map[string]interface{}{"headline": "Present"},
	}
	if _, _, _, err := RenderTemplate(`<h1>{{.headline}}</h1><div>{{.body}}</div>`, ctx, zap.NewNop()); err != nil {
		t.Fatalf("fixture must render: %v", err)
	}
	if strings.Join(ctx.AbsentRequiredFields, ",") != "body" {
		t.Fatalf("ctx.AbsentRequiredFields = %v, want [body] — without this a caller with a database "+
			"handle has nothing to file, and the fix is a log line again", ctx.AbsentRequiredFields)
	}
}

// CONTROL: a clean render must publish NOTHING, or every caller that reads the
// field files an item on every healthy page.
func TestCleanRenderPublishesNoAbsentRequired(t *testing.T) {
	ctx := &RenderContext{
		InputSchema: schemaRequiring("headline"),
		ContentData: map[string]interface{}{"headline": "Present"},
	}
	if _, _, _, err := RenderTemplate(`<h1>{{.headline}}</h1>`, ctx, zap.NewNop()); err != nil {
		t.Fatalf("fixture must render: %v", err)
	}
	if len(ctx.AbsentRequiredFields) != 0 {
		t.Errorf("CONTROL FAILED: a clean render published %v", ctx.AbsentRequiredFields)
	}
}

// The record is OPT-IN with the unsafe default OFF, because it is a new DB write
// on a shared render path — the shape three seats made the dead-URL record arm
// justify on council 98852baa. Unset must mean today's behaviour, byte for byte.
func TestAbsentRequiredRecordIsOptInAndFailsOpen(t *testing.T) {
	cases := map[string]struct {
		config map[string]interface{}
		want   bool
	}{
		"unset":              {map[string]interface{}{}, false},
		"nil config":         {nil, false},
		"explicitly false":   {map[string]interface{}{"record_absent_required_fields": false}, false},
		"armed":              {map[string]interface{}{"record_absent_required_fields": true}, true},
		"mistyped string":    {map[string]interface{}{"record_absent_required_fields": "true"}, false},
		"mistyped number":    {map[string]interface{}{"record_absent_required_fields": 1}, false},
		"someone else's key": {map[string]interface{}{"refuse_mistyped_llm_fields": true}, false},
	}
	for name, c := range cases {
		if got := recordAbsentRequiredFields(c.config); got != c.want {
			t.Errorf("%s: recordAbsentRequiredFields = %v, want %v — a config value that is not a "+
				"bool is a mistake, and a mistake must not switch on a fleet-wide DB write", name, got, c.want)
		}
	}
}

// The live-page routes ESCALATE, they do not merely detect (council bb7f5d0e
// round 5 — editquality and bug_historian, both HIGH, both against my own words:
// the submission called these two "the two with the most exposure" and then
// wired only a log line there).
//
// ONE emitter serves all three render-time producers since round 6 (reuse_agent:
// three near-identical writers of one item_type is how three subtly different
// behaviours start). The emitter needs a database, so what is asserted here is
// the DECISION, which is the part that was wrong: nothing to report must file
// nothing, and an identity-less call must refuse rather than panic or write.
func TestRequiredFieldsMissingEmitterDecidesCorrectlyWithoutADatabase(t *testing.T) {
	// Nothing absent → no attempt at all. If this ever reaches the db==nil guard
	// it would log a warning on every healthy edit, which is how a real signal
	// gets filtered out.
	core, logs := observer.New(zapcore.DebugLevel)
	emitRequiredFieldsMissing(context.Background(), nil, uuid.New(), pageContext{}, nil, "hero", "Section edit on hero",
		"page_component", "section_editor", nil, nil, zap.New(core))
	if logs.Len() != 0 {
		t.Errorf("a clean edit produced %d log line(s) — the empty case must return before every "+
			"other consideration: %v", logs.Len(), logs.All()[0].Message)
	}

	// Something absent but no identity → refuse loudly, never silently.
	core2, logs2 := observer.New(zapcore.DebugLevel)
	emitRequiredFieldsMissing(context.Background(), nil, uuid.Nil, pageContext{}, nil, "hero", "Section edit on hero",
		"page_component", "section_editor", []string{"body"}, nil, zap.New(core2))
	if logs2.Len() == 0 {
		t.Error("an identity-less call with real findings said nothing — a record that cannot be " +
			"written must still be visible, or the finding disappears twice over")
	}
}

// ── bugs_open/342, the REFUSAL half (2026-08-22) ───────────────────────────
//
// The two section-editor routes write rendered_html straight onto an
// already-live page. Until this change they filed the required_fields_missing
// item and then persisted the blank section anyway. The refusal declines the
// persist and leaves the live section untouched — opt-in, default OFF, per the
// owner ruling of 2026-08-02 §2, because such an edit SUCCEEDS today.

// Key semantics, mirroring both siblings exactly: unset is OFF, and a mistyped
// value is a mistake that must not switch a refusal on by accident.
func TestEditorAbsentRequiredRefusalIsOptInAndFailsOpen(t *testing.T) {
	cases := map[string]struct {
		config map[string]interface{}
		want   bool
	}{
		"unset":            {map[string]interface{}{}, false},
		"nil config":       {nil, false},
		"explicitly false": {map[string]interface{}{"refuse_absent_required_fields": false}, false},
		"armed":            {map[string]interface{}{"refuse_absent_required_fields": true}, true},
		"mistyped string":  {map[string]interface{}{"refuse_absent_required_fields": "true"}, false},
		"mistyped number":  {map[string]interface{}{"refuse_absent_required_fields": 1}, false},
		// Cross-contamination: neither sibling key may arm this one.
		"record key is not this key":  {map[string]interface{}{"record_absent_required_fields": true}, false},
		"mistype key is not this key": {map[string]interface{}{"refuse_mistyped_llm_fields": true}, false},
	}
	for name, c := range cases {
		if got := refuseAbsentRequiredFields(c.config); got != c.want {
			t.Errorf("%s: refuseAbsentRequiredFields = %v, want %v — a config value that is not a "+
				"bool is a mistake, and a mistake must not refuse a live-page edit by accident", name, got, c.want)
		}
	}
}

// The deciding arm as the action runs it: BOTH halves are required. Unarmed
// must mean today's behaviour byte for byte even when the render left required
// fields empty (that is what "opt-in" means), and an armed step with a clean
// render must persist normally — a refusal that fires on clean renders has
// merely stopped completing edits, which is the failure mode the positive
// control in bugs_open/348 §8 exists to catch.
func TestEditorRefusalNeedsBothArmingAndAFinding(t *testing.T) {
	armed := map[string]interface{}{"refuse_absent_required_fields": true}
	unarmed := map[string]interface{}{}
	finding := []string{"headline"}

	cases := map[string]struct {
		config map[string]interface{}
		absent []string
		want   bool
	}{
		"armed + finding = refuse":            {armed, finding, true},
		"armed + clean render = persist":      {armed, nil, false},
		"armed + empty slice = persist":       {armed, []string{}, false},
		"unarmed + finding = persist (today)": {unarmed, finding, false},
		"unarmed + clean = persist":           {unarmed, nil, false},
	}
	for name, c := range cases {
		if got := refusePersistForAbsentRequired(c.config, c.absent); got != c.want {
			t.Errorf("%s: refusePersistForAbsentRequired = %v, want %v", name, got, c.want)
		}
	}
}

// Pins the seam→outcome→gate chain SHAPE: a real render's published finding,
// carried on a sectionEditOutcome, must trip an armed gate. HONEST LIMIT: the
// copy inside applyContentEdit/applyComponentSwap is performed here by the
// test (those helpers need a database), so a branch that stops copying
// RenderContext.AbsentRequiredFields onto the outcome is NOT caught by this —
// that is what the post-roll live canary is for (PLAN, verification section).
func TestSeamFindingSurvivesOntoTheEditOutcome(t *testing.T) {
	ctx := &RenderContext{
		ContentData: map[string]interface{}{},
		InputSchema: schemaRequiring("body"),
	}
	core, _ := observer.New(zapcore.ErrorLevel)
	if _, _, _, err := RenderTemplate(`<article>{{.body}}</article>`, ctx, zap.New(core)); err != nil {
		t.Fatalf("fixture must render: %v", err)
	}
	if len(ctx.AbsentRequiredFields) == 0 {
		t.Fatal("seam published nothing — the fixture no longer exercises the defect")
	}
	// The copy the branches perform, shape-for-shape.
	outcome := sectionEditOutcome{AbsentRequiredFields: ctx.AbsentRequiredFields}
	if !refusePersistForAbsentRequired(map[string]interface{}{"refuse_absent_required_fields": true}, outcome.AbsentRequiredFields) {
		t.Error("an armed step did not refuse the outcome carrying the seam's own finding — " +
			"the seam→outcome→gate chain is broken somewhere the two half-tests cannot see")
	}
}

// The chrome store path gets the SAME refusal capability as the section editor
// (council 3626629a round 1, bug_historian, medium): arming DETECTION on chrome
// while only the editor got PROTECTION would reproduce this very bug's shape on
// the sibling call site — 016b §9's "one call site of a shared judgement gets
// the rigorous fix, the sibling stays heuristic".
//
// What is asserted here is that the two paths share ONE decision function, so
// they cannot drift: the chrome caller passes its own bool through the same
// refusePersistForAbsentRequired the editor's persist switch calls. A second
// implementation on the chrome side is exactly what this test exists to fail.
func TestChromeAndEditorShareOneRefusalDecision(t *testing.T) {
	finding := []string{"headline"}

	// The chrome path's call shape: a plain bool lifted into the same key the
	// editor reads from step config.
	chrome := func(armed bool, absent []string) bool {
		return refusePersistForAbsentRequired(
			map[string]interface{}{absentRequiredRefuseConfigKey: armed}, absent)
	}

	if chrome(true, finding) != true {
		t.Error("armed chrome slot with an absent required field did not refuse")
	}
	if chrome(true, nil) != false {
		t.Error("armed chrome slot with a CLEAN render refused — a refusal that fires on clean " +
			"renders has merely stopped storing chrome")
	}
	if chrome(false, finding) != false {
		t.Error("UNARMED chrome slot refused — unset must mean today's behaviour byte for byte; " +
			"no migration arms this key and the measured population is zero, so an unarmed " +
			"refusal here would be a fleet-wide behaviour change nobody opted into")
	}

	// And the editor, through its own config map, must agree on every case —
	// same function, so disagreement means someone forked the decision.
	for _, c := range []struct {
		armed  bool
		absent []string
	}{{true, finding}, {true, nil}, {false, finding}, {false, nil}} {
		editorCfg := map[string]interface{}{}
		if c.armed {
			editorCfg[absentRequiredRefuseConfigKey] = true
		}
		if got, want := refusePersistForAbsentRequired(editorCfg, c.absent), chrome(c.armed, c.absent); got != want {
			t.Errorf("editor and chrome disagree (armed=%v, absent=%v): editor=%v chrome=%v — "+
				"the two persist paths must share one decision or they will drift apart",
				c.armed, c.absent, got, want)
		}
	}
}

// ── bugs_open/342: the item must be ROUTABLE, not merely filed (2026-08-23) ──
//
// The first item this emitter ever filed in production (`a31da7f3`) was
// classified `malformed` by required-fields-missing-handler, failed three
// times and parked — because the emitter supplied neither `page_name` nor
// `slot_name`, which is what that handler's `classify` step resolves the page
// and component by, and keyed on `<site_id>:<function>` while the post-deploy
// producer keys on `<page_id>:<slot_name>`.
//
// Both defects came from the same false assumption, stated in the code, the
// bug file, the register AND the approved council submission: that reusing the
// item TYPE meant inheriting its router. **Reusing a type is not reusing its
// contract**, and nothing asserted the contract — so this test does.
//
// It deliberately tests the DECISION rather than the insert (the emitter needs
// a database): the key shape and the routing disposition are what were wrong,
// and both are computable without one.
func TestRequiredFieldsMissingItemsAreRoutable(t *testing.T) {
	pageID := uuid.New()
	siteID := uuid.New()

	// WITH a page: must take the post-deploy check's key shape exactly
	// (check_required_fields_missing.go:180) so the two producers co-dedup on
	// one key instead of filing two items for one defect.
	page := pageContext{id: &pageID, name: "ai-agent-roi-estimator", slot: "tool-cta"}
	wantKey := "required_fields_missing:" + pageID.String() + ":tool-cta"
	gotKey, gotHandler, gotStatus := requiredFieldsMissingRouting(siteID, page, "tool-cta")
	if gotKey != wantKey {
		t.Errorf("page-scoped item_key = %q, want %q — a key that does not match "+
			"check_required_fields_missing's means the two producers file TWO items for ONE defect",
			gotKey, wantKey)
	}
	if gotHandler != "required-fields-missing-handler" {
		t.Errorf("page-scoped handler = %q, want the router", gotHandler)
	}
	if gotStatus != "detected" {
		t.Errorf("page-scoped status = %q, want detected (the router's intake status)", gotStatus)
	}

	// WITHOUT a page (chrome: slots hang off the SITE): the page-resolving
	// router cannot classify this, so it must NOT be handed over. Parked for a
	// human is the honest disposition; routing it buys three failed attempts.
	gotKey, gotHandler, gotStatus = requiredFieldsMissingRouting(siteID, pageContext{}, "header")
	if gotHandler != "" {
		t.Errorf("chrome handler = %q, want empty — required-fields-missing-handler resolves the "+
			"page by spec.page_name, which a chrome slot does not have; handing it over reproduces "+
			"the malformed-then-failed loop that item a31da7f3 hit", gotHandler)
	}
	if gotStatus != "needs_human_review" {
		t.Errorf("chrome status = %q, want needs_human_review (the estate's parking vocabulary — "+
			"the router's own park_* steps use it)", gotStatus)
	}
	if gotKey != "required_fields_missing:"+siteID.String()+":header" {
		t.Errorf("chrome item_key = %q, want the site-scoped shape", gotKey)
	}

	// A page WITHOUT a slot cannot be keyed like the check either — it must
	// fall back rather than emit a half-formed page-scoped key.
	gotKey, gotHandler, _ = requiredFieldsMissingRouting(siteID, pageContext{id: &pageID, name: "x"}, "hero")
	if gotHandler != "" || gotKey != "required_fields_missing:"+siteID.String()+":hero" {
		t.Errorf("page-without-slot took the routed path (key=%q handler=%q) — the router needs BOTH "+
			"page_name and slot_name, so half the context must not look like all of it", gotKey, gotHandler)
	}
}
