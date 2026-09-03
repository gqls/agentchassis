package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func toJSON(v interface{}) string {
	if v == nil { return "null" }
	if s, ok := v.(string); ok { return s }
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil { return fmt.Sprintf("%v", v) }
	return string(b)
}

var repl = map[string]string{
	"{{.site_specs.specs.mission_brief.text}}": "{{if .site_specs.specs.mission_brief.text}}{{.site_specs.specs.mission_brief.text}}{{else}}{{toJSON .site_specs.specs.mission_brief}}{{end}}",
	"{{.site_specs.specs.roadmap_brief.text}}": "{{if .site_specs.specs.roadmap_brief.text}}{{.site_specs.specs.roadmap_brief.text}}{{else}}{{toJSON .site_specs.specs.roadmap_brief}}{{end}}",
}

func ctx(brief interface{}) map[string]interface{} {
	specs := map[string]interface{}{"briefing": "b", "classification": map[string]interface{}{"category": "hub"}, "identity": map[string]interface{}{"name": "x"}, "strategy": map[string]interface{}{"site_type": "authority-portal"}}
	if brief != nil { specs["mission_brief"] = brief }
	return map[string]interface{}{
		"site_specs": map[string]interface{}{"specs": specs},
		"input_data": map[string]interface{}{"domain": "copyonline.co.uk"},
		"site_record": map[string]interface{}{"domain": "copyonline.co.uk"},
	}
}

func main() {
	fails := 0
	for _, name := range []string{"build-site-planner", "domain-research-classifier"} {
		raw, err := os.ReadFile(name + ".tpl"); if err != nil { panic(err) }
		tpl := string(raw)
		for a, r := range repl {
			if n := strings.Count(tpl, a); n != 1 { fmt.Printf("FAIL %s: anchor %q occurs %d times\n", name, a, n); fails++ }
			tpl = strings.Replace(tpl, a, r, 1)
		}
		// funcs the real renderer exposes that these templates might touch; unknown funcs fail at Parse, which is the point
		fm := template.FuncMap{"toJSON": toJSON, "placeholder": func(...interface{}) string { return "" }}
		t, err := template.New(name).Funcs(fm).Parse(tpl)
		if err != nil { fmt.Printf("FAIL %s: PARSE: %v\n", name, err); fails++; continue }
		cases := []struct{ label string; brief interface{}; want, wantNot string }{
			{"brief WITH text (gamedesign shape)", map[string]interface{}{"text": "MISSION-PROSE-SENTINEL", "proposition": "p"}, "MISSION-PROSE-SENTINEL", "<no value>"},
			{"brief OBJECT without text (brief-writer shape)", map[string]interface{}{"proposition": "OBJECT-PROP-SENTINEL", "content_plan": []interface{}{map[string]interface{}{"name": "Get Copy Written"}}}, "OBJECT-PROP-SENTINEL", "<no value>"},
			{"NO brief", nil, "", "OBJECT-PROP-SENTINEL"},
		}
		for _, c := range cases {
			var buf, buf0 bytes.Buffer
			if err := t.Execute(&buf, ctx(c.brief)); err != nil { fmt.Printf("FAIL %s / %s: EXEC: %v\n", name, c.label, err); fails++; continue }
			out := buf.String()
			t0, _ := template.New(name+"-o").Funcs(fm).Parse(string(raw))
			_ = t0.Execute(&buf0, ctx(c.brief))
			out0 := buf0.String()
			// isolate the mission block: from its heading to the next "## "
			block := func(o string) string {
				i := strings.Index(o, "## Mission"); if i < 0 { i = strings.Index(o, "## Pre-Defined Mission") }
				if i < 0 { return "" }
				j := strings.Index(o[i+3:], "\n## "); if j < 0 { return o[i:] }
				return o[i : i+3+j]
			}
			mb, mb0 := block(out), block(out0)
			novFixed, novOrig := strings.Count(mb, "<no value>"), strings.Count(mb0, "<no value>")
			totalDelta := strings.Count(out, "<no value>") - strings.Count(out0, "<no value>")
			ok := true
			switch {
			case c.brief == nil:
				ok = mb == "" && mb0 == "" && totalDelta == 0
			case c.want == "MISSION-PROSE-SENTINEL":
				ok = strings.Contains(mb, c.want) && strings.Contains(mb0, c.want) && novFixed == novOrig && totalDelta == 0 // prose brief: byte-for-byte same behaviour
			default:
				ok = strings.Contains(mb, c.want) && !strings.Contains(mb0, c.want) && novFixed == novOrig-1 && totalDelta == -1
				// name every residual <no value> in the FIXED block so a reader can see they are harness-context artefacts, not the brief
				for _, seg := range strings.SplitAfter(mb, "<no value>") {
					if i := strings.LastIndex(seg, "{{"); false && i >= 0 { _ = i }
				}
				idx := 0
				for k := 0; k < novFixed; k++ {
					j := strings.Index(mb[idx:], "<no value>"); if j < 0 { break }
					start := idx + j - 70; if start < 0 { start = 0 }
					fmt.Printf("     residual empty near: %q\n", strings.ReplaceAll(mb[start:idx+j], "\n", " "))
					idx = idx + j + len("<no value>")
				}
			}
			status := "ok  "; if !ok { status = "FAIL"; fails++ }
			fmt.Printf("%s %s / %s: mission-block <no value> orig=%d fixed=%d, whole-render delta=%d, sentinel in fixed block=%v\n", status, name, c.label, novOrig, novFixed, totalDelta, c.want != "" && strings.Contains(mb, c.want))
		}
		// control: the UNMODIFIED template must still show the defect on the object case (so the harness can fail)
		t0, err := template.New(name+"-orig").Funcs(fm).Parse(string(raw)); if err != nil { fmt.Printf("FAIL %s: orig PARSE: %v\n", name, err); fails++; continue }
		var b0 bytes.Buffer
		_ = t0.Execute(&b0, ctx(map[string]interface{}{"proposition": "OBJECT-PROP-SENTINEL"}))
		if strings.Contains(b0.String(), "<no value>") && !strings.Contains(b0.String(), "OBJECT-PROP-SENTINEL") {
			fmt.Printf("ok   %s / CONTROL: unmodified template renders <no value> and hides the object (defect reproduced)\n", name)
		} else { fmt.Printf("FAIL %s / CONTROL: unmodified template did not reproduce the defect\n", name); fails++ }
	}
	if fails > 0 { fmt.Printf("\n%d FAILURE(S)\n", fails); os.Exit(1) }
	fmt.Println("\nALL PASS")
}
