package http

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/tkahng/playground/internal/tools/http/queryparam"
)

type UrlGetFunc func(r *http.Request, name string) string

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

var urlGetFuncs = []UrlGetFunc{
	func(r *http.Request, name string) string {
		name = strings.ReplaceAll(name, "_", "-")
		return GetParam(r, name)
	},
	func(r *http.Request, name string) string {
		name = strings.ReplaceAll(name, "-", "_")
		return GetQuery(r, name)
	},
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
