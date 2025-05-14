package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
)

func CreateMedia(ctx context.Context, exec db.Dbx, media *models.Medium) (*models.Medium, error) {
	data, err := crudrepo.Media.PostOne(
		ctx,
		exec,
		media,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func FindMediaByID(ctx context.Context, exec db.Dbx, id uuid.UUID) (*models.Medium, error) {
	data, err := crudrepo.Media.GetOne(
		ctx,
		exec,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return OptionalRow(data, err)
}
