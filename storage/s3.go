package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/lrndwy/gokil/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

func NewS3(settings config.StorageSettings) (*S3, error) {
	if strings.TrimSpace(settings.Bucket) == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{}
	if settings.Region != "" {
		opts = append(opts, awsconfig.WithRegion(settings.Region))
	}
	if settings.AccessKeyID != "" && settings.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(settings.AccessKeyID, settings.SecretAccessKey, ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if settings.Endpoint != "" {
			o.BaseEndpoint = aws.String(settings.Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3{
		client:  client,
		bucket:  settings.Bucket,
		baseURL: settings.BaseURL,
	}, nil
}

func (s *S3) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}
	_, err := s.client.PutObject(ctx, input)
	return err
}

func (s *S3) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3) URL(key string) (string, error) {
	if s.baseURL != "" {
		return strings.TrimSuffix(s.baseURL, "/") + "/" + key, nil
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key), nil
}
