package shared

import (
	"github.com/aarondl/opt/null"
)

// ProvidersGoogle      Providers = "google"
// ProvidersApple       Providers = "apple"
// ProvidersFacebook    Providers = "facebook"
// ProvidersGithub      Providers = "github"
// ProvidersCredentials Providers = "credentials"

type SortParams struct {
	SortBy    string `query:"sort_by,omitempty" required:"false"`
	SortOrder string `query:"sort_order,omitempty" required:"false" enum:"asc,desc"`
}

type PaginatedInput struct {
	Page    int `query:"page,omitempty" default:"1" minimum:"1" required:"false"`
	PerPage int `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100" required:"false"`
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
