package cookie

import (
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type CookieName string

const (
	SessionCookieName      string = "k2dv-session"
	AccessTokenCookieName  string = "access_token"
	RefreshTokenCookieName string = "refresh_token"
)

func (c CookieName) String() string {
	return string(c)
}

func TokenFromCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(string(name))
	if err != nil {
		return ""
	}
	return cookie.Value
}

func AppTokenFromCookie(ctx AppContext, name string) string {
	cookie, err := huma.ReadCookie(ctx, name)
	//  ctx.Header()
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SetTokenCookie(w http.ResponseWriter, name string, value string, expires time.Time) {
	rtCookie := http.Cookie{
		Name:     string(name),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Value:    value,
		Expires:  expires,
	}
	http.SetCookie(w, &rtCookie)
}

func CreateTokenCookie(name string, value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     string(name),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Value:    value,
		Expires:  expires,
	}
}

func SetTokenCookieApi(name string, value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     string(name),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Value:    value,
		Expires:  expires,
	}
}

func RemoveCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:    string(name),
		MaxAge:  -1,
		Expires: time.Now().Add(-100 * time.Hour), // Set expires for older versions of IE
		Path:    "/",
	}
}
func RemoveTokenCookie(w http.ResponseWriter, name string) {
	rtCookie := http.Cookie{
		Name:    string(name),
		MaxAge:  -1,
		Expires: time.Now().Add(-100 * time.Hour), // Set expires for older versions of IE
		Path:    "/",
	}
	http.SetCookie(w, &rtCookie)
}

type AppContext interface {
	// Operation returns the OpenAPI operation that matched the request.
	Operation() *huma.Operation

	// Context returns the underlying request context.
	Context() context.Context

	// TLS / SSL connection information.
	TLS() *tls.ConnectionState

	// Version of the HTTP protocol as text and integers.
	Version() huma.ProtoVersion

	// Method returns the HTTP method for the request.
	Method() string

	// Host returns the HTTP host for the request.
	Host() string

	// RemoteAddr returns the remote address of the client.
	RemoteAddr() string

	// URL returns the full URL for the request.
	URL() url.URL

	// Param returns the value for the given path parameter.
	Param(name string) string

	// Query returns the value for the given query parameter.
	Query(name string) string

	// Header returns the value for the given header.
	Header(name string) string

	// EachHeader iterates over all headers and calls the given callback with
	// the header name and value.
	EachHeader(cb func(name, value string))

	// BodyReader returns the request body reader.
	BodyReader() io.Reader

	// GetMultipartForm returns the parsed multipart form, if any.
	GetMultipartForm() (*multipart.Form, error)

	// SetReadDeadline sets the read deadline for the request body.
	SetReadDeadline(time.Time) error

	// SetStatus sets the HTTP status code for the response.
	SetStatus(code int)

	// Status returns the HTTP status code for the response.
	Status() int

	// SetHeader sets the given header to the given value, overwriting any
	// existing value. Use `AppendHeader` to append a value instead.
	SetHeader(name, value string)

	// AppendHeader appends the given value to the given header.
	AppendHeader(name, value string)

	// BodyWriter returns the response body writer.
	BodyWriter() io.Writer
}
