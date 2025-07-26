package urlshortner

import (
	"context"
	"fmt"
	"math"

	"github.com/tkahng/playground/internal/tools/security"
)

type UrlShortnerOptions struct {
	RetryCount int
}

type UrlShortner struct {
	store ShortUrlStore
	opt   UrlShortnerOptions
}

// generateShortCode implements UrlShortner.
func (u *UrlShortner) generateShortCode(ctx context.Context) (string, error) {
	var shortCode string
	var existingShort *ShortUrl
	var err error
	var tries int = u.opt.RetryCount
	totalShortCodeCount, err := u.store.CountShortUrls(ctx, &ShortUrlFilter{})
	if err != nil {
		return "", err
	}
	minLength := CalculateMinimumLength(totalShortCodeCount)
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
	shortCode, err := u.generateShortCode(ctx)
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

func NewUrlShortner(store ShortUrlStore) *UrlShortner {
	return &UrlShortner{
		store: store,
	}
}

func CalculateMinimumLength(n int64) int64 {
	length := len(security.DefaultRandomAlphabet)
	return int64(math.Max(float64(security.EstimateLength(n, int64(length))), 4.0))
}
