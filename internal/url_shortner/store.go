package urlshortner

import "context"

type ShortUrlFilter struct {
	ShortCode string
	SourceUrl string
}

type ShortUrlStore interface {
	FindByShortCode(ctx context.Context, shortCode string) (*ShortUrl, error)
	FindBySourceUrl(ctx context.Context, sourceUrl string) (*ShortUrl, error)
	SaveShortUrl(ctx context.Context, shortUrl *ShortUrl) error
}
