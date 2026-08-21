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
	"strings"
	"testing"

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
