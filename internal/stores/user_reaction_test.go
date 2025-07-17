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

func TestDbUserReactionStore_CountByCountry(t *testing.T) {
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
		for range 5 {
			reaction, err := store.CreateUserReaction(ctx, &models.UserReaction{
				Reaction:  types.Pointer("hello"),
				Type:      "hello",
				IpAddress: types.Pointer("hello"),
				City:      types.Pointer("paris"),
				Country:   types.Pointer("fr"),
			})
			if err != nil {
				t.Fatalf("Failed to create user reaction: %v", err)
			}
			reactions = append(reactions, reaction)
		}
		for range 3 {
			reaction, err := store.CreateUserReaction(ctx, &models.UserReaction{
				Reaction:  types.Pointer("hello"),
				Type:      "hello",
				IpAddress: types.Pointer("hello"),
				City:      types.Pointer("beijing"),
				Country:   types.Pointer("cn"),
			})
			if err != nil {
				t.Fatalf("Failed to create user reaction: %v", err)
			}
			reactions = append(reactions, reaction)
		}

		countryCounts, err := store.CountByCountry(ctx, &UserReactionFilter{
			PaginatedInput: PaginatedInput{
				Page:    0,
				PerPage: 3,
			},
		})
		if err != nil {
			t.Fatalf("Failed to count user reactions: %v", err)
		}
		if len(countryCounts) != 3 {
			t.Fatalf("Expected 10 user reactions, got %d", len(countryCounts))
		}
		for _, c := range countryCounts {
			switch c.Country {
			case "us":
				if c.TotalReactions != 10 {
					t.Fatalf("Expected 10 user reactions, got %d", c.TotalReactions)
				}
			case "fr":
				if c.TotalReactions != 5 {
					t.Fatalf("Expected 5 user reactions, got %d", c.TotalReactions)
				}
			case "cn":
				if c.TotalReactions != 3 {
					t.Fatalf("Expected 3 user reactions, got %d", c.TotalReactions)
				}

			}
		}
	})
}
