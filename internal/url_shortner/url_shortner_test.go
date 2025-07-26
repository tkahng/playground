package urlshortner

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestNewUrlShortner(t *testing.T) {
	type args struct {
		store ShortUrlStore
	}
	tests := []struct {
		name string
		args args
		want *UrlShortner
	}{
		{
			name: "test new url shortner",
			args: args{
				store: func() ShortUrlStore {
					return NewInMemoryShortUrlStore()
				}(),
			},
			want: &UrlShortner{
				store: NewInMemoryShortUrlStore(),
				opt: UrlShortnerOptions{
					RetryCount:    3,
					CodeMinLength: 3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUrlShortner(tt.args.store); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUrlShortner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUrlShortner_CalculateMinimumLength(t *testing.T) {
	type args struct {
		n int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "one mil needs 4 char long code",
			args: args{
				n: 1_000_000,
			},
			want: 4,
		},
		{
			name: "10 needs minimum 3 char long code",
			args: args{
				n: 10,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryShortUrlStore()
			u := NewUrlShortner(store)
			if got := u.CalculateMinimumLength(tt.args.n); got != tt.want {
				t.Errorf("UrlShortner.CalculateMinimumLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUrlShortner_ShortenUrl(t *testing.T) {
	type fields struct {
		store ShortUrlStore
		u     *UrlShortner
	}
	type args struct {
		ctx       context.Context
		sourceUrl string
	}
	tests := []struct {
		name     string
		setup    func() *fields
		args     args
		wantFunc func(string) error
		wantErr  bool
	}{
		{
			name: "create first short url",
			args: args{
				ctx:       context.Background(),
				sourceUrl: "https://google.com",
			},
			setup: func() *fields {
				store := NewInMemoryShortUrlStore()
				short := NewUrlShortner(store)
				return &fields{
					store: store,
					u:     short,
				}
			},
			wantFunc: func(s string) error {
				if len(s) == 3 {
					return nil
				}
				return errors.New("short url should be 3 char long")
			},
			wantErr: false,
		},
		{
			name: "create when count is one million",
			setup: func() *fields {
				store := NewInMemoryShortUrlStoreDecorator()
				store.CountShortUrlsFunc = func(ctx context.Context, filter *ShortUrlFilter) (int64, error) {
					return 1_000_000, nil
				}
				short := NewUrlShortner(store)
				return &fields{
					store: store,
					u:     short,
				}
			},
			args: args{
				ctx:       context.Background(),
				sourceUrl: "https://google.com",
			},
			wantFunc: func(s string) error {
				if len(s) == 4 {
					return nil
				}
				return errors.New("short url should be 4 char long")
			},
			wantErr: false,
		},
		{
			name: "error fail to create unique short code in retry",
			setup: func() *fields {
				store := NewInMemoryShortUrlStoreDecorator()
				store.CountShortUrlsFunc = func(ctx context.Context, filter *ShortUrlFilter) (int64, error) {
					return 1_000_000, nil
				}
				store.FindByShortCodeFunc = func(ctx context.Context, shortCode string) (*ShortUrl, error) {
					return &ShortUrl{}, nil
				}
				short := NewUrlShortner(store)
				return &fields{
					store: store,
					u:     short,
				}
			},
			args: args{
				ctx:       context.Background(),
				sourceUrl: "https://google.com",
			},
			wantFunc: func(string) error {
				return nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.setup()
			got, err := fields.u.ShortenUrl(tt.args.ctx, tt.args.sourceUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("UrlShortner.ShortenUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantFunc != nil {
				if err := tt.wantFunc(got); err != nil {
					t.Errorf("UrlShortner.ShortenUrl() wantFunc error = %v", err)
				}
			}
		})
	}
}
