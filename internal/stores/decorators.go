package stores

type StoreDecorators struct {
	*UserStoreDecorator
	*AccountStoreDecorator
	*TokenStoreDecorator
}

func NewStoreDecorators() *StoreDecorators {
	return &StoreDecorators{
		UserStoreDecorator:    &UserStoreDecorator{},
		AccountStoreDecorator: &AccountStoreDecorator{},
		TokenStoreDecorator:   &TokenStoreDecorator{},
	}
}

func NewAdapterDecorators() *StorageAdapterDecorator {
	return &StorageAdapterDecorator{
		UserFunc:           &UserStoreDecorator{},
		UserAccountFunc:    &AccountStoreDecorator{},
		TokenFunc:          &TokenStoreDecorator{},
		TeamGroupFunc:      &TeamGroupStoreDecorator{},
		TeamInvitationFunc: &TeamInvitationStoreDecorator{},
		TeamMemberFunc:     &TeamMemberStoreDecorator{},
	}
}

func (s *StoreDecorators) Cleanup() {
	if s == nil {
		return
	}
	if s.UserStoreDecorator != nil {
		s.UserStoreDecorator.Cleanup()
	}
	if s.AccountStoreDecorator != nil {
		s.AccountStoreDecorator.Cleanup()
	}
	if s.TokenStoreDecorator != nil {
		s.TokenStoreDecorator.Cleanup()
	}
}

type StorageAdapterDecorator struct {
	UserFunc           *UserStoreDecorator
	UserAccountFunc    *AccountStoreDecorator
	TokenFunc          *TokenStoreDecorator
	TeamGroupFunc      *TeamGroupStoreDecorator
	TeamInvitationFunc *TeamInvitationStoreDecorator
	TeamMemberFunc     *TeamMemberStoreDecorator
	RunInTxFunc        func(fn func(tx StorageAdapterInterface) error) error
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
}

// Customer implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Customer() DbCustomerStoreInterface {
	panic("unimplemented")
}

// Price implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Price() DbPriceStoreInterface {
	panic("unimplemented")
}

// Product implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Product() DbProductStoreInterface {
	panic("unimplemented")
}

// RunInTx implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) RunInTx(fn func(tx StorageAdapterInterface) error) error {
	if s.RunInTxFunc != nil {
		return s.RunInTxFunc(fn)
	}
	return ErrDelegateNil
}

// Subscription implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Subscription() DbSubscriptionStoreInterface {
	panic("unimplemented")
}

// TeamGroup implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamGroup() DbTeamGroupStoreInterface {
	return s.TeamGroupFunc
}

// TeamInvitation implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamInvitation() DbTeamInvitationStoreInterface {
	return s.TeamInvitationFunc
}

// TeamMember implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamMember() DbTeamMemberStoreInterface {
	return s.TeamMemberFunc
}

// Token implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Token() DbTokenStoreInterface {
	return s.TokenFunc
}

// User implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) User() DbUserStoreInterface {
	return s.UserFunc
}

// UserAccount implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) UserAccount() DbAccountStoreInterface {
	return s.UserAccountFunc
}

var _ StorageAdapterInterface = (*StorageAdapterDecorator)(nil)
