package stores

import (
	"context"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/services"
)

type DbTeamStore struct {
	db database.Dbx
	*DbUserStore
	*DbTeamGroupStore
	*DbTeamMemberStore
	*DbTeamInvitationStore
}

func (s *DbTeamStore) WithTx(db database.Dbx) *DbTeamStore {
	return &DbTeamStore{

		db:                    db,
		DbUserStore:           s.DbUserStore.WithTx(db),
		DbTeamGroupStore:      s.DbTeamGroupStore.WithTx(db),
		DbTeamMemberStore:     s.DbTeamMemberStore.WithTx(db),
		DbTeamInvitationStore: s.DbTeamInvitationStore.WithTx(db),
	}
}

func NewDbTeamStore(db database.Dbx) *DbTeamStore {
	return &DbTeamStore{
		db:                    db,
		DbUserStore:           NewDbUserStore(db),
		DbTeamGroupStore:      NewDbTeamGroupStore(db),
		DbTeamMemberStore:     NewDbTeamMemberStore(db),
		DbTeamInvitationStore: NewDbTeamInvitationStore(db),
	}
}

func (p *DbTeamStore) Transact(ctx context.Context, txFunc func(adapters *DbTeamStore) error) error {
	return database.WithTx(ctx, p.db, func(tx database.Dbx) error {
		adapters := p.WithTx(tx)

		return txFunc(adapters)
	})
}

var _ services.TeamStore = &DbTeamStore{}
