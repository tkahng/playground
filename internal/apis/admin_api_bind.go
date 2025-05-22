package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

func BindAdminApi(api huma.API, appApi *Api) {
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
		huma.Operation{
			OperationID: "admin-user-get",
			Method:      http.MethodGet,
			Path:        "/users/{user-id}",
			Summary:     "Get user",
			Description: "Get user",
			Tags:        []string{"Admin", "Users"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{
				{shared.BearerAuthSecurityKey: {}},
			},
		},
		appApi.AdminUsersGet,
	)
	//  admin user create
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-users-create",
			Method:      http.MethodPost,
			Path:        "/users",
			Summary:     "Create user",
			Description: "Create user",
			Tags:        []string{"Users", "Admin"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUsersCreate,
	)
	//  admin user delete
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-users-delete",
			Method:      http.MethodDelete,
			Path:        "/users/{user-id}",
			Summary:     "Delete user",
			Description: "Delete user",
			Tags:        []string{"Admin", "Users"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUsersDelete,
	)
	//  admin user update
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-users-update",
			Method:      http.MethodPut,
			Path:        "/users/{user-id}",
			Summary:     "Update user",
			Description: "Update user",
			Tags:        []string{"Admin", "Users"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUsersUpdate,
	)
	//  admin user update password
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-users-update-password",
			Method:      http.MethodPut,
			Path:        "/users/{user-id}/password",
			Summary:     "Update user password",
			Description: "Update user password",
			Tags:        []string{"Admin", "Users"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUsersUpdatePassword,
	)
	//  admin user update roles
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-update-user-roles",
			Method:      http.MethodPut,
			Path:        "/users/{user-id}/roles",
			Summary:     "Update user roles",
			Description: "Update user roles",
			Tags:        []string{"Admin", "Roles"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUserRolesUpdate,
	)
	// admin user create roles
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-create-user-roles",
			Method:      http.MethodPost,
			Path:        "/users/{user-id}/roles",
			Summary:     "Create user roles",
			Description: "Create user roles",
			Tags:        []string{"Admin", "Roles", "User"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUserRolesCreate,
	)
	// admin user delete roles
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-user-roles-delete",
			Method:      http.MethodDelete,
			Path:        "/users/{user-id}/roles/{role-id}",
			Summary:     "Delete user roles",
			Description: "Delete user roles",
			Tags:        []string{"Admin", "Roles", "User"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUserRolesDelete,
	)
	// admin user permission source list
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-user-permission-sources",
			Method:      http.MethodGet,
			Path:        "/users/{user-id}/permissions",
			Summary:     "Admin user permission sources",
			Description: "List of permission sources",
			Tags:        []string{"Admin", "Permissions", "User"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUserPermissionSourceList,
	)
	// admin user permissions create
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-user-permissions-create",
			Method:      http.MethodPost,
			Path:        "/users/{user-id}/permissions",
			Summary:     "Create user permission",
			Description: "Create user permission",
			Tags:        []string{"Admin", "Permissions", "User"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUserPermissionsCreate,
	)
	// admin user permissions delete
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-user-permissions-delete",
			Method:      http.MethodDelete,
			Path:        "/users/{user-id}/permissions/{permission-id}",
			Summary:     "Delete user permission",
			Description: "Delete user permission",
			Tags:        []string{"Admin", "Permissions", "User"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminUserPermissionsDelete,
	)
	// admin user accounts list
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-user-accounts",
			Method:      http.MethodGet,
			Path:        "/user-accounts",
			Summary:     "Admin user accounts",
			Description: "List of user accounts",
			Tags:        []string{"User Accounts", "Admin"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
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
	huma.Register(
		adminGroup,
		huma.Operation{
			OperationID: "admin-roles-create-permissions",
			Method:      http.MethodPost,
			Path:        "/roles/{id}/permissions",
			Summary:     "Create role permissions",
			Description: "Create role permissions",
			Tags:        []string{"Admin", "Roles", "Permissions"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.AdminRolesCreatePermissions,
	)
	// admin roles delete permissions
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-roles-delete-permissions", Method: http.MethodDelete, Path: "/roles/{roleId}/permissions/{permissionId}", Summary: "Delete role permissions", Description: "Delete role permissions", Tags: []string{"Admin", "Roles", "Permissions"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminRolesDeletePermissions)
	// admin roles delete
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-roles-delete", Method: http.MethodDelete, Path: "/roles/{id}", Summary: "Delete role", Description: "Delete role", Tags: []string{"Admin", "Roles"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminRolesDelete)
	// admin permissions list
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-permissions", Method: http.MethodGet, Path: "/permissions", Summary: "Admin permissions", Description: "List of permissions", Tags: []string{"Admin", "Permissions"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminPermissionsList)
	// admin permissions create
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-permissions-create", Method: http.MethodPost, Path: "/permissions", Summary: "Create permission", Description: "Create permission", Tags: []string{"Admin", "Permissions"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminPermissionsCreate)
	// admin permissions get
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-permissions-get", Method: http.MethodGet, Path: "/permissions/{id}", Summary: "Get permission", Description: "Get permission", Tags: []string{"Admin", "Permissions"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminPermissionsGet)
	// admin permissions update
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-permissions-update", Method: http.MethodPut, Path: "/permissions/{id}", Summary: "Update permission", Description: "Update permission", Tags: []string{"Admin", "Permissions"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminPermissionsUpdate)
	// admin permissions delete
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-permissions-delete", Method: http.MethodDelete, Path: "/permissions/{id}", Summary: "Delete permission", Description: "Delete permission", Tags: []string{"Admin", "Permissions"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminPermissionsDelete)

	// admin stripe subscriptions
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-stripe-subscriptions", Method: http.MethodGet, Path: "/subscriptions", Summary: "Admin stripe subscriptions", Description: "List of stripe subscriptions", Tags: []string{"Admin", "Subscription", "Stripe"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminStripeSubscriptions)
	// admin stripe subscriptions get
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-stripe-subscription-get", Method: http.MethodGet, Path: "/subscriptions/{subscription-id}", Summary: "Admin stripe subscription get", Description: "Get a stripe subscription by ID", Tags: []string{"Admin", "Subscription", "Stripe"}, Errors: []int{http.StatusNotFound, http.StatusBadRequest}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminStripeSubscriptionsGet)
	// admin stripe products
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-stripe-products", Method: http.MethodGet, Path: "/products", Summary: "Admin stripe products", Description: "List of stripe products", Tags: []string{"Admin", "Product", "Stripe"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminStripeProducts)
	//  admin stripe products get
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-stripe-product-get", Method: http.MethodGet, Path: "/products/{product-id}", Summary: "Admin stripe product get", Description: "Get a stripe product by ID", Tags: []string{"Admin", "Product", "Stripe"}, Errors: []int{http.StatusNotFound, http.StatusBadRequest}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminStripeProductsGet)
	// admin stripe products roles create
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-create-product-roles", Method: http.MethodPost, Path: "/products/{product-id}/roles", Summary: "Create product roles", Description: "Create product roles", Tags: []string{"Admin", "Roles", "Product", "Stripe"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminStripeProductsRolesCreate)
	// admin stripe products roles delete
	huma.Register(adminGroup, huma.Operation{OperationID: "admin-delete-product-roles", Method: http.MethodDelete, Path: "/products/{product-id}/roles/{role-id}", Summary: "Delete product roles", Description: "Delete product roles", Tags: []string{"Admin", "Roles", "Product", "Stripe"}, Errors: []int{http.StatusNotFound}, Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}}}, appApi.AdminStripeProductsRolesDelete)
}
