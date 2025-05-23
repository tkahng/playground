package apis

import (
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

func TestBindTeamsApi(t *testing.T) {
	type args struct {
		api    huma.API
		appApi *Api
	}
	tests := []struct {
		name      string
		args      args
		setupMock func()
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}
