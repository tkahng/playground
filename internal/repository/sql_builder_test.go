package repository

import (
	"testing"

	"github.com/tkahng/authgo/internal/models"
)

type buidlerWants struct {
	skipIdGeneration bool
	generator        bool
}

func TestNewSQLBuilder(t *testing.T) {
	type args struct {
		opts []SQLBuilderOptions[models.User]
	}
	tests := []struct {
		name string
		args args
		want buidlerWants
	}{
		{
			name: "Test case 1",
			args: args{opts: []SQLBuilderOptions[models.User]{
				UuidV7Generator[models.User],
			}},
			want: buidlerWants{skipIdGeneration: false, generator: true}, // expected result here
		},
		{
			name: "Test case 2",
			args: args{opts: []SQLBuilderOptions[models.User]{}},
			want: buidlerWants{skipIdGeneration: false, generator: false}, // expected result here
		},
		// TODO: Add more test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSQLBuilder(tt.args.opts...)
			if got.insertID != tt.want.skipIdGeneration {
				t.Errorf("NewSQLBuilder() skipIdGeneration = %v, want %v", got.insertID, tt.want.skipIdGeneration)
			}
			if tt.want.generator {
				if got.generator == nil {
					t.Errorf("NewSQLBuilder() idGenerator = nil, want non-nil")
				}
			} else {
				if got.generator != nil {
					t.Errorf("NewSQLBuilder() idGenerator non-nil, want nil")
				}
			}
		})
	}
}
