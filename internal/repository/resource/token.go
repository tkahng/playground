package resource

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/types"
)

type TokenFilter struct {
	PaginatedInput
	SortParams
	Q             string                         `query:"q,omitempty" required:"false"`
	UserIds       []uuid.UUID                    `query:"user_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Ids           []uuid.UUID                    `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Types         []models.TokenTypes            `query:"types,omitempty" required:"false" doc:"Filter by token type, e.g., 'access', 'refresh'"`
	Identifiers   []string                       `query:"identifiers,omitempty" required:"false" minimum:"1" maximum:"100"`
	Tokens        []string                       `query:"tokens,omitempty" required:"false" minimum:"1" maximum:"100"`
	ExpiresAfter  types.OptionalParam[time.Time] `query:"expires_after,omitempty" required:"false" minimum:"1" maximum:"100"`
	ExpiresBefore types.OptionalParam[time.Time] `query:"expires_before,omitempty" required:"false" minimum:"1" maximum:"100"`
}

func NewTokenRepositoryResource(db database.Dbx) *RepositoryResource[models.Token, uuid.UUID, TokenFilter] {
	return NewRepositoryResource[models.Token, uuid.UUID](
		db,
		repository.Token,
		func(filter *TokenFilter) *map[string]any {
			if filter == nil {
				return nil
			}
			where := make(map[string]any)
			if len(filter.UserIds) > 0 {
				where["user_id"] = map[string]any{"_in": filter.UserIds}
			}
			if len(filter.Ids) > 0 {
				where["id"] = map[string]any{"_in": filter.Ids}
			}
			if len(filter.Types) > 0 {
				where["type"] = map[string]any{"_in": filter.Types}
			}
			if len(filter.Identifiers) > 0 {
				where["identifier"] = map[string]any{"_in": filter.Identifiers}
			}
			if len(filter.Tokens) > 0 {
				where["token"] = map[string]any{"_in": filter.Tokens}
			}
			if filter.ExpiresAfter.IsSet {
				where["expires"] = map[string]any{"_gte": filter.ExpiresAfter.Value}
			}
			if filter.ExpiresBefore.IsSet {
				where["expires"] = map[string]any{"_lte": filter.ExpiresBefore.Value}
			}
			return &where
		},
		nil,
		nil,
	)
}
