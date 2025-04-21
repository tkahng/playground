package shared

import (
	"math"

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
	Page    int64 `query:"page,omitempty" minimum:"0" required:"false"`
	PerPage int64 `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100" required:"false"`
}

type PaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
type Meta struct {
	Page     int64  `json:"page"`
	PerPage  int64  `json:"per_page"`
	Total    int64  `json:"total"`
	NextPage *int64 `json:"next_page"`
	PrevPage *int64 `json:"prev_page"`
	HasMore  bool   `json:"has_more"`
}

func GenerateMeta(input PaginatedInput, total int64) Meta {
	var meta Meta = Meta{
		Page:    input.Page,
		PerPage: input.PerPage,
		Total:   total,
	}
	nextPage, prevPage := input.Page+1, input.Page-1

	perPage := input.PerPage
	if perPage == 0 {
		perPage = 10
	}
	pageCount := int64(math.Ceil(float64(total) / float64(perPage)))

	if prevPage >= 0 {
		meta.PrevPage = &prevPage
	} else {
		meta.PrevPage = nil
	}
	if nextPage < pageCount-1 {
		meta.NextPage = &nextPage
	} else {
		meta.NextPage = nil
	}
	return meta
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
