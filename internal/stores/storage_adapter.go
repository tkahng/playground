package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
)

var _ StorageAdapterInterface = (*StorageAdapter)(nil)

type StorageAdapterInterface interface {
	UserReaction() UserReactionStore
	Notification() NotificationStore
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
	Job() JobStore
	// WithTx(tx database.Dbx) *StorageAdapter
	RunInTxCtx(ctx context.Context, fn func(txCtx context.Context) error) error
	RunInTx(ctx context.Context, fn func(tx StorageAdapterInterface) error) error
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
	notification   *DbNotificationStore
	job            *DbJobStore
	userReaction   *DbUserReactionStore
}

// UserReaction implements StorageAdapterInterface.
func (s *StorageAdapter) UserReaction() UserReactionStore {
	return s.userReaction
}

func (s *StorageAdapter) Job() JobStore {
	return s.job
}
func (s *StorageAdapter) Notification() NotificationStore {
	return s.notification
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

// RunInTx implements StorageAdapterInterface.
func (s *StorageAdapter) RunInTx(ctx context.Context, fn func(tx StorageAdapterInterface) error) error {
	return s.db.RunInTx(ctx, func(db database.Dbx) error {
		tx := &StorageAdapter{
			db:             db,
			user:           s.user.WithTx(db),
			userAccount:    s.userAccount.WithTx(db),
			token:          s.token.WithTx(db),
			teamGroup:      s.teamGroup.WithTx(db),
			teamMember:     s.teamMember.WithTx(db),
			teamInvitation: s.teamInvitation.WithTx(db),
			customer:       s.customer.WithTx(db),
			price:          s.price.WithTx(db),
			product:        s.product.WithTx(db),
			subscription:   s.subscription.WithTx(db),
			rbac:           s.rbac.WithTx(db),
			task:           s.task.WithTx(db),
			media:          s.media.WithTx(db),
			notification:   s.notification.WithTx(db),
			job:            s.job.WithTx(db),
			userReaction:   s.userReaction.WithTx(db),
		}
		return fn(tx)
	})
}

func (s *StorageAdapter) RunInTxCtx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return s.db.RunInTxCtx(ctx, fn)
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
		job:            NewDbJobStore(db),
		media:          NewMediaStore(db),
		notification:   NewDbNotificationStore(db),
		userReaction:   NewDbUserReactionStore(db),
	}
}
