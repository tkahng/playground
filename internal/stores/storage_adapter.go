package stores

import "github.com/tkahng/authgo/internal/database"

var _ StorageAdapterInterface = (*StorageAdapter)(nil)

type StorageAdapterInterface interface {
	User() DbUserStoreInterface
	UserAccount() DbAccountStoreInterface
	Token() DbTokenStoreInterface
	TeamGroup() DbTeamGroupStoreInterface
	TeamMember() DbTeamMemberStoreInterface
	TeamInvitation() DbTeamInvitationStoreInterface
	Customer() DbCustomerStoreInterface
	Price() DbPriceStoreInterface
	Product() DbProductStoreInterface
	Subscription() DbSubscriptionStoreInterface
	Media() MediaStoreInterface
	Rbac() DbRbacStoreInterface
	Task() DbTaskStoreInterface
	// WithTx(tx database.Dbx) *StorageAdapter
	RunInTx(fn func(tx StorageAdapterInterface) error) error
}
type StorageAdapter struct {
	db             database.Dbx
	user           *DbUserStore
	userAccount    *DbAccountStore
	token          *DbTokenStore
	teamGroup      *DbTeamGroupStore
	teamMember     *DbTeamMemberStore
	teamInvitation *DbTeamInvitationStore
	customer       *DbCustomerStore
	price          *DbPriceStore
	product        *DbProductStore
	subscription   *DbSubscriptionStore
	rbac           *DbRbacStore
	task           *DbTaskStore
	media          *DbMediaStore
}

func (s *StorageAdapter) Media() MediaStoreInterface {
	return s.media
}

func (s *StorageAdapter) Task() DbTaskStoreInterface {
	return s.task
}

// Customer implements StorageAdapterInterface.
func (s *StorageAdapter) Customer() DbCustomerStoreInterface {
	return s.customer
}

// Price implements StorageAdapterInterface.
func (s *StorageAdapter) Price() DbPriceStoreInterface {
	return s.price
}

// Product implements StorageAdapterInterface.
func (s *StorageAdapter) Product() DbProductStoreInterface {
	return s.product
}
func (s *StorageAdapter) WithTx(tx database.Dbx) *StorageAdapter {
	return &StorageAdapter{
		db:             tx,
		user:           s.user.WithTx(tx),
		userAccount:    s.userAccount.WithTx(tx),
		token:          s.token.WithTx(tx),
		teamGroup:      s.teamGroup.WithTx(tx),
		teamMember:     s.teamMember.WithTx(tx),
		teamInvitation: s.teamInvitation.WithTx(tx),
		customer:       s.customer.WithTx(tx),
		price:          s.price.WithTx(tx),
		product:        s.product.WithTx(tx),
		subscription:   s.subscription.WithTx(tx),
		rbac:           s.rbac.WithTx(tx),
	}
}

// RunInTx implements StorageAdapterInterface.
func (s *StorageAdapter) RunInTx(fn func(tx StorageAdapterInterface) error) error {
	return s.db.RunInTx(func(d database.Dbx) error {
		tx := s.WithTx(d)
		return fn(tx)
	})
}

func (s *StorageAdapter) Rbac() DbRbacStoreInterface {
	return s.rbac
}

// Subscription implements StorageAdapterInterface.
func (s *StorageAdapter) Subscription() DbSubscriptionStoreInterface {
	return s.subscription
}

// TeamGroup implements StorageAdapterInterface.
func (s *StorageAdapter) TeamGroup() DbTeamGroupStoreInterface {
	return s.teamGroup
}

// TeamInvitation implements StorageAdapterInterface.
func (s *StorageAdapter) TeamInvitation() DbTeamInvitationStoreInterface {
	return s.teamInvitation
}

// TeamMember implements StorageAdapterInterface.
func (s *StorageAdapter) TeamMember() DbTeamMemberStoreInterface {
	return s.teamMember
}

// Token implements StorageAdapterInterface.
func (s *StorageAdapter) Token() DbTokenStoreInterface {
	return s.token
}

// User implements StorageAdapterInterface.
func (s *StorageAdapter) User() DbUserStoreInterface {
	return s.user
}

// UserAccount implements StorageAdapterInterface.
func (s *StorageAdapter) UserAccount() DbAccountStoreInterface {
	return s.userAccount
}

func NewStorageAdapter(db database.Dbx) *StorageAdapter {
	return &StorageAdapter{
		db:             db,
		user:           NewDbUserStore(db),
		userAccount:    NewDbAccountStore(db),
		token:          NewPostgresTokenStore(db),
		teamGroup:      NewDbTeamGroupStore(db),
		teamMember:     NewDbTeamMemberStore(db),
		teamInvitation: NewDbTeamInvitationStore(db),
		customer:       NewDbCustomerStore(db),
		price:          NewDbPriceStore(db),
		product:        NewDbProductStore(db),
		subscription:   NewDbSubscriptionStore(db),
		rbac:           NewDbRBACStore(db),
		task:           NewDbTaskStore(db),
	}
}
