package stores

import (
	"github.com/tkahng/authgo/internal/database"
)

type AllStore struct {
	db             database.Dbx
	user           *DbUserStore
	account        *DbAccountStore
	token          *DbTokenStore
	teamGroup      *DbTeamGroupStore
	teamMember     *DbTeamMemberStore
	teamInvitation *DbTeamInvitationStore
	customer       *DbCustomerStore
	price          *DbPriceStore
	product        *DbProductStore
	subscription   *DbSubscriptionStore
}

var _ AllStoreInterface = (*AllStore)(nil)

type AllStoreInterface interface {
	User() DbUserStoreInterface
	Account() DbAccountStoreInterface
	Token() DbTokenStoreInterface
	TeamGroup() DbTeamGroupStoreInterface
	TeamMember() DbTeamMemberStoreInterface
	TeamInvitation() DbTeamInvitationStoreInterface
	Customer() DbCustomerStoreInterface
	Price() DbPriceStoreInterface
	Product() DbProductStoreInterface
	Subscription() DbSubscriptionStoreInterface
}

func NewAllStore(db database.Dbx) *AllStore {
	return &AllStore{
		db:             db,
		user:           NewDbUserStore(db),
		account:        NewDbAccountStore(db),
		token:          NewPostgresTokenStore(db),
		teamGroup:      NewDbTeamGroupStore(db),
		teamMember:     NewDbTeamMemberStore(db),
		teamInvitation: NewDbTeamInvitationStore(db),
		customer:       NewDbCustomerStore(db),
		price:          NewDbPriceStore(db),
		product:        NewDbProductStore(db),
		subscription:   NewDbSubscriptionStore(db),
	}
}
func (s *AllStore) RunInTx(fn func(*AllStore) error) error {
	return s.db.RunInTx(func(d database.Dbx) error {
		store := s.WithTx(d)
		if err := fn(store); err != nil {
			return err
		}
		return nil
	})
}
func (s *AllStore) WithTx(dbx database.Dbx) *AllStore {
	return &AllStore{
		user:           s.user.WithTx(dbx),
		account:        s.account.WithTx(dbx),
		token:          s.token.WithTx(dbx),
		teamGroup:      s.teamGroup.WithTx(dbx),
		teamMember:     s.teamMember.WithTx(dbx),
		teamInvitation: s.teamInvitation.WithTx(dbx),
		customer:       s.customer.WithTx(dbx),
		price:          s.price.WithTx(dbx),
		product:        s.product.WithTx(dbx),
		subscription:   s.subscription.WithTx(dbx),
	}
}
func (s *AllStore) User() DbUserStoreInterface {
	return s.user
}
func (s *AllStore) Account() DbAccountStoreInterface {
	return s.account
}
func (s *AllStore) Token() DbTokenStoreInterface {
	return s.token
}

// Customer implements AllStoreInterface.
func (s *AllStore) Customer() DbCustomerStoreInterface {
	return s.customer
}

// Price implements AllStoreInterface.
func (s *AllStore) Price() DbPriceStoreInterface {
	return s.price
}

// Product implements AllStoreInterface.
func (s *AllStore) Product() DbProductStoreInterface {
	return s.product
}

// Subscription implements AllStoreInterface.
func (s *AllStore) Subscription() DbSubscriptionStoreInterface {
	return s.subscription
}

// TeamGroup implements AllStoreInterface.
func (s *AllStore) TeamGroup() DbTeamGroupStoreInterface {
	return s.teamGroup
}

// TeamInvitation implements AllStoreInterface.
func (s *AllStore) TeamInvitation() DbTeamInvitationStoreInterface {
	return s.teamInvitation
}

// TeamMember implements AllStoreInterface.
func (s *AllStore) TeamMember() DbTeamMemberStoreInterface {
	return s.teamMember
}
