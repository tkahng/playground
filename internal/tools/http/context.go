package http

import (
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/queryparam"
)

// MultipartMaxMemory is the maximum memory to use when parsing multipart
// form data.
var MultipartMaxMemory int64 = 8 * 1024

// Unwrap extracts the underlying HTTP request and response writer from a Huma
// context. If passed a context from a different adapter it will panic.
func Unwrap(ctx huma.Context) (*http.Request, http.ResponseWriter) {
	for {
		if c, ok := ctx.(interface{ Unwrap() huma.Context }); ok {
			ctx = c.Unwrap()
			continue
		}
		break
	}
	if c, ok := ctx.(*playgroundContext); ok {
		return c.Unwrap()
	}
	panic("not a playground context")
}

type playgroundContext struct {
	op     *huma.Operation
	r      *http.Request
	w      http.ResponseWriter
	status int
}

// check that chiContext implements huma.Context
var _ huma.Context = &playgroundContext{}

func (c *playgroundContext) Unwrap() (*http.Request, http.ResponseWriter) {
	return c.r, c.w
}

func (c *playgroundContext) Operation() *huma.Operation {
	return c.op
}

func (c *playgroundContext) Context() context.Context {
	return c.r.Context()
}

func (c *playgroundContext) Method() string {
	return c.r.Method
}

func (c *playgroundContext) Host() string {
	return c.r.Host
}

func (c *playgroundContext) RemoteAddr() string {
	return c.r.RemoteAddr
}

func (c *playgroundContext) URL() url.URL {
	return *c.r.URL
}

func (c *playgroundContext) Param(name string) string {
	v := c.r.PathValue(name)
	if c.r.URL.RawPath == "" {
		return v // RawPath empty means no escaping was done
	}
	u, err := url.PathUnescape(v)
	if err != nil {
		return v // not supposed to happen, but if it does, return the original value
	}
	return u
}

func (c *playgroundContext) Query(name string) string {
	return queryparam.Get(c.r.URL.RawQuery, name)
}

func (c *playgroundContext) Header(name string) string {
	return c.r.Header.Get(name)
}

func (c *playgroundContext) EachHeader(cb func(name, value string)) {
	for name, values := range c.r.Header {
		for _, value := range values {
			cb(name, value)
		}
	}
}

func (c *playgroundContext) BodyReader() io.Reader {
	return c.r.Body
}

func (c *playgroundContext) GetMultipartForm() (*multipart.Form, error) {
	err := c.r.ParseMultipartForm(MultipartMaxMemory)
	return c.r.MultipartForm, err
}

func (c *playgroundContext) SetReadDeadline(deadline time.Time) error {
	return huma.SetReadDeadline(c.w, deadline)
}

func (c *playgroundContext) SetStatus(code int) {
	c.status = code
	c.w.WriteHeader(code)
}

func (c *playgroundContext) Status() int {
	return c.status
}

func (c *playgroundContext) AppendHeader(name string, value string) {
	c.w.Header().Add(name, value)
}

func (c *playgroundContext) SetHeader(name string, value string) {
	c.w.Header().Set(name, value)
}

func (c *playgroundContext) BodyWriter() io.Writer {
	return c.w
}

func (c *playgroundContext) TLS() *tls.ConnectionState {
	return c.r.TLS
}

func (c *playgroundContext) Version() huma.ProtoVersion {
	return huma.ProtoVersion{
		Proto:      c.r.Proto,
		ProtoMajor: c.r.ProtoMajor,
		ProtoMinor: c.r.ProtoMinor,
	}
}

// NewPlaygroundContext creates a new Huma context from an HTTP request and response.
func NewPlaygroundContext(op *huma.Operation, r *http.Request, w http.ResponseWriter) huma.Context {
	return &playgroundContext{op: op, r: r, w: w}
}
