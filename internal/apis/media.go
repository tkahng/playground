package apis

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) UploadMediaOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "upload-media",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Upload media",
		Description: "Upload media",
		Tags:        []string{"Media"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) UploadMedia(ctx context.Context, input *struct {
	RawBody huma.MultipartFormFiles[struct {
		Files []huma.FormFile `form:"files" required:"false" description:"Files to upload"`
		Urls  []string        `form:"urls" format:"uri" required:"false" description:"Urls to upload"  minItems:"1" maxItems:"10" nullable:"false"`
	}] `contentType:"multipart/form-data"`
}) (*struct{}, error) {
	db := api.app.Db()
	user := core.GetContextUserClaims(ctx)
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

			dto, err := api.app.Fs().NewFileFromBytes(buf.Bytes(), file.Filename)
			if err != nil {
				return nil, err
			}
			_, err = repository.CreateMedia(ctx, db, &models.Medium{
				UserID:           null.From(user.User.ID),
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
			dto, err := api.app.Fs().NewFileFromURL(ctx, url)
			if err != nil {
				return nil, err
			}
			_, err = repository.CreateMedia(ctx, db, &models.Medium{
				UserID:           null.From(user.User.ID),
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

type MediaOuput struct {
	ID        uuid.UUID `json:"id" db:"id" format:"uuid"`
	Filename  string    `json:"filename" db:"filename"`
	URL       string    `json:"url" db:"url" format:"uri"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (api *Api) GetMediaOperation(path string /** /media/:id */) huma.Operation {
	return huma.Operation{
		OperationID: "get-media",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Get media",
		Description: "Get media",
		Tags:        []string{"Media"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) GetMedia(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true" description:"Id of the media"`
}) (*MediaOuput, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	media, err := repository.FindMediaByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	url, err := api.app.Fs().GeneratePresignedURL(ctx, media.Disk, path.Join(media.Directory, media.Filename))
	if err != nil {
		return nil, err
	}
	return &MediaOuput{
		ID:        media.ID,
		Filename:  media.Filename,
		URL:       url,
		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,
	}, nil
}
