package urlshortner

import (
	"context"
	"fmt"
	"math"

	"github.com/tkahng/playground/internal/tools/security"
)

type UrlShortner struct {
	store ShortUrlStore
}

// generateShortCode implements UrlShortner.
func (u *UrlShortner) generateShortCode(ctx context.Context) (string, error) {
	var shortCode string
	var existingShort *ShortUrl
	var err error
	for {
		shortCode = generateShortCode()
		existingShort, err = u.store.FindByShortCode(ctx, shortCode)
		if err != nil {
			return "", err
		}
		if existingShort == nil {
			break
		}
		fmt.Printf("Duplicate found, retrying: %s\n", shortCode)
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
	err = u.store.CreateShortUrl(ctx, shortUrl)
	if err != nil {
		return "", err
	}
	return shortCode, nil
}

func generateShortCode() string {
	return security.RandomString(6)
}

func NewUrlShortner(store ShortUrlStore) *UrlShortner {
	return &UrlShortner{
		store: store,
	}
}

func EstimateLength(n int64, alphabetSize int64) int64 {
	length := math.Log10(float64(n)) / math.Log10(float64(alphabetSize))
	fmt.Println("length", length)
	return int64(math.Ceil(length))
}

func CalculateMinimumLength(n int64, alphabetSize int64) int64 {
	return int64(math.Max(float64(EstimateLength(n, alphabetSize)), 4.0))
}
