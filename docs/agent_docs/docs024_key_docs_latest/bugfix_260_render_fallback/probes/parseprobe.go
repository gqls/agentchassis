// parseprobe — how many live component templates FAIL text/template Parse today?
//
// Parse is data-independent, so this needs no RenderContext replica. The only
// thing a replica could get wrong is the FuncMap NAME SET (an undefined function
// is a PARSE error, so a missing name would manufacture failures). The names
// below are pasted from the grep of executeGoTemplate's FuncMap in
// call_agent.go, and the program asserts the count it was built against.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
)

type comp struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Function string `json:"function"`
	IsActive bool   `json:"is_active"`
	Template string `json:"html_template"`
}

func main() {
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var comps []comp
	if err := json.Unmarshal(raw, &comps); err != nil {
		panic(err)
	}

	// Derived from call_agent.go:1172-1207 (7 names, asserted by the caller's grep).
	fm := template.FuncMap{
		"default": func(a, b interface{}) interface{} { return a },
		"eq":      func(a, b interface{}) bool { return false },
		"ne":      func(a, b interface{}) bool { return false },
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"isset":   func(v interface{}) bool { return false },
		"safe":    func(v interface{}) string { return "" },
	}
	if len(fm) != 7 {
		panic("FuncMap name count drifted from the 7 grepped out of call_agent.go")
	}

	// POSITIVE CONTROL: a template that MUST fail, and one that MUST pass.
	if _, e := template.New("c").Option("missingkey=zero").Funcs(fm).Parse(`{{if .a}}x`); e == nil {
		panic("control failed: an unclosed {{if}} parsed clean — the probe cannot detect anything")
	}
	if _, e := template.New("c").Option("missingkey=zero").Funcs(fm).Parse(`{{if .a}}{{range .b}}{{.c}}{{end}}{{end}}`); e != nil {
		panic("control failed: a valid template did not parse: " + e.Error())
	}

	var failedActive, failedInactive []string
	nActive := 0
	for _, c := range comps {
		if c.IsActive {
			nActive++
		}
		_, err := template.New("component").Option("missingkey=zero").Funcs(fm).Parse(c.Template)
		if err != nil {
			msg := err.Error()
			if len(msg) > 130 {
				msg = msg[:130]
			}
			line := fmt.Sprintf("%-34s %-22s active=%v  %s", c.Name, c.Function, c.IsActive, msg)
			if c.IsActive {
				failedActive = append(failedActive, line)
			} else {
				failedInactive = append(failedInactive, line)
			}
		}
	}
	sort.Strings(failedActive)
	sort.Strings(failedInactive)

	fmt.Printf("components: %d total, %d active\n", len(comps), nActive)
	fmt.Printf("PARSE FAILURES: %d active, %d inactive\n\n", len(failedActive), len(failedInactive))
	for _, l := range failedActive {
		fmt.Println("ACTIVE   ", l)
	}
	for _, l := range failedInactive {
		fmt.Println("inactive ", l)
	}
}
