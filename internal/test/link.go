package test

import (
	"fmt"
	"net/url"
	"regexp"
)

var (
	// LinkRegex is a regular expression to extract a link from an HTML email.
	LinkRegex = regexp.MustCompile(`href\s*=\s*"([^"]+)"`)
	CodeRegex = regexp.MustCompile(`code:\s*(\d{6})`)
)

func GetLinkParam(html, paramName string) (string, error) {
	// Compile regex to extract href value
	matches := LinkRegex.FindStringSubmatch(html)
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
func GetCodeParam(html string) (string, error) {
	// Compile regex to extract href value
	matches := CodeRegex.FindAllStringSubmatch(html, -1)
	if len(matches) < 1 {
		return "", fmt.Errorf("no code found in HTML")
	}
	code := matches[0][1]

	if code == "" {
		return "", fmt.Errorf("code parameter is empty")
	}
	return code, nil
}
