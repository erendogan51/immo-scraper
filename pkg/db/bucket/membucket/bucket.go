package membucket

import (
	"bytes"
	"context"
	"fmt"
	"io"

	common "github.com/erendogan51/immo-scrapper/pkg/db/bucket"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/gcerrors"
)

type MemBucket struct {
	b *blob.Bucket
}

func New() common.Bucket {
	b := memblob.OpenBucket(nil)
	return &MemBucket{
		b: b,
	}
}

func (b *MemBucket) UploadBytes(ctx context.Context, data []byte, filename string, writerOptions *blob.WriterOptions) error {
	if len(data) == 0 {
		return nil
	}

	if writerOptions == nil {
		writerOptions = &blob.WriterOptions{ContentType: "application/octet-stream"}
	}
	err := b.b.Upload(ctx, filename, bytes.NewReader(data), writerOptions)

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (b *MemBucket) Upload(ctx context.Context, key string, file io.Reader, writerOptions *blob.WriterOptions) error {
	if writerOptions == nil {
		writerOptions = &blob.WriterOptions{ContentType: "application/octet-stream"}
	}
	err := b.b.Upload(ctx, key, file, writerOptions)

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (b *MemBucket) Get(ctx context.Context, filename string) ([]byte, error) {
	data, err := b.b.ReadAll(ctx, filename)

	if err != nil {
		code := gcerrors.Code(err)

		if code == gcerrors.NotFound {
			return []byte{}, nil
		}

		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return data, nil
}

func (b *MemBucket) SignedURL(_ context.Context, filename string, opts *blob.SignedURLOptions) (string, error) {
	return "http://abc.com/" + filename, nil
}

func (b *MemBucket) Exists(ctx context.Context, key string) (bool, error) {
	return b.b.Exists(ctx, key)
}

func (b *MemBucket) List(ctx context.Context, prefix string) ([]blob.ListObject, error) {
	list := b.b.List(&blob.ListOptions{Prefix: prefix})

	var objects []blob.ListObject
	for {
		next, err := list.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		objects = append(objects, *next)
	}

	return objects, nil
}
