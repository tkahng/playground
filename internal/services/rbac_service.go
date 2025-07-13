package services

import "github.com/tkahng/playground/internal/stores"

type RBACService interface {
	Adapter() stores.StorageAdapterInterface
}

type rbacService struct {
	adapter stores.StorageAdapterInterface
}

// Adapter implements RBACService.
func (r *rbacService) Adapter() stores.StorageAdapterInterface {
	return r.adapter
}

func NewRBACService(adapter stores.StorageAdapterInterface) RBACService {
	return &rbacService{
		adapter: adapter,
	}
}
