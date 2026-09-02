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

// Candidate C, as a Go template. Sibling exclusion is by SUBJECT, because section
// NAMES repeat (generic-text-block x3 on the real playground row).
const block = `{{if .current_section.subject}}## This section

You'll want to know {{.current_section.subject}}. That's what this section is for.

{{.current_page.title}} also covers, each in its own section:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
{{end}}{{end}}
{{end}}## Verified Facts (marker: next block)`

// Same construction as datahelpers.RenderPromptTemplate: Funcs + Parse + Execute,
// default options (so a missing map key prints "<no value>").
func render(data map[string]interface{}) (string, error) {
	fm := template.FuncMap{"toJSON": func(v interface{}) string { b, _ := json.Marshal(v); return string(b) },
		"placeholder": func(s string) string { return s }, "rangeStart": func() string { return "" }, "rangeEnd": func() string { return "" }}
	t, err := template.New("agent_prompt").Funcs(fm).Parse(block)
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
	names := make([]string, 0, len(fx))
	for k := range fx {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		out, err := render(fx[n])
		fmt.Printf("===== %s =====\n", n)
		if err != nil {
			fmt.Printf("!! ERROR: %v\n", err)
		}
		fmt.Println(out)
		if strings.Contains(out, "<no value>") {
			fmt.Println("!! CONTAINS <no value>")
		}
		fmt.Println()
	}
}
