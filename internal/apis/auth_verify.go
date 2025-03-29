package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

type VerifyTokenType string

const (
	VerifyTokenTypeEmail VerifyTokenType = "verification_token"
	// VerifyTokenTypePassword VerifyTokenType = "password_reset_token"
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
		OperationID: "verify",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Verify",
		Description: "Verify",
		Tags:        []string{"Auth", "Verify"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		// Security: []map[string][]string{
		// 	{shared.BearerAuthSecurityKey: {}},
		// },
	}
}

func (api *Api) Verify(ctx context.Context, input *OtpInput) (*oauthLoginResponse, error) {
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
