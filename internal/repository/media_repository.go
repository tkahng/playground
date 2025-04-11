package repository

import (
	"context"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/tkahng/authgo/internal/db/models"
)

func CreateMedia(ctx context.Context, exec bob.Executor, media *models.Medium) (*models.Medium, error) {
	q := models.Media.Insert(
		&models.MediumSetter{
			UserID:           omitnull.FromNull(media.UserID),
			Disk:             omit.From(media.Disk),
			Directory:        omit.From(media.Directory),
			Filename:         omit.From(media.Filename),
			OriginalFilename: omit.From(media.OriginalFilename),
			Extension:        omit.From(media.Extension),
			MimeType:         omit.From(media.MimeType),
			Size:             omit.From(media.Size),
			CreatedAt:        omit.From(media.CreatedAt),
			UpdatedAt:        omit.From(media.UpdatedAt),
		},
		im.Returning("*"),
	)
	d, err := q.One(ctx, exec)
	d, err = OptionalRow(d, err)
	if err != nil {
		return nil, err
	}
	return d, nil
}
