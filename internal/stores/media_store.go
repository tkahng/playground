package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
)

type DbMediaStore struct {
	dbx database.Dbx
}

func (s *DbMediaStore) WithTx(tx database.Dbx) services.MediaStore {
	return &DbMediaStore{
		dbx: tx,
	}
}

var _ services.MediaStore = (*DbMediaStore)(nil)

func NewMediaStore(dbx database.Dbx) services.MediaStore {
	return &DbMediaStore{
		dbx: dbx,
	}
}

// UpdateMedia implements services.MediaStore.
func (s *DbMediaStore) UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	data, err := crudrepo.Media.PutOne(
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
	data, err := crudrepo.Media.PostOne(
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
	data, err := crudrepo.Media.GetOne(
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
