package datahelpers

import (
	"reflect"
	"testing"
)

func TestTemplateTopLevelFields(t *testing.T) {
	cases := []struct {
		name, tpl string
		want      []string
	}{
		{"bare", `<p>{{.body}}</p>`, []string{"body"}},
		{"trim and space", `{{- .title }} {{ .body }}`, []string{"title", "body"}},
		{"nested counts top", `{{.meta.author}}`, []string{"meta"}},
		{"range with if", `{{range .items}}{{.title}}{{end}}{{with .x}}{{end}}{{if .flag}}{{end}}`,
			[]string{"items", "title", "x", "flag"}},
		{"static", `<div class="a">no fields</div>`, nil},
		{"function args not matched", `{{if eq .a .b}}{{end}}`, nil},
		{"dedup", `{{.body}}{{.body}}`, []string{"body"}},
	}
	for _, c := range cases {
		if got := TemplateTopLevelFields(c.tpl); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestContentDataCanFillTemplate(t *testing.T) {
	portedPageTpl := `<section><div><article>{{.body}}</article></div></section>`
	// The worked case (bugs_open/277 §5.2): provenance-only content_data
	// against a body-only template — the population this test exists for.
	provenance := map[string]interface{}{
		"schema": "learn-page.v1", "sha256": "ab12", "source": "ported",
		"qa_tier": "1", "generator": "port",
	}
	if ContentDataCanFillTemplate(portedPageTpl, provenance) {
		t.Error("provenance-only content_data reported as able to fill a {{.body}} template")
	}
	if !ContentDataCanFillTemplate(portedPageTpl, map[string]interface{}{"body": "<p>prose</p>"}) {
		t.Error("content_data WITH body reported as unable to fill")
	}
	// Empty-string and nil values are absences, not fills.
	if ContentDataCanFillTemplate(portedPageTpl, map[string]interface{}{"body": ""}) {
		t.Error("empty-string body reported as a fill")
	}
	if ContentDataCanFillTemplate(portedPageTpl, map[string]interface{}{"body": nil}) {
		t.Error("nil body reported as a fill")
	}
	// A static template is trivially reproducible.
	if !ContentDataCanFillTemplate(`<hr>`, nil) {
		t.Error("static template reported as unfillable")
	}
	// Coarse-in-the-safe-direction: ANY present field counts as fillable.
	if !ContentDataCanFillTemplate(`{{.title}}{{.body}}`, map[string]interface{}{"title": "t"}) {
		t.Error("partially-fillable template must report fillable (keeps the regenerate route)")
	}
	// nil map reads as holding nothing.
	if ContentDataCanFillTemplate(portedPageTpl, nil) {
		t.Error("nil content_data reported as able to fill")
	}
}
