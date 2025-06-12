package services

import "github.com/tkahng/authgo/internal/stores"

type RBACService interface {
	Adapter() stores.StorageAdapterInterface
}

type rbacService struct {
	adapter *stores.StorageAdapter
}

// Adapter implements RBACService.
func (r *rbacService) Adapter() stores.StorageAdapterInterface {
	return r.adapter
}

func NewRBACService(adapter *stores.StorageAdapter) RBACService {
	return &rbacService{
		adapter: adapter,
	}
}
