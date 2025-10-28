// FILE: platform/storage/s3.go
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	platform_config "github.com/gqls/agentchassis/platform/config"
	"go.uber.org/zap"
)

// S3Client implements the Client interface for S3-compatible services
type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(ctx context.Context, cfg platform_config.ObjectStorageConfig, logger zap.Logger) (*S3Client, error) {
	// Use the env vars specified in config
	accessKey := os.Getenv(cfg.AccessKeyEnvVar)
	secretKey := os.Getenv(cfg.SecretKeyEnvVar)

	// Add debug logging
	logger.Info("NewS3Client",
		zap.String("DEBUG: AccessKeyEnvVar=%s", cfg.AccessKeyEnvVar),
		zap.String("DEBUG: cfg.SecretKeyEnvVar", cfg.AccessKeyEnvVar),
		zap.String("DEBUG: AccessKey", accessKey),
		zap.String("DEBUGaa: B2_APPLICATION_KEY_ID from env", os.Getenv("B2_APPLICATION_KEY_ID")),
		zap.String("DEBUGaa: B2_APPLICATION_KEY from env", os.Getenv("B2_APPLICATION_KEY")),
		zap.String("DEBUG: Endpoint", cfg.Endpoint),
		zap.String("DEBUG: Bucket", cfg.Bucket),
	)

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("object storage credentials not found in environment variables (%s, %s)",
			cfg.AccessKeyEnvVar, cfg.SecretKeyEnvVar)
	}

	// Get endpoint from config or environment
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = os.Getenv("S3_ENDPOINT")
		if endpoint == "" {
			endpoint = "https://s3.us-east-005.backblazeb2.com" // B2 default
		}
	}

	// Get region from environment (since it might not be in your struct)
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-005" // B2 default region
	}

	// Custom resolver for B2
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && endpoint != "" {
			return aws.Endpoint{
				URL:               endpoint,
				SigningRegion:     region,
				HostnameImmutable: true, // Important for B2
				SigningName:       "s3",
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load s3 config: %w", err)
	}

	// Determine path style - B2 doesn't use path-style
	usePathStyle := false
	if os.Getenv("S3_USE_PATH_STYLE") == "true" {
		usePathStyle = true
	}
	// Override for MinIO
	if strings.Contains(endpoint, "minio") || strings.Contains(endpoint, "localhost") {
		usePathStyle = true
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
	})

	// Get bucket from config or environment
	bucket := cfg.Bucket
	if bucket == "" {
		bucket = os.Getenv("IMAGE_BUCKET")
		if bucket == "" {
			return nil, fmt.Errorf("no bucket specified in config or IMAGE_BUCKET env var")
		}
	}

	return &S3Client{
		client: s3Client,
		bucket: bucket,
	}, nil
}

// Upload puts a new object into the storage bucket
func (c *S3Client) Upload(ctx context.Context, key, contentType string, body io.Reader) (string, error) {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to s3: %w", err)
	}
	// Return the S3 URI for the object
	return fmt.Sprintf("s3://%s/%s", c.bucket, key), nil
}

// Download retrieves an object from the storage bucket
func (c *S3Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object from s3: %w", err)
	}
	return output.Body, nil
}

// Delete removes an object from storage
func (c *S3Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// Exists checks if an object exists
func (c *S3Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// ListObjects lists objects with a given prefix
func (c *S3Client) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	})

	var objects []ObjectInfo
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			objects = append(objects, ObjectInfo{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
				ETag:         aws.ToString(obj.ETag),
			})
		}
	}

	return objects, nil
}

// GetPresignedURL generates a temporary access URL
func (c *S3Client) GetPresignedURL(ctx context.Context, key string, expiryMinutes int) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expiryMinutes) * time.Minute
	})

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return request.URL, nil
}
