// FILE: platform/storage/image_format.go
//
// One place that answers "what format are these bytes, really?", for the three
// call sites that each had their own answer and disagreed with each other.
//
// WHY THIS EXISTS (bugs_open/433). The estate decided an image's format in four
// places and measured it in none of them:
//
//   - internal/adapters/imagegenerator/dynamic_adapter.go named every generated
//     object `<uuid>.png` and uploaded it with Content-Type image/png, both
//     unconditional literals. Measured 2026-09-02 by the 417/420 lane: 12 of 12
//     `logo` source objects in B2, spanning 2026-08-10 to 2026-09-02, carry JPEG
//     magic (ffd8ffe0). None is a PNG. So B2 serves JPEG bytes as image/png, and
//     the extension, the filename and the url are all three unusable as evidence
//     of format.
//   - platform/storage/image_processing.go derived the deployed artefact's
//     Content-Type from the PURPOSE, not the bytes — which is right on the happy
//     path (OptimizeImageForWeb re-encodes to match) and wrong on the fallback at
//     DownloadAndOptimizeImage, which returns the ORIGINAL bytes when
//     optimisation fails and logs "Optimization failed, using original".
//   - assets.mime_type was left unwritten by the two largest writers, so 1,023 of
//     1,418 rows carried nothing at all [MEASURED 2026-09-03].
//
// THE RULE THIS FILE ENFORCES: the bytes are the only honest source. Not the
// extension, not the filename, not the url, not the provider's own claim (the
// banana provider reports the wire MIME and it can be empty; the stability
// provider hardcodes image/png for every response).
//
// ⚠ AND THE RULE THAT MATTERS MOST: THERE IS NO FALLBACK. Unrecognised input
// returns ("", ""), never a guess. Every defect above began as a plausible
// default — `mimeFromKey`'s `default: return "image/png"` is commented "PNG is
// the safest fallback", and it is exactly what makes a mislabelled object
// undetectable. An empty answer is honest and greppable; a confident wrong one
// looks repaired and can only be disproved by re-reading every object. If you
// are tempted to add a default arm here, that is the bug.
//
// Magic bytes, not image.DecodeConfig, deliberately: Go's image registry is
// process-global and populated by whichever blank imports the linked binary
// happens to pull in, so a DecodeConfig-based answer would differ between
// binaries. Magic bytes are deterministic and testable against fixed slices.
package storage

import "bytes"

// SniffImageExtAndMIME reports the file extension and MIME type of an image from
// its leading bytes. It returns ("", "") for anything it does not recognise —
// see the file header: that is the point, not an omission.
//
// The extension is returned WITHOUT a leading dot, to match GetImageConfig and
// ImagePurposes, whose Extension fields are spelled "jpg" / "png".
func SniffImageExtAndMIME(data []byte) (ext, mime string) {
	switch {
	case len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "png", "image/png"
	case len(data) >= 3 && bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "jpg", "image/jpeg"
	case len(data) >= 6 && (bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a"))):
		return "gif", "image/gif"
	case len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		// Recognised on purpose even though nothing in this repo can DECODE a
		// webp: the whole point is to be able to say what was stored. A webp
		// reaching the deploy path is a real possibility (the banana provider
		// lists image/webp as a wire type) and today it is committed under a
		// .png name with browsers left to sniff it.
		return "webp", "image/webp"
	default:
		return "", ""
	}
}

// ImageExtAndMIME maps a Go image format name (as returned by image.Decode or
// image.DecodeConfig) to the same extension/MIME pair, so a caller that has
// already decoded does not need a second, independently-drifting table.
//
// Same no-fallback rule as above.
func ImageExtAndMIME(format string) (ext, mime string) {
	switch format {
	case "png":
		return "png", "image/png"
	case "jpeg":
		return "jpg", "image/jpeg"
	case "gif":
		return "gif", "image/gif"
	case "webp":
		return "webp", "image/webp"
	default:
		return "", ""
	}
}
