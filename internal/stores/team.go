package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
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
	return database.WithTx(p.db, func(tx database.Dbx) error {
		adapters := p.WithTx(tx)

		return txFunc(adapters)
	})
}

// CreateTeamWithOwnerMember implements services.TeamStore.
func (s *DbTeamStore) CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	var teamInfo *shared.TeamInfoModel
	err := s.Transact(
		ctx,
		func(store *DbTeamStore) error {
			team, err := store.CreateTeam(ctx, name, slug)
			if err != nil {
				return err
			}
			if team == nil {
				return fmt.Errorf("team not found")
			}
			teamMember, err := store.CreateTeamMember(ctx, team.ID, userId, models.TeamMemberRoleOwner, true)
			if err != nil {
				return err
			}
			if teamMember == nil {
				return fmt.Errorf("team member not found")
			}
			teamInfo = &shared.TeamInfoModel{
				Team:   *team,
				Member: *teamMember,
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return teamInfo, nil
}
