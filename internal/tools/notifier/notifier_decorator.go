package notifier

import "context"

type SubscriptionDecorator struct {
	Delegate          Subscription
	EstablishedCFunc  func() <-chan struct{}
	NotificationCFunc func() <-chan []byte
	UnlistenFunc      func(ctx context.Context)
}

// EstablishedC implements Subscription.
func (s *SubscriptionDecorator) EstablishedC() <-chan struct{} {
	if s.EstablishedCFunc != nil {
		return s.EstablishedCFunc()
	}
	if s.Delegate == nil {
		panic("delegate is nil in EstablishedC in SubscriptionDecorator")
	}
	return s.Delegate.EstablishedC()
}

// NotificationC implements Subscription.
func (s *SubscriptionDecorator) NotificationC() <-chan []byte {
	if s.NotificationCFunc != nil {
		return s.NotificationCFunc()
	}
	if s.Delegate == nil {
		panic("delegate is nil in NotificationC in SubscriptionDecorator")
	}
	return s.Delegate.NotificationC()
}

// Unlisten implements Subscription.
func (s *SubscriptionDecorator) Unlisten(ctx context.Context) {
	if s.UnlistenFunc != nil {
		s.UnlistenFunc(ctx)
	}
	if s.Delegate == nil {
		panic("delegate is nil in Unlisten in SubscriptionDecorator")
	}
	s.Delegate.Unlisten(ctx)
}

var _ Subscription = (*SubscriptionDecorator)(nil)

type NotifierDecorator struct {
	Delegate      Notifier
	RunFunc       func(ctx context.Context) error
	SubscribeFunc func(channel string) Subscription
}

// Run implements Notifier.
func (n *NotifierDecorator) Run(ctx context.Context) error {
	if n.RunFunc != nil {
		return n.RunFunc(ctx)
	}
	if n.Delegate == nil {
		panic("delegate is nil in Run in NotifierDecorator")
	}
	return n.Delegate.Run(ctx)
}

// Subscribe implements Notifier.
func (n *NotifierDecorator) Subscribe(channel string) Subscription {
	if n.SubscribeFunc != nil {
		return n.SubscribeFunc(channel)
	}
	if n.Delegate == nil {
		panic("delegate is nil in Subscribe in NotifierDecorator")
	}
	return n.Delegate.Subscribe(channel)
}

var _ Notifier = (*NotifierDecorator)(nil)
