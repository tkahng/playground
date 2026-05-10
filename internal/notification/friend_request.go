package notification

import "github.com/google/uuid"

type FriendRequestNotificationData struct {
	RequestingPlayerID uuid.UUID `json:"requesting_player_id" required:"true"`
	RequestingEmail    string    `json:"requesting_email" required:"true"`
	FriendshipID       uuid.UUID `json:"friendship_id" required:"true"`
}

func (FriendRequestNotificationData) Kind() string {
	return "friend_request"
}
