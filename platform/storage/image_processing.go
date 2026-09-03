// FILE: platform/storage/image_processing.go
// Image download and optimization functions for the storage package

package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	_ "image/png" // PNG decoder
	"io"

	"github.com/nfnt/resize"
	"go.uber.org/zap"
)

// DownloadAndOptimizeImage fetches an image from S3 and optimizes it for web
// Accepts either an S3 URI (s3://bucket/key) or a raw key
func (c *S3Client) DownloadAndOptimizeImage(ctx context.Context, uriOrKey string, purpose string, logger *zap.Logger) ([]byte, error) {
	// Extract key using centralized helper
	key := ExtractKeyFromS3URI(uriOrKey)
	if key == "" {
		return nil, fmt.Errorf("invalid S3 URI or empty key: %s", uriOrKey)
	}

	logger.Info("Downloading image from S3",
		zap.String("key", key),
		zap.String("purpose", purpose))

	// Download
	reader, err := c.Download(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer reader.Close()

	imageData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	logger.Debug("Downloaded image", zap.Int("size_bytes", len(imageData)))

	// Optimize
	optimized, err := OptimizeImageForWeb(imageData, purpose, logger)
	if err != nil {
		logger.Warn("Optimization failed, using original", zap.Error(err))
		return imageData, nil
	}

	return optimized, nil
}

// OptimizeImageForWeb resizes and compresses an image for web delivery
func OptimizeImageForWeb(imageData []byte, purpose string, logger *zap.Logger) ([]byte, error) {
	// Decode
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	logger.Debug("Decoded image",
		zap.String("format", format),
		zap.Int("width", img.Bounds().Dx()),
		zap.Int("height", img.Bounds().Dy()))

	// Get target dimensions from centralized config
	maxWidth, maxHeight, quality, extension := GetImageConfig(purpose)

	// Resize if larger than target
	bounds := img.Bounds()
	if uint(bounds.Dx()) > maxWidth || uint(bounds.Dy()) > maxHeight {
		img = resize.Thumbnail(maxWidth, maxHeight, img, resize.Lanczos3)
		logger.Debug("Resized image",
			zap.Int("new_width", img.Bounds().Dx()),
			zap.Int("new_height", img.Bounds().Dy()))
	}

	// Encode based on target format
	var buf bytes.Buffer
	var encodeErr error

	if extension == "png" {
		encodeErr = png.Encode(&buf, img)
	} else {
		encodeErr = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	}

	if encodeErr != nil {
		return nil, fmt.Errorf("failed to encode image: %w", encodeErr)
	}

	logger.Debug("Optimized image",
		zap.Int("original_size", len(imageData)),
		zap.Int("optimized_size", buf.Len()),
		zap.Int("quality", quality),
		zap.String("format", extension))

	return buf.Bytes(), nil
}

// CoverCropResize scales an image so it fully covers targetW×targetH
// (preserving aspect), then centre-crops to exactly those dimensions —
// CSS object-fit:cover semantics. Unlike resize.Thumbnail (fit-inside,
// output size varies with the source aspect), the result is always exactly
// targetW×targetH, which purpose-derived assets (e.g. the Phase I3 card,
// a fixed 800×450 crop of a hero) require. A same-aspect source degrades
// to a pure resize with a zero-size crop offset.
func CoverCropResize(img image.Image, targetW, targetH uint) image.Image {
	if targetW == 0 || targetH == 0 {
		return img
	}
	b := img.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if srcW <= 0 || srcH <= 0 {
		return img
	}

	// Scale so BOTH dimensions cover the target: pick the larger scale factor.
	// Computing in cross-multiplied integers avoids float drift on the edge
	// cases that matter (exact multiples like 1600×900 → 800×450).
	var scaledW, scaledH uint
	if uint64(srcW)*uint64(targetH) >= uint64(srcH)*uint64(targetW) {
		// Source is wider than target aspect → height binds.
		scaledH = targetH
		scaledW = uint(uint64(srcW) * uint64(targetH) / uint64(srcH))
	} else {
		// Source is taller than target aspect → width binds.
		scaledW = targetW
		scaledH = uint(uint64(srcH) * uint64(targetW) / uint64(srcW))
	}
	scaled := resize.Resize(scaledW, scaledH, img, resize.Lanczos3)

	// Centre-crop the overhang.
	sb := scaled.Bounds()
	offX := (sb.Dx() - int(targetW)) / 2
	offY := (sb.Dy() - int(targetH)) / 2
	if offX < 0 {
		offX = 0
	}
	if offY < 0 {
		offY = 0
	}
	out := image.NewRGBA(image.Rect(0, 0, int(targetW), int(targetH)))
	draw.Draw(out, out.Bounds(), scaled, image.Pt(sb.Min.X+offX, sb.Min.Y+offY), draw.Src)
	return out
}

// ProcessedImage holds the result of image processing
type ProcessedImage struct {
	Data        []byte
	ContentType string
	Paths       AssetPaths
}

// DownloadOptimizeAndPrepare is a convenience function that does everything:
// downloads, optimizes, and prepares paths for deployment
func (c *S3Client) DownloadOptimizeAndPrepare(ctx context.Context, s3URI string, purpose string, logger *zap.Logger) (*ProcessedImage, error) {
	// Download and optimize
	data, err := c.DownloadAndOptimizeImage(ctx, s3URI, purpose, logger)
	if err != nil {
		return nil, err
	}

	// Get config for the deployed path's extension.
	_, _, _, extension := GetImageConfig(purpose)

	// CONTENT TYPE COMES FROM THE BYTES, NOT FROM THE PURPOSE (bugs_open/433).
	//
	// This used to be derived from `extension` alone, which is correct on the
	// happy path — OptimizeImageForWeb re-encodes to png or jpeg to match — and
	// WRONG on the fallback directly above: DownloadAndOptimizeImage returns the
	// ORIGINAL, un-re-encoded bytes when optimisation fails ("Optimization
	// failed, using original"), and this function then confidently reported the
	// format the purpose asked for rather than the one it is about to publish.
	// That value is what deploy_image_asset writes into assets.mime_type, so the
	// purpose-derived version would have filled ~1,000 empty rows with a
	// confidently wrong answer — strictly worse than the empty column it
	// replaced, because an empty mime_type is at least honest and greppable
	// whereas a wrong one looks repaired.
	//
	// A DISAGREEMENT IS LOGGED, NOT REFUSED, and that is deliberate. Refusing
	// would be the tidier invariant and it would take sites down: nothing in
	// this repo registers a webp decoder, yet image/webp is a possible provider
	// response, so today such an image is committed under a .png name and the
	// browser sniffs it. Turning that into a hard failure removes a site's logo
	// to fix a metadata field. So: publish what we have, and say plainly that
	// the name and the bytes disagree.
	contentType := ""
	if sniffedExt, sniffedMIME := SniffImageExtAndMIME(data); sniffedMIME != "" {
		contentType = sniffedMIME
		if sniffedExt != extension {
			logger.Warn("image bytes do not match the extension this purpose deploys under — recording the BYTES, not the name",
				zap.String("purpose", purpose),
				zap.String("deployed_extension", extension),
				zap.String("actual_format", sniffedExt),
				zap.String("content_type", sniffedMIME),
				zap.Int("size_bytes", len(data)))
		}
	} else {
		// Honest unknown. Callers must write NULL rather than invent a type;
		// there is no fallback here by design (see SniffImageExtAndMIME).
		logger.Warn("could not identify the image format from its bytes — content type left empty",
			zap.String("purpose", purpose),
			zap.Int("size_bytes", len(data)))
	}

	// Build paths
	paths := BuildAssetPaths(purpose, extension)

	return &ProcessedImage{
		Data:        data,
		ContentType: contentType,
		Paths:       paths,
	}, nil
}
