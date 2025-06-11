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
		user:           &UserStoreDecorator{},
		userAccount:    &AccountStoreDecorator{},
		token:          &TokenStoreDecorator{},
		teamGroup:      &TeamGroupStoreDecorator{},
		teamInvitation: &TeamInvitationStoreDecorator{},
		teamMember:     &TeamMemberStoreDecorator{},
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
	user           *UserStoreDecorator
	userAccount    *AccountStoreDecorator
	token          *TokenStoreDecorator
	teamGroup      *TeamGroupStoreDecorator
	teamInvitation *TeamInvitationStoreDecorator
	teamMember     *TeamMemberStoreDecorator
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
	panic("unimplemented")
}

// Subscription implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Subscription() DbSubscriptionStoreInterface {
	panic("unimplemented")
}

// TeamGroup implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamGroup() DbTeamGroupStoreInterface {
	return s.teamGroup
}

// TeamInvitation implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamInvitation() DbTeamInvitationStoreInterface {
	return s.teamInvitation
}

// TeamMember implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) TeamMember() DbTeamMemberStoreInterface {
	return s.teamMember
}

// Token implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) Token() DbTokenStoreInterface {
	return s.token
}

// User implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) User() DbUserStoreInterface {
	return s.user
}

// UserAccount implements StorageAdapterInterface.
func (s *StorageAdapterDecorator) UserAccount() DbAccountStoreInterface {
	return s.userAccount
}

var _ StorageAdapterInterface = (*StorageAdapterDecorator)(nil)
