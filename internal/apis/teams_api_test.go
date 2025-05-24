package apis_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/apis"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestCreateTeam(t *testing.T) {
	test.DbSetup()

	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)
		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		tests := []struct {
			name             string
			ctxUserInfo      *shared.UserInfo
			createTeamErr    error
			createTeamResult *shared.TeamInfo
			expectedErr      error
			expectedOutput   *apis.TeamOutput
		}{
			{
				name:        "unauthorized error",
				ctxUserInfo: nil,
				expectedErr: huma.Error401Unauthorized("unauthorized"),
			},
			{
				name:          "error propagation",
				ctxUserInfo:   &shared.UserInfo{User: shared.User{ID: uuid.New()}},
				createTeamErr: errors.New("test error"),
				expectedErr:   errors.New("test error"),
			},
			{
				name:             "team not found error",
				ctxUserInfo:      &shared.UserInfo{User: shared.User{ID: uuid.New()}},
				createTeamResult: nil,
				expectedErr:      huma.Error500InternalServerError("team not found"),
			},
			{
				name:             "successful team creation",
				ctxUserInfo:      &shared.UserInfo{User: shared.User{ID: uuid.New()}},
				createTeamResult: &shared.TeamInfo{Team: models.Team{}},
				expectedOutput:   &apis.TeamOutput{Body: shared.FromTeamModel(&models.Team{})},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				ctx := contextstore.SetContextUserInfo(context.Background(), tt.ctxUserInfo)
				resp := api.PostCtx(ctx,
					"/api/teams",
					&struct {
						Body apis.CreateTeamInput `json:"body" required:"true"`
					}{
						Body: apis.CreateTeamInput{Name: tt.name, Slug: "test-slug"},
					})
				if tt.expectedErr != nil {
					assert.GreaterOrEqual(t, resp.Code, 400)
				}
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err, "Error reading response body")
				fmt.Println(string(body))
				var expectedResponse apis.TeamOutput // Or your specific JSON struct
				err = json.Unmarshal(body, &expectedResponse)
				if tt.expectedErr != nil {
					if resp.Code >= 400 {
						return
					}

				}
				assert.Equal(t, tt.expectedOutput.Body.Name, expectedResponse.Body.Name)
			})
		}

	})
}
