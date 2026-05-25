package sse_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/tools/sse"
)

func TestPlayerChannel_RoundTrip(t *testing.T) {
	id := uuid.New()
	channel := sse.PlayerChannel(id.String())
	parsed, ok := sse.PlayerIDFromChannel(channel)
	require.True(t, ok)
	assert.Equal(t, id, parsed)
}

func TestPlayerIDFromChannel_WrongPrefix(t *testing.T) {
	_, ok := sse.PlayerIDFromChannel("user_id:some-uuid")
	assert.False(t, ok)
}

func TestPlayerIDFromChannel_EmptyString(t *testing.T) {
	_, ok := sse.PlayerIDFromChannel("")
	assert.False(t, ok)
}

func TestPlayerIDFromChannel_InvalidUUID(t *testing.T) {
	_, ok := sse.PlayerIDFromChannel("player_id:not-a-uuid")
	assert.False(t, ok)
}

func TestPlayerIDFromChannel_PrefixOnly(t *testing.T) {
	_, ok := sse.PlayerIDFromChannel("player_id:")
	assert.False(t, ok)
}
