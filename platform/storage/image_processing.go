// FILE: platform/storage/image_processing.go
// Image download and optimization functions for the storage package

package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
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

	// Get config for content type
	_, _, _, extension := GetImageConfig(purpose)
	contentType := "image/jpeg"
	if extension == "png" {
		contentType = "image/png"
	}

	// Build paths
	paths := BuildAssetPaths(purpose, extension)

	return &ProcessedImage{
		Data:        data,
		ContentType: contentType,
		Paths:       paths,
	}, nil
}
