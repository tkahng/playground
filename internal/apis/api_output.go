package apis

import (
	"math"
)

type ApiPaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
type ApiPaginatedOutput[T any] struct {
	Body ApiPaginatedResponse[T] `json:"body"`
}
type ApiSingleOutput[T any] struct {
	Body ApiSingleResponse[T] `json:"body"`
}
type ApiSingleResponse[T any] struct {
	Data T `json:"data"`
}
type ApiOutput[T any] struct {
	Body T `json:"body"`
}
type Meta struct {
	Page     int64  `json:"page"`
	PerPage  int64  `json:"per_page"`
	Total    int64  `json:"total"`
	NextPage *int64 `json:"next_page"`
	PrevPage *int64 `json:"prev_page"`
	HasMore  bool   `json:"has_more"`
}

func ApiGenerateMeta(input *PaginatedInput, total int64) Meta {
	meta := Meta{
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
		meta.HasMore = true
		meta.NextPage = &nextPage
	} else {
		meta.NextPage = nil
	}
	return meta
}
