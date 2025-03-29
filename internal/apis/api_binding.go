package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
)

// import (
// 	"net/http"

// 	"github.com/danielgtaylor/huma/v2"
// 	"github.com/tkahng/authgo/internal/core"
// 	"github.com/tkahng/authgo/internal/middlewares"
// )

func InitApiConfig() huma.Config {
	config := huma.DefaultConfig("My API", "1.0.0")
	config.Servers = []*huma.Server{{URL: "http://localhost:8080"}}
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		shared.BearerAuthSecurityKey: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	return config
}

func BindMiddlewares(api huma.API, app core.App) {
	api.UseMiddleware(AuthMiddleware(api, app))
	api.UseMiddleware(RequireAuthMiddleware(api))
}

// OmittableNullable is a field which can be omitted from the input,
// set to `null`, or set to a value. Each state is tracked and can
// be checked for in handling code.
type OmittableNullable[T any] struct {
	Sent  bool
	Null  bool
	Value T
}

// UnmarshalJSON unmarshals this value from JSON input.
func (o *OmittableNullable[T]) UnmarshalJSON(b []byte) error {
	if len(b) > 0 {
		o.Sent = true
		if bytes.Equal(b, []byte("null")) {
			o.Null = true
			return nil
		}
		return json.Unmarshal(b, &o.Value)
	}
	return nil
}

// Schema returns a schema representing this value on the wire.
// It returns the schema of the contained type.
func (o OmittableNullable[T]) Schema(r huma.Registry) *huma.Schema {
	return r.Schema(reflect.TypeOf(o.Value), true, "")
}

type IndexOutputBody struct {
	Access string `json:"access"`
}

type IndexOutput struct {
	Body IndexOutputBody `json:"body"`
}

func BindApis(api huma.API, app core.App) {
	appApi := &Api{
		app: app,
	}
	huma.Get(api, "/", func(ctx context.Context, input *struct {
		Page OmittableNullable[string] `query:"page" required:"false"`
	}) (*IndexOutput, error) {
		return &IndexOutput{
			Body: IndexOutputBody{
				Access: "public",
			},
		}, nil
	})
	huma.Register(api, appApi.SignupOperation("/auth/signup"), appApi.SignUp)
	huma.Register(api, appApi.SigninOperation("/auth/signin"), appApi.SignIn)
	huma.Register(api, appApi.MeOperation("/auth/me"), appApi.Me)
	huma.Register(api, appApi.RefreshTokenOperation("/auth/refresh-token"), appApi.RefreshToken)

	huma.Register(api, appApi.VerifyOperation("/auth/verify"), appApi.Verify)
	huma.Register(api, appApi.RequestPasswordResetOperation("/auth/request-password-reset"), appApi.RequestPasswordReset)
	huma.Register(api, appApi.ConfirmPasswordResetOperation("/auth/confirm-password-reset"), appApi.ConfirmPasswordReset)

	huma.Register(api, appApi.OauthCallbackGetOperation("/auth/oauth/callback"), appApi.OauthCallbackGet())

	huma.Register(api, appApi.AdminUsersOperation("/admin/users"), appApi.AdminUsers)

	huma.Register(api, appApi.AppSettingsOperation("/app/settings"), appApi.AppSettings)

	// bindUsersApi(api, app)
	// bindStripeApi(api, app)
	// bindViewAPI(api, app)
}

func AddRoutes(api huma.API, app core.App) {
	BindMiddlewares(api, app)
	BindApis(api, app)
}

// type SetCookieOutput struct {
// 	SetCookie []*http.Cookie `header:"Set-Cookie"`
// }

// type CookieInput struct {
// 	AccessToken  *http.Cookie `cookie:"access_token" required:"false"`
// 	RefreshToken *http.Cookie `cookie:"refresh_token" required:"false"`
// }
