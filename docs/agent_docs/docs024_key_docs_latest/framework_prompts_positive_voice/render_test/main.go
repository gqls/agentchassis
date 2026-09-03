package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
)

// Candidates for the 641 block under option A (one per-section field, the subject,
// authored in the voice and printed verbatim). Sibling exclusion/marking is by
// SUBJECT because section NAMES repeat. Order in `candidates` is fixed; the letters
// shown to the owner are assigned separately (NOTES holds the key).
var candidates = map[string]string{
	"C_control_as_committed_in_641": `{{if .current_section.subject}}## This section

You'll want to know {{.current_section.subject}}. That's what this section is for.

{{.current_page.title}} also covers, each in its own section:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
{{end}}{{end}}
{{end}}`,
	"A1_rest_of_the_page": `{{if .current_section.subject}}## This section

{{.current_section.subject}}

The rest of {{.current_page.title}}, section by section:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
{{end}}{{end}}
{{end}}`,
	"A2_page_in_order_this_one_marked": `{{if .current_section.subject}}## This section

{{.current_section.subject}}

{{.current_page.title}}, in order:
{{range $s := .sections_for_render.sections_ready}}{{if $s.subject}}- {{$s.subject}}{{if eq $s.subject $.current_section.subject}} (this section){{end}}
{{end}}{{end}}
{{end}}`,
	"A4_R_second_half_no_frame": `{{if .current_section.subject}}## This section

{{.current_section.subject}}

{{.current_page.title}} also covers, each in its own section:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
{{end}}{{end}}
{{end}}`,
	"A3_elsewhere_on_the_page": `{{if .current_section.subject}}## This section

{{.current_section.subject}}

Elsewhere on the page:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
{{end}}{{end}}
{{end}}`,
}

const tail = `## Verified Facts (marker: next block)`

// Same construction as datahelpers.RenderPromptTemplate: Funcs + Parse + Execute,
// default options (so a missing map key prints "<no value>").
func render(block string, data map[string]interface{}) (string, error) {
	fm := template.FuncMap{"toJSON": func(v interface{}) string { b, _ := json.Marshal(v); return string(b) },
		"placeholder": func(s string) string { return s }, "rangeStart": func() string { return "" }, "rangeEnd": func() string { return "" }}
	t, err := template.New("agent_prompt").Funcs(fm).Parse(block + tail)
	if err != nil {
		return "", fmt.Errorf("PARSE: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return buf.String(), fmt.Errorf("EXECUTE: %w", err)
	}
	return buf.String(), nil
}

func main() {
	raw, _ := os.ReadFile("fixtures.json")
	var fx map[string]map[string]interface{}
	json.Unmarshal(raw, &fx)
	fnames := make([]string, 0, len(fx))
	for k := range fx {
		fnames = append(fnames, k)
	}
	sort.Strings(fnames)
	cnames := make([]string, 0, len(candidates))
	for k := range candidates {
		cnames = append(cnames, k)
	}
	sort.Strings(cnames)
	for _, c := range cnames {
		fmt.Printf("########## CANDIDATE %s  (em dashes in template: %d)\n\n", c, strings.Count(candidates[c], "—"))
		for _, n := range fnames {
			out, err := render(candidates[c], fx[n])
			fmt.Printf("===== %s =====\n", n)
			if err != nil {
				fmt.Printf("!! ERROR: %v\n", err)
			}
			fmt.Println(out)
			if strings.Contains(out, "<no value>") {
				fmt.Println("!! CONTAINS <no value>")
			}
			if strings.Count(out, "(this section)") > 1 {
				fmt.Println("!! MARKER DOUBLED")
			}
			fmt.Println()
		}
	}
}
