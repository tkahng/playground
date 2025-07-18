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

func FromModelUserReaction(ur *models.UserReaction) *UserReaction {
	if ur == nil {
		return nil
	}
	if ur.IpAddress == nil || ur.City == nil || ur.Country == nil {
		return nil
	}
	return &UserReaction{
		ID:        ur.ID,
		Type:      ur.Type,
		IpAddress: *ur.IpAddress,
		Country:   *ur.Country,
		City:      *ur.City,
		CreatedAt: ur.CreatedAt,
		UpdatedAt: ur.UpdatedAt,
	}
}
