package shared

import "github.com/tkahng/authgo/internal/models"

type TeamInfo struct {
	User   models.User       `json:"user"`
	Team   models.Team       `json:"team"`
	Member models.TeamMember `json:"member"`
}
