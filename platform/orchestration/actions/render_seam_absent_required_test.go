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
	emitRequiredFieldsMissing(context.Background(), nil, uuid.New(), nil, "hero", "Section edit on hero",
		"page_component", "section_editor", nil, nil, zap.New(core))
	if logs.Len() != 0 {
		t.Errorf("a clean edit produced %d log line(s) — the empty case must return before every "+
			"other consideration: %v", logs.Len(), logs.All()[0].Message)
	}

	// Something absent but no identity → refuse loudly, never silently.
	core2, logs2 := observer.New(zapcore.DebugLevel)
	emitRequiredFieldsMissing(context.Background(), nil, uuid.Nil, nil, "hero", "Section edit on hero",
		"page_component", "section_editor", []string{"body"}, nil, zap.New(core2))
	if logs2.Len() == 0 {
		t.Error("an identity-less call with real findings said nothing — a record that cannot be " +
			"written must still be visible, or the finding disappears twice over")
	}
}
