package middlewares

import (
	"fmt"
	"net/http"

	"github.com/tkahng/authgo/internal/core"
)

func Verify(app core.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// token, err := VerifyRequest(app, w, r)
			// ctx = NewContext(ctx, token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}
func Authenticator(ja core.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fmt.Println("Authenticator")
		hfn := func(w http.ResponseWriter, r *http.Request) {
			if claims := core.GetContextUserInfo(r.Context()); claims == nil {
				fmt.Println("not authorized")
				http.Error(w, "not authorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

// func NewContext(ctx context.Context, t *shared.UserInfoDto, err error) context.Context {
// 	// ctx = core.SetContextUserClaims(ctx, t)
// 	// ctx = context.WithValue(ctx, ErrorCtxKey, err)
// 	return ctx
// }

// func VerifyRequest(ja core.App, w http.ResponseWriter, r *http.Request) (*shared.UserInfoDto, error) {
// 	fmt.Println("VerifyRequest")
// 	// user, err := ja.Auth().VerifyCookieTokenStd(w, r)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return nil, err
// 	// }
// 	user := models.User{}
// 	// fmt.Println(user)
// 	return &shared.UserInfoDto{
// 		User: user,
// 	}, nil

// }
