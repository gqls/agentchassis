package discovery_checks

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
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

// TestVerifyMissingComponentIsNotSuccess pins bugs_open/032. The verifier used to
// report Resolved:true when the page_components row was gone, on the assumption
// that a removed component cannot be empty. But a missing row is equally the
// signature of a rebuild silently deleting the component — so that branch
// recorded content-loss incidents as verified fixes, which is precisely the
// blind spot the completion gate exists to close.
//
// The contract now: never claim success from absence. Return an error, so the
// gate's fail-OPEN policy completes the item while recording under
// result._verification that verification could not be made. A false success
// becomes a visible unknown.
func TestVerifyMissingComponentIsNotSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	const componentID = "6f1b8a4e-6c2f-4a1e-9d3b-2f5c7a8e1b04"
	mock.ExpectQuery("SELECT rendered_html FROM page_components").
		WithArgs(componentID).
		WillReturnError(sql.ErrNoRows)

	res, err := VerifyEmptySectionResolved(
		context.Background(), db,
		VerifyTarget{Spec: map[string]interface{}{"component_id": componentID}},
		zap.NewNop(),
	)

	if err == nil {
		t.Fatal("a missing component must NOT verify as resolved: absence is equally deletion (bugs_open/032)")
	}
	if res.Resolved {
		t.Error("Resolved must be false when the component row is gone")
	}
	// The message has to say why it is unverifiable, or the recorded
	// _verification entry is as uninformative as the silent success it replaced.
	for _, want := range []string{"cannot verify", componentID, "indistinguishable"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must contain %q", err.Error(), want)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A present-but-empty component must still fail closed — the ordinary path must
// not regress while fixing the absent one.
func TestVerifyPresentButEmptyStillFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	const componentID = "6f1b8a4e-6c2f-4a1e-9d3b-2f5c7a8e1b04"
	mock.ExpectQuery("SELECT rendered_html FROM page_components").
		WithArgs(componentID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(""))

	res, err := VerifyEmptySectionResolved(
		context.Background(), db,
		VerifyTarget{Spec: map[string]interface{}{"component_id": componentID}},
		zap.NewNop(),
	)
	if err != nil {
		t.Fatalf("a present component must verify without error: %v", err)
	}
	if res.Resolved {
		t.Error("an empty rendered_html must not verify as resolved")
	}
}
