package http

import (
	"net/http"
	"net/url"
)

func GetParam(r *http.Request, name string) string {
	v := r.PathValue(name)
	if r.URL.RawPath == "" {
		return v // RawPath empty means no escaping was done
	}
	u, err := url.PathUnescape(v)
	if err != nil {
		return v // not supposed to happen, but if it does, return the original value
	}
	return u
}
