// FILE: platform/orchestration/datahelpers/template_missing_values_test.go
//
// bugs_open/453 shape 3. The positive fixture is the REAL prompt text the
// bugs_open/437 lane quoted from a live page-content-writer call on 2026-09-03 —
// the contact block whose heading declares the data authoritative and whose
// fourth line was the hole — so the fix is proven by FIRING against the thing it
// was written for, not by passing over an invented one.
//
// The acquittals are the harder half and most of them are measured rather than
// imagined: 87 of the 139 live templates mention "exact" or "verified"
// somewhere, so a marker set including those, or a document-level rather than
// block-level test, would escalate almost every render. A severity that fires on
// two thirds of the fleet is not a severity, and these tests are what stop it
// becoming one.
package datahelpers

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// The verbatim shape from the 437 lane's contribution to bugs_open/453.
const liveContactBlockTemplate = `## Company Context
Company: {{.render_context.company_name}}

## Official Contact Information (USE ONLY THESE - DO NOT INVENT)
Email: {{.render_context.email}}
Phone: {{.render_context.phone}}
Location: {{.reviewed_brief.headquarters}}
`

func liveContactBlockData() map[string]interface{} {
	return map[string]interface{}{
		"render_context": map[string]interface{}{
			"company_name": "Finetune",
			"email":        "finetune@contactforsales.com",
			"phone":        "+44 (0) 7934 524 911",
		},
		// headquarters is absent — the data gap that manufactured the token.
		"reviewed_brief": map[string]interface{}{"services": "x"},
	}
}

func TestAuthoritativeStandInIsNamedStrippedAndEscalated(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	out, err := RenderPromptTemplate(liveContactBlockTemplate, liveContactBlockData(), *zap.New(core))
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	// 1. The model must not receive the token.
	if strings.Contains(out, MissingValueToken) {
		t.Errorf("the stand-in reached the prompt:\n%s", out)
	}
	// 2. ...and the surrounding truth must survive the strip.
	for _, keep := range []string{"finetune@contactforsales.com", "+44 (0) 7934 524 911", "Location:"} {
		if !strings.Contains(out, keep) {
			t.Errorf("stripping removed more than the token: %q is gone\n%s", keep, out)
		}
	}

	// 3. Escalated, because it landed inside the DO-NOT-INVENT block.
	errs := logs.FilterLevelExact(zap.ErrorLevel).All()
	if len(errs) != 1 {
		t.Fatalf("expected exactly one Error; got %d entries: %v", len(errs), logs.All())
	}
	ctx := errs[0].ContextMap()
	if got := ctx["in_authoritative_block"]; got != int64(1) {
		t.Errorf("in_authoritative_block = %v, want 1", got)
	}
	// 4. NAMED, not counted — the whole point of the change.
	paths := strings.Join(strings.Fields(strings.Trim(sprint(ctx["unresolved_paths"]), "[]")), ",")
	if !strings.Contains(paths, "reviewed_brief.headquarters") {
		t.Errorf("unresolved_paths = %v, want it to name reviewed_brief.headquarters", ctx["unresolved_paths"])
	}
	// 5. ...and it must NOT accuse the fields that were supplied.
	for _, present := range []string{"render_context.email", "render_context.phone", "render_context.company_name"} {
		if strings.Contains(paths, present) {
			t.Errorf("named %q as unresolved, but it was supplied", present)
		}
	}
}

// A hole with no directive near it is a Warn, not an Error. Without this the
// escalation is indistinguishable from "log everything at Error".
func TestOrdinaryHoleWarnsAndDoesNotEscalate(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	out, err := RenderPromptTemplate("Some prose.\n\nTagline: {{.brief.tagline}}\n",
		map[string]interface{}{"brief": map[string]interface{}{}}, *zap.New(core))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(out, MissingValueToken) {
		t.Error("token not stripped")
	}
	if n := logs.FilterLevelExact(zap.ErrorLevel).Len(); n != 0 {
		t.Errorf("escalated with no anti-invention directive present (%d Errors)", n)
	}
	if n := logs.FilterLevelExact(zap.WarnLevel).Len(); n != 1 {
		t.Errorf("want exactly one Warn, got %d", n)
	}
}

// THE PROXIMITY PROPERTY, and the reason the rule is block-scoped. A directive
// in a DIFFERENT paragraph must not escalate a hole elsewhere — 87 of 139 live
// templates carry such a directive somewhere, so a document-level test escalates
// nearly every render and the signal stops discriminating.
func TestDirectiveInAnotherBlockDoesNotEscalate(t *testing.T) {
	const tpl = `## Rules
Never invent a statistic.

## Some other section
Tagline: {{.brief.tagline}}
`
	core, logs := observer.New(zap.WarnLevel)
	if _, err := RenderPromptTemplate(tpl, map[string]interface{}{"brief": map[string]interface{}{}}, *zap.New(core)); err != nil {
		t.Fatalf("render: %v", err)
	}
	if n := logs.FilterLevelExact(zap.ErrorLevel).Len(); n != 0 {
		t.Errorf("a directive two paragraphs away escalated a hole (%d Errors) — the rule is not block-scoped", n)
	}
}

// The marker set is deliberately narrow. "exact" and "verified" are the two most
// common words in this corpus (161 and 73 occurrences across 139 templates) and
// are instructions about FORM, not provenance.
func TestGenericEmphasisWordsDoNotEscalate(t *testing.T) {
	for _, marker := range []string{
		"Use the exact figures below.",
		"These are the verified facts.",
		"Preserve every fact, number and name exactly.",
	} {
		core, logs := observer.New(zap.WarnLevel)
		tpl := marker + "\nLocation: {{.brief.hq}}\n"
		if _, err := RenderPromptTemplate(tpl, map[string]interface{}{"brief": map[string]interface{}{}}, *zap.New(core)); err != nil {
			t.Fatalf("render: %v", err)
		}
		if n := logs.FilterLevelExact(zap.ErrorLevel).Len(); n != 0 {
			t.Errorf("%q escalated; it is an instruction about form, not provenance", marker)
		}
	}
}

// A clean render must be untouched and silent, or every prompt in the fleet
// gains a log line and the channel is worthless.
func TestCleanRenderIsUntouchedAndSilent(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	const tpl = "Company: {{.brief.name}}\nDO NOT INVENT anything.\n"
	out, err := RenderPromptTemplate(tpl, map[string]interface{}{"brief": map[string]interface{}{"name": "Acme"}}, *zap.New(core))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if out != "Company: Acme\nDO NOT INVENT anything.\n" {
		t.Errorf("clean render altered: %q", out)
	}
	if logs.Len() != 0 {
		t.Errorf("clean render logged %d entries: %v", logs.Len(), logs.All())
	}
}

// The honest gap, asserted so nobody later reads an empty field list as "no
// fields affected": inside a {{range}} the dot is a loop item this scan cannot
// see, so the occurrence is COUNTED but not attributable.
func TestRangeBodyHoleIsCountedButNotAttributed(t *testing.T) {
	const tpl = `{{range .items}}- {{.missing_leaf}}
{{end}}`
	data := map[string]interface{}{"items": []interface{}{map[string]interface{}{"other": 1}}}

	rep := ScanMissingValues(tpl, mustRender(t, tpl, data), data)
	if rep.Occurrences != 1 {
		t.Fatalf("occurrences = %d, want 1 — the count reads the OUTPUT and must be exact", rep.Occurrences)
	}
	if len(rep.Fields) != 0 {
		t.Errorf("attributed %v from inside a range body, where the dot is a loop item", rep.Fields)
	}

	core, logs := observer.New(zap.WarnLevel)
	if _, err := RenderPromptTemplate(tpl, data, *zap.New(core)); err != nil {
		t.Fatal(err)
	}
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("want one entry, got %d", len(entries))
	}
	if got := sprint(entries[0].ContextMap()["unresolved_paths"]); !strings.Contains(got, "range/with") {
		t.Errorf("unresolved_paths = %q; an empty list must SAY it is empty because the scan could not see, "+
			"not read as 'no fields affected'", got)
	}
}

func TestPathResolvesMirrorsTemplateIndexing(t *testing.T) {
	data := map[string]interface{}{
		"a":       map[string]interface{}{"b": map[string]interface{}{"c": "x"}},
		"nilkey":  nil,
		"emptied": map[string]interface{}{"leaf": nil},
	}
	cases := []struct {
		path string
		want bool
	}{
		{"a.b.c", true},
		{"a.b", true},
		{"a.missing", false},
		{"a.b.missing", false},
		{"nilkey", false},       // present but nil IS what renders the token
		{"emptied.leaf", false}, // ditto, one level down
		{"absent.entirely", false},
	}
	for _, c := range cases {
		if got := pathResolves(data, strings.Split(c.path, ".")); got != c.want {
			t.Errorf("pathResolves(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestStripMissingValuesReportsWhatItRemoved(t *testing.T) {
	out, n := StripMissingValues("a <no value> b <no value>")
	if out != "a  b " || n != 2 {
		t.Errorf("got %q / %d, want %q / 2", out, n, "a  b ")
	}
	if out, n := StripMissingValues("nothing here"); out != "nothing here" || n != 0 {
		t.Errorf("clean input altered: %q / %d", out, n)
	}
}

func mustRender(t *testing.T, tpl string, data map[string]interface{}) string {
	t.Helper()
	// Render WITHOUT the strip, to get the raw output ScanMissingValues reads.
	out, err := rawRenderForTest(tpl, data)
	if err != nil {
		t.Fatalf("raw render: %v", err)
	}
	return out
}

func sprint(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return strings.Trim(strings.Join(strings.Fields(strings.TrimSpace(sprintSlice(v))), " "), "")
}

func sprintSlice(v interface{}) string {
	switch t := v.(type) {
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, sprint(e))
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(t, ",")
	default:
		return ""
	}
}

// rawRenderForTest renders WITHOUT RenderPromptTemplate's strip, so a test can
// see the output ScanMissingValues is meant to read. Deliberately duplicates
// only the two lines that matter (parse with production's func map, execute) —
// anything more would be a second renderer, and a test asserting on a second
// renderer proves nothing about the first.
func rawRenderForTest(tpl string, data map[string]interface{}) (string, error) {
	t, err := template.New("raw").Funcs(PromptTemplateFuncs()).Parse(tpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
