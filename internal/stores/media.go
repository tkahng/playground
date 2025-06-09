package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
)

type DbMediaStore struct {
	dbx database.Dbx
}

func NewMediaStore(dbx database.Dbx) *DbMediaStore {
	return &DbMediaStore{
		dbx: dbx,
	}
}

func (s *DbMediaStore) UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	data, err := repository.Media.PutOne(
		ctx,
		s.dbx,
		media,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbMediaStore) CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	data, err := repository.Media.PostOne(
		ctx,
		s.dbx,
		media,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbMediaStore) FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error) {
	data, err := repository.Media.GetOne(
		ctx,
		s.dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": mediaId,
			},
		},
	)
	return database.OptionalRow(data, err)
}
