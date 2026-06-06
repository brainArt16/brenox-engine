package storage

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
}

func LoadConfig() Config {
	useSSL := strings.EqualFold(os.Getenv("S3_USE_SSL"), "true")
	return Config{
		Endpoint:  strings.TrimSpace(os.Getenv("S3_ENDPOINT")),
		Bucket:    strings.TrimSpace(os.Getenv("S3_BUCKET")),
		AccessKey: strings.TrimSpace(os.Getenv("S3_ACCESS_KEY")),
		SecretKey: strings.TrimSpace(os.Getenv("S3_SECRET_KEY")),
		Region:    envOr("S3_REGION", "us-east-1"),
		UseSSL:    useSSL,
	}
}

func (c Config) Enabled() bool {
	return c.Endpoint != "" && c.Bucket != "" && c.AccessKey != "" && c.SecretKey != ""
}

type Client struct {
	bucket string
	client *s3.Client
	presign *s3.PresignClient
}

func NewClient(cfg Config) (*Client, error) {
	if !cfg.Enabled() {
		return nil, fmt.Errorf("S3 storage is not configured")
	}

	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               fmt.Sprintf("%s://%s", scheme, cfg.Endpoint),
			HostnameImmutable: true,
		}, nil
	})

	awsCfg := aws.Config{
		Region:                      cfg.Region,
		Credentials:                 credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		EndpointResolverWithOptions: resolver,
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfg.Bucket)})
	if err != nil {
		_, createErr := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(cfg.Bucket)})
		if createErr != nil {
			return nil, fmt.Errorf("ensure bucket %s: %w", cfg.Bucket, createErr)
		}
	}

	return &Client{
		bucket:  cfg.Bucket,
		client:  s3Client,
		presign: s3.NewPresignClient(s3Client),
	}, nil
}

func (c *Client) PresignPut(ctx context.Context, key, contentType string, size int64) (string, time.Time, error) {
	expires := 15 * time.Minute
	result, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", time.Time{}, err
	}
	return result.URL, time.Now().UTC().Add(expires), nil
}

func (c *Client) PresignGet(ctx context.Context, key string) (string, time.Time, error) {
	expires := 15 * time.Minute
	result, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", time.Time{}, err
	}
	return result.URL, time.Now().UTC().Add(expires), nil
}

func (c *Client) HeadObject(ctx context.Context, key string) (int64, string, error) {
	out, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, "", err
	}

	size := int64(0)
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	contentType := ""
	if out.ContentType != nil {
		contentType = *out.ContentType
	}
	return size, contentType, nil
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func MaxUploadBytes() int64 {
	raw := strings.TrimSpace(os.Getenv("MAX_UPLOAD_BYTES"))
	if raw == "" {
		return 10 * 1024 * 1024
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 10 * 1024 * 1024
	}
	return value
}
