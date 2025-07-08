package apis

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/tkahng/authgo/internal/core"
)

func TestApi_GetStripeSubscriptions(t *testing.T) {
	type fields struct {
		app core.App
	}
	type args struct {
		ctx   context.Context
		input *struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *struct{ Body *SubscriptionWithPrice }
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appApi := &Api{
				a: tt.fields.app,
			}
			_, api := humatest.New(t)

			AddRoutes(api, appApi)
			resp := api.GetCtx(tt.args.ctx, "/api/subscriptions/active", tt.args.input)
			if resp.Code != 200 {
				t.Errorf("Api.GetStripeSubscriptions() = %v, want %v", resp.Code, 200)
			}
			if resp.Body == nil {
				t.Errorf("Api.GetStripeSubscriptions() = %v, want %v", resp.Body, "not nil")
			}
		})
	}
}
