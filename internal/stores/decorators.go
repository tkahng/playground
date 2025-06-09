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
