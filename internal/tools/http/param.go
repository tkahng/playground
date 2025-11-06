package http

import (
	"net/http"
	"net/url"

	"github.com/tkahng/playground/internal/tools/http/queryparam"
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

func GetQuery(r *http.Request, name string) string {
	return queryparam.Get(r.URL.RawQuery, name)
}

type UrlGetFunc func(r *http.Request, name string) string

var urlGetFuncs = []UrlGetFunc{
	GetParam,
	GetQuery,
}

func GetRequestValueByName(r *http.Request, name string) string {
	for _, urlGetFunc := range urlGetFuncs {
		value := urlGetFunc(r, name)
		if value != "" {
			return value
		}
	}
	return ""
}
