package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/test"
)

func TestPoller_Run(t *testing.T) {
	_, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		type fields struct {
			Store      JobStore
			Dispatcher Dispatcher
			opts       pollerOpts
		}
		type args struct {
			ctx context.Context
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			// TODO: Add test cases.
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p := NewPoller(tt.fields.Store, tt.fields.Dispatcher)

				if err := p.Run(tt.args.ctx); (err != nil) != tt.wantErr {
					t.Errorf("Poller.Run() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}
