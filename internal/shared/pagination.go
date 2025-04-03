package shared

import (
	"github.com/aarondl/opt/null"
	"github.com/tkahng/authgo/internal/db/models"
)

// ProvidersGoogle      Providers = "google"
// ProvidersApple       Providers = "apple"
// ProvidersFacebook    Providers = "facebook"
// ProvidersGithub      Providers = "github"
// ProvidersCredentials Providers = "credentials"

type UserListFilter struct {
	Providers []models.Providers `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	Q         string             `query:"q,omitempty" required:"false"`
	Ids       []string           `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Emails    []string           `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	// Roles         []string           `query:"roles,omitempty" required:"false" minimum:"1" maximum:"100"`
	// Permissions   []string           `query:"permissions,omitempty" required:"false" minimum:"1" maximum:"100"`
	RoleIds       []string `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	PermissionIds []string `query:"permission_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	// 	Provider  OmitNull[models.Providers]   `query:"provider,omitempty" required:"false" enum:"google,apple,facebook,github,credentials"`
	// 	Providers OmitNull[[]models.Providers] `query:"providers,omitempty" required:"false" uniqueItems:"true" minLength:"1" maxLength:"80" enum:"google,apple,facebook,github,credentials"`
	// 	Q         OmitNull[string]             `query:"q,omitempty" required:"false" default:""`
}
type UserListParams struct {
	PaginatedInput
	UserListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"roles,permissions,accounts,subscriptions"`
}
type RoleListFilter struct {
	Q      string   `query:"q,omitempty" required:"false"`
	Ids    []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	NotIds []string `query:"not_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names  []string `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	// Roles         []string           `query:"roles,omitempty" required:"false" minimum:"1" maximum:"100"`
	// Permissions   []string           `query:"permissions,omitempty" required:"false" minimum:"1" maximum:"100"`
	PermissionIds []string `query:"permission_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	// 	Provider  OmitNull[models.Providers]   `query:"provider,omitempty" required:"false" enum:"google,apple,facebook,github,credentials"`
	// 	Providers OmitNull[[]models.Providers] `query:"providers,omitempty" required:"false" uniqueItems:"true" minLength:"1" maxLength:"80" enum:"google,apple,facebook,github,credentials"`
	// 	Q         OmitNull[string]             `query:"q,omitempty" required:"false" default:""`
}
type RolesListParams struct {
	PaginatedInput
	RoleListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100"`
}
type PermissionsListFilter struct {
	Q     string   `query:"q,omitempty" required:"false"`
	Ids   []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names []string `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	// Roles         []string           `query:"roles,omitempty" required:"false" minimum:"1" maximum:"100"`
	// Permissions   []string           `query:"permissions,omitempty" required:"false" minimum:"1" maximum:"100"`
	RoleIds []string `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	// 	Provider  OmitNull[models.Providers]   `query:"provider,omitempty" required:"false" enum:"google,apple,facebook,github,credentials"`
	// 	Providers OmitNull[[]models.Providers] `query:"providers,omitempty" required:"false" uniqueItems:"true" minLength:"1" maxLength:"80" enum:"google,apple,facebook,github,credentials"`
	// 	Q         OmitNull[string]             `query:"q,omitempty" required:"false" default:""`
}
type PermissionsListParams struct {
	PaginatedInput
	PermissionsListFilter
	// 	SortBy    string `query:"sort_by,omitempty" required:"false" default:"created_at"`
	// 	SortOrder string `query:"sort_order,omitempty" required:"false" default:"desc"`
	SortBy    string `query:"sort_by,omitempty" required:"false" default:"created_at"`
	SortOrder string `query:"sort_order,omitempty" required:"false" default:"desc"`
}
type StripeProductListFilter struct {
	Q      string       `query:"q,omitempty" required:"false"`
	Ids    []string     `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Active ActiveStatus `query:"active,omitempty" required:"false" enum:"active,inactive"`
}

type SortParams struct {
	SortBy    string `query:"sort_by,omitempty" required:"false" default:"created_at"`
	SortOrder string `query:"sort_order,omitempty" required:"false" default:"desc"`
}
type StripeProductListParams struct {
	PaginatedInput
	StripeProductListFilter
	SortParams
	Expand      []string     `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100"`
	PriceActive ActiveStatus `query:"price_active,omitempty" required:"false" enum:"active,inactive"`
}
type BooleanType string

const (
	True  BooleanType = "true"
	False BooleanType = "false"
)

type ActiveStatus string

const (
	Active   ActiveStatus = "active"
	Inactive ActiveStatus = "inactive"
)

type StripePriceListFilter struct {
	Q      string       `query:"q,omitempty" required:"false"`
	Ids    []string     `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Active ActiveStatus `query:"active,omitempty" required:"false" enum:"active,inactive"`
}
type StripePriceListParams struct {
	PaginatedInput
	StripePriceListFilter
	SortParams
}

type StripeCustomerListFilter struct {
	Q   string   `query:"q,omitempty" required:"false"`
	Ids []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}
type StripeCustomerListParams struct {
	PaginatedInput
	StripeCustomerListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"users"`
}

type PaginatedInput struct {
	Page    int `query:"page,omitempty" default:"1" minimum:"1"`
	PerPage int `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100"`
}

type PaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type Link struct {
	URL    *string `json:"url"`
	Label  string  `json:"label"`
	Active bool    `json:"active"`
}
type MetaLink struct {
	First null.Val[string] `json:"first"`
	Last  null.Val[string] `json:"last"`
	Next  null.Val[string] `json:"next"`
	Prev  null.Val[string] `json:"prev"`
}

// meta: {
// 	current_page: number;
// 	from: number | null;
// 	last_page: number;
// 	/** @description Generated paginator links. */
// 	links: {
// 		url: string | null;
// 		label: string;
// 		active: boolean;
// 	}[];
// 	/** @description Base path for paginator generated URLs. */
// 	path: string | null;
// 	/** @description Number of items shown per page. */
// 	per_page: number;
// 	/** @description Number of the last item in the slice. */
// 	to: number | null;
// 	/** @description Total number of items being paginated. */
// 	total: number;
// };
