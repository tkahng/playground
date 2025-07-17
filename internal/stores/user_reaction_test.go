package stores

import (
	"context"
	"reflect"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestDbUserReactionStore_CountUserReactions(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDbUserReactionStore(db)
		var reactions []*models.UserReaction
		for range 10 {
			reaction, err := store.CreateUserReaction(ctx, &models.UserReaction{
				Reaction:  types.Pointer("hello"),
				Type:      "hello",
				IpAddress: types.Pointer("hello"),
				City:      types.Pointer("los angeles"),
				Country:   types.Pointer("us"),
			})
			if err != nil {
				t.Fatalf("Failed to create user reaction: %v", err)
			}
			reactions = append(reactions, reaction)
		}

		count, err := store.CountUserReactions(ctx, nil)
		if err != nil {
			t.Fatalf("Failed to count user reactions: %v", err)
		}
		if count != 10 {
			t.Fatalf("Expected 10 user reactions, got %d", count)
		}
	})
}

func TestNewDbUserReactionStore(t *testing.T) {
	type args struct {
		db database.Dbx
	}
	tests := []struct {
		name string
		args args
		want *DbUserReactionStore
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDbUserReactionStore(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDbUserReactionStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDbUserReactionStore_CreateUserReaction(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDbUserReactionStore(db)
		_, err := store.CreateUserReaction(ctx, &models.UserReaction{
			Reaction:  types.Pointer("hello"),
			Type:      "hello",
			IpAddress: types.Pointer("hello"),
			City:      types.Pointer("los angeles"),
			Country:   types.Pointer("us"),
		})
		if err != nil {
			t.Fatalf("Failed to count user reactions: %v", err)
		}

	})
}
