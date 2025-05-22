package apis

import (
	"net/http"

	"github.com/tkahng/authgo/internal/shared"
)

type SetCookieOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
}

type CookieInput struct {
	AccessToken  *http.Cookie `cookie:"access_token" required:"false"`
	RefreshToken *http.Cookie `cookie:"refresh_token" required:"false"`
}

type AuthenticatedResponse struct {
	// SetCookieOutput
	SetCookie []http.Cookie `header:"Set-Cookie"`

	Body shared.AuthenticatedDTO `json:"body"`
}
