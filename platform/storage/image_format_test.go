package storage

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"go.uber.org/zap"
)

// Tests for bugs_open/433 — assets.mime_type was empty on 1,023 of 1,418 rows,
// and the generated objects that DID carry a type were labelled by an
// unconditional literal rather than by their content.
//
// Every assertion here runs against REALLY ENCODED BYTES, in the shape of
// keyground_test.go's TestKeyOutBackground_EncodesRealAlphaChannel. A test that
// asserts a string equals a string would pass just as happily against the
// hardcoded constants this change exists to remove.

func realPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	img.Set(0, 0, color.RGBA{1, 2, 3, 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return b.Bytes()
}

func realJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return b.Bytes()
}

func realGIF(t *testing.T) []byte {
	t.Helper()
	img := image.NewPaletted(image.Rect(0, 0, 8, 8), color.Palette{color.Black, color.White})
	var b bytes.Buffer
	if err := gif.Encode(&b, img, nil); err != nil {
		t.Fatalf("encode gif: %v", err)
	}
	return b.Bytes()
}

// The core claim: the answer comes from the bytes.
//
// Mutation that must kill it: swap either magic-byte prefix, or make the
// function consult a filename.
func TestSniffImageExtAndMIMEReadsTheBytes(t *testing.T) {
	for _, tc := range []struct {
		name          string
		data          []byte
		wantExt, want string
	}{
		{"png", realPNG(t), "png", "image/png"},
		{"jpeg", realJPEG(t), "jpg", "image/jpeg"},
		{"gif", realGIF(t), "gif", "image/gif"},
	} {
		ext, mime := SniffImageExtAndMIME(tc.data)
		if ext != tc.wantExt || mime != tc.want {
			t.Errorf("%s: got (%q, %q), want (%q, %q)", tc.name, ext, mime, tc.wantExt, tc.want)
		}
	}
}

// THE MOST IMPORTANT TEST IN THIS FILE. Every defect bugs_open/433 documents
// began as a plausible default: mimeFromKey's `default: return "image/png"` is
// commented "PNG is the safest fallback", and it is precisely what makes a
// mislabelled object undetectable. An empty answer is honest and greppable.
//
// Mutation that must kill it: add any default arm returning a concrete type.
func TestSniffImageExtAndMIMEHasNoConfidentDefault(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"too short to be anything", []byte{0x89, 0x50}},
		{"html", []byte("<html><body>not an image</body></html>")},
		{"text", []byte("plain text that happens to be long enough to sniff")},
	} {
		ext, mime := SniffImageExtAndMIME(tc.data)
		if ext != "" || mime != "" {
			t.Errorf("%s: unrecognised input must return empty, got (%q, %q) — a confident wrong answer is worse than none (bugs_open/433)", tc.name, ext, mime)
		}
	}
}

// The two tables must not drift apart: a caller that has already decoded uses
// ImageExtAndMIME, one holding raw bytes uses the sniffer, and they must agree.
//
// Mutation that must kill it: change one table's spelling ("jpeg" vs "jpg").
func TestSniffAgreesWithTheDecodedFormatName(t *testing.T) {
	for _, tc := range []struct {
		data   []byte
		format string
	}{
		{realPNG(t), "png"},
		{realJPEG(t), "jpeg"},
		{realGIF(t), "gif"},
	} {
		sExt, sMIME := SniffImageExtAndMIME(tc.data)
		fExt, fMIME := ImageExtAndMIME(tc.format)
		if sExt != fExt || sMIME != fMIME {
			t.Errorf("format %q: sniff gave (%q,%q), table gave (%q,%q)", tc.format, sExt, sMIME, fExt, fMIME)
		}
	}
	if ext, mime := ImageExtAndMIME("tiff"); ext != "" || mime != "" {
		t.Errorf("an unmapped format must return empty, got (%q, %q)", ext, mime)
	}
}

// The invariant that licenses writing this value into assets.mime_type at all:
// every purpose the deployer can be asked for must re-encode to the extension
// it publishes under, so bytes and name agree on the happy path.
//
// ⚠ If this fails for a purpose, the deployed-artefact semantics is wrong for
// that purpose and the backfill discussed in bugs_open/433 must not be run
// against it.
//
// Mutation that must kill it: flip the `extension == "png"` branch in
// OptimizeImageForWeb, or add a purpose whose extension it cannot produce.
func TestEveryImagePurposeEncodesToItsDeclaredExtension(t *testing.T) {
	logger := zap.NewNop()
	for purpose, cfg := range ImagePurposes {
		// Feed the OPPOSITE format in, so a pass means it re-encoded rather
		// than passed the input through.
		var in []byte
		if cfg.Extension == "png" {
			in = realJPEG(t)
		} else {
			in = realPNG(t)
		}
		out, err := OptimizeImageForWeb(in, purpose, logger)
		if err != nil {
			t.Errorf("purpose %q: optimise failed: %v", purpose, err)
			continue
		}
		gotExt, gotMIME := SniffImageExtAndMIME(out)
		if gotExt != cfg.Extension {
			t.Errorf("purpose %q deploys under .%s but the encoder produced %s (%s) — the deployed extension is not evidence of format for this purpose",
				purpose, cfg.Extension, gotExt, gotMIME)
		}
	}
}
