package s3client

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client defines the S3 client interface.
type Client interface {
	// GeneratePresignedGetURL generates a presigned GET URL for the given key and duration.
	GeneratePresignedGetURL(key string, duration time.Duration) (string, error)

	// GeneratePresignedPutURL generates a presigned PUT URL for the given key and duration.
	GeneratePresignedPutURL(key string, duration time.Duration) (string, error)

	// DeleteFile deletes a file from the S3 bucket using the given key.
	DeleteFile(key string) error

	// ListFiles lists files in the S3 bucket with pagination support.
	ListFiles(pageSize int32, continuationToken *string) ([]string, *string, error)

	// UploadFile uploads data directly to S3 with configurable content type
	UploadFile(ctx context.Context, key string, data []byte, contentType string) error

	// Helper methods for common content types
	UploadCSV(ctx context.Context, key string, data []byte) error
	UploadJSON(ctx context.Context, key string, data []byte) error
	UploadText(ctx context.Context, key string, data []byte) error
}

// S3Client defines the structure that implements the Client interface.
type S3Client struct {
	s3Client *s3.Client
	bucket   string
}

// Option is a functional option type for configuring S3Client.
type Option func(*S3Client) error

// WithCredentials sets AWS credentials for the S3 client.
func WithCredentials(accessKey, secretKey, sessionToken string) Option {
	return func(c *S3Client) error {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)))
		if err != nil {
			return fmt.Errorf("failed to load configuration with custom credentials: %v", err)
		}
		c.s3Client = s3.NewFromConfig(cfg)
		return nil
	}
}

// WithRegion sets a custom AWS region for the S3 client.
func WithRegion(region string) Option {
	return func(c *S3Client) error {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region))
		if err != nil {
			return fmt.Errorf("failed to load configuration with custom region: %v", err)
		}
		c.s3Client = s3.NewFromConfig(cfg)
		return nil
	}
}

// NewS3Client creates a new S3 client with optional configuration parameters.
func NewS3Client(bucketName string, opts ...Option) (*S3Client, error) {
	client := &S3Client{
		bucket: bucketName,
	}

	// Load the default configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load default configuration: %v", err)
	}
	client.s3Client = s3.NewFromConfig(cfg)

	// Apply any additional options (e.g., credentials, region)
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// GeneratePresignedGetURL generates a presigned GET URL for the given key and duration.
func (c *S3Client) GeneratePresignedGetURL(key string, duration time.Duration) (string, error) {
	psClient := s3.NewPresignClient(c.s3Client)

	req, err := psClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(duration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %v", err)
	}

	return req.URL, nil
}

// GeneratePresignedPutURL generates a presigned PUT URL for the given key and duration.
func (c *S3Client) GeneratePresignedPutURL(key string, duration time.Duration) (string, error) {
	psClient := s3.NewPresignClient(c.s3Client)

	req, err := psClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(duration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %v", err)
	}

	return req.URL, nil
}

// DeleteFile deletes a file from the S3 bucket using the given key.
func (c *S3Client) DeleteFile(key string) error {
	_, err := c.s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// ListFiles lists files in the S3 bucket with pagination support.
func (c *S3Client) ListFiles(pageSize int32, continuationToken *string) ([]string, *string, error) {
	var files []string

	output, err := c.s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:            aws.String(c.bucket),
		MaxKeys:           &pageSize,
		ContinuationToken: continuationToken,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list files: %v", err)
	}

	for _, item := range output.Contents {
		files = append(files, *item.Key)
	}

	return files, output.NextContinuationToken, nil
}

// UploadFile uploads data directly to S3 with configurable content type
func (c *S3Client) UploadFile(ctx context.Context, key string, data []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// UploadCSV is a helper method for uploading CSV files
func (c *S3Client) UploadCSV(ctx context.Context, key string, data []byte) error {
	return c.UploadFile(ctx, key, data, "text/csv")
}

// UploadJSON is a helper method for uploading JSON files
func (c *S3Client) UploadJSON(ctx context.Context, key string, data []byte) error {
	return c.UploadFile(ctx, key, data, "application/json")
}

// UploadText is a helper method for uploading plain text files
func (c *S3Client) UploadText(ctx context.Context, key string, data []byte) error {
	return c.UploadFile(ctx, key, data, "text/plain")
}
