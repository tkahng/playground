package filesystem

import (
	"context"
	"io"

	"github.com/tkahng/authgo/internal/shared"
)

type FileSystem interface {
	GeneratePresignedURL(ctx context.Context, bucket string, key string) (string, error)
	NewFile(ctx context.Context, authority string, key string, file io.Reader) error
	NewFileFromBytes(ctx context.Context, b []byte, name string) (*shared.FileDto, error)
	NewFileFromURL(ctx context.Context, url string) (*shared.FileDto, error)
}
