// FILE: platform/storage/uri_helpers.go
// Centralized URI parsing and asset path utilities

package storage

import (
	"fmt"
	"strings"
)

// S3URIPrefix is the standard S3 URI scheme
const S3URIPrefix = "s3://"

// DefaultAssetBasePath is the base path for web assets
const DefaultAssetBasePath = "assets/images"

// ParseS3URI extracts bucket and key from an S3 URI
// Input: s3://bucket-name/path/to/object
// Returns: bucket, key, error
func ParseS3URI(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(uri, S3URIPrefix) {
		return "", "", fmt.Errorf("invalid S3 URI: must start with %s", S3URIPrefix)
	}

	// Remove prefix: bucket-name/path/to/object
	path := uri[len(S3URIPrefix):]

	// Split into bucket and key
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", "", fmt.Errorf("invalid S3 URI: missing bucket name")
	}

	bucket = parts[0]
	if len(parts) == 2 {
		key = parts[1]
	}

	return bucket, key, nil
}

// ExtractKeyFromS3URI extracts just the key (path) from an S3 URI
// This is a convenience function when you don't need the bucket
func ExtractKeyFromS3URI(uri string) string {
	if !strings.HasPrefix(uri, S3URIPrefix) {
		return uri // Return as-is if not an S3 URI
	}

	path := uri[len(S3URIPrefix):]
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// BuildS3URI constructs an S3 URI from bucket and key
func BuildS3URI(bucket, key string) string {
	return fmt.Sprintf("%s%s/%s", S3URIPrefix, bucket, key)
}

// IsS3URI checks if a string is an S3 URI
func IsS3URI(uri string) bool {
	return strings.HasPrefix(uri, S3URIPrefix)
}

// PresignedURLToS3URI converts a presigned S3/B2 HTTPS URL to an s3:// URI.
//
// Input:  https://s3.us-east-005.backblazeb2.com/bucket-name/path/to/file.png?X-Amz-Algorithm=...
// Output: s3://bucket-name/path/to/file.png
//
// Returns "" if the URL doesn't match the expected pattern.
// The query string (signature, expiry, etc.) is stripped.
func PresignedURLToS3URI(presignedURL string) string {
	if presignedURL == "" {
		return ""
	}

	// Strip query parameters
	base := presignedURL
	if idx := strings.Index(base, "?"); idx >= 0 {
		base = base[:idx]
	}

	// Find the path portion after the host
	// Pattern: https://s3.region.backblazeb2.com/bucket/key
	//          https://bucket.s3.region.amazonaws.com/key  (virtual-hosted)
	protocolEnd := strings.Index(base, "://")
	if protocolEnd < 0 {
		return ""
	}
	afterProtocol := base[protocolEnd+3:]

	// Find first slash after host
	slashIdx := strings.Index(afterProtocol, "/")
	if slashIdx < 0 {
		return ""
	}

	// Path is /bucket/key — trim leading slash, split into bucket + key
	path := strings.TrimPrefix(afterProtocol[slashIdx:], "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}

	return BuildS3URI(parts[0], parts[1])
}

// AssetPaths holds the different path formats for an asset
type AssetPaths struct {
	// RelativeURL is the web-accessible path (e.g., /assets/images/hero.jpg)
	RelativeURL string

	// FilePath is the file system path without leading slash (e.g., assets/images/hero.jpg)
	FilePath string

	// Filename is just the filename (e.g., hero.jpg)
	Filename string
}

// BuildAssetPaths generates consistent paths for an asset based on purpose
func BuildAssetPaths(purpose string, extension string) AssetPaths {
	if extension == "" {
		extension = "jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	filename := purpose + extension
	filePath := fmt.Sprintf("%s/%s", DefaultAssetBasePath, filename)
	relativeURL := "/" + filePath

	return AssetPaths{
		RelativeURL: relativeURL,
		FilePath:    filePath,
		Filename:    filename,
	}
}

// ImagePurposes defines valid image purposes and their configurations
var ImagePurposes = map[string]struct {
	Width     uint
	Height    uint
	Quality   int
	Extension string
}{
	"hero":          {1600, 900, 85, "jpg"},
	"hero_home":     {1600, 900, 85, "jpg"},
	"hero_about":    {1600, 900, 85, "jpg"},
	"hero_services": {1600, 900, 85, "jpg"},
	"logo":          {400, 400, 90, "png"},
	"icon":          {240, 240, 85, "jpg"},
	"default":       {1200, 800, 85, "jpg"},
}

// GetImageConfig returns the configuration for an image purpose
func GetImageConfig(purpose string) (width, height uint, quality int, extension string) {
	cfg, ok := ImagePurposes[purpose]
	if !ok {
		cfg = ImagePurposes["default"]
	}
	return cfg.Width, cfg.Height, cfg.Quality, cfg.Extension
}
