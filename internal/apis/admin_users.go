package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) AdminUsersOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin users",
		Description: "List of users",
		Tags:        []string{"Auth", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}
func (api *Api) AdminUsers(ctx context.Context, input *struct {
	shared.UserListParams
}) (*RequestPasswordResetOutput, error) {
	// data, err := repository.ListUsers(ctx, input)
	utils.PrettyPrintJSON(input)
	return nil, nil
}
