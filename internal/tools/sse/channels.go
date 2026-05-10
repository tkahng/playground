package sse

const (
	UserReactionsChannel = "user-reactions"
)

func PlayerChannel(playerID string) string {
	return "player_id:" + playerID
}
