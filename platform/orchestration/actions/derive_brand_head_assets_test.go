package actions

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func TestParseHexColour(t *testing.T) {
	cases := []struct {
		in   string
		want color.RGBA
	}{
		{"#0080FF", color.RGBA{0x00, 0x80, 0xFF, 0xff}},
		{"1a1a2e", color.RGBA{0x1a, 0x1a, 0x2e, 0xff}},
		{"#abc", color.RGBA{0xaa, 0xbb, 0xcc, 0xff}},
		// Gradients / junk fall back to the dark neutral — OG cards need a solid.
		{"linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)", color.RGBA{0x1a, 0x1a, 0x2e, 0xff}},
		{"", color.RGBA{0x1a, 0x1a, 0x2e, 0xff}},
	}
	for _, c := range cases {
		got := parseHexColour(c.in)
		if got != color.Color(c.want) {
			t.Errorf("parseHexColour(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestComposeFaviconPreservesAspect(t *testing.T) {
	// A wide wordmark: 200×100 solid red. The old resize.Resize(64,64)
	// stretched it to fill the square; composeFavicon must instead fit it
	// (64×32), centred, with the vertical padding transparent.
	wide := image.NewRGBA(image.Rect(0, 0, 200, 100))
	draw.Draw(wide, wide.Bounds(), &image.Uniform{C: color.RGBA{0xff, 0, 0, 0xff}}, image.Point{}, draw.Src)

	got := composeFavicon(wide)
	if b := got.Bounds(); b.Dx() != faviconSize || b.Dy() != faviconSize {
		t.Fatalf("favicon canvas is %dx%d, want %dx%d", b.Dx(), b.Dy(), faviconSize, faviconSize)
	}

	// Padding rows (top and bottom) transparent, centre opaque.
	if _, _, _, a := got.At(32, 2).RGBA(); a != 0 {
		t.Errorf("top padding not transparent (alpha=%d) — logo was stretched to fill", a)
	}
	if _, _, _, a := got.At(32, 61).RGBA(); a != 0 {
		t.Errorf("bottom padding not transparent (alpha=%d) — logo was stretched to fill", a)
	}
	if _, _, _, a := got.At(32, 32).RGBA(); a == 0 {
		t.Errorf("centre is transparent — logo missing from canvas")
	}

	// A square logo still fills the box edge to edge.
	square := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(square, square.Bounds(), &image.Uniform{C: color.RGBA{0, 0xff, 0, 0xff}}, image.Point{}, draw.Src)
	if _, _, _, a := composeFavicon(square).At(32, 2).RGBA(); a == 0 {
		t.Errorf("square logo should fill the box; top row is transparent")
	}
}

// The lock guard is the safety half of this action: a locked row is an owner
// approval and the derivation must skip that artefact BEFORE the git commit.
// Deliberately no status filter — assets.status is unconstrained text, so a
// locked row must fail closed whatever status it carries.
//
// The predicate itself moved to asset_lock_guard.go (bugs_open/143), so the SQL
// asserted here is the shared query's, and the no-status-filter / no-expiry
// guarantees are pinned directly in asset_lock_guard_test.go. What THIS test
// still owns is that brand-head asks about both of its own artefacts and reads
// the answer per key.
func TestLockedBrandHeadKeys(t *testing.T) {
	ctx := context.Background()
	siteID := uuid.New()

	cases := []struct {
		name string
		rows []string
		want map[string]bool
	}{
		{"no locked rows", nil, map[string]bool{}},
		{"og_card locked", []string{"og_card"}, map[string]bool{"og_card": true}},
		{"both locked (any status — filter must not exist)", []string{"og_card", "favicon"},
			map[string]bool{"og_card": true, "favicon": true}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			r := sqlmock.NewRows([]string{"asset_key", "locked_by", "lock_type", "locked_at"})
			for _, k := range c.rows {
				r.AddRow(k, "admin", "permanent", time.Now())
			}
			mock.ExpectQuery(`SELECT DISTINCT ON \(asset_key\)`).WillReturnRows(r)

			got, err := lockedBrandHeadKeys(ctx, db, siteID)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got.Keys(), c.want)
			}
			for k := range c.want {
				if !got.Locked(k) {
					t.Errorf("missing locked key %q in %v", k, got.Keys())
				}
			}
		})
	}
}

// Both artefacts locked → the action must refuse with derived:false BEFORE
// touching storage or git. StorageClient is deliberately nil here: reaching
// the storage type-assertion would error, so a clean refusal is also proof
// of the ordering (lock check precedes any write machinery).
func TestDeriveBrandHeadBothLockedRefuses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("FROM assets a").WillReturnRows(
		sqlmock.NewRows([]string{"url", "storage_path", "domain"}).
			AddRow("https://s3.example.com/bucket/images/logo.png", "", "example.com"))
	mock.ExpectQuery("color_palette").WillReturnRows(
		sqlmock.NewRows([]string{"color_palette"}).AddRow(`{"background":"#ffffff"}`))
	mock.ExpectQuery(`SELECT DISTINCT ON \(asset_key\)`).WillReturnRows(
		sqlmock.NewRows([]string{"asset_key", "locked_by", "lock_type", "locked_at"}).
			AddRow("favicon", "admin", "permanent", time.Now()).
			AddRow("og_card", "admin", "permanent", time.Now()))

	out, err := DeriveBrandHeadAssetsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    map[string]interface{}{"site_id": siteID.String()},
	})
	if err != nil {
		t.Fatalf("expected clean refusal, got error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok || m["derived"] != false {
		t.Fatalf("expected derived:false refusal, got %#v", out)
	}
	if reason, _ := m["reason"].(string); !strings.Contains(reason, "locked") {
		t.Fatalf("refusal reason should name the lock, got %q", reason)
	}
}

func TestInjectBrandHeadTags(t *testing.T) {
	log := zap.NewNop()
	ctx := &RenderContext{Domain: "robot-hands.com", CompanyName: "Robot-Hands", Tagline: "Grip intelligence", LogoURL: "/assets/images/logo.jpg"}

	head := "<head>\n  <title>x</title>\n</head>"
	out, _ := injectBrandHeadTags(head, ctx, true, log)

	for _, want := range []string{
		`rel="icon" href="/assets/images/favicon.png"`,
		`rel="icon" href="/assets/images/logo.jpg"`,
		`property="og:image" content="https://robot-hands.com/assets/images/og-card.png"`,
		`property="og:title" content="Robot-Hands"`,
		`name="twitter:card" content="summary_large_image"`,
		`rel="stylesheet" href="/assets/css/sprites.css"`, // hasSpriteCSS=true
	} {
		if !strings.Contains(out, want) {
			t.Errorf("injected head missing %q\n---\n%s", want, out)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(out), "</head>") {
		t.Errorf("tags must be injected BEFORE </head>; got tail: %q", out[len(out)-40:])
	}

	// CONTRACT CHANGED 2026-08-21 (bugs_open/322 item 4, owner-directed).
	// This used to assert a WHOLESALE no-op: a head containing rel="icon" OR
	// og:image was returned untouched. That is the defect, not the contract —
	// one foreign tag disabled the entire block, so webdesign.co.uk's
	// hand-authored favicon cost it every og and twitter tag on 117 pages,
	// and the guard could not see a BLANK og:title, which is how four sites
	// ended up serving duplicates (bugs_closed/252).
	//
	// The idempotence that MATTERS is per tag, and it is asserted below: an
	// authored tag is preserved byte-for-byte, and only the missing ones are
	// added. Changing this assertion is a deliberate contract change with the
	// reason on record — not a checker bent to agree with the code.
	already := `<head><link rel="icon" href="/x.png"></head>`
	got, _ := injectBrandHeadTags(already, ctx, false, log)
	if !strings.Contains(got, `rel="icon" href="/x.png"`) {
		t.Errorf("the authored favicon must be preserved exactly, got: %s", got)
	}
	if strings.Contains(got, `href="/assets/images/favicon.png"`) &&
		strings.Count(got, `rel="icon"`) != 1 {
		t.Errorf("must not add a second rel=\"icon\" beside an authored one, got: %s", got)
	}
	for _, want := range []string{
		`property="og:type"`, `property="og:site_name"`, `property="og:image"`,
		`name="twitter:card"`, `rel="apple-touch-icon"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("a partial head must still receive %q — the wholesale skip is the bug: %s", want, got)
		}
	}

	// Per-tag idempotence, both quote styles: a head that already declares a
	// tag keeps ITS version and gains no duplicate.
	authored := `<head><meta property='og:image' content='/mine.png'><meta name="twitter:card" content="summary"><link rel="icon" href="/i.png"><link rel="apple-touch-icon" href="/a.png"></head>`
	out3, _ := injectBrandHeadTags(authored, ctx, false, log)
	for _, tag := range []string{"og:image", "twitter:card", "rel=\"icon\"", "apple-touch-icon"} {
		if strings.Count(out3, tag) != 1 {
			t.Errorf("tag %q duplicated or lost (count %d): %s", tag, strings.Count(out3, tag), out3)
		}
	}
	if !strings.Contains(out3, `content='/mine.png'`) {
		t.Errorf("single-quoted authored og:image must be preserved: %s", out3)
	}

	// A head that already declares EVERYTHING comes back byte-identical — the
	// steady state after one render, and it must not churn the stored artefact
	// (the site_components archive trigger fires on a real change).
	full, _ := injectBrandHeadTags("<head><title>x</title></head>", ctx, true, log)
	if again, _ := injectBrandHeadTags(full, ctx, true, log); again != full {
		t.Errorf("second pass must be byte-identical:\n--- once ---\n%s\n--- twice ---\n%s", full, again)
	}

	// No </head> → returned unchanged, AND a non-empty decline reason so the
	// caller can record it durably. The reason is the point: council round 2
	// (bug_historian) established that a zap.Warn is not a fail-loud signal
	// here, because chassis log retention is measured in minutes. A silent
	// return is how webdesign.co.uk lost every brand tag on 117 pages.
	got2, declined := injectBrandHeadTags("no head here", ctx, false, log)
	if got2 != "no head here" {
		t.Errorf("expected unchanged on malformed head, got: %s", got2)
	}
	if declined == "" {
		t.Error("a decline must return a NON-EMPTY reason — the caller writes it to agent_error_log, and an empty string makes the skip silent again")
	}

	// The normal paths must NOT report a decline, or the caller writes a row
	// on every healthy render.
	if _, d := injectBrandHeadTags("<head><title>x</title></head>", ctx, true, log); d != "" {
		t.Errorf("healthy injection reported a decline: %q", d)
	}
	if _, d := injectBrandHeadTags(full, ctx, true, log); d != "" {
		t.Errorf("byte-identical no-op reported a decline: %q", d)
	}

	// Escaping: a company name with quotes/ampersand must not break the attr.
	ctx2 := &RenderContext{Domain: "x.com", CompanyName: `A&B "Co"`, Tagline: ""}
	out2, _ := injectBrandHeadTags("<head></head>", ctx2, false, log)
	if !strings.Contains(out2, `content="A&amp;B &quot;Co&quot;"`) {
		t.Errorf("attr not escaped: %s", out2)
	}
}
