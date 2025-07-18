package userreaction

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
)

type ReactionByCountry struct {
	Country        string `json:"country"`
	TotalReactions int64  `json:"total_reactions"`
}

type UserReactionStats struct {
	TotalReactions   int64               `json:"total_reactions"`
	TopFiveCountries []ReactionByCountry `json:"top_five_countries"`
	LastCreated      *UserReaction       `json:"last_created" required:"false"`
}

type UserReaction struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Type      string    `db:"type" json:"type"`
	IpAddress string    `db:"ip_address" json:"ip_address"`
	Country   string    `db:"country" json:"country"`
	City      string    `db:"city" json:"city"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func FromModelUserReaction(userReaction *models.UserReaction) *UserReaction {
	if userReaction == nil {
		return nil
	}
	if userReaction.IpAddress == nil || userReaction.City == nil || userReaction.Country == nil {
		return nil
	}
	return &UserReaction{
		ID:        userReaction.ID,
		Type:      userReaction.Type,
		IpAddress: *userReaction.IpAddress,
		Country:   *userReaction.Country,
		City:      *userReaction.City,
		CreatedAt: userReaction.CreatedAt,
		UpdatedAt: userReaction.UpdatedAt,
	}
}
