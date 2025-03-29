package apis

import (
	"context"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/auth"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
	"golang.org/x/oauth2"
)

type oauth2Response struct {
	Providers []providerInfo `json:"providers"`
	Enabled   bool           `json:"enabled"`
}

type providerInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	State       string `json:"state"`
	AuthURL     string `json:"authURL"`

	// @todo
	// deprecated: use AuthURL instead
	// AuthUrl will be removed after dropping v0.22 support
	AuthUrl string `json:"authUrl"`

	// technically could be omitted if the provider doesn't support PKCE,
	// but to avoid breaking existing typed clients we'll return them as empty string
	CodeVerifier        string `json:"codeVerifier"`
	CodeChallenge       string `json:"codeChallenge"`
	CodeChallengeMethod string `json:"codeChallengeMethod"`
}

// type passwordResponse struct {
// 	IdentityFields []string `json:"identityFields"`
// 	Enabled        bool     `json:"enabled"`
// }

type authMethodsResponse struct {
	// Password passwordResponse `json:"password"`
	OAuth2 oauth2Response `json:"oauth2"`
	// MFA      mfaResponse      `json:"mfa"`
	// OTP      otpResponse      `json:"otp"`

}

func (api *Api) AuthMethodsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "auth-methods",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "auth-methods",
		Description: "auth-methods",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AuthMethods(context.Context, *struct{}) (*struct{}, error) {
	// collection, err := findAuthCollection(e)
	collection := api.app.Settings().Auth
	// if err != nil {
	// 	return err
	// }

	result := authMethodsResponse{
		// Password: passwordResponse{
		// 	IdentityFields: make([]string, 0, len(collection.PasswordAuth.IdentityFields)),
		// },
		OAuth2: oauth2Response{
			Providers: make([]providerInfo, 0, len(collection.OAuth2.Providers)),
		},
		// OTP: otpResponse{
		// 	Enabled: collection.OTP.Enabled,
		// },
		// MFA: mfaResponse{
		// 	Enabled: collection.MFA.Enabled,
		// },
	}

	// if collection.PasswordAuth.Enabled {
	// 	result.Password.Enabled = true
	// 	result.Password.IdentityFields = collection.PasswordAuth.IdentityFields
	// }

	// if collection.OTP.Enabled {
	// 	result.OTP.Duration = collection.OTP.Duration
	// }

	// if collection.MFA.Enabled {
	// 	result.MFA.Duration = collection.MFA.Duration
	// }

	if !collection.OAuth2.Enabled {
		// result.fillLegacyFields()

		// return e.JSON(http.StatusOK, result)
	}

	result.OAuth2.Enabled = true

	for _, config := range collection.OAuth2.Providers {
		provider, err := config.InitProvider()
		if err != nil {
			log.Println("Failed to setup OAuth2 provider", "name", config.Name, "error", err.Error())
			// e.App.Logger().Debug(
			// 	"Failed to setup OAuth2 provider",
			// 	slog.String("name", config.Name),
			// 	slog.String("error", err.Error()),
			// )
			continue // skip provider
		}

		info := providerInfo{
			Name:        config.Name,
			DisplayName: provider.DisplayName(),
			State:       security.RandomString(30),
		}

		if info.DisplayName == "" {
			info.DisplayName = config.Name
		}

		urlOpts := []oauth2.AuthCodeOption{}

		// custom providers url options
		switch config.Name {
		case auth.NameApple:
			// see https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_js/incorporating_sign_in_with_apple_into_other_platforms#3332113
			urlOpts = append(urlOpts, oauth2.SetAuthURLParam("response_mode", "form_post"))
		}

		if provider.PKCE() {
			info.CodeVerifier = security.RandomString(43)
			info.CodeChallenge = security.S256Challenge(info.CodeVerifier)
			info.CodeChallengeMethod = "S256"
			urlOpts = append(urlOpts,
				oauth2.SetAuthURLParam("code_challenge", info.CodeChallenge),
				oauth2.SetAuthURLParam("code_challenge_method", info.CodeChallengeMethod),
			)
		}

		info.AuthURL = provider.BuildAuthURL(
			info.State,
			urlOpts...,
		) + "&redirect_uri=" // empty redirect_uri so that users can append their redirect url

		info.AuthUrl = info.AuthURL

		result.OAuth2.Providers = append(result.OAuth2.Providers, info)
	}

	return &struct{}{}, nil
}
