package apis

import (
	"bytes"
	"context"
	"io"
	"path"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) UploadMedia(ctx context.Context, input *struct {
	RawBody huma.MultipartFormFiles[struct {
		Files []huma.FormFile `form:"files" required:"false" description:"Files to upload"`
		Urls  []string        `form:"urls" format:"uri" required:"false" description:"Urls to upload"  minItems:"1" maxItems:"10" nullable:"false"`
	}] `contentType:"multipart/form-data"`
}) (*struct{}, error) {
	db := api.app.Db()
	user := contextstore.GetContextUserInfo(ctx)
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	formData := input.RawBody.Data()

	if formData.Files != nil {
		for _, file := range formData.Files {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, file.File); err != nil {
				return nil, err
			}

			dto, err := api.app.Fs().PutFileFromBytes(ctx, buf.Bytes(), file.Filename)
			if err != nil {
				return nil, err
			}
			_, err = queries.CreateMedia(ctx, db, &models.Medium{
				UserID:           &user.User.ID,
				Disk:             dto.Disk,
				Directory:        dto.Directory,
				Filename:         dto.Filename,
				OriginalFilename: dto.OriginalName,
				Extension:        dto.Extension,
				MimeType:         dto.MimeType,
				Size:             dto.Size,
			})
			if err != nil {
				return nil, err
			}

		}
	}

	if formData.Urls != nil {
		for _, url := range formData.Urls {
			dto, err := api.app.Fs().PutNewFileFromURL(ctx, url)
			if err != nil {
				return nil, err
			}
			_, err = queries.CreateMedia(ctx, db, &models.Medium{
				UserID:           &user.User.ID,
				Disk:             dto.Disk,
				Directory:        dto.Directory,
				Filename:         dto.Filename,
				OriginalFilename: dto.OriginalName,
				Extension:        dto.Extension,
				MimeType:         dto.MimeType,
				Size:             dto.Size,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

func (api *Api) GetMedia(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true" description:"Id of the media"`
}) (*shared.Media, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	media, err := queries.FindMediaByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	url, err := api.app.Fs().GeneratePresignedURL(ctx, media.Disk, path.Join(media.Directory, media.Filename))
	if err != nil {
		return nil, err
	}
	return &shared.Media{
		ID:        media.ID,
		Filename:  media.Filename,
		URL:       url,
		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,
	}, nil
}

func (api *Api) MediaList(ctx context.Context, input *shared.MediaListParams) (*shared.PaginatedOutput[*shared.Media], error) {
	db := api.app.Db()
	medias, err := queries.ListMedia(ctx, db, input)
	if err != nil {
		return nil, err
	}
	var data []*shared.Media
	for _, media := range medias {
		url, err := api.app.Fs().GeneratePresignedURL(ctx, media.Disk, path.Join(media.Directory, media.Filename))
		if err != nil {
			return nil, err
		}
		data = append(data, &shared.Media{
			ID:        media.ID,
			Filename:  media.Filename,
			URL:       url,
			CreatedAt: media.CreatedAt,
			UpdatedAt: media.UpdatedAt,
		})
	}
	count, err := queries.CountMedia(ctx, db, &input.MediaListFilter)
	if err != nil {
		return nil, err
	}

	return &shared.PaginatedOutput[*shared.Media]{
		Body: shared.PaginatedResponse[*shared.Media]{
			Data: data,
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
