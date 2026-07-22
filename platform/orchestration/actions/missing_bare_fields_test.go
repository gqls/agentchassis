package actions

import (
	"reflect"
	"testing"
)

// TestMissingBareFields_ScopeAware locks the behaviour the council flagged in
// round 3 of the idea.uk chrome submission (bugs_open/018): the detector must
// report ONLY ungated, root-scope bare output fields that render empty. A field
// nested inside {{range}} (per-item dot) or {{if}}/{{with}} (author-gated) must
// NOT be reported — reporting it was the false-positive Error noise that would
// have fired on ~30 active components fleet-wide.
func TestMissingBareFields_ScopeAware(t *testing.T) {
	cases := []struct {
		name        string
		tpl         string
		data        map[string]interface{}
		wantMissing []string
		wantURL     []string
	}{
		{
			name:        "root-scope missing field is reported",
			tpl:         `<h1>{{.Title}}</h1><p>{{.Sub}}</p>`,
			data:        map[string]interface{}{"Title": "Hi"},
			wantMissing: []string{"Sub"},
			wantURL:     nil,
		},
		{
			name:        "missing href field is a dead control",
			tpl:         `<a href="{{.CTAUrl}}">Go</a>`,
			data:        map[string]interface{}{},
			wantMissing: []string{"CTAUrl"},
			wantURL:     []string{"CTAUrl"},
		},
		{
			name:        "missing src field, single-quoted, is a dead control",
			tpl:         `<img src='{{.LogoSrc}}'>`,
			data:        map[string]interface{}{},
			wantMissing: []string{"LogoSrc"},
			wantURL:     []string{"LogoSrc"},
		},
		{
			name:        "filled field is not reported",
			tpl:         `<a href="{{.CTAUrl}}">Go</a>`,
			data:        map[string]interface{}{"CTAUrl": "/contact"},
			wantMissing: nil,
			wantURL:     nil,
		},
		{
			// THE round-3 objection (guardian): a bare {{.Url}} inside a range
			// refers to the per-item dot, not the top-level map. Must NOT report.
			name:        "range-scoped field is NOT reported",
			tpl:         `{{range .Items}}<a href="{{.Url}}">{{.Label}}</a>{{end}}`,
			data:        map[string]interface{}{"Items": []interface{}{}},
			wantMissing: nil,
			wantURL:     nil,
		},
		{
			// THE round-3 objection (bug_historian): a bare {{.LogoURL}} inside an
			// {{if}} gate is author-handled. Must NOT report.
			name:        "if-gated field is NOT reported",
			tpl:         `{{if .ShowLogo}}<img src="{{.LogoURL}}">{{end}}`,
			data:        map[string]interface{}{},
			wantMissing: nil,
			wantURL:     nil,
		},
		{
			name:        "with-scoped field is NOT reported",
			tpl:         `{{with .Meta}}<span>{{.Author}}</span>{{end}}`,
			data:        map[string]interface{}{},
			wantMissing: nil,
			wantURL:     nil,
		},
		{
			// Root-scope field AFTER a control block is still seen; the gated
			// field of the same name inside the block is not conflated with it.
			name:        "root field after an if-block is still reported",
			tpl:         `{{if .Show}}<img src="{{.Logo}}">{{end}}<a href="{{.CTAUrl}}">x</a>`,
			data:        map[string]interface{}{},
			wantMissing: []string{"CTAUrl"},
			wantURL:     []string{"CTAUrl"},
		},
		{
			// A piped field ({{.X | safe}}) is a deliberate author choice and
			// renders "" via safe, not "<no value>" — not a dead control.
			name:        "piped field is not reported",
			tpl:         `<span>{{.Name | safe}}</span>`,
			data:        map[string]interface{}{},
			wantMissing: nil,
			wantURL:     nil,
		},
		{
			// Nested access {{.Foo.Bar}}: top-level presence says nothing about
			// the leaf, so it is not reported (matches the old single-segment rule).
			name:        "nested access is not reported",
			tpl:         `<span>{{.Foo.Bar}}</span>`,
			data:        map[string]interface{}{},
			wantMissing: nil,
			wantURL:     nil,
		},
		{
			// A template that will not parse as Go (mismatched delimiters) must
			// fall back to the flat regex scan rather than silently reporting none.
			name:        "unparseable template falls back to regex scan",
			tpl:         `<a href="{{.CTAUrl}}">{{range}}`,
			data:        map[string]interface{}{},
			wantMissing: []string{"CTAUrl"},
			wantURL:     []string{"CTAUrl"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			missing, inURL := missingBareFields(c.tpl, c.data)
			if !reflect.DeepEqual(missing, c.wantMissing) {
				t.Errorf("missing = %#v, want %#v", missing, c.wantMissing)
			}
			if !reflect.DeepEqual(inURL, c.wantURL) {
				t.Errorf("inURLAttr = %#v, want %#v", inURL, c.wantURL)
			}
		})
	}
}
