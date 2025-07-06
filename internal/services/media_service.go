package services

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/filesystem"
)

type MediaStore interface {
	CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error)
}

type FsService interface {
	GeneratePresignedURL(ctx context.Context, bucket string, key string) (string, error)
	NewFile(ctx context.Context, authority string, key string, file io.Reader) error
	NewFileFromBytes(ctx context.Context, b []byte, name string) (*filesystem.FileDto, error)
	NewFileFromURL(ctx context.Context, url string) (*filesystem.FileDto, error)
}

type MediaService interface {
	Store() MediaStore
}
