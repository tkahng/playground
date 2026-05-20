package services

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/filesystem"
)

type MediaStore interface {
	CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error)
}

type FsService interface {
	NewFile(ctx context.Context, authority string, key string, file io.Reader) error
	NewFileFromBytes(ctx context.Context, b []byte, name string) (*filesystem.FileDto, error)
	NewFileFromURL(ctx context.Context, url string) (*filesystem.FileDto, error)
	PublicURL(key string) string
}

type MediaService interface {
	Store() MediaStore
}
