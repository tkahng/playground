package apis

import (
	"math"

	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

type SortParams struct {
	SortBy    string `query:"sort_by,omitempty" required:"false"`
	SortOrder string `query:"sort_order,omitempty" required:"false" enum:"asc,desc"`
}

func (s *SortParams) ToRepoSort() *repository.SortParams {
	if s == nil {
		return nil // default values
	}
	return &repository.SortParams{
		SortBy:    s.SortBy,
		SortOrder: s.SortOrder,
	}
}

func (s *SortParams) Sort() (sortBy, sortOrder string) {
	if s == nil {
		return "", "" // default values
	}
	return s.SortBy, s.SortOrder
}

type PaginatedInput struct {
	Page    int64 `query:"page,omitempty" minimum:"0" required:"false"`
	PerPage int64 `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100" required:"false"`
}

func (p *PaginatedInput) Pagination() (limit, offset int) {
	if p == nil {
		return 10, 0 // default values
	}
	if p.PerPage <= 0 {
		p.PerPage = 10 // default value
	}
	if p.Page < 0 {
		p.Page = 0 // default value
	}
	return int(p.PerPage), int(p.Page) * int(p.PerPage)
}

type PaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
type ApiPaginatedOutput[T any] struct {
	Body PaginatedResponse[T] `json:"body"`
}
type Meta struct {
	Page     int64  `json:"page"`
	PerPage  int64  `json:"per_page"`
	Total    int64  `json:"total"`
	NextPage *int64 `json:"next_page"`
	PrevPage *int64 `json:"prev_page"`
	HasMore  bool   `json:"has_more"`
}

func ApiGenerateMeta(input *PaginatedInput, total int64) shared.Meta {
	var meta shared.Meta = shared.Meta{
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
