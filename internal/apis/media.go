package apis

import (
	"bytes"
	"context"
	"io"
	"path"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/utils"
)

func (api *Api) UploadMedia(ctx context.Context, input *struct {
	RawBody huma.MultipartFormFiles[struct {
		Files []huma.FormFile `form:"files" required:"false" description:"Files to upload"`
		Urls  []string        `form:"urls" format:"uri" required:"false" description:"Urls to upload"  minItems:"1" maxItems:"10" nullable:"false"`
	}] `contentType:"multipart/form-data"`
}) (*struct{}, error) {
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

			dto, err := api.App().Fs().PutFileFromBytes(ctx, buf.Bytes(), file.Filename)
			if err != nil {
				return nil, err
			}
			_, err = api.App().Adapter().Media().CreateMedia(ctx, &models.Medium{
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
			dto, err := api.App().Fs().PutNewFileFromURL(ctx, url)
			if err != nil {
				return nil, err
			}
			_, err = api.App().Adapter().Media().CreateMedia(ctx, &models.Medium{
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

type Media struct {
	ID        uuid.UUID `json:"id" db:"id" format:"uuid"`
	Filename  string    `json:"filename" db:"filename"`
	URL       string    `json:"url" db:"url" format:"uri"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (api *Api) GetMedia(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true" description:"Id of the media"`
}) (*Media, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	media, err := api.App().Adapter().Media().FindMediaByID(ctx, id)
	if err != nil {
		return nil, err
	}
	url, err := api.App().Fs().GeneratePresignedURL(ctx, media.Disk, path.Join(media.Directory, media.Filename))
	if err != nil {
		return nil, err
	}
	return &Media{
		ID:        media.ID,
		Filename:  media.Filename,
		URL:       url,
		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,
	}, nil
}

type MediaListFilter struct {
	PaginatedInput
	SortParams
	Q       string   `query:"q,omitempty" required:"false"`
	UserIds []string `query:"user_ids,omitempty" format:"uuid" required:"false"`
}

func (api *Api) MediaList(ctx context.Context, input *MediaListFilter) (*ApiPaginatedOutput[*Media], error) {
	filter := &stores.MediaListFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.Q = input.Q
	filter.UserIds = utils.ParseValidUUIDs(input.UserIds...)

	medias, err := api.App().Adapter().Media().FindMedia(ctx, filter)
	if err != nil {
		return nil, err
	}
	var data []*Media
	for _, media := range medias {
		url, err := api.App().Fs().GeneratePresignedURL(ctx, media.Disk, path.Join(media.Directory, media.Filename))
		if err != nil {
			return nil, err
		}
		data = append(data, &Media{
			ID:        media.ID,
			Filename:  media.Filename,
			URL:       url,
			CreatedAt: media.CreatedAt,
			UpdatedAt: media.UpdatedAt,
		})
	}
	count, err := api.App().Adapter().Media().CountMedia(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ApiPaginatedOutput[*Media]{
		Body: ApiPaginatedResponse[*Media]{
			Data: data,
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
