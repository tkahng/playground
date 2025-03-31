package shared

import (
	"github.com/aarondl/opt/null"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
)

// ProvidersGoogle      Providers = "google"
// ProvidersApple       Providers = "apple"
// ProvidersFacebook    Providers = "facebook"
// ProvidersGithub      Providers = "github"
// ProvidersCredentials Providers = "credentials"

type UserListFilter struct {
	Provider  models.Providers   `query:"provider,omitempty" required:"false" enum:"google,apple,facebook,github,credentials"`
	Providers []models.Providers `query:"providers,omitempty" required:"false" uniqueItems:"true" minLength:"1" maxLength:"80" enum:"google,apple,facebook,github,credentials"`
	Q         string             `query:"q,omitempty" required:"false" default:""`
	Ids       []uuid.UUID        `query:"ids,omitempty" required:"false"`
	Emails    []string           `query:"emails,omitempty" required:"false"`
	// 	Provider  OmitNull[models.Providers]   `query:"provider,omitempty" required:"false" enum:"google,apple,facebook,github,credentials"`
	// 	Providers OmitNull[[]models.Providers] `query:"providers,omitempty" required:"false" uniqueItems:"true" minLength:"1" maxLength:"80" enum:"google,apple,facebook,github,credentials"`
	// 	Q         OmitNull[string]             `query:"q,omitempty" required:"false" default:""`
}
type UserListParams struct {
	PaginatedInput
	UserListFilter
	// 	SortBy    string `query:"sort_by,omitempty" required:"false" default:"created_at"`
	// 	SortOrder string `query:"sort_order,omitempty" required:"false" default:"desc"`
	SortBy    string `query:"sort_by,omitempty" required:"false" default:"created_at"`
	SortOrder string `query:"sort_order,omitempty" required:"false" default:"desc"`
}

type PaginatedInput struct {
	Page    int `query:"page,omitempty" default:"1" minimum:"1"`
	PerPage int `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100"`
	// 	Page    OmitNull[int] `query:"page,omitempty" default:"1" minimum:"1"`
	// 	PerPage OmitNull[int] `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100"`
}

type PaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
type Meta struct {
	Page int `json:"page"`
	// From        null.Val[string]    `json:"from"`
	// LastPage    int     `json:"last_page"`
	// Links   []Link  `json:"links"`
	// Path    *string `json:"path"`
	PerPage int `json:"per_page"`
	// To          *int    `json:"to"`
	Total int `json:"total"`
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
