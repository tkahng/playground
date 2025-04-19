package apis

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
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
		Page shared.OmittableNullable[string] `query:"page" required:"false"`
	}) (*IndexOutput, error) {
		fmt.Println("input", input)
		return &IndexOutput{
			Body: IndexOutputBody{
				Access: "public",
			},
		}, nil
	})
	// http://127.0.0.1:8080/auth/callback
	// huma.Register(api, appApi.AuthMethodsOperation("/auth/methods"), appApi.AuthMethods)
	// protected test routes -----------------------------------------------------------

	huma.Register(api, appApi.ApiProtectedBasicPermissionOperation("/protected/basic-permission"), appApi.ApiProtectedBasicPermission)
	huma.Register(api, appApi.ApiProtectedProPermissionOperation("/protected/pro-permission"), appApi.ApiProtectedProPermission)
	huma.Register(api, appApi.ApiProtectedAdvancedPermissionOperation("/protected/advanced-permission"), appApi.ApiProtectedAdvancedPermission)

	huma.Register(api, appApi.SignupOperation("/auth/signup"), appApi.SignUp)
	huma.Register(api, appApi.SigninOperation("/auth/signin"), appApi.SignIn)
	huma.Register(api, appApi.MeOperation("/auth/me"), appApi.Me)
	huma.Register(api, appApi.RefreshTokenOperation("/auth/refresh-token"), appApi.RefreshToken)

	huma.Register(api, appApi.VerifyOperation("/auth/verify"), appApi.Verify)
	huma.Register(api, appApi.VerifyPostOperation("/auth/verify"), appApi.VerifyPost)
	huma.Register(api, appApi.RequestPasswordResetOperation("/auth/request-password-reset"), appApi.RequestPasswordReset)
	huma.Register(api, appApi.ConfirmPasswordResetOperation("/auth/confirm-password-reset"), appApi.ConfirmPasswordReset)

	huma.Register(api, appApi.OAuth2CallbackGetOperation("/auth/callback"), appApi.OAuth2CallbackGet)
	huma.Register(api, appApi.OAuth2CallbackPostOperation("/auth/callback"), appApi.OAuth2CallbackPost)
	huma.Register(api, appApi.OAuth2AuthorizationUrlOperation("/auth/authorization-url"), appApi.OAuth2AuthorizationUrl)
	// authenticated routes -----------------------------------------------------------
	authenticatedGroup := huma.NewGroup(api)
	// ---- Upload File
	huma.Register(authenticatedGroup, appApi.UploadMediaOperation("/media"), appApi.UploadMedia)
	// ---- Get Media
	huma.Register(authenticatedGroup, appApi.GetMediaOperation("/media/{id}"), appApi.GetMedia)
	// ---- Get Media List
	huma.Register(authenticatedGroup, appApi.MedialListOperation("/media"), appApi.MediaList)
	// ---- notifications
	sse.Register(authenticatedGroup, appApi.NotificationsSseOperation("/notifications/sse"), map[string]any{
		// Mapping of event type name to Go struct for that event.
		"message": models.Notification{},
	}, appApi.NotificationsSsefunc)
	// ---- task routes -------------------------------------------------------------------------------------------------
	taskGroup := huma.NewGroup(api)
	taskGroup.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		db := app.Db()
		rawCtx := ctx.Context()
		taskId := ctx.Param("task-id")
		fmt.Println("taskId", taskId)
		if taskId == "" {
			next(ctx)
			return
		}
		id, err := uuid.Parse(taskId)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "invalid task id", err)
			return
		}
		task, err := repository.FindTaskByID(rawCtx, db, id)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting task", err)
			return
		}
		if task == nil {
			huma.WriteErr(api, ctx, http.StatusNotFound, "task not found")
			return
		}
		user := core.GetContextUserClaims(rawCtx)
		if user == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		if task.UserID != user.User.ID {
			if slices.Contains(user.Permissions, "superuser") {
				next(ctx)
				return
			}
			huma.WriteErr(api, ctx, http.StatusForbidden, "task not found")
			return
		}
		// if
		next(ctx)
	})
	// task list
	huma.Register(taskGroup, appApi.TaskListOperation("/tasks"), appApi.TaskList)
	// task create
	// huma.Register(taskGroup, appApi.TaskCreateOperation("/task"), appApi.TaskCreate)
	// task update
	huma.Register(taskGroup, appApi.TaskUpdateOperation("/tasks/{task-id}"), appApi.TaskUpdate)
	// task position
	huma.Register(taskGroup, appApi.UpdateTaskPositionOperation("/tasks/{task-id}/position"), appApi.UpdateTaskPosition)
	// task position status
	huma.Register(taskGroup, appApi.UpdateTaskPositionStatusOperation("/tasks/{task-id}/position-status"), appApi.UpdateTaskPositionStatus)
	// // task delete
	huma.Register(taskGroup, appApi.TaskDeleteOperation("/tasks/{task-id}"), appApi.TaskDelete)
	// // task get
	huma.Register(taskGroup, appApi.TaskGetOperation("/tasks/{task-id}"), appApi.TaskGet)

	// task project routes -------------------------------------------------------------------------------------------------
	taskProjectGroup := huma.NewGroup(api)
	// task project list
	huma.Register(taskProjectGroup, appApi.TaskProjectListOperation("/task-projects"), appApi.TaskProjectList)
	// task project create
	huma.Register(taskProjectGroup, appApi.TaskProjectCreateOperation("/task-projects"), appApi.TaskProjectCreate)
	// task project create with ai
	huma.Register(taskProjectGroup, appApi.TaskProjectCreateWithAiOperation("/task-projects/ai"), appApi.TaskProjectCreateWithAi)
	// task project update
	huma.Register(taskProjectGroup, appApi.TaskProjectUpdateOperation("/task-projects/{task-project-id}"), appApi.TaskProjectUpdate)
	// // task project delete
	huma.Register(taskProjectGroup, appApi.TaskProjectDeleteOperation("/task-projects/{task-project-id}"), appApi.TaskProjectDelete)
	// // task project get
	huma.Register(taskProjectGroup, appApi.TaskProjectGetOperation("/task-projects/{task-project-id}"), appApi.TaskProjectGet)
	// task project tasks create
	huma.Register(taskProjectGroup, appApi.TaskProjectTasksCreateOperation("/task-projects/{task-project-id}/tasks"), appApi.TaskProjectTasksCreate)

	// stripe routes -------------------------------------------------------------------------------------------------
	stripeGroup := huma.NewGroup(api, "/stripe")
	// stripe my subscriptions
	huma.Register(stripeGroup, appApi.MyStripeSubscriptionsOperation("/my-subscriptions"), appApi.MyStripeSubscriptions)
	// stripe webhook
	huma.Register(stripeGroup, appApi.StripeWebhookOperation("/webhook"), appApi.StripeWebhook)
	// stripe products with prices
	huma.Register(stripeGroup, appApi.StripeProductsWithPricesOperation("/products"), appApi.StripeProductsWithPrices)
	// stripe billing portal
	huma.Register(stripeGroup, appApi.StripeBillingPortalOperation("/billing-portal"), appApi.StripeBillingPortal)
	//  stripe checkout session
	huma.Register(stripeGroup, appApi.StripeCheckoutSessionOperation("/checkout-session"), appApi.StripeCheckoutSession)
	//  stripe get checkout session by checkoutSessionId
	huma.Register(stripeGroup, appApi.StripeCheckoutSessionGetOperation("/checkout-session/{checkoutSessionId}"), appApi.StripeCheckoutSessionGet)

	//  admin routes ----------------------------------------------------------------------------
	adminGroup := huma.NewGroup(api, "/admin")
	//  admin middleware
	adminGroup.UseMiddleware(CheckPermissionsMiddleware(api, "superuser"))
	//  admin user list
	huma.Register(adminGroup, appApi.AdminUsersOperation("/users"), appApi.AdminUsers)
	// admin user get
	huma.Register(adminGroup, appApi.AdminUsersGetOperation("/users/{id}"), appApi.AdminUsersGet)
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
	// admin user create roles
	huma.Register(adminGroup, appApi.AdminUserRolesCreateOperation("/users/{id}/roles"), appApi.AdminUserRolesCreate)
	// admin user delete roles
	huma.Register(adminGroup, appApi.AdminUserRolesDeleteOperation("/users/{userId}/roles/{roleId}"), appApi.AdminUserRolesDelete)
	// admin user permission source list
	huma.Register(adminGroup, appApi.AdminUserPermissionSourceListOperation("/users/{userId}/permissions"), appApi.AdminUserPermissionSourceList)
	// admin user permissions create
	huma.Register(adminGroup, appApi.AdminUserPermissionsCreateOperation("/users/{userId}/permissions"), appApi.AdminUserPermissionsCreate)
	// admin user permissions delete
	huma.Register(adminGroup, appApi.AdminUserPermissionsDeleteOperation("/users/{userId}/permissions/{permissionId}"), appApi.AdminUserPermissionsDelete)
	// admin roles
	huma.Register(adminGroup, appApi.AdminRolesOperation("/roles"), appApi.AdminRolesList)
	// admin roles create
	huma.Register(adminGroup, appApi.AdminRolesCreateOperation("/roles"), appApi.AdminRolesCreate)
	// admin roles update
	huma.Register(adminGroup, appApi.AdminRolesUpdateOperation("/roles/{id}"), appApi.AdminRolesUpdate)
	// admin role get
	huma.Register(adminGroup, appApi.AdminRolesGetOperation("/roles/{id}"), appApi.AdminRolesGet)
	// admin roles update permissions
	huma.Register(adminGroup, appApi.AdminRolesUpdatePermissionsOperation("/roles/{id}/permissions"), appApi.AdminRolesUpdatePermissions)
	// admin roles create permissions
	huma.Register(adminGroup, appApi.AdminRolesCreatePermissionsOperation("/roles/{id}/permissions"), appApi.AdminRolesCreatePermissions)
	// admin roles delete permissions
	huma.Register(adminGroup, appApi.AdminRolesDeletePermissionsOperation("/roles/{roleId}/permissions/{permissionId}"), appApi.AdminRolesDeletePermissions)
	// admin roles delete
	huma.Register(adminGroup, appApi.AdminRolesDeleteOperation("/roles/{id}"), appApi.AdminRolesDelete)
	// admin permissions list
	huma.Register(adminGroup, appApi.AdminPermissionsListOperation("/permissions"), appApi.AdminPermissionsList)
	// admin permissions create
	huma.Register(adminGroup, appApi.AdminPermissionsCreateOperation("/permissions"), appApi.AdminPermissionsCreate)
	// admin permissions get
	huma.Register(adminGroup, appApi.AdminPermissionsGetOperation("/permissions/{id}"), appApi.AdminPermissionsGet)
	// admin permissions update
	huma.Register(adminGroup, appApi.AdminPermissionsUpdateOperation("/permissions/{id}"), appApi.AdminPermissionsUpdate)
	// admin permissions delete
	huma.Register(adminGroup, appApi.AdminPermissionsDeleteOperation("/permissions/{id}"), appApi.AdminPermissionsDelete)

	// admin stripe subscriptions
	huma.Register(adminGroup, appApi.AdminStripeSubscriptionsOperation("/subscriptions"), appApi.AdminStripeSubscriptions)

}

func AddRoutes(api huma.API, app core.App) {
	BindMiddlewares(api, app)
	BindApis(api, app)
}
