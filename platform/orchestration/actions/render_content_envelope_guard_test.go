package actions

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// Tests for the RENDER-seam transport-envelope guard (bugs_open/199).
//
// The mutation that must break each test is named ON the test, following the
// storage guard's test file. That convention matters more here than anywhere,
// because this guard's dominant path is "change nothing" — which every broken
// version of it also does.
//
// The component fixture is the LIVE gate-blind case, not an invented one:
// `pricing` on gaswholesalers.com/how-pricing-works carries the one envelope
// still in page_components.content_data, and its schema declares no required
// source:"llm" field, so missingRequiredLLMFields can never speak for it. That
// is the whole population this guard exists for.

// componentRowFor builds the single-row result GetComponentByID scans, in
// queryComponent's column order (component_library.go:536-539).
func componentRowFor(id, function, htmlTemplate, inputSchema string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "function", "category", "html_template", "input_schema", "is_dark_section",
	}).AddRow(id, function, function, "content", htmlTemplate, inputSchema, false)
}

// renderParamsFor rigs RenderComponentAction with a step output at
// `generated_content`, which is the live content_from path
// (page-content-writer's render_section: content_from "generated_content.result").
func renderParamsFor(t *testing.T, componentID string, stepOutput map[string]interface{}) (ActionParams, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		AgentType:        "page-content-writer",
		ExecutionContext: &types.ExecutionContext{Action: "execute", StepName: "render_section"},
		StepConfig: models.Step{Config: map[string]interface{}{
			"component_id": componentID,
			"content_from": "generated_content.result",
			"context_from": "render_context",
		}},
		CollectedData: map[string]interface{}{
			"generated_content": stepOutput,
			"render_context":    map[string]interface{}{"company_name": "Gas Wholesalers"},
			"input_data": map[string]interface{}{
				"site_id":      "6f1d8c2e-0000-4000-8000-00000000abcd",
				"current_page": map[string]interface{}{"name": "how-pricing-works"},
			},
		},
	}
	return params, mock, func() { db.Close() }
}

// --- The bug file's demanded assertion #1 ----------------------------------
//
// "Take a component with an empty input schema, drive a step whose LLM output
// falls to the text path, and assert the render REFUSES rather than producing
// an empty section."
//
// The payload is the live REFUSE tier: prose that ParseLLMJSONWithProvenance
// cannot recover losslessly. The storage seam refuses this exact shape, so the
// two seams must agree.
//
// MUTATION THAT MUST BREAK IT: delete the normalizeRenderContentEnvelope call
// from RenderComponentAction. The action then returns rendered_html for a
// section whose every field is blank — the defect, reported as success.
func TestRenderRefusesEnvelopeForSchemalessComponent(t *testing.T) {
	componentID := uuid.NewString()
	params, mock, closeDB := renderParamsFor(t, componentID, map[string]interface{}{
		"type":   "text",
		"result": "Our pricing works on a tiered basis. Tier 1 covers the first 50,000 therms...",
	})
	defer closeDB()

	// Empty input_schema — the gate at v3_site_actions.go:1843 is skipped
	// entirely, so nothing but this guard stands between the envelope and the
	// template.
	mock.ExpectQuery("FROM content_components").
		WithArgs(componentID).
		WillReturnRows(componentRowFor(componentID, "pricing", "<section>{{.headline}}</section>", `{}`))
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := RenderComponentAction(context.Background(), params)
	if err == nil {
		t.Fatalf("render SUCCEEDED on a transport envelope with an empty input schema — the guard is not firing; got %#v", result)
	}
	if result != nil {
		t.Errorf("refused render must return a nil result, got %#v", result)
	}
	for _, want := range []string{"pricing", "generated_content.result", "bugs_open/199"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal must name %q so the operator can find the section; got: %v", want, err)
		}
	}
}

// --- The decode branch, which is why this is not merely a refusal ----------
//
// A losslessly-recoverable envelope renders its REAL content. Today the same
// payload renders a blank section and the storage guard repairs only the
// column, never the HTML it already shipped.
//
// MUTATION THAT MUST BREAK IT: make the render seam refuse every envelope
// instead of decoding the lossless tier. The section then fails where it should
// have rendered, and the storage seam's approved decode branch becomes
// unreachable on the only live path (render output feeds the save).
func TestRenderDecodesLosslessEnvelopeAndRendersRealContent(t *testing.T) {
	componentID := uuid.NewString()
	params, mock, closeDB := renderParamsFor(t, componentID, map[string]interface{}{
		"type":   "text",
		"result": `{"headline":"Real headline","body":"<p>Real copy.</p>"}`,
	})
	defer closeDB()

	mock.ExpectQuery("FROM content_components").
		WithArgs(componentID).
		WillReturnRows(componentRowFor(componentID, "pricing", "<section>{{.headline}}</section>", `{}`))
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := RenderComponentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("a losslessly-decodable envelope must render, not refuse: %v", err)
	}

	out, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if html, _ := out["rendered_html"].(string); !strings.Contains(html, "Real headline") {
		t.Errorf("decoded content never reached the template; rendered_html = %q", html)
	}

	// The resolved map also becomes result["content_data"], which is what the
	// save seam stores — so a decode here must leave the storage guard nothing
	// to do.
	contentData, ok := out["content_data"].(map[string]interface{})
	if !ok {
		t.Fatalf("content_data missing from the render result: %#v", out)
	}
	if isLLMTransportEnvelope(contentData) {
		t.Errorf("content_data is still an envelope after the decode: %#v", contentData)
	}
	for _, k := range []string{"type", "result"} {
		if _, present := contentData[k]; present {
			t.Errorf("decoded content_data still carries the transport key %q: %#v", k, contentData)
		}
	}
	if contentData["headline"] != "Real headline" {
		t.Errorf("decoded content_data lost the payload: %#v", contentData)
	}
}

// --- The bug file's demanded assertion #2 ----------------------------------
//
// "Confirm a normally-schema'd component is byte-identical through the change."
//
// Asserted two ways, because they fail differently: BYTE identity through the
// normaliser (a bad predicate rewrites a map nobody was looking at), and an
// action-level run proving the render still succeeds and hands the same
// content_data on to the save.
//
// MUTATION THAT MUST BREAK IT: weaken isLLMTransportEnvelope itself — predicate
// on the presence of `type` alone, or of `result` alone. Verified: that turns
// fixtures 1 and 2 into envelopes and the test fails.
//
// AND THE ONE THAT DOES NOT, which is the useful half. Weakening only the
// fast-exit at the top of normalizeRenderContentEnvelope leaves this test GREEN,
// because normalizeContentDataEnvelope re-checks the same predicate and returns
// (m, false, nil) for a non-envelope — the two are guards in SERIES. So the
// fast-exit is an optimisation and a log-suppression, NOT the decision: byte
// identity at this seam is inherited from the storage normaliser's no-op
// contract. Anyone hardening this file should mutate the shared predicate to
// test it, and anyone tempted to "simplify" by trusting the local check alone
// should read that sentence twice.
func TestRenderNonEnvelopeContentByteIdentical(t *testing.T) {
	fixtures := []struct {
		name string
		in   map[string]interface{}
	}{
		{"type text but no string result", map[string]interface{}{
			"type": "text", "headline": "Wholesale Fuel Supply", "body": "<p>Real content.</p>",
		}},
		{"string result but not type text", map[string]interface{}{
			"result": "42", "label": "Answer", "type": "calculator",
		}},
		{"json-path envelope, result is an object", map[string]interface{}{
			"type": "json", "result": map[string]interface{}{"headline": "Hi"},
		}},
		{"ordinary component content", map[string]interface{}{
			"headline": "Pricing", "items": []interface{}{"a", "b"},
		}},
	}

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			before, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			out, err := normalizeRenderContentEnvelope(
				context.Background(),
				ActionParams{Logger: zap.NewNop()}, // DB nil — the log writer must no-op
				&Component{Function: "pricing"},
				"generated_content.result",
				tc.in,
			)
			if err != nil {
				t.Fatalf("legitimate content was REFUSED: %v", err)
			}
			after, err := json.Marshal(out)
			if err != nil {
				t.Fatalf("marshal result: %v", err)
			}
			if string(before) != string(after) {
				t.Errorf("guard rewrote legitimate content\n before: %s\n  after: %s", before, after)
			}
		})
	}

	t.Run("action level, schema'd component with real content", func(t *testing.T) {
		componentID := uuid.NewString()
		params, mock, closeDB := renderParamsFor(t, componentID, map[string]interface{}{
			"type":   "json",
			"result": map[string]interface{}{"headline": "Pricing that scales"},
		})
		defer closeDB()

		schema := `{"fields":{"headline":{"source":"llm","required":true,"type":"text"}}}`
		mock.ExpectQuery("FROM content_components").
			WithArgs(componentID).
			WillReturnRows(componentRowFor(componentID, "pricing", "<section>{{.headline}}</section>", schema))
		// No agent_error_log INSERT expected: nothing fired.

		result, err := RenderComponentAction(context.Background(), params)
		if err != nil {
			t.Fatalf("a normally-schema'd component with good content must render unchanged: %v", err)
		}
		out := result.(map[string]interface{})
		contentData := out["content_data"].(map[string]interface{})
		if contentData["headline"] != "Pricing that scales" {
			t.Errorf("content_data changed through the guard: %#v", contentData)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unexpected database traffic for a no-op render: %v", err)
		}
	})
}

// --- The identity chain, which is NOT the save seam's -----------------------
//
// MUTATION THAT MUST BREAK IT: read `site_record.site_id` / `current_page.name`,
// the paths writeContentDataEnvelopeLog uses at the save seam. Measured live
// 2026-08-05 across every stored page-content-writer run (n=110), both resolve
// 0/110 here — copying them would produce a guard that fires correctly and
// records an unattributable row every time.
func TestRenderEnvelopeIdentityFallbacks(t *testing.T) {
	siteID := "6f1d8c2e-0000-4000-8000-00000000abcd"

	cases := []struct {
		name       string
		collected  map[string]interface{}
		wantSite   string
		wantPage   string
		wantDomain string
	}{
		{
			name: "the common live shape: input_data.site_id + input_data.current_page.name",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"site_id":      siteID,
					"current_page": map[string]interface{}{"name": "how-pricing-works"},
					"domain":       "gaswholesalers.com",
				},
			},
			wantSite: siteID, wantPage: "how-pricing-works", wantDomain: "gaswholesalers.com",
		},
		{
			name: "render_context fallbacks — current_page is a plain STRING here, not a map",
			collected: map[string]interface{}{
				"render_context": map[string]interface{}{
					"site_id": siteID, "current_page": "services", "domain": "gaswholesalers.com",
				},
			},
			wantSite: siteID, wantPage: "services", wantDomain: "gaswholesalers.com",
		},
		{
			name: "input_data.page_name, the third page rung",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{"site_id": siteID, "page_name": "contact"},
			},
			wantSite: siteID, wantPage: "contact", wantDomain: "",
		},
		{
			name:      "nothing resolvable — degrades, never panics",
			collected: map[string]interface{}{"unrelated": 1},
			wantSite:  uuid.Nil.String(), wantPage: "", wantDomain: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotSite, gotPage, gotDomain := renderEnvelopeIdentity(ActionParams{CollectedData: tc.collected})
			if gotSite.String() != tc.wantSite {
				t.Errorf("site_id: got %s, want %s", gotSite, tc.wantSite)
			}
			if gotPage != tc.wantPage {
				t.Errorf("page_name: got %q, want %q", gotPage, tc.wantPage)
			}
			if gotDomain != tc.wantDomain {
				t.Errorf("domain: got %q, want %q", gotDomain, tc.wantDomain)
			}
		})
	}
}
