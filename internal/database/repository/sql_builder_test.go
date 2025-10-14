package repository_test

import (
	"context"
	"fmt"
	"testing"
)

func TestSQLBuilder_WhereError(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		where   *map[string]any
		args    *[]any
		want    string
		wantErr bool
	}{
		{
			name: "where id _eq hello",
			where: &map[string]any{
				"id": map[string]any{
					"_eq": "hello",
				},
			},
			args:    &[]any{},
			wantErr: false,
			want:    "SELECT a.id,a.age FROM a WHERE a.id = $1",
		},
		{
			name: "where id _eq hello and age _eq 10",
			where: &map[string]any{
				"id": map[string]any{
					"_eq": "hello",
				},
				"age": map[string]any{
					"_eq": 10,
				},
			},
			args:    &[]any{},
			wantErr: false,
			want:    "SELECT a.id,a.age FROM a WHERE a.id = $1 AND a.age = $2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := fmt.Sprintf("SELECT %s FROM %s", ABuilder.QualifiedColumnNamesJoined(), ABuilder.TableName())
			got, gotErr := ABuilder.WhereError(context.Background(), tt.where, tt.args, nil)
			query += fmt.Sprintf(" WHERE %s", got)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("WhereError() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("WhereError() succeeded unexpectedly")
			}
			if query != tt.want {
				t.Errorf("WhereError() = %v, want %v", query, tt.want)
			}
		})
	}
}
