package stores

import "github.com/tkahng/authgo/internal/database"

type PostgresTeamInvitationStore struct {
	db database.Dbx
}
