package apis

import (
	"bytes"
	"context"
	"io"
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
			publicURL := dto.PublicURL
			_, err = api.App().Adapter().Media().CreateMedia(ctx, &models.Medium{
				UserID:           &user.User.ID,
				StorageKey:       dto.StorageKey,
				PublicURL:        &publicURL,
				MimeType:         dto.MimeType,
				Size:             dto.Size,
				OriginalFilename: dto.OriginalName,
				Extension:        dto.Extension,
				Disk:             dto.Disk,
				Directory:        dto.Directory,
				Filename:         dto.Filename,
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
			publicURL := dto.PublicURL
			_, err = api.App().Adapter().Media().CreateMedia(ctx, &models.Medium{
				UserID:           &user.User.ID,
				StorageKey:       dto.StorageKey,
				PublicURL:        &publicURL,
				MimeType:         dto.MimeType,
				Size:             dto.Size,
				OriginalFilename: dto.OriginalName,
				Extension:        dto.Extension,
				Disk:             dto.Disk,
				Directory:        dto.Directory,
				Filename:         dto.Filename,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

type Media struct {
	ID               uuid.UUID `json:"id" format:"uuid"`
	StorageKey       string    `json:"storage_key"`
	URL              string    `json:"url" format:"uri"`
	MimeType         string    `json:"mime_type"`
	Size             int64     `json:"size"`
	OriginalFilename string    `json:"original_filename"`
	AltText          *string   `json:"alt_text" nullable:"true"`
	Width            *int      `json:"width" nullable:"true"`
	Height           *int      `json:"height" nullable:"true"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (api *Api) mediaFromModel(m *models.Medium) *Media {
	url := ""
	if m.PublicURL != nil {
		url = *m.PublicURL
	} else {
		url = api.App().Fs().PublicURL(m.StorageKey)
	}
	return &Media{
		ID:               m.ID,
		StorageKey:       m.StorageKey,
		URL:              url,
		MimeType:         m.MimeType,
		Size:             m.Size,
		OriginalFilename: m.OriginalFilename,
		AltText:          m.AltText,
		Width:            m.Width,
		Height:           m.Height,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

type GetMediaOutput struct {
	Body *Media
}

func (api *Api) GetMedia(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true" description:"Id of the media"`
}) (*GetMediaOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	media, err := api.App().Adapter().Media().FindMediaByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, huma.Error404NotFound("media not found")
	}
	return &GetMediaOutput{Body: api.mediaFromModel(media)}, nil
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
	count, err := api.App().Adapter().Media().CountMedia(ctx, filter)
	if err != nil {
		return nil, err
	}

	data := make([]*Media, len(medias))
	for i, m := range medias {
		data[i] = api.mediaFromModel(m)
	}

	return &ApiPaginatedOutput[*Media]{
		Body: ApiPaginatedResponse[*Media]{
			Data: data,
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
