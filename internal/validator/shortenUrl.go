package validator

import (
	"net/url"
	"strings"
)

func ValidateURL(inputURL string) bool {
	inputURL = strings.TrimSpace(inputURL)

	if inputURL == "" {
		return false
	}

	parsedURL, err := url.ParseRequestURI(inputURL)
	if err != nil {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	if parsedURL.Scheme == "" {
		return false
	}

	return true
}
