// FILE: platform/orchestration/actions/vision_image_downscale.go
//
// downscaleVisionImage makes a screenshot legal for the vision providers.
//
// WHY THIS EXISTS (bugs_open/403-era 018 work, 2026-08-27→09-02): full-page
// captures grow with their pages, and a page that gains sections can push its
// capture past the providers' PER-IMAGE limits. Measured on leopardess after
// its hero batch, one run each, three different provider errors for one
// underlying fact:
//   - Gemini:    400 INVALID_ARGUMENT "Unable to process input image" (×2)
//   - Anthropic: 413 request_too_large                       (16 images)
//   - Anthropic: 400 "At least one of the image dimensions exceed max
//                 allowed size: 8000 pixels"                 (8 images)
// No config lever can exclude one too-tall page honestly (max_images drops
// from the END of the list, not by size), so the fix is here: any image whose
// longest edge exceeds the cap is scaled down to fit, preserving aspect, and
// re-encoded as JPEG (quality 85 — a large payload reduction on top of the
// scaling, and taste review does not need lossless).
//
// DEFAULT ON, at 7900 (just under Anthropic's stated 8000px cap), and this is
// a deliberate reading of the 2026-08-02 opt-in ruling rather than a breach of
// it: the ruling guards NEW AUTHORITY whose unsafe side must default OFF. This
// path's "authority" only touches images that today produce a GUARANTEED
// provider 400/413 — for every image at or under the cap (which includes every
// image of all 41 successful tool-acceptance-agent runs, and the one
// successful design-critique run) the bytes pass through UNTOUCHED, so every
// historically-working consumer is byte-identical. Opt out with
// max_image_dimension: 0 for the pre-2026-09 behaviour, cap included.
//
// The scaler is a plain box average (area average per target pixel). For large
// downscale factors on UI screenshots that is the RIGHT filter (it is exactly
// the anti-aliasing bilinear misses at big ratios), and it keeps this file
// free of new dependencies — golang.org/x/image is not in go.mod and one
// function does not justify adding it.

package actions

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png" // decode support; captures arrive as PNG

	"go.uber.org/zap"
)

// visionImageDimensionCapDefault is just under Anthropic's hard 8000px
// per-dimension limit; Gemini's effective limits are lower still, so an image
// legal here is legal there after this pass.
const visionImageDimensionCapDefault = 7900

// downscaleVisionImage returns image bytes whose longest edge is <= maxDim,
// with the media type they are encoded as, and whether scaling happened.
// maxDim <= 0 disables the pass entirely (bytes returned untouched).
// A decode failure returns the original bytes untouched with scaled=false —
// the provider then reports the undecodable image itself, which is a clearer
// error than failing here on a byte stream this function only needed to
// measure.
func downscaleVisionImage(data []byte, mediaType string, maxDim int, logger *zap.Logger) ([]byte, string, bool) {
	if maxDim <= 0 {
		return data, mediaType, false
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		if logger != nil {
			logger.Warn("downscaleVisionImage: undecodable image passed through", zap.Error(err))
		}
		return data, mediaType, false
	}
	if cfg.Width <= maxDim && cfg.Height <= maxDim {
		return data, mediaType, false
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		if logger != nil {
			logger.Warn("downscaleVisionImage: decode failed after DecodeConfig succeeded", zap.Error(err))
		}
		return data, mediaType, false
	}

	srcW, srcH := cfg.Width, cfg.Height
	long := srcW
	if srcH > long {
		long = srcH
	}
	// Integer-safe target: scale the long edge to maxDim exactly, the short
	// edge proportionally (minimum 1px).
	dstW := srcW * maxDim / long
	dstH := srcH * maxDim / long
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	scaled := boxAverageScale(img, srcW, srcH, dstW, dstH)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, scaled, &jpeg.Options{Quality: 85}); err != nil {
		if logger != nil {
			logger.Warn("downscaleVisionImage: jpeg encode failed, passing original through", zap.Error(err))
		}
		return data, mediaType, false
	}
	if logger != nil {
		logger.Info("downscaleVisionImage: image scaled for provider limits",
			zap.Int("src_w", srcW), zap.Int("src_h", srcH),
			zap.Int("dst_w", dstW), zap.Int("dst_h", dstH),
			zap.Int("bytes_in", len(data)), zap.Int("bytes_out", out.Len()))
	}
	return out.Bytes(), "image/jpeg", true
}

// boxAverageScale area-averages src (srcW×srcH) into an RGBA image of
// dstW×dstH. Each destination pixel is the mean of the source rectangle it
// covers — box filtering, the correct anti-aliasing for large downscales.
func boxAverageScale(src image.Image, srcW, srcH, dstW, dstH int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	b := src.Bounds()
	for dy := 0; dy < dstH; dy++ {
		sy0 := b.Min.Y + dy*srcH/dstH
		sy1 := b.Min.Y + (dy+1)*srcH/dstH
		if sy1 <= sy0 {
			sy1 = sy0 + 1
		}
		for dx := 0; dx < dstW; dx++ {
			sx0 := b.Min.X + dx*srcW/dstW
			sx1 := b.Min.X + (dx+1)*srcW/dstW
			if sx1 <= sx0 {
				sx1 = sx0 + 1
			}
			var rs, gs, bs, as, n uint64
			for sy := sy0; sy < sy1; sy++ {
				for sx := sx0; sx < sx1; sx++ {
					r, g, bb, a := src.At(sx, sy).RGBA()
					rs += uint64(r >> 8)
					gs += uint64(g >> 8)
					bs += uint64(bb >> 8)
					as += uint64(a >> 8)
					n++
				}
			}
			i := dst.PixOffset(dx, dy)
			dst.Pix[i+0] = uint8(rs / n)
			dst.Pix[i+1] = uint8(gs / n)
			dst.Pix[i+2] = uint8(bs / n)
			dst.Pix[i+3] = uint8(as / n)
		}
	}
	return dst
}
