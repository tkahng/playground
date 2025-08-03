package urlshortner

import (
	"context"
	"fmt"
	"testing"
)

func TestInMemoryShortUrlStore(t *testing.T) {

	store := NewInMemoryShortUrlStore()
	ctx := context.Background()
	for i := range 5 {
		err := store.SaveShortUrl(ctx, &ShortUrl{
			ShortCode: fmt.Sprintf("shortCode%d", i),
			SourceUrl: fmt.Sprintf("sourceUrl%d", i),
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := range 5 {
		shortUrl, err := store.FindByShortCode(ctx, fmt.Sprintf("shortCode%d", i))
		if err != nil {
			t.Fatal(err)
		}
		if shortUrl.ShortCode != fmt.Sprintf("shortCode%d", i) {
			t.Fatalf("Expected shortCode%d, got %s", i, shortUrl.ShortCode)
		}
		if shortUrl.SourceUrl != fmt.Sprintf("sourceUrl%d", i) {
			t.Fatalf("Expected sourceUrl%d, got %s", i, shortUrl.SourceUrl)
		}
	}
	for i := range 5 {
		shortUrl, err := store.FindBySourceUrl(ctx, fmt.Sprintf("sourceUrl%d", i))
		if err != nil {
			t.Fatal(err)
		}
		if shortUrl.ShortCode != fmt.Sprintf("shortCode%d", i) {
			t.Fatalf("Expected shortCode%d, got %s", i, shortUrl.ShortCode)
		}
		if shortUrl.SourceUrl != fmt.Sprintf("sourceUrl%d", i) {
			t.Fatalf("Expected sourceUrl%d, got %s", i, shortUrl.SourceUrl)
		}
	}
	count, err := store.CountShortUrls(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Fatalf("Expected 5, got %d", count)
	}
}
