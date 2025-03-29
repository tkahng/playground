package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type OAuth2CallbackInput struct {
	Code  string `json:"code" query:"code" required:"true"`
	State string `json:"state" query:"state" required:"true"`
	// Provider db.AuthProviders `json:"provider" path:"provider"`
}

// func (m *OAuth2CallbackInput) Resolve(ctx huma.Context) []error {
// 	// Get request info you don't normally have access to.
// 	if m.Provider == db.AuthProvidersEmail || m.Provider == db.AuthProvidersCredentials {
// 		return []error{errors.New("invalid provider")}
// 	}
// 	if m.Code == "" || m.Provider == "" {
// 		return []error{errors.New("missing provider or code")}
// 	}

// 	return nil
// }

func (h *Api) OAuth2CallbackOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "oauth2-callback",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "OAuth2 callback",
		Description: "Count the number of colors for all themes",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
		// Security: []map[string][]string{
		// 	middleware.BearerAuthSecurity("colors:read"),
		// },
	}
}
func (h *Api) Oatuh2Callback(ctx context.Context, input *OAuth2CallbackInput) (*AuthenticatedResponse, error) {
	// provider, err := auth.NewProviderByName(input.Provider, &h.app.Cfg().OAuth)
	// if err != nil {
	// 	log.Println(err)

	// 	return nil, fmt.Errorf("Error at Oatuh2Callback: %w", err)
	// }

	// ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	// defer cancel()

	// provider.SetContext(ctx)

	// var opts []oauth2.AuthCodeOption

	// // fetch token
	// token, err := provider.FetchToken(input.Code, opts...)
	// if err != nil {
	// 	log.Println(err)
	// 	return nil, fmt.Errorf("Error at Oatuh2Callback: %w", err)
	// }

	// authUser, err := provider.FetchAuthUser(token)
	// if err != nil {
	// 	log.Println(err)
	// 	return nil, fmt.Errorf("Error at Oatuh2Callback: %w", err)
	// }

	// params := &core.NewUserParams{
	// 	Email:             authUser.Email,
	// 	Name:              &authUser.Username,
	// 	EmailVerifiedAt:   &authUser.Expiry,
	// 	Provider:          input.Provider,
	// 	Key:               authUser.RefreshToken,
	// 	Type:              db.AuthProviderTypesOauth,
	// 	ProviderAccountID: authUser.Id,
	// }

	// user, err := h.app.Signup(ctx, params)
	// if err != nil {
	// 	return nil, fmt.Errorf("Error at Oatuh2Callback: %w", err)
	// }
	// return TokenDtoFromUserWithApp(ctx, h.app, user, uuid.NewString())
	return nil, nil
}
