package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

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

	checkTaskOwnerMiddleware := CheckTaskOwnerMiddleware(api, app)

	//  public list of permissions -----------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "permissions-list",
			Method:      http.MethodGet,
			Path:        "/permissions",
			Summary:     "permissions list",
			Description: "List of permissions",
			Tags:        []string{"Permissions"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.PermissionsList,
	)
	// protected test routes -----------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "api-protected",
			Method:      http.MethodGet,
			Path:        "/protected/{permission-name}",
			Summary:     "Api protected",
			Description: "Api protected",
			Tags:        []string{"Protected"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.ApiProtected,
	)

	// signup -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signup",
			Method:      http.MethodPost,
			Path:        "/auth/signup",
			Summary:     "Sign up",
			Description: "Count the number of colors for all themes",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.SignUp,
	)
	// signin -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signin",
			Method:      http.MethodPost,
			Path:        "/auth/signin",
			Summary:     "Sign in",
			Description: "Count the number of colors for all themes",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.SignIn,
	)
	//  me get ---------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "me",
			Method:      http.MethodGet,
			Path:        "/auth/me",
			Summary:     "Me",
			Description: "Me",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security: []map[string][]string{
				{shared.BearerAuthSecurityKey: {}},
			},
		},
		appApi.Me,
	)
	// me update -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "meUpdate",
			Method:      http.MethodPut,
			Path:        "/auth/me",
			Summary:     "Me Update",
			Description: "Me Update",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.MeUpdate,
	)
	// me delete -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "me-delete",
			Method:      http.MethodDelete,
			Path:        "/auth/me",
			Summary:     "Me delete",
			Description: "Me delete",
			Tags:        []string{"Auth", "Me"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.MeDelete,
	)
	// refresh token -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "refresh-token",
			Method:      http.MethodPost,
			Path:        "/auth/refresh-token",
			Summary:     "Refresh token",
			Description: "Count the number of colors for all themes",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.RefreshToken,
	)
	// signout -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signout",
			Method:      http.MethodPost,
			Path:        "/auth/signout",
			Summary:     "Signout",
			Description: "Signout",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.Signout,
	)
	// verify email -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "verify-get",
			Method:      http.MethodGet,
			Path:        "/auth/verify",
			Summary:     "Verify",
			Description: "Verify",
			Tags:        []string{"Auth", "Verify"},
			Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		},
		appApi.Verify,
	)

	// verify email post -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "verify-post",
			Method:      http.MethodPost,
			Path:        "/auth/verify",
			Summary:     "Verify",
			Description: "Verify",
			Tags:        []string{"Auth", "Verify"},
			Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		},
		appApi.VerifyPost,
	)
	// request verification -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "request-verification",
			Method:      http.MethodPost,
			Path:        "/auth/request-verification",
			Summary:     "Email verification request",
			Description: "Request email verification",
			Tags:        []string{"Auth", "Verify"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.RequestVerification,
	)
	// request password reset -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "request-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/request-password-reset",
			Summary:     "Request password reset",
			Description: "Request password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.RequestPasswordReset,
	)
	// confirm password reset -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "confirm-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/confirm-password-reset",
			Summary:     "Confirm password reset",
			Description: "Confirm password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.ConfirmPasswordReset,
	)
	// check password reset -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "check-password-reset",
			Method:      http.MethodGet,
			Path:        "/auth/check-password-reset",
			Summary:     "Check password reset",
			Description: "Check password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.CheckPasswordResetGet,
	)
	// password reset
	huma.Register(
		api,
		huma.Operation{
			OperationID: "reset-password",
			Method:      http.MethodPost,
			Path:        "/auth/password-reset",
			Summary:     "Reset Password",
			Description: "Reset Password",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.ResetPassword,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "oauth2-callback-get",
			Method:      http.MethodGet,
			Path:        "/auth/callback",
			Summary:     "OAuth2 Callback (GET)",
			Description: "Handle OAuth2 callback (GET)",
			Tags:        []string{"Auth", "OAuth2"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.OAuth2CallbackGet,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "oauth2-callback-post",
			Method:      http.MethodPost,
			Path:        "/auth/callback",
			Summary:     "OAuth2 Callback (POST)",
			Description: "Handle OAuth2 callback (POST)",
			Tags:        []string{"Auth", "OAuth2"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.OAuth2CallbackPost,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "oauth2-authorization-url",
			Method:      http.MethodGet,
			Path:        "/auth/authorization-url",
			Summary:     "OAuth2 Authorization URL",
			Description: "Get OAuth2 authorization URL",
			Tags:        []string{"Auth", "OAuth2"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.OAuth2AuthorizationUrl,
	)

	// authenticated routes ----------------------------------------------------------------------------------------
	// need to be authenticated to access these routes

	authenticatedGroup := huma.NewGroup(api)

	// ---- Upload File
	huma.Register(
		authenticatedGroup,
		huma.Operation{
			OperationID: "upload-media",
			Method:      http.MethodPost,
			Path:        "/media",
			Summary:     "Upload media",
			Description: "Upload a media file",
			Tags:        []string{"Media"},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Errors:      []int{http.StatusUnauthorized, http.StatusBadRequest, http.StatusInternalServerError},
		},
		appApi.UploadMedia,
	)
	// ---- Get Media
	huma.Register(
		authenticatedGroup,
		huma.Operation{
			OperationID: "get-media",
			Method:      http.MethodGet,
			Path:        "/media/{id}",
			Summary:     "Get media",
			Description: "Get a media file by ID",
			Tags:        []string{"Media"},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
		},
		appApi.GetMedia,
	)
	// ---- Get Media List
	huma.Register(
		authenticatedGroup,
		huma.Operation{
			OperationID: "list-media",
			Method:      http.MethodGet,
			Path:        "/media",
			Summary:     "List media",
			Description: "List all media files for the user",
			Tags:        []string{"Media"},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Errors:      []int{http.StatusUnauthorized, http.StatusInternalServerError},
		},
		appApi.MediaList,
	)

	// ---- notifications
	sse.Register(
		authenticatedGroup,
		huma.Operation{
			OperationID: "notifications-sse",
			Method:      http.MethodGet,
			Path:        "/notifications/sse",
			Summary:     "Notifications SSE",
			Description: "Notifications SSE",
			Tags:        []string{"Notifications"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		}, map[string]any{
			// Mapping of event type name to Go struct for that event.
			"message": models.Notification{},
		},
		appApi.NotificationsSsefunc)
	// stats routes -------------------------------------------------------------------------------------------------
	statsGroup := huma.NewGroup(api)
	huma.Register(
		statsGroup,
		huma.Operation{
			OperationID: "stats-get",
			Method:      http.MethodGet,
			Path:        "/stats",
			Summary:     "Get stats",
			Description: "Get stats",
			Tags:        []string{"Stats"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.Stats,
	)

	// ---- task routes -------------------------------------------------------------------------------------------------
	taskGroup := huma.NewGroup(api)
	taskGroup.UseMiddleware(checkTaskOwnerMiddleware)
	// task list
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-list",
			Method:      http.MethodGet,
			Path:        "/tasks",
			Summary:     "Task list",
			Description: "List of tasks",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskList,
	)
	// task create
	// huma.Register(taskGroup, appApi.TaskCreateOperation("/task"), appApi.TaskCreate)
	// task update
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-update",
			Method:      http.MethodPut,
			Path:        "/tasks/{task-id}",
			Summary:     "Task update",
			Description: "Update a task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskUpdate,
	)
	// task position
	// huma.Register(taskGroup, appApi.UpdateTaskPositionOperation("/tasks/{task-id}/position"), appApi.UpdateTaskPosition)
	// task position status
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "update-task-position-status",
			Method:      http.MethodPut,
			Path:        "/tasks/{task-id}/position-status",
			Summary:     "Update task position and status",
			Description: "Update task position and status",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.UpdateTaskPositionStatus,
	)
	// // task delete
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-delete",
			Method:      http.MethodDelete,
			Path:        "/tasks/{task-id}",
			Summary:     "Task delete",
			Description: "Delete a task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskDelete,
	)
	// // task get
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-get",
			Method:      http.MethodGet,
			Path:        "/tasks/{task-id}",
			Summary:     "Task get",
			Description: "Get a task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskGet,
	)

	// task project routes -------------------------------------------------------------------------------------------------
	taskProjectGroup := huma.NewGroup(api)
	// task project list
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-list",
			Method:      http.MethodGet,
			Path:        "/task-projects",
			Summary:     "Task project list",
			Description: "List of task projects",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskProjectList,
	)
	// task project create
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-create",
			Method:      http.MethodPost,
			Path:        "/task-projects",
			Summary:     "Task project create",
			Description: "Create a new task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskProjectCreate,
	)
	// task project create with ai
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-create-with-ai",
			Method:      http.MethodPost,
			Path:        "/task-projects/ai",
			Summary:     "Task project create with ai",
			Description: "Create a new task project with ai",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskProjectCreateWithAi,
	)
	// task project update
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-update",
			Method:      http.MethodPut,
			Path:        "/task-projects/{task-project-id}",
			Summary:     "Task project update",
			Description: "Update a task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskProjectUpdate,
	)
	// // task project delete
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-delete",
			Method:      http.MethodDelete,
			Path:        "/task-projects/{task-project-id}",
			Summary:     "Task project delete",
			Description: "Delete a task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskProjectDelete,
	)
	// // task project get
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-get",
			Method:      http.MethodGet,
			Path:        "/task-projects/{task-project-id}",
			Summary:     "Task project get",
			Description: "Get a task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskProjectGet,
	)
	// task project tasks create
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-tasks-create",
			Method:      http.MethodPost,
			Path:        "/task-projects/{task-project-id}/tasks",
			Summary:     "Task project tasks create",
			Description: "Create a new task project task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		appApi.TaskProjectTasksCreate,
	)

	// stripe routes -------------------------------------------------------------------------------------------------
	stripeGroup := huma.NewGroup(api, "/stripe")
	// stripe my subscriptions
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "stripe-my-subscriptions",
			Method:      http.MethodGet,
			Path:        "/my-subscriptions",
			Summary:     "stripe-my-subscriptions",
			Description: "stripe-my-subscriptions",
			Tags:        []string{"Payment", "Stripe", "Subscriptions"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.MyStripeSubscriptions,
	)
	// stripe webhook
	huma.Register(
		stripeGroup,
		appApi.StripeWebhookOperation("/webhook"),
		appApi.StripeWebhook,
	)
	// stripe products with prices
	huma.Register(
		stripeGroup,
		appApi.StripeProductsWithPricesOperation("/products"),
		appApi.StripeProductsWithPrices,
	)
	// stripe billing portal
	huma.Register(
		stripeGroup,
		appApi.StripeBillingPortalOperation("/billing-portal"),
		appApi.StripeBillingPortal,
	)
	//  stripe checkout session
	huma.Register(
		stripeGroup,
		appApi.StripeCheckoutSessionOperation("/checkout-session"),
		appApi.StripeCheckoutSession,
	)
	//  stripe get checkout session by checkoutSessionId
	huma.Register(
		stripeGroup,
		appApi.StripeCheckoutSessionGetOperation("/checkout-session/{checkoutSessionId}"),
		appApi.StripeCheckoutSessionGet,
	)

	//  admin routes ----------------------------------------------------------------------------
	adminGroup := huma.NewGroup(api, "/admin")
	//  admin middleware
	adminGroup.UseMiddleware(CheckPermissionsMiddleware(api, "superuser"))
	//  admin user list
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-users",
			Method:      http.MethodGet,
			Path:        "/users",
			Summary:     "Admin users",
			Description: "List of users",
			Tags:        []string{"Users", "Admin"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUsers,
	)
	// admin user get
	huma.Register(
		adminGroup,
		appApi.AdminUsersGetOperation("/users/{user-id}"),
		appApi.AdminUsersGet,
	)
	//  admin user create
	huma.Register(
		adminGroup,
		appApi.AdminUsersCreateOperation("/users"),
		appApi.AdminUsersCreate,
	)
	//  admin user delete
	huma.Register(
		adminGroup,
		appApi.AdminUsersDeleteOperation("/users/{user-id}"),
		appApi.AdminUsersDelete,
	)
	//  admin user update
	huma.Register(
		adminGroup,
		appApi.AdminUsersUpdateOperation("/users/{user-id}"),
		appApi.AdminUsersUpdate,
	)
	//  admin user update password
	huma.Register(
		adminGroup,
		appApi.AdminUsersUpdatePasswordOperation("/users/{user-id}/password"),
		appApi.AdminUsersUpdatePassword,
	)
	//  admin user update roles
	huma.Register(
		adminGroup,
		appApi.AdminUserRolesUpdateOperation("/users/{user-id}/roles"),
		appApi.AdminUserRolesUpdate,
	)
	// admin user create roles
	huma.Register(
		adminGroup,
		appApi.AdminUserRolesCreateOperation("/users/{user-id}/roles"),
		appApi.AdminUserRolesCreate,
	)
	// admin user delete roles
	huma.Register(
		adminGroup,
		appApi.AdminUserRolesDeleteOperation("/users/{user-id}/roles/{role-id}"),
		appApi.AdminUserRolesDelete,
	)
	// admin user permission source list
	huma.Register(
		adminGroup,
		appApi.AdminUserPermissionSourceListOperation("/users/{user-id}/permissions"),
		appApi.AdminUserPermissionSourceList,
	)
	// admin user permissions create
	huma.Register(
		adminGroup,
		appApi.AdminUserPermissionsCreateOperation("/users/{user-id}/permissions"),
		appApi.AdminUserPermissionsCreate,
	)
	// admin user permissions delete
	huma.Register(
		adminGroup,
		appApi.AdminUserPermissionsDeleteOperation("/users/{user-id}/permissions/{permission-id}"),
		appApi.AdminUserPermissionsDelete,
	)
	// admin user accounts list
	huma.Register(
		adminGroup,
		appApi.AdminUserAccountsOperation("/user-accounts"),
		appApi.AdminUserAccounts,
	)
	// admin roles
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-roles",
			Method:      http.MethodGet,
			Path:        "/roles",
			Summary:     "Admin roles",
			Description: "List of roles",
			Tags:        []string{"Admin", "Roles"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminRolesList,
	)
	// admin roles create
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-roles-create",
			Method:      http.MethodPost,
			Path:        "/roles",
			Summary:     "Create role",
			Description: "Create role",
			Tags:        []string{"Admin", "Roles"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminRolesCreate,
	)
	// admin roles update
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-roles-update",
			Method:      http.MethodPut,
			Path:        "/roles/{id}",
			Summary:     "Update role",
			Description: "Update role",
			Tags:        []string{"Admin", "Roles"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminRolesUpdate,
	)
	// admin role get
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-roles-get",
			Method:      http.MethodGet,
			Path:        "/roles/{id}",
			Summary:     "Get role",
			Description: "Get role",
			Tags:        []string{"Admin", "Roles"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminRolesGet,
	)
	// admin roles update permissions
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-roles-update-permissions",
			Method:      http.MethodPut,
			Path:        "/roles/{id}/permissions",
			Summary:     "Update role permissions",
			Description: "Update role permissions",
			Tags:        []string{"Admin", "Roles"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminRolesUpdatePermissions,
	)
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
	// admin stripe subscriptions get
	huma.Register(adminGroup, appApi.AdminStripeSubscriptionsGetOperation("/subscriptions/{subscription-id}"), appApi.AdminStripeSubscriptionsGet)
	// admin stripe products
	huma.Register(adminGroup, appApi.AdminStripeProductsOperation("/products"), appApi.AdminStripeProducts)
	//  admin stripe products get
	huma.Register(adminGroup, appApi.AdminStripeProductsGetOperation("/products/{product-id}"), appApi.AdminStripeProductsGet)
	// admin stripe products roles create
	huma.Register(adminGroup, appApi.AdminStripeProductsRolesCreateOperation("/products/{product-id}/roles"), appApi.AdminStripeProductsRolesCreate)
	// admin stripe products roles delete
	huma.Register(adminGroup, appApi.AdminStripeProductsRolesDeleteOperation("/products/{product-id}/roles/{role-id}"), appApi.AdminStripeProductsRolesDelete)
	// admin stripe products with prices

}

func AddRoutes(api huma.API, app core.App) {
	BindMiddlewares(api, app)
	BindApis(api, app)
}
