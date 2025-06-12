package stores

import (
	"context"

	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/models"
)

type StripeSubscriptionStoreDecorator struct {
	Delegate                                   *DbSubscriptionStore
	CountSubscriptionsFunc                     func(ctx context.Context, filter *StripeSubscriptionListFilter) (int64, error)
	FindActiveSubscriptionByCustomerIdFunc     func(ctx context.Context, customerId string) (*models.StripeSubscription, error)
	FindActiveSubscriptionsByCustomerIdsFunc   func(ctx context.Context, customerIds ...string) ([]*models.StripeSubscription, error)
	FindActiveSubscriptionsByTeamIdsFunc       func(ctx context.Context, teamIds ...uuid.UUID) ([]*models.StripeSubscription, error)
	FindActiveSubscriptionsByUserIdsFunc       func(ctx context.Context, userIds ...uuid.UUID) ([]*models.StripeSubscription, error)
	FindSubscriptionsWithPriceProductByIdsFunc func(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error)
	IsFirstSubscriptionFunc                    func(ctx context.Context, customerID string) (bool, error)
	ListSubscriptionsFunc                      func(ctx context.Context, input *StripeSubscriptionListFilter) ([]*models.StripeSubscription, error)
	UpsertSubscriptionFunc                     func(ctx context.Context, sub *models.StripeSubscription) error
	UpsertSubscriptionFromStripeFunc           func(ctx context.Context, sub *stripe.Subscription) error
}

// CountSubscriptions implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) CountSubscriptions(ctx context.Context, filter *StripeSubscriptionListFilter) (int64, error) {
	if s.CountSubscriptionsFunc != nil {
		return s.CountSubscriptionsFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return s.Delegate.CountSubscriptions(ctx, filter)
}

// FindActiveSubscriptionByCustomerId implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) FindActiveSubscriptionByCustomerId(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
	if s.FindActiveSubscriptionByCustomerIdFunc != nil {
		return s.FindActiveSubscriptionByCustomerIdFunc(ctx, customerId)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindActiveSubscriptionByCustomerId(ctx, customerId)
}

// FindActiveSubscriptionsByCustomerIds implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) FindActiveSubscriptionsByCustomerIds(ctx context.Context, customerIds ...string) ([]*models.StripeSubscription, error) {
	if s.FindActiveSubscriptionsByCustomerIdsFunc != nil {
		return s.FindActiveSubscriptionsByCustomerIdsFunc(ctx, customerIds...)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindActiveSubscriptionsByCustomerIds(ctx, customerIds...)
}

// FindActiveSubscriptionsByTeamIds implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) FindActiveSubscriptionsByTeamIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
	if s.FindActiveSubscriptionsByTeamIdsFunc != nil {
		return s.FindActiveSubscriptionsByTeamIdsFunc(ctx, teamIds...)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindActiveSubscriptionsByTeamIds(ctx, teamIds...)
}

// FindActiveSubscriptionsByUserIds implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) FindActiveSubscriptionsByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
	if s.FindActiveSubscriptionsByUserIdsFunc != nil {
		return s.FindActiveSubscriptionsByUserIdsFunc(ctx, userIds...)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindActiveSubscriptionsByUserIds(ctx, userIds...)
}

// FindSubscriptionsWithPriceProductByIds implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) FindSubscriptionsWithPriceProductByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error) {
	if s.FindSubscriptionsWithPriceProductByIdsFunc != nil {
		return s.FindSubscriptionsWithPriceProductByIdsFunc(ctx, subscriptionIds...)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindSubscriptionsWithPriceProductByIds(ctx, subscriptionIds...)
}

// IsFirstSubscription implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) IsFirstSubscription(ctx context.Context, customerID string) (bool, error) {
	if s.IsFirstSubscriptionFunc != nil {
		return s.IsFirstSubscriptionFunc(ctx, customerID)
	}
	if s.Delegate == nil {
		return false, ErrDelegateNil
	}
	return s.Delegate.IsFirstSubscription(ctx, customerID)
}

// ListSubscriptions implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) ListSubscriptions(ctx context.Context, input *StripeSubscriptionListFilter) ([]*models.StripeSubscription, error) {
	if s.ListSubscriptionsFunc != nil {
		return s.ListSubscriptionsFunc(ctx, input)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.ListSubscriptions(ctx, input)
}

// UpsertSubscription implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error {
	if s.UpsertSubscriptionFunc != nil {
		return s.UpsertSubscriptionFunc(ctx, sub)
	}
	if s.Delegate == nil {
		return ErrDelegateNil
	}
	return s.Delegate.UpsertSubscription(ctx, sub)
}

// UpsertSubscriptionFromStripe implements DbSubscriptionStoreInterface.
func (s *StripeSubscriptionStoreDecorator) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error {
	if s.UpsertSubscriptionFromStripeFunc != nil {
		return s.UpsertSubscriptionFromStripeFunc(ctx, sub)
	}
	if s.Delegate == nil {
		return ErrDelegateNil
	}
	return s.Delegate.UpsertSubscriptionFromStripe(ctx, sub)
}

var _ DbSubscriptionStoreInterface = (*StripeSubscriptionStoreDecorator)(nil)
