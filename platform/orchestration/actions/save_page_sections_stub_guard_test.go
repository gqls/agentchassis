package actions

import "testing"

// Tests for the unresolvable-section guard (bugs_open/039): an empty
// generic-text-block stub must be recognised so SavePageSectionsAction refuses
// to persist it as a deployed, component_id=NULL row, while a generic block
// that received real content — and any resolved component — is left untouched.

// emptyGenericStub is the exact 208-byte markup generic-text-block renders when
// it stands in for an unresolvable section name and receives no content.
const emptyGenericStub = `<section id="8d81e665-3ee0-443d-a873-690268c15fbb" class="section section--generic">
  <div class="container">
    <h2 class="section__title"></h2>
    <div class="section__content"></div>
  </div>
</section>`

func TestIsEmptyGenericStub(t *testing.T) {
	cases := []struct {
		name string
		html string
		want bool
	}{
		{
			name: "empty generic stub — the bug",
			html: emptyGenericStub,
			want: true,
		},
		{
			name: "generic block WITH content — legitimate, must not match",
			html: `<section id="x" class="section section--generic">
  <div class="container">
    <h2 class="section__title">Our Philosophy</h2>
    <div class="section__content">We build systems that run in production.</div>
  </div>
</section>`,
			want: false,
		},
		{
			name: "generic block, heading only — has visible text, must not match",
			html: `<section class="section section--generic"><div class="container"><h2 class="section__title">Insights</h2><div class="section__content"></div></div></section>`,
			want: false,
		},
		{
			name: "resolved non-generic component, empty-ish — no generic marker, must not match",
			html: `<section class="section section--hero"><div class="container"><h1></h1></div></section>`,
			want: false,
		},
		{
			name: "resolved non-generic component with content — must not match",
			html: `<section class="section section--testimonials"><blockquote>It just works.</blockquote></section>`,
			want: false,
		},
		{
			name: "empty string",
			html: "",
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isEmptyGenericStub(tc.html); got != tc.want {
				t.Errorf("isEmptyGenericStub() = %v, want %v", got, tc.want)
			}
		})
	}
}
