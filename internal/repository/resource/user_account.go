package resource

import (
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
)

type UserAccountFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Providers     []models.Providers     `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	ProviderTypes []models.ProviderTypes `query:"provider_types,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"oauth,credentials"`
	Q             string                 `query:"q,omitempty" required:"false"`
	Ids           []uuid.UUID            `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserIds       []uuid.UUID            `query:"user_ids,omitempty" minimum:"1" maximum:"100" required:"false" format:"uuid"`
}

func NewUserAccountRepositoryResource(
	db database.Dbx,
) *RepositoryResource[models.UserAccount, uuid.UUID, UserAccountFilter] {
	resource := NewRepositoryResource[models.UserAccount, uuid.UUID](
		db,
		repository.UserAccount,
		func(filter *UserAccountFilter) *map[string]any {
			where := make(map[string]any)
			if filter == nil {
				return &where
			}
			if len(filter.Providers) > 0 {
				where[models.UserAccountTable.Provider] = map[string]any{
					"_in": filter.Providers,
				}
			}
			if len(filter.ProviderTypes) > 0 {
				where[models.UserAccountTable.Type] = map[string]any{
					"_in": filter.ProviderTypes,
				}
			}
			if len(filter.Ids) > 0 {
				where[models.UserAccountTable.ID] = map[string]any{
					"_in": filter.Ids,
				}
			}
			if len(filter.UserIds) > 0 {
				where[models.UserAccountTable.UserID] = map[string]any{
					"_in": filter.UserIds,
				}
			}
			return &where
		},
		nil,
		nil,
	)
	return resource
}
