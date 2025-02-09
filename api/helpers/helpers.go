package helpers

import (
	"os"
	"strings"
)

func EnforceHTTP(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

func RemoveDomainError(url string) bool {
	domain := strings.ToLower(os.Getenv("DOMAIN"))

	// normalize input URL
	url = strings.ToLower(url)
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "www.")
	url = strings.Split(url, "/")[0]

	return url != domain
}
