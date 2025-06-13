package shared

type SortParams struct {
	SortBy    string `query:"sort_by,omitempty" required:"false"`
	SortOrder string `query:"sort_order,omitempty" required:"false" enum:"asc,desc"`
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

func (p *PaginatedInput) LimitOffset() (limit, offset int) {
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

func (p *PaginatedInput) Pagination() (page, perPage int) {
	return int(p.Page), int(p.PerPage)
}

type Meta struct {
	Page     int64  `json:"page"`
	PerPage  int64  `json:"per_page"`
	Total    int64  `json:"total"`
	NextPage *int64 `json:"next_page"`
	PrevPage *int64 `json:"prev_page"`
	HasMore  bool   `json:"has_more"`
}

type JoinedResult[T any, K any] struct {
	Key  K   `db:"key"`
	Data []T `db:"data"`
}
