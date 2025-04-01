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
	// http://127.0.0.1:8080/auth/callback
	// huma.Register(api, appApi.AuthMethodsOperation("/auth/methods"), appApi.AuthMethods)
	huma.Register(api, appApi.SignupOperation("/auth/signup"), appApi.SignUp)
	huma.Register(api, appApi.SigninOperation("/auth/signin"), appApi.SignIn)
	huma.Register(api, appApi.MeOperation("/auth/me"), appApi.Me)
	huma.Register(api, appApi.RefreshTokenOperation("/auth/refresh-token"), appApi.RefreshToken)

	huma.Register(api, appApi.VerifyOperation("/auth/verify"), appApi.Verify)
	huma.Register(api, appApi.RequestPasswordResetOperation("/auth/request-password-reset"), appApi.RequestPasswordReset)
	huma.Register(api, appApi.ConfirmPasswordResetOperation("/auth/confirm-password-reset"), appApi.ConfirmPasswordReset)

	huma.Register(api, appApi.OAuth2CallbackGetOperation("/auth/callback"), appApi.OAuth2CallbackGet)
	huma.Register(api, appApi.OAuth2CallbackPostOperation("/auth/callback"), appApi.OAuth2CallbackPost)
	huma.Register(api, appApi.OAuth2AuthorizationUrlOperation("/auth/authorization-url"), appApi.OAuth2AuthorizationUrl)

	//  admin routes
	adminGroup := huma.NewGroup(api, "/admin")
	//  admin middleware
	adminGroup.UseMiddleware(CheckRolesMiddleware(api, "superuser"))
	//  admin user list
	huma.Register(adminGroup, appApi.AdminUsersOperation("/users"), appApi.AdminUsers)
	//  admin user create
	huma.Register(adminGroup, appApi.AdminUsersCreateOperation("/users"), appApi.AdminUsersCreate)
	//  admin user delete
	huma.Register(adminGroup, appApi.AdminUsersDeleteOperation("/users/{id}"), appApi.AdminUsersDelete)
	//  admin user update
	huma.Register(adminGroup, appApi.AdminUsersUpdateOperation("/users/{id}"), appApi.AdminUsersUpdate)
	//  admin user update password
	huma.Register(adminGroup, appApi.AdminUsersUpdatePasswordOperation("/users/{id}/password"), appApi.AdminUsersUpdatePassword)
	//  admin user update roles
	huma.Register(adminGroup, appApi.AdminUserRolesUpdateOperation("/users/{id}/roles"), appApi.AdminUserRolesUpdate)
	// admin roles
	huma.Register(adminGroup, appApi.AdminRolesOperation("/roles"), appApi.AdminRoles)
	// admin roles create
	huma.Register(adminGroup, appApi.AdminRolesCreateOperation("/roles"), appApi.AdminRolesCreate)
	// admin roles update
	huma.Register(adminGroup, appApi.AdminRolesUpdateOperation("/roles/{id}"), appApi.AdminRolesUpdate)
	// admin roles delete
	huma.Register(adminGroup, appApi.AdminRolesDeleteOperation("/roles/{id}"), appApi.AdminRolesDelete)
	// admin permissions list
	huma.Register(adminGroup, appApi.AdminPermissionsOperation("/permissions"), appApi.AdminPermissions)
	// admin permissions create
	huma.Register(adminGroup, appApi.AdminPermissionsCreateOperation("/permissions"), appApi.AdminPermissionsCreate)
	// admin permissions update
	huma.Register(adminGroup, appApi.AdminPermissionsUpdateOperation("/permissions/{id}"), appApi.AdminPermissionsUpdate)
	// admin permissions delete
	huma.Register(adminGroup, appApi.AdminPermissionsDeleteOperation("/permissions/{id}"), appApi.AdminPermissionsDelete)

}

func AddRoutes(api huma.API, app core.App) {
	BindMiddlewares(api, app)
	BindApis(api, app)
}
