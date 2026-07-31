// render_tool.go — render a tool's html_template offline, the way production does.
//
// WHY GO AND NOT A PYTHON SUBSTITUTION. The platform renders html_template with
// Go's text/template, and a missing key renders as the literal string
// "<no value>" rather than erroring — which is exactly the stored-component
// corruption class TL-030 records (templates full of bare <no value>, zero
// {{.}} slots, empty input_schema). A Python `str.replace` prover would silently
// leave a slot unreplaced and look fine, so it would prove the wrong thing.
// Using the real engine means a typo'd field name is caught HERE, offline,
// before the component is stored and long before a page is built from it.
//
// It also asserts what production does not: that the rendered output contains no
// "<no value>", no residual "{{", and that every schema field was consumed.
//
// Usage:
//   go run render_tool.go <template> <schema.json> <out.html>
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type field struct {
	Type     string `json:"type"`
	Source   string `json:"source"`
	Required bool   `json:"required"`
	Fallback string `json:"fallback"`
}

type schema struct {
	Fields map[string]field `json:"fields"`
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: render_tool.go <template> <schema.json> <out.html>")
		os.Exit(2)
	}
	tmplPath, schemaPath, outPath := os.Args[1], os.Args[2], os.Args[3]

	raw, err := os.ReadFile(tmplPath)
	must(err)
	sraw, err := os.ReadFile(schemaPath)
	must(err)

	var sc schema
	must(json.Unmarshal(sraw, &sc))

	// Build the render data from each field's fallback. A `required` field with no
	// fallback is a defect in the schema, not something to paper over at render
	// time — a caller that forgets it would get "<no value>" in production.
	data := map[string]string{}
	var missing []string
	for name, f := range sc.Fields {
		if f.Fallback == "" && f.Required {
			missing = append(missing, name)
		}
		data[name] = f.Fallback
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "schema defect: required field(s) with no fallback: %s\n",
			strings.Join(missing, ", "))
		os.Exit(1)
	}

	// Option("missingkey=error") turns the silent "<no value>" into a hard failure.
	// Production does NOT do this, which is precisely why it must be done here:
	// this is the only place the mistake is cheap.
	tmpl, err := template.New("tool").Option("missingkey=error").Parse(string(raw))
	must(err)

	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		fmt.Fprintf(os.Stderr, "render failed: %v\n", err)
		os.Exit(1)
	}
	html := out.String()

	// Belt and braces: missingkey=error catches an absent key, but not a slot
	// written with the wrong syntax ({{ .foo }} vs {{.foo}} both work; {{foo}}
	// is a parse error; a stray literal "{{" survives).
	for _, bad := range []string{"<no value>", "{{"} {
		if strings.Contains(html, bad) {
			fmt.Fprintf(os.Stderr, "rendered output still contains %q — refusing to write\n", bad)
			os.Exit(1)
		}
	}

	// Every declared field should actually be used by the template; an unused one
	// is either dead schema or a renamed slot the template no longer reads.
	var unused []string
	for name := range sc.Fields {
		if !strings.Contains(string(raw), "{{."+name+"}}") &&
			!strings.Contains(string(raw), "{{ ."+name+" }}") {
			unused = append(unused, name)
		}
	}
	if len(unused) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: schema declares field(s) the template never reads: %s\n",
			strings.Join(unused, ", "))
	}

	must(os.WriteFile(outPath, []byte(html), 0o644))
	fmt.Printf("rendered %d fields -> %s (%d bytes)\n", len(sc.Fields), outPath, len(html))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
