package urlshortner

import (
	"context"

	"github.com/tkahng/playground/internal/tools/store"
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
