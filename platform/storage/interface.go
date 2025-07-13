// FILE: platform/storage/interface.go
package storage

import (
	"context"
	"io"
	"time"
)

// Client defines the interface for object storage operations
type Client interface {
	// Upload stores an object and returns its URI
	Upload(ctx context.Context, key, contentType string, body io.Reader) (string, error)

	// Download retrieves an object by its key
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an object
	Delete(ctx context.Context, key string) error

	// Exists checks if an object exists
	Exists(ctx context.Context, key string) (bool, error)

	// ListObjects lists objects with a given prefix
	ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error)

	// GetPresignedURL generates a temporary access URL
	GetPresignedURL(ctx context.Context, key string, expiry int) (string, error)
}

// ObjectInfo contains metadata about a stored object
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
}
