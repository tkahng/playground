package urlshortner

import (
	"context"
	"errors"

	"github.com/tkahng/playground/internal/tools/store"
)

var (
	ErrResourceNotImplemented = errors.New("delegate for InMemoryShortUrlStoreDecorator not implemented")
)

type InMemoryShortUrlStore struct {
	shortCodeStore *store.Store[string, *ShortUrl]
	sourceUrlStore *store.Store[string, *ShortUrl]
}

// CountShortUrls implements ShortUrlStore.
func (i *InMemoryShortUrlStore) CountShortUrls(ctx context.Context, filter *ShortUrlFilter) (int64, error) {
	count := i.shortCodeStore.Length()
	return int64(count), nil
}

// SaveShortUrl implements ShortUrlStore.
func (i *InMemoryShortUrlStore) SaveShortUrl(ctx context.Context, shortUrl *ShortUrl) error {
	i.shortCodeStore.Set(shortUrl.ShortCode, shortUrl)
	i.sourceUrlStore.Set(shortUrl.SourceUrl, shortUrl)
	return nil
}

// FindByShortCode implements ShortUrlStore.
func (i *InMemoryShortUrlStore) FindByShortCode(ctx context.Context, shortCode string) (*ShortUrl, error) {
	res := i.shortCodeStore.Get(shortCode)
	return res, nil
}

// FindBySourceUrl implements ShortUrlStore.
func (i *InMemoryShortUrlStore) FindBySourceUrl(ctx context.Context, sourceUrl string) (*ShortUrl, error) {
	res := i.sourceUrlStore.Get(sourceUrl)
	return res, nil
}

func NewInMemoryShortUrlStore() *InMemoryShortUrlStore {
	return &InMemoryShortUrlStore{
		shortCodeStore: store.New[string, *ShortUrl](nil),
		sourceUrlStore: store.New[string, *ShortUrl](nil),
	}
}

var _ ShortUrlStore = &InMemoryShortUrlStore{}

type InMemoryShortUrlStoreDecorator struct {
	Delegate            ShortUrlStore
	CountShortUrlsFunc  func(ctx context.Context, filter *ShortUrlFilter) (int64, error)
	FindByShortCodeFunc func(ctx context.Context, shortCode string) (*ShortUrl, error)
	FindBySourceUrlFunc func(ctx context.Context, sourceUrl string) (*ShortUrl, error)
	SaveShortUrlFunc    func(ctx context.Context, shortUrl *ShortUrl) error
}

func NewInMemoryShortUrlStoreDecorator() *InMemoryShortUrlStoreDecorator {
	return &InMemoryShortUrlStoreDecorator{
		Delegate: NewInMemoryShortUrlStore(),
	}
}

var _ ShortUrlStore = &InMemoryShortUrlStoreDecorator{}

// CountShortUrls implements ShortUrlStore.
func (i *InMemoryShortUrlStoreDecorator) CountShortUrls(ctx context.Context, filter *ShortUrlFilter) (int64, error) {
	if i.CountShortUrlsFunc != nil {
		return i.CountShortUrlsFunc(ctx, filter)
	}
	if i.Delegate == nil {
		return 0, ErrResourceNotImplemented
	}
	return i.Delegate.CountShortUrls(ctx, filter)
}

// FindByShortCode implements ShortUrlStore.
func (i *InMemoryShortUrlStoreDecorator) FindByShortCode(ctx context.Context, shortCode string) (*ShortUrl, error) {
	if i.FindByShortCodeFunc != nil {
		return i.FindByShortCodeFunc(ctx, shortCode)
	}
	if i.Delegate == nil {
		return nil, ErrResourceNotImplemented
	}
	return i.Delegate.FindByShortCode(ctx, shortCode)
}

// FindBySourceUrl implements ShortUrlStore.
func (i *InMemoryShortUrlStoreDecorator) FindBySourceUrl(ctx context.Context, sourceUrl string) (*ShortUrl, error) {
	if i.FindBySourceUrlFunc != nil {
		return i.FindBySourceUrlFunc(ctx, sourceUrl)
	}
	if i.Delegate == nil {
		return nil, ErrResourceNotImplemented
	}
	return i.Delegate.FindBySourceUrl(ctx, sourceUrl)
}

// SaveShortUrl implements ShortUrlStore.
func (i *InMemoryShortUrlStoreDecorator) SaveShortUrl(ctx context.Context, shortUrl *ShortUrl) error {
	if i.SaveShortUrlFunc != nil {
		return i.SaveShortUrlFunc(ctx, shortUrl)
	}
	if i.Delegate == nil {
		return ErrResourceNotImplemented
	}
	return i.Delegate.SaveShortUrl(ctx, shortUrl)
}
