package urlshortner

import (
	"context"
	"fmt"
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
		fmt.Printf("Duplicate found, retrying: %s\n. Tries left: %d\n", shortCode, tries)
		tries--
		if tries == 0 {
			return "", fmt.Errorf("unable to generate unique short code")
		}
	}
	return shortCode, nil
}

// ShortenUrl implements UrlShortner.
func (u *UrlShortner) ShortenUrl(ctx context.Context, sourceUrl string) (string, error) {
	existingShort, err := u.store.FindBySourceUrl(ctx, sourceUrl)
	if err != nil {
		return "", err
	}
	if existingShort != nil {
		return existingShort.ShortCode, nil
	}
	shortCode, err := u.GenerateShortCode(ctx)
	if err != nil {
		return "", err
	}
	shortUrl := &ShortUrl{

		ShortCode: shortCode,
		SourceUrl: sourceUrl,
	}
	err = u.store.SaveShortUrl(ctx, shortUrl)
	if err != nil {
		return "", err
	}
	return shortCode, nil
}

func (u *UrlShortner) CalculateMinimumLength(n int64) int64 {
	estimate := security.EstimateLength(
		n,
		int64(len(security.DefaultRandomAlphabet)),
	)
	fmt.Printf("Estimate: %d\n", estimate)
	return int64(math.Max(
		float64(estimate),
		float64(u.opt.CodeMinLength),
	))
}
