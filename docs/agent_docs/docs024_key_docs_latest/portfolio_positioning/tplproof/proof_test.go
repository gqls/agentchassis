//go:build tplproof

// Render proof for migration 764 (bugs_open/453), through the FLEET'S OWN renderer.
//
// Council round 1 (888e7319, editquality): the first harness built its own text/template + funcmap,
// which proves the template text but not the production injection path. This version calls
// datahelpers.RenderPromptTemplate — the function ai_actions.go:328 calls for execute_llm_prompt,
// the action BOTH edited steps and domain-strategist's analyze_strategy run under — and
// datahelpers.ScanMissingValues, PRC-003's own attribution scan, so the hole and its disappearance
// are measured by the instrument the fleet uses, not by a string count of my own.
//
// Run:  go test -tags tplproof ./docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/tplproof/ -run TestBrief764 -v
// after pulling the two live templates to <type>.tpl (README).
package tplproof

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var repl = map[string]string{
	"{{.site_specs.specs.mission_brief.text}}": "{{if .site_specs.specs.mission_brief.text}}{{.site_specs.specs.mission_brief.text}}{{else}}{{toJSON .site_specs.specs.mission_brief}}{{end}}",
	"{{.site_specs.specs.roadmap_brief.text}}": "{{if .site_specs.specs.roadmap_brief.text}}{{.site_specs.specs.roadmap_brief.text}}{{else}}{{toJSON .site_specs.specs.roadmap_brief}}{{end}}",
}

func ctx(brief interface{}) map[string]interface{} {
	specs := map[string]interface{}{"briefing": "b", "classification": map[string]interface{}{"category": "hub"}, "identity": map[string]interface{}{"name": "x"}, "strategy": map[string]interface{}{"site_type": "authority-portal"}}
	if brief != nil {
		specs["mission_brief"] = brief
	}
	return map[string]interface{}{
		"site_specs":  map[string]interface{}{"specs": specs},
		"input_data":  map[string]interface{}{"domain": "copyonline.co.uk"},
		"site_record": map[string]interface{}{"domain": "copyonline.co.uk"},
	}
}

// render returns what production SENDS (RenderPromptTemplate, which under PRC-003 strips <no value>
// before returning) and the fleet's own MissingValueReport measured the way RenderPromptTemplate
// measures it: on the RAW execution, parsed with the SAME PromptTemplateFuncs() production parses
// with (data_helpers.go:1204). Round-1 of this harness scanned the already-stripped output and read
// zero holes everywhere — recorded so nobody repeats it.
func render(t *testing.T, tpl string, data map[string]interface{}) (string, datahelpers.MissingValueReport) {
	t.Helper()
	sent, err := datahelpers.RenderPromptTemplate(tpl, data, *zap.NewNop())
	if err != nil {
		t.Fatalf("RenderPromptTemplate: %v", err)
	}
	tm, err := template.New("agent_prompt").Funcs(datahelpers.PromptTemplateFuncs()).Parse(tpl)
	if err != nil {
		t.Fatalf("parse with production funcmap: %v", err)
	}
	var raw bytes.Buffer
	if err := tm.Execute(&raw, data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	rep := datahelpers.ScanMissingValues(tpl, raw.String(), data)
	return sent, rep
}

func TestBrief764(t *testing.T) {
	for _, name := range []string{"build-site-planner", "domain-research-classifier"} {
		raw, err := os.ReadFile(name + ".tpl")
		if err != nil {
			t.Fatalf("%s: %v (pull the live template first — README)", name, err)
		}
		orig := string(raw)
		fixed := orig
		for a, r := range repl {
			if n := strings.Count(fixed, a); n != 1 {
				t.Fatalf("%s: anchor %q occurs %d times, expected 1", name, a, n)
			}
			fixed = strings.Replace(fixed, a, r, 1)
		}
		cases := []struct {
			label string
			brief interface{}
			want  string
			delta int
		}{
			{"brief WITH text (gamedesign shape)", map[string]interface{}{"text": "MISSION-PROSE-SENTINEL", "proposition": "p"}, "MISSION-PROSE-SENTINEL", 0},
			{"brief OBJECT without text (brief-writer shape)", map[string]interface{}{"proposition": "OBJECT-PROP-SENTINEL", "content_plan": []interface{}{map[string]interface{}{"name": "Get Copy Written"}}}, "OBJECT-PROP-SENTINEL", -1},
			{"NO brief", nil, "", 0},
		}
		for _, c := range cases {
			outO, repO := render(t, orig, ctx(c.brief))
			outF, repF := render(t, fixed, ctx(c.brief))
			delta := repF.Occurrences - repO.Occurrences
			if delta != c.delta {
				t.Errorf("%s / %s: ScanMissingValues occurrences orig=%d fixed=%d delta=%d, want %d", name, c.label, repO.Occurrences, repF.Occurrences, delta, c.delta)
			}
			if c.want != "" && !strings.Contains(outF, c.want) {
				t.Errorf("%s / %s: fixed render lacks sentinel %q", name, c.label, c.want)
			}
			if c.delta == -1 {
				if strings.Contains(outO, c.want) {
					t.Errorf("%s / %s: ORIGINAL render already shows the object — control failed to reproduce the defect", name, c.label)
				}
				// the fleet's own attribution must name the exact path 764 fixes
				named := false
				for _, f := range repO.Fields {
					if strings.Contains(f, "mission_brief.text") {
						named = true
					}
				}
				if !named {
					t.Errorf("%s / %s: original report did not attribute the hole to mission_brief.text (fields=%v)", name, c.label, repO.Fields)
				}
			}
			if c.brief == nil && (strings.Contains(outF, "## Mission") || strings.Contains(outF, "## Pre-Defined Mission")) {
				t.Errorf("%s / NO brief: a Mission block rendered with no brief", name)
			}
			// prose case must be byte-identical between orig and fixed
			if c.delta == 0 && c.want != "" && outO != outF {
				t.Errorf("%s / %s: prose brief render differs between original and fixed", name, c.label)
			}
			t.Logf("ok %s / %s: scan occurrences orig=%d fixed=%d (delta %d); orig attributed fields=%v", name, c.label, repO.Occurrences, repF.Occurrences, delta, repO.Fields)
		}
	}
}
