package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
)

func NewAdapterDecorators() *StorageAdapterDecorator {
	return &StorageAdapterDecorator{
		UserFunc:           &UserStoreDecorator{},
		UserAccountFunc:    &AccountStoreDecorator{},
		TokenFunc:          &TokenStoreDecorator{},
		TeamGroupFunc:      &TeamGroupStoreDecorator{},
		TeamInvitationFunc: &TeamInvitationStoreDecorator{},
		TeamMemberFunc:     &TeamMemberStoreDecorator{},
		RbacFunc:           &RbacStoreDecorator{},
		CustomerFunc:       &CustomerStoreDecorator{},
		ProductFunc:        &StripeProductStoreDecorator{},
		PriceFunc:          &StripePriceStoreDecorator{},
		SubscriptionFunc:   &StripeSubscriptionStoreDecorator{},
		TaskFunc:           &TaskDecorator{},
		MediaFunc:          &MediaStoreDecorator{},
		NotificationFunc:   &NotificationStoreDecorator{},
		Delegate:           &StorageAdapter{},
		JobFunc:            &JobStoreDecorator{},
		UserReactionFunc:   &DbUserReactionStoreDectorator{},
		GamingFunc:         &DbGamingStoreDecorator{},
		GisFunc:            &DBGisStoreDecorator{},
		LedgerFunc:         &DbLedgerStoreDecorator{},
		AiUsageFunc:        &AiUsageStoreDecorator{},
	}
}

func NewDbAdapterDecorators(db database.Dbx) *StorageAdapterDecorator {
	return &StorageAdapterDecorator{
		UserFunc:           NewUserStoreDecorator(db),
		UserAccountFunc:    NewAccountStoreDecorator(db),
		TokenFunc:          NewTokenStoreDecorator(db),
		TeamGroupFunc:      NewTeamGroupStoreDecorator(db),
		TeamInvitationFunc: NewTeamInvitationStoreDecorator(db),
		TeamMemberFunc:     NewTeamMemberStoreDecorator(db),
		RbacFunc:           NewRbacStoreDecorator(db),
		CustomerFunc:       NewCustomerStoreDecorator(db),
		ProductFunc:        NewStripeProductStoreDecorator(db),
		PriceFunc:          NewStripePriceStoreDecorator(db),
		SubscriptionFunc:   NewStripeSubscriptionStoreDecorator(db),
		TaskFunc:           NewTaskDecorator(db),
		MediaFunc:          NewDbMediaStoreDecorator(db),
		Delegate:           NewStorageAdapter(db),
		NotificationFunc:   NewNotificationStoreDecorator(db),
		JobFunc: &JobStoreDecorator{
			Delegate: NewDbJobStore(db),
		},
		UserReactionFunc: NewDbUserReactionStoreDectorator(db),
		GisFunc:          NewDBGisStoreDecorator(db),
		LedgerFunc:       NewDbLedgerStoreDecorator(db),
		AiUsageFunc:      NewAiUsageStoreDecorator(db),
	}
}

func NewCustomerStoreDecorator(db database.Dbx) *CustomerStoreDecorator {
	return &CustomerStoreDecorator{
		Delegate: NewDbCustomerStore(db),
	}
}

func NewStripeProductStoreDecorator(db database.Dbx) *StripeProductStoreDecorator {
	return &StripeProductStoreDecorator{
		Delegate: NewDbProductStore(db),
	}
}

func NewStripePriceStoreDecorator(db database.Dbx) *StripePriceStoreDecorator {
	return &StripePriceStoreDecorator{
		Delegate: NewDbPriceStore(db),
	}
}

func NewStripeSubscriptionStoreDecorator(db database.Dbx) *StripeSubscriptionStoreDecorator {
	return &StripeSubscriptionStoreDecorator{
		Delegate: NewDbSubscriptionStore(db),
	}
}

func NewTaskDecorator(db database.Dbx) *TaskDecorator {
	return &TaskDecorator{
		Delegate: NewDbTaskStore(db),
	}
}

func NewDbMediaStoreDecorator(db database.Dbx) *MediaStoreDecorator {
	return &MediaStoreDecorator{
		Delegate: NewMediaStore(db),
	}
}

type StorageAdapterDecorator struct {
	Delegate           StorageAdapterInterface
	NotificationFunc   *NotificationStoreDecorator
	UserFunc           *UserStoreDecorator
	UserAccountFunc    *AccountStoreDecorator
	TokenFunc          *TokenStoreDecorator
	TeamGroupFunc      *TeamGroupStoreDecorator
	TeamInvitationFunc *TeamInvitationStoreDecorator
	TeamMemberFunc     *TeamMemberStoreDecorator
	MediaFunc          *MediaStoreDecorator
	RbacFunc           *RbacStoreDecorator
	CustomerFunc       *CustomerStoreDecorator
	ProductFunc        *StripeProductStoreDecorator
	PriceFunc          *StripePriceStoreDecorator
	SubscriptionFunc   *StripeSubscriptionStoreDecorator
	TaskFunc           *TaskDecorator
	RunInTxFunc        func(ctx context.Context, fn func(tx StorageAdapterInterface) error) error
	RunInTxCtxFunc     func(ctx context.Context, fn func(ctx context.Context) error) error
	JobFunc            *JobStoreDecorator
	GamingFunc         *DbGamingStoreDecorator
	UserReactionFunc   *DbUserReactionStoreDectorator
	GisFunc            *DBGisStoreDecorator
	LedgerFunc         *DbLedgerStoreDecorator
	AiUsageFunc        *AiUsageStoreDecorator
}

// Gis implements [StorageAdapterInterface].
func (s *StorageAdapterDecorator) Gis() GisStore {
	if s.GisFunc != nil {
		return s.GisFunc
	}
	return s.Delegate.Gis()
}

// Ledger implements [StorageAdapterInterface].
func (s *StorageAdapterDecorator) Ledger() LedgerStore {
	if s.LedgerFunc != nil {
		return s.LedgerFunc
	}
	return s.Delegate.Ledger()
}

// Gaming implements [StorageAdapterInterface].
func (s *StorageAdapterDecorator) Gaming() GamingStore {
	if s.GamingFunc != nil {
		return s.GamingFunc
	}
	return s.Delegate.Gaming()
}

// RunInTx implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) RunInTx(ctx context.Context, fn func(tx StorageAdapterInterface) error) error {
	if s.RunInTxFunc != nil {
		return s.RunInTxFunc(ctx, fn)
	}
	return s.Delegate.RunInTx(ctx, fn)
}

// RunInTxCtx implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) RunInTxCtx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if s.RunInTxCtxFunc != nil {
		return s.RunInTxCtxFunc(ctx, fn)
	}
	return s.Delegate.RunInTxCtx(ctx, fn)
}

// UserReaction implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) UserReaction() UserReactionStore {
	if s.UserReactionFunc != nil {
		return s.UserReactionFunc
	}
	return s.Delegate.UserReaction()
}

// Job implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Job() JobStore {
	if s.JobFunc != nil {
		return s.JobFunc
	}
	return s.Delegate.Job()
}

var _ StorageAdapterInterface = (*StorageAdapterDecorator)(nil)

func (s *StorageAdapterDecorator) Notification() NotificationStore {
	if s.NotificationFunc != nil {
		return s.NotificationFunc
	}
	return s.Delegate.Notification()
}

// Media implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Media() MediaStoreInterface {
	if s.MediaFunc != nil {
		return s.MediaFunc
	}
	return s.Delegate.Media()
}

// Rbac implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Rbac() DbRbacStoreInterface {
	if s.RbacFunc != nil {
		return s.RbacFunc
	}
	return s.Delegate.Rbac()
}

func (s *StorageAdapterDecorator) Cleanup() {
	if s == nil {
		return
	}
	if s.UserFunc != nil {
		s.UserFunc.Cleanup()
	}
	if s.UserAccountFunc != nil {
		s.UserAccountFunc.Cleanup()
	}
	if s.TokenFunc != nil {
		s.TokenFunc.Cleanup()
	}
	if s.TeamGroupFunc != nil {
		s.TeamGroupFunc.Cleanup()
	}
	if s.TeamInvitationFunc != nil {
		s.TeamInvitationFunc.Cleanup()
	}
	if s.TeamMemberFunc != nil {
		s.TeamMemberFunc.Cleanup()
	}
	if s.RbacFunc != nil {
		s.RbacFunc.Cleanup()
	}
	if s.CustomerFunc != nil {
		s.CustomerFunc.Cleanup()
	}
	if s.ProductFunc != nil {
		s.ProductFunc.Cleanup()
	}
	if s.PriceFunc != nil {
		s.PriceFunc.Cleanup()
	}
	if s.SubscriptionFunc != nil {
		s.SubscriptionFunc.Cleanup()
	}
	if s.RunInTxFunc != nil {
		s.RunInTxFunc = nil // Clear the function to avoid memory leaks
	}
	if s.RunInTxCtxFunc != nil {
		s.RunInTxCtxFunc = nil // Clear the function to avoid memory leaks
	}
}

// Customer implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Customer() DbCustomerStoreInterface {
	if s.CustomerFunc != nil {
		return s.CustomerFunc
	}
	return s.Delegate.Customer()
}

// Price implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Price() DbPriceStoreInterface {
	if s.PriceFunc != nil {
		return s.PriceFunc
	}
	return s.Delegate.Price()
}

// Product implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Product() DbProductStoreInterface {
	if s.ProductFunc != nil {
		return s.ProductFunc
	}
	return s.Delegate.Product()
}

// Subscription implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Subscription() DbSubscriptionStoreInterface {
	if s.SubscriptionFunc != nil {
		return s.SubscriptionFunc
	}
	return s.Delegate.Subscription()
}

// TeamGroup implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamGroup() DbTeamGroupStoreInterface {
	if s.TeamGroupFunc != nil {
		return s.TeamGroupFunc
	}
	return s.Delegate.TeamGroup()
}

// TeamInvitation implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamInvitation() DbTeamInvitationStoreInterface {
	if s.TeamInvitationFunc != nil {
		return s.TeamInvitationFunc
	}
	return s.Delegate.TeamInvitation()
}

// TeamMember implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamMember() DbTeamMemberStoreInterface {
	if s.TeamMemberFunc != nil {
		return s.TeamMemberFunc
	}
	return s.Delegate.TeamMember()
}

// Token implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Token() DbTokenStoreInterface {
	if s.TokenFunc != nil {
		return s.TokenFunc
	}
	return s.Delegate.Token()
}

// User implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) User() DbUserStoreInterface {
	if s.UserFunc != nil {
		return s.UserFunc
	}
	return s.Delegate.User()
}

// UserAccount implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) UserAccount() DbAccountStoreInterface {
	if s.UserAccountFunc != nil {
		return s.UserAccountFunc
	}
	return s.Delegate.UserAccount()
}

func (s *StorageAdapterDecorator) Task() DbTaskStoreInterface {
	if s.TaskFunc != nil {
		return s.TaskFunc
	}
	return s.Delegate.Task()
}

func (s *StorageAdapterDecorator) AiUsage() AiUsageStoreInterface {
	if s.AiUsageFunc != nil {
		return s.AiUsageFunc
	}
	return s.Delegate.AiUsage()
}

var _ StorageAdapterInterface = (*StorageAdapterDecorator)(nil)
