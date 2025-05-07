package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
)

func CreateMedia(ctx context.Context, exec db.Dbx, media *models.Medium) (*models.Medium, error) {
	// q := models.Media.Insert(
	// 	&models.MediumSetter{
	// 		UserID:           omitnull.FromNull(media.UserID),
	// 		Disk:             omit.From(media.Disk),
	// 		Directory:        omit.From(media.Directory),
	// 		Filename:         omit.From(media.Filename),
	// 		OriginalFilename: omit.From(media.OriginalFilename),
	// 		Extension:        omit.From(media.Extension),
	// 		MimeType:         omit.From(media.MimeType),
	// 		Size:             omit.From(media.Size),
	// 	},
	// 	im.Returning("*"),
	// )
	// d, err := q.One(ctx, exec)
	// d, err = OptionalRow(d, err)
	data, err := repository.Media.PostOne(
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
	// data, err := models.Media.Query(
	// 	models.SelectWhere.Media.ID.EQ(id),
	// ).One(ctx, exec)
	data, err := repository.Media.GetOne(
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
