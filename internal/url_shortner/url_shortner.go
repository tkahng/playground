package urlshortner

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/tkahng/playground/internal/tools/security"
)

type UrlShortnerOptions struct {
	RetryCount    int
	CodeMinLength int
}

type UrlShortner struct {
	store ShortUrlStore
	opt   UrlShortnerOptions
}

type UrlShortnerStub interface {
	CalculateMinimumLength(n int64) int64
	GenerateShortCode(ctx context.Context) (string, error)
	GenerateShortUrl(ctx context.Context, sourceUrl string) (*ShortUrl, error)
}

func NewUrlShortner(store ShortUrlStore) *UrlShortner {
	return &UrlShortner{
		store: store,
		opt: UrlShortnerOptions{
			RetryCount:    3,
			CodeMinLength: 3,
		},
	}
}

// GenerateShortCode implements UrlShortner.
func (u *UrlShortner) GenerateShortCode(ctx context.Context) (string, error) {
	var shortCode string
	var existingShort *ShortUrl
	var err error
	var tries int = u.opt.RetryCount
	totalShortCodeCount, err := u.store.CountShortUrls(ctx, &ShortUrlFilter{})
	if err != nil {
		return "", err
	}
	minLength := u.CalculateMinimumLength(totalShortCodeCount)
	for {
		shortCode = security.RandomString(int(minLength))
		existingShort, err = u.store.FindByShortCode(ctx, shortCode)
		if err != nil {
			return "", err
		}
		if existingShort == nil {
			break
		}
		slog.Info(fmt.Sprintf("Duplicate found, retrying: %s\n. Tries left: %d\n", shortCode, tries))
		tries--
		if tries == 0 {
			return "", fmt.Errorf("unable to generate unique short code")
		}
	}
	return shortCode, nil
}

// GenerateShortUrl implements UrlShortner.
func (u *UrlShortner) GenerateShortUrl(ctx context.Context, sourceUrl string) (*ShortUrl, error) {
	existingShort, err := u.store.FindBySourceUrl(ctx, sourceUrl)
	if err != nil {
		return nil, err
	}
	if existingShort != nil {
		return existingShort, nil
	}
	shortCode, err := u.GenerateShortCode(ctx)
	if err != nil {
		return nil, err
	}
	shortUrl := &ShortUrl{
		ShortCode: shortCode,
		SourceUrl: sourceUrl,
	}
	err = u.store.SaveShortUrl(ctx, shortUrl)
	if err != nil {
		return nil, err
	}
	return shortUrl, nil
}

func (u *UrlShortner) CalculateMinimumLength(n int64) int64 {
	estimate := security.EstimateLength(
		n,
		int64(len(security.DefaultRandomAlphabet)),
	)
	return int64(math.Max(
		float64(estimate),
		float64(u.opt.CodeMinLength),
	))
}
