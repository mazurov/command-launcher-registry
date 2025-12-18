package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 timeout constants
const (
	S3UploadTimeout   = 60 * time.Second
	S3DownloadTimeout = 30 * time.Second
)

// S3Client wraps MinIO SDK for S3 operations
type S3Client struct {
	client *minio.Client
	bucket string
	key    string
	logger *slog.Logger
}

// NewS3Client creates a new S3 client for the given endpoint and credentials.
func NewS3Client(endpoint, bucket, key, accessKey, secretKey string, useSSL bool, region string, logger *slog.Logger) (*S3Client, error) {
	start := time.Now()

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	}

	// Set region if provided
	if region != "" {
		opts.Region = region
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		logger.Error("Failed to create S3 client",
			"endpoint", endpoint,
			"bucket", bucket,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeS3Error(S3OpConnect, fmt.Errorf("failed to create S3 client: %w", err))
	}

	logger.Info("S3 client created",
		"endpoint", endpoint,
		"bucket", bucket,
		"key", key,
		"ssl", useSSL,
		"region", region,
		"duration_ms", time.Since(start).Milliseconds())

	return &S3Client{
		client: client,
		bucket: bucket,
		key:    key,
		logger: logger,
	}, nil
}

// ValidateBucket checks if the bucket exists and is accessible
func (c *S3Client) ValidateBucket(ctx context.Context) error {
	start := time.Now()
	c.logger.Debug("Validating S3 bucket", "bucket", c.bucket)

	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		c.logger.Error("S3 bucket validation failed",
			"bucket", c.bucket,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return CategorizeS3Error(S3OpConnect, err)
	}

	if !exists {
		c.logger.Error("S3 bucket does not exist",
			"bucket", c.bucket,
			"duration_ms", time.Since(start).Milliseconds())
		return CategorizeS3Error(S3OpConnect, fmt.Errorf("bucket %q does not exist", c.bucket))
	}

	c.logger.Info("S3 bucket validated",
		"bucket", c.bucket,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

// Exists checks if the object exists in the S3 bucket
func (c *S3Client) Exists(ctx context.Context) (bool, error) {
	start := time.Now()
	c.logger.Debug("Checking S3 object existence", "bucket", c.bucket, "key", c.key)

	_, err := c.client.StatObject(ctx, c.bucket, c.key, minio.StatObjectOptions{})
	if err != nil {
		// Check if it's a "not found" error
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			c.logger.Info("S3 object does not exist",
				"bucket", c.bucket,
				"key", c.key,
				"duration_ms", time.Since(start).Milliseconds())
			return false, nil
		}
		c.logger.Error("S3 existence check failed",
			"bucket", c.bucket,
			"key", c.key,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return false, CategorizeS3Error(S3OpConnect, err)
	}

	c.logger.Info("S3 object exists",
		"bucket", c.bucket,
		"key", c.key,
		"duration_ms", time.Since(start).Milliseconds())
	return true, nil
}

// Upload uploads data to the S3 bucket
func (c *S3Client) Upload(ctx context.Context, data []byte) error {
	start := time.Now()
	c.logger.Info("Starting S3 upload",
		"bucket", c.bucket,
		"key", c.key,
		"size_bytes", len(data))

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, S3UploadTimeout)
	defer cancel()

	reader := bytes.NewReader(data)
	_, err := c.client.PutObject(ctx, c.bucket, c.key, reader, int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/json",
		},
	)
	if err != nil {
		c.logger.Error("S3 upload failed",
			"bucket", c.bucket,
			"key", c.key,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return CategorizeS3Error(S3OpUpload, err)
	}

	c.logger.Info("S3 upload completed",
		"bucket", c.bucket,
		"key", c.key,
		"size_bytes", len(data),
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}

// Download downloads data from the S3 bucket
func (c *S3Client) Download(ctx context.Context) ([]byte, error) {
	start := time.Now()
	c.logger.Debug("Starting S3 download", "bucket", c.bucket, "key", c.key)

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, S3DownloadTimeout)
	defer cancel()

	obj, err := c.client.GetObject(ctx, c.bucket, c.key, minio.GetObjectOptions{})
	if err != nil {
		c.logger.Error("S3 download failed",
			"bucket", c.bucket,
			"key", c.key,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeS3Error(S3OpDownload, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		c.logger.Error("S3 download read failed",
			"bucket", c.bucket,
			"key", c.key,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeS3Error(S3OpDownload, err)
	}

	c.logger.Info("S3 download completed",
		"bucket", c.bucket,
		"key", c.key,
		"size_bytes", len(data),
		"duration_ms", time.Since(start).Milliseconds())
	return data, nil
}

// ParseS3Token parses the storage token into access key and secret key.
// Token format: ACCESS_KEY:SECRET_KEY
// Falls back to AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY env vars if token is empty.
func ParseS3Token(token string) (accessKey, secretKey string, err error) {
	if token == "" {
		// Fallback to AWS environment variables
		accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey != "" && secretKey != "" {
			return accessKey, secretKey, nil
		}
		// Allow empty credentials for IAM role authentication
		if accessKey == "" && secretKey == "" {
			return "", "", nil
		}
		return "", "", fmt.Errorf("S3 credentials incomplete: set both AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY, or use --storage-token ACCESS_KEY:SECRET_KEY")
	}

	// Split on first colon only (secret key may contain colons)
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid token format: expected ACCESS_KEY:SECRET_KEY")
	}

	accessKey = parts[0]
	secretKey = parts[1]

	if accessKey == "" {
		return "", "", fmt.Errorf("invalid token format: access key cannot be empty")
	}
	if secretKey == "" {
		return "", "", fmt.Errorf("invalid token format: secret key cannot be empty")
	}

	return accessKey, secretKey, nil
}

// ExtractRegionFromEndpoint extracts AWS region from endpoint URL.
// Supports patterns: s3.REGION.amazonaws.com and s3-REGION.amazonaws.com
func ExtractRegionFromEndpoint(endpoint string) string {
	// Pattern: s3.REGION.amazonaws.com
	re1 := regexp.MustCompile(`s3\.([a-z]{2}-[a-z]+-\d+)\.amazonaws\.com`)
	if matches := re1.FindStringSubmatch(endpoint); len(matches) > 1 {
		return matches[1]
	}

	// Pattern: s3-REGION.amazonaws.com
	re2 := regexp.MustCompile(`s3-([a-z]{2}-[a-z]+-\d+)\.amazonaws\.com`)
	if matches := re2.FindStringSubmatch(endpoint); len(matches) > 1 {
		return matches[1]
	}

	return ""
}
