package filesystem

import (
	"context"
	"io"
)

type FileSystem interface {
	PutFile(ctx context.Context, authority string, key string, file io.Reader) error
	PutFileFromBytes(ctx context.Context, b []byte, name string) (*FileDto, error)
	PutNewFileFromURL(ctx context.Context, url string) (*FileDto, error)
	DeleteObject(ctx context.Context, key string) error
	PublicURL(key string) string
}
