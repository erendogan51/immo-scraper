package s3bucket

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsv2cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/erendogan51/immo-scraper/pkg/config"
	"gocloud.dev/blob/s3blob"

	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
)

type S3Bucket struct {
	S3Client *s3v2.Client
	Config   config.Bucket
}

func New(ctx context.Context, bucketCfg config.Bucket) (*S3Bucket, error) {
	cfg, err := awsv2cfg.LoadDefaultConfig(
		ctx,
		awsv2cfg.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
		awsv2cfg.WithRegion(bucketCfg.Region),
		awsv2cfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				bucketCfg.Key,
				bucketCfg.Secret,
				"",
			),
		),
	)

	if err != nil {
		return nil, err
	}

	clientV2 := s3v2.NewFromConfig(cfg, func(o *s3v2.Options) {
		o.BaseEndpoint = aws.String(bucketCfg.Endpoint)
		o.UsePathStyle = true
	})

	bucket, err := s3blob.OpenBucketV2(ctx, clientV2, bucketCfg.Name, nil)
	if err != nil {
		return nil, err
	}

	defer closeBucket(bucket)

	accessible, err := bucket.IsAccessible(ctx)
	if err != nil {
		log.Error().Msg("bucket is not accessible")
		return nil, fmt.Errorf("bucket %s is not accessible", bucketCfg.Name)
	}

	if !accessible {
		log.Warn().Msg("bucket is not accessible")

		return nil, fmt.Errorf("bucket %s is not accessible", bucketCfg.Name)
	}

	return &S3Bucket{
		S3Client: clientV2,
		Config:   bucketCfg,
	}, nil
}

func (b *S3Bucket) UploadBytes(ctx context.Context, data []byte, key string, writerOptions *blob.WriterOptions) error {
	if len(data) == 0 {
		return nil
	}

	bucket, err := s3blob.OpenBucketV2(ctx, b.S3Client, b.Config.Name, nil)

	if err != nil {
		return fmt.Errorf("failed to open bucket: %w", err)
	}

	defer closeBucket(bucket)

	if writerOptions == nil {
		writerOptions = &blob.WriterOptions{ContentType: "application/octet-stream"}
	}
	err = bucket.Upload(ctx, key, bytes.NewReader(data), writerOptions)

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (b *S3Bucket) Upload(ctx context.Context, key string, file io.Reader, writerOptions *blob.WriterOptions) error {
	bucket, err := s3blob.OpenBucketV2(ctx, b.S3Client, b.Config.Name, nil)

	if err != nil {
		return fmt.Errorf("failed to open bucket: %w", err)
	}

	defer closeBucket(bucket)

	if writerOptions == nil {
		writerOptions = &blob.WriterOptions{ContentType: "application/octet-stream"}
	}

	err = bucket.Upload(ctx, key, file, writerOptions)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (b *S3Bucket) Get(ctx context.Context, key string) ([]byte, error) {
	bucket, err := s3blob.OpenBucketV2(ctx, b.S3Client, b.Config.Name, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to open bucket: %w", err)
	}

	defer closeBucket(bucket)

	data, err := bucket.ReadAll(ctx, key)

	if err != nil {
		code := gcerrors.Code(err)

		if code == gcerrors.NotFound {
			return []byte{}, nil
		}

		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return data, nil
}

func (b *S3Bucket) SignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error) {
	bucket, err := s3blob.OpenBucketV2(ctx, b.S3Client, b.Config.Name, nil)

	if err != nil {
		return "", fmt.Errorf("failed to open bucket: %w", err)
	}

	defer closeBucket(bucket)

	var options *blob.SignedURLOptions
	if opts != nil {
		options = opts
	} else {
		options = &blob.SignedURLOptions{
			Expiry: 60 * time.Minute,
		}
	}

	url, err := bucket.SignedURL(ctx, key, options)
	if err != nil {
		return "", fmt.Errorf("failed to sign URL: %w", err)
	}

	return url, nil
}

func closeBucket(b *blob.Bucket) {
	if err := b.Close(); err != nil {
		log.Warn().Msgf("Failed to close bucket: %v", err)
	}
}

func (b *S3Bucket) Exists(ctx context.Context, key string) (bool, error) {
	bucket, err := s3blob.OpenBucketV2(ctx, b.S3Client, b.Config.Name, nil)

	if err != nil {
		return false, fmt.Errorf("failed to open bucket: %w", err)
	}

	defer closeBucket(bucket)

	exists, err := bucket.Exists(ctx, key)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (b *S3Bucket) List(ctx context.Context, prefix string) ([]blob.ListObject, error) {
	bucket, err := s3blob.OpenBucketV2(ctx, b.S3Client, b.Config.Name, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to open bucket: %w", err)
	}

	defer closeBucket(bucket)

	list := bucket.List(&blob.ListOptions{Prefix: prefix})

	var objects []blob.ListObject
	for {
		next, err := list.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		objects = append(objects, *next)
	}

	return objects, nil
}
