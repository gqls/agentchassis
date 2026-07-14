package discovery_checks

import (
	"strings"
	"testing"
)

func TestEmptySectionVerdict(t *testing.T) {
	// The real defect this verifier was built for: gripper-detail's
	// product-details section — 8.7kB of e-commerce chrome with every value
	// empty, led by an empty <h1 class="pd-title"></h1>.
	hollowProductSection := `<section class="pd">` +
		`<h1 class="pd-title"></h1>` +
		`<span class="pd-price"></span>` +
		`<span class="pd-meta-val"></span>` +
		`<ul class="pd-features"><li></li><li></li><li></li><li></li></ul>` +
		`<button>Add to Cart</button><button>Buy Now</button>` +
		`</section>`

	cases := []struct {
		name         string
		html         string
		wantResolved bool
	}{
		{"empty string", "", false},
		{"whitespace only", "  \n\t ", false},
		{"minimal html", "<div></div>", false},
		{"empty heading shell", hollowProductSection, false},
		{"empty heading with attrs", `<section>` + strings.Repeat("x", 60) + `<h2 class="title"></h2></section>`, false},
		{"uppercase empty heading", `<section>` + strings.Repeat("x", 60) + `<H3></H3></section>`, false},
		{"runtime-fill shell exempt", `<div data-runtime-fill="lobby"><h2></h2></div>`, true},
		{"filled section", `<section><h1 class="pd-title">PG-90 Parallel Gripper</h1><span class="pd-price">£1,240</span></section>`, true},
		{"headless but substantial", `<section><p>` + strings.Repeat("real content ", 10) + `</p></section>`, true},
	}

	for _, tc := range cases {
		got := emptySectionVerdict(tc.html)
		if got.Resolved != tc.wantResolved {
			t.Errorf("%s: emptySectionVerdict resolved = %v (detail %q), want %v",
				tc.name, got.Resolved, got.Detail, tc.wantResolved)
		}
	}
}

func TestEmptyHeadingReMirrorsSQL(t *testing.T) {
	// The SQL pattern is '<(h[1-6])[^>]*>\s*</\1>'. RE2 has no backrefs, so
	// the Go mirror accepts mismatched heading levels — assert the deliberate
	// broadening so a future "fix" doesn't silently diverge the other way.
	if !emptyHeadingRe.MatchString(`<h2 class="a">   </h2>`) {
		t.Error("expected match on empty h2 with whitespace")
	}
	if !emptyHeadingRe.MatchString(`<h1></h3>`) {
		t.Error("expected match on mismatched empty heading pair (documented broadening)")
	}
	if emptyHeadingRe.MatchString(`<h2>Real title</h2>`) {
		t.Error("must not match a filled heading")
	}
}
