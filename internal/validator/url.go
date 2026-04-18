package validator

import (
	"net/url"
	"strings"
)

func ValidateUrl(inputUrl string) bool {
	inputUrl = strings.TrimSpace(inputUrl)

	if inputUrl == "" {
		return false
	}

	parsedUrl, err := url.ParseRequestURI(inputUrl)
	if err != nil {
		return false
	}

	if parsedUrl.Host == "" {
		return false
	}

	if parsedUrl.Scheme == "" {
		return false
	}

	return true
}
