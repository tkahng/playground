package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

type OtpInput struct {
	Token string           `query:"token" json:"token" required:"true"`
	Type  shared.TokenType `query:"type" json:"type" required:"true"`
}

type oauthLoginResponse struct {
	Status int
	Url    string `header:"Location"`
	Cookie string `header:"Set-Cookie"`
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
	switch input.Type {
	case shared.VerificationTokenType:
		_, err := api.app.VerifyAndUseVerificationToken(ctx, db, input.Token)
		if err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, huma.Error400BadRequest(fmt.Sprintf("Invalid token type. only verification_token is supported. got %v", input.Type))
	}
}
