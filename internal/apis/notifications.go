package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2/sse"
)

func (api *Api) NotificationsSseEvents() map[string]any {
	// sse.
	return map[string]any{
		"notifications": "notifications",
	}
}

func (api *Api) NotificationsSsefunc(ctx context.Context, input *struct{},
	send sse.Sender,
) {

}

// func (api *Api) NotificationsListOperation(path string) huma.Operation {
// 	return huma.Operation{
// 		OperationID: "notifications-list",
// 		Method:      http.MethodGet,
// 		Path:        path,
// 		Summary:     "List notifications",
// 		Description: "List notifications",
// 		Tags:        []string{"Notifications"},
// 		Errors:      []int{http.StatusNotFound},
// 		Security: []map[string][]string{
// 			{shared.BearerAuthSecurityKey: {}},
// 		},
// 	}
// }

// func (api *Api) NotificationsList(ctx context.Context, input *struct {
// 	shared.NotificationsListParams
// }) (*shared.PaginatedOutput[*models.Notification], error) {
// 	db := api.app.Db()
// 	notifications, err := repository.ListNotifications(ctx, db, &input.NotificationsListParams)
// 	if err != nil {
// 		return nil, err
// 	}
// 	count, err := repository.CountNotifications(ctx, db, &input.NotificationsListFilter)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &shared.PaginatedOutput[*models.Notification]{
// 		Body: shared.PaginatedResponse[*models.Notification]{

// 			Data: notifications,
// 			Meta: shared.Meta{
// 				Page:    input.PaginatedInput.Page,
// 				PerPage: input.PaginatedInput.PerPage,
// 				Total:   count,
// 			},
// 		},
// 	}, nil

// }
