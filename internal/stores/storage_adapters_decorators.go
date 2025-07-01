package stores

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
		MediaFunc:          &DbMediaStore{},
	}
}

type StorageAdapterDecorator struct {
	Delegate           StorageAdapterInterface
	UserFunc           *UserStoreDecorator
	UserAccountFunc    *AccountStoreDecorator
	TokenFunc          *TokenStoreDecorator
	TeamGroupFunc      *TeamGroupStoreDecorator
	TeamInvitationFunc *TeamInvitationStoreDecorator
	TeamMemberFunc     *TeamMemberStoreDecorator
	MediaFunc          *DbMediaStore
	RbacFunc           *RbacStoreDecorator
	CustomerFunc       *CustomerStoreDecorator
	ProductFunc        *StripeProductStoreDecorator
	PriceFunc          *StripePriceStoreDecorator
	SubscriptionFunc   *StripeSubscriptionStoreDecorator
	TaskFunc           *TaskDecorator
	RunInTxFunc        func(fn func(tx StorageAdapterInterface) error) error
}

var _ StorageAdapterInterface = (*StorageAdapterDecorator)(nil)

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

// RunInTx implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) RunInTx(fn func(tx StorageAdapterInterface) error) error {
	if s.RunInTxFunc != nil {
		return s.RunInTxFunc(fn)
	}
	return s.Delegate.RunInTx(fn)
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

var _ StorageAdapterInterface = (*StorageAdapterDecorator)(nil)
