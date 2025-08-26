package test_test

import (
	"testing"

	"github.com/tkahng/playground/internal/test"
)

func TestGetLinkParam(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		html      string
		paramName string
		want      string
		wantErr   bool
	}{
		{
			name:      "get link param",
			paramName: "token",
			want:      "some random token with spaces&symbols!",
			html:      `<a href="https://example.com/callback?token=some+random+token+with+spaces%26symbols%21">link</a>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, gotErr := test.GetLinkParam(tt.html, tt.paramName)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetLinkParam() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetLinkParam() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GetLinkParam() = %v, want %v", got, tt.want)
			}
		})
	}
}
