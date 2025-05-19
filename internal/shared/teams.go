package shared

import "github.com/tkahng/authgo/internal/models"

type TeamInfo struct {
	Team   models.Team       `json:"team"`
	Member models.TeamMember `json:"member"`
}
