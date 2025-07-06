package repository

type Sortable interface {
	Sort() (sortBy, sortOrder string)
}
type DefaultFilter interface {
	Sortable
	Paginable
}
type SortParams struct {
	SortBy    string `query:"sort_by,omitempty" required:"false"`
	SortOrder string `query:"sort_order,omitempty" required:"false" enum:"asc,desc"`
}

func (s *SortParams) Sort() (sortBy, sortOrder string) {
	if s == nil {
		return "", "" // default values
	}
	// if s.SortBy == "" {
	// 	s.SortBy = "created_at" // default sort by
	// }
	// if s.SortOrder == "" {
	// 	s.SortOrder = "desc" // default sort order
	// }
	return s.SortBy, s.SortOrder
}

type PaginatedInput struct {
	Page    int64 `query:"page,omitempty" minimum:"0" required:"false"`
	PerPage int64 `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100" required:"false"`
}

type Paginable interface {
	Pagination() (limit, offset int)
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
