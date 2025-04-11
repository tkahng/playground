package apis

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/aarondl/opt/null"
	"github.com/danielgtaylor/huma/v2"
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
