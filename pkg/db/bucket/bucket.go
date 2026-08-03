package bucket

import (
	"context"
	"io"

	"gocloud.dev/blob"
)

type Bucket interface {
	UploadBytes(ctx context.Context, data []byte, key string, writerOptions *blob.WriterOptions) error
	Upload(ctx context.Context, key string, r io.Reader, opts *blob.WriterOptions) error
	Get(ctx context.Context, key string) ([]byte, error)
	SignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	List(ctx context.Context, prefix string) ([]blob.ListObject, error)
}
