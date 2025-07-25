package filesystem

import (
	"context"
	"io"
)

type FileSystem interface {
	GeneratePresignedURL(ctx context.Context, bucket string, key string) (string, error)
	PutFile(ctx context.Context, authority string, key string, file io.Reader) error
	PutFileFromBytes(ctx context.Context, b []byte, name string) (*FileDto, error)
	PutNewFileFromURL(ctx context.Context, url string) (*FileDto, error)
}
