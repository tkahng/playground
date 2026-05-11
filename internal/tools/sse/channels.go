package sse

import (
	"strings"

	"github.com/google/uuid"
)

const (
	UserReactionsChannel = "user-reactions"
	playerChannelPrefix  = "player_id:"
)

func PlayerChannel(playerID string) string {
	return playerChannelPrefix + playerID
}

// PlayerIDFromChannel extracts the player UUID from a PlayerChannel string.
// Returns (uuid.Nil, false) if the channel is not a player channel or the UUID is malformed.
func PlayerIDFromChannel(channel string) (uuid.UUID, bool) {
	s, ok := strings.CutPrefix(channel, playerChannelPrefix)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
