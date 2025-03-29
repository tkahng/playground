package mailer

import (
	URL "net/url"
)

func BuildUrl(origin string, params map[string]string) string {
	uri, err := URL.Parse(origin)
	if err != nil {
		return ""
	}
	val := URL.Values{}
	for k, v := range params {
		val.Add(k, v)
	}
	uri.RawQuery = val.Encode()
	return uri.String()

}
