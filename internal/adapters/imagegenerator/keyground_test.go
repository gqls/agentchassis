// FILE: internal/adapters/imagegenerator/keyground_test.go
//
// bugs_open/424. Builds synthetic images in-test rather than reading fixture
// files, so every assertion is checkable against a value written two lines
// above it — the pattern provider_test.go already uses for the negative-
// prompt fold.
package imagegenerator

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

const (
	testInner = 48.0
	testOuter = 110.0
)

var (
	testMagenta = color.NRGBA{R: 255, G: 0, B: 255, A: 255}   // the key colour
	testGrey    = color.NRGBA{R: 128, G: 128, B: 128, A: 255} // stands in for "the mark"

	// testBlend is a TRUE alpha compositing blend of testGrey over testMagenta
	// at alpha ≈ 0.303 — chosen so that this package's own distance-to-alpha
	// mapping (alpha = (dist-inner)/(outer-inner) at testInner/testOuter
	// below) recovers approximately that same alpha, which is what makes the
	// despill test's "recovers close to grey" assertion meaningful rather
	// than arbitrary: (128,128,128)*0.303 + (255,0,255)*0.697 ≈ (217,39,217),
	// at distance ≈66 from magenta, inside (testInner, testOuter).
	testBlend = color.NRGBA{R: 217, G: 39, B: 217, A: 255}
)

// buildTestLogo makes a 32x32 image: a magenta background, a 16x16 grey
// square in the middle ("the mark"), a 4x4 EXACT-magenta hole enclosed inside
// the square (disconnected from the border — reachable only by the
// enclosed-hole pass), one graded BLEND pixel on the background side of the
// square's edge (reachable by the border flood-fill, since its neighbours
// are pure magenta chained back to the border), and one graded-distance
// BLEND pixel INSIDE the square, surrounded on all sides by opaque grey —
// unreachable by the border flood-fill and not close enough for the
// enclosed-hole pass, so it must come back untouched.
func buildTestLogo() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.SetNRGBA(x, y, testMagenta)
		}
	}
	// The "mark": a 16x16 grey square, x/y in [8,23].
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetNRGBA(x, y, testGrey)
		}
	}
	// Enclosed hole: 4x4 exact magenta, deep inside the square (x/y in [14,17]).
	for y := 14; y < 18; y++ {
		for x := 14; x < 18; x++ {
			img.SetNRGBA(x, y, testMagenta)
		}
	}
	// Border-connected graded edge pixel, background side of the square.
	img.SetNRGBA(7, 15, testBlend)
	// Disconnected graded-distance pixel, deep inside the opaque square,
	// surrounded on all four sides by grey (which sits at ~220 units from
	// magenta — far outside outer, so the flood-fill cannot cross it).
	img.SetNRGBA(10, 10, testBlend)
	return img
}

func TestKeyOutBackground_CornersFullyTransparent(t *testing.T) {
	img := buildTestLogo()
	out, _ := KeyOutBackground(img, testMagenta, testInner, testOuter)
	for _, p := range [][2]int{{0, 0}, {31, 0}, {0, 31}, {31, 31}} {
		if a := out.NRGBAAt(p[0], p[1]).A; a != 0 {
			t.Errorf("corner (%d,%d): want alpha 0 (border-connected magenta), got %d", p[0], p[1], a)
		}
	}
}

func TestKeyOutBackground_MarkStaysFullyOpaque(t *testing.T) {
	img := buildTestLogo()
	out, _ := KeyOutBackground(img, testMagenta, testInner, testOuter)
	// A plain grey pixel well away from the enclosed hole and the test blend pixels.
	c := out.NRGBAAt(20, 20)
	if c.A != 255 {
		t.Fatalf("mark pixel (20,20): want alpha 255, got %d", c.A)
	}
	if c.R != testGrey.R || c.G != testGrey.G || c.B != testGrey.B {
		t.Errorf("mark pixel (20,20): colour must be untouched, want %v got %v", testGrey, c)
	}
}

func TestKeyOutBackground_EnclosedHoleKeyed(t *testing.T) {
	img := buildTestLogo()
	out, _ := KeyOutBackground(img, testMagenta, testInner, testOuter)
	// Centre of the enclosed 4x4 hole.
	if a := out.NRGBAAt(15, 15).A; a != 0 {
		t.Errorf("enclosed hole (15,15): want alpha 0 (pass 2, exact-magenta anywhere), got %d", a)
	}
}

// The safety property the whole design rests on: a pixel that merely
// RESEMBLES the key colour, but is not connected to the image border through
// other near-key pixels, must never be erased by the border pass — otherwise
// a real mark using a similar shade could be punched full of holes.
func TestKeyOutBackground_DisconnectedInteriorPixelUntouched(t *testing.T) {
	img := buildTestLogo()
	out, _ := KeyOutBackground(img, testMagenta, testInner, testOuter)
	c := out.NRGBAAt(10, 10)
	if c.A != 255 {
		t.Fatalf("disconnected interior pixel (10,10): want alpha 255 (untouched — not border-connected, "+
			"not within inner), got %d", c.A)
	}
	if c.R != testBlend.R || c.G != testBlend.G || c.B != testBlend.B {
		t.Errorf("disconnected interior pixel (10,10): colour must be untouched, want %v got %v", testBlend, c)
	}
}

func TestKeyOutBackground_BorderConnectedEdgeIsGraded(t *testing.T) {
	img := buildTestLogo()
	out, _ := KeyOutBackground(img, testMagenta, testInner, testOuter)
	c := out.NRGBAAt(7, 15)
	if c.A == 0 || c.A == 255 {
		t.Fatalf("border-connected blend pixel (7,15): want a GRADED alpha strictly between 0 and 255, got %d", c.A)
	}
	// Despill: the recovered colour should read as roughly grey, not magenta —
	// otherwise a partially-keyed edge would carry a visible magenta fringe.
	if c.G < testGrey.G-20 {
		t.Errorf("border-connected blend pixel (7,15): despill left green channel too low (%d) — "+
			"looks like a magenta fringe, want close to grey (%d)", c.G, testGrey.G)
	}
}

func TestKeyOutBackground_BorderKeyedStatIsHigh(t *testing.T) {
	img := buildTestLogo()
	_, stats := KeyOutBackground(img, testMagenta, testInner, testOuter)
	if stats.BorderKeyed < 0.95 {
		t.Fatalf("a background that IS the key colour must clear the fail-closed guard threshold: "+
			"want BorderKeyed >= 0.95, got %.3f", stats.BorderKeyed)
	}
}

// TestKeyOutBackground_GradedBorderIsNotBorderKeyed — the real bug, found live
// (bugfix_417_420 lane, dynamic testing, CONTRIB round 3, 2026-09-02): BEFORE
// this fix, BorderKeyed was computed from BFS flood-fill REACHABILITY (was
// this pixel within `outer`?), not from the pixel's FINAL alpha. A border
// wash sitting entirely in the GRADED band — reachable, so every border pixel
// counted as "keyed" — but nowhere near `inner`, so every border pixel's
// ACTUAL alpha stayed well above 0, used to report BorderKeyed=1.000 despite
// the image being 0% transparent at its own edges. Measured live: a
// 0.0%-actually-transparent failure and an 87.4%-actually-transparent success
// both read 1.000 on the old code — the fail-closed guard could not tell them
// apart. This test builds exactly that shape: a uniform border colour placed
// deliberately mid-graded-band (distance ~79, strictly between inner=48 and
// outer=110), so every border pixel is flood-reachable but none is fully
// transparent.
func TestKeyOutBackground_GradedBorderIsNotBorderKeyed(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	// (210,60,210): distance to testMagenta (255,0,255) is
	// sqrt(45^2+60^2+45^2) ~= 87.5 — comfortably inside (inner=48, outer=110),
	// so every pixel is flood-reachable but none is close enough to inner to
	// ever reach alpha 0.
	gradedBand := color.NRGBA{R: 210, G: 60, B: 210, A: 255}
	d := colourDistance(gradedBand.R, gradedBand.G, gradedBand.B, testMagenta.R, testMagenta.G, testMagenta.B)
	if d <= testInner || d >= testOuter {
		t.Fatalf("test construction error: gradedBand distance %.1f is not strictly inside (inner=%.0f, outer=%.0f)",
			d, testInner, testOuter)
	}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetNRGBA(x, y, gradedBand)
		}
	}

	out, stats := KeyOutBackground(img, testMagenta, testInner, testOuter)

	if stats.BorderKeyed > 0.05 {
		t.Fatalf("graded-band border (distance %.1f, reachable but never fully transparent): "+
			"want BorderKeyed near 0, got %.3f — this is the exact live-measured bug (0%% actually "+
			"transparent read as BorderKeyed=1.000)", d, stats.BorderKeyed)
	}
	// Confirm the image genuinely is NOT transparent at its border — the
	// assertion above is only meaningful if this is true.
	if a := out.NRGBAAt(0, 0).A; a == 0 {
		t.Fatalf("test construction error: corner (0,0) is fully transparent (alpha 0), " +
			"so this is not actually testing the graded-not-transparent case")
	}
}

// Guard cases: the model ignoring the instruction entirely, in the two
// concrete shapes this bug's own history produced — a checkerboard (the
// failure this bug is named for) and a solid unrelated ground (the interim
// workaround this fix supersedes). Both must leave BorderKeyed near zero, so
// the fail-closed guard in dynamic_adapter.go refuses to upload either.
func TestKeyOutBackground_ChequerboardGivesNearZeroBorderKeyed(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	light := color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	dark := color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if (x+y)%2 == 0 {
				img.SetNRGBA(x, y, light)
			} else {
				img.SetNRGBA(x, y, dark)
			}
		}
	}
	_, stats := KeyOutBackground(img, testMagenta, testInner, testOuter)
	if stats.BorderKeyed > 0.05 {
		t.Fatalf("checkerboard ground: want BorderKeyed near 0 (model ignored the key-colour "+
			"instruction), got %.3f", stats.BorderKeyed)
	}
}

func TestKeyOutBackground_SolidUnrelatedGroundGivesNearZeroBorderKeyed(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	nearBlack := color.NRGBA{R: 10, G: 16, B: 16, A: 255} // the interim workaround's own ground colour
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetNRGBA(x, y, nearBlack)
		}
	}
	_, stats := KeyOutBackground(img, testMagenta, testInner, testOuter)
	if stats.BorderKeyed > 0.05 {
		t.Fatalf("solid near-black ground: want BorderKeyed near 0, got %.3f", stats.BorderKeyed)
	}
}

// TestKeyOutBackground_16BitSourceDepth — the live asset this bug is about is
// measured 16-bit PNG depth (RUNBOOK_logo_transparency.md), which Go's PNG
// decoder returns as *image.RGBA64, not the 8-bit *image.NRGBA every other
// test in this file uses. KeyOutBackground reads exclusively through
// image.Image/color.Color's generic RGBA() method (always 16-bit-normalised
// regardless of source depth — see the function's own doc comment), so this
// should already work; this test exercises that claim rather than trusting it.
func TestKeyOutBackground_16BitSourceDepth(t *testing.T) {
	// Build the same three-way shape (background / mark / enclosed hole) as
	// buildTestLogo, but as a genuine 16-bit RGBA64 image — every 8-bit
	// channel value scaled by 0x101, the same conversion Go's own PNG decoder
	// applies going the other way (16-bit sample = 8-bit sample * 0x101).
	scale16 := func(c color.NRGBA) color.RGBA64 {
		return color.RGBA64{
			R: uint16(c.R) * 0x101, G: uint16(c.G) * 0x101,
			B: uint16(c.B) * 0x101, A: uint16(c.A) * 0x101,
		}
	}
	img := image.NewRGBA64(image.Rect(0, 0, 32, 32))
	magenta16, grey16 := scale16(testMagenta), scale16(testGrey)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.SetRGBA64(x, y, magenta16)
		}
	}
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA64(x, y, grey16)
		}
	}
	for y := 14; y < 18; y++ {
		for x := 14; x < 18; x++ {
			img.SetRGBA64(x, y, magenta16)
		}
	}

	out, stats := KeyOutBackground(img, testMagenta, testInner, testOuter)

	if stats.BorderKeyed < 0.95 {
		t.Fatalf("16-bit source: want BorderKeyed >= 0.95, got %.3f", stats.BorderKeyed)
	}
	if a := out.NRGBAAt(0, 0).A; a != 0 {
		t.Errorf("16-bit source: corner (0,0) want alpha 0, got %d", a)
	}
	c := out.NRGBAAt(20, 20)
	if c.A != 255 || c.R != testGrey.R || c.G != testGrey.G || c.B != testGrey.B {
		t.Errorf("16-bit source: mark pixel (20,20) want untouched %v, got %v", testGrey, c)
	}
	if a := out.NRGBAAt(15, 15).A; a != 0 {
		t.Errorf("16-bit source: enclosed hole (15,15) want alpha 0, got %d", a)
	}
}

// The artefact-level check this bug was actually caught by: the served PNG's
// own colour type. A checkerboard-painted PNG (this bug's shipped defect) is
// colour type 2 (RGB, no alpha) — a matted result must be colour type 6.
func TestKeyOutBackground_EncodesRealAlphaChannel(t *testing.T) {
	img := buildTestLogo()
	out, _ := KeyOutBackground(img, testMagenta, testInner, testOuter)

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	b := buf.Bytes()
	// IHDR is the first chunk: 8-byte signature, 4-byte length, 4-byte type
	// "IHDR", then width(4) height(4) bitdepth(1) colourtype(1) ...
	const colourTypeOffset = 8 + 4 + 4 + 4 + 4 + 1
	if len(b) <= colourTypeOffset {
		t.Fatalf("encoded PNG too short to read IHDR colour type: %d bytes", len(b))
	}
	if ct := b[colourTypeOffset]; ct != 6 {
		t.Errorf("encoded PNG colour type: want 6 (RGBA), got %d — this is exactly the chunk-scan "+
			"check that caught the original checkerboard defect (colour type 2, no alpha)", ct)
	}
}
