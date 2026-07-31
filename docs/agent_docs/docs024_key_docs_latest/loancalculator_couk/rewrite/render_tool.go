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

	// COPY MUST NOT BE INTERPOLATED INTO JAVASCRIPT. Caught the hard way on
	// tool-early-settlement, 2026-07-31: the schema fallback
	//     " in \"58-day\" interest charges."
	// was rendered into
	//     var BREAKDOWN_SUFFIX = " in "58-day" interest charges.";
	// which is a syntax error that kills the WHOLE script. The tool then showed
	// £0.00 for every input while still containing a <script> block, still
	// matching every selector, and still rendering perfectly. tool_health passes
	// it (check 4 only asks whether a script tag exists), Tier 2 passes it
	// (the anchors are all present), and only a check that reads the computed
	// values catches it.
	//
	// text/template does no escaping whatsoever — that is what it is for — so
	// there is no context-aware fix available here, and html/template is not what
	// production uses. The rule is therefore structural and worth keeping:
	// PUT COPY IN THE MARKUP AND LET JAVASCRIPT WRITE ONLY THE NUMBER. Text nodes
	// have no quoting hazard, and it has the better side effect that a content
	// agent can edit the wording without touching code.
	//
	// Conservative by design: it fires on any quote-bearing, backslash-bearing or
	// multi-line value interpolated inside a <script>, not on an attempt to guess
	// the surrounding quoting context. A false positive costs one markup move; a
	// false negative costs a silently dead calculator.
	var jsHazards []string
	for _, region := range scriptRegions(string(raw)) {
		for name, f := range sc.Fields {
			if !strings.Contains(region, "{{."+name+"}}") &&
				!strings.Contains(region, "{{ ."+name+" }}") {
				continue
			}
			if bad := unsafeInJS(f.Fallback); bad != "" {
				jsHazards = append(jsHazards, fmt.Sprintf(
					"%s (value contains %s)", name, bad))
			}
		}
	}
	if len(jsHazards) > 0 {
		fmt.Fprintf(os.Stderr,
			"refusing to write: field(s) interpolated INSIDE a <script> carry characters that break a JS string literal: %s\n"+
				"   Move the copy into the markup and have the script write only the computed value.\n",
			strings.Join(jsHazards, ", "))
		os.Exit(1)
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

// scriptRegions returns the contents of every <script> block in the template.
// Deliberately naive string scanning rather than a parser: the input is a
// component template, not arbitrary HTML, and a parser that silently normalised
// the markup would be reasoning about something other than what gets stored.
func scriptRegions(tmpl string) []string {
	var out []string
	lower := strings.ToLower(tmpl)
	for i := 0; ; {
		open := strings.Index(lower[i:], "<script")
		if open < 0 {
			return out
		}
		open += i
		gt := strings.Index(lower[open:], ">")
		if gt < 0 {
			return out
		}
		start := open + gt + 1
		end := strings.Index(lower[start:], "</script>")
		if end < 0 {
			out = append(out, tmpl[start:])
			return out
		}
		out = append(out, tmpl[start:start+end])
		i = start + end
	}
}

// unsafeInJS names the first character in v that cannot survive being pasted
// into a JavaScript string literal, or "" when the value is safe.
func unsafeInJS(v string) string {
	for _, c := range []struct {
		s, name string
	}{
		{`"`, `a double quote`},
		{`'`, `an apostrophe or single quote`},
		{`\`, `a backslash`},
		{"\n", `a newline`},
		{"`", `a backtick`},
	} {
		if strings.Contains(v, c.s) {
			return c.name
		}
	}
	return ""
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
