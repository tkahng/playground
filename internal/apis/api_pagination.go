package apis

import (
	"github.com/tkahng/playground/internal/database/repository"
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

const maxPerPage = 100

func (p *PaginatedInput) LimitOffset() (limit, offset int) {
	if p == nil {
		return 10, 0 // default values
	}
	if p.PerPage <= 0 {
		p.PerPage = 10
	}
	if p.PerPage > maxPerPage {
		p.PerPage = maxPerPage
	}
	if p.Page < 0 {
		p.Page = 0
	}
	return int(p.PerPage), int(p.Page) * int(p.PerPage)
}
