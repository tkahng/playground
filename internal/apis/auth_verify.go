package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type OtpInput struct {
	Token string `query:"token" json:"token" required:"true"`
}

func (api *Api) VerifyOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "verify-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Verify",
		Description: "Verify",
		Tags:        []string{"Auth", "Verify"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
	}
}

func (api *Api) Verify(ctx context.Context, input *OtpInput) (*struct{}, error) {
	return Verify(api, ctx, input)
}

func (api *Api) VerifyPostOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "verify-post",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Verify",
		Description: "Verify",
		Tags:        []string{"Auth", "Verify"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
	}
}

func (h *Api) VerifyPost(ctx context.Context, input *struct{ Body *OtpInput }) (*struct{}, error) {
	return Verify(h, ctx, input.Body)
}

func Verify(api *Api, ctx context.Context, input *OtpInput) (*struct{}, error) {
	db := api.app.Db()
	action := api.app.NewAuthActions(db)
	err := action.HandleVerificationToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	return nil, nil

}
