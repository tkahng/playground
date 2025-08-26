package test

import (
	"fmt"
	"net/url"
	"regexp"
)

var (
	// LinkRegex is a regular expression to extract a link from an HTML email.
	LinkRegex = regexp.MustCompile(`href\s*=\s*"([^"]+)"`)
)

func GetLinkParam(html, paramName string) (string, error) {
	// Compile regex to extract href value
	re := regexp.MustCompile(`href\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return "", fmt.Errorf("no link found in HTML")
	}
	href := matches[1]

	// Parse the URL
	parsed, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	// Extract query parameter
	val := parsed.Query().Get(paramName)
	return val, nil
}
