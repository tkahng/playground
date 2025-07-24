package middleware

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
)

func WithContext(parent huma.Context, ctx context.Context) huma.Context {
	r, w := humachi.Unwrap(parent)
	r = r.WithContext(ctx) // ✨
	humactx := humachi.NewContext(parent.Operation(), r, w)
	return humactx
}
