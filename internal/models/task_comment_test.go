package models_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/models"
)

func TestParseMentionedMemberIDs(t *testing.T) {
	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	id2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	tests := []struct {
		name    string
		content string
		want    []uuid.UUID
	}{
		{
			name:    "no mentions",
			content: "plain text with no mentions",
			want:    []uuid.UUID{},
		},
		{
			name:    "single mention",
			content: "@[Alice](" + id1.String() + ") please review",
			want:    []uuid.UUID{id1},
		},
		{
			name:    "multiple distinct mentions",
			content: "hey @[Alice](" + id1.String() + ") and @[Bob](" + id2.String() + ")",
			want:    []uuid.UUID{id1, id2},
		},
		{
			name:    "duplicate mention deduplicated",
			content: "@[Alice](" + id1.String() + ") and @[Alice](" + id1.String() + ") again",
			want:    []uuid.UUID{id1},
		},
		{
			name:    "invalid uuid ignored",
			content: "@[Bad](not-a-uuid) text",
			want:    []uuid.UUID{},
		},
		{
			name:    "mention embedded in larger text",
			content: "cc @[Alice](" + id1.String() + ") for follow-up",
			want:    []uuid.UUID{id1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := models.ParseMentionedMemberIDs(tt.content)
			if len(tt.want) == 0 {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
