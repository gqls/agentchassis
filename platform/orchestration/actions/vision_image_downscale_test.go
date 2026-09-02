package actions

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"go.uber.org/zap"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// A non-uniform fill so scaling actually averages something.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

func decodeDims(t *testing.T, data []byte) (int, int) {
	t.Helper()
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return cfg.Width, cfg.Height
}

// An image at or under the cap passes through BYTE-IDENTICAL — the property
// that makes the default safe for every historically-working consumer.
func TestDownscalePassesThroughLegalImagesUntouched(t *testing.T) {
	in := pngBytes(t, 640, 480)
	out, mt, scaled := downscaleVisionImage(in, "image/png", 7900, zap.NewNop())
	if scaled {
		t.Fatalf("640x480 must not scale under a 7900 cap")
	}
	if mt != "image/png" {
		t.Fatalf("media type changed on passthrough: %s", mt)
	}
	if !bytes.Equal(in, out) {
		t.Fatalf("passthrough bytes differ")
	}
}

// A too-tall capture (the leopardess failure shape: viewport-wide, very tall)
// comes back with its long edge exactly at the cap, aspect preserved, JPEG.
func TestDownscaleCapsTheTallCapture(t *testing.T) {
	in := pngBytes(t, 300, 9000) // >8000 tall, like a post-hero full-page capture
	out, mt, scaled := downscaleVisionImage(in, "image/png", 7900, zap.NewNop())
	if !scaled {
		t.Fatalf("9000px-tall image must scale under a 7900 cap")
	}
	if mt != "image/jpeg" {
		t.Fatalf("scaled image must re-encode as jpeg, got %s", mt)
	}
	w, h := decodeDims(t, out)
	if h != 7900 {
		t.Fatalf("long edge must be exactly the cap: got h=%d", h)
	}
	if w != 300*7900/9000 {
		t.Fatalf("aspect not preserved: got w=%d want %d", w, 300*7900/9000)
	}
}

// Width-long images cap on width — the guard is on the LONG edge, not height.
func TestDownscaleCapsTheWideImageToo(t *testing.T) {
	in := pngBytes(t, 9000, 200)
	out, _, scaled := downscaleVisionImage(in, "image/png", 7900, zap.NewNop())
	if !scaled {
		t.Fatalf("9000px-wide image must scale")
	}
	w, h := decodeDims(t, out)
	if w != 7900 || h != 200*7900/9000 {
		t.Fatalf("got %dx%d", w, h)
	}
}

// max_image_dimension: 0 is the opt-out — byte-identical pre-2026-09
// behaviour, oversized images included (the provider then errors, which is
// the caller's stated choice).
func TestDownscaleZeroDisables(t *testing.T) {
	in := pngBytes(t, 300, 9000)
	out, mt, scaled := downscaleVisionImage(in, "image/png", 0, zap.NewNop())
	if scaled || mt != "image/png" || !bytes.Equal(in, out) {
		t.Fatalf("cap 0 must be a pure passthrough")
	}
}

// Undecodable bytes pass through for the provider to report — this function
// only measures, it must not become a second, vaguer error site.
func TestDownscalePassesThroughUndecodableBytes(t *testing.T) {
	in := []byte("not an image at all")
	out, mt, scaled := downscaleVisionImage(in, "image/png", 7900, zap.NewNop())
	if scaled || mt != "image/png" || !bytes.Equal(in, out) {
		t.Fatalf("undecodable bytes must pass through untouched")
	}
}
