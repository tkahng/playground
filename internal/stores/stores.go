package stores

import (
	"github.com/tkahng/authgo/internal/database"
)

type AllEmbeddedStores struct {
	db database.Dbx
	*DbUserStore
	*DbAccountStore
	*DbTokenStore
	*DbTeamGroupStore
	*DbTeamMemberStore
	*DbTeamInvitationStore
	*DbCustomerStore
	*DbPriceStore
	*DbProductStore
	*DbSubscriptionStore
	*DbTaskStore
	*DbConstraintStore
}

func NewAllEmbeddedStores(db database.Dbx) *AllEmbeddedStores {
	return &AllEmbeddedStores{
		db:                    db,
		DbUserStore:           NewDbUserStore(db),
		DbAccountStore:        NewDbAccountStore(db),
		DbTokenStore:          NewPostgresTokenStore(db),
		DbTeamGroupStore:      NewDbTeamGroupStore(db),
		DbTeamMemberStore:     NewDbTeamMemberStore(db),
		DbTeamInvitationStore: NewDbTeamInvitationStore(db),
		DbCustomerStore:       NewDbCustomerStore(db),
		DbPriceStore:          NewDbPriceStore(db),
		DbProductStore:        NewDbProductStore(db),
		DbSubscriptionStore:   NewDbSubscriptionStore(db),
		DbTaskStore:           NewDbTaskStore(db),
		DbConstraintStore:     NewDbConstraintStore(db),
	}
}
