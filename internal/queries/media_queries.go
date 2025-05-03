package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
)

func CreateMedia(ctx context.Context, exec Queryer, media *crudModels.Medium) (*crudModels.Medium, error) {
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

func FindMediaByID(ctx context.Context, exec Queryer, id uuid.UUID) (*crudModels.Medium, error) {
	// data, err := models.Media.Query(
	// 	models.SelectWhere.Media.ID.EQ(id),
	// ).One(ctx, exec)
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
