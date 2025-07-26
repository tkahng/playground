package urlshortner

import (
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
