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
	}
}

func (s *AllEmbeddedStores) WithTx(dbx database.Dbx) *AllEmbeddedStores {
	return &AllEmbeddedStores{
		db:                    dbx,
		DbUserStore:           s.DbUserStore.WithTx(dbx),
		DbAccountStore:        s.DbAccountStore.WithTx(dbx),
		DbTokenStore:          s.DbTokenStore.WithTx(dbx),
		DbTeamGroupStore:      s.DbTeamGroupStore.WithTx(dbx),
		DbTeamMemberStore:     s.DbTeamMemberStore.WithTx(dbx),
		DbTeamInvitationStore: s.DbTeamInvitationStore.WithTx(dbx),
		DbCustomerStore:       s.DbCustomerStore.WithTx(dbx),
		DbPriceStore:          s.DbPriceStore.WithTx(dbx),
		DbProductStore:        s.DbProductStore.WithTx(dbx),
		DbSubscriptionStore:   s.DbSubscriptionStore.WithTx(dbx),
		DbTaskStore:           s.DbTaskStore.WithTx(dbx),
	}
}

type AllEmbeddedStoresInterface interface {
	DbUserStoreInterface
	DbAccountStoreInterface
	DbTokenStoreInterface
	DbTeamGroupStoreInterface
	DbTeamMemberStoreInterface
	DbTeamInvitationStoreInterface
	DbCustomerStoreInterface
	DbPriceStoreInterface
	DbProductStoreInterface
	DbSubscriptionStoreInterface
}
