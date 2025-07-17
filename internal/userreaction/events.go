package userreaction

import "github.com/tkahng/playground/internal/models"

type UserReactionCreatedEvent struct {
	UserReaction *models.UserReaction
}
