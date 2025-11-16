package http

import (
	"net/url"
	"strings"
)

func EncodeRedirectURL(referrerURL string) string {
	if len(referrerURL) > 0 {
		if strings.ContainsAny(referrerURL, "&=#") {
			// if the string contains &, = or # it has not been URL
			// encoded by the caller, which means it should be URL
			// encoded by us otherwise, it should be taken as-is
			referrerURL = url.QueryEscape(referrerURL)
		}
	}
	return referrerURL
}
