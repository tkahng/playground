package database

import (
	"github.com/Masterminds/squirrel"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

func Paginate(q squirrel.SelectBuilder, input *shared.PaginatedInput) squirrel.SelectBuilder {
	if input == nil {
		input = &shared.PaginatedInput{
			PerPage: 10,
			Page:    0,
		}
	}
	if input.PerPage == 0 {
		input.PerPage = 10
	}
	return q.Limit(uint64(input.PerPage)).Offset(uint64(input.Page * input.PerPage))
}

func PaginateRepo(input *shared.PaginatedInput) (*int, *int) {
	if input == nil {
		input = &shared.PaginatedInput{
			PerPage: 10,
			Page:    0,
		}
	}
	if input.PerPage == 0 {
		input.PerPage = 10
	}
	return types.Pointer(int(input.PerPage)), types.Pointer(int(input.Page * input.PerPage))
}
